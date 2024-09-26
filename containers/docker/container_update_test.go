package docker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeUpdatingDocker struct {
	fakeDocker
	summaries []types.ImageSummary
	cnts      []types.Container
	cntJSON   *types.ContainerJSON

	c chan struct{}

	Ports       nat.PortSet
	Env         []string
	Volumes     []mount.Mount
	ContainerID string

	Instance       string
	Duration       int
	RemoveInstance string

	InvocationContainerCreate  int
	InvocationContainerStart   int
	InvocationContainerList    int
	InvocationContainerStop    int
	InvocationContainerRemove  int
	InvocationImageList        int
	InvocationContainerInspect int
}

func (f *fakeUpdatingDocker) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	f.InvocationContainerCreate++

	// This enables the synchronous test to explicitly fail the m.ContainerStart call.
	if config.Cmd != nil && config.Cmd[0] == "fail-container-create" {
		return container.CreateResponse{}, status.Error(codes.Internal, "instructed to fail during test")
	}

	// Actually create container -- needed for multi-step async test.
	f.cnts = append(f.cnts, types.Container{
		Names: []string{containerName},
		ID:    containerName,
	})

	f.Ports = config.ExposedPorts
	f.Env = config.Env
	f.Volumes = hostConfig.Mounts

	return container.CreateResponse{
		ID: containerName,
	}, nil
}

func (f *fakeUpdatingDocker) ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error {
	f.InvocationContainerStart++

	// This enables the asynchronous test to explicitly block the m.ContainerStart call.
	if container == "block-till-released" {
		<-f.c
	}

	f.ContainerID = container
	return nil
}

func (f *fakeUpdatingDocker) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	f.InvocationContainerList++
	return f.cnts, nil
}

func (f *fakeUpdatingDocker) ContainerStop(ctx context.Context, container string, options container.StopOptions) error {
	f.InvocationContainerStop++
	f.Instance = container

	if options.Timeout != nil {
		f.Duration = *options.Timeout
	}

	// We actually have to perform a container removal here.
	newcnts := []types.Container{}
cntloop:
	for _, cnt := range f.cnts {
		for _, name := range cnt.Names {
			if name == container {
				continue cntloop
			}
		}
		newcnts = append(newcnts, cnt)
	}
	f.cnts = newcnts
	return nil
}

func (f *fakeUpdatingDocker) ContainerRemove(ctx context.Context, container string, options types.ContainerRemoveOptions) error {
	f.InvocationContainerRemove++
	return nil
}

func (f *fakeUpdatingDocker) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	f.InvocationImageList++
	return f.summaries, nil
}

func (f *fakeUpdatingDocker) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	f.InvocationContainerInspect++
	if f.cntJSON == nil {
		return types.ContainerJSON{}, fmt.Errorf("bonito-flakes")
	}
	return *f.cntJSON, nil
}

func TestContainerUpdateSync(t *testing.T) {
	tests := []struct {
		name         string
		inOpts       []options.Option
		inInstance   string
		inImage      string
		inInProgress []string
		inTag        string
		inCmd        string
		inSummaries  []types.ImageSummary
		inCnts       []types.Container
		inCntJSON    *types.ContainerJSON
		wantState    *fakeUpdatingDocker
		wantErr      error
	}{
		{
			name:    "failure-image-not-found",
			inImage: "no-such-image",
			inTag:   "no-such-tag",
			wantErr: status.Error(codes.NotFound, "image no-such-image:no-such-tag not found"),
		},
		{
			name:    "failure-image-found-container-not-found",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"not-the-right-container"},
				},
			},
			wantErr: status.Errorf(codes.NotFound, "instance name container-I-want not found"),
		},
		{
			name:    "failure-image-found-container-found-port-unavailable",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"container-I-want"},
				},
				types.Container{
					Names: []string{"conflicting-container"},
					Ports: []types.Port{types.Port{PublicPort: 1}},
				},
			},
			inOpts:  []options.Option{options.WithInstanceName("container-I-want"), options.WithPorts(map[uint32]uint32{1: 1})},
			wantErr: status.Errorf(codes.Unavailable, "port 1 already in use"),
		},
		{
			name:    "failure-image-found-container-found-port-reusable-instance-progressing",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"container-I-want"},
					Ports: []types.Port{types.Port{PublicPort: 1}},
				},
				types.Container{
					Names: []string{"conflicting-container"},
				},
			},
			inInProgress: []string{"container-I-want"},
			inOpts:       []options.Option{options.WithInstanceName("container-I-want"), options.WithPorts(map[uint32]uint32{1: 1})},
			wantErr:      status.Errorf(codes.Unavailable, "container container-I-want is already being updated"),
		},
		{
			name:    "failure-image-found-container-found-port-reused-another-instance-progressing-json-missing",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"some-container"},
					Ports: []types.Port{types.Port{PublicPort: 1}},
				},
				types.Container{
					Names: []string{"conflicting-container"},
				},
			},
			inInProgress: []string{"conflicting-container"},
			inOpts:       []options.Option{options.WithInstanceName("container-I-want"), options.WithPorts(map[uint32]uint32{1: 1})},
			wantErr:      status.Errorf(codes.NotFound, "instance name container-I-want not found"),
		},
		{
			name:    "failure-image-found-container-found-port-reused-another-instance-progressing-inspect-failed",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"container-I-want"},
					Ports: []types.Port{types.Port{PublicPort: 1}},
					ID:    "salmon-roe",
				},
				types.Container{
					Names: []string{"conflicting-container"},
				},
			},
			inInProgress: []string{"conflicting-container"},
			inOpts:       []options.Option{options.WithInstanceName("container-I-want"), options.WithPorts(map[uint32]uint32{1: 1})},
			wantErr:      status.Errorf(codes.Unknown, "failed to inspect container salmon-roe: bonito-flakes"),
		},
		{
			name:    "success-image-found-container-found-port-reused-another-instance-progressing-inspect-worked-not-running",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"container-I-want"},
					Ports: []types.Port{types.Port{PublicPort: 1}},
					ID:    "salmon-roe",
				},
				types.Container{
					Names: []string{"conflicting-container"},
				},
			},
			inCntJSON: &types.ContainerJSON{
				&types.ContainerJSONBase{
					HostConfig: &container.HostConfig{},
					State: &types.ContainerState{
						Paused: true,
					},
				},
				nil, &container.Config{}, &types.NetworkSettings{},
			},
			inInProgress: []string{"conflicting-container"},
			inOpts:       []options.Option{options.WithInstanceName("container-I-want"), options.WithPorts(map[uint32]uint32{1: 1})},
			inCmd:        "my-cmd",
			wantState: &fakeUpdatingDocker{
				Instance:                   "container-I-want",
				Duration:                   -1,
				Ports:                      nat.PortSet{"1/tcp": struct{}{}},
				ContainerID:                "container-I-want",
				Volumes:                    []mount.Mount{},
				InvocationContainerCreate:  1,
				InvocationContainerStart:   1,
				InvocationContainerList:    3, // ContainerUpdate, ContainerStop, and ContainerStart checks.
				InvocationContainerStop:    1,
				InvocationContainerRemove:  1,
				InvocationImageList:        2, // ContainerUpdate and ContainerStart checks.
				InvocationContainerInspect: 1,
			},
		},
		{
			name:    "restoration-success",
			inImage: "my-image",
			inTag:   "my-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"my-image:my-tag"},
				},
			},
			inInstance: "container-I-want",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"container-I-want"},
					Ports: []types.Port{types.Port{PublicPort: 1}},
					ID:    "salmon-roe",
				},
				types.Container{
					Names: []string{"conflicting-container"},
				},
			},
			inCntJSON: &types.ContainerJSON{
				&types.ContainerJSONBase{
					HostConfig: &container.HostConfig{},
					State: &types.ContainerState{
						Paused: true,
					},
				},
				nil, &container.Config{}, &types.NetworkSettings{},
			},
			inInProgress: []string{"conflicting-container"},
			inOpts:       []options.Option{options.WithInstanceName("container-I-want"), options.WithPorts(map[uint32]uint32{1: 1})},
			inCmd:        "fail-container-create",
			wantErr:      status.Error(codes.Internal, "failed to update instance container-I-want due to: rpc error: code = Internal desc = unable to create container: rpc error: code = Internal desc = instructed to fail during test; yet, restoration of previous state succeeded"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fsd := &fakeUpdatingDocker{
				summaries: tc.inSummaries,
				cnts:      tc.inCnts,
				cntJSON:   tc.inCntJSON,
			}
			mgr := New(fsd)
			for _, inProgress := range tc.inInProgress {
				mgr.updateInProgress[inProgress] = struct{}{}
			}

			if _, err := mgr.ContainerUpdate(context.Background(), tc.inInstance, tc.inImage, tc.inTag, tc.inCmd, false, tc.inOpts...); err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ContainerUpdate(%q, %q, %q, %q, %+v) returned unexpected error(-want, got):\n %s", tc.inInstance, tc.inImage, tc.inTag, tc.inCmd, tc.inOpts, diff)
					}
					return
				}
				t.Errorf("ContainerUpdate(%q, %q, %q, %q, %+v) returned error: %v", tc.inInstance, tc.inImage, tc.inTag, tc.inCmd, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fsd, cmpopts.IgnoreUnexported(fakeUpdatingDocker{}), cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("ContainerUpdate(%q, %q, %q, %q, %+v) returned diff(-want, +got):\n%s", tc.inInstance, tc.inImage, tc.inTag, tc.inCmd, tc.inOpts, diff)
				}
			}
		})
	}
}

func TestContainerUpdateAsync(t *testing.T) {
	// This test performs the following sequence of steps:
	// 1) Start Async & blocking update of container A (named block-till-released to trigger blocking
	//		in ContainerStart(...))
	// 2) Attempt another update of container A while last update still blocking -> should fail
	// 3) Attempt update of container B -> should work
	// 4) Unblock container A & give it some time to finish up.
	// 5) Attempt another update of container A -> should work

	fd := &fakeUpdatingDocker{
		summaries: []types.ImageSummary{
			types.ImageSummary{
				RepoTags: []string{"image-A1:tag-A1"},
			},
			types.ImageSummary{
				RepoTags: []string{"image-A2:tag-A2"},
			},
			types.ImageSummary{
				RepoTags: []string{"image-B1:tag-B1"},
			},
			types.ImageSummary{
				RepoTags: []string{"image-B2:tag-B2"},
			},
		},
		cnts: []types.Container{
			types.Container{
				Names: []string{"block-till-released"},
				ID:    "block-till-released",
			},
			types.Container{
				Names: []string{"container-B"},
				ID:    "container-B",
			},
		},
		cntJSON: &types.ContainerJSON{
			&types.ContainerJSONBase{
				HostConfig: &container.HostConfig{},
				State: &types.ContainerState{
					Running: true,
				},
			},
			nil, &container.Config{}, &types.NetworkSettings{},
		},
		c: make(chan struct{}),
	}
	mgr := New(fd)

	// Start Async & blocking update of container A.
	if _, err := mgr.ContainerUpdate(context.Background(), "block-till-released", "image-A1", "tag-A1", "", true); err != nil {
		t.Fatalf("ContainerUpdate(context.Background(), block-till-released, image-A1, tag-A1, , true) returned unexpected error: %v", err)
	}

	// Attempt another update of container A -> should fail.
	if _, err := mgr.ContainerUpdate(context.Background(), "block-till-released", "image-A2", "tag-A2", "", false); err == nil { // if NO error
		t.Fatalf("ContainerUpdate(ctx, block-till-released, image-A1, tag-A1, , false) unexpected succeeded. Expected: container ... is already being updated")
	}

	// Attempt update of container B -> should work.
	if _, err := mgr.ContainerUpdate(context.Background(), "container-B", "image-B1", "tag-B1", "", false); err != nil {
		t.Fatalf("ContainerUpdate(context.Background(), container-B, image-B1, tag-B1, , false) returned unexpected error: %v", err)
	}

	// Unblock container A & give it some time to finish up.
	fd.c <- struct{}{}
	time.Sleep(time.Millisecond * 100)

	// Start Async & blocking update of container A.
	if _, err := mgr.ContainerUpdate(context.Background(), "block-till-released", "image-A2", "tag-A2", "", true); err != nil {
		t.Fatalf("ContainerUpdate(context.Background(), block-till-released, image-A2, tag-A2, , true) returned unexpected error: %v", err)
	}
}

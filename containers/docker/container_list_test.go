package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeListContainerStreamer struct {
	msgs []*cpb.ListContainerResponse
}

func (f *fakeListContainerStreamer) Send(msg *cpb.ListContainerResponse) error {
	f.msgs = append(f.msgs, msg)
	return nil
}

type fakeListingDocker struct {
	fakeDocker
	cnts []types.Container

	Opts container.ListOptions
}

func (f *fakeListingDocker) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	f.Opts = options

	return f.cnts, nil
}

func TestContainerList(t *testing.T) {
	tests := []struct {
		name      string
		inOpts    []options.Option
		inCnts    []types.Container
		inAll     bool
		inLimit   int32
		wantState *fakeListingDocker
		wantMsgs  []*cpb.ListContainerResponse
	}{
		{
			name:    "no-containers",
			inAll:   true,
			inLimit: 10,
			wantState: &fakeListingDocker{
				Opts: container.ListOptions{
					Limit: 10,
					All:   true,
				},
			},
		},
		{
			name:    "containers-no-filter",
			inAll:   true,
			inLimit: 10,
			wantState: &fakeListingDocker{
				Opts: container.ListOptions{
					Limit: 10,
					All:   true,
				},
			},
			inCnts: []types.Container{
				types.Container{
					ID:    "some-id",
					Image: "some-image",
					Names: []string{"some-name"},
				},
				types.Container{
					ID:    "some-other-id",
					Image: "some-other-image",
					Names: []string{"some-other-name"},
				},
			},
			wantMsgs: []*cpb.ListContainerResponse{
				&cpb.ListContainerResponse{
					Id:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
				},
				&cpb.ListContainerResponse{
					Id:        "some-other-id",
					Name:      "some-other-name",
					ImageName: "some-other-image",
				},
			},
		},
		{
			name:    "filter",
			inAll:   true,
			inLimit: 10,
			inOpts: []options.Option{options.WithFilter(map[options.FilterKey][]string{
				options.Image: []string{"some-image"},
				options.State: []string{"RUNNING"},
			})},
			wantState: &fakeListingDocker{
				Opts: container.ListOptions{
					Limit:   10,
					All:     true,
					Filters: filters.NewArgs(filters.Arg("image", "some-image"), filters.Arg("state", "RUNNING")),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fsd := &fakeListingDocker{
				cnts: tc.inCnts,
			}
			mgr := New(fsd)

			stream := &fakeListContainerStreamer{}

			if err := mgr.ContainerList(ctx, tc.inAll, tc.inLimit, stream, tc.inOpts...); err != nil {
				t.Errorf("ContainerList(%t, %q, %+v) returned error: %v", tc.inAll, tc.inLimit, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fsd, cmpopts.IgnoreUnexported(fakeListingDocker{}, filters.Args{})); diff != "" {
					t.Errorf("ContainerList(%t, %q, %+v) returned diff(-want, +got):\n%s", tc.inAll, tc.inLimit, tc.inOpts, diff)
				}
			}

			if diff := cmp.Diff(tc.wantMsgs, stream.msgs, protocmp.Transform()); diff != "" {
				t.Errorf("ContainerList(%t, %q, %+v) returned diff(-want, +got):\n%s", tc.inAll, tc.inLimit, tc.inOpts, diff)
			}
		})
	}
}

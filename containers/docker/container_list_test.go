package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/moby/moby/v/v24/api/types/filters"
	"github.com/moby/moby/v/v24/api/types"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeListStreamer struct {
	msgs []*cpb.ListResponse
}

func (f *fakeListStreamer) Send(msg *cpb.ListResponse) error {
	f.msgs = append(f.msgs, msg)
	return nil
}

type fakeListingDocker struct {
	fakeDocker
	cnts []types.Container

	Opts types.ContainerListOptions
}

func (f *fakeListingDocker) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	f.Opts = options

	return f.cnts, nil
}

func TestContainerList(t *testing.T) {
	tests := []struct {
		name      string
		inOpts    []options.ImageOption
		inCnts    []types.Container
		inAll     bool
		inLimit   int32
		wantState *fakeListingDocker
		wantMsgs  []*cpb.ListResponse
	}{
		{
			name:    "no-containers",
			inAll:   true,
			inLimit: 10,
			wantState: &fakeListingDocker{
				Opts: types.ContainerListOptions{
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
				Opts: types.ContainerListOptions{
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
			wantMsgs: []*cpb.ListResponse{
				&cpb.ListResponse{
					Id:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
				},
				&cpb.ListResponse{
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
			inOpts: []options.ImageOption{options.WithFilter(map[options.FilterKey][]string{
				options.Image: []string{"some-image"},
				options.State: []string{"RUNNING"},
			})},
			wantState: &fakeListingDocker{
				Opts: types.ContainerListOptions{
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

			stream := &fakeListStreamer{}

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

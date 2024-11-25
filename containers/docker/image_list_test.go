package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeListImageStreamer struct {
	msgs []*cpb.ListImageResponse
}

func (f *fakeListImageStreamer) Send(msg *cpb.ListImageResponse) error {
	f.msgs = append(f.msgs, msg)
	return nil
}

type fakeImageListingDocker struct {
	fakeDocker
	imgs []imagetypes.Summary
	Opts imagetypes.ListOptions
}

func (f *fakeImageListingDocker) ImageList(ctx context.Context, options imagetypes.ListOptions) ([]imagetypes.Summary, error) {
	f.Opts = options
	return f.imgs, nil
}

func TestImageList(t *testing.T) {
	tests := []struct {
		name      string
		inOpts    []options.Option
		inImgs    []imagetypes.Summary
		inAll     bool
		inLimit   int32
		wantState *fakeImageListingDocker
		wantMsgs  []*cpb.ListImageResponse
	}{
		{
			name:  "no-containers",
			inAll: true,
			wantState: &fakeImageListingDocker{
				Opts: imagetypes.ListOptions{
					All: true,
				},
			},
		},
		{
			name:  "containers-no-filter",
			inAll: true,
			wantState: &fakeImageListingDocker{
				Opts: imagetypes.ListOptions{
					All: true,
				},
			},
			inImgs: []imagetypes.Summary{
				imagetypes.Summary{
					ID:       "some-id",
					RepoTags: []string{"some-image:some-tag"},
				},
				imagetypes.Summary{
					ID:       "some-other-id",
					RepoTags: []string{"some-other-image:some-other-tag"},
				},
			},
			wantMsgs: []*cpb.ListImageResponse{
				&cpb.ListImageResponse{
					Id:        "some-id",
					ImageName: "some-image",
					Tag:       "some-tag",
				},
				&cpb.ListImageResponse{
					Id:        "some-other-id",
					ImageName: "some-other-image",
					Tag:       "some-other-tag",
				},
			},
		},
		{
			name:  "containers-no-filter-limited",
			inAll: true,
			wantState: &fakeImageListingDocker{
				Opts: imagetypes.ListOptions{
					All: true,
				},
			},
			inImgs: []imagetypes.Summary{
				imagetypes.Summary{
					ID:       "some-id",
					RepoTags: []string{"some-image:some-tag"},
				},
				imagetypes.Summary{
					ID:       "some-other-id",
					RepoTags: []string{"some-other-image:some-other-tag"},
				},
			},
			inLimit: 1,
			wantMsgs: []*cpb.ListImageResponse{
				&cpb.ListImageResponse{
					Id:        "some-id",
					ImageName: "some-image",
					Tag:       "some-tag",
				},
			},
		},
		{
			name:  "containers-no-filter-multi-tags",
			inAll: true,
			wantState: &fakeImageListingDocker{
				Opts: imagetypes.ListOptions{
					All: true,
				},
			},
			inImgs: []imagetypes.Summary{
				imagetypes.Summary{
					ID:       "some-id",
					RepoTags: []string{"some-image:some-tag", "some-image:capybara"},
				},
				imagetypes.Summary{
					ID:       "some-other-id",
					RepoTags: []string{"some-other-image:some-other-tag"},
				},
			},
			wantMsgs: []*cpb.ListImageResponse{
				&cpb.ListImageResponse{
					Id:        "some-id",
					ImageName: "some-image",
					Tag:       "some-tag,capybara",
				},
				&cpb.ListImageResponse{
					Id:        "some-other-id",
					ImageName: "some-other-image",
					Tag:       "some-other-tag",
				},
			},
		},
		{
			name:  "filter",
			inAll: true,
			inOpts: []options.Option{options.WithFilter(map[options.FilterKey][]string{
				options.Image: []string{"some-image"},
				options.State: []string{"RUNNING"},
			})},
			wantState: &fakeImageListingDocker{
				Opts: imagetypes.ListOptions{
					All:     true,
					Filters: filters.NewArgs(filters.Arg("image", "some-image"), filters.Arg("state", "RUNNING")),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fsd := &fakeImageListingDocker{
				imgs: tc.inImgs,
			}
			mgr := New(fsd)

			stream := &fakeListImageStreamer{}

			if err := mgr.ImageList(ctx, tc.inAll, tc.inLimit, stream, tc.inOpts...); err != nil {
				t.Errorf("ImageList(%t, %+v) returned error: %v", tc.inAll, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fsd, cmpopts.IgnoreUnexported(fakeImageListingDocker{}, filters.Args{})); diff != "" {
					t.Errorf("ImageList(%t, %+v) returned diff(-want, +got):\n%s", tc.inAll, tc.inOpts, diff)
				}
			}

			if diff := cmp.Diff(tc.wantMsgs, stream.msgs, protocmp.Transform()); diff != "" {
				t.Errorf("ImageList(%t, %+v) returned diff(-want, +got):\n%s", tc.inAll, tc.inOpts, diff)
			}
		})
	}
}

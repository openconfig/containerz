package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/moby/moby/v/v24/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

type fakeRemovingDocker struct {
	fakeDocker
	summaries []types.ImageSummary
	cnts      []types.Container

	Name string
}

func (f fakeRemovingDocker) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return f.cnts, nil
}

func (f fakeRemovingDocker) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	return f.summaries, nil
}

func (f *fakeRemovingDocker) ImageRemove(ctx context.Context, image string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	f.Name = image
	return nil, nil
}

func TestContainerRemove(t *testing.T) {
	tests := []struct {
		name        string
		inOpts      []options.ImageOption
		inImage     string
		inTag       string
		inSummaries []types.ImageSummary
		inCnts      []types.Container
		wantState   *fakeRemovingDocker
		wantErr     error
	}{
		{
			name:    "no-such-image",
			inImage: "no-such-image",
			inTag:   "no-such-tag",
			wantErr: status.Error(codes.NotFound, "image no-such-image:no-such-tag not found"),
		},
		{
			name:    "container-running",
			inImage: "container-running",
			inTag:   "running-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"container-running:running-tag"},
				},
			},
			inCnts: []types.Container{
				types.Container{
					Image: "container-running:running-tag",
				},
			},
			wantErr: status.Error(codes.Unavailable, "image container-running:running-tag has a running container; use force to override"),
		},
		{
			name:    "container-running-with-force",
			inImage: "container-running",
			inTag:   "running-tag",
			inOpts:  []options.ImageOption{options.Force()},
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"container-running:running-tag"},
				},
			},
			inCnts: []types.Container{
				types.Container{
					Image: "container-running:running-tag",
				},
			},
			wantState: &fakeRemovingDocker{
				Name: "container-running:running-tag",
			},
		},
		{
			name:    "container-remove",
			inImage: "container-remove",
			inTag:   "remove-tag",
			inSummaries: []types.ImageSummary{
				types.ImageSummary{
					RepoTags: []string{"container-remove:remove-tag"},
				},
			},
			wantState: &fakeRemovingDocker{
				Name: "container-remove:remove-tag",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fpd := &fakeRemovingDocker{
				summaries: tc.inSummaries,
				cnts:      tc.inCnts,
			}
			mgr := New(fpd)

			if err := mgr.ContainerRemove(context.Background(), tc.inImage, tc.inTag, tc.inOpts...); err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ContainerRemove(%q, %q, %+v) returned unexpected error(-want, got):\n %s", tc.inImage, tc.inTag, tc.inOpts, diff)
					}
					return
				}
				t.Errorf("ContainerRemove(%q, %q, %+v) returned error: %v", tc.inImage, tc.inTag, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fpd, cmpopts.IgnoreUnexported(fakeRemovingDocker{})); diff != "" {
					t.Errorf("ContainerRemove(%q, %q, %+v) returned diff(-want, +got):\n%s", tc.inImage, tc.inTag, tc.inOpts, diff)
				}
			}
		})
	}
}

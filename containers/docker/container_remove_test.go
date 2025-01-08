package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

func (f *fakeRemovingDocker) ContainerRemove(ctx context.Context, cnt string, options container.RemoveOptions) error {
	f.Name = cnt
	return nil
}

func TestContainerRemove(t *testing.T) {
	tests := []struct {
		name        string
		inOpts      []options.Option
		inCnt       string
		inSummaries []imagetypes.Summary
		inCnts      []types.Container
		wantState   *fakeRemovingDocker
		wantErr     error
	}{
		{
			name:    "no-such-container",
			inCnt:   "no-such-container",
			wantErr: status.Error(codes.NotFound, "container no-such-container not found"),
		},
		{
			name:  "container-running",
			inCnt: "container-running",
			inSummaries: []imagetypes.Summary{
				imagetypes.Summary{
					RepoTags: []string{"container-running"},
				},
			},
			inCnts: []types.Container{
				types.Container{
					Image: "container-running",
				},
			},
			wantErr: status.Error(codes.Unavailable, "container container-running is running; use force to override"),
		},
		{
			name:   "container-running-with-force",
			inCnt:  "container-running",
			inOpts: []options.Option{options.Force()},
			inSummaries: []imagetypes.Summary{
				imagetypes.Summary{
					RepoTags: []string{"container-running"},
				},
			},
			inCnts: []types.Container{
				types.Container{
					Image: "container-running",
				},
			},
			wantState: &fakeRemovingDocker{
				Name: "container-running",
			},
		},
		{
			name:  "container-remove",
			inCnt: "container-remove",
			inSummaries: []imagetypes.Summary{
				imagetypes.Summary{
					RepoTags: []string{"container-remove"},
				},
			},
			wantState: &fakeRemovingDocker{
				Name: "container-remove",
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

			if err := mgr.ContainerRemove(context.Background(), tc.inCnt, tc.inOpts...); err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ContainerRemove(%q, %+v) returned unexpected error(-want, got):\n %s", tc.inCnt, tc.inOpts, diff)
					}
					return
				}
				t.Errorf("ContainerRemove(%q, %+v) returned error: %v", tc.inCnt, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fpd, cmpopts.IgnoreUnexported(fakeRemovingDocker{})); diff != "" {
					t.Errorf("ContainerRemove(%q, %+v) returned diff(-want, +got):\n%s", tc.inCnt, tc.inOpts, diff)
				}
			}
		})
	}
}

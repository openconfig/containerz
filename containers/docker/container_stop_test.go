package docker

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

type fakeStoppingDocker struct {
	fakeDocker
	cnts []types.Container

	Instance       string
	Duration       int
	RemoveInstance string
}

func (f *fakeStoppingDocker) ContainerStop(ctx context.Context, container string, options container.StopOptions) error {
	f.Instance = container

	if options.Timeout != nil {
		f.Duration = *options.Timeout
	}
	return nil
}

func (f fakeStoppingDocker) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	return f.cnts, nil
}

func (f *fakeStoppingDocker) ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error {
	f.RemoveInstance = container
	return nil
}

func TestContainerStop(t *testing.T) {
	tests := []struct {
		name       string
		inTimeout  time.Duration
		inOpts     []options.Option
		inInstance string
		inCnts     []types.Container
		wantState  *fakeStoppingDocker
		wantErr    error
	}{
		{
			name:       "no-such-instance",
			inInstance: "no-such-instance",
			wantErr:    status.Errorf(codes.NotFound, "container no-such-instance was not found"),
		},
		{
			name:       "stop-no-force",
			inInstance: "stop-no-force",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"stop-no-force"},
				},
			},
			wantState: &fakeStoppingDocker{
				Instance: "stop-no-force",
				Duration: -1,
			},
		},
		{
			name:       "stop-with-force-no-duration",
			inInstance: "stop-with-force-no-duration",
			inOpts:     []options.Option{options.Force()},
			inCnts: []types.Container{
				types.Container{
					Names: []string{"stop-with-force-no-duration"},
				},
			},
			wantState: &fakeStoppingDocker{
				Instance: "stop-with-force-no-duration",
			},
		},
		{
			name:       "stop-with-force-and-duration",
			inInstance: "stop-with-force-and-duration",
			inTimeout:  1 * time.Minute,
			inOpts:     []options.Option{options.Force()},
			inCnts: []types.Container{
				types.Container{
					Names: []string{"stop-with-force-and-duration"},
				},
			},
			wantState: &fakeStoppingDocker{
				Instance: "stop-with-force-and-duration",
				Duration: maximumStopTimeout,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.inTimeout != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), tc.inTimeout)
				defer cancel()
			}
			fsd := &fakeStoppingDocker{
				cnts: tc.inCnts,
			}
			mgr := New(fsd)

			if err := mgr.ContainerStop(ctx, tc.inInstance, tc.inOpts...); err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ContainerStop(%q, %+v) returned unexpected error(-want, got):\n %s", tc.inInstance, tc.inOpts, diff)
					}
					return
				}
				t.Errorf("ContainerStop(%q, %+v) returned error: %v", tc.inInstance, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fsd, cmpopts.IgnoreUnexported(fakeStoppingDocker{})); diff != "" {
					t.Errorf("ContainerStop(%q, %+v) returned diff(-want, +got):\n%s", tc.inInstance, tc.inOpts, diff)
				}
			}
		})
	}
}

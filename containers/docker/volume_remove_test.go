package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/openconfig/containerz/containers"
)

type fakeVolumeRemovingDocker struct {
	fakeDocker
	Name  string
	Force bool
}

func (f *fakeVolumeRemovingDocker) VolumeRemove(_ context.Context, id string, force bool) error {
	f.Name = id
	f.Force = force
	return nil
}

func TestVolumeRemove(t *testing.T) {
	tests := []struct {
		name      string
		inOpts    []options.Option
		inName    string
		wantState *fakeVolumeRemovingDocker
	}{
		{
			name:   "simple-name",
			inName: "simple-name",
			wantState: &fakeVolumeRemovingDocker{
				Name:  "simple-name",
				Force: false,
			},
		},
		{
			name:   "simple-name-forced",
			inName: "simple-name",
			inOpts: []options.Option{options.Force()},
			wantState: &fakeVolumeRemovingDocker{
				Name:  "simple-name",
				Force: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fcd := &fakeVolumeRemovingDocker{}
			mgr := New(fcd)

			if err := mgr.VolumeRemove(ctx, tc.inName, tc.inOpts...); err != nil {
				t.Errorf("VolumeRemove(%s, %+v) returned error: %v", tc.inName, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fcd, cmpopts.IgnoreUnexported(fakeVolumeRemovingDocker{})); diff != "" {
					t.Errorf("VolumeRemove(%s, %+v) returned diff(-want, +got):\n%s", tc.inName, tc.inOpts, diff)
				}
			}
		})
	}
}

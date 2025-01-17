package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/docker/docker/api/types"
)

type fakeRemovingPluginDocker struct {
	fakeDocker

	instance string
	options  types.PluginRemoveOptions
}

func (f *fakeRemovingPluginDocker) PluginRemove(ctx context.Context, instance string, options types.PluginRemoveOptions) error {
	f.instance = instance
	f.options = options
	return nil
}

func TestPluginRemove(t *testing.T) {
	tests := []struct {
		name       string
		inInstance string
		wantState  *fakeRemovingPluginDocker
	}{
		{
			name:       "success",
			inInstance: "some-instance",
			wantState: &fakeRemovingPluginDocker{
				instance: "some-instance",
				options: types.PluginRemoveOptions{
					Force: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Manager{
				client: &fakeRemovingPluginDocker{},
			}
			if err := m.PluginRemove(ctx, tt.inInstance); err != nil {
				t.Fatalf("PluginRemove() failed: %v", err)
			}
			if diff := cmp.Diff(tt.wantState, m.client, cmp.AllowUnexported(fakeRemovingPluginDocker{})); diff != "" {
				t.Errorf("PluginRemove() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

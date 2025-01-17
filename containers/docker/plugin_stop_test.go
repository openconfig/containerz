package docker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/docker/docker/api/types"
)

type fakeStoppingPluginDocker struct {
	fakeDocker

	instance string
	options  types.PluginDisableOptions
}

func (f *fakeStoppingPluginDocker) PluginDisable(ctx context.Context, instance string, options types.PluginDisableOptions) error {
	f.instance = instance
	f.options = options
	return nil
}

func TestPluginStop(t *testing.T) {
	tests := []struct {
		name       string
		inInstance string
		wantState  *fakeStoppingPluginDocker
	}{
		{
			name:       "success",
			inInstance: "some-instance",
			wantState: &fakeStoppingPluginDocker{
				instance: "some-instance",
				options: types.PluginDisableOptions{
					Force: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Manager{
				client: &fakeStoppingPluginDocker{},
			}
			if err := m.PluginStop(ctx, tt.inInstance); err != nil {
				t.Fatalf("PluginStop() failed: %v", err)
			}
			if diff := cmp.Diff(tt.wantState, m.client, cmp.AllowUnexported(fakeStoppingPluginDocker{})); diff != "" {
				t.Errorf("PluginStop() returned diff (-want +got):\n%s", diff)
			}
		})
	}
}

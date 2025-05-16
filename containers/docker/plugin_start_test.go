package docker

import (
	"context"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
)

type fakePluginStartingDocker struct {
	fakeDocker
}


func (f *fakePluginStartingDocker) PluginCreate(ctx context.Context, createCtx io.Reader, options types.PluginCreateOptions) error {
	return nil
}

func (f *fakePluginStartingDocker) PluginEnable(ctx context.Context, name string, options types.PluginEnableOptions) error {
	return nil
}

func TestPluginStart(t *testing.T) {
	pluginLocation = "testdata/"
	tests := []struct {
		name       string
		inName     string
		inInstance string
		inConfig   string
		wantErr    bool
	}{
		{
			name:       "valid-plugin",
			inName:     "data",
			inInstance: "test-instance",
			inConfig:   "test-config",
		},
		{
			name:       "invalid-plugin",
			inName:     "no-such-plugin",
			inInstance: "test-instance",
			inConfig:   "test-config",
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stagingLocation = t.TempDir()
			ctx := context.Background()
			mgr := New(&fakePluginStartingDocker{})
			if err := mgr.PluginStart(ctx, tc.inName, tc.inInstance, tc.inConfig); err != nil {
				if tc.wantErr {
					return
				}
				t.Errorf("PluginStart(%q, %q, %q) returned error: %v", tc.inName, tc.inInstance, tc.inConfig, err)
			}
		})
	}
}

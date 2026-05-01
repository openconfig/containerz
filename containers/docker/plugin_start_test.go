package docker

import (
	"context"
	"fmt"
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
		withHook   startPluginHookFunc
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
		{
			name:       "valid-plugin-working-hook",
			inName:     "data",
			inInstance: "test-instance",
			inConfig:   "test-config",
			withHook: func(ctx context.Context,
				pluginReader io.ReadCloser) (io.ReadCloser, error) {
				return pluginReader, nil
			},
		},
		{
			name:       "valid-plugin-failing-hook",
			inName:     "data",
			inInstance: "test-instance",
			inConfig:   "test-config",
			withHook: func(ctx context.Context,
				pluginReader io.ReadCloser) (io.ReadCloser, error) {
				return nil, fmt.Errorf("failed hook")
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stagingLocation = t.TempDir()
			ctx := context.Background()
			var ranHook bool
			if tc.withHook != nil {
				ctx = NewContextWithStartPluginHook(ctx, func(ctx context.Context,
					pluginReader io.ReadCloser) (io.ReadCloser, error) {
					ranHook = true
					return tc.withHook(ctx, pluginReader)
				})
			}

			mgr := New(&fakePluginStartingDocker{})
			if err := mgr.PluginStart(ctx, tc.inName, tc.inInstance,
				tc.inConfig); (err != nil) != tc.wantErr {
				t.Errorf("PluginStart(%q, %q, %q) returned error: %v, want error=%t",
					tc.inName, tc.inInstance, tc.inConfig, err, tc.wantErr)
			}
			if (tc.withHook != nil) != ranHook {
				t.Errorf("failed to run start plugin hook")
			}
		})
	}
}

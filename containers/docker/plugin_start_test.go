package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type fakePluginStartingDocker struct {
	plugins types.PluginsListResponse
	fakeDocker
}

func (f *fakePluginStartingDocker) PluginList(ctx context.Context, filter filters.Args) (types.PluginsListResponse, error) {
	return f.plugins, nil
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
		name        string
		inName      string
		inInstance  string
		inConfig    string
		expectedErr error
		withHook    startPluginHookFunc
		PluginList  []*types.Plugin
	}{
		{
			name:       "valid-plugin",
			inName:     "data",
			inInstance: "test-instance",
			inConfig:   "test-config",
		},
		{
			name:       "missing-name",
			inInstance: "test-instance",
			inConfig:   "test-config",
			expectedErr: nonNilErr(t, func() error {
				return checkStartPluginRequest("", "test-instance", "test-config")
			}),
		},
		{
			name:       "missing-config",
			inName:     "test-name",
			inInstance: "test-instance",
			expectedErr: nonNilErr(t, func() error {
				return checkStartPluginRequest("test-name", "test-instance", "")
			}),
		},
		{
			name:        "missing-instance",
			inName:      "test-name",
			inConfig:    "test-config",
			expectedErr: fmt.Errorf("instance name must be specified"),
		},
		{
			name:       "invalid-plugin",
			inName:     "no-such-plugin",
			inInstance: "test-instance",
			inConfig:   "test-config",
			expectedErr: func() error {
				_, err := os.Open(filepath.Join(pluginLocation, "no-such-plugin.tar"))
				return fmt.Errorf("failed to open plugin tar: %w", err)
			}(),
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
			expectedErr: fmt.Errorf(
				"failed to run startup plugin hook with error failed hook"),
		},
		{
			name:       "existing-plugin-restarted",
			inInstance: "test-instance",
			PluginList: []*types.Plugin{{
				Name: "test-instance:latest",
			}},
		},
		{
			name:       "existing-plugin-bad-restart-request-config",
			inInstance: "test-instance",
			inConfig:   "config",
			PluginList: []*types.Plugin{{
				Name: "test-instance:latest",
			}},
			expectedErr: nonNilErr(t, func() error {
				return checkRestartPluginRequest("", "test-instance", "config")
			}),
		},
		{
			name:       "existing-plugin-bad-restart-request-name",
			inInstance: "test-instance",
			inName:     "name",
			PluginList: []*types.Plugin{{
				Name: "test-instance:latest",
			}},
			expectedErr: nonNilErr(t, func() error {
				return checkRestartPluginRequest("name", "test-instance", "")
			}),
		},
		{
			name:       "existing-plugin-bad-restart-request-name+config",
			inInstance: "test-instance",
			inName:     "name",
			inConfig:   "config",
			PluginList: []*types.Plugin{{
				Name: "test-instance:latest",
			}},
			expectedErr: nonNilErr(t, func() error {
				return checkRestartPluginRequest("name", "test-instance", "config")
			}),
		},
		{
			name:       "existing-plugin-running",
			inInstance: "test-instance",
			PluginList: []*types.Plugin{{
				Enabled: true,
				Name:    "test-instance:latest",
			}},
			expectedErr: fmt.Errorf(`plugin "test-instance" is already started`),
		},
		{
			name:        "no-instance-name",
			inName:      "data",
			inInstance:  "",
			expectedErr: fmt.Errorf("instance name must be specified"),
		},
		{
			name:       "multiple-matching-instances",
			inInstance: "test-instance",
			PluginList: []*types.Plugin{{
				Name: "test-instance:v1",
			}, {
				Name: "test-instance:v0",
			}},
			expectedErr: fmt.Errorf(
				`mutliple plugins found for instance name "test-instance"`),
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

			mgr := New(&fakePluginStartingDocker{
				plugins: tc.PluginList,
			})
			err := mgr.PluginStart(ctx, tc.inName, tc.inInstance,
				tc.inConfig)
			wantErr := tc.expectedErr != nil
			if (err != nil) != wantErr {
				t.Errorf("PluginStart(%q, %q, %q) returned error: %v, want error=%s",
					tc.inName, tc.inInstance, tc.inConfig, err, tc.expectedErr)
			}
			if wantErr && err.Error() != tc.expectedErr.Error() {
				t.Errorf("expected error %s, got %s", tc.expectedErr, err)
			}
			if (tc.withHook != nil) != ranHook {
				t.Errorf("failed to run start plugin hook")
			}
		})
	}
}

func nonNilErr(t *testing.T, f func() error) error {
	if err := f(); err != nil {
		return err
	}
	t.Fatal("want non-nil error")
	return nil
}

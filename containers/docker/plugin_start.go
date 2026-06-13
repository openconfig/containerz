package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pluginState int

const (
	_ pluginState = iota
	running
	stopped
	notPresent
)

var (
	// pluginLocation is the location where plugins are expected to be written to.
	pluginLocation = "/plugins"

	// stagingLocation is the location where plugins are extracted to before being imported.
	stagingLocation = "/staging"
)

const (

	// rootfsDir is the name of the directory where the rootfs of a plugin is extracted to. Docker
	// expects this directory to exist when importing a plugin.
	rootfsDir = "rootfs"
)

// PluginStart does the following:
//
// If the plugin is already present but stopped, and the request
// contains only the instance name of that plugin, then the plugin will be re-started.
//
// If the plugin is not present, then it loads the deployed plugin tarball
// (expected to be in /plugins) into the container runtime.
//
// The operations performed to load and enabled the deployed plugin are based on
// this [documentation](https://docs.docker.com/engine/extend/#developing-a-plugin).
// The process is as follows:
//  0. The plugin image was uploaded in a previous deploy operation.
//  1. Unpack the plugin in a scratch space. The image must be unpacked under a `rootfs` directory.
//  2. Write the provided configuration alongside the `rootfs` directory.
//  3. Tar up the result
//  4. Push the tarball to docker and enable the plugin.
func (m *Manager) PluginStart(ctx context.Context, name, instance, config string) error {

	currentState, err := m.getPluginState(ctx, instance)
	if err != nil {
		return err
	}
	switch currentState {
	case running:
		return fmt.Errorf("plugin %q is already started", instance)
	case notPresent:
		if err := checkStartPluginRequest(name, instance, config); err != nil {
			return err
		}
		if err := m.createPlugin(ctx, name, instance, config); err != nil {
			return err
		}
		return m.enablePlugin(ctx, instance)
	case stopped:
		if err := checkRestartPluginRequest(name, instance, config); err != nil {
			return err
		}
		return m.enablePlugin(ctx, instance)
	}
	return nil
}

func (m *Manager) enablePlugin(ctx context.Context, instance string) error {
	if err := m.client.PluginEnable(ctx, instance, types.PluginEnableOptions{}); err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}
	return nil
}

func (m *Manager) createPlugin(ctx context.Context, name, instance, config string) error {
	f, err := os.Open(filepath.Join(pluginLocation, fmt.Sprintf("%s.tar", name)))
	if err != nil {
		return fmt.Errorf("failed to open plugin tar: %w", err)
	}
	defer f.Close()

	extractLocation := filepath.Join(stagingLocation, name)
	if err := os.MkdirAll(filepath.Join(extractLocation, rootfsDir), 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory %s: %w", filepath.Join(pluginLocation, name), err)
	}
	defer os.RemoveAll(stagingLocation)

	if err := archive.Untar(f, filepath.Join(extractLocation, rootfsDir), &archive.TarOptions{
		NoLchown: true,
	}); err != nil {
		return fmt.Errorf("failed to untar plugin: %w", err)
	}

	if err := os.WriteFile(filepath.Join(extractLocation, "config.json"), []byte(config), 0666); err != nil {
		return fmt.Errorf("failed to write plugin config: %w", err)
	}

	createCtx, err := archive.TarWithOptions(extractLocation, &archive.TarOptions{
		Compression: 0,
	})
	if err != nil {
		return fmt.Errorf("failed to create plugin tar: %w", err)
	}

	if hook := startPluginHookFromContext(ctx); hook != nil {
		if createCtx, err = hook(ctx, createCtx); err != nil {
			return fmt.Errorf("failed to run startup plugin hook with error %s", err)
		}
	}

	if err := m.client.PluginCreate(ctx, createCtx, types.PluginCreateOptions{
		RepoName: instance,
	}); err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	return nil
}

func (m *Manager) getPluginState(ctx context.Context, instance string) (pluginState, error) {
	if instance == "" {
		return 0, fmt.Errorf("instance name must be specified")
	}
	plugins, err := m.listMatchingPlugins(ctx, instance)
	if err != nil {
		return 0, err
	}
	if len(plugins) == 0 {
		// plugin doesnt exist yet, we'll be starting it for the first time
		return notPresent, nil
	}
	if len(plugins) > 1 {
		return 0, fmt.Errorf("mutliple plugins found for instance name %q", instance)
	}
	pluginToRestart := plugins[0]
	if pluginToRestart.Enabled {
		// handling this is the callers responsibility, we just want to report the state.
		return running, nil
	}
	return stopped, nil
}

func checkStartPluginRequest(name, instance, config string) error {
	if instance == "" || name == "" || config == "" {
		return status.Errorf(codes.InvalidArgument,
			"plugin instance %q is not present."+
				" please provide the instance_name, name, config to start it."+
				" got instance_name=%[1]q, name=%q, config=%q", instance, name, config)
	}
	return nil
}

func checkRestartPluginRequest(name, instance, config string) error {
	if instance == "" || name != "" || config != "" {
		return status.Errorf(codes.InvalidArgument,
			"plugin instance %q is not enabled."+
				" please provide only the instance_name to restart it."+
				" got instance_name=%[1]q, name=%q, config=%q", instance, name, config)
	}
	return nil
}

package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"

	"github.com/docker/docker/api/types"
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

// PluginStart loads the deployed plugin tarball (expected to be in /plugins) into the container
// runtime.
//
// The operations performed here are based on this [documentation](https://docs.docker.com/engine/extend/#developing-a-plugin).
// The process is as follows:
//  0. The plugin image was uploaded in a previous deploy operation.
//  1. Unpack the plugin in a scratch space. The image must be unpacked under a `rootfs` directory.
//  2. Write the provided configuration alongside the `rootfs` directory.
//  3. Tar up the result
//  4. Push the tarball to docker and enable the plugin.
func (m *Manager) PluginStart(ctx context.Context, name, instance, config string) error {
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

	if err := m.client.PluginCreate(ctx, createCtx, types.PluginCreateOptions{
		RepoName: instance,
	}); err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	if err := m.client.PluginEnable(ctx, instance, types.PluginEnableOptions{}); err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	return nil
}

package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

// PluginRemove removes a plugin named `instance` from the target system.
func (m *Manager) PluginRemove(ctx context.Context, instance string) error {
	return m.client.PluginRemove(ctx, instance, types.PluginRemoveOptions{
		Force: true,
	})
}

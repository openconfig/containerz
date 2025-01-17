package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

// PluginStop stops a plugin named `instance`.
func (m *Manager) PluginStop(ctx context.Context, instance string) error {
	return m.client.PluginDisable(ctx, instance, types.PluginDisableOptions{
		Force: true,
	})
}

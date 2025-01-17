package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/openconfig/containerz/containers"

	cpb "github.com/openconfig/gnoi/containerz"
)

// PluginList lists all plugins on a target. If instance is not empty, it will return a plugin
// named `instance` if it exists.
func (m *Manager) PluginList(ctx context.Context, instance string, srv options.ListPluginStreamer) error {
	var args filters.Args
	if instance != "" {
		args = filters.NewArgs(filters.KeyValuePair{Key: "name", Value: instance})
	}

	resp, err := m.client.PluginList(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	for _, plugin := range resp {
		conf, err := json.MarshalIndent(plugin.Config, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal plugin config: %v", err)
		}
		if err := srv.Send(&cpb.ListPluginsResponse{
			Plugins: []*cpb.Plugin{
				&cpb.Plugin{
					Id:           plugin.ID,
					InstanceName: plugin.Name,
					Config:       string(conf),
				},
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

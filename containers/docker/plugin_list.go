package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/filters"

	cpb "github.com/openconfig/gnoi/containerz"
)

// PluginList lists all plugins on a target. If instance is not empty, it will return a plugin
// named `instance` if it exists.
func (m *Manager) PluginList(ctx context.Context, instance string) (*cpb.ListPluginsResponse, error) {
	var args filters.Args
	if instance != "" {
		args = filters.NewArgs(filters.KeyValuePair{Key: "name", Value: instance})
	}

	resp, err := m.client.PluginList(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	res := &cpb.ListPluginsResponse{}
	for _, plugin := range resp {
		conf, err := json.MarshalIndent(plugin.Config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("unable to marshal plugin config: %v", err)
		}

		res.Plugins = append(res.Plugins, &cpb.Plugin{
			Id:           plugin.ID,
			InstanceName: plugin.Name,
			Config:       string(conf),
		})
	}

	return res, nil
}

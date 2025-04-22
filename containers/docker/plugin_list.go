package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/filters"

	cpb "github.com/openconfig/gnoi/containerz"
)

// PluginList lists all plugins on a target. If instance is not empty, it will return a plugin
// named `instance` if it exists.
func (m *Manager) PluginList(ctx context.Context, instance string) (*cpb.ListPluginsResponse, error) {
	resp, err := m.client.PluginList(ctx, filters.Args{})
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	res := &cpb.ListPluginsResponse{}
	for _, plugin := range resp {
		if instance != "" {
			// plugin.Name will have format <name>:<tag>.
			// instance_name (from StartPluginRequest) does not have a tag,
			// so cut off only the name here.
			if pluginName, _, _ := strings.Cut(plugin.Name, ":"); pluginName != instance {
				continue
			}
		}
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

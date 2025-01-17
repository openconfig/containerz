package client

import (
	"context"

	cpb "github.com/openconfig/gnoi/containerz"
)

// ListPlugin lists the plugins present on the target.
func (c *Client) ListPlugin(ctx context.Context, instance string) ([]*cpb.Plugin, error) {
	resp, err := c.cli.ListPlugins(ctx, &cpb.ListPluginsRequest{
		InstanceName: instance,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetPlugins(), nil
}
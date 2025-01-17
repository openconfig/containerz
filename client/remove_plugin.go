package client

import (
	"context"

	cpb "github.com/openconfig/gnoi/containerz"
)

// RemovePlugin removes the requested plugin identified by instance.
func (c *Client) RemovePlugin(ctx context.Context, instance string) error {
	if _, err := c.cli.RemovePlugin(ctx, &cpb.RemovePluginRequest{
		InstanceName: instance,
	}); err != nil {
		return err
	}

	return nil
}

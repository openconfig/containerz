package client

import (
	"context"

	cpb "github.com/openconfig/gnoi/containerz"
)

// StopPlugin stops the requested plugin identified by instance.
func (c *Client) StopPlugin(ctx context.Context, instance string) error {
	_, err := c.cli.StopPlugin(ctx, &cpb.StopPluginRequest{
		InstanceName: instance,
	})
	if err != nil {
		return err
	}

	return nil
}

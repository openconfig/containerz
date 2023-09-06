package client

import (
	"context"

	cpb "github.com/openconfig/gnoi/containerz"
)

// Stop stops the requested instance. Stop can also force termination.
func (c *Client) Stop(ctx context.Context, instance string, force bool) error {
	if _, err := c.cli.Stop(ctx, &cpb.StopRequest{
		InstanceName: instance,
		Force:        force,
	}); err != nil {
		return err
	}

	return nil
}

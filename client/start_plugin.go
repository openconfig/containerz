package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	cpb "github.com/openconfig/gnoi/containerz"
)

// StartPlugin starts the requested plugin identified by instance.
func (c *Client) StartPlugin(ctx context.Context, name, instance, configFile string) error {
	buf, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if !json.Valid(buf) {
		return fmt.Errorf("invalid json: %w", err)
	}

	if _, err := c.cli.StartPlugin(ctx, &cpb.StartPluginRequest{
		InstanceName: instance,
		Name:         name,
		Config:       string(buf),
	}); err != nil {
		return err
	}

	return nil
}

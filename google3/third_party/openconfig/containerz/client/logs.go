package client

import (
	"context"
	"io"

	cpb "github.com/openconfig/gnoi/containerz"
)

// Logs retrieves the logs for a given container. It can optionally follow the logs as they
// are being produced.
func (c *Client) Logs(ctx context.Context, instance string, follow bool) (<-chan *LogMessage, error) {
	lcli, err := c.cli.Log(ctx, &cpb.LogRequest{
		InstanceName: instance,
		Follow:       follow,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan *LogMessage, 100)
	go func() {
		defer lcli.CloseSend()
		defer close(ch)

		for {
			msg, err := lcli.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}

				nonBlockingChannelSend(ctx, ch, &LogMessage{
					Error: err,
				})
				return
			}

			if nonBlockingChannelSend(ctx, ch, &LogMessage{
				Msg: msg.GetMsg(),
			}) {
				return
			}
		}
	}()

	return ch, nil
}

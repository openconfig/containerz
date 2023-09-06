package client

import (
	"context"
	"io"

	"k8s.io/klog/v2"
	cpb "github.com/openconfig/gnoi/containerz"
)

// List implements the client logic for listing the existing containers on the target system.
func (c *Client) List(ctx context.Context, all bool, limit int32, filter map[string][]string) (<-chan *ContainerInfo, error) {
	req := &cpb.ListRequest{
		All:    all,
		Limit:  limit,
		Filter: toFilter(filter),
	}

	dcli, err := c.cli.List(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *ContainerInfo, 100)
	go func() {
		defer dcli.CloseSend()
		defer close(ch)
		for {
			msg, err := dcli.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				nonBlockingChannelSend(ctx, ch, &ContainerInfo{
					Error: err,
				})
				return
			}

			if nonBlockingChannelSend(ctx, ch, &ContainerInfo{
				ID:        msg.GetId(),
				Name:      msg.GetName(),
				ImageName: msg.GetImageName(),
				State:     msg.GetStatus().String(),
			}) {
				klog.Warningf("operation cancelled; returning")
				return
			}
		}
	}()

	return ch, nil
}

func toFilter(m map[string][]string) *cpb.ListRequest_Filter {
	// TODO(alshabib) implement this when filter field becomes a repeated.
	return nil
}

package client

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cpb "github.com/openconfig/gnoi/containerz"
)

// UpdateContainer updates a running container with the provided configuration.
func (c *Client) UpdateContainer(ctx context.Context, instance string, image string, tag string, cmd string, async bool, opts ...StartOption) (string, error) {
	optionz := &startOptions{}
	for _, opt := range opts {
		opt(optionz)
	}

	portMappings, err := ports(optionz.ports)
	if err != nil {
		return "", err
	}

	envMappings, err := envs(optionz.envs)
	if err != nil {
		return "", err
	}

	volumeMappings, err := volumes(optionz.volumes)
	if err != nil {
		return "", err
	}

	// Create the UpdateContainerRequest
	req := &cpb.UpdateContainerRequest{
		InstanceName: instance,
		ImageName:    image,
		ImageTag:     tag,
		Async:        async,
		Params: &cpb.StartContainerRequest{
			ImageName:    image,
			Tag:          tag,
		        Cmd:          cmd,
			Ports:        portMappings,
			Environment:  envMappings,
			InstanceName: instance,
			Volumes:      volumeMappings,
		},
	}

	resp, err := c.cli.UpdateContainer(ctx, req)
	if err != nil {
		return "", err
	}

	switch resp.GetResponse().(type) {
	case *cpb.UpdateContainerResponse_UpdateOk:
		return resp.GetUpdateOk().GetInstanceName(), nil
	case *cpb.UpdateContainerResponse_UpdateError:
		return "", status.Errorf(codes.Internal, "failed to update container: %s", resp.GetUpdateError().GetDetails())
	default:
		return "", status.Error(codes.Unknown, "unknown container state")
	}
}

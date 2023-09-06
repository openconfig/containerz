package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cpb "github.com/openconfig/gnoi/containerz"
)

// Start starts a container with the provided configuration and returns its instance name if the
// operation succeeded or an error otherwise.
func (c *Client) Start(ctx context.Context, image string, tag string, cmd string, instance string, opts ...StartOption) (string, error) {
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

	req := &cpb.StartRequest{
		ImageName:    image,
		Tag:          tag,
		Cmd:          cmd,
		Ports:        portMappings,
		Environment:  envMappings,
		InstanceName: instance,
	}

	resp, err := c.cli.Start(ctx, req)
	if err != nil {
		return "", err
	}

	switch resp.GetResponse().(type) {
	case *cpb.StartResponse_StartOk:
		return resp.GetStartOk().GetInstanceName(), nil
	case *cpb.StartResponse_StartError:
		return "", status.Errorf(codes.Internal, "failed to start container: %s", resp.GetStartError().GetDetails())
	default:
		return "", status.Error(codes.Unknown, "unknown container state")
	}
}

func ports(ports []string) ([]*cpb.StartRequest_Port, error) {
	mapping := make([]*cpb.StartRequest_Port, 0, len(ports))
	for _, port := range ports {
		parts := strings.SplitN(port, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("port definition %s is invalid", port)
		}

		in, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		out, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		mapping = append(mapping, &cpb.StartRequest_Port{Internal: uint32(in), External: uint32(out)})
	}

	return mapping, nil
}

func envs(envs []string) (map[string]string, error) {
	mapping := make(map[string]string, len(envs))

	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("env definition %s is invalid", env)
		}
		mapping[parts[0]] = parts[1]
	}

	return mapping, nil
}

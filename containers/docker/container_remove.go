package docker

import (
	"context"
	"strings"

	options "github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/docker/docker/api/types/container"
)

// ContainerRemove removes an image provided it is not related to a running container. Otherwise,
// it returns an error.
func (m *Manager) ContainerRemove(ctx context.Context, cnt string, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)

	cnts, err := m.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return status.Errorf(codes.Internal, "unable to list containers: %v", err)
	}

	for _, c := range cnts {
		for _, name := range c.Names {
			strippedname := strings.Replace(name, "/", "", 1)
			if strippedname == cnt {
				if stringToStatus(c.Status) == cpb.ListContainerResponse_RUNNING && !optionz.Force {
					return status.Errorf(codes.FailedPrecondition, "container %s is running", cnt)
				}
				if err := m.client.ContainerRemove(ctx, cnt, container.RemoveOptions{
					Force: optionz.Force,
				}); err != nil {
					return status.Errorf(codes.Internal, "unable to remove container: %v", err)
				}
				return nil
			}
		}
	}

	return status.Errorf(codes.NotFound, "container %s not found", cnt)
}

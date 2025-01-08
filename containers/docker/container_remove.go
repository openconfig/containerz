package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/openconfig/containerz/containers"
)

// ContainerRemove removes an image provided it is not related to a running container. Otherwise,
// it returns an error.
func (m *Manager) ContainerRemove(ctx context.Context, cnt string, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)
	return m.client.ContainerRemove(ctx, cnt, container.RemoveOptions{
		Force: optionz.Force,
	})
}

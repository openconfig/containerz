package docker

import (
	"context"

	"github.com/openconfig/containerz/containers"
)

// ContainerRemove removes an image provided it is not related to a running container. Otherwise,
// it returns an error.
//
// Deprecated - use RemoveImage instead.
func (m *Manager) ContainerRemove(ctx context.Context, image, tag string, opts ...options.Option) error {
	return m.ImageRemove(ctx, image, tag, opts...)
}

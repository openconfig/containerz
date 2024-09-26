package docker

import (
	"context"
	"testing"

	"github.com/openconfig/containerz/containers"
)

func TestContainerRemove(t *testing.T) {
	caller := func(m *Manager, ctx context.Context, image, tag string, opts ...options.Option) error {
		return m.ContainerRemove(ctx, image, tag, opts...)
	}
	objectRemoveTestInlet(t, caller)
}

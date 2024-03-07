package docker

import (
	"context"

	"github.com/openconfig/containerz/containers"
)

// VolumeRemove removes a volume.
func (m *Manager) VolumeRemove(ctx context.Context, name string, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)
	return m.client.VolumeRemove(ctx, name, optionz.Force)
}

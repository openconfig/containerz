package server

import (
	"context"
	"fmt"

	cpb "github.com/openconfig/gnoi/containerz"
)

// RemovePlugin removes a plugin. If the plugin does not exist this operation is a no-op.
func (s *Server) RemovePlugin(ctx context.Context, request *cpb.RemovePluginRequest) (*cpb.RemovePluginResponse, error) {
	if err := s.mgr.PluginRemove(ctx, request.GetInstanceName()); err != nil {
		return nil, fmt.Errorf("unable to remove plugin: %w", err)
	}
	return &cpb.RemovePluginResponse{}, nil
}

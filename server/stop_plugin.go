package server

import (
	"context"
	"fmt"

	cpb "github.com/openconfig/gnoi/containerz"
)

// StopPlugin stops a plugin. If the plugin does not exist this operation is a no-op.
func (s *Server) StopPlugin(ctx context.Context, request *cpb.StopPluginRequest) (*cpb.StopPluginResponse, error) {
	if err := s.mgr.PluginStop(ctx, request.GetInstanceName()); err != nil {
		return nil, fmt.Errorf("unable to stop plugin: %w", err)
	}
	return &cpb.StopPluginResponse{}, nil
}

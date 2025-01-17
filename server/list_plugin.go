package server

import (
	"context"
	"fmt"

	cpb "github.com/openconfig/gnoi/containerz"
)

// ListPlugins lists plugins on the target.
func (s *Server) ListPlugins(ctx context.Context, request *cpb.ListPluginsRequest) (*cpb.ListPluginsResponse, error) {
	resp, err := s.mgr.PluginList(ctx, request.GetInstanceName())
	if err != nil {
		return nil, fmt.Errorf("unable to list plugins: %w", err)
	}

	return resp, nil
}

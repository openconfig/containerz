package server

import (
	"context"
	"fmt"

	cpb "github.com/openconfig/gnoi/containerz"
)

func (s *Server) StartPlugin(ctx context.Context, request *cpb.StartPluginRequest) (*cpb.StartPluginResponse, error) {
	if err := s.mgr.PluginStart(ctx, request.GetName(), request.GetInstanceName(), request.GetConfig()); err != nil {
		return nil, fmt.Errorf("unable to start plugin: %w", err)
	}
	return &cpb.StartPluginResponse{
		InstanceName: request.GetInstanceName(),
	}, nil
}

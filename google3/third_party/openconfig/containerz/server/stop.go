package server

import (
	"context"

	"google3/third_party/openconfig/containerz/containers/options"
	cpb "github.com/openconfig/gnoi/containerz"
)

// Stop stops a container. If the container does not exist or is not running
// this operation returns an error. This operation can, optionally, force
// (i.e. kill) a container.
func (s *Server) Stop(ctx context.Context, request *cpb.StopRequest) (*cpb.StopResponse, error) {
	// TODO (alshabib): Consider adding a timeout to the request or use a containerz default.s
	opts := []options.ImageOption{}

	if request.GetForce() {
		opts = append(opts, options.Force())
	}

	if err := s.mgr.ContainerStop(ctx, request.GetInstanceName(), opts...); err != nil {
		return nil, err
	}
	return &cpb.StopResponse{}, nil
}

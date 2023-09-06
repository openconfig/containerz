package server

import (
	"context"

	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// Start starts a container. If the image does not exist on the target,
// Start returns an error. A started container is identified by an instance
// name, which  can optionally be supplied by the caller otherwise the target
// should provide one. If the instance name already exists, the target should
// return an error.
func (s *Server) Start(ctx context.Context, request *cpb.StartRequest) (*cpb.StartResponse, error) {
	opts := []options.ImageOption{}

	if request.GetPorts() != nil {
		ports := make(map[uint32]uint32, len(request.GetPorts()))
		for _, port := range request.GetPorts() {
			ports[port.GetInternal()] = port.GetExternal()
		}
		opts = append(opts, options.WithPorts(ports))
	}

	opts = append(opts, options.WithEnv(request.GetEnvironment()), options.WithInstanceName(request.GetInstanceName()))

	resp, err := s.mgr.ContainerStart(ctx, request.GetImageName(), request.GetTag(), request.GetCmd(), opts...)
	if err != nil {
		return nil, err
	}

	return &cpb.StartResponse{
		Response: &cpb.StartResponse_StartOk{
			StartOk: &cpb.StartOK{
				InstanceName: resp,
			},
		},
	}, nil
}

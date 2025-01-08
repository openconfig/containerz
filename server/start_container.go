// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"

	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// StartContainer starts a container. If the image does not exist on the target,
// Start returns an error. A started container is identified by an instance
// name, which  can optionally be supplied by the caller otherwise the target
// should provide one. If the instance name already exists, the target should
// return an error.
func (s *Server) StartContainer(ctx context.Context, request *cpb.StartContainerRequest) (*cpb.StartContainerResponse, error) {
	opts := optionsFromStartContainerRequest(request)
	resp, err := s.mgr.ContainerStart(ctx, request.GetImageName(), request.GetTag(), request.GetCmd(), opts...)
	if err != nil {
		return nil, err
	}

	return &cpb.StartContainerResponse{
		Response: &cpb.StartContainerResponse_StartOk{
			StartOk: &cpb.StartOK{
				InstanceName: resp,
			},
		},
	}, nil
}

func optionsFromStartContainerRequest(request *cpb.StartContainerRequest) []options.Option {
	var opts []options.Option
	if len(request.GetPorts()) != 0 {
		ports := make(map[uint32]uint32, len(request.GetPorts()))
		for _, port := range request.GetPorts() {
			ports[port.GetInternal()] = port.GetExternal()
		}
		opts = append(opts, options.WithPorts(ports))
	}
	if request.GetNetwork() != "" {
		opts = append(opts, options.WithNetwork(request.GetNetwork()))
	}
	if request.GetRestart() != nil {
		opts = append(opts, options.WithRestartPolicy(request.GetRestart()))
	}
	if request.GetRunAs() != nil {
		opts = append(opts, options.WithRunAs(request.GetRunAs()))
	}
	if request.GetCap() != nil {
		opts = append(opts, options.WithCapabilities(request.GetCap()))
	}
	if request.GetLimits() != nil {
		if request.GetLimits().GetMaxCpu() != 0 {
			opts = append(opts, options.WithCPUs(request.GetLimits().GetMaxCpu()))
		}
		if request.GetLimits().GetSoftMemBytes() != 0 {
			opts = append(opts, options.WithSoftLimit(request.GetLimits().GetSoftMemBytes()))
		}
		if request.GetLimits().GetHardMemBytes() != 0 {
			opts = append(opts, options.WithHardLimit(request.GetLimits().GetHardMemBytes()))
		}
	}

	opts = append(opts, options.WithLabels(request.GetLabels()), options.WithEnv(request.GetEnvironment()), options.WithInstanceName(request.GetInstanceName()), options.WithVolumes(request.GetVolumes()))
	return opts
}

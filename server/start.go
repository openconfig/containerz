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

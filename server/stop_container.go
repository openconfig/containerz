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
	"fmt"

	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// StopContainer stops a container. If the container does not exist or is not running
// this operation returns an error. This operation can, optionally, force
// (i.e. kill) a container.
func (s *Server) StopContainer(ctx context.Context, request *cpb.StopContainerRequest) (*cpb.StopContainerResponse, error) {
	// TODO (alshabib): Consider adding a timeout to the request or use a containerz default.s
	opts := []options.Option{}

	if request.GetForce() {
		opts = append(opts, options.Force())
	}

	if err := s.mgr.ContainerStop(ctx, request.GetInstanceName(), opts...); err != nil {
		return nil, err
	}
	return &cpb.StopContainerResponse{
		Code: cpb.StopContainerResponse_SUCCESS,
		Details: fmt.Sprintf("stopped %q", request.GetInstanceName()),
	}, nil
}

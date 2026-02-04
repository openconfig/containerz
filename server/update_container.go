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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	cpb "github.com/openconfig/gnoi/containerz"
)

// UpdateContainer updates a running container to the image specified in the
// request. By default the operation is synchronous which means that the
// request will only return once the container has either been successfully
// updated or the update has failed. If the client requests an asynchronous
// update then the server must perform all validations (e.g. does the
// requested image exist on the system or does the instance name exist) and
// return to the client and the update happens asynchronously. It is up to the
// client to check if the update actually updates the container to the
// requested version or not.
// In both synchronous and asynchronous mode, the update process is a
// break-before-make process as resources bound to the old container must be
// released prior to launching the new container.
// If the update fails, the server must restore the previous version of the
// container. This can either be a start of the previous container or by
// starting a new container with the old image.
// It must use the provided StartContainerRequest provided in the
// params field.
// If a container exists but is not running should still upgrade the container
// and start it.
// The client should only depend on the client being restarted. Any ephemeral
// state (date written to memory or the filesystem) cannot be depended upon.
// In particular, the contents of the filesystem are not guaranteed during a
// rollback.
func (s *Server) UpdateContainer(ctx context.Context, request *cpb.UpdateContainerRequest) (*cpb.UpdateContainerResponse, error) {
	startReq := request.GetParams()
	if startReq == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "expected request to contain populated params, yet was nil")
	}

	opts, err := optionsFromStartContainerRequest(startReq)
	if err != nil {
		return nil, err
	}
	instance, err := s.mgr.ContainerUpdate(ctx, request.GetInstanceName(), startReq.GetImageName(), startReq.GetTag(), startReq.GetCmd(), request.GetAsync(), opts...)
	if err != nil {
		return nil, err
	}

	return &cpb.UpdateContainerResponse{
		Response: &cpb.UpdateContainerResponse_UpdateOk{
			UpdateOk: &cpb.UpdateOK{
				InstanceName: instance,
				IsAsync:      request.GetAsync(),
			},
		},
	}, nil
}

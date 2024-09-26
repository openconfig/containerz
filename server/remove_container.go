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

	cpb "github.com/openconfig/gnoi/containerz"
)

// RemoveContainer deletes images that match the spec defined in the
// request. If the image is associated to a running container then an error
// is returned. If the specified container image does not exist, this
// operation is a no-op.
//
// Deprecated - use RemoveImage instead.
func (s *Server) RemoveContainer(ctx context.Context, request *cpb.RemoveContainerRequest) (*cpb.RemoveContainerResponse, error) {
	req := &cpb.RemoveImageRequest{
		Name:  request.GetName(),
		Tag:   request.GetTag(),
		Force: request.GetForce(),
	}

	imgResp, err := s.RemoveImage(ctx, req)
	if err != nil {
		return nil, err
	}

	var code cpb.RemoveContainerResponse_Code
	switch imgResp.GetCode() {
	case cpb.RemoveImageResponse_SUCCESS:
		code = cpb.RemoveContainerResponse_SUCCESS
	case cpb.RemoveImageResponse_NOT_FOUND:
		code = cpb.RemoveContainerResponse_NOT_FOUND
	case cpb.RemoveImageResponse_RUNNING:
		code = cpb.RemoveContainerResponse_RUNNING
	default:
		code = cpb.RemoveContainerResponse_UNSPECIFIED
	}

	return &cpb.RemoveContainerResponse{
		Code:   code,
		Detail: imgResp.GetDetail(),
	}, nil
}

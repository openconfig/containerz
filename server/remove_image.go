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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RemoveImage deletes containers that match the spec defined in the request. If
// the specified container does not exist, this operation is a no-op.
func (s *Server) RemoveImage(ctx context.Context, request *cpb.RemoveImageRequest) (*cpb.RemoveImageResponse, error) {

	var opts []options.Option
	if request.GetForce() {
		opts = append(opts, options.Force())
	}

	if err := s.mgr.ImageRemove(ctx, request.GetName(), request.GetTag(), opts...); err != nil {
		if _, ok := status.FromError(err); !ok {
			return nil, status.Errorf(codes.Internal,
				"unknown containerz state: %v", err)
		}
		return nil, err
	}

	return &cpb.RemoveImageResponse{}, nil
}

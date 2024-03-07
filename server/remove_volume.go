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

// RemoveVolume removes a volume.
func (s *Server) RemoveVolume(ctx context.Context, request *cpb.RemoveVolumeRequest) (*cpb.RemoveVolumeResponse, error) {
	var opts []options.Option
	if request.GetForce() {
		opts = append(opts, options.Force())
	}
	if err := s.mgr.VolumeRemove(ctx, request.GetName(), opts...); err != nil {
		return nil, err
	}

	return &cpb.RemoveVolumeResponse{}, nil
}

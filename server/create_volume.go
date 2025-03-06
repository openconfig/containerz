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

// CreateVolume creates a volume. If the volume already exists, this operation returns an
// error. A volume is expected to be backed by persistent datastore and is
// expected exist across device reboots along with the data it contained.
func (s *Server) CreateVolume(ctx context.Context, request *cpb.CreateVolumeRequest) (*cpb.CreateVolumeResponse, error) {
	opts := []options.Option{options.WithVolumeLabels(request.GetLabels())}

	name, driver := request.GetName(), request.GetDriver()

	switch driver {
	case cpb.Driver_DS_UNSPECIFIED:
		driver = cpb.Driver_DS_LOCAL
	case cpb.Driver_DS_LOCAL:
		driver = cpb.Driver_DS_LOCAL
		opts = append(opts, options.WithVolumeDriverOpts(request.GetLocalMountOptions()))
	case cpb.Driver_DS_CUSTOM:
		driver = cpb.Driver_DS_CUSTOM
		opts = append(opts, options.WithVolumeDriverOpts(request.GetCustomOptions()))
	}

	resp, err := s.mgr.VolumeCreate(ctx, name, driver, opts...)
	if err != nil {
		return nil, err
	}

	return &cpb.CreateVolumeResponse{
		Name: resp,
	}, nil
}

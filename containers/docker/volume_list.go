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

package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	tpb "google.golang.org/protobuf/types/known/timestamppb"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/filters/filters"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/volume/volume"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// VolumeList lists the volumes present on the target.
func (m *Manager) VolumeList(ctx context.Context, srv options.ListVolumeStreamer, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)
	kvPairs := []filters.KeyValuePair{}
	for key, values := range optionz.Filter {
		for _, value := range values {
			kvPairs = append(kvPairs, filters.KeyValuePair{Key: string(key), Value: value})
		}
	}

	volOpts := volume.ListOptions{
		Filters: filters.NewArgs(kvPairs...),
	}

	resp, err := m.client.VolumeList(ctx, volOpts)
	if err != nil {
		return err
	}

	for _, vol := range resp.Volumes {
		t, err := time.Parse(time.RFC3339, vol.CreatedAt)
		if err != nil {
			return fmt.Errorf("unable to parse creation time: %v", err)
		}
		if err := srv.Send(&cpb.ListVolumeResponse{
			Name:    vol.Name,
			Created: tpb.New(t),
			Driver:  vol.Driver,
			Options: vol.Options,
			Labels:  vol.Labels,
		}); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

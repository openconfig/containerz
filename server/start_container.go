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

	options "github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	locationLabel = "com.google.containerz.location"
)

// StartContainer starts a container. If the image does not exist on the target,
// Start returns an error. A started container is identified by an instance
// name, which  can optionally be supplied by the caller otherwise the target
// should provide one. If the instance name already exists, the target should
// return an error.
func (s *Server) StartContainer(ctx context.Context, request *cpb.StartContainerRequest) (*cpb.StartContainerResponse, error) {
	opts, err := optionsFromStartContainerRequest(request)
	if err != nil {
		return nil, err
	}
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

func optionsFromStartContainerRequest(request *cpb.StartContainerRequest) ([]options.Option, error) {
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

	labels, err := labelsWithLocation(request)
	if err != nil {
		return nil, err
	}

	opts = append(opts, options.WithLabels(labels), options.WithEnv(request.GetEnvironment()), options.WithInstanceName(request.GetInstanceName()), options.WithVolumes(request.GetVolumes()), options.WithDevices(request.GetDevices()))
	return opts, nil
}

// labelsWithLocation updates the labels map to include the location, based on the location
// field in the request. L_UNKNOWN is treated as L_PRIMARY
func labelsWithLocation(request *cpb.StartContainerRequest) (map[string]string, error) {
	location := request.GetLocation()
	if location == cpb.StartContainerRequest_L_UNKNOWN {
		location = cpb.StartContainerRequest_L_PRIMARY
	}
	locationStr := cpb.StartContainerRequest_Location_name[int32(location)]
	labels := request.GetLabels()
	// if the label is already set but has the same value as the location, then ignore it.
	if requestedLocation, ok := labels[locationLabel]; ok && requestedLocation != locationStr {
		return nil, status.Errorf(codes.InvalidArgument,
			"%q label (currently set to %q) should be not be set, or should match"+
				" location field %q. Unspecified location field is treated as L_PRIMARY",
			locationLabel, requestedLocation, locationStr)
	} else if !ok {
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[locationLabel] = locationStr
	}
	return labels, nil
}

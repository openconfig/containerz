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

package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cpb "github.com/openconfig/gnoi/containerz"
)

// StartContainer starts a container with the provided configuration and returns its instance name if the
// operation succeeded or an error otherwise.
func (c *Client) StartContainer(ctx context.Context, image string, tag string, cmd string, instance string, opts ...StartOption) (string, error) {
	optionz := &startOptions{}
	for _, opt := range opts {
		opt(optionz)
	}

	portMappings, err := ports(optionz.ports)
	if err != nil {
		return "", err
	}

	envMappings, err := envs(optionz.envs)
	if err != nil {
		return "", err
	}

	volumeMappings, err := volumes(optionz.volumes)
	if err != nil {
		return "", err
	}

	req := &cpb.StartContainerRequest{
		ImageName:    image,
		Tag:          tag,
		Cmd:          cmd,
		Ports:        portMappings,
		Environment:  envMappings,
		InstanceName: instance,
		Volumes:      volumeMappings,
	}

	resp, err := c.cli.StartContainer(ctx, req)
	if err != nil {
		return "", err
	}

	switch resp.GetResponse().(type) {
	case *cpb.StartContainerResponse_StartOk:
		return resp.GetStartOk().GetInstanceName(), nil
	case *cpb.StartContainerResponse_StartError:
                errorCode := resp.GetStartError().GetErrorCode().String()
                return "", status.Errorf(codes.Internal, "Failed to start container: %s (Error Code: %s)", resp.GetStartError().GetDetails(), errorCode)
		//return "", status.Errorf(resp.GetStartError().GetErrorCode(), "failed to start container: %s", resp.GetStartError().GetDetails())
	default:
		return "", status.Error(codes.Unknown, "unknown container state")
	}
}

func ports(ports []string) ([]*cpb.StartContainerRequest_Port, error) {
	mapping := make([]*cpb.StartContainerRequest_Port, 0, len(ports))
	for _, port := range ports {
		parts := strings.SplitN(port, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("port definition %s is invalid", port)
		}

		in, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		out, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		mapping = append(mapping, &cpb.StartContainerRequest_Port{Internal: uint32(in), External: uint32(out)})
	}

	return mapping, nil
}

func envs(envs []string) (map[string]string, error) {
	mapping := make(map[string]string, len(envs))

	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("env definition %s is invalid", env)
		}
		mapping[parts[0]] = parts[1]
	}

	return mapping, nil
}

func volumes(volumes []string) ([]*cpb.Volume, error) {
	vols := make([]*cpb.Volume, 0, len(volumes))

	for _, volume := range volumes {
		parts := strings.SplitN(volume, ":", 3)
		switch len(parts) {
		case 2:
			vols = append(vols, &cpb.Volume{
				Name:       parts[0],
				MountPoint: parts[1],
			})
		case 3:
			vols = append(vols, &cpb.Volume{
				Name:       parts[0],
				MountPoint: parts[1],
				ReadOnly:   parts[2] == "ro",
			})
		default:
			return nil, fmt.Errorf("volume definition %s is invalid", volume)
		}
	}

	return vols, nil
}

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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cpb "github.com/openconfig/gnoi/containerz"
)

// UpdateContainer updates an existing container with the provided configuration and returns its
// instance name if the operation succeeded or an error otherwise.
func (c *Client) UpdateContainer(ctx context.Context, image string, tag string, cmd string, instance string, async bool, opts ...StartOption) (string, error) {

	// First get the ContainerStartRequest.
	reqInlet, err := startContainerRequestWithOptions(ctx, image, tag, cmd, instance, opts...)
	if err != nil {
		return "", err
	}

	// Then wrap it inside a UpdateContainerRequest.
	req := &cpb.UpdateContainerRequest{
		InstanceName: instance,
		ImageName:    image,
		ImageTag:     tag,
		Params:       reqInlet,
		Async:        async,
	}

	resp, err := c.cli.UpdateContainer(ctx, req)
	if err != nil {
		return "", err
	}

	switch resp.GetResponse().(type) {
	case *cpb.UpdateContainerResponse_UpdateOk:
		return resp.GetUpdateOk().GetInstanceName(), nil
	case *cpb.UpdateContainerResponse_UpdateError:
		return "", status.Errorf(codes.Unknown,
			"got UpdateError returned - please update containerz server")
	default:
		return "", status.Error(codes.Unknown, "unknown container state")
	}
}

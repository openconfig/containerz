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

var (

	// ErrNotFound indicates that the specified image was not found on the target system.
	ErrNotFound = status.Error(codes.NotFound, "resource was not found")

	// ErrRunning indicates that there is a container running this image.
	ErrRunning = status.Error(codes.FailedPrecondition, "resource is running")
)

// RemoveContainer removes an image from the target system. It returns nil upon success. Otherwise it
// returns an error indicating whether the image was not found or is associated to running
// container.
//
// Deprecated -- Use ImageRemove instead.
func (c *Client) RemoveContainer(ctx context.Context, cnt string, forceopt ...bool) error {
	force := false
	if len(forceopt) > 0 {
		force = forceopt[0]
	}

	resp, err := c.cli.RemoveContainer(ctx, &cpb.RemoveContainerRequest{
		Name:  cnt,
		Force: force,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "unable to remove container: %v", err)
	}

	switch resp.GetCode() {
	case cpb.RemoveContainerResponse_SUCCESS:
		return nil
	case cpb.RemoveContainerResponse_NOT_FOUND:
		return ErrNotFound
	case cpb.RemoveContainerResponse_RUNNING:
		return ErrRunning
	default:
		return status.Errorf(codes.Unknown, "unknown error: %v", resp.GetCode())
	}
}

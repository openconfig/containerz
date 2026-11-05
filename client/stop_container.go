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

	cpb "github.com/openconfig/gnoi/containerz"
)

// stopOptions contains the optional parameters of a stop container call.
type stopOptions struct {
	restart bool
}

// StopOption is an option passed to a stop container call.
type StopOption func(*stopOptions)

// WithRestart requests that the target starts the container again immediately
// after stopping it.
func WithRestart() StopOption {
	return func(opt *stopOptions) {
		opt.restart = true
	}
}

// StopContainer stops the requested instance. Stop can also force termination
// and restart the container after it has stopped.
func (c *Client) StopContainer(ctx context.Context, instance string, force bool, opts ...StopOption) error {
	optionz := &stopOptions{}
	for _, opt := range opts {
		opt(optionz)
	}

	if _, err := c.cli.StopContainer(ctx, &cpb.StopContainerRequest{
		InstanceName: instance,
		Force:        force,
		Restart:      optionz.restart,
	}); err != nil {
		return err
	}

	return nil
}

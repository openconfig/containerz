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
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeStoppingContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.StopRequest
}

func (f *fakeStoppingContainerzServer) Stop(ctx context.Context, req *cpb.StopRequest) (*cpb.StopResponse, error) {
	f.receivedMsg = req
	return &cpb.StopResponse{}, nil
}

func TestStop(t *testing.T) {
	tests := []struct {
		name       string
		inInstance string
		inForce    bool

		wantMsg *cpb.StopRequest
	}{
		{
			name:       "simple",
			inInstance: "some-instance",
			inForce:    true,
			wantMsg: &cpb.StopRequest{
				InstanceName: "some-instance",
				Force:        true,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeStoppingContainerzServer{}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			if err := cli.Stop(ctx, tc.inInstance, tc.inForce); err != nil {
				t.Fatalf("Stop(%q, %t) returned an unexpected error: %v", tc.inInstance, tc.inForce, err)
			}

			if diff := cmp.Diff(fcm.receivedMsg, tc.wantMsg, protocmp.Transform()); diff != "" {
				t.Errorf("Stop(%q, %t) returned diff(-want, +got):\n%s", tc.inInstance, tc.inForce, diff)
			}
		})
	}
}

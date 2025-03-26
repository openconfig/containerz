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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestStop(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.StopContainerRequest
		inOpts    []Option
		wantResp  *cpb.StopContainerResponse
		wantState *fakeContainerManager
	}{
		{
			name: "no-force",
			inReq: &cpb.StopContainerRequest{
				InstanceName: "some-name",
			},
			wantResp: &cpb.StopContainerResponse{
				Code:    cpb.StopContainerResponse_SUCCESS,
				Details: `stopped "some-name"`,
			},
			wantState: &fakeContainerManager{
				Instance: "some-name",
				Force:    false,
			},
		},
		{
			name: "nforce",
			inReq: &cpb.StopContainerRequest{
				InstanceName: "some-name",
				Force:        true,
			},
			wantResp: &cpb.StopContainerResponse{
				Code:    cpb.StopContainerResponse_SUCCESS,
				Details: `stopped "some-name"`,
			},
			wantState: &fakeContainerManager{
				Instance: "some-name",
				Force:    true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{}
			tc.inOpts = append(tc.inOpts, WithAddr("localhost:0"))
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.StopContainer(ctx, tc.inReq)
			if err != nil {
				t.Errorf("Stop(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Stop(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}

			if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Stop(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

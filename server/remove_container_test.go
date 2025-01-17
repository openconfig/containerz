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

func TestContainerRemove(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.RemoveContainerRequest
		inOpts    []Option
		inCnts    []*cpb.ListContainerResponse
		wantError bool
		wantResp  *cpb.RemoveContainerResponse
		wantState *fakeContainerManager
	}{
		{
			name: "success",
			inCnts: []*cpb.ListContainerResponse{
				&cpb.ListContainerResponse{
					Name: "test",
				},
			},
			inReq: &cpb.RemoveContainerRequest{
				Name: "test",
			},
			wantResp: &cpb.RemoveContainerResponse{},
			wantState: &fakeContainerManager{
				Instance: "test",
			},
		},
		{
			name: "success-with-force",
			inCnts: []*cpb.ListContainerResponse{
				&cpb.ListContainerResponse{
					Name: "test",
				},
			},
			inReq: &cpb.RemoveContainerRequest{
				Name:  "test",
				Force: true,
			},
			wantResp: &cpb.RemoveContainerResponse{},
			wantState: &fakeContainerManager{
				Instance: "test",
				Force:    true,
			},
		},
		{
			name: "not-found",
			inReq: &cpb.RemoveContainerRequest{
				Name: "test",
			},
			wantError: true,
			wantResp:  &cpb.RemoveContainerResponse{},
			wantState: &fakeContainerManager{
				Instance: "test",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{
				listCntMsgs: tc.inCnts,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.RemoveContainer(ctx, tc.inReq)
			if err != nil {
				if tc.wantError {
					return
				}
				t.Errorf("Remove(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Remove(%+v) returned unexpected diff (-want +got):\n%s", tc.inReq, diff)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("List(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}

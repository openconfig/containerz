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

func TestContainerStart(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.StartRequest
		inOpts    []Option
		wantResp  *cpb.StartResponse
		wantState *fakeContainerManager
	}{
		{
			name: "simple",
			inReq: &cpb.StartRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
			},
			wantResp: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
			},
		},
		{
			name: "ports",
			inReq: &cpb.StartRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Ports: []*cpb.StartRequest_Port{
					&cpb.StartRequest_Port{
						Internal: 1,
						External: 2,
					},
					&cpb.StartRequest_Port{
						Internal: 3,
						External: 4,
					},
				},
			},
			wantResp: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Ports: map[uint32]uint32{1: 2, 3: 4},
			},
		},
		{
			name: "env+port+instance",
			inReq: &cpb.StartRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Ports: []*cpb.StartRequest_Port{
					&cpb.StartRequest_Port{
						Internal: 1,
						External: 2,
					},
					&cpb.StartRequest_Port{
						Internal: 3,
						External: 4,
					},
				},
				Environment:  map[string]string{"1": "2", "3": "4"},
				InstanceName: "some-instance",
			},
			wantResp: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Envs:     map[string]string{"1": "2", "3": "4"},
				Ports:    map[uint32]uint32{1: 2, 3: 4},
				Instance: "some-instance",
			},
		},
		{
			name: "env",
			inReq: &cpb.StartRequest{
				ImageName:   "some-image",
				Tag:         "some-tag",
				Cmd:         "some-cmd",
				Environment: map[string]string{"1": "2", "3": "4"},
			},
			wantResp: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Envs:  map[string]string{"1": "2", "3": "4"},
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

			resp, err := cli.Start(ctx, tc.inReq)
			if err != nil {
				t.Errorf("Start(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}

			if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

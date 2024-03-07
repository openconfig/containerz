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
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestListContainer(t *testing.T) {
	tests := []struct {
		name      string
		inErr     error
		inCnts    []*cpb.ListContainerResponse
		inReq     *cpb.ListContainerRequest
		inOpts    []Option
		wantResp  []*cpb.ListContainerResponse
		wantState *fakeContainerManager
	}{
		{
			name: "no-containers",
			inReq: &cpb.ListContainerRequest{
				All:   true,
				Limit: 10,
			},
			wantState: &fakeContainerManager{
				All:   true,
				Limit: 10,
			},
			wantResp: []*cpb.ListContainerResponse{},
		},
		{
			name: "containers",
			inReq: &cpb.ListContainerRequest{
				All:   true,
				Limit: 10,
			},
			inCnts: []*cpb.ListContainerResponse{
				&cpb.ListContainerResponse{
					Id: "some-id",
				},
				&cpb.ListContainerResponse{
					Id: "other-id",
				},
			},
			wantState: &fakeContainerManager{
				All:   true,
				Limit: 10,
			},
			wantResp: []*cpb.ListContainerResponse{
				&cpb.ListContainerResponse{
					Id: "some-id",
				},
				&cpb.ListContainerResponse{
					Id: "other-id",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{
				listMsgs: tc.inCnts,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			lCli, err := cli.ListContainer(ctx, tc.inReq)
			if err != nil {
				t.Errorf("List(%+v) returned error: %v", tc.inReq, err)
			}

			gotResp := []*cpb.ListContainerResponse{}
			for {
				msg, err := lCli.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Errorf("Recv() returned error: %v", err)
				}
				gotResp = append(gotResp, msg)
			}

			if diff := cmp.Diff(tc.wantResp, gotResp, protocmp.Transform()); diff != "" {
				t.Errorf("List(%+v) returned unexpected diff (-want +got):\n%s", tc.inReq, diff)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("List(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}

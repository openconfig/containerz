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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

func TestContainerRemove(t *testing.T) {
	tests := []struct {
		name     string
		inErr    error
		inReq    *cpb.RemoveRequest
		inOpts   []Option
		wantResp *cpb.RemoveResponse
	}{
		{
			name:  "success",
			inReq: &cpb.RemoveRequest{},
			wantResp: &cpb.RemoveResponse{
				Code: cpb.RemoveResponse_SUCCESS,
			},
		},
		{
			name:  "not-found",
			inReq: &cpb.RemoveRequest{},
			inErr: status.Error(codes.NotFound, "image not found"),
			wantResp: &cpb.RemoveResponse{
				Code:   cpb.RemoveResponse_NOT_FOUND,
				Detail: "image not found",
			},
		},
		{
			name:  "running-container",
			inReq: &cpb.RemoveRequest{},
			inErr: status.Error(codes.Unavailable, "container running"),
			wantResp: &cpb.RemoveResponse{
				Code:   cpb.RemoveResponse_RUNNING,
				Detail: "container running",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{
				removeError: tc.inErr,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.Remove(ctx, tc.inReq)
			if err != nil {
				t.Errorf("Remove(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Remove(%+v) returned unexpected diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

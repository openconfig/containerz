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
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeRemovingContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.RemoveRequest
	sendMsg     *cpb.RemoveResponse
}

func (f *fakeRemovingContainerzServer) Remove(_ context.Context, req *cpb.RemoveRequest) (*cpb.RemoveResponse, error) {
	f.receivedMsg = req
	return f.sendMsg, nil
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name    string
		inImage string
		inTag   string
		inMsg   *cpb.RemoveResponse

		wantMsg *cpb.RemoveRequest
		wantErr error
	}{
		{
			name: "success",
			inMsg: &cpb.RemoveResponse{
				Code: cpb.RemoveResponse_SUCCESS,
			},
			inImage: "some-image",
			inTag:   "some-tag",
			wantMsg: &cpb.RemoveRequest{
				Name: "some-image",
				Tag:  "some-tag",
			},
		},
		{
			name: "not-found",
			inMsg: &cpb.RemoveResponse{
				Code: cpb.RemoveResponse_NOT_FOUND,
			},
			inImage: "some-image",
			inTag:   "some-tag",
			wantMsg: &cpb.RemoveRequest{
				Name: "some-image",
				Tag:  "some-tag",
			},
			wantErr: ErrNotFound,
		},
		{
			name: "running",
			inMsg: &cpb.RemoveResponse{
				Code: cpb.RemoveResponse_RUNNING,
			},
			inImage: "some-image",
			inTag:   "some-tag",
			wantMsg: &cpb.RemoveRequest{
				Name: "some-image",
				Tag:  "some-tag",
			},
			wantErr: ErrRunning,
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeRemovingContainerzServer{
				sendMsg: tc.inMsg,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			err = cli.Remove(ctx, tc.inImage, tc.inTag)
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("Remove(%q, %q) returned an unexpected error: %v", tc.inImage, tc.inTag, err)
				}
			}

			if diff := cmp.Diff(tc.wantMsg, fcm.receivedMsg, protocmp.Transform()); diff != "" {
				t.Errorf("Remove(%q, %q) returned diff (-got, +want):\n%s", tc.inImage, tc.inTag, diff)
			}

			if diff := cmp.Diff(err, tc.wantErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Remove(%q, %q) returned diff (-got, +want):\n%s", tc.inImage, tc.inTag, diff)
			}
		})
	}
}

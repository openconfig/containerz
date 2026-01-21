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

type fakeImageRemovingContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.RemoveImageRequest
	sendMsg     *cpb.RemoveImageResponse
	sendErr     error
}

func (f *fakeImageRemovingContainerzServer) RemoveImage(_ context.Context, req *cpb.RemoveImageRequest) (*cpb.RemoveImageResponse, error) {
	f.receivedMsg = req
	return f.sendMsg, f.sendErr
}

func TestImageRemove(t *testing.T) {
	tests := []struct {
		name    string
		inImage string
		inTag   string
		inForce bool
		inMsg   *cpb.RemoveImageResponse

		wantMsg *cpb.RemoveImageRequest
		wantErr error
	}{
		{
			name:    "success",
			inMsg:   &cpb.RemoveImageResponse{},
			inImage: "some-image",
			inTag:   "some-tag",
			wantMsg: &cpb.RemoveImageRequest{
				Name:  "some-image",
				Tag:   "some-tag",
				Force: false,
			},
		},
		{
			name:    "success-with-force",
			inMsg:   &cpb.RemoveImageResponse{},
			inImage: "some-image",
			inTag:   "some-tag",
			inForce: true,
			wantMsg: &cpb.RemoveImageRequest{
				Name:  "some-image",
				Tag:   "some-tag",
				Force: true,
			},
		},
		{
			name:    "not-found",
			inImage: "some-image",
			inTag:   "some-tag",
			wantMsg: &cpb.RemoveImageRequest{
				Name:  "some-image",
				Tag:   "some-tag",
				Force: false,
			},
			wantErr: ErrNotFound,
		},
		{
			name:    "running",
			inImage: "some-image",
			inTag:   "some-tag",
			wantMsg: &cpb.RemoveImageRequest{
				Name:  "some-image",
				Tag:   "some-tag",
				Force: false,
			},
			wantErr: ErrRunning,
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeImageRemovingContainerzServer{
				sendMsg: tc.inMsg,
				// passthrough error from test.
				sendErr: tc.wantErr,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			err = cli.RemoveImage(ctx, tc.inImage, tc.inTag, tc.inForce)
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("RemoveImage(%q, %q, %t) returned an unexpected error: %v", tc.inImage, tc.inTag, tc.inForce, err)
				}
			}

			if diff := cmp.Diff(tc.wantMsg, fcm.receivedMsg, protocmp.Transform()); diff != "" {
				t.Errorf("RemoveImage(%q, %q, %t) returned diff (-got, +want):\n%s", tc.inImage, tc.inTag, tc.inForce, diff)
			}

			if diff := cmp.Diff(err, tc.wantErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("RemoveImage(%q, %q, %t) returned diff (-got, +want):\n%s", tc.inImage, tc.inTag, tc.inForce, diff)
			}
		})
	}
}

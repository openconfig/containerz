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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeImageRemoveContainerContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.RemoveContainerRequest
	sendErr     error
}

func (f *fakeImageRemoveContainerContainerzServer) RemoveContainer(_ context.Context, req *cpb.RemoveContainerRequest) (*cpb.RemoveContainerResponse, error) {
	f.receivedMsg = req
	return nil, f.sendErr
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name    string
		inImage string
		inErr   error
		inForce bool
		wantMsg *cpb.RemoveContainerRequest
		wantErr error
	}{
		{
			name:    "success",
			inImage: "some-image",
			wantMsg: &cpb.RemoveContainerRequest{
				Name:  "some-image",
				Force: false,
			},
		},
		{
			name:    "success-with-force",
			inImage: "some-image",
			inForce: true,
			wantMsg: &cpb.RemoveContainerRequest{
				Name:  "some-image",
				Force: true,
			},
		},
		{
			name:    "not-found",
			inErr:   status.Errorf(codes.NotFound, "resource was not found"),
			inImage: "some-image",
			wantMsg: &cpb.RemoveContainerRequest{
				Name:  "some-image",
				Force: false,
			},
			wantErr: ErrNotFound,
		},
		{
			name:    "running",
			inErr:   status.Errorf(codes.FailedPrecondition, "resource is running"),
			inImage: "some-image",
			wantMsg: &cpb.RemoveContainerRequest{
				Name:  "some-image",
				Force: false,
			},
			wantErr: ErrRunning,
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeImageRemoveContainerContainerzServer{
				sendErr: tc.inErr,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			err = cli.RemoveContainer(ctx, tc.inImage, tc.inForce)
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("Remove(%q, %t) returned an unexpected error: %v", tc.inImage, tc.inForce, err)
				}
			}

			if diff := cmp.Diff(tc.wantMsg, fcm.receivedMsg, protocmp.Transform()); diff != "" {
				t.Errorf("Remove(%q, %t) returned diff (-got, +want):\n%s", tc.inImage, tc.inForce, diff)
			}

			if diff := cmp.Diff(err, tc.wantErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Remove(%q, %t) returned diff (-got, +want):\n%s", tc.inImage, tc.inForce, diff)
			}
		})
	}
}

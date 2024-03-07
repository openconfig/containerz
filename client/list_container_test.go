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

type fakeListingContainerzServer struct {
	fakeContainerzServer

	sendMsgs         []*cpb.ListContainerResponse
	receivedMessages []*cpb.ListContainerRequest
}

func (f *fakeListingContainerzServer) ListContainer(req *cpb.ListContainerRequest, srv cpb.Containerz_ListContainerServer) error {
	f.receivedMessages = append(f.receivedMessages, req)

	for _, resp := range f.sendMsgs {
		if err := srv.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func TestList(t *testing.T) {
	tests := []struct {
		name     string
		inAll    bool
		inLimit  int32
		inFilter map[string][]string
		inMsgs   []*cpb.ListContainerResponse

		wantInfo []*ContainerInfo
		wantMsgs []*cpb.ListContainerRequest
	}{
		{
			name:  "all-no-limit",
			inAll: true,
			inMsgs: []*cpb.ListContainerResponse{
				{
					Id:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
					Status:    cpb.ListContainerResponse_RUNNING,
				},
			},
			wantInfo: []*ContainerInfo{
				&ContainerInfo{
					ID:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
					State:     "RUNNING",
				},
			},
			wantMsgs: []*cpb.ListContainerRequest{
				&cpb.ListContainerRequest{
					All:   true,
					Limit: 0,
				},
			},
		},
		{
			name:    "all-with-limit",
			inAll:   true,
			inLimit: 10,
			inMsgs: []*cpb.ListContainerResponse{
				{
					Id:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
					Status:    cpb.ListContainerResponse_RUNNING,
				},
			},
			wantInfo: []*ContainerInfo{
				&ContainerInfo{
					ID:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
					State:     "RUNNING",
				},
			},
			wantMsgs: []*cpb.ListContainerRequest{
				&cpb.ListContainerRequest{
					All:   true,
					Limit: 10,
				},
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeListingContainerzServer{sendMsgs: tc.inMsgs}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*ContainerInfo{}

			ch, err := cli.ListContainer(ctx, tc.inAll, tc.inLimit, tc.inFilter)
			if err != nil {
				t.Fatalf("List(%t, %d, %v) returned an unexpected error: %v", tc.inAll, tc.inLimit, tc.inFilter, err)
			}

			go func() {
				for info := range ch {
					got = append(got, info)
				}
				close(doneCh)
			}()
			<-doneCh

			if diff := cmp.Diff(tc.wantInfo, got); diff != "" {
				t.Errorf("List(%t, %d, %v) returned an unexpected diff (-want +got):\n%s", tc.inAll, tc.inLimit, tc.inFilter, diff)
			}

			if diff := cmp.Diff(tc.wantMsgs, fcm.receivedMessages, protocmp.Transform()); diff != "" {
				t.Errorf("List(%t, %d, %v) returned an unexpected diff (-want +got):\n%s", tc.inAll, tc.inLimit, tc.inFilter, diff)
			}
		})
	}
}

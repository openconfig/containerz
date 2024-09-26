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

type fakeImageListingContainerzServer struct {
	fakeContainerzServer

	sendMsgs         []*cpb.ListImageResponse
	receivedMessages []*cpb.ListImageRequest
}

func (f *fakeImageListingContainerzServer) ListImage(req *cpb.ListImageRequest, srv cpb.Containerz_ListImageServer) error {
	f.receivedMessages = append(f.receivedMessages, req)

	for _, resp := range f.sendMsgs {
		if err := srv.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func TestImageList(t *testing.T) {
	tests := []struct {
		name     string
		inLimit  int32
		inFilter map[string][]string
		inMsgs   []*cpb.ListImageResponse

		wantInfo []*ImageInfo
		wantMsgs []*cpb.ListImageRequest
	}{
		{
			name: "all-no-limit",
			inMsgs: []*cpb.ListImageResponse{
				{
					Id:        "some-id",
					ImageName: "some-name",
					Tag:       "some-tag",
				},
			},
			wantInfo: []*ImageInfo{
				&ImageInfo{
					ID:        "some-id",
					ImageName: "some-name",
					ImageTag:  "some-tag",
				},
			},
			wantMsgs: []*cpb.ListImageRequest{
				&cpb.ListImageRequest{
					Limit: 0,
				},
			},
		},
		{
			name:    "all-limit-10",
			inLimit: 10,
			inMsgs: []*cpb.ListImageResponse{
				{
					Id:        "some-id",
					ImageName: "some-name",
					Tag:       "some-tag",
				},
			},
			wantInfo: []*ImageInfo{
				&ImageInfo{
					ID:        "some-id",
					ImageName: "some-name",
					ImageTag:  "some-tag",
				},
			},
			wantMsgs: []*cpb.ListImageRequest{
				&cpb.ListImageRequest{
					Limit: 10,
				},
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeImageListingContainerzServer{sendMsgs: tc.inMsgs}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*ImageInfo{}

			ch, err := cli.ListImage(ctx, tc.inLimit, tc.inFilter)
			if err != nil {
				t.Fatalf("ListImage(%d, %v) returned an unexpected error: %v", tc.inLimit, tc.inFilter, err)
			}

			go func() {
				for info := range ch {
					got = append(got, info)
				}
				close(doneCh)
			}()
			<-doneCh

			if diff := cmp.Diff(tc.wantInfo, got); diff != "" {
				t.Errorf("ListImage(%d, %v) returned an unexpected diff (-want +got):\n%s", tc.inLimit, tc.inFilter, diff)
			}

			if diff := cmp.Diff(tc.wantMsgs, fcm.receivedMessages, protocmp.Transform()); diff != "" {
				t.Errorf("ListImage(%d, %v) returned an unexpected diff (-want +got):\n%s", tc.inLimit, tc.inFilter, diff)
			}
		})
	}
}

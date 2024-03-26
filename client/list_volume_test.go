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
	"time"

	tpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeListingVolumeServer struct {
	fakeContainerzServer

	sendMsgs         []*cpb.ListVolumeResponse
	receivedMessages []*cpb.ListVolumeRequest
}

func (f *fakeListingVolumeServer) ListVolume(req *cpb.ListVolumeRequest, srv cpb.Containerz_ListVolumeServer) error {
	f.receivedMessages = append(f.receivedMessages, req)

	for _, resp := range f.sendMsgs {
		if err := srv.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func TestListVolume(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name string

		inFilter map[string][]string
		inMsgs   []*cpb.ListVolumeResponse

		wantInfo []*VolumeInfo
		wantMsgs []*cpb.ListVolumeRequest
	}{
		{
			name: "all-no-limit",
			inMsgs: []*cpb.ListVolumeResponse{
				{
					Name:    "some-name",
					Driver:  "some-driver",
					Created: tpb.New(testTime),
				},
			},
			wantInfo: []*VolumeInfo{
				&VolumeInfo{
					Name:         "some-name",
					Driver:       "some-driver",
					CreationTime: testTime,
				},
			},
			wantMsgs: []*cpb.ListVolumeRequest{
				&cpb.ListVolumeRequest{},
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeListingVolumeServer{sendMsgs: tc.inMsgs}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*VolumeInfo{}

			ch, err := cli.ListVolume(ctx, tc.inFilter)
			if err != nil {
				t.Fatalf("List(%v) returned an unexpected error: %v", tc.inFilter, err)
			}

			go func() {
				for info := range ch {
					got = append(got, info)
				}
				close(doneCh)
			}()
			<-doneCh

			if diff := cmp.Diff(tc.wantInfo, got); diff != "" {
				t.Errorf("ListVolume(%v) returned an unexpected diff (-want +got):\n%s", tc.inFilter, diff)
			}

			if diff := cmp.Diff(tc.wantMsgs, fcm.receivedMessages, protocmp.Transform()); diff != "" {
				t.Errorf("ListVolume(%v) returned an unexpected diff (-want +got):\n%s", tc.inFilter, diff)
			}
		})
	}
}

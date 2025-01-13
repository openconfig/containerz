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
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakePushingContainerzServer struct {
	fakeContainerzServer

	sendMsgs         []*cpb.DeployResponse
	receivedMessages []*cpb.DeployRequest
}

func (f *fakePushingContainerzServer) Deploy(srv cpb.Containerz_DeployServer) error {
	for {
		msg, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		f.receivedMessages = append(f.receivedMessages, msg)

		if len(f.sendMsgs) == 0 {
			return nil
		}

		m := f.sendMsgs[0]
		f.sendMsgs = f.sendMsgs[1:]

		if err := srv.Send(m); err != nil {
			return err
		}
	}
}

func TestPushImage(t *testing.T) {
	tests := []struct {
		name     string
		inImage  string
		inTag    string
		inFile   string
		inPlugin bool
		inMsgs   []*cpb.DeployResponse

		wantProgress []*Progress
		wantMsgs     []*cpb.DeployRequest
	}{
		{
			name:    "invalid-msg",
			inImage: "some-image",
			inTag:   "some-tag",
			inFile:  "testdata/reader-data.txt",
			inMsgs: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{},
					},
				},
			},
			wantMsgs: []*cpb.DeployRequest{
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransfer{
						ImageTransfer: &cpb.ImageTransfer{
							Name:      "some-image",
							Tag:       "some-tag",
							ImageSize: 26,
						},
					},
				},
			},
			wantProgress: []*Progress{
				&Progress{
					Error: status.Errorf(codes.InvalidArgument, "received unexpected message type: *containerz.DeployResponse_ImageTransferProgress"),
				},
			},
		},
		{
			name:    "valid-transfer",
			inImage: "some-image",
			inTag:   "some-tag",
			inFile:  "testdata/reader-data.txt",
			inMsgs: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferReady{
						ImageTransferReady: &cpb.ImageTransferReady{
							ChunkSize: 28,
						},
					},
				},
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 26,
						},
					},
				},
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferSuccess{
						ImageTransferSuccess: &cpb.ImageTransferSuccess{
							Name: "some-image",
							Tag:  "some-tag",
						},
					},
				},
			},

			wantMsgs: []*cpb.DeployRequest{
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransfer{
						ImageTransfer: &cpb.ImageTransfer{
							Name:      "some-image",
							Tag:       "some-tag",
							ImageSize: 26,
						},
					},
				},
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_Content{
						Content: []byte("some really important data"),
					},
				},
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransferEnd{
						ImageTransferEnd: &cpb.ImageTransferEnd{},
					},
				},
			},
			wantProgress: []*Progress{
				&Progress{
					BytesReceived: 26,
				},
				&Progress{
					Finished: true,
					Image:    "some-image",
					Tag:      "some-tag",
				},
			},
		},
		{
			name:     "valid-plugin-transfer",
			inImage:  "some-image",
			inTag:    "some-tag",
			inFile:   "testdata/reader-data.txt",
			inPlugin: true,
			inMsgs: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferReady{
						ImageTransferReady: &cpb.ImageTransferReady{
							ChunkSize: 28,
						},
					},
				},
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 26,
						},
					},
				},
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferSuccess{
						ImageTransferSuccess: &cpb.ImageTransferSuccess{
							Name: "some-image",
							Tag:  "some-tag",
						},
					},
				},
			},

			wantMsgs: []*cpb.DeployRequest{
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransfer{
						ImageTransfer: &cpb.ImageTransfer{
							Name:      "some-image",
							Tag:       "some-tag",
							ImageSize: 26,
							IsPlugin:  true,
						},
					},
				},
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_Content{
						Content: []byte("some really important data"),
					},
				},
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransferEnd{
						ImageTransferEnd: &cpb.ImageTransferEnd{},
					},
				},
			},
			wantProgress: []*Progress{
				&Progress{
					BytesReceived: 26,
				},
				&Progress{
					Finished: true,
					Image:    "some-image",
					Tag:      "some-tag",
				},
			},
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakePushingContainerzServer{
				sendMsgs: tc.inMsgs,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*Progress{}

			ch, err := cli.PushImage(ctx, tc.inImage, tc.inTag, tc.inFile, tc.inPlugin)
			if err != nil {
				t.Fatalf("PushImage(%q, %q, %q, %t) returned an unexpected error: %v", tc.inImage, tc.inTag, tc.inFile, tc.inPlugin, err)
			}

			go func() {
				for prog := range ch {
					got = append(got, prog)
				}
				close(doneCh)
			}()
			<-doneCh

			if diff := cmp.Diff(tc.wantProgress, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("PushImage(%q, %q, %q) returned an unexpected diff (-want +got): %v", tc.inImage, tc.inTag, tc.inFile, diff)
			}

			if diff := cmp.Diff(tc.wantMsgs, fcm.receivedMessages, protocmp.Transform()); diff != "" {
				t.Errorf("PushImage(%q, %q, %q) returned an unexpected diff (-want +got): %v", tc.inImage, tc.inTag, tc.inFile, diff)
			}
		})
	}
}

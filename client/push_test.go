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
		name    string
		inImage string
		inTag   string
		inFile  string
		inMsgs  []*cpb.DeployResponse

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
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakePushingContainerzServer{
				sendMsgs: tc.inMsgs,
			}
			addr := newServer(t, fcm)
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*Progress{}

			ch, err := cli.PushImage(ctx, tc.inImage, tc.inTag, tc.inFile)
			if err != nil {
				t.Fatalf("PushImage(%q, %q, %q) returned an unexpected error: %v", tc.inImage, tc.inTag, tc.inFile, err)
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

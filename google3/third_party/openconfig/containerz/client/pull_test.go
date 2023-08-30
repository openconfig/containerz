package client

import (
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	commonpb "google3/third_party/openconfig/gnoi/common/common_go_proto"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakePullingContainerzServer struct {
	fakeContainerzServer

	sendMsgs         []*cpb.DeployResponse
	receivedMessages []*cpb.DeployRequest
}

func (f *fakePullingContainerzServer) Deploy(srv cpb.Containerz_DeployServer) error {
	msg, err := srv.Recv()
	if err == io.EOF {
		return nil
	}

	f.receivedMessages = append(f.receivedMessages, msg)

	for _, resp := range f.sendMsgs {
		if err := srv.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func TestPullImage(t *testing.T) {
	tests := []struct {
		name       string
		inImage    string
		inTag      string
		inProgress []*cpb.DeployResponse

		wantProgress []*Progress
		wantMsgs     []*cpb.DeployRequest
	}{
		{
			name:    "tiny-image",
			inImage: "tiny-image",
			inTag:   "tiny-tag",
			inProgress: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 1,
						},
					},
				},
			},
			wantMsgs: []*cpb.DeployRequest{
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransfer{
						ImageTransfer: &cpb.ImageTransfer{
							Name:           "tiny-image",
							Tag:            "tiny-tag",
							RemoteDownload: &commonpb.RemoteDownload{},
						},
					},
				},
			},
			wantProgress: []*Progress{
				&Progress{
					BytesReceived: 1,
				},
			},
		},
		{
			name:    "large-image",
			inImage: "large-image",
			inTag:   "large-tag",
			inProgress: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 10,
						},
					},
				},
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 10,
						},
					},
				},
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 1,
						},
					},
				},
			},
			wantMsgs: []*cpb.DeployRequest{
				&cpb.DeployRequest{
					Request: &cpb.DeployRequest_ImageTransfer{
						ImageTransfer: &cpb.ImageTransfer{
							Name:           "large-image",
							Tag:            "large-tag",
							RemoteDownload: &commonpb.RemoteDownload{},
						},
					},
				},
			},
			wantProgress: []*Progress{
				&Progress{
					BytesReceived: 10,
				},
				&Progress{
					BytesReceived: 10,
				},
				&Progress{
					BytesReceived: 1,
				},
			},
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakePullingContainerzServer{
				sendMsgs: tc.inProgress,
			}
			addr := newServer(t, fcm)
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*Progress{}

			ch, err := cli.PullImage(ctx, tc.inImage, tc.inTag, nil)
			if err != nil {
				t.Fatalf("PullImage(%q, %q) returned an unexpected error: %v", tc.inImage, tc.inTag, err)
			}

			go func() {
				for prog := range ch {
					got = append(got, prog)
				}
				close(doneCh)
			}()
			<-doneCh

			if diff := cmp.Diff(tc.wantProgress, got); diff != "" {
				t.Errorf("PullImage(%q, %q) returned an unexpected diff (-want +got): %v", tc.inImage, tc.inTag, diff)
			}

			if diff := cmp.Diff(tc.wantMsgs, fcm.receivedMessages, protocmp.Transform()); diff != "" {
				t.Errorf("PullImage(%q, %q) returned an unexpected diff (-want +got): %v", tc.inImage, tc.inTag, diff)
			}
		})
	}
}

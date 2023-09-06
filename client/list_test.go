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

	sendMsgs         []*cpb.ListResponse
	receivedMessages []*cpb.ListRequest
}

func (f *fakeListingContainerzServer) List(req *cpb.ListRequest, srv cpb.Containerz_ListServer) error {
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
		inMsgs   []*cpb.ListResponse

		wantInfo []*ContainerInfo
		wantMsgs []*cpb.ListRequest
	}{
		{
			name:  "all-no-limit",
			inAll: true,
			inMsgs: []*cpb.ListResponse{
				{
					Id:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
					Status:    cpb.ListResponse_RUNNING,
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
			wantMsgs: []*cpb.ListRequest{
				&cpb.ListRequest{
					All:   true,
					Limit: 0,
				},
			},
		},
		{
			name:    "all-with-limit",
			inAll:   true,
			inLimit: 10,
			inMsgs: []*cpb.ListResponse{
				{
					Id:        "some-id",
					Name:      "some-name",
					ImageName: "some-image",
					Status:    cpb.ListResponse_RUNNING,
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
			wantMsgs: []*cpb.ListRequest{
				&cpb.ListRequest{
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
			addr := newServer(t, fcm)
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*ContainerInfo{}

			ch, err := cli.List(ctx, tc.inAll, tc.inLimit, tc.inFilter)
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

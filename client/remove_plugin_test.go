package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeRemovePluginServer struct {
	fakeContainerzServer

	recvMsg *cpb.RemovePluginRequest
}

func (f *fakeRemovePluginServer) RemovePlugin(ctx context.Context, req *cpb.RemovePluginRequest) (*cpb.RemovePluginResponse, error) {
	f.recvMsg = req
	return nil, nil
}

func TestRemovePlugin(t *testing.T) {
	tests := []struct {
		name string

		inInstance string

		wantReq *cpb.RemovePluginRequest
	}{
		{
			name:       "remove-plugin",
			inInstance: "test",
			wantReq: &cpb.RemovePluginRequest{
				InstanceName: "test",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fcm := &fakeRemovePluginServer{}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			if err := cli.RemovePlugin(ctx, tc.inInstance); err != nil {
				t.Fatalf("RemovePlugin(%v) returned an unexpected error: %v", tc.inInstance, err)
			}

			if diff := cmp.Diff(tc.wantReq, fcm.recvMsg, protocmp.Transform()); diff != "" {
				t.Errorf("RemovePlugin(%v) returned an unexpected diff (-want +got):\n%s", tc.inInstance, diff)
			}

		})
	}
}

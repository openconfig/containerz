package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeStopPluginServer struct {
	fakeContainerzServer

	recvMsg *cpb.StopPluginRequest
}

func (f *fakeStopPluginServer) StopPlugin(ctx context.Context, req *cpb.StopPluginRequest) (*cpb.StopPluginResponse, error) {
	f.recvMsg = req
	return nil, nil
}

func TestStopPlugin(t *testing.T) {
	tests := []struct {
		name string

		inInstance string

		wantReq *cpb.StopPluginRequest
	}{
		{
			name:       "stop-plugin",
			inInstance: "test",
			wantReq: &cpb.StopPluginRequest{
				InstanceName: "test",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fcm := &fakeStopPluginServer{}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			if err := cli.StopPlugin(ctx, tc.inInstance); err != nil {
				t.Fatalf("StopPlugin(%v) returned an unexpected error: %v", tc.inInstance, err)
			}

			if diff := cmp.Diff(tc.wantReq, fcm.recvMsg, protocmp.Transform()); diff != "" {
				t.Errorf("StopPlugin(%v) returned an unexpected diff (-want +got):\n%s", tc.inInstance, diff)
			}

		})
	}
}

package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeListPluginServer struct {
	fakeContainerzServer
	sendMsgs *cpb.ListPluginsResponse
	recvMsg  *cpb.ListPluginsRequest
}

func (f *fakeListPluginServer) ListPlugins(ctx context.Context, req *cpb.ListPluginsRequest) (*cpb.ListPluginsResponse, error) {
	f.recvMsg = req

	return f.sendMsgs, nil
}

func TestListPlugin(t *testing.T) {
	tests := []struct {
		name string

		inInstance string
		inMsgs     *cpb.ListPluginsResponse

		wantReq  *cpb.ListPluginsRequest
		wantMsgs []*cpb.Plugin
	}{
		{
			name:       "with-instance",
			inInstance: "test",
			inMsgs: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "test",
					},
				},
			},
			wantReq: &cpb.ListPluginsRequest{
				InstanceName: "test",
			},
			wantMsgs: []*cpb.Plugin{
				{
					InstanceName: "test",
				},
			},
		},
		{
			name: "without-instance",
			inMsgs: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "test",
					},
				},
			},
			wantReq: &cpb.ListPluginsRequest{},
			wantMsgs: []*cpb.Plugin{
				&cpb.Plugin{
					InstanceName: "test",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fcm := &fakeListPluginServer{sendMsgs: tc.inMsgs}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			got, err := cli.ListPlugin(ctx, tc.inInstance)
			if err != nil {
				t.Fatalf("ListPlugin(%v) returned an unexpected error: %v", tc.inInstance, err)
			}

			if diff := cmp.Diff(tc.wantReq, fcm.recvMsg, protocmp.Transform()); diff != "" {
				t.Errorf("ListPlugin(%v) returned an unexpected diff (-want +got):\n%s", tc.inInstance, diff)
			}

			if diff := cmp.Diff(tc.wantMsgs, got, protocmp.Transform()); diff != "" {
				t.Errorf("ListPlugin(%v) returned an unexpected diff (-want +got):\n%s", tc.inInstance, diff)
			}
		})
	}
}

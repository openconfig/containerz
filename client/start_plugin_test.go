package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeStartPluginServer struct {
	fakeContainerzServer

	recvMsg *cpb.StartPluginRequest
}

func (f *fakeStartPluginServer) StartPlugin(ctx context.Context, req *cpb.StartPluginRequest) (*cpb.StartPluginResponse, error) {
	f.recvMsg = req
	return &cpb.StartPluginResponse{
		InstanceName: req.GetInstanceName(),
	}, nil
}

func TestStartPlugin(t *testing.T) {
	tests := []struct {
		name string

		inInstance string
		inName     string
		inConfig   string

		wantReq *cpb.StartPluginRequest
		wantErr bool
	}{
		{
			name:       "no-such-file",
			inInstance: "test",
			inName:     "test",
			inConfig:   "test",
			wantErr:    true,
		},
		{
			name:       "bad-json",
			inInstance: "test",
			inName:     "test",
			inConfig:   "testdata/bad.json",
			wantErr:    true,
		},
		{
			name:       "good-json",
			inInstance: "test",
			inName:     "test",
			inConfig:   "testdata/good.json",
			wantReq: &cpb.StartPluginRequest{
				InstanceName: "test",
				Name:         "test",
				Config:       "{\n  \"i-am\": \"good-json\"\n}",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fcm := &fakeStartPluginServer{}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			if err := cli.StartPlugin(ctx, tc.inName, tc.inInstance, tc.inConfig); err != nil {
				if tc.wantErr {
					return
				}
				t.Fatalf("StartPlugin(%q, %q, %q) returned an unexpected error: %v", tc.inName, tc.inInstance, tc.inConfig, err)
			}

			if diff := cmp.Diff(tc.wantReq, fcm.recvMsg, protocmp.Transform()); diff != "" {
				t.Errorf("StopPlugin(%v) returned an unexpected diff (-want +got):\n%s", tc.inInstance, diff)
			}

		})
	}
}

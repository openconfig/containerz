package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeStoppingContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.StopRequest
}

func (f *fakeStoppingContainerzServer) Stop(ctx context.Context, req *cpb.StopRequest) (*cpb.StopResponse, error) {
	f.receivedMsg = req
	return &cpb.StopResponse{}, nil
}

func TestStop(t *testing.T) {
	tests := []struct {
		name       string
		inInstance string
		inForce    bool

		wantMsg *cpb.StopRequest
	}{
		{
			name:       "simple",
			inInstance: "some-instance",
			inForce:    true,
			wantMsg: &cpb.StopRequest{
				InstanceName: "some-instance",
				Force:        true,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeStoppingContainerzServer{}
			addr := newServer(t, fcm)
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			if err := cli.Stop(ctx, tc.inInstance, tc.inForce); err != nil {
				t.Fatalf("Stop(%q, %t) returned an unexpected error: %v", tc.inInstance, tc.inForce, err)
			}

			if diff := cmp.Diff(fcm.receivedMsg, tc.wantMsg, protocmp.Transform()); diff != "" {
				t.Errorf("Stop(%q, %t) returned diff(-want, +got):\n%s", tc.inInstance, tc.inForce, diff)
			}
		})
	}
}

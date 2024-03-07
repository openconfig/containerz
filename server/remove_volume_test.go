package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestRemoveVolume(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		inReq    *cpb.RemoveVolumeRequest
		inOpts   []Option
		wantResp *cpb.RemoveVolumeResponse
	}{
		{
			name: "empty request",
			inReq: &cpb.RemoveVolumeRequest{
				Name: "test-volume",
			},
			inOpts:   []Option{},
			wantResp: &cpb.RemoveVolumeResponse{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeContainerManager{}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.RemoveVolume(ctx, tc.inReq)
			if err != nil {
				t.Errorf("RemoveVolume(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestStop(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.StopRequest
		inOpts    []Option
		wantResp  *cpb.StopResponse
		wantState *fakeContainerManager
	}{
		{
			name: "no-force",
			inReq: &cpb.StopRequest{
				InstanceName: "some-name",
			},
			wantResp: &cpb.StopResponse{},
			wantState: &fakeContainerManager{
				Instance: "some-name",
				Force:    false,
			},
		},
		{
			name: "nforce",
			inReq: &cpb.StopRequest{
				InstanceName: "some-name",
				Force:        true,
			},
			wantResp: &cpb.StopResponse{},
			wantState: &fakeContainerManager{
				Instance: "some-name",
				Force:    true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{}
			tc.inOpts = append(tc.inOpts, WithAddr("localhost:0"))
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.Stop(ctx, tc.inReq)
			if err != nil {
				t.Errorf("Stop(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Stop(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}

			if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Stop(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

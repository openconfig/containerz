package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestStartPlugin(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.StartPluginRequest
		inOpts    []Option
		wantResp  *cpb.StartPluginResponse
		wantState *fakeContainerManager
		wantErr   bool
	}{
		{
			name: "success",
			inReq: &cpb.StartPluginRequest{
				Name:         "test",
				InstanceName: "test",
				Config:       "test",
			},
			wantResp: &cpb.StartPluginResponse{
				InstanceName: "test",
			},
			wantState: &fakeContainerManager{
				Name:     "test",
				Instance: "test",
				Config:   "test",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.StartPlugin(ctx, tc.inReq)
			if err != nil {
				t.Errorf("Start(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}

			if diff := cmp.Diff(tc.wantState, fake, protocmp.Transform(), cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestListPlugins(t *testing.T) {
	tests := []struct {
		name      string
		inOpts    []Option
		inPlugins *cpb.ListPluginsResponse
		inReq     *cpb.ListPluginsRequest
		wantResp  *cpb.ListPluginsResponse
		wantState *fakeContainerManager
	}{
		{
			name:   "success",
			inOpts: []Option{},
			inPlugins: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "some-plugin",
					},
				},
			},
			inReq: &cpb.ListPluginsRequest{},
			wantResp: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "some-plugin",
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "",
			},
		},
		{
			name:   "specific-plugin",
			inOpts: []Option{},
			inPlugins: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "some-plugin",
					},
				},
			},
			inReq: &cpb.ListPluginsRequest{
				InstanceName: "some-plugin",
			},
			wantResp: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "some-plugin",
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-plugin",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{
				listPluginMsgs: tc.inPlugins,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			gotResp, err := cli.ListPlugins(ctx, tc.inReq)
			if err != nil {
				t.Errorf("ListPlugins(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, gotResp, protocmp.Transform()); diff != "" {
				t.Errorf("ListPlugins(%+v) returned unexpected diff (-want +got):\n%s", tc.inReq, diff)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("ListPlugins(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}

package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestStopPlugin(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.StopPluginRequest
		inPlugins *cpb.ListPluginsResponse
		inOpts    []Option
		wantState *fakeContainerManager
		wantErr   bool
	}{
		{
			name: "success",
			inPlugins: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "test",
					},
				},
			},
			inReq: &cpb.StopPluginRequest{
				InstanceName: "test",
			},
			wantState: &fakeContainerManager{
				Instance: "test",
			},
		},
		{
			name: "error",
			inPlugins: &cpb.ListPluginsResponse{
				Plugins: []*cpb.Plugin{
					&cpb.Plugin{
						InstanceName: "nope",
					},
				},
			},
			inReq: &cpb.StopPluginRequest{
				InstanceName: "test",
			},
			wantState: &fakeContainerManager{
				Instance: "test",
			},
			wantErr: true,
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

			if _, err := cli.StopPlugin(ctx, tc.inReq); err != nil && !tc.wantErr {
				t.Errorf("RemovePlugin(%+v) returned error: %v", tc.inReq, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("List(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}

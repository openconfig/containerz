package server

import (
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

func TestLogs(t *testing.T) {
	tests := []struct {
		name      string
		inErr     error
		inLogs    []string
		inReq     *cpb.LogRequest
		inOpts    []Option
		wantResp  []*cpb.LogResponse
		wantState *fakeContainerManager
	}{
		{
			name: "no-logs",
			inReq: &cpb.LogRequest{
				InstanceName: "some-instance",
			},
			wantResp: []*cpb.LogResponse{},
		},
		{
			name: "simple-logs",
			inReq: &cpb.LogRequest{
				InstanceName: "some-instance",
			},
			inLogs: []string{
				"we",
				"have",
				"the",
				"logs",
			},
			wantResp: []*cpb.LogResponse{
				&cpb.LogResponse{
					Msg: "we",
				},
				&cpb.LogResponse{
					Msg: "have",
				},
				&cpb.LogResponse{
					Msg: "the",
				},
				&cpb.LogResponse{
					Msg: "logs",
				},
			},
		},
		{
			name: "simple-logs-with-follow",
			inReq: &cpb.LogRequest{
				InstanceName: "some-instance",
				Follow:       true,
			},
			inLogs: []string{
				"we",
				"have",
				"the",
				"logs",
			},
			wantResp: []*cpb.LogResponse{
				&cpb.LogResponse{
					Msg: "we",
				},
				&cpb.LogResponse{
					Msg: "have",
				},
				&cpb.LogResponse{
					Msg: "the",
				},
				&cpb.LogResponse{
					Msg: "logs",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{
				msgs: tc.inLogs,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			lCli, err := cli.Log(ctx, tc.inReq)
			if err != nil {
				t.Errorf("Log(%+v) returned error: %v", tc.inReq, err)
			}

			gotResp := []*cpb.LogResponse{}
			for {
				msg, err := lCli.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Errorf("Recv() returned error: %v", err)
				}
				gotResp = append(gotResp, msg)
			}

			if diff := cmp.Diff(tc.wantResp, gotResp, protocmp.Transform()); diff != "" {
				t.Errorf("Log(%+v) returned unexpected diff (-want +got):\n%s", tc.inReq, diff)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("Log(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}

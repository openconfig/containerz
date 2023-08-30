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

func TestList(t *testing.T) {
	tests := []struct {
		name      string
		inErr     error
		inCnts    []*cpb.ListResponse
		inReq     *cpb.ListRequest
		inOpts    []Option
		wantResp  []*cpb.ListResponse
		wantState *fakeContainerManager
	}{
		{
			name: "no-containers",
			inReq: &cpb.ListRequest{
				All:   true,
				Limit: 10,
			},
			wantState: &fakeContainerManager{
				All:   true,
				Limit: 10,
			},
			wantResp: []*cpb.ListResponse{},
		},
		{
			name: "containers",
			inReq: &cpb.ListRequest{
				All:   true,
				Limit: 10,
			},
			inCnts: []*cpb.ListResponse{
				&cpb.ListResponse{
					Id: "some-id",
				},
				&cpb.ListResponse{
					Id: "other-id",
				},
			},
			wantState: &fakeContainerManager{
				All:   true,
				Limit: 10,
			},
			wantResp: []*cpb.ListResponse{
				&cpb.ListResponse{
					Id: "some-id",
				},
				&cpb.ListResponse{
					Id: "other-id",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{
				listMsgs: tc.inCnts,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			lCli, err := cli.List(ctx, tc.inReq)
			if err != nil {
				t.Errorf("List(%+v) returned error: %v", tc.inReq, err)
			}

			gotResp := []*cpb.ListResponse{}
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
				t.Errorf("List(%+v) returned unexpected diff (-want +got):\n%s", tc.inReq, diff)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("List(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}
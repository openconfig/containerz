package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeStartingContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.StartRequest
	sendMsg     *cpb.StartResponse
}

func (f *fakeStartingContainerzServer) Start(ctx context.Context, req *cpb.StartRequest) (*cpb.StartResponse, error) {
	f.receivedMsg = req
	return f.sendMsg, nil
}

func TestStart(t *testing.T) {
	tests := []struct {
		name       string
		inImage    string
		inTag      string
		inCmd      string
		inInstance string
		inPorts    []string
		inEnvs     []string
		inMsg      *cpb.StartResponse

		wantMsg *cpb.StartRequest
		wantID  string
		wantErr error
	}{
		{
			name:       "simple",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inMsg: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
			},
		},
		{
			name:       "simple-with-error",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inMsg: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartError{
					StartError: &cpb.StartError{
						Details: "oh no!",
					},
				},
			},
			wantMsg: &cpb.StartRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
			},
			wantErr: status.Error(codes.Internal, "failed to start container: oh no!"),
		},
		{
			name:       "simple-with-ports",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inPorts:    []string{"1:1", "2:2"},
			inMsg: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
				Ports: []*cpb.StartRequest_Port{
					&cpb.StartRequest_Port{
						Internal: 1,
						External: 1,
					},
					&cpb.StartRequest_Port{
						Internal: 2,
						External: 2,
					},
				},
			},
		},
		{
			name:       "simple-with-envs",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inEnvs:     []string{"env1=cool", "env2=cooler"},
			inMsg: &cpb.StartResponse{
				Response: &cpb.StartResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
				Environment:  map[string]string{"env1": "cool", "env2": "cooler"},
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeStartingContainerzServer{
				sendMsg: tc.inMsg,
			}
			addr := newServer(t, fcm)
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			gotID, err := cli.Start(ctx, tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, WithPorts(tc.inPorts), WithEnv(tc.inEnvs))
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("Start(%q, %q, %q, %q, %v, %v) returned an unexpected error: %v", tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inPorts, tc.inEnvs, err)
				}
			}

			if tc.wantID != gotID {
				t.Errorf("Start(%q, %q, %q, %q, %v, %v) returned incorrect id - want %s, got %s", tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inPorts, tc.inEnvs, tc.wantID, gotID)
			}

			if diff := cmp.Diff(err, tc.wantErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Start(%q, %q, %q, %q, %v, %v) returned diff (-got, +want):\n%s", tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inPorts, tc.inEnvs, diff)
			}
		})
	}
}

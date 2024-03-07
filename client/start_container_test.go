// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	receivedMsg *cpb.StartContainerRequest
	sendMsg     *cpb.StartContainerResponse
}

func (f *fakeStartingContainerzServer) StartContainer(ctx context.Context, req *cpb.StartContainerRequest) (*cpb.StartContainerResponse, error) {
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
		inVols     []string
		inMsg      *cpb.StartContainerResponse

		wantMsg *cpb.StartContainerRequest
		wantID  string
		wantErr error
	}{
		{
			name:       "simple",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inMsg: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartContainerRequest{
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
			inMsg: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartError{
					StartError: &cpb.StartError{
						Details: "oh no!",
					},
				},
			},
			wantMsg: &cpb.StartContainerRequest{
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
			inMsg: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartContainerRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
				Ports: []*cpb.StartContainerRequest_Port{
					&cpb.StartContainerRequest_Port{
						Internal: 1,
						External: 1,
					},
					&cpb.StartContainerRequest_Port{
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
			inMsg: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartContainerRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
				Environment:  map[string]string{"env1": "cool", "env2": "cooler"},
			},
		},
		{
			name:       "simple-with-envs-and-volumes",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inEnvs:     []string{"env1=cool", "env2=cooler"},
			inVols:     []string{"vol1:/aa", "vol2:/bb:ro"},
			inMsg: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: &cpb.StartContainerRequest{
				ImageName:    "some-image",
				Tag:          "some-tag",
				Cmd:          "some-cmd",
				InstanceName: "some-instance",
				Environment:  map[string]string{"env1": "cool", "env2": "cooler"},
				Volumes: []*cpb.Volume{
					&cpb.Volume{
						Name:       "vol1",
						MountPoint: "/aa",
					},
					&cpb.Volume{
						Name:       "vol2",
						MountPoint: "/bb",
						ReadOnly:   true,
					},
				},
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeStartingContainerzServer{
				sendMsg: tc.inMsg,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			gotID, err := cli.StartContainer(ctx, tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, WithPorts(tc.inPorts), WithEnv(tc.inEnvs), WithVolumes(tc.inVols))
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

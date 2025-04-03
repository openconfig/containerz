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
	cpb "github.com/openconfig/gnoi/containerz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUpdatingContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.UpdateContainerRequest
	sendMsg     *cpb.UpdateContainerResponse
}

func (f *fakeUpdatingContainerzServer) UpdateContainer(ctx context.Context, req *cpb.UpdateContainerRequest) (*cpb.UpdateContainerResponse, error) {
	f.receivedMsg = req
	return f.sendMsg, nil
}

func wrapInUpdateRequest(req *cpb.StartContainerRequest, async bool) *cpb.UpdateContainerRequest {
	return &cpb.UpdateContainerRequest{
		InstanceName: req.GetInstanceName(),
		ImageName:    req.GetImageName(),
		ImageTag:     req.GetTag(),
		Params:       req,
		Async:        async,
	}
}

func TestUpdateContainer(t *testing.T) {
	tests := []struct {
		name       string
		inImage    string
		inTag      string
		inCmd      string
		inInstance string
		inAsync    bool
		inPorts    []string
		inEnvs     []string
		inVols     []string
		inDevices  []string
		inMsg      *cpb.UpdateContainerResponse

		wantMsg *cpb.UpdateContainerRequest
		wantID  string
		wantErr error
	}{
		{
			name:       "simple",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
				},
				false,
			),
		},
		{
			name:       "simple-with-async",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      true,
					},
				},
			},
			inAsync: true,
			wantID:  "some-instance",
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
				},
				true,
			),
		},
		{
			name:       "simple-with-error",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateError{
					UpdateError: &cpb.UpdateError{
						Details: "oh no!",
					},
				},
			},
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
				},
				false,
			),
			wantErr: status.Error(codes.Internal, "failed to update container: oh no!"),
		},
		{
			name:       "simple-with-ports",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inPorts:    []string{"1:1", "2:2"},
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
					Ports: []*cpb.StartContainerRequest_Port{
						{
							Internal: 1,
							External: 1,
						},
						{
							Internal: 2,
							External: 2,
						},
					},
				},
				false,
			),
		},
		{
			name:       "simple-with-envs",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inEnvs:     []string{"env1=cool", "env2=cooler"},
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
					Environment:  map[string]string{"env1": "cool", "env2": "cooler"},
				},
				false,
			),
		},
		{
			name:       "simple-with-envs-and-volumes",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inEnvs:     []string{"env1=cool", "env2=cooler"},
			inVols:     []string{"vol1:/aa", "vol2:/bb:ro"},
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
					Environment:  map[string]string{"env1": "cool", "env2": "cooler"},
					Volumes: []*cpb.Volume{
						{
							Name:       "vol1",
							MountPoint: "/aa",
						},
						{
							Name:       "vol2",
							MountPoint: "/bb",
							ReadOnly:   true,
						},
					},
				},
				false,
			),
		},
		{
			name:       "simple-with-devices",
			inImage:    "some-image",
			inTag:      "some-tag",
			inInstance: "some-instance",
			inCmd:      "some-cmd",
			inEnvs:     []string{"env1=cool", "env2=cooler"},
			inDevices:  []string{"dev1", "dev2:mydev2", "dev3:mydev3:rw"},
			inMsg: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
					},
				},
			},
			wantID: "some-instance",
			wantMsg: wrapInUpdateRequest(
				&cpb.StartContainerRequest{
					ImageName:    "some-image",
					Tag:          "some-tag",
					Cmd:          "some-cmd",
					InstanceName: "some-instance",
					Environment:  map[string]string{"env1": "cool", "env2": "cooler"},
					Devices: []*cpb.Device{
						{
							SrcPath:     "dev1",
							DstPath:     "dev1",
							Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE, cpb.Device_MKNOD},
						},
						{
							SrcPath:     "dev2",
							DstPath:     "mydev2",
							Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE, cpb.Device_MKNOD},
						},
						{
							SrcPath:     "dev3",
							DstPath:     "mydev3",
							Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE},
						},
					},
				},
				false,
			),
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeUpdatingContainerzServer{
				sendMsg: tc.inMsg,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			gotID, err := cli.UpdateContainer(ctx, tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inAsync, WithPorts(tc.inPorts), WithEnv(tc.inEnvs), WithVolumes(tc.inVols), WithDevices(tc.inDevices))
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("Start(%q, %q, %q, %q, %t, %v, %v) returned an unexpected error: %v", tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inAsync, tc.inPorts, tc.inEnvs, err)
				}
			}

			if tc.wantID != gotID {
				t.Errorf("Start(%q, %q, %q, %q, %t, %v, %v) returned incorrect id - want %s, got %s", tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inAsync, tc.inPorts, tc.inEnvs, tc.wantID, gotID)
			}

			if diff := cmp.Diff(err, tc.wantErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Start(%q, %q, %q, %q, %t, %v, %v) returned diff (-got, +want):\n%s", tc.inImage, tc.inTag, tc.inCmd, tc.inInstance, tc.inAsync, tc.inPorts, tc.inEnvs, diff)
			}
		})
	}
}

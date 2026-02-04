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

package server

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cpb "github.com/openconfig/gnoi/containerz"
)

func TestContainerStart(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.StartContainerRequest
		inOpts    []Option
		wantResp  *cpb.StartContainerResponse
		wantState *fakeContainerManager
		wantErr   error
	}{
		{
			name: "simple",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_PRIMARY,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
			},
		},
		{
			name: "ports",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_PRIMARY,
				Ports: []*cpb.StartContainerRequest_Port{
					{
						Internal: 1,
						External: 2,
					},
					{
						Internal: 3,
						External: 4,
					},
				},
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Ports: map[uint32]uint32{1: 2, 3: 4},
			},
		},
		{
			name: "env+port+instance",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_PRIMARY,
				Ports: []*cpb.StartContainerRequest_Port{
					{
						Internal: 1,
						External: 2,
					},
					{
						Internal: 3,
						External: 4,
					},
				},
				Environment:  map[string]string{"1": "2", "3": "4"},
				InstanceName: "some-instance",
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Envs:     map[string]string{"1": "2", "3": "4"},
				Ports:    map[uint32]uint32{1: 2, 3: 4},
				Instance: "some-instance",
			},
		},
		{
			name: "env",
			inReq: &cpb.StartContainerRequest{
				ImageName:   "some-image",
				Tag:         "some-tag",
				Cmd:         "some-cmd",
				Location:    cpb.StartContainerRequest_L_ALL,
				Environment: map[string]string{"1": "2", "3": "4"},
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_ALL.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Envs:  map[string]string{"1": "2", "3": "4"},
			},
		},
		{
			name: "volumes",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_BACKUP,
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
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_BACKUP.String()},
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
		},
		{
			name: "network",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Network:   "some-network",
				Location:  cpb.StartContainerRequest_L_PRIMARY,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image:   "some-image",
				Tag:     "some-tag",
				Cmd:     "some-cmd",
				Network: "some-network",
			},
		},
		{
			name: "capabilities",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Cap: &cpb.StartContainerRequest_Capabilities{
					Add:    []string{"cap1", "cap2"},
					Remove: []string{"cap3", "cap4"},
				},
				Location: cpb.StartContainerRequest_L_ALL,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_ALL.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Capabilities: &cpb.StartContainerRequest_Capabilities{
					Add:    []string{"cap1", "cap2"},
					Remove: []string{"cap3", "cap4"},
				},
			},
		},
		{
			name: "restart-policy",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_ALL,
				Restart: &cpb.StartContainerRequest_Restart{
					Policy:   cpb.StartContainerRequest_Restart_ON_FAILURE,
					Attempts: 3,
				},
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_ALL.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				RestartPolicy: &cpb.StartContainerRequest_Restart{
					Policy:   cpb.StartContainerRequest_Restart_ON_FAILURE,
					Attempts: 3,
				},
			},
		},
		{
			name: "run-as",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_ALL,
				RunAs: &cpb.StartContainerRequest_RunAs{
					User:  "some-user",
					Group: "some-group",
				},
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_ALL.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				RunAs: &cpb.StartContainerRequest_RunAs{
					User:  "some-user",
					Group: "some-group",
				},
			},
		},
		{
			name: "labels",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels:    map[string]string{"key1": "value1"},
				Location:  cpb.StartContainerRequest_L_ALL,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
				Labels: map[string]string{"key1": "value1",
					locationLabel: cpb.StartContainerRequest_L_ALL.String()},
			},
		},
		{
			name: "limits",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Limits: &cpb.StartContainerRequest_Limits{
					MaxCpu:       1.0,
					SoftMemBytes: 1000,
					HardMemBytes: 2000,
				},
				Location: cpb.StartContainerRequest_L_ALL,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels:     map[string]string{locationLabel: cpb.StartContainerRequest_L_ALL.String()},
				Image:      "some-image",
				Tag:        "some-tag",
				Cmd:        "some-cmd",
				CPU:        1.0,
				SoftMemory: 1000,
				HardMemory: 2000,
			},
		},
		{
			name: "devices",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels:    map[string]string{locationLabel: cpb.StartContainerRequest_L_ALL.String()},
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
				Location: cpb.StartContainerRequest_L_ALL,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Image:  "some-image",
				Tag:    "some-tag",
				Cmd:    "some-cmd",
				Labels: map[string]string{locationLabel: cpb.StartContainerRequest_L_ALL.String()},
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
		},
		{
			name: "location-unknown-set-in-map",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_UNKNOWN.String(),
				},
			},
			wantErr: status.Errorf(codes.InvalidArgument,
				"%q label (currently set to %q) should be not be set, or should match"+
					" location field %q. Unspecified location field is treated as L_PRIMARY",
				locationLabel, cpb.StartContainerRequest_L_UNKNOWN.String(),
				cpb.StartContainerRequest_L_PRIMARY.String()),
			wantState: &fakeContainerManager{},
		},
		{
			name: "non-location-label-set",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels:    map[string]string{"foo": "bar"},
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					"foo":         "bar",
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
			},
		},
		{
			name: "location-unknown-treated-as-primary",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
			},
		},
		{
			name: "mismatching-location",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_ALL.String(),
				},
				Location: cpb.StartContainerRequest_L_BACKUP,
			},
			wantState: &fakeContainerManager{},
			wantErr: status.Errorf(codes.InvalidArgument,
				"%q label (currently set to %q) should be not be set, or should match"+
					" location field %q. Unspecified location field is treated as L_PRIMARY",
				locationLabel, cpb.StartContainerRequest_L_ALL.String(),
				cpb.StartContainerRequest_L_BACKUP.String()),
		},
		{
			name: "location-label-only",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_ALL.String(),
				},
			},
			wantState: &fakeContainerManager{},
			wantErr: status.Errorf(codes.InvalidArgument,
				"%q label (currently set to %q) should be not be set, or should match"+
					" location field %q. Unspecified location field is treated as L_PRIMARY",
				locationLabel, cpb.StartContainerRequest_L_ALL.String(),
				cpb.StartContainerRequest_L_PRIMARY.String()),
		},
		{
			name: "matching-location",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String(),
				},
				Location: cpb.StartContainerRequest_L_PRIMARY,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
			},
		},
		{
			name: "location-field-only",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Location:  cpb.StartContainerRequest_L_PRIMARY,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					locationLabel: cpb.StartContainerRequest_L_PRIMARY.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
			},
		},
		{
			name: "location-label-among-other-labels",
			inReq: &cpb.StartContainerRequest{
				ImageName: "some-image",
				Tag:       "some-tag",
				Cmd:       "some-cmd",
				Labels: map[string]string{
					"x":           "y",
					locationLabel: cpb.StartContainerRequest_L_ALL.String(),
				},
				Location: cpb.StartContainerRequest_L_ALL,
			},
			wantResp: &cpb.StartContainerResponse{
				Response: &cpb.StartContainerResponse_StartOk{
					StartOk: &cpb.StartOK{},
				},
			},
			wantState: &fakeContainerManager{
				Labels: map[string]string{
					"x":           "y",
					locationLabel: cpb.StartContainerRequest_L_ALL.String()},
				Image: "some-image",
				Tag:   "some-tag",
				Cmd:   "some-cmd",
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

			resp, err := cli.StartContainer(ctx, tc.inReq)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected Start(%+v) to return error %v, got error: %v",
					tc.inReq, tc.wantErr, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Start(%+v) returned resp diff (-want +got):\n%s", tc.inReq, diff)
			}

			if diff := cmp.Diff(tc.wantState, fake, protocmp.Transform(), cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

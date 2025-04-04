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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

func TestContainerUpdate(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.UpdateContainerRequest
		inOpts    []Option
		wantResp  *cpb.UpdateContainerResponse
		wantState *fakeContainerManager
	}{
		{
			name: "simple-sync",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        false,
				Params: &cpb.StartContainerRequest{
					ImageName: "some-image",
					Tag:       "some-tag",
					Cmd:       "some-cmd",
				},
			},
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      false,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Async:    false,
			},
		},
		{
			name: "only-inner-image-and-tag-used",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "ignore-this-image",
				ImageTag:     "ignore-this-tag",
				Async:        false,
				Params: &cpb.StartContainerRequest{
					ImageName: "some-image",
					Tag:       "some-tag",
					Cmd:       "some-cmd",
				},
			},
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      false,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Async:    false,
			},
		},
		{
			name: "simple-async",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        true,
				Params: &cpb.StartContainerRequest{
					ImageName: "some-image",
					Tag:       "some-tag",
					Cmd:       "some-cmd",
				},
			},
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      true,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Async:    true,
			},
		},
		{
			name: "ports",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        false,
				Params: &cpb.StartContainerRequest{
					ImageName: "some-image",
					Tag:       "some-tag",
					Cmd:       "some-cmd",
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
			},
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      false,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Ports:    map[uint32]uint32{1: 2, 3: 4},
			},
		},
		{
			name: "env",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        false,
				Params: &cpb.StartContainerRequest{
					ImageName:   "some-image",
					Tag:         "some-tag",
					Cmd:         "some-cmd",
					Environment: map[string]string{"1": "2", "3": "4"},
				},
			},
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      false,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Envs:     map[string]string{"1": "2", "3": "4"},
			},
		},
		{
			name: "volumes",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        false,
				Params: &cpb.StartContainerRequest{
					ImageName: "some-image",
					Tag:       "some-tag",
					Cmd:       "some-cmd",
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
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      false,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
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
			name: "devices",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        false,
				Params: &cpb.StartContainerRequest{
					ImageName: "some-image",
					Tag:       "some-tag",
					Cmd:       "some-cmd",
					Devices: []*cpb.Device{
						{
							SrcPath:     "dev1",
							DstPath:     "my-dev1",
							Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE, cpb.Device_MKNOD},
						},
						{
							SrcPath:     "dev2",
							DstPath:     "my-dev2",
							Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE, cpb.Device_MKNOD},
						},
					},
				},
			},
			wantResp: &cpb.UpdateContainerResponse{
				Response: &cpb.UpdateContainerResponse_UpdateOk{
					UpdateOk: &cpb.UpdateOK{
						InstanceName: "some-instance",
						IsAsync:      false,
					},
				},
			},
			wantState: &fakeContainerManager{
				Instance: "some-instance",
				Image:    "some-image",
				Tag:      "some-tag",
				Cmd:      "some-cmd",
				Devices: []*cpb.Device{
					{
						SrcPath:     "dev1",
						DstPath:     "my-dev1",
						Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE, cpb.Device_MKNOD},
					},
					{
						SrcPath:     "dev2",
						DstPath:     "my-dev2",
						Permissions: []cpb.Device_Permission{cpb.Device_READ, cpb.Device_WRITE, cpb.Device_MKNOD},
					},
				},
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

			resp, err := cli.UpdateContainer(ctx, tc.inReq)
			if err != nil {
				t.Errorf("UpdateContainer(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("UpdateContainer(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}

			if diff := cmp.Diff(tc.wantState, fake, protocmp.Transform(), cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("UpdateContainer(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

func TestContainerUpdateError(t *testing.T) {
	tests := []struct {
		name      string
		inReq     *cpb.UpdateContainerRequest
		wantError error
	}{
		{
			name: "missing-start-req",
			inReq: &cpb.UpdateContainerRequest{
				InstanceName: "some-instance",
				ImageName:    "some-image",
				ImageTag:     "some-tag",
				Async:        false,
			},
			wantError: status.Errorf(codes.FailedPrecondition, "expected request to contain populated params, yet was nil"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fake := &fakeContainerManager{}
			opts := []Option{WithAddr("localhost:0")}
			cli, s := startServerAndReturnClient(ctx, t, fake, opts)
			defer s.Halt(ctx)

			_, err := cli.UpdateContainer(ctx, tc.inReq)
			if err == nil {
				t.Errorf("UpdateContainer(%+v) succeeded despite wanting error %v", tc.inReq, err)
			}
			if diff := cmp.Diff(tc.wantError, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("UpdateContainer(%+v) returned diff for error (-want +got):\n%s", tc.inReq, diff)
			}
		})
	}
}

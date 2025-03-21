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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeCreateVolumeContainerzServer struct {
	fakeContainerzServer

	receivedMsg *cpb.CreateVolumeRequest
	sendMsg     *cpb.CreateVolumeResponse
}

func (f *fakeCreateVolumeContainerzServer) CreateVolume(ctx context.Context, req *cpb.CreateVolumeRequest) (*cpb.CreateVolumeResponse, error) {
	f.receivedMsg = req
	return f.sendMsg, nil
}

func TestCreateVolume(t *testing.T) {
	tests := []struct {
		name      string
		inName    string
		inDriver  string
		inLabels  map[string]string
		inOptions map[string]string
		inMsg     *cpb.CreateVolumeResponse

		wantRequest *cpb.CreateVolumeRequest
		wantName    string
		wantErr     error
	}{
		{
			name:   "simple",
			inName: "simple",
			inMsg:  &cpb.CreateVolumeResponse{Name: "simple"},
			wantRequest: &cpb.CreateVolumeRequest{
				Name:    "simple",
				Driver:  cpb.Driver_DS_LOCAL,
				Options: &cpb.CreateVolumeRequest_LocalMountOptions{},
			},
			wantName: "simple",
		},
		{
			name:      "custom-driver",
			inName:    "simple",
			inDriver:  "custom-driver",
			inOptions: map[string]string{"foo": "bar"},
			inLabels:  map[string]string{"label1": "value1", "label2": "value2"},
			inMsg:     &cpb.CreateVolumeResponse{Name: "simple"},
			wantRequest: &cpb.CreateVolumeRequest{
				Name:   "simple",
				Driver: cpb.Driver_DS_CUSTOM,
				Labels: map[string]string{"label1": "value1", "label2": "value2"},
				Options: &cpb.CreateVolumeRequest_CustomOptions{
					CustomOptions: &cpb.CustomOptions{
						Options: map[string]string{"foo": "bar"},
					},
				},
			},
			wantName: "simple",
		},
		{
			name:      "good-driver-wrong-option",
			inName:    "simple",
			inDriver:  "local",
			inOptions: map[string]string{"foo": "bar"},
			inMsg:     &cpb.CreateVolumeResponse{Name: "simple"},
			wantErr:   fmt.Errorf("invalid key: %q", "foo"),
		},
		{
			name:      "good-driver-wrong-type",
			inName:    "simple",
			inDriver:  "local",
			inOptions: map[string]string{"type": "bar"},
			inMsg:     &cpb.CreateVolumeResponse{Name: "simple"},
			wantErr:   fmt.Errorf("invalid type: %q", "bar"),
		},
		{
			name:     "bind-mount",
			inName:   "simple",
			inDriver: "local",
			inOptions: map[string]string{
				"type":       "none",
				"options":    "opt1,opt2",
				"mountpoint": "/here",
			},
			wantRequest: &cpb.CreateVolumeRequest{
				Name:   "simple",
				Driver: cpb.Driver_DS_LOCAL,
				Options: &cpb.CreateVolumeRequest_LocalMountOptions{
					LocalMountOptions: &cpb.LocalDriverOptions{
						Type:       cpb.LocalDriverOptions_TYPE_NONE,
						Options:    []string{"opt1", "opt2"},
						Mountpoint: "/here"},
				},
				Labels: map[string]string{},
			},
			wantName: "simple",
			inMsg:    &cpb.CreateVolumeResponse{Name: "simple"},
		},
		{
			name:     "bind-mount-with-labels",
			inName:   "simple",
			inDriver: "local",
			inLabels: map[string]string{"label1": "value1", "label2": "value2"},
			inOptions: map[string]string{
				"type":       "none",
				"options":    "opt1,opt2",
				"mountpoint": "/here",
			},
			wantRequest: &cpb.CreateVolumeRequest{
				Name:   "simple",
				Driver: cpb.Driver_DS_LOCAL,
				Options: &cpb.CreateVolumeRequest_LocalMountOptions{
					LocalMountOptions: &cpb.LocalDriverOptions{
						Type:       cpb.LocalDriverOptions_TYPE_NONE,
						Options:    []string{"opt1", "opt2"},
						Mountpoint: "/here"},
				},
				Labels: map[string]string{"label1": "value1", "label2": "value2"},
			},
			wantName: "simple",
			inMsg:    &cpb.CreateVolumeResponse{Name: "simple"},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeCreateVolumeContainerzServer{
				sendMsg: tc.inMsg,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			resp, err := cli.CreateVolume(ctx, tc.inName, tc.inDriver, tc.inLabels, tc.inOptions)
			if err != nil {
				if tc.wantErr == nil {
					t.Fatalf("CreateVolume(%q, %q, %v, %v) returned an unexpected error: %v", tc.inName, tc.inDriver, tc.inLabels, tc.inOptions, err)
				}
			}

			if tc.wantName != resp {
				t.Errorf("CreateVolume(%q, %q, %v, %v) = %q, want %q", tc.inName, tc.inDriver, tc.inLabels, tc.inOptions, resp, tc.wantName)
			}

			if diff := cmp.Diff(fcm.receivedMsg, tc.wantRequest, protocmp.Transform()); diff != "" {
				t.Errorf("CreateVolume(%q, %q, %v, %v) received diff (-got, +want):\n%s", tc.inName, tc.inDriver, tc.inLabels, tc.inOptions, diff)
			}

			if err != nil && tc.wantErr != nil {
				if diff := cmp.Diff(err.Error(), tc.wantErr.Error()); diff != "" {
					t.Errorf("CreateVolume(%q, %q, %v, %v) returned diff (-got, +want):\n%s", tc.inName, tc.inDriver, tc.inLabels, tc.inOptions, diff)
				}
			}
		})
	}
}

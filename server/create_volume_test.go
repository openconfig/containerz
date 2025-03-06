package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	cpb "github.com/openconfig/gnoi/containerz"
)

func TestCreateVolume(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		inReq     *cpb.CreateVolumeRequest
		inOpts    []Option
		wantName  string
		wantState *fakeContainerManager
		wantResp  *cpb.CreateVolumeResponse
	}{
		{
			name:     "empty request",
			inReq:    &cpb.CreateVolumeRequest{},
			inOpts:   []Option{},
			wantName: "some-volume",
			wantState: &fakeContainerManager{
				VolumeDriver: cpb.Driver_DS_LOCAL,
			},
			wantResp: &cpb.CreateVolumeResponse{
				Name: "some-volume",
			},
		},
		{
			name: "with options",
			inReq: &cpb.CreateVolumeRequest{
				Name:   "some-volume",
				Driver: cpb.Driver_DS_LOCAL,
				Options: &cpb.CreateVolumeRequest_LocalMountOptions{
					LocalMountOptions: &cpb.LocalDriverOptions{
						Mountpoint: "some-mountpoint",
					},
				},
				Labels: map[string]string{
					"some-label": "some-value",
				},
			},
			inOpts:   []Option{},
			wantName: "",
			wantState: &fakeContainerManager{
				VolumeDriver: cpb.Driver_DS_LOCAL,
				VolumeOpts: &cpb.LocalDriverOptions{
					Mountpoint: "some-mountpoint",
				},
				VolumeLabel: map[string]string{
					"some-label": "some-value",
				},
			},
			wantResp: &cpb.CreateVolumeResponse{
				Name: "some-volume",
			},
		},
		{
			name: "with driver and options",
			inReq: &cpb.CreateVolumeRequest{
				Name:   "some-volume",
				Driver: cpb.Driver_DS_CUSTOM,
				Options: &cpb.CreateVolumeRequest_CustomOptions{
					CustomOptions: &cpb.CustomOptions{
						Options: map[string]string{
							"some-option": "some-value",
						},
					},
				},
				Labels: map[string]string{
					"some-label": "some-value",
				},
			},
			inOpts:   []Option{},
			wantName: "",
			wantState: &fakeContainerManager{
				VolumeDriver: cpb.Driver_DS_CUSTOM,
				VolumeOpts: &cpb.CustomOptions{
					Options: map[string]string{
						"some-option": "some-value",
					},
				},
				VolumeLabel: map[string]string{
					"some-label": "some-value",
				},
			},
			wantResp: &cpb.CreateVolumeResponse{
				Name: "some-volume",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeContainerManager{
				createVolumeName: tc.wantName,
			}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			resp, err := cli.CreateVolume(ctx, tc.inReq)
			if err != nil {
				t.Errorf("CreateVolume(%+v) returned error: %v", tc.inReq, err)
			}

			if diff := cmp.Diff(tc.wantResp, resp, protocmp.Transform()); diff != "" {
				t.Errorf("Start(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fake, protocmp.Transform(), cmpopts.IgnoreUnexported(fakeContainerManager{}), cmpopts.SortMaps(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("List(%+v) returned diff (-want +got):\n%s", tc.inReq, diff)
				}
			}
		})
	}
}

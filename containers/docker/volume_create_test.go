package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/volume"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeVolumeCreatingDocker struct {
	fakeDocker
	V volume.Volume
}

func (f *fakeVolumeCreatingDocker) VolumeCreate(_ context.Context, opts volume.CreateOptions) (volume.Volume, error) {
	vol := volume.Volume{
		Name:    opts.Name,
		Driver:  opts.Driver,
		Options: opts.DriverOpts,
		Labels:  opts.Labels,
	}
	if vol.Name == "" {
		vol.Name = "test-volume"
	}

	f.V = vol
	return vol, nil
}

func TestVolumeCreate(t *testing.T) {
	tests := []struct {
		name      string
		inOpts    []options.Option
		inName    string
		inDriver  cpb.Driver
		wantState *fakeVolumeCreatingDocker
		wantResp  string
	}{
		{
			name:     "no-name",
			wantResp: "test-volume",
			wantState: &fakeVolumeCreatingDocker{
				V: volume.Volume{
					Name:    "test-volume",
					Driver:  "local",
					Options: map[string]string{},
				},
			},
		},
		{
			name:     "named-volume",
			inName:   "some-volume",
			wantResp: "some-volume",
			wantState: &fakeVolumeCreatingDocker{
				V: volume.Volume{
					Name:    "some-volume",
					Driver:  "local",
					Options: map[string]string{},
				},
			},
		},
		{
			name:     "named-volume-with-opts",
			inName:   "some-volume",
			wantResp: "some-volume",
			inDriver: cpb.Driver_DS_LOCAL,
			inOpts: []options.Option{
				options.WithVolumeDriverOpts(&cpb.LocalDriverOptions{
					Type:       cpb.LocalDriverOptions_TYPE_NONE,
					Options:    []string{"some-option"},
					Mountpoint: "some-mountpoint",
				}),
				options.WithVolumeLabels(map[string]string{"some-label": "some-label"}),
			},
			wantState: &fakeVolumeCreatingDocker{
				V: volume.Volume{
					Name:   "some-volume",
					Driver: "local",
					Options: map[string]string{
						"type":   "none",
						"o":      "some-option",
						"device": "some-mountpoint",
					},
					Labels: map[string]string{"some-label": "some-label"},
				},
			},
		},
		{
			name:     "named-volume-with-opts-and-custom-driver",
			inName:   "some-volume",
			wantResp: "some-volume",
			inDriver: cpb.Driver_DS_CUSTOM,
			inOpts: []options.Option{
				options.WithVolumeDriverOpts(&cpb.CustomOptions{
					Options: map[string]string{"some-option": "some-value"},
				}),
				options.WithVolumeLabels(map[string]string{"some-label": "some-label"}),
			},
			wantState: &fakeVolumeCreatingDocker{
				V: volume.Volume{
					Name:   "some-volume",
					Driver: "custom",
					Options: map[string]string{
						"some-option": "some-value",
					},
					Labels: map[string]string{"some-label": "some-label"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fcd := &fakeVolumeCreatingDocker{}
			mgr := New(fcd)

			resp, err := mgr.VolumeCreate(ctx, tc.inName, tc.inDriver, tc.inOpts...)
			if err != nil {
				t.Errorf("VolumeCreate(%s, %v, %+v) returned error: %v", tc.inName, tc.inDriver, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fcd, cmpopts.IgnoreUnexported(fakeVolumeCreatingDocker{})); diff != "" {
					t.Errorf("VolumeCreate(%s, %v, %+v) returned diff(-want, +got):\n%s", tc.inName, tc.inDriver, tc.inOpts, diff)
				}
			}

			if diff := cmp.Diff(tc.wantResp, resp); diff != "" {
				t.Errorf("VolumeCreate(%s, %v, %+v) returned diff(-want, +got):\n%s", tc.inName, tc.inDriver, tc.inOpts, diff)
			}
		})
	}
}

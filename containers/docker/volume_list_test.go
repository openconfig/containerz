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

package docker

import (
	"context"
	"testing"
	"time"

	tpb "google3/google/protobuf/timestamp_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/moby/moby/v/v24/api/types/filters"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/volume/volume"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeListVolumeStreamer struct {
	msgs []*cpb.ListVolumeResponse
}

func (f *fakeListVolumeStreamer) Send(msg *cpb.ListVolumeResponse) error {
	f.msgs = append(f.msgs, msg)
	return nil
}

type fakeVolumeListingDocker struct {
	fakeDocker
	volumes []*volume.Volume

	Opts volume.ListOptions
}

func (f *fakeVolumeListingDocker) VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error) {
	f.Opts = options

	return volume.ListResponse{
		Volumes: f.volumes,
	}, nil
}

func TestListVolume(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2024-02-09T13:07:31+01:00")
	if err != nil {
		t.Fatalf("time.Parse(%q) returned error: %v", "2024-02-09T13:07:31+01:00", err)
	}
	tests := []struct {
		name      string
		inOpts    []options.Option
		inVols    []*volume.Volume
		wantState *fakeVolumeListingDocker
		wantMsgs  []*cpb.ListVolumeResponse
	}{
		{
			name: "no-containers",
			wantState: &fakeVolumeListingDocker{
				Opts: volume.ListOptions{},
			},
		},
		{
			name: "containers-no-filter",
			wantState: &fakeVolumeListingDocker{
				Opts: volume.ListOptions{},
			},
			inVols: []*volume.Volume{
				&volume.Volume{
					Name:   "some-volume",
					Driver: "some-driver",
					Options: map[string]string{
						"some-option": "some-option",
					},
					CreatedAt: "2024-02-09T13:07:31+01:00",
				},
				&volume.Volume{
					Name:   "some-other-volume",
					Driver: "some-other-driver",
					Options: map[string]string{
						"some-other-option": "some-other-option",
					},
					CreatedAt: "2024-02-09T13:07:31+01:00",
				},
			},
			wantMsgs: []*cpb.ListVolumeResponse{
				&cpb.ListVolumeResponse{
					Name:   "some-volume",
					Driver: "some-driver",
					Options: map[string]string{
						"some-option": "some-option",
					},
					Created: tpb.New(ts),
				},
				&cpb.ListVolumeResponse{
					Name:   "some-other-volume",
					Driver: "some-other-driver",
					Options: map[string]string{
						"some-other-option": "some-other-option",
					},
					Created: tpb.New(ts),
				},
			},
		},
		{
			name: "filter",
			inOpts: []options.Option{
				options.WithFilter(map[options.FilterKey][]string{
					options.Volume: []string{"some-volume"},
				}),
			},
			wantState: &fakeVolumeListingDocker{
				Opts: volume.ListOptions{
					Filters: filters.NewArgs(filters.Arg("volme", "some-volume")),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fsd := &fakeVolumeListingDocker{
				volumes: tc.inVols,
			}
			mgr := New(fsd)

			stream := &fakeListVolumeStreamer{}

			if err := mgr.VolumeList(ctx, stream, tc.inOpts...); err != nil {
				t.Errorf("VolumeList(%+v) returned error: %v", tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fsd, cmpopts.IgnoreUnexported(fakeVolumeListingDocker{}, filters.Args{})); diff != "" {
					t.Errorf("VolumeList( %+v) returned diff(-want, +got):\n%s", tc.inOpts, diff)
				}
			}

			if diff := cmp.Diff(tc.wantMsgs, stream.msgs, protocmp.Transform()); diff != "" {
				t.Errorf("VolumeList(%+v) returned diff(-want, +got):\n%s", tc.inOpts, diff)
			}
		})
	}
}

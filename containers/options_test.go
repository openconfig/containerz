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

package options

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
	tpb "github.com/openconfig/gnoi/types"
)

type fakeStream struct{}

func (f fakeStream) Send(_ *cpb.DeployResponse) error { return nil }

func TestWithTarget(t *testing.T) {
	p := &options{}

	WithTarget("my-image", "my-tag")(p)

	if p.TargetName != "my-image" && p.TargetTag != "my-tag" {
		t.Errorf("WithTarget(my-image, my-tag) returned incorrect values: %+v", p)
	}

	WithTarget("my-image", "")(p)
	if p.TargetName != "my-image" && p.TargetTag != "latest" {
		t.Errorf("WithTarget(my-image, my-tag) returned incorrect values: %+v", p)
	}
}

func TestWithRegistryAuth(t *testing.T) {
	p := &options{}

	WithRegistryAuth(&tpb.Credentials{})(p)

	if p.Credentials == nil {
		t.Errorf("WithRegistryAuth(tpb.Credentials{}) did not set the credential")
	}
}

func TestWithStream(t *testing.T) {
	p := &options{}

	WithStream(&fakeStream{})(p)

	if p.StreamClient == nil {
		t.Errorf("WithStream(fakeStream) did not set the streamClient")
	}
}

func TestWithInstanceName(t *testing.T) {
	p := &options{}

	WithInstanceName("some-instance")(p)

	if p.InstanceName != "some-instance" {
		t.Errorf("WithInstanceName(some-instance) did not set the instance name")
	}
}

func TestWithPorts(t *testing.T) {
	p := &options{}

	in := map[uint32]uint32{1: 2}
	WithPorts(in)(p)

	if diff := cmp.Diff(p.PortMapping, in); diff != "" {
		t.Errorf("WithPorts(%v) returned diff (-got, +want):\n%s", in, diff)
	}
}

func TestWithEnv(t *testing.T) {
	p := &options{}

	in := map[string]string{"a": "b"}
	WithEnv(in)(p)

	if diff := cmp.Diff(p.EnvMapping, in); diff != "" {
		t.Errorf("WithEnv(%v) returned diff (-got, +want):\n%s", in, diff)
	}
}

func TestForce(t *testing.T) {
	p := &options{}

	Force()(p)

	if !p.Force {
		t.Errorf("Force() did not set the force flag")
	}
}

func TestFollow(t *testing.T) {
	p := &options{}

	Follow()(p)

	if !p.Follow {
		t.Errorf("Follow() did not set the follow flag")
	}
}

func TestWithSince(t *testing.T) {
	p := &options{}

	WithSince(time.Second)(p)

	if p.Since != time.Second {
		t.Errorf("WithSince(time.Second) did not set the since field")
	}
}

func TestWithUntil(t *testing.T) {
	p := &options{}

	WithUntil(time.Second)(p)

	if p.Until != time.Second {
		t.Errorf("WithUntil(time.Second) did not set the since field")
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		inOpts []Option
		want   *options
	}{
		{
			inOpts: []Option{WithTarget("my-image", "my-tag")},
			want:   &options{TargetName: "my-image", TargetTag: "my-tag"},
		},
		{
			inOpts: []Option{WithTarget("my-image", "")},
			want:   &options{TargetName: "my-image", TargetTag: "latest"},
		},
		{
			inOpts: []Option{WithTarget("my-image", "my-tag"), WithRegistryAuth(&tpb.Credentials{})},
			want:   &options{TargetName: "my-image", TargetTag: "my-tag", Credentials: &tpb.Credentials{}},
		},
		{
			inOpts: []Option{WithTarget("my-image", "my-tag"), WithRegistryAuth(&tpb.Credentials{}), WithStream(&fakeStream{})},
			want:   &options{TargetName: "my-image", TargetTag: "my-tag", Credentials: &tpb.Credentials{}, StreamClient: &fakeStream{}},
		},
	}

	for _, tc := range tests {
		got := ApplyOptions(tc.inOpts...)
		if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
			t.Errorf("ApplyOptions(%v) returned an unexpected diff (-want +got): %v", tc.inOpts, diff)
		}
	}
}

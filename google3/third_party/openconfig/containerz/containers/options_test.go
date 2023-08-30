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
	p := &imageOptions{}

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
	p := &imageOptions{}

	WithRegistryAuth(&tpb.Credentials{})(p)

	if p.Credentials == nil {
		t.Errorf("WithRegistryAuth(tpb.Credentials{}) did not set the credential")
	}
}

func TestWithStream(t *testing.T) {
	p := &imageOptions{}

	WithStream(&fakeStream{})(p)

	if p.StreamClient == nil {
		t.Errorf("WithStream(fakeStream) did not set the streamClient")
	}
}

func TestWithInstanceName(t *testing.T) {
	p := &imageOptions{}

	WithInstanceName("some-instance")(p)

	if p.InstanceName != "some-instance" {
		t.Errorf("WithInstanceName(some-instance) did not set the instance name")
	}
}

func TestWithPorts(t *testing.T) {
	p := &imageOptions{}

	in := map[uint32]uint32{1: 2}
	WithPorts(in)(p)

	if diff := cmp.Diff(p.PortMapping, in); diff != "" {
		t.Errorf("WithPorts(%v) returned diff (-got, +want):\n%s", in, diff)
	}
}

func TestWithEnv(t *testing.T) {
	p := &imageOptions{}

	in := map[string]string{"a": "b"}
	WithEnv(in)(p)

	if diff := cmp.Diff(p.EnvMapping, in); diff != "" {
		t.Errorf("WithEnv(%v) returned diff (-got, +want):\n%s", in, diff)
	}
}

func TestForce(t *testing.T) {
	p := &imageOptions{}

	Force()(p)

	if !p.Force {
		t.Errorf("Force() did not set the force flag")
	}
}

func TestFollow(t *testing.T) {
	p := &imageOptions{}

	Follow()(p)

	if !p.Follow {
		t.Errorf("Follow() did not set the follow flag")
	}
}

func TestWithSince(t *testing.T) {
	p := &imageOptions{}

	WithSince(time.Second)(p)

	if p.Since != time.Second {
		t.Errorf("WithSince(time.Second) did not set the since field")
	}
}

func TestWithUntil(t *testing.T) {
	p := &imageOptions{}

	WithUntil(time.Second)(p)

	if p.Until != time.Second {
		t.Errorf("WithUntil(time.Second) did not set the since field")
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		inOpts []ImageOption
		want   *imageOptions
	}{
		{
			inOpts: []ImageOption{WithTarget("my-image", "my-tag")},
			want:   &imageOptions{TargetName: "my-image", TargetTag: "my-tag"},
		},
		{
			inOpts: []ImageOption{WithTarget("my-image", "")},
			want:   &imageOptions{TargetName: "my-image", TargetTag: "latest"},
		},
		{
			inOpts: []ImageOption{WithTarget("my-image", "my-tag"), WithRegistryAuth(&tpb.Credentials{})},
			want:   &imageOptions{TargetName: "my-image", TargetTag: "my-tag", Credentials: &tpb.Credentials{}},
		},
		{
			inOpts: []ImageOption{WithTarget("my-image", "my-tag"), WithRegistryAuth(&tpb.Credentials{}), WithStream(&fakeStream{})},
			want:   &imageOptions{TargetName: "my-image", TargetTag: "my-tag", Credentials: &tpb.Credentials{}, StreamClient: &fakeStream{}},
		},
	}

	for _, tc := range tests {
		got := ApplyOptions(tc.inOpts...)
		if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
			t.Errorf("ApplyOptions(%v) returned an unexpected diff (-want +got): %v", tc.inOpts, diff)
		}
	}
}

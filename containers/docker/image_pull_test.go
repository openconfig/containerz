package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/image"
	"github.com/moby/moby/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
	tpb "github.com/openconfig/gnoi/types"
)

type fakePullingDocker struct {
	fakeDocker
	ImageRef  string
	SourceRef string
	TargetRef string
}

func (f *fakePullingDocker) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	f.ImageRef = ref
	jm := &jsonmessage.JSONMessage{
		Progress: &jsonmessage.JSONProgress{
			Current: 10,
		},
	}

	buf, err := json.Marshal(jm)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(buf)), nil

}

func (f *fakePullingDocker) ImageTag(ctx context.Context, source, target string) error {
	f.SourceRef = source
	f.TargetRef = target
	return nil
}

type fakeStream struct {
	resps []*cpb.DeployResponse
}

func (f *fakeStream) Send(resp *cpb.DeployResponse) error {
	f.resps = append(f.resps, resp)
	return nil
}

func TestImagePull(t *testing.T) {
	fakeStream := &fakeStream{}
	tests := []struct {
		name      string
		inImage   string
		inTag     string
		inOpts    []options.Option
		wantState *fakePullingDocker
		wantResp  []*cpb.DeployResponse
		wantErr   error
	}{
		{
			name:    "empty-image-name",
			wantErr: status.Error(codes.InvalidArgument, "an image name must be supplied."),
		},
		{
			name:    "empty-tag",
			inImage: "some-image",
			wantState: &fakePullingDocker{
				ImageRef: "some-image:latest",
			},
		},
		{
			name:    "non-nil-creds",
			inImage: "some-image",
			inOpts:  []options.Option{options.WithRegistryAuth(&tpb.Credentials{})},
			wantErr: status.Error(codes.Unimplemented, "registry auth not yet implemented"),
		},
		{
			name: "pull-with-tag",
			inOpts: []options.Option{
				options.WithTarget("another-name", "another-tag"),
			},
			inImage: "some-image",
			wantState: &fakePullingDocker{
				ImageRef:  "some-image:latest",
				SourceRef: "some-image:latest",
				TargetRef: "another-name:another-tag",
			},
		},
		{
			name: "pull-with-stream",
			inOpts: []options.Option{
				options.WithStream(fakeStream),
			},
			inImage: "some-image",
			wantState: &fakePullingDocker{
				ImageRef: "some-image:latest",
			},
			wantResp: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 10,
						},
					},
				},
			},
		},
		{
			name: "pull-with-stream-and-tag",
			inOpts: []options.Option{
				options.WithStream(fakeStream),
				options.WithTarget("another-name", "another-tag"),
			},
			inImage: "some-image",
			wantState: &fakePullingDocker{
				ImageRef:  "some-image:latest",
				SourceRef: "some-image:latest",
				TargetRef: "another-name:another-tag",
			},
			wantResp: []*cpb.DeployResponse{
				&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferProgress{
						ImageTransferProgress: &cpb.ImageTransferProgress{
							BytesReceived: 10,
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeStream.resps = nil
			fd := &fakePullingDocker{}
			mgr := New(fd)

			if err := mgr.ImagePull(context.Background(), tc.inImage, tc.inTag, tc.inOpts...); err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ImagePull(%q, %q) returned unexpected error(-want, got):\n %s", tc.inImage, tc.inTag, diff)
					}
					return
				}
				t.Errorf("ImagePull(%q, %q) returned error: %v", tc.inImage, tc.inTag, err)
			}

			if diff := cmp.Diff(tc.wantState, fd, cmpopts.IgnoreUnexported(fakePullingDocker{})); diff != "" {
				t.Errorf("ImagePull(%q, %q) returned diff(-want, +got):\n%s", tc.inImage, tc.inTag, diff)
			}

			if diff := cmp.Diff(tc.wantResp, fakeStream.resps, cmpopts.EquateEmpty(), protocmp.Transform()); diff != "" {
				t.Errorf("ImagePull(%q, %q) returned diff(-want, +got):\n%s", tc.inImage, tc.inTag, diff)
			}
		})
	}
}

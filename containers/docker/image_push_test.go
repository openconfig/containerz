package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"
	"google3/third_party/golang/github_com/moby/moby/v/v24/pkg/jsonmessage/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google3/third_party/openconfig/containerz/containers/options"
)

type fakePushingDocker struct {
	fakeDocker
	image, tag string
	isJSON     bool

	Source, Target string
}

func (f fakePushingDocker) ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
	jm := &jsonmessage.JSONMessage{
		Stream: fmt.Sprintf("Loaded image: %s\n", f.image+":"+f.tag),
	}

	data, err := json.Marshal(jm)
	if err != nil {
		return types.ImageLoadResponse{}, err
	}

	return types.ImageLoadResponse{
		Body: io.NopCloser(bytes.NewBuffer(data)),
		JSON: f.isJSON,
	}, nil
}

func (f *fakePushingDocker) ImageTag(ctx context.Context, source, target string) error {
	f.Source = source
	f.Target = target
	return nil
}

func TestContainerPush(t *testing.T) {
	tests := []struct {
		name               string
		inOpts             []options.ImageOption
		inFile             *os.File
		inImage, inTag     string
		isJSON             bool
		wantState          *fakePushingDocker
		wantImage, wantTag string
		wantErr            error
	}{
		{
			name:    "nil-file",
			wantErr: status.Error(codes.InvalidArgument, "file must be supplied"),
		},
		{
			name:      "plain-load",
			isJSON:    true,
			inFile:    os.Stdin, // using os.Stdin as a stand in for a file. We don't read from it.
			inImage:   "some-image",
			inTag:     "some-tag",
			wantImage: "some-image",
			wantTag:   "some-tag",
		},
		{
			name:    "plain-load-with-tagging",
			isJSON:  true,
			inFile:  os.Stdin,
			inImage: "some-image",
			inTag:   "some-tag",
			inOpts:  []options.ImageOption{options.WithTarget("another-image", "another-tag")},
			wantState: &fakePushingDocker{
				Source: "some-image:some-tag",
				Target: "another-image:another-tag",
			},
			wantImage: "another-image",
			wantTag:   "another-tag",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fd := &fakePushingDocker{
				isJSON: tc.isJSON,
				tag:    tc.inTag,
				image:  tc.inImage,
			}
			mgr := New(fd)

			gotImage, gotTag, err := mgr.ContainerPush(context.Background(), tc.inFile, tc.inOpts...)
			if err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ContainerPush(file) returned unexpected error(-want, got):\n %s", diff)
					}
					return
				}
				t.Errorf("ContainerPush(file) returned error: %v", err)
			}

			if gotImage != tc.wantImage || gotTag != tc.wantTag {
				t.Errorf("ContainerPush(file) returned wrong info; want %s/%s, got %s/%s", tc.wantImage, tc.wantTag, gotImage, gotTag)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fd, cmpopts.IgnoreUnexported(fakePushingDocker{})); diff != "" {
					t.Errorf("ContainerPush(file) returned diff(-want, +got):\n%s", diff)
				}
			}
		})
	}
}

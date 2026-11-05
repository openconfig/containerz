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
	"github.com/docker/docker/client"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/image"
	"github.com/moby/moby/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

type fakePushingDocker struct {
	fakeDocker
	image, tag string
	isJSON     bool
	// msgs, if set, replaces the default "Loaded image" message.
	msgs []*jsonmessage.JSONMessage

	Source, Target string
}

func (f fakePushingDocker) ImageLoad(ctx context.Context, input io.Reader, options ...client.ImageLoadOption) (image.LoadResponse, error) {
	msgs := f.msgs
	if msgs == nil {
		msgs = []*jsonmessage.JSONMessage{{
			Stream: fmt.Sprintf("Loaded image: %s\n", f.image+":"+f.tag),
		}}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, jm := range msgs {
		if err := enc.Encode(jm); err != nil {
			return image.LoadResponse{}, err
		}
	}

	return image.LoadResponse{
		Body: io.NopCloser(&buf),
		JSON: f.isJSON,
	}, nil
}

func (f *fakePushingDocker) ImageTag(ctx context.Context, source, target string) error {
	f.Source = source
	f.Target = target
	return nil
}

func TestImagePush(t *testing.T) {
	tests := []struct {
		name               string
		inOpts             []options.Option
		inFile             *os.File
		inImage, inTag     string
		inMsgs             []*jsonmessage.JSONMessage
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
			inOpts:  []options.Option{options.WithTarget("another-image", "another-tag")},
			wantState: &fakePushingDocker{
				Source: "some-image:some-tag",
				Target: "another-image:another-tag",
			},
			wantImage: "another-image",
			wantTag:   "another-tag",
		},
		{
			name:   "existing-image-warning-then-loaded",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "The image some-image:latest already exists, renaming the old one with ID sha256:1234 to empty string\n"},
				{Stream: "Loaded image: some-image:latest\n"},
			},
			wantImage: "some-image",
			wantTag:   "latest",
		},
		{
			name:   "existing-image-warning-then-loaded-with-tagging",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "The image some-image:latest already exists, renaming the old one with ID sha256:1234 to empty string\n"},
				{Stream: "Loaded image: some-image:latest\n"},
			},
			inOpts: []options.Option{options.WithTarget("another-image", "another-tag")},
			wantState: &fakePushingDocker{
				Source: "some-image:latest",
				Target: "another-image:another-tag",
			},
			wantImage: "another-image",
			wantTag:   "another-tag",
		},
		{
			name:   "warning-and-loaded-in-one-message",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "The image some-image:latest already exists, renaming the old one with ID sha256:1234 to empty string\nLoaded image: some-image:latest\n"},
			},
			wantImage: "some-image",
			wantTag:   "latest",
		},
		{
			name:   "registry-with-port",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "Loaded image: registry:5000/some-image:some-tag\n"},
			},
			wantImage: "registry:5000/some-image",
			wantTag:   "some-tag",
		},
		{
			name:   "no-tag-defaults-to-latest",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "Loaded image: registry:5000/some-image\n"},
			},
			wantImage: "registry:5000/some-image",
			wantTag:   "latest",
		},
		{
			name:   "image-id-only-with-tagging",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "Loaded image ID: sha256:1234\n"},
			},
			inOpts: []options.Option{options.WithTarget("another-image", "another-tag")},
			wantState: &fakePushingDocker{
				Source: "sha256:1234",
				Target: "another-image:another-tag",
			},
			wantImage: "another-image",
			wantTag:   "another-tag",
		},
		{
			name:   "image-id-only-without-target",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "Loaded image ID: sha256:1234\n"},
			},
			wantErr: status.Error(codes.InvalidArgument, "image was loaded as sha256:1234 without a name and tag; a target name must be provided"),
		},
		{
			name:   "error-detail",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Error: &jsonmessage.JSONError{Message: "archive/tar: invalid tar header"}},
			},
			wantErr: status.Error(codes.Internal, "unable to load image: archive/tar: invalid tar header"),
		},
		{
			name:   "error-after-loaded",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "Loaded image: some-image:latest\n"},
				{ErrorMessage: "something went wrong"},
			},
			wantErr: status.Error(codes.Internal, "unable to load image: something went wrong"),
		},
		{
			name:    "empty-stream",
			isJSON:  true,
			inFile:  os.Stdin,
			inMsgs:  []*jsonmessage.JSONMessage{},
			wantErr: status.Error(codes.Internal, "image load response did not report a loaded image"),
		},
		{
			name:   "empty-stream-with-tagging",
			isJSON: true,
			inFile: os.Stdin,
			inMsgs: []*jsonmessage.JSONMessage{
				{Stream: "\n"},
			},
			inOpts:  []options.Option{options.WithTarget("another-image", "another-tag")},
			wantErr: status.Error(codes.Internal, "image load response did not report a loaded image"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fd := &fakePushingDocker{
				isJSON: tc.isJSON,
				tag:    tc.inTag,
				image:  tc.inImage,
				msgs:   tc.inMsgs,
			}
			mgr := New(fd)

			gotImage, gotTag, err := mgr.ImagePush(context.Background(), tc.inFile, tc.inOpts...)
			if err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ImagePush(file) returned unexpected error(-want, got):\n %s", diff)
					}
					return
				}
				t.Errorf("ImagePush(file) returned error: %v", err)
			}

			if tc.wantErr != nil {
				t.Fatalf("ImagePush(file) returned no error, want %v", tc.wantErr)
			}

			if gotImage != tc.wantImage || gotTag != tc.wantTag {
				t.Errorf("ImagePush(file) returned wrong info; want %s/%s, got %s/%s", tc.wantImage, tc.wantTag, gotImage, gotTag)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fd, cmpopts.IgnoreUnexported(fakePushingDocker{})); diff != "" {
					t.Errorf("ImagePush(file) returned diff(-want, +got):\n%s", diff)
				}
			}
		})
	}
}

package docker

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"/containers/options"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeStreamer struct {
	msgs []string
}

func (f *fakeStreamer) Send(msg *cpb.LogResponse) error {
	f.msgs = append(f.msgs, msg.GetMsg())
	return nil
}

type fakeLoggingDocker struct {
	fakeDocker
	cnts  []types.Container
	inMsg string

	Instance string
	Follow   bool
}

func (f fakeLoggingDocker) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return f.cnts, nil
}

func (f *fakeLoggingDocker) ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	f.Instance = container
	f.Follow = options.Follow

	return io.NopCloser(bytes.NewBuffer([]byte(f.inMsg))), nil
}

func TestContainerLogs(t *testing.T) {
	tests := []struct {
		name       string
		inTimeout  time.Duration
		inOpts     []options.ImageOption
		inInstance string
		inMsg      string
		inCnts     []types.Container
		wantState  *fakeLoggingDocker
		wantMsgs   []string
		wantErr    error
	}{
		{
			name:       "no-such-instance",
			inInstance: "no-such-instance",
			wantErr:    status.Errorf(codes.NotFound, "container no-such-instance not found"),
		},
		{
			name:       "instance-with-logs",
			inInstance: "instance-with-logs",
			inCnts: []types.Container{
				types.Container{
					Names: []string{"/instance-with-logs"},
				},
			},
			inMsg: "we have the logs",
			wantState: &fakeLoggingDocker{
				Instance: "instance-with-logs",
				Follow:   false,
			},
			wantMsgs: []string{"we have the logs"},
		},
		{
			name:       "instance-follow-with-logs",
			inInstance: "instance-with-logs",
			inOpts:     []options.ImageOption{options.Follow()},
			inCnts: []types.Container{
				types.Container{
					Names: []string{"/instance-with-logs"},
				},
			},
			inMsg: "we have the logs",
			wantState: &fakeLoggingDocker{
				Instance: "instance-with-logs",
				Follow:   true,
			},
			wantMsgs: []string{"we have the logs"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fsd := &fakeLoggingDocker{
				cnts:  tc.inCnts,
				inMsg: tc.inMsg,
			}
			mgr := New(fsd)

			stream := &fakeStreamer{}

			if err := mgr.ContainerLogs(ctx, tc.inInstance, stream, tc.inOpts...); err != nil {
				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
						t.Errorf("ContainerLogs(%q, %+v) returned unexpected error(-want, got):\n %s", tc.inInstance, tc.inOpts, diff)
					}
					return
				}
				t.Errorf("ContainerLogs(%q, %+v) returned error: %v", tc.inInstance, tc.inOpts, err)
			}

			if tc.wantState != nil {
				if diff := cmp.Diff(tc.wantState, fsd, cmpopts.IgnoreUnexported(fakeLoggingDocker{})); diff != "" {
					t.Errorf("ContainerLogs(%q, %+v) returned diff(-want, +got):\n%s", tc.inInstance, tc.inOpts, diff)
				}
			}

			if diff := cmp.Diff(tc.wantMsgs, stream.msgs); diff != "" {
				t.Errorf("ContainerLogs(%q, %+v) returned diff(-want, +got):\n%s", tc.inInstance, tc.inOpts, diff)
			}
		})
	}
}

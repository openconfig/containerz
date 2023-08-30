package server

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google3/third_party/openconfig/containerz/containers/options"
	commonpb "google3/third_party/openconfig/gnoi/common/common_go_proto"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeContainerManager struct {
	Image    string
	Tag      string
	Contents string
	Cmd      string
	Instance string
	Ports    map[uint32]uint32
	Envs     map[string]string
	Force    bool
	Follow   bool
	All      bool
	Limit    int32

	listMsgs    []*cpb.ListResponse
	msgs        []string
	removeError error
}

func (f *fakeContainerManager) ContainerPull(ctx context.Context, image string, tag string, opts ...options.ImageOption) error {
	f.Image = image
	f.Tag = tag
	return nil
}

func (f *fakeContainerManager) ContainerPush(ctx context.Context, file *os.File, opts ...options.ImageOption) (string, string, error) {
	buf, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	f.Contents = string(buf)
	return "", "", nil
}

func (f fakeContainerManager) ContainerRemove(context.Context, string, string, ...options.ImageOption) error {
	return f.removeError
}

func (f *fakeContainerManager) ContainerStart(_ context.Context, image string, tag string, cmd string, opts ...options.ImageOption) (string, error) {
	optionz := options.ApplyOptions(opts...)
	f.Image = image
	f.Tag = tag
	f.Cmd = cmd
	f.Ports = optionz.PortMapping
	f.Envs = optionz.EnvMapping
	f.Instance = optionz.InstanceName
	return "", nil
}

func (f *fakeContainerManager) ContainerStop(_ context.Context, instance string, opts ...options.ImageOption) error {
	optionz := options.ApplyOptions(opts...)
	f.Instance = instance
	f.Force = optionz.Force
	return nil
}

func (f *fakeContainerManager) ContainerList(ctx context.Context, all bool, limit int32, srv options.ListStreamer, opts ...options.ImageOption) error {
	f.All = all
	f.Limit = limit

	for _, msg := range f.listMsgs {
		if err := srv.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeContainerManager) ContainerLogs(_ context.Context, instance string, srv options.LogStreamer, opts ...options.ImageOption) error {
	optionz := options.ApplyOptions(opts...)

	f.Follow = optionz.Follow
	for _, msg := range f.msgs {
		if err := srv.Send(&cpb.LogResponse{
			Msg: msg,
		}); err != nil {
			return err
		}
	}

	return nil
}

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name          string
		inOpts        []Option
		inReqs        []*cpb.DeployRequest
		wantResponses []*cpb.DeployResponse
		wantState     *fakeContainerManager
		wantErr       error
	}{
		{
			name:    "invalid-protocol",
			inOpts:  []Option{WithAddr("localhost:0")},
			inReqs:  buildRequests(t, &cpb.ImageTransferEnd{}),
			wantErr: status.Error(codes.Unavailable, "must send send a TransferImage message first"),
		},
		{
			name:   "gigantic-contents",
			inOpts: []Option{WithAddr("localhost:0")},
			inReqs: buildRequests(t, &cpb.ImageTransfer{
				Name:      "some-image",
				Tag:       "some-tag",
				ImageSize: 1e16,
			}),
			wantErr: status.Error(codes.ResourceExhausted, "not enough space to store image"),
		},
		{
			name:   "remote-download",
			inOpts: []Option{WithAddr("localhost:0")},
			inReqs: buildRequests(t, &cpb.ImageTransfer{
				Name:           "some-image",
				Tag:            "some-tag",
				RemoteDownload: &commonpb.RemoteDownload{},
			}),
			wantState: &fakeContainerManager{
				Image: "some-image",
				Tag:   "some-tag",
			},
			wantResponses: buildResponses(t, &cpb.ImageTransferSuccess{
				Name: "some-image",
				Tag:  "some-tag",
			}),
		},
		{
			name:   "too-much-data-sent",
			inOpts: []Option{WithAddr("localhost:0"), WithChunkSize(8)},
			inReqs: buildRequests(t, &cpb.ImageTransfer{
				Name:      "some-image",
				Tag:       "some-tag",
				ImageSize: 16,
			}, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_Content{
					Content: []byte("exactly "),
				},
			}, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_Content{
					Content: []byte("16 bytes"),
				},
			}, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_Content{
					Content: []byte("16 bytes"),
				},
			}),
			wantResponses: buildResponses(t, &cpb.ImageTransferReady{
				ChunkSize: 8,
			}, &cpb.ImageTransferProgress{
				BytesReceived: 8,
			}, &cpb.ImageTransferProgress{
				BytesReceived: 16,
			}, &cpb.ImageTransferSuccess{
				ImageSize: 16,
			}),
			wantErr: status.Errorf(codes.InvalidArgument, "too much data received"),
		},
		{
			name:   "successful-image-transfer",
			inOpts: []Option{WithAddr("localhost:0"), WithChunkSize(8)},
			inReqs: buildRequests(t, &cpb.ImageTransfer{
				Name:      "some-image",
				Tag:       "some-tag",
				ImageSize: 16,
			}, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_Content{
					Content: []byte("exactly "),
				},
			}, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_Content{
					Content: []byte("16 bytes"),
				},
			}, &cpb.ImageTransferEnd{}),
			wantResponses: buildResponses(t, &cpb.ImageTransferReady{
				ChunkSize: 8,
			}, &cpb.ImageTransferProgress{
				BytesReceived: 8,
			}, &cpb.ImageTransferProgress{
				BytesReceived: 16,
			}, &cpb.ImageTransferSuccess{
				ImageSize: 16,
			}),
			wantState: &fakeContainerManager{
				Contents: "exactly 16 bytes",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeContainerManager{}
			cli, s := startServerAndReturnClient(ctx, t, fake, tc.inOpts)
			defer s.Halt(ctx)

			dCli, err := cli.Deploy(ctx)
			if err != nil {
				t.Errorf("Deploy(ctx) returned error: %v", err)
			}
			defer dCli.CloseSend()

			msgIndex := 0
			if err := dCli.Send(tc.inReqs[msgIndex]); err != nil {
				t.Errorf("Send(%v) returned error: %v", tc.inReqs[msgIndex], err)
			}

			for {
				msg, err := dCli.Recv()
				if err != nil {
					if err == io.EOF {
						return
					}
					if tc.wantErr != nil {
						if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
							t.Errorf("Send returned the wrong error: want: %v, got: %v", tc.wantErr, err)
						}
						return
					}
					t.Errorf("Recv() returned error: %v", err)
				}

				if diff := cmp.Diff(tc.wantResponses[msgIndex], msg, protocmp.Transform()); diff != "" {
					t.Errorf("Recv() returned diff at index %d (-want, +got):\n%s", msgIndex, diff)
				}

				if _, ok := msg.GetResponse().(*cpb.DeployResponse_ImageTransferSuccess); ok {
					break
				}

				msgIndex++
				if err := dCli.Send(tc.inReqs[msgIndex]); err != nil {
					t.Errorf("Send(%v) returned error: %v", tc.inReqs[msgIndex], err)
				}
			}

			if diff := cmp.Diff(tc.wantState, fake, cmpopts.IgnoreUnexported(fakeContainerManager{})); diff != "" {
				t.Errorf("Deploy(ctx) returned diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func buildRequests(t *testing.T, msgs ...proto.Message) []*cpb.DeployRequest {
	t.Helper()

	requests := []*cpb.DeployRequest{}

	for _, msg := range msgs {
		switch m := msg.(type) {
		case *cpb.DeployRequest:
			requests = append(requests, m)
		case *cpb.ImageTransfer:
			requests = append(requests, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_ImageTransfer{
					ImageTransfer: m,
				},
			})
		case *cpb.ImageTransferEnd:
			requests = append(requests, &cpb.DeployRequest{
				Request: &cpb.DeployRequest_ImageTransferEnd{
					ImageTransferEnd: m,
				},
			})
		default:
			t.Fatalf("unknown type %T", m)
		}
	}

	return requests
}

func buildResponses(t *testing.T, msg ...proto.Message) []*cpb.DeployResponse {
	t.Helper()

	reponses := []*cpb.DeployResponse{}

	for _, msg := range msg {
		switch m := msg.(type) {
		case *cpb.ImageTransferSuccess:
			reponses = append(reponses, &cpb.DeployResponse{
				Response: &cpb.DeployResponse_ImageTransferSuccess{
					ImageTransferSuccess: m,
				},
			})
		case *cpb.ImageTransferProgress:
			reponses = append(reponses, &cpb.DeployResponse{
				Response: &cpb.DeployResponse_ImageTransferProgress{
					ImageTransferProgress: m,
				},
			})
		case *cpb.ImageTransferReady:
			reponses = append(reponses, &cpb.DeployResponse{
				Response: &cpb.DeployResponse_ImageTransferReady{
					ImageTransferReady: m,
				},
			})
		default:
			t.Fatalf("unknown type %T", m)
		}
	}

	return reponses
}

func startServerAndReturnClient(ctx context.Context, t *testing.T, fake *fakeContainerManager, opts []Option) (cpb.ContainerzClient, *Server) {
	t.Helper()
	s := New(fake, opts...)
	go func() {
		if err := s.Serve(ctx); err != nil {
			t.Logf("failed to serve containerz service: %v", err)
		}
	}()

	addr := s.lis.Addr().String()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("received error when dialing grpcServer(%v): got err: %v, want: nil\n", addr, err)
	}

	cli := cpb.NewContainerzClient(conn)
	return cli, s
}

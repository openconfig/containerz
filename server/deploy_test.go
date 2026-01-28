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

package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	options "github.com/openconfig/containerz/containers"
	commonpb "github.com/openconfig/gnoi/common"
	cpb "github.com/openconfig/gnoi/containerz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

type fakeContainerManager struct {
	Image         string
	Tag           string
	Contents      string
	Cmd           string
	Instance      string
	Name          string
	Config        string
	Ports         map[uint32]uint32
	Envs          map[string]string
	Force         bool
	Follow        bool
	All           bool
	Async         bool
	Limit         int32
	Devices       []*cpb.Device
	Volumes       []*cpb.Volume
	VolumeDriver  cpb.Driver
	VolumeOpts    proto.Message
	VolumeLabel   map[string]string
	Network       string
	Capabilities  proto.Message
	RunAs         proto.Message
	RestartPolicy proto.Message
	Labels        map[string]string
	CPU           float64
	HardMemory    int64
	SoftMemory    int64

	listVols         []*cpb.ListVolumeResponse
	listCntMsgs      []*cpb.ListContainerResponse
	listImgMsgs      []*cpb.ListImageResponse
	listPluginMsgs   *cpb.ListPluginsResponse
	createVolumeName string
	msgs             []string

	removeError error
}

func (f *fakeContainerManager) ImagePull(ctx context.Context, image string, tag string, opts ...options.Option) error {
	f.Image = image
	f.Tag = tag
	return nil
}

func (f *fakeContainerManager) ImagePush(ctx context.Context, file *os.File, opts ...options.Option) (string, string, error) {
	buf, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	f.Contents = string(buf)
	return "", "", nil
}

func (f *fakeContainerManager) ContainerRemove(_ context.Context, instance string, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)
	f.Instance = instance
	f.Force = optionz.Force

	for _, cnt := range f.listCntMsgs {
		if cnt.GetName() == instance {
			return nil
		}
	}

	return fmt.Errorf("not found")
}

func (f *fakeContainerManager) ContainerStart(_ context.Context, image string, tag string, cmd string, opts ...options.Option) (string, error) {
	optionz := options.ApplyOptions(opts...)
	f.Image = image
	f.Tag = tag
	f.Cmd = cmd
	f.Ports = optionz.PortMapping
	f.Envs = optionz.EnvMapping
	f.Instance = optionz.InstanceName
	f.Volumes = optionz.Volumes
	f.Devices = optionz.Devices
	f.Network = optionz.Network
	f.Capabilities = optionz.Capabilities
	f.RunAs = optionz.RunAs
	f.RestartPolicy = optionz.RestartPolicy
	f.Labels = optionz.Labels
	f.CPU = optionz.CPU
	f.HardMemory = optionz.HardMemory
	f.SoftMemory = optionz.SoftMemory
	return "", nil
}

func (f *fakeContainerManager) ContainerStop(_ context.Context, instance string, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)
	f.Instance = instance
	f.Force = optionz.Force
	return nil
}

func (f *fakeContainerManager) ContainerList(ctx context.Context, all bool, limit int32, srv options.ListContainerStreamer, opts ...options.Option) error {
	f.All = all
	f.Limit = limit

	for _, msg := range f.listCntMsgs {
		if err := srv.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeContainerManager) ContainerUpdate(_ context.Context, instance, image, tag, cmd string, async bool, opts ...options.Option) (string, error) {
	optionz := options.ApplyOptions(opts...)
	f.Instance = instance
	f.Image = image
	f.Tag = tag
	f.Cmd = cmd
	f.Async = async
	f.Ports = optionz.PortMapping
	f.Envs = optionz.EnvMapping
	f.Volumes = optionz.Volumes
	f.Devices = optionz.Devices
	f.Labels = optionz.Labels
	f.Network = optionz.Network
	f.Capabilities = optionz.Capabilities
	f.RunAs = optionz.RunAs
	f.RestartPolicy = optionz.RestartPolicy
	return instance, nil
}

func (f *fakeContainerManager) ContainerLogs(_ context.Context, instance string, srv options.LogStreamer, opts ...options.Option) error {
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

func (f *fakeContainerManager) ImageList(ctx context.Context, all bool, limit int32, srv options.ListImageStreamer, opts ...options.Option) error {
	f.All = all
	f.Limit = limit

	for _, msg := range f.listImgMsgs {
		if err := srv.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (f fakeContainerManager) ImageRemove(context.Context, string, string, ...options.Option) error {
	return f.removeError
}

func (f *fakeContainerManager) PluginList(ctx context.Context, instance string) (*cpb.ListPluginsResponse, error) {
	f.Instance = instance
	return f.listPluginMsgs, nil
}

func (f *fakeContainerManager) PluginRemove(ctx context.Context, instance string) error {
	f.Instance = instance
	for _, plugin := range f.listPluginMsgs.GetPlugins() {
		if plugin.GetInstanceName() == instance {
			return nil
		}
	}
	return status.Errorf(codes.NotFound, "plugin %s not found", instance)
}

func (f *fakeContainerManager) PluginStart(ctx context.Context, name, instance, config string) error {
	f.Name = name
	f.Instance = instance
	f.Config = config
	return nil
}

func (f *fakeContainerManager) PluginStop(ctx context.Context, instance string) error {
	f.Instance = instance
	for _, plugin := range f.listPluginMsgs.GetPlugins() {
		if plugin.GetInstanceName() == instance {
			return nil
		}
	}
	return status.Errorf(codes.NotFound, "plugin %s not found", instance)
}

func (f *fakeContainerManager) VolumeList(ctx context.Context, srv options.ListVolumeStreamer, opts ...options.Option) error {
	for _, msg := range f.listVols {
		if err := srv.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeContainerManager) VolumeCreate(ctx context.Context, name string, driver cpb.Driver, opts ...options.Option) (string, error) {
	optionz := options.ApplyOptions(opts...)

	f.VolumeOpts = optionz.VolumeDriverOptions
	f.VolumeLabel = optionz.VolumeLabels
	if f.createVolumeName == "" {
		f.createVolumeName = name
	}
	f.VolumeDriver = driver
	return f.createVolumeName, nil
}

func (f *fakeContainerManager) VolumeRemove(ctx context.Context, name string, opts ...options.Option) error {
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
		{
			name:   "successful-plugin-transfer",
			inOpts: []Option{WithAddr("localhost:0"), WithChunkSize(8)},
			inReqs: buildRequests(t, &cpb.ImageTransfer{
				Name:      "some-image",
				Tag:       "some-tag",
				ImageSize: 16,
				IsPlugin:  true,
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
				Name:      "some-image",
				ImageSize: 16,
			}),
			wantState: &fakeContainerManager{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pluginLocation = t.TempDir()
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

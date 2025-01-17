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

// Package server implements the containerz gNOI service.
package server

import (
	"context"
	"net"
	"os"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"github.com/openconfig/containerz/containers"

	cpb "github.com/openconfig/gnoi/containerz"
)

type containerManager interface {

	// ContainerList list the containers on the target.
	//
	// It takes:
	// all (bool): return all containers regardless of state
	// limit (int32): return limit number of results.
	//
	// It returns an error indicating the result of the operation.
	ContainerList(context.Context, bool, int32, options.ListContainerStreamer, ...options.Option) error

	// ContainerPull pulls a container from a registry to this instance of containerz.
	//
	// It takes as input:
	// - image (string): a container image name.
	// - tag (string): a container image tag.
	// - opts (ImageOption slice): a set of options.
	//
	// It returns an error indicating the result of the operation.
	ContainerPull(context.Context, string, string, ...options.Option) error

	// ContainerPush pushes a container image to this instance of containerz.
	//
	// It takes as input:
	// - file (os.File): a file containing the tarball of the container.
	// - opts (ImageOption slice): a set of options.
	//
	// It returns:
	// - image (string): the container image name of the container that was pushed.
	// - tag (string): the container image tag of the container that was pushed
	ContainerPush(context.Context, *os.File, ...options.Option) (string, string, error)

	// ContainerRemove removes an container provided that it is not running.
	//
	// It takes:
	// - container (string): the container name to remove.
	//
	// It returns an error indicating if the remove operation succeeded.
	ContainerRemove(context.Context, string, ...options.Option) error

	// ContainerStart starts a container based on the supplied image and tag.
	//
	// It takes:
	// - image (string): the image to use
	// - tag (string): the tag to use
	// - cmd (string): a command to run.
	//
	// It returns an error indicating if the start operation succeeded along with the ID of the
	// started container.
	ContainerStart(context.Context, string, string, string, ...options.Option) (string, error)

	// ContainerStop stops a container. If the Force option is passed it will forcefully stop
	// (kill) the container. A stop timeout can be provided via the context otherwise the
	// system default will be used.
	//
	// It takes:
	// - instance (string): the instance name of the running container.
	//
	// It returns an error indicating whether the result was successful
	ContainerStop(context.Context, string, ...options.Option) error

	// ContainerUpdates updates an existing container.
	//
	// It takes:
	// - instance (string): the instance name of the running container.
	// - image (string): the image to use.
	// - tag (string): the tag to use.
	// - cmd (string): a command to run.
	// - async (bool): whether to run immediately and perform the update asynchronously.
	//
	// It returns an error indicating if the start operation succeeded along with the ID of the
	// started container.
	ContainerUpdate(ctx context.Context, instance, image, tag, cmd string, async bool, opts ...options.Option) (string, error)

	// ContainerLogs fetches the logs from a container. It can optionally follow the logs
	// and send them back to the client.
	//
	// It takes:
	// - instance (string): the instance name of the container.
	// - srv (LogStreamer): to stream the logs back to the client.
	//
	// It returns an error indicating whether the operation was successful or not.
	ContainerLogs(context.Context, string, options.LogStreamer, ...options.Option) error

	// ImageList lists the images on the target.
	//
	// It takes:
	// all (bool): return all containers regardless of state
	// limit (int32): return limit number of results.
	//
	// It returns an error indicating the result of the operation.
	ImageList(context.Context, bool, int32, options.ListImageStreamer, ...options.Option) error

	// ImageRemove removes an image provided it is not linked to any running containers.
	//
	// It takes:
	// - image (string): the image name to remove.
	// - tag (string): the tage to remove
	//
	// It returns an error indicating if the remove operation succeeded.
	ImageRemove(context.Context, string, string, ...options.Option) error

	// PluginList lists plugins on the target.
	//
	// It takes:
	// - instance (string): the instance name of the plugin to list.
	//
	// It returns an error indicating whether the operation was successful or not and a list of
	// plugins.
	PluginList(context.Context, string) (*cpb.ListPluginsResponse, error)

	// PluginRemove removes a plugin from the target.
	//
	// It takes:
	// - instance (string): the instance name of the plugin to remove.
	//
	// It returns an error indicating whether the operation was successful or not.
	PluginRemove(context.Context, string) error

	// PluginStart starts a plugin on the target.
	//
	// It takes:
	// - name (string): the name of the plugin to start.
	// - instance (string): the instance name of the plugin to start.
	// - config (string): the configuration to use for the plugin.
	//
	// It returns an error indicating whether the operation was successful or not.
	PluginStart(context.Context, string, string, string) error

	// PluginStop stops a plugin on the target.
	//
	// It takes:
	// - instance (string): the instance name of the plugin to stop.
	//
	// It returns an error indicating whether the operation was successful or not.
	PluginStop(context.Context, string) error

	// VolumeList lists volumes on the target.
	//
	// It takes:
	// - optional filters
	//
	// It returns an error indicating the result of the operation.
	VolumeList(context.Context, options.ListVolumeStreamer, ...options.Option) error

	// VolumeCreate creates a volume. It will optionally apply labels or driver options to the volume
	// creation.
	//
	// It takes:
	// -  name (string): The name of the volume to create. If this this empty the name should be
	//		autogenerated.
	// -  driver (enum): The name of the driver to use. If this is empty the driver will default
	//		local driver.
	VolumeCreate(context.Context, string, cpb.Driver, ...options.Option) (string, error)

	// VolumeRemove removes a volume identified by the provided name.
	//
	// It takes:
	// - name (string): The name of the volume to remove.
	VolumeRemove(context.Context, string, ...options.Option) error
}

// Server represents a containerz service.
type Server struct {
	cpb.UnimplementedContainerzServer

	mgr        containerManager
	grpcServer *grpc.Server
	lis        net.Listener

	addr        string
	dockerHost  string
	tmpLocation string

	chunkSize int
}

// New constructs a new containerz server
func New(mgr containerManager, opts ...Option) *Server {
	s := &Server{
		grpcServer:  grpc.NewServer(),
		tmpLocation: "/tmp",
		chunkSize:   5e6, // 5mb chunks
		mgr:         mgr,
		addr:        ":9999",
	}

	for _, opt := range opts {
		opt(s)
	}

	var err error
	s.lis, err = net.Listen("tcp", s.addr)
	if err != nil {
		klog.Fatalf("server start: %e", err)
	}

	return s
}

// Serve starts this instance of the containerz server.
func (s *Server) Serve(ctx context.Context) error {
	klog.Info("server-start")
	cpb.RegisterContainerzServer(s.grpcServer, s)

	klog.Infof("Starting up on Containerz server, listening on: %s", s.lis.Addr())
	klog.Info("server-ready")
	return s.grpcServer.Serve(s.lis)
}

// Halt stops the containerz server gracefully.
func (s *Server) Halt(ctx context.Context) {
	klog.Info("server-stopping")
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	klog.Info("server stopped")
}

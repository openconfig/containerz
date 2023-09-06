// Package server implements the containerz gNOI service.
package server

import (
	"context"
	"net"
	"os"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"/containers/options"

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
	ContainerList(context.Context, bool, int32, options.ListStreamer, ...options.ImageOption) error

	// ContainerPull pulls a container from a registry to this instance of containerz.
	//
	// It takes as input:
	// - image (string): a container image name.
	// - tag (string): a container image tag.
	// - opts (ImageOption slice): a set of options.
	//
	// It returns an error indicating the result of the operation.
	ContainerPull(context.Context, string, string, ...options.ImageOption) error

	// ContainerPush pushes a container image to this instance of containerz.
	//
	// It takes as input:
	// - file (os.File): a file containing the tarball of the container.
	// - opts (ImageOption slice): a set of options.
	//
	// It returns:
	// - image (string): the container image name of the container that was pushed.
	// - tag (string): the container image tag of the container that was pushed
	ContainerPush(context.Context, *os.File, ...options.ImageOption) (string, string, error)

	// ContainerRemove removes an image provided it is not linked to any running containers.
	//
	// It takes:
	// - image (string): the image name to remove.
	// - tag (string): the tage to remove
	//
	// It returns an error indicating if the remove operation succeeded.
	ContainerRemove(context.Context, string, string, ...options.ImageOption) error

	// ContainerStart starts a container based on the supplied image and tag.
	//
	// It takes:
	// - image (string): the image to use
	// - tag (string): the tag to use
	// - cmd (string): a command to run.
	//
	// It returns an error indicating if the start operation succeeded along with the ID of the
	// started container.
	ContainerStart(context.Context, string, string, string, ...options.ImageOption) (string, error)

	// ContainerStop stops a container. If the Force option is passed it will forcefully stop
	// (kill) the container. A stop timeout can be provided via the context otherwise the
	// system default will be used.
	//
	// It takes:
	// - instance (string): the instance name of the running container.
	//
	// It returns an error indicating whether the result was successful
	ContainerStop(context.Context, string, ...options.ImageOption) error

	// ContainerLogs fetches the logs from a container. It can optionally follow the logs
	// and send them back to the client.
	//
	// It takes:
	// - instance (string): the instance name of the container.
	// - srv (LogStreamer): to stream the logs back to the client.
	//
	// It returns an error indicating whether the operation was successful or not.
	ContainerLogs(context.Context, string, options.LogStreamer, ...options.ImageOption) error
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

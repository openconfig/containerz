package client

import (
	"context"
	"testing"

	"google3/net/grpc/go/grpctest"
	"google.golang.org/grpc"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeContainerzServer struct {
	cpb.UnimplementedContainerzServer
}

func TestNewClient(t *testing.T) {
	addr := newServer(t, &fakeContainerzServer{})
	client, err := NewClient(context.Background(), addr)
	if err != nil {
		t.Fatalf("NewClient(%q) returned error: %v", addr, err)
	}

	if client.cli == nil {
		t.Errorf("NewClient(%q) did not initialize the client", addr)
	}
}

func newServer(t *testing.T, srv cpb.ContainerzServer) string {
	s := grpc.NewServer()
	cpb.RegisterContainerzServer(s, srv)
	return grpctest.StartServerT(t, s)
}

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

package client

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	"google.golang.org/grpc"

	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeContainerzServer struct {
	cpb.UnimplementedContainerzServer
}

func TestNewClient(t *testing.T) {
	addr, stop := newServer(t, &fakeContainerzServer{})
	defer stop()
	client, err := NewClient(context.Background(), addr)
	if err != nil {
		t.Fatalf("NewClient(%q) returned error: %v", addr, err)
	}

	if client.cli == nil {
		t.Errorf("NewClient(%q) did not initialize the client", addr)
	}
}

func TLSCreds() (string, string) {
	td := "testdata"
	return filepath.Join(td, "server.cert"), filepath.Join(td, "server.key")
}

func newServer(t *testing.T, srv cpb.ContainerzServer) (string, func()) {
	t.Helper()
	Dial = grpc.DialContext

	s := grpc.NewServer()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen on localhost:0, %v", err)
	}

	cpb.RegisterContainerzServer(s, srv)
	go s.Serve(l)
	return l.Addr().String(), s.Stop
}

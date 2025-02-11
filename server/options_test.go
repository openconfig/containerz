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
	"testing"

	"google.golang.org/grpc"
)

func TestWithAddr(t *testing.T) {
	s := &Server{}

	WithAddr("cool-addr")(s)

	if s.addr != "cool-addr" {
		t.Errorf("WithAddr('cool-addr') returned %s", s.addr)
	}
}

func TestWithTempLocation(t *testing.T) {
	s := &Server{}

	WithTempLocation("cool-location")(s)

	if s.tmpLocation != "cool-location" {
		t.Errorf("WithTempLocation('cool-location') returned %s", s.tmpLocation)
	}
}

func TestWithChunkSize(t *testing.T) {
	s := &Server{}

	WithChunkSize(10)(s)

	if s.chunkSize != 10 {
		t.Errorf("WithChunkSize(10) returned %d", s.chunkSize)
	}
}

func TestWithGrpcServer(t *testing.T) {
	s := &Server{}

	srv := grpc.NewServer()
	WithGrpcServer(srv)(s)

	if s.grpcServer != srv {
		t.Fatal("WithGrpcServer did not set first server")
	}

	srv2 := grpc.NewServer()
	WithGrpcServer(srv2)(s)
	if s.grpcServer != srv2 {
		t.Fatal("WithGrpcServer did not set second server")
	}
}

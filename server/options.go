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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/alts"
)

// Option represents an server option.
type Option func(*Server)

// WithAddr sets the listen for the server.
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithTempLocation sets the location where image files should be uploaded to.
func WithTempLocation(tmp string) Option {
	return func(s *Server) {
		s.tmpLocation = tmp
	}
}

// WithChunkSize sets the chunkSize supported by the server
func WithChunkSize(chunkSize int) Option {
	return func(s *Server) {
		s.chunkSize = chunkSize
	}
}

// UseALTS sets up the grpc server to use ALTS authentication.
// See https://cloud.google.com/docs/security/encryption-in-transit/application-layer-transport-security
// for more information.
func UseALTS() Option {
	return WithGrpcServer(
		grpc.NewServer(grpc.Creds(alts.NewServerCreds(alts.DefaultServerOptions()))),
	)
}

// WithGrpcServer sets the gRPC server that will host the containerz service
// WithGrpcServer may be called multiple times during New as options are processed.
// Each call to WithGrpcServer will stop the previously-running server. (it also ensures
// channelz entries are removed even if the server was not running).
func WithGrpcServer(g *grpc.Server) Option {
	return func(s *Server) {
		if s.grpcServer != nil {
			s.grpcServer.Stop()
		}
		s.grpcServer = g
	}
}

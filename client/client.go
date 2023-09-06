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

// Package client is a containerz grpc client.
package client

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"
	cpb "github.com/openconfig/gnoi/containerz"
)

var (
	dial = grpc.DialContext
)

// Client is a grpc containerz client.
type Client struct {
	cli cpb.ContainerzClient
}

// NewClient builds a new containerz client.
func NewClient(ctx context.Context, addr string) (*Client, error) {
	tls := &tls.Config{InsecureSkipVerify: true} // NOLINT
	tlsCred := grpc.WithTransportCredentials(credentials.NewTLS(tls))
	conn, err := dial(ctx, addr, tlsCred)
	if err != nil {
		return nil, err
	}

	return &Client{
		cli: cpb.NewContainerzClient(conn),
	}, nil
}

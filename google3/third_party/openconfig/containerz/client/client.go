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

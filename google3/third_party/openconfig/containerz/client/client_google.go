package client

import (
	"context"

	"google3/base/go/google"
	"google3/net/grpc/go/grpcprod"
	"google.golang.org/grpc"
)

const (
	transparentProxy = "/abns/ggn-engprod-jobs/grpc-transparent-proxy?drain-aware"
)

func prodDialer(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	return grpcprod.Dial(transparentProxy, nil, opts...)
}

func init() {
	google.Init()
	dial = prodDialer
}

package grpc

import (
	"context"
	"time"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"google.golang.org/grpc"
)

type Client struct {
}

func CC(address string) (*grpc.ClientConn, error) {
	var grpcOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_opentracing.UnaryClientInterceptor(),
			grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(5), grpc_retry.WithBackoff(grpc_retry.BackoffExponential(1*time.Second))),
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			grpc_opentracing.StreamClientInterceptor(),
			grpc_retry.StreamClientInterceptor(grpc_retry.WithMax(5), grpc_retry.WithBackoff(grpc_retry.BackoffExponential(1*time.Second))),
		)),
	}
	grpcOpts = append(grpcOpts, grpc.WithInsecure())
	return grpc.DialContext(context.TODO(), address, grpcOpts...)
}

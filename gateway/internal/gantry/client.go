package gantry

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ratdaddy/blockcloset/gateway/internal/requestid"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

type Client struct {
	buckets servicev1.GantryServiceClient
	cc      *grpc.ClientConn
}

func New(ctx context.Context, address string, opts ...grpc.DialOption) (*Client, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(requestIDUnaryInterceptor()),
	}

	dialOpts = append(dialOpts, opts...)

	cc, err := grpc.NewClient(address, dialOpts...)

	if err != nil {
		return nil, err
	}
	return &Client{buckets: servicev1.NewGantryServiceClient(cc), cc: cc}, nil
}

func requestIDUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id := requestid.RequestIDFromContext(ctx); id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", id)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *Client) Close() error {
	if c.cc != nil {
		return c.cc.Close()
	}
	return nil
}

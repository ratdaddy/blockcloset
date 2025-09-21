package gantry

import (
	"context"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	buckets servicev1.GantryServiceClient
	cc      *grpc.ClientConn
}

func New(ctx context.Context, address string) (*Client, error) {
	cc, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	return &Client{buckets: servicev1.NewGantryServiceClient(cc), cc: cc}, nil
}

func (c *Client) Close() error {
	if c.cc != nil {
		return c.cc.Close()
	}
	return nil
}

package cradle

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

type Client struct {
	svc servicev1.CradleServiceClient
	cc  *grpc.ClientConn
}

func New(ctx context.Context, address string, opts ...grpc.DialOption) (*Client, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	dialOpts = append(dialOpts, opts...)

	cc, err := grpc.NewClient(address, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{svc: servicev1.NewCradleServiceClient(cc), cc: cc}, nil
}

func (c *Client) Close() error {
	if c.cc != nil {
		return c.cc.Close()
	}
	return nil
}

func (c *Client) Heartbeat(ctx context.Context) error {
	resp, err := c.svc.Heartbeat(ctx, &servicev1.HeartbeatRequest{})
	if err != nil {
		slog.Debug("heartbeat failed", "addr", c.cc.Target(), "err", err)
		return err
	}
	slog.Debug("heartbeat ok", "addr", c.cc.Target(), "available_bytes", resp.GetAvailableBytes())
	return nil
}

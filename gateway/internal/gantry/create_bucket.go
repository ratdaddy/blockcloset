package gantry

import (
	"context"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (c *Client) CreateBucket(ctx context.Context, name string) (string, error) {
	resp, err := c.buckets.CreateBucket(ctx, &servicev1.CreateBucketRequest{Name: name})
	if err != nil {
		return "", err
	}
	return resp.Bucket.Name, nil
}

package gantry

import (
	"context"
	"time"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (c *Client) ListBuckets(ctx context.Context) ([]Bucket, error) {
	resp, err := c.buckets.ListBuckets(ctx, &servicev1.ListBucketsRequest{})
	if err != nil {
		return nil, err
	}

	buckets := make([]Bucket, 0, len(resp.GetBuckets()))
	for _, b := range resp.GetBuckets() {
		var createdAt time.Time
		if ts := b.GetCreatedAtRfc3339(); ts != "" {
			createdAt, err = time.Parse(time.RFC3339, ts)
			if err != nil {
				return nil, err
			}
		}

		buckets = append(buckets, Bucket{
			Name:      b.GetName(),
			CreatedAt: createdAt,
		})
	}

	return buckets, nil
}

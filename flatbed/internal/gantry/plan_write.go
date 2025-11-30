package gantry

import (
	"context"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
	writeplanv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/write_plan/v1"
)

func (c *Client) PlanWrite(ctx context.Context, bucket, key string, size int64) (*writeplanv1.WritePlan, error) {
	resp, err := c.buckets.PlanWrite(ctx, &servicev1.PlanWriteRequest{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetWritePlan(), nil
}

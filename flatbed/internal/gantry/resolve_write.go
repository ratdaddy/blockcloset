package gantry

import (
	"context"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (c *Client) ResolveWrite(ctx context.Context, bucket, key string, size int64) (objectID, cradleAddress string, err error) {
	resp, err := c.buckets.ResolveWrite(ctx, &servicev1.ResolveWriteRequest{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	})
	if err != nil {
		return "", "", err
	}
	return resp.GetObjectId(), resp.GetCradleAddress(), nil
}

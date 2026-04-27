package gantry

import (
	"context"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (c *Client) CommitObject(ctx context.Context, objectID string, size int64, lastModifiedMs int64) error {
	_, err := c.svc.CommitObject(ctx, &servicev1.CommitObjectRequest{
		ObjectId:       objectID,
		Size:           size,
		LastModifiedMs: lastModifiedMs,
	})
	return err
}

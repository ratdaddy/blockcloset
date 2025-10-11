package httpapi

import (
	"context"

	"github.com/ratdaddy/blockcloset/gateway/internal/gantry"
)

type GantryClient interface {
	CreateBucket(ctx context.Context, name string) (string, error)
	ListBuckets(ctx context.Context) ([]gantry.Bucket, error)
}

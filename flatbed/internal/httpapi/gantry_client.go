package httpapi

import (
	"context"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
)

type GantryClient interface {
	CreateBucket(ctx context.Context, name string) (string, error)
	ListBuckets(ctx context.Context) ([]gantry.Bucket, error)
	ResolveWrite(ctx context.Context, bucket, key string) error
}

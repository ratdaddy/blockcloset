package handlers

import (
	"context"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
	"github.com/ratdaddy/blockcloset/pkg/validation"
)

// GantryClient defines the operations needed from the Gantry service.
type GantryClient interface {
	CreateBucket(ctx context.Context, name string) (string, error)
	ListBuckets(ctx context.Context) ([]gantry.Bucket, error)
	ResolveWrite(ctx context.Context, bucket, key string) error
}

// Handlers provides HTTP handler implementations for S3-compatible operations.
// URL parameters are extracted using r.PathValue
type Handlers struct {
	Validator validation.BucketNameValidator
	Gantry    GantryClient
}

func NewHandlers(g GantryClient) *Handlers {
	return &Handlers{
		Validator: validation.DefaultBucketNameValidator{},
		Gantry:    g,
	}
}

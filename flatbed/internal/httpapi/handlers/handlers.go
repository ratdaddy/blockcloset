package handlers

import (
	"context"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
	"github.com/ratdaddy/blockcloset/pkg/validation"
	writeplanv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/write_plan/v1"
)

// GantryClient defines the operations needed from the Gantry service.
type GantryClient interface {
	CreateBucket(ctx context.Context, name string) (string, error)
	ListBuckets(ctx context.Context) ([]gantry.Bucket, error)
	PlanWrite(ctx context.Context, bucket, key string, size int64) (*writeplanv1.WritePlan, error)
}

// Handlers provides HTTP handler implementations for S3-compatible operations.
// URL parameters are extracted using r.PathValue
type Handlers struct {
	BucketValidator validation.BucketNameValidator
	KeyValidator    validation.KeyValidator
	Gantry          GantryClient
}

func NewHandlers(g GantryClient) *Handlers {
	return &Handlers{
		BucketValidator: validation.DefaultBucketNameValidator{},
		KeyValidator:    validation.DefaultKeyValidator{},
		Gantry:          g,
	}
}

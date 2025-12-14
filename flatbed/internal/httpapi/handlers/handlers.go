package handlers

import (
	"context"
	"io"

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

// CradleClient defines the operations needed from the Cradle service.
type CradleClient interface {
	WriteObject(ctx context.Context, address, objectID, bucket string, size int64, body io.Reader) (int64, int64, error)
}

// Handlers provides HTTP handler implementations for S3-compatible operations.
// URL parameters are extracted using r.PathValue
type Handlers struct {
	BucketValidator validation.BucketNameValidator
	KeyValidator    validation.KeyValidator
	Gantry          GantryClient
	Cradle          CradleClient
}

func NewHandlers(g GantryClient, c CradleClient) *Handlers {
	return &Handlers{
		BucketValidator: validation.DefaultBucketNameValidator{},
		KeyValidator:    validation.DefaultKeyValidator{},
		Gantry:          g,
		Cradle:          c,
	}
}

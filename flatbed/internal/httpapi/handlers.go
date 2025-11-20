package httpapi

import "github.com/ratdaddy/blockcloset/pkg/storage/bucket"

// Handlers provides HTTP handler implementations for S3-compatible operations.
// URL parameters are extracted using r.PathValue
type Handlers struct {
	Validator bucket.BucketNameValidator
	Gantry    GantryClient
}

func NewHandlers(g GantryClient) *Handlers {
	return &Handlers{
		Validator: bucket.DefaultBucketNameValidator{},
		Gantry:    g,
	}
}

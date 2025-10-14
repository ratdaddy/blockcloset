package httpapi

import "github.com/ratdaddy/blockcloset/pkg/storage/bucket"

type Handlers struct {
	Validator bucket.BucketNameValidator
	Gantry    GantryClient
}

func NewHandlers(g GantryClient) *Handlers {
	// panic if no gantry client provided
	return &Handlers{
		Validator: bucket.DefaultBucketNameValidator{},
		Gantry:    g,
	}
}

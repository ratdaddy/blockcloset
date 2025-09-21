package httpapi

import "context"

type GantryClient interface {
	CreateBucket(ctx context.Context, name string) (string, error)
}

type Handlers struct {
	Validator BucketNameValidator
	Gantry    GantryClient
}

func NewHandlers(g GantryClient) *Handlers {
	// panic if no gantry client provided
	return &Handlers{
		Validator: DefaultBucketNameValidator{},
		Gantry:    g,
	}
}

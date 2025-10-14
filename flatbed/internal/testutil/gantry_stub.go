package testutil

import (
	"context"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
)

type GantryStub struct {
	CreateFn    func(context.Context, string) (string, error)
	ListFn      func(context.Context) ([]gantry.Bucket, error)
	CreateCalls []string
	ListCalls   int
}

func NewGantryStub() *GantryStub {
	return &GantryStub{}
}

func (g *GantryStub) CreateBucket(ctx context.Context, name string) (string, error) {
	g.CreateCalls = append(g.CreateCalls, name)
	if g.CreateFn != nil {
		return g.CreateFn(ctx, name)
	}
	return "", nil
}

func (g *GantryStub) ListBuckets(ctx context.Context) ([]gantry.Bucket, error) {
	g.ListCalls++
	if g.ListFn != nil {
		return g.ListFn(ctx)
	}
	return nil, nil
}

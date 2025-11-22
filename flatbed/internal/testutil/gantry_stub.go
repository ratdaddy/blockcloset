package testutil

import (
	"context"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
)

type ResolveWriteCall struct {
	Bucket string
	Key    string
	Size   int64
}

type GantryStub struct {
	CreateFn          func(context.Context, string) (string, error)
	ListFn            func(context.Context) ([]gantry.Bucket, error)
	ResolveWriteFn    func(context.Context, string, string, int64) (string, string, error)
	CreateCalls       []string
	ListCalls         int
	ResolveWriteCalls []ResolveWriteCall
}

func NewGantryStub() *GantryStub {
	return &GantryStub{}
}

func (g *GantryStub) CreateCount() int {
	return len(g.CreateCalls)
}

func (g *GantryStub) ListCount() int {
	return g.ListCalls
}

func (g *GantryStub) ResolveWriteCount() int {
	return len(g.ResolveWriteCalls)
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

func (g *GantryStub) ResolveWrite(ctx context.Context, bucket, key string, size int64) (string, string, error) {
	g.ResolveWriteCalls = append(g.ResolveWriteCalls, ResolveWriteCall{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	})
	if g.ResolveWriteFn != nil {
		return g.ResolveWriteFn(ctx, bucket, key, size)
	}
	return "stub-object-id", "localhost:9002", nil
}

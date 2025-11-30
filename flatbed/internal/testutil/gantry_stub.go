package testutil

import (
	"context"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
	writeplanv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/write_plan/v1"
)

type PlanWriteCall struct {
	Bucket string
	Key    string
	Size   int64
}

type GantryStub struct {
	CreateFn       func(context.Context, string) (string, error)
	ListFn         func(context.Context) ([]gantry.Bucket, error)
	PlanWriteFn    func(context.Context, string, string, int64) (*writeplanv1.WritePlan, error)
	CreateCalls    []string
	ListCalls      int
	PlanWriteCalls []PlanWriteCall
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

func (g *GantryStub) PlanWriteCount() int {
	return len(g.PlanWriteCalls)
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

func (g *GantryStub) PlanWrite(ctx context.Context, bucket, key string, size int64) (*writeplanv1.WritePlan, error) {
	g.PlanWriteCalls = append(g.PlanWriteCalls, PlanWriteCall{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	})
	if g.PlanWriteFn != nil {
		return g.PlanWriteFn(ctx, bucket, key, size)
	}
	return &writeplanv1.WritePlan{
		ObjectId:      "stub-object-id",
		CradleAddress: "localhost:9002",
	}, nil
}

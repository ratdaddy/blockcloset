package testutil

import (
	"context"
	"io"
)

type WriteObjectCall struct {
	Address   string
	ObjectID  string
	Bucket    string
	Size      int64
	BodyBytes []byte
}

type CradleStub struct {
	WriteObjectFn    func(context.Context, string, string, string, int64, io.Reader) (int64, int64, error)
	WriteObjectCalls []WriteObjectCall
}

func NewCradleStub() *CradleStub {
	return &CradleStub{}
}

func (c *CradleStub) WriteObjectCount() int {
	return len(c.WriteObjectCalls)
}

func (c *CradleStub) WriteObject(ctx context.Context, address, objectID, bucket string, size int64, body io.Reader) (int64, int64, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return 0, 0, err
	}

	c.WriteObjectCalls = append(c.WriteObjectCalls, WriteObjectCall{
		Address:   address,
		ObjectID:  objectID,
		Bucket:    bucket,
		Size:      size,
		BodyBytes: bodyBytes,
	})

	if c.WriteObjectFn != nil {
		// Re-create reader with the bytes we just read
		return c.WriteObjectFn(ctx, address, objectID, bucket, size, io.NopCloser(io.Reader(nil)))
	}

	// Default: return successful write
	return size, 1234567890, nil
}

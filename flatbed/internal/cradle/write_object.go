package cradle

import (
	"context"
	"fmt"
	"io"

	"github.com/ratdaddy/blockcloset/flatbed/internal/config"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func (c *Client) WriteObject(ctx context.Context, address, objectID, bucket string, size int64, body io.Reader) (int64, int64, error) {
	conn, err := c.pool.GetConn(ctx, address)
	if err != nil {
		return 0, 0, err
	}

	serviceClient := servicev1.NewCradleServiceClient(conn)
	stream, err := serviceClient.WriteObject(ctx)
	if err != nil {
		return 0, 0, err
	}

	// Send metadata
	err = stream.Send(&servicev1.WriteObjectRequest{
		Payload: &servicev1.WriteObjectRequest_Metadata{
			Metadata: &servicev1.WriteObjectMetadata{
				ObjectId: objectID,
				Bucket:   bucket,
				Size:     size,
			},
		},
	})
	if err != nil {
		return 0, 0, err
	}

	// Stream chunks
	buf := make([]byte, config.PutObjectChunkSize)
	var totalBytesRead int64
	for {
		n, err := body.Read(buf)
		if n > 0 {
			totalBytesRead += int64(n)
			if err = stream.Send(&servicev1.WriteObjectRequest{
				Payload: &servicev1.WriteObjectRequest_Chunk{
					Chunk: buf[:n],
				},
			}); err != nil {
				return 0, 0, err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, err
		}
	}

	// Validate size matches what was declared
	if totalBytesRead != size {
		return 0, 0, fmt.Errorf("size mismatch: read %d bytes, expected %d", totalBytesRead, size)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return 0, 0, err
	}

	return resp.GetBytesWritten(), resp.GetCommittedAtMs(), nil
}

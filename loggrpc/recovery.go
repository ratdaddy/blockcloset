package loggrpc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoverToStatus(ctx context.Context, p any) (err error) {
	var (
		st  *status.Status
		msg string
		typ = fmt.Sprintf("%T", p)
	)

	switch v := p.(type) {
	case *status.Status:
		st = v
		msg = v.Message()
	case error:
		if s, ok := status.FromError(v); ok {
			st = s
			msg = s.Message()
		} else {
			msg = v.Error()
		}
	default:
		msg = fmt.Sprint(v)
	}

	frames := CaptureStack(3, 10)

	attrs := []slog.Attr{
		slog.String("event.name", "rpc_panic_recovered"),
		slog.Group("exception",
			slog.String("type", typ),
			slog.String("message", msg),
			slog.Any("stacktrace", frames),
		),
	}

	SetAttrs(ctx, attrs...)

	if st != nil {
		return st.Err()
	} else {
		return status.Error(codes.Internal, "internal server error")
	}
}

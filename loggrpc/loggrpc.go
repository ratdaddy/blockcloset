package loggrpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Also need a StreamServerInterceptor for streaming RPCs
func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		start := time.Now()

		resp, err := handler(ctx, req)

		dur := time.Since(start)

		service, method := splitFullMethod(info.FullMethod)
		code := status.Code(err)

		attrs := []slog.Attr{
			slog.String("grpc.full_method", info.FullMethod),
			slog.String("grpc.service", service),
			slog.String("grpc.method", method),
			slog.String("rpc.system", "grpc"),
			slog.String("grpc.code", code.String()),
			// mimics httplog but maybe should be network.protocol.name (HTTP) & network.protocol.version (2)
			slog.String("network.protocol.version", "HTTP/2"),
			slog.Float64("rpc.server.duration", float64(dur)/float64(time.Second)),
		}

		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			// Doesn't work if behind a proxy
			attrs = append(attrs, slog.String("client.address", p.Addr.String()))

			scheme := "grpc"
			if p.AuthInfo != nil {
				scheme = "grpcs"
			}
			attrs = append(attrs, slog.String("url.scheme", scheme))
		}

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if h := first(md.Get(":authority")); h != "" {
				attrs = append(attrs, slog.String("server.address", h))
			}

			if v := first(md.Get("user-agent")); v != "" {
				attrs = append(attrs, slog.String("user_agent.original", v))
			}
		}

		if pm, ok := req.(proto.Message); ok {
			attrs = append(attrs, slog.Int("grpc.request.size", proto.Size(pm)))
		}

		if pm, ok := resp.(proto.Message); ok {
			attrs = append(attrs, slog.Int("grpc.response.size", proto.Size(pm)))
		}

		level := codeToLevel(code)

		msg := fmt.Sprintf("Unary %s => gRPC %s (%s)", info.FullMethod, code.String(), humanDuration(dur))

		logger.LogAttrs(ctx, level, msg, attrs...)
		return resp, err
	}
}

func splitFullMethod(full string) (service, method string) {
	if len(full) == 0 {
		return "", ""
	}
	if full[0] == '/' {
		full = full[1:]
	}
	for i := 0; i < len(full); i++ {
		if full[i] == '/' {
			return full[:i], full[i+1:]
		}
	}
	return full, ""
}

func humanDuration(d time.Duration) string {
	const (
		us = time.Microsecond
		ms = time.Millisecond
		s  = time.Second
	)
	switch {
	case d >= s:
		return fmt.Sprintf("%.3fs", float64(d)/float64(s))
	case d >= ms:
		return fmt.Sprintf("%.3fms", float64(d)/float64(ms))
	case d >= us:
		return fmt.Sprintf("%.3fµs", float64(d)/float64(us))
	default:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
}

func first(ss []string) string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}

func codeToLevel(c codes.Code) slog.Level {
	switch c {
	case codes.OK, codes.Canceled:
		return slog.LevelInfo
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.ResourceExhausted,
		codes.FailedPrecondition,
		codes.OutOfRange:
		return slog.LevelWarn
	case codes.Unknown,
		codes.DeadlineExceeded,
		codes.Unimplemented,
		codes.Internal,
		codes.Unavailable,
		codes.DataLoss:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

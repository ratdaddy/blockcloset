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
func UnaryServerInterceptor(logger *slog.Logger, o *Options) grpc.UnaryServerInterceptor {
	if o == nil {
		o = &Options{Schema: SchemaOTEL}
	}

	s := o.Schema
	if s == nil {
		s = SchemaOTEL
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		start := time.Now()

		ctx = context.WithValue(ctx, ctxKeyLogAttrs{}, &[]slog.Attr{})

		resp, err := handler(ctx, req)

		dur := time.Since(start)

		service, method := splitFullMethod(info.FullMethod)
		code := status.Code(err)

		attrs := []slog.Attr{}

		attrs = appendAttrs(attrs,
			slog.String(s.FullMethod, info.FullMethod),
			slog.String(s.Service, service),
			slog.String(s.Method, method),
			slog.String(s.System, "grpc"),
			slog.Int(s.Code, int(code)),
			slog.String(s.ProtocolName, "HTTP"),
			slog.String(s.ProtocolVersion, "2"),
			slog.Float64(s.Duration, float64(dur)/float64(time.Second)),
		)

		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			// Doesn't work if behind a proxy
			attrs = appendAttrs(attrs, slog.String(s.RemoteIP, p.Addr.String()))

			scheme := "grpc"
			if p.AuthInfo != nil {
				scheme = "grpcs"
			}
			attrs = appendAttrs(attrs, slog.String(s.Scheme, scheme))
		}

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if h := first(md.Get(":authority")); h != "" {
				attrs = appendAttrs(attrs, slog.String(s.Host, h))
			}

			if v := first(md.Get("user-agent")); v != "" {
				attrs = appendAttrs(attrs, slog.String(s.UserAgent, v))
			}
		}

		if pm, ok := req.(proto.Message); ok {
			attrs = appendAttrs(attrs, slog.Int(s.RequestBytes, proto.Size(pm)))
		}

		if pm, ok := resp.(proto.Message); ok {
			attrs = appendAttrs(attrs, slog.Int(s.ResponseBytes, proto.Size(pm)))
		}

		attrs = appendAttrs(attrs, getAttrs(ctx)...)

		level := codeToLevel(code)

		msg := fmt.Sprintf("Unary %s => gRPC %v (%v)", info.FullMethod, code, dur)

		logger.LogAttrs(ctx, level, msg, attrs...)
		return resp, err
	}
}

func appendAttrs(attrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	for _, attr := range newAttrs {
		if attr.Key != "" {
			attrs = append(attrs, attr)
		}
	}
	return attrs
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

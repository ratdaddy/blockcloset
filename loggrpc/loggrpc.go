// Package loggrpc provides gRPC server interceptors for structured logging.
//
// It supports both unary and streaming RPCs with OpenTelemetry-compatible
// semantic conventions. The interceptors automatically extract request metadata
// (service, method, peer info, headers) and allow handlers to add custom
// attributes using SetAttrs.
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

// StreamServerInterceptor returns a gRPC server interceptor for streaming RPCs
// that logs request metadata and completion status using structured logging.
//
// The interceptor logs once at stream completion (not per-message) with the
// following attributes:
//   - rpc.full_method: The full gRPC method path (e.g., /package.Service/Method)
//   - rpc.service: The service name
//   - rpc.method: The method name
//   - rpc.system: Always "grpc"
//   - rpc.grpc.status_code: The gRPC status code (0 for OK)
//   - network.protocol.name: Always "HTTP"
//   - network.protocol.version: Always "2"
//   - server.duration: Duration in seconds
//   - network.peer.address: Client IP address
//   - url.scheme: "grpc" or "grpcs" (if using TLS)
//   - server.address: Authority header from request
//   - user_agent.original: User-Agent header
//   - rpc.request.header.x-request-id: Request ID if present
//
// Handlers can add custom attributes during stream processing using SetAttrs.
// The context passed to the handler has been prepared to collect these attributes
// via a wrapped ServerStream that preserves context values.
//
// Example usage:
//
//	func (s *Service) MyStream(stream pb.Service_MyStreamServer) error {
//	    ctx := stream.Context()
//	    loggrpc.SetAttrs(ctx, slog.String("user_id", userID))
//	    // ... handle streaming ...
//	}
//
// If options is nil, defaults to OpenTelemetry schema. The interceptor logs
// all RPCs including reflection calls.
func StreamServerInterceptor(logger *slog.Logger, o *Options) grpc.StreamServerInterceptor {
	if o == nil {
		o = &Options{Schema: SchemaOTEL}
	}

	s := o.Schema
	if s == nil {
		s = SchemaOTEL
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		ctx := ss.Context()
		ctx = context.WithValue(ctx, ctxKeyLogAttrs{}, &[]slog.Attr{})

		err := handler(srv, &loggingServerStream{ServerStream: ss, ctx: ctx})

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

			if v := first(md.Get("x-request-id")); v != "" && s.RequestID != "" {
				attrs = appendAttrs(attrs, slog.String(s.RequestID, v))
			}
		}

		attrs = appendAttrs(attrs, getAttrs(ctx)...)

		msg := fmt.Sprintf("Stream %s => gRPC %v (%v)", info.FullMethod, code, dur)
		logger.LogAttrs(ctx, codeToLevel(code), msg, attrs...)

		return err
	}
}

// UnaryServerInterceptor returns a gRPC server interceptor for unary RPCs
// that logs request metadata and completion status using structured logging.
//
// The interceptor logs once after each RPC completes with the following attributes:
//   - rpc.full_method: The full gRPC method path (e.g., /package.Service/Method)
//   - rpc.service: The service name
//   - rpc.method: The method name
//   - rpc.system: Always "grpc"
//   - rpc.grpc.status_code: The gRPC status code (0 for OK)
//   - network.protocol.name: Always "HTTP"
//   - network.protocol.version: Always "2"
//   - server.duration: Duration in seconds
//   - network.peer.address: Client IP address
//   - url.scheme: "grpc" or "grpcs" (if using TLS)
//   - server.address: Authority header from request
//   - user_agent.original: User-Agent header
//   - rpc.request.header.x-request-id: Request ID if present
//   - rpc.request.size: Request message size in bytes (if proto.Message)
//   - rpc.response.size: Response message size in bytes (if proto.Message)
//
// Handlers can add custom attributes by calling SetAttrs with the request context.
//
// Example usage:
//
//	func (s *Service) MyMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
//	    loggrpc.SetAttrs(ctx, slog.String("user_id", req.UserId))
//	    // ... handle request ...
//	}
//
// If options is nil, defaults to OpenTelemetry schema. The interceptor logs
// all RPCs including reflection calls.
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

			if v := first(md.Get("x-request-id")); v != "" && s.RequestID != "" {
				attrs = appendAttrs(attrs, slog.String(s.RequestID, v))
			}
		}

		if pm, ok := req.(proto.Message); ok {
			attrs = appendAttrs(attrs, slog.Int(s.RequestBytes, proto.Size(pm)))
		}

		if pm, ok := resp.(proto.Message); ok {
			attrs = appendAttrs(attrs, slog.Int(s.ResponseBytes, proto.Size(pm)))
		}

		attrs = appendAttrs(attrs, getAttrs(ctx)...)

		msg := fmt.Sprintf("Unary %s => gRPC %v (%v)", info.FullMethod, code, dur)

		logger.LogAttrs(ctx, codeToLevel(code), msg, attrs...)
		return resp, err
	}
}

type loggingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *loggingServerStream) Context() context.Context {
	return s.ctx
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

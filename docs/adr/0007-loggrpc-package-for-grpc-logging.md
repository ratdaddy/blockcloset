# 0007: `loggrpc` Package for gRPC Logging

- **Status**: Adopted
- **Date**: 2025-09-01
- **Author**: Brian VanLoo

## Context

The BlockCloset project uses gRPC for inter-service communication. While Go’s `httplog` provides structured, OTEL-friendly logging for HTTP services, there is no equivalent for gRPC. The official gRPC Go package provides `grpclog`, but it is limited in scope and does not offer the same features as `httplog`, such as:

- Structured, contextual logging (slog/json/tint).
- Standard OTEL attribute schemas.
- Request/response logging with latency and error handling.
- Middleware-style integration.

For observability, consistency, and developer experience, we want a solution that makes logging gRPC calls as easy and standardized as `httplog` does for HTTP.

## Decision

We will create a new Go package, tentatively named **`loggrpc`**, under the `blockcloset` repository.
- Import path: `github.com/ratdaddy/blockcloset/loggrpc` (for prototyping).
- Long-term home: `github.com/ratdaddy/loggrpc` (to be extracted and open sourced once stable).

The package will provide:
- gRPC server and client interceptors for logging calls.
- Logs in OTEL schema format, compatible with `httplog`.
- Configurable verbosity levels and log formats.
- Integration with Go’s `slog`.

## Consequences

- Ensures consistent developer experience across HTTP and gRPC services.
- Improves observability and diagnostics for BlockCloset gRPC services.
- Creates an opportunity for community adoption once open sourced.
- Introduces additional maintenance burden for a new library.
- Some overlap with existing `grpc-ecosystem` logging middlewares.

## Alternatives Considered

- **Use gRPC’s built-in `grpclog`**
  - Rejected because it lacks structured logging and OTEL integration.

- **Use `grpc-ecosystem/go-grpc-middleware/logging/*`**
  - Rejected because these focus on specific loggers (zap, logrus) and don’t match the httplog-style schema/approach.

- **Skip structured gRPC logging**
  - Rejected because observability is a core BlockCloset goal.

## References

- [httplog on pkg.go.dev](https://pkg.go.dev/github.com/go-chi/httplog)
- [grpc-ecosystem logging middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)

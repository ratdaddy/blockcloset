# 0006: Structured Logging with slog and go-chi/httplog

- **Status**: Adopted
- **Date**: 2025-08-16
- **Author**: Brian VanLoo

## Context

We want structured, production-grade logging for both application logs and HTTP request logs.
We also want logs that align with OpenTelemetry (OTEL) conventions to support interoperability with observability pipelines.

## Decision

- Use a **single default `slog.Logger`**, initialized at program startup, as the root logger for both application logs and request logs.
- For request logging:
  - Use `go-chi/httplog` v3 middleware with the OTEL schema.
  - In development/test environments, apply `.Concise(true)` for shorter log output; in other environments, use the full OTEL schema.
  - Enable `RecoverPanics: true` so request panics are logged and converted into 500 responses.
- Application and request logs share the same logger instance, ensuring consistent formatting and output configuration.
- Request-level log enrichment uses `httplog.SetAttrs(r.Context(), ...)`, which requires passing the request context into helpers (e.g. `respond.Error`).

## Consequences

- Logging is consistent across application and HTTP request boundaries, with environment-specific output suitable for local debugging and production ingestion.
- Using OTEL schema positions the system for future integration with observability tools without schema migration.
- Keeping a single root logger avoids duplication of handler setup and keeps configuration centralized.
- `respond.Error` and similar helpers must receive the `http.Request` so they can use the context for structured log enrichment.
- Default Chi 404/405 behaviors can be overridden to ensure logs and responses follow the same structured error patterns.

## Alternatives Considered

- **Separate loggers for app and request logs**
  Not needed; a single `slog.Logger` can be reused by both direct logging and by `httplog`.
- **Field filtering with ReplaceAttr**
  Not used; instead, rely on `SchemaOTEL` or `SchemaOTEL.Concise` depending on environment.
- **Additional panic recovery middleware**
  Not adopted; rely on `httplog`’s built-in `RecoverPanics`.
- **Other logging libraries (`zap`, `zerolog`, `logrus`)**
  These are widely used in Go projects, but all introduce an external dependency and diverge from the standard library’s direction. Since `log/slog` is now part of the standard library, provides structured logging, integrates directly with `httplog`, and supports interchangeable handlers (`tint`, JSON), it was chosen over third-party options.


## References

- [Go `log/slog`](https://pkg.go.dev/log/slog)
- [go-chi/httplog](https://pkg.go.dev/github.com/go-chi/httplog/v3)
- [lmittmann/tint](https://github.com/lmittmann/tint)
- [OpenTelemetry Logs Data Model](https://opentelemetry.io/docs/specs/otel/logs/data-model/)

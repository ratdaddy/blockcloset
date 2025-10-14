# 0010: Adoption of gRPC for Control-Plane API

- **Status**: Adopted
- **Date**: 2025-10-11
- **Author**: Brian VanLoo

## Context

The Gantry service provides BlockCloset’s control plane: creating buckets, managing placement metadata, and coordinating future replication and lifecycle workflows. It should expose an internal API that is:

- Strongly typed so that flatbed, admin UI, background jobs, and tooling share the same schema.
- Evolvable without breaking callers, supporting backwards-compatible field additions, explicit versioning, and robust metadata handling.
- Observable through OpenTelemetry, tying into the emerging `loggrpc` instrumentation (ADR 0007) and existing tracing/metrics pipelines living in `otel/`.
- Efficient for low-latency operations that run on the same trusted network as other BlockCloset services.

## Decision

Gantry’s control-plane API will be delivered exclusively via gRPC. Protobuf definitions under `proto/gantry` is the canonical contract, and all internal and future third-party clients will use generated stubs. Gantry will continue to expose a gRPC server (`gantry/cmd/gantry`) that composes interceptors for structured logging, telemetry, and resilience, mirroring the patterns set by ADR 0007.

## Consequences

- Enables code-generated clients and servers across services, ensuring consistency and minimizing hand-written serialization.
- Aligns control-plane observability with `loggrpc` interceptors and OTEL exporters, simplifying cross-service tracing.
- Provides builtin gRPC features (metadata, deadlines, retries) that flatbed can leverage when translating S3-style HTTP requests.
- Introduces a translation boundary inside flatbed: HTTP requests must be mapped to gRPC calls, and failure semantics must be mirrored (e.g., HTTP 500 ↔ `codes.Internal` per project error policy).
- Requires gRPC-aware infrastructure (load balancers, health checks) for multi-instance deployments; tooling like gRPC reflection remains optional but recommended for local introspection.

## Alternatives Considered

- **HTTP+JSON Control Plane**
  - Rejected because it duplicates the HTTP surface already exposed externally (ADR 0002) without offering the type safety and contract enforcement we need internally. Maintaining synchronized JSON payloads and validators would add drift and runtime parsing overhead.
- **Embedding Gantry Logic via a Shared Library**
  - Rejected because it would tightly couple callers to Gantry’s storage layer (`gantry/internal/store`), skip authorization boundaries, and complicate scaling Gantry independently.
- **GraphQL or gRPC-Gateway Hybrid**
  - Rejected for now; while a hybrid would benefit browser-based tooling, it would bring additional translation layers and schema management costs. We can layer gRPC Gateway later if browser clients require it, without changing the core decision.

## References

- [ADR 0002: External API – S3-Compatible HTTP Subset](0002-external-api-s3-compatible-http-subset.md)
- [ADR 0007: `loggrpc` Package for gRPC Logging](0007-loggrpc-package-for-grpc-logging.md)

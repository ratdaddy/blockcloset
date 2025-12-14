# Architecture Decision Records Index

This index lists all ADRs for the project and their current implementation state.
It represents the **set of governing architectural decisions** currently in effect.
Superseded and deprecated ADRs remain in the repository for historical purposes but are not listed here.

---

## Active ADRs

Only ADRs in the **Adopted** state appear here.

| ADR # | Title | Implementation | Summary |
|-------|-------|----------------|---------|
| [0001](0001-adr-process-and-template-usage.md) | ADR Process and Template Usage | Complete | Defines how ADRs are created, maintained, and used by humans and AI agents. |
| [0002](0002-external-api-s3-compatible-http-subset.md) | External API: S3-Compatible HTTP Subset | Partial | Specifies the design and implementation of an S3-compatible HTTP API for external access. |
| [0003](0003-unit-and-routing-testing-for-go-http-services.md) | Unit and Routing Testing for Go HTTP Services | Complete | Outlines the testing strategy for Go HTTP services, including unit tests and routing tests. |
| [0005](0005-standardizing-environment-variable-for-application-runtime-mode.md) | Standardizing Environment Variable for Application Runtime Mode | Complete | Establishes a standard environment variable to define the application runtime mode (e.g., development, production). |
| [0006](0006-structured-logging-with-slog-and-go-chi-httplog.md) | Structured Logging with slog and go-chi/httplog | Complete | Describes the adoption of structured logging using the slog package and integration with go-chi/httplog. |
| [0007](0007-loggrpc-package-for-grpc-logging.md) | loggrpc Package for gRPC Logging | Complete | Introduces the loggrpc package to facilitate structured logging in gRPC services. |
| [0008](0008-sqlite-for-local-metadata-store.md) | SQLite for Local Metadata Store | Complete | Chooses SQLite (via modernc driver) for per-node durable metadata storage. |
| [0009](0009-schema-migrations-with-golang-migrate.md) | Schema Migrations with golang-migrate | Complete | Establishes golang-migrate as the standard tool for SQLite schema upgrades and rollbacks. |
| [0010](0010-adoption-of-grpc-for-control-plane-api.md) | Adoption of gRPC for Control-Plane API | Complete | Documents the decision to expose Gantry’s control-plane API via gRPC contracts. |
| [0011](0011-gantry-schema-conventions.md) | Gantry Schema Conventions | Complete | Establishes ID, timestamp, and indexing standards for Gantry persistence. |
| [0012](0012-error-testing-only-for-transformations.md) | Error Testing Only for Transformations | Complete | Defines when to test error conditions: only when errors are transformed, not when simply propagated. |
| [0013](0013-adoption-of-enhanced-servemux-for-http-routing.md) | Adoption of Enhanced ServeMux for HTTP Routing | Complete | Uses Go 1.22+ enhanced ServeMux for HTTP routing, superseding the chi router decision. |

---

## Considering

ADRs here are being actively evaluated but no final decision has been made.

| ADR # | Title | Summary |
|-------|-------|---------|
| *None currently* | | |

---

## Future

ADRs here record context, constraints, and possible alternatives for decisions that will be made in the future.

| ADR # | Title | Summary |
|-------|-------|---------|
| *None currently* | | |

---

*Note*:
- Update this index whenever an ADR is created, adopted, superseded, deprecated, or changes implementation state.
- “Implementation” column tracks the extent to which the codebase reflects the decision: **Complete**, **Partial**, or **None**.

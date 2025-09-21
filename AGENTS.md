# Repository Guidelines

- Maintain this document by keeping the short bullet-oriented sections.

## Project Structure & Module Organization
- `gateway/` delivers the public API and future admin UI
    - keep HTTP handlers in `cmd/gateway` and service logic inside `gateway/internal/`.
- `gantry/` runs the control plane
    - entrypoints in `cmd/gantry` and domain code in `gantry/internal/`.
- Shared protobuf contracts live in `proto/` with generated Go under `proto/gen/`
- Production-like runtime sits in `otel/` with a focus on observability via OpenTelemetry.
- gRPC logging utilities in `loggrpc/`
    - This module is intended to grow into an open-sourced library for structured logging in gRPC services with some of the same intents as go-chi/httplog


## Architecture
- Architecture decisions are defined in ADRs in `docs/adr`. Refer to these documents for understanding the design choices made and incorporate them in any suggestions you make
- If there is a topic we are discussing or deciding on suggest that we create a new ADR on that topic.

## Environment & Configuration
- Load shared settings via `direnv`, sourcing `env/.env.development`
    - place machine-specific overrides in `env/.env.local` and keep them untracked.
- Services read `GATEWAY_*` and `GANTRY_*` variables, so adjust both sets when changing local ports, hosts, or TLS-related options.

## Build, Test, and Development Commands
- `make run` (from `gateway/` or `gantry/`) executes the service once with the current config
- `make dev` hot-reloads using `entr`.
- `make test` runs the full Go unit suite
- `make testdev` watches files and reruns tests with timestamps for tight inner loops.
- `make build` produces binaries in `bin/`
- in `proto/`, use `make lint` before editing schemas and `make gen` after edits to refresh generated stubs.

## Coding Style & Naming Conventions
- Format Go with `gofmt`
- keep imports grouped standard/third-party/internal and avoid unused exports.
    - Use a blank line between groups, sorted alphabetically within each group.
- Use lowercase package directories and filenames that match the primary type or feature (e.g., `router.go`, `config.go`).
- Prefer context-aware logging through the shared `log/slog` helpers (`gateway/internal/logger`, `gantry/internal/logger`)
- keep public structs slim with explicit JSON/proto tags when exposed.
- No trailing whitespace in any files, including documentation. All files end with a newline.

## Testing Guidelines
- Co-locate table-driven tests as `<name>_test.go` using the Go `testing` package, mirroring the directory of the code under test.
- Run `go test ./...` (or `go test -cover ./...` for critical changes) in both services before pushing.
- Favor interface fakes over network calls
- reference existing patterns in `gateway/internal/httpapi` and `gantry/internal/grpcsvc` for helpers.

## Commit & Pull Request Guidelines
- Write short, imperative commit subjects (e.g., "Add centralized environment configuration") and reference issues in the body when relevant.
- All Pull Requests should be atomic and focused on a single change or feature.
- Pull requests should be of the form `# Problem`, `# Solution` where the problem section describes the context and problem the change solves.

# Repository Guidelines

## Service Topology

- **Flatbed** — public HTTP API and future admin UI; the only component clients talk to directly.
- **Gantry** — control plane; manages metadata, cluster membership, and replication policy.
- **Cradle** — storage nodes; store and serve raw object data as blocks on local disks.

Request flow: client → flatbed (HTTP) → gantry (gRPC, metadata/coordination) → cradle (gRPC, blob writes/reads).

All three services share the same layout: `cmd/<service>/` for entrypoints, `internal/` for domain logic. Flatbed holds gRPC client wrappers for gantry (`flatbed/internal/gantry`) and cradle (`flatbed/internal/cradle`).

## Project Structure

- `flatbed/` — public API and future admin UI
- `gantry/` — control plane
- `cradle/` — block storage nodes
- `proto/` — shared protobuf contracts; generated Go under `proto/gen/`
- `loggrpc/` — gRPC logging utilities; intended to become an open-sourced structured-logging library (similar in intent to go-chi/httplog)

## Architecture

- Significant decisions are documented as ADRs in `docs/adr/`. Refer to them when making design choices and suggest creating a new ADR when a significant decision arises.

## Environment & Configuration

- Load shared settings via `direnv`, sourcing `env/.env.development`; place machine-specific overrides in `env/.env.local` (untracked).
- Services read `FLATBED_*` and `GANTRY_*` variables; adjust both when changing ports, hosts, or TLS options.

## Build & Development

- `make run` — run the service; `make dev` — hot-reload via `entr`; `make build` — binaries to `bin/`
- `make test` — full unit suite; `make testdev` — watch mode with timestamps
- In `proto/`, run `make lint` before editing schemas and `make gen` after to refresh stubs.
- Do not create `.gocache` directories; use `go clean -cache` to clear the cache.

## API Semantics

- Mirror AWS S3 responses: gantry `codes.Internal` → HTTP 500 `InternalError`; reserve 503 for retryable outages, 504 for timeouts.

## Testing

- Favor interface fakes over network calls.
- Do not ship new production behavior without a failing test in place unless explicitly approved.
- After writing a new test, do not write the production code until told to do so.
- Follow an interleaved TDD loop: one failing test → minimal code to pass → next test. Keep each pair small enough to review at each increment.
- For gRPC client wrapper tests, follow the pattern in `docs/adr/0014-grpc-client-wrapper-testing.md`.

## Coding Guardrails

- Never remove files without explicit approval.
- When editing files that contain user changes, assume those edits are intentional and preserve them unless told otherwise.

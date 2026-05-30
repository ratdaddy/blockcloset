# Repository Guidelines

@~/.claude/process/planning-workflow.md
@.claude/process/development-workflow.md

## Service Topology

- **Flatbed** — public HTTP API and future admin UI; the only component clients talk to directly.
- **Gantry** — control plane; manages metadata, cluster membership, and replication policy.
- **Cradle** — storage nodes; store and serve raw object data as blocks on local disks.

All three services share the same layout: `cmd/<service>/` for entrypoints, `internal/` for domain logic.

Object put/get: client → flatbed (HTTP) → gantry (gRPC, resolve which cradle nodes to use) → flatbed → cradle (gRPC, blob write/read). Flatbed is the hub and holds client wrappers for both (`flatbed/internal/gantry`, `flatbed/internal/cradle`).

Cluster management: gantry calls cradle directly for heartbeats and other cluster concerns, and will hold its own client wrapper for cradle (`gantry/internal/cradle`).

## Project Structure

- `flatbed/` — public API and future admin UI
- `gantry/` — control plane
- `cradle/` — block storage nodes
- `proto/` — shared protobuf contracts; generated Go under `proto/gen/`
- `loggrpc/` — gRPC logging utilities; intended to become an open-sourced structured-logging library

## Documentation

- **Architecture decisions** — `docs/adr/`; consult before making significant design choices;
  suggest a new ADR when a decision point is reached
- **Design references** — `docs/design/`; system structure, data models, interfaces
- **Planning** — `docs/initiatives/`; active initiatives, plans, and ROADMAP

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

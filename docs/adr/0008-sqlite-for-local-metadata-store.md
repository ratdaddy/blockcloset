# 0008: SQLite for Local Metadata Store

- **Status**: Adopted
- **Date**: 2025-09-28
- **Author**: Brian VanLoo

## Context

Each Gantry node in a BlockCloset deployment will need to persist bucket definitions and related metadata so it can operate autonomously while remaining coordinated through Raft replication. Homelab environments favor lightweight dependencies, simple upgrades, and self-contained binaries. The persistence layer must therefore provide durable writes, predictable crash recovery, and minimal operational overhead while aligning with Raft’s single-writer apply loop.

We also need tooling that lets operators inspect data easily using commodity CLI or GUI tools.

## Decision

BlockCloset will store Gantry’s local metadata in [SQLite](https://sqlite.org/), accessed from Go via the [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) driver. SQLite provides:
- A transactional, durable single-file database tuned for embedded use.
- Deterministic behavior compatible with Raft’s single-writer semantics via WAL mode.
- Broad ecosystem support, making inspection and backups straightforward for homelab operators.

The modernc driver keeps builds CGO-free so binaries cross-compile cleanly and avoid external toolchains, while still tracking upstream SQLite features closely.

Implementation:
- `gantry/internal/database` will own connection setup.
- Domain packages will work directly with SQLite-backed structs under `gantry/internal/store` (e.g., `bucket.go`). These types will embed a `*sql.DB` and expose concrete methods (`Create`, `Get`, etc.) without introducing intermediate repository interfaces.
- Database-facing structs (e.g., `bucketRecord`) remain separate from protobuf transports so persistence concerns stay decoupled from gRPC representations.

## Consequences

- Enables durable local state without introducing an external database dependency.
- Keeps deployment lightweight: one data file per node plus optional WAL artifacts.
- Simplifies manual inspection (e.g., via the `sqlite3` CLI or GUI browsers) for troubleshooting upgrades.
- Requires conscious coordination between migrations and Raft apply logic to avoid long-lived locks.
- Adds the modernc driver dependency, which may trail official SQLite releases by a short window and has slightly higher CPU cost than CGO builds.
- Couples Gantry’s data access directly to SQLite implementations; swapping back-ends later would require refactoring instead of interface substitution.

## Alternatives Considered

- **In-memory store**: unsuitable because it loses state on restart and complicates upgrades.
- **BoltDB/Bbolt**: durable, single-writer friendly, but forces manual serialization strategies and lacks the relational querying flexibility we anticipate needing.
- **BadgerDB**: high-performance LSM tree with GC overhead; introduces more operational tuning than necessary for the target scale.
- **Pebble (RocksDB-compatible)**: powerful and battle-tested for large clusters, but heavier to embed and maintain than SQLite for homelab scenarios.
- **External PostgreSQL/MySQL instance**: would offload storage but adds configuration complexity and breaks the “single binary deployment” goal for homelabs.

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [modernc.org/sqlite driver](https://pkg.go.dev/modernc.org/sqlite)

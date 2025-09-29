# 0008: Schema Migrations with golang-migrate

- **Status**: Adopted
- **Date**: 2025-09-28
- **Author**: Brian VanLoo

## Context

Gantry nodes will persist bucket definitions and related metadata locally so they can participate in a Raft cluster while remaining operable in homelab environments. We expect to ship the service with simple installation and maintenance tools that can be upgraded or downgraded without bespoke orchestration. This means every node needs a predictable way to apply schema changes, surface drift, and align its local SQLite database with the version of the running binary.

The migration approach must:
- Work reliably in constrained homelab deployments where ops tooling may be limited.
- Support both upgrades and rollbacks between released versions.
- Integrate cleanly with Go so services can detect schema version mismatches during startup.
- Avoid coupling the project to environment-specific dependencies (e.g., external migration services).

## Decision

BlockCloset will use [golang-migrate/migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4) as the migration engine for SQLite schemas. Migration definitions will be expressed as SQL `up` and `down` files stored in the repository and executed through the migrate CLI or library. Services may run migrations automatically or expose an explicit admin command, but migrate’s version table will remain the single source of truth for schema state.

## Consequences

- Provides a battle-tested mechanism for applying and rolling back migrations, reducing upgrade risk for homelab operators.
- Keeps migrations as plain SQL so changes remain transparent and reviewable.
- Enables services to programmatically check schema versions via the migrate library and warn or refuse to start when drift is detected.
- Introduces a dependency on maintaining paired `up`/`down` scripts; mistakes can make downgrade paths harder.
- Requires coordination so long-running transactions or Raft apply loops do not hold locks when migrations execute.

## Alternatives Considered

- **pressly/goose**: Simpler CLI with optional Go-based migrations, but lacks broad dialect support and version drift tooling that migrate provides.
- **ariga/atlas**: Powerful declarative schema management and diffing, yet adds operational complexity and a new DSL that is heavier than needed for BlockCloset’s smaller footprint.
- **amacneil/dbmate**: CLI-oriented and language-agnostic, but would not integrate as smoothly with Go services for programmatic version checks.
- **rubenv/sql-migrate**: Embeds migrations in Go and YAML, yet has a smaller community and fewer best-practice patterns for downgrade testing.

## References

- [golang-migrate/migrate](https://github.com/golang-migrate/migrate)
- [SQLite Write-Ahead Logging](https://www.sqlite.org/wal.html)

# 0011: Gantry Schema Conventions

- **Status**: Adopted
- **Date**: 2025-10-15
- **Author**: Brian VanLoo

## Context

Gantry’s control-plane metadata lives in SQLite as established in [ADR 0008](0008-sqlite-for-local-metadata-store.md) and is managed through migrations from [ADR 0009](0009-schema-migrations-with-golang-migrate.md). To keep schema migrations predictable and enable deterministic replication across the cluster, we need explicit conventions for identifiers, timestamps, foreign keys, and indexing.

Database triggers or implicit defaults would make Raft state divergence likely because Gantry applies mutations inside the state machine; all derived fields must therefore be computed in Go before writes are replicated.

## Decision

Adopt the following conventions for all Gantry SQLite schemas:

1. **Identifiers**
   - Primary keys are `TEXT` ULIDs generated in Go before persistence.
   - Each foreign key column references the corresponding `id` with `ON DELETE RESTRICT` unless a different cascade is explicitly required.
2. **Audit Columns**
   - Tables include `created_at` and `updated_at` columns defined as `INTEGER NOT NULL`.
   - Timestamp values represent Unix epoch microseconds in UTC.
   - Values are generated and updated in Go code; SQLite defaults, triggers, or `CURRENT_TIMESTAMP` are not used so Raft replication remains deterministic.
3. **Additional Temporal Columns**
   - Timestamp names will end in `_at` unless they're a timestamp that's named for what it represents in the API (e.g. `last_modfied`)
   - Any optional lifecycle timestamps (e.g., `last_modified`, `expires_at`) use the same integer microsecond format and nullable semantics where appropriate.
4. **Indexing Patterns**
   - Unique or partial indexes reflect S3 semantics (for example, ensuring only one committed object per bucket/key).
   - Composite indexes follow access paths used by service methods; avoid speculative indexes until a query warrants them.
5. **Migrations and Testing**
   - Migrations encode these conventions explicitly.

Implementation guidance:

- Normalize all timestamps to UTC, truncate to microsecond precision, and write with `time.Time.UnixMicro()`.
- Use `time.UnixMicro()` when hydrating records and convert to RFC-compliant strings at service boundaries.
- Validate foreign keys on insert/update paths and prefer explicit joins over denormalized text columns when relating entities (e.g., `bucket_id`, `cradle_server_id`).

## Consequences

- Enables indexed ordering, range predicates, and comparisons without parsing text.
- Unifies timestamp handling across schema modules, simplifying future migrations and tooling.
- Keeps precision compatible with S3-style header formatting while avoiding floating-point rounding issues.
- Guarantees IDs and timestamps are deterministic across Raft replicas because they are produced in Go before writes.
- Ensures referential integrity between core tables (buckets, objects, cradle servers) by default.
- Manual inspection via SQLite CLI requires explicit conversion (`datetime(created_at/1e6, 'unixepoch')`), adding minor friction for ad-hoc queries.
- Existing helpers must consistently truncate to microseconds to avoid test flake stemming from nanosecond differences.

## Alternatives Considered

- **Store RFC3339 text strings**
  - Rejected because it forces per-query parsing in Go, undermines index performance, and risks lexicographic ordering bugs.
- **Store Unix milliseconds as integers**
  - Rejected to preserve sub-millisecond precision for future observability or TTL requirements without adding trailing zeros later.
- **Use SQLite `REAL` values for fractional seconds**
  - Rejected due to floating-point rounding behavior and harder equality semantics compared with integral microseconds.
- **Populate timestamps and IDs with SQLite triggers**
  - Rejected because trigger-generated values are non-deterministic from the perspective of the Raft state machine, risking drift between nodes.

## References

- [ADR 0008 – SQLite for Local Metadata Store](0008-sqlite-for-local-metadata-store.md)
- [ADR 0009 – Schema Migrations with golang-migrate](0009-schema-migrations-with-golang-migrate.md)

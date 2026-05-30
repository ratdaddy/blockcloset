# Design: Gantry Background Task System

System design for the background task infrastructure running inside the gantry executable.
See `docs/initiatives/gantry-background-tasks/initiative.md` for the implementation plan.

---

## Blob State Machine

Every blob tracked in gantry's data store has a state. The valid states and transitions are:

```
PENDING → COMMITTED      (flatbed sends commit request)
PENDING → FAILED         (flatbed sends failure notification, or staleness timeout — future)
COMMITTED → SUPERSEDED   (a new COMMITTED blob replaces this one for the same key)
SUPERSEDED → DELETED     (cradle confirms deletion)
FAILED → DELETED         (cradle confirms deletion)
```

**State definitions:**

- **PENDING** — upload is in progress. Invisible to all cleanup and background workers.
- **COMMITTED** — live, current version for its object key. Only one blob per key may be
  COMMITTED at a time.
- **SUPERSEDED** — a newer commit has arrived for this key; this blob is the previous
  version and is eligible for deletion.
- **FAILED** — flatbed reported a detectable upload failure; eligible for deletion.
- **DELETED** — cradle has confirmed deletion. Terminal state.

**Deletion is idempotent on the cradle side.** Gantry will retry delete commands until it
receives confirmation. No intermediate DELETING state is needed at this stage. A DELETING
state will be reconsidered when replication is implemented and gantry needs to track
per-server deletion confirmation across multiple cradle instances.

**Out of scope for this phase:** detection of PENDING blobs orphaned because flatbed crashed
without sending a commit or failure notification. This requires a staleness timeout mechanism
and is a defined future step.

---

## Upload Concurrency and Overwrite Semantics

Two flatbed instances may concurrently upload different versions of the same object key.
The system must remain consistent at all times — a reader asking gantry for a key always
gets a single, fully committed answer.

The concurrency model:

1. Both uploads create PENDING blob records in gantry with unique blob IDs.
2. Whichever flatbed instance sends a commit request first causes gantry to atomically
   transition the key's current COMMITTED blob (if any) to SUPERSEDED and promote the
   new blob to COMMITTED.
3. When the second commit request arrives, its blob is already PENDING against a key that
   now has a different COMMITTED blob. Gantry transitions the second blob directly to
   SUPERSEDED.
4. Both SUPERSEDED blobs are eligible for deletion by the cleanup worker.

This resolution must be protected by a mutex or equivalent in gantry to prevent two
concurrent commit handlers from both believing they are the winning commit.

**Read leases are not in scope.** Blob deletion safety is currently limited to ensuring
no in-progress write operations are targeting a blob before issuing a delete command.
A blob in PENDING state has an in-progress write by definition; PENDING blobs are never
targeted for deletion. When reads are implemented, a lease/reference-tracking system will
be required before deletion commands are safe to issue against COMMITTED or SUPERSEDED blobs.

---

## Background Task Supervisor

All background goroutines run under a common supervisor inside the gantry executable.
The supervisor is responsible for:

- Starting all background goroutines on gantry startup
- Propagating a root `context.Context` with cancellation to all goroutines
- On shutdown signal, cancelling the root context and waiting for all goroutines to
  drain cleanly (`sync.WaitGroup` or `errgroup` at the supervisor level)
- Restarting goroutines that exit unexpectedly (optional but realistic)

Graceful shutdown is a first-class requirement, not an afterthought. Every goroutine must
select on `context.Done()` at appropriate points and exit cleanly when cancelled.

---

## Data Model

The following records need to exist in gantry's data store. Schema details (column types,
indexes) are left to implementation planning.

**storage_servers**
- id, address, status (HEALTHY / DEGRADED / OFFLINE), available_capacity_bytes,
  last_heartbeat_at, consecutive_miss_count

**object_keys**
- id, bucket, key, committed_blob_id (FK to blobs)

**blobs**
- id, object_key_id (FK), storage_server_id (FK), state (PENDING / COMMITTED /
  SUPERSEDED / FAILED / DELETED), created_at, updated_at

**blob_replicas** (Step 4+)
- id, blob_id (FK), storage_server_id (FK), status (PENDING / CONFIRMED), confirmed_at

---

## Key Design Constraints

- **Gantry is the source of truth.** Cradles are dumb stores; all state lives in
  gantry's data store.
- **Cradle deletes are idempotent.** Gantry will retry until confirmed without risk of
  double-delete side effects.
- **PENDING blobs are never touched by background workers.** State machine membership
  in PENDING is the write-in-progress signal.
- **Raft replication of gantry's data store is planned but not in scope.** All state
  management assumes a single gantry instance for now. Data model decisions should not
  preclude eventual Raft replication.
- **Read leases are not in scope.** Blob deletion safety rules will need revisiting when
  GET support is implemented.

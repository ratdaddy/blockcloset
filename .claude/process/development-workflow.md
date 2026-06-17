# Development Workflow

## Reference Documents
- Testing strategy: `docs/adr/0003-unit-and-routing-testing-for-go-http-services.md`
- gRPC client wrapper testing: `docs/adr/0014-grpc-client-wrapper-testing.md`

## The TDD Loop

Each cycle in a plan follows this sequence:

1. **Propose and get approval** — describe the intended approach before writing any code;
   get approval before implementation begins. When the prompt explicitly says to implement,
   that is the approval — do not re-ask before writing code.
2. **Red** — write a failing test asserting exactly the behavior being added; confirm the
   specific expected failure before touching production code
3. **Green** — write the minimum production code to make the test pass; run after each
   change; stop when all assertions pass
4. **Refactor** — clean up without changing behavior; run tests to confirm still passing
5. **Confirm complete** — confirm the cycle is complete before advancing to the next

## TDD Rules

- The Red section describes only what to assert, never what to implement
- The expected failure must be specific enough to distinguish the right failure from the
  wrong one — a compile error and a wrong value are different failures
- Green does the minimum to make the current test pass — nothing more
- Fake/test infrastructure grows in the Green step that needs it, not before
- **Green is iterative:** a compile error is fixed with a stub, then the first assertion
  fails and is fixed, then the next assertion fails and is fixed, and so on. The Checkpoint
  is reached when all assertions pass, not after the first fix
- A cycle whose Red is immediately green is a documentation cycle — note it, skip to Checkpoint

## Scope Rules

Only implement what is explicitly approved for the current cycle:

- Do ONLY that cycle — not related cycles, not the next cycle
- Do not make "while we're here" changes without authorization

If something noticed during implementation looks worth doing, surface it as a suggestion
at a natural pause point. The developer decides whether to fold it in, add it to deferred
hardening, or skip it.

## Verification

```
make test
```

All tests must pass before a cycle is considered complete. Run `make build` first if
there are compile errors to resolve. For proto changes: run `make lint` before editing
schemas, `make gen` after to refresh stubs.

## Deferred Hardening

Deferred hardening is for things consciously excluded from TDD — defensive error wrapping
on DB operations, observability hooks, logging — that still need to ship. Collect these
in the plan's Deferred Hardening section and address them as a final pass before marking
the plan complete.

## Input Validation Placement

Validation tests and guard clauses for a field belong after the cycle that first uses
that field in handler logic — never before. Writing a validation test before the field is
used anywhere produces a guard that protects code that doesn't exist yet. The right home
is either a small hardening step immediately following the cycle that introduces the field,
or folded into that cycle's Refactor step if it's tidy enough.

---

## Plan Document Format (Tier 2)

```markdown
# Plan: [Feature Name]

## Status
In Progress

## Goal
[One sentence.]

## Prerequisites
[What must exist before this plan can start, if anything.]

## Cycles
- **Cycle 1 (complete):** [one-line description]
- **Cycle 2 (complete):** [one-line description]
- **Cycle 3 (in progress):** [one-line description]
- **Cycle 4:** [one-line description]

## Cycle 3 (in progress): [description]

[Expanded cycle detail goes here, written just before work starts.]

## Deferred Hardening
- [item]

---

## Completed Cycles

### Cycle 1: [description]

[Preserved detail from when this cycle was active.]

### Cycle 2: [description]

[Preserved detail from when this cycle was active.]
```

## Cycle Format (Tier 3)

Expand only the cycle currently being worked. Write this just before starting, not upfront.
Place the expanded detail immediately after the Cycles list. When the cycle is complete,
move its detail to the "Completed Cycles" section at the bottom.

```markdown
## Cycle N (in progress): [description]

**Red:** [Which test file, which case name, which assertion to add.]
Expected failure: `[exact failure — compile error / "want error got nil" / assertion message]`
→ Checkpoint: run tests, confirm this exact failure.

**Green:** Run after each change; stop when all assertions pass.
[What to implement. If a compile error is expected first, note the stub.]
→ Checkpoint: run tests, confirm passing.

**Refactor:** [Specific thing to clean up — omit this section if there is nothing.]
→ Checkpoint: run tests, confirm still passing.
```

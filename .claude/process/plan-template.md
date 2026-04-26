# Plan Template

Use this template for feature implementation plans that follow a TDD workflow.

## How to use

**The plan document is an index, not a spec.** Each cycle is described in one line.
Detail is written for the *current cycle only*, just before starting work on it.
Once a cycle is complete, mark it done and move on — do not add retrospective detail.

**TDD rules this template enforces:**
- The Red section describes only what to assert, never what to implement
- The expected failure must be specific enough to distinguish the right failure from the wrong one
- The Green section does the minimum to make the current test pass — nothing more
- Fake/test infrastructure grows in the Green step that needs it, not before
- A cycle whose Red is immediately green is a documentation cycle — note it, skip to Checkpoint

**Green is iterative.** A single cycle may require multiple passes:
a compile failure is fixed with a stub, then the first assertion fails and is fixed,
then the next assertion fails and is fixed, and so on.
The Checkpoint is reached when all assertions pass, not after the first fix.

**Deferred hardening** is for things consciously excluded from TDD — defensive error
wrapping on DB operations, observability hooks, logging — that still need to ship.
Address these as a final pass before marking the plan complete.

**Input validation placement.** Validation tests and guard clauses for a field belong
*after* the cycle that first uses that field in handler logic — never before. Writing
a validation test before the field is used anywhere produces a guard that protects
code that doesn't exist yet. The right home is either a small hardening step immediately
following the cycle that introduces the field, or folded into that cycle's Refactor step
if it's tidy enough to do so.

---

## Plan Document Format

```markdown
# Plan: [Feature Name]

## Status
In Progress

## Cycles
- **Cycle 1 (done):** [one-line description]
- **Cycle 2 (done):** [one-line description]
- **Cycle 3 (in progress):** [one-line description]
- **Cycle 4:** [one-line description]

## Deferred Hardening
- [item]
- [item]
```

---

## Per-Cycle Format

Expand only the cycle currently being worked. Write this just before starting, not upfront.

```markdown
## Cycle N: [description]

**Red:** [Which test file, which case name, which assertion to add.]
Expected failure: `[exact failure — compile error / "want error got nil" / assertion message]`
→ Checkpoint: run tests, confirm this exact failure.

**Green:** Run after each change; stop when all assertions pass.
[What to implement. If a compile error is expected first, note the stub.]
→ Checkpoint: run tests, confirm passing.

**Refactor:** [Specific thing to clean up — omit this section if there is nothing.]
→ Checkpoint: run tests, confirm still passing.
```

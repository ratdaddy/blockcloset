# 0014: gRPC Client Wrapper Testing

- **Status**: Considering
- **Date**: 2026-04-26
- **Author**: Brian VanLoo

## Context

Flatbed contains gRPC client wrappers for two downstream services: Gantry
(`flatbed/internal/gantry`) and Cradle (`flatbed/internal/cradle`). These
wrappers translate Go method calls into gRPC requests, unmarshal responses into
Go types, and propagate a request ID via outgoing metadata through a stream/unary
interceptor.

Without a defined testing convention the gantry package accumulated inconsistent
tests: some flat, some table-driven, some asserting all request fields, some not,
and request ID propagation tested in a separate file from the method behavior it
belongs to. ADR 0012 clarified when to test errors; this ADR covers structure and
coverage for the happy path and cross-cutting concerns.

## Decision

### One test per method, one file per method

Each client method has exactly one test function in its own `_test.go` file
named after the method (e.g., `create_bucket_test.go`). Shared test
infrastructure lives in `test_helpers_test.go`.

Test files for infrastructure concerns unrelated to a specific method (e.g.,
connection pool behavior) live in their own file named after the component under
test (e.g., `pool_test.go`).

### Flat tests, not table-driven

Each method test is a flat single-case test. A table is only introduced when a
second case tests genuinely different behavior — not error propagation (excluded
by ADR 0012) and not defensive cases for data that is correct by system design
(since all downstream services are owned by this project).

### Every test asserts all three concerns

Each method test must verify:

1. **Request fields** — every field the method maps from its parameters to the
   proto request is asserted on the captured call.
2. **Response mapping** — every field the method maps from the proto response to
   its return value is asserted, confirming correct hydration.
3. **Request ID propagation** — the single call uses a request ID context (via
   `requestid.WithRequestID`), and the test asserts `x-request-id` metadata is
   present and correct on the captured call. The absent direction is not tested
   per method; the interceptor is shared infrastructure and the present-direction
   assertion is sufficient to confirm it is wired.

### Fake defaults return meaningful non-zero values

The `captureXxxService` fake must return meaningful, non-zero values for every
response field by default. This ensures the response mapping assertion in each
method test is actually exercised and cannot pass on a zero-value coincidence.

Hooks on the fake are infrastructure for future use; a test may use one when the
default response is insufficient for its assertions (e.g., response hydration
tests that need specific field values).

## Consequences

- Each method's complete observable behavior — what it sends, what it returns,
  that it carries the request ID — is verified in one place.
- The interceptor is covered once per method rather than in a dedicated file,
  eliminating the need to keep a separate propagation table up to date as new
  methods are added.

## Alternatives Considered

- **Separate file per method**
  *Rejected because*: it splits the complete picture of one method across two
  files (the method file and the interceptor table), requiring two places to
  look when adding or changing a method.

- **Table-driven for all methods**
  *Rejected because*: most client methods have exactly one meaningful case after
  applying ADR 0012 and excluding impossible-in-practice data scenarios. A table
  of one is overhead without benefit.

- **Single `client_test.go` containing all method tests**
  *Rejected because*: the file grows unboundedly as methods are added, with no
  natural organizing principle. Per-method files keep each file small and
  self-contained.

- **Dedicated interceptor test retained alongside per-method tests**
  *Rejected because*: it either duplicates coverage or tests nothing unique once
  each method already asserts propagation, and it requires manual maintenance to
  stay in sync with the method list.

## References

- [ADR 0012](0012-error-testing-only-for-transformations.md) — error testing scope
- [ADR 0010](0010-adoption-of-grpc-for-control-plane-api.md) — gRPC adoption

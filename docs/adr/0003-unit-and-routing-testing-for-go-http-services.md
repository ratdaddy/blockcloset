# 0003: Unit & Routing Testing for Go HTTP Services

* **Status**: Adopted
* **Date**: 2025-08-10
* **Author**: Brian VanLoo

## Context

A consistent, fast, idiomatic, and scalable testing approach is needed for all Go HTTP services in this project. Tests should provide quick feedback without introducing heavy frameworks or unnecessary complexity, while maintaining clarity about what is being tested and why.

## Decision

Use a two-tier, in-process testing strategy for Go HTTP services, relying  on the Go standard library for assertions and HTTP testing.

### Testing Tiers

1. **Unit Tests (Handler-Level)**

   * Exercise request handling logic directly via `http.Handler` and `ServeHTTP`.
   * Focus on verifying handler status codes, required headers, and minimal body structure.
   * Avoid assertions about internal implementation details.

2. **Routing Tests (In-Process Integration)**

   * Validate that HTTP method + path combinations are correctly wired to the intended handlers using the service’s public HTTP surface (e.g., mux/handler).
   * Confirm that expected routes work and that invalid methods or paths are handled appropriately.
   * Keep the scope limited to in-process verification without external dependencies.

Integration tests and storage/backend contract tests will be addressed in separate ADRs.

### Structure & Conventions

* **Test Naming**: Use one top-level function per behavior (e.g., `TestCreateBucket`) with table-driven subtests for variants.
* **Helpers**: Create small helper functions for repeated request/response setup; mark with `t.Helper()`.
* **Parallelism**: Call `t.Parallel()` in top-level tests when no shared state exists. Avoid globals and use `t.Setenv` for per-test configuration.
* **Packages**: Prefer `package <pkg>_test` for black-box testing; use `package <pkg>` only when white-box access is required.
* **Dependencies**: Stick to the Go standard library unless a third-party library becomes necessary for clarity or maintainability.

### Rationale

* **Speed & Determinism**: In-process tests avoid network flakiness and provide rapid feedback.
* **Separation of Concerns**: Differentiating handler logic tests from routing tests makes diagnosing failures easier.
* **Parallelization**: Stateless services and isolated test setup allow for high concurrency in CI.

## Consequences

This strategy ensures fast, deterministic tests that clearly identify whether failures are in handler logic or routing configuration. It promotes stateless, isolated components and scales well as features grow, without creating framework lock-in.

## Alternatives Considered

* **Integration-first testing**: Rejected due to slower feedback and greater flakiness.
* **Framework-dependent testing**: Rejected to avoid unnecessary dependencies and tight coupling.

## References

* [Go testing package](https://pkg.go.dev/testing)
* [net/http/httptest documentation](https://pkg.go.dev/net/http/httptest)

# Pre-Commit Checklist — yardmaster

Repo-specific items checked by `/pre-commit-check` after the global steps.

## TDD discipline

No new production behavior shipped without a prior failing test in place, unless
explicitly approved in this session. If production code was written before a test,
flag it.

## Interface fakes

New tests must use interface fakes over real network or database calls. If a test
spins up a real gRPC server or hits SQLite directly to test logic that could be
isolated, flag it.

## gRPC client wrappers

New gRPC client wrappers must follow the pattern in
`docs/adr/0014-grpc-client-wrapper-testing.md` — one test file, flat tests, all
three concerns per method (happy path, transport error, application error).

## Go version consistency

Application modules (`gantry`, `flatbed`, `cradle`) must declare the same Go version.
`loggrpc` is intentionally lower (`go 1.21`) as a library targeting broad compatibility.
`pkg` and `proto` should match the application modules. Check with:

```bash
grep -r "^go " */go.mod
```

If any application module differs from the others, standardize before committing.

## Error mapping

gRPC status codes must map to HTTP responses per `docs/adr/0002`. Verify any new
HTTP handlers that call gRPC services apply the correct mapping.

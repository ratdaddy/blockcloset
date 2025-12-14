# 0012: Error Testing Only for Transformations

- **Status**: Adopted
- **Date**: 2025-12-13
- **Author**: Brian VanLoo

## Context

When writing tests for Go code that handles errors, there's a question of when and where to test error conditions. Testing every `if err != nil { return err }` statement would result in excessive test code with little value, while not testing error handling at integration boundaries could miss critical bugs in error transformation logic.

The codebase has multiple layers with different responsibilities. Some layers transform errors between types (e.g., gRPC status codes to HTTP status codes), while others simply propagate errors unchanged. Without a clear testing pattern, the codebase could end up with inconsistent test coverage—some areas over-tested with redundant error propagation tests, others missing critical error transformation tests.

## Decision

**Test error conditions if and only if the error is being transformed in a non-trivial way.**

### What Constitutes "Transformation"

**Test errors when:**
- Converting between error types (e.g., gRPC codes → HTTP status codes)
- Adding context or detail to errors (wrapping with additional information)
- Mapping internal errors to public API error messages
- Generating semantic/business errors (e.g., "bucket already exists", "no cradle servers available")
- Validating inputs and returning specific error messages
- Parsing/transforming data that could fail (e.g., timestamp parsing)

**Do not test errors when:**
- Simply propagating errors unchanged (`if err != nil { return err }`)
- Wrapping with generic context that doesn't change the error type or meaning
- Passing through errors from dependencies without interpretation

High-level handlers at integration boundaries (HTTP/gRPC endpoints) should have comprehensive error tests because they transform errors to external formats. Low-level clients and wrappers should not test error propagation, though they should test any transformation logic they perform.

## Consequences

- The codebase will have less test code overall, with tests focused on meaningful error transformations rather than trivial propagation.
- Developers must use judgment about what constitutes "transformation" versus "propagation," but the reduction in redundant test code and clearer focus on integration boundaries makes this trade-off worthwhile.
- Error transformation bugs will be caught at the layers where errors are actually transformed, not redundantly tested throughout the call stack.

## Alternatives Considered

### Test all error paths everywhere
*Rejected because*: This creates massive test code overhead with little value. Testing trivial error propagation (`if err != nil { return err }`) doesn't catch meaningful bugs and significantly increases test maintenance burden.

### Never test error handling
*Rejected because*: Error transformation bugs at API boundaries are critical. Without tests, we could return incorrect HTTP status codes, missing error details, or malformed error responses to clients.

### Test errors only at the top-most layer (handlers)
*Rejected because*: Semantic errors generated deep in the stack have business meaning that should be verified where they're generated. For example, storage layer tests should verify that "bucket already exists" errors are returned correctly, not just assume handlers will map them properly.

## References

- Established during comprehensive codebase testing review (2025-12-13)
- Consistent with Go testing best practices: focus on behavior, not implementation

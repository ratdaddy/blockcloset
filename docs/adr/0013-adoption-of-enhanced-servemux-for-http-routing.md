# 0013: Adoption of Enhanced ServeMux for HTTP Routing

- **Status**: Adopted
- **Date**: 2025-12-13
- **Author**: Brian VanLoo
- **Supersedes**: [0004](0004-adoption-of-chi-router-for-http-routing.md)

## Context

ADR 0004 adopted the chi router to address limitations in Go's standard `http.ServeMux`, which at the time lacked support for path parameters, method-based routing, and flexible route matching.

Go 1.22 introduced significant enhancements to `http.ServeMux` that address these limitations:
- Method-based routing (e.g., `"GET /path"`, `"PUT /bucket/{key}"`)
- Path parameters with wildcards (e.g., `{key...}` for multi-segment captures)
- Exact path matching with `{$}` to prevent unintended prefix matching

These enhancements eliminate the primary motivations for using a third-party router while maintaining the advantages of using the standard library: zero external dependencies, idiomatic Go patterns, and guaranteed long-term support.

## Decision

Use Go's enhanced `http.ServeMux` (available in Go 1.22+) for HTTP routing in all services. Remove the chi router dependency.

The enhanced ServeMux provides:
- **Method-based routing**: `mux.HandleFunc("PUT /{bucket}", handler)` matches only PUT requests
- **Path parameters**: `{bucket}` and `{key...}` capture path segments
- **Exact matching**: `/{$}` matches only the root path, not all paths
- **Standard library**: No external dependencies, consistent with Go's stability guarantees

## Consequences

- The codebase uses only standard library components for HTTP routing, eliminating the chi dependency.
- All routing patterns are expressed using Go's native syntax.
- The code remains maximally portable and benefits from Go's backward compatibility guarantees.
- Path parameter extraction uses `r.PathValue("name")` instead of third-party APIs.

## Alternatives Considered

### Continue using chi router
*Rejected because*: The enhanced ServeMux provides equivalent functionality without external dependencies. Continuing to use chi would maintain an unnecessary dependency and non-standard patterns when the standard library now suffices.

### Adopt a heavier framework (e.g., Gin, Echo)
*Rejected because*: Same rationale as ADR 0004. These frameworks introduce additional abstraction layers and non-standard patterns that are unnecessary given the enhanced ServeMux capabilities.

## References

- [Go 1.22 Release Notes - Enhanced ServeMux](https://go.dev/blog/routing-enhancements)
- [net/http package documentation](https://pkg.go.dev/net/http)
- Supersedes [ADR 0004: Adoption of chi Router for HTTP Routing](0004-adoption-of-chi-router-for-http-routing.md)

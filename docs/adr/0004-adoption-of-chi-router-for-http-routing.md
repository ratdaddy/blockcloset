# 0004: Adoption of chi Router for HTTP Routing

* **Status**: Adopted
* **Date**: 2025-08-11
* **Author**: Brian VanLoo

## Context

The initial HTTP implementation for gateway services used Go’s standard library `http.ServeMux`. While sufficient for minimal prototypes, it lacks built-in support for path parameters, middleware composition, and more flexible route matching.

As the project grows, routing requirements will expand to include:

- Named path parameters for resources such as buckets and objects.
- Middleware for logging, metrics, and request shaping.
- Fine-grained method handling with clear separation of concerns.
- The ability to extend the routing layer without introducing complex custom code.

## Decision

Use [`chi`](https://github.com/go-chi/chi) as the HTTP router for new and refactored HTTP services provided they don't have extremely simple path schemas nor the need for middleware.

`chi` was chosen because it:

- **Remains idiomatic Go** — it uses `net/http` interfaces and standard patterns.
- **Provides path parameter parsing** without introducing large frameworks.
- **Has a minimal dependency footprint** and small API surface.
- **Supports middleware chaining** in a clear, composable way.
- **Is proven in production** and actively maintained.

## Consequences

- All new HTTP routers of non-trivial complexity should be implemented with `chi` rather than `http.ServeMux`.
- Code can take advantage of `chi.URLParam` for clean extraction of path parameters.
- Middleware can be applied at the router or route-group level without boilerplate.

## Alternatives Considered

- **Continue with `http.ServeMux`**: Rejected due to the need for significant custom code to support path parameters and middleware consistently.
- **Adopt a heavier framework (e.g., Gin, Echo)**: Rejected to avoid additional abstraction layers, non-standard patterns, and tighter coupling to framework-specific APIs.

## References

- [chi documentation](https://github.com/go-chi/chi)
- [Go net/http package](https://pkg.go.dev/net/http)


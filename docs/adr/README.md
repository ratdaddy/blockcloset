# Architecture Decision Records Index

This index lists all ADRs for the project and their current implementation state.
It represents the **set of governing architectural decisions** currently in effect.
Superseded and deprecated ADRs remain in the repository for historical purposes but are not listed here.

---

## Active ADRs

Only ADRs in the **Adopted** state appear here.

| ADR # | Title | Implementation | Summary |
|-------|-------|----------------|---------|
| [0001](0001-adr-process-and-template-usage.md) | ADR Process and Template Usage | Complete | Defines how ADRs are created, maintained, and used by humans and AI agents. |
| [0002](0002-external-api-s3-compatible-http-subset.md) | External API: S3-Compatible HTTP Subset | Partial | Specifies the design and implementation of an S3-compatible HTTP API for external access. |
| [0003](0003-unit-and-routing-testing-for-go-http-services.md) | Unit and Routing Testing for Go HTTP Services | Complete | Outlines the testing strategy for Go HTTP services, including unit tests and routing tests. |
| [0004](0004-adoption-of-chi-router-for-http-routing.md) | Adoption of Chi Router for HTTP Routing | Complete | Details the decision to use the Chi router for HTTP routing in Go services. |

---

## Considering

ADRs here are being actively evaluated but no final decision has been made.

| ADR # | Title | Summary |
|-------|-------|---------|
| *None currently* | | |

---

## Future

ADRs here record context, constraints, and possible alternatives for decisions that will be made in the future.

| ADR # | Title | Summary |
|-------|-------|---------|
| *None currently* | | |

---

*Note*:
- Update this index whenever an ADR is created, adopted, superseded, deprecated, or changes implementation state.
- “Implementation” column tracks the extent to which the codebase reflects the decision: **Complete**, **Partial**, or **None**.

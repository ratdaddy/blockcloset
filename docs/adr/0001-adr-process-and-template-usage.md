# 0001: ADR Process and Template Usage

- **Status**: Adopted
- **Date**: 2025-08-09
- **Author**: Brian VanLoo

## Context

Architecture Decision Records (ADRs) are structured, version-controlled documents that capture **why** a decision was made, **what** was decided, and **its expected consequences**.
They serve as a durable knowledge base for both human maintainers and AI coding agents, ensuring that the reasoning behind key choices remains accessible and consistent over the life of the project.

For AI agents, ADRs enable:
- Understanding of the rationale and constraints behind decisions.
- Avoidance of changes that contradict established design principles.
- Identification of when a decision should be revisited or superseded.
- Quick navigation to related ADRs and relevant code sections.

For humans, ADRs reduce the need to reconstruct decision history from commit logs, meetings, or memory, especially when trade-offs or subtle constraints influenced the outcome.

## Decision

This project will maintain ADRs in the `docs/adr/` directory, numbered sequentially starting at `0001`.

### ADR Template

When creating a new ADR, copy and paste the following structure and fill it in:

```markdown
# [ADR Number]: [Short Descriptive Title]

- **Status**: [Adopted | Superseded | Deprecated | Considering | Future]
- **Date**: YYYY-MM-DD
- **Author**: [Name(s)]

## Context

[Background, problem statement, constraints, assumptions, and trade-offs considered.]

## Decision

[The chosen solution, or proposed solution if not yet adopted.]

## Consequences

[Anticipated positive and negative outcomes, including impact on maintainability, performance, and AI agent reasoning.]

## Alternatives Considered

[List other viable options that were evaluated, along with the reasons they were not selected.]

## References

[Links to related ADRs, source code files, tickets, or external docs.]
```


### Index Maintenance

The `docs/adr/README.md` file will serve as the **index** for all ADRs and must be kept up to date whenever an ADR is created, adopted, superseded, deprecated, or changes status.

The index must:
- Include clickable links to each ADR using the format `[0001](0001-adr-title.md)` so it works both in GitHub and in plain Markdown viewers.
- Contain three sections:
  1. **Active ADRs** — Only ADRs in the Adopted state, each with an *Implementation* status.
  2. **Considering** — ADRs actively being evaluated.
  3. **Future** — ADRs that will require a decision in the future and may already contain context or alternatives.
- Track **Implementation** state for Active ADRs using:
  - **Complete** — Codebase fully reflects the decision.
  - **Partial** — Decision partially implemented in the codebase.
  - **None** — Decision adopted but no implementation work started.

### Maintenance Rules

- **Version Control**: All ADRs are plain Markdown files tracked in Git.
- **Immutability**: Adopted ADRs are not altered except for clarifications or typo fixes; changes to a decision require a new ADR that supersedes the old one.
- **Linking**: ADRs should cross-reference related decisions and affected code modules.
- **AI Readability**: Use explicit, consistent wording; avoid unnecessary jargon; clearly describe constraints, dependencies, and rejected alternatives.

## Consequences

- Creates a unified decision log that supports human and AI contributors.
- Reduces the risk of architectural drift or contradictory changes.
- Provides AI agents with reliable, structured context for automated reasoning.
- Enables traceability from code to the decision logic that shaped it.
- Makes it clear why certain paths were *not* taken, helping avoid re-evaluation of already-discarded ideas.
- Ensures a single, authoritative index of all architectural decisions.

## Alternatives Considered

- **Omitting rejected alternatives entirely**
  - *Rejected because*: This would remove valuable context for why specific paths were not chosen, increasing the risk of re-evaluating them unnecessarily in the future.
- **Relying solely on commit messages for decision history**
  - *Rejected because*: Commit messages are tied to implementation details and are often too granular, inconsistent, or incomplete to reconstruct architectural reasoning.
- **Maintaining decisions only in a separate design document**
  - *Rejected because*: Large, monolithic design documents are harder to keep updated, less discoverable for AI agents, and more prone to being ignored during day-to-day development.

## References

- [Michael Nygard’s ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions.html)
- [AWS ADR Best Practices](https://docs.aws.amazon.com/prescriptive-guidance/latest/architectural-decision-records/adr-process.html)
- [Lightweight ADRs](https://adr.github.io/)

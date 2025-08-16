# 0005: Standardizing Environment Variable for Application Runtime Mode

- **Status**: Adopted
- **Date**: 2025-08-14
- **Author**: Brian VanLoo

## Context

The Twelve-Factor methodology prescribes storing deploy-specific configuration in environment variables, but does not mandate names for them.
Different ecosystems have evolved their own conventions:

- **Rails/Rack**: `RAILS_ENV` or `RACK_ENV` with `development`, `test`, `production`.
- **.NET**: `ASPNETCORE_ENVIRONMENT` / `DOTNET_ENVIRONMENT` with `Development`, `Staging`, `Production`.
- **Spring Boot**: `SPRING_PROFILES_ACTIVE` for profile names such as `dev`, `staging`, `prod`.

In Go, there is no widely adopted `GO_ENV` equivalent, and using `GO_ENV` could be confused with Go toolchain settings.
A consistent, language-agnostic variable name across all services simplifies tooling, deployment pipelines, and operational visibility — especially in polyglot stacks.

## Decision

We will use a single, canonical environment variable:
```
APP_ENV
```

Allowed values and their meanings:

- **`production`** — Live environment serving real users; optimized for performance and stability.
- **`staging`** — Pre-production environment mirroring production for final testing before release.
- **`development`** — Local or shared dev environment used for active feature work.
- **`test`** — Automated testing environment (CI/CD jobs, integration tests).

The `APP_ENV` variable will be the primary source of truth for runtime mode.
Language/framework-specific variables (e.g., `NODE_ENV`, `RAILS_ENV`) may be set in addition to `APP_ENV` when needed for ecosystem compatibility, but `APP_ENV` will remain the authoritative value.

Default behavior:
- If `APP_ENV` is unset, applications will assume `development`.

## Consequences

- Provides consistent runtime mode handling across all services regardless of language.
- Simplifies CI/CD scripts, logging, monitoring, and configuration management.
- Avoids conflicts with language-specific environment variables.
- Requires some additional mapping for frameworks that expect a different variable name.
- Potential for drift if ecosystem-specific variables are set inconsistently with `APP_ENV`.

## Alternatives Considered

- **Use language-specific variables** (`NODE_ENV`, `RAILS_ENV`, etc.)
  *Rejected because*: Leads to inconsistency across services and complicates deployment pipelines.
- **Use `ENV`**
  *Rejected because*: Too generic; conflicts with Unix tools and existing conventions.
- **Use `GO_ENV` for Go services**
  *Rejected because*: Not recognized in Go tooling; could cause confusion with Go build environment variables.

## References

- [The Twelve-Factor App: Config](https://12factor.net/config)
- [.NET Environment Name conventions](https://learn.microsoft.com/en-us/aspnet/core/fundamentals/environments)
- [Rails Environment Variables](https://guides.rubyonrails.org/configuring.html#rails-environment-settings)




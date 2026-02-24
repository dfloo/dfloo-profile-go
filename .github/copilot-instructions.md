# Copilot Instructions for `dfloo-profile-go`

## Mission and scope

- Keep changes small, focused, and aligned with the existing layered architecture.
- Prefer root-cause fixes over superficial patches.
- Do not introduce broad refactors unless explicitly requested.

## Architecture and dependency boundaries

- Preserve dependency direction: `internal/router` -> `internal/handler` -> `internal/repository` -> database.
- `internal/model` is shared data shape only; it must not depend on repository, handler, or router concerns.
- Keep HTTP concerns in handlers/router and SQL concerns in repository.

## API and behavior stability

- Preserve existing route semantics unless explicitly requested, including auth behavior in router wiring.
- Keep request/response JSON field names backward compatible.
- Keep encoded/decoded ID behavior consistent with middleware helpers.

## Error handling and observability

- Return explicit, deterministic errors; avoid swallowing errors.
- Do not leak sensitive internals in HTTP responses.
- Add context to returned errors when it improves diagnosability.
- Keep logging structured and minimal; avoid noisy logs in hot paths.

## Data and migrations

- Align model/repository changes with migration schema.
- When schema expectations change, update migrations and related repository SQL in the same change.
- Preserve transaction safety for multi-step writes.

## Testing requirements

- For behavior changes, add or update tests in the closest package test file.
- Prefer table-driven tests for handler/latex logic.
- For `internal/repository`, `internal/router`, and `internal/middleware`, require tests where feasible when behavior changes.
- Keep tests deterministic: no real network calls, no dependency on wall-clock timing unless controlled.

## Style and implementation guardrails

- Match existing package style and naming conventions.
- Avoid introducing new dependencies unless clearly justified.
- Do not add inline comments unless necessary to explain non-obvious constraints.

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./...

# Run tests for a single package
go test ./internal/handler/...

# Build the server
go build ./cmd/server/...

# Start local environment (Postgres + API)
docker compose up

# Run DB migrations
docker compose run --rm migrate

# Seed F1 data (one-time, idempotent)
docker compose --profile seed up f1-loader
```

Required environment variables (see `.env`): `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `AUTH0_DOMAIN`, `AUTH0_AUDIENCE`, `CLIENT_ORIGIN`.

## Architecture

Strict layered dependency: `router` → `handler` → `repository` → Postgres (pgx/v5 pool).

- **`internal/model`** — shared data structs with JSON tags; no HTTP/SQL/middleware imports allowed.
- **`internal/repository`** — SQL only; interfaces defined here let handlers depend on abstractions, enabling mock-based handler tests.
- **`internal/handler`** — thin HTTP orchestration; parses requests, calls repo, maps errors to status codes. Uses `middleware.GetUserID`, `middleware.EncodeID`/`DecodeID`, and `middleware.HasPermission` for all user/permission work.
- **`internal/router`** — wires dependencies and registers routes. Uses `middleware.Core` (CORS + logging) for public routes and `middleware.CoreAuthenticated` (Core + JWT validation) for protected routes.
- **`internal/middleware`** — Auth0 JWT/JWKS validation, CORS, logging, base64 ID encode/decode. User identity lives in `ValidatedClaims` under `jwtmiddleware.ContextKey{}`.
- **`internal/latex`** — deterministic PDF generation via `lualatex`; no network calls, bounded temp usage.
- **`internal/testutil`** — shared test fixtures and mock helpers (`MockGetUserID`, `MockEncodeID`, `MockDecodeID`, `CreateRequestWithUserID`, etc.).
- **`cmd/server`** — entry point; initializes `pgxpool.Pool` and wires `router.New`.
- **`cmd/f1-loader`** — one-shot F1 CSV importer; idempotent (fails if only some tables are populated).

## Testing patterns

Tests are `httptest`-based and table-driven; no real DB or network calls. Handler tests inject mock repositories via interfaces. Use `testutil` fixtures rather than duplicating payloads. Repository and middleware tests also exist where feasible. CI runs `go test ./...`.

## Key conventions

- **ID encoding**: all external-facing IDs are base64-encoded via `middleware.EncodeID`/`DecodeID`. Never expose raw internal IDs.
- **Error mapping**: 400 bad input, 401/403 auth, 404 missing resource, 500 unexpected. Do not leak internal error details in response bodies.
- **Schema changes**: update migrations in `db/migrations/` and related repository SQL in the same change.
- **Route auth posture**: `Core` vs `CoreAuthenticated` wrapping is intentional — preserve it unless requirements change explicitly.
- **No new dependencies** unless clearly justified.

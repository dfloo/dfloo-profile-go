---
applyTo: "internal/repository/*"
---

# Repository package instructions

- Keep this layer responsible for SQL, transactions, and persistence mapping only.
- Do not introduce HTTP concerns (status codes, request/response types) in repository code.
- Keep SQL statements explicit and deterministic; avoid hidden side effects.
- Use transactions for multi-step writes and preserve rollback safety on partial failures.
- Keep nullable/optional field handling explicit (`NULL` semantics, pointer fields, and scan targets).
- Return actionable errors with enough context for handlers to map to stable responses.
- Keep schema assumptions synchronized with `db/migrations` in the same change when altered.
- For behavior changes, add or update repository tests where feasible.

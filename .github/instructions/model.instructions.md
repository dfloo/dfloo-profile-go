---
applyTo: "internal/model/*"
---

# Model package instructions

- Keep models as shared data structures only; no HTTP, SQL, or middleware dependencies.
- Preserve JSON tags and field names for API compatibility unless a migration plan is included.
- Keep time and optional/null field semantics explicit and consistent with repository scanning behavior.
- Align nested struct changes with JSONB storage expectations in migrations and repository SQL.
- Prefer additive, backward-compatible schema evolution over breaking shape changes.
- When modifying model fields used by persistence, update related repository code and tests in the same change.

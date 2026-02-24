---
applyTo: "internal/testutil/*"
---

# Test utility package instructions

- Keep helpers deterministic, side-effect free, and fast.
- Prefer reusable fixture builders over ad-hoc duplicated test payloads.
- Do not hide important assertions inside helpers; keep assertions in test files.
- Keep helper APIs small and explicit so callsites remain readable.
- Avoid introducing production-only dependencies into test utilities.
- When adding new helpers, ensure they support existing table-driven test style.

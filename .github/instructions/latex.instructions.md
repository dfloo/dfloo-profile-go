---
applyTo: "internal/latex/*"
---

# LaTeX package instructions

- Keep this package focused on document generation concerns only.
- Preserve strict LaTeX escaping behavior for all user-provided text fields.
- Keep formatting and transformation helpers deterministic and side-effect free.
- Preserve process-safety behavior for `lualatex` execution (timeouts, bounded temp usage, and explicit error returns).
- Keep the current template path contract compatible with container runtime expectations.
- Do not add network calls or external service dependencies here.
- For behavior changes, add or update table-driven tests in `internal/latex/latex_test.go`.

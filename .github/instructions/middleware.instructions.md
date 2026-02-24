---
applyTo: "internal/middleware/*"
---

# Middleware package instructions

- Keep middleware focused on cross-cutting HTTP concerns (auth, CORS, request logging, context propagation).
- Preserve JWT/JWKS validation behavior and fail closed on configuration or verification errors.
- Keep claim extraction helpers deterministic and defensive against missing/invalid claim shapes.
- Avoid leaking token contents or sensitive details through logs/errors.
- Keep CORS behavior explicit and stable; do not silently broaden allowed origins/methods/headers.
- Keep ID encode/decode helpers backward compatible for existing handlers/routes.
- For behavior changes, add or update tests in `internal/middleware` where feasible.

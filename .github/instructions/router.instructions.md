---
applyTo: "internal/router/*"
---

# Router package instructions

- Keep router code focused on dependency wiring and route registration.
- Preserve existing route paths, methods, and auth posture unless explicitly requested.
- Keep `Core` vs `CoreAuthenticated` usage intentional and consistent with current behavior.
- Preserve the current unauthenticated default resume PDF download route behavior unless requirements change.
- Avoid embedding business logic in route closures; delegate to handlers.
- Maintain explicit method gating and deterministic `405` handling when methods are unsupported.
- For route/auth behavior changes, add or update tests where feasible.

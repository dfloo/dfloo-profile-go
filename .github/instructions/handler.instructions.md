---
applyTo: "internal/handler/*"
---

# Handler package instructions

- Keep handlers as thin HTTP orchestration layers; business/data logic belongs in repository or dedicated helpers.
- Use constructor injection and interfaces for dependencies to preserve testability.
- Parse request JSON strictly and validate required fields before repository calls.
- Map failures to stable HTTP status codes consistently:
	- `400` for malformed input/validation errors
	- `401/403` for auth/permission failures
	- `404` for missing resources
	- `500` for unexpected internal failures
- Do not expose internal error details in response bodies.
- Always use middleware helpers for user/permission extraction and ID encode/decode.
- Keep response payload schemas backward compatible.
- Prefer small private helper functions to avoid duplicated response/error handling logic.
- For changes in handler behavior, add or update `httptest`-based tests in `internal/handler/*_test.go`.


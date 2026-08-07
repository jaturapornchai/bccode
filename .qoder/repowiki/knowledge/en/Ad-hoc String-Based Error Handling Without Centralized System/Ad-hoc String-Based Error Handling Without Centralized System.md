---
kind: error_handling
name: Ad-hoc String-Based Error Handling Without Centralized System
category: error_handling
scope:
    - '**'
source_files:
    - backend/error.md
    - backend/internal/authentication/models/authentication.go
---

This repository does not implement a centralized error handling system. Instead, errors are handled ad-hoc throughout the Go backend using scattered patterns:

**Approach used:**
- Plain `errors.New()` and `fmt.Errorf(...)` calls with human-readable string messages returned directly from functions
- No dedicated `errors` package or shared error types (beyond one custom `UserDisableLoginError` struct in `backend/internal/authentication/models/authentication.go`)
- No sentinel error variables defined via `var ErrNotFound = errors.New(...)` pattern
- No structured error codes or HTTP status mapping layer
- No middleware-based error normalization at the HTTP boundary
- No panic/recover strategy — panics are not used for control flow

**Evidence of inconsistency:**
- The file `backend/error.md` is a manually maintained catalog of raw error strings across services (e.g., "document not found", "code is exists", "username is not exists") — indicating developers track these informally rather than through code
- Many error messages are grammatically incorrect or inconsistent in style ("password is not invalid", "code is exists", "user payload invalid")
- Some callers wrap errors with `%w` for wrapping (`fmt.Errorf("put object: %w", err)`), while others return raw strings without context
- Database-specific errors like `mongo: no documents in result` are compared as plain strings rather than via sentinel checks

**HTTP-level handling:**
- There is no global error handler middleware that converts domain errors into consistent JSON responses
- Each HTTP handler appears to handle its own error-to-response mapping inline

**Frontend (Next.js):**
- Frontend error handling is not part of this category's scope; the focus here is on the Go backend.

**Consequences:**
- Hard to uniformly map errors to HTTP status codes
- Difficult to test error paths without parsing string messages
- No way to programmatically distinguish error categories (validation vs. not-found vs. conflict) without string matching
- Adding new error types requires touching many call sites instead of a central registry

**Rules developers should follow (informal conventions observed):**
- Return `error` values from service-layer functions
- Use descriptive English strings (though consistency is poor)
- Wrap lower-level errors with `fmt.Errorf("...: %w", err)` when adding context
- Do not use `panic` for expected error conditions
- Document new error strings in `backend/error.md` if they represent user-visible messages
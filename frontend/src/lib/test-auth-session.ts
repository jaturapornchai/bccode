import { afterEach, beforeEach } from "vitest";

import { clearAuthSession, setAuthSession } from "./client-auth-session";

/**
 * Module clients call `apiFetch`, which refuses to send a request without a
 * session — that refusal is the point: an anonymous request comes back 401 and
 * a screen renders as "no data". Tests that stub `fetch` therefore need a
 * session in place, otherwise they exercise the refusal instead of the code
 * under test.
 */
export function signInTestSession(): void {
  setAuthSession({
    token: "test-token",
    username: "test",
    backendUrl: "http://localhost:8888",
  });
}

/**
 * Call at file scope, not inside a `describe` — a `beforeEach` registered in one
 * describe block does not run for the others in the same file.
 */
export function setupTestAuthSession(): void {
  beforeEach(signInTestSession);
  afterEach(() => {
    clearAuthSession();
  });
}

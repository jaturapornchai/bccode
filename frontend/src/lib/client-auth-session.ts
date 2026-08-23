import type { AuthSession } from "./workspace-models";
import { workspaceStorageKeys } from "./workspace-models";

type StoredAuthSession = Omit<AuthSession, "token">;

let activeSession: AuthSession | null = null;
let refreshRequest: Promise<AuthSession | null> | null = null;
let authChannel: BroadcastChannel | null = null;

type AuthChannelMessage =
  | { type: "session"; session: AuthSession }
  | { type: "logout" };

function clearLocalAuthSession(): void {
  activeSession = null;
  refreshRequest = null;
  storage()?.removeItem(workspaceStorageKeys.auth);
}

function sessionChannel(): BroadcastChannel | null {
  if (authChannel || typeof window === "undefined" || typeof window.BroadcastChannel === "undefined") {
    return authChannel;
  }
  authChannel = new window.BroadcastChannel("bc-auth-session");
  authChannel.addEventListener("message", (event: MessageEvent<AuthChannelMessage>) => {
    if (event.data?.type === "logout") {
      clearLocalAuthSession();
      return;
    }
    if (event.data?.type === "session" && event.data.session?.token) {
      activeSession = event.data.session;
    }
  });
  return authChannel;
}

function storage(): Storage | null {
  return typeof localStorage === "undefined" ? null : localStorage;
}

function storedSession(): StoredAuthSession | null {
  const target = storage();
  if (!target) return null;
  try {
    const parsed = JSON.parse(target.getItem(workspaceStorageKeys.auth) ?? "null") as Partial<AuthSession> | null;
    if (!parsed || typeof parsed.username !== "string" || typeof parsed.backendUrl !== "string") {
      target.removeItem(workspaceStorageKeys.auth);
      return null;
    }
    const session = {
      username: parsed.username,
      backendUrl: parsed.backendUrl,
      method: parsed.method,
      holdingcode: parsed.holdingcode,
      profile: parsed.profile,
    };
    target.setItem(workspaceStorageKeys.auth, JSON.stringify(session));
    return session;
  } catch {
    target.removeItem(workspaceStorageKeys.auth);
    return null;
  }
}

export function getAuthSession(): AuthSession | null {
  return activeSession;
}

export function setAuthSession(session: AuthSession): void {
  activeSession = session;
  storage()?.setItem(
    workspaceStorageKeys.auth,
    JSON.stringify({
      username: session.username,
      backendUrl: session.backendUrl,
      method: session.method,
      holdingcode: session.holdingcode,
      profile: session.profile,
    } satisfies StoredAuthSession),
  );
  sessionChannel()?.postMessage({ type: "session", session } satisfies AuthChannelMessage);
}

export function clearAuthSession(): void {
  clearLocalAuthSession();
  sessionChannel()?.postMessage({ type: "logout" } satisfies AuthChannelMessage);
}

async function withRefreshLock(
  staleToken: string | undefined,
  operation: () => Promise<AuthSession | null>,
): Promise<AuthSession | null> {
  sessionChannel();
  if (typeof navigator !== "undefined" && navigator.locks?.request) {
    return navigator.locks.request("bc-auth-refresh", async () => {
      // Another tab may have refreshed while this tab waited for the lock and
      // delivered the new in-memory access token through BroadcastChannel.
      if (activeSession && (!staleToken || activeSession.token !== staleToken)) return activeSession;
      return operation();
    });
  }
  return operation();
}

async function refreshAuthSession(staleToken?: string): Promise<AuthSession | null> {
  if (refreshRequest) return refreshRequest;

  const stored: StoredAuthSession | null = activeSession
    ? {
        username: activeSession.username,
        backendUrl: activeSession.backendUrl,
        method: activeSession.method,
        holdingcode: activeSession.holdingcode,
        profile: activeSession.profile,
      }
    : storedSession();
  if (!stored) return null;

  refreshRequest = withRefreshLock(staleToken, async () => {
    try {
      const response = await fetch("/api/auth/refresh", {
        method: "POST",
        credentials: "same-origin",
        cache: "no-store",
      });
      const payload = (await response.json().catch(() => null)) as { success?: boolean; token?: string } | null;
      if (!response.ok || !payload?.success || !payload.token) {
        if (response.status === 401 || response.status === 403) clearAuthSession();
        return null;
      }
      const session = { ...stored, token: payload.token };
      setAuthSession(session);
      return session;
    } catch {
      return null;
    } finally {
      refreshRequest = null;
    }
  });

  return refreshRequest;
}

export async function restoreAuthSession(): Promise<AuthSession | null> {
  return activeSession ?? refreshAuthSession();
}

function requestUrl(input: RequestInfo | URL): string {
  return input instanceof Request ? input.url : String(input);
}

function isRefreshableRequest(input: RequestInfo | URL, backendUrl: string): boolean {
  const value = requestUrl(input);
  if (value.startsWith("/")) return true;
  try {
    const baseUrl = typeof window === "undefined" ? undefined : window.location.href;
    const requestOrigin = baseUrl ? new URL(value, baseUrl).origin : new URL(value).origin;
    const backendOrigin = baseUrl ? new URL(backendUrl, baseUrl).origin : new URL(backendUrl).origin;
    return requestOrigin === backendOrigin || (
      typeof window !== "undefined" && requestOrigin === window.location.origin
    );
  } catch {
    return false;
  }
}

function mergedHeaders(input: RequestInfo | URL, init?: RequestInit): Headers {
  const headers = new Headers(input instanceof Request ? input.headers : undefined);
  new Headers(init?.headers).forEach((value, key) => headers.set(key, value));
  return headers;
}

/** Fetch a protected browser API and refresh an expired access token once. */
export async function authFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const headers = mergedHeaders(input, init);
  const authorization = headers.get("authorization") ?? "";
  const session = activeSession;
  if (!session || !/^Bearer\s+\S+$/i.test(authorization) || !isRefreshableRequest(input, session.backendUrl)) {
    return fetch(input, init);
  }

  headers.set("Authorization", `Bearer ${session.token}`);
  const retryInput = input instanceof Request ? input.clone() : input;
  const response = await fetch(input, { ...init, headers });
  if (response.status !== 401) return response;

  const refreshed = await refreshAuthSession(session.token);
  if (!refreshed) return response;

  headers.set("Authorization", `Bearer ${refreshed.token}`);
  return fetch(retryInput, { ...init, headers });
}

export async function logoutAuthSession(): Promise<void> {
  const authorization = activeSession?.token ? { Authorization: `Bearer ${activeSession.token}` } : undefined;
  const response = await authFetch("/api/auth/logout", {
    method: "POST",
    headers: authorization,
    credentials: "same-origin",
    cache: "no-store",
  });
  const payload = (await response.json().catch(() => null)) as { success?: boolean; message?: string } | null;
  if (!response.ok || payload?.success !== true) {
    throw new Error(payload?.message ?? "ไม่สามารถเพิกถอน Session ได้");
  }
  clearAuthSession();
}

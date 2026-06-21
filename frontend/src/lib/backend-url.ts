export type BackendUrlCheck = {
  normalizedGoApiUrl: string;
  mainApiUrl: string;
};

export function publicGoApiUrlForOrigin(origin: string): string {
  return new URL("/backend/goapi", origin).toString().replace(/\/$/, "");
}

export function localGoApiUrlForOrigin(origin: string): string {
  return publicGoApiUrlForOrigin(origin);
}

export function runtimeGoApiUrlForOrigin(origin: string): string {
  return publicGoApiUrlForOrigin(origin);
}

export function migrateRuntimeBackendUrl(rawUrl: string, currentOrigin: string): string {
  try {
    const normalizedUrl = normalizeBackendUrl(rawUrl);
    const parsed = new URL(normalizedUrl);
    const current = new URL(currentOrigin);
    const path = parsed.pathname.replace(/\/+$/, "").toLowerCase();

    if (isDevServerBackend(parsed, current)) {
      return runtimeGoApiUrlForOrigin(current.origin);
    }

    if (parsed.origin === current.origin && path === "/goapi" && !isLocalWebHost(current.hostname)) {
      return publicGoApiUrlForOrigin(current.origin);
    }

    return normalizedUrl;
  } catch {
    return rawUrl;
  }
}

export function migrateSameOriginLegacyGoApiUrl(rawUrl: string, currentOrigin: string): string {
  try {
    const normalizedUrl = normalizeBackendUrl(rawUrl);
    const parsed = new URL(normalizedUrl);
    const current = new URL(currentOrigin);
    const path = parsed.pathname.replace(/\/+$/, "").toLowerCase();

    if (parsed.origin === current.origin && path === "/goapi" && !isLocalWebHost(current.hostname)) {
      return publicGoApiUrlForOrigin(current.origin);
    }

    return normalizedUrl;
  } catch {
    return rawUrl;
  }
}

export function normalizeBackendUrl(rawUrl: string): string {
  let value = rawUrl.trim();
  if (!value) {
    throw new Error("กรุณากรอก Backend URL");
  }

  if (!/^https?:\/\//i.test(value)) {
    value = `http://${value}`;
  }

  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error("รูปแบบ Backend URL ไม่ถูกต้อง");
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new Error("Backend URL ต้องเป็น http หรือ https เท่านั้น");
  }

  parsed.hash = "";
  parsed.search = "";
  parsed.pathname = parsed.pathname.replace(/\/+$/, "");

  return parsed.toString().replace(/\/$/, "");
}

export function deriveMainApiUrl(goApiUrl: string): string {
  const parsed = new URL(normalizeBackendUrl(goApiUrl));
  const path = parsed.pathname.replace(/\/+$/, "");

  if (path.toLowerCase().endsWith("/goapi")) {
    parsed.pathname = path.slice(0, -"/goapi".length) || "/";
  } else {
    parsed.pathname = "/";
  }

  parsed.search = "";
  parsed.hash = "";
  return parsed.toString().replace(/\/$/, "");
}

export function validateBackendUrl(rawUrl: string): BackendUrlCheck {
  const normalizedGoApiUrl = normalizeBackendUrl(rawUrl);
  const mainApiUrl = deriveMainApiUrl(normalizedGoApiUrl);
  const parsed = new URL(mainApiUrl);

  if (parsed.username || parsed.password) {
    throw new Error("Backend URL ห้ามมี username หรือ password");
  }

  return { normalizedGoApiUrl, mainApiUrl };
}

function isLocalWebHost(hostname: string): boolean {
  const normalized = hostname.toLowerCase();
  return normalized === "localhost" || normalized === "127.0.0.1" || normalized === "::1";
}

function isDevServerBackend(parsed: URL, current: URL): boolean {
  const hostname = parsed.hostname.toLowerCase();
  const path = parsed.pathname.replace(/\/+$/, "").toLowerCase();
  if (path !== "/goapi" && path !== "/backend/goapi") return false;

  if (isLocalWebHost(hostname)) {
    return parsed.origin !== current.origin || path === "/goapi";
  }

  return hostname === "45.144.166.112" || hostname === "dev.bcaicloud.com" || hostname === "api.bcaicloud.com";
}

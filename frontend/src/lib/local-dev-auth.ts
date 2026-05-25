export const LOCAL_GOOGLE_TEST_EMAIL = "jaturapornchai@gmail.com";
export const LOCAL_GOOGLE_TEST_NAME = "jaturapornchai";

const LOCAL_LOGIN_HOSTS = new Set(["localhost", "127.0.0.1", "::1"]);
const DEFAULT_DEV_LOGIN_HOSTS = new Set(["dev.bcaicloud.com"]);

export function isLocalLoginHost(hostname: string): boolean {
  return LOCAL_LOGIN_HOSTS.has(hostname.trim().toLowerCase());
}

export function isDevLoginHost(hostname: string, env: Record<string, string | undefined> = process.env): boolean {
  const normalized = hostname.trim().toLowerCase();
  const configuredHosts = (env.BC_DEV_GOOGLE_LOGIN_HOSTS ?? "")
    .split(",")
    .map((host) => host.trim().toLowerCase())
    .filter(Boolean);

  return DEFAULT_DEV_LOGIN_HOSTS.has(normalized) || configuredHosts.includes(normalized);
}

export function getRequestHostName(hostHeader: string | null): string {
  const firstHost = (hostHeader ?? "").split(",")[0]?.trim().toLowerCase() ?? "";
  if (!firstHost) return "";

  if (firstHost.startsWith("[")) {
    const closingBracket = firstHost.indexOf("]");
    return closingBracket > 1 ? firstHost.slice(1, closingBracket) : "";
  }

  return firstHost.split(":")[0] ?? "";
}

export function isLocalLoginRequest(request: Request): boolean {
  const forwardedHost = request.headers.get("x-forwarded-host");
  const host = forwardedHost || request.headers.get("host");
  return isLocalLoginHost(getRequestHostName(host));
}

export function isConfiguredDevLoginRequest(request: Request, env: Record<string, string | undefined> = process.env): boolean {
  if (env.BC_ENABLE_DEV_GOOGLE_LOGIN !== "true") return false;
  const forwardedHost = request.headers.get("x-forwarded-host");
  const host = forwardedHost || request.headers.get("host");
  return isDevLoginHost(getRequestHostName(host), env);
}

import { afterEach, describe, expect, it, vi } from "vitest";

const cookieValues = vi.hoisted(() => new Map<string, string>());

vi.mock("next/headers", () => ({
  cookies: async () => ({
    get: (name: string) => (cookieValues.has(name) ? { name, value: cookieValues.get(name) } : undefined),
  }),
}));

import { getInitialBackendLanguage } from "./backend-language-server";

afterEach(() => {
  cookieValues.clear();
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

describe("getInitialBackendLanguage (SSR preload)", () => {
  it("never fetches the host from the backend_url cookie; it always uses BCAI_LOCAL_BACKEND_URL", async () => {
    vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "http://mainapi.internal:8888");
    cookieValues.set("user_language", "lo");
    cookieValues.set("backend_url", "http://attacker.example/goapi");
    const fetchMock = vi.fn(async () => Response.json({ overview: "ພາບລວມ" }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await getInitialBackendLanguage();

    const fetchedUrls = fetchMock.mock.calls.map((call) => String((call as unknown[])[0]));
    expect(fetchedUrls).toEqual(["http://mainapi.internal:8888/goapi/api/language/lo"]);
    expect(fetchedUrls.some((url) => url.includes("attacker.example"))).toBe(false);
    expect(result.initialLanguage).toBe("lo");
    expect(result.initialBackendLanguage).toEqual({ overview: "ພາບລວມ" });
    // Still handed to the client screens as their fallback backend URL; only the server fetch ignores it.
    expect(result.initialBackendUrl).toBe("http://attacker.example/goapi");
  });

  it("renders with an empty dictionary instead of crashing when BCAI_LOCAL_BACKEND_URL is missing", async () => {
    vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "");
    cookieValues.set("backend_url", "http://attacker.example/goapi");
    const fetchMock = vi.fn(async () => Response.json({ overview: "x" }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await getInitialBackendLanguage();

    expect(fetchMock).not.toHaveBeenCalled();
    expect(result.initialBackendLanguage).toEqual({});
    expect(result.initialLanguage).toBe("th");
  });
});

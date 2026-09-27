import { afterEach, describe, expect, it, vi } from "vitest";
import {
  backendUrlPreferenceCookie,
  languagePreferenceCookie,
  loadBackendLanguageDictionary,
  serializePreferenceCookie,
} from "./backend-language-preload";

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("backend language preload helpers", () => {
  it("serializes language/backend cookies for client preference sync", () => {
    expect(serializePreferenceCookie(languagePreferenceCookie, "lo")).toContain("user_language=lo");
    expect(serializePreferenceCookie(backendUrlPreferenceCookie, "http://localhost:8888/goapi")).toContain(
      "backend_url=http%3A%2F%2Flocalhost%3A8888%2Fgoapi",
    );
  });

  it("loads the dictionary from BCAI_LOCAL_BACKEND_URL/goapi, the server-side backend address", async () => {
    vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "http://mainapi.internal:8888/");
    const fetcher = vi.fn(async () => Response.json({ overview: "ພາບລວມ" }));

    const dictionary = await loadBackendLanguageDictionary("lo", fetcher);

    expect(fetcher).toHaveBeenCalledTimes(1);
    expect(fetcher).toHaveBeenCalledWith(
      "http://mainapi.internal:8888/goapi/api/language/lo",
      expect.objectContaining({ cache: "no-store" }),
    );
    expect(dictionary).toEqual({ overview: "ພາບລວມ" });
  });

  it("returns an empty dictionary without fetching when BCAI_LOCAL_BACKEND_URL is missing", async () => {
    vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "");
    const fetcher = vi.fn(async () => Response.json({ overview: "ພາບລວມ" }));

    await expect(loadBackendLanguageDictionary("lo", fetcher)).resolves.toEqual({});
    expect(fetcher).not.toHaveBeenCalled();
  });

  it("returns an empty dictionary when the backend answers with an error", async () => {
    vi.stubEnv("BCAI_LOCAL_BACKEND_URL", "http://mainapi.internal:8888");
    const fetcher = vi.fn(async () => Response.json({ message: "boom" }, { status: 500 }));

    await expect(loadBackendLanguageDictionary("lo", fetcher)).resolves.toEqual({});
  });
});

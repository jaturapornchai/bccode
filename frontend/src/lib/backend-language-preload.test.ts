import { describe, expect, it, vi } from "vitest";
import {
  backendUrlPreferenceCookie,
  languagePreferenceCookie,
  loadBackendLanguageDictionary,
  serializePreferenceCookie,
} from "./backend-language-preload";

describe("backend language preload helpers", () => {
  it("serializes language/backend cookies for client preference sync", () => {
    expect(serializePreferenceCookie(languagePreferenceCookie, "lo")).toContain("user_language=lo");
    expect(serializePreferenceCookie(backendUrlPreferenceCookie, "http://localhost:8888/goapi")).toContain(
      "backend_url=http%3A%2F%2Flocalhost%3A8888%2Fgoapi",
    );
  });

  it("loads the backend dictionary from a validated goapi URL", async () => {
    const fetcher = vi.fn(async () => Response.json({ overview: "ພາບລວມ" }));

    const dictionary = await loadBackendLanguageDictionary("lo", "http://localhost:8888/goapi", fetcher);

    expect(fetcher).toHaveBeenCalledWith("http://localhost:8888/goapi/api/language/lo", expect.objectContaining({ cache: "no-store" }));
    expect(dictionary).toEqual({ overview: "ພາບລວມ" });
  });

  it("returns an empty dictionary for invalid backend URLs", async () => {
    const fetcher = vi.fn(async () => Response.json({ overview: "ພາບລວມ" }));

    await expect(loadBackendLanguageDictionary("lo", "ftp://localhost:8888/goapi", fetcher)).resolves.toEqual({});
    expect(fetcher).not.toHaveBeenCalled();
  });
});

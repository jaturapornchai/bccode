import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

const context = (path: string) => ({ params: Promise.resolve({ goPath: path.split("/") }) });

const postReport = (path: string) =>
  POST(
    new Request(`http://localhost/api/goapi/${path}`, {
      method: "POST",
      headers: {
        Authorization: "Bearer access-token",
        "Content-Type": "application/json",
        "x-bc-backend-url": "http://localhost:8888/goapi",
      },
      body: JSON.stringify({ fromdate: "2026-01-01", todate: "2026-01-31" }),
    }),
    context(path),
  );

describe("goapi BFF allowlist", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });

  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  // Every path the frontend calls through /api/goapi must be allowlisted; a missing entry is a silent 404.
  it.each(["api/report/tax/wht", "api/report/debt/query", "api/report/tax/vat-register"])(
    "proxies POST %s to mainapi /goapi",
    async (path) => {
      const fetchMock = vi.fn(async (url: string | URL | Request) => {
        expect(String(url)).toBe(`http://localhost:8888/goapi/${path}`);
        return Response.json({ success: true, data: [] });
      });
      vi.stubGlobal("fetch", fetchMock);

      const response = await postReport(path);

      expect(response.status).toBe(200);
      expect(fetchMock).toHaveBeenCalledTimes(1);
    },
  );

  it("rejects paths that are not allowlisted without calling mainapi", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const response = await postReport("api/transaction/calculate");

    expect(response.status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

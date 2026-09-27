import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GET } from "./route";

const context = (key: string[]) => ({ params: Promise.resolve({ key }) });

function request(query = "", headers: Record<string, string> = { Authorization: "Bearer access-token" }) {
  return new Request(`http://localhost/api/files/x${query}`, { headers });
}

describe("file BFF (private S3 objects)", () => {
  beforeEach(() => {
    process.env.BCAI_LOCAL_BACKEND_URL = "http://localhost:8888";
  });
  afterEach(() => {
    delete process.env.BCAI_LOCAL_BACKEND_URL;
    vi.unstubAllGlobals();
  });

  it("streams the object from mainapi with the caller's token and the upstream file headers", async () => {
    const calls: Array<{ url: string; init?: RequestInit }> = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      calls.push({ url: String(url), init });
      return new Response(new Uint8Array([137, 80, 78, 71]), {
        status: 200,
        headers: {
          "Content-Type": "image/png",
          "Content-Length": "4",
          "Cache-Control": "private, max-age=300, must-revalidate",
          ETag: '"abc"',
          "Last-Modified": "Sat, 26 Sep 2026 00:00:00 GMT",
          Vary: "Authorization",
          "X-Content-Type-Options": "nosniff",
          "Set-Cookie": "leak=1",
        },
      });
    }));

    const response = await GET(request("", { Authorization: "Bearer access-token", "If-None-Match": '"old"' }), context(["rungrueng", "system-settings", "logouri", "20260926_ab12.png"]));

    expect(calls).toHaveLength(1);
    expect(calls[0].url).toBe("http://localhost:8888/goapi/s3/file/rungrueng/system-settings/logouri/20260926_ab12.png");
    const sent = calls[0].init?.headers as Record<string, string>;
    expect(sent.Authorization).toBe("Bearer access-token");
    expect(sent["If-None-Match"]).toBe('"old"');
    expect(calls[0].init?.redirect).toBe("manual");
    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toBe("image/png");
    expect(response.headers.get("content-length")).toBe("4");
    expect(response.headers.get("cache-control")).toBe("private, max-age=300, must-revalidate");
    expect(response.headers.get("etag")).toBe('"abc"');
    expect(response.headers.get("vary")).toBe("Authorization");
    expect(response.headers.get("x-content-type-options")).toBe("nosniff");
    expect(response.headers.get("set-cookie")).toBeNull();
    expect([...new Uint8Array(await response.arrayBuffer())]).toEqual([137, 80, 78, 71]);
  });

  it("passes the caller's IP from Caddy's X-Forwarded-For so mainapi rate-limits per user", async () => {
    const sent: Array<Record<string, string>> = [];
    vi.stubGlobal("fetch", vi.fn(async (_url: string | URL | Request, init?: RequestInit) => {
      sent.push(init?.headers as Record<string, string>);
      return new Response(null, { status: 404 });
    }));
    const key = context(["rungrueng", "a.png"]);

    await GET(request("", { Authorization: "Bearer t", "X-Forwarded-For": "203.0.113.7, 172.18.0.1" }), key);
    await GET(request("", { Authorization: "Bearer t", "X-Forwarded-For": "2001:db8::1" }), key);
    await GET(request("", { Authorization: "Bearer t", "X-Forwarded-For": "evil<script>" }), key);
    await GET(request("", { Authorization: "Bearer t" }), key);

    expect(sent.map((headers) => headers["X-Forwarded-For"])).toEqual(["203.0.113.7", "2001:db8::1", undefined, undefined]);
  });

  it("forwards only the thumbnail variant and encodes each key segment", async () => {
    const urls: string[] = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request) => {
      urls.push(String(url));
      return new Response("x", { status: 200, headers: { "Content-Type": "image/webp" } });
    }));

    await GET(request("?variant=thumbnail&other=1"), context(["shop", "images", "a b#?.png"]));
    await GET(request("?variant=THUMBNAIL"), context(["shop", "รูป.png"]));

    expect(urls).toEqual([
      "http://localhost:8888/goapi/s3/file/shop/images/a%20b%23%3F.png?variant=thumbnail",
      `http://localhost:8888/goapi/s3/file/shop/${encodeURIComponent("รูป.png")}?variant=thumbnail`,
    ]);
  });

  it("relays 304 without a body and mainapi errors with their status", async () => {
    const responses = [
      new Response(null, { status: 304, headers: { ETag: '"abc"' } }),
      Response.json({ error: "forbidden" }, { status: 403 }),
    ];
    vi.stubGlobal("fetch", vi.fn(async () => responses.shift()));

    const notModified = await GET(request(), context(["shop", "a.png"]));
    const forbidden = await GET(request(), context(["other", "a.png"]));

    expect(notModified.status).toBe(304);
    expect(notModified.headers.get("etag")).toBe('"abc"');
    expect(await notModified.text()).toBe("");
    expect(forbidden.status).toBe(403);
    expect(await forbidden.json()).toEqual({ error: "forbidden" });
  });

  it("never relays or follows a redirect from upstream", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 302, headers: { Location: "https://evil.example/x" } })));

    const response = await GET(request(), context(["shop", "a.png"]));

    expect(response.status).toBe(502);
    expect(response.headers.get("location")).toBeNull();
  });

  it("rejects traversal, empty, encoded-slash, backslash and control-character segments without calling mainapi", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    for (const key of [
      [],
      ["shop", ".."],
      ["..", "..", "get"],
      ["shop", "."],
      ["shop", "", "a.png"],
      ["shop", "../../get"],
      ["shop", "a/b.png"],
      ["shop", "..\\..\\get"],
      ["shop", "a\u0000.png"],
      ["shop", "a\n.png"],
      ["shop", "x".repeat(256)],
    ]) {
      const response = await GET(request(), context(key));
      expect(response.status, JSON.stringify(key)).toBe(400);
    }
    expect((await GET(request("?variant=original"), context(["shop", "a.png"]))).status).toBe(400);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("requires a Bearer token and ignores any client-supplied backend host", async () => {
    const urls: string[] = [];
    vi.stubGlobal("fetch", vi.fn(async (url: string | URL | Request) => {
      urls.push(String(url));
      return new Response("x", { status: 200 });
    }));

    expect((await GET(request("", {}), context(["shop", "a.png"]))).status).toBe(401);
    expect((await GET(request("", { Authorization: "Basic dXNlcjpwYXNz" }), context(["shop", "a.png"]))).status).toBe(401);
    await GET(
      new Request("http://localhost/api/files/shop/a.png?backendUrl=https://evil.example/goapi", {
        headers: { Authorization: "Bearer access-token", "x-bc-backend-url": "https://evil.example/goapi", Host: "evil.example" },
      }),
      context(["shop", "a.png"]),
    );

    expect(urls).toEqual(["http://localhost:8888/goapi/s3/file/shop/a.png"]);
  });
});

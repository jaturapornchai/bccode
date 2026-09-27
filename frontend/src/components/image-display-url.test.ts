import { describe, expect, it } from "vitest";
import { imageDisplayUrl } from "./authenticated-image";
import { resolveDisplayUrl } from "./logo-avatar";

// Stored image URIs are "/goapi/s3/file/<key>". There is no public /backend proxy to mainapi any more
// (next.config.ts has no rewrites), so the browser reads them through the same-origin BFF
// src/app/api/files/[...key]/route.ts.
describe("image display URL builders (file BFF)", () => {
  it("resolves stored S3 URIs to the same-origin /api/files path, keeping the key and query", () => {
    for (const build of [imageDisplayUrl, resolveDisplayUrl]) {
      expect(build("/goapi/s3/file/tenant/a.png")).toBe("/api/files/tenant/a.png");
      expect(build("/goapi/s3/file/tenant/system-settings/logouri/a.png?variant=thumbnail")).toBe(
        "/api/files/tenant/system-settings/logouri/a.png?variant=thumbnail",
      );
    }
  });

  it("keeps same-origin BFF, absolute, data/blob and static URIs untouched", () => {
    for (const build of [imageDisplayUrl, resolveDisplayUrl]) {
      expect(build("/api/upload/image/x")).toBe("/api/upload/image/x");
      expect(build("https://cdn.example/x.png")).toBe("https://cdn.example/x.png");
      expect(build("data:image/png;base64,AAAA")).toBe("data:image/png;base64,AAAA");
      expect(build("/banks/kbank.png")).toBe("/banks/kbank.png");
    }
  });

  it("never builds a /backend URL", () => {
    expect(imageDisplayUrl("images/x.png")).toBe("");
    expect(imageDisplayUrl("x.png")).toBe("");
    expect(resolveDisplayUrl("images/x.png")).toBe("images/x.png");
    for (const build of [imageDisplayUrl, resolveDisplayUrl]) {
      for (const raw of ["images/x.png", "x.png", "/api/x", "/goapi/s3/file/a.png", "/goapi/api/x", "/s3/file/a.png"]) {
        expect(build(raw), raw).not.toMatch(/\/backend(\/|$)/);
      }
    }
  });
});

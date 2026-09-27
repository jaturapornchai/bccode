import { describe, expect, it } from "vitest";
import { imageDisplayUrl } from "./authenticated-image";
import { resolveDisplayUrl } from "./logo-avatar";

// The browser's backendUrl is "<origin>/backend/goapi" (publicGoApiUrlForOrigin in src/lib/backend-url.ts).
// Stored image URIs are "/goapi/s3/file/<key>", and the public proxy forwards only
// "/backend/goapi/s3/file/:path+" for images (next.config.ts), so that is the one shape that must resolve there.
const backendUrl = "https://account.example/backend/goapi";

describe("image display URL builders (public /backend allowlist)", () => {
  it("resolves stored S3 URIs to the allowlisted /backend/goapi/s3/file path", () => {
    for (const build of [imageDisplayUrl, resolveDisplayUrl]) {
      expect(build("/goapi/s3/file/tenant/a.png", backendUrl)).toBe(
        "https://account.example/backend/goapi/s3/file/tenant/a.png",
      );
      expect(build("/goapi/s3/file/a.png?variant=thumbnail", backendUrl)).toBe(
        "https://account.example/backend/goapi/s3/file/a.png?variant=thumbnail",
      );
    }
  });

  it("keeps same-origin BFF, absolute, data/blob and static URIs untouched", () => {
    for (const build of [imageDisplayUrl, resolveDisplayUrl]) {
      expect(build("/api/upload/image/x", backendUrl)).toBe("/api/upload/image/x");
      expect(build("https://cdn.example/x.png", backendUrl)).toBe("https://cdn.example/x.png");
      expect(build("data:image/png;base64,AAAA", backendUrl)).toBe("data:image/png;base64,AAAA");
      expect(build("/banks/kbank.png", backendUrl)).toBe("/banks/kbank.png");
    }
  });

  it("never builds the removed /backend/images/* or /backend/api/* URLs", () => {
    expect(imageDisplayUrl("images/x.png", backendUrl)).toBe("");
    expect(imageDisplayUrl("x.png", backendUrl)).toBe("");
    expect(resolveDisplayUrl("images/x.png", backendUrl)).toBe("images/x.png");
    for (const build of [imageDisplayUrl, resolveDisplayUrl]) {
      for (const raw of ["images/x.png", "x.png", "/api/x", "/goapi/s3/file/a.png"]) {
        expect(build(raw, backendUrl), raw).not.toMatch(/\/backend\/(images|api)\//);
      }
    }
  });
});

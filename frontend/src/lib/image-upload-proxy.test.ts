import { describe, expect, it, vi } from "vitest";
import { PRODUCT_VIDEO_REQUEST_MAX_BYTES } from "@/lib/product-barcode/types";
import {
  fileBffUrl,
  imageNeedsAuthenticatedFetch,
  imageThumbnailProxyUrl,
  normalizeImageUploadPayload,
  proxyImageUploadToGoApi,
} from "./image-upload-proxy";

describe("image upload proxy response", () => {
  it("normalizes GoAPI R2 upload metadata to a frontend image URL", () => {
    const payload = normalizeImageUploadPayload({
      status: "success",
      code: 200,
      data: {
        holdingcode: "SHOP001",
        category: "system-settings/imageuri",
        file_name: "20260525_111111_abcd.webp",
      },
    });

    expect(payload.success).toBe(true);
    expect(payload.url).toBe(
      "/goapi/s3/file/SHOP001/system-settings/imageuri/20260525_111111_abcd.webp",
    );
    expect(payload.data).toMatchObject({
      url: "/goapi/s3/file/SHOP001/system-settings/imageuri/20260525_111111_abcd.webp",
      key: "SHOP001/system-settings/imageuri/20260525_111111_abcd.webp",
    });
  });

  it("keeps upstream errors as failed envelopes", () => {
    const payload = normalizeImageUploadPayload(
      { status: "error", message: "R2 storage is not configured" },
      false,
    );

    expect(payload.success).toBe(false);
    expect(payload.message).toBe("R2 storage is not configured");
  });

  it("normalizes the private URL returned by the video uploader", () => {
    const payload = normalizeImageUploadPayload({
      success: true,
      fileurl: "/s3/file/SHOP001/companies/COMPANY-A/products/videos/demo.mp4",
      objectkey: "SHOP001/companies/COMPANY-A/products/videos/demo.mp4",
    });

    expect(payload).toMatchObject({
      success: true,
      url: "/goapi/s3/file/SHOP001/companies/COMPANY-A/products/videos/demo.mp4",
      data: {
        key: "SHOP001/companies/COMPANY-A/products/videos/demo.mp4",
        url: "/goapi/s3/file/SHOP001/companies/COMPANY-A/products/videos/demo.mp4",
      },
    });
  });

  it("preserves an encoded Company segment returned by the image uploader", () => {
    const payload = normalizeImageUploadPayload({
      status: "success",
      data: {
        category: "companies/~VUFUQg/products/images",
        file_name: "product.png",
        holdingcode: "test",
      },
    });

    expect(payload.url).toBe(
      "/goapi/s3/file/test/companies/~VUFUQg/products/images/product.png",
    );
  });

  it("maps stored S3 URIs to the same-origin file BFF only", () => {
    expect(fileBffUrl("/goapi/s3/file/SHOP/imageuri/a.webp")).toBe("/api/files/SHOP/imageuri/a.webp");
    expect(fileBffUrl(" /goapi/s3/file/SHOP/a.png?variant=thumbnail ")).toBe("/api/files/SHOP/a.png?variant=thumbnail");
    for (const other of ["", "/s3/file/SHOP/a.png", "/backend/goapi/s3/file/SHOP/a.png", "https://x.example/goapi/s3/file/a.png", "/api/x"]) {
      expect(fileBffUrl(other), other).toBe("");
    }
  });

  it("sends the Bearer token only to same-origin /api/files paths", () => {
    expect(imageNeedsAuthenticatedFetch("/api/files/SHOP/imageuri/a.webp")).toBe(true);
    expect(imageNeedsAuthenticatedFetch("http://localhost/api/files/SHOP/imageuri/a.webp")).toBe(true);
    for (const url of [
      // Stored form and the removed public proxy are never fetched directly any more.
      "/goapi/s3/file/SHOP/imageuri/a.webp",
      "/s3/file/SHOP/imageuri/a.webp",
      "http://localhost:3000/backend/goapi/s3/file/SHOP/imageuri/a.webp",
      "http://localhost:8888/goapi/s3/file/SHOP/imageuri/a.webp",
      // Other hosts never get the token, whatever the path.
      "https://evil.example/api/files/SHOP/imageuri/a.webp",
      "//evil.example/api/files/SHOP/imageuri/a.webp",
      String.raw`\\evil.example\api\files\SHOP\imageuri\a.webp`,
      "https:/evil.example/api/files/SHOP/imageuri/a.webp",
    ]) {
      expect(imageNeedsAuthenticatedFetch(url), url).toBe(false);
    }
  });

  it("ignores public, blob, data, and unrelated paths", () => {
    expect(imageNeedsAuthenticatedFetch("")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("blob:http://localhost/abc")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("data:image/png;base64,xxx")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("/flags/th.png")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("https://cdn.example.com/logo.png")).toBe(false);
  });

  it("adds the fixed WebP thumbnail variant without changing the object path", () => {
    expect(imageThumbnailProxyUrl("/api/files/SHOP/images/a.png")).toBe(
      "/api/files/SHOP/images/a.png?variant=thumbnail",
    );
    expect(imageThumbnailProxyUrl("/api/files/SHOP/images/a.png?download=0#image")).toBe(
      "/api/files/SHOP/images/a.png?download=0&variant=thumbnail#image",
    );
    expect(imageThumbnailProxyUrl("/api/files/SHOP/images/a.png?variant=original")).toBe(
      "/api/files/SHOP/images/a.png?variant=thumbnail",
    );
    expect(imageThumbnailProxyUrl("/goapi/s3/file/SHOP/images/a.png")).toBe("/goapi/s3/file/SHOP/images/a.png");
    expect(imageThumbnailProxyUrl("https://cdn.example.com/a.png")).toBe(
      "https://cdn.example.com/a.png",
    );
  });

  it("returns an explicit error when a required upload category is missing", async () => {
    const form = new FormData();
    form.append("file", new File(["image"], "branch.webp", { type: "image/webp" }));

    const response = await proxyImageUploadToGoApi(
      new Request("http://localhost/api/upload/image", {
        method: "POST",
        headers: {
          authorization: "Bearer test-token",
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
        body: form,
      }),
      { category: "system-settings", requireClientCategory: true },
    );

    expect(response.status).toBe(400);
    await expect(response.json()).resolves.toMatchObject({
      success: false,
      message: "ไม่พบหมวดหมู่รูปภาพสำหรับอัปโหลด",
    });
  });

  it("allows a streamed video larger than 50 MB within the configured limit", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ success: true }), {
        headers: { "content-type": "application/json" },
      }),
    );

    try {
      const response = await proxyImageUploadToGoApi(
        new Request("http://localhost/api/product-barcode/video", {
          method: "POST",
          headers: {
            authorization: "Bearer test-token",
            "content-length": String(51 * 1024 * 1024),
            "content-type": "multipart/form-data; boundary=video-test",
            "x-bc-backend-url": "http://localhost:8888/goapi",
          },
          body: new Uint8Array([1]),
        }),
        {
          backendPath: "/goapi/video/upload",
          category: "products/videos",
          forwardRequestBody: true,
          kind: "video",
          maxRequestBytes: PRODUCT_VIDEO_REQUEST_MAX_BYTES,
        },
      );

      expect(response.status).toBe(200);
      expect(fetchMock).toHaveBeenCalledOnce();
    } finally {
      fetchMock.mockRestore();
    }
  });

  it("rejects a streamed video above the configured limit before forwarding it", async () => {
    const response = await proxyImageUploadToGoApi(
      new Request("http://localhost/api/product-barcode/video", {
        method: "POST",
        headers: {
          authorization: "Bearer test-token",
          "content-length": String(PRODUCT_VIDEO_REQUEST_MAX_BYTES + 1),
          "content-type": "multipart/form-data; boundary=video-test",
          "x-bc-backend-url": "http://localhost:8888/goapi",
        },
        body: new Uint8Array([1]),
      }),
      {
        backendPath: "/goapi/video/upload",
        category: "products/videos",
        forwardRequestBody: true,
        kind: "video",
        maxRequestBytes: PRODUCT_VIDEO_REQUEST_MAX_BYTES,
      },
    );

    expect(response.status).toBe(413);
  });

  it("stops a chunked video stream that exceeds the limit", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (_input, init) => {
      const reader = (init?.body as ReadableStream<Uint8Array>).getReader();
      while (!(await reader.read()).done) {
        // Consume the forwarded stream like the backend connection.
      }
      return new Response(JSON.stringify({ success: true }), {
        headers: { "content-type": "application/json" },
      });
    });

    try {
      const response = await proxyImageUploadToGoApi(
        new Request("http://localhost/api/product-barcode/video", {
          method: "POST",
          headers: {
            authorization: "Bearer test-token",
            "content-type": "multipart/form-data; boundary=video-test",
            "x-bc-backend-url": "http://localhost:8888/goapi",
          },
          body: new Uint8Array([1, 2, 3]),
        }),
        {
          backendPath: "/goapi/video/upload",
          category: "products/videos",
          forwardRequestBody: true,
          kind: "video",
          maxRequestBytes: 2,
        },
      );

      expect(response.status).toBe(413);
      expect(fetchMock).toHaveBeenCalledOnce();
    } finally {
      fetchMock.mockRestore();
    }
  });
});

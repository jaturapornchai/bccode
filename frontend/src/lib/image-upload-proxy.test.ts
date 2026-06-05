import { describe, expect, it } from "vitest";
import {
  imageNeedsAuthenticatedFetch,
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

  it("detects private GoAPI image paths across all routing surfaces", () => {
    expect(imageNeedsAuthenticatedFetch("/goapi/s3/file/SHOP/imageuri/a.webp")).toBe(true);
    expect(imageNeedsAuthenticatedFetch("/s3/file/SHOP/imageuri/a.webp")).toBe(true);
    expect(
      imageNeedsAuthenticatedFetch("http://localhost:8888/goapi/s3/file/SHOP/imageuri/a.webp"),
    ).toBe(true);
    expect(
      imageNeedsAuthenticatedFetch("http://localhost:3000/backend/goapi/s3/file/SHOP/imageuri/a.webp"),
    ).toBe(true);
    expect(
      imageNeedsAuthenticatedFetch("https://dev.bcaicloud.com/backend/goapi/s3/file/SHOP/imageuri/a.webp"),
    ).toBe(true);
  });

  it("ignores public, blob, data, and unrelated paths", () => {
    expect(imageNeedsAuthenticatedFetch("")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("blob:http://localhost/abc")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("data:image/png;base64,xxx")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("/flags/th.png")).toBe(false);
    expect(imageNeedsAuthenticatedFetch("https://cdn.example.com/logo.png")).toBe(false);
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
});

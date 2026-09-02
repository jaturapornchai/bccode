import { describe, expect, it } from "vitest";
import { getAuthBridgeUrl, normalizeAuthBridgeUrl } from "./auth-bridge";

describe("auth bridge config", () => {
  it("normalizes a bridge base URL", () => {
    expect(normalizeAuthBridgeUrl("https://dev-api.bcaicloud.com/liff/?x=1#top")).toBe("https://dev-api.bcaicloud.com/liff");
  });

  it("rejects credentials in bridge URL", () => {
    expect(() => normalizeAuthBridgeUrl("https://user:pass@dev-api.bcaicloud.com/liff")).toThrow("ห้ามมี username");
  });

  it("fails closed when the bridge is not configured", () => {
    expect(() => getAuthBridgeUrl({})).toThrow("ยังไม่ได้ตั้งค่า Auth bridge");
  });
});

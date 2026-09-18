import { describe, it, expect } from "vitest";
import { thaiBahtText } from "./thai-baht-text";

describe("thaiBahtText", () => {
  it("converts 0 to ศูนย์บาทถ้วน", () => {
    expect(thaiBahtText(0)).toBe("ศูนย์บาทถ้วน");
    expect(thaiBahtText("0")).toBe("ศูนย์บาทถ้วน");
    expect(thaiBahtText("")).toBe("ศูนย์บาทถ้วน");
    expect(thaiBahtText(null)).toBe("ศูนย์บาทถ้วน");
  });

  it("converts exact integers correctly", () => {
    expect(thaiBahtText(1)).toBe("หนึ่งบาทถ้วน");
    expect(thaiBahtText(10)).toBe("สิบบาทถ้วน");
    expect(thaiBahtText(11)).toBe("สิบเอ็ดบาทถ้วน");
    expect(thaiBahtText(20)).toBe("ยี่สิบบาทถ้วน");
    expect(thaiBahtText(21)).toBe("ยี่สิบเอ็ดบาทถ้วน");
    expect(thaiBahtText(100)).toBe("หนึ่งร้อยบาทถ้วน");
    expect(thaiBahtText(101)).toBe("หนึ่งร้อยเอ็ดบาทถ้วน");
    expect(thaiBahtText(1000)).toBe("หนึ่งพันบาทถ้วน");
    expect(thaiBahtText(1500)).toBe("หนึ่งพันห้าร้อยบาทถ้วน");
    expect(thaiBahtText(10000)).toBe("หนึ่งหมื่นบาทถ้วน");
    expect(thaiBahtText(100000)).toBe("หนึ่งแสนบาทถ้วน");
    expect(thaiBahtText(1000000)).toBe("หนึ่งล้านบาทถ้วน");
    expect(thaiBahtText(10000000)).toBe("สิบล้านบาทถ้วน");
  });

  it("converts satang / decimals correctly", () => {
    expect(thaiBahtText(0.5)).toBe("ห้าสิบสตางค์");
    expect(thaiBahtText(0.25)).toBe("ยี่สิบห้าสตางค์");
    expect(thaiBahtText(0.05)).toBe("ห้าสตางค์");
    expect(thaiBahtText(0.01)).toBe("หนึ่งสตางค์");
    expect(thaiBahtText(10.5)).toBe("สิบบาทห้าสิบสตางค์");
    expect(thaiBahtText(1234.5)).toBe("หนึ่งพันสองร้อยสามสิบสี่บาทห้าสิบสตางค์");
    expect(thaiBahtText(1234.75)).toBe("หนึ่งพันสองร้อยสามสิบสี่บาทเจ็ดสิบห้าสตางค์");
  });

  it("handles negative numbers and strings with commas", () => {
    expect(thaiBahtText(-500)).toBe("ลบห้าร้อยบาทถ้วน");
    expect(thaiBahtText("1,250,500.00")).toBe("หนึ่งล้านสองแสนห้าหมื่นห้าร้อยบาทถ้วน");
  });
});

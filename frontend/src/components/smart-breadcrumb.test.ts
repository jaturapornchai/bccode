import { describe, expect, it } from "vitest";
import { getBreadcrumbTrail } from "./smart-breadcrumb";

describe("smart-breadcrumb", () => {
  it("returns null for root or empty path", () => {
    expect(getBreadcrumbTrail("")).toBeNull();
    expect(getBreadcrumbTrail("/")).toBeNull();
  });

  it("finds GL journal screen in breadcrumb hierarchy", () => {
    const trail = getBreadcrumbTrail("/gl/journals", "th");
    expect(trail).not.toBeNull();
    expect(trail?.item).toBe("สมุดรายวัน");
    expect(trail?.route).toBe("/gl/journals");
    expect(trail?.group).toBe("สมุดรายวันและงานประจำ");
  });

  it("finds english labels when language is en", () => {
    const trail = getBreadcrumbTrail("/gl/journals", "en");
    expect(trail).not.toBeNull();
    expect(trail?.item).toBe("Journals");
  });

  it("matches prefix routes for dynamic subpaths", () => {
    const trail = getBreadcrumbTrail("/gl/journals/edit/123", "th");
    expect(trail).not.toBeNull();
    expect(trail?.item).toBe("สมุดรายวัน");
  });
});

import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemMainMapScreen } from "./system-main-map-screen";

describe("SystemMainMapScreen (Champ Parity S2 / Main Logistics Map Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemMainMapScreen).toBe("function");
  });

  it("has active route for system-main-map in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-main-map");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/main-map");
    expect(item?.label.th).toContain("กำหนดแผนที่หลัก");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/main-map")).toBe(false);
  });
});

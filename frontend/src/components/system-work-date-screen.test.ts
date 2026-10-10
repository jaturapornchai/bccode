import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemWorkDateScreen } from "./system-work-date-screen";

describe("SystemWorkDateScreen (Champ Parity S2 / Default Work Date Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemWorkDateScreen).toBe("function");
  });

  it("has active route for system-work-date in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-work-date");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/work-date");
    expect(item?.label.th).toContain("กำหนดวันที่ทำงานเริ่มต้น");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/work-date")).toBe(false);
  });
});

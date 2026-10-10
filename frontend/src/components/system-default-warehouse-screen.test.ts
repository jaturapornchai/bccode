import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemDefaultWarehouseScreen } from "./system-default-warehouse-screen";

describe("SystemDefaultWarehouseScreen (Champ Parity S2 / Default Warehouse Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemDefaultWarehouseScreen).toBe("function");
  });

  it("has active route for system-default-warehouse in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-default-warehouse");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/default-warehouse");
    expect(item?.label.th).toContain("กำหนดคลังที่เก็บสินค้าเริ่มต้น");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/default-warehouse")).toBe(false);
  });
});

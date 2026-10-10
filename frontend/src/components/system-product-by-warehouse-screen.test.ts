import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemProductByWarehouseScreen } from "./system-product-by-warehouse-screen";

describe("SystemProductByWarehouseScreen (Champ Parity S2 / Product by Warehouse Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemProductByWarehouseScreen).toBe("function");
  });

  it("has active route for system-product-by-warehouse in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-product-by-warehouse");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/product-by-warehouse");
    expect(item?.label.th).toContain("กำหนดสินค้าตามคลัง");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/product-by-warehouse")).toBe(false);
  });
});

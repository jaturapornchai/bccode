import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { ProductSupplierUpdateScreen } from "./product-supplier-update-screen";

describe("ProductSupplierUpdateScreen (Champ Parity 50000+ / Product Supplier Update)", () => {
  it("exports component function", () => {
    expect(typeof ProductSupplierUpdateScreen).toBe("function");
  });

  it("has active route for product-supplier-update in menu data hierarchy under ic-pricing", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "product-supplier-update");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/inventory/product-supplier-update");
    expect(item?.label.th).toContain("ปรับปรุง ข้อมูล Supplier ตามสินค้า");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/inventory/product-supplier-update")).toBe(false);
  });
});

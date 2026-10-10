import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SaleDispatchScreen } from "./sale-dispatch-screen";

describe("SaleDispatchScreen (Champ Parity 30005 / Goods Dispatch Order)", () => {
  it("exports component function", () => {
    expect(typeof SaleDispatchScreen).toBe("function");
  });

  it("has active route for sale-dispatch in menu data hierarchy under sales-orders", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "sale-dispatch");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/sale/dispatch");
    expect(item?.label.th).toContain("ใบสั่งจ่ายสินค้า");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/sale/dispatch")).toBe(false);
  });
});

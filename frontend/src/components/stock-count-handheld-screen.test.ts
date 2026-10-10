import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { StockCountHandheldScreen } from "./stock-count-handheld-screen";

describe("StockCountHandheldScreen (Champ Parity 50000+ / Handheld Stock Count)", () => {
  it("exports component function", () => {
    expect(typeof StockCountHandheldScreen).toBe("function");
  });

  it("has active route for stock-count-handheld in menu data hierarchy under ic-count", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "stock-count-handheld");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/inventory/stock-count-handheld");
    expect(item?.label.th).toContain("ตรวจนับสต็อกโดย Hand Held");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/inventory/stock-count-handheld")).toBe(false);
  });
});

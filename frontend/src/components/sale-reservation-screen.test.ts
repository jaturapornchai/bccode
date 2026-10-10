import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SaleReservationScreen } from "./sale-reservation-screen";

describe("SaleReservationScreen (Champ Parity 30005 / Sales Reservation)", () => {
  it("exports component function", () => {
    expect(typeof SaleReservationScreen).toBe("function");
  });

  it("has active route for sale-reservation in menu data hierarchy under sales-orders", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "sale-reservation");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/sale/reservation");
    expect(item?.label.th).toContain("ใบสั่งจองสินค้า");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/sale/reservation")).toBe(false);
  });
});

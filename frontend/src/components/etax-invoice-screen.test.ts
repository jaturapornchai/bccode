import { describe, expect, it } from "vitest";
import { CUSTOM_MENU_SCREEN_ROUTES, isMenuScreenPending } from "@/lib/menu-screen-status";

describe("ETaxInvoiceScreen (/transaction/etax)", () => {
  it("includes /transaction/etax in CUSTOM_MENU_SCREEN_ROUTES", () => {
    expect(CUSTOM_MENU_SCREEN_ROUTES).toContain("/transaction/etax");
  });

  it("marks /transaction/etax as a connected screen (not pending)", () => {
    expect(isMenuScreenPending("/transaction/etax")).toBe(false);
  });
});

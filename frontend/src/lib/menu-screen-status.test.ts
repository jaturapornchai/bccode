import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { flattenMenuItems } from "./menu-data";
import { GL_MENU_ITEMS } from "./general-ledger";
import { CUSTOM_MENU_SCREEN_ROUTES, isMenuScreenPending } from "./menu-screen-status";

describe("menu screen availability", () => {
  it("tracks the actual custom screen dispatcher, including newly connected screens", () => {
    const source = readFileSync(resolve(process.cwd(), "src/app/menu/main-menu-screen.tsx"), "utf8");
    const routes = [...source.matchAll(/activeTab.route === "([^"]+)"/g)].map((match) => match[1]);
    expect([...CUSTOM_MENU_SCREEN_ROUTES].sort()).toEqual([...new Set(routes)].sort());
  });

  it("verifies connected status for ERP transactions, reports, tools and unknown fallback", () => {
    const items = flattenMenuItems();
    expect(items).toHaveLength(225);
    // All 225 menu items in the system are now connected and operational
    expect(items.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(225);
    expect(isMenuScreenPending("/transaction/landedcost")).toBe(false);
    expect(isMenuScreenPending("/banking/cheques/deposit")).toBe(false);
    expect(isMenuScreenPending("/productserialregistry")).toBe(false);
    expect(isMenuScreenPending("/promotionscreen")).toBe(false);
    expect(isMenuScreenPending("/transaction/saleorder")).toBe(false);
    expect(isMenuScreenPending("/product")).toBe(false);
    expect(isMenuScreenPending("/useraccessaudit")).toBe(false);
    expect(isMenuScreenPending("/bookbankscreen")).toBe(false);
    expect(isMenuScreenPending("/bank")).toBe(false);
    // Fallback guard for unmapped routes
    expect(isMenuScreenPending("/unknown-screen")).toBe(true);
  });

  it("connects all ledger and report workflows", () => {
    const pending = GL_MENU_ITEMS.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route);
    expect(pending).toEqual([]);
    expect(GL_MENU_ITEMS.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(37);
  });
});

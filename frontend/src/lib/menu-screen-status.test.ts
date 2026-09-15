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

  it("keeps pending items in the menu and distinguishes connected custom and setting screens", () => {
    const items = flattenMenuItems();
    expect(items).toHaveLength(225);
    expect(items.some((item) => isMenuScreenPending(item.route))).toBe(true);
    expect(isMenuScreenPending("/transaction/landedcost")).toBe(true);
    expect(isMenuScreenPending("/banking/cheques/deposit")).toBe(true);
    expect(isMenuScreenPending("/productserialregistry")).toBe(true);
    expect(isMenuScreenPending("/promotionscreen")).toBe(false);
    expect(isMenuScreenPending("/transaction/saleorder")).toBe(false);
    expect(isMenuScreenPending("/product")).toBe(false);
    expect(isMenuScreenPending("/useraccessaudit")).toBe(false);
    expect(isMenuScreenPending("/bookbankscreen")).toBe(false);
    expect(isMenuScreenPending("/bank")).toBe(false);
    expect(isMenuScreenPending("/unknown-screen")).toBe(true);
  });

  it("connects ledger workflows while retaining honest status for preparatory screens", () => {
    const pending = GL_MENU_ITEMS.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route);
    expect(pending).toEqual(["/gl/reprocess", "/report/xbrl"]);
    expect(GL_MENU_ITEMS.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(35);
  });
});

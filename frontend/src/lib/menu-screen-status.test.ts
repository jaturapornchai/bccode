import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { flattenMenuItems } from "./menu-data";
import { GL_MENU_ITEMS } from "./general-ledger";
import { CUSTOM_MENU_SCREEN_ROUTES, isMenuScreenPending, isMenuDataPending } from "./menu-screen-status";

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

describe("Menu data readiness", () => {
  it("flags screens that render but have no live API behind them", () => {
    // รายงาน: มี API จริง 4 ตัว
    expect(isMenuDataPending("/report/salesreportbydocument")).toBe(false);
    expect(isMenuDataPending("/report/stockbalanceitem")).toBe(false);
    expect(isMenuDataPending("/report/araging")).toBe(true);
    expect(isMenuDataPending("/report/xbrl")).toBe(true);

    // เครื่องมือ: มี API จริง 2 ตัว
    expect(isMenuDataPending("/rebuildproductbalancescreen")).toBe(false);
    expect(isMenuDataPending("/auditscreen")).toBe(false);
    expect(isMenuDataPending("/gl/reprocess")).toBe(true);

    // งานอนุมัติ: มีเฉพาะใบขอซื้อ
    expect(isMenuDataPending("/procurement/requisition-approval")).toBe(false);
    expect(isMenuDataPending("/sales/quotation-approval")).toBe(true);

    // ภาษี: ภาษีขาย/ซื้อ/ภ.พ.30 ต่อ API แล้ว ส่วน ภ.ง.ด. ยังไม่มีข้อมูลต้นทาง
    expect(isMenuDataPending("/report/vatsale")).toBe(false);
    expect(isMenuDataPending("/report/vatpnd3")).toBe(true);

    // จอที่ไม่ได้อยู่ในกลุ่มเหล่านี้ ไม่ถือว่ารอข้อมูล
    expect(isMenuDataPending("/product")).toBe(false);
  });
});

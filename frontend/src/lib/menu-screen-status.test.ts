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

  // Champ parity 2026-09-19: routes added from Champ menuconfig.xml that have no screen or backend yet.
  // AP/AR debt reports (8 routes) connected to PostgreSQL via /api/report/debt/query on 2026-09-19.
  const CHAMP_PENDING_ROUTES = [
    "/transaction/purchasereducedebt", "/transaction/rfqpricetable",
    "/report/chequereceived", "/report/chequeissued", "/report/creditcard", "/report/bankstatement",
    "/report/cashmovement", "/report/pettycashmovement", "/report/monthlypaymentbook",
    "/report/maxstock", "/report/nomovementstock", "/report/stockcountvariance",
    "/report/pendingreceive", "/report/pendingdelivery", "/report/serialmovement",
    "/report/depreciationmonthly", "/report/depreciationyearly", "/report/depreciationpnd50", "/report/assetdisposal",
    "/report/vatsummary",
  ];

  it("verifies connected status for ERP transactions, reports, tools and unknown fallback", () => {
    const items = flattenMenuItems();
    expect(items).toHaveLength(194);
    expect(items.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route).sort()).toEqual([...CHAMP_PENDING_ROUTES].sort());
    expect(items.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(194 - CHAMP_PENDING_ROUTES.length);
    expect(isMenuScreenPending("/gl/fiscal-years")).toBe(false);
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

  it("connects every ledger workflow (all 23 Champ items are fully implemented)", () => {
    const pending = GL_MENU_ITEMS.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route);
    expect(pending).toEqual([]);
    expect(GL_MENU_ITEMS.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(23);
    expect(isMenuScreenPending("/gl/journal-books")).toBe(false);
    expect(isMenuScreenPending("/report/gljournal")).toBe(false);
    expect(isMenuScreenPending("/report/budgetcomparison")).toBe(false);
    expect(isMenuScreenPending("/gl/account-groups")).toBe(false);
    expect(isMenuScreenPending("/gl/reprocess")).toBe(false);
  });
});

describe("Menu data readiness", () => {
  it("flags screens that render but have no live API behind them", () => {
    // รายงาน: มี API จริง 12 ตัว (4 รายงานสินค้า/ขาย + 8 รายงานเจ้าหนี้/ลูกหนี้)
    expect(isMenuDataPending("/report/salesreportbydocument")).toBe(false);
    expect(isMenuDataPending("/report/stockbalanceitem")).toBe(false);
    expect(isMenuDataPending("/report/apmovement")).toBe(false);
    expect(isMenuDataPending("/report/apstatus")).toBe(false);
    expect(isMenuDataPending("/report/armovement")).toBe(false);
    expect(isMenuDataPending("/report/arstatus")).toBe(false);
    expect(isMenuDataPending("/report/araging")).toBe(true);
    expect(isMenuDataPending("/report/xbrl")).toBe(true);

    // เครื่องมือ: ยังไม่มีเครื่องมือใดต่อ API (เครื่องมือตรวจ/สร้างยอดสินค้าใหม่ถูกตัดตาม Champ 2026-09-19)
    expect(isMenuDataPending("/tools/ar-recalculate")).toBe(true);
    expect(isMenuDataPending("/gl/reprocess")).toBe(false);

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

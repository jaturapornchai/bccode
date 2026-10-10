import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { flattenMenuItems } from "./menu-data";
import { GL_MENU_ITEMS } from "./general-ledger";
import { CUSTOM_MENU_SCREEN_ROUTES, isMenuBackendRetired, isMenuScreenPending, isMenuDataPending, isRetiredBackendScreen } from "./menu-screen-status";

describe("menu screen availability", () => {
  it("tracks the actual custom screen dispatcher, including newly connected screens", () => {
    const source = readFileSync(resolve(process.cwd(), "src/app/menu/main-menu-screen.tsx"), "utf8");
    const routes = [...source.matchAll(/activeTab.route === "([^"]+)"/g)].map((match) => match[1]);
    expect([...CUSTOM_MENU_SCREEN_ROUTES].sort()).toEqual([...new Set(routes)].sort());
  });

  // Champ parity 2026-09-19: all Champ pending routes are now fully connected!
  const CHAMP_PENDING_ROUTES: string[] = [];

  // 2026-10-09 (S1): cancel/approve screens Champ has; they render the operations workbench but no approval API exists (pending like the first six).
  const CHAMP_OPERATIONS_ROUTES = [
    "/procurement/requisition-cancellation",
    "/procurement/requisition-approval-cancellation",
    "/procurement/order-approval",
    "/inventory/issue-request-cancellation",
    "/inventory/transfer-request-cancellation",
  ];

  // 2026-10-08: เมนูที่ backend ไม่มีประเภทเอกสารรองรับ (ไม่มี config ใน erp-transaction.ts) → pending
  // 2026-10-10: /transaction/withholdingtax เชื่อมกับ /report/whtcertificate แล้ว; /banking/contacts เชื่อมกับ BankingContactsScreen แล้ว; /banking/petty-cash-limit เชื่อมกับ PettyCashLimitScreen แล้ว
  const NO_BACKEND_DOCUMENT_ROUTES = [
    "/transaction/etax",
  ];

  it("verifies connected status for ERP transactions, reports, tools and unknown fallback", () => {
    const items = flattenMenuItems();
    expect(items).toHaveLength(238);
    // 2026-09-23: จอที่เคยพึ่ง API บน MongoDB (ถอดแล้ว) เป็น "ยังไม่พร้อม" จนกว่าจะมี API บน PostgreSQL
    const retired = items.filter((item) => isMenuBackendRetired(item.route)).map((item) => item.route);
    const expected = [...new Set([...CHAMP_PENDING_ROUTES, ...NO_BACKEND_DOCUMENT_ROUTES, ...retired])].sort();
    expect(items.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route).sort()).toEqual(expected);
    expect(items.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(items.length - expected.length);
    expect(isMenuScreenPending("/gl/fiscal-years")).toBe(false);
    // เอกสารธุรกรรมที่ยังมี config ทำงานจริงกับตาราง doc (ไม่ใช่ retired / pending)
    expect(isMenuBackendRetired("/transaction/landedcost")).toBe(false);
    expect(isMenuBackendRetired("/banking/cheques/deposit")).toBe(false);
    for (const route of ["/transaction/sale", "/transaction/paid", "/banking/cash-deposit", "/transaction/landedcost"]) {
      expect(isMenuScreenPending(route), route).toBe(false);
    }
    // เมนูที่ไม่มีประเภทเอกสารใน backend เป็น pending แต่ยังอยู่ในเมนู; จอสินค้า 3 จอไม่มี API
    for (const route of [...NO_BACKEND_DOCUMENT_ROUTES, "/product", "/productbarcode", "/productbarcodeshelf"]) {
      expect(isMenuScreenPending(route), route).toBe(true);
    }
    expect(flattenMenuItems().map((item) => item.route)).toEqual(expect.arrayContaining(NO_BACKEND_DOCUMENT_ROUTES));
    // 2026-10-10: 11 Tier A approval/cancellation screens now connected to PostgreSQL backend
    for (const route of CHAMP_OPERATIONS_ROUTES) {
      expect(isMenuScreenPending(route), route).toBe(false);
      expect(isMenuDataPending(route), route).toBe(false);
    }
    expect(isRetiredBackendScreen("/product")).toBe(true);
    expect(isRetiredBackendScreen("/productbarcodeshelf?x=1")).toBe(true);
    expect(isRetiredBackendScreen("/transaction/sale")).toBe(false);
    expect(isMenuBackendRetired("/productserialregistry")).toBe(true);
    expect(isMenuBackendRetired("/promotionscreen")).toBe(true);
    expect(isMenuBackendRetired("/transaction/saleorder")).toBe(false);
    expect(isMenuBackendRetired("/product")).toBe(true);
    expect(isMenuBackendRetired("/productwarehousescreen")).toBe(true);
    expect(isMenuScreenPending("/useraccessaudit")).toBe(false);
    expect(isMenuBackendRetired("/organization/branch")).toBe(false);
    expect(isMenuBackendRetired("/gl/fiscal-years")).toBe(false);
    expect(isMenuBackendRetired("/bookbankscreen")).toBe(true);
    expect(isMenuBackendRetired("/bank")).toBe(true);
    // Fallback guard for unmapped routes
    expect(isMenuScreenPending("/unknown-screen")).toBe(true);
  });

  it("connects every ledger workflow including the monthly budget entry screen (2026-09-25)", () => {
    const pending = GL_MENU_ITEMS.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route);
    expect(pending).toEqual([]);
    expect(GL_MENU_ITEMS.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(29);
    expect(isMenuScreenPending("/gl/unposting")).toBe(false);
    expect(isMenuScreenPending("/gl/budget")).toBe(false);
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

    // งานอนุมัติ: 11 จอ Tier A ต่อ API บน PostgreSQL แล้ว (isApprovalApiReady = true)
    expect(isMenuDataPending("/procurement/requisition-approval")).toBe(false);
    expect(isMenuDataPending("/sales/quotation-approval")).toBe(false);

    // ภาษี (2026-09-23): รายงานภาษีหัก ณ ที่จ่าย/50 ทวิ และภาษีขาย/ซื้อ อ่านจาก GL จริง (details.vats);
    // แบบยื่นทุกแบบเป็นจอกรอกแบบ (TaxFormEditor) — แบบที่ยังไม่มีข้อมูลต้นทาง (ภ.พ.36) กรอกเองได้ จึงไม่ "รอเชื่อมข้อมูล"
    expect(isMenuDataPending("/report/vatpnd3")).toBe(false);
    expect(isMenuDataPending("/report/whtcertificate")).toBe(false);
    expect(isMenuDataPending("/report/reportvatsale")).toBe(false);
    expect(isMenuDataPending("/report/reportvatbuy")).toBe(false);
    expect(isMenuDataPending("/report/vatpp30")).toBe(false);
    expect(isMenuDataPending("/report/vatpp36")).toBe(false);
    expect(isMenuScreenPending("/report/vatpp36")).toBe(false);
    expect(isMenuScreenPending("/report/taxforms")).toBe(false);

    // จอที่ไม่ได้อยู่ในกลุ่มเหล่านี้ ไม่ถือว่ารอข้อมูล
    expect(isMenuDataPending("/product")).toBe(false);
  });
});

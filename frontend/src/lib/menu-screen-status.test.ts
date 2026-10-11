import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { BANK_CODE_KINDS, BANK_CODE_ROUTE_PREFIX } from "./bank-code-kinds";
import { CHEQUE_SCREEN_ROUTES } from "./cheque-doc-config";
import { flattenMenuItems } from "./menu-data";
import { GL_MENU_ITEMS } from "./general-ledger";
import { getErpModuleConfig } from "./erp-transaction";
import { CUSTOM_MENU_SCREEN_ROUTES, isMenuBackendRetired, isMenuScreenPending, isMenuDataPending, isRetiredBackendScreen } from "./menu-screen-status";

describe("menu screen availability", () => {
  it("tracks the actual custom screen dispatcher, including newly connected screens", () => {
    const source = readFileSync(resolve(process.cwd(), "src/app/menu/main-menu-screen.tsx"), "utf8");
    const routes = [...source.matchAll(/activeTab.route === "([^"]+)"/g)].map((match) => match[1]);
    // จอกำหนดรหัสธนาคารเลือกชนิดจาก route ด้วย bankCodeKindFromRoute() ตัวเดียว → ขยายเป็น route ของทุกชนิด
    if (source.includes("bankCodeKindFromRoute(activeTab.route)")) routes.push(...BANK_CODE_KINDS.map((kind) => BANK_CODE_ROUTE_PREFIX + kind));
    // จอเช็ค 10 จอเลือกจาก route ด้วย chequeScreenFromRoute() ตัวเดียว → ขยายเป็น route ของทั้ง 10 จอ
    if (source.includes("chequeScreenFromRoute(activeTab.route)")) routes.push(...CHEQUE_SCREEN_ROUTES);
    expect([...CUSTOM_MENU_SCREEN_ROUTES].sort()).toEqual([...new Set(routes)].sort());
  });

  // 2026-10-11: จอ custom ที่ branch อยู่ก่อน isErpTransactionRoute ใน WorkTabPanel จะบังจอ ERP จริงของ route เดียวกัน
  // (จอเงา deposit-receipt เคยบัง /transaction/receivedeposit) — route ใน CUSTOM ต้องไม่มี config ERP
  it("never lets a custom screen shadow an ERP transaction route", () => {
    expect(CUSTOM_MENU_SCREEN_ROUTES.filter((route) => getErpModuleConfig(route))).toEqual([]);
  });

  // 2026-10-11: จอ UI-only (ข้อมูลจำลอง/localStorage ไม่มี API+PostgreSQL จริง) ถอดเป็น "รอพัฒนา" จนกว่าจะสร้างจริง
  // (docs/kms/champ/design-2026-10-11/facade-screens.md). ลบ route ออกจากรายการนี้ใน commit เดียวกับที่สร้างจอจริง
  const CHAMP_PENDING_ROUTES: string[] = [
    "/ar/billing-note-auto",
    "/banking/bank-service",
    "/banking/cheque-issued-prepare",
    "/banking/cheque-movement",
    "/banking/petty-cash-limit",
    "/inventory/product-supplier-update",
    "/inventory/stock-count-handheld",
    "/sale/dispatch",
    "/sale/reservation",
    "/settings/about",
    "/settings/credit-debit-reasons",
    "/settings/currency",
    "/settings/customer-inquiry",
    "/settings/default-warehouse",
    "/settings/font-report",
    "/settings/font-report-column",
    "/settings/font-report-company",
    "/settings/font-report-detail",
    "/settings/font-report-title",
    "/settings/journal-format",
    "/settings/main-map",
    "/settings/payment-terms",
    "/settings/product-by-warehouse",
    "/settings/product-inquiry",
    "/settings/standard-software",
    "/settings/work-date",
    "/tools/transaction-lock",
    "/transaction/etax",
  ];

  // 2026-10-09 (S1): cancel/approve screens Champ has; they render the operations workbench (connected to the approval API since 2026-10-10).
  const CHAMP_OPERATIONS_ROUTES = [
    "/procurement/requisition-cancellation",
    "/procurement/requisition-approval-cancellation",
    "/procurement/order-approval",
    "/inventory/issue-request-cancellation",
    "/inventory/transfer-request-cancellation",
  ];

  // 2026-10-08: เมนูที่ backend ไม่มีประเภทเอกสารรองรับ (ไม่มี config ใน erp-transaction.ts) → pending
  // 2026-10-11: /transaction/etax กลับเป็น pending (คิวเอกสารเป็นข้อมูลจำลอง) — อยู่ใน CHAMP_PENDING_ROUTES ด้านบน
  const NO_BACKEND_DOCUMENT_ROUTES: string[] = [];

  it("verifies connected status for ERP transactions, reports, tools and unknown fallback", () => {
    const items = flattenMenuItems();
    expect(items).toHaveLength(255);
    // 2026-09-23: จอที่เคยพึ่ง API บน MongoDB (ถอดแล้ว) เป็น "ยังไม่พร้อม" จนกว่าจะมี API บน PostgreSQL
    const retired = items.filter((item) => isMenuBackendRetired(item.route)).map((item) => item.route);
    const expected = [...new Set([...CHAMP_PENDING_ROUTES, ...NO_BACKEND_DOCUMENT_ROUTES, ...retired])].sort();
    expect(items.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route).sort()).toEqual(expected);
    expect(items.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(items.length - expected.length);
    expect(isMenuScreenPending("/gl/fiscal-years")).toBe(false);
    expect(isMenuScreenPending("/transaction/etax")).toBe(true);
    // 2026-10-11: ถอดจอเงาแล้ว แต่ใบรับเงินมัดจำยังเป็น "รอพัฒนา" — จอ ERP จริง (hasLineItems:false) ไม่มีช่องกรอกยอดเงิน
    // บันทึกแล้ว totalamount = 0 เสมอ; WorkTabPanel เช็ก isRetiredBackendScreen ก่อน isErpTransactionRoute จึงเป็นการ์ดรอพัฒนา
    expect(getErpModuleConfig("/transaction/receivedeposit")?.hasLineItems).toBe(false);
    expect(isRetiredBackendScreen("/transaction/receivedeposit")).toBe(true);
    expect(isMenuScreenPending("/transaction/receivedeposit")).toBe(true);
    // เอกสารธุรกรรมที่ยังมี config ทำงานจริงกับตาราง doc (ไม่ใช่ retired / pending)
    expect(isMenuBackendRetired("/transaction/landedcost")).toBe(false);
    expect(isMenuBackendRetired("/banking/cheques/deposit")).toBe(false);
    for (const route of ["/transaction/sale", "/transaction/paid", "/banking/cash-deposit", "/transaction/landedcost"]) {
      expect(isMenuScreenPending(route), route).toBe(false);
    }
    // จอสินค้า 3 จอไม่มี API จึงเป็น retired backend screen
    for (const route of ["/product", "/productbarcode", "/productbarcodeshelf"]) {
      expect(isMenuScreenPending(route), route).toBe(true);
    }
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
    // จอสมุดเงินฝากเป็นจอใหม่ (CB-2) ไม่ใช่จอตั้งค่าที่ถอด backend แล้ว
    expect(isMenuBackendRetired("/bookbankscreen")).toBe(false);
    expect(isMenuScreenPending("/bookbankscreen")).toBe(false);
    // Fallback guard for unmapped routes
    expect(isMenuScreenPending("/unknown-screen")).toBe(true);
  });

  it("connects every ledger workflow including the monthly budget entry screen (2026-09-25)", () => {
    const pending = GL_MENU_ITEMS.filter((item) => isMenuScreenPending(item.route)).map((item) => item.route);
    expect(pending).toEqual([]);
    expect(GL_MENU_ITEMS.filter((item) => !isMenuScreenPending(item.route))).toHaveLength(41);
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
    // รายงาน: มี API จริง 14 ตัว (4 รายงานสินค้า/ขาย + 10 รายงานเจ้าหนี้/ลูกหนี้ รวมอายุลูกหนี้/เจ้าหนี้)
    expect(isMenuDataPending("/report/salesreportbydocument")).toBe(false);
    expect(isMenuDataPending("/report/stockbalanceitem")).toBe(false);
    expect(isMenuDataPending("/report/apmovement")).toBe(false);
    expect(isMenuDataPending("/report/apstatus")).toBe(false);
    expect(isMenuDataPending("/report/armovement")).toBe(false);
    expect(isMenuDataPending("/report/arstatus")).toBe(false);
    expect(isMenuDataPending("/report/araging")).toBe(false);
    expect(isMenuDataPending("/report/apaging")).toBe(false);
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

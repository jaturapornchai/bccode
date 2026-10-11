import { describe, it, expect } from "vitest";
import { MENU_SECTIONS } from "./menu-data";
import { isMenuBackendRetired, isMenuScreenPending } from "./menu-screen-status";

describe("ERP Menu Completeness Audit", () => {
  // 2026-10-11: จอ UI-only 29 จอเป็น "รอพัฒนา" จนกว่าจะสร้างจริง — ประวัติรายการ/เหตุผลที่ docs/kms/champ/design-2026-10-11/facade-screens.md
  // เมื่อสร้างจอใดจริง ให้ลบ route นั้นออกจากรายการนี้ (และ CHAMP_PENDING_ROUTES ใน menu-screen-status.test.ts) ใน commit เดียวกัน
  const FACADE_PENDING_ROUTES = [
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

  it("keeps pending routes limited to the 29 UI-only (facade) screens set to รอพัฒนา, besides screens whose MongoDB API was removed", () => {
    const pendingList: string[] = [];
    for (const section of MENU_SECTIONS) {
      for (const group of section.groups) {
        for (const item of group.items) {
          if (isMenuScreenPending(item.route)) {
            if (!isMenuBackendRetired(item.route)) pendingList.push(item.route);
          }
        }
      }
    }
    expect([...pendingList].sort()).toEqual(FACADE_PENDING_ROUTES);
  });
});



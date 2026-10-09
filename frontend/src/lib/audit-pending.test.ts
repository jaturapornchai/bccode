import { describe, it, expect } from "vitest";
import { MENU_SECTIONS } from "./menu-data";
import { isMenuBackendRetired, isMenuScreenPending } from "./menu-screen-status";

describe("ERP Menu Completeness Audit", () => {
  // 2026-09-19 Champ parity: originally 31 routes from Champ menuconfig.xml without screen/backend.
  // On 2026-09-19, 8 AP/AR debt reports were connected via /api/report/debt/query, leaving 23 pending.
  // Then, 3 GL routes (/gl/journal-books, /report/gljournal, /report/budgetcomparison) connected, leaving 20 pending.
  it("keeps pending routes limited to the 19 Champ-parity items, the 20 Champ menu-bar items without a screen yet and screens whose MongoDB API was removed", () => {
    const pendingList: { group: string; title: string; route: string }[] = [];
    for (const section of MENU_SECTIONS) {
      for (const group of section.groups) {
        for (const item of group.items) {
          if (isMenuScreenPending(item.route)) {
            if (!isMenuBackendRetired(item.route)) pendingList.push({ group: group.title.th, title: item.label.th, route: item.route });
          }
        }
      }
    }
    // 2026-09-25: /gl/budget back to pending — monthly budget API ready, entry screen not built yet.
    // 2026-09-25: /report/vatsummary (Champ 5539) connected via /api/report/tax/vat-summary.
    // 2026-09-25: /gl/budget monthly entry screen connected (gl-budgets.tsx).
    // 2026-10-07: +20 items the real Champ menu has but BC lacks ("รอพัฒนา", routes /pending/*) — see menu-bars.test.ts.
    // 2026-10-08: -15 non-Champ items pruned to align with Champ parity, leaving 5 Champ-only items + 19 original = 24.
    // 2026-10-08: +5 เมนูที่ backend ไม่มีประเภทเอกสารรองรับ (etax, withholdingtax, purchasereceiptadjustment, petty-cash-limit, bank contacts) = 29.
    // 2026-10-09 (S1): +10 Champ transaction-head items without a screen yet (/pending/*) = 39.
    // 2026-10-09 (S2): +10 items of the Champ จัดการระบบ head without a screen yet (/pending/system-*) = 49.
    // 2026-10-10: +8 items of the Champ หน้าต่าง/ช่วยเหลือ heads without a screen yet (/pending/*) = 57.
    // 2026-10-10 (Champ Parity Phase 3): -1 purchase-receipt-date-adjust connected (/procurement/delivery-dates) = 56.
    expect(pendingList.length).toBe(56);
  });
});



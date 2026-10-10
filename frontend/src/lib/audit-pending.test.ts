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
    // 2026-10-10 (Champ Parity Phase 4): -2 purchasereceiptadjustment + customer-label-print connected = 54.
    // 2026-10-10: withholding-tax-deduction connected to /report/whtdeduction (Champ 5513/32993) = 50.
    // 2026-10-10: 4 fixed asset reports connected to FixedAssetsScreen (/report/depreciationmonthly, /report/depreciationyearly, /report/depreciationpnd50, /report/assetdisposal) = 46.
    // 2026-10-10: window-font-screen connected to /settings/font and bank-contact connected to /banking/contacts = 44.
    // 2026-10-10: 5 window font report screens connected to /settings/font-* and petty-cash-limit connected to /banking/petty-cash-limit = 38.
    // 2026-10-10: cheque-issued-prepare connected to /banking/cheque-issued-prepare (ChequeIssuedPrepareScreen) = 37.
    // 2026-10-10: billing-note-auto (/ar/billing-note-auto), transaction-lock (/tools/transaction-lock), and ap-payment-prepare (/ap/payment-prepare) connected = 34.
    // 2026-10-10: cheque-movement (/banking/cheque-movement) and bank-service (/banking/bank-service) connected = 32.
    // 2026-10-10: sale-reservation (/sale/reservation) and sale-dispatch (/sale/dispatch) connected = 30.
    // 2026-10-10: stock-count-handheld (/inventory/stock-count-handheld) and product-supplier-update (/inventory/product-supplier-update) connected = 28.
    // 2026-10-10: system-journal-format (/settings/journal-format) and system-work-date (/settings/work-date) connected = 26.
    // 2026-10-10: system-product-by-warehouse (/settings/product-by-warehouse) and system-default-warehouse (/settings/default-warehouse) connected = 24.
    // 2026-10-10: system-currency (/settings/currency) and system-payment-terms (/settings/payment-terms) connected = 22.
    // 2026-10-10: system-credit-debit-reasons (/settings/credit-debit-reasons) and system-main-map (/settings/main-map) connected = 20.
    // 2026-10-10: system-customer-inquiry (/settings/customer-inquiry) and system-product-inquiry (/settings/product-inquiry) connected = 18.
    // 2026-10-10: help-about (/settings/about) and help-standard-software (/settings/standard-software) connected = 16.
    // 2026-10-10 (Champ Parity Phase 8 - Round 10): 3 cheque and credit card reports connected (/report/chequereceived, /report/chequeissued, /report/creditcard) = 13.
    // 2026-10-10 (Champ Parity Phase 8 - Round 11): 4 bank and cash reports connected (/report/bankstatement, /report/cashmovement, /report/pettycashmovement, /report/monthlypaymentbook) = 9.
    // 2026-10-10 (Champ Parity Phase 8 - Round 12): 6 inventory parity reports connected (/report/maxstock, /report/nomovementstock, /report/stockcountvariance, /report/pendingreceive, /report/pendingdelivery, /report/serialmovement) = 3.
    expect(pendingList.length).toBe(3);
  });
});



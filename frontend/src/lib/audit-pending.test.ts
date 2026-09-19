import { describe, it, expect } from "vitest";
import { MENU_SECTIONS } from "./menu-data";
import { isMenuScreenPending } from "./menu-screen-status";

describe("ERP Menu Completeness Audit", () => {
  // 2026-09-19 Champ parity: originally 31 routes from Champ menuconfig.xml without screen/backend.
  // On 2026-09-19, 8 AP/AR debt reports were connected via /api/report/debt/query, leaving 23 pending.
  // Then, 3 GL routes (/gl/journal-books, /report/gljournal, /report/budgetcomparison) connected, leaving 20 pending.
  it("keeps pending routes limited to the 20 remaining Champ-parity items", () => {
    const pendingList: { group: string; title: string; route: string }[] = [];
    for (const section of MENU_SECTIONS) {
      for (const group of section.groups) {
        for (const item of group.items) {
          if (isMenuScreenPending(item.route)) {
            pendingList.push({ group: group.title.th, title: item.label.th, route: item.route });
          }
        }
      }
    }
    expect(pendingList.length).toBe(20);
  });
});



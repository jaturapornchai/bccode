import { describe, it, expect } from "vitest";
import { MENU_SECTIONS } from "./menu-data";
import { isMenuScreenPending } from "./menu-screen-status";

describe("ERP Menu Completeness Audit", () => {
  // 2026-09-19 Champ parity: 31 routes were added straight from Champ menuconfig.xml without a screen/backend yet;
  // the exact list lives in menu-screen-status.test.ts. Nothing else may be pending.
  it("keeps pending routes limited to the 31 Champ-parity additions", () => {
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
    expect(pendingList.length).toBe(31);
  });
});



import { describe, it, expect } from "vitest";
import { MENU_SECTIONS } from "./menu-data";
import { isMenuScreenPending } from "./menu-screen-status";

describe("ERP Menu Completeness Audit", () => {
  it("verifies all menu screens are implemented with zero pending routes", () => {
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
    expect(pendingList.length).toBe(0);
  });
});



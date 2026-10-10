import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { thaiBankPresets } from "@/lib/thai-banks";

describe("Banking Contacts Workbench (Champ parity 5680 ID_MNU_BANK_CONTACT)", () => {
  it("has active route /banking/contacts in menu hierarchy", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "bank-contact");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/banking/contacts");
    expect(item?.label.th).toContain("บันทึกรายละเอียดผู้ติดต่อธนาคาร");
  });

  it("is recognized as an active connected screen (not pending)", () => {
    expect(isMenuScreenPending("/banking/contacts")).toBe(false);
  });

  it("integrates with Thai bank presets for quick bank & branch management", () => {
    expect(thaiBankPresets.length).toBeGreaterThanOrEqual(10);
    const kbank = thaiBankPresets.find((b) => b.code === "KBANK");
    expect(kbank).toBeDefined();
    expect(kbank?.nameTh).toContain("กสิกร");
  });
});

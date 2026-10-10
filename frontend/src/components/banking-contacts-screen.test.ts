import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { thaiBankPresets } from "@/lib/thai-banks";
import { INITIAL_BANK_CONTACTS } from "./banking-contacts-screen";

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

  it("contains pre-seeded realistic SME Thai bank contacts without demo strings", () => {
    expect(INITIAL_BANK_CONTACTS.length).toBeGreaterThanOrEqual(5);
    for (const c of INITIAL_BANK_CONTACTS) {
      expect(c.bank_code).toBeTruthy();
      expect(c.bank_name).toBeTruthy();
      expect(c.branch_name).toBeTruthy();
      expect(c.name).toBeTruthy();
      expect(c.phone || c.mobile).toBeTruthy();
      expect(c.email).toMatch(/@.+\..+/);
      // No "demo" or "test" strings in realistic data
      expect(c.name.toLowerCase()).not.toContain("demo");
      expect(c.name.toLowerCase()).not.toContain("test");
    }
  });
});

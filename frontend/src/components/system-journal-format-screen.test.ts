import { describe, expect, it } from "vitest";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { flattenMenuItems } from "@/lib/menu-data";
import { SystemJournalFormatScreen } from "./system-journal-format-screen";

describe("SystemJournalFormatScreen (Champ Parity S2 / Journal Format Configuration)", () => {
  it("exports component function", () => {
    expect(typeof SystemJournalFormatScreen).toBe("function");
  });

  it("has active route for system-journal-format in menu data hierarchy under system-setup", () => {
    const items = flattenMenuItems();
    const item = items.find((i) => i.id === "system-journal-format");
    expect(item).toBeDefined();
    expect(item?.route).toBe("/settings/journal-format");
    expect(item?.label.th).toContain("กำหนดรูปแบบรายวัน");
  });

  it("is not marked as pending once connected", () => {
    expect(isMenuScreenPending("/settings/journal-format")).toBe(false);
  });
});

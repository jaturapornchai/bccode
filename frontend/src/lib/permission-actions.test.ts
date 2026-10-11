import { describe, expect, it } from "vitest";
import { screenActionsFor, setSpecialPermission, specialPermissionGranted } from "@/lib/permission-actions";

describe("screenActionsFor cancel", () => {
  it("allows cancel with the delete permission, matching the backend", () => {
    expect(screenActionsFor(["ap-payment-prepare", "ap-payment-prepare:delete"], "ap-payment-prepare").cancel).toBe(true);
  });

  it("allows cancel with only the cancel permission", () => {
    const actions = screenActionsFor(["ap-payment-prepare", "ap-payment-prepare:cancel"], "ap-payment-prepare");
    expect(actions.cancel).toBe(true);
    expect(actions.delete).toBe(false);
  });

  it("denies cancel without delete or cancel, allows it for *", () => {
    expect(screenActionsFor(["ap-payment-prepare", "ap-payment-prepare:update"], "ap-payment-prepare").cancel).toBe(false);
    expect(screenActionsFor(["*"], "ap-payment-prepare").cancel).toBe(true);
  });
});

// Champ แท็บ "เฉพาะ": print / export are granted until the group unticks them (BCSWinApp.cpp:1636-1653)
describe("special rights of a permission group", () => {
  it("are ticked without a stored decision, untick stores the denial and a tick removes it", () => {
    const entries = new Set(["gl-journals", "report:print"]);
    expect(specialPermissionGranted(entries, "report:print")).toBe(true);
    expect(specialPermissionGranted(entries, "report:export")).toBe(true);
    setSpecialPermission(entries, "report:export", false);
    setSpecialPermission(entries, "report:print", false);
    expect([...entries].sort()).toEqual(["!report:export", "!report:print", "gl-journals"]);
    expect(specialPermissionGranted(entries, "report:export")).toBe(false);
    setSpecialPermission(entries, "report:print", true);
    expect([...entries].sort()).toEqual(["!report:export", "gl-journals"]);
    expect(specialPermissionGranted(entries, "report:print")).toBe(true);
  });
});

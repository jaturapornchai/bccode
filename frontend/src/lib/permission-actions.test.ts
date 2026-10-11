import { describe, expect, it } from "vitest";
import { screenActionsFor } from "@/lib/permission-actions";

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

import { describe, expect, it } from "vitest";
import { amountFieldIssue, droppedWarningShown, normalizeAmountText } from "./gl-common";

// adversarial review 2026-09-24: European amounts changed value silently; the "characters removed"
// warning stuck to the input (rows keyed by index) and hid the real problem.
describe("normalizeAmountText — commas that are not thousands separators", () => {
  it("flags European format instead of turning 1.500,00 into 1.50 silently", () => {
    expect(normalizeAmountText("1.500,00").dropped).toBe(true);
    expect(normalizeAmountText("1 500,00").dropped).toBe(true);
    expect(normalizeAmountText("(1.500,00)").dropped).toBe(true);
  });
  it("keeps real thousands commas unflagged", () => {
    expect(normalizeAmountText("1,500.00")).toEqual({ value: "1500.00", dropped: false });
    expect(normalizeAmountText("-12,345,678.90")).toEqual({ value: "-12345678.90", dropped: false });
  });
});

describe("droppedWarningShown — the warning belongs to the value that lost characters", () => {
  it("shows while editing and after blur on the committed value", () => {
    expect(droppedWarningShown("1500", true, "1500")).toBe(true);
    expect(droppedWarningShown("1500.00", false, "1500.00")).toBe(true);
  });
  it("hides when the input now shows another row's value or a new voucher", () => {
    expect(droppedWarningShown("1500.00", false, "2000.00")).toBe(false);
    expect(droppedWarningShown("1500.00", false, "0")).toBe(false);
    expect(droppedWarningShown(null, true, "1500")).toBe(false);
  });
});

describe("amountFieldIssue — real problems first", () => {
  it("says 'negative not allowed' for a-5, not 'characters removed'", () => {
    expect(amountFieldIssue("-5", 2, false, true)).toBe("negative");
    expect(amountFieldIssue("1.234", 2, false, true)).toBe("scale");
    expect(amountFieldIssue("15", 2, false, true)).toBe("characters");
    expect(amountFieldIssue("15", 2, false, false)).toBe("");
  });
});

import { describe, expect, it } from "vitest";
import { hasInvalidHoldingScopeRules, normalizeHoldingScopeRules, toggleHoldingScopeRules } from "./holding-scope-editor";

// Holding-wide access now covers companies created later, so only an explicit
// scopetype "holding" may produce it (decision 2026-09-24).
describe("holding scope rule normalization", () => {
  it("keeps an explicit holding rule", () => {
    expect(normalizeHoldingScopeRules([{ scopetype: "HOLDING", businesscode: "C01" }])).toEqual([{ scopetype: "holding" }]);
  });

  it("drops unknown or company-less rules instead of promoting them to holding", () => {
    expect(
      normalizeHoldingScopeRules([{}, { businesscode: "C01" }, { scopetype: "everything", businesscode: "C01" }, { scopetype: "company" }]),
    ).toEqual([]);
    expect(hasInvalidHoldingScopeRules([{ businesscode: "C01" }], true)).toBe(true);
  });

  it("keeps a companyuid-only company rule as a company rule", () => {
    expect(normalizeHoldingScopeRules([{ scopetype: "company", companyuid: "c01" }])).toEqual([
      { scopetype: "company", businesscode: "C01", allbranches: true },
    ]);
    expect(normalizeHoldingScopeRules([{ scopetype: "branch", companyuid: "C01", branchuid: "1" }])).toEqual([
      { scopetype: "branch", businesscode: "C01", branchcode: "00001", allbranches: false },
    ]);
  });

  it("restores the company/branch picks when 'whole holding' is ticked then unticked", () => {
    const picks = normalizeHoldingScopeRules([
      { scopetype: "company", businesscode: "C01" },
      { scopetype: "branch", businesscode: "C02", branchcode: "1" },
    ]);
    const ticked = toggleHoldingScopeRules(picks, true, []);
    expect(ticked.rules).toEqual([{ scopetype: "holding" }]);
    const unticked = toggleHoldingScopeRules(ticked.rules, false, ticked.picksBeforeHolding);
    expect(unticked.rules).toEqual(picks);
    expect(unticked.picksBeforeHolding).toEqual([]);
    // A record that was holding-wide when opened has nothing to restore.
    expect(toggleHoldingScopeRules([{ scopetype: "holding" }], false, []).rules).toEqual([]);
  });
});

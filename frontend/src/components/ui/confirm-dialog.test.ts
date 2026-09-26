import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { CONFIRM_EARLY_CLICK_MS, isEarlyConfirmClick } from "./confirm-dialog";

// A double-click whose first click opens the dialog must not also answer it with the second click.
describe("confirm dialog ignores the click that opened it", () => {
  it("ignores the second click of a double-click and any click right after opening", () => {
    expect(isEarlyConfirmClick(2, 1000, 3000)).toBe(true);
    expect(isEarlyConfirmClick(3, 1000, 3000)).toBe(true);
    expect(isEarlyConfirmClick(1, 1000, 1000 + CONFIRM_EARLY_CLICK_MS - 1)).toBe(true);
    expect(isEarlyConfirmClick(0, 1000, 1100)).toBe(true);
  });

  it("accepts a single click or keyboard press once the dialog has been visible long enough", () => {
    expect(isEarlyConfirmClick(1, 1000, 1000 + CONFIRM_EARLY_CLICK_MS)).toBe(false);
    expect(isEarlyConfirmClick(0, 1000, 5000)).toBe(false);
  });

  it("every mouse path that answers the dialog goes through the guard", () => {
    const source = readFileSync(resolve(process.cwd(), "src", "components", "ui", "confirm-dialog.tsx"), "utf8");
    expect(source).toMatch(/openedAtRef\.current = performance\.now\(\);\s*setPending\(/);
    expect(source).toMatch(/if \(event\.target === event\.currentTarget\) closeOnClick\(false, event\.detail\);/);
    expect(source.match(/onClick=\{\(event\) => closeOnClick\(false, event\.detail\)\}/g)).toHaveLength(2);
    expect(source).toMatch(/onClick=\{\(event\) => closeOnClick\(true, event\.detail\)\}/);
    expect(source).not.toMatch(/onClick=\{\(\) => close\(/);
  });
});

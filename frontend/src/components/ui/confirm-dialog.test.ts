import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement, type ComponentProps } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { BackendTextProvider } from "@/components/backend-text-provider";
import { CONFIRM_EARLY_CLICK_MS, isEarlyConfirmClick, useConfirmDialogDefaults, type UseConfirmDialogOptions } from "./confirm-dialog";

const readSource = () => readFileSync(resolve(process.cwd(), "src", "components", "ui", "confirm-dialog.tsx"), "utf8");

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
    const source = readSource();
    expect(source).toMatch(/openedAtRef\.current = performance\.now\(\);\s*setPending\(/);
    expect(source).toMatch(/if \(event\.target === event\.currentTarget\) closeOnClick\(false, event\.detail\);/);
    expect(source.match(/onClick=\{\(event\) => closeOnClick\(false, event\.detail\)\}/g)).toHaveLength(2);
    expect(source).toMatch(/onClick=\{\(event\) => closeOnClick\(true, event\.detail\)\}/);
    expect(source).not.toMatch(/onClick=\{\(\) => close\(/);
  });
});

// A shared component must not embed Thai (AGENTS.md i18n rule 2026-09-14): default labels follow the selected language.
function DefaultLabels({ defaults }: { defaults?: UseConfirmDialogOptions }) {
  const labels = useConfirmDialogDefaults(defaults);
  return createElement("span", null, `${labels.confirmLabel}|${labels.cancelLabel}`);
}

const renderDefaults = (dictionary: Record<string, string> | null, defaults?: UseConfirmDialogOptions) => {
  const probe = createElement(DefaultLabels, { defaults });
  return renderToStaticMarkup(dictionary ? createElement(BackendTextProvider, { dictionary } as ComponentProps<typeof BackendTextProvider>, probe) : probe);
};

describe("confirm dialog default labels come from the language table", () => {
  it("keeps no Thai text in the shared component source", () => {
    expect(readSource()).not.toMatch(/[\u0E00-\u0E7F]/);
  });

  it("uses common_confirm / common_cancel from the surrounding BackendTextProvider", () => {
    expect(renderDefaults({ common_confirm: "ยืนยัน", common_cancel: "ยกเลิก" }))
      .toBe("<span>ยืนยัน|ยกเลิก</span>");
    expect(renderDefaults({ common_confirm: "Confirm (en)", common_cancel: "Cancel (en)" })).toBe("<span>Confirm (en)|Cancel (en)</span>");
  });

  it("lets labels from the caller's own dictionary win over the provider", () => {
    expect(renderDefaults({ common_confirm: "Confirm (en)", common_cancel: "Cancel (en)" }, { defaultConfirmLabel: "OK", defaultCancelLabel: "Back" }))
      .toBe("<span>OK|Back</span>");
  });

  it("falls back to neutral English text outside any provider", () => {
    expect(renderDefaults(null)).toBe("<span>Confirm|Cancel</span>");
  });
});

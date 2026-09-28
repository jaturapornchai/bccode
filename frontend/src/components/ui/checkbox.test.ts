import { describe, expect, it } from "vitest";
import React from "react";
import { renderToString } from "react-dom/server";
import { Checkbox, CheckboxCard } from "./checkbox";

describe("Checkbox component", () => {
  it("renders unchecked state correctly", () => {
    const html = renderToString(React.createElement(Checkbox, { checked: false }));
    expect(html).toContain('type="checkbox"');
    expect(html).toContain("border-muted-foreground/40");
  });

  it("renders checked state with active primary styling and CheckIcon", () => {
    const html = renderToString(React.createElement(Checkbox, { checked: true }));
    expect(html).toContain('type="checkbox"');
    expect(html).toContain("bg-primary");
    expect(html).toContain("text-primary-foreground");
    expect(html).toContain("<svg");
  });

  it("renders disabled state with appropriate opacity", () => {
    const html = renderToString(React.createElement(Checkbox, { checked: false, disabled: true }));
    expect(html).toContain("opacity-50");
    expect(html).toContain('disabled=""');
    expect(html).not.toContain("peer-hover:border-primary/60");
  });

  // Bug 2026-09-28: the box was a dead aria-hidden span over an sr-only input, so clicking it
  // outside a <label> (account picker multi-select) did nothing. The real input must cover the box.
  it("lays the real input over the drawn box so clicking the box toggles it", () => {
    const html = renderToString(React.createElement(Checkbox, { checked: false }));
    const inputClasses = (html.match(/<input[^>]*class="([^"]*)"/)?.[1] ?? "").split(" ");
    expect(inputClasses).not.toContain("sr-only");
    // min-h-0! (important): the unlayered global `input { min-height: 2.6em }` otherwise makes the
    // invisible input taller than the box, so a click on the next row toggled this checkbox.
    for (const cls of ["peer", "absolute", "inset-0", "size-full", "min-h-0!", "opacity-0", "m-0", "appearance-none", "cursor-pointer"]) {
      expect(inputClasses).toContain(cls);
    }
    const box = html.match(/<span aria-hidden="true" class="([^"]*)"/)?.[1] ?? "";
    expect(box.split(" ")).toContain("pointer-events-none");
    expect(box).toContain("peer-focus-visible:ring-2");
    expect(box).toContain("peer-hover:border-primary/60");
    // peer-* only reaches later siblings: the input must come before the box.
    expect(html.indexOf("<input")).toBeLessThan(html.indexOf('aria-hidden="true"'));
  });

  it("renders CheckboxCard with label and padding", () => {
    const html = renderToString(
      React.createElement(CheckboxCard, {
        label: "เปิดใช้งานระบบ",
        hint: "อนุญาตให้ผู้ใช้เข้าถึง",
        checked: true,
      })
    );
    expect(html).toContain("เปิดใช้งานระบบ");
    expect(html).toContain("อนุญาตให้ผู้ใช้เข้าถึง");
    expect(html).toContain("bg-primary/10");
  });
});

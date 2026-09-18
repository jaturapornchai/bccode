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

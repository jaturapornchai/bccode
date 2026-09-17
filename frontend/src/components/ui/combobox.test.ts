import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { Combobox } from "./combobox";

describe("Combobox Component", () => {
  it("renders with options prop and displays selected label", () => {
    const html = renderToStaticMarkup(
      createElement(Combobox, {
        value: "asset",
        options: [
          { value: "asset", label: "สินทรัพย์" },
          { value: "liability", label: "หนี้สิน" },
        ],
      })
    );
    expect(html).toContain("สินทรัพย์");
    expect(html).toContain('role="combobox"');
    expect(html).toContain("min-h-[2.6em]");
  });

  it("renders with children <option> elements and displays selected label", () => {
    const html = renderToStaticMarkup(
      createElement(
        Combobox,
        {
          value: 3,
          "data-field": "level",
          "aria-label": "ระดับบัญชี",
        },
        createElement("option", { value: 1 }, "ระดับ 1"),
        createElement("option", { value: 2 }, "ระดับ 2"),
        createElement("option", { value: 3 }, "ระดับ 3")
      )
    );
    expect(html).toContain("ระดับ 3");
    expect(html).toContain('data-field="level"');
    expect(html).toContain('aria-label="ระดับบัญชี"');
  });

  it("renders custom placeholder when value is empty or not matched", () => {
    const html = renderToStaticMarkup(
      createElement(Combobox, {
        value: "",
        placeholder: "กรุณาเลือกระดับบัญชี...",
        options: [{ value: "1", label: "ระดับ 1" }],
      })
    );
    expect(html).toContain("กรุณาเลือกระดับบัญชี...");
  });

  it("renders depth shadow and accessible classes", () => {
    const html = renderToStaticMarkup(
      createElement(Combobox, {
        value: "1",
        options: [{ value: "1", label: "ระดับ 1" }],
      })
    );
    expect(html).toContain("shadow-[0_3px_10px_rgba(0,0,0,0.14),0_1px_3px_rgba(0,0,0,0.1)]");
    expect(html).toContain("dark:shadow-[0_3px_10px_rgba(0,0,0,0.6)]");
  });

  it("preserves shadow and styles when disabled (Thai 40+ read-only clarity)", () => {
    const html = renderToStaticMarkup(
      createElement(Combobox, {
        value: "1",
        disabled: true,
        options: [{ value: "1", label: "ระดับ 1" }],
      })
    );
    expect(html).toContain("disabled");
    expect(html).toContain("shadow-[0_2px_8px_rgba(0,0,0,0.1),0_1px_2px_rgba(0,0,0,0.07)]");
    expect(html).toContain("bg-muted/20");
  });

  it("renders hidden input with name when name prop is provided", () => {
    const html = renderToStaticMarkup(
      createElement(Combobox, {
        name: "account_level",
        value: "5",
        options: [{ value: "5", label: "ระดับ 5" }],
      })
    );
    expect(html).toContain('type="hidden"');
    expect(html).toContain('name="account_level"');
    expect(html).toContain('value="5"');
  });
});

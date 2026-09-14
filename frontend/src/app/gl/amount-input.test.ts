import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { AmountInput, cleanAmountValue, formatAmountValue } from "./gl-common";

describe("formatAmountValue", () => {
  it("formats integers with thousands commas and default 2 decimals", () => {
    expect(formatAmountValue("70000")).toBe("70,000.00");
    expect(formatAmountValue("1000000")).toBe("1,000,000.00");
    expect(formatAmountValue("0")).toBe("0.00");
  });

  it("pads single decimal places to specified scale", () => {
    expect(formatAmountValue("70000.5")).toBe("70,000.50");
    expect(formatAmountValue("1234.5", 2)).toBe("1,234.50");
  });

  it("handles custom scale (0, 3, 4)", () => {
    expect(formatAmountValue("70000", 0)).toBe("70,000");
    expect(formatAmountValue("70000", 3)).toBe("70,000.000");
    expect(formatAmountValue("70000.1234", 4)).toBe("70,000.1234");
  });

  it("handles negative values correctly", () => {
    expect(formatAmountValue("-70000")).toBe("-70,000.00");
    expect(formatAmountValue("-70000.5")).toBe("-70,000.50");
    expect(formatAmountValue("-0")).toBe("0.00");
    expect(formatAmountValue("-0.00")).toBe("0.00");
  });

  it("returns empty string for empty or whitespace values", () => {
    expect(formatAmountValue("")).toBe("");
    expect(formatAmountValue("   ")).toBe("");
  });

  it("strips pre-existing commas cleanly", () => {
    expect(formatAmountValue("70,000")).toBe("70,000.00");
    expect(formatAmountValue("1,234,567.89")).toBe("1,234,567.89");
  });
});

describe("cleanAmountValue", () => {
  it("strips thousands commas", () => {
    expect(cleanAmountValue("70,000.00")).toBe("70000.00");
    expect(cleanAmountValue("1,234,567.89")).toBe("1234567.89");
  });

  it("strips non-numeric characters", () => {
    expect(cleanAmountValue("abc70,000.50xyz")).toBe("70000.50");
    expect(cleanAmountValue("$1,000.00")).toBe("1000.00");
  });

  it("prevents multiple decimal points", () => {
    expect(cleanAmountValue("12.34.56")).toBe("12.3456");
  });

  it("handles negative sign based on allowNegative flag", () => {
    expect(cleanAmountValue("-70000", false)).toBe("70000");
    expect(cleanAmountValue("-70000", true)).toBe("-70000");
    expect(cleanAmountValue("70-000", true)).toBe("70000");
  });
});

describe("AmountInput Component Rendering", () => {
  it("renders with text-right and tabular-nums classes", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "70000",
        onChange: () => {},
      }),
    );
    expect(html).toContain("text-right");
    expect(html).toContain("tabular-nums");
    expect(html).toContain('inputMode="decimal"');
  });

  it("formats value with commas and decimals when idle", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "70000",
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="70,000.00"');
  });

  it("renders empty string when value is empty and allowEmpty is true", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "",
        allowEmpty: true,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value=""');
  });

  it("renders 0.00 when value is empty but required is true", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "",
        required: true,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="0.00"');
  });

  it("respects custom scale in idle rendering", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "50000",
        scale: 4,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="50,000.0000"');
  });

  it("renders aria-label and disabled state", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "1000",
        disabled: true,
        ariaLabel: "จำนวนเงินงบประมาณ",
        onChange: () => {},
      }),
    );
    expect(html).toContain('aria-label="จำนวนเงินงบประมาณ"');
    expect(html).toContain("disabled");
  });

  it("renders default placeholder matching scale", () => {
    const html2 = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "",
        scale: 2,
        onChange: () => {},
      }),
    );
    expect(html2).toContain('placeholder="0.00"');

    const html0 = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "",
        scale: 0,
        onChange: () => {},
      }),
    );
    expect(html0).toContain('placeholder="0"');
  });
});


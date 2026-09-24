import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import {
  AmountInput,
  amountBlurValue,
  amountFocusText,
  amountInputIssue,
  commitAmountText,
  formatAmountValue,
  normalizeAmountText,
} from "./gl-common";

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

describe("normalizeAmountText", () => {
  it("strips thousands commas, spaces and currency signs without flagging", () => {
    expect(normalizeAmountText("70,000.00")).toEqual({ value: "70000.00", dropped: false });
    expect(normalizeAmountText(" 1,234,567.89 ")).toEqual({ value: "1234567.89", dropped: false });
    expect(normalizeAmountText("$1,000.00")).toEqual({ value: "1000.00", dropped: false });
    expect(normalizeAmountText("\u0E3F1,000")).toEqual({ value: "1000", dropped: false });
  });

  it("converts Thai digits to 0-9", () => {
    expect(normalizeAmountText("\u0E51,\u0E52\u0E53\u0E54.\u0E55\u0E50").value).toBe("1234.50");
    expect(normalizeAmountText("\u0E50").value).toBe("0");
  });

  it("keeps the minus sign (a positive-only field shows a message instead of dropping it)", () => {
    expect(normalizeAmountText("-500").value).toBe("-500");
    expect(normalizeAmountText("(1,500.00)")).toEqual({ value: "-1500.00", dropped: false });
  });

  it("flags removed characters instead of changing the number silently", () => {
    expect(normalizeAmountText("abc70,000.50xyz")).toEqual({ value: "70000.50", dropped: true });
    expect(normalizeAmountText("12.34.56")).toEqual({ value: "12.3456", dropped: true });
    expect(normalizeAmountText("70-000").dropped).toBe(true);
  });
});

describe("amountInputIssue", () => {
  it("reports a negative in a positive-only field", () => {
    expect(amountInputIssue("-500", 2, false)).toBe("negative");
    expect(amountInputIssue("-500", 2, true)).toBe("");
    expect(amountInputIssue("-0.00", 2, false)).toBe("");
  });

  it("reports more decimals than the scale (never rounded)", () => {
    expect(amountInputIssue("1234.567", 2, false)).toBe("scale");
    expect(amountInputIssue("1234.560", 2, false)).toBe("");
    expect(amountInputIssue("1234.56", 2, false)).toBe("");
    expect(amountInputIssue("5.5", 0, false)).toBe("scale");
  });
});

describe("AmountInput focus/blur helpers", () => {
  it("tabbing through an untouched 0.00 field keeps the stored value (no \"\" emitted)", () => {
    expect(amountFocusText("0")).toBe("");
    expect(amountFocusText("0.00")).toBe("");
    expect(amountBlurValue(false, "", 2, "0.00")).toBeNull();
    expect(amountBlurValue(false, "", 2, "")).toBeNull();
  });

  it("a cleared positive-only line amount commits as zero, never an empty string", () => {
    expect(amountBlurValue(true, "", 2, "0.00")).toBe("0.00");
    expect(amountBlurValue(true, "-", 2, "0.00")).toBe("0.00");
  });

  it("focus shows plain digits without commas", () => {
    expect(amountFocusText("70,000.50")).toBe("70000.50");
  });
});

describe("commitAmountText", () => {
  it("pads to the scale and strips leading zeros", () => {
    expect(commitAmountText("1500", 2, "")).toBe("1500.00");
    expect(commitAmountText("007.5", 2, "")).toBe("7.50");
    expect(commitAmountText("1,234.5", 2, "")).toBe("1234.50");
  });

  it("never rounds: extra significant decimals stay as typed", () => {
    expect(commitAmountText("1234.567", 2, "")).toBe("1234.567");
    expect(commitAmountText("1234.560", 2, "")).toBe("1234.56");
  });

  it("keeps a real negative but not a negative zero", () => {
    expect(commitAmountText("-1500", 2, "")).toBe("-1500.00");
    expect(commitAmountText("-0", 2, "0.00")).toBe("0.00");
  });
});

describe("AmountInput Component Rendering", () => {
  it("shows an inline message when a positive-only field holds a negative", () => {
    const html = renderToStaticMarkup(
      createElement(AmountInput, {
        value: "-500",
        onChange: () => {},
      }),
    );
    expect(html).toContain('role="alert"');
    expect(html).toContain('aria-invalid="true"');
    expect(html).toContain("ไม่รับยอดติดลบ");
  });

  it("shows no message for a valid amount", () => {
    const html = renderToStaticMarkup(createElement(AmountInput, { value: "1500", onChange: () => {} }));
    expect(html).not.toContain('aria-invalid="true"');
  });

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


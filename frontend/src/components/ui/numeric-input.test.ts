import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { NumericInput } from "./numeric-input";

describe("NumericInput Component", () => {
  it("renders with text-right and tabular-nums classes", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 70000,
        onChange: () => {},
      }),
    );
    expect(html).toContain("text-right");
    expect(html).toContain("tabular-nums");
    expect(html).toContain('inputMode="decimal"');
  });

  it("formats value with commas and default 2 decimals when idle", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 70000,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="70,000.00"');
  });

  it("formats zero as 0.00 when idle", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 0,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="0.00"');
  });

  it("keeps stored precision beyond the default decimals instead of rounding", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 0.125,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="0.125"');
  });

  it("respects custom decimals prop", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 1234.5,
        decimals: 4,
        onChange: () => {},
      }),
    );
    expect(html).toContain('value="1,234.5000"');
  });

  it("renders calculator button alongside input", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 500,
        onChange: () => {},
      }),
    );
    expect(html).toContain("เปิดเครื่องคิดเลข");
  });

  it("renders default placeholder matching decimals", () => {
    const html = renderToStaticMarkup(
      createElement(NumericInput, {
        value: 0,
        decimals: 2,
        onChange: () => {},
      }),
    );
    expect(html).toContain('placeholder="0.00"');
  });
});

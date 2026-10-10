import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { SystemStandardSoftwareScreen } from "./system-standard-software-screen";

describe("SystemStandardSoftwareScreen", () => {
  it("renders standard software compliance screen with Champ Parity badge", () => {
    const html = renderToStaticMarkup(
      createElement(SystemStandardSoftwareScreen, { language: "th", holdingcode: "rungrueng" })
    );
    expect(html).toContain("มาตรฐานซอฟท์แวร์ของกรมสรรพากร (Standard Software)");
    expect(html).toContain("Champ Parity (S2)");
    expect(html).toContain("ประกาศอธิบดีกรมสรรพากร");
    expect(html).toContain("Double-Entry Balance 100%");
    expect(html).toContain("Tax ID Validation");
  });

  it("renders in English correctly", () => {
    const html = renderToStaticMarkup(
      createElement(SystemStandardSoftwareScreen, { language: "en", holdingcode: "rungrueng" })
    );
    expect(html).toContain("Revenue Department Standard Software Compliance");
    expect(html).toContain("Print Compliance");
  });
});

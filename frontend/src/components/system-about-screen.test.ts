import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { SystemAboutScreen } from "./system-about-screen";

describe("SystemAboutScreen", () => {
  it("renders about screen with version and Champ Parity badge", () => {
    const html = renderToStaticMarkup(
      createElement(SystemAboutScreen, { language: "th", holdingcode: "rungrueng" })
    );
    expect(html).toContain("เกี่ยวกับโปรแกรม (About BC Ai Account)");
    expect(html).toContain("Champ Parity (S2)");
    expect(html).toContain("BC Ai Account Cloud ERP");
    expect(html).toContain("PostgreSQL 18");
    expect(html).toContain("No Float Accounting");
  });

  it("renders in English correctly", () => {
    const html = renderToStaticMarkup(
      createElement(SystemAboutScreen, { language: "en", holdingcode: "rungrueng" })
    );
    expect(html).toContain("About BC Ai Account");
    expect(html).toContain("Enterprise Cloud ERP");
  });
});

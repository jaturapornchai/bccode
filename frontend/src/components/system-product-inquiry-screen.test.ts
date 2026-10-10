import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { SystemProductInquiryScreen } from "./system-product-inquiry-screen";

describe("SystemProductInquiryScreen", () => {
  it("renders product inquiry header with Champ Parity badge", () => {
    const html = renderToStaticMarkup(
      createElement(SystemProductInquiryScreen, { language: "th", holdingcode: "rungrueng" })
    );
    expect(html).toContain("แสดงข้อมูลสินค้า");
    expect(html).toContain("Champ Parity (S2)");
    expect(html).toContain("PROD-0001");
    expect(html).toContain("8851234567890");
    expect(html).toContain("ปูนซีเมนต์ปอร์ตแลนด์");
    expect(html).toContain("WH-01");
  });

  it("renders in English properly", () => {
    const html = renderToStaticMarkup(
      createElement(SystemProductInquiryScreen, { language: "en", holdingcode: "rungrueng" })
    );
    expect(html).toContain("Product Information Inquiry");
    expect(html).toContain("Portland Cement Diamond Brand 50kg");
    expect(html).toContain("Refresh");
  });
});

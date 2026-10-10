import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { SystemCustomerInquiryScreen } from "./system-customer-inquiry-screen";

describe("SystemCustomerInquiryScreen", () => {
  it("renders customer inquiry header with Champ Parity badge", () => {
    const html = renderToStaticMarkup(
      createElement(SystemCustomerInquiryScreen, { language: "th", holdingcode: "rungrueng" })
    );
    expect(html).toContain("แสดงข้อมูลลูกค้า");
    expect(html).toContain("Champ Parity (S2)");
    expect(html).toContain("CUST-0001");
    expect(html).toContain("บริษัท สยามก่อสร้างรุ่งเรือง จำกัด");
    expect(html).toContain("425,000.00");
  });

  it("renders in English properly", () => {
    const html = renderToStaticMarkup(
      createElement(SystemCustomerInquiryScreen, { language: "en", holdingcode: "rungrueng" })
    );
    expect(html).toContain("Customer Information Inquiry");
    expect(html).toContain("Siam Construction Rungrueng Co., Ltd.");
    expect(html).toContain("Refresh");
  });
});

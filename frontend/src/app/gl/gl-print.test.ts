import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { GLJournal, GLReport } from "@/lib/general-ledger";
import { GLStatementTable, GLVoucherPrint, printAmountCell, printOrientation, printPageStyle, statementAmountText, statementPeriodText, visiblePrintColumns } from "./gl-print";

const tr = (_key: string, fallback: string) => fallback;

// สมุดบัญชีต้องมีเลขหน้าเรียงทุกหน้า (ประกาศกรมทะเบียนการค้า 2544 ข้อ 5(2)) — ป้ายมาจาก languages.tsv จึงต้องเป็น CSS string ที่ปลอดภัย
describe("printPageStyle", () => {
  it("numbers every page of both A4 orientations", () => {
    const css = printPageStyle("หน้า");
    expect(css).toContain(`@page gl-print-portrait { size: A4 portrait;`);
    expect(css).toContain(`@page gl-print-landscape { size: A4 landscape;`);
    expect(css.match(/content: "หน้า" " " counter\(page\) " \/ " counter\(pages\)/g)).toHaveLength(2);
  });
  it("escapes quotes and cannot close the <style> element", () => {
    const css = printPageStyle(`Page "x" </style><script>`);
    expect(css).toContain(`"Page \\"x\\" \\3c /style>\\3c script>"`);
    expect(css).not.toContain("</style>");
  });
});

describe("visiblePrintColumns", () => {
  const columns = [{ key: "date" }, { key: "departmentcode" }, { key: "description" }, { key: "debit", amount: true }, { key: "credit", amount: true }];
  it("drops text columns that are blank on every row but keeps money columns", () => {
    const rows = [{ date: "2026-09-01", departmentcode: " ", description: "ซื้อปูน", debit: "", credit: "" }];
    expect(visiblePrintColumns(columns, rows).map((column) => column.key)).toEqual(["date", "description", "debit", "credit"]);
  });
  it("keeps every column when there are no rows", () => {
    expect(visiblePrintColumns(columns, [])).toHaveLength(5);
  });
  it("switches to landscape only for wide tables", () => {
    expect(printOrientation(7)).toBe("portrait");
    expect(printOrientation(8)).toBe("landscape");
  });
});

// ใบสำคัญที่ทำขึ้นใช้เอง (ข้อ 9 + ข้อ 11): ผู้จัดทำ ชื่อเอกสาร เลขที่ วันที่ ยอดรวม คำอธิบาย วิธีคำนวณ ลายมือชื่อผู้อนุมัติ
describe("printAmountCell", () => {
  it("blanks the empty debit/credit side but keeps zero balances", () => {
    expect(printAmountCell("debit", "0")).toBe("");
    expect(printAmountCell("endingcredit", "0.00000000")).toBe("");
    expect(printAmountCell("credit", "1234.5")).toBe("1,234.50");
    expect(printAmountCell("balance", "0")).toBe("0.00");
    expect(printAmountCell("debit", "0.001", 2)).toBe("0.001");
  });
});

describe("GLVoucherPrint", () => {
  const journal: GLJournal = {
    docno: "UV6909-0012", date: "2026-09-15", bookcode: "UV", fiscalyear: "2569", description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 500 ถุง เป็นเงินเชื่อ",
    reference: "INV-889", branchcode: "00000", kind: "normal", status: "draft",
    lines: [
      { accountcode: "1131", accountname: "สินค้าคงเหลือ - ปูนซีเมนต์", description: "ปูน 500 ถุง", debit: "70000.10", credit: "", departmentcode: "", projectcode: "", cashflow: "" },
      { accountcode: "1151", accountname: "ภาษีซื้อ", description: "", debit: "4900.01", credit: "", departmentcode: "", projectcode: "", cashflow: "" },
      { accountcode: "2111", accountname: "เจ้าหนี้การค้า", description: "", debit: "", credit: "74900.11", departmentcode: "", projectcode: "", cashflow: "" },
    ],
    details: { vats: [{ id: "v1", tax_type: 1, document_type: 1, tax_invoice_no: "IV-0098", tax_invoice_date: "2026-09-14", partner_name: "บริษัท สยามซีเมนต์ค้าส่ง จำกัด", base_amount: "70000.10", zero_rate_amount: "0", exempt_amount: "0", vat_rate: "7.00", vat_amount: "4900.01" }] },
  };
  const html = renderToStaticMarkup(createElement(GLVoucherPrint, { journal, books: [{ code: "UV", name: "สมุดรายวันซื้อ", booktype: 5, isactive: true }], company: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", tr, language: "th" }));
  it("carries the issuer, document name, number, date and exact totals", () => {
    expect(html).toContain("บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด");
    expect(html).toContain("ใบสำคัญรายวัน — สมุดรายวันซื้อ");
    expect(html).toContain("UV6909-0012");
    expect(html).toContain("15 ก.ย. 2569");
    // 70,000.10 + 4,900.01 = 74,900.11: บรรทัดเจ้าหนี้ + ยอดรวมเดบิต + ยอดรวมเครดิต (BigInt ไม่มีเศษสตางค์เพี้ยน)
    expect(html.match(/74,900\.11/g)).toHaveLength(3);
  });
  it("shows the explanation, the tax calculation, a draft mark and three signature boxes", () => {
    expect(html).toContain("ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 500 ถุง เป็นเงินเชื่อ");
    expect(html).toContain("รายละเอียดภาษีมูลค่าเพิ่ม");
    expect(html).toContain("IV-0098");
    expect(html).toContain("ฉบับร่าง — ยังไม่ผ่านรายการ");
    for (const label of ["ผู้จัดทำ", "ผู้ตรวจสอบ", "ผู้อนุมัติ"]) expect(html).toContain(label);
    expect(html).not.toContain("รหัสแผนก");
    // เจ้าหนี้ไม่มีเดบิต: ช่องว่าง ไม่ใช่ 0.00
    expect(html).not.toContain(">0.00<");
  });
});

// หัวงบตาม TFRS for NPAEs ย่อหน้า 4.7 (ชื่อกิจการ ชื่องบ วันที่/งวด หน่วยเงิน) + คอลัมน์ปีก่อน ย่อหน้า 4.3 — ตัวเลขมาจาก backend
describe("financial statement print", () => {
  it("dates the heading as a point in time, a full year or a partial period", () => {
    const period = { from: "2026-01-01", to: "2026-12-31" };
    expect(statementPeriodText(period, true, true, tr, "th")).toBe("ณ วันที่ 31 ธันวาคม 2569");
    expect(statementPeriodText(period, false, true, tr, "th")).toBe("สำหรับปีสิ้นสุดวันที่ 31 ธันวาคม 2569");
    expect(statementPeriodText({ from: "2026-04-01", to: "2026-06-30" }, false, false, tr, "th")).toBe("สำหรับงวดตั้งแต่วันที่ 1 เมษายน 2569 ถึงวันที่ 30 มิถุนายน 2569");
    expect(statementPeriodText(undefined, true, true, tr, "th")).toBe("");
  });

  it("shows zero as a dash unless the row asks for zero", () => {
    expect(statementAmountText("0.00", false)).toBe("-");
    expect(statementAmountText("0.00", true)).toBe("0.00");
    expect(statementAmountText("-1234.50", false)).toBe("-1,234.50");
    expect(statementAmountText("", false)).toBe("");
  });

  it("renders company, title, period, unit, comparative columns and underlines", () => {
    const report: GLReport = {
      columns: [{ key: "rowno", label: "ลำดับ" }, { key: "title", label: "รายการ" }, { key: "noteno", label: "หมายเหตุ" }, { key: "amount", label: "2569", amount: true }, { key: "prioramount", label: "2568", amount: true }],
      rows: [
        { rowno: "10", title: "สินทรัพย์", rowtype: "header", indent: "0", fontweight: "bold", amount: "", prioramount: "" },
        { rowno: "20", title: "เงินสด", rowtype: "account", indent: "2", noteno: "3", amount: "219.75", prioramount: "0.00", showzero: "false" },
        { rowno: "30", title: "", rowtype: "blank" },
        { rowno: "40", title: "รวมสินทรัพย์", rowtype: "subtotal", indent: "0", fontweight: "bold", underline: "double", amount: "219.75", prioramount: "69.75" },
      ],
      totals: {}, totalrows: 4, warnings: [], asof: "", sequence: 1,
    };
    const html = renderToStaticMarkup(createElement(GLStatementTable, { report, company: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", title: "งบแสดงฐานะการเงิน", period: "ณ วันที่ 31 ธันวาคม 2569", showNote: true, tr }));
    for (const text of ["บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", "งบแสดงฐานะการเงิน", "ณ วันที่ 31 ธันวาคม 2569", "(หน่วย : บาท)", ">2569<", ">2568<", ">หมายเหตุ<", ">219.75<", ">69.75<"]) expect(html).toContain(text);
    expect(html).toContain('class="gl-statement-bold gl-statement-u-double">219.75<');
    expect(html).toContain(">-<");
    expect(html).not.toContain("ลำดับ");
  });
});

import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { GLJournal, GLReport, GLStatementCheck } from "@/lib/general-ledger";
import { GLNotesPrint, GLReportWarnings, GLStatementChecks, GLStatementTable, GLVoucherPrint, statementCheckKey, printAmountCell, printOrientation, printPageStyle, statementAmountText, statementOrientation, statementPeriodText, visiblePrintColumns } from "./gl-print";

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
    for (const text of ["บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", "งบแสดงฐานะการเงิน", "ณ วันที่ 31 ธันวาคม 2569", "(หน่วย: บาท)", ">2569<", ">2568<", ">หมายเหตุ<", ">219.75<", ">69.75<"]) expect(html).toContain(text);
    expect(html).toContain('class="gl-statement-bold gl-statement-u-double">219.75<');
    expect(html).toContain(">-<");
    expect(html).not.toContain("ลำดับ");
    expect(statementOrientation(report)).toBe("portrait");
  });

  // งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น: คอลัมน์ = องค์ประกอบ + รวม, ชุดแถวปีก่อน/ปีนี้คั่นด้วยบรรทัดว่าง, พิมพ์แนวนอน
  it("renders equity component columns, year blocks and a landscape page", () => {
    const report: GLReport = {
      columns: [{ key: "rowno", label: "ลำดับ" }, { key: "title", label: "รายการ" }, { key: "noteno", label: "หมายเหตุ" }, { key: "c1", label: "ทุนที่ชำระแล้ว", amount: true }, { key: "c2", label: "ส่วนเกินมูลค่าหุ้น", amount: true }, { key: "c3", label: "กำไร (ขาดทุน) สะสม", amount: true }, { key: "total", label: "รวมส่วนของผู้ถือหุ้น", amount: true }],
      rows: [
        { block: "prioramount", rowno: "140", title: "ยอดคงเหลือ ณ ปลายงวด 2568", rowtype: "account", c1: "1000.00", c2: "0.00", c3: "69.75", total: "1069.75" },
        { block: "amount", rowno: "10", title: "ยอดคงเหลือ ณ ต้นงวด 2569", rowtype: "account", c1: "1000.00", c2: "0.00", c3: "69.75", total: "1069.75" },
      ],
      totals: {}, totalrows: 2, warnings: [], asof: "", sequence: 1,
    };
    const html = renderToStaticMarkup(createElement(GLStatementTable, { report, company: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", title: "งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", period: "สำหรับปีสิ้นสุดวันที่ 31 ธันวาคม 2569", showNote: false, tr }));
    for (const text of ["gl-statement-wide", ">ทุนที่ชำระแล้ว<", ">รวมส่วนของผู้ถือหุ้น<", ">1,069.75<", "ยอดคงเหลือ ณ ต้นงวด 2569"]) expect(html).toContain(text);
    expect(html.match(/<tr[ >]/g)?.length).toBe(5); // หัวงบ + หัวคอลัมน์ + ปีก่อน + บรรทัดคั่น + ปีนี้
    expect(statementOrientation(report)).toBe("landscape");
  });
});

// งบกระแสเงินสด: ผลตรวจเงินสดปลายงวดตามงบกับยอดคงเหลือตามบัญชี (backend) แสดงบนจอนอกตารางงบ — สถานะต้องมีข้อความ ไม่ใช่สีอย่างเดียว
describe("statement book checks", () => {
  const matched: GLStatementCheck = { key: "amount", fiscalyear: "2570", rowno: 180, title: "เงินสดและรายการเทียบเท่าเงินสด ณ วันปลายงวด", statement: "219750.50", book: "219750.50", difference: "0.00", matched: true };
  const mismatched: GLStatementCheck = { key: "prioramount", fiscalyear: "2569", rowno: 180, title: "", statement: "100000.00", book: "69750.25", difference: "30249.75", matched: false };
  it("shows statement and ledger amounts with a matched label", () => {
    const html = renderToStaticMarkup(createElement(GLStatementChecks, { checks: [matched], tr }));
    expect(html).toContain("ตรวจยอดกับบัญชี: เงินสดและรายการเทียบเท่าเงินสด ณ วันปลายงวด · ปีบัญชี 2570");
    expect(html).toContain("ตามงบ 219,750.50 · ตามบัญชี 219,750.50");
    expect(html).toContain("ตรงกัน");
    expect(html).not.toContain("ไม่ตรงกัน");
    expect(html).toContain("<svg");
  });
  it("labels a mismatch with its difference and falls back to the row number when the row has no title", () => {
    const html = renderToStaticMarkup(createElement(GLStatementChecks, { checks: [matched, mismatched], tr }));
    expect(html).toContain("ตรวจยอดกับบัญชี: บรรทัด 180 · ปีบัญชี 2569");
    expect(html).toContain("ไม่ตรงกัน ผลต่าง 30,249.75");
    expect(html.match(/<li/g)).toHaveLength(2);
  });
  it("keys two checks on the same row number apart (row numbers are user-editable and may repeat)", () => {
    const keys = [matched, { ...matched, matched: false }].map((check, index) => statementCheckKey(check, index));
    expect(new Set(keys).size).toBe(2);
    expect(renderToStaticMarkup(createElement(GLStatementChecks, { checks: [matched, { ...matched, matched: false }], tr })).match(/<li/g)).toHaveLength(2);
  });
  it("renders nothing when the report has no checks", () => {
    expect(renderToStaticMarkup(createElement(GLStatementChecks, { checks: undefined, tr }))).toBe("");
    expect(renderToStaticMarkup(createElement(GLStatementChecks, { checks: [], tr }))).toBe("");
  });
  it("is not part of the printed statement table", () => {
    const report: GLReport = { columns: [{ key: "title", label: "รายการ" }, { key: "amount", label: "2570", amount: true }], rows: [{ rowno: "180", title: "เงินสดปลายงวด", rowtype: "subtotal", amount: "219750.50" }], totals: {}, totalrows: 1, warnings: ["คำเตือน"], asof: "", sequence: 0, checks: [mismatched] };
    const html = renderToStaticMarkup(createElement(GLStatementTable, { report, company: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", title: "งบกระแสเงินสด", period: "", showNote: false, tr }));
    expect(html).not.toContain("ตรวจยอดกับบัญชี");
    expect(html).not.toContain("คำเตือน");
  });
});

describe("GLReportWarnings", () => {
  it("lists every warning under a titled notice with an icon", () => {
    const html = renderToStaticMarkup(createElement(GLReportWarnings, { warnings: ["ไม่พบปีบัญชีก่อนหน้า จึงไม่มีคอลัมน์เปรียบเทียบ", " ", "ปี 2569: เงินสดปลายงวดตามงบไม่ตรง"], tr }));
    expect(html).toContain("ข้อควรตรวจสอบก่อนออกงบ");
    expect(html).toContain("<svg");
    expect(html.match(/<li>/g)).toHaveLength(2);
    expect(html).toContain("role=\"status\"");
  });
  it("renders nothing without warnings", () => {
    expect(renderToStaticMarkup(createElement(GLReportWarnings, { warnings: [], tr }))).toBe("");
    expect(renderToStaticMarkup(createElement(GLReportWarnings, { warnings: null, tr }))).toBe("");
  });
});

// หมายเหตุประกอบงบการเงิน (แบบ 2 ข้อ 5) พิมพ์ต่อจากงบ: หัวกิจการ/ชื่อ/งวด แล้วแต่ละข้อ "เลขที่. หัวข้อ" + เนื้อหาที่คงการขึ้นบรรทัด
describe("GLNotesPrint", () => {
  const notes = [
    { id: "n1", noteno: "1", title: "ข้อมูลทั่วไป", body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด\nสำนักงานใหญ่ กรุงเทพมหานคร" },
    { id: "n2", noteno: "2", title: "เกณฑ์ในการจัดทำและนำเสนองบการเงิน", body: "" },
  ];
  const html = renderToStaticMarkup(createElement(GLNotesPrint, { notes, company: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", period: "สำหรับปีสิ้นสุดวันที่ 31 ธันวาคม 2569", tr }));
  it("prints the company, the notes title and the year-ended line in order", () => {
    const company = html.indexOf("gl-print-company"), title = html.indexOf("หมายเหตุประกอบงบการเงิน"), period = html.indexOf("สำหรับปีสิ้นสุดวันที่ 31 ธันวาคม 2569");
    expect(company).toBeGreaterThanOrEqual(0);
    expect(title).toBeGreaterThan(company);
    expect(period).toBeGreaterThan(title);
  });
  it("numbers each note, keeps the heading with its body and preserves line breaks", () => {
    expect(html).toContain(">1. ข้อมูลทั่วไป</div>");
    expect(html).toContain("break-after:avoid");
    expect(html).toContain("white-space:pre-wrap");
    expect(html).toContain("overflow-wrap:anywhere");
    expect(html).toContain("บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด\nสำนักงานใหญ่ กรุงเทพมหานคร");
    expect(html).not.toMatch(/color:\s*#/);
  });
  it("still prints a note whose body is empty", () => {
    expect(html).toContain(">2. เกณฑ์ในการจัดทำและนำเสนองบการเงิน</div>");
    expect(html.match(/gl-print-note-heading/g)).toHaveLength(2);
    expect(html.match(/gl-print-note-body/g)).toHaveLength(1);
  });
});

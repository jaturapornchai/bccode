import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { accountTypes, reportCsv, type GLJournalBook, type GLReport } from "@/lib/general-ledger";
import { ReportGrid } from "./gl-reports";

function render(patch: Partial<GLReport>) {
  const report: GLReport = { sequence: 1, columns: [], rows: [], totals: {}, totalrows: 0, warnings: [], asof: "2026-09-11", ...patch };
  return renderToStaticMarkup(createElement(ReportGrid, { report }));
}

describe("Thai general ledger report presentation", () => {
  it.each([
    ["accounttype", accountTypes],
    ["status", { draft: "ฉบับร่าง", posted: "ผ่านรายการแล้ว", reversed: "กลับรายการแล้ว", void: "ยกเลิกร่าง" }],
    ["direction", { in: "เงินเข้า", out: "เงินออก" }],
    ["category", { operating: "ดำเนินงาน", investing: "ลงทุน", financing: "จัดหาเงิน", unclassified: "ยังไม่ระบุ" }],
  ] as const)("renders %s as Thai labels", (key, labels) => {
    const html = render({ columns: [{ key, label: "รายการ", amount: false }], rows: Object.keys(labels).map(value => ({ [key]: value })) });
    for (const [value, label] of Object.entries(labels)) {
      expect(html).toContain(`>${label}</td>`);
      expect(html).not.toContain(`>${value}</td>`);
    }
  });

  it("renders book codes with the company's journal-book names (unknown codes stay as codes)", () => {
    const books: GLJournalBook[] = [
      { id: "b1", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true },
      { id: "b2", code: "POS1", name: "สมุดขายหน้าร้าน", booktype: 4, isactive: false },
    ];
    const report: GLReport = { sequence: 1, columns: [{ key: "bookcode", label: "สมุด", amount: false }], rows: [{ bookcode: "JV" }, { bookcode: "POS1" }, { bookcode: "ZZ" }], totals: {}, totalrows: 3, warnings: [], asof: "2026-09-11" };
    const html = renderToStaticMarkup(createElement(ReportGrid, { report, books }));
    expect(html).toContain(">สมุดรายวันทั่วไป</td>");
    expect(html).toContain(">สมุดขายหน้าร้าน</td>");
    expect(html).toContain(">ZZ</td>");
  });

  it("labels supplemental financial totals and shows a line count without money decimals", () => {
    const html = render({ totals: { currentearnings: "52000.00000000", cash: "341395.00000000", unclassifiedlines: "0" } });
    expect(html).toContain("กำไรขาดทุนที่ยังไม่ปิด");
    expect(html).toContain("เงินสดและเงินฝากธนาคาร");
    expect(html).toContain("บรรทัดที่ยังไม่ระบุประเภท");
    expect(html).toContain("52,000.00");
    expect(html).toContain("341,395.00");
    expect(html).toContain("0 รายการ");
    expect(html).not.toContain("ยอดรวม");
  });

  it("hides only the computed account key while preserving source data and exact CSV amounts", () => {
    const report: GLReport = { sequence: 1, columns: [{ key: "accountcode", label: "รหัสบัญชี", amount: false }, { key: "accountname", label: "ชื่อบัญชี", amount: false }, { key: "accounttype", label: "หมวดบัญชี", amount: false }, { key: "amount", label: "จำนวนเงิน", amount: true }], rows: [{ accountcode: "__current_earnings__", accountname: "กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม", accounttype: "equity", amount: "0.30000001" }, { accountcode: "BM69-310201", accountname: "กำไรสะสม", accounttype: "equity", amount: "100000.00000000" }], totals: {}, totalrows: 2, warnings: [], asof: "2026-09-11" };
    const before = JSON.stringify(report), csv = reportCsv(report);
    const html = render(report);
    expect(html).toContain(">—</td>");
    expect(html).not.toContain("__current_earnings__");
    expect(html).toContain("BM69-310201");
    expect(html).toContain("กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม");
    expect(html).toContain("0.30000001");
    expect(JSON.stringify(report)).toBe(before);
    expect(reportCsv(report)).toBe(csv);
    expect(csv).toContain("__current_earnings__");
    expect(csv).toContain("0.30000001");
  });

  it("does not translate business codes, descriptions or unknown enum values", () => {
    const html = render({ columns: [{ key: "accountcode", label: "รหัส", amount: false }, { key: "description", label: "รายละเอียด", amount: false }, { key: "status", label: "สถานะ", amount: false }], rows: [{ accountcode: "asset", description: "in", status: "future-status" }] });
    expect(html).toContain(">asset</td>");
    expect(html).toContain(">in</td>");
    expect(html).toContain(">future-status</td>");
  });

  it("renders interactive drill-down buttons when callbacks are provided", () => {
    const report: GLReport = {
      sequence: 1,
      columns: [
        { key: "accountcode", label: "รหัสบัญชี", amount: false },
        { key: "docno", label: "เลขที่เอกสาร", amount: false },
      ],
      rows: [
        { accountcode: "1111-01", docno: "JV202609001", journalid: "journal-1" },
      ],
      totals: {},
      totalrows: 1,
      warnings: [],
      asof: "2026-09-18",
    };

    const html = renderToStaticMarkup(createElement(ReportGrid, {
      report,
      onDrillAccount: () => {},
      onDrillDocNo: () => {},
    }));

    expect(html).toContain("JV202609001");
    expect(html).toContain("1111-01");
    expect(html).toContain("ดูรายงานแยกประเภท");
    expect(html).toContain("รายละเอียดใบสำคัญรายวัน");
  });

  it("renders budget comparison and gl journal totals properly", () => {
    const html = render({
      totals: {
        budgetamount: "500000.00",
        actualamount: "320000.00",
        variance: "180000.00",
      },
    });
    expect(html).toContain("งบประมาณ");
    expect(html).toContain("ยอดจริง");
    expect(html).toContain("ผลต่าง");
    expect(html).toContain("500,000.00");
    expect(html).toContain("320,000.00");
    expect(html).toContain("180,000.00");
  });
});


describe("report source identifiers", () => {
  it("keeps hidden journal IDs out of CSV and never makes a missing ID clickable", () => {
    const report: GLReport = { sequence: 1, columns: [{key: "docno", label: "เลขที่", amount: false}], rows: [{docno: "JV1", journalid: "hidden-id"}, {docno: "JV2"}], totals: {}, totalrows: 2, warnings: [], asof: "2026-09-20" };
    const html = renderToStaticMarkup(createElement(ReportGrid, {report, onDrillDocNo: () => {}}));
    expect((html.match(/title="รายละเอียดใบสำคัญรายวัน"/g) ?? []).length).toBe(1);
    expect(reportCsv(report)).not.toContain("hidden-id");
    expect(reportCsv(report)).toContain("JV2");
  });
});

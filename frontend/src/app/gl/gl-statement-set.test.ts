import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { statementSetRank, statementSetTemplates, type GLReport, type GLStatementTemplate, type StatementType } from "@/lib/general-ledger";
import { GLStatementSetDialog, preparedStatementSet, reportNeedsReview, statementSetPath, type GLStatementSetResult } from "./gl-statement-set";

const tr = (_key: string, fallback: string) => fallback;
const report = (warnings: string[] = []): GLReport => ({ columns: [], rows: [], totals: {}, totalrows: 0, warnings, asof: "2569-12-31", sequence: 7 });

// ยังไม่มีปีบัญชี (holding ใหม่ที่ใช้ GL อย่างเดียว) = ปุ่ม "ตรวจและเตรียมพิมพ์" กดไม่ได้ — ท้ายจอต้องบอกเหตุผล ไม่ใช่ชวนให้กดปุ่มที่กดไม่ได้
describe("GLStatementSetDialog without a fiscal year", () => {
  it("explains that a fiscal year is required instead of asking to press the disabled prepare button", () => {
    const html = renderToStaticMarkup(createElement(GLStatementSetDialog, { open: true, onClose: () => {}, unsavedChanges: false }));
    expect(html).toContain("เลือกปีบัญชีก่อน — ถ้ายังไม่มีปีบัญชี ให้สร้างที่เมนู “ปีบัญชีและบัญชีปิดปี”");
    expect(html).not.toContain("กด “ตรวจและเตรียมพิมพ์”");
    expect(html).toMatch(/<button[^>]*disabled=""[^>]*>(?:(?!<\/button>).)*ตรวจและเตรียมพิมพ์/);
  });
});

// ตรวจและเตรียมพิมพ์ = คำขอเดียว GET reports/statement-set (backend/internal/generalledger/statement_set.go)
describe("statement set request", () => {
  const year = { code: "2569", startdate: "2569-01-01", enddate: "2569-12-31" };

  it("asks for the whole fiscal year with the selected templates and the notes flag", () => {
    const path = statementSetPath(year, ["BS-1", "PL-1"], true);
    expect(path.startsWith("reports/statement-set?")).toBe(true);
    expect(Object.fromEntries(new URLSearchParams(path.split("?")[1]))).toEqual({ fiscalyear: "2569", from: "2569-01-01", to: "2569-12-31", templates: "BS-1,PL-1", notes: "true" });
  });

  it("sends an empty template list for a notes-only print (omitting it would ask the backend for the default set)", () => {
    const query = new URLSearchParams(statementSetPath(year, [], true).split("?")[1]);
    expect(query.has("templates")).toBe(true);
    expect(query.get("templates")).toBe("");
    expect(new URLSearchParams(statementSetPath(year, ["CF-1"], false).split("?")[1]).get("notes")).toBe("false");
  });

  it("keeps the backend order, display settings and per-statement errors without re-sorting in the browser", () => {
    const result: GLStatementSetResult = {
      sequence: 7, fiscalyear: "2569", from: "2569-01-01", to: "2569-12-31",
      sections: [
        { code: "Z-BS", name: "งบแสดงฐานะการเงิน", statementtype: "balance_sheet", shownotecolumn: true, scale: 2, report: report(["หมายเหตุ 9 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2569"]) },
        { code: "A-PL", name: "", statementtype: "pnl", shownotecolumn: false, scale: 0, report: report() },
        { code: "EQ-1", name: "งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", statementtype: "equity", shownotecolumn: true, scale: 2, error: "กรุณากำหนดคอลัมน์องค์ประกอบส่วนของผู้ถือหุ้นและเลือกบัญชีอย่างน้อย 1 คอลัมน์" },
        { code: "CF-1", name: "งบกระแสเงินสด", statementtype: "cash_flow", shownotecolumn: true, scale: 2, report: null },
      ],
      notes: [{ id: "n1", noteno: "1", title: "ข้อมูลทั่วไป", body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" }],
    };
    const prepared = preparedStatementSet(result, "2569", true, tr);
    expect(prepared.sections.map((section) => section.code)).toEqual(["Z-BS", "A-PL", "EQ-1", "CF-1"]);
    expect(prepared.sections[0]).toMatchObject({ title: "งบแสดงฐานะการเงิน", statementtype: "balance_sheet", showNote: true, scale: 2, error: "" });
    // ชื่อว่างใช้รหัส; ทศนิยม/คอลัมน์หมายเหตุเป็นค่าที่ backend ใช้คำนวณ (ไม่เดาใหม่ที่ browser)
    expect(prepared.sections[1]).toMatchObject({ title: "A-PL", showNote: false, scale: 0 });
    expect(prepared.sections[2]).toMatchObject({ report: null, error: "กรุณากำหนดคอลัมน์องค์ประกอบส่วนของผู้ถือหุ้นและเลือกบัญชีอย่างน้อย 1 คอลัมน์" });
    expect(prepared.sections[3]).toMatchObject({ report: null, error: "คำนวณไม่สำเร็จ" });
    expect(prepared.notes).toEqual({ items: result.notes, error: "" });
    expect(prepared.error).toBe("");
  });

  it("flags requested notes that the fiscal year does not have, and leaves notes out when not requested", () => {
    const empty: GLStatementSetResult = { sequence: 1, fiscalyear: "2569", from: "2569-01-01", to: "2569-12-31", sections: null, notes: null };
    expect(preparedStatementSet(empty, "2569", true, tr).notes).toEqual({ items: [], error: "ปีบัญชีนี้ยังไม่มีหมายเหตุประกอบงบการเงิน — เขียนได้ที่โหมด “หมายเหตุประกอบงบการเงิน”" });
    expect(preparedStatementSet(empty, "2569", false, tr)).toEqual({ year: "2569", sections: [], notes: null, error: "" });
  });
});

// ลำดับแบบ 2 มีสองที่: statementSetOrder ใน backend (ลำดับที่พิมพ์) และ statementSetRank ใน lib/general-ledger.ts (ลำดับรายการให้เลือก) — ต้องตรงกัน
describe("statement set order parity with the backend", () => {
  it("statementSetOrder in statement_set.go matches the checklist order", () => {
    const source = readFileSync(resolve(process.cwd(), "..", "backend", "internal", "generalledger", "statement_set.go"), "utf8");
    const match = /^var statementSetOrder = \[\]string\{([^}]*)\}/m.exec(source);
    expect(match, "var statementSetOrder = []string{...} not found in statement_set.go").not.toBeNull();
    const backendOrder = [...match![1].matchAll(/"([a-z_]+)"/g)].map((item) => item[1]);
    expect(backendOrder).toEqual(["balance_sheet", "pnl", "equity", "cash_flow"]);
    expect(backendOrder.map((type) => statementSetRank(type))).toEqual(backendOrder.map((_, index) => index));
    for (const other of ["production_cost", "custom"]) expect(statementSetRank(other)).toBe(backendOrder.length);
    const template = (code: string, statementtype: StatementType): GLStatementTemplate => ({ id: code, version: 1, code, name: code, statementtype, isactive: true, globalstyle: {}, rows: [] } as unknown as GLStatementTemplate);
    const checklist = statementSetTemplates([template("ZZ", "custom"), template("CF", "cash_flow"), template("EQ", "equity"), template("PL", "pnl"), template("BS", "balance_sheet"), template("PC", "production_cost")]);
    expect(checklist.map((item) => item.statementtype)).toEqual([...backendOrder, "production_cost", "custom"]);
  });
});

// งบที่ต้องตรวจ = มีคำเตือน, ผลตรวจความถูกต้องไม่ตรง หรือมีบัญชีที่มียอดแต่ไม่อยู่ในบรรทัดใด (report.unassigned)
describe("reportNeedsReview", () => {
  it("flags reports with accounts left out of every line", () => {
    expect(reportNeedsReview(report())).toBe(false);
    expect(reportNeedsReview({ ...report(), unassigned: [] })).toBe(false);
    expect(reportNeedsReview({ ...report(), unassigned: [{ key: "amount", fiscalyear: "2569", accountcode: "11140", accountname: "เงินฝากประจำ", accounttype: "asset", basis: "closing", amount: "10.00" }] })).toBe(true);
    expect(reportNeedsReview(report(["ยอดไม่ตรง"]))).toBe(true);
  });
});

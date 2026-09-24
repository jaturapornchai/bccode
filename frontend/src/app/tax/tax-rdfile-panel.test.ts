import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { TaxRdFileError, TaxRdFileIssue } from "@/lib/tax-forms";
import { RDFILE_ISSUE_ROWS, rdIssueText, TaxRdFileButton, TaxRdFileIssues, TaxRdFilePanel } from "./tax-rdfile-panel";

// ปุ่ม/แผงไฟล์ยื่นกรมสรรพากร: ไฟล์สร้างจากฉบับที่บันทึกเท่านั้น → ปิดเมื่อยังไม่บันทึก/มีค่าที่แก้ค้าง/จอทำงานอื่นอยู่

// buttonTag - แท็กเปิดของปุ่มที่มีข้อความนี้ (ใช้ตรวจ disabled/aria)
function buttonTag(html: string, text: string): string {
  for (const match of html.matchAll(/<button([^>]*)>([\s\S]*?)<\/button>/g)) {
    if (match[2].includes(text)) return match[1];
  }
  throw new Error(`button not found: ${text}`);
}

const button = (props: Partial<Parameters<typeof TaxRdFileButton>[0]> = {}) =>
  renderToStaticMarkup(createElement(TaxRdFileButton, { ready: true, working: false, busy: false, open: false, onToggle: () => {}, ...props }));

const panel = (props: Partial<Parameters<typeof TaxRdFilePanel>[0]> = {}) =>
  renderToStaticMarkup(
    createElement(TaxRdFilePanel, {
      filing: { id: 7, version: 3 },
      ready: true,
      working: false,
      branchNo: "00000",
      fieldLabel: (key: string) => key,
      onCreate: vi.fn(),
      onIssue: vi.fn(),
      onClose: vi.fn(),
      ...props,
    }),
  );

describe("TaxRdFileButton — ปิดจนกว่าฉบับจะบันทึกและไม่มีค่าที่แก้ค้าง", () => {
  it("ยังไม่บันทึก/แก้ค้าง → ปิด + บอกเหตุผล (title + ข้อความสำหรับโปรแกรมอ่านหน้าจอ)", () => {
    const html = button({ ready: false });
    expect(buttonTag(html, "ไฟล์ยื่นกรมสรรพากร (.txt)")).toContain('disabled=""');
    expect(buttonTag(html, "ไฟล์ยื่นกรมสรรพากร (.txt)")).toContain('aria-describedby="tax-rdfile-need-save"');
    expect(html).toContain('title="บันทึกฉบับนี้ก่อน ไฟล์สร้างจากฉบับที่บันทึกแล้ว"');
    expect(html).toContain('<span id="tax-rdfile-need-save" class="sr-only">บันทึกฉบับนี้ก่อน ไฟล์สร้างจากฉบับที่บันทึกแล้ว</span>');
  });

  it("บันทึกแล้วและไม่มีค่าที่แก้ค้าง → กดได้", () => {
    const tag = buttonTag(button(), "ไฟล์ยื่นกรมสรรพากร (.txt)");
    expect(tag).not.toContain('disabled=""');
    expect(tag).not.toContain("aria-describedby");
  });

  it("จอกำลังทำงานอื่น → ปิด; กำลังสร้างไฟล์ → หมุน; เปิดแผงอยู่ → aria-expanded + aria-controls", () => {
    expect(buttonTag(button({ working: true }), "ไฟล์ยื่นกรมสรรพากร (.txt)")).toContain('disabled=""');
    expect(button({ busy: true })).toContain("animate-spin");
    const open = buttonTag(button({ open: true }), "ไฟล์ยื่นกรมสรรพากร (.txt)");
    expect(open).toContain('aria-expanded="true"');
    expect(open).toContain('aria-controls="tax-rdfile-panel"');
  });
});

describe("TaxRdFilePanel — แผงตั้งค่าไฟล์", () => {
  it("คำอธิบาย: สำหรับ SWC-UI ต้องลงทะเบียน และอัปโหลดเข้า e-Filing ตรง ๆ ไม่ได้ + บอกรุ่นของฉบับที่ใช้สร้าง", () => {
    const html = panel();
    expect(html).toContain("SWC-UI");
    expect(html).toContain("ต้องลงทะเบียนบริการก่อน");
    expect(html).toContain("อัปโหลดเข้า e-Filing ตรง ๆ ไม่ได้");
    expect(html).toContain("สร้างจากฉบับที่บันทึก รุ่นที่ 3");
  });

  it("สาขา 00000 → เติมชื่อ สำนักงานใหญ่; สาขาอื่น → ว่างให้กรอกเอง; ครั้งที่ส่งเริ่มที่ 00", () => {
    expect(panel()).toMatch(/id="tax-rdfile-dept_name"[^>]*value="สำนักงานใหญ่"/);
    expect(panel({ branchNo: "00001" })).toMatch(/id="tax-rdfile-dept_name"[^>]*value=""/);
    expect(panel()).toMatch(/id="tax-rdfile-submission_no"[^>]*value="00"/);
  });

  it("ตัวเลือก LTO และประเภทสาขาเริ่มที่ ไม่ระบุ และมีครบ 3 ตัวเลือก", () => {
    const html = panel();
    for (const label of ["ไม่เป็น", "เป็น", "สาขาภาษีมูลค่าเพิ่ม (V)", "สาขาภาษีธุรกิจเฉพาะ (S)"]) expect(html).toContain(label);
    expect(html.match(/type="radio"[^>]*checked=""[^>]*value=""/g)?.length).toBe(2);
  });

  it("ปุ่มสร้างไฟล์: ปิดเมื่อยังไม่บันทึก (พร้อมข้อความบอก) หรือจอทำงานอื่น; เปิดเมื่อพร้อม", () => {
    const notReady = panel({ ready: false });
    expect(buttonTag(notReady, "สร้างไฟล์")).toContain('disabled=""');
    expect(notReady).toContain('role="status"');
    expect(notReady).toContain("บันทึกฉบับนี้ก่อน ไฟล์สร้างจากฉบับที่บันทึกแล้ว");
    expect(buttonTag(panel({ working: true }), "สร้างไฟล์")).toContain('disabled=""');
    const ready = panel();
    expect(buttonTag(ready, "สร้างไฟล์")).not.toContain('disabled=""');
    expect(ready).not.toContain("บันทึกฉบับนี้ก่อน");
  });
});

const issue = (overrides: Partial<TaxRdFileIssue> = {}): TaxRdFileIssue => ({ key: "tax_rdfile_title_required", field: "title", row: 1, ...overrides });
const failure = (issues: TaxRdFileIssue[], total = issues.length): TaxRdFileError => ({ code: "tax_rdfile_invalid", issues, total });

describe("TaxRdFileIssues — รายการจุดที่ต้องแก้", () => {
  it("แสดงไม่เกิน 50 รายการ + บอกจำนวนที่เหลือจาก total ของ backend", () => {
    const list = Array.from({ length: 60 }, (_, i) => issue({ row: i + 1, message: "ต้องมีคำนำหน้าชื่อ" }));
    const html = renderToStaticMarkup(createElement(TaxRdFileIssues, { error: failure(list, 120), fieldLabel: () => "คำนำหน้าชื่อ", onSelect: vi.fn() }));
    expect(html.match(/<li /g)?.length).toBe(RDFILE_ISSUE_ROWS);
    expect(html).toContain("สร้างไฟล์ไม่ได้ — พบ 120 จุดที่ต้องแก้");
    expect(html).toContain("และอีก 70 รายการ");
    expect(html).toContain("แถวที่ 1: ต้องมีคำนำหน้าชื่อ — ช่อง “คำนำหน้าชื่อ”");
  });

  it("รายการที่ชี้ช่องได้เป็นปุ่ม; ไม่มีช่อง (ทั้งไฟล์) เป็นข้อความ; ช่องในแผง (ชื่อแผนก) กดได้แม้ไม่มี field", () => {
    const html = renderToStaticMarkup(
      createElement(TaxRdFileIssues, {
        error: failure([
          issue(),
          issue({ key: "tax_rdfile_no_rows", field: undefined, row: 0, message: "ไม่มีรายการในใบแนบ" }),
          issue({ key: "tax_rdfile_dept_name_required", field: undefined, row: 0, message: "กรอกชื่อแผนก" }),
        ]),
        fieldLabel: (key: string) => (key === "title" ? "คำนำหน้าชื่อ" : key),
        onSelect: vi.fn(),
      }),
    );
    expect(html.match(/<button /g)?.length).toBe(2);
    expect(html).toMatch(/<p [^>]*><span [^>]*>•<\/span><span>ไม่มีรายการในใบแนบ<\/span><\/p>/);
    expect(html).not.toContain("และอีก");
  });

  // ช่องที่แบบยังไม่มีคอลัมน์ (ป้าย = รหัสดิบ) พาไปไม่ได้ → ข้อความธรรมดา ไม่ใช่ปุ่มที่กดแล้วไม่เกิดอะไร
  it("ช่องที่ไม่มีบนแบบ/ใบแนบ → ไม่เป็นปุ่ม", () => {
    const html = renderToStaticMarkup(
      createElement(TaxRdFileIssues, { error: failure([issue({ field: "addr_district", key: "tax_rdfile_address_required" })]), fieldLabel: (key: string) => key, onSelect: vi.fn() }),
    );
    expect(html).not.toContain("<button");
  });

  it("ข้อผิดพลาดที่ไม่มีรายการ (เช่น ชนเวอร์ชัน) → ข้อความเดียว", () => {
    const html = renderToStaticMarkup(
      createElement(TaxRdFileIssues, { error: { code: "tax_form_version_conflict", message: "มีผู้อื่นแก้ฉบับนี้ไปแล้ว", issues: [], total: 0 }, fieldLabel: (k: string) => k, onSelect: vi.fn() }),
    );
    expect(html).toContain('role="alert"');
    expect(html).toContain("มีผู้อื่นแก้ฉบับนี้ไปแล้ว");
    expect(html).not.toContain("<li");
  });

  it("ค่าในฉบับที่บันทึกผิดรูปแบบ (ไม่มีรายการ แต่มีช่อง+แถว) → บอกชื่อช่องและแถว", () => {
    const html = renderToStaticMarkup(
      createElement(TaxRdFileIssues, {
        error: { code: "tax_form_value_invalid", message: "ค่าไม่ถูกต้อง", field: "l1_amount", row: 2, issues: [], total: 0 },
        fieldLabel: (k: string) => (k === "l1_amount" ? "จำนวนเงินที่จ่าย" : k),
        onSelect: vi.fn(),
      }),
    );
    expect(html).toContain("ค่าไม่ถูกต้อง — ช่อง “จำนวนเงินที่จ่าย” (แถวที่ 2)");
  });
});

describe("rdIssueText — ข้อความของจุดที่ต้องแก้", () => {
  const dict: Record<string, string> = {
    tax_rdfile_forbidden_char: "มีอักขระต้องห้าม “{char}”",
    tax_rdfile_too_long: "ยาวเกิน {max} ตัวอักษร",
  };
  const tr = (key: string, fallback: string) => dict[key] ?? fallback;
  const label = (key: string) => ({ name: "ชื่อผู้มีเงินได้", media_ref_no: "เลขอ้างอิงการลงทะเบียน" })[key] ?? key;

  it("แถวในใบแนบ: แถวที่ n + ข้อความจากพจนานุกรม + ค่า {char} + ชื่อช่อง", () => {
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_forbidden_char", field: "name", row: 3, args: { char: "&" } }), label)).toBe("แถวที่ 3: มีอักขระต้องห้าม “&” — ช่อง “ชื่อผู้มีเงินได้”");
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_too_long", field: "name", row: 2, args: { max: "100" } }), label)).toBe("แถวที่ 2: ยาวเกิน 100 ตัวอักษร — ช่อง “ชื่อผู้มีเงินได้”");
  });

  it("ค่า $& / $1 แทนตรงตัว ไม่ถูกตีความเป็นรูปแบบพิเศษ", () => {
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_forbidden_char", field: undefined, row: 0, args: { char: "$&" } }), label)).toBe("มีอักขระต้องห้าม “$&”");
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_x", field: undefined, row: 4, message: "ยอด $1 ผิด" }), label)).toBe("แถวที่ 4: ยอด $1 ผิด");
  });

  it("หัวแบบ (row 0) ไม่มีคำว่าแถว; ไม่มีใน dictionary → ข้อความจาก backend → ข้อความกลาง", () => {
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_user_id_required", field: "media_ref_no", row: 0, message: "ต้องกรอกเลขอ้างอิงการลงทะเบียน" }), label)).toBe(
      "ต้องกรอกเลขอ้างอิงการลงทะเบียน — ช่อง “เลขอ้างอิงการลงทะเบียน”",
    );
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_new", field: undefined, row: 0 }), label)).toBe("ข้อมูลไม่ตรงรูปแบบไฟล์ของกรมสรรพากร");
  });

  it("ป้ายช่องตามตำแหน่ง: แถวใบแนบใช้ป้ายคอลัมน์ (name ของผู้มีเงินได้) ไม่ใช่ป้ายบนหัวแบบ", () => {
    const byRow = (key: string, row: number) => (key === "name" ? (row > 0 ? "ชื่อผู้มีเงินได้" : "ชื่อผู้มีหน้าที่หักภาษี") : key);
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_name_required", field: "name", row: 2, message: "ต้องกรอกชื่อ" }), byRow)).toBe("แถวที่ 2: ต้องกรอกชื่อ — ช่อง “ชื่อผู้มีเงินได้”");
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_name_required", field: "name", row: 0, message: "ต้องกรอกชื่อ" }), byRow)).toBe("ต้องกรอกชื่อ — ช่อง “ชื่อผู้มีหน้าที่หักภาษี”");
  });

  it("ช่องที่ไม่รู้จัก (ป้าย = รหัสช่อง) ไม่แสดงรหัสเทคนิค", () => {
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_unsupported_form", field: "code", row: 0, message: "แบบนี้ยังสร้างไฟล์ไม่ได้" }), label)).toBe("แบบนี้ยังสร้างไฟล์ไม่ได้");
  });

  it("backend ไม่ส่ง args → ใช้ข้อความที่ backend แทนค่าแล้วแทนแม่แบบที่ยังมี {char}", () => {
    expect(rdIssueText(tr, issue({ key: "tax_rdfile_forbidden_char", field: undefined, row: 5, message: "มีอักขระต้องห้าม “/”" }), label)).toBe("แถวที่ 5: มีอักขระต้องห้าม “/”");
  });
});

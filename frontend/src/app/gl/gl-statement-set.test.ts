import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { GLStatementSetDialog } from "./gl-statement-set";

// ยังไม่มีปีบัญชี (holding ใหม่ที่ใช้ GL อย่างเดียว) = ปุ่ม "ตรวจและเตรียมพิมพ์" กดไม่ได้ — ท้ายจอต้องบอกเหตุผล ไม่ใช่ชวนให้กดปุ่มที่กดไม่ได้
describe("GLStatementSetDialog without a fiscal year", () => {
  it("explains that a fiscal year is required instead of asking to press the disabled prepare button", () => {
    const html = renderToStaticMarkup(createElement(GLStatementSetDialog, { open: true, onClose: () => {}, unsavedChanges: false }));
    expect(html).toContain("เลือกปีบัญชีก่อน — ถ้ายังไม่มีปีบัญชี ให้สร้างที่เมนู “ปีบัญชีและบัญชีปิดปี”");
    expect(html).not.toContain("กด “ตรวจและเตรียมพิมพ์”");
    expect(html).toMatch(/<button[^>]*disabled=""[^>]*>(?:(?!<\/button>).)*ตรวจและเตรียมพิมพ์/);
  });
});

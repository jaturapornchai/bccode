import { afterEach, describe, expect, it, vi } from "vitest";
import {
  acceptsTyping,
  cleanDocument,
  computeTaxForm,
  groupFields,
  noteText,
  prefillTaxForm,
  requestTaxFormPdf,
  saveTaxForm,
  taxFormRouteCode,
  toDocument,
  type TaxFormField,
} from "./tax-forms";
import { setupTestAuthSession } from "./test-auth-session";

setupTestAuthSession();

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

const scope = { holdingcode: "rungrueng", businesscode: "01" };

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: new Headers({ "content-type": "application/json" }),
    json: async () => body,
  } as unknown as Response;
}

describe("tax form typing guard", () => {
  it("ช่องเงินรับตัวเลข คอมมา และทศนิยมไม่เกิน 2 ตำแหน่ง", () => {
    for (const ok of ["", "1,250", "1250.5", "0.01", "107,000.00"]) expect(acceptsTyping("money", ok), ok).toBe(true);
    for (const bad of ["1.234", "-5", "1e5", "abc", "12.3.4"]) expect(acceptsTyping("money", bad), bad).toBe(false);
  });

  it("เลขประจำตัวผู้เสียภาษี/เลขหลัก รับตัวเลข ช่องว่าง และขีด", () => {
    expect(acceptsTyping("taxid", "0-1055-58012-34-9")).toBe(true);
    expect(acceptsTyping("digits", "12a")).toBe(false);
    expect(acceptsTyping("int", "12")).toBe(true);
    expect(acceptsTyping("int", "1.5")).toBe(false);
  });

  it("ข้อความห้ามขึ้นบรรทัดใหม่และยาวเกิน 300 ตัว", () => {
    expect(acceptsTyping("text", "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด")).toBe(true);
    expect(acceptsTyping("text", "a\nb")).toBe(false);
    expect(acceptsTyping("text", "x".repeat(301))).toBe(false);
  });
});

describe("tax form document helpers", () => {
  it("toDocument ทิ้งค่าที่ไม่ใช่ string (เงินไม่เป็น JSON number)", () => {
    const doc = toDocument({ values: { output_tax: "6930.00", bad: 6930, flag: true }, rows: [{ name: "ก", amount: 1 }], sheets: "x" });
    expect(doc).toEqual({ values: { output_tax: "6930.00" }, rows: [{ name: "ก" }], sheets: undefined });
    expect(toDocument(null)).toEqual({ values: {}, rows: undefined, sheets: undefined });
  });

  it("cleanDocument ตัดช่องว่างและแถวที่ว่างทั้งแถว", () => {
    const doc = cleanDocument({ values: { a: " 1 ", b: "  " }, rows: [{ x: " " }, { x: "2" }], sheets: [{ y: "" }] });
    expect(doc).toEqual({ values: { a: "1" }, rows: [{ x: "2" }], sheets: undefined });
  });

  it("groupFields เรียงหมวดตามลำดับที่ปรากฏบนแบบ ช่องไม่มีหมวดไปอยู่ other", () => {
    const f = (key: string, group: string) => ({ key, label: key, type: "text", group }) as TaxFormField;
    const groups = groupFields([f("a", "payer"), f("b", "tax"), f("c", "payer"), f("d", "")]);
    expect(groups.map((g) => [g.group, g.fields.map((x) => x.key)])).toEqual([["payer", ["a", "c"]], ["tax", ["b"]], ["other", ["d"]]]);
  });

  it("noteText แทน {count} และ {amount}", () => {
    expect(noteText("พบ {count} รายการ", { key: "k", count: 3 }, (v) => v)).toBe("พบ 3 รายการ");
    expect(noteText("ยกมา {amount} บาท", { key: "k", amount: "400.00" }, (v) => `#${v}`)).toBe("ยกมา #400.00 บาท");
  });

  it("route เมนูแบบยื่น → รหัสแบบ; เมนูอื่นไม่ใช่จอแบบยื่น", () => {
    expect(taxFormRouteCode("/report/vatpp30")).toBe("pp30");
    expect(taxFormRouteCode("/report/vatpnd53?x=1")).toBe("pnd53");
    expect(taxFormRouteCode("/report/taxforms")).toBe("");
    expect(taxFormRouteCode("/report/vatsale")).toBeUndefined();
  });
});

describe("tax form API client", () => {
  it("prefill ส่ง scope + งวด และคืน notes", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, {
      success: true,
      data: { values: { output_tax: "6930.00" } },
      notes: [{ key: "tax_form_note_vat_records", count: 3 }, { bad: 1 }],
    }));
    vi.stubGlobal("fetch", fetchMock);
    const result = await prefillTaxForm(scope, { code: "pp30", year: 2026, month: 9 });
    expect(result).toEqual({ ok: true, data: { values: { output_tax: "6930.00" }, rows: undefined, sheets: undefined }, notes: [{ key: "tax_form_note_vat_records", count: 3, amount: undefined }] });
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/goapi/api/report/tax/form/prefill");
    expect(JSON.parse(String(init.body))).toMatchObject({ holdingcode: "rungrueng", businesscode: "01", code: "pp30", year: 2026, month: 9 });
  });

  it("compute คืนหมายเหตุที่ตรวจใหม่ + key ที่ต้องแทน (แก้เลขผู้เสียภาษีแล้วหมายเหตุหายได้)", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      success: true, data: { values: { total_income: "1000.00" }, rows: [] }, notes: [], rechecked: ["tax_form_note_missing_taxid", 7],
    })));
    const result = await computeTaxForm(scope, "pnd53", { values: {}, rows: [] });
    expect(result).toEqual({ ok: true, data: { values: { total_income: "1000.00" }, rows: [] }, notes: [], rechecked: ["tax_form_note_missing_taxid"] });
  });

  it("บันทึกชนเวอร์ชันคืน code + field จาก backend", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(409, { success: false, code: "tax_form_version_conflict", field: "x", row: 2 })));
    const result = await saveTaxForm(scope, { code: "pnd53", year: 2026, month: 9 }, { values: {} });
    expect(result).toEqual({ ok: false, error: { code: "tax_form_version_conflict", message: undefined, field: "x", row: 2 } });
  });

  it("PDF: content-type application/pdf คืน Blob; เครือข่ายล้มคืน connection_error", async () => {
    const blob = new Blob(["%PDF-1.7"], { type: "application/pdf" });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200, headers: new Headers({ "content-type": "application/pdf" }), blob: async () => blob } as unknown as Response));
    const pdf = await requestTaxFormPdf(scope, { code: "pnd53", year: 2026, month: 9 }, { values: {} });
    expect(pdf.ok && pdf.data.size).toBe(8);

    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    expect(await requestTaxFormPdf(scope, { code: "pnd53", year: 2026, month: 9 }, { values: {} })).toEqual({ ok: false, error: { code: "connection_error" } });
  });
});

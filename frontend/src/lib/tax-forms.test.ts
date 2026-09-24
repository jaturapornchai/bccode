import { afterEach, describe, expect, it, vi } from "vitest";
import {
  acceptsTyping,
  attachmentChanged,
  cleanDocument,
  computeTaxForm,
  contentDispositionFilename,
  fetchTaxFormCatalog,
  formChanges,
  groupFields,
  isHeadOfficeBranch,
  ledgerBaseline,
  ledgerDrift,
  normalizeMoneyText,
  normalizeSubmissionNo,
  noteText,
  prefillTaxForm,
  RD_FILE_FALLBACK_NAME,
  requestTaxFormPdf,
  requestTaxFormRdFile,
  saveTaxForm,
  taxFormRouteCode,
  toDocument,
  type TaxFormField,
  type TaxFormSchema,
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

  it("normalizeMoneyText เทียบยอดเงินเป็นข้อความ: จุลภาค/ศูนย์ท้าย/ค่าว่าง ไม่ทำให้ต่างกัน", () => {
    expect(normalizeMoneyText("6,561.40")).toBe("6561.4");
    expect(normalizeMoneyText("006561.400")).toBe("6561.4");
    expect(normalizeMoneyText("")).toBe("0");
    expect(normalizeMoneyText("0.00")).toBe("0");
    expect(normalizeMoneyText("-0.00")).toBe("0");
    expect(normalizeMoneyText("-12.50")).toBe("-12.5");
    // ทศนิยมยาวต้องไม่ถูกปัดแบบ float
    expect(normalizeMoneyText("0.1000000000000000055")).toBe("0.1000000000000000055");
  });

  // UAT V22 2026-09-24: บันทึก ภ.พ.30 ต.ค. (ภาษีขาย 6,561.42) แล้วลงใบขายเพิ่ม → เปิดฉบับเดิมต้องเตือนว่ายอดบัญชีเปลี่ยน
  it("ledgerBaseline + ledgerDrift: ยอดจากบัญชีเปลี่ยนหลังบันทึก → เตือนเฉพาะช่องเงินที่ต่างจริง ช่องกรอกเองไม่ทำให้เตือนผิด", () => {
    const f = (key: string, type: TaxFormField["type"] = "money") => ({ key, label: key, type }) as TaxFormField;
    const schema: TaxFormSchema = {
      code: "pp30",
      title: "ภ.พ.30",
      fields: [f("name", "text"), f("filing_type", "choice"), f("sales_amount"), f("output_tax"), f("excess_brought_forward"), f("tax_payable")],
    };
    const saved = { values: { name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", filing_type: "normal", sales_amount: "93734.50", output_tax: "6561.42", excess_brought_forward: "500.00", tax_payable: "6061.42" } };
    // prefill: ยอดจากบัญชีใหม่ + บรรทัดรวมที่คิดโดยไม่มียอดยกมา (ช่องกรอกเองว่าง)
    const fresh = { values: { name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", filing_type: "normal", sales_amount: "94734.50", output_tax: "6631.42", tax_payable: "6631.42" } };
    const baseline = ledgerBaseline(schema, saved, fresh);
    expect(baseline.values).toMatchObject({ sales_amount: "94734.50", output_tax: "6631.42", excess_brought_forward: "500.00" });
    // compute (backend) คิดบรรทัดรวมใหม่ด้วยยอดยกมาเดิม: 6,631.42 − 500.00
    const ledger = { values: { ...baseline.values, tax_payable: "6131.42" } };
    expect(ledgerDrift(schema, saved, ledger)).toEqual([
      { key: "sales_amount", before: "93734.50", after: "94734.50" },
      { key: "output_tax", before: "6561.42", after: "6631.42" },
      { key: "tax_payable", before: "6061.42", after: "6131.42" },
    ]);
    // บัญชีไม่เปลี่ยน: ไม่เตือน (รูปแบบตัวเลขต่างกันไม่นับ)
    const same = { values: { ...saved.values, sales_amount: "93,734.5" } };
    expect(ledgerDrift(schema, saved, same)).toEqual([]);
  });

  // เตือนผิดเดิม: ผู้ใช้แก้ยอดที่ดึงจากบัญชีเอง (ภาษีขาย 6,561.42 → 6,600.00) แล้วบันทึก → เปิดใหม่ขึ้นว่า "ยอดบัญชีเปลี่ยน" ทั้งที่บัญชีไม่เปลี่ยน
  describe("ยอดที่ผู้ใช้แก้เองต้องไม่ถูกเตือนว่ายอดบัญชีเปลี่ยน", () => {
    const f = (key: string, type: TaxFormField["type"] = "money", options?: TaxFormField["options"]) => ({ key, label: key, type, options }) as TaxFormField;
    const schema: TaxFormSchema = {
      code: "pp30",
      title: "ภ.พ.30",
      fields: [f("filing_type", "choice", [{ value: "normal", label: "ยื่นปกติ" }, { value: "additional", label: "ยื่นเพิ่มเติม" }]), f("sales_amount"), f("output_tax"), f("excess_brought_forward"), f("tax_payable")],
    };
    // prefill ตอนเตรียมฉบับ (ยอดจากบัญชี) และค่าที่บันทึก: ผู้ใช้แก้ภาษีขายเองเป็น 6,600.00
    const basis = { values: { filing_type: "normal", sales_amount: "93734.50", output_tax: "6561.42", tax_payable: "6561.42" } };
    const saved = { values: { ...basis.values, output_tax: "6600.00", tax_payable: "6600.00" } };

    it("มียอดตั้งต้น: เทียบ prefill ตอนนี้กับ prefill ตอนเตรียม — บัญชีไม่เปลี่ยน = ไม่เตือน แม้ค่าที่บันทึกต่างจากบัญชี", () => {
      expect(ledgerDrift(schema, basis, { values: { ...basis.values, sales_amount: "93,734.5" } })).toEqual([]);
      // บัญชีเปลี่ยนจริง → เตือนเฉพาะช่องที่ยอดจากบัญชีเปลี่ยน (ไม่เอาค่าที่ผู้ใช้แก้มาเทียบ)
      const fresh = { values: { ...basis.values, sales_amount: "94734.50", output_tax: "6631.42", tax_payable: "6631.42" } };
      expect(ledgerDrift(schema, basis, fresh)).toEqual([
        { key: "sales_amount", before: "93734.50", after: "94734.50" },
        { key: "output_tax", before: "6561.42", after: "6631.42" },
        { key: "tax_payable", before: "6561.42", after: "6631.42" },
      ]);
    });

    it("ไม่มียอดตั้งต้น: ช่องที่แก้ในรอบนี้คงค่าที่บันทึก จึงไม่ขึ้นว่าต่าง; ช่องอื่นยังเทียบกับบัญชีตามเดิม", () => {
      const fresh = { values: { ...basis.values } };
      const edited = new Set(["output_tax"]);
      const baseline = ledgerBaseline(schema, saved, fresh, edited);
      expect(baseline.values.output_tax).toBe("6600.00");
      expect(baseline.values.sales_amount).toBe("93734.50");
      // compute (backend) คิดบรรทัดรวมจากค่าชุดเดียวกับที่บันทึก → ตรงกัน ไม่เตือน
      expect(ledgerDrift(schema, saved, { values: { ...baseline.values, tax_payable: "6600.00" } })).toEqual([]);
      // ไม่บอกช่องที่แก้ = เทียบทุกช่อง (จอต้องบอกผู้ใช้ว่าความต่างอาจมาจากการแก้เองครั้งก่อน)
      expect(ledgerBaseline(schema, saved, fresh).values.output_tax).toBe("6561.42");
    });

    it("formChanges: ช่องที่จะถูกทับพร้อมค่าเดิม → ค่าใหม่ (รูปแบบตัวเลขต่างไม่นับ, ช่องกรอกเองที่จะหายก็นับ)", () => {
      const screen = { values: { ...saved.values, sales_amount: "93,734.50", excess_brought_forward: "500.00" } };
      expect(formChanges(schema, screen, basis)).toEqual([
        { key: "output_tax", before: "6600.00", after: "6561.42" },
        { key: "excess_brought_forward", before: "500.00", after: "" },
        { key: "tax_payable", before: "6600.00", after: "6561.42" },
      ]);
      expect(formChanges(schema, basis, { values: { ...basis.values } })).toEqual([]);
      expect(formChanges(schema, { values: { filing_type: "additional" } }, { values: { filing_type: " additional " } })).toEqual([]);
    });

    it("attachmentChanged: ลำดับ key/แถวว่างไม่นับ แถวต่างจริงนับ", () => {
      const a = { values: {}, rows: [{ name: "ก", amount: "100.00" }] };
      expect(attachmentChanged(a, { values: {}, rows: [{ amount: "100.00", name: "ก" }, { name: " " }] })).toBe(false);
      expect(attachmentChanged(a, { values: {}, rows: [{ name: "ก", amount: "150.00" }] })).toBe(true);
      expect(attachmentChanged({ values: {} }, { values: {}, rows: [] })).toBe(false);
    });
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

describe("ไฟล์ยื่นกรมสรรพากร (Format กลาง) — API client", () => {
  const filing = { id: 41, version: 3 };
  const options = { dept_name: "  สำนักงานใหญ่ ", submission_no: "5", lto: "" as const, branch_type: "V" as const };
  const fileResponse = (headers: Record<string, string>, bytes: Uint8Array<ArrayBuffer>) =>
    ({ ok: true, status: 200, headers: new Headers(headers), blob: async () => new Blob([bytes], { type: "text/plain" }) }) as unknown as Response;

  it("สำเร็จ: ส่ง id+version+rdfile (ตัดช่องว่าง, ครั้งที่ส่ง 2 หลัก) ได้ Blob ตามไบต์ (BOM คงอยู่) + ชื่อไฟล์จาก Content-Disposition", async () => {
    const bytes = new Uint8Array([0xef, 0xbb, 0xbf, 0x48, 0x7c, 0x30, 0x0d, 0x0a, 0x44]);
    const fetchMock = vi.fn().mockResolvedValue(fileResponse({
      "content-type": "text/plain; charset=utf-8",
      "content-disposition": 'attachment; filename="PND53_0105555555555_000000_2569_08_00_05.txt"',
    }, bytes));
    vi.stubGlobal("fetch", fetchMock);
    const result = await requestTaxFormRdFile(scope, filing, options);
    expect(result.ok).toBe(true);
    if (!result.ok) return;
    expect(result.data.filename).toBe("PND53_0105555555555_000000_2569_08_00_05.txt");
    expect([...new Uint8Array(await result.data.blob.arrayBuffer())]).toEqual([...bytes]);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/api/goapi/api/report/tax/form/rdfile");
    expect(JSON.parse(String(init.body))).toEqual({
      holdingcode: "rungrueng", businesscode: "01", id: 41, version: 3,
      rdfile: { dept_name: "สำนักงานใหญ่", submission_no: "05", lto: "", branch_type: "V" },
    });
  });

  it("ไม่มี Content-Disposition → ชื่อสำรอง", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(fileResponse({ "content-type": "text/plain" }, new Uint8Array([0x48]))));
    const result = await requestTaxFormRdFile(scope, filing, options);
    expect(result.ok && result.data.filename).toBe(RD_FILE_FALLBACK_NAME);
  });

  it("ไม่ผ่าน (BFF ส่ง 200 + success:false): อ่าน issues ทิ้งรายการเสีย, row ไม่ถูกต้อง = 0, total ไม่น้อยกว่าจำนวนรายการ", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      success: false,
      code: "tax_rdfile_invalid",
      message: "ข้อมูลยังสร้างไฟล์ไม่ได้",
      total: 130,
      issues: [
        { key: "tax_rdfile_forbidden_char", field: "name", row: 3, message: "มีอักขระต้องห้าม “&”", args: { char: "&", bad: 1 } },
        { key: "tax_rdfile_too_long", field: "l1_income_type", row: 2, args: { max: "100" } },
        { key: "tax_rdfile_taxid_invalid", field: "tax_id", row: 4, args: [] },
        { key: "tax_rdfile_user_id_required", field: "media_ref_no", row: -1 },
        { key: "tax_rdfile_no_rows", row: 1.5 },
        { field: "x", row: 1 },
        "broken",
      ],
    })));
    const result = await requestTaxFormRdFile(scope, filing, options);
    expect(result).toEqual({
      ok: false,
      error: {
        code: "tax_rdfile_invalid", message: "ข้อมูลยังสร้างไฟล์ไม่ได้", field: undefined, row: undefined, total: 130,
        issues: [
          { key: "tax_rdfile_forbidden_char", field: "name", row: 3, message: "มีอักขระต้องห้าม “&”", args: { char: "&" } },
          { key: "tax_rdfile_too_long", field: "l1_income_type", row: 2, message: undefined, args: { max: "100" } },
          { key: "tax_rdfile_taxid_invalid", field: "tax_id", row: 4, message: undefined, args: undefined },
          { key: "tax_rdfile_user_id_required", field: "media_ref_no", row: 0, message: undefined, args: undefined },
          { key: "tax_rdfile_no_rows", field: undefined, row: 0, message: undefined, args: undefined },
        ],
      },
    });
  });

  it("ไม่ผ่านแบบ 400 ตรงจาก backend + total น้อยกว่ารายการ → ใช้จำนวนรายการ", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(400, {
      success: false, code: "tax_rdfile_invalid", total: 0, issues: [{ key: "tax_rdfile_title_required", field: "title", row: 1 }],
    })));
    const result = await requestTaxFormRdFile(scope, filing, options);
    expect(!result.ok && result.error.total).toBe(1);
    expect(!result.ok && result.error.issues[0]).toMatchObject({ key: "tax_rdfile_title_required", field: "title", row: 1 });
  });

  it("ชนเวอร์ชัน / ไม่มีสิทธิ์ / เครือข่ายล้ม / คำตอบไม่ใช่ JSON", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, { success: false, code: "tax_form_version_conflict" })));
    expect(await requestTaxFormRdFile(scope, filing, options)).toEqual({
      ok: false, error: { code: "tax_form_version_conflict", message: undefined, field: undefined, row: undefined, issues: [], total: 0 },
    });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(403, { success: false })));
    expect(await requestTaxFormRdFile(scope, filing, options)).toEqual({ ok: false, error: { code: "unauthorized", issues: [], total: 0 } });
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    expect(await requestTaxFormRdFile(scope, filing, options)).toEqual({ ok: false, error: { code: "connection_error", issues: [], total: 0 } });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false, status: 502, headers: new Headers({ "content-type": "text/html" }), json: async () => { throw new SyntaxError("Unexpected token <"); },
    } as unknown as Response));
    const html = await requestTaxFormRdFile(scope, filing, options);
    expect(!html.ok && html.error.code).toBe("load_failed");
  });

  it("text/plain ที่ไม่ใช่ 2xx ไม่ถือเป็นไฟล์", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false, status: 500, headers: new Headers({ "content-type": "text/plain" }), json: async () => { throw new SyntaxError("x"); },
    } as unknown as Response));
    expect((await requestTaxFormRdFile(scope, filing, options)).ok).toBe(false);
  });

  it("catalog อ่านธง rdfile (ไม่มี = ไม่รองรับ, ต้องเป็น boolean true เท่านั้น)", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      success: true,
      data: [
        { code: "pnd53", title: "ภ.ง.ด.53", period: "month", source: "wht", hasattachment: true, rdfile: true },
        { code: "pp30", title: "ภ.พ.30", period: "month", source: "vat", hasattachment: true },
        { code: "pnd3", title: "ภ.ง.ด.3", period: "month", source: "wht", rdfile: "true" },
      ],
    })));
    const result = await fetchTaxFormCatalog(scope);
    expect(result.ok && result.data.map((c) => [c.code, c.rdfile])).toEqual([["pnd53", true], ["pp30", false], ["pnd3", false]]);
  });
});

describe("ไฟล์ยื่นกรมสรรพากร — ตัวช่วย", () => {
  it("contentDispositionFilename: filename* ก่อน, filename=\"…\", token, ตัด path/อักขระควบคุม", () => {
    expect(contentDispositionFilename(null)).toBe("");
    expect(contentDispositionFilename("attachment")).toBe("");
    expect(contentDispositionFilename('attachment; filename="PND3_0105555555555_000001_2569_09_00_00.txt"')).toBe("PND3_0105555555555_000001_2569_09_00_00.txt");
    expect(contentDispositionFilename("attachment; filename=PND2_x.txt; size=10")).toBe("PND2_x.txt");
    expect(contentDispositionFilename("attachment; filename=\"fallback.txt\"; filename*=UTF-8''%E0%B8%A0.txt")).toBe("ภ.txt");
    // filename* เสีย (ถอดรหัสไม่ได้) → ใช้ filename ธรรมดา
    expect(contentDispositionFilename("attachment; filename*=UTF-8''%E0%B8; filename=\"plain.txt\"")).toBe("plain.txt");
    expect(contentDispositionFilename('attachment; filename="..\\\\..\\\\evil/a\u0007b.txt"')).toBe("ab.txt");
  });

  it("normalizeSubmissionNo: ว่าง=00, หลักเดียวเติม 0, 2 หลักคงเดิม, อื่น=null", () => {
    expect(normalizeSubmissionNo("")).toBe("00");
    expect(normalizeSubmissionNo(" 7 ")).toBe("07");
    expect(normalizeSubmissionNo("42")).toBe("42");
    expect(normalizeSubmissionNo("100")).toBeNull();
    expect(normalizeSubmissionNo("1a")).toBeNull();
  });

  it("isHeadOfficeBranch: ศูนย์ล้วนเท่านั้น", () => {
    expect(isHeadOfficeBranch("00000")).toBe(true);
    expect(isHeadOfficeBranch("000000")).toBe(true);
    expect(isHeadOfficeBranch("00001")).toBe(false);
    expect(isHeadOfficeBranch("")).toBe(false);
    expect(isHeadOfficeBranch(undefined)).toBe(false);
  });
});

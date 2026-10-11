import { describe, expect, it } from "vitest";
import { DIMENSIONS, dimensionDef, dimensionOptions, emptyJournal, emptyLine, type GLAccount, type GLDimension, type GLDimensionSettingsInput, type GLJournal, type GLLine } from "./general-ledger";
import type { GLDetailVat } from "./gl-journal-details";
import {
  EXTRA_DIMENSIONS, JOURNAL_SAVE_OPTIONS, dropBlankTaxBoxRow, extraDimensionCodes, isBlankVatRow, journalDimensionColumns, journalFunctionKey, journalSaveOptionStorageKey, lineDimensionCell, linesHaveAmounts,
  newVatRow, nextJournalAfterSave, parseJournalSaveOption, patchTaxBox, setTaxBoxType, shouldOfferNewDocNo, taxBoxCalcInput, taxBoxView, taxLinesFromCalc, vatErrorTarget, withoutDisabledDimensions,
} from "./gl-journal-entry";

const vat = (patch: Partial<GLDetailVat> = {}): GLDetailVat => ({
  id: "V1", tax_type: 1, document_type: 1, tax_invoice_no: "IV6909-001", tax_invoice_date: "2026-09-10", tax_period_year: 2026, tax_period_month: 9,
  partner_name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", base_amount: "10000.00", zero_rate_amount: "0", exempt_amount: "0", vat_rate: "7", vat_amount: "700.00",
  claim_status: 1, remark: "ซื้อปูนซีเมนต์ปอร์ตแลนด์", ...patch,
});
const line = (accountcode: string, debit: string, credit: string, description = ""): GLLine => ({ ...emptyLine(), accountcode, debit, credit, description, departmentcode: "D01", projectcode: "P01" });

describe("journal save option (Champ combo)", () => {
  it("lists the two Champ options in Champ order with the exact Thai texts", () => {
    expect(JOURNAL_SAVE_OPTIONS.map((option) => [option.value, option.th])).toEqual([["keep", "บันทึกแล้วไม่ล้างหน้าจอ"], ["new", "บันทึกแล้วเพิ่มรายการใหม่"]]);
  });
  it("remembers the option per user and falls back to Champ's first option", () => {
    expect(journalSaveOptionStorageKey(" Somchai ")).toBe("bc_gl_journal_save_option:somchai");
    expect(journalSaveOptionStorageKey(undefined)).toBe("bc_gl_journal_save_option:anonymous");
    expect(parseJournalSaveOption("new")).toBe("new");
    expect(parseJournalSaveOption("keep")).toBe("keep");
    expect(parseJournalSaveOption(null)).toBe("keep");
    expect(parseJournalSaveOption("stay")).toBe("keep");
  });
});

describe("nextJournalAfterSave", () => {
  const saved: GLJournal = {
    ...emptyJournal("PV", "manual", [], "00000"), id: "J1", version: 3, docno: "PV6909-0007", date: "2026-09-10", fiscalyear: "2569", status: "draft",
    reference: "IV6909-001", refdate: "2026-09-09", description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 500 ถุง เป็นเงินเชื่อ", branchcode: "00001",
    lines: [line("11310", "10000.00", "0", "ปูนซีเมนต์"), line("11410", "700.00", "0"), line("21110", "0", "10700.00")],
    details: { vats: [vat()] },
  };
  const blank = emptyJournal("JV", "manual", [], "00000");

  it("'new' returns the blank voucher untouched", () => {
    expect(nextJournalAfterSave(saved, "new", blank)).toBe(blank);
  });

  it("'keep' keeps header, accounts and line details but clears doc number, amounts and invoice", () => {
    const next = nextJournalAfterSave(saved, "keep", blank);
    expect(next.id).toBeUndefined();
    expect(next.version).toBeUndefined();
    expect(next.docno).toBe("");
    expect(next.status).toBe(blank.status);
    expect([next.bookcode, next.date, next.fiscalyear, next.reference, next.refdate, next.branchcode]).toEqual(["PV", "2026-09-10", "2569", "IV6909-001", "2026-09-09", "00001"]);
    expect(next.description).toBe(saved.description);
    expect(next.lines.map((item) => [item.accountcode, item.description, item.departmentcode, item.projectcode, item.debit, item.credit])).toEqual([
      ["11310", "ปูนซีเมนต์", "D01", "P01", "0", "0"], ["11410", "", "D01", "P01", "0", "0"], ["21110", "", "D01", "P01", "0", "0"],
    ]);
    const [kept] = next.details?.vats ?? [];
    expect(kept.id).not.toBe("V1");
    expect([kept.tax_type, kept.tax_invoice_date, kept.vat_rate, kept.remark, kept.claim_status]).toEqual([1, "2026-09-10", "7", "ซื้อปูนซีเมนต์ปอร์ตแลนด์", 1]);
    expect([kept.tax_invoice_no, kept.base_amount, kept.vat_amount, kept.partner_name]).toEqual(["", "0", undefined, ""]);
    expect(isBlankVatRow(kept)).toBe(true);
  });

  it("'keep' also keeps the allocation, job, unit and division codes of a line, and adds no empty key when a line has none", () => {
    const withExtra: GLJournal = { ...saved, lines: [{ ...line("11310", "10000.00", "0"), allocatecode: "A01", jobcode: " J07 ", partcode: "U1", sidecode: "S2" }, line("21110", "0", "10000.00")] };
    const [first, second] = nextJournalAfterSave(withExtra, "keep", blank).lines;
    expect([first.allocatecode, first.jobcode, first.partcode, first.sidecode]).toEqual(["A01", "J07", "U1", "S2"]);
    expect(Object.keys(second).sort()).toEqual(Object.keys({ ...emptyLine() }).sort());
  });

  it("'keep' does not carry a multi-invoice tax table or the other detail sections", () => {
    const next = nextJournalAfterSave({ ...saved, details: { vats: [vat(), vat({ id: "V2" })], withholdings: [] } }, "keep", blank);
    expect(next.details).toBeUndefined();
  });
});

describe("tax box", () => {
  it("new VAT rows never get a default rate", () => {
    const row = newVatRow("2026-09-10", 2);
    expect(row.vat_rate).toBe("");
    expect(row.claim_status).toBeUndefined();
    expect([row.tax_period_year, row.tax_period_month]).toEqual([2026, 9]);
    expect(newVatRow("2026-09-10", 1).claim_status).toBe(1);
  });

  it("views none / single / multi", () => {
    expect(taxBoxView(undefined)).toEqual({ mode: "none" });
    expect(taxBoxView({ vats: [] })).toEqual({ mode: "none" });
    expect(taxBoxView({ vats: [vat()] })).toMatchObject({ mode: "single", row: { id: "V1" } });
    expect(taxBoxView({ vats: [vat(), vat({ id: "V2" })] })).toEqual({ mode: "multi", count: 2 });
  });

  it("choosing a type creates one row; 'no tax' removes it; multi is left to the tax table", () => {
    const created = setTaxBoxType(undefined, "2", "2026-09-10");
    expect(created?.vats).toHaveLength(1);
    expect(created?.vats?.[0]).toMatchObject({ tax_type: 2, vat_rate: "", tax_invoice_date: "2026-09-10" });
    expect(setTaxBoxType(created, "", "2026-09-10")?.vats).toEqual([]);
    expect(setTaxBoxType(undefined, "", "2026-09-10")).toBeUndefined();
    const multi = { vats: [vat(), vat({ id: "V2" })] };
    expect(setTaxBoxType(multi, "", "2026-09-10")).toBe(multi);
  });

  it("switching type keeps the row and applies the tax-table side effects", () => {
    const sale = setTaxBoxType({ vats: [vat({ claim_reason: "x" })] }, "2", "2026-10-01")?.vats?.[0];
    expect(sale).toMatchObject({ id: "V1", tax_type: 2, tax_invoice_no: "IV6909-001", tax_period_year: 2026, tax_period_month: 9 });
    expect(sale?.claim_status).toBeUndefined();
    expect(sale?.claim_reason).toBeUndefined();
    const purchase = setTaxBoxType({ vats: [vat({ tax_type: 2, claim_status: undefined })] }, "1", "2026-10-01")?.vats?.[0];
    expect(purchase).toMatchObject({ tax_type: 1, claim_status: 1 });
  });

  it("editing base or rate clears the stored tax so the backend recomputes; an explicit tax stays", () => {
    expect(patchTaxBox({ vats: [vat()] }, { base_amount: "20000.00" })?.vats?.[0].vat_amount).toBeUndefined();
    expect(patchTaxBox({ vats: [vat()] }, { vat_rate: "10" })?.vats?.[0].vat_amount).toBeUndefined();
    expect(patchTaxBox({ vats: [vat()] }, { base_amount: "100.00", vat_amount: "7.00" })?.vats?.[0]).toMatchObject({ base_amount: "100.00", vat_amount: "7.00" });
    expect(patchTaxBox({ vats: [vat()] }, { tax_invoice_no: "IV2" })?.vats?.[0].vat_amount).toBe("700.00");
    const multi = { vats: [vat(), vat({ id: "V2" })] };
    expect(patchTaxBox(multi, { remark: "x" })).toBe(multi);
  });

  it("a type-only box is blank and is dropped before saving; a filled one is kept", () => {
    const typeOnly = setTaxBoxType(undefined, "1", "2026-09-10");
    expect(isBlankVatRow(typeOnly!.vats![0])).toBe(true);
    expect(dropBlankTaxBoxRow(typeOnly)?.vats).toEqual([]);
    expect(dropBlankTaxBoxRow({ vats: [vat()] })?.vats).toHaveLength(1);
    expect(isBlankVatRow(vat({ tax_invoice_no: "", base_amount: "0.00", vat_amount: undefined, partner_name: "" }))).toBe(true);
    expect(isBlankVatRow(vat({ tax_invoice_no: "", base_amount: "0.01", vat_amount: undefined, partner_name: "" }))).toBe(false);
    expect(dropBlankTaxBoxRow(undefined)).toBeUndefined();
  });

  it("calc input needs a rate and a non-zero amount (no 7% default)", () => {
    expect(taxBoxCalcInput(undefined)).toEqual({ ok: false, problem: "no_row" });
    expect(taxBoxCalcInput({ vats: [vat({ vat_rate: " " })] })).toEqual({ ok: false, problem: "rate" });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0.00" })] })).toEqual({ ok: false, problem: "amount" });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "10700.00", tax_type: 2 })] })).toEqual({ ok: true, amount: "10700.00", rate: "7", taxType: 2 });
  });

  it("blank amount + typed tax asks the backend for the base like Champ OnTaxAuto (both filled = amount wins)", () => {
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0", vat_amount: "10.70" })] }, "exclusive")).toEqual({ ok: true, mode: "from_tax", amount: "10.70", rate: "7", taxType: 1 });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "", vat_amount: " 7.00 ", tax_type: 2 })] }, "inclusive")).toEqual({ ok: true, mode: "from_tax", amount: "7.00", rate: "7", taxType: 2 });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "10000.00", vat_amount: "10.70" })] }, "inclusive")).toEqual({ ok: true, mode: "inclusive", amount: "10000.00", rate: "7", taxType: 1 });
    // rate 0 goes to the backend (vat_calc_rate_zero) — the browser only checks that something was typed
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0", vat_amount: "7.00", vat_rate: "0" })] }, "exclusive")).toEqual({ ok: true, mode: "from_tax", amount: "7.00", rate: "0", taxType: 1 });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0", vat_amount: "7.00", vat_rate: "" })] }, "exclusive")).toEqual({ ok: false, problem: "rate" });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0", vat_amount: undefined })] }, "exclusive")).toEqual({ ok: false, problem: "amount" });
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0", vat_amount: "0.00" })] }, "exclusive")).toEqual({ ok: false, problem: "amount" });
    // without the pressed button the old contract stays: blank amount = ask for the amount
    expect(taxBoxCalcInput({ vats: [vat({ base_amount: "0", vat_amount: "10.70" })] })).toEqual({ ok: false, problem: "amount" });
  });
});

describe("taxLinesFromCalc (Champ OnTaxAuto amounts, no account codes)", () => {
  const calc = { base_amount: "10000.00", vat_amount: "700.00", total_amount: "10700.00" };
  it("purchase: Dr base · Dr tax · Cr total and clears every other amount", () => {
    const lines = [line("51110", "5.00", "0"), line("11410", "0", "0"), line("21110", "0", "3.00"), line("11120", "9.00", "0")];
    const next = taxLinesFromCalc(lines, 1, calc);
    expect(next.map((item) => [item.accountcode, item.debit, item.credit])).toEqual([["51110", "10000.00", "0"], ["11410", "700.00", "0"], ["21110", "0", "10700.00"], ["11120", "0", "0"]]);
  });
  it("sale: Dr total · Cr base · Cr tax, padding to three lines without guessing accounts", () => {
    const next = taxLinesFromCalc([line("11210", "0", "0")], 2, calc);
    expect(next.map((item) => [item.accountcode, item.debit, item.credit])).toEqual([["11210", "10700.00", "0"], ["", "0", "10000.00"], ["", "0", "700.00"]]);
  });
  it("detects existing amounts without floats", () => {
    expect(linesHaveAmounts([line("1", "0", "0"), line("2", "0.00", "")])).toBe(false);
    expect(linesHaveAmounts([line("1", "0", "0.01")])).toBe(true);
  });
});

describe("vatErrorTarget", () => {
  it("points a backend vat_ error at the single invoice row (fields the box lacks live in the F3 table)", () => {
    expect(vatErrorTarget({ vats: [vat()] }, "vat_partner_name_invalid", "partner_name")).toEqual({ section: "vats", row: 1, field: "partner_name" });
  });
  it("does not guess with several invoices, other error codes or no field", () => {
    expect(vatErrorTarget({ vats: [vat(), vat({ id: "V2" })] }, "vat_partner_name_invalid", "partner_name")).toBeNull();
    expect(vatErrorTarget({ vats: [vat()] }, "wht_payee_name_invalid", "partner_name")).toBeNull();
    expect(vatErrorTarget({ vats: [vat()] }, "vat_partner_name_invalid", "")).toBeNull();
    expect(vatErrorTarget(undefined, "vat_partner_name_invalid", "partner_name")).toBeNull();
  });
});

describe("shouldOfferNewDocNo (Champ: Document No is duplicated, run again?)", () => {
  const ctx = { isNew: true, typedDocNo: "JV6910-0001", bookFound: true };
  it("offers a new running number only for a typed number that the backend rejected as duplicate on a new voucher", () => {
    expect(shouldOfferNewDocNo("duplicate_code", ctx)).toBe(true);
  });
  it("keeps the plain error for other codes, edits, blank or unknown-book cases", () => {
    expect(shouldOfferNewDocNo("validation_failed", ctx)).toBe(false);
    expect(shouldOfferNewDocNo(undefined, ctx)).toBe(false);
    expect(shouldOfferNewDocNo("duplicate_code", { ...ctx, isNew: false })).toBe(false);
    expect(shouldOfferNewDocNo("duplicate_code", { ...ctx, typedDocNo: "  " })).toBe(false);
    expect(shouldOfferNewDocNo("duplicate_code", { ...ctx, bookFound: false })).toBe(false);
  });
});

describe("journalFunctionKey", () => {
  it("maps F2 / F3 / F5 without modifiers", () => {
    expect(journalFunctionKey({ key: "F2" })).toBe("list");
    expect(journalFunctionKey({ key: "F3" })).toBe("tax");
    expect(journalFunctionKey({ key: "F5" })).toBe("details");
  });
  it("leaves other keys, modified keys and handled events to the browser", () => {
    expect(journalFunctionKey({ key: "F9" })).toBeNull();
    expect(journalFunctionKey({ key: "F11" })).toBeNull();
    expect(journalFunctionKey({ key: "F5", ctrlKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F5", shiftKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F3", altKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F3", metaKey: true })).toBeNull();
    expect(journalFunctionKey({ key: "F2", defaultPrevented: true })).toBeNull();
  });
});

describe("analysis dimensions of a journal line (แผนก โครงการ การจัดสรร งาน หน่วยงาน ฝ่าย)", () => {
  const off = {} as Partial<GLDimensionSettingsInput>;
  const on = (...flags: (keyof GLDimensionSettingsInput)[]) => Object.fromEntries(flags.map((flag) => [flag, true])) as Partial<GLDimensionSettingsInput>;
  const emptyRegisters = { departments: [], projects: [], "allocate-codes": [], "job-codes": [], "part-codes": [], "side-codes": [] };
  const jobs: GLDimension[] = [{ code: "J01", name: "งานก่อสร้างอาคาร A", isactive: true }, { code: "J02", name: "งานปรับปรุงโกดัง", isactive: true }];
  const registers = { ...emptyRegisters, "job-codes": jobs };
  const account = (patch: Partial<GLAccount> = {}) => ({ accountcode: "51200", ...patch }) as GLAccount;
  const job = dimensionDef("job-codes");

  it("the extra dimensions are the four after department and project, in Champ order", () => {
    expect(EXTRA_DIMENSIONS.map((def) => def.line)).toEqual(["allocatecode", "jobcode", "partcode", "sidecode"]);
    expect(DIMENSIONS.map((def) => def.line)).toEqual(["departmentcode", "projectcode", "allocatecode", "jobcode", "partcode", "sidecode"]);
  });

  it("extraDimensionCodes keeps only filled, trimmed codes of the four extra dimensions", () => {
    expect(extraDimensionCodes({ departmentcode: "D01", projectcode: "P01" })).toEqual({});
    expect(extraDimensionCodes({ allocatecode: "", jobcode: " J01 ", partcode: undefined, sidecode: "  " })).toEqual({ jobcode: "J01" });
  });

  describe("journalDimensionColumns", () => {
    const lines = [{ departmentcode: "", allocatecode: "" }, { departmentcode: "", allocatecode: undefined }];
    it("shows no column when every switch is off and the voucher holds no code", () => {
      expect(journalDimensionColumns(off, lines)).toEqual([]);
      expect(journalDimensionColumns(undefined, undefined)).toEqual([]);
    });
    it("shows the columns of the switched-on dimensions in Champ order, whatever the order of the switches", () => {
      expect(journalDimensionColumns(on("sideseparate", "projectseparate", "jobseparate"), lines).map((def) => def.line)).toEqual(["projectcode", "jobcode", "sidecode"]);
      expect(journalDimensionColumns(on("departseparate", "projectseparate", "allocateseparate", "jobseparate", "partseparate", "sideseparate"), []).map((def) => def.label[1])).toEqual(["แผนก", "โครงการ", "การจัดสรร", "งาน", "หน่วยงาน", "ฝ่าย"]);
    });
    it("keeps the column of a dimension that was switched off later while the voucher still holds a code of it", () => {
      expect(journalDimensionColumns(off, [{ allocatecode: "A01" }, { departmentcode: "D01" }]).map((def) => def.line)).toEqual(["departmentcode", "allocatecode"]);
      expect(journalDimensionColumns(on("jobseparate"), [{ allocatecode: "A01" }]).map((def) => def.line)).toEqual(["allocatecode", "jobcode"]);
    });
    it("a blank code does not hold the column", () => {
      expect(journalDimensionColumns(off, [{ jobcode: "   " }])).toEqual([]);
    });
  });

  describe("lineDimensionCell", () => {
    it("company switch off: the cell is locked; without a code it points to the company defaults, with a code the saved code is shown and kept", () => {
      const empty = lineDimensionCell(job, off, registers, account({ jobseparate: true, joballcodes: true }), {});
      expect(empty).toMatchObject({ locked: true, hint: ["gl_dim_switch_off_hint", expect.any(String)], register: [], allowed: null });
      const held = lineDimensionCell(job, off, registers, account({ jobseparate: true }), { jobcode: " J01 " });
      expect(held).toMatchObject({ locked: true, hint: ["gl_dim_value_kept", expect.any(String)], allowed: null });
      expect(held.register).toEqual([{ code: "J01", name: "", isactive: true }]);
      // the held code is offered as a plain code, not flagged "not in the allowed list"
      expect(dimensionOptions(held.register, held.allowed, "J01")).toEqual([{ code: "J01", name: "", isactive: true }]);
    });
    it("company switch on, account not separated by it: locked with the account hint of that dimension while the line has no code", () => {
      const keys = DIMENSIONS.map((def) => lineDimensionCell(def, on(def.flag), emptyRegisters, account(), {}).hint?.[0]);
      expect(keys).toEqual(["gl_line_not_depart_separate", "gl_line_not_project_separate", "gl_line_not_allocate_separate", "gl_line_not_job_separate", "gl_line_not_part_separate", "gl_line_not_side_separate"]);
      expect(lineDimensionCell(job, on("jobseparate"), registers, account({ jobseparate: false }), {})).toMatchObject({ locked: true, allowed: [] });
      expect(lineDimensionCell(job, on("jobseparate"), registers, undefined, {}).locked).toBe(true);
    });
    it("company switch on, account not separated but the line already holds a code: still editable so it can be changed or cleared", () => {
      const cell = lineDimensionCell(job, on("jobseparate"), registers, account({ jobseparate: false }), { jobcode: "J01" });
      expect(cell).toMatchObject({ locked: false, hint: null, allowed: [] });
    });
    it("company switch on, account separated: editable; select-all = every code, a list = only those, neither = none (Champ)", () => {
      const switches = on("jobseparate");
      expect(lineDimensionCell(job, switches, registers, account({ jobseparate: true, joballcodes: true }), {})).toMatchObject({ locked: false, allowed: null, register: jobs });
      expect(lineDimensionCell(job, switches, registers, account({ jobseparate: true, jobcodes: ["J02"] }), {})).toMatchObject({ locked: false, allowed: ["J02"] });
      expect(lineDimensionCell(job, switches, registers, account({ jobseparate: true }), {})).toMatchObject({ locked: false, allowed: [] });
    });
    it("each dimension reads its own register, flag and list", () => {
      const all = on("departseparate", "projectseparate", "allocateseparate", "jobseparate", "partseparate", "sideseparate");
      const side = dimensionDef("side-codes");
      const sideRegister: GLDimension[] = [{ code: "S1", name: "ฝ่ายขาย", isactive: true }];
      const cell = lineDimensionCell(side, all, { ...emptyRegisters, "side-codes": sideRegister }, account({ sideseparate: true, sidecodes: ["S1"], jobseparate: true }), {});
      expect(cell).toMatchObject({ locked: false, allowed: ["S1"], register: sideRegister });
    });
  });

  describe("withoutDisabledDimensions", () => {
    const pasted = [
      { accountcode: "51200", departmentcode: "D01", projectcode: "", allocatecode: "A01", jobcode: "J01" },
      { accountcode: "11110", departmentcode: "", projectcode: "P01" },
    ];
    it("clears the codes of dimensions the company has not switched on and names them in Champ order", () => {
      const result = withoutDisabledDimensions(pasted, on("projectseparate", "jobseparate"));
      expect(result.dropped.map((def) => def.line)).toEqual(["departmentcode", "allocatecode"]);
      expect(result.lines).toEqual([
        { accountcode: "51200", departmentcode: "", projectcode: "", allocatecode: "", jobcode: "J01" },
        { accountcode: "11110", departmentcode: "", projectcode: "P01" },
      ]);
    });
    it("leaves the lines as they are when every code belongs to a switched-on dimension", () => {
      const result = withoutDisabledDimensions(pasted, on("departseparate", "projectseparate", "allocateseparate", "jobseparate"));
      expect(result.dropped).toEqual([]);
      expect(result.lines).toEqual(pasted);
    });
    it("does not mutate the lines it was given and adds no key to a line that never had one", () => {
      const original = JSON.stringify(pasted);
      const result = withoutDisabledDimensions(pasted, off);
      expect(JSON.stringify(pasted)).toBe(original);
      expect(Object.keys(result.lines[1]).sort()).toEqual(["accountcode", "departmentcode", "projectcode"]);
    });
  });
});

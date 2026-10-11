import { describe, expect, it } from "vitest";
import { dimensionDef, type GLDimension } from "./general-ledger";
import { blankRegisterRow, decodeRegisterFile, exportRegisterText, mergeImportedRows, parseRegisterImport, planRegisterSave, registerDirty, registerRowAction, registerRowsFrom, registerStepPayload, type RegisterRow } from "./gl-dimension-register";

const job = dimensionDef("job-codes");
const item = (code: string, name: string, extra: Partial<GLDimension> = {}): GLDimension => ({ id: `id-${code}`, version: 1, code, name, isactive: true, ...extra } as GLDimension);
let counter = 0;
const nextKey = () => `new-${++counter}`;

describe("register grid rows", () => {
  it("loads a register in code order with the stored values kept for change detection", () => {
    const rows = registerRowsFrom([item("B2", "สอง"), item("A1", "หนึ่ง", { usergroup: "ทีมขาย" })]);
    expect(rows.map((row) => row.code)).toEqual(["A1", "B2"]);
    expect(rows[0]).toMatchObject({ id: "id-A1", version: 1, usergroup: "ทีมขาย", saved: { code: "A1", name: "หนึ่ง", usergroup: "ทีมขาย", isactive: true } });
    expect(registerDirty(rows)).toBe(false);
  });
  it("decides the F11 action of each row like Champ (create / delete when the code is cleared / update)", () => {
    const [stored] = registerRowsFrom([item("A1", "หนึ่ง")]);
    expect(registerRowAction(blankRegisterRow("n"))).toBe("skip");
    expect(registerRowAction({ ...blankRegisterRow("n"), code: "X1", name: "ใหม่" })).toBe("create");
    expect(registerRowAction({ ...blankRegisterRow("n"), code: "X1" })).toBe("incomplete");
    expect(registerRowAction(stored)).toBe("skip");
    expect(registerRowAction({ ...stored, name: "หนึ่งใหม่" })).toBe("update");
    expect(registerRowAction({ ...stored, isactive: false })).toBe("update");
    expect(registerRowAction({ ...stored, code: " " })).toBe("delete");
    expect(registerRowAction({ ...stored, removed: true })).toBe("delete");
    expect(registerRowAction({ ...stored, name: "" })).toBe("incomplete");
  });
});

describe("planRegisterSave", () => {
  it("sends deletes first, then updates, then creates, and nothing for untouched rows", () => {
    const [a, b, c] = registerRowsFrom([item("A1", "หนึ่ง"), item("B2", "สอง"), item("C3", "สาม")]);
    const plan = planRegisterSave([{ ...a, name: "แก้" }, { ...b, removed: true }, c, { ...blankRegisterRow("n1"), code: "D4", name: "สี่" }], job);
    expect(plan.steps.map((step) => `${step.action}:${step.row.code}`)).toEqual(["delete:B2", "update:A1", "create:D4"]);
    expect(plan.problems).toEqual({});
  });
  it("reports a row problem per row: too long, duplicate code, half filled", () => {
    const plan = planRegisterSave([
      { ...blankRegisterRow("a"), code: "ABCDEFGHIJKLMNOP", name: "ยาว" },
      { ...blankRegisterRow("b"), code: "J1", name: "งาน" },
      { ...blankRegisterRow("c"), code: "J1", name: "ซ้ำ" },
      { ...blankRegisterRow("d"), code: "J2" },
      { ...blankRegisterRow("e"), code: "J3", name: "x".repeat(256) },
    ], job);
    expect(plan.problems).toEqual({
      a: { reason: "code_too_long", max: 15 },
      b: { reason: "duplicate_code" },
      c: { reason: "duplicate_code" },
      d: { reason: "name_required" },
      e: { reason: "name_too_long", max: 255 },
    });
    expect(plan.steps).toEqual([]);
  });
  it("lets a code move from a deleted row to a new one in the same save", () => {
    const [a] = registerRowsFrom([item("J1", "เดิม")]);
    const plan = planRegisterSave([{ ...a, removed: true }, { ...blankRegisterRow("n"), code: "J1", name: "ใหม่" }], job);
    expect(plan.steps.map((step) => step.action)).toEqual(["delete", "create"]);
  });
  it("builds the register payload with the stored id and version", () => {
    const [a] = registerRowsFrom([item("A1", "หนึ่ง")]);
    expect(registerStepPayload({ ...a, name: " ใหม่ ", usergroup: " กลุ่ม " })).toMatchObject({ id: "id-A1", version: 1, code: "A1", name: "ใหม่", usergroup: "กลุ่ม", isactive: true });
  });
});

describe("Export / Import of CODE<TAB>NAME", () => {
  const rows = registerRowsFrom([item("A1", "หนึ่ง"), item("B2", "ชื่อ\tมีแท็บ"), item("C3", "ถูกลบ")]).map((row) => row.code === "C3" ? { ...row, removed: true } : row);
  it("writes a BOM, the CODE/NAME header and one CRLF line per kept row", () => {
    expect(exportRegisterText(rows)).toBe("\uFEFFCODE\tNAME\r\nA1\tหนึ่ง\r\nB2\tชื่อ มีแท็บ\r\n");
  });
  it("reads its own export back", () => {
    expect(parseRegisterImport(exportRegisterText(rows), job)).toEqual({ rows: [{ code: "A1", name: "หนึ่ง" }, { code: "B2", name: "ชื่อ มีแท็บ" }] });
  });
  it("reads a file without a header, with blank lines and LF endings", () => {
    expect(parseRegisterImport("A1\tหนึ่ง\n\nB2\tสอง\n", job)).toEqual({ rows: [{ code: "A1", name: "หนึ่ง" }, { code: "B2", name: "สอง" }] });
  });
  it("refuses the whole file at the first bad line and names it", () => {
    expect(parseRegisterImport("CODE\tNAME\nA1\tหนึ่ง\nB2 ไม่มีแท็บ\n", job)).toEqual({ error: { line: 3, reason: "format" } });
    expect(parseRegisterImport("A1\t\n", job)).toEqual({ error: { line: 1, reason: "format" } });
    expect(parseRegisterImport("ABCDEFGHIJKLMNOP\tยาว\n", job)).toEqual({ error: { line: 1, reason: "length" } });
    expect(parseRegisterImport("A1\tหนึ่ง\nA1\tซ้ำ\n", job)).toEqual({ error: { line: 2, reason: "duplicate" } });
    expect(parseRegisterImport("CODE\tNAME\n", job)).toEqual({ error: { line: 1, reason: "format" } });
    expect(parseRegisterImport("", job)).toEqual({ error: { line: 1, reason: "format" } });
  });
  it("merges into the grid: a known code takes the new name, a new code becomes a new row, blank rows go", () => {
    const grid = [...registerRowsFrom([item("A1", "หนึ่ง"), item("B2", "สอง")]), blankRegisterRow("blank")];
    grid[1] = { ...grid[1], removed: true };
    const merged = mergeImportedRows(grid, [{ code: "B2", name: "สองใหม่" }, { code: "Z9", name: "ใหม่" }], nextKey);
    expect(merged).toMatchObject({ added: 1, updated: 1 });
    expect(merged.rows.map((row) => [row.code, row.name, Boolean(row.removed)])).toEqual([["A1", "หนึ่ง", false], ["B2", "สองใหม่", false], ["Z9", "ใหม่", false]]);
    expect(merged.rows[2].id).toBeUndefined();
    expect(registerRowAction(merged.rows[1] as RegisterRow)).toBe("update");
  });
});

describe("decodeRegisterFile", () => {
  it("reads UTF-8, and falls back to Windows-874 for a file saved by Champ", () => {
    expect(decodeRegisterFile(new TextEncoder().encode("A1\tแผนก").buffer as ArrayBuffer)).toBe("A1\tแผนก");
    const cp874 = Uint8Array.from([0x41, 0x31, 0x09, 0xe1, 0xbc, 0xb9, 0xa1]);
    expect(decodeRegisterFile(cp874.buffer)).toBe("A1\tแผนก");
  });
});

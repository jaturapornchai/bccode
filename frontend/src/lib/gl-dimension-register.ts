import { dimensionPayload, type GLDimension, type GLDimensionDef } from "./general-ledger";

// Grid logic of the code-setup screen "กำหนดรหัสต่างๆ" (Champ myfrmbasemaster.cpp): the register is edited in a grid, nothing is stored
// until F11 (OnMasterSave :393-560), Import only fills the grid (:743-864), Export writes the grid (:685-742).

export const REGISTER_USERGROUP_MAX = 255;
const IMPORT_HEADER = ["CODE", "NAME"] as const;

/** One grid row. `saved` is the stored record the row was loaded from (absent for a new row). */
export type RegisterRow = {
  key: string;
  id?: string;
  version?: number;
  code: string;
  name: string;
  usergroup: string;
  isactive: boolean;
  /** marked for deletion — the stored record goes when the grid is saved */
  removed?: boolean;
  saved?: { code: string; name: string; usergroup: string; isactive: boolean };
};

export type RegisterProblem = { reason: "name_required" | "code_too_long" | "name_too_long" | "usergroup_too_long" | "duplicate_code"; max?: number };
export type RegisterAction = "create" | "update" | "delete";
export type RegisterStep = { row: RegisterRow; action: RegisterAction };
export type RegisterPlan = { steps: RegisterStep[]; problems: Record<string, RegisterProblem> };

const text = (value: unknown) => typeof value === "string" ? value : "";
const length = (value: string) => Array.from(value).length;

/** Grid rows of a loaded register, in code order (Champ lists by code). */
export function registerRowsFrom(list: readonly GLDimension[]): RegisterRow[] {
  return [...list]
    .sort((a, b) => a.code < b.code ? -1 : a.code > b.code ? 1 : 0)
    .map((item) => {
      const saved = { code: item.code, name: text(item.name), usergroup: text(item.usergroup), isactive: item.isactive !== false };
      return { key: item.id || item.code, id: item.id, version: item.version, ...saved, saved };
    });
}

export function blankRegisterRow(key: string): RegisterRow { return { key, code: "", name: "", usergroup: "", isactive: true }; }

const changed = (row: RegisterRow) => !row.saved || row.saved.code !== row.code.trim() || row.saved.name !== row.name.trim() || row.saved.usergroup !== row.usergroup.trim() || row.saved.isactive !== row.isactive;

/** Champ F11 per row (:509-536): a new row with code + name is created, a stored row whose code was cleared is deleted, anything else changed is updated. */
export function registerRowAction(row: RegisterRow): RegisterAction | "skip" | "incomplete" {
  const code = row.code.trim(), name = row.name.trim();
  if (!row.id) return !code && !name && !row.usergroup.trim() ? "skip" : code && name ? "create" : "incomplete";
  if (row.removed || !code) return "delete";
  if (!name) return "incomplete";
  return changed(row) ? "update" : "skip";
}

export function registerDirty(rows: readonly RegisterRow[]): boolean {
  return rows.some((row) => {
    const action = registerRowAction(row);
    return action !== "skip";
  });
}

/** The commands F11 sends, deletes first so a code can move from one row to another, plus the problems found per row key. */
export function planRegisterSave(rows: readonly RegisterRow[], def: Pick<GLDimensionDef, "codeMax" | "nameMax">): RegisterPlan {
  const problems: Record<string, RegisterProblem> = {};
  const steps: RegisterStep[] = [];
  const kept = new Map<string, string>();
  for (const row of rows) {
    const action = registerRowAction(row);
    if (action === "skip") { if (row.id && !row.removed) kept.set(row.code.trim(), row.key); continue; }
    if (action === "incomplete") { problems[row.key] = { reason: "name_required" }; continue; }
    if (action === "delete") { steps.push({ row, action }); continue; }
    const code = row.code.trim(), name = row.name.trim(), usergroup = row.usergroup.trim();
    if (length(code) > def.codeMax) problems[row.key] = { reason: "code_too_long", max: def.codeMax };
    else if (length(name) > def.nameMax) problems[row.key] = { reason: "name_too_long", max: def.nameMax };
    else if (length(usergroup) > REGISTER_USERGROUP_MAX) problems[row.key] = { reason: "usergroup_too_long", max: REGISTER_USERGROUP_MAX };
    else if (kept.has(code)) { problems[row.key] = { reason: "duplicate_code" }; problems[kept.get(code)!] ??= { reason: "duplicate_code" }; }
    else { kept.set(code, row.key); steps.push({ row, action }); }
  }
  const order: Record<RegisterAction, number> = { delete: 0, update: 1, create: 2 };
  return { steps: steps.filter((step) => !problems[step.row.key]).sort((a, b) => order[a.action] - order[b.action]), problems };
}

/** The register record a create/update step sends. */
export function registerStepPayload(row: RegisterRow): GLDimension {
  return dimensionPayload({ id: row.id ?? "", version: row.version, code: row.code, name: row.name, usergroup: row.usergroup, isactive: row.isactive });
}

/** File Export writes: header line, then one `CODE<TAB>NAME` per grid row (Champ :685-742). UTF-8 with BOM so Excel and Notepad show Thai. */
export function exportRegisterText(rows: readonly RegisterRow[]): string {
  const clean = (value: string) => value.replace(/[\t\r\n]+/g, " ").trim();
  const lines = rows.filter((row) => !row.removed && row.code.trim()).map((row) => `${clean(row.code)}\t${clean(row.name)}`);
  return `\uFEFF${[IMPORT_HEADER.join("\t"), ...lines].join("\r\n")}\r\n`;
}

export type RegisterImportRow = { code: string; name: string };
export type RegisterImportResult = { rows: RegisterImportRow[]; error?: undefined } | { rows?: undefined; error: { line: number; reason: "format" | "length" | "duplicate" } };

/** Reads a `CODE<TAB>NAME` file (Champ :743-864). The whole file is refused at the first bad line so the grid never holds half a file. */
export function parseRegisterImport(content: string, def: Pick<GLDimensionDef, "codeMax" | "nameMax">): RegisterImportResult {
  const lines = content.replace(/^\uFEFF/, "").split(/\r\n|\n|\r/);
  const rows: RegisterImportRow[] = [], seen = new Set<string>();
  let headerChecked = false;
  for (const [index, raw] of lines.entries()) {
    if (!raw.trim()) continue;
    const line = index + 1;
    const cells = raw.split("\t");
    const code = (cells[0] ?? "").trim(), name = (cells[1] ?? "").trim();
    if (!headerChecked) {
      headerChecked = true;
      if (code.toUpperCase() === IMPORT_HEADER[0] && name.toUpperCase() === IMPORT_HEADER[1]) continue;
    }
    if (cells.length < 2 || !code || !name) return { error: { line, reason: "format" } };
    if (length(code) > def.codeMax || length(name) > def.nameMax) return { error: { line, reason: "length" } };
    if (seen.has(code)) return { error: { line, reason: "duplicate" } };
    seen.add(code);
    rows.push({ code, name });
  }
  return rows.length ? { rows } : { error: { line: 1, reason: "format" } };
}

/** Puts imported rows into the grid: a code already there gets the new name (and is un-removed), a new code becomes a new row. */
export function mergeImportedRows(rows: readonly RegisterRow[], imported: readonly RegisterImportRow[], nextKey: () => string): { rows: RegisterRow[]; added: number; updated: number } {
  const result = rows.filter((row) => row.id || row.code.trim() || row.name.trim()).map((row) => ({ ...row }));
  let added = 0, updated = 0;
  for (const item of imported) {
    const existing = result.find((row) => row.code.trim() === item.code);
    if (existing) { existing.name = item.name; existing.removed = false; updated += 1; }
    else { result.push({ ...blankRegisterRow(nextKey()), code: item.code, name: item.name }); added += 1; }
  }
  return { rows: result, added, updated };
}

/** Reads an import file as UTF-8, or as Windows-874 (Champ's own export, Thai ANSI) when it is not valid UTF-8. */
export function decodeRegisterFile(buffer: ArrayBuffer): string {
  try { return new TextDecoder("utf-8", { fatal: true }).decode(buffer); }
  catch { return new TextDecoder("windows-874").decode(buffer); }
}

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { accountDimensionCodes, dimensionOptions, lineDimensionsAfterAccountChange, normalizeDimensionCodes, type GLAccount, type GLDimension } from "@/lib/general-ledger";
import { DimensionCodeMultiSelect, DimensionSelect, dimensionLabel } from "./gl-dimension-select";
import { accountSavePayload, normalizeRecord } from "./gl-masters";

const tr = (_key: string, fallback: string) => fallback;
const dim = (code: string, name: string, isactive = true): GLDimension => ({ id: `id-${code}`, code, name, isactive });
const register = [dim("D02", "ฝ่ายขาย"), dim("D01", "ฝ่ายบัญชีและการเงิน"), dim("D03", "ฝ่ายผลิต", false)];
const account = (extra: Partial<GLAccount>): GLAccount => ({ accountcode: "5100", names: [{ code: "th", name: "ค่าใช้จ่ายในการขาย" }], accounttype: "expense", parentaccountcode: null, normalbalance: "debit", allowposting: true, isactive: true, accountgroup: "", iscash: false, level: 4, ...extra });

describe("dimensionOptions (journal line combo)", () => {
  it("offers active register codes in code order; an empty account list means every code", () => {
    expect(dimensionOptions(register, [], "").map((item) => item.code)).toEqual(["D01", "D02"]);
  });
  it("narrows to the account's allowed list", () => {
    expect(dimensionOptions(register, ["D02"], "").map((item) => item.code)).toEqual(["D02"]);
  });
  it("keeps the code the line already holds even when it is no longer offered, flagged as not offered", () => {
    const options = dimensionOptions(register, ["D02"], "D01");
    expect(options.map((item) => item.code)).toEqual(["D02", "D01"]);
    expect(options[1].isactive).toBe(false);
    expect(dimensionOptions(register, [], "D03").at(-1)).toMatchObject({ code: "D03", isactive: false });
  });
  it("labels code and name, or the code with a not-offered note", () => {
    expect(dimensionLabel(register[1], tr)).toBe("D01 · ฝ่ายบัญชีและการเงิน");
    expect(dimensionLabel({ code: "D09", name: "", isactive: false }, tr)).toBe("D09 · ไม่อยู่ในรายการที่เลือกได้");
  });
});

describe("lineDimensionsAfterAccountChange (Champ clears the cell)", () => {
  const line = { departmentcode: "D01", projectcode: "P01" };
  it("keeps codes when the new account separates them and lists nothing", () => {
    expect(lineDimensionsAfterAccountChange(line, account({ departseparate: true, projectseparate: true }))).toEqual({});
  });
  it("clears a dimension whose flag is off or whose account has no record", () => {
    expect(lineDimensionsAfterAccountChange(line, account({ departseparate: true }))).toEqual({ projectcode: "" });
    expect(lineDimensionsAfterAccountChange(line, undefined)).toEqual({ departmentcode: "", projectcode: "" });
  });
  it("clears a code outside the new account's list and keeps one inside it", () => {
    const listed = account({ departseparate: true, projectseparate: true, departcodes: ["D02"], projectcodes: ["P01"] });
    expect(lineDimensionsAfterAccountChange(line, listed)).toEqual({ departmentcode: "" });
  });
  it("leaves an empty line alone", () => {
    expect(lineDimensionsAfterAccountChange({ departmentcode: "", projectcode: " " }, undefined)).toEqual({});
  });
});

describe("account allowed-code lists in the master screen", () => {
  it("normalises codes: trimmed, unique, sorted; junk becomes empty", () => {
    expect(normalizeDimensionCodes([" D02", "D01", "D02", ""])).toEqual(["D01", "D02"]);
    expect(normalizeDimensionCodes(undefined)).toEqual([]);
    expect(accountDimensionCodes(account({ projectcodes: ["P01"] }), "projects")).toEqual(["P01"]);
    expect(accountDimensionCodes(account({}), "departments")).toEqual([]);
  });
  it("loads a record with empty lists when the account has none", () => {
    const norm = normalizeRecord("accounts", { accountcode: "51000", accounttype: "expense" } as never) as GLAccount;
    expect(norm.departcodes).toEqual([]);
    expect(norm.projectcodes).toEqual([]);
  });
  it("sends a list only with its flag on, normalised, and leaves the key out when empty", () => {
    const on = accountSavePayload(account({ departseparate: true, projectseparate: true, departcodes: ["D02", " D01 "], projectcodes: [] }), false);
    expect(on.departcodes).toEqual(["D01", "D02"]);
    expect("projectcodes" in on).toBe(false);
    const off = accountSavePayload(account({ departseparate: false, projectseparate: true, departcodes: ["D01"], projectcodes: ["P01"] }), false);
    expect("departcodes" in off).toBe(false);
    expect(off.projectcodes).toEqual(["P01"]);
  });
});

describe("DimensionSelect / DimensionCodeMultiSelect markup", () => {
  it("renders the held code on the line combo with its label", () => {
    const html = renderToStaticMarkup(createElement(DimensionSelect, { label: "แผนกบรรทัด 1", value: "D01", register, allowed: [], onChange: () => undefined }));
    expect(html).toContain("D01 · ฝ่ายบัญชีและการเงิน");
    expect(html).toContain('aria-label="แผนกบรรทัด 1"');
  });
  it("shows a locked cell as a dash and a held stale code with the not-offered note", () => {
    const locked = renderToStaticMarkup(createElement(DimensionSelect, { label: "แผนกบรรทัด 1", value: "", register: [], allowed: [], disabled: true, onChange: () => undefined }));
    expect(locked).toContain("—");
    const stale = renderToStaticMarkup(createElement(DimensionSelect, { label: "แผนกบรรทัด 1", value: "D03", register, allowed: ["D02"], onChange: () => undefined }));
    expect(stale).toContain("D03 · ไม่อยู่ในรายการที่เลือกได้");
  });
  it("account picker: nothing chosen reads as every code; chosen codes show as removable chips", () => {
    const empty = renderToStaticMarkup(createElement(DimensionCodeMultiSelect, { resource: "departments", value: [], onChange: () => undefined, register }));
    expect(empty).toContain("ทุกแผนก");
    expect(empty).not.toContain("ล้างทั้งหมด");
    const chosen = renderToStaticMarkup(createElement(DimensionCodeMultiSelect, { resource: "projects", value: ["D01", "D09"], onChange: () => undefined, register }));
    expect(chosen).toContain("D01 · ฝ่ายบัญชีและการเงิน");
    expect(chosen).toContain("D09 · ไม่อยู่ในรายการที่เลือกได้");
    expect(chosen).toContain("เอารหัส D09 ออก");
    expect(chosen).toContain("ล้างทั้งหมด (ใช้ได้ทุกรหัส)");
  });
});

describe("journal screen wiring", () => {
  const journals = readFileSync(resolve(process.cwd(), "src", "app", "gl", "gl-journals.tsx"), "utf8");
  it("uses the register combos, not free-text inputs, for the line department and project", () => {
    expect(journals).toContain("<DimensionSelect");
    expect(journals).not.toMatch(/<input[^>]*departmentcode: e\.target\.value/);
    expect(journals).not.toMatch(/<input[^>]*projectcode: e\.target\.value/);
  });
  it("clears a disallowed department / project when the line account changes", () => {
    expect(journals).toContain("...lineDimensionsAfterAccountChange(line, accountsByCode.get(accountcode))");
  });
});

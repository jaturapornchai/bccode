import { describe, expect, it } from "vitest";
import { accountEntriesText, accountGroupPayload, emptyMaster, replaceAccountGroupSingles, statementSuggestSignature, type GLAccountEntry, type GLMaster, type GLStatementTemplate } from "@/lib/general-ledger";

const master = (accountentries: GLAccountEntry[]): GLMaster => ({ ...emptyMaster(), code: " G-1 ", name: " Cash and deposits ", accountentries });

describe("account group entries", () => {
  it("summarises entries the way Champ writes AccGroup", () => {
    expect(accountEntriesText([{ from: "1110", to: "1130" }, { from: "1210" }, { from: "1300", to: "" }])).toBe("1110:1130, 1210, 1300");
    expect(accountEntriesText(undefined)).toBe("");
  });

  it("keeps ranges and still-selected singles in place, appends new picks, drops unticked singles", () => {
    const entries: GLAccountEntry[] = [{ from: "1110" }, { from: "1200", to: "1299" }, { from: "1210" }];
    expect(replaceAccountGroupSingles(entries, ["1210", "1110", "1300"])).toEqual([{ from: "1110" }, { from: "1200", to: "1299" }, { from: "1210" }, { from: "1300" }]);
    expect(replaceAccountGroupSingles(entries, [])).toEqual([{ from: "1200", to: "1299" }]);
  });

  it("builds the save payload: trims, drops blank rows, collapses an end equal to the start", () => {
    const payload = accountGroupPayload(master([{ from: " 1110 ", to: " 1130 " }, { from: "", to: "" }, { from: "1120", to: "1120" }, { from: "1210" }, { from: "1300", to: "" }]));
    expect(payload.code).toBe("G-1");
    expect(payload.name).toBe("Cash and deposits");
    expect(payload.accountentries).toEqual([{ from: "1110", to: "1130" }, { from: "1120" }, { from: "1210" }, { from: "1300" }]);
    expect(accountGroupPayload(master([])).accountentries).toEqual([]);
  });

  it("keeps a half-filled range so the backend can name the missing code", () => {
    expect(accountGroupPayload(master([{ from: "", to: "1130" }])).accountentries).toEqual([{ from: "", to: "1130" }]);
  });

  it("refreshes statement suggestions when a line switches to a group", () => {
    const template = {
      statementtype: "balance_sheet",
      rows: [{ id: "r1", rowno: 10, rowtype: "account", title: "x", accountcodes: ["1000"], suggestkey: "cash" }],
    } as unknown as GLStatementTemplate;
    const grouped = { ...template, rows: [{ ...template.rows[0], accountgroup: "G-1" }] } as GLStatementTemplate;
    expect(statementSuggestSignature(grouped)).not.toBe("");
    expect(statementSuggestSignature(grouped)).not.toBe(statementSuggestSignature(template));
  });
});

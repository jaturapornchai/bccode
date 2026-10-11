import { describe, expect, it } from "vitest";
import {
  CHEQUE_DOC_CONFIGS,
  CHEQUE_GL_STATUS,
  CHEQUE_REGISTER_ROUTES,
  CHEQUE_SCREEN_ROUTES,
  CHEQUE_STATUS_IN,
  CHEQUE_STATUS_OUT,
  CHEQUE_TYPE_ISSUED,
  CHEQUE_TYPE_RECEIVED,
  chequeDocBookPolicy,
  chequeDocConfig,
  chequeDocTitleByKind,
  chequeGlStatus,
  chequeScreenFromRoute,
  chequeStatusLabel,
  chequeStatusTone,
  registerStatusChoices,
} from "./cheque-doc-config";

const config = (flag: number) => {
  const found = chequeDocConfig(flag);
  if (!found) throw new Error(`no config for ${flag}`);
  return found;
};

describe("status document configs 419-426", () => {
  it("covers exactly the eight Champ status documents, each on its own route", () => {
    expect(CHEQUE_DOC_CONFIGS.map((item) => item.flag)).toEqual([419, 420, 421, 422, 423, 424, 425, 426]);
    expect(new Set(CHEQUE_DOC_CONFIGS.map((item) => item.route)).size).toBe(8);
    expect(new Set(CHEQUE_DOC_CONFIGS.map((item) => item.doctypeKey)).size).toBe(8);
    expect(chequeDocConfig(418)).toBeUndefined();
  });

  it("assigns received cheques to 419-424 and issued cheques to 425-426", () => {
    for (const flag of [419, 420, 421, 422, 423, 424]) expect(config(flag).chequeType, String(flag)).toBe(CHEQUE_TYPE_RECEIVED);
    for (const flag of [425, 426]) expect(config(flag).chequeType, String(flag)).toBe(CHEQUE_TYPE_ISSUED);
  });

  it("gives only 419 the remark column and no fee, only 423 the cancel type combo, only 426 the cancelled-passed box", () => {
    expect(CHEQUE_DOC_CONFIGS.filter((item) => !item.hasFee).map((item) => item.flag)).toEqual([419]);
    expect(CHEQUE_DOC_CONFIGS.filter((item) => item.showRemark).map((item) => item.flag)).toEqual([419]);
    expect(CHEQUE_DOC_CONFIGS.filter((item) => item.cancelTypeCombo).map((item) => item.flag)).toEqual([423]);
    expect(CHEQUE_DOC_CONFIGS.filter((item) => item.cancelPassedBox).map((item) => item.flag)).toEqual([426]);
  });
});

describe("chequeDocBookPolicy (mirrors backend cheque/state.go bookPolicyFor)", () => {
  it("419 needs a book but does not match cheques to it", () => {
    expect(chequeDocBookPolicy(config(419), null)).toEqual({ required: true, match: false, none: false });
  });

  it("420, 421, 422, 425 and 426 need a book and match cheques to it", () => {
    for (const flag of [420, 421, 422, 425, 426]) expect(chequeDocBookPolicy(config(flag), null), String(flag)).toEqual({ required: true, match: true, none: false });
  });

  it("424 matches cheques to a book but the book itself is optional", () => {
    expect(chequeDocBookPolicy(config(424), null)).toEqual({ required: false, match: true, none: false });
  });

  it("423 uses no book for type 0 (on-hand) and a matching required book for the other types", () => {
    expect(chequeDocBookPolicy(config(423), 0)).toEqual({ required: false, match: false, none: true });
    expect(chequeDocBookPolicy(config(423), null)).toEqual({ required: false, match: false, none: true });
    for (const type of [1, 2, 3]) expect(chequeDocBookPolicy(config(423), type), String(type)).toEqual({ required: true, match: true, none: false });
  });
});

describe("chequeScreenFromRoute", () => {
  it("routes the two register menus and the eight status documents, ignoring the query string", () => {
    expect(chequeScreenFromRoute("/transaction/chequereceived")).toEqual({ kind: "register", chequeType: CHEQUE_TYPE_RECEIVED });
    expect(chequeScreenFromRoute("/transaction/chequeissued")).toEqual({ kind: "register", chequeType: CHEQUE_TYPE_ISSUED });
    expect(chequeScreenFromRoute("/banking/cheques/deposit")).toEqual({ kind: "doc", flag: 419 });
    expect(chequeScreenFromRoute("/banking/cheques/issued-cancel?tab=1")).toEqual({ kind: "doc", flag: 426 });
  });

  it("returns null for every other route", () => {
    for (const route of ["", "/transaction/paid", "/banking/cheques", "/banking/cheques/deposit/x", "/report/chequereceived"]) expect(chequeScreenFromRoute(route), route).toBeNull();
  });

  it("lists exactly the ten cheque screen routes with no duplicates", () => {
    expect(CHEQUE_SCREEN_ROUTES).toHaveLength(10);
    expect(new Set(CHEQUE_SCREEN_ROUTES).size).toBe(10);
    for (const route of Object.values(CHEQUE_REGISTER_ROUTES)) expect(CHEQUE_SCREEN_ROUTES).toContain(route);
    for (const route of CHEQUE_SCREEN_ROUTES) expect(chequeScreenFromRoute(route), route).not.toBeNull();
  });
});

describe("status labels and choices", () => {
  it("knows 7 received and 4 issued statuses and falls back for anything else", () => {
    expect(CHEQUE_STATUS_IN).toHaveLength(7);
    expect(CHEQUE_STATUS_OUT).toHaveLength(4);
    expect(chequeStatusLabel(CHEQUE_TYPE_RECEIVED, 2)[0]).toBe("cheque_status_in_2");
    expect(chequeStatusLabel(CHEQUE_TYPE_ISSUED, 3)[0]).toBe("cheque_status_out_3");
    expect(chequeStatusLabel(CHEQUE_TYPE_RECEIVED, 9)[0]).toBe("cheque_status_unknown");
    expect(chequeStatusLabel(CHEQUE_TYPE_ISSUED, -1)[0]).toBe("cheque_status_unknown");
  });

  it("marks cleared as success and cancelled as warning for each cheque type", () => {
    expect(chequeStatusTone(CHEQUE_TYPE_RECEIVED, 2)).toBe("success");
    expect(chequeStatusTone(CHEQUE_TYPE_RECEIVED, 4)).toBe("warning");
    expect(chequeStatusTone(CHEQUE_TYPE_RECEIVED, 0)).toBe("outline");
    expect(chequeStatusTone(CHEQUE_TYPE_ISSUED, 1)).toBe("success");
    expect(chequeStatusTone(CHEQUE_TYPE_ISSUED, 2)).toBe("warning");
    expect(chequeStatusTone(CHEQUE_TYPE_ISSUED, 3)).toBe("outline");
  });

  it("offers every received status, but a new issued cheque only on-hand or prepared", () => {
    expect(registerStatusChoices(CHEQUE_TYPE_RECEIVED, true)).toEqual([0, 1, 2, 3, 4, 5, 6]);
    expect(registerStatusChoices(CHEQUE_TYPE_RECEIVED, false)).toEqual([0, 1, 2, 3, 4, 5, 6]);
    expect(registerStatusChoices(CHEQUE_TYPE_ISSUED, true)).toEqual([0, 3]);
    expect(registerStatusChoices(CHEQUE_TYPE_ISSUED, false)).toEqual([0, 1, 2, 3]);
  });
});

describe("titles and GL status", () => {
  it("finds the status document title by doctype key for the reference table", () => {
    expect(chequeDocTitleByKind("chequedeposit")?.[0]).toBe("cheque_deposit");
    expect(chequeDocTitleByKind("chequeissuedcancel")?.[0]).toBe("cheque_issued_cancel");
    expect(chequeDocTitleByKind("paid")).toBeUndefined();
  });

  it("labels the four backend gl_status values and falls back to not-posted", () => {
    expect(Object.keys(CHEQUE_GL_STATUS).sort()).toEqual(["no_mapping", "none", "not_required", "posted"]);
    expect(chequeGlStatus("posted").tone).toBe("success");
    expect(chequeGlStatus("no_mapping").tone).toBe("warning");
    expect(chequeGlStatus("something_new").label[0]).toBe("cheque_gl_none");
  });
});

import { beforeEach, describe, expect, it, vi } from "vitest";

const apiFetch = vi.fn();
vi.mock("./client-auth-session", () => ({ apiFetch: (...args: unknown[]) => apiFetch(...args) }));

import { BANK_CODE_KINDS, bankCodeKindFromRoute, isBankCodeKind } from "./bank-code-kinds";
import {
  ChequeApiError,
  changeBankBookNumber,
  createBankBook,
  createBankCode,
  deleteBankBook,
  deleteBankCode,
  listBankBooks,
  listBankCodes,
  listBankContacts,
  listBankGLAccounts,
  newRequestId,
  saveBankContacts,
  updateBankBook,
  updateBankCode,
  listPettyCashTrans,
  createPettyCashTrans,
  deletePettyCashTrans,
} from "./cheque-api";

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

beforeEach(() => apiFetch.mockReset());

describe("bank code kinds", () => {
  it("maps screen routes to kinds and rejects anything else", () => {
    expect(BANK_CODE_KINDS).toHaveLength(5);
    for (const kind of BANK_CODE_KINDS) expect(bankCodeKindFromRoute(`/settings/bank-codes/${kind}`)).toBe(kind);
    expect(bankCodeKindFromRoute("/settings/bank-codes/branch")).toBeNull();
    expect(bankCodeKindFromRoute("/settings/bank-codes/")).toBeNull();
    expect(bankCodeKindFromRoute("/settings/currency")).toBeNull();
    expect(isBankCodeKind("bank_branch")).toBe(true);
    expect(isBankCodeKind("constructor")).toBe(false);
  });
});

describe("listBankCodes", () => {
  it("sends only the filters that are set and asks in the UI language", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0, page: 1, limit: 50 }));

    await listBankCodes("bank_branch", { q: "  สีลม ", bankCode: "BBL", activeOnly: true, page: 2 }, "th");

    const [path, init] = apiFetch.mock.calls[0];
    expect(path).toBe(`/api/cheque/codes/bank_branch?q=${encodeURIComponent("สีลม")}&bank_code=BBL&active=1&page=2`);
    expect(init.headers["Accept-Language"]).toBe("th");
  });

  it("omits the query string when nothing is filtered", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0, page: 1, limit: 50 }));
    await listBankCodes("bank", { q: "   " }, "en");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/codes/bank");
  });
});

describe("code commands", () => {
  const saved = { kind: "bank", code: "BBL", name: "ธนาคารกรุงเทพ", bank_code: "", user_group: "", is_active: true, version: 1 };

  it("create / update / delete post one command each and return the saved row", async () => {
    apiFetch.mockImplementation(async () => json({ success: true, data: saved }));

    await createBankCode({ kind: "bank", code: "BBL", name: "ธนาคารกรุงเทพ", is_active: true }, "th");
    await updateBankCode({ kind: "bank", code: "BBL", name: "ธนาคารกรุงเทพ", version: 1 }, "th");
    expect(await deleteBankCode({ kind: "bank", code: "BBL", version: 2 }, "th")).toEqual(saved);

    const bodies = apiFetch.mock.calls.map(([, init]) => JSON.parse(init.body));
    expect(apiFetch.mock.calls.every(([path, init]) => path === "/api/cheque/command" && init.method === "POST")).toBe(true);
    expect(bodies[0]).toEqual({ resource: "codes", action: "create", kind: "bank", code: "BBL", name: "ธนาคารกรุงเทพ", is_active: true });
    expect(bodies[1]).toMatchObject({ resource: "codes", action: "update", version: 1 });
    expect(bodies[2]).toEqual({ resource: "codes", action: "delete", kind: "bank", code: "BBL", version: 2 });
  });
});

describe("bank contacts", () => {
  it("lists by bank and branch", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [{ contact_code: "C1" }], total: 1 }));
    expect(await listBankContacts("BBL", "0123", "th")).toHaveLength(1);
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/bank-contacts?bank_code=BBL&bank_branch_code=0123");
  });

  it("saves the whole set in one command, including an empty set", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0 }));
    await saveBankContacts("BBL", "0123", [], "th");
    expect(JSON.parse(apiFetch.mock.calls[0][1].body)).toEqual({
      resource: "bank-contacts", action: "save", bank_code: "BBL", bank_branch_code: "0123", contacts: [],
    });
  });
});

describe("errors", () => {
  it("turns a relayed user error (HTTP 200 + success:false) into ChequeApiError with the real status and server message", async () => {
    apiFetch.mockResolvedValue(json({ success: false, code: "bank_code_err_duplicate", message: "มีรหัส BBL แล้ว", field: "code", statuscode: 409 }));

    const error = await createBankCode({ kind: "bank", code: "BBL", name: "x" }, "th").catch((cause) => cause);

    expect(error).toBeInstanceOf(ChequeApiError);
    expect(error).toMatchObject({ status: 409, code: "bank_code_err_duplicate", message: "มีรหัส BBL แล้ว", field: "code" });
  });

  it("keeps the HTTP status of a 403 and survives a non-JSON failure body", async () => {
    apiFetch.mockResolvedValueOnce(json({ success: false, code: "cheque_err_read_forbidden", message: "ไม่มีสิทธิ์", statuscode: 403 }, 403));
    await expect(listBankCodes("bank", {}, "th")).rejects.toMatchObject({ status: 403, code: "cheque_err_read_forbidden" });

    apiFetch.mockResolvedValueOnce(new Response("<html>bad gateway</html>", { status: 502 }));
    await expect(listBankCodes("bank", {}, "th")).rejects.toMatchObject({ status: 502, code: "", message: "" });
  });
});

describe("bank book (GL resource bank-accounts)", () => {
  const input = {
    bank_account_code: "KBANK-01",
    account_number: "123-4-56789-0",
    account_name: "บริษัท รุ่งเรืองการค้า จำกัด",
    bank_code: "KBANK",
    bank_branch_code: "0123",
    account_type_code: "CURRENT",
    gl_account_code: "11120",
    currency_code: "THB" as const,
    is_active: true,
  };
  const sentBody = () => JSON.parse(apiFetch.mock.calls[0][1].body);

  it("lists through the GL BFF with only the filters that are set", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { items: [], total: 0, page: 1, limit: 50 } }));
    await listBankBooks({ q: " kbank ", page: 2, limit: 50 }, "th");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/gl/bank-accounts?q=kbank&page=2&limit=50");
    expect(apiFetch.mock.calls[0][1].headers["Accept-Language"]).toBe("th");
  });

  it("sends create/update/delete/change-number as GL commands with a UUID request id", async () => {
    apiFetch.mockImplementation(async () => json({ success: true, data: { id: "KBANK-01", version: 5 } }));

    await createBankBook(input, "th");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/gl/command");
    expect(sentBody()).toMatchObject({ resource: "bank-accounts", action: "create", bankaccount: input });
    expect(sentBody().requestid).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
    expect(sentBody()).not.toHaveProperty("version");

    apiFetch.mockClear();
    expect(await updateBankBook(input, 4, "en")).toEqual({ id: "KBANK-01", version: 5 });
    expect(sentBody()).toMatchObject({ resource: "bank-accounts", action: "update", id: "KBANK-01", version: 4, bankaccount: input });

    apiFetch.mockClear();
    await deleteBankBook("KBANK-01", 5, "th");
    expect(sentBody()).toMatchObject({ resource: "bank-accounts", action: "delete", id: "KBANK-01", version: 5 });
    expect(sentBody()).not.toHaveProperty("bankaccount");

    apiFetch.mockClear();
    await changeBankBookNumber("KBANK-01", 5, "999-9-99999-9", "th");
    expect(sentBody()).toMatchObject({ resource: "bank-accounts", action: "change-account-number", id: "KBANK-01", version: 5, newaccountnumber: "999-9-99999-9" });
    expect(sentBody()).not.toHaveProperty("bankaccount");
  });

  it("uses a fresh request id for every command", () => {
    const ids = new Set(Array.from({ length: 50 }, () => newRequestId()));
    expect(ids.size).toBe(50);
  });

  it("surfaces a user error relayed as HTTP 200 + success:false with its field", async () => {
    apiFetch.mockResolvedValue(json({ success: false, statuscode: 409, code: "bank_book_account_number_duplicate", message: "เลขที่บัญชี 123 ซ้ำกับสมุด KBANK-02", field: "new_account_number" }));
    const error = await changeBankBookNumber("KBANK-01", 5, "123", "th").catch((cause: unknown) => cause);
    expect(error).toBeInstanceOf(ChequeApiError);
    expect(error).toMatchObject({ status: 409, code: "bank_book_account_number_duplicate", field: "new_account_number" });
  });

  it("offers only active, postable asset accounts for the book's GL account, named in the UI language", async () => {
    const account = (accountcode: string, accounttype: string, extra: Record<string, unknown> = {}) => ({
      accountcode,
      accounttype,
      allowposting: true,
      isactive: true,
      names: [{ code: "th", name: "ธนาคารกสิกรไทย" }, { code: "en", name: "Kasikorn Bank" }],
      ...extra,
    });
    apiFetch.mockResolvedValue(
      json({
        success: true,
        data: {
          items: [
            account("11120", "asset"),
            account("11000", "asset", { allowposting: false }),
            account("11130", "asset", { isactive: false }),
            account("41100", "revenue"),
            account("11140", "asset", { names: [{ code: "th", name: "ออมทรัพย์" }] }),
          ],
        },
      }),
    );
    expect(await listBankGLAccounts("en")).toEqual([
      { code: "11120", name: "Kasikorn Bank" },
      { code: "11140", name: "ออมทรัพย์" },
    ]);
    expect(apiFetch.mock.calls[0][0]).toBe("/api/gl/accounts?limit=1000");
  });
});

describe("petty cash transactions API", () => {
  it("lists petty cash transactions with query parameters", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [{ id: "t1", doc_no: "PC01", amount: "500.00" }] }));
    const result = await listPettyCashTrans({ petty_cash_code: "PC-01", from_date: "2026-10-01", to_date: "2026-10-31" }, "th");
    expect(result).toHaveLength(1);
    expect(apiFetch.mock.calls[0][0]).toContain("/api/cheque/petty-cash/trans?petty_cash_code=PC-01&from_date=2026-10-01&to_date=2026-10-31");
  });

  it("creates a petty cash transaction through command endpoint", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { id: "t2", doc_no: "PC02", trans_type: 1, amount: "1000.00" } }));
    const input = {
      doc_no: "PC02",
      doc_date: "2026-10-11",
      petty_cash_code: "PC-01",
      trans_type: 1,
      amount: "1000.00",
      description: "เบิกชดเชย",
    };
    const created = await createPettyCashTrans(input, "th");
    expect(created.id).toBe("t2");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/command");
    expect(JSON.parse(apiFetch.mock.calls[0][1].body)).toMatchObject({
      resource: "petty-cash-trans",
      action: "create",
      doc_no: "PC02",
      amount: "1000.00",
    });
  });

  it("deletes a petty cash transaction through command endpoint", async () => {
    apiFetch.mockResolvedValue(json({ success: true }));
    await deletePettyCashTrans("t2", "th");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/command");
    expect(JSON.parse(apiFetch.mock.calls[0][1].body)).toMatchObject({
      resource: "petty-cash-trans",
      action: "delete",
      id: "t2",
    });
  });
});

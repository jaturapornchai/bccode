import { beforeEach, describe, expect, it, vi } from "vitest";

const apiFetch = vi.fn();
vi.mock("./client-auth-session", () => ({ apiFetch: (...args: unknown[]) => apiFetch(...args) }));

import {
  ChequeApiError,
  cancelChequeDoc,
  cancelChequeRegister,
  createChequeDoc,
  createChequeRegister,
  getChequeDoc,
  getChequeRegister,
  listChequeDocs,
  listChequeRegisters,
  listSelectableCheques,
  updateChequeRegister,
  type ChequeDocInput,
  type ChequeRegisterInput,
} from "./cheque-api";

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
const sentBody = (call = 0) => JSON.parse(apiFetch.mock.calls[call][1].body);
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

beforeEach(() => apiFetch.mockReset());

const registerInput: ChequeRegisterInput = {
  cheque_type: 0,
  doc_no: "",
  doc_date: "2026-10-01",
  due_date: "2026-10-20",
  partner_code: "AR-0012",
  cheque_no: "1023456",
  bank_code: "BBL",
  bank_branch_code: "0123",
  book_code: "",
  status: 0,
  ext_status_code: "",
  currency_code: "THB",
  exchange_rate: "1",
  amount: "15000.00",
  description: "ชำระค่าสินค้างวดที่ 2",
  sale_code: "",
  department_code: "",
};

describe("register reads", () => {
  it("lists with only the filters that are set, trimming the search text and keeping status 0", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0, page: 1, limit: 50 }));

    await listChequeRegisters({ type: 1, status: 0, q: " 1023 ", page: 2, limit: 50 }, "th");

    const [path, init] = apiFetch.mock.calls[0];
    expect(path).toBe("/api/cheque/registers?type=1&status=0&q=1023&page=2&limit=50");
    expect(init.headers["Accept-Language"]).toBe("th");
  });

  it("leaves status out when the filter is 'all statuses'", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0, page: 1, limit: 50 }));
    await listChequeRegisters({ type: 0, status: null, q: "   " }, "en");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/registers?type=0");
  });

  it("encodes the id and turns a missing reference list into an empty one", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { id: "a b", cheque_no: "1", last_event: null } }));

    const detail = await getChequeRegister("a b", "th");

    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/registers/a%20b");
    expect(detail.references).toEqual([]);
  });
});

describe("selectable cheques for a status document", () => {
  it("sends the document date, book and cancel type 0 (on-hand), and the cleared-cheque tick as 1", async () => {
    apiFetch.mockImplementation(async () => json({ success: true, items: [], total: 0, page: 1, limit: 50 }));

    await listSelectableCheques({ transflag: 423, docDate: "2026-10-25", cancelType: 0 }, "th");
    await listSelectableCheques({ transflag: 426, docDate: "2026-10-25", book: "SCB-01", cancelPassed: true }, "th");

    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/registers/selectable?transflag=423&doc_date=2026-10-25&cancel_type=0");
    expect(apiFetch.mock.calls[1][0]).toBe("/api/cheque/registers/selectable?transflag=426&doc_date=2026-10-25&book=SCB-01&cancel_passed=1");
  });

  it("omits a false tick, a blank book and an unset cancel type", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0, page: 1, limit: 50 }));
    await listSelectableCheques({ transflag: 426, book: "", cancelPassed: false, cancelType: null, q: " 55 " }, "en");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/registers/selectable?transflag=426&q=55");
  });
});

describe("register commands", () => {
  it("create and update post to the one command endpoint, update carrying id, version and the D8 confirmation only when given", async () => {
    apiFetch.mockImplementation(async () => json({ success: true, data: { id: "r1" } }));

    await createChequeRegister(registerInput, "th");
    await updateChequeRegister({ ...registerInput, id: "r1", version: 3 }, false, "th");
    await updateChequeRegister({ ...registerInput, id: "r1", version: 3 }, true, "th");

    expect(apiFetch.mock.calls.every(([path, init]) => path === "/api/cheque/command" && init.method === "POST")).toBe(true);
    expect(sentBody(0)).toMatchObject({ resource: "registers", action: "create", cheque_no: "1023456", amount: "15000.00", currency_code: "THB", exchange_rate: "1" });
    expect(sentBody(1)).toMatchObject({ resource: "registers", action: "update", id: "r1", version: 3 });
    expect(sentBody(1)).not.toHaveProperty("confirm");
    expect(sentBody(2)).toMatchObject({ action: "update", confirm: true });
  });

  it("keeps the amount as the exact decimal text, never a number", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: {} }));
    await createChequeRegister({ ...registerInput, amount: "0.10" }, "th");
    expect(apiFetch.mock.calls[0][1].body).toContain('"amount":"0.10"');
  });

  it("removes a cheque from the register with 'cancel' and its version", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { id: "r1" } }));

    await cancelChequeRegister({ chequeType: 1, id: "r1", version: 4 }, "th");

    expect(sentBody()).toMatchObject({ resource: "registers", action: "cancel", cheque_type: 1, id: "r1", version: 4 });
  });

  it("uses a fresh UUID request id for every press so a retry after a timeout is the only duplicate", async () => {
    apiFetch.mockImplementation(async () => json({ success: true, data: {} }));

    await createChequeRegister(registerInput, "th");
    await createChequeRegister(registerInput, "th");

    const ids = [sentBody(0).requestid, sentBody(1).requestid];
    expect(ids[0]).toMatch(UUID);
    expect(ids[1]).toMatch(UUID);
    expect(ids[0]).not.toBe(ids[1]);
  });

  it("relays a referenced-cheque refusal (HTTP 200 + success:false, 409) as ChequeApiError so the screen can ask for confirmation", async () => {
    apiFetch.mockResolvedValue(
      json({ success: false, code: "cheque_err_referenced", message: "เช็คนี้ถูกอ้างอิงแล้ว ต้องการแก้ไขต่อหรือไม่", field: "", statuscode: 409 }),
    );

    const error = await updateChequeRegister({ ...registerInput, id: "r1", version: 1 }, false, "th").catch((cause) => cause);

    expect(error).toBeInstanceOf(ChequeApiError);
    expect(error).toMatchObject({ status: 409, code: "cheque_err_referenced", message: "เช็คนี้ถูกอ้างอิงแล้ว ต้องการแก้ไขต่อหรือไม่" });
  });
});

describe("status documents", () => {
  const docInput: ChequeDocInput = {
    transflag: 420,
    doc_no: "",
    doc_date: "2026-10-25",
    book_code: "SCB-01",
    description: "",
    lines: [
      { cheque_id: "c1", fee_amount: "" },
      { cheque_id: "c2", fee_amount: "10.25" },
    ],
    totals: { total: "17500.50", fee: "10.25", net: "17490.25" },
  };

  it("lists one document type with the filters that are set", async () => {
    apiFetch.mockResolvedValue(json({ success: true, items: [], total: 0, page: 1, limit: 50 }));
    await listChequeDocs(421, { q: " CQRC ", status: "cancelled", page: 3 }, "th");
    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/docs/421?q=CQRC&status=cancelled&page=3");
  });

  it("encodes the document number and always returns a lines array", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { doc_no: "CQRC 6910/0001", transflag: 420 } }));

    const doc = await getChequeDoc(420, "CQRC 6910/0001", "th");

    expect(apiFetch.mock.calls[0][0]).toBe("/api/cheque/docs/420/CQRC%206910%2F0001");
    expect(doc.lines).toEqual([]);
  });

  it("creates with only cheque ids, fees and the screen's totals, and leaves cancel fields off when not set", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { doc_no: "CQRC6910-0001" } }));

    await createChequeDoc(docInput, "th");

    const body = sentBody();
    expect(body).toMatchObject({ resource: "docs", action: "create", transflag: 420, book_code: "SCB-01", totals: docInput.totals });
    expect(body.lines).toEqual(docInput.lines);
    expect(body).not.toHaveProperty("cancel_type");
    expect(body).not.toHaveProperty("cancel_passed");
  });

  it("sends the 423 cancel type, including 0, and the 426 tick when the screen sets them", async () => {
    apiFetch.mockImplementation(async () => json({ success: true, data: {} }));

    await createChequeDoc({ ...docInput, transflag: 423, book_code: "", cancel_type: 0 }, "th");
    await createChequeDoc({ ...docInput, transflag: 426, cancel_passed: true }, "th");

    expect(sentBody(0)).toMatchObject({ transflag: 423, cancel_type: 0, book_code: "" });
    expect(sentBody(1)).toMatchObject({ transflag: 426, cancel_passed: true });
  });

  it("cancels with the document number, version and the reason typed by the user", async () => {
    apiFetch.mockResolvedValue(json({ success: true, data: { doc_no: "CQRC6910-0001", status: "cancelled" } }));

    const result = await cancelChequeDoc({ transflag: 420, docNo: "CQRC6910-0001", version: 2, reason: "ระบุวันที่เอกสารผิด" }, "th");

    expect(sentBody()).toMatchObject({ resource: "docs", action: "cancel", transflag: 420, doc_no: "CQRC6910-0001", version: 2, reason: "ระบุวันที่เอกสารผิด" });
    expect(sentBody().requestid).toMatch(UUID);
    expect(result.status).toBe("cancelled");
  });

  it("relays a rule refusal, such as a cheque that is no longer in the right status, with the server's Thai message", async () => {
    apiFetch.mockResolvedValue(
      json({ success: false, code: "cheque_err_status_mismatch", message: "เช็คเลขที่ 1023456 ไม่อยู่ในสถานะที่เลือกได้", field: "lines", statuscode: 409 }),
    );

    await expect(createChequeDoc(docInput, "th")).rejects.toMatchObject({
      status: 409,
      code: "cheque_err_status_mismatch",
      field: "lines",
      message: "เช็คเลขที่ 1023456 ไม่อยู่ในสถานะที่เลือกได้",
    });
  });
});

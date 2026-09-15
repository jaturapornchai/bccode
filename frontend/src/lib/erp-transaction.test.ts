import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ERP_MODULE_CONFIGS,
  isErpTransactionRoute,
  getErpModuleConfig,
  fetchErpTransactions,
  saveErpTransaction,
  deleteErpTransaction,
  type ErpTransactionDoc,
} from "./erp-transaction";

describe("ERP Transaction domain configs and operations", () => {
  it("provides comprehensive coverage across all 6 ERP core operational domains", () => {
    const domains = new Set(ERP_MODULE_CONFIGS.map((c) => c.domain));
    expect(domains).toContain("inventory");
    expect(domains).toContain("sales");
    expect(domains).toContain("purchase");
    expect(domains).toContain("ar");
    expect(domains).toContain("ap");
    expect(domains).toContain("cash-bank");
    expect(ERP_MODULE_CONFIGS.length).toBeGreaterThanOrEqual(25);
  });

  it("correctly identifies ERP transaction routes", () => {
    // Inventory
    expect(isErpTransactionRoute("/transaction/stockbalance")).toBe(true);
    expect(isErpTransactionRoute("/transaction/stockreceiveproduct")).toBe(true);
    expect(isErpTransactionRoute("/transaction/stockpickupproduct")).toBe(true);
    expect(isErpTransactionRoute("/transaction/stocktransfer")).toBe(true);
    expect(isErpTransactionRoute("/transaction/adjust")).toBe(true);

    // Sales
    expect(isErpTransactionRoute("/transaction/quotation")).toBe(true);
    expect(isErpTransactionRoute("/transaction/saleorder")).toBe(true);
    expect(isErpTransactionRoute("/transaction/sale")).toBe(true);
    expect(isErpTransactionRoute("/transaction/taxinvoice")).toBe(true);
    expect(isErpTransactionRoute("/transaction/salereturn")).toBe(true);

    // Purchase
    expect(isErpTransactionRoute("/transaction/purchaserequisition")).toBe(true);
    expect(isErpTransactionRoute("/transaction/purchaseorder")).toBe(true);
    expect(isErpTransactionRoute("/transaction/purchase")).toBe(true);
    expect(isErpTransactionRoute("/transaction/expense")).toBe(true);

    // AR
    expect(isErpTransactionRoute("/debtorbeginningbalance")).toBe(true);
    expect(isErpTransactionRoute("/transaction/billingnote")).toBe(true);
    expect(isErpTransactionRoute("/transaction/paid")).toBe(true);

    // AP
    expect(isErpTransactionRoute("/creditorbeginningbalance")).toBe(true);
    expect(isErpTransactionRoute("/transaction/paymentvoucher")).toBe(true);
    expect(isErpTransactionRoute("/transaction/pay")).toBe(true);

    // Cash & Bank
    expect(isErpTransactionRoute("/transaction/accounttransfer")).toBe(true);
    expect(isErpTransactionRoute("/transaction/chequereceived")).toBe(true);
    expect(isErpTransactionRoute("/transaction/chequeissued")).toBe(true);
    expect(isErpTransactionRoute("/transaction/directoradvance")).toBe(true);

    // Non-transaction route
    expect(isErpTransactionRoute("/product")).toBe(false);
    expect(isErpTransactionRoute("/nonexistent")).toBe(false);
  });

  it("retrieves module config with titles and counterparty labels", () => {
    const saleConfig = getErpModuleConfig("/transaction/saleorder");
    expect(saleConfig).toBeDefined();
    expect(saleConfig?.domain).toBe("sales");
    expect(saleConfig?.title.th).toContain("ใบสั่งขาย");
    expect(saleConfig?.counterpartyLabel.th).toBe("ลูกค้า");
    expect(saleConfig?.defaultDocPrefix).toBe("SO");

    const poConfig = getErpModuleConfig("/transaction/purchaseorder");
    expect(poConfig).toBeDefined();
    expect(poConfig?.domain).toBe("purchase");
    expect(poConfig?.title.th).toContain("ใบสั่งซื้อ");
    expect(poConfig?.counterpartyLabel.th).toBe("ผู้จำหน่าย");
    expect(poConfig?.defaultDocPrefix).toBe("PO");
  });

  const config = getErpModuleConfig("/transaction/quotation")!;

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const mockResponse = (ok: boolean, status: number, body: unknown): Response =>
    ({ ok, status, json: async () => body }) as unknown as Response;

  const makeDoc = (id: string): ErpTransactionDoc =>
    ({ id } as unknown as ErpTransactionDoc);

  it("fetchErpTransactions returns items and total on 200", async () => {
    const doc = makeDoc("DOC1");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        mockResponse(true, 200, { data: [doc], pagination: { total: 1 } }),
      ),
    );

    const result = await fetchErpTransactions(config);

    expect(result.items).toHaveLength(1);
    expect(result.total).toBe(1);
    expect(result.error).toBeUndefined();
  });

  it("fetchErpTransactions returns unauthorized on 401", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 401, {})),
    );

    const result = await fetchErpTransactions(config);

    expect(result.error).toBe("unauthorized");
    expect(result.items).toEqual([]);
  });

  it("fetchErpTransactions returns load_failed on 500", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 500, {})),
    );

    const result = await fetchErpTransactions(config);

    expect(result.error).toBe("load_failed");
    expect(result.items).toEqual([]);
  });

  it("fetchErpTransactions returns connection_error when fetch throws", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("network down");
      }),
    );

    const result = await fetchErpTransactions(config);

    expect(result.error).toBe("connection_error");
  });

  it("saveErpTransaction returns save_success on 200", async () => {
    const doc = makeDoc("DOC1");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(true, 200, { data: doc })),
    );

    const result = await saveErpTransaction(config, doc, false);

    expect(result.success).toBe(true);
    expect(result.message).toBe("save_success");
  });

  it("saveErpTransaction returns save_failed on 500 (no fake success)", async () => {
    const doc = makeDoc("DOC1");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 500, {})),
    );

    const result = await saveErpTransaction(config, doc, false);

    expect(result.success).toBe(false);
    expect(result.message).toBe("save_failed");
  });

  it("saveErpTransaction returns connection_error when fetch throws", async () => {
    const doc = makeDoc("DOC1");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("network down");
      }),
    );

    const result = await saveErpTransaction(config, doc, false);

    expect(result.success).toBe(false);
    expect(result.message).toBe("connection_error");
  });

  it("deleteErpTransaction returns delete_success on 200", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(true, 200, {})),
    );

    const result = await deleteErpTransaction(config, "DOC1");

    expect(result.success).toBe(true);
    expect(result.message).toBe("delete_success");
  });

  it("deleteErpTransaction returns delete_failed on 500", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 500, {})),
    );

    const result = await deleteErpTransaction(config, "DOC1");

    expect(result.success).toBe(false);
    expect(result.message).toBe("delete_failed");
  });

  it("saveErpTransaction uses POST for create and PUT with id for edit", async () => {
    const fetchMock = vi.fn(async () =>
      mockResponse(true, 200, { data: makeDoc("NEW") }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await saveErpTransaction(config, makeDoc(""), false);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [createUrl, createInit] = fetchMock.mock.calls[0] as unknown as [
      string,
      RequestInit,
    ];
    expect(createInit.method).toBe("POST");
    expect(createUrl.endsWith(config.apiPath)).toBe(true);

    fetchMock.mockClear();

    await saveErpTransaction(config, makeDoc("ABC"), true);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [editUrl, editInit] = fetchMock.mock.calls[0] as unknown as [
      string,
      RequestInit,
    ];
    expect(editInit.method).toBe("PUT");
    expect(editUrl.endsWith("ABC")).toBe(true);
  });
});

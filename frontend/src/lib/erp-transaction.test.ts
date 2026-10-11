import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ERP_MODULE_CONFIGS,
  isErpTransactionRoute,
  getErpModuleConfig,
  fetchErpTransactions,
  saveErpTransaction,
  deleteErpTransaction,
  fetchErpDocEvents,
  approveBatchErpDocs,
  unapproveBatchErpDocs,
  cancelBatchErpDocs,
  erpDocLifecycle,
  erpLineAmount,
  erpDocTotals,
  erpDocForSave,
  erpDueDate,
  erpBeforeVatAmount,
  erpMoneyDisplay,
  type ErpTransactionDoc,
} from "./erp-transaction";
import { ERP_AUTO_POSTING_MODULES, ERP_DUE_DATE_MODULES, ERP_TRANSACTION_MODULES } from "./erp-transaction-modules";
import { flattenMenuItems } from "./menu-data";
import { setupTestAuthSession } from "./test-auth-session";

setupTestAuthSession();

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
    expect(isErpTransactionRoute("/transaction/pay")).toBe(true);

    // Cash & Bank
    expect(isErpTransactionRoute("/transaction/accounttransfer")).toBe(true);
    expect(isErpTransactionRoute("/transaction/chequereceived")).toBe(true);
    expect(isErpTransactionRoute("/transaction/chequeissued")).toBe(true);

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
    expect(saleConfig?.apiPath).toBe("sale-order");

    const poConfig = getErpModuleConfig("/transaction/purchaseorder");
    expect(poConfig).toBeDefined();
    expect(poConfig?.domain).toBe("purchase");
    expect(poConfig?.title.th).toContain("ใบสั่งซื้อ");
    expect(poConfig?.counterpartyLabel.th).toBe("เจ้าหนี้");
    expect(poConfig?.apiPath).toBe("purchase-order");
  });

  // เมนูที่ backend ไม่มีเอกสารรองรับ ถูกถอด config ออก → เป็นจอ "รอพัฒนา" แทนจอที่บันทึกไม่ได้
  it("has no config for menus without a backend document type", () => {
    for (const route of [
      "/transaction/purchasereceiptadjustment",
      "/transaction/etax",
      "/transaction/withholdingtax",
      "/banking/petty-cash-limit",
      "/banking/contacts",
    ]) {
      expect(isErpTransactionRoute(route)).toBe(false);
    }
  });

  it("maps every screen to a backend registry key, one document type per screen", () => {
    const paths = ERP_MODULE_CONFIGS.map((c) => c.apiPath);
    for (const path of paths) {
      expect(path).not.toContain("/");
      expect(ERP_TRANSACTION_MODULES).toContain(path);
    }
    // count-sheet and stockcount are the same Champ document (stock-count);
    // purchasereducedebt and purchasedebitnote are Champ document siblings (apcreditnote);
    // rfqpricetable and rfq are Champ document siblings (rfq)
    const shared = [...new Set(paths.filter((path, i) => paths.indexOf(path) !== i))].sort();
    expect(shared).toEqual(["apcreditnote", "rfq", "stock-count"].sort());
    expect(getErpModuleConfig("/inventory/count-sheet")?.apiPath).toBe("stock-count");
    expect(getErpModuleConfig("/transaction/stockcount")?.apiPath).toBe("stock-count");
    expect(getErpModuleConfig("/transaction/purchasereducedebt")?.apiPath).toBe("apcreditnote");
    expect(getErpModuleConfig("/transaction/rfqpricetable")?.apiPath).toBe("rfq");
    // route spelling is inverted vs the Thai label for these two: the label decides the type
    expect(getErpModuleConfig("/transaction/purchasedebitnote")?.apiPath).toBe("apcreditnote");
    expect(getErpModuleConfig("/transaction/purchasecreditnote")?.apiPath).toBe("apdebitnote");
    expect(getErpModuleConfig("/transaction/stockpickupproduct")?.apiPath).toBe("stock-pickup-product");
  });

  it("keeps the frontend module list identical to the backend doctype registry", () => {
    const source = readFileSync(resolve(process.cwd(), "../backend/internal/doctype/doctype.go"), "utf8");
    const registryKeys = [...source.matchAll(/^\t\{"([a-z-]+)", /gm)].map((match) => match[1]);
    expect(registryKeys).toHaveLength(63);
    expect([...ERP_TRANSACTION_MODULES].sort()).toEqual([...registryKeys].sort());
    expect(new Set(ERP_TRANSACTION_MODULES).size).toBe(ERP_TRANSACTION_MODULES.length);
  });

  it("limits the auto-posting list to registry types the backend posts or moves stock for on save", () => {
    const source = readFileSync(resolve(process.cwd(), "../backend/internal/doctype/doctype.go"), "utf8");
    const stockMovers = [...source.matchAll(/^	{"([a-z-]+)", [A-Za-z0-9]+, "[A-Z]+", (StockIn|StockOut),/gm)].map((match) => match[1]);
    const posting = ["expense", "paid", "pay", "purchase", "purchase-return", "sale-invoice", "sale-invoice-return",
      "apdebitnote", "apcreditnote", "apotherdebt"];
    for (const key of ERP_AUTO_POSTING_MODULES) expect(ERP_TRANSACTION_MODULES).toContain(key);
    expect(stockMovers.length).toBeGreaterThan(0);
    expect([...ERP_AUTO_POSTING_MODULES].sort()).toEqual([...new Set([...posting, ...stockMovers])].sort());
  });

  it("has all-column language rows for the document-number and store-only notices", () => {
    const lines = readFileSync(resolve(process.cwd(), "../backend/assets/language/languages.tsv"), "utf8").split(/\r?\n/);
    for (const key of ["docno_auto_placeholder", "erp_store_only_notice", "erp_gl_notice_posted", "erp_gl_notice_no_mapping", "erp_gl_notice_no_amount"]) {
      const row = lines.find((line) => line.startsWith(key + "\t"));
      expect(row, key).toBeDefined();
      const columns = row!.split("\t");
      expect(columns).toHaveLength(13);
      expect(columns.every((column) => column.trim() !== "")).toBe(true);
    }
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

  it("fetchErpTransactions maps 403 to the read-forbidden key, not a re-login", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 403, {})),
    );

    const result = await fetchErpTransactions(config);

    expect(result.error).toBe("erp_err_read_forbidden");
    expect(result.items).toEqual([]);
  });

  it("save and delete map 403 to the action-forbidden key", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 403, {})),
    );

    const saved = await saveErpTransaction(config, { docno: "DOC1" }, false);
    const deleted = await deleteErpTransaction(config, "DOC1");

    expect(saved).toMatchObject({ success: false, message: "erp_err_action_forbidden" });
    expect(deleted).toMatchObject({ success: false, message: "erp_err_action_forbidden" });
  });

  it("asks the backend to answer in the language chosen on screen, not the browser's", async () => {
    const calls: { headers?: HeadersInit }[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (_url: string, init?: { headers?: HeadersInit }) => {
        calls.push({ headers: init?.headers });
        return mockResponse(true, 200, { success: true, data: {} });
      }),
    );

    await saveErpTransaction(config, { docno: "DOC1" }, false, "en");
    await deleteErpTransaction(config, "DOC1", "en");
    await saveErpTransaction(config, { docno: "DOC1" }, false);

    expect(calls.map((c) => new Headers(c.headers).get("accept-language"))).toEqual(["en", "en", "th"]);
  });

  it("has all-column language rows for both forbidden messages", () => {
    const tsv = readFileSync(resolve(process.cwd(), "../backend/assets/language/languages.tsv"), "utf8");
    for (const key of ["erp_err_read_forbidden", "erp_err_action_forbidden"]) {
      const row = tsv.split(/\r?\n/).find((line) => line.startsWith(key + "\t"));
      expect(row, key).toBeTruthy();
      const cells = (row ?? "").split("\t");
      expect(cells).toHaveLength(13);
      expect(cells.every((cell) => cell.trim() !== "")).toBe(true);
    }
  });

  it("opens every ERP route under the menu id the backend checks for that document type", () => {
    const source = readFileSync(resolve(process.cwd(), "../backend/internal/doctype/doctype.go"), "utf8");
    const screenOf = new Map<string, string>(
      [...source.matchAll(/^\t\{"([a-z-]+)", [A-Za-z0-9]+, "[A-Z-]+", Stock[A-Za-z]+, "([a-z-]+)"\}/gm)].map(
        (match) => [match[1], match[2]] as [string, string],
      ),
    );
    expect(screenOf.size).toBe(63);
    // types reachable from multiple menu entries (doctype.screenAliases)
    const aliases: Record<string, string[]> = {
      "stock-count": ["stock-count-sheet"],
      "apcreditnote": ["purchase-reduce-debt"],
      "rfq": ["rfq-price-table"],
    };
    const menuIdOfRoute = new Map(flattenMenuItems().map((item) => [item.route, item.id]));
    for (const config of ERP_MODULE_CONFIGS) {
      const menuId = menuIdOfRoute.get(config.route);
      const allowed = [screenOf.get(config.apiPath), ...(aliases[config.apiPath] ?? [])];
      expect(allowed, `${config.route} -> ${config.apiPath}`).toContain(menuId);
    }
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

  it("saveErpTransaction and deleteErpTransaction never count a relayed ledger rejection (HTTP 200 + success:false) as success", async () => {
    const rejection = { success: false, code: "erp_doc_settled", message: "เอกสาร PUR6910-0001 มีการตัดหนี้แล้ว จึงแก้ไขหรือลบไม่ได้" };
    vi.stubGlobal("fetch", vi.fn(async () => mockResponse(true, 200, rejection)));

    const saved = await saveErpTransaction(config, makeDoc("PUR6910-0001"), true);
    expect(saved).toMatchObject({ success: false, message: "save_failed", detail: rejection.message });

    const deleted = await deleteErpTransaction(config, "7");
    expect(deleted).toMatchObject({ success: false, message: "delete_failed", detail: rejection.message });
  });

  it("saveErpTransaction passes the posting result of the backend through to the screen", async () => {
    const doc = { ...makeDoc("PUR6910-0001"), glstatus: "no_mapping" };
    vi.stubGlobal("fetch", vi.fn(async () => mockResponse(true, 200, { success: true, data: doc })));

    const result = await saveErpTransaction(config, makeDoc("PUR6910-0001"), false);
    expect(result.success).toBe(true);
    expect(result.data?.glstatus).toBe("no_mapping");
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

  it("fetchErpDocEvents correctly parses events list from data array and error states", async () => {
    // 1. Success with data array
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        mockResponse(true, 200, {
          success: true,
          data: [
            {
              id: 1,
              docno: "PO001",
              action: "approve",
              from_status: "",
              to_status: "approved",
              actor: "manager",
              created_at: "2026-10-10T08:00:00Z",
            },
          ],
        }),
      ),
    );

    const res1 = await fetchErpDocEvents("purchase-order", "PO001");
    expect(res1.success).toBe(true);
    expect(res1.events.length).toBe(1);
    expect(res1.events[0].action).toBe("approve");
    expect(res1.events[0].actor).toBe("manager");

    // 2. Failed request
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => mockResponse(false, 403, {})),
    );
    const res2 = await fetchErpDocEvents("purchase-order", "PO001");
    expect(res2.success).toBe(false);
    expect(res2.events.length).toBe(0);
    expect(res2.error).toBe("erp_err_read_forbidden");
  });
});

describe("ERP batch approve / unapprove / cancel helpers", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const reply = (ok: boolean, status: number, body: unknown): Response =>
    ({ ok, status, json: async () => body }) as unknown as Response;

  const stubFetch = (response: Response) => {
    const fetchMock = vi.fn(async (_url: string, _init?: RequestInit) => response);
    vi.stubGlobal("fetch", fetchMock);
    return fetchMock;
  };

  it("parses the per-document array and counts successes on partial failure", async () => {
    const fetchMock = stubFetch(
      reply(true, 200, {
        success: true,
        data: [
          { id: "PR-1", success: true },
          { id: "PR-2", success: false, code: "erp_doc_not_pending", message: "เอกสารนี้ไม่ได้รออนุมัติ" },
          { id: "PR-3", success: true },
        ],
      }),
    );

    const res = await approveBatchErpDocs("purchase-requisition", ["PR-1", "PR-2", "PR-3"]);

    expect(res.success).toBe(true);
    expect(res.approved).toBe(2);
    expect(res.total).toBe(3);
    expect(res.items?.filter((item) => !item.success)).toEqual([
      { id: "PR-2", success: false, code: "erp_doc_not_pending", message: "เอกสารนี้ไม่ได้รออนุมัติ" },
    ]);
    expect(String(fetchMock.mock.calls[0][0])).toBe("/api/erp-transaction/purchase-requisition/approve-batch");
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ docnos: ["PR-1", "PR-2", "PR-3"] });
  });

  it("reports zero approved when every document in a 200 response failed", async () => {
    stubFetch(
      reply(true, 200, {
        success: true,
        data: [
          { id: "PO-1", success: false, code: "erp_doc_locked" },
          { id: "PO-2", success: false, code: "erp_doc_locked" },
        ],
      }),
    );

    const res = await approveBatchErpDocs("purchase-order", ["PO-1", "PO-2"]);

    expect(res.approved).toBe(0);
    expect(res.total).toBe(2);
  });

  it("posts unapprove with the reason and cancel with the mandatory reason", async () => {
    const fetchMock = stubFetch(reply(true, 200, { success: true, data: [{ id: "SO-1", success: true }] }));

    const unapproved = await unapproveBatchErpDocs("sale-order", ["SO-1"], "ขอทบทวนยอด");
    expect(unapproved.unapproved).toBe(1);
    expect(String(fetchMock.mock.calls[0][0])).toBe("/api/erp-transaction/sale-order/unapprove-batch");
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).toEqual({ docnos: ["SO-1"], reason: "ขอทบทวนยอด" });

    const cancelled = await cancelBatchErpDocs("sale-order", ["SO-1"], "ลูกค้ายกเลิก");
    expect(cancelled.cancelled).toBe(1);
    expect(String(fetchMock.mock.calls[1][0])).toBe("/api/erp-transaction/sale-order/cancel-batch");
    expect(JSON.parse(String(fetchMock.mock.calls[1][1]?.body))).toEqual({ docnos: ["SO-1"], reason: "ลูกค้ายกเลิก" });
  });

  it("returns a failure key for forbidden responses and for success:false bodies", async () => {
    stubFetch(reply(false, 403, {}));
    const forbidden = await cancelBatchErpDocs("quotation", ["QT-1"], "เหตุผล");
    expect(forbidden.success).toBe(false);
    expect(forbidden.message).toBe("erp_err_action_forbidden");

    stubFetch(reply(true, 200, { success: false, message: "ไม่สำเร็จ" }));
    const rejected = await cancelBatchErpDocs("quotation", ["QT-1"], "เหตุผล");
    expect(rejected.success).toBe(false);
    expect(rejected.message).toBe("action_failed");
  });

  it("maps document lifecycle: cancelled beats approved beats draft", () => {
    expect(erpDocLifecycle({ status: 0, iscancel: false, approvalstatus: "" })).toBe("draft");
    expect(erpDocLifecycle({})).toBe("draft");
    expect(erpDocLifecycle({ status: 1, iscancel: false, approvalstatus: "approved" })).toBe("approved");
    expect(erpDocLifecycle({ approvalstatus: "approved" })).toBe("approved");
    expect(erpDocLifecycle({ status: 2 })).toBe("cancelled");
    expect(erpDocLifecycle({ status: 1, iscancel: true, approvalstatus: "approved" })).toBe("cancelled");
  });
});

// Rule 20: ยอดเงินเป็นข้อความทศนิยมจาก backend — ตัวอย่างยอดระหว่างแก้ต้องตรงตัว ไม่ผ่าน Number()
describe("ERP document amounts are exact decimal strings", () => {
  it("adds 0.1 + 0.2 exactly and multiplies without float drift", () => {
    expect(erpLineAmount({ qty: "3", price: "0.1" })).toBe("0.3");
    const details = [
      { sumamount: erpLineAmount({ qty: "1", price: "0.1" }) },
      { sumamount: erpLineAmount({ qty: "1", price: "0.2" }) },
    ];
    expect(erpDocTotals(details, 0)).toEqual({ totalbeforevat: "0.3", totalvatvalue: "0", totalamount: "0.3" });
    expect(erpLineAmount({ qty: "0.12345678", price: "1234567.89" })).toBe("152415.7763907942");
  });

  it("keeps the screen's rounding rule: lines unrounded, VAT rounded half up at document level to 2 places", () => {
    // 100.5 × 7% = 7.035 → 7.04 (ขอบ .005 ปัดขึ้น); 100.49 × 7% = 7.0343 → 7.03
    expect(erpDocTotals([{ sumamount: "100.5" }], 1, "7")).toEqual({ totalbeforevat: "100.5", totalvatvalue: "7.04", totalamount: "107.54" });
    expect(erpDocTotals([{ sumamount: "100.49" }], 1, "7")).toEqual({ totalbeforevat: "100.49", totalvatvalue: "7.03", totalamount: "107.52" });
    expect(erpDocTotals([{ sumamount: "0.5" }], 1)).toMatchObject({ totalvatvalue: "0.04" });
    // บรรทัดไม่ปัด: 3 × 0.333 = 0.999 รวมก่อนภาษีตามจริง
    const lines = [{ sumamount: erpLineAmount({ qty: "3", price: "0.333" }) }, { sumamount: "1000" }];
    expect(erpDocTotals(lines, 1, "7")).toEqual({ totalbeforevat: "1000.999", totalvatvalue: "70.07", totalamount: "1071.069" });
  });

  // ตัวอย่างระหว่างแก้แสดงยอดที่ระดับเงิน 2 ตำแหน่ง (ยอดบรรทัดที่บันทึกถูก PostgreSQL ปัดแบบเดียวกัน) — ค่าที่ส่งบันทึกไม่เปลี่ยน
  it("shows preview amounts at money scale 2, rounding the .005 boundary away from zero", () => {
    expect(erpMoneyDisplay(erpLineAmount({ qty: "0.12345678", price: "1234567.89" }))).toBe("152415.78");
    expect(erpMoneyDisplay("1.005")).toBe("1.01");
    expect(erpMoneyDisplay("1.0049")).toBe("1");
    expect(erpMoneyDisplay("-1.005")).toBe("-1.01");
    expect(erpMoneyDisplay("1071.069")).toBe("1071.07");
    expect(erpMoneyDisplay("")).toBe("0");
    expect(erpMoneyDisplay(undefined)).toBe("0");
  });

  it("treats blank or partial input as zero, clamps a discount larger than the line, and accepts exponent input", () => {
    expect(erpLineAmount({ qty: "", price: "100" })).toBe("0");
    expect(erpLineAmount({ qty: "-", price: "." })).toBe("0");
    expect(erpLineAmount({ qty: "1", price: "10", discountamount: "15" })).toBe("0");
    expect(erpLineAmount({ qty: "2", price: "10", discountamount: "0.5" })).toBe("19.5");
    expect(erpLineAmount({ qty: "1e2", price: "1.5" })).toBe("150");
    expect(erpLineAmount({ qty: "1e999", price: "1" })).toBe("0");
  });

  it("derives the before-VAT amount exactly when the backend omitted a zero field", () => {
    expect(erpBeforeVatAmount({ totalamount: "1070", totalvatvalue: "70" })).toBe("1000");
    expect(erpBeforeVatAmount({ totalamount: "0.3", totalvatvalue: "0.1" })).toBe("0.2");
    expect(erpBeforeVatAmount({ totalamount: "500" })).toBe("500");
    expect(erpBeforeVatAmount({ totalamount: "1070", totalbeforevat: "1000.5", totalvatvalue: "70" })).toBe("1000.5");
  });

  it("sends amounts as decimal strings and a cleared field as 0", async () => {
    const bodies: string[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (_url: string, init?: { body?: string }) => {
        bodies.push(String(init?.body));
        return { ok: true, status: 200, json: async () => ({ success: true, data: {} }) } as unknown as Response;
      }),
    );
    const config = getErpModuleConfig("/transaction/quotation")!;
    const doc: Partial<ErpTransactionDoc> = {
      docno: "",
      totalbeforevat: "0.3",
      totalvatvalue: "0.02",
      totalamount: "0.32",
      details: [{ linenumber: 1, itemcode: "P-1", unitcode: "PCS", qty: "", price: "0.1", sumamount: "0" }],
    };
    expect(erpDocForSave(doc).details?.[0].qty).toBe("0");
    await saveErpTransaction(config, doc, false);
    const sent = JSON.parse(bodies[0]);
    expect(sent).toMatchObject({ totalbeforevat: "0.3", totalvatvalue: "0.02", totalamount: "0.32" });
    expect(sent.details[0]).toMatchObject({ qty: "0", price: "0.1", sumamount: "0" });
    vi.unstubAllGlobals();
  });

  it("computes the due date as document date plus credit days like the backend", () => {
    expect(erpDueDate("2026-10-08", 30)).toBe("2026-11-07");
    expect(erpDueDate("2026-12-20", 15)).toBe("2027-01-04");
    expect(erpDueDate("2028-02-28", 1)).toBe("2028-02-29");
    expect(erpDueDate("2026-10-08", 0)).toBe("2026-10-08");
    expect(erpDueDate("", 30)).toBe("");
    expect(erpDueDate("2026-10-08", -1)).toBe("");
  });

  it("leaves a blank due date out of the request so the backend computes it", () => {
    const sent = erpDocForSave({ docno: "PUR6910-0001", creditday: 30, duedate: "  ", details: [] });
    expect(sent.duedate).toBeUndefined();
    expect(sent.creditday).toBe(30);
    expect(erpDocForSave({ docno: "PUR6910-0001", duedate: "2026-11-07" }).duedate).toBe("2026-11-07");
  });

  it("shows credit days and due date on exactly the backend due-date document types", () => {
    const source = readFileSync(resolve(process.cwd(), "../backend/internal/transaction/httpapi/due_date.go"), "utf8");
    const map = /var dueDateDocs = map\[string\]bool\{([^}]*)\}/.exec(source)?.[1] ?? "";
    const keys = [...map.matchAll(/"([a-z-]+)": true/g)].map((match) => match[1]);
    expect(keys.length).toBeGreaterThan(0);
    expect([...ERP_DUE_DATE_MODULES].sort()).toEqual(keys.sort());
    for (const key of ERP_DUE_DATE_MODULES) expect(ERP_TRANSACTION_MODULES).toContain(key);
  });
});

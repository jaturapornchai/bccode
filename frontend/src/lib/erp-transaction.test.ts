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
  type ErpTransactionDoc,
} from "./erp-transaction";
import { ERP_AUTO_POSTING_MODULES, ERP_TRANSACTION_MODULES } from "./erp-transaction-modules";
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
    // count-sheet and stockcount are the same Champ document (stock-count); no other screens share a type
    const shared = [...new Set(paths.filter((path, i) => paths.indexOf(path) !== i))];
    expect(shared).toEqual(["stock-count"]);
    expect(getErpModuleConfig("/inventory/count-sheet")?.apiPath).toBe("stock-count");
    expect(getErpModuleConfig("/transaction/stockcount")?.apiPath).toBe("stock-count");
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
    const posting = ["expense", "paid", "pay", "purchase", "purchase-return", "sale-invoice", "sale-invoice-return"];
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
    // stock-count is the only type reachable from two menu entries (doctype.screenAliases)
    const aliases: Record<string, string[]> = { "stock-count": ["stock-count-sheet"] };
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
});

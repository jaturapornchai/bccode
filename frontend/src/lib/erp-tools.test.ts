import { describe, it, expect, vi, afterEach } from "vitest";
import {
  ERP_TOOL_CONFIGS,
  isErpToolsRoute,
  getErpToolConfig,
  isErpToolApiReady,
  runErpTool,
} from "./erp-tools";

describe("ERP Tools & Integrity Recalculate Engine", () => {
  it("registers all 12 system recalculate and maintenance routes", () => {
    expect(ERP_TOOL_CONFIGS.length).toBe(12);

    expect(isErpToolsRoute("/gl/reprocess")).toBe(true);
    expect(isErpToolsRoute("/tools/ar-recalculate")).toBe(true);
    expect(isErpToolsRoute("/tools/ar-bill-balances")).toBe(true);
    expect(isErpToolsRoute("/tools/ap-recalculate")).toBe(true);
    expect(isErpToolsRoute("/tools/ap-bill-balances")).toBe(true);
    expect(isErpToolsRoute("/tools/cheque-balances")).toBe(true);
    expect(isErpToolsRoute("/tools/bank-balances")).toBe(true);
    expect(isErpToolsRoute("/rebuildstockscreen")).toBe(true);
    expect(isErpToolsRoute("/auditscreen")).toBe(true);
    expect(isErpToolsRoute("/rebuildproductsscreen")).toBe(true);
    expect(isErpToolsRoute("/rebuildproductbalancescreen")).toBe(true);
    expect(isErpToolsRoute("/inventory/daily-sequence")).toBe(true);
  });

  it("each tool has well-defined execution steps and action label", () => {
    for (const tool of ERP_TOOL_CONFIGS) {
      expect(tool.steps.length).toBeGreaterThan(0);
      expect(tool.actionLabel.th).toBeDefined();
      expect(tool.title.th).toBeDefined();
    }

    const glTool = getErpToolConfig("/gl/reprocess");
    expect(glTool?.domain).toBe("gl");
    expect(glTool?.steps.length).toBe(4);
  });
});

describe("runErpTool", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const scope = { holdingcode: "H01", businesscode: "C01", year: 2026 };

  it("marks only tools backed by a real endpoint as ready", () => {
    expect(isErpToolApiReady("rebuild_product_balance")).toBe(true);
    expect(isErpToolApiReady("audit_data")).toBe(true);
    expect(isErpToolApiReady("gl_reprocess")).toBe(false);
    expect(isErpToolApiReady("ar_recalculate")).toBe(false);
  });

  it("never calls the backend for tools without an endpoint", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "gl_reprocess", ...scope });

    expect(result).toEqual({ success: false, messageKey: "tool_not_available" });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("requires a selected business before calling the backend", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "audit_data", holdingcode: "", businesscode: "", year: 2026 });

    expect(result.messageKey).toBe("holding_required");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("reports success only when the backend confirms it", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ success: true, message: "Product balance updated successfully" }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "rebuild_product_balance", ...scope });

    expect(result.success).toBe(true);
    expect(result.messageKey).toBe("process_success");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toContain("/api/goapi/api/process/product-balance");
    expect(init.method).toBe("POST");
  });

  it("treats success:false from the backend as a failure", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({ ok: true, status: 200, json: async () => ({ success: false, message: "boom" }) })),
    );

    const result = await runErpTool({ code: "rebuild_product_balance", ...scope });

    expect(result.success).toBe(false);
    expect(result.messageKey).toBe("process_failed");
  });

  it("returns audit counters straight from the backend", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        status: "success",
        total_rows: 1250,
        total_documents: 320,
        total_products: 87,
        earliest_date: "2026-01-04T03:00:00Z",
        latest_date: "2026-09-15T10:00:00Z",
      }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "audit_data", ...scope });

    expect(result.success).toBe(true);
    expect(result.stats?.totalrows).toBe(1250);
    expect(result.stats?.totaldocuments).toBe(320);
    expect(result.stats?.totalproducts).toBe(87);
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(String(init.body)).toContain("2026-01-01");
  });

  it("maps auth, server and network failures to error keys", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 401, json: async () => ({}) })));
    expect((await runErpTool({ code: "audit_data", ...scope })).messageKey).toBe("unauthorized");

    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 500, json: async () => ({}) })));
    expect((await runErpTool({ code: "audit_data", ...scope })).messageKey).toBe("process_failed");

    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("offline"); }));
    expect((await runErpTool({ code: "audit_data", ...scope })).messageKey).toBe("connection_error");
  });
});

import { describe, it, expect, vi, afterEach } from "vitest";
import {
  ERP_TOOL_CONFIGS,
  isErpToolsRoute,
  getErpToolConfig,
  isErpToolApiReady,
  runErpTool,
} from "./erp-tools";
import { setupTestAuthSession } from "./test-auth-session";

setupTestAuthSession();

describe("ERP Tools & Integrity Recalculate Engine", () => {
  it("registers the 8 Champ recalculate and maintenance routes", () => {
    expect(ERP_TOOL_CONFIGS.length).toBe(8);

    expect(isErpToolsRoute("/tools/ar-recalculate")).toBe(true);
    expect(isErpToolsRoute("/tools/ar-bill-balances")).toBe(true);
    expect(isErpToolsRoute("/tools/ap-recalculate")).toBe(true);
    expect(isErpToolsRoute("/tools/ap-bill-balances")).toBe(true);
    expect(isErpToolsRoute("/tools/cheque-balances")).toBe(true);
    expect(isErpToolsRoute("/tools/bank-balances")).toBe(true);
    expect(isErpToolsRoute("/rebuildstockscreen")).toBe(true);
    expect(isErpToolsRoute("/inventory/daily-sequence")).toBe(true);
  });

  it("each tool has well-defined execution steps and action label", () => {
    for (const tool of ERP_TOOL_CONFIGS) {
      expect(tool.steps.length).toBeGreaterThan(0);
      expect(tool.actionLabel.th).toBeDefined();
      expect(tool.title.th).toBeDefined();
    }

    const arTool = getErpToolConfig("/tools/ar-recalculate");
    expect(arTool?.domain).toBe("ar");
    expect(arTool?.steps.length).toBe(3);
    // /gl/reprocess is a GL menu item now (general-ledger-screen → GLProcesses), not a tools route
    expect(isErpToolsRoute("/gl/reprocess")).toBe(false);
  });
});

describe("runErpTool", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const scope = { holdingcode: "H01", businesscode: "C01", year: 2026 };

  it("marks AR, AP, cheque, and bank tools as ready and others as not ready", () => {
    expect(isErpToolApiReady("ar_recalculate")).toBe(true);
    expect(isErpToolApiReady("ar_bill_balances")).toBe(true);
    expect(isErpToolApiReady("ap_recalculate")).toBe(true);
    expect(isErpToolApiReady("ap_bill_balances")).toBe(true);
    expect(isErpToolApiReady("cheque_balances")).toBe(true);
    expect(isErpToolApiReady("bank_balances")).toBe(true);

    const pendingCodes = ["rebuild_stock", "inventory_daily_sequence"];
    for (const code of pendingCodes) {
      expect(isErpToolApiReady(code), code).toBe(false);
    }
  });

  it("never calls the backend for tools without an endpoint", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "rebuild_stock", ...scope });

    expect(result).toEqual({ success: false, messageKey: "tool_not_available" });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("calls /api/gl/command for ar_recalculate", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "ar_recalculate", ...scope });

    expect(result).toEqual({ success: true, messageKey: "process_success" });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/gl/command",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          resource: "processes",
          action: "recalculate",
          reason: "คำนวณยอดลูกหนี้ใหม่ (Champ 5110)",
        }),
      }),
    );
  });

  it("calls /api/gl/command for ap_recalculate", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "ap_recalculate", ...scope });

    expect(result).toEqual({ success: true, messageKey: "process_success" });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/gl/command",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          resource: "processes",
          action: "recalculate",
          reason: "คำนวณยอดเจ้าหนี้ใหม่ (Champ 5035)",
        }),
      }),
    );
  });

  it("calls /api/cheque/command for cheque_balances recalculation", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: { checked: 10, total: 0, mismatches: [] } }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "cheque_balances", ...scope });

    expect(result).toEqual({ success: true, messageKey: "process_success" });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/cheque/command",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          resource: "recalculate",
          action: "apply",
          reason: "คำนวณยอดคงเหลือของเช็คใหม่ตามระบบอัตโนมัติ (Champ 5400)",
        }),
      }),
    );
  });

  it("calls /api/cheque/command for bank_balances recalculation", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: { total_books: 2, books: [] } }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "bank_balances", ...scope });

    expect(result).toEqual({ success: true, messageKey: "process_success" });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/cheque/command",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          resource: "bank-recalculate",
          action: "apply",
        }),
      }),
    );
  });
});

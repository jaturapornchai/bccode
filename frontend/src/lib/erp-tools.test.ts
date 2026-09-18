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

  it("marks no tool as ready until the backend exposes an endpoint", () => {
    for (const tool of ERP_TOOL_CONFIGS) expect(isErpToolApiReady(tool.code), tool.code).toBe(false);
  });

  it("never calls the backend for tools without an endpoint", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await runErpTool({ code: "ar_recalculate", ...scope });

    expect(result).toEqual({ success: false, messageKey: "tool_not_available" });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

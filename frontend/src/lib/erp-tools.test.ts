import { describe, it, expect } from "vitest";
import {
  ERP_TOOL_CONFIGS,
  isErpToolsRoute,
  getErpToolConfig,
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

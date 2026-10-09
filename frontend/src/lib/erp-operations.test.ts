import { describe, it, expect, vi, afterEach } from "vitest";
import {
  OPERATIONS_CONFIGS,
  isOperationsRoute,
  getOperationsConfig,
  isApprovalApiReady,
  fetchPendingApprovals,
  submitApprovalAction,
} from "./erp-operations";
import { setupTestAuthSession } from "./test-auth-session";

setupTestAuthSession();

describe("ERP Operations, Approvals, BOM & Import Workflows", () => {
  it("registers all 22 operations workflow routes across 6 categories", () => {
    expect(OPERATIONS_CONFIGS.length).toBe(22);

    // Approvals
    expect(isOperationsRoute("/procurement/requisition-approval")).toBe(true);
    expect(isOperationsRoute("/procurement/order-cancellation")).toBe(true);
    expect(isOperationsRoute("/sales/quotation-approval")).toBe(true);
    expect(isOperationsRoute("/sales/quotation-cancellation")).toBe(true);
    expect(isOperationsRoute("/sales/order-approval")).toBe(true);
    expect(isOperationsRoute("/sales/order-cancellation")).toBe(true);
    // 2026-10-09 (S1): cancel/approve screens of the real Champ 5.5 dropdowns
    expect(isOperationsRoute("/procurement/requisition-cancellation")).toBe(true);
    expect(isOperationsRoute("/procurement/requisition-approval-cancellation")).toBe(true);
    expect(isOperationsRoute("/procurement/order-approval")).toBe(true);
    expect(isOperationsRoute("/inventory/issue-request-cancellation")).toBe(true);
    expect(isOperationsRoute("/inventory/transfer-request-cancellation")).toBe(true);

    // Reservations
    expect(isOperationsRoute("/sales/reservations")).toBe(true);
    expect(isOperationsRoute("/sales/order-dates")).toBe(true);
    expect(isOperationsRoute("/sales/delivery-dates")).toBe(true);

    // Procurement
    expect(isOperationsRoute("/procurement/price-comparison")).toBe(true);
    expect(isOperationsRoute("/procurement/generate-orders")).toBe(true);

    // BOM & Serial
    expect(isOperationsRoute("/inventory/set-assembly")).toBe(true);
    expect(isOperationsRoute("/inventory/set-disassembly")).toBe(true);
    expect(isOperationsRoute("/inventory/set-components")).toBe(true);
    expect(isOperationsRoute("/productserialregistry")).toBe(true);

    // Pricing
    expect(isOperationsRoute("/inventory/selling-prices")).toBe(true);
    expect(isOperationsRoute("/inventory/price-adjustment")).toBe(true);

    // Import
  });

  it("provides operational action metadata for each route", () => {
    const prApproval = getOperationsConfig("/procurement/requisition-approval");
    expect(prApproval?.category).toBe("approval");
    expect(prApproval?.documentType).toBe("PR");

    expect(getOperationsConfig("/procurement/order-approval")?.documentType).toBe("PO");
    expect(getOperationsConfig("/procurement/requisition-cancellation")?.documentType).toBe("PR");
    expect(getOperationsConfig("/procurement/requisition-approval-cancellation")?.documentType).toBe("PR");
    expect(getOperationsConfig("/inventory/issue-request-cancellation")?.documentType).toBe("SIR");
    expect(getOperationsConfig("/inventory/transfer-request-cancellation")?.documentType).toBe("STR");

    const serial = getOperationsConfig("/productserialregistry");
    expect(serial?.category).toBe("bom");
  });
});

describe("approval workflow API", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("only exposes screens that have a real approval endpoint", () => {
    // backend ไม่มี endpoint อนุมัติ (pr-status/pending ไม่มีจริง) จึงไม่มีจอใดพร้อม
    expect(isApprovalApiReady("pr_approval")).toBe(false);
    expect(isApprovalApiReady("po_cancellation")).toBe(false);
    expect(isApprovalApiReady("qt_approval")).toBe(false);
    expect(isApprovalApiReady("so_approval")).toBe(false);
    for (const code of ["pr_cancellation", "pr_approval_cancellation", "po_approval", "stock_issue_request_cancellation", "stock_transfer_request_cancellation"]) {
      expect(isApprovalApiReady(code), code).toBe(false);
    }
  });

  it("never calls the backend for screens without an endpoint", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const list = await fetchPendingApprovals({ code: "qt_approval", holdingcode: "H01" });
    const action = await submitApprovalAction({
      code: "qt_approval",
      holdingcode: "H01",
      docno: "QT-1",
      action: "approve",
      actionby: "user1",
    });

    expect(list.error).toBe("approval_not_available");
    expect(action.messageKey).toBe("approval_not_available");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  // pr_approval ถูกถอดจาก APPROVAL_KINDS: backend ไม่มี endpoint อนุมัติ จึงห้ามยิง API ที่ไม่มีอยู่
  it("does not call the backend for pr_approval while no approval endpoint exists", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const list = await fetchPendingApprovals({ code: "pr_approval", holdingcode: "H01" });
    const approved = await submitApprovalAction({
      code: "pr_approval", holdingcode: "H01", docno: "PR-1", action: "approve", actionby: "user1",
    });

    expect(list).toEqual({ docs: [], error: "approval_not_available" });
    expect(approved).toEqual({ success: false, messageKey: "approval_not_available" });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

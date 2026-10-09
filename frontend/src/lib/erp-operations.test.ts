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
  it("registers all 23 operations workflow routes across 6 categories", () => {
    expect(OPERATIONS_CONFIGS.length).toBe(23);

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
    expect(isOperationsRoute("/procurement/delivery-dates")).toBe(true);
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

  it("exposes all 11 Tier A approval and cancellation screens with real endpoints", () => {
    const activeApprovalCodes = [
      "pr_approval",
      "po_approval",
      "qt_approval",
      "so_approval",
      "pr_approval_cancellation",
      "pr_cancellation",
      "po_cancellation",
      "qt_cancellation",
      "so_cancellation",
      "stock_issue_request_cancellation",
      "stock_transfer_request_cancellation",
    ];

    for (const code of activeApprovalCodes) {
      expect(isApprovalApiReady(code), code).toBe(true);
    }

    // Screens without an approval endpoint remain inactive
    for (const code of ["sales_reservations", "order_dates", "delivery_dates", "set_assembly", "serial_registry", "selling_prices"]) {
      expect(isApprovalApiReady(code), code).toBe(false);
    }
  });

  it("never calls the backend for screens without an endpoint", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const list = await fetchPendingApprovals({ code: "sales_reservations", holdingcode: "H01" });
    const action = await submitApprovalAction({
      code: "sales_reservations",
      holdingcode: "H01",
      docno: "SO-1",
      action: "approve",
      actionby: "user1",
    });

    expect(list.error).toBe("approval_not_available");
    expect(action.messageKey).toBe("approval_not_available");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("calls the real backend endpoint for active approval screens", async () => {
    const fetchMock = vi.fn().mockImplementation((url: string, _init?: RequestInit) => {
      const u = String(url);
      if (u.includes("/list")) {
        return Promise.resolve({
          ok: true,
          status: 200,
          json: () =>
            Promise.resolve({
              success: true,
              data: [
                {
                  docno: "PR-202610-001",
                  docdatetime: "2026-10-10",
                  custname: "ผู้จำหน่าย ก",
                  totalamount: 1500,
                  createdby: "staff",
                  approval_status: 0,
                  status: 0,
                  iscancel: false,
                },
              ],
              pagination: { total: 1 },
            }),
        });
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true }),
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    // List pending
    const list = await fetchPendingApprovals({ code: "pr_approval", holdingcode: "H01" });
    expect(list.docs).toHaveLength(1);
    expect(list.docs[0].docno).toBe("PR-202610-001");
    expect(list.docs[0].totalamount).toBe(1500);

    // Approve
    const approved = await submitApprovalAction({
      code: "pr_approval",
      docno: "PR-202610-001",
      action: "approve",
      actionby: "manager",
    });
    expect(approved.success).toBe(true);
    expect(approved.messageKey).toBe("approve_success");

    // Cancel
    const cancelled = await submitApprovalAction({
      code: "pr_cancellation",
      docno: "PR-202610-001",
      action: "cancel",
      comment: "ยกเลิกเพราะขอซ้ำ",
    });
    expect(cancelled.success).toBe(true);
    expect(cancelled.messageKey).toBe("cancel_success");

    // Unapprove
    const unapproved = await submitApprovalAction({
      code: "pr_approval_cancellation",
      docno: "PR-202610-001",
      action: "unapprove",
      comment: "ขอทบทวนยอดใหม่",
    });
    expect(unapproved.success).toBe(true);
    expect(unapproved.messageKey).toBe("unapprove_success");
  });
});

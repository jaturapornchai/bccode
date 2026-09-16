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
  it("registers all 27 operations workflow routes across 6 categories", () => {
    expect(OPERATIONS_CONFIGS.length).toBe(27);

    // Approvals
    expect(isOperationsRoute("/procurement/requisition-approval")).toBe(true);
    expect(isOperationsRoute("/procurement/order-cancellation")).toBe(true);
    expect(isOperationsRoute("/sales/quotation-approval")).toBe(true);
    expect(isOperationsRoute("/sales/quotation-cancellation")).toBe(true);
    expect(isOperationsRoute("/sales/order-approval")).toBe(true);
    expect(isOperationsRoute("/sales/order-cancellation")).toBe(true);

    // Reservations
    expect(isOperationsRoute("/sales/reservations")).toBe(true);
    expect(isOperationsRoute("/sales/order-dates")).toBe(true);
    expect(isOperationsRoute("/sales/delivery-dates")).toBe(true);

    // Procurement
    expect(isOperationsRoute("/procurement/dashboard")).toBe(true);
    expect(isOperationsRoute("/procurement/price-comparison")).toBe(true);
    expect(isOperationsRoute("/procurement/generate-orders")).toBe(true);
    expect(isOperationsRoute("/transaction/documentvault")).toBe(true);
    expect(isOperationsRoute("/transaction/documentinbox")).toBe(true);

    // BOM & Serial
    expect(isOperationsRoute("/inventory/set-assembly")).toBe(true);
    expect(isOperationsRoute("/inventory/set-disassembly")).toBe(true);
    expect(isOperationsRoute("/inventory/set-components")).toBe(true);
    expect(isOperationsRoute("/productserialregistry")).toBe(true);

    // Pricing
    expect(isOperationsRoute("/inventory/selling-prices")).toBe(true);
    expect(isOperationsRoute("/inventory/price-adjustment")).toBe(true);
    expect(isOperationsRoute("/inventory/cost-layers")).toBe(true);
    expect(isOperationsRoute("/inventory/lotexpiry")).toBe(true);

    // Import
    expect(isOperationsRoute("/importdocuments")).toBe(true);
    expect(isOperationsRoute("/importpartner")).toBe(true);
    expect(isOperationsRoute("/importproduct")).toBe(true);
    expect(isOperationsRoute("/importproductfromfile")).toBe(true);
    expect(isOperationsRoute("/importproductimage")).toBe(true);
  });

  it("provides operational action metadata for each route", () => {
    const prApproval = getOperationsConfig("/procurement/requisition-approval");
    expect(prApproval?.category).toBe("approval");
    expect(prApproval?.documentType).toBe("PR");

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
    expect(isApprovalApiReady("pr_approval")).toBe(true);
    expect(isApprovalApiReady("po_cancellation")).toBe(false);
    expect(isApprovalApiReady("qt_approval")).toBe(false);
    expect(isApprovalApiReady("so_approval")).toBe(false);
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

  it("maps pending documents from the backend payload", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        success: true,
        count: 1,
        data: [
          {
            docno: "PR-202609-001",
            docdatetime: "2026-09-12",
            createdbyname: "user one",
            custname: "Vendor Co.",
            totalamount: 128400,
            requiredlevelname: "Manager",
          },
        ],
      }),
    }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchPendingApprovals({ code: "pr_approval", holdingcode: "H01" });

    expect(result.error).toBeUndefined();
    expect(result.docs).toHaveLength(1);
    expect(result.docs[0].docno).toBe("PR-202609-001");
    expect(result.docs[0].totalamount).toBe(128400);
    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toContain("/api/goapi/api/approval/pr-status/pending");
  });

  it("returns an empty list with an error key when the backend fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 500, json: async () => ({}) })));
    const result = await fetchPendingApprovals({ code: "pr_approval", holdingcode: "H01" });
    expect(result.docs).toEqual([]);
    expect(result.error).toBe("load_failed");
  });

  it("requires holding, docno and an acting user before approving", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const base = { code: "pr_approval", action: "approve" as const };

    expect((await submitApprovalAction({ ...base, holdingcode: "", docno: "PR-1", actionby: "u" })).messageKey).toBe("holding_required");
    expect((await submitApprovalAction({ ...base, holdingcode: "H01", docno: "", actionby: "u" })).messageKey).toBe("docno_required");
    expect((await submitApprovalAction({ ...base, holdingcode: "H01", docno: "PR-1", actionby: "" })).messageKey).toBe("unauthorized");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("posts approve and reject to the matching endpoint", async () => {
    const fetchMock = vi.fn(async () => ({ ok: true, status: 200, json: async () => ({ success: true }) }));
    vi.stubGlobal("fetch", fetchMock);

    const approved = await submitApprovalAction({
      code: "pr_approval", holdingcode: "H01", docno: "PR-1", action: "approve", actionby: "user1",
    });
    const rejected = await submitApprovalAction({
      code: "pr_approval", holdingcode: "H01", docno: "PR-1", action: "reject", actionby: "user1",
    });

    expect(approved).toEqual({ success: true, messageKey: "approve_success" });
    expect(rejected).toEqual({ success: true, messageKey: "reject_success" });
    expect((fetchMock.mock.calls[0] as unknown as [string])[0]).toContain("/pr-status/approve");
    expect((fetchMock.mock.calls[1] as unknown as [string])[0]).toContain("/pr-status/reject");
    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(String(init.body)).toContain("\"actionby\":\"user1\"");
  });

  it("treats success:false and network failures as failed actions", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: true, status: 200, json: async () => ({ success: false }) })));
    expect((await submitApprovalAction({ code: "pr_approval", holdingcode: "H01", docno: "PR-1", action: "approve", actionby: "u" })).messageKey).toBe("action_failed");

    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("offline"); }));
    expect((await submitApprovalAction({ code: "pr_approval", holdingcode: "H01", docno: "PR-1", action: "approve", actionby: "u" })).messageKey).toBe("connection_error");
  });
});

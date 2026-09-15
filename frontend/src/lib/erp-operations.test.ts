import { describe, it, expect } from "vitest";
import {
  OPERATIONS_CONFIGS,
  isOperationsRoute,
  getOperationsConfig,
} from "./erp-operations";

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

import { describe, expect, it } from "vitest";
import {
  ERP_MODULE_CONFIGS,
  isErpTransactionRoute,
  getErpModuleConfig,
  fetchErpTransactions,
  saveErpTransaction,
  deleteErpTransaction,
  type ErpTransactionDoc,
} from "./erp-transaction";

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
    expect(isErpTransactionRoute("/transaction/taxinvoice")).toBe(true);
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
    expect(isErpTransactionRoute("/transaction/paymentvoucher")).toBe(true);
    expect(isErpTransactionRoute("/transaction/pay")).toBe(true);

    // Cash & Bank
    expect(isErpTransactionRoute("/transaction/accounttransfer")).toBe(true);
    expect(isErpTransactionRoute("/transaction/chequereceived")).toBe(true);
    expect(isErpTransactionRoute("/transaction/chequeissued")).toBe(true);
    expect(isErpTransactionRoute("/transaction/directoradvance")).toBe(true);

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
    expect(saleConfig?.defaultDocPrefix).toBe("SO");

    const poConfig = getErpModuleConfig("/transaction/purchaseorder");
    expect(poConfig).toBeDefined();
    expect(poConfig?.domain).toBe("purchase");
    expect(poConfig?.title.th).toContain("ใบสั่งซื้อ");
    expect(poConfig?.counterpartyLabel.th).toBe("ผู้จำหน่าย");
    expect(poConfig?.defaultDocPrefix).toBe("PO");
  });

  it("executes full CRUD cycle on transactions", async () => {
    const config = getErpModuleConfig("/transaction/quotation")!;
    expect(config).toBeDefined();

    // 1. Initial list
    const initial = await fetchErpTransactions(config);
    expect(initial.items.length).toBeGreaterThan(0);
    const initialCount = initial.items.length;

    // 2. Create
    const newDoc: Partial<ErpTransactionDoc> = {
      docno: "QT-TEST-9999",
      docdatetime: "2026-09-15T10:00:00Z",
      custcode: "CUST-TEST",
      custname: "บริษัท ทดสอบความถูกต้อง จำกัด",
      totalamount: 10700,
      totalbeforevat: 10000,
      totalvatvalue: 700,
      status: 0,
      details: [
        {
          linenumber: 1,
          itemcode: "P-TEST",
          itemname: "สินค้าทดสอบ",
          unitcode: "PCS",
          qty: 5,
          price: 2000,
          sumamount: 10000,
        },
      ],
    };
    const created = await saveErpTransaction(config, newDoc, false);
    expect(created.success).toBe(true);
    expect(created.data?.docno).toBe("QT-TEST-9999");
    const createdId = created.data?.id!;

    // Verify presence in list
    const afterCreate = await fetchErpTransactions(config);
    expect(afterCreate.items.length).toBe(initialCount + 1);
    expect(afterCreate.items.some((x) => x.id === createdId)).toBe(true);

    // 3. Update
    const updated = await saveErpTransaction(
      config,
      { id: createdId, docno: "QT-TEST-9999", totalamount: 12000, status: 1 },
      true,
    );
    expect(updated.success).toBe(true);
    expect(updated.data?.totalamount).toBe(12000);
    expect(updated.data?.status).toBe(1);

    // 4. Delete
    const deleted = await deleteErpTransaction(config, createdId);
    expect(deleted.success).toBe(true);

    const afterDelete = await fetchErpTransactions(config);
    expect(afterDelete.items.some((x) => x.id === createdId)).toBe(false);
  });
});

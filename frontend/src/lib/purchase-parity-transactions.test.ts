import { describe, expect, it } from "vitest";
import { getErpModuleConfig, isErpTransactionRoute } from "./erp-transaction";
import { isMenuScreenPending } from "./menu-screen-status";

describe("Purchase Parity Transactions (Champ 5801, 5809)", () => {
  it("recognizes /transaction/purchasereducedebt as a connected ERP transaction", () => {
    expect(isErpTransactionRoute("/transaction/purchasereducedebt")).toBe(true);
    expect(isMenuScreenPending("/transaction/purchasereducedebt")).toBe(false);

    const config = getErpModuleConfig("/transaction/purchasereducedebt");
    expect(config).toBeDefined();
    expect(config?.moduleKey).toBe("purchase_reduce_debt");
    expect(config?.apiPath).toBe("apcreditnote");
    expect(config?.domain).toBe("purchase");
    expect(config?.title.th).toBe("บันทึกลดหนี้");
    expect(config?.title.en).toBe("Purchase Debt Reduction");
    expect(config?.hasLineItems).toBe(true);
  });

  it("recognizes /transaction/rfqpricetable as a connected ERP transaction", () => {
    expect(isErpTransactionRoute("/transaction/rfqpricetable")).toBe(true);
    expect(isMenuScreenPending("/transaction/rfqpricetable")).toBe(false);

    const config = getErpModuleConfig("/transaction/rfqpricetable");
    expect(config).toBeDefined();
    expect(config?.moduleKey).toBe("rfq_price_table");
    expect(config?.apiPath).toBe("rfq");
    expect(config?.domain).toBe("purchase");
    expect(config?.title.th).toBe("บันทึกใบตารางราคารวม");
    expect(config?.title.en).toBe("Consolidated Price Table");
    expect(config?.hasLineItems).toBe(true);
  });
});

import { describe, it, expect } from "vitest";
import { calculateStockCard, type RawStockMovement } from "./inventory-stock-card";

describe("calculateStockCard", () => {
  it("computes moving weighted average and running balance correctly", () => {
    const movements: RawStockMovement[] = [
      {
        id: "1",
        date: "2026-09-01",
        docno: "PO2026-001",
        taxno: "INV-001",
        docType: "purchase",
        description: "ซื้อล็อต 1 จากซัพพลายเออร์ A",
        itemcode: "ITEM-A",
        qty: 100,
        unitCost: 50,
        totalAmount: 5000,
      },
      {
        id: "2",
        date: "2026-09-05",
        docno: "SO2026-001",
        taxno: "TAX-001",
        docType: "sale",
        description: "ขายให้ลูกค้า ก",
        itemcode: "ITEM-A",
        qty: 40,
      },
      {
        id: "3",
        date: "2026-09-10",
        docno: "PO2026-002",
        taxno: "INV-002",
        docType: "purchase",
        description: "ซื้อล็อต 2 ต้นทุนสูงขึ้น",
        itemcode: "ITEM-A",
        qty: 60,
        unitCost: 70,
        totalAmount: 4200,
      },
      {
        id: "4",
        date: "2026-09-15",
        docno: "SO2026-002",
        taxno: "TAX-002",
        docType: "sale",
        description: "ขายให้ลูกค้า ข",
        itemcode: "ITEM-A",
        qty: 50,
      },
    ];

    const result = calculateStockCard("ITEM-A", movements, {
      itemname: "สินค้าตัวอย่าง A",
      beginningQty: 0,
      beginningCost: 0,
    });

    expect(result.itemcode).toBe("ITEM-A");
    expect(result.totalInQty).toBe(160);
    expect(result.totalInTotal).toBe(9200);
    expect(result.totalOutQty).toBe(90);

    // After Step 1 (Purchase 100 @ 50): Bal = 100, Cost = 50, Total = 5000
    // After Step 2 (Sale 40 @ 50): Bal = 60, Cost = 50, Total = 3000
    // After Step 3 (Purchase 60 @ 70, Total 4200):
    //    NewQty = 60 + 60 = 120
    //    NewTotal = 3000 + 4200 = 7200
    //    NewAvgCost = 7200 / 120 = 60
    // After Step 4 (Sale 50 @ 60):
    //    NewQty = 120 - 50 = 70
    //    NewTotal = 7200 - (50 * 60) = 4200
    //    EndingAvgCost = 60
    expect(result.endingQty).toBe(70);
    expect(result.endingCost).toBe(60);
    expect(result.endingTotal).toBe(4200);
  });

  it("handles opening beginning inventory correctly", () => {
    const movements: RawStockMovement[] = [
      {
        id: "1",
        date: "2026-09-02",
        docno: "SO-001",
        docType: "sale",
        itemcode: "ITEM-B",
        qty: 10,
      },
    ];

    const result = calculateStockCard("ITEM-B", movements, {
      itemname: "สินค้า B",
      beginningQty: 50,
      beginningCost: 100,
    });

    expect(result.beginningQty).toBe(50);
    expect(result.beginningTotal).toBe(5000);
    expect(result.lines[0].docType).toBe("beginning");
    expect(result.lines[0].balanceQty).toBe(50);
    expect(result.lines[1].outQty).toBe(10);
    expect(result.lines[1].balanceQty).toBe(40);
    expect(result.endingQty).toBe(40);
    expect(result.endingCost).toBe(100);
    expect(result.endingTotal).toBe(4000);
  });
});

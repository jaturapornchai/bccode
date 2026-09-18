import { describe, it, expect } from "vitest";
import {
  calculateMovingAverageCost,
  calculateFifoCost,
  allocateLandedCost,
  generateStockVarianceJournal,
  calculatePeriodicStockClosing,
  type StockMovementItem,
} from "./inventory-costing";

describe("Inventory Costing & Stock Closing Engine (Group D)", () => {
  it("calculates Moving Average Cost and COGS on each issue accurately", () => {
    // 1. Buy 100 units @ 10 THB = 1,000 THB
    // 2. Buy 50 units @ 16 THB = 800 THB => Total 150 units, 1,800 THB => Avg = 12 THB
    // 3. Sell 80 units => COGS = 80 * 12 = 960 THB => Rem 70 units @ 12 THB = 840 THB
    const movements: StockMovementItem[] = [
      { id: "1", date: "2026-09-01", docNo: "RC-01", type: "RECEIPT", itemCode: "ITEM-A", quantity: 100, unitCost: 10 },
      { id: "2", date: "2026-09-05", docNo: "RC-02", type: "RECEIPT", itemCode: "ITEM-A", quantity: 50, unitCost: 16 },
      { id: "3", date: "2026-09-10", docNo: "INV-01", type: "ISSUE", itemCode: "ITEM-A", quantity: 80 },
    ];

    const res = calculateMovingAverageCost(movements);
    expect(res.currentQty).toBe(70);
    expect(res.averageCost).toBe(12);
    expect(res.totalValuation).toBe(840);
    expect(res.issueCogs).toHaveLength(1);
    expect(res.issueCogs[0].cogs).toBe(960);
    expect(res.issueCogs[0].unitCost).toBe(12);
  });

  it("calculates FIFO Cost layers and tracks depletion correctly", () => {
    // 1. Buy 10 units @ 100 THB
    // 2. Buy 10 units @ 120 THB
    // 3. Sell 15 units: takes 10 @ 100 (=1,000) + 5 @ 120 (=600) => COGS = 1,600 THB
    // Remaining: 5 units @ 120 THB = 600 THB
    const movements: StockMovementItem[] = [
      { id: "1", date: "2026-09-01", docNo: "RC-01", type: "RECEIPT", itemCode: "ITEM-B", quantity: 10, unitCost: 100 },
      { id: "2", date: "2026-09-05", docNo: "RC-02", type: "RECEIPT", itemCode: "ITEM-B", quantity: 10, unitCost: 120 },
      { id: "3", date: "2026-09-10", docNo: "INV-01", type: "ISSUE", itemCode: "ITEM-B", quantity: 15 },
    ];

    const res = calculateFifoCost(movements);
    expect(res.currentQty).toBe(5);
    expect(res.totalValuation).toBe(600);
    expect(res.issueCogs[0].cogs).toBe(1600);
    expect(res.remainingLayers).toHaveLength(1);
    expect(res.remainingLayers[0].remainingQty).toBe(5);
    expect(res.remainingLayers[0].unitCost).toBe(120);
  });

  it("allocates landed cost by value accurately to incoming goods", () => {
    // 2 items:
    // Item 1: 10 units @ 1,000 = 10,000 THB (50%)
    // Item 2: 5 units @ 2,000 = 10,000 THB (50%)
    // Freight/Duty: 2,000 THB => Allocated 1,000 each
    // Item 1 new unit cost: 1,000 + 100 = 1,100 THB
    // Item 2 new unit cost: 2,000 + 200 = 2,200 THB
    const receipts = [
      { itemCode: "PROD-1", quantity: 10, unitPrice: 1000 },
      { itemCode: "PROD-2", quantity: 5, unitPrice: 2000 },
    ];

    const result = allocateLandedCost(receipts, 2000, "BY_VALUE");
    expect(result[0].allocatedLandedCost).toBe(1000);
    expect(result[0].finalUnitCost).toBe(1100);

    expect(result[1].allocatedLandedCost).toBe(1000);
    expect(result[1].finalUnitCost).toBe(2200);
  });

  it("generates balanced stock variance adjustment journal entry", () => {
    // Physical: 95, Book: 100 => Shortage 5 units @ 200 THB = 1,000 THB
    const res = generateStockVarianceJournal({
      itemCode: "SKU-99",
      itemName: "เครื่องวัดอุณหภูมิ",
      bookQty: 100,
      physicalQty: 95,
      unitCost: 200,
    });

    expect(res.varianceQty).toBe(-5);
    expect(res.varianceAmount).toBe(1000);
    expect(res.isShortage).toBe(true);
    expect(res.journalLines).toHaveLength(2);

    const sumDr = res.journalLines.reduce((s, l) => s + l.debit, 0);
    const sumCr = res.journalLines.reduce((s, l) => s + l.credit, 0);
    expect(sumDr).toBe(1000);
    expect(sumDr).toBe(sumCr);
  });

  it("calculates Periodic stock closing and balances closing entries", () => {
    // Beg: 50,000, Purchases: 200,000, Physical Ending: 60,000
    // COGS = 50,000 + 200,000 - 60,000 = 190,000 THB
    const closing = calculatePeriodicStockClosing({
      beginningInventory: 50000,
      netPurchases: 200000,
      endingInventoryPhysical: 60000,
    });

    expect(closing.cogs).toBe(190000);
    expect(closing.closingJournalLines).toHaveLength(4);

    const sumDr = closing.closingJournalLines.reduce((s, l) => s + l.debit, 0);
    const sumCr = closing.closingJournalLines.reduce((s, l) => s + l.credit, 0);
    expect(sumDr).toBe(250000); // 60,000 + 190,000 = 250,000
    expect(sumCr).toBe(250000); // 50,000 + 200,000 = 250,000
    expect(sumDr).toBe(sumCr);
  });
});

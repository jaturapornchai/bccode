/**
 * Inventory Costing & Stock Closing Engine
 * Standards: TFRS for NPAEs / PAEs (Cost formulas: Moving Average & FIFO, Landed Cost Allocation, Stock Count Adjustment, and Periodic Closing)
 */

export type StockMovementType = "RECEIPT" | "ISSUE" | "ADJUSTMENT";

export type StockMovementItem = {
  id: string;
  date: string; // YYYY-MM-DD
  docNo: string;
  type: StockMovementType;
  itemCode: string;
  quantity: number; // positive
  unitCost?: number; // on receipt
  totalCost?: number;
};

export type StockLayer = {
  receivedDate: string;
  docNo: string;
  initialQty: number;
  remainingQty: number;
  unitCost: number;
};

/**
 * Calculates Moving Average Cost through a sequence of chronological stock movements.
 */
export function calculateMovingAverageCost(movements: StockMovementItem[]): {
  currentQty: number;
  averageCost: number;
  totalValuation: number;
  issueCogs: { docNo: string; quantity: number; unitCost: number; cogs: number }[];
} {
  let currentQty = 0;
  let totalValuation = 0;
  let currentAvg = 0;
  const issueCogs: { docNo: string; quantity: number; unitCost: number; cogs: number }[] = [];

  for (const m of movements) {
    if (m.type === "RECEIPT") {
      const inCost = m.unitCost ?? (m.totalCost && m.quantity > 0 ? m.totalCost / m.quantity : 0);
      const incomingValuation = Math.round(m.quantity * inCost * 100) / 100;

      const nextQty = currentQty + m.quantity;
      const nextValuation = Math.round((totalValuation + incomingValuation) * 100) / 100;
      currentAvg = nextQty > 0 ? Math.round((nextValuation / nextQty) * 10000) / 10000 : 0;

      currentQty = nextQty;
      totalValuation = nextValuation;
    } else if (m.type === "ISSUE") {
      const issueQty = Math.min(currentQty, m.quantity);
      const cogs = Math.round(issueQty * currentAvg * 100) / 100;

      currentQty = Math.max(0, currentQty - issueQty);
      totalValuation = Math.round(Math.max(0, totalValuation - cogs) * 100) / 100;
      if (currentQty === 0) currentAvg = 0;

      issueCogs.push({
        docNo: m.docNo,
        quantity: issueQty,
        unitCost: currentAvg,
        cogs,
      });
    } else if (m.type === "ADJUSTMENT") {
      currentQty = Math.max(0, currentQty + m.quantity);
      totalValuation = Math.round(currentQty * currentAvg * 100) / 100;
    }
  }

  return {
    currentQty,
    averageCost: currentAvg,
    totalValuation,
    issueCogs,
  };
}

/**
 * Calculates FIFO (First-In, First-Out) Cost and remaining stock layers.
 */
export function calculateFifoCost(movements: StockMovementItem[]): {
  currentQty: number;
  totalValuation: number;
  remainingLayers: StockLayer[];
  issueCogs: { docNo: string; quantity: number; cogs: number }[];
} {
  const layers: StockLayer[] = [];
  const issueCogs: { docNo: string; quantity: number; cogs: number }[] = [];

  for (const m of movements) {
    if (m.type === "RECEIPT") {
      const inCost = m.unitCost ?? (m.totalCost && m.quantity > 0 ? m.totalCost / m.quantity : 0);
      layers.push({
        receivedDate: m.date,
        docNo: m.docNo,
        initialQty: m.quantity,
        remainingQty: m.quantity,
        unitCost: inCost,
      });
    } else if (m.type === "ISSUE") {
      let needed = m.quantity;
      let cogs = 0;

      for (const layer of layers) {
        if (needed <= 0) break;
        if (layer.remainingQty <= 0) continue;

        const take = Math.min(layer.remainingQty, needed);
        cogs += Math.round(take * layer.unitCost * 100) / 100;
        layer.remainingQty -= take;
        needed -= take;
      }

      issueCogs.push({
        docNo: m.docNo,
        quantity: m.quantity - needed,
        cogs: Math.round(cogs * 100) / 100,
      });
    }
  }

  const activeLayers = layers.filter((l) => l.remainingQty > 0);
  const currentQty = activeLayers.reduce((s, l) => s + l.remainingQty, 0);
  const totalValuation =
    Math.round(activeLayers.reduce((s, l) => s + l.remainingQty * l.unitCost, 0) * 100) / 100;

  return {
    currentQty,
    totalValuation,
    remainingLayers: activeLayers,
    issueCogs,
  };
}

/**
 * Distributes Landed Costs (freight, duty, insurance) to incoming goods.
 */
export function allocateLandedCost(
  receipts: { itemCode: string; quantity: number; unitPrice: number }[],
  additionalCost: number,
  method: "BY_VALUE" | "BY_QUANTITY" = "BY_VALUE",
): {
  itemCode: string;
  quantity: number;
  originalUnitPrice: number;
  allocatedLandedCost: number;
  finalUnitCost: number;
}[] {
  const totalBase =
    method === "BY_VALUE"
      ? receipts.reduce((s, r) => s + r.quantity * r.unitPrice, 0)
      : receipts.reduce((s, r) => s + r.quantity, 0);

  if (totalBase <= 0 || additionalCost <= 0) {
    return receipts.map((r) => ({
      itemCode: r.itemCode,
      quantity: r.quantity,
      originalUnitPrice: r.unitPrice,
      allocatedLandedCost: 0,
      finalUnitCost: r.unitPrice,
    }));
  }

  return receipts.map((r) => {
    const baseShare = method === "BY_VALUE" ? r.quantity * r.unitPrice : r.quantity;
    const allocated = Math.round(((baseShare / totalBase) * additionalCost) * 100) / 100;
    const unitAddition = r.quantity > 0 ? allocated / r.quantity : 0;
    const finalUnitCost = Math.round((r.unitPrice + unitAddition) * 100) / 100;

    return {
      itemCode: r.itemCode,
      quantity: r.quantity,
      originalUnitPrice: r.unitPrice,
      allocatedLandedCost: allocated,
      finalUnitCost,
    };
  });
}

/**
 * Generates Balanced Stock Adjustment Journal Lines for Physical Count Variances.
 */
export function generateStockVarianceJournal(params: {
  itemCode: string;
  itemName: string;
  bookQty: number;
  physicalQty: number;
  unitCost: number;
  inventoryAccountCode?: string; // 114101 (สินค้าสำเร็จรูป)
  varianceAccountCode?: string; // 520201 (ผลต่างการตรวจนับสต๊อก/สินค้าขาดหาย)
}): {
  varianceQty: number;
  varianceAmount: number;
  isShortage: boolean;
  journalLines: { accountCode: string; debit: number; credit: number; description: string }[];
} {
  const {
    itemCode,
    itemName,
    bookQty,
    physicalQty,
    unitCost,
    inventoryAccountCode = "114101",
    varianceAccountCode = "520201",
  } = params;

  const varianceQty = physicalQty - bookQty;
  const varianceAmount = Math.round(Math.abs(varianceQty) * unitCost * 100) / 100;
  const isShortage = varianceQty < 0; // Shortage = สินค้าขาดหาย

  const journalLines: { accountCode: string; debit: number; credit: number; description: string }[] = [];

  if (varianceAmount > 0) {
    if (isShortage) {
      // Physical < Book: สินค้าขาดหาย
      // Dr. ผลต่างการตรวจนับสต๊อก / ต้นทุนสินค้าขาดหาย (5xxx)
      // Cr. สินค้าคงเหลือ (1xxx)
      journalLines.push({
        accountCode: varianceAccountCode,
        debit: varianceAmount,
        credit: 0,
        description: `ปรับปรุงสินค้าขาดจากการตรวจนับ - ${itemName} (${itemCode}) จำนวน ${Math.abs(varianceQty)} หน่วย`,
      });
      journalLines.push({
        accountCode: inventoryAccountCode,
        debit: 0,
        credit: varianceAmount,
        description: `ตัดลดยอดสินค้าคงเหลือ - ${itemName} (${itemCode})`,
      });
    } else {
      // Physical > Book: สินค้าเกิน
      // Dr. สินค้าคงเหลือ (1xxx)
      // Cr. ปรับปรุงสินค้าเกิน / รายได้อื่น (4xxx/5xxx)
      journalLines.push({
        accountCode: inventoryAccountCode,
        debit: varianceAmount,
        credit: 0,
        description: `ปรับปรุงเพิ่มยอดสินค้าคงเหลือจากการตรวจนับ - ${itemName} (${itemCode}) จำนวน ${varianceQty} หน่วย`,
      });
      journalLines.push({
        accountCode: varianceAccountCode,
        debit: 0,
        credit: varianceAmount,
        description: `ผลต่างสินค้าเกินจากการตรวจนับ - ${itemName} (${itemCode})`,
      });
    }
  }

  return {
    varianceQty,
    varianceAmount,
    isShortage,
    journalLines,
  };
}

/**
 * Calculates Periodic Inventory Cost of Goods Sold (COGS) and Closing Entry.
 * Formula: Beginning Inventory + Net Purchases - Ending Inventory = COGS
 */
export function calculatePeriodicStockClosing(params: {
  beginningInventory: number;
  netPurchases: number;
  endingInventoryPhysical: number;
  inventoryAccountCode?: string; // 114101
  cogsAccountCode?: string; // 510101
  purchaseAccountCode?: string; // 510201
}): {
  cogs: number;
  closingJournalLines: { accountCode: string; debit: number; credit: number; description: string }[];
} {
  const {
    beginningInventory,
    netPurchases,
    endingInventoryPhysical,
    inventoryAccountCode = "114101",
    cogsAccountCode = "510101",
    purchaseAccountCode = "510201",
  } = params;

  // COGS = Beg + Purchase - End
  const cogs = Math.round((beginningInventory + netPurchases - endingInventoryPhysical) * 100) / 100;

  const closingJournalLines = [
    // 1. Dr. Ending Inventory (Assets)
    {
      accountCode: inventoryAccountCode,
      debit: endingInventoryPhysical,
      credit: 0,
      description: "บันทึกสินค้าคงเหลือปลายงวด (ตรวจนับจริง)",
    },
    // 2. Dr. Cost of Goods Sold (Expenses)
    {
      accountCode: cogsAccountCode,
      debit: Math.max(0, cogs),
      credit: 0,
      description: "ต้นทุนขายประจำงวด (Periodic COGS)",
    },
    // 3. Cr. Beginning Inventory (Clear Beg)
    {
      accountCode: inventoryAccountCode,
      debit: 0,
      credit: beginningInventory,
      description: "โอนปิดสินค้าคงเหลือต้นงวด",
    },
    // 4. Cr. Net Purchases (Clear Purchases)
    {
      accountCode: purchaseAccountCode,
      debit: 0,
      credit: netPurchases,
      description: "โอนปิดบัญชีซื้อสินค้าสุทธิ",
    },
  ];

  return {
    cogs,
    closingJournalLines,
  };
}

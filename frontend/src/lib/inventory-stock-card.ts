/**
 * Thai Revenue Department Standard Stock Card (รายงานสินค้าและวัตถุดิบ / บัญชีคุมพิเศษสินค้า)
 * Modeled after BC Champ: ICRepSpecialAccountView.cpp:386-405
 * Complies with Section 87(3) of the Thai Revenue Code (มาตรา 87(3) แห่งประมวลรัษฎากร).
 *
 * Tracks:
 * - Date, DocNo, TaxInvoiceNo, DocType, Description
 * - IN (รับเข้า): Qty, Unit Cost, Total Value
 * - OUT (จ่ายออก): Qty, Unit Cost, Total Value
 * - BALANCE (คงเหลือ): Qty, Moving Weighted Average Unit Cost, Total Value
 */

export type StockMovementType =
  | "beginning"   // ยอดยกมา
  | "purchase"    // ซื้อสินค้า (IN)
  | "receive"     // รับสินค้าสำเร็จรูป / รับคืน (IN)
  | "sale"        // ขายสินค้า (OUT)
  | "issue"       // เบิกใช้สินค้า/วัตถุดิบ (OUT)
  | "transfer_in" // โอนเข้าคลัง (IN)
  | "transfer_out"// โอนออกจากคลัง (OUT)
  | "adjust_add"  // ปรับปรุงยอดเพิ่ม (IN)
  | "adjust_sub"; // ปรับปรุงยอดลด (OUT)

export interface RawStockMovement {
  id: string;
  date: string;         // YYYY-MM-DD
  docno: string;        // เลขที่เอกสาร
  taxno?: string;       // เลขที่ใบกำกับภาษี
  docType: StockMovementType;
  docTypeName?: string; // ชื่อประเภทเอกสารภาษาไทย
  description?: string; // รายละเอียด / ผู้ขาย / ลูกค้า
  itemcode: string;
  itemname?: string;
  unitcode?: string;
  warehousecode?: string;
  qty: number;          // Positive quantity
  unitCost?: number;    // Specified cost for IN transactions
  totalAmount?: number; // Specified amount for IN transactions
}

export interface StockCardLine {
  date: string;
  docno: string;
  taxno: string;
  docType: StockMovementType;
  docTypeName: string;
  description: string;
  
  // IN
  inQty: number;
  inCost: number;
  inTotal: number;

  // OUT
  outQty: number;
  outCost: number;
  outTotal: number;

  // BALANCE
  balanceQty: number;
  balanceCost: number;
  balanceTotal: number;
}

export interface StockCardSummary {
  itemcode: string;
  itemname: string;
  unitcode: string;
  warehousecode: string;
  fromDate: string;
  toDate: string;
  beginningQty: number;
  beginningCost: number;
  beginningTotal: number;
  totalInQty: number;
  totalInTotal: number;
  totalOutQty: number;
  totalOutTotal: number;
  endingQty: number;
  endingCost: number;
  endingTotal: number;
  lines: StockCardLine[];
}

export const MOVEMENT_TYPE_NAMES: Record<StockMovementType, string> = {
  beginning: "ยอดยกมาต้นงวด",
  purchase: "ซื้อสินค้า",
  receive: "รับสินค้าสำเร็จรูป",
  sale: "ขายสินค้า",
  issue: "เบิกใช้สินค้า/วัตถุดิบ",
  transfer_in: "โอนย้ายเข้าคลัง",
  transfer_out: "โอนย้ายออกจากคลัง",
  adjust_add: "ปรับปรุงเพิ่มสต็อก",
  adjust_sub: "ปรับปรุงลดสต็อก",
};

/**
 * Calculates moving average stock card lines from raw movements
 */
export function calculateStockCard(
  itemcode: string,
  movements: RawStockMovement[],
  options: {
    itemname?: string;
    unitcode?: string;
    warehousecode?: string;
    fromDate?: string;
    toDate?: string;
    beginningQty?: number;
    beginningCost?: number;
  } = {}
): StockCardSummary {
  const itemname = options.itemname || itemcode;
  const unitcode = options.unitcode || "ชิ้น";
  const warehousecode = options.warehousecode || "WH-01";
  const fromDate = options.fromDate || "";
  const toDate = options.toDate || "";

  let currentQty = Math.max(0, options.beginningQty ?? 0);
  let currentCost = Math.max(0, options.beginningCost ?? 0);
  let currentTotal = Math.round(currentQty * currentCost * 100) / 100;

  const beginningQty = currentQty;
  const beginningCost = currentCost;
  const beginningTotal = currentTotal;

  // Sort movements chronologically
  const sorted = [...movements]
    .filter((m) => m.itemcode === itemcode)
    .sort((a, b) => a.date.localeCompare(b.date) || a.docno.localeCompare(b.docno));

  const lines: StockCardLine[] = [];
  let totalInQty = 0;
  let totalInTotal = 0;
  let totalOutQty = 0;
  let totalOutTotal = 0;

  // Initial beginning balance line if there is an opening stock
  if (beginningQty > 0) {
    lines.push({
      date: fromDate || sorted[0]?.date || new Date().toISOString().split("T")[0],
      docno: "-",
      taxno: "-",
      docType: "beginning",
      docTypeName: MOVEMENT_TYPE_NAMES.beginning,
      description: "ยอดยกมาต้นปี/ต้นงวด",
      inQty: 0,
      inCost: 0,
      inTotal: 0,
      outQty: 0,
      outCost: 0,
      outTotal: 0,
      balanceQty: beginningQty,
      balanceCost: beginningCost,
      balanceTotal: beginningTotal,
    });
  }

  for (const m of sorted) {
    const isIncoming = ["purchase", "receive", "transfer_in", "adjust_add"].includes(m.docType);
    let inQty = 0;
    let inCost = 0;
    let inTotal = 0;
    let outQty = 0;
    let outCost = 0;
    let outTotal = 0;

    if (isIncoming) {
      inQty = Math.abs(m.qty);
      inCost = m.unitCost ?? (m.totalAmount && inQty > 0 ? m.totalAmount / inQty : currentCost);
      inTotal = m.totalAmount ?? Math.round(inQty * inCost * 100) / 100;

      totalInQty += inQty;
      totalInTotal += inTotal;

      const newQty = currentQty + inQty;
      const newTotal = currentTotal + inTotal;
      const newCost = newQty > 0 ? Math.round((newTotal / newQty) * 10000) / 10000 : 0;

      currentQty = newQty;
      currentTotal = Math.round(newTotal * 100) / 100;
      currentCost = newCost;
    } else {
      outQty = Math.abs(m.qty);
      outCost = currentCost;
      outTotal = Math.round(outQty * outCost * 100) / 100;

      totalOutQty += outQty;
      totalOutTotal += outTotal;

      const newQty = Math.max(0, currentQty - outQty);
      const newTotal = Math.max(0, currentTotal - outTotal);
      const newCost = newQty > 0 ? currentCost : 0;

      currentQty = newQty;
      currentTotal = Math.round(newTotal * 100) / 100;
      currentCost = newCost;
    }

    lines.push({
      date: m.date,
      docno: m.docno,
      taxno: m.taxno || "-",
      docType: m.docType,
      docTypeName: m.docTypeName || MOVEMENT_TYPE_NAMES[m.docType] || m.docType,
      description: m.description || "-",
      inQty,
      inCost,
      inTotal,
      outQty,
      outCost,
      outTotal,
      balanceQty: currentQty,
      balanceCost: currentCost,
      balanceTotal: currentTotal,
    });
  }

  return {
    itemcode,
    itemname,
    unitcode,
    warehousecode,
    fromDate,
    toDate,
    beginningQty,
    beginningCost,
    beginningTotal,
    totalInQty,
    totalInTotal,
    totalOutQty,
    totalOutTotal,
    endingQty: currentQty,
    endingCost: currentCost,
    endingTotal: currentTotal,
    lines,
  };
}

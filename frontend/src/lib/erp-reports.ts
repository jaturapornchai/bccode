// Unified ERP Reporting Engine for Thai SMEs & Thai Accounting
// Covers Inventory, Sales, Purchase, AR/AP Aging, Gross Profit, and DBD XBRL Export

import { authFetch } from "@/lib/client-auth-session";
import { catalogText } from "@/lib/catalog-text";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";

export type ReportCategory = "inventory" | "sales" | "purchase" | "ar" | "ap" | "xbrl";

export interface ReportColumn {
  key: string;
  label: { th: string; en: string };
  align?: "left" | "center" | "right";
  isNumeric?: boolean;
  isCurrency?: boolean;
}

export interface ErpReportConfig {
  route: string;
  code: string;
  category: ReportCategory;
  title: { th: string; en: string };
  description: { th: string; en: string };
  columns: ReportColumn[];
  defaultSortKey: string;
}

export const ERP_REPORT_CONFIGS: ErpReportConfig[] = [
  // --- สินค้าคงคลัง (Inventory) ---
  {
    route: "/report/stockbalanceitem",
    code: "stock_balance_item",
    category: "inventory",
    title: { th: "รายงานยอดคงเหลือสินค้า-ตามสินค้า", en: "Stock Balance by Item" },
    description: { th: "รายงานสรุปยอดสินค้าคงเหลือและมูลค่าสต็อก แยกรายรหัสสินค้า", en: "Inventory on-hand balance and valuation by item" },
    defaultSortKey: "itemcode",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "unitname", label: { th: "หน่วยนับ", en: "Unit" }, align: "center" },
      { key: "qty", label: { th: "จำนวนคงเหลือ", en: "Qty" }, align: "right", isNumeric: true },
      { key: "avgcost", label: { th: "ต้นทุนเฉลี่ย", en: "Avg Cost" }, align: "right", isCurrency: true },
      { key: "totalcost", label: { th: "มูลค่าคงเหลือ", en: "Total Value" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/stockbalancewarehouse",
    code: "stock_balance_warehouse",
    category: "inventory",
    title: { th: "รายงานยอดคงเหลือสินค้า-ตามคลัง,สินค้า", en: "Stock Balance by Warehouse" },
    description: { th: "แสดงยอดสินค้าคงเหลือแยกตามคลังสินค้าและสาขา", en: "Inventory on-hand balance grouped by warehouse and branch" },
    defaultSortKey: "warehouse",
    columns: [
      { key: "warehouse", label: { th: "คลังสินค้า", en: "Warehouse" } },
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "unitname", label: { th: "หน่วย", en: "Unit" }, align: "center" },
      { key: "qty", label: { th: "จำนวนคงเหลือ", en: "Qty" }, align: "right", isNumeric: true },
      { key: "totalcost", label: { th: "มูลค่าสต็อก", en: "Value" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/stockbalancelocation",
    code: "stock_balance_location",
    category: "inventory",
    title: { th: "คงเหลือตามที่เก็บสินค้า", en: "Stock Balance by Location/Bin" },
    description: { th: "แสดงตำแหน่งจัดเก็บ แถว ชั้น ช่อง (Bin Location) ของสินค้า", en: "Stock balance pinpointed by location, aisle and bin" },
    defaultSortKey: "location",
    columns: [
      { key: "location", label: { th: "ที่เก็บ", en: "Location/Bin" } },
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "qty", label: { th: "จำนวน", en: "Qty" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/reportdedebistockbalance",
    code: "general_stock_balance",
    category: "inventory",
    title: { th: "รายงานสินค้าคงเหลือ", en: "Comprehensive Stock Balance" },
    description: { th: "รายงานสินค้าคงเหลือภาพรวมทั้งบริษัทพร้อมยอดจองและยอดพร้อมขาย", en: "Company-wide stock balance with reservations and available to sell" },
    defaultSortKey: "itemcode",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "qty", label: { th: "คงเหลือจริง", en: "Physical Qty" }, align: "right", isNumeric: true },
      { key: "reservedqty", label: { th: "ยอดสั่งจองสินค้า", en: "Reserved" }, align: "right", isNumeric: true },
      { key: "availableqty", label: { th: "พร้อมขาย (ATP)", en: "Available" }, align: "right", isNumeric: true },
      { key: "totalcost", label: { th: "มูลค่ารวม", en: "Total Cost" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/lowstock",
    code: "low_stock_report",
    category: "inventory",
    title: { th: "รายงานยอดคงเหลือสินค้าที่ถึงจุดสั่งซื้อ", en: "Low Stock Alert" },
    description: { th: "เตือนรายการสินค้าที่ยอดคงเหลือต่ำกว่าจุดสั่งซื้อซ้ำ เพื่อความต่อเนื่องของธุรกิจ", en: "Items below minimum stock level and safety threshold" },
    defaultSortKey: "shortage",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "minqty", label: { th: "จุดสั่งซื้อขั้นต่ำ", en: "Min Point" }, align: "right", isNumeric: true },
      { key: "qty", label: { th: "ยอดคงเหลือปัจจุบัน", en: "Current Qty" }, align: "right", isNumeric: true },
      { key: "shortage", label: { th: "จำนวนที่ขาด", en: "Shortage" }, align: "right", isNumeric: true },
      { key: "reorderqty", label: { th: "แนะนำสั่งซื้อ", en: "Suggested Reorder" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/expiringstock",
    code: "expiring_stock_report",
    category: "inventory",
    title: { th: "สินค้าใกล้หมดอายุ", en: "Expiring Stock Report" },
    description: { th: "ติดตามวันหมดอายุของสินค้าตามล็อต (FEFO Management)", en: "Stock expiration tracking by batch and lot number" },
    defaultSortKey: "expirydate",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "lotno", label: { th: "เลขล็อต (Lot No.)", en: "Lot No." } },
      { key: "expirydate", label: { th: "วันหมดอายุ", en: "Expiry Date" }, align: "center" },
      { key: "daysremaining", label: { th: "วันคงเหลือ", en: "Days Left" }, align: "right", isNumeric: true },
      { key: "qty", label: { th: "จำนวน", en: "Qty" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/stockmovementcost",
    code: "stock_movement_cost",
    category: "inventory",
    title: { th: "รายงานบัญชีคุมพิเศษสินค้า", en: "Stock Card with Cost Movement" },
    description: { th: "บัตรคุมสต็อกแสดงการรับ-จ่าย-คงเหลือ พร้อมต้นทุน FIFO และถัวเฉลี่ยถ่วงน้ำหนัก", en: "Stock card ledger with in-out-balance quantities and moving costs" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่เอกสาร", en: "Doc No." } },
      { key: "transname", label: { th: "ประเภทรายการ", en: "Transaction" } },
      { key: "inqty", label: { th: "รับเข้า", en: "In Qty" }, align: "right", isNumeric: true },
      { key: "outqty", label: { th: "จ่ายออก", en: "Out Qty" }, align: "right", isNumeric: true },
      { key: "balanceqty", label: { th: "ยอดคงเหลือ", en: "Balance Qty" }, align: "right", isNumeric: true },
      { key: "unitcost", label: { th: "ต้นทุน/หน่วย", en: "Unit Cost" }, align: "right", isCurrency: true },
      { key: "totalcost", label: { th: "มูลค่าคงเหลือ", en: "Balance Value" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportstockmovement",
    code: "stock_movement_simple",
    category: "inventory",
    title: { th: "รายงานเคลื่อนไหวสินค้า", en: "Stock Movement Report" },
    description: { th: "สรุปการเคลื่อนไหวสต็อกเข้า-ออกตามช่วงเวลา", en: "Periodical in-out stock movement log" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่เอกสาร", en: "Doc No." } },
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "inqty", label: { th: "รับ", en: "In" }, align: "right", isNumeric: true },
      { key: "outqty", label: { th: "จ่าย", en: "Out" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/stocklotmovement",
    code: "stock_lot_movement",
    category: "inventory",
    title: { th: "ความเคลื่อนไหวล็อตสินค้า", en: "Lot Tracking Movement" },
    description: { th: "ตรวจสอบการรับเข้าและจ่ายออกของสินค้าแยกรายล็อตการผลิต", en: "Traceability ledger of batch/lot receipts and shipments" },
    defaultSortKey: "lotno",
    columns: [
      { key: "lotno", label: { th: "เลขล็อต", en: "Lot No." } },
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เอกสาร", en: "Doc No." } },
      { key: "qty", label: { th: "จำนวน", en: "Qty" }, align: "right", isNumeric: true },
    ],
  },

  // --- งานขายและวิเคราะห์ (Sales & Analytics) ---
  {
    route: "/report/reportdedebisales",
    code: "sales_summary",
    category: "sales",
    title: { th: "รายงานขาย", en: "Sales Summary Report" },
    description: { th: "สรุปยอดขายแยกตามหมวดหมู่ ยอดขายก่อนภาษี ภาษี และยอดสุทธิ", en: "Summary of gross sales, discounts, VAT and net sales" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่เอกสาร", en: "Invoice No." } },
      { key: "custname", label: { th: "ชื่อลูกค้า", en: "Customer" } },
      { key: "subtotal", label: { th: "มูลค่าก่อนภาษี", en: "Subtotal" }, align: "right", isCurrency: true },
      { key: "vatamount", label: { th: "ภาษีขาย 7%", en: "VAT 7%" }, align: "right", isCurrency: true },
      { key: "totalamount", label: { th: "รวมทั้งสิ้น", en: "Total" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportdedebisalesdaily",
    code: "sales_daily",
    category: "sales",
    title: { th: "รายงานรายวันขายสินค้าและบริการ", en: "Daily Sales Report" },
    description: { th: "สรุปยอดขายประจำวันเพื่อตรวจสอบเงินสด ธนาคาร และลูกหนี้", en: "Daily sales audit report by payment method" },
    defaultSortKey: "salesdate",
    columns: [
      { key: "salesdate", label: { th: "วันที่ขาย", en: "Date" }, align: "center" },
      { key: "doccount", label: { th: "จำนวนบิล", en: "Bills" }, align: "right", isNumeric: true },
      { key: "cashsales", label: { th: "เงินสด", en: "Cash" }, align: "right", isCurrency: true },
      { key: "transfersales", label: { th: "เงินโอน", en: "Transfer" }, align: "right", isCurrency: true },
      { key: "creditsales", label: { th: "ขายเชื่อ (ลูกหนี้)", en: "Credit" }, align: "right", isCurrency: true },
      { key: "totalamount", label: { th: "ยอดขายรวม", en: "Total Sales" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/salesbyseller",
    code: "sales_by_seller",
    category: "sales",
    title: { th: "รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย", en: "Sales by Salesperson" },
    description: { th: "วิเคราะห์ผลงานพนักงานขายและคำนวณค่าคอมมิชชั่น", en: "Sales performance ranking and commission basis by seller" },
    defaultSortKey: "totalamount",
    columns: [
      { key: "sellername", label: { th: "ชื่อพนักงานขาย", en: "Salesperson" } },
      { key: "target", label: { th: "เป้ายอดขาย", en: "Target" }, align: "right", isCurrency: true },
      { key: "actual", label: { th: "ยอดขายจริง", en: "Actual" }, align: "right", isCurrency: true },
      { key: "achievedpercent", label: { th: "% บรรลุเป้า", en: "% Achieved" }, align: "right", isNumeric: true },
      { key: "commission", label: { th: "คอมมิชชั่น", en: "Commission" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/salesreportbydocument",
    code: "sales_by_document",
    category: "sales",
    title: { th: "รายงานขายตามเอกสาร", en: "Sales by Document Register" },
    description: { th: "สมุดทะเบียนเอกสารขายเรียงตามเลขที่บิลและวันที่", en: "Sales register ordered sequentially by document number" },
    defaultSortKey: "docno",
    columns: [
      { key: "docno", label: { th: "เลขที่บิล", en: "Doc No." } },
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "custcode", label: { th: "รหัสลูกค้า", en: "Cust Code" } },
      { key: "custname", label: { th: "ชื่อลูกค้า", en: "Customer" } },
      { key: "totalamount", label: { th: "ยอดเงิน", en: "Amount" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportgrossprofitbydocument",
    code: "gross_profit_document",
    category: "sales",
    title: { th: "กำไรขั้นต้นตามเอกสาร", en: "Gross Profit by Document" },
    description: { th: "วิเคราะห์ต้นทุนขาย กำไรขั้นต้น และมาร์จิ้น (% GP) ในแต่ละบิล", en: "Gross profit margin analysis per individual invoice" },
    defaultSortKey: "docno",
    columns: [
      { key: "docno", label: { th: "เลขที่บิล", en: "Invoice" } },
      { key: "custname", label: { th: "ลูกค้า", en: "Customer" } },
      { key: "salesrevenue", label: { th: "ยอดขาย", en: "Revenue" }, align: "right", isCurrency: true },
      { key: "costofgoods", label: { th: "ต้นทุนขาย", en: "COGS" }, align: "right", isCurrency: true },
      { key: "grossprofit", label: { th: "กำไรขั้นต้น", en: "Gross Profit" }, align: "right", isCurrency: true },
      { key: "marginpercent", label: { th: "% GP", en: "% Margin" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/reportgrossprofitbyproduct",
    code: "gross_profit_product",
    category: "sales",
    title: { th: "กำไรขั้นต้นตามสินค้า", en: "Gross Profit by Product" },
    description: { th: "วิเคราะห์สินค้าขายดีและสินค้าที่สร้างกำไรสูงสุดให้ธุรกิจ", en: "Profitability matrix per product item" },
    defaultSortKey: "grossprofit",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "soldqty", label: { th: "จำนวนที่ขาย", en: "Sold Qty" }, align: "right", isNumeric: true },
      { key: "salesrevenue", label: { th: "ยอดขายรวม", en: "Sales" }, align: "right", isCurrency: true },
      { key: "costofgoods", label: { th: "ต้นทุนรวม", en: "Cost" }, align: "right", isCurrency: true },
      { key: "grossprofit", label: { th: "กำไรขั้นต้น", en: "Gross Profit" }, align: "right", isCurrency: true },
      { key: "marginpercent", label: { th: "% GP", en: "% Margin" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/reportdedebisalereturn",
    code: "sales_return_report",
    category: "sales",
    title: { th: "รายงานคืนขาย", en: "Sales Return Report" },
    description: { th: "วิเคราะห์สาเหตุการคืนสินค้าและลดหนี้ให้ลูกค้า", en: "Customer returns and credit note issue log" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่ใบลดหนี้", en: "CN No." } },
      { key: "refdocno", label: { th: "อ้างอิงบิลเดิม", en: "Ref Invoice" } },
      { key: "custname", label: { th: "ลูกค้า", en: "Customer" } },
      { key: "reason", label: { th: "สาเหตุการส่งคืน", en: "Reason" } },
      { key: "totalamount", label: { th: "ยอดลดหนี้", en: "Amount" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/salesbycustomer",
    code: "sales_by_customer",
    category: "sales",
    title: { th: "ยอดขายตามลูกค้า", en: "Sales by Customer" },
    description: { th: "จัดอันดับลูกค้าชั้นดี (Top Customers) ตามปริมาณยอดซื้อ", en: "Customer revenue breakdown and VIP tier ranking" },
    defaultSortKey: "totalamount",
    columns: [
      { key: "custcode", label: { th: "รหัสลูกค้า", en: "Code" } },
      { key: "custname", label: { th: "ชื่อลูกค้า", en: "Customer" } },
      { key: "billcount", label: { th: "จำนวนบิล", en: "Orders" }, align: "right", isNumeric: true },
      { key: "totalamount", label: { th: "ยอดซื้อรวม", en: "Total Purchases" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/salesbychannel",
    code: "sales_by_channel",
    category: "sales",
    title: { th: "ยอดขายตามช่องทางขาย", en: "Sales by Channel" },
    description: { th: "เปรียบเทียบยอดขายหน้าร้าน, Shopee, Lazada, TikTok, และตัวแทนจำหน่าย", en: "Omnichannel sales distribution analysis" },
    defaultSortKey: "totalamount",
    columns: [
      { key: "channel", label: { th: "ช่องทางการจัดจำหน่าย", en: "Channel" } },
      { key: "billcount", label: { th: "จำนวนคำสั่งซื้อ", en: "Orders" }, align: "right", isNumeric: true },
      { key: "totalamount", label: { th: "ยอดขายรวม", en: "Total Sales" }, align: "right", isCurrency: true },
      { key: "sharepercent", label: { th: "สัดส่วน (%)", en: "Share %" }, align: "right", isNumeric: true },
    ],
  },

  // --- งานจัดซื้อ & ค่าใช้จ่าย (Purchasing) ---
  {
    route: "/report/reportdedebipurchase",
    code: "purchase_summary",
    category: "purchase",
    title: { th: "รายงานซื้อ", en: "Purchase Summary Report" },
    description: { th: "สรุปยอดการสั่งซื้อสินค้าและบริการแยกตามเจ้าหนี้", en: "Summary of vendor purchases and expenses" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่ใบซื้อ", en: "PO/Bill No." } },
      { key: "vendorname", label: { th: "เจ้าหนี้", en: "Vendor" } },
      { key: "subtotal", label: { th: "มูลค่าก่อนภาษี", en: "Subtotal" }, align: "right", isCurrency: true },
      { key: "vatamount", label: { th: "ภาษีซื้อ 7%", en: "VAT 7%" }, align: "right", isCurrency: true },
      { key: "totalamount", label: { th: "ยอดรวมเงินสุทธิ", en: "Total" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/purchasebyproduct",
    code: "purchase_by_product",
    category: "purchase",
    title: { th: "รายงานซื้อตามสินค้า", en: "Purchase by Product" },
    description: { th: "วิเคราะห์ปริมาณและราคาซื้อเฉลี่ยของสินค้าแต่ละรายการ", en: "Product purchase volume and price trend" },
    defaultSortKey: "totalamount",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "purchasedqty", label: { th: "จำนวนที่ซื้อ", en: "Purchased Qty" }, align: "right", isNumeric: true },
      { key: "avgprice", label: { th: "ราคาเฉลี่ย", en: "Avg Price" }, align: "right", isCurrency: true },
      { key: "totalamount", label: { th: "ยอดซื้อรวม", en: "Total Value" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportdedebipurchasepartial",
    code: "purchase_partial_report",
    category: "purchase",
    title: { th: "รายงานทยอยรับ", en: "Partial Delivery Tracking" },
    description: { th: "ติดตามสินค้าค้างรับจากผู้ขายตามใบสั่งซื้อ (PO Backlog)", en: "Outstanding purchase orders and partial shipments status" },
    defaultSortKey: "pono",
    columns: [
      { key: "pono", label: { th: "เลขที่ PO", en: "PO No." } },
      { key: "vendorname", label: { th: "เจ้าหนี้", en: "Vendor" } },
      { key: "orderqty", label: { th: "จำนวนสั่งซื้อ", en: "Order Qty" }, align: "right", isNumeric: true },
      { key: "receivedqty", label: { th: "รับเข้าแล้ว", en: "Received" }, align: "right", isNumeric: true },
      { key: "pendingqty", label: { th: "ค้างส่งมอบ", en: "Pending" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/expensesummary",
    code: "expense_summary",
    category: "purchase",
    title: { th: "รายงานสรุปรายจ่าย", en: "Expense Summary Report" },
    description: { th: "สรุปค่าใช้จ่ายในการดำเนินงาน (OPEX) จัดกลุ่มตามผังบัญชีและแผนก", en: "Operating expense breakdown by chart of account and department" },
    defaultSortKey: "totalamount",
    columns: [
      { key: "accountcode", label: { th: "รหัสบัญชี", en: "Account" } },
      { key: "accountname", label: { th: "หมวดหมู่ค่าใช้จ่าย", en: "Expense Category" } },
      { key: "deptname", label: { th: "แผนก", en: "Department" } },
      { key: "totalamount", label: { th: "ยอดเงินรวม", en: "Amount" }, align: "right", isCurrency: true },
    ],
  },

  // --- ลูกหนี้และเจ้าหนี้ (Aging & Debt Collection) ---
  {
    route: "/report/araging",
    code: "ar_aging",
    category: "ar",
    title: { th: "รายงานอายุลูกหนี้", en: "Accounts Receivable Aging Report" },
    description: { th: "วิเคราะห์หนี้ค้างชำระแบ่งตามช่วงเวลา (ไม่เกินกำหนด, 1-30 วัน, 31-60 วัน, 61-90 วัน, เกิน 90 วัน)", en: "Debtor aging schedule with standard Thai accounting aging buckets" },
    defaultSortKey: "totaldue",
    columns: [
      { key: "custcode", label: { th: "รหัสลูกหนี้", en: "Code" } },
      { key: "custname", label: { th: "ชื่อลูกหนี้", en: "Customer / Debtor" } },
      { key: "currentdue", label: { th: "ยังไม่ถึงกำหนด", en: "Current" }, align: "right", isCurrency: true },
      { key: "days1_30", label: { th: "1-30 วัน", en: "1-30 Days" }, align: "right", isCurrency: true },
      { key: "days31_60", label: { th: "31-60 วัน", en: "31-60 Days" }, align: "right", isCurrency: true },
      { key: "days61_90", label: { th: "61-90 วัน", en: "61-90 Days" }, align: "right", isCurrency: true },
      { key: "over90", label: { th: "เกิน 90 วัน", en: "> 90 Days" }, align: "right", isCurrency: true },
      { key: "totaldue", label: { th: "ยอดหนี้รวม", en: "Total Due" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportdedebipaymentdaily",
    code: "ar_payment_daily",
    category: "ar",
    title: { th: "รายงานการรับชำระหนี้ประจำวัน", en: "Daily Debt Collection Register" },
    description: { th: "ทะเบียนใบเสร็จรับเงินและการตัดหนี้ลูกหนี้ประจำวัน", en: "Daily debtor settlement receipts and clearance log" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่รับเงิน", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่ใบเสร็จ", en: "Receipt No." } },
      { key: "custname", label: { th: "ลูกหนี้", en: "Customer" } },
      { key: "paymethod", label: { th: "วิธีชำระ", en: "Method" }, align: "center" },
      { key: "totalamount", label: { th: "ยอดรับชำระ", en: "Amount" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/apaging",
    code: "ap_aging",
    category: "ap",
    title: { th: "รายงานอายุเจ้าหนี้", en: "Accounts Payable Aging Report" },
    description: { th: "วิเคราะห์หนี้ที่ต้องจ่ายให้เจ้าหนี้ตามระยะเวลาครบกำหนดชำระเพื่อบริหารสภาพคล่อง", en: "Creditor aging schedule to manage vendor payments and liquidity" },
    defaultSortKey: "totaldue",
    columns: [
      { key: "vendorcode", label: { th: "รหัสเจ้าหนี้", en: "Code" } },
      { key: "vendorname", label: { th: "ชื่อเจ้าหนี้", en: "Vendor / Creditor" } },
      { key: "currentdue", label: { th: "ยังไม่ถึงกำหนด", en: "Current" }, align: "right", isCurrency: true },
      { key: "days1_30", label: { th: "1-30 วัน", en: "1-30 Days" }, align: "right", isCurrency: true },
      { key: "days31_60", label: { th: "31-60 วัน", en: "31-60 Days" }, align: "right", isCurrency: true },
      { key: "days61_90", label: { th: "61-90 วัน", en: "61-90 Days" }, align: "right", isCurrency: true },
      { key: "over90", label: { th: "เกิน 90 วัน", en: "> 90 Days" }, align: "right", isCurrency: true },
      { key: "totaldue", label: { th: "ยอดหนี้รวม", en: "Total Due" }, align: "right", isCurrency: true },
    ],
  },

  // --- ส่งออกงบการเงิน DBD XBRL ---
  {
    route: "/report/xbrl",
    code: "dbd_xbrl_export",
    category: "xbrl",
    title: { th: "ส่งออกงบการเงิน XBRL (ยื่น DBD)", en: "DBD XBRL Financial Statement Export" },
    description: { th: "ส่งออกไฟล์ข้อมูลทางการเงินตาม Taxonomy ของกรมพัฒนาธุรกิจการค้า กระทรวงพาณิชย์ สำหรับยื่นงบการเงินประจำปี", en: "Export statutory financial statements compliant with Thai DBD XBRL taxonomy" },
    defaultSortKey: "accountcode",
    columns: [
      { key: "accountcode", label: { th: "รหัสผังบัญชี", en: "GL Code" } },
      { key: "xbrltag", label: { th: "แท็กมาตรฐาน XBRL (DBD Taxonomy)", en: "XBRL Tag" } },
      { key: "description", label: { th: "รายการในงบการเงิน", en: "Statement Line" } },
      { key: "amount", label: { th: "ยอดเงินรอบปัจจุบัน", en: "Current Period" }, align: "right", isCurrency: true },
      { key: "prevamount", label: { th: "ยอดเงินรอบก่อน", en: "Prior Period" }, align: "right", isCurrency: true },
    ],
  },
];

const reportRouteMap = new Map<string, ErpReportConfig>(
  ERP_REPORT_CONFIGS.map((config) => [config.route, config]),
);

export function isErpReportRoute(route: string): boolean {
  const clean = route.split("?")[0];
  return reportRouteMap.has(clean);
}

export function getErpReportConfig(route: string): ErpReportConfig | undefined {
  const clean = route.split("?")[0];
  return reportRouteMap.get(clean);
}

export type ErpReportRow = Record<string, unknown>;

const API_READY_REPORTS: ReadonlySet<string> = new Set<string>([
  "sales_by_document",
  "gross_profit_document",
  "stock_balance_item",
  "stock_balance_warehouse",
]);

export function isErpReportApiReady(code: string): boolean {
  return API_READY_REPORTS.has(code);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function toNumber(value: unknown): number | undefined {
  if (typeof value === "number") {
    return Number.isFinite(value) ? value : undefined;
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function toText(value: unknown): string | undefined {
  if (typeof value === "string") return value;
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  return undefined;
}

type RowMapper = (source: Record<string, unknown>) => ErpReportRow;

function mapSalesByDocument(source: Record<string, unknown>): ErpReportRow {
  const row: ErpReportRow = {};
  const docno = toText(source.docno);
  if (docno !== undefined) row.docno = docno;
  const docdate = toText(source.docdate);
  if (docdate !== undefined) row.docdate = docdate;
  const custcode = toText(source.debtorcode);
  if (custcode !== undefined) row.custcode = custcode;
  const custname = toText(source.debtorname);
  if (custname !== undefined) row.custname = custname;
  const totalamount = toNumber(source.totalamount);
  if (totalamount !== undefined) row.totalamount = totalamount;
  return row;
}

function mapGrossProfitDocument(source: Record<string, unknown>): ErpReportRow {
  const row: ErpReportRow = {};
  const docno = toText(source.docno);
  if (docno !== undefined) row.docno = docno;
  const custname = toText(source.debtorname);
  if (custname !== undefined) row.custname = custname;

  const totalamount = toNumber(source.totalamount);
  const calcamount = toNumber(source.calcamount);
  const grossprofit = toNumber(source.grossprofit);

  if (totalamount !== undefined) row.salesrevenue = totalamount;
  if (calcamount !== undefined) row.costofgoods = calcamount;
  if (grossprofit !== undefined) row.grossprofit = grossprofit;

  if (grossprofit !== undefined && totalamount !== undefined) {
    const marginpercent =
      totalamount === 0
        ? 0
        : Math.round((grossprofit / totalamount) * 100 * 100) / 100;
    row.marginpercent = marginpercent;
  }
  return row;
}

function mapStockBalanceItem(source: Record<string, unknown>): ErpReportRow {
  const row: ErpReportRow = {};
  const itemcode = toText(source.itemcode);
  if (itemcode !== undefined) row.itemcode = itemcode;
  const itemname = toText(source.itemname);
  if (itemname !== undefined) row.itemname = itemname;
  const qty = toNumber(source.qty);
  if (qty !== undefined) row.qty = qty;
  const unitname = toText(source.unitcode);
  if (unitname !== undefined) row.unitname = unitname;
  const avgcost = toNumber(source.averagecost);
  if (avgcost !== undefined) row.avgcost = avgcost;
  const totalcost = toNumber(source.totalvalue);
  if (totalcost !== undefined) row.totalcost = totalcost;
  return row;
}

function mapStockBalanceWarehouse(source: Record<string, unknown>): ErpReportRow {
  const row: ErpReportRow = {};
  const warehouse = toText(source.whcode);
  if (warehouse !== undefined) row.warehouse = warehouse;
  const itemcode = toText(source.itemcode);
  if (itemcode !== undefined) row.itemcode = itemcode;
  const itemname = toText(source.itemname);
  if (itemname !== undefined) row.itemname = itemname;
  const qty = toNumber(source.qty);
  if (qty !== undefined) row.qty = qty;
  const unitname = toText(source.unitcode);
  if (unitname !== undefined) row.unitname = unitname;
  const totalcost = toNumber(source.totalvalue);
  if (totalcost !== undefined) row.totalcost = totalcost;
  return row;
}

const ROW_MAPPERS: Record<string, RowMapper> = {
  sales_by_document: mapSalesByDocument,
  gross_profit_document: mapGrossProfitDocument,
  stock_balance_item: mapStockBalanceItem,
  stock_balance_warehouse: mapStockBalanceWarehouse,
};

function extractRows(
  payload: unknown,
  key: "data" | "items",
): Record<string, unknown>[] | null {
  if (!isRecord(payload)) return null;
  const arr = payload[key];
  if (!Array.isArray(arr)) return null;
  const rows: Record<string, unknown>[] = [];
  for (const item of arr) {
    if (isRecord(item)) rows.push(item);
  }
  return rows;
}

export async function fetchErpReportData(params: {
  code: string;
  holdingcode: string;
  fromdate: string; // "YYYY-MM-DD"
  todate: string;   // "YYYY-MM-DD"
}): Promise<{ rows: ErpReportRow[]; error?: string }> {
  const { code, holdingcode, fromdate, todate } = params;

  if (!holdingcode) {
    return { rows: [], error: "holding_required" };
  }

  if (!isErpReportApiReady(code)) {
    return { rows: [], error: "report_not_available" };
  }

  const mapper = ROW_MAPPERS[code];
  if (!mapper) {
    return { rows: [], error: "report_not_available" };
  }

  const isDocumentReport =
    code === "sales_by_document" || code === "gross_profit_document";

  try {
    let payload: unknown;

    if (isDocumentReport) {
      const res = await authFetch("/api/goapi/api/report/sales/by-document", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          holdingcode,
          fromdate,
          todate,
          reporttype: "header",
          sortascending: true,
          limit: 1000,
          offset: 0,
        }),
      });
      if (!res.ok) {
        return {
          rows: [],
          error: res.status === 401 || res.status === 403 ? "unauthorized" : "load_failed",
        };
      }
      payload = await res.json();
      const rawRows = extractRows(payload, "data");
      if (rawRows === null) {
        return { rows: [], error: "load_failed" };
      }
      return { rows: rawRows.map(mapper) };
    }

    const url = `/api/goapi/api/reports/inventory-valuation?holdingcode=${encodeURIComponent(holdingcode)}`;
    const res = await authFetch(url);
    if (!res.ok) {
      return {
        rows: [],
        error: res.status === 401 || res.status === 403 ? "unauthorized" : "load_failed",
      };
    }
    payload = await res.json();
    const rawRows = extractRows(payload, "items");
    if (rawRows === null) {
      return { rows: [], error: "load_failed" };
    }
    return { rows: rawRows.map(mapper) };
  } catch {
    return { rows: [], error: "connection_error" };
  }
}

// 2026-09-16: every user-visible string above also lives in languages.tsv,
// keyed by `<code>.<part>`. The literals stay as the offline fallback.
const catalogKeys: Record<string, string> = {
  "stock_balance_item.title": "stock_balance_item",
  "stock_balance_item.description": "rpt_inventory_on_hand_balance_and",
  "stock_balance_item.col.itemcode": "import_productdetail.product_code",
  "stock_balance_item.col.itemname": "import_productdetail.product_name",
  "stock_balance_item.col.unitname": "import_productdetail.unit",
  "stock_balance_item.col.qty": "balance",
  "stock_balance_item.col.avgcost": "average_cost",
  "stock_balance_item.col.totalcost": "balance_amount",
  "stock_balance_warehouse.title": "stock_balance_warehouse",
  "stock_balance_warehouse.description": "rpt_inventory_on_hand_balance_grouped",
  "stock_balance_warehouse.col.warehouse": "form_design_cat_inventory",
  "stock_balance_warehouse.col.itemcode": "import_productdetail.product_code",
  "stock_balance_warehouse.col.itemname": "import_productdetail.product_name",
  "stock_balance_warehouse.col.unitname": "chatbot_product_unit",
  "stock_balance_warehouse.col.qty": "balance",
  "stock_balance_warehouse.col.totalcost": "rpt_value",
  "stock_balance_location.title": "stock_balance_location",
  "stock_balance_location.description": "rpt_stock_balance_pinpointed_by_location",
  "stock_balance_location.col.location": "location",
  "stock_balance_location.col.itemcode": "import_productdetail.product_code",
  "stock_balance_location.col.itemname": "import_productdetail.product_name",
  "stock_balance_location.col.qty": "common_qty",
  "general_stock_balance.title": "inventory",
  "general_stock_balance.description": "rpt_company_wide_stock_balance_with",
  "general_stock_balance.col.itemcode": "import_productdetail.product_code",
  "general_stock_balance.col.itemname": "import_productdetail.product_name",
  "general_stock_balance.col.qty": "rpt_physical_qty",
  "general_stock_balance.col.reservedqty": "rpt_reserved",
  "general_stock_balance.col.availableqty": "rpt_available",
  "general_stock_balance.col.totalcost": "payment_daily_totalamount_column",
  "low_stock_report.title": "low_stock_alert",
  "low_stock_report.description": "rpt_items_below_minimum_stock_level",
  "low_stock_report.col.itemcode": "import_productdetail.product_code",
  "low_stock_report.col.itemname": "import_productdetail.product_name",
  "low_stock_report.col.minqty": "rpt_min_point",
  "low_stock_report.col.qty": "rpt_current_qty",
  "low_stock_report.col.shortage": "rpt_shortage",
  "low_stock_report.col.reorderqty": "rpt_suggested_reorder",
  "expiring_stock_report.title": "expiring_stock_alert",
  "expiring_stock_report.description": "rpt_stock_expiration_tracking_by_batch",
  "expiring_stock_report.col.itemcode": "import_productdetail.product_code",
  "expiring_stock_report.col.itemname": "import_productdetail.product_name",
  "expiring_stock_report.col.lotno": "rpt_lot_no",
  "expiring_stock_report.col.expirydate": "coupon_expirydate",
  "expiring_stock_report.col.daysremaining": "rpt_days_left",
  "expiring_stock_report.col.qty": "common_qty",
  "stock_movement_cost.title": "stock_movement_cost",
  "stock_movement_cost.description": "rpt_stock_card_ledger_with_in",
  "stock_movement_cost.col.docdate": "date",
  "stock_movement_cost.col.docno": "docno",
  "stock_movement_cost.col.transname": "inquiry_type",
  "stock_movement_cost.col.inqty": "rpt_in_qty",
  "stock_movement_cost.col.outqty": "rpt_out_qty",
  "stock_movement_cost.col.balanceqty": "balanceqty",
  "stock_movement_cost.col.unitcost": "st_cost_per_unit",
  "stock_movement_cost.col.totalcost": "balance_amount",
  "stock_movement_simple.title": "product_movement",
  "stock_movement_simple.description": "rpt_periodical_in_out_stock_movement",
  "stock_movement_simple.col.docdate": "date",
  "stock_movement_simple.col.docno": "docno",
  "stock_movement_simple.col.itemcode": "import_productdetail.product_code",
  "stock_movement_simple.col.itemname": "import_productdetail.product_name",
  "stock_movement_simple.col.inqty": "rpt_in",
  "stock_movement_simple.col.outqty": "rpt_out",
  "stock_lot_movement.title": "stock_lot_movement",
  "stock_lot_movement.description": "rpt_traceability_ledger_of_batch_lot",
  "stock_lot_movement.col.lotno": "rpt_lot_no_2",
  "stock_lot_movement.col.itemcode": "import_productdetail.product_code",
  "stock_lot_movement.col.docdate": "date",
  "stock_lot_movement.col.docno": "doc_header",
  "stock_lot_movement.col.qty": "common_qty",
  "sales_summary.title": "sales",
  "sales_summary.description": "rpt_summary_of_gross_sales_discounts",
  "sales_summary.col.docdate": "date",
  "sales_summary.col.docno": "docno",
  "sales_summary.col.custname": "customer_name",
  "sales_summary.col.subtotal": "total_before_vat",
  "sales_summary.col.vatamount": "rpt_vat_7",
  "sales_summary.col.totalamount": "product_amount",
  "sales_daily.title": "sales_daily",
  "sales_daily.description": "rpt_daily_sales_audit_report_by",
  "sales_daily.col.salesdate": "rpt_date",
  "sales_daily.col.doccount": "rpt_bills",
  "sales_daily.col.cashsales": "cash",
  "sales_daily.col.transfersales": "money_transfer",
  "sales_daily.col.creditsales": "rpt_credit",
  "sales_daily.col.totalamount": "total_sales",
  "sales_by_seller.title": "sales_by_seller",
  "sales_by_seller.description": "rpt_sales_performance_ranking_and_commission",
  "sales_by_seller.col.sellername": "salename",
  "sales_by_seller.col.target": "rpt_target",
  "sales_by_seller.col.actual": "rpt_actual",
  "sales_by_seller.col.achievedpercent": "rpt_achieved",
  "sales_by_seller.col.commission": "rpt_commission",
  "sales_by_document.title": "sales_by_document",
  "sales_by_document.description": "rpt_sales_register_ordered_sequentially_by",
  "sales_by_document.col.docno": "rpt_doc_no",
  "sales_by_document.col.docdate": "date",
  "sales_by_document.col.custcode": "customer_code",
  "sales_by_document.col.custname": "customer_name",
  "sales_by_document.col.totalamount": "gl_amount_2",
  "gross_profit_document.title": "report_gross_profit_by_doc",
  "gross_profit_document.description": "rpt_gross_profit_margin_analysis_per",
  "gross_profit_document.col.docno": "rpt_doc_no",
  "gross_profit_document.col.custname": "customer",
  "gross_profit_document.col.salesrevenue": "sales_amount",
  "gross_profit_document.col.costofgoods": "cost_of_sales",
  "gross_profit_document.col.grossprofit": "gross_profit",
  "gross_profit_document.col.marginpercent": "rpt_margin",
  "gross_profit_product.title": "report_gross_profit_by_product",
  "gross_profit_product.description": "rpt_profitability_matrix_per_product_item",
  "gross_profit_product.col.itemcode": "import_productdetail.product_code",
  "gross_profit_product.col.itemname": "import_productdetail.product_name",
  "gross_profit_product.col.soldqty": "rpt_sold_qty",
  "gross_profit_product.col.salesrevenue": "total_sales",
  "gross_profit_product.col.costofgoods": "totalcost",
  "gross_profit_product.col.grossprofit": "gross_profit",
  "gross_profit_product.col.marginpercent": "rpt_margin",
  "sales_return_report.title": "sale_return_report",
  "sales_return_report.description": "rpt_customer_returns_and_credit_note",
  "sales_return_report.col.docdate": "date",
  "sales_return_report.col.docno": "rpt_cn_no",
  "sales_return_report.col.refdocno": "rpt_ref_invoice",
  "sales_return_report.col.custname": "customer",
  "sales_return_report.col.reason": "rpt_reason",
  "sales_return_report.col.totalamount": "rpt_amount",
  "sales_by_customer.title": "sales_by_customer",
  "sales_by_customer.description": "rpt_customer_revenue_breakdown_and_vip",
  "sales_by_customer.col.custcode": "customer_code",
  "sales_by_customer.col.custname": "customer_name",
  "sales_by_customer.col.billcount": "rpt_bills",
  "sales_by_customer.col.totalamount": "rpt_total_purchases",
  "sales_by_channel.title": "sales_by_channel",
  "sales_by_channel.description": "rpt_omnichannel_sales_distribution_analysis",
  "sales_by_channel.col.channel": "rpt_channel",
  "sales_by_channel.col.billcount": "rpt_orders",
  "sales_by_channel.col.totalamount": "total_sales",
  "sales_by_channel.col.sharepercent": "rpt_share",
  "purchase_summary.title": "purchase_report",
  "purchase_summary.description": "rpt_summary_of_vendor_purchases_and",
  "purchase_summary.col.docdate": "date",
  "purchase_summary.col.docno": "rpt_po_bill_no",
  "purchase_summary.col.vendorname": "report_group_creditor",
  "purchase_summary.col.subtotal": "total_before_vat",
  "purchase_summary.col.vatamount": "rpt_vat_7_2",
  "purchase_summary.col.totalamount": "rpt_total",
  "purchase_by_product.title": "purchase_by_product",
  "purchase_by_product.description": "rpt_product_purchase_volume_and_price",
  "purchase_by_product.col.itemcode": "import_productdetail.product_code",
  "purchase_by_product.col.itemname": "import_productdetail.product_name",
  "purchase_by_product.col.purchasedqty": "rpt_purchased_qty",
  "purchase_by_product.col.avgprice": "avg_price",
  "purchase_by_product.col.totalamount": "rpt_total_purchases",
  "purchase_partial_report.title": "gradual_receipt",
  "purchase_partial_report.description": "rpt_outstanding_purchase_orders_and_partial",
  "purchase_partial_report.col.pono": "ref_po_docno",
  "purchase_partial_report.col.vendorname": "report_group_creditor",
  "purchase_partial_report.col.orderqty": "rpt_order_qty",
  "purchase_partial_report.col.receivedqty": "rpt_received",
  "purchase_partial_report.col.pendingqty": "rpt_pending",
  "expense_summary.title": "expense_summary_report",
  "expense_summary.description": "rpt_operating_expense_breakdown_by_chart",
  "expense_summary.col.accountcode": "account_code",
  "expense_summary.col.accountname": "rpt_expense_category",
  "expense_summary.col.deptname": "company_branch_department",
  "expense_summary.col.totalamount": "ops_total",
  "ar_aging.title": "ar_aging",
  "ar_aging.description": "rpt_debtor_aging_schedule_with_standard",
  "ar_aging.col.custcode": "debtor_code",
  "ar_aging.col.custname": "debtor_name",
  "ar_aging.col.currentdue": "rpt_current",
  "ar_aging.col.days1_30": "rpt_1_30_days",
  "ar_aging.col.days31_60": "rpt_31_60_days",
  "ar_aging.col.days61_90": "rpt_61_90_days",
  "ar_aging.col.over90": "rpt_90_days",
  "ar_aging.col.totaldue": "rpt_total_due",
  "ar_payment_daily.title": "payment_daily",
  "ar_payment_daily.description": "rpt_daily_debtor_settlement_receipts_and",
  "ar_payment_daily.col.docdate": "rpt_date_2",
  "ar_payment_daily.col.docno": "rpt_receipt_no",
  "ar_payment_daily.col.custname": "report_group_debtor",
  "ar_payment_daily.col.paymethod": "rpt_method",
  "ar_payment_daily.col.totalamount": "rpt_amount_2",
  "ap_aging.title": "ap_aging",
  "ap_aging.description": "rpt_creditor_aging_schedule_to_manage",
  "ap_aging.col.vendorcode": "creditor_code",
  "ap_aging.col.vendorname": "creditor_name",
  "ap_aging.col.currentdue": "rpt_current",
  "ap_aging.col.days1_30": "rpt_1_30_days",
  "ap_aging.col.days31_60": "rpt_31_60_days",
  "ap_aging.col.days61_90": "rpt_61_90_days",
  "ap_aging.col.over90": "rpt_90_days",
  "ap_aging.col.totaldue": "rpt_total_due",
  "dbd_xbrl_export.title": "xbrl_export",
  "dbd_xbrl_export.description": "rpt_export_statutory_financial_statements_compliant",
  "dbd_xbrl_export.col.accountcode": "chart_of_acc",
  "dbd_xbrl_export.col.xbrltag": "rpt_xbrl_tag",
  "dbd_xbrl_export.col.description": "rpt_statement_line",
  "dbd_xbrl_export.col.amount": "rpt_current_period",
  "dbd_xbrl_export.col.prevamount": "rpt_prior_period",
};

export function reportText(
  config: ErpReportConfig,
  part: "title" | "description",
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  return catalogText(catalogKeys, `${config.code}.${part}`, config[part], language, dictionary);
}

export function reportColumnLabel(
  config: ErpReportConfig,
  column: ReportColumn,
  language: LanguageCode,
  dictionary?: BackendLanguageDictionary,
): string {
  return catalogText(
    catalogKeys,
    `${config.code}.col.${column.key}`,
    column.label,
    language,
    dictionary,
  );
}

// Unified ERP Reporting Engine for Thai SMEs & Thai Accounting
// Covers Inventory, Sales, Purchase, AR/AP Aging, Gross Profit, and DBD XBRL Export

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
    title: { th: "รายงานสินค้าคงเหลือตามสินค้า", en: "Stock Balance by Item" },
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
    title: { th: "รายงานสินค้าคงเหลือตามคลัง", en: "Stock Balance by Warehouse" },
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
    title: { th: "รายงานสินค้าคงเหลือตามโซนเก็บ", en: "Stock Balance by Location/Bin" },
    description: { th: "แสดงตำแหน่งจัดเก็บ แถว ชั้น ช่อง (Bin Location) ของสินค้า", en: "Stock balance pinpointed by location, aisle and bin" },
    defaultSortKey: "location",
    columns: [
      { key: "location", label: { th: "โซน/ช่องจัดเก็บ", en: "Location/Bin" } },
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "qty", label: { th: "จำนวน", en: "Qty" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/reportdedebistockbalance",
    code: "general_stock_balance",
    category: "inventory",
    title: { th: "รายงานสินค้าคงเหลือรวม", en: "Comprehensive Stock Balance" },
    description: { th: "รายงานสินค้าคงเหลือภาพรวมทั้งบริษัทพร้อมยอดจองและยอดพร้อมขาย", en: "Company-wide stock balance with reservations and available to sell" },
    defaultSortKey: "itemcode",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "qty", label: { th: "คงเหลือจริง", en: "Physical Qty" }, align: "right", isNumeric: true },
      { key: "reservedqty", label: { th: "ยอดสั่งจอง", en: "Reserved" }, align: "right", isNumeric: true },
      { key: "availableqty", label: { th: "พร้อมขาย (ATP)", en: "Available" }, align: "right", isNumeric: true },
      { key: "totalcost", label: { th: "มูลค่ารวม", en: "Total Cost" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/lowstock",
    code: "low_stock_report",
    category: "inventory",
    title: { th: "รายงานสินค้าใกล้หมดขั้นต่ำ (Reorder Point)", en: "Low Stock Alert" },
    description: { th: "เตือนรายการสินค้าที่ยอดคงเหลือต่ำกว่าจุดสั่งซื้อซ้ำ เพื่อความต่อเนื่องของธุรกิจ", en: "Items below minimum stock level and safety threshold" },
    defaultSortKey: "shortage",
    columns: [
      { key: "itemcode", label: { th: "รหัสสินค้า", en: "Item Code" } },
      { key: "itemname", label: { th: "ชื่อสินค้า", en: "Item Name" } },
      { key: "minqty", label: { th: "จุดสั่งซื้อขั้นต่ำ", en: "Min Point" }, align: "right", isNumeric: true },
      { key: "qty", label: { th: "คงเหลือปัจจุบัน", en: "Current Qty" }, align: "right", isNumeric: true },
      { key: "shortage", label: { th: "จำนวนที่ขาด", en: "Shortage" }, align: "right", isNumeric: true },
      { key: "reorderqty", label: { th: "แนะนำสั่งซื้อ", en: "Suggested Reorder" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/expiringstock",
    code: "expiring_stock_report",
    category: "inventory",
    title: { th: "รายงานสินค้าใกล้หมดอายุ", en: "Expiring Stock Report" },
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
    title: { th: "รายงานความเคลื่อนไหวสินค้าพร้อมต้นทุน (Stock Card)", en: "Stock Card with Cost Movement" },
    description: { th: "บัตรคุมสต็อกแสดงการรับ-จ่าย-คงเหลือ พร้อมต้นทุน FIFO และถัวเฉลี่ยถ่วงน้ำหนัก", en: "Stock card ledger with in-out-balance quantities and moving costs" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่เอกสาร", en: "Doc No." } },
      { key: "transname", label: { th: "ประเภทรายการ", en: "Transaction" } },
      { key: "inqty", label: { th: "รับเข้า", en: "In Qty" }, align: "right", isNumeric: true },
      { key: "outqty", label: { th: "จ่ายออก", en: "Out Qty" }, align: "right", isNumeric: true },
      { key: "balanceqty", label: { th: "คงเหลือ", en: "Balance Qty" }, align: "right", isNumeric: true },
      { key: "unitcost", label: { th: "ต้นทุน/หน่วย", en: "Unit Cost" }, align: "right", isCurrency: true },
      { key: "totalcost", label: { th: "มูลค่าคงเหลือ", en: "Balance Value" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportstockmovement",
    code: "stock_movement_simple",
    category: "inventory",
    title: { th: "รายงานความเคลื่อนไหวสินค้า", en: "Stock Movement Report" },
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
    title: { th: "รายงานความเคลื่อนไหวล็อตสินค้า", en: "Lot Tracking Movement" },
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
    title: { th: "รายงานสรุปยอดขายสินค้า", en: "Sales Summary Report" },
    description: { th: "สรุปยอดขายแยกตามหมวดหมู่ ยอดขายก่อนภาษี ภาษี และยอดสุทธิ", en: "Summary of gross sales, discounts, VAT and net sales" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่เอกสาร", en: "Invoice No." } },
      { key: "custname", label: { th: "ชื่อลูกค้า", en: "Customer" } },
      { key: "subtotal", label: { th: "มูลค่าก่อนภาษี", en: "Subtotal" }, align: "right", isCurrency: true },
      { key: "vatamount", label: { th: "ภาษีขาย 7%", en: "VAT 7%" }, align: "right", isCurrency: true },
      { key: "totalamount", label: { th: "ยอดรวมทั้งสิ้น", en: "Total" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/reportdedebisalesdaily",
    code: "sales_daily",
    category: "sales",
    title: { th: "รายงานยอดขายรายวัน", en: "Daily Sales Report" },
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
    title: { th: "รายงานยอดขายตามพนักงานขาย", en: "Sales by Salesperson" },
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
    title: { th: "รายงานกำไรขั้นต้นตามเอกสาร", en: "Gross Profit by Document" },
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
    title: { th: "รายงานกำไรขั้นต้นตามสินค้า", en: "Gross Profit by Product" },
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
    title: { th: "รายงานการรับคืนสินค้าจากการขาย", en: "Sales Return Report" },
    description: { th: "วิเคราะห์สาเหตุการคืนสินค้าและลดหนี้ให้ลูกค้า", en: "Customer returns and credit note issue log" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่ใบลดหนี้", en: "CN No." } },
      { key: "refdocno", label: { th: "อ้างอิงบิลเดิม", en: "Ref Invoice" } },
      { key: "custname", label: { th: "ลูกค้า", en: "Customer" } },
      { key: "reason", label: { th: "สาเหตุการคืน", en: "Reason" } },
      { key: "totalamount", label: { th: "ยอดลดหนี้", en: "Amount" }, align: "right", isCurrency: true },
    ],
  },
  {
    route: "/report/salesbycustomer",
    code: "sales_by_customer",
    category: "sales",
    title: { th: "รายงานยอดขายตามลูกค้า", en: "Sales by Customer" },
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
    title: { th: "รายงานยอดขายตามช่องทางจำหน่าย", en: "Sales by Channel" },
    description: { th: "เปรียบเทียบยอดขายหน้าร้าน, Shopee, Lazada, TikTok, และตัวแทนจำหน่าย", en: "Omnichannel sales distribution analysis" },
    defaultSortKey: "totalamount",
    columns: [
      { key: "channel", label: { th: "ช่องทางจำหน่าย", en: "Channel" } },
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
    title: { th: "รายงานการซื้อสินค้า", en: "Purchase Summary Report" },
    description: { th: "สรุปยอดการสั่งซื้อสินค้าและบริการแยกตามผู้จำหน่าย", en: "Summary of vendor purchases and expenses" },
    defaultSortKey: "docdate",
    columns: [
      { key: "docdate", label: { th: "วันที่", en: "Date" }, align: "center" },
      { key: "docno", label: { th: "เลขที่ใบซื้อ", en: "PO/Bill No." } },
      { key: "vendorname", label: { th: "ผู้จำหน่าย", en: "Vendor" } },
      { key: "subtotal", label: { th: "มูลค่าก่อนภาษี", en: "Subtotal" }, align: "right", isCurrency: true },
      { key: "vatamount", label: { th: "ภาษีซื้อ 7%", en: "VAT 7%" }, align: "right", isCurrency: true },
      { key: "totalamount", label: { th: "ยอดรวมสุทธิ", en: "Total" }, align: "right", isCurrency: true },
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
    title: { th: "รายงานการทยอยรับสินค้าตามใบสั่งซื้อ", en: "Partial Delivery Tracking" },
    description: { th: "ติดตามสินค้าค้างรับจากผู้ขายตามใบสั่งซื้อ (PO Backlog)", en: "Outstanding purchase orders and partial shipments status" },
    defaultSortKey: "pono",
    columns: [
      { key: "pono", label: { th: "เลขที่ PO", en: "PO No." } },
      { key: "vendorname", label: { th: "ผู้จำหน่าย", en: "Vendor" } },
      { key: "orderqty", label: { th: "จำนวนสั่งซื้อ", en: "Order Qty" }, align: "right", isNumeric: true },
      { key: "receivedqty", label: { th: "รับเข้าแล้ว", en: "Received" }, align: "right", isNumeric: true },
      { key: "pendingqty", label: { th: "ค้างส่งมอบ", en: "Pending" }, align: "right", isNumeric: true },
    ],
  },
  {
    route: "/report/expensesummary",
    code: "expense_summary",
    category: "purchase",
    title: { th: "รายงานสรุปค่าใช้จ่ายตามหมวดหมู่", en: "Expense Summary Report" },
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
    title: { th: "รายงานวิเคราะห์อายุลูกหนี้ (AR Aging)", en: "Accounts Receivable Aging Report" },
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
    title: { th: "รายงานรับชำระเงินรายวันจากลูกหนี้", en: "Daily Debt Collection Register" },
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
    title: { th: "รายงานวิเคราะห์อายุเจ้าหนี้ (AP Aging)", en: "Accounts Payable Aging Report" },
    description: { th: "วิเคราะห์หนี้ที่ต้องจ่ายให้เจ้าหนี้ตามระยะเวลาครบกำหนดชำระเพื่อบริหารสภาพคล่อง", en: "Creditor aging schedule to manage vendor payments and liquidity" },
    defaultSortKey: "totaldue",
    columns: [
      { key: "vendorcode", label: { th: "รหัสเจ้าหนี้", en: "Code" } },
      { key: "vendorname", label: { th: "ชื่อเจ้าหนี้ / ผู้จำหน่าย", en: "Vendor / Creditor" } },
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
    title: { th: "ส่งออกงบการเงินรูปแบบ XBRL (ยื่น DBD e-Filing)", en: "DBD XBRL Financial Statement Export" },
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

export function getSampleReportData(code: string): Record<string, unknown>[] {
  if (code.includes("stock_balance")) {
    return [
      { itemcode: "P-001", itemname: "ปากกาลูกลื่น 0.5 มม.", unitname: "ด้าม", qty: 1500, avgcost: 12.5, totalcost: 18750, warehouse: "คลังหลัก (WH-01)", location: "A-01-01", reservedqty: 200, availableqty: 1300 },
      { itemcode: "P-002", itemname: "สมุดบันทึกริมลวด A5", unitname: "เล่ม", qty: 820, avgcost: 45.0, totalcost: 36900, warehouse: "คลังหลัก (WH-01)", location: "A-02-04", reservedqty: 50, availableqty: 770 },
      { itemcode: "P-003", itemname: "กระดาษถ่ายเอกสาร A4 80 แกรม", unitname: "รีม", qty: 350, avgcost: 110.0, totalcost: 38500, warehouse: "คลังสาขา 1 (WH-02)", location: "B-01-02", reservedqty: 100, availableqty: 250 },
      { itemcode: "P-004", itemname: "แฟ้มห่วง 2 นิ้ว ตราช้าง", unitname: "เล่ม", qty: 450, avgcost: 65.0, totalcost: 29250, warehouse: "คลังหลัก (WH-01)", location: "B-03-01", reservedqty: 0, availableqty: 450 },
    ];
  }

  if (code === "low_stock_report") {
    return [
      { itemcode: "P-003", itemname: "กระดาษถ่ายเอกสาร A4 80 แกรม", minqty: 500, qty: 350, shortage: 150, reorderqty: 300 },
      { itemcode: "P-005", itemname: "หมึกพิมพ์เลเซอร์ HP LaserJet", minqty: 50, qty: 12, shortage: 38, reorderqty: 50 },
      { itemcode: "P-008", itemname: "เทปใสปิดกล่อง 2 นิ้ว", minqty: 200, qty: 45, shortage: 155, reorderqty: 200 },
    ];
  }

  if (code === "expiring_stock_report") {
    return [
      { itemcode: "MED-01", itemname: "น้ำยาทำความสะอาดฆ่าเชื้อ 5 ลิตร", lotno: "LOT-202512-01", expirydate: "2026-10-15", daysremaining: 30, qty: 45 },
      { itemcode: "FOOD-09", itemname: "กาแฟคั่วบดดอยช้าง 500g", lotno: "LOT-202601-88", expirydate: "2026-11-20", daysremaining: 66, qty: 120 },
    ];
  }

  if (code.includes("sales") || code.includes("gross_profit")) {
    return [
      {
        docdate: "2026-09-02", docno: "INV-202609-001", custcode: "C-001", custname: "บริษัท สยามการค้าปลีก จำกัด (มหาชน)",
        subtotal: 50000, vatamount: 3500, totalamount: 53500, salesdate: "2026-09-02", doccount: 12, cashsales: 15000, transfersales: 25000, creditsales: 13500,
        sellername: "สมเกียรติ มั่นคง", target: 500000, actual: 580000, achievedpercent: 116, commission: 17400,
        salesrevenue: 50000, costofgoods: 32000, grossprofit: 18000, marginpercent: 36.0, itemcode: "P-001", itemname: "ปากกาลูกลื่น 0.5 มม.", soldqty: 4000,
        channel: "ขายหน้าร้าน (Store POS)", billcount: 120, sharepercent: 45.5,
      },
      {
        docdate: "2026-09-05", docno: "INV-202609-002", custcode: "C-002", custname: "ห้างหุ้นส่วนจำกัด ไทยเจริญการช่าง",
        subtotal: 28000, vatamount: 1960, totalamount: 29960, salesdate: "2026-09-05", doccount: 8, cashsales: 8000, transfersales: 21960, creditsales: 0,
        sellername: "นภาลัย สดใส", target: 400000, actual: 420000, achievedpercent: 105, commission: 12600,
        salesrevenue: 28000, costofgoods: 16500, grossprofit: 11500, marginpercent: 41.1, itemcode: "P-002", itemname: "สมุดบันทึกริมลวด A5", soldqty: 620,
        channel: "Shopee Mall", billcount: 85, sharepercent: 32.2,
      },
    ];
  }

  if (code.includes("purchase") || code.includes("expense")) {
    return [
      {
        docdate: "2026-09-03", docno: "PUR-202609-01", vendorcode: "V-001", vendorname: "บริษัท สเตชั่นเนอรี่ ซัพพลาย จำกัด",
        subtotal: 45000, vatamount: 3150, totalamount: 48150, pono: "PO-202608-099", orderqty: 500, receivedqty: 350, pendingqty: 150,
        itemcode: "P-003", itemname: "กระดาษถ่ายเอกสาร A4 80 แกรม", purchasedqty: 350, avgprice: 110.0,
        accountcode: "510101", accountname: "ค่าเครื่องเขียนและวัสดุสำนักงาน", deptname: "ฝ่ายบริหารทั่วไป",
      },
      {
        docdate: "2026-09-07", docno: "PUR-202609-02", vendorcode: "V-002", vendorname: "บริษัท โลจิสติกส์ พลัส จำกัด",
        subtotal: 12000, vatamount: 840, totalamount: 12840, pono: "PO-202609-012", orderqty: 1, receivedqty: 1, pendingqty: 0,
        itemcode: "SRV-01", itemname: "ค่าขนส่งสินค้าเข้าคลัง", purchasedqty: 1, avgprice: 12000.0,
        accountcode: "510204", accountname: "ค่าระวางและขนส่งเข้า", deptname: "ฝ่ายคลังสินค้า",
      },
    ];
  }

  if (code === "ar_aging") {
    return [
      { custcode: "AR-001", custname: "บริษัท ซุปเปอร์ริช เทรดดิ้ง จำกัด", currentdue: 150000, days1_30: 45000, days31_60: 0, days61_90: 0, over90: 0, totaldue: 195000 },
      { custcode: "AR-002", custname: "ห้างหุ้นส่วนจำกัด โชคชัยวิศวกรรม", currentdue: 80000, days1_30: 30000, days31_60: 25000, days61_90: 0, over90: 0, totaldue: 135000 },
      { custcode: "AR-003", custname: "บริษัท สยามโมเดิร์น เฟอร์นิเจอร์ จำกัด", currentdue: 0, days1_30: 0, days31_60: 45000, days61_90: 18000, over90: 12000, totaldue: 75000 },
    ];
  }

  if (code === "ap_aging") {
    return [
      { vendorcode: "AP-001", vendorname: "บริษัท ปิโตรเคมีคอล กรุ๊ป จำกัด", currentdue: 320000, days1_30: 0, days31_60: 0, days61_90: 0, over90: 0, totaldue: 320000 },
      { vendorcode: "AP-002", vendorname: "บริษัท บางกอกเปเปอร์มิลล์ จำกัด", currentdue: 110000, days1_30: 45000, days31_60: 0, days61_90: 0, over90: 0, totaldue: 155000 },
      { vendorcode: "AP-003", vendorname: "ห้างหุ้นส่วนจำกัด นครหลวงขนส่ง", currentdue: 25000, days1_30: 12000, days31_60: 0, days61_90: 0, over90: 0, totaldue: 37000 },
    ];
  }

  if (code === "dbd_xbrl_export") {
    return [
      { accountcode: "111101", xbrltag: "th-gaap-ci:CashAndCashEquivalents", description: "เงินสดและรายการเทียบเท่าเงินสด", amount: 1450280.50, prevamount: 1120450.00 },
      { accountcode: "111201", xbrltag: "th-gaap-ci:TradeAndOtherCurrentReceivables", description: "ลูกหนี้การค้าและลูกหนี้หมุนเวียนอื่น", amount: 890400.00, prevamount: 760000.00 },
      { accountcode: "111301", xbrltag: "th-gaap-ci:Inventories", description: "สินค้าคงเหลือ", amount: 1280500.00, prevamount: 950000.00 },
      { accountcode: "121101", xbrltag: "th-gaap-ci:PropertyPlantAndEquipment", description: "ที่ดิน อาคารและอุปกรณ์ (สุทธิ)", amount: 3500000.00, prevamount: 3800000.00 },
      { accountcode: "211101", xbrltag: "th-gaap-ci:TradeAndOtherCurrentPayables", description: "เจ้าหนี้การค้าและเจ้าหนี้หมุนเวียนอื่น", amount: 512000.00, prevamount: 480000.00 },
      { accountcode: "311101", xbrltag: "th-gaap-ci:AuthorizedShareCapital", description: "ทุนจดทะเบียนชำระแล้ว", amount: 5000000.00, prevamount: 5000000.00 },
      { accountcode: "311201", xbrltag: "th-gaap-ci:RetainedEarningsUnappropriated", description: "กำไรสะสมยังไม่ได้จัดสรร", amount: 1609180.50, prevamount: 1150450.00 },
    ];
  }

  return [];
}

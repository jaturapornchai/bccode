import type { LanguageCode } from "./i18n";
import { isBackendLanguageReady, type BackendLanguageDictionary } from "./backend-language";

export type MenuCategory = "transaction" | "report" | "master" | "settings" | "finance" | "restaurant" | "approval";

export type MenuLabel = {
  key?: string;
  th: string;
  en: string;
} & Partial<Record<LanguageCode, string>>;

export type MenuItem = {
  id: string;
  label: MenuLabel;
  route: string;
  category: MenuCategory;
};

export type MenuGroup = {
  id: string;
  title: MenuLabel;
  items: MenuItem[];
};

export type MenuSection = {
  id: string;
  title: MenuLabel;
  groups: MenuGroup[];
};

const menuKey = (id: string) => id.replace(/-/g, "_");
const ml = (id: string, th: string, en: string): MenuLabel => ({ key: menuKey(id), th, en });

const tx = (
  id: string,
  th: string,
  en: string,
  route: string,
  category: MenuCategory = "transaction",
  languageKey: string = menuKey(id),
): MenuItem => ({
  id,
  label: { key: languageKey, th, en },
  route,
  category,
});

export const MENU_SECTIONS: MenuSection[] = [
  {
    id: "transactions",
    title: { key: "operations", th: "งานประจำ", en: "Operations" },
    groups: [
      {
        id: "procurement",
        title: ml("procurement", "จัดซื้อจัดหา", "Procurement"),
        items: [
          tx("procurement-dashboard", "ภาพรวมจัดซื้อ", "Purchase Overview", "/procurement/dashboard"),
          tx("purchase-requisition", "ใบขอซื้อ", "Purchase Requisition", "/transaction/purchaserequisition"),
          tx("rfq", "สืบราคาและเจรจา", "Price Inquiry & Negotiation", "/transaction/rfq"),
          tx("purchase-order", "ใบสั่งซื้อ", "Purchase Order", "/transaction/purchaseorder"),
        ],
      },
      {
        id: "purchase",
        title: ml("purchase_group", "ซื้อ", "Purchase"),
        items: [
          tx("purchase", "ซื้อสินค้า", "Purchase", "/transaction/purchase"),
          tx("purchase-return", "คืนซื้อ", "Purchase Return", "/transaction/purchasereturn"),
          tx("purchase-debit-note", "ใบลดหนี้เจ้าหนี้", "Purchase Debit Note", "/transaction/purchasedebitnote"),
          tx("expense-record", "บันทึกค่าใช้จ่าย", "Expense Record", "/transaction/expense"),
          tx("document-vault", "คลังเอกสารและสแกนบิล", "Document Vault & Scan", "/transaction/documentvault"),
          tx("withholding-tax-deduction", "ภาษีหัก ณ ที่จ่าย", "Withholding Tax", "/transaction/withholdingtax", "finance"),
          tx("purchase-partial", "รับสินค้าแบบทยอยรับ", "Gradual Receipt", "/transaction/purchasepartial"),
          tx("accrual-receive", "ตั้งหนี้จากทยอยรับ", "Set Debt from Receipt", "/transaction/accrualreceive"),
        ],
      },
      {
        id: "purchase-payment",
        title: ml("purchase_advance_deposit", "เงินล่วงหน้า/เงินมัดจำฝั่งซื้อ", "Purchase Advance / Deposit"),
        items: [
          tx("advance-payment", "จ่ายเงินล่วงหน้า", "Advance Payment", "/transaction/advancepayment", "finance"),
          tx("advance-payment-refund", "รับคืนเงินล่วงหน้า", "Advance Refund", "/transaction/advancepaymentrefund", "finance"),
          tx("deposit", "จ่ายเงินมัดจำ", "Pay Deposit", "/transaction/deposit", "finance"),
          tx("deposit-refund", "รับคืนเงินมัดจำ", "Deposit Refund", "/transaction/depositrefund", "finance"),
        ],
      },
      {
        id: "sales",
        title: ml("sales_group", "ขาย", "Sales"),
        items: [
          tx("quotation", "ใบเสนอราคา", "Quotation", "/transaction/quotation"),
          tx("sale-order", "ใบสั่งขาย", "Sale Order", "/transaction/saleorder"),
          tx("billing-note", "ใบวางบิล", "Billing Note", "/transaction/billingnote"),
          tx("sale-invoice", "ใบแจ้งหนี้", "Sale Invoice", "/transaction/saleinvoice"),
          tx("tax-invoice", "ใบเสร็จรับเงิน/ใบกำกับภาษี", "Tax Invoice / Receipt", "/transaction/taxinvoice"),
          tx("credit-note", "ใบลดหนี้", "Credit Note", "/transaction/creditnote"),
          tx("debit-note", "ใบเพิ่มหนี้", "Debit Note", "/transaction/debitnote"),
          tx("etax-invoice", "ใบกำกับภาษีอิเล็กทรอนิกส์", "e-Tax Invoice & e-Receipt", "/transaction/etax"),
          tx("receive-advance", "รับเงินล่วงหน้า", "Receive Advance", "/transaction/paidadvance", "finance"),
          tx("return-advance", "คืนเงินล่วงหน้า", "Return Advance", "/transaction/paidadvancerefund", "finance"),
          tx("receive-deposit", "รับเงินมัดจำ", "Receive Deposit", "/transaction/receivedeposit", "finance"),
          tx("return-deposit", "คืนเงินมัดจำ", "Return Deposit", "/transaction/receivedepositrefund", "finance"),
          tx("sale", "ขายสินค้า", "Sale", "/transaction/sale"),
          tx("sale-return", "คืนขาย", "Sale Return", "/transaction/salereturn"),
        ],
      },
      {
        id: "stock",
        title: ml("stock_group", "คลังสินค้า", "Stock"),
        items: [
          tx("stock-transfer", "โอนสินค้า", "Stock Transfer", "/transaction/stocktransfer"),
          tx("stock-receive", "รับสินค้า", "Stock Receive", "/transaction/stockreceiveproduct"),
          tx("stock-pickup", "เบิกสินค้า", "Stock Pickup", "/transaction/stockpickupproduct"),
          tx("stock-return", "คืนสินค้าเข้าคลัง", "Stock Return", "/transaction/stockreturnproduct"),
          tx("stock-adjust", "ปรับปรุงสต็อก", "Stock Adjustment", "/transaction/adjust"),
          tx("stock-count", "ตรวจนับสต็อก", "Stock Count", "/transaction/stockcount"),
          tx("fifo-cost-layers", "ชั้นต้นทุนสต็อกเข้าก่อนออกก่อน", "FIFO Cost Layers", "/inventory/cost-layers"),
          tx("stock-balance", "ยอดยกมาสินค้า", "Stock Balance", "/transaction/stockbalance"),
        ],
      },
      {
        id: "payment",
        title: ml("receive_pay_money", "รับ/จ่ายเงิน", "Payment"),
        items: [
          tx("paid", "รับชำระ", "Receive Payment", "/transaction/paid", "finance"),
          tx("pay", "จ่ายชำระ", "Pay", "/transaction/pay", "finance"),
          tx("bank-statement", "รายการเดินบัญชีธนาคาร", "Bank Statement", "/banking/statements", "finance"),
          tx("bank-reconcile", "กระทบยอดเงินฝากธนาคาร", "Bank Reconciliation", "/banking/reconciliation", "finance"),
          tx("slip-in", "รูปสลิปเงินเข้า", "Money In Slip", "/slipmoneyin", "finance"),
          tx("slip-out", "รูปสลิปเงินออก", "Money Out Slip", "/slipmoneyout", "finance"),
        ],
      },
      {
        id: "accounting",
        title: ml("accounting", "บัญชี", "Accounting"),
        items: [
          tx("uv-journal", "สมุดรายวันขาย", "Sales Journal (UV)", "/gl/journal/uv", "finance"),
          tx("sv-journal", "สมุดรายวันซื้อ", "Purchase Journal (SV)", "/gl/journal/sv", "finance"),
          tx("rv-journal", "สมุดรายวันรับเงิน", "Receipt Voucher Journal (RV)", "/gl/journal/rv", "finance"),
          tx("pv-journal", "สมุดรายวันจ่ายเงิน", "Payment Voucher Journal (PV)", "/gl/journal/pv", "finance"),
          tx("jv-journal", "สมุดรายวันทั่วไป", "General Journal (JV)", "/gl/journal/jv", "finance"),
          tx("chart-of-accounts", "ผังบัญชี", "Chart of Accounts", "/gl/chartofaccounts", "finance"),
          tx("working-paper", "กระดาษทำการ", "Working Paper", "/gl/workingpaper", "finance"),
          tx("period-lock", "ล็อกงวดบัญชี", "Period Lock", "/gl/periodlock", "finance"),
          tx("daily-info", "ตรวจสอบประจำวัน", "Daily Check", "/checkdaily/dailyinfoscreen", "finance"),
          tx("cash-drawer", "รับ-ส่งเงิน POS", "POS Cash Drawer", "/cashinginthedrawer", "finance"),
        ],
      },
    ],
  },
  {
    id: "reports",
    title: { key: "report", th: "รายงาน", en: "Reports" },
    groups: [
      {
        id: "financial-reports",
        title: ml("financial_reports", "รายงานการเงิน", "Financial Reports"),
        items: [
          tx("profit-loss", "งบกำไรขาดทุน", "Profit and Loss", "/report/pnl", "report"),
          tx("balance-sheet", "งบดุล", "Balance Sheet", "/report/balancesheet", "report"),
          tx("trial-balance", "งบทดลอง", "Trial Balance", "/report/trialbalance", "report"),
          tx("cash-flow", "งบกระแสเงินสด", "Cash Flow Statement", "/report/cashflow", "report"),
          tx("financial-graphs", "กราฟประกอบงบการเงิน", "Financial Statement Graphs", "/report/financialgraphs", "report"),
          tx("project-pnl", "กำไรขาดทุนตามโครงการ", "Project P&L", "/report/project-pnl", "report"),
        ],
      },
      {
        id: "tax-reports",
        title: ml("tax_reports", "รายงานภาษี", "Tax Reports"),
        items: [
          tx("vat-sale", "รายงานภาษีขาย", "VAT Sale", "/report/reportvatsale", "report"),
          tx("vat-buy", "รายงานภาษีซื้อ", "VAT Purchase", "/report/reportvatbuy", "report"),
          tx("vat-pp30", "แบบยื่นภาษี ภ.พ.30", "VAT Return (P.P.30)", "/report/vatpp30", "report"),
          tx("withholding-tax-report", "รายงานภาษีหัก ณ ที่จ่าย", "Withholding Tax Report", "/report/wht-reports", "report"),
        ],
      },
      {
        id: "inventory-reports",
        title: ml("inventory_reports", "รายงานสินค้า", "Inventory Reports"),
        items: [
          tx("stock-balance-item", "คงเหลือตามสินค้า", "Stock by Item", "/report/stockbalanceitem", "report"),
          tx("stock-balance-warehouse", "คงเหลือตามคลัง", "Stock by Warehouse", "/report/stockbalancewarehouse", "report"),
          tx("stock-balance-location", "คงเหลือตามโซนเก็บสินค้า", "Stock by Storage Zone", "/report/stockbalancelocation", "report"),
          tx("stock-movement-cost", "เคลื่อนไหวสินค้าพร้อมต้นทุน", "Stock Movement with Cost", "/report/stockmovementcost", "report"),
        ],
      },
      {
        id: "sales-reports",
        title: ml("sales_reports", "รายงานขาย", "Sales Reports"),
        items: [
          tx("gross-profit-doc", "กำไรขั้นต้นตามเอกสาร", "Gross Profit by Document", "/report/salesreportbydocument", "report"),
          tx("sales", "รายงานขาย", "Sales Report", "/report/reportdedebisales", "report"),
          tx("sales-daily", "รายงานขายรายวัน", "Daily Sales", "/report/reportdedebisalesdaily", "report"),
          tx("payment-daily", "รายงานรับชำระรายวัน", "Daily Payment", "/report/reportdedebipaymentdaily", "report"),
          tx("sale-return-report", "รายงานคืนขาย", "Sale Return Report", "/report/reportdedebisalereturn", "report"),
        ],
      },
      {
        id: "other-reports",
        title: ml("other_reports", "รายงานอื่น ๆ", "Other Reports"),
        items: [
          tx("product-movement", "ความเคลื่อนไหวสินค้า", "Product Movement", "/report/reportstockmovement", "report"),
          tx("inventory", "รายงานสินค้าคงเหลือ", "Inventory Report", "/report/reportdedebistockbalance", "report"),
          tx("gradual-receipt", "รายงานทยอยรับ", "Gradual Receipt Report", "/report/reportdedebipurchasepartial", "report"),
          tx("gross-profit-by-document", "กำไรขั้นต้นตามเอกสาร", "Gross Profit by Document", "/report/reportgrossprofitbydocument", "report"),
          tx("gross-profit-by-product", "กำไรขั้นต้นตามสินค้า", "Gross Profit by Product", "/report/reportgrossprofitbyproduct", "report"),
        ],
      },
      {
        id: "audit-reports",
        title: ml("audit_reprocess", "ตรวจและซ่อมข้อมูล", "Data Check and Repair"),
        items: [
          tx("reprocess", "คำนวณข้อมูลใหม่", "Recalculate Data", "/rebuildstockscreen", "report"),
          tx("audit-data", "ตรวจข้อมูล", "Check Data", "/auditscreen", "report"),
          tx("rebuild-products", "สร้างรายการสินค้าใหม่", "Rebuild Product List", "/rebuildproductsscreen", "report"),
          tx("rebuild-product-balance", "สร้างยอดคงเหลือใหม่", "Rebuild Stock Balance", "/rebuildproductbalancescreen", "report"),
        ],
      },
    ],
  },
  {
    id: "master",
    title: { key: "master_data", th: "ข้อมูลหลัก", en: "Master Data" },
    groups: [
      {
        id: "products",
        title: ml("product_catalog", "สินค้าและบาร์โค้ด", "Product Catalog"),
        items: [
          tx("product", "สินค้า", "Product", "/product", "master"),
          tx("service-product", "สินค้าบริการ", "Service Products", "/serviceproduct", "master"),
          tx("non-stock-product", "สินค้าไม่นับสต็อก", "Non-Stock Products", "/nonstockproduct", "master"),
          tx("product-extension", "ข้อมูลเสริมสินค้า", "Product Extended Data", "/productextension", "master"),
          { ...tx("barcode", "บาร์โค้ด", "Barcode", "/productbarcode", "master"), label: ml("barcode", "บาร์โค้ด", "Barcode") },
          tx("productset", "สินค้าชุด", "Product Set", "/productset", "master"),
          tx("bom", "สูตรผลิต", "Product BOM", "/productbom", "master"),
        ],
      },
      {
        id: "partners",
        title: ml("creditor_debtor", "คู่ค้า (ลูกค้า/ผู้ขาย)", "Trade Partners"),
        items: [
          tx("debtor", "ลูกหนี้", "Debtor", "/debtor", "master"),
          tx("creditor", "เจ้าหนี้", "Creditor", "/creditor", "master"),
          tx("debtor-beginning-balance", "ลูกหนี้ตั้งต้นรายเอกสาร", "Debtor Beginning Balance", "/debtorbeginningbalance", "master"),
          tx("creditor-beginning-balance", "เจ้าหนี้ตั้งต้นรายเอกสาร", "Creditor Beginning Balance", "/creditorbeginningbalance", "master"),
        ],
      },
      {
        id: "bank-accounts",
        title: ml("bookbank", "สมุดบัญชีธนาคาร", "Bank Accounts"),
        items: [
          tx("book-bank", "สมุดบัญชีธนาคาร", "Bank Book", "/bookbankscreen", "master"),
        ],
      },
      {
        id: "fixed-assets",
        title: ml("fixed_assets", "สินทรัพย์ถาวร", "Fixed Assets"),
        items: [
          tx("asset-registry", "ทะเบียนสินทรัพย์ถาวร", "Fixed Asset Registry", "/asset/registry", "master"),
          tx("asset-depreciation", "คำนวณค่าเสื่อมราคา", "Asset Depreciation", "/asset/depreciation", "master"),
        ],
      },
      {
        id: "dimensions",
        title: ml("dimensions", "มิติข้อมูล", "Dimensions"),
        items: [
          tx("project-department-tags", "แท็กโครงการและแผนก", "Project & Department Tags", "/dimensions/tags", "master"),
        ],
      },
      {
        id: "product-tools",
        title: ml("product_tools", "เครื่องมือสินค้า", "Product Tools"),
        items: [
          tx("product-serial-registry", "ทะเบียนเลขเครื่อง", "Serial Registry", "/productserialregistry", "master"),
          tx("price-history", "ประวัติแก้ไขราคา", "Price Edit History", "/pricehistory", "master"),
          tx("label-print", "พิมพ์ป้ายสินค้า", "Print Product Label", "/productbarcodeshelf", "master"),
          tx("add-product-kitchen", "เพิ่มสินค้าเข้าครัว", "Add Product to Kitchen", "/addproducttokitchenscreen", "restaurant"),
        ],
      },
      {
        id: "import-export",
        title: ml("import_export_data", "นำเข้า/ส่งออกข้อมูล", "Import / Export"),
        items: [
          tx("import-product", "นำเข้ารายการสินค้า", "Import Product List", "/importproduct", "master"),
          tx("import-product-file", "นำเข้าสินค้าจากไฟล์", "Import Product File", "/importproductfromfile", "master"),
          tx("import-product-image", "นำเข้ารูปสินค้า", "Import Product Image", "/importproductimage", "master"),
        ],
      },
      {
        id: "product-assistant",
        title: ml("product_assistant", "ผู้ช่วย AI และคลังความรู้", "AI Assistant & Knowledge Base"),
        items: [
          tx("knowledge-base", "คลังความรู้", "Knowledge Base", "/knowledgebasescreen", "master"),
          tx("alert-agent", "ผู้ช่วยแจ้งเตือน", "Alert Assistant", "/alertagentscreen", "master"),
          tx("api-dashboard", "ศูนย์การเชื่อมต่อ API", "API Integration Dashboard", "/apidashboard", "master"),
        ],
      },
    ],
  },
  {
    id: "defaults",
    title: { key: "menu_setup", th: "ค่าเริ่มต้น", en: "Defaults" },
    groups: [
      {
        id: "product-classification",
        title: ml("product_classification", "จัดกลุ่มสินค้า", "Product Grouping"),
        items: [
          tx("product-unit", "หน่วยนับสินค้า", "Product Unit", "/productunit", "master"),
          tx("product-group", "กลุ่มสินค้า", "Product Group", "/productgroup", "master"),
          tx("product-category", "จัดหมวดสินค้า", "Product Categories", "/productcategorygroupselectscreen", "master"),
          tx("product-category-list", "สินค้าในหมวด", "Products in Category", "/productcategorylist", "master"),
        ],
      },
      {
        id: "product-descriptors",
        title: ml("product_descriptors", "รายละเอียดประกอบสินค้า", "Product Details"),
        items: [
          tx("brand", "ยี่ห้อสินค้า", "Brand", "/masterbrandscreen", "master"),
          tx("model", "รุ่นสินค้า", "Model", "/mastermodelscreen", "master"),
          {
            ...tx("category", "คุณลักษณะสินค้า", "Product Attributes", "/mastercategoryscreen", "master"),
            label: ml("product_attribute_category", "คุณลักษณะสินค้า", "Product Attributes"),
          },
          tx("pattern", "รูปแบบสินค้า", "Pattern", "/masterpatternscreen", "master"),
          tx("dimension", "ขนาด/มิติสินค้า", "Product Dimensions", "/productdimension", "master"),
          tx("grade", "เกรดสินค้า", "Product Grade", "/mastergradescreen", "master"),
          tx("class", "ระดับสินค้า", "Product Class", "/masterclassscreen", "master"),
          tx("design", "รูปทรงสินค้า", "Product Shape", "/masterdesignscreen", "master"),
        ],
      },
      {
        id: "product-sku-options",
        title: ml("product_sku_options", "สี ไซซ์ และตัวเลือก", "Product Options"),
        items: [
          tx("product-color", "สีสินค้า", "Product Color", "/productcolor", "master"),
          tx("product-size", "ไซซ์/ขนาดสินค้า", "Product Size", "/productsize", "master"),
          tx("product-variant-matrix", "ชุดตัวเลือกสินค้า", "Product Option Sets", "/productvariantmatrix", "master"),
        ],
      },
      {
        id: "warehouse-setup",
        title: ml("warehouse", "คลังสินค้า", "Warehouse"),
        items: [
          tx("warehouse", "คลัง", "Warehouse", "/productwarehousescreen", "master"),
        ],
      },
      {
        id: "partner-groups",
        title: ml("partner_groups", "กลุ่มคู่ค้า", "Partner Groups"),
        items: [
          tx("debtor-group", "กลุ่มลูกหนี้", "Debtor Group", "/debtorgroup", "master"),
          tx("creditor-group", "กลุ่มเจ้าหนี้", "Creditor Group", "/creditorgroup", "master"),
        ],
      },
      {
        id: "sales-payment-banking",
        title: ml("sales_payment_banking", "การรับเงินและบัญชีธนาคาร", "Payment & Banking"),
        items: [
          tx("bank", "ธนาคาร", "Bank", "/bank", "master"),
          tx("qr-provider", "ผู้ให้บริการรับเงิน QR", "QR Payment Provider", "/qrprovider", "master"),
          tx("exchange-rate", "อัตราแลกเปลี่ยน", "Exchange Rate", "/exchangerate", "finance"),
          tx("bank-rules", "กฎจับคู่บัญชีอัตโนมัติ", "Bank Matching Rules", "/banking/rules", "master"),
        ],
      },
      {
        id: "sales-channel-pricing",
        title: ml("sales_channel_pricing", "ช่องทางขาย ราคา และขนส่ง", "Sales Channels, Pricing & Shipping"),
        items: [
          tx("sale-channel", "ช่องทางขาย", "Sale Channel", "/salechannelscreen", "master"),
          tx("transport-channel", "ช่องทางขนส่ง", "Transport Channel", "/transportchannelscreen", "master"),
          tx("channel-price", "ราคาตามช่องทางขาย", "Channel Prices", "/channelprice", "master"),
        ],
      },
      {
        id: "sales-documents",
        title: ml("sales_documents", "เอกสารและบิล", "Documents & Billing"),
        items: [
          tx("doc-format", "รูปแบบเอกสาร", "Document Format", "/docformat", "master"),
          tx("bill-design", "ออกแบบบิล", "Bill Design", "/billdesign", "master"),
          tx("etax-setting", "ตั้งค่าใบกำกับภาษีอิเล็กทรอนิกส์", "e-Tax Invoice Setting", "/etaxsetting", "master"),
          tx("tax-invoice-request-setting", "ขอใบกำกับภาษีออนไลน์", "Online Tax Invoice Request", "/taxinvoicerequestsetting", "master"),
        ],
      },
      {
        id: "approval",
        title: ml("approval", "อนุมัติ", "Approval"),
        items: [
          tx("purchase-type", "ประเภทซื้อ", "Purchase Type", "/purchasetypescreen", "approval"),
          tx("po-approval", "อนุมัติใบสั่งซื้อ", "PO Approval", "/poapprovalsettingscreen", "approval"),
          tx("quotation-type", "ประเภทใบเสนอราคา", "Quotation Type", "/quotationtypescreen", "approval"),
          tx("qt-approval", "อนุมัติใบเสนอราคา", "Quotation Approval", "/qtapprovalsettingscreen", "approval"),
          tx("sale-order-type", "ประเภทใบสั่งขาย", "Sale Order Type", "/saleordertypescreen", "approval"),
          tx("so-approval", "อนุมัติใบสั่งขาย", "Sale Order Approval", "/soapprovalsettingscreen", "approval"),
        ],
      },
      {
        id: "sales-pos",
        title: ml("sales_pos", "หน้าร้าน POS", "Point of Sale"),
        items: [
          tx("pos-setting", "ตั้งค่าเครื่องขายหน้าร้าน (POS)", "Point of Sale Settings", "/possetting", "master"),
          tx("pos-media", "รูป/สื่อหน้าจอขาย", "Point of Sale Media", "/posmedia", "master"),
          tx("thermal-printer-setting", "เครื่องพิมพ์ใบเสร็จเทอร์มัล", "Thermal Printer Settings", "/posprintersetting", "master"),
          { ...tx("color", "สีสำหรับงานขาย", "Sales Colors", "/colorscreen", "master"), label: ml("sales_color", "สีสำหรับงานขาย", "Sales Colors") },
        ],
      },
      {
        id: "sales-loyalty",
        title: ml("sales_loyalty", "สมาชิก คูปอง และโปรโมชัน", "Loyalty, Coupons & Promotions"),
        items: [
          tx("point-setting", "ตั้งค่าคะแนนสะสม", "Point Setting", "/pointsetting", "master"),
          tx("coupon-setting", "ตั้งค่าคูปอง", "Coupon Setting", "/couponsetting", "master"),
          tx("promotion", "โปรโมชั่น", "Promotion", "/promotionscreen", "master"),
          tx("customer-purchase-cycle", "รอบซื้อลูกค้าประจำ", "Customer Purchase Cycles", "/customerpurchasecycle", "master"),
        ],
      },
      {
        id: "restaurant-setup",
        title: ml("restaurant_cafe", "ร้านอาหาร/คาเฟ่", "Restaurant / Cafe"),
        items: [
          tx("zone", "โซน", "Zone", "/zonegroupselectscreen", "restaurant"),
          tx("table", "โต๊ะ", "Table", "/tablegroupselectscreen", "restaurant"),
          tx("table-map", "ผังโต๊ะ", "Table Map", "/tablemapgroupselectscreen", "restaurant"),
          tx("kitchen", "ครัว", "Kitchen", "/kitchengroupselectscreen", "restaurant"),
          tx("order-template", "ตั้งค่าเครื่องสั่งอาหาร", "Ordering Station Settings", "/ordertemplatsetting", "restaurant"),
          tx("order-setting", "ตั้งค่าการสั่งอาหาร", "Ordering Settings", "/ordersetting", "restaurant"),
          tx("qr-order", "สั่งอาหารด้วย QR", "QR Ordering", "/qrcodeordergroupselectscreen", "restaurant"),
        ],
      },
    ],
  },
];

export function menuText(label: MenuLabel, language: LanguageCode, dictionary?: BackendLanguageDictionary): string {
  const fromBackend = label.key ? dictionary?.[label.key] : "";
  if (fromBackend) return fromBackend;
  if (label.key && dictionary && !isBackendLanguageReady(dictionary)) return label[language] || label.en || label.th || "";
  return label[language] || label.en || label.th || label.key || "";
}

/**
 * Menu full-text search (2026-08-30): "ค้นหาไม่เจอ ต้องแบบ full text search
 * ด้วย ได้ทุกภาษา" — matching used to compare only the CURRENT UI language,
 * so typing English in a Thai UI (or vice versa) found nothing, and a Thai
 * query with a wrong tone mark missed exact labels.
 */
const SEARCH_NOISE = /[\u200B-\u200D\uFEFF\u2060]/g;
const THAI_COMBINING = /[\u0E31\u0E34-\u0E3A\u0E47-\u0E4E]/g;

export function normalizeMenuSearchText(value: string): string {
  return value
    .replace(SEARCH_NOISE, "")
    .replace(THAI_COMBINING, "")
    .toLocaleLowerCase()
    .replace(/\s+/g, " ")
    .trim();
}

/** Every searchable variant of one label: all shipped languages, the key and
    any backend dictionary override — a query in ANY language must match. */
export function menuSearchHaystack(label: MenuLabel, dictionary?: BackendLanguageDictionary): string {
  const parts: string[] = [label.th, label.en];
  for (const code of Object.keys(label)) {
    if (code === "th" || code === "en" || code === "key") continue;
    const value = (label as Record<string, unknown>)[code];
    if (typeof value === "string") parts.push(value);
  }
  if (label.key) {
    parts.push(label.key);
    const fromBackend = dictionary?.[label.key];
    if (fromBackend) parts.push(fromBackend);
  }
  return normalizeMenuSearchText(parts.filter(Boolean).join(" "));
}

export function menuSearchMatches(label: MenuLabel, needle: string, dictionary?: BackendLanguageDictionary): boolean {
  const normalizedNeedle = normalizeMenuSearchText(needle);
  return Boolean(normalizedNeedle) && menuSearchHaystack(label, dictionary).includes(normalizedNeedle);
}

export function flattenMenuItems(): MenuItem[] {
  return MENU_SECTIONS.flatMap((section) => section.groups.flatMap((group) => group.items));
}

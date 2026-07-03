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

const tx = (id: string, th: string, en: string, route: string, category: MenuCategory = "transaction"): MenuItem => ({
  id,
  label: ml(id, th, en),
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
          tx("purchase-requisition", "ใบขอซื้อ", "Purchase Requisition", "/transaction/purchaserequisition"),
          tx("rfq", "ขอใบเสนอราคา", "Request for Quotation", "/transaction/rfq"),
          tx("purchase-order", "ใบสั่งซื้อ", "Purchase Order", "/transaction/purchaseorder"),
          tx("procurement-dashboard", "ภาพรวมจัดซื้อ", "Purchase Overview", "/procurement/dashboard"),
        ],
      },
      {
        id: "purchase",
        title: ml("purchase_group", "ซื้อ", "Purchase"),
        items: [
          tx("purchase", "ซื้อสินค้า", "Purchase", "/transaction/purchase"),
          tx("purchase-return", "คืนซื้อ", "Purchase Return", "/transaction/purchasereturn"),
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
          tx("stock-balance", "ยอดยกมาสินค้า", "Stock Balance", "/transaction/stockbalance"),
        ],
      },
      {
        id: "payment",
        title: ml("receive_pay_money", "รับ/จ่ายเงิน", "Payment"),
        items: [
          tx("paid", "รับชำระ", "Receive Payment", "/transaction/paid", "finance"),
          tx("pay", "จ่ายชำระ", "Pay", "/transaction/pay", "finance"),
          tx("slip-in", "รูปสลิปเงินเข้า", "Money In Slip", "/slipmoneyin", "finance"),
          tx("slip-out", "รูปสลิปเงินออก", "Money Out Slip", "/slipmoneyout", "finance"),
        ],
      },
      {
        id: "accounting",
        title: ml("accounting", "บัญชี", "Accounting"),
        items: [
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
          tx("vat-sale", "รายงานภาษีขาย", "VAT Sale", "/report/reportvatsale", "report"),
          tx("vat-buy", "รายงานภาษีซื้อ", "VAT Purchase", "/report/reportvatbuy", "report"),
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
      // Reorganized 2026-07-03 as a natural first-time-setup workflow (scopeofwork feedback +
      // GPT UX review + Fable chief-architect review): products first (everything references a
      // product), then classify/variants/details, warehouse, trade partners, org basics, sales
      // settings split into 5 focused groups instead of one 15-item wall, approval, marketplace,
      // tools, a new small group for the 2 items that were miscategorized under import/export
      // (knowledge-base and alert-agent are not import/export tools), then import/export itself,
      // then restaurant/cafe last since it's business-type-specific, not core to every business.
      // Every item id/route is unchanged from before — only grouping, group order, and 2 group
      // titles (partners, organization) changed.
      {
        id: "products",
        title: ml("product_catalog", "สินค้าและบาร์โค้ด", "Product Catalog"),
        items: [
          // Core product management (daily use)
          tx("product", "สินค้า", "Product", "/product", "master"),
          { ...tx("barcode", "บาร์โค้ด", "Barcode", "/productbarcode", "master"), label: ml("barcode", "บาร์โค้ด", "Barcode") },
          tx("productset", "สินค้าชุด", "Product Set", "/productset", "master"),
          tx("product-unit", "หน่วยนับสินค้า", "Product Unit", "/productunit", "master"),
        ],
      },
      {
        id: "product-classification",
        title: ml("product_classification", "จัดกลุ่มสินค้า", "Product Grouping"),
        items: [
          tx("product-group", "กลุ่มสินค้า", "Product Group", "/productgroup", "master"),
          tx("product-category", "จัดหมวดสินค้า", "Product Categories", "/productcategorygroupselectscreen", "master"),
          tx("product-category-list", "สินค้าในหมวด", "Products in Category", "/productcategorylist", "master"),
        ],
      },
      {
        id: "product-sku-options",
        title: ml("product_sku_options", "สี ไซซ์ และตัวเลือก", "Product Options"),
        items: [
          tx("product-variant-matrix", "ชุดตัวเลือกสินค้า", "Product Option Sets", "/productvariantmatrix", "master"),
          tx("product-color", "สีสินค้า", "Product Color", "/productcolor", "master"),
          tx("product-size", "ไซซ์/ขนาดสินค้า", "Product Size", "/productsize", "master"),
        ],
      },
      {
        id: "product-descriptors",
        title: ml("product_descriptors", "รายละเอียดประกอบสินค้า", "Product Details"),
        items: [
          tx("brand", "ยี่ห้อสินค้า", "Brand", "/masterbrandscreen", "master"),
          tx("model", "รุ่นสินค้า", "Model", "/mastermodelscreen", "master"),
          tx("pattern", "รูปแบบสินค้า", "Pattern", "/masterpatternscreen", "master"),
          {
            ...tx("category", "คุณลักษณะสินค้า", "Product Attributes", "/mastercategoryscreen", "master"),
            label: ml("product_attribute_category", "คุณลักษณะสินค้า", "Product Attributes"),
          },
          tx("dimension", "ขนาด/มิติสินค้า", "Product Dimensions", "/productdimension", "master"),
          tx("class", "ระดับสินค้า", "Product Class", "/masterclassscreen", "master"),
          tx("design", "รูปทรงสินค้า", "Product Shape", "/masterdesignscreen", "master"),
          tx("grade", "เกรดสินค้า", "Product Grade", "/mastergradescreen", "master"),
        ],
      },
      {
        id: "product-stock-production",
        title: ml("product_stock_production", "คลังและการผลิต", "Stock and Production"),
        items: [
          tx("warehouse", "คลัง", "Warehouse", "/productwarehousescreen", "master"),
          tx("bom", "สูตรผลิต", "Product BOM", "/productbom", "master"),
        ],
      },
      {
        id: "partners",
        title: ml("creditor_debtor", "คู่ค้า (ลูกค้า/ผู้ขาย)", "Trade Partners"),
        items: [
          tx("creditor", "เจ้าหนี้", "Creditor", "/creditor", "master"),
          tx("creditor-group", "กลุ่มเจ้าหนี้", "Creditor Group", "/creditorgroup", "master"),
          tx("debtor", "ลูกหนี้", "Debtor", "/debtor", "master"),
          tx("debtor-group", "กลุ่มลูกหนี้", "Debtor Group", "/debtorgroup", "master"),
        ],
      },
      {
        id: "organization",
        title: ml("organization_master", "งาน โครงการ และศูนย์ต้นทุน", "Jobs, Projects & Cost Centers"),
        items: [
          tx("cost-center", "ศูนย์ต้นทุน", "Cost Center", "/costcenter", "master"),
          tx("project", "โครงการ", "Project", "/project", "master"),
          tx("job", "งาน", "Job", "/job", "master"),
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
        id: "sales-payment-banking",
        title: ml("sales_payment_banking", "การรับเงินและบัญชีธนาคาร", "Payment & Banking"),
        items: [
          tx("qr-provider", "ผู้ให้บริการรับเงิน QR", "QR Payment Provider", "/qrprovider", "master"),
          tx("bank", "ธนาคาร", "Bank", "/bank", "master"),
          tx("book-bank", "สมุดบัญชีธนาคาร", "Bank Book", "/bookbankscreen", "master"),
          tx("exchange-rate", "อัตราแลกเปลี่ยน", "Exchange Rate", "/exchangerate", "finance"),
        ],
      },
      {
        id: "sales-pos",
        title: ml("sales_pos", "หน้าร้าน POS", "Point of Sale"),
        items: [
          tx("pos-setting", "ตั้งค่าเครื่องขายหน้าร้าน (POS)", "Point of Sale Settings", "/possetting", "master"),
          tx("pos-media", "รูป/สื่อหน้าจอขาย", "Point of Sale Media", "/posmedia", "master"),
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
        ],
      },
      {
        id: "sales-documents",
        title: ml("sales_documents", "เอกสารและบิล", "Documents & Billing"),
        items: [
          tx("doc-format", "รูปแบบเอกสาร", "Document Format", "/docformat", "master"),
          tx("bill-design", "ออกแบบบิล", "Bill Design", "/billdesign", "master"),
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
        id: "marketplace-connectors",
        title: ml("marketplace_connectors", "เชื่อมข้อมูลตลาดออนไลน์", "Marketplace Connections"),
        items: [
          tx("shopee-mappings", "เชื่อม Shopee", "Shopee Connection", "/marketplace/shopee", "master"),
          tx("lazada-mappings", "เชื่อม Lazada", "Lazada Connection", "/marketplace/lazada", "master"),
          tx("tiktok-mappings", "เชื่อม TikTok", "TikTok Connection", "/marketplace/tiktok", "master"),
        ],
      },
      {
        id: "product-tools",
        title: ml("product_tools", "เครื่องมือสินค้า", "Product Tools"),
        items: [
          tx("product-serial-registry", "ทะเบียนเลขเครื่อง", "Serial Registry", "/productserialregistry", "master"),
          tx("price-history", "ประวัติแก้ไขราคา", "Price Edit History", "/pricehistory", "master"),
          tx("label-print", "พิมพ์ป้ายสินค้า", "Print Product Label", "/productbarcodeshelf", "master"),
        ],
      },
      {
        id: "product-assistant",
        title: ml("product_assistant", "ผู้ช่วย AI และคลังความรู้", "AI Assistant & Knowledge Base"),
        items: [
          tx("knowledge-base", "คลังความรู้", "Knowledge Base", "/knowledgebasescreen", "master"),
          tx("alert-agent", "ผู้ช่วยแจ้งเตือน", "Alert Assistant", "/alertagentscreen", "master"),
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
        id: "restaurant",
        title: ml("restaurant_cafe", "ร้านอาหาร/คาเฟ่", "Restaurant / Cafe"),
        items: [
          tx("zone", "โซน", "Zone", "/zonegroupselectscreen", "restaurant"),
          tx("table", "โต๊ะ", "Table", "/tablegroupselectscreen", "restaurant"),
          tx("table-map", "ผังโต๊ะ", "Table Map", "/tablemapgroupselectscreen", "restaurant"),
          tx("kitchen", "ครัว", "Kitchen", "/kitchengroupselectscreen", "restaurant"),
          tx("add-product-kitchen", "เพิ่มสินค้าเข้าครัว", "Add Product to Kitchen", "/addproducttokitchenscreen", "restaurant"),
          tx("qr-order", "สั่งอาหารด้วย QR", "QR Ordering", "/qrcodeordergroupselectscreen", "restaurant"),
          tx("order-template", "ตั้งค่าเครื่องสั่งอาหาร", "Ordering Station Settings", "/ordertemplatsetting", "restaurant"),
          tx("order-setting", "ตั้งค่าการสั่งอาหาร", "Ordering Settings", "/ordersetting", "restaurant"),
        ],
      },
    ],
  },
  {
    id: "settings",
    title: { key: "settings", th: "ตั้งค่า", en: "Settings" },
    groups: [
      {
        id: "company-system",
        title: { key: "system_settings", th: "ตั้งค่าระบบ", en: "System Settings" },
        items: [
          tx("active-languages", "ภาษาที่ใช้งาน", "Active Languages", "/activelanguages", "settings"),

          tx("currency", "สกุลเงิน", "Currency", "/currency", "settings"),
          tx("company-type", "ประเภทธุรกิจ", "Business Type", "/businesstypescreen", "settings"),
          tx("employee", "พนักงาน", "Employee", "/employee", "settings"),
          tx("line-oa-user-link", "เชื่อม LINE OA", "Connect LINE OA", "/line-oa", "settings"),
          tx("form-design", "ออกแบบฟอร์ม", "Form Design", "/formdesign", "settings"),
          tx("line-notify", "แจ้งเตือนผ่าน LINE", "LINE Notifications", "/linenotify", "settings"),
          tx("ai-provider", "ผู้ให้บริการ AI", "AI Service Provider", "/aiprovider", "settings"),
          tx("copy-uat-dev", "คัดลอกข้อมูลทดสอบ", "Copy Test Data", "/copyuattodev", "settings"),
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

export function flattenMenuItems(): MenuItem[] {
  return MENU_SECTIONS.flatMap((section) => section.groups.flatMap((group) => group.items));
}

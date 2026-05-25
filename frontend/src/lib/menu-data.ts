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
          tx("procurement-dashboard", "แดชบอร์ดจัดซื้อ", "Procurement Dashboard", "/procurement/dashboard"),
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
          tx("slip-in", "รูปสลิปเงินเข้า", "Money In Slip", "/slip_money_in", "finance"),
          tx("slip-out", "รูปสลิปเงินออก", "Money Out Slip", "/slip_money_out", "finance"),
        ],
      },
      {
        id: "accounting",
        title: ml("accounting", "บัญชี", "Accounting"),
        items: [
          tx("daily-info", "ตรวจสอบประจำวัน", "Daily Check", "/check_daily/daily_info_screen", "finance"),
          tx("cash-drawer", "รับ-ส่งเงิน POS", "POS Cash Drawer", "/cashing_in_the_drawer", "finance"),
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
          tx("stock-balance-item", "คงเหลือตามสินค้า", "Stock by Item", "/report/stock_balance_item", "report"),
          tx("stock-balance-warehouse", "คงเหลือตามคลัง", "Stock by Warehouse", "/report/stock_balance_warehouse", "report"),
          tx("stock-balance-location", "คงเหลือตามที่เก็บ", "Stock by Location", "/report/stock_balance_location", "report"),
          tx("stock-movement-cost", "เคลื่อนไหวสินค้าพร้อมต้นทุน", "Stock Movement with Cost", "/report/stock_movement_cost", "report"),
        ],
      },
      {
        id: "sales-reports",
        title: ml("sales_reports", "รายงานขาย", "Sales Reports"),
        items: [
          tx("gross-profit-doc", "กำไรขั้นต้นตามเอกสาร", "Gross Profit by Document", "/report/sales_report_by_document", "report"),
          tx("sales", "รายงานขาย", "Sales Report", "/report/report_dedebi_sales", "report"),
          tx("sales-daily", "รายงานขายรายวัน", "Daily Sales", "/report/report_dedebi_sales_daily", "report"),
          tx("payment-daily", "รายงานรับชำระรายวัน", "Daily Payment", "/report/report_dedebi_payment_daily", "report"),
          tx("sale-return-report", "รายงานคืนขาย", "Sale Return Report", "/report/report_dedebi_sale_return", "report"),
        ],
      },
      {
        id: "other-reports",
        title: ml("other_reports", "รายงานอื่น ๆ", "Other Reports"),
        items: [
          tx("product-movement", "ความเคลื่อนไหวสินค้า", "Product Movement", "/report/report_stock_movement", "report"),
          tx("inventory", "รายงานสินค้าคงเหลือ", "Inventory Report", "/report/report_dedebi_stock_balance", "report"),
          tx("gradual-receipt", "รายงานทยอยรับ", "Gradual Receipt Report", "/report/report_dedebi_purchase_partial", "report"),
          tx("gross-profit-by-document", "กำไรขั้นต้นตามเอกสาร", "Gross Profit by Document", "/report/report_gross_profit_by_document", "report"),
          tx("gross-profit-by-product", "กำไรขั้นต้นตามสินค้า", "Gross Profit by Product", "/report/report_gross_profit_by_product", "report"),
          tx("vat-sale", "รายงานภาษีขาย", "VAT Sale", "/report/report_vat_sale", "report"),
          tx("vat-buy", "รายงานภาษีซื้อ", "VAT Purchase", "/report/report_vat_buy", "report"),
        ],
      },
      {
        id: "audit-reports",
        title: ml("audit_reprocess", "ตรวจสอบ/ประมวลผล", "Audit / Reprocess"),
        items: [
          tx("reprocess", "ทดสอบประมวลผลใหม่", "Reprocess Test", "/rebuild_stock_screen", "report"),
          tx("audit-data", "ตรวจสอบข้อมูล", "Audit Data", "/audit_screen", "report"),
          tx("rebuild-products", "สร้างข้อมูลสินค้าใหม่", "Rebuild Products", "/rebuild_products_screen", "report"),
          tx("rebuild-product-balance", "สร้างยอดคงเหลือใหม่", "Rebuild Product Balance", "/rebuild_product_balance_screen", "report"),
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
        title: ml("product_management", "จัดการสินค้า", "Products"),
        items: [
          { ...tx("barcode", "บาร์โค้ด", "Barcode", "/product_barcode", "master"), label: ml("barcode", "บาร์โค้ด", "Barcode") },
          tx("product-unit", "หน่วยนับสินค้า", "Product Unit", "/productunit", "master"),
          tx("label-print", "พิมพ์ป้ายสินค้า", "Print Product Label", "/product_barcode_shelf", "master"),
          tx("promotion", "โปรโมชั่น", "Promotion", "/promotion_screen", "master"),
          tx("price-history", "ประวัติแก้ไขราคา", "Price Edit History", "/price_history", "master"),
        ],
      },
      {
        id: "product-settings",
        title: ml("product_settings", "ตั้งค่าสินค้า", "Product Settings"),
        items: [
          tx("product-category", "หมวดสินค้า", "Product Category", "/product_category_group_select_screen", "master"),
          tx("product-category-list", "รายการหมวดสินค้า", "Product Category List", "/productcategorylist", "master"),
          tx("product-group", "กลุ่มสินค้า", "Product Group", "/productgroup", "master"),
          tx("warehouse", "คลังสินค้า", "Warehouse", "/product_warehouse_screen", "master"),
          tx("location", "ที่เก็บสินค้า", "Location", "/product_location_screen", "master"),
          tx("order-type", "ประเภทคำสั่ง", "Order Type", "/order_type_screen", "master"),
          tx("product-type", "ประเภทสินค้า", "Product Type", "/product_type_screen", "master"),
          tx("dimension", "มิติสินค้า", "Product Dimension", "/product_dimension", "master"),
          tx("bom", "สูตรประกอบสินค้า", "Product BOM", "/product_bom", "master"),
          tx("brand", "ยี่ห้อ", "Brand", "/master_brand_screen", "master"),
          tx("category", "Category", "Category", "/master_category_screen", "master"),
          tx("class", "Class", "Class", "/master_class_screen", "master"),
          tx("design", "Design", "Design", "/master_design_screen", "master"),
          tx("grade", "Grade", "Grade", "/master_grade_screen", "master"),
          tx("model", "Model", "Model", "/master_model_screen", "master"),
          tx("pattern", "Pattern", "Pattern", "/master_pattern_screen", "master"),
          tx("group-main", "กลุ่มหลัก", "Main Group", "/master_group_screen", "master"),
          tx("group-sub1", "กลุ่มย่อย 1", "Sub Group 1", "/master_group_sub1_screen", "master"),
          tx("group-sub2", "กลุ่มย่อย 2", "Sub Group 2", "/master_group_sub2_screen", "master"),
        ],
      },
      {
        id: "database",
        title: ml("product_database", "จัดการฐานข้อมูลสินค้า", "Product Database"),
        items: [
          tx("add-product-branch", "เพิ่มสินค้าเข้าตามสาขา", "Add Product to Branch", "/add_product_to_branch_screen", "master"),
          tx("add-product-department", "เพิ่มสินค้าเข้าตามแผนก", "Add Product to Department", "/add_product_to_department_screen", "master"),
        ],
      },
      {
        id: "partners",
        title: ml("creditor_debtor", "เจ้าหนี้/ลูกหนี้", "Partners"),
        items: [
          tx("creditor", "เจ้าหนี้", "Creditor", "/creditor", "master"),
          tx("creditor-group", "กลุ่มเจ้าหนี้", "Creditor Group", "/creditorgroup", "master"),
          tx("debtor", "ลูกหนี้", "Debtor", "/debtor", "master"),
          tx("debtor-group", "กลุ่มลูกหนี้", "Debtor Group", "/debtorgroup", "master"),
        ],
      },
      {
        id: "organization",
        title: ml("organization_master", "ข้อมูลหลักองค์กร", "Organization"),
        items: [
          tx("cost-center", "ศูนย์ต้นทุน", "Cost Center", "/costcenter", "master"),
          tx("project", "โครงการ", "Project", "/project", "master"),
          tx("job", "งาน", "Job", "/job", "master"),
        ],
      },
      {
        id: "sales-settings",
        title: ml("sales_settings", "ตั้งค่าการขาย", "Sales Settings"),
        items: [
          tx("color", "สี", "Color", "/color_screen", "master"),
          tx("qr-provider", "ผู้ให้บริการ QR", "QR Provider", "/qrprovider", "master"),
          tx("bank", "ธนาคาร", "Bank", "/bank", "master"),
          tx("book-bank", "สมุดบัญชีธนาคาร", "Bank Book", "/book_bank_screen", "master"),
          tx("sale-channel", "ช่องทางขาย", "Sale Channel", "/sale_channel_screen", "master"),
          tx("transport-channel", "ช่องทางขนส่ง", "Transport Channel", "/transport_channel_screen", "master"),
          tx("pos-setting", "ตั้งค่า POS", "POS Setting", "/possetting", "master"),
          tx("exchange-rate", "อัตราแลกเปลี่ยน", "Exchange Rate", "/exchange_rate", "finance"),
          tx("pos-media", "สื่อ POS", "POS Media", "/posmedia", "master"),
          tx("point-setting", "ตั้งค่าแต้ม", "Point Setting", "/point_setting", "master"),
          tx("coupon-setting", "ตั้งค่าคูปอง", "Coupon Setting", "/coupon_setting", "master"),
          tx("doc-format", "รูปแบบเอกสาร", "Document Format", "/docformat", "master"),
          tx("bill-design", "ออกแบบบิล", "Bill Design", "/billdesign", "master"),
        ],
      },
      {
        id: "approval",
        title: ml("approval", "อนุมัติ", "Approval"),
        items: [
          tx("purchase-type", "ประเภทซื้อ", "Purchase Type", "/purchase_type_screen", "approval"),
          tx("po-approval", "อนุมัติใบสั่งซื้อ", "PO Approval", "/po_approval_setting_screen", "approval"),
          tx("quotation-type", "ประเภทใบเสนอราคา", "Quotation Type", "/quotation_type_screen", "approval"),
          tx("qt-approval", "อนุมัติใบเสนอราคา", "Quotation Approval", "/qt_approval_setting_screen", "approval"),
          tx("sale-order-type", "ประเภทใบสั่งขาย", "Sale Order Type", "/sale_order_type_screen", "approval"),
          tx("so-approval", "อนุมัติใบสั่งขาย", "Sale Order Approval", "/so_approval_setting_screen", "approval"),
        ],
      },
      {
        id: "import-export",
        title: ml("import_export_data", "นำเข้า/ส่งออกข้อมูล", "Import / Export"),
        items: [
          tx("import-product", "นำเข้าสินค้า", "Import Product", "/importproduct", "master"),
          tx("import-product-file", "นำเข้าข้อมูลสินค้า", "Import Product Data", "/importproductfromfile", "master"),
          tx("knowledge-base", "คลังความรู้", "Knowledge Base", "/knowledge_base_screen", "master"),
          tx("alert-agent", "BC Ai Account Alert Agent", "BC Ai Account Alert Agent", "/alert_agent_screen", "master"),
          tx("import-product-image", "นำเข้ารูปสินค้า", "Import Product Image", "/importproductimage", "master"),
        ],
      },
      {
        id: "restaurant",
        title: ml("restaurant_cafe", "ร้านอาหาร/คาเฟ่", "Restaurant / Cafe"),
        items: [
          tx("zone", "โซน", "Zone", "/zone_group_select_screen", "restaurant"),
          tx("table", "โต๊ะ", "Table", "/table_group_select_screen", "restaurant"),
          tx("table-map", "ผังโต๊ะ", "Table Map", "/table_map_group_select_screen", "restaurant"),
          tx("kitchen", "ครัว", "Kitchen", "/kitchen_group_select_screen", "restaurant"),
          tx("add-product-kitchen", "เพิ่มสินค้าเข้าครัว", "Add Product to Kitchen", "/add_product_to_kitchen_screen", "restaurant"),
          tx("qr-order", "QR Code สั่งอาหาร", "QR Order", "/qrcodeorder_group_select_screen", "restaurant"),
          tx("order-template", "ตั้งค่า Template Order", "Order Template Setting", "/ordertemplatsetting", "restaurant"),
          tx("order-setting", "ตั้งค่า Order", "Order Setting", "/ordersetting", "restaurant"),
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
          tx("active-languages", "ภาษาที่ใช้งาน", "Active Languages", "/active_languages", "settings"),
          tx("user", "ผู้ใช้งาน", "User", "/user", "settings"),
          tx("permission-definition", "กำหนดสิทธิ์หน้าจอ", "Permission Definition", "/permission_definition", "settings"),
          tx("permission-group", "กำหนดสิทธิ์ตามกลุ่ม", "Permission Group", "/permission_group", "settings"),
          tx("approval-setting", "สิทธิ์การอนุมัติ", "Approval Permission", "/approval_setting", "settings"),
          tx("permission-link", "กำหนดสิทธิ์พนักงาน", "Permission Link", "/permission_link", "settings"),
          tx("currency", "สกุลเงิน", "Currency", "/currency", "settings"),
          tx("company-type", "ประเภทธุรกิจ", "Business Type", "/business_type_screen", "settings"),
          tx("company", "บริษัท", "Company", "/company", "settings"),
          tx("branch", "สาขา", "Branch", "/branch", "settings"),
          tx("employee", "พนักงาน", "Employee", "/employee", "settings"),
          tx("line-oa-user-link", "เชื่อม LINE OA", "Connect LINE OA", "/line-oa", "settings"),
          tx("form-design", "ออกแบบฟอร์ม", "Form Design", "/formdesign", "settings"),
          tx("line-notify", "LINE Notify", "LINE Notify", "/line_notify", "settings"),
          tx("mcp-token", "MCP Token", "MCP Token", "/mcp_apikey", "settings"),
          tx("ai-provider", "AI Provider", "AI Provider", "/ai_provider", "settings"),
          tx("copy-uat-dev", "โอนข้อมูล UAT ไป DEV", "Copy UAT to DEV", "/copy_uat_to_dev", "settings"),
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

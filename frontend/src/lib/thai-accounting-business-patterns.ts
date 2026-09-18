// Thai Accounting Business Patterns & Smart Journal Templates
// ครอบคลุม 8 กลุ่มธุรกิจหลักของประเทศไทย ตามมาตรฐาน TFRS for NPAEs / ประมวลรัษฎากร

export type ThaiBusinessType =
  | "trading"             // ธุรกิจซื้อมาขายไป / ค้าปลีก-ส่ง
  | "service"             // ธุรกิจบริการ / ที่ปรึกษา / วิชาชีพอิสระ
  | "manufacturing"       // ธุรกิจโรงงาน / ผลิตสินค้า
  | "restaurant_cafe"     // ร้านอาหาร / คาเฟ่ / เครื่องดื่ม
  | "construction"        // รับเหมาก่อสร้าง / ตกแต่ง / งานระบบ
  | "ecommerce"           // ร้านค้าออนไลน์ / มาร์เก็ตเพลส / Social Commerce
  | "real_estate_rental"  // อสังหาริมทรัพย์ / ให้เช่าพื้นที่และโกดัง
  | "transport_logistics";// ขนส่ง / คลังสินค้า / โลจิสติกส์

export interface ThaiAccountTemplate {
  code: string;
  nameTh: string;
  nameEn: string;
  category: "asset" | "liability" | "equity" | "revenue" | "expense";
  normalBalance: "debit" | "credit";
  description?: string;
}

export interface ThaiJournalLineTemplate {
  accountCode: string;
  accountNameTh: string;
  side: "debit" | "credit";
  /** สูตรคำนวณสัดส่วนหรือมูลค่า: 'base' | 'vat7' | 'wht1' | 'wht2' | 'wht3' | 'wht5' | 'net_payable' | 'net_receivable' */
  formula:
    | "base"
    | "vat7"
    | "wht1"
    | "wht2"
    | "wht3"
    | "wht5"
    | "net_payable"
    | "net_receivable"
    | "custom";
  descriptionTh: string;
}

export interface ThaiJournalPatternTemplate {
  id: string;
  businessType: ThaiBusinessType;
  titleTh: string;
  titleEn: string;
  keywords: string[];
  lines: ThaiJournalLineTemplate[];
  taxNotesTh: string;
  tfrsReference?: string;
}

export interface ThaiBusinessMetadata {
  type: ThaiBusinessType;
  titleTh: string;
  titleEn: string;
  descriptionTh: string;
  taxCharacteristicsTh: string[];
  recommendedAccounts: ThaiAccountTemplate[];
}

/** 1. ธุรกิจซื้อมาขายไป / ค้าปลีก-ส่ง (Trading) */
const TRADING_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1111", nameTh: "เงินสดในมือ", nameEn: "Cash on Hand", category: "asset", normalBalance: "debit" },
  { code: "1112", nameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", nameEn: "Cash at Bank", category: "asset", normalBalance: "debit" },
  { code: "1131", nameTh: "ลูกหนี้การค้า", nameEn: "Trade Accounts Receivable", category: "asset", normalBalance: "debit" },
  { code: "1141", nameTh: "สินค้าคงเหลือสำเร็จรูป", nameEn: "Merchandise Inventory", category: "asset", normalBalance: "debit" },
  { code: "1151", nameTh: "ภาษีซื้อ", nameEn: "Input VAT", category: "asset", normalBalance: "debit" },
  { code: "2111", nameTh: "เจ้าหนี้การค้า", nameEn: "Trade Accounts Payable", category: "liability", normalBalance: "credit" },
  { code: "2141", nameTh: "ภาษีขาย", nameEn: "Output VAT", category: "liability", normalBalance: "credit" },
  { code: "4111", nameTh: "รายได้จากการขายสินค้า", nameEn: "Sales Revenue", category: "revenue", normalBalance: "credit" },
  { code: "4112", nameTh: "รับคืนและส่วนลดจ่าย", nameEn: "Sales Returns and Allowances", category: "revenue", normalBalance: "debit" },
  { code: "5111", nameTh: "ซื้อสินค้า", nameEn: "Purchases", category: "expense", normalBalance: "debit" },
  { code: "5112", nameTh: "ค่าขนส่งเข้า", nameEn: "Freight-in", category: "expense", normalBalance: "debit" },
  { code: "5113", nameTh: "ส่งคืนและส่วนลดรับ", nameEn: "Purchase Returns and Allowances", category: "expense", normalBalance: "credit" },
  { code: "5211", nameTh: "ค่าขนส่งออกและจัดส่ง", nameEn: "Freight-out", category: "expense", normalBalance: "debit" },
];

/** 2. ธุรกิจบริการ / ที่ปรึกษา / ฟรีแลนซ์ (Service) */
const SERVICE_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1112", nameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", nameEn: "Cash at Bank", category: "asset", normalBalance: "debit" },
  { code: "1131", nameTh: "ลูกหนี้การค้าค่าบริการ", nameEn: "Accounts Receivable - Services", category: "asset", normalBalance: "debit" },
  { code: "1151", nameTh: "ภาษีซื้อ", nameEn: "Input VAT", category: "asset", normalBalance: "debit" },
  { code: "1152", nameTh: "ภาษีซื้อยังไม่ถึงกำหนดชำระ", nameEn: "Undue Input VAT", category: "asset", normalBalance: "debit" },
  { code: "1161", nameTh: "ภาษีถูกหัก ณ ที่จ่าย (3%)", nameEn: "Withholding Tax Receivable (3%)", category: "asset", normalBalance: "debit" },
  { code: "2112", nameTh: "เจ้าหนี้ค่าบริการและค่าใช้จ่ายค้างจ่าย", nameEn: "Accrued Service Expenses", category: "liability", normalBalance: "credit" },
  { code: "2141", nameTh: "ภาษีขาย", nameEn: "Output VAT", category: "liability", normalBalance: "credit" },
  { code: "2142", nameTh: "ภาษีขายยังไม่ถึงกำหนดชำระ", nameEn: "Undue Output VAT", category: "liability", normalBalance: "credit" },
  { code: "2151", nameTh: "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย (ภ.ง.ด.3/53)", nameEn: "Withholding Tax Payable", category: "liability", normalBalance: "credit" },
  { code: "4121", nameTh: "รายได้ค่าบริการและที่ปรึกษา", nameEn: "Service Revenue", category: "revenue", normalBalance: "credit" },
  { code: "5311", nameTh: "ค่าจ้างทำของและบริการภายนอก", nameEn: "Subcontract and Outsourced Services", category: "expense", normalBalance: "debit" },
  { code: "5321", nameTh: "ค่าเช่าสำนักงานและสถานที่", nameEn: "Office Rent", category: "expense", normalBalance: "debit" },
  { code: "5322", nameTh: "ค่าโฆษณาและส่งเสริมการตลาด", nameEn: "Advertising and Marketing", category: "expense", normalBalance: "debit" },
];

/** 3. ธุรกิจโรงงาน / ผลิตสินค้า (Manufacturing) */
const MANUFACTURING_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1142", nameTh: "วัตถุดิบคงเหลือ", nameEn: "Raw Materials", category: "asset", normalBalance: "debit" },
  { code: "1143", nameTh: "งานระหว่างทำ (WIP)", nameEn: "Work in Process", category: "asset", normalBalance: "debit" },
  { code: "1144", nameTh: "สินค้าสำเร็จรูป (FG)", nameEn: "Finished Goods", category: "asset", normalBalance: "debit" },
  { code: "1148", nameTh: "วัสดุโรงงานและวัสดุสิ้นเปลือง", nameEn: "Factory Supplies", category: "asset", normalBalance: "debit" },
  { code: "5121", nameTh: "ซื้อวัตถุดิบทางตรง", nameEn: "Direct Materials Purchased", category: "expense", normalBalance: "debit" },
  { code: "5122", nameTh: "ค่าแรงงานทางตรง", nameEn: "Direct Labor", category: "expense", normalBalance: "debit" },
  { code: "5123", nameTh: "ค่าไฟฟ้าโรงงานและพลังงานผลิต", nameEn: "Factory Utilities & Power", category: "expense", normalBalance: "debit" },
  { code: "5124", nameTh: "ค่าเสื่อมราคาเครื่องจักรโรงงาน", nameEn: "Machinery Depreciation", category: "expense", normalBalance: "debit" },
  { code: "5125", nameTh: "ค่าซ่อมบำรุงโรงงานและเครื่องจักร", nameEn: "Factory Maintenance", category: "expense", normalBalance: "debit" },
  { code: "5129", nameTh: "คุมโสหุ้ยการผลิต (Overhead Control)", nameEn: "Manufacturing Overhead Control", category: "expense", normalBalance: "debit" },
  { code: "5100", nameTh: "ต้นทุนขายสินค้าผลิตสำเร็จ", nameEn: "Cost of Goods Sold - Manufactured", category: "expense", normalBalance: "debit" },
];

/** 4. ร้านอาหาร คาเฟ่ และ F&B (Restaurant & Cafe) */
const RESTAURANT_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1111", nameTh: "เงินสดในลิ้นชัก/หน้าร้าน", nameEn: "Cash Drawer / Front Store", category: "asset", normalBalance: "debit" },
  { code: "1113", nameTh: "เงินโอนรับชำระผ่าน QR Code", nameEn: "PromptPay & QR Clearing", category: "asset", normalBalance: "debit" },
  { code: "1132", nameTh: "ลูกหนี้แพลตฟอร์มเดลิเวอรี (Grab/Lineman/Foodpanda)", nameEn: "Delivery Platform Receivables", category: "asset", normalBalance: "debit" },
  { code: "1145", nameTh: "วัตถุดิบสด/เนื้อสัตว์/ผัก (ไม่มี VAT)", nameEn: "Fresh Ingredients (VAT Exempt)", category: "asset", normalBalance: "debit" },
  { code: "1146", nameTh: "วัตถุดิบเครื่องดื่ม/บรรจุภัณฑ์ (มี VAT)", nameEn: "Beverages & Packaging (Vatable)", category: "asset", normalBalance: "debit" },
  { code: "4131", nameTh: "รายได้จากการขายอาหารและเครื่องดื่มหน้าร้าน", nameEn: "Dine-in & Takeaway Sales", category: "revenue", normalBalance: "credit" },
  { code: "4132", nameTh: "รายได้จากการขายผ่านเดลิเวอรี", nameEn: "Delivery Sales Revenue", category: "revenue", normalBalance: "credit" },
  { code: "5331", nameTh: "ค่าคอมมิชชันและค่า GP เดลิเวอรี", nameEn: "Delivery Platform GP Commission", category: "expense", normalBalance: "debit" },
  { code: "5332", nameTh: "ค่าแก๊สหุงต้มและน้ำแข็ง", nameEn: "Cooking Gas & Ice Expenses", category: "expense", normalBalance: "debit" },
  { code: "5333", nameTh: "ค่าบรรจุภัณฑ์กล่องอาหารและถุง", nameEn: "Packaging Supplies", category: "expense", normalBalance: "debit" },
];

/** 5. รับเหมาก่อสร้างและติดตั้ง (Construction) */
const CONSTRUCTION_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1133", nameTh: "เงินประกันผลงานค้างรับ (Retention Receivable)", nameEn: "Retention Receivable", category: "asset", normalBalance: "debit" },
  { code: "1147", nameTh: "งานระหว่างก่อสร้างตามสัญญา (Construction in Progress - CIP)", nameEn: "Construction in Progress", category: "asset", normalBalance: "debit" },
  { code: "2113", nameTh: "เงินประกันผลงานค้างจ่ายผู้รับเหมาช่วง (Retention Payable)", nameEn: "Retention Payable to Subcontractors", category: "liability", normalBalance: "credit" },
  { code: "2131", nameTh: "เงินรับล่วงหน้าค่างวดงานก่อสร้าง", nameEn: "Advances from Construction Contracts", category: "liability", normalBalance: "credit" },
  { code: "4141", nameTh: "รายได้ตามสัญญาก่อสร้าง", nameEn: "Construction Contract Revenue", category: "revenue", normalBalance: "credit" },
  { code: "5131", nameTh: "ต้นทุนค่าวัสดุก่อสร้าง", nameEn: "Construction Materials Cost", category: "expense", normalBalance: "debit" },
  { code: "5132", nameTh: "ค่าจ้างผู้รับเหมาช่วง (Subcontractor)", nameEn: "Subcontractor Expenses", category: "expense", normalBalance: "debit" },
  { code: "5133", nameTh: "ค่าเช่าเครื่องจักรและนั่งร้านก่อสร้าง", nameEn: "Scaffolding & Heavy Equipment Rental", category: "expense", normalBalance: "debit" },
  { code: "5134", nameTh: "ค่าออกแบบและวิศวกรรมควบคุม", nameEn: "Architectural & Engineering Fees", category: "expense", normalBalance: "debit" },
];

/** 6. ร้านค้าออนไลน์ และ Social Commerce (E-Commerce) */
const ECOMMERCE_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1134", nameTh: "ลูกหนี้ Shopee / Lazada / TikTok Shop", nameEn: "Marketplace Pending Payouts", category: "asset", normalBalance: "debit" },
  { code: "2143", nameTh: "ภาษีมูลค่าเพิ่มรอนำส่ง ภ.พ.36 (ยิงแอดต่างประเทศ)", nameEn: "PP.36 VAT Remittance Payable", category: "liability", normalBalance: "credit" },
  { code: "4151", nameTh: "รายได้จากการขายผ่าน Shopee/Lazada/TikTok", nameEn: "Marketplace Sales Revenue", category: "revenue", normalBalance: "credit" },
  { code: "5341", nameTh: "ค่าธรรมเนียมธุรกรรมการชำระเงิน (Transaction Fee)", nameEn: "Marketplace Payment Transaction Fees", category: "expense", normalBalance: "debit" },
  { code: "5342", nameTh: "ค่าคอมมิชชันและค่าบริการมาร์เก็ตเพลส", nameEn: "Marketplace Commission Fees", category: "expense", normalBalance: "debit" },
  { code: "5343", nameTh: "ค่าบริการขนส่งพัสดุ (Flash/Kerry/J&T)", nameEn: "Courier & Parcel Delivery Fees", category: "expense", normalBalance: "debit" },
  { code: "5344", nameTh: "ค่าโฆษณายิงแอด Facebook / TikTok / Google Ads", nameEn: "Online Advertising Ads Spent", category: "expense", normalBalance: "debit" },
  { code: "5345", nameTh: "ค่าไลฟ์สดและคอมมิชชันครีเอเตอร์ (Affiliate)", nameEn: "Live Streaming & Affiliate Fees", category: "expense", normalBalance: "debit" },
];

/** 7. อสังหาริมทรัพย์และให้เช่าพื้นที่ (Real Estate & Rental) */
const REAL_ESTATE_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1162", nameTh: "ภาษีถูกหัก ณ ที่จ่ายค่าเช่า (5%)", nameEn: "Withholding Tax Receivable - Rent (5%)", category: "asset", normalBalance: "debit" },
  { code: "2161", nameTh: "เงินประกันและมัดจำการเช่ารับล่วงหน้า", nameEn: "Tenant Security Deposits", category: "liability", normalBalance: "credit" },
  { code: "4161", nameTh: "รายได้ค่าเช่าอสังหาริมทรัพย์ (ยกเว้น VAT ม.81(1)(ต))", nameEn: "Property Rental Revenue (VAT Exempt)", category: "revenue", normalBalance: "credit" },
  { code: "4162", nameTh: "รายได้ค่าบริการส่วนกลางและสาธารณูปโภค (มี VAT 7%)", nameEn: "Common Area & Facility Services (7% VAT)", category: "revenue", normalBalance: "credit" },
  { code: "4163", nameTh: "รายได้ค่าน้ำ-ค่าไฟฟ้าเก็บจากผู้เช่า", nameEn: "Tenant Reimbursed Utilities Revenue", category: "revenue", normalBalance: "credit" },
  { code: "5361", nameTh: "ค่ารักษาความปลอดภัยและทำความสะอาด", nameEn: "Security & Cleaning Services", category: "expense", normalBalance: "debit" },
  { code: "5362", nameTh: "ค่าบำรุงรักษาอาคาร ลิฟต์ และระบบดับเพลิง", nameEn: "Building & Facility Maintenance", category: "expense", normalBalance: "debit" },
  { code: "5363", nameTh: "ภาษีที่ดินและสิ่งปลูกสร้าง", nameEn: "Land and Building Taxes", category: "expense", normalBalance: "debit" },
];

/** 8. ขนส่งและโลจิสติกส์ (Transport & Logistics) */
const TRANSPORT_ACCOUNTS: ThaiAccountTemplate[] = [
  { code: "1163", nameTh: "ภาษีถูกหัก ณ ที่จ่ายค่าขนส่ง (1%)", nameEn: "Withholding Tax Receivable - Transport (1%)", category: "asset", normalBalance: "debit" },
  { code: "4171", nameTh: "รายได้ค่าบริการขนส่งทางบก (ยกเว้น VAT ม.81(1)(ณ))", nameEn: "Freight Transportation Revenue (VAT Exempt)", category: "revenue", normalBalance: "credit" },
  { code: "4172", nameTh: "รายได้ค่าบริการคลังสินค้าและยกขน (มี VAT 7%)", nameEn: "Warehousing & Handling Services (7% VAT)", category: "revenue", normalBalance: "credit" },
  { code: "5351", nameTh: "ค่าน้ำมันเชื้อเพลิงและก๊าซ NGV/LPG", nameEn: "Fleet Fuel Expenses", category: "expense", normalBalance: "debit" },
  { code: "5352", nameTh: "ค่าผ่านทางด่วนและมอเตอร์เวย์ (Easy Pass)", nameEn: "Toll & Expressway Fees", category: "expense", normalBalance: "debit" },
  { code: "5353", nameTh: "ค่าซ่อมบำรุงและเปลี่ยนยางรถบรรทุก", nameEn: "Fleet Vehicle Repairs & Tires", category: "expense", normalBalance: "debit" },
  { code: "5354", nameTh: "ค่าเบี้ยประกันภัยรถบรรทุกและสินค้า", nameEn: "Fleet & Cargo Insurance", category: "expense", normalBalance: "debit" },
  { code: "5355", nameTh: "ค่าต่อทะเบียน พ.ร.บ. และภาษีรถยนต์", nameEn: "Vehicle License & Annual Tax", category: "expense", normalBalance: "debit" },
];

/** รวบรวมข้อมูลประเภทธุรกิจ */
export const THAI_BUSINESS_METADATA: Record<ThaiBusinessType, ThaiBusinessMetadata> = {
  trading: {
    type: "trading",
    titleTh: "ซื้อมาขายไป / ค้าปลีก-ส่ง",
    titleEn: "Trading & Retail/Wholesale",
    descriptionTh: "กิจการจำหน่ายสินค้า มีสต๊อกสินค้าคงคลัง ออกใบกำกับภาษีขาย และได้รับใบกำกับภาษีซื้อสินค้า",
    taxCharacteristicsTh: [
      "ภาษีมูลค่าเพิ่ม 7% ทั้งฝั่งซื้อสินค้าและขายสินค้า",
      "การขายสินค้าไม่มีการหักภาษี ณ ที่จ่ายระหว่างนิติบุคคลทั่วไป",
      "มีระบบสต๊อกสินค้าคงคลัง (Inventory Counting) บันทึกปรับปรุงปลายงวด",
    ],
    recommendedAccounts: TRADING_ACCOUNTS,
  },
  service: {
    type: "service",
    titleTh: "ธุรกิจบริการ / ที่ปรึกษา / วิชาชีพอิสระ",
    titleEn: "Service & Consulting",
    descriptionTh: "กิจการให้บริการ ให้คำปรึกษา หรือรับจ้างทำของ จุดรับรู้ภาษีขายเมื่อได้รับชำระเงิน",
    taxCharacteristicsTh: [
      "รายได้ค่าบริการถูกผู้ว่าจ้างนิติบุคคลหักภาษี ณ ที่จ่าย 3% (ภ.ง.ด.53)",
      "จุดรับรู้ภาษีขายเกิดขึ้นเมื่อ 'ได้รับชำระเงิน' หรือออกใบกำกับภาษี",
      "ค่าเช่าสถานที่ถูกหัก ณ ที่จ่าย 5%, ค่าบริการ/จ้างทำของ หัก 3%, ค่าโฆษณา หัก 2%",
    ],
    recommendedAccounts: SERVICE_ACCOUNTS,
  },
  manufacturing: {
    type: "manufacturing",
    titleTh: "โรงงานอุตสาหกรรม / ผลิตสินค้า",
    titleEn: "Manufacturing & Industrial",
    descriptionTh: "กิจการแปรรูปวัตถุดิบเป็นสินค้าสำเร็จรูป มี 3 สต๊อก (วัตถุดิบ, งานระหว่างทำ, สินค้าสำเร็จรูป)",
    taxCharacteristicsTh: [
      "ซื้อวัตถุดิบและเครื่องจักรมีภาษีซื้อ 7%",
      "คำนวณต้นทุนการผลิต 3 องค์ประกอบ (Direct Materials, Direct Labor, Manufacturing Overhead)",
      "ต้องมีการกระทบยอดและตรวจนับวัตถุดิบและงานระหว่างทำสิ้นงวด",
    ],
    recommendedAccounts: MANUFACTURING_ACCOUNTS,
  },
  restaurant_cafe: {
    type: "restaurant_cafe",
    titleTh: "ร้านอาหาร / คาเฟ่ / F&B",
    titleEn: "Restaurant & Cafe / Food & Beverage",
    descriptionTh: "กิจการจำหน่ายอาหารและเครื่องดื่ม มีทั้งการขายหน้าร้าน (Cash/QR) และเดลิเวอรี (Grab/Lineman)",
    taxCharacteristicsTh: [
      "วัตถุดิบอาหารสด ผัก ผลไม้ เนื้อสัตว์ ยกเว้นภาษีมูลค่าเพิ่ม (ไม่มี VAT)",
      "เครื่องดื่ม วัตถุดิบแปรรูป บรรจุภัณฑ์ มีภาษีมูลค่าเพิ่ม 7%",
      "ยอดขายผ่านแอปเดลิเวอรีถูกหักค่า GP ประมาณ 30% (+ VAT 7% ของค่า GP) และร้านต้องออกใบกำกับให้ลูกค้าเต็มจำนวน",
    ],
    recommendedAccounts: RESTAURANT_ACCOUNTS,
  },
  construction: {
    type: "construction",
    titleTh: "รับเหมาก่อสร้าง / ตกแต่ง / งานติดตั้ง",
    titleEn: "Construction & Contractor",
    descriptionTh: "กิจการรับเหมาก่อสร้าง ออกใบแจ้งหนี้ตามงวดงาน (Progress Billing) และมีการหักเงินประกันผลงาน",
    taxCharacteristicsTh: [
      "ค่างวดงานถูกผู้ว่าจ้างหักภาษี ณ ที่จ่าย 3% (ภ.ง.ด.53)",
      "มีการหักเงินประกันผลงาน (Retention) 5% - 10% จนกว่าจะหมดระยะเวลารับประกัน",
      "ต้องบันทึกแยกงานระหว่างก่อสร้างตามสัญญา (CIP) แต่ละโครงการเพื่อคิดกำไรรายโครงการ",
    ],
    recommendedAccounts: CONSTRUCTION_ACCOUNTS,
  },
  ecommerce: {
    type: "ecommerce",
    titleTh: "ร้านค้าออนไลน์ / มาร์เก็ตเพลส / Social Commerce",
    titleEn: "E-Commerce & Marketplace Merchants",
    descriptionTh: "ขายสินค้าผ่าน Shopee, Lazada, TikTok Shop และช่องทางโซเชียล มีค่าธรรมเนียมและค่ายิงแอด",
    taxCharacteristicsTh: [
      "แพลตฟอร์มหักค่าธรรมเนียมการขาย ค่าคอมมิชชัน และออกใบกำกับภาษีให้กิจการ",
      "ค่ายิงแอด Facebook / Google / TikTok นอกประเทศ ต้องนำส่งภาษีมูลค่าเพิ่มด้วยแบบ ภ.พ.36 (VAT 7%)",
      "ต้องบันทึกกระทบยอดเงินโอนเข้าบัญชี (Payout Settlement) กับยอดขายจริง",
    ],
    recommendedAccounts: ECOMMERCE_ACCOUNTS,
  },
  real_estate_rental: {
    type: "real_estate_rental",
    titleTh: "อสังหาริมทรัพย์และให้เช่าพื้นที่",
    titleEn: "Real Estate & Property Rental",
    descriptionTh: "กิจการให้เช่าอาคาร สำนักงาน คอนโด หรือโกดังสินค้า มีการแยกค่าเช่าและค่าบริการส่วนกลาง",
    taxCharacteristicsTh: [
      "ค่าเช่าอสังหาริมทรัพย์ 'ยกเว้นภาษีมูลค่าเพิ่ม' ตามมาตรา 81(1)(ต) แต่ถูกหัก ณ ที่จ่าย 5%",
      "ค่าบริการส่วนกลาง ค่าเฟอร์นิเจอร์ หรือค่าที่จอดรถ 'มีภาษีมูลค่าเพิ่ม 7%' และถูกหัก ณ ที่จ่าย 3%",
      "เงินประกันการเช่า (Deposit) ไม่ถือเป็นรายได้จนกว่าจะริบมัดจำ",
    ],
    recommendedAccounts: REAL_ESTATE_ACCOUNTS,
  },
  transport_logistics: {
    type: "transport_logistics",
    titleTh: "ขนส่งและโลจิสติกส์ / คลังสินค้า",
    titleEn: "Transport & Logistics",
    descriptionTh: "กิจการรับจ้างขนส่งสินค้าทางบก บริการคลังสินค้า มีค่าน้ำมัน ค่าทางด่วน และค่าบำรุงรักษารถ",
    taxCharacteristicsTh: [
      "ค่าบริการขนส่งในราชอาณาจักร 'ยกเว้นภาษีมูลค่าเพิ่ม' ตามมาตรา 81(1)(ณ) และถูกหัก ณ ที่จ่าย 1%",
      "ค่าบริการคลังสินค้าและการยกขน 'มีภาษีมูลค่าเพิ่ม 7%' และถูกหัก ณ ที่จ่าย 3%",
      "ค่าน้ำมันเชื้อเพลิงมีภาษีซื้อ 7% นำมาหักได้, ค่าทางด่วน Easy Pass ไม่มี VAT",
    ],
    recommendedAccounts: TRANSPORT_ACCOUNTS,
  },
};

/** คลังแม่แบบสมุดรายวันมาตรฐาน 8 ธุรกิจ */
export const THAI_JOURNAL_PATTERNS: ThaiJournalPatternTemplate[] = [
  // --- 1. TRADING ---
  {
    id: "trading_purchase_credit",
    businessType: "trading",
    titleTh: "ซื้อสินค้าเป็นเงินเชื่อ (มีภาษีซื้อ 7%)",
    titleEn: "Purchase Merchandise on Credit with 7% VAT",
    keywords: ["ซื้อสินค้า", "ซื้อของ", "เจ้าหนี้การค้า", "บิลซื้อ", "ซื้อเชื่อ", "purchase goods"],
    taxNotesTh: "ได้รับใบกำกับภาษีแบบเต็มรูป นำภาษีซื้อ 7% มาหักในแบบ ภ.พ.30 ได้ทันที",
    tfrsReference: "TFRS for NPAEs บทที่ 8 สินค้าคงเหลือ",
    lines: [
      { accountCode: "5111", accountNameTh: "ซื้อสินค้า", side: "debit", formula: "base", descriptionTh: "มูลค่าสินค้าก่อนภาษี" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "vat7", descriptionTh: "ภาษีมูลค่าเพิ่ม 7%" },
      { accountCode: "2111", accountNameTh: "เจ้าหนี้การค้า", side: "credit", formula: "net_payable", descriptionTh: "ตั้งหนี้การค้าเต็มจำนวน" },
    ],
  },
  {
    id: "trading_sale_credit",
    businessType: "trading",
    titleTh: "ขายสินค้าเป็นเงินเชื่อ (ออกใบกำกับภาษีขาย 7%)",
    titleEn: "Credit Sales with 7% VAT",
    keywords: ["ขายสินค้า", "ขายเชื่อ", "ลูกหนี้การค้า", "ส่งมอบสินค้า", "ออกบิลขาย", "credit sale"],
    taxNotesTh: "จุดรับรู้ภาษีขายเกิดขึ้นเมื่อส่งมอบสินค้าหรือออกใบกำกับภาษี นำส่ง ภ.พ.30 ในเดือนที่ออกบิล",
    tfrsReference: "TFRS for NPAEs บทที่ 18 รายได้",
    lines: [
      { accountCode: "1131", accountNameTh: "ลูกหนี้การค้า", side: "debit", formula: "net_receivable", descriptionTh: "ยอดลูกหนี้รวมภาษี" },
      { accountCode: "4111", accountNameTh: "รายได้จากการขายสินค้า", side: "credit", formula: "base", descriptionTh: "มูลค่าขายก่อนภาษี" },
      { accountCode: "2141", accountNameTh: "ภาษีขาย", side: "credit", formula: "vat7", descriptionTh: "ภาษีมูลค่าเพิ่ม 7%" },
    ],
  },
  {
    id: "trading_sale_cash",
    businessType: "trading",
    titleTh: "ขายสินค้าเป็นเงินสด/เงินโอน (ออกใบกำกับภาษีขาย 7%)",
    titleEn: "Cash/Bank Sales with 7% VAT",
    keywords: ["ขายสด", "รับเงินสดขายของ", "ขายหน้าร้าน", "เงินโอนขายสินค้า", "cash sale"],
    taxNotesTh: "ภาษีขายรับรู้ทันทีเมื่อได้รับเงิน นำส่ง ภ.พ.30",
    lines: [
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "debit", formula: "net_receivable", descriptionTh: "รับเงินโอนสุทธิ" },
      { accountCode: "4111", accountNameTh: "รายได้จากการขายสินค้า", side: "credit", formula: "base", descriptionTh: "มูลค่าสินค้าก่อนภาษี" },
      { accountCode: "2141", accountNameTh: "ภาษีขาย", side: "credit", formula: "vat7", descriptionTh: "ภาษีขาย 7%" },
    ],
  },

  // --- 2. SERVICE ---
  {
    id: "service_income_wht3",
    businessType: "service",
    titleTh: "รับชำระค่าบริการ (ออกใบเสร็จ/ใบกำกับภาษี ถูกหัก WHT 3%)",
    titleEn: "Service Revenue with 7% VAT and 3% WHT Deducted",
    keywords: ["ค่าบริการ", "รับจ้าง", "ค่าที่ปรึกษา", "ถูกหัก 3%", "หัก ณ ที่จ่าย 3%", "service income"],
    taxNotesTh: "ออกใบกำกับภาษีขาย 7% และได้รับหนังสือรับรองหัก ณ ที่จ่าย 50 ทวิ (3%) นำไปเครดิตภาษีเงินได้นิติบุคคลตอนสิ้นปี (ภ.ง.ด.50/51)",
    lines: [
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "debit", formula: "custom", descriptionTh: "เงินที่ได้รับสุทธิ (ฐาน + VAT 7% - WHT 3%)" },
      { accountCode: "1161", accountNameTh: "ภาษีถูกหัก ณ ที่จ่าย (3%)", side: "debit", formula: "wht3", descriptionTh: "ภาษีถูกหัก ณ ที่จ่าย 3% ของฐาน" },
      { accountCode: "4121", accountNameTh: "รายได้ค่าบริการและที่ปรึกษา", side: "credit", formula: "base", descriptionTh: "มูลค่าค่าบริการก่อนภาษี" },
      { accountCode: "2141", accountNameTh: "ภาษีขาย", side: "credit", formula: "vat7", descriptionTh: "ภาษีขาย 7%" },
    ],
  },
  {
    id: "service_office_rent",
    businessType: "service",
    titleTh: "จ่ายค่าเช่าสำนักงาน (ยกเว้น VAT หัก ณ ที่จ่าย 5%)",
    titleEn: "Office Rent Payment (VAT Exempt, 5% WHT Deducted)",
    keywords: ["ค่าเช่า", "ค่าเช่าสำนักงาน", "ค่าเช่าออฟฟิศ", "ค่าเช่าตึก", "หัก 5%", "office rent"],
    taxNotesTh: "ค่าเช่าอสังหาริมทรัพย์ได้รับยกเว้นภาษีมูลค่าเพิ่มตาม ม.81(1)(ต) แต่ต้องหักภาษี ณ ที่จ่าย 5% ออกหนังสือรับรอง 50 ทวิ และนำส่ง ภ.ง.ด.53/3",
    lines: [
      { accountCode: "5321", accountNameTh: "ค่าเช่าสำนักงานและสถานที่", side: "debit", formula: "base", descriptionTh: "ค่าเช่าตามสัญญา" },
      { accountCode: "2151", accountNameTh: "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย (ภ.ง.ด.3/53)", side: "credit", formula: "wht5", descriptionTh: "หัก ณ ที่จ่าย 5%" },
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "credit", formula: "custom", descriptionTh: "จ่ายเงินสุทธิ 95%" },
    ],
  },
  {
    id: "service_subcontract_wht3",
    businessType: "service",
    titleTh: "จ่ายค่าจ้างทำของ/ฟรีแลนซ์/เอาท์ซอร์ส (มี VAT 7% หัก WHT 3%)",
    titleEn: "Subcontractor / Freelance Payment with 7% VAT and 3% WHT",
    keywords: ["จ้างทำของ", "ค่าจ้างทำของ", "ฟรีแลนซ์", "outsource", "ค่าโปรแกรมเมอร์", "subcontract"],
    taxNotesTh: "มีภาษีซื้อ 7% และต้องหักภาษี ณ ที่จ่าย 3% นำส่ง ภ.ง.ด.53 (นิติบุคคล) หรือ ภ.ง.ด.3 (บุคคลธรรมดา)",
    lines: [
      { accountCode: "5311", accountNameTh: "ค่าจ้างทำของและบริการภายนอก", side: "debit", formula: "base", descriptionTh: "มูลค่าค่าจ้างก่อนภาษี" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "vat7", descriptionTh: "ภาษีซื้อ 7%" },
      { accountCode: "2151", accountNameTh: "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย (ภ.ง.ด.3/53)", side: "credit", formula: "wht3", descriptionTh: "หัก ณ ที่จ่าย 3%" },
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "credit", formula: "custom", descriptionTh: "จ่ายสุทธิ (ฐาน + VAT 7% - WHT 3%)" },
    ],
  },

  // --- 3. MANUFACTURING ---
  {
    id: "mfg_buy_raw_material",
    businessType: "manufacturing",
    titleTh: "ซื้อวัตถุดิบเข้าโรงงาน (มีภาษีซื้อ 7%)",
    titleEn: "Purchase Direct Raw Materials with 7% VAT",
    keywords: ["ซื้อวัตถุดิบ", "วัตถุดิบ", "raw material", "สั่งซื้อวัตถุดิบโรงงาน", "แผ่นเหล็ก", "เม็ดพลาสติก"],
    taxNotesTh: "ภาษีซื้อนำส่ง ภ.พ.30 ตรวจสอบใบกำกับภาษีต้องระบุชื่อผู้ซื้อและเลขประจำตัวผู้เสียภาษีถูกต้อง",
    lines: [
      { accountCode: "1142", accountNameTh: "วัตถุดิบคงเหลือ", side: "debit", formula: "base", descriptionTh: "มูลค่าวัตถุดิบเข้าคลัง" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "vat7", descriptionTh: "ภาษีซื้อ 7%" },
      { accountCode: "2111", accountNameTh: "เจ้าหนี้การค้า", side: "credit", formula: "net_payable", descriptionTh: "ยอดหนี้รวมภาษี" },
    ],
  },
  {
    id: "mfg_issue_material_wip",
    businessType: "manufacturing",
    titleTh: "เบิกวัตถุดิบเข้าสู่สายการผลิต (WIP)",
    titleEn: "Issue Raw Materials into Work in Process (WIP)",
    keywords: ["เบิกวัตถุดิบ", "เบิกผลิต", "งานระหว่างทำ", "wip", "issue materials to production"],
    taxNotesTh: "รายการภายใน ไม่กระทบภาษีมูลค่าเพิ่ม แต่ต้องมีใบเบิกวัตถุดิบเป็นเอกสารประกอบการลงบัญชี",
    lines: [
      { accountCode: "1143", accountNameTh: "งานระหว่างทำ (WIP)", side: "debit", formula: "base", descriptionTh: "โอนเข้าต้นทุนงานระหว่างทำ" },
      { accountCode: "1142", accountNameTh: "วัตถุดิบคงเหลือ", side: "credit", formula: "base", descriptionTh: "ตัดยอดวัตถุดิบคงเหลือ" },
    ],
  },
  {
    id: "mfg_complete_finished_goods",
    businessType: "manufacturing",
    titleTh: "โอนปิดงานระหว่างทำเข้าสินค้าสำเร็จรูป (FG)",
    titleEn: "Transfer Completed WIP to Finished Goods (FG)",
    keywords: ["ผลิตเสร็จ", "โอนเข้าสินค้าสำเร็จรูป", "finished goods", "fg", "งานเสร็จสมบูรณ์"],
    taxNotesTh: "คำนวณต้นทุนการผลิตจริงตามรายงานสรุปต้นทุนการผลิต (Cost Sheet)",
    lines: [
      { accountCode: "1144", accountNameTh: "สินค้าสำเร็จรูป (FG)", side: "debit", formula: "base", descriptionTh: "มูลค่าสินค้าที่ผลิตเสร็จ" },
      { accountCode: "1143", accountNameTh: "งานระหว่างทำ (WIP)", side: "credit", formula: "base", descriptionTh: "ปิดยอดงานระหว่างทำ" },
    ],
  },

  // --- 4. RESTAURANT & CAFE ---
  {
    id: "rest_buy_fresh_market",
    businessType: "restaurant_cafe",
    titleTh: "ซื้อของสด ผัก ผลไม้ เนื้อสัตว์ (ยกเว้น VAT)",
    titleEn: "Purchase Fresh Market Groceries (VAT Exempt)",
    keywords: ["ซื้อของสด", "ผัก", "ผลไม้", "เนื้อหมู", "ไก่", "ตลาดสด", "วัตถุดิบอาหารสด", "fresh food"],
    taxNotesTh: "พืชผลทางการเกษตรและเนื้อสัตว์สดได้รับยกเว้นภาษีมูลค่าเพิ่มตาม ม.81(1)(ก) ไม่มีภาษีซื้อ",
    lines: [
      { accountCode: "1145", accountNameTh: "วัตถุดิบสด/เนื้อสัตว์/ผัก (ไม่มี VAT)", side: "debit", formula: "base", descriptionTh: "มูลค่าของสดที่จ่ายจริง" },
      { accountCode: "1111", accountNameTh: "เงินสดในลิ้นชัก/หน้าร้าน", side: "credit", formula: "base", descriptionTh: "จ่ายเงินสดซื้อของ" },
    ],
  },
  {
    id: "rest_daily_dinein_sales",
    businessType: "restaurant_cafe",
    titleTh: "สรุปยอดขายหน้าร้านประจำวัน (เงินสด/เงินโอน QR + VAT 7%)",
    titleEn: "Daily Dine-in Sales Summary with 7% VAT",
    keywords: ["ยอดขายหน้าร้าน", "ปิดกะ", "สรุปยอดขายประจำวัน", "ขายอาหาร", "คาเฟ่", "dine in sales"],
    taxNotesTh: "ออกใบกำกับภาษีอย่างย่อจากเครื่อง POS และสรุปรายงานภาษีขายประจำวันเพื่อนำส่ง ภ.พ.30",
    lines: [
      { accountCode: "1111", accountNameTh: "เงินสดในลิ้นชัก/หน้าร้าน", side: "debit", formula: "custom", descriptionTh: "เงินสดรับหน้าร้าน" },
      { accountCode: "1113", accountNameTh: "เงินโอนรับชำระผ่าน QR Code", side: "debit", formula: "custom", descriptionTh: "เงินโอนผ่าน QR PromptPay" },
      { accountCode: "4131", accountNameTh: "รายได้จากการขายอาหารและเครื่องดื่มหน้าร้าน", side: "credit", formula: "base", descriptionTh: "ยอดขายอาหารก่อนภาษี" },
      { accountCode: "2141", accountNameTh: "ภาษีขาย", side: "credit", formula: "vat7", descriptionTh: "ภาษีขาย 7%" },
    ],
  },
  {
    id: "rest_delivery_settlement",
    businessType: "restaurant_cafe",
    titleTh: "กระทบยอดเงินโอนเดลิเวอรี (Grab/Lineman หัก GP 30% + VAT 7% และ WHT 3%)",
    titleEn: "Delivery Platform Settlement (Grab/Lineman GP 30% + VAT, WHT 3%)",
    keywords: ["grab", "lineman", "foodpanda", "เดลิเวอรี", "gp grab", "gp lineman", "delivery settlement"],
    taxNotesTh: "ร้านต้องรับรู้รายได้ยอดขายเต็มจำนวน ส่วนค่า GP ถือเป็นค่าใช้จ่ายบริการที่มีภาษีซื้อ 7% และหัก ณ ที่จ่าย 3%",
    lines: [
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "debit", formula: "custom", descriptionTh: "เงินที่แอปโอนเข้าบัญชีจริง" },
      { accountCode: "5331", accountNameTh: "ค่าคอมมิชชันและค่า GP เดลิเวอรี", side: "debit", formula: "custom", descriptionTh: "ค่า GP 30% ก่อนภาษี" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "custom", descriptionTh: "ภาษีซื้อ 7% ของค่า GP (ตามใบกำกับแอป)" },
      { accountCode: "2151", accountNameTh: "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย (ภ.ง.ด.3/53)", side: "credit", formula: "custom", descriptionTh: "หัก ณ ที่จ่าย 3% ของค่า GP" },
      { accountCode: "1132", accountNameTh: "ลูกหนี้แพลตฟอร์มเดลิเวอรี (Grab/Lineman/Foodpanda)", side: "credit", formula: "net_receivable", descriptionTh: "ตัดยอดลูกหนี้เดลิเวอรีเต็มจำนวน" },
    ],
  },

  // --- 5. CONSTRUCTION ---
  {
    id: "const_progress_billing",
    businessType: "construction",
    titleTh: "วางบิลค่างวดงานก่อสร้าง (Progress Billing + VAT 7% หักเงินประกัน 10% และ WHT 3%)",
    titleEn: "Progress Billing with Retention (10%) and 3% WHT",
    keywords: ["ค่างวดงาน", "progress billing", "ส่งงวดงาน", "รับเหมาก่อสร้าง", "เงินประกันผลงาน", "retention"],
    taxNotesTh: "ภาษีขายรับรู้เมื่อส่งมอบงานหรือออกใบแจ้งหนี้/ใบเสร็จรับเงิน นิติบุคคลผู้ว่าจ้างจะหักเงินประกันผลงานและหัก WHT 3%",
    tfrsReference: "TFRS for NPAEs บทที่ 19 สัญญาก่อสร้าง",
    lines: [
      { accountCode: "1131", accountNameTh: "ลูกหนี้การค้า", side: "debit", formula: "custom", descriptionTh: "ยอดเงินที่จะได้รับจริงในงวดนี้" },
      { accountCode: "1133", accountNameTh: "เงินประกันผลงานค้างรับ (Retention Receivable)", side: "debit", formula: "custom", descriptionTh: "เงินประกันผลงานหักไว้ (ปกติ 5-10%)" },
      { accountCode: "1161", accountNameTh: "ภาษีถูกหัก ณ ที่จ่าย (3%)", side: "debit", formula: "wht3", descriptionTh: "ภาษีถูกหัก ณ ที่จ่าย 3% ของค่างวด" },
      { accountCode: "4141", accountNameTh: "รายได้ตามสัญญาก่อสร้าง", side: "credit", formula: "base", descriptionTh: "มูลค่างวดงานก่อนภาษี" },
      { accountCode: "2141", accountNameTh: "ภาษีขาย", side: "credit", formula: "vat7", descriptionTh: "ภาษีขาย 7%" },
    ],
  },
  {
    id: "const_buy_materials",
    businessType: "construction",
    titleTh: "ซื้อวัสดุก่อสร้าง/ปูน/เหล็ก เข้าหน้างาน (มีภาษีซื้อ 7%)",
    titleEn: "Purchase Construction Materials with 7% VAT",
    keywords: ["ซื้อปูน", "ซื้อเหล็ก", "วัสดุก่อสร้าง", "กระเบื้อง", "คอนกรีตผสมเสร็จ", "construction materials"],
    taxNotesTh: "บันทึกเข้าต้นทุนงานก่อสร้างโดยตรง (Cost of Construction) พร้อมนำภาษีซื้อ 7% มาหักใน ภ.พ.30",
    lines: [
      { accountCode: "5131", accountNameTh: "ต้นทุนค่าวัสดุก่อสร้าง", side: "debit", formula: "base", descriptionTh: "ค่าวัสดุก่อสร้างก่อนภาษี" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "vat7", descriptionTh: "ภาษีซื้อ 7%" },
      { accountCode: "2111", accountNameTh: "เจ้าหนี้การค้า", side: "credit", formula: "net_payable", descriptionTh: "เจ้าหนี้ร้านวัสดุก่อสร้าง" },
    ],
  },

  // --- 6. E-COMMERCE ---
  {
    id: "ecom_marketplace_payout",
    businessType: "ecommerce",
    titleTh: "รับเงินโอน Settlement จาก Shopee / Lazada / TikTok (หักค่าธรรมเนียม + VAT)",
    titleEn: "Marketplace Settlement Payout (Shopee/Lazada/TikTok)",
    keywords: ["shopee", "lazada", "tiktok shop", "settlement", "เงินเข้า shopee", "ค่าธรรมเนียม shopee"],
    taxNotesTh: "กระทบยอดเงินโอนเข้าบัญชีธนาคารกับยอดขายจริง แพลตฟอร์มจะออกใบกำกับภาษีสำหรับค่า Commission/Transaction ให้เป็นภาษีซื้อ 7%",
    lines: [
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "debit", formula: "custom", descriptionTh: "ยอดเงินโอนสุทธิเข้าบัญชี" },
      { accountCode: "5341", accountNameTh: "ค่าธรรมเนียมธุรกรรมการชำระเงิน (Transaction Fee)", side: "debit", formula: "custom", descriptionTh: "ค่าธรรมเนียมการชำระเงิน" },
      { accountCode: "5342", accountNameTh: "ค่าคอมมิชชันและค่าบริการมาร์เก็ตเพลส", side: "debit", formula: "custom", descriptionTh: "ค่าคอมมิชชันแพลตฟอร์ม" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "custom", descriptionTh: "ภาษีซื้อ 7% จากใบกำกับแพลตฟอร์ม" },
      { accountCode: "1134", accountNameTh: "ลูกหนี้ Shopee / Lazada / TikTok Shop", side: "credit", formula: "net_receivable", descriptionTh: "ตัดยอดลูกหนี้แพลตฟอร์ม" },
    ],
  },
  {
    id: "ecom_facebook_ads_pp36",
    businessType: "ecommerce",
    titleTh: "จ่ายค่ายิงแอด Facebook / Google / TikTok (ต่างประเทศ นำส่ง ภ.พ.36)",
    titleEn: "Overseas Ad Spend (Facebook/Google) with ภ.พ.36 VAT Remittance",
    keywords: ["ยิงแอด", "ค่าแอด", "facebook ads", "google ads", "tiktok ads", "โฆษณาออนไลน์", "ภ.พ.36"],
    taxNotesTh: "การจ่ายเงินค่าบริการไปต่างประเทศ ต้องนำส่งภาษีมูลค่าเพิ่มแทนผู้ให้บริการต่างประเทศด้วยแบบ ภ.พ.36 (7%) ภายในวันที่ 7 ของเดือนถัดไป และนำใบเสร็จ ภ.พ.36 มาเคลมเป็นภาษีซื้อในเดือนถัดไปได้",
    lines: [
      { accountCode: "5344", accountNameTh: "ค่าโฆษณายิงแอด Facebook / TikTok / Google Ads", side: "debit", formula: "base", descriptionTh: "ยอดตัดบัตรเครดิตค่าแอด" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "vat7", descriptionTh: "สิทธิ์ภาษีซื้อจากการยื่น ภ.พ.36" },
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "credit", formula: "base", descriptionTh: "ยอดเงินตัดบัญชี/บัตรเครดิต" },
      { accountCode: "2143", accountNameTh: "ภาษีมูลค่าเพิ่มรอนำส่ง ภ.พ.36 (ยิงแอดต่างประเทศ)", side: "credit", formula: "vat7", descriptionTh: "ตั้งหนี้ภาษีรอนำส่ง ภ.พ.36" },
    ],
  },

  // --- 7. REAL ESTATE RENTAL ---
  {
    id: "rental_monthly_billing",
    businessType: "real_estate_rental",
    titleTh: "เรียกเก็บค่าเช่าอาคารและค่าบริการส่วนกลาง (ค่าเช่า ม.81(1)(ต) + ค่าบริการมี VAT 7%)",
    titleEn: "Monthly Rental & Common Facility Billing (Exempt Rent + 7% VAT Service)",
    keywords: ["ค่าเช่าตึก", "ค่าบริการส่วนกลาง", "เก็บค่าเช่า", "สัญญาเช่าอาคาร", "ผู้เช่า", "rental billing"],
    taxNotesTh: "ค่าเช่าได้รับยกเว้น VAT ตาม ม.81(1)(ต) ถูกหัก WHT 5% ส่วนค่าบริการส่วนกลางมี VAT 7% และถูกหัก WHT 3%",
    lines: [
      { accountCode: "1131", accountNameTh: "ลูกหนี้การค้า", side: "debit", formula: "custom", descriptionTh: "ยอดเรียกเก็บรวมสุทธิ" },
      { accountCode: "1162", accountNameTh: "ภาษีถูกหัก ณ ที่จ่ายค่าเช่า (5%)", side: "debit", formula: "custom", descriptionTh: "หัก ณ ที่จ่าย 5% ของค่าเช่า" },
      { accountCode: "1161", accountNameTh: "ภาษีถูกหัก ณ ที่จ่าย (3%)", side: "debit", formula: "custom", descriptionTh: "หัก ณ ที่จ่าย 3% ของค่าบริการส่วนกลาง" },
      { accountCode: "4161", accountNameTh: "รายได้ค่าเช่าอสังหาริมทรัพย์ (ยกเว้น VAT ม.81(1)(ต))", side: "credit", formula: "custom", descriptionTh: "รายได้ค่าเช่า (ยกเว้น VAT)" },
      { accountCode: "4162", accountNameTh: "รายได้ค่าบริการส่วนกลางและสาธารณูปโภค (มี VAT 7%)", side: "credit", formula: "custom", descriptionTh: "รายได้ค่าบริการส่วนกลาง" },
      { accountCode: "2141", accountNameTh: "ภาษีขาย", side: "credit", formula: "custom", descriptionTh: "ภาษีขาย 7% ของค่าบริการส่วนกลาง" },
    ],
  },
  {
    id: "rental_security_deposit",
    businessType: "real_estate_rental",
    titleTh: "รับเงินมัดจำ/เงินประกันการเช่า (Deposit ไม่เสียภาษี)",
    titleEn: "Receive Tenant Security Deposit (Non-Taxable)",
    keywords: ["เงินมัดจำ", "เงินประกันการเช่า", "มัดจำห้อง", "เงินประกันห้อง", "security deposit"],
    taxNotesTh: "เงินประกันความเสียหายหรือเงินมัดจำการเช่าที่จะต้องคืนเมื่อสิ้นสุดสัญญา ไม่ถือเป็นรายได้และไม่เสียภาษีมูลค่าเพิ่ม",
    lines: [
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "debit", formula: "base", descriptionTh: "รับเงินมัดจำเข้าบัญชี" },
      { accountCode: "2161", accountNameTh: "เงินประกันและมัดจำการเช่ารับล่วงหน้า", side: "credit", formula: "base", descriptionTh: "ตั้งเป็นหนี้สินรอคืนเมื่อหมดสัญญา" },
    ],
  },

  // --- 8. TRANSPORT & LOGISTICS ---
  {
    id: "trans_freight_service",
    businessType: "transport_logistics",
    titleTh: "ให้บริการขนส่งสินค้าทางบก (ยกเว้น VAT ม.81(1)(ณ) ถูกหัก WHT 1%)",
    titleEn: "Freight Transportation Service (VAT Exempt, 1% WHT Deducted)",
    keywords: ["ค่าขนส่ง", "รับจ้างขนส่ง", "วิ่งงานขนส่ง", "ค่าเที่ยวรถ", "หัก 1%", "freight transport"],
    taxNotesTh: "การให้บริการขนส่งในราชอาณาจักรได้รับยกเว้นภาษีมูลค่าเพิ่มตาม ม.81(1)(ณ) แต่ลูกค้านิติบุคคลต้องหักภาษี ณ ที่จ่าย 1%",
    lines: [
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "debit", formula: "custom", descriptionTh: "เงินรับโอนสุทธิ (99%)" },
      { accountCode: "1163", accountNameTh: "ภาษีถูกหัก ณ ที่จ่ายค่าขนส่ง (1%)", side: "debit", formula: "wht1", descriptionTh: "ภาษีถูกหัก ณ ที่จ่าย 1%" },
      { accountCode: "4171", accountNameTh: "รายได้ค่าบริการขนส่งทางบก (ยกเว้น VAT ม.81(1)(ณ))", side: "credit", formula: "base", descriptionTh: "รายได้ค่าขนส่งตามใบเสร็จ" },
    ],
  },
  {
    id: "trans_fuel_fleet_card",
    businessType: "transport_logistics",
    titleTh: "เติมน้ำมันรถบรรทุก/ฟลีทการ์ด (มีภาษีซื้อ 7%)",
    titleEn: "Fleet Fuel Expense with 7% Input VAT",
    keywords: ["ค่าน้ำมัน", "เติมน้ำมัน", "ดีเซล", "fleet card", "ปั๊มน้ำมัน", "fuel expense"],
    taxNotesTh: "ใบกำกับภาษีค่าน้ำมันรถบรรทุกขนส่งสินค้าสามารถนำภาษีซื้อ 7% มาหักใน ภ.พ.30 ได้ 100% (ต่างจากรถยนต์นั่งไม่เกิน 10 ที่นั่ง)",
    lines: [
      { accountCode: "5351", accountNameTh: "ค่าน้ำมันเชื้อเพลิงและก๊าซ NGV/LPG", side: "debit", formula: "base", descriptionTh: "ค่าน้ำมันก่อนภาษี" },
      { accountCode: "1151", accountNameTh: "ภาษีซื้อ", side: "debit", formula: "vat7", descriptionTh: "ภาษีซื้อ 7%" },
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "credit", formula: "net_payable", descriptionTh: "จ่ายเงินค่าน้ำมันรวมภาษี" },
    ],
  },
  {
    id: "trans_toll_expressway",
    businessType: "transport_logistics",
    titleTh: "จ่ายค่าทางด่วนและมอเตอร์เวย์ (Easy Pass / M-Flow ไม่มี VAT)",
    titleEn: "Toll & Expressway Fees (Easy Pass / M-Flow, No VAT)",
    keywords: ["ค่าทางด่วน", "easy pass", "m-flow", "มอเตอร์เวย์", "ทางพิเศษ", "toll fee"],
    taxNotesTh: "ค่าผ่านทางพิเศษของการทางพิเศษแห่งประเทศไทยได้รับยกเว้นภาษีมูลค่าเพิ่ม ไม่มีภาษีซื้อ",
    lines: [
      { accountCode: "5352", accountNameTh: "ค่าผ่านทางด่วนและมอเตอร์เวย์ (Easy Pass)", side: "debit", formula: "base", descriptionTh: "ค่าทางด่วนตามใบรับเงิน" },
      { accountCode: "1112", accountNameTh: "เงินฝากกระแสรายวัน/ออมทรัพย์", side: "credit", formula: "base", descriptionTh: "ตัดยอดเงินสด/บัตรเติมเงิน" },
    ],
  },
];

/** ผลลัพธ์การคำนวณและข้อเสนอแนะผังบัญชี */
export interface JournalSuggestionResult {
  pattern: ThaiJournalPatternTemplate;
  matchScore: number;
  matchedKeywords: string[];
  calculatedLines: Array<{
    accountCode: string;
    accountNameTh: string;
    side: "debit" | "credit";
    amount: number;
    descriptionTh: string;
  }>;
}

/**
 * คำนวณภาษีมูลค่าเพิ่ม 7% แบบ Satang ป้องกันปัญหาเศษทศนิยม
 */
export function calculateVatSatang(amountSatang: bigint, ratePercent = 7): bigint {
  return (amountSatang * BigInt(ratePercent) + 50n) / 100n;
}

/**
 * คำนวณภาษีหัก ณ ที่จ่าย Satang ตามอัตรา WHT
 */
export function calculateWhtSatang(baseSatang: bigint, whtPercent: number): bigint {
  return (baseSatang * BigInt(whtPercent) + 50n) / 100n;
}

/**
 * Smart Journal Suggestion: ค้นหาและแนะนำรูปแบบการลงบัญชีที่ตรงกับคำอธิบายรายการ
 */
export function suggestJournalPatterns(
  query: string,
  businessType?: ThaiBusinessType,
  baseAmount = 0,
): JournalSuggestionResult[] {
  const cleanQuery = query.toLowerCase().trim();
  if (!cleanQuery) return [];

  const candidates = businessType
    ? THAI_JOURNAL_PATTERNS.filter((p) => p.businessType === businessType)
    : THAI_JOURNAL_PATTERNS;

  const results: JournalSuggestionResult[] = [];

  for (const pattern of candidates) {
    const matched: string[] = [];
    let score = 0;

    // เช็ค title
    if (pattern.titleTh.toLowerCase().includes(cleanQuery)) {
      score += 10;
    }

    // เช็ค keywords
    for (const kw of pattern.keywords) {
      const lowerKw = kw.toLowerCase();
      if (cleanQuery.includes(lowerKw)) {
        matched.push(kw);
        score += 5;
      } else if (lowerKw.includes(cleanQuery)) {
        matched.push(kw);
        score += 2;
      }
    }

    if (score > 0) {
      // คำนวณยอดเงินตามสูตรถ้ามีระบุ baseAmount
      const baseSatang = BigInt(Math.round(baseAmount * 100));
      const vatSatang = calculateVatSatang(baseSatang, 7);
      const wht1Satang = calculateWhtSatang(baseSatang, 1);
      const wht2Satang = calculateWhtSatang(baseSatang, 2);
      const wht3Satang = calculateWhtSatang(baseSatang, 3);
      const wht5Satang = calculateWhtSatang(baseSatang, 5);

      const calculatedLines = pattern.lines.map((line) => {
        let amount = baseAmount;
        if (line.formula === "vat7") {
          amount = Number(vatSatang) / 100;
        } else if (line.formula === "wht1") {
          amount = Number(wht1Satang) / 100;
        } else if (line.formula === "wht2") {
          amount = Number(wht2Satang) / 100;
        } else if (line.formula === "wht3") {
          amount = Number(wht3Satang) / 100;
        } else if (line.formula === "wht5") {
          amount = Number(wht5Satang) / 100;
        } else if (line.formula === "net_payable") {
          amount = Number(baseSatang + vatSatang) / 100;
        } else if (line.formula === "net_receivable") {
          amount = Number(baseSatang + vatSatang) / 100;
        }
        return {
          accountCode: line.accountCode,
          accountNameTh: line.accountNameTh,
          side: line.side,
          amount,
          descriptionTh: line.descriptionTh,
        };
      });

      results.push({
        pattern,
        matchScore: score,
        matchedKeywords: matched,
        calculatedLines,
      });
    }
  }

  // เรียงลำดับจากคะแนนสูงสุดลงมา
  return results.sort((a, b) => b.matchScore - a.matchScore);
}

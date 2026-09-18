// Thai Withholding Tax (WHT) Engine & Statutory Forms (ภ.ง.ด.3, ภ.ง.ด.53, 50 ทวิ)
// ตามประมวลรัษฎากร มาตรา 3 เตรส, มาตรา 50, มาตรา 50 ทวิ, และคำสั่งกรมสรรพากร ท.ป.4/2528

import { thaiBahtText } from "./thai-vat-closing";

export type ThaiWhtIncomeType =
  | "rent_40_5"                 // ค่าเช่าทรัพย์สิน / อาคาร (5%)
  | "service_subcontract_40_8"  // ค่าบริการ / รับจ้างทำของ (3%)
  | "transportation_40_8"       // ค่าขนส่ง (1%)
  | "advertising_40_8"          // ค่าโฆษณา (2%)
  | "professional_40_6"         // ค่าวิชาชีพอิสระ เช่น บัญชี/ทนาย (3%)
  | "contractor_40_7"           // ค่ารับเหมาก่อสร้าง (3%)
  | "interest_dividend_40_4"    // ดอกเบี้ย (15%) / เงินปันผล (10%)
  | "prize_reward_40_8"         // รางวัล / ชิงโชค (5%)
  | "other";

export interface ThaiWhtIncomeTypeConfig {
  code: ThaiWhtIncomeType;
  lawSection: string;
  nameTh: string;
  nameEn: string;
  defaultRate: number;
  revenueCodeTextTh: string;
}

export const THAI_WHT_INCOME_CONFIGS: Record<ThaiWhtIncomeType, ThaiWhtIncomeTypeConfig> = {
  rent_40_5: {
    code: "rent_40_5",
    lawSection: "มาตรา 40(5)",
    nameTh: "ค่าเช่าทรัพย์สิน / อาคาร / สถานที่",
    nameEn: "Rental of Property & Premises",
    defaultRate: 5,
    revenueCodeTextTh: "ค่าเช่าหรือประโยชน์อย่างอื่นที่ได้จากการให้เช่าทรัพย์สิน",
  },
  service_subcontract_40_8: {
    code: "service_subcontract_40_8",
    lawSection: "มาตรา 40(8)",
    nameTh: "ค่าบริการ / จ้างทำของ / ค่าธรรมเนียม",
    nameEn: "Service Fees & Subcontracting",
    defaultRate: 3,
    revenueCodeTextTh: "ค่าบริการ จ้างทำของ รางวัล ส่วนลดหรือประโยชน์ใดๆ เนื่องจากการส่งเสริมการขาย",
  },
  transportation_40_8: {
    code: "transportation_40_8",
    lawSection: "มาตรา 40(8)",
    nameTh: "ค่าบริการขนส่งสินค้า",
    nameEn: "Freight & Transportation Services",
    defaultRate: 1,
    revenueCodeTextTh: "ค่าจ้างขนส่งสินค้าหรือผู้โดยสาร (ยกเว้นเรือเดินทะเล)",
  },
  advertising_40_8: {
    code: "advertising_40_8",
    lawSection: "มาตรา 40(8)",
    nameTh: "ค่าโฆษณาและประชาสัมพันธ์",
    nameEn: "Advertising & Media Placements",
    defaultRate: 2,
    revenueCodeTextTh: "ค่าโฆษณาผ่านสื่อสิ่งพิมพ์ วิทยุ โทรทัศน์ หรือสื่อออนไลน์",
  },
  professional_40_6: {
    code: "professional_40_6",
    lawSection: "มาตรา 40(6)",
    nameTh: "วิชาชีพอิสระ (กฎหมาย / บัญชี / การแพทย์ / วิศวกรรม)",
    nameEn: "Liberal Professions (Legal, Accounting, etc.)",
    defaultRate: 3,
    revenueCodeTextTh: "การประกอบวิชาชีพอิสระ เช่น บัญชี ทนายความ แพทย์",
  },
  contractor_40_7: {
    code: "contractor_40_7",
    lawSection: "มาตรา 40(7)",
    nameTh: "การรับเหมาก่อสร้างและติดตั้ง",
    nameEn: "Construction & Contracting Works",
    defaultRate: 3,
    revenueCodeTextTh: "การรับเหมาที่ผู้รับเหมาต้องจัดหาสัมภาระในส่วนสำคัญ",
  },
  interest_dividend_40_4: {
    code: "interest_dividend_40_4",
    lawSection: "มาตรา 40(4)",
    nameTh: "ดอกเบี้ย / เงินปันผล / ส่วนแบ่งกำไร",
    nameEn: "Interest & Dividends",
    defaultRate: 10,
    revenueCodeTextTh: "ดอกเบี้ยเงินกู้ยืม ดอกเบี้ยพันธบัตร เงินปันผล",
  },
  prize_reward_40_8: {
    code: "prize_reward_40_8",
    lawSection: "มาตรา 40(8)",
    nameTh: "รางวัลจากการประกวด / แข่งขัน / ชิงโชค",
    nameEn: "Prizes & Rewards",
    defaultRate: 5,
    revenueCodeTextTh: "รางวัลจากการประกวด แข่งขัน การชิงโชค หรือการอื่นใดอันมีลักษณะทำนองเดียวกัน",
  },
  other: {
    code: "other",
    lawSection: "มาตรา 40",
    nameTh: "เงินได้พึงประเมินอื่นๆ",
    nameEn: "Other Taxable Incomes",
    defaultRate: 3,
    revenueCodeTextTh: "เงินได้พึงประเมินอื่นๆ ตามประมวลรัษฎากร",
  },
};

export interface ThaiWhtRecord {
  id: string;
  docNo: string;              // เลขที่ใบรับรอง เช่น 50TWI-2026/09-001
  docDate: string;            // วันเดือนปีที่จ่าย (YYYY-MM-DD)
  filingType: "pnd3" | "pnd53";
  payeeType: "individual" | "corporate";
  payeeTaxId: string;         // เลขประจำตัว 13 หลัก
  payeeName: string;          // ชื่อผู้มีเงินได้
  payeeAddress: string;       // ที่อยู่ผู้มีเงินได้
  payeeBranchNo: string;      // สาขา (00000 = สนญ.)
  isHeadOffice: boolean;
  incomeType: ThaiWhtIncomeType;
  incomeDescription: string;  // คำอธิบายประเภทเงินได้
  taxRate: number;            // อัตราภาษี 1, 2, 3, 5
  paymentAmount: number;      // จำนวนเงินที่จ่าย (บาท)
  whtAmount: number;          // ภาษีที่หักและนำส่ง (บาท)
  condition: "deducted" | "forever" | "once"; // 1=หัก ณ ที่จ่าย, 2=ออกให้ตลอดไป, 3=ออกให้ครั้งเดียว
  bookCode?: string;          // สมุดรายวันอ้างอิง
  status: "active" | "cancelled";
}

export interface PayerInfo {
  taxId: string;
  nameTh: string;
  addressTh: string;
  branchNo: string;
  isHeadOffice: boolean;
}

export interface PndFilingSummary {
  formType: "pnd3" | "pnd53";
  periodMonth: number;
  periodYearBe: number; // พ.ศ.
  totalPayees: number;
  totalPaymentAmount: number;
  totalWhtAmount: number;
  totalWhtTextTh: string;
  byIncomeType: Array<{
    incomeType: ThaiWhtIncomeType;
    incomeNameTh: string;
    rate: number;
    count: number;
    paymentAmount: number;
    whtAmount: number;
  }>;
}

/**
 * คำนวณสรุปแบบยื่น ภ.ง.ด.3 หรือ ภ.ง.ด.53 ประจำเดือน
 */
export function computePndSummary(
  records: ThaiWhtRecord[],
  formType: "pnd3" | "pnd53",
  month: number,
  yearCe: number,
): PndFilingSummary {
  const activeRecords = records.filter(
    (r) => r.filingType === formType && r.status === "active",
  );

  let totalPaymentSatang = 0n;
  let totalWhtSatang = 0n;
  const mapByType = new Map<
    ThaiWhtIncomeType,
    { count: number; paymentSatang: bigint; whtSatang: bigint; rate: number }
  >();

  for (const r of activeRecords) {
    const pSatang = BigInt(Math.round(r.paymentAmount * 100));
    const wSatang = BigInt(Math.round(r.whtAmount * 100));
    totalPaymentSatang += pSatang;
    totalWhtSatang += wSatang;

    const existing = mapByType.get(r.incomeType) ?? {
      count: 0,
      paymentSatang: 0n,
      whtSatang: 0n,
      rate: r.taxRate,
    };
    existing.count += 1;
    existing.paymentSatang += pSatang;
    existing.whtSatang += wSatang;
    mapByType.set(r.incomeType, existing);
  }

  const byIncomeType = Array.from(mapByType.entries()).map(([code, val]) => ({
    incomeType: code,
    incomeNameTh: THAI_WHT_INCOME_CONFIGS[code]?.nameTh || code,
    rate: val.rate,
    count: val.count,
    paymentAmount: Number(val.paymentSatang) / 100,
    whtAmount: Number(val.whtSatang) / 100,
  }));

  const totalWht = Number(totalWhtSatang) / 100;

  return {
    formType,
    periodMonth: month,
    periodYearBe: yearCe + 543,
    totalPayees: activeRecords.length,
    totalPaymentAmount: Number(totalPaymentSatang) / 100,
    totalWhtAmount: totalWht,
    totalWhtTextTh: thaiBahtText(totalWht),
    byIncomeType,
  };
}

/**
 * โครงสร้างข้อมูลหนังสือรับรองการหักภาษี ณ ที่จ่ายตามมาตรา 50 ทวิ (50 Twi Certificate)
 */
export interface Certificate50TwiData {
  payer: PayerInfo;
  payee: {
    taxId: string;
    name: string;
    address: string;
    branchNo: string;
    isHeadOffice: boolean;
  };
  certNo: string;
  certDate: string;
  pndType: "pnd1" | "pnd2" | "pnd3" | "pnd53";
  items: Array<{
    incomeCategoryName: string;
    paymentDate: string;
    paymentAmount: number;
    whtAmount: number;
  }>;
  totalPayment: number;
  totalWht: number;
  totalWhtBahtText: string;
  conditionTextTh: string; // "หัก ณ ที่จ่าย" | "ออกให้ตลอดไป" | "ออกให้ครั้งเดียว"
}

/**
 * สร้างข้อมูลหนังสือรับรอง 50 ทวิ ที่สมบูรณ์พร้อมพิมพ์
 */
export function generate50TwiCertificate(
  record: ThaiWhtRecord,
  payer: PayerInfo,
): Certificate50TwiData {
  const config = THAI_WHT_INCOME_CONFIGS[record.incomeType];
  const incomeName = config ? `${config.lawSection} ${config.nameTh}` : record.incomeDescription;

  const conditionMap: Record<string, string> = {
    deducted: "(1) หัก ณ ที่จ่าย",
    forever: "(2) ออกให้ตลอดไป",
    once: "(3) ออกให้ครั้งเดียว",
  };

  return {
    payer,
    payee: {
      taxId: record.payeeTaxId,
      name: record.payeeName,
      address: record.payeeAddress,
      branchNo: record.payeeBranchNo,
      isHeadOffice: record.isHeadOffice,
    },
    certNo: record.docNo,
    certDate: record.docDate,
    pndType: record.filingType,
    items: [
      {
        incomeCategoryName: incomeName,
        paymentDate: record.docDate,
        paymentAmount: record.paymentAmount,
        whtAmount: record.whtAmount,
      },
    ],
    totalPayment: record.paymentAmount,
    totalWht: record.whtAmount,
    totalWhtBahtText: thaiBahtText(record.whtAmount),
    conditionTextTh: conditionMap[record.condition] || "(1) หัก ณ ที่จ่าย",
  };
}

// Official ภ.พ.30 Form Calculator (ข้อ 1 ถึง ข้อ 10 ตามแบบฟอร์มกรมสรรพากร)
// แยกออกมาจาก thai-vat-reconciliation.ts เมื่อลบส่วนกระทบยอด GL (Champ parity 2026-09-19)

export interface OfficialPp30FormData {
  // ข้อมูลส่วนที่ 1: สถานประกอบการ
  taxId: string;
  companyName: string;
  branchNo: string;
  isHeadOffice: boolean;
  taxPeriodMonth: number;
  taxPeriodYearBe: number; // พ.ศ.
  filingType: "normal" | "additional";
  additionalTimes?: number;

  // ข้อมูลส่วนที่ 2: การคำนวณภาษี ข้อ 1 ถึง ข้อ 10
  item1_grossSales: number;           // 1. ยอดขายเดือนนี้
  item2_zeroRatedSales: number;       // 2. ยอดขายอัตราร้อยละ 0
  item3_exemptSales: number;           // 3. ยอดขายได้รับยกเว้นภาษี
  item4_taxableSales: number;         // 4. ยอดขายที่ต้องเสียภาษี (1 - 2 - 3)
  item5_outputVat: number;            // 5. ภาษีขาย (ร้อยละ 7 ของข้อ 4)
  item6_taxablePurchases: number;     // 6. ยอดซื้อที่มีสิทธินำภาษีซื้อมาหัก
  item7_inputVat: number;             // 7. ภาษีซื้อตามใบกำกับภาษี
  item8_creditBroughtForward: number; // 8. ภาษีชำระเกินยกมาจากเดือนก่อน
  item9_vatPayable: number;           // 9. ภาษีที่ต้องชำระเดือนนี้ (ถ้า 5 > 7 + 8)
  item10_vatOverpaid: number;         // 10. ภาษีชำระเกินเดือนนี้ (ถ้า 7 + 8 > 5)

  // ส่วนที่ 3: เงินเพิ่ม/เบี้ยปรับ
  surcharge: number;                  // เงินเพิ่ม (1.5% ต่อเดือน)
  penalty: number;                    // เบี้ยปรับ (ถ้ามี)
  netPaymentAmount: number;           // รวมยอดภาษีที่ต้องชำระทั้งสิ้น
}

/**
 * คำนวณกระทบยอดภาษีซื้อ-ภาษีขาย ระหว่างบัญชีแยกประเภท (GL) กับรายงานภาษี (VAT Register)
 */
/**
 * คำนวณข้อมูลแบบ ภ.พ. 30 ตามแบบฟอร์มกรมสรรพากร ข้อ 1 ถึง ข้อ 10
 */
export function computeOfficialPp30(params: {
  taxId: string;
  companyName: string;
  branchNo?: string;
  month: number;
  yearCe: number; // ค.ศ. เช่น 2026
  taxableSales: number;
  zeroRatedSales?: number;
  exemptSales?: number;
  outputVat?: number;
  taxablePurchases: number;
  inputVat?: number;
  creditBroughtForward?: number;
  isLateFiling?: boolean;
  lateMonths?: number;
}): OfficialPp30FormData {
  const zeroRated = params.zeroRatedSales ?? 0;
  const exempt = params.exemptSales ?? 0;
  const grossSales = params.taxableSales + zeroRated + exempt;

  // ภาษีขาย: ถ้าส่งมาให้ใช้ค่าจริง ถ้าไม่ส่งมาให้คำนวณ 7% แบบ Satang
  let outputVat = params.outputVat;
  if (outputVat === undefined) {
    const satang = (BigInt(Math.round(params.taxableSales * 100)) * 7n + 50n) / 100n;
    outputVat = Number(satang) / 100;
  }

  // ภาษีซื้อ
  let inputVat = params.inputVat;
  if (inputVat === undefined) {
    const satang = (BigInt(Math.round(params.taxablePurchases * 100)) * 7n + 50n) / 100n;
    inputVat = Number(satang) / 100;
  }

  const creditForward = params.creditBroughtForward ?? 0;

  // Net calculation
  const totalCreditsSatang = BigInt(Math.round((inputVat + creditForward) * 100));
  const outputSatang = BigInt(Math.round(outputVat * 100));
  const diffSatang = outputSatang - totalCreditsSatang;

  let item9_vatPayable = 0;
  let item10_vatOverpaid = 0;

  if (diffSatang > 0n) {
    item9_vatPayable = Number(diffSatang) / 100;
  } else if (diffSatang < 0n) {
    item10_vatOverpaid = Number(-diffSatang) / 100;
  }

  // เงินเพิ่ม (Surcharge 1.5% ต่อเดือนของภาษีที่ต้องชำระ ตาม ม.27 แห่งประมวลรัษฎากร)
  let surcharge = 0;
  if (params.isLateFiling && item9_vatPayable > 0 && params.lateMonths && params.lateMonths > 0) {
    const surchargeSatang = (BigInt(Math.round(item9_vatPayable * 100)) * 15n * BigInt(params.lateMonths) + 500n) / 1000n;
    surcharge = Number(surchargeSatang) / 100;
  }

  const netPaymentAmount = item9_vatPayable + surcharge;
  const branchNo = params.branchNo || "00000";

  return {
    taxId: params.taxId,
    companyName: params.companyName,
    branchNo,
    isHeadOffice: branchNo === "00000" || branchNo === "",
    taxPeriodMonth: params.month,
    taxPeriodYearBe: params.yearCe + 543,
    filingType: "normal",
    item1_grossSales: grossSales,
    item2_zeroRatedSales: zeroRated,
    item3_exemptSales: exempt,
    item4_taxableSales: params.taxableSales,
    item5_outputVat: outputVat,
    item6_taxablePurchases: params.taxablePurchases,
    item7_inputVat: inputVat,
    item8_creditBroughtForward: creditForward,
    item9_vatPayable,
    item10_vatOverpaid,
    surcharge,
    penalty: 0,
    netPaymentAmount,
  };
}

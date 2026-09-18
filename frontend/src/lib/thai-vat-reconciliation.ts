// Thai VAT Reconciliation Engine & Official ภ.พ.30 Form Calculator
// พัฒนาตามประมวลรัษฎากร มาตรา 79, 81, 82, 87 และมาตรฐานการรายงานทางการเงินไทย TFRS for NPAEs

export interface GlVatBalances {
  /** ยอดเดบิตบัญชี 1151 ภาษีซื้อ ใน GL งวดนี้ (บาท) */
  inputVatDebit: number;
  /** ยอดเครดิตบัญชี 2141 ภาษีขาย ใน GL งวดนี้ (บาท) */
  outputVatCredit: number;
  /** ยอดยกมาภาษีซื้อค้างรับ/ชำระเกิน (ถ้ามี) */
  vatCreditBroughtForward?: number;
}

export interface VatRegisterTotals {
  taxableSalesBase: number;
  zeroRatedSalesBase?: number;
  exemptSalesBase?: number;
  salesVat: number;
  taxablePurchasesBase: number;
  purchasesVat: number;
}

export interface VatReconciliationItem {
  accountCode: string;
  accountNameTh: string;
  glAmount: number;
  registerAmount: number;
  variance: number;
  status: "matched" | "discrepancy";
  noteTh: string;
}

export interface VatReconciliationReport {
  periodMonth: number;
  periodYear: number;
  inputVat: VatReconciliationItem;
  outputVat: VatReconciliationItem;
  totalVariance: number;
  isFullyBalanced: boolean;
  statusMessageTh: string;
  recommendationsTh: string[];
}

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
export function reconcileVatWithGl(
  periodMonth: number,
  periodYear: number,
  register: VatRegisterTotals,
  gl: GlVatBalances,
): VatReconciliationReport {
  // แปลงเป็น Satang เพื่อความแม่นยำ 100% ป้องกัน floating point error
  const glInputSatang = BigInt(Math.round(gl.inputVatDebit * 100));
  const regInputSatang = BigInt(Math.round(register.purchasesVat * 100));
  const diffInputSatang = glInputSatang - regInputSatang;

  const glOutputSatang = BigInt(Math.round(gl.outputVatCredit * 100));
  const regOutputSatang = BigInt(Math.round(register.salesVat * 100));
  const diffOutputSatang = glOutputSatang - regOutputSatang;

  const inputDiff = Number(diffInputSatang) / 100;
  const outputDiff = Number(diffOutputSatang) / 100;

  const inputMatched = Math.abs(inputDiff) < 0.01;
  const outputMatched = Math.abs(outputDiff) < 0.01;
  const isFullyBalanced = inputMatched && outputMatched;

  const totalVariance = Math.abs(inputDiff) + Math.abs(outputDiff);

  const recommendations: string[] = [];

  if (!inputMatched) {
    if (inputDiff > 0) {
      recommendations.push(
        `ภาษีซื้อใน GL สูงกว่ารายงานภาษีซื้อ ${inputDiff.toFixed(2)} บาท: อาจมีใบสำคัญจ่ายที่ลงบัญชีภาษีซื้อ 1151 ไว้ แต่ยังไม่ได้เพิ่มในทะเบียนใบกำกับภาษี หรือเป็นภาษีซื้อต้องห้ามที่ไม่ได้ตัดออก`,
      );
    } else {
      recommendations.push(
        `ภาษีซื้อในรายงานภาษีซื้อสูงกว่า GL ${Math.abs(inputDiff).toFixed(2)} บาท: มีใบกำกับภาษีซื้อที่ยังไม่ได้บันทึกสมุดรายวัน GL หรือบันทึกผิดผังบัญชี`,
      );
    }
  }

  if (!outputMatched) {
    if (outputDiff > 0) {
      recommendations.push(
        `ภาษีขายใน GL สูงกว่ารายงานภาษีขาย ${outputDiff.toFixed(2)} บาท: มีการลงบัญชีภาษีขาย 2141 ใน GL โดยไม่ได้ออกใบกำกับภาษี หรือยอดขายยังไม่ถึงกำหนดส่งมอบ`,
      );
    } else {
      recommendations.push(
        `ภาษีขายในรายงานภาษีขายสูงกว่า GL ${Math.abs(outputDiff).toFixed(2)} บาท: มีใบกำกับภาษีขายที่ยังไม่ได้ผ่านรายการเข้าบัญชี 2141 ใน GL`,
      );
    }
  }

  if (isFullyBalanced) {
    recommendations.push(
      "ยอดภาษีซื้อและภาษีขายใน GL ตรงกับรายงานภาษีและแบบ ภ.พ.30 ครบถ้วน 100% พร้อมยื่นแบบและปิดงวดบัญชี",
    );
  }

  return {
    periodMonth,
    periodYear,
    inputVat: {
      accountCode: "1151",
      accountNameTh: "ภาษีซื้อ (Input VAT)",
      glAmount: gl.inputVatDebit,
      registerAmount: register.purchasesVat,
      variance: inputDiff,
      status: inputMatched ? "matched" : "discrepancy",
      noteTh: inputMatched
        ? "ยอดตรงกันสมบูรณ์"
        : `ผลต่าง ${inputDiff > 0 ? "+" : ""}${inputDiff.toFixed(2)} บาท`,
    },
    outputVat: {
      accountCode: "2141",
      accountNameTh: "ภาษีขาย (Output VAT)",
      glAmount: gl.outputVatCredit,
      registerAmount: register.salesVat,
      variance: outputDiff,
      status: outputMatched ? "matched" : "discrepancy",
      noteTh: outputMatched
        ? "ยอดตรงกันสมบูรณ์"
        : `ผลต่าง ${outputDiff > 0 ? "+" : ""}${outputDiff.toFixed(2)} บาท`,
    },
    totalVariance,
    isFullyBalanced,
    statusMessageTh: isFullyBalanced
      ? "ยอดบัญชี GL และทะเบียนภาษีตรงกันสมบูรณ์ (100% Reconciled)"
      : `พบผลต่างการกระทบยอดภาษี ${totalVariance.toFixed(2)} บาท กรุณาตรวจสอบก่อนยื่นแบบ`,
    recommendationsTh: recommendations,
  };
}

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

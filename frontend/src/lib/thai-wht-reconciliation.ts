// Thai Withholding Tax (WHT) Reconciliation Engine
// ตรวจสอบกระทบยอดภาษีหัก ณ ที่จ่ายระหว่างบัญชีแยกประเภท (GL) กับทะเบียน ภ.ง.ด.3, ภ.ง.ด.53 และใบรับรอง 50 ทวิ
// พัฒนาตามประมวลรัษฎากร มาตรา 3 เตรส, มาตรา 50, มาตรา 52, และมาตรฐาน TFRS for NPAEs

import type { ThaiWhtRecord } from "./thai-wht";

export interface GlWhtBalances {
  /** ยอดเครดิตบัญชี 2151 ภาษีหัก ณ ที่จ่ายค้างจ่าย ใน GL งวดนี้ (บาท) */
  whtPayableCredit: number;
  /** ยอดเดบิตบัญชี 1161 ภาษีเงินได้ถูกหัก ณ ที่จ่าย ใน GL งวดนี้ (บาท) */
  whtReceivableDebit: number;
}

export interface WhtRegisterData {
  pnd3Records: ThaiWhtRecord[];
  pnd53Records: ThaiWhtRecord[];
  receivedRecords?: ThaiWhtRecord[];
}

export interface WhtReconciliationItem {
  accountCode: string;
  accountNameTh: string;
  glAmount: number;
  registerAmount: number;
  variance: number;
  status: "matched" | "discrepancy";
  noteTh: string;
}

export interface ThaiWhtPendingCertificate {
  docNo: string;
  docDate: string;
  customerName: string;
  customerTaxId: string;
  incomeDescription: string;
  baseAmount: number;
  whtAmount: number;
  daysOutstanding: number;
}

export interface WhtReconciliationReport {
  periodMonth: number;
  periodYear: number;
  payableItem: WhtReconciliationItem;     // บัญชี 2151 ภาษีหัก ณ ที่จ่ายค้างจ่าย
  receivableItem: WhtReconciliationItem;  // บัญชี 1161 ภาษีเงินได้ถูกหัก ณ ที่จ่าย
  isPayableMatched: boolean;
  isReceivableMatched: boolean;
  isFullyBalanced: boolean;
  totalVariance: number;
  duePaymentDatePaperTh: string;           // ครบกำหนดยื่นกระดาษ (วันที่ 7 เดือนถัดไป)
  duePaymentDateOnlineTh: string;          // ครบกำหนดยื่นอินเทอร์เน็ต (วันที่ 15 เดือนถัดไป)
  cashOutflowRequired: number;             // กระแสเงินสดที่ต้องเตรียมชำระให้กรมสรรพากร
  totalTaxShieldAnnual: number;            // ยอดภาษีถูกหักสะสมสำหรับเป็นเครดิตภาษีสิ้นปี (ภ.ง.ด.50)
  pendingCertificates: ThaiWhtPendingCertificate[];
  recommendationsTh: string[];
}

/**
 * คำนวณกระทบยอดภาษีหัก ณ ที่จ่าย 2 ด้าน:
 * 1. ภาษีหัก ณ ที่จ่ายค้างจ่าย (WHT Payable - บัญชี 2151) vs ทะเบียน ภ.ง.ด.3 + ภ.ง.ด.53
 * 2. ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT Receivable - บัญชี 1161) vs ทะเบียนหนังสือรับรอง 50 ทวิที่ได้รับ
 */
export function reconcileWhtWithGl(
  periodMonth: number,
  periodYearCe: number,
  registers: WhtRegisterData,
  gl: GlWhtBalances,
  pendingCerts: ThaiWhtPendingCertificate[] = [],
): WhtReconciliationReport {
  // 1. ด้านภาษีหัก ณ ที่จ่ายค้างจ่าย (WHT Payable)
  const activePnd3 = (registers.pnd3Records || []).filter((r) => r.status === "active");
  const activePnd53 = (registers.pnd53Records || []).filter((r) => r.status === "active");

  let pnd3Satang = 0n;
  for (const r of activePnd3) {
    pnd3Satang += BigInt(Math.round(r.whtAmount * 100));
  }

  let pnd53Satang = 0n;
  for (const r of activePnd53) {
    pnd53Satang += BigInt(Math.round(r.whtAmount * 100));
  }

  const registerPayableSatang = pnd3Satang + pnd53Satang;
  const glPayableSatang = BigInt(Math.round(gl.whtPayableCredit * 100));
  const diffPayableSatang = glPayableSatang - registerPayableSatang;

  const regPayableAmount = Number(registerPayableSatang) / 100;
  const payableDiff = Number(diffPayableSatang) / 100;
  const isPayableMatched = Math.abs(payableDiff) < 0.01;

  // 2. ด้านภาษีถูกหัก ณ ที่จ่าย (WHT Receivable)
  const activeReceived = (registers.receivedRecords || []).filter((r) => r.status === "active");
  let receivedSatang = 0n;
  for (const r of activeReceived) {
    receivedSatang += BigInt(Math.round(r.whtAmount * 100));
  }

  const registerReceivableSatang = receivedSatang;
  const glReceivableSatang = BigInt(Math.round(gl.whtReceivableDebit * 100));
  const diffReceivableSatang = glReceivableSatang - registerReceivableSatang;

  const regReceivableAmount = Number(registerReceivableSatang) / 100;
  const receivableDiff = Number(diffReceivableSatang) / 100;
  const isReceivableMatched = Math.abs(receivableDiff) < 0.01;

  const isFullyBalanced = isPayableMatched && isReceivableMatched;
  const totalVariance = Math.abs(payableDiff) + Math.abs(receivableDiff);

  // คำนวณวันครบกำหนดนำส่งภาษีของเดือนถัดไป
  let nextMonth = periodMonth + 1;
  let nextYearCe = periodYearCe;
  if (nextMonth > 12) {
    nextMonth = 1;
    nextYearCe += 1;
  }
  const nextYearBe = nextYearCe + 543;
  const monthNamesTh = [
    "",
    "มกราคม",
    "กุมภาพันธ์",
    "มีนาคม",
    "เมษายน",
    "พฤษภาคม",
    "มิถุนายน",
    "กรกฎาคม",
    "สิงหาคม",
    "กันยายน",
    "ตุลาคม",
    "พฤศจิกายน",
    "ธันวาคม",
  ];

  const duePaymentDatePaperTh = `7 ${monthNamesTh[nextMonth]} ${nextYearBe}`;
  const duePaymentDateOnlineTh = `15 ${monthNamesTh[nextMonth]} ${nextYearBe}`;

  // คำแนะนำเชิงรุก AI Accounting Advisor
  const recommendations: string[] = [];

  if (isPayableMatched) {
    recommendations.push(
      `✅ บัญชี 2151 ภาษีหัก ณ ที่จ่ายค้างจ่าย ดุลสมบูรณ์ 100% กับแบบยื่น ภ.ง.ด.3 (฿${(Number(pnd3Satang) / 100).toLocaleString("th-TH", { minimumFractionDigits: 2 })}) และ ภ.ง.ด.53 (฿${(Number(pnd53Satang) / 100).toLocaleString("th-TH", { minimumFractionDigits: 2 })}) พร้อมนำส่งเงินได้ทันที`,
    );
  } else {
    if (payableDiff > 0) {
      recommendations.push(
        `⚠️ บัญชี GL 2151 มียอดสูงกว่าแบบยื่น ฿${Math.abs(payableDiff).toLocaleString("th-TH", { minimumFractionDigits: 2 })}: อาจมีใบสำคัญจ่ายที่บันทึกหักภาษีใน GL แล้ว แต่ยังไม่ได้สร้างหนังสือรับรอง 50 ทวิ หรือยังไม่ได้นำเข้าแบบยื่นภาษี`,
      );
    } else {
      recommendations.push(
        `⚠️ แบบยื่นภาษีมียอดสูงกว่าบัญชี GL 2151 อยู่ ฿${Math.abs(payableDiff).toLocaleString("th-TH", { minimumFractionDigits: 2 })}: มีการออกใบรับรอง 50 ทวิ แต่ฝ่ายบัญชียังไม่ได้บันทึกใบสำคัญจ่าย (Payment Voucher) ในสมุดรายวัน GL`,
      );
    }
  }

  if (isReceivableMatched) {
    recommendations.push(
      `✅ บัญชี 1161 ภาษีเงินได้ถูกหัก ณ ที่จ่าย ตรงกับหนังสือรับรอง 50 ทวิที่ได้รับครบถ้วน พร้อมใช้เป็นเครดิตภาษีสิ้นปี`,
    );
  } else {
    if (receivableDiff > 0) {
      recommendations.push(
        `⚠️ บัญชี GL 1161 สูงกว่าทะเบียนใบ 50 ทวิที่ได้รับ ฿${Math.abs(receivableDiff).toLocaleString("th-TH", { minimumFractionDigits: 2 })}: บันทึกรายได้และภาษีถูกหักแล้ว แต่ยังไม่ได้รับใบ 50 ทวิฉบับจริงจากลูกค้า กรุณาติดตามเอกสารเพื่อใช้ยื่น ภ.ง.ด.50`,
      );
    } else {
      recommendations.push(
        `⚠️ ได้รับหนังสือรับรอง 50 ทวิแต่ยังไม่ได้ลงบัญชี GL 1161 จำนวน ฿${Math.abs(receivableDiff).toLocaleString("th-TH", { minimumFractionDigits: 2 })}: กรุณาบันทึกรับรู้สิทธิเครดิตภาษีใน GL`,
      );
    }
  }

  if (pendingCerts.length > 0) {
    recommendations.push(
      `📋 มีหนังสือรับรอง 50 ทวิค้างรับที่ต้องติดตามจากลูกค้า ${pendingCerts.length} ฉบับ รวมยอดภาษี ฿${pendingCerts.reduce((acc, c) => acc + c.whtAmount, 0).toLocaleString("th-TH", { minimumFractionDigits: 2 })}`,
    );
  }

  recommendations.push(
    `💡 กำหนดเวลานำส่งภาษี: ยื่นแบบกระดาษภายใน ${duePaymentDatePaperTh} หรือ ยื่นผ่านอินเทอร์เน็ต (e-Filing) ภายใน ${duePaymentDateOnlineTh}`,
  );

  return {
    periodMonth,
    periodYear: periodYearCe,
    payableItem: {
      accountCode: "2151",
      accountNameTh: "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย (ภ.ง.ด.3 / ภ.ง.ด.53)",
      glAmount: gl.whtPayableCredit,
      registerAmount: regPayableAmount,
      variance: payableDiff,
      status: isPayableMatched ? "matched" : "discrepancy",
      noteTh: isPayableMatched
        ? "ยอดดุลสมบูรณ์ 100%"
        : payableDiff > 0
          ? "GL สูงกว่าแบบยื่น (ตรวจใบสำคัญจ่ายค้างออก 50 ทวิ)"
          : "แบบยื่นสูงกว่า GL (ตรวจการลงบัญชีสมุดรายวันจ่าย)",
    },
    receivableItem: {
      accountCode: "1161",
      accountNameTh: "ภาษีเงินได้ถูกหัก ณ ที่จ่าย (ลูกหนี้สรรพากร / เครดิตภาษี)",
      glAmount: gl.whtReceivableDebit,
      registerAmount: regReceivableAmount,
      variance: receivableDiff,
      status: isReceivableMatched ? "matched" : "discrepancy",
      noteTh: isReceivableMatched
        ? "ยอดดุลสมบูรณ์ 100%"
        : receivableDiff > 0
          ? "GL สูงกว่าทะเบียน (มีใบ 50 ทวิค้างรับจากลูกค้า)"
          : "ทะเบียนสูงกว่า GL (ตรวจการบันทึกสิทธิเครดิตภาษี)",
    },
    isPayableMatched,
    isReceivableMatched,
    isFullyBalanced,
    totalVariance,
    duePaymentDatePaperTh,
    duePaymentDateOnlineTh,
    cashOutflowRequired: regPayableAmount,
    totalTaxShieldAnnual: gl.whtReceivableDebit,
    pendingCertificates: pendingCerts,
    recommendationsTh: recommendations,
  };
}

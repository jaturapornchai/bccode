// Thai Month-End VAT Closing Engine
// ระบบสร้างรายการโอนปิดบัญชีภาษีมูลค่าเพิ่มสิ้นงวด (VAT Closing Journal Voucher)
// ตามมาตรฐานการบัญชีไทย TFRS for NPAEs และประมวลรัษฎากร

export interface VatClosingJournalLine {
  accountCode: string;
  accountNameTh: string;
  debit: string;
  credit: string;
  descriptionTh: string;
}

export interface VatClosingResult {
  docno: string;
  date: string;
  description: string;
  bookCode: string;
  lines: VatClosingJournalLine[];
  totalDebit: number;
  totalCredit: number;
  isBalanced: boolean;
  netVatType: "payable" | "refundable" | "zero";
  netAmount: number;
  summaryNoteTh: string;
}

export interface VatClosingParams {
  year: number;
  month: number;
  outputVat: number;
  inputVat: number;
  creditBroughtForward?: number;
  outputVatAccount?: string;
  inputVatAccount?: string;
  vatPayableAccount?: string;
  vatRefundableAccount?: string;
  bookCode?: string;
  docno?: string;
}

/**
 * หาวันสุดท้ายของเดือนในรูปแบบ YYYY-MM-DD
 */
export function getLastDayOfMonth(year: number, month: number): string {
  const lastDay = new Date(year, month, 0).getDate();
  const m = String(month).padStart(2, "0");
  const d = String(lastDay).padStart(2, "0");
  return `${year}-${m}-${d}`;
}

/**
 * แปลงจำนวนเงินเป็นข้อความภาษาไทย (Thai Baht Text)
 * เช่น 1250.50 -> "หนึ่งพันสองร้อยห้าสิบบาทห้าสิบสตางค์"
 */
export function thaiBahtText(amount: number): string {
  if (isNaN(amount) || amount === 0) return "ศูนย์บาทถ้วน";

  const isNegative = amount < 0;
  const absAmount = Math.abs(amount);

  // แยกจำนวนเต็มและสตางค์
  const roundedSatang = Math.round(absAmount * 100);
  const baht = Math.floor(roundedSatang / 100);
  const satang = roundedSatang % 100;

  const numbersTh = ["ศูนย์", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"];
  const unitsTh = ["", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน", "ล้าน"];

  function convertGroup(num: number): string {
    const s = String(num);
    const len = s.length;
    let result = "";

    for (let i = 0; i < len; i++) {
      const digit = Number(s[i]);
      const pos = len - i - 1;

      if (digit === 0) continue;

      if (pos === 0 && digit === 1 && len > 1) {
        result += "เอ็ด";
      } else if (pos === 1 && digit === 1) {
        result += "สิบ";
      } else if (pos === 1 && digit === 2) {
        result += "ยี่สิบ";
      } else {
        result += numbersTh[digit] + unitsTh[pos];
      }
    }
    return result;
  }

  let text = "";
  if (baht === 0) {
    text = "ศูนย์บาท";
  } else {
    // รองรับจำนวนเกินล้าน
    if (baht >= 1000000) {
      const millionGroup = Math.floor(baht / 1000000);
      const remainGroup = baht % 1000000;
      text = `${convertGroup(millionGroup)}ล้าน${convertGroup(remainGroup)}บาท`;
    } else {
      text = `${convertGroup(baht)}บาท`;
    }
  }

  if (satang === 0) {
    text += "ถ้วน";
  } else {
    text += `${convertGroup(satang)}สตางค์`;
  }

  return isNegative ? `ลบ${text}` : text;
}

/**
 * สร้างรายการโอนปิดภาษีซื้อ-ภาษีขายสิ้นงวด (Month-End VAT Closing Voucher)
 */
export function generateVatClosingJournal(params: VatClosingParams): VatClosingResult {
  const outputAccount = params.outputVatAccount || "2141";
  const inputAccount = params.inputVatAccount || "1151";
  const payableAccount = params.vatPayableAccount || "2142";
  const refundableAccount = params.vatRefundableAccount || "1152";
  const bookCode = params.bookCode || "JV";

  const date = getLastDayOfMonth(params.year, params.month);
  const mStr = String(params.month).padStart(2, "0");
  const docno = params.docno || `JV-VAT-${params.year}${mStr}`;
  const description = `โอนปิดบัญชีภาษีมูลค่าเพิ่มประจำงวดเดือน ${params.month}/${params.year}`;

  const outputSatang = BigInt(Math.round(params.outputVat * 100));
  const inputSatang = BigInt(Math.round(params.inputVat * 100));
  const creditForwardSatang = BigInt(Math.round((params.creditBroughtForward || 0) * 100));

  const totalCreditsSatang = inputSatang + creditForwardSatang;
  const netSatang = outputSatang - totalCreditsSatang;

  const lines: VatClosingJournalLine[] = [];

  // 1. ล้างบัญชีภาษีขาย (เดบิตลดภาษีขาย)
  if (outputSatang > 0n) {
    lines.push({
      accountCode: outputAccount,
      accountNameTh: "ภาษีขาย (Output VAT)",
      debit: (Number(outputSatang) / 100).toFixed(2),
      credit: "",
      descriptionTh: `โอนปิดภาษีขายประจำเดือน ${params.month}/${params.year}`,
    });
  }

  // 2. ล้างบัญชีภาษีซื้อ (เครดิตลดภาษีซื้อ)
  if (inputSatang > 0n) {
    lines.push({
      accountCode: inputAccount,
      accountNameTh: "ภาษีซื้อ (Input VAT)",
      debit: "",
      credit: (Number(inputSatang) / 100).toFixed(2),
      descriptionTh: `โอนปิดภาษีซื้อประจำเดือน ${params.month}/${params.year}`,
    });
  }

  // 2.1 ล้างยอดภาษีชำระเกินยกมาจากงวดก่อนที่นำมาใช้ในงวดนี้ (ถ้ามี)
  if (creditForwardSatang > 0n) {
    lines.push({
      accountCode: refundableAccount,
      accountNameTh: "ลูกหนี้กรมสรรพากร / ภาษีมูลค่าเพิ่มรอขอคืน (เครดิตยกมา)",
      debit: "",
      credit: (Number(creditForwardSatang) / 100).toFixed(2),
      descriptionTh: `ตัดยอดภาษีมูลค่าเพิ่มชำระเกินยกมาจากเดือนก่อนที่นำมาเครดิตในเดือนนี้ (ภ.พ.30 ข้อ 8)`,
    });
  }

  // 3. ผลต่างสุทธิ
  let netVatType: "payable" | "refundable" | "zero" = "zero";
  let summaryNoteTh = "";

  if (netSatang > 0n) {
    // ภาษีขาย > ภาษีซื้อ -> มีภาษีต้องชำระ (ตั้งเป็นเจ้าหนี้สรรพากร ฝั่งเครดิต)
    netVatType = "payable";
    const payableAmount = Number(netSatang) / 100;
    lines.push({
      accountCode: payableAccount,
      accountNameTh: "ภาษีมูลค่าเพิ่มค้างจ่าย / เจ้าหนี้กรมสรรพากร",
      debit: "",
      credit: payableAmount.toFixed(2),
      descriptionTh: `ตั้งภาษีมูลค่าเพิ่มที่ต้องชำระงวดเดือน ${params.month}/${params.year} (ภ.พ.30 ข้อ 9)`,
    });
    summaryNoteTh = `ภาษีขายมากกว่าภาษีซื้อ มีภาษีมูลค่าเพิ่มที่ต้องชำระต่อกรมสรรพากร ${payableAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท (${thaiBahtText(payableAmount)})`;
  } else if (netSatang < 0n) {
    // ภาษีซื้อ > ภาษีขาย -> มีภาษีชำระเกิน (ตั้งเป็นลูกหนี้สรรพากร ฝั่งเดบิต)
    netVatType = "refundable";
    const refundableAmount = Number(-netSatang) / 100;
    lines.push({
      accountCode: refundableAccount,
      accountNameTh: "ลูกหนี้กรมสรรพากร / ภาษีมูลค่าเพิ่มรอขอคืน",
      debit: refundableAmount.toFixed(2),
      credit: "",
      descriptionTh: `ตั้งภาษีมูลค่าเพิ่มชำระเกินงวดเดือน ${params.month}/${params.year} เครดิตยกไปงวดถัดไป (ภ.พ.30 ข้อ 10)`,
    });
    summaryNoteTh = `ภาษีซื้อมากกว่าภาษีขาย มีภาษีมูลค่าเพิ่มชำระเกินยกไปเครดิตในเดือนถัดไป ${refundableAmount.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท (${thaiBahtText(refundableAmount)})`;
  } else {
    summaryNoteTh = "ภาษีขายและภาษีซื้อเท่ากันพอดี ไม่มียอดภาษีค้างชำระหรือชำระเกิน";
  }

  // คำนวณผลรวมเดบิต-เครดิต เพื่อยืนยันความสมดุล
  let totalDebitSatang = 0n;
  let totalCreditSatang = 0n;

  for (const line of lines) {
    if (line.debit) totalDebitSatang += BigInt(Math.round(Number(line.debit) * 100));
    if (line.credit) totalCreditSatang += BigInt(Math.round(Number(line.credit) * 100));
  }

  const isBalanced = totalDebitSatang === totalCreditSatang;

  return {
    docno,
    date,
    description,
    bookCode,
    lines,
    totalDebit: Number(totalDebitSatang) / 100,
    totalCredit: Number(totalCreditSatang) / 100,
    isBalanced,
    netVatType,
    netAmount: Number(netSatang > 0n ? netSatang : -netSatang) / 100,
    summaryNoteTh,
  };
}

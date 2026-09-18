// Thai Tax Export Engine for Revenue Department (RD Prep & e-Filing)
// รองรับการส่งออกไฟล์ ภ.พ.30, ภ.ง.ด.3, ภ.ง.ด.53 ตามรูปแบบของกรมสรรพากรและ New e-Filing
// พัฒนาตามระเบียบกรมสรรพากร และมาตรฐานการคำนวณเลขประจำตัวผู้เสียภาษี 13 หลัก (Mod 11)

import type { ThaiWhtRecord } from "./thai-wht";
import type { OfficialPp30FormData } from "./thai-vat-reconciliation";

/**
 * ตรวจสอบความถูกต้องของเลขประจำตัวผู้เสียภาษีอากร / เลขประจำตัวประชาชน 13 หลัก
 * ตามหลักการ Modulo 11 Checksum ของกรมการปกครองและกรมสรรพากร
 */
export function validateThaiTaxId(taxId: string): { isValid: boolean; reason?: string } {
  const cleanId = String(taxId || "").replace(/[^0-9]/g, "");

  if (cleanId.length !== 13) {
    return {
      isValid: false,
      reason: `ความยาวต้องมี 13 หลัก (ปัจจุบันมี ${cleanId.length} หลัก)`,
    };
  }

  // ป้องกันเลขตองซ้ำ 13 ตัว เช่น 0000000000000 หรือ 9999999999999
  if (/^(\d)\1{12}$/.test(cleanId)) {
    return {
      isValid: false,
      reason: "เลขประจำตัวต้องไม่ใช่ตัวเลขเดียวกันซ้ำกันทั้ง 13 หลัก",
    };
  }

  // คำนวณ Checksum Modulo 11
  let sum = 0;
  for (let i = 0; i < 12; i++) {
    sum += parseInt(cleanId.charAt(i), 10) * (13 - i);
  }

  const checkDigit = (11 - (sum % 11)) % 10;
  const actualCheckDigit = parseInt(cleanId.charAt(12), 10);

  if (checkDigit !== actualCheckDigit) {
    return {
      isValid: false,
      reason: `เลขตรวจสอบหลักที่ 13 ไม่ถูกต้อง (คำนวณได้ ${checkDigit} แต่ระบุเป็น ${actualCheckDigit})`,
    };
  }

  return { isValid: true };
}

/**
 * ปรับรูปแบบรหัสสาขาให้เป็น 5 หลักมาตรฐาน เช่น 0 -> 00000, 1 -> 00001
 */
export function normalizeBranchNo(branch: string | number | undefined | null): string {
  const str = String(branch ?? "").trim().replace(/[^0-9]/g, "");
  if (!str) return "00000";
  return str.padStart(5, "0").slice(-5);
}

/**
 * แปลงวันที่ YYYY-MM-DD เป็นรูปแบบ พ.ศ. (วว/ดด/ปปปป หรือ ววดดปปปป)
 */
export function formatThaiTaxDate(
  dateStr: string,
  format: "dd/mm/yyyy" | "ddmmyyyy" | "yyyy-mm-dd" = "dd/mm/yyyy",
): string {
  if (!dateStr) return "";
  const parts = dateStr.split("-");
  if (parts.length !== 3) return dateStr;

  const yearCe = parseInt(parts[0], 10);
  const month = parts[1].padStart(2, "0");
  const day = parts[2].padStart(2, "0");
  const yearBe = yearCe > 2400 ? yearCe : yearCe + 543;

  if (format === "dd/mm/yyyy") {
    return `${day}/${month}/${yearBe}`;
  }
  if (format === "ddmmyyyy") {
    return `${day}${month}${yearBe}`;
  }
  return `${yearCe}-${month}-${day}`;
}

export type RdPrepDelimiter = "|" | "," | "\t";

export interface RdPrepExportOptions {
  delimiter?: RdPrepDelimiter;
  includeHeader?: boolean;
}

/**
 * แยกคำนำหน้าชื่อ ชื่อ และนามสกุล สำหรับบุคคลธรรมดา (ภ.ง.ด.3)
 */
export function parseIndividualName(fullName: string): {
  prefix: string;
  firstName: string;
  lastName: string;
} {
  const trimmed = fullName.trim();
  const prefixes = [
    "นาย",
    "นางสาว",
    "นาง",
    "น.ส.",
    "ด.ช.",
    "ด.ญ.",
    "ว่าที่ร้อยตรี",
    "ร้อยตรี",
    "พันตรี",
    "พลเอก",
    "ดร.",
    "นพ.",
    "พญ.",
  ];

  let prefix = "";
  let remaining = trimmed;

  for (const p of prefixes) {
    if (trimmed.startsWith(p)) {
      prefix = p;
      remaining = trimmed.slice(p.length).trim();
      break;
    }
  }

  const nameParts = remaining.split(/\s+/);
  const firstName = nameParts[0] || "";
  const lastName = nameParts.slice(1).join(" ") || "";

  return { prefix, firstName, lastName };
}

/**
 * แปลงรหัสเงื่อนไขการหักภาษีเป็นรหัสของกรมสรรพากร:
 * 1 = หัก ณ ที่จ่าย
 * 2 = ออกให้ตลอดไป
 * 3 = ออกให้ครั้งเดียว
 */
export function mapTaxConditionCode(condition: "deducted" | "forever" | "once" | string): string {
  switch (condition) {
    case "forever":
      return "2";
    case "once":
      return "3";
    case "deducted":
    default:
      return "1";
  }
}

/**
 * สร้างข้อมูลไฟล์ RD Prep สำหรับแบบ ภ.ง.ด.3 (หัก ณ ที่จ่าย บุคคลธรรมดา)
 * ตามรูปแบบของโปรแกรม RD Prep กรมสรรพากร
 */
export function generateRdPrepPnd3(
  records: ThaiWhtRecord[],
  options: RdPrepExportOptions = {},
): string {
  const delimiter = options.delimiter ?? "|";
  const activeRecords = records.filter(
    (r) => r.filingType === "pnd3" && r.status === "active",
  );

  const lines: string[] = [];

  if (options.includeHeader) {
    lines.push(
      [
        "ลำดับที่",
        "เลขประจำตัวประชาชน",
        "สาขา",
        "คำนำหน้าชื่อ",
        "ชื่อ",
        "นามสกุล",
        "ที่อยู่",
        "วันที่จ่าย",
        "ประเภทเงินได้",
        "อัตราภาษี",
        "จำนวนเงินที่จ่าย",
        "ภาษีที่หัก",
        "เงื่อนไขการหัก",
      ].join(delimiter),
    );
  }

  activeRecords.forEach((r, idx) => {
    const { prefix, firstName, lastName } = parseIndividualName(r.payeeName);
    const taxId = r.payeeTaxId.replace(/[^0-9]/g, "");
    const branch = normalizeBranchNo(r.payeeBranchNo);
    const dateFormatted = formatThaiTaxDate(r.docDate, "dd/mm/yyyy");
    const incomeDesc = (r.incomeDescription || "ค่าบริการ").replace(/[\r\n|]/g, " ");
    const address = (r.payeeAddress || "-").replace(/[\r\n|]/g, " ");
    const rateStr = r.taxRate.toFixed(2);
    const amountStr = r.paymentAmount.toFixed(2);
    const whtStr = r.whtAmount.toFixed(2);
    const conditionCode = mapTaxConditionCode(r.condition);

    const row = [
      String(idx + 1),
      taxId,
      branch,
      prefix,
      firstName,
      lastName,
      address,
      dateFormatted,
      incomeDesc,
      rateStr,
      amountStr,
      whtStr,
      conditionCode,
    ].join(delimiter);

    lines.push(row);
  });

  return lines.join("\r\n");
}

/**
 * สร้างข้อมูลไฟล์ RD Prep สำหรับแบบ ภ.ง.ด.53 (หัก ณ ที่จ่าย นิติบุคคล)
 * ตามรูปแบบของโปรแกรม RD Prep กรมสรรพากร
 */
export function generateRdPrepPnd53(
  records: ThaiWhtRecord[],
  options: RdPrepExportOptions = {},
): string {
  const delimiter = options.delimiter ?? "|";
  const activeRecords = records.filter(
    (r) => r.filingType === "pnd53" && r.status === "active",
  );

  const lines: string[] = [];

  if (options.includeHeader) {
    lines.push(
      [
        "ลำดับที่",
        "เลขประจำตัวผู้เสียภาษีอากร",
        "สาขา",
        "ชื่อนิติบุคคล",
        "ที่อยู่",
        "วันที่จ่าย",
        "ประเภทเงินได้",
        "อัตราภาษี",
        "จำนวนเงินที่จ่าย",
        "ภาษีที่หัก",
        "เงื่อนไขการหัก",
      ].join(delimiter),
    );
  }

  activeRecords.forEach((r, idx) => {
    const taxId = r.payeeTaxId.replace(/[^0-9]/g, "");
    const branch = normalizeBranchNo(r.payeeBranchNo);
    const name = r.payeeName.replace(/[\r\n|]/g, " ");
    const address = (r.payeeAddress || "-").replace(/[\r\n|]/g, " ");
    const dateFormatted = formatThaiTaxDate(r.docDate, "dd/mm/yyyy");
    const incomeDesc = (r.incomeDescription || "ค่าบริการ").replace(/[\r\n|]/g, " ");
    const rateStr = r.taxRate.toFixed(2);
    const amountStr = r.paymentAmount.toFixed(2);
    const whtStr = r.whtAmount.toFixed(2);
    const conditionCode = mapTaxConditionCode(r.condition);

    const row = [
      String(idx + 1),
      taxId,
      branch,
      name,
      address,
      dateFormatted,
      incomeDesc,
      rateStr,
      amountStr,
      whtStr,
      conditionCode,
    ].join(delimiter);

    lines.push(row);
  });

  return lines.join("\r\n");
}

/**
 * สร้างข้อมูลไฟล์สรุปยื่นแบบ ภ.พ.30 สำหรับระบบ e-Filing กรมสรรพากร
 */
export function generateRdPrepPp30(
  form: OfficialPp30FormData,
  options: RdPrepExportOptions = {},
): string {
  const delimiter = options.delimiter ?? "|";
  const lines: string[] = [];

  if (options.includeHeader) {
    lines.push(
      [
        "เลขประจำตัวผู้เสียภาษี",
        "สาขา",
        "เดือนภาษี",
        "ปีภาษี",
        "ประเภทยื่น",
        "ยอดขายข้อ1",
        "ยอดขายข้อ2",
        "ยอดขายข้อ3",
        "ยอดขายข้อ4",
        "ภาษีขายข้อ5",
        "ยอดซื้อข้อ6",
        "ภาษีซื้อข้อ7",
        "เครดิตยกมาข้อ8",
        "ภาษีต้องชำระข้อ9",
        "ภาษีชำระเกินข้อ10",
        "เงินเพิ่ม",
        "เบี้ยปรับ",
        "ยอดชำระสุทธิ",
      ].join(delimiter),
    );
  }

  const taxId = form.taxId.replace(/[^0-9]/g, "");
  const branch = normalizeBranchNo(form.branchNo);
  const monthStr = String(form.taxPeriodMonth).padStart(2, "0");
  const yearStr = String(form.taxPeriodYearBe);
  const filingCode = form.filingType === "additional" ? "1" : "0";

  const row = [
    taxId,
    branch,
    monthStr,
    yearStr,
    filingCode,
    form.item1_grossSales.toFixed(2),
    form.item2_zeroRatedSales.toFixed(2),
    form.item3_exemptSales.toFixed(2),
    form.item4_taxableSales.toFixed(2),
    form.item5_outputVat.toFixed(2),
    form.item6_taxablePurchases.toFixed(2),
    form.item7_inputVat.toFixed(2),
    form.item8_creditBroughtForward.toFixed(2),
    form.item9_vatPayable.toFixed(2),
    form.item10_vatOverpaid.toFixed(2),
    form.surcharge.toFixed(2),
    form.penalty.toFixed(2),
    form.netPaymentAmount.toFixed(2),
  ].join(delimiter);

  lines.push(row);
  return lines.join("\r\n");
}

/**
 * สร้าง Data Blob สำหรับดาวน์โหลดไฟล์ Text / CSV พร้อม UTF-8 BOM
 * เพื่อให้ Microsoft Excel และโปรแกรมในไทยเปิดภาษาไทยได้ถูกต้อง 100%
 */
export function createDownloadBlob(
  content: string,
  type: "text" | "csv" = "text",
): Blob {
  const bom = "\uFEFF";
  const mime = type === "csv" ? "text/csv;charset=utf-8" : "text/plain;charset=utf-8";
  return new Blob([bom + content], { type: mime });
}


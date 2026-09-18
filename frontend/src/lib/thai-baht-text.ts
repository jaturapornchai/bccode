/**
 * Thai Baht Text (บาทข้อความ) Converter
 * Converts numeric amounts into official Thai Baht text representation.
 * Follows standard Thai accounting & Revenue Department guidelines.
 * E.g.: 1,234.50 -> "หนึ่งพันสองร้อยสามสิบสี่บาทห้าสิบสตางค์"
 *       10,000,000 -> "สิบล้านบาทถ้วน"
 *       0 -> "ศูนย์บาทถ้วน"
 */

const THAI_DIGITS = ["ศูนย์", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"];
const THAI_UNITS = ["", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน", "ล้าน"];

/**
 * Helper to convert integer part up to 6 digits (within one million cycle)
 */
function convertGroup(digits: string, isAbsoluteUnit: boolean = false): string {
  let result = "";
  const len = digits.length;

  for (let i = 0; i < len; i++) {
    const digit = parseInt(digits[i], 10);
    const pos = len - i - 1;

    if (digit === 0) continue;

    if (pos === 0 && digit === 1 && isAbsoluteUnit) {
      // Rule: Ending 1 in units place of numbers > 1 is "เอ็ด"
      result += "เอ็ด";
    } else if (pos === 1 && digit === 2) {
      // Rule: 20 is "ยี่สิบ"
      result += "ยี่สิบ";
    } else if (pos === 1 && digit === 1) {
      // Rule: 10 is "สิบ" (not "หนึ่งสิบ")
      result += "สิบ";
    } else {
      result += THAI_DIGITS[digit] + THAI_UNITS[pos];
    }
  }

  return result;
}

/**
 * Main BahtText conversion
 */
export function thaiBahtText(amount: number | string | null | undefined): string {
  if (amount === null || amount === undefined || amount === "") {
    return "ศูนย์บาทถ้วน";
  }

  const num = typeof amount === "number" ? amount : parseFloat(String(amount).replace(/,/g, ""));
  if (isNaN(num)) {
    return "ศูนย์บาทถ้วน";
  }

  if (num === 0) {
    return "ศูนย์บาทถ้วน";
  }

  const isNegative = num < 0;
  const absNum = Math.abs(num);

  // Round to 2 decimal places to avoid IEEE-754 precision issues
  const fixed = absNum.toFixed(2);
  const [intPart, decPart] = fixed.split(".");

  let text = "";

  if (intPart === "0") {
    text = "";
  } else {
    // Process groups of 6 digits from right to left (thousands, millions, billions...)
    const groups: string[] = [];
    let remaining = intPart;

    while (remaining.length > 0) {
      const take = Math.min(6, remaining.length);
      const chunk = remaining.slice(remaining.length - take);
      groups.unshift(chunk);
      remaining = remaining.slice(0, remaining.length - take);
    }

    const isMultiDigit = intPart.length > 1 || intPart > "1";
    for (let g = 0; g < groups.length; g++) {
      const isLowestGroup = g === groups.length - 1;
      const groupText = convertGroup(groups[g], isLowestGroup && isMultiDigit);
      const millionsIndex = groups.length - g - 1;

      if (groupText) {
        text += groupText;
        if (millionsIndex > 0) {
          text += "ล้าน".repeat(millionsIndex);
        }
      }
    }

    text += "บาท";
  }

  // Satang part
  const satangVal = parseInt(decPart, 10);
  if (satangVal === 0) {
    text += text ? "ถ้วน" : "ศูนย์บาทถ้วน";
  } else {
    const isSatangMulti = satangVal > 1;
    const satangText = convertGroup(decPart, isSatangMulti);
    text += satangText + "สตางค์";
  }

  return (isNegative ? "ลบ" : "") + text;
}

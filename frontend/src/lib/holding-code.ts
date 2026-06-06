export const holdingCodePattern = /^[a-z][a-z0-9]{2,29}$/;

export const holdingCodeValidationMessageTh =
  "holdingcode ต้องใช้ a-z และ 0-9 เท่านั้น ยาว 3-30 ตัว และขึ้นต้นด้วย a-z ห้ามใช้ _ หรือสัญลักษณ์";

export function normalizeHoldingCode(value: string): string {
  return value.trim().toLowerCase();
}

export function isValidHoldingCode(value: string): boolean {
  return holdingCodePattern.test(value);
}

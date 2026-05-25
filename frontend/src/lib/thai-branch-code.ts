export const THAI_HEAD_OFFICE_BRANCH_CODE = "00000";

const HEAD_OFFICE_ALIASES = new Set([
  "สำนักงานใหญ่",
  "สํานักงานใหญ่",
  "สนญ",
  "head-office",
  "headoffice",
  "hq",
  "ho",
]);

export function normalizeThaiTaxBranchCode(value: unknown): string {
  const compact = String(value ?? "")
    .trim()
    .replace(/\s+/g, "")
    .toLowerCase();
  if (!compact) throw new Error("branch code is required");
  if (HEAD_OFFICE_ALIASES.has(compact)) return THAI_HEAD_OFFICE_BRANCH_CODE;
  if (!/^\d+$/.test(compact))
    throw new Error("branch code must be numeric and no more than 5 digits");
  if (compact.length > 5)
    throw new Error("branch code must be no more than 5 digits");
  return compact.padStart(5, "0");
}

export function isThaiHeadOfficeBranchCode(value: unknown): boolean {
  try {
    return normalizeThaiTaxBranchCode(value) === THAI_HEAD_OFFICE_BRANCH_CODE;
  } catch {
    return false;
  }
}

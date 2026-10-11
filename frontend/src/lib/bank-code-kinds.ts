// ชนิดรหัสของโมดูลเช็ค/ธนาคาร — ไฟล์นี้ไม่ import อะไรเลย เพื่อให้ BFF (server) และจอ (client) ใช้ร่วมกันได้

export const BANK_CODE_KINDS = ["bank", "bank_branch", "bank_service", "bank_account_type", "cheque_ext_status"] as const;
export type BankCodeKind = (typeof BANK_CODE_KINDS)[number];

/** ความยาวสูงสุดตามตาราง Champ (BCBank* / BCBankContact) — backend ตรวจซ้ำเสมอ */
export const BANK_CODE_LIMITS = { code: 10, name: 50, userGroup: 255 } as const;
export const BANK_CONTACT_LIMITS = { code: 10, name: 50, position: 50, address: 255, telephone: 30, rows: 200 } as const;

export const BANK_CODE_ROUTE_PREFIX = "/settings/bank-codes/";

export function isBankCodeKind(value: string): value is BankCodeKind {
  return (BANK_CODE_KINDS as readonly string[]).includes(value);
}

/** route ของจอกำหนดรหัส → ชนิดรหัส (null = ไม่ใช่จอกำหนดรหัส) */
export function bankCodeKindFromRoute(route: string): BankCodeKind | null {
  if (!route.startsWith(BANK_CODE_ROUTE_PREFIX)) return null;
  const kind = route.slice(BANK_CODE_ROUTE_PREFIX.length);
  return isBankCodeKind(kind) ? kind : null;
}

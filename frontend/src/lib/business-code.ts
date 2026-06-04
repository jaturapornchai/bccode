export function normalizeBusinessCode(value: unknown): string {
  return String(value ?? "").trim().toUpperCase();
}

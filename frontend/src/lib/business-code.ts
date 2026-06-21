export function normalizeBusinessCode(value: unknown): string {
  // Business codes are uppercase AND contain no whitespace — strip every space, tab,
  // and newline (leading, trailing, and internal) so `qatest perm 01` -> `QATESTPERM01`.
  return String(value ?? "").toUpperCase().replace(/\s+/g, "");
}

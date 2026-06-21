import { describe, expect, it } from "vitest";

import { normalizeBusinessCode } from "./business-code";

describe("normalizeBusinessCode", () => {
  it("normalizes user-facing business codes to uppercase", () => {
    expect(normalizeBusinessCode(" a-01 ")).toBe("A-01");
    expect(normalizeBusinessCode("sku_abc")).toBe("SKU_ABC");
  });

  it("removes all whitespace inside a business code", () => {
    expect(normalizeBusinessCode("qatest perm 01")).toBe("QATESTPERM01");
    expect(normalizeBusinessCode(" a 01 ")).toBe("A01");
    expect(normalizeBusinessCode("ab\tcd")).toBe("ABCD");
  });
});

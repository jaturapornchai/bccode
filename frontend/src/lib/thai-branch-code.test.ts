import { describe, expect, it } from "vitest";
import {
  normalizeThaiTaxBranchCode,
  THAI_HEAD_OFFICE_BRANCH_CODE,
} from "./thai-branch-code";

describe("Thai tax branch code", () => {
  it("normalizes Thai head office and branch numbers to five digits", () => {
    expect(normalizeThaiTaxBranchCode("สำนักงานใหญ่")).toBe(
      THAI_HEAD_OFFICE_BRANCH_CODE,
    );
    expect(normalizeThaiTaxBranchCode("สนญ")).toBe(
      THAI_HEAD_OFFICE_BRANCH_CODE,
    );
    expect(normalizeThaiTaxBranchCode("0")).toBe(THAI_HEAD_OFFICE_BRANCH_CODE);
    expect(normalizeThaiTaxBranchCode("1")).toBe("00001");
    expect(normalizeThaiTaxBranchCode("01")).toBe("00001");
    expect(normalizeThaiTaxBranchCode("12345")).toBe("12345");
  });

  it("rejects non-tax branch code values", () => {
    expect(() => normalizeThaiTaxBranchCode("")).toThrow();
    expect(() => normalizeThaiTaxBranchCode("branch-a")).toThrow();
    expect(() => normalizeThaiTaxBranchCode("000000")).toThrow();
  });
});

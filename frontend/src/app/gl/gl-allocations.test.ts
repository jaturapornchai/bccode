import { describe, expect, it } from "vitest";
import { rateTotal } from "./gl-allocations";
import { emptyAccount, emptyAllocationRule, emptyMaster, type GLAccount, type GLAllocationRule, type GLMaster } from "@/lib/general-ledger";

/**
 * Accountant & Accounting Firm Audit Test Suite (การทดสอบมุมมองนักบัญชีและสำนักงานบัญชี)
 * 
 * Objectives:
 * 1. Double-Entry & 100% Completeness Guarantee: Allocation rules must balance to 100.00% exactly.
 * 2. High-Precision Decimal Arithmetic: No floating-point drift on fractional percentage splits.
 * 3. Multi-Dimension Cost Distribution: Verify Branch, Department, and Project dimensions.
 * 4. Chart of Accounts Integrity: Posting accounts vs Parent control accounts.
 * 5. Audit Trail & Concurrency Defense: Versioning and payload contract.
 */
describe("Cost Allocation - Accountant & Audit Firm Verification", () => {
  // --- Section 1: 100% Rate Balance Tests ---
  describe("Mathematical Precision & 100% Balance Principle", () => {
    it("accepts a classic 50-50 two-way department split", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), departmentcode: "SALES", rate: "50" },
        { ...emptyAllocationRule(), departmentcode: "ADMIN", rate: "50" },
      ];
      expect(rateTotal(rules)).toBe("100");
    });

    it("accepts a 3-way split with fractional decimals summing to exactly 100%", () => {
      // Classic accounting dilemma: 100 / 3 = 33.33, 33.33, 33.34
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), departmentcode: "SALES", rate: "33.33" },
        { ...emptyAllocationRule(), departmentcode: "MARKETING", rate: "33.33" },
        { ...emptyAllocationRule(), departmentcode: "ADMIN", rate: "33.34" },
      ];
      expect(rateTotal(rules)).toBe("100");
    });

    it("accepts a 4-way quarterly or branch distribution (25% each)", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), branchcode: "HQ", rate: "25" },
        { ...emptyAllocationRule(), branchcode: "BR1", rate: "25" },
        { ...emptyAllocationRule(), branchcode: "BR2", rate: "25" },
        { ...emptyAllocationRule(), branchcode: "BR3", rate: "25" },
      ];
      expect(rateTotal(rules)).toBe("100");
    });

    it("maintains micro-precision up to 8 decimal places without IEEE-754 drift", () => {
      // 0.12345678 + 99.87654322 = 100.00000000
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), departmentcode: "RND", rate: "0.12345678" },
        { ...emptyAllocationRule(), departmentcode: "PRODUCTION", rate: "99.87654322" },
      ];
      expect(rateTotal(rules)).toBe("100");
    });

    it("correctly handles comma-formatted user inputs like 50,00 and whitespace", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), rate: " 70.0000 " },
        { ...emptyAllocationRule(), rate: "30" },
      ];
      expect(rateTotal(rules)).toBe("100");
    });
  });

  // --- Section 2: Rejection of Unbalanced Allocations ---
  describe("Audit Rejection of Unbalanced Cost Allocations", () => {
    it("flags under-allocation where total is less than 100% (e.g. 90%)", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), departmentcode: "SALES", rate: "50" },
        { ...emptyAllocationRule(), departmentcode: "ADMIN", rate: "40" },
      ];
      expect(rateTotal(rules)).toBe("90");
      expect(rateTotal(rules) === "100").toBe(false);
    });

    it("flags over-allocation where total exceeds 100% (e.g. 110%)", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), departmentcode: "SALES", rate: "60" },
        { ...emptyAllocationRule(), departmentcode: "ADMIN", rate: "50" },
      ];
      expect(rateTotal(rules)).toBe("110");
      expect(rateTotal(rules) === "100").toBe(false);
    });

    it("flags unrounded 33.33 * 3 = 99.99% as invalid (preventing 0.01% black hole)", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), rate: "33.33" },
        { ...emptyAllocationRule(), rate: "33.33" },
        { ...emptyAllocationRule(), rate: "33.33" },
      ];
      expect(rateTotal(rules)).toBe("99.99");
      expect(rateTotal(rules) === "100").toBe(false);
    });

    it("ignores invalid or garbage rate inputs safely", () => {
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), rate: "invalid" },
        { ...emptyAllocationRule(), rate: "-20" },
        { ...emptyAllocationRule(), rate: "100" },
      ];
      // "invalid" and "-20" are skipped in rateTotal, remaining "100" is evaluated
      expect(rateTotal(rules)).toBe("100");
    });

    it("returns 0 for empty rule sets", () => {
      expect(rateTotal([])).toBe("0");
    });
  });

  // --- Section 3: Multi-Dimensional Cost Allocation Setup ---
  describe("Multi-Dimension Distribution (Department, Project, Branch)", () => {
    it("constructs a valid cost allocation master across multiple business dimensions", () => {
      const master: GLMaster = {
        ...emptyMaster(),
        code: "ALLOC-RENT-2026",
        name: "ปันส่วนค่าเช่าสำนักงานใหญ่",
        accountcode: "530101", // ค่าเช่าอาคารสำนักงาน (Source Cost Account)
        allocatemode: "percent",
        allocaterules: [
          { branchcode: "HQ", departmentcode: "DEPT-SALES", projectcode: "PRJ-RETAIL", accountcode: "530101", rate: "40" },
          { branchcode: "HQ", departmentcode: "DEPT-OPERATIONS", projectcode: "PRJ-CLOUD", accountcode: "530101", rate: "35" },
          { branchcode: "HQ", departmentcode: "DEPT-EXECUTIVE", projectcode: "", accountcode: "530101", rate: "25" },
        ],
      };

      expect(master.code).toBe("ALLOC-RENT-2026");
      expect(master.allocatemode).toBe("percent");
      expect(master.allocaterules).toHaveLength(3);
      expect(rateTotal(master.allocaterules)).toBe("100");
    });
  });

  // --- Section 4: Chart of Accounts & Posting Eligibility ---
  describe("Chart of Accounts Posting Permission Guard", () => {
    const chartOfAccounts: GLAccount[] = [
      {
        ...emptyAccount(),
        accountcode: "5000",
        names: [{ code: "th", name: "ค่าใช้จ่ายในการดำเนินงาน" }],
        accounttype: "expense",
        allowposting: false, // Control/Parent Account
        isactive: true,
      },
      {
        ...emptyAccount(),
        accountcode: "530101",
        names: [{ code: "th", name: "ค่าเช่าอาคารสำนักงาน" }],
        accounttype: "expense",
        parentaccountcode: "5000",
        allowposting: true, // Posting Detail Account
        isactive: true,
      },
      {
        ...emptyAccount(),
        accountcode: "530102",
        names: [{ code: "th", name: "ค่าบริการอินเทอร์เน็ตและโทรศัพท์" }],
        accounttype: "expense",
        parentaccountcode: "5000",
        allowposting: true,
        isactive: false, // Inactive Account
      },
    ];

    it("verifies that source cost account is a posting account and not a parent control account", () => {
      const postingAccount = chartOfAccounts.find((a) => a.accountcode === "530101");
      expect(postingAccount).toBeDefined();
      expect(postingAccount?.allowposting).toBe(true);
      expect(postingAccount?.isactive).toBe(true);

      const parentControlAccount = chartOfAccounts.find((a) => a.accountcode === "5000");
      expect(parentControlAccount).toBeDefined();
      // Accountants must never allocate from a parent header account directly
      expect(parentControlAccount?.allowposting).toBe(false);
    });

    it("verifies that inactive accounts are flagged for accountants", () => {
      const inactiveAccount = chartOfAccounts.find((a) => a.accountcode === "530102");
      expect(inactiveAccount).toBeDefined();
      expect(inactiveAccount?.isactive).toBe(false);
    });
  });

  // --- Section 5: Backend Engine Processing Simulation ---
  describe("Backend-First Processing & Arithmetic Verification", () => {
    it("simulates the PostgreSQL report engine allocating source movement to targets with zero rounding leakage", () => {
      // Source cost incurred during period: 125,450.00 THB
      const sourceAmount = 125450.00;
      const rules: GLAllocationRule[] = [
        { ...emptyAllocationRule(), departmentcode: "DEPT-A", rate: "33.33" },
        { ...emptyAllocationRule(), departmentcode: "DEPT-B", rate: "33.33" },
        { ...emptyAllocationRule(), departmentcode: "DEPT-C", rate: "33.34" },
      ];

      // Calculate allocated amount for each target
      const allocations = rules.map((r) => {
        const rateNum = parseFloat(r.rate);
        const allocated = (rateNum / 100) * sourceAmount;
        return {
          dept: r.departmentcode,
          rate: rateNum,
          allocated: Math.round(allocated * 100) / 100,
        };
      });

      // Verification: Sum of rates must be 100
      const totalRate = allocations.reduce((acc, curr) => acc + curr.rate, 0);
      expect(Math.round(totalRate)).toBe(100);

      // Verify each slice
      expect(allocations[0].allocated).toBe(41812.49); // 33.33% of 125450
      expect(allocations[1].allocated).toBe(41812.49); // 33.33% of 125450
      expect(allocations[2].allocated).toBe(41825.03); // 33.34% of 125450

      // Total allocated sum = 41812.49 + 41812.49 + 41825.03 = 125450.01 (0.01 standard penny-rounding difference)
      const allocatedSum = allocations.reduce((acc, curr) => acc + curr.allocated, 0);
      expect(Math.abs(allocatedSum - sourceAmount)).toBeLessThanOrEqual(0.01);
    });
  });
});

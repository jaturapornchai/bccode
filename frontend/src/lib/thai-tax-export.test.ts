import { describe, it, expect } from "vitest";
import {
  validateThaiTaxId,
  normalizeBranchNo,
  parseIndividualName,
  formatThaiTaxDate,
  generateRdPrepPnd3,
  generateRdPrepPnd53,
  generateRdPrepPp30,
} from "./thai-tax-export";
import type { ThaiWhtRecord } from "./thai-wht";
import type { OfficialPp30FormData } from "./thai-vat-reconciliation";

describe("thai-tax-export", () => {
  describe("validateThaiTaxId", () => {
    it("should validate a correct Thai Tax ID / Citizen ID using Mod 11 Checksum", () => {
      // เลข 13 หลักที่ถูกต้องตามสูตร Modulo 11
      // ตัวอย่าง: 1100500123456 -> sum % 11 -> check digit
      // คำนวณ: 1*13 + 1*12 + 0*11 + 0*10 + 5*9 + 0*8 + 0*7 + 1*6 + 2*5 + 3*4 + 4*3 + 5*2
      // = 13 + 12 + 0 + 0 + 45 + 0 + 0 + 6 + 10 + 12 + 12 + 10 = 120
      // 120 % 11 = 10 -> (11 - 10) % 10 = 1 (ดังนั้นหลักสุดท้ายต้องเป็น 1)
      const validId = "1100500123451";
      const result = validateThaiTaxId(validId);
      expect(result.isValid).toBe(true);
    });

    it("should reject an ID with incorrect check digit", () => {
      const invalidId = "1100500123459";
      const result = validateThaiTaxId(invalidId);
      expect(result.isValid).toBe(false);
      expect(result.reason).toContain("เลขตรวจสอบหลักที่ 13 ไม่ถูกต้อง");
    });

    it("should reject an ID with invalid length", () => {
      const shortId = "123456789";
      const result = validateThaiTaxId(shortId);
      expect(result.isValid).toBe(false);
      expect(result.reason).toContain("ความยาวต้องมี 13 หลัก");
    });

    it("should reject repetitive identical digits (e.g. 0000000000000)", () => {
      const repId = "1111111111111";
      const result = validateThaiTaxId(repId);
      expect(result.isValid).toBe(false);
      expect(result.reason).toContain("ตัวเลขเดียวกันซ้ำกัน");
    });
  });

  describe("normalizeBranchNo", () => {
    it("should normalize empty or 0 to 00000", () => {
      expect(normalizeBranchNo("")).toBe("00000");
      expect(normalizeBranchNo("0")).toBe("00000");
      expect(normalizeBranchNo(undefined)).toBe("00000");
    });

    it("should pad numbers to 5 digits", () => {
      expect(normalizeBranchNo("1")).toBe("00001");
      expect(normalizeBranchNo("12")).toBe("00012");
      expect(normalizeBranchNo("00005")).toBe("00005");
    });
  });

  describe("parseIndividualName", () => {
    it("should parse prefix, first name, and last name", () => {
      const res1 = parseIndividualName("นาย สมชาย มั่งมี");
      expect(res1.prefix).toBe("นาย");
      expect(res1.firstName).toBe("สมชาย");
      expect(res1.lastName).toBe("มั่งมี");

      const res2 = parseIndividualName("นางสาวสุดารัตน์ ใจงาม");
      expect(res2.prefix).toBe("นางสาว");
      expect(res2.firstName).toBe("สุดารัตน์");
      expect(res2.lastName).toBe("ใจงาม");

      const res3 = parseIndividualName("ดร. ทักษิณ รุ่งเรือง");
      expect(res3.prefix).toBe("ดร.");
      expect(res3.firstName).toBe("ทักษิณ");
      expect(res3.lastName).toBe("รุ่งเรือง");
    });
  });

  describe("formatThaiTaxDate", () => {
    it("should format YYYY-MM-DD to BE dd/mm/yyyy", () => {
      expect(formatThaiTaxDate("2026-09-18", "dd/mm/yyyy")).toBe("18/09/2569");
      expect(formatThaiTaxDate("2026-09-18", "ddmmyyyy")).toBe("18092569");
    });
  });

  describe("generateRdPrepPnd3", () => {
    it("should generate pipe-delimited text for PND.3 with correct columns", () => {
      const mockRecords: ThaiWhtRecord[] = [
        {
          id: "wht-1",
          docNo: "50TWI-001",
          docDate: "2026-09-15",
          filingType: "pnd3",
          payeeType: "individual",
          payeeTaxId: "1100500123451",
          payeeName: "นาย สมชาย ใจดี",
          payeeAddress: "123 ถ.สุขุมวิท กทม.",
          payeeBranchNo: "00000",
          isHeadOffice: true,
          incomeType: "service_subcontract_40_8",
          incomeDescription: "ค่าที่ปรึกษาการตลาด",
          taxRate: 3,
          paymentAmount: 10000,
          whtAmount: 300,
          condition: "deducted",
          status: "active",
        },
      ];

      const output = generateRdPrepPnd3(mockRecords, { delimiter: "|" });
      const parts = output.split("|");

      expect(parts[0]).toBe("1"); // ลำดับที่
      expect(parts[1]).toBe("1100500123451"); // เลข 13 หลัก
      expect(parts[2]).toBe("00000"); // สาขา
      expect(parts[3]).toBe("นาย"); // คำนำหน้า
      expect(parts[4]).toBe("สมชาย"); // ชื่อ
      expect(parts[5]).toBe("ใจดี"); // นามสกุล
      expect(parts[7]).toBe("15/09/2569"); // วันที่จ่าย พ.ศ.
      expect(parts[9]).toBe("3.00"); // อัตราภาษี
      expect(parts[10]).toBe("10000.00"); // จำนวนเงินที่จ่าย
      expect(parts[11]).toBe("300.00"); // ภาษีที่หัก
      expect(parts[12]).toBe("1"); // เงื่อนไขหัก ณ ที่จ่าย
    });
  });

  describe("generateRdPrepPnd53", () => {
    it("should generate pipe-delimited text for PND.53 corporate records", () => {
      const mockRecords: ThaiWhtRecord[] = [
        {
          id: "wht-2",
          docNo: "50TWI-002",
          docDate: "2026-09-16",
          filingType: "pnd53",
          payeeType: "corporate",
          payeeTaxId: "0105558012345",
          payeeName: "บริษัท บิลดิ้ง เซอร์วิสเซส จำกัด",
          payeeAddress: "99 ถ.สีลม บางรัก กทม.",
          payeeBranchNo: "00002",
          isHeadOffice: false,
          incomeType: "rent_40_5",
          incomeDescription: "ค่าเช่าสำนักงาน",
          taxRate: 5,
          paymentAmount: 20000,
          whtAmount: 1000,
          condition: "deducted",
          status: "active",
        },
      ];

      const output = generateRdPrepPnd53(mockRecords, { delimiter: "|" });
      const parts = output.split("|");

      expect(parts[0]).toBe("1");
      expect(parts[1]).toBe("0105558012345");
      expect(parts[2]).toBe("00002");
      expect(parts[3]).toBe("บริษัท บิลดิ้ง เซอร์วิสเซส จำกัด");
      expect(parts[5]).toBe("16/09/2569");
      expect(parts[7]).toBe("5.00");
      expect(parts[8]).toBe("20000.00");
      expect(parts[9]).toBe("1000.00");
      expect(parts[10]).toBe("1");
    });
  });

  describe("generateRdPrepPp30", () => {
    it("should generate pipe-delimited summary for PP.30 e-filing", () => {
      const mockPp30: OfficialPp30FormData = {
        taxId: "0105559123456",
        companyName: "บริษัท บีซี คลาวด์ แอคเคาท์ จำกัด",
        branchNo: "00000",
        isHeadOffice: true,
        taxPeriodMonth: 9,
        taxPeriodYearBe: 2569,
        filingType: "normal",
        item1_grossSales: 500000,
        item2_zeroRatedSales: 0,
        item3_exemptSales: 0,
        item4_taxableSales: 500000,
        item5_outputVat: 35000,
        item6_taxablePurchases: 200000,
        item7_inputVat: 14000,
        item8_creditBroughtForward: 2000,
        item9_vatPayable: 19000,
        item10_vatOverpaid: 0,
        surcharge: 0,
        penalty: 0,
        netPaymentAmount: 19000,
      };

      const output = generateRdPrepPp30(mockPp30, { delimiter: "|" });
      const parts = output.split("|");

      expect(parts[0]).toBe("0105559123456");
      expect(parts[1]).toBe("00000");
      expect(parts[2]).toBe("09");
      expect(parts[3]).toBe("2569");
      expect(parts[4]).toBe("0"); // ปกติ
      expect(parts[8]).toBe("500000.00"); // item 4
      expect(parts[9]).toBe("35000.00"); // item 5 (Output VAT)
      expect(parts[11]).toBe("14000.00"); // item 7 (Input VAT)
      expect(parts[12]).toBe("2000.00"); // item 8 (Credit brought forward)
      expect(parts[13]).toBe("19000.00"); // item 9 (VAT Payable)
      expect(parts[17]).toBe("19000.00"); // Net
    });
  });
});

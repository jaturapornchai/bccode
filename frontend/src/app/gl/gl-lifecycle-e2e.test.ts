import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import {
  type GLAccount,
  type GLFiscalYear,
  type GLMaster,
  type GLJournal,
  type GLReport,
  journalTotals,
  accountName,
  emptyAccount,
  emptyFiscalYear,
  emptyMaster,
  emptyJournal,
  validateJournal,
  formatAmount,
  reportCsv,
  GL_RESOURCES,
  GL_REPORTS,
  type GLJournalBook,
  journalBookTypeLabels,
} from "@/lib/general-ledger";
import { parseClipboardJournalLines } from "@/lib/clipboard-journal-parser";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { ReportGrid } from "./gl-reports";
import { GLMasters, normalizeRecord, editorAlert, errorStatePatch, saveFailureTarget } from "./gl-masters";
import { GLJournals } from "./gl-journals";
import { GLProcesses } from "./gl-processes";
import * as glCommon from "./gl-common";

vi.mock("./gl-common", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./gl-common")>();
  return {
    ...actual,
    useGLList: vi.fn(),
    useReferences: vi.fn().mockReturnValue({
      accounts: [
        { id: "a-1", accountcode: "1100", names: [{ code: "th", name: "สินทรัพย์หมุนเวียน" }], accounttype: "asset", allowposting: false, isactive: true, level: 1 },
        { id: "a-2", accountcode: "1111-01", names: [{ code: "th", name: "เงินสดในมือ" }], accounttype: "asset", parentaccountcode: "1100", normalbalance: "debit", allowposting: true, isactive: true, level: 2 },
        { id: "a-3", accountcode: "1112-01", names: [{ code: "th", name: "เงินฝากธนาคารกระแสรายวัน" }], accounttype: "asset", parentaccountcode: "1100", normalbalance: "debit", allowposting: true, isactive: true, level: 2 },
        { id: "a-4", accountcode: "2100", names: [{ code: "th", name: "เจ้าหนี้การค้า" }], accounttype: "liability", normalbalance: "credit", allowposting: true, isactive: true, level: 2 },
        { id: "a-5", accountcode: "3100", names: [{ code: "th", name: "ทุนเรือนหุ้น" }], accounttype: "equity", normalbalance: "credit", allowposting: true, isactive: true, level: 2 },
        { id: "a-6", accountcode: "3200", names: [{ code: "th", name: "กำไรสะสม" }], accounttype: "equity", normalbalance: "credit", allowposting: true, isactive: true, level: 2 },
        { id: "a-7", accountcode: "4100", names: [{ code: "th", name: "รายได้จากการขาย" }], accounttype: "income", normalbalance: "credit", allowposting: true, isactive: true, level: 2 },
        { id: "a-8", accountcode: "5100", names: [{ code: "th", name: "ต้นทุนขาย" }], accounttype: "expense", normalbalance: "debit", allowposting: true, isactive: true, level: 2 },
        { id: "a-9", accountcode: "5200", names: [{ code: "th", name: "ค่าใช้จ่ายในการบริหาร" }], accounttype: "expense", normalbalance: "debit", allowposting: true, isactive: true, level: 2 },
      ],
      years: [
        { id: "y-2026", code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, retainedearningsaccount: "3200", profitlossaccount: "3200", isactive: true, closed: false },
      ],
      books: [
        { id: "jb-1", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true },
        { id: "jb-2", code: "PV", name: "สมุดรายวันจ่ายเงิน", booktype: 2, isactive: true },
        { id: "jb-3", code: "RV", name: "สมุดรายวันรับเงิน", booktype: 3, isactive: true },
        { id: "jb-4", code: "SV", name: "สมุดรายวันขาย", booktype: 4, isactive: true },
        { id: "jb-5", code: "UV", name: "สมุดรายวันซื้อ", booktype: 5, isactive: true },
      ],
      error: "",
      reload: vi.fn(),
    }),
    useGLCommand: vi.fn().mockReturnValue({ busy: false, execute: vi.fn() }),
    useDirtyGuard: vi.fn(),
  };
});

describe("General Ledger Full Lifecycle End-to-End Simulation", () => {
  describe("1. Configuration & Master Data Setup", () => {
    it("verifies all GL resources and reports are defined without omissions", () => {
      expect(GL_RESOURCES).toContain("accounts");
      expect(GL_RESOURCES).toContain("fiscal-years");
      expect(GL_RESOURCES).toContain("journal-books");
      expect(GL_RESOURCES).toContain("account-groups");
      expect(GL_RESOURCES).toContain("mappings");
      expect(GL_RESOURCES).toContain("budgets");
      expect(GL_RESOURCES).toContain("periods");
      expect(GL_RESOURCES).toContain("allocations");
      expect(GL_RESOURCES).toContain("journals");

      expect(GL_REPORTS).toContain("ledger");
      expect(GL_REPORTS).toContain("trialbalance");
      expect(GL_REPORTS).toContain("pnl");
      expect(GL_REPORTS).toContain("balancesheet");
      expect(GL_REPORTS).toContain("workingpaper");
      expect(GL_REPORTS).toContain("gljournal");
      expect(GL_REPORTS).toContain("budgetcomparison");
    });

    it("ensures GL module routes are connected (22 of 23; monthly budget entry screen pending)", () => {
      const glRoutes = [
        "/gl/chartofaccounts",
        "/gl/fiscal-years",
        "/gl/journal-books",
        "/gl/account-groups",
        "/gl/account-mapping",
        "/gl/product-account-groups",
        "/gl/periodlock",
        "/gl/statement-designer",
        "/gl/journals",
        "/gl/journal/jv",
        "/gl/journal/uv",
        "/gl/journal/sv",
        "/gl/journal/rv",
        "/gl/journal/pv",
        "/gl/openingbalance",
        "/gl/posting",
        "/gl/financialclose",
        "/gl/year-end",
        "/gl/recalculate-posted",
        "/gl/reprocess",
        "/report/gljournal",
        "/report/budgetcomparison",
      ];
      for (const route of glRoutes) {
        expect(isMenuScreenPending(route)).toBe(false);
      }
      // 2026-09-25: monthly budgets moved to their own API; the entry screen is the next step.
      expect(isMenuScreenPending("/gl/budget")).toBe(true);
    });

    it("verifies account normalization and hierarchy constraints", () => {
      const parent = normalizeRecord("accounts", {
        accountcode: "1000",
        names: [{ code: "th", name: "สินทรัพย์รวม" }],
        accounttype: "asset",
        allowposting: false,
        level: 1,
      } as GLAccount) as GLAccount;

      expect(parent.allowposting).toBe(false);
      expect(parent.level).toBe(1);
      expect(accountName(parent)).toBe("สินทรัพย์รวม");

      const child = normalizeRecord("accounts", {
        accountcode: "1111-01",
        parentaccountcode: "1000",
        names: [{ code: "th", name: "เงินสด" }],
        accounttype: "asset",
        allowposting: true,
        level: 2,
      } as GLAccount) as GLAccount;

      expect(child.allowposting).toBe(true);
      expect(child.parentaccountcode).toBe("1000");
    });

    it("verifies fiscal year parameters and date intervals", () => {
      const fy: GLFiscalYear = {
        code: "2026",
        startdate: "2026-01-01",
        enddate: "2026-12-31",
        scale: 2,
        retainedearningsaccount: "3200",
        profitlossaccount: "3200",
        isactive: true,
        closed: false,
      };
      expect(fy.scale).toBe(2);
      expect(fy.startdate < fy.enddate).toBe(true);
      expect(fy.retainedearningsaccount).toBe("3200");
    });

    it("verifies journal books definition and book labels", () => {
      // Book types are fixed by the spec; codes are user-defined (standard defaults JV/PV/RV/SV/UV)
      expect(journalBookTypeLabels[1][1]).toBe("รายวันทั่วไป");
      expect(journalBookTypeLabels[2][1]).toBe("รายวันจ่ายเงิน");
      expect(journalBookTypeLabels[3][1]).toBe("รายวันรับเงิน");
      expect(journalBookTypeLabels[4][1]).toBe("รายวันขาย");
      expect(journalBookTypeLabels[5][1]).toBe("รายวันซื้อ");
      expect(journalBookTypeLabels[6][1]).toBe("ยอดยกมา");

      vi.mocked(glCommon.useGLList).mockReturnValue({
        data: {
          items: [
            { id: "jb-1", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true } as GLJournalBook,
            { id: "jb-2", code: "PV", name: "สมุดรายวันจ่ายเงิน", booktype: 2, isactive: true } as GLJournalBook,
            { id: "jb-3", code: "RV", name: "สมุดรายวันรับเงิน", booktype: 3, isactive: true } as GLJournalBook,
            { id: "jb-4", code: "SV", name: "สมุดรายวันขาย", booktype: 4, isactive: true } as GLJournalBook,
            { id: "jb-5", code: "UV", name: "สมุดรายวันซื้อ", booktype: 5, isactive: true } as GLJournalBook,
          ],
          total: 5,
          page: 1,
          limit: 30,
          sequence: 1,
        },
        page: 1,
        loading: false,
        error: "",
        reload: vi.fn(),
        setPage: vi.fn(),
      });

      const html = renderToStaticMarkup(createElement(GLMasters, { resource: "journal-books", route: "/gl/journal-books" }));
      expect(html).toContain("JV");
      expect(html).toContain("UV");
      expect(html).toContain("SV");
      expect(html).toContain("RV");
      expect(html).toContain("PV");
    });
  });

  describe("2. Journal Entry Operations & Validation Rules", () => {
    it("validates journal debit-credit balance check", () => {
      const balancedJournal: GLJournal = {
        docno: "JV202601001",
        date: "2026-01-15",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "รับชำระค่าหุ้นเพิ่มทุนเป็นเงินสด",
        reference: "REF-001",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1111-01", description: "เงินสด", debit: "100000.00", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "financing" },
          { accountcode: "3100", description: "ทุนเรือนหุ้น", debit: "0.00", credit: "100000.00", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      const totals = journalTotals(balancedJournal.lines);
      expect(totals.debit).toBe(10000000000000n);
      expect(totals.credit).toBe(10000000000000n);
      expect(totals.debit === totals.credit).toBe(true);

      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1111-01", isactive: true, allowposting: true } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(balancedJournal, fy, accList);
      expect(err).toBeNull();
    });

    it("rejects unbalanced journal entries (debit != credit)", () => {
      const unbalancedJournal: GLJournal = {
        docno: "JV202601002",
        date: "2026-01-15",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "รายการไม่สมดุล",
        reference: "",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1111-01", description: "เงินสด", debit: "50000.00", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "" },
          { accountcode: "3100", description: "ทุน", debit: "0.00", credit: "40000.00", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1111-01", isactive: true, allowposting: true } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(unbalancedJournal, fy, accList);
      expect(err).toContain("ยอดเดบิตและเครดิตต้องเท่ากันก่อนบันทึก");
    });

    it("rejects posting to non-allowposting accounts (control accounts)", () => {
      const invalidJournal: GLJournal = {
        docno: "JV202601003",
        date: "2026-01-15",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "ลงบัญชีคุม",
        reference: "",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1100", description: "สินทรัพย์คุมยอด", debit: "10000.00", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "" },
          { accountcode: "3100", description: "ทุน", debit: "0.00", credit: "10000.00", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1100", isactive: true, allowposting: false } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(invalidJournal, fy, accList);
      expect(err).toContain("เลือกบัญชีที่เปิดใช้งานและลงรายการได้");
    });

    it("rejects dates outside the fiscal year", () => {
      const outOfYearJournal: GLJournal = {
        docno: "JV202601004",
        date: "2025-12-31",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "ลงวันที่ปีก่อนหน้า",
        reference: "",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1111-01", description: "เงินสด", debit: "1000.00", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "" },
          { accountcode: "3100", description: "ทุน", debit: "0.00", credit: "1000.00", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1111-01", isactive: true, allowposting: true } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(outOfYearJournal, fy, accList);
      expect(err).toContain("กรุณาเลือกปีบัญชีที่เปิดใช้งานและวันที่ภายในปีบัญชี");
    });

    it("parses and validates tabular journal paste from Excel", () => {
      const excelText = "1111-01\tเงินสด\t15000.00\t0.00\tแผนกขาย\tโครงการ A\n4100\tขายสินค้า\t0.00\t15000.00\tแผนกขาย\tโครงการ A";
      const { lines: parsed, issues } = parseClipboardJournalLines(excelText);
      expect(issues).toEqual([]);
      expect(parsed.length).toBe(2);
      expect(parsed[0].accountcode).toBe("1111-01");
      expect(parsed[0].debit).toBe("15000.00");
      expect(parsed[0].credit).toBe("0");
      expect(parsed[1].accountcode).toBe("4100");
      expect(parsed[1].debit).toBe("0");
      expect(parsed[1].credit).toBe("15000.00");

      const totals = journalTotals(parsed);
      expect(totals.debit === totals.credit).toBe(true);
    });

    it("protects posted entries from direct mutations (immutable safety guard)", () => {
      vi.mocked(glCommon.useGLList).mockReturnValue({
        data: {
          items: [
            {
              id: "j-posted-1",
              docno: "JV202601001",
              date: "2026-01-15",
              bookcode: "JV",
              status: "posted",
              description: "ผ่านบัญชีแล้ว ห้ามแก้ไข",
              lines: [
                { accountcode: "1111-01", debit: "5000.00", credit: "0.00" },
                { accountcode: "4100", debit: "0.00", credit: "5000.00" },
              ],
            } as GLJournal,
          ],
          total: 1,
          page: 1,
          limit: 30,
          sequence: 1,
        },
        page: 1,
        loading: false,
        error: "",
        reload: vi.fn(),
        setPage: vi.fn(),
      });

      const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/journal/jv", book: "JV" }));
      expect(html).toContain("JV202601001");
      expect(html).toContain("ผ่านรายการแล้ว");
      // In posted state, direct deletion is not offered in table rows
      expect(html).not.toContain("title=\"ลบ (Delete)\"");
    });
  });

  describe("3. Financial Processes Execution", () => {
    it("renders GL process controls for recalculation, financial close, and year-end", () => {
      const htmlRecalc = renderToStaticMarkup(createElement(GLProcesses, { route: "/gl/recalculate-posted", action: "recalculate" }));
      expect(htmlRecalc).toContain("คำนวณยอดผ่านรายการใหม่");
      expect(htmlRecalc).toContain("ตรวจสอบยอดก่อนดำเนินการ");

      const htmlClose = renderToStaticMarkup(createElement(GLProcesses, { route: "/gl/financialclose", action: "close" }));
      expect(htmlClose).toContain("สร้างฉบับร่างปิดงวด");

      const htmlYearEnd = renderToStaticMarkup(createElement(GLProcesses, { route: "/gl/year-end", action: "year-end" }));
      expect(htmlYearEnd).toContain("ประมวลผลสิ้นปี");
    });
  });

  describe("4. Financial & Analytical Reports Parity", () => {
    it("renders Trial Balance report with opening, movement, ending, and balance check", () => {
      const trialReport: GLReport = {
        sequence: 10,
        columns: [
          { key: "accountcode", label: "รหัสบัญชี" },
          { key: "accountname", label: "ชื่อบัญชี" },
          { key: "accounttype", label: "หมวดบัญชี" },
          { key: "openingdebit", label: "ยกมาเดบิต", amount: true },
          { key: "openingcredit", label: "ยกมาเครดิต", amount: true },
          { key: "debit", label: "เดบิต", amount: true },
          { key: "credit", label: "เครดิต", amount: true },
          { key: "endingdebit", label: "คงเหลือเดบิต", amount: true },
          { key: "endingcredit", label: "คงเหลือเครดิต", amount: true },
        ],
        rows: [
          { accountcode: "1111-01", accountname: "เงินสดในมือ", accounttype: "asset", openingdebit: "10000.00", openingcredit: "0.00", debit: "5000.00", credit: "2000.00", endingdebit: "13000.00", endingcredit: "0.00" },
          { accountcode: "4100", accountname: "รายได้จากการขาย", accounttype: "income", openingdebit: "0.00", openingcredit: "0.00", debit: "0.00", credit: "5000.00", endingdebit: "0.00", endingcredit: "5000.00" },
          { accountcode: "5100", accountname: "ต้นทุนขาย", accounttype: "expense", openingdebit: "0.00", openingcredit: "0.00", debit: "2000.00", credit: "0.00", endingdebit: "2000.00", endingcredit: "0.00" },
          { accountcode: "3100", accountname: "ทุนเรือนหุ้น", accounttype: "equity", openingdebit: "0.00", openingcredit: "10000.00", debit: "0.00", credit: "0.00", endingdebit: "0.00", endingcredit: "10000.00" },
        ],
        totals: {
          openingdebit: "10000.00",
          openingcredit: "10000.00",
          debit: "7000.00",
          credit: "7000.00",
          endingdebit: "15000.00",
          endingcredit: "15000.00",
          difference: "0.00",
        },
        totalrows: 4,
        warnings: [],
        asof: "2026-09-19 08:00:00",
      };

      const html = renderToStaticMarkup(createElement(ReportGrid, { report: trialReport }));
      expect(html).toContain("1111-01");
      expect(html).toContain("เงินสดในมือ");
      expect(html).toContain("13,000.00");
      expect(html).toContain("4100");
      expect(html).toContain("รายได้จากการขาย");
      expect(html).toContain("5,000.00");
      expect(html).toContain("15,000.00"); // totals endingdebit and endingcredit
    });

    it("renders Profit & Loss statement with revenues, expenses, and net profit", () => {
      const pnlReport: GLReport = {
        sequence: 11,
        columns: [
          { key: "accountcode", label: "รหัสบัญชี" },
          { key: "accountname", label: "ชื่อบัญชี" },
          { key: "amount", label: "จำนวนเงิน", amount: true },
        ],
        rows: [
          { accountcode: "4100", accountname: "รายได้จากการขาย", amount: "500000.00" },
          { accountcode: "5100", accountname: "ต้นทุนขาย", amount: "-300000.00" },
          { accountcode: "5200", accountname: "ค่าใช้จ่ายในการบริหาร", amount: "-50000.00" },
        ],
        totals: {
          revenue: "500000.00",
          expense: "350000.00",
          profit: "150000.00",
        },
        totalrows: 3,
        warnings: [],
        asof: "2026-09-19",
      };

      const html = renderToStaticMarkup(createElement(ReportGrid, { report: pnlReport }));
      expect(html).toContain("รายได้จากการขาย");
      expect(html).toContain("500,000.00");
      expect(html).toContain("ต้นทุนขาย");
      expect(html).toContain("-300,000.00");
      expect(html).toContain("กำไรสุทธิ");
      expect(html).toContain("150,000.00");
    });

    it("renders Balance Sheet with Assets = Liabilities + Equity balance verification", () => {
      const bsReport: GLReport = {
        sequence: 12,
        columns: [
          { key: "accountcode", label: "รหัสบัญชี" },
          { key: "accountname", label: "ชื่อบัญชี" },
          { key: "amount", label: "ยอดคงเหลือ", amount: true },
        ],
        rows: [
          { accountcode: "1111-01", accountname: "เงินสดในมือ", amount: "200000.00" },
          { accountcode: "1112-01", accountname: "เงินฝากธนาคาร", amount: "400000.00" },
          { accountcode: "2100", accountname: "เจ้าหนี้การค้า", amount: "100000.00" },
          { accountcode: "3100", accountname: "ทุนเรือนหุ้น", amount: "350000.00" },
          { accountcode: "__current_earnings__", accountname: "กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม", amount: "150000.00" },
        ],
        totals: {
          assets: "600000.00",
          liabilities: "100000.00",
          equity: "500000.00", // 350000 + 150000
          difference: "0.00",   // assets - (liabilities + equity) = 0
        },
        totalrows: 5,
        warnings: [],
        asof: "2026-09-19",
      };

      const html = renderToStaticMarkup(createElement(ReportGrid, { report: bsReport }));
      expect(html).toContain("สินทรัพย์");
      expect(html).toContain("600,000.00");
      expect(html).toContain("หนี้สิน");
      expect(html).toContain("100,000.00");
      expect(html).toContain("ส่วนของเจ้าของ");
      expect(html).toContain("500,000.00");
      expect(html).toContain("0.00");
    });

    it("renders Journal Entries report (gljournal / Champ 104003) with vouchers & lines", () => {
      const glJournalReport: GLReport = {
        sequence: 13,
        columns: [
          { key: "date", label: "วันที่" },
          { key: "docno", label: "เลขที่เอกสาร" },
          { key: "bookcode", label: "สมุดรายวัน" },
          { key: "accountcode", label: "รหัสบัญชี" },
          { key: "accountname", label: "ชื่อบัญชี" },
          { key: "description", label: "คำอธิบาย" },
          { key: "debit", label: "เดบิต", amount: true },
          { key: "credit", label: "เครดิต", amount: true },
        ],
        rows: [
          { date: "2026-01-10", docno: "JV202601001", bookcode: "JV", accountcode: "1111-01", accountname: "เงินสดในมือ", description: "ตั้งวงเงินสดย่อย", debit: "20000.00", credit: "0.00" },
          { date: "2026-01-10", docno: "JV202601001", bookcode: "JV", accountcode: "1112-01", accountname: "เงินฝากธนาคาร", description: "ถอนเงินสดเข้ามือ", debit: "0.00", credit: "20000.00" },
        ],
        totals: {
          debit: "20000.00",
          credit: "20000.00",
        },
        totalrows: 2,
        warnings: [],
        asof: "2026-09-19",
      };

      const html = renderToStaticMarkup(createElement(ReportGrid, { report: glJournalReport }));
      expect(html).toContain("JV202601001");
      expect(html).toContain("1111-01");
      expect(html).toContain("เงินสดในมือ");
      expect(html).toContain("20,000.00");
    });

    it("renders Budget Comparison report (budgetcomparison / Champ 104017)", () => {
      const budgetReport: GLReport = {
        sequence: 14,
        columns: [
          { key: "accountcode", label: "รหัสบัญชี" },
          { key: "accountname", label: "ชื่อบัญชี" },
          { key: "budgetamount", label: "งบประมาณ", amount: true },
          { key: "actualamount", label: "ใช้จริง", amount: true },
          { key: "variance", label: "ผลต่างคงเหลือ", amount: true },
          { key: "percentused", label: "ร้อยละที่ใช้ (%)", amount: true },
        ],
        rows: [
          { accountcode: "5200", accountname: "ค่าใช้จ่ายในการบริหาร", budgetamount: "100000.00", actualamount: "65000.00", variance: "35000.00", percentused: "65.00" },
        ],
        totals: {
          budgetamount: "100000.00",
          actualamount: "65000.00",
          variance: "35000.00",
        },
        totalrows: 1,
        warnings: [],
        asof: "2026-09-19",
      };

      const html = renderToStaticMarkup(createElement(ReportGrid, { report: budgetReport }));
      expect(html).toContain("5200");
      expect(html).toContain("ค่าใช้จ่ายในการบริหาร");
      expect(html).toContain("100,000.00");
      expect(html).toContain("65,000.00");
      expect(html).toContain("35,000.00");
      expect(html).toContain("65.00");
    });

    it("verifies accurate CSV export serialization", () => {
      const report: GLReport = {
        sequence: 15,
        columns: [
          { key: "accountcode", label: "รหัสบัญชี" },
          { key: "accountname", label: "ชื่อบัญชี" },
          { key: "amount", label: "จำนวนเงิน", amount: true },
        ],
        rows: [
          { accountcode: "1111-01", accountname: "เงินสด", amount: "1500.50" },
        ],
        totals: { amount: "1500.50" },
        totalrows: 1,
        warnings: [],
        asof: "2026-09-19",
      };

      const csv = reportCsv(report);
      expect(csv).toContain('"รหัสบัญชี","ชื่อบัญชี","จำนวนเงิน"');
      expect(csv).toContain('"1111-01","เงินสด","1500.50"');
    });
  });

  describe("5. Comprehensive UI Controls, Buttons, and Operational Conditions", () => {
    it("tests GLMasters buttons: Add, Search, Reload, Density, and List/Tree toggle", () => {
      vi.mocked(glCommon.useGLList).mockReturnValue({
        data: {
          items: [
            { id: "a-1", accountcode: "1111-01", names: [{ code: "th", name: "เงินสดในมือ" }], accounttype: "asset", isactive: true, allowposting: true, level: 2 } as GLAccount,
          ],
          total: 1,
          page: 1,
          limit: 30,
          sequence: 1,
        },
        page: 1,
        loading: false,
        error: "",
        reload: vi.fn(),
        setPage: vi.fn(),
      });

      const html = renderToStaticMarkup(createElement(GLMasters, { resource: "accounts", route: "/gl/chartofaccounts" }));
      expect(html).toContain("เพิ่มรายการ");
      expect(html).toContain("ค้นหา");
      expect(html).toContain("โหลดใหม่");
      expect(html).toContain("มุมมองผังต้นไม้");
      expect(html).toContain("เงินสดในมือ");
    });

    it("tests Period Lock buttons and operational actions (Lock / Unlock Period)", () => {
      vi.mocked(glCommon.useGLList).mockReturnValue({
        data: {
          items: [
            { id: "p-1", code: "P01", name: "งวดมกราคม 2026", locked: true, startdate: "2026-01-01", enddate: "2026-01-31", isactive: true } as GLMaster,
            { id: "p-2", code: "P02", name: "งวดกุมภาพันธ์ 2026", locked: false, startdate: "2026-02-01", enddate: "2026-02-28", isactive: true } as GLMaster,
          ],
          total: 2,
          page: 1,
          limit: 30,
          sequence: 1,
        },
        page: 1,
        loading: false,
        error: "",
        reload: vi.fn(),
        setPage: vi.fn(),
      });

      const html = renderToStaticMarkup(createElement(GLMasters, { resource: "periods", route: "/gl/periodlock" }));
      expect(html).toContain("P01");
      expect(html).toContain("งวดมกราคม 2026");
      expect(html).toContain("P02");
      expect(html).toContain("งวดกุมภาพันธ์ 2026");
    });

    it("tests GLJournals posting and unposting mode screens", () => {
      vi.mocked(glCommon.useGLList).mockReturnValue({
        data: {
          items: [
            { id: "j-draft", docno: "JV-DRAFT-01", date: "2026-01-20", bookcode: "JV", status: "draft", description: "รอดำเนินการผ่านบัญชี", lines: [] } as unknown as GLJournal,
          ],
          total: 1,
          page: 1,
          limit: 30,
          sequence: 1,
        },
        page: 1,
        loading: false,
        error: "",
        reload: vi.fn(),
        setPage: vi.fn(),
      });

      const htmlPost = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/posting", mode: "post" }));
      expect(htmlPost).toContain("ผ่านรายการ");
      expect(htmlPost).toContain("JV-DRAFT-01");

      const htmlUnpost = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/unposting", mode: "reverse" }));
      expect(htmlUnpost).toContain("ผ่านรายการแล้ว (ขอยกเลิก)");
    });

    it("tests GLJournals opening balance mode screen", () => {
      vi.mocked(glCommon.useGLList).mockReturnValue({
        data: {
          items: [
            { id: "j-open", docno: "OB2026", date: "2026-01-01", bookcode: "JV", kind: "opening", status: "draft", description: "ยอดยกมาต้นปี 2026", lines: [] } as unknown as GLJournal,
          ],
          total: 1,
          page: 1,
          limit: 30,
          sequence: 1,
        },
        page: 1,
        loading: false,
        error: "",
        reload: vi.fn(),
        setPage: vi.fn(),
      });

      const htmlOpening = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/openingbalance", kind: "opening" }));
      expect(htmlOpening).toContain("OB2026");
      expect(htmlOpening).toContain("ยอดยกมาต้นปี 2026");
    });

    it("verifies accounting balance edge cases: zero debit & credit rejected", () => {
      const zeroLineJournal: GLJournal = {
        docno: "JV-ZERO",
        date: "2026-01-15",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "ยอดศูนย์สองฝั่ง",
        reference: "",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1111-01", description: "เงินสด", debit: "0.00", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "" },
          { accountcode: "3100", description: "ทุน", debit: "0.00", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1111-01", isactive: true, allowposting: true } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(zeroLineJournal, fy, accList);
      expect(err).toContain("ใส่ยอดมากกว่าศูนย์เพียงด้านเดียว");
    });

    it("verifies accounting balance edge cases: both sides filled on same line rejected", () => {
      const bothSidesJournal: GLJournal = {
        docno: "JV-BOTH",
        date: "2026-01-15",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "ใส่ยอดสองด้านในบรรทัดเดียว",
        reference: "",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1111-01", description: "เงินสด", debit: "100.00", credit: "100.00", departmentcode: "", projectcode: "", cashflow: "" },
          { accountcode: "3100", description: "ทุน", debit: "0.00", credit: "100.00", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1111-01", isactive: true, allowposting: true } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(bothSidesJournal, fy, accList);
      expect(err).toContain("ใส่ยอดมากกว่าศูนย์เพียงด้านเดียว");
    });

    it("verifies scale enforcement: fractions exceeding fiscal year scale rejected", () => {
      const excessDecimalsJournal: GLJournal = {
        docno: "JV-DECIMAL",
        date: "2026-01-15",
        bookcode: "JV",
        fiscalyear: "2026",
        description: "ทศนิยมเกิน scale",
        reference: "",
        branchcode: "",
        kind: "general",
        status: "draft",
        lines: [
          { accountcode: "1111-01", description: "เงินสด", debit: "100.123", credit: "0.00", departmentcode: "", projectcode: "", cashflow: "" },
          { accountcode: "3100", description: "ทุน", debit: "0.00", credit: "100.123", departmentcode: "", projectcode: "", cashflow: "" },
        ],
      };

      // Year has scale: 2, but entry has 3 decimals (100.123)
      const fy: GLFiscalYear = { code: "2026", startdate: "2026-01-01", enddate: "2026-12-31", scale: 2, isactive: true, closed: false } as GLFiscalYear;
      const accList: GLAccount[] = [
        { accountcode: "1111-01", isactive: true, allowposting: true } as GLAccount,
        { accountcode: "3100", isactive: true, allowposting: true } as GLAccount,
      ];

      const err = validateJournal(excessDecimalsJournal, fy, accList);
      expect(err).toContain("จำนวนเงินเกิน 2 ตำแหน่งทศนิยม");
    });
  });
});

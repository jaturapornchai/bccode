import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { GLJournal, GLLine } from "@/lib/general-ledger";
import * as glCommon from "./gl-common";
import { GLJournals, createSmartNextLine } from "./gl-journals";

vi.mock("./gl-common", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./gl-common")>();
  return {
    ...actual,
    useGLList: vi.fn(),
    useReferences: vi.fn().mockReturnValue({ accounts: [], years: [], error: "", reload: vi.fn() }),
    useGLCommand: vi.fn().mockReturnValue({ busy: false, execute: vi.fn() }),
    useDirtyGuard: vi.fn(),
  };
});

describe("createSmartNextLine auto-balance helper", () => {
  it("pre-fills credit when debit exceeds credit", () => {
    const lines: GLLine[] = [
      { accountcode: "1110-01", description: "เงินสด", debit: "15000.00", credit: "0", departmentcode: "", projectcode: "", cashflow: "" },
    ];
    const nextLine = createSmartNextLine(lines, 2);
    expect(nextLine.credit).toBe("15000.00");
    expect(nextLine.debit).toBe("0");
  });

  it("pre-fills debit when credit exceeds debit", () => {
    const lines: GLLine[] = [
      { accountcode: "4110-01", description: "รายได้จากการขาย", debit: "0", credit: "25000.50", departmentcode: "", projectcode: "", cashflow: "" },
    ];
    const nextLine = createSmartNextLine(lines, 2);
    expect(nextLine.debit).toBe("25000.50");
    expect(nextLine.credit).toBe("0");
  });

  it("calculates net difference across multiple mixed lines", () => {
    const lines: GLLine[] = [
      { accountcode: "5110-01", description: "ซื้อสินค้า", debit: "10000.00", credit: "0", departmentcode: "", projectcode: "", cashflow: "" },
      { accountcode: "1154-01", description: "ภาษีซื้อ", debit: "700.00", credit: "0", departmentcode: "", projectcode: "", cashflow: "" },
      { accountcode: "1110-01", description: "เงินสดจ่ายบางส่วน", debit: "0", credit: "3000.00", departmentcode: "", projectcode: "", cashflow: "" },
    ];
    const nextLine = createSmartNextLine(lines, 2);
    // Total debit = 10700, total credit = 3000 -> remaining credit needed = 7700
    expect(nextLine.credit).toBe("7700.00");
    expect(nextLine.debit).toBe("0");
  });

  it("returns empty debit and credit when journal is already balanced", () => {
    const lines: GLLine[] = [
      { accountcode: "1110-01", description: "เงินสด", debit: "5000.00", credit: "0", departmentcode: "", projectcode: "", cashflow: "" },
      { accountcode: "4110-01", description: "รายได้", debit: "0", credit: "5000.00", departmentcode: "", projectcode: "", cashflow: "" },
    ];
    const nextLine = createSmartNextLine(lines, 2);
    expect(nextLine.debit).toBe("0");
    expect(nextLine.credit).toBe("0");
  });
});

describe("GLJournals CRUD table presentation", () => {
  it("renders table headers with date/docno, description, status, and action column (จัดการ)", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "j-draft",
            docno: "JV-2026-001",
            date: "2026-09-13",
            fiscalyear: "2026",
            bookcode: "JV",
            description: "บันทึกปรับปรุงรายการ",
            status: "draft",
            version: 1,
            lines: [],
          } as unknown as GLJournal,
          {
            id: "j-posted",
            docno: "JV-2026-002",
            date: "2026-09-13",
            fiscalyear: "2026",
            bookcode: "JV",
            description: "ผ่านรายการแล้ว",
            status: "posted",
            version: 1,
            lines: [],
          } as unknown as GLJournal,
        ],
        total: 2,
        page: 1,
        limit: 30,
        sequence: 0,
      },
      page: 1,
      loading: false,
      error: "",
      reload: vi.fn(),
      setPage: vi.fn(),
    });

    const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/journal/jv", book: "JV" }));
    expect(html).toContain("วันที่ / เลขที่");
    expect(html).toContain("คำอธิบาย");
    expect(html).toContain("สถานะ");
    expect(html).toContain("จัดการ");
    expect(html).toContain("JV-2026-001");
    expect(html).toContain("JV-2026-002");
    expect(html).toContain("ฉบับร่าง");
    expect(html).toContain("ผ่านรายการแล้ว");
    // Draft item has Edit and Delete buttons in action column
    expect(html).toContain("title=\"แก้ไขฉบับร่าง (Edit)\"");
    expect(html).toContain("title=\"ลบฉบับร่าง (Delete)\"");
    // Posted item has View button in action column
    expect(html).toContain("title=\"แสดงข้อมูล (View)\"");
  });

  it("renders empty workbench placeholder prompting selection and no save button initially", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [],
        total: 0,
        page: 1,
        limit: 30,
        sequence: 0,
      },
      page: 1,
      loading: false,
      error: "",
      reload: vi.fn(),
      setPage: vi.fn(),
    });

    const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/journal/jv", book: "JV" }));
    expect(html).toContain("เลือกรายการเพื่อแสดงข้อมูลบัญชี");
    expect(html).toContain("คลิกที่แถวในตารางเพื่อแสดงข้อมูล");
    expect(html).not.toContain("บันทึกฉบับร่าง");
  });

  it("renders journal book category tabs (ทั้งหมด, UV, SV, RV, PV, JV) to separate books inside one screen", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "j-sales",
            docno: "UV-2026-001",
            date: "2026-09-18",
            fiscalyear: "2026",
            bookcode: "UV",
            description: "ขายสินค้าเงินเชื่อ",
            status: "draft",
            version: 1,
            lines: [],
          } as unknown as GLJournal,
        ],
        total: 1,
        page: 1,
        limit: 30,
        sequence: 0,
      },
      page: 1,
      loading: false,
      error: "",
      reload: vi.fn(),
      setPage: vi.fn(),
    });

    const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/journals", book: "" }));
    // Contains all book selector tabs
    expect(html).toContain("ทั้งหมด");
    expect(html).toContain("สมุดรายวันขาย");
    expect(html).toContain("สมุดรายวันซื้อ");
    expect(html).toContain("สมุดรายวันรับเงิน");
    expect(html).toContain("สมุดรายวันจ่ายเงิน");
    expect(html).toContain("สมุดรายวันทั่วไป");
    // Contains book code badge
    expect(html).toContain("UV");
    expect(html).toContain("UV-2026-001");
  });

  it("renders posting dual tabs (รอผ่านรายการ and ผ่านรายการแล้ว) when on posting route", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [],
        total: 0,
        page: 1,
        limit: 30,
        sequence: 0,
      },
      page: 1,
      loading: false,
      error: "",
      reload: vi.fn(),
      setPage: vi.fn(),
    });

    const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/posting", mode: "post" }));
    // Contains Posting & Reversal dual tabs
    expect(html).toContain("รอผ่านรายการ (ฉบับร่าง)");
    expect(html).toContain("ผ่านรายการแล้ว (ขอยกเลิก)");
  });
});

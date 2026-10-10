import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { GLJournal, GLLine } from "@/lib/general-ledger";
import * as glCommon from "./gl-common";
import { GLJournals, createSmartNextLine, journalBookBadgeClass, journalListFilters, parseChampDateSearch, pasteIssueText } from "./gl-journals";

vi.mock("./gl-common", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./gl-common")>();
  return {
    ...actual,
    useGLList: vi.fn(),
    useReferences: vi.fn().mockReturnValue({
      accounts: [],
      years: [],
      // สมุดเป็นข้อมูลหลักที่บริษัทกำหนดเอง — รวมรหัสที่ไม่ใช่มาตรฐาน (POS1) และสมุดที่ปิดใช้งาน (OLD)
      books: [
        { id: "b1", code: "JV", name: "สมุดรายวันทั่วไป", booktype: 1, isactive: true },
        { id: "b2", code: "PV", name: "สมุดรายวันจ่ายเงิน", booktype: 2, isactive: true },
        { id: "b3", code: "RV", name: "สมุดรายวันรับเงิน", booktype: 3, isactive: true },
        { id: "b4", code: "SV", name: "สมุดรายวันขาย", booktype: 4, isactive: true },
        { id: "b5", code: "UV", name: "สมุดรายวันซื้อ", booktype: 5, isactive: true },
        { id: "b6", code: "POS1", name: "สมุดขายหน้าร้าน", booktype: 4, isactive: true },
        { id: "b7", code: "OLD", name: "สมุดเก่าเลิกใช้", booktype: 1, isactive: false },
      ],
      error: "",
      reload: vi.fn(),
    }),
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

  it("renders journal book tabs from the company's active journal-books master (name label, colour by book type)", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "j-sales",
            docno: "SV-2026-001",
            date: "2026-09-18",
            fiscalyear: "2026",
            bookcode: "SV",
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
    // Tabs = "all" + every active book of the company, labelled by the book name
    expect(html).toContain("ทั้งหมด");
    for (const name of ["สมุดรายวันขาย", "สมุดรายวันซื้อ", "สมุดรายวันรับเงิน", "สมุดรายวันจ่ายเงิน", "สมุดรายวันทั่วไป", "สมุดขายหน้าร้าน"]) {
      expect(html).toContain(name);
    }
    expect(html).toContain("POS1");
    // Inactive books are not offered as tabs
    expect(html).not.toContain("สมุดเก่าเลิกใช้");
    // Badge shows the code as text; colour follows the book type (4 = sales), not the code
    expect(html).toContain("SV-2026-001");
    expect(html).toContain(journalBookBadgeClass(4));
  });

  it("the unposting screen has no draft/posted sub-tabs — Champ posts in batches (5512) and reverses one voucher at a time", () => {
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

    const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/unposting", mode: "reverse" }));
    expect(html).not.toContain("รอผ่านรายการ (ฉบับร่าง)");
    expect(html).not.toContain("ผ่านรายการแล้ว (ขอยกเลิก)");
  });
});

describe("journal book badge colours", () => {
  it("uses theme tokens only and depends on the book type, never the code", () => {
    for (const booktype of [undefined, 1, 2, 3, 4, 5, 6]) {
      expect(journalBookBadgeClass(booktype)).not.toMatch(/(blue|rose|emerald|amber|purple|red|green)-\d/);
    }
    expect(journalBookBadgeClass(4)).not.toBe(journalBookBadgeClass(5));
    expect(journalBookBadgeClass(undefined)).toBe(journalBookBadgeClass(1));
  });
});

describe("journal list date typed in the search box (Champ QList 1719-1758)", () => {
  it("a single day in Buddhist or Gregorian year filters that day only", () => {
    expect(parseChampDateSearch("15/10/2569")).toEqual({ from: "2026-10-15", to: "2026-10-15" });
    expect(parseChampDateSearch("15/10/2026")).toEqual({ from: "2026-10-15", to: "2026-10-15" });
    expect(parseChampDateSearch("5.1.2569")).toEqual({ from: "2026-01-05", to: "2026-01-05" });
    expect(parseChampDateSearch(" 05-01-2026 ")).toEqual({ from: "2026-01-05", to: "2026-01-05" });
  });

  it("day 0 means the whole month, with the right last day in leap and normal years", () => {
    expect(parseChampDateSearch("00/10/2569")).toEqual({ from: "2026-10-01", to: "2026-10-31" });
    expect(parseChampDateSearch("0/11/2026")).toEqual({ from: "2026-11-01", to: "2026-11-30" });
    expect(parseChampDateSearch("00/02/2567")).toEqual({ from: "2024-02-01", to: "2024-02-29" });
    expect(parseChampDateSearch("00/02/2569")).toEqual({ from: "2026-02-01", to: "2026-02-28" });
    expect(parseChampDateSearch("00/02/2100")).toEqual({ from: "2100-02-01", to: "2100-02-28" });
    expect(parseChampDateSearch("00/02/2000")).toEqual({ from: "2000-02-01", to: "2000-02-29" });
  });

  it("anything that is not a real date stays a text search", () => {
    for (const text of ["31/02/2569", "30/02/2024", "32/01/2026", "15/13/2026", "15/00/2026", "JV001", "JV-2026-001", "15/10/26", "15/10/20269", "15/10/1899", "15/10/4001", "15/10", "1/2/3/2026", "", "  "]) {
      expect(parseChampDateSearch(text), text).toBeNull();
    }
    expect(parseChampDateSearch("29/02/2567")).toEqual({ from: "2024-02-29", to: "2024-02-29" });
    expect(parseChampDateSearch("29/02/2569")).toBeNull();
  });

  it("sends the range as from/to next to the book, kind and status filters", () => {
    const query = new URLSearchParams(journalListFilters({ bookcode: "JV", kind: "manual", status: "posted", from: "2026-10-01", to: "2026-10-31" }));
    expect(query.get("bookcode")).toBe("JV");
    expect(query.get("kind")).toBe("manual");
    expect(query.get("status")).toBe("posted");
    expect(query.get("from")).toBe("2026-10-01");
    expect(query.get("to")).toBe("2026-10-31");
    expect(journalListFilters({ bookcode: "JV" })).toBe("bookcode=JV");
    expect(journalListFilters({})).toBe("");
  });

  it("has no date boxes: the search box carries the hint and the list starts without a date filter", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 30, sequence: 0 },
      page: 1,
      loading: false,
      error: "",
      reload: vi.fn(),
      setPage: vi.fn(),
    });
    const html = renderToStaticMarkup(createElement(GLJournals, { route: "/gl/journals" }));
    expect(html).not.toContain('type="date"');
    expect(html).toContain("00/10/2569");
    expect(vi.mocked(glCommon.useGLList).mock.lastCall?.[1]).toBe("");
    expect(vi.mocked(glCommon.useGLList).mock.lastCall?.[2]).toBe("");
  });
});

describe("pasteIssueText", () => {
  const tr = (_key: string, fallback: string) => fallback;
  it("names the row, the column and how to fix an accounting negative", () => {
    const text = pasteIssueText({ row: 3, field: "debit", value: "(1,500.00)", problem: "negative" }, tr);
    expect(text).toContain("แถวที่ 3");
    expect(text).toContain("เดบิต");
    expect(text).toContain("(1,500.00)");
    expect(text).toContain("ฝั่งตรงข้าม");
  });
  it("names the column of a non-numeric cell", () => {
    expect(pasteIssueText({ row: 2, field: "credit", value: "หนึ่งพัน", problem: "invalid" }, tr)).toContain("เครดิต");
  });
});

describe("SME_PRESET_TEMPLATES", () => {
  it("has valid Thai SME presets with accounts and descriptions", async () => {
    const { SME_PRESET_TEMPLATES, presetToJournalTemplate } = await import("@/lib/gl-journal-templates");
    expect(SME_PRESET_TEMPLATES.length).toBeGreaterThanOrEqual(5);

    const rent = SME_PRESET_TEMPLATES.find((p) => p.id === "preset-rent");
    expect(rent).toBeDefined();
    expect(rent?.bookcode).toBe("PV");
    expect(rent?.lines.some((l) => l.accountcode === "53210")).toBe(true);

    const template = presetToJournalTemplate(rent!);
    expect(template.name).toBe(rent?.name);
    expect(template.template.bookcode).toBe("PV");
    expect(template.template.lines.length).toBe(rent?.lines.length);
  });
});


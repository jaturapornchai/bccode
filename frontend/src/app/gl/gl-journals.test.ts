import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { GLJournal } from "@/lib/general-ledger";
import * as glCommon from "./gl-common";
import { GLJournals } from "./gl-journals";

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
});

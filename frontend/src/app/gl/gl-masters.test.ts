import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { GLAccount, GLMaster } from "@/lib/general-ledger";
import * as glCommon from "./gl-common";
import { FormErrorAlert, GLMasters, TreeNodeRow, editorAlert, errorStatePatch, normalizeRecord, pageErrorText, paneErrorText, parentAccountCandidates, saveFailureTarget } from "./gl-masters";

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

describe("GLMasters CRUD table presentation", () => {
  it("renders table headers with code, name, status, and action column (จัดการ)", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "m-1",
            code: "P01",
            name: "งวด 1",
            isactive: true,
            version: 1,
          } as GLMaster,
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

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "periods", route: "/gl/periodlock" }));
    expect(html).toContain("รหัส");
    expect(html).toContain("ชื่อ / รายละเอียด");
    expect(html).toContain("สถานะ");
    expect(html).toContain("จัดการ");
    expect(html).toContain("P01");
    expect(html).toContain("งวด 1");
    expect(html).toContain("ใช้งาน");
    expect(html).toContain("title=\"แก้ไข (Edit)\"");
    expect(html).toContain("title=\"ลบ (Delete)\"");
  });

  it("renders amount column for budgets with formatted number and right alignment", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "b-1",
            code: "DEMO-BM69-BUD-410101",
            name: "งบกันยายน - รายได้ขายปูนซีเมนต์",
            amount: "100000.00",
            isactive: true,
            version: 1,
          } as unknown as GLMaster,
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

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "budgets", route: "/gl/budget" }));
    expect(html).toContain("จำนวนเงิน");
    expect(html).toContain("100,000.00");
    expect(html).toContain("DEMO-BM69-BUD-410101");
    expect(html).toContain("งบกันยายน - รายได้ขายปูนซีเมนต์");
  });

  it("renders account level badge when resource is accounts", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "a-1",
            accountcode: "110101",
            names: [{ code: "th", name: "เงินสดหน้าร้าน" }],
            level: 2,
            isactive: true,
            version: 1,
          } as unknown as GLAccount,
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

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "accounts", route: "/gl/chartofaccounts" }));
    expect(html).toContain("ระดับ");
    expect(html).toContain("ระดับ 2");
    expect(html).toContain("110101");
    expect(html).toContain("เงินสดหน้าร้าน");
  });

  it("renders status badge variants: locked and inactive", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "m-locked",
            code: "LOCKED-01",
            name: "งวดที่ถูกล็อก",
            locked: true,
            isactive: true,
            version: 1,
          } as unknown as GLMaster,
          {
            id: "m-inactive",
            code: "INACTIVE-01",
            name: "รายการปิดใช้งาน",
            isactive: false,
            version: 1,
          } as unknown as GLMaster,
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

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "periods", route: "/gl/periodlock" }));
    expect(html).toContain("ล็อกแล้ว");
    expect(html).toContain("ปิดใช้งาน");
  });

  it("renders empty workbench placeholder prompting row selection and no save button initially", () => {
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

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "accounts", route: "/gl/chartofaccounts" }));
    expect(html).toContain("เลือกรายการเพื่อแสดงข้อมูล");
    expect(html).toContain("คลิกที่แถวในตารางเพื่อแสดงข้อมูล");
    expect(html).not.toContain("บันทึกข้อมูล");
  });
});

describe("ผังบัญชี: failed save shows one Thai alert in the editor pane", () => {
  const duplicateThai = "รหัสบัญชีถูกใช้แล้ว กรุณาใช้รหัสอื่น";

  it("renders exactly one role=alert with the Thai message the server sent", () => {
    const html = renderToStaticMarkup(createElement(FormErrorAlert, { text: duplicateThai }));
    expect((html.match(/role="alert"/g) ?? []).length).toBe(1);
    expect(html).toContain(duplicateThai);
    expect(html).toContain("text-[0.95rem]");
  });

  it("renders no alert at all while there is no error", () => {
    expect(renderToStaticMarkup(createElement(FormErrorAlert, { text: "" }))).toBe("");
  });

  it("never shows the page notice and the pane notice together", () => {
    for (const error of ["", duplicateThai]) {
      for (const hasRecord of [true, false]) {
        const visible = [pageErrorText(error, hasRecord), paneErrorText(error, hasRecord)].filter(Boolean);
        expect(visible.length).toBe(error ? 1 : 0);
      }
    }
  });

  it("shows exactly ONE alert when a failed save also has list/reference errors (the 2-alert regression)", () => {
    const listError = "โหลดรายการไม่สำเร็จ กรุณากดโหลดใหม่";
    const refsError = "โหลดข้อมูลอ้างอิงไม่สำเร็จ กรุณากดโหลดใหม่";
    for (const hasRecord of [true, false]) {
      for (const list of ["", listError]) {
        for (const refs of ["", refsError]) {
          const result = editorAlert({ error: duplicateThai, listError: list, refsError: refs, hasRecord });
          expect(result.count, `alerts=${result.count} hasRecord=${hasRecord}`).toBe(1);
          expect(result.pane || result.page).toContain(duplicateThai);
          if (hasRecord) {
            expect(result.pane).toBe(duplicateThai);
            expect(result.page).toBe("");
          }
        }
      }
    }
  });

  it("keeps the page notice alive for a list failure while no record is open", () => {
    const listError = "โหลดรายการไม่สำเร็จ กรุณากดโหลดใหม่";
    const result = editorAlert({ error: "", listError, refsError: "", hasRecord: false });
    expect(result).toEqual({ pane: "", page: listError, count: 1 });
    expect(editorAlert({ error: "", hasRecord: false }).count).toBe(0);
  });

  it("touches only the error state, so the values the user typed survive a failure", () => {
    const patch = errorStatePatch({ message: duplicateThai, field: "" }, "duplicate_code");
    expect(Object.keys(patch).sort()).toEqual(["error", "errorField"]);
    expect(patch).toEqual({ error: duplicateThai, errorField: "accountcode" });
  });


  it("a rejected save always has a field to focus, even when the API sends none", () => {
    expect(saveFailureTarget("accounts")).toBe("accountcode");
    expect(saveFailureTarget("mappings")).toBe("code");
    expect(errorStatePatch({ message: "ทำรายการไม่สำเร็จ กรุณาลองใหม่ หากยังไม่ได้ให้ติดต่อผู้ดูแลระบบ", field: "" }, "unavailable", saveFailureTarget("accounts"))).toEqual({ error: "ทำรายการไม่สำเร็จ กรุณาลองใหม่ หากยังไม่ได้ให้ติดต่อผู้ดูแลระบบ", errorField: "accountcode" });
    expect(errorStatePatch({ message: "ซ้ำ", field: "" }, "duplicate_code", "accountcode").errorField).toBe("accountcode");
  });
  it("falls back to a Thai sentence (never provider/English text) when the API sends none", () => {
    const patch = errorStatePatch({ message: "ทำรายการไม่สำเร็จ กรุณาลองใหม่ หากยังไม่ได้ให้ติดต่อผู้ดูแลระบบ", field: "" }, "unavailable");
    expect(/[ก-๛]/.test(patch.error)).toBe(true);
    expect(patch.error).not.toContain("E11000");
  });

  it("UnsavedBadge renders null when dirty is false to avoid layout space", () => {
    const html = renderToStaticMarkup(createElement(glCommon.UnsavedBadge, { dirty: false }));
    expect(html).toBe("");
  });

  it("UnsavedBadge renders status badge with Thai indicator when dirty is true", () => {
    const html = renderToStaticMarkup(createElement(glCommon.UnsavedBadge, { dirty: true }));
    expect(html).toContain("ยังไม่บันทึก");
    expect(html).toContain("role=\"status\"");
  });

  it("normalizeRecord provides safe fallbacks when account record has null or missing names", () => {
    const raw = {
      id: "a-null-names",
      accountcode: "110101",
      names: null as unknown as { code: string; name: string }[],
      version: 1,
    } as unknown as GLAccount;

    const normalized = normalizeRecord("accounts", raw) as GLAccount;
    expect(Array.isArray(normalized.names)).toBe(true);
    expect(normalized.names.length).toBeGreaterThan(0);
    expect(normalized.names[0].code).toBe("th");
    expect(normalized.accountcode).toBe("110101");
    expect(normalized.accounttype).toBe("asset");
    expect(normalized.normalbalance).toBe("debit");
    expect(normalized.level).toBe(1);
    expect(normalized.isactive).toBe(true);
    expect(normalized.allowposting).toBe(true);
    expect(normalized.iscash).toBe(false);
  });

  it("normalizeRecord guarantees Thai name exists in names array", () => {
    const raw = {
      id: "a-en-only",
      accountcode: "110102",
      names: [{ code: "en", name: "Cash on hand" }],
      version: 1,
    } as unknown as GLAccount;

    const normalized = normalizeRecord("accounts", raw) as GLAccount;
    expect(normalized.names.some((n) => n.code === "th")).toBe(true);
    expect(normalized.names.find((n) => n.code === "en")?.name).toBe("Cash on hand");
  });

  it("TreeNodeRow renders control account badge for header accounts and posting badge for sub accounts", () => {
    const controlNode = {
      account: {
        id: "acc-ctrl",
        accountcode: "1100-00",
        names: [{ code: "th", name: "สินทรัพย์หมุนเวียน" }],
        accounttype: "asset",
        allowposting: false,
        isactive: true,
        level: 1,
      } as GLAccount,
      children: [],
      level: 1,
      hasChildren: true,
      category: "asset" as const,
    };

    const postingNode = {
      account: {
        id: "acc-post",
        accountcode: "1111-01",
        names: [{ code: "th", name: "เงินสดในมือ" }],
        accounttype: "asset",
        allowposting: true,
        isactive: true,
        level: 2,
      } as GLAccount,
      children: [],
      level: 2,
      hasChildren: false,
      category: "asset" as const,
    };

    const ctrlHtml = renderToStaticMarkup(createElement(TreeNodeRow, {
      node: controlNode,
      depth: 0,
      isEditing: false,
      expandedNodes: {},
      onToggleNode: () => {},
      onSelect: () => {},
      onEdit: () => {},
      onDelete: () => {},
      tr: (key: string, fallback: string) => fallback,
    }));

    const postHtml = renderToStaticMarkup(createElement(TreeNodeRow, {
      node: postingNode,
      depth: 1,
      isEditing: false,
      expandedNodes: {},
      onToggleNode: () => {},
      onSelect: () => {},
      onEdit: () => {},
      onDelete: () => {},
      tr: (key: string, fallback: string) => fallback,
    }));

    expect(ctrlHtml).toContain("1100-00");
    expect(ctrlHtml).toContain("สินทรัพย์หมุนเวียน");
    expect(ctrlHtml).toContain("บัญชีคุม");
    expect(ctrlHtml).toContain("L1");

    expect(postHtml).toContain("1111-01");
    expect(postHtml).toContain("เงินสดในมือ");
    expect(postHtml).toContain("บัญชีย่อย");
    expect(postHtml).toContain("L2");
    expect(ctrlHtml).toContain("shrink-0");
    expect(postHtml).toContain("shrink-0");
  });

  it("renders tree view with shrink-0 and overflow-y-auto to prevent vertical card overlap", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "acc-1",
            accountcode: "110101",
            names: [{ code: "th", name: "เงินสดหน้าร้าน" }],
            accounttype: "asset",
            allowposting: true,
            isactive: true,
            version: 1,
          } as unknown as GLAccount,
        ],
        total: 1,
        page: 1,
        limit: 1000,
        sequence: 0,
      },
      page: 1,
      loading: false,
      error: "",
      reload: vi.fn(),
      setPage: vi.fn(),
    });

    vi.mocked(glCommon.useReferences).mockReturnValue({
      accounts: [
        {
          id: "acc-1",
          accountcode: "110101",
          names: [{ code: "th", name: "เงินสดหน้าร้าน" }],
          accounttype: "asset",
          allowposting: true,
          isactive: true,
          version: 1,
        } as GLAccount,
      ],
      years: [],
      books: [],
      error: "",
      reload: vi.fn(),
    });

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "accounts", route: "/gl/accounts" }));
    // Initially list view button exists with title 'มุมมองผังต้นไม้'
    expect(html).toContain("title=\"มุมมองผังต้นไม้\"");
  });

  it("renders standard CRUD table styling with bc-list-toolbar, bc-list-header, and bc-list-row", () => {
    vi.mocked(glCommon.useGLList).mockReturnValue({
      data: {
        items: [
          {
            id: "p-1",
            code: "P01",
            name: "งวดบัญชี 1",
            isactive: true,
            version: 1,
          } as GLMaster,
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

    const html = renderToStaticMarkup(createElement(GLMasters, { resource: "periods", route: "/gl/periodlock" }));
    expect(html).toContain("bc-list-toolbar");
    expect(html).toContain("bc-list-header");
    expect(html).toContain("bc-list-row");
    expect(html).toContain("P01");
    expect(html).toContain("งวดบัญชี 1");
    expect(html).toContain("รายการทั้งหมด");
  });

  it("enforces multi-level Thai accounting standard for chart of accounts seed in deploy_fresh_database.sql", async () => {
    const { readFileSync } = await import("node:fs");
    const { resolve } = await import("node:path");
    const sqlPath = resolve(process.cwd(), "..", "deploy_fresh_database.sql");
    const sql = readFileSync(sqlPath, "utf8");

    // Must have Level 1 to Level 4 accounts
    expect(sql).toContain('"level": 1');
    expect(sql).toContain('"level": 2');
    expect(sql).toContain('"level": 3');
    expect(sql).toContain('"level": 4');

    // Must have parentaccountcode linkage
    expect(sql).toContain('"parentaccountcode"');

    // Must not have demo words in account names
    expect(sql).not.toContain("ข้อมูลตัวอย่าง - สินทรัพย์");
    expect(sql).not.toContain("ข้อมูลตัวอย่าง - หนี้สิน");
  });
});

describe("ผังบัญชี: parent-account picker", () => {
  const account = (accountcode: string, parentaccountcode: string | null, patch: Partial<GLAccount> = {}): GLAccount => ({
    accountcode, names: [{ code: "th", name: accountcode }], accounttype: "asset", parentaccountcode, normalbalance: "debit",
    allowposting: false, isactive: true, accountgroup: "", iscash: false, ...patch,
  });
  const chart = [
    account("10000", null),
    account("11000", "10000"),
    account("11100", "11000"),
    account("11110", "11100", { allowposting: true }),
    account("12000", "10000", { isactive: false }),
    account("13000", "10000", { isdeleted: true }),
    account("20000", null),
    account("21000", "20000"),
  ];
  const codes = (list: GLAccount[]) => list.map((row) => row.accountcode);

  it("offers only active, non-posting, not-deleted accounts", () => {
    expect(codes(parentAccountCandidates(chart, ""))).toEqual(["10000", "11000", "11100", "20000", "21000"]);
  });

  it("excludes the account itself and every descendant so the chart can never loop", () => {
    expect(codes(parentAccountCandidates(chart, "11000"))).toEqual(["10000", "20000", "21000"]);
    expect(codes(parentAccountCandidates(chart, "10000"))).toEqual(["20000", "21000"]);
  });
});

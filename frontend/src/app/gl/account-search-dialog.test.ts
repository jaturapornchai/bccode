import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import type { GLAccount } from "@/lib/general-ledger";
import { AccountSearchDialog } from "./account-search-dialog";
import { AccountSelect } from "./gl-common";

const mockAccounts: GLAccount[] = [
  {
    id: "acc-1",
    accountcode: "1100",
    names: [{ code: "th", name: "สินทรัพย์หมุนเวียน" }],
    accounttype: "asset",
    parentaccountcode: "",
    normalbalance: "debit",
    allowposting: false,
    isactive: true,
    accountgroup: "10",
    iscash: false,
    level: 1,
  },
  {
    id: "acc-2",
    accountcode: "110101",
    names: [{ code: "th", name: "เงินสดหน้าร้าน" }],
    accounttype: "asset",
    parentaccountcode: "1100",
    normalbalance: "debit",
    allowposting: true,
    isactive: true,
    accountgroup: "10",
    iscash: true,
    level: 2,
  },
  {
    id: "acc-3",
    accountcode: "210101",
    names: [{ code: "th", name: "เจ้าหนี้การค้า" }],
    accounttype: "liability",
    parentaccountcode: "",
    normalbalance: "credit",
    allowposting: true,
    isactive: true,
    accountgroup: "20",
    iscash: false,
    level: 1,
  },
  {
    id: "acc-4",
    accountcode: "310101",
    names: [{ code: "th", name: "ทุนจดทะเบียน" }],
    accounttype: "equity",
    parentaccountcode: "",
    normalbalance: "credit",
    allowposting: true,
    isactive: true,
    accountgroup: "30",
    iscash: false,
    level: 1,
  },
  {
    id: "acc-5",
    accountcode: "410101",
    names: [{ code: "th", name: "รายได้จากการขาย" }],
    accounttype: "income",
    parentaccountcode: "",
    normalbalance: "credit",
    allowposting: true,
    isactive: true,
    accountgroup: "40",
    iscash: false,
    level: 1,
  },
  {
    id: "acc-6",
    accountcode: "510101",
    names: [{ code: "th", name: "ต้นทุนขายสินค้า" }],
    accounttype: "expense",
    parentaccountcode: "",
    normalbalance: "debit",
    allowposting: true,
    isactive: true,
    accountgroup: "50",
    iscash: false,
    level: 1,
  },
];

describe("AccountSearchDialog", () => {
  it("renders when open and displays title and accounts", () => {
    const html = renderToStaticMarkup(
      createElement(AccountSearchDialog, {
        open: true,
        onClose: vi.fn(),
        onSelect: vi.fn(),
        accounts: mockAccounts,
      })
    );

    expect(html).toContain("ค้นหาและเลือกผังบัญชี");
    expect(html).toContain("110101");
    expect(html).toContain("เงินสดหน้าร้าน");
    expect(html).toContain("210101");
    expect(html).toContain("เจ้าหนี้การค้า");
    expect(html).toContain("410101");
    expect(html).toContain("รายได้จากการขาย");
  });

  it("does not render when open=false", () => {
    const html = renderToStaticMarkup(
      createElement(AccountSearchDialog, {
        open: false,
        onClose: vi.fn(),
        onSelect: vi.fn(),
        accounts: mockAccounts,
      })
    );

    expect(html).toBe("");
  });

  it("displays 5 standard accounting category pills", () => {
    const html = renderToStaticMarkup(
      createElement(AccountSearchDialog, {
        open: true,
        onClose: vi.fn(),
        onSelect: vi.fn(),
        accounts: mockAccounts,
      })
    );

    expect(html).toContain("1. สินทรัพย์");
    expect(html).toContain("2. หนี้สิน");
    expect(html).toContain("3. ส่วนของเจ้าของ");
    expect(html).toContain("4. รายได้");
    expect(html).toContain("5. ค่าใช้จ่าย");
  });

  it("displays level indentation tree for nested accounts (level > 1)", () => {
    const html = renderToStaticMarkup(
      createElement(AccountSearchDialog, {
        open: true,
        onClose: vi.fn(),
        onSelect: vi.fn(),
        accounts: mockAccounts,
        all: true,
      })
    );

    // Level 2 should display branch symbol
    expect(html).toContain("└─");
    expect(html).toContain("ระดับ 2");
    expect(html).toContain("ระดับ 1");
  });

  it("supports multi-select mode with checkboxes", () => {
    const html = renderToStaticMarkup(
      createElement(AccountSearchDialog, {
        open: true,
        onClose: vi.fn(),
        onSelectMultiple: vi.fn(),
        accounts: mockAccounts,
        multiSelect: true,
        selectedCodes: ["110101", "210101"],
        all: true,
      })
    );

    expect(html).toContain('type="checkbox"');
    expect(html).toContain("ตกลงเลือก");
    expect(html).toContain("เลือกอยู่ 2 บัญชี");
  });

  it("sorts accounts hierarchically and renders control account badge when !all", () => {
    // Child first in input array to verify hierarchical sort reorders correctly
    const outOfOrderAccounts: GLAccount[] = [
      mockAccounts[1], // 110101 (child)
      mockAccounts[0], // 1100 (parent, control account)
    ];

    const html = renderToStaticMarkup(
      createElement(AccountSearchDialog, {
        open: true,
        onClose: vi.fn(),
        onSelect: vi.fn(),
        accounts: outOfOrderAccounts,
        all: false,
      })
    );

    // Parent account 1100 should appear before child 110101 in table
    const idx1100 = html.indexOf("1100");
    const idx110101 = html.indexOf("110101");
    expect(idx1100).toBeGreaterThan(-1);
    expect(idx110101).toBeGreaterThan(-1);
    expect(idx1100).toBeLessThan(idx110101);

    // Control account 1100 should show บัญชีคุม badge
    expect(html).toContain("บัญชีคุม");
    // Postable account 110101 should show เลือก button
    expect(html).toContain("เลือก");
    // Search input should have clearance padding and z-10 icon
    expect(html).toContain("!pl-11 !pr-10");
    expect(html).toContain("lucide-search");
  });
});

describe("AccountSelect in gl-common", () => {
  it("renders accessible select for Playwright and search trigger button", () => {
    const onChange = vi.fn();
    const html = renderToStaticMarkup(
      createElement(AccountSelect, {
        label: "บัญชีแม่",
        value: "110101",
        onChange,
        accounts: mockAccounts,
      })
    );

    // Accessible select with aria-label for Playwright
    expect(html).toContain('aria-label="บัญชีแม่"');
    expect(html).toContain("110101");
    expect(html).toContain("เงินสดหน้าร้าน");

    // Search button trigger
    expect(html).toContain("เปิดระบบค้นหาผังบัญชีแบบเต็มจอ");
    expect(html).toContain("ล้างค่าที่เลือก");
  });
});

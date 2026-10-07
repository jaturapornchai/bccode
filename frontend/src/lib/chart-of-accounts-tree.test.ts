import { describe, it, expect } from "vitest";
import {
  resolveAccountCategory,
  buildChartOfAccountsTree,
  filterAccountTree,
} from "./chart-of-accounts-tree";
import type { GLAccount } from "./general-ledger";

describe("chart-of-accounts-tree", () => {
  const sampleAccounts: GLAccount[] = [
    {
      accountcode: "1000",
      names: [{ code: "th", name: "สินทรัพย์" }],
      accounttype: "asset",
      parentaccountcode: "",
      normalbalance: "debit",
      allowposting: false,
      isactive: true,
      iscash: false,
      level: 1,
    },
    {
      accountcode: "1100",
      names: [{ code: "th", name: "สินทรัพย์หมุนเวียน" }],
      accounttype: "asset",
      parentaccountcode: "1000",
      normalbalance: "debit",
      allowposting: false,
      isactive: true,
      iscash: false,
      level: 2,
    },
    {
      accountcode: "1111",
      names: [{ code: "th", name: "เงินสดในมือ" }],
      accounttype: "asset",
      parentaccountcode: "1100",
      normalbalance: "debit",
      allowposting: true,
      isactive: true,
      iscash: true,
      level: 3,
    },
    {
      accountcode: "2000",
      names: [{ code: "th", name: "หนี้สิน" }],
      accounttype: "liability",
      parentaccountcode: "",
      normalbalance: "credit",
      allowposting: false,
      isactive: true,
      iscash: false,
      level: 1,
    },
    {
      accountcode: "2111",
      names: [{ code: "th", name: "เจ้าหนี้การค้า" }],
      accounttype: "liability",
      parentaccountcode: "2000",
      normalbalance: "credit",
      allowposting: true,
      isactive: true,
      iscash: false,
      level: 2,
    },
  ];

  it("should correctly resolve account categories from accounttype or code prefix", () => {
    expect(resolveAccountCategory(sampleAccounts[0])).toBe("asset");
    expect(resolveAccountCategory(sampleAccounts[3])).toBe("liability");

    const fallbackAccount: GLAccount = {
      accountcode: "4111",
      names: [{ code: "th", name: "รายได้จากการขาย" }],
      accounttype: "income",
      parentaccountcode: "",
      normalbalance: "credit",
      allowposting: true,
      isactive: true,
      iscash: false,
    };
    expect(resolveAccountCategory(fallbackAccount)).toBe("income");
  });

  it("should build hierarchical tree groups for all 5 categories", () => {
    const groups = buildChartOfAccountsTree(sampleAccounts);

    expect(groups.length).toBe(5);
    const assetGroup = groups.find((g) => g.category === "asset")!;
    expect(assetGroup.totalAccounts).toBe(3);
    expect(assetGroup.rootNodes.length).toBe(1); // 1000

    const root1000 = assetGroup.rootNodes[0];
    expect(root1000.account.accountcode).toBe("1000");
    expect(root1000.children.length).toBe(1); // 1100

    const child1100 = root1000.children[0];
    expect(child1100.account.accountcode).toBe("1100");
    expect(child1100.children.length).toBe(1); // 1111

    const child1111 = child1100.children[0];
    expect(child1111.account.accountcode).toBe("1111");
    expect(child1111.hasChildren).toBe(false);
  });

  it("should filter the tree by search query and preserve parent path", () => {
    const groups = buildChartOfAccountsTree(sampleAccounts);
    // ค้นหา "เงินสด" ซึ่งอยู่ที่ 1111
    const filtered = filterAccountTree(groups, "เงินสด");

    const assetGroup = filtered.find((g) => g.category === "asset")!;
    expect(assetGroup.totalAccounts).toBe(3); // 1000 -> 1100 -> 1111
    expect(assetGroup.rootNodes.length).toBe(1);

    const liabilityGroup = filtered.find((g) => g.category === "liability")!;
    expect(liabilityGroup.totalAccounts).toBe(0);
  });

  it("should automatically nest accounts by level when parentaccountcode is empty (Champ standard)", () => {
    const champAccounts: GLAccount[] = [
      {
        accountcode: "1000",
        names: [{ code: "th", name: "สินทรัพย์" }],
        accounttype: "asset",
        parentaccountcode: "",
        normalbalance: "debit",
        allowposting: false,
        isactive: true,
        iscash: false,
        level: 1,
      },
      {
        accountcode: "1100",
        names: [{ code: "th", name: "สินทรัพย์หมุนเวียน" }],
        accounttype: "asset",
        parentaccountcode: "",
        normalbalance: "debit",
        allowposting: false,
        isactive: true,
        iscash: false,
        level: 2,
      },
      {
        accountcode: "1111",
        names: [{ code: "th", name: "เงินสดในมือ" }],
        accounttype: "asset",
        parentaccountcode: "",
        normalbalance: "debit",
        allowposting: true,
        isactive: true,
        iscash: false,
        level: 3,
      },
      {
        accountcode: "1112",
        names: [{ code: "th", name: "เงินฝากธนาคาร" }],
        accounttype: "asset",
        parentaccountcode: "",
        normalbalance: "debit",
        allowposting: true,
        isactive: true,
        iscash: false,
        level: 3,
      },
    ];

    const groups = buildChartOfAccountsTree(champAccounts);
    const assetGroup = groups.find((g) => g.category === "asset")!;
    expect(assetGroup.rootNodes.length).toBe(1); // 1000 is the root

    const root1000 = assetGroup.rootNodes[0];
    expect(root1000.account.accountcode).toBe("1000");
    expect(root1000.children.length).toBe(1); // 1100

    const child1100 = root1000.children[0];
    expect(child1100.account.accountcode).toBe("1100");
    expect(child1100.children.length).toBe(2); // 1111 and 1112
    expect(child1100.children[0].account.accountcode).toBe("1111");
    expect(child1100.children[1].account.accountcode).toBe("1112");
  });

  it("should validate the 5-level Thai standard chart of accounts fixture (DBD 2566) with >= 100 accounts", () => {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const fs = require("node:fs");
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const path = require("node:path");

    const jsonPath = path.resolve(__dirname, "../../../backend/internal/generalledger/standard_chart_th.json");
    const raw = fs.readFileSync(jsonPath, "utf-8");
    const accounts = JSON.parse(raw);

    expect(accounts.length).toBeGreaterThanOrEqual(100);
    const level1 = accounts.filter((a: GLAccount) => a.level === 1);
    const level2 = accounts.filter((a: GLAccount) => a.level === 2);
    const level3 = accounts.filter((a: GLAccount) => a.level === 3);
    const level4 = accounts.filter((a: GLAccount) => a.level === 4);
    const level5 = accounts.filter((a: GLAccount) => a.level === 5);

    expect(level1.length).toBe(5);
    expect(level2.length).toBe(13);
    expect(level3.length).toBe(39);
    expect(level4.length).toBe(94);
    expect(level5.length).toBe(47);

    // Test tree building with full standard accounts
    const groups = buildChartOfAccountsTree(accounts);
    expect(groups.length).toBe(5);

    const totalTreeAccounts = groups.reduce((sum: number, g: { totalAccounts: number }) => sum + g.totalAccounts, 0);
    expect(totalTreeAccounts).toBe(accounts.length);

    // Verify specific rules from mydocs/specs/chatofaccount.md
    // 1. Sort by accountcode
    const codes = accounts.map((a: GLAccount) => a.accountcode);
    const sortedCodes = [...codes].sort((a: string, b: string) => a.localeCompare(b, "en", { numeric: true }));
    expect(codes).toEqual(sortedCodes);

    // 2. All level 1 accounts have parentaccountcode === null
    const level1Accounts = accounts.filter((a: GLAccount) => a.level === 1);
    for (const a of level1Accounts) {
      expect(a.parentaccountcode).toBeNull();
      expect(a.allowposting).toBe(false);
    }

    // 3. Parent accounts (having children) have allowposting === false
    const parentCodesSet = new Set(accounts.map((a: GLAccount) => a.parentaccountcode).filter(Boolean));
    for (const a of accounts) {
      if (parentCodesSet.has(a.accountcode)) {
        expect(a.allowposting).toBe(false);
      } else {
        expect(a.allowposting).toBe(true);
      }
    }
  });
});


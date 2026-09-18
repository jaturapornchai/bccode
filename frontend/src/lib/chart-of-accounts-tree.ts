// Chart of Accounts Hierarchical Tree Engine
// พัฒนาตามมาตรฐานการรายงานทางการเงินไทย TFRS for NPAEs / PAEs
// จัดโครงสร้างผังบัญชี 5 หมวดบัญชีมาตรฐาน พร้อมระบบคำนวณและแสดงผลแบบต้นไม้

import type { GLAccount } from "./general-ledger";

export type AccountCategory = "asset" | "liability" | "equity" | "income" | "expense";

export interface AccountTreeNode {
  account: GLAccount;
  children: AccountTreeNode[];
  level: number;
  hasChildren: boolean;
  category: AccountCategory;
}

export interface CategoryTreeGroup {
  category: AccountCategory;
  categoryNumber: number;
  nameTh: string;
  nameEn: string;
  colorClass: {
    badge: string;
    text: string;
    bg: string;
    border: string;
  };
  rootNodes: AccountTreeNode[];
  totalAccounts: number;
}

export const CATEGORY_CONFIGS: Record<
  AccountCategory,
  {
    number: number;
    nameTh: string;
    nameEn: string;
    colorClass: { badge: string; text: string; bg: string; border: string };
  }
> = {
  asset: {
    number: 1,
    nameTh: "สินทรัพย์",
    nameEn: "Assets",
    colorClass: {
      badge: "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/20",
      text: "text-blue-600 dark:text-blue-400",
      bg: "bg-blue-50/50 dark:bg-blue-950/20",
      border: "border-blue-500/30",
    },
  },
  liability: {
    number: 2,
    nameTh: "หนี้สิน",
    nameEn: "Liabilities",
    colorClass: {
      badge: "bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-500/20",
      text: "text-rose-600 dark:text-rose-400",
      bg: "bg-rose-50/50 dark:bg-rose-950/20",
      border: "border-rose-500/30",
    },
  },
  equity: {
    number: 3,
    nameTh: "ส่วนของเจ้าของ",
    nameEn: "Equity",
    colorClass: {
      badge: "bg-purple-500/10 text-purple-700 dark:text-purple-300 border-purple-500/20",
      text: "text-purple-600 dark:text-purple-400",
      bg: "bg-purple-50/50 dark:bg-purple-950/20",
      border: "border-purple-500/30",
    },
  },
  income: {
    number: 4,
    nameTh: "รายได้",
    nameEn: "Income",
    colorClass: {
      badge: "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/20",
      text: "text-emerald-600 dark:text-emerald-400",
      bg: "bg-emerald-50/50 dark:bg-emerald-950/20",
      border: "border-emerald-500/30",
    },
  },
  expense: {
    number: 5,
    nameTh: "ค่าใช้จ่าย",
    nameEn: "Expenses",
    colorClass: {
      badge: "bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/20",
      text: "text-amber-600 dark:text-amber-400",
      bg: "bg-amber-50/50 dark:bg-amber-950/20",
      border: "border-amber-500/30",
    },
  },
};

/**
 * ระบุหมวดบัญชีจากฟิลด์ accounttype หรือจากหลักแรกของรหัสบัญชี
 */
export function resolveAccountCategory(account: GLAccount): AccountCategory {
  if (account.accounttype && CATEGORY_CONFIGS[account.accounttype]) {
    return account.accounttype;
  }

  const firstChar = (account.accountcode || "").trim().charAt(0);
  switch (firstChar) {
    case "1":
      return "asset";
    case "2":
      return "liability";
    case "3":
      return "equity";
    case "4":
      return "income";
    case "5":
      return "expense";
    default:
      return "asset";
  }
}

/**
 * สร้างโครงสร้าง Tree Node จาก GLAccount
 */
function createTreeNode(
  account: GLAccount,
  category: AccountCategory,
  level = 1,
): AccountTreeNode {
  return {
    account,
    children: [],
    level: typeof account.level === "number" && account.level >= 1 ? account.level : level,
    hasChildren: false,
    category,
  };
}

/**
 * สร้างโครงสร้าง Tree สำหรับผังบัญชีทั้ง 5 หมวดหมู่
 */
export function buildChartOfAccountsTree(accounts: GLAccount[]): CategoryTreeGroup[] {
  const categories: AccountCategory[] = ["asset", "liability", "equity", "income", "expense"];

  // 1. แยกบัญชีตามหมวด
  const byCategory: Record<AccountCategory, GLAccount[]> = {
    asset: [],
    liability: [],
    equity: [],
    income: [],
    expense: [],
  };

  for (const acc of accounts) {
    const cat = resolveAccountCategory(acc);
    byCategory[cat].push(acc);
  }

  // 2. สร้าง Tree ในแต่ละหมวด
  return categories.map((cat) => {
    const catAccounts = byCategory[cat];
    // เรียงตามรหัสบัญชี
    catAccounts.sort((a, b) => a.accountcode.localeCompare(b.accountcode, "en", { numeric: true }));

    const nodeMap = new Map<string, AccountTreeNode>();
    for (const acc of catAccounts) {
      nodeMap.set(acc.accountcode, createTreeNode(acc, cat));
    }

    const rootNodes: AccountTreeNode[] = [];
    const stack: AccountTreeNode[] = [];

    for (const acc of catAccounts) {
      const node = nodeMap.get(acc.accountcode)!;
      const parentCode = (acc.parentaccountcode || "").trim();

      if (parentCode && nodeMap.has(parentCode) && parentCode !== acc.accountcode) {
        const parentNode = nodeMap.get(parentCode)!;
        parentNode.children.push(node);
        parentNode.hasChildren = true;
        // ปรับ level ให้ลึกกว่า parent
        node.level = Math.max(node.level, parentNode.level + 1);
      } else {
        // Stack-based hierarchical nesting using level (1..12) as in Champ / Thai ERP standards
        const nodeLevel = node.level || 1;
        while (stack.length > 0 && stack[stack.length - 1].level >= nodeLevel) {
          stack.pop();
        }
        if (stack.length > 0) {
          const parentNode = stack[stack.length - 1];
          parentNode.children.push(node);
          parentNode.hasChildren = true;
        } else {
          rootNodes.push(node);
        }
        stack.push(node);
      }
    }

    const config = CATEGORY_CONFIGS[cat];

    return {
      category: cat,
      categoryNumber: config.number,
      nameTh: config.nameTh,
      nameEn: config.nameEn,
      colorClass: config.colorClass,
      rootNodes,
      totalAccounts: catAccounts.length,
    };
  });
}

/**
 * ค้นหาและกรองโหนดในต้นไม้ตามข้อความค้นหา (Search Query)
 */
export function filterAccountTree(
  groups: CategoryTreeGroup[],
  query: string,
): CategoryTreeGroup[] {
  const cleanQuery = query.trim().toLowerCase();
  if (!cleanQuery) return groups;

  function filterNode(node: AccountTreeNode): AccountTreeNode | null {
    const codeMatch = node.account.accountcode.toLowerCase().includes(cleanQuery);
    const thName = node.account.names?.find((n) => n.code === "th")?.name || "";
    const enName = node.account.names?.find((n) => n.code === "en")?.name || "";
    const nameMatch =
      thName.toLowerCase().includes(cleanQuery) || enName.toLowerCase().includes(cleanQuery);

    const filteredChildren = node.children
      .map(filterNode)
      .filter((n): n is AccountTreeNode => n !== null);

    if (codeMatch || nameMatch || filteredChildren.length > 0) {
      return {
        ...node,
        children: filteredChildren,
        hasChildren: filteredChildren.length > 0,
      };
    }
    return null;
  }

  return groups.map((g) => {
    const filteredRoots = g.rootNodes
      .map(filterNode)
      .filter((n): n is AccountTreeNode => n !== null);

    let count = 0;
    function countNodes(nodes: AccountTreeNode[]) {
      for (const n of nodes) {
        count++;
        countNodes(n.children);
      }
    }
    countNodes(filteredRoots);

    return {
      ...g,
      rootNodes: filteredRoots,
      totalAccounts: count,
    };
  });
}

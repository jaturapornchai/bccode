import type { GLLine } from "./general-ledger";

/** Posting batch (Champ BCPostTrans) as returned by GET post-batches: names are free text typed at posting time; postedby is the user id. */
export type GLPostBatchInfo = { batchno: string; postedat?: string; postedby?: string; count?: number; remark?: string; check1?: string; check2?: string; approver?: string };

// Exact digits only (no float): "0", "0.00" and "" are zero, anything with a digit 1-9 is an amount.
const hasAmount = (value: string | undefined) => /[1-9]/.test(value ?? "");

/** Champ footer "จำนวนรายการ" (SumDebitOrCredit): lines that carry an account or a debit/credit amount; untouched filler rows are not counted. */
export function countJournalEntries(lines: readonly GLLine[]): number {
  return lines.filter((line) => line.accountcode.trim() !== "" || hasAmount(line.debit) || hasAmount(line.credit)).length;
}

/** Champ Prev/Next record: the voucher before/after the open one in the loaded list order; null at either end or when the open one is not in the list. */
export function adjacentJournal<T extends { id?: string }>(items: readonly T[], currentId: string | undefined, step: -1 | 1): T | null {
  if (!currentId) return null;
  const index = items.findIndex((item) => item.id === currentId);
  return index < 0 ? null : items[index + step] ?? null;
}

/** One GET by batch number: the list search is a substring match on the code, so the exact batch is picked from the (short) result. */
export function postBatchLookupPath(batchno: string): string {
  return `post-batches?${new URLSearchParams({ q: batchno, page: "1", limit: "50" })}`;
}

export function pickPostBatch(items: readonly GLPostBatchInfo[] | undefined, batchno: string): GLPostBatchInfo | null {
  return items?.find((item) => item.batchno === batchno) ?? null;
}

/** Posting date for the screen as YYYY-MM-DD in Thai time (the stored value is a UTC timestamp); "" when missing or invalid. */
export function postBatchDate(postedat: string | undefined): string {
  const time = Date.parse(postedat ?? "");
  return Number.isNaN(time) ? "" : new Date(time).toLocaleDateString("en-CA", { timeZone: "Asia/Bangkok" });
}

/** Names for the voucher print signature lines: reviewers 1 and 2 share the "checked by" box. */
export function postBatchSignatories(batch: GLPostBatchInfo | null | undefined): { checkedBy: string; approvedBy: string } {
  return {
    checkedBy: [batch?.check1, batch?.check2].map((name) => name?.trim() ?? "").filter(Boolean).join(" / "),
    approvedBy: batch?.approver?.trim() ?? "",
  };
}

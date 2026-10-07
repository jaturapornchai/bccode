import { describe, expect, it } from "vitest";
import { emptyLine } from "./general-ledger";
import { adjacentJournal, countJournalEntries, pickPostBatch, postBatchDate, postBatchLookupPath, postBatchSignatories } from "./gl-journal-voucher";

describe("countJournalEntries", () => {
  it("counts lines with an account or a non-zero amount and skips filler rows", () => {
    const lines = [
      { ...emptyLine(), accountcode: "11110", debit: "100.00" },
      { ...emptyLine(), credit: "100.00" },
      { ...emptyLine(), accountcode: "41100" },
      { ...emptyLine(), debit: "0", credit: "0.00" },
      emptyLine(),
    ];
    expect(countJournalEntries(lines)).toBe(3);
    expect(countJournalEntries([])).toBe(0);
  });
});

describe("adjacentJournal", () => {
  const items = [{ id: "a" }, { id: "b" }, { id: "c" }];
  it("moves along the loaded list and stops at both ends", () => {
    expect(adjacentJournal(items, "b", -1)).toBe(items[0]);
    expect(adjacentJournal(items, "b", 1)).toBe(items[2]);
    expect(adjacentJournal(items, "a", -1)).toBeNull();
    expect(adjacentJournal(items, "c", 1)).toBeNull();
  });
  it("has no neighbour for a new voucher or one outside the list", () => {
    expect(adjacentJournal(items, undefined, 1)).toBeNull();
    expect(adjacentJournal(items, "zz", 1)).toBeNull();
  });
});

describe("posting batch lookup", () => {
  it("asks the list for the batch number and picks the exact match", () => {
    expect(postBatchLookupPath("PB 1")).toBe("post-batches?q=PB+1&page=1&limit=50");
    const items = [{ batchno: "10" }, { batchno: "1", approver: "คุณสมศรี" }];
    expect(pickPostBatch(items, "1")).toBe(items[1]);
    expect(pickPostBatch(items, "2")).toBeNull();
    expect(pickPostBatch(undefined, "1")).toBeNull();
  });
  it("shows the posting date in Thai time", () => {
    expect(postBatchDate("2026-09-30T18:30:00Z")).toBe("2026-10-01");
    expect(postBatchDate("2026-09-30T05:00:00.123456Z")).toBe("2026-09-30");
    expect(postBatchDate("")).toBe("");
    expect(postBatchDate(undefined)).toBe("");
  });
  it("joins the reviewers for the print signature box", () => {
    expect(postBatchSignatories({ batchno: "1", check1: " สมชาย ", check2: "สมหญิง", approver: "สมศักดิ์" })).toEqual({ checkedBy: "สมชาย / สมหญิง", approvedBy: "สมศักดิ์" });
    expect(postBatchSignatories({ batchno: "1", check2: "สมหญิง" })).toEqual({ checkedBy: "สมหญิง", approvedBy: "" });
    expect(postBatchSignatories(null)).toEqual({ checkedBy: "", approvedBy: "" });
  });
});

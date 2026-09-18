import { describe, it, expect } from "vitest";
import {
  parseThaiBankStatement,
  autoReconcileBankStatement,
  calculateBankReconciliation,
  generateBankFastVoucher,
  type BankStatementTransaction,
  type BookTransaction,
} from "./bank-reconciliation";

describe("Bank Feeds & Smart Reconciliation Engine (Group A)", () => {
  it("parses Thai bank CSV statement lines correctly", () => {
    const rawCsv = `
Date,Description,Withdrawal,Deposit,Balance
2026-09-01,Transfer from SCB K.Somchai,0.00,15000.00,115000.00
2026-09-02,PromptPay Payment to OfficeMate,2140.00,0.00,112860.00
2026-09-03,Bank Monthly Fee,50.00,0.00,112810.00
2026-09-04,Bank Interest Deposit,0.00,125.50,112935.50
    `.trim();

    const items = parseThaiBankStatement(rawCsv);
    expect(items).toHaveLength(4);

    expect(items[0]).toMatchObject({
      date: "2026-09-01",
      type: "deposit",
      amount: 15000,
      status: "unmatched",
    });

    expect(items[1]).toMatchObject({
      date: "2026-09-02",
      type: "withdrawal",
      amount: 2140,
      channel: "promptpay",
    });

    expect(items[2]).toMatchObject({
      date: "2026-09-03",
      type: "withdrawal",
      amount: 50,
      channel: "fee",
    });

    expect(items[3]).toMatchObject({
      date: "2026-09-04",
      type: "deposit",
      amount: 125.5,
      channel: "interest",
    });
  });

  it("handles Thai Buddhist era date format (DD/MM/2569)", () => {
    const raw = "15/09/2569,PromptPay in,0,5000\n16/09/2569,Fee,30,0";
    const items = parseThaiBankStatement(raw);
    expect(items).toHaveLength(2);
    expect(items[0].date).toBe("2026-09-15");
    expect(items[0].amount).toBe(5000);
    expect(items[1].date).toBe("2026-09-16");
    expect(items[1].amount).toBe(30);
  });

  it("auto reconciles statement lines with book transactions within date window", () => {
    const statements: BankStatementTransaction[] = [
      {
        id: "stmt-1",
        date: "2026-09-10",
        type: "deposit",
        amount: 25000,
        description: "Customer deposit INV2026-001",
        reference: "INV2026001",
        status: "unmatched",
      },
      {
        id: "stmt-2",
        date: "2026-09-11",
        type: "withdrawal",
        amount: 4500,
        description: "Supplier payment PO99",
        status: "unmatched",
      },
      {
        id: "stmt-3",
        date: "2026-09-12",
        type: "withdrawal",
        amount: 200,
        description: "Bank fee",
        channel: "fee",
        status: "unmatched",
      },
    ];

    const books: BookTransaction[] = [
      {
        docno: "RV2026-001",
        date: "2026-09-09", // within 1 day window
        type: "receipt",
        amount: 25000,
        accountCode: "111200",
        accountName: "ธนาคารกสิกรไทย",
        description: "รับชำระหนี้ INV2026-001",
        reference: "INV2026001",
        isReconciled: false,
      },
      {
        docno: "PV2026-088",
        date: "2026-09-11", // exact date match
        type: "payment",
        amount: 4500,
        accountCode: "111200",
        accountName: "ธนาคารกสิกรไทย",
        description: "จ่ายเจ้าหนี้ PO99",
        isReconciled: false,
      },
    ];

    const result = autoReconcileBankStatement(statements, books, { dateWindowDays: 3 });
    expect(result.matchCount).toBe(2);

    expect(result.matchedStatements[0].status).toBe("matched");
    expect(result.matchedStatements[0].matchedJournalDocNo).toBe("RV2026-001");
    expect(result.matchedStatements[0].matchConfidence).toBeGreaterThanOrEqual(0.9);

    expect(result.matchedStatements[1].status).toBe("matched");
    expect(result.matchedStatements[1].matchedJournalDocNo).toBe("PV2026-088");

    // Unmatched item remains unmatched
    expect(result.matchedStatements[2].status).toBe("unmatched");
  });

  it("calculates balanced Bank Reconciliation statement (TFRS)", () => {
    const statements: BankStatementTransaction[] = [
      {
        id: "stmt-1",
        date: "2026-09-01",
        type: "deposit",
        amount: 10000,
        description: "Customer payment",
        status: "matched",
      },
      {
        id: "stmt-2",
        date: "2026-09-02",
        type: "withdrawal",
        amount: 100,
        description: "Bank service fee",
        channel: "fee",
        status: "unmatched", // In bank statement, not yet in book
      },
    ];

    const books: BookTransaction[] = [
      {
        docno: "RV-1",
        date: "2026-09-01",
        type: "receipt",
        amount: 10000,
        accountCode: "111200",
        accountName: "Bank",
        description: "Matched",
        isReconciled: true,
      },
      {
        docno: "RV-2",
        date: "2026-09-30",
        type: "receipt",
        amount: 5000,
        accountCode: "111200",
        accountName: "Bank",
        description: "Deposit in transit (sent end of day)",
        isReconciled: false, // Deposit in transit
      },
      {
        docno: "PV-1",
        date: "2026-09-28",
        type: "payment",
        amount: 2000,
        accountCode: "111200",
        accountName: "Bank",
        description: "Outstanding cheque issued to supplier",
        isReconciled: false, // Outstanding cheque
      },
    ];

    // Bank ending balance: 50,000
    // Book ending balance: 50,000 + 5,000 (transit) - 2,000 (cheque) + 100 (unrecorded fee) = 53,100
    const summary = calculateBankReconciliation({
      statementEndingBalance: 50000,
      statementTransactions: statements,
      bookEndingBalance: 53100,
      bookTransactions: books,
    });

    expect(summary.depositsInTransit).toBe(5000);
    expect(summary.outstandingCheques).toBe(2000);
    expect(summary.adjustedBankBalance).toBe(53000); // 50000 + 5000 - 2000 = 53000

    expect(summary.unrecordedBankCharges).toBe(100);
    expect(summary.adjustedBookBalance).toBe(53000); // 53100 - 100 = 53000

    expect(summary.isBalanced).toBe(true);
    expect(summary.difference).toBe(0);
  });

  it("generates fast balanced clearing voucher for bank fees and interest", () => {
    const feeStmt: BankStatementTransaction = {
      id: "stmt-fee",
      date: "2026-09-15",
      type: "withdrawal",
      amount: 150,
      description: "SMS Alert Fee",
      channel: "fee",
      status: "unmatched",
    };

    const voucher = generateBankFastVoucher(feeStmt, "111200");
    expect(voucher.journalType).toBe("PV");
    expect(voucher.lines).toHaveLength(2);
    expect(voucher.lines[0].debit).toBe(150);
    expect(voucher.lines[0].credit).toBe(0);
    expect(voucher.lines[1].debit).toBe(0);
    expect(voucher.lines[1].credit).toBe(150);
  });
});

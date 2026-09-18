/**
 * Bank Feeds & Smart Reconciliation Engine for Thai Banking System (TFRS)
 * Supports parsing KBank, SCB, BBL, KTB, TTB, BAY statements and 2-way matching with GL book transactions.
 */

export type BankTransactionType = "deposit" | "withdrawal";

export type BankChannel =
  | "promptpay"
  | "transfer"
  | "edc"
  | "cheque"
  | "atm"
  | "fee"
  | "interest"
  | "other";

export type BankStatementTransaction = {
  id: string;
  date: string; // YYYY-MM-DD
  time?: string; // HH:mm:ss
  type: BankTransactionType;
  amount: number; // positive number in satang-safe floating / 2-decimal
  description: string;
  reference?: string;
  channel?: BankChannel;
  balance?: number;
  status: "unmatched" | "matched" | "cleared" | "ignored";
  matchedJournalDocNo?: string;
  matchConfidence?: number;
};

export type BookTransaction = {
  docno: string;
  date: string; // YYYY-MM-DD
  type: "receipt" | "payment" | "journal";
  amount: number;
  accountCode: string;
  accountName: string;
  description: string;
  reference?: string;
  isReconciled: boolean;
  reconciledWithId?: string;
};

export type BankReconciliationSummary = {
  statementEndingBalance: number;
  depositsInTransit: number;
  depositsInTransitCount: number;
  outstandingCheques: number;
  outstandingChequesCount: number;
  adjustedBankBalance: number;
  bookEndingBalance: number;
  unrecordedBankCharges: number;
  unrecordedBankInterest: number;
  unrecordedTransfersIn: number;
  adjustedBookBalance: number;
  difference: number;
  isBalanced: boolean;
};

/**
 * Parses raw bank statement text (CSV, TSV, or copy-pasted table) into structured transactions.
 */
export function parseThaiBankStatement(
  rawText: string,
  _bankCode?: string,
): BankStatementTransaction[] {
  if (!rawText || !rawText.trim()) return [];

  const lines = rawText
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  const results: BankStatementTransaction[] = [];
  let seq = 1;

  for (const line of lines) {
    // Skip headers
    const lower = line.toLowerCase();
    if (
      lower.includes("date") ||
      lower.includes("วันที่") ||
      lower.includes("transaction") ||
      lower.includes("ถอน") ||
      lower.includes("ฝาก") ||
      lower.includes("balance")
    ) {
      if (
        lower.startsWith("date") ||
        lower.startsWith("วันที่") ||
        lower.startsWith("trans")
      ) {
        continue;
      }
    }

    // Split by comma, semicolon, or tab
    const delimiter = line.includes("\t")
      ? "\t"
      : line.includes(";")
        ? ";"
        : ",";
    const parts = line.split(delimiter).map((p) => p.replace(/^["']|["']$/g, "").trim());

    if (parts.length < 3) continue;

    // Detect date pattern (e.g. YYYY-MM-DD, DD/MM/YYYY, DD-MM-YYYY)
    let parsedDate = "";
    let rawDate = parts[0];

    if (/^\d{4}-\d{2}-\d{2}$/.test(rawDate)) {
      parsedDate = rawDate;
    } else if (/^\d{1,2}\/\d{1,2}\/\d{4}$/.test(rawDate)) {
      const [d, m, y] = rawDate.split("/");
      const yearNum = parseInt(y, 10);
      const gregorianYear = yearNum > 2500 ? yearNum - 543 : yearNum;
      parsedDate = `${gregorianYear}-${m.padStart(2, "0")}-${d.padStart(2, "0")}`;
    } else if (/^\d{1,2}-\d{1,2}-\d{4}$/.test(rawDate)) {
      const [d, m, y] = rawDate.split("-");
      const yearNum = parseInt(y, 10);
      const gregorianYear = yearNum > 2500 ? yearNum - 543 : yearNum;
      parsedDate = `${gregorianYear}-${m.padStart(2, "0")}-${d.padStart(2, "0")}`;
    } else {
      rawDate = parts[1];
      if (/^\d{4}-\d{2}-\d{2}$/.test(rawDate)) {
        parsedDate = rawDate;
      }
    }

    if (!parsedDate) continue;

    // Detect amounts and description
    let withdrawal = 0;
    let deposit = 0;
    let description = "";
    let reference = "";
    let channel: BankChannel = "transfer";

    // Scan parts for numbers
    const numericParts: { index: number; val: number }[] = [];
    const textParts: string[] = [];

    for (let i = 1; i < parts.length; i++) {
      const cleanNum = parts[i].replace(/[^\d.-]/g, "");
      const val = parseFloat(cleanNum);
      if (!isNaN(val) && cleanNum.length > 0 && /^-?\d+(\.\d+)?$/.test(cleanNum)) {
        numericParts.push({ index: i, val });
      } else if (parts[i].length > 0) {
        textParts.push(parts[i]);
      }
    }

    description = textParts.join(" ") || "รายการธนาคาร";

    // Detect channel based on keywords
    const descLower = description.toLowerCase();
    if (descLower.includes("promptpay") || descLower.includes("พร้อมเพย์")) {
      channel = "promptpay";
    } else if (descLower.includes("fee") || descLower.includes("ค่าธรรมเนียม")) {
      channel = "fee";
    } else if (descLower.includes("int") || descLower.includes("ดอกเบี้ย")) {
      channel = "interest";
    } else if (descLower.includes("edc") || descLower.includes("รูดบัตร")) {
      channel = "edc";
    } else if (descLower.includes("chq") || descLower.includes("เช็ค")) {
      channel = "cheque";
    } else if (descLower.includes("atm")) {
      channel = "atm";
    }

    // Assign deposit / withdrawal
    if (numericParts.length >= 2) {
      const w = numericParts[0].val;
      const d = numericParts[1].val;
      if (w > 0 && d === 0) {
        withdrawal = w;
      } else if (d > 0 && w === 0) {
        deposit = d;
      } else if (w > 0 && d > 0) {
        withdrawal = w;
        deposit = d;
      } else {
        const amt = numericParts[0].val;
        if (amt < 0) withdrawal = Math.abs(amt);
        else deposit = amt;
      }
    } else if (numericParts.length === 1) {
      const amt = numericParts[0].val;
      if (amt < 0) withdrawal = Math.abs(amt);
      else deposit = amt;
    }

    const type: BankTransactionType = withdrawal > 0 ? "withdrawal" : "deposit";
    const amount = type === "withdrawal" ? withdrawal : deposit;

    if (amount <= 0) continue;

    const refMatch = description.match(/\b[A-Z0-9]{6,16}\b/i);
    if (refMatch) {
      reference = refMatch[0];
    }

    results.push({
      id: `stmt-${parsedDate.replace(/-/g, "")}-${seq++}`,
      date: parsedDate,
      type,
      amount: Math.round(amount * 100) / 100,
      description,
      reference: reference || undefined,
      channel,
      status: "unmatched",
    });
  }

  return results;
}

/**
 * 2-Way Auto Match between Bank Statement and Book Transactions.
 * Matches deposit with receipt, withdrawal with payment.
 * Considers date window (default +-3 days) and reference matching.
 */
export function autoReconcileBankStatement(
  statements: BankStatementTransaction[],
  books: BookTransaction[],
  options: { dateWindowDays?: number } = {},
): {
  matchedStatements: BankStatementTransaction[];
  matchedBooks: BookTransaction[];
  matchCount: number;
} {
  const windowDays = options.dateWindowDays ?? 3;
  const updatedStatements = statements.map((s) => ({ ...s }));
  const updatedBooks = books.map((b) => ({ ...b }));

  let matchCount = 0;

  for (const stmt of updatedStatements) {
    if (stmt.status === "matched" || stmt.status === "cleared") continue;

    let bestMatchIdx = -1;
    let bestScore = 0;

    for (let i = 0; i < updatedBooks.length; i++) {
      const book = updatedBooks[i];
      if (book.isReconciled) continue;

      const directionMatches =
        (stmt.type === "deposit" && (book.type === "receipt" || book.type === "journal")) ||
        (stmt.type === "withdrawal" && (book.type === "payment" || book.type === "journal"));

      if (!directionMatches) continue;

      const amtDiff = Math.abs(stmt.amount - book.amount);
      if (amtDiff > 0.009) continue;

      const stmtTime = new Date(stmt.date).getTime();
      const bookTime = new Date(book.date).getTime();
      const dayDiff = Math.abs(stmtTime - bookTime) / (1000 * 60 * 60 * 24);

      if (dayDiff > windowDays) continue;

      let score = 0.8;
      if (dayDiff === 0) score += 0.15;
      else if (dayDiff <= 1) score += 0.1;

      if (
        stmt.reference &&
        book.reference &&
        stmt.reference.toLowerCase() === book.reference.toLowerCase()
      ) {
        score += 0.2;
      }

      if (score > bestScore) {
        bestScore = score;
        bestMatchIdx = i;
      }
    }

    if (bestMatchIdx >= 0 && bestScore >= 0.8) {
      const matchedBook = updatedBooks[bestMatchIdx];
      stmt.status = "matched";
      stmt.matchedJournalDocNo = matchedBook.docno;
      stmt.matchConfidence = Math.min(1.0, Math.round(bestScore * 100) / 100);

      matchedBook.isReconciled = true;
      matchedBook.reconciledWithId = stmt.id;
      matchCount++;
    }
  }

  return {
    matchedStatements: updatedStatements,
    matchedBooks: updatedBooks,
    matchCount,
  };
}

/**
 * Computes official Bank Reconciliation Statement according to Thai accounting standards.
 */
export function calculateBankReconciliation(params: {
  statementEndingBalance: number;
  statementTransactions: BankStatementTransaction[];
  bookEndingBalance: number;
  bookTransactions: BookTransaction[];
}): BankReconciliationSummary {
  const {
    statementEndingBalance,
    statementTransactions,
    bookEndingBalance,
    bookTransactions,
  } = params;

  // Deposits in Transit: Receipts recorded in books but not yet in bank statement
  const depositsInTransitList = bookTransactions.filter(
    (b) => !b.isReconciled && (b.type === "receipt" || (b.type === "journal" && b.amount > 0)),
  );
  const depositsInTransit = depositsInTransitList.reduce((sum, b) => sum + b.amount, 0);

  // Outstanding Cheques: Payments issued in books but not yet cleared by bank
  const outstandingChequesList = bookTransactions.filter(
    (b) => !b.isReconciled && (b.type === "payment" || (b.type === "journal" && b.amount < 0)),
  );
  const outstandingCheques = outstandingChequesList.reduce(
    (sum, b) => sum + Math.abs(b.amount),
    0,
  );

  const adjustedBankBalance =
    Math.round((statementEndingBalance + depositsInTransit - outstandingCheques) * 100) / 100;

  // Unrecorded Bank Charges / Fees
  const unrecordedChargesList = statementTransactions.filter(
    (s) => s.status === "unmatched" && s.type === "withdrawal",
  );
  const unrecordedBankCharges = unrecordedChargesList.reduce((sum, s) => sum + s.amount, 0);

  // Unrecorded Bank Interest
  const unrecordedInterestList = statementTransactions.filter(
    (s) => s.status === "unmatched" && s.type === "deposit" && s.channel === "interest",
  );
  const unrecordedBankInterest = unrecordedInterestList.reduce((sum, s) => sum + s.amount, 0);

  // Unrecorded customer transfers in
  const unrecordedTransfersInList = statementTransactions.filter(
    (s) => s.status === "unmatched" && s.type === "deposit" && s.channel !== "interest",
  );
  const unrecordedTransfersIn = unrecordedTransfersInList.reduce((sum, s) => sum + s.amount, 0);

  const adjustedBookBalance =
    Math.round(
      (bookEndingBalance + unrecordedBankInterest + unrecordedTransfersIn - unrecordedBankCharges) *
        100,
    ) / 100;

  const difference = Math.round(Math.abs(adjustedBankBalance - adjustedBookBalance) * 100) / 100;
  const isBalanced = difference < 0.01;

  return {
    statementEndingBalance,
    depositsInTransit: Math.round(depositsInTransit * 100) / 100,
    depositsInTransitCount: depositsInTransitList.length,
    outstandingCheques: Math.round(outstandingCheques * 100) / 100,
    outstandingChequesCount: outstandingChequesList.length,
    adjustedBankBalance,
    bookEndingBalance,
    unrecordedBankCharges: Math.round(unrecordedBankCharges * 100) / 100,
    unrecordedBankInterest: Math.round(unrecordedBankInterest * 100) / 100,
    unrecordedTransfersIn: Math.round(unrecordedTransfersIn * 100) / 100,
    adjustedBookBalance,
    difference,
    isBalanced,
  };
}

/**
 * 1-Click Fast Clearing Journal Line Generator for Unmatched Bank Items.
 */
export function generateBankFastVoucher(
  stmt: BankStatementTransaction,
  bankAccountCode: string = "111200",
  expenseOrIncomeCode?: string,
): {
  journalType: "PV" | "RV" | "JV";
  description: string;
  date: string;
  lines: { accountCode: string; debit: number; credit: number; description: string }[];
} {
  const amt = stmt.amount;

  if (stmt.type === "withdrawal") {
    const expCode = expenseOrIncomeCode || (stmt.channel === "fee" ? "530101" : "590101");
    return {
      journalType: "PV",
      description: stmt.description || "ค่าใช้จ่าย/ค่าธรรมเนียมธนาคาร",
      date: stmt.date,
      lines: [
        {
          accountCode: expCode,
          debit: amt,
          credit: 0,
          description: stmt.description,
        },
        {
          accountCode: bankAccountCode,
          debit: 0,
          credit: amt,
          description: "ตัดบัญชีเงินฝากธนาคาร",
        },
      ],
    };
  } else {
    const incCode =
      expenseOrIncomeCode || (stmt.channel === "interest" ? "420101" : "219101");
    return {
      journalType: "RV",
      description: stmt.description || "รับเงินโอนเข้าบัญชีธนาคาร",
      date: stmt.date,
      lines: [
        {
          accountCode: bankAccountCode,
          debit: amt,
          credit: 0,
          description: "เงินฝากธนาคาร",
        },
        {
          accountCode: incCode,
          debit: 0,
          credit: amt,
          description: stmt.description,
        },
      ],
    };
  }
}

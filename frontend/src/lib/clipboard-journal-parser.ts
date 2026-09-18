import type { GLLine } from "./general-ledger";

/**
 * Normalizes numeric currency text from Excel / clipboard:
 * - Strips currency symbols (฿, $), commas (,), spaces
 * - Replaces dashes ("-", "—") with ""
 */
export function cleanNumericAmount(val: string | undefined): string {
  if (!val) return "";
  const cleaned = val.replace(/[฿$,\s]/g, "").trim();
  if (cleaned === "-" || cleaned === "—" || cleaned === "") return "";
  const num = Number(cleaned);
  return isNaN(num) ? "" : cleaned;
}

/**
 * Detects whether a row contains typical header titles from Excel
 */
export function isHeaderRow(cells: string[]): boolean {
  const text = cells.join(" ").toLowerCase();
  return (
    text.includes("account") ||
    text.includes("รหัส") ||
    text.includes("เดบิต") ||
    text.includes("debit") ||
    text.includes("เครดิต") ||
    text.includes("credit") ||
    text.includes("คำอธิบาย") ||
    text.includes("description")
  );
}

/**
 * Parses tab-separated (or CSV) clipboard text into structured GLLine array:
 * Supports:
 * - 4 columns: [AccountCode] [Description] [Debit] [Credit]
 * - 3 columns: [AccountCode] [Debit] [Credit]
 * - 5+ columns: [AccountCode] [Description] [Debit] [Credit] [Department] [Project]
 */
export function parseClipboardJournalLines(clipboardText: string): GLLine[] {
  if (!clipboardText || !clipboardText.trim()) return [];

  const rawLines = clipboardText
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean);

  const results: GLLine[] = [];

  for (let i = 0; i < rawLines.length; i++) {
    const rawRow = rawLines[i];
    // Split by tab (standard for Excel/Sheets copy-paste), fallback to comma if tab not present
    const cells = rawRow.includes("\t")
      ? rawRow.split("\t").map((c) => c.trim())
      : rawRow.split(",").map((c) => c.trim());

    if (cells.length < 2) continue;

    // Check if the very first row is a header title row
    if (i === 0 && isHeaderRow(cells)) {
      continue;
    }

    let accountcode = "";
    let description = "";
    let debit = "";
    let credit = "";
    let departmentcode = "";
    let projectcode = "";

    if (cells.length === 2) {
      accountcode = cells[0];
      const amt = cleanNumericAmount(cells[1]);
      if (amt.startsWith("-")) {
        credit = amt.replace("-", "");
      } else {
        debit = amt;
      }
    } else if (cells.length === 3) {
      // Check if cells[1] is numeric (Account, Debit, Credit)
      const num1 = cleanNumericAmount(cells[1]);
      const num2 = cleanNumericAmount(cells[2]);
      if (num1 !== "" || num2 !== "") {
        accountcode = cells[0];
        debit = Number(num1) === 0 ? "" : num1;
        credit = Number(num2) === 0 ? "" : num2;
      } else {
        accountcode = cells[0];
        description = cells[1];
        debit = cleanNumericAmount(cells[2]);
      }
    } else if (cells.length >= 4) {
      // Standard 4+ columns: [Account] [Description] [Debit] [Credit] [Dept] [Proj]
      accountcode = cells[0];
      description = cells[1];
      debit = cleanNumericAmount(cells[2]);
      credit = cleanNumericAmount(cells[3]);
      if (Number(debit) === 0) debit = "";
      if (Number(credit) === 0) credit = "";

      if (cells.length >= 5) {
        departmentcode = cells[4];
      }
      if (cells.length >= 6) {
        projectcode = cells[5];
      }
    }

    if (accountcode || debit || credit) {
      results.push({
        accountcode,
        description,
        debit,
        credit,
        departmentcode,
        projectcode,
        cashflow: "",
      });
    }
  }

  return results;
}

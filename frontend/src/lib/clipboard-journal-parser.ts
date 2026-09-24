import type { GLLine } from "./general-ledger";

/** A pasted cell that cannot become a journal amount — the whole paste is rejected so nothing changes silently. */
export type ClipboardJournalIssue = {
  /** 1-based row number in the pasted block (header row included), so the user can find it in Excel */
  row: number;
  field: "debit" | "credit" | "amount";
  value: string;
  problem: "invalid" | "negative";
};

export type ClipboardJournalParse = { lines: GLLine[]; issues: ClipboardJournalIssue[] };

type ClipboardAmount = { value: string; negative: boolean; valid: boolean };

/**
 * Reads one amount cell copied from Excel / Google Sheets without changing its value:
 * - blank, "-", "–", "—" (Excel accounting format for zero) → "0"
 * - Thai digits → 0-9; thousands commas, spaces (incl. NBSP), baht / dollar signs removed
 * - accounting negative "(1,500.00)" and "-1500" → negative = true, value = "1500.00" / "1500"
 * - anything else (letters, "1500-", two decimal points) → valid = false (reported, never dropped)
 * - commas that are not thousands separators ("1.500,00" European format, "1,5") → valid = false:
 *   removing them would change the value 1,000 times silently
 * Values come back canonical (".5" → "0.5", "0500" → "500") so save-time validation accepts what the screen shows.
 * Zero values always come back as "0".
 */
export function parseClipboardAmount(cell: string | undefined): ClipboardAmount {
  let text = (cell ?? "")
    .replace(/[๐-๙]/g, (digit) => String(digit.charCodeAt(0) - 0x0e50))
    .replace(/[\s฿$]/g, "");
  if (text === "" || text === "-" || text === "–" || text === "—") {
    return { value: "0", negative: false, valid: true };
  }
  let negative = false;
  const parenthesised = /^\((.*)\)$/.exec(text);
  if (parenthesised) {
    negative = true;
    text = parenthesised[1];
  }
  if (text.startsWith("-")) {
    if (negative) return { value: text, negative: true, valid: false };
    negative = true;
    text = text.slice(1);
  }
  if (!thousandsCommasValid(text)) return { value: text, negative, valid: false };
  text = text.replace(/,/g, "");
  if (!/^(\d+(\.\d*)?|\.\d+)$/.test(text)) return { value: text, negative, valid: false };
  const isZero = !/[1-9]/.test(text);
  if (isZero) return { value: "0", negative: false, valid: true };
  const [whole = "", fraction = ""] = text.split(".");
  const canonicalWhole = whole.replace(/^0+(?=\d)/, "") || "0";
  return { value: fraction ? `${canonicalWhole}.${fraction}` : canonicalWhole, negative, valid: true };
}

/** Commas in an unsigned amount are thousands separators only ("1,234,567.89"); "1.500,00" or "1,5" are not. */
export function thousandsCommasValid(unsigned: string): boolean {
  return !unsigned.includes(",") || /^\d{1,3}(,\d{3})+(\.\d*)?$/.test(unsigned);
}

/**
 * The table takes over a paste only when it is a block of cells: a tab, or two or more non-empty lines.
 * One cell copied from Excel arrives as "1,500.00\r\n" — that must paste into the focused box as usual.
 */
export function isTabularPaste(text: string): boolean {
  const body = (text ?? "").replace(/(\r\n|\r|\n)$/, "");
  return body.includes("\t") || body.split(/\r\n|\r|\n/).filter((line) => line.trim()).length >= 2;
}

/** Detects the header row that Excel copies along with the data (Thai or English titles) */
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
 * Splits one pasted row on tabs (Excel/Sheets) or commas (CSV), honouring "quoted, cells" so "1,500.00" stays one cell.
 * A row without a tab that is one number with thousands commas (a single Excel amount cell) stays one cell —
 * splitting "12,345,678.90" on commas made account "12", debit "345", credit "678.90".
 */
function splitRow(row: string): string[] {
  if (!row.includes("\t") && /^[(-]?[฿$]?\d{1,3}(,\d{3})+(\.\d*)?\)?$/.test(row.trim())) return [row.trim()];
  const separator = row.includes("\t") ? "\t" : ",";
  const cells: string[] = [];
  let current = "";
  let quoted = false;
  for (let index = 0; index < row.length; index += 1) {
    const char = row[index];
    if (char === '"') {
      if (quoted && row[index + 1] === '"') {
        current += '"';
        index += 1;
      } else {
        quoted = !quoted;
      }
    } else if (char === separator && !quoted) {
      cells.push(current.trim());
      current = "";
    } else {
      current += char;
    }
  }
  cells.push(current.trim());
  return cells;
}

/**
 * Parses tab-separated (or CSV) clipboard text into journal lines. Cells are trimmed, rows are not,
 * so an empty last cell (debit filled, credit blank) keeps its column. Layouts by column count:
 * - 4+ columns: [Account] [Description] [Debit] [Credit] [Department] [Project]
 * - 3 columns:  [Account] [Debit] [Credit] when the 2nd cell is blank or a number,
 *               otherwise [Account] [Description] [Signed amount]
 * - 2 columns:  [Account] [Signed amount]
 * Signed amount rule (documented): positive → debit, negative ("-1500" or "(1,500.00)") → credit.
 * In Debit/Credit columns a negative or non-numeric amount is an issue: the caller must reject the paste.
 * The side without an amount is "0" (never ""), because the backend does not accept an empty amount.
 */
export function parseClipboardJournalLines(clipboardText: string): ClipboardJournalParse {
  const lines: GLLine[] = [];
  const issues: ClipboardJournalIssue[] = [];
  if (!clipboardText || !clipboardText.trim()) return { lines, issues };

  const rows = clipboardText.split(/\r?\n|\r/);
  let firstDataRow = true;
  rows.forEach((row, index) => {
    if (!row.trim()) return;
    const rowNo = index + 1;
    const cells = splitRow(row);
    if (cells.length < 2) return;
    if (firstDataRow) {
      firstDataRow = false;
      // A header has titles, not numbers — a data row whose description happens to contain "รหัส" is kept
      if (isHeaderRow(cells) && !cells.slice(1).some((cell) => cell !== "" && parseClipboardAmount(cell).valid)) return;
    }

    const accountcode = cells[0];
    let description = "";
    let debit = "0";
    let credit = "0";
    let departmentcode = "";
    let projectcode = "";

    const signed = (cell: string) => {
      const amount = parseClipboardAmount(cell);
      if (!amount.valid) {
        issues.push({ row: rowNo, field: "amount", value: cell, problem: "invalid" });
        return;
      }
      if (amount.negative) credit = amount.value;
      else debit = amount.value;
    };
    const sided = (cell: string, field: "debit" | "credit") => {
      const amount = parseClipboardAmount(cell);
      if (!amount.valid || amount.negative) {
        issues.push({ row: rowNo, field, value: cell, problem: amount.valid ? "negative" : "invalid" });
        return "0";
      }
      return amount.value;
    };

    if (cells.length === 2) {
      signed(cells[1]);
    } else if (cells.length === 3) {
      const second = parseClipboardAmount(cells[1]);
      if (second.valid) {
        debit = sided(cells[1], "debit");
        credit = sided(cells[2], "credit");
      } else {
        description = cells[1];
        signed(cells[2]);
      }
    } else {
      description = cells[1];
      debit = sided(cells[2], "debit");
      credit = sided(cells[3], "credit");
      departmentcode = cells[4] ?? "";
      projectcode = cells[5] ?? "";
    }

    lines.push({ accountcode, description, debit, credit, departmentcode, projectcode, cashflow: "" });
  });

  return { lines, issues };
}

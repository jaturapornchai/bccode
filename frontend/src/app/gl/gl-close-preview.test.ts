import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { readFileSync } from "node:fs";
import { describe, expect, it, vi } from "vitest";
import { glRequest } from "@/lib/general-ledger-api";
import type { GLClosePreview } from "@/lib/general-ledger";
import { ClosePreviewTable, fetchClosePreview, isZeroAmount } from "./gl-close-preview";

vi.mock("@/lib/general-ledger-api", () => ({ glRequest: vi.fn() }));

const sample: GLClosePreview = {
  fiscalyear: "2026", date: "2026-02-28", scale: 2, pendingdrafts: false, netprofit: "15.25", totaldebit: "1500000.50", totalcredit: "1500000.50", balanced: true,
  branches: [{
    branchcode: "00001", docno: "CLOSE-FEB-001", netprofit: "75.25", totaldebit: "1500000.50", totalcredit: "1500000.50",
    lines: [
      { accountcode: "41000", accountname: "รายได้จากการขาย", departmentcode: "D1", projectcode: "", debit: "1500000.50", credit: "0.00", transfer: false },
      { accountcode: "32000", accountname: "กำไรขาดทุนสุทธิ", departmentcode: "D1", projectcode: "", debit: "0.00", credit: "1500000.50", transfer: true },
    ],
  }],
};

describe("closing entries preview (Champ 5504)", () => {
  it("asks the backend for processes/close-preview with the typed document number only when filled", async () => {
    vi.mocked(glRequest).mockResolvedValue({ closepreview: sample });
    await fetchClosePreview({ year: "2026", version: 4, date: "2026-02-28", docno: " CLOSE-FEB " });
    const [path, init] = vi.mocked(glRequest).mock.calls.at(-1)!;
    expect(path).toBe("command");
    expect(JSON.parse(String(init?.body))).toMatchObject({ resource: "processes", action: "close-preview", id: "2026", version: 4, date: "2026-02-28", docno: "CLOSE-FEB" });
    await fetchClosePreview({ year: "2026", version: 4, date: "2026-02-28", docno: "  " });
    expect(JSON.parse(String(vi.mocked(glRequest).mock.calls.at(-1)![1]?.body))).not.toHaveProperty("docno");
  });

  it("returns null when the backend sends no preview and fills the missing parts", async () => {
    vi.mocked(glRequest).mockResolvedValue({});
    expect(await fetchClosePreview({ year: "2026", date: "2026-02-28", docno: "" })).toBeNull();
    vi.mocked(glRequest).mockResolvedValue({ closepreview: { branches: [{ branchcode: "B1", docno: "", netprofit: "0", totaldebit: "0", totalcredit: "0" }] } });
    const preview = await fetchClosePreview({ year: "2026", date: "2026-02-28", docno: "" });
    expect(preview).toMatchObject({ scale: 2, pendingdrafts: false, balanced: false });
    expect(preview?.branches[0].lines).toEqual([]);
  });

  it("treats only zero decimal strings as the empty side of a line", () => {
    for (const value of ["", "0", "0.00", "-0.00", "0.0000"]) expect(isZeroAmount(value)).toBe(true);
    for (const value of ["0.01", "10", "100.50", "-5.00", "1500000.50"]) expect(isZeroAmount(value)).toBe(false);
  });

  it("shows each branch voucher with the transfer section, formatted amounts and the debit = credit proof", () => {
    const html = renderToStaticMarkup(createElement(ClosePreviewTable, { preview: sample }));
    expect(html).toContain("CLOSE-FEB-001");
    expect(html).toContain("41000");
    expect(html).toContain("1,500,000.50");
    expect(html).toContain("โอนกำไรขาดทุนสุทธิเข้ากำไรสะสม");
    expect(html).toContain("ยอดเดบิตเท่ากับเครดิต");
    expect(html).not.toContain("0.00");
    expect(html).not.toContain('role="alert"');
    expect(html.match(/data-close-line="transfer"/g)).toHaveLength(1);
  });

  it("warns and states the mismatch instead of hiding them", () => {
    const html = renderToStaticMarkup(createElement(ClosePreviewTable, { preview: { ...sample, pendingdrafts: true, balanced: false } }));
    expect(html).toContain('role="alert"');
    expect(html).toContain("ยังมีรายการร่าง");
    expect(html).toContain("ยอดเดบิตไม่เท่ากับเครดิต");
  });
});

describe("close screen uses the real closing entries", () => {
  it("no longer builds the close preview from the working paper report", () => {
    const source = readFileSync(new URL("./gl-processes.tsx", import.meta.url), "utf8");
    expect(source).toContain("fetchClosePreview(");
    expect(source).not.toMatch(/action === "close" \? date : year\.enddate/);
  });
});

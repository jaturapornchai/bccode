import { describe, expect, it } from "vitest";
import type { TaxFormCatalogItem } from "@/lib/tax-forms";
import { periodForForm } from "./tax-form-editor";

const monthly = { code: "PND3", title: "ภ.ง.ด.3", period: "month" } as TaxFormCatalogItem;
const yearly = { code: "PND50", title: "ภ.ง.ด.50", period: "year" } as TaxFormCatalogItem;
const today = new Date(2026, 8, 24); // 24 ก.ย. 2569 → ค่าเริ่มต้นรายเดือน = ส.ค.

describe("periodForForm", () => {
  // UAT 2026-09-24: ตั้ง ต.ค. ใน ภ.ง.ด.53 แล้วเปลี่ยนเป็น ภ.ง.ด.3 เดือนเด้งกลับเป็น ส.ค.
  it("keeps the chosen month when switching between monthly forms", () => {
    expect(periodForForm(monthly, { year: 2026, month: 10 }, today)).toEqual({ year: 2026, month: 10 });
  });

  it("uses the default period on first open or when the period kind changes", () => {
    expect(periodForForm(monthly, null, today)).toEqual({ year: 2026, month: 8 });
    expect(periodForForm(monthly, { year: 2025, month: 0 }, today)).toEqual({ year: 2026, month: 8 });
    expect(periodForForm(yearly, { year: 2026, month: 10 }, today)).toEqual({ year: 2025, month: 0 });
    expect(periodForForm(yearly, { year: 2024, month: 0 }, today)).toEqual({ year: 2024, month: 0 });
  });
});

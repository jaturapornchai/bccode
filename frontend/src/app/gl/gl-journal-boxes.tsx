"use client";

import type { KeyboardEvent, ReactNode } from "react";
import { Calculator, Table2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { isBlankVatRow, taxBoxView } from "@/lib/gl-journal-entry";
import type { GLDetailVat, GLJournalDetails } from "@/lib/gl-journal-details";
import { AmountInput, actionClass, control, useGLText } from "./gl-common";

// กรอบข้อมูลของจอบันทึกข้อมูลรายวันแบบ Champ (CGLFrmJournal, docs/kms/champ/09-gl.md §2):
// กรอบหัวเอกสาร + กรอบภาษี วางป้ายซ้ายช่องสองคอลัมน์เมื่อพื้นที่ของแผงกว้างพอ (container query — แผงแบ่งจอยืดหดได้ จึงไม่ใช้ความกว้างหน้าจอ)

/** กรอบกลุ่มข้อมูล (group box ของ Champ) — ชื่อกรอบเป็น legend ให้โปรแกรมอ่านหน้าจอรู้ว่าช่องเป็นของกลุ่มไหน */
export function ChampGroupBox({ title, children, className, onKeyDown }: { title: string; children: ReactNode; className?: string; onKeyDown?: (event: KeyboardEvent<HTMLFieldSetElement>) => void }) {
  return (
    <fieldset className={cn("grid min-w-0 gap-x-6 gap-y-2.5 rounded-2xl border border-border bg-card/70 px-3 pb-3 pt-1 shadow-sm @2xl:grid-cols-2", className)} onKeyDown={onKeyDown}>
      <legend className="px-1.5 text-[0.9rem] font-semibold text-muted-foreground">{title}</legend>
      {children}
    </fieldset>
  );
}

/** ช่องในกรอบ: แผงแคบ = ป้ายอยู่บนช่อง, แผงกว้าง (@2xl) = ป้ายชิดขวาอยู่หน้าช่องแบบ Champ (สูงเท่าช่อง 2.6em จึงตรงกับช่อง แม้ใต้ช่องมีคำแนะนำ/ปุ่ม); wide = กินสองคอลัมน์ */
export function ChampField({ label, children, hint, wide = false }: { label: string; children: ReactNode; hint?: string; wide?: boolean }) {
  return (
    <div className={cn("grid min-w-0 gap-1 text-[0.95rem] leading-normal @2xl:grid-cols-[8.5rem_minmax(0,1fr)] @2xl:items-start @2xl:gap-x-3", wide && "@2xl:col-span-2")}>
      <span className="font-medium text-foreground @2xl:flex @2xl:min-h-[2.6em] @2xl:items-center @2xl:justify-end @2xl:text-right">{label}</span>
      <div className="grid min-w-0 gap-1">{children}{hint && <span className="text-[0.9rem] text-muted-foreground">{hint}</span>}</div>
    </div>
  );
}

/** ช่องว่างแทนตำแหน่งที่ Champ ไม่มีช่อง (เช่น ขวาของสมุดรายวัน) ให้ช่องถัดไปลงคอลัมน์เดียวกับ Champ */
export function ChampSpacer() {
  return <div aria-hidden className="hidden @2xl:block" />;
}

/** ป้ายปุ่มลัดบนปุ่ม (F11, F3 ...) — ข้อความบนปุ่มยังเป็นภาษาไทยเสมอ */
export function ChampKey({ children, primary = false }: { children: ReactNode; primary?: boolean }) {
  return (
    <kbd className={cn("ml-auto rounded border px-1.5 py-0.5 font-mono text-[0.75rem]", primary ? "border-primary-foreground/30 bg-primary-foreground/15 text-primary-foreground" : "border-border bg-muted/60 text-muted-foreground")}>
      {children}
    </kbd>
  );
}

const cell = (field: string) => ({ "data-detail-field": `vats.1.${field}` });

/**
 * กรอบภาษีของ Champ = ตัวแก้ด่วนของใบกำกับภาษีใบเดียว (journal.details.vats[0]); ข้อมูลจริงอยู่ในตารางภาษี F3 ชุดเดียวกัน
 * ไม่มีใบกำกับ = "ไม่มีภาษี" (ช่องอื่นปิดไว้ ไม่สร้างแถวภาษีเปล่า); 2 ใบขึ้นไป = แก้ในตารางภาษีเท่านั้น
 * ช่องแต่ละช่องมี data-detail-field="vats.1.<field>" ให้ข้อความตรวจก่อนบันทึกพามาที่ช่องในกรอบนี้ก่อนตาราง
 */
export function ChampTaxBox({ details, scale, disabled, calcBusy = false, onTaxType, onPatch, onCalc, onOpenTaxTable }: {
  details: GLJournalDetails | undefined;
  scale: number;
  disabled: boolean;
  calcBusy?: boolean;
  onTaxType?: (value: "" | "1" | "2") => void;
  onPatch?: (patch: Partial<GLDetailVat>) => void;
  onCalc?: (mode: "exclusive" | "inclusive") => void;
  onOpenTaxTable: () => void;
}) {
  const tr = useGLText();
  const view = taxBoxView(details);
  const title = tr("gl_tax_box_title", "ภาษีมูลค่าเพิ่ม");
  if (view.mode === "multi") {
    return (
      <ChampGroupBox title={title}>
        <div className="flex flex-wrap items-center gap-3 @2xl:col-span-2" role="status">
          <span className="font-medium">{tr("gl_tax_box_multi", "มีใบกำกับ {0} ใบ — แก้ในตารางภาษี F3").replace("{0}", String(view.count))}</span>
          <Button type="button" variant="outline" className={actionClass} onClick={onOpenTaxTable}>
            <Table2 className="mr-1.5 size-4" />{tr("gl_tax_table", "ตารางภาษี")}<ChampKey>F3</ChampKey>
          </Button>
        </div>
      </ChampGroupBox>
    );
  }
  const row = view.mode === "single" ? view.row : undefined;
  const locked = disabled || !row;
  const rate = String(row?.vat_rate ?? "").trim();
  const missingRate = row !== undefined && !disabled && !rate && !isBlankVatRow(row);
  const patch = (value: Partial<GLDetailVat>) => { if (!locked) onPatch?.(value); };
  const typeValue = row ? (Number(row.tax_type) === 2 ? "2" : "1") : "";
  return (
    <ChampGroupBox title={title}>
      <ChampField label={tr("gl_tax_type", "ประเภทภาษี")}>
        <div {...cell("tax_type")} className="contents">
          <select className={control} aria-label={tr("gl_tax_type", "ประเภทภาษี")} value={typeValue} disabled={disabled || !onTaxType} onChange={(event) => onTaxType?.(event.target.value as "" | "1" | "2")}>
            <option value="">{tr("gl_tax_type_none", "ไม่มีภาษี")}</option>
            <option value="2">{tr("gl_tax_type_output", "ภาษีขาย")}</option>
            <option value="1">{tr("gl_tax_type_input", "ภาษีซื้อ")}</option>
          </select>
        </div>
      </ChampField>
      <ChampField label={tr("gl_detail_vats_vat_rate", "อัตราภาษี (%)")}>
        <div {...cell("vat_rate")} className="contents">
          <AmountInput ariaLabel={tr("gl_detail_vats_vat_rate", "อัตราภาษี (%)")} value={rate} scale={2} allowEmpty disabled={locked} placeholder={tr("gl_example_7", "เช่น 7")} onChange={(value) => patch({ vat_rate: value })} />
        </div>
        {missingRate && <span role="alert" className="text-[0.9rem] leading-snug text-destructive">{tr("gl_detail_rate_required_inline", "กรุณาใส่อัตราภาษี (%) — ระบบไม่ใส่ 0% ให้เอง")}</span>}
      </ChampField>
      <ChampField label={tr("gl_detail_vats_tax_invoice_no", "เลขที่ใบกำกับภาษี")}>
        <div {...cell("tax_invoice_no")} className="contents">
          <input className={control} aria-label={tr("gl_detail_vats_tax_invoice_no", "เลขที่ใบกำกับภาษี")} maxLength={50} value={row?.tax_invoice_no ?? ""} disabled={locked} onChange={(event) => patch({ tax_invoice_no: event.target.value })} />
        </div>
      </ChampField>
      <ChampField label={tr("gl_detail_vats_tax_invoice_date", "วันที่ใบกำกับภาษี")}>
        <div {...cell("tax_invoice_date")} className="contents">
          <input className={control} type="date" aria-label={tr("gl_detail_vats_tax_invoice_date", "วันที่ใบกำกับภาษี")} value={row?.tax_invoice_date ?? ""} disabled={locked} onChange={(event) => patch({ tax_invoice_date: event.target.value })} />
        </div>
      </ChampField>
      <ChampField label={tr("gl_amount", "จำนวนเงิน")}>
        <div {...cell("base_amount")} className="contents">
          <AmountInput ariaLabel={tr("gl_amount", "จำนวนเงิน")} value={row ? String(row.base_amount ?? "") : ""} scale={scale} allowEmpty={false} disabled={locked} onChange={(value) => patch({ base_amount: value || "0" })} />
        </div>
        {onCalc && (
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" className={actionClass} disabled={locked || calcBusy} onClick={() => onCalc("exclusive")} title={tr("gl_tax_calc_exclusive_hint", "จำนวนเงินคือฐานภาษี — คำนวณภาษีจากฐาน × อัตรา")}>
              <Calculator className="mr-1.5 size-4" />{tr("gl_tax_calc_exclusive", "ภาษีแยกนอก")}
            </Button>
            <Button type="button" variant="outline" className={actionClass} disabled={locked || calcBusy} onClick={() => onCalc("inclusive")} title={tr("gl_tax_calc_inclusive_hint", "จำนวนเงินรวมภาษีแล้ว — แยกฐานภาษีและภาษีออกจากยอดรวม")}>
              <Calculator className="mr-1.5 size-4" />{tr("gl_tax_calc_inclusive", "ภาษีรวมใน")}
            </Button>
          </div>
        )}
      </ChampField>
      <ChampField label={tr("gl_tax_box_vat_amount", "ภาษี")}>
        <div {...cell("vat_amount")} className="contents">
          <AmountInput ariaLabel={tr("gl_tax_box_vat_amount", "ภาษี")} value={row ? String(row.vat_amount ?? "") : ""} scale={scale} allowEmpty disabled={locked} placeholder={row ? tr("gl_wht_tax_auto", "เว้นว่าง = คำนวณ ฐาน × อัตรา ตอนบันทึก") : undefined} onChange={(value) => patch({ vat_amount: value === "" ? undefined : value })} />
        </div>
      </ChampField>
      <ChampField label={tr("gl_description", "คำอธิบาย")} wide>
        <div {...cell("remark")} className="contents">
          <input className={control} aria-label={tr("gl_tax_box_remark", "คำอธิบายภาษี")} maxLength={500} value={row?.remark ?? ""} disabled={locked} onChange={(event) => patch({ remark: event.target.value })} />
        </div>
      </ChampField>
    </ChampGroupBox>
  );
}

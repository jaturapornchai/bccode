"use client";
// พิมพ์สมุดบัญชีและใบสำคัญรายวันตามประกาศกรมทะเบียนการค้า เรื่อง กำหนดชนิดของบัญชีที่ต้องจัดทำฯ พ.ศ. 2544
// ข้อ 4 (ชื่อผู้มีหน้าที่จัดทำบัญชี ชนิดของบัญชี ลำดับเล่ม), ข้อ 5 (ชื่อบัญชี วันที่ เลขที่เอกสาร รายการ จำนวนเงิน + เลขหน้าเรียงทุกหน้า),
// ข้อ 11 (เอกสารที่ทำขึ้นใช้เอง: คำอธิบายรายการ วิธีคำนวณ ลายมือชื่อผู้จัดทำบัญชีหรือผู้อนุมัติ) — ทะเบียนอ้างอิง docs/kms/21 §11
// งบการเงิน (GLStatementTable) ตามประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566 + TFRS for NPAEs บทที่ 4 — docs/kms/21
// หน้าพิมพ์เป็นแค่การแสดงผล: ตัวเลขทุกตัวมาจาก API ของ backend ตามเดิม
import { useCallback, useEffect, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import type { LanguageCode } from "@/lib/i18n";
import { formatAppDate, localeForDate } from "@/lib/date-time";
import { amountString, formatAmount, journalBookName, journalTotals, type GLJournal, type GLJournalBook, type GLReport, type GLTextFn } from "@/lib/general-ledger";
import { companyBaseName, workspaceCompanyDisplayName, workspaceStorageKeys, type WorkspaceSession } from "@/lib/workspace-models";
import { useGLText } from "./gl-common";

export type GLPrintOrientation = "portrait" | "landscape";

/** เลขหน้า "หน้า n / N" ทุกหน้าด้วย @page margin box (Chrome 131+, Safari 18.2); ป้ายมาจาก languages.tsv จึงต้อง escape ให้เป็น CSS string */
export function printPageStyle(pageLabel: string): string {
  const label = JSON.stringify(pageLabel).replace(/</g, "\\3c ");
  return (["portrait", "landscape"] as const).map((orientation) =>
    `@page gl-print-${orientation} { size: A4 ${orientation}; margin: ${orientation === "portrait" ? "12mm 10mm 14mm" : "10mm 10mm 12mm"}; @bottom-right { content: ${label} " " counter(page) " / " counter(pages); font-size: 8pt; } }`,
  ).join("\n");
}

/** คอลัมน์ข้อความที่ว่างทุกแถว (เช่น แผนก/โครงการที่ไม่ได้ใช้) ไม่ต้องกินที่บนกระดาษ; คอลัมน์เงินเก็บไว้เสมอ */
export function visiblePrintColumns<T extends { key: string; amount?: boolean }>(columns: T[], rows: Record<string, string>[]): T[] {
  if (!rows.length) return columns;
  return columns.filter((column) => column.amount || rows.some((row) => (row[column.key] ?? "").trim() !== ""));
}

// สมุดบัญชีเว้นช่องเดบิต/เครดิตที่เป็นศูนย์ให้ว่าง (อ่านง่ายแบบสมุดเขียนมือ) — ยอดคงเหลือ/ยอดรวมยังแสดง 0.00 เพราะศูนย์มีความหมาย
export function printAmountCell(key: string, value: string, scale = 2): string {
  const text = formatAmount(value, scale);
  return /(debit|credit)$/.test(key) && /^0(.0+)?$/.test(text) ? "" : text;
}
export function printOrientation(columnCount: number): GLPrintOrientation {
  return columnCount > 7 ? "landscape" : "portrait";
}

/** ชื่อผู้มีหน้าที่จัดทำบัญชี = บริษัทที่เลือกตอนเข้าระบบ (ข้อมูลเดียวกับทะเบียนบริษัท) */
export function printCompanyName(): string {
  try {
    const workspace = JSON.parse(window.localStorage.getItem(workspaceStorageKeys.workspace) || "null") as WorkspaceSession | null;
    if (!workspace?.shop) return "";
    const name = workspace.company ? companyBaseName(workspace.company) : workspaceCompanyDisplayName(workspace);
    return name === "-" ? "" : name;
  } catch {
    return "";
  }
}

export function printedAtText(language: LanguageCode, now = new Date()): string {
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${formatAppDate(`${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`, language)} ${pad(now.getHours())}:${pad(now.getMinutes())}`;
}

/** พิมพ์เอกสารเดียวโดยซ่อนแอปทั้งหมด (globals.css "GL print"): เอกสารอยู่ใน portal ลูกของ body แล้วเรียก window.print() */
export function useGLPrint() {
  const tr = useGLText();
  const [job, setJob] = useState<{ node: ReactNode; orientation: GLPrintOrientation } | null>(null);
  const print = useCallback((node: ReactNode, orientation: GLPrintOrientation = "portrait") => setJob({ node, orientation }), []);
  useEffect(() => {
    if (!job) return;
    document.body.classList.add("gl-printing");
    const finish = () => setJob(null);
    window.addEventListener("afterprint", finish);
    // setTimeout (not requestAnimationFrame): rAF never fires while the tab is hidden, which left the print dialog unopened
    const timer = window.setTimeout(() => window.print(), 0);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener("afterprint", finish);
      document.body.classList.remove("gl-printing");
    };
  }, [job]);
  const portal = job
    ? createPortal(<div className="gl-print-root" data-orientation={job.orientation}><style>{printPageStyle(tr("common_page", "หน้า"))}</style>{job.node}</div>, document.body)
    : null;
  return { print, portal, printing: job !== null };
}

export function PrintHeading({ company, title, meta }: { company: string; title: string; meta: string[] }) {
  const lines = meta.filter(Boolean);
  return <>
    <div className="gl-print-company">{company}</div>
    <div className="gl-print-title">{title}</div>
    {lines.length > 0 && <div className="gl-print-meta">{lines.join(" · ")}</div>}
  </>;
}

const vatTypeLabels: Record<string, [string, string]> = { "1": ["gl_vat_tax_type_purchase", "ภาษีซื้อ (ใบกำกับที่ได้รับ)"], "2": ["gl_vat_tax_type_sale", "ภาษีขาย (ใบกำกับที่ออก)"] };
function whtFormLabel(form: string, tr: GLTextFn) {
  return form === "PND53" ? tr("wht_cert_ui_form_53", "ภ.ง.ด.53 (นิติบุคคล)") : form === "PND3" ? tr("wht_cert_ui_form_3", "ภ.ง.ด.3 (บุคคลธรรมดา)") : form === "PND2" ? tr("wht_cert_ui_form_2", "ภ.ง.ด.2 (ดอกเบี้ย/เงินปันผล)") : form;
}

/** ใบสำคัญรายวัน (ข้อ 9 + ข้อ 11): ผู้จัดทำ ชื่อเอกสาร เลขที่ วันที่ ยอดรวม คำอธิบาย วิธีคำนวณภาษี และช่องลายมือชื่อ */
export function GLVoucherPrint({ journal, books, company, tr, language, scale = 2 }: { journal: GLJournal; books: GLJournalBook[]; company: string; tr: GLTextFn; language: LanguageCode; scale?: number }) {
  const book = books.find((item) => item.code === journal.bookcode);
  const totals = journalTotals(journal.lines);
  const showDepartment = journal.lines.some((line) => line.departmentcode?.trim());
  const showProject = journal.lines.some((line) => line.projectcode?.trim());
  const vats = journal.details?.vats ?? [];
  const withholdings = journal.details?.withholdings ?? [];
  const mark = journal.status === "draft" ? tr("gl_print_draft_mark", "ฉบับร่าง — ยังไม่ผ่านรายการ") : journal.status === "reversed" ? tr("gl_reversed", "กลับรายการแล้ว") : "";
  const money = (value: string | undefined) => formatAmount(value ?? "", scale);
  const fields: [string, string][] = [
    [tr("gl_document_no", "เลขที่เอกสาร"), journal.docno],
    [tr("gl_date", "วันที่"), formatAppDate(journal.date, language)],
    [tr("gl_journal", "สมุดรายวัน"), journal.bookcode],
    [tr("gl_reference", "เอกสารอ้างอิง"), journal.reference],
    [tr("gl_branch_code", "รหัสสาขา"), journal.branchcode],
  ];
  return <article className="gl-print-voucher">
    <header className="gl-print-voucher-head">
      <div>
        <PrintHeading company={company} title={`${tr("gl_print_voucher_title", "ใบสำคัญรายวัน")}${book ? ` — ${journalBookName(book, journal.bookcode, language)}` : ""}`} meta={[]} />
        {mark && <div className="gl-print-mark">{mark}</div>}
      </div>
      <dl className="gl-print-fields">{fields.filter(([, value]) => value?.trim()).map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
    </header>
    <p className="gl-print-desc"><strong>{tr("gl_entry_description", "คำอธิบายรายการ")}:</strong> {journal.description || "—"}</p>
    <table>
      <thead><tr>
        <th>{tr("gl_account_code", "รหัสบัญชี")}</th><th>{tr("gl_account_name", "ชื่อบัญชี")}</th><th>{tr("gl_description", "คำอธิบาย")}</th>
        {showDepartment && <th>{tr("gl_department_code", "รหัสแผนก")}</th>}{showProject && <th>{tr("gl_project_code", "รหัสโครงการ")}</th>}
        <th className="gl-print-num">{tr("gl_debit", "เดบิต")}</th><th className="gl-print-num">{tr("gl_credit", "เครดิต")}</th>
      </tr></thead>
      <tbody>
        {journal.lines.map((line, index) => <tr key={index}>
          <td>{line.accountcode}</td><td>{line.accountname ?? ""}</td><td>{line.description}</td>
          {showDepartment && <td>{line.departmentcode}</td>}{showProject && <td>{line.projectcode}</td>}
          <td className="gl-print-num">{printAmountCell("debit", line.debit, scale)}</td><td className="gl-print-num">{printAmountCell("credit", line.credit, scale)}</td>
        </tr>)}
        <tr className="gl-print-total">
          <td colSpan={3 + (showDepartment ? 1 : 0) + (showProject ? 1 : 0)}>{tr("gl_total", "ยอดรวม")}</td>
          <td className="gl-print-num">{money(amountString(totals.debit))}</td><td className="gl-print-num">{money(amountString(totals.credit))}</td>
        </tr>
      </tbody>
    </table>
    {vats.length > 0 && <section className="gl-print-section">
      <h3>{tr("gl_print_vat_section", "รายละเอียดภาษีมูลค่าเพิ่ม")}</h3>
      <table>
        <thead><tr><th>{tr("gl_detail_vats_tax_type", "ประเภทภาษี")}</th><th>{tr("gl_detail_vats_tax_invoice_no", "เลขที่ใบกำกับภาษี")}</th><th>{tr("gl_detail_vats_tax_invoice_date", "วันที่ใบกำกับภาษี")}</th><th>{tr("gl_detail_vats_partner_name", "ชื่อผู้ซื้อ / ผู้ขาย")}</th><th className="gl-print-num">{tr("gl_detail_vats_base_amount", "ฐานภาษี (มูลค่าที่ต้องเสียภาษี)")}</th><th className="gl-print-num">{tr("gl_detail_vats_vat_rate", "อัตราภาษี (%)")}</th><th className="gl-print-num">{tr("gl_detail_vats_vat_amount", "ภาษีมูลค่าเพิ่ม")}</th></tr></thead>
        <tbody>{vats.map((vat) => { const label = vatTypeLabels[String(vat.tax_type)]; return <tr key={vat.id}><td>{label ? tr(label[0], label[1]) : String(vat.tax_type)}</td><td>{vat.tax_invoice_no}</td><td>{vat.tax_invoice_date ? formatAppDate(vat.tax_invoice_date, language) : ""}</td><td>{vat.partner_name}</td><td className="gl-print-num">{money(vat.base_amount)}</td><td className="gl-print-num">{formatAmount(vat.vat_rate ?? "", 0)}</td><td className="gl-print-num">{money(vat.vat_amount)}</td></tr>; })}</tbody>
      </table>
    </section>}
    {withholdings.length > 0 && <section className="gl-print-section">
      <h3>{tr("gl_details_section_withholdings", "ภาษีหัก ณ ที่จ่าย")}</h3>
      <table>
        <thead><tr><th>{tr("gl_detail_withholdings_form_type", "แบบยื่น")}</th><th>{tr("gl_detail_withholdings_wht_cert_no", "เลขที่หนังสือรับรอง")}</th><th>{tr("gl_detail_withholdings_payment_date", "วันที่จ่ายเงิน")}</th><th>{tr("gl_detail_withholdings_income_description", "รายละเอียดเงินได้")}</th><th className="gl-print-num">{tr("gl_detail_withholdings_base_amount", "ฐานภาษี (จำนวนเงินที่จ่าย)")}</th><th className="gl-print-num">{tr("gl_detail_withholdings_wht_rate", "อัตราภาษี (%)")}</th><th className="gl-print-num">{tr("gl_detail_withholdings_tax_amount", "ภาษีที่หัก")}</th></tr></thead>
        <tbody>{withholdings.map((wht) => <tr key={wht.id}><td>{whtFormLabel(wht.form_type, tr)}</td><td>{wht.wht_cert_no ?? ""}</td><td>{wht.payment_date ? formatAppDate(wht.payment_date, language) : ""}</td><td>{wht.income_description ?? ""}</td><td className="gl-print-num">{money(wht.base_amount)}</td><td className="gl-print-num">{formatAmount(wht.wht_rate ?? "", 0)}</td><td className="gl-print-num">{money(wht.tax_amount)}</td></tr>)}</tbody>
      </table>
    </section>}
    {journal.reason?.trim() && <p className="gl-print-desc"><strong>{tr("gl_reason", "เหตุผล")}:</strong> {journal.reason}</p>}
    <div className="gl-print-sign">
      {([["gl_print_prepared_by", "ผู้จัดทำ"], ["gl_print_checked_by", "ผู้ตรวจสอบ"], ["gl_print_approved_by", "ผู้อนุมัติ"]] as const).map(([key, fallback]) => <div key={key}>
        <div className="gl-print-sign-line" />
        <div>{tr(key, fallback)}</div>
        <div className="gl-print-sign-date">{tr("gl_print_sign_date", "วันที่ ........../........../..........")}</div>
      </div>)}
    </div>
  </article>;
}

/** วันที่หัวงบแบบเต็ม (เช่น 31 ธันวาคม 2569) ตามแบบงบการเงินของกรมพัฒนาธุรกิจการค้า — ภาษาไทยปี พ.ศ. */
function statementDate(value: string, language: LanguageCode): string {
  const [year, month, day] = value.split("-").map(Number);
  if (!year || !month || !day) return value;
  return new Intl.DateTimeFormat(localeForDate(language, language === "th" ? "buddhist" : "christian"), { day: "numeric", month: "long", year: "numeric", timeZone: "UTC" }).format(new Date(Date.UTC(year, month - 1, day, 12)));
}

/** บรรทัดวันที่ของหัวงบ (TFRS for NPAEs ย่อหน้า 4.7): งบ ณ วันใดวันหนึ่ง "ณ วันที่", งบช่วงเวลา "สำหรับปีสิ้นสุดวันที่" หรือ "สำหรับงวดตั้งแต่วันที่ … ถึงวันที่ …" */
export function statementPeriodText(period: { from: string; to: string } | undefined, pointInTime: boolean, fullYear: boolean, tr: GLTextFn, language: LanguageCode): string {
  if (!period?.to) return "";
  const to = statementDate(period.to, language);
  if (pointInTime) return tr("gl_statement_as_of", "ณ วันที่ {0}").replace("{0}", to);
  if (fullYear) return tr("gl_statement_year_ended", "สำหรับปีสิ้นสุดวันที่ {0}").replace("{0}", to);
  return tr("gl_statement_period_range", "สำหรับงวดตั้งแต่วันที่ {0} ถึงวันที่ {1}").replace("{0}", statementDate(period.from, language)).replace("{1}", to);
}

/** ยอดในงบ: backend ปัดตามทศนิยมของรูปแบบแล้ว; ศูนย์แสดง "-" เว้นแต่แถวตั้งให้แสดงศูนย์ */
export function statementAmountText(value: string, showZero: boolean, scale = 2): string {
  const text = formatAmount(value, scale);
  return text && !showZero && /^-?0(\.0+)?$/.test(text) ? "-" : text;
}

const statementUnderline: Record<string, string> = { single: "gl-statement-u-single", double: "gl-statement-u-double", top_single: "gl-statement-u-top", top_single_bottom_double: "gl-statement-u-top gl-statement-u-double" };

/** งบการเงินจากรูปแบบงบ (GET reports/statement) ใช้ทั้งพรีวิวบนจอและหน้าพิมพ์: หัวงบ = ชื่อกิจการ ชื่องบ วันที่/งวด หน่วยเงิน;
 *  คอลัมน์เงินตามที่ backend ส่ง (ปีนี้ + ปีก่อนเมื่อรูปแบบตั้งเปรียบเทียบ — TFRS for NPAEs ย่อหน้า 4.3) */
export function GLStatementTable({ report, company, title, period, showNote, scale = 2, tr }: { report: GLReport; company: string; title: string; period: string; showNote: boolean; scale?: number; tr: GLTextFn }) {
  const amounts = report.columns.filter((column) => column.amount);
  const span = 1 + (showNote ? 1 : 0) + amounts.length;
  return (
    <table className="gl-print-statement gl-statement">
      <thead>
        <tr className="gl-print-head">
          <th colSpan={span} className="gl-statement-heading">
            <div className="gl-print-company">{company}</div>
            <div className="gl-print-title">{title}</div>
            {period && <div className="gl-print-meta">{period}</div>}
            <div className="gl-print-meta">{tr("gl_statement_unit_baht", "(หน่วย : บาท)")}</div>
          </th>
        </tr>
        <tr className="gl-statement-columns">
          <th />
          {showNote && <th className="gl-statement-note">{tr("gl_note", "หมายเหตุ")}</th>}
          {amounts.map((column) => <th key={column.key} className="gl-print-num">{column.label}</th>)}
        </tr>
      </thead>
      <tbody>
        {(report.rows ?? []).map((row, index) => {
          if (row.rowtype === "blank") return <tr key={index}><td colSpan={span}>&nbsp;</td></tr>;
          if (row.rowtype === "divider") return <tr key={index}><td colSpan={span} className="gl-statement-divider" /></tr>;
          const font = `${row.fontweight === "bold" || row.fontweight === "semibold" ? "gl-statement-bold" : ""} ${row.fontstyle === "italic" ? "gl-statement-italic" : ""}`.trim();
          return (
            <tr key={index}>
              <td className={font} style={{ paddingLeft: `${Number(row.indent || 0) * 1.25 + 0.25}em` }}>{row.title}</td>
              {showNote && <td className="gl-statement-note">{row.noteno}</td>}
              {amounts.map((column) => (
                <td key={column.key} className="gl-print-num">
                  <span className={`${font} ${statementUnderline[row.underline ?? ""] ?? ""}`.trim()}>{statementAmountText(row[column.key] ?? "", row.showzero === "true", scale)}</span>
                </td>
              ))}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

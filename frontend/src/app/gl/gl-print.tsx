"use client";
// พิมพ์สมุดบัญชีและใบสำคัญรายวันตามประกาศกรมทะเบียนการค้า เรื่อง กำหนดชนิดของบัญชีที่ต้องจัดทำฯ พ.ศ. 2544
// ข้อ 4 (ชื่อผู้มีหน้าที่จัดทำบัญชี ชนิดของบัญชี ลำดับเล่ม), ข้อ 5 (ชื่อบัญชี วันที่ เลขที่เอกสาร รายการ จำนวนเงิน + เลขหน้าเรียงทุกหน้า),
// ข้อ 11 (เอกสารที่ทำขึ้นใช้เอง: คำอธิบายรายการ วิธีคำนวณ ลายมือชื่อผู้จัดทำบัญชีหรือผู้อนุมัติ) — ทะเบียนอ้างอิง docs/kms/21 §11
// งบการเงิน (GLStatementTable) ตามประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566 + TFRS for NPAEs บทที่ 4 — docs/kms/21
// หน้าพิมพ์เป็นแค่การแสดงผล: ตัวเลขทุกตัวมาจาก API ของ backend ตามเดิม
import { Fragment, useCallback, useEffect, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { AlertTriangle, CheckCircle2 } from "lucide-react";
import type { LanguageCode } from "@/lib/i18n";
import { formatAppDate, localeForDate } from "@/lib/date-time";
import { postBatchSignatories, type GLPostBatchInfo } from "@/lib/gl-journal-voucher";
import { DIMENSIONS, STATEMENT_CURRENT_EARNINGS, accountName, dimensionDef, rowsHoldDimension, accountTypeLabels, amountString, fillText, formatAmount, journalBookName, journalTotals, labelText, sortAccountsHierarchically, statementIsPeriodic, type GLAccount, type GLDimension, type GLDimensionResource, type GLFiscalYear, type GLJournal, type GLJournalBook, type GLReport, type GLStatementCheck, type GLStatementUnassigned as GLStatementUnassignedItem, type GLTextFn, type StatementNote } from "@/lib/general-ledger";
import { companyBaseName, workspaceCompanyDisplayName, workspaceStorageKeys, type WorkspaceSession } from "@/lib/workspace-models";
import { getAuthSession } from "@/lib/client-auth-session";
import { useGLText } from "./gl-common";

export type GLPrintOrientation = "portrait" | "landscape";

/** เลขหน้า "หน้า n / N" มุมขวาบนทุกหน้าแบบ Champ (Page No. : n/N) ด้วย @page margin box (Chrome 131+, Safari 18.2); ป้ายมาจาก languages.tsv จึงต้อง escape ให้เป็น CSS string */
export function printPageStyle(pageLabel: string): string {
  const label = JSON.stringify(pageLabel).replace(/</g, "\\3c ");
  return (["portrait", "landscape"] as const).map((orientation) =>
    `@page gl-print-${orientation} { size: A4 ${orientation}; margin: ${orientation === "portrait" ? "12mm 10mm 14mm" : "10mm 10mm 12mm"}; @top-right { content: ${label} " " counter(page) " / " counter(pages); font-size: 8pt; } }`,
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

/** ชื่อผู้พิมพ์ (Printed By ของ Champ): ชื่อในโปรไฟล์ หรือชื่อผู้ใช้ที่เข้าระบบ */
export function printUserName(): string {
  const session = getAuthSession();
  return session?.profile?.name?.trim() || session?.username?.trim() || "";
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

/** หัวรายงานแบบ Champ (ซ้ำทุกหน้าใน thead): ชื่อบริษัท, ชื่อรายงาน, พิมพ์โดย ↔ วันที่พิมพ์, รายละเอียด — เลขหน้าอยู่มุมขวาบนจาก @page */
export function ReportPrintHeading({ company, title, printedBy, printedAt, description, tr }: { company: string; title: string; printedBy: string; printedAt: string; description: string[]; tr: GLTextFn }) {
  const details = description.filter(Boolean);
  return <>
    <div className="gl-print-company">{company}</div>
    <div className="gl-print-title">{title}</div>
    <div className="gl-print-meta" style={{ display: "flex", justifyContent: "space-between", gap: "8mm" }}>
      <span>{tr("gl_print_printed_by", "พิมพ์โดย")} : {printedBy || "-"}</span>
      <span>{tr("gl_print_printed_date", "วันที่พิมพ์")} : {printedAt}</span>
    </div>
    {details.length > 0 && <div className="gl-print-meta">{tr("gl_print_description", "รายละเอียด")} : {details.join(" · ")}</div>}
  </>;
}

/** "เงื่อนไขการพิมพ์รายงาน" ท้ายรายงานแบบ Champ — ตัวกรองที่ใช้ (ช่วงบัญชี สาขา แผนก โครงการ) */
export function ReportPrintConditions({ lines, tr }: { lines: string[]; tr: GLTextFn }) {
  const items = lines.filter(Boolean);
  if (!items.length) return null;
  return <div className="gl-print-meta" style={{ marginTop: "4mm", breakInside: "avoid", pageBreakInside: "avoid", textAlign: "left" }}>
    <div style={{ fontWeight: 700 }}>{tr("gl_print_conditions", "เงื่อนไขการพิมพ์รายงาน")}</div>
    {items.map((line) => <div key={line}>{line}</div>)}
  </div>;
}

const vatTypeLabels: Record<string, [string, string]> = { "1": ["gl_vat_tax_type_purchase", "ภาษีซื้อ (ใบกำกับที่ได้รับ)"], "2": ["gl_vat_tax_type_sale", "ภาษีขาย (ใบกำกับที่ออก)"] };
function whtFormLabel(form: string, tr: GLTextFn) {
  return form === "PND53" ? tr("wht_cert_ui_form_53", "ภ.ง.ด.53 (นิติบุคคล)") : form === "PND3" ? tr("wht_cert_ui_form_3", "ภ.ง.ด.3 (บุคคลธรรมดา)") : form === "PND2" ? tr("wht_cert_ui_form_2", "ภ.ง.ด.2 (ดอกเบี้ย/เงินปันผล)") : form;
}

/** ใบสำคัญรายวัน (ข้อ 9 + ข้อ 11): ผู้จัดทำ ชื่อเอกสาร เลขที่ วันที่ ยอดรวม คำอธิบาย วิธีคำนวณภาษี และช่องลายมือชื่อ */
export function GLVoucherPrint({ journal, books, company, tr, language, scale = 2, postBatch }: { journal: GLJournal; books: GLJournalBook[]; company: string; tr: GLTextFn; language: LanguageCode; scale?: number; postBatch?: GLPostBatchInfo | null }) {
  const book = books.find((item) => item.code === journal.bookcode);
  const totals = journalTotals(journal.lines);
  // a dimension column is printed only when some line holds a code of it (all six, in Champ's journal order)
  const dimensions = DIMENSIONS.filter((def) => rowsHoldDimension(journal.lines, def));
  const vats = journal.details?.vats ?? [];
  const withholdings = journal.details?.withholdings ?? [];
  const signatories = postBatchSignatories(postBatch);
  const signNames: Record<string, string> = { gl_print_checked_by: signatories.checkedBy, gl_print_approved_by: signatories.approvedBy };
  const mark = journal.status === "draft" ? tr("gl_print_draft_mark", "ฉบับร่าง — ยังไม่ผ่านรายการ") : journal.status === "reversed" ? tr("gl_reversed", "กลับรายการแล้ว") : "";
  const money = (value: string | undefined) => formatAmount(value ?? "", scale);
  const fields: [string, string][] = [
    [tr("gl_document_no", "เลขที่เอกสาร"), journal.docno],
    [tr("gl_date", "วันที่"), formatAppDate(journal.date, language)],
    [tr("gl_journal", "สมุดรายวัน"), journal.bookcode],
    [tr("gl_reference_document", "เอกสารอ้างอิง"), journal.reference],
    [tr("gl_ref_date", "วันที่อ้างอิง"), journal.refdate ? formatAppDate(journal.refdate, language) : ""],
    [tr("gl_branch_code", "รหัสสาขา"), journal.branchcode],
  ];
  return <article className="gl-print-voucher">
    <header className="gl-print-voucher-head">
      <div>
        <PrintHeading company={company} title={book?.titlename?.trim() || `${tr("gl_print_voucher_title", "ใบสำคัญรายวัน")}${book ? ` — ${journalBookName(book, journal.bookcode, language)}` : ""}`} meta={[]} />
        {mark && <div className="gl-print-mark">{mark}</div>}
      </div>
      <dl className="gl-print-fields">{fields.filter(([, value]) => value?.trim()).map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
    </header>
    <p className="gl-print-desc"><strong>{tr("gl_description", "คำอธิบาย")}:</strong> {journal.description || "—"}</p>
    <table>
      <thead><tr>
        <th>{tr("gl_account_code", "รหัสบัญชี")}</th><th>{tr("gl_account_name", "ชื่อบัญชี")}</th><th>{tr("gl_description", "คำอธิบาย")}</th>
        {dimensions.map((def) => <th key={def.line}>{tr(...def.codeLabel)}</th>)}
        <th className="gl-print-num">{tr("gl_debit", "เดบิต")}</th><th className="gl-print-num">{tr("gl_credit", "เครดิต")}</th>
      </tr></thead>
      <tbody>
        {journal.lines.map((line, index) => <tr key={index}>
          <td>{line.accountcode}</td><td>{line.accountname ?? ""}</td><td>{line.description}</td>
          {dimensions.map((def) => <td key={def.line}>{line[def.line] ?? ""}</td>)}
          <td className="gl-print-num">{printAmountCell("debit", line.debit, scale)}</td><td className="gl-print-num">{printAmountCell("credit", line.credit, scale)}</td>
        </tr>)}
        <tr className="gl-print-total">
          <td colSpan={3 + dimensions.length}>{tr("gl_total", "ยอดรวม")}</td>
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
        {signNames[key] && <div>({signNames[key]})</div>}
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

/** บรรทัดวันที่ของงบที่ backend คำนวณแล้ว: งวดแรกของรายงาน (ปีนี้) เทียบช่วงของปีบัญชี — ทั้งปี = "สำหรับปีสิ้นสุดวันที่",
 *  ช่วงย่อย = "สำหรับงวดตั้งแต่ … ถึง …", งบ ณ วันที่ = "ณ วันที่" (กติกาเดียวกับ backend); ใช้ทั้งพรีวิวในจอออกแบบและชุดงบการเงิน */
export function statementPeriodLine(statementtype: string, report: GLReport, years: GLFiscalYear[], tr: GLTextFn, language: LanguageCode): string {
  const current = report.periods?.[0];
  const year = years.find((item) => item.code === current?.fiscalyear);
  const fullYear = Boolean(year && current && current.from === year.startdate && current.to === year.enddate);
  return statementPeriodText(current, !statementIsPeriodic(statementtype), fullYear, tr, language);
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
    <table className={`gl-print-statement gl-statement${amounts.length > 3 ? " gl-statement-wide" : ""}`}>
      <thead>
        <tr className="gl-print-head">
          <th colSpan={span} className="gl-statement-heading">
            <div className="gl-print-company">{company}</div>
            <div className="gl-print-title">{title}</div>
            {period && <div className="gl-print-meta">{period}</div>}
            <div className="gl-print-meta">{tr("gl_statement_unit_baht", "(หน่วย: บาท)")}</div>
          </th>
        </tr>
        <tr className="gl-statement-columns">
          <th />
          {showNote && <th className="gl-statement-note">{tr("gl_note", "หมายเหตุ")}</th>}
          {amounts.map((column) => <th key={column.key} className="gl-print-num">{column.key === "total" ? tr("gl_statement_equity_total", column.label) : column.label}</th>)}
        </tr>
      </thead>
      <tbody>
        {(report.rows ?? []).map((row, index, rows) => {
          // งบส่วนของผู้ถือหุ้นมีชุดแถวต่อปี (ปีก่อน → ปีนี้): เว้นบรรทัดคั่นระหว่างชุด
          const spacer = index > 0 && row.block && row.block !== rows[index - 1].block ? <tr key={`gap-${index}`}><td colSpan={span}>&nbsp;</td></tr> : null;
          if (row.rowtype === "blank") return <tr key={index}><td colSpan={span}>&nbsp;</td></tr>;
          if (row.rowtype === "divider") return <tr key={index}><td colSpan={span} className="gl-statement-divider" /></tr>;
          const font = `${row.fontweight === "bold" || row.fontweight === "semibold" ? "gl-statement-bold" : ""} ${row.fontstyle === "italic" ? "gl-statement-italic" : ""}`.trim();
          return [spacer, (
            <tr key={index}>
              <td className={font} style={{ paddingLeft: `${Number(row.indent || 0) * 1.25 + 0.25}em` }}>{row.title}</td>
              {showNote && <td className="gl-statement-note">{row.noteno}</td>}
              {amounts.map((column) => (
                <td key={column.key} className="gl-print-num">
                  <span className={`${font} ${statementUnderline[row.underline ?? ""] ?? ""}`.trim()}>{statementAmountText(row[column.key] ?? "", row.showzero === "true", scale)}</span>
                </td>
              ))}
            </tr>
          )];
        })}
      </tbody>
    </table>
  );
}

/** หมายเหตุประกอบงบการเงิน (แบบ 2 ข้อ 5 ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566) พิมพ์ต่อจากงบ: หัว = ชื่อกิจการ, "หมายเหตุประกอบงบการเงิน",
 *  บรรทัดงวด (สำหรับปีสิ้นสุดวันที่ …); แต่ละข้อ "เลขที่. หัวข้อ" ตัวหนา (ไม่ขึ้นหน้าใหม่ทันทีหลังหัวข้อ) ตามด้วยเนื้อหาที่คงการขึ้นบรรทัดของผู้ใช้.
 *  ไม่กำหนดสี — ใช้สีตัวอักษรของหน้าพิมพ์ (currentColor) */
export function GLNotesPrint({ notes, company, period, tr }: { notes: StatementNote[]; company: string; period: string; tr: GLTextFn }) {
  return (
    <div className="gl-print-notes">
      <div style={{ textAlign: "center", marginBottom: "4mm" }}>
        <div className="gl-print-company">{company}</div>
        <div className="gl-print-title">{tr("gl_statement_notes_title", "หมายเหตุประกอบงบการเงิน")}</div>
        {period && <div className="gl-print-meta">{period}</div>}
      </div>
      {notes.map((note, index) => (
        <section key={note.id || index} style={{ marginTop: "3mm" }}>
          <div className="gl-print-note-heading" style={{ fontWeight: 700, breakAfter: "avoid", pageBreakAfter: "avoid", breakInside: "avoid" }}>{`${(note.noteno ?? "").trim()}. ${(note.title ?? "").trim()}`}</div>
          {(note.body ?? "") !== "" && <div className="gl-print-note-body" style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere", marginTop: "1mm" }}>{note.body}</div>}
        </section>
      ))}
    </div>
  );
}

/** งบที่มีคอลัมน์ยอดเงินมากกว่า 3 คอลัมน์ (งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น) พิมพ์แนวนอน */
export function statementOrientation(report: GLReport): GLPrintOrientation {
  return report.columns.filter((column) => column.amount).length > 3 ? "landscape" : "portrait";
}

/** งบหนึ่งรายการในชุดงบการเงิน: ผลคำนวณจาก backend + หัวงบ (ชื่อรูปแบบ, บรรทัดวันที่) และรูปแบบแสดงผลของรูปแบบงบ */
export type GLStatementSetSection = { code: string; title: string; period: string; report: GLReport; showNote: boolean; scale: number };

/** แนวกระดาษของ .gl-print-root = แนวของส่วนแรกที่พิมพ์ (หน้าแรกได้ชื่อ @page ถูกต้อง); มีแต่หมายเหตุ = แนวตั้ง */
export function statementSetRootOrientation(sections: { report: GLReport }[]): GLPrintOrientation {
  return sections.length ? statementOrientation(sections[0].report) : "portrait";
}

/** ชุดงบการเงินในงานพิมพ์เดียว: งบตามลำดับที่ส่งมา (แบบ 2 — backend เรียงให้ใน reports/statement-set) แล้วหมายเหตุประกอบงบการเงินท้ายสุด;
 *  แต่ละส่วนขึ้นหน้าใหม่และใช้แนวกระดาษของตัวเอง (globals.css "GL statement set print") — ตัวเลขทุกตัวมาจาก backend */
export function GLStatementSetPrint({ sections, notes, company, notesPeriod, tr }: { sections: GLStatementSetSection[]; notes: StatementNote[]; company: string; notesPeriod: string; tr: GLTextFn }) {
  return <>
    {sections.map((section) => (
      <section key={section.code} className="gl-print-set-section" data-orientation={statementOrientation(section.report)}>
        <GLStatementTable report={section.report} company={company} title={section.title} period={section.period} showNote={section.showNote} scale={section.scale} tr={tr} />
      </section>
    ))}
    {notes.length > 0 && (
      <section className="gl-print-set-section" data-orientation="portrait">
        <GLNotesPrint notes={notes} company={company} period={notesPeriod} tr={tr} />
      </section>
    )}
  </>;
}

/** ผลตรวจยอดของงบกับบัญชี (งบกระแสเงินสด: เงินสดปลายงวดตามงบเทียบยอดคงเหลือตามบัญชี คำนวณที่ backend) — แสดงบนจอเท่านั้น
 *  ไม่อยู่ใน GLStatementTable จึงไม่ถูกพิมพ์; สถานะบอกด้วยไอคอน + ข้อความเสมอ ไม่ใช้สีอย่างเดียว */
/** key ของผลตรวจ: เลขบรรทัดของรูปแบบงบแก้ได้และซ้ำได้ จึงต่อท้ายลำดับในรายการ ไม่ให้ React ทิ้ง/ซ้ำการ์ด "ไม่ตรงกัน" */
export function statementCheckKey(check: GLStatementCheck, index: number) {
  return `${check.key ?? ""}-${check.rowno ?? ""}-${index}`;
}
export function GLStatementChecks({ checks, scale = 2, tr }: { checks?: GLStatementCheck[] | null; scale?: number; tr: GLTextFn }) {
  const items = checks ?? [];
  if (!items.length) return null;
  const title = tr("gl_statement_check_title", "ตรวจยอดกับบัญชี");
  return (
    <ul className="grid gap-2 lg:grid-cols-2" aria-label={title}>
      {items.map((check, index) => (
        <li key={statementCheckKey(check, index)} className={`rounded-xl border p-3 text-[0.95rem] leading-relaxed text-foreground shadow-[0_2px_8px_rgba(0,0,0,0.08)] ${check.matched ? "border-border bg-muted/40" : "border-primary/40 bg-primary/10"}`}>
          <div className="font-semibold">
            {title}: {check.title?.trim() || tr("gl_statement_check_row", "บรรทัด {0}").replace("{0}", String(check.rowno ?? ""))} · {tr("gl_fiscal_year", "ปีบัญชี")} {check.fiscalyear ?? ""}
          </div>
          <div className="mt-1 tabular-nums">
            {tr("gl_statement_check_statement", "ตามงบ")} {formatAmount(check.statement ?? "", scale)} · {tr("gl_statement_check_book", "ตามบัญชี")} {formatAmount(check.book ?? "", scale)}
          </div>
          {check.matched ? (
            <div className="mt-1.5 inline-flex items-center gap-1.5 font-semibold text-primary">
              <CheckCircle2 aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_check_matched", "ตรงกัน")}
            </div>
          ) : (
            <div className="mt-1.5 inline-flex items-center gap-1.5 font-semibold tabular-nums text-primary">
              <AlertTriangle aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_check_mismatched", "ไม่ตรงกัน")} {tr("gl_difference", "ผลต่าง")} {formatAmount(check.difference ?? "", scale)}
            </div>
          )}
        </li>
      ))}
    </ul>
  );
}

/** คำเตือนของรายงานงบจาก backend: กล่องสีหลัก (primary) มีไอคอนและหัวข้อ (ไม่ใช้สีอย่างเดียว) — แสดงบนจอเท่านั้น ไม่อยู่ในหน้าพิมพ์ */
export function GLReportWarnings({ warnings, tr }: { warnings?: string[] | null; tr: GLTextFn }) {
  const items = (warnings ?? []).filter((warning) => warning?.trim());
  if (!items.length) return null;
  return (
    <div role="status" className="rounded-xl border border-primary/40 bg-primary/10 p-3 text-[0.95rem] leading-relaxed text-foreground shadow-[0_2px_8px_rgba(0,0,0,0.08)]">
      <div className="flex items-center gap-1.5 font-semibold text-primary">
        <AlertTriangle aria-hidden className="size-4 shrink-0" /> {tr("gl_statement_warnings_title", "ข้อควรตรวจสอบก่อนออกงบ")}
      </div>
      <ul className="mt-1.5 list-disc space-y-1 pl-6 [overflow-wrap:anywhere]">
        {items.map((warning, index) => <li key={index}>{warning}</li>)}
      </ul>
    </div>
  );
}

/** บัญชีที่มียอดแต่ยังไม่อยู่ในบรรทัดใดของงบ (backend statements_unassigned.go — งบฐานะการเงิน/งบกำไรขาดทุน): ยอดรวมของงบไม่ครบ
 *  จนกว่าผู้ใช้เพิ่มบัญชีเข้าบรรทัดในแท็บออกแบบ. กล่องสีหลัก (primary) + ไอคอน + หัวข้อ (ไม่ใช้สีอย่างเดียว) อยู่นอก GLStatementTable จึงไม่ถูกพิมพ์ */
export const STATEMENT_UNASSIGNED_SHOWN = 30;
export function GLStatementUnassigned({ items, scale = 2, tr }: { items?: GLStatementUnassignedItem[] | null; scale?: number; tr: GLTextFn }) {
  const list = items ?? [];
  if (!list.length) return null;
  const title = tr("gl_statement_unassigned_title", "บัญชีที่มียอดแต่ยังไม่อยู่ในบรรทัดใดของงบนี้");
  const description = list[0].basis === "movement"
    ? tr("gl_statement_unassigned_desc_movement", "ยอดเคลื่อนไหวในงวดของบัญชีเหล่านี้ไม่ถูกนับในงบ รายได้ ค่าใช้จ่าย และกำไร (ขาดทุน) สุทธิจึงอาจไม่ตรงกับบัญชี — เพิ่มบัญชีเข้าบรรทัดที่ถูกต้องในแท็บออกแบบ แล้วบันทึกแม่แบบ")
    : tr("gl_statement_unassigned_desc_balance", "ยอดคงเหลือ ณ วันสิ้นงวดของบัญชีเหล่านี้ไม่ถูกนับในงบ ยอดรวมจึงไม่ครบ — เพิ่มบัญชีเข้าบรรทัดที่ถูกต้องในแท็บออกแบบ แล้วบันทึกแม่แบบ");
  return (
    <div role="status" aria-label={title} className="rounded-xl border border-primary/40 bg-primary/10 p-3 text-[0.95rem] leading-relaxed text-foreground shadow-[0_2px_8px_rgba(0,0,0,0.08)]">
      <div className="flex items-center gap-1.5 font-semibold text-primary">
        <AlertTriangle aria-hidden className="size-4 shrink-0" /> {title}
      </div>
      <p className="mt-1 [overflow-wrap:anywhere]">{description}</p>
      <ul className="mt-1.5 list-disc space-y-1 pl-6 tabular-nums [overflow-wrap:anywhere]">
        {list.slice(0, STATEMENT_UNASSIGNED_SHOWN).map((item, index) => (
          <li key={`${item.key ?? ""}-${item.fiscalyear ?? ""}-${item.accountcode ?? ""}-${index}`}>
            {item.accountcode === STATEMENT_CURRENT_EARNINGS
              ? fillText(tr("gl_statement_unassigned_current_earnings", "ปีบัญชี {0} · กำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี {1} ยังไม่อยู่ในบรรทัดใด — กด “ใช้แม่แบบมาตรฐาน” แล้วเลือกงบแสดงฐานะการเงิน (บรรทัด “ยังไม่ได้จัดสรร” รวมยอดนี้ให้) หรือทำ “ประมวลผลสิ้นปี” เพื่อโอนเข้ากำไรสะสม"), item.fiscalyear ?? "", formatAmount(item.amount ?? "", scale))
              : fillText(tr("gl_statement_unassigned_line", "ปีบัญชี {0} · {1} {2} · {3} · {4}"), item.fiscalyear ?? "", item.accountcode ?? "", item.accountname ?? "", labelText(accountTypeLabels, item.accounttype ?? "", tr), formatAmount(item.amount ?? "", scale))}
          </li>
        ))}
      </ul>
      {list.length > STATEMENT_UNASSIGNED_SHOWN && (
        <p className="mt-1">{fillText(tr("gl_statement_unassigned_more", "และอีก {0} บัญชี"), list.length - STATEMENT_UNASSIGNED_SHOWN)}</p>
      )}
    </div>
  );
}

/** รายงานรายละเอียดสมุดบัญชีแบบ Champ 5531 (GLRepGLBookView): รหัสสมุด · ชื่อสมุดรายวัน · ชื่อ/รายละเอียด (หัวเอกสาร) เรียงตามรหัส */
export function GLJournalBooksPrint({ books, company, tr, language, conditions = [] }: { books: GLJournalBook[]; company: string; tr: GLTextFn; language: LanguageCode; conditions?: string[] }) {
  const rows = books.filter((book) => !book.isdeleted).sort((a, b) => a.code.localeCompare(b.code));
  return (
    <article className="gl-print-chart-of-accounts">
      <table>
        <thead>
          <tr className="gl-print-head">
            <th colSpan={3}>
              <ReportPrintHeading company={company} title={tr("gl_book_report_title", "รายงานรายละเอียดสมุดบัญชี")} printedBy={printUserName()} printedAt={printedAtText(language)} description={[]} tr={tr} />
            </th>
          </tr>
          <tr>
            <th style={{ width: "18%" }}>{tr("gl_book_col_code", "รหัสสมุด")}</th>
            <th style={{ width: "42%" }}>{tr("gl_book_name", "ชื่อสมุดรายวัน")}</th>
            <th style={{ width: "40%" }}>{tr("gl_book_col_detail", "ชื่อ/รายละเอียด")}</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((book) => (
            <tr key={book.code}>
              <td className="font-mono">{book.code}</td>
              <td>{journalBookName(book, book.code, language)}</td>
              <td>{book.titlename ?? ""}</td>
            </tr>
          ))}
          <tr className="gl-print-total"><td colSpan={3}>{fillText(tr("gl_report_doc_total", "รวมทั้งสิ้น {0} รายการ"), rows.length.toLocaleString("th-TH"))}</td></tr>
        </tbody>
      </table>
      <ReportPrintConditions lines={conditions} tr={tr} />
    </article>
  );
}

/** Register listing of the six analysis dimensions like Champ 5538 / 5521 / 5524 / 5526 / 5527 / 5528 (GLRepDeaprtmentView, GLRepProjectView …): code · name/description, rows in the order
 *  the backend returned them (ORDER BY code, already filtered by the code range), total-count line, the range as the print condition. Title and code header come from the dimension definition. */
export function GLDimensionsPrint({ kind, rows, conditions, company, tr, language }: { kind: GLDimensionResource; rows: GLDimension[]; conditions: string[]; company: string; tr: GLTextFn; language: LanguageCode }) {
  const def = dimensionDef(kind);
  return (
    <article className="gl-print-chart-of-accounts">
      <table>
        <thead>
          <tr className="gl-print-head">
            <th colSpan={2}>
              <ReportPrintHeading company={company} title={tr(...def.reportTitle)} printedBy={printUserName()} printedAt={printedAtText(language)} description={[]} tr={tr} />
            </th>
          </tr>
          <tr>
            <th style={{ width: "20%" }}>{tr(...def.codeLabel)}</th>
            <th style={{ width: "80%" }}>{tr("gl_name_description", "ชื่อ / รายละเอียด")}</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id ?? row.code}>
              <td className="font-mono">{row.code}</td>
              <td>{row.name}</td>
            </tr>
          ))}
          <tr className="gl-print-total"><td colSpan={2}>{fillText(tr("gl_report_doc_total", "รวมทั้งสิ้น {0} รายการ"), rows.length.toLocaleString("th-TH"))}</td></tr>
        </tbody>
      </table>
      <ReportPrintConditions lines={conditions} tr={tr} />
    </article>
  );
}

const BUDGET_PRINT_DIMENSIONS: readonly string[] = ["branchcode", ...DIMENSIONS.map((def) => def.line)];
/** Dimension columns of the budget detail report (Champ 5525): branch + the six analysis dimensions, shown only when some budget row uses them. */
function budgetPrintDimensions(rows: Record<string, string>[]) {
  return BUDGET_PRINT_DIMENSIONS.filter((key) => rows.some((row) => (row[key] ?? "").trim() !== ""));
}
/** Column header of a budget dimension column: the dimension's register name, "สาขา" for the branch. */
function budgetPrintDimensionLabel(key: string, tr: GLTextFn): string {
  const def = DIMENSIONS.find((item) => item.line === key);
  return def ? tr(...def.label) : tr("gl_budget_col_branch", "สาขา");
}
/** Code, name, start, end, amount + the dimensions in use; Champ prints landscape from 3 dimensions, which printOrientation gives at 8 columns. */
export function budgetPrintColumnCount(rows: Record<string, string>[]) {
  return 5 + budgetPrintDimensions(rows).length;
}

/** Budget detail report like Champ 5525 (GLRepBudgetView), rendering the rows of the backend report "budgetdetail" as they come:
 *  budgetaccount heading → budgetline rows → budgettotal (linecount, amount); report totals = grand total. */
export function GLBudgetsPrint({ report, company, tr, language, conditions = [] }: { report: Pick<GLReport, "rows" | "totals">; company: string; tr: GLTextFn; language: LanguageCode; conditions?: string[] }) {
  const rows = report.rows ?? [];
  const dimensions = budgetPrintDimensions(rows);
  const columnCount = budgetPrintColumnCount(rows);
  const date = (value: string) => (value ? formatAppDate(value, language) : "");
  return (
    <article className="gl-print-chart-of-accounts">
      <table>
        <thead>
          <tr className="gl-print-head">
            <th colSpan={columnCount}>
              <ReportPrintHeading company={company} title={tr("gl_budget_report_title", "รายงานรายละเอียดงบประมาณ")} printedBy={printUserName()} printedAt={printedAtText(language)} description={[]} tr={tr} />
            </th>
          </tr>
          <tr>
            <th>{tr("gl_coa_col_code", "รหัสผังบัญชี")}</th>
            <th colSpan={columnCount - 1}>{tr("gl_coa_col_name", "ชื่อผังบัญชี")}</th>
          </tr>
          <tr>
            <th>{tr("gl_code", "รหัส")}</th>
            <th>{tr("gl_budget_list_name", "ชื่อ/รายละเอียด")}</th>
            {dimensions.map((key) => <th key={key}>{budgetPrintDimensionLabel(key, tr)}</th>)}
            <th>{tr("gl_budget_col_start", "วันที่เริ่ม")}</th>
            <th>{tr("gl_budget_col_end", "วันที่สิ้นสุด")}</th>
            <th className="gl-print-num">{tr("gl_amount", "จำนวนเงิน")}</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => {
            const key = `${row.rowtype}-${row.accountcode}-${row.budgetcode}`;
            if (row.rowtype === "budgetaccount") {
              return (
                <tr key={key}>
                  <td className="font-mono font-semibold">{row.accountcode}</td>
                  <td colSpan={columnCount - 1} className="font-semibold">{row.accountname}</td>
                </tr>
              );
            }
            if (row.rowtype === "budgettotal") {
              return (
                <tr key={key} className="gl-print-total">
                  <td />
                  <td colSpan={columnCount - 2}>{fillText(tr("gl_budget_account_total", "รวม {0} รายการ"), row.linecount ?? "")}</td>
                  <td className="gl-print-num">{formatAmount(row.amount ?? "")}</td>
                </tr>
              );
            }
            return (
              <tr key={key}>
                <td className="font-mono">{row.budgetcode}</td>
                <td>{row.budgetname}</td>
                {dimensions.map((dimension) => <td key={dimension}>{row[dimension]}</td>)}
                <td>{date(row.startdate ?? "")}</td>
                <td>{date(row.enddate ?? "")}</td>
                <td className="gl-print-num">{formatAmount(row.amount ?? "")}</td>
              </tr>
            );
          })}
          {rows.length > 0 && (
            <tr className="gl-print-total">
              <td />
              <td colSpan={columnCount - 2}>{fillText(tr("gl_report_doc_total", "รวมทั้งสิ้น {0} รายการ"), report.totals?.linecount ?? "0")}</td>
              <td className="gl-print-num">{formatAmount(report.totals?.amount ?? "0")}</td>
            </tr>
          )}
          {!rows.length && <tr><td colSpan={columnCount} style={{ textAlign: "center", padding: "2rem" }}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}
        </tbody>
      </table>
      <ReportPrintConditions lines={conditions} tr={tr} />
    </article>
  );
}

/** รายงานรายละเอียดผังบัญชีแบบ Champ 5520 (GLRepChartOfAccountView): หมวด · รหัสผังบัญชี · ชื่อผังบัญชี (ไทย / อังกฤษ) · ระดับ
 *  เรียงตามต้นไม้ในแต่ละหมวด, ชื่อหมวดพิมพ์แถวแรกของหมวด, "รวม<หมวด> n รายการ" ท้ายหมวด แล้วจำนวนบัญชีคุม/บัญชีย่อยท้ายรายงาน */
/** "แสดงชื่อบัญชี" ของ Champ 5520: ทั้งสอง / ชื่อบัญชี 1 (ไทย) / ชื่อบัญชี 2 (อังกฤษ) */
export type ChartNameMode = "both" | "name1" | "name2";
/** ชื่อที่พิมพ์ตามตัวเลือก; ชื่อ 2 ที่ไม่มีใช้ชื่อไทยแทน ไม่ให้แถวชื่อว่าง */
export function chartPrintName(thai: string, english: string | undefined, mode: ChartNameMode): string {
  const second = english?.trim();
  if (mode === "name1") return thai;
  if (mode === "name2") return second || thai;
  return second && second !== thai ? `${thai} / ${second}` : thai;
}

export function GLChartOfAccountsPrint({
  accounts,
  company,
  tr,
  language,
  conditions = [],
  nameMode = "both",
}: {
  accounts: GLAccount[];
  company: string;
  tr: GLTextFn;
  language: LanguageCode;
  conditions?: string[];
  nameMode?: ChartNameMode;
}) {
  const sorted = sortAccountsHierarchically(accounts);
  const codes = new Set(sorted.map((acc) => acc.accountcode));
  // หมวดของแถว = หมวดของบัญชีรากในต้นไม้ (sortAccountsHierarchically จัดรากตามหมวดอยู่แล้ว ลูกตามพ่อ)
  const rows: { acc: (typeof sorted)[number]; category: string }[] = [];
  let current = "";
  for (const acc of sorted) {
    const parent = acc.parentaccountcode?.trim();
    if (!parent || !codes.has(parent) || parent === acc.accountcode) current = acc.accounttype;
    rows.push({ acc, category: current });
  }
  const categoryName = (category: string) => labelText(accountTypeLabels, category, tr);
  const controlCount = sorted.filter((acc) => !acc.allowposting).length;
  const count = (value: number) => value.toLocaleString("th-TH");
  const categoryCounts = new Map<string, number>();
  for (const row of rows) categoryCounts.set(row.category, (categoryCounts.get(row.category) ?? 0) + 1);

  return (
    <article className="gl-print-chart-of-accounts">
      <table>
        <thead>
          <tr className="gl-print-head">
            <th colSpan={4}>
              <ReportPrintHeading
                company={company}
                title={tr("gl_coa_report_title", "รายงานรายละเอียดผังบัญชี")}
                printedBy={printUserName()}
                printedAt={printedAtText(language)}
                description={[]}
                tr={tr}
              />
            </th>
          </tr>
          <tr>
            <th style={{ width: "16%" }}>{tr("gl_coa_col_category", "หมวด")}</th>
            <th style={{ width: "18%" }}>{tr("gl_coa_col_code", "รหัสผังบัญชี")}</th>
            <th style={{ width: "56%" }}>{tr("gl_coa_col_name", "ชื่อผังบัญชี")}</th>
            <th style={{ width: "10%", textAlign: "center" }}>{tr("gl_level", "ระดับ")}</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(({ acc, category }, index) => {
            const lvl = acc.effectiveLevel ?? acc.level ?? 1;
            const english = acc.names?.find((name) => name.code === "en")?.name?.trim();
            const thai = accountName(acc);
            const firstOfCategory = rows[index - 1]?.category !== category;
            const lastOfCategory = rows[index + 1]?.category !== category;
            return (
              <Fragment key={acc.accountcode}>
                <tr>
                  <td>{firstOfCategory ? categoryName(category) : ""}</td>
                  <td className="font-mono">{acc.accountcode}</td>
                  <td style={{ paddingLeft: `${(lvl > 1 ? (lvl - 1) * 1.25 : 0) + 0.5}em` }} className={acc.allowposting ? "" : "font-semibold"}>
                    {chartPrintName(thai, english, nameMode)}
                  </td>
                  <td style={{ textAlign: "center" }}>{lvl}</td>
                </tr>
                {lastOfCategory && (
                  <tr className="gl-print-total">
                    <td colSpan={4}>{fillText(tr("gl_coa_category_total", "รวม{0} {1} รายการ"), categoryName(category), count(categoryCounts.get(category) ?? 0))}</td>
                  </tr>
                )}
              </Fragment>
            );
          })}
          {sorted.length > 0 && <>
            <tr className="gl-print-total"><td colSpan={4}>{fillText(tr("gl_coa_control_total", "รวมเป็นบัญชีคุม {1} รายการ"), "", count(controlCount))}</td></tr>
            <tr className="gl-print-total"><td colSpan={4}>{fillText(tr("gl_coa_posting_total", "รวมเป็นบัญชีย่อย {1} รายการ"), "", count(sorted.length - controlCount))}</td></tr>
          </>}
          {!sorted.length && (
            <tr>
              <td colSpan={4} style={{ textAlign: "center", padding: "2rem" }}>
                {tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}
              </td>
            </tr>
          )}
        </tbody>
      </table>
      <ReportPrintConditions lines={conditions} tr={tr} />
    </article>
  );
}

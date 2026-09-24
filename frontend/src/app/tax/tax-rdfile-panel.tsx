"use client";

import { useState } from "react";
import { AlertTriangle, CheckCircle2, FileDown, Info, Loader2, X } from "lucide-react";
import { useBackendText, type BackendTextFn } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import {
  isHeadOfficeBranch, RD_HEAD_OFFICE_DEPT_NAME,
  type TaxRdFileBranchType, type TaxRdFileError, type TaxRdFileIssue, type TaxRdFileLto, type TaxRdFileOptions, type TaxRdFileResult,
} from "@/lib/tax-forms";
import { cn } from "@/lib/utils";
import { focusTaxFormField } from "./tax-form-fields";

// ไฟล์ยื่นกรมสรรพากร (รูปแบบข้อมูล Format กลาง V2.0) สำหรับโปรแกรม SWC-UI — ปุ่มในแถบปุ่มของจอแบบยื่น + แผงตั้งค่าในหน้า
// ไฟล์สร้างจากฉบับที่บันทึกแล้วเท่านั้น (backend อ่านจาก tax_filings ด้วย id + version) จึงปิดปุ่มเมื่อยังไม่บันทึกหรือมีค่าที่แก้ค้าง

export const TAX_RDFILE_PANEL_ID = "tax-rdfile-panel";
const NEED_SAVE_ID = "tax-rdfile-need-save";
// จำนวนจุดที่ต้องแก้ที่แสดง (ที่เหลือบอกเป็นจำนวน) — รายการยาวเกินไปคนอ่านไม่ไหวและดันแบบลงไป
export const RDFILE_ISSUE_ROWS = 50;
// ช่องที่อยู่ในแผงนี้ (ไม่ใช่ช่องบนแบบ): field จาก backend หรือ key ของจุดที่ต้องแก้ → id ของช่องในแผง
const PANEL_FIELD_IDS: Record<string, string> = {
  dept_name: "tax-rdfile-dept_name",
  submission_no: "tax-rdfile-submission_no",
  lto: "tax-rdfile-lto",
  branch_type: "tax-rdfile-branch_type",
};
const PANEL_ISSUE_FIELDS: Record<string, string> = { tax_rdfile_dept_name_required: "dept_name", tax_rdfile_submission_invalid: "submission_no" };

// fill - แทน {name} ด้วยค่า (ใช้ callback: ค่าอย่าง "$&" ต้องไม่ถูกตีความเป็นรูปแบบพิเศษของ String.replace)
const fill = (template: string, vars: Record<string, string>) => template.replace(/\{(\w+)\}/g, (match, name: string) => vars[name] ?? match);

// panelField - จุดที่ต้องแก้นี้อยู่ที่ช่องในแผง (ชื่อแผนก/ครั้งที่ส่ง) หรือไม่
const panelField = (issue: TaxRdFileIssue) => (issue.field && PANEL_FIELD_IDS[issue.field] ? issue.field : PANEL_ISSUE_FIELDS[issue.key]);

// ป้ายของช่องตามตำแหน่ง: row 0 = หัวแบบ/แผงนี้, row n = คอลัมน์ในแถวใบแนบ (tax_id/name มีทั้งบนหัวแบบและในใบแนบ)
export type FieldLabel = (key: string, row: number) => string;

// fieldSuffix - " — ช่อง “ป้าย”" เมื่อรู้จักช่องนั้น (ป้ายที่ได้กลับมาเป็นรหัสช่องดิบ = ไม่รู้จัก ไม่แสดงรหัสเทคนิคให้ผู้ใช้)
const fieldSuffix = (tr: BackendTextFn, field: string | undefined, row: number, fieldLabel: FieldLabel) => {
  const label = field ? fieldLabel(field, row) : "";
  return label && label !== field ? ` — ${tr("tax_form_field", "ช่อง")} “${label}”` : "";
};

// rdIssueText - ข้อความของจุดที่ต้องแก้: ข้อความจากพจนานุกรม (key) → ข้อความที่ backend แปลมา → ข้อความกลาง
// + ชื่อช่อง + "แถวที่ n:" เมื่อเป็นแถวในใบแนบ; {char}/{max} แทนด้วย args ที่ backend ส่งมา
export function rdIssueText(tr: BackendTextFn, issue: TaxRdFileIssue, fieldLabel: FieldLabel): string {
  let message = fill(tr(issue.key, issue.message || tr("tax_rdfile_issue_unknown", "ข้อมูลไม่ตรงรูปแบบไฟล์ของกรมสรรพากร")), issue.args ?? {});
  // แม่แบบยังมี {…} ที่ไม่มีค่ามาแทน → ใช้ข้อความที่ backend แทนค่าแล้ว
  if (/\{\w+\}/.test(message) && issue.message) message = issue.message;
  const body = `${message}${fieldSuffix(tr, issue.field, issue.row, fieldLabel)}`;
  return issue.row > 0 ? fill(tr("tax_rdfile_issue_row", "แถวที่ {row}: {message}"), { row: String(issue.row), message: body }) : body;
}

type ButtonProps = {
  ready: boolean;
  working: boolean;
  busy: boolean;
  open: boolean;
  onToggle: () => void;
};

// TaxRdFileButton - ปุ่มในแถบปุ่ม (ถัดจากพิมพ์แบบ PDF) เปิด/ปิดแผง; ครอบด้วย span ให้ title แสดงแม้ปุ่มปิดอยู่
export function TaxRdFileButton({ ready, working, busy, open, onToggle }: ButtonProps) {
  const tr = useBackendText();
  const needSave = tr("tax_rdfile_need_save", "บันทึกฉบับนี้ก่อน ไฟล์สร้างจากฉบับที่บันทึกแล้ว");
  const label = tr("tax_rdfile_button", "ไฟล์ยื่นกรมสรรพากร (.txt)");
  return (
    <span className="inline-flex" title={ready ? label : needSave}>
      <Button
        type="button"
        variant="outline"
        className="h-11 gap-2"
        onClick={onToggle}
        disabled={working || !ready}
        aria-expanded={open}
        aria-controls={open ? TAX_RDFILE_PANEL_ID : undefined}
        aria-describedby={ready ? undefined : NEED_SAVE_ID}
      >
        {busy ? <Loader2 className="animate-spin" /> : <FileDown />}
        {label}
      </Button>
      {ready ? null : (
        <span id={NEED_SAVE_ID} className="sr-only">
          {needSave}
        </span>
      )}
    </span>
  );
}

type IssuesProps = {
  error: TaxRdFileError;
  fieldLabel: FieldLabel;
  onSelect: (issue: TaxRdFileIssue) => void;
};

// TaxRdFileIssues - สร้างไฟล์ไม่ผ่าน: รายการจุดที่ต้องแก้ (ไม่เกิน 50 + บอกจำนวนที่เหลือ) กดแล้วพาไปช่องนั้น
// ข้อผิดพลาดอื่น (ไม่มีรายการ เช่น ฉบับถูกแก้โดยผู้อื่น/เชื่อมต่อไม่ได้) แสดงเป็นข้อความเดียว
export function TaxRdFileIssues({ error, fieldLabel, onSelect }: IssuesProps) {
  const tr = useBackendText();
  if (error.issues.length === 0) {
    const where = error.row ? ` (${fill(tr("tax_form_row_label", "แถวที่ {n}"), { n: String(error.row) })})` : "";
    return (
      <div role="alert" className="flex items-start gap-2 rounded-xl border border-destructive/40 bg-destructive/10 px-3 py-2 leading-[1.5] text-destructive">
        <AlertTriangle className="mt-0.5 size-5 shrink-0" aria-hidden />
        <span>
          {tr(error.code, error.message ?? tr("tax_form_failed", "ทำรายการแบบยื่นภาษีไม่สำเร็จ กรุณาลองใหม่"))}
          {fieldSuffix(tr, error.field, error.row ?? 0, fieldLabel)}
          {where}
        </span>
      </div>
    );
  }
  const shown = error.issues.slice(0, RDFILE_ISSUE_ROWS);
  const more = Math.max(0, error.total - shown.length);
  return (
    <div role="alert" className="grid gap-2 rounded-xl border border-destructive/40 bg-destructive/10 px-3 py-2 leading-[1.5] text-foreground">
      <p className="flex items-center gap-2 font-semibold text-destructive">
        <AlertTriangle className="size-5 shrink-0" aria-hidden />
        {fill(tr("tax_rdfile_invalid_title", "สร้างไฟล์ไม่ได้ — พบ {n} จุดที่ต้องแก้"), { n: String(error.total) })}
      </p>
      <p className="text-muted-foreground">{tr("tax_rdfile_issues_hint", "กดที่รายการเพื่อไปยังช่องนั้น แก้แล้วกดบันทึก แล้วสร้างไฟล์ใหม่")}</p>
      <ol className="grid max-h-[40dvh] gap-0.5 overflow-auto">
        {shown.map((issue, i) => {
          const text = rdIssueText(tr, issue, fieldLabel);
          // กดได้เฉพาะเมื่อพาไปช่องได้จริง: ช่องในแผงนี้ หรือช่องที่มีบนแบบ/ใบแนบ (รู้จักป้าย) — ช่องที่แบบยังไม่มีคอลัมน์ให้เป็นข้อความ ไม่ใช่ปุ่มที่กดแล้วไม่เกิดอะไร
          const target = Boolean(panelField(issue)) || (issue.field ? fieldLabel(issue.field, issue.row) !== issue.field : false);
          return (
            <li key={`${issue.row}-${issue.field ?? ""}-${issue.key}-${i}`} className="[overflow-wrap:anywhere]">
              {target ? (
                <button
                  type="button"
                  onClick={() => onSelect(issue)}
                  className="flex min-h-[2.4em] w-full items-center gap-2 rounded-lg px-2 py-1 text-left underline-offset-4 transition-colors hover:bg-destructive/10 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
                >
                  <span className="shrink-0 font-semibold text-destructive" aria-hidden>
                    ›
                  </span>
                  <span>{text}</span>
                </button>
              ) : (
                <p className="flex min-h-[2.4em] items-center gap-2 px-2 py-1">
                  <span className="shrink-0 font-semibold text-destructive" aria-hidden>
                    •
                  </span>
                  <span>{text}</span>
                </p>
              )}
            </li>
          );
        })}
      </ol>
      {more > 0 ? <p className="font-medium">{fill(tr("tax_rdfile_more", "และอีก {n} รายการ"), { n: String(more) })}</p> : null}
    </div>
  );
}

type PanelProps = {
  filing: { id: number; version: number } | null;
  // มีฉบับที่บันทึก และไม่มีค่าที่แก้ค้าง
  ready: boolean;
  // จอกำลังทำงานอื่น (บันทึก/พิมพ์/ดึงยอด)
  working: boolean;
  // เลขสาขาบนแบบ: ศูนย์ล้วน = สำนักงานใหญ่ → เติมชื่อแผนกให้
  branchNo?: string;
  fieldLabel: FieldLabel;
  onCreate: (options: TaxRdFileOptions) => Promise<TaxRdFileResult | null>;
  // จุดที่ต้องแก้อยู่ที่ช่องบนแบบ/ใบแนบ → จอหลักไฮไลต์และพาไปช่องนั้น
  onIssue: (issue: TaxRdFileIssue) => void;
  onClose: () => void;
};

// ผลของการสร้างไฟล์ผูกกับฉบับ (id:version) — บันทึกรุ่นใหม่/เปิดฉบับอื่นแล้วผลเดิมไม่ตรงกับไฟล์ที่จะได้ จึงซ่อนเอง
type Outcome = { filingKey: string } & ({ kind: "ready"; filename: string } | { kind: "failed"; error: TaxRdFileError });

// downloadBlob - ดาวน์โหลดไบต์ตามที่ backend ส่ง (ห้ามแปลงเป็นข้อความ: BOM และ CRLF ต้องคงเดิม)
function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.rel = "noopener";
  document.body.appendChild(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export function TaxRdFilePanel({ filing, ready, working, branchNo, fieldLabel, onCreate, onIssue, onClose }: PanelProps) {
  const tr = useBackendText();
  const [options, setOptions] = useState<TaxRdFileOptions>(() => ({
    dept_name: isHeadOfficeBranch(branchNo) ? RD_HEAD_OFFICE_DEPT_NAME : "",
    submission_no: "00",
    lto: "",
    branch_type: "",
  }));
  const [running, setRunning] = useState(false);
  const [outcome, setOutcome] = useState<Outcome | null>(null);
  const filingKey = filing ? `${filing.id}:${filing.version}` : "";
  const shown = outcome && outcome.filingKey === filingKey ? outcome : null;
  const issues = shown?.kind === "failed" ? shown.error.issues : [];
  const invalidPanelField = (field: string) => issues.some((issue) => panelField(issue) === field);
  const set = (patch: Partial<TaxRdFileOptions>) => setOptions((current) => ({ ...current, ...patch }));
  const notSpecified = tr("tax_rdfile_not_specified", "ไม่ระบุ");

  const create = async () => {
    if (!ready || !filing || running) return;
    const key = filingKey;
    setRunning(true);
    const r = await onCreate(options);
    setRunning(false);
    if (!r) return;
    if (!r.ok) return setOutcome({ filingKey: key, kind: "failed", error: r.error });
    downloadBlob(r.data.blob, r.data.filename);
    setOutcome({ filingKey: key, kind: "ready", filename: r.data.filename });
  };

  const selectIssue = (issue: TaxRdFileIssue) => {
    const own = panelField(issue);
    if (own) {
      focusTaxFormField(PANEL_FIELD_IDS[own]);
      return;
    }
    onIssue(issue);
  };

  const panelLabels: Record<string, string> = {
    dept_name: tr("tax_rdfile_dept_name", "ชื่อแผนก/ฝ่าย หรือสำนักงานใหญ่"),
    submission_no: tr("tax_rdfile_submission_no", "ครั้งที่ส่ง (00–99)"),
    lto: tr("tax_rdfile_lto", "ผู้ประกอบการรายใหญ่ (LTO)"),
    branch_type: tr("tax_rdfile_branch_type", "ประเภทสาขา"),
  };
  const panelLabel: FieldLabel = (key, row) => (row === 0 ? panelLabels[key] : undefined) ?? fieldLabel(key, row);
  // กรอบแดงรอบกลุ่มตัวเลือกที่ต้องแก้ (ตัวเลือกแบบปุ่มไม่มีช่องให้ใส่ aria-invalid)
  const choiceGroup = (field: string) => cn("col-span-full grid gap-1 rounded-xl", invalidPanelField(field) && "p-1.5 ring-2 ring-destructive/40");

  return (
    <section
      id={TAX_RDFILE_PANEL_ID}
      aria-labelledby="tax-rdfile-title"
      className="grid gap-3 rounded-xl border border-primary/40 bg-primary/5 p-3 text-sm leading-[1.5] text-foreground shadow-[var(--shadow-card)] sm:p-4"
    >
      <header className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-2">
          <FileDown className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden />
          <div className="grid min-w-0 gap-0.5">
            <h2 id="tax-rdfile-title" className="text-base font-semibold leading-[1.45]">
              {tr("tax_rdfile_title", "สร้างไฟล์ยื่นแบบด้วยสื่อ (รูปแบบข้อมูล Format กลาง)")}
            </h2>
            <p className="text-muted-foreground">
              {tr("tax_rdfile_help", "สำหรับโปรแกรม SWC-UI ของกรมสรรพากร (ยื่นแบบด้วยสื่อฝากไฟล์ออนไลน์) ต้องลงทะเบียนบริการก่อน · อัปโหลดเข้า e-Filing ตรง ๆ ไม่ได้")}
            </p>
            {filing ? (
              <p className="text-muted-foreground">{fill(tr("tax_rdfile_source", "สร้างจากฉบับที่บันทึก รุ่นที่ {version}"), { version: String(filing.version) })}</p>
            ) : null}
          </div>
        </div>
        <Button type="button" variant="outline" size="sm" className="shrink-0 gap-1" onClick={onClose}>
          <X />
          {tr("close", "ปิด")}
        </Button>
      </header>

      <div className="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(min(100%,17rem),1fr))]">
        <div className="grid content-start gap-1">
          <label htmlFor={PANEL_FIELD_IDS.dept_name} className="font-medium">
            {panelLabels.dept_name}
          </label>
          <Input
            id={PANEL_FIELD_IDS.dept_name}
            value={options.dept_name}
            maxLength={80}
            onChange={(e) => set({ dept_name: e.target.value })}
            aria-invalid={invalidPanelField("dept_name") || undefined}
            aria-describedby="tax-rdfile-dept-hint"
            className={cn("min-h-[2.6em]", invalidPanelField("dept_name") && "border-destructive ring-2 ring-destructive/25")}
          />
          <span id="tax-rdfile-dept-hint" className="text-xs leading-[1.5] text-muted-foreground">
            {tr("tax_rdfile_dept_name_hint", "ไม่เกิน 80 ตัวอักษร · สาขาที่ไม่ใช่สำนักงานใหญ่ต้องกรอกเอง")}
          </span>
        </div>
        <div className="grid content-start gap-1">
          <label htmlFor={PANEL_FIELD_IDS.submission_no} className="font-medium">
            {panelLabels.submission_no}
          </label>
          <Input
            id={PANEL_FIELD_IDS.submission_no}
            value={options.submission_no}
            inputMode="numeric"
            maxLength={2}
            placeholder="00"
            onChange={(e) => {
              if (/^\d{0,2}$/.test(e.target.value)) set({ submission_no: e.target.value });
            }}
            aria-invalid={invalidPanelField("submission_no") || undefined}
            aria-describedby="tax-rdfile-submission-hint"
            className={cn("min-h-[2.6em] tabular-nums", invalidPanelField("submission_no") && "border-destructive ring-2 ring-destructive/25")}
          />
          <span id="tax-rdfile-submission-hint" className="text-xs leading-[1.5] text-muted-foreground">
            {tr("tax_rdfile_submission_hint", "ใช้เป็นส่วนท้ายของชื่อไฟล์ ค่าเริ่มต้น 00")}
          </span>
        </div>
        <div id={PANEL_FIELD_IDS.lto} className={choiceGroup("lto")}>
          <p className="font-medium">{panelLabels.lto}</p>
          <ChoiceSelect<TaxRdFileLto>
            value={options.lto}
            onChange={(lto) => set({ lto })}
            options={[
              { value: "", label: notSpecified },
              { value: "0", label: tr("tax_rdfile_lto_no", "ไม่เป็น") },
              { value: "1", label: tr("tax_rdfile_lto_yes", "เป็น") },
            ]}
            aria-label={panelLabels.lto}
          />
          <span className="text-xs leading-[1.5] text-muted-foreground">{tr("tax_rdfile_optional_hint", "ยื่นด้วยสื่อไม่บังคับ")}</span>
        </div>
        <div id={PANEL_FIELD_IDS.branch_type} className={choiceGroup("branch_type")}>
          <p className="font-medium">{panelLabels.branch_type}</p>
          <ChoiceSelect<TaxRdFileBranchType>
            value={options.branch_type}
            onChange={(branch_type) => set({ branch_type })}
            options={[
              { value: "", label: notSpecified },
              { value: "V", label: tr("tax_rdfile_branch_type_v", "สาขาภาษีมูลค่าเพิ่ม (V)") },
              { value: "S", label: tr("tax_rdfile_branch_type_s", "สาขาภาษีธุรกิจเฉพาะ (S)") },
            ]}
            aria-label={panelLabels.branch_type}
          />
          <span className="text-xs leading-[1.5] text-muted-foreground">
            {tr("tax_rdfile_optional_hint", "ยื่นด้วยสื่อไม่บังคับ")} · {tr("tax_rdfile_branch_type_hint", "เป็นทั้งสาขาภาษีมูลค่าเพิ่มและภาษีธุรกิจเฉพาะ ให้เลือก V")}
          </span>
        </div>
      </div>

      {ready ? null : (
        <p role="status" className="flex items-center gap-2 font-medium">
          <Info className="size-5 shrink-0 text-primary" aria-hidden />
          {tr("tax_rdfile_need_save", "บันทึกฉบับนี้ก่อน ไฟล์สร้างจากฉบับที่บันทึกแล้ว")}
        </p>
      )}
      <div>
        <Button type="button" className="h-11 gap-2 px-5" onClick={() => void create()} disabled={!ready || working || running}>
          {running ? <Loader2 className="animate-spin" /> : <FileDown />}
          {tr("tax_rdfile_create", "สร้างไฟล์")}
        </Button>
      </div>
      {shown?.kind === "ready" ? (
        <p role="status" className="flex items-center gap-2 font-medium text-primary [overflow-wrap:anywhere]">
          <CheckCircle2 className="size-5 shrink-0" aria-hidden />
          {fill(tr("tax_rdfile_ready", "สร้างไฟล์แล้ว: {filename}"), { filename: shown.filename })}
        </p>
      ) : null}
      {shown?.kind === "failed" ? <TaxRdFileIssues error={shown.error} fieldLabel={panelLabel} onSelect={selectIssue} /> : null}
    </section>
  );
}

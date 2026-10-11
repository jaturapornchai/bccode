"use client";

import { Fragment } from "react";
import { AlertTriangle, CheckCircle2 } from "lucide-react";
import { formatAmount, rowsHoldDimension, type GLClosePreview, type GLClosePreviewBranch, type GLClosePreviewLine, type GLDimensionDef } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { EXTRA_DIMENSIONS } from "@/lib/gl-journal-entry";
import { Notice, useGLText, useRowDensity } from "./gl-common";

/** Asks the backend for the entries the real close would create (processes "close-preview"); it reads the same generator as the close and saves nothing. */
export async function fetchClosePreview(input: { year: string; version?: number; date: string; docno: string }): Promise<GLClosePreview | null> {
  const docno = input.docno.trim();
  const data = await glRequest<{ closepreview?: Partial<GLClosePreview> | null }>("command", { method: "POST", body: JSON.stringify({ resource: "processes", action: "close-preview", id: input.year, version: input.version, date: input.date, ...(docno ? { docno } : {}), requestid: crypto.randomUUID() }) });
  const preview = data?.closepreview;
  if (!preview) return null;
  return {
    fiscalyear: preview.fiscalyear ?? input.year, date: preview.date ?? input.date, scale: preview.scale ?? 2, pendingdrafts: Boolean(preview.pendingdrafts),
    netprofit: preview.netprofit ?? "0", totaldebit: preview.totaldebit ?? "0", totalcredit: preview.totalcredit ?? "0", balanced: Boolean(preview.balanced),
    branches: (preview.branches ?? []).map((branch) => ({ ...branch, lines: branch.lines ?? [] })),
  };
}

/** Blank cell for the empty side of a line (Champ shows only the side that has an amount); works on the decimal string, no number math. */
export function isZeroAmount(value: string): boolean { return /^-?0*\.?0*$/.test(value.trim()); }

/** Department and project are always columns of the closing entries; the other four analysis dimensions only when a closing line carries a code of it. */
export function closePreviewExtraDimensions(lines: readonly GLClosePreviewLine[]): GLDimensionDef[] {
  return EXTRA_DIMENSIONS.filter((def) => rowsHoldDimension(lines, def));
}

function BranchEntries({ branch, scale }: { branch: GLClosePreviewBranch; scale: number }) {
  const tr = useGLText(), density = useRowDensity();
  const extra = closePreviewExtraDimensions(branch.lines);
  const firstTransfer = branch.lines.findIndex((line) => line.transfer);
  const money = (value: string) => (isZeroAmount(value) ? "" : formatAmount(value, scale));
  return <section className="grid gap-2" data-close-branch={branch.branchcode}>
    <h3 className="flex flex-wrap items-baseline gap-x-4 gap-y-1 text-[1rem] font-semibold">
      <span>{tr("gl_branch_code", "รหัสสาขา")} {branch.branchcode || "-"}</span>
      {branch.docno && <span>{tr("gl_document_no", "เลขที่เอกสาร")} {branch.docno}</span>}
      <span className="tabular-nums">{tr("gl_periods_net_profit_loss", "กำไร(ขาดทุน)สุทธิ")} {formatAmount(branch.netprofit, scale)}</span>
    </h3>
    <div className="overflow-auto rounded-xl border border-border">
      <table className={`w-full min-w-[820px] border-collapse text-[0.95rem] ${density.tableClass}`} style={extra.length > 0 ? { minWidth: 820 + extra.length * 110 } : undefined}>
        <thead className="sticky top-0 z-10 bg-muted text-left">
          <tr className="[&_th]:px-3 [&_th]:py-2 [&_th]:font-semibold">
            <th>{tr("gl_account_code", "รหัสบัญชี")}</th>
            <th>{tr("gl_account_name", "ชื่อบัญชี")}</th>
            <th>{tr("gl_department", "แผนก")}</th>
            <th>{tr("gl_project", "โครงการ")}</th>
            {extra.map((def) => <th key={def.line}>{tr(def.label[0], def.label[1])}</th>)}
            <th className="text-right">{tr("gl_debit", "เดบิต")}</th>
            <th className="text-right">{tr("gl_credit", "เครดิต")}</th>
          </tr>
        </thead>
        <tbody>
          {branch.lines.map((line, index) => <Fragment key={index}>
            {index === firstTransfer && <tr className="border-t border-border bg-muted/50"><td colSpan={6 + extra.length} className="px-3 py-2 font-semibold">{tr("gl_close_preview_transfer", "โอนกำไรขาดทุนสุทธิเข้ากำไรสะสม")}</td></tr>}
            <tr className="border-t border-border align-top [&_td]:px-3 hover:bg-muted/40" data-close-line={line.transfer ? "transfer" : "close"}>
              <td className="whitespace-nowrap font-medium">{line.accountcode}</td>
              <td className="[overflow-wrap:anywhere]">{line.accountname}</td>
              <td className="whitespace-nowrap">{line.departmentcode}</td>
              <td className="whitespace-nowrap">{line.projectcode}</td>
              {extra.map((def) => <td key={def.line} className="whitespace-nowrap">{line[def.line] ?? ""}</td>)}
              <td className="whitespace-nowrap text-right tabular-nums">{money(line.debit)}</td>
              <td className="whitespace-nowrap text-right tabular-nums">{money(line.credit)}</td>
            </tr>
          </Fragment>)}
        </tbody>
        <tfoot>
          <tr className="border-t-2 border-border bg-muted/40 font-semibold [&_td]:px-3 [&_td]:py-2">
            <td colSpan={4 + extra.length}>{tr("gl_total", "ยอดรวม")}</td>
            <td className="whitespace-nowrap text-right tabular-nums">{formatAmount(branch.totaldebit, scale)}</td>
            <td className="whitespace-nowrap text-right tabular-nums">{formatAmount(branch.totalcredit, scale)}</td>
          </tr>
        </tfoot>
      </table>
    </div>
  </section>;
}

/** The closing entries per branch (Champ 5504 tab "net profit/loss" + tab "retained earnings") with totals proving debit = credit. */
export function ClosePreviewTable({ preview }: { preview: GLClosePreview }) {
  const tr = useGLText();
  const stat = "min-w-36 flex-1 rounded-xl border border-border bg-muted/40 p-3 shadow-sm";
  return <div className="grid gap-3" data-close-preview="true">
    <h2 className="text-[1.05rem] font-semibold">{tr("gl_close_preview_title", "รายการปิดบัญชีที่ระบบจะสร้างเป็นฉบับร่าง")}</h2>
    {preview.pendingdrafts && <Notice error text={tr("gl_close_preview_pending", "ยังมีรายการร่างที่ลงวันที่ถึงวันที่นี้ กรุณาผ่านรายการให้ครบก่อน แล้วกด ประมวลผล ใหม่อีกครั้ง")} />}
    <div className="flex flex-wrap items-stretch gap-2">
      <div className={stat}><div className="text-[0.9rem] text-muted-foreground">{tr("gl_periods_net_profit_loss", "กำไร(ขาดทุน)สุทธิ")}</div><div className="text-lg font-semibold tabular-nums">{formatAmount(preview.netprofit, preview.scale)}</div></div>
      <div className={stat}><div className="text-[0.9rem] text-muted-foreground">{tr("gl_debit", "เดบิต")}</div><div className="text-lg font-semibold tabular-nums">{formatAmount(preview.totaldebit, preview.scale)}</div></div>
      <div className={stat}><div className="text-[0.9rem] text-muted-foreground">{tr("gl_credit", "เครดิต")}</div><div className="text-lg font-semibold tabular-nums">{formatAmount(preview.totalcredit, preview.scale)}</div></div>
      <div className={`${stat} flex items-center`}>{preview.balanced
        ? <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-1 font-semibold text-primary"><CheckCircle2 className="size-4 shrink-0" aria-hidden="true" />{tr("gl_close_preview_balanced", "ยอดเดบิตเท่ากับเครดิต")}</span>
        : <span className="inline-flex items-center gap-1.5 rounded-full bg-destructive/10 px-2.5 py-1 font-semibold text-foreground"><AlertTriangle className="size-4 shrink-0" aria-hidden="true" />{tr("gl_close_preview_unbalanced", "ยอดเดบิตไม่เท่ากับเครดิต กรุณาแจ้งผู้ดูแลระบบ")}</span>}</div>
    </div>
    {preview.branches.map((branch) => <BranchEntries key={`${branch.branchcode}/${branch.docno}`} branch={branch} scale={preview.scale} />)}
  </div>;
}

"use client";

import { useEffect, useRef, useState } from "react";
import { Plus, RefreshCw, Save, Search, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { emptyAllocationRule, emptyMaster, type GLAllocationRule, type GLMaster, type GLRecord, type GLResource } from "@/lib/general-ledger";
import { GLCommandError, commandFailure, glRequest } from "@/lib/general-ledger-api";
import { AccountSelect, AmountInput, Check, Field, Notice, Pager, SearchInput, SplitWorkbench, actionClass, useDebouncedSearch, useDirtyGuard, useGLCommand, useGLList, useReferences, useGLText } from "./gl-common";

type AllocationResource = Extract<GLResource, "allocations">;
const RESOURCE: AllocationResource = "allocations";

function allocationFrom(value: GLRecord): GLMaster {
  const base = emptyMaster();
  const master = { ...base, ...(value as GLMaster) };
  master.allocatemode = master.allocatemode || "percent";
  master.allocaterules = Array.isArray(master.allocaterules) ? master.allocaterules : [];
  return master;
}

/** Percentage total of the rules as an exact string for display; the server is the
 *  authority (it rejects any total other than 100), this only guides the user. */
export function rateTotal(rules: GLAllocationRule[]): string {
  let units = 0n;
  for (const rule of rules) {
    const text = (rule.rate || "0").replace(/,/g, "").trim() || "0";
    if (!/^\d+(\.\d{1,8})?$/.test(text)) continue;
    const [whole, fraction = ""] = text.split(".");
    units += BigInt(whole) * 100000000n + BigInt(fraction.padEnd(8, "0"));
  }
  const whole = units / 100000000n;
  const fraction = (units % 100000000n).toString().padStart(8, "0").replace(/0+$/, "");
  return fraction ? `${whole}.${fraction}` : String(whole);
}

export function GLAllocations({ route }: { route: string }) {
  const tr = useGLText();
  const [search, setSearch] = useState("");
  const list = useGLList<GLMaster>(RESOURCE, search);
  const searchDebounce = useDebouncedSearch({ onSearch: (value) => { list.setPage(1); setSearch(value); }, debounceMs: 2000 });
  const refs = useReferences();
  const [record, setRecord] = useState<GLMaster | null>(null), [original, setOriginal] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [message, setMessage] = useState(""), [error, setError] = useState("");
  const [errorField, setErrorField] = useState("");
  const formRef = useRef<HTMLFormElement>(null);
  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const dirty = isEditing && record !== null && JSON.stringify(record) !== original;
  useDirtyGuard(route, dirty);
  const set = (patch: object) => setRecord((current) => current ? { ...current, ...patch } : current);

  function showFormError(cause: unknown, fallback: string, field = "") {
    const info = commandFailure(cause, fallback, field);
    setError(info.message);
    setErrorField(cause instanceof GLCommandError && (cause.code === "duplicate_code" || cause.code === "immutable_code") ? "code" : info.field || field || "code");
  }
  useEffect(() => {
    if (!error || !errorField || busy) return;
    const node = formRef.current?.querySelector<HTMLElement>(`[data-field="${errorField}"]`);
    const target = node?.matches("input,select,textarea,button") ? node : node?.querySelector<HTMLElement>("input,select,textarea,button");
    if (target && document.activeElement !== target) target.focus({ preventScroll: true });
  }, [error, errorField, busy]);

  async function open(item: GLMaster | null, editing: boolean) {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      const value = item?.id ? await glRequest<GLRecord>(`${RESOURCE}/${encodeURIComponent(item.id)}`) : emptyMaster();
      const master = allocationFrom(value);
      setRecord(master);
      setOriginal(JSON.stringify(master));
      setIsEditing(editing);
      setMessage(""); setError(""); setErrorField("");
    } catch (e) { showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")); }
  }

  async function cancelEdit() {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    if (record?.id) { setRecord(JSON.parse(original)); setIsEditing(false); }
    else { setRecord(null); setOriginal(""); setIsEditing(false); }
    setError(""); setErrorField("");
  }

  async function save() {
    if (!record || busy || !isEditing) return;
    if (!record.code.trim() || !record.name.trim()) { showFormError(null, tr("gl_enter_code_and_name", "กรุณาระบุรหัสและชื่อให้ครบ"), !record.code.trim() ? "code" : "name"); return; }
    if (!record.accountcode) { showFormError(null, tr("gl_alloc_account_required", "กรุณาเลือกบัญชีต้นทุนที่ต้องการปันส่วน"), "accountcode"); return; }
    if (record.allocaterules.length === 0) { showFormError(null, tr("gl_alloc_rules_required", "กรุณากำหนดรายการปันส่วนอย่างน้อย 1 รายการ")); return; }
    if (rateTotal(record.allocaterules) !== "100") { showFormError(null, tr("gl_alloc_rate_total_100", "อัตราการปันส่วนรวมต้องเท่ากับ 100 เปอร์เซ็นต์")); return; }
    if (record.id && !await confirm({ title: tr("gl_save_changes_confirm", "บันทึกการแก้ไขข้อมูล?"), description: tr("gl_edit_with_history", "แก้ไข {0} โดยเก็บประวัติการเปลี่ยนแปลง").replace("{0}", String(record.code)), confirmLabel: tr("gl_save_changes", "บันทึกการแก้ไข"), tone: "info" })) return;
    try {
      setError("");
      const result = await execute({ resource: RESOURCE, action: record.id ? "update" : "create", id: record.id, version: record.version, reason: tr("gl_alloc_save_reason", "บันทึกการปันส่วนค่าใช้จ่าย"), master: record });
      const saved = { ...record, id: result.id, version: result.version };
      setRecord(saved); setOriginal(JSON.stringify(saved)); setIsEditing(false); list.reload();
      setMessage(result.projectionpending ? tr("gl_saved_updating_reload", "บันทึกแล้ว กำลังปรับปรุงข้อมูลสำหรับรายงาน กดโหลดใหม่เพื่อตรวจสอบ") : tr("gl_saved_successfully", "บันทึกเรียบร้อยแล้ว"));
    } catch (e) { showFormError(e, tr("gl_save_failed_retry", "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"), "code"); }
  }

  async function remove(item: GLMaster) {
    if (!item.id || busy) return;
    if (!await confirm({ title: tr("gl_confirm_delete", "ยืนยันลบรายการ?"), description: `${item.code} · ${item.name}`, confirmLabel: tr("gl_delete_item", "ลบรายการ"), tone: "danger" })) return;
    try {
      await execute({ resource: RESOURCE, id: item.id, version: item.version, action: "delete", reason: tr("gl_delete_item", "ลบรายการ") });
      if (record?.id === item.id) { setRecord(null); setOriginal(""); setIsEditing(false); }
      list.reload();
      setMessage(tr("gl_delete_success", "ลบ {0} เรียบร้อยแล้ว").replace("{0}", String(item.code)));
    } catch (e) { showFormError(e, tr("gl_transaction_failed_try_again", "ทำรายการไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")); }
  }

  const total = record ? rateTotal(record.allocaterules) : "0";
  const balanced = total === "100";

  return <div className="flex flex-col flex-1 min-h-0 gap-2">
    <div className="shrink-0 flex flex-col gap-2"><Notice error text={record ? "" : [error, list.error, refs.error].filter(Boolean).join(" ")} /><Notice text={message} /></div>
    <SplitWorkbench list={<div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="shrink-0 pb-3 border-b border-border/70 flex flex-col gap-2">
        <form className="flex flex-wrap items-center gap-2 w-full" onSubmit={(event) => { event.preventDefault(); searchDebounce.searchNow(); }}>
          <SearchInput className="min-w-44 flex-1" value={searchDebounce.query} onChange={searchDebounce.setQuery} onClear={searchDebounce.clear} onSearch={searchDebounce.searchNow} placeholder={tr("gl_search_code_name", "ค้นหารหัสหรือชื่อ")} />
          <div className="flex flex-wrap items-center gap-2 shrink-0">
            <Button type="submit" variant="outline" className={actionClass}><Search className="size-4 mr-1.5" />{tr("gl_search", "ค้นหา")}</Button>
            <Button type="button" variant="outline" className={actionClass} onClick={() => list.reload()} disabled={list.loading}><RefreshCw className="size-4 mr-1.5" />{tr("gl_reload", "โหลดใหม่")}</Button>
            <Button type="button" className={actionClass} onClick={() => void open(null, true)} disabled={busy}><Plus className="size-4 mr-1.5" />{tr("gl_add_row", "เพิ่มรายการ")}</Button>
          </div>
        </form>
        <div className="text-xs text-muted-foreground px-1 font-medium text-foreground/80">{tr("gl_x_items", "{0} รายการ").replace("{0}", String(list.data.total.toLocaleString("th-TH")))}</div>
      </div>
      <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-xs" aria-busy={list.loading}>
        <table className="w-full text-left text-[0.95rem] leading-normal">
          <thead className="sticky top-0 bg-muted z-10"><tr>
            <th className="p-2.5">{tr("gl_code", "รหัส")}</th>
            <th className="p-2.5">{tr("gl_name_description", "ชื่อ / รายละเอียด")}</th>
            <th className="p-2.5">{tr("gl_alloc_cost_account", "บัญชีต้นทุน")}</th>
            <th className="p-2.5 text-center w-24">{tr("gl_alloc_lines", "จำนวนบรรทัด")}</th>
            <th className="p-2.5 text-center w-28">{tr("gl_status", "สถานะ")}</th>
            <th className="p-2.5 text-right pr-3 w-20">{tr("gl_manage", "จัดการ")}</th>
          </tr></thead>
          <tbody className="divide-y divide-border">
            {list.data.items.map((item, index) => (
              <tr key={item.id} className={`cursor-pointer transition-colors ${record?.id === item.id ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium" : index % 2 === 0 ? "bg-background hover:bg-accent/60" : "bg-muted/20 hover:bg-accent/60"}`} onClick={() => void open(item, false)}>
                <td className="p-2 whitespace-nowrap font-mono font-bold text-primary">{item.code}</td>
                <td className="p-2 max-w-72 truncate" title={item.name}>{item.name}</td>
                <td className="p-2 whitespace-nowrap font-mono">{item.accountcode || "-"}</td>
                <td className="p-2 text-center">{item.allocaterules?.length ?? 0}</td>
                <td className="p-2 text-center">{item.isactive ? <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary border border-primary/20">{tr("gl_enable", "ใช้งาน")}</span> : <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground border border-border">{tr("gl_disable", "ปิดใช้งาน")}</span>}</td>
                <td className="p-2 text-right pr-2" onClick={(event) => event.stopPropagation()}>
                  <Button type="button" variant="ghost" className="!min-h-0 !p-1.5" aria-label={tr("gl_delete_item", "ลบรายการ")} onClick={() => void remove(item)}><Trash2 className="size-4 text-destructive" /></Button>
                </td>
              </tr>
            ))}
            {!list.data.items.length && <tr><td className="p-5 text-center text-muted-foreground" colSpan={6}>{tr("gl_no_records_found_criteria", "ไม่พบรายการตามเงื่อนไขที่เลือก")}</td></tr>}
          </tbody>
        </table>
      </div>
      <Pager page={list.page} total={list.data.total} onPage={list.setPage} loading={list.loading} limit={30} />
    </div>} editor={<form ref={formRef} className="flex flex-col flex-1 min-h-0 gap-3" onSubmit={(event) => { event.preventDefault(); void save(); }}>
      {!record ? <div className="grid place-items-center h-full text-muted-foreground text-[0.95rem]">{tr("gl_select_row_to_view", "เลือกรายการจากตารางเพื่อดูข้อมูล")}</div> : <>
        <div className="shrink-0 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-semibold">{isEditing ? (record.id ? tr("gl_edit_data", "แก้ไขรายการ") : tr("gl_add_row", "เพิ่มรายการ")) : tr("gl_view_item", "ดูข้อมูลรายการ")}</h2>
          <div className="flex gap-2">
            {isEditing ? <>
              <Button type="button" variant="outline" className={actionClass} onClick={() => void cancelEdit()} disabled={busy}>{tr("common_cancel", "ยกเลิก")}</Button>
              <Button type="submit" className={actionClass} disabled={busy}><Save className="size-4 mr-1.5" />{tr("gl_save", "บันทึก")}</Button>
            </> : <Button type="button" className={actionClass} onClick={() => setIsEditing(true)} disabled={busy}>{tr("gl_edit", "แก้ไข")}</Button>}
          </div>
        </div>
        {record && error && <div className="shrink-0"><Notice error text={error} /></div>}
        <div className="flex-1 min-h-0 overflow-auto grid gap-3 content-start">
          <fieldset disabled={!isEditing || busy} className="grid gap-3">
            <div className="grid gap-3 sm:grid-cols-2">
              <Field label={tr("gl_code", "รหัส")}><input data-field="code" className="min-h-[2.6em] w-full rounded-xl border border-input bg-background px-3 py-1.5" value={record.code} onChange={(event) => set({ code: event.target.value })} maxLength={60} /></Field>
              <Field label={tr("gl_name_description", "ชื่อ / รายละเอียด")}><input data-field="name" className="min-h-[2.6em] w-full rounded-xl border border-input bg-background px-3 py-1.5" value={record.name} onChange={(event) => set({ name: event.target.value })} maxLength={300} /></Field>
            </div>
            <Field label={tr("gl_alloc_cost_account", "บัญชีต้นทุนที่ต้องการปันส่วน")}><AccountSelect value={record.accountcode} onChange={(value) => set({ accountcode: value })} accounts={refs.accounts} field="accountcode" /></Field>
            <div className="flex flex-wrap items-center gap-3"><Check label={tr("gl_enable", "ใช้งาน")} checked={record.isactive} onChange={(checked) => set({ isactive: checked })} />
              <span className={`text-[0.95rem] ${balanced ? "text-muted-foreground" : "text-destructive font-medium"}`}>{tr("gl_alloc_rate_total", "อัตรารวม {0}%").replace("{0}", total)}{balanced ? "" : ` · ${tr("gl_alloc_rate_total_100", "อัตราการปันส่วนรวมต้องเท่ากับ 100 เปอร์เซ็นต์")}`}</span>
            </div>
          </fieldset>

          <section className="grid gap-2 rounded-xl border border-border p-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-medium">{tr("gl_alloc_rules", "รายการปันส่วนตามสัดส่วน")}</h3>
              <Button type="button" variant="outline" className={actionClass} disabled={!isEditing || busy || record.allocaterules.length >= 100} onClick={() => set({ allocaterules: [...record.allocaterules, emptyAllocationRule()] })}><Plus className="size-4 mr-1.5" />{tr("gl_alloc_add_rule", "เพิ่มรายการปันส่วน")}</Button>
            </div>
            <div className="overflow-auto rounded-xl border border-border shadow-xs">
              <table className="w-full text-left text-[0.95rem]">
                <thead className="bg-muted"><tr>
                  <th className="p-2 w-14 text-center">{tr("gl_sequence", "ลำดับ")}</th>
                  <th className="p-2">{tr("gl_alloc_target_account", "บัญชีปลายทาง")}</th>
                  <th className="p-2 w-32">{tr("gl_branch_code", "รหัสสาขา")}</th>
                  <th className="p-2 w-32">{tr("gl_department_code", "รหัสแผนก")}</th>
                  <th className="p-2 w-32">{tr("gl_project_code", "รหัสโครงการ")}</th>
                  <th className="p-2 w-32 text-right">{tr("gl_alloc_rate", "อัตรา %")}</th>
                  <th className="p-2 w-16"></th>
                </tr></thead>
                <tbody className="divide-y divide-border">
                  {record.allocaterules.map((rule, index) => (
                    <tr key={index}>
                      <td className="p-2 text-center text-muted-foreground">{index + 1}</td>
                      <td className="p-2"><AccountSelect value={rule.accountcode} onChange={(value) => set({ allocaterules: record.allocaterules.map((entry, i) => i === index ? { ...entry, accountcode: value } : entry) })} accounts={refs.accounts} allowEmpty label={tr("gl_alloc_target_account", "บัญชีปลายทาง")} /></td>
                      <td className="p-2"><input className="min-h-[2.4em] w-full rounded-lg border border-input bg-background px-2 py-1" value={rule.branchcode} onChange={(event) => set({ allocaterules: record.allocaterules.map((entry, i) => i === index ? { ...entry, branchcode: event.target.value } : entry) })} maxLength={60} /></td>
                      <td className="p-2"><input className="min-h-[2.4em] w-full rounded-lg border border-input bg-background px-2 py-1" value={rule.departmentcode} onChange={(event) => set({ allocaterules: record.allocaterules.map((entry, i) => i === index ? { ...entry, departmentcode: event.target.value } : entry) })} maxLength={60} /></td>
                      <td className="p-2"><input className="min-h-[2.4em] w-full rounded-lg border border-input bg-background px-2 py-1" value={rule.projectcode} onChange={(event) => set({ allocaterules: record.allocaterules.map((entry, i) => i === index ? { ...entry, projectcode: event.target.value } : entry) })} maxLength={60} /></td>
                      <td className="p-2"><AmountInput value={rule.rate} onChange={(value) => set({ allocaterules: record.allocaterules.map((entry, i) => i === index ? { ...entry, rate: value } : entry) })} scale={2} ariaLabel={tr("gl_alloc_rate", "อัตรา %")} /></td>
                      <td className="p-2 text-center"><Button type="button" variant="ghost" className="!min-h-0 !p-1.5" disabled={!isEditing || busy} aria-label={tr("gl_delete_item", "ลบรายการ")} onClick={() => set({ allocaterules: record.allocaterules.filter((_, i) => i !== index) })}><Trash2 className="size-4 text-destructive" /></Button></td>
                    </tr>
                  ))}
                  {!record.allocaterules.length && <tr><td className="p-4 text-center text-muted-foreground" colSpan={7}>{tr("gl_alloc_no_rules", "ยังไม่มีรายการปันส่วน กดเพิ่มรายการปันส่วน")}</td></tr>}
                </tbody>
                <tfoot className="bg-muted/60"><tr>
                  <td className="p-2 text-right font-medium" colSpan={5}>{tr("gl_alloc_rate_total_short", "รวมอัตรา")}</td>
                  <td className={`p-2 text-right font-mono font-semibold tabular-nums ${balanced ? "" : "text-destructive"}`}>{total}%</td>
                  <td></td>
                </tr></tfoot>
              </table>
            </div>
          </section>
        </div>
      </>}
    </form>} />
    {confirmationDialog}
  </div>;
}

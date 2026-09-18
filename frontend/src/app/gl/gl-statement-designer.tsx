"use client";

import { useEffect, useState } from "react";
import {
  Plus,
  RefreshCw,
  Save,
  Trash2,
  Copy,
  SlidersHorizontal,
  Eye,
  Printer,
  Download,
  ArrowUp,
  ArrowDown,
  FileSpreadsheet,
  Sparkles,
  Search,
  Pencil,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Checkbox } from "@/components/ui/checkbox";
import {
  formatAmount,
  emptyStatementTemplate,
  statementTypeLabels,
  statementRowTypeLabels,
  labelText,
  type GLLabel,
  generateStarterTemplates,
  calculateStatement,
  type GLStatementTemplate,
  type StatementRow,
  type StatementType,
  type StatementRowType,
  type StatementRowUnderline,
  type CalculatedStatement,
} from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import {
  Combobox,
  Field,
  Notice,
  Pager,
  SplitWorkbench,
  YearSelect,
  actionClass,
  control,
  useDirtyGuard,
  useGLCommand,
  useGLList,
  useReferences,
  useRowDensity,
  downloadText,
  AccountSearchDialog,
  SearchInput,
  useDebouncedSearch,
  useGLText,
} from "./gl-common";
import { fetchReport, type ReportFilters, emptyReportFilters } from "./gl-reports";

const FONT_OPTIONS: { id: string; name: GLLabel; family: string; href: string }[] = [
  { id: "sarabun", name: ["gl_font_family_sarabun", "Sarabun (สารบรรณ - มาตรฐานทางการ)"], family: '"Sarabun", sans-serif', href: "https://fonts.googleapis.com/css2?family=Sarabun:wght@400;500;600;700&display=swap" },
  { id: "prompt", name: ["gl_font_family_prompt", "Prompt (พร้อม - อ่านง่าย ผู้บริหาร)"], family: '"Prompt", sans-serif', href: "https://fonts.googleapis.com/css2?family=Prompt:wght@400;500;600;700&display=swap" },
  { id: "kanit", name: ["gl_font_kanit", "Kanit (คณิต - คมชัด ทันสมัย)"], family: '"Kanit", sans-serif', href: "https://fonts.googleapis.com/css2?family=Kanit:wght@400;500;600;700&display=swap" },
  { id: "noto-sans-thai", name: ["gl_font_noto_sans_thai", "Noto Sans Thai (มาตรฐานสากล)"], family: '"Noto Sans Thai", sans-serif', href: "https://fonts.googleapis.com/css2?family=Noto+Sans+Thai:wght@400;500;600;700&display=swap" },
  { id: "inter", name: ["gl_font_inter", "Inter (สากล โมเดิร์น)"], family: '"Inter", sans-serif', href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" },
  { id: "monospace", name: ["gl_font_courier_monospace", "Courier / Monospace (ตัวเลขพิมพ์ดีด)"], family: "ui-monospace, monospace", href: "" },
];

function ensureFontLoaded(fontId: string) {
  const opt = FONT_OPTIONS.find((f) => f.id === fontId);
  if (!opt || !opt.href) return;
  const linkId = `font-link-${opt.id}`;
  if (!document.getElementById(linkId)) {
    const link = document.createElement("link");
    link.id = linkId;
    link.rel = "stylesheet";
    link.href = opt.href;
    document.head.appendChild(link);
  }
}

export function GLStatementDesigner({ route = "/gl/statement-designer" }: { route?: string }) {
  const tr = useGLText();
  const [search, setSearch] = useState("");
  const list = useGLList<GLStatementTemplate>("statement-templates", search);
  const searchDebounce = useDebouncedSearch({
    onSearch: (val) => {
      list.setPage(1);
      setSearch(val);
    },
    debounceMs: 2000,
  });
  const density = useRowDensity();
  const refs = useReferences();
  const [template, setTemplate] = useState<GLStatementTemplate | null>(null);
  const [original, setOriginal] = useState("");
  const [activeTab, setActiveTab] = useState<"design" | "preview">("design");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [starterModalOpen, setStarterModalOpen] = useState(false);
  const [accountPickerRowId, setAccountPickerRowId] = useState<string | null>(null);

  // Live preview state
  const [filters, setFilters] = useState<ReportFilters>({ ...emptyReportFilters });
  const [calculating, setCalculating] = useState(false);
  const [calculated, setCalculated] = useState<CalculatedStatement | null>(null);
  const [previewError, setPreviewError] = useState("");

  const { busy, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });

  const dirty = template !== null && JSON.stringify(template) !== original;

  useDirtyGuard(route, dirty);

  // Ensure current font is loaded
  useEffect(() => {
    if (template?.globalstyle?.fontfamily) {
      ensureFontLoaded(template.globalstyle.fontfamily);
    }
  }, [template?.globalstyle?.fontfamily]);

  // Set default year for preview
  useEffect(() => {
    if (refs.years.length && !filters.fiscalyear) {
      const active = refs.years.find((y) => y.isactive && !y.closed) ?? refs.years[0];
      if (active) {
        setFilters((prev) => ({ ...prev, fiscalyear: active.code, from: active.startdate, to: active.enddate }));
      }
    }
  }, [refs.years, filters.fiscalyear]);

  async function open(item?: GLStatementTemplate, targetTab: "preview" | "design" = "preview") {
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_editing_data_not_saved", "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก"), tone: "warning", confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    try {
      if (item?.id) {
        const value = await glRequest<GLStatementTemplate>(`statement-templates/${encodeURIComponent(item.id)}`);
        const normalized = { ...emptyStatementTemplate(), ...value };
        setTemplate(normalized);
        setOriginal(JSON.stringify(normalized));
        setActiveTab(targetTab);
      } else {
        const fresh = emptyStatementTemplate();
        setTemplate(fresh);
        setOriginal(JSON.stringify(fresh));
        setActiveTab("design");
      }
      setMessage("");
      setError("");
      setCalculated(null);
    } catch (e) {
      setError((e as Error).message);
    }
  }

  function applyStarterTemplate(starter: GLStatementTemplate) {
    const fresh: GLStatementTemplate = {
      ...starter,
      id: template?.id,
      version: template?.version,
      code: template?.id ? (template.code || starter.code) : starter.code,
      name: starter.name,
    };
    setTemplate(fresh);
    setStarterModalOpen(false);
    setMessage(tr("gl_template_loaded", "โหลดแม่แบบ \"{0}\" เรียบร้อยแล้ว").replace("{0}", String(starter.name)));
  }

  async function save() {
    if (!template || busy) return;
    if (!template.code.trim() || !template.name.trim()) {
      setError(tr("gl_specify_fin_stmt_template_code_name", "กรุณาระบุรหัสและชื่อแม่แบบงบการเงิน"));
      return;
    }
    if (template.id && !await confirm({ title: tr("gl_save_fin_stmt_template_changes", "บันทึกการแก้ไขแม่แบบงบ?"), description: tr("gl_edit_item", "แก้ไข {0} ({1})").replace("{0}", String(template.code)).replace("{1}", String(template.name)), confirmLabel: tr("gl_save_changes", "บันทึกการแก้ไข"), tone: "info" })) return;
    try {
      setError("");
      const result = await execute({
        resource: "statement-templates",
        action: template.id ? "update" : "create",
        id: template.id,
        version: template.version,
        master: {
          code: template.code,
          name: template.name,
          statementtype: template.statementtype,
          isactive: template.isactive,
          globalstyle: template.globalstyle,
          rows: template.rows,
        } as never,
      });
      const saved: GLStatementTemplate = { ...template, id: result.id, version: result.version };
      setTemplate(saved);
      setOriginal(JSON.stringify(saved));
      list.reload();
      setMessage(tr("gl_fin_stmt_template_saved", "บันทึกแม่แบบงบการเงินเรียบร้อยแล้ว"));
    } catch (e) {
      setError((e as Error).message);
    }
  }

  async function cloneTemplate() {
    if (!template) return;
    const cloned: GLStatementTemplate = {
      ...template,
      id: undefined,
      version: undefined,
      code: `${template.code}-COPY`,
      name: tr("gl_copy_of", "{0} (คัดลอก)").replace("{0}", String(template.name)),
    };
    setTemplate(cloned);
    setOriginal("");
    setMessage(tr("gl_template_copied", "คัดลอกแม่แบบเป็น \"{0}\" เรียบร้อยแล้ว (กรุณากดบันทึก)").replace("{0}", String(cloned.name)));
  }

  async function remove() {
    if (!template?.id || busy) return;
    if (!await confirm({
      title: tr("gl_delete_fin_template_confirm", "ลบแม่แบบงบการเงิน?"),
      description: tr("gl_confirm_delete_template", "คุณต้องการลบแม่แบบ {0} ({1}) ใช่หรือไม่?").replace("{0}", String(template.code)).replace("{1}", String(template.name)),
      confirmLabel: tr("gl_delete_template", "ลบแม่แบบ"),
      tone: "danger",
    })) return;
    try {
      setError("");
      await execute({
        resource: "statement-templates",
        id: template.id,
        version: template.version,
        action: "delete",
        reason: tr("gl_delete_fin_template", "ลบแม่แบบงบการเงิน"),
      });
      setTemplate(null);
      setOriginal("");
      list.reload();
      setMessage(tr("gl_delete_fin_template_success", "ลบแม่แบบงบการเงินเรียบร้อยแล้ว"));
    } catch (e) {
      setError((e as Error).message);
    }
  }

  // Row Manipulation Helpers
  function addRow(type: StatementRowType) {
    if (!template) return;
    const currentRows = template.rows ?? [];
    const maxRowNo = currentRows.reduce((max, r) => (r.rowno > max ? r.rowno : max), 0);
    const nextRowNo = Math.floor(maxRowNo / 10 + 1) * 10;
    const newRow: StatementRow = {
      id: `row-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
      rowno: nextRowNo,
      rowtype: type,
      title: type === "blank" ? "" : type === "divider" ? "—" : tr("gl_line_no", "รายการที่ {0}").replace("{0}", String(nextRowNo)),
      style: {
        indent: type === "header" ? 0 : type === "subtotal" ? 1 : 2,
        fontweight: type === "header" || type === "subtotal" ? "bold" : "normal",
        underline: type === "subtotal" ? "single" : "none",
      },
    };
    setTemplate({ ...template, rows: [...currentRows, newRow] });
  }

  function updateRow(rowId: string, patch: Partial<StatementRow>) {
    if (!template) return;
    const nextRows = template.rows.map((r) => (r.id === rowId ? { ...r, ...patch } : r));
    setTemplate({ ...template, rows: nextRows });
  }

  function deleteRow(rowId: string) {
    if (!template) return;
    setTemplate({ ...template, rows: template.rows.filter((r) => r.id !== rowId) });
  }

  function moveRow(index: number, direction: "up" | "down") {
    if (!template) return;
    const targetIndex = direction === "up" ? index - 1 : index + 1;
    if (targetIndex < 0 || targetIndex >= template.rows.length) return;
    const nextRows = [...template.rows];
    const [moved] = nextRows.splice(index, 1);
    nextRows.splice(targetIndex, 0, moved);
    setTemplate({ ...template, rows: nextRows });
  }

  // Calculation for preview
  async function runCalculation() {
    if (!template) return;
    if (!filters.fiscalyear) {
      setPreviewError(tr("gl_select_fiscal_year_before_process", "กรุณาเลือกปีบัญชีก่อนประมวลผล"));
      return;
    }
    setCalculating(true);
    setPreviewError("");
    try {
      // The server caps a page at 1000 rows; follow the pages on the same snapshot.
      const first = await fetchReport("trialbalance", filters, 1, 1000);
      const rows = [...(first.rows ?? [])];
      for (let page = 2; rows.length < first.totalrows; page++) {
        const next = await fetchReport("trialbalance", filters, page, 1000, first.sequence);
        if (!next.rows?.length) break;
        rows.push(...next.rows);
      }
      const res = calculateStatement(template, rows);
      setCalculated(res);
    } catch (e) {
      setPreviewError((e as Error).message);
    } finally {
      setCalculating(false);
    }
  }

  const updateGlobalStyle = (patch: Partial<GLStatementTemplate["globalstyle"]>) => {
    if (!template) return;
    setTemplate({
      ...template,
      globalstyle: { ...template.globalstyle, ...patch },
    });
  };

  const selectedFont = FONT_OPTIONS.find((f) => f.id === (template?.globalstyle?.fontfamily ?? "sarabun")) ?? FONT_OPTIONS[0];

  return (
    <div className="flex flex-col flex-1 min-h-0 gap-2">
      <div className="shrink-0 flex flex-col gap-2">
        <Notice error text={error || list.error || refs.error} />
        <Notice text={message} />
      </div>

      <SplitWorkbench
        list={
          <div className="flex flex-col flex-1 min-h-0 gap-2">
            <div className="flex flex-wrap items-center justify-between gap-2 shrink-0">
              <h2 className="text-base font-semibold">{tr("gl_financial_statement_template", "แม่แบบงบการเงิน")}</h2>
              <Button type="button" className={actionClass} onClick={() => void open()} disabled={busy}>
                <Plus className="mr-1.5 h-4 w-4" /> {tr("gl_create_new_template", "สร้างแม่แบบใหม่")}
              </Button>
            </div>

            <div className="shrink-0 pb-3 border-b border-border/70 flex flex-col gap-2">
              <form className="flex flex-wrap items-center gap-2 w-full" onSubmit={(e) => { e.preventDefault(); searchDebounce.searchNow(); }}>
                <SearchInput
                  className="min-w-44 flex-1"
                  ariaLabel={tr("gl_search_code_or_template_name", "ค้นหารหัสหรือชื่อแม่แบบ")}
                  placeholder={tr("gl_search_code_or_template_name_dots", "ค้นหารหัสหรือชื่อแม่แบบ...")}
                  value={searchDebounce.query}
                  onChange={searchDebounce.setQuery}
                  onClear={searchDebounce.clear}
                  onSearch={searchDebounce.searchNow}
                />
                <div className="flex flex-wrap items-center gap-2 shrink-0">
                  <Button type="submit" variant="outline" className={actionClass}>
                    <Search className="h-4 w-4 mr-1.5" />
                    {tr("gl_search", "ค้นหา")}
                  </Button>
                  <Button type="button" variant="outline" className={actionClass} onClick={() => list.reload()} disabled={list.loading}>
                    <RefreshCw className="h-4 w-4 mr-1.5" />
                    {tr("gl_reload", "โหลดใหม่")}
                  </Button>
                </div>
              </form>
            </div>

            <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm" aria-busy={list.loading}>
              <table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}>
                <thead className="sticky top-0 bg-muted z-10">
                  <tr>
                    <th className="p-2.5">{tr("gl_code", "รหัส")}</th>
                    <th className="p-2.5 min-w-44">{tr("gl_statement_template_name", "ชื่อแม่แบบงบ")}</th>
                    <th className="p-2.5 text-center w-28">{tr("gl_type", "ประเภท")}</th>
                    <th className="p-2.5 text-right pr-3 w-20">{tr("gl_manage", "จัดการ")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {list.data.items.map((item, index) => {
                    const isSelected = template?.id === item.id;
                    const isRowEditing = isSelected && activeTab === "design";
                    return (
                      <tr
                        key={item.id}
                        className={`group cursor-pointer transition-colors ${
                          isRowEditing
                            ? "bg-primary/15 hover:bg-primary/20 text-foreground ring-1 ring-inset ring-primary/50 font-medium"
                            : isSelected
                              ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
                              : index % 2 === 0
                                ? "bg-background hover:bg-accent/60"
                                : "bg-muted/20 hover:bg-accent/60"
                        }`}
                        onClick={() => void open(item, "preview")}
                      >
                        <td className="p-2 whitespace-nowrap">
                          <span className="font-mono font-bold text-primary">{item.code}</span>
                        </td>
                      <td className="max-w-48 truncate p-2" title={item.name}>
                        {item.name}
                      </td>
                      <td className="whitespace-nowrap p-2 text-center text-xs">
                        <span className="rounded-md border border-border bg-muted/60 px-2 py-0.5 font-medium">
                          {labelText(statementTypeLabels, item.statementtype, tr)}
                        </span>
                      </td>
                      <td className="p-2 text-right whitespace-nowrap pr-2" onClick={(e) => e.stopPropagation()}>
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            type="button"
                            size="icon"
                            variant="outline"
                            className="size-7 rounded-md bg-background text-primary border-primary/30 hover:bg-primary/15 hover:border-primary/50 shadow-none transition-colors shrink-0"
                            onClick={(e) => {
                              e.stopPropagation();
                              void open(item, "design");
                            }}
                            aria-label={tr("gl_edit", "แก้ไข")}
                            title={tr("gl_edit_chart_of_accounts", "แก้ไขผังงบ (Edit)")}
                          >
                            <Pencil className="size-3.5 shrink-0" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
                  {!list.data.items.length && (
                    <tr>
                      <td colSpan={4} className="p-6 text-center text-muted-foreground">
                        {list.loading ? tr("gl_loading_data_2", "กำลังโหลดข้อมูล...") : tr("gl_no_fin_stmt_template_yet", "ยังไม่มีแม่แบบงบการเงิน กดสร้างใหม่หรือใช้แม่แบบมาตรฐาน")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
            <Pager page={list.page} total={list.data.total} onPage={list.setPage} loading={list.loading} />
          </div>
        }
        editor={
          template ? (
            <div className="flex flex-col h-full min-h-0 gap-3">
              {/* Editor Header Bar */}
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3 shrink-0">
                <div className="flex flex-wrap items-center gap-2">
                  <div className="flex rounded-xl border border-border bg-muted/30 p-1">
                    <Button
                      type="button"
                      size="sm"
                      variant={activeTab === "design" ? "default" : "ghost"}
                      className="rounded-lg text-sm font-semibold"
                      onClick={() => setActiveTab("design")}
                    >
                      <SlidersHorizontal className="mr-1.5 h-4 w-4" /> {tr("gl_layout_design_mode", "โหมดออกแบบผัง")}
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant={activeTab === "preview" ? "default" : "ghost"}
                      className="rounded-lg text-sm font-semibold"
                      onClick={() => {
                        setActiveTab("preview");
                        if (!calculated) void runCalculation();
                      }}
                    >
                      <Eye className="mr-1.5 h-4 w-4" /> {tr("gl_preview_print_statements", "พรีวิวและพิมพ์งบจริง")}
                    </Button>
                  </div>

                  <Button
                    type="button"
                    variant="outline"
                    className={actionClass}
                    onClick={() => setStarterModalOpen(true)}
                  >
                    <Sparkles className="mr-1.5 h-4 w-4 text-primary" /> {tr("gl_use_standard_template", "ใช้แม่แบบมาตรฐาน...")}
                  </Button>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  {template.id && (
                    <>
                      <Button type="button" variant="outline" className={actionClass} onClick={() => void cloneTemplate()}>
                        <Copy className="mr-1.5 h-4 w-4" /> {tr("gl_copy", "คัดลอก")}
                      </Button>
                      <Button type="button" variant="outline" className={`${actionClass} text-destructive hover:bg-destructive/10`} onClick={() => void remove()} disabled={busy}>
                        <Trash2 className="mr-1.5 h-4 w-4" /> {tr("gl_delete", "ลบ")}
                      </Button>
                    </>
                  )}
                  {activeTab === "design" ? (
                    <Button type="button" className={actionClass} onClick={() => void save()} disabled={busy}>
                      <Save className="mr-1.5 h-4 w-4" /> {busy ? tr("gl_saving_2", "กำลังบันทึก...") : tr("gl_save_template", "บันทึกแม่แบบ")}
                    </Button>
                  ) : (
                    <Button type="button" className={actionClass} onClick={() => setActiveTab("design")} disabled={busy}>
                      <SlidersHorizontal className="mr-1.5 h-4 w-4" /> {tr("gl_edit_statement_layout", "แก้ไขผังงบ")}
                    </Button>
                  )}
                </div>
              </div>

              {/* Template Metadata & Global Styling */}
              <div className="grid gap-3 rounded-xl border border-border bg-muted/20 p-3 sm:grid-cols-2 xl:grid-cols-4 shrink-0">
                <Field label={tr("gl_fin_stmt_template_code", "รหัสแม่แบบงบ")}>
                  <input
                    className={control}
                    value={template.code}
                    onChange={(e) => setTemplate({ ...template, code: e.target.value })}
                    placeholder={tr("gl_fin_stmt_template_code_ex", "เช่น BS-01, PNL-01")}
                  />
                </Field>
                <Field label={tr("gl_fin_stmt_template_name", "ชื่อแม่แบบงบการเงิน")}>
                  <input
                    className={control}
                    value={template.name}
                    onChange={(e) => setTemplate({ ...template, name: e.target.value })}
                    placeholder={tr("gl_fin_stmt_template_name_ex", "เช่น งบแสดงฐานะการเงิน (แบบ DBD)")}
                  />
                </Field>
                <Field label={tr("gl_fin_stmt_type", "ประเภทงบ")}>
                  <Combobox
                    value={template.statementtype}
                    onChange={(val) => setTemplate({ ...template, statementtype: String(val) as StatementType })}
                    placeholder={tr("gl_fin_stmt_type", "ประเภทงบ")}
                  >
                    {Object.entries(statementTypeLabels).map(([key, label]) => (
                      <option key={key} value={key}>{tr(...label)}</option>
                    ))}
                  </Combobox>
                </Field>

                {/* Font Customizer */}
                <Field label={tr("gl_font_family", "แบบตัวอักษร (Font Family)")}>
                  <Combobox
                    value={template.globalstyle?.fontfamily ?? "sarabun"}
                    onChange={(val) => updateGlobalStyle({ fontfamily: String(val) })}
                    placeholder={tr("gl_font_family", "แบบตัวอักษร")}
                  >
                    {FONT_OPTIONS.map((font) => (
                      <option key={font.id} value={font.id}>{tr(...font.name)}</option>
                    ))}
                  </Combobox>
                </Field>

                <Field label={tr("gl_base_font_size", "ขนาดตัวอักษรพื้นฐาน")}>
                  <Combobox
                    value={template.globalstyle?.fontsize ?? "15px"}
                    onChange={(val) => updateGlobalStyle({ fontsize: String(val) })}
                    placeholder={tr("gl_base_font_size", "ขนาดตัวอักษร")}
                  >
                    <option value="13px">{tr("gl_font_size_13_compact", "13px - กะทัดรัด")}</option>
                    <option value="14px">{tr("gl_font_size_14_normal", "14px - ปกติ")}</option>
                    <option value="15px">{tr("gl_font_size_15_comfort", "15px - สบายตา (แนะนำ 40+)")}</option>
                    <option value="16px">{tr("gl_font_size_16_large", "16px - ตัวใหญ่")}</option>
                    <option value="18px">{tr("gl_font_size_18_extra", "18px - พิเศษ")}</option>
                  </Combobox>
                </Field>

                <div className="flex flex-wrap items-center gap-4 xl:col-span-2 pt-1">
                  <label className="inline-flex w-auto shrink-0 items-center gap-2.5 text-sm font-medium cursor-pointer select-none">
                    <Checkbox
                      checked={template.globalstyle?.shownotecolumn ?? true}
                      onCheckedChange={(checked) => updateGlobalStyle({ shownotecolumn: checked })}
                    />
                    <span className="whitespace-nowrap">{tr("gl_show_notes_column", "แสดงคอลัมน์หมายเหตุประกอบงบ")}</span>
                  </label>
                  <label className="inline-flex w-auto shrink-0 items-center gap-2.5 text-sm font-medium cursor-pointer select-none">
                    <Checkbox
                      checked={template.isactive}
                      onCheckedChange={(checked) => setTemplate({ ...template, isactive: checked })}
                    />
                    <span className="whitespace-nowrap">{tr("gl_activate_template", "เปิดใช้งานแม่แบบนี้")}</span>
                  </label>
                </div>
              </div>

              {/* TAB 1: DESIGN MODE */}
              {activeTab === "design" && (
                <div className="flex flex-col flex-1 min-h-0 gap-3">
                  {/* Row actions */}
                  <div className="flex flex-wrap items-center justify-between gap-2 shrink-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <Button type="button" size="sm" variant="outline" className={actionClass} onClick={() => addRow("header")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> {tr("gl_heading", "หัวข้อ")}
                      </Button>
                      <Button type="button" size="sm" variant="default" className={actionClass} onClick={() => addRow("account")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> {tr("gl_account_row", "แถวบัญชี")}
                      </Button>
                      <Button type="button" size="sm" variant="outline" className={actionClass} onClick={() => addRow("formula")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> {tr("gl_formula_row", "แถวสูตร")}
                      </Button>
                      <Button type="button" size="sm" variant="outline" className={actionClass} onClick={() => addRow("subtotal")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> {tr("gl_subtotal_row", "แถวรวมย่อย")}
                      </Button>
                      <Button type="button" size="sm" variant="ghost" className={actionClass} onClick={() => addRow("blank")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> {tr("gl_blank_line_2", "บรรทัดว่าง")}
                      </Button>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {tr("gl_total_lines", "ทั้งหมด {0} บรรทัด").replace("{0}", String(template.rows.length))}
                    </span>
                  </div>

                  {/* Rows Table */}
                  <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border shadow-sm">
                    <table className="w-full text-left text-sm">
                      <thead className="sticky top-0 bg-muted font-semibold">
                        <tr>
                          <th className="w-16 p-2 text-center">{tr("gl_sequence", "ลำดับ")}</th>
                          <th className="w-24 p-2">{tr("gl_type", "ประเภท")}</th>
                          <th className="p-2">{tr("gl_line_item_name", "ชื่อรายการในงบ")}</th>
                          <th className="w-16 p-2 text-center">{tr("gl_note", "หมายเหตุ")}</th>
                          <th className="p-2">{tr("gl_account_mapping_formula", "การผูกบัญชี / สูตรคำนวณ")}</th>
                          <th className="w-48 p-2">{tr("gl_styling", "การจัดสไตล์")}</th>
                          <th className="w-24 p-2 text-center">{tr("gl_manage", "จัดการ")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {template.rows.map((row, index) => (
                          <tr key={row.id} className="hover:bg-muted/30">
                            {/* Row No */}
                            <td className="p-2 text-center">
                              <input
                                type="number"
                                className="w-14 rounded border border-input bg-background p-1 text-center text-xs font-mono font-bold"
                                value={row.rowno}
                                onChange={(e) => updateRow(row.id, { rowno: parseInt(e.target.value, 10) || 0 })}
                              />
                            </td>

                            {/* Row Type */}
                            <td className="p-2">
                              <select
                                className="w-full rounded border border-input bg-background p-1 text-xs"
                                value={row.rowtype}
                                onChange={(e) => updateRow(row.id, { rowtype: e.target.value as StatementRowType })}
                              >
                                {Object.entries(statementRowTypeLabels).map(([k, v]) => (
                                  <option key={k} value={k}>{tr(...v)}</option>
                                ))}
                              </select>
                            </td>

                            {/* Title */}
                            <td className="p-2">
                              {row.rowtype === "blank" ? (
                                <span className="text-xs italic text-muted-foreground">{tr("gl_blank_line", "(บรรทัดว่าง)")}</span>
                              ) : row.rowtype === "divider" ? (
                                <span className="text-xs font-mono text-muted-foreground">────────────────</span>
                              ) : (
                                <input
                                  className="w-full rounded border border-input bg-background px-2 py-1 text-sm font-medium"
                                  value={row.title}
                                  onChange={(e) => updateRow(row.id, { title: e.target.value })}
                                  placeholder={tr("gl_enter_item_name", "ระบุชื่อรายการ...")}
                                />
                              )}
                            </td>

                            {/* Note No */}
                            <td className="p-2 text-center">
                              {row.rowtype !== "blank" && row.rowtype !== "divider" && (
                                <input
                                  className="w-12 rounded border border-input bg-background p-1 text-center text-xs font-mono"
                                  value={row.noteno ?? ""}
                                  onChange={(e) => updateRow(row.id, { noteno: e.target.value })}
                                  placeholder={tr("gl_example_3", "เช่น 3")}
                                />
                              )}
                            </td>

                            {/* Account Codes or Formula */}
                            <td className="p-2">
                              {row.rowtype === "account" && (
                                <div className="flex flex-wrap items-center gap-1.5">
                                  <button
                                    type="button"
                                    onClick={() => setAccountPickerRowId(row.id)}
                                    className="inline-flex min-h-7 items-center rounded-md border border-primary/30 bg-primary/5 px-2 py-1 text-xs font-semibold text-primary hover:bg-primary/10"
                                  >
                                    {row.accountcodes?.length
                                      ? tr("gl_selected_accounts", "เลือกแล้ว {0} บัญชี").replace("{0}", String(row.accountcodes.length))
                                      : tr("gl_select_chart_of_accounts", "+ เลือกผังบัญชี")}
                                  </button>

                                  <select
                                    className="rounded border border-input bg-background p-1 text-xs"
                                    value={row.normalbalance ?? "debit"}
                                    onChange={(e) => updateRow(row.id, { normalbalance: e.target.value as "debit" | "credit" | "net" })}
                                  >
                                    <option value="debit">{tr("gl_debit_plus", "เดบิต (+)")}</option>
                                    <option value="credit">{tr("gl_credit_plus", "เครดิต (+)")}</option>
                                    <option value="net">{tr("gl_net", "สุทธิ")}</option>
                                  </select>

                                  <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer select-none" title={tr("gl_reverse_sign", "กลับเครื่องหมายบวกลบ")}>
                                    <Checkbox
                                      checked={row.reversesign ?? false}
                                      onCheckedChange={(checked) => updateRow(row.id, { reversesign: checked })}
                                    />
                                    +/-
                                  </label>
                                </div>
                              )}

                              {(row.rowtype === "formula" || row.rowtype === "subtotal") && (
                                <div className="flex items-center gap-2">
                                  <input
                                    className="w-full rounded border border-input bg-background px-2 py-1 font-mono text-xs"
                                    value={row.formula ?? ""}
                                    onChange={(e) => updateRow(row.id, { formula: e.target.value })}
                                    placeholder={row.rowtype === "subtotal" ? tr("gl_example_sum_range", "เช่น SUM(R10:R40)") : tr("gl_example_add_subtract", "เช่น R10 + R20 - R30")}
                                  />
                                </div>
                              )}
                            </td>

                            {/* Row Styling */}
                            <td className="p-2">
                              <div className="flex flex-wrap items-center gap-1">
                                {/* Bold */}
                                <button
                                  type="button"
                                  className={`h-7 w-7 rounded border font-bold text-xs ${row.style?.fontweight === "bold" ? "bg-primary text-primary-foreground border-primary" : "border-input bg-background"}`}
                                  onClick={() => updateRow(row.id, { style: { ...row.style, fontweight: row.style?.fontweight === "bold" ? "normal" : "bold" } })}
                                  title={tr("gl_bold", "ตัวหนา")}
                                >
                                  B
                                </button>
                                {/* Italic */}
                                <button
                                  type="button"
                                  className={`h-7 w-7 rounded border italic text-xs ${row.style?.fontstyle === "italic" ? "bg-primary text-primary-foreground border-primary" : "border-input bg-background"}`}
                                  onClick={() => updateRow(row.id, { style: { ...row.style, fontstyle: row.style?.fontstyle === "italic" ? "normal" : "italic" } })}
                                  title={tr("gl_italic", "ตัวเอียง")}
                                >
                                  I
                                </button>

                                {/* Indent */}
                                <select
                                  className="h-7 rounded border border-input bg-background px-1 text-xs"
                                  value={row.style?.indent ?? 0}
                                  onChange={(e) => updateRow(row.id, { style: { ...row.style, indent: parseInt(e.target.value, 10) } })}
                                  title={tr("gl_indent_level", "ระดับการเยื้อง")}
                                >
                                  <option value={0}>{tr("gl_indent_none", "ไม่เยื้อง")}</option>
                                  <option value={1}>{tr("gl_indent_1", "เยื้อง 1")}</option>
                                  <option value={2}>{tr("gl_indent_2", "เยื้อง 2")}</option>
                                  <option value={3}>{tr("gl_indent_3", "เยื้อง 3")}</option>
                                  <option value={4}>{tr("gl_indent_4", "เยื้อง 4")}</option>
                                </select>

                                {/* Underline */}
                                <select
                                  className="h-7 rounded border border-input bg-background px-1 text-xs"
                                  value={row.style?.underline ?? "none"}
                                  onChange={(e) => updateRow(row.id, { style: { ...row.style, underline: e.target.value as StatementRowUnderline } })}
                                  title={tr("gl_underline_account", "เส้นใต้บัญชี")}
                                >
                                  <option value="none">{tr("gl_no_line", "ไร้เส้น")}</option>
                                  <option value="single">{tr("gl_single_underline", "ขีดเดี่ยว _")}</option>
                                  <option value="double">{tr("gl_double_underline_net", "ขีดคู่ = (ยอดสุทธิ)")}</option>
                                  <option value="top_single_bottom_double">{tr("gl_top_single_bottom_double", "บนเดี่ยว ล่างคู่")}</option>
                                </select>
                              </div>
                            </td>

                            {/* Ordering / Delete */}
                            <td className="p-2 text-center">
                              <div className="flex items-center justify-center gap-1">
                                <button
                                  type="button"
                                  className="h-6 w-6 rounded hover:bg-muted"
                                  disabled={index === 0}
                                  onClick={() => moveRow(index, "up")}
                                  title={tr("gl_move_up", "เลื่อนขึ้น")}
                                >
                                  <ArrowUp className="h-3.5 w-3.5 mx-auto" />
                                </button>
                                <button
                                  type="button"
                                  className="h-6 w-6 rounded hover:bg-muted"
                                  disabled={index === template.rows.length - 1}
                                  onClick={() => moveRow(index, "down")}
                                  title={tr("gl_move_down", "เลื่อนลง")}
                                >
                                  <ArrowDown className="h-3.5 w-3.5 mx-auto" />
                                </button>
                                <button
                                  type="button"
                                  className="h-6 w-6 rounded text-destructive hover:bg-destructive/10"
                                  onClick={() => deleteRow(row.id)}
                                  title={tr("gl_delete_this_row", "ลบแถวนี้")}
                                >
                                  <Trash2 className="h-3.5 w-3.5 mx-auto" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {/* TAB 2: LIVE PREVIEW & PRINT */}
              {activeTab === "preview" && (
                <div className="flex flex-col flex-1 min-h-0 gap-3">
                  <div className="shrink-0 flex flex-col gap-2">
                    <Notice error text={previewError} />
                  </div>

                  {/* Preview Criteria Toolbar */}
                  <div className="flex flex-wrap items-end justify-between gap-3 rounded-xl border border-border bg-muted/20 p-3 shrink-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <div className="w-48">
                        <Field label={tr("gl_fiscal_year", "ปีบัญชี")}>
                          <YearSelect
                            years={refs.years}
                            value={filters.fiscalyear}
                            onChange={(val) => {
                              const y = refs.years.find((item) => item.code === val);
                              setFilters((prev) => ({ ...prev, fiscalyear: val, from: y?.startdate ?? "", to: y?.enddate ?? "" }));
                            }}
                          />
                        </Field>
                      </div>
                      <div className="w-36">
                        <Field label={tr("gl_from_date", "ตั้งแต่วันที่")}>
                          <input type="date" className={control} value={filters.from} onChange={(e) => setFilters((prev) => ({ ...prev, from: e.target.value }))} />
                        </Field>
                      </div>
                      <div className="w-36">
                        <Field label={tr("gl_to_date", "ถึงวันที่")}>
                          <input type="date" className={control} value={filters.to} onChange={(e) => setFilters((prev) => ({ ...prev, to: e.target.value }))} />
                        </Field>
                      </div>
                      <Button type="button" className={`${actionClass} mt-auto`} onClick={() => void runCalculation()} disabled={calculating}>
                        <RefreshCw className={`mr-1.5 h-4 w-4 ${calculating ? "animate-spin" : ""}`} />
                        {calculating ? tr("gl_calculating", "กำลังคำนวณ...") : tr("gl_calculate_and_display", "คำนวณและแสดงผล")}
                      </Button>
                    </div>

                    <div className="flex items-center gap-2">
                      <Button type="button" variant="outline" className={actionClass} onClick={() => window.print()}>
                        <Printer className="mr-1.5 h-4 w-4" /> {tr("gl_print_financial_statements", "พิมพ์งบการเงิน")}
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        className={actionClass}
                        onClick={() => {
                          if (!calculated) return;
                          const csv = "\uFEFF" + [
                            [tr("gl_sequence", "ลำดับ"), tr("gl_items", "รายการ"), tr("gl_note", "หมายเหตุ"), tr("gl_amount", "จำนวนเงิน")].join(","),
                            ...calculated.rows.map((r) => [`"'${r.rowno}"`, `"${r.title.replace(/"/g, '""')}"`, `"${r.noteno ?? ""}"`, `"'${r.amountFormatted}"`].join(",")),
                          ].join("\r\n");
                          downloadText(tr("gl_financial_statements_csv", "งบการเงิน-{0}-{1}.csv").replace("{0}", String(template.code)).replace("{1}", String(filters.fiscalyear)), csv, "text/csv;charset=utf-8");
                        }}
                      >
                        <Download className="mr-1.5 h-4 w-4" /> {tr("gl_export_csv", "ส่งออก CSV")}
                      </Button>
                    </div>
                  </div>

                  {/* WYSIWYG Live Report Viewer */}
                  <div
                    className="flex-1 min-h-[300px] overflow-auto rounded-2xl border border-border bg-card p-6 shadow-sm print:m-0 print:border-none print:p-0 print:shadow-none"
                    style={{
                      fontFamily: selectedFont.family,
                      fontSize: template.globalstyle?.fontsize ?? "15px",
                    }}
                  >
                    {/* Header */}
                    <div className="mb-6 text-center">
                      <h1 className="text-xl font-bold tracking-tight">{template.name}</h1>
                      <p className="text-sm text-muted-foreground">{tr("gl_for_period_as_of", "สำหรับงวดบัญชี {0} (ณ วันที่ {1})").replace("{0}", String(filters.fiscalyear)).replace("{1}", String(filters.to || "-"))}</p>
                      <p className="text-xs text-muted-foreground">{tr("gl_unit_baht", "(หน่วย: บาท)")}</p>
                    </div>

                    {/* Statement Table */}
                    <table className="w-full border-collapse">
                      <thead>
                        <tr className="border-b-2 border-foreground/30">
                          <th className="py-2 text-left font-bold">{tr("gl_items", "รายการ")}</th>
                          {template.globalstyle?.shownotecolumn && (
                            <th className="w-24 py-2 text-center font-bold">{tr("gl_note", "หมายเหตุ")}</th>
                          )}
                          <th className="w-44 py-2 text-right font-bold">{tr("gl_amount_2", "ยอดเงิน")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {calculated?.rows.map((row) => {
                          const indentPx = (row.style?.indent ?? 0) * 20;
                          const isBold = row.style?.fontweight === "bold";
                          const isItalic = row.style?.fontstyle === "italic";

                          return (
                            <tr key={row.id} className="hover:bg-muted/10">
                              <td
                                className={`py-1.5 ${isBold ? "font-bold" : ""} ${isItalic ? "italic" : ""}`}
                                style={{ paddingLeft: `${indentPx}px` }}
                              >
                                {row.title}
                              </td>

                              {template.globalstyle?.shownotecolumn && (
                                <td className="py-1.5 text-center text-xs text-muted-foreground">
                                  {row.noteno ?? ""}
                                </td>
                              )}

                              <td className="py-1.5 text-right tabular-nums">
                                {row.amountFormatted && (
                                  <span
                                    className={`inline-block min-w-24 ${isBold ? "font-bold" : ""} ${
                                      row.style?.underline === "single"
                                        ? "border-b border-foreground"
                                        : row.style?.underline === "double"
                                        ? "border-b-4 border-double border-foreground"
                                        : row.style?.underline === "top_single_bottom_double"
                                        ? "border-t border-b-4 border-double border-foreground"
                                        : ""
                                    }`}
                                  >
                                    {row.amountFormatted}
                                  </span>
                                )}
                              </td>
                            </tr>
                          );
                        })}

                        {!calculated?.rows.length && (
                          <tr>
                            <td colSpan={3} className="py-8 text-center text-muted-foreground">
                              {calculating ? tr("gl_processing_financial_amounts", "กำลังประมวลผลยอดงบการเงิน...") : tr("gl_press_calculate_display_to_process", "กดปุ่ม 'คำนวณและแสดงผล' เพื่อประมวลผลยอดบัญชี")}
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="rounded-2xl border border-dashed border-border p-12 text-center text-muted-foreground">
              <FileSpreadsheet className="mx-auto mb-3 h-10 w-10 text-muted-foreground/50" />
              <h3 className="text-base font-semibold text-foreground">{tr("gl_select_template_to_start", "เลือกแม่แบบเพื่อเริ่มออกแบบ")}</h3>
              <p className="mt-1 text-sm">{tr("gl_select_template_left_or_create", "เลือกแม่แบบจากแถบด้านซ้าย หรือกดสร้างแม่แบบใหม่ / ใช้แม่แบบมาตรฐาน")}</p>
              <div className="mt-4 flex justify-center gap-2">
                <Button type="button" className={actionClass} onClick={() => void open()}>
                  <Plus className="mr-1.5 h-4 w-4" /> {tr("gl_create_new_template", "สร้างแม่แบบใหม่")}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  className={actionClass}
                  onClick={() => {
                    const fresh = emptyStatementTemplate();
                    setTemplate(fresh);
                    setStarterModalOpen(true);
                  }}
                >
                  <Sparkles className="mr-1.5 h-4 w-4 text-primary" /> {tr("gl_use_standard_template_2", "ใช้แม่แบบมาตรฐาน")}
                </Button>
              </div>
            </div>
          )
        }
      />

      {/* Starter Templates Modal */}
      {starterModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-xs">
          <div className="w-full max-w-2xl rounded-2xl border border-border bg-card p-6 shadow-xl">
            <h2 className="text-lg font-bold text-foreground">{tr("gl_select_standard_template", "เลือกแม่แบบมาตรฐาน (Starter Templates)")}</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {tr("gl_ready_made_template_dbd_fap", "แม่แบบสำเร็จรูปที่ออกแบบตามมาตรฐานกรมพัฒนาธุรกิจการค้า (DBD) และสภาวิชาชีพบัญชี")}
            </p>

            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              {generateStarterTemplates().map((starter) => (
                <div
                  key={starter.code}
                  className="flex flex-col justify-between rounded-xl border border-border bg-muted/20 p-4 transition hover:border-primary hover:bg-primary/5"
                >
                  <div>
                    <span className="rounded-md border border-border bg-background px-2 py-0.5 text-xs font-semibold text-primary">
                      {starter.code}
                    </span>
                    <h3 className="mt-2 text-base font-semibold leading-snug">{starter.name}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {labelText(statementTypeLabels, starter.statementtype, tr)} · {starter.rows.length} {tr("gl_line", tr("gl_line", "บรรทัด"))}
                    </p>
                  </div>
                  <Button
                    type="button"
                    className="mt-4 w-full"
                    onClick={() => applyStarterTemplate(starter)}
                  >
                    {tr("gl_use_this_template", "ใช้แม่แบบนี้")}
                  </Button>
                </div>
              ))}
            </div>

            <div className="mt-6 flex justify-end">
              <Button type="button" variant="outline" onClick={() => setStarterModalOpen(false)}>
                {tr("gl_cancel", "ยกเลิก")}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Full-screen Account Search Dialog for Statement Row */}
      {accountPickerRowId && (
        <AccountSearchDialog
          open={Boolean(accountPickerRowId)}
          onClose={() => setAccountPickerRowId(null)}
          accounts={refs.accounts}
          multiSelect={true}
          selectedCodes={template?.rows.find((r) => r.id === accountPickerRowId)?.accountcodes ?? []}
          onSelectMultiple={(codes) => {
            updateRow(accountPickerRowId, { accountcodes: codes });
          }}
          title={tr("gl_select_coa_for", "เลือกผังบัญชีสำหรับ \"{0}\"").replace("{0}", String(template?.rows.find((r) => r.id === accountPickerRowId)?.title || tr("gl_this_row", "แถวนี้")))}
          all={true}
        />
      )}

      {confirmationDialog}
    </div>
  );
}

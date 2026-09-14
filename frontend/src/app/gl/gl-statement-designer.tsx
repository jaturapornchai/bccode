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
import {
  formatAmount,
  emptyStatementTemplate,
  statementTypeLabels,
  statementRowTypeLabels,
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
} from "./gl-common";
import { fetchReport, type ReportFilters, emptyReportFilters } from "./gl-reports";

const FONT_OPTIONS = [
  { id: "sarabun", name: "Sarabun (สารบรรณ - มาตรฐานทางการ)", family: '"Sarabun", sans-serif', href: "https://fonts.googleapis.com/css2?family=Sarabun:wght@400;500;600;700&display=swap" },
  { id: "prompt", name: "Prompt (พร้อม - อ่านง่าย ผู้บริหาร)", family: '"Prompt", sans-serif', href: "https://fonts.googleapis.com/css2?family=Prompt:wght@400;500;600;700&display=swap" },
  { id: "kanit", name: "Kanit (คณิต - คมชัด ทันสมัย)", family: '"Kanit", sans-serif', href: "https://fonts.googleapis.com/css2?family=Kanit:wght@400;500;600;700&display=swap" },
  { id: "noto-sans-thai", name: "Noto Sans Thai (มาตรฐานสากล)", family: '"Noto Sans Thai", sans-serif', href: "https://fonts.googleapis.com/css2?family=Noto+Sans+Thai:wght@400;500;600;700&display=swap" },
  { id: "inter", name: "Inter (สากล โมเดิร์น)", family: '"Inter", sans-serif', href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" },
  { id: "monospace", name: "Courier / Monospace (ตัวเลขพิมพ์ดีด)", family: "ui-monospace, monospace", href: "" },
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
  const { confirm, confirmationDialog } = useConfirmDialog();

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
    if (dirty && !await confirm({ title: "ละทิ้งข้อมูลที่ยังไม่บันทึก?", description: "ข้อมูลที่กำลังแก้ไขจะไม่ถูกบันทึก", tone: "warning", confirmLabel: "ละทิ้งการแก้ไข" })) return;
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
    setMessage(`โหลดแม่แบบ "${starter.name}" เรียบร้อยแล้ว`);
  }

  async function save() {
    if (!template || busy) return;
    if (!template.code.trim() || !template.name.trim()) {
      setError("กรุณาระบุรหัสและชื่อแม่แบบงบการเงิน");
      return;
    }
    if (template.id && !await confirm({ title: "บันทึกการแก้ไขแม่แบบงบ?", description: `แก้ไข ${template.code} (${template.name})`, confirmLabel: "บันทึกการแก้ไข", tone: "info" })) return;
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
      setMessage("บันทึกแม่แบบงบการเงินเรียบร้อยแล้ว");
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
      name: `${template.name} (คัดลอก)`,
    };
    setTemplate(cloned);
    setOriginal("");
    setMessage(`คัดลอกแม่แบบเป็น "${cloned.name}" เรียบร้อยแล้ว (กรุณากดบันทึก)`);
  }

  async function remove() {
    if (!template?.id || busy) return;
    if (!await confirm({
      title: "ลบแม่แบบงบการเงิน?",
      description: `คุณต้องการลบแม่แบบ ${template.code} (${template.name}) ใช่หรือไม่?`,
      confirmLabel: "ลบแม่แบบ",
      tone: "danger",
    })) return;
    try {
      setError("");
      await execute({
        resource: "statement-templates",
        id: template.id,
        version: template.version,
        action: "delete",
        reason: "ลบแม่แบบงบการเงิน",
      });
      setTemplate(null);
      setOriginal("");
      list.reload();
      setMessage("ลบแม่แบบงบการเงินเรียบร้อยแล้ว");
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
      title: type === "blank" ? "" : type === "divider" ? "—" : `รายการที่ ${nextRowNo}`,
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
      setPreviewError("กรุณาเลือกปีบัญชีก่อนประมวลผล");
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
              <h2 className="text-base font-semibold">แม่แบบงบการเงิน</h2>
              <Button type="button" className={actionClass} onClick={() => void open()} disabled={busy}>
                <Plus className="mr-1.5 h-4 w-4" /> สร้างแม่แบบใหม่
              </Button>
            </div>

            <form className="flex flex-wrap gap-2 shrink-0" onSubmit={(e) => { e.preventDefault(); searchDebounce.searchNow(); }}>
              <SearchInput
                className="min-w-28 flex-1"
                ariaLabel="ค้นหารหัสหรือชื่อแม่แบบ"
                placeholder="ค้นหารหัสหรือชื่อแม่แบบ..."
                value={searchDebounce.query}
                onChange={searchDebounce.setQuery}
                onClear={searchDebounce.clear}
                onSearch={searchDebounce.searchNow}
              />
              <Button type="submit" variant="outline" className={actionClass}>
                <Search className="h-4 w-4 mr-1.5" />
                ค้นหา
              </Button>
              <Button type="button" variant="outline" className={actionClass} onClick={() => list.reload()} disabled={list.loading}>
                <RefreshCw className="h-4 w-4" />
              </Button>
            </form>

            <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border" aria-busy={list.loading}>
              <table className={`w-full text-left text-[0.95rem] leading-normal ${density.tableClass}`}>
                <thead className="sticky top-0 bg-muted z-10">
                  <tr>
                    <th className="p-2.5">รหัส</th>
                    <th className="p-2.5 min-w-44">ชื่อแม่แบบงบ</th>
                    <th className="p-2.5 text-center w-28">ประเภท</th>
                    <th className="p-2.5 text-right pr-3 w-20">จัดการ</th>
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
                            ? "bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 ring-1 ring-inset ring-amber-500/50 font-medium"
                            : isSelected
                              ? "bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium"
                              : index % 2 === 0
                                ? "bg-background hover:bg-accent/60"
                                : "bg-muted/20 hover:bg-accent/60"
                        }`}
                        onClick={() => void open(item, "preview")}
                      >
                        <td className="p-2 whitespace-nowrap">
                          <span className={`font-mono font-bold ${isRowEditing ? "text-amber-900 dark:text-amber-300" : "text-primary"}`}>{item.code}</span>
                        </td>
                      <td className="max-w-48 truncate p-2" title={item.name}>
                        {item.name}
                      </td>
                      <td className="whitespace-nowrap p-2 text-center text-xs">
                        <span className="rounded-md border border-border bg-muted/60 px-2 py-0.5 font-medium">
                          {statementTypeLabels[item.statementtype] ?? item.statementtype}
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
                            aria-label="แก้ไข"
                            title="แก้ไขผังงบ (Edit)"
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
                        {list.loading ? "กำลังโหลดข้อมูล..." : "ยังไม่มีแม่แบบงบการเงิน กดสร้างใหม่หรือใช้แม่แบบมาตรฐาน"}
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
                      <SlidersHorizontal className="mr-1.5 h-4 w-4" /> โหมดออกแบบผัง
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
                      <Eye className="mr-1.5 h-4 w-4" /> พรีวิวและพิมพ์งบจริง
                    </Button>
                  </div>

                  <Button
                    type="button"
                    variant="outline"
                    className={actionClass}
                    onClick={() => setStarterModalOpen(true)}
                  >
                    <Sparkles className="mr-1.5 h-4 w-4 text-amber-500" /> ใช้แม่แบบมาตรฐาน...
                  </Button>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  {template.id && (
                    <>
                      <Button type="button" variant="outline" className={actionClass} onClick={() => void cloneTemplate()}>
                        <Copy className="mr-1.5 h-4 w-4" /> คัดลอก
                      </Button>
                      <Button type="button" variant="outline" className={`${actionClass} text-destructive hover:bg-destructive/10`} onClick={() => void remove()} disabled={busy}>
                        <Trash2 className="mr-1.5 h-4 w-4" /> ลบ
                      </Button>
                    </>
                  )}
                  {activeTab === "design" ? (
                    <Button type="button" className={actionClass} onClick={() => void save()} disabled={busy}>
                      <Save className="mr-1.5 h-4 w-4" /> {busy ? "กำลังบันทึก..." : "บันทึกแม่แบบ"}
                    </Button>
                  ) : (
                    <Button type="button" className={actionClass} onClick={() => setActiveTab("design")} disabled={busy}>
                      <SlidersHorizontal className="mr-1.5 h-4 w-4" /> แก้ไขผังงบ
                    </Button>
                  )}
                </div>
              </div>

              {/* Template Metadata & Global Styling */}
              <div className="grid gap-3 rounded-xl border border-border bg-muted/20 p-3 sm:grid-cols-2 xl:grid-cols-4 shrink-0">
                <Field label="รหัสแม่แบบงบ">
                  <input
                    className={control}
                    value={template.code}
                    onChange={(e) => setTemplate({ ...template, code: e.target.value })}
                    placeholder="เช่น BS-01, PNL-01"
                  />
                </Field>
                <Field label="ชื่อแม่แบบงบการเงิน">
                  <input
                    className={control}
                    value={template.name}
                    onChange={(e) => setTemplate({ ...template, name: e.target.value })}
                    placeholder="เช่น งบแสดงฐานะการเงิน (แบบ DBD)"
                  />
                </Field>
                <Field label="ประเภทงบ">
                  <select
                    className={control}
                    value={template.statementtype}
                    onChange={(e) => setTemplate({ ...template, statementtype: e.target.value as StatementType })}
                  >
                    {Object.entries(statementTypeLabels).map(([key, label]) => (
                      <option key={key} value={key}>{label}</option>
                    ))}
                  </select>
                </Field>

                {/* Font Customizer */}
                <Field label="แบบตัวอักษร (Font Family)">
                  <select
                    className={control}
                    value={template.globalstyle?.fontfamily ?? "sarabun"}
                    onChange={(e) => updateGlobalStyle({ fontfamily: e.target.value })}
                  >
                    {FONT_OPTIONS.map((font) => (
                      <option key={font.id} value={font.id}>{font.name}</option>
                    ))}
                  </select>
                </Field>

                <Field label="ขนาดตัวอักษรพื้นฐาน">
                  <select
                    className={control}
                    value={template.globalstyle?.fontsize ?? "15px"}
                    onChange={(e) => updateGlobalStyle({ fontsize: e.target.value })}
                  >
                    <option value="13px">13px - กะทัดรัด</option>
                    <option value="14px">14px - ปกติ</option>
                    <option value="15px">15px - สบายตา (แนะนำ 40+)</option>
                    <option value="16px">16px - ตัวใหญ่</option>
                    <option value="18px">18px - พิเศษ</option>
                  </select>
                </Field>

                <div className="flex flex-wrap items-end gap-3 xl:col-span-2">
                  <label className="flex items-center gap-2 text-sm font-medium">
                    <input
                      type="checkbox"
                      checked={template.globalstyle?.shownotecolumn ?? true}
                      onChange={(e) => updateGlobalStyle({ shownotecolumn: e.target.checked })}
                      className="size-4 rounded"
                    />
                    แสดงคอลัมน์หมายเหตุประกอบงบ
                  </label>
                  <label className="flex items-center gap-2 text-sm font-medium">
                    <input
                      type="checkbox"
                      checked={template.isactive}
                      onChange={(e) => setTemplate({ ...template, isactive: e.target.checked })}
                      className="size-4 rounded"
                    />
                    เปิดใช้งานแม่แบบนี้
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
                        <Plus className="mr-1 h-3.5 w-3.5" /> หัวข้อ
                      </Button>
                      <Button type="button" size="sm" variant="default" className={actionClass} onClick={() => addRow("account")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> แถวบัญชี
                      </Button>
                      <Button type="button" size="sm" variant="outline" className={actionClass} onClick={() => addRow("formula")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> แถวสูตร
                      </Button>
                      <Button type="button" size="sm" variant="outline" className={actionClass} onClick={() => addRow("subtotal")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> แถวรวมย่อย
                      </Button>
                      <Button type="button" size="sm" variant="ghost" className={actionClass} onClick={() => addRow("blank")}>
                        <Plus className="mr-1 h-3.5 w-3.5" /> บรรทัดว่าง
                      </Button>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      ทั้งหมด {template.rows.length} บรรทัด
                    </span>
                  </div>

                  {/* Rows Table */}
                  <div className="flex-1 min-h-[300px] overflow-auto rounded-xl border border-border">
                    <table className="w-full text-left text-sm">
                      <thead className="sticky top-0 bg-muted font-semibold">
                        <tr>
                          <th className="w-16 p-2 text-center">ลำดับ</th>
                          <th className="w-24 p-2">ประเภท</th>
                          <th className="p-2">ชื่อรายการในงบ</th>
                          <th className="w-16 p-2 text-center">หมายเหตุ</th>
                          <th className="p-2">การผูกบัญชี / สูตรคำนวณ</th>
                          <th className="w-48 p-2">การจัดสไตล์</th>
                          <th className="w-24 p-2 text-center">จัดการ</th>
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
                                  <option key={k} value={k}>{v}</option>
                                ))}
                              </select>
                            </td>

                            {/* Title */}
                            <td className="p-2">
                              {row.rowtype === "blank" ? (
                                <span className="text-xs italic text-muted-foreground">(บรรทัดว่าง)</span>
                              ) : row.rowtype === "divider" ? (
                                <span className="text-xs font-mono text-muted-foreground">────────────────</span>
                              ) : (
                                <input
                                  className="w-full rounded border border-input bg-background px-2 py-1 text-sm font-medium"
                                  value={row.title}
                                  onChange={(e) => updateRow(row.id, { title: e.target.value })}
                                  placeholder="ระบุชื่อรายการ..."
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
                                  placeholder="เช่น 3"
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
                                      ? `เลือกแล้ว ${row.accountcodes.length} บัญชี`
                                      : "+ เลือกผังบัญชี"}
                                  </button>

                                  <select
                                    className="rounded border border-input bg-background p-1 text-xs"
                                    value={row.normalbalance ?? "debit"}
                                    onChange={(e) => updateRow(row.id, { normalbalance: e.target.value as "debit" | "credit" | "net" })}
                                  >
                                    <option value="debit">เดบิต (+)</option>
                                    <option value="credit">เครดิต (+)</option>
                                    <option value="net">สุทธิ</option>
                                  </select>

                                  <label className="flex items-center gap-1 text-xs text-muted-foreground" title="กลับเครื่องหมายบวกลบ">
                                    <input
                                      type="checkbox"
                                      checked={row.reversesign ?? false}
                                      onChange={(e) => updateRow(row.id, { reversesign: e.target.checked })}
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
                                    placeholder={row.rowtype === "subtotal" ? "เช่น SUM(R10:R40)" : "เช่น R10 + R20 - R30"}
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
                                  title="ตัวหนา"
                                >
                                  B
                                </button>
                                {/* Italic */}
                                <button
                                  type="button"
                                  className={`h-7 w-7 rounded border italic text-xs ${row.style?.fontstyle === "italic" ? "bg-primary text-primary-foreground border-primary" : "border-input bg-background"}`}
                                  onClick={() => updateRow(row.id, { style: { ...row.style, fontstyle: row.style?.fontstyle === "italic" ? "normal" : "italic" } })}
                                  title="ตัวเอียง"
                                >
                                  I
                                </button>

                                {/* Indent */}
                                <select
                                  className="h-7 rounded border border-input bg-background px-1 text-xs"
                                  value={row.style?.indent ?? 0}
                                  onChange={(e) => updateRow(row.id, { style: { ...row.style, indent: parseInt(e.target.value, 10) } })}
                                  title="ระดับการเยื้อง"
                                >
                                  <option value={0}>ไม่เยื้อง</option>
                                  <option value={1}>เยื้อง 1</option>
                                  <option value={2}>เยื้อง 2</option>
                                  <option value={3}>เยื้อง 3</option>
                                  <option value={4}>เยื้อง 4</option>
                                </select>

                                {/* Underline */}
                                <select
                                  className="h-7 rounded border border-input bg-background px-1 text-xs"
                                  value={row.style?.underline ?? "none"}
                                  onChange={(e) => updateRow(row.id, { style: { ...row.style, underline: e.target.value as StatementRowUnderline } })}
                                  title="เส้นใต้บัญชี"
                                >
                                  <option value="none">ไร้เส้น</option>
                                  <option value="single">ขีดเดี่ยว _</option>
                                  <option value="double">ขีดคู่ = (ยอดสุทธิ)</option>
                                  <option value="top_single_bottom_double">บนเดี่ยว ล่างคู่</option>
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
                                  title="เลื่อนขึ้น"
                                >
                                  <ArrowUp className="h-3.5 w-3.5 mx-auto" />
                                </button>
                                <button
                                  type="button"
                                  className="h-6 w-6 rounded hover:bg-muted"
                                  disabled={index === template.rows.length - 1}
                                  onClick={() => moveRow(index, "down")}
                                  title="เลื่อนลง"
                                >
                                  <ArrowDown className="h-3.5 w-3.5 mx-auto" />
                                </button>
                                <button
                                  type="button"
                                  className="h-6 w-6 rounded text-destructive hover:bg-destructive/10"
                                  onClick={() => deleteRow(row.id)}
                                  title="ลบแถวนี้"
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
                        <Field label="ปีบัญชี">
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
                        <Field label="ตั้งแต่วันที่">
                          <input type="date" className={control} value={filters.from} onChange={(e) => setFilters((prev) => ({ ...prev, from: e.target.value }))} />
                        </Field>
                      </div>
                      <div className="w-36">
                        <Field label="ถึงวันที่">
                          <input type="date" className={control} value={filters.to} onChange={(e) => setFilters((prev) => ({ ...prev, to: e.target.value }))} />
                        </Field>
                      </div>
                      <Button type="button" className={`${actionClass} mt-auto`} onClick={() => void runCalculation()} disabled={calculating}>
                        <RefreshCw className={`mr-1.5 h-4 w-4 ${calculating ? "animate-spin" : ""}`} />
                        {calculating ? "กำลังคำนวณ..." : "คำนวณและแสดงผล"}
                      </Button>
                    </div>

                    <div className="flex items-center gap-2">
                      <Button type="button" variant="outline" className={actionClass} onClick={() => window.print()}>
                        <Printer className="mr-1.5 h-4 w-4" /> พิมพ์งบการเงิน
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        className={actionClass}
                        onClick={() => {
                          if (!calculated) return;
                          const csv = "\uFEFF" + [
                            ["ลำดับ", "รายการ", "หมายเหตุ", "จำนวนเงิน"].join(","),
                            ...calculated.rows.map((r) => [`"'${r.rowno}"`, `"${r.title.replace(/"/g, '""')}"`, `"${r.noteno ?? ""}"`, `"'${r.amountFormatted}"`].join(",")),
                          ].join("\r\n");
                          downloadText(`งบการเงิน-${template.code}-${filters.fiscalyear}.csv`, csv, "text/csv;charset=utf-8");
                        }}
                      >
                        <Download className="mr-1.5 h-4 w-4" /> ส่งออก CSV
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
                      <p className="text-sm text-muted-foreground">สำหรับงวดบัญชี {filters.fiscalyear} (ณ วันที่ {filters.to || "-"})</p>
                      <p className="text-xs text-muted-foreground">(หน่วย: บาท)</p>
                    </div>

                    {/* Statement Table */}
                    <table className="w-full border-collapse">
                      <thead>
                        <tr className="border-b-2 border-foreground/30">
                          <th className="py-2 text-left font-bold">รายการ</th>
                          {template.globalstyle?.shownotecolumn && (
                            <th className="w-24 py-2 text-center font-bold">หมายเหตุ</th>
                          )}
                          <th className="w-44 py-2 text-right font-bold">ยอดเงิน</th>
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
                              {calculating ? "กำลังประมวลผลยอดงบการเงิน..." : "กดปุ่ม 'คำนวณและแสดงผล' เพื่อประมวลผลยอดบัญชี"}
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
              <h3 className="text-base font-semibold text-foreground">เลือกแม่แบบเพื่อเริ่มออกแบบ</h3>
              <p className="mt-1 text-sm">เลือกแม่แบบจากแถบด้านซ้าย หรือกดสร้างแม่แบบใหม่ / ใช้แม่แบบมาตรฐาน</p>
              <div className="mt-4 flex justify-center gap-2">
                <Button type="button" className={actionClass} onClick={() => void open()}>
                  <Plus className="mr-1.5 h-4 w-4" /> สร้างแม่แบบใหม่
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
                  <Sparkles className="mr-1.5 h-4 w-4 text-amber-500" /> ใช้แม่แบบมาตรฐาน
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
            <h2 className="text-lg font-bold text-foreground">เลือกแม่แบบมาตรฐาน (Starter Templates)</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              แม่แบบสำเร็จรูปที่ออกแบบตามมาตรฐานกรมพัฒนาธุรกิจการค้า (DBD) และสภาวิชาชีพบัญชี
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
                      {statementTypeLabels[starter.statementtype]} · {starter.rows.length} บรรทัด
                    </p>
                  </div>
                  <Button
                    type="button"
                    className="mt-4 w-full"
                    onClick={() => applyStarterTemplate(starter)}
                  >
                    ใช้แม่แบบนี้
                  </Button>
                </div>
              ))}
            </div>

            <div className="mt-6 flex justify-end">
              <Button type="button" variant="outline" onClick={() => setStarterModalOpen(false)}>
                ยกเลิก
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
          title={`เลือกผังบัญชีสำหรับ "${template?.rows.find((r) => r.id === accountPickerRowId)?.title || "แถวนี้"}"`}
          all={true}
        />
      )}

      {confirmationDialog}
    </div>
  );
}

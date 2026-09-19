"use client";

import { useBackendDictionary, useBackendText, type BackendTextFn } from "@/components/backend-text-provider";

import { authFetch } from "@/lib/client-auth-session";
import React, { useCallback, useMemo, useState, useEffect, useRef } from "react";
import {
  Building2,
  GitBranch,
  Plus,
  ChevronRight,
  ChevronDown,
  Edit3,
  Trash2,
  Star,
  Save,
  Loader2,
  Check,
  KeyRound,
  UploadCloud,
  FileText,
  X,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChoiceSelect } from "@/components/ui/select";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { Checkbox, CheckboxCard } from "@/components/ui/checkbox";

import { LogoAvatar } from "@/components/logo-avatar";
import { type LanguageCode, LANGUAGES } from "@/lib/i18n";
import { DEFAULT_TIME_ZONE, timezoneMeta, timezoneSelectOptions } from "@/lib/date-time";
import { cn } from "@/lib/utils";
import {
  normalizeLanguageConfigs,
  ThailandAddressSelect,
  filterThailandProvinces,
  filterThailandDistricts,
  filterThailandSubdistricts,
  singleThailandAddressCode,
  thailandAddressOptionLabel,
  postalAddressHint,
  thailandAddressUi,
} from "./system-settings-screen";
import {
  loadThailandAddressData,
  getThailandCountry,
  findThailandProvince,
  findThailandDistrict,
  findThailandSubdistrict,
  findThailandAddressMatchesByPostalCode,
  getThailandDistrictPostalCode,
  getThailandSubdistrictPostalCode,
  normalizeThaiPostalCode,
  type ThailandAddressData,
} from "@/lib/thailand-addresses";
import { deriveMainApiUrl } from "@/lib/backend-url";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { AddressesEditor } from "@/components/product-barcode/addresses-editor";
import { isThaiHeadOfficeBranchCode, normalizeThaiTaxBranchCode } from "@/lib/thai-branch-code";
import { notifyWorkspaceChanged } from "@/lib/workspace-models";
import { normalizeBusinessCode } from "@/lib/business-code";

interface CompanyBranchTreeViewProps {
  auth: { token: string; backendUrl: string; profile?: { email?: string } | null } | null;
  workspace: CompanyWorkspace | null;
  language: LanguageCode;
  onRefresh?: () => void;
}

interface CompanyWorkspace {
  shop: { holdingcode: string; role?: number };
  shopInfo?: {
    settings?: {
      language?: string;
      languageconfigs?: unknown;
    };
  } | null;
}

interface LocalizedNameEntry {
  code?: string;
  name?: string;
}

type LocalizedNames = LocalizedNameEntry[] | Record<string, unknown> | null | undefined;

interface CompanyRecord {
  guidfixed?: string;
  holdinguid?: string;
  companyuid?: string;
  code?: string;
  logouri?: string;
  names?: LocalizedNames;
  taxid?: string;
  isactive?: boolean;
  isdeleted?: boolean;
  deletedat?: string | null;
}

interface BranchRecord {
  guidfixed?: string;
  holdinguid?: string;
  companyuid?: string;
  branchuid?: string;
  companyguid?: string;
  code?: string;
  logouri?: string;
  names?: LocalizedNames;
  timezone?: string;
  timezonelabel?: string;
  timezoneoffset?: string;
  language?: string;
  dateformat?: string;
  yeartype?: string;
  branchtype?: string;
  isvatregistered?: boolean;
  companyregistrationno?: string;
  email?: string;
  managername?: string;
  fiscalstartmonth?: number;
  documentformats?: Partial<DocFormat>[];
  addresses?: { code?: string; address?: string }[];
  countrycode?: string;
  provincecode?: string;
  districtcode?: string;
  subdistrictcode?: string;
  zipcode?: string;
  etaxenabled?: boolean;
  isactive?: boolean;
  isdeleted?: boolean;
  deletedat?: string | null;
}

interface OrganizationSaveResponse {
  success?: boolean;
  message?: string;
  data?: {
    entity?: CompanyRecord | BranchRecord;
    kafka_sync?: string;
  };
}

const MONTH_OPTIONS = [
  { value: 1, label: ["month_january", "มกราคม"] },
  { value: 2, label: ["month_february", "กุมภาพันธ์"] },
  { value: 3, label: ["month_march", "มีนาคม"] },
  { value: 4, label: ["month_april", "เมษายน"] },
  { value: 5, label: ["month_may", "พฤษภาคม"] },
  { value: 6, label: ["month_june", "มิถุนายน"] },
  { value: 7, label: ["month_july", "กรกฎาคม"] },
  { value: 8, label: ["month_august", "สิงหาคม"] },
  { value: 9, label: ["month_september", "กันยายน"] },
  { value: 10, label: ["month_october", "ตุลาคม"] },
  { value: 11, label: ["month_november", "พฤศจิกายน"] },
  { value: 12, label: ["month_december", "ธันวาคม"] },
] as const;

// ภ.พ.20 branch registration type.
const BRANCH_TYPE_OPTIONS = [
  { value: "head", label: ["head_of", "สำนักงานใหญ่"] },
  { value: "permanent", label: ["st_permanent_branch", "สาขาถาวร"] },
  { value: "temporary", label: ["st_temporary_branch", "สาขาชั่วคราว"] },
] as const;

const YEAR_TYPE_OPTIONS = [
  { value: "buddhist", label: ["st_buddhist_era", "พ.ศ. (Buddhist Era)"] },
  { value: "christian", label: ["st_christian_era", "ค.ศ. (Christian Era)"] },
] as const;

// ตัวอย่างปีในรูปแบบวันที่เปลี่ยนตามประเภทปี: พ.ศ. -> 2568, ค.ศ. -> 2025
function dateFormatOptionsFor(yearType: string) {
  const y = yearType === "buddhist" ? 2568 : 2025;
  return [
    { value: "dd/MM/yyyy", label: `dd/MM/yyyy (31/12/${y})` },
    { value: "dd-MM-yyyy", label: `dd-MM-yyyy (31-12-${y})` },
    { value: "yyyy-MM-dd", label: `yyyy-MM-dd (${y}-12-31)` },
    { value: "MM/dd/yyyy", label: `MM/dd/yyyy (12/31/${y})` },
  ];
}


// ประเภทเอกสารหลักที่สาขากำหนดคำนำหน้าเลขที่เอกสารแยกได้ (subset จาก ~46 types ในระบบ).
// code ตรงกับ MODULE_NAME ของ transaction module ใน backend (เก็บค่าอย่างเดียว — generator ยังไม่ใช้).
const DOC_PREFIX_TYPES = [
  { code: "SI", label: ["st_tax_invoice_receipt", "ใบกำกับภาษี / ใบเสร็จ"] },
  { code: "ST", label: ["st_credit_note_sales", "ใบลดหนี้ (ขาย)"] },
  { code: "SA", label: ["st_debit_note_sales", "ใบเพิ่มหนี้ (ขาย)"] },
  { code: "SO", label: ["transaction_sale_order", "ใบสั่งขาย/สั่งจองสินค้า"] },
  { code: "QT", label: ["transaction_quotation", "ใบเสนอราคา"] },
  { code: "PU", label: ["st_goods_receipt_purchase", "ใบรับสินค้า (ซื้อ)"] },
  { code: "PO", label: ["transaction_purchase_order", "ใบสั่งซื้อสินค้า"] },
  { code: "PT", label: ["st_purchase_return", "ใบรับคืน (ซื้อ)"] },
  { code: "TF", label: ["st_interbranch_stock_transfer", "ใบโอนสินค้าระหว่างสาขา"] },
  { code: "AJ", label: ["st_stock_adjustment", "ใบปรับปรุงสต็อก"] },
  { code: "EE", label: ["st_receipt_voucher", "ใบสำคัญรับเงิน"] },
  { code: "DE", label: ["pdf_payment_voucher", "ใบสำคัญจ่าย"] },
  { code: "PC", label: ["st_petty_cash_deposit", "เงินสดย่อย / มัดจำ"] },
] as const;

// รูปแบบเลขที่เอกสาร 1 รูปแบบ (แต่ละประเภทมีได้หลายรูปแบบ; generator จริงยังไม่ใช้ค่านี้).
type DocFormat = {
  doctype: string;
  name: string;
  prefix: string;
  usebranch: boolean;
  yearmode: string; // none | be2 | be4 | ce2 | ce4
  usemonth: boolean;
  useday: boolean;
  separator: boolean;
  runlength: number; // 4-6
  resetmode: string; // never | yearly | monthly | daily
  startnumber: number;
  enabled: boolean;
  isdefault: boolean;
};

const DOC_YEAR_MODES = [
  { value: "none", label: ["st_not_use", "ไม่ใช้"] },
  { value: "be2", label: ["st_be_2_digit", "พ.ศ. 2 หลัก"] },
  { value: "be4", label: ["st_be_4_digit", "พ.ศ. 4 หลัก"] },
  { value: "ce2", label: ["st_ce_2_digit", "ค.ศ. 2 หลัก"] },
  { value: "ce4", label: ["st_ce_4_digit", "ค.ศ. 4 หลัก"] },
] as const;
const DOC_RESET_MODES = [
  { value: "never", label: ["st_no_reset", "ไม่รีเซ็ต"] },
  { value: "yearly", label: ["st_yearly", "รายปี"] },
  { value: "monthly", label: ["alert_monthly", "รายเดือน"] },
  { value: "daily", label: ["st_daily", "รายวัน"] },
] as const;
// เอกสารภาษี — default รีเซ็ตรายปี (ตามแนวสรรพากร)
const TAX_DOC_TYPES = ["SI", "ST", "SA"];

const normalizeDocPrefix = (v: string) => v.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 10);

function defaultDocFormat(doctype: string): DocFormat {
  return {
    doctype,
    name: "",
    prefix: doctype,
    usebranch: false,
    yearmode: "ce2",
    usemonth: true,
    useday: true,
    separator: false,
    runlength: 5,
    resetmode: TAX_DOC_TYPES.includes(doctype) ? "yearly" : "never",
    startnumber: 1,
    enabled: true,
    isdefault: false,
  };
}

// สร้างตัวอย่างเลขที่เอกสารจากรูปแบบ (ใช้วันที่ปัจจุบันสำหรับ preview)
function buildDocExample(f: DocFormat, branchCode: string): string {
  const p = normalizeDocPrefix(f.prefix) || f.doctype || "XX";
  const now = new Date();
  const ce = now.getFullYear();
  const be = ce + 543;
  const year =
    f.yearmode === "be2" ? String(be % 100).padStart(2, "0")
    : f.yearmode === "be4" ? String(be)
    : f.yearmode === "ce2" ? String(ce % 100).padStart(2, "0")
    : f.yearmode === "ce4" ? String(ce)
    : "";
  const br = f.usebranch ? (branchCode || "00000") : "";
  const mm = f.usemonth ? String(now.getMonth() + 1).padStart(2, "0") : "";
  const dd = f.useday ? String(now.getDate()).padStart(2, "0") : "";
  const date = year + mm + dd;
  const num = String(Math.max(1, f.startnumber || 1)).padStart(Math.max(1, f.runlength || 5), "0");
  return [p, br, date, num].filter(Boolean).join(f.separator ? "-" : "");
}

// คำนำหน้าที่ซ้ำกันทั้งสาขา (ข้ามทุกประเภท) — คืน set ของ prefix ที่ซ้ำ
function duplicateDocPrefixes(formats: DocFormat[]): Set<string> {
  const counts = new Map<string, number>();
  for (const f of formats) {
    const p = normalizeDocPrefix(f.prefix);
    if (p) counts.set(p, (counts.get(p) ?? 0) + 1);
  }
  return new Set([...counts.entries()].filter(([, n]) => n > 1).map(([p]) => p));
}

function DocSelect({ label, value, onChange, options, disabled }: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: [string, string][];
  disabled: boolean;
}) {
  return (
    <label className="flex flex-col gap-0.5 text-[11px] text-muted-foreground">
      {label}
      <select
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value)}
        className="h-8 w-full rounded-md border border-border bg-background px-2 text-sm"
      >
        {options.map(([v, t]) => <option key={v} value={v}>{t}</option>)}
      </select>
    </label>
  );
}

function DocFormatBuilder({ formats, onChange, branchCode, disabled }: {
  formats: DocFormat[];
  onChange: (next: DocFormat[]) => void;
  branchCode: string;
  disabled: boolean;
}) {
  const tr = useBackendText();
  const dups = duplicateDocPrefixes(formats);
  const update = (i: number, patch: Partial<DocFormat>) =>
    onChange(formats.map((f, idx) => (idx === i ? { ...f, ...patch } : f)));
  const addFormat = (doctype: string) =>
    onChange([...formats, { ...defaultDocFormat(doctype), prefix: "", isdefault: !formats.some((f) => f.doctype === doctype) }]);
  const removeFormat = (i: number) => onChange(formats.filter((_, idx) => idx !== i));
  const setDefault = (i: number, doctype: string) =>
    onChange(formats.map((f, idx) => (f.doctype === doctype ? { ...f, isdefault: idx === i } : f)));

  return (
    <div className="space-y-2">
      <label className="text-sm font-semibold text-foreground">
        {tr("st_doc_num_format_multiple", "รูปแบบเลขที่เอกสาร (แต่ละประเภทมีได้หลายรูปแบบ)")}
        <span className="ml-1 font-normal text-xs text-muted-foreground">{tr("st_adjust_prefix_ymd_running_preview", "— ปรับคำนำหน้า/ปี/เดือน/วัน/รันนิ่ง เห็นตัวอย่างทันที")}</span>
      </label>
      {dups.size > 0 ? (
        <div className="rounded-md bg-destructive/10 px-3 py-1.5 text-xs text-destructive">
          {tr("st_duplicate_prefix_fix_before_save", "คำนำหน้าซ้ำ: {0} — ต้องแก้ให้ไม่ซ้ำก่อนบันทึก").replace("{0}", String([...dups].join(", ")))}
        </div>
      ) : null}
      <div className="space-y-3">
        {DOC_PREFIX_TYPES.map((dt) => {
          const rows = formats.map((f, i) => ({ f, i })).filter((r) => r.f.doctype === dt.code);
          return (
            <div key={dt.code} className="rounded-xl border border-border bg-card p-3">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-medium">
                  {tr(dt.label[0], dt.label[1])} <span className="font-mono text-xs text-muted-foreground">({dt.code})</span>
                </span>
                <Button type="button" variant="outline" size="sm" disabled={disabled} onClick={() => addFormat(dt.code)}>
                  <Plus className="h-3.5 w-3.5" /> {tr("st_add_format", "เพิ่มรูปแบบ")}
                </Button>
              </div>
              {rows.length === 0 ? (
                <p className="text-xs text-muted-foreground">{tr("st_no_format_press_add_format", "ยังไม่มีรูปแบบ — กด “เพิ่มรูปแบบ”")}</p>
              ) : (
                <div className="space-y-2">
                  {rows.map(({ f, i }) => {
                    const pfx = normalizeDocPrefix(f.prefix);
                    const dup = !!pfx && dups.has(pfx);
                    return (
                      <div key={i} className={cn("rounded-lg border p-2.5", f.isdefault ? "border-primary ring-1 ring-primary/30" : "border-border", !f.enabled && "opacity-60")}>
                        <div className="flex items-center gap-2">
                          <Checkbox checked={f.enabled} disabled={disabled} aria-label={tr("st_activate_format", "เปิดใช้งานรูปแบบ")} onCheckedChange={(checked) => update(i, { enabled: checked })} />
                          <Input value={f.name} placeholder={tr("pattern_name", "ชื่อรูปแบบ")} disabled={disabled} onChange={(e) => update(i, { name: e.target.value })} className="h-8 flex-1 text-sm" />
                          <Input value={f.prefix} placeholder={f.doctype} maxLength={10} disabled={disabled} onChange={(e) => update(i, { prefix: normalizeDocPrefix(e.target.value) })} className={cn("h-8 w-24 text-center text-sm uppercase", dup && "border-destructive text-destructive")} />
                          <button type="button" aria-label={tr("st_set_as_default", "ตั้งเป็นค่าเริ่มต้น")} title={tr("st_set_as_default", "ตั้งเป็นค่าเริ่มต้น")} disabled={disabled} onClick={() => setDefault(i, f.doctype)} className={cn("grid h-8 w-8 shrink-0 place-items-center rounded-md", f.isdefault ? "bg-primary/15 text-primary" : "text-muted-foreground hover:bg-muted")}>
                            <Star className={cn("h-4 w-4", f.isdefault && "fill-primary")} />
                          </button>
                          <span className="rounded-md bg-primary/10 px-2.5 py-1 font-mono text-sm font-medium text-primary whitespace-nowrap">{buildDocExample(f, branchCode)}</span>
                          <button type="button" aria-label={tr("st_delete_format", "ลบรูปแบบ")} disabled={disabled} onClick={() => removeFormat(i)} className="grid h-8 w-8 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-destructive">
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                        {f.isdefault ? (
                          <span className="mt-1.5 inline-flex items-center gap-1 rounded bg-primary/15 px-2 py-0.5 text-[11px] font-medium text-primary">
                            <Star className="h-3 w-3 fill-primary" /> {tr("menu_setup", "ค่าเริ่มต้น")}
                          </span>
                        ) : null}
                        <div className="mt-2 grid grid-cols-2 gap-x-3 gap-y-2 sm:grid-cols-3 md:grid-cols-4">
                          <DocSelect label={tr("branchcode", "รหัสสาขา")} value={f.usebranch ? "1" : "0"} disabled={disabled} onChange={(v) => update(i, { usebranch: v === "1" })} options={[["0", tr("st_exclude", "ไม่ใส่")], ["1", tr("st_include_value", "ใส่ {0}").replace("{0}", String(branchCode || "00000"))]]} />
                          <DocSelect label={tr("year", "ปี")} value={f.yearmode} disabled={disabled} onChange={(v) => update(i, { yearmode: v })} options={DOC_YEAR_MODES.map((y) => [y.value, tr(y.label[0], y.label[1])])} />
                          <DocSelect label={tr("st_month", "เดือน")} value={f.usemonth ? "1" : "0"} disabled={disabled} onChange={(v) => update(i, { usemonth: v === "1" })} options={[["1", tr("use", "ใช้")], ["0", tr("st_not_use", "ไม่ใช้")]]} />
                          <DocSelect label={tr("day", "วัน")} value={f.useday ? "1" : "0"} disabled={disabled} onChange={(v) => update(i, { useday: v === "1" })} options={[["1", tr("use", "ใช้")], ["0", tr("st_not_use", "ไม่ใช้")]]} />
                          <DocSelect label={tr("st_separator", "ตัวคั่น")} value={f.separator ? "1" : "0"} disabled={disabled} onChange={(v) => update(i, { separator: v === "1" })} options={[["0", tr("st_none", "ไม่มี")], ["1", tr("st_hyphen", "ขีดกลาง")]]} />
                          <DocSelect label={tr("st_running", "รันนิ่ง")} value={String(f.runlength)} disabled={disabled} onChange={(v) => update(i, { runlength: +v })} options={[["4", tr("st_4_digits", "4 หลัก")], ["5", tr("st_5_digits", "5 หลัก")], ["6", tr("st_6_digits", "6 หลัก")]]} />
                          <DocSelect label={tr("form_design_reset", "รีเซ็ต")} value={f.resetmode} disabled={disabled} onChange={(v) => update(i, { resetmode: v })} options={DOC_RESET_MODES.map((r) => [r.value, tr(r.label[0], r.label[1])])} />
                          <label className="flex flex-col gap-0.5 text-[11px] text-muted-foreground">
                            {tr("st_starts_at", "เริ่มที่")}
                            <Input type="number" min={1} value={f.startnumber} disabled={disabled} onChange={(e) => update(i, { startnumber: Math.max(1, Number(e.target.value) || 1) })} className="h-8 w-full text-sm" />
                          </label>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

type NodeType = "company" | "branch";
type OrganizationFormType = "viewcompany" | "viewbranch" | "editcompany" | "editbranch" | "createcompany" | "createbranch";

interface SelectedNode {
  type: NodeType;
  guidfixed?: string;
  companyuid?: string;
  companyguid?: string;
  data: Partial<CompanyRecord> | Partial<BranchRecord>;
}

const getNameFromObject = (names: LocalizedNames, code: string): string => {
  if (!names) return "";
  if (Array.isArray(names)) {
    const entry = names.find((n) => n.code === code);
    return entry?.name || "";
  }
  const value = names[code];
  return typeof value === "string" ? value : "";
};

const getAddressFromList = (
  list: { code?: string; address?: string }[] | null | undefined,
  code: string,
): string => {
  if (!Array.isArray(list)) return "";
  const entry = list.find((item) => item?.code === code);
  return entry?.address || "";
};

// Cascade province -> district -> subdistrict -> auto zipcode picker (+ reverse
// zipcode lookup), matching ThailandAddressFieldEditor's UX in system-settings-screen.tsx.
// Plain-useState here since branch-tree-view doesn't use the generic FormState machinery.
function BranchGeoAddressPicker({
  backendUrl,
  countryCode,
  provinceCode,
  districtCode,
  subdistrictCode,
  zipCode,
  language,
  disabled,
  onChange,
}: {
  backendUrl?: string;
  countryCode: string;
  provinceCode: string;
  districtCode: string;
  subdistrictCode: string;
  zipCode: string;
  language: LanguageCode;
  disabled?: boolean;
  onChange: (next: {
    countrycode?: string;
    provincecode?: string;
    districtcode?: string;
    subdistrictcode?: string;
    zipcode?: string;
  }) => void;
}) {
  const [data, setData] = useState<ThailandAddressData | null>(null);
  const [loadError, setLoadError] = useState("");
  const postalCode = normalizeThaiPostalCode(zipCode);

  useEffect(() => {
    if (countryCode !== "TH") return;
    let active = true;
    setLoadError("");
    loadThailandAddressData(backendUrl)
      .then((nextData) => {
        if (active) setData(nextData);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setLoadError(error instanceof Error ? error.message : "load failed");
      });
    return () => {
      active = false;
    };
  }, [backendUrl, countryCode]);

  const country = data ? getThailandCountry(data) : undefined;
  const postalMatches = useMemo(
    () =>
      data && postalCode.length === 5
        ? findThailandAddressMatchesByPostalCode(data, postalCode)
        : [],
    [data, postalCode],
  );
  const selectedProvince = data ? findThailandProvince(data, provinceCode) : undefined;
  const selectedDistrict = data
    ? findThailandDistrict(data, provinceCode, districtCode)
    : undefined;
  const selectedSubdistrict = data
    ? findThailandSubdistrict(data, provinceCode, districtCode, subdistrictCode)
    : undefined;

  const provinces = filterThailandProvinces(country?.provinces ?? [], postalMatches);
  const districts = filterThailandDistricts(
    selectedProvince?.districts ?? [],
    postalMatches,
    provinceCode,
  );
  const subdistricts = filterThailandSubdistricts(
    selectedDistrict?.subdistricts ?? [],
    postalMatches,
    provinceCode,
    districtCode,
  );

  function chooseProvince(nextProvinceCode: string) {
    const next: Parameters<typeof onChange>[0] = {
      provincecode: nextProvinceCode,
      districtcode: "",
      subdistrictcode: "",
      zipcode: "",
    };
    if (data && postalCode.length === 5 && nextProvinceCode) {
      const matches = postalMatches.filter((item) => item.provinceCode === nextProvinceCode);
      if (matches.length > 0) {
        next.zipcode = postalCode;
        const nextDistrictCode = singleThailandAddressCode(matches, (item) => item.districtCode);
        if (nextDistrictCode) {
          next.districtcode = nextDistrictCode;
          const nextSubdistrictCode = singleThailandAddressCode(
            matches.filter((item) => item.districtCode === nextDistrictCode),
            (item) => item.subdistrictCode,
          );
          if (nextSubdistrictCode) next.subdistrictcode = nextSubdistrictCode;
        }
      }
    }
    onChange(next);
  }

  function chooseDistrict(nextDistrictCode: string) {
    const next: Parameters<typeof onChange>[0] = {
      districtcode: nextDistrictCode,
      subdistrictcode: "",
      zipcode: "",
    };
    if (data && provinceCode && nextDistrictCode) {
      const matches =
        postalCode.length === 5
          ? postalMatches.filter(
              (item) => item.provinceCode === provinceCode && item.districtCode === nextDistrictCode,
            )
          : [];
      next.zipcode =
        matches.length > 0 ? postalCode : getThailandDistrictPostalCode(data, provinceCode, nextDistrictCode);
      if (postalCode.length === 5) {
        const nextSubdistrictCode = singleThailandAddressCode(matches, (item) => item.subdistrictCode);
        if (nextSubdistrictCode) next.subdistrictcode = nextSubdistrictCode;
      }
    }
    onChange(next);
  }

  function chooseSubdistrict(nextSubdistrictCode: string) {
    const next: Parameters<typeof onChange>[0] = {
      subdistrictcode: nextSubdistrictCode,
      zipcode: "",
    };
    if (data && provinceCode && districtCode) {
      next.zipcode = nextSubdistrictCode
        ? getThailandSubdistrictPostalCode(data, provinceCode, districtCode, nextSubdistrictCode)
        : getThailandDistrictPostalCode(data, provinceCode, districtCode);
    }
    onChange(next);
  }

  function applyPostalCode(rawValue: string) {
    const normalized = normalizeThaiPostalCode(rawValue);
    const next: Parameters<typeof onChange>[0] = { zipcode: normalized };
    if (data && normalized.length === 5) {
      const matches = findThailandAddressMatchesByPostalCode(data, normalized);
      const nextProvinceCode = singleThailandAddressCode(matches, (item) => item.provinceCode);
      if (nextProvinceCode) {
        next.provincecode = nextProvinceCode;
        const provinceMatches = matches.filter((item) => item.provinceCode === nextProvinceCode);
        const nextDistrictCode = singleThailandAddressCode(provinceMatches, (item) => item.districtCode);
        next.districtcode = nextDistrictCode;
        if (nextDistrictCode) {
          next.subdistrictcode = singleThailandAddressCode(
            provinceMatches.filter((item) => item.districtCode === nextDistrictCode),
            (item) => item.subdistrictCode,
          );
        } else {
          next.subdistrictcode = "";
        }
      } else {
        next.provincecode = "";
        next.districtcode = "";
        next.subdistrictcode = "";
      }
    }
    onChange(next);
  }

  const dictionary = useBackendDictionary();
  const labels = thailandAddressUi(dictionary);
  const loading = !country && !loadError;

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{labels.title}</span>
        <span className="text-xs font-medium text-muted-foreground">
          {loading ? labels.loading : loadError ? labels.loadError : `${country?.provinces.length ?? 0} ${labels.provinces}`}
        </span>
      </div>
      <div className="grid gap-2 md:grid-cols-2">
        <ThailandAddressSelect
          label={labels.province}
          value={provinceCode}
          disabled={disabled || loading || Boolean(loadError)}
          placeholder={labels.selectProvince}
          onChange={chooseProvince}
          options={provinces.map((province) => ({
            code: province.code,
            label: thailandAddressOptionLabel(province, language),
          }))}
        />
        <ThailandAddressSelect
          label={labels.district}
          value={districtCode}
          disabled={disabled || !selectedProvince || loading || Boolean(loadError)}
          placeholder={labels.selectDistrict}
          onChange={chooseDistrict}
          options={districts.map((district) => ({
            code: district.code,
            label: thailandAddressOptionLabel(district, language),
          }))}
        />
        <ThailandAddressSelect
          label={labels.subdistrict}
          value={subdistrictCode}
          disabled={disabled || !selectedDistrict || loading || Boolean(loadError)}
          placeholder={labels.selectSubdistrict}
          onChange={chooseSubdistrict}
          options={subdistricts.map((subdistrict) => ({
            code: subdistrict.code,
            label: thailandAddressOptionLabel(subdistrict, language),
          }))}
        />
        <label className="grid gap-1">
          <span>{labels.postalCode}</span>
          <Input
            inputMode="numeric"
            maxLength={5}
            value={postalCode}
            disabled={disabled}
            onChange={(event) => applyPostalCode(event.target.value)}
            placeholder="10200"
          />
        </label>
      </div>
      <p className="text-xs font-medium text-muted-foreground">
        {postalAddressHint(postalCode, postalMatches, selectedSubdistrict, language, dictionary)}
      </p>
    </section>
  );
}

const isVisibleOrganizationRecord = <T extends { isdeleted?: boolean; deletedat?: string | null }>(record: T): boolean => {
  if (record.isdeleted === true) return false;
  return !record.deletedat || String(record.deletedat).trim().length === 0;
};

const companyTreeKey = (company: CompanyRecord): string =>
  company.companyuid?.trim() || company.guidfixed?.trim() || "";

const branchParentTreeKey = (branch: BranchRecord): string =>
  branch.companyuid?.trim() || branch.companyguid?.trim() || "";

const ORG_TREE_SIDEBAR_MIN_WIDTH = 260;
const ORG_TREE_SIDEBAR_MAX_WIDTH = 620;
const ORG_TREE_SIDEBAR_DEFAULT_WIDTH = 340;
const ORG_TREE_SIDEBAR_WIDTH_STORAGE_KEY = "bc_org_tree_sidebar_width";

export function CompanyBranchTreeView({
  auth,
  workspace,
  language,
  onRefresh,
}: CompanyBranchTreeViewProps) {
  const tr = useBackendText();
  const [companies, setCompanies] = useState<CompanyRecord[]>([]);
  const [branches, setBranches] = useState<BranchRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [loadError, setLoadError] = useState("");

  const containerRef = useRef<HTMLDivElement>(null);
  const {
    splitPercent: sidebarWidth,
    setSplitPercent: setSidebarWidth,
    isResizing: isResizingSidebar,
    startResize: handleSidebarResizeStart,
    resetSplit: handleSidebarResizeReset,
    adjustWithKeyboard: handleSidebarKeyDown,
  } = useSplitPercent({
    storageKey: ORG_TREE_SIDEBAR_WIDTH_STORAGE_KEY,
    defaultLeft: ORG_TREE_SIDEBAR_DEFAULT_WIDTH,
    min: ORG_TREE_SIDEBAR_MIN_WIDTH,
    max: ORG_TREE_SIDEBAR_MAX_WIDTH,
    mode: "pixel",
    containerRef,
  });


  const mainApiUrl = useMemo(() => {
    if (!auth?.backendUrl) return "";
    try {
      return deriveMainApiUrl(auth.backendUrl);
    } catch {
      return auth.backendUrl;
    }
  }, [auth]);

  const ensureActiveWorkspaceHolding = useCallback(async () => {
    const holdingcode = workspace?.shop?.holdingcode?.trim();
    if (!auth || !holdingcode) return;

    const res = await authFetch("/api/workspace/select-holding", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-bc-backend-url": auth.backendUrl,
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ backendUrl: auth.backendUrl, holdingcode }),
    });
    const json = (await res.json().catch(() => ({}))) as { success?: boolean; message?: string };
    if (!res.ok || json.success === false) {
      throw new Error(json.message || tr("st_select_company_load_failed", "เลือกบริษัทสำหรับโหลดข้อมูลไม่สำเร็จ"));
    }
  }, [auth, workspace, tr]);

  // Active languages for multilingual names
  const editorLanguages = useMemo(() => {
    if (!workspace) return ["th"];
    const configs = workspace.shopInfo?.settings?.languageconfigs || [];
    const defaultCode = workspace.shopInfo?.settings?.language || "th";
    return normalizeLanguageConfigs(configs, defaultCode, { forcePrimaryFirst: true }).map((row) => row.code);
  }, [workspace]);

  const timezoneChoices = useMemo(() => timezoneSelectOptions(language), [language]);
  const workspaceDefaultLanguage = useMemo(
    () => workspace?.shopInfo?.settings?.language || "th",
    [workspace],
  );
  const canCreateOrganization = Boolean(
    auth?.profile?.email?.trim() && [1, 2].includes(Number(workspace?.shop?.role)),
  );

  // Collapsed states
  const [collapsedCompanies, setCollapsedCompanies] = useState<Record<string, boolean>>({});

  // Selected Node for Right Form Editing
  const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
  const [formType, setFormType] = useState<OrganizationFormType | null>(null);

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [randomCode, setRandomCode] = useState("");
  const [inputCode, setInputCode] = useState("");
  const [codeError, setCodeError] = useState(false);

  // Fetch Companies & Branches
  const loadData = useCallback(async () => {
    if (!auth || !mainApiUrl) return;
    setLoading(true);
    setLoadError("");
    try {
      await ensureActiveWorkspaceHolding();

      // Load Companies
      const cacheBuster = Date.now().toString();
      const resComp = await authFetch(`${mainApiUrl}/organization/company?management=true&_=${cacheBuster}`, {
        headers: { Authorization: `Bearer ${auth.token}` },
        cache: "no-store",
      });
      const jsonComp = await resComp.json();
      if (!resComp.ok || jsonComp.success === false) {
        throw new Error(jsonComp.message || tr("st_load_company_data_failed", "โหลดข้อมูลบริษัทไม่สำเร็จ"));
      }
      if (jsonComp.success && Array.isArray(jsonComp.data)) {
        setCompanies(jsonComp.data.filter(isVisibleOrganizationRecord));
      }

      // Load Branches
      const resBranch = await authFetch(`${mainApiUrl}/organization/branch?management=true&_=${cacheBuster}`, {
        headers: { Authorization: `Bearer ${auth.token}` },
        cache: "no-store",
      });
      const jsonBranch = await resBranch.json();
      if (!resBranch.ok || jsonBranch.success === false) {
        throw new Error(jsonBranch.message || tr("st_load_branch_data_failed", "โหลดข้อมูลสาขาไม่สำเร็จ"));
      }
      if (jsonBranch.success && Array.isArray(jsonBranch.data)) {
        setBranches(jsonBranch.data.filter(isVisibleOrganizationRecord));
      }
    } catch (e) {
      setLoadError(e instanceof Error && e.message ? e.message : tr("st_load_org_structure_failed", "โหลดข้อมูลโครงสร้างองค์กรไม่สำเร็จ"));
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [auth, ensureActiveWorkspaceHolding, mainApiUrl, tr]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  // Right Form Values
  const [formCode, setFormCode] = useState("");
  const [formTaxId, setFormTaxId] = useState("");
  const [formIsActive, setFormIsActive] = useState(true);
  const [formStatusReason, setFormStatusReason] = useState("");
  const [formNames, setFormNames] = useState<LocalizedNameEntry[]>([]);
  const [formLogoUri, setFormLogoUri] = useState("");
  const [formTimezone, setFormTimezone] = useState(DEFAULT_TIME_ZONE);
  const [formLanguage, setFormLanguage] = useState("th");
  const [formDateFormat, setFormDateFormat] = useState("dd/MM/yyyy");
  const [formYearType, setFormYearType] = useState("buddhist");
  const [formBranchType, setFormBranchType] = useState("permanent");
  const [formIsVatRegistered, setFormIsVatRegistered] = useState(false);
  const [formCompanyRegNo, setFormCompanyRegNo] = useState("");
  const [formEmail, setFormEmail] = useState("");
  const [formManagerName, setFormManagerName] = useState("");
  const [formFiscalStartMonth, setFormFiscalStartMonth] = useState(1);
  const [formDocFormats, setFormDocFormats] = useState<DocFormat[]>([]);
  const [docDrawerOpen, setDocDrawerOpen] = useState(false);
  const [formAddresses, setFormAddresses] = useState<Record<string, string>>({});
  const [formCountryCode, setFormCountryCode] = useState("TH");
  const [formProvinceCode, setFormProvinceCode] = useState("");
  const [formDistrictCode, setFormDistrictCode] = useState("");
  const [formSubdistrictCode, setFormSubdistrictCode] = useState("");
  const [formZipCode, setFormZipCode] = useState("");
  const [formETaxEnabled, setFormETaxEnabled] = useState(false);
  const [logoUploading, setLogoUploading] = useState(false);
  const [logoError, setLogoError] = useState("");
  const logoInputRef = React.useRef<HTMLInputElement>(null);
  const primaryLanguage = editorLanguages[0] || "th";
  const hasRequiredName = formNames.some(
    (entry) => entry.code?.trim().toLowerCase() === primaryLanguage.toLowerCase() && Boolean(entry.name?.trim()),
  );
  const requiredFormError = !formType || formType.startsWith("view")
    ? ""
    : !formCode.trim()
      ? formType.includes("company")
        ? tr("st_please_enter_company_code", "กรุณากรอกรหัสบริษัท")
        : tr("st_enter_branch_code", "กรุณากรอกรหัสสาขา")
      : !hasRequiredName
        ? tr("st_enter_first_lang_code", "กรุณากรอก{0}ในช่องภาษาแรก ({1}) — รหัสภาษาไม่ใช่ชื่อ").replace("{0}", String(formType.includes("company") ? tr("company_name", "ชื่อบริษัท") : tr("company_branch_name", "ชื่อสาขา"))).replace("{1}", String(primaryLanguage.toUpperCase()))
        : formType.includes("branch") && (!formTimezone.trim() || !formLanguage.trim())
          ? tr("st_select_branch_timezone_lang", "กรุณาเลือกเขตเวลาและภาษาของสาขา")
          : "";

  const showConfirmCodeDialog = () => {
    if (requiredFormError) return;
    setSaveError("");
    const code = Math.floor(1000 + Math.random() * 9000).toString();
    setRandomCode(code);
    setInputCode("");
    setCodeError(false);
    setConfirmOpen(true);
  };

  const closeConfirmCodeDialog = () => {
    setConfirmOpen(false);
  };

  const handleConfirmCodeSubmit = () => {
    if (inputCode !== randomCode) {
      setCodeError(true);
      return;
    }

    setConfirmOpen(false);
    void handleSave();
  };

  useEffect(() => {
    if (!selectedNode) return;
    const list: LocalizedNameEntry[] = [];
    const rawNames = selectedNode.data.names;
    editorLanguages.forEach((lang) => {
      list.push({
        code: lang,
        name: getNameFromObject(rawNames, lang),
      });
    });

    setFormCode(selectedNode.type === "company" ? normalizeBusinessCode(selectedNode.data.code) : selectedNode.data.code || "");
    setFormIsActive(selectedNode.data.isactive !== false);
    setFormStatusReason("");
    setFormNames(list);
    setFormLogoUri(String(selectedNode.data.logouri ?? ""));
    setLogoError("");

    if (selectedNode.type === "company") {
      setFormTaxId((selectedNode.data as CompanyRecord).taxid || "");
      setFormAddresses({});
      setFormCountryCode("TH");
      setFormProvinceCode("");
      setFormDistrictCode("");
      setFormSubdistrictCode("");
      setFormZipCode("");
    } else {
      const branchData = selectedNode.data as BranchRecord;
      setFormTimezone(branchData.timezone || DEFAULT_TIME_ZONE);
      setFormLanguage(branchData.language || workspaceDefaultLanguage);
      setFormDateFormat(branchData.dateformat || "dd/MM/yyyy");
      setFormYearType(branchData.yeartype || "buddhist");
      setFormBranchType(
        branchData.branchtype || (isThaiHeadOfficeBranchCode(branchData.code) ? "head" : "permanent"),
      );
      setFormIsVatRegistered(branchData.isvatregistered === true);
      setFormCompanyRegNo(branchData.companyregistrationno || "");
      setFormEmail(branchData.email || "");
      setFormManagerName(branchData.managername || "");
      setFormFiscalStartMonth(branchData.fiscalstartmonth || 1);
      setFormDocFormats(
        (branchData.documentformats || [])
          .filter((e) => e.doctype)
          .map((e) => ({ ...defaultDocFormat(e.doctype as string), ...e, prefix: normalizeDocPrefix(e.prefix ?? (e.doctype as string)) })),
      );
      setFormETaxEnabled(branchData.etaxenabled === true);
      const addrMap: Record<string, string> = {};
      editorLanguages.forEach((lang) => {
        addrMap[lang] = getAddressFromList(branchData.addresses, lang);
      });
      setFormAddresses(addrMap);
      setFormCountryCode(branchData.countrycode || "TH");
      setFormProvinceCode(branchData.provincecode || "");
      setFormDistrictCode(branchData.districtcode || "");
      setFormSubdistrictCode(branchData.subdistrictcode || "");
      setFormZipCode(branchData.zipcode || "");
    }
  }, [selectedNode, editorLanguages, workspaceDefaultLanguage]);

  const handleLogoUpload = async (file: File | undefined) => {
    if (!file || !auth || logoUploading) return;
    // PNG-only for logos.
    const isPng = /\.png$/i.test(file.name) || file.type === "image/png";
    if (!isPng) {
      setLogoError(tr("st_logo_png_only", "โลโก้ต้องเป็นไฟล์ PNG เท่านั้น"));
      return;
    }
    setLogoUploading(true);
    setLogoError("");
    try {
      const uploadForm = new FormData();
      uploadForm.append("file", file, file.name);
      uploadForm.append("category", `system-settings/logouri`);
      const response = await authFetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: uploadForm,
      });
      const payload = (await response.json()) as {
        success?: boolean;
        message?: string;
        data?: {
          holdingcode?: string;
          filename?: string;
          category?: string;
        };
      };
      if (!response.ok || payload.success === false) {
        throw new Error(
          typeof payload.message === "string" ? payload.message : tr("st_logo_upload_failed", "อัปโหลดโลโก้ไม่สำเร็จ"),
        );
      }
      // Backend (image_r2.go) stores images as private R2 objects and only returns
      // holdingcode/filename/category — no URL. We construct the proxy URI
      // `/s3/file/{holdingcode}/{category}/{filename}` which the backend serves
      // through the authenticated S3FileProxyHandler.
      const meta = payload.data ?? {};
      const holding = (meta.holdingcode ?? "").trim();
      const category = (meta.category ?? "").trim();
      const filename = (meta.filename ?? "").trim();
      if (!filename) throw new Error(tr("st_logo_upload_failed", "อัปโหลดโลโก้ไม่สำเร็จ"));
      const pathSegments = [holding, category, filename]
        .filter((segment) => segment.length > 0)
        .map((segment) => segment.replace(/^\/+|\/+$/g, ""));
      const proxyUri = `/goapi/s3/file/${pathSegments.join("/")}`;
      setFormLogoUri(proxyUri);
    } catch (error) {
      setLogoError(error instanceof Error && error.message ? error.message : tr("st_logo_upload_failed", "อัปโหลดโลโก้ไม่สำเร็จ"));
    } finally {
      setLogoUploading(false);
      if (logoInputRef.current) logoInputRef.current.value = "";
    }
  };

  // Handle Save
  const handleSave = async () => {
    if (!auth || !selectedNode || !formType || formType.startsWith("view")) return;
    if (formType.startsWith("create") && !canCreateOrganization) {
      setSaveError(tr("st_only_owner_admin_email_can_create", "เฉพาะ OWNER หรือ ADMIN ที่เชื่อมอีเมลแล้วเท่านั้นที่สร้างบริษัทหรือสาขาได้"));
      return;
    }
    setSaving(true);
    setSaveError("");

    try {
      await ensureActiveWorkspaceHolding();

      const namesList = formNames.map((entry) => ({
        ...entry,
        code: entry.code?.trim(),
        name: entry.name?.trim(),
      }));
      const normalizedCompanyCode = formType.includes("company") ? normalizeBusinessCode(formCode) : "";
      const normalizedBranchCode = formType.includes("branch") ? normalizeThaiTaxBranchCode(formCode) : "";

      const isBranchForm = formType.includes("branch");
      if (!hasRequiredName) {
        setSaveError(formType.includes("company") ? tr("st_enter_company_name_first_lang", "กรุณากรอกชื่อบริษัทภาษาแรก") : tr("st_enter_branch_name_first_lang", "กรุณากรอกชื่อสาขาภาษาแรก"));
        setSaving(false);
        return;
      }
      const statusChanged = !formType.startsWith("create") && formIsActive !== (selectedNode.data.isactive !== false);
      if (statusChanged && !formStatusReason.trim()) {
        setSaveError(tr("st_specify_status_change_reason", "กรุณาระบุเหตุผลที่เปลี่ยนสถานะ"));
        setSaving(false);
        return;
      }
      if (isBranchForm && (!formTimezone.trim() || !formLanguage.trim())) {
        setSaveError(tr("st_select_branch_timezone_lang", "กรุณาเลือกเขตเวลาและภาษาของสาขา"));
        setSaving(false);
        return;
      }
      const branchTzMeta = timezoneMeta(formTimezone);
      // คำนำหน้าห้ามซ้ำกันทั้งสาขา (ข้ามทุกประเภท) — กันบันทึกถ้าซ้ำ
      const dupPrefixes = duplicateDocPrefixes(formDocFormats);
      if (dupPrefixes.size > 0) {
        setSaveError(tr("st_document_prefix_duplicate", "คำนำหน้าเลขที่เอกสารซ้ำ: {0} — ต้องแก้ให้ไม่ซ้ำก่อนบันทึก").replace("{0}", String([...dupPrefixes].join(", "))));
        setSaving(false);
        return;
      }
      const docFormatsPayload = formDocFormats.map((f) => ({ ...f, prefix: normalizeDocPrefix(f.prefix), doctype: f.doctype.toUpperCase() }));
      const addressesPayload = editorLanguages
        .map((lang) => ({ code: lang, address: (formAddresses[lang] || "").trim() }))
        .filter((item) => item.address);

      let url = "";
      let method = "POST";
      let body: Record<string, unknown> = {};

      if (formType === "createcompany") {
        url = `${mainApiUrl}/organization/company`;
        method = "POST";
        body = {
          code: normalizedCompanyCode,
          names: namesList,
          taxid: formTaxId,
          logouri: formLogoUri,
          isactive: true,
        };
      } else if (formType === "editcompany") {
        url = `${mainApiUrl}/organization/company/${selectedNode.guidfixed}`;
        method = "PUT";
        body = {
          code: normalizedCompanyCode,
          names: namesList,
          taxid: formTaxId,
          logouri: formLogoUri,
          isactive: formIsActive,
          statusreason: formStatusReason.trim(),
        };
      } else if (formType === "createbranch") {
        url = `${mainApiUrl}/organization/branch`;
        method = "POST";
        body = {
          companyuid: selectedNode.companyuid,
          code: normalizedBranchCode,
          names: namesList,
          logouri: formLogoUri,
          timezone: formTimezone,
          timezonelabel: branchTzMeta.label,
          timezoneoffset: branchTzMeta.offset,
          language: formLanguage,
          dateformat: formDateFormat,
          yeartype: formYearType,
          branchtype: formBranchType,
          isvatregistered: formIsVatRegistered,
          companyregistrationno: formCompanyRegNo,
          email: formEmail,
          managername: formManagerName,
          fiscalstartmonth: formFiscalStartMonth,
          documentformats: docFormatsPayload,
          addresses: addressesPayload,
          countrycode: formCountryCode,
          provincecode: formProvinceCode,
          districtcode: formDistrictCode,
          subdistrictcode: formSubdistrictCode,
          zipcode: formZipCode,
          etaxenabled: formETaxEnabled,
          isactive: true,
        };
      } else if (formType === "editbranch") {
        url = `${mainApiUrl}/organization/branch/${selectedNode.guidfixed}`;
        method = "PUT";
        body = {
          companyuid: selectedNode.companyuid,
          code: normalizedBranchCode,
          names: namesList,
          logouri: formLogoUri,
          timezone: formTimezone,
          timezonelabel: branchTzMeta.label,
          timezoneoffset: branchTzMeta.offset,
          language: formLanguage,
          dateformat: formDateFormat,
          yeartype: formYearType,
          branchtype: formBranchType,
          isvatregistered: formIsVatRegistered,
          companyregistrationno: formCompanyRegNo,
          email: formEmail,
          managername: formManagerName,
          fiscalstartmonth: formFiscalStartMonth,
          documentformats: docFormatsPayload,
          addresses: addressesPayload,
          countrycode: formCountryCode,
          provincecode: formProvinceCode,
          districtcode: formDistrictCode,
          subdistrictcode: formSubdistrictCode,
          zipcode: formZipCode,
          etaxenabled: formETaxEnabled,
          isactive: formIsActive,
          statusreason: formStatusReason.trim(),
        };
      }

      const res = await authFetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(body),
      });

      const json = (await res.json().catch(() => ({}))) as OrganizationSaveResponse;
      if (!res.ok || json.success === false) {
        setSaveError(saveErrorMessage(json.message, formType, tr));
        return;
      }
      if (json.success) {
        const createdData = formType.startsWith("create") ? json.data?.entity : undefined;
        const createdGuid = createdData?.guidfixed?.trim() ?? "";
        const hasSavedIdentity = formType === "createcompany"
          ? Boolean((createdData as CompanyRecord | undefined)?.holdinguid?.trim() && (createdData as CompanyRecord | undefined)?.companyuid?.trim())
          : formType === "createbranch"
            ? Boolean(
                (createdData as BranchRecord | undefined)?.holdinguid?.trim()
                && (createdData as BranchRecord | undefined)?.companyuid?.trim()
                && (createdData as BranchRecord | undefined)?.branchuid?.trim(),
              )
            : true;
        if (
          formType.startsWith("create")
          && (!createdData || !createdGuid || !hasSavedIdentity || createdData.isactive !== true || createdData.isdeleted === true)
        ) {
          await loadData();
          setSaveError(tr("st_saved_but_backend_incomplete", "บันทึกสำเร็จ แต่ Backend ไม่คืนข้อมูลที่บันทึกครบถ้วน กรุณารีเฟรชหน้าจอ"));
          return;
        }

        setSaveSuccess(true);
        setTimeout(() => setSaveSuccess(false), 2000);
        await loadData();
        onRefresh?.();
        notifyWorkspaceChanged();
        // Switch back to view mode and refresh the selected node data so the
        // logo (and any other changed fields) show immediately without a manual
        // close/reopen of the form.
        if (formType.startsWith("create")) {
          const savedEntity = createdData as CompanyRecord | BranchRecord;
          if (formType === "createcompany") {
            const savedCompany = savedEntity as CompanyRecord;
            setCompanies((prev) => [...prev.filter((row) => row.guidfixed !== createdGuid), savedCompany]);
          } else {
            const savedBranch = savedEntity as BranchRecord;
            setBranches((prev) => [...prev.filter((row) => row.guidfixed !== createdGuid), savedBranch]);
          }
          // Select newly created node
          setSelectedNode({
            type: formType === "createcompany" ? "company" : "branch",
            guidfixed: createdGuid,
            companyuid: formType === "createbranch" ? (savedEntity as BranchRecord).companyuid : undefined,
            companyguid: formType === "createbranch" ? (savedEntity as BranchRecord).companyguid : undefined,
            data: savedEntity,
          });
          setFormType(formType === "createcompany" ? "editcompany" : "editbranch");
        } else {
          // Edit mode: switch back to view mode and refresh selectedNode data so
          // the logo and other changed fields display immediately without a
          // manual close/reopen of the form.
          const updatedData = {
            ...selectedNode.data,
            code: selectedNode.type === "company" ? normalizedCompanyCode : normalizedBranchCode,
            names: namesList,
            logouri: formLogoUri,
            isactive: formIsActive,
            ...(selectedNode.type === "company"
              ? { taxid: formTaxId }
              : {
                  timezone: formTimezone,
                  timezonelabel: branchTzMeta.label,
                  timezoneoffset: branchTzMeta.offset,
                  language: formLanguage,
                  dateformat: formDateFormat,
                  yeartype: formYearType,
                  branchtype: formBranchType,
                  isvatregistered: formIsVatRegistered,
                  companyregistrationno: formCompanyRegNo,
                  email: formEmail,
                  managername: formManagerName,
                  fiscalstartmonth: formFiscalStartMonth,
                  documentformats: docFormatsPayload,
                  addresses: addressesPayload,
                  countrycode: formCountryCode,
                  provincecode: formProvinceCode,
                  districtcode: formDistrictCode,
                  subdistrictcode: formSubdistrictCode,
                  zipcode: formZipCode,
                  etaxenabled: formETaxEnabled,
                }),
          };
          if (selectedNode.type === "company") {
            setCompanies((prev) => prev.map((row) => (
              row.guidfixed === selectedNode.guidfixed ? { ...updatedData, guidfixed: selectedNode.guidfixed } : row
            )));
          } else {
            setBranches((prev) => prev.map((row) => (
              row.guidfixed === selectedNode.guidfixed
                ? {
                    ...updatedData,
                    guidfixed: selectedNode.guidfixed,
                    companyuid: selectedNode.companyuid,
                    companyguid: selectedNode.companyguid,
                  }
                : row
            )));
          }
          setSelectedNode({
            ...selectedNode,
            data: updatedData,
          });
          setFormType(selectedNode.type === "company" ? "viewcompany" : "viewbranch");
        }
      }
    } catch (e) {
      setSaveError(e instanceof Error && e.message ? e.message : tr("st_save_data_failed", "บันทึกข้อมูลไม่สำเร็จ"));
      console.error(e);
    } finally {
      setSaving(false);
    }
  };

  const sortedCompanies = useMemo(() => {
    return [...companies].sort((a, b) => {
      const codeA = normalizeBusinessCode(a.code);
      const codeB = normalizeBusinessCode(b.code);
      const cmp = codeA.localeCompare(codeB, undefined, { numeric: true, sensitivity: "base" });
      if (cmp !== 0) return cmp;

      const nameA = getNameFromObject(a.names, language) || "";
      const nameB = getNameFromObject(b.names, language) || "";
      return nameA.localeCompare(nameB, "th", { sensitivity: "base" });
    });
  }, [companies, language]);

  const sortedBranches = useMemo(() => {
    return [...branches].sort((a, b) => {
      const codeA = a.code || "";
      const codeB = b.code || "";
      const cmp = codeA.localeCompare(codeB, undefined, { numeric: true, sensitivity: "base" });
      if (cmp !== 0) return cmp;

      const nameA = getNameFromObject(a.names, language) || "";
      const nameB = getNameFromObject(b.names, language) || "";
      return nameA.localeCompare(nameB, "th", { sensitivity: "base" });
    });
  }, [branches, language]);

  const isReadOnlyMode = formType?.startsWith("view") ?? false;
  const selectedCompanyUID = selectedNode?.type === "company"
    ? (selectedNode.data as CompanyRecord).companyuid?.trim() || ""
    : "";

  return (
    <div
      ref={containerRef}
      style={{ ["--org-sidebar-width" as any]: `${sidebarWidth}px` }}
      className="relative grid w-full min-w-0 items-stretch gap-2 min-h-[calc(100dvh-12rem)] grid-cols-1 lg:grid-cols-[var(--org-sidebar-width)_auto_minmax(0,1fr)]"
    >
      {/* Left panel: Company -> Branch list */}
      <Card className="min-h-[calc(100vh-12rem)] shadow-lg border-primary/10">
        <CardContent className="p-4">
          <div className="flex items-center justify-between border-b pb-3 mb-4">
            <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
              <Building2 className="w-5 h-5 text-primary" />
              {tr("st_organization_structure", "โครงสร้างองค์กร")}
            </h2>
            <Button
              size="sm"
              className="gap-1 font-semibold"
              disabled={!canCreateOrganization}
              title={!canCreateOrganization ? tr("st_owner_admin_email_required", "ต้องเป็น OWNER/ADMIN และเชื่อมอีเมลก่อน") : tr("st_add_company", "เพิ่มบริษัท")}
              onClick={() => {
                setFormType("createcompany");
                setSelectedNode({
                  type: "company",
                  data: {},
                });
              }}
            >
              <Plus className="w-4 h-4" />
              {tr("st_add_company", "เพิ่มบริษัท")}
            </Button>
          </div>
          {loadError && (
            <div className="mb-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
              {loadError}
            </div>
          )}

          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
            </div>
          ) : sortedCompanies.length === 0 ? (
            <div className="rounded-lg border border-dashed border-border bg-muted/30 px-3 py-8 text-center text-sm font-semibold text-muted-foreground">
              {tr("st_no_company_data", "ยังไม่มีข้อมูลบริษัทในกิจการนี้")}
            </div>
          ) : (
            <div className="space-y-2">
              {sortedCompanies.map((comp) => {
                const compGuid = comp.guidfixed || "";
                const companyUID = comp.companyuid?.trim() || "";
                const companyKey = companyTreeKey(comp);
                const companyCode = normalizeBusinessCode(comp.code);
                const isCollapsed = collapsedCompanies[companyKey];
                const isSelected = selectedNode?.type === "company" && selectedNode.guidfixed === compGuid;
                const isEditingCompany = isSelected && formType === "editcompany";
                const compBranches = sortedBranches.filter((branch) => branchParentTreeKey(branch) === companyKey);
                const canAddBranch = canCreateOrganization && Boolean(companyUID);

                return (
                  <div key={companyKey || compGuid} className="space-y-1">
                    <div
                      className={cn(
                        "group flex items-center justify-between p-2 rounded-lg cursor-pointer transition-all",
                        isEditingCompany
                          ? "bg-amber-100/70 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-l-4 border-amber-400 pl-1 font-bold"
                          : isSelected
                            ? "bg-primary/10 border-l-4 border-primary pl-1 font-bold"
                            : "hover:bg-accent pl-2"
                      )}
                      onClick={() => {
                        setSelectedNode({
                          type: "company",
                          guidfixed: compGuid,
                          data: comp,
                        });
                        setFormType("viewcompany");
                      }}
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <button
                          type="button"
                          aria-label={tr("st_collapse_expand_branch", "ย่อ/ขยายสาขา")}
                          onClick={(e) => {
                            e.stopPropagation();
                            setCollapsedCompanies((prev) => ({
                              ...prev,
                              [companyKey]: !prev[companyKey],
                            }));
                          }}
                          className="p-1 hover:bg-black/5 dark:hover:bg-white/5 rounded"
                        >
                          {isCollapsed ? (
                            <ChevronRight className="w-4 h-4 text-muted-foreground" />
                          ) : (
                            <ChevronDown className="w-4 h-4 text-muted-foreground" />
                          )}
                        </button>
                        <LogoAvatar
                          uri={comp.logouri}
                          auth={auth}
                          alt={getNameFromObject(comp.names, language) || companyCode}
                          sizeClass="size-5 rounded-md shrink-0"
                          iconSize={14}
                          width={64}
                          className="text-primary"
                        />
                        <span className="min-w-0 break-words">
                          {getNameFromObject(comp.names, language) || companyCode}
                        </span>
                        {companyCode ? (
                          <span className="font-mono text-[9px] px-1.5 py-0.2 bg-muted border border-border/50 text-muted-foreground rounded uppercase font-bold shrink-0">
                            {companyCode}
                          </span>
                        ) : null}
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 text-primary hover:text-primary hover:bg-primary/10"
                          title={tr("st_edit_company", "แก้ไขบริษัท")}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({
                              type: "company",
                              guidfixed: compGuid,
                              data: comp,
                            });
                            setFormType("editcompany");
                          }}
                        >
                          <Edit3 className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 text-sky-500 hover:text-sky-600 hover:bg-sky-500/10"
                          title={
                            !canCreateOrganization
                              ? tr("st_owner_admin_email_required", "ต้องเป็น OWNER/ADMIN และเชื่อมอีเมลก่อน")
                              : !companyUID
                                ? tr("st_company_no_uid_cannot_add_branch", "บริษัทนี้ยังไม่มีรหัสถาวร companyuid จึงเพิ่มสาขาไม่ได้")
                                : tr("add_branch", "เพิ่มสาขา")
                          }
                          disabled={!canAddBranch}
                          onClick={(e) => {
                            e.stopPropagation();
                            if (!canAddBranch) return;
                            setSelectedNode({
                              type: "branch",
                              companyuid: companyUID,
                              data: {},
                            });
                            setFormType("createbranch");
                          }}
                        >
                          <Plus className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </div>

                    {/* Branches list */}
                    {!isCollapsed && compBranches.length > 0 && (
                      <div className="pl-6 border-l ml-4 space-y-1 my-1">
                        {compBranches.map((br) => {
                          const brGuid = br.guidfixed || "";
                          const isBrSelected = selectedNode?.type === "branch" && selectedNode.guidfixed === brGuid;
                          const isEditingBranch = isBrSelected && formType === "editbranch";
                          return (
                            <div
                              key={brGuid}
                              className={cn(
                                "group flex items-center justify-between p-2 rounded-lg cursor-pointer transition-all",
                                isEditingBranch
                                  ? "bg-amber-100/70 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-l-4 border-amber-400 pl-1 font-semibold"
                                  : isBrSelected
                                    ? "bg-sky-500/10 border-l-4 border-sky-500 pl-1 font-semibold"
                                    : "hover:bg-accent pl-2"
                              )}
                              onClick={() => {
                                setSelectedNode({
                                  type: "branch",
                                  guidfixed: brGuid,
                                  companyuid: br.companyuid?.trim() || undefined,
                                  companyguid: br.companyguid?.trim() || undefined,
                                  data: br,
                                });
                                setFormType("viewbranch");
                              }}
                            >
                              <div className="flex items-center gap-2 min-w-0">
                                <LogoAvatar
                                  uri={br.logouri}
                                  auth={auth}
                                  alt={getNameFromObject(br.names, language) || br.code || ""}
                                  sizeClass="size-5 rounded-md shrink-0"
                                  iconSize={14}
                                  width={64}
                                  className="text-sky-500"
                                />
                                <span className="min-w-0 text-sm break-words">
                                  {getNameFromObject(br.names, language) || br.code}
                                </span>
                                {br.code ? (
                                  <span className="font-mono text-[9px] px-1.5 py-0.2 bg-muted border border-border/50 text-muted-foreground rounded uppercase font-bold shrink-0">
                                    {br.code}
                                  </span>
                                ) : null}
                              </div>
                              <div className="flex items-center gap-1">
                                <Button
                                  size="icon"
                                  variant="ghost"
                                  className="w-7 h-7 text-primary hover:text-primary hover:bg-primary/10"
                                  title={tr("st_edit_branch", "แก้ไขสาขา")}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setSelectedNode({
                                      type: "branch",
                                      guidfixed: brGuid,
                                      companyuid: br.companyuid?.trim() || undefined,
                                      companyguid: br.companyguid?.trim() || undefined,
                                      data: br,
                                    });
                                    setFormType("editbranch");
                                  }}
                                >
                                  <Edit3 className="w-3.5 h-3.5" />
                                </Button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      <ResizableSplitter
        breakpoint="lg"
        value={sidebarWidth}
        min={ORG_TREE_SIDEBAR_MIN_WIDTH}
        max={ORG_TREE_SIDEBAR_MAX_WIDTH}
        label={
          tr("st_resize_org_list_details", "ปรับขนาดรายชื่อองค์กรและรายละเอียด (ลากเพื่อปรับ, ดับเบิ้ลคลิกเพื่อรีเซ็ต)")
        }
        isResizing={isResizingSidebar}
        onPointerDown={handleSidebarResizeStart}
        onDoubleClick={handleSidebarResizeReset}
        onKeyDown={handleSidebarKeyDown}
      />

      {/* Right panel: Detail / Edit Form */}
      <Card className="shadow-lg border-primary/10">
        <CardContent className="p-6">
          {formType ? (
            <div className="space-y-6">
              <div className="flex flex-wrap items-start justify-between border-b pb-4 gap-4">
                <div>
                  <h3 className="text-xl font-bold text-foreground flex items-center gap-2">
                    {formType.includes("company") ? (
                      <Building2 className="w-5.5 h-5.5 text-primary shrink-0" />
                    ) : (
                      <GitBranch className="w-5.5 h-5.5 text-sky-500 shrink-0" />
                    )}
                    {formType === "viewcompany" && tr("company_profile", "ข้อมูลบริษัท")}
                    {formType === "viewbranch" && tr("company_branch_data", "ข้อมูลสาขา")}
                    {formType === "createcompany" && tr("st_add_new_company", "เพิ่มบริษัทใหม่")}
                    {formType === "editcompany" && tr("st_edit_company_info", "แก้ไขข้อมูลบริษัท")}
                    {formType === "createbranch" && tr("st_add_new_branch", "เพิ่มสาขาใหม่")}
                    {formType === "editbranch" && tr("st_edit_branch_info", "แก้ไขข้อมูลสาขา")}
                  </h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    {formType.startsWith("create")
                      ? tr("st_enter_main_details_add_data", "ระบุข้อมูลรายละเอียดหลักเพื่อเพิ่มข้อมูลเข้าระบบ")
                      : isReadOnlyMode
                        ? tr("st_show_selected_item_details", "แสดงรายละเอียดข้อมูลจากรายการที่เลือก")
                        : tr("st_edit_details_save_history", "แก้ไขรายละเอียดข้อมูลและบันทึกประวัติ")}
                  </p>
                </div>
                <div className="flex w-full shrink-0 flex-wrap items-center justify-end gap-2 sm:w-auto">
                  {isReadOnlyMode && selectedNode?.guidfixed && (
                    <Button
                      size="icon"
                      variant="outline"
                      className="size-9 border-primary/30 text-primary hover:bg-primary/10"
                      title={selectedNode.type === "company" ? tr("st_edit_company", "แก้ไขบริษัท") : tr("st_edit_branch", "แก้ไขสาขา")}
                      onClick={() => setFormType(selectedNode.type === "company" ? "editcompany" : "editbranch")}
                    >
                      <Edit3 className="w-4 h-4" />
                    </Button>
                  )}
                  {(formType === "viewcompany" || formType === "editcompany") && selectedNode?.guidfixed && (
                    <Button
                      size="sm"
                      variant="outline"
                      className="gap-1 text-xs border-sky-500/30 text-sky-600 hover:bg-sky-500/10 hover:text-sky-700 font-bold shrink-0"
                      disabled={!canCreateOrganization || !selectedCompanyUID}
                      title={
                        !canCreateOrganization
                          ? tr("st_owner_admin_email_required", "ต้องเป็น OWNER/ADMIN และเชื่อมอีเมลก่อน")
                          : !selectedCompanyUID
                            ? tr("st_company_no_uid_cannot_add_branch", "บริษัทนี้ยังไม่มีรหัสถาวร companyuid จึงเพิ่มสาขาไม่ได้")
                            : tr("st_add_branch_to_company", "เพิ่มสาขาในบริษัทนี้")
                      }
                      onClick={() => {
                        if (!canCreateOrganization || !selectedCompanyUID) return;
                        setSelectedNode({
                          type: "branch",
                          companyuid: selectedCompanyUID,
                          data: {},
                        });
                        setFormType("createbranch");
                      }}
                    >
                      <Plus className="w-3.5 h-3.5" />
                      {tr("st_add_branch_to_company", "เพิ่มสาขาในบริษัทนี้")}
                    </Button>
                  )}
                  {!isReadOnlyMode ? (
                    <Button
                      type="button"
                      size="sm"
                      onClick={showConfirmCodeDialog}
                      disabled={saving || saveSuccess || logoUploading}
                      title={requiredFormError || undefined}
                      aria-describedby={requiredFormError ? "company-branch-required-fields" : undefined}
                      className="gap-1.5 text-xs font-bold"
                      data-testid="company-branch-save-action"
                    >
                      {saving ? (
                        <Loader2 className="size-4 animate-spin" />
                      ) : saveSuccess ? (
                        <Check className="size-4" />
                      ) : (
                        <Save className="size-4" />
                      )}
                      {saving ? tr("saving", "กำลังบันทึก...") : saveSuccess ? tr("saved", "บันทึกแล้ว") : tr("fd_save", "บันทึก")}
                    </Button>
                  ) : null}
                </div>
              </div>

              {!isReadOnlyMode && requiredFormError ? (
                <div
                  id="company-branch-required-fields"
                  className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm font-semibold text-amber-800 dark:text-amber-200"
                  role="status"
                >
                  {requiredFormError}
                </div>
              ) : null}

              <div className="space-y-4">
                {/* Logo */}
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-foreground">
                    {formType.includes("company") ? tr("company_logo", "โลโก้บริษัท") : tr("st_branch_logo", "โลโก้สาขา")}
                  </label>
                  <div className="flex items-start gap-3">
                    <LogoAvatar
                      uri={formLogoUri}
                      auth={auth}
                      alt={formType.includes("company") ? tr("company_logo", "โลโก้บริษัท") : tr("st_branch_logo", "โลโก้สาขา")}
                      sizeClass="size-20 rounded-2xl"
                      iconSize={32}
                      width={256}
                      className="border border-input"
                    />
                    {!isReadOnlyMode && (
                      <div className="flex flex-col gap-1.5 min-w-48 flex-1">
                        <div className="flex gap-2">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            disabled={logoUploading}
                            onClick={() => logoInputRef.current?.click()}
                          >
                            {logoUploading ? <Loader2 className="w-4 h-4 animate-spin" /> : <UploadCloud className="w-4 h-4" />}
                            {logoUploading ? tr("import_product_file_uploading", "กำลังอัปโหลด...") : formLogoUri ? tr("st_change_logo", "เปลี่ยนโลโก้") : tr("st_select_png_file", "เลือกไฟล์ PNG")}
                          </Button>
                          {formLogoUri && (
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              className="text-destructive hover:bg-destructive/5"
                              disabled={logoUploading}
                              onClick={() => {
                                setFormLogoUri("");
                                setLogoError("");
                              }}
                            >
                              <X className="w-4 h-4" /> {tr("delete", "ลบ")}
                            </Button>
                          )}
                        </div>
                        <p className="text-xs text-muted-foreground">
                          {tr("st_png_transparent_form_print", "รองรับเฉพาะไฟล์ PNG พื้นหลังโปร่งใสได้ ใช้สำหรับออกแบบฟอร์มและพิมพ์เอกสาร")}
                          {formType.includes("company")
                            ? " " + tr("st_logo_saved_company_only", "โลโก้นี้บันทึกเฉพาะบริษัท")
                            : " " + tr("st_logo_saved_this_branch_only", "โลโก้นี้บันทึกเฉพาะสาขานี้ และแตกต่างจากสาขาอื่นได้")}
                        </p>
                        {logoError && (
                          <p className="text-xs font-semibold text-destructive">{logoError}</p>
                        )}
                        <input
                          ref={logoInputRef}
                          className="sr-only"
                          type="file"
                          accept="image/png"
                          onChange={(e) => void handleLogoUpload(e.target.files?.[0])}
                        />
                      </div>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm font-semibold text-foreground">
                      {formType.includes("company") ? tr("st_company_code", "รหัสบริษัท *") : tr("st_branch_code", "รหัสสาขา *")}
                    </label>
                    <Input
                      value={formCode}
                      onChange={(e) => {
                        setFormCode(formType.includes("company") ? normalizeBusinessCode(e.target.value) : e.target.value);
                      }}
                      placeholder={formType.includes("company") ? tr("st_enter_company_code_00000", "ระบุรหัสบริษัท เช่น 00000") : tr("st_enter_branch_code_5_digits_00001", "ระบุรหัสสาขา 5 หลัก เช่น 00001")}
                      className="bg-accent/20"
                      disabled={isReadOnlyMode}
                    />
                  </div>
                  {formType.includes("company") && (
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">{tr("st_tax_id", "เลขประจำตัวผู้เสียภาษี (Tax ID)")}</label>
                      <Input
                        value={formTaxId}
                        onChange={(e) => setFormTaxId(e.target.value)}
                        placeholder={tr("st_taxpayer_number_13_digits", "เลขผู้เสียภาษี 13 หลัก")}
                        className="bg-accent/20"
                        disabled={isReadOnlyMode}
                      />
                    </div>
                  )}
                </div>

                {/* Branch locale: timezone + language (both required — branches may differ) */}
                {formType.includes("branch") && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">{tr("st_timezone", "เขตเวลา (Timezone) *")}</label>
                      <select
                        value={formTimezone}
                        onChange={(e) => setFormTimezone(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {timezoneChoices.map((tz) => (
                          <option key={tz.value} value={tz.value}>
                            {tz.label}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">{tr("st_branch_language", "ภาษาของสาขา *")}</label>
                      <select
                        value={formLanguage}
                        onChange={(e) => setFormLanguage(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {LANGUAGES.map((lang) => (
                          <option key={lang.code} value={lang.code}>
                            {lang.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {/* Branch date settings (backend already carries these) */}
                {formType.includes("branch") && (
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">{tr("yeartype", "ปีศักราชที่ใช้")}</label>
                      <ChoiceSelect
                        value={formYearType}
                        onChange={(val) => setFormYearType(String(val))}
                        disabled={isReadOnlyMode}
                        options={YEAR_TYPE_OPTIONS.map((opt) => ({
                          value: opt.value,
                          label: tr(opt.label[0], opt.label[1]),
                        }))}
                      />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">{tr("dateformat", "รูปแบบวันที่")}</label>
                      <select
                        value={formDateFormat}
                        onChange={(e) => setFormDateFormat(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {dateFormatOptionsFor(formYearType).map((opt) => (
                          <option key={opt.value} value={opt.value}>
                            {opt.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {/* Branch tax / registration (ภ.พ.20) */}
                {formType.includes("branch") && (
                  <div className="space-y-4 rounded-lg border border-border/60 bg-muted/20 p-4">
                    <p className="text-sm font-bold text-foreground">{tr("st_tax_registration_pp20", "ข้อมูลภาษี / ทะเบียน (ภ.พ.20)")}</p>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">{tr("st_branch_type_pp20", "ประเภทสาขา (ภ.พ.20)")}</label>
                        <ChoiceSelect
                          value={formBranchType}
                          onChange={(val) => setFormBranchType(String(val))}
                          disabled={isReadOnlyMode}
                          options={BRANCH_TYPE_OPTIONS.map((opt) => ({
                            value: opt.value,
                            label: tr(opt.label[0], opt.label[1]),
                          }))}
                        />
                      </div>
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">{tr("st_juristic_registration_number", "เลขทะเบียนนิติบุคคล")}</label>
                        <Input
                          value={formCompanyRegNo}
                          onChange={(e) => setFormCompanyRegNo(e.target.value)}
                          placeholder={tr("st_juristic_registration_number_13", "เลขทะเบียนนิติบุคคล 13 หลัก")}
                          className="bg-accent/20"
                          disabled={isReadOnlyMode}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">{tr("st_branch_manager", "ผู้จัดการสาขา")}</label>
                        <Input
                          value={formManagerName}
                          onChange={(e) => setFormManagerName(e.target.value)}
                          placeholder={tr("st_branch_manager_name", "ชื่อผู้จัดการสาขา")}
                          className="bg-accent/20"
                          disabled={isReadOnlyMode}
                        />
                      </div>
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">{tr("st_branch_email", "อีเมลสาขา")}</label>
                        <Input
                          type="email"
                          value={formEmail}
                          onChange={(e) => setFormEmail(e.target.value)}
                          placeholder={tr("st_email_for_documents", "อีเมลสำหรับส่งเอกสาร")}
                          className="bg-accent/20"
                          disabled={isReadOnlyMode}
                        />
                      </div>
                    </div>
                    <div className="pt-1">
                      <CheckboxCard
                        id="isvatregistered"
                        label={tr("st_vat_registration", "จดทะเบียนภาษีมูลค่าเพิ่ม (VAT)")}
                        checked={formIsVatRegistered}
                        disabled={isReadOnlyMode}
                        onCheckedChange={setFormIsVatRegistered}
                        cardClassName="w-full"
                      />
                    </div>
                  </div>
                )}

                {/* Document / fiscal config — value-only (generator + e-Tax engine are separate, not active) */}
                {formType.includes("branch") && (
                  <div className="space-y-4 rounded-lg border border-border/60 bg-muted/20 p-4">
                    <p className="text-sm font-bold text-foreground">{tr("st_accounting_doc_period", "รอบบัญชี / เอกสาร")}</p>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">{tr("st_accounting_period_start_month", "เดือนเริ่มรอบบัญชี")}</label>
                        <select
                          value={formFiscalStartMonth}
                          onChange={(e) => setFormFiscalStartMonth(Number(e.target.value))}
                          disabled={isReadOnlyMode}
                          className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          {MONTH_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>
                              {tr(opt.label[0], opt.label[1])}
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>

                    {/* รูปแบบเลขที่เอกสาร — สรุปย่อ + ปุ่มเปิด drawer จัดการ (builder เต็มอยู่ใน drawer) */}
                    {(() => {
                      const dups = duplicateDocPrefixes(formDocFormats);
                      const summaries = DOC_PREFIX_TYPES
                        .map((dt) => {
                          const f = formDocFormats.find((x) => x.doctype === dt.code && x.isdefault)
                            ?? formDocFormats.find((x) => x.doctype === dt.code);
                          return f ? { dt, ex: buildDocExample(f, formCode || "00000") } : null;
                        })
                        .filter((x): x is { dt: (typeof DOC_PREFIX_TYPES)[number]; ex: string } => x !== null);
                      return (
                        <div className="space-y-2">
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <label className="text-sm font-semibold text-foreground">
                              {tr("order_setting_doc_format", "รูปแบบเลขที่เอกสาร")}
                              <span className="ml-1 font-normal text-xs text-muted-foreground">{tr("st_formats_in_document_types", "— {0} รูปแบบ ใน {1} ประเภท").replace("{0}", String(formDocFormats.length)).replace("{1}", String(new Set(formDocFormats.map((f) => f.doctype)).size))}</span>
                            </label>
                            <Button type="button" variant="outline" size="sm" onClick={() => setDocDrawerOpen(true)} disabled={isReadOnlyMode}>
                              <FileText className="h-3.5 w-3.5" /> {tr("st_manage_doc_number_formats", "จัดการรูปแบบเลขที่เอกสาร")}
                            </Button>
                          </div>
                          {dups.size > 0 ? (
                            <div className="rounded-md bg-destructive/10 px-3 py-1.5 text-xs text-destructive">
                              {tr("st_duplicate_prefix_manage", "คำนำหน้าซ้ำ: {0} — กดจัดการเพื่อแก้ก่อนบันทึก").replace("{0}", String([...dups].join(", ")))}
                            </div>
                          ) : null}
                          {summaries.length > 0 ? (
                            <div className="flex flex-wrap gap-1.5">
                              {summaries.map(({ dt, ex }) => (
                                <span key={dt.code} className="inline-flex items-center gap-1 rounded-md border border-border bg-muted/40 px-2 py-1 text-xs">
                                  <span className="text-muted-foreground">{tr(dt.label[0], dt.label[1])}</span>
                                  <span className="font-mono font-medium text-primary">{ex}</span>
                                </span>
                              ))}
                            </div>
                          ) : (
                            <p className="text-xs text-muted-foreground">{tr("st_no_format_add_manage", "ยังไม่มีรูปแบบ — กด “จัดการรูปแบบเลขที่เอกสาร” เพื่อเพิ่ม")}</p>
                          )}
                        </div>
                      );
                    })()}
                    <div className="pt-1">
                      <CheckboxCard
                        id="etaxenabled"
                        label={tr("st_enable_e_tax", "เปิดใช้ใบกำกับภาษีอิเล็กทรอนิกส์ (e-Tax)")}
                        checked={formETaxEnabled}
                        disabled={isReadOnlyMode}
                        onCheckedChange={setFormETaxEnabled}
                        cardClassName="w-full"
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {tr("st_auto_doc_number_note", "* การออกเลขที่เอกสารอัตโนมัติและการส่ง e-Tax เป็นระบบแยก ยังไม่เปิดใช้งาน — ค่านี้เก็บไว้ตั้งค่าล่วงหน้า")}
                    </p>
                  </div>
                )}

                {/* Multilingual names */}
                <div className="space-y-3">
                  <NamesEditor
                    key={`${formType}:${selectedNode?.guidfixed || selectedNode?.companyuid || "new"}`}
                    names={formNames}
                    onChange={setFormNames}
                    languages={editorLanguages}
                    label={formType.includes("company") ? tr("company_name", "ชื่อบริษัท") : tr("company_branch_name", "ชื่อสาขา")}
                    firstRequired
                    language={language}
                    disabled={isReadOnlyMode}
                    languageSelect
                  />
                  {formType.includes("branch") && (
                    <AddressesEditor
                      addresses={formAddresses}
                      onChange={setFormAddresses}
                      languages={editorLanguages}
                      label={tr("st_branch_address_documents", "ที่อยู่สาขา (สำหรับออกเอกสาร)")}
                      language={language}
                      disabled={isReadOnlyMode}
                    />
                  )}
                  {formType.includes("branch") && (
                    <BranchGeoAddressPicker
                      backendUrl={auth?.backendUrl}
                      countryCode={formCountryCode}
                      provinceCode={formProvinceCode}
                      districtCode={formDistrictCode}
                      subdistrictCode={formSubdistrictCode}
                      zipCode={formZipCode}
                      language={language}
                      disabled={isReadOnlyMode}
                      onChange={(next) => {
                        if (next.countrycode !== undefined) setFormCountryCode(next.countrycode);
                        if (next.provincecode !== undefined) setFormProvinceCode(next.provincecode);
                        if (next.districtcode !== undefined) setFormDistrictCode(next.districtcode);
                        if (next.subdistrictcode !== undefined) setFormSubdistrictCode(next.subdistrictcode);
                        if (next.zipcode !== undefined) setFormZipCode(next.zipcode);
                      }}
                    />
                  )}
                </div>
                {saveError && (
                  <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
                    {saveError}
                  </div>
                )}

                  <div className="pt-2">
                    <CheckboxCard
                      id="isactive"
                      label={tr("st_enable_in_system", "เปิดใช้งานในระบบ")}
                      checked={formType.startsWith("create") || formIsActive}
                      disabled={isReadOnlyMode || formType.startsWith("create")}
                      onCheckedChange={setFormIsActive}
                      cardClassName="w-full"
                    />
                  </div>
                {selectedNode && !formType.startsWith("create") && formIsActive !== (selectedNode.data.isactive !== false) && (
                  <label className="block space-y-1.5">
                    <span className="text-sm font-semibold text-foreground">{tr("st_reason_status_change", "เหตุผลที่เปลี่ยนสถานะ")}</span>
                    <textarea
                      value={formStatusReason}
                      onChange={(event) => setFormStatusReason(event.target.value)}
                      disabled={isReadOnlyMode}
                      rows={3}
                      className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                      placeholder={tr("st_reason_for_audit_log", "ระบุเหตุผลเพื่อบันทึก Audit")}
                    />
                  </label>
                )}

              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-24 text-muted-foreground">
              <Building2 className="w-16 h-16 text-muted-foreground/30 mb-4" />
              <p className="text-lg font-semibold">{tr("st_select_company_or_branch", "กรุณาเลือก บริษัท หรือ สาขา")}</p>
              <p className="text-sm mt-1">{tr("st_click_left_menu_view_edit_data", "คลิกเลือกรายการที่แถบเมนูด้านซ้ายเพื่อดูหรือแก้ไขข้อมูล")}</p>
            </div>
          )}
        </CardContent>
      </Card>
      {docDrawerOpen && (
        <div className="fixed inset-0 z-50 flex bg-black/40" onClick={() => setDocDrawerOpen(false)}>
          <div
            className="ml-auto flex h-full w-full max-w-4xl flex-col bg-card shadow-2xl animate-in slide-in-from-right duration-200"
            onClick={(e) => e.stopPropagation()}
            role="dialog"
            aria-label={tr("st_manage_doc_number_formats", "จัดการรูปแบบเลขที่เอกสาร")}
          >
            <div className="flex shrink-0 items-center justify-between border-b px-5 py-3">
              <h3 className="flex items-center gap-2 text-base font-bold text-foreground">
                <FileText className="h-4 w-4 text-primary" /> {tr("st_manage_doc_number_formats", "จัดการรูปแบบเลขที่เอกสาร")}
              </h3>
              <button
                type="button"
                onClick={() => setDocDrawerOpen(false)}
                aria-label={tr("bill_close", "ปิด")}
                className="grid size-8 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto p-5">
              <DocFormatBuilder
                formats={formDocFormats}
                onChange={setFormDocFormats}
                branchCode={formCode || "00000"}
                disabled={isReadOnlyMode}
              />
            </div>
            <div className="flex shrink-0 justify-end border-t px-5 py-3">
              <Button type="button" onClick={() => setDocDrawerOpen(false)}>{tr("st_done", "เสร็จ")}</Button>
            </div>
          </div>
        </div>
      )}
      {confirmOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div
            role="dialog"
            aria-modal="true"
            aria-label={tr("st_confirm_data_save", "ยืนยันการบันทึกข้อมูล")}
            className="bg-card border rounded-2xl w-full max-w-sm p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in duration-200"
          >
            <div className="text-center space-y-2">
              <div className="w-12 h-12 rounded-full bg-primary/10 text-primary flex items-center justify-center mx-auto">
                <KeyRound size={22} className="animate-pulse" />
              </div>
              <h3 className="text-lg font-bold text-foreground">
                {tr("st_confirm_data_save", "ยืนยันการบันทึกข้อมูล")}
              </h3>
              <p className="text-xs text-muted-foreground">
                {tr("st_enter_4_digit_code_org_save", "กรุณากรอกรหัสยืนยันตัวเลข 4 หลักเพื่อดำเนินการบันทึกข้อมูลโครงสร้างองค์กร")}
              </p>
            </div>

            <div className="bg-accent/40 rounded-xl p-3 border border-border/80 text-center">
              <span className="text-xs font-semibold text-muted-foreground block mb-1">{tr("st_verification_code_is", "รหัสยืนยันของคุณคือ")}</span>
              <span className="text-2xl font-black tracking-widest text-primary font-mono select-none">{randomCode}</span>
            </div>

            <div className="space-y-1.5">
              <Input
                value={inputCode}
                onChange={(e) => {
                  setInputCode(e.target.value);
                  if (codeError) setCodeError(false);
                }}
                placeholder={tr("st_enter_4_digit_code_above", "กรอกรหัส 4 หลักที่แสดงด้านบน")}
                className={`bg-accent/20 h-11 text-center font-bold tracking-widest font-mono text-base ${codeError ? "border-destructive focus-visible:ring-destructive" : ""}`}
                maxLength={4}
              />
              {codeError && (
                <p className="text-[10px] text-destructive font-semibold text-center">{tr("st_invalid_verification_code_try_again", "รหัสยืนยันไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง")}</p>
              )}
            </div>

            <div className="flex gap-3 pt-2">
              <Button
                variant="outline"
                className="flex-1 rounded-xl h-11 text-xs font-semibold"
                onClick={closeConfirmCodeDialog}
              >
                {tr("cancel", "ยกเลิก")}
              </Button>
              <Button
                className="flex-1 rounded-xl h-11 text-xs font-semibold"
                onClick={handleConfirmCodeSubmit}
                disabled={inputCode.length !== 4}
              >
                {tr("st_confirm_save", "ยืนยันบันทึก")}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function saveErrorMessage(message: string | undefined, formType: OrganizationFormType, tr: BackendTextFn): string {
  void formType;
  switch (message) {
    case "branch code is required":
      return tr("st_enter_branch_code", "กรุณากรอกรหัสสาขา");
    case "branch code must be numeric and no more than 5 digits":
    case "branch code must be no more than 5 digits":
      return tr("st_branch_code_numeric_max_5", "รหัสสาขาต้องเป็นตัวเลขไม่เกิน 5 หลัก");
    case "companyuid is required":
    case "companyguid is required":
      return tr("st_company_not_found_for_branch_add", "ไม่พบบริษัทของสาขาที่กำลังเพิ่ม กรุณากดเพิ่มสาขาจากบริษัทอีกครั้ง");
    case "company not found":
      return tr("st_no_company_in_business_reload", "ไม่พบบริษัทในกิจการนี้ กรุณาโหลดข้อมูลใหม่แล้วลองอีกครั้ง");
    case "branch code is exists":
      return tr("st_branch_code_exists_in_company", "รหัสสาขานี้มีอยู่แล้วในบริษัทนี้");
    default:
      if (message?.includes("duplicate key")) return tr("st_duplicate_code", "รหัสนี้ซ้ำกับข้อมูลเดิม");
      return message || tr("st_save_data_failed", "บันทึกข้อมูลไม่สำเร็จ");
  }
}

"use client";

import { Fragment, type ReactNode } from "react";
import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DateField, TimeField } from "@/components/ui/date-time-field";
import type { LanguageCode } from "@/lib/i18n";
import type { CalendarYearType } from "@/lib/date-time";
import { normalizeTimeInput } from "@/lib/date-time";
import type {
  SystemSettingConfig,
  SystemSettingField,
} from "@/lib/system-setting-screens";
import {
  backendText,
  type BackendLanguageDictionary,
} from "@/lib/backend-language";
import {
  type DateTimeScope,
  type FormState,
  type SettingRecord,
  isRecord,
  safeJsonParse,
  stringValue,
} from "../types";

/* ------------------------------------------------------------------ */
/* Shared day names                                                    */
/* ------------------------------------------------------------------ */

export const dayNames: Record<LanguageCode, string[]> = {
  th: ["จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์", "อาทิตย์"],
  en: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"],
  cn: ["星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"],
  ja: ["月曜", "火曜", "水曜", "木曜", "金曜", "土曜", "日曜"],
  ko: ["월요일", "화요일", "수요일", "목요일", "금요일", "토요일", "일요일"],
  lo: ["ຈັນ", "ອັງຄານ", "ພຸດ", "ພະຫັດ", "ສຸກ", "ເສົາ", "ອາທິດ"],
  my: ["တနင်္လာ", "အင်္ဂါ", "ဗုဒ္ဓဟူး", "ကြာသပတေး", "သောကြာ", "စနေ", "တနင်္ဂနွေ"],
  km: ["ចន្ទ", "អង្គារ", "ពុធ", "ព្រហស្បតិ៍", "សុក្រ", "សៅរ៍", "អាទិត្យ"],
  vi: ["Thứ hai", "Thứ ba", "Thứ tư", "Thứ năm", "Thứ sáu", "Thứ bảy", "Chủ nhật"],
  ms: ["Isnin", "Selasa", "Rabu", "Khamis", "Jumaat", "Sabtu", "Ahad"],
  id: ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"],
  fil: ["Lunes", "Martes", "Miyerkules", "Huwebes", "Biyernes", "Sabado", "Linggo"],
};

/* ------------------------------------------------------------------ */
/* Product Variant Structured Field                                    */
/* ------------------------------------------------------------------ */

type VariantColumn = {
  key: string;
  labelTh: string;
  labelEn: string;
  kind?: "number" | "csv";
  placeholder?: string;
};

const variantFieldColumns: Record<string, VariantColumn[]> = {
  optiontiers: [
    { key: "tierno", labelTh: "ลำดับ", labelEn: "Tier", kind: "number" },
    { key: "optioncode", labelTh: "รหัสแกน", labelEn: "Option code", placeholder: "COLOR" },
    { key: "name", labelTh: "ชื่อแกน", labelEn: "Option name", placeholder: "สี/Color" },
  ],
  skucombinations: [
    { key: "sellersku", labelTh: "SKU", labelEn: "SKU" },
    { key: "barcode", labelTh: "บาร์โค้ด", labelEn: "Barcode" },
    { key: "gtin", labelTh: "GTIN", labelEn: "GTIN" },
    { key: "optionvalues", labelTh: "ค่าตัวเลือก", labelEn: "Options", kind: "csv", placeholder: "BLACK, 128GB" },
    { key: "saleprice", labelTh: "ราคาขาย", labelEn: "Price", kind: "number" },
    { key: "cost", labelTh: "ต้นทุน", labelEn: "Cost", kind: "number" },
    { key: "openingstock", labelTh: "สต๊อกต้น", labelEn: "Opening stock", kind: "number" },
  ],
  mediaassets: [
    { key: "kind", labelTh: "ชนิดสื่อ", labelEn: "Media type", placeholder: "main / gallery / sku / video" },
    { key: "uri", labelTh: "ที่อยู่ไฟล์", labelEn: "File path" },
    { key: "optioncode", labelTh: "รหัสแกน", labelEn: "Option code" },
    { key: "optionvalue", labelTh: "ค่าตัวเลือก", labelEn: "Option value" },
    { key: "sortorder", labelTh: "ลำดับ", labelEn: "Sort", kind: "number" },
  ],
  specificationgroups: [
    { key: "groupcode", labelTh: "รหัสกลุ่ม", labelEn: "Group code" },
    { key: "groupname", labelTh: "ชื่อกลุ่ม", labelEn: "Group name" },
  ],
  importattributemaps: [
    { key: "sourcename", labelTh: "ชื่อจากไฟล์นำเข้า", labelEn: "Imported name" },
    { key: "targetoptioncode", labelTh: "รหัสแกนในระบบ", labelEn: "System option code" },
  ],
  integrationprofiles: [
    { key: "channel", labelTh: "ช่องทาง", labelEn: "Channel", placeholder: "shopee / lazada / external" },
    { key: "skufields", labelTh: "ช่อง SKU", labelEn: "SKU fields", kind: "csv", placeholder: "sellersku, barcode, price, stock" },
    { key: "media_keys", labelTh: "ช่องรูป/วิดีโอ", labelEn: "Media keys", kind: "csv" },
    { key: "price_keys", labelTh: "ช่องราคา", labelEn: "Price keys", kind: "csv" },
    { key: "stock_keys", labelTh: "ช่องสต๊อก", labelEn: "Stock keys", kind: "csv" },
  ],
  payloadexamples: [
    { key: "direction", labelTh: "ทิศทาง", labelEn: "Direction", placeholder: "import / export" },
    { key: "usecase", labelTh: "กรณีใช้งาน", labelEn: "Use case" },
    { key: "channel", labelTh: "ช่องทาง", labelEn: "Channel" },
    { key: "note", labelTh: "หมายเหตุ", labelEn: "Note" },
  ],
};

function variantArrayValue(value: unknown): SettingRecord[] {
  const source =
    typeof value === "string" ? safeJsonParse(value.trim() || "[]", []) : value;
  if (!Array.isArray(source)) return [];
  return source.map((item) => (isRecord(item) ? item : { value: item }));
}

function defaultVariantItem(key: string, index: number): SettingRecord {
  if (key === "optiontiers") return { tierno: index + 1, values: [] };
  if (key === "specificationgroups") return { attributes: [] };
  return {};
}

export function ProductVariantStructuredFieldEditor({
  field,
  form,
  label,
  language,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const items = variantArrayValue(form[field.key]);
  const columns = variantFieldColumns[field.key] ?? [];
  const addLabel = language === "th" ? "เพิ่มรายการ" : "Add item";
  const emptyText =
    language === "th"
      ? "ยังไม่มีรายการ กดเพิ่มรายการเพื่อเริ่มกรอก"
      : "No items yet. Add an item to start.";

  const setItems = (nextItems: SettingRecord[]) => {
    setForm({ ...form, [field.key]: nextItems });
  };

  const updateItem = (index: number, key: string, value: unknown) => {
    setItems(
      items.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [key]: value } : item,
      ),
    );
  };
  const moveItem = (index: number, direction: -1 | 1) => {
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= items.length) return;
    const nextItems = [...items];
    [nextItems[index], nextItems[nextIndex]] = [
      nextItems[nextIndex],
      nextItems[index],
    ];
    setItems(
      field.key === "optiontiers"
        ? nextItems.map((item, itemIndex) => ({
            ...item,
            tierno: itemIndex + 1,
          }))
        : nextItems,
    );
  };

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <header className="flex flex-wrap items-start justify-between gap-2 border-b border-border pb-2">
        <div className="grid gap-0.5">
          <h3 className="text-sm font-semibold">
            {label}
            {field.required ? " *" : ""}
          </h3>
          {field.helper ? (
            <p className="text-xs leading-snug text-muted-foreground">
              {field.helper[language] ?? field.helper.en ?? field.helper.th}
            </p>
          ) : null}
        </div>
        <Button
          className="h-8 gap-1 rounded-xl px-2 text-xs"
          type="button"
          variant="outline"
          onClick={() => setItems([...items, defaultVariantItem(field.key, items.length)])}
        >
          <Plus className="size-3.5" />
          {addLabel}
        </Button>
      </header>
      {items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-3 text-xs text-muted-foreground">
          {emptyText}
        </div>
      ) : (
        <div className="grid gap-2">
          {items.map((item, index) => (
            <article
              className="grid gap-2 rounded-xl border border-border bg-muted/20 p-2"
              key={`${field.key}-${index}`}
            >
              <div className="flex items-center justify-between gap-2">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant="secondary" className="rounded-lg text-[11px]">
                    {field.key === "optiontiers"
                      ? language === "th"
                        ? `ลำดับเลือก ${index + 1}`
                        : `Choice order ${index + 1}`
                      : language === "th"
                        ? `รายการ ${index + 1}`
                        : `Item ${index + 1}`}
                  </Badge>
                  {field.key === "optiontiers" ? (
                    <span className="text-[11px] font-medium text-muted-foreground">
                      {language === "th"
                        ? "ผู้ใช้เลือกตามลำดับนี้ เช่น สีก่อนไซซ์ หรือไซซ์ก่อนสี"
                        : "Users choose in this order, such as color before size or size before color."}
                    </span>
                  ) : null}
                </div>
                <div className="flex items-center gap-1">
                  {field.key === "optiontiers" ? (
                    <>
                      <Button
                        aria-label={language === "th" ? "เลื่อนขึ้น" : "Move up"}
                        className="size-8 rounded-xl p-0"
                        disabled={index === 0}
                        type="button"
                        variant="outline"
                        onClick={() => moveItem(index, -1)}
                      >
                        <ArrowUp className="size-4" />
                      </Button>
                      <Button
                        aria-label={language === "th" ? "เลื่อนลง" : "Move down"}
                        className="size-8 rounded-xl p-0"
                        disabled={index === items.length - 1}
                        type="button"
                        variant="outline"
                        onClick={() => moveItem(index, 1)}
                      >
                        <ArrowDown className="size-4" />
                      </Button>
                    </>
                  ) : null}
                  <Button
                    aria-label={language === "th" ? "ลบรายการ" : "Remove item"}
                    className="size-8 rounded-xl p-0 text-destructive"
                    type="button"
                    variant="outline"
                    onClick={() =>
                      setItems(items.filter((_, itemIndex) => itemIndex !== index))
                    }
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              </div>
              <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                {columns.map((column) => (
                  <VariantInput
                    column={column}
                    key={column.key}
                    language={language}
                    value={item[column.key]}
                    onChange={(nextValue) => updateItem(index, column.key, nextValue)}
                  />
                ))}
              </div>
              {field.key === "optiontiers" ? (
                <VariantNestedRowsEditor
                  columns={[
                    { key: "valuecode", labelTh: "รหัสค่า", labelEn: "Value code" },
                    { key: "valuetext", labelTh: "ชื่อค่า", labelEn: "Value name" },
                  ]}
                  items={variantArrayValue(item.values)}
                  language={language}
                  title={language === "th" ? "ค่าของแกนนี้" : "Option values"}
                  onChange={(nextRows) => updateItem(index, "values", nextRows)}
                />
              ) : null}
              {field.key === "specificationgroups" ? (
                <VariantNestedRowsEditor
                  columns={[
                    { key: "attributecode", labelTh: "รหัสคุณสมบัติ", labelEn: "Attribute code" },
                    { key: "attributename", labelTh: "ชื่อคุณสมบัติ", labelEn: "Attribute name" },
                    { key: "inputtype", labelTh: "ชนิดช่องกรอก", labelEn: "Input type" },
                    { key: "scope", labelTh: "ระดับข้อมูล", labelEn: "Scope" },
                  ]}
                  items={variantArrayValue(item.attributes)}
                  language={language}
                  title={language === "th" ? "คุณสมบัติในกลุ่มนี้" : "Attributes"}
                  onChange={(nextRows) => updateItem(index, "attributes", nextRows)}
                />
              ) : null}
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

function VariantNestedRowsEditor({
  columns,
  items,
  language,
  onChange,
  title,
}: {
  columns: VariantColumn[];
  items: SettingRecord[];
  language: LanguageCode;
  onChange: (items: SettingRecord[]) => void;
  title: string;
}) {
  const updateRow = (index: number, key: string, value: unknown) => {
    onChange(
      items.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [key]: value } : item,
      ),
    );
  };
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-background/80 p-2">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-semibold text-muted-foreground">
          {title}
        </span>
        <Button
          className="h-7 gap-1 rounded-lg px-2 text-[11px]"
          type="button"
          variant="outline"
          onClick={() => onChange([...items, {}])}
        >
          <Plus className="size-3" />
          {language === "th" ? "เพิ่ม" : "Add"}
        </Button>
      </div>
      {items.length === 0 ? (
        <p className="text-xs text-muted-foreground">
          {language === "th" ? "ยังไม่มีรายการย่อย" : "No rows yet."}
        </p>
      ) : (
        <div className="grid gap-2">
          {items.map((item, index) => (
            <div
              className="grid gap-2 rounded-lg border border-border bg-muted/20 p-2 md:grid-cols-[1fr_1fr_auto]"
              key={`nested-${index}`}
            >
              {columns.map((column) => (
                <VariantInput
                  column={column}
                  key={column.key}
                  language={language}
                  value={item[column.key]}
                  onChange={(nextValue) => updateRow(index, column.key, nextValue)}
                />
              ))}
              <Button
                aria-label={language === "th" ? "ลบรายการย่อย" : "Remove row"}
                className="size-8 self-end rounded-xl p-0 text-destructive"
                type="button"
                variant="outline"
                onClick={() => onChange(items.filter((_, itemIndex) => itemIndex !== index))}
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function VariantInput({
  column,
  language,
  onChange,
  value,
}: {
  column: VariantColumn;
  language: LanguageCode;
  onChange: (value: unknown) => void;
  value: unknown;
}) {
  const label = language === "th" ? column.labelTh : column.labelEn;
  const inputValue =
    column.kind === "csv" && Array.isArray(value)
      ? value.map((item) => stringValue(item)).join(", ")
      : stringValue(value);
  return (
    <label className="grid min-w-0 gap-1 text-xs font-semibold">
      <span className="text-muted-foreground">{label}</span>
      <Input
        className="h-9"
        inputMode={column.kind === "number" ? "decimal" : undefined}
        placeholder={column.placeholder}
        type={column.kind === "number" ? "number" : "text"}
        value={inputValue}
        onChange={(event) => {
          const nextValue = event.target.value;
          if (column.kind === "number") {
            onChange(nextValue === "" ? "" : Number(nextValue));
            return;
          }
          if (column.kind === "csv") {
            onChange(
              nextValue
                .split(",")
                .map((item) => item.trim())
                .filter(Boolean),
            );
            return;
          }
          onChange(nextValue);
        }}
      />
    </label>
  );
}

export function ProductVariantStructuredReadOnlyDetail({
  field,
  label,
  language,
  value,
}: {
  field: SystemSettingField;
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const items = variantArrayValue(value);
  const columns = variantFieldColumns[field.key] ?? [];
  return (
    <section className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)]">
      <header className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-2">
        <span className="text-xs font-semibold text-muted-foreground">
          {label}
        </span>
        <Badge variant="secondary" className="rounded-lg text-[11px]">
          {items.length} {language === "th" ? "รายการ" : "items"}
        </Badge>
      </header>
      {items.length === 0 ? (
        <span className="text-sm font-medium text-muted-foreground">-</span>
      ) : (
        <div className="grid gap-2">
          {items.map((item, index) => (
            <article
              className="grid gap-2 rounded-lg border border-border bg-muted/20 p-2"
              key={`${field.key}-readonly-${index}`}
            >
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="secondary" className="rounded-lg text-[11px]">
                  {field.key === "optiontiers"
                    ? language === "th"
                      ? `ลำดับเลือก ${index + 1}`
                      : `Choice order ${index + 1}`
                    : language === "th"
                      ? `รายการ ${index + 1}`
                      : `Item ${index + 1}`}
                </Badge>
                {field.key === "optiontiers" ? (
                  <span className="text-[11px] text-muted-foreground">
                    {language === "th"
                      ? "ใช้กำหนดว่าผู้ใช้เลือกสี/ไซซ์/ตัวเลือกใดก่อนหลัง"
                      : "Controls which option tier users choose first."}
                  </span>
                ) : null}
              </div>
              <div className="grid gap-1 md:grid-cols-2 xl:grid-cols-3">
                {columns.map((column) => {
                  const cellValue =
                    column.kind === "csv" && Array.isArray(item[column.key])
                      ? (item[column.key] as unknown[]).map(stringValue).join(", ")
                      : stringValue(item[column.key]);
                  return (
                    <div className="min-w-0" key={column.key}>
                      <span className="text-[11px] font-medium text-muted-foreground">
                        {language === "th" ? column.labelTh : column.labelEn}
                      </span>
                      <p className="truncate text-xs font-semibold">
                        {cellValue || "-"}
                      </p>
                    </div>
                  );
                })}
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

/* ------------------------------------------------------------------ */
/* TimeSaleListEditor                                                  */
/* ------------------------------------------------------------------ */

export type TimeSaleRow = {
  daysofweek: number[];
  fromdate: string;
  todate: string;
  fromtime: string;
  totime: string;
};

function emptyTimeSaleRow(): TimeSaleRow {
  return {
    daysofweek: [1, 2, 3, 4, 5, 6, 7],
    fromdate: "",
    todate: "",
    fromtime: "",
    totime: "",
  };
}

function normalizeTimeSaleDays(value: unknown): number[] {
  const values = Array.isArray(value) ? value : [];
  const normalized = values
    .map((item) => Number(item))
    .filter((item) => Number.isInteger(item) && item >= 1 && item <= 7);
  return Array.from(new Set(normalized)).sort((a, b) => a - b);
}

function timeSaleDateInputValue(value: unknown): string {
  const raw = stringValue(value);
  if (!raw) return "";
  const match = raw.match(/^(\d{4}-\d{2}-\d{2})/);
  return match?.[1] ?? "";
}

function timeSaleDateInputToIso(value: unknown): string {
  const raw = timeSaleDateInputValue(value);
  if (!raw) return "";
  return new Date(`${raw}T00:00:00.000Z`).toISOString();
}

export function normalizeTimeSaleFormList(value: unknown): TimeSaleRow[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item) => ({
    daysofweek: normalizeTimeSaleDays(item.daysofweek),
    fromdate: timeSaleDateInputValue(item.fromdate),
    todate: timeSaleDateInputValue(item.todate),
    fromtime: normalizeTimeInput(stringValue(item.fromtime)),
    totime: normalizeTimeInput(stringValue(item.totime)),
  }));
}

export function normalizeTimeSalePayload(value: unknown): TimeSaleRow[] {
  return normalizeTimeSaleFormList(value).map((item) => ({
    daysofweek: item.daysofweek,
    fromdate: timeSaleDateInputToIso(item.fromdate),
    todate: timeSaleDateInputToIso(item.todate),
    fromtime: item.fromtime,
    totime: item.totime,
  }));
}

export function TimeSaleListEditor({
  dateTimeScope,
  field,
  form,
  label,
  language,
  setForm,
}: {
  dateTimeScope: DateTimeScope;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const rows = normalizeTimeSaleFormList(form[field.key]);
  const enabled = rows.length > 0;

  function commit(nextRows: TimeSaleRow[]) {
    setForm({ ...form, [field.key]: nextRows });
  }

  function updateRow(index: number, patch: Partial<TimeSaleRow>) {
    commit(rows.map((row, rowIndex) => (rowIndex === index ? { ...row, ...patch } : row)));
  }

  function toggleEnabled(checked: boolean) {
    commit(checked ? [emptyTimeSaleRow()] : []);
  }

  function toggleDay(index: number, day: number, checked: boolean) {
    const current = new Set(rows[index]?.daysofweek ?? []);
    if (checked) current.add(day);
    else current.delete(day);
    updateRow(index, { daysofweek: Array.from(current).sort((a, b) => a - b) });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold">
      <label className="flex min-h-10 items-center gap-2">
        <input
          className="size-4 accent-primary"
          type="checkbox"
          checked={enabled}
          onChange={(event) => toggleEnabled(event.target.checked)}
        />
        <span>
          {label}
          {enabled ? ` (${rows.length})` : ""}
        </span>
      </label>
      {enabled ? (
        <div className="grid gap-2">
          {rows.map((row, index) => (
            <div
              className="grid gap-2 rounded-xl border border-border bg-card p-2"
              key={index}
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <b>
                  {language === "th" ? "ช่วงเวลา" : "Time window"} {index + 1}
                </b>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8"
                  onClick={() => commit(rows.filter((_, rowIndex) => rowIndex !== index))}
                >
                  <Trash2 className="size-3.5" />
                  {language === "th" ? "ลบ" : "Delete"}
                </Button>
              </div>
              <div className="grid gap-2 md:grid-cols-2">
                <DateField
                  language={language}
                  yearType={dateTimeScope.calendarYearType}
                  label={language === "th" ? "วันที่เริ่ม" : "From date"}
                  value={row.fromdate}
                  onChange={(event) => updateRow(index, { fromdate: event.target.value })}
                />
                <DateField
                  language={language}
                  yearType={dateTimeScope.calendarYearType}
                  label={language === "th" ? "วันที่สิ้นสุด" : "To date"}
                  value={row.todate}
                  onChange={(event) => updateRow(index, { todate: event.target.value })}
                />
                <TimeField
                  label={language === "th" ? "เวลาเริ่ม" : "From time"}
                  value={row.fromtime}
                  onChange={(event) => updateRow(index, { fromtime: normalizeTimeInput(event.target.value) })}
                />
                <TimeField
                  label={language === "th" ? "เวลาสิ้นสุด" : "To time"}
                  value={row.totime}
                  onChange={(event) => updateRow(index, { totime: normalizeTimeInput(event.target.value) })}
                />
              </div>
              <div className="flex flex-wrap gap-1">
                {dayNames[language].map((dayName, dayIndex) => {
                  const day = dayIndex + 1;
                  const checked = row.daysofweek.includes(day);
                  return (
                    <label
                      className={cn(
                        "inline-flex min-h-8 cursor-pointer items-center gap-1 rounded-lg border px-2 text-xs font-semibold",
                        checked
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border bg-background text-muted-foreground",
                      )}
                      key={day}
                    >
                      <input
                        className="size-3 accent-primary"
                        type="checkbox"
                        checked={checked}
                        onChange={(event) => toggleDay(index, day, event.target.checked)}
                      />
                      {dayName}
                    </label>
                  );
                })}
              </div>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            className="h-9 justify-center"
            onClick={() => commit([...rows, emptyTimeSaleRow()])}
          >
            <Plus className="size-4" />
            {language === "th" ? "เพิ่มช่วงเวลา" : "Add time window"}
          </Button>
        </div>
      ) : null}
    </section>
  );
}

export function TimeSaleListReadOnlyDetail({
  label,
  language,
  value,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const rows = normalizeTimeSaleFormList(value);
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)]">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {rows.length === 0 ? (
        <b className="text-foreground">-</b>
      ) : (
        <div className="grid gap-1">
          {rows.map((row, index) => (
            <div className="rounded-lg border border-border bg-background p-2" key={index}>
              <b>
                {language === "th" ? "ช่วงเวลา" : "Time window"} {index + 1}
              </b>
              <p className="text-xs font-medium text-muted-foreground">
                {row.fromdate || "-"} {row.fromtime || "--:--"} - {row.todate || "-"} {row.totime || "--:--"}
              </p>
              <p className="text-xs font-medium text-muted-foreground">
                {(row.daysofweek.length ? row.daysofweek : [1, 2, 3, 4, 5, 6, 7])
                  .map((day) => dayNames[language][day - 1])
                  .filter(Boolean)
                  .join(", ")}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/* ------------------------------------------------------------------ */
/* BankAccountsEditor                                                  */
/* ------------------------------------------------------------------ */

export type BankAccountRow = { bankcode: string; accountnumber: string; accountname: string };

function emptyBankAccountRow(): BankAccountRow {
  return { bankcode: "", accountnumber: "", accountname: "" };
}

export function normalizeBankAccountFormList(value: unknown): BankAccountRow[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item) => ({
    bankcode: stringValue(item.bankcode),
    accountnumber: stringValue(item.accountnumber),
    accountname: stringValue(item.accountname),
  }));
}

export function normalizeBankAccountPayload(value: unknown): BankAccountRow[] {
  return normalizeBankAccountFormList(value).filter(
    (row) => row.bankcode || row.accountnumber || row.accountname,
  );
}

export function BankAccountsEditor({
  field,
  form,
  language,
  label,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  label: string;
  setForm: (form: FormState) => void;
}) {
  const rows = normalizeBankAccountFormList(form[field.key]);

  function commit(nextRows: BankAccountRow[]) {
    setForm({ ...form, [field.key]: nextRows });
  }

  function updateRow(index: number, patch: Partial<BankAccountRow>) {
    commit(rows.map((row, rowIndex) => (rowIndex === index ? { ...row, ...patch } : row)));
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <span>
        {label}
        {rows.length ? ` (${rows.length})` : ""}
      </span>
      <div className="grid gap-2">
        {rows.map((row, index) => (
          <div
            className="grid gap-2 rounded-xl border border-border bg-card p-2 md:grid-cols-[repeat(3,1fr)_auto]"
            key={index}
          >
            <label className="grid gap-1">
              <span className="text-xs font-medium text-muted-foreground">
                {language === "th" ? "ธนาคาร" : "Bank"}
              </span>
              <Input
                value={row.bankcode}
                onChange={(event) => updateRow(index, { bankcode: event.target.value })}
              />
            </label>
            <label className="grid gap-1">
              <span className="text-xs font-medium text-muted-foreground">
                {language === "th" ? "เลขที่บัญชี" : "Account number"}
              </span>
              <Input
                value={row.accountnumber}
                onChange={(event) => updateRow(index, { accountnumber: event.target.value })}
              />
            </label>
            <label className="grid gap-1">
              <span className="text-xs font-medium text-muted-foreground">
                {language === "th" ? "ชื่อบัญชี" : "Account name"}
              </span>
              <Input
                value={row.accountname}
                onChange={(event) => updateRow(index, { accountname: event.target.value })}
              />
            </label>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="h-9 self-end"
              onClick={() => commit(rows.filter((_, rowIndex) => rowIndex !== index))}
            >
              <Trash2 className="size-3.5" />
              {language === "th" ? "ลบ" : "Delete"}
            </Button>
          </div>
        ))}
        <Button
          type="button"
          variant="outline"
          className="h-9 justify-center"
          onClick={() => commit([...rows, emptyBankAccountRow()])}
        >
          <Plus className="size-4" />
          {language === "th" ? "เพิ่มบัญชีธนาคาร" : "Add bank account"}
        </Button>
      </div>
    </section>
  );
}

export function BankAccountsReadOnlyDetail({
  label,
  language,
  value,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const rows = normalizeBankAccountFormList(value);
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)]">
      <span className="text-xs font-semibold text-muted-foreground">{label}</span>
      {rows.length === 0 ? (
        <b className="text-foreground">-</b>
      ) : (
        <div className="grid gap-1">
          {rows.map((row, index) => (
            <div className="rounded-lg border border-border bg-background p-2" key={index}>
              <b>{row.bankcode || "-"}</b>{" "}
              <span className="text-xs font-medium text-muted-foreground">
                {row.accountnumber || "-"} · {row.accountname || "-"}
              </span>
              {index === 0 ? (
                <span className="ml-1 text-xs font-medium text-muted-foreground">
                  ({language === "th" ? "บัญชีหลัก" : "default"})
                </span>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function normalizeHexColor(value: unknown): string {
  const raw = stringValue(value).replace(/^#/, "");
  if (/^[0-9a-fA-F]{6}$/.test(raw)) return `#${raw}`;
  if (/^[0-9a-fA-F]{8}$/.test(raw)) return `#${raw.slice(2)}`;
  return "#ffffff";
}

/* ------------------------------------------------------------------ */
/* Payment Rounding + Point Config types & helpers                     */
/* ------------------------------------------------------------------ */

export const paymentMethodKeys = [
  "cash",
  "creditcard",
  "banktransfer",
  "cheque",
  "coupon",
  "delivery",
  "qrcode",
] as const;

export type PaymentMethodKey = (typeof paymentMethodKeys)[number];
export type RoundingRule = {
  lowerbound: number;
  upperbound: number;
  roundto: number;
};
export type PaymentMethodRounding = {
  enabled: boolean;
  rules: RoundingRule[];
};
export type PaymentRoundingConfig = Record<PaymentMethodKey, PaymentMethodRounding>;
export type GeneralPointRule = {
  startdate: string;
  enddate: string;
  payperpoint: number;
  pointvalue: number;
};
export type SpecialPointRule = {
  startdate: string;
  enddate: string;
  multiplier: number;
  sunday: boolean;
  monday: boolean;
  tuesday: boolean;
  wednesday: boolean;
  thursday: boolean;
  friday: boolean;
  saturday: boolean;
  maxpointperbill: number;
};
export type PointConfig = {
  generalrules: GeneralPointRule[];
  specialrules: SpecialPointRule[];
  pointusagetype: number;
};
export type BranchStructuredParseResult<T> = {
  isRecovered: boolean;
  value: T;
};

const pointWeekDays = [
  "monday",
  "tuesday",
  "wednesday",
  "thursday",
  "friday",
  "saturday",
  "sunday",
] as const;

type PointWeekDay = (typeof pointWeekDays)[number];

function normalizeConfigNumber(value: unknown, fallback: number): number {
  const next = Number(value);
  return Number.isFinite(next) ? next : fallback;
}

function numberFromInput(value: string): number {
  const next = Number(value);
  return Number.isFinite(next) ? next : 0;
}

function formatConfigNumber(value: number): string {
  return Number.isFinite(value) ? String(value) : "";
}

function normalizeIsoDate(value: unknown, fallback: string): string {
  if (typeof value !== "string" || !value.trim()) return fallback;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? fallback : date.toISOString();
}

function isoDateToInput(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString().slice(0, 10);
}

function dateInputToIso(value: string, endOfDay?: boolean): string {
  if (!value) return "";
  const suffix = endOfDay ? "T23:59:59" : "T00:00:00";
  return new Date(`${value}${suffix}`).toISOString();
}

function invalidConfigText(language: LanguageCode): string {
  return language === "th"
    ? "ข้อมูลเดิมไม่ถูกต้อง ใช้ค่าเริ่มต้น"
    : "Invalid saved config; default values are shown";
}

export function defaultPaymentRoundingJson(): string {
  const defaultRules = [
    { lowerbound: 0.01, upperbound: 0.12, roundto: 0 },
    { lowerbound: 0.13, upperbound: 0.37, roundto: 0.25 },
    { lowerbound: 0.38, upperbound: 0.62, roundto: 0.5 },
    { lowerbound: 0.63, upperbound: 0.87, roundto: 0.75 },
    { lowerbound: 0.88, upperbound: 0.99, roundto: 1 },
  ];
  const method = { enabled: true, rules: defaultRules };
  return JSON.stringify(
    {
      banktransfer: method,
      cash: method,
      cheque: method,
      coupon: method,
      creditcard: method,
      delivery: method,
      qrcode: method,
    },
    null,
    2,
  );
}

export function defaultPointConfigJson(): string {
  return JSON.stringify(
    {
      generalrules: [],
      specialrules: [],
      pointusagetype: 1,
    },
    null,
    2,
  );
}

function defaultRoundingRules(): RoundingRule[] {
  return [
    { lowerbound: 0.01, upperbound: 0.12, roundto: 0 },
    { lowerbound: 0.13, upperbound: 0.37, roundto: 0.25 },
    { lowerbound: 0.38, upperbound: 0.62, roundto: 0.5 },
    { lowerbound: 0.63, upperbound: 0.87, roundto: 0.75 },
    { lowerbound: 0.88, upperbound: 0.99, roundto: 1 },
  ];
}

function defaultRoundingRule(): RoundingRule {
  return { lowerbound: 0.01, upperbound: 0.99, roundto: 0 };
}

function defaultGeneralPointRule(): GeneralPointRule {
  const start = new Date();
  const end = new Date(start);
  end.setDate(end.getDate() + 365);
  return {
    startdate: start.toISOString(),
    enddate: end.toISOString(),
    payperpoint: 20,
    pointvalue: 1,
  };
}

function defaultSpecialPointRule(): SpecialPointRule {
  const start = new Date();
  const end = new Date(start);
  end.setDate(end.getDate() + 30);
  return {
    startdate: start.toISOString(),
    enddate: end.toISOString(),
    multiplier: 2,
    sunday: false,
    monday: true,
    tuesday: true,
    wednesday: true,
    thursday: true,
    friday: true,
    saturday: false,
    maxpointperbill: 100,
  };
}

function parseBranchSettingRecord(
  value: unknown,
  fallbackJson: string,
): { isRecovered: boolean; record: SettingRecord } {
  const fallback = JSON.parse(fallbackJson) as SettingRecord;
  if (isRecord(value)) return { isRecovered: false, record: value };
  if (typeof value !== "string" || !value.trim())
    return { isRecovered: false, record: fallback };
  try {
    const parsed = JSON.parse(value) as unknown;
    if (isRecord(parsed)) return { isRecovered: false, record: parsed };
  } catch {
    return { isRecovered: true, record: fallback };
  }
  return { isRecovered: true, record: fallback };
}

function normalizeRoundingRule(value: unknown): RoundingRule {
  const record = isRecord(value) ? value : {};
  return {
    lowerbound: normalizeConfigNumber(record.lowerbound, 0),
    upperbound: normalizeConfigNumber(record.upperbound, 0),
    roundto: normalizeConfigNumber(record.roundto, 0),
  };
}

function normalizeGeneralPointRule(value: unknown): GeneralPointRule {
  const record = isRecord(value) ? value : {};
  const fallback = defaultGeneralPointRule();
  return {
    startdate: normalizeIsoDate(record.startdate, fallback.startdate),
    enddate: normalizeIsoDate(record.enddate, fallback.enddate),
    payperpoint: normalizeConfigNumber(record.payperpoint, 20),
    pointvalue: normalizeConfigNumber(record.pointvalue, 1),
  };
}

function normalizeSpecialPointRule(value: unknown): SpecialPointRule {
  const record = isRecord(value) ? value : {};
  const fallback = defaultSpecialPointRule();
  return {
    startdate: normalizeIsoDate(record.startdate, fallback.startdate),
    enddate: normalizeIsoDate(record.enddate, fallback.enddate),
    multiplier: normalizeConfigNumber(record.multiplier, 2),
    sunday: Boolean(record.sunday),
    monday: record.monday === undefined ? true : Boolean(record.monday),
    tuesday: record.tuesday === undefined ? true : Boolean(record.tuesday),
    wednesday: record.wednesday === undefined ? true : Boolean(record.wednesday),
    thursday: record.thursday === undefined ? true : Boolean(record.thursday),
    friday: record.friday === undefined ? true : Boolean(record.friday),
    saturday: Boolean(record.saturday),
    maxpointperbill: normalizeConfigNumber(record.maxpointperbill, 100),
  };
}

export function normalizePaymentRoundingConfig(
  value: unknown,
): BranchStructuredParseResult<PaymentRoundingConfig> {
  const parsed = parseBranchSettingRecord(value, defaultPaymentRoundingJson());
  const source = parsed.record;
  return {
    isRecovered: parsed.isRecovered,
    value: Object.fromEntries(
      paymentMethodKeys.map((method) => {
        const methodValue = isRecord(source[method]) ? source[method] : {};
        const rules = Array.isArray(methodValue.rules)
          ? methodValue.rules.map(normalizeRoundingRule)
          : defaultRoundingRules();
        return [
          method,
          {
            enabled:
              typeof methodValue.enabled === "boolean"
                ? methodValue.enabled
                : true,
            rules: rules.length ? rules : [defaultRoundingRule()],
          },
        ];
      }),
    ) as PaymentRoundingConfig,
  };
}

export function normalizePointConfig(
  value: unknown,
): BranchStructuredParseResult<PointConfig> {
  const parsed = parseBranchSettingRecord(value, defaultPointConfigJson());
  const source = parsed.record;
  const pointUsageType = Number(source.pointusagetype);
  return {
    isRecovered: parsed.isRecovered,
    value: {
      generalrules: Array.isArray(source.generalrules)
        ? source.generalrules.map(normalizeGeneralPointRule)
        : [],
      specialrules: Array.isArray(source.specialrules)
        ? source.specialrules.map(normalizeSpecialPointRule)
        : [],
      pointusagetype: pointUsageType === 2 ? 2 : 1,
    },
  };
}

function paymentRoundingSummary(
  value: PaymentRoundingConfig,
  language: LanguageCode,
): string {
  const methods = paymentMethodKeys.map((method) => value[method]);
  const enabledCount = methods.filter((method) => Boolean(method.enabled)).length;
  const maxRules = Math.max(
    0,
    ...methods.map((method) => method.rules.length),
  );
  return language === "th"
    ? `เปิดใช้ ${enabledCount}/${methods.length} ช่องทาง, กฎสูงสุด ${maxRules} รายการ`
    : `Enabled ${enabledCount}/${methods.length} methods, max ${maxRules} rules`;
}

function pointConfigSummary(value: PointConfig, language: LanguageCode): string {
  return language === "th"
    ? `กฎทั่วไป ${value.generalrules.length} รายการ, กฎพิเศษ ${value.specialrules.length} รายการ, ประเภทใช้แต้ม ${value.pointusagetype}`
    : `${value.generalrules.length} general rules, ${value.specialrules.length} special rules, usage type ${value.pointusagetype}`;
}

function paymentMethodLabel(
  method: PaymentMethodKey,
  dictionary: BackendLanguageDictionary,
): string {
  const keys: Record<PaymentMethodKey, [string, string]> = {
    banktransfer: ["bank_transfer", "การโอนเงิน"],
    cash: ["cash", "เงินสด"],
    cheque: ["cheque", "เช็ค"],
    coupon: ["coupon", "คูปอง"],
    creditcard: ["creditcard", "บัตรเครดิต"],
    delivery: ["delivery", "เดลิเวอรี่"],
    qrcode: ["qrcode", "คิวอาร์โค้ด"],
  };
  const [key, fallback] = keys[method];
  return backendText(dictionary, key, fallback);
}

function shortDayLabel(
  day: PointWeekDay,
  dictionary: BackendLanguageDictionary,
): string {
  const fallback: Record<PointWeekDay, string> = {
    friday: "ศุกร์",
    monday: "จันทร์",
    saturday: "เสาร์",
    sunday: "อาทิตย์",
    thursday: "พฤหัส",
    tuesday: "อังคาร",
    wednesday: "พุธ",
  };
  return backendText(dictionary, day, fallback[day]);
}

export function isBranchStructuredSettingField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    config.slug === "branch" &&
    field.type === "json" &&
    (field.key === "paymentrounding" || field.key === "pointconfig")
  );
}

/* ------------------------------------------------------------------ */
/* Small helper components                                             */
/* ------------------------------------------------------------------ */

function PaymentMethodToggle({
  checked,
  disabled = false,
  label,
  onChange,
  text,
}: {
  checked: boolean;
  disabled?: boolean;
  label: string;
  onChange: (checked: boolean) => void;
  text: (key: string, fallback: string) => string;
}) {
  return (
    <div className="grid gap-1">
      <span className="font-semibold">{label}</span>
      <label className="inline-flex w-auto max-w-none items-center gap-2 text-xs font-medium text-muted-foreground">
        <input
          aria-readonly={disabled}
          className="size-4 accent-primary"
          type="checkbox"
          checked={checked}
          onClick={(event) => {
            if (disabled) event.preventDefault();
          }}
          onChange={(event) => {
            if (disabled) return;
            onChange(event.target.checked);
          }}
        />
        {text("enable_rounding", "เปิดการปัดเศษ")}
      </label>
    </div>
  );
}

function CompactNumberInput({
  disabled,
  onChange,
  value,
}: {
  disabled?: boolean;
  onChange: (value: number) => void;
  value: number;
}) {
  return (
    <input
      className="h-8 w-24 rounded-lg border border-input bg-background px-2 text-right text-xs text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
      disabled={disabled}
      inputMode="decimal"
      type="number"
      step="0.01"
      value={formatConfigNumber(value)}
      onChange={(event) => onChange(numberFromInput(event.target.value))}
    />
  );
}

function CompactDateInput({
  disabled,
  endOfDay,
  onChange,
  value,
}: {
  disabled?: boolean;
  endOfDay?: boolean;
  onChange: (value: string) => void;
  value: string;
}) {
  return (
    <input
      className="h-8 w-36 rounded-lg border border-input bg-background px-2 text-xs text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
      disabled={disabled}
      type="date"
      value={isoDateToInput(value)}
      onChange={(event) => onChange(dateInputToIso(event.target.value, endOfDay))}
    />
  );
}

function FragmentLikeRows({ children }: { children: ReactNode }) {
  return <>{children}</>;
}

/* ------------------------------------------------------------------ */
/* BranchStructuredSettingEditor                                       */
/* ------------------------------------------------------------------ */

export function BranchStructuredSettingEditor({
  dictionary,
  field,
  form,
  label,
  language,
  readOnly = false,
  setForm,
}: {
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
}) {
  const commit = (nextValue: unknown) => {
    if (readOnly || !setForm) return;
    setForm({ ...form, [field.key]: JSON.stringify(nextValue, null, 2) });
  };

  if (field.key === "paymentrounding") {
    const parsed = normalizePaymentRoundingConfig(form[field.key]);
    return (
      <PaymentRoundingTableEditor
        dictionary={dictionary}
        isRecovered={parsed.isRecovered}
        label={label}
        language={language}
        onChange={commit}
        readOnly={readOnly}
        value={parsed.value}
      />
    );
  }

  const parsed = normalizePointConfig(form[field.key]);
  return (
    <PointConfigTableEditor
      dictionary={dictionary}
      isRecovered={parsed.isRecovered}
      label={label}
      language={language}
      onChange={commit}
      readOnly={readOnly}
      value={parsed.value}
    />
  );
}

/* ------------------------------------------------------------------ */
/* PaymentRoundingTableEditor                                          */
/* ------------------------------------------------------------------ */

export function PaymentRoundingTableEditor({
  dictionary,
  isRecovered,
  label,
  language,
  onChange,
  readOnly = false,
  value,
}: {
  dictionary: BackendLanguageDictionary;
  isRecovered: boolean;
  label: string;
  language: LanguageCode;
  onChange: (value: PaymentRoundingConfig) => void;
  readOnly?: boolean;
  value: PaymentRoundingConfig;
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);

  function updateMethod(
    method: PaymentMethodKey,
    patch: Partial<PaymentMethodRounding>,
  ) {
    if (readOnly) return;
    const current = value[method];
    onChange({
      ...value,
      [method]: {
        ...current,
        ...patch,
        rules: patch.rules ?? current.rules,
      },
    });
  }

  function updateRule(
    method: PaymentMethodKey,
    index: number,
    patch: Partial<RoundingRule>,
  ) {
    if (readOnly) return;
    const current = value[method];
    updateMethod(method, {
      rules: current.rules.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    });
  }

  function addRule(method: PaymentMethodKey) {
    if (readOnly) return;
    updateMethod(method, {
      enabled: true,
      rules: [...value[method].rules, defaultRoundingRule()],
    });
  }

  function removeRule(method: PaymentMethodKey, index: number) {
    if (readOnly) return;
    const current = value[method];
    if (current.rules.length <= 1) return;
    updateMethod(method, {
      rules: current.rules.filter((_, ruleIndex) => ruleIndex !== index),
    });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold">{label}</span>
            {isRecovered ? <Badge variant="warning">{invalidConfigText(language)}</Badge> : null}
          </div>
          <p className="mt-1 text-xs font-medium text-muted-foreground">
            {paymentRoundingSummary(value, language)}
          </p>
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[760px] border-collapse text-left text-xs">
          <thead className="bg-muted/60 text-muted-foreground">
            <tr>
              <th className="w-52 px-2 py-2 font-semibold">{text("payment_method", "ช่องทาง")}</th>
              <th className="px-2 py-2 font-semibold">{text("lower_bound", "ต่ำสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("upper_bound", "สูงสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("round_to", "ปัดเป็น")}</th>
              {!readOnly ? (
                <th className="w-36 px-2 py-2 text-right font-semibold">{text("manage", "จัดการ")}</th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {paymentMethodKeys.map((method) => {
              const methodConfig = value[method];
              if (!methodConfig.enabled) {
                return (
                  <tr className="border-t border-border" key={method}>
                    <td className="px-2 py-2 align-top">
                      <PaymentMethodToggle
                        checked={methodConfig.enabled}
                        label={paymentMethodLabel(method, dictionary)}
                        disabled={readOnly}
                        onChange={(checked) =>
                          updateMethod(method, {
                            enabled: checked,
                            rules: methodConfig.rules.length
                              ? methodConfig.rules
                              : defaultRoundingRules(),
                          })
                        }
                        text={text}
                      />
                    </td>
                    <td className="px-2 py-2 text-muted-foreground" colSpan={3}>
                      {language === "th" ? "ปิดการปัดเศษ" : "Rounding disabled"}
                    </td>
                    {!readOnly ? (
                      <td className="px-2 py-2 text-right">
                        <Button type="button" size="sm" variant="outline" onClick={() => addRule(method)}>
                          <Plus />
                          {text("add_rule", "เพิ่มกฎ")}
                        </Button>
                      </td>
                    ) : null}
                  </tr>
                );
              }
              return (
                <FragmentLikeRows key={method}>
                  {methodConfig.rules.map((rule, index) => (
                    <tr className="border-t border-border" key={`${method}-${index}`}>
                      <td className="px-2 py-2 align-top">
                        {index === 0 ? (
                          <PaymentMethodToggle
                            checked={methodConfig.enabled}
                            label={paymentMethodLabel(method, dictionary)}
                            disabled={readOnly}
                            onChange={(checked) => updateMethod(method, { enabled: checked })}
                            text={text}
                          />
                        ) : null}
                      </td>
                      <td className="px-2 py-1">
                        <CompactNumberInput
                          disabled={readOnly}
                          value={rule.lowerbound}
                          onChange={(nextValue) => updateRule(method, index, { lowerbound: nextValue })}
                        />
                      </td>
                      <td className="px-2 py-1">
                        <CompactNumberInput
                          disabled={readOnly}
                          value={rule.upperbound}
                          onChange={(nextValue) => updateRule(method, index, { upperbound: nextValue })}
                        />
                      </td>
                      <td className="px-2 py-1">
                        <CompactNumberInput
                          disabled={readOnly}
                          value={rule.roundto}
                          onChange={(nextValue) => updateRule(method, index, { roundto: nextValue })}
                        />
                      </td>
                      {!readOnly ? (
                        <td className="px-2 py-1">
                          <div className="flex justify-end gap-1">
                            {index === 0 ? (
                              <Button type="button" size="sm" variant="outline" onClick={() => addRule(method)}>
                                <Plus />
                              </Button>
                            ) : null}
                            <Button
                              type="button"
                              size="sm"
                              variant="outline"
                              disabled={methodConfig.rules.length <= 1}
                              onClick={() => removeRule(method, index)}
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </td>
                      ) : null}
                    </tr>
                  ))}
                </FragmentLikeRows>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

/* ------------------------------------------------------------------ */
/* PointConfigTableEditor                                              */
/* ------------------------------------------------------------------ */

export function PointConfigTableEditor({
  dictionary,
  isRecovered,
  label,
  language,
  onChange,
  readOnly = false,
  value,
}: {
  dictionary: BackendLanguageDictionary;
  isRecovered: boolean;
  label: string;
  language: LanguageCode;
  onChange: (value: PointConfig) => void;
  readOnly?: boolean;
  value: PointConfig;
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);

  function updateGeneralRule(index: number, patch: Partial<GeneralPointRule>) {
    if (readOnly) return;
    onChange({
      ...value,
      generalrules: value.generalrules.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    });
  }

  function updateSpecialRule(index: number, patch: Partial<SpecialPointRule>) {
    if (readOnly) return;
    onChange({
      ...value,
      specialrules: value.specialrules.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    });
  }

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold">{label}</span>
            {isRecovered ? <Badge variant="warning">{invalidConfigText(language)}</Badge> : null}
          </div>
          <p className="mt-1 text-xs font-medium text-muted-foreground">
            {pointConfigSummary(value, language)}
          </p>
        </div>
        <div className="flex flex-wrap gap-1">
          {[1, 2].map((usageType) => (
            <Button
              className={value.pointusagetype === usageType ? "" : "bg-background text-foreground"}
              key={usageType}
              size="sm"
              type="button"
              disabled={readOnly}
              variant={value.pointusagetype === usageType ? "default" : "outline"}
              onClick={() => {
                if (!readOnly) onChange({ ...value, pointusagetype: usageType });
              }}
            >
              {usageType === 1
                ? text("discount", "ส่วนลด")
                : text("cash", "เงินสด")}
            </Button>
          ))}
        </div>
      </div>

      <PointGeneralRulesTable
        dictionary={dictionary}
        language={language}
        onAdd={() =>
          !readOnly
            ? onChange({
                ...value,
                generalrules: [...value.generalrules, defaultGeneralPointRule()],
              })
            : undefined
        }
        onRemove={(index) =>
          !readOnly
            ? onChange({
                ...value,
                generalrules: value.generalrules.filter((_, ruleIndex) => ruleIndex !== index),
              })
            : undefined
        }
        onUpdate={updateGeneralRule}
        readOnly={readOnly}
        rules={value.generalrules}
      />

      <PointSpecialRulesTable
        dictionary={dictionary}
        onAdd={() =>
          !readOnly
            ? onChange({
                ...value,
                specialrules: [...value.specialrules, defaultSpecialPointRule()],
              })
            : undefined
        }
        onRemove={(index) =>
          !readOnly
            ? onChange({
                ...value,
                specialrules: value.specialrules.filter((_, ruleIndex) => ruleIndex !== index),
              })
            : undefined
        }
        onUpdate={updateSpecialRule}
        readOnly={readOnly}
        rules={value.specialrules}
      />
    </section>
  );
}

function PointGeneralRulesTable({
  dictionary,
  language,
  onAdd,
  onRemove,
  onUpdate,
  readOnly = false,
  rules,
}: {
  dictionary: BackendLanguageDictionary;
  language: LanguageCode;
  onAdd: () => void;
  onRemove: (index: number) => void;
  onUpdate: (index: number, patch: Partial<GeneralPointRule>) => void;
  readOnly?: boolean;
  rules: GeneralPointRule[];
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);
  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="font-semibold">{text("general_rules", "กฎทั่วไป")}</h4>
        {!readOnly ? (
          <Button type="button" size="sm" onClick={onAdd}>
            <Plus />
            {text("add_rule", "เพิ่มกฎ")}
          </Button>
        ) : null}
      </div>
      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[720px] border-collapse text-left text-xs">
          <thead className="bg-muted/60 text-muted-foreground">
            <tr>
              <th className="px-2 py-2 font-semibold">#</th>
              <th className="px-2 py-2 font-semibold">{text("startdate", "วันที่เริ่มต้น")}</th>
              <th className="px-2 py-2 font-semibold">{text("enddate", "วันที่สิ้นสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("amount_per_point", "จำนวนเงินต่อ 1 แต้ม")}</th>
              <th className="px-2 py-2 font-semibold">{text("points_per_baht", "แต้มต่อบาท")}</th>
              {!readOnly ? (
                <th className="w-20 px-2 py-2 text-right font-semibold">{text("manage", "จัดการ")}</th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {rules.length === 0 ? (
              <tr className="border-t border-border">
                <td className="px-2 py-3 text-muted-foreground" colSpan={readOnly ? 5 : 6}>
                  {language === "th" ? "ยังไม่มีกฎทั่วไป" : "No general rules"}
                </td>
              </tr>
            ) : (
              rules.map((rule, index) => (
                <tr className="border-t border-border" key={index}>
                  <td className="px-2 py-1 font-semibold">{index + 1}</td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      value={rule.startdate}
                      onChange={(nextValue) => onUpdate(index, { startdate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      endOfDay
                      value={rule.enddate}
                      onChange={(nextValue) => onUpdate(index, { enddate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.payperpoint}
                      onChange={(nextValue) => onUpdate(index, { payperpoint: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.pointvalue}
                      onChange={(nextValue) => onUpdate(index, { pointvalue: nextValue })}
                    />
                  </td>
                  {!readOnly ? (
                    <td className="px-2 py-1 text-right">
                      <Button type="button" size="sm" variant="outline" onClick={() => onRemove(index)}>
                        <Trash2 />
                      </Button>
                    </td>
                  ) : null}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function PointSpecialRulesTable({
  dictionary,
  onAdd,
  onRemove,
  onUpdate,
  readOnly = false,
  rules,
}: {
  dictionary: BackendLanguageDictionary;
  onAdd: () => void;
  onRemove: (index: number) => void;
  onUpdate: (index: number, patch: Partial<SpecialPointRule>) => void;
  readOnly?: boolean;
  rules: SpecialPointRule[];
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);
  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="font-semibold">{text("special_rules", "กฎพิเศษ")}</h4>
        {!readOnly ? (
          <Button type="button" size="sm" onClick={onAdd}>
            <Plus />
            {text("add_rule", "เพิ่มกฎ")}
          </Button>
        ) : null}
      </div>
      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[980px] border-collapse text-left text-xs">
          <thead className="bg-muted/60 text-muted-foreground">
            <tr>
              <th className="px-2 py-2 font-semibold">#</th>
              <th className="px-2 py-2 font-semibold">{text("startdate", "วันที่เริ่มต้น")}</th>
              <th className="px-2 py-2 font-semibold">{text("enddate", "วันที่สิ้นสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("point_multiplier", "ตัวคูณแต้ม")}</th>
              <th className="px-2 py-2 font-semibold">{text("max_points_per_bill", "แต้มสูงสุดต่อบิล")}</th>
              <th className="px-2 py-2 font-semibold">{text("days_of_week", "วันในสัปดาห์")}</th>
              {!readOnly ? (
                <th className="w-20 px-2 py-2 text-right font-semibold">{text("manage", "จัดการ")}</th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {rules.length === 0 ? (
              <tr className="border-t border-border">
                <td className="px-2 py-3 text-muted-foreground" colSpan={readOnly ? 6 : 7}>
                  {text("no_special_rules", "ไม่มีกฎพิเศษ")}
                </td>
              </tr>
            ) : (
              rules.map((rule, index) => (
                <tr className="border-t border-border" key={index}>
                  <td className="px-2 py-1 font-semibold">{index + 1}</td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      value={rule.startdate}
                      onChange={(nextValue) => onUpdate(index, { startdate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      endOfDay
                      value={rule.enddate}
                      onChange={(nextValue) => onUpdate(index, { enddate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.multiplier}
                      onChange={(nextValue) => onUpdate(index, { multiplier: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.maxpointperbill}
                      onChange={(nextValue) => onUpdate(index, { maxpointperbill: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <div className="flex min-w-[340px] flex-wrap gap-1">
                      {pointWeekDays.map((day) => (
                        <label className="inline-flex w-auto max-w-none items-center gap-1 rounded-lg border border-border bg-muted/20 px-2 py-1" key={day}>
                          <input
                            aria-readonly={readOnly}
                            className="size-3.5 accent-primary"
                            type="checkbox"
                            checked={rule[day]}
                            onClick={(event) => {
                              if (readOnly) event.preventDefault();
                            }}
                            onChange={(event) => {
                              if (readOnly) return;
                              onUpdate(index, { [day]: event.target.checked } as Partial<SpecialPointRule>);
                            }}
                          />
                          <span>{shortDayLabel(day, dictionary)}</span>
                        </label>
                      ))}
                    </div>
                  </td>
                  {!readOnly ? (
                    <td className="px-2 py-1 text-right">
                      <Button type="button" size="sm" variant="outline" onClick={() => onRemove(index)}>
                        <Trash2 />
                      </Button>
                    </td>
                  ) : null}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

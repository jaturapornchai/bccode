"use client";

/**
 * ProductBarcodeFormDialog — structured editor for a Product Barcode record.
 *
 * Replaces the old raw-JSON textarea with a 12-tab form covering every field
 * in backend Go `ProductBarcodeBase` + `ProductBarcode`. Uses MasterPicker for
 * code/name lookups against the live mainapi master endpoints.
 *
 * - Pure controlled component: parent owns the `ProductBarcode` value via
 *   `value` + `onChange` props.
 * - `onSave` / `onCancel` for outer Save / Cancel actions.
 * - Validates: barcode regex, names[0] required, ItemUnit required.
 */

import {
  ChevronLeft,
  ChevronRight,
  ImagePlus,
  Loader2,
  Minus,
  Plus,
  Save,
  Trash2,
  Upload,
  X,
} from "lucide-react";
import {
  type ChangeEvent,
  type Dispatch,
  type FormEvent,
  type KeyboardEvent,
  type ReactNode,
  type SetStateAction,
  useCallback,
  useMemo,
  useRef,
  useState,
} from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { uploadProductImage, type MasterName, type MasterEntry } from "@/lib/product-barcode/api";
import { getBarcodeText } from "@/lib/product-barcode/language";
import {
  FOOD_TYPE,
  ITEM_TYPE,
  MATERIAL_TYPE,
  type NameX,
  type ProductBarcode,
  type ProductPrice,
  type RefProductBarcode,
  type BOMProductBarcode,
  type ProductChoice,
  type ProductOption,
  type ProductTimeForSale,
  type ProductDimension,
  type ProductOrderType,
  type ProductBarcodeBusinessType,
  type ProductBarcodeBranch,
} from "@/lib/product-barcode/types";
import { isValidBarcode, pickName, setNameXEntry } from "@/lib/product-barcode/utils";
import type { LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import type { AuthSession } from "@/lib/workspace-models";

// ─── Tab definitions ──────────────────────────────────────────────────────

type TabKey =
  | "basic"
  | "classification"
  | "pricing"
  | "stock"
  | "units"
  | "bom"
  | "media"
  | "restaurant"
  | "timeforsales"
  | "dimensions"
  | "business"
  | "misc";

interface TabDef {
  key: TabKey;
  label: keyof ReturnType<typeof getBarcodeText>;
}

const TABS: TabDef[] = [
  { key: "basic", label: "tabBasic" },
  { key: "classification", label: "tabClassification" },
  { key: "pricing", label: "tabPricing" },
  { key: "stock", label: "tabStock" },
  { key: "units", label: "tabUnits" },
  { key: "bom", label: "tabBom" },
  { key: "media", label: "tabMedia" },
  { key: "restaurant", label: "tabRestaurant" },
  { key: "timeforsales", label: "tabTimeForSales" },
  { key: "dimensions", label: "tabDimensions" },
  { key: "business", label: "tabBusinessBranch" },
  { key: "misc", label: "tabMisc" },
];

// ─── Props ────────────────────────────────────────────────────────────────

export interface ProductBarcodeFormDialogProps {
  open: boolean;
  mode: "create" | "edit";
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  onSave: (value: ProductBarcode) => void;
  onCancel: () => void;
  saving?: boolean;
  language: LanguageCode | string;
  auth: AuthSession | null;
  shopLanguages?: string[];
  /** Optional extra controls (e.g. quick links to BOM / Price History / Label print). */
  extraActions?: ReactNode;
  embedded?: boolean;
}

export function ProductBarcodeFormDialog(props: ProductBarcodeFormDialogProps) {
  const { open, mode, value, onChange, onSave, onCancel, saving = false, language, auth, extraActions, embedded = false } = props;
  const text = getBarcodeText(language);
  const [tab, setTab] = useState<TabKey>("basic");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const shopLanguages = useMemo(
    () => (props.shopLanguages && props.shopLanguages.length > 0 ? props.shopLanguages : ["th", "en"]),
    [props.shopLanguages],
  );

  const validate = useCallback((): boolean => {
    const next: Record<string, string> = {};
    if (!value.barcode.trim()) next.barcode = text.required_error;
    else if (!isValidBarcode(value.barcode)) next.barcode = text.invalidBarcode;
    const firstName = value.names[0]?.name ?? "";
    if (!firstName.trim()) next.name0 = text.required_error;
    if (!value.item_unit_code.trim()) next.item_unit_code = text.required_error;
    setErrors(next);
    return Object.keys(next).length === 0;
  }, [value, text.required_error, text.invalidBarcode]);

  const saveIfValid = useCallback(() => {
      if (validate()) onSave(value);
  }, [validate, onSave, value]);

  const submit = useCallback(
    (event: FormEvent) => {
      event.preventDefault();
      saveIfValid();
    },
    [saveIfValid],
  );

  const handleFormKeyDown = useCallback(
    (event: KeyboardEvent<HTMLFormElement>) => {
      if (event.key !== "F10") return;
      event.preventDefault();
      if (!saving) saveIfValid();
    },
    [saveIfValid, saving],
  );

  if (!open) return null;

  const form = (
      <form
        onSubmit={submit}
        onKeyDown={handleFormKeyDown}
        className={cn(
          "flex w-full flex-col overflow-hidden border border-border bg-card text-card-foreground",
          embedded
            ? "rounded-lg shadow-sm xl:h-full xl:min-h-0"
            : "h-full max-w-6xl shadow-2xl md:h-[90vh] md:rounded-lg",
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between gap-2 border-b border-border bg-card/95 px-4 py-3">
          <div className="flex min-w-0 flex-col">
            <span className="text-sm text-muted-foreground">{text.detailTitle}</span>
            <span className="truncate text-base font-semibold">
              {mode === "create" ? text.createTitle : text.editTitle}
              {value.barcode ? ` — ${value.barcode}` : ""}
            </span>
          </div>
          <div className="flex items-center gap-2">
            {extraActions}
            <Button type="button" variant="ghost" size="sm" onClick={onCancel} aria-label={text.close}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Tab bar (scrollable on mobile) */}
        <div className="border-b border-border bg-muted/30">
          <div className="flex flex-wrap gap-1 px-2 py-1.5">
            {TABS.map((tabDef) => {
              const isActive = tab === tabDef.key;
              return (
                <button
                  key={tabDef.key}
                  type="button"
                  onClick={() => setTab(tabDef.key)}
                  className={cn(
                    "rounded-md px-3 py-1.5 text-xs font-medium transition",
                    isActive
                      ? "bg-primary text-primary-foreground shadow-sm"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground",
                  )}
                >
                  {text[tabDef.label]}
                </button>
              );
            })}
          </div>
        </div>

        {/* Body */}
        <div className="min-h-0 flex-1 overflow-y-auto bg-background px-4 py-4">
          {tab === "basic" && (
            <TabBasic value={value} onChange={onChange} text={text} errors={errors} shopLanguages={shopLanguages} auth={auth} language={language} />
          )}
          {tab === "classification" && (
            <TabClassification value={value} onChange={onChange} text={text} auth={auth} language={language} />
          )}
          {tab === "pricing" && <TabPricing value={value} onChange={onChange} text={text} />}
          {tab === "stock" && <TabStock value={value} onChange={onChange} text={text} />}
          {tab === "units" && <TabUnits value={value} onChange={onChange} text={text} />}
          {tab === "bom" && <TabBom value={value} onChange={onChange} text={text} />}
          {tab === "media" && (
            <TabMedia value={value} onChange={onChange} text={text} auth={auth} />
          )}
          {tab === "restaurant" && (
            <TabRestaurant value={value} onChange={onChange} text={text} auth={auth} language={language} shopLanguages={shopLanguages} />
          )}
          {tab === "timeforsales" && <TabTimeForSales value={value} onChange={onChange} text={text} />}
          {tab === "dimensions" && <TabDimensions value={value} onChange={onChange} text={text} />}
          {tab === "business" && (
            <TabBusinessBranch value={value} onChange={onChange} text={text} auth={auth} language={language} />
          )}
          {tab === "misc" && <TabMisc value={value} onChange={onChange} text={text} />}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between gap-2 border-t border-border bg-card/95 px-4 py-3">
          <div className="flex gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                const idx = TABS.findIndex((t) => t.key === tab);
                if (idx > 0) setTab(TABS[idx - 1].key);
              }}
              disabled={TABS.findIndex((t) => t.key === tab) === 0}
              aria-label="Previous tab"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                const idx = TABS.findIndex((t) => t.key === tab);
                if (idx < TABS.length - 1) setTab(TABS[idx + 1].key);
              }}
              disabled={TABS.findIndex((t) => t.key === tab) === TABS.length - 1}
              aria-label="Next tab"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            {Object.keys(errors).length > 0 ? (
              <Badge variant="warning" className="ml-2">
                {Object.keys(errors).length}
              </Badge>
            ) : null}
          </div>
          <div className="flex gap-2">
            <Button type="button" variant="outline" size="sm" onClick={onCancel} disabled={saving}>
              {text.cancel}
            </Button>
            <Button type="submit" size="sm" disabled={saving}>
              {saving ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : <Save className="mr-1 h-4 w-4" />}
              {saving ? text.saving : text.save}
            </Button>
          </div>
        </div>
      </form>
  );

  if (embedded) return form;

  return (
    <div className="fixed inset-0 z-40 flex items-stretch justify-center bg-black/40 p-0 md:items-center md:p-4">
      {form}
    </div>
  );
}

// ─── Shared field helpers ─────────────────────────────────────────────────

type Text = ReturnType<typeof getBarcodeText>;

function FieldRow({ label, hint, children, required }: { label: string; hint?: string; children: ReactNode; required?: boolean }) {
  return (
    <label className="flex flex-col gap-1 text-sm">
      <span className="font-medium text-foreground">
        {label}
        {required ? <span className="ml-1 text-destructive">*</span> : null}
      </span>
      {children}
      {hint ? <span className="text-xs text-muted-foreground">{hint}</span> : null}
    </label>
  );
}

function FieldGrid({ children }: { children: ReactNode }) {
  return <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{children}</div>;
}

function Section({ title, action, children }: { title: string; action?: ReactNode; children: ReactNode }) {
  return (
    <section className="mb-4 overflow-hidden rounded-lg border border-border bg-card">
      <header className="flex items-center justify-between border-b border-border px-3 py-2">
        <h3 className="text-sm font-semibold">{title}</h3>
        {action}
      </header>
      <div className="p-3">{children}</div>
    </section>
  );
}

function Toggle({
  checked,
  onCheckedChange,
  label,
  disabled,
}: {
  checked: boolean;
  onCheckedChange: (next: boolean) => void;
  label: string;
  disabled?: boolean;
}) {
  return (
    <label className={cn("flex cursor-pointer items-center gap-2 text-sm", disabled && "cursor-not-allowed opacity-60")}>
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(event) => onCheckedChange(event.target.checked)}
        className="size-4 rounded border-input"
      />
      <span>{label}</span>
    </label>
  );
}

type RadioOptionValue = string | number | boolean;

function RadioOptionGroup<T extends RadioOptionValue>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: Array<{ value: T; label: string; disabled?: boolean }>;
  onChange: (next: T) => void;
}) {
  return (
    <fieldset className="rounded-md border border-border bg-background px-3 py-2">
      <legend className="px-1 text-xs font-semibold text-muted-foreground">{label}</legend>
      <div className="flex flex-wrap gap-x-5 gap-y-2">
        {options.map((option) => (
          <label
            key={String(option.value)}
            className={cn(
              "flex min-h-8 cursor-pointer items-center gap-2 text-sm",
              option.disabled && "cursor-not-allowed opacity-60",
            )}
          >
            <input
              type="radio"
              checked={Object.is(value, option.value)}
              disabled={option.disabled}
              onChange={() => onChange(option.value)}
              className="size-4"
            />
            <span>{option.label}</span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}

function NumberField({
  value,
  onChange,
  min,
  step = "any",
  className,
}: {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  step?: string | number;
  className?: string;
}) {
  return (
    <Input
      type="number"
      value={Number.isFinite(value) ? value : 0}
      step={step}
      min={min}
      onChange={(event: ChangeEvent<HTMLInputElement>) => {
        const n = Number(event.target.value);
        onChange(Number.isFinite(n) ? n : 0);
      }}
      className={className}
    />
  );
}

function emptyReferenceBarcode(): RefProductBarcode {
  return {
    guid_fixed: "",
    names: [],
    item_unit_code: "",
    itemunitnames: [],
    barcode: "",
    condition: false,
    dividevalue: 1,
    standvalue: 1,
    qty: 1,
  };
}

function canUseReferenceBarcode(itemType: ProductBarcode["item_type"]): boolean {
  return itemType !== ITEM_TYPE.SERVICE && itemType !== ITEM_TYPE.NOT_STOCK;
}

function NamesEditor({
  names,
  onChange,
  languages,
  label,
  firstRequired,
  error,
  progressiveLanguages = false,
  addLanguageLabel = "Add language",
  otherLanguageLabel = "Other language",
}: {
  names: NameX[];
  onChange: (next: NameX[]) => void;
  languages: string[];
  label: string;
  firstRequired?: boolean;
  error?: string;
  progressiveLanguages?: boolean;
  addLanguageLabel?: string;
  otherLanguageLabel?: string;
}) {
  const allLanguages = useMemo(() => {
    const codes = new Set<string>();
    const primary = languages.find((code) => code.trim())?.trim() || "th";
    codes.add(primary);
    languages.forEach((code) => {
      if (code.trim()) codes.add(code.trim());
    });
    names.forEach((entry) => {
      const code = entry.code?.trim();
      if (code) codes.add(code);
    });
    return Array.from(codes);
  }, [languages, names]);
  const primaryLanguage = languages.find((code) => code.trim())?.trim() || allLanguages[0] || "th";
  const [extraLanguages, setExtraLanguages] = useState<string[]>([]);
  const visibleLanguages = progressiveLanguages
    ? allLanguages.filter(
        (code) =>
          code === primaryLanguage ||
          extraLanguages.includes(code) ||
          names.some((entry) => entry.code === code && (entry.name ?? "").trim()),
      )
    : allLanguages;
  const hiddenLanguages = progressiveLanguages
    ? allLanguages.filter((code) => code !== primaryLanguage && !visibleLanguages.includes(code))
    : [];
  const [languageToAdd, setLanguageToAdd] = useState("");

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">
        {label}
        {firstRequired ? <span className="ml-1 text-destructive">*</span> : null}
      </div>
      <div className="grid gap-2 sm:grid-cols-2">
        {visibleLanguages.map((code) => {
          const entry = names.find((n) => n.code === code);
          return (
            <div key={code} className="flex items-center gap-2">
              <span className="w-10 text-center text-xs font-mono uppercase text-muted-foreground">{code}</span>
              <Input
                value={entry?.name ?? ""}
                onChange={(event) => onChange(setNameXEntry(names, code, event.target.value))}
                placeholder={code}
                aria-invalid={firstRequired && code === primaryLanguage && !entry?.name ? true : undefined}
              />
            </div>
          );
        })}
      </div>
      {progressiveLanguages && hiddenLanguages.length > 0 ? (
        <div className="flex flex-wrap items-center gap-2">
          <select
            className="h-9 rounded-md border border-input bg-background px-2 text-sm"
            value={languageToAdd}
            onChange={(event) => setLanguageToAdd(event.target.value)}
            aria-label={otherLanguageLabel}
          >
            <option value="">{otherLanguageLabel}</option>
            {hiddenLanguages.map((code) => (
              <option key={code} value={code}>
                {code.toUpperCase()}
              </option>
            ))}
          </select>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              const nextCode = languageToAdd || hiddenLanguages[0];
              if (!nextCode) return;
              setExtraLanguages((current) => (current.includes(nextCode) ? current : [...current, nextCode]));
              setLanguageToAdd("");
            }}
          >
            <Plus className="mr-1 h-4 w-4" />
            {addLanguageLabel}
          </Button>
        </div>
      ) : null}
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
    </div>
  );
}

// ─── Master picker field ──────────────────────────────────────────────────

function MasterField({
  label,
  code,
  names,
  master,
  language,
  auth,
  onPick,
  onClear,
  required,
  error,
}: {
  label: string;
  code: string;
  names: NameX[];
  master: MasterName;
  language: LanguageCode | string;
  auth: AuthSession | null;
  onPick: (entry: MasterEntry) => void;
  onClear: () => void;
  required?: boolean;
  error?: string;
}) {
  const [open, setOpen] = useState(false);
  const pickerAnchorRef = useRef<HTMLButtonElement | null>(null);
  const displayName = pickName(names, language);
  return (
    <FieldRow label={label} required={required}>
      <div className="relative flex gap-2">
        <button
          ref={pickerAnchorRef}
          type="button"
          onClick={() => setOpen(true)}
          className="flex h-10 min-w-0 flex-1 items-center justify-between rounded-lg border border-input bg-background px-3 text-left text-sm text-foreground hover:bg-muted/40"
        >
          <span className="truncate">{displayName || code || <span className="text-muted-foreground">—</span>}</span>
          {code ? <span className="ml-2 truncate text-xs text-muted-foreground">{code}</span> : null}
        </button>
        {code ? (
          <Button type="button" variant="ghost" size="icon" onClick={onClear} aria-label="Clear">
            <X className="h-4 w-4" />
          </Button>
        ) : null}
        <MasterPicker
          open={open}
          onClose={() => setOpen(false)}
          auth={auth}
          language={language}
          master={master}
          title={label}
          onSelect={onPick}
          placement="field"
          anchorRef={pickerAnchorRef}
        />
      </div>
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
    </FieldRow>
  );
}

// ─── Tabs ────────────────────────────────────────────────────────────────

function TabBasic({
  auth,
  language,
  value,
  onChange,
  text,
  errors,
  shopLanguages,
}: {
  auth: AuthSession | null;
  language: LanguageCode | string;
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  errors: Record<string, string>;
  shopLanguages: string[];
}) {
  const upd = useCallback(
    <K extends keyof ProductBarcode>(key: K, val: ProductBarcode[K]) =>
      onChange((current) => ({ ...current, [key]: val })),
    [onChange],
  );
  const setItemType = useCallback(
    (nextType: ProductBarcode["item_type"]) =>
      onChange((current) => {
        const next: ProductBarcode = { ...current, item_type: nextType };
        if (!canUseReferenceBarcode(nextType)) {
          next.isusesubbarcodes = false;
          next.refbarcodes = [];
        } else if (nextType === ITEM_TYPE.STOCK && next.refbarcodes.length > 1) {
          next.refbarcodes = next.refbarcodes.slice(0, 1);
        }
        return next;
      }),
    [onChange],
  );
  const setReferenceBarcodeEnabled = useCallback(
    (enabled: boolean) =>
      onChange((current) => {
        if (!enabled || !canUseReferenceBarcode(current.item_type)) {
          return { ...current, isusesubbarcodes: false, refbarcodes: [] };
        }
        return {
          ...current,
          isusesubbarcodes: true,
          refbarcodes: current.refbarcodes.length > 0 ? current.refbarcodes : [emptyReferenceBarcode()],
        };
      }),
    [onChange],
  );
  const referenceDisabled = !canUseReferenceBarcode(value.item_type);

  return (
    <div className="space-y-3">
      <Section title={text.tabBasic}>
        <div className="space-y-3">
          <FieldGrid>
            <FieldRow label={text.barcode} hint={text.barcodeHelp} required>
              <Input
                value={value.barcode}
                onChange={(event) =>
                  upd(
                    "barcode",
                    event.target.value
                      .toUpperCase()
                      .replace(/[^A-Z0-9-]/g, ""),
                  )
                }
                aria-invalid={Boolean(errors.barcode) || undefined}
              />
              {errors.barcode ? <p className="text-xs text-destructive">{errors.barcode}</p> : null}
            </FieldRow>
            <FieldRow label={text.itemCode}>
              <Input
                value={value.itemcode}
                onChange={(event) => upd("itemcode", event.target.value.toUpperCase())}
              />
            </FieldRow>
            <MasterField
              label={text.itemUnit}
              code={value.item_unit_code}
              names={value.itemunitnames}
              master="unit"
              language={language}
              auth={auth}
              required
              error={errors.item_unit_code}
              onPick={(entry) =>
                onChange((current) => ({
                  ...current,
                  itemunitguid: entry.guidfixed,
                  item_unit_code: entry.code,
                  itemunitnames: entry.names,
                }))
              }
              onClear={() =>
                onChange((current) => ({
                  ...current,
                  itemunitguid: "",
                  item_unit_code: "",
                  itemunitnames: [],
                }))
              }
            />
          </FieldGrid>
          <NamesEditor
            names={value.names}
            onChange={(next) => upd("names", next)}
            languages={shopLanguages}
            label={text.names}
            firstRequired
            error={errors.name0}
            progressiveLanguages
            addLanguageLabel={text.addLanguage}
            otherLanguageLabel={text.otherLanguage}
          />
        </div>
      </Section>

      <Section title={text.itemType}>
        <div className="space-y-3">
          <RadioOptionGroup
            label={text.itemType}
            value={value.item_type}
            onChange={(next) => setItemType(next)}
            options={[
              { value: ITEM_TYPE.STOCK, label: text.itemTypeStock },
              { value: ITEM_TYPE.SERVICE, label: text.itemTypeService },
              { value: ITEM_TYPE.SET, label: text.itemTypeSet },
              { value: ITEM_TYPE.NOT_STOCK, label: text.itemTypeNotStock },
            ]}
          />
          <div className="rounded-md border border-border bg-background px-3 py-2">
            <Toggle
              checked={value.isusesubbarcodes && !referenceDisabled}
              disabled={referenceDisabled}
              onCheckedChange={setReferenceBarcodeEnabled}
              label={`${text.isUseSubBarcodes}${value.refbarcodes.length ? ` (${value.refbarcodes.length})` : ""}`}
            />
          </div>
          <RadioOptionGroup
            label={text.materialType}
            value={value.materialtype}
            onChange={(next) => upd("materialtype", next)}
            options={[
              { value: MATERIAL_TYPE.GENERAL, label: text.materialGeneral },
              { value: MATERIAL_TYPE.MATERIAL, label: text.materialMaterial },
              { value: MATERIAL_TYPE.SEMI_FINISHED, label: text.materialSemiFinished },
            ]}
          />
        </div>
      </Section>

      <Section title={text.vatType}>
        <div className="grid gap-3 lg:grid-cols-2">
          <RadioOptionGroup
            label={text.vatType}
            value={value.vatcal}
            onChange={(next) => upd("vatcal", next)}
            options={[
              { value: 0, label: text.vatIncluded },
              { value: 1, label: text.vatExcluded },
            ]}
          />
          <RadioOptionGroup
            label={text.isSumPoint}
            value={value.issumpoint}
            onChange={(next) => upd("issumpoint", next)}
            options={[
              { value: true, label: text.isSumPointYes },
              { value: false, label: text.isSumPointNo },
            ]}
          />
        </div>
      </Section>

      <Section title={text.tabBasic + " — flags"}>
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          <Toggle
            checked={value.is_main_barcode}
            onCheckedChange={(n) => upd("is_main_barcode", n)}
            label={text.isMainBarcode}
          />
          <Toggle checked={value.isdividend} onCheckedChange={(n) => upd("isdividend", n)} label={text.isDividend} />
          <Toggle
            checked={value.isdiscountpointofpurchase}
            onCheckedChange={(n) => upd("isdiscountpointofpurchase", n)}
            label={text.isDiscountPointOfPurchase}
          />
        </div>
      </Section>
    </div>
  );
}

function TabClassification({
  value,
  onChange,
  text,
  auth,
  language,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  auth: AuthSession | null;
  language: LanguageCode | string;
}) {
  const setMaster = useCallback(
    (
      keys: { guid: keyof ProductBarcode; code: keyof ProductBarcode; names: keyof ProductBarcode },
      entry: MasterEntry,
    ) =>
      onChange((current) => ({
        ...current,
        [keys.guid]: entry.guidfixed,
        [keys.code]: entry.code,
        [keys.names]: entry.names,
      })),
    [onChange],
  );
  const clearMaster = useCallback(
    (keys: { guid: keyof ProductBarcode; code: keyof ProductBarcode; names: keyof ProductBarcode }) =>
      onChange((current) => ({ ...current, [keys.guid]: "", [keys.code]: "", [keys.names]: [] })),
    [onChange],
  );
  return (
    <Section title={text.tabClassification}>
      <FieldGrid>
        <MasterField
          label={text.group}
          code={value.group_code}
          names={value.group_names}
          master="group"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "groupguid", code: "group_code", names: "group_names" }, entry)}
          onClear={() => clearMaster({ guid: "groupguid", code: "group_code", names: "group_names" })}
        />
        <MasterField
          label={text.groupSubOne}
          code={value.groupsubonecode}
          names={value.groupsubonenames}
          master="groupsubone"
          language={language}
          auth={auth}
          onPick={(entry) =>
            setMaster({ guid: "groupsuboneguid", code: "groupsubonecode", names: "groupsubonenames" }, entry)
          }
          onClear={() =>
            clearMaster({ guid: "groupsuboneguid", code: "groupsubonecode", names: "groupsubonenames" })
          }
        />
        <MasterField
          label={text.groupSubTwo}
          code={value.groupsubtwocode}
          names={value.groupsubtwonames}
          master="groupsubtwo"
          language={language}
          auth={auth}
          onPick={(entry) =>
            setMaster({ guid: "groupsubtwoguid", code: "groupsubtwocode", names: "groupsubtwonames" }, entry)
          }
          onClear={() =>
            clearMaster({ guid: "groupsubtwoguid", code: "groupsubtwocode", names: "groupsubtwonames" })
          }
        />
        <MasterField
          label={text.brand}
          code={value.brand_code}
          names={value.brandnames}
          master="brand"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "brandguid", code: "brand_code", names: "brandnames" }, entry)}
          onClear={() => clearMaster({ guid: "brandguid", code: "brand_code", names: "brandnames" })}
        />
        <MasterField
          label={text.category}
          code={value.categorycode}
          names={value.category_names}
          master="category"
          language={language}
          auth={auth}
          onPick={(entry) =>
            setMaster({ guid: "category_guid", code: "categorycode", names: "category_names" }, entry)
          }
          onClear={() => clearMaster({ guid: "category_guid", code: "categorycode", names: "category_names" })}
        />
        <MasterField
          label={text.classification}
          code={value.classcode}
          names={value.classnames}
          master="class"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "classguid", code: "classcode", names: "classnames" }, entry)}
          onClear={() => clearMaster({ guid: "classguid", code: "classcode", names: "classnames" })}
        />
        <MasterField
          label={text.design}
          code={value.designcode}
          names={value.designnames}
          master="design"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "designguid", code: "designcode", names: "designnames" }, entry)}
          onClear={() => clearMaster({ guid: "designguid", code: "designcode", names: "designnames" })}
        />
        <MasterField
          label={text.grade}
          code={value.gradecode}
          names={value.gradenames}
          master="grade"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "gradeguid", code: "gradecode", names: "gradenames" }, entry)}
          onClear={() => clearMaster({ guid: "gradeguid", code: "gradecode", names: "gradenames" })}
        />
        <MasterField
          label={text.model}
          code={value.modelcode}
          names={value.modelnames}
          master="model"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "modelguid", code: "modelcode", names: "modelnames" }, entry)}
          onClear={() => clearMaster({ guid: "modelguid", code: "modelcode", names: "modelnames" })}
        />
        <MasterField
          label={text.pattern}
          code={value.patterncode}
          names={value.patternnames}
          master="pattern"
          language={language}
          auth={auth}
          onPick={(entry) => setMaster({ guid: "patternguid", code: "patterncode", names: "patternnames" }, entry)}
          onClear={() => clearMaster({ guid: "patternguid", code: "patterncode", names: "patternnames" })}
        />
        <MasterField
          label={text.manufacturer}
          code={value.manufacturercode}
          names={value.manufacturernames}
          master="producttype"
          language={language}
          auth={auth}
          onPick={(entry) =>
            setMaster(
              { guid: "manufacturerguid", code: "manufacturercode", names: "manufacturernames" },
              entry,
            )
          }
          onClear={() =>
            clearMaster({ guid: "manufacturerguid", code: "manufacturercode", names: "manufacturernames" })
          }
        />
        <FieldRow label={text.productType}>
          <NumberField value={value.producttype} onChange={(n) => onChange((c) => ({ ...c, producttype: n }))} step={1} />
        </FieldRow>
      </FieldGrid>
    </Section>
  );
}

function TabPricing({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const setPrices = useCallback(
    (mutator: (prices: ProductPrice[]) => ProductPrice[]) =>
      onChange((current) => ({ ...current, prices: mutator(current.prices) })),
    [onChange],
  );
  const setFixed = useCallback(
    (mutator: (rows: ProductBarcode["fixedcost"]) => ProductBarcode["fixedcost"]) =>
      onChange((current) => ({ ...current, fixedcost: mutator(current.fixedcost) })),
    [onChange],
  );

  return (
    <div className="space-y-4">
      <Section
        title={text.prices}
        action={
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() =>
              setPrices((rows) => [...rows, { key_number: rows.length + 1, price: 0 }])
            }
          >
            <Plus className="mr-1 h-4 w-4" />
            {text.addPrice}
          </Button>
        }
      >
        <div className="space-y-2">
          {value.prices.length === 0 ? (
            <p className="text-sm text-muted-foreground">—</p>
          ) : (
            value.prices.map((entry, idx) => (
              <div key={idx} className="grid grid-cols-[80px_1fr_40px] items-center gap-2">
                <Input
                  type="number"
                  value={entry.key_number}
                  min={1}
                  step={1}
                  onChange={(event) =>
                    setPrices((rows) =>
                      rows.map((row, rowIdx) =>
                        rowIdx === idx ? { ...row, key_number: Number(event.target.value) || 0 } : row,
                      ),
                    )
                  }
                  aria-label={text.priceKey}
                />
                <Input
                  type="number"
                  value={entry.price}
                  step="any"
                  onChange={(event) =>
                    setPrices((rows) =>
                      rows.map((row, rowIdx) =>
                        rowIdx === idx ? { ...row, price: Number(event.target.value) || 0 } : row,
                      ),
                    )
                  }
                  aria-label={text.price}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => setPrices((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))
          )}
        </div>
      </Section>

      <Section
        title={text.fixedCost}
        action={
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() =>
              setFixed((rows) => [...rows, { effectdate: new Date().toISOString().slice(0, 10), amount: 0 }])
            }
          >
            <Plus className="mr-1 h-4 w-4" />
            {text.addFixedCost}
          </Button>
        }
      >
        <div className="space-y-2">
          {value.fixedcost.length === 0 ? (
            <p className="text-sm text-muted-foreground">—</p>
          ) : (
            value.fixedcost.map((entry, idx) => (
              <div key={idx} className="grid grid-cols-[1fr_1fr_40px] items-center gap-2">
                <Input
                  type="date"
                  value={entry.effectdate.slice(0, 10)}
                  onChange={(event) =>
                    setFixed((rows) =>
                      rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, effectdate: event.target.value } : row)),
                    )
                  }
                  aria-label={text.fixedCostDate}
                />
                <Input
                  type="number"
                  value={entry.amount}
                  step="any"
                  onChange={(event) =>
                    setFixed((rows) =>
                      rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, amount: Number(event.target.value) || 0 } : row)),
                    )
                  }
                  aria-label={text.fixedCostAmount}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => setFixed((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))
          )}
        </div>
      </Section>

      <Section title={text.discount}>
        <FieldGrid>
          <FieldRow label={text.discount}>
            <Input value={value.discount} onChange={(event) => onChange((c) => ({ ...c, discount: event.target.value }))} />
          </FieldRow>
          <FieldRow label={text.maxDiscount}>
            <Input value={value.maxdiscount} onChange={(event) => onChange((c) => ({ ...c, maxdiscount: event.target.value }))} />
          </FieldRow>
        </FieldGrid>
      </Section>
    </div>
  );
}

function TabStock({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const upd = useCallback(
    <K extends keyof ProductBarcode>(key: K, val: ProductBarcode[K]) =>
      onChange((c) => ({ ...c, [key]: val })),
    [onChange],
  );
  return (
    <Section title={text.tabStock}>
      <FieldGrid>
        <FieldRow label={text.orderPoint}>
          <NumberField value={value.orderpoint} onChange={(n) => upd("orderpoint", n)} min={0} />
        </FieldRow>
        <FieldRow label={text.minPoint}>
          <NumberField value={value.minpoint} onChange={(n) => upd("minpoint", n)} min={0} />
        </FieldRow>
        <FieldRow label={text.maxPoint}>
          <NumberField value={value.maxpoint} onChange={(n) => upd("maxpoint", n)} min={0} />
        </FieldRow>
        <FieldRow label={text.qty}>
          <NumberField value={value.qty} onChange={(n) => upd("qty", n)} />
        </FieldRow>
        <FieldRow label={text.stockBarcode}>
          <Input value={value.stockbarcode} onChange={(event) => upd("stockbarcode", event.target.value)} />
        </FieldRow>
      </FieldGrid>
    </Section>
  );
}

function TabUnits({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const setRef = useCallback(
    (mutator: (rows: RefProductBarcode[]) => RefProductBarcode[]) =>
      onChange((c) => ({ ...c, refbarcodes: mutator(c.refbarcodes) })),
    [onChange],
  );
  return (
    <div className="space-y-4">
      <Section title={text.tabUnits}>
        <FieldGrid>
          <FieldRow label={text.divideValue}>
            <NumberField
              value={value.dividevalue}
              onChange={(n) => onChange((c) => ({ ...c, dividevalue: n }))}
              min={0.0001}
            />
          </FieldRow>
          <FieldRow label={text.standValue}>
            <NumberField
              value={value.standvalue}
              onChange={(n) => onChange((c) => ({ ...c, standvalue: n }))}
              min={0.0001}
            />
          </FieldRow>
          <FieldRow label={text.condition}>
            <Toggle
              checked={value.condition}
              onCheckedChange={(n) => onChange((c) => ({ ...c, condition: n }))}
              label={text.condition}
            />
          </FieldRow>
        </FieldGrid>
      </Section>

      {value.isusesubbarcodes ? (
        <Section
          title={text.refBarcodes}
          action={
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() =>
                setRef((rows) => [
                  ...rows,
                  {
                    guid_fixed: "",
                    names: [],
                    item_unit_code: "",
                    itemunitnames: [],
                    barcode: "",
                    condition: false,
                    dividevalue: 1,
                    standvalue: 1,
                    qty: 1,
                  },
                ])
              }
            >
              <Plus className="mr-1 h-4 w-4" />
              {text.refBarcodeAdd}
            </Button>
          }
        >
          <div className="space-y-2">
            {value.refbarcodes.length === 0 ? (
              <p className="text-sm text-muted-foreground">—</p>
            ) : (
              value.refbarcodes.map((entry, idx) => (
                <div
                  key={idx}
                  className="grid grid-cols-1 items-center gap-2 rounded-md border border-border p-2 md:grid-cols-[1.5fr_1fr_1fr_1fr_40px]"
                >
                  <Input
                    placeholder={text.bomBarcode}
                    value={entry.barcode}
                    onChange={(event) =>
                      setRef((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, barcode: event.target.value } : row)),
                      )
                    }
                  />
                  <NumberField
                    value={entry.qty}
                    onChange={(n) => setRef((rows) => rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, qty: n } : row)))}
                    min={0}
                  />
                  <NumberField
                    value={entry.dividevalue}
                    onChange={(n) =>
                      setRef((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, dividevalue: n } : row)),
                      )
                    }
                    min={0.0001}
                  />
                  <NumberField
                    value={entry.standvalue}
                    onChange={(n) =>
                      setRef((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, standvalue: n } : row)),
                      )
                    }
                    min={0.0001}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => setRef((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))
            )}
          </div>
        </Section>
      ) : null}
    </div>
  );
}

function TabBom({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const setBom = useCallback(
    (mutator: (rows: BOMProductBarcode[]) => BOMProductBarcode[]) =>
      onChange((c) => ({ ...c, bom: mutator(c.bom) })),
    [onChange],
  );
  return (
    <Section
      title={text.bom}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setBom((rows) => [
              ...rows,
              {
                barcodeguidfixed: "",
                names: [],
                item_unit_code: "",
                itemunitnames: [],
                barcode: "",
                qty: 1,
              },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {text.bomAdd}
        </Button>
      }
    >
      {value.bom.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-2">
          {value.bom.map((entry, idx) => (
            <div key={idx} className="grid grid-cols-1 items-center gap-2 md:grid-cols-[1.5fr_1fr_40px]">
              <Input
                placeholder={text.bomBarcode}
                value={entry.barcode}
                onChange={(event) =>
                  setBom((rows) =>
                    rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, barcode: event.target.value } : row)),
                  )
                }
              />
              <NumberField
                value={entry.qty ?? 1}
                onChange={(n) => setBom((rows) => rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, qty: n } : row)))}
                min={0}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setBom((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

function TabMedia({
  value,
  onChange,
  text,
  auth,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  auth: AuthSession | null;
}) {
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string>("");

  const handleUpload = useCallback(
    async (file: File | null, target: "main" | "gallery") => {
      if (!file) return;
      setUploading(true);
      setUploadError("");
      const result = await uploadProductImage(auth, file);
      if (!result.success || !result.data?.url) {
        setUploadError(result.message ?? "Upload failed");
      } else if (target === "main") {
        onChange((c) => ({ ...c, imageuri: result.data!.url }));
      } else {
        onChange((c) => ({
          ...c,
          images: [...c.images, { xorder: c.images.length + 1, uri: result.data!.url }],
        }));
      }
      setUploading(false);
    },
    [auth, onChange],
  );

  return (
    <div className="space-y-4">
      <Section title={text.useImageOrColor}>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={value.useimageorcolor}
              onChange={() => onChange((c) => ({ ...c, useimageorcolor: true }))}
            />
            <ImagePlus className="h-4 w-4" />
            <span>{text.imageMain}</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={!value.useimageorcolor}
              onChange={() => onChange((c) => ({ ...c, useimageorcolor: false }))}
            />
            <span className="inline-block size-4 rounded border" style={{ background: value.colorselecthex || "#888" }} />
            <span>{text.colorPick}</span>
          </label>
        </div>
      </Section>

      {value.useimageorcolor ? (
        <>
          <Section title={text.imageMain}>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <div className="size-32 overflow-hidden rounded-lg border border-border bg-muted">
                {value.imageuri ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={value.imageuri}
                    alt={text.imageMain}
                    className="size-full object-cover"
                  />
                ) : (
                  <div className="flex size-full items-center justify-center text-xs text-muted-foreground">
                    {text.noData}
                  </div>
                )}
              </div>
              <div className="space-y-2">
                <Input
                  placeholder="https://… or upload"
                  value={value.imageuri}
                  onChange={(event) => onChange((c) => ({ ...c, imageuri: event.target.value }))}
                />
                <div className="flex flex-wrap items-center gap-2">
                  <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input px-3 py-1.5 text-xs hover:bg-muted">
                    <Upload className="h-3.5 w-3.5" />
                    {uploading ? text.pickerLoading : text.imageUpload}
                    <input
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(event) => handleUpload(event.target.files?.[0] ?? null, "main")}
                      disabled={uploading}
                    />
                  </label>
                  {value.imageuri ? (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => onChange((c) => ({ ...c, imageuri: "" }))}
                    >
                      <Trash2 className="mr-1 h-3.5 w-3.5" />
                      {text.imageDelete}
                    </Button>
                  ) : null}
                </div>
                {uploadError ? <p className="text-xs text-destructive">{uploadError}</p> : null}
              </div>
            </div>
          </Section>
          <Section
            title={text.imageGallery}
            action={
              <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input px-3 py-1.5 text-xs hover:bg-muted">
                <Plus className="h-3.5 w-3.5" />
                {text.imageGallery}
                <input
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(event) => handleUpload(event.target.files?.[0] ?? null, "gallery")}
                  disabled={uploading}
                />
              </label>
            }
          >
            {value.images.length === 0 ? (
              <p className="text-sm text-muted-foreground">—</p>
            ) : (
              <ul className="grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6">
                {value.images.map((img, idx) => (
                  <li key={idx} className="relative overflow-hidden rounded-md border border-border">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={img.uri} alt={`#${img.xorder}`} className="aspect-square w-full object-cover" />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="absolute right-1 top-1 size-6 bg-background/80"
                      onClick={() =>
                        onChange((c) => ({ ...c, images: c.images.filter((_, imgIdx) => imgIdx !== idx) }))
                      }
                      aria-label={text.imageDelete}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </Section>
        </>
      ) : (
        <Section title={text.colorPick}>
          <FieldGrid>
            <FieldRow label={text.colorName}>
              <Input
                value={value.colorselect}
                onChange={(event) => onChange((c) => ({ ...c, colorselect: event.target.value }))}
              />
            </FieldRow>
            <FieldRow label={text.colorHex}>
              <div className="flex items-center gap-2">
                <Input
                  type="color"
                  value={value.colorselecthex || "#888888"}
                  onChange={(event) => onChange((c) => ({ ...c, colorselecthex: event.target.value }))}
                  className="h-10 w-16 p-1"
                />
                <Input
                  value={value.colorselecthex}
                  onChange={(event) => onChange((c) => ({ ...c, colorselecthex: event.target.value }))}
                  placeholder="#RRGGBB"
                />
              </div>
            </FieldRow>
          </FieldGrid>
        </Section>
      )}
    </div>
  );
}

function TabRestaurant({
  value,
  onChange,
  text,
  auth,
  language,
  shopLanguages,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  auth: AuthSession | null;
  language: LanguageCode | string;
  shopLanguages: string[];
}) {
  const updR = useCallback(
    <K extends keyof ProductBarcode["restaurant"]>(key: K, val: ProductBarcode["restaurant"][K]) =>
      onChange((c) => ({ ...c, restaurant: { ...c.restaurant, [key]: val } })),
    [onChange],
  );
  return (
    <div className="space-y-4">
      <Section title={text.tabRestaurant}>
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          <Toggle checked={value.restaurant.isforrestaurant} onCheckedChange={(n) => updR("isforrestaurant", n)} label={text.isForRestaurant} />
          <Toggle checked={value.restaurant.isfortakeaway} onCheckedChange={(n) => updR("isfortakeaway", n)} label={text.isForTakeaway} />
          <Toggle checked={value.restaurant.isfordelivery} onCheckedChange={(n) => updR("isfordelivery", n)} label={text.isForDelivery} />
          <Toggle checked={value.restaurant.isforcustomer} onCheckedChange={(n) => updR("isforcustomer", n)} label={text.isForCustomer} />
          <Toggle checked={value.restaurant.isforcustomerpreorder} onCheckedChange={(n) => updR("isforcustomerpreorder", n)} label={text.isForCustomerPreOrder} />
          <Toggle checked={value.isalacarte} onCheckedChange={(n) => onChange((c) => ({ ...c, isalacarte: n }))} label={text.isALaCarte} />
          <Toggle checked={value.isstockforrestaurant} onCheckedChange={(n) => onChange((c) => ({ ...c, isstockforrestaurant: n }))} label={text.isStockForRestaurant} />
          <Toggle checked={value.issplitunitprint} onCheckedChange={(n) => onChange((c) => ({ ...c, issplitunitprint: n }))} label={text.isSplitUnitPrint} />
          <Toggle checked={value.isonlystaff} onCheckedChange={(n) => onChange((c) => ({ ...c, isonlystaff: n }))} label={text.isOnlyStaff} />
        </div>
        <div className="mt-3">
          <FieldRow label={text.foodType}>
            <select
              className="h-10 rounded-lg border border-input bg-background px-3 text-sm"
              value={value.foodtype}
              onChange={(event) => onChange((c) => ({ ...c, foodtype: Number(event.target.value) as ProductBarcode["foodtype"] }))}
            >
              <option value={FOOD_TYPE.FOOD}>{text.foodTypeFood}</option>
              <option value={FOOD_TYPE.DRINK}>{text.foodTypeDrink}</option>
              <option value={FOOD_TYPE.ALCOHOL}>{text.foodTypeAlcohol}</option>
              <option value={FOOD_TYPE.OTHER}>{text.foodTypeOther}</option>
            </select>
          </FieldRow>
        </div>
      </Section>

      <OrderTypesEditor value={value} onChange={onChange} text={text} language={language} auth={auth} />
      <OptionsEditor value={value} onChange={onChange} text={text} shopLanguages={shopLanguages} />
    </div>
  );
}

function OrderTypesEditor({
  value,
  onChange,
  text,
  language,
  auth,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  language: LanguageCode | string;
  auth: AuthSession | null;
}) {
  const [picker, setPicker] = useState<{ open: boolean; idx?: number }>({ open: false });
  const setRows = useCallback(
    (mutator: (rows: ProductOrderType[]) => ProductOrderType[]) =>
      onChange((c) => ({ ...c, ordertypes: mutator(c.ordertypes) })),
    [onChange],
  );
  return (
    <Section
      title={text.orderTypes}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [...rows, { guidfixed: "", code: "", names: [], chargeprice: 0 }])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {text.orderTypeAdd}
        </Button>
      }
    >
      {value.ordertypes.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-2">
          {value.ordertypes.map((entry, idx) => (
            <div key={idx} className="grid grid-cols-1 items-center gap-2 md:grid-cols-[2fr_1fr_40px]">
              <button
                type="button"
                onClick={() => setPicker({ open: true, idx })}
                className="flex h-10 w-full items-center justify-between rounded-lg border border-input bg-background px-3 text-left text-sm hover:bg-muted/40"
              >
                <span className="truncate">{pickName(entry.names, language) || entry.code || "—"}</span>
                <span className="text-xs text-muted-foreground">{entry.code}</span>
              </button>
              <NumberField
                value={entry.chargeprice ?? 0}
                onChange={(n) =>
                  setRows((rows) => rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, chargeprice: n } : row)))
                }
                step="any"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
      <MasterPicker
        open={picker.open}
        onClose={() => setPicker({ open: false })}
        auth={auth}
        language={language}
        master="ordertype"
        title={text.orderTypes}
        onSelect={(entry) =>
          setRows((rows) =>
            rows.map((row, rowIdx) =>
              rowIdx === picker.idx
                ? { ...row, guidfixed: entry.guidfixed, code: entry.code, names: entry.names }
                : row,
            ),
          )
        }
      />
    </Section>
  );
}

function OptionsEditor({
  value,
  onChange,
  text,
  shopLanguages,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  shopLanguages: string[];
}) {
  const setOptions = useCallback(
    (mutator: (rows: ProductOption[]) => ProductOption[]) =>
      onChange((c) => ({ ...c, options: mutator(c.options) })),
    [onChange],
  );
  return (
    <Section
      title={text.options}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setOptions((rows) => [
              ...rows,
              { guid: cryptoRandomId(), names: [], choicetype: 0, choices: [] },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {text.optionAdd}
        </Button>
      }
    >
      {value.options.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-4">
          {value.options.map((opt, optIdx) => (
            <div key={opt.guid} className="space-y-2 rounded-md border border-border p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground">#{optIdx + 1}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => setOptions((rows) => rows.filter((_, idx) => idx !== optIdx))}
                  aria-label={text.delete}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
              <NamesEditor
                names={opt.names}
                onChange={(next) =>
                  setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, names: next } : row)))
                }
                languages={shopLanguages}
                label={text.optionType}
              />
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-1 text-sm">
                  <input
                    type="radio"
                    checked={opt.choicetype === 0}
                    onChange={() =>
                      setOptions((rows) =>
                        rows.map((row, idx) => (idx === optIdx ? { ...row, choicetype: 0 } : row)),
                      )
                    }
                  />
                  {text.optionTypeMulti}
                </label>
                <label className="flex items-center gap-1 text-sm">
                  <input
                    type="radio"
                    checked={opt.choicetype === 1}
                    onChange={() =>
                      setOptions((rows) =>
                        rows.map((row, idx) => (idx === optIdx ? { ...row, choicetype: 1 } : row)),
                      )
                    }
                  />
                  {text.optionTypeSingle}
                </label>
              </div>
              {opt.choicetype === 0 ? (
                <FieldGrid>
                  <FieldRow label={text.optionMinSelect}>
                    <NumberField
                      value={opt.minselect ?? 0}
                      onChange={(n) =>
                        setOptions((rows) =>
                          rows.map((row, idx) => (idx === optIdx ? { ...row, minselect: n } : row)),
                        )
                      }
                      min={0}
                      step={1}
                    />
                  </FieldRow>
                  <FieldRow label={text.optionMaxSelect}>
                    <NumberField
                      value={opt.maxselect ?? 0}
                      onChange={(n) =>
                        setOptions((rows) =>
                          rows.map((row, idx) => (idx === optIdx ? { ...row, maxselect: n } : row)),
                        )
                      }
                      min={0}
                      step={1}
                    />
                  </FieldRow>
                </FieldGrid>
              ) : null}

              <ChoicesEditor
                option={opt}
                onChange={(updated) =>
                  setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? updated : row)))
                }
                text={text}
                shopLanguages={shopLanguages}
              />
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

function ChoicesEditor({
  option,
  onChange,
  text,
  shopLanguages,
}: {
  option: ProductOption;
  onChange: (next: ProductOption) => void;
  text: Text;
  shopLanguages: string[];
}) {
  const setChoices = (mutator: (rows: ProductChoice[]) => ProductChoice[]) =>
    onChange({ ...option, choices: mutator(option.choices) });
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">{text.optionChoices}</span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setChoices((rows) => [
              ...rows,
              { guid: cryptoRandomId(), names: [], isstock: false, isdefault: false },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {text.optionChoiceAdd}
        </Button>
      </div>
      {option.choices.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        option.choices.map((choice, choiceIdx) => (
          <div key={choice.guid} className="grid grid-cols-1 gap-2 rounded border border-border p-2 md:grid-cols-[2fr_1fr_auto]">
            <NamesEditor
              names={choice.names}
              onChange={(next) =>
                setChoices((rows) =>
                  rows.map((row, idx) => (idx === choiceIdx ? { ...row, names: next } : row)),
                )
              }
              languages={shopLanguages}
              label={text.optionChoices}
            />
            <div className="flex flex-col gap-1">
              <Input
                placeholder={text.optionChoicePrice}
                value={choice.price ?? ""}
                onChange={(event) =>
                  setChoices((rows) =>
                    rows.map((row, idx) => (idx === choiceIdx ? { ...row, price: event.target.value } : row)),
                  )
                }
              />
              <Toggle
                checked={choice.isstock}
                onCheckedChange={(n) =>
                  setChoices((rows) =>
                    rows.map((row, idx) => (idx === choiceIdx ? { ...row, isstock: n } : row)),
                  )
                }
                label={text.optionChoiceIsStock}
              />
              <Toggle
                checked={choice.isdefault}
                onCheckedChange={(n) =>
                  setChoices((rows) =>
                    rows.map((row, idx) => (idx === choiceIdx ? { ...row, isdefault: n } : row)),
                  )
                }
                label={text.optionChoiceIsDefault}
              />
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setChoices((rows) => rows.filter((_, idx) => idx !== choiceIdx))}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))
      )}
    </div>
  );
}

function TabTimeForSales({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const setRows = useCallback(
    (mutator: (rows: ProductTimeForSale[]) => ProductTimeForSale[]) =>
      onChange((c) => ({ ...c, timeforsales: mutator(c.timeforsales) })),
    [onChange],
  );
  const dayLabels = useMemo(
    () => [text.sun, text.mon, text.tue, text.wed, text.thu, text.fri, text.sat],
    [text.sun, text.mon, text.tue, text.wed, text.thu, text.fri, text.sat],
  );
  return (
    <Section
      title={text.timeForSales}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [
              ...rows,
              { fromdate: "", todate: "", fromtime: "", totime: "", daysofweek: [] },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {text.timeForSalesAdd}
        </Button>
      }
    >
      {value.timeforsales.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-3">
          {value.timeforsales.map((entry, idx) => (
            <div key={idx} className="space-y-2 rounded border border-border p-2">
              <FieldGrid>
                <FieldRow label={text.fromDate}>
                  <Input
                    type="date"
                    value={entry.fromdate}
                    onChange={(event) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, fromdate: event.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={text.toDate}>
                  <Input
                    type="date"
                    value={entry.todate}
                    onChange={(event) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, todate: event.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={text.fromTime}>
                  <Input
                    type="time"
                    value={entry.fromtime}
                    onChange={(event) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, fromtime: event.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={text.toTime}>
                  <Input
                    type="time"
                    value={entry.totime}
                    onChange={(event) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, totime: event.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
              </FieldGrid>
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm font-medium">{text.daysOfWeek}</span>
                {dayLabels.map((label, dayIdx) => {
                  const selected = entry.daysofweek?.includes(dayIdx);
                  return (
                    <button
                      key={dayIdx}
                      type="button"
                      onClick={() =>
                        setRows((rows) =>
                          rows.map((row, rowIdx) => {
                            if (rowIdx !== idx) return row;
                            const current = new Set(row.daysofweek ?? []);
                            if (current.has(dayIdx)) current.delete(dayIdx);
                            else current.add(dayIdx);
                            return { ...row, daysofweek: Array.from(current).sort() };
                          }),
                        )
                      }
                      className={cn(
                        "rounded-full border px-3 py-1 text-xs",
                        selected ? "border-primary bg-primary text-primary-foreground" : "border-input hover:bg-muted",
                      )}
                    >
                      {label}
                    </button>
                  );
                })}
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="ml-auto"
                  onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                  aria-label={text.delete}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

function TabDimensions({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const setRows = useCallback(
    (mutator: (rows: ProductDimension[]) => ProductDimension[]) =>
      onChange((c) => ({ ...c, dimensions: mutator(c.dimensions) })),
    [onChange],
  );
  return (
    <Section
      title={text.dimensions}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [
              ...rows,
              {
                guidfixed: "",
                names: [],
                isdisabled: false,
                item: { guidfixed: "", names: [], isdisabled: false },
              },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {text.dimensionAdd}
        </Button>
      }
    >
      {value.dimensions.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-2">
          {value.dimensions.map((entry, idx) => (
            <div key={idx} className="grid grid-cols-1 items-center gap-2 md:grid-cols-[1fr_1fr_40px]">
              <Input
                placeholder={text.dimensionGroup}
                value={pickName(entry.names, "th")}
                onChange={(event) =>
                  setRows((rows) =>
                    rows.map((row, rowIdx) =>
                      rowIdx === idx
                        ? { ...row, names: setNameXEntry(row.names, "th", event.target.value) }
                        : row,
                    ),
                  )
                }
              />
              <Input
                placeholder={text.dimensionItem}
                value={pickName(entry.item.names, "th")}
                onChange={(event) =>
                  setRows((rows) =>
                    rows.map((row, rowIdx) =>
                      rowIdx === idx
                        ? {
                            ...row,
                            item: { ...row.item, names: setNameXEntry(row.item.names, "th", event.target.value) },
                          }
                        : row,
                    ),
                  )
                }
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

function TabBusinessBranch({
  value,
  onChange,
  text,
  auth,
  language,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  auth: AuthSession | null;
  language: LanguageCode | string;
}) {
  return (
    <div className="space-y-4">
      <BusinessTypesEditor value={value} onChange={onChange} text={text} auth={auth} language={language} />
      <IgnoreBranchesEditor value={value} onChange={onChange} text={text} auth={auth} language={language} />
    </div>
  );
}

function BusinessTypesEditor({
  value,
  onChange,
  text,
  auth,
  language,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  auth: AuthSession | null;
  language: LanguageCode | string;
}) {
  const [picker, setPicker] = useState(false);
  const setRows = (mutator: (rows: ProductBarcodeBusinessType[]) => ProductBarcodeBusinessType[]) =>
    onChange((c) => ({ ...c, businesstypes: mutator(c.businesstypes) }));
  return (
    <Section
      title={text.businessTypes}
      action={
        <Button type="button" variant="outline" size="sm" onClick={() => setPicker(true)}>
          <Plus className="mr-1 h-4 w-4" />
          {text.businessTypeAdd}
        </Button>
      }
    >
      {value.businesstypes.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <ul className="space-y-1">
          {value.businesstypes.map((entry, idx) => (
            <li key={idx} className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1.5">
              <span className="truncate text-sm">{pickName(entry.names, language) || entry.code}</span>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Minus className="h-4 w-4" />
              </Button>
            </li>
          ))}
        </ul>
      )}
      <MasterPicker
        open={picker}
        onClose={() => setPicker(false)}
        auth={auth}
        language={language}
        master="businesstype"
        title={text.businessTypes}
        onSelect={(entry) =>
          setRows((rows) => [
            ...rows,
            { guidfixed: entry.guidfixed, code: entry.code, names: entry.names, isignore: false },
          ])
        }
      />
    </Section>
  );
}

function IgnoreBranchesEditor({
  value,
  onChange,
  text,
  auth,
  language,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  auth: AuthSession | null;
  language: LanguageCode | string;
}) {
  const [picker, setPicker] = useState(false);
  const setRows = (mutator: (rows: ProductBarcodeBranch[]) => ProductBarcodeBranch[]) =>
    onChange((c) => ({ ...c, ignorebranches: mutator(c.ignorebranches) }));
  return (
    <Section
      title={text.ignoreBranches}
      action={
        <Button type="button" variant="outline" size="sm" onClick={() => setPicker(true)}>
          <Plus className="mr-1 h-4 w-4" />
          {text.ignoreBranchAdd}
        </Button>
      }
    >
      <p className="mb-2 text-xs text-muted-foreground">{text.branchHint}</p>
      {value.ignorebranches.length === 0 ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <ul className="space-y-1">
          {value.ignorebranches.map((entry, idx) => (
            <li key={idx} className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1.5">
              <span className="truncate text-sm">{pickName(entry.names, language) || entry.code}</span>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Minus className="h-4 w-4" />
              </Button>
            </li>
          ))}
        </ul>
      )}
      <MasterPicker
        open={picker}
        onClose={() => setPicker(false)}
        auth={auth}
        language={language}
        master="branch"
        title={text.ignoreBranches}
        onSelect={(entry) =>
          setRows((rows) => [
            ...rows,
            { guidfixed: entry.guidfixed, code: entry.code, names: entry.names, isignore: true },
          ])
        }
      />
    </Section>
  );
}

function TabMisc({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  return (
    <div className="space-y-4">
      <Section title={text.tabMisc}>
        <FieldGrid>
          <FieldRow label={text.isAlert}>
            <Toggle
              checked={value.isalert}
              onCheckedChange={(n) => onChange((c) => ({ ...c, isalert: n }))}
              label={text.isAlert}
            />
          </FieldRow>
          {value.isalert ? (
            <FieldRow label={text.alertDescription}>
              <textarea
                className="min-h-[80px] rounded-lg border border-input bg-background p-2 text-sm"
                value={value.alertdescription}
                onChange={(event) => onChange((c) => ({ ...c, alertdescription: event.target.value }))}
              />
            </FieldRow>
          ) : null}
        </FieldGrid>
      </Section>
      <Section title={text.description}>
        <textarea
          className="min-h-[120px] w-full rounded-lg border border-input bg-background p-2 text-sm"
          value={value.description}
          onChange={(event) => onChange((c) => ({ ...c, description: event.target.value }))}
        />
      </Section>
      <Section title="ID / GUID">
        <FieldGrid>
          <FieldRow label={text.guid}>
            <Input value={value.guidfixed} readOnly />
          </FieldRow>
          <FieldRow label={text.shopId}>
            <Input value={value.shopid ?? ""} readOnly />
          </FieldRow>
          <FieldRow label={text.itemGuid}>
            <Input value={value.item_guid} readOnly />
          </FieldRow>
          <FieldRow label={text.refGuid}>
            <Input value={value.refguidfixed} readOnly />
          </FieldRow>
        </FieldGrid>
      </Section>
    </div>
  );
}

// ─── Internal helpers ────────────────────────────────────────────────────

function cryptoRandomId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return (crypto as Crypto).randomUUID();
  }
  return Math.random().toString(36).slice(2);
}

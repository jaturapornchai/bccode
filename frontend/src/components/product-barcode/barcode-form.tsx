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
  Info,
  Loader2,
  Minus,
  Plus,
  Save,
  Trash2,
  Upload,
  Wand2,
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
  useEffect,
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
  ITEM_TYPE,
  MARKETPLACE_PLATFORMS,
  MATERIAL_TYPE,
  emptyMarketplaceProductMap,
  type MarketplacePlatform,
  type MarketplaceProductMap,
  type NameX,
  type Product,
  type ProductBarcode,
  type ProductChoice,
  type ProductImage,
  type ProductManufacturer,
  type ProductOption,
  type ProductPrice,
  type ProductSupplier,
} from "@/lib/product-barcode/types";
import { isValidBarcode, pickName, setNameXEntry } from "@/lib/product-barcode/utils";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import type { AuthSession } from "@/lib/workspace-models";
import { NamesEditor } from "./names-editor";

// ─── Tab definitions ──────────────────────────────────────────────────────

type TabKey =
  | "basic"
  | "pricing"
  | "media"
  | "logistics"
  | "marketplace"
  | "product_detail";

interface TabDef {
  key: TabKey;
  label: keyof ReturnType<typeof getBarcodeText> | "tabProductDetail";
}

const TABS: TabDef[] = [
  { key: "basic", label: "tabBasic" },
  { key: "pricing", label: "tabPricing" },
  { key: "media", label: "tabMedia" },
  { key: "logistics", label: "tabDimensions" },
  { key: "marketplace", label: "tabMarketplace" },
];

// ─── Props ────────────────────────────────────────────────────────────────

export interface ProductBarcodeFormDialogProps {
  open: boolean;
  mode: "create" | "edit";
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  onSave: (value: ProductBarcode) => void;
  /** Optional: called instead of onSave when "Save & add new" is clicked in create mode. */
  onSaveAndNew?: (value: ProductBarcode) => void;
  onCancel: () => void;
  saving?: boolean;
  language: LanguageCode | string;
  auth: AuthSession | null;
  shopLanguages?: string[];
  /** Optional extra controls (e.g. quick links to BOM / Price History / Label print). */
  extraActions?: ReactNode;
  embedded?: boolean;
  companyGuid?: string;
}

export function ProductBarcodeFormDialog(props: ProductBarcodeFormDialogProps) {
  const { open, mode, value, onChange, onSave, onSaveAndNew, onCancel, saving = false, language, auth, extraActions, embedded = false, companyGuid } = props;
  const text = getBarcodeText(language);
  const [tab, setTab] = useState<TabKey>("basic");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const shopLanguages = useMemo(
    () => (props.shopLanguages && props.shopLanguages.length > 0 ? props.shopLanguages : ["th", "en"]),
    [props.shopLanguages],
  );

  const [productDetail, setProductDetail] = useState<Product | null>(null);
  const [loadingProductDetail, setLoadingProductDetail] = useState(false);

  useEffect(() => {
    if (!auth || !value.item_guid) {
      setProductDetail(null);
      return;
    }
    let active = true;
    setLoadingProductDetail(true);
    fetch(`/api/product/${encodeURIComponent(value.item_guid)}`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
        "x-bc-backend-url": auth.backendUrl,
      },
    })
      .then((res) => res.json())
      .then((resData) => {
        if (active && resData && resData.success !== false) {
          setProductDetail(resData.data || null);
        }
      })
      .catch((err) => {
        console.error("Failed to load product details", err);
      })
      .finally(() => {
        if (active) setLoadingProductDetail(false);
      });

    return () => {
      active = false;
    };
  }, [auth, value.item_guid]);

  useEffect(() => {
    if (productDetail) {
      onChange((current) => {
        const nextItemType = typeof productDetail.item_type === "number" ? productDetail.item_type as ProductBarcode["item_type"] : current.item_type;
        const nextMaterialType = typeof productDetail.materialtype === "number" ? productDetail.materialtype as ProductBarcode["materialtype"] : current.materialtype;
        const nextVatCal = typeof productDetail.vat_type === "number" ? productDetail.vat_type : current.vatcal;
        const nextSumPoint = typeof productDetail.issumpoint === "boolean" ? productDetail.issumpoint : current.issumpoint;

        if (
          current.item_type !== nextItemType ||
          current.materialtype !== nextMaterialType ||
          current.vatcal !== nextVatCal ||
          current.issumpoint !== nextSumPoint
        ) {
          return {
            ...current,
            item_type: nextItemType,
            materialtype: nextMaterialType,
            vatcal: nextVatCal,
            issumpoint: nextSumPoint,
          };
        }
        return current;
      });
    }
  }, [productDetail, onChange]);

  const visibleTabs = useMemo<TabDef[]>(() => {
    if (value.item_guid) {
      return [
        { key: "basic", label: "tabBasic" },
        { key: "pricing", label: "tabPricing" },
        { key: "logistics", label: "tabDimensions" },
        { key: "marketplace", label: "tabMarketplace" },
        { key: "product_detail", label: "tabProductDetail" },
      ];
    }
    return TABS;
  }, [value.item_guid]);

  useEffect(() => {
    const isCurrentTabVisible = visibleTabs.some((t) => t.key === tab);
    if (!isCurrentTabVisible) {
      setTab("basic");
    }
  }, [visibleTabs, tab]);

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

  const saveAndNewIfValid = useCallback(() => {
    if (validate()) {
      if (onSaveAndNew) onSaveAndNew(value);
      else onSave(value);
    }
  }, [validate, onSave, onSaveAndNew, value]);

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
            {visibleTabs.map((tabDef) => {
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
                  {text[tabDef.label as keyof typeof text]}
                </button>
              );
            })}
          </div>
        </div>

        {/* Body */}
        <div className="min-h-0 flex-1 overflow-y-auto bg-background px-4 py-4">
          {tab === "basic" && (
            <TabBasic value={value} onChange={onChange} text={text} errors={errors} shopLanguages={shopLanguages} auth={auth} language={language} companyGuid={companyGuid} />
          )}
          {tab === "pricing" && <TabPricing value={value} onChange={onChange} text={text} />}
          {tab === "media" && (
            <TabMedia value={value} onChange={onChange} text={text} auth={auth} />
          )}
          {tab === "logistics" && (
            <TabLogistics value={value} onChange={onChange} text={text} />
          )}
          {tab === "marketplace" && (
            <div className="space-y-4">
              {MARKETPLACE_PLATFORMS.map((platform) => (
                <TabMarketplace
                  key={platform}
                  platform={platform}
                  value={value}
                  onChange={onChange}
                  text={text}
                />
              ))}
            </div>
          )}
          {tab === "product_detail" && (
            <TabProductDetail productDetail={productDetail} loading={loadingProductDetail} text={text} language={language} />
          )}

        </div>

        {/* Footer */}
        <div className="flex items-center justify-between gap-2 border-t border-border bg-card/95 px-4 py-3">
          <div className="flex gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                const idx = visibleTabs.findIndex((t) => t.key === tab);
                if (idx > 0) setTab(visibleTabs[idx - 1].key);
              }}
              disabled={visibleTabs.findIndex((t) => t.key === tab) === 0}
              aria-label="Previous tab"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => {
                const idx = visibleTabs.findIndex((t) => t.key === tab);
                if (idx < visibleTabs.length - 1) setTab(visibleTabs[idx + 1].key);
              }}
              disabled={visibleTabs.findIndex((t) => t.key === tab) === visibleTabs.length - 1}
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
            {mode === "create" && (
              <Button type="button" variant="secondary" size="sm" disabled={saving} onClick={saveAndNewIfValid}>
                {saving ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : <Plus className="mr-1 h-4 w-4" />}
                {text.saveAndNew}
              </Button>
            )}
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
    <div className="flex flex-col gap-1 text-sm">
      <span className="font-medium text-foreground">
        {label}
        {required ? <span className="ml-1 text-destructive">*</span> : null}
      </span>
      {children}
      {hint ? <span className="text-xs text-muted-foreground">{hint}</span> : null}
    </div>
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
    <fieldset className="rounded-md border border-border bg-background px-2.5 py-2">
      <legend className="px-1 text-xs font-semibold text-muted-foreground">{label}</legend>
      <div className="grid gap-1.5 sm:grid-cols-2 lg:grid-cols-3">
        {options.map((option) => (
          <label
            key={String(option.value)}
            className={cn(
              "flex min-h-7 w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1 text-xs font-medium hover:bg-muted/60",
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
            <span className="whitespace-normal break-words">{option.label}</span>
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
  companyGuid,
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
  companyGuid?: string;
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
          aria-label={label}
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
          companyGuid={companyGuid}
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
  companyGuid,
}: {
  auth: AuthSession | null;
  language: LanguageCode | string;
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
  errors: Record<string, string>;
  shopLanguages: string[];
  companyGuid?: string;
}) {
  const upd = useCallback(
    <K extends keyof ProductBarcode>(key: K, val: ProductBarcode[K]) =>
      onChange((current) => ({ ...current, [key]: val })),
    [onChange],
  );
  const setItemType = useCallback(
    (nextType: ProductBarcode["item_type"]) =>
      onChange((current) => ({ ...current, item_type: nextType })),
    [onChange],
  );
  const hasProduct = !!value.item_guid;

  const itemTypeLabel = value.item_type === 0
    ? text.itemTypeStock
    : value.item_type === 1
    ? text.itemTypeService
    : value.item_type === 2
    ? text.itemTypeSet
    : "-";

  const materialTypeLabel = value.materialtype === 0
    ? text.materialGeneral
    : value.materialtype === 1
    ? text.materialMaterial
    : value.materialtype === 2
    ? text.materialSemiFinished
    : value.materialtype === 3
    ? text.materialSet
    : value.materialtype === 4
    ? text.materialAgricultural
    : "-";

  const vatCalLabel = value.vatcal === 0 ? text.vatIncluded : value.vatcal === 1 ? text.vatExcluded : "-";
  const sumPointLabel = value.issumpoint ? text.isSumPointYes : text.isSumPointNo;

  return (
    <div className="space-y-3">
      {hasProduct && (
        <div className="flex items-start gap-3 rounded-lg border border-blue-200/50 bg-blue-50/50 p-3.5 dark:border-blue-900/30 dark:bg-blue-950/20 text-xs text-blue-800 dark:text-blue-300">
          <Info className="h-4.5 w-4.5 shrink-0 text-blue-500 mt-0.5" />
          <div className="space-y-1">
            <p className="font-semibold text-blue-900 dark:text-blue-200">
              {language === "th"
                ? `บาร์โค้ดนี้ผูกอยู่กับสินค้าหลัก: ${value.itemcode} — ${pickName(value.names, language)}`
                : `This barcode is linked to product: ${value.itemcode} — ${pickName(value.names, language)}`}
            </p>
            <p className="text-muted-foreground/90 dark:text-muted-foreground/80 leading-normal">
              {language === "th"
                ? "ข้อมูลชื่อสินค้า ประเภทไอเทม ประเภทสินค้า และประเภทภาษี จะถูกควบคุมตามสินค้าหลักโดยอัตโนมัติ"
                : "Product name, item type, product type, and tax configuration are automatically inherited from the parent product."}
            </p>
          </div>
        </div>
      )}

      {hasProduct && (
        <Section title={language === "th" ? "ข้อมูลควบคุมจากสินค้าหลัก" : "Inherited Control Properties"}>
          <FieldGrid>
            <ReadOnlyField label={text.itemTypeLabel} value={itemTypeLabel} />
            <ReadOnlyField label={text.materialTypeLabel} value={materialTypeLabel} />
            <ReadOnlyField label={text.vatType} value={vatCalLabel} />
            <ReadOnlyField label={language === "th" ? "สะสมแต้ม" : "Sum Point"} value={sumPointLabel} />
          </FieldGrid>
        </Section>
      )}

      <Section title={text.tabBasic}>
        <div className="space-y-3">
          <FieldGrid>
            <FieldRow label={text.barcode} hint={text.barcodeHelp} required>
              <div className="flex gap-2">
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
                  className="flex-1"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  aria-label={text.generateBarcode}
                  title={text.generateBarcode}
                  onClick={() => {
                    const ts = String(Date.now());
                    // Produce a 13-digit numeric string (EAN-like)
                    const candidate = ts.slice(-13).padStart(13, "0");
                    upd("barcode", candidate);
                  }}
                >
                  <Wand2 className="h-4 w-4" />
                </Button>
              </div>
              {errors.barcode ? <p className="text-xs text-destructive">{errors.barcode}</p> : null}
            </FieldRow>
            <MasterField
              label={text.product}
              code={value.itemcode}
              names={value.names}
              master="product"
              language={language}
              auth={auth}
              companyGuid={companyGuid}
              onPick={(entry) =>
                onChange((current) => ({
                  ...current,
                  item_guid: entry.guidfixed,
                  itemcode: entry.code,
                  names: entry.names && entry.names.length > 0 ? entry.names : current.names,
                }))
              }
              onClear={() =>
                onChange((current) => ({
                  ...current,
                  item_guid: "",
                  itemcode: "",
                }))
              }
            />
            <MasterField
              label={text.itemUnit}
              code={value.item_unit_code}
              names={value.itemunitnames}
              master="unit"
              language={language}
              auth={auth}
              required
              error={errors.item_unit_code}
              companyGuid={companyGuid}
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
            language={language}
            disabled={hasProduct}
          />
        </div>
      </Section>

      <Section title={text.vatType}>
        <div className="grid gap-3 lg:grid-cols-1">
          <RadioOptionGroup
            label={text.vatType}
            value={value.vatcal}
            onChange={(next) => upd("vatcal", next)}
            options={[
              { value: 0, label: text.vatIncluded, disabled: hasProduct },
              { value: 1, label: text.vatExcluded, disabled: hasProduct },
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
          <Toggle
            checked={value.isdividend}
            onCheckedChange={(n) => upd("isdividend", n)}
            disabled={hasProduct}
            label={text.isDividend}
          />
          <Toggle
            checked={value.isdiscountpointofpurchase}
            onCheckedChange={(n) => upd("isdiscountpointofpurchase", n)}
            disabled={hasProduct}
            label={text.isDiscountPointOfPurchase}
          />
        </div>
      </Section>
    </div>
  );
}

function ReadOnlyField({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0 rounded-lg border border-border bg-muted/30 p-2">
      <span className="text-[11px] font-semibold text-muted-foreground block">{label}</span>
      <span className="mt-0.5 block truncate text-xs font-semibold text-foreground/90">{value}</span>
    </div>
  );
}

function TabProductDetail({
  productDetail,
  loading,
  text,
  language,
}: {
  productDetail: Product | null;
  loading: boolean;
  text: Text;
  language: string;
}) {
  if (loading) {
    return (
      <div className="flex flex-col justify-center items-center p-12 gap-3 text-sm text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
        <span>{text.loadingProductDetail}</span>
      </div>
    );
  }

  if (!productDetail) {
    return (
      <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-center text-sm text-destructive">
        {text.noProductDetailFound}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Alert Banner */}
      <div className="rounded-lg border border-border bg-muted/50 p-3.5 text-xs text-muted-foreground">
        <div className="flex items-start gap-2">
          <span className="text-base">ℹ️</span>
          <div className="flex-1">
            <span className="font-semibold block mb-0.5">{text.inheritedInfoBanner}</span>
            {text.inheritedInfoDetail}{" "}
            <strong> {productDetail.code} — {pickName(productDetail.names, language)}</strong>{" "}
            {text.inheritedInfoEditHint} <strong>{text.productMenuName}</strong>
          </div>
        </div>
      </div>

      {/* Grid container */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Left Column: Classification & Creditors */}
        <div className="space-y-4">
          <Section title={text.classificationSectionTitle}>
            <div className="grid gap-2 grid-cols-2">
              <ReadOnlyField label={text.group} value={productDetail.group_code ? `${productDetail.group_code} — ${pickName(productDetail.group_names, language)}` : "-"} />
              <ReadOnlyField label={text.groupSubOne} value={productDetail.groupsubonecode ? `${productDetail.groupsubonecode} — ${pickName(productDetail.groupsubonenames, language)}` : "-"} />
              <ReadOnlyField label={text.groupSubTwo} value={productDetail.groupsubtwocode ? `${productDetail.groupsubtwocode} — ${pickName(productDetail.groupsubtwonames, language)}` : "-"} />
              <ReadOnlyField label={text.brand} value={productDetail.brand_code ? `${productDetail.brand_code} — ${pickName(productDetail.brandnames, language)}` : "-"} />
              <ReadOnlyField label={text.category} value={productDetail.categorycode ? `${productDetail.categorycode} — ${pickName(productDetail.category_names, language)}` : "-"} />
              <ReadOnlyField label={text.classification} value={productDetail.classcode ? `${productDetail.classcode} — ${pickName(productDetail.classnames, language)}` : "-"} />
              <ReadOnlyField label={text.design} value={productDetail.designcode ? `${productDetail.designcode} — ${pickName(productDetail.designnames, language)}` : "-"} />
              <ReadOnlyField label={text.grade} value={productDetail.gradecode ? `${productDetail.gradecode} — ${pickName(productDetail.gradenames, language)}` : "-"} />
              <ReadOnlyField label={text.model} value={productDetail.modelcode ? `${productDetail.modelcode} — ${pickName(productDetail.modelnames, language)}` : "-"} />
              <ReadOnlyField label={text.pattern} value={productDetail.patterncode ? `${productDetail.patterncode} — ${pickName(productDetail.patternnames, language)}` : "-"} />
            </div>
          </Section>

          <Section title={text.creditorsSectionTitle}>
            <div className="space-y-3">
              <div>
                <span className="text-[11px] font-semibold text-muted-foreground block mb-1">{text.manufacturersLabel}</span>
                <div className="flex flex-wrap gap-1">
                  {!productDetail.manufacturers || productDetail.manufacturers.length === 0 ? (
                    <span className="text-xs text-muted-foreground">—</span>
                  ) : (
                    productDetail.manufacturers.map((m: ProductManufacturer) => (
                      <Badge key={m.guid_fixed} variant="outline" className="text-xs py-1">
                        {m.code} — {pickName(m.names, language)}
                      </Badge>
                    ))
                  )}
                </div>
              </div>
              <div className="border-t border-border/50 pt-2.5">
                <span className="text-[11px] font-semibold text-muted-foreground block mb-1">{text.suppliersLabel}</span>
                <div className="flex flex-wrap gap-1">
                  {!productDetail.suppliers || productDetail.suppliers.length === 0 ? (
                    <span className="text-xs text-muted-foreground">—</span>
                  ) : (
                    productDetail.suppliers.map((s: ProductSupplier) => (
                      <Badge key={s.guid_fixed} variant="outline" className="text-xs py-1">
                        {s.code} — {pickName(s.names, language)}
                      </Badge>
                    ))
                  )}
                </div>
              </div>
            </div>
          </Section>

          <Section title={text.mediaSectionTitle}>
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="size-16 overflow-hidden rounded-md border bg-muted flex items-center justify-center">
                  {productDetail.imageuri ? (
                    <img src={productDetail.imageuri} alt="Main" className="size-full object-cover" />
                  ) : (
                    <span className="text-[10px] text-muted-foreground">{text.noImage}</span>
                  )}
                </div>
                <div className="flex-1 space-y-1 text-xs">
                  <div><strong>{text.displayMode}:</strong> {productDetail.useimageorcolor ?? true ? text.displayModeImage : text.displayModeColor}</div>
                  {productDetail.colorselecthex && (
                    <div className="flex items-center gap-1.5">
                      <strong>{text.tagColor}:</strong>
                      <span className="inline-block size-3 rounded border" style={{ background: productDetail.colorselecthex }} />
                      <span>{productDetail.colorselect} ({productDetail.colorselecthex})</span>
                    </div>
                  )}
                </div>
              </div>

              {productDetail.images && productDetail.images.length > 0 && (
                <div className="border-t border-border/50 pt-2.5">
                  <span className="text-[11px] font-semibold text-muted-foreground block mb-1">{text.galleryLabel}</span>
                  <div className="flex flex-wrap gap-1.5">
                    {productDetail.images.map((img: ProductImage, idx: number) => (
                      <img key={idx} src={img.uri} alt="Gallery" className="size-10 rounded border object-cover" />
                    ))}
                  </div>
                </div>
              )}
            </div>
          </Section>
        </div>

        {/* Right Column: Restaurant & POS */}
        <div className="space-y-4">
          <Section title={text.restaurantSectionTitle}>
            <div className="space-y-3 text-xs">
              <div className="grid gap-2 grid-cols-2">
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.restaurant?.isforrestaurant ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isForRestaurantLabel}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.restaurant?.isfortakeaway ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isForTakeawayLabel}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.restaurant?.isfordelivery ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isForDeliveryLabel}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.restaurant?.isforcustomer ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isForCustomerLabel}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.restaurant?.isforcustomerpreorder ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isForCustomerPreOrderLabel}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.isalacarte ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isALaCarte}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.isstockforrestaurant ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isStockForRestaurantLabel}</span>
                </div>
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.issplitunitprint ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isSplitUnitPrintLabel}</span>
                </div>
              </div>

              <div className="border-t border-border/50 pt-2.5 grid gap-2 grid-cols-2">
                <ReadOnlyField label={text.foodTypeSectionLabel} value={productDetail.foodtype === 0 ? text.foodTypeFood : productDetail.foodtype === 1 ? text.foodTypeDrink : productDetail.foodtype === 2 ? text.foodTypeAlcohol : text.foodTypeOther} />
                <div className="flex items-center gap-2 p-1.5 rounded border bg-muted/20">
                  <input type="checkbox" checked={productDetail.isonlystaff ?? false} readOnly disabled className="pointer-events-none" />
                  <span>{text.isOnlyStaffLabel}</span>
                </div>
              </div>

              {productDetail.options && productDetail.options.length > 0 && (
                <div className="border-t border-border/50 pt-2.5">
                  <span className="text-[11px] font-semibold text-muted-foreground block mb-1">{text.optionsSectionLabel}</span>
                  <div className="space-y-1">
                    {productDetail.options.map((opt: ProductOption) => (
                      <div key={opt.guid} className="text-[11px] bg-muted/40 p-1.5 rounded border">
                        <span className="font-semibold">{pickName(opt.names, language)}</span>
                        <span className="text-muted-foreground"> ({opt.choicetype === 0 ? text.optionTypeMultiLabel : text.optionTypeSingleLabel}, {text.optionSelectRange.replace("%min", String(opt.minselect)).replace("%max", String(opt.maxselect))})</span>
                        <div className="mt-1 flex flex-wrap gap-1">
                          {opt.choices?.map((c: ProductChoice) => (
                            <Badge key={c.guid} variant="secondary" className="text-[10px] py-0.5">
                              {pickName(c.names, language)} {Number(c.price) > 0 ? `(+${c.price}฿)` : ""}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </Section>

          <Section title={text.timeSectionTitle}>
            <div className="space-y-2 text-xs">
              <Toggle checked={productDetail.isalert ?? false} onCheckedChange={() => {}} disabled label={text.alertToggleLabel} />
              {productDetail.isalert && productDetail.alertdescription && (
                <div className="bg-muted border border-border p-2 rounded text-xs text-foreground">
                  <strong>{text.alertMessageLabel}:</strong> {productDetail.alertdescription}
                </div>
              )}
              {productDetail.description && (
                <div className="border-t border-border/50 pt-2">
                  <span className="text-[11px] font-semibold text-muted-foreground block mb-0.5">{text.descriptionLabel}</span>
                  <p className="text-xs text-muted-foreground whitespace-pre-wrap">{productDetail.description}</p>
                </div>
              )}
            </div>
          </Section>
        </div>
      </div>
    </div>
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

// ─── Marketplace tab (Shopee / Lazada / AliExpress / TikTok) ───────────────
// One section per platform; all platforms share the same unified MarketplaceProductMap
// shape. A product holds at most one entry per platform (filtered by `platform`).

function TabMarketplace({
  platform,
  value,
  onChange,
  text,
}: {
  platform: MarketplacePlatform;
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: Text;
}) {
  const platformLabel =
    platform === "shopee"
      ? text.tabShopee
      : platform === "lazada"
        ? text.tabLazada
        : platform === "aliexpress"
          ? text.tabAliexpress
          : text.tabTiktok;

  const entry = (value.marketplace_products ?? []).find((m) => m.platform === platform) ?? null;

  const enable = useCallback(() => {
    onChange((current) => {
      const cur = current.marketplace_products ?? [];
      if (cur.some((m) => m.platform === platform)) return current;
      return { ...current, marketplace_products: [...cur, emptyMarketplaceProductMap(platform)] };
    });
  }, [onChange, platform]);

  const disable = useCallback(() => {
    onChange((current) => ({
      ...current,
      marketplace_products: (current.marketplace_products ?? []).filter((m) => m.platform !== platform),
    }));
  }, [onChange, platform]);

  const upd = useCallback(
    <K extends keyof MarketplaceProductMap>(key: K, val: MarketplaceProductMap[K]) => {
      onChange((current) => {
        const cur = current.marketplace_products ?? [];
        const idx = cur.findIndex((m) => m.platform === platform);
        if (idx < 0) return current;
        const next = [...cur];
        next[idx] = { ...next[idx], [key]: val };
        return { ...current, marketplace_products: next };
      });
    },
    [onChange, platform],
  );

  return (
    <div className="space-y-3">
      <Section title={platformLabel}>
        <Toggle
          checked={!!entry}
          onCheckedChange={(next) => (next ? enable() : disable())}
          label={text.mkEnableOnPlatform.replace("%s", platformLabel)}
        />
      </Section>

      {!entry ? (
        <div className="rounded-lg border border-dashed border-border bg-muted/30 p-6 text-center text-sm text-muted-foreground">
          {text.mkNotLinked.replace("%s", platformLabel)}
        </div>
      ) : (
        <>
          <Section title={text.mkSectionListing}>
            <FieldGrid>
              <FieldRow label={text.mkAccountId}>
                <Input value={entry.account_id} onChange={(e) => upd("account_id", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkMarketItemId}>
                <Input value={entry.market_item_id} onChange={(e) => upd("market_item_id", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkMarketModelId}>
                <Input value={entry.market_model_id} onChange={(e) => upd("market_model_id", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkSellerSku}>
                <Input value={entry.seller_sku} onChange={(e) => upd("seller_sku", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkShopSku}>
                <Input value={entry.shop_sku} onChange={(e) => upd("shop_sku", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkGtin}>
                <Input value={entry.gtin} onChange={(e) => upd("gtin", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkCategoryId}>
                <Input value={entry.category_id} onChange={(e) => upd("category_id", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkCategoryName}>
                <Input value={entry.category_name} onChange={(e) => upd("category_name", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkBrandId}>
                <Input value={entry.brand_id} onChange={(e) => upd("brand_id", e.target.value)} />
              </FieldRow>
              <FieldRow label={text.mkItemUrl}>
                <Input value={entry.item_url} onChange={(e) => upd("item_url", e.target.value)} placeholder="https://…" />
              </FieldRow>
            </FieldGrid>
            <div className="mt-3 space-y-3">
              <RadioOptionGroup
                label={text.mkStatus}
                value={entry.status}
                onChange={(next) => upd("status", next)}
                options={[
                  { value: "", label: text.mkStatusNone },
                  { value: "LIVE", label: text.mkStatusLive },
                  { value: "UNLIST", label: text.mkStatusUnlist },
                  { value: "REVIEWING", label: text.mkStatusReviewing },
                  { value: "REJECTED", label: text.mkStatusRejected },
                  { value: "DELETED", label: text.mkStatusDeleted },
                ]}
              />
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                <FieldRow label={text.mkDaysToShip}>
                  <NumberField value={entry.days_to_ship} min={0} step={1} onChange={(n) => upd("days_to_ship", n)} />
                </FieldRow>
                <div className="flex items-end pb-1">
                  <Toggle checked={entry.is_pre_order} onCheckedChange={(n) => upd("is_pre_order", n)} label={text.mkIsPreOrder} />
                </div>
              </div>
              {entry.status === "REJECTED" && (
                <FieldRow label={text.mkRejectReason}>
                  <Input value={entry.reject_reason} onChange={(e) => upd("reject_reason", e.target.value)} />
                </FieldRow>
              )}
            </div>
          </Section>

          <Section title={text.mkSectionMediaSpecs}>
            <div className="grid gap-3 lg:grid-cols-2">
              <MarketplaceJsonField
                label={text.mkMediaAssets}
                helper={text.mkJsonArrayHelp}
                invalidText={text.mkJsonInvalid}
                value={entry.media_assets}
                onCommit={(next) => upd("media_assets", next)}
              />
              <MarketplaceJsonField
                label={text.mkSpecificationGroups}
                helper={text.mkJsonArrayHelp}
                invalidText={text.mkJsonInvalid}
                value={entry.specification_groups}
                onCommit={(next) => upd("specification_groups", next)}
              />
              <MarketplaceJsonField
                label={text.mkRawAttributes}
                helper={text.mkJsonArrayHelp}
                invalidText={text.mkJsonInvalid}
                value={entry.raw_attributes}
                onCommit={(next) => upd("raw_attributes", next)}
              />
              <MarketplaceJsonField
                label={text.mkPayloadExamples}
                helper={text.mkJsonArrayHelp}
                invalidText={text.mkJsonInvalid}
                value={entry.payload_examples}
                onCommit={(next) => upd("payload_examples", next)}
              />
            </div>
          </Section>

          <Section title={text.mkSectionPriceStock}>
            <FieldGrid>
              <FieldRow label={text.mkCurrency}>
                <Input value={entry.currency} onChange={(e) => upd("currency", e.target.value.toUpperCase())} />
              </FieldRow>
              <FieldRow label={text.mkCustomPrice}>
                <NumberField value={entry.custom_price} min={0} onChange={(n) => upd("custom_price", n)} />
              </FieldRow>
              <FieldRow label={text.mkPlatformPrice}>
                <NumberField value={entry.platform_price} min={0} onChange={(n) => upd("platform_price", n)} />
              </FieldRow>
              <FieldRow label={text.mkPlatformStock}>
                <NumberField value={entry.platform_stock} min={0} step={1} onChange={(n) => upd("platform_stock", n)} />
              </FieldRow>
            </FieldGrid>
            <div className="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              <Toggle checked={entry.sync_stock} onCheckedChange={(n) => upd("sync_stock", n)} label={text.mkSyncStock} />
              <Toggle checked={entry.sync_price} onCheckedChange={(n) => upd("sync_price", n)} label={text.mkSyncPrice} />
            </div>
          </Section>

          <Section title={text.mkSectionSync}>
            <div className="space-y-3">
              <Toggle checked={entry.sync_enabled} onCheckedChange={(n) => upd("sync_enabled", n)} label={text.mkSyncEnabled} />
              <FieldGrid>
                <ReadOnlyField label={text.mkLastSyncAt} value={entry.last_sync_at || "—"} />
                <ReadOnlyField label={text.mkLastSyncError} value={entry.last_sync_error || "—"} />
              </FieldGrid>
            </div>
          </Section>
        </>
      )}
    </div>
  );
}

function MarketplaceJsonField({
  label,
  helper,
  invalidText,
  value,
  onCommit,
}: {
  label: string;
  helper: string;
  invalidText: string;
  value: unknown;
  onCommit: (value: any[]) => void;
}) {
  const [draft, setDraft] = useState(() => JSON.stringify(value ?? [], null, 2));
  const [invalid, setInvalid] = useState(false);

  useEffect(() => {
    if (invalid) return;
    setDraft(JSON.stringify(value ?? [], null, 2));
  }, [invalid, value]);

  return (
    <label className="grid gap-1 text-xs font-semibold">
      <span>{label}</span>
      <textarea
        className={cn(
          "min-h-32 w-full rounded-xl border bg-background px-3 py-2 font-mono text-[11px] text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring",
          invalid ? "border-destructive" : "border-input",
        )}
        value={draft}
        onChange={(event) => {
          setDraft(event.target.value);
          setInvalid(false);
        }}
        onBlur={() => {
          try {
            const parsed = JSON.parse(draft || "[]");
            onCommit(Array.isArray(parsed) ? parsed : [parsed]);
            setInvalid(false);
          } catch {
            setInvalid(true);
          }
        }}
        spellCheck={false}
      />
      <span className={cn("text-[10px]", invalid ? "text-destructive" : "text-muted-foreground")}>
        {invalid ? invalidText : helper}
      </span>
    </label>
  );
}

function TabLogistics({
  value,
  onChange,
  text,
}: {
  value: ProductBarcode;
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  text: any;
}) {
  const upd = useCallback(
    <K extends keyof ProductBarcode>(key: K, val: ProductBarcode[K]) =>
      onChange((current) => ({ ...current, [key]: val })),
    [onChange]
  );

  const volumetricWeight = useMemo(() => {
    const w = value.package_width ?? 0;
    const l = value.package_length ?? 0;
    const h = value.package_height ?? 0;
    return Number(((w * l * h) / 5000).toFixed(3));
  }, [value.package_width, value.package_length, value.package_height]);

  return (
    <div className="space-y-4">
      <Section title="ข้อมูลขนาดและน้ำหนักพัสดุ (Logistics Dimensions)">
        <div className="grid gap-4 sm:grid-cols-2">
          <FieldRow label="น้ำหนักพัสดุรวมกล่อง (kg)">
            <Input
              type="number"
              min={0}
              step="any"
              value={value.package_weight ?? 0}
              onChange={(e) => upd("package_weight", Math.max(0, Number(e.target.value) || 0))}
            />
          </FieldRow>
          <FieldRow label="น้ำหนักเชิงปริมาตรประเมิน (kg)">
            <Input
              type="text"
              value={`${volumetricWeight} kg`}
              readOnly
              className="bg-muted/40 font-mono"
            />
            <p className="text-[10px] text-muted-foreground mt-1">คำนวณจาก (กว้าง x ยาว x สูง) / 5000</p>
          </FieldRow>
        </div>
      </Section>

      <Section title="มิติกล่องพัสดุ (เซนติเมตร)">
        <div className="grid gap-4 sm:grid-cols-3">
          <FieldRow label="ความกว้างกล่อง (cm)">
            <Input
              type="number"
              min={0}
              value={value.package_width ?? 0}
              onChange={(e) => upd("package_width", Math.max(0, Number(e.target.value) || 0))}
            />
          </FieldRow>
          <FieldRow label="ความยาวกล่อง (cm)">
            <Input
              type="number"
              min={0}
              value={value.package_length ?? 0}
              onChange={(e) => upd("package_length", Math.max(0, Number(e.target.value) || 0))}
            />
          </FieldRow>
          <FieldRow label="ความสูงกล่อง (cm)">
            <Input
              type="number"
              min={0}
              value={value.package_height ?? 0}
              onChange={(e) => upd("package_height", Math.max(0, Number(e.target.value) || 0))}
            />
          </FieldRow>
        </div>
      </Section>

      <Section title="คุณลักษณะขนส่งพิเศษ (สติ๊กเกอร์ปะหน้า)">
        <div className="p-3 bg-muted/20 border rounded-lg space-y-3 text-xs">
          <p className="font-semibold text-muted-foreground">ธงสถานะสินค้าสำหรับเตรียมแพ็คและติดสติ๊กเกอร์จัดส่ง:</p>
          <div className="grid grid-cols-2 gap-3">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={value.isalert ?? false}
                onChange={(e) => upd("isalert", e.target.checked)}
                className="rounded accent-primary size-4"
              />
              <div>
                <span className="font-semibold block">สินค้าต้องระวังเป็นพิเศษ / แตกง่าย (Fragile)</span>
                <span className="text-[10px] text-muted-foreground">ติดสัญลักษณ์ระวังแตกบนใบปะหน้า</span>
              </div>
            </label>
          </div>
          {value.isalert && (
            <FieldRow label="คำเตือนเพิ่มเติมสำหรับพิมพ์ป้าย">
              <Input
                placeholder="เช่น ระวังแตกห้ามโยน / มีของเหลวซึมง่าย"
                value={value.alertdescription || ""}
                onChange={(e) => upd("alertdescription", e.target.value)}
              />
            </FieldRow>
          )}
        </div>
      </Section>
    </div>
  );
}

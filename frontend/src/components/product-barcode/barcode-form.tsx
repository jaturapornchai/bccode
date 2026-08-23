"use client";

/**
 * ProductBarcodeFormDialog — compact editor for a company-scoped barcode.
 *
 * - Pure controlled component: parent owns the `ProductBarcode` value via
 *   `value` + `onChange` props.
 * - `onSave` / `onCancel` for outer Save / Cancel actions.
 * - Validates: barcode regex, names[0] required, ItemUnit required.
 */

import {
  Info,
  Loader2,
  Plus,
  Save,
  Wand2,
  X,
} from "lucide-react";
import {
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
import { Ean13Barcode } from "@/components/product-barcode/ean13-barcode";
import { BusinessImageEditor } from "@/components/product-barcode/business-image-editor";
import { RadioOptionGroup } from "@/app/menu/product-tab-shared";
import { Textarea } from "@/components/ui/textarea";
import { listBarcodes, type MasterName, type MasterEntry } from "@/lib/product-barcode/api";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { type NameX, type ProductBarcode } from "@/lib/product-barcode/types";
import { ean13CheckDigit, isValidBarcode, pickName } from "@/lib/product-barcode/utils";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import type { AuthSession } from "@/lib/workspace-models";
import { NamesEditor } from "./names-editor";

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
  const [errors, setErrors] = useState<Record<string, string>>({});
  const shopLanguages = useMemo(
    () => (props.shopLanguages && props.shopLanguages.length > 0 ? props.shopLanguages : ["th", "en"]),
    [props.shopLanguages],
  );

  const validate = useCallback((): boolean => {
    const next: Record<string, string> = {};
    if (!value.barcode.trim()) next.barcode = text.required_error;
    else if (!isValidBarcode(value.barcode)) next.barcode = text.invalidBarcode;
    if (!value.itemcode.trim()) next.itemcode = text.required_error;
    const firstName = value.names[0]?.name ?? "";
    if (!firstName.trim()) next.name0 = text.required_error;
    if (!value.itemunitcode.trim()) next.itemunitcode = text.required_error;
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

        {/* Body */}
        <div className="min-h-0 flex-1 overflow-y-auto bg-background px-4 py-4">
          <QuickBarcodeFields
            auth={auth}
            companyGuid={companyGuid}
            errors={errors}
            language={language}
            mode={mode}
            onChange={onChange}
            shopLanguages={shopLanguages}
            text={text}
            value={value}
          />
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between gap-2 border-t border-border bg-card/95 px-4 py-3">
          <div className="flex gap-1">
            {Object.keys(errors).length > 0 ? (
              <Badge variant="warning">
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

function QuickBarcodeFields({
  auth,
  companyGuid,
  errors,
  language,
  mode,
  onChange,
  shopLanguages,
  text,
  value,
}: {
  auth: AuthSession | null;
  companyGuid?: string;
  errors: Record<string, string>;
  language: LanguageCode | string;
  mode: "create" | "edit";
  onChange: Dispatch<SetStateAction<ProductBarcode>>;
  shopLanguages: string[];
  text: Text;
  value: ProductBarcode;
}) {
  const isThai = normalizeLanguage(language) === "th";
  const [barcodePrefix, setBarcodePrefix] = useState<"200" | "885">("200");
  const [generatingBarcode, setGeneratingBarcode] = useState(false);
  const [generateError, setGenerateError] = useState("");
  const upd = useCallback(
    <K extends keyof ProductBarcode>(key: K, val: ProductBarcode[K]) =>
      onChange((current) => ({ ...current, [key]: val })),
    [onChange],
  );
  const updateBarcode = useCallback(
    (barcode: string) =>
      onChange((current) => {
        const shouldMirrorItemCode =
          mode === "create" &&
          (!current.itemcode.trim() || current.itemcode.trim() === current.barcode.trim());
        return {
          ...current,
          barcode,
          ...(shouldMirrorItemCode ? { itemcode: barcode } : {}),
        };
      }),
    [mode, onChange],
  );
  const generateBarcode = useCallback(async () => {
    if (!auth || !value.holdingcode || !value.businesscode) {
      setGenerateError(isThai ? "กรุณาเลือกบริษัทก่อนสร้างบาร์โค้ด" : "Select a company first.");
      return;
    }

    setGeneratingBarcode(true);
    setGenerateError("");
    try {
      for (let attempt = 0; attempt < 5; attempt += 1) {
        const randomDigits = Array.from(crypto.getRandomValues(new Uint8Array(9)), (digit) =>
          String(digit % 10),
        ).join("");
        const base = `${barcodePrefix}${randomDigits}`;
        const candidate = `${base}${ean13CheckDigit(base)}`;
        const result = await listBarcodes(auth, {
          holdingcode: value.holdingcode,
          businesscode: value.businesscode,
          keyword: candidate,
          limit: 5,
          offset: 0,
          sortfield: "barcode",
          sortorder: "asc",
        });
        if (!result.success) throw new Error(result.message || "Barcode lookup failed");
        if (!(result.data ?? []).some((item) => item.barcode === candidate)) {
          updateBarcode(candidate);
          return;
        }
      }
      setGenerateError(isThai ? "สร้างเลขไม่สำเร็จ กรุณาลองอีกครั้ง" : "Could not generate a unique barcode. Try again.");
    } catch (error) {
      setGenerateError(
        error instanceof Error && error.message
          ? error.message
          : isThai
            ? "ตรวจเลขซ้ำไม่สำเร็จ"
            : "Duplicate check failed.",
      );
    } finally {
      setGeneratingBarcode(false);
    }
  }, [auth, barcodePrefix, isThai, updateBarcode, value.businesscode, value.holdingcode]);

  return (
    <div className="space-y-3">
      <div className="flex items-start gap-3 rounded-lg border border-blue-200/60 bg-blue-50/70 p-3 text-sm text-blue-900 dark:border-blue-900/40 dark:bg-blue-950/20 dark:text-blue-100">
        <Info className="mt-0.5 h-4 w-4 shrink-0 text-blue-600 dark:text-blue-400" />
        <p>
          {isThai
            ? "สร้างบาร์โค้ดให้ขาย รับสินค้า และเริ่มงานสต๊อกได้ก่อน บาร์โค้ดใส่รูปและรายละเอียดของตัวเองได้ ส่วนราคา ต้นทุน และยอดคงเหลือยังจัดการที่สินค้า"
            : "Create the barcode first for sales, receiving, and stock operations. A barcode can keep its own images and description; price, cost, and balance remain on Product."}
        </p>
      </div>

      <Section title={isThai ? "ข้อมูลบาร์โค้ดที่จำเป็น" : "Required barcode data"}>
        <div className="space-y-4">
          <FieldGrid>
            <FieldRow label={text.barcode} hint={text.barcodeHelp} required>
              {mode === "create" ? (
                <RadioOptionGroup
                  label={isThai ? "คำนำหน้าบาร์โค้ดที่ระบบสร้าง" : "Generated barcode prefix"}
                  onChange={setBarcodePrefix}
                  options={[
                    { value: "200", label: isThai ? "200 — ใช้ภายในร้าน" : "200 — Store internal" },
                    { value: "885", label: "885 — GS1 Thailand" },
                  ]}
                  value={barcodePrefix}
                />
              ) : null}
              <div className="flex gap-2">
                <Input
                  value={value.barcode}
                  disabled={mode === "edit"}
                  onChange={(event) =>
                    updateBarcode(event.target.value.toUpperCase().replace(/[^A-Z0-9-]/g, ""))
                  }
                  aria-invalid={Boolean(errors.barcode) || undefined}
                  className={cn("flex-1", mode === "edit" && "bg-muted/40")}
                />
                {mode === "create" ? (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    aria-label={text.generateBarcode}
                    title={text.generateBarcode}
                    disabled={generatingBarcode}
                    onClick={() => void generateBarcode()}
                  >
                    {generatingBarcode ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wand2 className="h-4 w-4" />}
                  </Button>
                ) : null}
              </div>
              {mode === "create" && barcodePrefix === "885" ? (
                <p className="text-xs text-amber-700 dark:text-amber-300">
                  {isThai
                    ? "ใช้ 885 เฉพาะกิจการที่ได้รับเลขจาก GS1 Thailand"
                    : "Use 885 only with a number allocated by GS1 Thailand."}
                </p>
              ) : null}
              {generateError ? <p className="text-xs text-destructive">{generateError}</p> : null}
              <Ean13Barcode className="h-24 w-full max-w-sm rounded-md border border-border" value={value.barcode} />
              {mode === "edit" ? (
                <p className="text-xs text-muted-foreground">{text.barcodeLockedHint}</p>
              ) : null}
              {errors.barcode ? <p className="text-xs text-destructive">{errors.barcode}</p> : null}
            </FieldRow>

            <FieldRow
              label={text.itemCode}
              hint={isThai ? "รหัสสินค้าภายในบริษัทนี้" : "Product code within this company"}
              required
            >
              <Input
                value={value.itemcode}
                disabled={mode === "edit"}
                onChange={(event) => {
                  const itemcode = event.target.value.toUpperCase().replace(/\s/g, "");
                  onChange((current) => ({
                    ...current,
                    itemcode,
                    itemguid: itemcode === current.itemcode ? current.itemguid : "",
                  }));
                }}
                aria-invalid={Boolean(errors.itemcode) || undefined}
                className={cn(mode === "edit" && "bg-muted/40")}
              />
              {mode === "edit" ? (
                <p className="text-xs text-muted-foreground">
                  {isThai
                    ? "รหัสสินค้าเป็นตัวตนของบาร์โค้ด จึงเปลี่ยนไม่ได้หลังบันทึก"
                    : "Product code is part of the barcode identity and cannot be changed after saving."}
                </p>
              ) : null}
              {errors.itemcode ? <p className="text-xs text-destructive">{errors.itemcode}</p> : null}
            </FieldRow>

            <MasterField
              label={text.itemUnit}
              code={value.itemunitcode}
              names={value.itemunitnames}
              master="unit"
              language={language}
              auth={auth}
              required
              error={errors.itemunitcode}
              companyGuid={companyGuid}
              onPick={(entry) =>
                onChange((current) => ({
                  ...current,
                  itemunitguid: entry.guidfixed,
                  itemunitcode: entry.code,
                  itemunitnames: entry.names,
                }))
              }
              onClear={() =>
                onChange((current) => ({
                  ...current,
                  itemunitguid: "",
                  itemunitcode: "",
                  itemunitnames: [],
                }))
              }
            />
          </FieldGrid>

          <NamesEditor
            names={value.names}
            onChange={(next) => upd("names", next)}
            languages={shopLanguages}
            label={text.productName}
            firstRequired
            error={errors.name0}
            language={language}
          />

        </div>
      </Section>

      <Section title={isThai ? "รูป วิดีโอ และรายละเอียดบาร์โค้ด" : "Barcode images, videos, and description"}>
        <div className="space-y-3">
          <BusinessImageEditor
            auth={auth}
            galleryTitle={isThai ? "รูปเพิ่มเติมของบาร์โค้ด" : "Additional barcode images"}
            language={language}
            mainTitle={isThai ? "รูปหลักของบาร์โค้ด" : "Main barcode image"}
            value={value}
            onChange={(patch) =>
              onChange((current) => ({ ...current, ...patch }))
            }
          />
          <FieldRow
            label={isThai ? "รายละเอียดบาร์โค้ด" : "Barcode description"}
            hint={
              isThai
                ? "เช่น ลักษณะบรรจุภัณฑ์ สี รุ่น หรือข้อมูลที่ต่างจากสินค้าหลัก"
                : "For packaging, color, model, or details that differ from the product master."
            }
          >
            <Textarea
              maxLength={1500}
              onChange={(event) => upd("description", event.target.value)}
              placeholder={
                isThai
                  ? "รายละเอียดเฉพาะของบาร์โค้ดนี้..."
                  : "Details specific to this barcode..."
              }
              rows={4}
              value={value.description}
            />
            <p className="text-right text-xs text-muted-foreground">
              {value.description.length.toLocaleString(isThai ? "th-TH" : "en-US")} / 1,500
            </p>
          </FieldRow>
        </div>
      </Section>
    </div>
  );
}

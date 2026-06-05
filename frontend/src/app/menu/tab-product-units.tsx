"use client";

import { useCallback } from "react";
import { Plus, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import { type Product, type RefProductBarcode } from "@/lib/product-barcode/types";
import { cn } from "@/lib/utils";

type ProductStateAction = (
  value: Product | null | ((current: Product | null) => Product | null),
) => void;

function FieldRow({ label, hint, children, required }: { label: string; hint?: string; children: React.ReactNode; required?: boolean }) {
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

function FieldGrid({ children }: { children: React.ReactNode }) {
  return <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{children}</div>;
}

function Section({ title, action, children }: { title: string; action?: React.ReactNode; children: React.ReactNode }) {
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

function NumberField({
  value,
  onChange,
  min,
  step = "any",
  className,
  disabled,
}: {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  step?: string | number;
  className?: string;
  disabled?: boolean;
}) {
  return (
    <Input
      type="number"
      value={Number.isFinite(value) ? value : 0}
      step={step}
      min={min}
      disabled={disabled}
      onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
        const n = Number(event.target.value);
        onChange(Number.isFinite(n) ? n : 0);
      }}
      className={className}
    />
  );
}

export function TabProductUnits({
  value,
  onChange,
  lang,
  openPicker,
  clearPickerField,
}: {
  value: Product;
  onChange: ProductStateAction;
  lang: string;
  openPicker: (target: string, master: string, anchorEl?: HTMLElement | null) => void;
  clearPickerField: (field: string) => void;
}) {
  const textU = getBarcodeText(lang);
  const setRef = useCallback(
    (mutator: (rows: RefProductBarcode[]) => RefProductBarcode[]) =>
      onChange((c) => c ? ({ ...c, refbarcodes: mutator(c.refbarcodes ?? []) } as Product) : null),
    [onChange],
  );

  return (
    <div className="space-y-4">
      <Section title={textU.unitSection}>
        <FieldGrid>
          <FieldRow label={textU.unitBaseUnit}>
            <div className="flex gap-1.5">
              <Input
                readOnly
                value={
                  value.unitcode
                    ? `${value.unitcode} — ${pickName(value.unitnames, lang)}`
                    : ""
                }
                placeholder={textU.unitBasePlaceholder}
              />
              <Button type="button" variant="outline" onClick={(e) => openPicker("unit", "unit", e.currentTarget.parentElement)}>
                ...
              </Button>
              {value.unitcode && (
                <Button type="button" variant="ghost" onClick={() => clearPickerField("unit")}>
                  <X className="h-4 w-4" />
                </Button>
              )}
            </div>
          </FieldRow>
          <FieldRow label={textU.unitDivide}>
            <NumberField
              value={1}
              onChange={() => {}}
              disabled
              className="bg-muted opacity-60 cursor-not-allowed"
            />
          </FieldRow>
          <FieldRow label={textU.unitStand}>
            <NumberField
              value={1}
              onChange={() => {}}
              disabled
              className="bg-muted opacity-60 cursor-not-allowed"
            />
          </FieldRow>
        </FieldGrid>
      </Section>

      <Section title={textU.unitMultiSection}>
        <div className="space-y-3">
          <Toggle
            checked={value.isusesubbarcodes ?? false}
            onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isusesubbarcodes: n } as Product) : null)}
            label={textU.unitMultiToggle}
          />

          {value.isusesubbarcodes ? (
            <div className="border border-border/50 rounded-lg p-4 space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium">{textU.unitRefListTitle}</h4>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setRef((rows) => [
                      ...rows,
                      {
                        guidfixed: "",
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
                  {textU.unitRefAddBtn}
                </Button>
              </div>

              {(!value.refbarcodes || value.refbarcodes.length === 0) ? (
                <p className="text-sm text-muted-foreground text-center py-6">{textU.unitRefNoData}</p>
              ) : (
                <div className="space-y-2">
                  {value.refbarcodes.map((entry, idx) => (
                    <div
                      key={idx}
                      className="grid grid-cols-1 items-center gap-2 rounded-md border border-border p-2 md:grid-cols-[1.5fr_1.2fr_0.8fr_0.8fr_40px]"
                    >
                      {/* หน่วยนับ เป็นตัวหลัก */}
                      <FieldRow label={textU.unitRefUnitCode}>
                        <div className="flex gap-1.5">
                          <Input
                            readOnly
                            placeholder={textU.unitRefUnitPlaceholder}
                            value={entry.item_unit_code ? `${entry.item_unit_code} — ${pickName(entry.itemunitnames, lang)}` : ""}
                          />
                          <Button
                            type="button"
                            variant="outline"
                            onClick={(e) => openPicker("unit", "unit-ref-" + idx, e.currentTarget.parentElement)}
                          >
                            ...
                          </Button>
                        </div>
                      </FieldRow>

                      {/* บาร์โค้ด เป็นทางเลือก */}
                      <FieldRow label={textU.unitRefBarcode}>
                        <Input
                          placeholder={textU.unitRefBarcodePlaceholder}
                          value={entry.barcode || ""}
                          onChange={(event) =>
                            setRef((rows) =>
                              rows.map((row, rowIdx) =>
                                rowIdx === idx ? { ...row, barcode: event.target.value } : row,
                              ),
                            )
                          }
                        />
                      </FieldRow>


                      <FieldRow label={textU.unitRefDivide}>
                        <NumberField
                          value={entry.dividevalue ?? 1}
                          onChange={(n) =>
                            setRef((rows) =>
                              rows.map((row, rowIdx) =>
                                rowIdx === idx ? { ...row, dividevalue: n } : row,
                              ),
                            )
                          }
                          min={0.0001}
                        />
                      </FieldRow>
                      <FieldRow label={textU.unitRefStand}>
                        <NumberField
                          value={entry.standvalue ?? 1}
                          onChange={(n) =>
                            setRef((rows) =>
                              rows.map((row, rowIdx) =>
                                rowIdx === idx ? { ...row, standvalue: n } : row,
                              ),
                            )
                          }
                          min={0.0001}
                        />
                      </FieldRow>
                      <div className="flex justify-end pt-5">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => setRef((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : null}
        </div>
      </Section>
    </div>
  );
}

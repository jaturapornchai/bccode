"use client";

import { useCallback } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { NumericInput } from "@/components/ui/numeric-input";
import { normalizeBusinessCode } from "@/lib/business-code";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import {
  type Product,
  type ProductUnitConversion,
} from "@/lib/product-barcode/types";

type ProductStateAction = (
  value: Product | null | ((current: Product | null) => Product | null),
) => void;

function FieldRow({
  label,
  hint,
  children,
  required,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
  required?: boolean;
}) {
  return (
    <div className="flex flex-col gap-1 text-sm">
      <span className="font-medium text-foreground">
        {label}
        {required ? <span className="ml-1 text-destructive">*</span> : null}
      </span>
      {children}
      {hint ? (
        <span className="text-xs text-muted-foreground">{hint}</span>
      ) : null}
    </div>
  );
}

function FieldGrid({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{children}</div>
  );
}

function Section({
  title,
  action,
  children,
}: {
  title: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
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
    <NumericInput
      value={value}
      onChange={onChange}
      min={min}
      step={step}
      disabled={disabled}
      className={className}
    />
  );
}

function MatchedBarcodes({
  product,
  unitCode,
  label,
  emptyLabel,
}: {
  product: Product;
  unitCode: string;
  label: string;
  emptyLabel: string;
}) {
  const normalizedUnit = normalizeBusinessCode(unitCode);
  const matches = (product.barcodes ?? []).filter(
    (barcode) => normalizeBusinessCode(barcode.itemunitcode) === normalizedUnit,
  );
  return (
    <div
      className="rounded-md bg-muted/35 px-3 py-2"
      data-testid="matched-barcodes"
    >
      <div className="mb-1 text-xs font-medium text-muted-foreground">
        {label}
      </div>
      {matches.length > 0 ? (
        <div className="flex flex-wrap gap-1.5">
          {matches.map((barcode) => (
            <span
              key={barcode.guidfixed || barcode.barcode}
              className="rounded-md border border-border bg-background px-2 py-1 font-mono text-xs"
            >
              {barcode.barcode}
            </span>
          ))}
        </div>
      ) : (
        <p className="text-xs text-muted-foreground">{emptyLabel}</p>
      )}
    </div>
  );
}

export function TabProductUnits({
  value,
  onChange,
  lang,
  openPicker,
}: {
  value: Product;
  onChange: ProductStateAction;
  lang: string;
  openPicker: (
    target: string,
    master: string,
    anchorEl?: HTMLElement | null,
  ) => void;
}) {
  const textU = getBarcodeText(lang);
  const setUnits = useCallback(
    (mutator: (rows: ProductUnitConversion[]) => ProductUnitConversion[]) =>
      onChange((c) =>
        c
          ? ({
              ...c,
              unitconversions: mutator(c.unitconversions ?? []),
            } as Product)
          : null,
      ),
    [onChange],
  );

  return (
    <div className="space-y-4">
      <Section title={textU.unitSection}>
        <FieldGrid>
          <FieldRow label={textU.unitBaseUnit} required>
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
              <Button
                type="button"
                variant="outline"
                aria-label={textU.unitBaseUnit}
                onClick={(e) =>
                  openPicker("unit", "unit", e.currentTarget.parentElement)
                }
              >
                ...
              </Button>
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
        <div className="mt-3">
          <MatchedBarcodes
            product={value}
            unitCode={value.unitcode ?? ""}
            label={textU.unitRefBarcode}
            emptyLabel={textU.unitRefBarcodePlaceholder}
          />
        </div>
      </Section>

      <Section
        title={textU.unitMultiSection}
        action={
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() =>
              setUnits((rows) => [
                ...rows,
                { unitcode: "", unitnames: [], dividevalue: 1, standvalue: 1 },
              ])
            }
          >
            <Plus className="mr-1 h-4 w-4" />
            {textU.unitRefAddBtn}
          </Button>
        }
      >
        {!value.unitconversions || value.unitconversions.length === 0 ? (
          <p
            className="py-6 text-center text-sm text-muted-foreground"
            data-testid="product-unit-conversions"
          >
            {textU.unitRefNoData}
          </p>
        ) : (
          <div className="space-y-2" data-testid="product-unit-conversions">
            {value.unitconversions.map((entry, idx) => (
              <div
                key={`${entry.unitcode}-${idx}`}
                className="space-y-2 rounded-md border border-border p-3"
                data-testid="product-unit-conversion"
              >
                <div className="grid grid-cols-1 items-center gap-2 md:grid-cols-[1.5fr_0.8fr_0.8fr_40px]">
                  <FieldRow label={textU.unitRefUnitCode} required>
                    <div className="flex gap-1.5">
                      <Input
                        readOnly
                        placeholder={textU.unitRefUnitPlaceholder}
                        value={
                          entry.unitcode
                            ? `${entry.unitcode} — ${pickName(entry.unitnames, lang)}`
                            : ""
                        }
                      />
                      <Button
                        type="button"
                        variant="outline"
                        aria-label={textU.unitRefUnitCode}
                        onClick={(event) =>
                          openPicker(
                            "unit",
                            `unit-conversion-${idx}`,
                            event.currentTarget.parentElement,
                          )
                        }
                      >
                        ...
                      </Button>
                    </div>
                  </FieldRow>
                  <FieldRow label={textU.unitRefDivide}>
                    <NumberField
                      value={entry.dividevalue ?? 1}
                      onChange={(next) =>
                        setUnits((rows) =>
                          rows.map((row, rowIdx) =>
                            rowIdx === idx
                              ? { ...row, dividevalue: next }
                              : row,
                          ),
                        )
                      }
                      min={1}
                      step={1}
                    />
                  </FieldRow>
                  <FieldRow label={textU.unitRefStand}>
                    <NumberField
                      value={entry.standvalue ?? 1}
                      onChange={(next) =>
                        setUnits((rows) =>
                          rows.map((row, rowIdx) =>
                            rowIdx === idx ? { ...row, standvalue: next } : row,
                          ),
                        )
                      }
                      min={1}
                      step={1}
                    />
                  </FieldRow>
                  <div className="flex justify-end pt-5">
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      aria-label={textU.delete}
                      onClick={() =>
                        setUnits((rows) =>
                          rows.filter((_, rowIdx) => rowIdx !== idx),
                        )
                      }
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
                <MatchedBarcodes
                  product={value}
                  unitCode={entry.unitcode}
                  label={textU.unitRefBarcode}
                  emptyLabel={textU.unitRefBarcodePlaceholder}
                />
              </div>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}

"use client";

import { Input } from "@/components/ui/input";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { type Product } from "@/lib/product-barcode/types";
import { FieldGrid, FieldRow, Section, Toggle, type ProductStateAction } from "./product-tab-shared";

export function TabProductMisc({
  value,
  onChange,
  language = "th",
}: {
  value: Product;
  onChange: ProductStateAction;
  language?: string;
}) {
  const textMisc = getBarcodeText(language);
  return (
    <div className="space-y-4">
      <Section title={textMisc.miscAlertSection}>
        <div className="space-y-3">
          <Toggle
            checked={value.isalert ?? false}
            onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isalert: n } as Product) : null)}
            label={textMisc.miscAlertToggle}
          />
          {value.isalert ? (
            <FieldRow label={textMisc.miscAlertDetail}>
              <textarea
                className="min-h-[80px] w-full rounded-lg border border-input bg-background p-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                value={value.alertdescription || ""}
                onChange={(event) => onChange((c) => c ? ({ ...c, alertdescription: event.target.value } as Product) : null)}
                placeholder={textMisc.miscAlertPlaceholder}
              />
            </FieldRow>
          ) : null}
        </div>
      </Section>

      <Section title={textMisc.miscDescSection}>
        <textarea
          className="min-h-[120px] w-full rounded-lg border border-input bg-background p-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          value={value.description || ""}
          onChange={(event) => onChange((c) => c ? ({ ...c, description: event.target.value } as Product) : null)}
          placeholder={textMisc.miscDescPlaceholder}
        />
      </Section>

      <Section title={textMisc.miscGuidSection}>
        <FieldGrid>
          <FieldRow label={textMisc.miscGuidFixed}>
            <Input value={value.guidfixed || ""} readOnly className="bg-muted/50" />
          </FieldRow>
          <FieldRow label={textMisc.miscHoldingCode}>
            <Input value={value.holdingcode || ""} readOnly className="bg-muted/50" />
          </FieldRow>
          <FieldRow label={textMisc.miscUnitGuid}>
            <Input value={value.unitguid || ""} readOnly className="bg-muted/50" />
          </FieldRow>
        </FieldGrid>
      </Section>
    </div>
  );
}

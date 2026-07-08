"use client";

import { useCallback } from "react";
import { Input } from "@/components/ui/input";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { type Product } from "@/lib/product-barcode/types";
import { FieldGrid, FieldRow, NumberField, Section, type ProductStateAction } from "./product-tab-shared";

export function TabProductStock({
  value,
  onChange,
  text,
}: {
  value: Product;
  onChange: ProductStateAction;
  text: ReturnType<typeof getBarcodeText>;
}) {
  const upd = useCallback(
    <K extends keyof Product>(key: K, val: Product[K]) =>
      onChange((c) => c ? ({ ...c, [key]: val } as Product) : null),
    [onChange],
  );
  return (
    <Section title={text.tabStock}>
      <FieldGrid>
        <FieldRow label={text.orderPoint}>
          <NumberField value={value.orderpoint ?? 0} onChange={(n) => upd("orderpoint", n)} min={0} />
        </FieldRow>
        <FieldRow label={text.minPoint}>
          <NumberField value={value.minpoint ?? 0} onChange={(n) => upd("minpoint", n)} min={0} />
        </FieldRow>
        <FieldRow label={text.maxPoint}>
          <NumberField value={value.maxpoint ?? 0} onChange={(n) => upd("maxpoint", n)} min={0} />
        </FieldRow>
        <FieldRow label={text.qty}>
          <NumberField value={value.qty ?? 0} onChange={(n) => upd("qty", n)} />
        </FieldRow>
        <FieldRow label={text.stockBarcode}>
          <Input value={value.stockbarcode || ""} onChange={(event) => upd("stockbarcode", event.target.value)} />
        </FieldRow>
      </FieldGrid>
    </Section>
  );
}

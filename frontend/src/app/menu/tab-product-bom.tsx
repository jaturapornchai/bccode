"use client";

import { useCallback } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { type BOMProductBarcode, type Product } from "@/lib/product-barcode/types";
import { FieldRow, NumberField, Section, type ProductStateAction } from "./product-tab-shared";

export function TabProductBom({
  value,
  onChange,
  lang,
}: {
  value: Product;
  onChange: ProductStateAction;
  lang: string;
}) {
  const textB = getBarcodeText(lang);
  const setBom = useCallback(
    (mutator: (rows: BOMProductBarcode[]) => BOMProductBarcode[]) =>
      onChange((c) => c ? ({ ...c, bom: mutator(c.bom ?? []) } as Product) : null),
    [onChange],
  );

  return (
    <div className="space-y-4">
      <Section
        title={textB.bomSection}
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
                  itemunitcode: "",
                  itemunitnames: [],
                  barcode: "",
                  qty: 1,
                },
              ])
            }
          >
            <Plus className="mr-1 h-4 w-4" />
            {textB.bomAddBtn}
          </Button>
        }
      >
        {(!value.bom || value.bom.length === 0) ? (
          <p className="text-sm text-muted-foreground text-center py-8">{textB.bomNoData}</p>
        ) : (
          <div className="space-y-2">
            {value.bom.map((entry, idx) => (
              <div key={idx} className="grid grid-cols-1 items-center gap-2 md:grid-cols-[1.5fr_1fr_1fr_40px] rounded-md border border-border p-2">
                <FieldRow label={textB.bomBarcodeLabel}>
                  <Input
                    placeholder={textB.bomBarcodePlaceholder}
                    value={entry.barcode || ""}
                    onChange={(event) =>
                      setBom((rows) =>
                        rows.map((row, rowIdx) =>
                          rowIdx === idx ? { ...row, barcode: event.target.value } : row,
                        ),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textB.bomUnitLabel}>
                  <Input
                    placeholder={textB.bomUnitPlaceholder}
                    value={entry.itemunitcode || ""}
                    onChange={(event) =>
                      setBom((rows) =>
                        rows.map((row, rowIdx) =>
                          rowIdx === idx ? { ...row, itemunitcode: event.target.value } : row,
                        ),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textB.bomQtyLabel}>
                  <NumberField
                    value={entry.qty ?? 1}
                    onChange={(n) =>
                      setBom((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, qty: n } : row)),
                      )
                    }
                    min={0}
                  />
                </FieldRow>
                <div className="flex justify-end pt-5">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => setBom((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}

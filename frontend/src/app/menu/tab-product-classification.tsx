"use client";

import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import { type Product } from "@/lib/product-barcode/types";

export function TabProductClassification({
  value,
  text,
  lang,
  openPicker,
  clearPickerField,
}: {
  value: Product;
  text: ReturnType<typeof getBarcodeText>;
  lang: string;
  openPicker: (type: string, targetField: string, anchorEl?: HTMLElement | null) => void;
  clearPickerField: (field: string) => void;
}) {
  return (
    <div className="border border-border rounded-lg p-4 space-y-4">
      <h4 className="font-semibold text-sm border-b border-border pb-1">{text.groupsAndCategories}</h4>
      <div className="grid gap-4 sm:grid-cols-2">
        {[
          { key: "group", label: text.group, code: value.groupcode, names: value.groupnames },
          { key: "subgroup", label: text.groupsubone, code: value.subgroupcode, names: value.subgroupnames },
          { key: "brand", label: text.brand, code: value.brandcode, names: value.brandnames },
          { key: "category", label: text.category, code: value.categorycode, names: value.categorynames },
          { key: "class", label: text.class, code: value.classcode, names: value.classnames },
          { key: "design", label: text.design, code: value.designcode, names: value.designnames },
          { key: "model", label: text.model, code: value.modelcode, names: value.modelnames },
          { key: "pattern", label: text.pattern, code: value.patterncode, names: value.patternnames },
          { key: "grade", label: text.grade, code: value.gradecode, names: value.gradenames },
        ].map((field) => (
          <div key={field.key} className="space-y-1">
            <label className="text-xs text-muted-foreground">{field.label}</label>
            <div className="flex gap-1.5">
              <Input readOnly value={field.code ? `${field.code} — ${pickName(field.names, lang)}` : ""} />
              <Button type="button" variant="outline" aria-label={`Select ${field.label}`} onClick={(e) => openPicker(field.key, field.key, e.currentTarget.parentElement)}>...</Button>
              {field.code && (
                <Button type="button" variant="ghost" aria-label={`Clear ${field.label}`} onClick={() => clearPickerField(field.key)}>
                  <X className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

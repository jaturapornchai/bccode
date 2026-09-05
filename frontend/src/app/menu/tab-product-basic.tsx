"use client";

import { PlusCircle, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import { type Product } from "@/lib/product-barcode/types";
import { RadioOptionGroup } from "./product-tab-shared";

export function TabProductBasic({
  value,
  onChange,
  text,
  lang,
  activeLanguages,
  editorMode,
  itemTypes,
  materialTypes,
  vatTypes,
  openPicker,
  removeManufacturer,
  removeSupplier,
}: {
  value: Product;
  onChange: (next: Product) => void;
  text: ReturnType<typeof getBarcodeText>;
  lang: string;
  activeLanguages: string[];
  editorMode: "create" | "edit";
  itemTypes: Array<{ value: number; label: string }>;
  materialTypes: Array<{ value: number; label: string }>;
  vatTypes: Array<{ value: number; label: string }>;
  openPicker: (type: string, targetField: string, anchorEl?: HTMLElement | null) => void;
  removeManufacturer: (guid: string) => void;
  removeSupplier: (guid: string) => void;
}) {
  return (
    <div className="space-y-3">
      <div className="grid gap-3 md:grid-cols-1">
        <div className="space-y-1">
          <label className="text-xs font-semibold">{text.itemCode} *</label>
          <Input
            required
            disabled={editorMode === "edit"}
            value={value.code}
            onChange={(e) => onChange({ ...value, code: e.target.value.toUpperCase() })}
            className="h-9"
          />
        </div>
      </div>

      {/* Localized Names */}
      <div className="border border-border rounded-lg p-3">
        <NamesEditor
          names={value.names || []}
          onChange={(nextNames) => onChange({ ...value, names: nextNames })}
          languages={activeLanguages}
          label={text.productName}
          firstRequired
          language={lang}
        />
      </div>

      {/* ประเภทสินค้าหลัก (Radio Groups) */}
      <div className="rounded-lg border border-border bg-card p-3 shadow-sm">
        <h4 className="border-b border-border pb-1.5 text-xs sm:text-sm font-bold text-foreground">ประเภทสินค้าหลัก</h4>
        <div className="mt-2 grid items-start gap-2.5 md:grid-cols-2">
          <RadioOptionGroup
            label={text.itemTypeLabel}
            value={value.itemtype ?? 0}
            onChange={(n) => onChange({ ...value, itemtype: n })}
            options={itemTypes}
          />
          <RadioOptionGroup
            label={text.materialTypeLabel}
            value={value.materialtype ?? 0}
            onChange={(n) => onChange({ ...value, materialtype: n })}
            options={materialTypes}
          />
        </div>
      </div>

      {/* การตั้งค่าภาษี (Radio Groups) */}
      <div className="rounded-lg border border-border bg-card p-3 shadow-sm">
        <h4 className="border-b border-border pb-1.5 text-xs sm:text-sm font-bold text-foreground">การตั้งค่าภาษี</h4>
        <div className="mt-2 grid items-start gap-2.5">
          <RadioOptionGroup
            label="ประเภทภาษี"
            value={value.vattype ?? 0}
            onChange={(n) => onChange({ ...value, vattype: n })}
            options={vatTypes}
          />
        </div>
      </div>

      {/* Manufacturers & Suppliers section */}
      <div className="grid gap-3 md:grid-cols-2">
        {/* Manufacturers List (Multi Select) */}
        <div className="border border-border rounded-lg p-3 space-y-2">
          <div className="flex items-center justify-between border-b border-border pb-1">
            <h4 className="font-semibold text-sm">{text.manufacturers}</h4>
            <Button type="button" size="sm" variant="outline" onClick={(e) => openPicker("creditor", "manufacturers", e.currentTarget)}>
              <PlusCircle className="mr-1 h-3.5 w-3.5" />
              {text.addManufacturer}
            </Button>
          </div>
          <div className="space-y-2">
            {(!value.manufacturers || value.manufacturers.length === 0) ? (
              <p className="text-xs text-muted-foreground text-center py-2">{text.noManufacturerData}</p>
            ) : (
              value.manufacturers.map((m) => (
                <div key={m.guidfixed} className="flex items-center justify-between bg-muted/40 p-2 rounded text-sm border border-border/40">
                  <span className="truncate font-medium">{m.code} — {pickName(m.names, lang)}</span>
                  <Button type="button" size="sm" variant="ghost" className="text-destructive hover:bg-destructive/10" onClick={() => removeManufacturer(m.guidfixed)}>
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Suppliers List (Multi Select) */}
        <div className="border border-border rounded-lg p-4 space-y-3">
          <div className="flex items-center justify-between border-b border-border pb-1">
            <h4 className="font-semibold text-sm">{text.suppliers}</h4>
            <Button type="button" size="sm" variant="outline" onClick={(e) => openPicker("creditor", "suppliers", e.currentTarget)}>
              <PlusCircle className="mr-1 h-3.5 w-3.5" />
              {text.addSupplier}
            </Button>
          </div>
          <div className="space-y-2">
            {(!value.suppliers || value.suppliers.length === 0) ? (
              <p className="text-xs text-muted-foreground text-center py-2">{text.noSupplierData}</p>
            ) : (
              value.suppliers.map((s) => (
                <div key={s.guidfixed} className="flex items-center justify-between bg-muted/40 p-2 rounded text-sm border border-border/40">
                  <span className="truncate font-medium">{s.code} — {pickName(s.names, lang)}</span>
                  <Button type="button" size="sm" variant="ghost" className="text-destructive hover:bg-destructive/10" onClick={() => removeSupplier(s.guidfixed)}>
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

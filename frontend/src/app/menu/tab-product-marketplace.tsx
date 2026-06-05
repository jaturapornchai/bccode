"use client";

import { useCallback, useMemo, useState, useRef } from "react";
import { Plus, Trash2, X, AlertCircle, Upload, FileSpreadsheet, Check, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { pickName } from "@/lib/product-barcode/utils";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { cn } from "@/lib/utils";
import {
  type Product,
  type MarketplaceProductMap,
  type MarketplaceDimensionStock,
  type MarketplaceSKUMap,
  type RefProductBarcode,
  emptyMarketplaceProductMap,
  emptyMarketplaceSKUMap,
} from "@/lib/product-barcode/types";

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
  return <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">{children}</div>;
}

function Section({ title, action, children }: { title: string; action?: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="mb-4 overflow-hidden rounded-lg border border-border bg-card">
      <header className="flex items-center justify-between border-b border-border bg-muted/20 px-3 py-2">
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

export function TabProductMarketplace({
  platform,
  value,
  onChange,
  language,
}: {
  platform: "shopee" | "lazada" | "tiktok";
  value: Product;
  onChange: ProductStateAction;
  language: string;
}) {
  const textU = getBarcodeText(language);

  const [showImport, setShowImport] = useState(false);
  const [importText, setImportText] = useState("");
  const [importStatus, setImportStatus] = useState<{ success: number; failed: number; logs: string[] } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Platform specific configurations
  const platformName = useMemo(() => {
    switch (platform) {
      case "shopee": return "Shopee";
      case "lazada": return "Lazada";
      case "tiktok": return "TikTok Shop";
      default: return platform;
    }
  }, [platform]);

  const handleImportData = useCallback((textData: string) => {
    if (!textData.trim()) return;

    const lines = textData.split(/\r?\n/).map(l => l.trim()).filter(Boolean);
    if (lines.length < 2) {
      setImportStatus({
        success: 0,
        failed: 0,
        logs: ["Error: Data must have at least 2 rows (header + content)."],
      });
      return;
    }

    const firstLine = lines[0];
    const delimiter = firstLine.includes("\t") ? "\t" : (firstLine.includes(",") ? "," : ";");
    const headers = firstLine.split(delimiter).map(h => h.trim().toLowerCase());

    const holdingCodeIdx = headers.findIndex(h => h.includes("shop") || h.includes("ร้านค้า") || h.includes("บัญชี"));
    const marketItemIdIdx = headers.findIndex(h => h.includes("product_id") || h.includes("item_id") || h.includes("สินค้าบนเว็บ") || h.includes("market_item_id"));
    const sellerSkuIdx = headers.findIndex(h => h.includes("seller_sku") || h.includes("sku") || h.includes("รหัสคู่ค้า") || h.includes("sellersku"));
    const barcodeIdx = headers.findIndex(h => h.includes("barcode") || h.includes("บาร์โค้ด") || h.includes("รหัสบาร์โค้ด") || h.includes("บาร์โค๊ด"));
    const marketModelIdIdx = headers.findIndex(h => h.includes("variant_id") || h.includes("model_id") || h.includes("รหัสตัวเลือกย่อย") || h.includes("market_model_id"));

    if (sellerSkuIdx === -1 && barcodeIdx === -1) {
      setImportStatus({
        success: 0,
        failed: 0,
        logs: ["Error: Mapping column not found. Must contain at least 'Barcode' or 'Seller SKU' column."],
      });
      return;
    }

    let successCount = 0;
    let failedCount = 0;
    const logs: string[] = [];

    let nextProductMaps = [...(value.marketplace_products || [])];
    let nextRefBarcodes = [...(value.refbarcodes || [])];

    for (let i = 1; i < lines.length; i++) {
      const row = lines[i].split(delimiter).map(v => v.trim());
      if (row.length === 0 || (row.length === 1 && !row[0])) continue;

      const holdingCode = holdingCodeIdx !== -1 && row[holdingCodeIdx] ? row[holdingCodeIdx] : "shop_01";
      const marketItemId = marketItemIdIdx !== -1 && row[marketItemIdIdx] ? row[marketItemIdIdx] : "";
      const sellerSku = sellerSkuIdx !== -1 && row[sellerSkuIdx] ? row[sellerSkuIdx] : "";
      const barcodeValue = barcodeIdx !== -1 && row[barcodeIdx] ? row[barcodeIdx] : "";
      const marketModelId = marketModelIdIdx !== -1 && row[marketModelIdIdx] ? row[marketModelIdIdx] : "";

      let matchedBarcodeIdx = -1;
      if (barcodeValue) {
        matchedBarcodeIdx = nextRefBarcodes.findIndex(b => b.barcode === barcodeValue);
      }
      if (matchedBarcodeIdx === -1 && sellerSku) {
        matchedBarcodeIdx = nextRefBarcodes.findIndex(b => b.seller_sku === sellerSku);
      }

      if (matchedBarcodeIdx === -1) {
        failedCount++;
        logs.push(`Row ${i + 1}: No matching barcode/SKU for "${barcodeValue || sellerSku}" in system.`);
        continue;
      }

      const entry = nextRefBarcodes[matchedBarcodeIdx];

      if (holdingCode && marketItemId) {
        const hasShopMap = nextProductMaps.some(
          m => m.platform === platform && m.holdingcode === holdingCode && m.market_item_id === marketItemId
        );
        if (!hasShopMap) {
          nextProductMaps.push({
            ...emptyMarketplaceProductMap(platform),
            holdingcode: holdingCode,
            market_item_id: marketItemId,
            sync_status: "linked",
          });
        }
      }

      const mappings = entry.marketplace_sku_mappings || [];
      const matchMapIdx = mappings.findIndex(m => m.platform === platform && m.holdingcode === holdingCode);
      let nextMappings = [...mappings];

      const mappingData: MarketplaceSKUMap = {
        ...emptyMarketplaceSKUMap(platform, holdingCode, marketItemId),
        status: "LIVE",
        market_model_id: marketModelId,
        seller_sku: sellerSku || entry.seller_sku || "",
        last_sync_at: new Date().toISOString(),
      };

      if (matchMapIdx >= 0) {
        nextMappings[matchMapIdx] = { ...nextMappings[matchMapIdx], ...mappingData };
      } else {
        nextMappings.push(mappingData);
      }

      nextRefBarcodes[matchedBarcodeIdx] = {
        ...entry,
        seller_sku: sellerSku || entry.seller_sku,
        marketplace_sku_mappings: nextMappings,
      };

      successCount++;
      logs.push(`Row ${i + 1}: Mapped (Barcode: ${entry.barcode || "—"}, SKU: ${sellerSku || "—"} -> Shop: ${holdingCode})`);
    }

    onChange((c) => {
      if (!c) return null;
      return {
        ...c,
        marketplace_products: nextProductMaps,
        refbarcodes: nextRefBarcodes,
      } as Product;
    });

    setImportStatus({
      success: successCount,
      failed: failedCount,
      logs,
    });
  }, [platform, value.marketplace_products, value.refbarcodes, onChange]);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      setImportText(text);
      handleImportData(text);
    };
    reader.readAsText(file);
  };

  // Filter mappings for this platform
  const productMaps = useMemo(() => {
    return value.marketplace_products || [];
  }, [value.marketplace_products]);

  const platformProductMaps = useMemo(() => {
    return productMaps.map((item, idx) => ({ item, originalIdx: idx })).filter(x => x.item.platform === platform);
  }, [productMaps, platform]);

  const setProductMaps = useCallback(
    (mutator: (rows: MarketplaceProductMap[]) => MarketplaceProductMap[]) =>
      onChange((c) => c ? ({ ...c, marketplace_products: mutator(c.marketplace_products || []) } as Product) : null),
    [onChange],
  );

  const addShopMapping = useCallback(() => {
    setProductMaps((rows) => [
      ...rows,
      emptyMarketplaceProductMap(platform),
    ]);
  }, [platform, setProductMaps]);

  const removeShopMapping = useCallback(
    (originalIdx: number) => {
      setProductMaps((rows) => rows.filter((_, idx) => idx !== originalIdx));
    },
    [setProductMaps],
  );

  const updateShopMapping = useCallback(
    (originalIdx: number, fields: Partial<MarketplaceProductMap>) => {
      setProductMaps((rows) =>
        rows.map((row, idx) => (idx === originalIdx ? { ...row, ...fields } : row)),
      );
    },
    [setProductMaps],
  );

  // Helper to update refbarcodes list
  const setRefBarcodes = useCallback(
    (mutator: (rows: RefProductBarcode[]) => RefProductBarcode[]) =>
      onChange((c) => c ? ({ ...c, refbarcodes: mutator(c.refbarcodes || []) } as Product) : null),
    [onChange],
  );

  // Update seller sku for a variation
  const handleSellerSkuChange = useCallback((barcodeIdx: number, sku: string) => {
    setRefBarcodes((rows) =>
      rows.map((row, idx) => (idx === barcodeIdx ? { ...row, seller_sku: sku } : row)),
    );
  }, [setRefBarcodes]);

  // Update specific SKU mapping
  const handleSkuMapChange = useCallback((barcodeIdx: number, holdingCode: string, fields: Partial<MarketplaceSKUMap>) => {
    const market_item_id = value.marketplace_products?.find(
      (m) => m.platform === platform && m.holdingcode === holdingCode
    )?.market_item_id || "";

    setRefBarcodes((rows) =>
      rows.map((row, idx) => {
        if (idx !== barcodeIdx) return row;
        const mappings = row.marketplace_sku_mappings || [];
        const matchIdx = mappings.findIndex(m => m.platform === platform && m.holdingcode === holdingCode);
        let nextMappings = [...mappings];
        if (matchIdx >= 0) {
          nextMappings[matchIdx] = { ...nextMappings[matchIdx], ...fields };
        } else {
          nextMappings.push({
            ...emptyMarketplaceSKUMap(platform, holdingCode, market_item_id),
            seller_sku: row.seller_sku || "",
            ...fields,
          });
        }
        return { ...row, marketplace_sku_mappings: nextMappings };
      }),
    );
  }, [platform, setRefBarcodes]);

  const updateDimensionStocks = useCallback(
    (barcodeIdx: number, holdingCode: string, mutator: (rows: MarketplaceDimensionStock[]) => MarketplaceDimensionStock[]) => {
      const marketItemId = value.marketplace_products?.find(
        (m) => m.platform === platform && m.holdingcode === holdingCode,
      )?.market_item_id || "";

      handleSkuMapChange(barcodeIdx, holdingCode, {
        marketplace_dimension_stocks: mutator(
          value.refbarcodes?.[barcodeIdx]?.marketplace_sku_mappings?.find(
            (m) => m.platform === platform && m.holdingcode === holdingCode,
          )?.marketplace_dimension_stocks || [],
        ),
        market_item_id: marketItemId,
      });
    },
    [handleSkuMapChange, platform, value.marketplace_products, value.refbarcodes],
  );

  const addDimensionStock = useCallback(
    (barcodeIdx: number, holdingCode: string) => {
      updateDimensionStocks(barcodeIdx, holdingCode, (rows) => [
        ...rows,
        {
          dimension_key: "",
          dimension_name: "",
          market_dimension_id: "",
          available_qty: 0,
          reserved_qty: 0,
          inbound_qty: 0,
          oversell_buffer_qty: 0,
          last_platform_stock: 0,
          last_synced_at: "",
          last_sync_status: "",
          last_sync_error: "",
        },
      ]);
    },
    [updateDimensionStocks],
  );

  const updateDimensionStock = useCallback(
    (barcodeIdx: number, holdingCode: string, rowIdx: number, fields: Partial<MarketplaceDimensionStock>) => {
      updateDimensionStocks(barcodeIdx, holdingCode, (rows) =>
        rows.map((row, idx) => (idx === rowIdx ? { ...row, ...fields } : row)),
      );
    },
    [updateDimensionStocks],
  );

  const removeDimensionStock = useCallback(
    (barcodeIdx: number, holdingCode: string, rowIdx: number) => {
      updateDimensionStocks(barcodeIdx, holdingCode, (rows) => rows.filter((_, idx) => idx !== rowIdx));
    },
    [updateDimensionStocks],
  );

  // Update SKU dimensions (TikTok specific package size overrides)
  const handleSkuDimensionsChange = useCallback((barcodeIdx: number, fields: Partial<RefProductBarcode>) => {
    setRefBarcodes((rows) =>
      rows.map((row, idx) => (idx === barcodeIdx ? { ...row, ...fields } : row)),
    );
  }, [setRefBarcodes]);

  return (
    <div className="space-y-4">
      {/* 1. Dimensions & Logistics (Shared) */}
      <Section title={`ขนาดและน้ำหนักพัสดุสำหรับจัดส่ง (${platformName})`}>
        <div className="text-xs text-muted-foreground mb-3 bg-muted/40 p-2 rounded flex items-center gap-1.5 border border-border/40">
          <AlertCircle className="size-4 shrink-0 text-primary" />
          <span>ขนาดและน้ำหนักกล่องพัสดุที่แพ็คแล้ว ใช้เพื่อคำนวณราคาค่าส่งของแพลตฟอร์ม</span>
        </div>
        <FieldGrid>
          <FieldRow label="น้ำหนักรวมกล่อง (kg)">
            <NumberField
              value={value.package_weight ?? 0}
              onChange={(n) => onChange((c) => c ? ({ ...c, package_weight: n } as Product) : null)}
              min={0}
            />
          </FieldRow>
          <FieldRow label="ความยาวกล่อง (cm)">
            <NumberField
              value={value.package_length ?? 0}
              onChange={(n) => onChange((c) => c ? ({ ...c, package_length: n } as Product) : null)}
              min={0}
            />
          </FieldRow>
          <FieldRow label="ความกว้างกล่อง (cm)">
            <NumberField
              value={value.package_width ?? 0}
              onChange={(n) => onChange((c) => c ? ({ ...c, package_width: n } as Product) : null)}
              min={0}
            />
          </FieldRow>
          <FieldRow label="ความสูงกล่อง (cm)">
            <NumberField
              value={value.package_height ?? 0}
              onChange={(n) => onChange((c) => c ? ({ ...c, package_height: n } as Product) : null)}
              min={0}
            />
          </FieldRow>
        </FieldGrid>
      </Section>

      {/* 2. Product-level Shop Mapping */}
      <Section
        title={`การผูกรหัสสินค้าหลัก (${platformName} Shop Mappings)`}
        action={
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="text-primary border-primary/30 hover:bg-primary/5 h-8"
              onClick={() => setShowImport((prev) => !prev)}
            >
              <Upload className="mr-1 h-3.5 w-3.5" />
              นำเข้าด่วน
            </Button>
            <Button type="button" variant="outline" size="sm" className="h-8" onClick={addShopMapping}>
              <Plus className="mr-1 h-3.5 w-3.5" />
              เพิ่มร้านค้า
            </Button>
          </div>
        }
      >
        {showImport && (
          <div className="mb-4 rounded-md border border-primary/20 bg-primary/5 p-3 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-primary uppercase tracking-wider flex items-center gap-1.5">
                <FileSpreadsheet className="h-4.5 w-4.5" />
                นำเข้าข้อมูลเชื่อมโยงสินค้า (CSV หรือคัดลอกจาก Excel)
              </span>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-6 w-6 text-muted-foreground hover:text-foreground"
                onClick={() => {
                  setShowImport(false);
                  setImportStatus(null);
                }}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="text-xs space-y-1 text-muted-foreground">
              <p className="font-semibold text-foreground">💡 วิธีการใช้งาน:</p>
              <ul className="list-disc pl-4 space-y-0.5">
                <li>เตรียมคอลัมน์ใน Excel/CSV อย่างน้อย: <code className="bg-muted px-1 py-0.5 rounded font-mono">Barcode</code> หรือ <code className="bg-muted px-1 py-0.5 rounded font-mono">Seller SKU</code> เพื่อใช้ระบุสินค้า</li>
                <li>สามารถระบุข้อมูลที่จะจับคู่ด้วย: <code className="bg-muted px-1 py-0.5 rounded font-mono">Holding Code</code> (รหัสร้านค้า), <code className="bg-muted px-1 py-0.5 rounded font-mono">Market Product ID</code> (รหัสสินค้าบนเว็บ), และ <code className="bg-muted px-1 py-0.5 rounded font-mono">Market Variant ID</code> (รหัสย่อย)</li>
                <li>บันทึกเป็นไฟล์ CSV หรือคัดลอก (Copy) ตารางจาก Excel แล้ววางในช่องข้อความด้านล่างได้ทันที</li>
              </ul>
            </div>

            <div className="flex flex-col gap-2">
              <div className="flex flex-wrap items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="bg-background h-8"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <Upload className="mr-1 h-3.5 w-3.5" />
                  เลือกไฟล์ CSV / TXT
                </Button>
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileUpload}
                  accept=".csv,.txt"
                  className="hidden"
                />
                <span className="text-xs text-muted-foreground">หรือวางข้อมูลที่คัดลอกจาก Excel (ใช้ Tab แยกคอลัมน์) ในช่องด้านล่าง:</span>
              </div>

              <textarea
                placeholder="วางข้อมูลแถวที่นี่ (เช่น:&#10;Holding Code&#9;Barcode&#9;Seller SKU&#9;Market Product ID&#9;Market Variant ID&#10;shop_01&#9;8850123456789&#9;sku-red-01&#9;12345678&#9;98765432)"
                value={importText}
                onChange={(e) => setImportText(e.target.value)}
                className="h-24 w-full rounded border border-input bg-background p-2 font-mono text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />

              <div className="flex justify-end gap-2">
                {importText.trim() && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="text-muted-foreground h-8"
                    onClick={() => {
                      setImportText("");
                      setImportStatus(null);
                    }}
                  >
                    ล้างค่า
                  </Button>
                )}
                <Button
                  type="button"
                  size="sm"
                  className="h-8"
                  disabled={!importText.trim()}
                  onClick={() => handleImportData(importText)}
                >
                  <RefreshCw className="mr-1 h-3.5 w-3.5 animate-none" />
                  เริ่มประมวลผลข้อมูล
                </Button>
              </div>
            </div>

            {importStatus && (
              <div className="rounded border border-border bg-background p-2 space-y-1.5 text-xs">
                <div className="flex items-center gap-2 font-semibold">
                  <span className="text-green-600 flex items-center gap-0.5">
                    <Check className="h-3.5 w-3.5" /> สำเร็จ {importStatus.success} รายการ
                  </span>
                  {importStatus.failed > 0 && (
                    <span className="text-destructive flex items-center gap-0.5">
                      <AlertCircle className="h-3.5 w-3.5" /> ไม่พบสินค้า {importStatus.failed} รายการ
                    </span>
                  )}
                </div>
                <div className="max-h-24 overflow-y-auto space-y-0.5 font-mono text-[10px] text-muted-foreground border-t border-border pt-1">
                  {importStatus.logs.map((log, idx) => (
                    <div key={idx} className={log.startsWith("Error") || log.includes("No matching") ? "text-destructive/90" : "text-green-600/90"}>
                      {log}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {platformProductMaps.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">ยังไม่ได้ผูกกับร้านค้า {platformName}</p>
        ) : (
          <div className="space-y-3">
            {platformProductMaps.map(({ item, originalIdx }, mapIdx) => (
              <div
                key={originalIdx}
                className="grid grid-cols-1 items-end gap-3 rounded-md border border-border p-3 md:grid-cols-[1.5fr_1.5fr_1fr_40px]"
              >
                <FieldRow label={`รหัสร้านค้า (Holding Code) #${mapIdx + 1}`} required>
                  <Input
                    placeholder="เช่น shop_shopee_01"
                    value={item.holdingcode || ""}
                    onChange={(e) => updateShopMapping(originalIdx, { holdingcode: e.target.value })}
                  />
                </FieldRow>
                <FieldRow label={`รหัสสินค้าบนเว็บ (Marketplace Product ID)`} required>
                  <Input
                    placeholder="เช่น 2348910283"
                    value={item.market_item_id || ""}
                    onChange={(e) => updateShopMapping(originalIdx, { market_item_id: e.target.value })}
                  />
                </FieldRow>
                <FieldRow label="สถานะการซิงค์">
                  <Input value={item.sync_status || "unlinked"} readOnly className="bg-muted/50 text-xs font-semibold capitalize" />
                </FieldRow>
                <div className="flex justify-end pb-1">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="text-destructive hover:bg-destructive/10"
                    onClick={() => removeShopMapping(originalIdx)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* 3. SKU / Variant Mappings per Shop */}
      {platformProductMaps.length > 0 && (value.refbarcodes && value.refbarcodes.length > 0) && (
        <div className="space-y-4">
          <h4 className="text-sm font-semibold border-b border-border pb-1.5">
            ตั้งค่าตัวเลือกย่อย / บาร์โค้ดแยกตามร้านค้า (Variation Sync Settings)
          </h4>

          {platformProductMaps.map(({ item }) => {
            const holdingCode = item.holdingcode;
            if (!holdingCode) return null;

            return (
              <Section key={holdingCode} title={`ร้านค้า: ${holdingCode}`}>
                <div className="space-y-3">
                  {(value.refbarcodes || []).map((entry, barcodeIdx) => {
                    // Find mapping for this platform and holdingcode
                    const mapping = entry.marketplace_sku_mappings?.find(
                      (m) => m.platform === platform && m.holdingcode === holdingCode
                    ) || emptyMarketplaceSKUMap(platform, holdingCode, item.market_item_id);

                    return (
                      <div
                        key={barcodeIdx}
                        className="rounded-md border border-border p-3 bg-muted/10 space-y-3"
                      >
                        {/* Title Row */}
                        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border/50 pb-2">
                          <span className="text-sm font-medium text-primary">
                            {pickName(entry.itemunitnames, language)} ({entry.barcode || "ไม่มีบาร์โค้ด"})
                          </span>
                          <span className="text-xs text-muted-foreground">หน่วยนับ: {entry.item_unit_code || "—"}</span>
                        </div>

                        {/* Mappings Form */}
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                          <FieldRow label="รหัสคู่ค้า/SKU (Seller SKU)" required>
                            <Input
                              placeholder="ระบุรหัส SKU ตัวเลือก"
                              value={entry.seller_sku || ""}
                              onChange={(e) => handleSellerSkuChange(barcodeIdx, e.target.value)}
                            />
                          </FieldRow>
                          <FieldRow label="รหัสตัวเลือกย่อยเว็บ (Market Variant ID)">
                            <Input
                              placeholder="เช่น 55678912"
                              value={mapping.market_model_id || ""}
                              onChange={(e) => handleSkuMapChange(barcodeIdx, holdingCode, { market_model_id: e.target.value })}
                            />
                          </FieldRow>
                          <FieldRow label="ราคาขายเฉพาะช่องทาง">
                            <NumberField
                              value={mapping.custom_price ?? 0}
                              onChange={(n) => handleSkuMapChange(barcodeIdx, holdingCode, { custom_price: n })}
                              min={0}
                            />
                          </FieldRow>
                          <FieldRow label="ยอดคงเหลือรวมบน Marketplace">
                            <NumberField
                              value={mapping.platform_stock ?? 0}
                              onChange={(n) => handleSkuMapChange(barcodeIdx, holdingCode, { platform_stock: n })}
                              min={0}
                            />
                          </FieldRow>
                          <div className="flex flex-col justify-end gap-2 pb-2">
                            <Toggle
                              checked={mapping.sync_stock ?? true}
                              onCheckedChange={(n) => handleSkuMapChange(barcodeIdx, holdingCode, { sync_stock: n })}
                              label="ซิงค์จำนวนสต๊อกสินค้าหลัก"
                            />
                            <Toggle
                              checked={mapping.sync_price ?? true}
                              onCheckedChange={(n) => handleSkuMapChange(barcodeIdx, holdingCode, { sync_price: n })}
                              label="ซิงค์ราคาขายบนเว็บบอร์ด"
                            />
                          </div>
                        </div>

                        <div className="rounded-md border border-border/60 bg-background/60 p-2">
                          <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
                            <div>
                              <span className="block text-xs font-semibold text-foreground">ยอดคงเหลือตามมิติบน Marketplace</span>
                              <span className="text-[11px] text-muted-foreground">
                                ใช้เป็นยอดพร้อมขายแยกสี/ไซซ์/ตัวเลือกของแต่ละ marketplace ไม่ใช่ยอดบัญชีสต๊อกจริง
                              </span>
                            </div>
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              className="h-7 text-xs"
                              onClick={() => addDimensionStock(barcodeIdx, holdingCode)}
                            >
                              <Plus className="mr-1 h-3.5 w-3.5" />
                              เพิ่มมิติ
                            </Button>
                          </div>

                          {(!mapping.marketplace_dimension_stocks || mapping.marketplace_dimension_stocks.length === 0) ? (
                            <p className="rounded border border-dashed border-border px-2 py-2 text-center text-xs text-muted-foreground">
                              ยังไม่มียอดคงเหลือตามมิติ
                            </p>
                          ) : (
                            <div className="space-y-2">
                              {mapping.marketplace_dimension_stocks.map((dimensionStock, rowIdx) => (
                                <div
                                  key={`${dimensionStock.dimension_key || "dimension"}-${rowIdx}`}
                                  className="grid gap-2 rounded-md border border-border/60 bg-muted/10 p-2 lg:grid-cols-[minmax(110px,1fr)_minmax(120px,1fr)_minmax(120px,1fr)_repeat(4,minmax(86px,0.75fr))_36px]"
                                >
                                  <FieldRow label="รหัสมิติ">
                                    <Input
                                      value={dimensionStock.dimension_key}
                                      placeholder="color:red|size:m"
                                      onChange={(e) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { dimension_key: e.target.value })}
                                    />
                                  </FieldRow>
                                  <FieldRow label="ชื่อมิติ">
                                    <Input
                                      value={dimensionStock.dimension_name}
                                      placeholder="แดง / M"
                                      onChange={(e) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { dimension_name: e.target.value })}
                                    />
                                  </FieldRow>
                                  <FieldRow label="รหัสมิติบนเว็บ">
                                    <Input
                                      value={dimensionStock.market_dimension_id}
                                      placeholder="model id"
                                      onChange={(e) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { market_dimension_id: e.target.value })}
                                    />
                                  </FieldRow>
                                  <FieldRow label="พร้อมขาย">
                                    <NumberField
                                      value={dimensionStock.available_qty}
                                      onChange={(n) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { available_qty: n })}
                                      min={0}
                                    />
                                  </FieldRow>
                                  <FieldRow label="จอง">
                                    <NumberField
                                      value={dimensionStock.reserved_qty}
                                      onChange={(n) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { reserved_qty: n })}
                                      min={0}
                                    />
                                  </FieldRow>
                                  <FieldRow label="รับเข้า">
                                    <NumberField
                                      value={dimensionStock.inbound_qty}
                                      onChange={(n) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { inbound_qty: n })}
                                      min={0}
                                    />
                                  </FieldRow>
                                  <FieldRow label="กัน oversell">
                                    <NumberField
                                      value={dimensionStock.oversell_buffer_qty}
                                      onChange={(n) => updateDimensionStock(barcodeIdx, holdingCode, rowIdx, { oversell_buffer_qty: n })}
                                      min={0}
                                    />
                                  </FieldRow>
                                  <div className="flex items-end justify-end">
                                    <Button
                                      type="button"
                                      variant="ghost"
                                      size="icon"
                                      className="size-8 text-destructive hover:bg-destructive/10"
                                      onClick={() => removeDimensionStock(barcodeIdx, holdingCode, rowIdx)}
                                    >
                                      <Trash2 className="h-4 w-4" />
                                    </Button>
                                  </div>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>

                        {/* TikTok Shop variant logistics packages */}
                        {platform === "tiktok" && (
                          <div className="pt-2 border-t border-dashed border-border/50">
                            <span className="text-xs font-semibold text-muted-foreground block mb-2">ขนาดบรรจุภัณฑ์ย่อยเฉพาะตัวเลือกนี้ (TikTok SKU Dimensions)</span>
                            <div className="grid gap-2 grid-cols-2 lg:grid-cols-4">
                              <FieldRow label="น้ำหนัก SKU (kg)">
                                <NumberField
                                  value={entry.sku_package_weight ?? 0}
                                  onChange={(n) => handleSkuDimensionsChange(barcodeIdx, { sku_package_weight: n })}
                                  min={0}
                                />
                              </FieldRow>
                              <FieldRow label="ยาว SKU (cm)">
                                <NumberField
                                  value={entry.sku_package_length ?? 0}
                                  onChange={(n) => handleSkuDimensionsChange(barcodeIdx, { sku_package_length: n })}
                                  min={0}
                                />
                              </FieldRow>
                              <FieldRow label="กว้าง SKU (cm)">
                                <NumberField
                                  value={entry.sku_package_width ?? 0}
                                  onChange={(n) => handleSkuDimensionsChange(barcodeIdx, { sku_package_width: n })}
                                  min={0}
                                />
                              </FieldRow>
                              <FieldRow label="สูง SKU (cm)">
                                <NumberField
                                  value={entry.sku_package_height ?? 0}
                                  onChange={(n) => handleSkuDimensionsChange(barcodeIdx, { sku_package_height: n })}
                                  min={0}
                                />
                              </FieldRow>
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </Section>
            );
          })}
        </div>
      )}
    </div>
  );
}

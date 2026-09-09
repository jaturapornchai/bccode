"use client";

import { authFetch } from "@/lib/client-auth-session";
import React, { useState, useEffect, useLayoutEffect, useMemo } from "react";
import { createPortal } from "react-dom";
import {
  Loader2,
  Search,
  Plus,
  Trash2,
  Save,
  X,
  Info,
  Check,
  GripVertical,
  Edit3,
  Barcode,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { type AuthSession } from "@/lib/workspace-models";
import { cn } from "@/lib/utils";
import { pushNotice } from "@/lib/toast";

type SettingRecord = Record<string, unknown>;

interface NameX {
  code: string;
  name: string;
}

interface CategoryProduct {
  code: string;
  xorder: number;
  names: NameX[];
  itemcode?: string;
  unitname?: string;
  price?: number;
}

interface BarcodeSearchResult {
  barcode: string;
  names: NameX[];
  itemcode: string;
  itemunitcode: string;
  itemunitnames: NameX[];
  unitname: string;
  price: number;
}

function formatMoney(amount: number): string {
  return Number(amount || 0).toLocaleString("th-TH", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function normalizeProductCode(value: unknown): string {
  return String(value ?? "").trim().toUpperCase();
}

function categoryNamesFrom(value: unknown): NameX[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((item) => {
    if (!item || typeof item !== "object") return [];
    const raw = item as Record<string, unknown>;
    const code = String(raw.code ?? "").trim();
    const name = String(raw.name ?? "").trim();
    return code && name ? [{ code, name }] : [];
  });
}

function pickName(value: unknown, language: string): string {
  const names = categoryNamesFrom(value);
  const languageCode = language.trim().toLowerCase();
  return (
    names.find((item) => item.code.toLowerCase() === languageCode)?.name ??
    names.find((item) => item.code.toLowerCase() === "th")?.name ??
    names.find((item) => item.code.toLowerCase() === "en")?.name ??
    names[0]?.name ??
    ""
  );
}

function categoryProductsFrom(value: unknown): CategoryProduct[] {
  if (!Array.isArray(value)) return [];

  const seen = new Set<string>();
  const products: CategoryProduct[] = [];
  for (const item of value) {
    if (!item || typeof item !== "object") continue;
    const raw = item as Record<string, unknown>;
    const code = normalizeProductCode(raw.code);
    if (!code || seen.has(code)) continue;
    seen.add(code);
    products.push({
      code,
      xorder: products.length,
      names: categoryNamesFrom(raw.names),
    });
  }
  return products;
}

export interface ProductCategoryItemsEditorProps {
  auth: AuthSession | null;
  workspace: { shop: { holdingcode: string } } | null;
  language: string;
  categorySelectedGuid: string;
  categoryRecord: SettingRecord | null;
  onRefresh?: () => void;
  setEditing: (record: SettingRecord | null) => void;
  saving: boolean;
  setSaving: (saving: boolean) => void;
  onUnsavedChangesChange?: (hasChanges: boolean) => void;
  onOpenEdit?: () => void;
  configSlug?: string;
  className?: string;
  triggerSearchNonce?: number;
}

export function ProductCategoryItemsEditor({
  auth,
  workspace,
  language,
  categorySelectedGuid,
  categoryRecord,
  onRefresh,
  setEditing,
  saving,
  setSaving,
  onUnsavedChangesChange,
  onOpenEdit,
  configSlug,
  className,
  triggerSearchNonce,
}: ProductCategoryItemsEditorProps) {
  const [localCodelist, setLocalCodelist] = useState<CategoryProduct[]>([]);
  const [searchDialogOpen, setSearchDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [searchResults, setSearchResults] = useState<BarcodeSearchResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState("");

  useEffect(() => {
    if (triggerSearchNonce && triggerSearchNonce > 0) {
      setSearchQuery("");
      setSearchDialogOpen(true);
    }
  }, [triggerSearchNonce]);
  const setNotice = pushNotice;
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  // Compare current localCodelist with original categoryRecord.codelist
  const hasChanges = useMemo(() => {
    if (!categoryRecord) return false;
    const originalList = categoryProductsFrom(categoryRecord.codelist);
    if (originalList.length !== localCodelist.length) return true;
    return localCodelist.some((item, idx) => {
      const orig = originalList[idx];
      return !orig || item.code !== orig.code;
    });
  }, [categoryRecord, localCodelist]);

  // Notify parent of unsaved changes
  useEffect(() => {
    onUnsavedChangesChange?.(hasChanges);
  }, [hasChanges, onUnsavedChangesChange]);

  // Alert before unloading tab/page
  useEffect(() => {
    if (!hasChanges) return;
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [hasChanges]);

  // Sync categoryRecord codelist when selected category changes
  useLayoutEffect(() => {
    if (categoryRecord) {
      setLocalCodelist(categoryProductsFrom(categoryRecord.codelist));
      setNotice(null);
    } else {
      setLocalCodelist([]);
    }
  }, [categoryRecord]);

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Query Barcode list from backend through /api/product-barcode/list.
  useEffect(() => {
    if (!searchDialogOpen || !auth || !workspace) return;

    const authSession = auth;
    const workspaceSession = workspace;
    let active = true;
    async function fetchBarcodes() {
      setSearching(true);
      setSearchError("");
      try {
        const response = await authFetch("/api/product-barcode/list", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${authSession.token}`,
            "x-bc-backend-url": authSession.backendUrl,
            "Accept-Language": language,
          },
          body: JSON.stringify({
            holdingcode: workspaceSession.shop.holdingcode,
            keyword: debouncedQuery,
            limit: 50,
            backendUrl: authSession.backendUrl,
          }),
          cache: "no-store",
        });
        const payload = await response.json();
        if (!active) return;
        if (response.ok && payload.success !== false) {
          const rawList = Array.isArray(payload.data) ? payload.data : [];
          const barcodes: BarcodeSearchResult[] = rawList
            .map((raw: any): BarcodeSearchResult | null => {
              if (!raw || typeof raw !== "object") return null;
              const barcode = normalizeProductCode(raw.barcode);
              if (!barcode) return null;
              const unitNames = categoryNamesFrom(raw.itemunitnames);
              const unitName = pickName(unitNames, language) || String(raw.itemunitcode || "");
              const price =
                typeof raw.price === "number"
                  ? raw.price
                  : Array.isArray(raw.prices) && raw.prices[0]
                    ? Number(raw.prices[0].price || 0)
                    : 0;
              return {
                barcode,
                names: categoryNamesFrom(raw.names),
                itemcode: String(raw.itemcode || "").trim(),
                itemunitcode: String(raw.itemunitcode || "").trim(),
                itemunitnames: unitNames,
                unitname: unitName,
                price,
              };
            })
            .filter((item: BarcodeSearchResult | null): item is BarcodeSearchResult => item !== null);
          setSearchResults(barcodes);
        } else {
          setSearchError(payload.message || (language === "th" ? "โหลดบาร์โค้ดไม่สำเร็จ" : "Failed to load barcodes"));
        }
      } catch (err: any) {
        if (!active) return;
        setSearchError(err.message || (language === "th" ? "เกิดข้อผิดพลาดในการเชื่อมต่อ" : "Network error"));
      } finally {
        if (active) setSearching(false);
      }
    }

    void fetchBarcodes();
    return () => {
      active = false;
    };
  }, [debouncedQuery, searchDialogOpen, auth, workspace, language]);

  const categoryRecordGuid = String(
    categoryRecord?.guidfixed ?? categoryRecord?.guid ?? "",
  );
  const categoryPending = Boolean(
    categorySelectedGuid && categoryRecordGuid !== categorySelectedGuid,
  );

  if (!categoryRecord) {
    return (
      <Card className={cn("flex h-full min-h-0 flex-col items-center justify-center border-border bg-card p-6 shadow-sm", className)}>
        <Loader2 className="size-6 animate-spin text-primary" />
        <span className="mt-2 text-xs text-muted-foreground">
          {language === "th" ? "กำลังโหลดข้อมูลบาร์โค้ดในหมวด..." : "Loading barcodes in category..."}
        </span>
      </Card>
    );
  }

  const categoryName = pickName(categoryRecord.names, language) || String(categoryRecord.code || "");

  // Add barcode to local codelist
  const handleAddItem = (item: BarcodeSearchResult) => {
    const code = normalizeProductCode(item.barcode);
    const isDuplicate = localCodelist.some((x) => x.code === code);
    if (isDuplicate) return;

    const newItem: CategoryProduct = {
      code,
      xorder: localCodelist.length,
      names: item.names,
      itemcode: item.itemcode,
      unitname: item.unitname,
      price: item.price,
    };

    setLocalCodelist((prev) => [...prev, newItem]);
  };

  // Remove item from local codelist
  const handleRemoveItem = (code: string) => {
    setLocalCodelist((prev) =>
      prev
        .filter((item) => item.code !== code)
        .map((item, index) => ({ ...item, xorder: index })),
    );
  };

  // Drag & Drop Handlers
  const handleDragStart = (e: React.DragEvent, index: number) => {
    e.dataTransfer.effectAllowed = "move";
    setDraggedIndex(index);
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === index) return;
    setDragOverIndex(index);
  };

  const handleDrop = (e: React.DragEvent, targetIndex: number) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === targetIndex) return;

    const list = [...localCodelist];
    const draggedItem = list[draggedIndex];
    list.splice(draggedIndex, 1);
    list.splice(targetIndex, 0, draggedItem);
    setLocalCodelist(list.map((item, index) => ({ ...item, xorder: index })));
    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  const handleDragEnd = () => {
    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  // Save changes to backend
  const handleSave = async () => {
    if (!auth || !workspace) return;
    setSaving(true);
    setNotice(null);

    const guid = String(categoryRecord.guidfixed || categoryRecord.guid || "");
    try {
      const codelist = localCodelist.map((item, index) => ({
        code: item.code,
        xorder: index,
        names: item.names,
      }));
      const payload = {
        ...categoryRecord,
        codelist,
      };

      const endpointSlug = configSlug || "productcategorygroupselectscreen";
      const response = await authFetch(
        `/api/system-settings/${endpointSlug}/${encodeURIComponent(guid)}?holdingcode=${encodeURIComponent(workspace.shop.holdingcode)}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
          body: JSON.stringify({
            ...payload,
            backendUrl: auth.backendUrl,
            holdingcode: workspace.shop.holdingcode,
          }),
        }
      );

      const data = await response.json();
      if (!response.ok || data.success === false) {
        throw new Error(data.message || "Failed to save category items");
      }

      setNotice({
        type: "success",
        text: language === "th" ? "บันทึกข้อมูลเรียบร้อยแล้ว" : "Saved successfully",
      });

      // Update parent state
      setLocalCodelist(codelist);
      setEditing(payload);
      onRefresh?.();
    } catch (err: any) {
      setNotice({
        type: "error",
        text: err.message || (language === "th" ? "เกิดข้อผิดพลาดในการบันทึก" : "Save failed"),
      });
    } finally {
      setSaving(false);
    }
  };

  // Render barcode search modal dialog
  const renderSearchDialog = () => {
    if (!searchDialogOpen) return null;

    return createPortal(
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
        onClick={() => setSearchDialogOpen(false)}
      >
        <div
          role="dialog"
          aria-modal="true"
          aria-label={language === "th" ? "ค้นหาและเพิ่มบาร์โค้ด" : "Search and add barcodes"}
          className="flex h-[80vh] w-full max-w-2xl flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          <header className="flex items-center justify-between border-b border-border px-4 py-3">
            <div className="flex items-center gap-2">
              <Barcode className="size-5 text-primary" />
              <span className="text-base font-bold">
                {language === "th" ? "ค้นหาและเพิ่มบาร์โค้ด" : "Search and Add Barcodes"}
              </span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="size-8 rounded-full"
              onClick={() => setSearchDialogOpen(false)}
            >
              <X className="size-4" />
            </Button>
          </header>

          <div className="border-b border-border p-4 bg-muted/20">
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                autoFocus
                placeholder={
                  language === "th"
                    ? "ค้นหาด้วยบาร์โค้ด, รหัสสินค้า, หรือชื่อสินค้า..."
                    : "Search by barcode, product code, or name..."
                }
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-10 !pl-10 rounded-lg border-input bg-background"
              />
            </div>
          </div>

          <div className="flex-1 overflow-y-auto p-2">
            {searching ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground">
                <Loader2 className="size-6 animate-spin text-primary" />
                <span>{language === "th" ? "กำลังค้นหาบาร์โค้ด..." : "Searching barcodes..."}</span>
              </div>
            ) : searchError ? (
              <div className="p-4 text-center text-sm text-destructive font-medium">
                {searchError}
              </div>
            ) : searchResults.length === 0 ? (
              <div className="py-12 text-center text-sm text-muted-foreground">
                {language === "th" ? "ไม่พบบาร์โค้ดที่ตรงตามคำค้นหา" : "No barcodes found"}
              </div>
            ) : (
              <div className="divide-y divide-border/60">
                {searchResults.map((item, idx) => {
                  const code = normalizeProductCode(item.barcode);
                  const isAdded = localCodelist.some((x) => x.code === code);
                  const pName = pickName(item.names, language) || code;
                  return (
                    <div
                      key={code || idx}
                      className="flex items-center justify-between gap-4 p-3 hover:bg-muted/40 rounded-lg transition-colors"
                    >
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs font-bold text-foreground">
                            <Barcode className="size-3 text-muted-foreground" />
                            {code}
                          </span>
                          <span className="font-semibold text-sm text-foreground truncate">
                            {pName}
                          </span>
                        </div>
                        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1 text-xs text-muted-foreground font-medium">
                          {item.itemcode ? (
                            <span>{language === "th" ? "รหัสสินค้า" : "Code"}: {item.itemcode}</span>
                          ) : null}
                          {item.unitname ? (
                            <span>{language === "th" ? "หน่วย" : "Unit"}: {item.unitname}</span>
                          ) : null}
                          {item.price > 0 ? (
                            <span className="font-semibold text-primary">
                              {language === "th" ? "ราคา" : "Price"}: {formatMoney(item.price)}
                            </span>
                          ) : null}
                        </div>
                      </div>
                      <Button
                        type="button"
                        size="sm"
                        variant={isAdded ? "ghost" : "outline"}
                        disabled={isAdded}
                        className={cn(
                          "h-8 rounded-lg font-semibold shrink-0",
                          isAdded
                            ? "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/20"
                            : "hover:border-primary/50 hover:bg-primary/5 hover:text-primary"
                        )}
                        onClick={() => handleAddItem(item)}
                      >
                        {isAdded ? (
                          <>
                            <Check className="size-3.5 mr-1" />
                            {language === "th" ? "เพิ่มแล้ว" : "Added"}
                          </>
                        ) : (
                          <>
                            <Plus className="size-3.5 mr-1" />
                            {language === "th" ? "เพิ่ม" : "Add"}
                          </>
                        )}
                      </Button>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <footer className="border-t border-border px-4 py-3 bg-muted/10 text-right">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setSearchDialogOpen(false)}
              className="rounded-lg font-semibold"
            >
              {language === "th" ? "ปิด" : "Close"}
            </Button>
          </footer>
        </div>
      </div>,
      document.body
    );
  };

  return (
    <Card
      aria-busy={categoryPending}
      className={cn("flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm", className)}
      inert={categoryPending ? true : undefined}
    >
      <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-2 border-b border-border bg-muted/20 px-4 py-3 shrink-0">
        <div className="grid gap-0.5">
          <CardTitle className="text-sm font-bold text-foreground">
            {language === "th" ? "บาร์โค้ดในหมวด" : "Barcodes in Category"}
          </CardTitle>
          <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5 text-xs font-semibold">
            <span className="text-primary">{categoryName}</span>
            <span className="text-muted-foreground">
              {language === "th" ? `ทั้งหมด ${localCodelist.length} รายการ` : `Total ${localCodelist.length} items`}
            </span>
          </div>
        </div>
        <div className="flex w-full flex-wrap items-center justify-end gap-2 sm:w-auto" data-testid="product-category-item-actions">
          {onOpenEdit ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="h-8 rounded-lg gap-1 font-semibold"
              onClick={onOpenEdit}
            >
              <Edit3 className="size-3.5" />
              {language === "th" ? "แก้ไขข้อมูลหมวด" : "Edit Category"}
            </Button>
          ) : null}
          <Button
            type="button"
            size="sm"
            className="h-8 rounded-lg gap-1 font-semibold"
            onClick={() => {
              setSearchQuery("");
              setSearchDialogOpen(true);
            }}
          >
            <Plus className="size-4" />
            {language === "th" ? "เพิ่มบาร์โค้ด" : "Add Barcode"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={saving}
            onClick={() => {
              if (categoryRecord) {
                setLocalCodelist(categoryProductsFrom(categoryRecord.codelist));
              }
              setNotice(null);
            }}
            className="h-8 rounded-lg font-semibold"
          >
            {language === "th" ? "ยกเลิก" : "Cancel"}
          </Button>
          <Button
            type="button"
            size="sm"
            disabled={saving}
            onClick={handleSave}
            className="h-8 rounded-lg gap-1.5 font-semibold"
          >
            {saving ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Save className="size-4" />
            )}
            {saving
              ? (language === "th" ? "กำลังบันทึก..." : "Saving...")
              : (language === "th" ? "บันทึก" : "Save")}
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex min-h-0 flex-1 flex-col p-0">
        <div className="flex-1 overflow-y-auto">
          {localCodelist.length === 0 ? (
            <div className="flex h-60 flex-col items-center justify-center gap-2 p-4 text-center text-sm text-muted-foreground">
              <Info className="size-8 text-muted-foreground/60" />
              <span className="font-semibold text-foreground">
                {language === "th" ? "ยังไม่มีบาร์โค้ดในหมวดนี้" : "No barcodes in this category"}
              </span>
              <span className="text-xs text-muted-foreground max-w-xs leading-normal">
                {language === "th"
                  ? "กดปุ่มด้านล่างเพื่อค้นหาและเพิ่มบาร์โค้ดเข้าหมวดนี้"
                  : "Click the button below to search and add barcodes to this category."}
              </span>
              <Button
                type="button"
                size="sm"
                className="mt-1 h-8 rounded-lg gap-1.5 font-semibold"
                onClick={() => {
                  setSearchQuery("");
                  setSearchDialogOpen(true);
                }}
              >
                <Plus className="size-4" />
                {language === "th" ? "เพิ่มบาร์โค้ด" : "Add Barcode"}
              </Button>
            </div>
          ) : (
            <div className="w-full overflow-x-auto">
              <table className="w-full text-left text-sm border-collapse">
                <thead>
                  <tr className="border-b border-border bg-muted/30 text-xs font-bold text-muted-foreground uppercase tracking-wider select-none">
                    <th className="w-8 px-2 py-2.5"></th>
                    <th className="px-4 py-2.5 font-bold">{language === "th" ? "บาร์โค้ด" : "Barcode"}</th>
                    <th className="px-4 py-2.5 font-bold">{language === "th" ? "ชื่อสินค้า/บาร์โค้ด" : "Name"}</th>
                    <th className="px-4 py-2.5 font-bold text-center w-24">{language === "th" ? "หน่วยนับ" : "Unit"}</th>
                    <th className="px-4 py-2.5 font-bold text-right w-28">{language === "th" ? "ราคา" : "Price"}</th>
                    <th className="px-4 py-2.5 text-center font-bold w-14">{language === "th" ? "จัดการ" : "Action"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {localCodelist.map((item, idx) => {
                    const pName = pickName(item.names, language) || item.code;
                    return (
                      <tr
                        key={item.code}
                        draggable
                        onDragStart={(e) => handleDragStart(e, idx)}
                        onDragOver={(e) => handleDragOver(e, idx)}
                        onDrop={(e) => handleDrop(e, idx)}
                        onDragEnd={handleDragEnd}
                        className={cn(
                          "hover:bg-muted/20 transition-colors border-b border-border/40",
                          draggedIndex === idx && "opacity-40 bg-muted",
                          dragOverIndex === idx && "bg-primary/5 border-t-2 border-primary"
                        )}
                      >
                        <td className="w-8 px-2 py-3 text-center cursor-grab active:cursor-grabbing select-none">
                          <GripVertical className="size-4 text-muted-foreground/60 mx-auto" />
                        </td>
                        <td className="px-4 py-3 text-xs font-mono font-bold text-foreground">
                          <span className="inline-flex items-center gap-1.5 rounded-md bg-muted px-2 py-1">
                            <Barcode className="size-3 text-muted-foreground" />
                            {item.code}
                          </span>
                        </td>
                        <td className="px-4 py-3 font-semibold text-foreground max-w-xs truncate">
                          {pName}
                        </td>
                        <td className="px-4 py-3 text-xs text-center text-muted-foreground">
                          {item.unitname || "-"}
                        </td>
                        <td className="px-4 py-3 text-xs text-right font-mono font-medium text-foreground">
                          {item.price !== undefined ? formatMoney(item.price) : "-"}
                        </td>
                        <td className="px-4 py-3 text-center">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                            onClick={() => handleRemoveItem(item.code)}
                            aria-label={language === "th" ? "ลบบาร์โค้ดออกจากหมวด" : "Remove barcode"}
                            title={language === "th" ? "ลบบาร์โค้ดออกจากหมวด" : "Remove barcode"}
                          >
                            <Trash2 className="size-3.5" />
                          </Button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>

      </CardContent>
      {renderSearchDialog()}
    </Card>
  );
}

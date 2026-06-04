"use client";

import React, { useState, useEffect, useMemo, useRef } from "react";
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
  ShoppingBag,
  GripVertical,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { listBarcodes } from "@/lib/product-barcode/api";
import { pickName } from "@/lib/product-barcode/utils";
import { type AuthSession } from "@/lib/workspace-models";
import { type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";

type SettingRecord = Record<string, unknown>;

interface NameX {
  code: string;
  name: string;
}

interface CodeXSort {
  code: string;       // itemcode
  barcode: string;
  unitcode: string;
  names: NameX[];
  unitnames: NameX[];
}

export interface ProductCategoryItemsEditorProps {
  auth: AuthSession | null;
  workspace: { shop: { holding_code: string } } | null;
  language: string;
  categorySelectedGuid: string;
  categoryRecord: SettingRecord | null;
  onRefresh?: () => void;
  setEditing: (record: SettingRecord | null) => void;
  saving: boolean;
  setSaving: (saving: boolean) => void;
  onUnsavedChangesChange?: (hasChanges: boolean) => void;
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
}: ProductCategoryItemsEditorProps) {
  const [localCodelist, setLocalCodelist] = useState<CodeXSort[]>([]);
  const [searchDialogOpen, setSearchDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState("");
  const [notice, setNotice] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  // Compare current localCodelist with original categoryRecord.codelist
  const hasChanges = useMemo(() => {
    if (!categoryRecord) return false;
    const originalList = Array.isArray(categoryRecord.codelist) ? categoryRecord.codelist : [];
    if (originalList.length !== localCodelist.length) return true;
    return localCodelist.some((item, idx) => {
      const orig = originalList[idx];
      if (!orig) return true;
      const origBarcode = orig.barcode ?? orig.itemcode ?? "";
      return item.barcode !== origBarcode;
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
  useEffect(() => {
    if (categoryRecord) {
      const rawCodelist = categoryRecord.codelist;
      const parsedList: CodeXSort[] = Array.isArray(rawCodelist)
        ? rawCodelist.map((item: any) => ({
            code: String(item.code ?? item.itemcode ?? ""),
            barcode: String(item.barcode ?? ""),
            unitcode: String(item.unitcode ?? item.itemunitcode ?? ""),
            names: Array.isArray(item.names) ? item.names : [],
            unitnames: Array.isArray(item.unitnames) ? item.unitnames : [],
          }))
        : [];
      setLocalCodelist(parsedList);
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

  // Query barcodes from backend
  useEffect(() => {
    if (!searchDialogOpen || !auth || !workspace) return;

    const holding_code = workspace.shop.holding_code;
    let active = true;
    async function fetchBarcodes() {
      setSearching(true);
      setSearchError("");
      try {
        const response = await listBarcodes(auth, {
          holding_code,
          keyword: debouncedQuery,
          limit: 50,
        });
        if (!active) return;
        if (response.success) {
          setSearchResults(response.data ?? []);
        } else {
          setSearchError(response.message || "Failed to load products");
        }
      } catch (err: any) {
        if (!active) return;
        setSearchError(err.message || "Network error");
      } finally {
        if (active) setSearching(false);
      }
    }

    void fetchBarcodes();
    return () => {
      active = false;
    };
  }, [debouncedQuery, searchDialogOpen, auth, workspace]);

  const isSelectedLoading = categorySelectedGuid && (!categoryRecord || (categoryRecord.guid_fixed !== categorySelectedGuid && categoryRecord.guidfixed !== categorySelectedGuid));

  if (isSelectedLoading) {
    return (
      <Card className="h-full border-border bg-card shadow-sm">
        <CardContent className="grid h-full min-h-60 place-items-center p-4 text-center text-sm text-muted-foreground">
          <div className="flex flex-col items-center gap-2">
            <Loader2 className="size-8 animate-spin text-primary" />
            <span>{language === "th" ? "กำลังโหลดข้อมูลสินค้า..." : "Loading category products..."}</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!categoryRecord) {
    return (
      <Card className="h-full border-border bg-card shadow-sm">
        <CardContent className="grid h-full min-h-60 place-items-center p-4 text-center text-sm text-muted-foreground">
          <div className="grid gap-2">
            <ShoppingBag className="mx-auto size-8 text-primary/70" />
            <b className="text-foreground">
              {language === "th" ? "เลือกหมวดหมู่สินค้า" : "Select a category"}
            </b>
            <span>
              {language === "th"
                ? "กรุณาเลือกหมวดสินค้าจากโครงสร้างต้นไม้ด้านซ้ายเพื่อจัดการสินค้าในหมวด"
                : "Select a category from the tree on the left to manage its product items."}
            </span>
          </div>
        </CardContent>
      </Card>
    );
  }

  const categoryName = pickName(categoryRecord.names as any[], language) || String(categoryRecord.code || "");

  // Add item to local codelist
  const handleAddItem = (item: any) => {
    const isDuplicate = localCodelist.some((x) => x.barcode === item.barcode);
    if (isDuplicate) return;

    const newItem: CodeXSort = {
      code: item.itemcode,
      barcode: item.barcode,
      unitcode: item.itemunitcode || "",
      names: Array.isArray(item.names) ? item.names : [],
      unitnames: Array.isArray(item.itemunitnames) ? item.itemunitnames : [],
    };

    setLocalCodelist((prev) => [...prev, newItem]);
  };

  // Remove item from local codelist
  const handleRemoveItem = (barcode: string) => {
    setLocalCodelist((prev) => prev.filter((item) => item.barcode !== barcode));
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
    setLocalCodelist(list);
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

    const guid = String(categoryRecord.guid_fixed || categoryRecord.guidfixed || "");
    try {
      const payload = {
        ...categoryRecord,
        codelist: localCodelist,
      };

      const response = await fetch(
        `/api/system-settings/productcategorylist/${encodeURIComponent(guid)}?holding_code=${encodeURIComponent(workspace.shop.holding_code)}`,
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
            holding_code: workspace.shop.holding_code,
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

  // Render product search modal dialog
  const renderSearchDialog = () => {
    if (!searchDialogOpen) return null;

    return createPortal(
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
        onClick={() => setSearchDialogOpen(false)}
      >
        <div
          className="flex h-[80vh] w-full max-w-2xl flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          <header className="flex items-center justify-between border-b border-border px-4 py-3">
            <span className="text-base font-bold">
              {language === "th" ? "ค้นหาและเพิ่มสินค้า" : "Search and Add Products"}
            </span>
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
              <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                autoFocus
                placeholder={language === "th" ? "ค้นหาด้วยรหัสสินค้า, บาร์โค้ด หรือชื่อสินค้า..." : "Search by code, barcode, or name..."}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-10 pl-9 rounded-lg border-input bg-background"
              />
            </div>
          </div>

          <div className="flex-1 overflow-y-auto p-2">
            {searching ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground">
                <Loader2 className="size-6 animate-spin text-primary" />
                <span>{language === "th" ? "กำลังโหลดสินค้า..." : "Loading products..."}</span>
              </div>
            ) : searchError ? (
              <div className="p-4 text-center text-sm text-destructive font-medium">
                {searchError}
              </div>
            ) : searchResults.length === 0 ? (
              <div className="py-12 text-center text-sm text-muted-foreground">
                {language === "th" ? "ไม่พบสินค้าที่ตรงตามคำค้นหา" : "No products found"}
              </div>
            ) : (
              <div className="divide-y divide-border/60">
                {searchResults.map((item, idx) => {
                  const isAdded = localCodelist.some((x) => x.barcode === item.barcode);
                  const pName = pickName(item.names, language) || item.barcode;
                  const uName = pickName(item.itemunitnames, language) || item.itemunitcode || "";
                  return (
                    <div
                      key={item.guidfixed || item.guid_fixed || `${item.barcode || ""}-${item.itemunitcode || ""}-${idx}`}
                      className="flex items-center justify-between gap-4 p-3 hover:bg-muted/40 rounded-lg transition-colors"
                    >
                      <div className="min-w-0 flex-1">
                        <div className="font-semibold text-sm text-foreground truncate">
                          {pName}
                        </div>
                        <div className="flex items-center gap-2 mt-0.5 text-xs text-muted-foreground font-medium">
                          <span>รหัส: {item.itemcode}</span>
                          <span>•</span>
                          <span>บาร์โค้ด: {item.barcode}</span>
                          {uName && (
                            <>
                              <span>•</span>
                              <span>หน่วย: {uName}</span>
                            </>
                          )}
                        </div>
                      </div>
                      <Button
                        type="button"
                        size="sm"
                        variant={isAdded ? "ghost" : "outline"}
                        disabled={isAdded}
                        className={cn(
                          "h-8 rounded-lg font-semibold",
                          isAdded ? "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/20" : "hover:border-primary/50 hover:bg-primary/5 hover:text-primary"
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
    <Card className="flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm">
      <CardHeader className="flex flex-row items-center justify-between border-b border-border bg-muted/20 px-4 py-3 shrink-0">
        <div className="grid gap-0.5">
          <CardTitle className="text-sm font-bold text-foreground">
            {language === "th" ? "สินค้าในหมวด" : "Products in Category"}
          </CardTitle>
          <span className="text-xs font-semibold text-primary">
            {categoryName}
          </span>
        </div>
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
          {language === "th" ? "เพิ่มสินค้า" : "Add Product"}
        </Button>
      </CardHeader>

      <CardContent className="flex min-h-0 flex-1 flex-col p-0">
        {notice && (
          <div
            className={cn(
              "px-4 py-2 text-xs font-semibold border-b shrink-0",
              notice.type === "success"
                ? "bg-emerald-50 text-emerald-800 border-emerald-100 dark:bg-emerald-950/20 dark:text-emerald-400 dark:border-emerald-900/40"
                : "bg-destructive/10 text-destructive border-destructive/20"
            )}
          >
            {notice.text}
          </div>
        )}

        <div className="flex-1 overflow-y-auto">
          {localCodelist.length === 0 ? (
            <div className="flex h-60 flex-col items-center justify-center gap-1.5 p-4 text-center text-sm text-muted-foreground">
              <Info className="size-8 text-muted-foreground/60" />
              <span className="font-semibold text-foreground">
                {language === "th" ? "ยังไม่มีสินค้าในหมวดนี้" : "No products in this category"}
              </span>
              <span className="text-xs text-muted-foreground max-w-xs leading-normal">
                {language === "th"
                  ? "กดปุ่ม 'เพิ่มสินค้า' ด้านบนเพื่อค้นหาบาร์โค้ดสินค้าที่ต้องการนำมาแสดงในหมวดหมู่นี้"
                  : "Click 'Add Product' above to search and link barcodes to this category."}
              </span>
            </div>
          ) : (
            <div className="w-full overflow-x-auto">
              <table className="w-full text-left text-sm border-collapse">
                <thead>
                  <tr className="border-b border-border bg-muted/30 text-xs font-bold text-muted-foreground uppercase tracking-wider select-none">
                    <th className="w-8 px-2 py-2.5"></th>
                    <th className="px-4 py-2.5 font-bold">{language === "th" ? "บาร์โค้ด" : "Barcode"}</th>
                    <th className="px-4 py-2.5 font-bold">{language === "th" ? "รหัสสินค้า" : "Product Code"}</th>
                    <th className="px-4 py-2.5 font-bold">{language === "th" ? "ชื่อสินค้า" : "Name"}</th>
                    <th className="px-4 py-2.5 font-bold">{language === "th" ? "หน่วย" : "Unit"}</th>
                    <th className="px-4 py-2.5 text-center font-bold w-12">{language === "th" ? "จัดการ" : "Action"}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {localCodelist.map((item, idx) => {
                    const pName = pickName(item.names, language) || item.barcode;
                    const uName = pickName(item.unitnames, language) || item.unitcode || "";
                    return (
                      <tr
                        key={`${item.barcode}:${idx}`}
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
                        <td className="px-4 py-3 font-semibold text-xs font-mono select-all">
                          {item.barcode}
                        </td>
                        <td className="px-4 py-3 text-xs font-mono font-medium">
                          {item.code}
                        </td>
                        <td className="px-4 py-3 font-semibold text-foreground max-w-xs truncate">
                          {pName}
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground font-semibold">
                          {uName}
                        </td>
                        <td className="px-4 py-3 text-center">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                            onClick={() => handleRemoveItem(item.barcode)}
                            aria-label={language === "th" ? "ลบสินค้าออกจากหมวด" : "Remove product"}
                            title={language === "th" ? "ลบสินค้าออกจากหมวด" : "Remove product"}
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

        <footer className="flex items-center justify-between gap-4 border-t border-border bg-muted/20 px-4 py-3 shrink-0">
          <span className="text-xs font-semibold text-muted-foreground">
            {language === "th" ? `ทั้งหมด ${localCodelist.length} รายการ` : `Total ${localCodelist.length} items`}
          </span>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={saving}
              onClick={() => {
                if (categoryRecord) {
                  const rawCodelist = categoryRecord.codelist;
                  const parsedList: CodeXSort[] = Array.isArray(rawCodelist)
                    ? rawCodelist.map((item: any) => ({
                        code: String(item.code ?? item.itemcode ?? ""),
                        barcode: String(item.barcode ?? ""),
                        unitcode: String(item.unitcode ?? item.itemunitcode ?? ""),
                        names: Array.isArray(item.names) ? item.names : [],
                        unitnames: Array.isArray(item.unitnames) ? item.unitnames : [],
                      }))
                    : [];
                  setLocalCodelist(parsedList);
                }
                setNotice(null);
              }}
              className="rounded-lg font-semibold"
            >
              {language === "th" ? "ยกเลิก" : "Cancel"}
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={saving}
              onClick={handleSave}
              className="rounded-lg gap-1.5 font-semibold"
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
        </footer>
      </CardContent>
      {renderSearchDialog()}
    </Card>
  );
}

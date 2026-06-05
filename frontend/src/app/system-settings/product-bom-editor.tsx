"use client";

import { useState, useEffect, useMemo, type Dispatch, type DragEvent, type SetStateAction } from "react";
import {
  Plus,
  Trash2,
  ChevronRight,
  ChevronDown,
  GripVertical,
  AlertCircle,
  Save,
  Loader2,
  RefreshCcw,
  Network,
  X,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { type MasterEntry } from "@/lib/product-barcode/api";
import {
  pickName,
  toProductUnitOptions,
  type ProductUnitOption,
} from "@/lib/product-barcode/utils";
import type { AuthSession, LocalizedName } from "@/lib/workspace-models";
import type { LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";

// Define TypeScript interfaces for our BOM structures
interface BOMItemInput {
  barcodeguidfixed?: string;
  barcode: string;
  ref_type?: "product" | "recipe";
  names: LocalizedName[];
  item_unit_code: string;
  itemunitnames: LocalizedName[];
  qty: number;
  yield_percent: number; // default 100
  averagecost: number;
  materialtype?: number;
  bom?: BOMItemInput[];
}

interface BOMVersion {
  guidfixed: string;
  start_date: string;
  end_date: string | null;
  bom: BOMItemInput[];
}

interface ProductBOMRecord {
  guidfixed: string;
  barcode_guid?: string;
  barcode: string;
  itemcode?: string;
  names: LocalizedName[];
  item_unit_code: string;
  itemunitnames?: LocalizedName[];
  price?: number;
  prices?: { price: number }[];
  materialtype?: number;
  bom?: BOMItemInput[];
  boms?: BOMVersion[];
  unitOptions?: ProductUnitOption[];
}

interface ProductBomEditorProps {
  auth: AuthSession | null;
  workspace: any;
  language: LanguageCode;
  selectedRecord: ProductBOMRecord | null;
  records?: ProductBOMRecord[];
  onRefresh: () => void;
  onClose: () => void;
  setSelectedRecordId?: (id: string) => void;
  setRecords?: Dispatch<SetStateAction<any[]>>;
}

const RECIPE_COMPONENT_MATERIALTYPE_QUERY = "1,2,4";
const RECIPE_COMPONENT_MATERIAL_TYPES = new Set([1, 2, 4]);

export function ProductBomEditor({
  auth,
  workspace,
  language,
  selectedRecord,
  records = [],
  onRefresh,
  onClose,
  setSelectedRecordId,
  setRecords,
}: ProductBomEditorProps) {
  const { confirm, confirmationDialog } = useConfirmDialog();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [bomVersions, setBomVersions] = useState<BOMVersion[]>([]);
  const [selectedVersionIdx, setSelectedVersionIdx] = useState<number>(0);
  const [bomItems, setBomItems] = useState<BOMItemInput[]>([]);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({});
  const [unsavedChanges, setUnsavedChanges] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);

  // States for setting up new recipe
  const [parentItemCode, setParentItemCode] = useState("");
  const [parentNames, setParentNames] = useState<LocalizedName[]>([]);
  const [parentUnitCode, setParentUnitCode] = useState("RECIPE");
  const [parentUnitNames, setParentUnitNames] = useState<LocalizedName[]>([{ code: "th", name: "สูตร" }]);
  const [parentPrice, setParentPrice] = useState(0);
  const [ingredientProduct, setIngredientProduct] = useState<any | null>(null);
  const [ingredientUnitOpen, setIngredientUnitOpen] = useState(false);
  const [recipePickerOpen, setRecipePickerOpen] = useState(false);

  // Helper to update current version's BOM items
  const updateCurrentVersionBOM = (newItems: BOMItemInput[]) => {
    setBomItems(newItems);
    setBomVersions((prev) => {
      const updated = [...prev];
      if (updated[selectedVersionIdx]) {
        updated[selectedVersionIdx] = {
          ...updated[selectedVersionIdx],
          bom: newItems,
        };
      }
      return updated;
    });
    setUnsavedChanges(true);
  };

  const updateVersionDate = (key: "start_date" | "end_date", value: string | null) => {
    setBomVersions((prev) => {
      const updated = [...prev];
      if (updated[selectedVersionIdx]) {
        updated[selectedVersionIdx] = {
          ...updated[selectedVersionIdx],
          [key]: value,
        };
      }
      return updated;
    });
    setUnsavedChanges(true);
  };

  const handleAddVersion = () => {
    const currentVerBOM = bomVersions[selectedVersionIdx]?.bom || [];
    const copiedBOM = currentVerBOM.map((item) => ({
      ...item,
      bom: item.bom ? item.bom.map((child) => ({ ...child })) : undefined,
    }));

    const newVer: BOMVersion = {
      guidfixed: "",
      start_date: new Date().toISOString().substring(0, 10),
      end_date: null,
      bom: copiedBOM,
    };

    setBomVersions((prev) => [...prev, newVer]);
    setSelectedVersionIdx(bomVersions.length);
    setBomItems(copiedBOM);
    setUnsavedChanges(true);
  };

  const handleDeleteVersion = (index: number) => {
    if (bomVersions.length <= 1) {
      confirm({
        title: language === "th" ? "ไม่สามารถลบได้" : "Cannot Delete",
        description:
          language === "th"
            ? "ต้องมีสูตรผลิตอย่างน้อย 1 เวอร์ชัน"
            : "At least one recipe version is required.",
        confirmLabel: language === "th" ? "ตกลง" : "OK",
        cancelLabel: "",
      });
      return;
    }

    confirm({
      title: language === "th" ? "ยืนยันการลบสูตรเวอร์ชันนี้" : "Confirm Delete Version",
      description:
        language === "th"
          ? "คุณต้องการลบสูตรผลิตเวอร์ชันนี้ใช่หรือไม่?"
          : "Are you sure you want to delete this recipe version?",
      confirmLabel: language === "th" ? "ลบ" : "Delete",
      cancelLabel: language === "th" ? "ยกเลิก" : "Cancel",
    }).then((confirmed) => {
      if (confirmed) {
        const updated = bomVersions.filter((_, i) => i !== index);
        setBomVersions(updated);
        const nextIdx = Math.max(0, index - 1);
        setSelectedVersionIdx(nextIdx);
        setBomItems(updated[nextIdx]?.bom || []);
        setUnsavedChanges(true);
      }
    });
  };

  // Load and map recipe master data when selectedRecord changes.
  useEffect(() => {
    const mapItems = (items: any[] = []): BOMItemInput[] =>
      items.map((item) => ({
        barcodeguidfixed: item.guidfixed || item.barcodeguidfixed || item.barcode_guid || "",
        barcode: item.barcode || "",
        ref_type: item.ref_type === "recipe" ? "recipe" : "product",
        names: item.names || [],
        item_unit_code: item.item_unit_code || "",
        itemunitnames: item.itemunitnames || [],
        qty: item.qty ?? 1,
        yield_percent: item.yield_percent ?? 100,
        averagecost: item.averagecost ?? item.unit_cost ?? 0,
        materialtype: item.materialtype ?? 0,
        bom: item.bom ? mapItems(item.bom) : [],
      }));

    if (!selectedRecord) {
      setBomVersions([]);
      setBomItems([]);
      setSelectedVersionIdx(0);
      setUnsavedChanges(false);
      return;
    }

    const loadFromRecord = (record: ProductBOMRecord, markDirty: boolean) => {
      const defaultBom = mapItems(record.bom || []);
      const versions =
        record.boms && record.boms.length > 0
          ? record.boms.map((ver: any) => ({
              guidfixed: ver.guidfixed || "",
              start_date: ver.start_date ? ver.start_date.substring(0, 10) : new Date().toISOString().substring(0, 10),
              end_date: ver.end_date ? ver.end_date.substring(0, 10) : null,
              bom: mapItems(ver.bom || []),
            }))
          : [
              {
                guidfixed: "",
                start_date: new Date().toISOString().substring(0, 10),
                end_date: null,
                bom: defaultBom,
              },
            ];

      setParentItemCode(record.barcode || record.itemcode || "");
      setParentNames(record.names || []);
      setParentUnitCode(record.item_unit_code || "RECIPE");
      setParentUnitNames(record.itemunitnames || [{ code: "th", name: "สูตร" }]);
      setParentPrice(record.price ?? record.prices?.[0]?.price ?? 0);
      setBomVersions(versions);
      setSelectedVersionIdx(0);
      setBomItems(versions[0]?.bom || []);
      setUnsavedChanges(markDirty);
    };

    if (selectedRecord.guidfixed.startsWith("virtual-")) {
      loadFromRecord(selectedRecord, true);
      return;
    }

    if (!auth || !workspace) {
      loadFromRecord(selectedRecord, false);
      return;
    }

    setLoading(true);
    fetch(`/api/system-settings/product_bom/${encodeURIComponent(selectedRecord.guidfixed)}?holdingcode=${encodeURIComponent(workspace.shop.holdingcode)}`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
        "x-bc-backend-url": auth.backendUrl || "",
      },
    })
      .then((res) => {
        if (!res.ok) throw new Error("Failed to load recipe");
        return res.json();
      })
      .then((resJson) => {
        if (resJson.success && resJson.data) loadFromRecord(resJson.data, false);
        else loadFromRecord(selectedRecord, false);
      })
      .catch((e) => {
        console.error(e);
        loadFromRecord(selectedRecord, false);
      })
      .finally(() => setLoading(false));
  }, [selectedRecord, auth, workspace]);

  // Handle Drag & Drop sorting
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

  const handleDragStart = (index: number) => {
    setDraggedIndex(index);
  };

  const handleDragOver = (e: DragEvent<HTMLTableRowElement>, index: number) => {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === index) return;
  };

  const handleDrop = (index: number) => {
    if (draggedIndex === null || draggedIndex === index) return;
    const items = [...bomItems];
    const draggedItem = items[draggedIndex];
    items.splice(draggedIndex, 1);
    items.splice(index, 0, draggedItem);
    updateCurrentVersionBOM(items);
    setDraggedIndex(null);
  };

  // Calculate costs recursively
  const calculateTotalCost = (items: BOMItemInput[]): number => {
    return items.reduce((sum, item) => {
      const qty = item.qty || 0;
      const yieldPct = item.yield_percent || 100;
      const unitCost =
        item.bom && item.bom.length > 0 ? calculateTotalCost(item.bom) : item.averagecost || 0;
      const divisor = yieldPct > 0 ? yieldPct / 100 : 1;
      return sum + (qty / divisor) * unitCost;
    }, 0);
  };

  const hasNestedBOM = (item: BOMItemInput) => Boolean(item.bom && item.bom.length > 0);
  const isRecipeComponent = (item: BOMItemInput) => item.ref_type === "recipe" || hasNestedBOM(item);
  const componentUnitCost = (item: BOMItemInput) =>
    hasNestedBOM(item) ? calculateTotalCost(item.bom || []) : item.averagecost || 0;

  const rollupCost = useMemo(() => calculateTotalCost(bomItems), [bomItems]);

  const salePrice = useMemo(() => {
    if (!selectedRecord) return 0;
    return parentPrice;
  }, [selectedRecord, parentPrice]);

  const profitMarginPercent = useMemo(() => {
    if (salePrice <= 0) return 0;
    return ((salePrice - rollupCost) / salePrice) * 100;
  }, [salePrice, rollupCost]);

  // Safe names selection helper
  const productName = useMemo(() => {
    if (!selectedRecord) return "";
    return pickName(parentNames, language) || parentItemCode || (language === "th" ? "(สูตรใหม่)" : "(New Recipe)");
  }, [selectedRecord, language, parentNames, parentItemCode]);

  const saveDisabled = saving || loading || !unsavedChanges;

  // Toggle sub-recipe expand/collapse
  const toggleExpand = (barcode: string) => {
    setExpandedItems((prev) => ({
      ...prev,
      [barcode]: !prev[barcode],
    }));
  };

  // Update ingredient properties
  const updateItemProperty = (index: number, key: keyof BOMItemInput, value: any) => {
    const updated = [...bomItems];
    updated[index] = {
      ...updated[index],
      [key]: value,
    };
    updateCurrentVersionBOM(updated);
  };

  // Remove ingredient
  const removeItem = (index: number) => {
    const updated = bomItems.filter((_, i) => i !== index);
    updateCurrentVersionBOM(updated);
  };

  const isAllowedRecipeComponent = (productData: any) =>
    RECIPE_COMPONENT_MATERIAL_TYPES.has(Number(productData?.materialtype));

  const showInvalidRecipeComponentNotice = () => {
    confirm({
      title: language === "th" ? "เลือกสินค้าไม่ได้" : "Invalid Ingredient",
      description:
        language === "th"
          ? "สูตรผลิตเลือกได้เฉพาะวัตถุดิบ กึ่งสำเร็จรูป และสินค้าเกษตร"
          : "Recipes can only use Material, Semi-Finished, or Agricultural products as ingredients.",
      confirmLabel: language === "th" ? "รับทราบ" : "OK",
      cancelLabel: "",
    });
  };

  // Add selected item from MasterPicker
  const addIngredientToBOM = (productData: any, unitOpt: ProductUnitOption) => {
    if (!isAllowedRecipeComponent(productData)) {
      showInvalidRecipeComponentNotice();
      return;
    }

    if (bomItems.some((item) => item.barcode === unitOpt.barcode)) {
      confirm({
        title: language === "th" ? "สินค้าอยู่ในสูตรแล้ว" : "Item Already in Formula",
        description:
          language === "th"
            ? "ส่วนประกอบนี้อยู่ในรายการอยู่แล้ว ไม่จำเป็นต้องเพิ่มซ้ำ"
            : "This ingredient is already in the list.",
        confirmLabel: language === "th" ? "รับทราบ" : "OK",
        cancelLabel: "",
      });
      return;
    }

    const materialType = Number(productData.materialtype);
    const newItem: BOMItemInput = {
      barcodeguidfixed: unitOpt.guidfixed,
      barcode: unitOpt.barcode,
      ref_type: "product",
      names: productData.names || [],
      item_unit_code: unitOpt.item_unit_code || "",
      itemunitnames: unitOpt.itemunitnames || [],
      qty: 1,
      yield_percent: 100,
      averagecost: unitOpt.averagecost || productData.averagecost || 0,
      materialtype: materialType,
      bom: productData.bom || [],
    };
    updateCurrentVersionBOM([...bomItems, newItem]);
  };

  const recipeOptions = useMemo(
    () =>
      records.filter((record) => {
        const code = String(record.barcode || record.itemcode || "").trim();
        if (!code) return false;
        if (record.guidfixed === selectedRecord?.guidfixed) return false;
        if (record.guidfixed?.startsWith("virtual-")) return false;
        return true;
      }),
    [records, selectedRecord?.guidfixed],
  );

  const addRecipeToBOM = (recipe: ProductBOMRecord) => {
    const recipeCode = String(recipe.barcode || recipe.itemcode || "").trim();
    if (!recipeCode) return;
    if (recipeCode === parentItemCode.trim()) {
      confirm({
        title: language === "th" ? "อ้างสูตรตัวเองไม่ได้" : "Invalid Sub-recipe",
        description: language === "th" ? "สูตรผลิตไม่สามารถอ้างอิงตัวเองเป็นสูตรย่อยได้" : "A recipe cannot reference itself.",
        confirmLabel: language === "th" ? "รับทราบ" : "OK",
        cancelLabel: "",
      });
      return;
    }
    if (bomItems.some((item) => item.ref_type === "recipe" && item.barcode === recipeCode)) {
      confirm({
        title: language === "th" ? "สูตรย่อยอยู่ในรายการแล้ว" : "Sub-recipe Already Added",
        description: language === "th" ? "สูตรนี้อยู่ในรายการส่วนประกอบแล้ว" : "This sub-recipe is already in the component list.",
        confirmLabel: language === "th" ? "รับทราบ" : "OK",
        cancelLabel: "",
      });
      return;
    }

    const newItem: BOMItemInput = {
      barcodeguidfixed: recipe.guidfixed,
      barcode: recipeCode,
      ref_type: "recipe",
      names: recipe.names || [],
      item_unit_code: recipe.item_unit_code || "RECIPE",
      itemunitnames: recipe.itemunitnames || [{ code: "th", name: "สูตร" }],
      qty: 1,
      yield_percent: 100,
      averagecost: calculateTotalCost(recipe.bom || []),
      materialtype: recipe.materialtype ?? 0,
      bom: recipe.bom || [],
    };
    updateCurrentVersionBOM([...bomItems, newItem]);
    setRecipePickerOpen(false);
  };

  // Add selected item from MasterPicker
  const handleAddIngredient = async (entry: MasterEntry) => {
    setPickerOpen(false);
    setLoading(true);
    try {
      // Fetch full product details to use existing barcode/unit choices from product.barcodes.
      const response = await fetch(`/api/product/${encodeURIComponent(entry.guidfixed)}`, {
        headers: {
          Authorization: `Bearer ${auth?.token}`,
          "x-bc-backend-url": auth?.backendUrl || "",
        },
      });
      const resJson = await response.json();
      if (resJson.success && resJson.data) {
        const productData = resJson.data;
        if (!isAllowedRecipeComponent(productData)) {
          showInvalidRecipeComponentNotice();
          return;
        }

        const unitOptions = toProductUnitOptions(productData);
        if (unitOptions.length === 0) {
          confirm({
            title: language === "th" ? "ไม่พบหน่วยนับสินค้า" : "No Unit Found",
            description:
              language === "th"
                ? "สินค้านี้ยังไม่มี barcode/unit ให้เลือก จึงเพิ่มเข้าสูตรไม่ได้"
                : "This product has no barcode/unit options and cannot be added to the recipe.",
            confirmLabel: language === "th" ? "รับทราบ" : "OK",
            cancelLabel: "",
          });
          return;
        }

        if (unitOptions.length === 1) {
          addIngredientToBOM(productData, unitOptions[0]);
        } else {
          setIngredientProduct({ ...productData, unitOptions });
          setIngredientUnitOpen(true);
        }
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  // Save changes
  const handleSave = async () => {
    if (!selectedRecord || !auth || !workspace) return;

    const recipeCode = parentItemCode.trim();
    if (!recipeCode) {
      confirm({
        title: language === "th" ? "ข้อมูลไม่ครบถ้วน" : "Incomplete Information",
        description: language === "th" ? "กรุณากรอกรหัสสูตรผลิต" : "Please enter the recipe code.",
        confirmLabel: language === "th" ? "ตกลง" : "OK",
        cancelLabel: "",
      });
      return;
    }
    if (parentNames.length === 0 || !parentNames.some((n) => n.name?.trim())) {
      confirm({
        title: language === "th" ? "ข้อมูลไม่ครบถ้วน" : "Incomplete Information",
        description: language === "th" ? "กรุณากรอกชื่อสูตรผลิตอย่างน้อย 1 ภาษา" : "Please enter at least one recipe name.",
        confirmLabel: language === "th" ? "ตกลง" : "OK",
        cancelLabel: "",
      });
      return;
    }

    setSaving(true);
    try {
      const isCreate = selectedRecord.guidfixed.startsWith("virtual-");
      const serializeItem = (item: BOMItemInput): Record<string, unknown> => ({
        guidfixed: item.barcodeguidfixed,
        barcode: item.barcode,
        ref_type: item.ref_type || "product",
        names: item.names || [],
        item_unit_code: item.item_unit_code || "",
        itemunitnames: item.itemunitnames || [],
        condition: true,
        dividevalue: 1,
        standvalue: 1,
        qty: item.qty,
        yield_percent: item.yield_percent,
        averagecost: item.averagecost || 0,
        materialtype: item.materialtype ?? 0,
        bom: (item.bom || []).map(serializeItem),
      });

      const sortedVersions = [...bomVersions].sort((a, b) => a.start_date.localeCompare(b.start_date));
      const payloadBOMs = sortedVersions.map((v, idx) => {
        const nextVersion = sortedVersions[idx + 1];
        const calculatedEndDate = nextVersion ? new Date(nextVersion.start_date).toISOString() : null;

        return {
          guidfixed: v.guidfixed || undefined,
          start_date: new Date(v.start_date).toISOString(),
          end_date: calculatedEndDate,
          bom: v.bom.map(serializeItem),
        };
      });

      const nowStr = new Date().toISOString().substring(0, 10);
      const activeVerIndex = sortedVersions.reduce((activeIndex, currentVer, currentIndex) => {
        if (currentVer.start_date <= nowStr) {
          return currentIndex;
        }
        return activeIndex;
      }, -1);

      const targetVer = activeVerIndex !== -1 ? sortedVersions[activeVerIndex] : sortedVersions[0];
      const activeBOM = targetVer ? targetVer.bom.map(serializeItem) : [];

      const payload = {
        guidfixed: isCreate ? undefined : selectedRecord.guidfixed,
        barcode: recipeCode,
        names: parentNames,
        item_unit_code: parentUnitCode.trim() || "RECIPE",
        itemunitnames: parentUnitNames,
        price: parentPrice,
        bom: activeBOM,
        boms: payloadBOMs,
        holdingcode: workspace.shop.holdingcode,
      };

      const saveResponse = await fetch(
        isCreate
          ? `/api/system-settings/product_bom?holdingcode=${encodeURIComponent(workspace.shop.holdingcode)}`
          : `/api/system-settings/product_bom/${encodeURIComponent(selectedRecord.guidfixed)}?holdingcode=${encodeURIComponent(workspace.shop.holdingcode)}`,
        {
          method: isCreate ? "POST" : "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
          body: JSON.stringify({
            backendUrl: auth.backendUrl,
            ...payload,
          }),
        },
      );

      const saveData = await saveResponse.json();
      if (!saveResponse.ok || !saveData.success) {
        throw new Error(saveData.message || "Failed to save recipe");
      }

      const savedId = String(saveData.id || "");
      setUnsavedChanges(false);
      confirm({
        title: language === "th" ? "บันทึกสำเร็จ" : "Saved Successfully",
        description:
          language === "th"
            ? "สูตรผลิตได้รับการบันทึกและคำนวณต้นทุนเรียบร้อยแล้ว"
            : "The bill of materials has been successfully saved.",
        confirmLabel: language === "th" ? "ตกลง" : "OK",
        cancelLabel: "",
      });

      onRefresh();
      if (isCreate && savedId) setSelectedRecordId?.(savedId);
    } catch (error: any) {
      confirm({
        title: language === "th" ? "เกิดข้อผิดพลาด" : "Error Occurred",
        description: error.message || "Request failed",
        confirmLabel: language === "th" ? "ปิด" : "Close",
        cancelLabel: "",
      });
    } finally {
      setSaving(false);
    }
  };

  // Render sub-recipe tree recursively (sidebar details view)
  const renderSubRecipeTree = (items: BOMItemInput[], depth = 1) => {
    return items.map((subItem) => {
      const hasChildren = hasNestedBOM(subItem);
      const isSubRecipe = isRecipeComponent(subItem);
      const isExpanded = expandedItems[subItem.barcode];
      const name = pickName(subItem.names, language) || subItem.barcode;
      const unit = pickName(subItem.itemunitnames, language) || subItem.item_unit_code;
      const cost = componentUnitCost(subItem);

      return (
        <div key={subItem.barcode} className="my-1.5 font-sans">
          <div
            className={cn(
              "flex items-center justify-between gap-4 py-1.5 px-3 rounded-md bg-muted/25 border border-dashed border-border/60 hover:bg-muted/40 transition-colors",
              depth === 1 ? "ml-4 border-l-sky-400" : "ml-8 border-l-indigo-400"
            )}
          >
            <div className="flex items-center gap-2 min-w-0">
              {hasChildren ? (
                <button
                  onClick={() => toggleExpand(subItem.barcode)}
                  className="p-0.5 rounded hover:bg-muted shrink-0"
                >
                  {isExpanded ? (
                    <ChevronDown className="size-3.5 text-sky-600" />
                  ) : (
                    <ChevronRight className="size-3.5 text-sky-600" />
                  )}
                </button>
              ) : (
                <div className="size-4" />
              )}
              <span className="truncate text-xs font-semibold text-foreground/80">{name}</span>
              {isSubRecipe && (
                <Badge variant="outline" className="text-[9px] h-3.5 bg-sky-50 text-sky-700 border-sky-200 shrink-0">
                  {language === "th" ? "สูตรย่อย" : "Sub-recipe"}
                </Badge>
              )}
            </div>

            <div className="flex items-center gap-4 text-xs shrink-0 font-medium">
              <span className="text-muted-foreground font-mono">
                {subItem.qty} {unit}
              </span>
              <span className="text-muted-foreground w-20 text-right">
                {language === "th" ? "ต้นทุน" : "Cost"}: ฿{cost.toFixed(2)}
              </span>
            </div>
          </div>
          {hasChildren && isExpanded && renderSubRecipeTree(subItem.bom || [], depth + 1)}
        </div>
      );
    });
  };

  // Render tree node inside dialog recursively (exploded structure view)
  const renderExplodedTreeNode = (item: BOMItemInput, path: string, level = 0) => {
    const hasChildren = hasNestedBOM(item);
    const isSub = isRecipeComponent(item);
    const itemCost = componentUnitCost(item);
    const yieldPct = item.yield_percent || 100;
    const divisor = yieldPct > 0 ? yieldPct / 100 : 1;
    const rowCost = (item.qty / divisor) * itemCost;
    const name = pickName(item.names, language) || item.barcode;
    const unit = pickName(item.itemunitnames, language) || item.item_unit_code;
    const share = rollupCost > 0 ? (rowCost / rollupCost) * 100 : 0;

    return (
      <div key={path} className="flex flex-col">
        <div className="flex items-center justify-between gap-4 py-2.5 border-b border-border/40 hover:bg-muted/30 px-3 transition-colors text-xs">
          <div className="flex items-center gap-2 min-w-0" style={{ paddingLeft: `${level * 24}px` }}>
            {level > 0 && (
              <span className="text-muted-foreground/30 font-light mr-1 shrink-0 font-mono select-none">
                └──
              </span>
            )}
            <div className="flex flex-col min-w-0">
              <span className="font-semibold text-foreground flex items-center gap-2 flex-wrap">
                <span className="truncate">{name}</span>
                {isSub && (
                  <Badge variant="outline" className="text-[9px] h-3.5 px-1 bg-sky-50 text-sky-700 border-sky-200">
                    {language === "th" ? "สูตรย่อย" : "Sub-recipe"}
                  </Badge>
                )}
                {item.materialtype === 1 && (
                  <Badge variant="outline" className="text-[9px] h-3.5 px-1 bg-green-50 text-green-700 border-green-200">
                    {language === "th" ? "วัตถุดิบ" : "Material"}
                  </Badge>
                )}
              </span>
              <span className="text-[10px] text-muted-foreground font-mono">{item.barcode}</span>
            </div>
          </div>

          <div className="flex items-center gap-6 shrink-0 font-semibold font-mono text-right">
            <span className="text-muted-foreground w-28">
              {item.qty} {unit}
              {yieldPct < 100 && ` (${language === "th" ? "สูญเสีย" : "Yield"} ${yieldPct}%)`}
            </span>
            <span className="text-foreground w-24">฿{itemCost.toFixed(2)}</span>
            <span className="text-sky-700 w-24">฿{rowCost.toFixed(2)}</span>
            <span className="text-emerald-600 w-16">{share.toFixed(1)}%</span>
          </div>
        </div>

        {hasChildren && item.bom?.map((child, i) => renderExplodedTreeNode(child, `${path}-${i}`, level + 1))}
      </div>
    );
  };


  if (!selectedRecord) {
    return (
      <Card className="h-full border-border bg-card shadow-sm flex items-center justify-center p-8">
        <div className="text-center space-y-2 max-w-sm">
          <AlertCircle className="mx-auto size-12 text-muted-foreground/60" />
          <h3 className="font-bold text-foreground">
            {language === "th" ? "เลือกสูตรผลิตเพื่อแก้ไข" : "Select a Recipe to Edit"}
          </h3>
          <p className="text-sm text-muted-foreground">
            {language === "th"
              ? "เลือกสูตรผลิตจากรายการด้านซ้าย หรือเพิ่มรหัสสูตรผลิตใหม่"
              : "Choose a recipe from the list on the left or create a new recipe code."}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-5 h-full items-stretch">
      {/* LEFT: Ingredients Editor Grid */}
      <Card className="border-border/80 bg-card/50 backdrop-blur-sm shadow-md rounded-xl flex flex-col h-full overflow-hidden transition-all duration-300">
        <CardHeader className="border-b border-border bg-gradient-to-r from-muted/30 via-background to-muted/20 py-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 shrink-0">
          <div className="min-w-0 space-y-1">
            <CardTitle className="text-lg font-bold flex items-center gap-2.5 flex-wrap">
              <span className="bg-gradient-to-r from-foreground to-foreground/80 bg-clip-text text-transparent">
                {language === "th" ? "รายการวัตถุดิบและส่วนประกอบ" : "Ingredients List"}
              </span>
              {unsavedChanges && (
                <Badge variant="outline" className="animate-pulse bg-destructive/10 border-destructive/30 text-destructive text-[10px] h-5 rounded-full font-semibold px-2.5">
                  {language === "th" ? "ยังไม่ได้บันทึก" : "Unsaved Changes"}
                </Badge>
              )}
            </CardTitle>
            <CardDescription className="text-xs text-muted-foreground/90 font-medium">
              {language === "th" ? "สูตรผลิต" : "Recipe"}{" "}
              <span className="text-foreground font-semibold">{productName}</span>{" "}
              {selectedRecord.guidfixed.startsWith("virtual-")
                ? (parentItemCode ? `[รหัสสูตร: ${parentItemCode}]` : "")
                : (selectedRecord.itemcode || selectedRecord.barcode ? `[รหัสสูตร: ${selectedRecord.itemcode || selectedRecord.barcode}]` : "")}
            </CardDescription>
          </div>
          <div className="flex flex-wrap gap-2 w-full sm:w-auto">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setDetailModalOpen(true)}
              disabled={bomItems.length === 0}
              className="h-8.5 rounded-lg border-indigo-200 text-indigo-700 bg-indigo-50/10 hover:bg-indigo-50 dark:border-indigo-900/50 dark:text-indigo-400 dark:hover:bg-indigo-950/30 text-xs transition-all"
            >
              <Network className="size-3.5 mr-1.5" />
              {language === "th" ? "โครงสร้างสูตรละเอียด" : "Explode Tree"}
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={handleSave}
              disabled={saveDisabled}
              className={cn(
                "h-8.5 rounded-lg text-white font-semibold shadow-sm text-xs transition-all px-4",
                saveDisabled
                  ? "bg-muted text-muted-foreground cursor-not-allowed border-0"
                  : "bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 border-0 hover:shadow-md hover:scale-[1.01]"
              )}
            >
              {saving ? <Loader2 className="size-3.5 animate-spin mr-1.5" /> : <Save className="size-3.5 mr-1.5" />}
              {language === "th" ? "บันทึกสูตร" : "Save BOM"}
            </Button>
          </div>
        </CardHeader>

        <CardContent className="p-0 flex-1 overflow-y-auto scrollbar-thin flex flex-col">
          <div className="p-5 border-b border-border bg-gradient-to-b from-muted/20 to-transparent space-y-4 shrink-0">
            <h4 className="text-xs font-bold text-foreground/80 uppercase tracking-wider select-none">
              {language === "th" ? "ข้อมูลสูตรผลิต" : "Recipe Master"}
            </h4>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 p-4 border border-sky-100 dark:border-sky-950/40 rounded-xl bg-sky-500/5 shadow-sm">
              <div className="space-y-1.5">
                <label className="text-[11px] font-bold text-sky-800 dark:text-sky-400 uppercase">
                  {language === "th" ? "รหัสสูตรผลิต *" : "Recipe Code *"}
                </label>
                <Input
                  className="h-8.5 text-xs font-mono rounded-lg focus-visible:ring-sky-500 bg-background"
                  placeholder="RECIPE-001"
                  value={parentItemCode}
                  onChange={(e) => {
                    const next = e.target.value.trimStart();
                    setParentItemCode(next);
                    setUnsavedChanges(true);
                  }}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-[11px] font-bold text-sky-800 dark:text-sky-400 uppercase">
                  {language === "th" ? "หน่วยสูตร" : "Recipe Unit"}
                </label>
                <Input
                  className="h-8.5 text-xs font-mono rounded-lg focus-visible:ring-sky-500 bg-background"
                  value={parentUnitCode}
                  onChange={(e) => {
                    const next = e.target.value.trimStart().toUpperCase();
                    setParentUnitCode(next);
                    setParentUnitNames([{ code: language, name: next || "สูตร" }]);
                    setUnsavedChanges(true);
                  }}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-[11px] font-bold text-sky-800 dark:text-sky-400 uppercase">
                  {language === "th" ? "ราคาประเมินของสูตร" : "Estimated Recipe Price"}
                </label>
                <Input
                  type="number"
                  className="h-8.5 text-xs font-mono rounded-lg focus-visible:ring-sky-500 bg-background"
                  placeholder="0.00"
                  value={parentPrice || ""}
                  onChange={(e) => {
                    setParentPrice(parseFloat(e.target.value) || 0);
                    setUnsavedChanges(true);
                  }}
                />
              </div>

              <div className="md:col-span-3 border-t border-dashed border-sky-200/50 pt-3">
                <NamesEditor
                  names={parentNames}
                  onChange={(next) => {
                    setParentNames(next);
                    setUnsavedChanges(true);
                  }}
                  languages={["th", "en"]}
                  label={language === "th" ? "ชื่อสูตรผลิต" : "Recipe names"}
                  firstRequired
                  language={language}
                />
              </div>
            </div>
          </div>

          {/* Recipe Version History Control */}
          <div className="bg-muted/20 border-b border-border p-4.5 space-y-4 shrink-0">
            <div className="flex items-center justify-between gap-4">
              <span className="text-xs font-bold text-muted-foreground uppercase tracking-wider select-none">
                {language === "th" ? "ช่วงเวลาการใช้งานสูตรผลิต (Recipe Versions)" : "Recipe Versions"}
              </span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-7.5 text-xs rounded-lg border-sky-200 bg-sky-50/10 text-sky-700 hover:bg-sky-50 dark:border-sky-900/40 dark:text-sky-400 dark:hover:bg-sky-950/20 shadow-sm"
                onClick={handleAddVersion}
              >
                <Plus className="size-3.5 mr-1" />
                {language === "th" ? "เพิ่มเวอร์ชันสูตร" : "Add Version"}
              </Button>
            </div>

            <div className="flex flex-wrap gap-2.5">
              {(() => {
                const sorted = [...bomVersions].sort((a, b) => a.start_date.localeCompare(b.start_date));
                return sorted.map((ver, idx) => {
                  const originalIdx = bomVersions.findIndex((v) => v.start_date === ver.start_date);
                  const isActive = originalIdx === selectedVersionIdx;
                  const nextVer = sorted[idx + 1];
                  const displayEndDate = nextVer ? nextVer.start_date : (language === "th" ? 'ปัจจุบัน' : 'Present');

                  return (
                    <div key={idx} className="flex items-center gap-1 group/btn select-none">
                      <Button
                        type="button"
                        variant={isActive ? "default" : "outline"}
                        size="sm"
                        className={cn(
                          "h-8.5 text-xs font-bold rounded-lg transition-all duration-200 px-3",
                          isActive
                            ? "bg-gradient-to-r from-sky-600 to-blue-600 hover:from-sky-700 hover:to-blue-700 text-white shadow-sm border-0 scale-[1.02]"
                            : "bg-background text-muted-foreground border-border hover:bg-muted/50 hover:text-foreground"
                        )}
                        onClick={() => {
                          setSelectedVersionIdx(originalIdx);
                          setBomItems(ver.bom || []);
                        }}
                      >
                        {language === "th" ? `สูตรที่ ${idx + 1}` : `Recipe v${idx + 1}`}
                        <span className="ml-1.5 text-[9px] opacity-75 font-mono">
                          ({ver.start_date} - {displayEndDate})
                        </span>
                      </Button>
                      {bomVersions.length > 1 && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="h-8.5 w-8.5 text-destructive/60 hover:text-destructive hover:bg-destructive/10 rounded-lg shrink-0 transition-colors"
                          onClick={() => handleDeleteVersion(originalIdx)}
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                      )}
                    </div>
                  );
                });
              })()}
            </div>

            {bomVersions[selectedVersionIdx] && (
              <div className="grid grid-cols-1 gap-3 pt-3 border-t border-dashed border-border/80 max-w-xs">
                <div className="space-y-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase select-none">
                    {language === "th" ? "วันที่เริ่มใช้งานสูตร" : "Effective Start Date"}
                  </label>
                  <Input
                    type="date"
                    className="h-8.5 text-xs font-mono rounded-lg bg-background border-border/80 focus-visible:ring-sky-500"
                    value={bomVersions[selectedVersionIdx].start_date || ""}
                    onChange={(e) => updateVersionDate('start_date', e.target.value)}
                  />
                </div>
              </div>
            )}
          </div>

          <div className="flex-1 overflow-y-auto w-full">
            {bomItems.length === 0 ? (
              <div className="py-24 text-center text-muted-foreground space-y-3 bg-muted/5">
                <RefreshCcw className="mx-auto size-9 text-muted-foreground/20" />
                <p className="text-sm font-semibold text-foreground/60 select-none">
                  {language === "th" ? "ยังไม่มีส่วนประกอบในสูตรนี้" : "No ingredients added yet."}
                </p>
                <Button variant="link" size="sm" onClick={() => setPickerOpen(true)} className="text-sky-600 dark:text-sky-400 font-semibold">
                  {language === "th" ? "คลิกเพื่อเริ่มเพิ่มวัตถุดิบ" : "Click here to add first ingredient"}
                </Button>
              </div>
            ) : (
              <table className="w-full text-sm border-collapse text-left">
                <thead className="sticky top-0 bg-muted/90 backdrop-blur-md border-b border-border text-[10px] text-muted-foreground font-bold uppercase tracking-wider z-15 select-none shadow-sm">
                  <tr>
                    <th className="w-8 py-3 px-3"></th>
                    <th className="py-3 px-2">{language === "th" ? "ส่วนประกอบ / รหัสอ้างอิง" : "Component / Reference Code"}</th>
                    <th className="w-24 py-3 px-2 text-right">{language === "th" ? "จำนวน" : "Quantity"}</th>
                    <th className="w-20 py-3 px-2">{language === "th" ? "หน่วย" : "Unit"}</th>
                    <th className="w-24 py-3 px-2 text-right">{language === "th" ? "สูญเสีย (%)" : "Yield (%)"}</th>
                    <th className="w-24 py-3 px-2 text-right">{language === "th" ? "ต้นทุน/หน่วย" : "Cost/Unit"}</th>
                    <th className="w-28 py-3 px-2 text-right">{language === "th" ? "ราคารวม" : "Total"}</th>
                    <th className="w-12 py-3 px-3"></th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {bomItems.map((item, index) => {
                    const hasChildren = hasNestedBOM(item);
                    const isSubRecipe = isRecipeComponent(item);
                    const isExpanded = expandedItems[item.barcode];
                    const name = pickName(item.names, language) || item.barcode;
                    const unit = pickName(item.itemunitnames, language) || item.item_unit_code;

                    const unitCost = componentUnitCost(item);
                    const divisor = item.yield_percent > 0 ? item.yield_percent / 100 : 1;
                    const rowTotalCost = (item.qty / divisor) * unitCost;

                    return (
                      <tr
                        key={item.barcode}
                        draggable
                        onDragStart={() => handleDragStart(index)}
                        onDragOver={(e) => handleDragOver(e, index)}
                        onDrop={() => handleDrop(index)}
                        onDragEnd={() => setDraggedIndex(null)}
                        className={cn(
                          "group/row hover:bg-muted/20 transition-all duration-150 align-top",
                          draggedIndex === index && "opacity-40 bg-muted",
                          index % 2 === 0 ? "bg-background" : "bg-muted/5"
                        )}
                      >
                        {/* Drag Handle */}
                        <td className="py-4.5 px-3 align-middle select-none">
                          <GripVertical className="size-4 text-muted-foreground/30 group-hover/row:text-muted-foreground/60 cursor-grab hover:text-muted-foreground active:cursor-grabbing transition-colors" />
                        </td>

                        {/* Name & Details */}
                        <td className="py-3 px-2">
                          <div className="space-y-1.5">
                            <div className="flex items-center gap-1.5 flex-wrap">
                              {hasChildren && (
                                <button
                                  onClick={() => toggleExpand(item.barcode)}
                                  className="p-0.5 rounded-lg hover:bg-muted/80 shrink-0 text-sky-600 transition-colors"
                                >
                                  {isExpanded ? (
                                    <ChevronDown className="size-4" />
                                  ) : (
                                    <ChevronRight className="size-4" />
                                  )}
                                </button>
                              )}
                              <span className="font-semibold text-foreground/90">{name}</span>
                              {item.materialtype === 1 && (
                                <Badge
                                  variant="outline"
                                  className="text-[9px] h-4 bg-emerald-50 text-emerald-700 border-emerald-200/50 shrink-0 rounded-full font-bold select-none"
                                >
                                  {language === "th" ? "วัตถุดิบ" : "Material"}
                                </Badge>
                              )}
                            </div>
                            <div className="flex items-center gap-2 text-xs text-muted-foreground/80 pl-1 font-medium select-text">
                              <span className="font-mono bg-muted/40 px-1 rounded">{item.barcode}</span>
                              {isSubRecipe && (
                                <Badge
                                  variant="outline"
                                  className="text-[9px] h-4 bg-indigo-50 text-indigo-700 border-indigo-200/50 rounded-full font-bold select-none"
                                >
                                  {language === "th" ? "สูตรย่อย" : "Sub-recipe"}
                                </Badge>
                              )}
                            </div>
                          </div>

                          {/* Collapsible Sub-recipe tree */}
                          {hasChildren && isExpanded && (
                            <div className="mt-3 pr-2 border-t border-border/30 pt-2.5">
                              {renderSubRecipeTree(item.bom || [])}
                            </div>
                          )}
                        </td>

                        {/* Quantity Input */}
                        <td className="py-3 px-2 text-right align-middle">
                          <Input
                            type="number"
                            className="h-8 text-right font-bold w-22 ml-auto rounded-lg font-mono focus-visible:ring-indigo-500 bg-background"
                            value={item.qty}
                            onChange={(e) =>
                              updateItemProperty(index, "qty", Math.max(0, parseFloat(e.target.value) || 0))
                            }
                            min="0"
                            step="any"
                          />
                        </td>

                        {/* Unit */}
                        <td className="py-3 px-2 text-muted-foreground/90 font-medium align-middle select-none">{unit}</td>

                        {/* Yield % Input */}
                        <td className="py-3 px-2 text-right align-middle">
                          <div className="flex items-center gap-1 justify-end select-none">
                            <Input
                              type="number"
                              className="h-8 text-right w-18 rounded-lg font-mono focus-visible:ring-indigo-500 bg-background"
                              value={item.yield_percent}
                              onChange={(e) =>
                                updateItemProperty(
                                  index,
                                  "yield_percent",
                                  Math.min(100, Math.max(1, parseFloat(e.target.value) || 100))
                                )
                              }
                              min="1"
                              max="100"
                            />
                            <span className="text-xs text-muted-foreground/80 font-bold">%</span>
                          </div>
                        </td>

                        {/* Unit Cost */}
                        <td className="py-3 px-2 text-right font-mono align-middle select-text text-muted-foreground font-medium">฿{unitCost.toFixed(2)}</td>

                        {/* Total Cost */}
                        <td className="py-3 px-2 text-right font-bold font-mono align-middle select-text text-foreground">
                          ฿{rowTotalCost.toFixed(2)}
                        </td>

                        {/* Delete */}
                        <td className="py-3 px-3 text-right align-middle select-none">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-destructive/50 hover:text-destructive hover:bg-destructive/10 rounded-xl transition-all"
                            onClick={() => removeItem(index)}
                          >
                            <Trash2 className="size-3.5" />
                          </Button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}
          </div>

          {/* Add buttons at the bottom */}
          <div className="flex items-center gap-2 p-3 border-t border-border bg-muted/10 shrink-0">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setPickerOpen(true)}
              disabled={loading || saving}
              className="h-8.5 rounded-lg border-sky-200 text-sky-700 bg-sky-50/10 hover:bg-sky-50 dark:border-sky-900/50 dark:text-sky-400 dark:hover:bg-sky-950/30 text-xs transition-all"
            >
              <Plus className="size-3.5 mr-1.5" />
              {language === "th" ? "เพิ่มวัตถุดิบ" : "Add Material"}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setRecipePickerOpen(true)}
              disabled={loading || saving || recipeOptions.length === 0}
              className="h-8.5 rounded-lg border-violet-200 text-violet-700 bg-violet-50/10 hover:bg-violet-50 dark:border-violet-900/50 dark:text-violet-400 dark:hover:bg-violet-950/30 text-xs transition-all"
            >
              <Network className="size-3.5 mr-1.5" />
              {language === "th" ? "เพิ่มสูตรย่อย" : "Add Sub-recipe"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* MASTER DATA PICKER */}
      {pickerOpen && (
        <MasterPicker
          open={pickerOpen}
          onClose={() => setPickerOpen(false)}
          auth={auth}
          language={language}
          master="product"
          title={language === "th" ? "ค้นหาวัตถุดิบ / กึ่งสำเร็จรูป / สินค้าเกษตร" : "Search Materials / Semi-Finished / Agricultural"}
          filters={{ materialtype: RECIPE_COMPONENT_MATERIALTYPE_QUERY }}
          onSelect={handleAddIngredient}
        />
      )}

      {/* SUB-RECIPE PICKER */}
      {recipePickerOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <Card className="w-full max-w-lg max-h-[80dvh] flex flex-col overflow-hidden shadow-2xl border-border bg-background animate-in zoom-in-95 duration-200">
            <CardHeader className="border-b border-border/60 bg-muted/20 py-4 flex flex-row items-center justify-between shrink-0">
              <div>
                <CardTitle className="text-base font-bold flex items-center gap-2">
                  <Network className="size-5 text-violet-600" />
                  <span>{language === "th" ? "เลือกสูตรย่อย" : "Select Sub-recipe"}</span>
                </CardTitle>
                <CardDescription>
                  {language === "th" ? "อ้างอิงสูตรผลิตอื่นเป็น subset ในสูตรนี้" : "Reference another recipe as a subset."}
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 rounded-full"
                onClick={() => setRecipePickerOpen(false)}
              >
                <X className="size-4" />
              </Button>
            </CardHeader>
            <CardContent className="p-3 overflow-y-auto space-y-2">
              {recipeOptions.length === 0 ? (
                <div className="text-sm text-muted-foreground p-4 text-center">
                  {language === "th" ? "ยังไม่มีสูตรอื่นให้เลือก" : "No other recipes are available."}
                </div>
              ) : (
                recipeOptions.map((recipe) => {
                  const recipeCode = recipe.barcode || recipe.itemcode || "";
                  const recipeName = pickName(recipe.names, language) || recipeCode;
                  return (
                    <button
                      key={recipe.guidfixed}
                      type="button"
                      onClick={() => addRecipeToBOM(recipe)}
                      className="w-full rounded-lg border border-border bg-card p-3 text-left hover:bg-muted/40 transition-colors"
                    >
                      <div className="flex items-center justify-between gap-3">
                        <span className="font-semibold text-sm text-foreground">{recipeName}</span>
                        <Badge variant="secondary" className="font-mono text-[10px]">{recipeCode}</Badge>
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        {(recipe.bom || []).length} {language === "th" ? "รายการส่วนประกอบ" : "components"}
                      </div>
                    </button>
                  );
                })
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* DETAILED STRUCTURE TREE MODAL */}
      {detailModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <Card className="w-full max-w-4xl max-h-[85dvh] flex flex-col overflow-hidden shadow-2xl border-border bg-background animate-in zoom-in-95 duration-200">
            <CardHeader className="border-b border-border/60 bg-muted/20 py-4 flex flex-row items-center justify-between shrink-0">
              <div>
                <CardTitle className="text-base font-bold flex items-center gap-2">
                  <Network className="size-5 text-sky-600" />
                  <span>{language === "th" ? "โครงสร้างสูตรผลิตละเอียด (Exploded BOM)" : "Detailed BOM Structure"}</span>
                </CardTitle>
                <CardDescription>
                  {productName}{" "}
                  {selectedRecord.guidfixed.startsWith("virtual-")
                    ? (parentItemCode ? `(รหัสสูตร: ${parentItemCode})` : "")
                    : (selectedRecord.itemcode || selectedRecord.barcode ? `(รหัสสูตร: ${selectedRecord.itemcode || selectedRecord.barcode})` : "")}
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 rounded-full"
                onClick={() => setDetailModalOpen(false)}
              >
                <X className="size-4" />
              </Button>
            </CardHeader>

            <CardContent className="p-0 flex-1 overflow-y-auto scrollbar-thin flex flex-col">
              {/* Tree Table Header */}
              <div className="sticky top-0 z-10 flex items-center justify-between gap-4 py-2.5 px-4 bg-muted/80 backdrop-blur border-b border-border text-[11px] font-bold text-muted-foreground uppercase shrink-0">
                <span>{language === "th" ? "โครงสร้างและรหัสอ้างอิง" : "Component Hierarchy & Reference"}</span>
                <div className="flex items-center gap-6 shrink-0 font-semibold text-right">
                  <span className="w-28">{language === "th" ? "จำนวนสุทธิ" : "Quantity"}</span>
                  <span className="w-24">{language === "th" ? "ทุน/หน่วย" : "Cost/Unit"}</span>
                  <span className="w-24">{language === "th" ? "ราคารวม" : "Subtotal"}</span>
                  <span className="w-16">{language === "th" ? "สัดส่วน" : "Share"}</span>
                </div>
              </div>

              {/* Tree Content */}
              <div className="p-4 flex-1 divide-y divide-border/20">
                {bomItems.map((item, idx) => renderExplodedTreeNode(item, `root-${idx}`, 0))}
              </div>

              {/* Summary Footer */}
              <div className="border-t border-border/60 bg-muted/15 p-4 grid grid-cols-1 md:grid-cols-3 gap-4 shrink-0 text-sm">
                <div className="flex items-center justify-between p-2 rounded-lg bg-card border border-border/40">
                  <span className="text-muted-foreground">{language === "th" ? "ราคาขายของสูตร" : "Selling Price"}</span>
                  <span className="font-bold text-foreground font-mono">฿{salePrice.toFixed(2)}</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded-lg bg-card border border-border/40">
                  <span className="text-muted-foreground">{language === "th" ? "ต้นทุน Rollup รวม" : "Total Rollup Cost"}</span>
                  <span className="font-bold text-sky-600 font-mono">฿{rollupCost.toFixed(2)}</span>
                </div>
                <div className="flex items-center justify-between p-2 rounded-lg bg-card border border-border/40">
                  <span className="text-muted-foreground">{language === "th" ? "อัตรากำไรขั้นต้น" : "Profit Margin"}</span>
                  <span className={cn("font-bold font-mono", profitMarginPercent >= 40 ? "text-emerald-600" : profitMarginPercent >= 0 ? "text-amber-600" : "text-destructive")}>
                    {profitMarginPercent.toFixed(1)}%
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* INGREDIENT UNIT PICKER DIALOG */}
      {ingredientUnitOpen && ingredientProduct && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <Card className="w-full max-w-md flex flex-col overflow-hidden shadow-2xl border-border bg-background animate-in zoom-in-95 duration-200">
            <CardHeader className="border-b border-border/60 bg-muted/20 py-4 flex flex-row items-center justify-between shrink-0">
              <div>
                <CardTitle className="text-base font-bold flex items-center gap-2">
                  <span>{language === "th" ? "เลือกหน่วยวัตถุดิบ" : "Select Ingredient Unit"}</span>
                </CardTitle>
                <CardDescription>
                  {pickName(ingredientProduct.names, language) || ingredientProduct.code}
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 rounded-full"
                onClick={() => {
                  setIngredientUnitOpen(false);
                  setIngredientProduct(null);
                }}
              >
                <X className="size-4" />
              </Button>
            </CardHeader>
            <CardContent className="p-4 space-y-3">
              <p className="text-xs text-muted-foreground">
                {language === "th"
                  ? "เนื่องจากวัตถุดิบมีหลายหน่วยนับ กรุณาเลือกหน่วยนับที่เหมาะสมเพื่อนำมาใช้ในสูตรผลิตนี้"
                  : "This ingredient has multiple units. Please select the correct unit to use in this recipe."}
              </p>

              <div className="flex flex-col gap-2">
                {(() => {
                  const unitOptions = ingredientProduct.unitOptions || toProductUnitOptions(ingredientProduct);
                  return unitOptions.map((opt: ProductUnitOption) => {
                    const unitName = pickName(opt.itemunitnames, language) || opt.item_unit_code || "PCS";
                    return (
                      <button
                        key={opt.barcode}
                        type="button"
                        onClick={() => {
                          addIngredientToBOM(ingredientProduct, opt);
                          setIngredientUnitOpen(false);
                          setIngredientProduct(null);
                        }}
                        className="flex flex-col items-start gap-1 p-3 rounded-xl border border-border bg-background hover:bg-muted/40 text-left transition-all duration-200 shadow-sm"
                      >
                        <span className="text-xs font-bold text-foreground">{unitName} ({opt.item_unit_code})</span>
                        <span className="text-[10px] font-mono text-muted-foreground">
                          {language === "th" ? "บาร์โค้ดหน่วย: " : "Barcode: "}{opt.barcode}
                        </span>
                      </button>
                    );
                  });
                })()}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {confirmationDialog}
    </div>
  );
}

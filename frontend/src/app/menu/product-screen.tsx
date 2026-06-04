"use client";

import {
  AlertCircle,
  FolderOpen,
  Loader2,
  Package,
  Pencil,
  Plus,
  RefreshCcw,
  Search,
  Trash2,
  X,
  Save,
  PlusCircle,
  Link,
  Upload,
  ImagePlus,
  Minus,
  ChevronLeft,
  ChevronRight,
  Box,
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type Dispatch,
  type SetStateAction,
  type ChangeEvent,
  type ReactNode,
} from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { listBarcodes, uploadProductImage, type MasterEntry, type MasterName } from "@/lib/product-barcode/api";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import {
  localizedName,
  shopDisplayName,
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
  WORKSPACE_CHANGED_EVENT,
} from "@/lib/workspace-models";
import {
  type Product,
  type ProductBarcode,
  type ProductBarcodeListRow,
  type ProductManufacturer,
  type ProductSupplier,
  type NameX,
  type ProductImage,
  type ProductOption,
  type ProductChoice,
  type ProductRestaurant,
  type ProductOrderType,
  type ProductTimeForSale,
  type ProductBarcodeBusinessType,
  type ProductBarcodeBranch,
  type RefProductBarcode,
  type BOMProductBarcode,
} from "@/lib/product-barcode/types";
import { pickName, rawToProduct } from "@/lib/product-barcode/utils";
import { NamesEditor, languageCodesFromWorkspace } from "@/components/product-barcode/names-editor";
import { getBarcodeText } from "@/lib/product-barcode/language";

import { TabProductUnits } from "./tab-product-units";
function readAuthSession(): AuthSession | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(workspaceStorageKeys.auth);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function readWorkspaceSession(): WorkspaceSession | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(workspaceStorageKeys.workspace);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

async function ensureActiveProductHolding(auth: AuthSession, holding_code: string): Promise<void> {
  const response = await fetch("/api/workspace/select-holding", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-bc-backend-url": auth.backendUrl,
      Authorization: `Bearer ${auth.token}`,
    },
    body: JSON.stringify({ backendUrl: auth.backendUrl, holding_code }),
    cache: "no-store",
  });
  const data = await response.json().catch(() => null) as { success?: boolean; message?: string } | null;
  if (!response.ok || data?.success === false) {
    throw new Error(data?.message || "ไม่สามารถเลือกบริษัทใน token ได้");
  }
}

const PRODUCT_SPLIT_DEFAULT_LEFT = 30;
const PRODUCT_SPLIT_STORAGE_KEY = "bc_product_split_left_v3";
const PRODUCT_SPLIT_MIN_LEFT = 5;
const PRODUCT_SPLIT_MAX_LEFT = 95;

function clampProductSplitLeft(value: number) {
  if (!Number.isFinite(value)) return PRODUCT_SPLIT_DEFAULT_LEFT;
  return Math.min(PRODUCT_SPLIT_MAX_LEFT, Math.max(PRODUCT_SPLIT_MIN_LEFT, value));
}

export function ProductScreen({
  embedded = false,
  language = "th",
  isSetOnly = false,
}: {
  embedded?: boolean;
  language?: LanguageCode;
  isSetOnly?: boolean;
}) {
  const lang = normalizeLanguage(language);
  const text = getBarcodeText(lang);
  const { confirm, confirmationDialog } = useConfirmDialog();

  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [items, setItems] = useState<Product[]>([]);
  const [selectedCode, setSelectedCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState<{ type: "success" | "error" | "info"; text: string } | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  // Resizable split states
  const [splitLeftPercent, setSplitLeftPercent] = useState(PRODUCT_SPLIT_DEFAULT_LEFT);
  const [resizingSplit, setResizingSplit] = useState(false);
  const splitContainerRef = useRef<HTMLDivElement | null>(null);
  const selectedShopTokenRef = useRef("");

  // Restore split settings
  useEffect(() => {
    if (typeof window === "undefined") return;
    const saved = window.localStorage.getItem(PRODUCT_SPLIT_STORAGE_KEY);
    if (saved) {
      const next = clampProductSplitLeft(Number(saved));
      setSplitLeftPercent(next);
    }
  }, []);

  const productSplitStyle = useMemo(
    () =>
      ({
        "--product-list-fr": `${splitLeftPercent}fr`,
        "--product-detail-fr": `${100 - splitLeftPercent}fr`,
      } as React.CSSProperties),
    [splitLeftPercent],
  );

  const updateSplitFromClientX = useCallback((clientX: number) => {
    const container = splitContainerRef.current;
    if (!container) return;
    const rect = container.getBoundingClientRect();
    if (rect.width <= 0) return;
    const offset = clientX - rect.left;
    const next = (offset / rect.width) * 100;
    setSplitLeftPercent(clampProductSplitLeft(next));
  }, []);

  useEffect(() => {
    if (!resizingSplit) return;
    const onPointerMove = (event: PointerEvent) => {
      updateSplitFromClientX(event.clientX);
    };
    const onPointerUp = () => {
      setResizingSplit(false);
    };
    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", onPointerUp);
    return () => {
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", onPointerUp);
    };
  }, [resizingSplit, updateSplitFromClientX]);

  const startSplitResize = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      setResizingSplit(true);
      updateSplitFromClientX(event.clientX);
    },
    [updateSplitFromClientX],
  );

  const startSplitMouseResize = useCallback(
    (event: React.MouseEvent<HTMLDivElement>) => {
      setResizingSplit(true);
      updateSplitFromClientX(event.clientX);
    },
    [updateSplitFromClientX],
  );

  const moveSplitResize = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      if (!resizingSplit) return;
      updateSplitFromClientX(event.clientX);
    },
    [resizingSplit, updateSplitFromClientX],
  );

  const stopSplitResize = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
    event.currentTarget.releasePointerCapture(event.pointerId);
    setResizingSplit(false);
    setSplitLeftPercent((current) => {
      const next = clampProductSplitLeft(current);
      if (typeof window !== "undefined") {
        window.localStorage.setItem(PRODUCT_SPLIT_STORAGE_KEY, String(Math.round(next)));
      }
      return next;
    });
  }, []);

  const adjustSplitWithKeyboard = useCallback((event: React.KeyboardEvent<HTMLDivElement>) => {
    let direction = 0;
    if (event.key === "ArrowLeft") direction = -2;
    else if (event.key === "ArrowRight") direction = 2;
    if (direction === 0) return;
    event.preventDefault();
    setSplitLeftPercent((current) => clampProductSplitLeft(current + direction));
  }, []);

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(searchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchInput]);

  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<"create" | "edit">("create");
  const [editProduct, setEditProduct] = useState<Product | null>(null);
  const [saving, setSaving] = useState(false);
  /** Flag: when true, successful save stays open and resets to blank instead of closing. */
  const saveAndNewPending = useRef(false);
  /** Ref to the product editor form element for programmatic submission. */
  const productFormRef = useRef<HTMLFormElement>(null);
  const pickerAnchorRef = useRef<HTMLElement | null>(null);
  const [productTab, setProductTab] = useState<"basic" | "classification" | "units" | "bom" | "stock" | "media" | "logistics" | "restaurant" | "timeforsales" | "business" | "misc">("basic");


  const itemTypes = useMemo(() => [
    { value: 0, label: text.itemTypeStock },
    { value: 1, label: text.itemTypeService },
  ], [text]);

  const vatTypes = useMemo(() => [
    { value: 0, label: text.vatIncluded },
    { value: 1, label: text.vatExcluded },
  ], [text]);

  const materialTypes = useMemo(() => [
    { value: 0, label: text.materialGeneral },
    { value: 1, label: text.materialMaterial },
    { value: 2, label: text.materialSemiFinished },
    { value: 3, label: text.materialSet },
    { value: 4, label: text.materialAgricultural },
  ], [text]);

  const sumPointTypes = useMemo(() => [
    { value: "true", label: text.isSumPointYes },
    { value: "false", label: text.isSumPointNo },
  ], [text]);

  const sumPointOptions = useMemo(() => [
    { value: true, label: text.isSumPointYes },
    { value: false, label: text.isSumPointNo },
  ], [text]);

  const foodTypes = useMemo(() => [
    { value: 0, label: text.foodTypeFood },
    { value: 1, label: text.foodTypeDrink },
    { value: 2, label: text.foodTypeAlcohol },
    { value: 3, label: text.foodTypeOther },
  ], [text]);

  const [pickerOpen, setPickerOpen] = useState(false);

const [pickerType, setPickerType] = useState<string>("");
  const [pickerTarget, setPickerTarget] = useState<string>("");

  // Unlinked Barcode Picker states
  const [showBarcodePicker, setShowBarcodePicker] = useState(false);
  const [barcodeSearchInput, setBarcodeSearchInput] = useState("");
  const [barcodeSearch, setBarcodeSearch] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setBarcodeSearch(barcodeSearchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [barcodeSearchInput]);
  const [barcodeList, setBarcodeList] = useState<ProductBarcodeListRow[]>([]);
  const [loadingBarcodes, setLoadingBarcodes] = useState(false);
  const [bindBarcodeOnSave, setBindBarcodeOnSave] = useState<ProductBarcode | null>(null);

  useEffect(() => {
    setAuth(readAuthSession());
    setWorkspace(readWorkspaceSession());
  }, []);

  useEffect(() => {
    const handleWorkspaceChange = () => {
      const nextWorkspace = readWorkspaceSession();
      if (nextWorkspace) setWorkspace(nextWorkspace);
    };
    window.addEventListener(WORKSPACE_CHANGED_EVENT, handleWorkspaceChange);
    window.addEventListener("storage", handleWorkspaceChange);
    return () => {
      window.removeEventListener(WORKSPACE_CHANGED_EVENT, handleWorkspaceChange);
      window.removeEventListener("storage", handleWorkspaceChange);
    };
  }, []);

  const activeHoldingCode = workspace?.shop.holding_code ?? "";

  useEffect(() => {
    if (!showBarcodePicker || !auth) return;
    let active = true;
    setLoadingBarcodes(true);
    listBarcodes(auth, { holding_code: activeHoldingCode, keyword: barcodeSearch, limit: 100 })
      .then((resData) => {
        if (active && resData.success && resData.data) {
          // Filter unlinked barcodes (itemcode is empty — no product linked yet)
          const unlinked = resData.data.filter((b) => !b.itemcode);
          setBarcodeList(unlinked);
        }
      })
      .catch((err) => console.error("Error loading unlinked barcodes", err))
      .finally(() => {
        if (active) setLoadingBarcodes(false);
      });

    return () => {
      active = false;
    };
  }, [showBarcodePicker, auth, activeHoldingCode, barcodeSearch]);

  const activeLanguages = useMemo(() => languageCodesFromWorkspace(workspace), [workspace]);

  const selectedProduct = useMemo(() => {
    return items.find((item) => item.code === selectedCode) ?? items[0] ?? null;
  }, [items, selectedCode]);

  const isFormDirty = useMemo(() => {
    if (!editorOpen || !editProduct || !selectedProduct) return false;
    // Compare basic fields for dirty detection
    return JSON.stringify(editProduct) !== JSON.stringify(selectedProduct);
  }, [editorOpen, editProduct, selectedProduct]);

  const handleSelectProduct = useCallback((code: string) => {
    if (isFormDirty) {
      void confirm({
        title: "ข้อมูลมีการเปลี่ยนแปลง",
        description: "คุณมีข้อมูลที่ยังไม่ได้บันทึก ต้องการละทิ้งการเปลี่ยนแปลงแล้วเปลี่ยนสินค้าหรือไม่?",
        confirmLabel: "เปลี่ยนสินค้า",
        cancelLabel: "ยกเลิก",
      }).then((ok) => {
        if (ok) {
          setSelectedCode(code);
          setEditorOpen(false);
        }
      });
    } else {
      setSelectedCode(code);
      setEditorOpen(false);
    }
  }, [isFormDirty, confirm]);

  const handleCancelEdit = useCallback(() => {
    if (isFormDirty) {
      void confirm({
        title: "ข้อมูลมีการเปลี่ยนแปลง",
        description: "คุณมีข้อมูลที่ยังไม่ได้บันทึก ต้องการยกเลิกการแก้ไขใช่หรือไม่?",
        confirmLabel: "ใช่, ยกเลิก",
        cancelLabel: "กลับไปแก้ไข",
      }).then((ok) => {
        if (ok) {
          setEditorOpen(false);
        }
      });
    } else {
      setEditorOpen(false);
    }
  }, [isFormDirty, confirm]);

  const loadProducts = useCallback(async () => {
    if (!auth || !activeHoldingCode) return;
    setLoading(true);
    setNotice(null);
    try {
      const tokenShopKey = `${auth.token}:${activeHoldingCode}`;
      if (selectedShopTokenRef.current !== tokenShopKey) {
        await ensureActiveProductHolding(auth, activeHoldingCode);
        selectedShopTokenRef.current = tokenShopKey;
      }
      const params = new URLSearchParams({
        q: search,
        limit: isSetOnly ? "120" : "80",
      });
      if (isSetOnly) {
        params.set("item_type", "2");
        params.set("materialtype", "3");
      }
      const response = await fetch(`/api/product?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${auth.token}`,
          "x-bc-backend-url": auth.backendUrl,
        },
      });
      const data = await response.json();
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text.requestFailed);
      }
      const rawData = Array.isArray(data.data) ? data.data : [];
      const normalized: Product[] = rawData.map(rawToProduct);
      const filtered = isSetOnly
        ? normalized.filter((item) => item.item_type === 2)
        : normalized.filter((item) => item.item_type !== 2);

      setItems(filtered);
      if (filtered.length > 0) {
        // Only auto-select first item on md+ screens; on mobile the list stays visible
        const isDesktop = typeof window !== "undefined" && window.matchMedia("(min-width: 768px)").matches;
        if (isDesktop) {
          setSelectedCode((prev) => prev || filtered[0].code);
        }
      }
    } catch (err) {
      setItems([]);
      setNotice({ type: "error", text: err instanceof Error ? err.message : text.requestFailed });
    } finally {
      setLoading(false);
    }
  }, [auth, activeHoldingCode, search, text.requestFailed, isSetOnly]);

  useEffect(() => {
    if (auth && activeHoldingCode) {
      void loadProducts();
    }
  }, [auth, activeHoldingCode, loadProducts]);

  const makeBlankProduct = useCallback((): Product => ({
    guidfixed: "",
    holding_code: activeHoldingCode,
    code: "",
    names: [{ code: "th", name: "" }, { code: "en", name: "" }],
    group_code: "",
    group_names: [],
    item_type: isSetOnly ? 2 : 0,
    vat_type: 0,
    materialtype: isSetOnly ? 3 : 0,
    issumpoint: false,
    manufacturers: [],
    suppliers: [],
    condition: false,
    dividevalue: 1,
    standvalue: 1,
    isusesubbarcodes: false,
    refbarcodes: [],
    bom: [],
    orderpoint: 0,
    minpoint: 0,
    maxpoint: 0,
    qty: 0,
    stockbarcode: "",
  }), [activeHoldingCode, isSetOnly]);

  const handleCreateOpen = () => {
    setEditorMode("create");
    setBindBarcodeOnSave(null);
    setEditProduct(makeBlankProduct());
    setEditorOpen(true);
  };

  const handleSelectBarcode = async (barcodeRow: ProductBarcodeListRow) => {
    if (!auth) return;
    try {
      setLoading(true);
      const res = await fetch(`/api/product-barcode/${encodeURIComponent(barcodeRow.guidfixed)}`, {
        headers: {
          Authorization: `Bearer ${auth.token}`,
          "x-bc-backend-url": auth.backendUrl,
        },
      });
      const resData = await res.json();
      if (resData.success && resData.data) {
        const b = resData.data as ProductBarcode;
        setEditProduct((current) => {
          if (!current) return null;
          return {
            ...current,
            names: b.names && b.names.length > 0 ? b.names : current.names,
            code: b.itemcode || b.barcode || current.code,
            group_code: b.group_code || "",
            group_names: b.group_names || [],
            groupsuboneguid: b.groupsuboneguid || "",
            groupsubonecode: b.groupsubonecode || "",
            groupsubonenames: b.groupsubonenames || [],
            groupsubtwoguid: b.groupsubtwoguid || "",
            groupsubtwocode: b.groupsubtwocode || "",
            groupsubtwonames: b.groupsubtwonames || [],
            brandguid: b.brandguid || "",
            brand_code: b.brand_code || "",
            brandnames: b.brandnames || [],
            category_guid: b.category_guid || "",
            categorycode: b.categorycode || "",
            category_names: b.category_names || [],
            classguid: b.classguid || "",
            classcode: b.classcode || "",
            classnames: b.classnames || [],
            designguid: b.designguid || "",
            designcode: b.designcode || "",
            designnames: b.designnames || [],
            modelguid: b.modelguid || "",
            modelcode: b.modelcode || "",
            modelnames: b.modelnames || [],
            patternguid: b.patternguid || "",
            patterncode: b.patterncode || "",
            patternnames: b.patternnames || [],
            gradeguid: b.gradeguid || "",
            gradecode: b.gradecode || "",
            gradenames: b.gradenames || [],
            vat_type: b.vat_type ?? 0,
            item_type: b.item_type ?? 0,
            materialtype: b.materialtype ?? 0,
            tax_type: b.tax_type ?? 0,
            manufacturers: b.manufacturers || [],
            suppliers: b.suppliers || [],
            condition: b.condition ?? false,
            dividevalue: b.dividevalue ?? 1,
            standvalue: b.standvalue ?? 1,
            isusesubbarcodes: b.isusesubbarcodes ?? false,
            refbarcodes: b.refbarcodes || [],
            bom: b.bom || [],
            orderpoint: b.orderpoint ?? 0,
            minpoint: b.minpoint ?? 0,
            maxpoint: b.maxpoint ?? 0,
            qty: b.qty ?? 0,
            stockbarcode: b.stockbarcode || "",
          };
        });
        setBindBarcodeOnSave(b);
        setNotice({ type: "info", text: text.barcodeFetchSuccess.replace("%s", b.barcode) });
      }
    } catch (err) {
      console.error(err);
      setNotice({ type: "error", text: text.barcodeFetchFail });
    } finally {
      setLoading(false);
      setShowBarcodePicker(false);
    }
  };

  const handleEditOpen = (p: Product) => {
    setEditorMode("edit");
    setEditProduct({ ...p });
    setEditorOpen(true);
  };

  const handleDelete = async (p: Product) => {
    if (!auth || !p.guidfixed) return;
    const ok = await confirm({
      title: text.deleteConfirm,
      description: `${text.itemCode}: ${p.code}`,
      tone: "danger",
      confirmLabel: text.delete,
      cancelLabel: text.cancel,
    });
    if (!ok) return;

    try {
      const res = await fetch(`/api/product/${encodeURIComponent(p.guidfixed)}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${auth.token}`,
          "x-bc-backend-url": auth.backendUrl,
        },
      });
      const data = await res.json();
      if (!res.ok || data.success === false) {
        throw new Error(data.message || "Delete failed");
      }
      setNotice({ type: "success", text: text.deleteSuccess });
      void loadProducts();
      setSelectedCode("");
    } catch (err) {
      setNotice({ type: "error", text: err instanceof Error ? err.message : "Delete failed" });
    }
  };

  const handleSave = async (e: FormEvent) => {
    e.preventDefault();
    // Capture & reset immediately so a validation/exception early-return can't leak the flag into the next save.
    const saveAndNew = saveAndNewPending.current;
    saveAndNewPending.current = false;
    if (!auth || !editProduct) return;
    if (!editProduct.code.trim()) {
      setNotice({ type: "error", text: text.codeRequired });
      return;
    }
    const hasName = editProduct.names.some((n) => n.name && n.name.trim() !== "");
    if (!hasName) {
      setNotice({ type: "error", text: text.nameRequired });
      return;
    }

    setSaving(true);
    setNotice(null);
    try {
      const method = editorMode === "create" ? "POST" : "PUT";
      const url = editorMode === "create" ? "/api/product" : `/api/product/${encodeURIComponent(editProduct.guidfixed)}`;

      const payload = {
        ...editProduct,
        dividevalue: 1,
        standvalue: 1,
        condition: false,
      };

      const res = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${auth.token}`,
          "x-bc-backend-url": auth.backendUrl,
        },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (!res.ok || data.success === false) {
        throw new Error(data.message || "Save failed");
      }

      // Link Barcode if selected
      if (editorMode === "create" && bindBarcodeOnSave && data.data?.guidfixed) {
        const createdGuid = data.data.guidfixed;
        const createdCode = editProduct.code;

        const bcRes = await fetch(`/api/product-barcode/${encodeURIComponent(bindBarcodeOnSave.guidfixed)}`, {
          headers: {
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
        });
        const bcJson = await bcRes.json();
        if (!bcRes.ok || !bcJson.success || !bcJson.data) {
          throw new Error(bcJson.message || "Failed to fetch barcode details for linking");
        }

        const fullBarcode = bcJson.data;
        fullBarcode.item_guid = createdGuid;
        fullBarcode.itemcode = createdCode;

        const putRes = await fetch(`/api/product-barcode/${encodeURIComponent(bindBarcodeOnSave.guidfixed)}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
          body: JSON.stringify({
            backendUrl: auth.backendUrl,
            data: fullBarcode,
          }),
        });
        const putJson = await putRes.json();
        if (!putRes.ok || !putJson.success) {
          throw new Error(putJson.message || "Failed to save barcode link");
        }
      }

      setNotice({ type: "success", text: text.saveSuccess });
      void loadProducts();
      if (saveAndNew) {
        // "Save & add new": reset to blank and stay open
        setBindBarcodeOnSave(null);
        setProductTab("basic");
        setEditProduct(makeBlankProduct());
      } else {
        setEditorOpen(false);
        setBindBarcodeOnSave(null);
        if (editorMode === "create" && data.data?.code) {
          setSelectedCode(data.data.code);
        }
      }
    } catch (err) {
      setNotice({ type: "error", text: err instanceof Error ? err.message : "Save failed" });
    } finally {
      setSaving(false);
    }
  };

  const openPicker = (type: string, targetField: string, anchorEl: HTMLElement | null = null) => {
    pickerAnchorRef.current = anchorEl;
    setPickerType(type);
    setPickerTarget(targetField);
    setPickerOpen(true);
  };

  const handlePickerSelect = (entry: MasterEntry) => {
    if (!editProduct) return;

    // Check if picker target is unit-ref-IDX
    if (pickerTarget.startsWith("unit-ref-")) {
      const idx = parseInt(pickerTarget.split("-")[2], 10);
      const current = editProduct.refbarcodes || [];
      setEditProduct({
        ...editProduct,
        refbarcodes: current.map((row, rowIdx) =>
          rowIdx === idx
            ? {
                ...row,
                item_unit_code: entry.code,
                itemunitnames: entry.names,
              }
            : row
        ),
      });
      return;
    }

    // Check if picker type is creditor (used for manufacturer and supplier multi-select)
    if (pickerType === "creditor") {
      if (pickerTarget === "manufacturers") {
        const current = editProduct.manufacturers || [];
        if (!current.some((m) => m.guid_fixed === entry.guidfixed)) {
          setEditProduct({
            ...editProduct,
            manufacturers: [...current, { guid_fixed: entry.guidfixed, code: entry.code, names: entry.names }],
          });
        }
      } else if (pickerTarget === "suppliers") {
        const current = editProduct.suppliers || [];
        if (!current.some((s) => s.guid_fixed === entry.guidfixed)) {
          setEditProduct({
            ...editProduct,
            suppliers: [...current, { guid_fixed: entry.guidfixed, code: entry.code, names: entry.names }],
          });
        }
      }
      return;
    }

    // Unit select mapping
    if (pickerTarget === "unit") {
      setEditProduct({
        ...editProduct,
        unitguid: entry.guidfixed,
        unitcode: entry.code,
        unitnames: entry.names,
        item_unit_code: entry.code,
        itemunitnames: entry.names,
      });
      return;
    }

    // Single select classification pickers — explicit field map to avoid unsafe keyof cast
    const classificationFields: Record<string, { guid: keyof Product; code: keyof Product; names: keyof Product }> = {
      group:        { guid: "groupguid" as keyof Product,        code: "group_code",       names: "group_names" },
      groupsubone:  { guid: "groupsuboneguid" as keyof Product,  code: "groupsubonecode",  names: "groupsubonenames" },
      groupsubtwo:  { guid: "groupsubtwoguid" as keyof Product,  code: "groupsubtwocode",  names: "groupsubtwonames" },
      brand:        { guid: "brandguid" as keyof Product,        code: "brand_code",       names: "brandnames" },
      category:     { guid: "category_guid" as keyof Product,    code: "categorycode",     names: "category_names" },
      class:        { guid: "classguid" as keyof Product,        code: "classcode",        names: "classnames" },
      design:       { guid: "designguid" as keyof Product,       code: "designcode",       names: "designnames" },
      model:        { guid: "modelguid" as keyof Product,        code: "modelcode",        names: "modelnames" },
      pattern:      { guid: "patternguid" as keyof Product,      code: "patterncode",      names: "patternnames" },
      grade:        { guid: "gradeguid" as keyof Product,        code: "gradecode",        names: "gradenames" },
    };
    const cf = classificationFields[pickerTarget];
    if (!cf) return;
    setEditProduct({
      ...editProduct,
      [cf.guid]: entry.guidfixed,
      [cf.code]: entry.code,
      [cf.names]: entry.names,
    });
  };

  const clearPickerField = (field: string) => {
    if (!editProduct) return;
    if (field === "unit") {
      setEditProduct({
        ...editProduct,
        unitguid: "",
        unitcode: "",
        unitnames: [],
        item_unit_code: "",
        itemunitnames: [],
      });
      return;
    }
    const clearFields: Record<string, { guid: keyof Product; code: keyof Product; names: keyof Product }> = {
      group:        { guid: "groupguid" as keyof Product,        code: "group_code",       names: "group_names" },
      groupsubone:  { guid: "groupsuboneguid" as keyof Product,  code: "groupsubonecode",  names: "groupsubonenames" },
      groupsubtwo:  { guid: "groupsubtwoguid" as keyof Product,  code: "groupsubtwocode",  names: "groupsubtwonames" },
      brand:        { guid: "brandguid" as keyof Product,        code: "brand_code",       names: "brandnames" },
      category:     { guid: "category_guid" as keyof Product,    code: "categorycode",     names: "category_names" },
      class:        { guid: "classguid" as keyof Product,        code: "classcode",        names: "classnames" },
      design:       { guid: "designguid" as keyof Product,       code: "designcode",       names: "designnames" },
      model:        { guid: "modelguid" as keyof Product,        code: "modelcode",        names: "modelnames" },
      pattern:      { guid: "patternguid" as keyof Product,      code: "patterncode",      names: "patternnames" },
      grade:        { guid: "gradeguid" as keyof Product,        code: "gradecode",        names: "gradenames" },
    };
    const cf = clearFields[field];
    if (!cf) return;
    setEditProduct({
      ...editProduct,
      [cf.guid]: "",
      [cf.code]: "",
      [cf.names]: [],
    });
  };

  const removeManufacturer = (guid: string) => {
    if (!editProduct) return;
    const current = editProduct.manufacturers || [];
    setEditProduct({
      ...editProduct,
      manufacturers: current.filter((m) => m.guid_fixed !== guid),
    });
  };

  const removeSupplier = (guid: string) => {
    if (!editProduct) return;
    const current = editProduct.suppliers || [];
    setEditProduct({
      ...editProduct,
      suppliers: current.filter((s) => s.guid_fixed !== guid),
    });
  };

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-background">
      {/* Header Toolbar */}
      <div className="flex shrink-0 items-center justify-between border-b border-border bg-card px-4 py-3">
        <div>
          <h2 className="text-lg font-bold">{isSetOnly ? "สินค้าชุด" : text.title}</h2>
          <p className="text-xs text-muted-foreground">{isSetOnly ? "จัดการข้อมูลสินค้าชุดและส่วนประกอบทั้งหมด" : text.subtitle}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => void loadProducts()} disabled={loading}>
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
            {text.refresh}
          </Button>
          <Button variant="default" size="sm" onClick={handleCreateOpen}>
            <Plus className="h-4 w-4" />
            {text.add}
          </Button>
        </div>
      </div>

      {notice && (
        <div className={cn(
          "px-4 py-2 text-sm flex items-center gap-2 rounded-md border",
          notice.type === "success"
            ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20"
            : notice.type === "info"
              ? "bg-blue-500/10 text-blue-500 border-blue-500/20"
              : "bg-destructive/10 text-destructive border-destructive/20"
        )}>
          <AlertCircle className="h-4 w-4" />
          <span>{notice.text}</span>
        </div>
      )}

      {/* Main split layout */}
      <div
        ref={splitContainerRef}
        className={cn(
          "grid min-h-0 min-w-0 gap-3 xl:grid-cols-[minmax(0,var(--product-list-fr))_8px_minmax(0,var(--product-detail-fr))] xl:gap-0 flex-1 xl:h-full flex-col xl:flex-row",
        )}
        style={productSplitStyle}
      >
        {/* Left Side: Product List */}
        <Card className={cn(
          "min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col border-b xl:border-b-0 xl:border-r border-border bg-muted/10 rounded-none border-y-0 border-l-0 shadow-none bg-card",
          selectedCode && !editorOpen ? "hidden xl:flex" : "flex"
        )}>
          <div className="p-3 border-b border-border">
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder={text.search}
                className="pl-8"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
              />
            </div>
          </div>
          <div className="bc-list-toolbar shrink-0">
            <span>{isSetOnly ? "สินค้าชุดทั้งหมด" : "สินค้าทั้งหมด"}</span>
            <span>{items.length} รายการ</span>
          </div>

          {/* Table Header inside list on Desktop */}
          <div className="bc-list-header hidden lg:grid grid-cols-[minmax(90px,1.2fr)_minmax(150px,2.8fr)_minmax(80px,1fr)_72px] gap-x-3 shrink-0">
            <span>{text.itemCode ?? "รหัสสินค้า"}</span>
            <span>{text.productName ?? "ชื่อสินค้า"}</span>
            <span>{text.itemType ?? "ประเภท"}</span>
            <span className="text-center">จัดการ</span>
          </div>

          <div className="flex-1 overflow-y-auto min-h-[360px] xl:h-full xl:min-h-0">
            {loading ? (
              <div className="p-8 text-center text-sm text-muted-foreground flex justify-center items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                {text.loading}
              </div>
            ) : items.length === 0 ? (
              <div className="p-8 text-center text-sm text-muted-foreground">{text.noData}</div>
            ) : (
              items.map((item, index) => {
                const active = item.code === selectedCode;
                const isEditing = active && editorOpen && editorMode === "edit";
                const typeLabel = item.item_type === 2
                  ? text.itemTypeSet
                  : (itemTypes.find((t) => t.value === item.item_type)?.label ?? String(item.item_type));
                return (
                  <div
                    key={item.guidfixed || item.code}
                    className={cn(
                      "bc-list-row grid lg:grid-cols-[minmax(90px,1.2fr)_minmax(150px,2.8fr)_minmax(80px,1fr)_72px] gap-x-3",
                      isEditing
                        ? "bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-amber-200/50"
                        : active
                          ? "bg-primary/10 hover:bg-primary/15"
                          : index % 2 === 0
                            ? "bg-background hover:bg-primary/5"
                            : "bg-muted/10 hover:bg-primary/5",
                    )}
                    onClick={() => handleSelectProduct(item.code)}
                  >
                    <div className="min-w-0">
                      <span className="lg:hidden text-[10px] font-semibold text-muted-foreground block">{text.itemCode ?? "รหัสสินค้า"}</span>
                      <div className={cn("truncate font-medium text-foreground", isEditing && "text-amber-950 dark:text-amber-100")}>{item.code || "-"}</div>
                    </div>
                    <div className="min-w-0">
                      <span className="lg:hidden text-[10px] font-semibold text-muted-foreground block">{text.productName ?? "ชื่อสินค้า"}</span>
                      <div className="line-clamp-2 font-medium">{pickName(item.names, lang) || "-"}</div>
                    </div>
                    <div className="min-w-0">
                      <span className="lg:hidden text-[10px] font-semibold text-muted-foreground block">{text.itemType ?? "ประเภท"}</span>
                      <div className={cn("truncate text-muted-foreground", isEditing && "text-amber-900/60 dark:text-amber-200/60")}>{typeLabel}</div>
                    </div>
                    <div className="flex items-center gap-1.5 justify-start lg:justify-center" onClick={(e) => e.stopPropagation()}>
                      <Button
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-lg bg-background text-primary hover:bg-primary/10 border-border"
                        onClick={() => handleEditOpen(item)}
                        title={text.edit}
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                      <Button
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-lg bg-background text-destructive hover:bg-destructive/10 border-border"
                        onClick={() => void handleDelete(item)}
                        title={text.delete}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </Card>

        {/* Resizable split separator bar */}
        <div
          aria-label={text.resizeAriaLabel ?? "Adjust layout split"}
          aria-orientation="vertical"
          aria-valuemax={PRODUCT_SPLIT_MAX_LEFT}
          aria-valuemin={PRODUCT_SPLIT_MIN_LEFT}
          aria-valuenow={Math.round(splitLeftPercent)}
          className={cn(
            "group hidden cursor-col-resize touch-none items-stretch justify-center rounded-md outline-none xl:flex",
            resizingSplit && "cursor-col-resize",
          )}
          onKeyDown={adjustSplitWithKeyboard}
          onMouseDown={startSplitMouseResize}
          onPointerCancel={stopSplitResize}
          onPointerDown={startSplitResize}
          onPointerMove={moveSplitResize}
          onPointerUp={stopSplitResize}
          role="separator"
          tabIndex={0}
        >
          <div
            className={cn(
              "my-1 w-1 rounded-full bg-border transition-colors group-hover:bg-primary group-focus-visible:bg-primary",
              resizingSplit && "bg-primary",
            )}
          />
        </div>

        {/* Right Side: Detail or Editor */}
        <div className="flex-1 overflow-y-auto bg-background min-h-0 p-4 xl:h-full xl:min-h-0">
          {selectedCode && !editorOpen && (
            <Button
              variant="ghost"
              size="sm"
              className="mb-4 xl:hidden flex items-center gap-2"
              onClick={() => setSelectedCode("")}
            >
              <ChevronLeft className="h-4 w-4" />
              {text.backToList}
            </Button>
          )}
          {editorOpen && editProduct ? (
            /* Product Edit Form */
            <form ref={productFormRef} onSubmit={handleSave} className="max-w-4xl mx-auto space-y-6">
              <div className="flex items-center justify-between border-b border-border pb-3">
                <div className="flex items-center gap-3">
                  <h3 className="text-xl font-bold">
                    {editorMode === "create"
                      ? (isSetOnly ? "เพิ่มสินค้าชุดใหม่" : text.productMasterCreateTitle)
                      : (isSetOnly ? "แก้ไขสินค้าชุด" : text.productMasterEditTitle)}
                  </h3>
                  {editorMode === "create" && (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="text-primary hover:text-primary-foreground border-primary hover:bg-primary gap-1"
                      onClick={() => {
                        setBarcodeSearchInput("");
                        setBarcodeSearch("");
                        setBarcodeList([]);
                        setShowBarcodePicker(true);
                      }}
                    >
                      <Link className="h-3.5 w-3.5" />
                      {text.pullFromBarcode}
                    </Button>
                  )}
                </div>
                <div className="flex gap-2">
                  <Button type="button" variant="outline" size="sm" onClick={handleCancelEdit} disabled={saving}>
                    {text.cancel}
                  </Button>
                  {editorMode === "create" && (
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      disabled={saving}
                      onClick={() => {
                        saveAndNewPending.current = true;
                        productFormRef.current?.requestSubmit();
                      }}
                    >
                      {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
                      {text.saveAndNew}
                    </Button>
                  )}
                  <Button type="submit" size="sm" disabled={saving}>
                    {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                    {text.save}
                  </Button>
                </div>
              </div>

              {/* Product Form Tab bar */}
              <div className="border-b border-border bg-muted/30 -mx-4 px-4 py-1.5 flex flex-wrap gap-1">
                {(Object.entries({
                  basic: text.tabBasic,
                  classification: text.tabClassification,
                  units: text.tabUnitsBarcode,
                  bom: text.tabBomShort,
                  stock: text.tabStock,
                  media: text.tabMedia,
                  logistics: "การจัดส่ง / โลจิสติกส์",
                  restaurant: isSetOnly ? "ตัวเลือกสินค้าชุด" : text.tabRestaurant,
                  timeforsales: text.tabTimeForSales,
                  business: text.tabBusinessBranchShort,
                  misc: text.tabMisc,
                }) as ["basic" | "classification" | "units" | "bom" | "stock" | "media" | "logistics" | "restaurant" | "timeforsales" | "business" | "misc", string][]).map(([k, label]) => {
                  const isActive = productTab === k;
                  return (
                    <button
                      key={k}
                      type="button"
                      onClick={() => setProductTab(k)}
                      className={cn(
                        "rounded-md px-3 py-1.5 text-xs font-medium transition",
                        isActive
                          ? "bg-primary text-primary-foreground shadow-sm"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground"
                      )}
                    >
                      {label}
                    </button>
                  );
                })}
              </div>

              {/* Tab: basic */}
              {productTab === "basic" && (
                <div className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-1">
                    <div className="space-y-2">
                      <label className="text-sm font-semibold">{text.itemCode} *</label>
                      <Input
                        required
                        disabled={editorMode === "edit"}
                        value={editProduct.code}
                        onChange={(e) => setEditProduct({ ...editProduct, code: e.target.value.toUpperCase() })}
                      />
                    </div>
                  </div>

                  {/* Localized Names */}
                  <div className="border border-border rounded-lg p-4">
                    <NamesEditor
                      names={editProduct.names || []}
                      onChange={(nextNames) => setEditProduct({ ...editProduct, names: nextNames })}
                      languages={activeLanguages}
                      label={text.productName}
                      firstRequired
                      language={lang}
                    />
                  </div>

                  {/* ประเภทสินค้าหลัก (Radio Groups) */}
                  {!isSetOnly && (
                    <div className="border border-border bg-card rounded-lg p-4 space-y-4 shadow-sm">
                      <h4 className="font-bold text-base text-foreground border-b border-border pb-2">ประเภทสินค้าหลัก</h4>
                      <div className="grid gap-6 md:grid-cols-2">
                        <RadioOptionGroup
                          label={text.itemTypeLabel}
                          value={editProduct.item_type ?? 0}
                          onChange={(n) => setEditProduct({ ...editProduct, item_type: n })}
                          options={itemTypes}
                        />
                        <RadioOptionGroup
                          label={text.materialTypeLabel}
                          value={editProduct.materialtype ?? 0}
                          onChange={(n) => setEditProduct({ ...editProduct, materialtype: n })}
                          options={materialTypes}
                        />
                      </div>
                    </div>
                  )}

                  {/* การตั้งค่าภาษี (Radio Groups) */}
                  <div className="border border-border bg-card rounded-lg p-4 space-y-4 shadow-sm">
                    <h4 className="font-bold text-base text-foreground border-b border-border pb-2">การตั้งค่าภาษี</h4>
                    <div className="grid gap-6 md:grid-cols-1">
                      <RadioOptionGroup
                        label="ประเภทภาษี"
                        value={editProduct.vat_type ?? 0}
                        onChange={(n) => setEditProduct({ ...editProduct, vat_type: n })}
                        options={vatTypes}
                      />
                    </div>
                  </div>

                  {/* Manufacturers & Suppliers section */}
                  <div className="grid gap-4 md:grid-cols-2">
                    {/* Manufacturers List (Multi Select) */}
                    <div className="border border-border rounded-lg p-4 space-y-3">
                      <div className="flex items-center justify-between border-b border-border pb-1">
                        <h4 className="font-semibold text-sm">{text.manufacturers}</h4>
                        <Button type="button" size="sm" variant="outline" onClick={(e) => openPicker("creditor", "manufacturers", e.currentTarget)}>
                          <PlusCircle className="mr-1 h-3.5 w-3.5" />
                          {text.addManufacturer}
                        </Button>
                      </div>
                      <div className="space-y-2">
                        {(!editProduct.manufacturers || editProduct.manufacturers.length === 0) ? (
                          <p className="text-xs text-muted-foreground text-center py-2">{text.noManufacturerData}</p>
                        ) : (
                          editProduct.manufacturers.map((m) => (
                            <div key={m.guid_fixed} className="flex items-center justify-between bg-muted/40 p-2 rounded text-sm border border-border/40">
                              <span className="truncate font-medium">{m.code} — {pickName(m.names, lang)}</span>
                              <Button type="button" size="sm" variant="ghost" className="text-destructive hover:bg-destructive/10" onClick={() => removeManufacturer(m.guid_fixed)}>
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
                        {(!editProduct.suppliers || editProduct.suppliers.length === 0) ? (
                          <p className="text-xs text-muted-foreground text-center py-2">{text.noSupplierData}</p>
                        ) : (
                          editProduct.suppliers.map((s) => (
                            <div key={s.guid_fixed} className="flex items-center justify-between bg-muted/40 p-2 rounded text-sm border border-border/40">
                              <span className="truncate font-medium">{s.code} — {pickName(s.names, lang)}</span>
                              <Button type="button" size="sm" variant="ghost" className="text-destructive hover:bg-destructive/10" onClick={() => removeSupplier(s.guid_fixed)}>
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </div>
                          ))
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              )}

                            {/* Tab: classification */}
              {productTab === "classification" && (
                <div className="border border-border rounded-lg p-4 space-y-4">
                  <h4 className="font-semibold text-sm border-b border-border pb-1">{text.groupsAndCategories}</h4>
                  <div className="grid gap-4 sm:grid-cols-2">
                    {[
                      { key: "group", label: text.group, code: editProduct.group_code, names: editProduct.group_names },
                      { key: "groupsubone", label: text.groupsubone, code: editProduct.groupsubonecode, names: editProduct.groupsubonenames },
                      { key: "groupsubtwo", label: text.groupsubtwo, code: editProduct.groupsubtwocode, names: editProduct.groupsubtwonames },
                      { key: "brand", label: text.brand, code: editProduct.brand_code, names: editProduct.brandnames },
                      { key: "category", label: text.category, code: editProduct.categorycode, names: editProduct.category_names },
                      { key: "class", label: text.class, code: editProduct.classcode, names: editProduct.classnames },
                      { key: "design", label: text.design, code: editProduct.designcode, names: editProduct.designnames },
                      { key: "model", label: text.model, code: editProduct.modelcode, names: editProduct.modelnames },
                      { key: "pattern", label: text.pattern, code: editProduct.patterncode, names: editProduct.patternnames },
                      { key: "grade", label: text.grade, code: editProduct.gradecode, names: editProduct.gradenames },
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
              )}

{/* Tab: units */}
              {productTab === "units" && (
                <TabProductUnits value={editProduct} onChange={setEditProduct} lang={lang} openPicker={openPicker} clearPickerField={clearPickerField} />
              )}

              {/* Tab: bom */}
              {productTab === "bom" && (
                <TabProductBom value={editProduct} onChange={setEditProduct} lang={lang} />
              )}

              {/* Tab: stock */}
              {productTab === "stock" && (
                <TabProductStock value={editProduct} onChange={setEditProduct} text={text} />
              )}

              {/* Tab: media */}
              {productTab === "media" && (
                <TabProductMedia value={editProduct} onChange={setEditProduct} auth={auth} language={lang} />
              )}

              {/* Tab: restaurant */}
              {productTab === "restaurant" && (
                <TabProductRestaurant value={editProduct} onChange={setEditProduct} auth={auth} language={lang} shopLanguages={activeLanguages} isSetOnly={isSetOnly} />
              )}

              {/* Tab: timeforsales */}
              {productTab === "timeforsales" && (
                <TabProductTimeForSale value={editProduct} onChange={setEditProduct} language={lang} />
              )}

              {/* Tab: business */}
              {productTab === "business" && (
                <TabProductBusinessBranch value={editProduct} onChange={setEditProduct} auth={auth} language={lang} />
              )}



              {/* Tab: logistics */}
              {productTab === "logistics" && (
                <div className="space-y-4">
                  <div className="border border-border bg-card rounded-lg p-4 space-y-4">
                    <h4 className="font-bold text-base text-foreground border-b border-border pb-2">ข้อมูลขนส่งและขนาดพัสดุ (Logistics & Shipping)</h4>
                    <div className="grid gap-4 sm:grid-cols-2">
                      <div className="space-y-1">
                        <label className="text-xs text-muted-foreground">น้ำหนักพัสดุรวมกล่อง (kg)</label>
                        <Input
                          type="number"
                          min={0}
                          step="any"
                          value={editProduct.package_weight ?? 0}
                          onChange={(e) => setEditProduct({ ...editProduct, package_weight: Math.max(0, Number(e.target.value) || 0) })}
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-xs text-muted-foreground">น้ำหนักเชิงปริมาตรประเมิน (kg)</label>
                        <Input
                          type="text"
                          readOnly
                          className="bg-muted/40 font-mono"
                          value={`${(((editProduct.package_width ?? 0) * (editProduct.package_length ?? 0) * (editProduct.package_height ?? 0)) / 5000).toFixed(3)} kg`}
                        />
                        <p className="text-[10px] text-muted-foreground mt-1">คำนวณจาก (กว้าง x ยาว x สูง) / 5000</p>
                      </div>
                    </div>

                    <div className="grid gap-4 sm:grid-cols-3">
                      <div className="space-y-1">
                        <label className="text-xs text-muted-foreground">ความกว้างกล่อง (cm)</label>
                        <Input
                          type="number"
                          min={0}
                          value={editProduct.package_width ?? 0}
                          onChange={(e) => setEditProduct({ ...editProduct, package_width: Math.max(0, Number(e.target.value) || 0) })}
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-xs text-muted-foreground">ความยาวกล่อง (cm)</label>
                        <Input
                          type="number"
                          min={0}
                          value={editProduct.package_length ?? 0}
                          onChange={(e) => setEditProduct({ ...editProduct, package_length: Math.max(0, Number(e.target.value) || 0) })}
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-xs text-muted-foreground">ความสูงกล่อง (cm)</label>
                        <Input
                          type="number"
                          min={0}
                          value={editProduct.package_height ?? 0}
                          onChange={(e) => setEditProduct({ ...editProduct, package_height: Math.max(0, Number(e.target.value) || 0) })}
                        />
                      </div>
                    </div>

                    <div className="p-3 bg-muted/20 border border-border rounded-lg space-y-3">
                      <p className="text-xs font-semibold text-muted-foreground">คุณลักษณะการจัดส่งและพิมพ์ฉลาก (Shipping Badges):</p>
                      <div className="grid grid-cols-2 gap-3 text-xs">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={editProduct.isalert ?? false}
                            onChange={(e) => setEditProduct({ ...editProduct, isalert: e.target.checked })}
                            className="rounded accent-primary size-4"
                          />
                          <div>
                            <span className="font-semibold block">สินค้าแตกหักง่าย / ระวังแตก (Fragile)</span>
                            <span className="text-[10px] text-muted-foreground">ติดป้ายเตือนและพิมพ์สติ๊กเกอร์เตือนพิเศษ</span>
                          </div>
                        </label>
                      </div>
                      {editProduct.isalert && (
                        <div className="space-y-1">
                          <label className="text-xs text-muted-foreground">คำเตือนสำหรับสติ๊กเกอร์จัดส่ง</label>
                          <Input
                            placeholder="ระบุข้อความ เช่น ห้ามโยน ระวังของแตกหักง่าย"
                            value={editProduct.alertdescription || ""}
                            onChange={(e) => setEditProduct({ ...editProduct, alertdescription: e.target.value })}
                          />
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {/* Tab: misc */}
              {productTab === "misc" && (
                <TabProductMisc value={editProduct} onChange={setEditProduct} language={lang} />
              )}
            </form>
          ) : selectedProduct ? (
            /* Product View Details Mode (Aligned with ProductBarcodeDetail layout) */
            <Card className="min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col shadow-sm border border-border bg-card">
              <CardHeader className="shrink-0 border-b border-border p-4 bg-muted/5">
                <div className="flex items-center justify-between gap-4">
                  <div className="min-w-0">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">{text.detailTitle}</span>
                    <h3 className="text-xl font-bold text-foreground truncate mt-0.5">{selectedProduct.code}</h3>
                    <p className="text-sm text-muted-foreground truncate mt-0.5">{pickName(selectedProduct.names, lang)}</p>
                  </div>
                  <div className="flex flex-wrap justify-end gap-2 shrink-0">
                    <Button variant="outline" size="sm" onClick={() => handleEditOpen(selectedProduct)}>
                      <Pencil className="h-4 w-4 mr-1 text-primary" />
                      {text.edit}
                    </Button>
                    <Button variant="destructive" size="sm" onClick={() => void handleDelete(selectedProduct)}>
                      <Trash2 className="h-4 w-4 mr-1" />
                      {text.delete}
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="grid gap-4 overflow-y-auto p-4 xl:min-h-0 xl:flex-1">
                {/* 1. Basic Info */}
                <DetailSection
                  title={text.basicInfoCard ?? "ข้อมูลพื้นฐานสินค้า"}
                  fields={[
                    ...(!isSetOnly ? [{
                      label: text.itemType,
                      value: selectedProduct.item_type === 2
                        ? text.itemTypeSet
                        : (itemTypes.find((t) => t.value === selectedProduct.item_type)?.label ?? String(selectedProduct.item_type ?? "-"))
                    }] : []),
                    {
                      label: text.vatType,
                      value: vatTypes.find((t) => t.value === selectedProduct.vat_type)?.label ?? String(selectedProduct.vat_type ?? "-")
                    },
                    ...(!isSetOnly ? [{
                      label: text.materialType,
                      value: materialTypes.find((t) => t.value === selectedProduct.materialtype)?.label ?? String(selectedProduct.materialtype ?? "-")
                    }] : [])
                  ]}
                />

                {/* 2. Grouping */}
                <DetailSection
                  title={text.groupingCard ?? "ข้อมูลการจัดกลุ่ม"}
                  fields={[
                    {
                      label: text.group,
                      value: selectedProduct.group_code ? `${selectedProduct.group_code} — ${pickName(selectedProduct.group_names, lang)}` : "-"
                    },
                    {
                      label: text.brand,
                      value: selectedProduct.brand_code ? `${selectedProduct.brand_code} — ${pickName(selectedProduct.brandnames, lang)}` : "-"
                    },
                    {
                      label: text.category,
                      value: selectedProduct.categorycode ? `${selectedProduct.categorycode} — ${pickName(selectedProduct.category_names, lang)}` : "-"
                    }
                  ]}
                />

                {/* 3. Stock */}
                <DetailSection
                  title={text.tabStock ?? "การควบคุมคลังสินค้า"}
                  fields={[
                    { label: text.orderPoint, value: String(selectedProduct.orderpoint ?? 0) },
                    { label: text.minPoint, value: String(selectedProduct.minpoint ?? 0) },
                    { label: text.maxPoint, value: String(selectedProduct.maxpoint ?? 0) },
                    { label: text.qty, value: String(selectedProduct.qty ?? 0) },
                    { label: text.stockBarcode, value: selectedProduct.stockbarcode || "-" }
                  ]}
                />

                {/* 4. Logistics & Dimensions */}
                <DetailSection
                  title="ข้อมูลขนาดและน้ำหนักพัสดุ (Logistics & Dimensions)"
                  fields={[
                    { label: "น้ำหนักรวมพัสดุ", value: `${selectedProduct.package_weight ?? 0} kg` },
                    {
                      label: "มิติตัวกล่อง (ก x ย x ส)",
                      value: `${selectedProduct.package_width ?? 0} x ${selectedProduct.package_length ?? 0} x ${selectedProduct.package_height ?? 0} cm`
                    },
                    {
                      label: "น้ำหนักปริมาตร (ประเมิน)",
                      value: `${(((selectedProduct.package_width ?? 0) * (selectedProduct.package_length ?? 0) * (selectedProduct.package_height ?? 0)) / 5000).toFixed(3)} kg`
                    },
                    { label: "คุณลักษณะพิเศษ", value: selectedProduct.isalert ? "ระวังแตก (Fragile)" : "-" }
                  ]}
                />

                {/* 5. Manufacturers & Suppliers */}
                <div className="grid gap-4 sm:grid-cols-2">
                  <section className="rounded-2xl border border-border p-3 space-y-2 bg-background">
                    <h3 className="text-sm font-semibold">{text.manufacturers}</h3>
                    {(!selectedProduct.manufacturers || selectedProduct.manufacturers.length === 0) ? (
                      <p className="text-xs text-muted-foreground italic py-1">{text.noManufacturerInfo}</p>
                    ) : (
                      <div className="space-y-1">
                        {selectedProduct.manufacturers.map((m) => (
                          <div key={m.guid_fixed} className="bg-muted/40 p-2 rounded text-xs border border-border/40 truncate">
                            {m.code} — {pickName(m.names, lang)}
                          </div>
                        ))}
                      </div>
                    )}
                  </section>

                  <section className="rounded-2xl border border-border p-3 space-y-2 bg-background">
                    <h3 className="text-sm font-semibold">{text.suppliers}</h3>
                    {(!selectedProduct.suppliers || selectedProduct.suppliers.length === 0) ? (
                      <p className="text-xs text-muted-foreground italic py-1">{text.noSupplierInfo}</p>
                    ) : (
                      <div className="space-y-1">
                        {selectedProduct.suppliers.map((s) => (
                          <div key={s.guid_fixed} className="bg-muted/40 p-2 rounded text-xs border border-border/40 truncate">
                            {s.code} — {pickName(s.names, lang)}
                          </div>
                        ))}
                      </div>
                    )}
                  </section>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
              <FolderOpen className="h-10 w-10 mb-2 opacity-50" />
              <span>{text.noSelection}</span>
            </div>
          )}
        </div>
      </div>

      {/* Master Pickers Dialog */}
      <MasterPicker
        open={pickerOpen}
        onClose={() => setPickerOpen(false)}
        auth={auth}
        language={lang}
        master={pickerType as any}
        title={pickerType ? `ค้นหา ${text[pickerType as keyof typeof text] || pickerType}` : ""}
        onSelect={handlePickerSelect}
        placement={pickerAnchorRef.current ? "field" : "dialog"}
        anchorRef={pickerAnchorRef}
      />

      {/* Quick Barcode Picker Dialog */}
      {showBarcodePicker && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-2 md:p-6"
          onClick={() => setShowBarcodePicker(false)}
          role="dialog"
          aria-modal="true"
        >
          <div
            className="flex w-full flex-col overflow-hidden rounded-lg bg-card text-card-foreground shadow-2xl max-h-[85vh] max-w-2xl border border-border"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <div className="text-base font-bold flex items-center gap-2">
                <Link className="h-5 w-5 text-primary" />
                <span>{text.selectUnlinkedBarcode}</span>
              </div>
              <Button variant="ghost" size="sm" onClick={() => setShowBarcodePicker(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="border-b border-border px-4 py-3">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  autoFocus
                  type="search"
                  value={barcodeSearchInput}
                  onChange={(event) => setBarcodeSearchInput(event.target.value)}
                  placeholder={text.searchBarcodeOrName}
                  className="h-9 pl-9"
                />
              </div>
            </div>

            <div className="flex-1 overflow-y-auto">
              {loadingBarcodes ? (
                <div className="flex items-center justify-center gap-2 p-8 text-sm text-muted-foreground">
                  <Loader2 className="h-5 w-5 animate-spin" />
                  <span>{text.searchingBarcode}</span>
                </div>
              ) : barcodeList.length === 0 ? (
                <div className="p-8 text-center text-sm text-muted-foreground">
                  {barcodeSearch ? text.noUnlinkedBarcode : text.allBarcodesLinked}
                </div>
              ) : (
                <ul className="divide-y divide-border">
                  {barcodeList.map((b) => (
                    <li key={b.guidfixed}>
                      <button
                        type="button"
                        onClick={() => handleSelectBarcode(b)}
                        className="grid min-h-12 w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-2 text-left transition hover:bg-muted/60 focus:bg-muted focus:outline-none"
                      >
                        <div className="min-w-0 flex-1">
                          <div className="text-sm font-semibold text-foreground truncate">
                            {pickName(b.names, lang) || text.noName}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {text.unitLabel}: {b.itemunitcode || "-"} {b.itemunitnames && b.itemunitnames.length > 0 ? `(${pickName(b.itemunitnames, lang)})` : ""}
                          </div>
                        </div>
                        <div className="text-right shrink-0">
                          <span className="inline-block rounded bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">
                            {b.barcode}
                          </span>
                        </div>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <div className="border-t border-border px-4 py-3 text-right">
              <Button variant="outline" size="sm" type="button" onClick={() => setShowBarcodePicker(false)}>
                {text.closeWindow}
              </Button>
            </div>
          </div>
        </div>
      )}

      {confirmationDialog}
    </div>
  );
}

// ─── Helpers for Product Form Tabs ────────────────────────────────────────

function FieldRow({ label, hint, children, required }: { label: string; hint?: string; children: ReactNode; required?: boolean }) {
  return (
    <label className="flex flex-col gap-1 text-sm">
      <span className="font-medium text-foreground">
        {label}
        {required ? <span className="ml-1 text-destructive">*</span> : null}
      </span>
      {children}
      {hint ? <span className="text-xs text-muted-foreground">{hint}</span> : null}
    </label>
  );
}

function FieldGrid({ children }: { children: ReactNode }) {
  return <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{children}</div>;
}

function Section({ title, action, children }: { title: string; action?: ReactNode; children: ReactNode }) {
  return (
    <section className="mb-4 overflow-hidden rounded-lg border border-border bg-card">
      <header className="flex items-center justify-between border-b border-border px-3 py-2 bg-muted/20">
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

type RadioOptionValue = string | number | boolean;

function RadioOptionGroup<T extends RadioOptionValue>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: Array<{ value: T; label: string; disabled?: boolean }>;
  onChange: (next: T) => void;
}) {
  return (
    <fieldset className="rounded-md border border-border bg-background px-3 py-2">
      <legend className="px-1 text-xs font-semibold text-muted-foreground">{label}</legend>
      <div className="flex flex-wrap gap-x-5 gap-y-2">
        {options.map((option) => (
          <label
            key={String(option.value)}
            className={cn(
              "flex min-h-8 cursor-pointer items-center gap-2 text-sm",
              option.disabled && "cursor-not-allowed opacity-60",
            )}
          >
            <input
              type="radio"
              checked={Object.is(value, option.value)}
              disabled={option.disabled}
              onChange={() => onChange(option.value)}
              className="size-4"
            />
            <span>{option.label}</span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}

function NumberField({
  value,
  onChange,
  min,
  step = "any",
  className,
}: {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  step?: string | number;
  className?: string;
}) {
  return (
    <Input
      type="number"
      value={Number.isFinite(value) ? value : 0}
      step={step}
      min={min}
      onChange={(event: ChangeEvent<HTMLInputElement>) => {
        const n = Number(event.target.value);
        onChange(Number.isFinite(n) ? n : 0);
      }}
      className={className}
    />
  );
}

function cryptoRandomId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return (crypto as Crypto).randomUUID();
  }
  return Math.random().toString(36).slice(2);
}

function setNameXEntry(list: NameX[] | undefined, code: string, name: string): NameX[] {
  const arr = list ? [...list] : [];
  const idx = arr.findIndex((x) => x.code === code);
  if (idx >= 0) {
    arr[idx] = { ...arr[idx], name };
  } else {
    arr.push({ code, name });
  }
  return arr;
}

// ─── Tab Components ───────────────────────────────────────────────────────

type ProductStateAction = Dispatch<SetStateAction<Product | null>>;

function TabProductMedia({
  value,
  onChange,
  auth,
  language = "th",
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language?: string;
}) {
  const textM = getBarcodeText(language);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string>("");

  const handleUpload = useCallback(
    async (file: File | null, target: "main" | "gallery") => {
      if (!file) return;
      setUploading(true);
      setUploadError("");
      const result = await uploadProductImage(auth, file);
      if (!result.success || !result.data?.url) {
        setUploadError(result.message ?? "Upload failed");
      } else if (target === "main") {
        onChange((c) => c ? ({ ...c, imageuri: result.data!.url }) : null);
      } else {
        onChange((c) => c ? ({
          ...c,
          images: [...(c.images || []), { xorder: (c.images || []).length + 1, uri: result.data!.url }],
        }) : null);
      }
      setUploading(false);
    },
    [auth, onChange],
  );

  return (
    <div className="space-y-4">
      <Section title={textM.mediaUseSection}>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={value.useimageorcolor ?? true}
              onChange={() => onChange((c) => c ? ({ ...c, useimageorcolor: true } as Product) : null)}
            />
            <ImagePlus className="h-4 w-4" />
            <span>แสดงรูปภาพหลัก</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={!(value.useimageorcolor ?? true)}
              onChange={() => onChange((c) => c ? ({ ...c, useimageorcolor: false } as Product) : null)}
            />
            <span className="inline-block size-4 rounded border" style={{ background: value.colorselecthex || "#888" }} />
            <span>แสดงสีป้ายสินค้า</span>
          </label>
        </div>
      </Section>

      {value.useimageorcolor ?? true ? (
        <>
          <Section title={textM.mediaMainSection}>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <div className="size-32 overflow-hidden rounded-lg border border-border bg-muted">
                {value.imageuri ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={value.imageuri}
                    alt="Main Product"
                    className="size-full object-cover"
                  />
                ) : (
                  <div className="flex size-full items-center justify-center text-xs text-muted-foreground">
                    {textM.mediaNoImage}
                  </div>
                )}
              </div>
              <div className="space-y-2 flex-1 max-w-md">
                <Input
                  placeholder="https://… หรือ อัปโหลดไฟล์"
                  value={value.imageuri || ""}
                  onChange={(event) => onChange((c) => c ? ({ ...c, imageuri: event.target.value } as Product) : null)}
                />
                <div className="flex flex-wrap items-center gap-2">
                  <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input px-3 py-1.5 text-xs hover:bg-muted bg-background">
                    <Upload className="h-3.5 w-3.5" />
                    {uploading ? textM.mediaUploading : textM.mediaUploadBtn}
                    <input
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(event) => handleUpload(event.target.files?.[0] ?? null, "main")}
                      disabled={uploading}
                    />
                  </label>
                  {value.imageuri ? (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => onChange((c) => c ? ({ ...c, imageuri: "" } as Product) : null)}
                    >
                      <Trash2 className="mr-1 h-3.5 w-3.5" />
                      {textM.mediaDeleteBtn}
                    </Button>
                  ) : null}
                </div>
                {uploadError ? <p className="text-xs text-destructive">{uploadError}</p> : null}
              </div>
            </div>
          </Section>
          <Section
            title={textM.mediaGallerySection}
            action={
              <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input px-3 py-1.5 text-xs hover:bg-muted bg-background">
                <Plus className="h-3.5 w-3.5" />
                {textM.mediaAddGallery}
                <input
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(event) => handleUpload(event.target.files?.[0] ?? null, "gallery")}
                  disabled={uploading}
                />
              </label>
            }
          >
            {(!value.images || value.images.length === 0) ? (
              <p className="text-sm text-muted-foreground">{textM.mediaNoGallery}</p>
            ) : (
              <ul className="grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6">
                {value.images.map((img, idx) => (
                  <li key={img.xorder} className="relative overflow-hidden rounded-md border border-border">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={img.uri} alt={`#${img.xorder}`} className="aspect-square w-full object-cover" />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="absolute right-1 top-1 size-6 bg-background/80"
                      onClick={() =>
                        onChange((c) => c ? ({ ...c, images: (c.images || []).filter((_, imgIdx) => imgIdx !== idx) } as Product) : null)
                      }
                      aria-label="Delete Image"
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </Section>
        </>
      ) : (
        <Section title={textM.mediaColorSection}>
          <FieldGrid>
            <FieldRow label={textM.mediaColorName}>
              <Input
                value={value.colorselect || ""}
                onChange={(event) => onChange((c) => c ? ({ ...c, colorselect: event.target.value } as Product) : null)}
              />
            </FieldRow>
            <FieldRow label={textM.mediaColorHex}>
              <div className="flex items-center gap-2">
                <Input
                  type="color"
                  value={value.colorselecthex || "#888888"}
                  onChange={(event) => onChange((c) => c ? ({ ...c, colorselecthex: event.target.value } as Product) : null)}
                  className="h-10 w-16 p-1 cursor-pointer"
                />
                <Input
                  value={value.colorselecthex || ""}
                  onChange={(event) => onChange((c) => c ? ({ ...c, colorselecthex: event.target.value } as Product) : null)}
                  placeholder="#RRGGBB"
                />
              </div>
            </FieldRow>
          </FieldGrid>
        </Section>
      )}
    </div>
  );
}

function TabProductRestaurant({
  value,
  onChange,
  auth,
  language,
  shopLanguages,
  isSetOnly = false,
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language: string;
  shopLanguages: string[];
  isSetOnly?: boolean;
}) {
  const text = getBarcodeText(language);
  const foodTypes = useMemo(() => [
    { value: 0, label: text.foodTypeFood },
    { value: 1, label: text.foodTypeDrink },
    { value: 2, label: text.foodTypeAlcohol },
    { value: 3, label: text.foodTypeOther },
  ], [text]);

  const updR = useCallback(
    (key: keyof ProductRestaurant, val: boolean) =>
      onChange((c) => {
        if (!c) return null;
        return {
          ...c,
          restaurant: {
            ...(c.restaurant || {
              isforrestaurant: false,
              isfortakeaway: false,
              isfordelivery: false,
              isforcustomer: false,
              isforcustomerpreorder: false,
            }),
            [key]: val,
          },
        } as Product;
      }),
    [onChange],
  );

  return (
    <div className="space-y-4">
      {!isSetOnly && (
        <>
          <Section title={text.tabRestaurant}>
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              <Toggle checked={value.restaurant?.isforrestaurant ?? false} onCheckedChange={(n) => updR("isforrestaurant", n)} label={text.isForRestaurant} />
              <Toggle checked={value.restaurant?.isfortakeaway ?? false} onCheckedChange={(n) => updR("isfortakeaway", n)} label={text.isForTakeaway} />
              <Toggle checked={value.restaurant?.isfordelivery ?? false} onCheckedChange={(n) => updR("isfordelivery", n)} label={text.isForDelivery} />
              <Toggle checked={value.restaurant?.isforcustomer ?? false} onCheckedChange={(n) => updR("isforcustomer", n)} label={text.isForCustomer} />
              <Toggle checked={value.restaurant?.isforcustomerpreorder ?? false} onCheckedChange={(n) => updR("isforcustomerpreorder", n)} label={text.isForCustomerPreOrder} />
              <Toggle checked={value.isalacarte ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isalacarte: n } as Product) : null)} label={text.isALaCarte} />
              <Toggle checked={value.isstockforrestaurant ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isstockforrestaurant: n } as Product) : null)} label={text.isStockForRestaurant} />
              <Toggle checked={value.issplitunitprint ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, issplitunitprint: n } as Product) : null)} label={text.isSplitUnitPrint} />
              <Toggle checked={value.isonlystaff ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isonlystaff: n } as Product) : null)} label={text.isOnlyStaff} />
            </div>
            <div className="mt-3">
              <RadioOptionGroup
                label={text.foodType}
                value={value.foodtype ?? 0}
                onChange={(val) => onChange((c) => c ? ({ ...c, foodtype: val } as Product) : null)}
                options={foodTypes}
              />
            </div>
          </Section>

          <ProductOrderTypesEditor value={value} onChange={onChange} language={language} auth={auth} />
        </>
      )}

      <ProductOptionsEditor value={value} onChange={onChange} shopLanguages={shopLanguages} language={language} isSetOnly={isSetOnly} />
    </div>
  );
}

function ProductOrderTypesEditor({
  value,
  onChange,
  language,
  auth,
}: {
  value: Product;
  onChange: ProductStateAction;
  language: string;
  auth: AuthSession | null;
}) {
  const [picker, setPicker] = useState({ open: false, idx: undefined as number | undefined });
  const setRows = useCallback(
    (mutator: (rows: ProductOrderType[]) => ProductOrderType[]) =>
      onChange((c) => c ? ({ ...c, ordertypes: mutator(c.ordertypes || []) } as Product) : null),
    [onChange],
  );

  const textOT = getBarcodeText(language);
  return (
    <Section
      title={textOT.orderTypes}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [...rows, { guidfixed: "", code: "", names: [], chargeprice: 0, isdisabled: false } as any])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {textOT.orderTypeAdd}
        </Button>
      }
    >
      {(!value.ordertypes || value.ordertypes.length === 0) ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-2">
          {value.ordertypes.map((entry, idx) => (
            <div key={idx} className="grid grid-cols-1 items-center gap-2 md:grid-cols-[2fr_1fr_40px]">
              <button
                type="button"
                onClick={() => setPicker({ open: true, idx })}
                className="flex h-10 w-full items-center justify-between rounded-lg border border-input bg-background px-3 text-left text-sm hover:bg-muted/40"
              >
                <span className="truncate">{pickName(entry.names, language) || entry.code || "— เลือกบริการสั่งอาหาร —"}</span>
                <span className="text-xs text-muted-foreground">{entry.code}</span>
              </button>
              <NumberField
                value={entry.chargeprice ?? 0}
                onChange={(n) =>
                  setRows((rows) => rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, chargeprice: n } : row)))
                }
                step="any"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
      <MasterPicker
        open={picker.open}
        onClose={() => setPicker({ open: false, idx: undefined })}
        auth={auth}
        language={language}
        master="ordertype"
        title={textOT.orderTypes}
        onSelect={(entry) =>
          setRows((rows) =>
            rows.map((row, rowIdx) =>
              rowIdx === picker.idx
                ? ({ ...row, guidfixed: entry.guidfixed, code: entry.code, names: entry.names } as any)
                : row,
            ),
          )
        }
      />
    </Section>
  );
}

function ProductOptionsEditor({
  value,
  onChange,
  shopLanguages,
  language,
  isSetOnly = false,
}: {
  value: Product;
  onChange: ProductStateAction;
  shopLanguages: string[];
  language: string;
  isSetOnly?: boolean;
}) {
  const textOpt = getBarcodeText(language);
  const setOptions = useCallback(
    (mutator: (rows: ProductOption[]) => ProductOption[]) =>
      onChange((c) => c ? ({ ...c, options: mutator(c.options || []) } as Product) : null),
    [onChange],
  );

  return (
    <Section
      title={isSetOnly ? "จัดการตัวเลือกสินค้าในชุด (เช่น เลือก Case, RAM, CPU, สี)" : textOpt.options}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setOptions((rows) => [
              ...rows,
              { guid: cryptoRandomId(), names: [], choicetype: 0, choices: [] },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {textOpt.optionAdd}
        </Button>
      }
    >
      {(!value.options || value.options.length === 0) ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-4">
          {value.options.map((opt, optIdx) => (
            <div key={opt.guid} className="space-y-2 rounded-md border border-border p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground font-semibold">กลุ่มตัวเลือกที่ #{optIdx + 1}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="text-destructive hover:bg-destructive/10"
                  onClick={() => setOptions((rows) => rows.filter((_, idx) => idx !== optIdx))}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>

              <NamesEditor
                names={opt.names || []}
                onChange={(nextNames) =>
                  setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, names: nextNames } : row)))
                }
                languages={shopLanguages}
                label={textOpt.optionGroupName}
                language={language}
              />

              <div className="grid gap-3 sm:grid-cols-3">
                <RadioOptionGroup
                  label={textOpt.optionChoiceType}
                  value={opt.choicetype}
                  onChange={(n) =>
                    setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, choicetype: n } : row)))
                  }
                  options={[
                    { value: 0, label: textOpt.optionChoiceTypeMulti },
                    { value: 1, label: textOpt.optionChoiceTypeSingle },
                  ]}
                />
                <FieldRow label={textOpt.optionMinSelectLabel}>
                  <NumberField
                    value={opt.minselect ?? 0}
                    onChange={(n) =>
                      setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, minselect: n } : row)))
                    }
                    min={0}
                    step={1}
                  />
                </FieldRow>
                <FieldRow label={textOpt.optionMaxSelectLabel}>
                  <NumberField
                    value={opt.maxselect ?? 0}
                    onChange={(n) =>
                      setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, maxselect: n } : row)))
                    }
                    min={0}
                    step={1}
                  />
                </FieldRow>
              </div>

              {/* Choices inside Option */}
              <div className="mt-3 border-t border-border pt-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-semibold text-muted-foreground">{textOpt.optionChoiceList}</span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setOptions((rows) =>
                        rows.map((row, idx) =>
                          idx === optIdx
                            ? {
                                ...row,
                                choices: [
                                  ...(row.choices || []),
                                  {
                                    guid: cryptoRandomId(),
                                    names: [],
                                    imageuri: "",
                                    refbarcode: "",
                                    refbarcodenames: [],
                                    refproductcode: "",
                                    refunitcode: "",
                                    isstock: false,
                                    isdefault: false,
                                    qty: 1,
                                    price: "",
                                    vatcal: 0,
                                  },
                                ],
                              }
                            : row,
                        ),
                      )
                    }
                  >
                    <Plus className="mr-1 h-3.5 w-3.5" />
                    {textOpt.optionAddChoice}
                  </Button>
                </div>

                {(!opt.choices || opt.choices.length === 0) ? (
                  <p className="text-xs text-muted-foreground text-center py-2">—</p>
                ) : (
                  <div className="space-y-3">
                    {opt.choices.map((choice, choiceIdx) => (
                      <div key={choice.guid} className="p-3 border border-border/60 rounded bg-muted/20 space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-xs text-muted-foreground font-semibold">ตัวเลือกย่อย #{choiceIdx + 1}</span>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-6 text-destructive hover:bg-destructive/10"
                            onClick={() =>
                              setOptions((rows) =>
                                rows.map((row, idx) =>
                                  idx === optIdx
                                    ? { ...row, choices: (row.choices || []).filter((_, cIdx) => cIdx !== choiceIdx) }
                                    : row,
                                ),
                              )
                            }
                          >
                            <X className="h-3.5 w-3.5" />
                          </Button>
                        </div>

                        <NamesEditor
                          names={choice.names || []}
                          onChange={(nextNames) =>
                            setOptions((rows) =>
                              rows.map((row, idx) =>
                                idx === optIdx
                                  ? {
                                      ...row,
                                      choices: (row.choices || []).map((c, cIdx) =>
                                        cIdx === choiceIdx ? { ...c, names: nextNames } : c,
                                      ),
                                    }
                                  : row,
                              ),
                            )
                          }
                          languages={shopLanguages}
                          label={textOpt.optionChoiceName}
                          language={language}
                        />

                        <div className="grid gap-3 sm:grid-cols-4">
                          <FieldRow label={textOpt.optionChoicePriceLabel}>
                            <Input
                              value={choice.price || ""}
                              onChange={(e) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, price: e.target.value } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                              placeholder="0.00"
                            />
                          </FieldRow>
                          <FieldRow label={textOpt.optionChoiceQtyLabel}>
                            <NumberField
                              value={choice.qty ?? 0}
                              onChange={(n) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, qty: n } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                            />
                          </FieldRow>
                          <div className="flex items-center pt-5">
                            <Toggle
                              checked={choice.isdefault}
                              onCheckedChange={(n) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, isdefault: n } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                              label={textOpt.optionChoiceDefaultLabel}
                            />
                          </div>
                          <div className="flex items-center pt-5">
                            <Toggle
                              checked={choice.isstock}
                              onCheckedChange={(n) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, isstock: n } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                              label={textOpt.optionChoiceStockLabel}
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

function TabProductTimeForSale({
  value,
  onChange,
  language = "th",
}: {
  value: Product;
  onChange: ProductStateAction;
  language?: string;
}) {
  const textT = getBarcodeText(language);
  const setRows = useCallback(
    (mutator: (rows: ProductTimeForSale[]) => ProductTimeForSale[]) =>
      onChange((c) => c ? ({ ...c, timeforsales: mutator(c.timeforsales || []) } as Product) : null),
    [onChange],
  );

  const DAYS = [textT.sun, textT.mon, textT.tue, textT.wed, textT.thu, textT.fri, textT.sat];

  return (
    <Section
      title={textT.timeSection}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [
              ...rows,
              { daysofweek: [1, 2, 3, 4, 5], fromdate: "", todate: "", fromtime: "00:00", totime: "23:59" },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {textT.timeAddBtn}
        </Button>
      }
    >
      {(!value.timeforsales || value.timeforsales.length === 0) ? (
        <p className="text-sm text-muted-foreground">{textT.timeNoLimit}</p>
      ) : (
        <div className="space-y-4">
          {value.timeforsales.map((entry, idx) => (
            <div key={idx} className="p-3 border border-border rounded bg-muted/10 space-y-3">
              <div className="flex items-center justify-between border-b border-border/60 pb-1">
                <span className="text-xs font-semibold text-muted-foreground">ข้อกำหนดเวลาขายที่ #{idx + 1}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 text-destructive hover:bg-destructive/10"
                  onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>

              {/* Days of week checklist */}
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground font-semibold">{textT.timeDaysLabel}</label>
                <div className="flex flex-wrap gap-2">
                  {DAYS.map((dayLabel, dayIdx) => {
                    const checked = (entry.daysofweek || []).includes(dayIdx as any);
                    return (
                      <label key={dayIdx} className="flex items-center gap-1 text-xs cursor-pointer bg-background p-1.5 rounded border border-border hover:bg-muted/40">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(e) => {
                            const activeDays = entry.daysofweek || [];
                            const nextDays = e.target.checked
                              ? [...activeDays, dayIdx as any]
                              : activeDays.filter((d) => d !== dayIdx);
                            setRows((rows) =>
                              rows.map((row, rowIdx) =>
                                rowIdx === idx ? { ...row, daysofweek: nextDays } : row,
                              ),
                            );
                          }}
                          className="size-3"
                        />
                        <span>{dayLabel}</span>
                      </label>
                    );
                  })}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-4">
                <FieldRow label={textT.timeFromDate}>
                  <Input
                    type="date"
                    value={entry.fromdate || ""}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, fromdate: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textT.timeToDate}>
                  <Input
                    type="date"
                    value={entry.todate || ""}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, todate: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textT.timeFromTime}>
                  <Input
                    type="time"
                    value={entry.fromtime || "00:00"}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, fromtime: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textT.timeToTime}>
                  <Input
                    type="time"
                    value={entry.totime || "23:59"}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, totime: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
              </div>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

function TabProductBusinessBranch({
  value,
  onChange,
  auth,
  language,
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language: string;
}) {
  const textBB = getBarcodeText(language);
  const [pickerType, setPickerType] = useState<"branch" | "businesstype" | null>(null);

  const setBusinessRows = useCallback(
    (mutator: (rows: ProductBarcodeBusinessType[]) => ProductBarcodeBusinessType[]) =>
      onChange((c) => c ? ({ ...c, businesstypes: mutator(c.businesstypes || []) } as Product) : null),
    [onChange],
  );

  const setBranchRows = useCallback(
    (mutator: (rows: ProductBarcodeBranch[]) => ProductBarcodeBranch[]) =>
      onChange((c) => c ? ({ ...c, ignorebranches: mutator(c.ignorebranches || []) } as Product) : null),
    [onChange],
  );

  return (
    <div className="space-y-4">
      {/* Business Types (Ignore) */}
      <Section
        title={textBB.businessTypeSection}
        action={
          <Button type="button" variant="outline" size="sm" onClick={() => setPickerType("businesstype")}>
            <Plus className="mr-1 h-4 w-4" />
            {textBB.businessTypeAddBtn}
          </Button>
        }
      >
        {(!value.businesstypes || value.businesstypes.length === 0) ? (
          <p className="text-sm text-muted-foreground">{textBB.businessTypeNoData}</p>
        ) : (
          <ul className="space-y-1">
            {value.businesstypes.map((entry, idx) => (
              <li key={entry.guidfixed || idx} className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1.5 bg-muted/10 text-sm">
                <span>{pickName(entry.names, language) || entry.code}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 text-destructive hover:bg-destructive/10"
                  onClick={() => setBusinessRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Minus className="h-4 w-4" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </Section>

      {/* Ignore Branches */}
      <Section
        title={textBB.branchSection}
        action={
          <Button type="button" variant="outline" size="sm" onClick={() => setPickerType("branch")}>
            <Plus className="mr-1 h-4 w-4" />
            {textBB.branchAddBtn}
          </Button>
        }
      >
        <p className="mb-2 text-xs text-muted-foreground">{textBB.branchHintDetail}</p>
        {(!value.ignorebranches || value.ignorebranches.length === 0) ? (
          <p className="text-sm text-muted-foreground">{textBB.branchNoData}</p>
        ) : (
          <ul className="space-y-1">
            {value.ignorebranches.map((entry, idx) => (
              <li key={entry.guidfixed || idx} className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1.5 bg-muted/10 text-sm">
                <span>{pickName(entry.names, language) || entry.code}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 text-destructive hover:bg-destructive/10"
                  onClick={() => setBranchRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Minus className="h-4 w-4" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </Section>

      <MasterPicker
        open={pickerType !== null}
        onClose={() => setPickerType(null)}
        auth={auth}
        language={language}
        master={pickerType === "branch" ? "branch" : "businesstype"}
        title={pickerType === "branch" ? textBB.pickerBranch : textBB.pickerBusinessType}
        onSelect={(entry) => {
          if (pickerType === "branch") {
            setBranchRows((rows) => [
              ...rows,
              { guidfixed: entry.guidfixed, code: entry.code, names: entry.names, isignore: true },
            ]);
          } else {
            setBusinessRows((rows) => [
              ...rows,
              { guidfixed: entry.guidfixed, code: entry.code, names: entry.names, isignore: false },
            ]);
          }
        }}
      />
    </div>
  );
}

function TabProductMisc({
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
            <Input value={value.holding_code || ""} readOnly className="bg-muted/50" />
          </FieldRow>
          <FieldRow label={textMisc.miscUnitGuid}>
            <Input value={value.unitguid || ""} readOnly className="bg-muted/50" />
          </FieldRow>
        </FieldGrid>
      </Section>
    </div>
  );
}

function TabProductStock({
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


function TabProductBom({
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
                  item_unit_code: "",
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
                    value={entry.item_unit_code || ""}
                    onChange={(event) =>
                      setBom((rows) =>
                        rows.map((row, rowIdx) =>
                          rowIdx === idx ? { ...row, item_unit_code: event.target.value } : row,
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

type DetailFieldItem = {
  label: string;
  value: string;
};

function DetailSection({ fields, title }: { fields: DetailFieldItem[]; title: string }) {
  return (
    <section className="rounded-2xl border border-border p-3">
      <h3 className="mb-2 text-sm font-semibold">{title}</h3>
      <div className="grid gap-2 sm:grid-cols-2">
        {fields.map((field) => (
          <DetailField key={`${title}-${field.label}`} label={field.label} value={field.value} />
        ))}
      </div>
    </section>
  );
}

function DetailField({ label, value }: DetailFieldItem) {
  return (
    <div className="min-w-0 rounded-xl border border-border bg-background p-2">
      <p className="text-xs font-semibold text-muted-foreground">{label}</p>
      <p className="mt-1 break-words text-sm font-medium">{value || "-"}</p>
    </div>
  );
}

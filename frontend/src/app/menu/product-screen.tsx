"use client";

import {
  CheckSquare,
  FolderOpen,
  Filter,
  ImageIcon,
  ImageOff,
  Loader2,
  Pencil,
  Plus,
  Copy,
  AlertCircle,
  Eye,
  EyeOff,
  RefreshCcw,
  Search,
  Trash2,
  X,
  Save,
  Link,
  ChevronLeft,
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { AuthenticatedImg } from "@/components/authenticated-image";
import { BusinessImageGallery } from "@/components/product-barcode/business-image-editor";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { listBarcodes, type MasterEntry } from "@/lib/product-barcode/api";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { normalizeBusinessCode } from "@/lib/business-code";
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
  type NameX,
  type ProductOption,
  type ProductTimeForSale,
  type ProductUnitConversion,
  type RefProductBarcode,
  type BOMProductBarcode,
} from "@/lib/product-barcode/types";
import { pickName, rawToProduct } from "@/lib/product-barcode/utils";
import { languageCodesFromWorkspace } from "@/components/product-barcode/names-editor";
import { getBarcodeText } from "@/lib/product-barcode/language";

import { TabProductUnits } from "./tab-product-units";
import { TabProductMedia } from "./tab-product-media";
import { TabProductRestaurant } from "./tab-product-restaurant";
import { TabProductTimeForSale } from "./tab-product-timeforsale";
import { TabProductBusinessBranch } from "./tab-product-business-branch";
import { TabProductMisc } from "./tab-product-misc";
import { TabProductStock } from "./tab-product-stock";
import { TabProductBom } from "./tab-product-bom";
import { TabProductBasic } from "./tab-product-basic";
import { TabProductClassification } from "./tab-product-classification";
import { TabProductLogistics } from "./tab-product-logistics";
import {
  FieldRow,
  FieldGrid,
  Section,
  Toggle,
  RadioOptionGroup,
  NumberField,
  type ProductStateAction,
} from "./product-tab-shared";
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

async function ensureActiveProductHolding(
  auth: AuthSession,
  holdingcode: string,
  businesscode: string,
): Promise<void> {
  const response = await fetch("/api/workspace/select-holding", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-bc-backend-url": auth.backendUrl,
      Authorization: `Bearer ${auth.token}`,
    },
    body: JSON.stringify({
      backendUrl: auth.backendUrl,
      holdingcode,
      businesscode,
    }),
    cache: "no-store",
  });
  const data = (await response.json().catch(() => null)) as {
    success?: boolean;
    message?: string;
  } | null;
  if (!response.ok || data?.success === false) {
    throw new Error(data?.message || "ไม่สามารถเลือกบริษัทใน token ได้");
  }
}

const PRIMARY_PRODUCT_TABS = (text: ReturnType<typeof getBarcodeText>) => ({
  basic: text.tabBasic,
  classification: text.tabClassification,
  units: text.tabUnitsBarcode,
  bom: text.tabBomShort,
  stock: text.tabStock,
  media: text.tabMedia,
});

const ADVANCED_PRODUCT_TABS = (text: ReturnType<typeof getBarcodeText>) => ({
  logistics: "การจัดส่ง / โลจิสติกส์",
  restaurant: text.tabRestaurant,
  timeforsales: text.tabTimeForSales,
  business: text.tabBusinessBranchShort,
  misc: text.tabMisc,
});

const PRODUCT_SPLIT_DEFAULT_LEFT = 30;
const PRODUCT_SPLIT_STORAGE_KEY = "bc_product_split_left_v3";
const PRODUCT_SPLIT_MIN_LEFT = 24;
const PRODUCT_SPLIT_MAX_LEFT = 50;

function clampProductSplitLeft(value: number) {
  if (!Number.isFinite(value)) return PRODUCT_SPLIT_DEFAULT_LEFT;
  return Math.min(
    PRODUCT_SPLIT_MAX_LEFT,
    Math.max(PRODUCT_SPLIT_MIN_LEFT, value),
  );
}

export function ProductScreen({
  active = true,
  embedded = false,
  focusRequest,
  language = "th",
}: {
  active?: boolean;
  embedded?: boolean;
  focusRequest?: { code: string; requestId: string };
  language?: LanguageCode;
}) {
  const lang = normalizeLanguage(language);
  const text = getBarcodeText(lang);
  const { confirm, confirmationDialog } = useConfirmDialog();

  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [items, setItems] = useState<Product[]>([]);
  const [selectedCode, setSelectedCode] = useState("");
  const [loading, setLoading] = useState(false);
  const setNotice = pushNotice;
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [filterOpen, setFilterOpen] = useState(false);
  const [listItemTypeFilter, setListItemTypeFilter] = useState("all");
  const [showListImage, setShowListImage] = useState(false);
  const [selectMode, setSelectMode] = useState(false);
  const [checkedProductKeys, setCheckedProductKeys] = useState<string[]>([]);

  // Resizable split states
  const [splitLeftPercent, setSplitLeftPercent] = useState(
    PRODUCT_SPLIT_DEFAULT_LEFT,
  );
  const [resizingSplit, setResizingSplit] = useState(false);
  const splitContainerRef = useRef<HTMLDivElement | null>(null);
  const selectedShopTokenRef = useRef("");
  const productListRequestRef = useRef(0);
  const handledFocusRequestRef = useRef("");
  const productDetailRequestRef = useRef(0);
  const productDetailRef = useRef<Product | null>(null);

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
      }) as React.CSSProperties,
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

  const stopSplitResize = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      event.currentTarget.releasePointerCapture(event.pointerId);
      setResizingSplit(false);
      setSplitLeftPercent((current) => {
        const next = clampProductSplitLeft(current);
        if (typeof window !== "undefined") {
          window.localStorage.setItem(
            PRODUCT_SPLIT_STORAGE_KEY,
            String(Math.round(next)),
          );
        }
        return next;
      });
    },
    [],
  );

  const adjustSplitWithKeyboard = useCallback(
    (event: React.KeyboardEvent<HTMLDivElement>) => {
      let direction = 0;
      if (event.key === "ArrowLeft") direction = -2;
      else if (event.key === "ArrowRight") direction = 2;
      if (direction === 0) return;
      event.preventDefault();
      setSplitLeftPercent((current) =>
        clampProductSplitLeft(current + direction),
      );
    },
    [],
  );

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
  const [productTab, setProductTab] = useState<
    | "basic"
    | "classification"
    | "units"
    | "bom"
    | "stock"
    | "media"
    | "logistics"
    | "restaurant"
    | "timeforsales"
    | "business"
    | "misc"
  >("basic");

  const [showEmptyDetails, setShowEmptyDetails] = useState(false);
  const [selectedProductDetail, setSelectedProductDetail] =
    useState<Product | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailError, setDetailError] = useState("");
  const [detailReloadKey, setDetailReloadKey] = useState(0);

  const itemTypes = useMemo(
    () => [
      { value: 0, label: text.itemTypeStock },
      { value: 1, label: text.itemTypeService },
    ],
    [text],
  );

  const vatTypes = useMemo(
    () => [
      { value: 0, label: text.vatIncluded },
      { value: 1, label: text.vatExcluded },
    ],
    [text],
  );

  const materialTypes = useMemo(
    () => [
      { value: 0, label: text.materialGeneral },
      { value: 1, label: text.materialMaterial },
      { value: 2, label: text.materialSemiFinished },
      { value: 3, label: text.materialSet },
      { value: 4, label: text.materialAgricultural },
    ],
    [text],
  );

  const sumPointTypes = useMemo(
    () => [
      { value: "true", label: text.isSumPointYes },
      { value: "false", label: text.isSumPointNo },
    ],
    [text],
  );

  const sumPointOptions = useMemo(
    () => [
      { value: true, label: text.isSumPointYes },
      { value: false, label: text.isSumPointNo },
    ],
    [text],
  );

  const foodTypes = useMemo(
    () => [
      { value: 0, label: text.foodTypeFood },
      { value: 1, label: text.foodTypeDrink },
      { value: 2, label: text.foodTypeAlcohol },
      { value: 3, label: text.foodTypeOther },
    ],
    [text],
  );

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
  const [bindBarcodeOnSave, setBindBarcodeOnSave] =
    useState<ProductBarcode | null>(null);
  const activeHoldingCode = workspace?.shop.holdingcode ?? "";
  const activeBusinessCode = normalizeBusinessCode(workspace?.company?.code);

  useEffect(() => {
    setAuth(readAuthSession());
    setWorkspace(readWorkspaceSession());
  }, []);

  useEffect(() => {
    const handleWorkspaceChange = () => {
      const nextWorkspace = readWorkspaceSession();
      const nextHoldingCode = nextWorkspace?.shop.holdingcode ?? "";
      const nextBusinessCode = normalizeBusinessCode(
        nextWorkspace?.company?.code,
      );
      if (
        nextHoldingCode !== activeHoldingCode ||
        nextBusinessCode !== activeBusinessCode
      ) {
        productListRequestRef.current += 1;
        productDetailRequestRef.current += 1;
        productDetailRef.current = null;
        selectedShopTokenRef.current = "";
        setItems([]);
        setSelectedCode("");
        setSelectedProductDetail(null);
        setDetailLoading(false);
        setDetailError("");
        setCheckedProductKeys([]);
        setSelectMode(false);
        setEditorOpen(false);
        setEditProduct(null);
        setBindBarcodeOnSave(null);
      }
      setWorkspace(nextWorkspace);
    };
    window.addEventListener(WORKSPACE_CHANGED_EVENT, handleWorkspaceChange);
    window.addEventListener("storage", handleWorkspaceChange);
    return () => {
      window.removeEventListener(
        WORKSPACE_CHANGED_EVENT,
        handleWorkspaceChange,
      );
      window.removeEventListener("storage", handleWorkspaceChange);
    };
  }, [activeBusinessCode, activeHoldingCode]);

  useEffect(() => {
    if (!showBarcodePicker || !auth || !activeBusinessCode) return;
    let active = true;
    setLoadingBarcodes(true);
    listBarcodes(auth, {
      holdingcode: activeHoldingCode,
      businesscode: activeBusinessCode,
      keyword: barcodeSearch,
      limit: 100,
    })
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
  }, [
    showBarcodePicker,
    auth,
    activeBusinessCode,
    activeHoldingCode,
    barcodeSearch,
  ]);

  const activeLanguages = useMemo(
    () => languageCodesFromWorkspace(workspace),
    [workspace],
  );

  const selectedListProduct = useMemo(() => {
    if (!selectedCode) return null;
    return items.find((item) => item.code === selectedCode) ?? null;
  }, [items, selectedCode]);

  useEffect(() => {
    const requestId = ++productDetailRequestRef.current;
    const listProduct = selectedListProduct;
    const previousDetail = productDetailRef.current;

    setDetailError("");
    setShowEmptyDetails(false);

    if (!active || !auth || !listProduct?.code) {
      setDetailLoading(false);
      if (!listProduct?.code) {
        productDetailRef.current = null;
        setSelectedProductDetail(null);
      }
      return () => {
        if (productDetailRequestRef.current === requestId) {
          productDetailRequestRef.current += 1;
        }
      };
    }

    setDetailLoading(true);
    void fetch(`/api/product/${encodeURIComponent(listProduct.code)}`, {
      headers: {
        Authorization: `Bearer ${auth.token}`,
        "x-bc-backend-url": auth.backendUrl,
      },
      cache: "no-store",
    })
      .then(async (response) => {
        const data = (await response.json().catch(() => null)) as {
          success?: boolean;
          data?: unknown;
          message?: string;
        } | null;
        if (!response.ok || data?.success === false || !data?.data) {
          throw new Error(data?.message || text.requestFailed);
        }
        return data.data;
      })
      .then((rawDetail) => {
        if (requestId !== productDetailRequestRef.current) return;
        const normalized = rawToProduct(rawDetail);
        if (
          normalized.holdingcode &&
          normalized.holdingcode !== activeHoldingCode
        ) {
          throw new Error("ข้อมูลสินค้าไม่ตรงกับกลุ่มกิจการที่เลือก");
        }
        const rawBusinessCode = normalizeBusinessCode(
          typeof (rawDetail as { businesscode?: unknown }).businesscode ===
            "string"
            ? (rawDetail as { businesscode: string }).businesscode
            : "",
        );
        if (rawBusinessCode && rawBusinessCode !== activeBusinessCode) {
          throw new Error("ข้อมูลสินค้าไม่ตรงกับบริษัทที่เลือก");
        }
        const detail = {
          ...listProduct,
          ...(rawDetail as Product),
          ...normalized,
          qty: listProduct.qty,
          _unit_count: listProduct._unit_count,
        };
        productDetailRef.current = detail;
        setSelectedProductDetail(detail);
      })
      .catch((error: unknown) => {
        if (requestId !== productDetailRequestRef.current) return;
        const message =
          error instanceof Error ? error.message : text.requestFailed;
        setDetailError(message);
        setNotice({ type: "error", text: message });
        if (previousDetail?.code && previousDetail.code !== listProduct.code) {
          setSelectedCode(previousDetail.code);
        }
      })
      .finally(() => {
        if (requestId === productDetailRequestRef.current) {
          setDetailLoading(false);
        }
      });

    return () => {
      if (productDetailRequestRef.current === requestId) {
        productDetailRequestRef.current += 1;
      }
    };
  }, [
    active,
    activeBusinessCode,
    activeHoldingCode,
    auth,
    detailReloadKey,
    selectedListProduct,
    setNotice,
    text.requestFailed,
  ]);

  const selectedProduct = selectedProductDetail;

  const visibleItems = useMemo(() => {
    if (listItemTypeFilter === "all") return items;
    return items.filter(
      (item) => String(item.itemtype ?? 0) === listItemTypeFilter,
    );
  }, [items, listItemTypeFilter]);

  const isFormDirty = useMemo(() => {
    if (!editorOpen || !editProduct || !selectedProduct) return false;
    // Compare basic fields for dirty detection
    return JSON.stringify(editProduct) !== JSON.stringify(selectedProduct);
  }, [editorOpen, editProduct, selectedProduct]);

  const requestProductSelection = useCallback(
    async (code: string): Promise<boolean> => {
      if (isFormDirty) {
        const ok = await confirm({
          title: "ข้อมูลมีการเปลี่ยนแปลง",
          description:
            "คุณมีข้อมูลที่ยังไม่ได้บันทึก ต้องการละทิ้งการเปลี่ยนแปลงแล้วเปลี่ยนสินค้าหรือไม่?",
          confirmLabel: "เปลี่ยนสินค้า",
          cancelLabel: "ยกเลิก",
        });
        if (!ok) return false;
      }

      setSelectedCode(code);
      setEditorOpen(false);
      return true;
    },
    [isFormDirty, confirm],
  );

  const handleSelectProduct = useCallback(
    (code: string) => {
      void requestProductSelection(code);
    },
    [requestProductSelection],
  );

  const handleCancelEdit = useCallback(() => {
    if (isFormDirty) {
      void confirm({
        title: "ข้อมูลมีการเปลี่ยนแปลง",
        description:
          "คุณมีข้อมูลที่ยังไม่ได้บันทึก ต้องการยกเลิกการแก้ไขใช่หรือไม่?",
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

  useEffect(() => {
    const requestId = focusRequest?.requestId ?? "";
    const code = focusRequest?.code.trim().toUpperCase() ?? "";
    if (!active || !requestId || !code) return;
    if (handledFocusRequestRef.current === requestId) return;
    handledFocusRequestRef.current = requestId;

    let cancelled = false;
    void requestProductSelection(code).then((selected) => {
      if (cancelled || !selected) return;
      setListItemTypeFilter("all");
      setSearchInput(code);
      setSearch(code);
    });

    return () => {
      cancelled = true;
    };
  }, [
    active,
    focusRequest?.code,
    focusRequest?.requestId,
    requestProductSelection,
  ]);

  const loadProducts = useCallback(async () => {
    if (!auth || !activeHoldingCode || !activeBusinessCode) return;
    const requestId = ++productListRequestRef.current;
    setLoading(true);
    setNotice(null);
    try {
      const tokenShopKey = `${auth.token}:${activeHoldingCode}:${activeBusinessCode}`;
      if (selectedShopTokenRef.current !== tokenShopKey) {
        await ensureActiveProductHolding(
          auth,
          activeHoldingCode,
          activeBusinessCode,
        );
        selectedShopTokenRef.current = tokenShopKey;
      }
      const params = new URLSearchParams({
        q: search,
        limit: "80",
        holdingcode: activeHoldingCode,
      });
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
      const filtered = normalized.filter((item) => item.itemtype !== 2);
      if (requestId !== productListRequestRef.current) return;

      setItems(filtered);
      if (filtered.length > 0) {
        // Only auto-select first item on md+ screens; on mobile the list stays visible
        const isDesktop =
          typeof window !== "undefined" &&
          window.matchMedia("(min-width: 768px)").matches;
        if (isDesktop) {
          setSelectedCode((prev) => prev || filtered[0].code);
        }
      }
    } catch (err) {
      if (requestId !== productListRequestRef.current) return;
      setItems([]);
      setNotice({
        type: "error",
        text: err instanceof Error ? err.message : text.requestFailed,
      });
    } finally {
      if (requestId === productListRequestRef.current) setLoading(false);
    }
  }, [auth, activeBusinessCode, activeHoldingCode, search, text.requestFailed]);

  useEffect(() => {
    if (active && auth && activeHoldingCode && activeBusinessCode) {
      void loadProducts();
    }
  }, [active, auth, activeBusinessCode, activeHoldingCode, loadProducts]);

  const makeBlankProduct = useCallback(
    (): Product => ({
      guidfixed: "",
      holdingcode: activeHoldingCode,
      code: "",
      names: [
        { code: "th", name: "" },
        { code: "en", name: "" },
      ],
      groupcode: "",
      groupnames: [],
      itemtype: 0,
      vattype: 0,
      materialtype: 0,
      issumpoint: false,
      manufacturers: [],
      suppliers: [],
      condition: false,
      unitcode: "",
      unitnames: [],
      dividevalue: 1,
      standvalue: 1,
      unitconversions: [],
      isusesubbarcodes: false,
      refbarcodes: [],
      bom: [],
      orderpoint: 0,
      minpoint: 0,
      maxpoint: 0,
      qty: 0,
      stockbarcode: "",
    }),
    [activeHoldingCode],
  );

  const handleCreateOpen = () => {
    setEditorMode("create");
    setBindBarcodeOnSave(null);
    setEditProduct(makeBlankProduct());
    setEditorOpen(true);
  };

  const handleCreateCopyOpen = () => {
    if (!selectedProduct) return;
    setEditorMode("create");
    setBindBarcodeOnSave(null);
    setEditProduct({
      ...selectedProduct,
      guidfixed: "",
      holdingcode: activeHoldingCode,
      barcodes: [],
    });
    setProductTab("basic");
    setEditorOpen(true);
  };

  const handleSelectBarcode = async (barcodeRow: ProductBarcodeListRow) => {
    if (!auth) return;
    try {
      setLoading(true);
      const res = await fetch(
        `/api/product-barcode/${encodeURIComponent(barcodeRow.guidfixed)}`,
        {
          headers: {
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
        },
      );
      const resData = await res.json();
      if (resData.success && resData.data) {
        const b = resData.data as ProductBarcode;
        setEditProduct((current) => {
          if (!current) return null;
          return {
            ...current,
            names: b.names && b.names.length > 0 ? b.names : current.names,
            code: b.itemcode || b.barcode || current.code,
            groupcode: b.groupcode || "",
            groupnames: b.groupnames || [],
            subgroupguid: b.subgroupguid || "",
            subgroupcode: b.subgroupcode || "",
            subgroupnames: b.subgroupnames || [],
            brandguid: b.brandguid || "",
            brandcode: b.brandcode || "",
            brandnames: b.brandnames || [],
            categoryguid: b.categoryguid || "",
            categorycode: b.categorycode || "",
            categorynames: b.categorynames || [],
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
            vattype: b.vattype ?? 0,
            itemtype: b.itemtype ?? 0,
            materialtype: b.materialtype ?? 0,
            taxtype: b.taxtype ?? 0,
            manufacturers: b.manufacturers || [],
            suppliers: b.suppliers || [],
            unitguid: b.itemunitguid || "",
            unitcode: b.itemunitcode || "",
            unitnames: b.itemunitnames || [],
            condition: false,
            dividevalue: 1,
            standvalue: 1,
            unitconversions: [],
            bom: b.bom || [],
            orderpoint: b.orderpoint ?? 0,
            minpoint: b.minpoint ?? 0,
            maxpoint: b.maxpoint ?? 0,
            qty: b.qty ?? 0,
            stockbarcode: b.stockbarcode || "",
          };
        });
        setBindBarcodeOnSave(b);
        setNotice({
          type: "info",
          text: text.barcodeFetchSuccess.replace("%s", b.barcode),
        });
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
      const res = await fetch(
        `/api/product/${encodeURIComponent(p.guidfixed)}`,
        {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${auth.token}`,
            "x-bc-backend-url": auth.backendUrl,
          },
        },
      );
      const data = await res.json();
      if (!res.ok || data.success === false) {
        throw new Error(data.message || "Delete failed");
      }
      setNotice({ type: "success", text: text.deleteSuccess });
      void loadProducts();
      setSelectedCode("");
    } catch (err) {
      setNotice({
        type: "error",
        text: err instanceof Error ? err.message : "Delete failed",
      });
    }
  };

  const toggleCheckedProduct = (key: string) => {
    setCheckedProductKeys((current) =>
      current.includes(key)
        ? current.filter((item) => item !== key)
        : [...current, key],
    );
  };

  const handleDeleteSelectedProducts = async () => {
    if (!auth || checkedProductKeys.length === 0) return;
    const selectedItems = visibleItems.filter((item) =>
      checkedProductKeys.includes(productRowKey(item)),
    );
    const codes = selectedItems.map((item) => item.code).filter(Boolean);
    if (codes.length === 0) return;
    const ok = await confirm({
      title: text.deleteConfirm,
      description: `เลือกไว้ ${codes.length.toLocaleString("th-TH")} รายการ`,
      tone: "danger",
      confirmLabel: text.delete,
      cancelLabel: text.cancel,
    });
    if (!ok) return;
    try {
      for (const code of codes) {
        const res = await fetch(`/api/product/${encodeURIComponent(code)}`, {
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
      }
      setCheckedProductKeys([]);
      setSelectMode(false);
      setSelectedCode("");
      setNotice({ type: "success", text: text.deleteSuccess });
      void loadProducts();
    } catch (err) {
      setNotice({
        type: "error",
        text: err instanceof Error ? err.message : "Delete failed",
      });
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
    const hasName = editProduct.names.some(
      (n) => n.name && n.name.trim() !== "",
    );
    if (!hasName) {
      setNotice({ type: "error", text: text.nameRequired });
      return;
    }
    if (!editProduct.unitcode?.trim()) {
      setProductTab("units");
      setNotice({
        type: "error",
        text: `${text.unitBaseUnit}: ${text.required_error}`,
      });
      return;
    }
    const conversions = editProduct.unitconversions ?? [];
    if (conversions.some((unit) => !normalizeBusinessCode(unit.unitcode))) {
      setProductTab("units");
      setNotice({ type: "error", text: text.unitConversionRequired });
      return;
    }
    const unitCodes = [
      editProduct.unitcode,
      ...conversions.map((unit) => unit.unitcode),
    ].map(normalizeBusinessCode);
    if (new Set(unitCodes).size !== unitCodes.length) {
      setProductTab("units");
      setNotice({ type: "error", text: text.unitConversionDuplicate });
      return;
    }
    if (
      conversions.some(
        (unit) =>
          !Number.isSafeInteger(unit.dividevalue) ||
          !Number.isSafeInteger(unit.standvalue) ||
          unit.dividevalue <= 0 ||
          unit.standvalue <= 0,
      )
    ) {
      setProductTab("units");
      setNotice({ type: "error", text: text.unitConversionInvalidRatio });
      return;
    }

    setSaving(true);
    setNotice(null);
    try {
      const method = editorMode === "create" ? "POST" : "PUT";
      const url =
        editorMode === "create"
          ? "/api/product"
          : `/api/product/${encodeURIComponent(editProduct.guidfixed)}`;

      const payload = {
        ...editProduct,
        barcodes: undefined,
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
      if (
        editorMode === "create" &&
        bindBarcodeOnSave &&
        data.data?.guidfixed
      ) {
        const createdGuid = data.data.guidfixed;
        const createdCode = editProduct.code;

        const bcRes = await fetch(
          `/api/product-barcode/${encodeURIComponent(bindBarcodeOnSave.guidfixed)}`,
          {
            headers: {
              Authorization: `Bearer ${auth.token}`,
              "x-bc-backend-url": auth.backendUrl,
            },
          },
        );
        const bcJson = await bcRes.json();
        if (!bcRes.ok || !bcJson.success || !bcJson.data) {
          throw new Error(
            bcJson.message || "Failed to fetch barcode details for linking",
          );
        }

        const fullBarcode = bcJson.data;
        fullBarcode.itemguid = createdGuid;
        fullBarcode.itemcode = createdCode;

        const putRes = await fetch(
          `/api/product-barcode/${encodeURIComponent(bindBarcodeOnSave.guidfixed)}`,
          {
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
          },
        );
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
      setNotice({
        type: "error",
        text: err instanceof Error ? err.message : "Save failed",
      });
    } finally {
      setSaving(false);
    }
  };

  const openPicker = (
    type: string,
    targetField: string,
    anchorEl: HTMLElement | null = null,
  ) => {
    pickerAnchorRef.current = anchorEl;
    setPickerType(type);
    setPickerTarget(targetField);
    setPickerOpen(true);
  };

  const handlePickerSelect = (entry: MasterEntry) => {
    if (!editProduct) return;

    if (pickerTarget.startsWith("unit-conversion-")) {
      const idx = Number(pickerTarget.slice("unit-conversion-".length));
      const current = editProduct.unitconversions || [];
      setEditProduct({
        ...editProduct,
        unitconversions: current.map((row, rowIdx) =>
          rowIdx === idx
            ? {
                ...row,
                unitcode: entry.code,
                unitnames: entry.names,
              }
            : row,
        ),
      });
      return;
    }

    // Check if picker type is creditor (used for manufacturer and supplier multi-select)
    if (pickerType === "creditor") {
      if (pickerTarget === "manufacturers") {
        const current = editProduct.manufacturers || [];
        if (!current.some((m) => m.guidfixed === entry.guidfixed)) {
          setEditProduct({
            ...editProduct,
            manufacturers: [
              ...current,
              {
                guidfixed: entry.guidfixed,
                code: entry.code,
                names: entry.names,
              },
            ],
          });
        }
      } else if (pickerTarget === "suppliers") {
        const current = editProduct.suppliers || [];
        if (!current.some((s) => s.guidfixed === entry.guidfixed)) {
          setEditProduct({
            ...editProduct,
            suppliers: [
              ...current,
              {
                guidfixed: entry.guidfixed,
                code: entry.code,
                names: entry.names,
              },
            ],
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
        itemunitcode: entry.code,
        itemunitnames: entry.names,
      });
      return;
    }

    // Single select classification pickers — explicit field map to avoid unsafe keyof cast
    const classificationFields: Record<
      string,
      { guid: keyof Product; code: keyof Product; names: keyof Product }
    > = {
      group: {
        guid: "groupguid" as keyof Product,
        code: "groupcode",
        names: "groupnames",
      },
      subgroup: {
        guid: "subgroupguid" as keyof Product,
        code: "subgroupcode",
        names: "subgroupnames",
      },
      brand: {
        guid: "brandguid" as keyof Product,
        code: "brandcode",
        names: "brandnames",
      },
      category: {
        guid: "categoryguid" as keyof Product,
        code: "categorycode",
        names: "categorynames",
      },
      class: {
        guid: "classguid" as keyof Product,
        code: "classcode",
        names: "classnames",
      },
      design: {
        guid: "designguid" as keyof Product,
        code: "designcode",
        names: "designnames",
      },
      model: {
        guid: "modelguid" as keyof Product,
        code: "modelcode",
        names: "modelnames",
      },
      pattern: {
        guid: "patternguid" as keyof Product,
        code: "patterncode",
        names: "patternnames",
      },
      grade: {
        guid: "gradeguid" as keyof Product,
        code: "gradecode",
        names: "gradenames",
      },
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
        itemunitcode: "",
        itemunitnames: [],
      });
      return;
    }
    const clearFields: Record<
      string,
      { guid: keyof Product; code: keyof Product; names: keyof Product }
    > = {
      group: {
        guid: "groupguid" as keyof Product,
        code: "groupcode",
        names: "groupnames",
      },
      subgroup: {
        guid: "subgroupguid" as keyof Product,
        code: "subgroupcode",
        names: "subgroupnames",
      },
      brand: {
        guid: "brandguid" as keyof Product,
        code: "brandcode",
        names: "brandnames",
      },
      category: {
        guid: "categoryguid" as keyof Product,
        code: "categorycode",
        names: "categorynames",
      },
      class: {
        guid: "classguid" as keyof Product,
        code: "classcode",
        names: "classnames",
      },
      design: {
        guid: "designguid" as keyof Product,
        code: "designcode",
        names: "designnames",
      },
      model: {
        guid: "modelguid" as keyof Product,
        code: "modelcode",
        names: "modelnames",
      },
      pattern: {
        guid: "patternguid" as keyof Product,
        code: "patterncode",
        names: "patternnames",
      },
      grade: {
        guid: "gradeguid" as keyof Product,
        code: "gradecode",
        names: "gradenames",
      },
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
      manufacturers: current.filter((m) => m.guidfixed !== guid),
    });
  };

  const removeSupplier = (guid: string) => {
    if (!editProduct) return;
    const current = editProduct.suppliers || [];
    setEditProduct({
      ...editProduct,
      suppliers: current.filter((s) => s.guidfixed !== guid),
    });
  };

  const selectedTypeLabel = selectedProduct
    ? selectedProduct.itemtype === 2
      ? text.itemTypeSet
      : (itemTypes.find((item) => item.value === selectedProduct.itemtype)
          ?.label ?? String(selectedProduct.itemtype ?? "-"))
    : "-";
  const selectedMaterialLabel = selectedProduct
    ? (materialTypes.find((item) => item.value === selectedProduct.materialtype)
        ?.label ?? String(selectedProduct.materialtype ?? "-"))
    : "-";
  const selectedVatLabel = selectedProduct
    ? (vatTypes.find((item) => item.value === selectedProduct.vattype)?.label ??
      String(selectedProduct.vattype ?? "-"))
    : "-";
  const selectedUnitLabel = selectedProduct
    ? pickName(
        selectedProduct.unitnames || selectedProduct.itemunitnames,
        lang,
      ) ||
      selectedProduct.unitcode ||
      selectedProduct.itemunitcode ||
      "ยังไม่กำหนด"
    : "-";
  const selectedGroupLabel = selectedProduct?.groupcode
    ? `${selectedProduct.groupcode} — ${pickName(selectedProduct.groupnames, lang)}`
    : "ยังไม่จัดกลุ่ม";

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-background">
      {/* Header Toolbar */}
      <div className="flex shrink-0 items-center justify-between border-b border-border bg-card px-4 py-3">
        <div>
          <h2 className="text-lg font-bold">{text.productMenuName}</h2>
          <p className="text-xs text-muted-foreground">
            {lang === "th"
              ? "จัดการสินค้าและข้อมูลที่เกี่ยวข้องทั้งหมด"
              : "Manage products and related data"}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => void loadProducts()}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCcw className="h-4 w-4" />
            )}
            {text.refresh}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleCreateCopyOpen}
            disabled={
              loading ||
              !selectedProduct?.guidfixed ||
              detailLoading ||
              Boolean(detailError)
            }
          >
            <Copy className="h-4 w-4" />
            คัดลอก
          </Button>
          <Button
            variant="default"
            size="sm"
            onClick={handleCreateOpen}
            disabled={loading || !activeHoldingCode}
          >
            <Plus className="h-4 w-4" />
            {text.add}
          </Button>
        </div>
      </div>

      {/* Main split layout */}
      <div
        ref={splitContainerRef}
        data-testid="product-workbench"
        className={cn(
          "grid min-h-0 min-w-0 gap-3 xl:grid-cols-[minmax(0,var(--product-list-fr))_8px_minmax(0,var(--product-detail-fr))] xl:gap-0 flex-1 xl:h-full flex-col xl:flex-row",
        )}
        style={productSplitStyle}
      >
        {/* Left Side: Product List */}
        <Card
          data-testid="product-list-pane"
          className={cn(
            "min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col border-b xl:border-b-0 xl:border-r border-border bg-muted/10 rounded-none border-y-0 border-l-0 shadow-none bg-card",
            selectedCode || editorOpen ? "hidden xl:flex" : "flex",
          )}
        >
          <div className="p-3 border-b border-border">
            <div className="flex flex-wrap gap-2">
              <div className="relative min-w-[240px] flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  type="search"
                  placeholder={text.search}
                  className="!pl-10"
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                />
              </div>
              <Button
                variant={filterOpen ? "secondary" : "outline"}
                size="sm"
                type="button"
                onClick={() => setFilterOpen((current) => !current)}
              >
                <Filter className="h-4 w-4" />
                ตัวกรอง
              </Button>
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setShowListImage((current) => !current)}
              >
                {showListImage ? (
                  <ImageOff className="h-4 w-4" />
                ) : (
                  <ImageIcon className="h-4 w-4" />
                )}
                รูป
              </Button>
              <Button
                variant={selectMode ? "secondary" : "outline"}
                size="sm"
                type="button"
                onClick={() => {
                  setSelectMode((current) => !current);
                  setCheckedProductKeys([]);
                }}
              >
                {selectMode ? (
                  <X className="h-4 w-4" />
                ) : (
                  <CheckSquare className="h-4 w-4" />
                )}
                {selectMode ? "ยกเลิกเลือก" : "เลือกเพื่อลบ"}
              </Button>
              {selectMode ? (
                <Button
                  variant="outline"
                  size="sm"
                  type="button"
                  onClick={() => void handleDeleteSelectedProducts()}
                  disabled={checkedProductKeys.length === 0}
                >
                  <Trash2 className="h-4 w-4" />
                  ลบ {checkedProductKeys.length} รายการ
                </Button>
              ) : null}
            </div>
            {filterOpen ? (
              <div className="mt-3 flex flex-wrap gap-2 rounded-lg border border-border bg-muted/20 p-2">
                <Button
                  variant={
                    listItemTypeFilter === "all" ? "secondary" : "outline"
                  }
                  size="sm"
                  type="button"
                  onClick={() => setListItemTypeFilter("all")}
                >
                  ทั้งหมด
                </Button>
                {itemTypes.map((itemType) => (
                  <Button
                    key={itemType.value}
                    variant={
                      listItemTypeFilter === String(itemType.value)
                        ? "secondary"
                        : "outline"
                    }
                    size="sm"
                    type="button"
                    onClick={() =>
                      setListItemTypeFilter(String(itemType.value))
                    }
                  >
                    {itemType.label}
                  </Button>
                ))}
              </div>
            ) : null}
          </div>
          <div className="bc-list-toolbar shrink-0">
            <span>สินค้าทั้งหมด</span>
            <span>
              {visibleItems.length} / {items.length} รายการ
            </span>
          </div>

          {/* Table Header inside list on Desktop */}
          <div className="bc-list-header grid grid-cols-[minmax(76px,0.9fr)_minmax(0,1.7fr)_minmax(72px,0.8fr)] gap-x-2 shrink-0">
            <span>{text.itemCode ?? "รหัสสินค้า"}</span>
            <span>{text.productName ?? "ชื่อสินค้า"}</span>
            <span className="text-right">ยอดคงเหลือ</span>
          </div>

          <div
            data-testid="product-list-scroll"
            className="flex-1 overflow-y-auto min-h-[360px] xl:h-full xl:min-h-0"
          >
            {loading ? (
              <div className="p-8 text-center text-sm text-muted-foreground flex justify-center items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                {text.loading}
              </div>
            ) : visibleItems.length === 0 ? (
              <div className="p-8 text-center text-sm text-muted-foreground">
                ไม่พบข้อมูลสินค้า
              </div>
            ) : (
              visibleItems.map((item, index) => {
                const active = item.code === selectedCode;
                const rowKey = productRowKey(item);
                const isEditing = active && editorOpen && editorMode === "edit";
                const typeLabel =
                  item.itemtype === 2
                    ? text.itemTypeSet
                    : (itemTypes.find((t) => t.value === item.itemtype)
                        ?.label ?? String(item.itemtype));
                return (
                  <div
                    key={rowKey}
                    data-testid="product-row"
                    role="button"
                    tabIndex={0}
                    aria-label={`${item.code || text.itemCode} ${pickName(item.names, lang)}`}
                    aria-pressed={
                      selectMode ? checkedProductKeys.includes(rowKey) : active
                    }
                    className={cn(
                      "bc-list-row grid grid-cols-[minmax(76px,0.9fr)_minmax(0,1.7fr)_minmax(72px,0.8fr)] gap-x-2 py-2",
                      isEditing
                        ? "bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-amber-200/50"
                        : active
                          ? "bg-primary/10 hover:bg-primary/15"
                          : index % 2 === 0
                            ? "bg-background hover:bg-primary/5"
                            : "bg-muted/10 hover:bg-primary/5",
                    )}
                    onClick={() =>
                      selectMode
                        ? toggleCheckedProduct(rowKey)
                        : handleSelectProduct(item.code)
                    }
                    onKeyDown={(event) => {
                      if (event.key !== "Enter" && event.key !== " ") return;
                      event.preventDefault();
                      if (selectMode) toggleCheckedProduct(rowKey);
                      else handleSelectProduct(item.code);
                    }}
                  >
                    <div className="flex min-w-0 items-start gap-1.5">
                      {selectMode ? (
                        <span
                          className={cn(
                            "mt-0.5 grid size-5 shrink-0 place-items-center rounded-md border",
                            checkedProductKeys.includes(rowKey) &&
                              "border-primary bg-primary text-primary-foreground",
                          )}
                        >
                          {checkedProductKeys.includes(rowKey) ? (
                            <CheckSquare className="h-3 w-3" />
                          ) : null}
                        </span>
                      ) : null}
                      <span
                        className={cn(
                          "min-w-0 break-all font-semibold text-foreground",
                          isEditing && "text-amber-950 dark:text-amber-100",
                        )}
                      >
                        {item.code || "-"}
                      </span>
                    </div>
                    <div className="flex min-w-0 items-start gap-2">
                      {showListImage ? (
                        item.imageuri ? (
                          <AuthenticatedImg
                            alt=""
                            className="size-9 shrink-0 rounded-lg border border-border object-cover"
                            src={item.imageuri}
                            auth={auth}
                            fallback={
                              <span className="grid size-9 shrink-0 place-items-center rounded-lg border border-border text-muted-foreground">
                                <ImageOff className="h-3.5 w-3.5" />
                              </span>
                            }
                          />
                        ) : (
                          <span className="grid size-9 shrink-0 place-items-center rounded-lg border border-border text-muted-foreground">
                            <ImageOff className="h-3.5 w-3.5" />
                          </span>
                        )
                      ) : null}
                      <div className="min-w-0">
                        <div className="break-words font-medium leading-snug">
                          {pickName(item.names, lang) || "-"}
                        </div>
                        <div
                          className={cn(
                            "mt-0.5 break-words text-[10px] leading-tight text-muted-foreground",
                            isEditing &&
                              "text-amber-900/60 dark:text-amber-200/60",
                          )}
                        >
                          {typeLabel} · {formatProductUnitType(item)}
                        </div>
                      </div>
                    </div>
                    <div className="min-w-0 text-right">
                      <div
                        className={cn(
                          "break-words font-semibold",
                          (item.qty ?? 0) <= 0 && "text-destructive",
                        )}
                      >
                        {formatAutoPackingBalance(item, lang)}
                      </div>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </Card>

        {/* Resizable split separator bar */}
        <div
          aria-label="ปรับขนาดรายการสินค้าและรายละเอียดสินค้า"
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
        <div
          data-testid="product-detail-pane"
          className={cn(
            "min-h-0 flex-1 flex-col overflow-hidden bg-background p-3 xl:h-full xl:min-h-0",
            selectedCode || editorOpen ? "flex" : "hidden xl:flex",
          )}
        >
          {selectedCode && !editorOpen && (
            <Button
              variant="ghost"
              size="sm"
              className="mb-2 shrink-0 self-start xl:hidden flex items-center gap-2"
              onClick={() => setSelectedCode("")}
            >
              <ChevronLeft className="h-4 w-4" />
              {text.backToList}
            </Button>
          )}
          {editorOpen && editProduct ? (
            /* Product Edit Form */
            <form
              ref={productFormRef}
              onSubmit={handleSave}
              className="mx-auto flex h-full min-h-0 w-full max-w-6xl flex-col"
            >
              <div className="flex shrink-0 flex-col gap-3 border-b border-border pb-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex min-w-0 flex-wrap items-center gap-3">
                  <h3 className="text-xl font-bold">
                    {editorMode === "create"
                      ? text.productMasterCreateTitle
                      : text.productMasterEditTitle}
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
                <div className="flex shrink-0 flex-wrap gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={handleCancelEdit}
                    disabled={saving}
                  >
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
                      {saving ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Plus className="h-4 w-4" />
                      )}
                      {text.saveAndNew}
                    </Button>
                  )}
                  <Button type="submit" size="sm" disabled={saving}>
                    {saving ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4" />
                    )}
                    {text.save}
                  </Button>
                </div>
              </div>

              {/* Product Form Tab bar — flex-wrap so every tab stays visible (no hidden overflow) */}
              <div className="mt-2 flex shrink-0 flex-wrap gap-1 border-b border-border bg-muted/30 px-1 py-1.5">
                {(
                  Object.entries(PRIMARY_PRODUCT_TABS(text)) as [
                    (
                      | "basic"
                      | "classification"
                      | "units"
                      | "bom"
                      | "stock"
                      | "media"
                    ),
                    string,
                  ][]
                ).map(([k, label]) => {
                  const isActive = productTab === k;
                  return (
                    <button
                      key={k}
                      type="button"
                      onClick={() => setProductTab(k)}
                      className={cn(
                        "shrink-0 rounded-md px-3 py-1.5 text-xs font-medium transition",
                        isActive
                          ? "bg-primary text-primary-foreground shadow-sm"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground",
                      )}
                    >
                      {label}
                    </button>
                  );
                })}
                {// Show every tab directly — no "ขั้นสูง" overflow toggle
                // (per Jead 2026-07-19: ไม่ต้องมีปุ่มขั้นสูง ให้แสดงเมนูทุกตัวเลย).
                (
                  Object.entries(ADVANCED_PRODUCT_TABS(text)) as [
                    (
                      | "logistics"
                      | "restaurant"
                      | "timeforsales"
                      | "business"
                      | "misc"
                    ),
                    string,
                  ][]
                ).map(([k, label]) => {
                  const isActive = productTab === k;
                  return (
                    <button
                      key={k}
                      type="button"
                      onClick={() => setProductTab(k)}
                      className={cn(
                        "shrink-0 rounded-md px-3 py-1.5 text-xs font-medium transition",
                        isActive
                          ? "bg-primary text-primary-foreground shadow-sm"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground",
                      )}
                    >
                      {label}
                    </button>
                  );
                })}
              </div>

              <div className="min-h-0 flex-1 overflow-y-auto py-3 pr-1">
                {/* Tab: basic */}
                {productTab === "basic" && (
                  <TabProductBasic
                    value={editProduct}
                    onChange={setEditProduct}
                    text={text}
                    lang={lang}
                    activeLanguages={activeLanguages}
                    editorMode={editorMode}
                    itemTypes={itemTypes}
                    materialTypes={materialTypes}
                    vatTypes={vatTypes}
                    openPicker={openPicker}
                    removeManufacturer={removeManufacturer}
                    removeSupplier={removeSupplier}
                  />
                )}

                {/* Tab: classification */}
                {productTab === "classification" && (
                  <TabProductClassification
                    value={editProduct}
                    text={text}
                    lang={lang}
                    openPicker={openPicker}
                    clearPickerField={clearPickerField}
                  />
                )}

                {/* Tab: units */}
                {productTab === "units" && (
                  <TabProductUnits
                    value={editProduct}
                    onChange={setEditProduct}
                    lang={lang}
                    openPicker={openPicker}
                  />
                )}

                {/* Tab: bom */}
                {productTab === "bom" && (
                  <TabProductBom
                    value={editProduct}
                    onChange={setEditProduct}
                    lang={lang}
                  />
                )}

                {/* Tab: stock */}
                {productTab === "stock" && (
                  <TabProductStock
                    value={editProduct}
                    onChange={setEditProduct}
                    text={text}
                  />
                )}

                {/* Tab: media */}
                {productTab === "media" && (
                  <TabProductMedia
                    value={editProduct}
                    onChange={setEditProduct}
                    auth={auth}
                    language={lang}
                  />
                )}

                {/* Tab: restaurant */}
                {productTab === "restaurant" && (
                  <TabProductRestaurant
                    value={editProduct}
                    onChange={setEditProduct}
                    auth={auth}
                    language={lang}
                    shopLanguages={activeLanguages}
                  />
                )}

                {/* Tab: timeforsales */}
                {productTab === "timeforsales" && (
                  <TabProductTimeForSale
                    value={editProduct}
                    onChange={setEditProduct}
                    language={lang}
                  />
                )}

                {/* Tab: business */}
                {productTab === "business" && (
                  <TabProductBusinessBranch
                    value={editProduct}
                    onChange={setEditProduct}
                    auth={auth}
                    language={lang}
                  />
                )}

                {/* Tab: logistics */}
                {productTab === "logistics" && (
                  <TabProductLogistics
                    value={editProduct}
                    onChange={setEditProduct}
                  />
                )}

                {/* Tab: misc */}
                {productTab === "misc" && (
                  <TabProductMisc
                    value={editProduct}
                    onChange={setEditProduct}
                    language={lang}
                  />
                )}
              </div>
            </form>
          ) : selectedProduct ? (
            <Card
              data-testid="product-detail-card"
              aria-busy={detailLoading}
              className="relative flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card shadow-sm"
            >
              <div className="absolute left-0 right-0 top-0 h-1 bg-primary" />

              <CardHeader className="shrink-0 border-b border-border bg-muted/5 p-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="flex min-w-0 items-start gap-3">
                    {selectedProduct.imageuri ? (
                      <AuthenticatedImg
                        alt={
                          pickName(selectedProduct.names, lang) ||
                          selectedProduct.code
                        }
                        className="size-16 shrink-0 rounded-xl border border-border bg-background object-cover shadow-sm"
                        src={selectedProduct.imageuri}
                        auth={auth}
                        fallback={
                          <span className="grid size-16 shrink-0 place-items-center rounded-xl border border-border bg-muted text-muted-foreground">
                            <ImageIcon className="h-6 w-6" />
                          </span>
                        }
                      />
                    ) : (
                      <span className="grid size-16 shrink-0 place-items-center rounded-xl border border-border bg-muted text-muted-foreground">
                        <ImageIcon className="h-6 w-6" />
                      </span>
                    )}
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-1.5">
                        <Badge variant="secondary">{selectedTypeLabel}</Badge>
                        <Badge variant="outline">{selectedMaterialLabel}</Badge>
                      </div>
                      <h3 className="mt-2 break-all text-sm font-extrabold tracking-wide text-primary">
                        {selectedProduct.code}
                      </h3>
                      <p className="mt-0.5 break-words text-xl font-bold leading-tight text-foreground">
                        {pickName(selectedProduct.names, lang) ||
                          "ยังไม่ได้ระบุชื่อสินค้า"}
                      </p>
                    </div>
                  </div>
                  <div className="flex shrink-0 flex-wrap gap-2 sm:justify-end">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleEditOpen(selectedProduct)}
                      disabled={
                        loading ||
                        detailLoading ||
                        Boolean(detailError) ||
                        !selectedProduct.guidfixed
                      }
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      {text.edit}
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => void handleDelete(selectedProduct)}
                      disabled={
                        loading ||
                        detailLoading ||
                        Boolean(detailError) ||
                        !selectedProduct.guidfixed
                      }
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      {text.delete}
                    </Button>
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-2 gap-2 lg:grid-cols-4">
                  <DetailSummary
                    label="ยอดคงเหลือ"
                    value={formatAutoPackingBalance(selectedProduct, lang)}
                    emphasize={(selectedProduct.qty ?? 0) <= 0}
                  />
                  <DetailSummary label="หน่วยหลัก" value={selectedUnitLabel} />
                  <DetailSummary
                    label="กลุ่มสินค้า"
                    value={selectedGroupLabel}
                  />
                  <DetailSummary
                    label="ภาษีมูลค่าเพิ่ม"
                    value={selectedVatLabel}
                  />
                </div>
              </CardHeader>

              <>
                {detailError ? (
                  <div
                    role="alert"
                    className="flex shrink-0 items-start gap-2 border-b border-destructive/30 bg-destructive/5 px-3 py-2 text-xs"
                  >
                    <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
                    <div className="min-w-0 flex-1">
                      <p className="font-bold text-foreground">
                        โหลดข้อมูลฉบับเต็มไม่สำเร็จ —
                        ยังคงแสดงข้อมูลล่าสุดที่โหลดสำเร็จ
                      </p>
                      <p className="mt-0.5 break-words text-muted-foreground">
                        {detailError}
                      </p>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 shrink-0 text-xs"
                      onClick={() =>
                        setDetailReloadKey((current) => current + 1)
                      }
                    >
                      <RefreshCcw className="h-3.5 w-3.5" />
                      ลองใหม่
                    </Button>
                  </div>
                ) : null}
                <div className="flex shrink-0 items-center justify-end border-b border-border bg-background px-3 py-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 shrink-0 text-xs"
                    onClick={() => setShowEmptyDetails((current) => !current)}
                  >
                    {showEmptyDetails ? (
                      <EyeOff className="h-3.5 w-3.5" />
                    ) : (
                      <Eye className="h-3.5 w-3.5" />
                    )}
                    {showEmptyDetails ? "ซ่อนข้อมูลว่าง" : "แสดงข้อมูลทั้งหมด"}
                  </Button>
                </div>

                <CardContent
                  id="product-detail-content"
                  data-testid="product-detail-content"
                  className="min-h-0 flex-1 overflow-y-auto bg-muted/10 p-3"
                >
                  <div className="mb-3 grid gap-3 2xl:grid-cols-2">
                    <BusinessImageGallery
                      auth={auth}
                      sources={[
                        {
                          key:
                            selectedProduct.guidfixed || selectedProduct.code,
                          label: `${selectedProduct.code} — ${pickName(selectedProduct.names, lang)}`,
                          imageuri: selectedProduct.imageuri,
                          images: selectedProduct.images,
                          videos: selectedProduct.videos,
                        },
                      ]}
                      title={lang === "th" ? "สื่อสินค้า" : "Product media"}
                    />
                    <BusinessImageGallery
                      auth={auth}
                      sources={(selectedProduct.barcodes ?? []).map(
                        (barcode) => ({
                          key: barcode.guidfixed || barcode.barcode,
                          label: `${lang === "th" ? "บาร์โค้ด" : "Barcode"} ${barcode.barcode}${barcode.itemunitcode ? ` · ${barcode.itemunitcode}` : ""}`,
                          imageuri: barcode.imageuri,
                          images: barcode.images,
                          videos: barcode.videos,
                        }),
                      )}
                      title={
                        lang === "th"
                          ? "สื่อจากบาร์โค้ด"
                          : "Media from barcodes"
                      }
                    />
                  </div>
                  <div className="grid gap-3 2xl:grid-cols-2">
                    <DetailSection
                      title={text.basicInfoCard ?? "ข้อมูลพื้นฐานสินค้า"}
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        { label: text.itemType, value: selectedTypeLabel },
                        {
                          label: text.materialType,
                          value: selectedMaterialLabel,
                        },
                        {
                          label: "สถานะภาษีมูลค่าเพิ่ม",
                          value: selectedVatLabel,
                        },
                        {
                          label: "รหัสประเภทภาษี",
                          value: String(
                            selectedProduct.taxtype ??
                              selectedProduct.vattype ??
                              "-",
                          ),
                        },
                        {
                          label: "รหัสหน่วยหลัก",
                          value:
                            selectedProduct.unitcode ||
                            selectedProduct.itemunitcode ||
                            "-",
                        },
                        {
                          label: "ชื่อหน่วยหลัก",
                          value:
                            pickName(
                              selectedProduct.unitnames ||
                                selectedProduct.itemunitnames,
                              lang,
                            ) || "-",
                        },
                        {
                          label: "มิติสินค้า",
                          value: formatDimensionList(
                            selectedProduct.dimensions,
                            lang,
                          ),
                        },
                        {
                          label: "ประเภทสินค้า (ระบบ)",
                          value: selectedProduct.producttype?.code
                            ? `${selectedProduct.producttype.code} — ${pickName(selectedProduct.producttype.names, lang)}`
                            : "-",
                        },
                        {
                          label: "วิธีคำนวณ VAT",
                          value: String(selectedProduct.vatcal ?? 0),
                        },
                      ]}
                    />
                    <DetailSection
                      title="คู่ค้าและรายละเอียด"
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: text.manufacturers,
                          value: formatNamedList(
                            selectedProduct.manufacturers,
                            lang,
                          ),
                        },
                        {
                          label: text.suppliers,
                          value: formatNamedList(
                            selectedProduct.suppliers,
                            lang,
                          ),
                        },
                        {
                          label: "รายละเอียดสินค้า",
                          value: selectedProduct.description || "-",
                        },
                        {
                          label: "คำเตือน",
                          value: selectedProduct.alertdescription || "-",
                        },
                      ]}
                    />
                  </div>

                  <DetailSection
                    title={text.groupingCard ?? "ข้อมูลการจัดกลุ่ม"}
                    showEmptyFields={showEmptyDetails}
                    fields={[
                      {
                        label: text.group,
                        value: selectedProduct.groupcode
                          ? `${selectedProduct.groupcode} — ${pickName(selectedProduct.groupnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.groupsubone,
                        value: selectedProduct.subgroupcode
                          ? `${selectedProduct.subgroupcode} — ${pickName(selectedProduct.subgroupnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.brand,
                        value: selectedProduct.brandcode
                          ? `${selectedProduct.brandcode} — ${pickName(selectedProduct.brandnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.category,
                        value: selectedProduct.categorycode
                          ? `${selectedProduct.categorycode} — ${pickName(selectedProduct.categorynames, lang)}`
                          : "-",
                      },
                      {
                        label: text.class,
                        value: selectedProduct.classcode
                          ? `${selectedProduct.classcode} — ${pickName(selectedProduct.classnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.design,
                        value: selectedProduct.designcode
                          ? `${selectedProduct.designcode} — ${pickName(selectedProduct.designnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.model,
                        value: selectedProduct.modelcode
                          ? `${selectedProduct.modelcode} — ${pickName(selectedProduct.modelnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.pattern,
                        value: selectedProduct.patterncode
                          ? `${selectedProduct.patterncode} — ${pickName(selectedProduct.patternnames, lang)}`
                          : "-",
                      },
                      {
                        label: text.grade,
                        value: selectedProduct.gradecode
                          ? `${selectedProduct.gradecode} — ${pickName(selectedProduct.gradenames, lang)}`
                          : "-",
                      },
                    ]}
                  />

                  <div className="grid gap-3 2xl:grid-cols-2">
                    <DetailSection
                      title={text.tabStock ?? "การควบคุมคลังสินค้า"}
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: text.qty,
                          value: formatAutoPackingBalance(
                            selectedProduct,
                            lang,
                          ),
                        },
                        {
                          label: text.orderPoint,
                          value: String(selectedProduct.orderpoint ?? 0),
                        },
                        {
                          label: text.minPoint,
                          value: String(selectedProduct.minpoint ?? 0),
                        },
                        {
                          label: text.maxPoint,
                          value: String(selectedProduct.maxpoint ?? 0),
                        },
                        {
                          label: text.stockBarcode,
                          value: selectedProduct.stockbarcode || "-",
                        },
                        {
                          label: "ต้นทุนคงที่",
                          value:
                            selectedProduct.fixedcost != null
                              ? String(selectedProduct.fixedcost)
                              : "-",
                        },
                      ]}
                    />
                    <DetailSection
                      title={text.tabUnitsBarcode ?? "หน่วยนับและบาร์โค้ด"}
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: "หน่วยนับมาตรฐาน",
                          value: [
                            selectedProduct.unitcode,
                            pickName(selectedProduct.unitnames, lang),
                            "1:1",
                          ]
                            .filter(Boolean)
                            .join(" — "),
                        },
                        {
                          label: "หน่วยนับเพิ่มเติม",
                          value: formatUnitConversionList(
                            selectedProduct.unitconversions,
                            lang,
                          ),
                        },
                        {
                          label: "บาร์โค้ดสินค้า",
                          value: formatRefBarcodeList(
                            selectedProduct.barcodes,
                            lang,
                          ),
                        },
                        {
                          label: "ส่วนประกอบ BOM",
                          value: formatBomList(selectedProduct.bom, lang),
                        },
                      ]}
                    />
                    <DetailSection
                      title="ขนาดและน้ำหนักพัสดุ"
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: "น้ำหนักรวมพัสดุ",
                          value: `${selectedProduct.packageweight ?? 0} kg`,
                        },
                        {
                          label: "มิติตัวกล่อง (ก × ย × ส)",
                          value: `${selectedProduct.packagewidth ?? 0} × ${selectedProduct.packagelength ?? 0} × ${selectedProduct.packageheight ?? 0} cm`,
                        },
                        {
                          label: "น้ำหนักปริมาตร (ประเมิน)",
                          value: `${(((selectedProduct.packagewidth ?? 0) * (selectedProduct.packagelength ?? 0) * (selectedProduct.packageheight ?? 0)) / 5000).toFixed(3)} kg`,
                        },
                        {
                          label: "คุณลักษณะพิเศษ",
                          value: selectedProduct.isalert ? "ระวังแตก" : "-",
                        },
                      ]}
                    />
                  </div>

                  <div className="grid gap-3 2xl:grid-cols-2">
                    <DetailSection
                      title={text.tabRestaurant ?? "ร้านอาหาร/POS"}
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: text.isForRestaurant,
                          value: formatYesNo(
                            selectedProduct.restaurant?.isforrestaurant,
                          ),
                        },
                        {
                          label: text.isForTakeaway,
                          value: formatYesNo(
                            selectedProduct.restaurant?.isfortakeaway,
                          ),
                        },
                        {
                          label: text.isForDelivery,
                          value: formatYesNo(
                            selectedProduct.restaurant?.isfordelivery,
                          ),
                        },
                        {
                          label: text.isForCustomer,
                          value: formatYesNo(
                            selectedProduct.restaurant?.isforcustomer,
                          ),
                        },
                        {
                          label: text.isForCustomerPreOrder,
                          value: formatYesNo(
                            selectedProduct.restaurant?.isforcustomerpreorder,
                          ),
                        },
                        {
                          label: text.isALaCarte,
                          value: formatYesNo(selectedProduct.isalacarte),
                        },
                        {
                          label: text.isStockForRestaurant,
                          value: formatYesNo(
                            selectedProduct.isstockforrestaurant,
                          ),
                        },
                        {
                          label: text.isSplitUnitPrint,
                          value: formatYesNo(selectedProduct.issplitunitprint),
                        },
                        {
                          label: text.isOnlyStaff,
                          value: formatYesNo(selectedProduct.isonlystaff),
                        },
                        {
                          label: text.foodType,
                          value: selectedProduct.restaurant?.isforrestaurant
                            ? (foodTypes.find(
                                (item) =>
                                  item.value === selectedProduct.foodtype,
                              )?.label ??
                              String(selectedProduct.foodtype ?? "-"))
                            : "-",
                        },
                        {
                          label: "บริการสั่งอาหาร",
                          value: formatNamedList(
                            selectedProduct.ordertypes,
                            lang,
                          ),
                        },
                        {
                          label: "ชุดตัวเลือกสินค้า",
                          value: formatOptionList(
                            selectedProduct.options,
                            lang,
                          ),
                        },
                        {
                          label: "สะสมแต้ม",
                          value: formatYesNo(selectedProduct.issumpoint),
                        },
                        {
                          label: "ส่วนลดสูงสุด",
                          value: selectedProduct.maxdiscount || "-",
                        },
                        {
                          label: "ส่วนลด",
                          value: selectedProduct.discount || "-",
                        },
                        {
                          label: "หักส่วนลด ณ จุดขาย",
                          value: formatYesNo(
                            selectedProduct.isdiscountpointofpurchase,
                          ),
                        },
                        {
                          label: "เงินปันผล",
                          value: formatYesNo(selectedProduct.isdividend),
                        },
                      ]}
                    />
                    <DetailSection
                      title={`${text.tabTimeForSales} / ${text.tabBusinessBranchShort}`}
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: text.tabTimeForSales,
                          value: formatTimeForSaleList(
                            selectedProduct.timeforsales,
                          ),
                        },
                        {
                          label: text.businessTypes ?? "ประเภทธุรกิจ",
                          value: formatNamedList(
                            selectedProduct.businesstypes,
                            lang,
                          ),
                        },
                        {
                          label: text.ignoreBranches ?? "สาขาที่ยกเว้น",
                          value: formatNamedList(
                            selectedProduct.ignorebranches,
                            lang,
                          ),
                        },
                      ]}
                    />
                  </div>

                  <div className="grid gap-3 2xl:grid-cols-2">
                    <DetailSection
                      title={`${text.tabMedia} / Marketplace`}
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: "ใช้รูปหรือสี",
                          value: formatYesNo(selectedProduct.useimageorcolor),
                        },
                        {
                          label: "สี",
                          value:
                            selectedProduct.colorselect ||
                            selectedProduct.colorselecthex ||
                            "-",
                        },
                        {
                          label: "รูปหลัก",
                          value: selectedProduct.imageuri ? "มีรูปหลัก" : "-",
                        },
                        {
                          label: "รูปทั้งหมด",
                          value: selectedProduct.images?.length
                            ? `${selectedProduct.images.length} รูป`
                            : "-",
                        },
                        {
                          label: "Marketplace",
                          value: formatMarketplaceProductList(
                            selectedProduct.marketplaceproducts,
                          ),
                        },
                        {
                          label: "คำเตือน",
                          value: selectedProduct.alertdescription || "-",
                        },
                        {
                          label: "รายละเอียด",
                          value: selectedProduct.description || "-",
                        },
                      ]}
                    />
                    <DetailSection
                      title="ข้อมูลระบบ"
                      showEmptyFields={showEmptyDetails}
                      fields={[
                        {
                          label: "รหัสภายในสินค้า",
                          value: selectedProduct.guidfixed || "-",
                        },
                        {
                          label: "รหัสกลุ่มกิจการ",
                          value: selectedProduct.holdingcode || "-",
                        },
                        {
                          label: "GUID หน่วยนับ",
                          value: selectedProduct.unitguid || "-",
                        },
                        {
                          label: "เปิดคำเตือน",
                          value: formatYesNo(selectedProduct.isalert),
                        },
                      ]}
                    />
                  </div>
                </CardContent>
              </>
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
        title={
          pickerType
            ? `ค้นหา ${text[pickerType as keyof typeof text] || pickerType}`
            : ""
        }
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
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowBarcodePicker(false)}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="border-b border-border px-4 py-3">
              <div className="relative">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  autoFocus
                  type="search"
                  value={barcodeSearchInput}
                  onChange={(event) =>
                    setBarcodeSearchInput(event.target.value)
                  }
                  placeholder={text.searchBarcodeOrName}
                  className="h-9 !pl-10"
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
                  {barcodeSearch
                    ? text.noUnlinkedBarcode
                    : text.allBarcodesLinked}
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
                            {text.unitLabel}: {b.itemunitcode || "-"}{" "}
                            {b.itemunitnames && b.itemunitnames.length > 0
                              ? `(${pickName(b.itemunitnames, lang)})`
                              : ""}
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
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setShowBarcodePicker(false)}
              >
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

function setNameXEntry(
  list: NameX[] | undefined,
  code: string,
  name: string,
): NameX[] {
  const arr = list ? [...list] : [];
  const idx = arr.findIndex((x) => x.code === code);
  if (idx >= 0) {
    arr[idx] = { ...arr[idx], name };
  } else {
    arr.push({ code, name });
  }
  return arr;
}

function productRowKey(item: Product): string {
  return item.code || item.guidfixed;
}

function productUnitRows(item: Product): ProductUnitConversion[] {
  return item.unitconversions ?? [];
}

function formatProductUnitType(item: Product): string {
  return productUnitRows(item).length > 0 ? "หลายหน่วยนับ" : "หน่วยนับเดียว";
}

function formatAutoPackingBalance(item: Product, language: string): string {
  const total = Math.max(0, Math.floor(Number(item.qty ?? 0)));
  const baseUnit =
    pickName(item.unitnames || item.itemunitnames, language) ||
    item.unitcode ||
    item.itemunitcode ||
    "หน่วย";
  const unitRows = productUnitRows(item)
    .map((row) => ({
      name: pickName(row.unitnames, language) || row.unitcode,
      size: Number(row.standvalue) / Number(row.dividevalue),
    }))
    .filter((row) => row.name && Number.isFinite(row.size) && row.size > 0)
    .sort((a, b) => b.size - a.size);

  const units = [...unitRows, { name: baseUnit, size: 1 }];
  if (total === 0) return `0 ${baseUnit}`;

  let remaining = total;
  const parts: string[] = [];
  for (const unit of units) {
    const count = Math.floor(remaining / unit.size);
    if (count <= 0) continue;
    parts.push(`${count.toLocaleString("th-TH")} ${unit.name}`);
    remaining -= count * unit.size;
  }
  if (remaining > 0)
    parts.push(`${remaining.toLocaleString("th-TH")} ${baseUnit}`);
  return parts.length > 0
    ? parts.join(" x ")
    : `${total.toLocaleString("th-TH")} ${baseUnit}`;
}

// ─── Tab Components ───────────────────────────────────────────────────────

type DetailFieldItem = {
  label: string;
  value: string;
};

function formatYesNo(value?: boolean) {
  return value ? "ใช่" : "ไม่ใช่";
}

function formatNamedList(
  items: Array<{ code?: string; names?: NameX[] }> | undefined,
  language: string,
) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) =>
      [item.code, pickName(item.names, language)].filter(Boolean).join(" — "),
    )
    .filter(Boolean)
    .join(", ");
}

function formatUnitConversionList(
  items: ProductUnitConversion[] | undefined,
  language: string,
) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) => {
      const unit = [item.unitcode, pickName(item.unitnames, language)]
        .filter(Boolean)
        .join(" — ");
      return `${unit} (${item.standvalue}:${item.dividevalue})`;
    })
    .join(", ");
}

function formatRefBarcodeList(
  items: RefProductBarcode[] | undefined,
  language: string,
) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) => {
      const barcodeText = [
        item.barcode,
        pickName(item.names, language),
        item.itemunitcode,
      ]
        .filter(Boolean)
        .join(" / ");
      const marketplaceText = (item.marketplaceskumappings || [])
        .map((mapping) => {
          const stockText = (mapping.marketplacedimensionstocks || [])
            .map(
              (stock) =>
                `${stock.dimensionname || stock.dimensionkey}: พร้อมขาย ${stock.availableqty}`,
            )
            .join("; ");
          return [
            [mapping.platform, mapping.holdingcode, mapping.sellersku]
              .filter(Boolean)
              .join(" / "),
            stockText,
          ]
            .filter(Boolean)
            .join(" => ");
        })
        .filter(Boolean)
        .join(" | ");
      return marketplaceText
        ? `${barcodeText}\n${marketplaceText}`
        : barcodeText;
    })
    .join("\n");
}

function formatBomList(
  items: BOMProductBarcode[] | undefined,
  language: string,
) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) => {
      const name = pickName(item.names, language);
      const qty = item.qty == null ? "" : ` x ${item.qty}`;
      return (
        [item.barcode, name, item.itemunitcode].filter(Boolean).join(" / ") +
        qty
      );
    })
    .join(", ");
}

function formatOptionList(
  items: ProductOption[] | undefined,
  language: string,
) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) => {
      const name = pickName(item.names, language) || "ชุดตัวเลือก";
      const choiceText = (item.choices || [])
        .map((choice) => {
          const choiceName =
            pickName(choice.names, language) || choice.refbarcode || "ตัวเลือก";
          const price = choice.price ? `ราคา ${choice.price}` : "";
          const qty = choice.qty == null ? "" : `จำนวน ${choice.qty}`;
          return [choiceName, price, qty].filter(Boolean).join(" ");
        })
        .join("; ");
      return choiceText ? `${name}: ${choiceText}` : name;
    })
    .join(", ");
}

function formatDimensionList(items: Product["dimensions"], language: string) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) => {
      const dimensionName = pickName(item.names, language) || item.guidfixed;
      const choiceName =
        pickName(item.item?.names, language) || item.item?.guidfixed || "-";
      return `${dimensionName}: ${choiceName}${item.isdisabled || item.item?.isdisabled ? " (ปิดใช้)" : ""}`;
    })
    .join(", ");
}

function formatMarketplaceProductList(items: Product["marketplaceproducts"]) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) =>
      [
        item.platform,
        item.accountid,
        item.marketitemid,
        item.sellersku || item.shopsku,
        item.status,
        item.syncenabled ? "sync" : "ไม่ sync",
      ]
        .filter(Boolean)
        .join(" / "),
    )
    .join("\n");
}

function formatTimeForSaleList(items: ProductTimeForSale[] | undefined) {
  if (!items || items.length === 0) return "-";
  return items
    .map((item) => {
      const dateRange = [item.fromdate, item.todate].filter(Boolean).join("-");
      const timeRange = [item.fromtime, item.totime].filter(Boolean).join("-");
      return [dateRange, timeRange].filter(Boolean).join(" ");
    })
    .filter(Boolean)
    .join(", ");
}

function isEmptyDetailValue(value: string) {
  const normalized = value.trim().toLowerCase();
  return (
    normalized === "" ||
    normalized === "-" ||
    normalized === "—" ||
    normalized === "ไม่ใช่" ||
    normalized === "0 kg" ||
    normalized === "0.000 kg" ||
    /^0\s*[×x]\s*0\s*[×x]\s*0\s*cm$/.test(normalized)
  );
}

function DetailSummary({
  emphasize = false,
  label,
  value,
}: {
  emphasize?: boolean;
  label: string;
  value: string;
}) {
  return (
    <div
      className={cn(
        "min-w-0 rounded-lg border border-border bg-background px-3 py-2",
        emphasize && "border-destructive/30 bg-destructive/5",
      )}
    >
      <p className="text-[10px] font-semibold text-muted-foreground">{label}</p>
      <p
        className={cn(
          "mt-0.5 break-words text-sm font-bold leading-snug text-foreground",
          emphasize && "text-destructive",
        )}
      >
        {value || "-"}
      </p>
    </div>
  );
}

function DetailSection({
  fields,
  showEmptyFields = false,
  title,
}: {
  fields: DetailFieldItem[];
  showEmptyFields?: boolean;
  title: string;
}) {
  const visibleFields = showEmptyFields
    ? fields
    : fields.filter((field) => !isEmptyDetailValue(field.value));

  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="border-b border-border bg-background px-3 py-2.5">
        <h3 className="text-sm font-bold text-primary">{title}</h3>
      </div>
      {visibleFields.length > 0 ? (
        <dl className="grid md:grid-cols-2">
          {visibleFields.map((field, index) => (
            <DetailField
              key={`${title}-${field.label}-${index}`}
              label={field.label}
              value={field.value}
            />
          ))}
        </dl>
      ) : (
        <p className="px-3 py-5 text-center text-sm text-muted-foreground">
          ยังไม่มีข้อมูลในส่วนนี้
        </p>
      )}
    </section>
  );
}

function DetailField({ label, value }: DetailFieldItem) {
  return (
    <div className="min-w-0 border-b border-border/60 px-3 py-2.5 md:odd:border-r">
      <dt className="text-[10px] font-semibold leading-tight text-muted-foreground">
        {label}
      </dt>
      <dd className="mt-1 whitespace-pre-line break-words text-sm font-semibold leading-snug text-foreground">
        {value || "-"}
      </dd>
    </div>
  );
}

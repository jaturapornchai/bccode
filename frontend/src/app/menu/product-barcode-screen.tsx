"use client";

import {
  Barcode,
  CheckSquare,
  Copy,
  Filter,
  GitFork,
  History,
  ImageIcon,
  ImageOff,
  Loader2,
  Package,
  Pencil,
  Plus,
  Printer,
  RefreshCcw,
  Search,
  Trash2,
  X,
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type FormEvent,
  type KeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type PointerEvent,
} from "react";
import { AuthenticatedImg } from "@/components/authenticated-image";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ProductBarcodeFormDialog } from "@/components/product-barcode/barcode-form";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import {
  emptyProductBarcode,
  type ProductBarcode as ProductBarcodeObject,
} from "@/lib/product-barcode/types";
import {
  createBarcode,
  deleteBarcodes,
  listBarcodes,
  updateBarcode,
} from "@/lib/product-barcode/api";
import {
  rawToProductBarcode,
  toNumberOrNull,
} from "@/lib/product-barcode/utils";
import { cn } from "@/lib/utils";
import {
  localizedName,
  type AuthSession,
  type LocalizedName,
  type WorkspaceSession,
  workspaceStorageKeys,
  WORKSPACE_CHANGED_EVENT,
} from "@/lib/workspace-models";
import { languageCodesFromWorkspace } from "@/components/product-barcode/names-editor";
import {
  getBarcodeText,
  type BarcodeText,
} from "@/lib/product-barcode/language";
import { pushNotice } from "@/lib/toast";

type ProductBarcodeScreenProps = {
  embedded?: boolean;
  language?: LanguageCode;
};

type ProductBarcodeRecord = {
  raw: Record<string, unknown>;
  guidFixed: string;
  holdingCode: string;
  barcode: string;
  barcodeRef: string;
  name: string;
  unitName: string;
  unitCode: string;
  itemCode: string;
  itemGuid: string;
  itemGuidFixed: string;
  parentGuid: string;
  groupName: string;
  groupCode: string;
  brandName: string;
  brandCode: string;
  categoryName: string;
  categoryCode: string;
  className: string;
  classCode: string;
  designName: string;
  designCode: string;
  gradeName: string;
  gradeCode: string;
  modelName: string;
  modelCode: string;
  patternName: string;
  patternCode: string;
  groupSubOneName: string;
  groupSubOneCode: string;
  groupSubTwoName: string;
  groupSubTwoCode: string;
  manufacturerName: string;
  manufacturerCode: string;
  shelfCode: string;
  shelfName: string;
  checksum: string;
  price: number;
  imageUri: string;
  colorSelect: string;
  colorSelectHex: string;
  unitCount: number;
  allUnitNames: string;
  balanceQty: number;
  balanceFormatted: string;
  standValue: number;
  divideValue: number;
  itemType: number;
  productType: string;
  foodType: number;
  materialType: number;
  taxType: number;
  vatType: number;
  vatCal: number;
  isStock: number;
  isMainBarcode: boolean;
  isMainItem: boolean;
  isUseSubBarcodes: boolean;
  useImageOrColor: boolean;
  condition: boolean;
  isSumPoint: boolean;
  isDividend: boolean;
  isALaCarte: boolean;
  isSplitUnitPrint: boolean;
  isOnlyStaff: boolean;
  isStockForRestaurant: boolean;
  isDiscountPointOfPurchase: boolean;
  isAlert: boolean;
  isDisable: boolean;
  showIsDividend: string;
  rowNumber: number;
  maxDiscount: string;
  discount: string;
  description: string;
  alertDescription: string;
  priceCount: number;
  refBarcodeCount: number;
  subBarcodeCount: number;
  bomCount: number;
  optionCount: number;
  orderTypeCount: number;
  dimensionCount: number;
  businessTypeCount: number;
  ignoreBranchCount: number;
  branchCount: number;
  categoryCount: number;
  timeForSaleCount: number;
  fixedCostCount: number;
};

const PRODUCT_SPLIT_DEFAULT_LEFT = 30;
const PRODUCT_SPLIT_STORAGE_KEY = "bcproductbarcodesplitleftv2";
const PRODUCT_SPLIT_MIN_LEFT = 5;
const PRODUCT_SPLIT_MAX_LEFT = 95;

function clampProductSplitLeft(value: number) {
  if (!Number.isFinite(value)) return PRODUCT_SPLIT_DEFAULT_LEFT;
  return Math.min(
    PRODUCT_SPLIT_MAX_LEFT,
    Math.max(PRODUCT_SPLIT_MIN_LEFT, value),
  );
}

type BarcodeApiResponse = {
  success?: boolean;
  message?: string;
  data?: unknown;
  total?: number;
};

const pageSize = 80;

export function ProductBarcodeScreen({
  embedded = false,
  language = "th",
}: ProductBarcodeScreenProps) {
  const lang = normalizeLanguage(language);
  const text = getBarcodeText(lang);
  const { confirm, confirmationDialog } = useConfirmDialog();

  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [items, setItems] = useState<ProductBarcodeRecord[]>([]);
  const [detailItems, setDetailItems] = useState<
    Record<string, ProductBarcodeRecord>
  >({});
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const setNotice = pushNotice;
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(searchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchInput]);
  const [filterOpen, setFilterOpen] = useState(false);
  const [showImage, setShowImage] = useState(false);
  const [selectMode, setSelectMode] = useState(false);
  const [selectedKey, setSelectedKey] = useState("");
  const [checkedBarcodes, setCheckedBarcodes] = useState<string[]>([]);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<"create" | "edit">("create");
  const [editorGuid, setEditorGuid] = useState("");
  const [editorBarcode, setEditorBarcode] = useState<ProductBarcodeObject>(() =>
    emptyProductBarcode(),
  );
  const [editorDirty, setEditorDirty] = useState(false);
  const [editorSaving, setEditorSaving] = useState(false);
  const [splitLeftPercent, setSplitLeftPercent] = useState(
    PRODUCT_SPLIT_DEFAULT_LEFT,
  );
  const [resizingSplit, setResizingSplit] = useState(false);
  const splitContainerRef = useRef<HTMLDivElement | null>(null);
  const loadBarcodesAbortRef = useRef<AbortController | undefined>(undefined);
  const loadDetailAbortRef = useRef<AbortController | undefined>(undefined);
  const [filters, setFilters] = useState({
    groupCode: "",
    brandCode: "",
    categoryCode: "",
    classCode: "",
    designCode: "",
    gradeCode: "",
    modelCode: "",
    patternCode: "",
    priceMin: "",
    priceMax: "",
  });
  const [debouncedFilters, setDebouncedFilters] = useState(filters);
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedFilters(filters);
    }, 300);
    return () => clearTimeout(handler);
  }, [filters]);

  useEffect(() => {
    setAuth(readAuthSession());
    setWorkspace(readWorkspaceSession());
  }, []);

  useEffect(() => {
    const handleWorkspaceChange = () => {
      const nextWorkspace = readWorkspaceSession();
      if (nextWorkspace) {
        setWorkspace(nextWorkspace);
      }
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
  }, []);

  useEffect(() => {
    if (typeof window === "undefined") return;
    const saved = window.localStorage.getItem(PRODUCT_SPLIT_STORAGE_KEY);
    if (!saved) return;
    const next = clampProductSplitLeft(Number(saved));
    setSplitLeftPercent(next);
  }, []);

  const activeHoldingCode = workspace?.shop.holdingcode ?? "";
  const shopLanguages = useMemo(
    () => languageCodesFromWorkspace(workspace),
    [workspace],
  );

  const selectedBase = useMemo(
    () =>
      items.find((item) => barcodeIdentity(item) === selectedKey) ??
      items[0] ??
      null,
    [items, selectedKey],
  );
  const selected = useMemo(() => {
    if (!selectedBase) return null;
    return detailItems[detailKey(selectedBase)] ?? selectedBase;
  }, [detailItems, selectedBase]);
  const selectedIndex = useMemo(
    () => items.findIndex((item) => barcodeIdentity(item) === selectedKey),
    [items, selectedKey],
  );

  const productSplitStyle = useMemo(
    () =>
      ({
        "--product-list-fr": `${splitLeftPercent}fr`,
        "--product-detail-fr": `${100 - splitLeftPercent}fr`,
      }) as CSSProperties,
    [splitLeftPercent],
  );

  const updateSplitFromClientX = useCallback((clientX: number) => {
    const container = splitContainerRef.current;
    if (!container) return;
    const rect = container.getBoundingClientRect();
    if (rect.width <= 0) return;
    const next = ((clientX - rect.left) / rect.width) * 100;
    setSplitLeftPercent(clampProductSplitLeft(next));
  }, []);

  useEffect(() => {
    if (!resizingSplit) return;

    function moveWithMouse(event: MouseEvent) {
      updateSplitFromClientX(event.clientX);
    }

    function stopMouseResize() {
      setResizingSplit(false);
    }

    window.addEventListener("mousemove", moveWithMouse);
    window.addEventListener("mouseup", stopMouseResize);
    return () => {
      window.removeEventListener("mousemove", moveWithMouse);
      window.removeEventListener("mouseup", stopMouseResize);
    };
  }, [resizingSplit, updateSplitFromClientX]);

  const startSplitResize = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      event.preventDefault();
      event.currentTarget.setPointerCapture(event.pointerId);
      setResizingSplit(true);
      updateSplitFromClientX(event.clientX);
    },
    [updateSplitFromClientX],
  );

  const startSplitMouseResize = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      event.preventDefault();
      setResizingSplit(true);
      updateSplitFromClientX(event.clientX);
    },
    [updateSplitFromClientX],
  );

  const moveSplitResize = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (!resizingSplit) return;
      updateSplitFromClientX(event.clientX);
    },
    [resizingSplit, updateSplitFromClientX],
  );

  const stopSplitResize = useCallback((event: PointerEvent<HTMLDivElement>) => {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    setResizingSplit(false);
    setSplitLeftPercent((current) => {
      if (typeof window !== "undefined") {
        window.localStorage.setItem(
          PRODUCT_SPLIT_STORAGE_KEY,
          String(Math.round(current)),
        );
      }
      return current;
    });
  }, []);

  const adjustSplitWithKeyboard = useCallback(
    (event: KeyboardEvent<HTMLDivElement>) => {
      if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
      event.preventDefault();
      const direction = event.key === "ArrowLeft" ? -2 : 2;
      setSplitLeftPercent((current) =>
        clampProductSplitLeft(current + direction),
      );
    },
    [],
  );

  const activeFilterCount = useMemo(
    () => Object.values(filters).filter((value) => value.trim() !== "").length,
    [filters],
  );

  const loadBarcodes = useCallback(async () => {
    if (!auth || !activeHoldingCode) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    loadBarcodesAbortRef.current?.abort();
    const controller = new AbortController();
    loadBarcodesAbortRef.current = controller;

    setLoading(true);
    setNotice(null);
    try {
      const data = await listBarcodes(auth, {
        holdingcode: activeHoldingCode,
        keyword: search.trim(),
        groupcode: debouncedFilters.groupCode.trim(),
        brandcode: debouncedFilters.brandCode.trim(),
        categorycode: debouncedFilters.categoryCode.trim(),
        classcode: debouncedFilters.classCode.trim(),
        designcode: debouncedFilters.designCode.trim(),
        gradecode: debouncedFilters.gradeCode.trim(),
        modelcode: debouncedFilters.modelCode.trim(),
        patterncode: debouncedFilters.patternCode.trim(),
        pricemin: toNumberOrNull(debouncedFilters.priceMin),
        pricemax: toNumberOrNull(debouncedFilters.priceMax),
        limit: pageSize,
        offset: 0,
        sortfield: search.trim() ? "relevance" : "barcode",
        sortorder: search.trim() ? "desc" : "asc",
      });
      if (controller.signal.aborted) return;
      if (!data.success) {
        throw new Error(data.message || text.requestFailed);
      }

      const nextItems = normalizeBarcodeList(data.data);
      setItems(nextItems);
      setTotal(Number(data.total ?? nextItems.length));
      setDetailItems({});
      setSelectedKey((current) => {
        if (
          current &&
          nextItems.some((item) => barcodeIdentity(item) === current)
        )
          return current;
        return nextItems[0] ? barcodeIdentity(nextItems[0]) : "";
      });
      setCheckedBarcodes([]);
    } catch (error) {
      if (error instanceof Error && error.name === "AbortError") return;
      setItems([]);
      setTotal(0);
      setSelectedKey("");
      setNotice({
        type: "error",
        text:
          error instanceof Error && error.message
            ? error.message
            : text.requestFailed,
      });
    } finally {
      if (!controller.signal.aborted) setLoading(false);
    }
  }, [
    activeHoldingCode,
    auth,
    debouncedFilters,
    search,
    text.apiRequired,
    text.requestFailed,
  ]);

  const loadBarcodeDetail = useCallback(
    async (item: ProductBarcodeRecord) => {
      if (!auth || !item.guidFixed) return;
      const key = detailKey(item);
      if (detailItems[key]) return;

      loadDetailAbortRef.current?.abort();
      const controller = new AbortController();
      loadDetailAbortRef.current = controller;

      try {
        const response = await fetch(
          `/api/product-barcode/${encodeURIComponent(item.guidFixed)}`,
          {
            method: "GET",
            headers: {
              "x-bc-backend-url": auth.backendUrl,
              Authorization: `Bearer ${auth.token}`,
            },
            signal: controller.signal,
          },
        );
        const data = (await response.json()) as BarcodeApiResponse;
        if (!response.ok || data.success === false) return;

        const detail = extractBarcodeRecordPayload(data.data);
        if (!detail) return;
        const mergedDetail = isRecord(detail)
          ? { ...item.raw, ...detail }
          : detail;
        const normalized = normalizeBarcodeRecord(mergedDetail);
        setDetailItems((current) => ({ ...current, [key]: normalized }));
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") return;
        // Detail loading is best-effort; the list row still has enough data to remain usable.
      }
    },
    [auth, detailItems],
  );

  useEffect(() => {
    if (!auth || !activeHoldingCode) return;
    const timer = window.setTimeout(() => {
      void loadBarcodes();
    }, 300);
    return () => window.clearTimeout(timer);
  }, [activeHoldingCode, auth, loadBarcodes]);

  useEffect(() => {
    if (!selectedBase) return;
    void loadBarcodeDetail(selectedBase);
  }, [loadBarcodeDetail, selectedBase]);

  function toggleChecked(barcode: string) {
    setCheckedBarcodes((current) =>
      current.includes(barcode)
        ? current.filter((item) => item !== barcode)
        : [...current, barcode],
    );
  }

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void loadBarcodes();
  }

  async function canDiscardEditor() {
    if (!editorOpen || !editorDirty) return true;
    return confirm({
      title: text.unsavedConfirm,
      description: editorBarcode.barcode
        ? `${text.barcode}: ${editorBarcode.barcode}`
        : undefined,
      details: text.unsavedDetail,
      confirmLabel: text.discardLabel,
      cancelLabel: text.close,
      tone: "warning",
    });
  }

  async function closeEditor() {
    if (!(await canDiscardEditor())) return;
    setEditorOpen(false);
    setEditorGuid("");
    setEditorDirty(false);
  }

  function handleEditorChange(
    next:
      | ProductBarcodeObject
      | ((current: ProductBarcodeObject) => ProductBarcodeObject),
  ) {
    setEditorDirty(true);
    setEditorBarcode(next);
  }

  async function selectListItem(item: ProductBarcodeRecord) {
    if (!(await canDiscardEditor())) return;
    setEditorOpen(false);
    setEditorDirty(false);
    setSelectedKey(barcodeIdentity(item));
  }

  function selectRowByIndex(nextIndex: number) {
    const next = items[nextIndex];
    if (!next) return;
    void selectListItem(next);
  }

  function handleListKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (!items.length || selectMode) return;
    if (event.key !== "ArrowUp" && event.key !== "ArrowDown") return;
    event.preventDefault();
    const fallbackIndex = selectedIndex >= 0 ? selectedIndex : 0;
    const nextIndex =
      event.key === "ArrowUp"
        ? Math.max(0, fallbackIndex - 1)
        : Math.min(items.length - 1, fallbackIndex + 1);
    selectRowByIndex(nextIndex);
  }

  function clearFilters() {
    setFilters({
      groupCode: "",
      brandCode: "",
      categoryCode: "",
      classCode: "",
      designCode: "",
      gradeCode: "",
      modelCode: "",
      patternCode: "",
      priceMin: "",
      priceMax: "",
    });
  }

  async function openCreateEditor() {
    if (!(await canDiscardEditor())) return;
    setEditorMode("create");
    setEditorGuid("");
    setEditorDirty(false);
    const next = emptyProductBarcode();
    next.holdingcode = activeHoldingCode;
    setEditorBarcode(next);
    setEditorOpen(true);
  }

  async function openEditEditor(itemOverride?: ProductBarcodeRecord) {
    if (!(await canDiscardEditor())) return;
    const target = itemOverride ?? selected;
    if (!target) return;
    setEditorMode("edit");
    setSelectedKey(barcodeIdentity(target));
    setEditorGuid(target.guidFixed);
    setEditorDirty(false);
    const base = emptyProductBarcode();
    base.holdingcode = activeHoldingCode;
    setEditorBarcode(rawToProductBarcode(target.raw, base));
    setEditorOpen(true);
  }

  async function openCopyEditor() {
    if (!(await canDiscardEditor())) return;
    if (!selected) return;
    setEditorMode("create");
    setEditorGuid("");
    setEditorDirty(true);
    const base = emptyProductBarcode();
    base.holdingcode = activeHoldingCode;
    const next = rawToProductBarcode(selected.raw, base);
    next.guidfixed = "";
    next.holdingcode = activeHoldingCode;
    setEditorBarcode(next);
    setEditorOpen(true);
  }

  function copyCurrentEditorValue() {
    const next = JSON.parse(
      JSON.stringify(editorBarcode),
    ) as ProductBarcodeObject;
    next.guidfixed = "";
    next.holdingcode = activeHoldingCode;
    setEditorMode("create");
    setEditorGuid("");
    setEditorBarcode(next);
    setEditorDirty(true);
  }

  async function saveEditor(value: ProductBarcodeObject, keepOpen = false) {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    const guid = editorGuid || String(value.guidfixed ?? "");
    if (editorMode === "edit" && !guid) {
      setNotice({ type: "error", text: text.missingGuid });
      return;
    }

    setEditorSaving(true);
    setNotice(null);
    try {
      const data =
        editorMode === "edit"
          ? await updateBarcode(auth, guid, value)
          : await createBarcode(auth, value);
      if (!data.success) {
        throw new Error(data.message || text.requestFailed);
      }

      if (keepOpen) {
        // "Save & add new": reset to a blank record and stay open
        const next = emptyProductBarcode();
        next.holdingcode = activeHoldingCode;
        setEditorGuid("");
        setEditorDirty(false);
        setEditorBarcode(next);
      } else {
        setEditorOpen(false);
        setEditorGuid("");
        setEditorDirty(false);
      }
      await loadBarcodes();
      setNotice({ type: "success", text: text.saveSuccess });
    } catch (error) {
      setNotice({
        type: "error",
        text:
          error instanceof Error && error.message
            ? error.message
            : text.requestFailed,
      });
    } finally {
      setEditorSaving(false);
    }
  }

  async function deleteSelected() {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    const selectedItems = items.filter((item) =>
      checkedBarcodes.includes(barcodeRowKey(item)),
    );
    const guids = selectedItems
      .map((item) => item.guidFixed.trim())
      .filter(Boolean);
    if (guids.length === 0) {
      setNotice({ type: "error", text: text.missingGuid });
      return;
    }

    const confirmed = await confirm({
      title: text.deleteConfirm,
      description: `${text.selected}: ${guids.length.toLocaleString(lang === "th" ? "th-TH" : "en-US")}`,
      details: selectedItems
        .slice(0, 8)
        .map((item) => item.barcode)
        .filter(Boolean)
        .join(", "),
      confirmLabel: text.delete,
      cancelLabel: text.close,
      tone: "danger",
    });
    if (!confirmed) return;

    setLoading(true);
    setNotice(null);
    try {
      const data = await deleteBarcodes(auth, guids);
      if (!data.success) {
        throw new Error(data.message || text.requestFailed);
      }

      setSelectMode(false);
      setCheckedBarcodes([]);
      await loadBarcodes();
      setNotice({ type: "success", text: text.deleteSuccess });
    } catch (error) {
      setNotice({
        type: "error",
        text:
          error instanceof Error && error.message
            ? error.message
            : text.requestFailed,
      });
    } finally {
      setLoading(false);
    }
  }

  async function deleteCurrentItem(itemOverride?: ProductBarcodeRecord) {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    const target = itemOverride ?? selected;
    const guid = (
      target?.guidFixed ||
      editorGuid ||
      String(editorBarcode.guidfixed ?? "")
    ).trim();
    if (!guid) {
      setNotice({ type: "error", text: text.missingGuid });
      return;
    }

    const barcode = target?.barcode ?? editorBarcode.barcode ?? "";
    const confirmed = await confirm({
      title: text.deleteCurrentConfirm,
      description: barcode ? `${text.barcode}: ${barcode}` : undefined,
      details: guid ? `GUID: ${guid}` : undefined,
      confirmLabel: text.delete,
      cancelLabel: text.close,
      tone: "danger",
    });
    if (!confirmed) return;

    setLoading(true);
    setNotice(null);
    try {
      const data = await deleteBarcodes(auth, [guid]);
      if (!data.success) {
        throw new Error(data.message || text.requestFailed);
      }

      setEditorOpen(false);
      setEditorGuid("");
      setEditorMode("create");
      setEditorDirty(false);
      setCheckedBarcodes([]);
      setSelectMode(false);
      await loadBarcodes();
      setNotice({ type: "success", text: text.deleteSuccess });
    } catch (error) {
      setNotice({
        type: "error",
        text:
          error instanceof Error && error.message
            ? error.message
            : text.requestFailed,
      });
    } finally {
      setLoading(false);
    }
  }

  const content = (
    <div
      className={cn(
        embedded
          ? "flex h-full min-h-0 w-full min-w-0 flex-col gap-3 overflow-hidden"
          : "grid w-full min-w-0 gap-3",
      )}
    >
      {!embedded ? (
        <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="flex min-w-0 items-start gap-3">
              <span className="grid size-10 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
                <Barcode size={22} />
              </span>
              <div className="min-w-0">
                <p className="text-xs font-semibold uppercase text-muted-foreground">
                  BC Ai Account
                </p>
                <h1 className="truncate text-3xl font-semibold leading-tight">
                  {text.title}
                </h1>
                <p className="text-sm text-muted-foreground">{text.subtitle}</p>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => void loadBarcodes()}
                disabled={loading}
              >
                <RefreshCcw size={16} />
                {text.refresh}
              </Button>
              <Button
                variant={selectMode ? "secondary" : "outline"}
                size="sm"
                onClick={() => {
                  setSelectMode((current) => !current);
                  setCheckedBarcodes([]);
                }}
              >
                {selectMode ? <X size={16} /> : <CheckSquare size={16} />}
                {selectMode ? text.cancelSelect : text.selectDelete}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => void deleteSelected()}
                disabled={!selectMode || checkedBarcodes.length === 0}
              >
                <Trash2 size={16} />
                {checkedBarcodes.length || ""}
              </Button>
              <Button
                size="sm"
                onClick={() => void openCopyEditor()}
                disabled={!selected}
              >
                <Copy size={16} />
                คัดลอก
              </Button>
              <Button size="sm" onClick={() => void openCreateEditor()}>
                <Plus size={16} />
                {text.add}
              </Button>
            </div>
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            <Badge variant="outline">
              {text.total}: {total.toLocaleString("th-TH")}
            </Badge>
            {selectMode ? (
              <Badge variant="warning">
                {text.selected}:{" "}
                {checkedBarcodes.length.toLocaleString("th-TH")}
              </Badge>
            ) : null}
          </div>
        </header>
      ) : null}

      {embedded ? (
        <div className="flex shrink-0 items-center justify-between border-b border-border bg-card px-4 py-3">
          <div>
            <h2 className="text-lg font-bold">{text.title}</h2>
            <p className="text-xs text-muted-foreground">{text.subtitle}</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              type="button"
              onClick={() => void loadBarcodes()}
              disabled={loading}
            >
              {loading ? (
                <Loader2 size={16} className="animate-spin" />
              ) : (
                <RefreshCcw size={16} />
              )}
              {text.refresh}
            </Button>
            <Button
              size="sm"
              type="button"
              onClick={() => void openCopyEditor()}
              disabled={!selected}
            >
              <Copy size={16} />
              คัดลอก
            </Button>
            <Button
              size="sm"
              type="button"
              onClick={() => void openCreateEditor()}
            >
              <Plus size={16} />
              {text.add}
            </Button>
          </div>
        </div>
      ) : null}

      <div
        ref={splitContainerRef}
        className={cn(
          "grid min-h-0 min-w-0 gap-3 xl:grid-cols-[minmax(0,var(--product-list-fr))_8px_minmax(0,var(--product-detail-fr))] xl:gap-0",
          embedded && "flex-1 xl:h-full",
        )}
        style={productSplitStyle}
      >
        <Card
          aria-label={text.title}
          className="min-w-0 overflow-hidden focus-visible:ring-2 focus-visible:ring-primary/40 focus-visible:outline-none xl:flex xl:h-full xl:min-h-0 xl:flex-col"
          onKeyDown={handleListKeyDown}
          role="list"
          tabIndex={0}
        >
          <CardHeader className="shrink-0 border-b border-border p-3">
            <form className="flex flex-wrap gap-2" onSubmit={submitSearch}>
              <div className="relative min-w-[240px] flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  className="h-9 !pl-10"
                  value={searchInput}
                  onChange={(event) => setSearchInput(event.target.value)}
                  placeholder={text.search}
                />
              </div>
              <Button
                variant={filterOpen ? "secondary" : "outline"}
                size="sm"
                type="button"
                onClick={() => setFilterOpen((current) => !current)}
              >
                <Filter size={16} />
                {text.filter}
                {activeFilterCount > 0 ? (
                  <Badge variant="warning">{activeFilterCount}</Badge>
                ) : null}
              </Button>
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setShowImage((current) => !current)}
              >
                {showImage ? <ImageOff size={16} /> : <ImageIcon size={16} />}
                {text.image}
              </Button>
              {embedded ? (
                <>
                  <Button
                    variant={selectMode ? "secondary" : "outline"}
                    size="sm"
                    type="button"
                    onClick={() => {
                      setSelectMode((current) => !current);
                      setCheckedBarcodes([]);
                    }}
                  >
                    {selectMode ? <X size={16} /> : <CheckSquare size={16} />}
                    {selectMode ? text.cancelSelect : text.selectDelete}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    type="button"
                    onClick={() => void deleteSelected()}
                    disabled={!selectMode || checkedBarcodes.length === 0}
                  >
                    <Trash2 size={16} />
                    {checkedBarcodes.length || ""}
                  </Button>
                </>
              ) : null}
            </form>
            {embedded ? (
              <div className="mt-2 flex flex-wrap gap-2">
                <Badge variant="outline">
                  {text.total}: {total.toLocaleString("th-TH")}
                </Badge>
                {selectMode ? (
                  <Badge variant="warning">
                    {text.selected}:{" "}
                    {checkedBarcodes.length.toLocaleString("th-TH")}
                  </Badge>
                ) : null}
              </div>
            ) : null}
            {filterOpen ? (
              <div className="mt-3 grid gap-2 rounded-2xl border border-border bg-muted/30 p-3 md:grid-cols-4 xl:grid-cols-5">
                <Input
                  value={filters.groupCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      groupCode: event.target.value,
                    }))
                  }
                  placeholder={text.groupCode}
                />
                <Input
                  value={filters.brandCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      brandCode: event.target.value,
                    }))
                  }
                  placeholder={text.brandCode}
                />
                <Input
                  value={filters.categoryCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      categoryCode: event.target.value,
                    }))
                  }
                  placeholder={text.categoryCode}
                />
                <Input
                  value={filters.classCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      classCode: event.target.value,
                    }))
                  }
                  placeholder={text.classCode}
                />
                <Input
                  value={filters.designCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      designCode: event.target.value,
                    }))
                  }
                  placeholder={text.designCode}
                />
                <Input
                  value={filters.gradeCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      gradeCode: event.target.value,
                    }))
                  }
                  placeholder={text.gradeCode}
                />
                <Input
                  value={filters.modelCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      modelCode: event.target.value,
                    }))
                  }
                  placeholder={text.modelCode}
                />
                <Input
                  value={filters.patternCode}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      patternCode: event.target.value,
                    }))
                  }
                  placeholder={text.patternCode}
                />
                <Input
                  value={filters.priceMin}
                  onChange={(event) =>
                    setFilters((current) => ({
                      ...current,
                      priceMin: event.target.value,
                    }))
                  }
                  placeholder={text.priceMin}
                  inputMode="decimal"
                />
                <div className="flex gap-2">
                  <Input
                    value={filters.priceMax}
                    onChange={(event) =>
                      setFilters((current) => ({
                        ...current,
                        priceMax: event.target.value,
                      }))
                    }
                    placeholder={text.priceMax}
                    inputMode="decimal"
                  />
                  <Button
                    variant="outline"
                    type="button"
                    onClick={clearFilters}
                    aria-label={text.clearFilter}
                  >
                    <X size={16} />
                  </Button>
                </div>
              </div>
            ) : null}
          </CardHeader>
          <CardContent className="grid min-h-[420px] p-0 xl:min-h-0 xl:flex-1 xl:grid-rows-[auto_minmax(0,1fr)]">
            <div className="bc-list-header hidden lg:grid grid-cols-[1.35fr_2.2fr_0.8fr_1.1fr_0.9fr] gap-x-2">
              <span>{text.barcode}</span>
              <span>{text.productName}</span>
              <span>{text.unit}</span>
              <span>{text.itemCode}</span>
              <span className="text-right">{text.retailPrice}</span>
            </div>
            <div className="relative min-h-[360px] xl:min-h-0">
              {loading ? (
                <div className="absolute inset-0 z-10 grid place-items-center bg-background/70">
                  <div className="flex items-center gap-2 rounded-2xl border border-border bg-card px-4 py-3 text-sm font-medium shadow-sm">
                    <Loader2 className="size-4 animate-spin" />
                    {text.loading}
                  </div>
                </div>
              ) : null}
              {items.length === 0 && !loading ? (
                <div className="grid min-h-[360px] place-items-center p-6 text-center text-sm text-muted-foreground xl:h-full xl:min-h-0">
                  <div>
                    <Package className="mx-auto mb-2 size-10 text-muted-foreground" />
                    {text.noData}
                  </div>
                </div>
              ) : (
                <div className="min-h-[360px] overflow-auto xl:h-full xl:min-h-0">
                  {items.map((item, index) => {
                    const rowKey = barcodeRowKey(item);
                    return (
                      <BarcodeRow
                        auth={auth}
                        checked={checkedBarcodes.includes(rowKey)}
                        imageUrl={item.imageUri}
                        index={index}
                        item={item}
                        key={rowKey}
                        onSelect={() => void selectListItem(item)}
                        onToggleChecked={() => toggleChecked(rowKey)}
                        selected={
                          selected
                            ? barcodeIdentity(selected) ===
                              barcodeIdentity(item)
                            : false
                        }
                        editing={
                          selected
                            ? barcodeIdentity(selected) ===
                                barcodeIdentity(item) &&
                              editorOpen &&
                              editorMode === "edit"
                            : false
                        }
                        selectMode={selectMode}
                        showImage={showImage}
                        text={text}
                      />
                    );
                  })}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <div
          aria-label={text.resizeAriaLabel}
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

        {editorOpen ? (
          <ProductBarcodeFormDialog
            open={editorOpen}
            mode={editorMode}
            value={editorBarcode}
            onChange={handleEditorChange}
            onSave={(val) => void saveEditor(val)}
            onSaveAndNew={(val) => void saveEditor(val, true)}
            onCancel={() => void closeEditor()}
            saving={editorSaving}
            language={lang}
            shopLanguages={shopLanguages}
            auth={auth}
            companyGuid={workspace?.branch?.companyguid}
            extraActions={
              editorMode === "edit" ? (
                <>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={copyCurrentEditorValue}
                    aria-label={text.copy}
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => void deleteCurrentItem()}
                    aria-label={text.deleteConfirm}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </>
              ) : null
            }
            embedded
          />
        ) : (
          <ProductBarcodeDetail
            item={selected}
            onDelete={() => void deleteCurrentItem()}
            onEdit={() => void openEditEditor()}
            text={text}
          />
        )}
      </div>
      {selected && !editorOpen ? (
        <BarcodeQuickActions
          barcode={selected.barcode}
          guid={selected.guidFixed}
          itemCode={selected.itemCode}
          text={text}
        />
      ) : null}
      {confirmationDialog}
    </div>
  );

  if (embedded) return content;
  return <main className="min-h-screen bg-background p-3">{content}</main>;
}

function BarcodeRow({
  auth,
  checked,
  imageUrl,
  index,
  item,
  onSelect,
  onToggleChecked,
  selected,
  editing,
  selectMode,
  showImage,
  text,
}: {
  auth: AuthSession | null;
  checked: boolean;
  imageUrl: string;
  index: number;
  item: ProductBarcodeRecord;
  onSelect: () => void;
  onToggleChecked: () => void;
  selected: boolean;
  editing: boolean;
  selectMode: boolean;
  showImage: boolean;
  text: BarcodeText;
}) {
  function handleRowKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key !== "Enter" && event.key !== " ") return;
    event.preventDefault();
    if (selectMode) onToggleChecked();
    else onSelect();
  }

  return (
    <div
      aria-label={`${text.barcode}: ${item.barcode || item.itemCode || "-"}`}
      aria-pressed={selected}
      className={cn(
        "bc-list-row grid focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 lg:grid-cols-[1.35fr_2.2fr_0.8fr_1.1fr_0.9fr] gap-x-2",
        editing
          ? "bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-amber-200/50"
          : selected
            ? "bg-primary/10 hover:bg-primary/15"
            : index % 2 === 0
              ? "bg-background hover:bg-primary/5"
              : "bg-muted/20 hover:bg-primary/5",
      )}
      onClick={selectMode ? onToggleChecked : onSelect}
      onKeyDown={handleRowKeyDown}
      role="button"
      tabIndex={0}
    >
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.barcode}
        </span>
        <div className="flex min-w-0 items-center gap-2">
          {selectMode ? (
            <span
              className={cn(
                "grid size-6 shrink-0 place-items-center rounded-md border",
                checked && "border-primary bg-primary text-primary-foreground",
              )}
            >
              {checked ? <CheckSquare size={14} /> : null}
            </span>
          ) : null}
          {showImage ? (
            imageUrl ? (
              <AuthenticatedImg
                alt=""
                className="size-8 shrink-0 rounded-lg border border-border object-cover"
                src={imageUrl}
                auth={auth}
                fallback={
                  <span className="grid size-8 shrink-0 place-items-center rounded-lg border border-border text-muted-foreground">
                    <ImageOff size={14} />
                  </span>
                }
              />
            ) : (
              <span className="grid size-8 shrink-0 place-items-center rounded-lg border border-border text-muted-foreground">
                <ImageOff size={14} />
              </span>
            )
          ) : null}
          <span className="truncate">{item.barcode || "-"}</span>
        </div>
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.productName}
        </span>
        <div className="line-clamp-2">{item.name || "-"}</div>
        {item.groupName ? (
          <div className="truncate text-xs text-muted-foreground">
            {item.groupName}
          </div>
        ) : null}
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.unit}
        </span>
        <div className={cn("truncate", item.unitCount > 1 && "text-primary")}>
          {item.unitName || item.unitCode || "-"}
        </div>
        {item.unitCount > 1 ? (
          <div className="truncate text-xs text-muted-foreground">
            {item.allUnitNames}
          </div>
        ) : null}
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.itemCode}
        </span>
        <div className="truncate">{item.itemCode || "-"}</div>
      </div>
      <div className="lg:text-right">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.retailPrice}
        </span>
        <div>{formatMoney(item.price)}</div>
      </div>
    </div>
  );
}

function barcodeIdentity(item: ProductBarcodeRecord): string {
  return item.guidFixed || `${item.itemCode}\u0000${item.barcode}`;
}

function barcodeRowKey(item: ProductBarcodeRecord): string {
  return barcodeIdentity(item);
}

function ProductBarcodeDetail({
  item,
  onDelete,
  onEdit,
  text,
}: {
  item: ProductBarcodeRecord | null;
  onDelete: () => void;
  onEdit: () => void;
  text: BarcodeText;
}) {
  const basicFields = item
    ? [
        { label: text.guid, value: item.guidFixed },
        { label: text.holdingCode, value: item.holdingCode },
        { label: text.barcode, value: item.barcode },
        { label: text.barcodeRef, value: item.barcodeRef },
        { label: text.productName, value: item.name },
        { label: text.itemCode, value: item.itemCode },
        { label: text.itemGuid, value: item.itemGuid },
        { label: text.itemGuidFixed, value: item.itemGuidFixed },
        { label: text.parentGuid, value: item.parentGuid },
        { label: text.checksum, value: item.checksum },
      ]
    : [];
  const classificationFields = item
    ? [
        {
          label: text.groupCode,
          value: formatCodeName(item.groupCode, item.groupName),
        },
        {
          label: text.brandCode,
          value: formatCodeName(item.brandCode, item.brandName),
        },
        {
          label: text.categoryCode,
          value: formatCodeName(item.categoryCode, item.categoryName),
        },
        {
          label: text.classCode,
          value: formatCodeName(item.classCode, item.className),
        },
        {
          label: text.designCode,
          value: formatCodeName(item.designCode, item.designName),
        },
        {
          label: text.gradeCode,
          value: formatCodeName(item.gradeCode, item.gradeName),
        },
        {
          label: text.modelCode,
          value: formatCodeName(item.modelCode, item.modelName),
        },
        {
          label: text.patternCode,
          value: formatCodeName(item.patternCode, item.patternName),
        },
        {
          label: text.subGroup1,
          value: formatCodeName(item.groupSubOneCode, item.groupSubOneName),
        },
        {
          label: text.subGroup2,
          value: formatCodeName(item.groupSubTwoCode, item.groupSubTwoName),
        },
        {
          label: text.manufacturer,
          value: formatCodeName(item.manufacturerCode, item.manufacturerName),
        },
        {
          label: text.shelf,
          value: formatCodeName(item.shelfCode, item.shelfName),
        },
      ]
    : [];
  const stockFields = item
    ? [
        {
          label: text.unit,
          value: formatCodeName(item.unitCode, item.unitName),
        },
        { label: text.multiUnit, value: item.allUnitNames },
        {
          label: text.balance,
          value: item.balanceFormatted || formatNumber(item.balanceQty),
        },
        { label: text.retailPrice, value: formatMoney(item.price) },
        { label: text.standValue, value: formatNumber(item.standValue) },
        { label: text.divideValue, value: formatNumber(item.divideValue) },
        { label: text.unitCount, value: String(item.unitCount || "") },
        { label: text.color, value: item.colorSelectHex || item.colorSelect },
      ]
    : [];
  const statusFields = item
    ? [
        {
          label: text.itemType,
          value:
            item.itemType === 0
              ? text.itemTypeStock
              : item.itemType === 1
                ? text.itemTypeService
                : item.itemType === 2
                  ? text.itemTypeSet
                  : item.itemType === 3
                    ? text.itemTypeNotStock
                    : String(item.itemType || ""),
        },
        { label: text.productType, value: String(item.productType || "") },
        {
          label: text.foodType,
          value:
            item.foodType === 0
              ? text.foodTypeFood
              : item.foodType === 1
                ? text.foodTypeDrink
                : item.foodType === 2
                  ? text.foodTypeAlcohol
                  : item.foodType === 3
                    ? text.foodTypeOther
                    : String(item.foodType || ""),
        },
        {
          label: text.materialType,
          value:
            item.materialType === 0
              ? text.materialGeneral
              : item.materialType === 1
                ? text.materialMaterial
                : item.materialType === 2
                  ? text.materialSemiFinished
                  : item.materialType === 3
                    ? text.materialSet
                    : item.materialType === 4
                      ? text.materialAgricultural
                      : String(item.materialType || ""),
        },
        { label: text.isStock, value: formatBoolean(item.isStock === 1, text) },
        {
          label: text.isMainBarcode,
          value: formatBoolean(item.isMainBarcode, text),
        },
        { label: text.isMainItem, value: formatBoolean(item.isMainItem, text) },
        {
          label: text.isUseSubBarcodes,
          value: formatBoolean(item.isUseSubBarcodes, text),
        },
        {
          label: text.useImageOrColor,
          value: formatBoolean(item.useImageOrColor, text),
        },
        { label: text.condition, value: formatBoolean(item.condition, text) },
        { label: text.isSumPoint, value: formatBoolean(item.isSumPoint, text) },
        { label: text.isDividend, value: formatBoolean(item.isDividend, text) },
        { label: text.isALaCarte, value: formatBoolean(item.isALaCarte, text) },
        {
          label: text.isSplitUnitPrint,
          value: formatBoolean(item.isSplitUnitPrint, text),
        },
        {
          label: text.isOnlyStaff,
          value: formatBoolean(item.isOnlyStaff, text),
        },
        {
          label: text.isStockForRestaurant,
          value: formatBoolean(item.isStockForRestaurant, text),
        },
        {
          label: text.isDiscountPointOfPurchase,
          value: formatBoolean(item.isDiscountPointOfPurchase, text),
        },
        { label: text.isAlert, value: formatBoolean(item.isAlert, text) },
        { label: text.isDisable, value: formatBoolean(item.isDisable, text) },
        { label: text.showIsDividend, value: item.showIsDividend },
        {
          label: text.rowNumber,
          value: item.rowNumber ? String(item.rowNumber) : "",
        },
      ]
    : [];
  const relationFields = item
    ? [
        { label: text.prices, value: formatCount(item.priceCount) },
        { label: text.refBarcodes, value: formatCount(item.refBarcodeCount) },
        { label: text.subBarcodes, value: formatCount(item.subBarcodeCount) },
        { label: text.bom, value: formatCount(item.bomCount) },
        { label: text.options, value: formatCount(item.optionCount) },
        { label: text.orderTypes, value: formatCount(item.orderTypeCount) },
        { label: text.dimensions, value: formatCount(item.dimensionCount) },
        {
          label: text.businessTypes,
          value: formatCount(item.businessTypeCount),
        },
        {
          label: text.ignoreBranches,
          value: formatCount(item.ignoreBranchCount),
        },
        { label: text.branches, value: formatCount(item.branchCount) },
        { label: text.categories, value: formatCount(item.categoryCount) },
        { label: text.timeForSales, value: formatCount(item.timeForSaleCount) },
        { label: text.fixedCost, value: formatCount(item.fixedCostCount) },
        { label: text.maxDiscount, value: item.maxDiscount },
        { label: text.discount, value: item.discount },
        { label: text.description, value: item.description },
        { label: text.alertDescription, value: item.alertDescription },
      ]
    : [];

  return (
    <Card className="min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col">
      <CardHeader className="shrink-0 border-b border-border p-3">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Package size={18} />
            {text.detailTitle}
          </CardTitle>
          <div className="flex flex-wrap justify-end gap-2">
            <Button
              disabled={!item}
              onClick={onDelete}
              size="sm"
              variant="outline"
            >
              <Trash2 size={16} />
              {text.delete}
            </Button>
            <Button
              disabled={!item}
              onClick={onEdit}
              size="sm"
              variant="outline"
            >
              <Pencil size={16} />
              {text.edit}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 overflow-auto p-3 xl:min-h-0 xl:flex-1">
        {!item ? (
          <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">
            {text.noSelection}
          </div>
        ) : (
          <>
            <div className="rounded-2xl border border-border bg-muted/30 p-3">
              <p className="text-xs font-semibold text-muted-foreground">
                {text.barcode}
              </p>
              <p className="break-all text-2xl font-semibold">{item.barcode}</p>
              <p className="mt-1 text-sm text-muted-foreground">{item.name}</p>
            </div>
            <DetailSection fields={basicFields} title={text.basicInfo} />
            <DetailSection
              fields={classificationFields}
              title={text.classificationSection}
            />
            <DetailSection fields={stockFields} title={text.stockAndUnit} />
            <DetailSection fields={statusFields} title={text.taxAndFlags} />
            <DetailSection fields={relationFields} title={text.relations} />
            <details className="rounded-2xl border border-border p-3">
              <summary className="cursor-pointer text-sm font-semibold">
                {text.rawFields}
              </summary>
              <pre className="mt-3 max-h-80 overflow-auto whitespace-pre-wrap break-words rounded-xl bg-muted p-3 text-xs">
                {JSON.stringify(item.raw, null, 2)}
              </pre>
            </details>
          </>
        )}
      </CardContent>
    </Card>
  );
}

/**
 * Floating quick-action chip — shown next to the selected row.
 * Provides direct links to:
 * - Label print (existing product-barcode-shelf-screen route)
 * - BOM view (TODO: open BOM read-only dialog)
 * - Price history (existing product-price-history-screen route)
 */
function BarcodeQuickActions({
  barcode,
  guid,
  itemCode,
  text,
}: {
  barcode: string;
  guid: string;
  itemCode: string;
  text: BarcodeText;
}) {
  if (!barcode) return null;
  const encoded = encodeURIComponent(barcode);
  const query = new URLSearchParams({ barcode });
  if (itemCode) query.set("itemcode", itemCode);
  const queryString = query.toString();
  return (
    <div className="fixed bottom-3 right-3 z-30 flex gap-2 rounded-full border border-border bg-card/95 p-1 shadow-lg backdrop-blur md:bottom-6 md:right-6">
      <Button asChild size="sm" variant="ghost" title={text.labelPrint}>
        <a
          href={`/product_barcode_shelf?${queryString}`}
          target="_blank"
          rel="noreferrer"
        >
          <Printer size={16} />
          <span className="hidden md:inline">{text.labelPrint}</span>
        </a>
      </Button>
      <Button
        asChild
        size="sm"
        variant="ghost"
        title={text.priceHistory ?? "Price history"}
      >
        <a
          href={`/price_history?${queryString}`}
          target="_blank"
          rel="noreferrer"
        >
          <History size={16} />
          <span className="hidden md:inline">
            {text.priceHistory ?? "Price history"}
          </span>
        </a>
      </Button>
      {guid ? (
        <Button asChild size="sm" variant="ghost" title={text.bomView ?? "BOM"}>
          <a
            href={`/api/product-barcode/bom/${encoded}${itemCode ? `?itemcode=${encodeURIComponent(itemCode)}` : ""}`}
            target="_blank"
            rel="noreferrer"
          >
            <GitFork size={16} />
            <span className="hidden md:inline">{text.bomView ?? "BOM"}</span>
          </a>
        </Button>
      ) : null}
    </div>
  );
}

type DetailFieldItem = {
  label: string;
  value: string;
};

function DetailSection({
  fields,
  title,
}: {
  fields: DetailFieldItem[];
  title: string;
}) {
  return (
    <section className="rounded-2xl border border-border p-3">
      <h3 className="mb-2 text-sm font-semibold">{title}</h3>
      <div className="grid gap-2 sm:grid-cols-2">
        {fields.map((field) => (
          <DetailField
            key={`${title}-${field.label}`}
            label={field.label}
            value={field.value}
          />
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

function readAuthSession(): AuthSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.auth);
    if (!raw) return null;
    const auth = JSON.parse(raw) as AuthSession;
    return auth.token && auth.backendUrl ? auth : null;
  } catch {
    return null;
  }
}

function readWorkspaceSession(): WorkspaceSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!raw) return null;
    const workspace = JSON.parse(raw) as WorkspaceSession;
    return workspace.shop?.holdingcode ? workspace : null;
  } catch {
    return null;
  }
}

// Language logic is now imported from names-editor.tsx

function normalizeBarcodeList(value: unknown): ProductBarcodeRecord[] {
  if (!Array.isArray(value)) return [];
  return value
    .map(normalizeBarcodeRecord)
    .filter((item) => item.barcode || item.itemCode);
}

function normalizeBarcodeRecord(value: unknown): ProductBarcodeRecord {
  const record = isRecord(value) ? value : {};
  return {
    raw: record,
    guidFixed: getFirstString(record, ["guidfixed", "guidfixed"]),
    holdingCode: getFirstString(record, ["holdingcode", "holdingcode"]),
    barcode: getFirstString(record, ["barcode"]),
    barcodeRef: getFirstString(record, [
      "barcoderef",
      "barcode_ref",
      "refbarcode",
    ]),
    name: localizedNameFromKeys(
      record,
      ["names"],
      getFirstString(record, ["name0", "name", "itemname"]),
    ),
    unitName: localizedNameFromKeys(
      record,
      ["itemunitnames", "unitnames"],
      getFirstString(record, ["unitname", "unitname"]),
    ),
    unitCode: getFirstString(record, [
      "itemunitcode",
      "itemunitcode",
      "unitcode",
    ]),
    itemCode: getFirstString(record, ["itemcode", "itemcode"]),
    itemGuid: getFirstString(record, ["itemguid", "itemguid"]),
    itemGuidFixed: getFirstString(record, ["itemguidfixed", "itemguidfixed"]),
    parentGuid: getFirstString(record, ["parentguid"]),
    groupName: localizedNameFromKeys(
      record,
      ["groupnames", "groupnames"],
      getFirstString(record, ["groupname", "groupname"]),
    ),
    groupCode: getFirstString(record, ["groupcode", "groupcode"]),
    brandName: localizedNameFromKeys(
      record,
      ["brand_names", "brandnames"],
      getFirstString(record, ["brand_name", "brandname"]),
    ),
    brandCode: getFirstString(record, ["brandcode", "brandcode"]),
    categoryName: localizedNameFromKeys(
      record,
      ["categorynames", "categorynames"],
      getFirstString(record, ["categoryname", "categoryname"]),
    ),
    categoryCode: getFirstString(record, ["category_code", "categorycode"]),
    className: localizedNameFromKeys(
      record,
      ["class_names", "classnames"],
      getFirstString(record, ["class_name", "classname"]),
    ),
    classCode: getFirstString(record, ["class_code", "classcode"]),
    designName: localizedNameFromKeys(
      record,
      ["design_names", "designnames"],
      getFirstString(record, ["design_name", "designname"]),
    ),
    designCode: getFirstString(record, ["design_code", "designcode"]),
    gradeName: localizedNameFromKeys(
      record,
      ["grade_names", "gradenames"],
      getFirstString(record, ["grade_name", "gradename"]),
    ),
    gradeCode: getFirstString(record, ["grade_code", "gradecode"]),
    modelName: localizedNameFromKeys(
      record,
      ["model_names", "modelnames"],
      getFirstString(record, ["model_name", "modelname"]),
    ),
    modelCode: getFirstString(record, ["model_code", "modelcode"]),
    patternName: localizedNameFromKeys(
      record,
      ["pattern_names", "patternnames"],
      getFirstString(record, ["pattern_name", "patternname"]),
    ),
    patternCode: getFirstString(record, ["pattern_code", "patterncode"]),
    groupSubOneName: localizedNameFromKeys(
      record,
      ["groupsubonenames", "group_sub_one_names"],
      getFirstString(record, ["groupsubonename", "group_sub_one_name"]),
    ),
    groupSubOneCode: getFirstString(record, [
      "groupsubonecode",
      "group_sub_one_code",
    ]),
    groupSubTwoName: localizedNameFromKeys(
      record,
      ["groupsubtwonames", "group_sub_two_names"],
      getFirstString(record, ["groupsubtwoname", "group_sub_two_name"]),
    ),
    groupSubTwoCode: getFirstString(record, [
      "groupsubtwocode",
      "group_sub_two_code",
    ]),
    manufacturerName: localizedNameFromKeys(
      record,
      ["manufacturernames", "manufacturer_names"],
      getFirstString(record, ["manufacturername", "manufacturer_name"]),
    ),
    manufacturerCode: getFirstString(record, [
      "manufacturercode",
      "manufacturer_code",
    ]),
    shelfCode: getFirstString(record, ["shelfcode", "shelf_code", "shelfCode"]),
    shelfName: getFirstString(record, ["shelfname", "shelf_name", "shelfName"]),
    checksum: getFirstString(record, ["checksum"]),
    price: getPrice(record),
    imageUri: getFirstString(record, ["imageuri", "image_uri"]),
    colorSelect: getFirstString(record, ["colorselect", "color_select"]),
    colorSelectHex: getFirstString(record, [
      "colorselecthex",
      "color_select_hex",
    ]),
    unitCount: getFirstNumber(record, ["unit_count", "unitcount"]),
    allUnitNames: getFirstString(record, ["all_unitnames", "allunitnames"]),
    balanceQty: getFirstNumber(record, [
      "availableqty",
      "balanceqty",
      "balanceqty",
    ]),
    balanceFormatted: getFirstString(record, [
      "balance_formatted",
      "balanceformatted",
    ]),
    standValue: getFirstNumber(record, [
      "standvalue",
      "stand_value",
      "barcoderefunitstand",
      "barcode_ref_unit_stand",
    ]),
    divideValue: getFirstNumber(record, [
      "dividevalue",
      "divide_value",
      "barcoderefunitdivide",
      "barcode_ref_unit_divide",
    ]),
    itemType: getFirstNumber(record, ["itemtype", "itemtype"]),
    productType: isRecord(record["producttype"])
      ? String((record["producttype"] as Record<string, unknown>).code ?? "")
      : "",
    foodType: getFirstNumber(record, ["foodtype", "food_type"]),
    materialType: getFirstNumber(record, ["materialtype", "material_type"]),
    taxType: getFirstNumber(record, ["taxtype", "taxtype"]),
    vatType: getFirstNumber(record, ["vattype", "vattype"]),
    vatCal: getFirstNumber(record, ["vatcal", "vat_cal"]),
    isStock: getFirstNumber(record, ["isstock", "is_stock"]),
    isMainBarcode: getFirstBoolean(record, ["ismainbarcode", "ismainbarcode"]),
    isMainItem: getFirstBoolean(record, ["ismainitem", "is_main_item"]),
    isUseSubBarcodes: getFirstBoolean(record, [
      "isusesubbarcodes",
      "is_use_sub_barcodes",
    ]),
    useImageOrColor: getFirstBoolean(record, [
      "useimageorcolor",
      "use_image_or_color",
    ]),
    condition: getFirstBoolean(record, ["condition"]),
    isSumPoint: getFirstBoolean(record, ["issumpoint", "is_sum_point"]),
    isDividend: getFirstBoolean(record, ["isdividend", "is_dividend"]),
    isALaCarte: getFirstBoolean(record, ["isalacarte", "is_a_la_carte"]),
    isSplitUnitPrint: getFirstBoolean(record, [
      "issplitunitprint",
      "is_split_unit_print",
    ]),
    isOnlyStaff: getFirstBoolean(record, ["isonlystaff", "is_only_staff"]),
    isStockForRestaurant: getFirstBoolean(record, [
      "isstockforrestaurant",
      "is_stock_for_restaurant",
    ]),
    isDiscountPointOfPurchase: getFirstBoolean(record, [
      "isdiscountpointofpurchase",
      "is_discount_point_of_purchase",
    ]),
    isAlert: getFirstBoolean(record, ["isalert", "is_alert"]),
    isDisable: getFirstBoolean(record, ["isdisable", "is_disable"]),
    showIsDividend: getFirstString(record, [
      "showisdividend",
      "show_is_dividend",
    ]),
    rowNumber: getFirstNumber(record, ["rownumber", "row_number"]),
    maxDiscount: getFirstString(record, ["maxdiscount", "max_discount"]),
    discount: getFirstString(record, ["discount"]),
    description: getFirstString(record, ["description"]),
    alertDescription: getFirstString(record, [
      "alertdescription",
      "alert_description",
    ]),
    priceCount: getArrayCount(record, ["prices"]),
    refBarcodeCount: getArrayCount(record, ["refbarcodes", "ref_barcodes"]),
    subBarcodeCount: getArrayCount(record, [
      "barcodes",
      "subbarcodes",
      "sub_barcodes",
    ]),
    bomCount: getArrayCount(record, ["bom"]),
    optionCount: getArrayCount(record, ["options"]),
    orderTypeCount: getArrayCount(record, ["ordertypes", "order_types"]),
    dimensionCount: getArrayCount(record, ["dimensions"]),
    businessTypeCount: getArrayCount(record, [
      "businesstypes",
      "business_types",
    ]),
    ignoreBranchCount: getArrayCount(record, [
      "ignorebranches",
      "ignore_branches",
    ]),
    branchCount: getArrayCount(record, ["branches"]),
    categoryCount: getArrayCount(record, ["categorys", "categories"]),
    timeForSaleCount: getArrayCount(record, ["timeforsales", "time_for_sales"]),
    fixedCostCount: getArrayCount(record, ["fixedcost", "fixed_cost"]),
  };
}

function getPrice(record: Record<string, unknown>): number {
  const direct = getFirstNumber(record, ["price1", "price", "price_retail"]);
  if (direct > 0) return direct;
  const prices = record.prices;
  if (!Array.isArray(prices)) return 0;
  const retail = prices.find((price) => {
    if (!isRecord(price)) return false;
    return (
      getNumber(price, "keynumber") === 1 || getNumber(price, "keynumber") === 1
    );
  });
  return isRecord(retail) ? getNumber(retail, "price") : 0;
}

function extractBarcodeRecordPayload(value: unknown): unknown {
  if (Array.isArray(value)) return value[0];
  if (!isRecord(value)) return null;
  if (isRecord(value.data)) return value.data;
  if (Array.isArray(value.data)) return value.data[0];
  if (isRecord(value.item)) return value.item;
  if (isRecord(value.barcode)) return value.barcode;
  return value;
}

function detailKey(item: ProductBarcodeRecord): string {
  return barcodeIdentity(item);
}

function getFirstString(
  record: Record<string, unknown>,
  keys: string[],
): string {
  for (const key of keys) {
    const value = getString(record, key);
    if (value) return value;
  }
  return "";
}

function getFirstNumber(
  record: Record<string, unknown>,
  keys: string[],
): number {
  for (const key of keys) {
    if (!Object.prototype.hasOwnProperty.call(record, key)) continue;
    return getNumber(record, key);
  }
  return 0;
}

function getFirstBoolean(
  record: Record<string, unknown>,
  keys: string[],
): boolean {
  for (const key of keys) {
    if (!Object.prototype.hasOwnProperty.call(record, key)) continue;
    const value = record[key];
    if (typeof value === "boolean") return value;
    if (typeof value === "number") return value !== 0;
    if (typeof value === "string") {
      const normalized = value.trim().toLowerCase();
      if (!normalized) return false;
      return (
        normalized === "true" ||
        normalized === "1" ||
        normalized === "yes" ||
        normalized === "y"
      );
    }
  }
  return false;
}

function localizedNameFromKeys(
  record: Record<string, unknown>,
  keys: string[],
  fallback = "",
): string {
  for (const key of keys) {
    const names = getNames(record, key);
    const name = localizedName(names, "");
    if (name) return name;
    const raw = getString(record, key);
    if (raw) return raw;
  }
  return fallback;
}

function getArrayCount(
  record: Record<string, unknown>,
  keys: string[],
): number {
  for (const key of keys) {
    const value = record[key];
    if (Array.isArray(value)) return value.length;
  }
  return 0;
}

function getNames(
  record: Record<string, unknown>,
  key: string,
): LocalizedName[] | undefined {
  const value = record[key];
  return Array.isArray(value)
    ? value.filter(isRecord).map((item) => ({
        code: getString(item, "code"),
        name: getString(item, "name"),
      }))
    : undefined;
}

function getString(record: Record<string, unknown>, key: string): string {
  const value = record[key];
  return typeof value === "string" ? value.trim() : "";
}

function getNumber(record: Record<string, unknown>, key: string): number {
  const value = record[key];
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const parsed = Number(value.replace(/,/g, ""));
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

const moneyFormatter = new Intl.NumberFormat("th-TH", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});
const numberFormatter = new Intl.NumberFormat("th-TH", {
  maximumFractionDigits: 4,
});

function formatMoney(value: number): string {
  return moneyFormatter.format(value || 0);
}

function formatNumber(value: number): string {
  return numberFormatter.format(value || 0);
}

function formatCount(value: number): string {
  return value > 0 ? value.toLocaleString("th-TH") : "";
}

function formatBoolean(value: boolean, text: BarcodeText): string {
  return value ? text.yes : text.no;
}

function formatCodeName(code: string, name: string): string {
  if (code && name) return `${code} - ${name}`;
  return code || name;
}

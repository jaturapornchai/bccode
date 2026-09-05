"use client";

import {
  Barcode,
  CheckSquare,
  ChevronLeft,
  ChevronRight,
  Copy,
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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ProductBarcodeFormDialog } from "@/components/product-barcode/barcode-form";
import { BusinessImageGallery } from "@/components/product-barcode/business-image-editor";
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
} from "@/lib/product-barcode/utils";
import { normalizeBusinessCode } from "@/lib/business-code";
import { cn } from "@/lib/utils";
import { authFetch, getAuthSession } from "@/lib/client-auth-session";
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
  onOpenLabelPrint?: () => void;
  onOpenProduct?: (itemCode: string) => void;
};

type ProductBarcodeRecord = {
  raw: Record<string, unknown>;
  guidFixed: string;
  barcode: string;
  name: string;
  unitName: string;
  unitCode: string;
  itemCode: string;
  sellingPrice: string;
  standValue: number;
  divideValue: number;
  imageURI: string;
  images: ProductBarcodeObject["images"];
  videos: ProductBarcodeObject["videos"];
  description: string;
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
  onOpenLabelPrint,
  onOpenProduct,
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
  const [pageIndex, setPageIndex] = useState(0);
  const [loading, setLoading] = useState(false);
  const setNotice = pushNotice;
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setPageIndex(0);
      setSearch(searchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchInput]);
  const [selectMode, setSelectMode] = useState(false);
  const [selectedKey, setSelectedKey] = useState("");
  const [checkedBarcodes, setCheckedBarcodes] = useState<string[]>([]);
  const [compactRows, setCompactRows] = useState<boolean>(() => {
    if (typeof window === "undefined") return true;
    const saved = window.localStorage.getItem("bcproductbarcodecompact");
    return saved !== null ? saved === "true" : true;
  });
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<"create" | "edit">("create");
  const [editorGuid, setEditorGuid] = useState("");
  const [editorScopeKey, setEditorScopeKey] = useState("");
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
  const activeBusinessCode = normalizeBusinessCode(workspace?.company?.code);
  const companyScopeKey = `${activeHoldingCode}\u0000${activeBusinessCode}`;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const companyRequiredMessage =
    lang === "th"
      ? "กรุณาเลือกบริษัทก่อนจัดการบาร์โค้ด"
      : "Select a company before managing barcodes.";
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

  const loadBarcodes = useCallback(async () => {
    if (!auth || !activeHoldingCode) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }
    if (!activeBusinessCode) {
      setItems([]);
      setDetailItems({});
      setTotal(0);
      setSelectedKey("");
      setCheckedBarcodes([]);
      setNotice({ type: "error", text: companyRequiredMessage });
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
        businesscode: activeBusinessCode,
        keyword: search.trim(),
        limit: pageSize,
        offset: pageIndex * pageSize,
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
    activeBusinessCode,
    activeHoldingCode,
    auth,
    companyRequiredMessage,
    pageIndex,
    search,
    text.apiRequired,
    text.requestFailed,
  ]);

  const loadBarcodeDetail = useCallback(
    async (item: ProductBarcodeRecord): Promise<ProductBarcodeRecord | null> => {
      if (!auth || !item.guidFixed) return null;
      const key = detailKey(item);
      if (detailItems[key]) return detailItems[key];

      loadDetailAbortRef.current?.abort();
      const controller = new AbortController();
      loadDetailAbortRef.current = controller;

      try {
        const response = await authFetch(
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
        if (!response.ok || data.success === false) return null;

        const detail = extractBarcodeRecordPayload(data.data);
        if (!detail) return null;
        const mergedDetail = isRecord(detail)
          ? { ...item.raw, ...detail }
          : detail;
        const normalized = normalizeBarcodeRecord(mergedDetail);
        setDetailItems((current) => ({ ...current, [key]: normalized }));
        return normalized;
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") return null;
        return null;
      }
    },
    [auth, detailItems],
  );

  useEffect(() => {
    loadBarcodesAbortRef.current?.abort();
    loadDetailAbortRef.current?.abort();
    setItems([]);
    setDetailItems({});
    setTotal(0);
    setSelectedKey("");
    setCheckedBarcodes([]);
    setPageIndex(0);
  }, [companyScopeKey]);

  useEffect(() => {
    if (pageIndex < pageCount) return;
    setPageIndex(pageCount - 1);
  }, [pageCount, pageIndex]);

  useEffect(() => {
    if (!auth) return;
    if (!activeHoldingCode || !activeBusinessCode) {
      void loadBarcodes();
      return;
    }
    const timer = window.setTimeout(() => {
      void loadBarcodes();
    }, 300);
    return () => window.clearTimeout(timer);
  }, [activeBusinessCode, activeHoldingCode, auth, loadBarcodes]);

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

  async function goToPage(nextPage: number) {
    if (nextPage < 0 || nextPage >= pageCount || nextPage === pageIndex) return;
    if (!(await canDiscardEditor())) return;
    setEditorOpen(false);
    setEditorDirty(false);
    setPageIndex(nextPage);
  }

  async function closeEditor() {
    if (!(await canDiscardEditor())) return;
    setEditorOpen(false);
    setEditorGuid("");
    setEditorScopeKey("");
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

  function hasCompanyScope() {
    if (activeHoldingCode && activeBusinessCode) return true;
    setNotice({
      type: "error",
      text: activeHoldingCode ? companyRequiredMessage : text.apiRequired,
    });
    return false;
  }

  async function openCreateEditor() {
    if (!(await canDiscardEditor())) return;
    if (!hasCompanyScope()) return;
    setEditorMode("create");
    setEditorGuid("");
    setEditorScopeKey(companyScopeKey);
    setEditorDirty(false);
    const next = emptyProductBarcode();
    next.holdingcode = activeHoldingCode;
    next.businesscode = activeBusinessCode;
    setEditorBarcode(next);
    setEditorOpen(true);
  }

  async function openEditEditor(itemOverride?: ProductBarcodeRecord) {
    if (!(await canDiscardEditor())) return;
    if (!hasCompanyScope()) return;
    const target = itemOverride ?? selected;
    if (!target) return;
    const detail = await loadBarcodeDetail(target);
    if (!detail) {
      setNotice({
        type: "error",
        text: lang === "th" ? "โหลดรายละเอียดบาร์โค้ดไม่สำเร็จ" : "Unable to load barcode details.",
      });
      return;
    }
    setEditorMode("edit");
    setSelectedKey(barcodeIdentity(detail));
    setEditorGuid(detail.guidFixed);
    setEditorScopeKey(companyScopeKey);
    setEditorDirty(false);
    const base = emptyProductBarcode();
    base.holdingcode = activeHoldingCode;
    base.businesscode = activeBusinessCode;
    setEditorBarcode(rawToProductBarcode(detail.raw, base));
    setEditorOpen(true);
  }

  async function openCopyEditor() {
    if (!(await canDiscardEditor())) return;
    if (!hasCompanyScope()) return;
    if (!selected) return;
    const detail = await loadBarcodeDetail(selected);
    if (!detail) {
      setNotice({
        type: "error",
        text: lang === "th" ? "โหลดรายละเอียดบาร์โค้ดไม่สำเร็จ" : "Unable to load barcode details.",
      });
      return;
    }
    setEditorMode("create");
    setEditorGuid("");
    setEditorScopeKey(companyScopeKey);
    setEditorDirty(true);
    const base = emptyProductBarcode();
    base.holdingcode = activeHoldingCode;
    base.businesscode = activeBusinessCode;
    const next = rawToProductBarcode(detail.raw, base);
    next.guidfixed = "";
    next.holdingcode = activeHoldingCode;
    next.businesscode = activeBusinessCode;
    setEditorBarcode(next);
    setEditorOpen(true);
  }

  function copyCurrentEditorValue() {
    const next = JSON.parse(
      JSON.stringify(editorBarcode),
    ) as ProductBarcodeObject;
    next.guidfixed = "";
    next.holdingcode = activeHoldingCode;
    next.businesscode = activeBusinessCode;
    setEditorMode("create");
    setEditorGuid("");
    setEditorScopeKey(companyScopeKey);
    setEditorBarcode(next);
    setEditorDirty(true);
  }

  async function saveEditor(value: ProductBarcodeObject, keepOpen = false) {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }
    if (!activeHoldingCode || !activeBusinessCode) {
      setNotice({ type: "error", text: companyRequiredMessage });
      return;
    }
    if (editorScopeKey !== companyScopeKey) {
      setNotice({
        type: "error",
        text:
          lang === "th"
            ? "บริษัทถูกเปลี่ยนระหว่างแก้ไข ข้อมูลที่กรอกยังอยู่ กรุณากลับไปเลือกบริษัทเดิมหรือปิดฟอร์มแล้วเริ่มใหม่"
            : "The company changed while editing. Your draft is preserved; return to the original company or close and start again.",
      });
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
      const scopedValue = {
        ...value,
        holdingcode: activeHoldingCode,
        businesscode: activeBusinessCode,
      };
      const data =
        editorMode === "edit"
          ? await updateBarcode(auth, guid, scopedValue)
          : await createBarcode(auth, scopedValue);
      if (!data.success) {
        throw new Error(data.message || text.requestFailed);
      }

      if (keepOpen) {
        // "Save & add new": reset to a blank record and stay open
        const next = emptyProductBarcode();
        next.holdingcode = activeHoldingCode;
        next.businesscode = activeBusinessCode;
        setEditorGuid("");
        setEditorScopeKey(companyScopeKey);
        setEditorDirty(false);
        setEditorBarcode(next);
      } else {
        setEditorOpen(false);
        setEditorGuid("");
        setEditorScopeKey("");
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
      setEditorScopeKey("");
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
              <Button
                variant={compactRows ? "secondary" : "outline"}
                size="sm"
                type="button"
                onClick={() => {
                  setCompactRows((prev) => {
                    const next = !prev;
                    if (typeof window !== "undefined") {
                      window.localStorage.setItem("bcproductbarcodecompact", String(next));
                    }
                    return next;
                  });
                }}
                className="h-9 text-xs"
                title={compactRows ? "คลิกเพื่อขยายบรรทัด" : "คลิกเพื่อย่อบรรทัด"}
              >
                {compactRows ? "ย่อบรรทัด" : "ขยายบรรทัด"}
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
              variant={compactRows ? "secondary" : "outline"}
              size="sm"
              type="button"
              onClick={() => {
                setCompactRows((prev) => {
                  const next = !prev;
                  if (typeof window !== "undefined") {
                    window.localStorage.setItem("bcproductbarcodecompact", String(next));
                  }
                  return next;
                });
              }}
              className="h-9 text-xs"
              title={compactRows ? "คลิกเพื่อขยายบรรทัด" : "คลิกเพื่อย่อบรรทัด"}
            >
              {compactRows ? "ย่อบรรทัด" : "ขยายบรรทัด"}
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
          </CardHeader>
          <CardContent className="grid min-h-[420px] p-0 xl:min-h-0 xl:flex-1 xl:grid-rows-[auto_minmax(0,1fr)_auto]">
            <div className="bc-list-header hidden grid-cols-[1.2fr_1.8fr] gap-x-2 lg:grid 2xl:grid-cols-[1.35fr_2.2fr_0.9fr_1.1fr]">
              <span>{text.barcode}</span>
              <span>{text.productName}</span>
              <span className="hidden 2xl:block">{text.unit}</span>
              <span className="hidden 2xl:block">{text.itemCode}</span>
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
                        checked={checkedBarcodes.includes(rowKey)}
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
                        compact={compactRows}
                        editing={
                          selected
                            ? barcodeIdentity(selected) ===
                                barcodeIdentity(item) &&
                              editorOpen &&
                              editorMode === "edit"
                            : false
                        }
                        selectMode={selectMode}
                        text={text}
                      />
                    );
                  })}
                </div>
              )}
            </div>
            {total > pageSize ? (
              <div
                className="flex flex-wrap items-center justify-between gap-2 border-t border-border px-3 py-2 text-xs text-muted-foreground"
                data-testid="barcode-pagination"
              >
                <span>
                  {lang === "th" ? "รายการ" : "Items"} {pageIndex * pageSize + 1}
                  –{Math.min(pageIndex * pageSize + items.length, total)} {lang === "th" ? "จาก" : "of"}{" "}
                  {total.toLocaleString(lang === "th" ? "th-TH" : "en-US")}
                </span>
                <div className="flex items-center gap-2">
                  <Button
                    aria-label={lang === "th" ? "หน้าก่อนหน้า" : "Previous page"}
                    data-testid="barcode-page-prev"
                    disabled={loading || pageIndex === 0}
                    onClick={() => void goToPage(pageIndex - 1)}
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    <ChevronLeft size={15} />
                  </Button>
                  <span className="min-w-20 text-center font-medium text-foreground">
                    {lang === "th" ? "หน้า" : "Page"} {pageIndex + 1} / {pageCount}
                  </span>
                  <Button
                    aria-label={lang === "th" ? "หน้าถัดไป" : "Next page"}
                    data-testid="barcode-page-next"
                    disabled={loading || pageIndex + 1 >= pageCount}
                    onClick={() => void goToPage(pageIndex + 1)}
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    <ChevronRight size={15} />
                  </Button>
                </div>
              </div>
            ) : null}
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
            companyGuid={workspace?.company?.guidfixed ?? workspace?.branch?.companyguid}
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
            auth={auth}
            item={selected}
            onDelete={() => void deleteCurrentItem()}
            onEdit={() => void openEditEditor()}
            language={lang}
            onOpenLabelPrint={onOpenLabelPrint}
            onOpenProduct={onOpenProduct}
            text={text}
          />
        )}
      </div>
      {confirmationDialog}
    </div>
  );

  if (embedded) return content;
  return <main className="min-h-screen bg-background p-3">{content}</main>;
}

function BarcodeRow({
  checked,
  compact = true,
  index,
  item,
  onSelect,
  onToggleChecked,
  selected,
  editing,
  selectMode,
  text,
}: {
  checked: boolean;
  compact?: boolean;
  index: number;
  item: ProductBarcodeRecord;
  onSelect: () => void;
  onToggleChecked: () => void;
  selected: boolean;
  editing: boolean;
  selectMode: boolean;
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
      data-compact={compact ? "true" : "false"}
      className={cn(
        "bc-list-row grid gap-x-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 lg:grid-cols-[1.2fr_1.8fr] 2xl:grid-cols-[1.35fr_2.2fr_0.9fr_1.1fr]",
        compact ? "is-compact py-0.5 items-center min-h-[28px]" : "py-2 items-start",
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
                "grid size-5 shrink-0 place-items-center rounded-md border",
                checked && "border-primary bg-primary text-primary-foreground",
              )}
            >
              {checked ? <CheckSquare size={12} /> : null}
            </span>
          ) : null}
          <span className={cn("min-w-0 font-semibold", compact ? "bc-cell-text text-xs sm:text-sm" : "truncate")} title={item.barcode || "-"}>
            {item.barcode || "-"}
          </span>
        </div>
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.productName}
        </span>
        <div
          className={cn("font-medium", compact ? "bc-cell-text text-xs sm:text-sm" : "line-clamp-2")}
          title={`${item.name || "-"} (${text.retailPrice} 1: ${item.sellingPrice || "-"})`}
        >
          {item.name || "-"}
        </div>
        {!compact ? (
          <div className="mt-0.5 truncate text-xs text-muted-foreground">
            {text.retailPrice} 1: {item.sellingPrice || "-"}
          </div>
        ) : null}
      </div>
      <div className="hidden min-w-0 2xl:block">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.unit}
        </span>
        <div className={cn(compact ? "bc-cell-text text-xs text-muted-foreground" : "truncate")} title={item.unitName || item.unitCode || "-"}>
          {item.unitName || item.unitCode || "-"}
        </div>
      </div>
      <div className="hidden min-w-0 2xl:block">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">
          {text.itemCode}
        </span>
        <div className={cn(compact ? "bc-cell-text text-xs text-muted-foreground" : "truncate")} title={item.itemCode || "-"}>
          {item.itemCode || "-"}
        </div>
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
  auth,
  item,
  language,
  onDelete,
  onEdit,
  onOpenLabelPrint,
  onOpenProduct,
  text,
}: {
  auth: AuthSession | null;
  item: ProductBarcodeRecord | null;
  language: LanguageCode;
  onDelete: () => void;
  onEdit: () => void;
  onOpenLabelPrint?: () => void;
  onOpenProduct?: (itemCode: string) => void;
  text: BarcodeText;
}) {
  const fields = item
    ? [
        { label: text.itemCode, value: item.itemCode },
        { label: text.productName, value: item.name },
        {
          label: text.unit,
          value: formatCodeName(item.unitCode, item.unitName),
        },
        { label: `${text.retailPrice} 1`, value: item.sellingPrice || "-" },
        { label: text.divideValue, value: formatNumber(item.divideValue) },
        { label: text.standValue, value: formatNumber(item.standValue) },
      ]
    : [];

  return (
    <Card className="min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col">
      <CardHeader className="shrink-0 border-b border-border p-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Package size={18} />
            {text.detailTitle}
          </CardTitle>
          <div className="flex flex-wrap justify-end gap-2">
            <Button
              disabled={!item || !onOpenLabelPrint}
              onClick={onOpenLabelPrint}
              size="sm"
              variant="secondary"
            >
              <Printer size={16} />
              {language === "th" ? "พิมพ์ฉลาก" : "Print labels"}
            </Button>
            <Button
              disabled={!item?.itemCode || !onOpenProduct}
              onClick={() => {
                if (item?.itemCode && onOpenProduct) onOpenProduct(item.itemCode);
              }}
              size="sm"
              variant="secondary"
            >
              <Package size={16} />
              {language === "th" ? "ไปเติมรายละเอียดสินค้า" : "Complete product details"}
            </Button>
            <Button disabled={!item} onClick={onEdit} size="sm" variant="outline">
              <Pencil size={16} />
              {text.edit}
            </Button>
            <Button disabled={!item} onClick={onDelete} size="sm" variant="outline">
              <Trash2 size={16} />
              {text.delete}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent
        className="grid content-start items-start gap-3 overflow-auto p-3 xl:min-h-0 xl:flex-1"
        style={{
          gridTemplateColumns:
            "repeat(auto-fit, minmax(min(100%, 20rem), 1fr))",
        }}
      >
        {!item ? (
          <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">
            {text.noSelection}
          </div>
        ) : (
          <>
            <div
              className="rounded-2xl border border-border bg-muted/30 p-3"
              data-testid="barcode-detail-summary"
            >
              <p className="text-xs font-semibold text-muted-foreground">{text.barcode}</p>
              <p className="break-all text-2xl font-semibold">{item.barcode}</p>
              <p className="mt-1 text-sm text-muted-foreground">{item.name}</p>
            </div>
            <DetailSection fields={fields} title={text.basicInfo} />
            <BusinessImageGallery
              auth={auth}
              sources={[
                {
                  key: item.guidFixed || item.barcode,
                  label: `${text.barcode} ${item.barcode}`,
                  imageuri: item.imageURI,
                  images: item.images,
                  videos: item.videos,
                },
              ]}
              title={language === "th" ? "สื่อบาร์โค้ดและบรรจุภัณฑ์" : "Barcode and package media"}
            />
            <section className="rounded-2xl border border-border p-3">
              <h3 className="mb-2 text-sm font-semibold">
                {language === "th" ? "รายละเอียดบาร์โค้ด" : "Barcode description"}
              </h3>
              <p className="whitespace-pre-wrap break-words text-sm text-muted-foreground">
                {item.description || "—"}
              </p>
            </section>
          </>
        )}
      </CardContent>
    </Card>
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
    <section
      className="rounded-2xl border border-border p-3"
      data-testid="barcode-detail-fields"
    >
      <h3 className="mb-1.5 text-xs sm:text-sm font-bold text-primary">{title}</h3>
      <div className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
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
    <div className="min-w-0 rounded-lg border border-border bg-background px-2.5 py-1.5 flex flex-col justify-center">
      <p className="text-[10px] sm:text-[11px] font-semibold text-muted-foreground truncate" title={label}>{label}</p>
      <p className="mt-0.5 break-words text-xs sm:text-sm font-semibold">{value || "-"}</p>
    </div>
  );
}

function readAuthSession(): AuthSession | null {
  return getAuthSession();
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
    guidFixed: getFirstString(record, ["guidfixed"]),
    barcode: getFirstString(record, ["barcode"]),
    name: localizedNameFromKeys(
      record,
      ["names"],
      getFirstString(record, ["name0", "name", "itemname"]),
    ),
    unitName: localizedNameFromKeys(
      record,
      ["itemunitnames", "unitnames"],
      getFirstString(record, ["unitname"]),
    ),
    unitCode: getFirstString(record, ["itemunitcode", "unitcode"]),
    itemCode: getFirstString(record, ["itemcode"]),
    sellingPrice: getSellingPrice(record),
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
    imageURI: getFirstString(record, ["imageuri"]),
    images: getProductImages(record),
    videos: getProductVideos(record),
    description: getFirstString(record, ["description"]),
  };
}

function getProductImages(record: Record<string, unknown>): ProductBarcodeObject["images"] {
  if (!Array.isArray(record.images)) return [];
  return record.images
    .filter(isRecord)
    .map((image, index) => ({
      xorder: getFirstNumber(image, ["xorder"]) || index + 1,
      uri: getFirstString(image, ["uri"]),
    }))
    .filter((image) => image.uri);
}

function getProductVideos(record: Record<string, unknown>): ProductBarcodeObject["videos"] {
  if (!Array.isArray(record.videos)) return [];
  return record.videos
    .filter(isRecord)
    .map((video, index) => ({
      xorder: getFirstNumber(video, ["xorder"]) || index + 1,
      uri: getFirstString(video, ["uri"]),
      posteruri: getFirstString(video, ["posteruri"]),
    }))
    .filter((video) => video.uri);
}

function getSellingPrice(record: Record<string, unknown>): string {
  const prices = Array.isArray(record.prices) ? record.prices.filter(isRecord) : [];
  const price = prices.find((item) => String(item.keynumber ?? "") === "1") ?? prices[0];
  const value = price?.price ?? record.price;
  if (typeof value === "string") return value.trim();
  return typeof value === "number" && Number.isFinite(value) ? String(value) : "";
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

const numberFormatter = new Intl.NumberFormat("th-TH", {
  maximumFractionDigits: 4,
});

function formatNumber(value: number): string {
  return numberFormatter.format(value || 0);
}

function formatCodeName(code: string, name: string): string {
  if (code && name) return `${code} - ${name}`;
  return code || name;
}

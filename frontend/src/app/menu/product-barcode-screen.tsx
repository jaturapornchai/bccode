"use client";

import {
  AlertCircle,
  Barcode,
  CheckSquare,
  Copy,
  Download,
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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { ProductBarcodeFormDialog } from "@/components/product-barcode/barcode-form";
import { deriveMainApiUrl } from "@/lib/backend-url";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import {
  emptyProductBarcode,
  type ProductBarcode as ProductBarcodeObject,
} from "@/lib/product-barcode/types";
import { rawToProductBarcode } from "@/lib/product-barcode/utils";
import { cn } from "@/lib/utils";
import {
  branchDisplayName,
  localizedName,
  shopDisplayName,
  type AuthSession,
  type LocalizedName,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";

type ProductBarcodeScreenProps = {
  embedded?: boolean;
  language?: LanguageCode;
};

type ProductBarcodeRecord = {
  raw: Record<string, unknown>;
  guidFixed: string;
  shopId: string;
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
  productType: number;
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
const PRODUCT_SPLIT_STORAGE_KEY = "bc_product_barcode_split_left_v2";
const PRODUCT_SPLIT_MIN_LEFT = 5;
const PRODUCT_SPLIT_MAX_LEFT = 95;

function clampProductSplitLeft(value: number) {
  if (!Number.isFinite(value)) return PRODUCT_SPLIT_DEFAULT_LEFT;
  return Math.min(PRODUCT_SPLIT_MAX_LEFT, Math.max(PRODUCT_SPLIT_MIN_LEFT, value));
}

type BarcodeApiResponse = {
  success?: boolean;
  message?: string;
  data?: unknown;
  total?: number;
};

type Notice = { type: "success" | "error" | "info"; text: string } | null;

const pageSize = 80;

const barcodeText = {
  th: {
    title: "สินค้า",
    subtitle: "จัดการรายการสินค้าแบบเดียวกับหน้าจอ Flutter เดิม",
    refresh: "รีเฟรช",
    export: "ส่งออก",
    selectDelete: "เลือกเพื่อลบ",
    cancelSelect: "ยกเลิกเลือก",
    add: "เพิ่ม",
    search: "ค้นหา บาร์โค้ด ชื่อสินค้า หรือรหัสสินค้า",
    filter: "ตัวกรอง",
    groupCode: "กลุ่มสินค้า",
    brandCode: "ยี่ห้อ",
    categoryCode: "หมวดสินค้า",
    classCode: "Class",
    designCode: "Design",
    gradeCode: "Grade",
    modelCode: "Model",
    patternCode: "Pattern",
    priceMin: "ราคาต่ำสุด",
    priceMax: "ราคาสูงสุด",
    clearFilter: "ล้างตัวกรอง",
    total: "ทั้งหมด",
    selected: "เลือกแล้ว",
    barcode: "บาร์โค้ด",
    productName: "ชื่อสินค้า",
    unit: "หน่วย",
    itemCode: "รหัสสินค้า",
    balance: "คงเหลือ",
    retailPrice: "ราคาขาย",
    image: "รูป",
    detailTitle: "รายละเอียดสินค้า",
    noSelection: "เลือกรายการด้านซ้ายเพื่อดูรายละเอียด",
    loading: "กำลังโหลดข้อมูลสินค้า",
    noData: "ไม่พบข้อมูลสินค้าจากฐานข้อมูลจริง",
    apiRequired: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้",
    requestFailed: "โหลดข้อมูลไม่สำเร็จ",
    edit: "แก้ไข",
    copy: "คัดลอก",
    delete: "ลบ",
    save: "บันทึก",
    close: "ปิด",
    createTitle: "เพิ่มสินค้า",
    editTitle: "แก้ไขสินค้า",
    jsonPayload: "ข้อมูล JSON ครบทุก field",
    invalidJson: "JSON ไม่ถูกต้อง",
    saveSuccess: "บันทึกสินค้าแล้ว",
    deleteConfirm: "ต้องการลบสินค้าที่เลือกจริงหรือไม่",
    deleteCurrentConfirm: "ต้องการลบจริงหรือไม่",
    unsavedConfirm: "มีข้อมูลที่ยังไม่ได้บันทึก ต้องการทิ้งการแก้ไขหรือไม่?",
    deleteSuccess: "ลบสินค้าแล้ว",
    missingGuid: "รายการที่เลือกไม่มี GUID สำหรับลบ",
    exportSuccess: "ส่งออกข้อมูลที่แสดงอยู่แล้ว",
    tenant: "บริษัท",
    branch: "สาขา",
    multiUnit: "หลายหน่วย",
    basicInfo: "ข้อมูลหลัก",
    classification: "กลุ่ม/หมวด/คุณสมบัติ",
    stockAndUnit: "หน่วยนับ/สต็อก/ราคา",
    taxAndFlags: "ภาษีและสถานะ",
    relations: "ข้อมูลเชื่อมโยง",
    rawFields: "ข้อมูลทั้งหมดจาก API",
    guid: "GUID",
    shopId: "รหัสบริษัท",
    barcodeRef: "บาร์โค้ดอ้างอิง",
    itemGuid: "GUID สินค้า",
    itemGuidFixed: "GUID สินค้าถาวร",
    parentGuid: "GUID แม่",
    checksum: "Checksum",
    shelf: "ชั้นวาง",
    color: "สี",
    standValue: "ตัวตั้ง",
    divideValue: "ตัวหาร",
    unitCount: "จำนวนหน่วยนับ",
    itemType: "ประเภทสินค้า",
    productType: "ประเภทบาร์โค้ด",
    foodType: "ประเภทอาหาร",
    materialType: "ประเภทวัตถุดิบ",
    taxType: "ประเภทภาษี",
    vatType: "ประเภท VAT",
    vatCal: "วิธีคำนวณ VAT",
    isStock: "นับสต็อก",
    isMainBarcode: "บาร์โค้ดหลัก",
    isMainItem: "สินค้าหลัก",
    isUseSubBarcodes: "ใช้บาร์โค้ดย่อย",
    useImageOrColor: "ใช้รูป/สี",
    condition: "มีเงื่อนไข",
    isSumPoint: "สะสมแต้ม",
    isDividend: "ร่วมปันผล",
    isALaCarte: "อลาคาร์ท",
    isSplitUnitPrint: "พิมพ์แยกหน่วย",
    isOnlyStaff: "เฉพาะพนักงาน",
    isStockForRestaurant: "สต็อกร้านอาหาร",
    isDiscountPointOfPurchase: "ลดราคา ณ จุดขาย",
    isAlert: "แจ้งเตือน",
    isDisable: "ปิดใช้งาน",
    showIsDividend: "ข้อความปันผล",
    rowNumber: "ลำดับแถว",
    maxDiscount: "ส่วนลดสูงสุด",
    discount: "ส่วนลด",
    description: "รายละเอียดสินค้า",
    alertDescription: "รายละเอียดแจ้งเตือน",
    manufacturer: "ผู้ผลิต",
    subGroup1: "กลุ่มย่อย 1",
    subGroup2: "กลุ่มย่อย 2",
    prices: "ราคา",
    refBarcodes: "บาร์โค้ดอ้างอิง",
    subBarcodes: "บาร์โค้ดย่อย",
    bom: "สูตรประกอบ",
    options: "ตัวเลือก",
    orderTypes: "ประเภทคำสั่ง",
    dimensions: "มิติ",
    businessTypes: "ประเภทธุรกิจ",
    ignoreBranches: "สาขาที่ไม่ใช้",
    branches: "สาขา",
    categories: "หมวดสินค้า",
    timeForSales: "เวลาขาย",
    fixedCost: "ต้นทุนมาตรฐาน",
    labelPrint: "พิมพ์ป้ายสินค้า",
    priceHistory: "ประวัติราคา",
    bomView: "ดูสูตรการผลิต",
    yes: "ใช่",
    no: "ไม่ใช่",
  },
  en: {
    title: "Product",
    subtitle: "Manage products using the legacy Flutter workflow as reference.",
    refresh: "Refresh",
    export: "Export",
    selectDelete: "Select delete",
    cancelSelect: "Cancel select",
    add: "Add",
    search: "Search barcode, product name, or item code",
    filter: "Filter",
    groupCode: "Product group",
    brandCode: "Brand",
    categoryCode: "Category",
    classCode: "Class",
    designCode: "Design",
    gradeCode: "Grade",
    modelCode: "Model",
    patternCode: "Pattern",
    priceMin: "Min price",
    priceMax: "Max price",
    clearFilter: "Clear filters",
    total: "Total",
    selected: "Selected",
    barcode: "Barcode",
    productName: "Product name",
    unit: "Unit",
    itemCode: "Item code",
    balance: "Balance",
    retailPrice: "Retail price",
    image: "Image",
    detailTitle: "Product detail",
    noSelection: "Select a row on the left to view detail",
    loading: "Loading product data",
    noData: "No product data found in the real database",
    apiRequired: "Please login and select a company before opening this screen.",
    requestFailed: "Could not load data",
    edit: "Edit",
    copy: "Copy",
    delete: "Delete",
    save: "Save",
    close: "Close",
    createTitle: "Add product",
    editTitle: "Edit product",
    jsonPayload: "Full-field JSON payload",
    invalidJson: "Invalid JSON",
    saveSuccess: "Product saved.",
    deleteConfirm: "Confirm delete selected products?",
    deleteCurrentConfirm: "Confirm delete this product?",
    unsavedConfirm: "You have unsaved changes. Discard them?",
    deleteSuccess: "Products deleted.",
    missingGuid: "Selected rows do not have GUIDs for deletion.",
    exportSuccess: "Exported the currently displayed data.",
    tenant: "Company",
    branch: "Branch",
    multiUnit: "Multi-unit",
    basicInfo: "Basic info",
    classification: "Classification",
    stockAndUnit: "Unit, stock, and price",
    taxAndFlags: "Tax and status",
    relations: "Linked data",
    rawFields: "All API fields",
    guid: "GUID",
    shopId: "Company ID",
    barcodeRef: "Barcode reference",
    itemGuid: "Item GUID",
    itemGuidFixed: "Fixed item GUID",
    parentGuid: "Parent GUID",
    checksum: "Checksum",
    shelf: "Shelf",
    color: "Color",
    standValue: "Stand value",
    divideValue: "Divide value",
    unitCount: "Unit count",
    itemType: "Item type",
    productType: "Product type",
    foodType: "Food type",
    materialType: "Material type",
    taxType: "Tax type",
    vatType: "VAT type",
    vatCal: "VAT calculation",
    isStock: "Stock item",
    isMainBarcode: "Main barcode",
    isMainItem: "Main item",
    isUseSubBarcodes: "Use sub barcodes",
    useImageOrColor: "Use image/color",
    condition: "Condition",
    isSumPoint: "Sum point",
    isDividend: "Dividend",
    isALaCarte: "A la carte",
    isSplitUnitPrint: "Split unit print",
    isOnlyStaff: "Staff only",
    isStockForRestaurant: "Restaurant stock",
    isDiscountPointOfPurchase: "POS discount",
    isAlert: "Alert",
    isDisable: "Disabled",
    showIsDividend: "Dividend text",
    rowNumber: "Row number",
    maxDiscount: "Max discount",
    discount: "Discount",
    description: "Description",
    alertDescription: "Alert description",
    manufacturer: "Manufacturer",
    subGroup1: "Sub group 1",
    subGroup2: "Sub group 2",
    prices: "Prices",
    refBarcodes: "Ref barcodes",
    subBarcodes: "Sub barcodes",
    bom: "BOM",
    options: "Options",
    orderTypes: "Order types",
    dimensions: "Dimensions",
    businessTypes: "Business types",
    ignoreBranches: "Ignored branches",
    branches: "Branches",
    categories: "Categories",
    timeForSales: "Sale times",
    fixedCost: "Fixed cost",
    labelPrint: "Print label",
    priceHistory: "Price history",
    bomView: "View BOM",
    yes: "Yes",
    no: "No",
  },
} as const;

type BarcodeText = Record<keyof typeof barcodeText.th, string>;

export function ProductBarcodeScreen({ embedded = false, language = "th" }: ProductBarcodeScreenProps) {
  const lang = normalizeLanguage(language);
  const text: BarcodeText = lang === "th" ? barcodeText.th : barcodeText.en;
  const { confirm, confirmationDialog } = useConfirmDialog();

  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [items, setItems] = useState<ProductBarcodeRecord[]>([]);
  const [detailItems, setDetailItems] = useState<Record<string, ProductBarcodeRecord>>({});
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);
  const [search, setSearch] = useState("");
  const [filterOpen, setFilterOpen] = useState(false);
  const [showImage, setShowImage] = useState(false);
  const [selectMode, setSelectMode] = useState(false);
  const [selectedBarcode, setSelectedBarcode] = useState("");
  const [checkedBarcodes, setCheckedBarcodes] = useState<string[]>([]);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<"create" | "edit">("create");
  const [editorGuid, setEditorGuid] = useState("");
  const [editorBarcode, setEditorBarcode] = useState<ProductBarcodeObject>(() => emptyProductBarcode());
  const [editorDirty, setEditorDirty] = useState(false);
  const [editorSaving, setEditorSaving] = useState(false);
  const [splitLeftPercent, setSplitLeftPercent] = useState(PRODUCT_SPLIT_DEFAULT_LEFT);
  const [resizingSplit, setResizingSplit] = useState(false);
  const splitContainerRef = useRef<HTMLDivElement | null>(null);
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

  useEffect(() => {
    setAuth(readAuthSession());
    setWorkspace(readWorkspaceSession());
  }, []);

  useEffect(() => {
    if (typeof window === "undefined") return;
    const saved = window.localStorage.getItem(PRODUCT_SPLIT_STORAGE_KEY);
    if (!saved) return;
    const next = clampProductSplitLeft(Number(saved));
    setSplitLeftPercent(next);
  }, []);

  useEffect(() => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem(PRODUCT_SPLIT_STORAGE_KEY, String(Math.round(splitLeftPercent)));
  }, [splitLeftPercent]);

  const activeShopId = workspace?.shop.shopid ?? "";
  const activeBranch = workspace?.branch ? branchDisplayName(workspace.branch) : "-";
  const activeTenant = workspace?.shop ? shopDisplayName(workspace.shop) : "-";
  const shopLanguages = useMemo(() => languageCodesFromWorkspace(workspace), [workspace]);

  const selectedBase = useMemo(
    () => items.find((item) => item.barcode === selectedBarcode) ?? items[0] ?? null,
    [items, selectedBarcode],
  );
  const selected = useMemo(() => {
    if (!selectedBase) return null;
    return detailItems[detailKey(selectedBase)] ?? selectedBase;
  }, [detailItems, selectedBase]);
  const selectedIndex = useMemo(
    () => items.findIndex((item) => item.barcode === selectedBarcode),
    [items, selectedBarcode],
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
  }, []);

  const adjustSplitWithKeyboard = useCallback((event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
    event.preventDefault();
    const direction = event.key === "ArrowLeft" ? -2 : 2;
    setSplitLeftPercent((current) => clampProductSplitLeft(current + direction));
  }, []);

  const activeFilterCount = useMemo(
    () => Object.values(filters).filter((value) => value.trim() !== "").length,
    [filters],
  );

  const loadBarcodes = useCallback(async () => {
    if (!auth || !activeShopId) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    setLoading(true);
    setNotice(null);
    try {
      const payload = {
        backendUrl: auth.backendUrl,
        shopid: activeShopId,
        keyword: search.trim(),
        group_code: filters.groupCode.trim(),
        brand_code: filters.brandCode.trim(),
        categorycode: filters.categoryCode.trim(),
        classcode: filters.classCode.trim(),
        designcode: filters.designCode.trim(),
        gradecode: filters.gradeCode.trim(),
        modelcode: filters.modelCode.trim(),
        patterncode: filters.patternCode.trim(),
        price_min: toNumberOrNull(filters.priceMin),
        price_max: toNumberOrNull(filters.priceMax),
        limit: pageSize,
        offset: 0,
        sort_field: search.trim() ? "relevance" : "barcode",
        sort_order: search.trim() ? "desc" : "asc",
      };

      const response = await fetch("/api/product-barcode/list", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(payload),
      });
      const data = (await response.json()) as BarcodeApiResponse;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text.requestFailed);
      }

      const nextItems = normalizeBarcodeList(data.data);
      setItems(nextItems);
      setTotal(Number(data.total ?? nextItems.length));
      setDetailItems({});
      setSelectedBarcode((current) => {
        if (current && nextItems.some((item) => item.barcode === current)) return current;
        return nextItems[0]?.barcode ?? "";
      });
      setCheckedBarcodes([]);
    } catch (error) {
      setItems([]);
      setTotal(0);
      setSelectedBarcode("");
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text.requestFailed });
    } finally {
      setLoading(false);
    }
  }, [activeShopId, auth, filters, search, text.apiRequired, text.requestFailed]);

  const loadBarcodeDetail = useCallback(
    async (item: ProductBarcodeRecord) => {
      if (!auth || !item.guidFixed) return;
      const key = detailKey(item);
      if (detailItems[key]) return;

      try {
        const response = await fetch(`/api/product-barcode/${encodeURIComponent(item.guidFixed)}`, {
          method: "GET",
          headers: {
            "x-bc-backend-url": auth.backendUrl,
            Authorization: `Bearer ${auth.token}`,
          },
        });
        const data = (await response.json()) as BarcodeApiResponse;
        if (!response.ok || data.success === false) return;

        const detail = extractBarcodeRecordPayload(data.data);
        if (!detail) return;
        const mergedDetail = isRecord(detail) ? { ...item.raw, ...detail } : detail;
        const normalized = normalizeBarcodeRecord(mergedDetail);
        setDetailItems((current) => ({ ...current, [key]: normalized }));
      } catch {
        // Detail loading is best-effort; the list row still has enough data to remain usable.
      }
    },
    [auth, detailItems],
  );

  useEffect(() => {
    if (!auth || !activeShopId) return;
    const timer = window.setTimeout(() => {
      void loadBarcodes();
    }, 300);
    return () => window.clearTimeout(timer);
  }, [activeShopId, auth, loadBarcodes]);

  useEffect(() => {
    if (!selectedBase) return;
    void loadBarcodeDetail(selectedBase);
  }, [loadBarcodeDetail, selectedBase]);

  function toggleChecked(barcode: string) {
    setCheckedBarcodes((current) =>
      current.includes(barcode) ? current.filter((item) => item !== barcode) : [...current, barcode],
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
      description: editorBarcode.barcode ? `${text.barcode}: ${editorBarcode.barcode}` : undefined,
      details: lang === "th" ? "การเปลี่ยนแปลงในฟอร์มนี้ยังไม่ได้บันทึก ถ้ายืนยัน ระบบจะปิดฟอร์มและทิ้งข้อมูลที่แก้ไขอยู่" : "Unsaved form changes will be discarded.",
      confirmLabel: lang === "th" ? "ทิ้งการแก้ไข" : "Discard",
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

  function handleEditorChange(next: ProductBarcodeObject | ((current: ProductBarcodeObject) => ProductBarcodeObject)) {
    setEditorDirty(true);
    setEditorBarcode(next);
  }

  async function selectListItem(item: ProductBarcodeRecord) {
    if (!(await canDiscardEditor())) return;
    setEditorOpen(false);
    setEditorDirty(false);
    setSelectedBarcode(item.barcode);
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
    next.shopid = activeShopId;
    setEditorBarcode(next);
    setEditorOpen(true);
  }

  async function openEditEditor(itemOverride?: ProductBarcodeRecord) {
    if (!(await canDiscardEditor())) return;
    const target = itemOverride ?? selected;
    if (!target) return;
    setEditorMode("edit");
    setSelectedBarcode(target.barcode);
    setEditorGuid(target.guidFixed);
    setEditorDirty(false);
    const base = emptyProductBarcode();
    base.shopid = activeShopId;
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
    base.shopid = activeShopId;
    const next = rawToProductBarcode(selected.raw, base);
    next.guidfixed = "";
    next.barcode = "";
    next.shopid = activeShopId;
    setEditorBarcode(next);
    setEditorOpen(true);
  }

  function copyCurrentEditorValue() {
    const next = JSON.parse(JSON.stringify(editorBarcode)) as ProductBarcodeObject;
    next.guidfixed = "";
    next.barcode = "";
    next.shopid = activeShopId;
    setEditorMode("create");
    setEditorGuid("");
    setEditorBarcode(next);
    setEditorDirty(true);
  }

  async function saveEditor(value: ProductBarcodeObject) {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    const payload: Record<string, unknown> = { ...(value as unknown as Record<string, unknown>) };
    const guid = editorGuid || String(value.guidfixed ?? "");
    const url = editorMode === "edit" ? `/api/product-barcode/${encodeURIComponent(guid)}` : "/api/product-barcode";
    if (editorMode === "edit" && !guid) {
      setNotice({ type: "error", text: text.missingGuid });
      return;
    }

    setEditorSaving(true);
    setNotice(null);
    try {
      const response = await fetch(url, {
        method: editorMode === "edit" ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({ backendUrl: auth.backendUrl, data: payload }),
      });
      const data = (await response.json()) as BarcodeApiResponse;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text.requestFailed);
      }

      setEditorOpen(false);
      setEditorGuid("");
      setEditorDirty(false);
      await loadBarcodes();
      setNotice({ type: "success", text: text.saveSuccess });
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text.requestFailed });
    } finally {
      setEditorSaving(false);
    }
  }

  async function deleteSelected() {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    const selectedItems = items.filter((item) => checkedBarcodes.includes(item.barcode));
    const guids = selectedItems.map((item) => item.guidFixed.trim()).filter(Boolean);
    if (guids.length === 0) {
      setNotice({ type: "error", text: text.missingGuid });
      return;
    }

    const confirmed = await confirm({
      title: text.deleteConfirm,
      description: `${text.selected}: ${guids.length.toLocaleString(lang === "th" ? "th-TH" : "en-US")}`,
      details: selectedItems.slice(0, 8).map((item) => item.barcode).filter(Boolean).join(", "),
      confirmLabel: text.delete,
      cancelLabel: text.close,
      tone: "danger",
    });
    if (!confirmed) return;

    setLoading(true);
    setNotice(null);
    try {
      const response = await fetch("/api/product-barcode", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({ backendUrl: auth.backendUrl, guids }),
      });
      const data = (await response.json()) as BarcodeApiResponse;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text.requestFailed);
      }

      setSelectMode(false);
      setCheckedBarcodes([]);
      await loadBarcodes();
      setNotice({ type: "success", text: text.deleteSuccess });
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text.requestFailed });
    } finally {
      setLoading(false);
    }
  }

  async function deleteCurrentItem() {
    if (!auth) {
      setNotice({ type: "error", text: text.apiRequired });
      return;
    }

    const guid = (selected?.guidFixed || editorGuid || String(editorBarcode.guidfixed ?? "")).trim();
    if (!guid) {
      setNotice({ type: "error", text: text.missingGuid });
      return;
    }

    const barcode = selected?.barcode ?? editorBarcode.barcode ?? "";
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
      const response = await fetch(`/api/product-barcode/${encodeURIComponent(guid)}`, {
        method: "DELETE",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
      });
      const data = (await response.json()) as BarcodeApiResponse;
      if (!response.ok || data.success === false) {
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
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text.requestFailed });
    } finally {
      setLoading(false);
    }
  }

  function exportCsv() {
    if (items.length === 0) return;
    const rows = [
      [
        text.barcode,
        text.productName,
        text.unit,
        text.itemCode,
        text.groupCode,
        text.brandCode,
        text.categoryCode,
        text.classCode,
        text.designCode,
        text.gradeCode,
        text.modelCode,
        text.patternCode,
        text.subGroup1,
        text.subGroup2,
        text.shelf,
        text.productType,
        text.foodType,
        text.balance,
        text.retailPrice,
        text.standValue,
        text.divideValue,
        text.itemType,
        text.isUseSubBarcodes,
      ],
      ...items.map((item) => [
        item.barcode,
        item.name,
        item.unitName,
        item.itemCode,
        item.groupCode,
        item.brandCode,
        item.categoryCode,
        item.classCode,
        item.designCode,
        item.gradeCode,
        item.modelCode,
        item.patternCode,
        item.groupSubOneCode,
        item.groupSubTwoCode,
        formatCodeName(item.shelfCode, item.shelfName),
        String(item.productType || ""),
        String(item.foodType || ""),
        item.balanceFormatted || String(item.balanceQty),
        formatMoney(item.price),
        formatNumber(item.standValue),
        formatNumber(item.divideValue),
        String(item.itemType || ""),
        formatBoolean(item.isUseSubBarcodes, text),
      ]),
    ];
    const csv = rows.map((row) => row.map(csvCell).join(",")).join("\n");
    const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "product-barcodes.csv";
    link.click();
    URL.revokeObjectURL(url);
    setNotice({ type: "success", text: text.exportSuccess });
  }

  const content = (
    <div
      className={cn(
        embedded ? "flex h-full min-h-0 w-full min-w-0 flex-col gap-3 overflow-hidden" : "grid w-full min-w-0 gap-3",
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
              <p className="text-xs font-semibold uppercase text-muted-foreground">BC AI ACCOUNT</p>
              <h1 className="truncate text-3xl font-semibold leading-tight">{text.title}</h1>
              <p className="text-sm text-muted-foreground">{text.subtitle}</p>
            </div>
          </div>
          <div className="flex flex-wrap justify-end gap-2">
            <Button variant="outline" size="sm" onClick={() => void loadBarcodes()} disabled={loading}>
              <RefreshCcw size={16} />
              {text.refresh}
            </Button>
            <Button variant="outline" size="sm" onClick={exportCsv} disabled={items.length === 0}>
              <Download size={16} />
              {text.export}
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
            <Button variant="outline" size="sm" onClick={() => void deleteSelected()} disabled={!selectMode || checkedBarcodes.length === 0}>
              <Trash2 size={16} />
              {checkedBarcodes.length || ""}
            </Button>
            <Button size="sm" onClick={() => void openCreateEditor()}>
              <Plus size={16} />
              {text.add}
            </Button>
          </div>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          <Badge variant="secondary">{text.tenant}: {activeTenant}</Badge>
          <Badge variant="secondary">{text.branch}: {activeBranch}</Badge>
          <Badge variant="outline">{text.total}: {total.toLocaleString("th-TH")}</Badge>
          {selectMode ? <Badge variant="warning">{text.selected}: {checkedBarcodes.length.toLocaleString("th-TH")}</Badge> : null}
        </div>
      </header>
      ) : null}

      {notice ? (
        <div
          className={cn(
            "flex items-center gap-2 rounded-2xl border p-3 text-sm font-medium",
            notice.type === "error" && "border-destructive/30 bg-destructive/10 text-destructive",
            notice.type === "success" && "border-emerald-400/30 bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300",
            notice.type === "info" && "border-primary/25 bg-primary/10 text-primary",
          )}
        >
          <AlertCircle size={18} />
          {notice.text}
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
          className="min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col"
          onKeyDown={handleListKeyDown}
          tabIndex={0}
        >
          <CardHeader className="shrink-0 border-b border-border p-3">
            <form className="flex flex-wrap gap-2" onSubmit={submitSearch}>
              <div className="relative min-w-[240px] flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  className="pl-9"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder={text.search}
                />
              </div>
              <Button variant={filterOpen ? "secondary" : "outline"} type="button" onClick={() => setFilterOpen((current) => !current)}>
                <Filter size={16} />
                {text.filter}
                {activeFilterCount > 0 ? <Badge variant="warning">{activeFilterCount}</Badge> : null}
              </Button>
              <Button variant="outline" type="button" onClick={() => setShowImage((current) => !current)}>
                {showImage ? <ImageOff size={16} /> : <ImageIcon size={16} />}
                {text.image}
              </Button>
              {embedded ? (
                <>
                  <Button variant="outline" size="sm" type="button" onClick={() => void loadBarcodes()} disabled={loading}>
                    <RefreshCcw size={16} />
                    {text.refresh}
                  </Button>
                  <Button variant="outline" size="sm" type="button" onClick={exportCsv} disabled={items.length === 0}>
                    <Download size={16} />
                    {text.export}
                  </Button>
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
                  <Button variant="outline" size="sm" type="button" onClick={() => void deleteSelected()} disabled={!selectMode || checkedBarcodes.length === 0}>
                    <Trash2 size={16} />
                    {checkedBarcodes.length || ""}
                  </Button>
                  <Button size="sm" type="button" onClick={() => void openCreateEditor()}>
                    <Plus size={16} />
                    {text.add}
                  </Button>
                </>
              ) : null}
            </form>
            {embedded ? (
              <div className="mt-2 flex flex-wrap gap-2">
                <Badge variant="secondary">{text.tenant}: {activeTenant}</Badge>
                <Badge variant="secondary">{text.branch}: {activeBranch}</Badge>
                <Badge variant="outline">{text.total}: {total.toLocaleString("th-TH")}</Badge>
                {selectMode ? <Badge variant="warning">{text.selected}: {checkedBarcodes.length.toLocaleString("th-TH")}</Badge> : null}
              </div>
            ) : null}
            {filterOpen ? (
              <div className="mt-3 grid gap-2 rounded-2xl border border-border bg-muted/30 p-3 md:grid-cols-4 xl:grid-cols-5">
                <Input value={filters.groupCode} onChange={(event) => setFilters((current) => ({ ...current, groupCode: event.target.value }))} placeholder={text.groupCode} />
                <Input value={filters.brandCode} onChange={(event) => setFilters((current) => ({ ...current, brandCode: event.target.value }))} placeholder={text.brandCode} />
                <Input value={filters.categoryCode} onChange={(event) => setFilters((current) => ({ ...current, categoryCode: event.target.value }))} placeholder={text.categoryCode} />
                <Input value={filters.classCode} onChange={(event) => setFilters((current) => ({ ...current, classCode: event.target.value }))} placeholder={text.classCode} />
                <Input value={filters.designCode} onChange={(event) => setFilters((current) => ({ ...current, designCode: event.target.value }))} placeholder={text.designCode} />
                <Input value={filters.gradeCode} onChange={(event) => setFilters((current) => ({ ...current, gradeCode: event.target.value }))} placeholder={text.gradeCode} />
                <Input value={filters.modelCode} onChange={(event) => setFilters((current) => ({ ...current, modelCode: event.target.value }))} placeholder={text.modelCode} />
                <Input value={filters.patternCode} onChange={(event) => setFilters((current) => ({ ...current, patternCode: event.target.value }))} placeholder={text.patternCode} />
                <Input value={filters.priceMin} onChange={(event) => setFilters((current) => ({ ...current, priceMin: event.target.value }))} placeholder={text.priceMin} inputMode="decimal" />
                <div className="flex gap-2">
                  <Input value={filters.priceMax} onChange={(event) => setFilters((current) => ({ ...current, priceMax: event.target.value }))} placeholder={text.priceMax} inputMode="decimal" />
                  <Button variant="outline" type="button" onClick={clearFilters} aria-label={text.clearFilter}>
                    <X size={16} />
                  </Button>
                </div>
              </div>
            ) : null}
          </CardHeader>
          <CardContent className="grid min-h-[420px] p-0 xl:min-h-0 xl:flex-1 xl:grid-rows-[auto_minmax(0,1fr)]">
            <div className="hidden grid-cols-[1.2fr_2fr_0.9fr_1fr_0.9fr_0.9fr_auto] gap-2 border-b border-border bg-muted/70 px-3 py-2 text-xs font-semibold text-foreground lg:grid">
              <span>{text.barcode}</span>
              <span>{text.productName}</span>
              <span>{text.unit}</span>
              <span>{text.itemCode}</span>
              <span>{text.balance}</span>
              <span className="text-right">{text.retailPrice}</span>
              <span className="text-center">{showImage || selectMode ? text.image : ""}</span>
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
                  {items.map((item, index) => (
                    <BarcodeRow
                      checked={checkedBarcodes.includes(item.barcode)}
                      imageUrl={resolveImageUrl(item.imageUri, auth?.backendUrl)}
                      index={index}
                      item={item}
                      key={item.guidFixed || item.barcode}
                      onEdit={() => void openEditEditor(item)}
                      onSelect={() => void selectListItem(item)}
                      onToggleChecked={() => toggleChecked(item.barcode)}
                      selected={selected?.barcode === item.barcode}
                      selectMode={selectMode}
                      showImage={showImage}
                      text={text}
                    />
                  ))}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

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

        {editorOpen ? (
          <ProductBarcodeFormDialog
            open={editorOpen}
            mode={editorMode}
            value={editorBarcode}
            onChange={handleEditorChange}
            onSave={(val) => void saveEditor(val)}
            onCancel={() => void closeEditor()}
            saving={editorSaving}
            language={lang}
            shopLanguages={shopLanguages}
            auth={auth}
            extraActions={
              editorMode === "edit" ? (
                <>
                  <Button type="button" variant="ghost" size="sm" onClick={copyCurrentEditorValue} aria-label={text.copy}>
                    <Copy className="h-4 w-4" />
                  </Button>
                  <Button type="button" variant="ghost" size="sm" onClick={() => void deleteCurrentItem()} aria-label={text.deleteConfirm}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </>
              ) : null
            }
            embedded
          />
        ) : (
          <ProductBarcodeDetail item={selected} onCopy={() => void openCopyEditor()} onDelete={() => void deleteCurrentItem()} onEdit={() => void openEditEditor()} text={text} />
        )}
      </div>
      {selected && !editorOpen ? (
        <BarcodeQuickActions
          barcode={selected.barcode}
          guid={selected.guidFixed}
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
  checked,
  imageUrl,
  index,
  item,
  onEdit,
  onSelect,
  onToggleChecked,
  selected,
  selectMode,
  showImage,
  text,
}: {
  checked: boolean;
  imageUrl: string;
  index: number;
  item: ProductBarcodeRecord;
  onEdit: () => void;
  onSelect: () => void;
  onToggleChecked: () => void;
  selected: boolean;
  selectMode: boolean;
  showImage: boolean;
  text: BarcodeText;
}) {
  return (
    <button
      className={cn(
        "grid w-full min-w-0 gap-2 border-b border-border px-3 py-2 text-left text-sm transition hover:bg-primary/5 lg:grid-cols-[1.2fr_2fr_0.9fr_1fr_0.9fr_0.9fr_auto]",
        index % 2 === 0 ? "bg-background" : "bg-muted/20",
        selected && "bg-primary/10 font-semibold",
      )}
      onClick={selectMode ? onToggleChecked : onSelect}
      onDoubleClick={selectMode ? undefined : onEdit}
      type="button"
    >
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">{text.barcode}</span>
        <div className="truncate">{item.barcode || "-"}</div>
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">{text.productName}</span>
        <div className="line-clamp-2">{item.name || "-"}</div>
        {item.groupName ? <div className="truncate text-xs text-muted-foreground">{item.groupName}</div> : null}
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">{text.unit}</span>
        <div className={cn("truncate", item.unitCount > 1 && "text-primary")}>
          {item.unitName || item.unitCode || "-"}
        </div>
        {item.unitCount > 1 ? <div className="truncate text-xs text-muted-foreground">{item.allUnitNames}</div> : null}
      </div>
      <div className="min-w-0">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">{text.itemCode}</span>
        <div className="truncate">{item.itemCode || "-"}</div>
      </div>
      <div>
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">{text.balance}</span>
        <div className={cn(item.balanceQty <= 0 && "text-destructive")}>{item.balanceFormatted || formatNumber(item.balanceQty)}</div>
      </div>
      <div className="lg:text-right">
        <span className="lg:hidden text-xs font-semibold text-muted-foreground">{text.retailPrice}</span>
        <div>{formatMoney(item.price)}</div>
      </div>
      <div className="flex items-center justify-start gap-2 lg:justify-center">
        {showImage ? (
          imageUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img alt="" className="size-8 rounded-lg border border-border object-cover" src={imageUrl} />
          ) : (
            <span className="grid size-8 place-items-center rounded-lg border border-border text-muted-foreground">
              <ImageOff size={14} />
            </span>
          )
        ) : null}
        {selectMode ? (
          <span className={cn("grid size-6 place-items-center rounded-md border", checked && "border-primary bg-primary text-primary-foreground")}>
            {checked ? <CheckSquare size={14} /> : null}
          </span>
        ) : null}
      </div>
    </button>
  );
}

function ProductBarcodeDetail({
  item,
  onCopy,
  onDelete,
  onEdit,
  text,
}: {
  item: ProductBarcodeRecord | null;
  onCopy: () => void;
  onDelete: () => void;
  onEdit: () => void;
  text: BarcodeText;
}) {
  const basicFields = item
    ? [
        { label: text.guid, value: item.guidFixed },
        { label: text.shopId, value: item.shopId },
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
        { label: text.groupCode, value: formatCodeName(item.groupCode, item.groupName) },
        { label: text.brandCode, value: formatCodeName(item.brandCode, item.brandName) },
        { label: text.categoryCode, value: formatCodeName(item.categoryCode, item.categoryName) },
        { label: text.classCode, value: formatCodeName(item.classCode, item.className) },
        { label: text.designCode, value: formatCodeName(item.designCode, item.designName) },
        { label: text.gradeCode, value: formatCodeName(item.gradeCode, item.gradeName) },
        { label: text.modelCode, value: formatCodeName(item.modelCode, item.modelName) },
        { label: text.patternCode, value: formatCodeName(item.patternCode, item.patternName) },
        { label: text.subGroup1, value: formatCodeName(item.groupSubOneCode, item.groupSubOneName) },
        { label: text.subGroup2, value: formatCodeName(item.groupSubTwoCode, item.groupSubTwoName) },
        { label: text.manufacturer, value: formatCodeName(item.manufacturerCode, item.manufacturerName) },
        { label: text.shelf, value: formatCodeName(item.shelfCode, item.shelfName) },
      ]
    : [];
  const stockFields = item
    ? [
        { label: text.unit, value: formatCodeName(item.unitCode, item.unitName) },
        { label: text.multiUnit, value: item.allUnitNames },
        { label: text.balance, value: item.balanceFormatted || formatNumber(item.balanceQty) },
        { label: text.retailPrice, value: formatMoney(item.price) },
        { label: text.standValue, value: formatNumber(item.standValue) },
        { label: text.divideValue, value: formatNumber(item.divideValue) },
        { label: text.unitCount, value: String(item.unitCount || "") },
        { label: text.color, value: item.colorSelectHex || item.colorSelect },
      ]
    : [];
  const statusFields = item
    ? [
        { label: text.itemType, value: String(item.itemType || "") },
        { label: text.productType, value: String(item.productType || "") },
        { label: text.foodType, value: String(item.foodType || "") },
        { label: text.materialType, value: String(item.materialType || "") },
        { label: text.isStock, value: String(item.isStock || "") },
        { label: text.isMainBarcode, value: formatBoolean(item.isMainBarcode, text) },
        { label: text.isMainItem, value: formatBoolean(item.isMainItem, text) },
        { label: text.isUseSubBarcodes, value: formatBoolean(item.isUseSubBarcodes, text) },
        { label: text.useImageOrColor, value: formatBoolean(item.useImageOrColor, text) },
        { label: text.condition, value: formatBoolean(item.condition, text) },
        { label: text.isSumPoint, value: formatBoolean(item.isSumPoint, text) },
        { label: text.isDividend, value: formatBoolean(item.isDividend, text) },
        { label: text.isALaCarte, value: formatBoolean(item.isALaCarte, text) },
        { label: text.isSplitUnitPrint, value: formatBoolean(item.isSplitUnitPrint, text) },
        { label: text.isOnlyStaff, value: formatBoolean(item.isOnlyStaff, text) },
        { label: text.isStockForRestaurant, value: formatBoolean(item.isStockForRestaurant, text) },
        { label: text.isDiscountPointOfPurchase, value: formatBoolean(item.isDiscountPointOfPurchase, text) },
        { label: text.isAlert, value: formatBoolean(item.isAlert, text) },
        { label: text.isDisable, value: formatBoolean(item.isDisable, text) },
        { label: text.showIsDividend, value: item.showIsDividend },
        { label: text.rowNumber, value: item.rowNumber ? String(item.rowNumber) : "" },
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
        { label: text.businessTypes, value: formatCount(item.businessTypeCount) },
        { label: text.ignoreBranches, value: formatCount(item.ignoreBranchCount) },
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
            <Button disabled={!item} onClick={onCopy} size="sm" variant="outline">
              <Copy size={16} />
              {text.copy}
            </Button>
            <Button disabled={!item} onClick={onDelete} size="sm" variant="outline">
              <Trash2 size={16} />
              {text.delete}
            </Button>
            <Button disabled={!item} onClick={onEdit} size="sm" variant="outline">
              <Pencil size={16} />
              {text.edit}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 overflow-auto p-3 xl:min-h-0 xl:flex-1">
        {!item ? (
          <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">{text.noSelection}</div>
        ) : (
          <>
            <div className="rounded-2xl border border-border bg-muted/30 p-3">
              <p className="text-xs font-semibold text-muted-foreground">{text.barcode}</p>
              <p className="break-all text-2xl font-semibold">{item.barcode}</p>
              <p className="mt-1 text-sm text-muted-foreground">{item.name}</p>
            </div>
            <DetailSection fields={basicFields} title={text.basicInfo} />
            <DetailSection fields={classificationFields} title={text.classification} />
            <DetailSection fields={stockFields} title={text.stockAndUnit} />
            <DetailSection fields={statusFields} title={text.taxAndFlags} />
            <DetailSection fields={relationFields} title={text.relations} />
            <details className="rounded-2xl border border-border p-3">
              <summary className="cursor-pointer text-sm font-semibold">{text.rawFields}</summary>
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
  text,
}: {
  barcode: string;
  guid: string;
  text: BarcodeText;
}) {
  if (!barcode) return null;
  const encoded = encodeURIComponent(barcode);
  return (
    <div className="fixed bottom-3 right-3 z-30 flex gap-2 rounded-full border border-border bg-card/95 p-1 shadow-lg backdrop-blur md:bottom-6 md:right-6">
      <Button asChild size="sm" variant="ghost" title={text.labelPrint}>
        <a href={`/product_barcode_shelf?barcode=${encoded}`} target="_blank" rel="noreferrer">
          <Printer size={16} />
          <span className="hidden md:inline">{text.labelPrint}</span>
        </a>
      </Button>
      <Button asChild size="sm" variant="ghost" title={text.priceHistory ?? "Price history"}>
        <a href={`/product-price-history?barcode=${encoded}`} target="_blank" rel="noreferrer">
          <History size={16} />
          <span className="hidden md:inline">{text.priceHistory ?? "Price history"}</span>
        </a>
      </Button>
      {guid ? (
        <Button asChild size="sm" variant="ghost" title={text.bomView ?? "BOM"}>
          <a href={`/api/product-barcode/bom/${encoded}`} target="_blank" rel="noreferrer">
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
    return workspace.shop?.shopid ? workspace : null;
  } catch {
    return null;
  }
}

function languageCodesFromWorkspace(workspace: WorkspaceSession | null): string[] {
  const shopInfo = isRecord(workspace?.shopInfo) ? workspace.shopInfo : {};
  const settings = isRecord(shopInfo.settings) ? shopInfo.settings : {};
  const rows = Array.isArray(settings.languageconfigs)
    ? settings.languageconfigs
        .filter(isRecord)
        .map((item) => ({
          code: supportedLanguageCode(item.code, ""),
          isUse: item.is_use === undefined && item.isuse === undefined ? true : booleanFromUnknown(item.is_use ?? item.isuse),
          isDefault: booleanFromUnknown(item.isdefault),
        }))
        .filter((item) => item.code && item.isUse)
    : [];
  const configuredDefault = supportedLanguageCode(settings.language, "");
  const primary = rows.find((item) => item.isDefault)?.code || configuredDefault || rows[0]?.code || "th";
  const ordered = [primary, ...rows.map((item) => item.code).filter((code) => code && code !== primary)];
  return Array.from(new Set(ordered));
}

function supportedLanguageCode(value: unknown, fallback: string): string {
  const raw = typeof value === "string" ? value.trim().toLowerCase() : "";
  if (!raw) return fallback;
  const normalized = normalizeLanguage(raw);
  return LANGUAGES.some((item) => item.code === normalized) ? normalized : fallback;
}

function booleanFromUnknown(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    return normalized === "true" || normalized === "1" || normalized === "yes" || normalized === "y";
  }
  return false;
}

function normalizeBarcodeList(value: unknown): ProductBarcodeRecord[] {
  if (!Array.isArray(value)) return [];
  return value.map(normalizeBarcodeRecord).filter((item) => item.barcode || item.itemCode);
}

function normalizeBarcodeRecord(value: unknown): ProductBarcodeRecord {
  const record = isRecord(value) ? value : {};
  return {
    raw: record,
    guidFixed: getFirstString(record, ["guid_fixed", "guidfixed"]),
    shopId: getFirstString(record, ["shopid", "shop_id"]),
    barcode: getFirstString(record, ["barcode"]),
    barcodeRef: getFirstString(record, ["barcoderef", "barcode_ref", "refbarcode"]),
    name: localizedNameFromKeys(record, ["names"], getFirstString(record, ["name0", "name", "item_name"])),
    unitName: localizedNameFromKeys(record, ["itemunitnames", "unit_names"], getFirstString(record, ["unit_name", "unitname"])),
    unitCode: getFirstString(record, ["item_unit_code", "itemunitcode", "unitcode"]),
    itemCode: getFirstString(record, ["itemcode", "item_code"]),
    itemGuid: getFirstString(record, ["itemguid", "item_guid"]),
    itemGuidFixed: getFirstString(record, ["itemguidfixed", "item_guid_fixed"]),
    parentGuid: getFirstString(record, ["parentguid", "parent_guid"]),
    groupName: localizedNameFromKeys(record, ["group_names", "groupnames"], getFirstString(record, ["group_name", "groupname"])),
    groupCode: getFirstString(record, ["group_code", "groupcode"]),
    brandName: localizedNameFromKeys(record, ["brand_names", "brandnames"], getFirstString(record, ["brand_name", "brandname"])),
    brandCode: getFirstString(record, ["brand_code", "brandcode"]),
    categoryName: localizedNameFromKeys(record, ["category_names", "categorynames"], getFirstString(record, ["category_name", "categoryname"])),
    categoryCode: getFirstString(record, ["category_code", "categorycode"]),
    className: localizedNameFromKeys(record, ["class_names", "classnames"], getFirstString(record, ["class_name", "classname"])),
    classCode: getFirstString(record, ["class_code", "classcode"]),
    designName: localizedNameFromKeys(record, ["design_names", "designnames"], getFirstString(record, ["design_name", "designname"])),
    designCode: getFirstString(record, ["design_code", "designcode"]),
    gradeName: localizedNameFromKeys(record, ["grade_names", "gradenames"], getFirstString(record, ["grade_name", "gradename"])),
    gradeCode: getFirstString(record, ["grade_code", "gradecode"]),
    modelName: localizedNameFromKeys(record, ["model_names", "modelnames"], getFirstString(record, ["model_name", "modelname"])),
    modelCode: getFirstString(record, ["model_code", "modelcode"]),
    patternName: localizedNameFromKeys(record, ["pattern_names", "patternnames"], getFirstString(record, ["pattern_name", "patternname"])),
    patternCode: getFirstString(record, ["pattern_code", "patterncode"]),
    groupSubOneName: localizedNameFromKeys(record, ["groupsubonenames", "group_sub_one_names"], getFirstString(record, ["groupsubonename", "group_sub_one_name"])),
    groupSubOneCode: getFirstString(record, ["groupsubonecode", "group_sub_one_code"]),
    groupSubTwoName: localizedNameFromKeys(record, ["groupsubtwonames", "group_sub_two_names"], getFirstString(record, ["groupsubtwoname", "group_sub_two_name"])),
    groupSubTwoCode: getFirstString(record, ["groupsubtwocode", "group_sub_two_code"]),
    manufacturerName: localizedNameFromKeys(record, ["manufacturernames", "manufacturer_names"], getFirstString(record, ["manufacturername", "manufacturer_name"])),
    manufacturerCode: getFirstString(record, ["manufacturercode", "manufacturer_code"]),
    shelfCode: getFirstString(record, ["shelfcode", "shelf_code", "shelfCode"]),
    shelfName: getFirstString(record, ["shelfname", "shelf_name", "shelfName"]),
    checksum: getFirstString(record, ["checksum"]),
    price: getPrice(record),
    imageUri: getFirstString(record, ["imageuri", "image_uri"]),
    colorSelect: getFirstString(record, ["colorselect", "color_select"]),
    colorSelectHex: getFirstString(record, ["colorselecthex", "color_select_hex"]),
    unitCount: getFirstNumber(record, ["unit_count", "unitcount"]),
    allUnitNames: getFirstString(record, ["all_unit_names", "allunitnames"]),
    balanceQty: getFirstNumber(record, ["balance_qty", "balanceqty"]),
    balanceFormatted: getFirstString(record, ["balance_formatted", "balanceformatted"]),
    standValue: getFirstNumber(record, ["standvalue", "stand_value", "barcoderefunitstand", "barcode_ref_unit_stand"]),
    divideValue: getFirstNumber(record, ["dividevalue", "divide_value", "barcoderefunitdivide", "barcode_ref_unit_divide"]),
    itemType: getFirstNumber(record, ["itemtype", "item_type"]),
    productType: getFirstNumber(record, ["producttype", "product_type"]),
    foodType: getFirstNumber(record, ["foodtype", "food_type"]),
    materialType: getFirstNumber(record, ["materialtype", "material_type"]),
    taxType: getFirstNumber(record, ["taxtype", "tax_type"]),
    vatType: getFirstNumber(record, ["vattype", "vat_type"]),
    vatCal: getFirstNumber(record, ["vatcal", "vat_cal"]),
    isStock: getFirstNumber(record, ["isstock", "is_stock"]),
    isMainBarcode: getFirstBoolean(record, ["ismainbarcode", "is_main_barcode"]),
    isMainItem: getFirstBoolean(record, ["ismainitem", "is_main_item"]),
    isUseSubBarcodes: getFirstBoolean(record, ["isusesubbarcodes", "is_use_sub_barcodes"]),
    useImageOrColor: getFirstBoolean(record, ["useimageorcolor", "use_image_or_color"]),
    condition: getFirstBoolean(record, ["condition"]),
    isSumPoint: getFirstBoolean(record, ["issumpoint", "is_sum_point"]),
    isDividend: getFirstBoolean(record, ["isdividend", "is_dividend"]),
    isALaCarte: getFirstBoolean(record, ["isalacarte", "is_a_la_carte"]),
    isSplitUnitPrint: getFirstBoolean(record, ["issplitunitprint", "is_split_unit_print"]),
    isOnlyStaff: getFirstBoolean(record, ["isonlystaff", "is_only_staff"]),
    isStockForRestaurant: getFirstBoolean(record, ["isstockforrestaurant", "is_stock_for_restaurant"]),
    isDiscountPointOfPurchase: getFirstBoolean(record, ["isdiscountpointofpurchase", "is_discount_point_of_purchase"]),
    isAlert: getFirstBoolean(record, ["isalert", "is_alert"]),
    isDisable: getFirstBoolean(record, ["isdisable", "is_disable"]),
    showIsDividend: getFirstString(record, ["showisdividend", "show_is_dividend"]),
    rowNumber: getFirstNumber(record, ["rownumber", "row_number"]),
    maxDiscount: getFirstString(record, ["maxdiscount", "max_discount"]),
    discount: getFirstString(record, ["discount"]),
    description: getFirstString(record, ["description"]),
    alertDescription: getFirstString(record, ["alertdescription", "alert_description"]),
    priceCount: getArrayCount(record, ["prices"]),
    refBarcodeCount: getArrayCount(record, ["refbarcodes", "ref_barcodes"]),
    subBarcodeCount: getArrayCount(record, ["barcodes", "subbarcodes", "sub_barcodes"]),
    bomCount: getArrayCount(record, ["bom"]),
    optionCount: getArrayCount(record, ["options"]),
    orderTypeCount: getArrayCount(record, ["ordertypes", "order_types"]),
    dimensionCount: getArrayCount(record, ["dimensions"]),
    businessTypeCount: getArrayCount(record, ["businesstypes", "business_types"]),
    ignoreBranchCount: getArrayCount(record, ["ignorebranches", "ignore_branches"]),
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
    return getNumber(price, "key_number") === 1 || getNumber(price, "keynumber") === 1;
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
  return item.guidFixed || item.barcode || item.itemCode;
}

function getFirstString(record: Record<string, unknown>, keys: string[]): string {
  for (const key of keys) {
    const value = getString(record, key);
    if (value) return value;
  }
  return "";
}

function getFirstNumber(record: Record<string, unknown>, keys: string[]): number {
  for (const key of keys) {
    if (!Object.prototype.hasOwnProperty.call(record, key)) continue;
    return getNumber(record, key);
  }
  return 0;
}

function getFirstBoolean(record: Record<string, unknown>, keys: string[]): boolean {
  for (const key of keys) {
    if (!Object.prototype.hasOwnProperty.call(record, key)) continue;
    const value = record[key];
    if (typeof value === "boolean") return value;
    if (typeof value === "number") return value !== 0;
    if (typeof value === "string") {
      const normalized = value.trim().toLowerCase();
      if (!normalized) return false;
      return normalized === "true" || normalized === "1" || normalized === "yes" || normalized === "y";
    }
  }
  return false;
}

function localizedNameFromKeys(record: Record<string, unknown>, keys: string[], fallback = ""): string {
  for (const key of keys) {
    const names = getNames(record, key);
    const name = localizedName(names, "");
    if (name) return name;
    const raw = getString(record, key);
    if (raw) return raw;
  }
  return fallback;
}

function getArrayCount(record: Record<string, unknown>, keys: string[]): number {
  for (const key of keys) {
    const value = record[key];
    if (Array.isArray(value)) return value.length;
  }
  return 0;
}

function getNames(record: Record<string, unknown>, key: string): LocalizedName[] | undefined {
  const value = record[key];
  return Array.isArray(value)
    ? value.filter(isRecord).map((item) => ({ code: getString(item, "code"), name: getString(item, "name") }))
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

function toNumberOrNull(value: string): number | null {
  const trimmed = value.trim();
  if (!trimmed) return null;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : null;
}

function formatMoney(value: number): string {
  return new Intl.NumberFormat("th-TH", { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value || 0);
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat("th-TH", { maximumFractionDigits: 4 }).format(value || 0);
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

function csvCell(value: string): string {
  return `"${String(value ?? "").replace(/"/g, '""')}"`;
}

function resolveImageUrl(imageUri: string, backendUrl: string | undefined): string {
  if (!imageUri) return "";
  if (/^https?:\/\//i.test(imageUri)) return imageUri;
  if (!backendUrl) return imageUri;
  try {
    return new URL(imageUri.replace(/^\/+/, ""), `${deriveMainApiUrl(backendUrl)}/`).toString();
  } catch {
    return imageUri;
  }
}

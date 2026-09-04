"use client";

import { useState, useEffect, useMemo, useRef, useCallback, type FormEvent } from "react";
import {
  FolderOpen,
  CheckSquare,
  Filter,
  Loader2,
  Package,
  Pencil,
  Plus,
  Copy,
  RefreshCw,
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
  Sparkles,
  Info,
  DollarSign,
  Layers,
  Settings2,
  CheckCircle2,
  AlertCircle,
  Eye,
  Percent,
  TrendingUp,
  Box,
  Scale,
  Play,
  ShoppingBag,
  HelpCircle,
  ArrowRight,
  Check
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";

import { MasterPicker } from "@/components/product-barcode/master-picker";
import { listBarcodes, type MasterEntry } from "@/lib/product-barcode/api";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { getBarcodeText } from "@/lib/product-barcode/language";
import {
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
  WORKSPACE_CHANGED_EVENT
} from "@/lib/workspace-models";
import { pickName, rawToProduct } from "@/lib/product-barcode/utils";
import { NamesEditor, languageCodesFromWorkspace } from "@/components/product-barcode/names-editor";
import {
  type Product,
  type NameX,
  type ProductImage,
  type ProductOption,
  type ProductChoice,
  type ProductBarcodeListRow
} from "@/lib/product-barcode/types";
import { cn, randomId } from "@/lib/utils";
import { pushNotice } from "@/lib/toast";
import { normalizeBusinessCode } from "@/lib/business-code";
import { authFetch, getAuthSession } from "@/lib/client-auth-session";

type ProductSetScreenProps = {
  active?: boolean;
  embedded?: boolean;
  language?: LanguageCode;
};

function productSetRowKey(item: Product, index: number): string {
  return item.guidfixed || `${item.code || "product-set"}-${index}`;
}

async function ensureActiveProductSetHolding(auth: AuthSession, holdingcode: string, businesscode: string): Promise<void> {
  const response = await authFetch("/api/workspace/select-holding", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-bc-backend-url": auth.backendUrl,
      Authorization: `Bearer ${auth.token}`,
    },
    body: JSON.stringify({ backendUrl: auth.backendUrl, holdingcode, businesscode }),
    cache: "no-store",
  });
  const data = await response.json().catch(() => null) as { success?: boolean; message?: string } | null;
  if (!response.ok || data?.success === false) {
    throw new Error(data?.message || "ไม่สามารถเลือกบริษัทใน token ได้");
  }
}

// Custom Barcode Picker Modal for rich barcode selection
function BarcodePickerModal({
  open,
  onClose,
  auth,
  holdingCode,
  businessCode,
  language,
  onSelect,
}: {
  open: boolean;
  onClose: () => void;
  auth: AuthSession | null;
  holdingCode: string;
  businessCode: string;
  language: string;
  onSelect: (row: ProductBarcodeListRow) => void;
}) {
  const text = getBarcodeText(language);
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<ProductBarcodeListRow[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open) return;
    const fetchBarcodes = async () => {
      setLoading(true);
      try {
        const response = await listBarcodes(auth, {
          keyword: query,
          holdingcode: holdingCode,
          businesscode: businessCode,
          limit: 30,
        });
        if (response.success && response.data) {
          setItems(response.data);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    const handler = setTimeout(fetchBarcodes, 300);
    return () => clearTimeout(handler);
  }, [open, query, auth, businessCode, holdingCode]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
      <div className="bg-card border border-border w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh] animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between border-b border-border px-4 py-3 bg-muted/20">
          <div className="flex items-center gap-2">
            <Package className="h-5 w-5 text-primary" />
            <h3 className="text-sm font-bold text-foreground">เลือกบาร์โค้ดสินค้าหลักร่วมชุด</h3>
          </div>
          <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="border-b border-border p-3">
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              autoFocus
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="ค้นหาบาร์โค้ด หรือชื่อสินค้า..."
              className="h-9 !pl-10"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {loading ? (
            <div className="flex flex-col items-center justify-center py-12 gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
              <span>กำลังค้นหาบาร์โค้ด...</span>
            </div>
          ) : items.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground italic">
              ไม่พบข้อมูลบาร์โค้ดสินค้า
            </div>
          ) : (
            <div className="grid gap-1">
              {items.map((row, index) => {
                const price = row.price ?? (row.prices?.[0]?.price ?? 0);
                const stock = row.availableqty ?? row.balanceqty ?? 0;
                const unit = pickName(row.itemunitnames, language) || "ชิ้น";

                return (
                  <button
                    key={row.guidfixed || `${row.barcode || "barcode"}-${index}`}
                    type="button"
                    onClick={() => {
                      onSelect(row);
                      onClose();
                    }}
                    className="w-full text-left p-2.5 rounded-lg border border-border/40 hover:border-primary/30 hover:bg-primary/5 transition flex items-center justify-between gap-3 text-xs"
                  >
                    <div className="min-w-0 flex-1">
                      <span className="font-bold block text-foreground truncate">{pickName(row.names, language)}</span>
                      <div className="flex items-center gap-2 mt-1 text-[10px] text-muted-foreground">
                        <span className="font-mono bg-muted px-1.5 py-0.5 rounded">{row.barcode}</span>
                        <span>•</span>
                        <span>หน่วย: {unit}</span>
                      </div>
                    </div>

                    <div className="text-right shrink-0 flex items-center gap-4">
                      <div>
                        <span className="font-bold text-foreground block">฿{price.toLocaleString()}</span>
                        <span className={cn(
                          "text-[10px] font-semibold",
                          stock > 0 ? "text-green-600 dark:text-green-400" : "text-destructive"
                        )}>
                          สต๊อก: {stock.toLocaleString()}
                        </span>
                      </div>
                      <ArrowRight className="h-4 w-4 text-muted-foreground/60" />
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>

        <div className="border-t border-border px-4 py-2.5 bg-muted/10 text-right">
          <Button variant="outline" size="sm" onClick={onClose}>
            ปิดหน้าต่าง
          </Button>
        </div>
      </div>
    </div>
  );
}

export function ProductSetScreen({ active = true, embedded = false, language = "th" }: ProductSetScreenProps) {
  const lang = normalizeLanguage(language);
  const text = getBarcodeText(lang);
  const { confirm, confirmationDialog } = useConfirmDialog();

  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const selectedShopTokenRef = useRef("");
  const [items, setItems] = useState<Product[]>([]);
  const [selectedGuid, setSelectedGuid] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const setNotice = pushNotice;
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [filterOpen, setFilterOpen] = useState(false);
  const [setFilter, setSetFilter] = useState("all");
  const [selectMode, setSelectMode] = useState(false);
  const [checkedSetKeys, setCheckedSetKeys] = useState<string[]>([]);

  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<"create" | "edit">("create");
  const [editProduct, setEditProduct] = useState<Product | null>(null);

  // Tabs for set editing
  const [activeTab, setActiveTab] = useState<"general" | "components" | "pricing_stock">("general");

  // State for picker
  const [pickerOpen, setPickerOpen] = useState(false);
  const [pickerType, setPickerType] = useState<string>("");
  const [pickerTarget, setPickerTarget] = useState<string>("");
  const [activeOptionIndex, setActiveOptionIndex] = useState<number>(-1);
  const [customBarcodePickerOpen, setCustomBarcodePickerOpen] = useState(false);

  // Cache for component details (price, stock, unit) to drive live simulator
  const [barcodeDetails, setBarcodeDetails] = useState<Record<string, { price: number; stock: number; unit: string; name: string }>>({});

  // User selections in Customer Simulator
  const [simulatorSelections, setSimulatorSelections] = useState<Record<string, string[]>>({});
  const [showSimulator, setShowSimulator] = useState(false);

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(searchInput);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchInput]);

  useEffect(() => {
    // Read auth/workspace session
    const workspaceRaw = localStorage.getItem(workspaceStorageKeys.workspace);
    setAuth(getAuthSession());
    if (workspaceRaw) setWorkspace(JSON.parse(workspaceRaw));
  }, []);

  const activeHoldingCode = workspace?.shop.holdingcode ?? "";
  const activeBusinessCode = normalizeBusinessCode(workspace?.company?.code);
  const shopLanguages = useMemo(() => languageCodesFromWorkspace(workspace), [workspace]);

  // Load products of type SET (itemtype: 2)
  const loadProductSets = useCallback(async () => {
    if (!auth || !activeHoldingCode || !activeBusinessCode) return;
    setLoading(true);
    setNotice(null);
    try {
      const tokenShopKey = `${auth.token}:${activeHoldingCode}:${activeBusinessCode}`;
      if (selectedShopTokenRef.current !== tokenShopKey) {
        await ensureActiveProductSetHolding(auth, activeHoldingCode, activeBusinessCode);
        selectedShopTokenRef.current = tokenShopKey;
      }
      const params = new URLSearchParams({
        q: search,
        limit: "120",
        itemtype: "2",
        materialtype: "3",
      });
      const response = await authFetch(`/api/product?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${auth.token}`,
          "x-bc-backend-url": auth.backendUrl,
        },
      });
      const data = await response.json();
      if (!response.ok || data.success === false) {
        throw new Error(data.message || "Failed to load product sets");
      }
      const rawData = Array.isArray(data.data) ? data.data : [];
      const normalized: Product[] = rawData.map(rawToProduct);

      // Filter only SET items
      const setList = normalized.filter((item) => item.itemtype === 2);
      setItems(setList);

      if (setList.length > 0) {
        const isDesktop = typeof window !== "undefined" && window.matchMedia("(min-width: 768px)").matches;
        if (isDesktop) {
          setSelectedGuid((prev) => prev || setList[0].guidfixed);
        }
      }
    } catch (err: any) {
      setItems([]);
      setNotice({ type: "error", text: err.message || "Error fetching product sets" });
    } finally {
      setLoading(false);
    }
  }, [auth, activeHoldingCode, activeBusinessCode, search]);

  useEffect(() => {
    if (active && auth && activeHoldingCode) {
      void loadProductSets();
    }
  }, [active, auth, activeHoldingCode, loadProductSets]);

  const selectedProduct = useMemo(() => {
    return items.find((item) => item.guidfixed === selectedGuid) ?? items[0] ?? null;
  }, [items, selectedGuid]);

  const visibleSets = useMemo(() => {
    if (setFilter === "component_stock") return items.filter((item) => item.isusesubbarcodes);
    if (setFilter === "bundle_stock") return items.filter((item) => !item.isusesubbarcodes);
    return items;
  }, [items, setFilter]);

  // Load component live details (price, stock, unit)
  const loadComponentDetails = useCallback(async (barcodes: string[]) => {
    if (!auth || !activeHoldingCode || barcodes.length === 0) return;
    try {
      const promises = barcodes.map(async (code) => {
        const response = await authFetch(`/api/product-barcode/list`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.token}`,
          },
          body: JSON.stringify({
            keyword: code,
            holdingcode: activeHoldingCode,
            limit: 1
          })
        });
        const resJson = await response.json();
        if (resJson.success && resJson.data && resJson.data.length > 0) {
          const row = resJson.data[0] as ProductBarcodeListRow;
          return {
            code: row.barcode,
            price: row.price ?? (row.prices?.[0]?.price ?? 0),
            stock: row.availableqty ?? row.balanceqty ?? 0,
            unit: pickName(row.itemunitnames, lang) || "ชิ้น",
            name: pickName(row.names, lang) || row.barcode
          };
        }
        return null;
      });
      const results = await Promise.all(promises);
      setBarcodeDetails((prev) => {
        const nextMap = { ...prev };
        results.forEach((res) => {
          if (res) {
            nextMap[res.code] = res;
          }
        });
        return nextMap;
      });
    } catch (err) {
      console.error("Failed to load component details", err);
    }
  }, [auth, activeHoldingCode, lang]);

  // Gather unique component barcodes to fetch details
  useEffect(() => {
    const barcodes: string[] = [];
    const source = editProduct || selectedProduct;
    if (source?.options) {
      source.options.forEach((group) => {
        group.choices?.forEach((choice) => {
          if (choice.refbarcode && !barcodes.includes(choice.refbarcode)) {
            barcodes.push(choice.refbarcode);
          }
        });
      });
    }
    if (barcodes.length > 0) {
      void loadComponentDetails(barcodes);
    }
  }, [selectedProduct, editProduct, loadComponentDetails]);

  // Reset simulator selections when active product changes
  useEffect(() => {
    const source = editProduct || selectedProduct;
    if (!source?.options) {
      setSimulatorSelections({});
      return;
    }
    const defaultSelections: Record<string, string[]> = {};
    source.options.forEach((group) => {
      const defaultChoice = group.choices?.find((c) => c.isdefault) || group.choices?.[0];
      if (defaultChoice?.guid) {
        defaultSelections[group.guid] = [defaultChoice.guid];
      } else {
        defaultSelections[group.guid] = [];
      }
    });
    setSimulatorSelections(defaultSelections);
  }, [selectedProduct, editProduct]);

  // Helper: Calculate simulator values
  const simulatedBundle = useMemo(() => {
    const source = editProduct || selectedProduct;
    if (!source) return null;

    let price = 0;
    const weight = source.packageweight ?? 0;
    const componentsList: Array<{
      barcode: string;
      name: string;
      qty: number;
      stock: number;
      price: number;
      unit: string;
    }> = [];

    // pricing policy: condition === true (Dynamic pricing), false (Fixed set pricing)
    const isDynamic = source.condition ?? false;

    if (source.options) {
      source.options.forEach((group) => {
        const selectedGuids = simulatorSelections[group.guid] || [];
        const selectedChoices = group.choices?.filter((c) => selectedGuids.includes(c.guid)) || [];

        selectedChoices.forEach((choice) => {
          const detail = choice.refbarcode ? barcodeDetails[choice.refbarcode] : null;
          const compPrice = detail?.price ?? 0;
          const choiceQty = choice.qty ?? 1;
          const addedPrice = Number(choice.price || 0);

          if (isDynamic) {
            price += (compPrice + addedPrice) * choiceQty;
          }

          if (choice.refbarcode) {
            componentsList.push({
              barcode: choice.refbarcode,
              name: detail?.name || pickName(choice.names, lang) || choice.refbarcode,
              qty: choiceQty,
              stock: detail?.stock ?? 0,
              price: compPrice + addedPrice,
              unit: detail?.unit || "ชิ้น"
            });
          }
        });
      });
    }

    // If fixed price, we show cost estimate
    let estimatedCost = 0;
    componentsList.forEach((comp) => {
      estimatedCost += comp.price * comp.qty;
    });

    // Calculate limit stock (Min of stock / qty per choice)
    let totalStock = source.qty ?? 0;
    const isComponentStock = source.isusesubbarcodes ?? false;

    if (isComponentStock && componentsList.length > 0) {
      const stockLimits = componentsList.map((comp) => Math.floor(comp.stock / comp.qty));
      totalStock = Math.min(...stockLimits);
    }

    return {
      price: isDynamic ? price : 0, // 0 means fixed, handle display on frontend
      weight,
      componentsList,
      totalStock,
      isComponentStock,
      isDynamic,
      estimatedCost
    };
  }, [selectedProduct, editProduct, simulatorSelections, barcodeDetails, lang]);

  const makeBlankProductSet = useCallback((): Product => ({
    guidfixed: "",
    holdingcode: activeHoldingCode,
    code: "",
    names: [{ code: "th", name: "" }, { code: "en", name: "" }],
    groupcode: "",
    groupnames: [],
    itemtype: 2, // Set
    vattype: 0,
    materialtype: 3, // Set
    issumpoint: false,
    condition: false, // false = Fixed Price, true = Dynamic Price
    dividevalue: 1,
    standvalue: 1,
    isusesubbarcodes: true, // Default to component stock
    refbarcodes: [],
    bom: [],
    options: [], // Options containing choices (set components)
    packageweight: 0,
    packagelength: 0,
    packagewidth: 0,
    packageheight: 0,
  }), [activeHoldingCode]);

  const handleCreateOpen = () => {
    setEditorMode("create");
    setEditProduct(makeBlankProductSet());
    setActiveTab("general");
    setEditorOpen(true);
  };

  const handleCreateCopyOpen = () => {
    if (!selectedProduct) return;
    setEditorMode("create");
    setEditProduct({
      ...selectedProduct,
      guidfixed: "",
      holdingcode: activeHoldingCode,
      itemtype: 2,
      materialtype: 3,
    });
    setActiveTab("general");
    setEditorOpen(true);
  };

  const handleEditOpen = (p: Product) => {
    setEditorMode("edit");
    setEditProduct({ ...p });
    setActiveTab("general");
    setEditorOpen(true);
  };

  const handleDelete = async (p: Product) => {
    if (!auth || !p.guidfixed) return;
    const ok = await confirm({
      title: "ยืนยันการลบสินค้าชุด?",
      description: `รหัสสินค้าชุด: ${p.code}`,
      tone: "danger",
      confirmLabel: text.delete,
      cancelLabel: text.cancel,
    });
    if (!ok) return;

    try {
      const res = await authFetch(`/api/product/${encodeURIComponent(p.guidfixed)}`, {
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
      void loadProductSets();
      setSelectedGuid("");
    } catch (err: any) {
      setNotice({ type: "error", text: err.message || "Delete failed" });
    }
  };

  const toggleCheckedSet = (key: string) => {
    setCheckedSetKeys((current) =>
      current.includes(key) ? current.filter((item) => item !== key) : [...current, key],
    );
  };

  const handleDeleteSelectedSets = async () => {
    if (!auth || checkedSetKeys.length === 0) return;
    const selectedItems = visibleSets.filter((item, index) => checkedSetKeys.includes(productSetRowKey(item, index)));
    const guids = selectedItems.map((item) => item.guidfixed).filter(Boolean);
    if (guids.length === 0) return;
    const ok = await confirm({
      title: "ยืนยันการลบสินค้าชุด?",
      description: `เลือกไว้ ${guids.length.toLocaleString("th-TH")} รายการ`,
      tone: "danger",
      confirmLabel: text.delete,
      cancelLabel: text.cancel,
    });
    if (!ok) return;
    try {
      for (const guid of guids) {
        const res = await authFetch(`/api/product/${encodeURIComponent(guid)}`, {
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
      setCheckedSetKeys([]);
      setSelectMode(false);
      setSelectedGuid("");
      setNotice({ type: "success", text: text.deleteSuccess });
      void loadProductSets();
    } catch (err: any) {
      setNotice({ type: "error", text: err.message || "Delete failed" });
    }
  };

  const handleSave = async (e: FormEvent) => {
    e.preventDefault();
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
        itemtype: 2, // Always ensure Set type
        materialtype: 3, // Always ensure Set material type
        dividevalue: 1,
        standvalue: 1,
        condition: editProduct.condition ?? false,
      };

      const res = await authFetch(url, {
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

      setNotice({ type: "success", text: text.saveSuccess });
      void loadProductSets();
      setEditorOpen(false);
      if (editorMode === "create" && data.data?.guidfixed) {
        setSelectedGuid(data.data.guidfixed);
      }
    } catch (err: any) {
      setNotice({ type: "error", text: err.message || "Save failed" });
    } finally {
      setSaving(false);
    }
  };

  // Option Group Actions (Set Components)
  const addOptionGroup = () => {
    if (!editProduct) return;
    const current = editProduct.options || [];
    const newGroup: ProductOption = {
      guid: `group_${randomId()}`,
      names: [{ code: "th", name: "กลุ่มส่วนประกอบย่อยใหม่" }],
      choicetype: 1, // Default to single choice
      minselect: 1,
      maxselect: 1,
      choices: [],
    };
    setEditProduct({
      ...editProduct,
      options: [...current, newGroup],
    });
  };

  const removeOptionGroup = (groupIndex: number) => {
    if (!editProduct) return;
    const current = editProduct.options || [];
    setEditProduct({
      ...editProduct,
      options: current.filter((_, idx) => idx !== groupIndex),
    });
  };

  const updateOptionGroupFields = (groupIndex: number, fields: Partial<ProductOption>) => {
    if (!editProduct) return;
    const current = editProduct.options || [];
    setEditProduct({
      ...editProduct,
      options: current.map((group, idx) =>
        idx === groupIndex ? { ...group, ...fields } : group
      ),
    });
  };

  // Option Choice Actions (Items inside group)
  const addChoiceToGroup = (groupIndex: number, entry: ProductBarcodeListRow) => {
    if (!editProduct) return;
    const current = editProduct.options || [];
    const group = current[groupIndex];
    if (!group) return;

    const choices = group.choices || [];

    // Prevent a set from containing itself as a component (infinite expansion risk).
    if (entry.itemcode && editProduct?.code && entry.itemcode === editProduct.code) {
      pushNotice({ type: "error", text: "ไม่สามารถเพิ่มสินค้าชุดนี้เป็นส่วนประกอบของตัวเองได้" });
      return;
    }

    // Avoid duplicates
    if (choices.some(c => c.refbarcode === entry.barcode)) return;

    // Cache details immediately
    const price = entry.price ?? (entry.prices?.[0]?.price ?? 0);
    const stock = entry.availableqty ?? entry.balanceqty ?? 0;
    const unit = pickName(entry.itemunitnames, lang) || "ชิ้น";

    setBarcodeDetails(prev => ({
      ...prev,
      [entry.barcode]: {
        code: entry.barcode,
        price,
        stock,
        unit,
        name: pickName(entry.names, lang)
      }
    }));

    const newChoice: ProductChoice = {
      guid: `choice_${randomId()}`,
      names: entry.names,
      refbarcode: entry.barcode,
      refbarcodenames: entry.names,
      isstock: true,
      isdefault: choices.length === 0, // Default first choice
      qty: 1,
      price: "0",
    };

    updateOptionGroupFields(groupIndex, {
      choices: [...choices, newChoice],
    });
  };

  const removeChoiceFromGroup = (groupIndex: number, choiceIndex: number) => {
    if (!editProduct) return;
    const current = editProduct.options || [];
    const group = current[groupIndex];
    if (!group) return;

    const choices = group.choices || [];
    const nextChoices = choices.filter((_, idx) => idx !== choiceIndex);

    // Auto adjust default choice if the removed one was default
    if (choices[choiceIndex]?.isdefault && nextChoices.length > 0) {
      nextChoices[0].isdefault = true;
    }

    updateOptionGroupFields(groupIndex, {
      choices: nextChoices,
    });
  };

  const updateChoiceFields = (groupIndex: number, choiceIndex: number, fields: Partial<ProductChoice>) => {
    if (!editProduct) return;
    const current = editProduct.options || [];
    const group = current[groupIndex];
    if (!group) return;

    const choices = group.choices || [];
    const nextChoices = choices.map((c, idx) => {
      if (idx !== choiceIndex) {
        // If turning this one to default, unset default from others
        if (fields.isdefault && c.isdefault) {
          return { ...c, isdefault: false };
        }
        return c;
      }
      return { ...c, ...fields };
    });

    updateOptionGroupFields(groupIndex, {
      choices: nextChoices,
    });
  };

  const openPickerForChoice = (groupIndex: number) => {
    setActiveOptionIndex(groupIndex);
    setCustomBarcodePickerOpen(true);
  };

  const handlePickerSelect = (entry: MasterEntry) => {
    if (!editProduct) return;

    if (pickerTarget === "group") {
      setEditProduct({
        ...editProduct,
        groupcode: entry.code,
        groupnames: entry.names,
      });
    }
  };

  // Toggle selections inside Customer Simulator
  const handleSimulatorToggle = (groupGuid: string, choiceGuid: string, multiSelect: boolean) => {
    setSimulatorSelections((prev) => {
      const current = prev[groupGuid] || [];
      if (multiSelect) {
        const next = current.includes(choiceGuid)
          ? current.filter((id) => id !== choiceGuid)
          : [...current, choiceGuid];
        return { ...prev, [groupGuid]: next };
      } else {
        return { ...prev, [groupGuid]: [choiceGuid] };
      }
    });
  };

  // Stats summaries for Left sidebar
  const setsStats = useMemo(() => {
    const total = items.length;
    const componentDeduct = items.filter((item) => item.isusesubbarcodes).length;
    const bundleDeduct = total - componentDeduct;
    return { total, componentDeduct, bundleDeduct };
  }, [items]);

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-background">
      {/* Header Toolbar */}
      <div className="flex shrink-0 items-center justify-between border-b border-border bg-card px-6 py-4 shadow-sm">
        <div>
          <h2 className="text-xl font-extrabold flex items-center gap-2 text-foreground">
            <span className="p-1.5 rounded-lg bg-primary/10 text-primary">
              <Sparkles className="h-5 w-5 animate-pulse" />
            </span>
            จัดการระบบสินค้าชุด (Product Bundles)
          </h2>
          <p className="text-xs text-muted-foreground mt-0.5">จัดกลุ่มคอมโบเซ็ต คอนฟิกตัวเลือกรวม และกติกาการตัดสต๊อกสินค้าหลัก</p>
          <p className="text-xs text-muted-foreground mt-0.5">สินค้าชุด = จับสินค้าหลายตัวขายรวมกันเป็นเซ็ต · ต่างจาก &quot;สูตรผลิต (BOM)&quot; ซึ่งใช้ผลิต/แปรรูปเป็นสินค้าใหม่</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => void loadProductSets()} disabled={loading} className="h-9 hover:bg-muted">
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
            โหลดใหม่
          </Button>
          <Button variant="default" size="sm" onClick={handleCreateCopyOpen} disabled={!selectedProduct} className="h-9 bg-primary hover:bg-primary/95 text-primary-foreground font-bold shadow-md shadow-primary/20">
            <Copy className="h-4 w-4 mr-1" />
            คัดลอก
          </Button>
          <Button variant="default" size="sm" onClick={handleCreateOpen} className="h-9 bg-primary hover:bg-primary/95 text-primary-foreground font-bold shadow-md shadow-primary/20">
            <Plus className="h-4 w-4 mr-1" />
            สร้างสินค้าชุดใหม่
          </Button>
        </div>
      </div>

      {/* Main Split Layout */}
      <div className="flex flex-col md:flex-row flex-1 min-h-0">
        {/* Left Side: Product Set List */}
        <div className={cn("w-full md:w-80 shrink-0 border-b md:border-b-0 md:border-r border-border bg-muted/5 flex flex-col min-h-0", selectedGuid && !editorOpen ? "hidden md:flex" : "flex")}>

          {/* Quick stats badges */}
          <div className="p-3 border-b border-border/60 grid grid-cols-3 gap-1.5 text-center bg-card">
            <div className="p-1.5 rounded-lg border border-border/80 bg-muted/20">
              <span className="text-[10px] text-muted-foreground block">ทั้งหมด</span>
              <strong className="text-sm font-bold text-foreground">{setsStats.total}</strong>
            </div>
            <div className="p-1.5 rounded-lg border border-violet-500/10 bg-violet-500/5">
              <span className="text-[10px] text-violet-500 block truncate">ตัดชิ้นแยก</span>
              <strong className="text-sm font-bold text-violet-600 dark:text-violet-400">{setsStats.componentDeduct}</strong>
            </div>
            <div className="p-1.5 rounded-lg border border-amber-500/10 bg-amber-500/5">
              <span className="text-[10px] text-amber-500 block truncate">ตัดคลังชุด</span>
              <strong className="text-sm font-bold text-amber-600 dark:text-amber-400">{setsStats.bundleDeduct}</strong>
            </div>
          </div>

          <div className="p-3 border-b border-border bg-card">
            <div className="flex flex-wrap gap-2">
              <div className="relative min-w-[220px] flex-1">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  type="search"
                  placeholder="ค้นหารหัส หรือชื่อสินค้าชุด..."
                  className="h-9 !pl-10 bg-background"
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                />
              </div>
              <Button variant={filterOpen ? "secondary" : "outline"} size="sm" type="button" onClick={() => setFilterOpen((current) => !current)}>
                <Filter className="h-4 w-4" />
                ตัวกรอง
              </Button>
              <Button
                variant={selectMode ? "secondary" : "outline"}
                size="sm"
                type="button"
                onClick={() => {
                  setSelectMode((current) => !current);
                  setCheckedSetKeys([]);
                }}
              >
                {selectMode ? <X className="h-4 w-4" /> : <CheckSquare className="h-4 w-4" />}
                {selectMode ? "ยกเลิกเลือก" : "เลือกเพื่อลบ"}
              </Button>
              <Button variant="outline" size="sm" type="button" onClick={() => void handleDeleteSelectedSets()} disabled={!selectMode || checkedSetKeys.length === 0}>
                <Trash2 className="h-4 w-4" />
                {checkedSetKeys.length || ""}
              </Button>
            </div>
            {filterOpen ? (
              <div className="mt-3 flex flex-wrap gap-2 rounded-lg border border-border bg-muted/20 p-2">
                <Button variant={setFilter === "all" ? "secondary" : "outline"} size="sm" type="button" onClick={() => setSetFilter("all")}>ทั้งหมด</Button>
                <Button variant={setFilter === "component_stock" ? "secondary" : "outline"} size="sm" type="button" onClick={() => setSetFilter("component_stock")}>ตัดชิ้นส่วน</Button>
                <Button variant={setFilter === "bundle_stock" ? "secondary" : "outline"} size="sm" type="button" onClick={() => setSetFilter("bundle_stock")}>สต๊อกชุด</Button>
              </div>
            ) : null}
            <div className="mt-2 flex flex-wrap items-center justify-between gap-2 text-xs font-semibold text-muted-foreground">
              <span>สินค้าชุดทั้งหมด</span>
              <span>{visibleSets.length} / {items.length} รายการ</span>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
            {loading ? (
              <div className="p-4 text-center text-sm text-muted-foreground flex justify-center items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin text-primary" />
                กำลังโหลด...
              </div>
            ) : visibleSets.length === 0 ? (
              <div className="py-12 text-center text-sm text-muted-foreground italic">ไม่พบข้อมูลสินค้าชุด</div>
            ) : (
              visibleSets.map((item, index) => {
                const active = item.guidfixed === selectedGuid;
                const rowKey = productSetRowKey(item, index);
                const optionCount = item.options?.length || 0;
                const isCompStock = item.isusesubbarcodes ?? false;

                return (
                  <button
                    key={item.guidfixed}
                    className={cn(
                      "w-full text-left p-3 rounded-xl border transition-all duration-200 flex items-start gap-3 hover:bg-muted/80",
                      active
                        ? "bg-primary/5 border-primary/40 text-foreground shadow-sm shadow-primary/5 translate-x-1"
                        : "border-transparent text-foreground"
                    )}
                    onClick={() => {
                      if (selectMode) {
                        toggleCheckedSet(rowKey);
                      } else {
                        setSelectedGuid(item.guidfixed);
                        setEditorOpen(false);
                      }
                    }}
                  >
                    <div className={cn(
                      "p-2 rounded-lg shrink-0 mt-0.5",
                      active ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground"
                    )}>
                      {selectMode ? (
                        checkedSetKeys.includes(rowKey) ? <CheckSquare className="h-4 w-4" /> : null
                      ) : (
                        <Layers className="h-4 w-4" />
                      )}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="font-bold truncate text-sm text-foreground">{item.code}</div>
                      <div className="text-xs text-muted-foreground truncate">{pickName(item.names, lang)}</div>
                      <div className="flex items-center gap-1.5 mt-1.5 flex-wrap">
                        <Badge variant="secondary" className="text-[9px] py-0 px-1.5 font-medium bg-muted/60">
                          {optionCount} ตัวเลือก
                        </Badge>
                        <Badge
                          variant="outline"
                          className={cn(
                            "text-[9px] py-0 px-1.5 font-semibold",
                            isCompStock
                              ? "border-violet-500/20 text-violet-600 bg-violet-500/5"
                              : "border-amber-500/20 text-amber-600 bg-amber-500/5"
                          )}
                        >
                          {isCompStock ? "ตัดชิ้นส่วน" : "สต๊อกชุด"}
                        </Badge>
                      </div>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* Right Side: Detail or Editor */}
        <div className="flex-1 overflow-hidden bg-background min-h-0 flex flex-col lg:flex-row">

          {/* Middle Card: Detail/Edit Form */}
          <div className="flex-1 overflow-y-auto p-6 border-r border-border">
            {selectedGuid && !editorOpen && (
              <Button
                variant="ghost"
                size="sm"
                className="mb-4 md:hidden flex items-center gap-2 hover:bg-muted"
                onClick={() => setSelectedGuid("")}
              >
                <ChevronLeft className="h-4 w-4" />
                กลับไปหน้ารายชื่อ
              </Button>
            )}

            {editorOpen && editProduct ? (
              /* PRODUCT BUNDLE EDITOR FORM */
              <form onSubmit={handleSave} className="space-y-6 max-w-3xl mx-auto">
                <div className="flex items-center justify-between border-b border-border pb-4">
                  <div>
                    <span className="text-[10px] uppercase tracking-wider font-extrabold text-primary block">
                      {editorMode === "create" ? "NEW PRODUCT BUNDLE" : "EDIT CONFIGURATION"}
                    </span>
                    <h3 className="text-lg font-black text-foreground mt-0.5">
                      {editorMode === "create" ? "✨ สร้างสินค้าชุดคอมโบเซ็ตใหม่" : `📝 แก้ไขสินค้าชุด: ${editProduct.code}`}
                    </h3>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => setShowSimulator(!showSimulator)}
                      className={cn(
                        "h-9 hover:bg-muted font-bold text-xs",
                        showSimulator && "bg-primary/10 text-primary border-primary/20 hover:bg-primary/15"
                      )}
                    >
                      <Play className="mr-1.5 h-3.5 w-3.5 fill-current" />
                      {showSimulator ? "ปิดจำลองการขาย" : "ทดสอบจำลองการขาย"}
                    </Button>
                    <Button type="button" variant="outline" size="sm" onClick={() => setEditorOpen(false)} disabled={saving} className="hover:bg-muted">
                      ยกเลิก
                    </Button>
                    <Button type="submit" variant="default" size="sm" disabled={saving} className="bg-primary hover:bg-primary/95 text-primary-foreground font-bold">
                      {saving && <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />}
                      บันทึกข้อมูลชุด
                    </Button>
                  </div>
                </div>

                {/* Form Navigation Tabs */}
                <div className="flex border-b border-border/60 bg-muted/20 p-1 rounded-xl">
                  <button
                    type="button"
                    onClick={() => setActiveTab("general")}
                    className={cn(
                      "flex-1 py-2 px-3 text-xs font-semibold rounded-lg transition-all flex items-center justify-center gap-1.5",
                      activeTab === "general"
                        ? "bg-card text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    <Info className="h-3.5 w-3.5" />
                    ข้อมูลทั่วไป
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveTab("components")}
                    className={cn(
                      "flex-1 py-2 px-3 text-xs font-semibold rounded-lg transition-all flex items-center justify-center gap-1.5",
                      activeTab === "components"
                        ? "bg-card text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    <Layers className="h-3.5 w-3.5" />
                    สินค้าประกอบชุด ({editProduct.options?.length || 0})
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveTab("pricing_stock")}
                    className={cn(
                      "flex-1 py-2 px-3 text-xs font-semibold rounded-lg transition-all flex items-center justify-center gap-1.5",
                      activeTab === "pricing_stock"
                        ? "bg-card text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    <Settings2 className="h-3.5 w-3.5" />
                    กติกา & ขนาดพัสดุ
                  </button>
                </div>

                {/* Tab Content 1: General Info */}
                {activeTab === "general" && (
                  <div className="space-y-6">
                    <Card className="border-border shadow-sm">
                      <CardHeader className="pb-3">
                        <CardTitle className="text-sm font-bold text-foreground">รายละเอียดพื้นฐาน</CardTitle>
                        <CardDescription className="text-xs">ตั้งค่ารหัสและชื่อเรียกสินค้าชุดคอมโบ</CardDescription>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div className="grid gap-4 sm:grid-cols-2">
                          <div className="space-y-1.5">
                            <span className="font-semibold text-xs text-foreground">รหัสสินค้าชุด (Set SKU) *</span>
                            <Input
                              placeholder="เช่น SET-COMBO-01"
                              value={editProduct.code}
                              onChange={(e) => setEditProduct({ ...editProduct, code: e.target.value })}
                              required
                              disabled={editorMode === "edit"}
                              className="h-9"
                            />
                          </div>
                          <div className="space-y-1.5">
                            <span className="font-semibold text-xs text-foreground">กลุ่มสินค้าหลัก</span>
                            <div className="flex gap-2">
                              <Input
                                placeholder="เลือกกลุ่มสินค้า"
                                value={editProduct.groupcode ? `${editProduct.groupcode} - ${pickName(editProduct.groupnames, lang)}` : ""}
                                readOnly
                                className="bg-muted/40 cursor-default h-9"
                              />
                              <Button
                                type="button"
                                variant="outline"
                                className="h-9"
                                onClick={() => {
                                  setPickerType("group");
                                  setPickerTarget("group");
                                  setPickerOpen(true);
                                }}
                              >
                                เลือก
                              </Button>
                            </div>
                          </div>
                        </div>

                        <div className="space-y-2">
                          <NamesEditor
                            names={editProduct.names}
                            onChange={(names) => setEditProduct({ ...editProduct, names })}
                            languages={shopLanguages}
                            label="ชื่อสินค้าชุด (รองรับหลายภาษา) *"
                          />
                        </div>

                        <div className="space-y-1.5">
                          <span className="font-semibold text-xs text-foreground">รายละเอียดชุดเซ็ต</span>
                          <textarea
                            placeholder="ระบุคำอธิบายสินค้าชุดนี้เพื่อความเข้าใจในการจัดเซ็ตหรือทำรายงาน..."
                            value={editProduct.description || ""}
                            onChange={(e) => setEditProduct({ ...editProduct, description: e.target.value })}
                            className="min-h-24 w-full rounded-lg border border-input bg-background px-3 py-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                          />
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                )}

                {/* Tab Content 2: Set Components */}
                {activeTab === "components" && (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between border-b border-border pb-3">
                      <div>
                        <span className="text-xs font-bold text-foreground">กำหนดกลุ่มตัวเลือกส่วนประกอบ</span>
                        <p className="text-[11px] text-muted-foreground mt-0.5">แบ่งกลุ่มตัวเลือกสินค้า (เช่น กล้องเสริม, เลนส์) ให้ลูกค้าเลือกได้ใน POS</p>
                      </div>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        className="text-primary border-primary/20 hover:bg-primary/5 h-8 font-bold"
                        onClick={addOptionGroup}
                      >
                        <Plus className="mr-1 h-3.5 w-3.5" />
                        เพิ่มกลุ่มตัวเลือกใหม่
                      </Button>
                    </div>

                    {(!editProduct.options || editProduct.options.length === 0) ? (
                      <div className="text-center py-12 border border-dashed border-border rounded-2xl bg-muted/5 flex flex-col items-center justify-center gap-3">
                        <Layers className="h-8 w-8 text-muted-foreground/40" />
                        <div className="space-y-1">
                          <p className="text-xs font-bold text-muted-foreground">ยังไม่มีตัวเลือกส่วนประกอบ</p>
                          <p className="text-[11px] text-muted-foreground">กดปุ่มเพิ่มกลุ่มตัวเลือกด้านบน เพื่อเริ่มสร้างชุดสินค้าคอมโบ</p>
                        </div>
                      </div>
                    ) : (
                      <div className="space-y-6">
                        {editProduct.options.map((group, groupIdx) => (
                          <Card key={group.guid || groupIdx} className="border border-border/80 shadow-sm overflow-hidden">
                            {/* Option Group Header */}
                            <header className="bg-muted/30 px-4 py-3 border-b border-border flex flex-wrap items-center justify-between gap-3">
                              <div className="flex items-center gap-2 flex-1 min-w-[220px]">
                                <span className="text-xs font-extrabold text-primary uppercase">กลุ่มที่ {groupIdx + 1}</span>
                                <Input
                                  placeholder="เช่น กล้องถ่ายรูป / ขาตั้งกล้อง"
                                  value={pickName(group.names, lang)}
                                  onChange={(e) => {
                                    const name = e.target.value;
                                    const names = group.names.map(n => n.code === lang ? { ...n, name } : n);
                                    if (!names.some(n => n.code === lang)) {
                                      names.push({ code: lang, name });
                                    }
                                    updateOptionGroupFields(groupIdx, { names });
                                  }}
                                  className="h-8 text-xs font-bold w-full max-w-[200px]"
                                />
                              </div>

                              <div className="flex items-center gap-3">
                                <span className="text-xs font-semibold text-muted-foreground">กติกา:</span>
                                <div className="flex items-center gap-3 bg-muted/50 px-2.5 py-1 rounded-md border border-border">
                                  <label className="flex items-center gap-1.5 text-xs font-medium cursor-pointer">
                                    <input
                                      type="radio"
                                      name={`choicetype-${groupIdx}`}
                                      checked={group.choicetype === 1}
                                      onChange={() => updateOptionGroupFields(groupIdx, {
                                        choicetype: 1,
                                        minselect: 1,
                                        maxselect: 1,
                                      })}
                                      className="size-3.5 accent-primary"
                                    />
                                    <span>เลือกได้อย่างเดียว (Single)</span>
                                  </label>
                                  <label className="flex items-center gap-1.5 text-xs font-medium cursor-pointer">
                                    <input
                                      type="radio"
                                      name={`choicetype-${groupIdx}`}
                                      checked={group.choicetype === 0}
                                      onChange={() => updateOptionGroupFields(groupIdx, {
                                        choicetype: 0,
                                        minselect: 0,
                                        maxselect: 99,
                                      })}
                                      className="size-3.5 accent-primary"
                                    />
                                    <span>เลือกได้หลายแบบ (Multi)</span>
                                  </label>
                                </div>

                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="icon"
                                  className="h-8 w-8 text-destructive hover:bg-destructive/10 rounded-full"
                                  onClick={() => removeOptionGroup(groupIdx)}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </div>
                            </header>

                            {/* Items inside Group */}
                            <div className="p-4 space-y-3 bg-card">
                              <div className="flex justify-between items-center pb-1">
                                <span className="text-[11px] font-bold text-muted-foreground">รายการบาร์โค้ดในตัวเลือกนี้:</span>
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  className="h-7 text-[11px] font-bold text-primary border-primary/10 hover:bg-primary/5"
                                  onClick={() => openPickerForChoice(groupIdx)}
                                >
                                  <Plus className="mr-1 h-3 w-3" />
                                  ดึงบาร์โค้ดเข้ามา
                                </Button>
                              </div>

                              {(!group.choices || group.choices.length === 0) ? (
                                <p className="text-center py-6 text-xs text-muted-foreground italic border border-dashed border-border/80 rounded-xl">
                                  ยังไม่มีบาร์โค้ดชิ้นส่วน ดึงบาร์โค้ดสินค้าที่ต้องการโดยกดปุ่มขวาบน
                                </p>
                              ) : (
                                <div className="space-y-2">
                                  {group.choices.map((choice, choiceIdx) => {
                                    const detail = choice.refbarcode ? barcodeDetails[choice.refbarcode] : null;
                                    const price = detail?.price ?? 0;
                                    const stock = detail?.stock ?? 0;
                                    const unit = detail?.unit ?? "ชิ้น";

                                    return (
                                      <div
                                        key={choice.guid || choiceIdx}
                                        className="flex flex-wrap items-center justify-between gap-3 p-3 border border-border/50 rounded-xl bg-muted/10 hover:bg-muted/20 transition-all text-xs"
                                      >
                                        <div className="flex-1 min-w-[200px]">
                                          <span className="font-bold text-foreground block">{pickName(choice.names, lang)}</span>
                                          <div className="flex items-center gap-2 mt-1 text-[10px] text-muted-foreground">
                                            <span className="font-mono bg-background px-1.5 py-0.5 rounded border border-border/40">{choice.refbarcode}</span>
                                            <span>•</span>
                                            <span>ราคาตลาด: ฿{price.toLocaleString()}</span>
                                            <span>•</span>
                                            <span className={stock > 0 ? "text-green-600 dark:text-green-400" : "text-destructive"}>
                                              คลัง: {stock} {unit}
                                            </span>
                                          </div>
                                        </div>

                                        {/* Qty & Price adjusting controls */}
                                        <div className="flex items-center gap-3 flex-wrap">
                                          <div className="flex items-center gap-1.5">
                                            <span className="text-[10px] text-muted-foreground">จำนวน:</span>
                                            <Input
                                              type="number"
                                              min={1}
                                              value={choice.qty ?? 1}
                                              onChange={(e) => updateChoiceFields(groupIdx, choiceIdx, { qty: Math.max(1, Number(e.target.value) || 1) })}
                                              className="w-14 h-7 text-center text-xs font-bold"
                                            />
                                          </div>

                                          <div className="flex items-center gap-1.5">
                                            <span className="text-[10px] text-muted-foreground">บวก/ลด (฿):</span>
                                            <Input
                                              type="number"
                                              value={choice.price || "0"}
                                              onChange={(e) => updateChoiceFields(groupIdx, choiceIdx, { price: String(Number(e.target.value) || 0) })}
                                              className="w-18 h-7 text-xs font-bold"
                                            />
                                          </div>

                                          {group.choicetype === 1 && (
                                            <label className="flex items-center gap-1.5 cursor-pointer">
                                              <input
                                                type="checkbox"
                                                checked={choice.isdefault || false}
                                                onChange={(e) => updateChoiceFields(groupIdx, choiceIdx, { isdefault: e.target.checked })}
                                                className="rounded border-border size-3.5 accent-primary cursor-pointer"
                                              />
                                              <span className="text-[10px] text-muted-foreground">ค่าเริ่มต้น</span>
                                            </label>
                                          )}

                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="h-7 w-7 text-destructive/80 hover:bg-destructive/10 rounded-full"
                                            onClick={() => removeChoiceFromGroup(groupIdx, choiceIdx)}
                                          >
                                            <X className="h-4 w-4" />
                                          </Button>
                                        </div>
                                      </div>
                                    );
                                  })}
                                </div>
                              )}
                            </div>
                          </Card>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Tab Content 3: Pricing & Stock Rules */}
                {activeTab === "pricing_stock" && (
                  <div className="space-y-6">
                    {/* Pricing policy */}
                    <Card className="border-border shadow-sm">
                      <CardHeader className="pb-3">
                        <CardTitle className="text-sm font-bold flex items-center gap-2">
                          <DollarSign className="h-4 w-4 text-primary" />
                          นโยบายราคาขาย
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div className="grid gap-3">
                          <label className="flex items-start gap-3 p-3 rounded-xl border border-border bg-muted/10 cursor-pointer hover:bg-muted/20 transition duration-200">
                            <input
                              type="radio"
                              name="price_policy"
                              checked={!editProduct.condition}
                              onChange={() => setEditProduct({ ...editProduct, condition: false })}
                              className="size-4 mt-0.5 text-primary accent-primary"
                            />
                            <div className="space-y-1">
                              <span className="font-bold text-xs text-foreground block">ราคาคงที่ (Fixed Set Price)</span>
                              <span className="text-muted-foreground text-[10px] leading-relaxed block">
                                ยอดราคาชำระคงที่ตามราคาชุดเซ็ตตั้งต้น ลูกค้าเลือกสินค้าตัวเลือกเสริมได้โดยไม่มีการปรับราคาบวก/ลด (ผูกราคาขายผ่านหน้าต่างบาร์โค้ด)
                              </span>
                            </div>
                          </label>

                          <label className="flex items-start gap-3 p-3 rounded-xl border border-border bg-muted/10 cursor-pointer hover:bg-muted/20 transition duration-200">
                            <input
                              type="radio"
                              name="price_policy"
                              checked={editProduct.condition}
                              onChange={() => setEditProduct({ ...editProduct, condition: true })}
                              className="size-4 mt-0.5 text-primary accent-primary"
                            />
                            <div className="space-y-1">
                              <span className="font-bold text-xs text-foreground block">ราคาแปรผันตามรายการที่เลือก (Dynamic Pricing)</span>
                              <span className="text-muted-foreground text-[10px] leading-relaxed block">
                                ราคาชุดจะเปลี่ยนแปลงอัตโนมัติ คำนวณจากผลรวมของราคาบาร์โค้ดสินค้าหลักบวกเพิ่ม/ลดตามจริงของแต่ละชิ้นส่วนที่เลือก
                              </span>
                            </div>
                          </label>
                        </div>
                      </CardContent>
                    </Card>

                    {/* Stock policy */}
                    <Card className="border-border shadow-sm">
                      <CardHeader className="pb-3">
                        <CardTitle className="text-sm font-bold flex items-center gap-2">
                          <Layers className="h-4 w-4 text-primary" />
                          นโยบายตัดสต๊อก
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div className="grid gap-3">
                          <label className="flex items-start gap-3 p-3 rounded-xl border border-border bg-muted/10 cursor-pointer hover:bg-muted/20 transition duration-200">
                            <input
                              type="radio"
                              name="stock_policy"
                              checked={editProduct.isusesubbarcodes ?? false}
                              onChange={() => setEditProduct({ ...editProduct, isusesubbarcodes: true })}
                              className="size-4 mt-0.5 text-primary accent-primary"
                            />
                            <div className="space-y-1">
                              <span className="font-bold text-xs text-foreground block">หักคลังตามจริงชิ้นส่วนประกอบ (Component Inventory Deduct)</span>
                              <span className="text-muted-foreground text-[10px] leading-relaxed block">
                                เมื่อเกิดคำสั่งซื้อ ระบบจะไปหักสต๊อกจากบาร์โค้ดสินค้าจริงแต่ละชิ้นที่ลูกค้าเลือกทันที และสต๊อกพร้อมจำหน่ายของสินค้าชุดนี้จะประเมินแบบพลวัตตามรายการสินค้าที่มีคลังเหลือน้อยที่สุด
                              </span>
                            </div>
                          </label>

                          <label className="flex items-start gap-3 p-3 rounded-xl border border-border bg-muted/10 cursor-pointer hover:bg-muted/20 transition duration-200">
                            <input
                              type="radio"
                              name="stock_policy"
                              checked={!(editProduct.isusesubbarcodes ?? false)}
                              onChange={() => setEditProduct({ ...editProduct, isusesubbarcodes: false })}
                              className="size-4 mt-0.5 text-primary accent-primary"
                            />
                            <div className="space-y-1">
                              <span className="font-bold text-xs text-foreground block">ตัดคลังที่ SKU สินค้าชุดโดยตรง (Bundle Set Inventory)</span>
                              <span className="text-muted-foreground text-[10px] leading-relaxed block">
                                หักสต๊อกจากรายการสินค้าชุดนี้โดยตรง (เหมาะสำหรับชุดคอมโบที่นำมาแพ็คเตรียมกล่องผูกริบบิ้นพร้อมขายไว้ล่วงหน้าแล้ว และมีคลังส่วนตัวไม่ยุ่งเกี่ยวกับคลังหลัก)
                              </span>
                            </div>
                          </label>
                        </div>
                      </CardContent>
                    </Card>

                    {/* Logistics weight & dimensions */}
                    <Card className="border-border shadow-sm">
                      <CardHeader className="pb-3">
                        <CardTitle className="text-sm font-bold flex items-center gap-2">
                          <Box className="h-4 w-4 text-primary" />
                          ขนาดและน้ำหนักกล่องจัดส่ง (Logistics)
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="grid gap-4 sm:grid-cols-4 text-xs">
                        <div className="space-y-1.5">
                          <span className="font-semibold text-muted-foreground">น้ำหนักรวม (kg):</span>
                          <Input
                            type="number"
                            min={0}
                            value={editProduct.packageweight ?? 0}
                            onChange={(e) => setEditProduct({ ...editProduct, packageweight: Math.max(0, Number(e.target.value) || 0) })}
                            className="h-9"
                          />
                        </div>
                        <div className="space-y-1.5">
                          <span className="font-semibold text-muted-foreground">กว้าง (cm):</span>
                          <Input
                            type="number"
                            min={0}
                            value={editProduct.packagewidth ?? 0}
                            onChange={(e) => setEditProduct({ ...editProduct, packagewidth: Math.max(0, Number(e.target.value) || 0) })}
                            className="h-9"
                          />
                        </div>
                        <div className="space-y-1.5">
                          <span className="font-semibold text-muted-foreground">ยาว (cm):</span>
                          <Input
                            type="number"
                            min={0}
                            value={editProduct.packagelength ?? 0}
                            onChange={(e) => setEditProduct({ ...editProduct, packagelength: Math.max(0, Number(e.target.value) || 0) })}
                            className="h-9"
                          />
                        </div>
                        <div className="space-y-1.5">
                          <span className="font-semibold text-muted-foreground">สูง (cm):</span>
                          <Input
                            type="number"
                            min={0}
                            value={editProduct.packageheight ?? 0}
                            onChange={(e) => setEditProduct({ ...editProduct, packageheight: Math.max(0, Number(e.target.value) || 0) })}
                            className="h-9"
                          />
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                )}
              </form>
            ) : selectedProduct ? (
              /* PRODUCT BUNDLE DETAILS VIEW MODE */
              <div className="space-y-6 w-full">
                {/* Premium Banner Header */}
                <div className="relative overflow-hidden rounded-2xl border border-border bg-gradient-to-r from-primary/10 via-violet-500/5 to-card p-6 shadow-sm">
                  <div className="absolute top-0 right-0 p-4 opacity-10">
                    <Sparkles className="h-24 w-24 text-primary animate-pulse" />
                  </div>
                  <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-2">
                        <Badge variant="default" className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 bg-primary/20 text-primary border border-primary/20">
                          สินค้าชุด (Combo Set)
                        </Badge>
                        <Badge variant="outline" className={cn(
                          "text-[10px] font-bold px-2 py-0.5",
                          selectedProduct.isusesubbarcodes
                            ? "border-violet-500/20 text-violet-600 bg-violet-500/5"
                            : "border-amber-500/20 text-amber-600 bg-amber-500/5"
                        )}>
                          {selectedProduct.isusesubbarcodes ? "หักสต๊อกตามส่วนประกอบ" : "หักสต๊อกตาม SKU ชุด"}
                        </Badge>
                      </div>
                      <h3 className="text-2xl font-black text-foreground mt-2 flex items-center gap-2">
                        {selectedProduct.code}
                      </h3>
                      <p className="text-sm font-bold text-muted-foreground mt-0.5">
                        {pickName(selectedProduct.names, lang)}
                      </p>
                      {selectedProduct.description && (
                        <p className="text-xs text-muted-foreground/80 mt-2 max-w-xl leading-relaxed italic bg-background/50 p-2 rounded-lg border border-border/40">
                          {selectedProduct.description}
                        </p>
                      )}
                    </div>

                    <div className="flex flex-wrap items-center gap-2 shrink-0 self-end md:self-center">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => setShowSimulator(!showSimulator)}
                        className={cn(
                          "h-9 hover:bg-muted font-bold text-xs border-border bg-card",
                          showSimulator && "bg-primary/10 text-primary border-primary/20 hover:bg-primary/15"
                        )}
                      >
                        <Play className="mr-1.5 h-3.5 w-3.5 fill-current" />
                        {showSimulator ? "ปิดจำลองการขาย" : "ทดสอบจำลองการขาย"}
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => handleEditOpen(selectedProduct)} className="h-9 hover:bg-muted text-xs font-bold border-border bg-card">
                        <Pencil className="mr-1.5 h-3.5 w-3.5 text-muted-foreground" />
                        แก้ไขข้อมูลชุด
                      </Button>
                      <Button variant="ghost" size="sm" className="h-9 text-xs text-destructive hover:bg-destructive/10 font-bold" onClick={() => void handleDelete(selectedProduct)}>
                        <Trash2 className="mr-1.5 h-3.5 w-3.5" />
                        ลบสินค้าชุด
                      </Button>
                    </div>
                  </div>
                </div>

                {/* Dashboard layout: General info on left, choices on right */}
                <div className="grid gap-6 xl:grid-cols-[1fr_1.5fr]">
                  {/* Left Column: Specs & Settings */}
                  <div className="space-y-6">
                    {/* General specs and policies */}
                    <Card className="border border-border/80 shadow-sm overflow-hidden bg-card/50 backdrop-blur-sm">
                      <CardHeader className="p-4 border-b border-border/60 bg-muted/20">
                        <CardTitle className="text-xs font-bold text-foreground flex items-center gap-1.5">
                          <Settings2 className="h-4 w-4 text-primary" />
                          กติกาและข้อมูลทั่วไป
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="p-4 space-y-4 text-xs">
                        <div className="flex justify-between items-center py-2 border-b border-border/40">
                          <span className="text-muted-foreground">กลุ่มสินค้าหลัก:</span>
                          <strong className="text-foreground font-semibold">
                            {selectedProduct.groupcode ? `${selectedProduct.groupcode} - ${pickName(selectedProduct.groupnames, lang)}` : "ไม่ระบุ"}
                          </strong>
                        </div>
                        <div className="flex justify-between items-center py-2 border-b border-border/40">
                          <span className="text-muted-foreground">นโยบายคิดราคา:</span>
                          <Badge variant="outline" className={cn("text-[10px] font-bold", selectedProduct.condition ? "border-violet-500/20 text-violet-600 bg-violet-500/5" : "border-muted-foreground/20 text-muted-foreground bg-muted/10")}>
                            {selectedProduct.condition ? "ราคาผันแปรตามสินค้าที่เลือกจริง" : "ราคาคงที่ (กำหนดแยกที่บาร์โค้ด)"}
                          </Badge>
                        </div>
                        <div className="flex justify-between items-center py-2 border-b border-border/40">
                          <span className="text-muted-foreground">การตัดสต๊อกจริง:</span>
                          <Badge variant="outline" className={cn("text-[10px] font-bold", selectedProduct.isusesubbarcodes ? "border-violet-500/20 text-violet-600 bg-violet-500/5" : "border-muted-foreground/20 text-muted-foreground bg-muted/10")}>
                            {selectedProduct.isusesubbarcodes ? "ตัดแยกทีละชิ้นส่วนตามที่เลือก" : "ตัดที่ตัว SKU สินค้าชุดโดยตรง"}
                          </Badge>
                        </div>
                        <div className="flex justify-between items-center py-2">
                          <span className="text-muted-foreground">จำนวนกลุ่มตัวเลือกสินค้า:</span>
                          <strong className="text-foreground font-semibold">{selectedProduct.options?.length || 0} กลุ่ม</strong>
                        </div>
                      </CardContent>
                    </Card>

                    {/* Logistics cards */}
                    <Card className="border border-border/80 shadow-sm overflow-hidden bg-card/50 backdrop-blur-sm">
                      <CardHeader className="p-4 border-b border-border/60 bg-muted/20">
                        <CardTitle className="text-xs font-bold text-foreground flex items-center gap-1.5">
                          <Box className="h-4 w-4 text-primary" />
                          ข้อมูลขนส่ง & พัสดุ (Logistics)
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="grid grid-cols-2 gap-3 text-xs p-4 bg-card/40">
                        <div className="bg-background p-3 rounded-xl border border-border/60 flex items-center justify-between">
                          <span className="text-muted-foreground text-[10px]">น้ำหนักรวม:</span>
                          <strong className="text-xs font-extrabold text-foreground">{selectedProduct.packageweight ?? 0} kg</strong>
                        </div>
                        <div className="bg-background p-3 rounded-xl border border-border/60 flex items-center justify-between">
                          <span className="text-muted-foreground text-[10px]">ขนาดกล่อง (กxยxส):</span>
                          <strong className="text-xs font-extrabold text-foreground">
                            {selectedProduct.packagewidth ?? 0}x{selectedProduct.packagelength ?? 0}x{selectedProduct.packageheight ?? 0} cm
                          </strong>
                        </div>
                      </CardContent>
                    </Card>
                  </div>

                  {/* Right Column: Choices and Options */}
                  <div className="space-y-4">
                    <Card className="border border-border/80 shadow-sm overflow-hidden bg-card/50 backdrop-blur-sm">
                      <CardHeader className="p-4 border-b border-border/60 bg-muted/20 flex flex-row items-center justify-between">
                        <CardTitle className="text-xs font-bold text-foreground flex items-center gap-1.5">
                          <Layers className="h-4 w-4 text-primary" />
                          รายการชิ้นส่วนและตัวเลือกภายในเซ็ต
                        </CardTitle>
                        <Badge variant="secondary" className="text-[10px] font-semibold bg-primary/10 text-primary border border-primary/10">
                          {selectedProduct.options?.length || 0} กลุ่มตัวเลือก
                        </Badge>
                      </CardHeader>
                      <CardContent className="p-4 space-y-4">
                        {(!selectedProduct.options || selectedProduct.options.length === 0) ? (
                          <div className="text-center py-10 border border-dashed rounded-xl bg-muted/5 italic text-xs text-muted-foreground">
                            สินค้าชุดนี้ยังไม่มีการกำหนดบาร์โค้ดชิ้นส่วนประกอบ
                          </div>
                        ) : (
                          <div className="space-y-4">
                            {selectedProduct.options.map((group, gIdx) => (
                              <div key={group.guid || gIdx} className="border border-border/60 rounded-xl overflow-hidden bg-background shadow-xs">
                                <div className="bg-muted/15 px-3 py-2 border-b border-border/60 flex justify-between items-center">
                                  <div>
                                    <span className="text-[11px] font-bold text-foreground">
                                      กลุ่มที่ {gIdx + 1}: {pickName(group.names, lang)}
                                    </span>
                                    <span className="text-[9px] text-muted-foreground block mt-0.5">
                                      กติกา: {group.choicetype === 1 ? "ลูกค้าเลือกได้ชิ้นเดียว (Single)" : "ลูกค้าเลือกผสมได้หลายชิ้น (Multi)"}
                                    </span>
                                  </div>
                                </div>
                                <div className="divide-y divide-border/40">
                                  {group.choices?.map((choice, cIdx) => {
                                    const detail = choice.refbarcode ? barcodeDetails[choice.refbarcode] : null;
                                    const price = detail?.price ?? 0;
                                    const stock = detail?.stock ?? 0;

                                    return (
                                      <div key={choice.guid || cIdx} className="flex justify-between items-center p-2.5 text-xs hover:bg-muted/5 transition-all">
                                        <div className="min-w-0 pr-3">
                                          <div className="flex items-center gap-1.5">
                                            <span className="font-semibold text-foreground truncate">{pickName(choice.names, lang)}</span>
                                            {choice.isdefault && (
                                              <Badge className="text-[8px] py-0 px-1 bg-green-500/10 text-green-700 border border-green-500/20 hover:bg-green-500/10 font-bold shrink-0">
                                                เริ่มต้น
                                              </Badge>
                                            )}
                                          </div>
                                          <span className="text-[9px] text-muted-foreground font-mono mt-0.5 block truncate">
                                            บาร์โค้ด: {choice.refbarcode} • คลัง: {stock} • ราคาตลาด: ฿{price.toLocaleString()}
                                          </span>
                                        </div>
                                        <div className="flex items-center gap-3 text-xs shrink-0">
                                          <span className="text-muted-foreground">x{choice.qty ?? 1}</span>
                                          {Number(choice.price) !== 0 && (
                                            <span className={cn("font-bold text-[10px]", Number(choice.price) > 0 ? "text-primary" : "text-green-600")}>
                                              {Number(choice.price) > 0 ? `+฿${Number(choice.price)}` : `-฿${Math.abs(Number(choice.price))}`}
                                            </span>
                                          )}
                                        </div>
                                      </div>
                                    );
                                  })}
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex h-[350px] flex-col items-center justify-center text-muted-foreground gap-3">
                <Layers className="h-12 w-12 text-muted-foreground/30 animate-bounce duration-1000" />
                <div className="text-center space-y-1">
                  <p className="font-bold text-foreground">ยินดีต้อนรับสู่เมนูจัดสินค้าชุด</p>
                  <p className="text-xs text-muted-foreground max-w-xs">กรุณาเลือกรายการสินค้าชุดจากแถบรายชื่อด้านซ้าย เพื่อเริ่มการแก้ไขหรือตรวจสอบกติกาสินค้าชุด</p>
                </div>
              </div>
            )}
          </div>

          {/* Right Column: Customer Storefront Simulator & Analytics */}
          {showSimulator && (
            <div className="w-full lg:w-96 shrink-0 bg-muted/10 p-6 overflow-y-auto flex flex-col gap-6 border-l border-border animate-in slide-in-from-right duration-300">

            {/* Storefront Viewport Simulator */}
            <div className="border border-border/80 rounded-3xl bg-card shadow-lg overflow-hidden flex flex-col relative">

              {/* Device Notch/Top status bar */}
              <div className="bg-muted px-4 py-2 border-b border-border/40 flex items-center justify-between text-[10px] text-muted-foreground font-mono">
                <span>COMBO SIMULATOR</span>
                <span className="px-1.5 py-0.5 rounded bg-primary/10 text-primary font-bold">PREVIEW</span>
              </div>

              {/* Mock product picture */}
              <div className="aspect-[4/3] bg-gradient-to-br from-primary/5 via-violet-500/5 to-fuchsia-500/5 flex items-center justify-center border-b border-border relative overflow-hidden group">
                <div className="absolute inset-0 opacity-10 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]" />
                <div className="p-4 text-center z-10">
                  <ShoppingBag className="h-10 w-10 text-primary/70 mx-auto animate-bounce duration-3000" />
                  <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-extrabold mt-2 block">
                    {editProduct?.code || selectedProduct?.code || "SKU BUNDLE"}
                  </span>
                </div>
              </div>

              {/* Simulator interactive option controls */}
              <div className="p-4 flex-1 space-y-4">
                <div>
                  <h4 className="text-sm font-black text-foreground truncate">
                    {editProduct ? pickName(editProduct.names, lang) : selectedProduct ? pickName(selectedProduct.names, lang) : "ตัวอย่างสินค้าชุดคอมโบเซ็ต"}
                  </h4>
                  <p className="text-[10px] text-muted-foreground mt-1 leading-relaxed">
                    {editProduct?.description || selectedProduct?.description || "จำลองพฤติกรรมการซื้อเพื่อคำนวณราคาและคลังสต๊อกแบบเรียลไทม์"}
                  </p>
                </div>

                {/* Option Groups Selectors */}
                <div className="space-y-4 pt-2 border-t border-border/50">
                  {((editProduct || selectedProduct)?.options || []).map((group) => {
                    const selectedGuids = simulatorSelections[group.guid] || [];
                    const isMulti = group.choicetype === 0;

                    return (
                      <div key={group.guid} className="space-y-2">
                        <span className="text-[11px] font-bold text-foreground flex items-center gap-1.5">
                          {pickName(group.names, lang)}
                          <Badge variant="outline" className="text-[8px] py-0 px-1 font-medium bg-muted">
                            {isMulti ? "เลือกได้หลายอย่าง" : "เลือกได้ชิ้นเดียว"}
                          </Badge>
                        </span>

                        <div className="grid gap-1.5">
                          {group.choices?.map((choice) => {
                            const isSelected = selectedGuids.includes(choice.guid);
                            const detail = choice.refbarcode ? barcodeDetails[choice.refbarcode] : null;
                            const price = detail?.price ?? 0;
                            const addedPrice = Number(choice.price || 0);
                            const finalPrice = price + addedPrice;

                            return (
                              <button
                                key={choice.guid}
                                type="button"
                                onClick={() => handleSimulatorToggle(group.guid, choice.guid, isMulti)}
                                className={cn(
                                  "w-full text-left p-2 rounded-xl border text-xs transition-all duration-200 flex items-center justify-between",
                                  isSelected
                                    ? "bg-primary/10 border-primary/50 text-foreground font-bold shadow-sm shadow-primary/5"
                                    : "border-border/50 hover:bg-muted/50 text-muted-foreground"
                                )}
                              >
                                <span className="truncate pr-2">{pickName(choice.names, lang)}</span>
                                <div className="flex items-center gap-2 shrink-0">
                                  <span className="text-[10px] font-bold text-foreground">
                                    ฿{finalPrice.toLocaleString()}
                                  </span>
                                  <div className={cn(
                                    "size-4 rounded-full border flex items-center justify-center",
                                    isSelected ? "bg-primary border-primary text-primary-foreground" : "border-border"
                                  )}>
                                    {isSelected && <Check className="h-2.5 w-2.5 stroke-[3]" />}
                                  </div>
                                </div>
                              </button>
                            );
                          })}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Total simulated values and Stock trace */}
              <div className="bg-muted/40 p-4 border-t border-border/80 space-y-3">
                <div className="flex items-baseline justify-between">
                  <span className="text-xs text-muted-foreground font-bold">ราคารวมของเซ็ต:</span>
                  <div className="text-right">
                    {simulatedBundle?.isDynamic ? (
                      <span className="text-2xl font-black text-primary animate-pulse">
                        ฿{simulatedBundle.price.toLocaleString()}
                      </span>
                    ) : (
                      <span className="text-xs font-bold text-foreground">
                        ราคาคงที่ (ตามระบบบาร์โค้ด)
                      </span>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-2 text-[10px] bg-card p-2.5 rounded-xl border border-border/80">
                  <div>
                    <span className="text-muted-foreground block">สต๊อกจำลอง:</span>
                    <strong className="text-xs font-extrabold text-foreground">
                      {simulatedBundle ? simulatedBundle.totalStock : 0} ชิ้น
                    </strong>
                  </div>
                  <div>
                    <span className="text-muted-foreground block">การตัดคลัง:</span>
                    <strong className="text-xs font-extrabold text-foreground">
                      {simulatedBundle?.isComponentStock ? "หักตามชิ้นจริง" : "หักสต๊อกชุด"}
                    </strong>
                  </div>
                </div>

                {/* Stock Deduction Path animation */}
                {simulatedBundle && simulatedBundle.isComponentStock && simulatedBundle.componentsList.length > 0 && (
                  <div className="space-y-1.5 pt-2 border-t border-border/50">
                    <span className="text-[9px] uppercase font-bold text-muted-foreground tracking-wider block">เส้นทางการหักคลังสต๊อก (Stock Trace):</span>
                    <div className="space-y-1">
                      {simulatedBundle.componentsList.map((comp, idx) => (
                        <div key={idx} className="flex justify-between items-center text-[9px] text-muted-foreground">
                          <span className="truncate pr-1">• หัก {comp.qty}x {comp.name}</span>
                          <span className="font-bold text-foreground shrink-0 bg-background px-1 border border-border rounded">
                            คลัง: {comp.stock} &rarr; {Math.max(0, comp.stock - comp.qty)}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Profit Margin & Analysis Card */}
            {simulatedBundle && (
              <Card className="border border-border/80 shadow-sm overflow-hidden bg-gradient-to-br from-card via-card to-primary/5">
                <CardHeader className="p-4 pb-2">
                  <CardTitle className="text-xs font-bold text-foreground uppercase flex items-center gap-1.5">
                    <TrendingUp className="h-4 w-4 text-primary" />
                    วิเคราะห์ต้นทุนและกำไร
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-4 pt-0 text-xs space-y-3">
                  <div className="space-y-1.5">
                    <div className="flex justify-between text-muted-foreground text-[11px]">
                      <span>ต้นทุนเฉลี่ยของส่วนประกอบ:</span>
                      <strong className="text-foreground">฿{simulatedBundle.estimatedCost.toLocaleString()}</strong>
                    </div>
                    {simulatedBundle.isDynamic && (
                      <div className="flex justify-between text-muted-foreground text-[11px]">
                        <span>ราคาขายจำลอง:</span>
                        <strong className="text-foreground">฿{simulatedBundle.price.toLocaleString()}</strong>
                      </div>
                    )}
                  </div>

                  {simulatedBundle.isDynamic && simulatedBundle.price > 0 && (
                    <div className="space-y-1.5 pt-2 border-t border-border/40">
                      {(() => {
                        const profit = simulatedBundle.price - simulatedBundle.estimatedCost;
                        const marginPercent = Math.round((profit / simulatedBundle.price) * 100);
                        const isLoss = profit < 0;

                        return (
                          <>
                            <div className="flex justify-between text-[11px]">
                              <span className="font-bold text-muted-foreground">กำไรขั้นต้นประเมิน:</span>
                              <strong className={cn("font-extrabold", isLoss ? "text-destructive" : "text-green-600 dark:text-green-400")}>
                                {isLoss ? "-" : ""}฿{Math.abs(profit).toLocaleString()} ({marginPercent}%)
                              </strong>
                            </div>
                            <div className="w-full bg-muted rounded-full h-1.5 overflow-hidden">
                              <div
                                className={cn("h-full rounded-full transition-all duration-500", isLoss ? "bg-destructive" : "bg-green-500")}
                                style={{ width: `${Math.min(100, Math.max(0, marginPercent))}%` }}
                              />
                            </div>
                            {isLoss ? (
                              <p className="text-[10px] text-destructive flex items-center gap-1">
                                <AlertCircle className="h-3.5 w-3.5 shrink-0" />
                                คำเตือน: ราคาเสนอขายต่ำกว่าต้นทุนรวมของสินค้าในเซ็ต!
                              </p>
                            ) : (
                              <p className="text-[10px] text-muted-foreground flex items-center gap-1">
                                <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-green-500" />
                                อัตรากำไรอยู่ในเกณฑ์มาตรฐานที่ระบบประเมินไว้
                              </p>
                            )}
                          </>
                        );
                      })()}
                    </div>
                  )}

                  {!simulatedBundle.isDynamic && (
                    <div className="p-2 rounded-lg bg-muted/40 border border-border text-[10px] text-muted-foreground leading-relaxed">
                      * เนื่องจากเลือกใช้นโยบาย <strong>ราคาเซ็ตคงที่</strong> อัตราส่วนกำไรที่แน่นอนจะขึ้นอยู่กับราคาขายที่ผูกไว้กับบาร์โค้ดของสินค้าชุด SKU นี้เอง
                    </div>
                  )}
                </CardContent>
              </Card>
            )}
          </div>
          )}
        </div>
      </div>

      {/* Shared Master Pickers */}
      {pickerOpen && (
        <MasterPicker
          open={pickerOpen}
          onClose={() => setPickerOpen(false)}
          auth={auth}
          language={lang}
          master={pickerType as any}
          title={pickerType ? `ค้นหา ${pickerType}` : ""}
          onSelect={handlePickerSelect}
          placement="dialog"
        />
      )}

      {/* Custom Barcode Picker Modal */}
      {customBarcodePickerOpen && (
        <BarcodePickerModal
          open={customBarcodePickerOpen}
          onClose={() => setCustomBarcodePickerOpen(false)}
          auth={auth}
          holdingCode={activeHoldingCode}
          businessCode={activeBusinessCode}
          language={lang}
          onSelect={(entry) => {
            if (activeOptionIndex >= 0) {
              addChoiceToGroup(activeOptionIndex, entry);
              setActiveOptionIndex(-1);
            }
          }}
        />
      )}

      {confirmationDialog}
    </div>
  );
}

"use client";

import { useState, useEffect, useMemo, useRef, useCallback } from "react";
import {
  ShoppingBag,
  Store,
  Settings,
  FileSpreadsheet,
  Upload,
  Check,
  AlertCircle,
  RefreshCw,
  Search,
  ExternalLink,
  ChevronRight,
  Database,
  ArrowRightLeft,
  X,
  CloudLightning,
  CheckCircle2,
  Trash2
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { listBarcodes, updateBarcode } from "@/lib/product-barcode/api";
import { formatDefaultDate, resolveWorkspaceDateTimeDisplayOptions } from "@/lib/date-time";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { getBarcodeText } from "@/lib/product-barcode/language";
import {
  branchDisplayName,
  shopDisplayName,
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
  WORKSPACE_CHANGED_EVENT
} from "@/lib/workspace-models";
import { pickName } from "@/lib/product-barcode/utils";
import {
  type ProductBarcode,
  type MarketplaceProductMap,
  type MarketplaceSKUMap,
  emptyMarketplaceProductMap,
  emptyMarketplaceSKUMap,
} from "@/lib/product-barcode/types";
import { cn } from "@/lib/utils";
import { pushNotice } from "@/lib/toast";
import { normalizeBusinessCode } from "@/lib/business-code";
import { authFetch, getAuthSession } from "@/lib/client-auth-session";

type MarketplaceScreenProps = {
  platform: "shopee" | "lazada" | "tiktok";
  embedded?: boolean;
  language?: LanguageCode;
};

type ShopConnection = {
  id: string;
  holdingcode: string;
  shopname: string;
  platform: "shopee" | "lazada" | "tiktok";
  status: "connected" | "disconnected";
  connectedat: string;
  itemcount: number;
};

export function MarketplaceMappingsScreen({ platform, embedded = false, language = "th" }: MarketplaceScreenProps) {
  const lang = normalizeLanguage(language);
  const textU = getBarcodeText(lang);

  const [activeTab, setActiveTab] = useState<"connections" | "mappings" | "logs">("connections");
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const setNotice = pushNotice;

  // Mappings page states
  const [barcodes, setBarcodes] = useState<any[]>([]);
  const [searchQuery, setSearchQuery] = useState("");

  // Bulk import states
  const [showImportPanel, setShowImportPanel] = useState(false);
  const [importText, setImportText] = useState("");
  const [importStatus, setImportStatus] = useState<{ success: number; failed: number; logs: string[] } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const platformName = useMemo(() => {
    switch (platform) {
      case "shopee": return "Shopee";
      case "lazada": return "Lazada";
      case "tiktok": return "TikTok Shop";
      default: return platform;
    }
  }, [platform]);

  // Shop connection mock data (persisted in local state/simulated)
  const [shops, setShops] = useState<ShopConnection[]>([
    {
      id: "conn_1",
      holdingcode: "shop_shopee_th",
      shopname: "Ban Chiang Official Store (Shopee)",
      platform: "shopee",
      status: "connected",
      connectedat: "2026-05-10T14:30:00Z",
      itemcount: 142
    },
    {
      id: "conn_2",
      holdingcode: "shop_laz_b2c",
      shopname: "Ban Chiang Outlet (Lazada)",
      platform: "lazada",
      status: "connected",
      connectedat: "2026-05-12T09:15:00Z",
      itemcount: 88
    },
    {
      id: "conn_3",
      holdingcode: "shop_tiktok_mall",
      shopname: "Ban Chiang Store (TikTok Shop)",
      platform: "tiktok",
      status: "disconnected",
      connectedat: "-",
      itemcount: 0
    }
  ]);

  // Auth Dialog state
  const [authDialog, setAuthDialog] = useState<{
    open: boolean;
    holdingCode: string;
    shopName: string;
  }>({
    open: false,
    holdingCode: "",
    shopName: ""
  });

  useEffect(() => {
    // Read auth/workspace session
    const workspaceRaw = localStorage.getItem(workspaceStorageKeys.workspace);
    setAuth(getAuthSession());
    if (workspaceRaw) setWorkspace(JSON.parse(workspaceRaw));
  }, []);

  const activeHoldingCode = workspace?.shop.holdingcode ?? "";
  const activeBusinessCode = normalizeBusinessCode(workspace?.company?.code);
  const dateDisplayOptions = useMemo(
    () => resolveWorkspaceDateTimeDisplayOptions(workspace, lang),
    [lang, workspace],
  );

  // Load product list
  const loadProductBarcodes = useCallback(async () => {
    if (!auth || !activeHoldingCode || !activeBusinessCode) return;
    setLoading(true);
    setNotice(null);
    try {
      const res = await listBarcodes(auth, {
        holdingcode: activeHoldingCode,
        businesscode: activeBusinessCode,
        limit: 100,
        offset: 0,
        sortfield: "barcode",
        sortorder: "asc"
      });
      if (res.success && Array.isArray(res.data)) {
        setBarcodes(res.data);
      } else {
        throw new Error(res.message || "Failed to load product list.");
      }
    } catch (err: any) {
      setNotice({ type: "error", text: err.message || "Error fetching barcodes" });
    } finally {
      setLoading(false);
    }
  }, [auth, activeBusinessCode, activeHoldingCode]);

  useEffect(() => {
    if (auth && activeHoldingCode && activeBusinessCode && activeTab === "mappings") {
      void loadProductBarcodes();
    }
  }, [auth, activeBusinessCode, activeHoldingCode, activeTab, loadProductBarcodes]);

  // Handle new shop auth
  const handleConnectShop = () => {
    if (!authDialog.holdingCode) return;

    setShops((prev) =>
      prev.map(s => s.platform === platform ? {
        ...s,
        holdingcode: authDialog.holdingCode,
        shopname: authDialog.shopName || `ร้านค้า ${platform.toUpperCase()} (${authDialog.holdingCode})`,
        status: "connected",
        connectedat: new Date().toISOString(),
        itemcount: 0
      } : s)
    );

    setAuthDialog({ open: false, holdingCode: "", shopName: "" });
    setNotice({ type: "success", text: `เชื่อมต่อร้านค้า ${platformName} สำเร็จ!` });
  };

  const handleDisconnectShop = (id: string) => {
    setShops((prev) =>
      prev.map(s => s.id === id ? {
        ...s,
        status: "disconnected",
        connectedat: "-",
        itemcount: 0
      } : s)
    );
    setNotice({ type: "info", text: `ยกเลิกการเชื่อมต่อร้านค้า ${platformName} แล้ว` });
  };

  // Bulk import processor
  const handleBulkImport = useCallback(async (textData: string) => {
    if (!textData.trim() || !auth) return;

    const lines = textData.split(/\r?\n/).map(l => l.trim()).filter(Boolean);
    if (lines.length < 2) {
      setImportStatus({
        success: 0,
        failed: 0,
        logs: ["Error: ข้อมูลต้องมีอย่างน้อย 2 แถว (หัวตาราง + เนื้อหา)"],
      });
      return;
    }

    const firstLine = lines[0];
    const delimiter = firstLine.includes("\t") ? "\t" : (firstLine.includes(",") ? "," : ";");
    const headers = firstLine.split(delimiter).map(h => h.trim().toLowerCase());

    const holdingCodeIdx = headers.findIndex(h => h.includes("shop") || h.includes("ร้านค้า") || h.includes("บัญชี"));
    const marketItemIdIdx = headers.findIndex(h => h.includes("product_id") || h.includes("item_id") || h.includes("สินค้าบนเว็บ") || h.includes("market_item_id"));
    const sellerSkuIdx = headers.findIndex(h => h.includes("sellersku") || h.includes("sku") || h.includes("รหัสคู่ค้า") || h.includes("sellersku"));
    const barcodeIdx = headers.findIndex(h => h.includes("barcode") || h.includes("บาร์โค้ด") || h.includes("รหัสบาร์โค้ด") || h.includes("บาร์โค๊ด"));
    const marketModelIdIdx = headers.findIndex(h => h.includes("variant_id") || h.includes("model_id") || h.includes("รหัสตัวเลือกย่อย") || h.includes("market_model_id"));

    if (sellerSkuIdx === -1 && barcodeIdx === -1) {
      setImportStatus({
        success: 0,
        failed: 0,
        logs: ["Error: ไม่พบคอลัมน์หลักสำหรับจับคู่ (ต้องมีอย่างน้อย 'Barcode' หรือ 'Seller SKU')"],
      });
      return;
    }

    setSaving(true);
    let successCount = 0;
    let failedCount = 0;
    const logs: string[] = [];

    const updatedBarcodes = [...barcodes];

    for (let i = 1; i < lines.length; i++) {
      const row = lines[i].split(delimiter).map(v => v.trim());
      if (row.length === 0 || (row.length === 1 && !row[0])) continue;

      const holdingCode = holdingCodeIdx !== -1 && row[holdingCodeIdx] ? row[holdingCodeIdx] : `shop_${platform}_01`;
      const marketItemId = marketItemIdIdx !== -1 && row[marketItemIdIdx] ? row[marketItemIdIdx] : "";
      const sellerSku = sellerSkuIdx !== -1 && row[sellerSkuIdx] ? row[sellerSkuIdx] : "";
      const barcodeValue = barcodeIdx !== -1 && row[barcodeIdx] ? row[barcodeIdx] : "";
      const marketModelId = marketModelIdIdx !== -1 && row[marketModelIdIdx] ? row[marketModelIdIdx] : "";

      let matchedIdx = -1;
      if (barcodeValue) {
        matchedIdx = updatedBarcodes.findIndex(b => b.barcode === barcodeValue);
      }
      if (matchedIdx === -1 && sellerSku) {
        matchedIdx = updatedBarcodes.findIndex(b => b.sellersku === sellerSku);
      }

      if (matchedIdx === -1) {
        failedCount++;
        logs.push(`Row ${i + 1}: ไม่พบบาร์โค้ด/SKU "${barcodeValue || sellerSku}" ในคลังสินค้า`);
        continue;
      }

      const item = updatedBarcodes[matchedIdx];

      try {
        const fetchRes = await authFetch(`/api/product-barcode/${encodeURIComponent(item.guidfixed)}`, {
          headers: {
            "x-bc-backend-url": auth.backendUrl,
            Authorization: `Bearer ${auth.token}`,
          }
        });
        const productData = await fetchRes.json();
        if (!fetchRes.ok || !productData.success) {
          throw new Error("Failed to load details");
        }

        // Modify the RAW product doc from the GET response directly (preserve every backend field —
        // names, producttype, units, etc.). Re-saving through the display-oriented rawToProductBarcode
        // transform was lossy and failed backend validation ("names must contain", producttype struct).
        const raw = (productData.data ?? {}) as Record<string, unknown>;

        const mProducts = Array.isArray(raw.marketplaceproducts) ? (raw.marketplaceproducts as MarketplaceProductMap[]) : [];
        const hasShopMap = mProducts.some(
          m => m.platform === platform && m.holdingcode === holdingCode && m.marketitemid === marketItemId
        );
        if (!hasShopMap) {
          mProducts.push({
            ...emptyMarketplaceProductMap(platform),
            holdingcode: holdingCode,
            marketitemid: marketItemId,
            syncstatus: "linked",
          });
        }
        raw.marketplaceproducts = mProducts;

        const refBarcodes = Array.isArray(raw.refbarcodes) ? (raw.refbarcodes as Record<string, unknown>[]) : [];
        const subBarcodeIdx = refBarcodes.findIndex(b => b.barcode === (barcodeValue || item.barcode));
        if (subBarcodeIdx !== -1) {
          const subB = refBarcodes[subBarcodeIdx];
          const mappings = Array.isArray(subB.marketplaceskumappings) ? (subB.marketplaceskumappings as MarketplaceSKUMap[]) : [];
          const matchMapIdx = mappings.findIndex(m => m.platform === platform && m.holdingcode === holdingCode);

          const mappingData: MarketplaceSKUMap = {
            ...emptyMarketplaceSKUMap(platform, holdingCode, marketItemId),
            marketmodelid: marketModelId,
            sellersku: sellerSku || (typeof subB.sellersku === "string" ? subB.sellersku : "") || "",
            status: "LIVE",
            lastsyncat: new Date().toISOString(),
          };

          if (matchMapIdx >= 0) {
            mappings[matchMapIdx] = { ...mappings[matchMapIdx], ...mappingData };
          } else {
            mappings.push(mappingData);
          }
          subB.marketplaceskumappings = mappings;
          if (sellerSku) subB.sellersku = sellerSku;
          refBarcodes[subBarcodeIdx] = subB;
        }
        raw.refbarcodes = refBarcodes;

        const guid = typeof raw.guidfixed === "string" ? raw.guidfixed : item.guidfixed;
        const saveRes = await updateBarcode(auth, guid, raw as unknown as ProductBarcode);
        if (!saveRes.success) {
          throw new Error(saveRes.message || "Save error");
        }

        successCount++;
        logs.push(`Row ${i + 1}: สำเร็จ (Barcode: ${item.barcode} -> Shop: ${holdingCode})`);
      } catch (err: any) {
        failedCount++;
        logs.push(`Row ${i + 1}: ล้มเหลวขณะบันทึกข้อมูล (${err.message})`);
      }
    }

    setSaving(false);
    setImportStatus({
      success: successCount,
      failed: failedCount,
      logs,
    });

    void loadProductBarcodes();
  }, [auth, barcodes, platform, activeHoldingCode, loadProductBarcodes]);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      setImportText(text);
      void handleBulkImport(text);
    };
    reader.readAsText(file);
  };

  const filteredBarcodes = useMemo(() => {
    return barcodes.filter((item) => {
      const matchesSearch =
        item.barcode.includes(searchQuery) ||
        item.itemcode.toLowerCase().includes(searchQuery.toLowerCase()) ||
        pickName(item.names, lang).toLowerCase().includes(searchQuery.toLowerCase());

      return matchesSearch;
    });
  }, [barcodes, searchQuery, lang]);

  const activePlatformShops = useMemo(() => {
    return shops.filter(s => s.platform === platform);
  }, [shops, platform]);

  return (
    <div className="flex h-full flex-col bg-background p-4 text-foreground space-y-4 overflow-y-auto">
      {/* Header */}
      <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border pb-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight flex items-center gap-2">
            <ShoppingBag className="text-primary h-6 w-6" />
            เชื่อมต่อร้านค้า {platformName} (Mappings & Integration)
          </h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            จัดการบัญชีและการแมปจับคู่รหัสคู่ค้า (Product SKU Mappings) ของช่องทาง {platformName} ทั้งหมด
          </p>
        </div>
        <div className="flex gap-2">
          {activeTab === "mappings" && (
            <Button
              onClick={() => setShowImportPanel((p) => !p)}
              variant="outline"
              className="text-primary border-primary/20 hover:bg-primary/5 h-9"
            >
              <Upload className="mr-1.5 h-4 w-4" />
              นำเข้าไฟล์จับคู่สินค้า (Bulk Import)
            </Button>
          )}
          <Button
            onClick={() => {
              if (activeTab === "mappings") void loadProductBarcodes();
              setNotice({ type: "success", text: "อัปเดตข้อมูลล่าสุดเรียบร้อย" });
            }}
            variant="ghost"
            size="icon"
            className="h-9 w-9 shrink-0 border border-border"
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </header>

      {/* Tabs list */}
      <div className="flex border-b border-border/60">
        <button
          onClick={() => setActiveTab("connections")}
          className={cn(
            "px-4 py-2 text-sm font-medium border-b-2 transition-all -mb-px flex items-center gap-1.5",
            activeTab === "connections"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          )}
        >
          <Store className="h-4 w-4" />
          บัญชีเชื่อมต่อ ({activePlatformShops.filter(s => s.status === "connected").length})
        </button>
        <button
          onClick={() => setActiveTab("mappings")}
          className={cn(
            "px-4 py-2 text-sm font-medium border-b-2 transition-all -mb-px flex items-center gap-1.5",
            activeTab === "mappings"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          )}
        >
          <ArrowRightLeft className="h-4 w-4" />
          ตารางจับคู่สินค้า ({filteredBarcodes.length})
        </button>
        <button
          onClick={() => setActiveTab("logs")}
          className={cn(
            "px-4 py-2 text-sm font-medium border-b-2 transition-all -mb-px flex items-center gap-1.5",
            activeTab === "logs"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          )}
        >
          <Database className="h-4 w-4" />
          ประวัติการอัปเดตล่าสุด / Logs
        </button>
      </div>

      {/* 1. Connections Tab */}
      {activeTab === "connections" && (
        <div className="grid gap-4 md:grid-cols-3">
          {activePlatformShops.map((shop) => (
            <Card key={shop.id} className="relative overflow-hidden border-border bg-card/60 backdrop-blur-sm">
              <div className={cn(
                "absolute top-0 left-0 right-0 h-1",
                shop.platform === "shopee" && "bg-[#ee4d2d]",
                shop.platform === "lazada" && "bg-[#101566]",
                shop.platform === "tiktok" && "bg-black"
              )} />
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <Badge variant={shop.status === "connected" ? "default" : "secondary"} className={cn(
                    shop.status === "connected" && "bg-green-500/10 text-green-700 hover:bg-green-500/15 border-green-500/30"
                  )}>
                    {shop.status === "connected" ? "เชื่อมต่อแล้ว" : "ยังไม่ได้ผูกบัญชี"}
                  </Badge>
                  <span className="text-xs uppercase font-bold text-muted-foreground">{shop.platform}</span>
                </div>
                <CardTitle className="text-lg mt-2 flex items-center gap-2">
                  <Store className="h-5 w-5 text-muted-foreground" />
                  {shop.status === "connected" ? shop.shopname : `ผูกบัญชี ${platformName}`}
                </CardTitle>
                <CardDescription className="font-mono text-xs">
                  กลุ่มกิจการ: {shop.status === "connected" ? shop.holdingcode : "-"}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4 pt-0">
                <div className="text-xs text-muted-foreground space-y-1 bg-muted/40 p-2.5 rounded-lg border border-border/40">
                  <div className="flex justify-between">
                    <span>วันที่เชื่อมต่อ:</span>
                    <span className="font-medium text-foreground">
                      {shop.status === "connected" ? formatDefaultDate(shop.connectedat, dateDisplayOptions) : "-"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span>สินค้าที่จับคู่สำเร็จ:</span>
                    <span className="font-medium text-foreground">
                      {shop.status === "connected" ? `${shop.itemcount} รายการ` : "-"}
                    </span>
                  </div>
                </div>

                <div className="flex justify-end gap-2 pt-2">
                  {shop.status === "connected" ? (
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-destructive border-destructive/20 hover:bg-destructive/5 hover:text-destructive h-8"
                      onClick={() => handleDisconnectShop(shop.id)}
                    >
                      <Trash2 className="mr-1.5 h-3.5 w-3.5" />
                      ตัดการเชื่อมต่อ
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      className="h-8"
                      onClick={() => setAuthDialog({
                        open: true,
                        holdingCode: "",
                        shopName: ""
                      })}
                    >
                      <CloudLightning className="mr-1.5 h-3.5 w-3.5" />
                      เชื่อมโยงสิทธิ์การซิงค์
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 2. Mappings Tab */}
      {activeTab === "mappings" && (
        <div className="space-y-4">
          {/* Bulk Import Panel */}
          {showImportPanel && (
            <Card className="border-primary/20 bg-primary/5">
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-semibold flex items-center gap-1.5 text-primary">
                    <FileSpreadsheet className="h-4.5 w-4.5" />
                    นำเข้าข้อมูลผูกรหัสบาร์โค้ด {platformName} จำนวนมาก (Bulk Mappings Import)
                  </CardTitle>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 text-muted-foreground hover:text-foreground"
                    onClick={() => {
                      setShowImportPanel(false);
                      setImportStatus(null);
                    }}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
                <CardDescription className="text-xs text-muted-foreground">
                  นำเข้าไฟล์ CSV หรือ วางคอลัมน์คัดลอกจาก Excel เพื่ออัปเดตรหัส SKU คู่ค้าและข้อมูลการเชื่อมโยงระบบของช่องทาง {platformName}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4 text-xs">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <span className="font-semibold text-foreground">1. วิธีการนำเข้าข้อมูล</span>
                    <div className="flex items-center gap-2">
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
                      <span className="text-muted-foreground">หรือคัดลอกจาก Excel แล้ววางในฟอร์มด้านล่าง:</span>
                    </div>
                  </div>
                </div>

                <div className="space-y-1">
                  <textarea
                    placeholder={`วางข้อมูลแถวที่มีหัวตารางที่นี่ (เช่น:&#10;รหัสกลุ่มกิจการ&#9;Barcode&#9;Seller SKU&#9;Market Product ID&#9;Market Variant ID&#10;shop_${platform}_01&#9;8850123456789&#9;sku-red-01&#9;123456789&#9;98765432)`}
                    value={importText}
                    onChange={(e) => setImportText(e.target.value)}
                    className="h-28 w-full rounded border border-input bg-background p-2 font-mono text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  />
                  <div className="flex justify-between items-center text-[10px] text-muted-foreground">
                    <span>* จำเป็นต้องมีคอลัมน์ Barcode หรือ Seller SKU เพื่อจับคู่และแมปสินค้าปลายทาง</span>
                    <div className="flex gap-2">
                      {importText && (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2 text-muted-foreground"
                          onClick={() => { setImportText(""); setImportStatus(null); }}
                        >
                          ล้างค่า
                        </Button>
                      )}
                      <Button
                        size="sm"
                        className="h-6 px-3"
                        disabled={!importText.trim() || saving}
                        onClick={() => void handleBulkImport(importText)}
                      >
                        {saving && <RefreshCw className="mr-1 h-3 w-3 animate-spin" />}
                        นำเข้ารหัสจับคู่ทั้งหมด
                      </Button>
                    </div>
                  </div>
                </div>

                {/* Import result logs */}
                {importStatus && (
                  <div className="rounded border border-border bg-background p-3 space-y-2 text-xs shadow-inner">
                    <div className="flex items-center gap-4 font-semibold text-sm">
                      <span className="text-green-600 flex items-center gap-1">
                        <CheckCircle2 className="h-4 w-4" /> ประมวลผลสำเร็จ {importStatus.success} รายการ
                      </span>
                      {importStatus.failed > 0 && (
                        <span className="text-destructive flex items-center gap-1">
                          <AlertCircle className="h-4 w-4" /> ล้มเหลว {importStatus.failed} รายการ
                        </span>
                      )}
                    </div>
                    <div className="max-h-28 overflow-y-auto space-y-1 font-mono text-[10px] text-muted-foreground border-t border-border pt-2">
                      {importStatus.logs.map((log, idx) => (
                        <div
                          key={idx}
                          className={log.includes("ไม่พบ") || log.includes("ล้มเหลว") ? "text-destructive/90 bg-destructive/5 p-1 rounded" : "text-green-600/90"}
                        >
                          {log}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Search & Filter Toolbar */}
          <div className="flex items-center gap-3">
            <div className="relative w-full max-w-xs">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="ค้นหาบาร์โค้ด, รหัสสินค้า, ชื่อ..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-9 !pl-10"
              />
            </div>
          </div>

          {/* Mappings Table */}
          <div className="overflow-hidden rounded-lg border border-border bg-card">
            <div className="overflow-x-auto">
              <table className="w-full border-collapse text-left text-sm">
                <thead className="border-b border-border bg-muted/40 font-medium">
                  <tr>
                    <th className="p-3">สินค้าหลัก / บาร์โค้ด</th>
                    <th className="p-3">รหัสพัสดุ (Seller SKU)</th>
                    <th className="p-3">ร้านค้าที่เชื่อมโยง ({platformName})</th>
                    <th className="p-3">รหัสสินค้าปลายทาง (Market Product ID)</th>
                    <th className="p-3">รหัสตัวเลือกสินค้า (Market Variant ID)</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {loading ? (
                    <tr>
                      <td colSpan={5} className="p-8 text-center text-muted-foreground">
                        <RefreshCw className="mx-auto h-6 w-6 animate-spin text-primary" />
                        <span className="mt-2 block text-xs">กำลังโหลดสินค้าคลังหลัก...</span>
                      </td>
                    </tr>
                  ) : filteredBarcodes.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="p-8 text-center text-muted-foreground text-xs">
                        ไม่พบข้อมูลสินค้าจับคู่ช่องทาง {platformName} ในระบบ
                      </td>
                    </tr>
                  ) : (
                    filteredBarcodes.map((item) => {
                      const platformMap = item.raw?.marketplace_products?.find((m: any) => m.platform === platform);
                      const subMappings = item.raw?.refbarcodes?.[0]?.marketplace_sku_mappings?.find((m: any) => m.platform === platform);

                      return (
                        <tr key={item.guidFixed} className="hover:bg-muted/10">
                          <td className="p-3">
                            <div className="font-semibold text-foreground">{pickName(item.names, lang)}</div>
                            <div className="flex gap-2 items-center text-xs text-muted-foreground mt-0.5">
                              <span className="bg-muted px-1.5 py-0.5 rounded font-mono">{item.barcode || "ไม่มีบาร์โค้ด"}</span>
                              <span>รหัส: {item.itemCode || "—"}</span>
                            </div>
                          </td>
                          <td className="p-3 font-mono text-xs">
                            {item.raw?.sellersku || <span className="text-muted-foreground italic">—</span>}
                          </td>
                          <td className="p-3">
                            {platformMap ? (
                              <Badge className={cn(
                                "text-[10px]",
                                platform === "shopee" && "bg-[#ee4d2d]/10 text-[#ee4d2d] border-[#ee4d2d]/20",
                                platform === "lazada" && "bg-[#101566]/10 text-[#101566] border-[#101566]/20",
                                platform === "tiktok" && "bg-black/5 text-black border-black/10 dark:bg-white/10 dark:text-white"
                              )}>
                                {platformMap.holdingcode}
                              </Badge>
                            ) : (
                              <span className="text-[11px] text-muted-foreground italic">ไม่มีข้อมูลร้านค้า</span>
                            )}
                          </td>
                          <td className="p-3 font-mono text-xs">
                            {platformMap?.market_item_id || <span className="text-muted-foreground italic">—</span>}
                          </td>
                          <td className="p-3 font-mono text-xs">
                            {subMappings?.market_model_id || <span className="text-muted-foreground italic">—</span>}
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* 3. Sync Logs Tab */}
      {activeTab === "logs" && (
        <Card className="border-border">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-1.5">
              <Database className="h-4.5 w-4.5 text-primary" />
              รายงานประวัติการซิงค์ {platformName} (Sync Event Log)
            </CardTitle>
            <CardDescription className="text-xs">
              รายงานประวัติและสถานะการปรับปรุงปริมาณสต๊อกสินค้าและราคาขายล่าสุดไปยังช่องทาง {platformName} ในรอบ 24 ชั่วโมง
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 font-mono text-xs">
            <div className="space-y-1 text-muted-foreground">
              <div className="p-2 border-b border-border flex justify-between bg-muted/20">
                <span className="text-foreground">2026-05-29 11:25:03</span>
                <span className="text-green-600 font-bold">[SUCCESS]</span>
                <span className="truncate">{platformName} Sync: บาร์โค้ด 8850123456789 (ปรับสต๊อกเป็น 25 ชิ้น)</span>
              </div>
              <div className="p-2 border-b border-border flex justify-between">
                <span>2026-05-29 11:21:44</span>
                <span className="text-green-600 font-bold">[SUCCESS]</span>
                <span className="truncate">{platformName} Sync: บาร์โค้ด 8850123456789 (ปรับราคาเป็น 390.00 THB)</span>
              </div>
              <div className="p-2 border-b border-border flex justify-between bg-muted/20">
                <span>2026-05-29 10:45:12</span>
                <span className="text-destructive font-bold">[FAILED]</span>
                <span className="truncate text-destructive font-semibold">{platformName} Sync: บาร์โค้ด 885022211333 (ข้อผิดพลาด API: Invalid Seller SKU)</span>
              </div>
            </div>
            <div className="pt-2 text-center text-xs text-muted-foreground italic font-sans">
              * แสดงผลเฉพาะความเคลื่อนไหวซิงค์ล่าสุดของแพลตฟอร์ม {platformName}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Auth Setup Dialog */}
      {authDialog.open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <Card className="w-full max-w-md border-border bg-card shadow-2xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base font-bold flex items-center gap-2">
                  <CloudLightning className="text-primary h-5 w-5" />
                  เชื่อมโยงสิทธิ์ร้านค้า {platformName}
                </CardTitle>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 text-muted-foreground hover:text-foreground"
                  onClick={() => setAuthDialog({ open: false, holdingCode: "", shopName: "" })}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
              <CardDescription className="text-xs">
                กรอกรหัสร้านค้าของคุณจากระบบ {platformName} Seller Center เพื่อเริ่มต้นระบบซิงค์ข้อมูล
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <div className="space-y-1">
                <span className="font-semibold text-xs text-muted-foreground">รหัสร้านค้า (รหัสกลุ่มกิจการ) *</span>
                <Input
                  placeholder={`เช่น shop_${platform}_01`}
                  value={authDialog.holdingCode}
                  onChange={(e) => setAuthDialog(prev => ({ ...prev, holdingCode: e.target.value }))}
                />
              </div>
              <div className="space-y-1">
                <span className="font-semibold text-xs text-muted-foreground">ชื่อร้านค้าสำหรับแสดงผล</span>
                <Input
                  placeholder="เช่น บ้านเชียง แบรนด์ ออฟฟิเชียล"
                  value={authDialog.shopName}
                  onChange={(e) => setAuthDialog(prev => ({ ...prev, shopName: e.target.value }))}
                />
              </div>

              <div className="text-xs text-muted-foreground bg-primary/5 p-3 rounded-lg border border-primary/10">
                💡 การเชื่อมต่อผ่าน API ตรงอย่างเป็นทางการ จะต้องมี App Key และ Sign Token ซึ่งสามารถเปิดใช้งานผ่านพอร์ทัลผู้พัฒนา หรือใช้การจับคู่รหัสแบบ Import ไฟล์ด่วนได้ฟรี
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setAuthDialog({ open: false, holdingCode: "", shopName: "" })}
                >
                  ยกเลิก
                </Button>
                <Button
                  size="sm"
                  disabled={!authDialog.holdingCode}
                  onClick={handleConnectShop}
                >
                  บันทึกและเชื่อมต่อ
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

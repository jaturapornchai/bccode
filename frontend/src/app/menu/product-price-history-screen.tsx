"use client";

import { AlertCircle, Barcode, History, Loader2, RefreshCcw, Search } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { formatDefaultDateTime, resolveWorkspaceDateTimeDisplayOptions } from "@/lib/date-time";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import {
  localizedName,
  type AuthSession,
  type LocalizedName,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";

type ProductPriceHistoryScreenProps = {
  embedded?: boolean;
  language?: LanguageCode;
};

type ProductSummary = {
  guidFixed: string;
  barcode: string;
  name: string;
  itemCode: string;
  unitName: string;
  price: number;
};

type PriceHistoryRecord = {
  guidFixed: string;
  barcode: string;
  productName: string;
  priceType: string;
  keyNumber: number;
  oldPrice: number;
  newPrice: number;
  difference: number;
  action: string;
  createdBy: string;
  createdAt: string;
  remark: string;
};

type Notice = { type: "error" | "info" | "success"; text: string } | null;

const text = {
  th: {
    title: "ประวัติแก้ไขราคา",
    subtitle: "เลือกสินค้าจากฐานข้อมูลจริงเพื่อดูประวัติการปรับราคา",
    search: "ค้นหา บาร์โค้ด ชื่อสินค้า หรือรหัสสินค้า",
    products: "สินค้า",
    history: "ประวัติราคา",
    oldPrice: "ราคาเดิม",
    newPrice: "ราคาใหม่",
    difference: "ผลต่าง",
    action: "การทำรายการ",
    createdBy: "ผู้ทำรายการ",
    createdAt: "วันที่",
    refresh: "รีเฟรช",
    noProduct: "ไม่พบข้อมูลสินค้า",
    noHistory: "ไม่พบประวัติราคาของสินค้านี้",
    loading: "กำลังโหลดข้อมูล",
    apiRequired: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้",
  },
  en: {
    title: "Price Edit History",
    subtitle: "Select a real product to inspect price-change history.",
    search: "Search barcode, product name, or item code",
    products: "Products",
    history: "Price history",
    oldPrice: "Old price",
    newPrice: "New price",
    difference: "Difference",
    action: "Action",
    createdBy: "Created by",
    createdAt: "Date",
    refresh: "Refresh",
    noProduct: "No product data found",
    noHistory: "No price history found for this product",
    loading: "Loading data",
    apiRequired: "Please sign in and select a company before opening this screen.",
  },
} as const;

export function ProductPriceHistoryScreen({ embedded = false, language: externalLanguage }: ProductPriceHistoryScreenProps) {
  const [language, setLanguage] = useState<LanguageCode>(externalLanguage ?? "th");
  const dictionary = language === "th" ? text.th : text.en;
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [products, setProducts] = useState<ProductSummary[]>([]);
  const [selectedBarcode, setSelectedBarcode] = useState("");
  const [history, setHistory] = useState<PriceHistoryRecord[]>([]);
  const [query, setQuery] = useState("");
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);

  const selectedProduct = useMemo(
    () => products.find((product) => product.barcode === selectedBarcode) ?? products[0] ?? null,
    [products, selectedBarcode],
  );
  const dateTimeDisplayOptions = useMemo(
    () => resolveWorkspaceDateTimeDisplayOptions(workspace, language),
    [language, workspace],
  );

  const loadProducts = useCallback(async (currentAuth: AuthSession | null, currentWorkspace: WorkspaceSession | null, searchText: string) => {
    if (!currentAuth || !currentWorkspace) return;
    setLoadingProducts(true);
    setNotice(null);
    try {
      const response = await fetch("/api/product-barcode/list", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": currentAuth.backendUrl,
          Authorization: `Bearer ${currentAuth.token}`,
        },
        body: JSON.stringify({
          backendUrl: currentAuth.backendUrl,
          holdingcode: currentWorkspace.shop.holdingcode,
          keyword: searchText.trim(),
          limit: 80,
          offset: 0,
          sortfield: searchText.trim() ? "relevance" : "barcode",
          sortorder: "asc",
        }),
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? "load failed");
      const nextProducts = normalizeProducts(isRecord(payload) ? payload.data : payload);
      setProducts(nextProducts);
      setSelectedBarcode((current) => nextProducts.some((item) => item.barcode === current) ? current : nextProducts[0]?.barcode ?? "");
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : "load failed" });
      setProducts([]);
      setSelectedBarcode("");
    } finally {
      setLoadingProducts(false);
    }
  }, []);

  const loadHistory = useCallback(async (currentAuth: AuthSession | null, barcode: string) => {
    if (!currentAuth || !barcode) {
      setHistory([]);
      return;
    }
    setLoadingHistory(true);
    try {
      const response = await fetch(`/api/product-price-history/${encodeURIComponent(barcode)}?page=1&limit=100`, {
        headers: {
          "x-bc-backend-url": currentAuth.backendUrl,
          Authorization: `Bearer ${currentAuth.token}`,
        },
        cache: "no-store",
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? "load failed");
      setHistory(normalizeHistory(isRecord(payload) ? payload.data : payload));
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : "load failed" });
      setHistory([]);
    } finally {
      setLoadingHistory(false);
    }
  }, []);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? externalLanguage ?? "th");
    setLanguage(savedLanguage);
    const nextAuth = readAuth();
    const nextWorkspace = readWorkspace();
    setAuth(nextAuth);
    setWorkspace(nextWorkspace);
    if (!nextAuth || !nextWorkspace) {
      setNotice({ type: "info", text: dictionary.apiRequired });
      return;
    }
    void loadProducts(nextAuth, nextWorkspace, "");
  }, [dictionary.apiRequired, externalLanguage, loadProducts]);

  useEffect(() => {
    void loadHistory(auth, selectedProduct?.barcode ?? "");
  }, [auth, loadHistory, selectedProduct?.barcode]);

  const content = (
    <div className="grid w-full min-w-0 gap-3">
      <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
        <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase text-muted-foreground">BC Ai Account</p>
            <h1 className="truncate text-xl font-semibold sm:text-2xl">{dictionary.title}</h1>
            <p className="text-sm text-muted-foreground">{dictionary.subtitle}</p>
          </div>
          <Badge variant="outline">{products.length.toLocaleString()} {dictionary.products}</Badge>
        </div>
      </header>

      {notice ? (
        <div className={`message ${notice.type === "error" ? "error" : notice.type === "success" ? "success" : "info"}`}>
          <AlertCircle size={18} />
          <span>{notice.text}</span>
        </div>
      ) : null}

      <Card>
        <CardContent className="grid gap-2 p-3">
          <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto]">
            <label className="relative block min-w-0">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="!pl-10"
                placeholder={dictionary.search}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") void loadProducts(auth, workspace, query);
                }}
              />
            </label>
            <Button type="button" variant="outline" onClick={() => void loadProducts(auth, workspace, query)} disabled={loadingProducts || !auth}>
              {loadingProducts ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {dictionary.refresh}
            </Button>
          </div>
        </CardContent>
      </Card>

      <section className="grid min-w-0 gap-3 xl:grid-cols-[420px_minmax(0,1fr)]">
        <Card className="min-w-0">
          <CardHeader className="p-3 pb-1">
            <CardTitle className="text-base">{dictionary.products}</CardTitle>
          </CardHeader>
          <CardContent className="grid max-h-[65dvh] gap-2 overflow-auto p-3">
            {loadingProducts && !products.length ? (
              <div className="flex min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="animate-spin" /> {dictionary.loading}
              </div>
            ) : products.length ? products.map((product) => (
              <button
                className={cn(
                  "grid w-full min-w-0 gap-1 rounded-xl border border-border bg-background p-2 text-left transition hover:border-primary",
                  selectedProduct?.barcode === product.barcode && "border-primary bg-primary/5",
                )}
                key={product.guidFixed || product.barcode}
                onClick={() => setSelectedBarcode(product.barcode)}
                type="button"
              >
                <b className="truncate">{product.name || product.barcode}</b>
                <div className="grid gap-1 text-xs text-muted-foreground">
                  <span className="truncate"><Barcode className="mr-1 inline size-3" />{product.barcode || "-"}</span>
                  <span className="truncate">{product.itemCode || "-"} · {product.unitName || "-"}</span>
                </div>
              </button>
            )) : (
              <div className="grid min-h-32 place-items-center text-sm text-muted-foreground">{dictionary.noProduct}</div>
            )}
          </CardContent>
        </Card>

        <Card className="min-w-0">
          <CardHeader className="flex flex-row items-start justify-between gap-2 p-3 pb-1">
            <div className="min-w-0">
              <CardTitle className="truncate text-base">{selectedProduct?.name || dictionary.history}</CardTitle>
              <p className="truncate text-xs text-muted-foreground">{selectedProduct?.barcode || "-"}</p>
            </div>
            <Badge variant="outline"><History className="mr-1 size-3" />{history.length.toLocaleString()}</Badge>
          </CardHeader>
          <CardContent className="grid max-h-[65dvh] gap-2 overflow-auto p-3">
            {loadingHistory ? (
              <div className="flex min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="animate-spin" /> {dictionary.loading}
              </div>
            ) : history.length ? history.map((item) => (
              <div className="grid gap-2 rounded-xl border border-border bg-background p-2" key={item.guidFixed || `${item.barcode}-${item.createdAt}-${item.keyNumber}`}>
                <div className="flex min-w-0 flex-wrap items-center justify-between gap-2">
                  <b className="truncate">{item.priceType || item.action || "-"}</b>
                  <Badge variant={item.difference >= 0 ? "success" : "warning"}>{formatNumber(item.difference)}</Badge>
                </div>
                <div className="grid gap-1 text-xs text-muted-foreground sm:grid-cols-2 lg:grid-cols-3">
                  <span>{dictionary.oldPrice}: <b className="text-foreground">{formatNumber(item.oldPrice)}</b></span>
                  <span>{dictionary.newPrice}: <b className="text-foreground">{formatNumber(item.newPrice)}</b></span>
                  <span>{dictionary.createdBy}: <b className="text-foreground">{item.createdBy || "-"}</b></span>
                  <span>{dictionary.action}: <b className="text-foreground">{item.action || "-"}</b></span>
                  <span className="sm:col-span-2">{dictionary.createdAt}: <b className="text-foreground">{formatDefaultDateTime(item.createdAt, dateTimeDisplayOptions)}</b></span>
                </div>
                {item.remark ? <p className="text-xs text-muted-foreground">{item.remark}</p> : null}
              </div>
            )) : (
              <div className="grid min-h-32 place-items-center text-sm text-muted-foreground">{dictionary.noHistory}</div>
            )}
          </CardContent>
        </Card>
      </section>
    </div>
  );

  if (embedded) return <section className="grid w-full min-w-0 gap-3">{content}</section>;
  return <main className="min-h-dvh w-full bg-background p-3 text-foreground">{content}</main>;
}

function readAuth(): AuthSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.auth);
    if (!raw) return null;
    const auth = JSON.parse(raw) as AuthSession;
    return auth?.token && auth?.backendUrl ? auth : null;
  } catch {
    return null;
  }
}

function readWorkspace(): WorkspaceSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!raw) return null;
    const workspace = JSON.parse(raw) as WorkspaceSession;
    return workspace?.shop?.holdingcode ? workspace : null;
  } catch {
    return null;
  }
}

function normalizeProducts(value: unknown): ProductSummary[] {
  if (!Array.isArray(value)) return [];
  return value.map(normalizeProduct).filter((item): item is ProductSummary => Boolean(item?.barcode || item?.itemCode));
}

function normalizeProduct(value: unknown): ProductSummary | null {
  if (!isRecord(value)) return null;
  return {
    guidFixed: getString(value, "guidfixed") || getString(value, "guidFixed"),
    barcode: getString(value, "barcode"),
    name: localizedName(getArray(value, "names") as LocalizedName[], "th") || getString(value, "name") || getString(value, "productname"),
    itemCode: getString(value, "itemcode") || getString(value, "itemcode"),
    unitName: localizedName(getArray(value, "itemunitnames") as LocalizedName[], "th") || getString(value, "unitname"),
    price: getNumber(value, "prices") || getNumber(value, "price"),
  };
}

function normalizeHistory(value: unknown): PriceHistoryRecord[] {
  if (!Array.isArray(value)) return [];
  return value.map(normalizeHistoryRecord).filter((item): item is PriceHistoryRecord => Boolean(item));
}

function normalizeHistoryRecord(value: unknown): PriceHistoryRecord | null {
  if (!isRecord(value)) return null;
  return {
    guidFixed: getString(value, "guidfixed") || getString(value, "guidFixed"),
    barcode: getString(value, "barcode"),
    productName: getString(value, "productname") || getString(value, "productName"),
    priceType: getString(value, "pricetype") || getString(value, "priceType"),
    keyNumber: getNumber(value, "keynumber") || getNumber(value, "keyNumber"),
    oldPrice: getNumber(value, "oldprice") || getNumber(value, "oldPrice"),
    newPrice: getNumber(value, "newprice") || getNumber(value, "newPrice"),
    difference: getNumber(value, "pricedifference") || getNumber(value, "priceDifference"),
    action: getString(value, "action"),
    createdBy: getString(value, "createdby") || getString(value, "createdBy"),
    createdAt: getString(value, "createdat") || getString(value, "createdAt"),
    remark: getString(value, "remark"),
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function getString(record: Record<string, unknown>, key: string): string {
  const value = record[key];
  return typeof value === "string" ? value : value === null || value === undefined ? "" : String(value);
}

function getNumber(record: Record<string, unknown>, key: string): number {
  const value = record[key];
  const number = typeof value === "number" ? value : Number(value ?? 0);
  return Number.isFinite(number) ? number : 0;
}

function getArray(record: Record<string, unknown>, key: string): unknown[] {
  const value = record[key];
  return Array.isArray(value) ? value : [];
}

function isFailed(payload: unknown): boolean {
  return isRecord(payload) && payload.success === false;
}

function extractMessage(payload: unknown): string | undefined {
  if (!isRecord(payload)) return undefined;
  const message = payload.message ?? payload.error;
  return typeof message === "string" ? message : undefined;
}

function formatNumber(value: number): string {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}

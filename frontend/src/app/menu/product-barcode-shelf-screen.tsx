"use client";

import { AlertCircle, Barcode, Loader2, Minus, Plus, Printer, RefreshCcw, Search, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import {
  localizedName,
  type AuthSession,
  type LocalizedName,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";

type ProductBarcodeShelfScreenProps = {
  embedded?: boolean;
  language?: LanguageCode;
};

type ProductForLabel = {
  guidFixed: string;
  barcode: string;
  name: string;
  itemCode: string;
  unitName: string;
  price: number;
  shelfName: string;
};

type SelectedLabel = {
  product: ProductForLabel;
  copies: number;
};

type Notice = { type: "error" | "info" | "success"; text: string } | null;

const text = {
  th: {
    title: "พิมพ์ป้ายสินค้า",
    subtitle: "เลือกสินค้าจากฐานข้อมูลจริง กำหนดจำนวนป้าย แล้วพิมพ์ตามหน้าจอ Flutter เดิม",
    search: "ค้นหา บาร์โค้ด ชื่อสินค้า หรือรหัสสินค้า",
    products: "รายการสินค้า",
    selected: "สินค้าที่เลือก",
    total: "ทั้งหมด",
    copies: "จำนวนป้าย",
    addVisible: "เพิ่มที่แสดง",
    clear: "ล้าง",
    print: "พิมพ์",
    refresh: "รีเฟรช",
    noData: "ไม่พบข้อมูลสินค้าจากฐานข้อมูลจริง",
    noSelected: "ยังไม่ได้เลือกสินค้า",
    loading: "กำลังโหลดสินค้า",
    apiRequired: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้",
  },
  en: {
    title: "Print Product Label",
    subtitle: "Select real products, set label copies, and print using the legacy Flutter workflow.",
    search: "Search barcode, product name, or item code",
    products: "Products",
    selected: "Selected products",
    total: "Total",
    copies: "Copies",
    addVisible: "Add visible",
    clear: "Clear",
    print: "Print",
    refresh: "Refresh",
    noData: "No product data found in the real database",
    noSelected: "No product selected",
    loading: "Loading products",
    apiRequired: "Please sign in and select a company before opening this screen.",
  },
} as const;

export function ProductBarcodeShelfScreen({ embedded = false, language: externalLanguage }: ProductBarcodeShelfScreenProps) {
  const [language, setLanguage] = useState<LanguageCode>(externalLanguage ?? "th");
  const dictionary = language === "th" ? text.th : text.en;
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [products, setProducts] = useState<ProductForLabel[]>([]);
  const [selected, setSelected] = useState<Record<string, SelectedLabel>>({});
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);

  const loadProducts = useCallback(async (currentAuth: AuthSession | null, currentWorkspace: WorkspaceSession | null, searchText: string) => {
    if (!currentAuth || !currentWorkspace) return;
    setLoading(true);
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
      setProducts(normalizeProducts(isRecord(payload) ? payload.data : payload));
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : "load failed" });
      setProducts([]);
    } finally {
      setLoading(false);
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

  const selectedList = useMemo(() => Object.values(selected), [selected]);
  const totalCopies = selectedList.reduce((sum, item) => sum + item.copies, 0);

  function toggleProduct(product: ProductForLabel) {
    setSelected((current) => {
      const next = { ...current };
      if (next[product.barcode]) delete next[product.barcode];
      else next[product.barcode] = { product, copies: 1 };
      return next;
    });
  }

  function updateCopies(barcode: string, delta: number) {
    setSelected((current) => {
      const item = current[barcode];
      if (!item) return current;
      const copies = Math.max(1, item.copies + delta);
      return { ...current, [barcode]: { ...item, copies } };
    });
  }

  function addVisible() {
    setSelected((current) => {
      const next = { ...current };
      for (const product of products) {
        if (!next[product.barcode]) next[product.barcode] = { product, copies: 1 };
      }
      return next;
    });
  }

  function printLabels() {
    if (!selectedList.length) {
      setNotice({ type: "info", text: dictionary.noSelected });
      return;
    }
    document.body.classList.add("product-label-printing");
    window.setTimeout(() => document.body.classList.remove("product-label-printing"), 1000);
    window.print();
  }

  const content = (
    <div className="grid w-full min-w-0 gap-3">
      <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
        <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase text-muted-foreground">BC Ai Account</p>
            <h1 className="truncate text-xl font-semibold sm:text-2xl">{dictionary.title}</h1>
            <p className="text-sm text-muted-foreground">{dictionary.subtitle}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant="outline">{dictionary.total}: {products.length.toLocaleString()}</Badge>
            <Badge variant="success">{dictionary.selected}: {selectedList.length.toLocaleString()} / {totalCopies.toLocaleString()}</Badge>
          </div>
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
          <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto]">
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
            <Button type="button" variant="outline" onClick={() => void loadProducts(auth, workspace, query)} disabled={loading || !auth}>
              {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {dictionary.refresh}
            </Button>
            <Button type="button" variant="outline" onClick={addVisible} disabled={!products.length}>
              <Plus />
              {dictionary.addVisible}
            </Button>
            <Button type="button" onClick={printLabels} disabled={!selectedList.length}>
              <Printer />
              {dictionary.print}
            </Button>
          </div>
        </CardContent>
      </Card>

      <section className="grid min-w-0 gap-3 xl:grid-cols-[minmax(0,1fr)_420px]">
        <Card className="min-w-0">
          <CardHeader className="p-3 pb-1">
            <CardTitle className="text-base">{dictionary.products}</CardTitle>
          </CardHeader>
          <CardContent className="grid max-h-[62dvh] gap-2 overflow-auto p-3">
            {loading && !products.length ? (
              <div className="flex min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="animate-spin" /> {dictionary.loading}
              </div>
            ) : products.length ? products.map((product) => {
              const active = Boolean(selected[product.barcode]);
              return (
                <button
                  className={cn(
                    "grid w-full min-w-0 gap-1 rounded-xl border border-border bg-background p-2 text-left transition hover:border-primary",
                    active && "border-primary bg-primary/5",
                  )}
                  key={product.guidFixed || product.barcode}
                  onClick={() => toggleProduct(product)}
                  type="button"
                >
                  <div className="flex min-w-0 items-center justify-between gap-2">
                    <b className="truncate">{product.name || product.barcode}</b>
                    <Badge variant={active ? "success" : "outline"}>{product.price.toLocaleString()}</Badge>
                  </div>
                  <div className="grid gap-1 text-xs text-muted-foreground sm:grid-cols-3">
                    <span className="truncate"><Barcode className="mr-1 inline size-3" />{product.barcode || "-"}</span>
                    <span className="truncate">{product.itemCode || "-"}</span>
                    <span className="truncate">{product.unitName || "-"}</span>
                  </div>
                </button>
              );
            }) : (
              <div className="grid min-h-32 place-items-center text-sm text-muted-foreground">{dictionary.noData}</div>
            )}
          </CardContent>
        </Card>

        <Card className="min-w-0">
          <CardHeader className="flex flex-row items-center justify-between p-3 pb-1">
            <CardTitle className="text-base">{dictionary.selected}</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setSelected({})} disabled={!selectedList.length}>
              <Trash2 /> {dictionary.clear}
            </Button>
          </CardHeader>
          <CardContent className="grid max-h-[62dvh] gap-2 overflow-auto p-3">
            {selectedList.length ? selectedList.map(({ product, copies }) => (
              <div className="grid gap-2 rounded-xl border border-border bg-background p-2" key={product.barcode}>
                <div className="min-w-0">
                  <b className="block truncate">{product.name || product.barcode}</b>
                  <p className="truncate text-xs text-muted-foreground">{product.barcode} · {product.itemCode || "-"}</p>
                </div>
                <div className="flex items-center justify-between gap-2">
                  <span className="text-sm text-muted-foreground">{dictionary.copies}</span>
                  <div className="flex items-center gap-1">
                    <Button type="button" variant="outline" size="icon" onClick={() => updateCopies(product.barcode, -1)}><Minus /></Button>
                    <b className="min-w-10 text-center">{copies}</b>
                    <Button type="button" variant="outline" size="icon" onClick={() => updateCopies(product.barcode, 1)}><Plus /></Button>
                  </div>
                </div>
              </div>
            )) : (
              <div className="grid min-h-32 place-items-center text-sm text-muted-foreground">{dictionary.noSelected}</div>
            )}
          </CardContent>
        </Card>
      </section>

      <div className="print-only">
        <div className="label-sheet">
          {selectedList.flatMap(({ product, copies }) =>
            Array.from({ length: copies }, (_, index) => (
              <div className="print-label" key={`${product.barcode}-${index}`}>
                <b>{product.name || product.barcode}</b>
                <span>{product.itemCode || product.barcode}</span>
                <svg className="label-bars" viewBox="0 0 120 32" aria-hidden="true">
                  {Array.from(product.barcode || "000000").map((char, i) => (
                    <rect height="32" key={`${char}-${i}`} width={i % 3 === 0 ? 3 : 1} x={i * 6} y="0" />
                  ))}
                </svg>
                <span>{product.barcode}</span>
                {product.price ? <strong>{product.price.toLocaleString()}.-</strong> : null}
              </div>
            )),
          )}
        </div>
      </div>
    </div>
  );

  if (embedded) return <section className="grid w-full min-w-0 gap-3 product-label-screen">{content}</section>;
  return <main className="min-h-dvh w-full bg-background p-3 text-foreground product-label-screen">{content}</main>;
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

function normalizeProducts(value: unknown): ProductForLabel[] {
  if (!Array.isArray(value)) return [];
  return value.map(normalizeProduct).filter((item): item is ProductForLabel => Boolean(item?.barcode || item?.itemCode));
}

function normalizeProduct(value: unknown): ProductForLabel | null {
  if (!isRecord(value)) return null;
  return {
    guidFixed: getString(value, "guidfixed") || getString(value, "guidFixed"),
    barcode: getString(value, "barcode"),
    name: localizedName(getArray(value, "names") as LocalizedName[], "th") || getString(value, "name") || getString(value, "productname"),
    itemCode: getString(value, "itemcode") || getString(value, "itemcode"),
    unitName: localizedName(getArray(value, "itemunitnames") as LocalizedName[], "th") || getString(value, "unitname"),
    price: getNumber(value, "prices") || getNumber(value, "price"),
    shelfName: getString(value, "shelfname") || getString(value, "shelf_name"),
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

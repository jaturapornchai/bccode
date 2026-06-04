"use client";

import dynamic from "next/dynamic";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  Loader2,
  LocateFixed,
  MapPin,
  Search,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { LanguageCode } from "@/lib/i18n";

const BANGKOK_LAT = 13.7563;
const BANGKOK_LNG = 100.5018;
const DEFAULT_ZOOM = 16;
const FALLBACK_ZOOM = 11;
const SEARCH_ZOOM = 17;
const NOMINATIM_URL = "https://nominatim.openstreetmap.org/search";

type Locale = LanguageCode;

const mapPickerText = {
  title: {
    th: "เลือกตำแหน่งจากแผนที่",
    en: "Pick a location",
    cn: "在地图上选择位置",
    ja: "地図で位置を選択",
    ko: "지도에서 위치 선택",
    lo: "ເລືອກຕຳແໜ່ງຈາກແຜນທີ່",
    my: "မြေပုံတွင် တည်နေရာ ရွေးချယ်ပါ",
    km: "ជ្រើសរើសទីតាំងពីផែនទី",
    vi: "Chọn vị trí trên bản đồ",
    ms: "Pilih lokasi pada peta",
    id: "Pilih lokasi di peta",
    fil: "Pumili ng lokasyon sa mapa",
  },
  use: {
    th: "ใช้ตำแหน่งนี้",
    en: "Use this location",
    cn: "使用此位置",
    ja: "この位置を使用",
    ko: "이 위치 사용",
    lo: "ໃຊ້ຕຳແໜ່ງນີ້",
    my: "ဤတည်နေရာကို သုံးမည်",
    km: "ប្រើទីតាំងនេះ",
    vi: "Dùng vị trí này",
    ms: "Guna lokasi ini",
    id: "Gunakan lokasi ini",
    fil: "Gamitin ang lokasyon na ito",
  },
  cancel: {
    th: "ยกเลิก",
    en: "Cancel",
    cn: "取消",
    ja: "キャンセル",
    ko: "취소",
    lo: "ຍົກເລີກ",
    my: "ပယ်ဖျက်",
    km: "បោះបង់",
    vi: "Huỷ",
    ms: "Batal",
    id: "Batal",
    fil: "Kanselahin",
  },
  hint: {
    th: "แตะที่แผนที่เพื่อปักหมุด",
    en: "Tap the map to drop a pin",
    cn: "点击地图放置标记",
    ja: "地図をタップしてピンを置く",
    ko: "지도를 탭하여 핀 추가",
    lo: "ແຕະທີ່ແຜນທີ່ເພື່ອປັກໝຸດ",
    my: "မြေပုံကို တို့ပြီး အမှတ်အသား ထည့်ပါ",
    km: "ប៉ះផែនទីដើម្បីដាក់ម្ជុល",
    vi: "Nhấn vào bản đồ để đặt ghim",
    ms: "Ketik peta untuk meletakkan pin",
    id: "Ketuk peta untuk menjatuhkan pin",
    fil: "I-tap ang mapa para maglagay ng pin",
  },
  search: {
    th: "ค้นหาสถานที่ ที่อยู่ หรือเขต",
    en: "Search place, address, or area",
    cn: "搜索地点、地址或区域",
    ja: "場所・住所・エリアを検索",
    ko: "장소, 주소 또는 지역 검색",
    lo: "ຄົ້ນຫາສະຖານທີ່ ທີ່ຢູ່ ຫຼື ເຂດ",
    my: "နေရာ၊ လိပ်စာ၊ ဧရိယာ ရှာဖွေပါ",
    km: "ស្វែងរកទីកន្លែង អាសយដ្ឋាន ឬតំបន់",
    vi: "Tìm địa điểm, địa chỉ, khu vực",
    ms: "Cari tempat, alamat atau kawasan",
    id: "Cari tempat, alamat, atau wilayah",
    fil: "Hanapin ang lugar, address, o lugar",
  },
  searchAction: {
    th: "ค้นหา",
    en: "Search",
    cn: "搜索",
    ja: "検索",
    ko: "검색",
    lo: "ຄົ້ນຫາ",
    my: "ရှာဖွေ",
    km: "ស្វែងរក",
    vi: "Tìm",
    ms: "Cari",
    id: "Cari",
    fil: "Hanapin",
  },
  locate: {
    th: "ตำแหน่งปัจจุบัน",
    en: "Use my current location",
    cn: "使用当前位置",
    ja: "現在地を使用",
    ko: "현재 위치 사용",
    lo: "ໃຊ້ຕຳແໜ່ງປັດຈຸບັນ",
    my: "လက်ရှိ တည်နေရာကို သုံးပါ",
    km: "ប្រើទីតាំងបច្ចុប្បន្ន",
    vi: "Vị trí hiện tại của tôi",
    ms: "Guna lokasi semasa",
    id: "Gunakan lokasi saat ini",
    fil: "Gamitin ang kasalukuyang lokasyon",
  },
  searching: {
    th: "กำลังค้นหา…",
    en: "Searching…",
    cn: "搜索中…",
    ja: "検索中…",
    ko: "검색 중…",
    lo: "ກຳລັງຄົ້ນຫາ…",
    my: "ရှာဖွေနေသည်…",
    km: "កំពុងស្វែងរក…",
    vi: "Đang tìm…",
    ms: "Sedang mencari…",
    id: "Mencari…",
    fil: "Naghahanap…",
  },
  noResults: {
    th: "ไม่พบผลลัพธ์",
    en: "No results",
    cn: "无结果",
    ja: "結果なし",
    ko: "결과 없음",
    lo: "ບໍ່ມີຜົນ",
    my: "ရလဒ် မရှိပါ",
    km: "គ្មានលទ្ធផល",
    vi: "Không có kết quả",
    ms: "Tiada keputusan",
    id: "Tidak ada hasil",
    fil: "Walang resulta",
  },
  searchFailed: {
    th: "ค้นหาไม่สำเร็จ",
    en: "Search failed",
    cn: "搜索失败",
    ja: "検索に失敗しました",
    ko: "검색 실패",
    lo: "ຄົ້ນຫາລົ້ມເຫລວ",
    my: "ရှာဖွေမှု မအောင်မြင်ပါ",
    km: "ការស្វែងរកបរាជ័យ",
    vi: "Tìm kiếm thất bại",
    ms: "Carian gagal",
    id: "Pencarian gagal",
    fil: "Nabigo ang paghahanap",
  },
  locateUnsupported: {
    th: "เบราว์เซอร์ไม่รองรับการระบุตำแหน่ง",
    en: "Geolocation is not supported in this browser",
    cn: "此浏览器不支持地理位置",
    ja: "このブラウザでは位置情報を取得できません",
    ko: "이 브라우저는 위치 정보를 지원하지 않습니다",
    lo: "ບຣາວເຊີນີ້ບໍ່ຮອງຮັບການລະບຸຕຳແໜ່ງ",
    my: "ဤဘရောက်ဆာသည် တည်နေရာသိရှိမှုကို မထောက်ပံ့ပါ",
    km: "កម្មវិធីរុករកនេះមិនគាំទ្រការកំណត់ទីតាំងទេ",
    vi: "Trình duyệt không hỗ trợ định vị",
    ms: "Pelayar tidak menyokong lokasi",
    id: "Browser tidak mendukung lokasi",
    fil: "Hindi sinusuportahan ang geolocation",
  },
  locateDenied: {
    th: "ไม่สามารถใช้ตำแหน่งปัจจุบันได้",
    en: "Could not get current location",
    cn: "无法获取当前位置",
    ja: "現在地を取得できませんでした",
    ko: "현재 위치를 가져올 수 없습니다",
    lo: "ບໍ່ສາມາດເອົາຕຳແໜ່ງປັດຈຸບັນ",
    my: "လက်ရှိတည်နေရာ မရရှိနိုင်ပါ",
    km: "មិនអាចទទួលបានទីតាំងបច្ចុប្បន្ន",
    vi: "Không thể lấy vị trí hiện tại",
    ms: "Tidak dapat lokasi semasa",
    id: "Tidak bisa mendapatkan lokasi saat ini",
    fil: "Hindi makuha ang kasalukuyang lokasyon",
  },
} satisfies Record<string, Record<Locale, string>>;

type MapTextKey = keyof typeof mapPickerText;
function text(language: Locale, key: MapTextKey): string {
  return mapPickerText[key][language] ?? mapPickerText[key].en;
}

const LeafletMapPicker = dynamic(() => import("./leaflet-map-picker"), {
  ssr: false,
  loading: () => (
    <div className="grid h-full w-full place-items-center bg-muted/40 text-muted-foreground">
      <div className="flex items-center gap-2 text-sm">
        <Loader2 className="size-4 animate-spin" />
        Loading map…
      </div>
    </div>
  ),
});

type SearchResult = {
  lat: number;
  lng: number;
  label: string;
};

export type MapPickerDialogProps = {
  open: boolean;
  initialLat?: number | null;
  initialLng?: number | null;
  language: Locale;
  onCancel: () => void;
  onSelect: (lat: number, lng: number) => void;
};

export function MapPickerDialog({
  open,
  initialLat,
  initialLng,
  language,
  onCancel,
  onSelect,
}: MapPickerDialogProps) {
  const startLat = isFiniteCoordinate(initialLat) ? initialLat! : BANGKOK_LAT;
  const startLng = isFiniteCoordinate(initialLng) ? initialLng! : BANGKOK_LNG;
  const startZoom =
    isFiniteCoordinate(initialLat) && isFiniteCoordinate(initialLng)
      ? DEFAULT_ZOOM
      : FALLBACK_ZOOM;
  const [position, setPosition] = useState<{
    lat: number;
    lng: number;
    zoom: number;
  }>({ lat: startLat, lng: startLng, zoom: startZoom });
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [searchStatus, setSearchStatus] = useState<
    "idle" | "loading" | "no-results" | "failed"
  >("idle");
  const [locating, setLocating] = useState(false);
  const [locateError, setLocateError] = useState("");
  const searchControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    if (!open) return;
    setPosition({ lat: startLat, lng: startLng, zoom: startZoom });
    setSearchQuery("");
    setSearchResults([]);
    setSearchStatus("idle");
    setLocateError("");
    setLocating(false);
  }, [open, startLat, startLng, startZoom]);

  useEffect(() => {
    if (!open) return;
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onCancel, open]);

  useEffect(() => {
    return () => {
      searchControllerRef.current?.abort();
    };
  }, []);

  const runSearch = useCallback(async () => {
    const query = searchQuery.trim();
    if (!query) {
      setSearchResults([]);
      setSearchStatus("idle");
      return;
    }
    searchControllerRef.current?.abort();
    const controller = new AbortController();
    searchControllerRef.current = controller;
    setSearchStatus("loading");
    setLocateError("");
    try {
      const params = new URLSearchParams({
        q: query,
        format: "json",
        limit: "8",
        addressdetails: "0",
        "accept-language": language === "en" ? "en" : `${language},en`,
      });
      const response = await fetch(`${NOMINATIM_URL}?${params.toString()}`, {
        headers: { Accept: "application/json" },
        signal: controller.signal,
        cache: "no-store",
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const payload = (await response.json()) as Array<{
        lat: string;
        lon: string;
        display_name: string;
      }>;
      const results = payload
        .map((item) => ({
          lat: Number(item.lat),
          lng: Number(item.lon),
          label: item.display_name,
        }))
        .filter((item) => Number.isFinite(item.lat) && Number.isFinite(item.lng));
      setSearchResults(results);
      setSearchStatus(results.length ? "idle" : "no-results");
    } catch (error) {
      if ((error as { name?: string }).name === "AbortError") return;
      setSearchResults([]);
      setSearchStatus("failed");
    }
  }, [language, searchQuery]);

  const useCurrentLocation = useCallback(() => {
    if (typeof navigator === "undefined" || !navigator.geolocation) {
      setLocateError(text(language, "locateUnsupported"));
      return;
    }
    setLocating(true);
    setLocateError("");
    navigator.geolocation.getCurrentPosition(
      (geo) => {
        setPosition({
          lat: geo.coords.latitude,
          lng: geo.coords.longitude,
          zoom: SEARCH_ZOOM,
        });
        setLocating(false);
      },
      () => {
        setLocateError(text(language, "locateDenied"));
        setLocating(false);
      },
      { enableHighAccuracy: true, maximumAge: 60_000, timeout: 15_000 },
    );
  }, [language]);

  const handleResultPick = (result: SearchResult) => {
    setPosition({ lat: result.lat, lng: result.lng, zoom: SEARCH_ZOOM });
    setSearchResults([]);
  };

  useEffect(() => {
    if (!open) return;
    const trimmed = searchQuery.trim();
    if (trimmed.length < 3) {
      setSearchResults([]);
      setSearchStatus("idle");
      return;
    }
    const id = window.setTimeout(() => {
      void runSearch();
    }, 600);
    return () => window.clearTimeout(id);
  }, [open, runSearch, searchQuery]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col bg-card text-card-foreground"
      role="dialog"
      aria-modal="true"
      aria-label={text(language, "title")}
    >
      <header className="flex items-center justify-between gap-2 border-b border-border px-3 py-2">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <MapPin className="size-4" />
          {text(language, "title")}
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCancel}
          aria-label={text(language, "cancel")}
        >
          <X />
        </Button>
      </header>
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-3 py-2">
        <div
          className="relative flex min-w-0 flex-1 items-center gap-2"
          role="search"
        >
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="h-9 !pl-10"
            placeholder={text(language, "search")}
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                void runSearch();
              }
            }}
          />
          <Button
            type="button"
            size="sm"
            onClick={() => void runSearch()}
            disabled={searchStatus === "loading" || !searchQuery.trim()}
          >
            {searchStatus === "loading" ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Search />
            )}
            {text(language, "searchAction")}
          </Button>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={useCurrentLocation}
          disabled={locating}
        >
          {locating ? <Loader2 className="animate-spin" /> : <LocateFixed />}
          {text(language, "locate")}
        </Button>
      </div>
      <div className="relative min-h-0 flex-1">
        <LeafletMapPicker
          lat={position.lat}
          lng={position.lng}
          zoom={position.zoom}
          onChange={(lat, lng) =>
            setPosition((prev) => ({ lat, lng, zoom: prev.zoom }))
          }
        />
        {searchResults.length > 0 ? (
          <div className="absolute left-3 top-3 z-[401] w-[min(420px,calc(100vw-24px))] overflow-hidden rounded-xl border border-border bg-card shadow-lg">
            <ul className="max-h-64 overflow-y-auto text-sm">
              {searchResults.map((result, index) => (
                <li key={`${result.lat}-${result.lng}-${index}`}>
                  <button
                    type="button"
                    className="block w-full px-3 py-2 text-left hover:bg-muted"
                    onClick={() => handleResultPick(result)}
                  >
                    <span className="line-clamp-2 break-words">
                      {result.label}
                    </span>
                    <span className="block font-mono text-xs text-muted-foreground">
                      {result.lat.toFixed(6)}, {result.lng.toFixed(6)}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        ) : null}
      </div>
      <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border px-3 py-2 text-xs">
        <div className="flex flex-wrap items-center gap-3 text-muted-foreground">
          <span>{text(language, "hint")}</span>
          {searchStatus === "no-results" ? (
            <span className="text-amber-600 dark:text-amber-400">
              {text(language, "noResults")}
            </span>
          ) : null}
          {searchStatus === "failed" ? (
            <span className="text-destructive">
              {text(language, "searchFailed")}
            </span>
          ) : null}
          {locateError ? (
            <span className="text-destructive">{locateError}</span>
          ) : null}
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <span className="font-mono">
            {position.lat.toFixed(6)}, {position.lng.toFixed(6)}
          </span>
          <Button type="button" variant="outline" onClick={onCancel}>
            {text(language, "cancel")}
          </Button>
          <Button
            type="button"
            onClick={() => onSelect(position.lat, position.lng)}
          >
            <MapPin />
            {text(language, "use")}
          </Button>
        </div>
      </footer>
    </div>
  );
}

function isFiniteCoordinate(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value) && value !== 0;
}

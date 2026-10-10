"use client";

import { useMemo, useRef, useState } from "react";
import {
  Building2,
  CheckCircle2,
  Compass,
  Map,
  MapPin,
  Navigation,
  Network,
  Plus,
  RotateCcw,
  Save,
  Search,
  Sliders,
  Truck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface LogisticsHubItem {
  id: string;
  code: string;
  nameTh: string;
  nameEn: string;
  region: string;
  latitude: string;
  longitude: string;
  serviceRadiusKm: number;
  baseDeliveryFee: string;
  address: string;
  contactPhone: string;
  isActive: boolean;
}

const DEFAULT_HUBS: LogisticsHubItem[] = [
  {
    id: "hub-01",
    code: "HUB-01",
    nameTh: "ศูนย์กระจายสินค้าหลักกรุงเทพฯ-ปริมณฑล (ลาดกระบัง)",
    nameEn: "Bangkok Central Hub (Lat Krabang)",
    region: "กรุงเทพฯ และปริมณฑล",
    latitude: "13.7290",
    longitude: "100.7780",
    serviceRadiusKm: 80,
    baseDeliveryFee: "150.00",
    address: "99/1 ถนนร่มเกล้า แขวงคลองสามประเวศ เขตลาดกระบัง กรุงเทพมหานคร 10520",
    contactPhone: "02-329-8800",
    isActive: true,
  },
  {
    id: "hub-02",
    code: "HUB-02",
    nameTh: "ศูนย์จัดส่งและกระจายสินค้าภาคเหนือ (เชียงใหม่)",
    nameEn: "Northern Logistics Hub (Chiang Mai)",
    region: "ภาคเหนือ",
    latitude: "18.7883",
    longitude: "98.9853",
    serviceRadiusKm: 120,
    baseDeliveryFee: "250.00",
    address: "124 หมู่ 4 ถนนซุปเปอร์ไฮเวย์ ตำบลช้างเผือก อำเภอเมืองเชียงใหม่ เชียงใหม่ 50300",
    contactPhone: "053-214-550",
    isActive: true,
  },
  {
    id: "hub-03",
    code: "HUB-03",
    nameTh: "ศูนย์จัดส่งและกระจายสินค้าภาคอีสาน (ขอนแก่น)",
    nameEn: "Northeastern Hub (Khon Kaen)",
    region: "ภาคตะวันออกเฉียงเหนือ",
    latitude: "16.4419",
    longitude: "102.8360",
    serviceRadiusKm: 150,
    baseDeliveryFee: "250.00",
    address: "88 ถนนมิตรภาพ ตำบลในเมือง อำเภอเมืองขอนแก่น ขอนแก่น 40000",
    contactPhone: "043-332-110",
    isActive: true,
  },
  {
    id: "hub-04",
    code: "HUB-04",
    nameTh: "ศูนย์จัดส่งและกระจายสินค้าภาคใต้ (สุราษฎร์ธานี)",
    nameEn: "Southern Logistics Hub (Surat Thani)",
    region: "ภาคใต้",
    latitude: "9.1382",
    longitude: "99.3215",
    serviceRadiusKm: 140,
    baseDeliveryFee: "300.00",
    address: "45/2 ถนนเลี่ยงเมือง ตำบลมะขามเตี้ย อำเภอเมืองสุราษฎร์ธานี สุราษฎร์ธานี 84000",
    contactPhone: "077-285-440",
    isActive: true,
  },
];

export interface SystemMainMapScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemMainMapScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemMainMapScreenProps) {
  const isTh = language === "th";
  const [hubs, setHubs] = useState<LogisticsHubItem[]>(DEFAULT_HUBS);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_HUBS[0].id);
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 38,
    min: 25,
    max: 55,
    storageKey: "bcai_split_main_map",
  });

  const selectedHub = useMemo(() => {
    return hubs.find((h) => h.id === selectedId) || hubs[0];
  }, [hubs, selectedId]);

  const handleUpdate = <K extends keyof LogisticsHubItem>(
    field: K,
    val: LogisticsHubItem[K]
  ) => {
    setHubs((prev) =>
      prev.map((h) => (h.id === selectedHub.id ? { ...h, [field]: val } : h))
    );
  };

  const handleAddNewHub = () => {
    const newHub: LogisticsHubItem = {
      id: `hub-${Date.now()}`,
      code: `HUB-0${hubs.length + 1}`,
      nameTh: "ศูนย์กระจายสินค้าใหม่",
      nameEn: "New Logistics Hub",
      region: "ภาคกลาง",
      latitude: "13.7563",
      longitude: "100.5018",
      serviceRadiusKm: 50,
      baseDeliveryFee: "150.00",
      address: "ที่ตั้งศูนย์กระจายสินค้า",
      contactPhone: "02-000-0000",
      isActive: true,
    };
    setHubs((prev) => [...prev, newHub]);
    setSelectedId(newHub.id);
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? `บันทึกข้อมูลแผนที่หลักและพิกัดศูนย์กระจายสินค้า (${selectedHub.code} - ${selectedHub.nameTh}) เรียบร้อยแล้ว`
        : `Hub location (${selectedHub.code}) saved successfully.`
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setHubs(DEFAULT_HUBS);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset hubs to defaults");
    setTimeout(() => setNotification(null), 3000);
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <Map className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "กำหนดแผนที่หลัก" : "Main Logistics Map Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดพิกัดทางภูมิศาสตร์ (GPS), รัศมีการบริการ และเส้นทางศูนย์กระจายสินค้า"
                : "System Settings › Setup geographical coordinates, delivery service radius and logistics hub network"}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={handleReset}
            className="h-8 gap-1.5 text-xs text-slate-700 dark:text-slate-300"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            {isTh ? "คืนค่าเริ่มต้น" : "Reset Default"}
          </Button>
          <Button
            size="sm"
            onClick={handleSave}
            className="h-8 gap-1.5 text-xs bg-indigo-600 hover:bg-indigo-700 text-white"
          >
            <Save className="w-3.5 h-3.5" />
            {isTh ? "บันทึกการตั้งค่า" : "Save Settings"}
          </Button>
        </div>
      </div>

      {/* Notification Banner */}
      {notification && (
        <div className="flex-none bg-emerald-50 dark:bg-emerald-950/50 border-b border-emerald-200 dark:border-emerald-800 px-4 py-2 flex items-center gap-2 text-emerald-800 dark:text-emerald-300 text-xs animate-in fade-in duration-200">
          <CheckCircle2 className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
          <span>{notification}</span>
        </div>
      )}

      {/* Main Split Content */}
      <div ref={containerRef} className="flex-1 flex overflow-hidden relative">
        {/* Left Column: Hubs List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
        >
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
            <span className="font-semibold text-slate-700 dark:text-slate-200">
              {isTh ? "ศูนย์กระจายสินค้าทั้งหมด" : "Logistics Hubs"} ({hubs.length})
            </span>
            <Button
              size="sm"
              variant="ghost"
              onClick={handleAddNewHub}
              className="h-7 text-xs gap-1 text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50"
            >
              <Plus className="w-3.5 h-3.5" />
              {isTh ? "เพิ่มศูนย์" : "Add"}
            </Button>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
            {hubs.map((hub) => {
              const isSelected = hub.id === selectedHub.id;
              return (
                <div
                  key={hub.id}
                  onClick={() => setSelectedId(hub.id)}
                  className={`p-3.5 cursor-pointer transition-colors ${
                    isSelected
                      ? "bg-indigo-50/80 dark:bg-indigo-950/40 border-l-4 border-indigo-600"
                      : "hover:bg-slate-50 dark:hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-sm">
                      {hub.code}
                    </span>
                    <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                      {hub.region}
                    </span>
                  </div>
                  <p className="text-slate-800 dark:text-slate-200 font-medium">{hub.nameTh}</p>
                  <div className="mt-2 flex items-center justify-between text-[11px] text-slate-500">
                    <span className="flex items-center gap-1 font-mono text-[10px] text-slate-400">
                      <Navigation className="w-3 h-3 text-indigo-500" />
                      {hub.latitude}, {hub.longitude}
                    </span>
                    <span className="font-mono font-semibold text-emerald-600 dark:text-emerald-400">
                      {hub.serviceRadiusKm} {isTh ? "กม." : "km"}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Resizable Splitter Handle */}
        <ResizableSplitter
          onMouseDown={startResize}
          onKeyDown={adjustWithKeyboard}
          isResizing={isResizing}
        />

        {/* Right Column: Hub Profile & Visual Map Grid */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="h-full flex flex-col bg-slate-50/50 dark:bg-slate-950 overflow-y-auto p-4 gap-4"
        >
          {/* Coordinates Overview Banner */}
          <div className="bg-gradient-to-r from-indigo-900 to-slate-900 text-white p-5 rounded-xl shadow-md border border-indigo-800/60">
            <div className="flex items-center justify-between mb-2">
              <span className="text-indigo-200 font-medium flex items-center gap-1.5 text-xs">
                <MapPin className="w-4 h-4 text-rose-400" />
                {isTh ? "พิกัดและข้อมูลศูนย์จัดส่ง" : "Hub GPS Coordinates & Range"}
              </span>
              <span className="text-[11px] font-mono text-indigo-300 bg-white/10 px-2.5 py-0.5 rounded-full">
                {selectedHub.code} • {selectedHub.region}
              </span>
            </div>
            <div className="py-1">
              <h2 className="text-xl font-bold text-amber-300">{selectedHub.nameTh}</h2>
              <p className="text-xs text-indigo-200 mt-1">{selectedHub.address}</p>
            </div>
            <div className="mt-3 pt-3 border-t border-indigo-800/60 flex flex-wrap items-center justify-between gap-2 text-xs text-indigo-200">
              <div className="flex items-center gap-4">
                <span>
                  {isTh ? "ละติจูด (Lat): " : "Lat: "}
                  <strong className="text-white font-mono">{selectedHub.latitude}</strong>
                </span>
                <span>
                  {isTh ? "ลองจิจูด (Lng): " : "Lng: "}
                  <strong className="text-white font-mono">{selectedHub.longitude}</strong>
                </span>
              </div>
              <div>
                <span>{isTh ? "เบอร์ติดต่อ: " : "Phone: "}</span>
                <strong className="text-white font-mono">{selectedHub.contactPhone}</strong>
              </div>
            </div>
          </div>

          {/* Form Card */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-xs">
            <div className="flex items-center gap-2 pb-3 mb-4 border-b border-slate-100 dark:border-slate-800">
              <Sliders className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
              <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
                {isTh ? "ตั้งค่าพิกัดและขอบเขตการจัดส่ง" : "Hub Coordinates & Service Scope"}
              </h3>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รหัสศูนย์ (Hub Code)" : "Hub Code"}
                </label>
                <Input
                  value={selectedHub.code}
                  onChange={(e) => handleUpdate("code", e.target.value.toUpperCase())}
                  className="font-mono uppercase font-bold text-indigo-600 dark:text-indigo-400"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ภูมิภาค / โซนหลัก" : "Region / Territory"}
                </label>
                <Input
                  value={selectedHub.region}
                  onChange={(e) => handleUpdate("region", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อศูนย์ภาษาไทย" : "Thai Hub Name"}
                </label>
                <Input
                  value={selectedHub.nameTh}
                  onChange={(e) => handleUpdate("nameTh", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อศูนย์ภาษาอังกฤษ" : "English Hub Name"}
                </label>
                <Input
                  value={selectedHub.nameEn}
                  onChange={(e) => handleUpdate("nameEn", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ละติจูด (Latitude)" : "Latitude"}
                </label>
                <Input
                  value={selectedHub.latitude}
                  onChange={(e) => handleUpdate("latitude", e.target.value)}
                  className="font-mono"
                  placeholder="13.7290"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ลองจิจูด (Longitude)" : "Longitude"}
                </label>
                <Input
                  value={selectedHub.longitude}
                  onChange={(e) => handleUpdate("longitude", e.target.value)}
                  className="font-mono"
                  placeholder="100.7780"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รัศมีการบริการจัดส่ง (กิโลเมตร)" : "Service Radius (KM)"}
                </label>
                <Input
                  type="number"
                  min={1}
                  value={selectedHub.serviceRadiusKm}
                  onChange={(e) => handleUpdate("serviceRadiusKm", Math.max(1, parseInt(e.target.value) || 1))}
                  className="font-mono"
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ค่าจัดส่งเริ่มต้น (บาท)" : "Base Delivery Fee (THB)"}
                </label>
                <Input
                  type="number"
                  step="0.01"
                  min={0}
                  value={selectedHub.baseDeliveryFee}
                  onChange={(e) => handleUpdate("baseDeliveryFee", e.target.value)}
                  className="font-mono"
                />
              </div>

              <div className="sm:col-span-2">
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ที่อยู่ศูนย์กระจายสินค้า" : "Hub Address"}
                </label>
                <Input
                  value={selectedHub.address}
                  onChange={(e) => handleUpdate("address", e.target.value)}
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

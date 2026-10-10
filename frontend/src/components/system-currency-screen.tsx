"use client";

import { useMemo, useRef, useState } from "react";
import {
  Banknote,
  CheckCircle2,
  Coins,
  DollarSign,
  Globe,
  Plus,
  RotateCcw,
  Save,
  Search,
  Sliders,
  TrendingUp,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";

export interface ExchangeRateEntry {
  id: string;
  effectiveDate: string;
  buyRate: string;
  sellRate: string;
  accountRate: string; // Rate used for accounting conversion
}

export interface CurrencyItem {
  id: string;
  code: string;
  nameTh: string;
  nameEn: string;
  symbol: string;
  decimalPlaces: number;
  isBaseCurrency: boolean;
  isActive: boolean;
  rates: ExchangeRateEntry[];
}

const DEFAULT_CURRENCIES: CurrencyItem[] = [
  {
    id: "cur-thb",
    code: "THB",
    nameTh: "บาทไทย (Thai Baht)",
    nameEn: "Thai Baht",
    symbol: "฿",
    decimalPlaces: 2,
    isBaseCurrency: true,
    isActive: true,
    rates: [
      {
        id: "rate-thb-1",
        effectiveDate: "2026-01-01",
        buyRate: "1.0000",
        sellRate: "1.0000",
        accountRate: "1.0000",
      },
    ],
  },
  {
    id: "cur-usd",
    code: "USD",
    nameTh: "ดอลลาร์สหรัฐ (US Dollar)",
    nameEn: "United States Dollar",
    symbol: "$",
    decimalPlaces: 2,
    isBaseCurrency: false,
    isActive: true,
    rates: [
      {
        id: "rate-usd-1",
        effectiveDate: "2026-10-01",
        buyRate: "35.8000",
        sellRate: "36.2000",
        accountRate: "36.0000",
      },
      {
        id: "rate-usd-2",
        effectiveDate: "2026-09-01",
        buyRate: "35.5000",
        sellRate: "35.9000",
        accountRate: "35.7000",
      },
    ],
  },
  {
    id: "cur-eur",
    code: "EUR",
    nameTh: "ยูโร (Euro)",
    nameEn: "Euro",
    symbol: "€",
    decimalPlaces: 2,
    isBaseCurrency: false,
    isActive: true,
    rates: [
      {
        id: "rate-eur-1",
        effectiveDate: "2026-10-01",
        buyRate: "38.9000",
        sellRate: "39.4000",
        accountRate: "39.1500",
      },
    ],
  },
  {
    id: "cur-jpy",
    code: "JPY",
    nameTh: "เยนญี่ปุ่น (Japanese Yen 100)",
    nameEn: "Japanese Yen (per 100)",
    symbol: "¥",
    decimalPlaces: 2,
    isBaseCurrency: false,
    isActive: true,
    rates: [
      {
        id: "rate-jpy-1",
        effectiveDate: "2026-10-01",
        buyRate: "23.4000",
        sellRate: "23.8500",
        accountRate: "23.6000",
      },
    ],
  },
  {
    id: "cur-cny",
    code: "CNY",
    nameTh: "หยวนจีน (Chinese Yuan)",
    nameEn: "Chinese Yuan Renminbi",
    symbol: "¥",
    decimalPlaces: 2,
    isBaseCurrency: false,
    isActive: true,
    rates: [
      {
        id: "rate-cny-1",
        effectiveDate: "2026-10-01",
        buyRate: "4.9500",
        sellRate: "5.1000",
        accountRate: "5.0200",
      },
    ],
  },
];

export interface SystemCurrencyScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemCurrencyScreen({
  language = "th",
  holdingcode: _holdingcode = "",
}: SystemCurrencyScreenProps) {
  const isTh = language === "th";
  const [currencies, setCurrencies] = useState<CurrencyItem[]>(DEFAULT_CURRENCIES);
  const [selectedId, setSelectedId] = useState<string>(DEFAULT_CURRENCIES[0].id);
  const [notification, setNotification] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const { splitPercent, isResizing, startResize, adjustWithKeyboard } = useSplitPercent({
    containerRef,
    defaultLeft: 35,
    min: 25,
    max: 55,
    storageKey: "bcai_split_currency",
  });

  const selectedCurrency = useMemo(() => {
    return currencies.find((c) => c.id === selectedId) || currencies[0];
  }, [currencies, selectedId]);

  const handleUpdateField = <K extends keyof CurrencyItem>(
    field: K,
    val: CurrencyItem[K]
  ) => {
    setCurrencies((prev) =>
      prev.map((c) => (c.id === selectedCurrency.id ? { ...c, [field]: val } : c))
    );
  };

  const handleSetBaseCurrency = () => {
    setCurrencies((prev) =>
      prev.map((c) => ({
        ...c,
        isBaseCurrency: c.id === selectedCurrency.id,
      }))
    );
    setNotification(
      isTh
        ? `กำหนด ${selectedCurrency.code} เป็นสกุลเงินหลักของระบบเรียบร้อย`
        : `Set ${selectedCurrency.code} as base system currency.`
    );
    setTimeout(() => setNotification(null), 3000);
  };

  const handleAddRate = () => {
    const today = new Date().toISOString().split("T")[0];
    const newRate: ExchangeRateEntry = {
      id: `rate-${Date.now()}`,
      effectiveDate: today,
      buyRate: "1.0000",
      sellRate: "1.0000",
      accountRate: "1.0000",
    };
    setCurrencies((prev) =>
      prev.map((c) =>
        c.id === selectedCurrency.id
          ? { ...c, rates: [newRate, ...c.rates] }
          : c
      )
    );
  };

  const handleUpdateRate = <K extends keyof ExchangeRateEntry>(
    rateId: string,
    field: K,
    val: ExchangeRateEntry[K]
  ) => {
    setCurrencies((prev) =>
      prev.map((c) =>
        c.id === selectedCurrency.id
          ? {
              ...c,
              rates: c.rates.map((r) =>
                r.id === rateId ? { ...r, [field]: val } : r
              ),
            }
          : c
      )
    );
  };

  const handleSave = () => {
    setNotification(
      isTh
        ? `บันทึกข้อมูลสกุลเงิน (${selectedCurrency.code}) และอัตราแลกเปลี่ยนเรียบร้อยแล้ว`
        : `Currency settings for (${selectedCurrency.code}) saved successfully.`
    );
    setTimeout(() => setNotification(null), 3500);
  };

  const handleReset = () => {
    setCurrencies(DEFAULT_CURRENCIES);
    setNotification(isTh ? "คืนค่าการตั้งค่าเริ่มต้นเรียบร้อย" : "Reset currencies to defaults");
    setTimeout(() => setNotification(null), 3000);
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 dark:bg-slate-950 font-sans text-xs">
      {/* Header Bar */}
      <div className="flex-none p-3.5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-wrap items-center justify-between gap-3 shadow-xs">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
            <Coins className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-semibold text-slate-900 dark:text-slate-100">
                {isTh ? "กำหนดสกุลเงิน" : "Currencies Configuration"}
              </h1>
              <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {isTh
                ? "จัดการระบบ › กำหนดสกุลเงิน สกุลเงินหลักของระบบ และประวัติอัตราแลกเปลี่ยนตามวันที่มีผล"
                : "System Settings › Setup foreign currencies, system base currency and historical exchange rates"}
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
        {/* Left Column: Currencies List */}
        <div
          style={{ width: `${splitPercent}%` }}
          className="h-full flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900"
        >
          <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
            <span className="font-semibold text-slate-700 dark:text-slate-200">
              {isTh ? "สกุลเงินทั้งหมด" : "Currencies"} ({currencies.length})
            </span>
            <span className="text-[11px] text-slate-500 font-mono">Select Currency</span>
          </div>

          <div className="flex-1 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
            {currencies.map((cur) => {
              const isSelected = cur.id === selectedCurrency.id;
              const latestRate = cur.rates[0];
              return (
                <div
                  key={cur.id}
                  onClick={() => setSelectedId(cur.id)}
                  className={`p-3.5 cursor-pointer transition-colors ${
                    isSelected
                      ? "bg-indigo-50/80 dark:bg-indigo-950/40 border-l-4 border-indigo-600"
                      : "hover:bg-slate-50 dark:hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-sm">
                        {cur.code}
                      </span>
                      <span className="text-slate-400 font-mono">({cur.symbol})</span>
                    </div>
                    {cur.isBaseCurrency ? (
                      <span className="px-2 py-0.5 rounded text-[10px] font-medium bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400">
                        {isTh ? "สกุลเงินหลัก" : "Base"}
                      </span>
                    ) : (
                      <span className="text-[11px] font-mono font-bold text-slate-600 dark:text-slate-300">
                        {latestRate ? latestRate.accountRate : "1.0000"}
                      </span>
                    )}
                  </div>
                  <p className="text-slate-800 dark:text-slate-200 font-medium">
                    {isTh ? cur.nameTh : cur.nameEn}
                  </p>
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

        {/* Right Column: Currency Details & Exchange Rates */}
        <div
          style={{ width: `${100 - splitPercent}%` }}
          className="h-full flex flex-col bg-slate-50/50 dark:bg-slate-950 overflow-y-auto p-4 gap-4"
        >
          {/* Currency Profile Card */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-4 shadow-xs">
            <div className="flex items-center justify-between pb-3 mb-4 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <Sliders className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
                <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
                  {isTh ? "รายละเอียดสกุลเงิน" : "Currency Specifications"}
                </h2>
              </div>
              {!selectedCurrency.isBaseCurrency && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleSetBaseCurrency}
                  className="h-7 text-[11px] text-emerald-600 border-emerald-300 hover:bg-emerald-50"
                >
                  {isTh ? "ตั้งเป็นสกุลเงินหลักของระบบ" : "Set as Base Currency"}
                </Button>
              )}
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "รหัสสกุลเงิน (Currency Code)" : "Currency Code"}
                </label>
                <Input
                  value={selectedCurrency.code}
                  onChange={(e) => handleUpdateField("code", e.target.value.toUpperCase())}
                  className="font-mono uppercase font-bold text-indigo-600 dark:text-indigo-400"
                  maxLength={5}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "สัญลักษณ์ (Symbol)" : "Symbol"}
                </label>
                <Input
                  value={selectedCurrency.symbol}
                  onChange={(e) => handleUpdateField("symbol", e.target.value)}
                  className="font-mono text-center font-bold w-24"
                  maxLength={4}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อสกุลเงินภาษาไทย" : "Thai Name"}
                </label>
                <Input
                  value={selectedCurrency.nameTh}
                  onChange={(e) => handleUpdateField("nameTh", e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-600 dark:text-slate-300 font-medium mb-1">
                  {isTh ? "ชื่อสกุลเงินภาษาอังกฤษ" : "English Name"}
                </label>
                <Input
                  value={selectedCurrency.nameEn}
                  onChange={(e) => handleUpdateField("nameEn", e.target.value)}
                />
              </div>
            </div>
          </div>

          {/* Exchange Rates History Card */}
          <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-xs flex-1 flex flex-col min-h-[300px]">
            <div className="p-3.5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/50 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
                <span className="font-semibold text-slate-800 dark:text-slate-100">
                  {isTh ? "ตารางอัตราแลกเปลี่ยนตามวันที่มีผล" : "Exchange Rates Table"}
                </span>
                <span className="text-[11px] text-slate-500 font-mono">
                  (เทียบ 1 {selectedCurrency.code} = THB)
                </span>
              </div>
              {!selectedCurrency.isBaseCurrency && (
                <Button
                  size="sm"
                  onClick={handleAddRate}
                  className="h-7 text-xs gap-1 bg-indigo-600 hover:bg-indigo-700 text-white"
                >
                  <Plus className="w-3.5 h-3.5" />
                  {isTh ? "เพิ่มอัตราแลกเปลี่ยน" : "Add Rate"}
                </Button>
              )}
            </div>

            <div className="flex-1 overflow-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead className="bg-slate-100/80 dark:bg-slate-800/80 text-slate-600 dark:text-slate-300 font-medium border-b border-slate-200 dark:border-slate-700">
                  <tr>
                    <th className="py-2.5 px-3 w-36">{isTh ? "วันที่มีผลบังคับใช้" : "Effective Date"}</th>
                    <th className="py-2.5 px-3 text-right">{isTh ? "อัตราซื้อ (Buy)" : "Buy Rate"}</th>
                    <th className="py-2.5 px-3 text-right">{isTh ? "อัตราขาย (Sell)" : "Sell Rate"}</th>
                    <th className="py-2.5 px-3 text-right font-semibold text-indigo-600 dark:text-indigo-400">
                      {isTh ? "อัตราทางบัญชี (Accounting)" : "Accounting Rate"}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {selectedCurrency.rates.map((rate) => (
                    <tr key={rate.id} className="hover:bg-slate-50 dark:hover:bg-slate-800/40">
                      <td className="py-2 px-3">
                        <Input
                          type="date"
                          value={rate.effectiveDate}
                          disabled={selectedCurrency.isBaseCurrency}
                          onChange={(e) => handleUpdateRate(rate.id, "effectiveDate", e.target.value)}
                          className="h-7 text-xs font-mono w-32"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          step="0.0001"
                          value={rate.buyRate}
                          disabled={selectedCurrency.isBaseCurrency}
                          onChange={(e) => handleUpdateRate(rate.id, "buyRate", e.target.value)}
                          className="h-7 text-xs font-mono text-right w-28 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          step="0.0001"
                          value={rate.sellRate}
                          disabled={selectedCurrency.isBaseCurrency}
                          onChange={(e) => handleUpdateRate(rate.id, "sellRate", e.target.value)}
                          className="h-7 text-xs font-mono text-right w-28 ml-auto"
                        />
                      </td>
                      <td className="py-2 px-3 text-right">
                        <Input
                          type="number"
                          step="0.0001"
                          value={rate.accountRate}
                          disabled={selectedCurrency.isBaseCurrency}
                          onChange={(e) => handleUpdateRate(rate.id, "accountRate", e.target.value)}
                          className="h-7 text-xs font-mono text-right font-bold text-indigo-600 dark:text-indigo-400 w-28 ml-auto"
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

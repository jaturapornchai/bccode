"use client";

import {
  AlertCircle,
  BadgeCheck,
  CircleDollarSign,
  Edit3,
  Loader2,
  Plus,
  RefreshCcw,
  Save,
  Search,
  Trash2,
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { applyCurrencySymbolPreset, currencyPresetSource, filterCurrencySymbolPresets, findCurrencySymbolPreset } from "@/lib/currency-presets";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import {
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { LanguageDialog } from "../language-dialog";
import { ManualLink } from "../manual-link";
import { ThemeToggle } from "../theme-toggle";

type CurrencyScreenProps = {
  embedded?: boolean;
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage?: LanguageCode;
  language?: LanguageCode;
};

type CurrencyRecord = {
  guidfixed: string;
  duplicate_guidfixeds?: string[];
  code: string;
  name: string;
  symbol: string;
  isdisabled: boolean;
  exchangerates?: ExchangeRateEntry[];
};

type ExchangeRateEntry = {
  guidfixed?: string;
  date?: string;
  rate?: number;
};

type CurrencyForm = {
  code: string;
  name: string;
  symbol: string;
  isdisabled: boolean;
};

type ApiResponse<T> = {
  success?: boolean;
  message?: string;
  data?: T;
  id?: string;
  total?: number;
  pagination?: {
    total?: number;
    totalitem?: number;
    totalItem?: number;
  };
};

type Notice = {
  type: "success" | "error" | "info";
  text: string;
} | null;

const currencyTextEn = {
  title: "Currency",
  subtitle: "Manage document currencies, symbols, and active status for the selected company.",
  refresh: "Refresh",
  total: "Total",
  active: "Active",
  disabled: "Disabled",
  baseCurrency: "Base currency",
  search: "Search currency",
  add: "Add currency",
  edit: "Edit currency",
  delete: "Delete",
  save: "Save",
  cancel: "Cancel",
  saving: "Saving",
  loading: "Loading currency data",
  noData: "No currency data",
  emptyHint: "Add the first currency or change the search term.",
  code: "Currency code",
  name: "Currency name",
  symbol: "Symbol",
  status: "Status",
  enabled: "Enabled",
  baseBadge: "Base",
  formNew: "New currency",
  formEdit: "Edit currency",
  editLocked: "The code is locked in edit mode to match the legacy workflow.",
  requiredCode: "Please enter currency code.",
  requiredName: "Please enter currency name.",
  codeLength: "Currency code must be 2-5 characters.",
  selectSymbol: "Please select currency symbol.",
  symbolSearch: "Search code, name, or symbol",
  currencyVerified: "Verified",
  currencyNotVerified: "Not verified",
  currencyVerifySource: "Code/name: ISO 4217 / SIX. Symbol/abbreviation: Wikipedia.",
  currencyVerifyRequired: "Select a currency from the verified list.",
  duplicateCode: "This currency code already exists.",
  noSymbolResult: "No matching currency",
  loadFailed: "Could not load currencies.",
  requestFailed: "Request failed.",
  deleteConfirm: "Confirm delete this currency?",
  addSuccess: "Currency added.",
  editSuccess: "Currency updated.",
  deleteSuccess: "Currency deleted.",
  apiRequired: "Please login and select a company before opening this screen.",
  tenant: "Company",
  branch: "Branch",
  exchangeRates: "Exchange rates",
  none: "None",
} as const;

type CurrencyTextKey = keyof typeof currencyTextEn;

const currencyText: Partial<Record<LanguageCode, Partial<Record<CurrencyTextKey, string>>>> = {
  th: {
    title: "สกุลเงิน",
    subtitle: "จัดการสกุลเงิน สัญลักษณ์ และสถานะใช้งานของบริษัทที่เลือก",
    refresh: "โหลดใหม่",
    total: "ทั้งหมด",
    active: "ใช้งาน",
    disabled: "ปิดใช้งาน",
    baseCurrency: "สกุลเงินหลัก",
    search: "ค้นหาสกุลเงิน",
    add: "เพิ่มสกุลเงิน",
    edit: "แก้ไขสกุลเงิน",
    delete: "ลบ",
    save: "บันทึก",
    cancel: "ยกเลิก",
    saving: "กำลังบันทึก",
    loading: "กำลังโหลดสกุลเงิน",
    noData: "ไม่มีข้อมูลสกุลเงิน",
    emptyHint: "เพิ่มสกุลเงินแรก หรือเปลี่ยนคำค้นหา",
    code: "รหัสสกุลเงิน",
    name: "ชื่อสกุลเงิน",
    symbol: "สัญลักษณ์",
    status: "สถานะ",
    enabled: "เปิดใช้งาน",
    baseBadge: "หลัก",
    formNew: "สกุลเงินใหม่",
    formEdit: "แก้ไขสกุลเงิน",
    editLocked: "โหมดแก้ไขล็อกรหัสตาม workflow เดิม",
    requiredCode: "กรุณากรอกรหัสสกุลเงิน",
    requiredName: "กรุณากรอกชื่อสกุลเงิน",
    codeLength: "รหัสสกุลเงินต้องมีความยาว 2-5 ตัวอักษร",
    selectSymbol: "กรุณาเลือกสัญลักษณ์สกุลเงิน",
    symbolSearch: "ค้นหารหัส ชื่อ หรือสัญลักษณ์",
    currencyVerified: "ยืนยันแล้ว",
    currencyNotVerified: "ยังไม่ยืนยัน",
    currencyVerifySource: "รหัส/ชื่อ: ISO 4217 / SIX; สัญลักษณ์/ตัวย่อ: Wikipedia",
    currencyVerifyRequired: "กรุณาเลือกสกุลเงินจากรายการที่ค้นพบเพื่อยืนยันว่ามีอยู่จริง",
    duplicateCode: "รหัสสกุลเงินนี้มีอยู่แล้ว",
    noSymbolResult: "ไม่พบสกุลเงินที่ตรงกัน",
    loadFailed: "โหลดสกุลเงินไม่สำเร็จ",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    deleteConfirm: "ยืนยันลบสกุลเงินนี้?",
    addSuccess: "เพิ่มสกุลเงินสำเร็จ",
    editSuccess: "แก้ไขสกุลเงินสำเร็จ",
    deleteSuccess: "ลบสกุลเงินสำเร็จ",
    apiRequired: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้",
    tenant: "บริษัท",
    branch: "สาขา",
    exchangeRates: "อัตราแลกเปลี่ยน",
    none: "ไม่มี",
  },
  en: currencyTextEn,
  cn: {
    subtitle: "管理所选公司的单据货币、符号和启用状态。",
    refresh: "刷新",
    total: "全部",
    active: "启用",
    disabled: "停用",
    status: "状态",
    enabled: "启用",
    emptyHint: "新增第一个货币或更改搜索条件。",
    editLocked: "编辑时锁定代码以符合旧流程。",
    apiRequired: "请先登录并选择公司。",
    tenant: "公司",
    branch: "分支",
    exchangeRates: "汇率",
    none: "无",
  },
  ja: {
    subtitle: "選択した会社の通貨、記号、有効状態を管理します。",
    refresh: "再読み込み",
    total: "合計",
    active: "有効",
    disabled: "無効",
    status: "状態",
    enabled: "有効",
    emptyHint: "最初の通貨を追加するか検索条件を変更してください。",
    editLocked: "編集時は旧ワークフローに合わせてコードを固定します。",
    apiRequired: "ログインして会社を選択してください。",
    tenant: "会社",
    branch: "支店",
    exchangeRates: "為替レート",
    none: "なし",
  },
  ko: {
    subtitle: "선택한 회사의 문서 통화, 기호, 사용 상태를 관리합니다.",
    refresh: "새로고침",
    total: "전체",
    active: "사용",
    disabled: "중지",
    status: "상태",
    enabled: "사용",
    emptyHint: "첫 통화를 추가하거나 검색어를 변경하세요.",
    editLocked: "수정 모드에서는 기존 흐름에 맞춰 코드가 잠깁니다.",
    apiRequired: "먼저 로그인하고 회사를 선택하세요.",
    tenant: "회사",
    branch: "지점",
    exchangeRates: "환율",
    none: "없음",
  },
  lo: {
    subtitle: "ຈັດການສະກຸນເງິນ ສັນຍາລັກ ແລະສະຖານະໃຊ້ງານຂອງກິດຈະການທີ່ເລືອກ.",
    refresh: "ໂຫຼດໃໝ່",
    total: "ທັງໝົດ",
    active: "ໃຊ້ງານ",
    disabled: "ປິດ",
    status: "ສະຖານະ",
    enabled: "ເປີດໃຊ້",
    emptyHint: "ເພີ່ມສະກຸນເງິນທໍາອິດ ຫຼືປ່ຽນຄໍາຄົ້ນຫາ.",
    editLocked: "ໂໝດແກ້ໄຂຈະລັອກລະຫັດຕາມ workflow ເກົ່າ.",
    apiRequired: "ກະລຸນາເຂົ້າລະບົບ ແລະເລືອກກິດຈະການກ່ອນ.",
    tenant: "ກິດຈະການ",
    branch: "ສາຂາ",
    exchangeRates: "ອັດຕາແລກປ່ຽນ",
    none: "ບໍ່ມີ",
  },
  my: {
    subtitle: "ရွေးထားသောကုမ္ပဏီအတွက် ငွေကြေး၊ သင်္ကေတနှင့် အသုံးပြုမှုအခြေအနေကို စီမံပါ။",
    refresh: "ပြန်ဖွင့်",
    total: "စုစုပေါင်း",
    active: "အသုံးပြု",
    disabled: "ပိတ်ထား",
    status: "အခြေအနေ",
    enabled: "ဖွင့်ထား",
    emptyHint: "ပထမဆုံးငွေကြေးကိုထည့်ပါ သို့မဟုတ် ရှာဖွေချက်ပြောင်းပါ။",
    editLocked: "ပြင်ဆင်ချိန်တွင် ကုဒ်ကို workflow ဟောင်းအတိုင်း ပိတ်ထားသည်။",
    apiRequired: "အရင် login ဝင်ပြီး ကုမ္ပဏီရွေးပါ။",
    tenant: "ကုမ္ပဏီ",
    branch: "ဌာနခွဲ",
    exchangeRates: "လဲလှယ်နှုန်း",
    none: "မရှိပါ",
  },
  km: {
    subtitle: "គ្រប់គ្រងរូបិយប័ណ្ណ សញ្ញា និងស្ថានភាពប្រើប្រាស់របស់ក្រុមហ៊ុនដែលបានជ្រើស។",
    refresh: "ផ្ទុកឡើងវិញ",
    total: "សរុប",
    active: "កំពុងប្រើ",
    disabled: "បានបិទ",
    status: "ស្ថានភាព",
    enabled: "បើកប្រើ",
    emptyHint: "បន្ថែមរូបិយប័ណ្ណដំបូង ឬប្តូរពាក្យស្វែងរក។",
    editLocked: "ពេលកែប្រែ កូដត្រូវបានចាក់សោតាម workflow ចាស់។",
    apiRequired: "សូមចូលប្រើ និងជ្រើសក្រុមហ៊ុនជាមុន។",
    tenant: "ក្រុមហ៊ុន",
    branch: "សាខា",
    exchangeRates: "អត្រាប្តូរប្រាក់",
    none: "គ្មាន",
  },
  vi: {
    subtitle: "Quản lý tiền tệ, ký hiệu và trạng thái sử dụng của công ty đã chọn.",
    refresh: "Tải lại",
    total: "Tất cả",
    active: "Đang dùng",
    disabled: "Đã tắt",
    status: "Trạng thái",
    enabled: "Bật",
    emptyHint: "Thêm tiền tệ đầu tiên hoặc đổi từ khóa tìm kiếm.",
    editLocked: "Khi chỉnh sửa, mã bị khóa theo quy trình cũ.",
    apiRequired: "Vui lòng đăng nhập và chọn công ty trước.",
    tenant: "Công ty",
    branch: "Chi nhánh",
    exchangeRates: "Tỷ giá",
    none: "Không có",
  },
  ms: {
    subtitle: "Urus mata wang, simbol dan status aktif syarikat yang dipilih.",
    refresh: "Muat semula",
    total: "Jumlah",
    active: "Aktif",
    disabled: "Dilumpuhkan",
    status: "Status",
    enabled: "Aktif",
    emptyHint: "Tambah mata wang pertama atau tukar carian.",
    editLocked: "Kod dikunci semasa edit mengikut aliran lama.",
    apiRequired: "Sila log masuk dan pilih syarikat dahulu.",
    tenant: "Syarikat",
    branch: "Cawangan",
    exchangeRates: "Kadar tukaran",
    none: "Tiada",
  },
  id: {
    subtitle: "Kelola mata uang, simbol, dan status aktif perusahaan yang dipilih.",
    refresh: "Muat ulang",
    total: "Total",
    active: "Aktif",
    disabled: "Dinonaktifkan",
    status: "Status",
    enabled: "Aktif",
    emptyHint: "Tambah mata uang pertama atau ubah pencarian.",
    editLocked: "Kode dikunci saat edit sesuai alur lama.",
    apiRequired: "Login dan pilih perusahaan terlebih dahulu.",
    tenant: "Perusahaan",
    branch: "Cabang",
    exchangeRates: "Kurs",
    none: "Tidak ada",
  },
  fil: {
    subtitle: "Pamahalaan ang currency, symbol, at active status ng napiling company.",
    refresh: "I-refresh",
    total: "Kabuuan",
    active: "Aktibo",
    disabled: "Naka-disable",
    status: "Status",
    enabled: "Aktibo",
    emptyHint: "Magdagdag ng unang currency o baguhin ang search.",
    editLocked: "Naka-lock ang code sa edit mode gaya ng lumang workflow.",
    apiRequired: "Mag-login at pumili muna ng company.",
    tenant: "Company",
    branch: "Branch",
    exchangeRates: "Exchange rates",
    none: "Wala",
  },
};

const backendKeys: Partial<Record<CurrencyTextKey, string>> = {
  add: "add_currency",
  addSuccess: "add_currency_success",
  baseCurrency: "basecurrency",
  code: "currency_code",
  codeLength: "currency_code_length_error",
  deleteConfirm: "confirm_delete_currency",
  deleteSuccess: "delete_currency_success",
  disabled: "disabled",
  edit: "edit_currency",
  editSuccess: "edit_currency_success",
  name: "currency_name",
  noData: "no_currency_data",
  requiredCode: "please_enter_currency_code",
  requiredName: "please_enter_currency_name",
  save: "save",
  saving: "saving",
  search: "search_currency",
  selectSymbol: "please_select_currency_symbol",
  symbolSearch: "search_currency_symbol",
  currencyVerifyRequired: "currency_verify_required",
  currencyVerifySource: "currency_verify_source",
  duplicateCode: "duplicate_currency_code",
  symbol: "symbol",
  title: "currency",
};

const emptyForm: CurrencyForm = {
  code: "",
  name: "",
  symbol: "",
  isdisabled: false,
};

export function CurrencyScreen({ embedded = false, initialBackendLanguage, initialBackendUrl, initialLanguage = "th", language: externalLanguage }: CurrencyScreenProps) {
  const router = useRouter();
  const [language, setLanguage] = useState<LanguageCode>(externalLanguage ?? initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [currencies, setCurrencies] = useState<CurrencyRecord[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);
  const [editing, setEditing] = useState<CurrencyRecord | null>(null);
  const [form, setForm] = useState<CurrencyForm>(emptyForm);
  const [formOpen, setFormOpen] = useState(false);
  const { confirm, confirmationDialog } = useConfirmDialog();
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);

  const text = useCallback(
    (key: CurrencyTextKey) => {
      const fallback = currencyText[language]?.[key] ?? currencyTextEn[key] ?? key;
      return backendText(backendLanguage, backendKeys[key] ?? key, fallback);
    },
    [backendLanguage, language],
  );

  const loadCurrencies = useCallback(async (currentAuth: AuthSession | null) => {
    if (!currentAuth) return;
    setLoading(true);
    setNotice(null);
    try {
      const response = await fetch("/api/currency?page=1&limit=1000", {
        headers: {
          "x-bc-backend-url": currentAuth.backendUrl,
          Authorization: `Bearer ${currentAuth.token}`,
        },
        cache: "no-store",
      });
      const payload = await response.json() as ApiResponse<unknown>;
      if (!response.ok || payload.success === false) {
        throw new Error(payload.message || currencyTextEn.loadFailed);
      }
      setCurrencies(normalizeCurrencies(payload.data));
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : currencyTextEn.loadFailed });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (externalLanguage) setLanguage(externalLanguage);
  }, [externalLanguage]);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? externalLanguage ?? initialLanguage);
    if (!externalLanguage) setLanguage(savedLanguage);

    const nextAuth = readAuth();
    const nextWorkspace = readWorkspace();
    if (!nextAuth || !nextWorkspace) {
      setNotice({ type: "info", text: getLocalCurrencyText(savedLanguage, "apiRequired") });
      if (!embedded) router.replace("/");
      return;
    }

    setAuth(nextAuth);
    setWorkspace(nextWorkspace);
    void loadCurrencies(nextAuth);
  }, [embedded, externalLanguage, initialLanguage, loadCurrencies, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    if (!externalLanguage) localStorage.setItem("user_language", language);
  }, [externalLanguage, language]);

  const baseCurrency = (workspace?.branch?.basecurrency || "").trim().toUpperCase();
  const activeCount = currencies.filter((item) => !item.isdisabled).length;
  const disabledCount = currencies.length - activeCount;
  const visibleCurrencies = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return currencies;
    return currencies.filter((currency) =>
      `${currency.code} ${currency.name} ${currency.symbol}`.toLowerCase().includes(needle),
    );
  }, [currencies, query]);

  function openCreate() {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
    setNotice(null);
  }

  function openEdit(currency: CurrencyRecord) {
    setEditing(currency);
    setForm({
      code: currency.code,
      name: currency.name,
      symbol: currency.symbol || "฿",
      isdisabled: currency.isdisabled,
    });
    setFormOpen(true);
    setNotice(null);
  }

  async function saveCurrency(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!auth) return;

    const payload = normalizeForm(form);
    const validation = validateForm(payload, text);
    if (validation) {
      setNotice({ type: "error", text: validation });
      return;
    }
    const duplicate = currencies.some((currency) =>
      currency.code === payload.code &&
      (!editing || currency.guidfixed !== editing.guidfixed),
    );
    if (duplicate) {
      setNotice({ type: "error", text: text("duplicateCode") });
      return;
    }

    setSaving(true);
    setNotice(null);
    try {
      const path = editing ? `/api/currency/${encodeURIComponent(editing.guidfixed)}` : "/api/currency";
      const response = await fetch(path, {
        method: editing ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify({
          backendUrl: auth.backendUrl,
          guidfixed: editing?.guidfixed ?? "",
          ...payload,
        }),
      });
      const data = await response.json() as ApiResponse<unknown>;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text("requestFailed"));
      }

      setFormOpen(false);
      setEditing(null);
      setNotice({ type: "success", text: editing ? text("editSuccess") : text("addSuccess") });
      await loadCurrencies(auth);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("requestFailed") });
    } finally {
      setSaving(false);
    }
  }

  async function deleteCurrency(currency: CurrencyRecord) {
    if (!auth) return;
    const deleteIds = uniqueStrings([currency.guidfixed, ...(currency.duplicate_guidfixeds ?? [])]);
    if (!deleteIds.length) {
      setNotice({ type: "error", text: text("requestFailed") });
      return;
    }
    const confirmed = await confirm({
      title: text("deleteConfirm"),
      description: `${currency.code} ${currency.name}`.trim(),
      details: currency.duplicate_guidfixeds && currency.duplicate_guidfixeds.length > 1
        ? `${language === "th" ? "จะลบรายการซ้ำทั้งหมด" : "All duplicate entries will be deleted"}: ${currency.duplicate_guidfixeds.length}`
        : undefined,
      confirmLabel: text("delete"),
      cancelLabel: text("cancel"),
      tone: "danger",
    });
    if (!confirmed) return;

    setLoading(true);
    setNotice(null);
    try {
      const response = await fetch(deleteIds.length === 1 ? `/api/currency/${encodeURIComponent(deleteIds[0])}` : "/api/currency", {
        method: "DELETE",
        headers: {
          ...(deleteIds.length > 1 ? { "Content-Type": "application/json" } : {}),
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: deleteIds.length > 1 ? JSON.stringify(deleteIds) : undefined,
      });
      const data = await response.json() as ApiResponse<unknown>;
      if (!response.ok || data.success === false) {
        throw new Error(data.message || text("requestFailed"));
      }
      setNotice({ type: "success", text: text("deleteSuccess") });
      await loadCurrencies(auth);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("requestFailed") });
    } finally {
      setLoading(false);
    }
  }

  const content = (
    <div className="grid w-full min-w-0 gap-3">
      <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
        <div className="flex w-full flex-wrap items-start justify-between gap-2">
          <div className="flex min-w-0 items-start gap-2">
            <span className="grid size-10 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              <CircleDollarSign size={22} />
            </span>
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase text-muted-foreground">BC Ai Account</p>
              <h1 className="truncate text-xl font-semibold sm:text-2xl">{text("title")}</h1>
              <p className="max-w-[92ch] text-sm leading-6 text-muted-foreground">{text("subtitle")}</p>
            </div>
          </div>
          <div className="flex min-w-0 flex-wrap justify-end gap-2">
            {embedded ? null : <LanguageDialog language={language} onLanguageChange={setLanguage} />}
            <ManualLink compact language={language} screen="currency" />
            {embedded ? null : <ThemeToggle language={language} />}
          </div>
        </div>
        {baseCurrency ? (
          <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
            <Badge variant="success">{text("baseCurrency")}: {baseCurrency}</Badge>
          </div>
        ) : null}
      </header>

      {notice ? (
        <div className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"}`}>
          {notice.type === "success" ? <BadgeCheck size={18} /> : <AlertCircle size={18} />}
          <span>{notice.text}</span>
        </div>
      ) : null}

      <section className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label={text("total")} value={formatCount(currencies.length, language)} />
        <StatCard label={text("active")} value={formatCount(activeCount, language)} tone="success" />
        <StatCard label={text("disabled")} value={formatCount(disabledCount, language)} tone="warning" />
        <StatCard label={text("baseCurrency")} value={baseCurrency || "-"} />
      </section>

      <Card>
        <CardContent className="grid gap-2 p-3">
          <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto]">
            <label className="relative block min-w-0">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input className="!pl-10" value={query} onChange={(event) => setQuery(event.target.value)} placeholder={text("search")} />
            </label>
            <Button type="button" variant="outline" onClick={() => void loadCurrencies(auth)} disabled={loading || !auth}>
              {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {text("refresh")}
            </Button>
            <Button type="button" onClick={openCreate} disabled={!auth}>
              <Plus />
              {text("add")}
            </Button>
          </div>
        </CardContent>
      </Card>

      {formOpen ? (
        <CurrencyFormPanel
          editing={editing}
          form={form}
          onClose={() => {
            if (!saving) setFormOpen(false);
          }}
          onFormChange={setForm}
          onSubmit={saveCurrency}
          saving={saving}
          text={text}
        />
      ) : null}

      {loading && currencies.length === 0 ? (
        <Card>
          <CardContent className="flex min-h-40 items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
            <Loader2 className="animate-spin" />
            {text("loading")}
          </CardContent>
        </Card>
      ) : visibleCurrencies.length ? (
        <section className="grid min-w-0 gap-2 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" aria-label={text("title")}>
          {visibleCurrencies.map((currency) => (
            <CurrencyCard
              baseCurrency={baseCurrency}
              currency={currency}
              key={currency.guidfixed || currency.code}
              onDelete={deleteCurrency}
              onEdit={openEdit}
              text={text}
            />
          ))}
        </section>
      ) : (
        <Card>
          <CardContent className="grid min-h-44 place-items-center p-4 text-center">
            <div className="grid gap-2">
              <CircleDollarSign className="mx-auto size-10 text-muted-foreground" />
              <h2 className="text-base font-semibold">{text("noData")}</h2>
              <p className="text-sm text-muted-foreground">{text("emptyHint")}</p>
              <Button type="button" onClick={openCreate} disabled={!auth}>
                <Plus />
                {text("add")}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {confirmationDialog}
    </div>
  );

  if (embedded) return <section className="grid w-full min-w-0 gap-3">{content}</section>;
  return <main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">{content}</main>;
}

function StatCard({ label, tone, value }: { label: string; tone?: "success" | "warning"; value: string }) {
  return (
    <Card className="shadow-sm">
      <CardHeader className="p-3 pb-1">
        <CardDescription>{label}</CardDescription>
        <CardTitle className={tone === "success" ? "text-emerald-600 dark:text-emerald-300" : tone === "warning" ? "text-amber-600 dark:text-amber-300" : ""}>
          {value}
        </CardTitle>
      </CardHeader>
    </Card>
  );
}

function CurrencyCard({
  baseCurrency,
  currency,
  onDelete,
  onEdit,
  text,
}: {
  baseCurrency: string;
  currency: CurrencyRecord;
  onDelete: (currency: CurrencyRecord) => void;
  onEdit: (currency: CurrencyRecord) => void;
  text: (key: CurrencyTextKey) => string;
}) {
  const isBase = baseCurrency && currency.code.toUpperCase() === baseCurrency;
  const verifiedPreset = findCurrencySymbolPreset(currency);

  return (
    <Card className="min-w-0 shadow-sm">
      <CardContent className="grid gap-2 p-3">
        <div className="flex min-w-0 items-start justify-between gap-2">
          <div className="flex min-w-0 items-center gap-2">
            <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-primary/10 text-xl font-semibold text-primary">
              {currency.symbol || "?"}
            </span>
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-1.5">
                <h2 className="truncate text-base font-semibold">{currency.code}</h2>
                <Badge variant={verifiedPreset ? "success" : "warning"}>
                  {verifiedPreset ? text("currencyVerified") : text("currencyNotVerified")}
                </Badge>
                {isBase ? <Badge variant="success">{text("baseBadge")}</Badge> : null}
                {currency.isdisabled ? <Badge variant="warning">{text("disabled")}</Badge> : null}
              </div>
              <p className="truncate text-sm text-muted-foreground">{currency.name}</p>
              {verifiedPreset ? (
                <p className="truncate text-xs text-muted-foreground">
                  {verifiedPreset.isoName} · {currencyPresetSource.standard} · {currencyPresetSource.symbolSource}
                </p>
              ) : null}
            </div>
          </div>
          <div className="flex shrink-0 gap-1">
            <Button type="button" variant="outline" size="icon" onClick={() => onEdit(currency)} aria-label={text("edit")} title={text("edit")}>
              <Edit3 />
            </Button>
            <Button type="button" variant="outline" size="icon" onClick={() => onDelete(currency)} aria-label={text("delete")} title={text("delete")}>
              <Trash2 />
            </Button>
          </div>
        </div>
        <div className="grid gap-1 text-xs text-muted-foreground">
          <span className="flex items-center justify-between gap-2 rounded-xl border border-border bg-background px-2 py-1.5">
            <span>{text("status")}</span>
            <b className="text-foreground">{currency.isdisabled ? text("disabled") : text("enabled")}</b>
          </span>
          <span className="flex items-center justify-between gap-2 rounded-xl border border-border bg-background px-2 py-1.5">
            <span>{text("exchangeRates")}</span>
            <b className="text-foreground">{(currency.exchangerates?.length ?? 0).toLocaleString()}</b>
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

function CurrencyFormPanel({
  editing,
  form,
  onClose,
  onFormChange,
  onSubmit,
  saving,
  text,
}: {
  editing: CurrencyRecord | null;
  form: CurrencyForm;
  onClose: () => void;
  onFormChange: (form: CurrencyForm) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  saving: boolean;
  text: (key: CurrencyTextKey) => string;
}) {
  const [symbolQuery, setSymbolQuery] = useState("");
  const filteredSymbols = useMemo(() => filterCurrencySymbolPresets(symbolQuery), [symbolQuery]);
  const verifiedPreset = useMemo(() => findCurrencySymbolPreset(form), [form]);

  return (
    <Card className="shadow-sm">
      <form className="grid gap-3 p-3" onSubmit={onSubmit} aria-label={editing ? text("formEdit") : text("formNew")}>
        <header className="flex min-w-0 items-center justify-between gap-2">
          <div className="min-w-0">
            <h2 className="truncate text-lg font-semibold">{editing ? text("formEdit") : text("formNew")}</h2>
            {editing ? <p className="text-xs text-muted-foreground">{text("editLocked")}</p> : null}
          </div>
          <Button type="button" variant="outline" size="icon" onClick={onClose} disabled={saving} aria-label={text("cancel")}>
            <X />
          </Button>
        </header>

        <div className="grid gap-2">
          <div className="grid gap-2 sm:grid-cols-2">
            <label className="grid gap-1 text-sm font-semibold">
              <span>{text("code")} *</span>
              <Input
                autoFocus
                disabled={Boolean(editing)}
                maxLength={5}
                value={form.code}
                onChange={(event) => onFormChange({ ...form, code: event.target.value.toUpperCase() })}
                placeholder="USD"
              />
            </label>
            <label className="grid gap-1 text-sm font-semibold">
              <span>{text("name")} *</span>
              <Input
                value={form.name}
                onChange={(event) => onFormChange({ ...form, name: event.target.value })}
                placeholder="US Dollar"
              />
            </label>
          </div>

          <section className="grid gap-2 rounded-2xl border border-border bg-background p-2">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="min-w-0 flex-1 text-sm font-semibold">{text("symbol")} *</h3>
              <Badge variant={verifiedPreset ? "success" : "warning"}>
                {verifiedPreset ? text("currencyVerified") : text("currencyNotVerified")}
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground">
              {text("currencyVerifySource")} {currencyPresetSource.list} ({currencyPresetSource.published})
            </p>
            <label className="relative block min-w-0">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="!pl-10"
                value={symbolQuery}
                onChange={(event) => setSymbolQuery(event.target.value)}
                placeholder={text("symbolSearch")}
              />
            </label>
            <div className="grid gap-1.5 sm:grid-cols-2 md:grid-cols-3 min-[900px]:grid-cols-4">
              {filteredSymbols.map((item) => {
                const selected = form.symbol === item.symbol && form.code.toUpperCase() === item.code;
                return (
                  <button
                    className={`grid min-h-11 grid-cols-[32px_minmax(0,1fr)] items-center gap-2 rounded-xl border px-2 text-left text-sm transition ${selected ? "border-primary bg-primary/10 text-primary" : "border-border bg-card text-foreground hover:bg-accent"}`}
                    key={item.code}
                    onClick={() => onFormChange(applyCurrencySymbolPreset(form, item, Boolean(editing)))}
                    type="button"
                  >
                    <span className="text-center text-xl font-semibold">{item.symbol}</span>
                    <span className="min-w-0 truncate text-xs text-muted-foreground">{item.name} ({item.code})</span>
                  </button>
                );
              })}
              {filteredSymbols.length ? null : (
                <div className="rounded-xl border border-dashed border-border bg-card px-3 py-4 text-center text-sm font-semibold text-muted-foreground sm:col-span-2 md:col-span-3 min-[900px]:col-span-4">
                  {text("noSymbolResult")}
                </div>
              )}
            </div>
          </section>
        </div>

        <footer className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={saving}>
            {text("cancel")}
          </Button>
          <Button type="submit" disabled={saving}>
            {saving ? <Loader2 className="animate-spin" /> : <Save />}
            {saving ? text("saving") : text("save")}
          </Button>
        </footer>
      </form>
    </Card>
  );
}

function readAuth(): AuthSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.auth);
    if (!raw) return null;
    const auth = JSON.parse(raw) as AuthSession;
    return auth.token && auth.backendUrl ? auth : null;
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

function normalizeCurrencies(value: unknown): CurrencyRecord[] {
  if (!Array.isArray(value)) return [];
  const currencies = value.map(normalizeCurrency).filter((currency): currency is CurrencyRecord => Boolean(currency));
  const byCode = new Map<string, CurrencyRecord & { duplicate_guidfixeds: string[] }>();
  for (const currency of currencies) {
    const existing = byCode.get(currency.code);
    if (!existing) {
      byCode.set(currency.code, { ...currency, duplicate_guidfixeds: uniqueStrings([currency.guidfixed]) });
      continue;
    }

    existing.duplicate_guidfixeds = uniqueStrings([...existing.duplicate_guidfixeds, currency.guidfixed]);
    if ((existing.isdisabled && !currency.isdisabled) || (!existing.guidfixed && currency.guidfixed)) {
      byCode.set(currency.code, { ...currency, duplicate_guidfixeds: existing.duplicate_guidfixeds });
    }
  }
  return [...byCode.values()];
}

function normalizeCurrency(value: unknown): CurrencyRecord | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  const item = value as Record<string, unknown>;
  const code = toStringValue(item.code).trim().toUpperCase();
  if (!code) return null;

  return {
    guidfixed: toStringValue(item.guidfixed || item.guidfixed || item.guidFixed || item.id || item._id),
    code,
    name: toStringValue(item.name),
    symbol: toStringValue(item.symbol),
    isdisabled: Boolean(item.isdisabled),
    exchangerates: Array.isArray(item.exchangerates) ? item.exchangerates as ExchangeRateEntry[] : [],
  };
}

function uniqueStrings(values: unknown[]): string[] {
  return Array.from(new Set(values.map((value) => toStringValue(value).trim()).filter(Boolean)));
}

function toStringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function normalizeForm(form: CurrencyForm): CurrencyForm {
  return {
    code: form.code.trim().toUpperCase(),
    name: form.name.trim(),
    symbol: form.symbol.trim(),
    isdisabled: form.isdisabled,
  };
}

function validateForm(form: CurrencyForm, text: (key: CurrencyTextKey) => string): string | null {
  if (!form.code) return text("requiredCode");
  if (form.code.length < 2 || form.code.length > 5) return text("codeLength");
  if (!form.name) return text("requiredName");
  if (!form.symbol) return text("selectSymbol");
  if (!findCurrencySymbolPreset(form)) return text("currencyVerifyRequired");
  return null;
}

function getLocalCurrencyText(language: LanguageCode, key: CurrencyTextKey): string {
  return currencyText[language]?.[key] ?? currencyTextEn[key] ?? key;
}

function formatCount(value: number, language: LanguageCode): string {
  const localeByLanguage: Record<LanguageCode, string> = {
    th: "th-TH",
    en: "en-US",
    cn: "zh-CN",
    ja: "ja-JP",
    ko: "ko-KR",
    lo: "lo-LA",
    my: "my-MM",
    km: "km-KH",
    vi: "vi-VN",
    ms: "ms-MY",
    id: "id-ID",
    fil: "fil-PH",
  };
  return value.toLocaleString(localeByLanguage[language]);
}

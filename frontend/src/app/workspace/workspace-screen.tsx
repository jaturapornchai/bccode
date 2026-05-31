"use client";

import {
  AlertCircle,
  ArrowLeft,
  Building2,
  CalendarDays,
  CheckCircle2,
  Coins,
  Copy,
  Crown,
  ExternalLink,
  GitBranch,
  Languages,
  Loader2,
  LogOut,
  MessageCircle,
  Plus,
  Search,
  Store,
  UserRound,
} from "lucide-react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import QRCode from "qrcode";
import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
import {
  branchDisplayName,
  shopDisplayName,
  type AuthSession,
  type BranchListItem,
  type ShopListItem,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { LanguageDialog } from "../language-dialog";
import { ManualLink } from "../manual-link";
import { ThemeToggle } from "../theme-toggle";

type Step = "loading" | "shops" | "create" | "branches";
type Notice = { type: "success" | "error" | "info"; text?: string; textKey?: WorkspaceTextKey } | null;
type LineDialogState = {
  open: boolean;
  loading: boolean;
  code: string;
  loginUrl: string;
  qrDataUrl: string;
  expiresAt: string;
  error: string;
  expired: boolean;
};
type LineLinkStatusResponse = {
  success?: boolean;
  status?: "pending" | "success" | "failed" | "expired";
  message?: string;
};
type PendingUnitSetup = {
  shop: ShopListItem;
  branch: BranchListItem | null;
  shopInfo: Record<string, unknown> | null;
  selectedCodes: string[];
  units: ProductUnitOption[];
};
type ProductUnitOption = {
  unitcode: string;
  names?: { code?: string; name?: string }[];
};
type StandardProductUnitResponse = {
  success?: boolean;
  data?: ProductUnitOption[];
  total?: number;
  source?: string;
  message?: string;
};

const SOCIAL_POLL_TIMEOUT_MS = 5 * 60 * 1000;
const DEFAULT_SOCIAL_POLL_INTERVAL_MS = 2000;
const emptyLineDialog: LineDialogState = {
  open: false,
  loading: false,
  code: "",
  loginUrl: "",
  qrDataUrl: "",
  expiresAt: "",
  error: "",
  expired: false,
};

const workspaceTextEn = {
  backToCompanies: "Back to companies",
  changeCompany: "Change company",
  companyName: "Company name",
  companyNamePlaceholder: "Example Co., Ltd.",
  companyNameRequired: "Please enter a company name.",
  createCompany: "Create company",
  createCompanyNew: "Create new company",
  createCompanyRequiresGoogle: "Sign in with Google to create a new company.",
  createShopFailed: "Could not create company.",
  createShopSuccess: "Company created. Reloading list.",
  loadShopsFailed: "Could not load companies.",
  loadingCompanies: "Loading companies",
  logout: "Log out",
  noCompanies: "No company yet. Create a company before starting.",
  signedInAs: "Signed in as",
  availableCompanies: "Available companies",
  availableBranches: "Available branches",
  requestFailed: "Request failed.",
  searchBranch: "Search branches",
  searchCompany: "Search companies",
  selectBranchTitle: "Select branch",
  selectCompanyTitle: "Select company",
  selectShopFailed: "Could not select company.",
  staff: "Staff",
  stepBranch: "3 Branch",
  stepCompany: "2 Company",
  stepLogin: "1 Login",
  stepMenu: "4 Main menu",
  owner: "Owner",
} as const;

type WorkspaceTextKey = keyof typeof workspaceTextEn;

const workspaceBackendKeys: Partial<Record<WorkspaceTextKey, string>> = {
  changeCompany: "change_company",
  companyName: "company_name",
  createCompany: "create_company",
  createCompanyNew: "create_company_new",
  createCompanyRequiresGoogle: "create_company_requires_google",
  logout: "logout",
  owner: "owner",
  searchBranch: "search_branch",
  searchCompany: "search_company",
  selectBranchTitle: "select_branch",
  selectCompanyTitle: "select_company",
  staff: "staff",
  stepBranch: "step_branch",
  stepCompany: "step_company",
  stepLogin: "step_login",
  stepMenu: "step_main_menu",
};

const workspaceText: Partial<Record<LanguageCode, Record<WorkspaceTextKey, string>>> = {
  th: {
    backToCompanies: "กลับไปเลือกบริษัท",
    changeCompany: "เปลี่ยนบริษัท",
    companyName: "ชื่อบริษัท",
    companyNamePlaceholder: "เช่น บริษัท ตัวอย่าง จำกัด",
    companyNameRequired: "กรุณากรอกชื่อบริษัท",
    createCompany: "สร้างบริษัท",
    createCompanyNew: "สร้างบริษัทใหม่",
    createCompanyRequiresGoogle: "ต้องเข้าสู่ระบบด้วย Google เพื่อสร้างบริษัทใหม่",
    createShopFailed: "สร้างบริษัทไม่สำเร็จ",
    createShopSuccess: "สร้างบริษัทแล้ว กำลังโหลดรายการใหม่",
    loadShopsFailed: "โหลดบริษัทไม่สำเร็จ",
    loadingCompanies: "กำลังโหลดข้อมูลบริษัท",
    logout: "ออกจากระบบ",
    noCompanies: "ยังไม่มีบริษัท ให้สร้างบริษัทใหม่ก่อนเริ่มใช้งาน",
    signedInAs: "เข้าสู่ระบบเป็น",
    availableCompanies: "บริษัทที่เข้าได้",
    availableBranches: "สาขาที่เข้าได้",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    searchBranch: "ค้นหาสาขา",
    searchCompany: "ค้นหาบริษัท",
    selectBranchTitle: "เลือกสาขา",
    selectCompanyTitle: "เลือกบริษัท",
    selectShopFailed: "เลือกบริษัทไม่สำเร็จ",
    staff: "พนักงาน",
    stepBranch: "3 เลือกสาขา",
    stepCompany: "2 เลือกบริษัท",
    stepLogin: "1 เข้าสู่ระบบ",
    stepMenu: "4 เมนูหลัก",
    owner: "เจ้าของ",
  },
  en: workspaceTextEn,
  km: {
    backToCompanies: "ត្រឡប់ទៅជ្រើសក្រុមហ៊ុន",
    changeCompany: "ប្តូរក្រុមហ៊ុន",
    companyName: "ឈ្មោះក្រុមហ៊ុន",
    companyNamePlaceholder: "ឧទាហរណ៍ Example Co., Ltd.",
    companyNameRequired: "សូមបញ្ចូលឈ្មោះក្រុមហ៊ុន។",
    createCompany: "បង្កើតក្រុមហ៊ុន",
    createCompanyNew: "បង្កើតក្រុមហ៊ុនថ្មី",
    createCompanyRequiresGoogle: "សូមចូលដោយ Google ដើម្បីបង្កើតក្រុមហ៊ុនថ្មី។",
    createShopFailed: "មិនអាចបង្កើតក្រុមហ៊ុនបាន។",
    createShopSuccess: "បានបង្កើតក្រុមហ៊ុន កំពុងផ្ទុកបញ្ជីឡើងវិញ។",
    loadShopsFailed: "មិនអាចផ្ទុកក្រុមហ៊ុនបាន។",
    loadingCompanies: "កំពុងផ្ទុកក្រុមហ៊ុន",
    logout: "ចេញពីប្រព័ន្ធ",
    noCompanies: "មិនទាន់មានក្រុមហ៊ុន សូមបង្កើតក្រុមហ៊ុនមុនចាប់ផ្តើម។",
    signedInAs: "ចូលប្រើជា",
    availableCompanies: "ក្រុមហ៊ុនដែលអាចចូលបាន",
    availableBranches: "សាខាដែលអាចចូលបាន",
    requestFailed: "សំណើបរាជ័យ។",
    searchBranch: "ស្វែងរកសាខា",
    searchCompany: "ស្វែងរកក្រុមហ៊ុន",
    selectBranchTitle: "ជ្រើសសាខា",
    selectCompanyTitle: "ជ្រើសក្រុមហ៊ុន",
    selectShopFailed: "មិនអាចជ្រើសក្រុមហ៊ុនបាន។",
    staff: "បុគ្គលិក",
    stepBranch: "3 ជ្រើសសាខា",
    stepCompany: "2 ជ្រើសក្រុមហ៊ុន",
    stepLogin: "1 ចូលប្រើ",
    stepMenu: "4 ម៉ឺនុយ",
    owner: "ម្ចាស់",
  },
  vi: {
    backToCompanies: "Quay lại chọn công ty",
    changeCompany: "Đổi công ty",
    companyName: "Tên công ty",
    companyNamePlaceholder: "Ví dụ: Công ty TNHH Mẫu",
    companyNameRequired: "Vui lòng nhập tên công ty.",
    createCompany: "Tạo công ty",
    createCompanyNew: "Tạo công ty mới",
    createCompanyRequiresGoogle: "Đăng nhập bằng Google để tạo công ty mới.",
    createShopFailed: "Không thể tạo công ty.",
    createShopSuccess: "Đã tạo công ty. Đang tải lại danh sách.",
    loadShopsFailed: "Không thể tải danh sách công ty.",
    loadingCompanies: "Đang tải công ty",
    logout: "Đăng xuất",
    noCompanies: "Chưa có công ty. Hãy tạo công ty trước khi bắt đầu.",
    signedInAs: "Đăng nhập bằng",
    availableCompanies: "Công ty có thể truy cập",
    availableBranches: "Chi nhánh có thể truy cập",
    requestFailed: "Yêu cầu thất bại.",
    searchBranch: "Tìm chi nhánh",
    searchCompany: "Tìm công ty",
    selectBranchTitle: "Chọn chi nhánh",
    selectCompanyTitle: "Chọn công ty",
    selectShopFailed: "Không thể chọn công ty.",
    staff: "Nhân viên",
    stepBranch: "3 Chi nhánh",
    stepCompany: "2 Công ty",
    stepLogin: "1 Đăng nhập",
    stepMenu: "4 Menu chính",
    owner: "Chủ sở hữu",
  },
};

function wt(language: LanguageCode, key: WorkspaceTextKey): string {
  return workspaceText[language]?.[key] ?? workspaceTextEn[key] ?? key;
}

type WorkspaceScreenProps = {
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage?: LanguageCode;
};

export function WorkspaceScreen({ initialBackendLanguage, initialBackendUrl, initialLanguage = "th" }: WorkspaceScreenProps = {}) {
  const router = useRouter();
  const [language, setLanguage] = useState<LanguageCode>(initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [step, setStep] = useState<Step>("loading");
  const [shops, setShops] = useState<ShopListItem[]>([]);
  const [branches, setBranches] = useState<BranchListItem[]>([]);
  const [selectedShop, setSelectedShop] = useState<ShopListItem | null>(null);
  const [query, setQuery] = useState("");
  const [branchQuery, setBranchQuery] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);
  const [lineDialog, setLineDialog] = useState<LineDialogState>(emptyLineDialog);
  const [pendingUnitSetup, setPendingUnitSetup] = useState<PendingUnitSetup | null>(null);
  const [unitSetupSaving, setUnitSetupSaving] = useState(false);
  const linePollTimer = useRef<number | null>(null);
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);
  const text = useCallback(
    (key: WorkspaceTextKey) => {
      const backendKey = workspaceBackendKeys[key] ?? key;
      const fallback = wt(language, key);
      const resolved = backendText(backendLanguage, backendKey, fallback);
      return resolved === backendKey ? fallback : resolved;
    },
    [backendLanguage, language],
  );
  const connectLineText = backendText(backendLanguage, "connect_line", t(language, "loginWithLine"));
  const lineLinkDescription = backendText(backendLanguage, "scan_qr_with_line", t(language, "lineLoginDescription"));
  const lineLinkSuccessText = backendText(backendLanguage, "link_line_success", t(language, "loginSuccess"));
  const lineLinkWaitingText = backendText(backendLanguage, "waiting_for_link", t(language, "lineLoginWaiting"));
  const unitSetupTitle = backendText(backendLanguage, "product_unit_setup_required", language === "th" ? "ยังไม่มีหน่วยนับสินค้า" : "Product units are missing");
  const unitSetupDescription = backendText(
    backendLanguage,
    "product_unit_setup_description",
    language === "th"
      ? "บริษัทนี้ยังไม่มีหน่วยนับสินค้า ต้องการเพิ่มหน่วยนับเริ่มต้นอัตโนมัติหรือไม่"
      : "This company has no product units. Add default product units automatically?",
  );
  const unitSetupConfirmText = backendText(backendLanguage, "product_unit_setup_confirm", language === "th" ? "เพิ่มอัตโนมัติ" : "Add automatically");
  const unitSetupSkipText = backendText(backendLanguage, "product_unit_setup_skip", language === "th" ? "เข้าเมนูก่อน" : "Enter menu first");
  const unitSetupSelectedText = backendText(backendLanguage, "selected", language === "th" ? "เลือกแล้ว" : "Selected");
  const unitSetupSelectAllText = backendText(backendLanguage, "select_all", language === "th" ? "เลือกทั้งหมด" : "Select all");
  const unitSetupClearText = backendText(backendLanguage, "clear_selection", language === "th" ? "ล้างการเลือก" : "Clear");
  const canCreateCompany = canAuthCreateCompany(auth);

  const loadShops = useCallback(async (currentAuth: AuthSession | null) => {
    if (!currentAuth) return;
    setStep("loading");
    setNotice(null);
    try {
      const payload = await callWorkspaceApi<{ data?: ShopListItem[] }>(currentAuth, "shops");
      const nextShops = Array.isArray(payload.data) ? payload.data : [];
      setShops(nextShops);
      setStep("shops");
      if (nextShops.length === 0 && !canAuthCreateCompany(currentAuth)) {
        setNotice({ type: "info", textKey: "createCompanyRequiresGoogle" });
      }
    } catch (error) {
      setStep("shops");
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "loadShopsFailed" });
    }
  }, []);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? initialLanguage);
    setLanguage(savedLanguage);
    document.documentElement.lang = savedLanguage;

    const savedAuth = readAuth();
    if (!savedAuth) {
      router.replace("/");
      return;
    }

    setAuth(savedAuth);
    void loadShops(savedAuth);
  }, [initialLanguage, loadShops, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem("user_language", language);
  }, [language]);

  useEffect(() => {
    return () => stopLinePolling();
  }, []);

  const filteredShops = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return shops;
    return shops.filter((shop) => `${shopDisplayName(shop)} ${shop.shopid} ${shop.createdby ?? ""}`.toLowerCase().includes(needle));
  }, [query, shops]);

  const filteredBranches = useMemo(() => {
    const needle = branchQuery.trim().toLowerCase();
    if (!needle) return branches;
    return branches.filter((branch) => `${branchDisplayName(branch)} ${branch.code ?? ""}`.toLowerCase().includes(needle));
  }, [branchQuery, branches]);

  const signedInAs = auth?.profile?.email || auth?.username || "";
  const currentTitle = step === "branches" ? text("selectBranchTitle") : text("selectCompanyTitle");
  const currentSummary =
    step === "branches"
      ? `${text("availableBranches")}: ${filteredBranches.length}`
      : `${text("availableCompanies")}: ${filteredShops.length}`;

  function stopLinePolling() {
    if (linePollTimer.current !== null) {
      window.clearInterval(linePollTimer.current);
      linePollTimer.current = null;
    }
  }

  async function handleLineLink() {
    if (!auth) return;

    stopLinePolling();
    setNotice(null);
    setLineDialog({ ...emptyLineDialog, open: true, loading: true });

    try {
      const response = await fetch("/api/auth/line/code", { method: "POST" });
      const data = (await response.json()) as {
        success?: boolean;
        message?: string;
        code?: string;
        loginUrl?: string;
        expiresAt?: string;
      };

      if (!response.ok || !data.success || !data.code || !data.loginUrl) {
        throw new Error(data.message ?? text("requestFailed"));
      }

      const qrDataUrl = await QRCode.toDataURL(data.loginUrl, {
        margin: 1,
        width: 220,
        color: { dark: "#111827", light: "#ffffff" },
      });

      setLineDialog({
        open: true,
        loading: false,
        code: data.code,
        loginUrl: data.loginUrl,
        qrDataUrl,
        expiresAt: data.expiresAt ?? "",
        error: "",
        expired: false,
      });
      startLinePolling(data.code, data.expiresAt ?? "");
    } catch (error) {
      setLineDialog((current) => ({
        ...current,
        loading: false,
        error: error instanceof Error ? error.message : text("requestFailed"),
      }));
    }
  }

  function startLinePolling(code: string, expiresAt: string) {
    const startedAt = Date.now();
    linePollTimer.current = window.setInterval(() => {
      const isExpiredByTime = expiresAt ? Date.now() > Date.parse(expiresAt) : false;
      if (Date.now() - startedAt > SOCIAL_POLL_TIMEOUT_MS || isExpiredByTime) {
        stopLinePolling();
        setLineDialog((current) => ({ ...current, expired: true }));
        setNotice({ type: "error", text: t(language, "lineLoginExpired") });
        return;
      }
      void pollLineLink(code);
    }, DEFAULT_SOCIAL_POLL_INTERVAL_MS);
  }

  async function pollLineLink(code: string) {
    if (!auth) return;

    const response = await fetch("/api/auth/line/link/status", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ backendUrl: auth.backendUrl, code }),
    });
    const data = (await response.json()) as LineLinkStatusResponse;

    if (!response.ok || data.success === false || data.status === "failed" || data.status === "expired") {
      stopLinePolling();
      const message = data.message ?? text("requestFailed");
      setNotice({ type: "error", text: message });
      setLineDialog((current) => ({ ...current, error: message }));
      return;
    }

    if (data.status !== "success") return;

    stopLinePolling();
    setLineDialog(emptyLineDialog);
    setNotice({ type: "success", text: lineLinkSuccessText });
  }

  function closeLineDialog() {
    stopLinePolling();
    setLineDialog(emptyLineDialog);
  }

  async function copyLineLoginUrl() {
    if (!lineDialog.loginUrl) return;
    await navigator.clipboard.writeText(lineDialog.loginUrl);
    setNotice({ type: "success", text: t(language, "copied") });
  }

  async function selectShop(shop: ShopListItem) {
    if (!auth) return;
    setBusy(true);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "select-shop", {
        method: "POST",
        body: { shopid: shop.shopid },
      });
      const shopInfo = await callWorkspaceApi<{ data?: Record<string, unknown> }>(auth, `shop-info?shopid=${encodeURIComponent(shop.shopid)}`);
      const branchPayload = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
      let nextBranches = Array.isArray(branchPayload.data) ? branchPayload.data : [];

      if (nextBranches.length === 0) {
        const created = await callWorkspaceApi<{ data?: BranchListItem; id?: string; ID?: string }>(auth, "branch", {
          method: "POST",
          body: { branch: createDefaultBranch(shop, shopInfo.data ?? null) },
        });
        if (created.data) {
          nextBranches = [created.data];
        } else {
          const reloaded = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
          nextBranches = Array.isArray(reloaded.data) ? reloaded.data : [];
          if (nextBranches.length === 0) nextBranches = [createDefaultBranchListItem(created.id ?? created.ID)];
        }
      }

      setSelectedShop(shop);
      setBranches(nextBranches);
      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));

      setStep("branches");
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "selectShopFailed" });
    } finally {
      setBusy(false);
    }
  }

  async function createShop(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!auth) return;
    if (!canCreateCompany) {
      setNotice({ type: "error", textKey: "createCompanyRequiresGoogle" });
      setStep("shops");
      return;
    }
    const name = companyName.trim();
    if (!name) {
      setNotice({ type: "error", textKey: "companyNameRequired" });
      return;
    }

    setBusy(true);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "create-shop", {
        method: "POST",
        body: createShopPayload(name),
      });
      setCompanyName("");
      setNotice({ type: "success", textKey: "createShopSuccess" });
      await loadShops(auth);
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "createShopFailed" });
    } finally {
      setBusy(false);
    }
  }

  async function selectBranch(branch: BranchListItem) {
    if (!selectedShop || !auth) return;
    setBusy(true);
    setNotice(null);
    const shopInfoRaw = localStorage.getItem(workspaceStorageKeys.shopInfo);
    const shopInfo = parseShopInfo(shopInfoRaw);
    try {
      await enterWorkspaceWithUnitCheck(selectedShop, branch, shopInfo);
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", text: text("requestFailed") });
    } finally {
      setBusy(false);
    }
  }

  async function enterWorkspaceWithUnitCheck(shop: ShopListItem, branch: BranchListItem | null, shopInfo: Record<string, unknown> | null) {
    if (!auth) return;
    persistWorkspace(shop, branch, shopInfo);
    const payload = await callWorkspaceApi<{ data?: unknown[]; total?: number }>(auth, "product-units?offset=0&limit=1&q=");

    if (getApiTotal(payload) > 0) {
      router.push("/menu");
      return;
    }

    const mainShopId = getMainShopId(shopInfo);
    const standardUnits = await callWorkspaceApi<StandardProductUnitResponse>(
      auth,
      `product-units/standard?mainShopId=${encodeURIComponent(mainShopId)}&q=`,
    );
    const units = Array.isArray(standardUnits.data) ? standardUnits.data : [];
    setPendingUnitSetup({
      shop,
      branch,
      shopInfo,
      units,
      selectedCodes: units.map((unit) => unit.unitcode),
    });
    setNotice({ type: "info", text: unitSetupTitle });
  }

  async function confirmUnitSetup() {
    if (!auth || !pendingUnitSetup) return;
    setUnitSetupSaving(true);
    setNotice(null);
    try {
      const payload = await callWorkspaceApi<{ success?: boolean; message?: string; data?: Record<string, unknown> }>(auth, "product-units/defaults", {
        method: "POST",
        body: { mainShopId: getMainShopId(pendingUnitSetup.shopInfo), unitcodes: pendingUnitSetup.selectedCodes },
      });
      if (payload.success === false) throw new Error(payload.message ?? text("requestFailed"));
      setPendingUnitSetup(null);
      setNotice({ type: "success", text: payload.message ?? unitSetupConfirmText });
      router.push("/menu");
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", text: text("requestFailed") });
    } finally {
      setUnitSetupSaving(false);
    }
  }

  function skipUnitSetup() {
    setPendingUnitSetup(null);
    router.push("/menu");
  }

  function togglePendingUnit(unitcode: string, checked: boolean) {
    setPendingUnitSetup((current) => {
      if (!current) return current;
      const selected = new Set(current.selectedCodes);
      if (checked) selected.add(unitcode);
      else selected.delete(unitcode);
      return { ...current, selectedCodes: Array.from(selected) };
    });
  }

  function selectAllPendingUnits() {
    setPendingUnitSetup((current) => current ? { ...current, selectedCodes: current.units.map((unit) => unit.unitcode) } : current);
  }

  function clearPendingUnits() {
    setPendingUnitSetup((current) => current ? { ...current, selectedCodes: [] } : current);
  }

  function logout() {
    localStorage.removeItem(workspaceStorageKeys.auth);
    localStorage.removeItem(workspaceStorageKeys.workspace);
    localStorage.removeItem(workspaceStorageKeys.shopInfo);
    localStorage.removeItem(workspaceStorageKeys.branch);
    router.replace("/");
  }

  return (
    <main className="workspace-page">
      <header className="workspace-topbar">
        <div className="workspace-title">
          <span className="workspace-mark"><Building2 size={18} /></span>
          <div>
            <p>BC Ai Account</p>
            <h1>{currentTitle}</h1>
            {signedInAs ? <small>{text("signedInAs")}: {signedInAs}</small> : null}
          </div>
        </div>
        <div className="workspace-actions">
          <ManualLink compact language={language} screen="workspace" />
          <button
            className="icon-button line-link-action"
            type="button"
            onClick={() => void handleLineLink()}
            disabled={!auth || lineDialog.loading}
            aria-label={connectLineText}
            title={connectLineText}
          >
            {lineDialog.loading ? <Loader2 className="spin" size={18} /> : <MessageCircle size={18} />}
          </button>
          <ThemeToggle language={language} />
          <LanguageDialog language={language} onLanguageChange={setLanguage} />
          <button className="icon-button" type="button" onClick={logout} aria-label={text("logout")} title={text("logout")}>
            <LogOut size={18} />
          </button>
        </div>
      </header>

      <section className="workspace-panel">
        <div className="workspace-panel-head">
          <div>
            <h2>{currentTitle}</h2>
            <p>{currentSummary}</p>
          </div>
          {step === "shops" && canCreateCompany ? (
            <button className="primary-button workspace-head-action" type="button" onClick={() => setStep("create")}>
              <Plus size={17} /> {text("createCompanyNew")}
            </button>
          ) : null}
        </div>

        <div className="step-strip">
          <span className="active">{text("stepLogin")}</span>
          <span className={step === "shops" || step === "branches" || step === "create" ? "active" : ""}>{text("stepCompany")}</span>
          <span className={step === "branches" ? "active" : ""}>{text("stepBranch")}</span>
          <span>{text("stepMenu")}</span>
        </div>

        {notice ? (
          <div className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"}`}>
            {notice.type === "success" ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
            <span>{notice.text ?? (notice.textKey ? text(notice.textKey) : "")}</span>
          </div>
        ) : null}

        {step === "loading" ? (
          <div className="loading-state"><Loader2 className="spin" size={24} /> {text("loadingCompanies")}</div>
        ) : null}

        {step === "shops" ? (
          <>
            <div className="workspace-toolbar">
              <label className="search-shell">
                <Search size={17} />
                <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={text("searchCompany")} />
              </label>
            </div>
            {filteredShops.length === 0 ? (
              <div className="workspace-empty-state">
                <Store size={24} />
                <strong>{text("noCompanies")}</strong>
              </div>
            ) : (
              <div className="workspace-card-grid">
                {filteredShops.map((shop, index) => {
                const isCreator = shop.is_creator === true
                  || Boolean(auth?.username && shop.createdby && shop.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase());
                const languageCodes = shopLanguageCodes(shop);
                const currencyLabel = shopCurrencyLabel(shop, language);
                const dateFormat = shopDateFormatLabel(shop, language);
                return (
                  <button className="shop-card" disabled={busy} key={shop.shopid} type="button" onClick={() => void selectShop(shop)}>
                    <span className={`shop-avatar tone-${index % 6}`}><Store size={20} /></span>
                    <span className="shop-main">
                      <strong>{shopDisplayName(shop)}</strong>
                      <div className="shop-details-row">
                        <small>{shop.shopid}</small>
                        {shop.createdby ? (
                          <span className="shop-creator">
                            <UserRound size={12} />
                            {language === "th" ? `ผู้สร้าง: ${shop.createdby}` : `Creator: ${shop.createdby}`}
                          </span>
                        ) : null}
                      </div>
                      <span className="shop-config-row">
                        <span className="shop-config-chip">
                          <Languages size={12} />
                          {language === "th" ? "ภาษา" : "Languages"}: {languageCodes.join(", ")}
                        </span>
                        <span className="shop-config-chip">
                          <Coins size={12} />
                          {language === "th" ? "สกุลเงิน" : "Currency"}: {currencyLabel}
                        </span>
                        <span className="shop-config-chip">
                          <CalendarDays size={12} />
                          {language === "th" ? "วันที่" : "Date"}: {dateFormat}
                        </span>
                      </span>
                    </span>
                    <span className="shop-badges">
                      <b className={isCreator ? "creator-badge" : "user-badge"}>
                        {isCreator ? <Crown size={13} /> : <UserRound size={13} />}
                        {isCreator ? "OWNER" : "USER"}
                      </b>
                      {shop.branchcode ? <em>{shop.branchcode}</em> : null}
                    </span>
                  </button>
                );
                })}
              </div>
            )}
          </>
        ) : null}

        {step === "create" ? (
          <form className="create-company-form" onSubmit={createShop}>
            <button className="text-action compact-action" type="button" onClick={() => setStep("shops")}>
              <ArrowLeft size={16} /> {text("backToCompanies")}
            </button>
            <label className="field-group">
              <span>{text("companyName")}</span>
              <div className="input-shell">
                <Building2 size={18} />
                <input value={companyName} onChange={(event) => setCompanyName(event.target.value)} placeholder={text("companyNamePlaceholder")} />
              </div>
            </label>
            <button className="primary-button" type="submit" disabled={busy}>
              {busy ? <Loader2 className="spin" size={18} /> : <Plus size={18} />}
              <span>{text("createCompany")}</span>
            </button>
          </form>
        ) : null}

        {step === "branches" ? (
          <>
            <div className="workspace-toolbar">
              <button className="text-action compact-action" type="button" onClick={() => setStep("shops")}>
                <ArrowLeft size={16} /> {text("changeCompany")}
              </button>
              <label className="search-shell">
                <Search size={17} />
                <input value={branchQuery} onChange={(event) => setBranchQuery(event.target.value)} placeholder={text("searchBranch")} />
              </label>
            </div>
            {filteredBranches.length === 0 ? (
              <div className="workspace-empty-state">
                <GitBranch size={24} />
                <strong>{text("searchBranch")}</strong>
              </div>
            ) : (
              <div className="workspace-card-grid">
                {filteredBranches.map((branch, index) => (
                <button className="shop-card branch-card" disabled={busy} key={branch.guid_fixed || branch.code} type="button" onClick={() => void selectBranch(branch)}>
                  <span className={`shop-avatar tone-${index % 6}`}><GitBranch size={20} /></span>
                  <span className="shop-main">
                    <strong>{branchDisplayName(branch)}</strong>
                    <small>{branch.code || branch.guid_fixed}</small>
                  </span>
                  <span className="shop-badges">
                    {branch.base_currency ? <b>{branch.base_currency}</b> : null}
                    {branch.language ? <em>{branch.language.toUpperCase()}</em> : null}
                  </span>
                </button>
                ))}
              </div>
            )}
          </>
        ) : null}
      </section>

      {lineDialog.open ? (
        <div className="dialog-backdrop" role="presentation">
          <section className="line-login-dialog" aria-label={connectLineText} role="dialog" aria-modal="true">
            <div className="dialog-header">
              <div>
                <p className="eyebrow">LINE</p>
                <h2>{connectLineText}</h2>
              </div>
              <button className="icon-button dialog-close" type="button" onClick={closeLineDialog} aria-label={t(language, "lineLoginClose")}>
                ×
              </button>
            </div>

            {lineDialog.loading ? (
              <div className="line-login-status">
                <Loader2 className="spin" aria-hidden="true" size={28} />
                <span>{lineLinkWaitingText}</span>
              </div>
            ) : lineDialog.error ? (
              <div className="message error">
                <AlertCircle size={18} />
                <span>{lineDialog.error}</span>
              </div>
            ) : lineDialog.expired ? (
              <div className="message error">
                <AlertCircle size={18} />
                <span>{t(language, "lineLoginExpired")}</span>
              </div>
            ) : (
              <>
                <p className="line-login-description">{lineLinkDescription}</p>
                <div className="line-qr-box">
                  {lineDialog.qrDataUrl ? <Image alt={connectLineText} height={220} src={lineDialog.qrDataUrl} unoptimized width={220} /> : null}
                </div>
                <div className="line-code">
                  <span>{t(language, "lineLoginCode")}</span>
                  <strong>{lineDialog.code}</strong>
                </div>
                <div className="line-dialog-actions">
                  <a className="secondary-button" href={lineDialog.loginUrl} target="_blank" rel="noreferrer">
                    <ExternalLink aria-hidden="true" size={17} />
                    <span>{t(language, "lineLoginOpen")}</span>
                  </a>
                  <button className="secondary-button" type="button" onClick={copyLineLoginUrl}>
                    <Copy aria-hidden="true" size={17} />
                    <span>{t(language, "lineLoginCopy")}</span>
                  </button>
                </div>
                <div className="line-login-status">
                  <Loader2 className="spin" aria-hidden="true" size={16} />
                  <span>{lineLinkWaitingText}</span>
                </div>
              </>
            )}

            <div className="line-dialog-actions">
              <button className="secondary-button" type="button" onClick={() => void handleLineLink()}>
                {t(language, "lineLoginCreateNew")}
              </button>
              <button className="secondary-button" type="button" onClick={closeLineDialog}>
                {t(language, "lineLoginClose")}
              </button>
            </div>
          </section>
        </div>
      ) : null}

      {pendingUnitSetup ? (
        <div className="dialog-backdrop" role="presentation">
          <section className="line-login-dialog" aria-label={unitSetupTitle} role="dialog" aria-modal="true">
            <div className="dialog-header">
              <div>
                <p className="eyebrow">PRODUCT UNIT</p>
                <h2>{unitSetupTitle}</h2>
              </div>
              <button className="icon-button dialog-close" type="button" onClick={skipUnitSetup} aria-label={unitSetupSkipText} disabled={unitSetupSaving}>
                ×
              </button>
            </div>
            <div className="message info">
              <AlertCircle size={18} />
              <span>{unitSetupDescription}</span>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2 text-sm font-semibold">
              <span>{unitSetupSelectedText}: {pendingUnitSetup.selectedCodes.length.toLocaleString(localeOf(language))}/{pendingUnitSetup.units.length.toLocaleString(localeOf(language))}</span>
              <span className="line-dialog-actions">
                <button className="secondary-button" type="button" onClick={selectAllPendingUnits} disabled={unitSetupSaving || pendingUnitSetup.units.length === 0}>
                  {unitSetupSelectAllText}
                </button>
                <button className="secondary-button" type="button" onClick={clearPendingUnits} disabled={unitSetupSaving || pendingUnitSetup.selectedCodes.length === 0}>
                  {unitSetupClearText}
                </button>
              </span>
            </div>
            <div className="grid max-h-[38dvh] gap-2 overflow-y-auto pr-1">
              {pendingUnitSetup.units.length ? pendingUnitSetup.units.map((unit) => {
                const checked = pendingUnitSetup.selectedCodes.includes(unit.unitcode);
                return (
                  <label className="flex min-w-0 cursor-pointer items-center gap-2 rounded-xl border border-border bg-background px-3 py-2 text-sm font-semibold" key={unit.unitcode}>
                    <input
                      className="size-4 shrink-0 accent-primary"
                      type="checkbox"
                      checked={checked}
                      disabled={unitSetupSaving}
                      onChange={(event) => togglePendingUnit(unit.unitcode, event.target.checked)}
                    />
                    <span className="min-w-0 flex-1 truncate">{unitDisplayName(unit, language)}</span>
                    <b className="shrink-0 text-xs text-muted-foreground">{unit.unitcode}</b>
                  </label>
                );
              }) : (
                <p className="rounded-xl border border-border bg-background px-3 py-2 text-sm font-semibold text-muted-foreground">
                  {backendText(backendLanguage, "no_standard_product_unit", language === "th" ? "ไม่พบหน่วยนับมาตรฐานที่เพิ่มได้" : "No standard product units are available.")}
                </p>
              )}
            </div>
            <div className="line-dialog-actions">
              <button className="secondary-button" type="button" onClick={skipUnitSetup} disabled={unitSetupSaving}>
                {unitSetupSkipText}
              </button>
              <button className="primary-button" type="button" onClick={() => void confirmUnitSetup()} disabled={unitSetupSaving || pendingUnitSetup.selectedCodes.length === 0}>
                {unitSetupSaving ? <Loader2 className="spin" aria-hidden="true" size={18} /> : <Plus aria-hidden="true" size={18} />}
                <span>{unitSetupConfirmText}</span>
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </main>
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

function canAuthCreateCompany(auth: AuthSession | null): boolean {
  return auth?.method === "google";
}

async function callWorkspaceApi<T extends Record<string, unknown>>(
  auth: AuthSession,
  path: string,
  init?: { method?: "GET" | "POST"; body?: Record<string, unknown> },
): Promise<T> {
  const response = await fetch(`/api/workspace/${path}`, {
    method: init?.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      "x-bc-backend-url": auth.backendUrl,
      Authorization: `Bearer ${auth.token}`,
    },
    body: init?.method === "POST" ? JSON.stringify({ backendUrl: auth.backendUrl, ...(init.body ?? {}) }) : undefined,
    cache: "no-store",
  });
  const data = await response.json() as T & { success?: boolean; message?: string };
  if (!response.ok || data.success === false) {
    throw new Error(data.message ?? "");
  }
  return data;
}

function persistWorkspace(shop: ShopListItem, branch: BranchListItem | null, shopInfo: Record<string, unknown> | null) {
  localStorage.setItem(workspaceStorageKeys.workspace, JSON.stringify({ shop, branch, shopInfo }));
  localStorage.setItem(workspaceStorageKeys.branch, JSON.stringify(branch));
}

function getApiTotal(payload: { data?: unknown[]; total?: number }): number {
  if (typeof payload.total === "number") return payload.total;
  return Array.isArray(payload.data) ? payload.data.length : 0;
}

function getMainShopId(shopInfo: Record<string, unknown> | null): string {
  if (!shopInfo) return "";
  const mainShopId = shopInfo.main_shop_id ?? shopInfo.mainshopid ?? shopInfo.mainShopId;
  return typeof mainShopId === "string" ? mainShopId.trim() : "";
}

function unitDisplayName(unit: ProductUnitOption, language: LanguageCode): string {
  const names = Array.isArray(unit.names) ? unit.names : [];
  return names.find((name) => name.code?.toLowerCase() === language && name.name?.trim())?.name?.trim()
    ?? names.find((name) => name.code?.toLowerCase() === "th" && name.name?.trim())?.name?.trim()
    ?? names.find((name) => name.name?.trim())?.name?.trim()
    ?? unit.unitcode;
}

function localeOf(language: LanguageCode): string {
  return language === "th" ? "th-TH" : "en-US";
}

function shopLanguageCodes(shop: ShopListItem): string[] {
  return normalizedCodeList(shop.active_languages, ["th"]).map((code) => code.toUpperCase());
}

function shopCurrencyCodes(shop: ShopListItem): string[] {
  return normalizedCodeList(shop.currencies, shop.base_currency ? [shop.base_currency] : ["THB"]).map((code) => code.toUpperCase());
}

function shopCurrencyLabel(shop: ShopListItem, language: LanguageCode): string {
  const currencies = shopCurrencyCodes(shop);
  const hasConfiguredCurrency = (Array.isArray(shop.currencies) && shop.currencies.length > 0) || Boolean(stringValue(shop.base_currency));
  const defaultMarker = hasConfiguredCurrency ? "" : language === "th" ? " ค่าเริ่มต้น" : " default";
  return `${currencies.join(", ")}${defaultMarker}`;
}

function shopDateFormatLabel(shop: ShopListItem, language: LanguageCode): string {
  const hasConfiguredDateFormat = Boolean(stringValue(shop.date_format));
  const format = stringValue(shop.date_format) || "dd/MM/yyyy";
  const yearType = hasConfiguredDateFormat ? normalizedYearType(shop) : "buddhist";
  const yearLabel = language === "th"
    ? yearType === "buddhist" ? "พ.ศ." : "ค.ศ."
    : yearType === "buddhist" ? "BE" : "CE";
  const defaultMarker = hasConfiguredDateFormat ? "" : language === "th" ? " ค่าเริ่มต้น" : " default";
  return `${format} ${yearLabel}${defaultMarker}`;
}

function normalizedYearType(shop: ShopListItem): "buddhist" | "christian" {
  const yearType = stringValue(shop.year_type).toLowerCase();
  if (["buddhist", "be", "พ.ศ."].includes(yearType)) return "buddhist";
  if (["christian", "ce", "ค.ศ."].includes(yearType)) return "christian";
  if (typeof shop.usebuddhistcalendar === "boolean") return shop.usebuddhistcalendar ? "buddhist" : "christian";
  return "buddhist";
}

function normalizedCodeList(value: unknown, fallback: string[]): string[] {
  const source: unknown[] = Array.isArray(value) && value.length > 0 ? value : fallback;
  return Array.from(
    new Set(
      source
        .map((item) => stringValue(item))
        .map((item) => item.trim())
        .filter(Boolean),
    ),
  );
}

function parseShopInfo(raw: string | null): Record<string, unknown> | null {
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as unknown;
    return parsed && typeof parsed === "object" && !Array.isArray(parsed) ? parsed as Record<string, unknown> : null;
  } catch {
    return null;
  }
}

function createDefaultBranch(shop?: ShopListItem, shopInfo?: Record<string, unknown> | null): Record<string, unknown> {
  const settings = recordValue(shopInfo?.settings);
  const languageCodes = activeLanguageCodes(settings);
  const companyNames = normalizedNames(shopInfo?.names, shop ? shopDisplayName(shop) : "");
  const companyAddress = normalizedNames(shopInfo?.address, "");
  return {
    guid_fixed: "",
    code: "00000",
    companynames: companyNames,
    names: defaultBranchNames(languageCodes),
    departments: [],
    languages: languageCodes,
    contact: {
      address: companyAddress,
      country_code: stringValue(settings?.country_code) || "TH",
      province_code: "",
      district_code: "",
      sub_district_code: "",
      zip_code: "",
      phone_number: stringValue(shopInfo?.telephone),
      latitude: numberValue(settings?.latitude),
      longitude: numberValue(settings?.longitude),
    },
    imageuri: "",
    logouri: stringValue(shopInfo?.logo),
    pos: {
      tax_id: stringValue(settings?.tax_id),
      isbom: false,
      vatrate: numberValue(settings?.vatrate),
      vattypesale: numberValue(settings?.vattypesale),
      vattypepurchase: numberValue(settings?.vattypepurchase),
      inquirytypesale: numberValue(settings?.inquirytypesale),
      inquirytypepurchase: numberValue(settings?.inquirytypepurchase),
      headerreceiptpos: "",
      footerreceiptpos: "",
    },
    businesstype: {},
    paymentrounding: createDefaultPaymentRounding(),
    pointconfig: { generalrules: [], specialrules: [], pointusagetype: 1 },
    machinetype: 0,
    couponusetype: 0,
    company_registration_no: stringValue(settings?.company_registration_no),
    is_vat_registered: booleanValue(settings?.is_vat_registered),
    base_currency: stringValue(settings?.base_currency) || "THB",
    language: languageCodes[0] ?? "th",
    timezone: stringValue(settings?.timezone) || "Asia/Bangkok",
    timezone_offset: stringValue(settings?.timezone_offset) || "+07:00",
    timezone_label: stringValue(settings?.timezone_label) || "(UTC+07:00) Bangkok",
    date_format: stringValue(settings?.date_format) || "dd/MM/yyyy",
    year_type: booleanValue(settings?.usebuddhistcalendar, true) ? "buddhist" : "christian",
    decimal_quantity: numberValue(settings?.decimal_quantity, 2),
    decimal_price: numberValue(settings?.decimal_price, 2),
    decimal_document: numberValue(settings?.decimal_document, 2),
    is_restaurant: false,
    is_tire: false,
    is_agriculture: false,
    is_pharmacy: false,
    is_retail: false,
    is_service: false,
    is_wholesale: false,
    is_manufacturing: false,
    is_import_export: false,
    is_contractor: false,
    is_rental: false,
    is_ecommerce: false,
    is_logistics: false,
    is_education: false,
    is_hotel: false,
    is_beauty: false,
    is_gold_shop: false,
    is_accounting_firm: false,
    is_construction: false,
    is_electronics: false,
    is_mobile_shop: false,
  };
}

function activeLanguageCodes(settings: Record<string, unknown> | null): string[] {
  const configs = Array.isArray(settings?.languageconfigs) ? settings.languageconfigs : [];
  const codes = configs
    .filter((item) => !recordValue(item) || recordValue(item)?.is_use !== false)
    .map((item) => stringValue(recordValue(item)?.code).toLowerCase())
    .filter(Boolean);
  return codes.length > 0 ? Array.from(new Set(codes)) : ["th"];
}

function defaultBranchNames(languageCodes: string[]): Array<{ code: string; name: string }> {
  return languageCodes.map((code) => ({ code, name: code === "th" ? "สำนักงานใหญ่" : "" }));
}

function normalizedNames(value: unknown, fallbackName: string): Array<{ code: string; name: string }> {
  if (Array.isArray(value)) {
    const names = value
      .map((item) => recordValue(item))
      .filter((item): item is Record<string, unknown> => Boolean(item))
      .map((item) => ({ code: stringValue(item.code).toLowerCase(), name: stringValue(item.name) }))
      .filter((item) => item.code && item.name);
    if (names.length > 0) return names;
  }
  const fallback = fallbackName.trim();
  return fallback ? [{ code: "th", name: fallback }] : [];
}

function recordValue(value: unknown): Record<string, unknown> | null {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : null;
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function numberValue(value: unknown, fallback = 0): number {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return fallback;
}

function booleanValue(value: unknown, fallback = false): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") return ["true", "1", "yes", "y", "buddhist"].includes(value.trim().toLowerCase());
  return fallback;
}

function createDefaultBranchListItem(guidFixed: unknown): BranchListItem {
  return {
    guid_fixed: typeof guidFixed === "string" && guidFixed.trim() ? guidFixed.trim() : "00000",
    code: "00000",
    names: [{ code: "th", name: "สำนักงานใหญ่" }],
    base_currency: "THB",
    language: "th",
    timezone: "Asia/Bangkok",
    timezone_offset: "+07:00",
    timezone_label: "(UTC+07:00) Bangkok",
    year_type: "buddhist",
  };
}

function createDefaultPaymentRounding(): Record<string, unknown> {
  const defaultRules = [
    { lowerbound: 0.01, upperbound: 0.12, roundto: 0 },
    { lowerbound: 0.13, upperbound: 0.37, roundto: 0.25 },
    { lowerbound: 0.38, upperbound: 0.62, roundto: 0.5 },
    { lowerbound: 0.63, upperbound: 0.87, roundto: 0.75 },
    { lowerbound: 0.88, upperbound: 0.99, roundto: 1 },
  ];
  const method = { enabled: true, rules: defaultRules };
  return {
    banktransfer: method,
    cash: method,
    cheque: method,
    coupon: method,
    creditcard: method,
    delivery: method,
    qrcode: method,
  };
}

function createShopPayload(name: string): Record<string, unknown> {
  return {
    address: [],
    branchcode: "",
    images: [],
    logo: "",
    name1: name,
    names: [{ code: "th", name }],
    profilepicture: "",
    settings: {
      emailowners: [],
      emailstaffs: [],
      isusebranch: false,
      isusedepartment: false,
      language: "th",
      languageconfigs: [{ code: "th", codetranslator: "th", name: "ภาษาไทย", is_use: true, isdefault: true }],
      latitude: 0,
      longitude: 0,
      tax_id: "",
      vatrate: 7,
      vattypesale: 0,
      vattypepurchase: 0,
      inquirytypesale: 0,
      inquirytypepurchase: 0,
    },
    telephone: "",
    businesstype: {},
  };
}


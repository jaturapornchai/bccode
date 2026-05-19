"use client";

import {
  AlertCircle,
  ArrowLeft,
  Building2,
  CheckCircle2,
  Copy,
  Crown,
  ExternalLink,
  GitBranch,
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
    backToCompanies: "กลับไปเลือกกิจการ",
    changeCompany: "เปลี่ยนกิจการ",
    companyName: "ชื่อกิจการ",
    companyNamePlaceholder: "เช่น บริษัท ตัวอย่าง จำกัด",
    companyNameRequired: "กรุณากรอกชื่อกิจการ",
    createCompany: "สร้างกิจการ",
    createCompanyNew: "สร้างกิจการใหม่",
    createCompanyRequiresGoogle: "ต้องเข้าสู่ระบบด้วย Google เพื่อสร้างกิจการใหม่",
    createShopFailed: "สร้างกิจการไม่สำเร็จ",
    createShopSuccess: "สร้างกิจการแล้ว กำลังโหลดรายการใหม่",
    loadShopsFailed: "โหลดกิจการไม่สำเร็จ",
    loadingCompanies: "กำลังโหลดข้อมูลกิจการ",
    logout: "ออกจากระบบ",
    noCompanies: "ยังไม่มีกิจการ ให้สร้างกิจการใหม่ก่อนเริ่มใช้งาน",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    searchBranch: "ค้นหาสาขา",
    searchCompany: "ค้นหากิจการ",
    selectBranchTitle: "เลือกสาขา",
    selectCompanyTitle: "เลือกกิจการ",
    selectShopFailed: "เลือกกิจการไม่สำเร็จ",
    staff: "พนักงาน",
    stepBranch: "3 เลือกสาขา",
    stepCompany: "2 เลือกกิจการ",
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
  const linePollTimer = useRef<number | null>(null);
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);
  const text = useCallback(
    (key: WorkspaceTextKey) => backendText(backendLanguage, workspaceBackendKeys[key] ?? key, wt(language, key)),
    [backendLanguage, language],
  );
  const connectLineText = backendText(backendLanguage, "connect_line", t(language, "loginWithLine"));
  const lineLinkDescription = backendText(backendLanguage, "scan_qr_with_line", t(language, "lineLoginDescription"));
  const lineLinkSuccessText = backendText(backendLanguage, "link_line_success", t(language, "loginSuccess"));
  const lineLinkWaitingText = backendText(backendLanguage, "waiting_for_link", t(language, "lineLoginWaiting"));
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
      if (nextShops.length === 0) {
        setNotice({ type: "info", textKey: canAuthCreateCompany(currentAuth) ? "noCompanies" : "createCompanyRequiresGoogle" });
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
        const created = await callWorkspaceApi<{ data?: BranchListItem }>(auth, "branch", {
          method: "POST",
          body: { branch: createDefaultBranch() },
        });
        if (created.data) nextBranches = [created.data];
      }

      setSelectedShop(shop);
      setBranches(nextBranches);
      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));

      if (nextBranches.length <= 1) {
        persistWorkspace(shop, nextBranches[0] ?? null, shopInfo.data ?? null);
        router.push("/menu");
        return;
      }

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

  function selectBranch(branch: BranchListItem) {
    if (!selectedShop) return;
    const shopInfoRaw = localStorage.getItem(workspaceStorageKeys.shopInfo);
    const shopInfo = parseShopInfo(shopInfoRaw);
    persistWorkspace(selectedShop, branch, shopInfo);
    router.push("/menu");
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
            <h1>{step === "branches" ? text("selectBranchTitle") : text("selectCompanyTitle")}</h1>
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
              {canCreateCompany ? (
                <button className="secondary-button" type="button" onClick={() => setStep("create")}>
                  <Plus size={17} /> {text("createCompanyNew")}
                </button>
              ) : null}
            </div>
            <div className="workspace-card-grid">
              {filteredShops.map((shop, index) => {
                const isCreator = shop.iscreator === true
                  || Boolean(auth?.username && shop.createdby && shop.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase());
                return (
                  <button className="shop-card" disabled={busy} key={shop.shopid} type="button" onClick={() => void selectShop(shop)}>
                    <span className={`shop-avatar tone-${index % 6}`}><Store size={20} /></span>
                    <span className="shop-main">
                      <strong>{shopDisplayName(shop)}</strong>
                      <small>{shop.shopid}</small>
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
            <div className="workspace-card-grid">
              {filteredBranches.map((branch, index) => (
                <button className="shop-card branch-card" key={branch.guidfixed || branch.code} type="button" onClick={() => selectBranch(branch)}>
                  <span className={`shop-avatar tone-${index % 6}`}><GitBranch size={20} /></span>
                  <span className="shop-main">
                    <strong>{branchDisplayName(branch)}</strong>
                    <small>{branch.code || branch.guidfixed}</small>
                  </span>
                  <span className="shop-badges">
                    {branch.base_currency ? <b>{branch.base_currency}</b> : null}
                    {branch.language ? <em>{branch.language.toUpperCase()}</em> : null}
                  </span>
                </button>
              ))}
            </div>
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

function parseShopInfo(raw: string | null): Record<string, unknown> | null {
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as unknown;
    return parsed && typeof parsed === "object" && !Array.isArray(parsed) ? parsed as Record<string, unknown> : null;
  } catch {
    return null;
  }
}

function createDefaultBranch(): Record<string, unknown> {
  return {
    guidfixed: "",
    code: "00000",
    names: [{ code: "th", name: "สำนักงานใหญ่" }],
    languages: ["th"],
    base_currency: "THB",
    language: "th",
    timezone: "Asia/Bangkok",
    timezoneoffset: "+07:00",
    timezonelabel: "(UTC+07:00) Bangkok",
    yeartype: "buddhist",
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
      languageconfigs: [{ code: "th", codeTranslator: "th", name: "Thai", isuse: true, isdefault: true }],
      latitude: 0,
      longitude: 0,
      taxid: "",
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


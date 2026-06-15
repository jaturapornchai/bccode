"use client";

import {
  AlertCircle,
  ArrowLeft,
  ArrowRight,
  Building2,
  CalendarDays,
  CheckCircle2,
  Coins,
  Copy,
  Crown,
  ExternalLink,
  GitBranch,
  HelpCircle,
  KeyRound,
  Languages,
  Loader2,
  LogOut,
  MessageCircle,
  Plus,
  Search,
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
  localizedName,
  shopDisplayName,
  type AuthSession,
  type BranchListItem,
  type ShopListItem,
  type WorkspaceCompany,
  WORKSPACE_CHANGED_EVENT,
  workspaceStorageKeys,
  notifyWorkspaceChanged,
} from "@/lib/workspace-models";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { LanguageDialog } from "../language-dialog";
import { ManualLink } from "../manual-link";
import { ThemeToggle } from "../theme-toggle";
import { SystemSettingsScreen } from "../system-settings/system-settings-screen";

type Step = "loading" | "shops" | "create" | "branches" | "access";
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
  company: WorkspaceCompany | null;
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
const accessSettingNavItems = [
  {
    route: "/activelanguages",
    label: { th: "ภาษาที่ใช้งาน", en: "Active Languages" },
    helper: { th: "กำหนดก่อนข้อมูลอื่น", en: "Set before other data" },
    banner: "/settings/banner-language.webp",
  },
  {
    route: "/company",
    label: { th: "ข้อมูลบริษัทและสาขา", en: "Company & Branch" },
    helper: { th: "สร้างบริษัทและสำนักงานใหญ่", en: "Create companies and branches" },
    banner: "/settings/banner-company.webp",
  },
  {
    route: "/user",
    label: { th: "ผู้ใช้งาน", en: "Users" },
    helper: { th: "เพิ่มคนเข้าใช้งาน", en: "Add system users" },
    banner: "/settings/banner-users.webp",
  },
  {
    route: "/permissiondefinition",
    label: { th: "กำหนดสิทธิ์หน้าจอ", en: "Permission Definition" },
    helper: { th: "เลือกหน้าจอที่เข้าได้", en: "Choose accessible screens" },
    banner: "/settings/banner-permission-screen.webp",
  },
  {
    route: "/permissiongroup",
    label: { th: "กำหนดสิทธิ์ตามกลุ่ม", en: "Permission Group" },
    helper: { th: "รวมสิทธิ์เป็นชุด", en: "Group permission sets" },
    banner: "/settings/banner-permission-group.webp",
  },
  {
    route: "/permissionlink",
    label: { th: "กำหนดสิทธิ์ผู้ใช้งาน", en: "User Permissions" },
    helper: { th: "ผูกกลุ่มกับผู้ใช้", en: "Assign groups to users" },
    banner: "/settings/banner-permission-user.webp",
  },
  {
    route: "/approvalsetting",
    label: { th: "สิทธิ์การอนุมัติ", en: "Approval Permission" },
    helper: { th: "วงเงินและเอกสารอนุมัติ", en: "Approval limits and documents" },
    banner: "/settings/banner-approval.webp",
  },
  {
    route: "/useraccessaudit",
    label: { th: "ตรวจสอบสถานะผู้ใช้งาน", en: "User Access Audit" },
    helper: { th: "รายงานสิทธิ์และการเข้าถึง", en: "Access and permission report" },
    banner: "/settings/banner-audit.webp",
  },
] as const;

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
  createCompany: "createcompany",
  createCompanyNew: "createcompany_new",
  createCompanyRequiresGoogle: "createcompany_requires_google",
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
  const [selectedCompany, setSelectedCompany] = useState<any | null>(null);
  const [query, setQuery] = useState("");
  const [branchQuery, setBranchQuery] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);
  const [lineDialog, setLineDialog] = useState<LineDialogState>(emptyLineDialog);
  const [pendingUnitSetup, setPendingUnitSetup] = useState<PendingUnitSetup | null>(null);
  const [unitSetupSaving, setUnitSetupSaving] = useState(false);
  const [activeAccessRoute, setActiveAccessRoute] = useState<string | null>(null);
  const [selectedShopForAccess, setSelectedShopForAccess] = useState<ShopListItem | null>(null);
  const isSelectedShopOwner = useMemo(() => {
    if (!selectedShop || !auth) return false;
    const isCreator = selectedShop.is_creator === true
      || Boolean(auth.username && selectedShop.createdby && selectedShop.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase())
      || Boolean(auth.profile?.email && selectedShop.createdby && selectedShop.createdby.trim().toLowerCase() === auth.profile.email.trim().toLowerCase());
    return isCreator || Number(selectedShop.role) === 2;
  }, [selectedShop, auth]);
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
      const selectedHoldingCode = activeHoldingCodeFromAuth(currentAuth);
      const endpoint = selectedHoldingCode
        ? `holdings?activeholdingcode=${encodeURIComponent(selectedHoldingCode)}`
        : "holdings";
      const payload = await callWorkspaceApi<{ data?: ShopListItem[] }>(currentAuth, endpoint);
      const allShops = Array.isArray(payload.data) ? payload.data : [];
      const nextShops = selectedHoldingCode
        ? allShops.filter((shop) => tenantCodeForShop(shop) === selectedHoldingCode)
        : allShops;
      setShops(nextShops);
      setStep("shops");
      if (selectedHoldingCode && nextShops.length === 0) {
        setNotice({
          type: "error",
          text: language === "th"
            ? `ไม่พบ Holding ${selectedHoldingCode} สำหรับบัญชีนี้`
            : `Holding ${selectedHoldingCode} is not available for this account.`,
        });
      } else if (nextShops.length === 0 && !canAuthCreateCompany(currentAuth)) {
        setNotice({ type: "info", textKey: "createCompanyRequiresGoogle" });
      }
    } catch (error) {
      setStep("shops");
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "loadShopsFailed" });
    }
  }, [language]);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? initialLanguage);
    setLanguage(savedLanguage);
    document.documentElement.lang = savedLanguage;

    let savedAuth = readAuth();
    if (!savedAuth) {
      router.replace("/");
      return;
    }

    const selectedHoldingCode = activeHoldingCodeFromAuth(savedAuth);
    if (savedAuth.method === "google" && !selectedHoldingCode) {
      router.replace("/holding");
      return;
    }
    if (selectedHoldingCode && !savedAuth.holdingcode) {
      savedAuth = { ...savedAuth, holdingcode: selectedHoldingCode };
      localStorage.setItem(workspaceStorageKeys.auth, JSON.stringify(savedAuth));
    }

    setAuth(savedAuth);
    void loadShops(savedAuth);
  }, [initialLanguage, loadShops, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem("user_language", language);
  }, [language]);

  useEffect(() => {
    const reloadWorkspaceCompanies = () => {
      if (step !== "shops") return;
      const nextAuth = readAuth();
      if (!nextAuth) return;
      setAuth(nextAuth);
      void loadShops(nextAuth);
    };
    window.addEventListener(WORKSPACE_CHANGED_EVENT, reloadWorkspaceCompanies);
    window.addEventListener("storage", reloadWorkspaceCompanies);
    return () => {
      window.removeEventListener(WORKSPACE_CHANGED_EVENT, reloadWorkspaceCompanies);
      window.removeEventListener("storage", reloadWorkspaceCompanies);
    };
  }, [loadShops, step]);

  useEffect(() => {
    return () => stopLinePolling();
  }, []);

  const flatCompanies = useMemo(() => {
    const list: Array<{
      shop: ShopListItem;
      company: any;
      branches: any[];
    }> = [];

    shops.forEach((shop) => {
      const shopCompanies = Array.isArray((shop as any).companies) ? (shop as any).companies.filter(isVisibleOrganizationRecord) : [];
      const shopBranches = Array.isArray((shop as any).branches) ? (shop as any).branches.filter(isVisibleOrganizationRecord) : [];

      shopCompanies.forEach((company: any) => {
        const compBranches = shopBranches.filter((b: any) => b.companyguid === company.guidfixed);
        list.push({
          shop,
          company,
          branches: compBranches,
        });
      });
    });

    const needle = query.trim().toLowerCase();
    let filtered = list;
    if (needle) {
      filtered = list.filter((item) => {
        const companyName = localizedName(item.company.names, item.company.code || "");
        const companyCode = item.company.code || "";
        const shopName = shopDisplayName(item.shop);

        const matchCompany = `${companyName} ${companyCode}`.toLowerCase().includes(needle);
        const matchShop = `${shopName} ${tenantCodeForShop(item.shop)} ${item.shop.holdingcode}`.toLowerCase().includes(needle);
        const matchBranch = item.branches.some((b) =>
          `${branchDisplayName(b)} ${b.code || ""}`.toLowerCase().includes(needle)
        );

        return matchCompany || matchShop || matchBranch;
      });
    }

    return filtered.map((item) => {
      const sortedBranches = [...item.branches].sort((a, b) => {
        const codeA = a.code || "";
        const codeB = b.code || "";
        return codeA.localeCompare(codeB, undefined, { numeric: true, sensitivity: "base" });
      });
      return {
        ...item,
        branches: sortedBranches,
      };
    }).sort((a, b) => {
      // Sort by shop name/id first
      const shopA = shopDisplayName(a.shop) || tenantCodeForShop(a.shop) || "";
      const shopB = shopDisplayName(b.shop) || tenantCodeForShop(b.shop) || "";
      const shopCmp = shopA.localeCompare(shopB, "th", { sensitivity: "base" });
      if (shopCmp !== 0) return shopCmp;

      // Under the same shop, sort by company code
      const codeA = a.company.code || "";
      const codeB = b.company.code || "";
      const cmp = codeA.localeCompare(codeB, undefined, { numeric: true, sensitivity: "base" });
      if (cmp !== 0) return cmp;

      const nameA = localizedName(a.company.names, a.company.code || "");
      const nameB = localizedName(b.company.names, b.company.code || "");
      return nameA.localeCompare(nameB, "th", { sensitivity: "base" });
    });
  }, [query, shops, language]);

  const filteredBranches = useMemo(() => {
    const needle = branchQuery.trim().toLowerCase();
    let list = branches;
    if (selectedCompany) {
      list = branches.filter((b) => b.companyguid === selectedCompany.guidfixed);
    }
    if (!needle) return list;
    return list.filter((branch) => `${branchDisplayName(branch)} ${branch.code ?? ""}`.toLowerCase().includes(needle));
  }, [branchQuery, branches, selectedCompany]);

  const accessShopOptions = useMemo(() => {
    const seen = new Set<string>();
    const options: Array<{ shop: ShopListItem; label: string }> = [];

    shops.forEach((shop) => {
      const holdingCode = tenantCodeForShop(shop);
      if (!holdingCode || seen.has(holdingCode)) return;
      seen.add(holdingCode);
      options.push({ shop, label: holdingAccessDisplayName(shop, language) });
    });

    return options.sort((a, b) => a.label.localeCompare(b.label, language === "th" ? "th" : "en", { sensitivity: "base" }));
  }, [language, shops]);

  const selectedAccessShopLabel = useMemo(() => {
    if (!selectedShopForAccess) return language === "th" ? "เลือก Holding" : "Select Holding";
    return accessShopOptions.find((option) => tenantCodeForShop(option.shop) === tenantCodeForShop(selectedShopForAccess))?.label
      ?? holdingAccessDisplayName(selectedShopForAccess, language);
  }, [accessShopOptions, language, selectedShopForAccess]);
  const activeHoldingContext = useMemo(() => {
    const activeHoldingCode = auth ? activeHoldingCodeFromAuth(auth) : "";
    const activeShop =
      selectedShopForAccess ??
      selectedShop ??
      (activeHoldingCode
        ? shops.find((shop) => tenantCodeForShop(shop) === activeHoldingCode)
        : null) ??
      (shops.length === 1 ? shops[0] : null);
    const code = tenantCodeForShop(activeShop) || activeHoldingCode;
    if (!code) return null;
    const name = activeShop ? shopDisplayName(activeShop) : code;
    return { code, name: name || code };
  }, [auth, selectedShop, selectedShopForAccess, shops]);

  const signedInAs = auth?.profile?.email || auth?.username || "";
  const currentTitle = step === "branches" ? text("selectBranchTitle") : text("selectCompanyTitle");
  const currentSummary =
    step === "branches"
      ? `${text("availableBranches")}: ${filteredBranches.length}`
      : `${text("availableCompanies")}: ${flatCompanies.length}`;

  const returnToShopSelection = useCallback(async () => {
    setActiveAccessRoute(null);
    setSelectedShopForAccess(null);
    setSelectedShop(null);
    setSelectedCompany(null);
    setBranches([]);
    setBranchQuery("");

    if (!auth) {
      setStep("shops");
      return;
    }

    await loadShops(auth);
  }, [auth, loadShops]);

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

  async function selectShopAndBranch(shop: ShopListItem, branch: BranchListItem) {
    if (!auth) return;
    setBusy(true);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "select-holding", {
        method: "POST",
        body: { holdingcode: tenantCodeForShop(shop) },
      });
      const shopInfo = await callWorkspaceApi<{ data?: Record<string, unknown> }>(auth, `holding-info?holdingcode=${encodeURIComponent(tenantCodeForShop(shop))}`);
      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));
      setSelectedShop(shop);

      const branchPayload = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
      const nextBranches = Array.isArray(branchPayload.data) ? branchPayload.data : [branch];
      setBranches(nextBranches);

      await enterWorkspaceWithUnitCheck(shop, branch, shopInfo.data ?? null);
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "selectShopFailed" });
    } finally {
      setBusy(false);
    }
  }

  async function selectShop(shop: ShopListItem) {
    if (!auth) return;
    setBusy(true);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "select-holding", {
        method: "POST",
        body: { holdingcode: tenantCodeForShop(shop) },
      });
      const shopInfo = await callWorkspaceApi<{ data?: Record<string, unknown> }>(auth, `holding-info?holdingcode=${encodeURIComponent(tenantCodeForShop(shop))}`);
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
      const defaultBranch = nextBranches[0] ?? null;
      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));
      localStorage.setItem(workspaceStorageKeys.workspace, JSON.stringify({ shop, branch: defaultBranch, shopInfo: shopInfo.data ?? null }));

      setStep("branches");
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "selectShopFailed" });
    } finally {
      setBusy(false);
    }
  }

  async function selectCompany(shop: ShopListItem, company: any) {
    if (!auth) return;
    setBusy(true);
    setNotice(null);
    try {
      await callWorkspaceApi(auth, "select-holding", {
        method: "POST",
        body: { holdingcode: tenantCodeForShop(shop) },
      });
      const shopInfo = await callWorkspaceApi<{ data?: Record<string, unknown> }>(auth, `holding-info?holdingcode=${encodeURIComponent(tenantCodeForShop(shop))}`);
      const branchPayload = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
      let nextBranches = Array.isArray(branchPayload.data) ? branchPayload.data : [];

      // Filter branches of the selected company
      let compBranches = nextBranches.filter((b) => b.companyguid === company.guidfixed);

      if (compBranches.length === 0) {
        // Create default branch (สำนักงานใหญ่) for this company
        const created = await callWorkspaceApi<{ data?: BranchListItem; id?: string; ID?: string }>(auth, "branch", {
          method: "POST",
          body: { branch: createDefaultBranch(shop, shopInfo.data ?? null, company.guidfixed) },
        });
        if (created.data) {
          nextBranches = [...nextBranches, created.data];
        } else {
          const reloaded = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
          nextBranches = Array.isArray(reloaded.data) ? reloaded.data : [];
        }
      }

      setSelectedShop(shop);
      setSelectedCompany(company);
      setBranches(nextBranches);
      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));
      setStep("branches");
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", textKey: "selectShopFailed" });
    } finally {
      setBusy(false);
    }
  }

  async function openAccessSettings(route: string) {
    if (!auth) return;
    const representativeShop = accessShopOptions[0]?.shop;
    if (!representativeShop) {
      setNotice({ type: "error", text: language === "th" ? "ไม่พบ Holding สำหรับกำหนดสิทธิ์" : "No Holding found for access control." });
      return;
    }
    setBusy(true);
    setNotice(null);
    try {
      setSelectedShopForAccess(representativeShop);
      await callWorkspaceApi(auth, "select-holding", {
        method: "POST",
        body: { holdingcode: tenantCodeForShop(representativeShop) },
      });
      const shopInfo = await callWorkspaceApi<{ data?: Record<string, unknown> }>(auth, `holding-info?holdingcode=${encodeURIComponent(tenantCodeForShop(representativeShop))}`);
      const branchPayload = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
      const nextBranches = Array.isArray(branchPayload.data) ? branchPayload.data : [];
      const defaultBranch = nextBranches[0] ?? null;

      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));
      localStorage.setItem(workspaceStorageKeys.workspace, JSON.stringify({ shop: representativeShop, branch: defaultBranch, shopInfo: shopInfo.data ?? null }));

      setActiveAccessRoute(route);
      setStep("access");
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", text: language === "th" ? "เตรียมระบบกำหนดสิทธิ์ล้มเหลว" : "Failed to initialize access control." });
    } finally {
      setBusy(false);
    }
  }

  async function handleAccessShopChange(shop: ShopListItem) {
    if (!auth) return;
    setBusy(true);
    setNotice(null);
    try {
      setSelectedShopForAccess(shop);
      await callWorkspaceApi(auth, "select-holding", {
        method: "POST",
        body: { holdingcode: tenantCodeForShop(shop) },
      });
      const shopInfo = await callWorkspaceApi<{ data?: Record<string, unknown> }>(auth, `holding-info?holdingcode=${encodeURIComponent(tenantCodeForShop(shop))}`);
      const branchPayload = await callWorkspaceApi<{ data?: BranchListItem[] }>(auth, "branches?offset=0&limit=100&q=");
      const nextBranches = Array.isArray(branchPayload.data) ? branchPayload.data : [];
      const defaultBranch = nextBranches[0] ?? null;

      localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(shopInfo.data ?? null));
      localStorage.setItem(workspaceStorageKeys.workspace, JSON.stringify({ shop, branch: defaultBranch, shopInfo: shopInfo.data ?? null }));

      notifyWorkspaceChanged();
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", text: language === "th" ? "เปลี่ยนบริษัทสำหรับกำหนดสิทธิ์ล้มเหลว" : "Failed to switch company." });
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
      await callWorkspaceApi(auth, "create-holding", {
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
      await enterWorkspaceWithUnitCheck(selectedShop, branch, shopInfo, selectedCompany);
    } catch (error) {
      setNotice(error instanceof Error && error.message ? { type: "error", text: error.message } : { type: "error", text: text("requestFailed") });
    } finally {
      setBusy(false);
    }
  }

  async function enterWorkspaceWithUnitCheck(
    shop: ShopListItem,
    branch: BranchListItem | null,
    shopInfo: Record<string, unknown> | null,
    company: WorkspaceCompany | null = null,
  ) {
    if (!auth) return;
    persistWorkspace(shop, branch, shopInfo, company);
    const payload = await callWorkspaceApi<{ data?: unknown[]; total?: number }>(auth, "product-units?offset=0&limit=1&q=");

    if (getApiTotal(payload) > 0) {
      router.push("/menu");
      return;
    }

    const mainHoldingCode = getMainHoldingCode(shopInfo);
    const standardUnits = await callWorkspaceApi<StandardProductUnitResponse>(
      auth,
      `product-units/standard?mainHoldingCode=${encodeURIComponent(mainHoldingCode)}&q=`,
    );
    const units = Array.isArray(standardUnits.data) ? standardUnits.data : [];
    setPendingUnitSetup({
      shop,
      company,
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
        body: { mainHoldingCode: getMainHoldingCode(pendingUnitSetup.shopInfo), unitcodes: pendingUnitSetup.selectedCodes },
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

  if (step === "access" && activeAccessRoute) {
    return (
      <main className="w-screen h-screen bg-background flex flex-col overflow-hidden">
        <section className="w-full h-full flex flex-col bg-card" role="dialog" aria-modal="true">
          <div className="dialog-header shrink-0 flex items-center justify-between gap-3 px-6">
            <div className="flex items-center gap-6 min-w-0">
              <button
                className="secondary-button flex items-center gap-2 px-3 py-1.5 text-sm font-bold text-foreground hover:bg-muted hover:text-primary border border-border rounded-xl transition-all shadow-sm shrink-0"
                type="button"
                onClick={() => void returnToShopSelection()}
              >
                <ArrowLeft size={16} className="text-muted-foreground" />
                <span>{language === "th" ? "ย้อนกลับ" : "Back"}</span>
              </button>
              <div>
                <p className="eyebrow">{language === "th" ? "การตั้งค่าระบบ" : "SYSTEM CONFIGURATION"}</p>
                <h2 className="text-xl font-bold">
                  {language === "th" ? "ตั้งค่าระบบและการเข้าถึง" : "Settings & Access Control"}
                </h2>
              </div>

            </div>
            <div className="flex min-w-0 items-center gap-2">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    className="secondary-button flex min-w-0 max-w-[42vw] items-center gap-2 px-3 py-1.5 text-sm font-bold text-foreground hover:bg-muted hover:text-primary border border-border rounded-xl transition-all shadow-sm"
                    type="button"
                    disabled={busy || accessShopOptions.length <= 1}
                  >
                    <Crown size={16} className="shrink-0 text-muted-foreground" />
                    <span className="min-w-0 truncate">
                      {selectedAccessShopLabel}
                    </span>
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-80">
                  {accessShopOptions.map(({ shop, label }) => {
                    const active = tenantCodeForShop(shop) === tenantCodeForShop(selectedShopForAccess);
                    return (
                      <DropdownMenuItem
                        className="gap-2"
                        disabled={busy || active}
                        key={tenantCodeForShop(shop)}
                        onClick={() => void handleAccessShopChange(shop)}
                      >
                        <Crown className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <span className="min-w-0 flex-1 truncate">{label}</span>
                        {active ? <CheckCircle2 className="h-4 w-4 shrink-0 text-primary" /> : null}
                      </DropdownMenuItem>
                    );
                  })}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
          <div className="flex-1 min-h-0 flex bg-card overflow-hidden">
            {/* Sidebar ภายใน Modal */}
            <aside className="w-60 shrink-0 border-r border-border bg-muted/20 p-4 flex flex-col gap-1 overflow-y-auto">
              <p className="px-2 mb-2 text-xs font-bold text-muted-foreground uppercase tracking-wider">
                {language === "th" ? "ตั้งค่าระบบและการเข้าถึง" : "Settings & Access"}
              </p>
              {accessSettingNavItems.map((item, index) => {
                const isActive = activeAccessRoute === item.route;
                return (
                  <button
                    key={item.route}
                    className={`w-full text-left px-2.5 py-2 rounded-lg text-sm font-semibold transition-all duration-200 flex items-start gap-2 ${
                      isActive
                        ? "bg-primary text-primary-foreground shadow-md ring-1 ring-primary/40"
                        : "text-foreground hover:bg-muted hover:translate-x-0.5"
                    }`}
                    type="button"
                    onClick={() => setActiveAccessRoute(item.route)}
                  >
                    <span
                      className={`mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full text-[10px] font-bold ${
                        isActive
                          ? "bg-primary-foreground/20 text-primary-foreground"
                          : "bg-primary/10 text-primary"
                      }`}
                    >
                      {index + 1}
                    </span>
                    <span className="min-w-0">
                      <span className="block truncate">
                        {language === "th" ? item.label.th : item.label.en}
                      </span>
                      <span
                        className={`block truncate text-[10px] font-medium ${
                          isActive
                            ? "text-primary-foreground/80"
                            : "text-muted-foreground"
                        }`}
                      >
                        {language === "th" ? item.helper.th : item.helper.en}
                      </span>
                    </span>
                  </button>
                );
              })}
            </aside>

            {/* คอนเทนต์แสดงผลฝั่งขวา */}
            <div className="flex-1 min-h-0 overflow-y-auto p-4">
              {(() => {
                const bIdx = accessSettingNavItems.findIndex((i) => i.route === activeAccessRoute);
                const bItem = bIdx >= 0 ? accessSettingNavItems[bIdx] : null;
                if (!bItem?.banner) return null;
                return (
                  <div
                    className="relative mb-4 h-36 w-full overflow-hidden rounded-2xl border border-border/60 bg-muted shadow-sm sm:h-44"
                    style={{ backgroundImage: `url(${bItem.banner})`, backgroundSize: "cover", backgroundPosition: "center", backgroundRepeat: "no-repeat" }}
                  >
                    <div className="absolute inset-0 bg-gradient-to-r from-[#812920]/92 via-[#812920]/55 to-[#812920]/10" />
                    <div className="relative z-10 flex h-full items-center gap-3 px-5">
                      <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-white/25 text-lg font-black text-white ring-1 ring-white/50 backdrop-blur-sm [text-shadow:0_1px_2px_rgba(0,0,0,0.6)]">
                        {bIdx + 1}
                      </span>
                      <div className="min-w-0">
                        <h3 className="text-xl font-black leading-tight text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.75),0_0_6px_rgba(0,0,0,0.55)] sm:text-2xl">
                          {language === "th" ? bItem.label.th : bItem.label.en}
                        </h3>
                        <p className="text-xs font-semibold text-white/95 [text-shadow:0_1px_2px_rgba(0,0,0,0.75),0_0_5px_rgba(0,0,0,0.5)] sm:text-sm">
                          {language === "th" ? bItem.helper.th : bItem.helper.en}
                        </p>
                      </div>
                    </div>
                  </div>
                );
              })()}
              {activeAccessRoute !== "/activelanguages" &&
              !hasExplicitLanguageSettings(selectedShopForAccess) ? (
                <div className="mb-3 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-primary/25 bg-primary/10 px-3 py-2 text-sm text-foreground">
                  <span className="min-w-0 text-xs font-semibold text-muted-foreground">
                    {language === "th"
                      ? "ยังไม่ได้ยืนยันภาษาที่ใช้งาน ระบบใช้ภาษาไทยเป็นค่าเริ่มต้น ควรกำหนดภาษาก่อนกรอกชื่อหลายภาษา"
                      : "Active languages are not confirmed yet. The system uses Thai by default; set languages before entering multilingual names."}
                  </span>
                  <button
                    className="secondary-button inline-flex items-center gap-1.5 px-2.5 py-1 text-xs"
                    type="button"
                    onClick={() => setActiveAccessRoute("/activelanguages")}
                  >
                    <Languages size={14} />
                    <span>{language === "th" ? "ไปตั้งภาษา" : "Set languages"}</span>
                  </button>
                </div>
              ) : null}
              <SystemSettingsScreen
                key={`${activeAccessRoute}:${tenantCodeForShop(selectedShopForAccess)}`}
                route={activeAccessRoute}
                embedded
                hideChrome
                branchOverride={null}
                language={language}
                initialLanguage={language}
              />
            </div>
          </div>
        </section>
      </main>
    );
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
            {activeHoldingContext ? (
              <div
                className="workspace-holding-context"
                title={`${activeHoldingContext.name} (${activeHoldingContext.code})`}
              >
                <span>{language === "th" ? "Holding ที่ใช้งาน" : "Active Holding"}</span>
                <strong>{activeHoldingContext.name}</strong>
                {activeHoldingContext.name !== activeHoldingContext.code ? (
                  <code>holdingcode: {activeHoldingContext.code}</code>
                ) : null}
              </div>
            ) : null}
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
          <div className="flex items-center gap-2">
            {step === "shops" ? (
              <button
                className="secondary-button workspace-head-action flex items-center gap-1.5"
                type="button"
                onClick={() => void openAccessSettings("/activelanguages")}
                disabled={busy}
              >
                <KeyRound size={17} />
                <span>{language === "th" ? "ตั้งค่าระบบ" : "Settings"}</span>
              </button>
            ) : null}
          </div>
        </div>

        <div className="step-strip">
          <span className="active">{text("stepLogin")}</span>
          <span className={step === "shops" ? "active" : ""}>
            {language === "th" ? "2 เลือกบริษัทและสาขา" : "2 Select Company & Branch"}
          </span>
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
            {flatCompanies.length === 0 ? (
              <div className="workspace-empty-state flex flex-col items-center justify-center p-6 text-center max-w-lg mx-auto">
                <div className="w-16 h-16 rounded-2xl bg-primary/10 text-primary flex items-center justify-center mb-4">
                  <Building2 size={32} />
                </div>
                <strong className="text-base text-foreground font-bold mb-2">
                  {language === "th"
                    ? "ยังไม่มีบริษัทเปิดใช้งานในระบบของคุณ"
                    : "No active companies found in your system"}
                </strong>
                <p className="text-xs text-muted-foreground mb-6">
                  {language === "th"
                    ? "เริ่มจากกำหนดภาษาที่ใช้งานก่อน แล้วค่อยสร้างบริษัทและสาขา"
                    : "Start with active languages, then create the company and branches."}
                </p>

                <button
                  className="secondary-button mb-4 inline-flex items-center gap-1.5"
                  type="button"
                  onClick={() => void openAccessSettings("/activelanguages")}
                  disabled={busy || accessShopOptions.length === 0}
                >
                  <Languages size={16} />
                  <span>{language === "th" ? "ตั้งค่าภาษาก่อน" : "Set languages first"}</span>
                </button>

                <div className="w-full text-left bg-accent/35 border border-border/60 rounded-xl p-4 space-y-3.5">
                  <h4 className="text-xs font-bold text-foreground flex items-center gap-1.5 border-b pb-2">
                    <HelpCircle size={14} className="text-primary" />
                    <span>
                      {language === "th" ? "ขั้นตอนเริ่มต้นใช้งาน" : "Getting Started"}
                    </span>
                  </h4>

                  <ul className="space-y-3 text-[11px] text-muted-foreground">
                    <li className="flex items-start gap-2.5">
                      <span className="w-4 h-4 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">1</span>
                      <div>
                        <strong className="text-foreground block">
                          {language === "th" ? "เข้าเมนูตั้งค่าระบบ" : "Go to Settings"}
                        </strong>
                        <span>
                          {language === "th"
                            ? "คลิกปุ่ม 'ตั้งค่าระบบ' 🔑 สีน้ำเงินที่มุมขวาบนของหน้านี้"
                            : "Click the 'Settings' 🔑 button at the top-right of this panel."}
                        </span>
                      </div>
                    </li>

                    <li className="flex items-start gap-2.5">
                      <span className="w-4 h-4 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">2</span>
                      <div>
                        <strong className="text-foreground block">
                          {language === "th" ? "เลือกหัวข้อภาษาที่ใช้งาน" : "Select Active Languages"}
                        </strong>
                        <span>
                          {language === "th"
                            ? "ตรวจภาษาไทยที่เป็นค่าเริ่มต้น เพิ่มภาษาอื่นที่ต้องใช้ และลากภาษาแรกไว้บนสุด"
                            : "Review Thai as the default, add any other active languages, and keep the primary language first."}
                        </span>
                      </div>
                    </li>

                    <li className="flex items-start gap-2.5">
                      <span className="w-4 h-4 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">3</span>
                      <div>
                        <strong className="text-foreground block">
                          {language === "th" ? "เลือกข้อมูลบริษัทและสาขา" : "Select Company & Branch Info"}
                        </strong>
                        <span>
                          {language === "th"
                            ? "หลังตั้งภาษาแล้ว ให้เลือกหัวข้อ 'ข้อมูลบริษัทและสาขา'"
                            : "After setting languages, select 'Company & Branch Info'."}
                        </span>
                      </div>
                    </li>

                    <li className="flex items-start gap-2.5">
                      <span className="w-4 h-4 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">4</span>
                      <div>
                        <strong className="text-foreground block">
                          {language === "th" ? "เพิ่มบริษัทใหม่และบันทึก" : "Add Company & Save"}
                        </strong>
                        <span>
                          {language === "th"
                            ? "คลิก '+ เพิ่มบริษัท' กรอกรหัสและชื่อบริษัทตามภาษาที่ตั้งไว้ แล้วบันทึก"
                            : "Click '+ Add Company', fill in the code and localized company names, then save."}
                        </span>
                      </div>
                    </li>

                    <li className="flex items-start gap-2.5">
                      <span className="w-4 h-4 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-[9px] shrink-0 mt-0.5">5</span>
                      <div>
                        <strong className="text-foreground block">
                          {language === "th" ? "สร้างผู้ใช้และสิทธิ์" : "Create Users & Permissions"}
                        </strong>
                        <span>
                          {language === "th"
                            ? "เพิ่มผู้ใช้งาน กำหนดสิทธิ์หน้าจอ รวมเป็นกลุ่ม แล้วผูกกลุ่มให้ผู้ใช้"
                            : "Add users, define screen permissions, group them, then assign groups to users."}
                        </span>
                      </div>
                    </li>
                  </ul>
                </div>
              </div>
            ) : (
              <div className="flex flex-wrap gap-6 justify-center w-full py-2">
                {flatCompanies.map((item, index) => {
                  const { shop, company } = item;
                  const isCreator = shop.is_creator === true
                    || Boolean(auth?.username && shop.createdby && shop.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase());
                  const languageCodes = shopLanguageCodes(shop);
                  const currencyLabel = shopCurrencyLabel(shop, language);
                  const compName = localizedName(company.names, company.code || "");
                  const holdingLabel = shop.holdingcode?.trim()
                    ? `holdingcode: ${shop.holdingcode.trim()}`
                    : `legacy_holdingcode: ${shop.holdingcode}`;

                  return (
                    <button
                      key={`${shop.holdingcode}-${company.guidfixed || company.code}`}
                      disabled={busy}
                      onClick={() => void selectCompany(shop, company)}
                      className="group/company text-left relative w-full md:w-[calc(50%-12px)] lg:w-[350px] shrink-0 border border-border/80 rounded-2xl bg-card/85 backdrop-blur-md overflow-hidden shadow-md hover:shadow-xl hover:border-primary/40 hover:scale-[1.01] transition-all duration-300"
                    >
                      {/* Left color bar accent */}
                      <div className={`absolute left-0 top-0 bottom-0 w-1.5 bg-gradient-to-b ${index % 2 === 0 ? "from-indigo-500 to-indigo-600" : "from-teal-500 to-emerald-500"}`} />

                      {/* Company Header */}
                      <div className="flex flex-col p-4 pl-6">
                        <div className="flex items-center gap-3">
                          <span className={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${index % 2 === 0 ? "bg-indigo-500/10 text-indigo-500" : "bg-teal-500/10 text-teal-500"}`}>
                            <Building2 size={20} />
                          </span>
                          <div className="min-w-0">
                            <h3 className="font-bold text-foreground text-sm sm:text-base tracking-tight truncate" title={compName}>
                              [{company.code}] {compName}
                            </h3>
                            <p className="text-[10px] text-muted-foreground/80 mt-0.5 font-mono truncate">{holdingLabel}</p>
                          </div>
                        </div>

                        <div className="flex flex-wrap items-center gap-1.5 mt-3 mb-4">
                          <span className={`px-2 py-0.5 text-[9px] font-bold rounded border uppercase shrink-0 ${
                            isCreator
                              ? "bg-amber-500/10 text-amber-600 dark:text-amber-500 border-amber-500/20"
                              : "bg-blue-500/10 text-blue-600 dark:text-blue-500 border-blue-500/20"
                          }`}>
                            {isCreator ? "OWNER" : "USER"}
                          </span>
                          {languageCodes.map((code) => (
                            <span key={code} className="px-1.5 py-0.5 text-[9px] bg-muted border border-border/50 text-muted-foreground rounded font-semibold uppercase shrink-0">
                              {code}
                            </span>
                          ))}
                          {currencyLabel && (
                            <span className="px-1.5 py-0.5 text-[9px] bg-muted border border-border/50 text-muted-foreground rounded font-semibold uppercase shrink-0">
                              {currencyLabel.split(" ")[0]}
                            </span>
                          )}
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </>
        ) : null}


        {step === "branches" ? (
          <>
            <div className="workspace-toolbar">
              <button className="text-action compact-action" type="button" onClick={() => void returnToShopSelection()}>
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
                <button className="shop-card branch-card" disabled={busy} key={branch.guidfixed || branch.code} type="button" onClick={() => void selectBranch(branch)}>
                  <span className={`shop-avatar tone-${index % 6}`}><GitBranch size={20} /></span>
                  <span className="shop-main">
                    <strong>{branchDisplayName(branch)}</strong>
                    <small>{branch.code || branch.guidfixed}</small>
                  </span>
                  <span className="shop-badges">
                    {branch.basecurrency ? <b>{branch.basecurrency}</b> : null}
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

function activeHoldingCodeFromAuth(auth: AuthSession | null): string {
  const authHoldingCode = auth?.holdingcode?.trim();
  if (authHoldingCode) return authHoldingCode;
  if (typeof window === "undefined") return "";
  return localStorage.getItem(workspaceStorageKeys.holdingCode)?.trim() || "";
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

function persistWorkspace(
  shop: ShopListItem,
  branch: BranchListItem | null,
  shopInfo: Record<string, unknown> | null,
  company: WorkspaceCompany | null = null,
) {
  localStorage.setItem(workspaceStorageKeys.workspace, JSON.stringify({ shop, company, branch, shopInfo }));
  localStorage.setItem(workspaceStorageKeys.branch, JSON.stringify(branch));
}

function tenantCodeForShop(shop: ShopListItem | null | undefined): string {
  return shop?.holdingcode?.trim() || shop?.holdingcode?.trim() || "";
}

function hasExplicitLanguageSettings(shop: ShopListItem | null | undefined): boolean {
  return Array.isArray(shop?.languageconfigs) && shop.languageconfigs.length > 0;
}

function getApiTotal(payload: { data?: unknown[]; total?: number }): number {
  if (typeof payload.total === "number") return payload.total;
  return Array.isArray(payload.data) ? payload.data.length : 0;
}

function getMainHoldingCode(shopInfo: Record<string, unknown> | null): string {
  if (!shopInfo) return "";
  const mainHoldingCode = shopInfo.main_holdingcode ?? shopInfo.mainholdingcode ?? shopInfo.mainHoldingCode;
  return typeof mainHoldingCode === "string" ? mainHoldingCode.trim() : "";
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
  return normalizedCodeList(shop.activelanguages, ["th"]).map((code) => code.toUpperCase());
}

function shopCurrencyCodes(shop: ShopListItem): string[] {
  return normalizedCodeList(shop.currencies, shop.basecurrency ? [shop.basecurrency] : ["THB"]).map((code) => code.toUpperCase());
}

function shopCurrencyLabel(shop: ShopListItem, language: LanguageCode): string {
  const currencies = shopCurrencyCodes(shop);
  const hasConfiguredCurrency = (Array.isArray(shop.currencies) && shop.currencies.length > 0) || Boolean(stringValue(shop.basecurrency));
  const defaultMarker = hasConfiguredCurrency ? "" : language === "th" ? " ค่าเริ่มต้น" : " default";
  return `${currencies.join(", ")}${defaultMarker}`;
}

function shopDateFormatLabel(shop: ShopListItem, language: LanguageCode): string {
  const hasConfiguredDateFormat = Boolean(stringValue(shop.dateformat));
  const format = stringValue(shop.dateformat) || "dd/MM/yyyy";
  const yearType = hasConfiguredDateFormat ? normalizedYearType(shop) : "buddhist";
  const yearLabel = language === "th"
    ? yearType === "buddhist" ? "พ.ศ." : "ค.ศ."
    : yearType === "buddhist" ? "BE" : "CE";
  const defaultMarker = hasConfiguredDateFormat ? "" : language === "th" ? " ค่าเริ่มต้น" : " default";
  return `${format} ${yearLabel}${defaultMarker}`;
}

function normalizedYearType(shop: ShopListItem): "buddhist" | "christian" {
  const yearType = stringValue(shop.yeartype).toLowerCase();
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

function createDefaultBranch(shop?: ShopListItem, shopInfo?: Record<string, unknown> | null, companyGuid?: string): Record<string, unknown> {
  const settings = recordValue(shopInfo?.settings);
  const languageCodes = activeLanguageCodes(settings);
  const companyNames = normalizedNames(shopInfo?.names, shop ? shopDisplayName(shop) : "");
  const companyAddress = normalizedNames(shopInfo?.address, "");
  return {
    guidfixed: "",
    code: "00000",
    companyguid: companyGuid || "",
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
      taxid: stringValue(settings?.taxid),
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
    companyregistrationno: stringValue(settings?.companyregistrationno),
    isvatregistered: booleanValue(settings?.isvatregistered),
    basecurrency: stringValue(settings?.basecurrency) || "THB",
    language: languageCodes[0] ?? "th",
    timezone: stringValue(settings?.timezone) || "Asia/Bangkok",
    timezoneoffset: stringValue(settings?.timezoneoffset) || "+07:00",
    timezonelabel: stringValue(settings?.timezonelabel) || "(UTC+07:00) Bangkok",
    dateformat: stringValue(settings?.dateformat) || "dd/MM/yyyy",
    yeartype: booleanValue(settings?.usebuddhistcalendar, true) ? "buddhist" : "christian",
    decimalquantity: numberValue(settings?.decimalquantity, 2),
    decimalprice: numberValue(settings?.decimalprice, 2),
    decimaldocument: numberValue(settings?.decimaldocument, 2),
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
    .filter((item) => !recordValue(item) || recordValue(item)?.isuse !== false)
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

function holdingAccessDisplayName(shop: ShopListItem, language: LanguageCode): string {
  const holdingCode = tenantCodeForShop(shop);
  const holdingName = shopDisplayName(shop) || holdingCode;
  const prefix = language === "th" ? "Holding" : "Holding";
  if (!holdingCode || holdingName === holdingCode) return `${prefix}: ${holdingName}`;
  return `${prefix}: ${holdingName} (${holdingCode})`;
}

function isVisibleOrganizationRecord(value: unknown): boolean {
  const record = recordValue(value);
  if (!record) return false;
  if (record.isactive === false) return false;
  const deletedAt = stringValue(record.deletedat);
  return deletedAt.length === 0;
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
    guidfixed: typeof guidFixed === "string" && guidFixed.trim() ? guidFixed.trim() : "00000",
    code: "00000",
    names: [{ code: "th", name: "สำนักงานใหญ่" }],
    basecurrency: "THB",
    language: "th",
    timezone: "Asia/Bangkok",
    timezoneoffset: "+07:00",
    timezonelabel: "(UTC+07:00) Bangkok",
    yeartype: "buddhist",
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
      languageconfigs: [{ code: "th", codetranslator: "th", name: "ภาษาไทย", isuse: true, isdefault: true }],
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


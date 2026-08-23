"use client";

import { useQuery } from "@tanstack/react-query";
import {
  AlertCircle,
  AlertTriangle,
  Bell,
  Building2,
  CheckCircle2,
  ChevronDown,
  ChevronsUpDown,
  CircleX,
  Command,
  Copy,
  Crown,
  ExternalLink,
  GitBranch,
  KeyRound,
  LayoutDashboard,
  LayoutPanelTop,
  Lock,
  Loader2,
  LogOut,
  MessageCircle,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
  RefreshCcw,
  Search,
  Settings,
  Star,
  Store,
  UserRound,
} from "lucide-react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import QRCode from "qrcode";
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type DragEvent, type FormEvent, type ReactNode, type UIEvent } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LogoAvatar } from "@/components/logo-avatar";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
import {
  MENU_SECTIONS,
  flattenMenuItems,
  menuText,
  type MenuGroup,
  type MenuItem,
  type MenuSection,
} from "@/lib/menu-data";
import { getFrequentMenuEntries, menuUsageStorageKey, readMenuUsage, recordMenuUsage, type MenuUsageMap } from "@/lib/menu-usage";
import { getSystemSettingConfig } from "@/lib/system-setting-screens";
import { authFetch, clearAuthSession, getAuthSession, logoutAuthSession } from "@/lib/client-auth-session";
import { pushNotice } from "@/lib/toast";
import { cn } from "@/lib/utils";
import {
  holdingDisplayName,
  shopDisplayName,
  WORKSPACE_CHANGED_EVENT,
  workspaceBranchDisplayName,
  workspaceCompanyDisplayName,
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { CurrencyScreen } from "../currency/currency-screen";
import { AppHeaderControls } from "../app-header-controls";
import { LineOaLinkScreen } from "../line-oa/line-oa-link-screen";
import { ManualLink } from "../manual-link";
import { SystemSettingsScreen } from "../system-settings/system-settings-screen";
import { ZoomControl } from "../zoom-control";
import { HomeMenuIcon, MenuRouteIcon } from "./menu-icon";
import { MenuDataTable } from "./menu-data-table";
import { buildChartData, buildKpis, fetchErpMenuRows, type ErpMenuRow } from "./menu-dashboard-data";
import { MenuKpiChart } from "./menu-kpi-chart";
import { MenuQueryProvider } from "./menu-query-provider";
import { ProductBarcodeScreen } from "./product-barcode-screen";
import { ProductScreen } from "./product-screen";
import { ProductSetScreen } from "./product-set-screen";
import { ProductBarcodeShelfScreen } from "./product-barcode-shelf-screen";
import { ProductPriceHistoryScreen } from "./product-price-history-screen";
import { MarketplaceMappingsScreen } from "./marketplace-screen";
import { DataModelGraphScreen } from "./datamodel-graph-screen";

type WorkTab = {
  id: string;
  title: string;
  route: string;
  item?: MenuItem;
  closable: boolean;
  productFocusRequest?: { code: string; requestId: string };
};

type TabInsertSide = "before" | "after";
type MenuLayoutMode = "left" | "top";
type LineDialogState = {
  open: boolean;
  loading: boolean;
  success: boolean;
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
type ProfileResponse = {
  success?: boolean;
  data?: {
    email?: string;
    name?: string;
    username?: string;
    avatar?: string;
    avatarthumb?: string;
  };
  message?: string;
};
type MenuTreeNode =
  | { id: string; item: MenuItem; type: "item" }
  | { children: MenuItem[]; id: string; label: string; seedItem?: MenuItem; type: "folder" };
type SettingRecord = Record<string, unknown>;

const firstTab: WorkTab = { id: "home", title: "ภาพรวม", route: "/menu", closable: false };
const menuLayoutStorageKey = "bc_menu_layout_mode";
const SOCIAL_POLL_TIMEOUT_MS = 5 * 60 * 1000;
const DEFAULT_SOCIAL_POLL_INTERVAL_MS = 2000;
const emptyLineDialog: LineDialogState = {
  open: false,
  loading: false,
  success: false,
  code: "",
  loginUrl: "",
  qrDataUrl: "",
  expiresAt: "",
  error: "",
  expired: false,
};
const companyMenuIds = new Set(["company", "branch", "employee", "company-type", "currency"]);
const generalMenuIds = new Set([
  "line-oa-user-link",
  "form-design",
  "line-notify",
  "ai-provider",
  "copy-uat-dev"
]);

const systemTreeFolders = [
  { id: "company-settings", itemIds: companyMenuIds, label: { key: "company_settings", th: "ตั้งค่าบริษัท", en: "Company Settings" }, seedId: "company" },
  { id: "general-settings", itemIds: generalMenuIds, label: { key: "general_settings", th: "ตั้งค่าทั่วไป", en: "General Settings" }, seedId: "line-oa-user-link" },
];
const menuUiKeys = {
  closeTab: "close_tab",
  dashboardRoute: "dashboard",
  frequentMenu: "frequent_menu",
  hideMenu: "hide_menu",
  navigation: "erp_navigation",
  noFrequentMenu: "no_frequent_menu",
  openTabs: "open_tabs",
  noMenu: "no_menu_found",
  overview: "overview",
  overviewErp: "erp_overview",
  searchMenu: "search_menu_document_route",
  showMenu: "show_menu",
  items: "items",
} as const;

function mt(dictionary: BackendLanguageDictionary, key: keyof typeof menuUiKeys): string {
  const id = menuUiKeys[key];
  return backendText(dictionary, id);
}

function isRecord(value: unknown): value is SettingRecord {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : value === null || value === undefined ? "" : String(value).trim();
}

function stringArray(value: unknown): string[] {
  if (Array.isArray(value)) return Array.from(new Set(value.map(stringValue).filter(Boolean)));
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  try {
    const parsed = JSON.parse(trimmed) as unknown;
    return Array.isArray(parsed) ? Array.from(new Set(parsed.map(stringValue).filter(Boolean))) : [];
  } catch {
    return Array.from(new Set(trimmed.split(",").map((item) => item.trim()).filter(Boolean)));
  }
}

function normalizeSettingRecords(payload: unknown): SettingRecord[] {
  if (Array.isArray(payload)) return payload.filter(isRecord);
  if (!isRecord(payload)) return [];
  if (Array.isArray(payload.data)) return payload.data.filter(isRecord);
  return isRecord(payload.data) ? [payload.data] : [];
}

function WorkspaceContextPanel({
  language,
  mode,
  workspace,
}: {
  language: LanguageCode;
  mode: "sidebar" | "topbar";
  workspace: WorkspaceSession | null;
}) {
  if (!workspace) {
    return (
      <div className={cn(
        "min-w-0 rounded-xl border border-border bg-background/90 px-2 py-1 shadow-sm",
        mode === "topbar" && "flex-[1_1_18rem]",
      )}>
        <p className="text-xs font-bold leading-tight">BC Ai Account</p>
        <p className="text-[10px] leading-tight text-muted-foreground">Workspace</p>
      </div>
    );
  }

  const companyLabel = language === "th" ? "บริษัท" : "Company";
  const branchLabel = language === "th" ? "สาขา" : "Branch";
  const rows = [
    { icon: Crown, label: "กลุ่มกิจการ", value: holdingDisplayName(workspace) },
    { icon: Building2, label: companyLabel, value: workspaceCompanyDisplayName(workspace) },
    { icon: GitBranch, label: branchLabel, value: workspaceBranchDisplayName(workspace) },
  ];

  return (
    <div
      className={cn(
        "min-w-0 rounded-xl border border-border bg-background/90 shadow-sm",
        mode === "sidebar"
          ? "grid gap-1 p-1.5"
          : "grid w-full grid-cols-1 gap-1 p-1 sm:grid-cols-2 xl:grid-cols-[minmax(10rem,0.8fr)_minmax(20rem,1.45fr)_minmax(12rem,0.9fr)]",
      )}
    >
      {rows.map((row) => {
        const Icon = row.icon;
        return (
          <div
            className={cn(
              "min-w-0 rounded-lg border border-border/60 bg-card/80",
              mode === "sidebar"
                ? "grid grid-cols-[16px_52px_minmax(0,1fr)] items-start gap-1.5 px-1.5 py-1"
                : "grid grid-cols-[16px_58px_minmax(0,1fr)] items-start gap-1.5 px-2 py-1",
            )}
            key={row.label}
          >
            <Icon className="mt-0.5 h-3.5 w-3.5 shrink-0 text-primary" />
            {mode === "sidebar" || mode === "topbar" ? (
              <span className="pt-0.5 text-[9px] font-black uppercase leading-tight text-muted-foreground">
                {row.label}
              </span>
            ) : null}
            <div className="min-w-0">
              <p className="break-words text-xs font-bold leading-snug text-foreground">
                {row.value || "-"}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

async function fetchAllowedMenuIds(auth: AuthSession, workspace: WorkspaceSession): Promise<Set<string>> {
  const headers = {
    "Content-Type": "application/json",
    "x-bc-backend-url": auth.backendUrl,
    Authorization: `Bearer ${auth.token}`,
  };
  const holdingcode = encodeURIComponent(workspace.shop.holdingcode);
  try {
    const response = await authFetch(
      `/api/system-settings/permissiongroup/me?holdingcode=${holdingcode}`,
      { headers, cache: "no-store" },
    );
    if (!response.ok) return new Set();
    const payload = await response.json() as unknown;
    const rolePermission = normalizeSettingRecords(payload)[0];
    if (!rolePermission || rolePermission.isactive === false) return new Set();
    return new Set(stringArray(rolePermission.permissions));
  } catch {
    return new Set();
  }
}

type MainMenuScreenProps = {
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage?: LanguageCode;
};

export function MainMenuScreen(props: MainMenuScreenProps = {}) {
  return (
    <MenuQueryProvider>
      <MainMenuDashboard {...props} />
    </MenuQueryProvider>
  );
}

function MainMenuDashboard({ initialBackendLanguage, initialBackendUrl, initialLanguage = "th" }: MainMenuScreenProps) {
  const router = useRouter();
  const [language, setLanguage] = useState<LanguageCode>(initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [activeSection, setActiveSection] = useState("all");
  const [globalSearch, setGlobalSearch] = useState("");
  const [tabs, setTabs] = useState<WorkTab[]>([firstTab]);
  const [activeTabId, setActiveTabId] = useState(firstTab.id);
  const [expandedSections, setExpandedSections] = useState<string[]>([]);
  const [expandedGroups, setExpandedGroups] = useState<string[]>([]);
  const [menuUsageKey, setMenuUsageKey] = useState("");
  const [menuUsage, setMenuUsage] = useState<MenuUsageMap>({});
  const [menuLayout, setMenuLayout] = useState<MenuLayoutMode>("left");
  const [sidebarHidden, setSidebarHidden] = useState(false);
  const [topChromeHidden, setTopChromeHidden] = useState(false);
  const [lineDialog, setLineDialog] = useState<LineDialogState>(emptyLineDialog);
  const setLineNotice = pushNotice;
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const setPasswordNotice = pushNotice;
  const [profileAvatar, setProfileAvatar] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [allowedMenuIds, setAllowedMenuIds] = useState<Set<string>>(new Set());
  const [permissionRevision, setPermissionRevision] = useState(0);
  const linePollTimer = useRef<number | null>(null);
  const lastContentScrollTopRef = useRef(0);
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const authToken = auth?.token ?? "";
  const authBackendUrl = auth?.backendUrl ?? "";
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);
  const connectLineText = backendText(backendLanguage, "connect_line", t(language, "loginWithLine"));
  const lineLinkDescription = backendText(backendLanguage, "scan_qr_with_line", t(language, "lineLoginDescription"));
  const lineLinkSuccessText = backendText(backendLanguage, "link_line_success", t(language, "loginSuccess"));
  const lineLinkWaitingText = backendText(backendLanguage, "waiting_for_link", t(language, "lineLoginWaiting"));
  const requestFailedText = backendText(backendLanguage, "request_failed", language === "th" ? "เรียกข้อมูลไม่สำเร็จ" : "Request failed");
  const changePasswordText = backendText(backendLanguage, "change_password", language === "th" ? "เปลี่ยนรหัสผ่าน" : "Change password");
  const loginIdentity = auth?.profile?.email?.trim() || auth?.username?.trim() || "-";
  const loginText = backendText(backendLanguage, "login", language === "th" ? "เข้าสู่ระบบ" : "Login");
  const canAccessMenuItem = useCallback((item: MenuItem) => allowedMenuIds.has(item.id), [allowedMenuIds]);
  const allMenuItems = useMemo(() => flattenMenuItems(), []);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? initialLanguage);
    setLanguage(savedLanguage);
    document.documentElement.lang = savedLanguage;
    setMenuLayout(localStorage.getItem(menuLayoutStorageKey) === "top" ? "top" : "left");

    const savedAuth = getAuthSession();
    const workspaceRaw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!savedAuth) {
      router.replace("/");
      return;
    }
    if (!workspaceRaw) {
      router.replace("/workspace");
      return;
    }

    setAuth(savedAuth);
    const nextMenuUsageKey = menuUsageStorageKey(savedAuth);
    setMenuUsageKey(nextMenuUsageKey);
    setMenuUsage(readMenuUsage(localStorage, nextMenuUsageKey));

    try {
      setWorkspace(JSON.parse(workspaceRaw) as WorkspaceSession);
    } catch {
      router.replace("/workspace");
    }
  }, [initialLanguage, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem("user_language", language);
  }, [language]);

  useEffect(() => {
    localStorage.setItem(menuLayoutStorageKey, menuLayout);
    if (menuLayout === "top") setSidebarHidden(true);
  }, [menuLayout]);

  useEffect(() => {
    if (!authToken || !authBackendUrl) return;
    let cancelled = false;

    async function loadProfile() {
      try {
        const response = await authFetch(`/api/auth/profile?backendUrl=${encodeURIComponent(authBackendUrl)}`, {
          headers: {
            Authorization: `Bearer ${authToken}`,
            "x-bc-backend-url": authBackendUrl,
          },
          cache: "no-store",
        });
        const payload = await response.json() as ProfileResponse;
        if (cancelled || !response.ok || payload.success === false) return;
        setProfileAvatar(payload.data?.avatarthumb || payload.data?.avatar || "");
      } catch {
        // Profile picture is optional; authentication remains valid.
      }
    }

    void loadProfile();
    return () => {
      cancelled = true;
    };
  }, [authBackendUrl, authToken]);

  useEffect(() => {
    if (!auth || !workspace) return;
    const currentAuth = auth;
    const currentWorkspace = workspace;
    let cancelled = false;
    async function loadMenuPermissions() {
      const nextAllowed = await fetchAllowedMenuIds(currentAuth, currentWorkspace);
      if (!cancelled) setAllowedMenuIds(nextAllowed);
    }

    void loadMenuPermissions();
    return () => {
      cancelled = true;
    };
  }, [auth, permissionRevision, workspace]);

  useEffect(() => {
    const reloadPermissions = () => setPermissionRevision((current) => current + 1);
    window.addEventListener(WORKSPACE_CHANGED_EVENT, reloadPermissions);
    window.addEventListener("focus", reloadPermissions);
    return () => {
      window.removeEventListener(WORKSPACE_CHANGED_EVENT, reloadPermissions);
      window.removeEventListener("focus", reloadPermissions);
    };
  }, []);

  useEffect(() => {
    return () => stopLinePolling();
  }, []);

  const menuQuery = useQuery({
    queryKey: ["erp-menu-rows", language, backendLanguage],
    queryFn: () => fetchErpMenuRows(language, backendLanguage),
  });

  const rows = useMemo(() => menuQuery.data ?? [], [menuQuery.data]);
  const frequentMenuEntries = useMemo(() => getFrequentMenuEntries(allMenuItems, menuUsage, 20), [allMenuItems, menuUsage]);
  const activeWorkTab = useMemo(() => tabs.find((tab) => tab.id === activeTabId) ?? firstTab, [activeTabId, tabs]);
  const activeTabNeedsFixedViewport = activeWorkTab.route === "/productbarcode" || activeWorkTab.route === "/product" || activeWorkTab.route === "/productset" || activeWorkTab.route === "/datamodelgraph";

  function openMenuItem(
    item: MenuItem,
    options: { forceNew?: boolean; productCode?: string } = {},
  ) {
    if (!canAccessMenuItem(item)) return;
    const title = menuText(item.label, language, backendLanguage);
    const productCode = options.productCode?.trim().toUpperCase();
    const productFocusRequest = productCode
      ? { code: productCode, requestId: crypto.randomUUID() }
      : undefined;
    if (menuUsageKey) {
      setMenuUsage(recordMenuUsage(localStorage, menuUsageKey, item.id));
    }

    if (!options.forceNew) {
      const existingTab = tabs.find((tab) => tab.route === item.route);
      if (existingTab) {
        if (productFocusRequest) {
          setTabs((current) =>
            current.map((tab) =>
              tab.id === existingTab.id ? { ...tab, productFocusRequest } : tab,
            ),
          );
        }
        setActiveTabId(existingTab.id);
        return;
      }
    }

    const tabId = `${item.route}::${crypto.randomUUID()}`;
    setTabs((current) => {
      return [
        ...current,
        {
          id: tabId,
          title,
          route: item.route,
          item,
          closable: true,
          productFocusRequest,
        },
      ];
    });
    setActiveTabId(tabId);
  }

  function openMenuItemInNewTab(item: MenuItem) {
    openMenuItem(item, { forceNew: true });
  }

  function openOverview() {
    setActiveSection("all");
    setActiveTabId(firstTab.id);
    setGlobalSearch("");
    setExpandedSections([]);
    setExpandedGroups([]);
  }

  function toggleSection(sectionId: string) {
    setExpandedSections((current) =>
      current.includes(sectionId) ? current.filter((id) => id !== sectionId) : [...current, sectionId],
    );
    setActiveSection(sectionId);
  }

  function toggleGroup(groupKey: string) {
    setExpandedGroups((current) =>
      current.includes(groupKey) ? current.filter((id) => id !== groupKey) : [...current, groupKey],
    );
  }

  function closeTab(tabId: string) {
    setTabs((current) => {
      const nextTabs = current.filter((tab) => tab.id !== tabId);
      if (activeTabId === tabId) setActiveTabId(nextTabs[nextTabs.length - 1]?.id ?? firstTab.id);
      return nextTabs.length ? nextTabs : [firstTab];
    });
  }

  function reorderTabs(sourceId: string, targetId: string, side: TabInsertSide) {
    if (sourceId === targetId) return;
    setTabs((current) => {
      const sourceIndex = current.findIndex((tab) => tab.id === sourceId);
      const targetIndex = current.findIndex((tab) => tab.id === targetId);
      if (sourceIndex < 0 || targetIndex < 0) return current;
      let insertIndex = targetIndex + (side === "after" ? 1 : 0);
      if (sourceIndex < insertIndex) insertIndex -= 1;
      if (sourceIndex === insertIndex) return current;
      const nextTabs = [...current];
      const [movedTab] = nextTabs.splice(sourceIndex, 1);
      nextTabs.splice(insertIndex, 0, movedTab);
      return nextTabs;
    });
  }

  function stopLinePolling() {
    if (linePollTimer.current !== null) {
      window.clearInterval(linePollTimer.current);
      linePollTimer.current = null;
    }
  }

  async function handleLineLink() {
    if (!auth) return;

    stopLinePolling();
    setLineNotice(null);
    setLineDialog({ ...emptyLineDialog, open: true, loading: true });

    try {
      const response = await authFetch("/api/auth/line/code", { method: "POST" });
      const data = (await response.json()) as {
        success?: boolean;
        message?: string;
        code?: string;
        loginUrl?: string;
        expiresAt?: string;
      };

      if (!response.ok || !data.success || !data.code || !data.loginUrl) {
        throw new Error(data.message ?? requestFailedText);
      }

      const qrDataUrl = await QRCode.toDataURL(data.loginUrl, {
        margin: 1,
        width: 220,
        color: { dark: "#111827", light: "#ffffff" },
      });

      setLineDialog({
        open: true,
        loading: false,
        success: false,
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
        error: error instanceof Error ? error.message : requestFailedText,
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
        setLineNotice({ type: "error", text: t(language, "lineLoginExpired") });
        return;
      }
      void pollLineLink(code);
    }, DEFAULT_SOCIAL_POLL_INTERVAL_MS);
  }

  async function pollLineLink(code: string) {
    if (!auth) return;

    const response = await authFetch("/api/auth/line/link/status", {
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
      const message = data.message ?? requestFailedText;
      setLineNotice({ type: "error", text: message });
      setLineDialog((current) => ({ ...current, error: message }));
      return;
    }

    if (data.status !== "success") return;

    stopLinePolling();
    setLineNotice({ type: "success", text: lineLinkSuccessText });
    setLineDialog((current) => ({
      ...current,
      loading: false,
      success: true,
      error: "",
      expired: false,
    }));
  }

  function closeLineDialog() {
    stopLinePolling();
    setLineNotice(null);
    setLineDialog(emptyLineDialog);
  }

  async function copyLineLoginUrl() {
    if (!lineDialog.loginUrl) return;
    try {
      await navigator.clipboard.writeText(lineDialog.loginUrl);
      setLineNotice({ type: "success", text: t(language, "copied") });
    } catch {
      setLineNotice({ type: "error", text: requestFailedText });
    }
  }

  async function handleChangePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!auth) return;
    if (!currentPassword || !newPassword) {
      setPasswordNotice({ type: "error", text: backendText(backendLanguage, "please_enter_password", language === "th" ? "กรุณาใส่รหัสผ่าน" : "Please enter password") });
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordNotice({ type: "error", text: t(language, "passwordMismatch") });
      return;
    }
    if (newPassword.length < 15 || newPassword.length > 64) {
      setPasswordNotice({ type: "error", text: language === "th" ? "รหัสผ่านใหม่ต้องยาว 15–64 ตัวอักษร" : "The new password must be 15–64 characters." });
      return;
    }

    setPasswordSaving(true);
    setPasswordNotice(null);
    try {
      const response = await authFetch("/api/auth/profile", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${auth.token}`,
          "x-bc-backend-url": auth.backendUrl,
        },
        body: JSON.stringify({
          backendUrl: auth.backendUrl,
          currentpassword: currentPassword,
          newpassword: newPassword,
        }),
      });
      const payload = await response.json() as { success?: boolean; message?: string };
      if (!response.ok || payload.success === false) throw new Error(payload.message ?? requestFailedText);

      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setPasswordDialogOpen(false);
      clearAuthSession();
      localStorage.removeItem(workspaceStorageKeys.workspace);
      localStorage.removeItem(workspaceStorageKeys.shopInfo);
      localStorage.removeItem(workspaceStorageKeys.branch);
      router.replace("/");
    } catch (error) {
      setPasswordNotice({ type: "error", text: error instanceof Error ? backendText(backendLanguage, error.message, error.message) : requestFailedText });
    } finally {
      setPasswordSaving(false);
    }
  }

  async function logout() {
    await logoutAuthSession();
    localStorage.removeItem(workspaceStorageKeys.workspace);
    localStorage.removeItem(workspaceStorageKeys.shopInfo);
    localStorage.removeItem(workspaceStorageKeys.branch);
    router.replace("/");
  }

  function handleContentScroll(event: UIEvent<HTMLDivElement>) {
    const nextScrollTop = event.currentTarget.scrollTop;

    if (topChromeHidden) {
      setTopChromeHidden(false);
    }

    lastContentScrollTopRef.current = nextScrollTop;
  }

  const showLeftMenu = menuLayout === "left" && !sidebarHidden;
  const menuLayoutLeftText = backendText(backendLanguage, "menu_layout_left", language === "th" ? "เมนูซ้าย" : "Left menu");
  const menuLayoutTopText = backendText(backendLanguage, "menu_layout_top", language === "th" ? "เมนูบน" : "Top menu");

  return (
    <main className="min-h-dvh overflow-x-hidden bg-background text-foreground lg:h-dvh lg:min-h-0 lg:max-h-dvh lg:overflow-hidden">
      <div className={cn("grid min-h-dvh min-w-0 grid-cols-[minmax(0,1fr)] lg:h-dvh lg:min-h-0 lg:max-h-dvh lg:overflow-hidden", showLeftMenu && "lg:grid-cols-[280px_minmax(0,1fr)]")}>
        {showLeftMenu ? (
        <aside className="flex max-h-dvh min-w-0 flex-col overflow-x-hidden overflow-y-auto overscroll-contain border-b border-border bg-card/80 p-3 lg:sticky lg:top-0 lg:h-dvh lg:border-b-0 lg:border-r">
          <label className="relative mb-3 block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="!pl-10" placeholder={mt(backendLanguage, "searchMenu")} value={globalSearch} onChange={(event) => setGlobalSearch(event.target.value)} />
          </label>

          <nav aria-label={mt(backendLanguage, "navigation")} className="grid w-full max-w-full gap-2" role="tree">
            <SidebarButton active={activeSection === "all" && activeTabId === firstTab.id} count={rows.length} icon={<LayoutDashboard className="h-4 w-4" />} label={mt(backendLanguage, "overview")} onClick={openOverview} />
            {MENU_SECTIONS.map((section) => {
              const label = menuText(section.title, language, backendLanguage);
              return (
                <MenuSectionAccordion
                  active={activeSection === section.id}
                  backendLanguage={backendLanguage}
                  canAccessMenuItem={canAccessMenuItem}
                  count={countSectionItems(section)}
                  expanded={expandedSections.includes(section.id) || Boolean(globalSearch.trim())}
                  expandedGroups={expandedGroups}
                  key={section.id}
                  language={language}
                  label={label}
                  onOpenNewItem={openMenuItemInNewTab}
                  onOpenItem={openMenuItem}
                  onToggleGroup={toggleGroup}
                  onToggle={() => toggleSection(section.id)}
                  search={globalSearch}
                  section={section}
                />
              );
            })}
          </nav>
          <div className="pointer-events-none sticky bottom-3 z-20 mt-auto flex justify-end pt-3">
            <Button
              type="button"
              variant="outline"
              size="icon"
              className="pointer-events-auto h-9 w-9 rounded-xl bg-background/95 shadow-lg backdrop-blur"
              aria-label={mt(backendLanguage, "hideMenu")}
              title={mt(backendLanguage, "hideMenu")}
              onClick={() => setSidebarHidden(true)}
            >
              <PanelLeftClose className="h-4 w-4" />
            </Button>
          </div>
        </aside>
        ) : null}

        <section
          className="min-w-0 lg:flex lg:h-dvh lg:min-h-0 lg:flex-col lg:overflow-hidden"
          onPointerMove={(event) => {
            if (topChromeHidden && event.clientY < 26) setTopChromeHidden(false);
          }}
        >
          <header
            className="menu-top-chrome sticky top-0 z-30 shrink-0 border-b border-border bg-background/90 px-2 py-1 backdrop-blur"
            data-hidden={topChromeHidden ? "true" : "false"}
          >
            <div className="grid min-w-0 gap-1.5">
              <WorkspaceContextPanel language={language} mode="topbar" workspace={workspace} />
              <div className="flex min-w-0 flex-wrap items-center gap-1.5">
                <div className="flex min-w-0 shrink-0 items-center gap-1.5">
                {menuLayout === "left" && sidebarHidden ? (
                  <Button type="button" variant="outline" size="icon" className="shrink-0" aria-label={mt(backendLanguage, "showMenu")} title={mt(backendLanguage, "showMenu")} onClick={() => setSidebarHidden(false)}>
                    <PanelLeftOpen className="h-4 w-4" />
                  </Button>
                ) : null}
                <div className="flex shrink-0 rounded-lg border border-border bg-card p-0.5">
                  <Button
                    type="button"
                    variant={menuLayout === "left" ? "secondary" : "ghost"}
                    size="sm"
                    className="h-8 gap-1 px-2"
                    aria-label={menuLayoutLeftText}
                    title={menuLayoutLeftText}
                    onClick={() => {
                      setMenuLayout("left");
                      setSidebarHidden(false);
                    }}
                  >
                    <PanelLeftOpen className="h-4 w-4" />
                    <span className="hidden sm:inline">{menuLayoutLeftText}</span>
                  </Button>
                  <Button
                    type="button"
                    variant={menuLayout === "top" ? "secondary" : "ghost"}
                    size="sm"
                    className="h-8 gap-1 px-2"
                    aria-label={menuLayoutTopText}
                    title={menuLayoutTopText}
                    onClick={() => setMenuLayout("top")}
                  >
                    <LayoutPanelTop className="h-4 w-4" />
                    <span className="hidden sm:inline">{menuLayoutTopText}</span>
                  </Button>
                </div>
                </div>
                <label className="relative min-w-52 flex-[1_1_22rem]">
                  <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input className="h-8 !pl-10" placeholder={mt(backendLanguage, "searchMenu")} value={globalSearch} onChange={(event) => setGlobalSearch(event.target.value)} />
                </label>
                <div className="ml-auto flex min-w-0 flex-wrap items-center justify-end gap-1.5">
                <Button variant="outline" size="icon" className="h-8 w-8" aria-label={backendText(backendLanguage, "notification")}>
                  <Bell className="h-4 w-4" />
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" className="h-8 gap-1 px-2">
                      <Star className="h-4 w-4" />
                      {mt(backendLanguage, "frequentMenu")}
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="max-h-[70vh] w-72 overflow-y-auto">
                    {frequentMenuEntries.length ? (
                      frequentMenuEntries.map(({ item, percent }) => (
                        <DropdownMenuItem key={item.id} className="gap-2" disabled={!canAccessMenuItem(item)} onClick={() => openMenuItem(item)}>
                          <MenuRouteIcon item={item} size={15} />
                          <span className="min-w-0 flex-1 truncate">{menuText(item.label, language, backendLanguage)}</span>
                          {canAccessMenuItem(item) ? null : <Lock className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />}
                          <Badge variant="secondary" className="shrink-0">{percent}%</Badge>
                        </DropdownMenuItem>
                      ))
                    ) : (
                      <DropdownMenuItem disabled>{mt(backendLanguage, "noFrequentMenu")}</DropdownMenuItem>
                    )}
                  </DropdownMenuContent>
                </DropdownMenu>
                <AppHeaderControls language={language} onLanguageChange={setLanguage} showSettings={false} />
                <ManualLink compact language={language} screen="menu" />
                <ZoomControl dictionary={backendLanguage} language={language} />
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-auto min-h-8 max-w-[min(18rem,34vw)] shrink-0 items-center gap-1.5 px-2 py-1 text-left"
                      aria-label={loginIdentity}
                      title={loginIdentity}
                    >
                      <LogoAvatar
                        uri={profileAvatar}
                        auth={auth}
                        alt={loginIdentity}
                        sizeClass="size-6 rounded-full shrink-0"
                        iconSize={14}
                        width={48}
                        fallbackIcon={UserRound}
                      />
                      <span className="hidden min-w-0 max-w-56 whitespace-normal break-all text-xs font-medium leading-tight sm:inline">
                        {loginIdentity}
                      </span>
                      <ChevronsUpDown className="h-3.5 w-3.5 opacity-60" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-72">
                    <div className="flex min-w-0 items-center gap-2 rounded-lg px-2 py-2 text-sm">
                      <LogoAvatar
                        uri={profileAvatar}
                        auth={auth}
                        alt={loginIdentity}
                        sizeClass="size-9 rounded-full shrink-0"
                        iconSize={18}
                        width={64}
                        fallbackIcon={UserRound}
                      />
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-muted-foreground">{loginText}</p>
                        <p className="break-all font-semibold leading-snug text-foreground">{loginIdentity}</p>
                      </div>
                    </div>
                    <DropdownMenuItem onClick={() => setPasswordDialogOpen(true)}>
                      <KeyRound className="h-4 w-4" />
                      {changePasswordText}
                    </DropdownMenuItem>
                    <DropdownMenuItem disabled={!auth || lineDialog.loading} onClick={() => void handleLineLink()}>
                      {lineDialog.loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <MessageCircle className="h-4 w-4" />}
                      {connectLineText}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={() => router.push("/workspace")}>
                      <Store className="h-4 w-4" />
                      {backendText(backendLanguage, "change_company")}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={logout}>
                    <LogOut className="h-4 w-4" />
                      {backendText(backendLanguage, "logout")}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                </div>
              </div>
            </div>
          </header>
          {menuLayout === "top" ? (
            <TopMenuChrome
              activeSection={activeSection}
              backendLanguage={backendLanguage}
              canAccessMenuItem={canAccessMenuItem}
              language={language}
              onOpenNewItem={openMenuItemInNewTab}
              onOpenItem={openMenuItem}
              onOpenOverview={openOverview}
              onSelectSection={setActiveSection}
            />
          ) : null}

          <div className="grid min-w-0 grid-cols-[minmax(0,1fr)] gap-2 p-2 lg:min-h-0 lg:flex-1 lg:grid-rows-[auto_minmax(0,1fr)] lg:overflow-hidden">
            <div
              className="menu-tabs-chrome min-w-0 overflow-hidden"
              data-hidden={topChromeHidden ? "true" : "false"}
            >
              <OpenTabs tabs={tabs} activeTabId={activeTabId} backendLanguage={backendLanguage} language={language} onSelect={setActiveTabId} onClose={closeTab} onReorder={reorderTabs} />
            </div>

            <div
              className={cn(
                "min-w-0 lg:min-h-0 lg:overscroll-contain",
                activeTabNeedsFixedViewport ? "lg:h-full lg:overflow-hidden" : "lg:overflow-y-auto lg:pr-1",
              )}
              onScroll={handleContentScroll}
            >
              {menuQuery.isLoading ? (
                <DashboardLoading backendLanguage={backendLanguage} />
              ) : menuQuery.isError ? (
                <Card className="border-destructive/30">
                  <CardContent className="flex items-center gap-3 p-6 text-destructive">
                    <AlertTriangle className="h-5 w-5" />
                    {backendText(backendLanguage, "menu_load_failed")}
                    <Button variant="outline" size="sm" onClick={() => void menuQuery.refetch()}>
                      <RefreshCcw className="h-4 w-4" />
                      {backendText(backendLanguage, "try_again")}
                    </Button>
                  </CardContent>
                </Card>
              ) : (
                <div className={cn("grid min-w-0 grid-cols-[minmax(0,1fr)]", activeTabNeedsFixedViewport && "lg:h-full lg:min-h-0")}>
                  {tabs.map((tab) => (
                    <section
                      aria-hidden={tab.id !== activeTabId}
                      className={cn("min-w-0", (tab.route === "/productbarcode" || tab.route === "/product" || tab.route === "/productset" || tab.route === "/datamodelgraph") && "lg:h-full lg:min-h-0 lg:overflow-hidden")}
                      hidden={tab.id !== activeTabId}
                      key={tab.id}
                      role="tabpanel"
                    >
                      {tab.id === "home" ? (
                        <DashboardHome />
                      ) : (
                        <WorkTabPanel
                          active={tab.id === activeTabId}
                          activeTab={tab}
                          backendLanguage={backendLanguage}
                          language={language}
                          onOpenRoute={(route, productCode) => {
                             const item = allMenuItems.find((candidate) => candidate.route === route);
                             if (item) openMenuItem(item, { productCode });
                           }}
                          tabCount={tabs.length}
                        />
                      )}
                    </section>
                  ))}
                </div>
              )}
            </div>
          </div>
        </section>
      </div>

      {passwordDialogOpen ? (
        <div className="dialog-backdrop" role="presentation">
          <form className="line-login-dialog" aria-label={changePasswordText} role="dialog" aria-modal="true" onSubmit={handleChangePassword}>
            <div className="dialog-header">
              <div>
                <p className="eyebrow">{loginText}</p>
                <h2>{changePasswordText}</h2>
              </div>
              <button className="icon-button dialog-close" type="button" onClick={() => setPasswordDialogOpen(false)} aria-label={t(language, "lineLoginClose")}>
                ×
              </button>
            </div>
            <label className="grid gap-1 text-sm font-medium">
              <span>{backendText(backendLanguage, "current_password", language === "th" ? "รหัสผ่านปัจจุบัน" : "Current password")}</span>
              <Input value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} type="password" autoComplete="current-password" />
            </label>
            <label className="grid gap-1 text-sm font-medium">
              <span>{backendText(backendLanguage, "new_password", language === "th" ? "รหัสผ่านใหม่" : "New password")}</span>
              <Input value={newPassword} onChange={(event) => setNewPassword(event.target.value)} type="password" autoComplete="new-password" />
            </label>
            <label className="grid gap-1 text-sm font-medium">
              <span>{backendText(backendLanguage, "confirmPassword", t(language, "confirmPassword"))}</span>
              <Input value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} type="password" autoComplete="new-password" />
            </label>
            <div className="line-dialog-actions">
              <button className="secondary-button" type="button" onClick={() => setPasswordDialogOpen(false)}>
                {t(language, "lineLoginClose")}
              </button>
              <button className="primary-button" type="submit" disabled={passwordSaving}>
                {passwordSaving ? <Loader2 className="spin" size={17} /> : <KeyRound aria-hidden="true" size={17} />}
                <span>{changePasswordText}</span>
              </button>
            </div>
          </form>
        </div>
      ) : null}

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
            ) : lineDialog.success ? (
              <div className="message success">
                <CheckCircle2 size={18} />
                <span>{lineLinkSuccessText}</span>
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

function TopMenuChrome({
  activeSection,
  backendLanguage,
  canAccessMenuItem,
  language,
  onOpenNewItem,
  onOpenItem,
  onOpenOverview,
  onSelectSection,
}: {
  activeSection: string;
  backendLanguage: BackendLanguageDictionary;
  canAccessMenuItem: (item: MenuItem) => boolean;
  language: LanguageCode;
  onOpenNewItem: (item: MenuItem) => void;
  onOpenItem: (item: MenuItem) => void;
  onOpenOverview: () => void;
  onSelectSection: (sectionId: string) => void;
}) {
  const [openSectionId, setOpenSectionId] = useState<string | null>(null);
  const [activeGroupId, setActiveGroupId] = useState<string | null>(null);
  const [activeFolderId, setActiveFolderId] = useState<string | null>(null);
  const [activeFolderIndex, setActiveFolderIndex] = useState<number>(0);
  const [openSectionLeft, setOpenSectionLeft] = useState(0);
  const [flyoutSide, setFlyoutSide] = useState<"left" | "right">("right");
  const menuRootRef = useRef<HTMLDivElement | null>(null);
  const openSection = MENU_SECTIONS.find((section) => section.id === openSectionId) ?? null;
  const visibleGroups = openSection
    ? getVisibleGroups(openSection, language, "", backendLanguage)
        .map((group) => ({ group, items: getVisibleItems(group.items, language, "", backendLanguage) }))
        .filter((entry) => entry.items.length > 0)
    : [];
  const activeGroupEntry = visibleGroups.find((entry) => entry.group.id === activeGroupId) ?? visibleGroups[0] ?? null;
  const activeGroupIndex = Math.max(0, visibleGroups.findIndex((entry) => entry.group.id === activeGroupEntry?.group.id));

  useEffect(() => {
    if (!openSectionId) {
      setActiveFolderId(null);
      setActiveFolderIndex(0);
      return;
    }

    function closeOnOutsidePointer(event: PointerEvent) {
      const target = event.target;
      if (target instanceof Node && menuRootRef.current?.contains(target)) return;
      setOpenSectionId(null);
      setActiveGroupId(null);
      setActiveFolderId(null);
      setActiveFolderIndex(0);
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key !== "Escape") return;
      setOpenSectionId(null);
      setActiveGroupId(null);
      setActiveFolderId(null);
      setActiveFolderIndex(0);
    }

    document.addEventListener("pointerdown", closeOnOutsidePointer);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnOutsidePointer);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [openSectionId]);

  function positionOpenSection(target: HTMLElement, section: MenuSection) {
    const rect = target.getBoundingClientRect();
    const firstPanelWidth = 434;
    const secondPanelWidth = 480;
    const gap = 4;
    const viewportPadding = 12;
    setOpenSectionLeft(target.offsetLeft);
    setFlyoutSide(rect.left + firstPanelWidth + gap + secondPanelWidth <= window.innerWidth - viewportPadding ? "right" : "left");
    setOpenSectionId(section.id);
    setActiveGroupId(getVisibleGroups(section, language, "", backendLanguage)[0]?.id ?? null);
    setActiveFolderId(null);
    setActiveFolderIndex(0);
  }

  function closeTopMenu() {
    setOpenSectionId(null);
    setActiveGroupId(null);
    setActiveFolderId(null);
    setActiveFolderIndex(0);
  }

  function topMenuItemRow(item: MenuItem, key: string, onMouseEnter?: () => void) {
    const locked = !canAccessMenuItem(item);
    const label = menuText(item.label, language, backendLanguage);
    const newTabLabel = language === "th" ? `เปิดแท็บใหม่ ${label}` : `Open new tab ${label}`;

    return (
      <div
        key={key}
        className={cn(
          "group/top-item flex min-h-10 w-full min-w-0 items-center gap-1 rounded-lg border border-transparent px-1.5 py-1 text-sm text-muted-foreground transition-colors hover:border-border/70 hover:bg-accent hover:text-foreground",
          locked && "cursor-not-allowed opacity-55",
        )}
        onMouseEnter={onMouseEnter}
      >
        <button
          type="button"
          disabled={locked}
          className="flex min-h-8 min-w-0 flex-1 items-center gap-2 rounded-md px-1 text-left disabled:cursor-not-allowed"
          onClick={() => {
            closeTopMenu();
            onOpenItem(item);
          }}
        >
          <span className="grid h-7 w-7 shrink-0 place-items-center rounded-md bg-muted text-muted-foreground transition-colors group-hover/top-item:bg-background group-hover/top-item:text-primary">
            <MenuRouteIcon item={item} size={15} />
          </span>
          <span className="line-clamp-2 min-w-0 flex-1 break-words font-medium leading-5">{label}</span>
          {locked ? <Lock className="h-3.5 w-3.5 shrink-0 text-muted-foreground" /> : null}
        </button>
        {locked ? null : (
          <button
            type="button"
            className="grid h-7 w-7 shrink-0 place-items-center rounded-md border border-transparent text-muted-foreground transition-colors hover:border-border hover:bg-background hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
            aria-label={newTabLabel}
            title={newTabLabel}
            onClick={(event) => {
              event.stopPropagation();
              closeTopMenu();
              onOpenNewItem(item);
            }}
          >
            <Plus className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
    );
  }

  const hasSingleGroup = openSection ? openSection.groups.length === 1 : false;

  // Find active folder node under active group
  const activeFolderNode = (() => {
    if (!activeFolderId) return null;
    const targetGroup = hasSingleGroup ? openSection?.groups[0] : activeGroupEntry?.group;
    if (!targetGroup) return null;
    return getMenuTreeNodes(targetGroup, language, "", backendLanguage).find(
      (node) => node.type === "folder" && node.id === activeFolderId
    ) as Extract<MenuTreeNode, { type: "folder" }> | undefined;
  })();

  return (
    <nav className="shrink-0 border-b border-border bg-card/95 px-2 py-1" aria-label={mt(backendLanguage, "navigation")}>
      <div className="relative" ref={menuRootRef}>
      <div className="flex min-w-0 gap-1 overflow-x-auto">
        <Button
          type="button"
          variant={activeSection === "all" ? "secondary" : "ghost"}
          size="sm"
          className="h-7 shrink-0 gap-1 px-2 text-[12px]"
          onClick={() => {
            setOpenSectionId(null);
            setActiveGroupId(null);
            setActiveFolderId(null);
            onOpenOverview();
          }}
        >
          <LayoutDashboard className="h-4 w-4" />
          {mt(backendLanguage, "overview")}
        </Button>
        {MENU_SECTIONS.map((section) => {
          const sectionLabel = menuText(section.title, language, backendLanguage);
          return (
            <Button
              key={section.id}
              type="button"
              variant={activeSection === section.id ? "secondary" : "ghost"}
              size="sm"
              className="h-7 shrink-0 gap-1 px-2 text-[12px]"
              aria-expanded={openSectionId === section.id}
              onMouseEnter={(event) => {
                onSelectSection(section.id);
                positionOpenSection(event.currentTarget, section);
              }}
              onFocus={(event) => {
                onSelectSection(section.id);
                positionOpenSection(event.currentTarget, section);
              }}
              onClick={(event) => {
                onSelectSection(section.id);
                const rect = event.currentTarget.getBoundingClientRect();
                const firstPanelWidth = 434;
                const secondPanelWidth = 480;
                setOpenSectionLeft(event.currentTarget.offsetLeft);
                setFlyoutSide(rect.left + firstPanelWidth + 4 + secondPanelWidth <= window.innerWidth - 12 ? "right" : "left");
                setOpenSectionId((current) => {
                  const next = current === section.id ? null : section.id;
                  setActiveGroupId(next ? getVisibleGroups(section, language, "", backendLanguage)[0]?.id ?? null : null);
                  setActiveFolderId(null);
                  return next;
                });
              }}
            >
              <SectionIcon sectionId={section.id} />
              <span>{sectionLabel}</span>
              <Badge variant="outline" className="h-4 px-1.5 text-[10px]">{countSectionItems(section)}</Badge>
              <ChevronDown className={cn("h-3.5 w-3.5 opacity-70 transition", openSectionId === section.id && "rotate-180")} />
            </Button>
          );
        })}
      </div>
      {openSection ? (
        <div
          className="absolute top-[calc(100%+6px)] z-40 flex overflow-visible rounded-lg border border-border bg-popover text-popover-foreground shadow-lg"
          style={{ left: `${openSectionLeft}px` }}
        >
          {hasSingleGroup && openSection.groups[0] ? (
            <>
              {/* Column 1: Items and folders of the single group */}
              <div className="w-72 min-w-0 max-w-[calc(100vw-1.5rem)] overflow-y-auto p-1.5">
                <div className="mb-1 px-2 text-xs font-semibold text-muted-foreground">
                  {menuText(openSection.title, language, backendLanguage)}
                </div>
                <div className="grid gap-1">
                  {getMenuTreeNodes(openSection.groups[0], language, "", backendLanguage).map((node, nodeIndex) => {
                    if (node.type === "folder") {
                      const active = activeFolderId === node.id;
                      return (
                        <button
                          key={node.id}
                          type="button"
                          className={cn(
                            "flex h-9 w-full min-w-0 items-center justify-between gap-2 rounded-md px-2 text-left text-sm hover:bg-muted",
                            active && "bg-primary/10 text-primary"
                          )}
                          onMouseEnter={() => {
                            setActiveFolderId(node.id);
                            setActiveFolderIndex(nodeIndex);
                          }}
                          onClick={() => {
                            setActiveFolderId(node.id);
                            setActiveFolderIndex(nodeIndex);
                          }}
                        >
                          <span className="flex min-w-0 items-center gap-2">
                            {node.seedItem ? <MenuRouteIcon item={node.seedItem} size={15} /> : <Command className="h-4 w-4" />}
                            <span className="truncate">{node.label}</span>
                          </span>
                          <span className="flex shrink-0 items-center gap-1">
                            <Badge variant={active ? "secondary" : "outline"} className="h-5 px-1.5">{node.children.length}</Badge>
                            <ChevronDown className={cn("h-3.5 w-3.5 -rotate-90 opacity-70", active && "opacity-100")} />
                          </span>
                        </button>
                      );
                    }

                    return topMenuItemRow(node.item, node.id, () => {
                      setActiveFolderId(null);
                      setActiveFolderIndex(0);
                    });
                  })}
                </div>
              </div>

              {/* Column 2: sub-items of the active folder */}
              {activeFolderNode ? (
                <div
                  className={cn(
                    "absolute w-80 max-h-[72vh] overflow-y-auto rounded-lg border border-border bg-popover p-1.5 shadow-lg text-popover-foreground",
                    flyoutSide === "left" ? "right-[calc(100%+4px)]" : "left-[calc(100%+4px)]"
                  )}
                  style={{ top: `${28 + activeFolderIndex * 40}px` }}
                >
                  <div className="mb-1 flex items-center justify-between gap-2 px-2 text-xs font-semibold text-muted-foreground">
                    <span className="truncate">{activeFolderNode.label}</span>
                    <Badge variant="secondary" className="h-5 px-1.5">{activeFolderNode.children.length}</Badge>
                  </div>
                  <div className="grid gap-1">
                    {activeFolderNode.children.map((item) => {
                      return topMenuItemRow(item, item.id);
                    })}
                  </div>
                </div>
              ) : null}
            </>
          ) : (
            <>
              {/* Column 1: Groups list */}
              <div className="w-72 min-w-0 max-w-[calc(100vw-1.5rem)] overflow-y-auto p-1.5">
                <div className="mb-1 px-2 text-xs font-semibold text-muted-foreground">
                  {menuText(openSection.title, language, backendLanguage)}
                </div>
                <div className="grid gap-1">
                  {visibleGroups.map(({ group, items }) => {
                    const active = activeGroupEntry?.group.id === group.id;
                    return (
                      <button
                        key={group.id}
                        type="button"
                        className={cn(
                          "flex h-9 min-w-0 items-center justify-between gap-2 rounded-md px-2 text-left text-sm hover:bg-muted",
                          active && "bg-primary/10 text-primary",
                        )}
                        onClick={() => {
                          setActiveGroupId(group.id);
                          setActiveFolderId(null);
                        }}
                        onMouseEnter={() => {
                          setActiveGroupId(group.id);
                          setActiveFolderId(null);
                        }}
                      >
                        <span className="truncate">{menuText(group.title, language, backendLanguage)}</span>
                        <span className="flex shrink-0 items-center gap-1">
                          <Badge variant={active ? "secondary" : "outline"} className="h-5 px-1.5">{items.length}</Badge>
                          <ChevronDown className={cn("h-3.5 w-3.5 -rotate-90 opacity-70", active && "opacity-100")} />
                        </span>
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Column 2: Items of the active group */}
              {activeGroupEntry ? (
                <div
                  className="absolute z-50 w-80 max-w-[calc(100vw-19rem)] overflow-y-auto rounded-lg border border-border bg-popover p-1.5 text-popover-foreground shadow-lg"
                  style={{
                    maxHeight: "72vh",
                    top: `${28 + activeGroupIndex * 40}px`,
                    ...(flyoutSide === "left" ? { right: "calc(100% + 4px)" } : { left: "calc(100% + 4px)" }),
                  }}
                >
                  <div className="mb-1 flex items-center justify-between gap-2 px-2 text-xs font-semibold text-muted-foreground">
                    <span className="truncate">{menuText(activeGroupEntry.group.title, language, backendLanguage)}</span>
                    <Badge variant="secondary" className="h-5 px-1.5">{activeGroupEntry.items.length}</Badge>
                  </div>
                  <div className="grid gap-1">
                    {activeGroupEntry.items.map((item) => {
                      return topMenuItemRow(item, item.id);
                    })}
                  </div>
                </div>
              ) : null}
            </>
          )}
        </div>
      ) : null}
      </div>
    </nav>
  );
}

function SidebarButton({
  active,
  count,
  icon,
  label,
  onClick,
}: {
  active: boolean;
  count: number;
  icon: ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      role="treeitem"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "flex w-full min-w-0 items-center justify-between gap-3 rounded-2xl border border-transparent px-3 py-2 text-left text-xs! font-medium transition-colors",
        active ? "border-border bg-primary text-primary-foreground shadow-sm" : "bg-transparent text-muted-foreground hover:bg-muted hover:text-foreground",
      )}
    >
      <span className="flex min-w-0 items-center gap-2">
        {icon}
        <span className="truncate">{label}</span>
      </span>
      <Badge variant={active ? "secondary" : "outline"}>{count}</Badge>
    </button>
  );
}

function MenuSectionAccordion({
  active,
  backendLanguage,
  canAccessMenuItem,
  count,
  expanded,
  expandedGroups,
  label,
  language,
  onOpenNewItem,
  onOpenItem,
  onToggleGroup,
  onToggle,
  search,
  section,
}: {
  active: boolean;
  backendLanguage: BackendLanguageDictionary;
  canAccessMenuItem: (item: MenuItem) => boolean;
  count: number;
  expanded: boolean;
  expandedGroups: string[];
  label: string;
  language: LanguageCode;
  onOpenNewItem: (item: MenuItem) => void;
  onOpenItem: (item: MenuItem) => void;
  onToggleGroup: (groupKey: string) => void;
  onToggle: () => void;
  search: string;
  section: MenuSection;
}) {
  const visibleGroups = getVisibleGroups(section, language, search, backendLanguage);
  const hasSearch = Boolean(search.trim());

  return (
    <div className={cn("grid gap-1 rounded-2xl", active && "bg-primary/5")} role="treeitem" aria-expanded={expanded} aria-selected={active}>
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={expanded}
        className={cn(
          "flex w-full min-w-0 items-center justify-between gap-3 rounded-2xl border px-3 py-2 text-left text-xs! font-medium transition-colors",
          active ? "border-border bg-primary text-primary-foreground shadow-sm" : "border-transparent bg-transparent text-muted-foreground hover:bg-muted hover:text-foreground",
        )}
      >
        <span className="flex min-w-0 items-center gap-2">
          <SectionIcon sectionId={section.id} />
          <span className="truncate">{label}</span>
        </span>
        <span className="flex shrink-0 items-center gap-2">
          <Badge variant={active ? "secondary" : "outline"}>{count}</Badge>
          <ChevronDown className={cn("h-4 w-4 transition-transform", expanded && "rotate-180")} aria-hidden="true" />
        </span>
      </button>

      {expanded ? (
        <div className="grid gap-1 rounded-2xl border border-border bg-background/70 p-1.5" role="group">
          {visibleGroups.length ? (
            section.groups.length === 1 ? (
              (() => {
                const group = section.groups[0];
                const nodes = getMenuTreeNodes(group, language, search, backendLanguage);
                return nodes.map((node) => node.type === "folder" ? (
                  <MenuTreeFolder
                    backendLanguage={backendLanguage}
                    canAccessMenuItem={canAccessMenuItem}
                    folder={node}
                    key={node.id}
                    language={language}
                    onOpenNewItem={onOpenNewItem}
                    onOpenItem={onOpenItem}
                    search={search}
                  />
                ) : (
                  <MenuTreeItemButton
                    backendLanguage={backendLanguage}
                    isLocked={!canAccessMenuItem(node.item)}
                    item={node.item}
                    key={node.id}
                    language={language}
                    onOpenNewItem={onOpenNewItem}
                    onOpenItem={onOpenItem}
                  />
                ));
              })()
            ) : (
              visibleGroups.map((group) => {
                const groupKey = `${section.id}:${group.id}`;
                return (
                  <MenuTreeGroup
                    backendLanguage={backendLanguage}
                    canAccessMenuItem={canAccessMenuItem}
                    expanded={expandedGroups.includes(groupKey) || Boolean(search.trim())}
                    group={group}
                    groupKey={groupKey}
                    key={group.id}
                    language={language}
                    onOpenNewItem={onOpenNewItem}
                    onOpenItem={onOpenItem}
                    onToggle={() => onToggleGroup(groupKey)}
                    search={search}
                  />
                );
              })
            )
          ) : (
            <div className="rounded-xl border border-dashed border-border px-3 py-2 text-sm text-muted-foreground">{mt(backendLanguage, "noMenu")}</div>
          )}
          {hasSearch ? null : <span className="px-2 pt-1 text-[11px] font-semibold text-muted-foreground">{count} {mt(backendLanguage, "items")}</span>}
        </div>
      ) : null}
    </div>
  );
}

function MenuTreeGroup({
  backendLanguage,
  canAccessMenuItem,
  expanded,
  group,
  groupKey,
  language,
  onOpenNewItem,
  onOpenItem,
  onToggle,
  search,
}: {
  backendLanguage: BackendLanguageDictionary;
  canAccessMenuItem: (item: MenuItem) => boolean;
  expanded: boolean;
  group: MenuGroup;
  groupKey: string;
  language: LanguageCode;
  onOpenNewItem: (item: MenuItem) => void;
  onOpenItem: (item: MenuItem) => void;
  onToggle: () => void;
  search: string;
}) {
  const nodes = getMenuTreeNodes(group, language, search, backendLanguage);
  const count = nodes.reduce((total, node) => total + (node.type === "folder" ? node.children.length : 1), 0);
  if (!nodes.length) return null;

  return (
    <div className="grid gap-1 rounded-xl border border-border bg-card" role="treeitem" aria-expanded={expanded} aria-selected={expanded}>
      <button
        aria-controls={`menu-tree-${groupKey}`}
        aria-expanded={expanded}
        className="grid min-h-9 w-full grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 px-2 py-1.5 text-left text-sm font-semibold"
        onClick={onToggle}
        type="button"
      >
        <ChevronDown className={cn("h-3.5 w-3.5 text-muted-foreground transition-transform", expanded && "rotate-180")} aria-hidden="true" />
        <span className="min-w-0 truncate">{menuText(group.title, language, backendLanguage)}</span>
        <Badge variant="outline" className="h-6 px-2 text-[11px]">{count}</Badge>
      </button>
      {expanded ? (
      <div id={`menu-tree-${groupKey}`} className="ml-4 grid gap-1 border-l border-border/80 pb-1 pl-2" role="group">
        {nodes.map((node) => node.type === "folder" ? (
          <MenuTreeFolder backendLanguage={backendLanguage} canAccessMenuItem={canAccessMenuItem} folder={node} key={node.id} language={language} onOpenNewItem={onOpenNewItem} onOpenItem={onOpenItem} search={search} />
        ) : (
          <MenuTreeItemButton backendLanguage={backendLanguage} isLocked={!canAccessMenuItem(node.item)} item={node.item} key={node.id} language={language} onOpenNewItem={onOpenNewItem} onOpenItem={onOpenItem} />
        ))}
      </div>
      ) : null}
    </div>
  );
}

function MenuTreeFolder({
  backendLanguage,
  canAccessMenuItem,
  folder,
  language,
  onOpenNewItem,
  onOpenItem,
  search,
}: {
  backendLanguage: BackendLanguageDictionary;
  canAccessMenuItem: (item: MenuItem) => boolean;
  folder: Extract<MenuTreeNode, { type: "folder" }>;
  language: LanguageCode;
  onOpenNewItem: (item: MenuItem) => void;
  onOpenItem: (item: MenuItem) => void;
  search: string;
}) {
  const [expanded, setExpanded] = useState(Boolean(search.trim()));
  useEffect(() => {
    if (search.trim()) setExpanded(true);
  }, [search]);

  return (
    <div className="relative grid gap-1" role="treeitem" aria-expanded={expanded} aria-selected={expanded}>
      <button
        className="relative grid min-h-9 w-full min-w-0 grid-cols-[auto_auto_minmax(0,1fr)_auto] items-center gap-2 rounded-lg px-2 py-1.5 text-left text-sm font-semibold text-muted-foreground hover:bg-accent hover:text-foreground"
        onClick={() => setExpanded((current) => !current)}
        type="button"
      >
        <span className="absolute -left-2 top-1/2 h-px w-2 bg-border/80" aria-hidden="true" />
        <ChevronDown className={cn("h-3.5 w-3.5 transition-transform", expanded && "rotate-180")} aria-hidden="true" />
        {folder.seedItem ? <MenuRouteIcon item={folder.seedItem} size={15} /> : <Command className="h-4 w-4" />}
        <span className="min-w-0 truncate">{folder.label}</span>
        <Badge variant="outline" className="h-6 px-2 text-[11px]">{folder.children.length}</Badge>
      </button>
      {expanded ? (
        <div className="ml-5 grid gap-1 border-l border-border/80 pl-2" role="group">
          {folder.children.map((item) => (
            <MenuTreeItemButton backendLanguage={backendLanguage} isLocked={!canAccessMenuItem(item)} item={item} key={item.id} language={language} onOpenNewItem={onOpenNewItem} onOpenItem={onOpenItem} nested />
          ))}
        </div>
      ) : null}
    </div>
  );
}

function MenuTreeItemButton({
  backendLanguage,
  isLocked,
  item,
  language,
  nested = false,
  onOpenNewItem,
  onOpenItem,
}: {
  backendLanguage: BackendLanguageDictionary;
  isLocked: boolean;
  item: MenuItem;
  language: LanguageCode;
  nested?: boolean;
  onOpenNewItem: (item: MenuItem) => void;
  onOpenItem: (item: MenuItem) => void;
}) {
  const noPermissionText = backendText(backendLanguage, "no_permission", "No permission");
  const label = menuText(item.label, language, backendLanguage);
  const newTabLabel = language === "th" ? `เปิดแท็บใหม่ ${label}` : `Open new tab ${label}`;

  return (
    <div
      className={cn(
        "relative flex min-h-9 w-full min-w-0 items-center gap-1 rounded-lg px-1.5 py-1 text-sm text-muted-foreground hover:bg-accent hover:text-foreground",
        isLocked && "cursor-not-allowed opacity-70",
        nested && "text-[13px]",
      )}
      key={item.id}
      role="treeitem"
      aria-selected={false}
      aria-disabled={isLocked}
      title={isLocked ? noPermissionText : undefined}
    >
      <span className="absolute -left-2 top-1/2 h-px w-2 bg-border/80" aria-hidden="true" />
      <button
        type="button"
        className="flex h-full min-w-0 flex-1 items-center gap-2 rounded-md px-1.5 text-left disabled:cursor-not-allowed"
        disabled={isLocked}
        onClick={() => onOpenItem(item)}
      >
        <MenuRouteIcon item={item} size={15} />
        <span className="min-w-0 flex-1 truncate">{label}</span>
        {isLocked ? <Lock className="h-3.5 w-3.5 shrink-0 text-muted-foreground" aria-label={noPermissionText} /> : null}
      </button>
      {isLocked ? null : (
        <button
          type="button"
          className="grid h-7 w-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground"
          aria-label={newTabLabel}
          title={newTabLabel}
          onClick={(event) => {
            event.stopPropagation();
            onOpenNewItem(item);
          }}
        >
          <Plus className="h-3.5 w-3.5" />
        </button>
      )}
    </div>
  );
}

function countSectionItems(section: MenuSection): number {
  return section.groups.reduce((total, group) => total + group.items.length, 0);
}

function getVisibleGroups(section: MenuSection, language: LanguageCode, search: string, dictionary: BackendLanguageDictionary): MenuGroup[] {
  const needle = search.trim().toLowerCase();
  if (!needle) return section.groups;
  return section.groups.filter((group) => {
    if (menuText(group.title, language, dictionary).toLowerCase().includes(needle)) return true;
    return getVisibleItems(group.items, language, search, dictionary).length > 0;
  });
}

function getVisibleItems(items: MenuItem[], language: LanguageCode, search: string, dictionary: BackendLanguageDictionary): MenuItem[] {
  const needle = search.trim().toLowerCase();
  if (!needle) return items;
  return items.filter((item) => `${menuText(item.label, language, dictionary)} ${item.route} ${item.id}`.toLowerCase().includes(needle));
}

function getMenuTreeNodes(group: MenuGroup, language: LanguageCode, search: string, dictionary: BackendLanguageDictionary): MenuTreeNode[] {
  const needle = search.trim().toLowerCase();
  const groupTitleMatches = Boolean(needle) && menuText(group.title, language, dictionary).toLowerCase().includes(needle);

  if (group.id !== "company-system") {
    const items = groupTitleMatches ? group.items : getVisibleItems(group.items, language, search, dictionary);
    return items.map((item) => ({ id: item.id, item, type: "item" }));
  }

  const nodes: MenuTreeNode[] = [];
  const insertedFolderIds = new Set<string>();

  const pushSystemFolder = (folder: (typeof systemTreeFolders)[number]) => {
    const folderItems = group.items.filter((item) => folder.itemIds.has(item.id));
    const seedItem = folderItems.find((item) => item.id === folder.seedId) ?? folderItems[0];
    const label = "label" in folder && folder.label ? menuText(folder.label, language, dictionary) : seedItem ? menuText(seedItem.label, language, dictionary) : folder.id;
    const labelMatches = Boolean(needle) && label.toLowerCase().includes(needle);
    const children = !needle || groupTitleMatches || labelMatches ? folderItems : getVisibleItems(folderItems, language, search, dictionary);
    if (!children.length) return;

    nodes.push({
      children,
      id: folder.id,
      label,
      seedItem,
      type: "folder",
    });
  };

  for (const item of group.items) {
    const folder = systemTreeFolders.find((candidate) => candidate.itemIds.has(item.id));
    if (folder) {
      if (!insertedFolderIds.has(folder.id)) {
        pushSystemFolder(folder);
        insertedFolderIds.add(folder.id);
      }
      continue;
    }

    if (groupTitleMatches || !needle || getVisibleItems([item], language, search, dictionary).length) {
      nodes.push({ id: item.id, item, type: "item" });
    }
  }

  return nodes;
}

function SectionIcon({ sectionId }: { sectionId: string }) {
  if (sectionId === "transactions") return <Store className="h-4 w-4" />;
  if (sectionId === "reports") return <LayoutDashboard className="h-4 w-4" />;
  if (sectionId === "settings") return <Settings className="h-4 w-4" />;
  return <Command className="h-4 w-4" />;
}

function OpenTabs({
  activeTabId,
  backendLanguage,
  language,
  onClose,
  onReorder,
  onSelect,
  tabs,
}: {
  activeTabId: string;
  backendLanguage: BackendLanguageDictionary;
  language: LanguageCode;
  onClose: (id: string) => void;
  onReorder: (sourceId: string, targetId: string, side: TabInsertSide) => void;
  onSelect: (id: string) => void;
  tabs: WorkTab[];
}) {
  const [draggingTabId, setDraggingTabId] = useState<string | null>(null);
  const [insertMarker, setInsertMarker] = useState<{ tabId: string; side: TabInsertSide } | null>(null);
  const draggingTabIdRef = useRef<string | null>(null);
  const lastInsertionRef = useRef<string | null>(null);
  const tabRefs = useRef(new Map<string, HTMLDivElement>());
  const previousTabRects = useRef(new Map<string, DOMRect>());

  useLayoutEffect(() => {
    if (!previousTabRects.current.size) return;
    const previousRects = previousTabRects.current;
    previousTabRects.current = new Map();

    requestAnimationFrame(() => {
      tabs.forEach((tab) => {
        const element = tabRefs.current.get(tab.id);
        const previous = previousRects.get(tab.id);
        if (!element || !previous) return;

        const next = element.getBoundingClientRect();
        const deltaX = previous.left - next.left;
        const deltaY = previous.top - next.top;
        if (Math.abs(deltaX) < 1 && Math.abs(deltaY) < 1) return;

        element.animate(
          [
            { transform: `translate(${deltaX}px, ${deltaY}px)` },
            { transform: "translate(0, 0)" },
          ],
          {
            duration: 220,
            easing: "cubic-bezier(0.22, 1, 0.36, 1)",
          },
        );
      });
    });
  }, [tabs]);

  function captureTabRects() {
    previousTabRects.current = new Map(
      tabs
        .map((tab) => {
          const element = tabRefs.current.get(tab.id);
          return element ? [tab.id, element.getBoundingClientRect()] as const : null;
        })
        .filter((item): item is readonly [string, DOMRect] => Boolean(item)),
    );
  }

  function getInsertSide(event: DragEvent<HTMLDivElement>): TabInsertSide {
    const rect = event.currentTarget.getBoundingClientRect();
    return event.clientX > rect.left + rect.width / 2 ? "after" : "before";
  }

  function clearDragState() {
    draggingTabIdRef.current = null;
    lastInsertionRef.current = null;
    setDraggingTabId(null);
    setInsertMarker(null);
  }

  function reorderFromPointer(event: DragEvent<HTMLDivElement>, targetId: string, sourceId?: string | null) {
    const activeSourceId = sourceId ?? draggingTabIdRef.current ?? draggingTabId;
    if (!activeSourceId || activeSourceId === targetId) {
      setInsertMarker(null);
      return;
    }

    const side = getInsertSide(event);
    const insertionKey = `${activeSourceId}:${targetId}:${side}`;
    setInsertMarker((current) => (current?.tabId === targetId && current.side === side ? current : { tabId: targetId, side }));
    if (lastInsertionRef.current === insertionKey) return;

    lastInsertionRef.current = insertionKey;
    captureTabRects();
    onReorder(activeSourceId, targetId, side);
  }

  return (
    <div className="flex max-w-full flex-wrap gap-1 rounded-xl border border-border bg-card p-1 shadow-sm" role="tablist" aria-label={mt(backendLanguage, "openTabs")}>
      {tabs.map((tab) => (
        <div
          aria-grabbed={draggingTabId === tab.id}
          draggable
          key={tab.id}
          ref={(element) => {
            if (element) tabRefs.current.set(tab.id, element);
            else tabRefs.current.delete(tab.id);
          }}
          onDragEnd={clearDragState}
          onDragOver={(event) => {
            event.preventDefault();
            event.dataTransfer.dropEffect = "move";
            reorderFromPointer(event, tab.id);
          }}
          onDragLeave={() => setInsertMarker((current) => (current?.tabId === tab.id ? null : current))}
          onDragStart={(event) => {
            captureTabRects();
            draggingTabIdRef.current = tab.id;
            lastInsertionRef.current = null;
            setDraggingTabId(tab.id);
            event.dataTransfer.effectAllowed = "move";
            event.dataTransfer.setData("text/plain", tab.id);
          }}
          onDrop={(event) => {
            event.preventDefault();
            const sourceId = event.dataTransfer.getData("text/plain") || draggingTabId;
            if (sourceId && sourceId !== tab.id) {
              reorderFromPointer(event, tab.id, sourceId);
            }
            clearDragState();
          }}
          className={cn(
            "group relative flex w-full min-w-0 cursor-grab items-center overflow-hidden rounded-t-md border border-b-0 transition-[background-color,border-color,box-shadow,opacity,transform] duration-200 ease-out active:cursor-grabbing sm:w-auto sm:min-w-28 sm:max-w-44",
            tab.id === activeTabId
              ? "border-primary border-t-2 border-t-primary bg-primary text-primary-foreground shadow-md ring-1 ring-primary/30 font-semibold"
              : "border-border bg-muted/65 text-muted-foreground hover:bg-background/80 hover:text-foreground",
            draggingTabId === tab.id && "scale-[0.98] opacity-60 ring-2 ring-ring/30 shadow-lg",
            insertMarker?.tabId === tab.id && draggingTabId !== tab.id && "translate-y-[-2px] border-primary/40 bg-primary/5 shadow-md",
          )}
        >
          {insertMarker?.tabId === tab.id && draggingTabId !== tab.id ? (
            <span
              className={cn(
                "pointer-events-none absolute bottom-2 top-2 w-1 rounded-full bg-primary shadow-[0_0_0_4px_var(--focus-ring)]",
                insertMarker.side === "after" ? "right-1" : "left-1",
              )}
            />
          ) : null}
          <button
            type="button"
            role="tab"
            aria-selected={tab.id === activeTabId}
            onClick={() => onSelect(tab.id)}
            title={`${tab.item ? menuText(tab.item.label, language, backendLanguage) : mt(backendLanguage, "overviewErp")} ${tab.id === "home" ? mt(backendLanguage, "dashboardRoute") : tab.route}`}
            className="grid min-w-0 flex-1 grid-cols-[auto_minmax(0,1fr)] items-center gap-x-1.5 px-2 py-1 text-left"
          >
            <span
              className={cn(
                "row-span-2 grid h-5 w-5 place-items-center rounded-md",
                tab.id === activeTabId ? "bg-primary-foreground/20 text-primary-foreground" : "bg-background",
              )}
            >
              {tab.item ? <MenuRouteIcon item={tab.item} size={13} /> : <HomeMenuIcon size={13} />}
            </span>
            <b className="flex min-w-0 items-center gap-1 text-[12px] leading-[14px]">
              <span className="truncate">{tab.item ? menuText(tab.item.label, language, backendLanguage) : mt(backendLanguage, "overviewErp")}</span>
            </b>
            <small className={cn("truncate text-[10px] leading-3", tab.id === activeTabId ? "text-primary-foreground/80" : "text-muted-foreground")}>
              {tab.id === "home" ? mt(backendLanguage, "dashboardRoute") : tab.route}
            </small>
          </button>
          {tab.closable ? (
            <button
              type="button"
              className={cn(
                "mr-1 grid h-5 w-5 place-items-center rounded-md opacity-75 transition-[background-color,color,opacity] hover:opacity-100",
                tab.id === activeTabId
                  ? "text-primary-foreground/85 hover:bg-primary-foreground/15 hover:text-primary-foreground"
                  : "hover:bg-background opacity-0 focus-visible:opacity-100 group-hover:opacity-70",
              )}
              onClick={() => onClose(tab.id)}
              aria-label={`${mt(backendLanguage, "closeTab")} ${tab.item ? menuText(tab.item.label, language, backendLanguage) : mt(backendLanguage, "overviewErp")}`}
            >
              <CircleX className="h-3.5 w-3.5" />
            </button>
          ) : null}
        </div>
      ))}
    </div>
  );
}

function DashboardHome() {
  return <div className="min-h-[320px] min-w-0" aria-label="overview" />;
}

function WorkTabPanel({ active, activeTab, backendLanguage, language, onOpenRoute, tabCount }: { active: boolean; activeTab: WorkTab; backendLanguage: BackendLanguageDictionary; language: LanguageCode; onOpenRoute: (route: string, productCode?: string) => void; tabCount: number }) {
  if (activeTab.route === "/currency") {
    return <CurrencyScreen embedded language={language} />;
  }

  if (activeTab.route === "/line-oa") {
    return <LineOaLinkScreen embedded language={language} />;
  }

  if (activeTab.route === "/product") {
    return (
      <ProductScreen
        active={active}
        embedded
        focusRequest={activeTab.productFocusRequest}
        language={language}
      />
    );
  }

  if (activeTab.route === "/productset") {
    return <ProductSetScreen active={active} embedded language={language} />;
  }

  if (activeTab.route === "/productbarcode") {
    return (
      <ProductBarcodeScreen
        embedded
        language={language}
        onOpenLabelPrint={() => onOpenRoute("/productbarcodeshelf")}
        onOpenProduct={(itemCode) => onOpenRoute("/product", itemCode)}
      />
    );
  }

  if (activeTab.route === "/productbarcodeshelf") {
    return <ProductBarcodeShelfScreen embedded language={language} />;
  }

  if (activeTab.route === "/pricehistory") {
    return <ProductPriceHistoryScreen embedded language={language} />;
  }

  if (activeTab.route === "/marketplace/shopee") {
    return <MarketplaceMappingsScreen platform="shopee" embedded language={language} />;
  }

  if (activeTab.route === "/marketplace/lazada") {
    return <MarketplaceMappingsScreen platform="lazada" embedded language={language} />;
  }

  if (activeTab.route === "/marketplace/tiktok") {
    return <MarketplaceMappingsScreen platform="tiktok" embedded language={language} />;
  }

  if (activeTab.route === "/datamodelgraph") {
    return <DataModelGraphScreen embedded language={language} />;
  }

  const systemSettingConfig = getSystemSettingConfig(activeTab.route);
  if (systemSettingConfig) {
    return <SystemSettingsScreen embedded language={language} route={activeTab.route} />;
  }

  return (
    <Card className="min-h-[420px]">
      <CardContent className="grid min-w-0 grid-cols-[minmax(0,1fr)] gap-4 p-5">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="flex min-w-0 gap-4">
            <span className="grid h-14 w-14 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              {activeTab.item ? <MenuRouteIcon item={activeTab.item} size={28} /> : <HomeMenuIcon size={28} />}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-medium text-muted-foreground">{backendText(backendLanguage, "old_flutter_route")}</p>
              <h2 className="truncate text-2xl font-semibold">{activeTab.item ? menuText(activeTab.item.label, language, backendLanguage) : activeTab.title}</h2>
              <code className="mt-1 block truncate rounded-xl bg-muted px-3 py-1 text-sm text-muted-foreground">{activeTab.route}</code>
            </div>
          </div>
          <Badge variant="warning">{backendText(backendLanguage, "migrate_business_screen_pending")}</Badge>
        </div>

        <div className="grid grid-cols-[minmax(0,1fr)] gap-3 md:grid-cols-3">
          <Metric label={backendText(backendLanguage, "open_tabs")} value={tabCount.toLocaleString("th-TH")} />
          <Metric label={backendText(backendLanguage, "status")} value={backendText(backendLanguage, "placeholder")} />
          <Metric label={backendText(backendLanguage, "target")} value={backendText(backendLanguage, "nextjs_screen")} />
        </div>

        <div className="rounded-2xl border border-dashed border-border bg-muted/30 p-5 text-sm leading-6 text-muted-foreground">
          {backendText(backendLanguage, "migration_placeholder_description")}
        </div>
      </CardContent>
    </Card>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-border bg-card p-4 shadow-sm">
      <p className="text-xs font-semibold text-muted-foreground">{label}</p>
      <strong className="mt-1 block truncate text-lg">{value}</strong>
    </div>
  );
}

function DashboardLoading({ backendLanguage }: { backendLanguage: BackendLanguageDictionary }) {
  return (
    <div className="grid grid-cols-[minmax(0,1fr)] gap-3">
      <div className="grid grid-cols-[minmax(0,1fr)] gap-3 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton className="h-28" key={index} />
        ))}
      </div>
      <Skeleton className="h-80" />
      <div className="flex items-center gap-2 rounded-2xl border border-border bg-card p-4 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        {backendText(backendLanguage, "loading_erp_dashboard")}
      </div>
    </div>
  );
}

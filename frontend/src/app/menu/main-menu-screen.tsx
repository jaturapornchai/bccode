"use client";

import { useQuery } from "@tanstack/react-query";
import {
  AlertCircle,
  AlertTriangle,
  Bell,
  CheckCircle2,
  ChevronDown,
  ChevronsUpDown,
  CircleX,
  Command,
  Copy,
  ExternalLink,
  KeyRound,
  LayoutDashboard,
  Lock,
  Loader2,
  LogOut,
  Menu as MenuIcon,
  MessageCircle,
  PanelLeftClose,
  PanelLeftOpen,
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
import { MENU_SECTIONS, flattenMenuItems, menuText, type MenuGroup, type MenuItem, type MenuSection } from "@/lib/menu-data";
import { getFrequentMenuEntries, menuUsageStorageKey, readMenuUsage, recordMenuUsage, type MenuUsageMap } from "@/lib/menu-usage";
import { getSystemSettingConfig } from "@/lib/system-setting-screens";
import { cn } from "@/lib/utils";
import { branchDisplayName, shopDisplayName, type AuthSession, type WorkspaceSession, workspaceStorageKeys } from "@/lib/workspace-models";
import { CurrencyScreen } from "../currency/currency-screen";
import { LanguageDialog } from "../language-dialog";
import { LineOaLinkScreen } from "../line-oa/line-oa-link-screen";
import { ManualLink } from "../manual-link";
import { SystemSettingsScreen } from "../system-settings/system-settings-screen";
import { ThemeToggle } from "../theme-toggle";
import { ZoomControl } from "../zoom-control";
import { HomeMenuIcon, MenuRouteIcon } from "./menu-icon";
import { MenuDataTable } from "./menu-data-table";
import { buildChartData, buildKpis, fetchErpMenuRows, type ErpMenuRow } from "./menu-dashboard-data";
import { MenuKpiChart } from "./menu-kpi-chart";
import { MenuQueryProvider } from "./menu-query-provider";

type WorkTab = {
  id: string;
  title: string;
  route: string;
  item?: MenuItem;
  closable: boolean;
};

type TabInsertSide = "before" | "after";
type LineNotice = { type: "success" | "error" | "info"; text: string } | null;
type PasswordNotice = { type: "success" | "error" | "info"; text: string } | null;
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
    isdefaultpassword?: boolean;
    name?: string;
    username?: string;
  };
  message?: string;
};
type MenuTreeNode =
  | { id: string; item: MenuItem; type: "item" }
  | { children: MenuItem[]; id: string; label: string; seedItem?: MenuItem; type: "folder" };
type SettingRecord = Record<string, unknown>;

const firstTab: WorkTab = { id: "home", title: "ภาพรวม ERP", route: "/menu", closable: false };
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
const accessMenuIds = new Set(["user", "permission-definition", "approval-setting", "permission-link"]);
const branchScopedMenuIds = new Set(["branch", "department", "workday", "holiday"]);
const systemTreeFolders = [
  { id: "access-control", itemIds: accessMenuIds, label: { key: "access_control", th: "การเข้าถึง", en: "Access" }, seedId: "user" },
  { id: "branch-scope", itemIds: branchScopedMenuIds, seedId: "branch" },
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

function isWorkspaceOwner(workspace: WorkspaceSession | null, auth: AuthSession | null): boolean {
  if (!workspace) return false;
  if (workspace.shop.iscreator) return true;
  if (Number(workspace.shop.role) === 2) return true;
  const creator = stringValue(workspace.shop.createdby ?? workspace.shopInfo?.createdby).toLowerCase();
  const identities = [auth?.username, auth?.profile?.email].map((item) => stringValue(item).toLowerCase()).filter(Boolean);
  return Boolean(creator && identities.includes(creator));
}

function workspacePermissionKeys(workspace: WorkspaceSession): string[] {
  return Array.from(new Set([
    stringValue(workspace.branch?.guidfixed),
    stringValue(workspace.branch?.code),
    stringValue(workspace.shop.branchcode),
    "company",
  ].filter(Boolean)));
}

async function fetchAllowedMenuIds(auth: AuthSession, workspace: WorkspaceSession): Promise<Set<string>> {
  const headers = {
    "Content-Type": "application/json",
    "x-bc-backend-url": auth.backendUrl,
    Authorization: `Bearer ${auth.token}`,
  };
  const shopid = encodeURIComponent(workspace.shop.shopid);
  try {
    const [linksResponse, definitionsResponse] = await Promise.all([
      fetch(`/api/system-settings/permission_link?limit=1000&offset=0&shopid=${shopid}`, { headers, cache: "no-store" }),
      fetch(`/api/system-settings/permission_definition?limit=1000&offset=0&shopid=${shopid}`, { headers, cache: "no-store" }),
    ]);
    if (!linksResponse.ok || !definitionsResponse.ok) return new Set();

    const [linksPayload, definitionsPayload] = await Promise.all([linksResponse.json() as Promise<unknown>, definitionsResponse.json() as Promise<unknown>]);
    const userKeys = new Set([auth.username, auth.profile?.email].map(stringValue).filter(Boolean).map((item) => item.toLowerCase()));
    const permissionLink = normalizeSettingRecords(linksPayload).find((record) => userKeys.has(stringValue(record.employeeCode).toLowerCase()));
    const permissionCodes = new Set(stringArray(permissionLink?.permissionCodes));
    if (!permissionCodes.size) return new Set();

    const branchKeys = workspacePermissionKeys(workspace);
    const allowed = new Set<string>();
    for (const definition of normalizeSettingRecords(definitionsPayload)) {
      if (!permissionCodes.has(stringValue(definition.permissionCode))) continue;
      const branches = toRecord(definition.branches);
      for (const branchKey of branchKeys) {
        const branch = toRecord(branches[branchKey]);
        const menus = toRecord(branch.menus);
        for (const [menuId, value] of Object.entries(menus)) {
          if (Boolean(toRecord(value).access)) allowed.add(menuId);
        }
      }
    }
    return allowed;
  } catch {
    return new Set();
  }
}

function toRecord(value: unknown): SettingRecord {
  if (isRecord(value)) return value;
  if (typeof value !== "string" || !value.trim()) return {};
  try {
    const parsed = JSON.parse(value) as unknown;
    return isRecord(parsed) ? parsed : {};
  } catch {
    return {};
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
  const [sidebarHidden, setSidebarHidden] = useState(false);
  const [topChromeHidden, setTopChromeHidden] = useState(false);
  const [lineDialog, setLineDialog] = useState<LineDialogState>(emptyLineDialog);
  const [lineNotice, setLineNotice] = useState<LineNotice>(null);
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const [passwordNotice, setPasswordNotice] = useState<PasswordNotice>(null);
  const [isDefaultPassword, setIsDefaultPassword] = useState(false);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [allowedMenuIds, setAllowedMenuIds] = useState<Set<string>>(new Set());
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
  const defaultPasswordWarningText = backendText(
    backendLanguage,
    "default_password_warning",
    language === "th" ? "รหัสผ่านยังเป็นค่าเริ่มต้น 12345 กรุณาเปลี่ยนรหัสผ่าน" : "Password is still the default 12345. Please change it.",
  );
  const loginIdentity = auth?.profile?.email?.trim() || auth?.username?.trim() || "-";
  const loginText = backendText(backendLanguage, "login", language === "th" ? "เข้าสู่ระบบ" : "Login");
  const isOwner = useMemo(() => isWorkspaceOwner(workspace, auth), [auth, workspace]);
  const canAccessMenuItem = useCallback((item: MenuItem) => isOwner || allowedMenuIds.has(item.id), [allowedMenuIds, isOwner]);
  const allMenuItems = useMemo(() => flattenMenuItems(), []);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? initialLanguage);
    setLanguage(savedLanguage);
    document.documentElement.lang = savedLanguage;

    const authRaw = localStorage.getItem(workspaceStorageKeys.auth);
    const workspaceRaw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!authRaw) {
      router.replace("/");
      return;
    }
    if (!workspaceRaw) {
      router.replace("/workspace");
      return;
    }

    try {
      const auth = JSON.parse(authRaw) as AuthSession;
      setAuth(auth);
      const nextMenuUsageKey = menuUsageStorageKey(auth);
      setMenuUsageKey(nextMenuUsageKey);
      setMenuUsage(readMenuUsage(localStorage, nextMenuUsageKey));
    } catch {
      setAuth(null);
      const nextMenuUsageKey = menuUsageStorageKey(null);
      setMenuUsageKey(nextMenuUsageKey);
      setMenuUsage(readMenuUsage(localStorage, nextMenuUsageKey));
    }

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
    if (!authToken || !authBackendUrl) return;
    let cancelled = false;

    async function loadProfile() {
      try {
        const response = await fetch(`/api/auth/profile?backendUrl=${encodeURIComponent(authBackendUrl)}`, {
          headers: {
            Authorization: `Bearer ${authToken}`,
            "x-bc-backend-url": authBackendUrl,
          },
          cache: "no-store",
        });
        const payload = await response.json() as ProfileResponse;
        if (cancelled || !response.ok || payload.success === false) return;
        setIsDefaultPassword(Boolean(payload.data?.isdefaultpassword));
      } catch {
        if (!cancelled) setIsDefaultPassword(false);
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
    if (isOwner) {
      setAllowedMenuIds(new Set(allMenuItems.map((item) => item.id)));
      return;
    }

    let cancelled = false;
    async function loadMenuPermissions() {
      const nextAllowed = await fetchAllowedMenuIds(currentAuth, currentWorkspace);
      if (!cancelled) setAllowedMenuIds(nextAllowed);
    }

    void loadMenuPermissions();
    return () => {
      cancelled = true;
    };
  }, [allMenuItems, auth, isOwner, workspace]);

  useEffect(() => {
    return () => stopLinePolling();
  }, []);

  const menuQuery = useQuery({
    queryKey: ["erp-menu-rows", language, backendLanguage],
    queryFn: () => fetchErpMenuRows(language, backendLanguage),
  });

  const rows = useMemo(() => menuQuery.data ?? [], [menuQuery.data]);
  const frequentMenuEntries = useMemo(() => getFrequentMenuEntries(allMenuItems, menuUsage, 20), [allMenuItems, menuUsage]);
  const sectionRows = useMemo(() => {
    if (activeSection === "all") return rows;
    const section = MENU_SECTIONS.find((item) => item.id === activeSection);
    if (!section) return rows;
    const label = menuText(section.title, language, backendLanguage);
    return rows.filter((row) => row.module === label);
  }, [activeSection, backendLanguage, language, rows]);
  const kpis = useMemo(() => buildKpis(rows, backendLanguage), [backendLanguage, rows]);
  const chartData = useMemo(() => buildChartData(rows, language, backendLanguage), [backendLanguage, language, rows]);

  function openMenuItem(item: MenuItem) {
    if (!canAccessMenuItem(item)) return;
    const title = menuText(item.label, language, backendLanguage);
    const tabId = `${item.route}::${crypto.randomUUID()}`;
    if (menuUsageKey) {
      setMenuUsage(recordMenuUsage(localStorage, menuUsageKey, item.id));
    }
    setTabs((current) => {
      return [...current, { id: tabId, title, route: item.route, item, closable: true }];
    });
    setActiveTabId(tabId);
  }

  function toggleSection(sectionId: string) {
    setExpandedSections((current) =>
      current.includes(sectionId) ? current.filter((id) => id !== sectionId) : [...current, sectionId],
    );
    setActiveSection(sectionId);
    setActiveTabId(firstTab.id);
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
      const response = await fetch("/api/auth/line/code", { method: "POST" });
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

    setPasswordSaving(true);
    setPasswordNotice(null);
    try {
      const response = await fetch("/api/auth/profile", {
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
      setIsDefaultPassword(false);
      setPasswordNotice({ type: "success", text: backendText(backendLanguage, "saved", language === "th" ? "บันทึกแล้ว" : "Saved.") });
      setPasswordDialogOpen(false);
    } catch (error) {
      setPasswordNotice({ type: "error", text: error instanceof Error ? backendText(backendLanguage, error.message, error.message) : requestFailedText });
    } finally {
      setPasswordSaving(false);
    }
  }

  function logout() {
    localStorage.removeItem(workspaceStorageKeys.auth);
    localStorage.removeItem(workspaceStorageKeys.workspace);
    localStorage.removeItem(workspaceStorageKeys.shopInfo);
    localStorage.removeItem(workspaceStorageKeys.branch);
    router.replace("/");
  }

  function handleContentScroll(event: UIEvent<HTMLDivElement>) {
    const nextScrollTop = event.currentTarget.scrollTop;
    const delta = nextScrollTop - lastContentScrollTopRef.current;

    if (nextScrollTop < 24) {
      setTopChromeHidden(false);
    } else if (delta > 12) {
      setTopChromeHidden(true);
    } else if (delta < -12) {
      setTopChromeHidden(false);
    }

    lastContentScrollTopRef.current = nextScrollTop;
  }

  return (
    <main className="min-h-dvh overflow-x-hidden bg-background text-foreground lg:h-dvh lg:overflow-hidden">
      <div className={cn("grid min-h-dvh min-w-0 grid-cols-[minmax(0,1fr)] lg:h-dvh lg:overflow-hidden", !sidebarHidden && "lg:grid-cols-[280px_minmax(0,1fr)]")}>
        {sidebarHidden ? null : (
        <aside className="max-h-dvh min-w-0 overflow-x-hidden overflow-y-auto overscroll-contain border-b border-border bg-card/80 p-3 lg:sticky lg:top-0 lg:h-dvh lg:border-b-0 lg:border-r">
          <div className="mb-3 flex items-center gap-3 rounded-2xl border border-border bg-background p-3 shadow-sm">
            <div className="grid h-10 w-10 place-items-center rounded-2xl bg-primary text-primary-foreground">
              <MenuIcon className="h-5 w-5" />
            </div>
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold">{workspace ? shopDisplayName(workspace.shop) : "BC Ai Account"}</p>
              <p className="truncate text-xs text-muted-foreground">{workspace?.branch ? branchDisplayName(workspace.branch) : "ERP Workspace"}</p>
            </div>
            <Button type="button" variant="outline" size="icon" className="ml-auto shrink-0" aria-label={mt(backendLanguage, "hideMenu")} title={mt(backendLanguage, "hideMenu")} onClick={() => setSidebarHidden(true)}>
              <PanelLeftClose className="h-4 w-4" />
            </Button>
          </div>

          <label className="relative mb-3 block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="pl-9" placeholder={mt(backendLanguage, "searchMenu")} value={globalSearch} onChange={(event) => setGlobalSearch(event.target.value)} />
          </label>

          <nav aria-label={mt(backendLanguage, "navigation")} className="grid w-full max-w-full gap-2" role="tree">
            <SidebarButton active={activeSection === "all"} count={rows.length} icon={<LayoutDashboard className="h-4 w-4" />} label={mt(backendLanguage, "overview")} onClick={() => setActiveSection("all")} />
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
                  onOpenItem={openMenuItem}
                  onToggleGroup={toggleGroup}
                  onToggle={() => toggleSection(section.id)}
                  search={globalSearch}
                  section={section}
                />
              );
            })}
          </nav>
        </aside>
        )}

        <section
          className="min-w-0 lg:flex lg:h-dvh lg:min-h-0 lg:flex-col lg:overflow-hidden"
          onPointerMove={(event) => {
            if (topChromeHidden && event.clientY < 26) setTopChromeHidden(false);
          }}
        >
          <header
            className="menu-top-chrome sticky top-0 z-30 shrink-0 border-b border-border bg-background/90 px-3 py-2 backdrop-blur"
            data-hidden={topChromeHidden ? "true" : "false"}
          >
            <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
              <div className="flex min-w-0 items-center gap-2">
                {sidebarHidden ? (
                  <Button type="button" variant="outline" size="icon" className="shrink-0" aria-label={mt(backendLanguage, "showMenu")} title={mt(backendLanguage, "showMenu")} onClick={() => setSidebarHidden(false)}>
                    <PanelLeftOpen className="h-4 w-4" />
                  </Button>
                ) : null}
              </div>

              <div className="flex w-full min-w-0 flex-wrap items-center justify-end gap-2 lg:flex-1">
                <label className="relative min-w-52 flex-1 lg:max-w-xs xl:max-w-sm">
                  <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input className="pl-9" placeholder={mt(backendLanguage, "searchMenu")} value={globalSearch} onChange={(event) => setGlobalSearch(event.target.value)} />
                </label>
                <Button variant="outline" size="icon" aria-label={backendText(backendLanguage, "notification")}>
                  <Bell className="h-4 w-4" />
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline">
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
                <div className="w-28">
                  <LanguageDialog language={language} onLanguageChange={setLanguage} />
                </div>
                <ManualLink compact language={language} screen="menu" />
                <ZoomControl dictionary={backendLanguage} language={language} />
                <ThemeToggle language={language} />
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" className="min-w-0">
                      <UserRound className="h-4 w-4" />
                      <span className="hidden max-w-28 truncate sm:inline">{loginIdentity}</span>
                      <ChevronsUpDown className="h-3.5 w-3.5 opacity-60" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-72">
                    <div className="flex min-w-0 items-center gap-2 rounded-lg px-2 py-2 text-sm">
                      <UserRound className="h-4 w-4 shrink-0 text-muted-foreground" />
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-muted-foreground">{loginText}</p>
                        <p className="truncate font-semibold text-foreground">{loginIdentity}</p>
                      </div>
                    </div>
                    <DropdownMenuItem onClick={() => router.push("/settings")}>
                      <Settings className="h-4 w-4" />
                      {backendText(backendLanguage, "settings")}
                    </DropdownMenuItem>
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
          </header>

          <div className={cn(
            "grid min-w-0 grid-cols-[minmax(0,1fr)] gap-3 p-3 lg:min-h-0 lg:flex-1 lg:overflow-hidden",
            isDefaultPassword ? "lg:grid-rows-[auto_auto_minmax(0,1fr)]" : "lg:grid-rows-[auto_minmax(0,1fr)]",
          )}>
            {isDefaultPassword ? (
              <Card className="border-amber-300 bg-amber-50 text-amber-950 shadow-sm dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-100">
                <CardContent className="flex flex-wrap items-center justify-between gap-2 p-3 text-sm font-medium">
                  <span className="flex min-w-0 items-center gap-2">
                    <AlertTriangle className="h-4 w-4 shrink-0" />
                    <span>{defaultPasswordWarningText}</span>
                  </span>
                  <Button type="button" size="sm" variant="outline" onClick={() => setPasswordDialogOpen(true)}>
                    <KeyRound className="h-4 w-4" />
                    {changePasswordText}
                  </Button>
                </CardContent>
              </Card>
            ) : null}
            <div
              className="menu-tabs-chrome min-w-0 overflow-hidden"
              data-hidden={topChromeHidden ? "true" : "false"}
            >
              <OpenTabs tabs={tabs} activeTabId={activeTabId} backendLanguage={backendLanguage} language={language} onSelect={setActiveTabId} onClose={closeTab} onReorder={reorderTabs} />
            </div>

            <div className="min-w-0 lg:min-h-0 lg:overflow-y-auto lg:overscroll-contain lg:pr-1" onScroll={handleContentScroll}>
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
                <div className="grid min-w-0 grid-cols-[minmax(0,1fr)]">
                  {tabs.map((tab) => (
                    <section
                      aria-hidden={tab.id !== activeTabId}
                      className="min-w-0"
                      hidden={tab.id !== activeTabId}
                      key={tab.id}
                      role="tabpanel"
                    >
                      {tab.id === "home" ? (
                        <DashboardHome
                          chartData={chartData}
                          backendLanguage={backendLanguage}
                          globalSearch={globalSearch}
                          kpis={kpis}
                          onGlobalSearchChange={setGlobalSearch}
                          onOpen={openMenuItem}
                          rows={sectionRows}
                        />
                      ) : (
                        <WorkTabPanel activeTab={tab} backendLanguage={backendLanguage} language={language} tabCount={tabs.length} />
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
            {isDefaultPassword ? (
              <div className="message error">
                <AlertCircle size={18} />
                <span>{defaultPasswordWarningText}</span>
              </div>
            ) : null}
            {passwordNotice ? (
              <div className={`message ${passwordNotice.type === "success" ? "success" : passwordNotice.type === "error" ? "error" : "info"}`}>
                {passwordNotice.type === "success" ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
                <span>{passwordNotice.text}</span>
              </div>
            ) : null}
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
                <span>{lineNotice?.text ?? lineLinkSuccessText}</span>
              </div>
            ) : lineDialog.error ? (
              <div className="message error">
                <AlertCircle size={18} />
                <span>{lineDialog.error}</span>
              </div>
            ) : lineDialog.expired ? (
              <div className="message error">
                <AlertCircle size={18} />
                <span>{lineNotice?.text ?? t(language, "lineLoginExpired")}</span>
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

            {lineNotice && !lineDialog.success && !lineDialog.expired && !lineDialog.error ? (
              <div className={`message ${lineNotice.type === "success" ? "success" : lineNotice.type === "error" ? "error" : "info"}`}>
                {lineNotice.type === "success" ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
                <span>{lineNotice.text}</span>
              </div>
            ) : null}

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
        "flex w-full min-w-0 items-center justify-between gap-3 rounded-2xl border border-transparent px-3 py-2 text-left text-sm font-medium transition-colors",
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
          "flex w-full min-w-0 items-center justify-between gap-3 rounded-2xl border px-3 py-2 text-left text-sm font-medium transition-colors",
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
                  onOpenItem={onOpenItem}
                  onToggle={() => onToggleGroup(groupKey)}
                  search={search}
                />
              );
            })
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
          <MenuTreeFolder backendLanguage={backendLanguage} canAccessMenuItem={canAccessMenuItem} folder={node} key={node.id} language={language} onOpenItem={onOpenItem} search={search} />
        ) : (
          <MenuTreeItemButton backendLanguage={backendLanguage} isLocked={!canAccessMenuItem(node.item)} item={node.item} key={node.id} language={language} onOpenItem={onOpenItem} />
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
  onOpenItem,
  search,
}: {
  backendLanguage: BackendLanguageDictionary;
  canAccessMenuItem: (item: MenuItem) => boolean;
  folder: Extract<MenuTreeNode, { type: "folder" }>;
  language: LanguageCode;
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
            <MenuTreeItemButton backendLanguage={backendLanguage} isLocked={!canAccessMenuItem(item)} item={item} key={item.id} language={language} onOpenItem={onOpenItem} nested />
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
  onOpenItem,
}: {
  backendLanguage: BackendLanguageDictionary;
  isLocked: boolean;
  item: MenuItem;
  language: LanguageCode;
  nested?: boolean;
  onOpenItem: (item: MenuItem) => void;
}) {
  const noPermissionText = backendText(backendLanguage, "no_permission", "No permission");
  return (
    <button
      className={cn(
        "relative grid min-h-9 w-full min-w-0 grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 rounded-lg px-2 py-1.5 text-left text-sm text-muted-foreground hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-70",
        nested && "text-[13px]",
      )}
      disabled={isLocked}
      key={item.id}
      onClick={() => onOpenItem(item)}
      role="treeitem"
      aria-selected={false}
      aria-disabled={isLocked}
      title={isLocked ? noPermissionText : undefined}
      type="button"
    >
      <span className="absolute -left-2 top-1/2 h-px w-2 bg-border/80" aria-hidden="true" />
      <MenuRouteIcon item={item} size={15} />
      <span className="min-w-0 truncate">{menuText(item.label, language, backendLanguage)}</span>
      {isLocked ? <Lock className="h-3.5 w-3.5 shrink-0 text-muted-foreground" aria-label={noPermissionText} /> : null}
    </button>
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
    <div className="flex max-w-full flex-wrap gap-2 rounded-2xl border border-border bg-card p-2 shadow-sm" role="tablist" aria-label={mt(backendLanguage, "openTabs")}>
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
            "relative flex w-full min-w-0 cursor-grab items-center overflow-hidden rounded-2xl border transition-[background-color,border-color,box-shadow,opacity,transform] duration-200 ease-out active:cursor-grabbing sm:w-auto sm:min-w-44 sm:max-w-64",
            tab.id === activeTabId ? "border-primary/30 bg-primary/10 text-primary" : "border-border bg-background text-foreground",
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
            className="grid min-w-0 flex-1 grid-cols-[auto_minmax(0,1fr)] items-center gap-x-2 px-3 py-2 text-left"
          >
            <span className="row-span-2 grid h-7 w-7 place-items-center rounded-xl bg-background">
              {tab.item ? <MenuRouteIcon item={tab.item} size={15} /> : <HomeMenuIcon size={15} />}
            </span>
            <b className="truncate text-sm">{tab.item ? menuText(tab.item.label, language, backendLanguage) : mt(backendLanguage, "overviewErp")}</b>
            <small className="truncate text-xs text-muted-foreground">{tab.id === "home" ? mt(backendLanguage, "dashboardRoute") : tab.route}</small>
          </button>
          {tab.closable ? (
            <button
              type="button"
              className="mr-2 grid h-7 w-7 place-items-center rounded-xl hover:bg-background"
              onClick={() => onClose(tab.id)}
              aria-label={`${mt(backendLanguage, "closeTab")} ${tab.item ? menuText(tab.item.label, language, backendLanguage) : mt(backendLanguage, "overviewErp")}`}
            >
              <CircleX className="h-4 w-4" />
            </button>
          ) : null}
        </div>
      ))}
    </div>
  );
}

function DashboardHome({
  backendLanguage,
  chartData,
  globalSearch,
  kpis,
  onGlobalSearchChange,
  onOpen,
  rows,
}: {
  backendLanguage: BackendLanguageDictionary;
  chartData: ReturnType<typeof buildChartData>;
  globalSearch: string;
  kpis: ReturnType<typeof buildKpis>;
  onGlobalSearchChange: (value: string) => void;
  onOpen: (item: MenuItem) => void;
  rows: ErpMenuRow[];
}) {
  return (
    <div className="grid min-w-0 grid-cols-[minmax(0,1fr)] gap-3">
      <section className="grid grid-cols-[minmax(0,1fr)] gap-3 md:grid-cols-2 xl:grid-cols-4" aria-label="KPI">
        {kpis.map((item) => (
          <Card key={item.label} className="shadow-sm">
            <CardHeader className="pb-2">
              <CardDescription>{item.label}</CardDescription>
              <CardTitle className="text-2xl">{item.value}</CardTitle>
            </CardHeader>
            <CardContent>
              <Badge variant={item.tone === "success" ? "success" : item.tone === "warning" ? "warning" : "secondary"}>{item.change}</Badge>
            </CardContent>
          </Card>
        ))}
      </section>

      <section className="grid grid-cols-[minmax(0,1fr)] gap-3">
        <MenuKpiChart data={chartData} dictionary={backendLanguage} />
      </section>

      <Card>
        <CardHeader>
          <CardTitle>{backendText(backendLanguage, "erp_data_table")}</CardTitle>
          <CardDescription>{backendText(backendLanguage, "erp_data_table_description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <MenuDataTable data={rows} dictionary={backendLanguage} globalSearch={globalSearch} onGlobalSearchChange={onGlobalSearchChange} onOpen={onOpen} />
        </CardContent>
      </Card>
    </div>
  );
}

function WorkTabPanel({ activeTab, backendLanguage, language, tabCount }: { activeTab: WorkTab; backendLanguage: BackendLanguageDictionary; language: LanguageCode; tabCount: number }) {
  if (activeTab.route === "/currency") {
    return <CurrencyScreen embedded language={language} />;
  }

  if (activeTab.route === "/line-oa") {
    return <LineOaLinkScreen embedded language={language} />;
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

"use client";

import {
  AlertCircle,
  ArrowLeft,
  BookOpen,
  CheckCircle2,
  Cloud,
  Copy,
  Database,
  Download,
  Eye,
  EyeOff,
  HardDrive,
  KeyRound,
  Loader2,
  LockKeyhole,
  LogIn,
  LogOut,
  Network,
  RefreshCcw,
  Save,
  Server,
  Settings as SettingsIcon,
  Table2,
  Trash2,
  Upload,
  Wifi,
  X,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { persistLanguagePreferenceCookies } from "@/lib/backend-language-preload";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import { catalogText } from "@/lib/catalog-text";
import {
  AI_PROVIDERS,
  SETUP_CATEGORY_DEFS,
  buildConnectionPayload,
  configMapToExportJson,
  createDefaultConfigMap,
  fieldLabels,
  getCategoryDef,
  getFieldControl,
  importConfigJson,
  mergeBackendConfig,
  normalizeBooleanValue,
  normalizeSetupBackendUrl,
  serializeConfig,
  updateConfigItem,
  validateConfig,
  type ConfigItem,
  type ConfigMap,
  type FieldOption,
  type TestResult,
} from "@/lib/setup-config";
import { AppHeaderControls } from "../app-header-controls";
import { ManualLink } from "../manual-link";

type ConnectionState = "idle" | "testing" | "success" | "error";
type MessageState = "idle" | "success" | "error";
type LoadingAction = "idle" | "verify" | "load" | "save" | "test-all" | "change-password";
type DialogMode = "import" | "change-password" | null;

type SetupResponse = {
  success?: boolean;
  message?: string;
  data?: unknown;
  latencyms?: unknown;
  errorcode?: unknown;
  database?: unknown;
};

const DEFAULT_BACKEND_URL =
  process.env.NEXT_PUBLIC_DEFAULT_BACKEND_URL ?? "http://192.168.2.202:8888/goapi";
const storageKeys = {
  backendUrl: "backend_url",
  backendUrlHistory: "backend_url_history",
  language: "user_language",
};

// 2026-09-16: the Thai/English literals below are the fallback — these keys
// pull the same words out of languages.tsv for all twelve languages.
const settingsTextKeys: Record<string, string> = {
  "group.databases": "set_databases",
  "testedSummary": "set_tested_0_1_passed",
  "reachable": "set_reachable_0",
  "group.storage": "set_image_file_storage",
  "storageShort": "set_storage",
  "serverConnection": "set_server_connection",
  "serverConnectionHint": "set_enter_a_single_host_and",
  "autofillTitle": "set_fill_this_host_into_every",
  "autofillAll": "set_auto_fill_all",
  "storageHint": "set_configure_s3_compatible_storage_e",
  "storageDescription": "set_s3_compatible_e_g_minio",
  "s3EndpointRequired": "set_please_fill_s3_endpoint_first",
  "testing": "testing",
  "unreachable": "set_unreachable",
  "fields": "set_fields",
  "testStorage": "set_test_storage",
};
export function SettingsScreen() {
  const [backendUrl, setBackendUrl] = useState(DEFAULT_BACKEND_URL);
  const [urlHistory, setUrlHistory] = useState<string[]>([]);
  const [language, setLanguage] = useState<LanguageCode>("th");
  const backendLanguage = useBackendLanguage(language, backendUrl || undefined);
  const st = useCallback(
    (id: string, bilingual: { th: string; en: string }) =>
      catalogText(settingsTextKeys, id, bilingual, language, backendLanguage),
    [backendLanguage, language],
  );
  // Strings that never had an English twin: the key is the identity and the
  // Thai is only the offline fallback.
  const tr = useCallback(
    (key: string, fallback: string) => backendText(backendLanguage, key, fallback),
    [backendLanguage],
  );
  const testBadgeLabel = useCallback(
    (result: TestResult) =>
      result.status === "success"
        ? tr("passed", "ผ่าน")
        : result.status === "testing"
          ? tr("set_testing", "กำลังทดสอบ")
          : tr("set_not_passed", "ไม่ผ่าน"),
    [tr],
  );
  const [connectionState, setConnectionState] = useState<ConnectionState>("idle");
  const [messageState, setMessageState] = useState<MessageState>("idle");
  const [message, setMessage] = useState("");
  const [setupPassword, setSetupPassword] = useState("");
  const [showSetupPassword, setShowSetupPassword] = useState(false);
  const [authenticated, setAuthenticated] = useState(false);
  const [loadingAction, setLoadingAction] = useState<LoadingAction>("idle");
  const [testingCategory, setTestingCategory] = useState("");
  const [configMap, setConfigMap] = useState<ConfigMap>(() => createDefaultConfigMap());
  const [testResults, setTestResults] = useState<Record<string, TestResult>>({});
  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [importText, setImportText] = useState("");
  const [currentSetupPassword, setCurrentSetupPassword] = useState("");
  const [newSetupPassword, setNewSetupPassword] = useState("");
  const [serverHost, setServerHost] = useState("");
  const { confirm, confirmationDialog } = useConfirmDialog();
  const router = useRouter();
  // After "บันทึก Config" the backend reloads/reconnects with the new config; show a popup
  // while that happens, then return to the login screen so the user re-enters with fresh config.
  const [reloadingConfig, setReloadingConfig] = useState(false);

  // Section groups for the redesigned layout. Each group renders its own heading
  // and contains a list of config categories. Pure PostgreSQL is the single database.
  const storageIntegrationKeys = new Set([
    "s3endpoint",
    "s3publicendpoint",
    "s3accesskeyid",
    "s3secretaccesskey",
    "s3bucketname",
  ]);

  const sectionGroups: { id: string; titleTh: string; titleEn: string; categories: string[] }[] = [
    { id: "databases", titleTh: "ฐานข้อมูล", titleEn: "Database", categories: ["postgresql"] },
    { id: "storage", titleTh: "ที่เก็บรูปและไฟล์", titleEn: "Image & File Storage", categories: [] },
  ];

  // Categories that require connection testing before saving config
  const testableCategories = useMemo(() => {
    const cats: { id: string; title: string }[] = [];
    for (const def of SETUP_CATEGORY_DEFS) {
      if (def.testType && (configMap[def.id]?.length ?? 0) > 0) {
        cats.push({ id: def.id, title: def.title });
      }
    }
    const hasStorage = (configMap["integrations"] ?? []).some((item) => storageIntegrationKeys.has(item.key));
    if (hasStorage) {
      cats.push({
        id: "storage",
        title: st("group.storage", { th: "ที่เก็บรูปและไฟล์", en: "Image & File Storage" }),
      });
    }
    return cats;
  }, [configMap, storageIntegrationKeys, st]);

  const pendingCategories = useMemo(() => {
    return testableCategories.filter((cat) => testResults[cat.id]?.status !== "success");
  }, [testableCategories, testResults]);

  const allTestsPassed = testableCategories.length > 0 && pendingCategories.length === 0;

  // Parse host from a backend URL like http://192.168.2.202:8888/goapi -> 192.168.2.202
  function parseHostFromUrl(url: string): string {
    try {
      const withProtocol = /^https?:\/\//i.test(url) ? url : `http://${url}`;
      return new URL(withProtocol).hostname || "";
    } catch {
      return "";
    }
  }

  // Default Docker on-prem values — pure PostgreSQL architecture.
  const DOCKER_DEFAULTS: Record<string, Record<string, string>> = {
    postgresql: {
      host: "postgres",
      port: "5432",
      user: "postgres",
      password: "***",
      dbname: "bcai_projection",
      sslmode: "disable",
      timezone: "Asia/Bangkok",
      loggerlevel: "Info",
    },
    integrations: {
      s3endpoint: "http://minio:9000",
      s3publicendpoint: "",
      s3accesskeyid: "admin",
      s3secretaccesskey: "***",
      s3bucketname: "bcai-media",
    },
  };

  // Auto-fill: populate PostgreSQL and storage fields with Docker defaults.
  function handleAutoFillHost() {
    if (!serverHost.trim()) return;
    const publicHost = serverHost.trim();
    setConfigMap((current) => {
      const next: ConfigMap = {};
      for (const [category, items] of Object.entries(current)) {
        const defaults = DOCKER_DEFAULTS[category] ?? {};
        next[category] = items.map((item) => {
          // Host fields: use Docker container name (fixed), not the public host.
          if (item.key === "host" && defaults.host) {
            return { ...item, value: defaults.host };
          }
          // S3 public endpoint: leave empty for standard internal Docker setup.
          if (item.key === "s3publicendpoint") {
            return { ...item, value: defaults.s3publicendpoint ?? "" };
          }
          // All other fields: use Docker default if available.
          if (defaults[item.key] !== undefined) {
            return { ...item, value: defaults[item.key] };
          }
          return item;
        });
      }
      return next;
    });
  }


  // null = empty (no indicator yet), true/false = valid/invalid backend URL.
  // normalizeSetupBackendUrl is lenient (coerces almost anything into a URL), so
  // also reject input that contains whitespace to catch obvious typos.
  const backendUrlValid = useMemo(() => {
    const value = backendUrl.trim();
    if (value.length === 0) return null;
    if (/\s/.test(value)) return false;
    return Boolean(normalizeSetupBackendUrl(value));
  }, [backendUrl]);
  const canSaveBackend = backendUrlValid === true;
  const isBusy = loadingAction !== "idle";
  const orderedCategories = useMemo(() => {
    const known = SETUP_CATEGORY_DEFS.map((item) => item.id).filter((category) => configMap[category]?.length);
    const unknown = Object.keys(configMap).filter((category) => !known.includes(category));
    return [...known, ...unknown];
  }, [configMap]);

  useEffect(() => {
    const savedBackendUrl = localStorage.getItem(storageKeys.backendUrl);
    const savedLanguage = normalizeLanguage(localStorage.getItem(storageKeys.language) ?? "th");
    const history = readUrlHistory();

    setLanguage(savedLanguage);
    setUrlHistory(history);
    if (savedBackendUrl) {
      setBackendUrl(savedBackendUrl);
      setServerHost(parseHostFromUrl(savedBackendUrl));
    } else {
      void loadConfigUrl();
    }
  }, []);

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem(storageKeys.language, language);
    persistLanguagePreferenceCookies(language, backendUrl);
  }, [backendUrl, language]);

  async function loadConfigUrl() {
    try {
      const response = await fetch("/config.json", { cache: "no-store" });
      if (response.ok) {
        const data = (await response.json()) as { goapi_url?: string };
        if (data.goapi_url && !data.goapi_url.includes("localhost")) {
          setBackendUrl(data.goapi_url);
          return;
        }
      }

      // Fallback อัจฉริยะ: ชี้ไปที่โดเมนปัจจุบันทันทีเมื่อเปิดจาก VPS / Production
      if (typeof window !== "undefined" && !window.location.hostname.includes("localhost") && !window.location.hostname.includes("127.0.0.1")) {
        setBackendUrl(`${window.location.origin}/goapi`);
      } else {
        setBackendUrl(DEFAULT_BACKEND_URL);
      }
    } catch {
      setBackendUrl(DEFAULT_BACKEND_URL);
    }
  }


  function handleSaveBackendUrl() {
    const normalized = normalizeSetupBackendUrl(backendUrl);
    if (!normalized) {
      setStatus(t(language, "enterBackendUrl"), "error");
      return;
    }

    persistBackendUrl(normalized);
    setBackendUrl(normalized);
    setConnectionState("idle");
    setStatus(t(language, "saved"), "success");
  }

  async function handleBackendConnectionTest() {
    const normalized = normalizeSetupBackendUrl(backendUrl);
    if (!normalized) {
      setStatus(t(language, "enterBackendUrl"), "error");
      return;
    }

    setConnectionState("testing");
    setMessage("");
    setMessageState("idle");

    try {
      const response = await fetch("/api/backend/check", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl: normalized }),
      });
      const data = (await response.json()) as { success?: boolean; message?: string };
      if (!response.ok || !data.success) throw new Error(data.message ?? t(language, "connectionFailed"));
      setBackendUrl(normalized);
      persistBackendUrl(normalized);
      setConnectionState("success");
      setStatus(t(language, "connectionSuccess"), "success");
    } catch (error) {
      setConnectionState("error");
      setStatus(error instanceof Error ? error.message : t(language, "connectionFailed"), "error");
    }
  }

  async function handleVerifySetup() {
    const normalized = normalizeSetupBackendUrl(backendUrl);
    if (!normalized) {
      setStatus(t(language, "enterBackendUrl"), "error");
      return;
    }
    if (!setupPassword) {
      setStatus(tr("set_enter_setup_password", "กรุณาระบุรหัสผ่าน Setup"), "error");
      return;
    }

    setLoadingAction("verify");
    try {
      const data = await callSetup("verify-password", { password: setupPassword }, normalized);
      persistBackendUrl(normalized);
      setBackendUrl(normalized);
      setAuthenticated(true);
      setStatus(data.message ?? tr("set_password_verified", "ยืนยันรหัสผ่านสำเร็จ"), "success");
      await loadSetupConfig(setupPassword, normalized);
    } catch (error) {
      setAuthenticated(false);
      setStatus(error instanceof Error ? error.message : tr("set_setup_login_failed", "เข้าสู่ระบบ Setup ไม่สำเร็จ"), "error");
    } finally {
      setLoadingAction("idle");
    }
  }

  async function loadSetupConfig(password = setupPassword, backendUrlOverride = backendUrl) {
    if (!password) return;

    setLoadingAction("load");
    try {
      const data = await callSetup("config/get", { password }, backendUrlOverride);
      const merged = mergeBackendConfig(data.data);
      setConfigMap(merged);
      setAuthenticated(true);
      setStatus(tr("set_config_loaded", "โหลด config สำเร็จ"), "success");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : tr("set_config_load_failed", "โหลด config ล้มเหลว"), "error");
    } finally {
      setLoadingAction("idle");
    }
  }


  async function handleSaveConfig() {
    // Guard: ต้องทดสอบการเชื่อมต่อให้ผ่านครบทุกรายการก่อนบันทึก
    if (!allTestsPassed) {
      const pendingNames = pendingCategories.map((c) => c.title).join(", ");
      setStatus(
        st("mustPassAllTestsFirst", {
          th: `กรุณาทดสอบการเชื่อมต่อให้ผ่านครบทุกรายการก่อนบันทึก Config (ยังไม่ได้ทดสอบหรือยังไม่ผ่าน: ${pendingNames})`,
          en: `Please test and pass all connections before saving config (pending: ${pendingNames})`,
        }),
        "error",
      );
      return;
    }

    const errors = validateConfig(configMap);
    if (errors.length > 0) {
      setStatus(tr("set_config_check_failed", "ตรวจสอบ config ไม่ผ่าน: {0}").replace("{0}", errors.slice(0, 3).join(", ")), "error");
      return;
    }

    const confirmed = await confirm({
      title: tr("set_confirm_save_config", "ยืนยันบันทึก Config"),
      description: tr("set_save_config_overwrites_bootstrap", "บันทึก config ไปที่ backend และเขียนทับ bootstrap.json"),
      details: tr("set_check_values_before_confirm", "ตรวจสอบ Backend URL และค่าที่แก้ไขให้เรียบร้อยก่อนยืนยัน เพราะค่าเหล่านี้มีผลกับการเชื่อมต่อระบบ"),
      confirmLabel: tr("set_save_config", "บันทึก Config"),
      cancelLabel: tr("cancel", "ยกเลิก"),
      tone: "warning",
    });
    if (!confirmed) {
      return;
    }

    setLoadingAction("save");
    try {
      const data = await callSetup("config/save", {
        password: setupPassword,
        configs: serializeConfig(configMap),
      });
      setStatus(data.message ?? tr("set_config_saved", "บันทึก config สำเร็จ"), "success");
      // Backend reloads + reconnects with the new config (async). Show a popup while it
      // settles, then send the user back to the login screen to re-enter with fresh config.
      setReloadingConfig(true);
      await new Promise((resolve) => setTimeout(resolve, 2200));
      router.push("/");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : tr("set_config_save_failed", "บันทึก config ล้มเหลว"), "error");
      setReloadingConfig(false);
    } finally {
      setLoadingAction("idle");
    }
  }

  async function handleTestConnection(category: string): Promise<boolean> {
    const items = configMap[category] ?? [];
    setTestingCategory(category);
    setTestResults((current) => ({
      ...current,
      [category]: { status: "testing", message: "กำลังทดสอบ..." },
    }));

    try {
      const payload = buildConnectionPayload(category, items, setupPassword);
      const data = await callSetup("test-connection", payload);
      const result = responseToTestResult(data, { success: tr("success", "สำเร็จ"), failed: tr("failed", "ไม่สำเร็จ") });
      setTestResults((current) => ({ ...current, [category]: result }));
      return result.status === "success";
    } catch (error) {
      setTestResults((current) => ({
        ...current,
        [category]: {
          status: "failed",
          message: error instanceof Error ? error.message : tr("set_test_failed", "ทดสอบไม่สำเร็จ"),
        },
      }));
      return false;
    } finally {
      setTestingCategory("");
    }
  }

  async function handleTestAllConnections() {
    const categories = SETUP_CATEGORY_DEFS.filter((category) => category.testType && configMap[category.id]?.length).map((category) => category.id);
    if (categories.length === 0) return;

    setLoadingAction("test-all");
    // Run DB category tests in parallel, then run storage test, then summarize.
    const dbResults = await Promise.all(categories.map((category) => handleTestConnection(category)));
    // Run storage test (S3/MinIO) separately since it uses a different probe path.
    const storageResult = await handleTestStorageConnection();

    // Build detailed summary line: "PostgreSQL 12ms ✓, ..."
    const summaryParts: string[] = [];
    const labelMap: Record<string, string> = {
      postgresql: "PostgreSQL",
    };
    for (let i = 0; i < categories.length; i++) {
      const category = categories[i];
      const passed = dbResults[i];
      const tr = testResults[category];
      const latency = tr?.latencyMs != null ? ` ${tr.latencyMs}ms` : "";
      const mark = passed ? "✓" : "✗";
      summaryParts.push(`${labelMap[category] ?? category}${latency} ${mark}`);
    }
    if (storageResult) {
      const sLatency = storageResult.latencyMs != null ? ` ${storageResult.latencyMs}ms` : "";
      const sMark = storageResult.status === "success" ? "✓" : "✗";
      summaryParts.push(`${st("storageShort", { th: "ที่เก็บรูป", en: "Storage" })}${sLatency} ${sMark}`);
    }

    const dbSuccess = dbResults.filter(Boolean).length;
    const storageSuccess = storageResult?.status === "success" ? 1 : 0;
    const totalSuccess = dbSuccess + storageSuccess;
    const total = categories.length + (storageResult ? 1 : 0);
    const headline =
      st("testedSummary", { th: "ทดสอบแล้ว {0}/{1} สำเร็จ", en: "Tested {0}/{1} passed" })
        .replace("{0}", String(totalSuccess))
        .replace("{1}", String(total));
    const detailLine = summaryParts.join(" · ");

    setStatus(`${headline} — ${detailLine}`, totalSuccess === total ? "success" : "error");
    setLoadingAction("idle");
  }

  async function handleExportConfig() {
    try {
      await navigator.clipboard.writeText(configMapToExportJson(configMap));
      setStatus(tr("set_config_json_copied", "คัดลอก config JSON แล้ว"), "success");
    } catch {
      setStatus(t(language, "copyFailed"), "error");
    }
  }

  function handleImportConfig(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setConfigMap((current) => importConfigJson(current, importText));
      setDialogMode(null);
      setImportText("");
      setStatus(tr("set_config_imported_not_saved", "Import config แล้ว ยังไม่ได้บันทึกลง backend"), "success");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : tr("set_config_import_failed", "Import config ล้มเหลว"), "error");
    }
  }

  async function handleChangePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!currentSetupPassword || !newSetupPassword) {
      setStatus(tr("set_enter_old_and_new_password", "กรุณาระบุรหัสผ่านเดิมและรหัสผ่านใหม่"), "error");
      return;
    }

    setLoadingAction("change-password");
    try {
      const data = await callSetup("change-password", {
        currentpassword: currentSetupPassword,
        newpassword: newSetupPassword,
      });
      setSetupPassword(newSetupPassword);
      setDialogMode(null);
      setCurrentSetupPassword("");
      setNewSetupPassword("");
      setStatus(data.message ?? tr("set_password_changed", "เปลี่ยนรหัสผ่านสำเร็จ"), "success");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : tr("set_password_change_failed", "เปลี่ยนรหัสผ่านล้มเหลว"), "error");
    } finally {
      setLoadingAction("idle");
    }
  }

  async function handleCopyBackendUrl() {
    if (!backendUrl.trim()) {
      setStatus(t(language, "enterBackendUrl"), "error");
      return;
    }

    try {
      await navigator.clipboard.writeText(backendUrl.trim());
      setStatus(t(language, "copied"), "success");
    } catch {
      setStatus(t(language, "copyFailed"), "error");
    }
  }

  function handleRemoveHistory(url: string) {
    const nextHistory = urlHistory.filter((item) => item !== url);
    localStorage.setItem(storageKeys.backendUrlHistory, JSON.stringify(nextHistory));
    setUrlHistory(nextHistory);
  }

  function handleClearHistory() {
    localStorage.removeItem(storageKeys.backendUrlHistory);
    setUrlHistory([]);
  }

  function updateItem(category: string, key: string, value: string) {
    setConfigMap((current) => updateConfigItem(current, category, key, value));
  }

  function persistBackendUrl(nextBackendUrl: string) {
    localStorage.setItem(storageKeys.backendUrl, nextBackendUrl);
    const nextHistory = [nextBackendUrl, ...urlHistory.filter((url) => url !== nextBackendUrl)].slice(0, 10);
    localStorage.setItem(storageKeys.backendUrlHistory, JSON.stringify(nextHistory));
    setUrlHistory(nextHistory);
  }

  function logoutSetup() {
    setAuthenticated(false);
    setSetupPassword("");
    setShowSetupPassword(false);
    setConfigMap(createDefaultConfigMap());
    setTestResults({});
    setStatus(tr("set_signed_out_of_setup", "ออกจากระบบ Setup แล้ว"), "success");
  }

  function setStatus(nextMessage: string, nextState: Exclude<MessageState, "idle">) {
    setMessage(nextMessage);
    setMessageState(nextState);
  }

  async function callSetup(path: string, payload: Record<string, unknown>, backendUrlOverride = backendUrl): Promise<SetupResponse> {
    const normalized = normalizeSetupBackendUrl(backendUrlOverride);
    if (!normalized) throw new Error(t(language, "enterBackendUrl"));

    const response = await fetch(`/api/setup/${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backendUrl: normalized, ...payload }),
    });
    const data = (await response.json().catch(() => ({}))) as SetupResponse;
    if (!response.ok || data.success === false) {
      throw new Error(data.message ?? tr("set_setup_backend_bad_response", "Setup backend ตอบกลับผิดปกติ"));
    }
    return data;
  }



  return (
    <main className="settings-shell">
      <section className="settings-layout" aria-label={t(language, "settingsTitle")}>
        <header className="settings-header settings-hero">
          <Link className="icon-button settings-back-link" href="/" aria-label={t(language, "backToLogin")} title={t(language, "backToLogin")}>
            <ArrowLeft aria-hidden="true" size={19} />
          </Link>

          <div className="settings-title-block">
            <p className="eyebrow">BC Ai Account</p>
            <h1>{tr("system_settings_center", "ศูนย์ตั้งค่าระบบ")}</h1>
            <p>{tr("set_manage_everything_one_screen", "จัดการ Backend URL, Setup Config, database connections, storage, AI providers, ธีม และภาษาในจอเดียว")}</p>
          </div>

          <div className="header-actions settings-header-actions">
            <ManualLink compact language={language} screen="settings" />
            <AppHeaderControls language={language} onLanguageChange={setLanguage} showSettings={false} />
          </div>
        </header>

        <section className="settings-overview" aria-label={tr("set_system_status", "สถานะระบบ")}>
          <div className="overview-tile">
            <span>Backend</span>
            <strong>{normalizeSetupBackendUrl(backendUrl) || tr("not_configured", "ยังไม่ได้ตั้งค่า")}</strong>
          </div>
          <div className="overview-tile">
            <span>Setup</span>
            <strong>{authenticated ? tr("set_verified", "ยืนยันแล้ว") : tr("set_not_signed_in_to_setup", "ยังไม่เข้าสู่ Setup")}</strong>
          </div>
          <div className="overview-tile">
            <span>Config Sections</span>
            <strong>{authenticated ? tr("set_group_count", "{0} กลุ่ม").replace("{0}", String(orderedCategories.length)) : tr("set_locked_until_password_verified", "ล็อกไว้ก่อนยืนยันรหัส")}</strong>
          </div>
        </section>

        <div className="settings-grid settings-dashboard">
          <section className="settings-card settings-card-main connection-card" aria-label="Backend Connection">
            <div className="settings-card-title">
              <span className="settings-card-icon">
                <Server aria-hidden="true" size={19} />
              </span>
              <div>
                <p className="eyebrow">{t(language, "api")}</p>
                <h2>{tr("connect_backend", "เชื่อมต่อ Backend")}</h2>
                <p className="config-description">{tr("set_check_url_then_verify", "เริ่มจากตรวจ URL แล้วค่อยยืนยันรหัส Setup เพื่อโหลด config จาก backend")}</p>
              </div>
            </div>

            <div className="connection-fields">
              <label className="field-group">
                <span>{t(language, "backendUrl")}</span>
                <div className="input-shell input-shell-large">
                  <Server aria-hidden="true" size={18} />
                  <input
                    value={backendUrl}
                    onChange={(event) => {
                      setBackendUrl(event.target.value);
                      setConnectionState("idle");
                    }}
                    placeholder="http://localhost:8888/goapi"
                    inputMode="url"
                  />
                  {backendUrlValid === true ? (
                    <CheckCircle2 aria-label={tr("set_url_valid", "URL ถูกต้อง")} size={18} style={{ color: "var(--success, #16a34a)" }} />
                  ) : backendUrlValid === false ? (
                    <AlertCircle aria-label={tr("set_url_invalid", "รูปแบบ URL ไม่ถูกต้อง")} size={18} style={{ color: "var(--destructive, #dc2626)" }} />
                  ) : null}
                </div>
              </label>

              <label className="field-group">
                <span>{tr("setup_password", "รหัสผ่าน Setup")}</span>
                <div className="input-shell input-shell-large">
                  <LockKeyhole aria-hidden="true" size={18} />
                  <input
                    value={setupPassword}
                    disabled={authenticated}
                    onChange={(event) => setSetupPassword(event.target.value)}
                    placeholder={tr("set_enter_setup_code", "กรอกรหัสตั้งค่า")}
                    type={showSetupPassword ? "text" : "password"}
                  />
                  <button
                    className="icon-button"
                    type="button"
                    onClick={() => setShowSetupPassword((current) => !current)}
                    aria-label={showSetupPassword ? t(language, "hidePassword") : t(language, "showPassword")}
                  >
                    {showSetupPassword ? <EyeOff aria-hidden="true" size={18} /> : <Eye aria-hidden="true" size={18} />}
                  </button>
                </div>
              </label>
            </div>

            <div className="settings-button-row connection-actions">
              <button className="secondary-button" type="button" onClick={handleBackendConnectionTest} disabled={connectionState === "testing" || !backendUrlValid}>
                {connectionState === "testing" ? <Loader2 className="spin" size={17} /> : <CheckCircle2 aria-hidden="true" size={17} />}
                <span>{t(language, "testConnection")}</span>
              </button>
              <button className="secondary-button" type="button" onClick={handleCopyBackendUrl}>
                <Copy aria-hidden="true" size={17} />
                <span>{t(language, "copy")}</span>
              </button>
              <button className="secondary-button" type="button" onClick={handleSaveBackendUrl} disabled={!canSaveBackend}>
                <Save aria-hidden="true" size={17} />
                <span>{t(language, "save")}</span>
              </button>
              {!authenticated ? (
                <button className="primary-button settings-primary" type="button" onClick={handleVerifySetup} disabled={isBusy || !backendUrlValid}>
                  {loadingAction === "verify" ? <Loader2 className="spin" size={17} /> : <LogIn aria-hidden="true" size={17} />}
                  <span>{tr("set_sign_in_to_setup", "เข้าสู่ระบบ Setup")}</span>
                </button>
              ) : (
                <button className="secondary-button" type="button" onClick={logoutSetup}>
                  <LogOut aria-hidden="true" size={17} />
                  <span>{tr("set_sign_out_of_setup", "ออกจาก Setup")}</span>
                </button>
              )}
            </div>

            {authenticated ? (
              <div className="setup-connected">
                <CheckCircle2 aria-hidden="true" size={18} />
                <span>เชื่อมต่อแล้ว: {normalizeSetupBackendUrl(backendUrl)}</span>
              </div>
            ) : null}

            {message ? (
              <div className={`message ${messageState === "success" ? "success" : "error"}`}>
                {messageState === "success" ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
                <span>{message}</span>
              </div>
            ) : null}
          </section>

          <aside className="settings-side settings-rail">
            <section className="settings-card" aria-label={t(language, "displaySettings")}>
              <div className="settings-card-title compact">
                <span className="settings-card-icon">
                  <SettingsIcon aria-hidden="true" size={19} />
                </span>
                <div>
                  <p className="eyebrow">{t(language, "displaySettings")}</p>
                  <h2>{tr("theme_and_language", "ธีมและภาษา")}</h2>
                </div>
              </div>
              <div className="settings-control-stack">
                <AppHeaderControls language={language} onLanguageChange={setLanguage} showSettings={false} />
              </div>
            </section>

            <section className="settings-card" aria-label={t(language, "backendHistory")}>
              <div className="settings-card-title compact">
                <div>
                  <p className="eyebrow">{t(language, "backendUrl")}</p>
                  <h2>{t(language, "backendHistory")}</h2>
                </div>
              </div>

              {urlHistory.length > 0 ? (
                <>
                  <div className="history-list">
                    {urlHistory.map((url) => (
                      <div className="history-row" key={url}>
                        <button className="history-value" type="button" onClick={() => setBackendUrl(url)}>
                          {url}
                        </button>
                        <button className="icon-button danger-icon" type="button" onClick={() => handleRemoveHistory(url)} aria-label={t(language, "remove")}>
                          <Trash2 aria-hidden="true" size={17} />
                        </button>
                      </div>
                    ))}
                  </div>
                  <button className="secondary-button" type="button" onClick={handleClearHistory}>
                    <Trash2 aria-hidden="true" size={16} />
                    <span>{tr("set_clear_all_history", "ล้างประวัติทั้งหมด")}</span>
                  </button>
                </>
              ) : (
                <p className="settings-empty">{t(language, "noBackendHistory")}</p>
              )}
            </section>

            <section className="settings-card" aria-label={t(language, "help")}>
              <div className="settings-card-title compact">
                <span className="settings-card-icon">
                  <BookOpen aria-hidden="true" size={19} />
                </span>
                <div>
                  <p className="eyebrow">{t(language, "help")}</p>
                  <h2>{t(language, "manual")}</h2>
                </div>
              </div>
              <ManualLink language={language} screen="settings" />
            </section>
          </aside>
        </div>

        {authenticated ? (
          <>
            <div className="setup-workbench">
              <section className="setup-action-bar" aria-label="Setup actions">
                <div className="setup-action-copy">
                  <p className="eyebrow">Setup Config</p>
                  <h2>{tr("set_manage_system_values", "จัดการค่าระบบ")}</h2>
                  <p>{tr("set_load_test_import_save", "โหลด ทดสอบ Import/Export และบันทึก config กลับ backend")}</p>
                </div>
                <div className="setup-action-buttons">
                  <button className="secondary-button" type="button" onClick={() => loadSetupConfig()} disabled={isBusy}>
                    {loadingAction === "load" ? <Loader2 className="spin" size={17} /> : <RefreshCcw aria-hidden="true" size={17} />}
                    <span>{tr("gl_reload", "โหลดใหม่")}</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={handleTestAllConnections} disabled={isBusy}>
                    {loadingAction === "test-all" ? <Loader2 className="spin" size={17} /> : <Wifi aria-hidden="true" size={17} />}
                    <span>{tr("set_test_all", "ทดสอบทั้งหมด")}</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={handleExportConfig} disabled={isBusy}>
                    <Upload aria-hidden="true" size={17} />
                    <span>Export</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={() => setDialogMode("import")} disabled={isBusy}>
                    <Download aria-hidden="true" size={17} />
                    <span>Import</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={() => setDialogMode("change-password")} disabled={isBusy}>
                    <KeyRound aria-hidden="true" size={17} />
                    <span>{tr("set_change_password_short", "เปลี่ยนรหัส")}</span>
                  </button>
                </div>
              </section>

              {/* Server host + Auto-fill card — quick setup for single-server on-prem */}
              <section className="settings-section-group" aria-label={st("serverConnection", { th: "การเชื่อมต่อเซิร์ฟเวอร์", en: "Server Connection" })}>
                <header className="settings-section-heading">
                  <h2>{st("serverConnection", { th: "การเชื่อมต่อเซิร์ฟเวอร์", en: "Server Connection" })}</h2>
                  <p>
                    {st("serverConnectionHint", {
                      th: "ใส่ host เดียวแล้วกด Auto-fill เพื่อกรอก host ให้ทุกฐานข้อมูลและที่เก็บรูปอัตโนมัติ",
                      en: "Enter a single host and click Auto-fill to populate every database and storage endpoint automatically.",
                    })}
                  </p>
                </header>
                <section className="settings-card server-host-card">
                  <div className="server-host-row">
                    <label className="field-group">
                      <span>Server Host</span>
                      <div className="input-shell">
                        <Server aria-hidden="true" size={18} />
                        <input
                          type="text"
                          value={serverHost}
                          onChange={(event) => setServerHost(event.target.value)}
                          placeholder={tr("set_host_example", "เช่น 192.168.2.202 หรือ localhost")}
                          spellCheck={false}
                          autoComplete="off"
                        />
                      </div>
                    </label>
                    <button
                      className="primary-button auto-fill-button"
                      type="button"
                      onClick={handleAutoFillHost}
                      disabled={isBusy || !serverHost.trim()}
                      title={st("autofillTitle", { th: "กรอก host นี้ให้ทุกฐานข้อมูลและที่เก็บรูป", en: "Fill this host into every database and storage endpoint" })}
                    >
                      <Zap aria-hidden="true" size={17} />
                      <span>{st("autofillAll", { th: "Auto-fill ทุกฐานข้อมูล", en: "Auto-fill all" })}</span>
                    </button>
                  </div>
                </section>
              </section>

              {/* Render each section group with its heading */}
              {sectionGroups.map((group) => {
                if (group.id === "storage") {
                  // Storage: extract R2/S3 items from integrations, render as dedicated card.
                  const storageItems = (configMap["integrations"] ?? []).filter((item) =>
                    storageIntegrationKeys.has(item.key),
                  );
                  if (storageItems.length === 0) return null;
                  return (
                    <section className="settings-section-group" key={group.id} aria-label={group.titleTh}>
                      <header className="settings-section-heading">
                        <h2>{st(`group.${group.id}`, { th: group.titleTh, en: group.titleEn })}</h2>
                        <p>
                          {st("storageHint", {
                            th: "ตั้งค่าที่เก็บไฟล์แบบ S3 (เช่น MinIO) สำหรับเก็บรูปและไฟล์",
                            en: "Configure S3-compatible storage (e.g. MinIO) for images and files.",
                          })}
                        </p>
                      </header>
                      <div className="setup-config-sections">
                        {renderStorageCard(storageItems)}
                      </div>
                    </section>
                  );
                }

                const categoriesInGroup = group.categories.filter(
                  (cat) => cat === "integrations" || (configMap[cat]?.length ?? 0) > 0,
                );
                if (categoriesInGroup.length === 0) return null;

                return (
                  <section className="settings-section-group" key={group.id} aria-label={group.titleTh}>
                    <header className="settings-section-heading">
                      <h2>{st(`group.${group.id}`, { th: group.titleTh, en: group.titleEn })}</h2>
                    </header>
                    {group.id === "ai" ? (
                      <div className="integrations-wrapper">
                        {categoriesInGroup.map((category) =>
                          renderConfigSection(category, splitAiOnlyItems(configMap[category] ?? [])),
                        )}
                      </div>
                    ) : (
                      <div className="setup-config-sections">
                        {categoriesInGroup.map((category) =>
                          renderConfigSection(category, configMap[category] ?? []),
                        )}
                      </div>
                    )}
                  </section>
                );
              })}

              {/* Bottom Action Area: Save Config with Complete Connection Testing Guard */}
              <section
                className={`settings-card setup-save-card ${allTestsPassed ? "ready" : "pending"}`}
                aria-label={tr("set_save_config", "บันทึก Config")}
              >
                <div className="setup-save-status-block">
                  <div className="setup-save-icon-wrap">
                    {allTestsPassed ? (
                      <CheckCircle2 className="text-emerald-500" size={24} aria-hidden="true" />
                    ) : (
                      <AlertCircle className="text-amber-500" size={24} aria-hidden="true" />
                    )}
                  </div>
                  <div className="setup-save-copy">
                    <p className="eyebrow">
                      {allTestsPassed
                        ? st("readyToSave", { th: "พร้อมบันทึก", en: "Ready to Save" })
                        : st("testingRequired", { th: "ต้องทดสอบให้ครบก่อนบันทึก", en: "Testing Required" })}
                    </p>
                    <h3>
                      {allTestsPassed
                        ? st("allTestsPassedTitle", {
                            th: `ทดสอบการเชื่อมต่อผ่านครบถ้วนแล้ว (${testableCategories.length}/${testableCategories.length} รายการ)`,
                            en: `All connection tests passed (${testableCategories.length}/${testableCategories.length})`,
                          })
                        : st("testsPendingTitle", {
                            th: `ต้องทดสอบการเชื่อมต่อให้ครบก่อนบันทึก (ผ่านแล้ว ${testableCategories.length - pendingCategories.length}/${testableCategories.length} รายการ)`,
                            en: `Please test all connections before saving (${testableCategories.length - pendingCategories.length}/${testableCategories.length} passed)`,
                          })}
                    </h3>
                    <p className="setup-save-desc">
                      {allTestsPassed
                        ? st("allTestsPassedDesc", {
                            th: "ระบบตรวจสอบการเชื่อมต่อฐานข้อมูล PostgreSQL และที่เก็บรูปเรียบร้อยแล้ว พร้อมบันทึกการตั้งค่ากลับสู่ Backend",
                            en: "PostgreSQL and Storage connections are verified. Safe to save configuration to Backend.",
                          })
                        : st("testsPendingDesc", {
                            th: `รายการที่ต้องทดสอบให้ผ่าน: ${pendingCategories.map((c) => c.title).join(", ")} (กดปุ่ม "ทดสอบทั้งหมด" เพื่อตรวจสอบ)`,
                            en: `Required tests: ${pendingCategories.map((c) => c.title).join(", ")} (Click "Test All" to verify)`,
                          })}
                    </p>
                  </div>
                </div>

                <div className="setup-save-actions">
                  {!allTestsPassed ? (
                    <button
                      className="secondary-button"
                      type="button"
                      onClick={handleTestAllConnections}
                      disabled={isBusy}
                    >
                      {loadingAction === "test-all" ? <Loader2 className="spin" size={17} /> : <Wifi aria-hidden="true" size={17} />}
                      <span>{tr("set_test_all", "ทดสอบทั้งหมด")}</span>
                    </button>
                  ) : null}
                  <button
                    className={`primary-button settings-primary ${!allTestsPassed ? "opacity-60 cursor-not-allowed" : ""}`}
                    type="button"
                    onClick={handleSaveConfig}
                    disabled={isBusy || !allTestsPassed}
                    title={
                      !allTestsPassed
                        ? st("cannotSaveTooltip", {
                            th: `ต้องทดสอบการเชื่อมต่อให้ผ่านครบก่อน: ${pendingCategories.map((c) => c.title).join(", ")}`,
                            en: `Must pass all connection tests: ${pendingCategories.map((c) => c.title).join(", ")}`,
                          })
                        : undefined
                    }
                  >
                    {loadingAction === "save" ? <Loader2 className="spin" size={17} /> : <Save aria-hidden="true" size={17} />}
                    <span>{tr("set_save_config", "บันทึก Config")}</span>
                  </button>
                </div>
              </section>
            </div>
          </>
        ) : null}
      </section>

      {dialogMode === "import" ? renderImportDialog() : null}
      {dialogMode === "change-password" ? renderChangePasswordDialog() : null}
      {confirmationDialog}
      {reloadingConfig ? (
        <div
          role="status"
          aria-live="polite"
          className="fixed inset-0 z-[3000] grid place-items-center bg-background/70 px-4 backdrop-blur-sm"
        >
          <div className="grid max-w-sm justify-items-center gap-3 rounded-2xl border border-border bg-card px-8 py-7 text-center shadow-2xl">
            <Loader2 className="spin" size={34} aria-hidden="true" />
            <strong className="text-base text-foreground">{tr("set_reloading_config", "กำลังโหลด config ใหม่…")}</strong>
            <span className="text-sm text-muted-foreground">
              {tr("set_saved_reconnecting_then_login", "บันทึกแล้ว — ระบบกำลัง reconnect ฐานข้อมูล แล้วจะกลับไปหน้าเข้าสู่ระบบ")}
            </span>
          </div>
        </div>
      ) : null}
    </main>
  );

  function renderConfigSection(category: string, items: ConfigItem[]) {
    const categoryDef = getCategoryDef(category);
    const testResult = testResults[category];
    const canTest = Boolean(categoryDef.testType);

    if (category === "integrations") {
      return renderIntegrationsSection(items);
    }

    const displayItems = items;

    return (
      <section className="settings-card config-card" key={category} aria-label={categoryDef.title}>
        <div className="config-card-header">
          <div className="settings-card-title">
            <span className="settings-card-icon">{categoryIcon(category)}</span>
            <div>
              <p className="eyebrow">{category}</p>
              <h2>{categoryDef.title}</h2>
              <p className="config-description">{categoryDef.description}</p>
            </div>
          </div>
          <div className="config-card-meta">
            <span className="field-count">{displayItems.length} fields</span>
            {testResult ? <TestBadge result={testResult} label={testBadgeLabel(testResult)} /> : null}
          </div>
        </div>

        <div className="config-field-grid">{displayItems.map((item) => renderConfigField(item))}</div>

        {canTest ? (
          <div className="config-test-row">
            <button className="secondary-button" type="button" onClick={() => void handleTestConnection(category)} disabled={Boolean(testingCategory)}>
              {testingCategory === category ? <Loader2 className="spin" size={17} /> : <Wifi aria-hidden="true" size={17} />}
              <span>ทดสอบ {categoryDef.title}</span>
            </button>
            {testResult?.message ? <p className={testResult.status === "success" ? "test-message success" : "test-message error"}>{formatTestMessage(testResult)}</p> : null}
          </div>
        ) : null}
      </section>
    );
  }

  // Split out AI-provider items only (R2/S3 storage items render in the Storage section).
  function splitAiOnlyItems(items: ConfigItem[]): ConfigItem[] {
    return items.filter((item) => !storageIntegrationKeys.has(item.key));
  }

  // Dedicated storage card — R2/S3 settings grouped together under "ที่เก็บรูปและไฟล์".
  // Test storage connection: simple direct HTTP probe to the S3/MinIO endpoint.
  // No backend setup password needed — just check the endpoint is reachable.
  // Returns the result so callers (e.g. test-all) can read it synchronously.
  async function handleTestStorageConnection(): Promise<TestResult | null> {
    if (isBusy) return null;
    const items = (configMap["integrations"] ?? []).filter((item) =>
      storageIntegrationKeys.has(item.key),
    );
    // Probe the S3 endpoint that backend actually connects to (s3endpoint e.g. http://minio:9000).
    // If empty, fallback to s3publicendpoint.
    const s3Endpoint = items.find((item) => item.key === "s3endpoint")?.value?.trim() || "";
    const s3PublicEndpoint = items.find((item) => item.key === "s3publicendpoint")?.value?.trim() || "";
    const endpoint = s3Endpoint || s3PublicEndpoint;
    if (!endpoint) {
      const failed: TestResult = {
        status: "failed",
        message: st("s3EndpointRequired", { th: "กรุณากรอก S3 Endpoint ก่อน", en: "Please fill S3 Endpoint first" }),
      };
      setTestResults((current) => ({ ...current, storage: failed }));
      return failed;
    }
    setTestingCategory("storage");
    setTestResults((current) => ({
      ...current,
      storage: { status: "testing", message: st("testing", { th: "กำลังทดสอบ...", en: "Testing..." }) },
    }));
    const start = Date.now();
    try {
      // Server-side probe via /api/storage/health — avoids browser CORS blocks
      // that occur when fetching a cross-origin S3/MinIO endpoint directly.
      const probeResponse = await fetch("/api/storage/health", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ endpoint, publicEndpoint: s3PublicEndpoint }),
        signal: AbortSignal.timeout(12000),
      });
      const data = (await probeResponse.json()) as {
        success?: boolean;
        message?: string;
        latencyMs?: number;
        httpStatus?: number;
      };
      const latencyMs = data.latencyMs ?? Date.now() - start;
      if (data.success) {
        const success: TestResult = {
          status: "success",
          message:
            st("reachable", { th: "เชื่อมต่อได้ ({0})", en: "Reachable ({0})" }).replace(
              "{0}",
              `${latencyMs}ms${data.httpStatus ? `, HTTP ${data.httpStatus}` : ""}`,
            ),
          latencyMs,
        };
        setTestResults((current) => ({ ...current, storage: success }));
        return success;
      }
      const failed: TestResult = {
        status: "failed",
        message: data.message ?? st("unreachable", { th: "เชื่อมต่อไม่ได้", en: "Unreachable" }),
        latencyMs,
      };
      setTestResults((current) => ({ ...current, storage: failed }));
      return failed;
    } catch (error) {
      const failed: TestResult = {
        status: "failed",
        message: error instanceof Error && error.message
          ? error.message
          : st("unreachable", { th: "เชื่อมต่อไม่ได้", en: "Unreachable" }),
      };
      setTestResults((current) => ({ ...current, storage: failed }));
      return failed;
    } finally {
      setTestingCategory("");
    }
  }

  function renderStorageCard(items: ConfigItem[]) {
    // Show only S3-compatible fields (R2 has been removed).
    const displayItems = items.filter((item) => storageIntegrationKeys.has(item.key));
    const testResult = testResults["storage"];
    const isTesting = testingCategory === "storage";
    return (
      <section className="settings-card config-card storage-card" key="storage" aria-label="Storage">
        <div className="config-card-header">
          <div className="settings-card-title">
            <span className="settings-card-icon">
              <Cloud aria-hidden="true" size={19} />
            </span>
            <div>
              <p className="eyebrow">storage</p>
              <h2>{st("group.storage", { th: "ที่เก็บรูปและไฟล์", en: "Image & File Storage" })}</h2>
              <p className="config-description">
                {st("storageDescription", {
                  th: "S3-compatible (เช่น MinIO) — เก็บรูปและไฟล์ในเซิร์ฟเวอร์ของเรา (on-prem)",
                  en: "S3-compatible (e.g. MinIO) — store images and files on our own server (on-prem)",
                })}
              </p>
            </div>
          </div>
          <div className="config-card-meta">
            <span className="field-count-pill">{displayItems.length} {st("fields", { th: "ฟิลด์", en: "fields" })}</span>
            {testResult ? <TestBadge result={testResult} label={testBadgeLabel(testResult)} /> : null}
          </div>
        </div>
        <div className="config-field-grid">
          {displayItems.map((item) => renderConfigField(item))}
        </div>
        <div className="config-test-row">
          <button
            className="secondary-button"
            type="button"
            onClick={() => void handleTestStorageConnection()}
            disabled={isBusy || isTesting}
          >
            {isTesting ? <Loader2 className="spin" size={17} /> : <Wifi aria-hidden="true" size={17} />}
            <span>{st("testStorage", { th: "ทดสอบที่เก็บรูป", en: "Test storage" })}</span>
          </button>
          {testResult?.message ? (
            <span className={`test-message ${testResult.status}`}>{formatTestMessage(testResult)}</span>
          ) : null}
        </div>
      </section>
    );
  }


  function renderIntegrationsSection(items: ConfigItem[]) {
    const providerItem = items.find((item) => item.key === "aiprovider");
    const providerIds = AI_PROVIDERS.map((provider) => provider.id);
    const otherItems = items.filter((item) => {
      if (item.key === "aiprovider") return false;
      return !providerIds.some((providerId) => item.key.startsWith(providerId));
    });

    return (
      <section className="settings-card config-card integrations-card" key="integrations" aria-label="Integrations">
        <div className="config-card-header">
          <div className="settings-card-title">
            <span className="settings-card-icon">
              <Zap aria-hidden="true" size={19} />
            </span>
            <div>
              <p className="eyebrow">integrations</p>
              <h2>{tr("set_ai_providers_and_integrations", "AI Providers และ Integrations")}</h2>
              <p className="config-description">{tr("set_provider_fallback_order", "อิงลำดับ provider fallback จาก Flutter: OpenRouter, Groq, DeepSeek, Gemini")}</p>
            </div>
          </div>
          <div className="config-card-meta">
            <span className="field-count">{items.length} fields</span>
          </div>
        </div>

        {providerItem ? (
          <div className="provider-selector">
            <p className="provider-selector-title">AI Provider — Auto Fallback</p>
            <div className="provider-options">
              {["", ...providerIds].map((providerId) => {
                const label = providerId ? AI_PROVIDERS.find((provider) => provider.id === providerId)?.name ?? providerId : "Auto";
                const selected = providerItem.value.trim().toLowerCase() === providerId || (!providerItem.value.trim() && !providerId);
                return (
                  <button
                    aria-pressed={selected}
                    className={selected ? "provider-option selected" : "provider-option"}
                    key={providerId || "auto"}
                    type="button"
                    onClick={() => updateItem("integrations", "aiprovider", providerId)}
                  >
                    {selected ? <CheckCircle2 aria-hidden="true" size={15} /> : null}
                    <span>{label}</span>
                  </button>
                );
              })}
            </div>
          </div>
        ) : null}

        <div className="provider-card-grid">
          {AI_PROVIDERS.map((provider) => {
            const providerItems = items.filter((item) => item.key.startsWith(provider.id));
            const modelItem = providerItems.find((item) => item.key.endsWith("model"));
            const freeMode = !modelItem?.value || modelItem.value.endsWith(":free") || provider.freeModels.includes(modelItem.value);

            return (
              <section className="provider-card" key={provider.id} aria-label={provider.name}>
                <div className="provider-card-header">
                  <div>
                    <h3>{provider.name}</h3>
                    <p>{provider.description}</p>
                  </div>
                  <span>{provider.badge}</span>
                </div>

                <div className="config-field-grid compact-grid">{providerItems.map((item) => renderConfigField(item))}</div>

                {modelItem ? (
                  <div className="model-mode">
                    <button
                      className={freeMode ? "provider-option selected" : "provider-option"}
                      type="button"
                      onClick={() => updateItem("integrations", modelItem.key, "")}
                      aria-pressed={freeMode}
                    >
                      {freeMode ? <CheckCircle2 aria-hidden="true" size={15} /> : null}
                      <span>{tr("set_use_free_models_automatically", "ใช้โมเดลฟรีอัตโนมัติ")}</span>
                    </button>
                    <ModelChips
                      title={tr("set_free_model_examples", "โมเดลฟรีตัวอย่าง")}
                      models={provider.freeModels}
                      onSelect={(model) => updateItem("integrations", modelItem.key, model)}
                    />
                    <ModelChips
                      title={tr("set_paid_models", "โมเดลเสียเงิน")}
                      models={provider.paidModels}
                      onSelect={(model) => updateItem("integrations", modelItem.key, model)}
                    />
                  </div>
                ) : null}

                <a className="manual-link provider-link" href={provider.registerUrl} target="_blank" rel="noreferrer">
                  ลงทะเบียน API Key — {provider.name}
                </a>
              </section>
            );
          })}
        </div>

        {otherItems.length > 0 ? (
          <>
            <h3 className="subsection-title">Storage / External Services</h3>
            <div className="config-field-grid">{otherItems.map((item) => renderConfigField(item))}</div>
          </>
        ) : null}
      </section>
    );
  }

  function getDefaultValue(category: string, key: string): string | null {
    if (key === "port" && category === "postgresql") return "5432";
    if (key === "timezone" && category === "postgresql") return "Asia/Bangkok";
    if (key === "serviceport" && category === "service") return "8888";
    return null;
  }

  function renderConfigField(item: ConfigItem) {
    const control = getFieldControl(item);
    if (control?.type === "checkbox") return renderCheckboxField(item);
    if (control?.type === "radio") return renderRadioField(item, control.options);

    const defaultValue = getDefaultValue(item.category, item.key);
    const hasDefault = defaultValue !== null;

    const isWideField =
      item.key === "uri" ||
      item.key === "serverurl" ||
      item.key === "hostapi" ||
      item.key === "corsallowedorigins" ||
      item.key === "jwtsecretkey" ||
      item.key.endsWith("apikey") ||
      item.key.endsWith("secretaccesskey") ||
      item.key.endsWith("accesskeyid") ||
      item.key === "azureaccountkey";

    return (
      <label className={`field-group config-field ${isWideField ? "field-group-wide" : ""}`} key={`${item.category}.${item.key}`}>
        <span>{fieldLabels[item.key] ?? item.key}</span>
        <div className="input-shell">
          {item.isSecret ? <KeyRound aria-hidden="true" size={17} /> : <SettingsIcon aria-hidden="true" size={17} />}
          <input
            autoComplete="off"
            value={item.value}
            onChange={(event) => updateItem(item.category, item.key, event.target.value)}
            placeholder={item.description || item.key}
            type={item.isSecret ? "password" : "text"}
          />
          {hasDefault && item.value.trim() !== defaultValue && (
            <button
              className="default-button-inline"
              type="button"
              onClick={() => updateItem(item.category, item.key, defaultValue)}
              title={tr("set_use_default_value", "ใช้ค่าเริ่มต้น: {0}").replace("{0}", String(defaultValue))}
            >
              Default
            </button>
          )}
        </div>
        {item.description ? <small>{item.description}</small> : null}
      </label>
    );
  }


  function renderCheckboxField(item: ConfigItem) {
    const checked = normalizeBooleanValue(item.value);

    return (
      <div className="field-group config-field choice-field" key={`${item.category}.${item.key}`}>
        <span>{fieldLabels[item.key] ?? item.key}</span>
        <button
          aria-pressed={checked}
          className={checked ? "checkbox-control checked" : "checkbox-control"}
          type="button"
          onClick={() => updateItem(item.category, item.key, checked ? "false" : "true")}
        >
          <span className="checkbox-box">{checked ? <CheckCircle2 aria-hidden="true" size={16} /> : null}</span>
          <span>{checked ? tr("alert_enabled", "เปิดใช้งาน") : tr("alert_disabled", "ปิดใช้งาน")}</span>
        </button>
        {item.description ? <small>{item.description}</small> : null}
      </div>
    );
  }

  function renderRadioField(item: ConfigItem, options: FieldOption[]) {
    const currentValue = item.value.trim();
    const selectedValue = currentValue || options[0]?.value || "";

    return (
      <fieldset className="field-group config-field choice-field" key={`${item.category}.${item.key}`}>
        <legend>{fieldLabels[item.key] ?? item.key}</legend>
        <div className="radio-options">
          {options.map((option) => {
            const selected = selectedValue.toLowerCase() === option.value.toLowerCase();

            return (
              <button
                aria-pressed={selected}
                className={selected ? "radio-option selected" : "radio-option"}
                key={option.value}
                type="button"
                onClick={() => updateItem(item.category, item.key, option.value)}
              >
                <span>{option.label}</span>
                {option.description ? <small>{option.description}</small> : null}
              </button>
            );
          })}
        </div>
        {item.description ? <small>{item.description}</small> : null}
      </fieldset>
    );
  }

  function renderImportDialog() {
    return (
      <div className="dialog-backdrop" onClick={() => setDialogMode(null)}>
        <form className="settings-modal" onClick={(event) => event.stopPropagation()} onSubmit={handleImportConfig}>
          <div className="dialog-header">
            <div>
              <p className="eyebrow">Import</p>
              <h2>Import Config JSON</h2>
            </div>
            <button className="icon-button dialog-close" type="button" onClick={() => setDialogMode(null)} aria-label={t(language, "close")}>
              <X aria-hidden="true" size={18} />
            </button>
          </div>
          <textarea
            className="settings-textarea"
            value={importText}
            onChange={(event) => setImportText(event.target.value)}
            placeholder='{"postgresql":{"host":"postgres"} }'
            rows={12}
          />
          <div className="settings-button-row">
            <button className="secondary-button" type="button" onClick={() => setDialogMode(null)}>
              {tr("cancel", "ยกเลิก")}
            </button>
            <button className="primary-button settings-primary" type="submit">
              Import
            </button>
          </div>
        </form>
      </div>
    );
  }

  function renderChangePasswordDialog() {
    return (
      <div className="dialog-backdrop" onClick={() => setDialogMode(null)}>
        <form className="settings-modal" onClick={(event) => event.stopPropagation()} onSubmit={handleChangePassword}>
          <div className="dialog-header">
            <div>
              <p className="eyebrow">Setup Password</p>
              <h2>{tr("set_change_setup_password", "เปลี่ยนรหัสผ่าน Setup")}</h2>
            </div>
            <button className="icon-button dialog-close" type="button" onClick={() => setDialogMode(null)} aria-label={t(language, "close")}>
              <X aria-hidden="true" size={18} />
            </button>
          </div>
          <label className="field-group">
            <span>{tr("set_current_password", "รหัสผ่านเดิม")}</span>
            <div className="input-shell">
              <LockKeyhole aria-hidden="true" size={18} />
              <input value={currentSetupPassword} onChange={(event) => setCurrentSetupPassword(event.target.value)} type="password" />
            </div>
          </label>
          <label className="field-group">
            <span>{tr("new_password", "รหัสผ่านใหม่")}</span>
            <div className="input-shell">
              <KeyRound aria-hidden="true" size={18} />
              <input value={newSetupPassword} onChange={(event) => setNewSetupPassword(event.target.value)} type="password" />
            </div>
          </label>
          <div className="settings-button-row">
            <button className="secondary-button" type="button" onClick={() => setDialogMode(null)}>
              {tr("cancel", "ยกเลิก")}
            </button>
            <button className="primary-button settings-primary" type="submit" disabled={loadingAction === "change-password"}>
              {loadingAction === "change-password" ? <Loader2 className="spin" size={17} /> : <Save aria-hidden="true" size={17} />}
              <span>{tr("save", "บันทึก")}</span>
            </button>
          </div>
        </form>
      </div>
    );
  }
}

function readUrlHistory(): string[] {
  const raw = localStorage.getItem(storageKeys.backendUrlHistory);
  if (!raw) return [];

  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((value): value is string => typeof value === "string").slice(0, 10);
  } catch {
    return [];
  }
}

function responseToTestResult(data: SetupResponse, labels: { success: string; failed: string }): TestResult {
  const success = data.success === true;
  return {
    status: success ? "success" : "failed",
    message: data.message ?? (success ? labels.success : labels.failed),
    latencyMs: typeof data.latencyms === "number" ? data.latencyms : undefined,
    errorCode: typeof data.errorcode === "string" ? data.errorcode : undefined,
    database: typeof data.database === "string" ? data.database : undefined,
  };
}

function formatTestMessage(result: TestResult): string {
  return `${result.message}${result.latencyMs != null ? ` (${result.latencyMs}ms)` : ""}`;
}

function categoryIcon(category: string): ReactNode {
  switch (category) {
    case "postgresql":
      return <Table2 aria-hidden="true" size={19} />;
    case "storage":
      return <Cloud aria-hidden="true" size={19} />;
    default:
      return <SettingsIcon aria-hidden="true" size={19} />;
  }
}

function TestBadge({ result, label }: { result: TestResult; label: string }) {
  return <span className={`test-badge ${result.status}`}>{label}</span>;
}

function ModelChips({ title, models, onSelect }: { title: string; models: string[]; onSelect: (model: string) => void }) {
  if (models.length === 0) return null;

  return (
    <div className="model-chip-group">
      <p>{title}</p>
      <div>
        {models.slice(0, 8).map((model) => (
          <button className="model-chip" key={model} type="button" onClick={() => onSelect(model)}>
            {model}
          </button>
        ))}
      </div>
    </div>
  );
}

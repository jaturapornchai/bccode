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
import { FormEvent, ReactNode, useEffect, useMemo, useState } from "react";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { persistLanguagePreferenceCookies } from "@/lib/backend-language-preload";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
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
type LoadingAction = "idle" | "verify" | "load" | "save" | "seed" | "test-all" | "change-password";
type DialogMode = "import" | "change-password" | null;

type SetupResponse = {
  success?: boolean;
  message?: string;
  data?: unknown;
  latency_ms?: unknown;
  error_code?: unknown;
  database?: unknown;
};

const DEFAULT_BACKEND_URL = "http://localhost:8888/goapi";
const storageKeys = {
  backendUrl: "backend_url",
  backendUrlHistory: "backend_url_history",
  language: "user_language",
};

export function SettingsScreen() {
  const [backendUrl, setBackendUrl] = useState(DEFAULT_BACKEND_URL);
  const [urlHistory, setUrlHistory] = useState<string[]>([]);
  const [language, setLanguage] = useState<LanguageCode>("th");
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
  const [mongodbMode, setMongodbMode] = useState<"uri" | "fields">("uri");
  const { confirm, confirmationDialog } = useConfirmDialog();

  const handleSetMongodbMode = (mode: "uri" | "fields") => {
    setMongodbMode(mode);
    setConfigMap((current) => {
      const items = current["mongodb"] ?? [];
      const nextItems = items.map((item) => {
        if (mode === "uri") {
          // เคลียร์ค่า host, port, username, password เพื่อไม่ให้สับสน
          if (item.key === "host" || item.key === "port" || item.key === "username" || item.key === "password") {
            return { ...item, value: "" };
          }
        } else {
          // เคลียร์ค่า URI ก่อน เพื่อให้ฟังก์ชันประกอบเริ่มคำนวณจากฟิลด์ใหม่
          if (item.key === "uri") {
            return { ...item, value: "" };
          }
        }
        return item;
      });
      return { ...current, mongodb: nextItems };
    });
  };


  const canSaveBackend = useMemo(() => backendUrl.trim().length > 0, [backendUrl]);
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
      setStatus("กรุณาระบุรหัสผ่าน Setup", "error");
      return;
    }

    setLoadingAction("verify");
    try {
      const data = await callSetup("verify-password", { password: setupPassword }, normalized);
      persistBackendUrl(normalized);
      setBackendUrl(normalized);
      setAuthenticated(true);
      setStatus(data.message ?? "ยืนยันรหัสผ่านสำเร็จ", "success");
      await loadSetupConfig(setupPassword, normalized);
    } catch (error) {
      setAuthenticated(false);
      setStatus(error instanceof Error ? error.message : "เข้าสู่ระบบ Setup ไม่สำเร็จ", "error");
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
      setStatus(`โหลด config สำเร็จ`, "success");

      // ตรวจสอบว่าโหมดเริ่มต้นควรเป็นแบบกรอก URI หรือแบบแยกฟิลด์
      const mongoItems = merged["mongodb"] ?? [];
      const host = mongoItems.find((i) => i.key === "host")?.value.trim() ?? "";
      if (host && !host.includes("xxxx")) {
        setMongodbMode("fields");
      } else {
        setMongodbMode("uri");
      }
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "โหลด config ล้มเหลว", "error");
    } finally {
      setLoadingAction("idle");
    }
  }


  async function handleSaveConfig() {
    const errors = validateConfig(configMap);
    if (errors.length > 0) {
      setStatus(`ตรวจสอบ config ไม่ผ่าน: ${errors.slice(0, 3).join(", ")}`, "error");
      return;
    }

    const confirmed = await confirm({
      title: "ยืนยันบันทึก Config",
      description: "บันทึก config ไปที่ backend และเขียนทับ bootstrap.json",
      details: "ตรวจสอบ Backend URL และค่าที่แก้ไขให้เรียบร้อยก่อนยืนยัน เพราะค่าเหล่านี้มีผลกับการเชื่อมต่อระบบ",
      confirmLabel: "บันทึก Config",
      cancelLabel: "ยกเลิก",
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
      setStatus(data.message ?? "บันทึก config สำเร็จ", "success");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "บันทึก config ล้มเหลว", "error");
    } finally {
      setLoadingAction("idle");
    }
  }

  async function handleSeedConfig() {
    setLoadingAction("seed");
    try {
      const data = await callSetup("config/seed", { password: setupPassword });
      setStatus(data.message ?? "Seed config สำเร็จ", "success");
      await loadSetupConfig();
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Seed config ล้มเหลว", "error");
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
      const result = responseToTestResult(data);
      setTestResults((current) => ({ ...current, [category]: result }));
      return result.status === "success";
    } catch (error) {
      setTestResults((current) => ({
        ...current,
        [category]: {
          status: "failed",
          message: error instanceof Error ? error.message : "ทดสอบไม่สำเร็จ",
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
    const results = await Promise.all(categories.map((category) => handleTestConnection(category)));
    const successCount = results.filter(Boolean).length;
    setStatus(`ทดสอบ connection แล้ว ${successCount}/${categories.length} สำเร็จ`, successCount === categories.length ? "success" : "error");
    setLoadingAction("idle");
  }

  async function handleCreateClickHouseDatabase(database: string) {
    const items = configMap.clickhouse ?? [];
    const payload = buildConnectionPayload("clickhouse", items, setupPassword);
    setTestingCategory("clickhouse");
    try {
      const data = await callSetup("create-clickhouse-database", {
        password: setupPassword,
        host: payload.host,
        port: payload.port,
        user: payload.user,
        password2: payload.password2,
        database,
      });
      setStatus(data.message ?? `สร้าง database ${database} สำเร็จ`, "success");
      await handleTestConnection("clickhouse");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "สร้าง ClickHouse database ล้มเหลว", "error");
    } finally {
      setTestingCategory("");
    }
  }

  async function handleExportConfig() {
    try {
      await navigator.clipboard.writeText(configMapToExportJson(configMap));
      setStatus("คัดลอก config JSON แล้ว", "success");
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
      setStatus("Import config แล้ว ยังไม่ได้บันทึกลง backend", "success");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Import config ล้มเหลว", "error");
    }
  }

  async function handleChangePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!currentSetupPassword || !newSetupPassword) {
      setStatus("กรุณาระบุรหัสผ่านเดิมและรหัสผ่านใหม่", "error");
      return;
    }

    setLoadingAction("change-password");
    try {
      const data = await callSetup("change-password", {
        current_password: currentSetupPassword,
        new_password: newSetupPassword,
      });
      setSetupPassword(newSetupPassword);
      setDialogMode(null);
      setCurrentSetupPassword("");
      setNewSetupPassword("");
      setStatus(data.message ?? "เปลี่ยนรหัสผ่านสำเร็จ", "success");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "เปลี่ยนรหัสผ่านล้มเหลว", "error");
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
    setStatus("ออกจากระบบ Setup แล้ว", "success");
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
      throw new Error(data.message ?? "Setup backend ตอบกลับผิดปกติ");
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
            <h1>ศูนย์ตั้งค่าระบบ</h1>
            <p>จัดการ Backend URL, Setup Config, database connections, storage, AI providers, ธีม และภาษาในจอเดียว</p>
          </div>

          <div className="header-actions settings-header-actions">
            <ManualLink compact language={language} screen="settings" />
            <AppHeaderControls language={language} onLanguageChange={setLanguage} showSettings={false} />
          </div>
        </header>

        <section className="settings-overview" aria-label="สถานะระบบ">
          <div className="overview-tile">
            <span>Backend</span>
            <strong>{normalizeSetupBackendUrl(backendUrl) || "ยังไม่ได้ตั้งค่า"}</strong>
          </div>
          <div className="overview-tile">
            <span>Setup</span>
            <strong>{authenticated ? "ยืนยันแล้ว" : "ยังไม่เข้าสู่ Setup"}</strong>
          </div>
          <div className="overview-tile">
            <span>Config Sections</span>
            <strong>{authenticated ? `${orderedCategories.length} กลุ่ม` : "ล็อกไว้ก่อนยืนยันรหัส"}</strong>
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
                <h2>เชื่อมต่อ Backend</h2>
                <p className="config-description">เริ่มจากตรวจ URL แล้วค่อยยืนยันรหัส Setup เพื่อโหลด config จาก backend</p>
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
                </div>
              </label>

              <label className="field-group">
                <span>รหัสผ่าน Setup</span>
                <div className="input-shell input-shell-large">
                  <LockKeyhole aria-hidden="true" size={18} />
                  <input
                    value={setupPassword}
                    disabled={authenticated}
                    onChange={(event) => setSetupPassword(event.target.value)}
                    placeholder="default: 12345"
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
              <button className="secondary-button" type="button" onClick={handleBackendConnectionTest} disabled={connectionState === "testing"}>
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
                <button className="primary-button settings-primary" type="button" onClick={handleVerifySetup} disabled={isBusy}>
                  {loadingAction === "verify" ? <Loader2 className="spin" size={17} /> : <LogIn aria-hidden="true" size={17} />}
                  <span>เข้าสู่ระบบ Setup</span>
                </button>
              ) : (
                <button className="secondary-button" type="button" onClick={logoutSetup}>
                  <LogOut aria-hidden="true" size={17} />
                  <span>ออกจาก Setup</span>
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
                  <h2>ธีมและภาษา</h2>
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
                  <h2>จัดการค่าระบบ</h2>
                  <p>โหลด ทดสอบ Import/Export และบันทึก config กลับ backend</p>
                </div>
                <div className="setup-action-buttons">
                  <button className="secondary-button" type="button" onClick={() => loadSetupConfig()} disabled={isBusy}>
                    {loadingAction === "load" ? <Loader2 className="spin" size={17} /> : <RefreshCcw aria-hidden="true" size={17} />}
                    <span>โหลดใหม่</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={handleTestAllConnections} disabled={isBusy}>
                    {loadingAction === "test-all" ? <Loader2 className="spin" size={17} /> : <Wifi aria-hidden="true" size={17} />}
                    <span>ทดสอบทั้งหมด</span>
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
                    <span>เปลี่ยนรหัส</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={handleSeedConfig} disabled={isBusy}>
                    {loadingAction === "seed" ? <Loader2 className="spin" size={17} /> : <Cloud aria-hidden="true" size={17} />}
                    <span>Seed Env</span>
                  </button>
                  <button className="primary-button settings-primary" type="button" onClick={handleSaveConfig} disabled={isBusy}>
                    {loadingAction === "save" ? <Loader2 className="spin" size={17} /> : <Save aria-hidden="true" size={17} />}
                    <span>บันทึก Config</span>
                  </button>
                </div>
              </section>

              <div className="setup-config-sections">
                {orderedCategories
                  .filter((cat) => cat !== "integrations")
                  .map((category) => renderConfigSection(category, configMap[category] ?? []))}
              </div>

              {orderedCategories.includes("integrations") && (
                <div className="integrations-wrapper">
                  {renderConfigSection("integrations", configMap["integrations"] ?? [])}
                </div>
              )}
            </div>
          </>
        ) : null}
      </section>

      {dialogMode === "import" ? renderImportDialog() : null}
      {dialogMode === "change-password" ? renderChangePasswordDialog() : null}
      {confirmationDialog}
    </main>
  );

  function renderConfigSection(category: string, items: ConfigItem[]) {
    const categoryDef = getCategoryDef(category);
    const testResult = testResults[category];
    const canTest = Boolean(categoryDef.testType);

    if (category === "integrations") {
      return renderIntegrationsSection(items);
    }

    // สำหรับ MongoDB: กรองฟิลด์ที่จะแสดงผลตามโหมดเชื่อมต่อ
    let displayItems = items;
    if (category === "mongodb") {
      if (mongodbMode === "uri") {
        displayItems = items.filter((item) => item.key === "uri" || item.key === "database");
      } else {
        displayItems = items;
      }
    }

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
            {testResult ? <TestBadge result={testResult} /> : null}
          </div>
        </div>

        {category === "mongodb" && (
          <div className="mongodb-mode-row">
            <div className="mongodb-mode-selector">
              <button
                className={`mongodb-mode-button ${mongodbMode === "uri" ? "active" : ""}`}
                type="button"
                onClick={() => handleSetMongodbMode("uri")}
              >
                กรอก URI โดยตรง
              </button>
              <button
                className={`mongodb-mode-button ${mongodbMode === "fields" ? "active" : ""}`}
                type="button"
                onClick={() => handleSetMongodbMode("fields")}
              >
                ระบุรายละเอียดแยกฟิลด์
              </button>
            </div>
          </div>
        )}

        <div className="config-field-grid">{displayItems.map((item) => renderConfigField(item))}</div>

        {canTest ? (
          <div className="config-test-row">
            <button className="secondary-button" type="button" onClick={() => void handleTestConnection(category)} disabled={Boolean(testingCategory)}>
              {testingCategory === category ? <Loader2 className="spin" size={17} /> : <Wifi aria-hidden="true" size={17} />}
              <span>ทดสอบ {categoryDef.title}</span>
            </button>
            {testResult?.message ? <p className={testResult.status === "success" ? "test-message success" : "test-message error"}>{formatTestMessage(testResult)}</p> : null}
            {category === "clickhouse" && testResult?.errorCode === "database_not_exist" && testResult.database ? (
              <button className="secondary-button warning-button" type="button" onClick={() => void handleCreateClickHouseDatabase(testResult.database!)}>
                <Database aria-hidden="true" size={17} />
                <span>สร้าง database &quot;{testResult.database}&quot;</span>
              </button>
            ) : null}
          </div>
        ) : null}
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
              <h2>AI Providers และ Integrations</h2>
              <p className="config-description">อิงลำดับ provider fallback จาก Flutter: OpenRouter, Groq, DeepSeek, Gemini</p>
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
                      <span>ใช้โมเดลฟรีอัตโนมัติ</span>
                    </button>
                    <ModelChips
                      title="โมเดลฟรีตัวอย่าง"
                      models={provider.freeModels}
                      onSelect={(model) => updateItem("integrations", modelItem.key, model)}
                    />
                    <ModelChips
                      title="โมเดลเสียเงิน"
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
    if (key === "port") {
      if (category === "postgresql") return "5432";
      if (category === "clickhouse") return "9000";
      if (category === "mongodb") return "27017";
    }
    if (key === "serverurl" && category === "kafka") return "127.0.0.1:9092";
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

    const isMongoUriReadOnly = item.category === "mongodb" && item.key === "uri" && mongodbMode === "fields";

    return (
      <label className={`field-group config-field ${isWideField ? "field-group-wide" : ""}`} key={`${item.category}.${item.key}`}>
        <span>{fieldLabels[item.key] ?? item.key}</span>
        <div className="input-shell">
          {item.isSecret ? <KeyRound aria-hidden="true" size={17} /> : <SettingsIcon aria-hidden="true" size={17} />}
          <input
            autoComplete="off"
            value={item.value}
            onChange={(event) => updateItem(item.category, item.key, event.target.value)}
            placeholder={isMongoUriReadOnly ? "ระบบประกอบ URI อัตโนมัติ..." : (item.description || item.key)}
            type={item.isSecret ? "password" : "text"}
            readOnly={isMongoUriReadOnly}
            disabled={isMongoUriReadOnly}
            style={isMongoUriReadOnly ? { opacity: 0.8, cursor: "not-allowed" } : undefined}
          />
          {hasDefault && item.value.trim() !== defaultValue && !isMongoUriReadOnly && (
            <button
              className="default-button-inline"
              type="button"
              onClick={() => updateItem(item.category, item.key, defaultValue)}
              title={`ใช้ค่าเริ่มต้น: ${defaultValue}`}
            >
              Default
            </button>
          )}
        </div>
        {item.description ? (
          <small>
            {isMongoUriReadOnly 
              ? "URI (ประกอบให้อัตโนมัติจากการกรอก Host, Port, User, Pass, Database)" 
              : item.description}
          </small>
        ) : null}
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
          <span>{checked ? "เปิดใช้งาน" : "ปิดใช้งาน"}</span>
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
            placeholder='{"mongodb":{"database":"..."} }'
            rows={12}
          />
          <div className="settings-button-row">
            <button className="secondary-button" type="button" onClick={() => setDialogMode(null)}>
              ยกเลิก
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
              <h2>เปลี่ยนรหัสผ่าน Setup</h2>
            </div>
            <button className="icon-button dialog-close" type="button" onClick={() => setDialogMode(null)} aria-label={t(language, "close")}>
              <X aria-hidden="true" size={18} />
            </button>
          </div>
          <label className="field-group">
            <span>รหัสผ่านเดิม</span>
            <div className="input-shell">
              <LockKeyhole aria-hidden="true" size={18} />
              <input value={currentSetupPassword} onChange={(event) => setCurrentSetupPassword(event.target.value)} type="password" />
            </div>
          </label>
          <label className="field-group">
            <span>รหัสผ่านใหม่</span>
            <div className="input-shell">
              <KeyRound aria-hidden="true" size={18} />
              <input value={newSetupPassword} onChange={(event) => setNewSetupPassword(event.target.value)} type="password" />
            </div>
          </label>
          <div className="settings-button-row">
            <button className="secondary-button" type="button" onClick={() => setDialogMode(null)}>
              ยกเลิก
            </button>
            <button className="primary-button settings-primary" type="submit" disabled={loadingAction === "change-password"}>
              {loadingAction === "change-password" ? <Loader2 className="spin" size={17} /> : <Save aria-hidden="true" size={17} />}
              <span>บันทึก</span>
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

function responseToTestResult(data: SetupResponse): TestResult {
  const success = data.success === true;
  return {
    status: success ? "success" : "failed",
    message: data.message ?? (success ? "สำเร็จ" : "ไม่สำเร็จ"),
    latencyMs: typeof data.latency_ms === "number" ? data.latency_ms : undefined,
    errorCode: typeof data.error_code === "string" ? data.error_code : undefined,
    database: typeof data.database === "string" ? data.database : undefined,
  };
}

function formatTestMessage(result: TestResult): string {
  return `${result.message}${result.latencyMs != null ? ` (${result.latencyMs}ms)` : ""}`;
}

function categoryIcon(category: string): ReactNode {
  switch (category) {
    case "mongodb":
      return <Database aria-hidden="true" size={19} />;
    case "postgresql":
      return <Table2 aria-hidden="true" size={19} />;
    case "clickhouse":
      return <HardDrive aria-hidden="true" size={19} />;
    case "kafka":
      return <Network aria-hidden="true" size={19} />;
    case "storage":
      return <Cloud aria-hidden="true" size={19} />;
    default:
      return <SettingsIcon aria-hidden="true" size={19} />;
  }
}

function TestBadge({ result }: { result: TestResult }) {
  return <span className={`test-badge ${result.status}`}>{result.status === "success" ? "ผ่าน" : result.status === "testing" ? "กำลังทดสอบ" : "ไม่ผ่าน"}</span>;
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

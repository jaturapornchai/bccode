"use client";

import {
  AlertCircle,
  Building2,
  CheckCircle2,
  Copy,
  ExternalLink,
  Eye,
  EyeOff,
  Loader2,
  LockKeyhole,
  MessageCircle,
  Server,
  Settings as SettingsIcon,
  ShieldCheck,
  UserPlus,
  UserRound,
} from "lucide-react";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import QRCode from "qrcode";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { runtimeGoApiUrlForOrigin } from "@/lib/backend-url";
import { persistLanguagePreferenceCookies } from "@/lib/backend-language-preload";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
import { isLocalLoginHost, LOCAL_GOOGLE_TEST_EMAIL } from "@/lib/local-dev-auth";
import { LanguageDialog } from "./language-dialog";
import { ThemeToggle } from "./theme-toggle";

type LoginState = "idle" | "loading" | "success" | "error";
type SignUpState = "idle" | "loading" | "success" | "error";
type ConnectionState = "idle" | "testing" | "success" | "error";
type ProviderLoginState = "idle" | "google" | "line" | "local-google";
type AuthMethod = "password" | "google" | "line";
type RuntimeMode = {
  ready: boolean;
  sameServerBackend: boolean;
};
type SocialLoginResponse = {
  success?: boolean;
  status?: "pending" | "success" | "failed" | "expired";
  message?: string;
  token?: string;
  refresh?: string;
  backendUrl?: string;
  user?: {
    username?: string;
    email?: string;
    name?: string;
    pictureUrl?: string;
  };
};
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
type SignUpForm = {
  name: string;
  username: string;
  password: string;
  confirmPassword: string;
};

const SOCIAL_POLL_TIMEOUT_MS = 5 * 60 * 1000;
const DEFAULT_SOCIAL_POLL_INTERVAL_MS = 2000;
const storageKeys = {
  backendUrl: "backend_url",
  backendUrlHistory: "backend_url_history",
  language: "user_language",
  username: "saved_username",
  legacyPassword: "saved_password",
  legacyRememberPassword: "remember_password",
  rememberUsername: "remember_username",
  auth: "bc_auth",
};

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

export function LoginScreen() {
  const router = useRouter();
  const [backendUrl, setBackendUrl] = useState("");
  const [language, setLanguage] = useState<LanguageCode>("th");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [rememberUsername, setRememberUsername] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [loginState, setLoginState] = useState<LoginState>("idle");
  const [signUpState, setSignUpState] = useState<SignUpState>("idle");
  const [connectionState, setConnectionState] = useState<ConnectionState>("idle");
  const [connectionMessage, setConnectionMessage] = useState("");
  const [providerLoginState, setProviderLoginState] = useState<ProviderLoginState>("idle");
  const [runtimeMode, setRuntimeMode] = useState<RuntimeMode>({ ready: false, sameServerBackend: false });
  const [devGoogleLoginEnabled, setDevGoogleLoginEnabled] = useState(false);
  const [isLocalTestHost, setIsLocalTestHost] = useState(false);
  const [lineDialog, setLineDialog] = useState<LineDialogState>(emptyLineDialog);
  const [signUpOpen, setSignUpOpen] = useState(false);
  const [signUpForm, setSignUpForm] = useState<SignUpForm>({
    name: "",
    username: "",
    password: "",
    confirmPassword: "",
  });
  const [message, setMessage] = useState("");
  const googlePollTimer = useRef<number | null>(null);
  const linePollTimer = useRef<number | null>(null);
  const autoConnectionTested = useRef(false);

  const canSubmit = useMemo(() => {
    return backendUrl.trim().length > 0 &&
      username.trim().length > 0 &&
      password.length > 0 &&
      loginState !== "loading" &&
      providerLoginState === "idle";
  }, [backendUrl, username, password, loginState, providerLoginState]);

  const canSignUp = useMemo(() => {
    return backendUrl.trim().length > 0 &&
      signUpForm.username.trim().length > 0 &&
      signUpForm.password.length > 0 &&
      signUpForm.confirmPassword.length > 0 &&
      signUpState !== "loading" &&
      loginState !== "loading" &&
      providerLoginState === "idle";
  }, [backendUrl, loginState, providerLoginState, signUpForm.confirmPassword, signUpForm.password, signUpForm.username, signUpState]);

  useEffect(() => {
    setRuntimeMode({ ready: true, sameServerBackend: isPublicRuntimeHost() });
    const savedLanguage = normalizeLanguage(localStorage.getItem(storageKeys.language) ?? "th");
    const savedUsername = localStorage.getItem(storageKeys.username);
    const savedRemember =
      localStorage.getItem(storageKeys.rememberUsername) === "true" ||
      localStorage.getItem(storageKeys.legacyRememberPassword) === "true";

    void loadLocalTestLoginAvailability();
    setBackendUrl(runtimeBackendUrlForCurrentPage());

    setLanguage(savedLanguage);
    setUsername(savedUsername ?? "");
    setRememberUsername(savedRemember);
    localStorage.removeItem(storageKeys.backendUrl);
    localStorage.removeItem(storageKeys.backendUrlHistory);
    localStorage.removeItem(storageKeys.legacyPassword);
    localStorage.removeItem(storageKeys.legacyRememberPassword);

    return () => {
      stopGooglePolling();
      stopLinePolling();
    };
  }, []);

  useEffect(() => {
    document.documentElement.lang = language;
    localStorage.setItem(storageKeys.language, language);
    persistLanguagePreferenceCookies(language, backendUrl);
  }, [backendUrl, language]);

  useEffect(() => {
    if (!backendUrl.trim() || autoConnectionTested.current) return;

    const timer = window.setTimeout(() => {
      autoConnectionTested.current = true;
      void (async () => {
        const targetBackendUrl = runtimeBackendUrlForCurrentPage();
        if (targetBackendUrl !== backendUrl) {
          setBackendUrl(targetBackendUrl);
        }

        setConnectionState("testing");
        setConnectionMessage("");

        try {
          const response = await fetch("/api/backend/check", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ backendUrl: targetBackendUrl }),
          });
          const data = (await response.json()) as { success?: boolean; message?: string };
          if (!response.ok || !data.success) throw new Error(data.message ?? t(language, "connectionFailed"));
          setConnectionState("success");
          setConnectionMessage(t(language, "connectionSuccess"));
        } catch (error) {
          setConnectionState("error");
          setConnectionMessage(error instanceof Error ? error.message : t(language, "connectionFailed"));
        }
      })();
    }, 350);

    return () => window.clearTimeout(timer);
  }, [backendUrl, language]);

  async function loadLocalTestLoginAvailability() {
    const isLocalHost = isLocalLoginHost(window.location.hostname);
    if (!isLocalHost && !isPublicRuntimeHost()) {
      setDevGoogleLoginEnabled(false);
      setIsLocalTestHost(false);
      return;
    }

    try {
      const response = await fetch("/api/auth/google/dev-login", { cache: "no-store" });
      const data = (await response.json()) as { enabled?: boolean };
      const enabled = response.ok && data.enabled === true;
      setDevGoogleLoginEnabled(enabled);
      setIsLocalTestHost(isLocalHost && enabled);
    } catch {
      setDevGoogleLoginEnabled(false);
      setIsLocalTestHost(false);
    }
  }

  function stopGooglePolling() {
    if (googlePollTimer.current !== null) {
      window.clearInterval(googlePollTimer.current);
      googlePollTimer.current = null;
    }
  }

  function stopLinePolling() {
    if (linePollTimer.current !== null) {
      window.clearInterval(linePollTimer.current);
      linePollTimer.current = null;
    }
  }

  async function handleGoogleLogin() {
    if (!backendUrl.trim()) return setMessage(t(language, "enterBackendUrl"));

    stopGooglePolling();
    setProviderLoginState("google");
    setLoginState("loading");
    setMessage(t(language, "googleLoginOpening"));

    try {
      if (devGoogleLoginEnabled && (await tryDevGoogleLoginFallback("", true))) return;

      const response = await fetch("/api/auth/google/session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ deviceInfo: navigator.userAgent }),
      });
      const data = (await response.json()) as {
        success?: boolean;
        message?: string;
        sessionId?: string;
        loginUrl?: string;
        pollInterval?: number;
      };

      if (!response.ok || !data.success || !data.sessionId || !data.loginUrl) {
        const errorMessage = data.message ?? t(language, "loginFailed");
        if (await tryDevGoogleLoginFallback(errorMessage)) return;
        throw new Error(errorMessage);
      }

      const popup = window.open(data.loginUrl, "bc-google-login", "popup=yes,width=480,height=720");
      if (!popup) {
        throw new Error(t(language, "popupBlocked"));
      }

      const startedAt = Date.now();
      googlePollTimer.current = window.setInterval(() => {
        if (Date.now() - startedAt > SOCIAL_POLL_TIMEOUT_MS) {
          stopGooglePolling();
          setProviderLoginState("idle");
          setLoginState("error");
          setMessage(t(language, "googleLoginTimeout"));
          return;
        }
        void pollGoogleLogin(data.sessionId!);
      }, Math.max(data.pollInterval ?? DEFAULT_SOCIAL_POLL_INTERVAL_MS, 1000));
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t(language, "loginFailed");
      if (await tryDevGoogleLoginFallback(errorMessage)) return;
      stopGooglePolling();
      setProviderLoginState("idle");
      setLoginState("error");
      setMessage(errorMessage);
    }
  }

  async function tryDevGoogleLoginFallback(reason: string, force = false): Promise<boolean> {
    if (!force && !reason.includes("Google login service")) return false;

    try {
      const targetBackendUrl = runtimeBackendUrlForCurrentPage();
      if (targetBackendUrl !== backendUrl) {
        setBackendUrl(targetBackendUrl);
      }

      const response = await fetch("/api/auth/google/dev-login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl: targetBackendUrl }),
      });
      const data = (await response.json()) as SocialLoginResponse;
      if (!response.ok || !data.success || !data.token) return false;

      const nextUsername = data.user?.username || data.user?.email || data.user?.name || LOCAL_GOOGLE_TEST_EMAIL;
      persistLogin(data.backendUrl ?? targetBackendUrl, nextUsername, data.token, data.refresh ?? "", "google", data.user);
      setLoginState("success");
      setProviderLoginState("idle");
      setMessage(t(language, "loginSuccess"));
      router.push("/workspace");
      return true;
    } catch {
      return false;
    }
  }

  async function pollGoogleLogin(sessionId: string) {
    const response = await fetch("/api/auth/google/status", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backendUrl, sessionId }),
    });
    const data = (await response.json()) as SocialLoginResponse;

    if (!response.ok || data.success === false || data.status === "failed" || data.status === "expired") {
      stopGooglePolling();
      setProviderLoginState("idle");
      setLoginState("error");
      setMessage(data.message ?? t(language, "loginFailed"));
      return;
    }

    if (data.status !== "success" || !data.token) return;

    const nextUsername = data.user?.username || data.user?.email || data.user?.name || "google";
    persistLogin(data.backendUrl ?? backendUrl, nextUsername, data.token, data.refresh ?? "", "google", data.user);
    stopGooglePolling();
    setProviderLoginState("idle");
    setLoginState("success");
    setMessage(t(language, "loginSuccess"));
    router.push("/workspace");
  }

  async function handleLocalGoogleTestLogin() {
    if (!backendUrl.trim()) return setMessage(t(language, "enterBackendUrl"));

    setProviderLoginState("local-google");
    setLoginState("loading");
    setMessage("");

    try {
      const response = await fetch("/api/auth/google/dev-login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl }),
      });
      const data = (await response.json()) as SocialLoginResponse;

      if (!response.ok || !data.success || !data.token) {
        throw new Error(data.message ?? t(language, "loginFailed"));
      }

      const nextUsername = data.user?.username || data.user?.email || data.user?.name || LOCAL_GOOGLE_TEST_EMAIL;
      persistLogin(data.backendUrl ?? backendUrl, nextUsername, data.token, data.refresh ?? "", "google", data.user);
      setProviderLoginState("idle");
      setLoginState("success");
      setMessage(t(language, "loginSuccess"));
      router.push("/workspace");
    } catch (error) {
      setProviderLoginState("idle");
      setLoginState("error");
      setMessage(error instanceof Error ? error.message : t(language, "loginFailed"));
    }
  }

  async function handleLineLogin() {
    if (!backendUrl.trim()) return setMessage(t(language, "enterBackendUrl"));

    stopLinePolling();
    setProviderLoginState("line");
    setLoginState("loading");
    setMessage("");
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
        throw new Error(data.message ?? t(language, "loginFailed"));
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
      setProviderLoginState("idle");
      setLoginState("error");
      setLineDialog((current) => ({
        ...current,
        loading: false,
        error: error instanceof Error ? error.message : t(language, "loginFailed"),
      }));
    }
  }

  function startLinePolling(code: string, expiresAt: string) {
    const startedAt = Date.now();
    linePollTimer.current = window.setInterval(() => {
      const isExpiredByTime = expiresAt ? Date.now() > Date.parse(expiresAt) : false;
      if (Date.now() - startedAt > SOCIAL_POLL_TIMEOUT_MS || isExpiredByTime) {
        stopLinePolling();
        setProviderLoginState("idle");
        setLoginState("error");
        setLineDialog((current) => ({ ...current, expired: true }));
        setMessage(t(language, "lineLoginExpired"));
        return;
      }
      void pollLineLogin(code);
    }, DEFAULT_SOCIAL_POLL_INTERVAL_MS);
  }

  async function pollLineLogin(code: string) {
    const response = await fetch("/api/auth/line/status", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backendUrl, code }),
    });
    const data = (await response.json()) as SocialLoginResponse;

    if (!response.ok || data.success === false || data.status === "failed" || data.status === "expired") {
      stopLinePolling();
      setProviderLoginState("idle");
      setLoginState("error");
      setMessage(data.message ?? t(language, "loginFailed"));
      setLineDialog((current) => ({ ...current, error: data.message ?? t(language, "loginFailed") }));
      return;
    }

    if (data.status !== "success" || !data.token) return;

    const nextUsername = data.user?.username || data.user?.email || data.user?.name || "line";
    persistLogin(data.backendUrl ?? backendUrl, nextUsername, data.token, data.refresh ?? "", "line", data.user);
    stopLinePolling();
    setProviderLoginState("idle");
    setLoginState("success");
    setMessage(t(language, "loginSuccess"));
    setLineDialog(emptyLineDialog);
    router.push("/workspace");
  }

  function closeLineDialog() {
    stopLinePolling();
    setProviderLoginState("idle");
    if (loginState === "loading") setLoginState("idle");
    setLineDialog(emptyLineDialog);
  }

  async function copyLineLoginUrl() {
    if (!lineDialog.loginUrl) return;
    await navigator.clipboard.writeText(lineDialog.loginUrl);
    setMessage(t(language, "copied"));
  }

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!backendUrl.trim()) return setMessage(t(language, "enterBackendUrl"));
    if (!username.trim()) return setMessage(t(language, "enterUsername"));
    if (!password) return setMessage(t(language, "enterPassword"));

    setLoginState("loading");
    setMessage("");

    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl, username, password }),
      });
      const data = (await response.json()) as {
        success?: boolean;
        message?: string;
        token?: string;
        refresh?: string;
        user?: string;
        backendUrl?: string;
      };

      if (!response.ok || !data.success || !data.token) {
        throw new Error(data.message ?? t(language, "loginFailed"));
      }

      persistLogin(data.backendUrl ?? backendUrl, username.trim(), data.token, data.refresh ?? "", "password");
      setLoginState("success");
      setMessage(t(language, "loginSuccess"));
      router.push("/workspace");
    } catch (error) {
      setLoginState("error");
      setMessage(error instanceof Error ? error.message : t(language, "loginFailed"));
    }
  }

  async function handleSignUp(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!backendUrl.trim()) return setMessage(t(language, "enterBackendUrl"));

    const nextUsername = signUpForm.username.trim();
    if (!nextUsername) return setMessage(t(language, "enterUsername"));
    if (!signUpForm.password) return setMessage(t(language, "enterPassword"));
    if (signUpForm.password !== signUpForm.confirmPassword) {
      setSignUpState("error");
      setMessage(t(language, "passwordMismatch"));
      return;
    }

    setSignUpState("loading");
    setLoginState("loading");
    setMessage("");

    try {
      const response = await fetch("/api/auth/register-username", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          backendUrl,
          username: nextUsername,
          password: signUpForm.password,
          name: signUpForm.name.trim(),
        }),
      });
      const data = (await response.json()) as { success?: boolean; message?: string; backendUrl?: string };

      if (!response.ok || !data.success) {
        throw new Error(data.message ?? t(language, "registerFailed"));
      }

      const loginResponse = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl: data.backendUrl ?? backendUrl, username: nextUsername, password: signUpForm.password }),
      });
      const loginData = (await loginResponse.json()) as {
        success?: boolean;
        message?: string;
        token?: string;
        refresh?: string;
        backendUrl?: string;
      };

      if (!loginResponse.ok || !loginData.success || !loginData.token) {
        throw new Error(loginData.message ?? t(language, "loginFailed"));
      }

      persistLogin(loginData.backendUrl ?? data.backendUrl ?? backendUrl, nextUsername, loginData.token, loginData.refresh ?? "", "password");
      setUsername(nextUsername);
      setPassword("");
      setSignUpState("success");
      setLoginState("success");
      setSignUpOpen(false);
      setSignUpForm({ name: "", username: "", password: "", confirmPassword: "" });
      setMessage(t(language, "registerSuccess"));
      router.push("/workspace");
    } catch (error) {
      setSignUpState("error");
      setLoginState("error");
      setMessage(error instanceof Error ? error.message : t(language, "registerFailed"));
    }
  }

  async function handleConnectionTest() {
    await runConnectionTest({ automatic: false });
  }

  async function runConnectionTest(options: { automatic: boolean }) {
    const targetBackendUrl = runtimeBackendUrlForCurrentPage();
    if (targetBackendUrl !== backendUrl) {
      setBackendUrl(targetBackendUrl);
    }

    setConnectionState("testing");
    setConnectionMessage("");
    if (!options.automatic) setMessage("");

    try {
      const response = await fetch("/api/backend/check", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl: targetBackendUrl }),
      });
      const data = (await response.json()) as { success?: boolean; message?: string };
      if (!response.ok || !data.success) throw new Error(data.message ?? t(language, "connectionFailed"));
      setConnectionState("success");
      setConnectionMessage(t(language, "connectionSuccess"));
    } catch (error) {
      setConnectionState("error");
      const errMsg = error instanceof Error ? error.message : t(language, "connectionFailed");
      setConnectionMessage(errMsg);
    }
  }

  function persistLogin(
    nextBackendUrl: string,
    nextUsername: string,
    token: string,
    refresh: string,
    method: AuthMethod,
    profile?: SocialLoginResponse["user"],
  ) {
    const runtimeBackendUrl = runtimeBackendUrlForCurrentPage(nextBackendUrl);
    localStorage.setItem(storageKeys.rememberUsername, String(rememberUsername));
    if (rememberUsername) {
      localStorage.setItem(storageKeys.username, nextUsername);
    } else {
      localStorage.removeItem(storageKeys.username);
    }
    localStorage.removeItem(storageKeys.legacyPassword);
    localStorage.removeItem(storageKeys.backendUrl);
    localStorage.removeItem(storageKeys.backendUrlHistory);
    localStorage.setItem(
      storageKeys.auth,
      JSON.stringify({ token, refresh, username: nextUsername, backendUrl: runtimeBackendUrl, method, profile: profile ?? null }),
    );
  }

  return (
    <main className="login-shell">
      <section className="brand-panel" aria-label="BC Ai Account">
        <div className="brand-badge-row">
          <div className="brand-mark">AI</div>
          <div>
            <p className="eyebrow">{t(language, "brandEyebrow")}</p>
            <strong>{t(language, "secureWorkspace")}</strong>
          </div>
        </div>
        <div className="brand-copy">
          <h1>{t(language, "loginTitle")}</h1>
          <p>{t(language, "brandDescription")}</p>
        </div>
        <div className="brand-feature-grid" aria-label={t(language, "firstUseTitle")}>
          <div className="brand-feature-card">
            <Building2 aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "firstUseTitle")}</strong>
              <span>{t(language, "firstUseDescription")}</span>
            </div>
          </div>
          <div className="brand-feature-card">
            <ShieldCheck aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "secureWorkspace")}</strong>
              <span>{t(language, "multiCompanyDescription")}</span>
            </div>
          </div>
          <div className="brand-feature-card">
            <MessageCircle aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "loginWithLine")}</strong>
              <span>{t(language, "lineLoginDescription")}</span>
            </div>
          </div>
        </div>
        <div className="status-strip" aria-label="System status">
          <div>
            <span>{t(language, "api")}</span>
            <strong>{backendUrl || "-"}</strong>
          </div>
          <ShieldCheck aria-hidden="true" size={22} />
        </div>
      </section>

      <section className="form-panel" aria-label="Login form">
        <form className="login-card" onSubmit={handleLogin}>
          <div className="card-header">
            <div>
              <p className="eyebrow">{t(language, "secureWorkspace")}</p>
              <h2>{t(language, "signInTitle")}</h2>
            </div>
            <div className="header-actions">
              <ThemeToggle language={language} />
              <LanguageDialog language={language} onLanguageChange={setLanguage} />
              <Link className="icon-button settings-link" href="/settings" title={t(language, "settings")} aria-label={t(language, "settings")}>
                <SettingsIcon aria-hidden="true" size={18} />
              </Link>
            </div>
          </div>

          {runtimeMode.ready && !runtimeMode.sameServerBackend ? (
            <label className="field-group">
              <span>{t(language, "backendUrl")}</span>
              <div className="input-shell">
                <Server aria-hidden="true" size={18} />
                <input
                  value={backendUrl}
                  readOnly
                  aria-readonly="true"
                  placeholder="/backend/goapi"
                />
                <button
                  className="text-action"
                  type="button"
                  onClick={handleConnectionTest}
                  disabled={connectionState === "testing"}
                >
                  {connectionState === "testing" ? <Loader2 className="spin" size={16} /> : t(language, "testConnection")}
                </button>
              </div>
              {connectionMessage && (
                <div
                  className={`connection-status-msg ${connectionState === "success" ? "success" : "error"}`}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "6px",
                    fontSize: "0.82rem",
                    fontWeight: "bold",
                    marginTop: "2px",
                    color: connectionState === "success" ? "var(--success)" : "var(--danger)",
                  }}
                >
                  {connectionState === "success" ? (
                    <CheckCircle2 size={14} />
                  ) : (
                    <AlertCircle size={14} />
                  )}
                  <span>{connectionMessage}</span>
                </div>
              )}
            </label>
          ) : null}

          <label className="field-group">
            <span>{t(language, "username")}</span>
            <div className="input-shell">
              <UserRound aria-hidden="true" size={18} />
              <input
                autoComplete="username"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                placeholder={t(language, "usernamePlaceholder")}
              />
            </div>
          </label>

          <label className="field-group">
            <span>{t(language, "password")}</span>
            <div className="input-shell">
              <LockKeyhole aria-hidden="true" size={18} />
              <input
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder={t(language, "password")}
                type={showPassword ? "text" : "password"}
              />
              <button
                className="icon-button"
                type="button"
                onClick={() => setShowPassword((current) => !current)}
                aria-label={showPassword ? t(language, "hidePassword") : t(language, "showPassword")}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </label>

          <div className="form-row">
            <label className="check-row">
              <input
                checked={rememberUsername}
                onChange={(event) => setRememberUsername(event.target.checked)}
                type="checkbox"
              />
              <span>{t(language, "rememberUsername")}</span>
            </label>
          </div>

          {message ? (
            <div className={`message ${loginState === "success" || connectionState === "success" ? "success" : "error"}`}>
              {loginState === "success" || connectionState === "success" ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
              <span>{message}</span>
            </div>
          ) : null}

          <button className="primary-button" type="submit" disabled={!canSubmit}>
            {loginState === "loading" ? <Loader2 className="spin" size={18} /> : <LockKeyhole size={18} />}
            <span>{loginState === "loading" ? t(language, "loggingIn") : t(language, "login")}</span>
          </button>

          <div className="social-login-separator">
            <span>{t(language, "socialLoginSeparator")}</span>
          </div>

          <div className="social-login-grid">
            <button
              className="social-login-button google-login"
              type="button"
              onClick={handleGoogleLogin}
              disabled={providerLoginState !== "idle" || loginState === "loading"}
            >
              {providerLoginState === "google" ? (
                <Loader2 className="spin" aria-hidden="true" size={20} />
              ) : (
                <Image alt="" height={20} src="/google_logo.png" width={20} />
              )}
              <span>{t(language, "loginWithGoogle")}</span>
            </button>
            {isLocalTestHost ? (
              <button
                className="social-login-button local-google-test-login"
                type="button"
                onClick={handleLocalGoogleTestLogin}
                disabled={providerLoginState !== "idle" || loginState === "loading"}
              >
                {providerLoginState === "local-google" ? (
                  <Loader2 className="spin" aria-hidden="true" size={20} />
                ) : (
                  <Image alt="" height={20} src="/google_logo.png" width={20} />
                )}
                <span>{t(language, "localGoogleTestLogin")}</span>
              </button>
            ) : null}
            <button
              className="social-login-button line-login"
              type="button"
              onClick={handleLineLogin}
              disabled={providerLoginState !== "idle" || loginState === "loading"}
            >
              {providerLoginState === "line" ? (
                <Loader2 className="spin" aria-hidden="true" size={20} />
              ) : (
                <MessageCircle aria-hidden="true" size={20} />
              )}
              <span>{t(language, "loginWithLine")}</span>
            </button>
          </div>

          <div className="first-use-note">
            <Building2 aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "firstUseTitle")}</strong>
              <span>{t(language, "firstUseDescription")}</span>
              <small>{t(language, "multiCompanyDescription")}</small>
            </div>
          </div>

          <button
            className="primary-button signup-open-button email-signup-button"
            type="button"
            onClick={handleGoogleLogin}
            disabled={loginState === "loading" || providerLoginState !== "idle"}
          >
            {providerLoginState === "google" ? <Loader2 className="spin" aria-hidden="true" size={18} /> : <UserPlus aria-hidden="true" size={18} />}
            <span>{t(language, "signUpWithEmail")}</span>
          </button>
        </form>

        {signUpOpen ? (
          <div className="dialog-backdrop" role="presentation">
            <section className="line-login-dialog signup-dialog" aria-label={t(language, "signUp")} role="dialog" aria-modal="true">
              <div className="dialog-header">
                <div>
                  <p className="eyebrow">{t(language, "secureWorkspace")}</p>
                  <h2>{t(language, "signUp")}</h2>
                </div>
                <button className="icon-button dialog-close" type="button" onClick={() => setSignUpOpen(false)} aria-label={t(language, "close")}>
                  ×
                </button>
              </div>

              <p className="line-login-description">{t(language, "signUpDescription")}</p>

              <form className="signup-form" onSubmit={handleSignUp}>
                <label className="field-group">
                  <span>{t(language, "displayName")}</span>
                  <div className="input-shell">
                    <UserRound aria-hidden="true" size={18} />
                    <input
                      autoComplete="name"
                      value={signUpForm.name}
                      onChange={(event) => setSignUpForm((current) => ({ ...current, name: event.target.value }))}
                      placeholder={t(language, "displayName")}
                    />
                  </div>
                </label>

                <label className="field-group">
                  <span>{t(language, "username")}</span>
                  <div className="input-shell">
                    <UserRound aria-hidden="true" size={18} />
                    <input
                      autoComplete="username"
                      value={signUpForm.username}
                      onChange={(event) => setSignUpForm((current) => ({ ...current, username: event.target.value }))}
                      placeholder={t(language, "usernamePlaceholder")}
                    />
                  </div>
                </label>

                <label className="field-group">
                  <span>{t(language, "password")}</span>
                  <div className="input-shell">
                    <LockKeyhole aria-hidden="true" size={18} />
                    <input
                      autoComplete="new-password"
                      value={signUpForm.password}
                      onChange={(event) => setSignUpForm((current) => ({ ...current, password: event.target.value }))}
                      placeholder={t(language, "password")}
                      type="password"
                    />
                  </div>
                </label>

                <label className="field-group">
                  <span>{t(language, "confirmPassword")}</span>
                  <div className="input-shell">
                    <LockKeyhole aria-hidden="true" size={18} />
                    <input
                      autoComplete="new-password"
                      value={signUpForm.confirmPassword}
                      onChange={(event) => setSignUpForm((current) => ({ ...current, confirmPassword: event.target.value }))}
                      placeholder={t(language, "confirmPassword")}
                      type="password"
                    />
                  </div>
                </label>

                <div className="line-dialog-actions">
                  <button className="primary-button" type="submit" disabled={!canSignUp}>
                    {signUpState === "loading" ? <Loader2 className="spin" size={18} /> : <UserPlus size={18} />}
                    <span>{signUpState === "loading" ? t(language, "creatingAccount") : t(language, "createAccount")}</span>
                  </button>
                  <button className="secondary-button" type="button" onClick={() => setSignUpOpen(false)} disabled={signUpState === "loading"}>
                    {t(language, "close")}
                  </button>
                </div>
              </form>
            </section>
          </div>
        ) : null}

        {lineDialog.open ? (
          <div className="dialog-backdrop" role="presentation">
            <section className="line-login-dialog" aria-label={t(language, "lineLoginTitle")} role="dialog" aria-modal="true">
              <div className="dialog-header">
                <div>
                  <p className="eyebrow">{t(language, "loginWithLine")}</p>
                  <h2>{t(language, "lineLoginTitle")}</h2>
                </div>
                <button className="icon-button dialog-close" type="button" onClick={closeLineDialog} aria-label={t(language, "lineLoginClose")}>
                  ×
                </button>
              </div>

              {lineDialog.loading ? (
                <div className="line-login-status">
                  <Loader2 className="spin" aria-hidden="true" size={28} />
                  <span>{t(language, "lineLoginWaiting")}</span>
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
                  <p className="line-login-description">{t(language, "lineLoginDescription")}</p>
                  <div className="line-qr-box">
                    {lineDialog.qrDataUrl ? <Image alt={t(language, "lineLoginTitle")} height={220} src={lineDialog.qrDataUrl} unoptimized width={220} /> : null}
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
                    <span>{t(language, "lineLoginWaiting")}</span>
                  </div>
                </>
              )}

              <div className="line-dialog-actions">
                <button className="secondary-button" type="button" onClick={handleLineLogin}>
                  {t(language, "lineLoginCreateNew")}
                </button>
                <button className="secondary-button" type="button" onClick={closeLineDialog}>
                  {t(language, "lineLoginClose")}
                </button>
              </div>
            </section>
          </div>
        ) : null}
      </section>
    </main>
  );
}

function runtimeBackendUrlForCurrentPage(fallback = ""): string {
  if (typeof window === "undefined") return fallback;
  return runtimeGoApiUrlForOrigin(window.location.origin);
}

function isPublicRuntimeHost(): boolean {
  return typeof window !== "undefined" && !window.location.hostname.includes("localhost") && !window.location.hostname.includes("127.0.0.1");
}

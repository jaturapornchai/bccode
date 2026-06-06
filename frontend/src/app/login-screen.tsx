"use client";

import {
  AlertCircle,
  Building2,
  CheckCircle2,
  Eye,
  EyeOff,
  Loader2,
  LockKeyhole,
  Server,
  Settings as SettingsIcon,
  ShieldCheck,
  UserRound,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { runtimeGoApiUrlForOrigin } from "@/lib/backend-url";
import { persistLanguagePreferenceCookies } from "@/lib/backend-language-preload";
import { isValidHoldingCode, normalizeHoldingCode } from "@/lib/holding-code";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
import { isLocalLoginHost, LOCAL_GOOGLE_TEST_EMAIL } from "@/lib/local-dev-auth";
import { LanguageDialog } from "./language-dialog";
import { ThemeToggle } from "./theme-toggle";

type LoginState = "idle" | "loading" | "success" | "error";
type ConnectionState = "idle" | "testing" | "success" | "error";
type ProviderLoginState = "idle" | "google" | "local-google";
type AuthMethod = "password" | "google";
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
  holdingCode: "saved_holdingcode",
  auth: "bc_auth",
};

export function LoginScreen() {
  const router = useRouter();
  const [mounted, setMounted] = useState(false);
  const [backendUrl, setBackendUrl] = useState("");
  const [language, setLanguage] = useState<LanguageCode>("th");
  const [holdingCode, setHoldingCode] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [rememberUsername, setRememberUsername] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [loginState, setLoginState] = useState<LoginState>("idle");
  const [connectionState, setConnectionState] = useState<ConnectionState>("idle");
  const [connectionMessage, setConnectionMessage] = useState("");
  const [providerLoginState, setProviderLoginState] = useState<ProviderLoginState>("idle");
  const [runtimeMode, setRuntimeMode] = useState<RuntimeMode>({ ready: false, sameServerBackend: false });
  const [devGoogleLoginEnabled, setDevGoogleLoginEnabled] = useState(false);
  const [isLocalTestHost, setIsLocalTestHost] = useState(false);
  const [message, setMessage] = useState("");
  const googlePollTimer = useRef<number | null>(null);
  const autoConnectionTested = useRef(false);

  const canSubmit = useMemo(() => {
    return backendUrl.trim().length > 0 &&
      username.trim().length > 0 &&
      password.length > 0 &&
      loginState !== "loading" &&
      providerLoginState === "idle";
  }, [backendUrl, holdingCode, username, password, loginState, providerLoginState]);

  useEffect(() => {
    setMounted(true);
    setRuntimeMode({ ready: true, sameServerBackend: isPublicRuntimeHost() });
    const savedLanguage = normalizeLanguage(localStorage.getItem(storageKeys.language) ?? "th");
    const savedUsername = localStorage.getItem(storageKeys.username);
    const savedHoldingCode = localStorage.getItem(storageKeys.holdingCode);
    const savedRemember =
      localStorage.getItem(storageKeys.rememberUsername) === "true" ||
      localStorage.getItem(storageKeys.legacyRememberPassword) === "true";

    void loadLocalTestLoginAvailability();
    setBackendUrl(runtimeBackendUrlForCurrentPage());

    setLanguage(savedLanguage);
    setHoldingCode(savedHoldingCode ?? "");
    setUsername(savedUsername ?? "");
    setRememberUsername(savedRemember);
    localStorage.removeItem(storageKeys.backendUrl);
    localStorage.removeItem(storageKeys.backendUrlHistory);
    localStorage.removeItem(storageKeys.legacyPassword);
    localStorage.removeItem(storageKeys.legacyRememberPassword);

    return () => {
      stopGooglePolling();
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
      router.push("/holding");
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
    router.push("/holding");
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
      router.push("/holding");
    } catch (error) {
      setProviderLoginState("idle");
      setLoginState("error");
      setMessage(error instanceof Error ? error.message : t(language, "loginFailed"));
    }
  }

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!backendUrl.trim()) return setMessage(t(language, "enterBackendUrl"));
    const normalizedHoldingCode = normalizeHoldingCode(holdingCode);
    if (!normalizedHoldingCode) return setMessage(t(language, "enterHoldingCode"));
    if (!isValidHoldingCode(normalizedHoldingCode)) return setMessage(t(language, "holdingCodeInvalid"));
    if (!username.trim()) return setMessage(t(language, "enterUsername"));
    if (!password) return setMessage(t(language, "enterPassword"));

    setLoginState("loading");
    setMessage("");

    try {
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backendUrl, username, password, holdingcode: normalizedHoldingCode }),
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

      persistLogin(data.backendUrl ?? backendUrl, username.trim(), data.token, data.refresh ?? "", "password", undefined, normalizedHoldingCode);
      setLoginState("success");
      setMessage(t(language, "loginSuccess"));
      router.push("/workspace");
    } catch (error) {
      setLoginState("error");
      setMessage(error instanceof Error ? error.message : t(language, "loginFailed"));
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
    nextHoldingCode = "",
  ) {
    const runtimeBackendUrl = runtimeBackendUrlForCurrentPage(nextBackendUrl);
    localStorage.setItem(storageKeys.rememberUsername, String(rememberUsername));
    if (rememberUsername) {
      localStorage.setItem(storageKeys.username, nextUsername);
    } else {
      localStorage.removeItem(storageKeys.username);
    }
    if (nextHoldingCode) {
      localStorage.setItem(storageKeys.holdingCode, nextHoldingCode);
    } else {
      localStorage.removeItem(storageKeys.holdingCode);
    }
    localStorage.removeItem(storageKeys.legacyPassword);
    localStorage.removeItem(storageKeys.backendUrl);
    localStorage.removeItem(storageKeys.backendUrlHistory);
    localStorage.setItem(
      storageKeys.auth,
      JSON.stringify({
        token,
        refresh,
        username: nextUsername,
        backendUrl: runtimeBackendUrl,
        method,
        profile: profile ?? null,
        ...(nextHoldingCode ? { holdingcode: nextHoldingCode } : {}),
      }),
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
        <div className="brand-hero">
          <div className="brand-copy">
            <h1>{t(language, "loginTitle")}</h1>
            <p>{t(language, "brandDescription")}</p>
          </div>
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
            <LockKeyhole aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "passwordLoginSectionTitle")}</strong>
              <span>{t(language, "passwordLoginSectionDescription")}</span>
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

          <section className="password-login-section" aria-label={t(language, "passwordLoginSectionTitle")}>
            <div className="password-login-heading">
              <LockKeyhole aria-hidden="true" size={18} />
              <div>
                <strong>{t(language, "passwordLoginSectionTitle")}</strong>
                <span>{t(language, "passwordLoginSectionDescription")}</span>
              </div>
            </div>

            <div className="field-group holding-code-field">
              <span id="holding-code-label">{t(language, "holdingCode")}</span>
              <div className="holding-code-control-row">
                <div className="input-shell">
                  <Building2 aria-hidden="true" size={18} />
                  <input
                    aria-labelledby="holding-code-label"
                    autoComplete="organization"
                    value={holdingCode}
                    onChange={(event) => setHoldingCode(normalizeHoldingCode(event.target.value))}
                    placeholder="bcdemo01"
                  />
                </div>
              </div>
              <small className="field-help">{t(language, "holdingCodeHint")}</small>
            </div>

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

            <button className="primary-button" type="submit" disabled={!mounted || !canSubmit}>
              {loginState === "loading" ? <Loader2 className="spin" size={18} /> : <LockKeyhole size={18} />}
              <span>{loginState === "loading" ? t(language, "loggingIn") : t(language, "login")}</span>
            </button>
          </section>

          <section className="auth-login-section" aria-label={t(language, "authLoginSectionTitle")}>
            <div className="auth-login-heading">
              <ShieldCheck aria-hidden="true" size={18} />
              <strong>{t(language, "authLoginSectionTitle")}</strong>
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
            </div>
          </section>

          <div className="first-use-note">
            <Building2 aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "firstUseTitle")}</strong>
              <span>{t(language, "firstUseDescription")}</span>
              <ol className="first-use-steps">
                <li>{t(language, "firstUseStepGoogle")}</li>
                <li>{t(language, "firstUseStepHolding")}</li>
                <li>{t(language, "firstUseStepWorkspace")}</li>
              </ol>
              <small>{t(language, "createHoldingDescription")}</small>
              <small>{t(language, "multiCompanyDescription")}</small>
            </div>
          </div>

        </form>
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

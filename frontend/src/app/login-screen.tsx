"use client";

import {
  AlertCircle,
  Building2,
  CheckCircle2,
  Eye,
  EyeOff,
  Loader2,
  LockKeyhole,
  LogIn,
  Settings as SettingsIcon,
  ShieldCheck,
  Sparkles,
  UserRound,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { motion, AnimatePresence, type Variants } from "motion/react";
import { runtimeGoApiUrlForOrigin } from "@/lib/backend-url";
import { persistLanguagePreferenceCookies } from "@/lib/backend-language-preload";
import { isValidHoldingCode, normalizeHoldingCode } from "@/lib/holding-code";
import { normalizeLanguage, t, type LanguageCode } from "@/lib/i18n";
import { isLocalLoginHost, LOCAL_GOOGLE_TEST_EMAIL } from "@/lib/local-dev-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LanguageDialog } from "./language-dialog";
import { ThemeToggle } from "./theme-toggle";
import { FontPicker } from "./font-picker";

// Shared motion variants — subtle, premium, never cluttered.
// Reduced-motion is handled by the client-only <MotionConfig> in LoginWrapper.
// The login screen is mounted after hydration, so Motion can read the user's
// preference without creating a server/client style mismatch.
const EASE_OUT = [0.22, 1, 0.36, 1] as const;

const panelEnter = {
  initial: { opacity: 0, y: 16 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.5, ease: EASE_OUT },
} as const;

const staggerParent: Variants = {
  animate: { transition: { staggerChildren: 0.08, delayChildren: 0.15 } },
};

const staggerChild: Variants = {
  initial: { opacity: 0, y: 12 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.45, ease: EASE_OUT },
  },
};

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
      <motion.section
        className="brand-panel"
        aria-label="BC Ai Account"
        initial={panelEnter.initial}
        animate={panelEnter.animate}
        transition={panelEnter.transition}
      >
        <motion.div className="brand-badge-row" variants={staggerChild} {...staggerParent}>
          <div className="brand-mark" aria-hidden="true">
            <Sparkles size={22} strokeWidth={2.2} />
          </div>
          <div>
            <p className="eyebrow">{t(language, "brandEyebrow")}</p>
            <strong>{t(language, "secureWorkspace")}</strong>
          </div>
        </motion.div>
        <div className="brand-hero">
          <motion.div className="brand-copy" variants={staggerParent}>
            <motion.p
              className="brand-eyebrow-line"
              variants={staggerChild}
            >
              {t(language, "brandEyebrow")}
            </motion.p>
            <motion.h1 variants={staggerChild}>
              {t(language, "loginTitle")}
            </motion.h1>
            <motion.p className="brand-tagline" variants={staggerChild}>
              {t(language, "brandDescription")}
            </motion.p>
          </motion.div>
        </div>
        <motion.div
          className="brand-feature-grid"
          aria-label={t(language, "firstUseTitle")}
          variants={staggerParent}
        >
          <motion.div className="brand-feature-card" variants={staggerChild}>
            <Building2 aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "firstUseTitle")}</strong>
              <span>{t(language, "firstUseDescription")}</span>
            </div>
          </motion.div>
          <motion.div className="brand-feature-card" variants={staggerChild}>
            <ShieldCheck aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "secureWorkspace")}</strong>
              <span>{t(language, "multiCompanyDescription")}</span>
            </div>
          </motion.div>
        </motion.div>
        <motion.div
          className="brand-status-strip"
          variants={staggerChild}
          aria-label="System status"
          title={backendUrl ? `${t(language, "api")}: ${backendUrl}` : undefined}
        >
          <span className="brand-status-dot" data-state={connectionState} aria-hidden="true" />
          <span>{backendUrl ? t(language, "connectionSuccess") : "-"}</span>
        </motion.div>
      </motion.section>

      <motion.section
        className="form-panel"
        aria-label="Login form"
        initial={panelEnter.initial}
        animate={panelEnter.animate}
        transition={{ ...panelEnter.transition, delay: 0.1 }}
      >
        <form className="login-card" onSubmit={handleLogin}>
          <motion.div className="card-header" variants={staggerParent}>
            <motion.div variants={staggerChild}>
              <p className="eyebrow">{t(language, "secureWorkspace")}</p>
              <h2>{t(language, "signInTitle")}</h2>
            </motion.div>
            <motion.div className="header-actions" variants={staggerChild}>
              <FontPicker language={language} />
              <ThemeToggle language={language} />
              <LanguageDialog language={language} onLanguageChange={setLanguage} />
              <Button asChild variant="ghost" size="icon" className="rounded-full">
                <Link href="/settings" title={t(language, "settings")} aria-label={t(language, "settings")}>
                  <SettingsIcon aria-hidden="true" size={18} />
                </Link>
              </Button>
            </motion.div>
          </motion.div>

          <motion.section
            className="auth-login-section"
            aria-label={t(language, "authLoginSectionTitle")}
            variants={staggerChild}
          >
            <div className="auth-login-heading">
              <ShieldCheck aria-hidden="true" size={18} />
              <strong>{t(language, "authLoginSectionTitle")}</strong>
            </div>

            <div className="social-login-grid">
              <Button
                className="social-login-button google-login"
                type="button"
                variant="outline"
                size="lg"
                onClick={handleGoogleLogin}
                disabled={providerLoginState !== "idle" || loginState === "loading"}
              >
                {providerLoginState === "google" ? (
                  <Loader2 className="spin" aria-hidden="true" size={20} />
                ) : (
                  <Image alt="" height={20} src="/google_logo.png" width={20} />
                )}
                <span>{t(language, "loginWithGoogle")}</span>
              </Button>
              {isLocalTestHost ? (
                <Button
                  className="social-login-button local-google-test-login"
                  type="button"
                  variant="outline"
                  size="lg"
                  onClick={handleLocalGoogleTestLogin}
                  disabled={providerLoginState !== "idle" || loginState === "loading"}
                  title={t(language, "localGoogleTestLogin")}
                >
                  {providerLoginState === "local-google" ? (
                    <Loader2 className="spin" aria-hidden="true" size={20} />
                  ) : (
                    <Image alt="" height={20} src="/google_logo.png" width={20} />
                  )}
                  <span>
                    <span className="dev-test-badge" aria-hidden="true">DEV</span>
                    {language === "th" ? "ทดสอบเข้าระบบ Google" : "Test Google login"}
                  </span>
                </Button>
              ) : null}
            </div>
          </motion.section>

          <motion.div
            className="login-divider"
            role="separator"
            aria-label={t(language, "socialLoginSeparator")}
            variants={staggerChild}
          >
            <span>{t(language, "socialLoginSeparator")}</span>
          </motion.div>

          <motion.section
            className="password-login-section"
            aria-label={t(language, "passwordLoginSectionTitle")}
            variants={staggerChild}
          >
            <div className="password-login-heading">
              <LockKeyhole aria-hidden="true" size={18} />
              <div>
                <span>{t(language, "passwordLoginSectionTitle")}</span>
                <small>{t(language, "passwordLoginSectionDescription")}</small>
              </div>
            </div>

            <div className="field-group holding-code-field">
              <span id="holding-code-label">{t(language, "holdingCode")}</span>
              <div className="input-with-icon">
                <Building2 aria-hidden="true" size={18} className="input-leading-icon" />
                <Input
                  aria-labelledby="holding-code-label"
                  aria-describedby="holding-code-help"
                  autoComplete="organization"
                  value={holdingCode}
                  onChange={(event) => setHoldingCode(normalizeHoldingCode(event.target.value))}
                  placeholder="bcdemo01"
                  className="!pl-10 h-11"
                />
              </div>
              <small id="holding-code-help" className="field-help">{t(language, "holdingCodeHint")}</small>
            </div>

            <label className="field-group" htmlFor="login-username">
              <span>{t(language, "username")}</span>
              <div className="input-with-icon">
                <UserRound aria-hidden="true" size={18} className="input-leading-icon" />
                <Input
                  id="login-username"
                  autoComplete="username"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  placeholder={t(language, "usernamePlaceholder")}
                  className="!pl-10 h-11"
                />
              </div>
            </label>

            <label className="field-group" htmlFor="login-password">
              <span>{t(language, "password")}</span>
              <div className="input-with-icon">
                <LockKeyhole aria-hidden="true" size={18} className="input-leading-icon" />
                <Input
                  id="login-password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={t(language, "password")}
                  type={showPassword ? "text" : "password"}
                  className="!pl-10 !pr-10 h-11"
                />
                <button
                  className="input-trailing-icon"
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

            <AnimatePresence mode="wait">
              {message ? (
                <motion.div
                  key={message + loginState}
                  className={`message ${loginState === "success" || connectionState === "success" ? "success" : "error"}`}
                  initial={{ opacity: 0, y: -6, height: 0 }}
                  animate={{ opacity: 1, y: 0, height: "auto" }}
                  exit={{ opacity: 0, y: -6, height: 0 }}
                  transition={{ duration: 0.24, ease: [0.22, 1, 0.36, 1] }}
                >
                  {loginState === "success" || connectionState === "success" ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
                  <span>{message}</span>
                </motion.div>
              ) : null}
            </AnimatePresence>

            <Button
              className="primary-button w-full h-12 text-base font-semibold"
              type="submit"
              size="lg"
              disabled={!mounted || !canSubmit}
            >
              {loginState === "loading" ? <Loader2 className="spin" size={18} /> : <LogIn size={18} />}
              <span>{loginState === "loading" ? t(language, "loggingIn") : t(language, "login")}</span>
            </Button>

            <p className="sign-up-hint">
              {t(language, "noAccountPrompt")}{" "}
              <button
                type="button"
                className="sign-up-link"
                onClick={handleGoogleLogin}
                disabled={providerLoginState !== "idle" || loginState === "loading"}
              >
                {t(language, "loginWithGoogle")} →
              </button>
            </p>
          </motion.section>

          <motion.div
            className="first-use-note"
            aria-label={t(language, "firstUseTitle")}
            variants={staggerChild}
          >
            <Building2 aria-hidden="true" size={20} />
            <div>
              <strong>{t(language, "firstUseTitle")}</strong>
              <span>{t(language, "firstUseDescription")}</span>
              <ol className="first-use-steps">
                <li>{t(language, "firstUseStepHolding")}</li>
                <li>{t(language, "firstUseStepGoogle")}</li>
                <li>{t(language, "firstUseStepWorkspace")}</li>
              </ol>
            </div>
          </motion.div>

        </form>
      </motion.section>
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

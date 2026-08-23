"use client";

import {
  CheckCircle2,
  Copy,
  ExternalLink,
  Loader2,
  MessageCircle,
  QrCode,
  RefreshCcw,
  ShieldCheck,
  UserRound,
} from "lucide-react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import QRCode from "qrcode";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { LogoAvatar, useProfileAvatar } from "@/components/logo-avatar";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { formatDefaultDateTime, resolveWorkspaceDateTimeDisplayOptions } from "@/lib/date-time";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { pushNotice } from "@/lib/toast";
import { authFetch, getAuthSession } from "@/lib/client-auth-session";
import {
  branchDisplayName,
  shopDisplayName,
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { AppHeaderControls } from "../app-header-controls";

type LineOaLinkScreenProps = {
  embedded?: boolean;
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage?: LanguageCode;
  language?: LanguageCode;
};

type LineOaTextKey = keyof typeof lineOaTextEn;
type LinkState = "idle" | "loading" | "ready" | "success" | "error";

type LineOaUserProfile = {
  linked: boolean;
  lineUserId: string;
  displayName: string;
  pictureUrl: string;
  linkedAt: string;
};

type LineOaApiResponse = {
  success?: boolean;
  status?: string;
  message?: string;
  link?: string;
  linked?: boolean;
  data?: unknown;
};

const lineOaTextEn = {
  title: "Connect LINE OA",
  subtitle: "Link the selected company user with LINE OA for alerts and approvals.",
  currentUser: "Current user",
  selectedCompany: "Selected company",
  selectedBranch: "Selected branch",
  status: "Status",
  linked: "Linked",
  notLinked: "Not linked",
  lineUserId: "LINE User ID",
  lineDisplayName: "LINE display name",
  linkedAt: "Linked at",
  noLineProfile: "No LINE OA profile found for this user.",
  createLink: "Create link",
  creatingLink: "Creating link",
  refreshStatus: "Refresh status",
  qrTitle: "LINE OA QR",
  scanQr: "Scan this QR with LINE or open the link on mobile.",
  openLine: "Open LINE",
  copyLink: "Copy link",
  copied: "Copied",
  linkReady: "Link is ready. Complete the connection in LINE.",
  linkedSuccess: "LINE OA connected.",
  loadFailed: "Could not load LINE OA profile.",
  requestFailed: "Request failed.",
  loginRequired: "Please login and select a company before opening this screen.",
  setupRequired: "LINE OA or LIFF ID is not configured for this company.",
  secureNote: "User-level link for the active company only.",
  stepCompany: "Company",
  stepGenerate: "Link",
  stepConfirm: "Confirm",
  none: "None",
} as const;

const lineOaBackendKeys: Record<LineOaTextKey, string> = {
  title: "lineoa_user_link_title",
  subtitle: "lineoa_user_link_subtitle",
  currentUser: "current_user",
  selectedCompany: "selected_company",
  selectedBranch: "selected_branch",
  status: "status",
  linked: "linked",
  notLinked: "not_linked",
  lineUserId: "lineuserid",
  lineDisplayName: "linedisplayname",
  linkedAt: "linked_at",
  noLineProfile: "lineoa_no_user_profile",
  createLink: "lineoa_create_link",
  creatingLink: "lineoa_creating_link",
  refreshStatus: "refresh_status",
  qrTitle: "line_oa_qrcode",
  scanQr: "lineoa_scan_qr",
  openLine: "open_line",
  copyLink: "copy_link",
  copied: "copied",
  linkReady: "lineoa_link_ready",
  linkedSuccess: "lineoa_linked_success",
  loadFailed: "lineoa_profile_load_failed",
  requestFailed: "request_failed",
  loginRequired: "lineoa_login_required",
  setupRequired: "lineoa_setup_required",
  secureNote: "lineoa_secure_note",
  stepCompany: "step_company",
  stepGenerate: "lineoa_step_generate",
  stepConfirm: "lineoa_step_confirm",
  none: "none",
};

const lineOaText: Partial<Record<LanguageCode, Partial<Record<LineOaTextKey, string>>>> = {
  th: {
    title: "เชื่อมต่อ LINE OA",
    subtitle: "ผูกผู้ใช้ของบริษัทที่เลือกกับ LINE OA สำหรับแจ้งเตือนและอนุมัติเอกสาร",
    currentUser: "ผู้ใช้ปัจจุบัน",
    selectedCompany: "บริษัทที่เลือก",
    selectedBranch: "สาขาที่เลือก",
    status: "สถานะ",
    linked: "เชื่อมแล้ว",
    notLinked: "ยังไม่เชื่อม",
    lineUserId: "LINE User ID",
    lineDisplayName: "ชื่อ LINE",
    linkedAt: "เชื่อมเมื่อ",
    noLineProfile: "ยังไม่พบโปรไฟล์ LINE OA ของผู้ใช้นี้",
    createLink: "สร้างลิงก์เชื่อมต่อ",
    creatingLink: "กำลังสร้างลิงก์",
    refreshStatus: "โหลดสถานะใหม่",
    qrTitle: "QR สำหรับ LINE OA",
    scanQr: "สแกน QR ด้วย LINE หรือเปิดลิงก์บนมือถือ",
    openLine: "เปิด LINE",
    copyLink: "คัดลอกลิงก์",
    copied: "คัดลอกแล้ว",
    linkReady: "สร้างลิงก์แล้ว ให้ยืนยันต่อใน LINE",
    linkedSuccess: "เชื่อมต่อ LINE OA สำเร็จ",
    loadFailed: "โหลดโปรไฟล์ LINE OA ไม่สำเร็จ",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    loginRequired: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้",
    setupRequired: "บริษัทนี้ยังไม่ได้ตั้งค่า LINE OA หรือ LIFF ID",
    secureNote: "ผูกระดับ user เฉพาะบริษัทที่ใช้งานอยู่",
    stepCompany: "บริษัท",
    stepGenerate: "ลิงก์",
    stepConfirm: "ยืนยัน",
    none: "ไม่มี",
  },
  en: lineOaTextEn,
};

const emptyProfile: LineOaUserProfile = {
  linked: false,
  lineUserId: "",
  displayName: "",
  pictureUrl: "",
  linkedAt: "",
};

export function LineOaLinkScreen({
  embedded = false,
  initialBackendLanguage,
  initialBackendUrl,
  initialLanguage = "th",
  language: externalLanguage,
}: LineOaLinkScreenProps) {
  const router = useRouter();
  const [language, setLanguage] = useState<LanguageCode>(externalLanguage ?? initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const profileAvatar = useProfileAvatar(auth);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [profile, setProfile] = useState<LineOaUserProfile>(emptyProfile);
  const setNotice = pushNotice;
  const [loadingProfile, setLoadingProfile] = useState(false);
  const [linkState, setLinkState] = useState<LinkState>("idle");
  const [linkUrl, setLinkUrl] = useState("");
  const [qrDataUrl, setQrDataUrl] = useState("");
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);
  const dateTimeDisplayOptions = useMemo(
    () => resolveWorkspaceDateTimeDisplayOptions(workspace, language),
    [language, workspace],
  );

  const text = useCallback(
    (key: LineOaTextKey) => {
      const fallback = lineOaText[language]?.[key] ?? lineOaTextEn[key] ?? key;
      return backendText(backendLanguage, lineOaBackendKeys[key], fallback);
    },
    [backendLanguage, language],
  );

  const loadProfile = useCallback(async (currentAuth: AuthSession | null, currentWorkspace: WorkspaceSession | null, silent = false) => {
    if (!currentAuth || !currentWorkspace) return;
    if (!silent) {
      setLoadingProfile(true);
      setNotice(null);
    }

    try {
      const payload = await callLineOaUserApi(currentAuth, currentWorkspace, "profile");
      const nextProfile = normalizeLineOaProfile(payload);
      setProfile(nextProfile);
      if (nextProfile.linked) {
        setLinkState("success");
        setNotice({ type: "success", text: text("linkedSuccess") });
      } else if (!silent) {
        setNotice({ type: "info", text: text("noLineProfile") });
      }
    } catch (error) {
      if (!silent) {
        setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("loadFailed") });
      }
    } finally {
      if (!silent) setLoadingProfile(false);
    }
  }, [text]);

  useEffect(() => {
    if (externalLanguage) setLanguage(externalLanguage);
  }, [externalLanguage]);

  useEffect(() => {
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? externalLanguage ?? initialLanguage);
    if (!externalLanguage) setLanguage(savedLanguage);

    const nextAuth = readAuth();
    const nextWorkspace = readWorkspace();
    if (!nextAuth || !nextWorkspace) {
      setNotice({ type: "error", text: getLocalText(savedLanguage, "loginRequired") });
      if (!embedded) router.replace("/");
      return;
    }

    setAuth(nextAuth);
    setWorkspace(nextWorkspace);
    void loadProfile(nextAuth, nextWorkspace);
  }, [embedded, externalLanguage, initialLanguage, loadProfile, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    if (!externalLanguage) localStorage.setItem("user_language", language);
  }, [externalLanguage, language]);

  useEffect(() => {
    if (!linkUrl || profile.linked || !auth || !workspace) return;
    const interval = window.setInterval(() => {
      void loadProfile(auth, workspace, true);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [auth, linkUrl, loadProfile, profile.linked, workspace]);

  const companyName = workspace ? shopDisplayName(workspace.shop) : "-";
  const branchName = workspace?.branch ? branchDisplayName(workspace.branch) : text("none");
  const canCreateLink = Boolean(auth && workspace && !loadingProfile && linkState !== "loading");
  const stepItems = useMemo(() => [
    { label: text("stepCompany"), done: Boolean(workspace) },
    { label: text("stepGenerate"), done: Boolean(linkUrl) || profile.linked },
    { label: text("stepConfirm"), done: profile.linked },
  ], [linkUrl, profile.linked, text, workspace]);

  async function createLink() {
    if (!auth || !workspace) return;
    setLinkState("loading");
    setNotice(null);

    try {
      const payload = await callLineOaUserApi(auth, workspace, "link");
      if (payload.status !== "success" || !payload.link) {
        throw new Error(payload.message || text("setupRequired"));
      }

      const qr = await QRCode.toDataURL(payload.link, {
        margin: 1,
        width: 240,
        color: { dark: "#111827", light: "#ffffff" },
      });
      setLinkUrl(payload.link);
      setQrDataUrl(qr);
      setLinkState("ready");
      setNotice({ type: "info", text: text("linkReady") });
    } catch (error) {
      setLinkState("error");
      setNotice({ type: "error", text: error instanceof Error && error.message ? error.message : text("requestFailed") });
    }
  }

  async function copyLink() {
    if (!linkUrl) return;
    try {
      await navigator.clipboard.writeText(linkUrl);
      setNotice({ type: "success", text: text("copied") });
    } catch {
      setNotice({ type: "error", text: text("requestFailed") });
    }
  }

  const content = (
    <div className="grid w-full min-w-0 gap-3">
      <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
        <div className="flex w-full flex-wrap items-start justify-between gap-2">
          <div className="flex min-w-0 items-start gap-2">
            <span className="grid size-10 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              <MessageCircle size={22} />
            </span>
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase text-muted-foreground">BC Ai Account</p>
              <h1 className="truncate text-xl font-semibold sm:text-2xl">{text("title")}</h1>
              <p className="max-w-[92ch] text-sm leading-6 text-muted-foreground">{text("subtitle")}</p>
            </div>
          </div>
          {embedded ? null : (
            <div className="flex min-w-0 flex-wrap justify-end gap-2">
              <AppHeaderControls language={language} onLanguageChange={setLanguage} showSettings={false} />
            </div>
          )}
        </div>
        <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
          <Badge variant="outline">{text("selectedCompany")}: {companyName}</Badge>
          <Badge variant="outline">{text("selectedBranch")}: {branchName}</Badge>
          <Badge variant={profile.linked ? "success" : "secondary"}>{text("status")}: {profile.linked ? text("linked") : text("notLinked")}</Badge>
        </div>
      </header>

      <section className="grid gap-2 md:grid-cols-3">
        {stepItems.map((item) => (
          <Card key={item.label}>
            <CardContent className="flex items-center gap-2 p-3">
              <span className={item.done ? "text-emerald-600" : "text-muted-foreground"}>
                {item.done ? <CheckCircle2 size={18} /> : <ShieldCheck size={18} />}
              </span>
              <span className="min-w-0 truncate text-sm font-semibold">{item.label}</span>
            </CardContent>
          </Card>
        ))}
      </section>

      <section className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_360px]">
        <Card>
          <CardHeader className="p-3 pb-1">
            <CardTitle className="flex items-center gap-2 text-base">
              <LogoAvatar
                uri={profileAvatar}
                auth={auth}
                alt={auth?.username ?? text("currentUser")}
                sizeClass="size-6 rounded-full shrink-0"
                iconSize={14}
                width={48}
                fallbackIcon={UserRound}
              />
              {text("currentUser")}
            </CardTitle>
            <CardDescription>{text("secureNote")}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 p-3">
            <div className="grid gap-2 sm:grid-cols-2">
              <InfoTile label={text("currentUser")} value={auth?.username ?? "-"} />
              <InfoTile label={text("status")} value={profile.linked ? text("linked") : text("notLinked")} />
              <InfoTile label={text("lineDisplayName")} value={profile.displayName || "-"} />
              <InfoTile label={text("lineUserId")} value={profile.lineUserId || "-"} />
              <InfoTile label={text("linkedAt")} value={profile.linkedAt ? formatDefaultDateTime(profile.linkedAt, dateTimeDisplayOptions) : "-"} />
              <InfoTile label={text("selectedCompany")} value={companyName} />
            </div>

            <div className="flex flex-wrap gap-2">
              <Button type="button" onClick={() => void createLink()} disabled={!canCreateLink}>
                {linkState === "loading" ? <Loader2 className="animate-spin" /> : <QrCode />}
                {linkState === "loading" ? text("creatingLink") : profile.linked ? backendText(backendLanguage, "reconnect_line", text("createLink")) : text("createLink")}
              </Button>
              <Button type="button" variant="outline" onClick={() => void loadProfile(auth, workspace)} disabled={!auth || !workspace || loadingProfile}>
                {loadingProfile ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
                {text("refreshStatus")}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="p-3 pb-1">
            <CardTitle className="flex items-center gap-2 text-base">
              <QrCode className="size-4" />
              {text("qrTitle")}
            </CardTitle>
            <CardDescription>{text("scanQr")}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 p-3">
            {qrDataUrl ? (
              <div className="grid place-items-center gap-3">
                <div className="rounded-2xl border border-border bg-white p-2 shadow-sm">
                  <Image alt={text("qrTitle")} height={240} src={qrDataUrl} unoptimized width={240} />
                </div>
                <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-1">
                  <Button asChild>
                    <a href={linkUrl} rel="noreferrer" target="_blank">
                      <ExternalLink />
                      {text("openLine")}
                    </a>
                  </Button>
                  <Button type="button" variant="outline" onClick={() => void copyLink()}>
                    <Copy />
                    {text("copyLink")}
                  </Button>
                </div>
              </div>
            ) : (
              <div className="grid min-h-60 place-items-center rounded-2xl border border-dashed border-border bg-muted/30 p-4 text-center text-sm text-muted-foreground">
                <div className="grid gap-2">
                  <QrCode className="mx-auto size-10" />
                  <span>{text("noLineProfile")}</span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </section>
    </div>
  );

  if (embedded) return <section className="grid w-full min-w-0 gap-3">{content}</section>;
  return <main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">{content}</main>;
}

function InfoTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0 rounded-xl border border-border bg-muted/30 p-3">
      <p className="text-xs font-semibold text-muted-foreground">{label}</p>
      <p className="mt-1 truncate text-sm font-semibold">{value}</p>
    </div>
  );
}

async function callLineOaUserApi(
  auth: AuthSession,
  workspace: WorkspaceSession,
  action: "link" | "profile",
): Promise<LineOaApiResponse> {
  const response = await authFetch("/api/line-oa/user", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${auth.token}`,
    },
    body: JSON.stringify({
      action,
      backendUrl: auth.backendUrl,
      holdingCode: workspace.shop.holdingcode,
      username: auth.username,
    }),
  });
  const payload = (await response.json().catch(() => ({}))) as LineOaApiResponse;
  if (!response.ok || payload.success === false) {
    throw new Error(payload.message || "LINE OA request failed");
  }
  return payload;
}

function normalizeLineOaProfile(payload: LineOaApiResponse): LineOaUserProfile {
  const data = isRecord(payload.data) ? payload.data : {};
  const lineUserId = getString(data, "lineuserid");
  return {
    linked: payload.linked === true || Boolean(lineUserId),
    lineUserId,
    displayName: getString(data, "linedisplayname"),
    pictureUrl: getString(data, "linepictureurl"),
    linkedAt: getString(data, "linelinkedat"),
  };
}

function readAuth(): AuthSession | null {
  return getAuthSession();
}

function readWorkspace(): WorkspaceSession | null {
  return readStorage<WorkspaceSession>(workspaceStorageKeys.workspace);
}

function readStorage<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return null;
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

function getLocalText(language: LanguageCode, key: LineOaTextKey): string {
  return lineOaText[language]?.[key] ?? lineOaTextEn[key] ?? key;
}

function getString(payload: Record<string, unknown>, key: string): string {
  const value = payload[key];
  return typeof value === "string" ? value : "";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

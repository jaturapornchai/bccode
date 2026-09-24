"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { ArrowLeft, Copy, KeyRound, Plus, RefreshCw, ShieldCheck, Trash2, Building2, Bot, Code2, ArrowUpRight, LockKeyhole } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { backendText, useBackendLanguage } from "@/lib/backend-language";
import { getAuthSession } from "@/lib/client-auth-session";
import { normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { WORKSPACE_CHANGED_EVENT, workspaceStorageKeys, type WorkspaceSession } from "@/lib/workspace-models";
import { AppHeaderControls } from "../app-header-controls";
import { Check, control } from "../gl/gl-common";
import { tokenRequest, tokenStatus, TokenRequestError, type MCPToken, type TokenCompany } from "./token-client";

import styles from "./token-screen.module.css";

export function MCPTokenScreen() {
  const [language, setLanguage] = useState<LanguageCode>("th");
  const [context, setContext] = useState<string | null>(null);
  useEffect(() => {
    setLanguage(normalizeLanguage(localStorage.getItem("user_language") ?? "th"));
    const sync = () => setContext(`${localStorage.getItem(workspaceStorageKeys.workspace) ?? ""}|${localStorage.getItem(workspaceStorageKeys.auth) ?? ""}`);
    sync();
    window.addEventListener("storage", sync); window.addEventListener(WORKSPACE_CHANGED_EVENT, sync); window.addEventListener("focus", sync);
    return () => { window.removeEventListener("storage", sync); window.removeEventListener(WORKSPACE_CHANGED_EVENT, sync); window.removeEventListener("focus", sync); };
  }, []);
  return context === null ? null : <TokenWorkbench key={context} language={language} onLanguageChange={setLanguage} />;
}

function TokenWorkbench({ language, onLanguageChange }: { language: LanguageCode; onLanguageChange: (language: LanguageCode) => void }) {
  const router = useRouter();
  const dictionary = useBackendLanguage(language, getAuthSession()?.backendUrl);
  const tr = useCallback((key: string, fallback: string) => backendText(dictionary, key, fallback), [dictionary]);
  const [tokens, setTokens] = useState<MCPToken[]>([]);
  const [companies, setCompanies] = useState<TokenCompany[]>([]);
  const [companyCodes, setCompanyCodes] = useState<string[]>([]);
  const [selected, setSelected] = useState<MCPToken | null>(null);
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState("");
  const [kind, setKind] = useState<"api" | "mcp">("mcp");
  const [mode, setMode] = useState<"readonly" | "readwrite">("readonly");
  const [days, setDays] = useState("90");
  const [secret, setSecret] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [denied, setDenied] = useState(false);
  const [authorized, setAuthorized] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const alive = useRef(true);
  const containerRef = useRef<HTMLDivElement>(null);
  const split = useSplitPercent({ storageKey: "bc_mcp_tokens_split", defaultLeft: 38, min: 25, max: 65, containerRef });
  const { confirm, confirmationDialog } = useConfirmDialog();
  const dirty = creating && (name !== "" || companyCodes.length > 0 || kind !== "mcp" || mode !== "readonly" || days !== "90");
  const connectionKind = selected?.kind ?? kind;
  const allowedCompanies = (token: MCPToken) => token.companyCodes?.length ? token.companyCodes : token.companyCode ? [token.companyCode] : [];
  const companyLabel = (code: string) => { const company = companies.find((item) => item.code === code); return company?.name ? `${code} — ${company.name}` : code; };
  const statusText = (token: MCPToken) => ({ active: tr("mcp_active", "ใช้งานได้"), expired: tr("mcp_expired", "หมดอายุ"), revoked: tr("mcp_revoked", "เพิกถอนแล้ว") })[tokenStatus(token)];
  const dateText = (value?: string | null) => value ? new Date(value).toLocaleString(language, { timeZone: "Asia/Bangkok" }) : "—";
  const [workspace] = useState<WorkspaceSession | null>(() => {
    try { return JSON.parse(localStorage.getItem(workspaceStorageKeys.workspace) ?? "null"); } catch { return null; }
  });

  const fail = useCallback((cause: unknown) => {
    if (!alive.current) return;
    if (cause instanceof TokenRequestError && (cause.status === 401 || cause.status === 403)) {
      setDenied(true); setSecret(""); setTokens([]); setSelected(null); setCreating(false);
    }
    setError(cause instanceof Error ? cause.message : tr("mcp_request_failed", "ไม่สามารถจัดการ token ได้ กรุณาลองใหม่"));
  }, [tr]);
  const load = useCallback(async () => {
    setLoading(true); setError(""); setAuthorized(false);
    try { const [data, choices] = await Promise.all([tokenRequest<MCPToken[]>(), tokenRequest<TokenCompany[]>("/companies")]); if (alive.current) { setTokens((data ?? []).filter((token) => !token.revokedAt)); setCompanies(choices ?? []); setSelected((current) => current ? data.find((token) => token.id === current.id && !token.revokedAt) ?? null : null); setDenied(false); setAuthorized(true); } }
    catch (cause) { fail(cause); }
    finally { if (alive.current) setLoading(false); }
  }, [fail]);
  useEffect(() => {
    alive.current = true; setEndpoint(window.location.origin); void load();
    return () => { alive.current = false; };
  }, [load]);
  useEffect(() => {
    if (!dirty && !secret && !busy) return;
    const prevent = (event: BeforeUnloadEvent) => { event.preventDefault(); };
    window.addEventListener("beforeunload", prevent);
    return () => window.removeEventListener("beforeunload", prevent);
  }, [dirty, secret, busy]);
  async function mayLeave() {
    if (busy) return false;
    if (!dirty && !secret) return true;
    return confirm({ title: tr("mcp_leave_title", "ออกจากข้อมูลนี้หรือไม่"), description: secret ? tr("mcp_secret_leave", "Token นี้แสดงครั้งเดียว กรุณาคัดลอกก่อนออกจากหน้านี้") : tr("mcp_unsaved", "ข้อมูลที่ยังไม่ได้สร้างจะถูกยกเลิก"), confirmLabel: tr("mcp_continue", "ดำเนินการต่อ"), cancelLabel: tr("common_cancel", "ยกเลิก") });
  }
  async function choose(token: MCPToken | null, newKind: "api" | "mcp" = "mcp") {
    if (!await mayLeave()) return;
    setSecret(""); setMessage(""); setError(""); setSelected(token); setCreating(!token); setName(""); setKind(newKind); setCompanyCodes([]); setMode("readonly"); setDays("90");
  }
  async function create(event: FormEvent) {
    event.preventDefault(); if (busy) return;
    const duration = Number(days);
    if (!companyCodes.length) { setError(tr("mcp_companies_required", "กรุณาเลือกบริษัทที่อนุญาตอย่างน้อยหนึ่งบริษัท")); return; }
    if (!name.trim() || !Number.isInteger(duration) || duration < 1 || duration > 365) { setError(tr("mcp_validation", "กรุณาระบุชื่อและอายุ 1–365 วัน")); return; }
    setBusy(true); setError(""); setMessage("");
    try {
      const created = await tokenRequest<MCPToken & { token: string }>("", { method: "POST", body: JSON.stringify({ name: name.trim(), kind, mode, companyCodes, expiresAt: new Date(Date.now() + duration * 86400000).toISOString() }) });
      if (!alive.current) return;
      const { token, ...metadata } = created;
      setSecret(token); setSelected(metadata); setTokens((old) => [metadata, ...old]); setCreating(false); setName("");
    } catch (cause) { fail(cause); } finally { if (alive.current) setBusy(false); }
  }
  async function remove(token: MCPToken) {
    if (!await confirm({ title: tr("mcp_delete", "ลบ token"), description: tr("mcp_delete_hint", "ลบ token นี้ออกจากรายการและยกเลิกการเชื่อมต่อทันที โดยเก็บประวัติไว้ตรวจสอบ ไม่สามารถนำ token เดิมกลับมาใช้ได้"), details: token.name, confirmLabel: tr("mcp_delete", "ลบ token"), cancelLabel: tr("common_cancel", "ยกเลิก"), tone: "danger" })) return;
    setBusy(true); setError("");
    try {
      await tokenRequest(`/${token.id}/revoke`, { method: "POST" });
      if (!alive.current) return;
      setSecret(""); setSelected(null); setMessage(tr("mcp_delete_success", "ลบ token แล้ว")); await load();
    } catch (cause) { fail(cause); } finally { if (alive.current) setBusy(false); }
  }
  async function copySecret() {
    try { await navigator.clipboard.writeText(secret); if (alive.current) setMessage(tr("mcp_copied", "คัดลอก token แล้ว")); }
    catch { setError(tr("mcp_copy_failed", "คัดลอกไม่สำเร็จ กรุณาเลือกข้อความและคัดลอกด้วยตนเอง")); }
  }
  const holding = workspace?.shop?.holdingcode || getAuthSession()?.holdingcode || "—";
  const connectionGuide = <div className={styles.guide}>
    <p className="break-all text-muted-foreground"><span className="font-semibold text-foreground">{connectionKind.toUpperCase()} URL: </span>{endpoint}{connectionKind === "api" ? "/api/integration/gl" : "/mcp/gl"}</p>
    <details className="mt-2"><summary>{tr("mcp_connection_guide", "วิธีเชื่อมต่อและขอบเขตสิทธิ์")}</summary><div className="grid gap-2 pt-2 text-muted-foreground">
      <p>{connectionKind === "api" ? tr("mcp_api_hint", "REST API: ส่ง API token แบบ Bearer; เลือกบริษัทด้วย X-BC-Company-Code หรือ readonly หลายบริษัทด้วย X-BC-Company-Codes คั่นด้วยจุลภาค") : tr("mcp_bearer_hint", "MCP Streamable HTTP: ส่ง MCP token แบบ Bearer; gl_* tool ระบุ companyCode หรือ readonly หลายบริษัทระบุ companyCodes")}</p>
      <p>{tr("mcp_batch_scope_hint", "readonly อ่านหลายบริษัทในคำขอเดียวได้; readwrite ทุกคำขอระบุได้บริษัทเดียว รวมคำขออ่าน ผลอ่านหลายบริษัทแยกตามบริษัท ไม่รวมยอดอัตโนมัติ")}</p>
    </div></details>
  </div>;
  return <main className={`${styles.screen} flex min-h-dvh flex-col text-base leading-relaxed text-foreground lg:h-dvh lg:overflow-hidden`}>
    <header className={`${styles.surface} ${styles.header} flex shrink-0 flex-wrap items-center justify-between gap-4`}>
      <div className="flex items-center gap-4"><span className={styles.icon}><KeyRound className="size-6" /></span><div><h1 className="text-2xl font-bold tracking-tight">{tr("mcp_title", "จัดการ API / MCP token")}</h1><p className="mt-1 text-muted-foreground">{tr("mcp_subtitle", "เชื่อมต่อโปรแกรมและผู้ช่วย AI กับห้องบัญชีของคุณ")}</p></div></div>
      <div className="flex flex-wrap items-center gap-3"><Button asChild variant="outline" size="lg"><Link href="/workspace" onClick={async (event) => { event.preventDefault(); if (await mayLeave()) router.push("/workspace"); }}><ArrowLeft />{tr("mcp_back", "กลับหน้าเลือกบริษัท")}</Link></Button><AppHeaderControls language={language} onLanguageChange={onLanguageChange} /></div>
    </header>
    <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 px-1"><span className={styles.badge}><Building2 className="size-4 text-primary" />{tr("mcp_holding", "Holding")}: {holding}</span><span className="flex items-center gap-2 text-sm text-muted-foreground"><ShieldCheck className="size-4 text-primary" />{tr("mcp_admin_only", "เฉพาะผู้ดูแลระบบ (OWNER / ADMIN)")}</span></div>
    {error && <div role="alert" className="rounded-xl border border-destructive/40 bg-destructive/5 p-3">{error}</div>}
    {message && <div role="status" className="rounded-xl border border-primary/30 bg-primary/5 p-3">{message}</div>}
    {denied ? <section className={`${styles.surface} ${styles.detail} grid justify-items-start gap-4`}><ShieldCheck className="size-8 text-primary" /><h2 className="font-semibold">{tr("mcp_denied", "ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")}</h2><Button size="lg" variant="outline" disabled={loading || busy} onClick={() => void load()}><RefreshCw className={loading ? "animate-spin" : ""} />{tr("mcp_reload", "โหลดใหม่")}</Button></section> : !authorized ? <section className={`${styles.surface} ${styles.detail}`}>{loading ? <p role="status">{tr("mcp_loading", "กำลังตรวจสอบสิทธิ์และโหลดข้อมูล…")}</p> : <Button size="lg" variant="outline" onClick={() => void load()}>{tr("mcp_reload", "โหลดใหม่")}</Button>}</section> : <div ref={containerRef} className="flex min-h-0 flex-1 flex-col gap-3 lg:flex-row lg:gap-0">
      <section style={{ "--token-list-width": `${split.splitPercent}%` } as React.CSSProperties} className={`${styles.surface} flex min-h-0 flex-col lg:!w-[var(--token-list-width)] lg:shrink-0`}>
        <div className={styles.listHeader}>
          <div className="mb-4 flex flex-wrap items-center justify-between gap-2"><h2 className="flex items-center gap-2 font-semibold">{tr("mcp_list", "รายการ token")}<span className={styles.badge}>{tokens.length}</span></h2><Button size="lg" variant="ghost" disabled={loading || busy} onClick={() => void load()}><RefreshCw className={loading ? "animate-spin" : ""} />{tr("mcp_reload", "โหลดใหม่")}</Button></div>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2"><Button size="lg" disabled={loading || busy} onClick={() => void choose(null, "mcp")}><Plus />{tr("mcp_add_mcp", "เพิ่ม MCP token")}</Button><Button size="lg" variant="outline" disabled={loading || busy} onClick={() => void choose(null, "api")}><Plus />{tr("mcp_add_api", "เพิ่ม API token")}</Button></div>
        </div>
        <div className={`${styles.listBody} ${creating && tokens.length === 0 ? styles.compactEmpty : ""} min-h-0 flex-1 overflow-y-auto`}>
          {loading ? <p role="status">{tr("mcp_loading", "กำลังตรวจสอบสิทธิ์และโหลดข้อมูล…")}</p> : tokens.length === 0 ? <div className="grid justify-items-center gap-3 px-4 py-12 text-center"><span className={styles.icon}><KeyRound className="size-6" /></span><p className="font-semibold">{tr("mcp_empty", "ยังไม่มี token สำหรับ Holding นี้")}</p><p className="max-w-xs text-muted-foreground">{tr("mcp_empty_hint", "เริ่มเพิ่ม token ตามประเภทการเชื่อมต่อที่ต้องการ")}</p></div> : tokens.map((token) => <button key={token.id} disabled={busy} onClick={() => void choose(token)} aria-pressed={selected?.id === token.id} className={styles.row}>
            <span className="mb-3 flex items-start justify-between gap-2"><span className="break-words font-semibold">{token.name}</span><span className={styles.badge}>{token.kind.toUpperCase()}</span></span>
            <span className="flex flex-wrap items-center gap-2 text-sm"><span className="font-medium text-primary">{statusText(token)}</span><span className="text-muted-foreground">· {token.mode}</span></span>
            <span className="mt-2 block break-words text-sm text-muted-foreground">{allowedCompanies(token).join(", ")}</span><span className="mt-1 block text-sm text-muted-foreground">{tr("mcp_expires", "หมดอายุ")}: {dateText(token.expiresAt)}</span>
          </button>)}
        </div>
        <p className="flex items-center gap-2 border-t border-border px-4 py-3 text-sm text-muted-foreground"><LockKeyhole className="size-4 shrink-0" />{tr("mcp_list_footnote", "ลบ token เพื่อยกเลิกการเชื่อมต่อได้ทุกเมื่อ")}</p>
      </section>
      <ResizableSplitter {...{ min: split.min, max: split.max, value: split.splitPercent, isResizing: split.isResizing, onPointerDown: split.startResize, onDoubleClick: split.resetSplit, onKeyDown: split.adjustWithKeyboard }} label={tr("mcp_resize", "ปรับความกว้างรายการ token")} />
      <section className={`${styles.surface} ${styles.detail} min-h-0 flex-1 overflow-y-auto`}>
        {creating ? <form onSubmit={(event) => void create(event)} className="mx-auto grid max-w-4xl gap-5">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-5"><div><p className="mb-1 text-sm font-semibold text-primary">{kind.toUpperCase()} TOKEN</p><h2 className="text-xl font-bold">{tr("mcp_new_connection", "เพิ่มการเชื่อมต่อใหม่")}</h2></div><Button size="lg" type="submit" disabled={busy}><Plus />{busy ? tr("mcp_saving", "กำลังสร้าง…") : tr("mcp_create", "สร้าง token")}</Button></div>
          <label className={styles.field}>{tr("mcp_name", "ชื่อการเชื่อมต่อ")}<input className={control} value={name} onChange={(event) => setName(event.target.value)} maxLength={100} required disabled={busy} autoFocus placeholder={tr("mcp_name_example", "เช่น ผู้ช่วยตรวจสอบบัญชี")}/></label>
          <div className="grid gap-4 xl:grid-cols-2"><label className={styles.field}>{tr("mcp_kind", "ประเภทการเชื่อมต่อ")}<select className={control} value={kind} onChange={(event) => setKind(event.target.value as typeof kind)} disabled={busy}><option value="mcp">MCP</option><option value="api">API</option></select></label><label className={styles.field}>{tr("mcp_days", "อายุ token (1–365 วัน)")}<input className={control} type="number" min="1" max="365" step="1" value={days} onChange={(event) => setDays(event.target.value)} required disabled={busy} /></label></div>
          <div className={styles.section}><label className={styles.field}>{tr("mcp_mode", "สิทธิ์การเข้าถึง")}<select className={control} value={mode} onChange={(event) => setMode(event.target.value as typeof mode)} disabled={busy}><option value="readonly">{tr("mcp_readonly", "readonly — อ่านและตรวจสอบเท่านั้น")}</option><option value="readwrite">{tr("mcp_readwrite", "readwrite — อ่านและบันทึกข้อมูลบัญชี")}</option></select></label><p className="mt-3 text-sm text-muted-foreground">{mode === "readonly" ? tr("mcp_read_scope", "อ่านข้อมูลได้หลายบริษัทพร้อมกัน เหมาะกับงานตรวจสอบ") : tr("mcp_write_scope", "อ่านและบันทึกข้อมูลได้ครั้งละหนึ่งบริษัท รวมถึงผ่านรายการและกลับรายการตามสิทธิ์ผู้สร้าง")}</p></div>
          <fieldset className={styles.section}><legend className="px-2 font-semibold">{tr("mcp_companies", "บริษัทที่อนุญาต")}</legend><p className="mb-3 text-sm text-muted-foreground">{tr("mcp_companies_hint", "เลือกอย่างน้อยหนึ่งบริษัท สิทธิ์ครอบคลุมทุกสาขาของบริษัทที่เลือกเท่านั้น")}</p><div className="grid max-h-56 gap-3 overflow-y-auto">{companies.map((company) => <Check className="!w-full !max-w-full [&>span:last-child]:whitespace-normal [&>span:last-child]:break-words" key={company.code} label={companyLabel(company.code)} checked={companyCodes.includes(company.code)} disabled={busy} onChange={(checked) => setCompanyCodes((codes) => checked ? [...codes, company.code] : codes.filter((code) => code !== company.code))} />)}{companies.length === 0 && <p>{tr("mcp_companies_empty", "ไม่มีบริษัทที่เปิดใช้งานใน Holding นี้ จึงยังสร้าง token ไม่ได้")}</p>}</div></fieldset>
          <p className="flex items-start gap-2 text-sm text-muted-foreground"><LockKeyhole className="mt-1 size-4 shrink-0" />{tr("mcp_secret_notice", "ระบบจะแสดง token เต็มเพียงครั้งเดียวหลังสร้าง เก็บไว้ในที่ปลอดภัย")}</p>
          {connectionGuide}
        </form> : selected ? <div className="mx-auto grid max-w-4xl gap-5">
          <div className="flex flex-wrap items-start justify-between gap-3 border-b border-border pb-5"><div><span className={styles.badge}>{selected.kind.toUpperCase()} TOKEN</span><h2 className="mt-3 break-words text-2xl font-bold">{selected.name}</h2><p className="mt-1 text-muted-foreground">{selected.mode} · {statusText(selected)}</p></div><Button className={styles.delete} size="lg" variant="destructive" disabled={busy} onClick={() => void remove(selected)}><Trash2 />{tr("mcp_delete", "ลบ token")}</Button></div>
          {secret && <div className="grid gap-3 rounded-2xl border border-primary/30 bg-primary/5 p-5"><p className="font-semibold">{tr("mcp_secret_notice", "ระบบจะแสดง token เต็มเพียงครั้งเดียวหลังสร้าง เก็บไว้ในที่ปลอดภัย")}</p><textarea className={`${control} break-all font-mono`} aria-label={tr("mcp_secret", "Token สำหรับเชื่อมต่อ")} value={secret} readOnly rows={3} spellCheck={false} autoComplete="off" /><Button className="justify-self-start" size="lg" variant="outline" onClick={() => void copySecret()}><Copy />{tr("mcp_copy", "คัดลอก token")}</Button></div>}
          <div className={styles.section}><h3 className="mb-3 flex items-center gap-2 font-semibold"><Building2 className="size-5 text-primary" />{tr("mcp_companies", "บริษัทที่อนุญาต")}</h3><ul className="grid gap-2">{allowedCompanies(selected).map((code) => <li key={code} className="break-words">{companyLabel(code)}</li>)}</ul>{selected.branchCode && <p className="mt-3">{tr("mcp_legacy_branch", "สาขาที่จำกัดไว้ใน token เดิม")}: {selected.branchCode}</p>}</div>
          <dl className="grid gap-4 xl:grid-cols-3">{[[tr("mcp_created", "สร้างเมื่อ"), selected.createdAt], [tr("mcp_expires", "หมดอายุ"), selected.expiresAt], [tr("mcp_last_used", "ใช้งานล่าสุด"), selected.lastUsedAt]].map(([label, value]) => <div key={label} className={styles.section}><dt className="text-sm text-muted-foreground">{label}</dt><dd className="mt-2 font-medium">{dateText(value)}</dd></div>)}</dl>{connectionGuide}
        </div> : <div className={styles.welcome}>
          <span className={styles.badge}><ShieldCheck className="size-4 text-primary" />{tr("mcp_controlled_access", "กำหนดสิทธิ์ได้ทุกการเชื่อมต่อ")}</span><h2 className="mt-5 text-3xl font-bold leading-snug">{tr("mcp_welcome_title", "เริ่มต้นเชื่อมต่อห้องบัญชี")}</h2><p className="mb-7 mt-3 max-w-xl text-muted-foreground">{tr("mcp_welcome_body", "เลือกประเภท token ให้ตรงกับงาน แล้วกำหนดบริษัทและสิทธิ์ที่ต้องการในขั้นตอนเดียว")}</p>
          <div className="grid gap-4 xl:grid-cols-2"><div className={styles.choice}><span className={styles.icon}><Bot className="size-6" /></span><div><h3 className="text-xl font-bold">MCP token</h3><p className="mt-2 text-muted-foreground">{tr("mcp_choice_mcp", "สำหรับผู้ช่วย AI และเครื่องมือที่รองรับ MCP เพื่ออ่าน ตรวจสอบ และทำงานกับข้อมูลบัญชี")}</p></div><Button className="mt-2 justify-self-start" size="lg" disabled={busy || loading} onClick={() => void choose(null, "mcp")}>{tr("mcp_connect_ai", "เชื่อมต่อผู้ช่วย AI")}<ArrowUpRight /></Button></div>
          <div className={styles.choice}><span className={styles.icon}><Code2 className="size-6" /></span><div><h3 className="text-xl font-bold">API token</h3><p className="mt-2 text-muted-foreground">{tr("mcp_choice_api", "สำหรับโปรแกรมภายนอกที่รับส่งข้อมูลบัญชีผ่าน REST API โดยใช้สิทธิ์แยกแต่ละการเชื่อมต่อ")}</p></div><Button className="mt-2 justify-self-start" size="lg" variant="outline" disabled={busy || loading} onClick={() => void choose(null, "api")}>{tr("mcp_connect_app", "เชื่อมต่อโปรแกรม")}<ArrowUpRight /></Button></div></div>
          <div className="mt-6 grid gap-3 text-sm text-muted-foreground"><p className="flex items-center gap-2"><LockKeyhole className="size-4 shrink-0 text-primary" />{tr("mcp_kind_separate", "API token และ MCP token แยกกัน ใช้ข้ามช่องทางไม่ได้")}</p><p>{tr("mcp_welcome_permissions", "readonly อ่านได้หลายบริษัทพร้อมกัน · readwrite ทำงานครั้งละหนึ่งบริษัท")}</p></div>
        </div>}
        {(creating || selected) && <p className="mt-4 text-sm text-muted-foreground">{tr("mcp_kind_separate", "API token และ MCP token แยกกัน ใช้ข้ามช่องทางไม่ได้")}</p>}
      </section>
    </div>}{confirmationDialog}
  </main>;
}

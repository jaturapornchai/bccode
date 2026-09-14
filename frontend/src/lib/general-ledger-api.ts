import { authFetch, getAuthSession, restoreAuthSession } from "./client-auth-session";
import type { GLCommand, GLPage, GLRecord, GLResource } from "./general-ledger";

const projectionWaitMs = 15_000;
const projectionBackoffMs = [100, 250, 500, 1000];
const projectionWaitMessage = "ข้อมูลรายงานยังปรับปรุงไม่เสร็จ กรุณารอสักครู่แล้วกดโหลดใหม่";
const thaiText = /[\u0E01-\u0E5B]/;
const genericCommandMessage = "ทำรายการไม่สำเร็จ กรุณาลองใหม่ หากยังไม่ได้ให้ติดต่อผู้ดูแลระบบ";

/**
 * Thai fallback per stable backend `code` (backend/internal/generalledger/errors.go).
 * The Thai text from the server always wins; this map only covers a missing/empty `message`.
 */
const commandCodeMessages: Record<string, string> = {
  duplicate_code: "รหัสบัญชีซ้ำกับที่มีอยู่แล้วในผังบัญชี กรุณาใช้รหัสอื่น",
  duplicate_request: "คำขอนี้ถูกบันทึกไปแล้ว กรุณาตรวจสอบข้อมูลล่าสุด",
  validation_failed: "ข้อมูลไม่ครบถ้วนหรือไม่ถูกต้อง กรุณาตรวจสอบช่องที่พิมพ์อีกครั้ง",
  stale_version: "ข้อมูลถูกแก้ไขโดยผู้ใช้อื่น กรุณาโหลดใหม่แล้วทำรายการอีกครั้ง",
  parent_not_found: "ไม่พบรหัสบัญชีแม่ที่เลือก กรุณาเลือกใหม่",
  parent_invalid: "บัญชีแม่ไม่ถูกต้อง (ต้องไม่ใช่บัญชีตัวเองหรือบัญชีลูกของตัวเอง)",
  account_tree_invalid: "โครงสร้างบัญชีซ้ำหรือมีระดับเกิน 12 ระดับ กรุณาเลือกบัญชีแม่อื่น",
  level_out_of_range: "ระดับบัญชีต้องอยู่ระหว่าง 1 ถึง 12",
  level_not_deeper_than_parent: "ระดับบัญชีต้องลึกกว่าบัญชีแม่อย่างน้อย 1 ระดับ",
  account_has_children: "บัญชีนี้มีบัญชีลูกอยู่ ลบไม่ได้ กรุณาลบบัญชีลูกก่อน",
  account_referenced: "บัญชีนี้ถูกใช้ในสมุดรายวันแล้ว ลบไม่ได้ กรุณาปิดใช้งานแทน",
  account_referenced_master: "บัญชีนี้ถูกใช้ในการตั้งค่าอื่น ลบไม่ได้ กรุณาปิดใช้งานแทน",
  account_posted_immutable: "บัญชีนี้ผ่านรายการแล้ว แก้ไขหมวดบัญชีหรือโครงสร้างไม่ได้",
  code_immutable: "แก้ไขรหัสบัญชีไม่ได้ กรุณาสร้างบัญชีใหม่",
  account_group_not_found: "ไม่พบรหัสกลุ่มผังบัญชีนี้ในระบบ กรุณาตรวจสอบอีกครั้ง",
  account_payload_required: "ไม่พบข้อมูลบัญชีที่ส่งมา กรุณาลองใหม่อีกครั้ง",
  unsupported_command: "ระบบไม่รองรับคำสั่งนี้ กรุณาแจ้งผู้ดูแลระบบ",
  not_found: "ไม่พบข้อมูลที่ต้องการ อาจถูกลบไปแล้ว กรุณาโหลดใหม่",
  projection_pending: "ข้อมูลรายงานยังปรับปรุงไม่เสร็จ กรุณารอสักครู่แล้วกดโหลดใหม่",
  unavailable: "ระบบปลายทางไม่พร้อมใช้งานชั่วคราว กรุณาลองใหม่อีกครั้ง",
};
/** Codes that point at a specific input so the screen can focus it. */
const commandCodeFields: Record<string, string> = {
  duplicate_code: "accountcode",
  code_immutable: "accountcode",
  validation_failed: "accountcode",
  parent_not_found: "parentaccountcode",
  parent_invalid: "parentaccountcode",
  account_tree_invalid: "parentaccountcode",
  level_not_deeper_than_parent: "level",
};

export class GLCommandError extends Error {
  readonly code: string;
  readonly field: string;
  constructor(message: string, code = "", field = "") {
    super(message);
    this.name = "GLCommandError";
    this.code = code;
    this.field = field;
  }
}

/**
 * Never surface raw provider/HTTP text: use the server's Thai `message`, else the Thai fallback for
 * the machine code, else a generic Thai sentence. Exported for unit tests.
 */
/**
 * Maps any thrown failure to what the editor pane must show: Thai text + the input to focus.
 * Server Thai `message` wins (already carried by GLCommandError), otherwise the caller's Thai
 * fallback is used — a raw English/technical error text is never surfaced.
 */
export function commandFailure(cause: unknown, fallback = genericCommandMessage, field = ""): { message: string; field: string } {
  const raw = cause instanceof Error ? cause.message : "";
  const message = raw && thaiText.test(raw) ? raw : fallback;
  return { message, field: field || (cause instanceof GLCommandError ? cause.field : "") };
}

export function commandErrorInfo(payload: unknown, status: number): { code: string; message: string; field: string } {
  const body = (payload && typeof payload === "object" ? payload : {}) as Record<string, unknown>;
  const nested = (body.error && typeof body.error === "object" ? body.error : {}) as Record<string, unknown>;
  const code = typeof body.code === "string" ? body.code : typeof body.errorcode === "string" ? body.errorcode : "";
  const raw = typeof body.message === "string" ? body.message : typeof nested.message === "string" ? nested.message : "";
  const serverThai = raw.trim() && thaiText.test(raw) ? raw.trim() : "";
  const message = serverThai || commandCodeMessages[code] || (status === 409 ? "ข้อมูลถูกแก้ไขโดยผู้ใช้อื่น กรุณาโหลดใหม่" : genericCommandMessage);
  return { code, message, field: commandCodeFields[code] ?? "" };
}

function waitForProjection(delay: number, signal: AbortSignal) {
  return new Promise<void>((resolve, reject) => {
    signal.throwIfAborted();
    const onAbort = () => { clearTimeout(timer); reject(signal.reason); };
    const timer = setTimeout(() => { signal.removeEventListener("abort", onAbort); resolve(); }, delay);
    signal.addEventListener("abort", onAbort, { once: true });
  });
}

export async function glRequest<T>(path: string, init?: RequestInit): Promise<T> {
  init?.signal?.throwIfAborted();
  const auth = getAuthSession() ?? await restoreAuthSession();
  if (!auth?.token) throw new Error("กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนใช้งานบัญชี");
  const isRead = (init?.method ?? "GET").toUpperCase() === "GET";
  const deadline = Date.now() + projectionWaitMs;
  let attempt = 0, signal = init?.signal;
  let projectionTimeout: ReturnType<typeof setTimeout> | undefined;
  let projectionController: AbortController | undefined;
  try {
    for (;;) {
      signal?.throwIfAborted();
      const response = await authFetch(`/api/gl/${path}`, { ...init, signal, cache: "no-store", headers: {
        Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl,
        "Content-Type": "application/json", ...init?.headers,
      } });
      const payload = await response.json().catch(() => ({})) as { success?: boolean; data: T; message?: string; code?: string; errorcode?: string; error?: { message?: string } };
      signal?.throwIfAborted();
      if (isRead && response.status === 409 && payload.errorcode === "GL_PROJECTION_PENDING") {
        const remaining = deadline - Date.now();
        if (remaining <= 0) throw new Error(projectionWaitMessage);
        if (!projectionController) {
          // Bound only projection retries; abort their in-flight fetch as well as backoff waits.
          projectionController = new AbortController();
          projectionTimeout = setTimeout(() => projectionController!.abort(), remaining);
          signal = init?.signal ? AbortSignal.any([init.signal, projectionController.signal]) : projectionController.signal;
        }
        await waitForProjection(Math.min(projectionBackoffMs[Math.min(attempt++, projectionBackoffMs.length - 1)], remaining), signal!);
        continue;
      }
      if (!response.ok || payload.success === false) {
        const info = commandErrorInfo(payload, response.status);
        throw new GLCommandError(info.message, info.code, info.field);
      }
      return payload.data;
    }
  } catch (error) {
    if (projectionController?.signal.aborted && !init?.signal?.aborted) throw new Error(projectionWaitMessage);
    throw error;
  } finally {
    clearTimeout(projectionTimeout);
  }
}
export function glCommand(command: Omit<GLCommand, "requestid">, requestid: string) {
  return glRequest<{ id: string; version: number; projectionpending?: boolean; createdjournals?: number }>("command", { method: "POST", body: JSON.stringify({ ...command, requestid }) });
}
/** Abort instead of silently producing a partial accounting export. */
export async function glAllRecords<T extends GLRecord>(resource: GLResource, query = "", cap = 100000, snapshot?: number): Promise<T[]> {
  const items: T[] = []; let expected: number | undefined; let sequence = snapshot;
  for (let page = 1; ; page++) {
    const data = await glRequest<GLPage<T>>(`${resource}?${new URLSearchParams({ q: query, page: String(page), limit: "500", ...(sequence === undefined ? {} : { snapshot: String(sequence) }) })}`);
    if (sequence !== undefined && sequence !== data.sequence) throw new Error("ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่");
    sequence = data.sequence;
    if (expected !== undefined && data.total !== expected) throw new Error("ข้อมูลเปลี่ยนแปลงระหว่างส่งออก กรุณาลองใหม่");
    expected = data.total;
    if (expected > cap) throw new Error(`ข้อมูลเกิน ${cap.toLocaleString("th-TH")} รายการ กรุณาจำกัดช่วงข้อมูลก่อนส่งออก`);
    items.push(...(data.items ?? []));
    if (items.length >= expected) break;
    if (!data.items?.length) throw new Error("ได้รับข้อมูลไม่ครบ กรุณาลองส่งออกใหม่");
  }
  if (new Set(items.map((item) => item.id)).size !== items.length) throw new Error("พบข้อมูลซ้ำระหว่างส่งออก กรุณาลองใหม่");
  return items;
}

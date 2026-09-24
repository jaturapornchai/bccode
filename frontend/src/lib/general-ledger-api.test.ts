import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
vi.mock("./client-auth-session", () => ({ getAuthSession: () => ({ token: "test", backendUrl: "http://localhost:8888" }), restoreAuthSession: vi.fn(), authFetch: (...args: unknown[]) => fetch(...args as Parameters<typeof fetch>) }));
import { GLCommandError, commandErrorInfo, commandFailure, glAllRecords, glCommand, glRequest } from "./general-ledger-api";

describe("general ledger API pagination", () => {
  afterEach(() => vi.unstubAllGlobals());
  it("reads every page under the same snapshot", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(Response.json({ success: true, data: { items: [{ id: "a" }], total: 2, sequence: 7 } })).mockResolvedValueOnce(Response.json({ success: true, data: { items: [{ id: "b" }], total: 2, sequence: 7 } }));
    vi.stubGlobal("fetch", fetchMock);
    expect(await glAllRecords("accounts")).toEqual([{ id: "a" }, { id: "b" }]);
    expect(fetchMock.mock.calls[1][0]).toContain("snapshot=7");
  });
  it("refuses incomplete, duplicate or changed exports", async () => {
    for (const second of [{ items: [], total: 2, sequence: 7 }, { items: [{ id: "a" }], total: 2, sequence: 7 }, { items: [{ id: "b" }], total: 2, sequence: 8 }]) {
      vi.stubGlobal("fetch", vi.fn().mockResolvedValueOnce(Response.json({ success: true, data: { items: [{ id: "a" }], total: 2, sequence: 7 } })).mockResolvedValueOnce(Response.json({ success: true, data: second })));
      await expect(glAllRecords("accounts")).rejects.toThrow();
    }
  });
  it("aborts above cap and sanitizes server diagnostics", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: true, data: { items: [], total: 100001, sequence: 1 } })));
    await expect(glAllRecords("accounts")).rejects.toThrow("เกิน");
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: false, message: "SQL error SELECT private" }, { status: 500 })));
    await expect(glRequest("accounts")).rejects.toThrow("ติดต่อผู้ดูแลระบบ");
  });
});


describe("general ledger projection readiness", () => {
  const pending = () => Response.json({ success: false, errorcode: "GL_PROJECTION_PENDING", message: "กำลังปรับปรุงข้อมูลรายงาน" }, { status: 409 });
  beforeEach(() => { vi.useFakeTimers(); vi.setSystemTime(0); });
  afterEach(() => { vi.unstubAllGlobals(); vi.useRealTimers(); });

  it("waits with bounded backoff until the GET projection is ready", async () => {
    const fetchMock = vi.fn().mockImplementationOnce(pending).mockImplementationOnce(pending).mockImplementationOnce(pending).mockImplementationOnce(pending)
      .mockResolvedValueOnce(Response.json({ success: true, data: { balance: "0.30000000" } }));
    vi.stubGlobal("fetch", fetchMock);
    const result = expect(glRequest("reports/trialbalance?fiscalyear=2569")).resolves.toEqual({ balance: "0.30000000" });
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    for (const [index, delay] of [100, 250, 500, 1000].entries()) {
      await vi.advanceTimersByTimeAsync(delay - 1);
      expect(fetchMock).toHaveBeenCalledTimes(index + 1);
      await vi.advanceTimersByTimeAsync(1);
      expect(fetchMock).toHaveBeenCalledTimes(index + 2);
    }
    await result;
    expect(fetchMock.mock.calls.every(([url]) => url === "/api/gl/reports/trialbalance?fiscalyear=2569")).toBe(true);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("stops pending retries after 15 seconds with actionable Thai feedback", async () => {
    const fetchMock = vi.fn().mockImplementation(pending);
    vi.stubGlobal("fetch", fetchMock);
    const result = expect(glRequest("accounts")).rejects.toThrow("ข้อมูลรายงานยังปรับปรุงไม่เสร็จ");
    await vi.advanceTimersByTimeAsync(15_000);
    await result;
    const calls = fetchMock.mock.calls.length;
    expect(calls).toBeGreaterThan(1);
    await vi.advanceTimersByTimeAsync(30_000);
    expect(fetchMock).toHaveBeenCalledTimes(calls);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("also bounds an in-flight retry request", async () => {
    const fetchMock = vi.fn().mockImplementationOnce(pending).mockImplementationOnce((_url: string, init: RequestInit) => new Promise((_resolve, reject) => {
      init.signal!.addEventListener("abort", () => reject(init.signal!.reason), { once: true });
    }));
    vi.stubGlobal("fetch", fetchMock);
    const result = expect(glRequest("accounts")).rejects.toThrow("ข้อมูลรายงานยังปรับปรุงไม่เสร็จ");
    await vi.advanceTimersByTimeAsync(15_000);
    await result;
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(vi.getTimerCount()).toBe(0);
  });

  it.each([
    [409, undefined], [409, "GL_SNAPSHOT_CHANGED"], [503, "GL_PROJECTION_PENDING"],
  ])("does not retry HTTP %s with code %s", async (status, errorcode) => {
    const fetchMock = vi.fn().mockImplementation(() => Response.json({ success: false, errorcode }, { status }));
    vi.stubGlobal("fetch", fetchMock);
    await expect(glRequest("accounts")).rejects.toThrow();
    await vi.advanceTimersByTimeAsync(15_000);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("never retries a command or replaces its idempotency key", async () => {
    const requestid = "00000000-0000-4000-8000-000000000001";
    const command = { resource: "journals" as const, action: "post" as const, id: "draft-1", version: 1 };
    const fetchMock = vi.fn().mockImplementation(pending);
    vi.stubGlobal("fetch", fetchMock);
    await expect(glCommand(command, requestid)).rejects.toThrow("กำลังปรับปรุงข้อมูลรายงาน");
    await vi.advanceTimersByTimeAsync(15_000);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock.mock.calls[0][1].method).toBe("POST");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ ...command, requestid });
    expect(vi.getTimerCount()).toBe(0);
  });

  it("honors caller cancellation while waiting without another request", async () => {
    const controller = new AbortController(), reason = new Error("cancelled by caller");
    const fetchMock = vi.fn().mockImplementation(pending);
    vi.stubGlobal("fetch", fetchMock);
    const result = expect(glRequest("accounts", { signal: controller.signal })).rejects.toBe(reason);
    await vi.advanceTimersByTimeAsync(50);
    controller.abort(reason);
    await result;
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("does not start a request when already aborted", async () => {
    const controller = new AbortController(), fetchMock = vi.fn();
    controller.abort(); vi.stubGlobal("fetch", fetchMock);
    await expect(glRequest("accounts", { signal: controller.signal })).rejects.toMatchObject({ name: "AbortError" });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

describe("command failure -> Thai message for the editor pane", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("uses the server Thai message and code for a duplicate account code (409)", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: false, code: "duplicate_code", message: "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น" }, { status: 409 })));
    const error = await glCommand({ resource: "accounts", action: "create", reason: "test" }, "req-1").catch((e) => e);
    expect(error).toBeInstanceOf(GLCommandError);
    expect(error.message).toBe("รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น");
    expect(error.code).toBe("duplicate_code");
    expect(error.field).toBe("accountcode");
  });

  it("falls back to Thai when the server omits the message or sends an unknown code", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: false, code: "duplicate_code" }, { status: 409 })));
    await expect(glCommand({ resource: "accounts", action: "create" }, "req-1")).rejects.toThrow("ซ้ำ");
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: false, code: "brand_new_code" }, { status: 409 })));
    const error = await glCommand({ resource: "accounts", action: "create" }, "req-1").catch((e) => e);
    expect(error.message).toBe(commandErrorInfo({}, 409).message);
    expect(/[ก-๛]/.test(error.message)).toBe(true);
  });

  it("never shows raw English/technical provider text", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(Response.json({ success: false, code: "unavailable", message: "E11000 duplicate key error collection: appdb.chart_of_accounts" }, { status: 503 })));
    const error = await glCommand({ resource: "accounts", action: "create" }, "req-1").catch((e) => e);
    expect(error.message).not.toContain("E11000");
    expect(/[ก-๛]/.test(error.message)).toBe(true);
  });

  // UAT S3 2026-09-24: browser Accept-Language en-US → backend message was English and the screen fell back to the generic text
  it("shows message_th when the server message is not Thai and sends the app language", async () => {
    vi.stubGlobal("localStorage", { getItem: (key: string) => (key === "user_language" ? "th" : null) });
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ success: false, code: "journal_book_in_use_delete", field: "code", message: "This journal book is used by journals", message_th: "สมุดรายวันนี้มีใบสำคัญใช้อยู่ ลบไม่ได้ — ปิดใช้งานแทน" }, { status: 409 }));
    vi.stubGlobal("fetch", fetchMock);
    const error = await glCommand({ resource: "journal-books", action: "delete" }, "req-1").catch((e) => e);
    expect(error.message).toBe("สมุดรายวันนี้มีใบสำคัญใช้อยู่ ลบไม่ได้ — ปิดใช้งานแทน");
    expect(error.code).toBe("journal_book_in_use_delete");
    expect(error.field).toBe("code");
    expect((fetchMock.mock.calls[0][1] as RequestInit).headers).toMatchObject({ "Accept-Language": "th" });
  });

  it("maps a non-JSON error body to Thai instead of a parse error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("<html>Bad Gateway</html>", { status: 502 })));
    const error = await glCommand({ resource: "accounts", action: "create" }, "req-1").catch((e) => e);
    expect(error.message).not.toContain("Unexpected token");
    expect(commandFailure(error, "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง").field).toBe("");
  });
});

describe("commandFailure focuses the offending input", () => {
  it("keeps the plain-Thai server message and its field", () => {
    expect(commandFailure(new GLCommandError("ไม่พบบัญชีแม่", "parent_not_found", "parentaccountcode"))).toEqual({ message: "ไม่พบบัญชีแม่", field: "parentaccountcode" });
  });
  it("replaces non-Thai text with the Thai fallback", () => {
    expect(commandFailure(new Error("fetch failed"), "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง")).toEqual({ message: "บันทึกไม่สำเร็จ กรุณาลองใหม่อีกครั้ง", field: "" });
  });
});

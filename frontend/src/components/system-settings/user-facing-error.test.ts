import { afterEach, describe, expect, it, vi } from "vitest";
import {
  dictionaryMessage,
  offersExistingLogin,
  requestErrorText,
  settingsRequestError,
  userFacingErrorText,
} from "./user-facing-error";

const th = {
  ss_err_invalid_data: "ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบแล้วลองใหม่อีกครั้ง",
  ss_err_session_expired: "หมดเวลาการเข้าใช้งาน กรุณาเข้าสู่ระบบใหม่อีกครั้ง",
  ss_err_no_permission: "คุณไม่มีสิทธิ์ทำรายการนี้ กรุณาติดต่อเจ้าของหรือผู้ดูแลกลุ่มกิจการ",
  ss_err_not_found: "ไม่พบข้อมูลที่ต้องการ อาจถูกลบไปแล้ว กรุณารีเฟรชรายการแล้วลองใหม่",
  ss_err_duplicate: "มีข้อมูลนี้อยู่ในระบบแล้ว (รหัสซ้ำ) กรุณาใช้รหัสอื่น",
  ss_err_conflict: "ทำรายการนี้ไม่ได้ในตอนนี้",
  ss_err_try_again: "ระบบขัดข้องชั่วคราว กรุณาลองใหม่อีกครั้ง",
  ss_err_cannot_delete_self: "คุณลบบัญชีเข้าระบบของตัวเองไม่ได้",
  creator_cannot_delete: "ผู้สร้างไม่สามารถลบได้",
};
const en = {
  ss_err_invalid_data: "The information is not valid. Please check it and try again.",
  ss_err_no_permission: "You don't have permission to do this.",
  ss_err_not_found: "The record could not be found.",
  ss_err_try_again: "Something went wrong. Please try again.",
  ss_err_cannot_delete_self: "You can't remove your own login account.",
};

describe("user-facing settings errors", () => {
  afterEach(() => vi.restoreAllMocks());

  it("translates a message that is a dictionary key", () => {
    const error = settingsRequestError(409, { success: false, message: "creator_cannot_delete" });
    expect(userFacingErrorText(error, "th", th)).toBe(th.creator_cannot_delete);
  });

  it("maps an apperr code to its row in the active language", () => {
    const payload = { success: false, errorcode: "FORBIDDEN", code: "FORBIDDEN", message: "forbidden", message_th: "ไม่มีสิทธิ์" };
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    expect(userFacingErrorText(settingsRequestError(403, payload), "th", th)).toBe(th.ss_err_no_permission);
    expect(userFacingErrorText(settingsRequestError(403, payload), "en", en)).toBe(en.ss_err_no_permission);
  });

  it("shows a specific Thai message_th before the generic errorcode row, only in Thai", () => {
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    const payload = {
      success: false,
      errorcode: "FORBIDDEN",
      message: "holding is required",
      message_th: "กรุณาเลือกกลุ่มกิจการก่อน",
    };
    expect(userFacingErrorText(settingsRequestError(403, payload), "th", th)).toBe("กรุณาเลือกกลุ่มกิจการก่อน");
    expect(userFacingErrorText(settingsRequestError(403, payload), "en", en)).toBe(en.ss_err_no_permission);
    // The bare apperr default only restates the code: the row that says what to do next wins.
    const generic = { success: false, errorcode: "NOT_FOUND", message: "record not found", message_th: "ไม่พบข้อมูล" };
    expect(userFacingErrorText(settingsRequestError(404, generic), "th", th)).toBe(th.ss_err_not_found);
    // Technical tails stay hidden even when the sentence starts in Thai.
    const technical = { success: false, errorcode: "INTERNAL_ERROR", message_th: "บันทึกไม่สำเร็จ: context canceled" };
    expect(userFacingErrorText(settingsRequestError(500, technical), "th", th)).toBe(th.ss_err_try_again);
  });

  it("treats decoder, driver and cancellation tails as technical", () => {
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    for (const raw of [
      "context canceled",
      "unexpected EOF",
      "EOF",
      "invalid UUID length: 5",
      "E11000 duplicate key { __v: 0 }",
    ]) {
      expect(userFacingErrorText(new Error(raw), "th", th)).toBe(th.ss_err_try_again);
      expect(requestErrorText({ status: 400, message: `ข้อมูล: ${raw}` }, "th", th)).toBe(th.ss_err_invalid_data);
    }
    // Ordinary words that merely contain the letters are not technical.
    expect(userFacingErrorText(new Error("Geoffrey's branch"), "en", en)).toBe("Geoffrey's branch");
  });

  it("hides Safari's 'Load failed' but keeps the screen's own upload-failed text", () => {
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    expect(userFacingErrorText(new TypeError("Load failed"), "en", en)).toBe(en.ss_err_try_again);
    expect(userFacingErrorText(new Error("load failed"), "th", th)).toBe(th.ss_err_try_again);
    const upload = "Image upload failed. Please choose a smaller PNG or JPG file.";
    expect(userFacingErrorText(new Error(upload), "en", en)).toBe(upload);
    const uploadTh = "อัปโหลดรูปไม่สำเร็จ (upload failed) กรุณาเลือกไฟล์ใหม่";
    expect(userFacingErrorText(new Error(uploadTh), "th", th)).toBe(uploadTh);
  });

  it("shows a backend row rendered in the active language, from message or message_th", () => {
    const payload = {
      success: false,
      errorcode: "CONFLICT",
      message: en.ss_err_cannot_delete_self,
      message_th: th.ss_err_cannot_delete_self,
    };
    expect(userFacingErrorText(settingsRequestError(409, payload), "th", th)).toBe(th.ss_err_cannot_delete_self);
    expect(userFacingErrorText(settingsRequestError(409, payload), "en", en)).toBe(en.ss_err_cannot_delete_self);
  });

  it("shows plain Thai backend text only when the screen is in Thai", () => {
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    const details = { status: 400, message: "กรุณาระบุชื่อธนาคาร" };
    expect(requestErrorText(details, "th", th)).toBe("กรุณาระบุชื่อธนาคาร");
    expect(requestErrorText(details, "en", en)).toBe(en.ss_err_invalid_data);
  });

  it("hides the known raw strings and logs them for developers", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    const unmarshal = settingsRequestError(400, {
      success: false,
      message: "json: cannot unmarshal object into Go struct field UserRoleRequest.accessscopes of type []models.AccessScope",
    });
    expect(userFacingErrorText(unmarshal, "th", th)).toBe(th.ss_err_invalid_data);
    expect(userFacingErrorText(settingsRequestError(400, { success: false, message: "find failed" }), "th", th)).toBe(
      th.ss_err_invalid_data,
    );
    expect(requestErrorText({ status: 400, message: "บันทึกไม่สำเร็จ: json: cannot unmarshal" }, "th", th)).toBe(
      th.ss_err_invalid_data,
    );
    expect(warn).toHaveBeenCalled();
  });

  it("maps 5xx, 404, 401 and network failures to what the user can do next", () => {
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    expect(userFacingErrorText(settingsRequestError(500, { message: "pq: connection refused" }), "th", th)).toBe(
      th.ss_err_try_again,
    );
    expect(userFacingErrorText(settingsRequestError(502, "<html>Bad gateway</html>"), "en", en)).toBe(en.ss_err_try_again);
    expect(userFacingErrorText(settingsRequestError(404, {}), "th", th)).toBe(th.ss_err_not_found);
    expect(userFacingErrorText(settingsRequestError(401, { message: "unauthorized request" }), "th", th)).toBe(
      th.ss_err_session_expired,
    );
    expect(userFacingErrorText(new TypeError("Failed to fetch"), "th", th)).toBe(th.ss_err_try_again);
    expect(userFacingErrorText(new SyntaxError("Unexpected token < in JSON"), "en", en)).toBe(en.ss_err_try_again);
  });

  it("keeps messages the screen built itself and falls back when empty", () => {
    expect(userFacingErrorText(new Error("รหัสสาขาภาษีไทยต้องเป็นเลขไม่เกิน 5 หลัก"), "th", th)).toBe(
      "รหัสสาขาภาษีไทยต้องเป็นเลขไม่เกิน 5 หลัก",
    );
    expect(userFacingErrorText(new Error(""), "th", th, "เรียกข้อมูลไม่สำเร็จ")).toBe("เรียกข้อมูลไม่สำเร็จ");
  });

  it("success toasts only translate dictionary keys", () => {
    expect(dictionaryMessage(th, "creator_cannot_delete")).toBe(th.creator_cannot_delete);
    expect(dictionaryMessage(th, "saved 3 units into productunit")).toBe("");
  });
});

// เพิ่มบัญชีเข้าระบบ: 409 ที่ช่องรหัสผู้ใช้ = รหัสนี้มีบัญชีอยู่แล้ว → ต้องถามก่อนส่ง addexistinguser (ยกเว้น "เป็นสมาชิกอยู่แล้ว")
describe("offersExistingLogin", () => {
  const attachable = {
    success: false,
    errorcode: "LOGIN_EXISTS",
    code: "LOGIN_EXISTS",
    field: "username",
    message: "รหัสผู้ใช้นี้มีบัญชีเข้าระบบอยู่แล้ว",
    message_th: "รหัสผู้ใช้นี้มีบัญชีเข้าระบบอยู่แล้ว",
  };

  it("offers to attach only the login account the backend can attach (LOGIN_EXISTS)", () => {
    expect(offersExistingLogin(409, attachable)).toBe(true);
  });

  // review 2026-09-24: DUPLICATE = used by another business group / several accounts / already a member —
  // the backend refuses the attach, so the dialog must not promise it
  it("does not offer it for DUPLICATE, another field or another status", () => {
    const taken = { ...attachable, errorcode: "DUPLICATE", code: "DUPLICATE", message: "ช่อง รหัสผู้ใช้ ซ้ำกับบัญชีที่มีอยู่แล้วในระบบ" };
    expect(offersExistingLogin(409, taken)).toBe(false);
    expect(offersExistingLogin(409, { ...attachable, field: "email" })).toBe(false);
    expect(offersExistingLogin(409, { ...attachable, errorcode: "CONFLICT", code: "CONFLICT" })).toBe(false);
    expect(offersExistingLogin(400, attachable)).toBe(false);
    expect(offersExistingLogin(409, "duplicate")).toBe(false);
  });
});

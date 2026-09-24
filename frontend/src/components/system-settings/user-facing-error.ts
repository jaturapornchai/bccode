// Settings screens never show raw backend/technical errors (owner rule: users are Thai, 40+, and
// every error must say what to do next). A failed request keeps its HTTP status and apperr code
// so userFacingErrorText can pick a plain-language languages.tsv row instead of the raw text.
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import type { LanguageCode } from "@/lib/i18n";
import { extractErrorCode, extractMessage, isRecord } from "./types";

/** A failed settings API call: HTTP status, apperr code and both backend messages. */
export class SettingsRequestError extends Error {
  readonly status: number;
  readonly code: string;
  readonly thaiMessage: string;

  constructor(status: number, code: string, message: string, thaiMessage = "") {
    super(message);
    this.name = "SettingsRequestError";
    this.status = status;
    this.code = code;
    this.thaiMessage = thaiMessage;
  }
}

/** Error for a failed response; reads ctx.ResponseError {message} and apperr {errorcode, message, message_th}. */
export function settingsRequestError(status: number, payload: unknown): SettingsRequestError {
  const thaiMessage = isRecord(payload) && typeof payload.message_th === "string" ? payload.message_th : "";
  return new SettingsRequestError(status, extractErrorCode(payload) ?? "", extractMessage(payload) ?? "", thaiMessage);
}

/** Invalid JSON typed into a JSON field; fieldLabel names the field for the user. */
export class JsonFieldError extends Error {
  readonly fieldLabel: string;

  constructor(fieldLabel: string) {
    super(fieldLabel);
    this.name = "JsonFieldError";
    this.fieldLabel = fieldLabel;
  }
}

// languages.tsv rows (English fallback only for a dictionary that has not loaded yet).
const errorRows = {
  invalidData: ["ss_err_invalid_data", "The information is not valid. Please check it and try again."],
  sessionExpired: ["ss_err_session_expired", "Your session has expired. Please sign in again."],
  noPermission: ["ss_err_no_permission", "You don't have permission to do this. Please contact the holding owner or an administrator."],
  notFound: ["ss_err_not_found", "The record could not be found. It may have been deleted. Please refresh the list and try again."],
  duplicate: ["ss_err_duplicate", "This record already exists (duplicate code). Please use a different code."],
  conflict: ["ss_err_conflict", "This can't be done right now. The data may be in use or may have changed. Please refresh and check again."],
  tryAgain: ["ss_err_try_again", "Something went wrong. Please try again. If the problem continues, contact your system administrator."],
} as const;
type ErrorRow = keyof typeof errorRows;

// apperr codes (backend/pkg/apperr/codes.go) → row.
const codeRows: Record<string, ErrorRow> = {
  BAD_REQUEST: "invalidData",
  VALIDATION_FAILED: "invalidData",
  UNAUTHORIZED: "sessionExpired",
  FORBIDDEN: "noPermission",
  DISABLED: "noPermission",
  NOT_FOUND: "notFound",
  DUPLICATE: "duplicate",
  CONFLICT: "conflict",
  INTERNAL_ERROR: "tryAgain",
  DEPENDENCY_FAILED: "tryAgain",
  RATE_LIMITED: "tryAgain",
};

function statusRow(status: number): ErrorRow {
  if (status === 400 || status === 422) return "invalidData";
  if (status === 401) return "sessionExpired";
  if (status === 403) return "noPermission";
  if (status === 404) return "notFound";
  if (status === 409) return "conflict";
  return "tryAgain";
}

// Decoder, driver, network and runtime text that must never reach the user.
// "load failed" is Safari fetch rejection text; the word boundary keeps translated upload errors
// (e.g. "Image upload failed") from being mistaken for it.
const technicalPattern =
  /json:|invalid character|unexpected end of json|unexpected token|cannot unmarshal|go struct field|parsing time|\bpq:|\bsql:|sqlstate|connect central database|find failed|syntaxerror|typeerror|failed to fetch|networkerror|\bload failed\b|econnrefused|dial tcp|context deadline|context canceled|\beof\b|invalid uuid|__v\b|runtime error|nil pointer|\bhttp \d{3}\b/i;
const thaiPattern = /[\u0E00-\u0E7F]/;

// The default message_th of the bare apperr codes (backend/pkg/apperr/codes.go). They only
// restate the code, so the languages.tsv row (which says what to do next) is shown instead.
const genericThaiMessages = new Set([
  "\u0E44\u0E21\u0E48\u0E1E\u0E1A\u0E02\u0E49\u0E2D\u0E21\u0E39\u0E25",
  "\u0E23\u0E2B\u0E31\u0E2A\u0E0B\u0E49\u0E33",
  "\u0E02\u0E49\u0E2D\u0E21\u0E39\u0E25\u0E44\u0E21\u0E48\u0E16\u0E39\u0E01\u0E15\u0E49\u0E2D\u0E07",
  "\u0E44\u0E21\u0E48\u0E44\u0E14\u0E49\u0E23\u0E31\u0E1A\u0E2D\u0E19\u0E38\u0E0D\u0E32\u0E15",
  "\u0E44\u0E21\u0E48\u0E21\u0E35\u0E2A\u0E34\u0E17\u0E18\u0E34\u0E4C",
  "\u0E40\u0E01\u0E34\u0E14\u0E02\u0E49\u0E2D\u0E1C\u0E34\u0E14\u0E1E\u0E25\u0E32\u0E14\u0E43\u0E19\u0E23\u0E30\u0E1A\u0E1A",
  "\u0E04\u0E33\u0E02\u0E2D\u0E44\u0E21\u0E48\u0E16\u0E39\u0E01\u0E15\u0E49\u0E2D\u0E07",
  "\u0E44\u0E21\u0E48\u0E2A\u0E32\u0E21\u0E32\u0E23\u0E16\u0E14\u0E33\u0E40\u0E19\u0E34\u0E19\u0E01\u0E32\u0E23\u0E44\u0E14\u0E49\u0E40\u0E19\u0E37\u0E48\u0E2D\u0E07\u0E08\u0E32\u0E01\u0E2A\u0E16\u0E32\u0E19\u0E30\u0E1B\u0E31\u0E08\u0E08\u0E38\u0E1A\u0E31\u0E19",
  "\u0E1A\u0E23\u0E34\u0E01\u0E32\u0E23\u0E20\u0E32\u0E22\u0E19\u0E2D\u0E01\u0E40\u0E01\u0E34\u0E14\u0E02\u0E49\u0E2D\u0E1C\u0E34\u0E14\u0E1E\u0E25\u0E32\u0E14",
  "\u0E2A\u0E48\u0E07\u0E04\u0E33\u0E02\u0E2D\u0E16\u0E35\u0E48\u0E40\u0E01\u0E34\u0E19\u0E44\u0E1B",
  "\u0E16\u0E39\u0E01\u0E23\u0E30\u0E07\u0E31\u0E1A\u0E01\u0E32\u0E23\u0E43\u0E0A\u0E49\u0E07\u0E32\u0E19",
  "\u0E2B\u0E21\u0E14\u0E2D\u0E32\u0E22\u0E38\u0E41\u0E25\u0E49\u0E27",
]);

function isTechnicalErrorText(text: string): boolean {
  return technicalPattern.test(text);
}

// A Thai sentence the backend wrote for this case (e.g. "กรุณาเลือกกลุ่มกิจการก่อน"): more
// useful to a Thai user than the generic row of its errorcode.
function isSpecificThaiText(text: string): boolean {
  return thaiPattern.test(text) && !isTechnicalErrorText(text) && !genericThaiMessages.has(text);
}

const dictionaryValues = new WeakMap<BackendLanguageDictionary, Set<string>>();

// A backend message rendered from a languages.tsv row in the active language is already user text.
function isDictionaryValue(dictionary: BackendLanguageDictionary, text: string): boolean {
  let values = dictionaryValues.get(dictionary);
  if (!values) {
    values = new Set(Object.values(dictionary).map((value) => String(value).trim()).filter(Boolean));
    dictionaryValues.set(dictionary, values);
  }
  return values.has(text);
}

function rowText(dictionary: BackendLanguageDictionary, row: ErrorRow): string {
  const [key, fallback] = errorRows[row];
  return backendText(dictionary, key, fallback);
}

/** The text of message when it is a languages.tsv key, otherwise "" (never the raw message). */
export function dictionaryMessage(dictionary: BackendLanguageDictionary, message: unknown): string {
  const key = String(message ?? "").trim();
  return key && dictionary[key] ? backendText(dictionary, key) : "";
}

/**
 * Adding a login account: mainapi answers 409 LOGIN_EXISTS on "username" only when the user code has
 * exactly one login account that no other business group uses — the one case the same request resent
 * with addexistinguser:true attaches (backend/internal/shop/shopuser_postgres_repository.go addMember),
 * so the screen asks first. DUPLICATE (already a member here, used by another group, several accounts)
 * can never be attached: offering it made the dialog promise what the save then refused (review 2026-09-24).
 */
export function offersExistingLogin(status: number, payload: unknown): boolean {
  if (status !== 409 || !isRecord(payload) || payload.field !== "username") return false;
  return String(extractErrorCode(payload) ?? "").toUpperCase() === "LOGIN_EXISTS";
}

function hiddenForUser(raw: string): void {
  // Developers still need the original text; users get the plain-language row.
  if (raw) console.warn("[system-settings] backend error hidden from user:", raw);
}

export type UserFacingErrorDetails = {
  status?: number;
  code?: string;
  message?: string;
  thaiMessage?: string;
};

/** Plain-language text for a failed request (status, apperr code and backend messages). */
export function requestErrorText(
  details: UserFacingErrorDetails,
  language: LanguageCode,
  dictionary: BackendLanguageDictionary,
): string {
  const candidates = [details.message, details.thaiMessage]
    .map((value) => String(value ?? "").trim())
    .filter(Boolean);
  for (const candidate of candidates) {
    const translated = dictionaryMessage(dictionary, candidate);
    if (translated) return translated;
  }
  const vetted = candidates.find((candidate) => !isTechnicalErrorText(candidate) && isDictionaryValue(dictionary, candidate));
  if (vetted) return vetted;
  // Before the errorcode row: a specific Thai sentence beats the generic text of its code.
  if (language === "th") {
    const thai = [details.thaiMessage, details.message]
      .map((value) => String(value ?? "").trim())
      .find(isSpecificThaiText);
    if (thai) return thai;
  }
  hiddenForUser(candidates.join(" | "));
  const codeRow = codeRows[String(details.code ?? "").toUpperCase()];
  return rowText(dictionary, codeRow ?? statusRow(details.status ?? 0));
}

/**
 * Plain-language text for any error a settings screen caught: request errors go through
 * requestErrorText; messages the screen built itself (already translated) are kept; network,
 * unreadable-response and other technical errors become the "try again" row.
 */
export function userFacingErrorText(
  error: unknown,
  language: LanguageCode,
  dictionary: BackendLanguageDictionary,
  fallback?: string,
): string {
  if (error instanceof SettingsRequestError)
    return requestErrorText(
      { status: error.status, code: error.code, message: error.message, thaiMessage: error.thaiMessage },
      language,
      dictionary,
    );
  const message = error instanceof Error ? error.message.trim() : "";
  if (!message) return fallback ?? rowText(dictionary, "tryAgain");
  // fetch() rejects with TypeError, response.json() with SyntaxError, browser APIs with DOMException:
  // anything but a plain Error the screen built itself is technical.
  if (!(error instanceof Error) || error.name !== "Error" || isTechnicalErrorText(message)) {
    hiddenForUser(message);
    return rowText(dictionary, "tryAgain");
  }
  return backendText(dictionary, message, message);
}

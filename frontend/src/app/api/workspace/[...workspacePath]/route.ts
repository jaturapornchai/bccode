import { NextResponse } from "next/server";
import {
  extractMessage,
  getBackendUrlFromRequest,
  getArray,
  getMainApiUrl,
  isRecord,
  proxyMainApiJson,
  readJsonOrText,
  requireBearerToken,
  type ApiProxyBody,
} from "@/lib/workspace-api";
import { holdingCodeValidationMessageTh, isValidHoldingCode, normalizeHoldingCode } from "@/lib/holding-code";

type WorkspaceProxyContext = {
  params: Promise<{ workspacePath: string[] }>;
};

type ProductUnitName = {
  code: string;
  name: string;
  isauto?: boolean;
  isdelete?: boolean;
};

type ProductUnit = {
  unitcode: string;
  names: ProductUnitName[];
};

type HoldingNameEntry = {
  code?: string;
  name?: string;
  isauto?: boolean;
  isdelete?: boolean;
};

type MainApiResult = {
  ok: boolean;
  status: number;
  payload: unknown;
};

const UNIT_TEMPLATE_URL = "https://raw.githubusercontent.com/smlsoft/dedepos_template/main/unit.json";
const API_TIMEOUT_MS = 20000;

export async function GET(request: Request, context: WorkspaceProxyContext) {
  const { workspacePath } = await context.params;
  const path = workspacePath.join("/");

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const url = new URL(request.url);
  switch (path) {
    case "holdings":
      return listHoldingsWithDisplayNames(request, mainApiUrl);
    case "holding-info": {
      const holdingCode = holdingCodeFromSearchParams(url.searchParams);
      const holdingcode = url.searchParams.get("holdingcode")?.trim() || holdingCode;
      if (!holdingCode && !holdingcode) return NextResponse.json({ success: false, message: "ไม่พบรหัส holding" }, { status: 400 });
      return proxyMainApiJson(request, mainApiUrl, `/holding/${encodeURIComponent(holdingcode)}`, { method: "GET" });
    }
    case "branches": {
      const offset = url.searchParams.get("offset") ?? "0";
      const limit = url.searchParams.get("limit") ?? "100";
      const query = url.searchParams.get("q") ?? "";
      const branchPath = `/organization/branch/list?offset=${encodeURIComponent(offset)}&limit=${encodeURIComponent(limit)}&q=${encodeURIComponent(query)}`;
      return proxyMainApiJson(request, mainApiUrl, branchPath, { method: "GET" });
    }
    case "product-units": {
      const offset = url.searchParams.get("offset") ?? "0";
      const limit = url.searchParams.get("limit") ?? "1";
      const query = url.searchParams.get("q") ?? "";
      const unitPath = `/unit/list?offset=${encodeURIComponent(offset)}&limit=${encodeURIComponent(limit)}&q=${encodeURIComponent(query)}&sort=unitcode:1`;
      return proxyMainApiJson(request, mainApiUrl, unitPath, { method: "GET" });
    }
    case "product-units/search": {
      const limit = url.searchParams.get("limit") ?? "1";
      const holdingCode = url.searchParams.get("holdingcode") ?? "";
      const unitPath = `/unit?limit=${encodeURIComponent(limit)}&holdingcode=${encodeURIComponent(holdingCode)}`;
      return proxyMainApiJson(request, mainApiUrl, unitPath, { method: "GET" });
    }
    case "product-units/standard": {
      const mainHoldingCode = url.searchParams.get("mainHoldingCode")?.trim() ?? url.searchParams.get("main_holdingcode")?.trim() ?? "";
      const query = url.searchParams.get("q")?.trim() ?? "";
      const includeExisting = url.searchParams.get("includeExisting") === "true";
      return listMissingStandardProductUnits(request, mainApiUrl, mainHoldingCode, query, includeExisting);
    }
    default:
      return NextResponse.json({ success: false, message: "ไม่พบ workspace endpoint" }, { status: 404 });
  }
}

export async function POST(request: Request, context: WorkspaceProxyContext) {
  const { workspacePath } = await context.params;
  const path = workspacePath.join("/");

  let body: ApiProxyBody;
  try {
    body = (await request.json()) as ApiProxyBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;

  switch (path) {
    case "select-holding": {
      const holdingCode = holdingCodeFromPayload(payload);
      if (!holdingCode) return NextResponse.json({ success: false, message: "ไม่พบรหัส holding" }, { status: 400 });
      return proxyMainApiJson(request, mainApiUrl, "/select-holding", {
        method: "POST",
        body: JSON.stringify({ holdingcode: holdingCode }),
      });
    }
    case "branch": {
      const branch = payload.branch;
      if (!branch || typeof branch !== "object") {
        return NextResponse.json({ success: false, message: "ไม่พบข้อมูลสาขา" }, { status: 400 });
      }
      return proxyMainApiJson(request, mainApiUrl, "/organization/branch", {
        method: "POST",
        body: JSON.stringify(branch),
      });
    }
    case "create-holding": {
      const holdingCode = normalizeHoldingCode(holdingCodeFromPayload(payload));
      if (!holdingCode) {
        return NextResponse.json({ success: false, message: "กรุณากรอกรหัส Holding" }, { status: 400 });
      }
      if (!isValidHoldingCode(holdingCode)) {
        return NextResponse.json(
          { success: false, message: holdingCodeValidationMessageTh },
          { status: 400 },
        );
      }
      payload.holdingcode = holdingCode;
      const holdingName = holdingNameFromPayload(payload);
      if (!holdingName) {
        return NextResponse.json({ success: false, message: "กรุณากรอกชื่อ Holding" }, { status: 400 });
      }
      applyHoldingDisplayName(payload, holdingName);
      return proxyMainApiJson(request, mainApiUrl, "/create-holding", {
        method: "POST",
        body: JSON.stringify(payload),
      });
    }
    case "update-holding":
      return updateHoldingDisplayName(request, mainApiUrl, payload);
    case "product-units/defaults": {
      const mainHoldingCode = getPayloadString(payload, "mainHoldingCode") ?? getPayloadString(payload, "main_holdingcode") ?? getPayloadString(payload, "mainholdingcode") ?? "";
      return createDefaultProductUnits(request, mainApiUrl, mainHoldingCode, getPayloadStringArray(payload, "unitcodes"));
    }
    default:
      return NextResponse.json({ success: false, message: "ไม่พบ workspace endpoint" }, { status: 404 });
  }
}

async function listMissingStandardProductUnits(request: Request, mainApiUrl: string, mainHoldingCode: string, query: string, includeExisting = false): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  try {
    const existing = await loadExistingUnitCodes(request, mainApiUrl, authorization);
    if (!existing.ok) return mainApiError(existing.result, "ตรวจสอบหน่วยนับสินค้าไม่สำเร็จ");

    const source = await loadStandardUnits(request, mainApiUrl, authorization, mainHoldingCode);
    const units = filterMissingUnits(source.units, existing.codes, query, [], includeExisting);

    return NextResponse.json({
      success: true,
      data: units,
      total: units.length,
      source: source.source,
    });
  } catch (error) {
    return NextResponse.json({ success: false, message: productUnitErrorMessage(error) }, { status: 504 });
  }
}

async function listHoldingsWithDisplayNames(request: Request, mainApiUrl: string): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;
  const url = new URL(request.url);
  const activeHoldingCode = holdingCodeFromSearchParams(url.searchParams);

  try {
    const result = await callMainApiJson(request, mainApiUrl, "/list-holding?limit=100", { method: "GET" }, authorization);
    if (!result.ok || isApiFailure(result.payload)) return mainApiError(result, "โหลดบริษัทไม่สำเร็จ");

    const holdings = getArrayFromPayload(result.payload, "data");
    const enriched = await Promise.all(holdings.map((holding) => enrichHoldingDisplayName(request, mainApiUrl, authorization, holding)));

    // Fetch companies and branches sequentially for each holding using session select-holding.
    const enrichedWithBranches = [];
    for (const holding of enriched) {
      if (!isRecord(holding)) {
        enrichedWithBranches.push(holding);
        continue;
      }

      const holdingCode = holdingCodeFromPayload(holding);
      if (!holdingCode) {
        enrichedWithBranches.push({ ...holding, companies: [], branches: [] });
        continue;
      }

      try {
        const selectResult = await callMainApiJson(
          request,
          mainApiUrl,
          "/select-holding",
          {
            method: "POST",
            body: JSON.stringify({ holdingcode: holdingCode }),
          },
          authorization,
        );

        if (selectResult.ok) {
          // Fetch Companies
          const compResult = await callMainApiJson(
            request,
            mainApiUrl,
            "/organization/company",
            { method: "GET" },
            authorization,
          );

          // Fetch Branches
          const branchResult = await callMainApiJson(
            request,
            mainApiUrl,
            "/organization/branch",
            { method: "GET" },
            authorization,
          );

          if (compResult.ok && branchResult.ok) {
            const companies = visibleOrganizationRecords(compResult.payload);
            const visibleCompanyGuids = new Set(companies.map(organizationGuid).filter(Boolean));
            const branches = visibleOrganizationRecords(branchResult.payload)
              .filter((branch) => {
                const companyGuid = stringFromUnknown(branch.companyguid);
                return !companyGuid || visibleCompanyGuids.has(companyGuid);
              });
            enrichedWithBranches.push({
              ...holding,
              companies,
              branches,
            });
            continue;
          }
        }
      } catch (err) {
        console.error(`Error loading organization data for holding ${holdingCode}:`, err);
      }

      enrichedWithBranches.push({
        ...holding,
        companies: [],
        branches: [],
      });
    }

    if (activeHoldingCode) {
      const restoreResult = await callMainApiJson(
        request,
        mainApiUrl,
        "/select-holding",
        {
          method: "POST",
          body: JSON.stringify({ holdingcode: activeHoldingCode }),
        },
        authorization,
      );
      if (!restoreResult.ok || isApiFailure(restoreResult.payload)) {
        return mainApiError(restoreResult, "คืนค่าบริษัทที่เลือกไม่สำเร็จ");
      }
    }

    if (isRecord(result.payload)) {
      return NextResponse.json({ ...result.payload, data: enrichedWithBranches }, { status: result.status });
    }

    return NextResponse.json({ success: true, data: enrichedWithBranches, total: enrichedWithBranches.length }, { status: result.status });
  } catch (error) {
    return NextResponse.json({ success: false, message: workspaceErrorMessage(error, "โหลดบริษัทไม่สำเร็จ") }, { status: 504 });
  }
}

async function updateHoldingDisplayName(
  request: Request,
  mainApiUrl: string,
  payload: Record<string, unknown>,
): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const holdingCode = normalizeHoldingCode(holdingCodeFromPayload(payload));
  if (!holdingCode) return NextResponse.json({ success: false, message: "กรุณากรอกรหัส Holding" }, { status: 400 });
  if (!isValidHoldingCode(holdingCode)) {
    return NextResponse.json(
      { success: false, message: holdingCodeValidationMessageTh },
      { status: 400 },
    );
  }

  const holdingName = holdingNameFromPayload(payload);
  if (!holdingName) return NextResponse.json({ success: false, message: "กรุณากรอกชื่อ Holding" }, { status: 400 });

  try {
    const selected = await callMainApiJson(
      request,
      mainApiUrl,
      "/select-holding",
      {
        method: "POST",
        body: JSON.stringify({ holdingcode: holdingCode }),
      },
      authorization,
    );
    if (!selected.ok || isApiFailure(selected.payload)) return mainApiError(selected, "ตรวจสอบสิทธิ์ owner ไม่สำเร็จ");

    const current = await callMainApiJson(
      request,
      mainApiUrl,
      `/holding/${encodeURIComponent(holdingCode)}`,
      { method: "GET" },
      authorization,
    );
    if (!current.ok || isApiFailure(current.payload)) return mainApiError(current, "โหลดข้อมูล Holding ไม่สำเร็จ");

    const currentHolding = payloadDataRecord(current.payload);
    if (!currentHolding) return NextResponse.json({ success: false, message: "ข้อมูล Holding ไม่ถูกต้อง" }, { status: 502 });

    const updatedHolding: Record<string, unknown> = {
      ...currentHolding,
      holdingcode: holdingCode,
    };
    applyHoldingDisplayName(updatedHolding, holdingName);

    const saved = await callMainApiJson(
      request,
      mainApiUrl,
      `/holding/${encodeURIComponent(holdingCode)}`,
      {
        method: "PUT",
        body: JSON.stringify(updatedHolding),
      },
      authorization,
    );
    if (!saved.ok || isApiFailure(saved.payload)) return mainApiError(saved, "บันทึกชื่อ Holding ไม่สำเร็จ");

    return NextResponse.json(saved.payload, { status: saved.status });
  } catch (error) {
    return NextResponse.json({ success: false, message: workspaceErrorMessage(error, "บันทึกชื่อ Holding ไม่สำเร็จ") }, { status: 504 });
  }
}

async function enrichHoldingDisplayName(request: Request, mainApiUrl: string, authorization: string, holding: unknown): Promise<unknown> {
  if (!isRecord(holding)) return holding;

  const holdingcode = getPayloadString(holding, "holdingcode")?.trim();
  const holdingCode = holdingCodeFromPayload(holding);
  if (!holdingcode && !holdingCode) return holding;

  let holdingInfo: Record<string, unknown> = holding;
  if (!hasHoldingDisplayName(holding) || !hasWorkspaceMetadata(holding)) {
    const result = await callMainApiJson(request, mainApiUrl, `/holding/${encodeURIComponent(holdingcode || holdingCode)}`, { method: "GET" }, authorization);
    const detailed = result.ok && !isApiFailure(result.payload) ? payloadDataRecord(result.payload) : null;
    if (detailed) holdingInfo = { ...holding, ...detailed };
  }

  const names = getArray(holdingInfo, "names")
    .map((name) => normalizeHoldingNameEntry(name))
    .filter((name): name is HoldingNameEntry => Boolean(name));
  const name = firstPayloadString(holdingInfo, ["name1", "companyname", "company_name", "name"]);
  const settings = holdingSettings(holdingInfo);
  const activeLanguages = activeLanguageCodes(settings);
  const currencies = currencyCodes(settings);

  return {
    ...holding,
    ...(realHoldingCodeFromPayload(holdingInfo) ? { holdingcode: realHoldingCodeFromPayload(holdingInfo) } : {}),
    ...(!hasHoldingDisplayName(holding) && name ? { name, name1: name } : {}),
    ...(!hasHoldingDisplayName(holding) && names.length > 0 ? { names } : {}),
    activelanguages: activeLanguages,
    language: getPayloadString(settings, "language") ?? activeLanguages[0],
    languageconfigs: getArray(settings, "languageconfigs"),
    basecurrency: getPayloadString(settings, "basecurrency")?.trim().toUpperCase() ?? "",
    currencies,
    dateformat: getPayloadString(settings, "dateformat")?.trim() ?? "",
    timezone: getPayloadString(settings, "timezone")?.trim() ?? "",
    timezonelabel: getPayloadString(settings, "timezonelabel")?.trim() ?? "",
    timezoneoffset: getPayloadString(settings, "timezoneoffset")?.trim() ?? "",
    yeartype: yearTypeFromSettings(settings),
    usebuddhistcalendar: booleanPayloadValue(settings, "usebuddhistcalendar"),
  };
}

async function createDefaultProductUnits(request: Request, mainApiUrl: string, mainHoldingCode: string, selectedCodes: string[]): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  try {
    const existing = await loadExistingUnitCodes(request, mainApiUrl, authorization);
    if (!existing.ok) return mainApiError(existing.result, "ตรวจสอบหน่วยนับสินค้าไม่สำเร็จ");

    const source = await loadStandardUnits(request, mainApiUrl, authorization, mainHoldingCode);
    const units = filterMissingUnits(
      source.units,
      existing.codes,
      "",
      selectedCodes,
      false,
    );

    if (units.length === 0) {
      return NextResponse.json({ success: true, message: "ไม่มีหน่วยนับใหม่ที่ต้องเพิ่ม", data: { count: 0, source: source.source } });
    }

    const saved = await callMainApiJson(
      request,
      mainApiUrl,
      "/unit/bulk",
      {
        method: "POST",
        body: JSON.stringify(units),
      },
      authorization,
    );

    if (!saved.ok || isApiFailure(saved.payload)) return mainApiError(saved, "เพิ่มหน่วยนับสินค้าไม่สำเร็จ");

    return NextResponse.json({
      success: true,
      message: "เพิ่มหน่วยนับสินค้าเริ่มต้นแล้ว",
      data: { count: units.length, source: source.source },
    });
  } catch (error) {
    return NextResponse.json({ success: false, message: productUnitErrorMessage(error) }, { status: 504 });
  }
}

async function loadExistingUnitCodes(
  request: Request,
  mainApiUrl: string,
  authorization: string,
): Promise<{ ok: true; codes: Set<string> } | { ok: false; result: MainApiResult }> {
  const result = await callMainApiJson(request, mainApiUrl, "/unit/list?offset=0&limit=10000&q=&sort=unitcode:1", { method: "GET" }, authorization);
  if (!result.ok) return { ok: false, result };

  const codes = new Set<string>();
  for (const item of getArrayFromPayload(result.payload, "data")) {
    if (!isRecord(item)) continue;
    const unitcode = firstPayloadString(item, ["unitcode", "unitcode", "code"]);
    if (unitcode) codes.add(normalizeUnitCode(unitcode));
  }
  return { ok: true, codes };
}

async function loadStandardUnits(
  request: Request,
  mainApiUrl: string,
  authorization: string,
  mainHoldingCode: string,
): Promise<{ source: string; units: ProductUnit[] }> {
  if (mainHoldingCode) {
    const mainHoldingUnits = await callMainApiJson(
      request,
      mainApiUrl,
      `/unit?limit=1000&holdingcode=${encodeURIComponent(mainHoldingCode)}`,
      { method: "GET" },
      authorization,
    );
    if (mainHoldingUnits.ok) {
      const units = normalizeUnitList(getArrayFromPayload(mainHoldingUnits.payload, "data"));
      if (units.length > 0) return { source: "main-holding", units };
    }
  }

  return { source: "template", units: await loadTemplateUnits() };
}

function filterMissingUnits(
  units: ProductUnit[],
  existingCodes: Set<string>,
  query: string,
  selectedCodes: string[] = [],
  includeExisting = false,
): ProductUnit[] {
  const needle = query.trim().toLowerCase();
  const selected = new Set(selectedCodes.map(normalizeUnitCode));

  return units.filter((unit) => {
    const normalizedCode = normalizeUnitCode(unit.unitcode);
    if (!includeExisting && existingCodes.has(normalizedCode)) return false;
    if (selected.size > 0 && !selected.has(normalizedCode)) return false;
    if (!needle) return true;
    const haystack = `${unit.unitcode} ${unit.names.map((name) => name.name).join(" ")}`.toLowerCase();
    return haystack.includes(needle);
  });
}

async function callMainApiJson(
  request: Request,
  mainApiUrl: string,
  path: string,
  init: RequestInit,
  authorization: string,
): Promise<MainApiResult> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), API_TIMEOUT_MS);

  try {
    const response = await fetch(`${mainApiUrl}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
        ...(init.headers ?? {}),
      },
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    return { ok: response.ok, status: response.status, payload };
  } finally {
    clearTimeout(timeout);
  }
}

async function loadTemplateUnits(): Promise<ProductUnit[]> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), API_TIMEOUT_MS);

  try {
    const response = await fetch(UNIT_TEMPLATE_URL, {
      headers: { Accept: "application/json" },
      signal: controller.signal,
      cache: "no-store",
    });

    if (!response.ok) throw new Error("โหลด template หน่วยนับสินค้าไม่สำเร็จ");
    const payload = await response.json();
    return normalizeUnitList(Array.isArray(payload) ? payload : []);
  } finally {
    clearTimeout(timeout);
  }
}

function normalizeUnitList(value: unknown[]): ProductUnit[] {
  const units: ProductUnit[] = [];
  const seen = new Set<string>();

  for (const item of value) {
    const unit = normalizeUnit(item);
    if (!unit || seen.has(unit.unitcode)) continue;
    seen.add(unit.unitcode);
    units.push(unit);
  }

  return units;
}

function normalizeUnit(value: unknown): ProductUnit | null {
  if (!isRecord(value)) return null;
  const unitcode = firstPayloadString(value, ["unitcode", "unitcode", "code"]);
  const names = getArray(value, "names")
    .map((name) => normalizeUnitName(name))
    .filter((name): name is ProductUnitName => Boolean(name));

  if (!unitcode || names.length === 0) return null;
  return { unitcode: normalizeUnitCode(unitcode), names };
}

function normalizeUnitName(value: unknown): ProductUnitName | null {
  if (!isRecord(value)) return null;
  const code = getPayloadString(value, "code")?.trim();
  if (!code) return null;
  return {
    code,
    name: getPayloadString(value, "name")?.trim() ?? "",
    isauto: false,
    isdelete: false,
  };
}

function normalizeHoldingNameEntry(value: unknown): HoldingNameEntry | null {
  if (!isRecord(value)) return null;
  const name = getPayloadString(value, "name")?.trim();
  if (!name) return null;
  return {
    code: getPayloadString(value, "code")?.trim(),
    name,
    isauto: value.isauto === true,
    isdelete: value.isdelete === true,
  };
}

function holdingNameFromPayload(payload: Record<string, unknown>): string {
  return normalizeHoldingName(
    firstPayloadString(payload, ["name1", "name", "companyname", "company_name"]) ??
      firstNameFromNames(payload.names),
  );
}

function firstNameFromNames(value: unknown): string {
  if (!Array.isArray(value)) return "";
  for (const item of value) {
    if (!isRecord(item)) continue;
    const name = getPayloadString(item, "name")?.trim();
    if (name) return name;
  }
  return "";
}

function normalizeHoldingName(value: string | undefined): string {
  return (value ?? "")
    .replace(/[\u0000-\u001F\u007F]/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, 120);
}

function applyHoldingDisplayName(payload: Record<string, unknown>, name: string): void {
  payload.name1 = name;
  payload.names = upsertThaiName(payload.names, name);
}

function upsertThaiName(value: unknown, name: string): HoldingNameEntry[] {
  const names = Array.isArray(value)
    ? value.map((item) => normalizeHoldingNameEntry(item)).filter((item): item is HoldingNameEntry => Boolean(item))
    : [];
  const thaiIndex = names.findIndex((item) => item.code?.trim().toLowerCase() === "th");
  if (thaiIndex >= 0) {
    names[thaiIndex] = { ...names[thaiIndex], code: names[thaiIndex].code || "th", name };
    return names;
  }
  return [{ code: "th", name }, ...names];
}

function hasWorkspaceMetadata(holding: Record<string, unknown>): boolean {
  const settings = holdingSettings(holding);
  return getArray(settings, "languageconfigs").length > 0 ||
    payloadStringArray(settings, "currencies").length > 0 ||
    Boolean(
      getPayloadString(settings, "language") ||
      getPayloadString(settings, "basecurrency") ||
      getPayloadString(settings, "dateformat") ||
      getPayloadString(settings, "timezone"),
    );
}

function holdingSettings(holdingInfo: Record<string, unknown>): Record<string, unknown> {
  const settings = isRecord(holdingInfo.settings) ? { ...holdingInfo.settings } : {};
  for (const key of [
    "language",
    "languageconfigs",
    "basecurrency",
    "currencies",
    "currency",
    "currency_codes",
    "currencylist",
    "currency_list",
    "dateformat",
    "timezone",
    "timezonelabel",
    "timezoneoffset",
    "yeartype",
    "usebuddhistcalendar",
  ]) {
    if (settings[key] === undefined && holdingInfo[key] !== undefined) settings[key] = holdingInfo[key];
  }
  return settings;
}

function activeLanguageCodes(settings: Record<string, unknown>): string[] {
  const configs = getArray(settings, "languageconfigs")
    .map((item) => isRecord(item) ? item : null)
    .filter((item): item is Record<string, unknown> => Boolean(item))
    .filter((item) => languageConfigEnabled(item))
    .map((item) => getPayloadString(item, "code")?.trim().toLowerCase() ?? "")
    .filter(Boolean);
  const language = getPayloadString(settings, "language")?.trim().toLowerCase();
  return uniqueStrings(configs.length > 0 ? configs : [language || "th"]);
}

function currencyCodes(settings: Record<string, unknown>): string[] {
  const codes = [
    getPayloadString(settings, "basecurrency"),
    getPayloadString(settings, "currency"),
    ...payloadStringArray(settings, "currencies"),
    ...payloadStringArray(settings, "currency_codes"),
    ...payloadStringArray(settings, "currencylist"),
    ...payloadStringArray(settings, "currency_list"),
  ]
    .map((item) => item?.trim().toUpperCase() ?? "")
    .filter(Boolean);
  return uniqueStrings(codes);
}

function languageConfigEnabled(item: Record<string, unknown>): boolean {
  const value = item.isuse ?? item.isuse ?? item.isUse;
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") return !["false", "0", "no", "n"].includes(value.trim().toLowerCase());
  return true;
}

function yearTypeFromSettings(settings: Record<string, unknown>): string {
  const yearType = getPayloadString(settings, "yeartype")?.trim().toLowerCase();
  if (yearType) return yearType;
  const useBuddhistCalendar = booleanPayloadValue(settings, "usebuddhistcalendar");
  if (useBuddhistCalendar === undefined) return "";
  return useBuddhistCalendar ? "buddhist" : "christian";
}

function booleanPayloadValue(payload: Record<string, unknown>, key: string): boolean | undefined {
  const value = payload[key];
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    if (["true", "1", "yes", "y", "buddhist"].includes(normalized)) return true;
    if (["false", "0", "no", "n", "christian"].includes(normalized)) return false;
  }
  return undefined;
}

function payloadStringArray(payload: Record<string, unknown>, key: string): string[] {
  const value = payload[key];
  if (!Array.isArray(value)) return [];
  return value.map((item) => {
    if (typeof item === "string") return item;
    if (isRecord(item)) return getPayloadString(item, "code") ?? getPayloadString(item, "currency") ?? "";
    return "";
  });
}

function uniqueStrings(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean)));
}

function getPayloadString(payload: Record<string, unknown>, key: string): string | undefined {
  const value = payload[key];
  return typeof value === "string" ? value : undefined;
}

function firstPayloadString(payload: Record<string, unknown>, keys: string[]): string | undefined {
  for (const key of keys) {
    const value = getPayloadString(payload, key)?.trim();
    if (value) return value;
  }
  return undefined;
}

function getPayloadStringArray(payload: Record<string, unknown>, key: string): string[] {
  const value = payload[key];
  if (!Array.isArray(value)) return [];
  return value.map((item) => stringFromUnknown(item)).filter(Boolean);
}

function holdingCodeFromPayload(payload: Record<string, unknown>): string {
  return getPayloadString(payload, "holdingcode")?.trim() || "";
}

function realHoldingCodeFromPayload(payload: Record<string, unknown>): string {
  return getPayloadString(payload, "holdingcode")?.trim() || "";
}

function holdingCodeFromSearchParams(searchParams: URLSearchParams): string {
  return searchParams.get("holdingcode")?.trim() || searchParams.get("activeholdingcode")?.trim() || "";
}

function getArrayFromPayload(payload: unknown, key: string): unknown[] {
  return isRecord(payload) ? getArray(payload, key) : [];
}

function visibleOrganizationRecords(payload: unknown): Record<string, unknown>[] {
  return getArrayFromPayload(payload, "data")
    .filter((item): item is Record<string, unknown> => isVisibleOrganizationRecord(item));
}

function isVisibleOrganizationRecord(value: unknown): boolean {
  if (!isRecord(value)) return false;
  if (value.isactive === false || value.isdelete === true || value.is_delete === true) return false;
  return stringFromUnknown(value.deletedat).length === 0;
}

function organizationGuid(value: Record<string, unknown>): string {
  return stringFromUnknown(value.guidfixed) || stringFromUnknown(value.guid);
}

function isApiFailure(payload: unknown): boolean {
  return isRecord(payload) && payload.success === false;
}

function payloadDataRecord(payload: unknown): Record<string, unknown> | null {
  if (!isRecord(payload)) return null;
  return isRecord(payload.data) ? payload.data : payload;
}

function hasHoldingDisplayName(holding: Record<string, unknown>): boolean {
  const holdingcode = getPayloadString(holding, "holdingcode")?.trim();
  const directName = firstPayloadString(holding, ["name1", "companyname", "company_name", "name"]);
  if (directName && directName !== holdingcode) return true;
  return getArray(holding, "names").some((name) => {
    const displayName = isRecord(name) ? getPayloadString(name, "name")?.trim() : "";
    return Boolean(displayName && displayName !== holdingcode);
  });
}

function mainApiError(result: MainApiResult, fallback: string): NextResponse {
  return NextResponse.json(
    { success: false, message: extractMessage(result.payload) ?? fallback },
    { status: result.status >= 400 ? result.status : 400 },
  );
}

function normalizeUnitCode(value: string): string {
  return value.trim().toUpperCase();
}

function stringFromUnknown(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function productUnitErrorMessage(error: unknown): string {
  return error instanceof Error && error.name === "AbortError"
    ? "Server ไม่ตอบกลับทันเวลา"
    : error instanceof Error && error.message
      ? error.message
      : "ไม่สามารถจัดการหน่วยนับสินค้าเริ่มต้นได้";
}

function workspaceErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.name === "AbortError"
    ? "Server ไม่ตอบกลับทันเวลา"
    : error instanceof Error && error.message
      ? error.message
      : fallback;
}

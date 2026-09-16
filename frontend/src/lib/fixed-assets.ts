import { authFetch, getAuthSession, restoreAuthSession } from "./client-auth-session";
import { type LanguageCode } from "./i18n";

export type FAIdentity = {
  id?: string;
  version?: number;
  isdeleted?: boolean;
};

export type FAName = {
  code: string;
  name: string;
};

export type AssetType = FAIdentity & {
  typecode: string;
  names: FAName[];
  defaultusefullifeyears: number;
  defaultdeprecpercent: string;
  assetaccountcode: string;
  accumdeprecaccountcode: string;
  deprecexpenseaccountcode: string;
  isactive: boolean;
};

export type FixedAsset = FAIdentity & {
  assetcode: string;
  names: FAName[];
  assettypecode: string;
  branchcode: string;
  departmentcode: string;
  locationcode: string;
  purchasedate: string; // YYYY-MM-DD
  startcalcdate: string;
  cost: string;
  scrapvalue: string;
  usefullifeyears: number;
  deprecpercent: string;
  method: string;
  firstyearpercent: string;
  beginaccumdeprec: string;
  assetaccountcode: string;
  accumdeprecaccountcode: string;
  deprecexpenseaccountcode: string;
  status: "active" | "disposed" | "written_off";
  serialnumber: string;
  brand: string;
  model: string;
  suppliercode: string;
  istaxdeductible: boolean;
  notes: string;
};

export type DepreciationScheduleItem = FAIdentity & {
  assetcode: string;
  fiscalyear: string;
  period: number;
  startdate: string;
  stopdate: string;
  days: number;
  perioddeprec: string;
  accumdeprec: string;
  netbookvalue: string;
  isposted: boolean;
  journaldocno: string;
  postedat?: string;
};

export type AssetDisposal = FAIdentity & {
  docno: string;
  assetcode: string;
  disposaldate: string;
  disposaltype: "sale" | "write_off" | "scrap";
  saleprice: string;
  vatamount: string;
  accumdeprecatdisposal: string;
  netbookvalueatdisposal: string;
  gainloss: string;
  settlementaccountcode: string;
  gainlossaccountcode: string;
  journaldocno: string;
  reason: string;
};

export type AssetReportColumn = {
  key: string;
  label: string;
  amount?: boolean;
};

export type AssetScheduleReport = {
  columns: AssetReportColumn[];
  rows: Record<string, string>[];
  totals: Record<string, string>;
  totalrows: number;
  asof: string;
};

export type FACommand = {
  resource: "assets" | "types" | "depreciations" | "disposals" | "processes";
  action: "create" | "update" | "delete" | "calculate" | "post-gl" | "reverse-gl" | "dispose" | "recalculate";
  requestid: string;
  id?: string;
  version?: number;
  date?: string;
  fiscalyear?: string;
  period?: number;
  assetcode?: string;
  reason?: string;
  docno?: string;
  asset?: Partial<FixedAsset>;
  assettype?: Partial<AssetType>;
  disposal?: Partial<AssetDisposal>;
};


// authFetch only refreshes a token that is already on the request, so every
// call has to attach the session itself or the backend answers 401.
async function faFetch(path: string, init?: RequestInit): Promise<Response> {
  const auth = getAuthSession() ?? (await restoreAuthSession());
  if (!auth?.token) throw new Error("กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนใช้งานสินทรัพย์ถาวร");
  return authFetch(path, {
    ...init,
    cache: "no-store",
    headers: {
      Authorization: `Bearer ${auth.token}`,
      "x-bc-backend-url": auth.backendUrl,
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
}
// API Helpers
export async function getFixedAssets(params: { q?: string; type?: string; status?: string; page?: number; limit?: number } = {}) {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  if (params.type) query.set("type", params.type);
  if (params.status) query.set("status", params.status);
  if (params.page) query.set("page", String(params.page));
  if (params.limit) query.set("limit", String(params.limit));

  const res = await faFetch(`/api/fa/assets?${query.toString()}`);
  return res.json();
}

export async function getFixedAsset(id: string) {
  const res = await faFetch(`/api/fa/assets/${encodeURIComponent(id)}`);
  return res.json();
}

export async function getAssetSchedule(id: string) {
  const res = await faFetch(`/api/fa/assets/${encodeURIComponent(id)}/schedule`);
  return res.json();
}

export async function getAssetTypes() {
  const res = await faFetch(`/api/fa/types`);
  return res.json();
}

export async function getFixedAssetScheduleReport(fiscalYear: string, period?: number, typeCode?: string) {
  const query = new URLSearchParams({ fiscalyear: fiscalYear });
  if (period) query.set("period", String(period));
  if (typeCode) query.set("type", typeCode);

  const res = await faFetch(`/api/fa/reports/schedule?${query.toString()}`);
  return res.json();
}

export async function getTaxReconciliationReport(fiscalYear: string) {
  const query = new URLSearchParams({ fiscalyear: fiscalYear });
  const res = await faFetch(`/api/fa/reports/tax-reconciliation?${query.toString()}`);
  return res.json();
}

export async function sendFixedAssetCommand(command: FACommand) {
  const res = await faFetch("/api/fa/command", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(command),
  });
  return res.json();
}

export function assetName(asset: FixedAsset, lang: LanguageCode = "th"): string {
  const n = asset.names?.find((x) => x.code === lang) || asset.names?.find((x) => x.code === "th");
  return n?.name || asset.assetcode;
}

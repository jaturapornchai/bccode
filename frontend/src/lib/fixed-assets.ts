import { authFetch } from "./client-auth-session";
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

// API Helpers
export async function getFixedAssets(params: { q?: string; type?: string; status?: string; page?: number; limit?: number } = {}) {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  if (params.type) query.set("type", params.type);
  if (params.status) query.set("status", params.status);
  if (params.page) query.set("page", String(params.page));
  if (params.limit) query.set("limit", String(params.limit));

  const res = await authFetch(`/api/fa/assets?${query.toString()}`);
  return res.json();
}

export async function getFixedAsset(id: string) {
  const res = await authFetch(`/api/fa/assets/${encodeURIComponent(id)}`);
  return res.json();
}

export async function getAssetSchedule(id: string) {
  const res = await authFetch(`/api/fa/assets/${encodeURIComponent(id)}/schedule`);
  return res.json();
}

export async function getAssetTypes() {
  const res = await authFetch(`/api/fa/types`);
  return res.json();
}

export async function getFixedAssetScheduleReport(fiscalYear: string, period?: number, typeCode?: string) {
  const query = new URLSearchParams({ fiscalyear: fiscalYear });
  if (period) query.set("period", String(period));
  if (typeCode) query.set("type", typeCode);

  const res = await authFetch(`/api/fa/reports/schedule?${query.toString()}`);
  return res.json();
}

export async function getTaxReconciliationReport(fiscalYear: string) {
  const query = new URLSearchParams({ fiscalyear: fiscalYear });
  const res = await authFetch(`/api/fa/reports/tax-reconciliation?${query.toString()}`);
  return res.json();
}

export async function sendFixedAssetCommand(command: FACommand) {
  const res = await authFetch("/api/fa/command", {
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

export const FA_LABELS = {
  assetRegistry: { th: "รายละเอียดสินทรัพย์", en: "Fixed Asset Registry" },
  depreciationCalc: { th: "ประมวลผลสินทรัพย์", en: "Calculate Depreciation" },
  postGL: { th: "โอนค่าเสื่อมเข้าบัญชีแยกประเภท", en: "Post Depreciation to GL" },
  assetDisposal: { th: "จำหน่ายและตัดจำหน่ายสินทรัพย์", en: "Asset Disposal & Write-off" },
  assetSchedule: { th: "ตารางค่าเสื่อมและสินทรัพย์", en: "Fixed Asset Schedule" },
  taxReconciliation: { th: "กระทบยอดภาษี (ภ.ง.ด.50)", en: "Tax Reconciliation (PND 50)" },
  assetTypes: { th: "กำหนดประเภทสินทรัพย์", en: "Asset Types" },
  createAsset: { th: "เพิ่มสินทรัพย์ใหม่", en: "Add New Asset" },
  editAsset: { th: "แก้ไขสินทรัพย์", en: "Edit Asset" },
  deleteAsset: { th: "ลบสินทรัพย์", en: "Delete Asset" },
  cost: { th: "ราคาทุน", en: "Cost" },
  scrapValue: { th: "ราคาซาก", en: "Scrap Value" },
  usefulLife: { th: "อายุการใช้งาน", en: "Useful Life (Years)" },
  deprecRate: { th: "อัตราค่าเสื่อมราคา", en: "Depreciation Rate (%)" },
  purchaseDate: { th: "วันที่ได้มา", en: "Purchase Date" },
  netBookValue: { th: "มูลค่าสุทธิตามบัญชี", en: "Net Book Value" },
  accumDeprec: { th: "ค่าเสื่อมราคาสะสม", en: "Accumulated Depreciation" },
  disposalDate: { th: "วันที่จำหน่าย", en: "Disposal Date" },
  salePrice: { th: "ราคาจำหน่าย", en: "Sale Price" },
  gainLoss: { th: "กำไร (ขาดทุน) จากการจำหน่าย", en: "Gain (Loss) on Disposal" },
} as const;

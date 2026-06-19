export type LocalizedName = {
  code?: string;
  name?: string;
};

export type ShopListItem = {
  holdingcode: string;
  companies?: WorkspaceCompany[];
  name?: string;
  name1?: string;
  names?: LocalizedName[];
  companyname?: string;
  company_name?: string;
  branchcode?: string;
  role?: number;
  is_favorite?: boolean;
  last_accessed_at?: string;
  createdby?: string;
  is_creator?: boolean;
  isaccessdisabled?: boolean;
  activelanguages?: string[];
  language?: string;
  languageconfigs?: Array<LocalizedName & { codetranslator?: string; isuse?: boolean; isdefault?: boolean }>;
  basecurrency?: string;
  currencies?: string[];
  dateformat?: string;
  timezone?: string;
  timezoneoffset?: string;
  timezonelabel?: string;
  yeartype?: string;
  usebuddhistcalendar?: boolean;
};

export type BranchListItem = {
  guidfixed: string;
  companyguid?: string;
  code?: string;
  names?: LocalizedName[];
  companynames?: LocalizedName[];
  basecurrency?: string;
  language?: string;
  timezone?: string;
  timezoneoffset?: string;
  timezonelabel?: string;
  yeartype?: string;
};

export type WorkspaceCompany = {
  guidfixed?: string;
  code?: string;
  names?: LocalizedName[];
  name?: string;
  name1?: string;
  companyname?: string;
  company_name?: string;
};

export type AuthSession = {
  token: string;
  refresh?: string;
  username: string;
  backendUrl: string;
  method?: string;
  holdingcode?: string;
  profile?: {
    email?: string;
    name?: string;
    pictureUrl?: string;
  } | null;
};

export type WorkspaceSession = {
  shop: ShopListItem;
  company?: WorkspaceCompany | null;
  branch?: BranchListItem | null;
  shopInfo?: Record<string, unknown> | null;
};

export const workspaceStorageKeys = {
  auth: "bc_auth",
  holdingCode: "saved_holdingcode",
  workspace: "bc_workspace",
  shopInfo: "bc_shop_info",
  branch: "bc_branch",
};

export const WORKSPACE_CHANGED_EVENT = "bc-workspace-changed";

export function notifyWorkspaceChanged() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(WORKSPACE_CHANGED_EVENT));
  }
}

export function localizedName(names: LocalizedName[] | undefined, fallback = ""): string {
  if (!Array.isArray(names) || names.length === 0) return fallback;
  return names.find((item) => item.code?.toLowerCase() === "th" && item.name)?.name
    ?? names.find((item) => item.name)?.name
    ?? fallback;
}

export function shopDisplayName(shop: ShopListItem): string {
  return localizedName(shop.names, "") ||
    shop.name1?.trim() ||
    shop.companyname?.trim() ||
    shop.company_name?.trim() ||
    shop.name?.trim() ||
    shop.holdingcode;
}

export function branchDisplayName(branch: BranchListItem): string {
  return localizedName(branch.names, branch.code || branch.guidfixed);
}

export function companyBaseName(company: WorkspaceCompany): string {
  return (
    localizedName(company.names, "") ||
    company.name1?.trim() ||
    company.companyname?.trim() ||
    company.company_name?.trim() ||
    company.name?.trim() ||
    company.code?.trim() ||
    company.guidfixed?.trim() ||
    "-"
  );
}

export function companyDisplayName(company: WorkspaceCompany): string {
  const name = companyBaseName(company);
  return company.code?.trim() && name !== company.code.trim()
    ? `[${company.code.trim()}] ${name}`
    : name;
}

function stringRecordValue(record: Record<string, unknown> | null | undefined, key: string): string {
  const value = record?.[key];
  return typeof value === "string" ? value.trim() : "";
}

function localizedNameFromUnknown(value: unknown, fallback = ""): string {
  if (!Array.isArray(value)) return fallback;
  return localizedName(
    value
      .filter((item): item is LocalizedName => Boolean(item) && typeof item === "object")
      .map((item) => ({
        code: typeof item.code === "string" ? item.code : undefined,
        name: typeof item.name === "string" ? item.name : undefined,
      })),
    fallback,
  );
}

export function holdingDisplayName(workspace: WorkspaceSession): string {
  const holdingCode = stringRecordValue(workspace.shopInfo, "holdingcode") || workspace.shop.holdingcode.trim();
  const holdingName =
    localizedNameFromUnknown(workspace.shopInfo?.names) ||
    stringRecordValue(workspace.shopInfo, "name1") ||
    stringRecordValue(workspace.shopInfo, "companyname") ||
    stringRecordValue(workspace.shopInfo, "company_name") ||
    stringRecordValue(workspace.shopInfo, "name") ||
    shopDisplayName(workspace.shop) ||
    holdingCode;

  if (!holdingCode || holdingName === holdingCode) return holdingName || holdingCode || "-";
  return `${holdingName} (${holdingCode})`;
}

export function workspaceCompanyDisplayName(workspace: WorkspaceSession): string {
  if (workspace.company) return companyDisplayName(workspace.company);
  const branchCompanyGuid = workspace.branch?.companyguid?.trim();
  const companyFromHolding = Array.isArray(workspace.shop.companies)
    ? workspace.shop.companies.find((company) => company.guidfixed?.trim() === branchCompanyGuid)
    : undefined;
  if (companyFromHolding) return companyDisplayName(companyFromHolding);
  if (workspace.branch?.companynames?.length) {
    const name = localizedName(workspace.branch.companynames, "");
    if (name) return name;
  }
  return "-";
}

export function workspaceBranchDisplayName(workspace: WorkspaceSession): string {
  if (!workspace.branch) return "-";
  const name = branchDisplayName(workspace.branch);
  return workspace.branch.code?.trim() && name !== workspace.branch.code.trim()
    ? `[${workspace.branch.code.trim()}] ${name}`
    : name;
}

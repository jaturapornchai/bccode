export type LocalizedName = {
  code?: string;
  name?: string;
};

export type ShopListItem = {
  shopid: string;
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
  is_access_disabled?: boolean;
};

export type BranchListItem = {
  guid_fixed: string;
  code?: string;
  names?: LocalizedName[];
  companynames?: LocalizedName[];
  base_currency?: string;
  language?: string;
  timezone?: string;
  timezone_offset?: string;
  timezone_label?: string;
  year_type?: string;
};

export type AuthSession = {
  token: string;
  refresh?: string;
  username: string;
  backendUrl: string;
  method?: string;
  profile?: {
    email?: string;
    name?: string;
    pictureUrl?: string;
  } | null;
};

export type WorkspaceSession = {
  shop: ShopListItem;
  branch?: BranchListItem | null;
  shopInfo?: Record<string, unknown> | null;
};

export const workspaceStorageKeys = {
  auth: "bc_auth",
  workspace: "bc_workspace",
  shopInfo: "bc_shop_info",
  branch: "bc_branch",
};

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
    shop.shopid;
}

export function branchDisplayName(branch: BranchListItem): string {
  return localizedName(branch.names, branch.code || branch.guid_fixed);
}

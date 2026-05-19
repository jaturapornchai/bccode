export type LocalizedName = {
  code?: string;
  name?: string;
};

export type ShopListItem = {
  shopid: string;
  name?: string;
  names?: LocalizedName[];
  branchcode?: string;
  role?: number;
  isfavorite?: boolean;
  lastaccessedat?: string;
  createdby?: string;
  iscreator?: boolean;
  isaccessdisabled?: boolean;
};

export type BranchListItem = {
  guidfixed: string;
  code?: string;
  names?: LocalizedName[];
  companynames?: LocalizedName[];
  base_currency?: string;
  language?: string;
  timezone?: string;
  timezoneoffset?: string;
  timezonelabel?: string;
  yeartype?: string;
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
  return shop.name?.trim() || localizedName(shop.names, shop.shopid);
}

export function branchDisplayName(branch: BranchListItem): string {
  return localizedName(branch.names, branch.code || branch.guidfixed);
}

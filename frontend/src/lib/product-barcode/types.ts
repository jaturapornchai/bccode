/**
 * TypeScript types for Product Barcode feature.
 * Mirrors backend Go struct `ProductBarcodeBase` and friends in
 * `backend/internal/product/productbarcode/models/product_barcode.go`.
 *
 * Source of truth: backend Go. Add fields here only when backend has them.
 */

import type { LocalizedName } from "@/lib/workspace-models";

/** Language entry — matches Go `models.NameX` (`{code,name,description}`). */
export type NameX = LocalizedName;

/** Sort key — matches Go `models.XSort`. */
export interface XSort {
  code: string;
  xorder: number;
}

/** Identity used by master pickers (code+guid+names). */
export interface MasterIdentity {
  guid: string;
  code: string;
  names: NameX[];
}

/** Price entry — matches Go `ProductPrice`. */
export interface ProductPrice {
  key_number: number;
  price: number;
}

/** Image entry — matches Go `ProductImage`. */
export interface ProductImage {
  xorder: number;
  uri: string;
}

export interface MarketplaceMediaAsset {
  kind: string;
  uri: string;
  external_id: string;
  external_url: string;
  option_code: string;
  option_value: string;
  sort_order: number;
  alt_text: string;
  use_case: string;
  mime_type: string;
  width: number;
  height: number;
  last_imported_at: string;
}

export interface MarketplaceAttributeValue {
  value_id: string;
  value_code: string;
  value_text: string;
  display_text: string;
  unit_code: string;
  sort_order: number;
  is_custom_value: boolean;
}

export interface MarketplaceAttribute {
  attribute_id: string;
  attribute_code: string;
  attribute_name: string;
  input_type: string;
  scope: string;
  is_required: boolean;
  is_sale_prop: boolean;
  is_custom: boolean;
  values: MarketplaceAttributeValue[];
}

export interface MarketplaceSpecificationGroup {
  group_code: string;
  group_name: string;
  attributes: MarketplaceAttribute[];
}

export interface MarketplacePayloadExample {
  direction: string;
  use_case: string;
  payload: unknown;
}

/** Fixed cost entry — matches Go `FixedCost`. */
export interface FixedCost {
  effectdate: string; // ISO date string
  amount: number;
}

/** Restaurant flags — matches Go `ProductRestaurant`. */
export interface ProductRestaurant {
  isforrestaurant: boolean;
  isfortakeaway: boolean;
  isfordelivery: boolean;
  isforcustomer: boolean;
  isforcustomerpreorder: boolean;
}

/** Order type entry. */
export interface ProductOrderType {
  guidfixed: string;
  code: string;
  names: NameX[];
  chargeprice?: number;
}

/** Supported marketplace platforms (matches backend `platform` values). */
export const MARKETPLACE_PLATFORMS = ["shopee", "lazada", "aliexpress", "tiktok"] as const;
export type MarketplacePlatform = (typeof MARKETPLACE_PLATFORMS)[number];

/** Listing status options shared across platforms. */
export const MARKETPLACE_STATUS = ["", "LIVE", "UNLIST", "REVIEWING", "REJECTED", "DELETED"] as const;
export type MarketplaceStatus = (typeof MARKETPLACE_STATUS)[number];

/**
 * One marketplace listing mapping for a product/barcode — matches Go
 * `MarketplaceProductMap`. Unified shape across Shopee/Lazada/AliExpress/TikTok; unused
 * fields per platform stay blank.
 */
export interface MarketplaceProductMap {
  platform: string;
  account_id: string;
  holding_code: string;
  market_item_id: string;
  market_model_id: string;
  item_url: string;
  seller_sku: string;
  shop_sku: string;
  gtin: string;
  category_id: string;
  category_name: string;
  brand_id: string;
  media_assets: MarketplaceMediaAsset[];
  specification_groups: MarketplaceSpecificationGroup[];
  raw_attributes: MarketplaceAttribute[];
  payload_examples: MarketplacePayloadExample[];
  currency: string;
  custom_price: number;
  platform_price: number;
  platform_stock: number;
  sync_stock: boolean;
  sync_price: boolean;
  status: string;
  reject_reason: string;
  days_to_ship: number;
  is_pre_order: boolean;
  sync_enabled: boolean;
  sync_status: string;
  last_sync_at: string;
  last_sync_error: string;
}

/** Sub/variation barcode marketplace mapping — matches Go `MarketplaceSKUMap`. */
export interface MarketplaceSKUMap {
  platform: string;
  account_id: string;
  holding_code: string;
  market_item_id: string;
  market_model_id: string;
  seller_sku: string;
  shop_sku: string;
  gtin: string;
  media_assets?: MarketplaceMediaAsset[];
  raw_attributes?: MarketplaceAttribute[];
  currency: string;
  sync_stock: boolean;
  sync_price: boolean;
  custom_price: number;
  platform_price: number;
  platform_stock: number;
  status: string;
  sync_enabled: boolean;
  last_sync_at: string;
}

/** Build an empty marketplace listing map for a given platform. */
export function emptyMarketplaceProductMap(platform: MarketplacePlatform): MarketplaceProductMap {
  return {
    platform,
    account_id: "",
    holding_code: "",
    market_item_id: "",
    market_model_id: "",
    item_url: "",
    seller_sku: "",
    shop_sku: "",
    gtin: "",
    category_id: "",
    category_name: "",
    brand_id: "",
    media_assets: [],
    specification_groups: [],
    raw_attributes: [],
    payload_examples: [],
    currency: "THB",
    custom_price: 0,
    platform_price: 0,
    platform_stock: 0,
    sync_stock: false,
    sync_price: false,
    status: "",
    reject_reason: "",
    days_to_ship: 0,
    is_pre_order: false,
    sync_enabled: false,
    sync_status: "",
    last_sync_at: "",
    last_sync_error: "",
  };
}

/** Sub/ref barcode — matches Go `RefProductBarcode`. */
export interface RefProductBarcode {
  guid_fixed: string;
  names: NameX[];
  item_unit_code: string;
  itemunitnames: NameX[];
  barcode: string;
  condition: boolean;
  dividevalue: number;
  standvalue: number;
  qty: number;

  // Marketplace & SKU Logistics
  seller_sku?: string;
  sku_package_weight?: number;
  sku_package_length?: number;
  sku_package_width?: number;
  sku_package_height?: number;
  marketplace_sku_mappings?: MarketplaceSKUMap[];
}

/** BOM entry — matches Go `BOMProductBarcode`. */
export interface BOMProductBarcode {
  barcodeguidfixed: string;
  names: NameX[];
  item_unit_code: string;
  itemunitnames: NameX[];
  barcode: string;
  qty?: number;
  standvalue?: number;
  dividevalue?: number;
}

/** Dimension — matches Go `ProductDimension`. */
export interface ProductDimension {
  guidfixed: string;
  names: NameX[];
  isdisabled: boolean;
  item: {
    guidfixed: string;
    names: NameX[];
    isdisabled: boolean;
  };
}

export interface ProductStockDimension {
  dimension_guid?: string;
  dimension_name?: string;
  item_guid?: string;
  item_name?: string;
}

/** Business type — matches Go `ProductBarcodeBusinessType`. */
export interface ProductBarcodeBusinessType {
  guidfixed: string;
  code: string;
  names: NameX[];
  isignore: boolean;
}

/** Ignore branch — matches Go `ProductBarcodeBranch`. */
export interface ProductBarcodeBranch {
  guidfixed: string;
  code: string;
  names: NameX[];
  isignore: boolean;
}

/** Time-for-sale entry — matches Go `ProductTimeForSale`. */
export interface ProductTimeForSale {
  daysofweek: number[]; // 0=Sun, 6=Sat (per Go int8)
  fromdate: string;
  todate: string;
  fromtime: string; // HH:mm
  totime: string;
}

/** Product option choice. */
export interface ProductChoice {
  guid: string;
  names: NameX[];
  imageuri?: string;
  refbarcode?: string;
  refbarcodenames?: NameX[];
  refproductcode?: string;
  refunitcode?: string;
  isstock: boolean;
  isdefault: boolean;
  qty?: number;
  price?: string;
}

/** Product option — multiselect/single choice group. */
export interface ProductOption {
  guid: string;
  names: NameX[];
  choicetype: 0 | 1; // 0=multi, 1=single
  minselect?: number;
  maxselect?: number;
  choices: ProductChoice[];
}

/** Enum: item type. */
export const ITEM_TYPE = {
  STOCK: 0,
  SERVICE: 1,
  SET: 2,
  NOT_STOCK: 3,
} as const;
export type ItemType = (typeof ITEM_TYPE)[keyof typeof ITEM_TYPE];

/** Enum: material type. */
export const MATERIAL_TYPE = {
  GENERAL: 0,
  MATERIAL: 1,
  SEMI_FINISHED: 2,
} as const;
export type MaterialType = (typeof MATERIAL_TYPE)[keyof typeof MATERIAL_TYPE];

/** Enum: food type. */
export const FOOD_TYPE = {
  FOOD: 0,
  DRINK: 1,
  ALCOHOL: 2,
  OTHER: 3,
} as const;
export type FoodType = (typeof FOOD_TYPE)[keyof typeof FOOD_TYPE];

/** Enum: legacy product VAT flag (Flutter default: 0). */
export const VAT_TYPE = {
  TAXABLE: 0,
  EXEMPT: 1,
} as const;
export type VatType = (typeof VAT_TYPE)[keyof typeof VAT_TYPE];

/** Enum: product type (POS classifier). */
export type ProductType = number;

/**
 * Full ProductBarcode — matches Go `ProductBarcode` + `ProductBarcodeBase`.
 * Field names use snake_case where backend JSON does, otherwise lowercase.
 */
export interface ProductBarcode {
  // Identity
  guidfixed: string;
  holding_code?: string;
  itemcode: string;
  barcode: string;
  names: NameX[];
  xsorts: XSort[];
  item_guid: string;

  // Unit
  itemunitguid: string;
  item_unit_code: string;
  itemunitnames: NameX[];
  itemunitsize: number;

  // Classification — Group hierarchy
  groupguid: string;
  group_code: string;
  group_names: NameX[];
  groupsubonecode: string;
  groupsubonenames: NameX[];
  groupsuboneguid: string;
  groupsubtwoguid: string;
  groupsubtwocode: string;
  groupsubtwonames: NameX[];

  // Other masters
  brandguid: string;
  brand_code: string;
  brandnames: NameX[];
  designguid: string;
  designcode: string;
  designnames: NameX[];
  modelguid: string;
  modelcode: string;
  modelnames: NameX[];
  patternguid: string;
  patterncode: string;
  patternnames: NameX[];
  gradeguid: string;
  gradecode: string;
  gradenames: NameX[];
  category_guid: string;
  categorycode: string;
  category_names: NameX[];
  classguid: string;
  classcode: string;
  classnames: NameX[];
  manufacturerguid: string;
  manufacturercode: string;
  manufacturernames: NameX[];
  manufacturers?: ProductManufacturer[];
  suppliers?: ProductSupplier[];

  // Stock
  orderpoint: number;
  minpoint: number;
  maxpoint: number;
  qty: number;
  stockbarcode: string;
  refguidfixed: string;
  refdividevalue: number;
  refstandvalue: number;
  refunitnames: NameX[];

  // Conversion
  condition: boolean;
  dividevalue: number;
  standvalue: number;
  isusesubbarcodes: boolean;
  is_main_barcode: boolean;

  // Pricing
  prices: ProductPrice[];
  fixedcost: FixedCost[];
  discount: string;
  maxdiscount: string;
  isdividend: boolean;
  isdiscountpointofpurchase: boolean;

  // Image & color
  imageuri: string;
  images: ProductImage[];
  useimageorcolor: boolean;
  colorselect: string;
  colorselecthex: string;

  // Type flags
  item_type: ItemType;
  materialtype: MaterialType;
  tax_type: number;
  vat_type: VatType;
  vatcal: number;
  producttype: ProductType;
  foodtype: FoodType;
  issumpoint: boolean;
  isalacarte: boolean;
  isstockforrestaurant: boolean;
  issplitunitprint: boolean;
  isonlystaff: boolean;

  // Restaurant
  restaurant: ProductRestaurant;
  ordertypes: ProductOrderType[];
  options: ProductOption[];

  // Sub & BOM
  refbarcodes: RefProductBarcode[];
  bom: BOMProductBarcode[];

  // Multi-tenant filters
  businesstypes: ProductBarcodeBusinessType[];
  ignorebranches: ProductBarcodeBranch[];

  // Time-based sales
  timeforsales: ProductTimeForSale[];

  // Dimensions
  dimensions: ProductDimension[];

  // Alert & description
  isalert: boolean;
  alertdescription: string;
  description: string;

  // Marketplace & logistics (matches Go ProductBarcodeBase)
  package_weight: number;
  package_length: number;
  package_width: number;
  package_height: number;
  marketplace_products: MarketplaceProductMap[];
}

/** Request body for create. */
export type ProductBarcodeCreateRequest = Omit<ProductBarcode, "guidfixed"> & {
  guidfixed?: string;
};

/** Request body for update. */
export type ProductBarcodeUpdateRequest = ProductBarcode;

/** Row shape returned by `POST /goapi/api/product/barcode/list` (PG list). */
export interface ProductBarcodeListRow {
  guidfixed: string;
  holding_code?: string;
  barcode: string;
  names: NameX[];
  itemunitcode: string;
  itemunitnames: NameX[];
  itemcode: string;
  groupcode: string;
  groupnames: NameX[];
  group_names?: NameX[];
  brand_code?: string;
  brandnames?: NameX[];
  categorycode?: string;
  category_names?: NameX[];
  prices: ProductPrice[];
  price?: number;
  imageuri: string;
  balance_qty?: number;
  reserved_qty?: number;
  available_qty?: number;
  stock_dimension_key?: string;
  stock_dimensions?: ProductStockDimension[];
  balanceamount?: number;
  balance_amount?: number;
  averagecost?: number;
  mainbarcoderef?: string;
  standvalue?: number;
  dividevalue?: number;
  bom?: BOMProductBarcode[];
}

/** Sort field whitelist accepted by PG list endpoint. */
export type BarcodeListSortField =
  | "barcode"
  | "name0"
  | "itemcode"
  | "groupnames"
  | "price1"
  | "unitname"
  | "relevance";

export type SortOrder = "asc" | "desc";

/** Filters supported by the PG list endpoint. */
export interface ProductBarcodeListFilters {
  keyword?: string;
  groupcode?: string;
  brandcode?: string;
  categorycode?: string;
  classcode?: string;
  designcode?: string;
  gradecode?: string;
  modelcode?: string;
  patterncode?: string;
  price_min?: number | null;
  price_max?: number | null;
  limit?: number;
  offset?: number;
  sort_field?: BarcodeListSortField;
  sort_order?: SortOrder;
}

/** Request body for list endpoint. */
export interface ProductBarcodeListRequest extends ProductBarcodeListFilters {
  holding_code: string;
}

/** Response envelope from list endpoint. */
export interface ProductBarcodeListResponse {
  success: boolean;
  message?: string;
  data?: ProductBarcodeListRow[];
  total?: number;
}

/** Build an empty barcode with sensible defaults (for "Add new"). */
export function emptyProductBarcode(): ProductBarcode {
  return {
    guidfixed: "",
    holding_code: "",
    itemcode: "",
    barcode: "",
    names: [],
    xsorts: [],
    item_guid: "",

    itemunitguid: "",
    item_unit_code: "",
    itemunitnames: [],
    itemunitsize: 0,

    groupguid: "",
    group_code: "",
    group_names: [],
    groupsubonecode: "",
    groupsubonenames: [],
    groupsuboneguid: "",
    groupsubtwoguid: "",
    groupsubtwocode: "",
    groupsubtwonames: [],

    brandguid: "",
    brand_code: "",
    brandnames: [],
    designguid: "",
    designcode: "",
    designnames: [],
    modelguid: "",
    modelcode: "",
    modelnames: [],
    patternguid: "",
    patterncode: "",
    patternnames: [],
    gradeguid: "",
    gradecode: "",
    gradenames: [],
    category_guid: "",
    categorycode: "",
    category_names: [],
    classguid: "",
    classcode: "",
    classnames: [],
    manufacturerguid: "",
    manufacturercode: "",
    manufacturernames: [],

    orderpoint: 0,
    minpoint: 0,
    maxpoint: 0,
    qty: 0,
    stockbarcode: "",
    refguidfixed: "",
    refdividevalue: 0,
    refstandvalue: 0,
    refunitnames: [],

    condition: false,
    dividevalue: 1,
    standvalue: 1,
    isusesubbarcodes: false,
    is_main_barcode: true,

    prices: [{ key_number: 1, price: 0 }],
    fixedcost: [],
    discount: "",
    maxdiscount: "",
    isdividend: false,
    isdiscountpointofpurchase: false,

    imageuri: "",
    images: [],
    useimageorcolor: true,
    colorselect: "",
    colorselecthex: "",

    item_type: ITEM_TYPE.STOCK,
    materialtype: MATERIAL_TYPE.GENERAL,
    tax_type: 0,
    vat_type: VAT_TYPE.TAXABLE,
    vatcal: 0,
    producttype: 0,
    foodtype: FOOD_TYPE.FOOD,
    issumpoint: false,
    isalacarte: false,
    isstockforrestaurant: false,
    issplitunitprint: false,
    isonlystaff: false,

    restaurant: {
      isforrestaurant: false,
      isfortakeaway: false,
      isfordelivery: false,
      isforcustomer: false,
      isforcustomerpreorder: false,
    },
    ordertypes: [],
    options: [],

    refbarcodes: [],
    bom: [],

    businesstypes: [],
    ignorebranches: [],

    timeforsales: [],

    dimensions: [],

    isalert: false,
    alertdescription: "",
    description: "",

    package_weight: 0,
    package_length: 0,
    package_width: 0,
    package_height: 0,
    marketplace_products: [],
  };
}

export interface ProductManufacturer {
  guid_fixed: string;
  code: string;
  names: NameX[];
}

export interface ProductSupplier {
  guid_fixed: string;
  code: string;
  names: NameX[];
}

export interface Product {
  guidfixed: string;
  holding_code: string;
  code: string;
  names: NameX[];
  group_code: string;
  group_names: NameX[];
  manufacturerguid?: string;
  manufacturercode?: string;
  manufacturernames?: NameX[];
  dimensions?: ProductDimension[];
  vat_type?: number;
  item_type?: number;
  unitguid?: string;

  // Classifications
  groupsuboneguid?: string;
  groupsubonecode?: string;
  groupsubonenames?: NameX[];
  groupsubtwoguid?: string;
  groupsubtwocode?: string;
  groupsubtwonames?: NameX[];
  brandguid?: string;
  brand_code?: string;
  brandnames?: NameX[];
  designguid?: string;
  designcode?: string;
  designnames?: NameX[];
  modelguid?: string;
  modelcode?: string;
  modelnames?: NameX[];
  patternguid?: string;
  patterncode?: string;
  patternnames?: NameX[];
  gradeguid?: string;
  gradecode?: string;
  gradenames?: NameX[];
  category_guid?: string;
  categorycode?: string;
  category_names?: NameX[];
  classguid?: string;
  classcode?: string;
  classnames?: NameX[];
  materialtype?: number;
  tax_type?: number;
  manufacturers?: ProductManufacturer[];
  suppliers?: ProductSupplier[];

  // Core Product Properties Moved from ProductBarcode
  imageuri?: string;
  images?: ProductImage[];
  useimageorcolor?: boolean;
  colorselect?: string;
  colorselecthex?: string;
  issumpoint?: boolean;
  isalacarte?: boolean;
  issplitunitprint?: boolean;
  isonlystaff?: boolean;
  foodtype?: number;
  isstockforrestaurant?: boolean;
  restaurant?: ProductRestaurant;
  ordertypes?: ProductOrderType[];
  options?: ProductOption[];
  isalert?: boolean;
  alertdescription?: string;
  description?: string;
  timeforsales?: ProductTimeForSale[];
  businesstypes?: ProductBarcodeBusinessType[];
  ignorebranches?: ProductBarcodeBranch[];

  // Units and BOM properties
  condition?: boolean;
  dividevalue?: number;
  standvalue?: number;
  isusesubbarcodes?: boolean;
  refbarcodes?: RefProductBarcode[];
  bom?: BOMProductBarcode[];
  barcodes?: RefProductBarcode[];
  unitcode?: string;
  unitnames?: NameX[];
  item_unit_code?: string;
  itemunitnames?: NameX[];

  // Stock properties
  orderpoint?: number;
  minpoint?: number;
  maxpoint?: number;
  qty?: number;
  stockbarcode?: string;

  // Marketplace & Logistics
  package_weight?: number;
  package_length?: number;
  package_width?: number;
  package_height?: number;
  marketplace_products?: MarketplaceProductMap[];
}

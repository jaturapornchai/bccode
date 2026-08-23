/**
 * TypeScript types for Product Barcode feature.
 * Mirrors backend Go struct `ProductBarcodeBase` and friends in
 * the Go productbarcode model.
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
  keynumber: number;
  price: number;
}

/** Image entry — matches Go `ProductImage`. */
export interface ProductImage {
  xorder: number;
  uri: string;
}

/** Video entry — matches Go `ProductVideo`. */
export interface ProductVideo {
  xorder: number;
  uri: string;
  posteruri: string;
}

export const PRODUCT_VIDEO_MAX_MB = 500;
export const PRODUCT_VIDEO_MAX_BYTES = PRODUCT_VIDEO_MAX_MB * 1024 * 1024;
export const PRODUCT_VIDEO_REQUEST_MAX_BYTES =
  PRODUCT_VIDEO_MAX_BYTES + 1024 * 1024;
export const PRODUCT_VIDEO_UPLOAD_TIMEOUT_MS = 30 * 60 * 1000;

export interface MarketplaceMediaAsset {
  kind: string;
  uri: string;
  externalid: string;
  externalurl: string;
  optioncode: string;
  optionvalue: string;
  sortorder: number;
  alttext: string;
  usecase: string;
  mimetype: string;
  width: number;
  height: number;
  lastimportedat: string;
}

export interface MarketplaceAttributeValue {
  valueid: string;
  valuecode: string;
  valuetext: string;
  displaytext: string;
  unitcode: string;
  sortorder: number;
  iscustomvalue: boolean;
}

export interface MarketplaceAttribute {
  attributeid: string;
  attributecode: string;
  attributename: string;
  inputtype: string;
  scope: string;
  isrequired: boolean;
  issaleprop: boolean;
  iscustom: boolean;
  values: MarketplaceAttributeValue[];
}

export interface MarketplaceSpecificationGroup {
  groupcode: string;
  groupname: string;
  attributes: MarketplaceAttribute[];
}

export interface MarketplacePayloadExample {
  direction: string;
  usecase: string;
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
export const MARKETPLACE_PLATFORMS = [
  "shopee",
  "lazada",
  "aliexpress",
  "tiktok",
] as const;
export type MarketplacePlatform = (typeof MARKETPLACE_PLATFORMS)[number];

/** Listing status options shared across platforms. */
export const MARKETPLACE_STATUS = [
  "",
  "LIVE",
  "UNLIST",
  "REVIEWING",
  "REJECTED",
  "DELETED",
] as const;
export type MarketplaceStatus = (typeof MARKETPLACE_STATUS)[number];

/**
 * One marketplace listing mapping for a product/barcode — matches Go
 * `MarketplaceProductMap`. Unified shape across Shopee/Lazada/AliExpress/TikTok; unused
 * fields per platform stay blank.
 */
export interface MarketplaceProductMap {
  platform: string;
  accountid: string;
  holdingcode: string;
  marketitemid: string;
  marketmodelid: string;
  itemurl: string;
  sellersku: string;
  shopsku: string;
  gtin: string;
  categoryid: string;
  categoryname: string;
  brandid: string;
  mediaassets: MarketplaceMediaAsset[];
  specificationgroups: MarketplaceSpecificationGroup[];
  rawattributes: MarketplaceAttribute[];
  payloadexamples: MarketplacePayloadExample[];
  currency: string;
  customprice: number;
  platformprice: number;
  platformstock: number;
  syncstock: boolean;
  syncprice: boolean;
  status: string;
  rejectreason: string;
  daystoship: number;
  ispreorder: boolean;
  syncenabled: boolean;
  syncstatus: string;
  lastsyncat: string;
  lastsyncerror: string;
}

/** Sub/variation barcode marketplace mapping — matches Go `MarketplaceSKUMap`. */
export interface MarketplaceDimensionStock {
  dimensionkey: string;
  dimensionname: string;
  marketdimensionid: string;
  availableqty: number;
  reservedqty: number;
  inboundqty: number;
  oversellbufferqty: number;
  lastplatformstock: number;
  lastsyncedat: string;
  lastsyncstatus: string;
  lastsyncerror: string;
}

export interface MarketplaceSKUMap {
  platform: string;
  accountid: string;
  holdingcode: string;
  marketitemid: string;
  marketmodelid: string;
  sellersku: string;
  shopsku: string;
  gtin: string;
  mediaassets?: MarketplaceMediaAsset[];
  rawattributes?: MarketplaceAttribute[];
  currency: string;
  syncstock: boolean;
  syncprice: boolean;
  customprice: number;
  platformprice: number;
  platformstock: number;
  marketplacedimensionstocks: MarketplaceDimensionStock[];
  status: string;
  syncenabled: boolean;
  lastsyncat: string;
}

export function emptyMarketplaceSKUMap(
  platform: MarketplacePlatform,
  holdingCode = "",
  marketItemId = "",
): MarketplaceSKUMap {
  return {
    platform,
    accountid: "",
    holdingcode: holdingCode,
    marketitemid: marketItemId,
    marketmodelid: "",
    sellersku: "",
    shopsku: "",
    gtin: "",
    currency: "THB",
    syncstock: true,
    syncprice: true,
    customprice: 0,
    platformprice: 0,
    platformstock: 0,
    marketplacedimensionstocks: [],
    status: "",
    syncenabled: true,
    lastsyncat: "",
  };
}

/** Build an empty marketplace listing map for a given platform. */
export function emptyMarketplaceProductMap(
  platform: MarketplacePlatform,
): MarketplaceProductMap {
  return {
    platform,
    accountid: "",
    holdingcode: "",
    marketitemid: "",
    marketmodelid: "",
    itemurl: "",
    sellersku: "",
    shopsku: "",
    gtin: "",
    categoryid: "",
    categoryname: "",
    brandid: "",
    mediaassets: [],
    specificationgroups: [],
    rawattributes: [],
    payloadexamples: [],
    currency: "THB",
    customprice: 0,
    platformprice: 0,
    platformstock: 0,
    syncstock: false,
    syncprice: false,
    status: "",
    rejectreason: "",
    daystoship: 0,
    ispreorder: false,
    syncenabled: false,
    syncstatus: "",
    lastsyncat: "",
    lastsyncerror: "",
  };
}

/** Sub/ref barcode — matches Go `RefProductBarcode`. */
export interface RefProductBarcode {
  guidfixed: string;
  itemcode: string;
  names: NameX[];
  itemunitcode: string;
  itemunitnames: NameX[];
  barcode: string;
  condition: boolean;
  dividevalue: number;
  standvalue: number;
  qty: number;
  imageuri?: string;
  images?: ProductImage[];
  videos?: ProductVideo[];
  description?: string;

  // Marketplace & SKU Logistics
  sellersku?: string;
  skupackageweight?: number;
  skupackagelength?: number;
  skupackagewidth?: number;
  skupackageheight?: number;
  marketplaceskumappings?: MarketplaceSKUMap[];
}

/** Additional Product unit; Barcode matching is derived by `itemunitcode`. */
export interface ProductUnitConversion {
  unitcode: string;
  unitnames: NameX[];
  dividevalue: number;
  standvalue: number;
}

/** BOM entry — matches Go `BOMProductBarcode`. */
export interface BOMProductBarcode {
  barcodeguidfixed: string;
  itemcode: string;
  names: NameX[];
  itemunitcode: string;
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
  dimensionguid?: string;
  dimensionname?: string;
  itemguid?: string;
  itemname?: string;
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

/** Product type master reference — matches Go struct `models.ProductType` ({guidfixed, code, names}).
 * NOTE: this is a struct, NOT the numeric POS classifier (that is `itemtype: ItemType`). */
export interface ProductType {
  guidfixed?: string;
  code: string;
  names: NameX[];
}

/**
 * Full ProductBarcode — matches Go `ProductBarcode` + `ProductBarcodeBase`.
 * Field names use snake_case where backend JSON does, otherwise lowercase.
 */
export interface ProductBarcode {
  // Identity
  guidfixed: string;
  holdingcode: string;
  businesscode: string;
  itemcode: string;
  barcode: string;
  names: NameX[];
  xsorts: XSort[];
  itemguid: string;

  // Unit
  itemunitguid: string;
  itemunitcode: string;
  itemunitnames: NameX[];
  itemunitsize: number;

  // Classification — Group hierarchy
  groupguid: string;
  groupcode: string;
  groupnames: NameX[];
  subgroupguid: string;
  subgroupcode: string;
  subgroupnames: NameX[];

  // Other masters
  brandguid: string;
  brandcode: string;
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
  categoryguid: string;
  categorycode: string;
  categorynames: NameX[];
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
  videos: ProductVideo[];
  useimageorcolor: boolean;
  colorselect: string;
  colorselecthex: string;

  // Type flags
  itemtype: ItemType;
  materialtype: MaterialType;
  taxtype: number;
  vattype: VatType;
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
  packageweight: number;
  packagelength: number;
  packagewidth: number;
  packageheight: number;
  marketplaceproducts: MarketplaceProductMap[];
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
  holdingcode?: string;
  barcode: string;
  names: NameX[];
  itemunitcode: string;
  itemunitnames: NameX[];
  itemcode: string;
  groupcode: string;
  groupnames: NameX[];
  brandcode?: string;
  brandnames?: NameX[];
  categorycode?: string;
  categorynames?: NameX[];
  prices: ProductPrice[];
  price?: number;
  imageuri: string;
  balanceqty?: number;
  reservedqty?: number;
  availableqty?: number;
  stockdimensionkey?: string;
  stockdimensions?: ProductStockDimension[];
  balanceamount?: number;
  averagecost?: number;
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
  pricemin?: number | null;
  pricemax?: number | null;
  limit?: number;
  offset?: number;
  sortfield?: BarcodeListSortField;
  sortorder?: SortOrder;
}

/** Request body for list endpoint. */
export interface ProductBarcodeListRequest extends ProductBarcodeListFilters {
  holdingcode: string;
  businesscode: string;
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
    holdingcode: "",
    businesscode: "",
    itemcode: "",
    barcode: "",
    names: [],
    xsorts: [],
    itemguid: "",

    itemunitguid: "",
    itemunitcode: "",
    itemunitnames: [],
    itemunitsize: 0,

    groupguid: "",
    groupcode: "",
    groupnames: [],
    subgroupguid: "",
    subgroupcode: "",
    subgroupnames: [],

    brandguid: "",
    brandcode: "",
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
    categoryguid: "",
    categorycode: "",
    categorynames: [],
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

    prices: [{ keynumber: 1, price: 0 }],
    fixedcost: [],
    discount: "",
    maxdiscount: "",
    isdividend: false,
    isdiscountpointofpurchase: false,

    imageuri: "",
    images: [],
    videos: [],
    useimageorcolor: true,
    colorselect: "",
    colorselecthex: "",

    itemtype: ITEM_TYPE.STOCK,
    materialtype: MATERIAL_TYPE.GENERAL,
    taxtype: 0,
    vattype: VAT_TYPE.TAXABLE,
    vatcal: 0,
    producttype: { guidfixed: "", code: "", names: [] },
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

    packageweight: 0,
    packagelength: 0,
    packagewidth: 0,
    packageheight: 0,
    marketplaceproducts: [],
  };
}

export interface ProductManufacturer {
  guidfixed: string;
  code: string;
  names: NameX[];
}

export interface ProductSupplier {
  guidfixed: string;
  code: string;
  names: NameX[];
}

export interface Product {
  guidfixed: string;
  holdingcode: string;
  code: string;
  names: NameX[];
  groupcode: string;
  groupnames: NameX[];
  manufacturerguid?: string;
  manufacturercode?: string;
  manufacturernames?: NameX[];
  dimensions?: ProductDimension[];
  vattype?: number;
  itemtype?: number;
  unitguid?: string;

  // Classifications
  subgroupguid?: string;
  subgroupcode?: string;
  subgroupnames?: NameX[];
  brandguid?: string;
  brandcode?: string;
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
  categoryguid?: string;
  categorycode?: string;
  categorynames?: NameX[];
  classguid?: string;
  classcode?: string;
  classnames?: NameX[];
  materialtype?: number;
  taxtype?: number;
  manufacturers?: ProductManufacturer[];
  suppliers?: ProductSupplier[];

  // Core Product Properties Moved from ProductBarcode
  imageuri?: string;
  images?: ProductImage[];
  videos?: ProductVideo[];
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
  unitconversions?: ProductUnitConversion[];
  isusesubbarcodes?: boolean;
  refbarcodes?: RefProductBarcode[];
  bom?: BOMProductBarcode[];
  barcodes?: RefProductBarcode[];
  unitcode?: string;
  unitnames?: NameX[];
  itemunitcode?: string;
  itemunitnames?: NameX[];

  // Stock properties
  orderpoint?: number;
  minpoint?: number;
  maxpoint?: number;
  qty?: number;
  stockbarcode?: string;
  _unit_count?: number;
  _source?: string;

  // Pricing & Discounts
  maxdiscount?: string;
  discount?: string;
  isdividend?: boolean;
  isdiscountpointofpurchase?: boolean;
  fixedcost?: number | null;
  vatcal?: number;
  producttype?: { guidfixed?: string; code?: string; names?: NameX[] };

  // Marketplace & Logistics
  packageweight?: number;
  packagelength?: number;
  packagewidth?: number;
  packageheight?: number;
  marketplaceproducts?: MarketplaceProductMap[];
}

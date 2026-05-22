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
  shopid?: string;
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
  shopid?: string;
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
  balanceamount?: number;
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
  shopid: string;
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
    shopid: "",
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
  };
}

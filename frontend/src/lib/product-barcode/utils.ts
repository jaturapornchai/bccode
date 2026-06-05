/**
 * Pure helpers for Product Barcode payloads (no React).
 * Used by both the list/normalize path and the structured form.
 */

import { type LanguageCode } from "@/lib/i18n";
import type {
  BOMProductBarcode,
  NameX,
  Product,
  ProductBarcode,
  ProductBarcodeListRow,
  ProductPrice,
  RefProductBarcode,
} from "./types";

/** True if value is a plain JS object (record). */
export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

/** Read string at `key`, else "" — null/undefined safe. */
export function getString(record: Record<string, unknown> | undefined, key: string): string {
  if (!record) return "";
  const value = record[key];
  return typeof value === "string" ? value : "";
}

/** Read number at `key`, else default. */
export function getNumber(record: Record<string, unknown> | undefined, key: string, fallback = 0): number {
  if (!record) return fallback;
  const value = record[key];
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const n = Number(value);
    if (Number.isFinite(n)) return n;
  }
  return fallback;
}

/** Read boolean at `key`, accepts true/false, "true"/"false", 1/0. Empty/missing → fallback. */
export function getBoolean(record: Record<string, unknown> | undefined, key: string, fallback = false): boolean {
  if (!record) return fallback;
  const value = record[key];
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") {
    const lower = value.trim().toLowerCase();
    if (!lower) return fallback;
    if (["true", "1", "yes", "y"].includes(lower)) return true;
    if (["false", "0", "no", "n"].includes(lower)) return false;
  }
  return fallback;
}

/** Read first non-empty string from a list of keys (handles legacy aliases). */
export function getFirstString(record: Record<string, unknown> | undefined, keys: readonly string[]): string {
  if (!record) return "";
  for (const key of keys) {
    const value = record[key];
    if (typeof value === "string" && value.trim()) return value;
  }
  return "";
}

/** Convert "" → null, numeric string → number, else null. */
export function toNumberOrNull(value: string): number | null {
  const trimmed = value.trim();
  if (!trimmed) return null;
  const n = Number(trimmed);
  return Number.isFinite(n) ? n : null;
}

/** Normalize a NameX array. Accepts `[{code,name}]` or legacy `[{code,value}]`. */
export function toNameXArray(value: unknown): NameX[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    const code = getString(entry, "code") || getString(entry, "lang");
    const name = getString(entry, "name") || getString(entry, "value");
    if (!code) return [];
    return [{ code, name, description: getString(entry, "description") || undefined }];
  });
}

/** Pick localized name for given language with TH/EN fallback. */
export function pickName(names: NameX[] | undefined, language: LanguageCode | string | undefined): string {
  if (!names || names.length === 0) return "";
  const lang = String(language ?? "th").toLowerCase();
  const exact = names.find((n) => n.code?.toLowerCase() === lang);
  if (exact?.name) return exact.name;
  const th = names.find((n) => n.code?.toLowerCase() === "th");
  if (th?.name) return th.name;
  const en = names.find((n) => n.code?.toLowerCase() === "en");
  if (en?.name) return en.name;
  return names.find((n) => n.name)?.name ?? "";
}

/** Set or replace a NameX entry by code. */
export function setNameXEntry(names: NameX[] | undefined, code: string, name: string): NameX[] {
  const base = Array.isArray(names) ? [...names] : [];
  const idx = base.findIndex((entry) => entry.code === code);
  if (idx >= 0) {
    base[idx] = { ...base[idx], name };
  } else {
    base.push({ code, name });
  }
  return base;
}

/** Normalize price array — accept `{keynumber,price}` or `{keynumber,price}`. */
export function toPriceArray(value: unknown): ProductPrice[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    const keyNumber = getNumber(entry, "keynumber", getNumber(entry, "keynumber", 0));
    const price = getNumber(entry, "price", 0);
    return [{ keynumber: keyNumber, price }];
  });
}

/** Normalize ref-barcode array. */
export function toRefBarcodeArray(value: unknown): RefProductBarcode[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    return [
      {
        guidfixed: getFirstString(entry, ["guidfixed", "guidfixed"]),
        names: toNameXArray(entry.names),
        itemunitcode: getFirstString(entry, ["itemunitcode", "itemunitcode"]),
        itemunitnames: toNameXArray(entry.itemunitnames),
        barcode: getString(entry, "barcode"),
        condition: getBoolean(entry, "condition"),
        dividevalue: getNumber(entry, "dividevalue", 1),
        standvalue: getNumber(entry, "standvalue", 1),
        qty: getNumber(entry, "qty", 1),
      },
    ];
  });
}

export type ProductUnitOption = RefProductBarcode & {
  prices?: ProductPrice[];
  averagecost?: number;
  ismainbarcode?: boolean;
  productguid?: string;
  productcode?: string;
};

/** Normalize unit/barcode choices from product detail. */
export function toProductUnitOptions(product: unknown): ProductUnitOption[] {
  if (!isRecord(product)) return [];
  const productGuid = getFirstString(product, ["guidfixed", "guidfixed"]);
  const productCode = getString(product, "code");
  const productNames = toNameXArray(product.names);
  const source = Array.isArray(product.barcodes) ? product.barcodes : [];

  return source.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    const barcode = getString(entry, "barcode");
    if (!barcode) return [];
    return [
      {
        guidfixed: getFirstString(entry, ["guidfixed", "guidfixed", "barcodeguidfixed"]),
        names: toNameXArray(entry.names).length ? toNameXArray(entry.names) : productNames,
        itemunitcode: getFirstString(entry, ["itemunitcode", "itemunitcode", "unitcode"]),
        itemunitnames: toNameXArray(entry.itemunitnames).length
          ? toNameXArray(entry.itemunitnames)
          : toNameXArray(entry.unitnames),
        barcode,
        condition: getBoolean(entry, "condition", true),
        dividevalue: getNumber(entry, "dividevalue", 1),
        standvalue: getNumber(entry, "standvalue", 1),
        qty: getNumber(entry, "qty", 1),
        prices: toPriceArray(entry.prices),
        averagecost: getNumber(entry, "averagecost", getNumber(entry, "unitcost", 0)),
        ismainbarcode: getBoolean(entry, "ismainbarcode", false),
        productguid: productGuid,
        productcode: productCode,
      },
    ];
  });
}

/** Normalize BOM array. */
export function toBomArray(value: unknown): BOMProductBarcode[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    return [
      {
        barcodeguidfixed: getFirstString(entry, ["barcodeguidfixed", "guidfixed", "guidfixed"]),
        names: toNameXArray(entry.names),
        itemunitcode: getFirstString(entry, ["itemunitcode", "itemunitcode"]),
        itemunitnames: toNameXArray(entry.itemunitnames),
        barcode: getString(entry, "barcode"),
        qty: getNumber(entry, "qty", 1),
        standvalue: getNumber(entry, "standvalue", 1),
        dividevalue: getNumber(entry, "dividevalue", 1),
      },
    ];
  });
}

/** Read array of generic master-data entries (`{code,guid,names}`). */
export function toMasterArray(
  value: unknown,
  fields: { guid?: string; code?: string; names?: string } = {},
): { guidfixed: string; code: string; names: NameX[] }[] {
  if (!Array.isArray(value)) return [];
  const guidKey = fields.guid ?? "guidfixed";
  const codeKey = fields.code ?? "code";
  const namesKey = fields.names ?? "names";
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    return [
      {
        guidfixed: getString(entry, guidKey) || getString(entry, "guidfixed"),
        code: getString(entry, codeKey),
        names: toNameXArray(entry[namesKey]),
      },
    ];
  });
}

/** Convert a "list endpoint" row to the structured `ProductBarcode` shape (best-effort, partial). */
export function listRowToBarcode(row: ProductBarcodeListRow): Partial<ProductBarcode> {
  return {
    guidfixed: row.guidfixed,
    holdingcode: row.holdingcode,
    barcode: row.barcode,
    names: row.names ?? [],
    itemunitcode: row.itemunitcode ?? "",
    itemunitnames: row.itemunitnames ?? [],
    itemcode: row.itemcode ?? "",
    groupcode: row.groupcode ?? "",
    groupnames: row.groupnames ?? row.groupnames ?? [],
    brandcode: row.brandcode ?? "",
    brandnames: row.brandnames ?? [],
    categorycode: row.categorycode ?? "",
    categorynames: row.categorynames ?? [],
    prices: row.prices ?? [],
    imageuri: row.imageuri ?? "",
    qty: row.availableqty ?? row.balanceqty ?? 0,
    standvalue: row.standvalue ?? 1,
    dividevalue: row.dividevalue ?? 1,
    bom: row.bom ?? [],
  };
}

/** Validate barcode string — A-Z, 0-9, -, no space. */
export function isValidBarcode(value: string): boolean {
  return /^[A-Za-z0-9-]+$/.test(value.trim());
}

/** Normalize raw API JSON from `/api/product` list into a `Product` shape. */
export function rawToProduct(raw: unknown): Product {
  const r: Record<string, unknown> = isRecord(raw) ? raw : {};
  return {
    guidfixed: getFirstString(r, ["guidfixed", "guidfixed"]),
    holdingcode: getString(r, "holdingcode"),
    code: getString(r, "code"),
    names: toNameXArray(r.names),
    groupcode: getString(r, "groupcode"),
    groupnames: toNameXArray(r.groupnames),
    vattype: getNumber(r, "vattype", 0),
    itemtype: getNumber(r, "itemtype", 0),
    materialtype: getNumber(r, "materialtype", 0),
    taxtype: getNumber(r, "taxtype", 0),
    groupsuboneguid: getString(r, "groupsuboneguid"),
    groupsubonecode: getString(r, "groupsubonecode"),
    groupsubonenames: toNameXArray(r.groupsubonenames),
    groupsubtwoguid: getString(r, "groupsubtwoguid"),
    groupsubtwocode: getString(r, "groupsubtwocode"),
    groupsubtwonames: toNameXArray(r.groupsubtwonames),
    brandguid: getString(r, "brandguid"),
    brandcode: getString(r, "brandcode"),
    brandnames: toNameXArray(r.brandnames),
    designguid: getString(r, "designguid"),
    designcode: getString(r, "designcode"),
    designnames: toNameXArray(r.designnames),
    modelguid: getString(r, "modelguid"),
    modelcode: getString(r, "modelcode"),
    modelnames: toNameXArray(r.modelnames),
    patternguid: getString(r, "patternguid"),
    patterncode: getString(r, "patterncode"),
    patternnames: toNameXArray(r.patternnames),
    gradeguid: getString(r, "gradeguid"),
    gradecode: getString(r, "gradecode"),
    gradenames: toNameXArray(r.gradenames),
    categoryguid: getString(r, "categoryguid"),
    categorycode: getString(r, "categorycode"),
    categorynames: toNameXArray(r.categorynames),
    classguid: getString(r, "classguid"),
    classcode: getString(r, "classcode"),
    classnames: toNameXArray(r.classnames),
    manufacturers: Array.isArray(r.manufacturers) ? (r.manufacturers as Product["manufacturers"]) : [],
    suppliers: Array.isArray(r.suppliers) ? (r.suppliers as Product["suppliers"]) : [],
    condition: getBoolean(r, "condition", false),
    dividevalue: getNumber(r, "dividevalue", 1),
    standvalue: getNumber(r, "standvalue", 1),
    isusesubbarcodes: getBoolean(r, "isusesubbarcodes", false),
    refbarcodes: toRefBarcodeArray(r.refbarcodes),
    bom: toBomArray(r.bom),
    orderpoint: getNumber(r, "orderpoint", 0),
    minpoint: getNumber(r, "minpoint", 0),
    maxpoint: getNumber(r, "maxpoint", 0),
    qty: getNumber(r, "qty", 0),
    stockbarcode: getString(r, "stockbarcode"),
    _unit_count: getNumber(r, "_unit_count", 0),
    _source: getString(r, "_source"),
  };
}

/** Normalize raw API JSON into a full `ProductBarcode`, merging with defaults. */
export function rawToProductBarcode(raw: unknown, base: ProductBarcode): ProductBarcode {
  if (!isRecord(raw)) return base;
  const r = raw;
  return {
    ...base,
    guidfixed: getFirstString(r, ["guidfixed", "guidfixed"]) || base.guidfixed,
    holdingcode: getFirstString(r, ["holdingcode", "holdingcode"]) || base.holdingcode,
    itemcode: getString(r, "itemcode") || base.itemcode,
    barcode: getString(r, "barcode") || base.barcode,
    names: toNameXArray(r.names) || base.names,
    xsorts: Array.isArray(r.xsorts) ? (r.xsorts as ProductBarcode["xsorts"]) : base.xsorts,
    itemguid: getString(r, "itemguid") || base.itemguid,
    itemunitguid: getString(r, "itemunitguid") || base.itemunitguid,
    itemunitcode: getFirstString(r, ["itemunitcode", "itemunitcode"]) || base.itemunitcode,
    itemunitnames: toNameXArray(r.itemunitnames),
    itemunitsize: getNumber(r, "itemunitsize", base.itemunitsize),

    groupguid: getString(r, "groupguid") || base.groupguid,
    groupcode: getFirstString(r, ["groupcode", "groupcode"]) || base.groupcode,
    groupnames: toNameXArray(r.groupnames ?? r.groupnames),
    groupsubonecode: getString(r, "groupsubonecode"),
    groupsubonenames: toNameXArray(r.groupsubonenames),
    groupsuboneguid: getString(r, "groupsuboneguid"),
    groupsubtwoguid: getString(r, "groupsubtwoguid"),
    groupsubtwocode: getString(r, "groupsubtwocode"),
    groupsubtwonames: toNameXArray(r.groupsubtwonames),

    brandguid: getString(r, "brandguid"),
    brandcode: getFirstString(r, ["brandcode", "brandcode"]),
    brandnames: toNameXArray(r.brandnames),
    designguid: getString(r, "designguid"),
    designcode: getString(r, "designcode"),
    designnames: toNameXArray(r.designnames),
    modelguid: getString(r, "modelguid"),
    modelcode: getString(r, "modelcode"),
    modelnames: toNameXArray(r.modelnames),
    patternguid: getString(r, "patternguid"),
    patterncode: getString(r, "patterncode"),
    patternnames: toNameXArray(r.patternnames),
    gradeguid: getString(r, "gradeguid"),
    gradecode: getString(r, "gradecode"),
    gradenames: toNameXArray(r.gradenames),
    categoryguid: getFirstString(r, ["categoryguid", "categoryguid"]),
    categorycode: getString(r, "categorycode"),
    categorynames: toNameXArray(r.categorynames ?? r.categorynames),
    classguid: getString(r, "classguid"),
    classcode: getString(r, "classcode"),
    classnames: toNameXArray(r.classnames),
    manufacturerguid: getString(r, "manufacturerguid"),
    manufacturercode: getString(r, "manufacturercode"),
    manufacturernames: toNameXArray(r.manufacturernames),

    orderpoint: getNumber(r, "orderpoint", base.orderpoint),
    minpoint: getNumber(r, "minpoint", base.minpoint),
    maxpoint: getNumber(r, "maxpoint", base.maxpoint),
    qty: getNumber(r, "qty", base.qty),
    stockbarcode: getString(r, "stockbarcode") || base.stockbarcode,
    refguidfixed: getString(r, "refguidfixed"),
    refdividevalue: getNumber(r, "refdividevalue", 0),
    refstandvalue: getNumber(r, "refstandvalue", 0),
    refunitnames: toNameXArray(r.refunitnames),

    condition: getBoolean(r, "condition", base.condition),
    dividevalue: getNumber(r, "dividevalue", base.dividevalue),
    standvalue: getNumber(r, "standvalue", base.standvalue),
    isusesubbarcodes: getBoolean(r, "isusesubbarcodes", base.isusesubbarcodes),
    ismainbarcode: getBoolean(r, "ismainbarcode", base.ismainbarcode),

    prices: toPriceArray(r.prices),
    fixedcost: Array.isArray(r.fixedcost) ? (r.fixedcost as ProductBarcode["fixedcost"]) : base.fixedcost,
    discount: getString(r, "discount") || base.discount,
    maxdiscount: getString(r, "maxdiscount") || base.maxdiscount,
    isdividend: getBoolean(r, "isdividend", base.isdividend),
    isdiscountpointofpurchase: getBoolean(r, "isdiscountpointofpurchase", base.isdiscountpointofpurchase),

    imageuri: getString(r, "imageuri") || base.imageuri,
    images: Array.isArray(r.images) ? (r.images as ProductBarcode["images"]) : base.images,
    useimageorcolor: getBoolean(r, "useimageorcolor", base.useimageorcolor),
    colorselect: getString(r, "colorselect"),
    colorselecthex: getString(r, "colorselecthex"),

    itemtype: getNumber(r, "itemtype", base.itemtype) as ProductBarcode["itemtype"],
    materialtype: getNumber(r, "materialtype", base.materialtype) as ProductBarcode["materialtype"],
    taxtype: getNumber(r, "taxtype", base.taxtype),
    vattype: getNumber(r, "vattype", base.vattype) as ProductBarcode["vattype"],
    vatcal: getNumber(r, "vatcal", base.vatcal),
    producttype: getNumber(r, "producttype", base.producttype),
    foodtype: getNumber(r, "foodtype", base.foodtype) as ProductBarcode["foodtype"],
    issumpoint: getBoolean(r, "issumpoint", base.issumpoint),
    isalacarte: getBoolean(r, "isalacarte", base.isalacarte),
    isstockforrestaurant: getBoolean(r, "isstockforrestaurant", base.isstockforrestaurant),
    issplitunitprint: getBoolean(r, "issplitunitprint", base.issplitunitprint),
    isonlystaff: getBoolean(r, "isonlystaff", base.isonlystaff),

    restaurant: isRecord(r.restaurant)
      ? {
          isforrestaurant: getBoolean(r.restaurant, "isforrestaurant"),
          isfortakeaway: getBoolean(r.restaurant, "isfortakeaway"),
          isfordelivery: getBoolean(r.restaurant, "isfordelivery"),
          isforcustomer: getBoolean(r.restaurant, "isforcustomer"),
          isforcustomerpreorder: getBoolean(r.restaurant, "isforcustomerpreorder"),
        }
      : base.restaurant,
    ordertypes: Array.isArray(r.ordertypes) ? (r.ordertypes as ProductBarcode["ordertypes"]) : base.ordertypes,
    options: Array.isArray(r.options) ? (r.options as ProductBarcode["options"]) : base.options,

    refbarcodes: toRefBarcodeArray(r.refbarcodes),
    bom: toBomArray(r.bom),

    businesstypes: Array.isArray(r.businesstypes)
      ? (r.businesstypes as ProductBarcode["businesstypes"])
      : base.businesstypes,
    ignorebranches: Array.isArray(r.ignorebranches)
      ? (r.ignorebranches as ProductBarcode["ignorebranches"])
      : base.ignorebranches,
    timeforsales: Array.isArray(r.timeforsales)
      ? (r.timeforsales as ProductBarcode["timeforsales"])
      : base.timeforsales,
    dimensions: Array.isArray(r.dimensions) ? (r.dimensions as ProductBarcode["dimensions"]) : base.dimensions,

    isalert: getBoolean(r, "isalert", base.isalert),
    alertdescription: getString(r, "alertdescription"),
    description: getString(r, "description"),
  };
}

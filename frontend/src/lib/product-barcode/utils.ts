/**
 * Pure helpers for Product Barcode payloads (no React).
 * Used by both the list/normalize path and the structured form.
 */

import { type LanguageCode } from "@/lib/i18n";
import type {
  BOMProductBarcode,
  NameX,
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

/** Normalize price array — accept `{key_number,price}` or `{keynumber,price}`. */
export function toPriceArray(value: unknown): ProductPrice[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    const keyNumber = getNumber(entry, "key_number", getNumber(entry, "keynumber", 0));
    const price = getNumber(entry, "price", 0);
    return [{ key_number: keyNumber, price }];
  });
}

/** Normalize ref-barcode array. */
export function toRefBarcodeArray(value: unknown): RefProductBarcode[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    return [
      {
        guid_fixed: getFirstString(entry, ["guid_fixed", "guidfixed"]),
        names: toNameXArray(entry.names),
        item_unit_code: getFirstString(entry, ["item_unit_code", "itemunitcode"]),
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

/** Normalize BOM array. */
export function toBomArray(value: unknown): BOMProductBarcode[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((entry) => {
    if (!isRecord(entry)) return [];
    return [
      {
        barcodeguidfixed: getFirstString(entry, ["barcodeguidfixed", "guid_fixed", "guidfixed"]),
        names: toNameXArray(entry.names),
        item_unit_code: getFirstString(entry, ["item_unit_code", "itemunitcode"]),
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
        guidfixed: getString(entry, guidKey) || getString(entry, "guid_fixed"),
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
    shopid: row.shopid,
    barcode: row.barcode,
    names: row.names ?? [],
    item_unit_code: row.itemunitcode ?? "",
    itemunitnames: row.itemunitnames ?? [],
    itemcode: row.itemcode ?? "",
    group_code: row.groupcode ?? "",
    group_names: row.group_names ?? row.groupnames ?? [],
    brand_code: row.brand_code ?? "",
    brandnames: row.brandnames ?? [],
    categorycode: row.categorycode ?? "",
    category_names: row.category_names ?? [],
    prices: row.prices ?? [],
    imageuri: row.imageuri ?? "",
    qty: row.balance_qty ?? 0,
    standvalue: row.standvalue ?? 1,
    dividevalue: row.dividevalue ?? 1,
    bom: row.bom ?? [],
  };
}

/** Validate barcode string — A-Z, 0-9, -, no space. */
export function isValidBarcode(value: string): boolean {
  return /^[A-Za-z0-9-]+$/.test(value.trim());
}

/** Normalize raw API JSON into a full `ProductBarcode`, merging with defaults. */
export function rawToProductBarcode(raw: unknown, base: ProductBarcode): ProductBarcode {
  if (!isRecord(raw)) return base;
  const r = raw;
  return {
    ...base,
    guidfixed: getFirstString(r, ["guidfixed", "guid_fixed"]) || base.guidfixed,
    shopid: getFirstString(r, ["shopid", "shop_id"]) || base.shopid,
    itemcode: getString(r, "itemcode") || base.itemcode,
    barcode: getString(r, "barcode") || base.barcode,
    names: toNameXArray(r.names) || base.names,
    xsorts: Array.isArray(r.xsorts) ? (r.xsorts as ProductBarcode["xsorts"]) : base.xsorts,
    item_guid: getString(r, "item_guid") || base.item_guid,
    itemunitguid: getString(r, "itemunitguid") || base.itemunitguid,
    item_unit_code: getFirstString(r, ["item_unit_code", "itemunitcode"]) || base.item_unit_code,
    itemunitnames: toNameXArray(r.itemunitnames),
    itemunitsize: getNumber(r, "itemunitsize", base.itemunitsize),

    groupguid: getString(r, "groupguid") || base.groupguid,
    group_code: getFirstString(r, ["group_code", "groupcode"]) || base.group_code,
    group_names: toNameXArray(r.group_names ?? r.groupnames),
    groupsubonecode: getString(r, "groupsubonecode"),
    groupsubonenames: toNameXArray(r.groupsubonenames),
    groupsuboneguid: getString(r, "groupsuboneguid"),
    groupsubtwoguid: getString(r, "groupsubtwoguid"),
    groupsubtwocode: getString(r, "groupsubtwocode"),
    groupsubtwonames: toNameXArray(r.groupsubtwonames),

    brandguid: getString(r, "brandguid"),
    brand_code: getFirstString(r, ["brand_code", "brandcode"]),
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
    category_guid: getFirstString(r, ["category_guid", "categoryguid"]),
    categorycode: getString(r, "categorycode"),
    category_names: toNameXArray(r.category_names ?? r.categorynames),
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
    is_main_barcode: getBoolean(r, "is_main_barcode", base.is_main_barcode),

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

    item_type: getNumber(r, "item_type", base.item_type) as ProductBarcode["item_type"],
    materialtype: getNumber(r, "materialtype", base.materialtype) as ProductBarcode["materialtype"],
    tax_type: getNumber(r, "tax_type", base.tax_type),
    vat_type: getNumber(r, "vat_type", base.vat_type) as ProductBarcode["vat_type"],
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

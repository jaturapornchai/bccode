import { describe, expect, it } from "vitest";
import {
  getBoolean,
  getFirstString,
  getNumber,
  getString,
  isRecord,
  isValidBarcode,
  pickName,
  rawToProductBarcode,
  setNameXEntry,
  toBomArray,
  toNameXArray,
  toNumberOrNull,
  toPriceArray,
  toRefBarcodeArray,
} from "./utils";
import { emptyProductBarcode, ITEM_TYPE } from "./types";

describe("utils — primitives", () => {
  it("isRecord recognizes objects only", () => {
    expect(isRecord({})).toBe(true);
    expect(isRecord([])).toBe(false);
    expect(isRecord(null)).toBe(false);
    expect(isRecord("x")).toBe(false);
  });

  it("getString returns string or empty", () => {
    expect(getString({ a: "hi" }, "a")).toBe("hi");
    expect(getString({ a: 1 }, "a")).toBe("");
    expect(getString(undefined, "a")).toBe("");
  });

  it("getNumber falls back when value not numeric", () => {
    expect(getNumber({ x: 12 }, "x")).toBe(12);
    expect(getNumber({ x: "8.5" }, "x")).toBe(8.5);
    expect(getNumber({ x: "abc" }, "x", -1)).toBe(-1);
    expect(getNumber(undefined, "x", 0)).toBe(0);
  });

  it("getBoolean accepts multiple truthy representations", () => {
    expect(getBoolean({ b: true }, "b")).toBe(true);
    expect(getBoolean({ b: 1 }, "b")).toBe(true);
    expect(getBoolean({ b: "true" }, "b")).toBe(true);
    expect(getBoolean({ b: 0 }, "b")).toBe(false);
    expect(getBoolean({ b: "no" }, "b")).toBe(false);
    expect(getBoolean({ b: "" }, "b", true)).toBe(true);
  });

  it("getFirstString picks first non-empty key", () => {
    expect(getFirstString({ a: "", b: "ok", c: "later" }, ["a", "b", "c"])).toBe("ok");
    expect(getFirstString({ a: "" }, ["a"])).toBe("");
  });

  it("toNumberOrNull handles trims and invalid", () => {
    expect(toNumberOrNull("12")).toBe(12);
    expect(toNumberOrNull("  ")).toBeNull();
    expect(toNumberOrNull("abc")).toBeNull();
  });
});

describe("utils — names", () => {
  it("toNameXArray normalizes legacy `value` to `name`", () => {
    expect(
      toNameXArray([
        { code: "th", value: "สวัสดี" },
        { code: "en", name: "Hi" },
        { code: "", name: "ignored" },
      ]),
    ).toEqual([
      { code: "th", name: "สวัสดี", description: undefined },
      { code: "en", name: "Hi", description: undefined },
    ]);
  });

  it("pickName prefers requested lang then th, en, then first non-empty", () => {
    const names = [
      { code: "th", name: "ไทย" },
      { code: "en", name: "EN" },
    ];
    expect(pickName(names, "th")).toBe("ไทย");
    expect(pickName(names, "en")).toBe("EN");
    expect(pickName(names, "lo")).toBe("ไทย");
    expect(pickName(undefined, "th")).toBe("");
    expect(pickName([], "th")).toBe("");
    expect(pickName([{ code: "fr", name: "Salut" }], "th")).toBe("Salut");
  });

  it("setNameXEntry replaces existing code, appends new code", () => {
    const start = [{ code: "th", name: "ก" }];
    expect(setNameXEntry(start, "th", "ข")).toEqual([{ code: "th", name: "ข" }]);
    expect(setNameXEntry(start, "en", "B")).toEqual([
      { code: "th", name: "ก" },
      { code: "en", name: "B" },
    ]);
  });
});

describe("utils — price/refbarcode/bom arrays", () => {
  it("toPriceArray supports key_number and legacy keynumber", () => {
    expect(
      toPriceArray([
        { key_number: 1, price: 10 },
        { keynumber: 2, price: 20 },
        { price: 30 },
      ]),
    ).toEqual([
      { key_number: 1, price: 10 },
      { key_number: 2, price: 20 },
      { key_number: 0, price: 30 },
    ]);
  });

  it("toRefBarcodeArray defaults divide/stand/qty to 1", () => {
    const result = toRefBarcodeArray([{ barcode: "X1" }]);
    expect(result).toEqual([
      {
        guid_fixed: "",
        names: [],
        item_unit_code: "",
        itemunitnames: [],
        barcode: "X1",
        condition: false,
        dividevalue: 1,
        standvalue: 1,
        qty: 1,
      },
    ]);
  });

  it("toBomArray normalizes barcode reference fields", () => {
    expect(toBomArray([{ barcode: "X1", barcodeguidfixed: "G1", qty: 2 }])).toEqual([
      {
        barcodeguidfixed: "G1",
        names: [],
        item_unit_code: "",
        itemunitnames: [],
        barcode: "X1",
        qty: 2,
        standvalue: 1,
        dividevalue: 1,
      },
    ]);
  });
});

describe("utils — barcode validation", () => {
  it("isValidBarcode accepts A-Z 0-9 -", () => {
    expect(isValidBarcode("ABC-123")).toBe(true);
    expect(isValidBarcode("X")).toBe(true);
  });
  it("isValidBarcode rejects spaces, lowercase-only-bad-chars", () => {
    expect(isValidBarcode("AB C")).toBe(false);
    expect(isValidBarcode("AB#1")).toBe(false);
    expect(isValidBarcode("")).toBe(false);
  });
});

describe("utils — rawToProductBarcode", () => {
  it("merges raw payload over base defaults", () => {
    const base = emptyProductBarcode();
    const merged = rawToProductBarcode(
      {
        barcode: "ABC-1",
        names: [{ code: "th", name: "สินค้า" }],
        item_unit_code: "PCS",
        prices: [{ key_number: 1, price: 99 }],
        item_type: 2,
      },
      base,
    );
    expect(merged.barcode).toBe("ABC-1");
    expect(merged.names).toEqual([{ code: "th", name: "สินค้า", description: undefined }]);
    expect(merged.item_unit_code).toBe("PCS");
    expect(merged.prices).toEqual([{ key_number: 1, price: 99 }]);
    expect(merged.item_type).toBe(ITEM_TYPE.SET);
  });

  it("returns base unchanged for non-record input", () => {
    const base = emptyProductBarcode();
    expect(rawToProductBarcode(null, base)).toBe(base);
    expect(rawToProductBarcode("oops", base)).toBe(base);
  });
});

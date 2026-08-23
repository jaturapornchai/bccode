import { describe, expect, it } from "vitest";
import {
  ean13CheckDigit,
  encodeEan13,
  formatProductBalance,
  getBoolean,
  getFirstString,
  getNumber,
  getString,
  isRecord,
  isValidBarcode,
  pickName,
  rawToProduct,
  rawToProductBarcode,
  setNameXEntry,
  toBomArray,
  toNameXArray,
  toNumberOrNull,
  toPriceArray,
  toRefBarcodeArray,
  toProductUnitOptions,
} from "./utils";
import { emptyProductBarcode, ITEM_TYPE, type Product } from "./types";

describe("utils — Product balance", () => {
  const product = (overrides: Partial<Product>): Product => ({
    guidfixed: "P1",
    holdingcode: "H1",
    code: "P1",
    names: [{ code: "th", name: "สินค้า" }],
    groupcode: "",
    groupnames: [],
    qty: 0,
    unitcode: "EA",
    unitnames: [{ code: "th", name: "ชิ้น" }],
    unitconversions: [],
    ...overrides,
  });

  it("preserves fractional base-unit balance", () => {
    expect(
      formatProductBalance(
        product({
          qty: 25.5,
          unitconversions: [
            {
              unitcode: "BOX",
              unitnames: [{ code: "th", name: "กล่อง" }],
              dividevalue: 1,
              standvalue: 12,
            },
          ],
        }),
        "th",
      ),
    ).toBe("2 กล่อง + 1.5 ชิ้น");
  });

  it("shows non-terminating remainder as an exact fraction", () => {
    expect(
      formatProductBalance(
        product({
          qty: 2,
          unitconversions: [
            {
              unitcode: "PACK",
              unitnames: [{ code: "th", name: "แพ็ก" }],
              dividevalue: 3,
              standvalue: 4,
            },
          ],
        }),
        "th",
      ),
    ).toBe("1 แพ็ก + 2/3 ชิ้น");
  });

  it("ignores invalid ratios and clamps negative balances to zero", () => {
    expect(
      formatProductBalance(
        product({
          qty: -3,
          unitconversions: [
            {
              unitcode: "BAD",
              unitnames: [{ code: "th", name: "เสีย" }],
              dividevalue: 0,
              standvalue: 1,
            },
          ],
        }),
        "th",
      ),
    ).toBe("0 ชิ้น");
  });
});

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
    expect(
      getFirstString({ a: "", b: "ok", c: "later" }, ["a", "b", "c"]),
    ).toBe("ok");
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
    expect(setNameXEntry(start, "th", "ข")).toEqual([
      { code: "th", name: "ข" },
    ]);
    expect(setNameXEntry(start, "en", "B")).toEqual([
      { code: "th", name: "ก" },
      { code: "en", name: "B" },
    ]);
  });
});

describe("utils — price/refbarcode/bom arrays", () => {
  it("toPriceArray supports keynumber and legacy keynumber", () => {
    expect(
      toPriceArray([
        { keynumber: 1, price: 10 },
        { keynumber: 2, price: 20 },
        { price: 30 },
      ]),
    ).toEqual([
      { keynumber: 1, price: 10 },
      { keynumber: 2, price: 20 },
      { keynumber: 0, price: 30 },
    ]);
  });

  it("toRefBarcodeArray defaults divide/stand/qty to 1", () => {
    const result = toRefBarcodeArray([{ barcode: "X1" }]);
    expect(result).toEqual([
      {
        guidfixed: "",
        itemcode: "",
        names: [],
        itemunitcode: "",
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
    expect(
      toBomArray([{ barcode: "X1", barcodeguidfixed: "G1", qty: 2 }]),
    ).toEqual([
      {
        barcodeguidfixed: "G1",
        itemcode: "",
        names: [],
        itemunitcode: "",
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
    expect(isValidBarcode("4006381333931")).toBe(true);
  });
  it("isValidBarcode rejects spaces, lowercase-only-bad-chars", () => {
    expect(isValidBarcode("AB C")).toBe(false);
    expect(isValidBarcode("AB#1")).toBe(false);
    expect(isValidBarcode("")).toBe(false);
    expect(isValidBarcode("4006381333932")).toBe(false);
  });
});

describe("utils — ean13CheckDigit", () => {
  it("computes the correct check digit for a known-good EAN-13", () => {
    // 4006381333931 is a real, valid EAN-13 (Ferrero Kinder base example).
    expect(ean13CheckDigit("400638133393")).toBe("1");
  });
  it("computes the correct check digit for a zero-padded base", () => {
    expect(ean13CheckDigit("000000000000")).toBe("0");
  });
});

describe("utils — encodeEan13", () => {
  it("encodes a known EAN-13 into the standard 95 modules", () => {
    expect(encodeEan13("4006381333931")).toBe(
      [
        "101",
        "0001101",
        "0100111",
        "0101111",
        "0111101",
        "0001001",
        "0110011",
        "01010",
        "1000010",
        "1000010",
        "1000010",
        "1110100",
        "1000010",
        "1100110",
        "101",
      ].join(""),
    );
  });

  it("rejects a bad check digit and non-EAN value", () => {
    expect(encodeEan13("4006381333932")).toBeNull();
    expect(encodeEan13("ABC-123")).toBeNull();
  });
});

describe("utils — rawToProductBarcode", () => {
  it("merges raw payload over base defaults", () => {
    const base = emptyProductBarcode();
    const merged = rawToProductBarcode(
      {
        barcode: "ABC-1",
        names: [{ code: "th", name: "สินค้า" }],
        itemunitcode: "PCS",
        prices: [{ keynumber: 1, price: 99 }],
        videos: [
          {
            xorder: 1,
            uri: "/goapi/s3/file/H/companies/C/products/videos/demo.mp4",
          },
        ],
        itemtype: 2,
      },
      base,
    );
    expect(merged.barcode).toBe("ABC-1");
    expect(merged.names).toEqual([
      { code: "th", name: "สินค้า", description: undefined },
    ]);
    expect(merged.itemunitcode).toBe("PCS");
    expect(merged.prices).toEqual([{ keynumber: 1, price: 99 }]);
    expect(merged.videos).toEqual([
      {
        xorder: 1,
        uri: "/goapi/s3/file/H/companies/C/products/videos/demo.mp4",
      },
    ]);
    expect(merged.itemtype).toBe(ITEM_TYPE.SET);
  });

  it("returns base unchanged for non-record input", () => {
    const base = emptyProductBarcode();
    expect(rawToProductBarcode(null, base)).toBe(base);
    expect(rawToProductBarcode("oops", base)).toBe(base);
  });
});

describe("utils — rawToProduct", () => {
  it("keeps Product units independent from Barcode references", () => {
    const product = rawToProduct({
      code: "P001",
      unitcode: "PCS",
      unitnames: [{ code: "th", name: "ชิ้น" }],
      unitconversions: [
        {
          unitcode: "BOX",
          unitnames: [{ code: "th", name: "กล่อง" }],
          dividevalue: 1,
          standvalue: 12,
        },
      ],
    });
    expect(product.unitcode).toBe("PCS");
    expect(product.unitconversions).toEqual([
      {
        unitcode: "BOX",
        unitnames: [{ code: "th", name: "กล่อง", description: undefined }],
        dividevalue: 1,
        standvalue: 12,
      },
    ]);
  });
});

describe("utils — toProductUnitOptions", () => {
  it("uses product.barcodes as the unit choices returned by /api/product detail", () => {
    const options = toProductUnitOptions({
      guidfixed: "PRODUCT-GUID",
      code: "P001",
      names: [{ code: "th", name: "สินค้า" }],
      barcodes: [
        {
          guidfixed: "BARCODE-GUID-1",
          barcode: "885-PCS",
          itemunitcode: "PCS",
          itemunitnames: [{ code: "th", name: "ชิ้น" }],
        },
        {
          guidfixed: "BARCODE-GUID-2",
          barcode: "885-BOX",
          itemunitcode: "BOX",
          itemunitnames: [{ code: "th", name: "กล่อง" }],
        },
      ],
      refbarcodes: [{ barcode: "WRONG", itemunitcode: "OLD" }],
    });

    expect(options).toEqual([
      expect.objectContaining({
        guidfixed: "BARCODE-GUID-1",
        barcode: "885-PCS",
        itemunitcode: "PCS",
      }),
      expect.objectContaining({
        guidfixed: "BARCODE-GUID-2",
        barcode: "885-BOX",
        itemunitcode: "BOX",
      }),
    ]);
  });

  it("does not invent a barcode when a product has no barcode/unit choices", () => {
    expect(
      toProductUnitOptions({
        guidfixed: "PRODUCT-GUID",
        code: "P001",
        names: [],
      }),
    ).toEqual([]);
  });

  it("does not use legacy refbarcodes as recipe unit choices", () => {
    expect(
      toProductUnitOptions({
        guidfixed: "PRODUCT-GUID",
        code: "P001",
        names: [],
        refbarcodes: [{ barcode: "LEGACY", itemunitcode: "OLD" }],
      }),
    ).toEqual([]);
  });
});

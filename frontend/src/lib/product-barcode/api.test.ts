import { describe, expect, it } from "vitest";
import { toQuickBarcodePayload } from "./api";
import { emptyProductBarcode } from "./types";

describe("toQuickBarcodePayload", () => {
  it("sends only core barcode fields and leaves tenant scope to the session", () => {
    const barcode = emptyProductBarcode();
    Object.assign(barcode, {
      holdingcode: " HOLDING-01 ",
      businesscode: " COMPANY-01 ",
      barcode: " 8850000000001 ",
      itemcode: " item-01 ",
      names: [{ code: "TH", name: " สินค้าทดสอบ " }],
      itemunitguid: "unit-guid",
      itemunitcode: " pcs ",
      itemunitnames: [{ code: "TH", name: " ชิ้น " }],
      dividevalue: 1,
      standvalue: 12,
      ismainbarcode: true,
      qty: 99,
      prices: [{ keynumber: 1, price: 250 }],
      fixedcost: [{ branchguid: "branch", cost: 100 }],
      groupcode: "GROUP",
      imageuri: " /goapi/s3/file/HOLDING-01/products/main.jpg ",
      images: [
        { xorder: 8, uri: " /goapi/s3/file/HOLDING-01/products/gallery.jpg " },
        { xorder: 9, uri: " " },
      ],
      videos: [
        {
          xorder: 4,
          uri: " /goapi/s3/file/HOLDING-01/companies/COMPANY-01/products/videos/demo.mp4 ",
          posteruri: " /goapi/s3/file/HOLDING-01/companies/COMPANY-01/products/video-posters/demo.jpg ",
        },
        { xorder: 5, uri: " ", posteruri: " " },
      ],
      description: " รายละเอียดเฉพาะบาร์โค้ด ",
      bom: [{ itemcode: "CHILD" }],
      marketplaceproducts: [{ platform: "shopee" }],
    });

    const payload = toQuickBarcodePayload(barcode);

    expect(payload).toEqual({
      barcode: "8850000000001",
      itemcode: "ITEM-01",
      names: [{ code: "th", name: "สินค้าทดสอบ" }],
      itemunitguid: "unit-guid",
      itemunitcode: "PCS",
      itemunitnames: [{ code: "th", name: "ชิ้น" }],
      dividevalue: 1,
      standvalue: 12,
      ismainbarcode: true,
      imageuri: "/goapi/s3/file/HOLDING-01/products/main.jpg",
      images: [
        {
          xorder: 1,
          uri: "/goapi/s3/file/HOLDING-01/products/gallery.jpg",
        },
      ],
      videos: [
        {
          xorder: 1,
          uri: "/goapi/s3/file/HOLDING-01/companies/COMPANY-01/products/videos/demo.mp4",
          posteruri: "/goapi/s3/file/HOLDING-01/companies/COMPANY-01/products/video-posters/demo.jpg",
        },
      ],
      description: "รายละเอียดเฉพาะบาร์โค้ด",
    });
    expect(payload).not.toHaveProperty("qty");
    expect(payload).not.toHaveProperty("prices");
    expect(payload).not.toHaveProperty("fixedcost");
    expect(payload).not.toHaveProperty("groupcode");
    expect(payload).not.toHaveProperty("bom");
    expect(payload).not.toHaveProperty("marketplaceproducts");
    expect(payload).not.toHaveProperty("holdingcode");
    expect(payload).not.toHaveProperty("businesscode");
    expect(payload).not.toHaveProperty("itemguid");
  });
});

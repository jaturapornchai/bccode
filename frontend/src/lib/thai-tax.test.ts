import { afterEach, describe, expect, it, vi } from "vitest";
import {
  THAI_TAX_CONFIGS,
  fetchPp30Summary,
  fetchVatRegister,
  getThaiTaxConfig,
  isThaiTaxRoute,
} from "./thai-tax";
import { setupTestAuthSession } from "./test-auth-session";

setupTestAuthSession();

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  } as unknown as Response;
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("thai tax configs", () => {
  it("ให้ config ครบ 12 รายการ และ resolve ได้จาก route", () => {
    expect(THAI_TAX_CONFIGS).toHaveLength(12);
    THAI_TAX_CONFIGS.forEach((config) => {
      expect(isThaiTaxRoute(config.route)).toBe(true);
      expect(getThaiTaxConfig(config.route)).toEqual(config);
    });
  });
});

describe("fetchVatRegister", () => {
  it("200 → map ข้อมูลจาก API เป็น ThaiTaxRecord", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(200, {
        status: "success",
        data: [
          {
            docdate: "2026-09-01",
            taxinvoiceno: "INV-001",
            counterpartyname: "บริษัท ตัวอย่าง จำกัด",
            taxid: "1101700000000",
            branchno: "00000",
            amountbeforevat: 10000,
            vatamount: 700,
            totalamount: 10700,
          },
        ],
        count: 1,
        limit: 200,
        offset: 0,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchVatRegister({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
      type: "sale",
    });

    expect(result.error).toBeUndefined();
    expect(result.records).toHaveLength(1);
    expect(result.total).toBe(1);
    expect(result.records[0]?.vatamount).toBe(700);
    expect(result.records[0]?.amountbeforevat).toBe(10000);
    expect(result.records[0]?.taxinvoiceno).toBe("INV-001");
    expect(result.records[0]?.status).toBe("active");
    expect(result.records[0]?.isheadoffice).toBe(true);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/goapi/api/report/tax/vat-register",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("401 → unauthorized และ records ว่าง", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(401, { status: "error" }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchVatRegister({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
      type: "sale",
    });

    expect(result.error).toBe("unauthorized");
    expect(result.records).toEqual([]);
  });

  it("500 → load_failed และ records ว่าง", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(500, { status: "error" }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchVatRegister({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
      type: "purchase",
    });

    expect(result.error).toBe("load_failed");
    expect(result.records).toEqual([]);
  });

  it("fetch throw → connection_error", async () => {
    const fetchMock = vi.fn().mockRejectedValue(new Error("network down"));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchVatRegister({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
      type: "sale",
    });

    expect(result.error).toBe("connection_error");
    expect(result.records).toEqual([]);
  });

  it("ไม่มี businesscode → company_required และไม่ยิง fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchVatRegister({
      holdingcode: "H001",
      businesscode: "",
      year: 2026,
      month: 9,
      type: "sale",
    });

    expect(result.error).toBe("company_required");
    expect(result.records).toEqual([]);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

describe("fetchPp30Summary", () => {
  it("200 → คืนค่าตามที่ API ส่งมาทุกฟิลด์ ไม่คำนวณ VAT ใหม่จากฐาน", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(200, {
        status: "success",
        data: {
          year: 2026,
          month: 9,
          salestaxable: 100000,
          saleszerorated: 5000,
          salesexempt: 2000,
          outputvat: 6999.37,
          purchasetaxable: 40000,
          inputvat: 2800.12,
          netvat: 4199.25,
          payable: 4199.25,
          creditable: 0,
        },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchPp30Summary({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
    });

    expect(result.error).toBeUndefined();
    expect(result.summary).toEqual({
      year: 2026,
      month: 9,
      salestaxable: 100000,
      saleszerorated: 5000,
      salesexempt: 2000,
      outputvat: 6999.37,
      purchasetaxable: 40000,
      inputvat: 2800.12,
      netvat: 4199.25,
      payable: 4199.25,
      creditable: 0,
    });
    // 100000 × 0.07 = 7000 จึงยืนยันว่าไม่ได้คำนวณซ้ำบน browser
    expect(result.summary?.outputvat).toBe(6999.37);
    expect(result.summary?.outputvat).not.toBe(7000);
  });

  it("500 → summary เป็น null และมี error key", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(500, { status: "error" }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchPp30Summary({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
    });

    expect(result.summary).toBeNull();
    expect(result.error).toBe("load_failed");
  });
});

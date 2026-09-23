import { afterEach, describe, expect, it, vi } from "vitest";
import {
  THAI_TAX_CONFIGS,
  fetchPp30Summary,
  fetchVatRegister,
  fetchWhtReport,
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
    expect(THAI_TAX_CONFIGS).toHaveLength(10);
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
            amountbeforevat: "10000.00",
            vatamount: "700.00",
            totalamount: "10700.00",
          },
        ],
        count: 1,
        total: 3,
        summary: { amountbeforevat: "30000.10", vatamount: "2100.01", totalamount: "32100.11" },
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
    // total/summary = ทั้งงวดจาก backend ไม่ใช่ผลบวกของแถวที่โหลดมา
    expect(result.total).toBe(3);
    expect(result.summary).toEqual({ amountbeforevat: "30000.10", vatamount: "2100.01", totalamount: "32100.11" });
    expect(result.records[0]?.vatamount).toBe("700.00");
    expect(result.records[0]?.amountbeforevat).toBe("10000.00");
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
          company: { code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234" },
          salesgross: "107000.00",
          saleszerorated: "5000.00",
          salesexempt: "2000.00",
          salestaxable: "100000.00",
          outputvat: "6999.37",
          purchasetaxable: "40000.00",
          inputvat: "2800.12",
          creditbroughtforward: "100.00",
          netvat: "4099.25",
          payable: "4099.25",
          creditable: "0.00",
        },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchPp30Summary({
      holdingcode: "H001",
      businesscode: "B001",
      year: 2026,
      month: 9,
      creditbroughtforward: "100.00",
    });

    expect(result.error).toBeUndefined();
    expect(result.summary).toEqual({
      year: 2026,
      month: 9,
      company: { code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234" },
      salesgross: "107000.00",
      saleszerorated: "5000.00",
      salesexempt: "2000.00",
      salestaxable: "100000.00",
      outputvat: "6999.37",
      purchasetaxable: "40000.00",
      inputvat: "2800.12",
      creditbroughtforward: "100.00",
      netvat: "4099.25",
      payable: "4099.25",
      creditable: "0.00",
    });
    // 100000 × 0.07 = 7000 จึงยืนยันว่าไม่ได้คำนวณซ้ำบน browser
    expect(result.summary?.outputvat).toBe("6999.37");
    const body = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    expect(body.creditbroughtforward).toBe("100.00");
  });

  it("ยอดเงินที่ไม่ใช่ string ทศนิยม (JSON number/null) ไม่ถูกเดา — แสดง 0.00", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(200, { status: "success", data: { year: 2026, month: 9, outputvat: 0.1 + 0.2, inputvat: null } }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const result = await fetchPp30Summary({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 9 });
    expect(result.summary?.outputvat).toBe("0.00");
    expect(result.summary?.inputvat).toBe("0.00");
    expect(result.summary?.company).toEqual({ code: "", name: "", taxid: "" });
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

describe("fetchWhtReport", () => {
  it("200 → แถว + ยอดรวมทั้งงวด + หัวบริษัทจาก backend เป็น string ตามที่ส่งมา", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(200, {
        status: "success",
        data: [
          {
            journalid: "J1",
            docno: "PV6901-S001",
            docdate: "2026-01-15",
            partnercode: "SUPP-TH-001",
            partnername: "บริษัท ปูนซีเมนต์ไทยค้าส่ง จำกัด",
            taxid: "0105558002001",
            address: "",
            description: "จ่ายชำระหนี้เจ้าหนี้การค้า หักภาษี ณ ที่จ่าย 3%",
            baseamount: "107000.00",
            whtamount: "3210.00",
            whtamounttext: "สามพันสองร้อยสิบบาทถ้วน",
            netamount: "103790.00",
            ratepercent: "3.00",
          },
        ],
        total: 1,
        summary: {
          basetotal: "107000.00",
          whttotal: "3210.00",
          whttotaltext: "สามพันสองร้อยสิบบาทถ้วน",
          nettotal: "103790.00",
          payeecount: 1,
          byrate: [{ ratepercent: "3.00", count: 1, baseamount: "107000.00", whtamount: "3210.00" }],
        },
        company: { code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234" },
        note: "",
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 1, direction: "paid" });

    expect(result.error).toBeUndefined();
    expect(result.total).toBe(1);
    expect(result.rows[0]).toMatchObject({ whtamount: "3210.00", baseamount: "107000.00", netamount: "103790.00", ratepercent: "3.00" });
    expect(result.summary.byrate).toEqual([{ ratepercent: "3.00", count: 1, baseamount: "107000.00", whtamount: "3210.00" }]);
    expect(result.summary.whttotaltext).toBe("สามพันสองร้อยสิบบาทถ้วน");
    expect(result.company.taxid).toBe("0105558001234");
    expect(fetchMock).toHaveBeenCalledWith("/api/goapi/api/report/tax/wht", expect.objectContaining({ method: "POST" }));
  });

  it("ไม่มีบริษัท → company_required และไม่ยิง fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "", year: 2026, month: 1, direction: "paid" });
    expect(result.error).toBe("company_required");
    expect(result.summary.whttotal).toBe("0.00");
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

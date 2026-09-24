import { afterEach, describe, expect, it, vi } from "vitest";
import {
  THAI_TAX_CONFIGS,
  countTaxIdIssues,
  fetchVatRegister,
  fetchWhtReport,
  getThaiTaxConfig,
  hasThirteenDigitTaxId,
  isHeadOfficeBranch,
  isThaiTaxRoute,
  requestWhtCertificatePdf,
  taxAddressProblems,
  taxCompanyLabel,
  toTaxAddress,
  trimTaxAddress,
  whtCertificateForm,
  whtRegisterRecords,
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
  it("ให้ config ครบ 5 รายการ และ resolve ได้จาก route", () => {
    expect(THAI_TAX_CONFIGS).toHaveLength(5);
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
    expect(result.summary).toEqual({ amountbeforevat: "30000.10", vatamount: "2100.01", totalamount: "32100.11", duplicatecount: 0 });
    // backend ไม่ส่งรายการซ้ำ = ไม่ซ้ำ (อาร์เรย์ว่าง ไม่ใช่ undefined)
    expect(result.records[0]?.duplicatedocnos).toEqual([]);
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

  it("ยอดสุทธิว่าง (backend ยังไม่รู้ฐาน) คงว่าง ไม่กลายเป็น 0.00 — UAT S14/S25 2026-09-24", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      status: "success",
      data: [{ journalid: "J9", docno: "PV6912-W06", baseamount: "0.00", whtamount: "200.00", netamount: "", ratepercent: "", taxbasesource: "inferred" }],
      total: 1,
      summary: { basetotal: "0.00", whttotal: "200.00", nettotal: "0.00", payeecount: 1, byrate: [] },
    })));
    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 12, direction: "paid" });
    expect(result.rows[0]).toMatchObject({ baseamount: "0.00", whtamount: "200.00", netamount: "" });
    expect(whtRegisterRecords(result.rows)[0].totalamount).toBe("");
  });

  // แถวประมาณ (ไม่มีรายละเอียดภาษีหักในใบสำคัญ): backend อ่านแบบยื่นจากชื่อบัญชีภาษีหัก → 50 ทวิ ใช้ค่านั้นเป็นค่าเริ่มต้น
  it("แถวประมาณที่ backend เติม formtype จากชื่อบัญชี → ช่องแบบ 50 ทวิ ตามนั้น; ระบุไม่ได้ = ว่าง", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      data: [
        { journalid: "J7", docno: "PV6912-W07", baseamount: "10000.00", whtamount: "300.00", netamount: "9700.00", ratepercent: "3.00", taxbasesource: "inferred", formtype: "PND3" },
        { journalid: "J8", docno: "PV6912-W08", baseamount: "0.00", whtamount: "150.00", netamount: "", ratepercent: "", taxbasesource: "inferred", formtype: "" },
      ],
      total: 2,
      summary: {},
    })));
    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 12, direction: "paid" });
    expect(result.rows.map((r) => [r.taxbasesource, r.formtype, whtCertificateForm(r.formtype)])).toEqual([["inferred", "PND3", "3"], ["inferred", "", ""]]);
  });

  it("ไม่มีบริษัท → company_required และไม่ยิง fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "", year: 2026, month: 1, direction: "paid" });
    expect(result.error).toBe("company_required");
    expect(result.summary.whttotal).toBe("0.00");
    expect(fetchMock).not.toHaveBeenCalled();
  });

  // UAT 2026-09-24: ผู้ใช้เลือกไทยในแอปแต่เบราว์เซอร์เป็นอังกฤษ → ประเภทเงินได้ขึ้น "Section 3 Tres"
  it("ส่งภาษาที่เลือกในแอปเป็น Accept-Language และไม่ส่ง language ไปใน body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { data: [], total: 0, summary: {} }));
    vi.stubGlobal("fetch", fetchMock);
    await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 10, direction: "paid", language: "th" });
    await fetchVatRegister({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 10, type: "purchase", language: "th" });
    for (const [, init] of fetchMock.mock.calls as [string, RequestInit][]) {
      expect(new Headers(init.headers).get("accept-language")).toBe("th");
      expect(JSON.parse(String(init.body))).not.toHaveProperty("language");
    }
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});

describe("แถวภาษีหัก ณ ที่จ่ายหลายรายการในใบสำคัญเดียว", () => {
  const whtItem = (overrides: Record<string, unknown>) => ({
    journalid: "J1",
    docno: "PV6901-S001",
    docdate: "2026-01-15",
    partnername: "บริษัท ปูนซีเมนต์ไทยค้าส่ง จำกัด",
    taxid: "0105558002001",
    baseamount: "100000.00",
    whtamount: "3000.00",
    netamount: "97000.00",
    ratepercent: "3.00",
    taxbasesource: "recorded",
    ...overrides,
  });

  it("rowid ไม่ซ้ำต่อรายการ และคลิกแถวแล้ว map กลับได้รายการที่ถูกต้อง (ไม่ใช่รายการแรกของใบ)", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      status: "success",
      data: [
        whtItem({ formtype: "PND53" }),
        whtItem({ formtype: "PND3", baseamount: "20000.00", whtamount: "600.00", netamount: "19400.00" }),
        whtItem({ journalid: "J2", formtype: "PND53" }),
      ],
      total: 3,
      summary: {},
      company: {},
    })));

    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 1, direction: "paid", offset: 10 });

    const ids = result.rows.map((row) => row.rowid);
    expect(new Set(ids).size).toBe(3);
    expect(ids).toEqual(["wht-J1#10", "wht-J1#11", "wht-J2#12"]);
    expect(result.rows.map((row) => row.formtype)).toEqual(["PND53", "PND3", "PND53"]);

    const records = whtRegisterRecords(result.rows);
    expect(records.map((record) => record.id)).toEqual(ids);
    const clicked = records[1]!;
    const matched = result.rows.find((row) => row.rowid === clicked.id);
    expect(matched).toMatchObject({ baseamount: "20000.00", whtamount: "600.00", formtype: "PND3" });
    // แถวที่ backend ไม่ส่งสาขาคู่ค้า → ไม่ติดป้ายสำนักงานใหญ่
    expect(records.every((record) => record.branchno === "" && !record.isheadoffice)).toBe(true);
    // UAT 2026-09-24: สาขาในรายละเอียดรายงานว่างทั้งที่คู่ค้าเป็นสำนักงานใหญ่ — ใช้เลขสาขาที่ backend ส่งมา
    const withBranch = whtRegisterRecords([{ ...result.rows[0]!, branchno: "00000" }])[0]!;
    expect(withBranch).toMatchObject({ branchno: "00000", isheadoffice: true });
  });

  it("หมายเหตุของแถว (กลับรายการภายหลัง) จาก backend แสดงเป็น remark ของรายการ — แถวปกติไม่มี remark", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      status: "success",
      data: [whtItem({ reversedmonth: "2026-02", note: "กลับรายการภายหลังในเดือน กุมภาพันธ์ 2569" }), whtItem({ journalid: "J2" })],
      total: 2,
      summary: {},
      company: {},
    })));

    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 1, direction: "paid" });
    const records = whtRegisterRecords(result.rows);
    expect(records.map((record) => record.remark)).toEqual(["กลับรายการภายหลังในเดือน กุมภาพันธ์ 2569", undefined]);
  });
});

describe("whtCertificateForm — แบบยื่นที่บันทึก → ช่องลำดับที่ในแบบของ 50 ทวิ", () => {
  it("PND3 → 3, PND53 → 53, PND2 → 2 (ไม่สนตัวพิมพ์/ช่องว่าง)", () => {
    expect(whtCertificateForm("PND3")).toBe("3");
    expect(whtCertificateForm("PND53")).toBe("53");
    expect(whtCertificateForm("PND2")).toBe("2");
    expect(whtCertificateForm(" pnd53 ")).toBe("53");
  });

  it("ไม่รู้จัก/ว่าง → '' ให้จอใช้ค่าเริ่มต้นเดิม", () => {
    expect(whtCertificateForm("")).toBe("");
    expect(whtCertificateForm("PND1")).toBe("");
  });
});

// ใบกำกับเลขที่เดียวกันจากผู้ขายต่างรายเป็นคนละฉบับ (ม.86/4: ใบกำกับระบุผู้ออก + เลขที่) — backend ตัดสินว่าซ้ำจริง จอแค่แสดง
describe("fetchVatRegister — ใบกำกับฉบับเดียวกันถูกบันทึกในใบสำคัญอื่น", () => {
  it("map duplicatedocnos (ทิ้งค่าที่ไม่ใช่ข้อความ/ว่าง) และ duplicatecount ทั้งงวดจาก summary", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      status: "success",
      data: [
        { docdate: "2026-09-03", taxinvoiceno: "IV-0001", counterpartyname: "บริษัท ปูนซีเมนต์ไทยค้าส่ง จำกัด", taxid: "0105558002001", branchno: "00000",
          amountbeforevat: "10000.00", vatamount: "700.00", totalamount: "10700.00", duplicatedocnos: ["PV6909-0012", " JV6909-0003 ", "", 7, null] },
        { docdate: "2026-09-04", taxinvoiceno: "IV-0001", counterpartyname: "ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย", taxid: "0103560001234", branchno: "00000",
          amountbeforevat: "500.00", vatamount: "35.00", totalamount: "535.00", duplicatedocnos: [] },
        { docdate: "2026-09-05", taxinvoiceno: "IV-0002", counterpartyname: "บริษัท สยามพาณิชย์ โฮลดิ้ง จำกัด", taxid: "0105561009999", branchno: "00001",
          amountbeforevat: "100.00", vatamount: "7.00", totalamount: "107.00", duplicatedocnos: "PV6909-0012" },
      ],
      total: 3,
      summary: { amountbeforevat: "10600.00", vatamount: "742.00", totalamount: "11342.00", duplicatecount: 4 },
    })));
    const result = await fetchVatRegister({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 9, type: "purchase" });
    expect(result.records.map((r) => r.duplicatedocnos)).toEqual([["PV6909-0012", "JV6909-0003"], [], []]);
    // ยอดนับทั้งงวดจาก backend ไม่ใช่นับแถวที่โหลด
    expect(result.summary.duplicatecount).toBe(4);
  });

  it("duplicatecount ผิดสัญญา (ติดลบ/ทศนิยม/ข้อความ) → 0 ไม่เดา", async () => {
    for (const bad of [-1, 1.5, "2"]) {
      vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, { data: [], total: 0, summary: { duplicatecount: bad } })));
      const result = await fetchVatRegister({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 9, type: "sale" });
      expect(result.summary.duplicatecount).toBe(0);
    }
  });
});

describe("taxCompanyLabel — ป้ายบริษัทบนหัวรายงานภาษี", () => {
  const workspace = JSON.stringify({ shop: { holdingcode: "rungrueng" }, company: { code: "01", names: [{ code: "th", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" }] } });
  it("หัวบริษัทจาก backend มาก่อน", () => {
    expect(taxCompanyLabel({ code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234" }, "", "rungrueng", "01")).toBe("[01] บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด");
  });
  it("ทะเบียน VAT ไม่มีหัวบริษัท → ใช้บริษัทที่เลือกใน workspace เมื่อรหัสตรงกันเท่านั้น", () => {
    expect(taxCompanyLabel(null, workspace, "rungrueng", "01")).toBe("[01] บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด");
    // คนละบริษัท/คนละกลุ่มกิจการ → ห้ามโชว์ชื่อบริษัทอื่น แสดงแค่รหัส
    expect(taxCompanyLabel(null, workspace, "rungrueng", "02")).toBe("[02]");
    expect(taxCompanyLabel(null, workspace, "other", "01")).toBe("[01]");
  });
  it("workspace เสีย/ว่าง หรือไม่มีรหัสบริษัท", () => {
    expect(taxCompanyLabel(null, "{not json", "rungrueng", "01")).toBe("[01]");
    expect(taxCompanyLabel({ code: "", name: "", taxid: "" }, "", "rungrueng", "")).toBe("");
  });
});

describe("isHeadOfficeBranch — สาขาว่างไม่ใช่สำนักงานใหญ่", () => {
  it("เฉพาะ 00000 เท่านั้น", () => {
    expect(isHeadOfficeBranch("00000")).toBe(true);
    expect(isHeadOfficeBranch("")).toBe(false);
    expect(isHeadOfficeBranch("00001")).toBe(false);
  });

  it("ทะเบียนภาษีขายที่ไม่ได้บันทึกสาขา → isheadoffice false", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      status: "success",
      data: [{ docdate: "2026-09-01", taxinvoiceno: "INV-002", counterpartyname: "ลูกค้าเงินสด", taxid: "", branchno: "", amountbeforevat: "100.00", vatamount: "7.00", totalamount: "107.00" }],
      total: 1,
      summary: {},
    })));
    const result = await fetchVatRegister({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 9, type: "sale" });
    expect(result.records[0]?.branchno).toBe("");
    expect(result.records[0]?.isheadoffice).toBe(false);
  });
});

describe("ตรวจเลขประจำตัวผู้เสียภาษี 13 หลัก", () => {
  it("นับเฉพาะตัวเลข เหมือน backend (ขีดคั่นไม่ผิด, ว่าง/สั้นผิด)", () => {
    expect(hasThirteenDigitTaxId("0105558002001")).toBe(true);
    expect(hasThirteenDigitTaxId("0-1055-58002-00-1")).toBe(true);
    expect(hasThirteenDigitTaxId("")).toBe(false);
    expect(hasThirteenDigitTaxId("12345")).toBe(false);
    expect(countTaxIdIssues([{ taxid: "0105558002001" }, { taxid: "" }, { taxid: "010555800200" }])).toBe(2);
  });
});

// review 2026-09-24: 403 ของ backend มี code ที่บอกทางแก้ — ต้องถึงจอ ไม่ถูกกลืนเป็น "unauthorized" (จอขึ้น "ลองใหม่")
describe("requestWhtCertificatePdf 403", () => {
  afterEach(() => vi.unstubAllGlobals());
  const respond = (status: number, body: unknown) =>
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(body), { status, headers: { "content-type": "application/json" } })));
  const certificate = {} as Parameters<typeof requestWhtCertificatePdf>[2];

  it("passes the backend code through (scope denied / access expired)", async () => {
    respond(403, { success: false, code: "user_access_expired", field: "", message: "หมดอายุ" });
    expect(await requestWhtCertificatePdf("h", "01", certificate)).toMatchObject({ ok: false, error: "user_access_expired" });
    respond(403, { success: false, code: "wht_cert_scope_denied", field: "journalid", message: "ไม่มีสิทธิ์" });
    expect(await requestWhtCertificatePdf("h", "01", certificate)).toMatchObject({ ok: false, error: "wht_cert_scope_denied", field: "journalid" });
  });

  // 401 ผ่านขั้น refresh session ของ authFetch ก่อน จึงไม่ทดสอบที่นี่
  it("keeps unauthorized for a 403 without a code (BFF)", async () => {
    respond(403, { success: false, message: "forbidden" });
    expect(await requestWhtCertificatePdf("h", "01", certificate)).toMatchObject({ ok: false, error: "unauthorized" });
  });
});

describe("ชื่อผู้ถูกหักภาษีพร้อมคำนำหน้าในทะเบียนภาษีหัก ณ ที่จ่าย", () => {
  it("ใช้ partnerfullname จาก backend แสดงบนจอ/CSV/พิมพ์ — partnername ไม่มีคำนำหน้ายังคงเดิม; backend เก่า/ค่าว่าง → ใช้ partnername", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(200, {
      status: "success",
      data: [
        { journalid: "J1", partnername: "วิไลวรรณ ศรีสุข", title: "นางสาว", partnerfullname: "นางสาว วิไลวรรณ ศรีสุข", whtamount: "300.00" },
        { journalid: "J2", partnername: "บริษัท ปูนกรุงไทย จำกัด", partnerfullname: null, whtamount: "3210.00" },
        { journalid: "J3", whtamount: "100.00" },
      ],
      total: 3,
      summary: {},
      company: {},
    })));

    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "B001", year: 2026, month: 1, direction: "paid" });

    expect(result.rows.map((row) => [row.partnername, row.partnerfullname])).toEqual([
      ["วิไลวรรณ ศรีสุข", "นางสาว วิไลวรรณ ศรีสุข"],
      ["บริษัท ปูนกรุงไทย จำกัด", "บริษัท ปูนกรุงไทย จำกัด"],
      ["", ""],
    ]);
    expect(whtRegisterRecords(result.rows).map((record) => record.counterpartyname)).toEqual([
      "นางสาว วิไลวรรณ ศรีสุข",
      "บริษัท ปูนกรุงไทย จำกัด",
      "",
    ]);
  });
});

describe("ที่อยู่สำหรับภาษีของบริษัท (ทะเบียนบริษัท → หัวแบบ/50 ทวิ)", () => {
  it("หัวบริษัทของรายงานภาษีหัก ณ ที่จ่ายอ่าน address/phone/addressline แบบกัน null", async () => {
    const payload = (company: unknown) => jsonResponse(200, { status: "success", data: [], total: 0, summary: {}, company, note: "" });
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(payload({
      code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234",
      address: { addr_no: "88/12", addr_province: "ปทุมธานี", addr_postcode: 12140, addr_road: null },
      phone: " 02-123-4567 ",
      addressline: "เลขที่ 88/12 จังหวัดปทุมธานี 12140",
    })));
    const result = await fetchWhtReport({ holdingcode: "H001", businesscode: "01", year: 2026, month: 1, direction: "paid" });
    expect(result.company.address?.addr_no).toBe("88/12");
    expect(result.company.address?.addr_road).toBe("");
    expect(result.company.address?.addr_postcode).toBe(""); // ตัวเลข JSON ไม่ใช่ข้อความ → ไม่เดา
    expect(result.company.phone).toBe("02-123-4567");
    expect(result.company.addressline).toBe("เลขที่ 88/12 จังหวัดปทุมธานี 12140");

    // backend เก่า / ทะเบียนยังไม่มีที่อยู่ → null + ค่าว่าง ไม่พัง
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(payload({ code: "01", name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", taxid: "0105558001234", address: null })));
    const blank = await fetchWhtReport({ holdingcode: "H001", businesscode: "01", year: 2026, month: 1, direction: "paid" });
    expect(blank.company).toMatchObject({ address: null, phone: "", addressline: "" });
  });

  it("toTaxAddress ครบ 13 ช่องเสมอ + trimTaxAddress ตัดช่องว่างหัวท้าย", () => {
    const empty = toTaxAddress(undefined);
    expect(Object.keys(empty)).toHaveLength(13);
    expect(Object.values(empty).every((v) => v === "")).toBe(true);
    expect(trimTaxAddress({ ...empty, addr_no: " 88/12 ", addr_province: "  " })).toMatchObject({ addr_no: "88/12", addr_province: "" });
  });

  it("taxAddressProblems ตรวจเหมือน backend: รหัสไปรษณีย์ 5 หลัก, ขึ้นบรรทัด/แท็บ, ความยาวนับตัวอักษร", () => {
    const base = toTaxAddress({});
    expect(taxAddressProblems(base, "")).toEqual({});
    expect(taxAddressProblems({ ...base, addr_postcode: "12120" }, "")).toEqual({});
    expect(taxAddressProblems({ ...base, addr_postcode: "1212" }, "").addr_postcode).toBe("postcode");
    expect(taxAddressProblems({ ...base, addr_postcode: "๑๒๑๒๐" }, "").addr_postcode).toBe("postcode");
    expect(taxAddressProblems({ ...base, addr_road: "พหลโยธิน\nกม. 40" }, "").addr_road).toBe("control_char");
    expect(taxAddressProblems({ ...base, addr_soi: "ลาดพร้าว\t71" }, "").addr_soi).toBe("control_char");
    expect(taxAddressProblems({ ...base, addr_building: "ก".repeat(40) }, "")).toEqual({});
    expect(taxAddressProblems({ ...base, addr_building: "ก".repeat(41) }, "").addr_building).toBe("too_long");
    expect(taxAddressProblems(base, "0".repeat(51)).phone).toBe("too_long");
    expect(taxAddressProblems({ ...base, addr_no: " 88/12 " }, " 02-123-4567 ")).toEqual({});
  });
});

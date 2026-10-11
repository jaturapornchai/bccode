import { describe, expect, it } from "vitest";
import {
  isPrepareAmountText,
  payLinesFromPrepare,
  prepareFormFromDoc,
  preparePayload,
  prepareStatus,
  prepareTotals,
  type ApPrepareDoc,
  type ApPrepareLine,
} from "./ap-payment-prepare";
import { erpDecimalSign, erpSignedSum, type ErpDetailItem } from "./erp-transaction";

function line(patch: Partial<ApPrepareLine>): ApPrepareLine {
  return {
    linenumber: 1,
    debt_document_id: "AP-RR6910-0001",
    document_no: "RR6910-0001",
    document_date: "2026-10-01",
    debt_amount: "10700.00",
    billtype: 1,
    billtype_key: "goods_receive",
    sign: 1,
    description: "ซื้อกระดาษ A4 บจก.ศรีสยามเปเปอร์",
    invbalance: "10700.00",
    payamount: "10700.00",
    used_amount: "0.00",
    open_amount: "10700.00",
    ...patch,
  };
}

function doc(lines: ApPrepareLine[], patch: Partial<ApPrepareDoc> = {}): ApPrepareDoc {
  return {
    docno: "PP6910-0001",
    paydate: "2026-10-11",
    partner_code: "V-0012",
    partner_name: "บจก.ศรีสยามเปเปอร์",
    branch_code: "00000",
    remark: "",
    projectcode: "",
    allocatecode: "",
    billgroupcode: "",
    departmentcode: "",
    total_amount: "10700.00",
    total_home_amount: "10700.00",
    discount_amount: "0.00",
    net_amount: "10700.00",
    exchange_profit: "0.00",
    billstatus: 0,
    iscancel: false,
    cancel_reason: "",
    version: 2,
    lines,
    ...patch,
  };
}

describe("ap payment prepare helpers", () => {
  it("sums signed amounts exactly (0.1 + 0.2, credit notes subtract)", () => {
    expect(erpSignedSum([{ amount: "0.1", sign: 1 }, { amount: "0.2", sign: 1 }])).toBe("0.3");
    const totals = prepareTotals(
      [
        { payamount: "10700.00", sign: 1 },
        { payamount: "1070.00", sign: -1 },
        { payamount: "0.01", sign: 1 },
      ],
      "100.50",
    );
    expect(erpDecimalSign(totals.total)).toBe(1);
    expect(erpSignedSum([{ amount: totals.total, sign: 1 }, { amount: "9630.01", sign: -1 }])).toBe("0");
    expect(erpSignedSum([{ amount: totals.net, sign: 1 }, { amount: "9529.51", sign: -1 }])).toBe("0");
  });

  it("reports the sign of a decimal string without Number()", () => {
    expect(erpDecimalSign("9999999999999999.99")).toBe(1);
    expect(erpDecimalSign("-0.01")).toBe(-1);
    expect(erpDecimalSign("0.00")).toBe(0);
    expect(erpDecimalSign("")).toBe(0);
  });

  it.each([
    ["10700", true],
    ["10700.5", true],
    ["0.01", true],
    [" 25.00 ", true],
    ["9999999999999999.99", true],
    ["1.234", false],
    ["-5", false],
    ["1,000", false],
    ["", false],
    ["abc", false],
  ])("amount text %j valid=%s", (value, ok) => {
    expect(isPrepareAmountText(value)).toBe(ok);
  });

  it("derives status from cancel flag first, then bill status", () => {
    expect(prepareStatus({ iscancel: true, billstatus: 1 })).toBe("cancelled");
    expect(prepareStatus({ iscancel: false, billstatus: 1 })).toBe("paid");
    expect(prepareStatus({ iscancel: false, billstatus: 0 })).toBe("unpaid");
  });

  it("builds a trimmed payload without allocation/bill group fields", () => {
    const form = prepareFormFromDoc(doc([line({ payamount: "500.00", description: "งวดแรก" })], { departmentcode: " ACC " }));
    form.lines[0].payamount = " 500.00 ";
    const payload = preparePayload(form);
    expect(payload).toEqual({
      docno: "PP6910-0001",
      paydate: "2026-10-11",
      partner_code: "V-0012",
      remark: "",
      projectcode: "",
      departmentcode: "ACC",
      discount_amount: "",
      version: 2,
      lines: [{ debt_document_id: "AP-RR6910-0001", payamount: "500.00", description: "งวดแรก" }],
    });
    expect(payload).not.toHaveProperty("allocatecode");
  });

  it("pulls only open positive lines into payment lines and skips debts already pulled", () => {
    const prepare = doc([
      line({}),
      line({ debt_document_id: "AP-RR6910-0002", document_no: "RR6910-0002", open_amount: "0.00", used_amount: "5350.00" }),
      line({ debt_document_id: "AP-CN6910-0001", document_no: "CN6910-0001", sign: -1, billtype: 2, open_amount: "1070.00" }),
      line({ debt_document_id: "AP-RR6910-0003", document_no: "RR6910-0003", description: "", open_amount: "2140.50" }),
    ]);
    const existing: ErpDetailItem[] = [
      { linenumber: 1, itemcode: "RR6910-0001", itemname: "x", unitcode: "", qty: "1", price: "10700.00", sumamount: "10700.00", debt_document_id: "AP-RR6910-0001", paybillno: "PP6910-0001" },
    ];
    const pulled = payLinesFromPrepare(prepare, existing);
    expect(pulled).toEqual([
      {
        linenumber: 2,
        itemcode: "RR6910-0003",
        itemname: "RR6910-0003",
        unitcode: "",
        qty: "1",
        price: "2140.50",
        sumamount: "2140.50",
        docref: "RR6910-0003",
        debt_document_id: "AP-RR6910-0003",
        paybillno: "PP6910-0001",
      },
    ]);
    // หนี้เดียวกันจากใบเตรียมจ่ายใบอื่นไม่ถือว่าดึงแล้ว
    expect(payLinesFromPrepare(prepare, [{ ...existing[0], paybillno: "PP6910-0009" }]).map((item) => item.debt_document_id)).toEqual([
      "AP-RR6910-0001",
      "AP-RR6910-0003",
    ]);
    expect(payLinesFromPrepare(doc([], { lines: null }), [])).toEqual([]);
  });
});

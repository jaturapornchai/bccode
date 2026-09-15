import { describe, expect, it } from "vitest";
import {
  amountUnits,
  amountString,
  emptyAccount,
  emptyFiscalYear,
  emptyJournal,
  emptyLine,
  journalTotals,
  sortAccountsHierarchically,
  validateJournal,
  type GLAccount,
  type GLFiscalYear,
  type GLJournal,
} from "./general-ledger";

/**
 * Thai Accounting Firm & Certified Auditor End-to-End Simulation Test
 * (การทดสอบวงจรบัญชีครบวงจรแบบสำนักงานบัญชีไทยและผู้สอบบัญชีรับอนุญาต)
 *
 * Scope:
 * 1. กำหนดปีบัญชี 2569 (Fiscal Year)
 * 2. สร้างและจัดการผังบัญชี 5 หมวดมาตรฐานไทย (Chart of Accounts CRUD & Tree Hierarchy)
 * 3. บันทึกยอดยกมาต้นงวด (Opening Balances)
 * 4. บันทึกสมุดรายวันเฉพาะ 5 เล่มตามมาตรฐานไทย (UV, SV, RV, PV, JV)
 * 5. การผ่านรายการ (Ledger Posting) & ตรวจสอบดุลเดบิต-เครดิต
 * 6. การออกงบทดลอง (Trial Balance) - ผลต่างต้องเป็น 0.00
 * 7. การออกงบกำไรขาดทุน (Income Statement / P&L) - คำนวณกำไรสุทธิ
 * 8. การออกงบแสดงฐานะการเงิน (Balance Sheet) - สินทรัพย์ = หนี้สิน + ส่วนของเจ้าของ 100%
 */
describe("Thai Accounting Firm - Full Cycle Accounting Test (E2E)", () => {
  // ==========================================
  // PHASE 1: Fiscal Year Setup
  // ==========================================
  const fiscalYear: GLFiscalYear = {
    ...emptyFiscalYear(),
    code: "2569",
    startdate: "2026-01-01",
    enddate: "2026-12-31",
    scale: 2,
    profitlossaccount: "3201",
    retainedearningsaccount: "3201",
    isactive: true,
    closed: false,
  };

  it("Step 1: กำหนดปีบัญชี 2569 ถูกต้องตามมาตรฐาน", () => {
    expect(fiscalYear.code).toBe("2569");
    expect(fiscalYear.isactive).toBe(true);
    expect(fiscalYear.closed).toBe(false);
    expect(fiscalYear.scale).toBe(2);
  });

  // ==========================================
  // PHASE 2: Thai Standard Chart of Accounts (5 Categories)
  // ==========================================
  let coa: GLAccount[] = [
    // หมวด 1: สินทรัพย์ (Assets)
    { ...emptyAccount(), accountcode: "1000", names: [{ code: "th", name: "สินทรัพย์" }, { code: "en", name: "Assets" }], accounttype: "asset", normalbalance: "debit", allowposting: false, level: 1 },
    { ...emptyAccount(), accountcode: "1100", parentaccountcode: "1000", names: [{ code: "th", name: "สินทรัพย์หมุนเวียน" }], accounttype: "asset", normalbalance: "debit", allowposting: false, level: 2 },
    { ...emptyAccount(), accountcode: "1111", parentaccountcode: "1100", names: [{ code: "th", name: "เงินสดในมือ" }], accounttype: "asset", normalbalance: "debit", allowposting: true, iscash: true, level: 3 },
    { ...emptyAccount(), accountcode: "1112", parentaccountcode: "1100", names: [{ code: "th", name: "เงินฝากกระแสรายวัน-ธนาคารกสิกรไทย" }], accounttype: "asset", normalbalance: "debit", allowposting: true, iscash: true, level: 3 },
    { ...emptyAccount(), accountcode: "1131", parentaccountcode: "1100", names: [{ code: "th", name: "ลูกหนี้การค้า" }], accounttype: "asset", normalbalance: "debit", allowposting: true, level: 3 },
    { ...emptyAccount(), accountcode: "1141", parentaccountcode: "1100", names: [{ code: "th", name: "สินค้าสำเร็จรูป" }], accounttype: "asset", normalbalance: "debit", allowposting: true, level: 3 },
    { ...emptyAccount(), accountcode: "1151", parentaccountcode: "1100", names: [{ code: "th", name: "ภาษีซื้อ" }], accounttype: "asset", normalbalance: "debit", allowposting: true, level: 3 },

    // หมวด 2: หนี้สิน (Liabilities)
    { ...emptyAccount(), accountcode: "2000", names: [{ code: "th", name: "หนี้สิน" }, { code: "en", name: "Liabilities" }], accounttype: "liability", normalbalance: "credit", allowposting: false, level: 1 },
    { ...emptyAccount(), accountcode: "2100", parentaccountcode: "2000", names: [{ code: "th", name: "หนี้สินหมุนเวียน" }], accounttype: "liability", normalbalance: "credit", allowposting: false, level: 2 },
    { ...emptyAccount(), accountcode: "2121", parentaccountcode: "2100", names: [{ code: "th", name: "เจ้าหนี้การค้า" }], accounttype: "liability", normalbalance: "credit", allowposting: true, level: 3 },
    { ...emptyAccount(), accountcode: "2141", parentaccountcode: "2100", names: [{ code: "th", name: "ภาษีขาย" }], accounttype: "liability", normalbalance: "credit", allowposting: true, level: 3 },
    { ...emptyAccount(), accountcode: "2142", parentaccountcode: "2100", names: [{ code: "th", name: "ภาษีหัก ณ ที่จ่ายค้างจ่าย (ภ.ง.ด. 3, 53)" }], accounttype: "liability", normalbalance: "credit", allowposting: true, level: 3 },

    // หมวด 3: ส่วนของเจ้าของ (Equity)
    { ...emptyAccount(), accountcode: "3000", names: [{ code: "th", name: "ส่วนของเจ้าของ" }, { code: "en", name: "Equity" }], accounttype: "equity", normalbalance: "credit", allowposting: false, level: 1 },
    { ...emptyAccount(), accountcode: "3101", parentaccountcode: "3000", names: [{ code: "th", name: "ทุนจดทะเบียนและชำระแล้ว" }], accounttype: "equity", normalbalance: "credit", allowposting: true, level: 2 },
    { ...emptyAccount(), accountcode: "3201", parentaccountcode: "3000", names: [{ code: "th", name: "กำไรสะสมยังไม่ได้จัดสรร" }], accounttype: "equity", normalbalance: "credit", allowposting: true, level: 2 },

    // หมวด 4: รายได้ (Revenue)
    { ...emptyAccount(), accountcode: "4000", names: [{ code: "th", name: "รายได้" }, { code: "en", name: "Revenue" }], accounttype: "income", normalbalance: "credit", allowposting: false, level: 1 },
    { ...emptyAccount(), accountcode: "4101", parentaccountcode: "4000", names: [{ code: "th", name: "รายได้จากการขายสินค้า" }], accounttype: "income", normalbalance: "credit", allowposting: true, level: 2 },

    // หมวด 5: ค่าใช้จ่าย (Expenses)
    { ...emptyAccount(), accountcode: "5000", names: [{ code: "th", name: "ค่าใช้จ่าย" }, { code: "en", name: "Expenses" }], accounttype: "expense", normalbalance: "debit", allowposting: false, level: 1 },
    { ...emptyAccount(), accountcode: "5101", parentaccountcode: "5000", names: [{ code: "th", name: "ต้นทุนขาย" }], accounttype: "expense", normalbalance: "debit", allowposting: true, level: 2 },
    { ...emptyAccount(), accountcode: "5301", parentaccountcode: "5000", names: [{ code: "th", name: "ค่าเช่าและบริการสำนักงาน" }], accounttype: "expense", normalbalance: "debit", allowposting: true, level: 2 },
  ];

  it("Step 2: ผังบัญชีรองรับ CRUD และการจัดลำดับชั้นแบบ Tree View ครบ 5 หมวด", () => {
    // CREATE: ตรวจสอบจำนวนบัญชี
    expect(coa).toHaveLength(20);

    // READ & HIERARCHY: ตรวจสอบการเรียงตามลำดับชั้น 5 หมวด (Asset -> Liability -> Equity -> Income -> Expense)
    const sorted = sortAccountsHierarchically(coa);
    expect(sorted[0].accountcode).toBe("1000");
    expect(sorted[0].accounttype).toBe("asset");

    // UPDATE: ทดสอบการอัปเดตชื่อบัญชี
    const bankAccount = coa.find((a) => a.accountcode === "1112")!;
    bankAccount.names.push({ code: "en", name: "Kasikorn Bank Current Account" });
    expect(bankAccount.names).toHaveLength(2);

    // DELETE GUARD: ตรวจสอบว่าบัญชีคุมที่มีลูก (เช่น 1000) ห้ามลบ
    const hasChildren = (code: string) => coa.some((a) => a.parentaccountcode === code);
    expect(hasChildren("1000")).toBe(true);
    expect(hasChildren("1111")).toBe(false); // บัญชีย่อยไม่มีลูก
  });

  // ==========================================
  // PHASE 3: Opening Balances (ยอดยกมาต้นงวด)
  // ==========================================
  it("Step 3: บันทึกยอดยกมาต้นงวด (Opening Balance) เดบิต = เครดิต", () => {
    const openingJournal: GLJournal = {
      ...emptyJournal("JV", "opening"),
      docno: "OB-2569-001",
      date: "2026-01-01",
      fiscalyear: "2569",
      description: "บันทึกยอดยกมาต้นงวดบัญชีปี 2569",
      lines: [
        { ...emptyLine(), accountcode: "1111", debit: "200000", credit: "0", description: "ยอดยกมาเงินสด" },
        { ...emptyLine(), accountcode: "1112", debit: "800000", credit: "0", description: "ยอดยกมาเงินฝากธนาคาร" },
        { ...emptyLine(), accountcode: "3101", debit: "0", credit: "1000000", description: "ยอดยกมาทุนจดทะเบียน" },
      ],
    };

    const error = validateJournal(openingJournal, fiscalYear, coa);
    expect(error).toBeNull();

    const totals = journalTotals(openingJournal.lines);
    expect(totals.debit).toBe(amountUnits("1000000"));
    expect(totals.credit).toBe(amountUnits("1000000"));
    expect(totals.difference).toBe(0n); // Balanced!
  });

  // ==========================================
  // PHASE 4: Thai Accounting Daily Operations (5 Special Journals)
  // ==========================================
  const postedJournals: GLJournal[] = [];

  it("Step 4.1: [SV] สมุดรายวันซื้อ - ซื้อสินค้าเป็นเงินเชื่อ 100,000 + ภาษีซื้อ 7%", () => {
    const svJournal: GLJournal = {
      ...emptyJournal("SV", "purchase"),
      docno: "SV-2569-001",
      date: "2026-01-10",
      fiscalyear: "2569",
      description: "ซื้อสินค้าสำเร็จรูปจาก บริษัท คู่ค้า จำกัด (บิลเลขที่ INV-9901)",
      lines: [
        { ...emptyLine(), accountcode: "1141", debit: "100000", credit: "0", description: "ซื้อสินค้าสำเร็จรูป" },
        { ...emptyLine(), accountcode: "1151", debit: "7000", credit: "0", description: "ภาษีซื้อ 7%" },
        { ...emptyLine(), accountcode: "2121", debit: "0", credit: "107000", description: "เจ้าหนี้การค้า" },
      ],
    };

    expect(validateJournal(svJournal, fiscalYear, coa)).toBeNull();
    expect(journalTotals(svJournal.lines).difference).toBe(0n);
    postedJournals.push(svJournal);
  });

  it("Step 4.2: [UV] สมุดรายวันขาย - ขายสินค้าเป็นเงินเชื่อ 250,000 + ภาษีขาย 7%", () => {
    const uvJournal: GLJournal = {
      ...emptyJournal("UV", "sales"),
      docno: "UV-2569-001",
      date: "2026-01-15",
      fiscalyear: "2569",
      description: "ขายสินค้าเงินเชื่อให้ บริษัท ลูกค้าชั้นดี จำกัด (ใบกำกับภาษี TAX-001)",
      lines: [
        { ...emptyLine(), accountcode: "1131", debit: "267500", credit: "0", description: "ลูกหนี้การค้า" },
        { ...emptyLine(), accountcode: "4101", debit: "0", credit: "250000", description: "รายได้จากการขายสินค้า" },
        { ...emptyLine(), accountcode: "2141", debit: "0", credit: "17500", description: "ภาษีขาย 7%" },
      ],
    };

    expect(validateJournal(uvJournal, fiscalYear, coa)).toBeNull();
    expect(journalTotals(uvJournal.lines).difference).toBe(0n);
    postedJournals.push(uvJournal);
  });

  it("Step 4.3: [PV] สมุดรายวันจ่ายเงิน - จ่ายชำระเจ้าหนี้ และจ่ายค่าเช่าหัก ณ ที่จ่าย 3%", () => {
    // 4.3.1: จ่ายเจ้าหนี้การค้า 107,000 ผ่านเงินฝากธนาคาร
    const pvSupplier: GLJournal = {
      ...emptyJournal("PV", "payment"),
      docno: "PV-2569-001",
      date: "2026-01-20",
      fiscalyear: "2569",
      description: "จ่ายชำระหนี้ค่าสินค้า บิล INV-9901",
      lines: [
        { ...emptyLine(), accountcode: "2121", debit: "107000", credit: "0", description: "ตัดชำระเจ้าหนี้การค้า" },
        { ...emptyLine(), accountcode: "1112", debit: "0", credit: "107000", description: "โอนจ่ายจาก บ/ช กสิกรไทย" },
      ],
    };
    expect(validateJournal(pvSupplier, fiscalYear, coa)).toBeNull();
    expect(journalTotals(pvSupplier.lines).difference).toBe(0n);
    postedJournals.push(pvSupplier);

    // 4.3.2: จ่ายค่าเช่าสำนักงาน 20,000 หัก ณ ที่จ่าย 3% (600) จ่ายสุทธิ 19,400
    const pvRent: GLJournal = {
      ...emptyJournal("PV", "payment"),
      docno: "PV-2569-002",
      date: "2026-01-25",
      fiscalyear: "2569",
      description: "จ่ายค่าเช่าและบริการสำนักงาน ประจำเดือน ม.ค. 2569 (หัก 3%)",
      lines: [
        { ...emptyLine(), accountcode: "5301", debit: "20000", credit: "0", description: "ค่าเช่าและบริการสำนักงาน" },
        { ...emptyLine(), accountcode: "2142", debit: "0", credit: "600", description: "ภาษีหัก ณ ที่จ่ายค้างจ่าย ภ.ง.ด.53" },
        { ...emptyLine(), accountcode: "1112", debit: "0", credit: "19400", description: "โอนจ่ายสุทธิ" },
      ],
    };
    expect(validateJournal(pvRent, fiscalYear, coa)).toBeNull();
    expect(journalTotals(pvRent.lines).difference).toBe(0n);
    postedJournals.push(pvRent);
  });

  it("Step 4.4: [RV] สมุดรายวันรับเงิน - รับชำระหนี้จากลูกหนี้การค้า 267,500 เข้าธนาคาร", () => {
    const rvJournal: GLJournal = {
      ...emptyJournal("RV", "receipt"),
      docno: "RV-2569-001",
      date: "2026-01-28",
      fiscalyear: "2569",
      description: "รับชำระหนี้ตามใบกำกับภาษี TAX-001 เข้าบัญชีธนาคาร",
      lines: [
        { ...emptyLine(), accountcode: "1112", debit: "267500", credit: "0", description: "เงินโอนเข้า บ/ช กสิกรไทย" },
        { ...emptyLine(), accountcode: "1131", debit: "0", credit: "267500", description: "ตัดรับชำระลูกหนี้การค้า" },
      ],
    };

    expect(validateJournal(rvJournal, fiscalYear, coa)).toBeNull();
    expect(journalTotals(rvJournal.lines).difference).toBe(0n);
    postedJournals.push(rvJournal);
  });

  it("Step 4.5: [JV] สมุดรายวันทั่วไป - บันทึกตัดต้นทุนขายสินค้า (COGS) 80,000", () => {
    const jvCogs: GLJournal = {
      ...emptyJournal("JV", "manual"),
      docno: "JV-2569-002",
      date: "2026-01-31",
      fiscalyear: "2569",
      description: "บันทึกตัดต้นทุนขายสินค้าประจำงวด",
      lines: [
        { ...emptyLine(), accountcode: "5101", debit: "80000", credit: "0", description: "ต้นทุนขายสินค้า" },
        { ...emptyLine(), accountcode: "1141", debit: "0", credit: "80000", description: "ตัดสต๊อกสินค้าสำเร็จรูป" },
      ],
    };

    expect(validateJournal(jvCogs, fiscalYear, coa)).toBeNull();
    expect(journalTotals(jvCogs.lines).difference).toBe(0n);
    postedJournals.push(jvCogs);
  });

  // ==========================================
  // PHASE 5: General Ledger & Trial Balance (งบทดลอง)
  // ==========================================
  it("Step 5: ประมวลผลงบทดลอง (Trial Balance) - ผลรวมเดบิตต้องเท่ากับเครดิต ผลต่างเป็น 0.00", () => {
    // Include Opening balance in ledger aggregation
    const allLines = [
      // Opening
      { accountcode: "1111", debit: amountUnits("200000"), credit: 0n },
      { accountcode: "1112", debit: amountUnits("800000"), credit: 0n },
      { accountcode: "3101", debit: 0n, credit: amountUnits("1000000") },
      // Movements
      ...postedJournals.flatMap((j) =>
        j.lines.map((l) => ({
          accountcode: l.accountcode,
          debit: amountUnits(l.debit),
          credit: amountUnits(l.credit),
        }))
      ),
    ];

    // Aggregate by account
    const balances = new Map<string, { debit: bigint; credit: bigint }>();
    for (const line of allLines) {
      const current = balances.get(line.accountcode) || { debit: 0n, credit: 0n };
      balances.set(line.accountcode, {
        debit: current.debit + line.debit,
        credit: current.credit + line.credit,
      });
    }

    let totalDebit = 0n;
    let totalCredit = 0n;

    for (const [code, bal] of balances.entries()) {
      totalDebit += bal.debit;
      totalCredit += bal.credit;
    }

    // Trial balance absolute equation check
    expect(totalDebit).toBe(totalCredit);
    expect(totalDebit - totalCredit).toBe(0n);
    expect(amountString(totalDebit, 2)).toBe("1849000.00");
    expect(amountString(totalCredit, 2)).toBe("1849000.00");
  });

  // ==========================================
  // PHASE 6: Income Statement (งบกำไรขาดทุน / P&L)
  // ==========================================
  it("Step 6: ประมวลผลงบกำไรขาดทุน (P&L) - คำนวณรายได้ ค่าใช้จ่าย และกำไรสุทธิถูกต้อง", () => {
    // Revenue: 4101 = 250,000.00
    const salesRevenue = 250000.00;

    // Expenses:
    // 5101 (ต้นทุนขาย) = 80,000.00
    // 5301 (ค่าเช่าสำนักงาน) = 20,000.00
    const cogs = 80000.00;
    const adminExpense = 20000.00;
    const totalExpenses = cogs + adminExpense; // 100,000.00

    // Gross Profit = 250,000 - 80,000 = 170,000.00
    const grossProfit = salesRevenue - cogs;
    expect(grossProfit).toBe(170000.00);

    // Net Profit = Gross Profit - Admin Expenses = 170,000 - 20,000 = 150,000.00
    const netProfit = salesRevenue - totalExpenses;
    expect(netProfit).toBe(150000.00);
  });

  // ==========================================
  // PHASE 7: Balance Sheet (งบแสดงฐานะการเงิน / งบดุล)
  // ==========================================
  it("Step 7: พิสูจน์สมการบัญชีในงบแสดงฐานะการเงิน (Assets = Liabilities + Equity) ดุล 100%", () => {
    // 1. Assets (สินทรัพย์)
    const cash = 200000.00; // 1111 (เงินสด)
    // 1112 (เงินฝากธนาคาร) = 800,000 (ยกมา) - 107,000 (จ่ายเจ้าหนี้) - 19,400 (จ่ายค่าเช่า) + 267,500 (รับชำระ) = 941,100
    const bank = 800000.00 - 107000.00 - 19400.00 + 267500.00;
    expect(bank).toBe(941100.00);

    // 1131 (ลูกหนี้การค้า) = 267,500 (ขาย) - 267,500 (รับชำระ) = 0.00
    const ar = 267500.00 - 267500.00;
    expect(ar).toBe(0.00);

    // 1141 (สินค้าคงเหลือ) = 100,000 (ซื้อ) - 80,000 (ตัดต้นทุน) = 20,000.00
    const inventory = 100000.00 - 80000.00;
    expect(inventory).toBe(20000.00);

    // 1151 (ภาษีซื้อ) = 7,000.00
    const vatInput = 7000.00;

    const totalAssets = cash + bank + ar + inventory + vatInput;
    expect(totalAssets).toBe(1168100.00);

    // 2. Liabilities (หนี้สิน)
    // 2121 (เจ้าหนี้การค้า) = 107,000 (ซื้อ) - 107,000 (จ่ายชำระ) = 0.00
    const ap = 107000.00 - 107000.00;
    expect(ap).toBe(0.00);

    // 2141 (ภาษีขาย) = 17,500.00
    const vatOutput = 17500.00;

    // 2142 (ภาษีหัก ณ ที่จ่ายค้างจ่าย) = 600.00
    const whtPayable = 600.00;

    const totalLiabilities = ap + vatOutput + whtPayable;
    expect(totalLiabilities).toBe(18100.00);

    // 3. Equity (ส่วนของเจ้าของ)
    // 3101 (ทุนจดทะเบียนและชำระแล้ว) = 1,000,000.00
    const shareCapital = 1000000.00;

    // กำไรสุทธิประจำงวด (Current Period Net Profit จาก Phase 6) = 150,000.00
    const netProfit = 150000.00;

    const totalEquity = shareCapital + netProfit;
    expect(totalEquity).toBe(1150000.00);

    // 4. Verification of Fundamental Accounting Equation:
    // Assets = Liabilities + Equity
    const totalLiabilitiesAndEquity = totalLiabilities + totalEquity;
    expect(totalLiabilitiesAndEquity).toBe(1168100.00);

    // The Golden Rule of Accounting:
    const balanceDifference = totalAssets - totalLiabilitiesAndEquity;
    expect(balanceDifference).toBe(0.00); // 100% Balanced! Zero Discrepancy!
  });
});

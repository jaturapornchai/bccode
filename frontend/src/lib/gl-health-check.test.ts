import { describe, expect, it } from "vitest";
import { auditGLHealth } from "./gl-health-check";
import type { GLAccount } from "./general-ledger";

describe("GL Health & Anomaly Detector (auditGLHealth)", () => {
  const mockAccounts: GLAccount[] = [
    {
      accountcode: "1111-01",
      names: [{ code: "th", name: "เงินสดในมือ" }],
      accounttype: "asset",
      parentaccountcode: "1111",
      normalbalance: "debit",
      allowposting: true,
      isactive: true,
      accountgroup: "",
      iscash: true,
    },
    {
      accountcode: "1112-01",
      names: [{ code: "th", name: "เงินฝากกระแสรายวัน กสิกรไทย" }],
      accounttype: "asset",
      parentaccountcode: "1112",
      normalbalance: "debit",
      allowposting: true,
      isactive: true,
      accountgroup: "",
      iscash: true,
    },
    {
      accountcode: "1412-01",
      names: [{ code: "th", name: "ค่าเสื่อมราคาสะสม - อาคาร" }],
      accounttype: "asset",
      parentaccountcode: "1412",
      normalbalance: "credit",
      allowposting: true,
      isactive: true,
      accountgroup: "",
      iscash: false,
    },
    {
      accountcode: "2111-01",
      names: [{ code: "th", name: "เจ้าหนี้การค้าในประเทศ" }],
      accounttype: "liability",
      parentaccountcode: "2111",
      normalbalance: "credit",
      allowposting: true,
      isactive: true,
      accountgroup: "",
      iscash: false,
    },
    {
      accountcode: "1199-01",
      names: [{ code: "th", name: "เงินโอนระหว่างทาง (Clearing Suspense)" }],
      accounttype: "asset",
      parentaccountcode: "1199",
      normalbalance: "debit",
      allowposting: true,
      isactive: true,
      accountgroup: "",
      iscash: false,
    },
    {
      accountcode: "1299-99",
      names: [{ code: "th", name: "บัญชีลูกหนี้เก่าปิดใช้งานแล้ว" }],
      accounttype: "asset",
      parentaccountcode: "1299",
      normalbalance: "debit",
      allowposting: true,
      isactive: false,
      accountgroup: "",
      iscash: false,
    },
  ];

  it("returns 100% score and excellent status for clean balanced records", () => {
    const reportRows = [
      {
        accountcode: "1111-01",
        endingdebit: "50000.00",
        endingcredit: "0.00",
        debit: "10000.00",
        credit: "5000.00",
      },
      {
        accountcode: "2111-01",
        endingdebit: "0.00",
        endingcredit: "50000.00",
        debit: "5000.00",
        credit: "10000.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    expect(health.score).toBe(100);
    expect(health.status).toBe("excellent");
    expect(health.summary.criticalCount).toBe(0);
    expect(health.summary.warningCount).toBe(0);
    expect(health.anomalies.length).toBe(0);
  });

  it("detects trial balance imbalance as CRITICAL anomaly", () => {
    const reportRows = [
      {
        accountcode: "1111-01",
        endingdebit: "50000.00",
        endingcredit: "0.00",
      },
      {
        accountcode: "2111-01",
        endingdebit: "0.00",
        endingcredit: "45000.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    expect(health.status).toBe("critical");
    expect(health.summary.criticalCount).toBeGreaterThanOrEqual(1);
    const imbalance = health.anomalies.find((a) => a.category === "trial_balance_imbalance");
    expect(imbalance).toBeDefined();
    expect(imbalance?.severity).toBe("critical");
    expect(imbalance?.amount).toBe("5,000.00");
  });

  it("detects negative cash on hand as CRITICAL anomaly", () => {
    const reportRows = [
      {
        accountcode: "1111-01",
        endingdebit: "0.00",
        endingcredit: "12500.00", // Cash with credit balance
        debit: "0.00",
        credit: "12500.00",
      },
      {
        accountcode: "2111-01",
        endingdebit: "12500.00",
        endingcredit: "0.00",
        debit: "12500.00",
        credit: "0.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    const cashOverdrawn = health.anomalies.find((a) => a.category === "cash_bank_overdrawn");
    expect(cashOverdrawn).toBeDefined();
    expect(cashOverdrawn?.severity).toBe("critical");
    expect(cashOverdrawn?.messageTh).toContain("เงินสดในมือ");
    expect(cashOverdrawn?.canDrillLedger).toBe(true);
  });

  it("detects bank account credit balance as WARNING (Overdraft)", () => {
    const reportRows = [
      {
        accountcode: "1112-01",
        endingdebit: "0.00",
        endingcredit: "85000.00", // Bank overdraft
      },
      {
        accountcode: "2111-01",
        endingdebit: "85000.00",
        endingcredit: "0.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    const bankOverdrawn = health.anomalies.find((a) => a.accountCode === "1112-01");
    expect(bankOverdrawn).toBeDefined();
    expect(bankOverdrawn?.severity).toBe("warning");
    expect(bankOverdrawn?.messageTh).toContain("เบิกเกินบัญชี");
  });

  it("does NOT flag accumulated depreciation (contra-asset) with credit balance", () => {
    const reportRows = [
      {
        accountcode: "1412-01", // Accumulated depreciation (contra-asset)
        endingdebit: "0.00",
        endingcredit: "200000.00",
        debit: "0.00",
        credit: "10000.00",
      },
      {
        accountcode: "2111-01",
        endingdebit: "200000.00",
        endingcredit: "0.00",
        debit: "0.00",
        credit: "0.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    const contraViolation = health.anomalies.find(
      (a) => a.accountCode === "1412-01" && a.category === "normal_balance_violation"
    );
    expect(contraViolation).toBeUndefined();
  });

  it("detects inactive account with ending balance", () => {
    const reportRows = [
      {
        accountcode: "1299-99", // Inactive account
        endingdebit: "15000.00",
        endingcredit: "0.00",
      },
      {
        accountcode: "2111-01",
        endingdebit: "0.00",
        endingcredit: "15000.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    const inactiveAnomaly = health.anomalies.find((a) => a.category === "inactive_account_with_balance");
    expect(inactiveAnomaly).toBeDefined();
    expect(inactiveAnomaly?.severity).toBe("warning");
    expect(inactiveAnomaly?.accountCode).toBe("1299-99");
  });

  it("detects uncleared suspense / clearing account", () => {
    const reportRows = [
      {
        accountcode: "1199-01", // Clearing suspense
        endingdebit: "7500.00",
        endingcredit: "0.00",
        debit: "7500.00",
        credit: "0.00",
      },
      {
        accountcode: "2111-01",
        endingdebit: "0.00",
        endingcredit: "7500.00",
        debit: "0.00",
        credit: "0.00",
      },
    ];

    const health = auditGLHealth({
      reportRows,
      accounts: mockAccounts,
    });

    const suspenseAnomaly = health.anomalies.find((a) => a.category === "suspense_account_uncleared");
    expect(suspenseAnomaly).toBeDefined();
    expect(suspenseAnomaly?.severity).toBe("info");
    expect(suspenseAnomaly?.accountCode).toBe("1199-01");
  });
});

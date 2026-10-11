import { describe, expect, it } from "vitest";
import { centsToMoney, isMoneyText, moneyToCents, subtractMoney, sumMoney, todayIso } from "./cheque-forms";

// เงินของเช็คเป็นข้อความทศนิยม + BigInt สตางค์เท่านั้น (AGENTS.md Rule 20: ห้าม float)

describe("isMoneyText / moneyToCents", () => {
  it("accepts only the plain positive decimals the backend accepts", () => {
    for (const value of ["0", "12", "12.5", "12.50", " 7.00 ", "9999999999999999.99"]) expect(isMoneyText(value), value).toBe(true);
    for (const value of ["", "-1", "1,000", "1e3", "1.234", ".5", "12.", "abc", "12345678901234567"]) expect(isMoneyText(value), value).toBe(false);
  });

  it("converts to whole satang without float rounding", () => {
    expect(moneyToCents("12.5")).toBe(BigInt(1250));
    expect(moneyToCents("0.07")).toBe(BigInt(7));
    expect(moneyToCents("9007199254740993.01")).toBe(BigInt("900719925474099301"));
    expect(moneyToCents("1,000")).toBeNull();
    expect(moneyToCents("")).toBeNull();
  });
});

describe("centsToMoney", () => {
  it("always prints two decimals and keeps the sign", () => {
    expect(centsToMoney(BigInt(0))).toBe("0.00");
    expect(centsToMoney(BigInt(5))).toBe("0.05");
    expect(centsToMoney(BigInt(123450))).toBe("1234.50");
    expect(centsToMoney(BigInt(-250))).toBe("-2.50");
  });
});

describe("sumMoney / subtractMoney", () => {
  it("adds decimals exactly (0.1 + 0.2 is 0.30) and skips blank or invalid cells", () => {
    expect(sumMoney(["0.1", "0.2"])).toBe("0.30");
    expect(sumMoney(["1.00", "abc", ""])).toBe("1.00");
    expect(sumMoney([])).toBe("0.00");
    expect(sumMoney(["9007199254740993.01", "0.01"])).toBe("9007199254740993.02");
  });

  it("subtracts a fee from an amount, treating a blank fee as zero", () => {
    expect(subtractMoney("100.00", "0.01")).toBe("99.99");
    expect(subtractMoney("10", "")).toBe("10.00");
    expect(subtractMoney("5.00", "10.00")).toBe("-5.00");
  });
});

describe("todayIso", () => {
  it("uses the Asia/Bangkok calendar day, not the UTC day", () => {
    expect(todayIso(new Date("2026-10-11T16:59:00Z"))).toBe("2026-10-11");
    expect(todayIso(new Date("2026-10-11T17:01:00Z"))).toBe("2026-10-12");
  });
});

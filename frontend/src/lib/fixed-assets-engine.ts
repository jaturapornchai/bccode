/**
 * Fixed Assets Engine & Thai Revenue Code (Sec 65 bis (2)) Depreciation / Disposal Calculator
 * Standards: TFRS for NPAEs / PAEs & Thai Revenue Department
 */

export type ThaiAssetCategoryPreset = {
  categoryCode: string;
  nameTh: string;
  nameEn: string;
  standardUsefulLifeYears: number;
  standardDeprecPercent: number;
  taxCostLimit?: number; // e.g. 1,000,000 THB for passenger cars
  isSmeSpecialRuleAvailable?: boolean;
};

export const THAI_ASSET_CATEGORIES: ThaiAssetCategoryPreset[] = [
  {
    categoryCode: "BUILDING_PERM",
    nameTh: "อาคารถาวร",
    nameEn: "Permanent Building",
    standardUsefulLifeYears: 20,
    standardDeprecPercent: 5.0,
  },
  {
    categoryCode: "BUILDING_TEMP",
    nameTh: "อาคารชั่วคราว",
    nameEn: "Temporary Building",
    standardUsefulLifeYears: 1,
    standardDeprecPercent: 100.0,
  },
  {
    categoryCode: "VEHICLE_PASSENGER",
    nameTh: "ยานพาหนะ - รถยนต์นั่งไม่เกิน 10 ที่นั่ง (จำกัดภาษี 1 ลบ.)",
    nameEn: "Passenger Vehicle (Tax Limit 1MB)",
    standardUsefulLifeYears: 5,
    standardDeprecPercent: 20.0,
    taxCostLimit: 1000000.0,
  },
  {
    categoryCode: "VEHICLE_COMMERCIAL",
    nameTh: "ยานพาหนะ - รถบรรทุก/เชิงพาณิชย์",
    nameEn: "Commercial Vehicle",
    standardUsefulLifeYears: 5,
    standardDeprecPercent: 20.0,
  },
  {
    categoryCode: "MACHINERY",
    nameTh: "เครื่องจักรและอุปกรณ์การผลิต",
    nameEn: "Machinery & Equipment",
    standardUsefulLifeYears: 5,
    standardDeprecPercent: 20.0,
  },
  {
    categoryCode: "OFFICE_EQUIPMENT",
    nameTh: "เครื่องใช้และอุปกรณ์สำนักงาน",
    nameEn: "Office Equipment & Furniture",
    standardUsefulLifeYears: 5,
    standardDeprecPercent: 20.0,
  },
  {
    categoryCode: "COMPUTER",
    nameTh: "คอมพิวเตอร์และอุปกรณ์อิเล็กทรอนิกส์ (3 ปี)",
    nameEn: "Computers & Electronics",
    standardUsefulLifeYears: 3,
    standardDeprecPercent: 33.33,
    isSmeSpecialRuleAvailable: true,
  },
];

/**
 * Checks if a year is a leap year (366 days).
 */
export function isLeapYear(year: number): boolean {
  return (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
}

/**
 * Calculates Straight-Line Depreciation according to Thai Revenue Code pro-rata days.
 */
export function calculateStraightLineDepreciation(params: {
  cost: number;
  scrapValue?: number;
  usefulLifeYears: number;
  purchaseDate: string; // YYYY-MM-DD
  fiscalYear: number;
  priorAccumDeprec?: number;
}): {
  depreciableCost: number;
  fullYearDeprec: number;
  actualDaysInYear: number;
  totalDaysInYear: number;
  periodDeprec: number;
  accumDeprecToYearEnd: number;
  netBookValue: number;
} {
  const scrapValue = params.scrapValue ?? 1.0; // standard 1 THB scrap in Thailand
  const depreciableCost = Math.max(0, Math.round((params.cost - scrapValue) * 100) / 100);

  if (depreciableCost <= 0 || params.usefulLifeYears <= 0) {
    return {
      depreciableCost: 0,
      fullYearDeprec: 0,
      actualDaysInYear: 0,
      totalDaysInYear: 365,
      periodDeprec: 0,
      accumDeprecToYearEnd: 0,
      netBookValue: params.cost,
    };
  }

  const fullYearDeprec = Math.round((depreciableCost / params.usefulLifeYears) * 100) / 100;
  const totalDaysInYear = isLeapYear(params.fiscalYear) ? 366 : 365;

  const [pYear, pMonth, pDay] = params.purchaseDate.split("-").map((n) => parseInt(n, 10));
  const purchaseTime = new Date(Date.UTC(pYear, pMonth - 1, pDay)).getTime();
  const yearStartTime = new Date(Date.UTC(params.fiscalYear, 0, 1)).getTime();
  const yearEndTime = new Date(Date.UTC(params.fiscalYear, 11, 31)).getTime();

  let actualDays = 0;
  if (pYear > params.fiscalYear) {
    // Purchased in future year
    actualDays = 0;
  } else if (pYear === params.fiscalYear) {
    // Purchased during this fiscal year: pro-rata from purchase date to Dec 31
    const diffDays = Math.floor((yearEndTime - purchaseTime) / (1000 * 60 * 60 * 24)) + 1;
    actualDays = Math.min(totalDaysInYear, Math.max(1, diffDays));
  } else {
    // Purchased in prior year: full year
    actualDays = totalDaysInYear;
  }

  // Calculate pro-rata period depreciation
  let periodDeprec = 0;
  if (actualDays === totalDaysInYear) {
    periodDeprec = fullYearDeprec;
  } else {
    periodDeprec = Math.round(((fullYearDeprec * actualDays) / totalDaysInYear) * 100) / 100;
  }

  const priorAccum = params.priorAccumDeprec ?? 0;
  // Guard: Cannot depreciate more than depreciableCost
  const maxPossible = Math.max(0, depreciableCost - priorAccum);
  periodDeprec = Math.min(periodDeprec, maxPossible);

  const accumDeprecToYearEnd = Math.round((priorAccum + periodDeprec) * 100) / 100;
  const netBookValue = Math.round((params.cost - accumDeprecToYearEnd) * 100) / 100;

  return {
    depreciableCost,
    fullYearDeprec,
    actualDaysInYear: actualDays,
    totalDaysInYear,
    periodDeprec,
    accumDeprecToYearEnd,
    netBookValue,
  };
}

/**
 * Generates Balanced Depreciation Journal lines for Period Closing.
 */
export function generatePeriodicDepreciationJournalLines(
  items: {
    assetCode: string;
    assetName: string;
    expenseAccountCode: string;
    accumAccountCode: string;
    periodDeprec: number;
  }[],
): {
  totalDeprec: number;
  lines: { accountCode: string; debit: number; credit: number; description: string }[];
} {
  const validItems = items.filter((i) => i.periodDeprec > 0);
  const totalDeprec =
    Math.round(validItems.reduce((sum, i) => sum + i.periodDeprec, 0) * 100) / 100;

  const lines: { accountCode: string; debit: number; credit: number; description: string }[] = [];

  for (const item of validItems) {
    // Dr. Depreciation Expense
    lines.push({
      accountCode: item.expenseAccountCode || "520103",
      debit: item.periodDeprec,
      credit: 0,
      description: `ค่าเสื่อมราคา - ${item.assetName} (${item.assetCode})`,
    });

    // Cr. Accumulated Depreciation (Contra Asset)
    lines.push({
      accountCode: item.accumAccountCode || "129101",
      debit: 0,
      credit: item.periodDeprec,
      description: `ค่าเสื่อมราคาสะสม - ${item.assetName} (${item.assetCode})`,
    });
  }

  return { totalDeprec, lines };
}

/**
 * Calculates Asset Disposal and Generates Balanced Disposal Journal Entry.
 */
export function calculateAssetDisposal(params: {
  cost: number;
  accumDeprecAtDisposal: number;
  salePrice: number;
  hasVat?: boolean;
  settlementAccountCode?: string; // default 111201 (Bank)
  assetAccountCode?: string; // default 120101 (Cost)
  accumAccountCode?: string; // default 129101 (Accum Deprec)
  gainAccountCode?: string; // default 420101 (Gain on disposal)
  lossAccountCode?: string; // default 590101 (Loss on disposal)
  vatAccountCode?: string; // default 213101 (Output VAT)
}): {
  netBookValue: number;
  gainLoss: number;
  isGain: boolean;
  vatAmount: number;
  totalReceived: number;
  journalLines: { accountCode: string; debit: number; credit: number; description: string }[];
} {
  const {
    cost,
    accumDeprecAtDisposal,
    salePrice,
    hasVat = false,
    settlementAccountCode = "111201",
    assetAccountCode = "120101",
    accumAccountCode = "129101",
    gainAccountCode = "420101",
    lossAccountCode = "590101",
    vatAccountCode = "213101",
  } = params;

  const netBookValue = Math.round(Math.max(0, cost - accumDeprecAtDisposal) * 100) / 100;
  const gainLoss = Math.round((salePrice - netBookValue) * 100) / 100;
  const isGain = gainLoss >= 0;

  const vatAmount = hasVat ? Math.round(salePrice * 0.07 * 100) / 100 : 0;
  const totalReceived = Math.round((salePrice + vatAmount) * 100) / 100;

  const journalLines: {
    accountCode: string;
    debit: number;
    credit: number;
    description: string;
  }[] = [];

  // 1. Dr. Cash/Bank received
  if (totalReceived > 0) {
    journalLines.push({
      accountCode: settlementAccountCode,
      debit: totalReceived,
      credit: 0,
      description: "รับเงินจากการจำหน่ายสินทรัพย์",
    });
  }

  // 2. Dr. Clear Accumulated Depreciation
  if (accumDeprecAtDisposal > 0) {
    journalLines.push({
      accountCode: accumAccountCode,
      debit: accumDeprecAtDisposal,
      credit: 0,
      description: "ล้างค่าเสื่อมราคาสะสม",
    });
  }

  // 3. Dr. Loss on Disposal (if loss)
  if (!isGain && Math.abs(gainLoss) > 0) {
    journalLines.push({
      accountCode: lossAccountCode,
      debit: Math.abs(gainLoss),
      credit: 0,
      description: "ขาดทุนจากการจำหน่ายสินทรัพย์",
    });
  }

  // 4. Cr. Clear Original Asset Cost
  journalLines.push({
    accountCode: assetAccountCode,
    debit: 0,
    credit: cost,
    description: "ล้างราคาทุนสินทรัพย์",
  });

  // 5. Cr. Output VAT (if applicable)
  if (vatAmount > 0) {
    journalLines.push({
      accountCode: vatAccountCode,
      debit: 0,
      credit: vatAmount,
      description: "ภาษีขายจากการจำหน่ายสินทรัพย์ (7%)",
    });
  }

  // 6. Cr. Gain on Disposal (if gain)
  if (isGain && gainLoss > 0) {
    journalLines.push({
      accountCode: gainAccountCode,
      debit: 0,
      credit: gainLoss,
      description: "กำไรจากการจำหน่ายสินทรัพย์",
    });
  }

  // Balance Check (Satang Precision Guarantee)
  const sumDr = Math.round(journalLines.reduce((s, l) => s + l.debit, 0) * 100) / 100;
  const sumCr = Math.round(journalLines.reduce((s, l) => s + l.credit, 0) * 100) / 100;
  const diff = Math.round((sumDr - sumCr) * 100) / 100;

  if (Math.abs(diff) >= 0.01) {
    // Adjust satang difference into gain/loss line
    const target = journalLines.find(
      (l) => l.accountCode === gainAccountCode || l.accountCode === lossAccountCode,
    );
    if (target) {
      if (target.credit > 0) target.credit = Math.round((target.credit + diff) * 100) / 100;
      else target.debit = Math.round((target.debit - diff) * 100) / 100;
    }
  }

  return {
    netBookValue,
    gainLoss,
    isGain,
    vatAmount,
    totalReceived,
    journalLines,
  };
}

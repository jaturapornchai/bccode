/**
 * CFO Financial Health Dashboard & Cash Flow Statement Engine
 * Standards: TFRS for NPAEs / PAEs (Executive Key Financial Indicators & Indirect Cash Flow Statement)
 */

export type CFOBalanceMetrics = {
  cashAndBank: number;
  tradeReceivables: number;
  inventory: number;
  otherCurrentAssets: number;
  nonCurrentAssets: number;
  totalAssets: number;

  tradePayables: number;
  shortTermLoans: number;
  otherCurrentLiabilities: number;
  longTermLiabilities: number;
  totalLiabilities: number;

  totalEquity: number;

  revenue: number;
  costOfGoodsSold: number;
  grossProfit: number;
  operatingExpenses: number;
  netProfit: number;

  monthlyOperatingBurnRate?: number;
};

export type RatioStatus = "HEALTHY" | "WARNING" | "CRITICAL";

export type FinancialRatioMetric = {
  key: string;
  nameTh: string;
  nameEn: string;
  value: number;
  formatted: string;
  status: RatioStatus;
  benchmark: string;
  advice: string;
};

export type CFOHealthDashboardData = {
  healthScore: number; // 0 - 100
  overallStatus: RatioStatus;
  cashRunwayMonths: number;
  ratios: FinancialRatioMetric[];
};

export type CashFlowStatementData = {
  operating: {
    netProfit: number;
    depreciation: number;
    arChange: number; // negative if AR increased
    inventoryChange: number; // negative if inventory increased
    apChange: number; // positive if AP increased
    netOperatingCashFlow: number;
  };
  investing: {
    capexPurchases: number; // cash outflow (-)
    assetDisposalProceeds: number; // cash inflow (+)
    netInvestingCashFlow: number;
  };
  financing: {
    loanProceeds: number; // cash inflow (+)
    loanRepayments: number; // cash outflow (-)
    dividendsPaid: number; // cash outflow (-)
    netFinancingCashFlow: number;
  };
  summary: {
    netCashChange: number;
    beginningCash: number;
    endingCash: number;
  };
};

/**
 * Calculates Executive Financial Ratios and overall Financial Health Score.
 */
export function calculateCFOFinancialHealth(m: CFOBalanceMetrics): CFOHealthDashboardData {
  const currentAssets = m.cashAndBank + m.tradeReceivables + m.inventory + m.otherCurrentAssets;
  const currentLiabilities = m.tradePayables + m.shortTermLoans + m.otherCurrentLiabilities;

  // 1. Current Ratio
  const currentRatio = currentLiabilities > 0 ? currentAssets / currentLiabilities : 99.0;
  const currentRatioStatus: RatioStatus =
    currentRatio >= 1.5 ? "HEALTHY" : currentRatio >= 1.0 ? "WARNING" : "CRITICAL";

  // 2. Quick Ratio
  const quickAssets = m.cashAndBank + m.tradeReceivables;
  const quickRatio = currentLiabilities > 0 ? quickAssets / currentLiabilities : 99.0;
  const quickRatioStatus: RatioStatus =
    quickRatio >= 1.0 ? "HEALTHY" : quickRatio >= 0.7 ? "WARNING" : "CRITICAL";

  // 3. Debt to Equity (D/E)
  const deRatio = m.totalEquity > 0 ? m.totalLiabilities / m.totalEquity : 99.0;
  const deRatioStatus: RatioStatus =
    deRatio <= 1.5 ? "HEALTHY" : deRatio <= 2.5 ? "WARNING" : "CRITICAL";

  // 4. Gross Profit Margin (%)
  const gpMargin = m.revenue > 0 ? (m.grossProfit / m.revenue) * 100 : 0;
  const gpMarginStatus: RatioStatus =
    gpMargin >= 30 ? "HEALTHY" : gpMargin >= 15 ? "WARNING" : "CRITICAL";

  // 5. Net Profit Margin (%)
  const npMargin = m.revenue > 0 ? (m.netProfit / m.revenue) * 100 : 0;
  const npMarginStatus: RatioStatus =
    npMargin >= 10 ? "HEALTHY" : npMargin >= 3 ? "WARNING" : "CRITICAL";

  // 6. Cash Runway
  const burnRate = m.monthlyOperatingBurnRate || Math.max(1, m.operatingExpenses / 12);
  const runwayMonths = burnRate > 0 ? Math.round((m.cashAndBank / burnRate) * 10) / 10 : 99.0;

  // Calculate Overall Health Score (0-100)
  let score = 50;
  if (currentRatioStatus === "HEALTHY") score += 10;
  else if (currentRatioStatus === "CRITICAL") score -= 15;

  if (quickRatioStatus === "HEALTHY") score += 10;
  else if (quickRatioStatus === "CRITICAL") score -= 10;

  if (deRatioStatus === "HEALTHY") score += 10;
  else if (deRatioStatus === "CRITICAL") score -= 15;

  if (gpMarginStatus === "HEALTHY") score += 10;
  else if (gpMarginStatus === "CRITICAL") score -= 10;

  if (npMarginStatus === "HEALTHY") score += 10;
  else if (npMarginStatus === "CRITICAL") score -= 15;

  if (runwayMonths >= 6) score += 10;
  else if (runwayMonths < 3) score -= 15;

  score = Math.max(0, Math.min(100, score));
  const overallStatus: RatioStatus = score >= 75 ? "HEALTHY" : score >= 50 ? "WARNING" : "CRITICAL";

  const ratios: FinancialRatioMetric[] = [
    {
      key: "current_ratio",
      nameTh: "อัตราส่วนทุนหมุนเวียน (Current Ratio)",
      nameEn: "Current Ratio",
      value: Math.round(currentRatio * 100) / 100,
      formatted: `${currentRatio.toFixed(2)}x`,
      status: currentRatioStatus,
      benchmark: "≥ 1.50x",
      advice:
        currentRatio >= 1.5
          ? "สภาพคล่องแข็งแกร่ง สินทรัพย์หมุนเวียนคุ้มหนี้สินระยะสั้นได้สบาย"
          : "ควรเร่งติดตามลูกหนี้และระวังภาระหนี้สินระยะสั้นที่จะถึงกำหนด",
    },
    {
      key: "quick_ratio",
      nameTh: "อัตราส่วนทุนหมุนเวียนเร็ว (Quick Ratio)",
      nameEn: "Quick Ratio",
      value: Math.round(quickRatio * 100) / 100,
      formatted: `${quickRatio.toFixed(2)}x`,
      status: quickRatioStatus,
      benchmark: "≥ 1.00x",
      advice:
        quickRatio >= 1.0
          ? "เงินสดและลูกหนี้เพียงพอชำระหนี้สินระยะสั้นโดยไม่ต้องพึ่งการระบายสต๊อก"
          : "อาจต้องพึ่งการขายสต๊อกเพื่อนำเงินมาชำระหนี้ระยะสั้น",
    },
    {
      key: "debt_to_equity",
      nameTh: "อัตราส่วนหนี้สินต่อส่วนของผู้ถือหุ้น (D/E)",
      nameEn: "Debt-to-Equity",
      value: Math.round(deRatio * 100) / 100,
      formatted: `${deRatio.toFixed(2)}x`,
      status: deRatioStatus,
      benchmark: "≤ 1.50x",
      advice:
        deRatio <= 1.5
          ? "โครงสร้างทางการเงินปลอดภัย ภาระหนี้สินอยู่ในเกณฑ์ควบคุมได้ดี"
          : "กิจการพึ่งพาเงินกู้ยืมสูง ควรระวังดอกเบี้ยจ่ายและความเสี่ยงสภาพคล่อง",
    },
    {
      key: "gross_profit_margin",
      nameTh: "อัตรากำไรขั้นต้น (GP Margin)",
      nameEn: "Gross Profit Margin",
      value: Math.round(gpMargin * 10) / 10,
      formatted: `${gpMargin.toFixed(1)}%`,
      status: gpMarginStatus,
      benchmark: "≥ 30.0%",
      advice:
        gpMargin >= 30
          ? "ส่วนต่างราคาขายต่อต้นทุนสินค้าดีเยี่ยม สะท้อนความสามารถในการแข่งขัน"
          : "ต้นทุนสินค้าค่อนข้างสูง ควรพิจารณาเจรจาต่อรองผู้จำหน่ายหรือปรับราคาขาย",
    },
    {
      key: "net_profit_margin",
      nameTh: "อัตรากำไรสุทธิ (NP Margin)",
      nameEn: "Net Profit Margin",
      value: Math.round(npMargin * 10) / 10,
      formatted: `${npMargin.toFixed(1)}%`,
      status: npMarginStatus,
      benchmark: "≥ 10.0%",
      advice:
        npMargin >= 10
          ? "ความสามารถในการทำกำไรสุทธิแข็งแรง แปลงรายได้เป็นผลกำไรได้ดี"
          : "กำไรสุทธิค่อนข้างบาง ควรตรวจสอบและควบคุมค่าใช้จ่ายในการดำเนินงาน (OPEX)",
    },
    {
      key: "cash_runway",
      nameTh: "ระยะเวลาความอยู่รอดของเงินสด (Cash Runway)",
      nameEn: "Cash Runway",
      value: runwayMonths,
      formatted: `${runwayMonths.toFixed(1)} เดือน`,
      status: runwayMonths >= 6 ? "HEALTHY" : runwayMonths >= 3 ? "WARNING" : "CRITICAL",
      benchmark: "≥ 6.0 เดือน",
      advice:
        runwayMonths >= 6
          ? "เงินสดสำรองเพียงพอสำหรับการดำเนินงานเกิน 6 เดือน ปลอดภัยมาก"
          : "เงินสดสำรองเหลือน้อยกว่า 6 เดือน ควรวางแผนกระแสเงินสดและจัดหาเงินทุนเพิ่ม",
    },
  ];

  return {
    healthScore: score,
    overallStatus,
    cashRunwayMonths: runwayMonths,
    ratios,
  };
}

/**
 * Calculates Cash Flow Statement using Indirect Method (TFRS for NPAEs / PAEs).
 */
export function calculateCashFlowStatement(params: {
  beginningCash: number;
  netProfit: number;
  depreciation: number;
  beginningAR: number;
  endingAR: number;
  beginningInventory: number;
  endingInventory: number;
  beginningAP: number;
  endingAP: number;
  capexPurchases?: number;
  assetDisposalProceeds?: number;
  loanProceeds?: number;
  loanRepayments?: number;
  dividendsPaid?: number;
}): CashFlowStatementData {
  const {
    beginningCash,
    netProfit,
    depreciation,
    beginningAR,
    endingAR,
    beginningInventory,
    endingInventory,
    beginningAP,
    endingAP,
    capexPurchases = 0,
    assetDisposalProceeds = 0,
    loanProceeds = 0,
    loanRepayments = 0,
    dividendsPaid = 0,
  } = params;

  // Operating Activities (Indirect Method)
  // Increase in AR => Cash Outflow (-)
  const arChange = Math.round((beginningAR - endingAR) * 100) / 100;
  // Increase in Inventory => Cash Outflow (-)
  const inventoryChange = Math.round((beginningInventory - endingInventory) * 100) / 100;
  // Increase in AP => Cash Inflow (+)
  const apChange = Math.round((endingAP - beginningAP) * 100) / 100;

  const netOperatingCashFlow =
    Math.round((netProfit + depreciation + arChange + inventoryChange + apChange) * 100) / 100;

  // Investing Activities
  const netInvestingCashFlow =
    Math.round((-Math.abs(capexPurchases) + assetDisposalProceeds) * 100) / 100;

  // Financing Activities
  const netFinancingCashFlow =
    Math.round((loanProceeds - Math.abs(loanRepayments) - Math.abs(dividendsPaid)) * 100) / 100;

  // Total Summary
  const netCashChange =
    Math.round((netOperatingCashFlow + netInvestingCashFlow + netFinancingCashFlow) * 100) / 100;
  const endingCash = Math.round((beginningCash + netCashChange) * 100) / 100;

  return {
    operating: {
      netProfit,
      depreciation,
      arChange,
      inventoryChange,
      apChange,
      netOperatingCashFlow,
    },
    investing: {
      capexPurchases: -Math.abs(capexPurchases),
      assetDisposalProceeds,
      netInvestingCashFlow,
    },
    financing: {
      loanProceeds,
      loanRepayments: -Math.abs(loanRepayments),
      dividendsPaid: -Math.abs(dividendsPaid),
      netFinancingCashFlow,
    },
    summary: {
      netCashChange,
      beginningCash,
      endingCash,
    },
  };
}

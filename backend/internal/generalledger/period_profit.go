package generalledger

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"
)

// PeriodProfit holds the two read-only columns of Champ 5511 (GLFrmPeriod.cpp:125-128, 211-214):
// NetProfitLoss = revenue - expense of the period (profit positive, loss negative; Champ stores
// -(expense - revenue) at close, glfrmcloseperiod.cpp:474-476) and SummProfitLoss = accumulated
// profit up to the end of the period.
type PeriodProfit struct {
	Code           string `json:"code"`
	NetProfitLoss  Amount `json:"netprofitloss"`
	SummProfitLoss Amount `json:"summprofitloss"`
}

type PeriodProfitResult struct {
	FiscalYear string         `json:"fiscalyear"`
	Periods    []PeriodProfit `json:"periods"`
}

// Champ writes both numbers into BCPeriod when a period is closed; the ledger here has no closing
// snapshot per period, so they are derived from the posted lines on every read. Posted lines only
// (gl_lines), opening/closing journals never count as profit, and revenue/expense is decided by
// the account type kept on each line (account master), never by account code.
//
// Net(p) is SUM(credit - debit) of the income and expense lines dated inside p. Summ(p) is the fiscal
// year's retained earnings account balance (credit - debit, closing journals skipped because their profit
// transfer is already counted through the income and expense lines) plus the profit of the fiscal year,
// both up to the end of p. Champ reads the retained earnings of the periods before p only; lines posted
// straight to retained earnings inside p (for example a dividend) are counted here in the period they happen.
// Champ adds the retained earnings balance as debit - credit (glfrmcloseperiod.cpp:479-484) although
// profit is credited to it (GLFrmClosePeriodPage2.cpp:219-237); credit - debit is used here so a
// credit balance (accumulated profit) is positive like the Net column.
const periodProfitSQL = `WITH retained AS (
  SELECT COALESCE((SELECT payload->>'retainedearningsaccount' FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)),'') AS code
), periods AS (
  SELECT code,(payload->>'startdate')::date AS start_date,(payload->>'enddate')::date AS end_date
  FROM gl_records WHERE company=$1 AND kind='periods' AND payload->>'fiscalyear'=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)
)
SELECT p.code,
  COALESCE(SUM(l.credit-l.debit) FILTER (WHERE l.account_type IN ('income','expense') AND l.kind NOT IN ('opening','closing') AND l.entry_date>=p.start_date),0)::text,
  COALESCE(SUM(l.credit-l.debit) FILTER (WHERE (l.account_type IN ('income','expense') AND l.kind NOT IN ('opening','closing'))
    OR (r.code<>'' AND l.account_code=r.code AND l.kind<>'closing')),0)::text
FROM periods p CROSS JOIN retained r
LEFT JOIN gl_lines l ON l.company=$1 AND l.fiscal_year=$2 AND l.entry_date<=p.end_date AND ($3='' OR l.branch_code=$3)
GROUP BY p.code,p.start_date ORDER BY p.start_date,p.code`

// PeriodProfit returns the Net/Summ profit of every period of a fiscal year. A branch session only
// sees the lines of its own branch, the same as the reports.
func (p *Postgres) PeriodProfit(ctx context.Context, scope Scope, fiscalYear string) (PeriodProfitResult, error) {
	result := PeriodProfitResult{FiscalYear: strings.TrimSpace(fiscalYear), Periods: []PeriodProfit{}}
	if result.FiscalYear == "" {
		return result, nil
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return result, err
	}
	rows, err := db.QueryContext(ctx, periodProfitSQL, scope.Company, result.FiscalYear, scope.Branch)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var code, net, summ string
		if err = rows.Scan(&code, &net, &summ); err != nil {
			return result, err
		}
		netValue, err := decimal.NewFromString(net)
		if err != nil {
			return result, err
		}
		summValue, err := decimal.NewFromString(summ)
		if err != nil {
			return result, err
		}
		result.Periods = append(result.Periods, PeriodProfit{Code: code, NetProfitLoss: Amount(netValue.String()), SummProfitLoss: Amount(summValue.String())})
	}
	return result, rows.Err()
}

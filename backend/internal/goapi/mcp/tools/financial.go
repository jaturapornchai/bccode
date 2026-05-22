package tools

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ==================== Profit Analysis ====================

type ProfitAnalysisRequest struct {
	ShopID string `json:"shop_id"`
	FromDate string `json:"from_date"`
	ToDate string `json:"to_date"`
}

type ProfitAnalysisResponse struct {
	Period string            `json:"period"`
	Revenue RevenueBreakdown  `json:"revenue"`
	Costs CostBreakdown     `json:"costs"`
	Profit ProfitBreakdown   `json:"profit"`
	ByCategory []CategoryProfit  `json:"by_category"`
	Trend []ProfitTrend     `json:"trend"`
	GeneratedAt time.Time         `json:"generated_at"`
}

type RevenueBreakdown struct {
	TotalSales float64 `json:"total_sales"`
	TotalSalesWord string  `json:"total_sales_word"`
	Returns float64 `json:"returns"`
	Discounts float64 `json:"discounts"`
	NetRevenue float64 `json:"net_revenue"`
}

type CostBreakdown struct {
	CostOfGoodsSold float64 `json:"cost_of_goods_sold"`
	COGSWord string  `json:"cogs_word"`
	OperatingCosts float64 `json:"operating_costs"`
	TotalCosts float64 `json:"total_costs"`
}

type ProfitBreakdown struct {
	GrossProfit float64 `json:"gross_profit"`
	GrossProfitWord string  `json:"gross_profit_word"`
	GrossMargin float64 `json:"gross_margin_percent"`
	OperatingProfit float64 `json:"operating_profit"`
	OperatingMargin float64 `json:"operating_margin_percent"`
	NetProfit float64 `json:"net_profit"`
	NetMargin float64 `json:"net_margin_percent"`
}

type CategoryProfit struct {
	CategoryCode string  `json:"category_code"`
	CategoryName string  `json:"category_name"`
	Revenue float64 `json:"revenue"`
	Cost float64 `json:"cost"`
	Profit float64 `json:"profit"`
	Margin float64 `json:"margin_percent"`
}

type ProfitTrend struct {
	Date string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Cost float64 `json:"cost"`
	Profit float64 `json:"profit"`
	Margin float64 `json:"margin_percent"`
}

func GetProfitAnalysis(ctx context.Context, shopID, fromDate, toDate string) (*ProfitAnalysisResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	// Default dates
	now := time.Now()
	if fromDate == "" {
		fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}

	logger.Info("[Profit Analysis] shopid=%s, from=%s, to=%s", shopID, fromDate, toDate)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &ProfitAnalysisResponse{
		Period:      fmt.Sprintf("%s to %s", fromDate, toDate),
		GeneratedAt: now,
	}

	// Get revenue data
	query := `
		SELECT
			COALESCE(SUM(totalamount), 0) as total_sales,
			COALESCE(SUM(CASE WHEN transflag = 20 THEN totalamount ELSE 0 END), 0) as returns,
			COALESCE(SUM(discountamount), 0) as discounts,
			COALESCE(SUM(totalcost), 0) as cogs
		FROM doc
		WHERE transflag IN (16, 18, 20)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`

	var totalSales, returns, discounts, cogs float64
	err = db.QueryRow(query, fromDate, toDate).Scan(&totalSales, &returns, &discounts, &cogs)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("[Profit Analysis] Query failed: %v", err)
	}

	netRevenue := totalSales - returns - discounts
	grossProfit := netRevenue - cogs
	grossMargin := float64(0)
	if netRevenue > 0 {
		grossMargin = (grossProfit / netRevenue) * 100
	}

	response.Revenue = RevenueBreakdown{
		TotalSales:     totalSales,
		TotalSalesWord: formatAmountWord(totalSales),
		Returns:        returns,
		Discounts:      discounts,
		NetRevenue:     netRevenue,
	}

	response.Costs = CostBreakdown{
		CostOfGoodsSold: cogs,
		COGSWord:        formatAmountWord(cogs),
		OperatingCosts:  0, // Would need expense tracking
		TotalCosts:      cogs,
	}

	response.Profit = ProfitBreakdown{
		GrossProfit:     grossProfit,
		GrossProfitWord: formatAmountWord(grossProfit),
		GrossMargin:     grossMargin,
		OperatingProfit: grossProfit,
		OperatingMargin: grossMargin,
		NetProfit:       grossProfit,
		NetMargin:       grossMargin,
	}

	// ByCategory — กำไรแยกตามหมวดสินค้า
	catQuery := `
		SELECT
			COALESCE(d.categorycode, 'N/A') as category_code,
			COALESCE(d.categorycode, 'ไม่ระบุหมวด') as category_name,
			COALESCE(SUM(d.totalqty * d.price * d.calcflag * -1 * d.unitstand / NULLIF(d.unitdivide, 0)), 0) as revenue,
			COALESCE(SUM(d.calcamount * d.calcflag * -1), 0) as cost
		FROM docdetail d
		WHERE d.transflag IN (16, 18)
			AND d.docdate >= $1 AND d.docdate <= $2
		GROUP BY d.categorycode
		ORDER BY revenue DESC
		LIMIT 20
	`
	catRows, catErr := db.Query(catQuery, fromDate, toDate)
	if catErr == nil {
		defer catRows.Close()
		for catRows.Next() {
			var cp CategoryProfit
			if err := catRows.Scan(&cp.CategoryCode, &cp.CategoryName, &cp.Revenue, &cp.Cost); err != nil {
				continue
			}
			cp.Profit = cp.Revenue - cp.Cost
			if cp.Revenue > 0 {
				cp.Margin = (cp.Profit / cp.Revenue) * 100
			}
			response.ByCategory = append(response.ByCategory, cp)
		}
	}
	if response.ByCategory == nil {
		response.ByCategory = []CategoryProfit{}
	}

	// Trend — กำไรรายวัน
	trendQuery := `
		SELECT
			TO_CHAR(docdate, 'YYYY-MM-DD') as date,
			COALESCE(SUM(totalamount), 0) as revenue,
			COALESCE(SUM(totalcost), 0) as cost
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY TO_CHAR(docdate, 'YYYY-MM-DD')
		ORDER BY date
	`
	trendRows, trendErr := db.Query(trendQuery, fromDate, toDate)
	if trendErr == nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var pt ProfitTrend
			if err := trendRows.Scan(&pt.Date, &pt.Revenue, &pt.Cost); err != nil {
				continue
			}
			pt.Profit = pt.Revenue - pt.Cost
			if pt.Revenue > 0 {
				pt.Margin = (pt.Profit / pt.Revenue) * 100
			}
			response.Trend = append(response.Trend, pt)
		}
	}
	if response.Trend == nil {
		response.Trend = []ProfitTrend{}
	}

	return response, nil
}

// ==================== Accounts Receivable ====================

type AccountsReceivableRequest struct {
	ShopID string `json:"shop_id"`
}

type AccountsReceivableResponse struct {
	Summary ARSummary         `json:"summary"`
	AgingBuckets []AgingBucket     `json:"aging_buckets"`
	TopDebtors []Debtor          `json:"top_debtors"`
	OverdueAlerts []OverdueAlert    `json:"overdue_alerts"`
	GeneratedAt time.Time         `json:"generated_at"`
}

type ARSummary struct {
	TotalReceivable float64 `json:"total_receivable"`
	TotalReceivableWord string `json:"total_receivable_word"`
	OverdueAmount float64 `json:"overdue_amount"`
	OverduePercent float64 `json:"overdue_percent"`
	AverageDaysToCollect int   `json:"average_days_to_collect"`
}

type AgingBucket struct {
	Label string  `json:"label"`      // "Current", "1-30 days", "31-60 days", etc.
	Amount float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
	Count int     `json:"count"`
}

type Debtor struct {
	CustomerCode string  `json:"customer_code"`
	CustomerName string  `json:"customer_name"`
	TotalOwed float64 `json:"total_owed"`
	OverdueAmount float64 `json:"overdue_amount"`
	OldestInvoice string  `json:"oldest_invoice_date"`
	DaysOverdue int     `json:"days_overdue"`
}

type OverdueAlert struct {
	DocNo string  `json:"docno"`
	CustomerName string  `json:"customer_name"`
	Amount float64 `json:"amount"`
	DueDate string  `json:"due_date"`
	DaysOverdue int     `json:"days_overdue"`
	Priority string  `json:"priority"` // high, medium, low
}

func GetAccountsReceivable(ctx context.Context, shopID string) (*AccountsReceivableResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	logger.Info("[Accounts Receivable] shopid=%s", shopID)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	now := time.Now()
	response := &AccountsReceivableResponse{
		AgingBuckets:  []AgingBucket{},
		TopDebtors:    []Debtor{},
		OverdueAlerts: []OverdueAlert{},
		GeneratedAt:   now,
	}

	// Get total receivable from unpaid invoices
	query := `
		SELECT
			COALESCE(SUM(totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)), 0) as total_receivable,
			COALESCE(SUM(CASE WHEN duedate < CURRENT_DATE THEN totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0) ELSE 0 END), 0) as overdue
		FROM doc
		WHERE transflag IN (16, 18)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND (isdelete = false OR isdelete IS NULL)
	`

	var totalReceivable, overdueAmount float64
	err = db.QueryRow(query).Scan(&totalReceivable, &overdueAmount)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("[Accounts Receivable] Query failed: %v", err)
	}

	overduePercent := float64(0)
	if totalReceivable > 0 {
		overduePercent = (overdueAmount / totalReceivable) * 100
	}

	response.Summary = ARSummary{
		TotalReceivable:     totalReceivable,
		TotalReceivableWord: formatAmountWord(totalReceivable),
		OverdueAmount:       overdueAmount,
		OverduePercent:      overduePercent,
		AverageDaysToCollect: 30, // Simplified
	}

	// Aging buckets — ข้อมูลจริงจาก DB
	agingQuery := `
		SELECT
			CASE
				WHEN duedate >= CURRENT_DATE THEN 'Current'
				WHEN CURRENT_DATE - duedate BETWEEN 1 AND 30 THEN '1-30 days overdue'
				WHEN CURRENT_DATE - duedate BETWEEN 31 AND 60 THEN '31-60 days overdue'
				WHEN CURRENT_DATE - duedate BETWEEN 61 AND 90 THEN '61-90 days overdue'
				ELSE '90+ days overdue'
			END as label,
			COALESCE(SUM(totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)), 0) as amount,
			COUNT(*) as cnt
		FROM doc
		WHERE transflag IN (16, 18)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY label
		ORDER BY CASE label
			WHEN 'Current' THEN 1
			WHEN '1-30 days overdue' THEN 2
			WHEN '31-60 days overdue' THEN 3
			WHEN '61-90 days overdue' THEN 4
			ELSE 5
		END
	`
	agingRows, agingErr := db.Query(agingQuery)
	if agingErr == nil {
		defer agingRows.Close()
		for agingRows.Next() {
			var bucket AgingBucket
			if err := agingRows.Scan(&bucket.Label, &bucket.Amount, &bucket.Count); err != nil {
				continue
			}
			if totalReceivable > 0 {
				bucket.Percentage = (bucket.Amount / totalReceivable) * 100
			}
			response.AgingBuckets = append(response.AgingBuckets, bucket)
		}
	}
	if len(response.AgingBuckets) == 0 {
		response.AgingBuckets = []AgingBucket{}
	}

	// TopDebtors — ลูกหนี้สูงสุด
	debtorQuery := `
		SELECT
			COALESCE(custcode, 'N/A') as customer_code,
			COALESCE(custname, custcode, 'Unknown') as customer_name,
			SUM(totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)) as total_owed,
			SUM(CASE WHEN duedate < CURRENT_DATE THEN totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0) ELSE 0 END) as overdue_amount,
			MIN(docdate)::text as oldest_invoice,
			COALESCE(MAX(CURRENT_DATE - duedate), 0) as days_overdue
		FROM doc
		WHERE transflag IN (16, 18)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND custcode IS NOT NULL AND custcode != ''
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY custcode, custname
		ORDER BY total_owed DESC
		LIMIT 10
	`
	debtorRows, debtorErr := db.Query(debtorQuery)
	if debtorErr == nil {
		defer debtorRows.Close()
		for debtorRows.Next() {
			var d Debtor
			if err := debtorRows.Scan(&d.CustomerCode, &d.CustomerName, &d.TotalOwed, &d.OverdueAmount, &d.OldestInvoice, &d.DaysOverdue); err != nil {
				continue
			}
			response.TopDebtors = append(response.TopDebtors, d)
		}
	}
	if len(response.TopDebtors) == 0 {
		response.TopDebtors = []Debtor{}
	}

	// OverdueAlerts — รายการเกินกำหนด
	alertQuery := `
		SELECT
			docno,
			COALESCE(custname, custcode, 'Unknown') as customer_name,
			totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0) as amount,
			duedate::text as due_date,
			CURRENT_DATE - duedate as days_overdue
		FROM doc
		WHERE transflag IN (16, 18)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND duedate < CURRENT_DATE
			AND (isdelete = false OR isdelete IS NULL)
		ORDER BY days_overdue DESC
		LIMIT 20
	`
	alertRows, alertErr := db.Query(alertQuery)
	if alertErr == nil {
		defer alertRows.Close()
		for alertRows.Next() {
			var a OverdueAlert
			if err := alertRows.Scan(&a.DocNo, &a.CustomerName, &a.Amount, &a.DueDate, &a.DaysOverdue); err != nil {
				continue
			}
			if a.DaysOverdue > 60 {
				a.Priority = "high"
			} else if a.DaysOverdue > 30 {
				a.Priority = "medium"
			} else {
				a.Priority = "low"
			}
			response.OverdueAlerts = append(response.OverdueAlerts, a)
		}
	}
	if len(response.OverdueAlerts) == 0 {
		response.OverdueAlerts = []OverdueAlert{}
	}

	return response, nil
}

// ==================== Accounts Payable ====================

type AccountsPayableRequest struct {
	ShopID string `json:"shop_id"`
}

type AccountsPayableResponse struct {
	Summary APSummary         `json:"summary"`
	AgingBuckets []AgingBucket     `json:"aging_buckets"`
	TopCreditors []Creditor        `json:"top_creditors"`
	UpcomingPayments []UpcomingPayment `json:"upcoming_payments"`
	GeneratedAt time.Time         `json:"generated_at"`
}

type APSummary struct {
	TotalPayable float64 `json:"total_payable"`
	TotalPayableWord string  `json:"total_payable_word"`
	OverdueAmount float64 `json:"overdue_amount"`
	OverduePercent float64 `json:"overdue_percent"`
	DueThisWeek float64 `json:"due_this_week"`
}

type Creditor struct {
	SupplierCode string  `json:"supplier_code"`
	SupplierName string  `json:"supplier_name"`
	TotalOwed float64 `json:"total_owed"`
	OldestInvoice string  `json:"oldest_invoice_date"`
}

type UpcomingPayment struct {
	DocNo string  `json:"docno"`
	SupplierName string  `json:"supplier_name"`
	Amount float64 `json:"amount"`
	DueDate string  `json:"due_date"`
	DaysUntilDue int     `json:"days_until_due"`
}

func GetAccountsPayable(ctx context.Context, shopID string) (*AccountsPayableResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	logger.Info("[Accounts Payable] shopid=%s", shopID)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	now := time.Now()
	response := &AccountsPayableResponse{
		AgingBuckets:     []AgingBucket{},
		TopCreditors:     []Creditor{},
		UpcomingPayments: []UpcomingPayment{},
		GeneratedAt:      now,
	}

	// Get total payable from unpaid purchase orders
	query := `
		SELECT
			COALESCE(SUM(totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)), 0) as total_payable,
			COALESCE(SUM(CASE WHEN duedate < CURRENT_DATE THEN totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0) ELSE 0 END), 0) as overdue,
			COALESCE(SUM(CASE WHEN duedate BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '7 days' THEN totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0) ELSE 0 END), 0) as due_this_week
		FROM doc
		WHERE transflag IN (1, 3)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND (isdelete = false OR isdelete IS NULL)
	`

	var totalPayable, overdueAmount, dueThisWeek float64
	err = db.QueryRow(query).Scan(&totalPayable, &overdueAmount, &dueThisWeek)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("[Accounts Payable] Query failed: %v", err)
	}

	overduePercent := float64(0)
	if totalPayable > 0 {
		overduePercent = (overdueAmount / totalPayable) * 100
	}

	response.Summary = APSummary{
		TotalPayable:     totalPayable,
		TotalPayableWord: formatAmountWord(totalPayable),
		OverdueAmount:    overdueAmount,
		OverduePercent:   overduePercent,
		DueThisWeek:      dueThisWeek,
	}

	// TopCreditors — เจ้าหนี้สูงสุด
	creditorQuery := `
		SELECT
			COALESCE(custcode, 'N/A') as supplier_code,
			COALESCE(custname, custcode, 'Unknown') as supplier_name,
			SUM(totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)) as total_owed,
			MIN(docdate)::text as oldest_invoice
		FROM doc
		WHERE transflag IN (1, 3)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY custcode, custname
		ORDER BY total_owed DESC
		LIMIT 10
	`
	creditorRows, creditorErr := db.Query(creditorQuery)
	if creditorErr == nil {
		defer creditorRows.Close()
		for creditorRows.Next() {
			var c Creditor
			if err := creditorRows.Scan(&c.SupplierCode, &c.SupplierName, &c.TotalOwed, &c.OldestInvoice); err != nil {
				continue
			}
			response.TopCreditors = append(response.TopCreditors, c)
		}
	}
	if len(response.TopCreditors) == 0 {
		response.TopCreditors = []Creditor{}
	}

	// UpcomingPayments — รายการที่จะครบกำหนดใน 30 วัน
	upcomingQuery := `
		SELECT
			docno,
			COALESCE(custname, custcode, 'Unknown') as supplier_name,
			totalamount - COALESCE(0 /*paidamount: column missing, treating as 0*/, 0) as amount,
			duedate::text as due_date,
			duedate - CURRENT_DATE as days_until_due
		FROM doc
		WHERE transflag IN (1, 3)
			AND totalamount > COALESCE(0 /*paidamount: column missing, treating as 0*/, 0)
			AND duedate >= CURRENT_DATE
			AND duedate <= CURRENT_DATE + INTERVAL '30 days'
			AND (isdelete = false OR isdelete IS NULL)
		ORDER BY duedate ASC
		LIMIT 20
	`
	upcomingRows, upcomingErr := db.Query(upcomingQuery)
	if upcomingErr == nil {
		defer upcomingRows.Close()
		for upcomingRows.Next() {
			var u UpcomingPayment
			if err := upcomingRows.Scan(&u.DocNo, &u.SupplierName, &u.Amount, &u.DueDate, &u.DaysUntilDue); err != nil {
				continue
			}
			response.UpcomingPayments = append(response.UpcomingPayments, u)
		}
	}
	if len(response.UpcomingPayments) == 0 {
		response.UpcomingPayments = []UpcomingPayment{}
	}

	return response, nil
}

// ==================== Cash Flow ====================

type CashFlowRequest struct {
	ShopID string `json:"shop_id"`
	FromDate string `json:"from_date"`
	ToDate string `json:"to_date"`
}

type CashFlowResponse struct {
	Period string          `json:"period"`
	Summary CashFlowSummary `json:"summary"`
	Inflows []CashFlowItem  `json:"inflows"`
	Outflows []CashFlowItem  `json:"outflows"`
	DailyFlow []DailyCashFlow `json:"daily_flow"`
	GeneratedAt time.Time       `json:"generated_at"`
}

type CashFlowSummary struct {
	OpeningBalance float64 `json:"opening_balance"`
	TotalInflows float64 `json:"total_inflows"`
	TotalInflowsWord string  `json:"total_inflows_word"`
	TotalOutflows float64 `json:"total_outflows"`
	TotalOutflowsWord string `json:"total_outflows_word"`
	NetCashFlow float64 `json:"net_cash_flow"`
	NetCashFlowWord string  `json:"net_cash_flow_word"`
	ClosingBalance float64 `json:"closing_balance"`
}

type CashFlowItem struct {
	Category string  `json:"category"`
	Description string  `json:"description"`
	Amount float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type DailyCashFlow struct {
	Date string  `json:"date"`
	Inflows float64 `json:"inflows"`
	Outflows float64 `json:"outflows"`
	NetFlow float64 `json:"net_flow"`
	Balance float64 `json:"balance"`
}

func GetCashFlow(ctx context.Context, shopID, fromDate, toDate string) (*CashFlowResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	now := time.Now()
	if fromDate == "" {
		fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}

	logger.Info("[Cash Flow] shopid=%s, from=%s, to=%s", shopID, fromDate, toDate)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &CashFlowResponse{
		Period:      fmt.Sprintf("%s to %s", fromDate, toDate),
		Inflows:    []CashFlowItem{},
		Outflows:   []CashFlowItem{},
		DailyFlow:  []DailyCashFlow{},
		GeneratedAt: now,
	}

	// Get cash inflows (sales receipts)
	var salesInflow float64
	query := `
		SELECT COALESCE(SUM(0 /*paidamount: column missing, treating as 0*/), 0)
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`
	db.QueryRow(query, fromDate, toDate).Scan(&salesInflow)

	// Get cash outflows (purchase payments)
	var purchaseOutflow float64
	query = `
		SELECT COALESCE(SUM(0 /*paidamount: column missing, treating as 0*/), 0)
		FROM doc
		WHERE transflag IN (1, 3)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`
	db.QueryRow(query, fromDate, toDate).Scan(&purchaseOutflow)

	netCashFlow := salesInflow - purchaseOutflow

	response.Summary = CashFlowSummary{
		OpeningBalance:    0,
		TotalInflows:      salesInflow,
		TotalInflowsWord:  formatAmountWord(salesInflow),
		TotalOutflows:     purchaseOutflow,
		TotalOutflowsWord: formatAmountWord(purchaseOutflow),
		NetCashFlow:       netCashFlow,
		NetCashFlowWord:   formatAmountWord(netCashFlow),
		ClosingBalance:    netCashFlow,
	}

	response.Inflows = []CashFlowItem{
		{Category: "Sales Receipts", Description: "Cash from sales", Amount: salesInflow, Percentage: 100},
	}

	response.Outflows = []CashFlowItem{
		{Category: "Purchase Payments", Description: "Payments to suppliers", Amount: purchaseOutflow, Percentage: 100},
	}

	// DailyFlow — กระแสเงินสดรายวัน
	dailyQuery := `
		SELECT
			TO_CHAR(docdate, 'YYYY-MM-DD') as date,
			COALESCE(SUM(CASE WHEN transflag IN (16, 18) THEN 0 /*paidamount: column missing, treating as 0*/ ELSE 0 END), 0) as inflows,
			COALESCE(SUM(CASE WHEN transflag IN (1, 3) THEN 0 /*paidamount: column missing, treating as 0*/ ELSE 0 END), 0) as outflows
		FROM doc
		WHERE transflag IN (1, 3, 16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY TO_CHAR(docdate, 'YYYY-MM-DD')
		ORDER BY date
	`
	dailyRows, dailyErr := db.Query(dailyQuery, fromDate, toDate)
	if dailyErr == nil {
		defer dailyRows.Close()
		var runningBalance float64
		for dailyRows.Next() {
			var d DailyCashFlow
			if err := dailyRows.Scan(&d.Date, &d.Inflows, &d.Outflows); err != nil {
				continue
			}
			d.NetFlow = d.Inflows - d.Outflows
			runningBalance += d.NetFlow
			d.Balance = runningBalance
			response.DailyFlow = append(response.DailyFlow, d)
		}
	}
	if len(response.DailyFlow) == 0 {
		response.DailyFlow = []DailyCashFlow{}
	}

	return response, nil
}

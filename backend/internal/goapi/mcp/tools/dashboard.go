package tools

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"
)

// DashboardKPIsRequest represents the request for dashboard KPIs
type DashboardKPIsRequest struct {
	ShopID string `json:"shop_id"`
	Period string `json:"period"` // today, this_week, this_month, this_year
}

// DashboardKPIsResponse represents the dashboard KPIs
type DashboardKPIsResponse struct {
	Period           string             `json:"period"`
	DateRange        DateRangeInfo      `json:"date_range"`
	Sales            SalesKPI           `json:"sales"`
	Orders           OrdersKPI          `json:"orders"`
	Profit           ProfitKPI          `json:"profit"`
	Customers        CustomersKPI       `json:"customers"`
	Inventory        InventoryKPI       `json:"inventory"`
	TopProducts      []TopProductKPI    `json:"top_products"`
	SalesTrend       []SalesTrendPoint  `json:"sales_trend"`
	GeneratedAt      time.Time          `json:"generated_at"`
}

type DateRangeInfo struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SalesKPI struct {
	TotalAmount      float64 `json:"total_amount"`
	TotalAmountWord  string  `json:"total_amount_word"`
	GrowthPercent    float64 `json:"growth_percent"`
	GrowthDirection  string  `json:"growth_direction"` // up, down, stable
	PreviousAmount   float64 `json:"previous_amount"`
}

type OrdersKPI struct {
	TotalCount       int     `json:"total_count"`
	AverageValue     float64 `json:"average_value"`
	GrowthPercent    float64 `json:"growth_percent"`
	GrowthDirection  string  `json:"growth_direction"`
}

type ProfitKPI struct {
	GrossProfit      float64 `json:"gross_profit"`
	GrossProfitWord  string  `json:"gross_profit_word"`
	GrossMargin      float64 `json:"gross_margin_percent"`
	NetProfit        float64 `json:"net_profit"`
	NetMargin        float64 `json:"net_margin_percent"`
}

type CustomersKPI struct {
	TotalActive      int     `json:"total_active"`
	NewCustomers     int     `json:"new_customers"`
	ReturningRate    float64 `json:"returning_rate_percent"`
}

type InventoryKPI struct {
	TotalValue       float64 `json:"total_value"`
	TotalValueWord   string  `json:"total_value_word"`
	LowStockCount    int     `json:"low_stock_count"`
	OutOfStockCount  int     `json:"out_of_stock_count"`
}

type TopProductKPI struct {
	ItemCode    string  `json:"itemcode"`
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity"`
	Amount      float64 `json:"amount"`
}

type SalesTrendPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

// GetDashboardKPIs returns comprehensive dashboard KPIs
func GetDashboardKPIs(ctx context.Context, shopID, period string) (*DashboardKPIsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	// Default period
	if period == "" {
		period = "this_month"
	}

	// Calculate date range based on period
	now := time.Now()
	var fromDate, toDate time.Time
	var prevFromDate, prevToDate time.Time

	switch period {
	case "today":
		fromDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		toDate = now
		prevFromDate = fromDate.AddDate(0, 0, -1)
		prevToDate = fromDate.Add(-time.Second)
	case "this_week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		fromDate = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		toDate = now
		prevFromDate = fromDate.AddDate(0, 0, -7)
		prevToDate = fromDate.Add(-time.Second)
	case "this_month":
		fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		toDate = now
		prevFromDate = fromDate.AddDate(0, -1, 0)
		prevToDate = fromDate.Add(-time.Second)
	case "this_year":
		fromDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		toDate = now
		prevFromDate = fromDate.AddDate(-1, 0, 0)
		prevToDate = fromDate.Add(-time.Second)
	default:
		fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		toDate = now
		prevFromDate = fromDate.AddDate(0, -1, 0)
		prevToDate = fromDate.Add(-time.Second)
	}

	logger.Info("[Dashboard KPIs] shopid=%s, period=%s, from=%s, to=%s",
		shopID, period, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &DashboardKPIsResponse{
		Period: period,
		DateRange: DateRangeInfo{
			From: fromDate.Format("2006-01-02"),
			To:   toDate.Format("2006-01-02"),
		},
		GeneratedAt: now,
	}

	// Get current period sales
	currentSales, currentOrders, currentProfit, err := getSalesSummary(db, fromDate, toDate)
	if err != nil {
		logger.Error("[Dashboard KPIs] Failed to get current sales: %v", err)
	}

	// Get previous period sales for comparison
	prevSales, prevOrders, _, err := getSalesSummary(db, prevFromDate, prevToDate)
	if err != nil {
		logger.Error("[Dashboard KPIs] Failed to get previous sales: %v", err)
	}

	// Calculate growth
	salesGrowth := calculateGrowth(currentSales, prevSales)
	ordersGrowth := calculateGrowth(float64(currentOrders), float64(prevOrders))

	response.Sales = SalesKPI{
		TotalAmount:     currentSales,
		TotalAmountWord: formatAmountWord(currentSales),
		GrowthPercent:   salesGrowth,
		GrowthDirection: getGrowthDirection(salesGrowth),
		PreviousAmount:  prevSales,
	}

	avgOrderValue := float64(0)
	if currentOrders > 0 {
		avgOrderValue = currentSales / float64(currentOrders)
	}

	response.Orders = OrdersKPI{
		TotalCount:      currentOrders,
		AverageValue:    avgOrderValue,
		GrowthPercent:   ordersGrowth,
		GrowthDirection: getGrowthDirection(ordersGrowth),
	}

	response.Profit = ProfitKPI{
		GrossProfit:     currentProfit,
		GrossProfitWord: formatAmountWord(currentProfit),
		GrossMargin:     calculateMargin(currentProfit, currentSales),
		NetProfit:       currentProfit * 0.8, // Estimated net (80% of gross)
		NetMargin:       calculateMargin(currentProfit*0.8, currentSales),
	}

	// Get customer stats
	newCustomers, err := getNewCustomersCount(db, fromDate, toDate)
	if err != nil {
		logger.Error("[Dashboard KPIs] Failed to get customer stats: %v", err)
	}

	response.Customers = CustomersKPI{
		TotalActive:   0, // Would need customer activity tracking
		NewCustomers:  newCustomers,
		ReturningRate: 0,
	}

	// Get inventory stats
	invValue, lowStock, outOfStock, err := getInventoryStats(db)
	if err != nil {
		logger.Error("[Dashboard KPIs] Failed to get inventory stats: %v", err)
	}

	response.Inventory = InventoryKPI{
		TotalValue:      invValue,
		TotalValueWord:  formatAmountWord(invValue),
		LowStockCount:   lowStock,
		OutOfStockCount: outOfStock,
	}

	// Get top products
	topProducts, err := getTopProductsKPI(db, fromDate, toDate, 5)
	if err != nil {
		logger.Error("[Dashboard KPIs] Failed to get top products: %v", err)
	}
	response.TopProducts = topProducts

	// Get sales trend
	trend, err := getSalesTrend(db, fromDate, toDate, period)
	if err != nil {
		logger.Error("[Dashboard KPIs] Failed to get sales trend: %v", err)
	}
	response.SalesTrend = trend

	return response, nil
}

// BusinessHealthRequest represents the request for business health
type BusinessHealthRequest struct {
	ShopID string `json:"shop_id"`
}

// BusinessHealthResponse represents overall business health
type BusinessHealthResponse struct {
	OverallScore     int                `json:"overall_score"`      // 0-100
	OverallStatus    string             `json:"overall_status"`     // excellent, good, fair, poor
	Metrics          []HealthMetric     `json:"metrics"`
	Recommendations  []string           `json:"recommendations"`
	GeneratedAt      time.Time          `json:"generated_at"`
}

type HealthMetric struct {
	Name        string  `json:"name"`
	Score       int     `json:"score"`       // 0-100
	Status      string  `json:"status"`      // excellent, good, fair, poor
	Value       string  `json:"value"`
	Target      string  `json:"target"`
	Description string  `json:"description"`
}

// GetBusinessHealth returns overall business health assessment
func GetBusinessHealth(ctx context.Context, shopID string) (*BusinessHealthResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonth := thisMonth.AddDate(0, -1, 0)

	response := &BusinessHealthResponse{
		Metrics:       []HealthMetric{},
		GeneratedAt:   now,
	}

	var totalScore int
	var metricCount int

	// 1. Sales Growth Metric
	currentSales, _, _, _ := getSalesSummary(db, thisMonth, now)
	prevSales, _, _, _ := getSalesSummary(db, lastMonth, thisMonth.Add(-time.Second))
	salesGrowth := calculateGrowth(currentSales, prevSales)
	salesScore := calculateHealthScore(salesGrowth, -10, 20)

	response.Metrics = append(response.Metrics, HealthMetric{
		Name:        "Sales Growth",
		Score:       salesScore,
		Status:      getHealthStatus(salesScore),
		Value:       fmt.Sprintf("%.1f%%", salesGrowth),
		Target:      ">10%",
		Description: "Month-over-month sales growth",
	})
	totalScore += salesScore
	metricCount++

	// 2. Profit Margin Metric
	_, _, profit, _ := getSalesSummary(db, thisMonth, now)
	margin := calculateMargin(profit, currentSales)
	marginScore := calculateHealthScore(margin, 10, 40)

	response.Metrics = append(response.Metrics, HealthMetric{
		Name:        "Profit Margin",
		Score:       marginScore,
		Status:      getHealthStatus(marginScore),
		Value:       fmt.Sprintf("%.1f%%", margin),
		Target:      ">25%",
		Description: "Gross profit margin",
	})
	totalScore += marginScore
	metricCount++

	// 3. Inventory Health
	_, lowStock, outOfStock, _ := getInventoryStats(db)
	invScore := 100
	if outOfStock > 10 {
		invScore -= 30
	}
	if lowStock > 20 {
		invScore -= 20
	}
	if invScore < 0 {
		invScore = 0
	}

	response.Metrics = append(response.Metrics, HealthMetric{
		Name:        "Inventory Health",
		Score:       invScore,
		Status:      getHealthStatus(invScore),
		Value:       fmt.Sprintf("%d low, %d out", lowStock, outOfStock),
		Target:      "<5 out of stock",
		Description: "Stock availability status",
	})
	totalScore += invScore
	metricCount++

	// 4. Order Frequency
	_, orderCount, _, _ := getSalesSummary(db, thisMonth, now)
	daysInPeriod := now.Sub(thisMonth).Hours() / 24
	ordersPerDay := float64(orderCount) / daysInPeriod
	orderScore := calculateHealthScore(ordersPerDay, 1, 10)

	response.Metrics = append(response.Metrics, HealthMetric{
		Name:        "Order Frequency",
		Score:       orderScore,
		Status:      getHealthStatus(orderScore),
		Value:       fmt.Sprintf("%.1f/day", ordersPerDay),
		Target:      ">5/day",
		Description: "Average orders per day",
	})
	totalScore += orderScore
	metricCount++

	// Calculate overall
	if metricCount > 0 {
		response.OverallScore = totalScore / metricCount
	}
	response.OverallStatus = getHealthStatus(response.OverallScore)

	// Generate recommendations
	response.Recommendations = generateRecommendations(response.Metrics)

	return response, nil
}

// Helper functions
func getSalesSummary(db interface{}, from, to time.Time) (totalAmount float64, orderCount int, profit float64, err error) {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid database connection")
	}

	query := `
		SELECT
			COALESCE(SUM(totalamount), 0) as total_amount,
			COUNT(DISTINCT docno) as order_count,
			COALESCE(SUM(totalamount - COALESCE(totalcost, 0)), 0) as profit
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`

	row := sqlDB.QueryRow(query, from, to)
	err = row.Scan(&totalAmount, &orderCount, &profit)
	return
}

func getNewCustomersCount(db interface{}, from, to time.Time) (int, error) {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return 0, fmt.Errorf("invalid database connection")
	}

	// นับลูกค้าที่ซื้อครั้งแรกในช่วงเวลานี้
	query := `
		WITH first_purchase AS (
			SELECT custcode, MIN(docdate) as first_date
			FROM doc
			WHERE transflag IN (16, 18)
				AND custcode IS NOT NULL AND custcode != ''
				AND (isdelete = false OR isdelete IS NULL)
			GROUP BY custcode
		)
		SELECT COUNT(*)
		FROM first_purchase
		WHERE first_date >= $1 AND first_date <= $2
	`
	var count int
	err := sqlDB.QueryRow(query, from, to).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getInventoryStats(db interface{}) (totalValue float64, lowStock, outOfStock int, err error) {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid database connection")
	}

	// มูลค่าสินค้าคงคลังรวม
	query := `
		SELECT COALESCE(SUM(ABS(balance_qty) * COALESCE(avgcost, 0)), 0) as total_value
		FROM (
			SELECT
				d.itemcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty,
				(SELECT COALESCE(avgcost, 0) FROM productbarcode pb WHERE pb.itemcode = d.itemcode LIMIT 1) as avgcost
			FROM docdetail d
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY d.itemcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) != 0
		) inv
	`
	_ = sqlDB.QueryRow(query).Scan(&totalValue)

	// นับสินค้า low stock (1-10) และ out of stock (<=0)
	query = `
		SELECT
			COUNT(CASE WHEN balance_qty <= 0 THEN 1 END) as out_of_stock,
			COUNT(CASE WHEN balance_qty > 0 AND balance_qty <= 10 THEN 1 END) as low_stock
		FROM (
			SELECT
				itemcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty
			FROM docdetail
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY itemcode
		) inv
	`
	_ = sqlDB.QueryRow(query).Scan(&outOfStock, &lowStock)

	return totalValue, lowStock, outOfStock, nil
}

func getTopProductsKPI(db interface{}, from, to time.Time, limit int) ([]TopProductKPI, error) {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return []TopProductKPI{}, fmt.Errorf("invalid database connection")
	}

	query := `
		SELECT
			d.itemcode,
			COALESCE(pb.name0, d.itemcode) as name,
			SUM((d.totalqty * d.calcflag * -1) * d.unitstand / NULLIF(d.unitdivide, 0)) as qty,
			SUM(d.totalqty * d.price * d.calcflag * -1 * d.unitstand / NULLIF(d.unitdivide, 0)) as amount
		FROM docdetail d
		LEFT JOIN (SELECT DISTINCT ON (itemcode) itemcode, name0 FROM productbarcode) pb ON pb.itemcode = d.itemcode
		WHERE d.transflag IN (16, 18)
			AND d.docdate >= $1 AND d.docdate <= $2
		GROUP BY d.itemcode, pb.name0
		ORDER BY amount DESC
		LIMIT $3
	`

	rows, err := sqlDB.Query(query, from, to, limit)
	if err != nil {
		return []TopProductKPI{}, err
	}
	defer rows.Close()

	var results []TopProductKPI
	for rows.Next() {
		var item TopProductKPI
		if err := rows.Scan(&item.ItemCode, &item.Name, &item.Quantity, &item.Amount); err != nil {
			continue
		}
		results = append(results, item)
	}

	if results == nil {
		results = []TopProductKPI{}
	}
	return results, nil
}

func getSalesTrend(db interface{}, from, to time.Time, period string) ([]SalesTrendPoint, error) {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return []SalesTrendPoint{}, fmt.Errorf("invalid database connection")
	}

	// เลือก group by ตาม period
	var dateExpr string
	switch period {
	case "today":
		dateExpr = "TO_CHAR(docdate, 'HH24:00')"
	case "this_week", "this_month":
		dateExpr = "TO_CHAR(docdate, 'YYYY-MM-DD')"
	case "this_year":
		dateExpr = "TO_CHAR(docdate, 'YYYY-MM')"
	default:
		dateExpr = "TO_CHAR(docdate, 'YYYY-MM-DD')"
	}

	query := fmt.Sprintf(`
		SELECT
			%s as period_date,
			COALESCE(SUM(totalamount), 0) as amount
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY period_date
		ORDER BY period_date
	`, dateExpr)

	rows, err := sqlDB.Query(query, from, to)
	if err != nil {
		return []SalesTrendPoint{}, err
	}
	defer rows.Close()

	var results []SalesTrendPoint
	for rows.Next() {
		var point SalesTrendPoint
		if err := rows.Scan(&point.Date, &point.Amount); err != nil {
			continue
		}
		results = append(results, point)
	}

	if results == nil {
		results = []SalesTrendPoint{}
	}
	return results, nil
}

func calculateGrowth(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return ((current - previous) / previous) * 100
}

func calculateMargin(profit, sales float64) float64 {
	if sales == 0 {
		return 0
	}
	return (profit / sales) * 100
}

func getGrowthDirection(growth float64) string {
	if growth > 1 {
		return "up"
	} else if growth < -1 {
		return "down"
	}
	return "stable"
}

func formatAmountWord(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.2f ล้านบาท", amount/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.2f พันบาท", amount/1000)
	}
	return fmt.Sprintf("%.2f บาท", amount)
}

func calculateHealthScore(value, minGood, maxGood float64) int {
	if value >= maxGood {
		return 100
	} else if value >= minGood {
		return int(50 + ((value-minGood)/(maxGood-minGood))*50)
	} else if value >= 0 {
		return int((value / minGood) * 50)
	}
	return 0
}

func getHealthStatus(score int) string {
	switch {
	case score >= 80:
		return "excellent"
	case score >= 60:
		return "good"
	case score >= 40:
		return "fair"
	default:
		return "poor"
	}
}

func generateRecommendations(metrics []HealthMetric) []string {
	var recs []string
	for _, m := range metrics {
		if m.Score < 60 {
			switch m.Name {
			case "Sales Growth":
				recs = append(recs, "Consider running promotions to boost sales")
			case "Profit Margin":
				recs = append(recs, "Review pricing strategy to improve margins")
			case "Inventory Health":
				recs = append(recs, "Restock low inventory items urgently")
			case "Order Frequency":
				recs = append(recs, "Focus on customer acquisition and retention")
			}
		}
	}
	if len(recs) == 0 {
		recs = append(recs, "Business is performing well! Keep up the good work.")
	}
	return recs
}

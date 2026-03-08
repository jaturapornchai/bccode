package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/myclickhouse"
)

// SalesTool handles sales-related queries using ClickHouse for fast performance
type SalesTool struct{}

// NewSalesTool creates a new sales tool
func NewSalesTool() *SalesTool {
	return &SalesTool{}
}

// DailySalesRequest represents a request for daily sales
type DailySalesRequest struct {
	ShopID     string `json:"shop_id"`
	Date       string `json:"date"` // Format: YYYY-MM-DD
	BranchCode string `json:"branch_code,omitempty"`
}

// DailySalesResponse represents the response for daily sales
type DailySalesResponse struct {
	Date          string  `json:"date"`
	TotalAmount   float64 `json:"total_amount"`
	TotalCost     float64 `json:"total_cost"`
	TotalProfit   float64 `json:"total_profit"`
	TotalQty      float64 `json:"total_qty"`
	DocumentCount int64   `json:"document_count"`
	BranchCode    string  `json:"branch_code,omitempty"`
}

// GetDailySales retrieves daily sales data from ClickHouse
func (st *SalesTool) GetDailySales(ctx context.Context, req DailySalesRequest) (*DailySalesResponse, error) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}

	startDate := date.Format("2006-01-02")
	endDate := date.AddDate(0, 0, 1).Format("2006-01-02")

	// Build query
	query := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT docno) as doc_count,
			SUM(totalqty * -1) as total_qty,
			SUM(totalqty * price * -1) as total_amount,
			SUM(calcamount * -1) as total_cost,
			SUM((totalqty * price * -1) - (calcamount * -1)) as total_profit
		FROM ` + myclickhouse.TableName("processstockcost") + `
		WHERE shopid = '%s'
			AND transflag = 44
			AND docdatetime >= '%s'
			AND docdatetime < '%s'`,
		req.ShopID, startDate, endDate)

	if req.BranchCode != "" {
		query += fmt.Sprintf(" AND whcode = '%s'", req.BranchCode)
	}

	results, err := myclickhouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily sales: %w", err)
	}

	if len(results) == 0 {
		return &DailySalesResponse{
			Date:          req.Date,
			TotalAmount:   0,
			TotalCost:     0,
			TotalProfit:   0,
			TotalQty:      0,
			DocumentCount: 0,
			BranchCode:    req.BranchCode,
		}, nil
	}

	row := results[0]
	return &DailySalesResponse{
		Date:          req.Date,
		TotalAmount:   parseFloat(row["total_amount"]),
		TotalCost:     parseFloat(row["total_cost"]),
		TotalProfit:   parseFloat(row["total_profit"]),
		TotalQty:      parseFloat(row["total_qty"]),
		DocumentCount: parseInt64(row["doc_count"]),
		BranchCode:    req.BranchCode,
	}, nil
}

// SalesByDateRangeRequest represents a request for sales by date range
type SalesByDateRangeRequest struct {
	ShopID     string `json:"shop_id"`
	FromDate   string `json:"from_date"` // Format: YYYY-MM-DD
	ToDate     string `json:"to_date"`   // Format: YYYY-MM-DD
	BranchCode string `json:"branch_code,omitempty"`
	GroupBy    string `json:"group_by,omitempty"` // "day", "week", "month"
}

// SalesByDateRangeResponse represents sales data grouped by date
type SalesByDateRangeResponse struct {
	FromDate string           `json:"from_date"`
	ToDate   string           `json:"to_date"`
	GroupBy  string           `json:"group_by"`
	Data     []SalesGroupData `json:"data"`
	Summary  SalesSummary     `json:"summary"`
}

// SalesGroupData represents sales data for a specific period
type SalesGroupData struct {
	Period      string  `json:"period"`
	TotalAmount float64 `json:"total_amount"`
	TotalCost   float64 `json:"total_cost"`
	TotalProfit float64 `json:"total_profit"`
	TotalQty    float64 `json:"total_qty"`
	DocCount    int64   `json:"doc_count"`
}

// SalesSummary represents overall summary
type SalesSummary struct {
	TotalAmount   float64 `json:"total_amount"`
	TotalCost     float64 `json:"total_cost"`
	TotalProfit   float64 `json:"total_profit"`
	TotalQty      float64 `json:"total_qty"`
	DocumentCount int64   `json:"document_count"`
	AvgDailySales float64 `json:"avg_daily_sales"`
}

// GetSalesByDateRange retrieves sales data for a date range from ClickHouse
func (st *SalesTool) GetSalesByDateRange(ctx context.Context, req SalesByDateRangeRequest) (*SalesByDateRangeResponse, error) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from_date format: %w", err)
	}
	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to_date format: %w", err)
	}

	// Set default group by
	if req.GroupBy == "" {
		req.GroupBy = "day"
	}

	// Build group by clause
	var groupByClause string
	switch req.GroupBy {
	case "week":
		groupByClause = "toStartOfWeek(docdatetime)"
	case "month":
		groupByClause = "toStartOfMonth(docdatetime)"
	default: // day
		groupByClause = "toDate(docdatetime)"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as period,
			COUNT(DISTINCT docno) as doc_count,
			SUM(totalqty * -1) as total_qty,
			SUM(totalqty * price * -1) as total_amount,
			SUM(calcamount * -1) as total_cost,
			SUM((totalqty * price * -1) - (calcamount * -1)) as total_profit
		FROM ` + myclickhouse.TableName("processstockcost") + `
		WHERE shopid = '%s'
			AND transflag = 44
			AND docdatetime >= '%s'
			AND docdatetime < '%s'`,
		groupByClause, req.ShopID,
		fromDate.Format("2006-01-02"),
		toDate.AddDate(0, 0, 1).Format("2006-01-02"))

	if req.BranchCode != "" {
		query += fmt.Sprintf(" AND whcode = '%s'", req.BranchCode)
	}

	query += fmt.Sprintf(`
		GROUP BY period
		ORDER BY period ASC`)

	results, err := myclickhouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales by date range: %w", err)
	}

	var data []SalesGroupData
	var summary SalesSummary

	for _, row := range results {
		salesData := SalesGroupData{
			Period:      parseString(row["period"]),
			DocCount:    parseInt64(row["doc_count"]),
			TotalQty:    parseFloat(row["total_qty"]),
			TotalAmount: parseFloat(row["total_amount"]),
			TotalCost:   parseFloat(row["total_cost"]),
			TotalProfit: parseFloat(row["total_profit"]),
		}
		data = append(data, salesData)

		// Update summary
		summary.TotalAmount += salesData.TotalAmount
		summary.TotalCost += salesData.TotalCost
		summary.TotalProfit += salesData.TotalProfit
		summary.TotalQty += salesData.TotalQty
		summary.DocumentCount += salesData.DocCount
	}

	// Calculate average daily sales
	days := int(toDate.Sub(fromDate).Hours()/24) + 1
	if days > 0 {
		summary.AvgDailySales = summary.TotalAmount / float64(days)
	}

	return &SalesByDateRangeResponse{
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
		GroupBy:  req.GroupBy,
		Data:     data,
		Summary:  summary,
	}, nil
}

// TopSellingProductsRequest represents a request for top selling products
type TopSellingProductsRequest struct {
	ShopID     string `json:"shop_id"`
	FromDate   string `json:"from_date"` // Format: YYYY-MM-DD
	ToDate     string `json:"to_date"`   // Format: YYYY-MM-DD
	Limit      int    `json:"limit"`     // Default: 10
	BranchCode string `json:"branch_code,omitempty"`
}

// TopSellingProduct represents a top selling product
type TopSellingProduct struct {
	ItemCode    string  `json:"item_code"`
	ItemName    string  `json:"item_name"`
	Barcode     string  `json:"barcode"`
	TotalQty    float64 `json:"total_qty"`
	TotalAmount float64 `json:"total_amount"`
	TotalCost   float64 `json:"total_cost"`
	TotalProfit float64 `json:"total_profit"`
	AvgPrice    float64 `json:"avg_price"`
}

// TopSellingProductsResponse represents the response for top selling products
type TopSellingProductsResponse struct {
	FromDate   string              `json:"from_date"`
	ToDate     string              `json:"to_date"`
	Limit      int                 `json:"limit"`
	Products   []TopSellingProduct `json:"products"`
	TotalCount int                 `json:"total_count"`
}

// GetTopSellingProducts retrieves top selling products from ClickHouse
func (st *SalesTool) GetTopSellingProducts(ctx context.Context, req TopSellingProductsRequest) (*TopSellingProductsResponse, error) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Set default limit
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}

	// Parse dates
	fromDate, _ := time.Parse("2006-01-02", req.FromDate)
	toDate, _ := time.Parse("2006-01-02", req.ToDate)

	query := fmt.Sprintf(`
		SELECT 
			p.itemcode,
			MAX(pb.name0) as itemname,
			MAX(p.barcode) as barcode,
			SUM(p.totalqty * -1) as total_qty,
			SUM(p.totalqty * p.price * -1) as total_amount,
			SUM(p.calcamount * -1) as total_cost,
			SUM((p.totalqty * p.price * -1) - (p.calcamount * -1)) as total_profit,
			AVG(p.price) as avg_price
		FROM ` + myclickhouse.TableName("processstockcost") + ` p
		LEFT JOIN ` + myclickhouse.TableName("productbarcode") + ` pb ON p.shopid = pb.shopid AND p.itemcode = pb.itemcode
		WHERE p.shopid = '%s'
			AND p.transflag = 44
			AND p.docdatetime >= '%s'
			AND p.docdatetime < '%s'`,
		req.ShopID,
		fromDate.Format("2006-01-02"),
		toDate.AddDate(0, 0, 1).Format("2006-01-02"))

	if req.BranchCode != "" {
		query += fmt.Sprintf(" AND p.whcode = '%s'", req.BranchCode)
	}

	query += fmt.Sprintf(`
		GROUP BY p.itemcode
		ORDER BY total_amount DESC
		LIMIT %d`, req.Limit)

	results, err := myclickhouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query top selling products: %w", err)
	}

	var products []TopSellingProduct
	for _, row := range results {
		products = append(products, TopSellingProduct{
			ItemCode:    parseString(row["itemcode"]),
			ItemName:    parseString(row["itemname"]),
			Barcode:     parseString(row["barcode"]),
			TotalQty:    parseFloat(row["total_qty"]),
			TotalAmount: parseFloat(row["total_amount"]),
			TotalCost:   parseFloat(row["total_cost"]),
			TotalProfit: parseFloat(row["total_profit"]),
			AvgPrice:    parseFloat(row["avg_price"]),
		})
	}

	return &TopSellingProductsResponse{
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		Limit:      req.Limit,
		Products:   products,
		TotalCount: len(products),
	}, nil
}

// SalesBySellerRequest represents a request for sales by seller
type SalesBySellerRequest struct {
	ShopID   string `json:"shop_id"`
	FromDate string `json:"from_date"` // Format: YYYY-MM-DD
	ToDate   string `json:"to_date"`   // Format: YYYY-MM-DD
}

// SellerSales represents sales data for a seller
type SellerSales struct {
	SellerCode  string  `json:"seller_code"`
	SellerName  string  `json:"seller_name"`
	TotalAmount float64 `json:"total_amount"`
	TotalQty    float64 `json:"total_qty"`
	DocCount    int64   `json:"doc_count"`
	AvgDocValue float64 `json:"avg_doc_value"`
}

// SalesBySellerResponse represents the response for sales by seller
type SalesBySellerResponse struct {
	FromDate   string        `json:"from_date"`
	ToDate     string        `json:"to_date"`
	Sellers    []SellerSales `json:"sellers"`
	TotalCount int           `json:"total_count"`
}

// GetSalesBySeller retrieves sales data grouped by sale channel from ClickHouse
// Note: ClickHouse doc table ไม่มี salescode/salesname — ใช้ salechannelcode + JOIN salechannel แทน
func (st *SalesTool) GetSalesBySeller(ctx context.Context, req SalesBySellerRequest) (*SalesBySellerResponse, error) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Parse dates
	fromDate, _ := time.Parse("2006-01-02", req.FromDate)
	toDate, _ := time.Parse("2006-01-02", req.ToDate)

	query := fmt.Sprintf(`
		SELECT
			d.salechannelcode as seller_code,
			COALESCE(MAX(sc.name), d.salechannelcode) as seller_name,
			COUNT(DISTINCT p.docno) as doc_count,
			SUM(p.totalqty * -1) as total_qty,
			SUM(p.totalqty * p.price * -1) as total_amount
		FROM `+myclickhouse.TableName("processstockcost")+` p
		INNER JOIN `+myclickhouse.TableName("doc")+` d ON p.shopid = d.shopid AND p.docno = d.docno
		LEFT JOIN `+myclickhouse.TableName("salechannel")+` sc ON d.shopid = sc.shopid AND d.salechannelcode = sc.code
		WHERE p.shopid = '%s'
			AND p.transflag = 44
			AND p.docdatetime >= '%s'
			AND p.docdatetime < '%s'
		GROUP BY d.salechannelcode
		ORDER BY total_amount DESC`,
		req.ShopID,
		fromDate.Format("2006-01-02"),
		toDate.AddDate(0, 0, 1).Format("2006-01-02"))

	results, err := myclickhouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales by seller: %w", err)
	}

	var sellers []SellerSales
	for _, row := range results {
		docCount := parseInt64(row["doc_count"])
		totalAmount := parseFloat(row["total_amount"])

		avgDocValue := 0.0
		if docCount > 0 {
			avgDocValue = totalAmount / float64(docCount)
		}

		sellers = append(sellers, SellerSales{
			SellerCode:  parseString(row["seller_code"]),
			SellerName:  parseString(row["seller_name"]),
			TotalAmount: totalAmount,
			TotalQty:    parseFloat(row["total_qty"]),
			DocCount:    docCount,
			AvgDocValue: avgDocValue,
		})
	}

	return &SalesBySellerResponse{
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		Sellers:    sellers,
		TotalCount: len(sellers),
	}, nil
}

// MonthlySummaryRequest represents a request for monthly summary
type MonthlySummaryRequest struct {
	ShopID string `json:"shop_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"` // 1-12
}

// MonthlySummaryResponse represents the response for monthly summary
type MonthlySummaryResponse struct {
	Year          int                `json:"year"`
	Month         int                `json:"month"`
	MonthName     string             `json:"month_name"`
	TotalAmount   float64            `json:"total_amount"`
	TotalCost     float64            `json:"total_cost"`
	TotalProfit   float64            `json:"total_profit"`
	TotalQty      float64            `json:"total_qty"`
	DocumentCount int64              `json:"document_count"`
	DailyAverage  float64            `json:"daily_average"`
	ProfitMargin  float64            `json:"profit_margin"`
	Comparison    *MonthlyComparison `json:"comparison,omitempty"`
}

// MonthlyComparison compares with previous month
type MonthlyComparison struct {
	PrevMonthAmount     float64 `json:"prev_month_amount"`
	AmountChange        float64 `json:"amount_change"`
	AmountChangePercent float64 `json:"amount_change_percent"`
}

// GetMonthlySummary retrieves monthly sales summary from ClickHouse
func (st *SalesTool) GetMonthlySummary(ctx context.Context, req MonthlySummaryRequest) (*MonthlySummaryResponse, error) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Default to current year/month if not provided
	now := time.Now()
	year := req.Year
	month := req.Month
	if year <= 0 {
		year = now.Year()
	}
	if month <= 0 || month > 12 {
		month = int(now.Month())
	}

	// Calculate date range
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	// Get current month data
	query := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT docno) as doc_count,
			SUM(totalqty * -1) as total_qty,
			SUM(totalqty * price * -1) as total_amount,
			SUM(calcamount * -1) as total_cost,
			SUM((totalqty * price * -1) - (calcamount * -1)) as total_profit
		FROM ` + myclickhouse.TableName("processstockcost") + `
		WHERE shopid = '%s'
			AND transflag = 44
			AND docdatetime >= '%s'
			AND docdatetime < '%s'`,
		req.ShopID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	results, err := myclickhouse.QuerySelectAll(conn, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly summary: %w", err)
	}

	monthNames := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	var summary MonthlySummaryResponse
	summary.Year = year
	summary.Month = month
	summary.MonthName = monthNames[month]

	if len(results) > 0 {
		row := results[0]
		summary.TotalAmount = parseFloat(row["total_amount"])
		summary.TotalCost = parseFloat(row["total_cost"])
		summary.TotalProfit = parseFloat(row["total_profit"])
		summary.TotalQty = parseFloat(row["total_qty"])
		summary.DocumentCount = parseInt64(row["doc_count"])
	}

	// Calculate daily average
	daysInMonth := endDate.AddDate(0, 0, -1).Day()
	if daysInMonth > 0 {
		summary.DailyAverage = summary.TotalAmount / float64(daysInMonth)
	}

	// Calculate profit margin
	if summary.TotalAmount > 0 {
		summary.ProfitMargin = (summary.TotalProfit / summary.TotalAmount) * 100
	}

	// Get previous month data for comparison
	prevMonthQuery := fmt.Sprintf(`
		SELECT 
			SUM(totalqty * price * -1) as total_amount
		FROM ` + myclickhouse.TableName("processstockcost") + `
		WHERE shopid = '%s'
			AND transflag = 44
			AND docdatetime >= '%s'
			AND docdatetime < '%s'`,
		req.ShopID,
		startDate.AddDate(0, -1, 0).Format("2006-01-02"),
		startDate.Format("2006-01-02"))

	prevResults, err := myclickhouse.QuerySelectAll(conn, prevMonthQuery)
	if err == nil && len(prevResults) > 0 {
		prevAmount := parseFloat(prevResults[0]["total_amount"])
		change := summary.TotalAmount - prevAmount
		changePercent := 0.0
		if prevAmount > 0 {
			changePercent = (change / prevAmount) * 100
		}

		summary.Comparison = &MonthlyComparison{
			PrevMonthAmount:     prevAmount,
			AmountChange:        change,
			AmountChangePercent: changePercent,
		}
	}

	return &summary, nil
}

// Helper functions for parsing ClickHouse results
func parseFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func parseInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	}
	return 0
}

func parseString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// sanitizeSQL sanitizes SQL input to prevent injection
func sanitizeSQL(input string) string {
	// Remove potentially dangerous characters
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "--", "")
	return input
}

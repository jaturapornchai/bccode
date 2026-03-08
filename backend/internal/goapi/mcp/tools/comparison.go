package tools

import (
	"context"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ==================== Year-over-Year Comparison ====================

type YoYComparisonRequest struct {
	ShopID string `json:"shop_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"` // Optional - if provided, compare specific month
}

type YoYComparisonResponse struct {
	Period           string              `json:"period"`
	CurrentYear      int                 `json:"current_year"`
	PreviousYear     int                 `json:"previous_year"`
	Summary          ComparisonSummary   `json:"summary"`
	MonthlyBreakdown []MonthComparison   `json:"monthly_breakdown"`
	TopChanges       []ChangeHighlight   `json:"top_changes"`
	GeneratedAt      time.Time           `json:"generated_at"`
}

type ComparisonSummary struct {
	CurrentRevenue      float64 `json:"current_revenue"`
	CurrentRevenueWord  string  `json:"current_revenue_word"`
	PreviousRevenue     float64 `json:"previous_revenue"`
	PreviousRevenueWord string  `json:"previous_revenue_word"`
	RevenueChange       float64 `json:"revenue_change"`
	RevenueChangePercent float64 `json:"revenue_change_percent"`
	ChangeDirection     string  `json:"change_direction"` // up, down, stable

	CurrentOrders       int     `json:"current_orders"`
	PreviousOrders      int     `json:"previous_orders"`
	OrdersChangePercent float64 `json:"orders_change_percent"`

	CurrentProfit       float64 `json:"current_profit"`
	PreviousProfit      float64 `json:"previous_profit"`
	ProfitChangePercent float64 `json:"profit_change_percent"`

	CurrentAvgOrder     float64 `json:"current_avg_order"`
	PreviousAvgOrder    float64 `json:"previous_avg_order"`
}

type MonthComparison struct {
	Month              string  `json:"month"`
	MonthName          string  `json:"month_name"`
	CurrentRevenue     float64 `json:"current_revenue"`
	PreviousRevenue    float64 `json:"previous_revenue"`
	ChangePercent      float64 `json:"change_percent"`
	ChangeDirection    string  `json:"change_direction"`
}

type ChangeHighlight struct {
	Metric      string  `json:"metric"`
	Description string  `json:"description"`
	Change      float64 `json:"change_percent"`
	Trend       string  `json:"trend"` // positive, negative
}

func GetYoYComparison(ctx context.Context, shopID string, year, month int) (*YoYComparisonResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	now := time.Now()
	if year <= 0 {
		year = now.Year()
	}
	previousYear := year - 1

	logger.Info("[YoY Comparison] shopid=%s, year=%d, month=%d", shopID, year, month)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &YoYComparisonResponse{
		CurrentYear:     year,
		PreviousYear:    previousYear,
		MonthlyBreakdown: []MonthComparison{},
		TopChanges:      []ChangeHighlight{},
		GeneratedAt:    now,
	}

	// Determine date ranges
	var currentFrom, currentTo, previousFrom, previousTo string
	if month > 0 && month <= 12 {
		response.Period = fmt.Sprintf("Month %d comparison: %d vs %d", month, year, previousYear)
		currentFrom = fmt.Sprintf("%d-%02d-01", year, month)
		currentTo = fmt.Sprintf("%d-%02d-31", year, month)
		previousFrom = fmt.Sprintf("%d-%02d-01", previousYear, month)
		previousTo = fmt.Sprintf("%d-%02d-31", previousYear, month)
	} else {
		response.Period = fmt.Sprintf("Year comparison: %d vs %d", year, previousYear)
		currentFrom = fmt.Sprintf("%d-01-01", year)
		currentTo = fmt.Sprintf("%d-12-31", year)
		previousFrom = fmt.Sprintf("%d-01-01", previousYear)
		previousTo = fmt.Sprintf("%d-12-31", previousYear)
	}

	// Get current year data
	var currentRevenue, currentProfit float64
	var currentOrders int
	query := `
		SELECT
			COALESCE(SUM(totalamount), 0) as revenue,
			COUNT(*) as orders,
			COALESCE(SUM(totalamount - COALESCE(totalcost, 0)), 0) as profit
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`
	db.QueryRow(query, currentFrom, currentTo).Scan(&currentRevenue, &currentOrders, &currentProfit)

	// Get previous year data
	var previousRevenue, previousProfit float64
	var previousOrders int
	db.QueryRow(query, previousFrom, previousTo).Scan(&previousRevenue, &previousOrders, &previousProfit)

	// Calculate changes
	revenueChange := currentRevenue - previousRevenue
	revenueChangePercent := calculateGrowth(currentRevenue, previousRevenue)
	ordersChangePercent := calculateGrowth(float64(currentOrders), float64(previousOrders))
	profitChangePercent := calculateGrowth(currentProfit, previousProfit)

	currentAvgOrder := float64(0)
	previousAvgOrder := float64(0)
	if currentOrders > 0 {
		currentAvgOrder = currentRevenue / float64(currentOrders)
	}
	if previousOrders > 0 {
		previousAvgOrder = previousRevenue / float64(previousOrders)
	}

	response.Summary = ComparisonSummary{
		CurrentRevenue:       currentRevenue,
		CurrentRevenueWord:   formatAmountWord(currentRevenue),
		PreviousRevenue:      previousRevenue,
		PreviousRevenueWord:  formatAmountWord(previousRevenue),
		RevenueChange:        revenueChange,
		RevenueChangePercent: revenueChangePercent,
		ChangeDirection:      getGrowthDirection(revenueChangePercent),
		CurrentOrders:        currentOrders,
		PreviousOrders:       previousOrders,
		OrdersChangePercent:  ordersChangePercent,
		CurrentProfit:        currentProfit,
		PreviousProfit:       previousProfit,
		ProfitChangePercent:  profitChangePercent,
		CurrentAvgOrder:      currentAvgOrder,
		PreviousAvgOrder:     previousAvgOrder,
	}

	// Get monthly breakdown for full year comparison
	if month == 0 {
		monthNames := []string{"", "January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"}

		for m := 1; m <= 12; m++ {
			mFrom := fmt.Sprintf("%d-%02d-01", year, m)
			mTo := fmt.Sprintf("%d-%02d-31", year, m)
			pFrom := fmt.Sprintf("%d-%02d-01", previousYear, m)
			pTo := fmt.Sprintf("%d-%02d-31", previousYear, m)

			var mCurrent, mPrevious float64
			db.QueryRow(`SELECT COALESCE(SUM(totalamount), 0) FROM doc WHERE transflag IN (16, 18) AND docdate >= $1 AND docdate <= $2 AND (isdelete = false OR isdelete IS NULL)`, mFrom, mTo).Scan(&mCurrent)
			db.QueryRow(`SELECT COALESCE(SUM(totalamount), 0) FROM doc WHERE transflag IN (16, 18) AND docdate >= $1 AND docdate <= $2 AND (isdelete = false OR isdelete IS NULL)`, pFrom, pTo).Scan(&mPrevious)

			change := calculateGrowth(mCurrent, mPrevious)

			response.MonthlyBreakdown = append(response.MonthlyBreakdown, MonthComparison{
				Month:           fmt.Sprintf("%02d", m),
				MonthName:       monthNames[m],
				CurrentRevenue:  mCurrent,
				PreviousRevenue: mPrevious,
				ChangePercent:   change,
				ChangeDirection: getGrowthDirection(change),
			})
		}
	}

	// Generate top changes highlights
	response.TopChanges = []ChangeHighlight{
		{
			Metric:      "Revenue",
			Description: fmt.Sprintf("Revenue %s by %.1f%%", getGrowthDirection(revenueChangePercent), abs(revenueChangePercent)),
			Change:      revenueChangePercent,
			Trend:       getTrend(revenueChangePercent),
		},
		{
			Metric:      "Orders",
			Description: fmt.Sprintf("Orders %s by %.1f%%", getGrowthDirection(ordersChangePercent), abs(ordersChangePercent)),
			Change:      ordersChangePercent,
			Trend:       getTrend(ordersChangePercent),
		},
		{
			Metric:      "Profit",
			Description: fmt.Sprintf("Profit %s by %.1f%%", getGrowthDirection(profitChangePercent), abs(profitChangePercent)),
			Change:      profitChangePercent,
			Trend:       getTrend(profitChangePercent),
		},
	}

	return response, nil
}

// ==================== Month-over-Month Comparison ====================

type MoMComparisonRequest struct {
	ShopID string `json:"shop_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
}

type MoMComparisonResponse struct {
	Period           string            `json:"period"`
	CurrentMonth     MonthInfo         `json:"current_month"`
	PreviousMonth    MonthInfo         `json:"previous_month"`
	Summary          ComparisonSummary `json:"summary"`
	WeeklyBreakdown  []WeekComparison  `json:"weekly_breakdown"`
	DailyTrend       []DailyComparison `json:"daily_trend"`
	GeneratedAt      time.Time         `json:"generated_at"`
}

type MonthInfo struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	MonthName string `json:"month_name"`
}

type WeekComparison struct {
	Week            int     `json:"week"`
	CurrentRevenue  float64 `json:"current_revenue"`
	PreviousRevenue float64 `json:"previous_revenue"`
	ChangePercent   float64 `json:"change_percent"`
}

type DailyComparison struct {
	Day             int     `json:"day"`
	CurrentRevenue  float64 `json:"current_revenue"`
	PreviousRevenue float64 `json:"previous_revenue"`
}

func GetMoMComparison(ctx context.Context, shopID string, year, month int) (*MoMComparisonResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	now := time.Now()
	if year <= 0 {
		year = now.Year()
	}
	if month <= 0 || month > 12 {
		month = int(now.Month())
	}

	// Calculate previous month
	prevYear := year
	prevMonth := month - 1
	if prevMonth <= 0 {
		prevMonth = 12
		prevYear = year - 1
	}

	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	logger.Info("[MoM Comparison] shopid=%s, current=%d-%02d, previous=%d-%02d", shopID, year, month, prevYear, prevMonth)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &MoMComparisonResponse{
		Period: fmt.Sprintf("%s %d vs %s %d", monthNames[month], year, monthNames[prevMonth], prevYear),
		CurrentMonth: MonthInfo{
			Year:      year,
			Month:     month,
			MonthName: monthNames[month],
		},
		PreviousMonth: MonthInfo{
			Year:      prevYear,
			Month:     prevMonth,
			MonthName: monthNames[prevMonth],
		},
		WeeklyBreakdown: []WeekComparison{},
		DailyTrend:      []DailyComparison{},
		GeneratedAt:    now,
	}

	currentFrom := fmt.Sprintf("%d-%02d-01", year, month)
	currentTo := fmt.Sprintf("%d-%02d-31", year, month)
	previousFrom := fmt.Sprintf("%d-%02d-01", prevYear, prevMonth)
	previousTo := fmt.Sprintf("%d-%02d-31", prevYear, prevMonth)

	// Get current month data
	var currentRevenue, currentProfit float64
	var currentOrders int
	query := `
		SELECT
			COALESCE(SUM(totalamount), 0) as revenue,
			COUNT(*) as orders,
			COALESCE(SUM(totalamount - COALESCE(totalcost, 0)), 0) as profit
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`
	db.QueryRow(query, currentFrom, currentTo).Scan(&currentRevenue, &currentOrders, &currentProfit)

	// Get previous month data
	var previousRevenue, previousProfit float64
	var previousOrders int
	db.QueryRow(query, previousFrom, previousTo).Scan(&previousRevenue, &previousOrders, &previousProfit)

	// Calculate changes
	revenueChangePercent := calculateGrowth(currentRevenue, previousRevenue)
	ordersChangePercent := calculateGrowth(float64(currentOrders), float64(previousOrders))
	profitChangePercent := calculateGrowth(currentProfit, previousProfit)

	currentAvgOrder := float64(0)
	previousAvgOrder := float64(0)
	if currentOrders > 0 {
		currentAvgOrder = currentRevenue / float64(currentOrders)
	}
	if previousOrders > 0 {
		previousAvgOrder = previousRevenue / float64(previousOrders)
	}

	response.Summary = ComparisonSummary{
		CurrentRevenue:       currentRevenue,
		CurrentRevenueWord:   formatAmountWord(currentRevenue),
		PreviousRevenue:      previousRevenue,
		PreviousRevenueWord:  formatAmountWord(previousRevenue),
		RevenueChange:        currentRevenue - previousRevenue,
		RevenueChangePercent: revenueChangePercent,
		ChangeDirection:      getGrowthDirection(revenueChangePercent),
		CurrentOrders:        currentOrders,
		PreviousOrders:       previousOrders,
		OrdersChangePercent:  ordersChangePercent,
		CurrentProfit:        currentProfit,
		PreviousProfit:       previousProfit,
		ProfitChangePercent:  profitChangePercent,
		CurrentAvgOrder:      currentAvgOrder,
		PreviousAvgOrder:     previousAvgOrder,
	}

	// Get weekly breakdown
	for week := 1; week <= 4; week++ {
		startDay := (week - 1) * 7 + 1
		endDay := week * 7
		if endDay > 31 {
			endDay = 31
		}

		cFrom := fmt.Sprintf("%d-%02d-%02d", year, month, startDay)
		cTo := fmt.Sprintf("%d-%02d-%02d", year, month, endDay)
		pFrom := fmt.Sprintf("%d-%02d-%02d", prevYear, prevMonth, startDay)
		pTo := fmt.Sprintf("%d-%02d-%02d", prevYear, prevMonth, endDay)

		var wCurrent, wPrevious float64
		db.QueryRow(`SELECT COALESCE(SUM(totalamount), 0) FROM doc WHERE transflag IN (16, 18) AND docdate >= $1 AND docdate <= $2 AND (isdelete = false OR isdelete IS NULL)`, cFrom, cTo).Scan(&wCurrent)
		db.QueryRow(`SELECT COALESCE(SUM(totalamount), 0) FROM doc WHERE transflag IN (16, 18) AND docdate >= $1 AND docdate <= $2 AND (isdelete = false OR isdelete IS NULL)`, pFrom, pTo).Scan(&wPrevious)

		response.WeeklyBreakdown = append(response.WeeklyBreakdown, WeekComparison{
			Week:            week,
			CurrentRevenue:  wCurrent,
			PreviousRevenue: wPrevious,
			ChangePercent:   calculateGrowth(wCurrent, wPrevious),
		})
	}

	// DailyTrend — เปรียบเทียบยอดขายรายวัน เดือนปัจจุบัน vs เดือนก่อน
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	daysInPrevMonth := time.Date(prevYear, time.Month(prevMonth)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	// Query ยอดขายรายวันของเดือนปัจจุบัน
	currentDailyQuery := `
		SELECT
			EXTRACT(DAY FROM docdate)::int as day,
			COALESCE(SUM(totalamount), 0) as revenue
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY EXTRACT(DAY FROM docdate)
		ORDER BY day
	`
	currentDailyMap := make(map[int]float64)
	curRows, curErr := db.Query(currentDailyQuery, currentFrom, currentTo)
	if curErr == nil {
		defer curRows.Close()
		for curRows.Next() {
			var day int
			var rev float64
			if err := curRows.Scan(&day, &rev); err == nil {
				currentDailyMap[day] = rev
			}
		}
	}

	// Query ยอดขายรายวันของเดือนก่อน
	prevDailyMap := make(map[int]float64)
	prevRows, prevErr := db.Query(currentDailyQuery, previousFrom, previousTo)
	if prevErr == nil {
		defer prevRows.Close()
		for prevRows.Next() {
			var day int
			var rev float64
			if err := prevRows.Scan(&day, &rev); err == nil {
				prevDailyMap[day] = rev
			}
		}
	}

	// สร้าง DailyTrend (ใช้จำนวนวันของเดือนที่มากกว่า)
	maxDays := daysInMonth
	if daysInPrevMonth > maxDays {
		maxDays = daysInPrevMonth
	}
	for d := 1; d <= maxDays; d++ {
		response.DailyTrend = append(response.DailyTrend, DailyComparison{
			Day:             d,
			CurrentRevenue:  currentDailyMap[d],
			PreviousRevenue: prevDailyMap[d],
		})
	}

	return response, nil
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func getTrend(change float64) string {
	if change > 0 {
		return "positive"
	} else if change < 0 {
		return "negative"
	}
	return "neutral"
}

package tools

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ==================== Inventory Value ====================

type InventoryValueRequest struct {
	ShopID     string `json:"shop_id"`
	WHCode     string `json:"whcode"`
}

type InventoryValueResponse struct {
	Summary          InvValueSummary    `json:"summary"`
	ByWarehouse      []WarehouseValue   `json:"by_warehouse"`
	ByCategory       []CategoryValue    `json:"by_category"`
	TopValueItems    []TopValueItem     `json:"top_value_items"`
	GeneratedAt      time.Time          `json:"generated_at"`
}

type InvValueSummary struct {
	TotalValue       float64 `json:"total_value"`
	TotalValueWord   string  `json:"total_value_word"`
	TotalItems       int     `json:"total_items"`
	TotalSKUs        int     `json:"total_skus"`
	AverageValuePerSKU float64 `json:"average_value_per_sku"`
}

type WarehouseValue struct {
	WarehouseCode string  `json:"warehouse_code"`
	WarehouseName string  `json:"warehouse_name"`
	Value         float64 `json:"value"`
	Percentage    float64 `json:"percentage"`
	ItemCount     int     `json:"item_count"`
}

type CategoryValue struct {
	CategoryCode string  `json:"category_code"`
	CategoryName string  `json:"category_name"`
	Value        float64 `json:"value"`
	Percentage   float64 `json:"percentage"`
	ItemCount    int     `json:"item_count"`
}

type TopValueItem struct {
	ItemCode     string  `json:"itemcode"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity"`
	UnitCost     float64 `json:"unit_cost"`
	TotalValue   float64 `json:"total_value"`
	Percentage   float64 `json:"percentage"`
}

func GetInventoryValue(ctx context.Context, shopID, whcode string) (*InventoryValueResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	logger.Info("[Inventory Value] shopid=%s, whcode=%s", shopID, whcode)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &InventoryValueResponse{
		ByWarehouse:   []WarehouseValue{},
		ByCategory:    []CategoryValue{},
		TopValueItems: []TopValueItem{},
		GeneratedAt:   time.Now(),
	}

	// Get total inventory value
	whereWH := ""
	if whcode != "" {
		whereWH = fmt.Sprintf(" AND whcode = '%s'", whcode)
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(DISTINCT itemcode) as sku_count,
			COALESCE(SUM(ABS(balance_qty)), 0) as total_qty,
			COALESCE(SUM(ABS(balance_qty) * COALESCE(avgcost, 0)), 0) as total_value
		FROM (
			SELECT
				itemcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty,
				(SELECT COALESCE(avgcost, 0) FROM productbarcode pb WHERE pb.itemcode = d.itemcode LIMIT 1) as avgcost
			FROM docdetail d
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
				%s
			GROUP BY itemcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) != 0
		) inv
	`, whereWH)

	var skuCount int
	var totalQty, totalValue float64
	err = db.QueryRow(query).Scan(&skuCount, &totalQty, &totalValue)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("[Inventory Value] Query failed: %v", err)
	}

	avgValuePerSKU := float64(0)
	if skuCount > 0 {
		avgValuePerSKU = totalValue / float64(skuCount)
	}

	response.Summary = InvValueSummary{
		TotalValue:        totalValue,
		TotalValueWord:    formatAmountWord(totalValue),
		TotalItems:        int(totalQty),
		TotalSKUs:         skuCount,
		AverageValuePerSKU: avgValuePerSKU,
	}

	// ByWarehouse — มูลค่าแยกตามคลัง
	whQuery := `
		SELECT
			COALESCE(whcode, 'DEFAULT') as wh_code,
			COALESCE(whcode, 'คลังหลัก') as wh_name,
			COALESCE(SUM(ABS(balance_qty) * COALESCE(avgcost, 0)), 0) as value,
			COUNT(DISTINCT itemcode) as item_count
		FROM (
			SELECT
				d.itemcode, d.whcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty,
				(SELECT COALESCE(avgcost, 0) FROM productbarcode pb WHERE pb.itemcode = d.itemcode LIMIT 1) as avgcost
			FROM docdetail d
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY d.itemcode, d.whcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) != 0
		) inv
		GROUP BY whcode
		ORDER BY value DESC
	`
	whRows, whErr := db.Query(whQuery)
	if whErr == nil {
		defer whRows.Close()
		for whRows.Next() {
			var wv WarehouseValue
			if err := whRows.Scan(&wv.WarehouseCode, &wv.WarehouseName, &wv.Value, &wv.ItemCount); err != nil {
				continue
			}
			if totalValue > 0 {
				wv.Percentage = (wv.Value / totalValue) * 100
			}
			response.ByWarehouse = append(response.ByWarehouse, wv)
		}
	}
	if len(response.ByWarehouse) == 0 {
		response.ByWarehouse = []WarehouseValue{}
	}

	// ByCategory — มูลค่าแยกตามหมวด
	catQuery := `
		SELECT
			COALESCE(categorycode, 'N/A') as cat_code,
			COALESCE(categorycode, 'ไม่ระบุหมวด') as cat_name,
			COALESCE(SUM(ABS(balance_qty) * COALESCE(avgcost, 0)), 0) as value,
			COUNT(DISTINCT itemcode) as item_count
		FROM (
			SELECT
				d.itemcode, d.categorycode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty,
				(SELECT COALESCE(avgcost, 0) FROM productbarcode pb WHERE pb.itemcode = d.itemcode LIMIT 1) as avgcost
			FROM docdetail d
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY d.itemcode, d.categorycode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) != 0
		) inv
		GROUP BY categorycode
		ORDER BY value DESC
		LIMIT 20
	`
	catRows, catErr := db.Query(catQuery)
	if catErr == nil {
		defer catRows.Close()
		for catRows.Next() {
			var cv CategoryValue
			if err := catRows.Scan(&cv.CategoryCode, &cv.CategoryName, &cv.Value, &cv.ItemCount); err != nil {
				continue
			}
			if totalValue > 0 {
				cv.Percentage = (cv.Value / totalValue) * 100
			}
			response.ByCategory = append(response.ByCategory, cv)
		}
	}
	if len(response.ByCategory) == 0 {
		response.ByCategory = []CategoryValue{}
	}

	// TopValueItems — สินค้ามูลค่าสูงสุด
	topQuery := fmt.Sprintf(`
		SELECT
			inv.itemcode,
			COALESCE(pb.name0, inv.itemcode) as name,
			ABS(inv.balance_qty) as qty,
			COALESCE(pb.avgcost, 0) as unit_cost,
			ABS(inv.balance_qty) * COALESCE(pb.avgcost, 0) as total_value
		FROM (
			SELECT
				itemcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty
			FROM docdetail
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
				%s
			GROUP BY itemcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) != 0
		) inv
		LEFT JOIN productbarcode pb ON pb.itemcode = inv.itemcode
		ORDER BY total_value DESC
		LIMIT 20
	`, whereWH)
	topRows, topErr := db.Query(topQuery)
	if topErr == nil {
		defer topRows.Close()
		for topRows.Next() {
			var tv TopValueItem
			if err := topRows.Scan(&tv.ItemCode, &tv.Name, &tv.Quantity, &tv.UnitCost, &tv.TotalValue); err != nil {
				continue
			}
			if totalValue > 0 {
				tv.Percentage = (tv.TotalValue / totalValue) * 100
			}
			response.TopValueItems = append(response.TopValueItems, tv)
		}
	}
	if len(response.TopValueItems) == 0 {
		response.TopValueItems = []TopValueItem{}
	}

	return response, nil
}

// ==================== Low Stock Alerts ====================

type LowStockAlertsRequest struct {
	ShopID    string `json:"shop_id"`
	Threshold int    `json:"threshold"` // Default 10
	Limit     int    `json:"limit"`     // Default 50
}

type LowStockAlertsResponse struct {
	Summary       LowStockSummary  `json:"summary"`
	Alerts        []LowStockItem   `json:"alerts"`
	GeneratedAt   time.Time        `json:"generated_at"`
}

type LowStockSummary struct {
	TotalLowStock    int `json:"total_low_stock"`
	TotalOutOfStock  int `json:"total_out_of_stock"`
	CriticalCount    int `json:"critical_count"`    // < 5 units
	WarningCount     int `json:"warning_count"`     // 5-10 units
}

type LowStockItem struct {
	ItemCode       string  `json:"itemcode"`
	Name           string  `json:"name"`
	CurrentStock   float64 `json:"current_stock"`
	StockWord      string  `json:"stock_word"`
	MinStock       float64 `json:"min_stock"`
	ReorderQty     float64 `json:"reorder_qty"`
	LastSaleDate   string  `json:"last_sale_date"`
	AvgDailySales  float64 `json:"avg_daily_sales"`
	DaysOfStock    int     `json:"days_of_stock"`
	Priority       string  `json:"priority"` // critical, high, medium, low
	WarehouseCode  string  `json:"warehouse_code"`
}

func GetLowStockAlerts(ctx context.Context, shopID string, threshold, limit int) (*LowStockAlertsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	if threshold <= 0 {
		threshold = 10
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	logger.Info("[Low Stock Alerts] shopid=%s, threshold=%d, limit=%d", shopID, threshold, limit)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &LowStockAlertsResponse{
		Alerts:      []LowStockItem{},
		GeneratedAt: time.Now(),
	}

	// Get low stock items
	query := fmt.Sprintf(`
		SELECT
			inv.itemcode,
			COALESCE(pb.name0, inv.itemcode) as name,
			inv.balance_qty,
			COALESCE(inv.whcode, '') as whcode
		FROM (
			SELECT
				itemcode,
				whcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty
			FROM docdetail
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY itemcode, whcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) > 0
				AND SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) <= %d
		) inv
		LEFT JOIN productbarcode pb ON pb.itemcode = inv.itemcode
		ORDER BY inv.balance_qty ASC
		LIMIT %d
	`, threshold, limit)

	rows, err := db.Query(query)
	if err != nil {
		logger.Error("[Low Stock Alerts] Query failed: %v", err)
		return response, nil
	}
	defer rows.Close()

	var criticalCount, warningCount int
	for rows.Next() {
		var itemCode, name, whcode string
		var balanceQty float64

		if err := rows.Scan(&itemCode, &name, &balanceQty, &whcode); err != nil {
			continue
		}

		priority := "low"
		if balanceQty <= 0 {
			priority = "critical"
			criticalCount++
		} else if balanceQty <= 5 {
			priority = "high"
			criticalCount++
		} else if balanceQty <= float64(threshold) {
			priority = "medium"
			warningCount++
		}

		response.Alerts = append(response.Alerts, LowStockItem{
			ItemCode:      itemCode,
			Name:          name,
			CurrentStock:  balanceQty,
			StockWord:     fmt.Sprintf("%.0f", balanceQty),
			Priority:      priority,
			WarehouseCode: whcode,
		})
	}

	response.Summary = LowStockSummary{
		TotalLowStock:   len(response.Alerts),
		TotalOutOfStock: 0,
		CriticalCount:   criticalCount,
		WarningCount:    warningCount,
	}

	return response, nil
}

// ==================== Dead Stock ====================

type DeadStockRequest struct {
	ShopID  string `json:"shop_id"`
	Days    int    `json:"days"`  // No movement for X days (default 90)
	Limit   int    `json:"limit"`
}

type DeadStockResponse struct {
	Summary       DeadStockSummary  `json:"summary"`
	Items         []DeadStockItem   `json:"items"`
	GeneratedAt   time.Time         `json:"generated_at"`
}

type DeadStockSummary struct {
	TotalItems      int     `json:"total_items"`
	TotalValue      float64 `json:"total_value"`
	TotalValueWord  string  `json:"total_value_word"`
	ByAgeBucket     []AgeBucket `json:"by_age_bucket"`
}

type AgeBucket struct {
	Label      string  `json:"label"`
	ItemCount  int     `json:"item_count"`
	Value      float64 `json:"value"`
}

type DeadStockItem struct {
	ItemCode       string  `json:"itemcode"`
	Name           string  `json:"name"`
	CurrentStock   float64 `json:"current_stock"`
	StockValue     float64 `json:"stock_value"`
	LastMovement   string  `json:"last_movement_date"`
	DaysSinceMove  int     `json:"days_since_movement"`
	Recommendation string  `json:"recommendation"`
}

func GetDeadStock(ctx context.Context, shopID string, days, limit int) (*DeadStockResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	if days <= 0 {
		days = 90
	}
	if limit <= 0 {
		limit = 50
	}

	logger.Info("[Dead Stock] shopid=%s, days=%d, limit=%d", shopID, days, limit)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &DeadStockResponse{
		Items:       []DeadStockItem{},
		GeneratedAt: time.Now(),
	}

	// Find items with no movement in X days
	cutoffDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := fmt.Sprintf(`
		WITH current_stock AS (
			SELECT
				itemcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty
			FROM docdetail
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY itemcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) > 0
		),
		last_movement AS (
			SELECT
				itemcode,
				MAX(docdate) as last_move_date
			FROM docdetail
			WHERE transflag IN (16, 18, 20)
			GROUP BY itemcode
		)
		SELECT
			cs.itemcode,
			COALESCE(pb.name0, cs.itemcode) as name,
			cs.balance_qty,
			COALESCE(lm.last_move_date, '2000-01-01'::date) as last_move,
			COALESCE(pb.avgcost, 0) * cs.balance_qty as stock_value
		FROM current_stock cs
		LEFT JOIN last_movement lm ON lm.itemcode = cs.itemcode
		LEFT JOIN productbarcode pb ON pb.itemcode = cs.itemcode
		WHERE COALESCE(lm.last_move_date, '2000-01-01'::date) < '%s'
		ORDER BY stock_value DESC
		LIMIT %d
	`, cutoffDate, limit)

	rows, err := db.Query(query)
	if err != nil {
		logger.Error("[Dead Stock] Query failed: %v", err)
		return response, nil
	}
	defer rows.Close()

	var totalValue float64
	for rows.Next() {
		var itemCode, name string
		var balanceQty, stockValue float64
		var lastMove time.Time

		if err := rows.Scan(&itemCode, &name, &balanceQty, &lastMove, &stockValue); err != nil {
			continue
		}

		daysSince := int(time.Since(lastMove).Hours() / 24)

		recommendation := "Consider discount sale"
		if daysSince > 180 {
			recommendation = "Consider liquidation or write-off"
		} else if daysSince > 120 {
			recommendation = "Run clearance promotion"
		}

		response.Items = append(response.Items, DeadStockItem{
			ItemCode:       itemCode,
			Name:           name,
			CurrentStock:   balanceQty,
			StockValue:     stockValue,
			LastMovement:   lastMove.Format("2006-01-02"),
			DaysSinceMove:  daysSince,
			Recommendation: recommendation,
		})

		totalValue += stockValue
	}

	response.Summary = DeadStockSummary{
		TotalItems:     len(response.Items),
		TotalValue:     totalValue,
		TotalValueWord: formatAmountWord(totalValue),
		ByAgeBucket: []AgeBucket{
			{Label: "30-60 days", ItemCount: 0, Value: 0},
			{Label: "61-90 days", ItemCount: 0, Value: 0},
			{Label: "91-180 days", ItemCount: 0, Value: 0},
			{Label: "180+ days", ItemCount: 0, Value: 0},
		},
	}

	return response, nil
}

// ==================== Inventory Turnover ====================

type InventoryTurnoverRequest struct {
	ShopID   string `json:"shop_id"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

type InventoryTurnoverResponse struct {
	Summary        TurnoverSummary    `json:"summary"`
	ByCategory     []CategoryTurnover `json:"by_category"`
	FastMovers     []TurnoverItem     `json:"fast_movers"`
	SlowMovers     []TurnoverItem     `json:"slow_movers"`
	GeneratedAt    time.Time          `json:"generated_at"`
}

type TurnoverSummary struct {
	OverallTurnover      float64 `json:"overall_turnover_ratio"`
	TurnoverDescription  string  `json:"turnover_description"`
	AverageDaysToSell    int     `json:"average_days_to_sell"`
	COGS                 float64 `json:"cogs"`
	AverageInventory     float64 `json:"average_inventory"`
}

type CategoryTurnover struct {
	CategoryCode string  `json:"category_code"`
	CategoryName string  `json:"category_name"`
	Turnover     float64 `json:"turnover_ratio"`
	DaysToSell   int     `json:"days_to_sell"`
}

type TurnoverItem struct {
	ItemCode     string  `json:"itemcode"`
	Name         string  `json:"name"`
	Turnover     float64 `json:"turnover_ratio"`
	DaysToSell   int     `json:"days_to_sell"`
	UnitsSold    float64 `json:"units_sold"`
	CurrentStock float64 `json:"current_stock"`
}

func GetInventoryTurnover(ctx context.Context, shopID, fromDate, toDate string) (*InventoryTurnoverResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	now := time.Now()
	if fromDate == "" {
		fromDate = now.AddDate(0, -3, 0).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}

	logger.Info("[Inventory Turnover] shopid=%s, from=%s, to=%s", shopID, fromDate, toDate)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &InventoryTurnoverResponse{
		ByCategory:  []CategoryTurnover{},
		FastMovers:  []TurnoverItem{},
		SlowMovers:  []TurnoverItem{},
		GeneratedAt: now,
	}

	// Calculate COGS for period
	var cogs float64
	query := `
		SELECT COALESCE(SUM(totalcost), 0)
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND (isdelete = false OR isdelete IS NULL)
	`
	db.QueryRow(query, fromDate, toDate).Scan(&cogs)

	// Estimate average inventory (simplified)
	var avgInventory float64
	query = `
		SELECT COALESCE(SUM(ABS(balance_qty) * avgcost), 0) / 2
		FROM (
			SELECT
				d.itemcode,
				SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty,
				COALESCE((SELECT avgcost FROM productbarcode WHERE itemcode = d.itemcode LIMIT 1), 0) as avgcost
			FROM docdetail d
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY d.itemcode
		) inv
	`
	db.QueryRow(query).Scan(&avgInventory)

	turnover := float64(0)
	daysToSell := 365
	if avgInventory > 0 {
		turnover = cogs / avgInventory
		if turnover > 0 {
			daysToSell = int(365 / turnover)
		}
	}

	description := "Low turnover - consider reducing inventory"
	if turnover >= 12 {
		description = "Excellent turnover - inventory moves quickly"
	} else if turnover >= 6 {
		description = "Good turnover - healthy inventory movement"
	} else if turnover >= 4 {
		description = "Average turnover - room for improvement"
	}

	response.Summary = TurnoverSummary{
		OverallTurnover:     turnover,
		TurnoverDescription: description,
		AverageDaysToSell:   daysToSell,
		COGS:                cogs,
		AverageInventory:    avgInventory,
	}

	// FastMovers — สินค้าหมุนเวียนเร็ว (ขายเยอะ)
	fastQuery := `
		SELECT
			d.itemcode,
			COALESCE(pb.name0, d.itemcode) as name,
			SUM((d.totalqty * d.calcflag * -1) * d.unitstand / NULLIF(d.unitdivide, 0)) as units_sold,
			COALESCE(inv.balance_qty, 0) as current_stock
		FROM docdetail d
		LEFT JOIN (SELECT DISTINCT ON (itemcode) itemcode, name0 FROM productbarcode) pb ON pb.itemcode = d.itemcode
		LEFT JOIN (
			SELECT itemcode, SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty
			FROM docdetail
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY itemcode
		) inv ON inv.itemcode = d.itemcode
		WHERE d.transflag IN (16, 18)
			AND d.docdate >= $1 AND d.docdate <= $2
		GROUP BY d.itemcode, pb.name0, inv.balance_qty
		HAVING SUM((d.totalqty * d.calcflag * -1) * d.unitstand / NULLIF(d.unitdivide, 0)) > 0
		ORDER BY units_sold DESC
		LIMIT 10
	`
	fastRows, fastErr := db.Query(fastQuery, fromDate, toDate)
	if fastErr == nil {
		defer fastRows.Close()
		for fastRows.Next() {
			var t TurnoverItem
			if err := fastRows.Scan(&t.ItemCode, &t.Name, &t.UnitsSold, &t.CurrentStock); err != nil {
				continue
			}
			if t.CurrentStock > 0 {
				t.Turnover = t.UnitsSold / t.CurrentStock
				if t.Turnover > 0 {
					t.DaysToSell = int(365 / t.Turnover)
				}
			}
			response.FastMovers = append(response.FastMovers, t)
		}
	}
	if len(response.FastMovers) == 0 {
		response.FastMovers = []TurnoverItem{}
	}

	// SlowMovers — สินค้าหมุนเวียนช้า (stock เยอะ ขายน้อย)
	slowQuery := `
		SELECT
			inv.itemcode,
			COALESCE(pb.name0, inv.itemcode) as name,
			COALESCE(sold.units_sold, 0) as units_sold,
			inv.balance_qty as current_stock
		FROM (
			SELECT itemcode, SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance_qty
			FROM docdetail
			WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
			GROUP BY itemcode
			HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) > 10
		) inv
		LEFT JOIN (SELECT DISTINCT ON (itemcode) itemcode, name0 FROM productbarcode) pb ON pb.itemcode = inv.itemcode
		LEFT JOIN (
			SELECT itemcode, SUM((totalqty * calcflag * -1) * unitstand / NULLIF(unitdivide, 0)) as units_sold
			FROM docdetail
			WHERE transflag IN (16, 18) AND docdate >= $1 AND docdate <= $2
			GROUP BY itemcode
		) sold ON sold.itemcode = inv.itemcode
		ORDER BY COALESCE(sold.units_sold, 0) / inv.balance_qty ASC, inv.balance_qty DESC
		LIMIT 10
	`
	slowRows, slowErr := db.Query(slowQuery, fromDate, toDate)
	if slowErr == nil {
		defer slowRows.Close()
		for slowRows.Next() {
			var t TurnoverItem
			if err := slowRows.Scan(&t.ItemCode, &t.Name, &t.UnitsSold, &t.CurrentStock); err != nil {
				continue
			}
			if t.CurrentStock > 0 && t.UnitsSold > 0 {
				t.Turnover = t.UnitsSold / t.CurrentStock
				if t.Turnover > 0 {
					t.DaysToSell = int(365 / t.Turnover)
				}
			} else {
				t.DaysToSell = 999
			}
			response.SlowMovers = append(response.SlowMovers, t)
		}
	}
	if len(response.SlowMovers) == 0 {
		response.SlowMovers = []TurnoverItem{}
	}

	return response, nil
}

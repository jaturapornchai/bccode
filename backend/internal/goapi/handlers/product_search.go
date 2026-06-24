package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// ==================== Request/Response Structs ====================

// UnifiedSearchRequest - Request body สำหรับค้นหาสินค้าแบบ Unified
type UnifiedSearchRequest struct {
	HoldingCode    string `json:"holdingcode"`
	Keyword        string `json:"keyword"`
	WHCode         string `json:"whcode"`
	LocationCode   string `json:"locationcode"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"` // สำหรับ pagination (infinite scroll)
	IncludeBalance bool   `json:"includebalance"`
}

// SearchProductUnit - หน่วยสินค้าพร้อมราคา
type SearchProductUnit struct {
	Barcode     string  `json:"barcode"`
	UnitCode    string  `json:"unitcode"`
	UnitName    string  `json:"unitname"`
	UnitStand   float64 `json:"unitstand"`
	UnitDivide  float64 `json:"unitdivide"`
	Price1      float64 `json:"price1"`
	PriceRetail float64 `json:"priceretail"`
}

// SearchLocationBalance - ยอดคงเหลือแยกตาม Location
type SearchLocationBalance struct {
	LocationCode string  `json:"locationcode"`
	BalanceQty   float64 `json:"balanceqty"`
	BalanceWord  string  `json:"balanceword"`
}

// SearchWarehouseBalance - ยอดคงเหลือแยกตาม Warehouse พร้อม Locations
type SearchWarehouseBalance struct {
	WarehouseCode string                  `json:"warehousecode"`
	BalanceQty    float64                 `json:"balanceqty"`
	BalanceWord   string                  `json:"balanceword"`
	Locations     []SearchLocationBalance `json:"locations,omitempty"`
}

// SearchBalanceInfo - ยอดคงเหลือพร้อม format และแยกตาม warehouse/location
type SearchBalanceInfo struct {
	Total      float64                  `json:"total"`
	Formatted  string                   `json:"formatted"`
	Warehouses []SearchWarehouseBalance `json:"warehouses,omitempty"`
	// ค้างรับ (PO ที่ยังไม่ได้รับของ)
	PendingRecvQty  float64 `json:"pendingrecvqty"`
	PendingRecvWord string  `json:"pendingrecvword"`
	// ค้างส่ง (SO ที่ยังไม่ได้ส่งของ)
	PendingSendQty  float64 `json:"pendingsendqty"`
	PendingSendWord string  `json:"pendingsendword"`
}

// SearchProductItem - สินค้าที่ค้นพบ
type SearchProductItem struct {
	ItemCode string              `json:"itemcode"`
	Name0    string              `json:"name0"`
	Units    []SearchProductUnit `json:"units"`
	Balance  *SearchBalanceInfo  `json:"balance,omitempty"`
	Score    int                 `json:"score"`
}

// UnifiedSearchResponse - Response สำหรับการค้นหา
type UnifiedSearchResponse struct {
	Status   string              `json:"status"`
	Count    int                 `json:"count"`
	Total    int                 `json:"total"`   // จำนวนสินค้าทั้งหมดที่พบ (ก่อน pagination)
	HasMore  bool                `json:"hasmore"` // มีข้อมูลเพิ่มหรือไม่ (สำหรับ infinite scroll)
	Products []SearchProductItem `json:"products"`
	Tokens   []string            `json:"tokens"`
}

// ==================== Main Handler ====================

// UnifiedProductSearchHandler - API endpoint สำหรับค้นหาสินค้าพร้อมยอดคงเหลือ
// POST /api/product/search
func UnifiedProductSearchHandler(c echo.Context) error {
	// 1. Parse request
	var req UnifiedSearchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Missing required parameter: holdingcode",
		})
	}

	if req.Keyword == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Missing required parameter: keyword",
		})
	}

	// Default limit
	if req.Limit <= 0 {
		req.Limit = 200
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	logger.Info("[UnifiedSearch] holdingcode=%s, keyword=%s, limit=%d, include_balance=%v",
		req.HoldingCode, req.Keyword, req.Limit, req.IncludeBalance)

	totalStart := time.Now()

	// 2. Tokenize keyword
	t0 := time.Now()
	tokens, err := tokenizeKeyword(req.Keyword)
	if err != nil {
		logger.Error("[UnifiedSearch] Tokenize failed: %v", err)
		// Fallback: ใช้ keyword เดิมเป็น token เดียว
		tokens = []string{req.Keyword}
	}
	logger.Info("[UnifiedSearch] Tokens: %v (took %dms)", tokens, time.Since(t0).Milliseconds())

	// 3. Connect to database
	t1 := time.Now()
	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Database connection failed",
			"error":   err.Error(),
		})
	}
	logger.Info("[UnifiedSearch] DB connect took %dms", time.Since(t1).Milliseconds())

	// 4. Search products (รองรับ pagination)
	t2 := time.Now()
	searchResult, err := searchProductsUnified(db, tokens, req.Limit, req.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Search failed",
			"error":   err.Error(),
		})
	}

	products := searchResult.Products
	total := searchResult.Total
	hasMore := (req.Offset + len(products)) < total

	logger.Info("[UnifiedSearch] Search took %dms — found %d products (offset=%d, total=%d, hasMore=%v)",
		time.Since(t2).Milliseconds(), len(products), req.Offset, total, hasMore)

	if len(products) == 0 {
		return c.JSON(http.StatusOK, UnifiedSearchResponse{
			Status:   "success",
			Count:    0,
			Total:    total,
			HasMore:  false,
			Products: []SearchProductItem{},
			Tokens:   tokens,
		})
	}

	// 5. Get all units for found itemcodes
	itemCodes := make([]string, len(products))
	for i, p := range products {
		itemCodes[i] = p.ItemCode
	}

	// 5+6. Fetch units, balance, packing พร้อมกัน (parallel goroutines)
	type unitsResult struct {
		data map[string][]SearchProductUnit
		err  error
		ms   int64
	}
	type balanceResult struct {
		data map[string]*SearchBalanceInfo
		err  error
		ms   int64
	}
	type packingResult struct {
		data map[string][]models.ProductBarcodePackingStruct
		err  error
		ms   int64
	}
	type pendingResult struct {
		data map[string]*pendingInfo
		err  error
		ms   int64
	}

	unitsCh := make(chan unitsResult, 1)
	balanceCh := make(chan balanceResult, 1)
	packingCh := make(chan packingResult, 1)
	pendingCh := make(chan pendingResult, 1)

	// Goroutine 1: Fetch units
	go func() {
		t := time.Now()
		data, err := fetchAllUnitsForSearch(db, itemCodes)
		unitsCh <- unitsResult{data: data, err: err, ms: time.Since(t).Milliseconds()}
	}()

	// Goroutine 2: Calculate balance (ถ้าต้องการ)
	if req.IncludeBalance {
		go func() {
			t := time.Now()
			data, err := calculateBalancesByLocation(db, itemCodes, req.WHCode, req.LocationCode)
			balanceCh <- balanceResult{data: data, err: err, ms: time.Since(t).Milliseconds()}
		}()

		// Goroutine 3: Fetch packing units
		go func() {
			t := time.Now()
			data, err := fetchPackingUnits(db, itemCodes)
			packingCh <- packingResult{data: data, err: err, ms: time.Since(t).Milliseconds()}
		}()

		// Goroutine 4: Fetch pending recv/send จาก product table
		go func() {
			t := time.Now()
			data, err := fetchPendingFromProduct(db, itemCodes)
			pendingCh <- pendingResult{data: data, err: err, ms: time.Since(t).Milliseconds()}
		}()
	}

	// รอ units result
	unitsRes := <-unitsCh
	logger.Info("[UnifiedSearch] Fetch units took %dms", unitsRes.ms)
	if unitsRes.err != nil {
		logger.Error("[UnifiedSearch] Fetch units failed: %v", unitsRes.err)
	} else {
		for i := range products {
			if units, ok := unitsRes.data[products[i].ItemCode]; ok {
				products[i].Units = units
			}
		}
	}

	// 6. Assign balance if requested
	if req.IncludeBalance {
		balanceRes := <-balanceCh
		logger.Info("[UnifiedSearch] Balance calc took %dms", balanceRes.ms)
		if balanceRes.err != nil {
			logger.Error("[UnifiedSearch] Calculate balance by location failed: %v", balanceRes.err)
		}

		packingRes := <-packingCh
		logger.Info("[UnifiedSearch] Packing units took %dms", packingRes.ms)
		if packingRes.err != nil {
			logger.Error("[UnifiedSearch] Fetch packing units failed: %v", packingRes.err)
		}

		pendingRes := <-pendingCh
		logger.Info("[UnifiedSearch] Pending recv/send took %dms", pendingRes.ms)
		if pendingRes.err != nil {
			logger.Error("[UnifiedSearch] Fetch pending failed: %v", pendingRes.err)
		}

		balanceMap := balanceRes.data
		packingMap := packingRes.data
		pendingMap := pendingRes.data

		// Assign balance to products และ format ด้วย packing
		for i := range products {
			itemCode := products[i].ItemCode

			// หา default unit name
			defaultUnitName := ""
			if len(products[i].Units) > 0 {
				for _, u := range products[i].Units {
					ratio := u.UnitStand / u.UnitDivide
					if ratio == 1.0 {
						defaultUnitName = u.UnitName
						if defaultUnitName == "" {
							defaultUnitName = u.UnitCode
						}
						break
					}
				}
				if defaultUnitName == "" {
					defaultUnitName = products[i].Units[len(products[i].Units)-1].UnitName
					if defaultUnitName == "" {
						defaultUnitName = products[i].Units[len(products[i].Units)-1].UnitCode
					}
				}
			}

			if balanceInfo, ok := balanceMap[itemCode]; ok && balanceInfo != nil {
				// Format ด้วย packing units
				packingUnits := packingMap[itemCode]
				formatBalanceWithPacking(balanceInfo, packingUnits, defaultUnitName)
				products[i].Balance = balanceInfo
			} else {
				// ไม่มี balance = 0
				products[i].Balance = &SearchBalanceInfo{
					Total:      0,
					Formatted:  "0",
					Warehouses: []SearchWarehouseBalance{},
				}
			}

			// Merge pending recv/send ลง balance info
			if pendingMap != nil {
				if p, ok := pendingMap[itemCode]; ok && p != nil {
					if products[i].Balance == nil {
						products[i].Balance = &SearchBalanceInfo{}
					}
					products[i].Balance.PendingRecvQty = p.RecvQty
					products[i].Balance.PendingRecvWord = p.RecvWord
					products[i].Balance.PendingSendQty = p.SendQty
					products[i].Balance.PendingSendWord = p.SendWord
				}
			}
		}
	}

	// 7. Calculate scores and rank
	for i := range products {
		products[i].Score = calculateSearchScore(products[i], tokens)
	}

	// Sort by score descending
	sort.Slice(products, func(i, j int) bool {
		return products[i].Score > products[j].Score
	})

	// 8. Return response
	logger.Info("[UnifiedSearch] TOTAL handler took %dms", time.Since(totalStart).Milliseconds())
	return c.JSON(http.StatusOK, UnifiedSearchResponse{
		Status:   "success",
		Count:    len(products),
		Total:    total,
		HasMore:  hasMore,
		Products: products,
		Tokens:   tokens,
	})
}

// ==================== Internal Functions ====================

// tokenizeKeyword - เรียก Thai NLP service เพื่อตัดคำ (พร้อม circuit breaker)
func tokenizeKeyword(keyword string) ([]string, error) {
	// Circuit breaker: ถ้า NLP service ล้มเหลวก่อนหน้า → ข้ามไปเลย ไม่ต้องรอ timeout
	if !nlpIsAvailable() {
		return fallbackTokenize(keyword), nil
	}

	thaiNLPURL := getThaiNLPURL()

	requestBody, err := json.Marshal(TokenizeRequest{Text: keyword})
	if err != nil {
		return fallbackTokenize(keyword), nil
	}

	resp, err := thaiNLPClient.Post(
		fmt.Sprintf("%s/tokenize", thaiNLPURL),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		nlpMarkFailed()
		return fallbackTokenize(keyword), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		nlpMarkFailed()
		return fallbackTokenize(keyword), nil
	}

	var tokenResp TokenizeResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fallbackTokenize(keyword), nil
	}

	if tokenResp.Status != "success" || len(tokenResp.Tokens) == 0 {
		return fallbackTokenize(keyword), nil
	}

	// NLP สำเร็จ → ปิด circuit breaker
	nlpMarkSuccess()

	// กรอง tokens ที่มีความยาว > 1
	var filtered []string
	for _, t := range tokenResp.Tokens {
		t = strings.TrimSpace(t)
		if len(t) > 1 {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return fallbackTokenize(keyword), nil
	}

	return filtered, nil
}

// fallbackTokenize - ตัดคำแบบง่ายเมื่อ NLP ไม่พร้อม (split by space, ใช้ keyword ทั้งก้อนถ้าไม่มี space)
func fallbackTokenize(keyword string) []string {
	fields := strings.Fields(keyword)
	if len(fields) == 0 {
		return []string{keyword}
	}
	return fields
}

// searchProductsUnified - ค้นหาสินค้าจาก productbarcode
// SearchResult - ผลลัพธ์การค้นหาพร้อม pagination info
type SearchResult struct {
	Products []SearchProductItem
	Total    int // จำนวนทั้งหมดที่พบ (ก่อน pagination)
}

func searchProductsUnified(db *sql.DB, tokens []string, limit int, offset int) (*SearchResult, error) {
	// Build WHERE clause - all tokens must match somewhere
	var conditions []string
	for _, token := range tokens {
		// Escape single quotes
		token = strings.ReplaceAll(token, "'", "''")
		condition := fmt.Sprintf(`(
			itemcode ILIKE '%%%s%%' OR
			barcode ILIKE '%%%s%%' OR
			name0 ILIKE '%%%s%%' OR
			unitcode ILIKE '%%%s%%' OR
			unitname ILIKE '%%%s%%'
		)`, token, token, token, token, token)
		conditions = append(conditions, condition)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Single query: COUNT(*) OVER() ให้ total + pagination ใน query เดียว
	// ลด ILIKE scan จาก 2 ครั้งเหลือ 1 ครั้ง
	query := fmt.Sprintf(`
		SELECT itemcode, name0, COUNT(*) OVER() as total_count
		FROM (
			SELECT DISTINCT itemcode, name0
			FROM productbarcode
			WHERE %s
		) sub
		ORDER BY name0
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []SearchProductItem
	var total int
	for rows.Next() {
		var itemCode, name0 string
		var totalCount int
		if err := rows.Scan(&itemCode, &name0, &totalCount); err != nil {
			continue
		}
		total = totalCount
		products = append(products, SearchProductItem{
			ItemCode: itemCode,
			Name0:    name0,
			Units:    []SearchProductUnit{},
		})
	}

	return &SearchResult{
		Products: products,
		Total:    total,
	}, rows.Err()
}

// fetchAllUnitsForSearch - ดึงหน่วยทั้งหมดของสินค้า
func fetchAllUnitsForSearch(db *sql.DB, itemCodes []string) (map[string][]SearchProductUnit, error) {
	if len(itemCodes) == 0 {
		return nil, nil
	}

	// Build IN clause
	quoted := make([]string, len(itemCodes))
	for i, code := range itemCodes {
		code = strings.ReplaceAll(code, "'", "''")
		quoted[i] = fmt.Sprintf("'%s'", code)
	}

	query := fmt.Sprintf(`
		SELECT
			itemcode,
			barcode,
			COALESCE(unitcode, '') as unitcode,
			COALESCE(unitname, '') as unitname,
			COALESCE(barcoderefunitstand, 1) as unitstand,
			COALESCE(barcoderefunitdivide, 1) as unitdivide,
			COALESCE(price1, 0) as price1,
			COALESCE(price_retail, 0) as price_retail
		FROM productbarcode
		WHERE itemcode IN (%s)
		ORDER BY itemcode ASC, barcoderefunitstand DESC, barcoderefunitdivide ASC, barcode ASC
	`, strings.Join(quoted, ","))

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]SearchProductUnit)
	for rows.Next() {
		var itemCode, barcode, unitCode, unitName string
		var unitStand, unitDivide, price1, priceRetail float64

		if err := rows.Scan(&itemCode, &barcode, &unitCode, &unitName, &unitStand, &unitDivide, &price1, &priceRetail); err != nil {
			continue
		}

		result[itemCode] = append(result[itemCode], SearchProductUnit{
			Barcode:     barcode,
			UnitCode:    unitCode,
			UnitName:    unitName,
			UnitStand:   unitStand,
			UnitDivide:  unitDivide,
			Price1:      price1,
			PriceRetail: priceRetail,
		})
	}

	return result, rows.Err()
}

// calculateBalancesByLocation - คำนวณยอดคงเหลือแยกตาม itemcode → warehouse → location
// ใช้ iscalcstock = 1 และ calcflag เพื่อกำหนดทิศทาง (+/-) ของ stock movement
func calculateBalancesByLocation(db *sql.DB, itemCodes []string, whcodeFilter, locationcodeFilter string) (map[string]*SearchBalanceInfo, error) {
	if len(itemCodes) == 0 {
		return nil, nil
	}

	// Build IN clause
	quoted := make([]string, len(itemCodes))
	for i, code := range itemCodes {
		code = strings.ReplaceAll(code, "'", "''")
		quoted[i] = fmt.Sprintf("'%s'", code)
	}

	// Build WHERE clause with optional filters
	whereExtra := ""
	if whcodeFilter != "" {
		whcodeFilter = strings.ReplaceAll(whcodeFilter, "'", "''")
		whereExtra += fmt.Sprintf(" AND whcode = '%s'", whcodeFilter)
	}
	if locationcodeFilter != "" {
		locationcodeFilter = strings.ReplaceAll(locationcodeFilter, "'", "''")
		whereExtra += fmt.Sprintf(" AND locationcode = '%s'", locationcodeFilter)
	}

	// Query balance แยกตาม itemcode, whcode, locationcode
	// ใช้ transflag IN (...) เพื่อ filter เฉพาะ stock movement transactions
	// คูณ totalqty * calcflag เพื่อกำหนดทิศทาง: +1 = รับเข้า, -1 = เบิกออก
	transFlagList := myglobal.GetTransFlagsForQuery()

	query := fmt.Sprintf(`
		SELECT
			itemcode,
			COALESCE(whcode, '') as whcode,
			COALESCE(locationcode, '') as locationcode,
			SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance
		FROM docdetail
		WHERE transflag IN (%s)
			AND itemcode IN (%s)
			%s
		GROUP BY itemcode, whcode, locationcode
		HAVING SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) != 0
		ORDER BY itemcode, whcode, locationcode
	`, transFlagList, strings.Join(quoted, ","), whereExtra)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Structure: itemcode -> warehouse -> location -> balance
	type locationData struct {
		LocationCode string
		BalanceQty   float64
	}
	type warehouseData struct {
		WarehouseCode string
		BalanceQty    float64
		Locations     map[string]*locationData
	}
	type itemData struct {
		TotalBalance float64
		Warehouses   map[string]*warehouseData
	}

	itemMap := make(map[string]*itemData)

	for rows.Next() {
		var itemCode, whcode, locationcode string
		var balance float64

		if err := rows.Scan(&itemCode, &whcode, &locationcode, &balance); err != nil {
			continue
		}

		// Initialize item if not exists
		if itemMap[itemCode] == nil {
			itemMap[itemCode] = &itemData{
				TotalBalance: 0,
				Warehouses:   make(map[string]*warehouseData),
			}
		}
		itemMap[itemCode].TotalBalance += balance

		// Initialize warehouse (ใช้ whcode เป็น key)
		// หมายเหตุ: ระบบใช้ "X" เป็น default เมื่อไม่มี warehouse
		// Frontend จะแสดง "คลังหลัก" ถ้า warehouse_code เป็นค่าว่าง หรือ "X"
		whKey := whcode
		whDisplay := whcode
		if whcode == "" || whcode == "X" {
			whKey = "__default__" // ใช้ key พิเศษสำหรับ warehouse ที่ไม่ระบุ
			whDisplay = ""        // ส่งค่าว่างให้ Frontend แสดง "คลังหลัก"
		}

		if itemMap[itemCode].Warehouses[whKey] == nil {
			itemMap[itemCode].Warehouses[whKey] = &warehouseData{
				WarehouseCode: whDisplay,
				BalanceQty:    0,
				Locations:     make(map[string]*locationData),
			}
		}
		itemMap[itemCode].Warehouses[whKey].BalanceQty += balance

		// Add location (ถ้ามี และไม่ใช่ "X")
		if locationcode != "" && locationcode != "X" {
			if itemMap[itemCode].Warehouses[whKey].Locations[locationcode] == nil {
				itemMap[itemCode].Warehouses[whKey].Locations[locationcode] = &locationData{
					LocationCode: locationcode,
					BalanceQty:   0,
				}
			}
			itemMap[itemCode].Warehouses[whKey].Locations[locationcode].BalanceQty += balance
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert to result format
	result := make(map[string]*SearchBalanceInfo)
	for itemCode, item := range itemMap {
		balanceInfo := &SearchBalanceInfo{
			Total:      item.TotalBalance,
			Formatted:  "", // จะถูก format ภายหลังด้วย formatPackingString
			Warehouses: []SearchWarehouseBalance{},
		}

		for _, wh := range item.Warehouses {
			whBalance := SearchWarehouseBalance{
				WarehouseCode: wh.WarehouseCode,
				BalanceQty:    wh.BalanceQty,
				BalanceWord:   "", // จะถูก format ภายหลัง
				Locations:     []SearchLocationBalance{},
			}

			for _, loc := range wh.Locations {
				whBalance.Locations = append(whBalance.Locations, SearchLocationBalance{
					LocationCode: loc.LocationCode,
					BalanceQty:   loc.BalanceQty,
					BalanceWord:  "", // จะถูก format ภายหลัง
				})
			}

			balanceInfo.Warehouses = append(balanceInfo.Warehouses, whBalance)
		}

		result[itemCode] = balanceInfo
	}

	return result, nil
}

// pendingInfo — ค้างรับ + ค้างส่ง ที่คำนวณไว้ใน product table
type pendingInfo struct {
	RecvQty  float64
	RecvWord string
	SendQty  float64
	SendWord string
}

// fetchPendingFromProduct — ดึง pendingrecvqty, pendingsendqty จากตาราง product
func fetchPendingFromProduct(db *sql.DB, itemCodes []string) (map[string]*pendingInfo, error) {
	if len(itemCodes) == 0 {
		return nil, nil
	}

	quoted := make([]string, len(itemCodes))
	for i, code := range itemCodes {
		code = strings.ReplaceAll(code, "'", "''")
		quoted[i] = fmt.Sprintf("'%s'", code)
	}

	query := fmt.Sprintf(`
		SELECT
			itemcode,
			COALESCE(pendingrecvqty, 0) as pendingrecvqty,
			COALESCE(pendingrecvqtyword, '') as pendingrecvqtyword,
			COALESCE(pendingsendqty, 0) as pendingsendqty,
			COALESCE(pendingsendqtyword, '') as pendingsendqtyword
		FROM product
		WHERE itemcode IN (%s)
			AND (pendingrecvqty != 0 OR pendingsendqty != 0)
	`, strings.Join(quoted, ","))

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*pendingInfo)
	for rows.Next() {
		var itemCode, recvWord, sendWord string
		var recvQty, sendQty float64

		if err := rows.Scan(&itemCode, &recvQty, &recvWord, &sendQty, &sendWord); err != nil {
			continue
		}
		result[itemCode] = &pendingInfo{
			RecvQty:  recvQty,
			RecvWord: recvWord,
			SendQty:  sendQty,
			SendWord: sendWord,
		}
	}

	return result, rows.Err()
}

// fetchPackingUnits - ดึงข้อมูล packing units สำหรับคำนวณ BalanceWord
func fetchPackingUnits(db *sql.DB, itemCodes []string) (map[string][]models.ProductBarcodePackingStruct, error) {
	if len(itemCodes) == 0 {
		return nil, nil
	}

	// Build IN clause
	quoted := make([]string, len(itemCodes))
	for i, code := range itemCodes {
		code = strings.ReplaceAll(code, "'", "''")
		quoted[i] = fmt.Sprintf("'%s'", code)
	}

	query := fmt.Sprintf(`
		SELECT
			itemcode,
			COALESCE(unitname, unitcode, '') as unitname,
			COALESCE(barcoderefunitstand, 1) as unitstand,
			COALESCE(barcoderefunitdivide, 1) as unitdivide
		FROM productbarcode
		WHERE itemcode IN (%s)
			AND barcoderefunitstand > 0
			AND barcoderefunitdivide > 0
		ORDER BY itemcode, barcoderefunitstand DESC
	`, strings.Join(quoted, ","))

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]models.ProductBarcodePackingStruct)
	for rows.Next() {
		var itemCode, unitName string
		var unitStand, unitDivide float64

		if err := rows.Scan(&itemCode, &unitName, &unitStand, &unitDivide); err != nil {
			continue
		}

		result[itemCode] = append(result[itemCode], models.ProductBarcodePackingStruct{
			UnitName:             unitName,
			BarcodeRefUnitStand:  unitStand,
			BarcodeRefUnitDivide: unitDivide,
		})
	}

	return result, rows.Err()
}

// formatBalanceWithPacking - format balance ทุก level ด้วย packing units
func formatBalanceWithPacking(balanceInfo *SearchBalanceInfo, packingUnits []models.ProductBarcodePackingStruct, defaultUnitName string) {
	if balanceInfo == nil {
		return
	}

	// Format total balance
	if len(packingUnits) > 0 {
		balanceInfo.Formatted = myglobal.CalcStockQtyWord(balanceInfo.Total, packingUnits)
	} else if defaultUnitName != "" {
		balanceInfo.Formatted = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(balanceInfo.Total), defaultUnitName)
	} else {
		balanceInfo.Formatted = myglobal.CalcStockQtyWordCut(balanceInfo.Total)
	}

	// Format warehouse level
	for i := range balanceInfo.Warehouses {
		if len(packingUnits) > 0 {
			balanceInfo.Warehouses[i].BalanceWord = myglobal.CalcStockQtyWord(balanceInfo.Warehouses[i].BalanceQty, packingUnits)
		} else if defaultUnitName != "" {
			balanceInfo.Warehouses[i].BalanceWord = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(balanceInfo.Warehouses[i].BalanceQty), defaultUnitName)
		} else {
			balanceInfo.Warehouses[i].BalanceWord = myglobal.CalcStockQtyWordCut(balanceInfo.Warehouses[i].BalanceQty)
		}

		// Format location level
		for j := range balanceInfo.Warehouses[i].Locations {
			if len(packingUnits) > 0 {
				balanceInfo.Warehouses[i].Locations[j].BalanceWord = myglobal.CalcStockQtyWord(balanceInfo.Warehouses[i].Locations[j].BalanceQty, packingUnits)
			} else if defaultUnitName != "" {
				balanceInfo.Warehouses[i].Locations[j].BalanceWord = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(balanceInfo.Warehouses[i].Locations[j].BalanceQty), defaultUnitName)
			} else {
				balanceInfo.Warehouses[i].Locations[j].BalanceWord = myglobal.CalcStockQtyWordCut(balanceInfo.Warehouses[i].Locations[j].BalanceQty)
			}
		}
	}
}

// calculateSearchScore - คำนวณคะแนน relevance
func calculateSearchScore(product SearchProductItem, tokens []string) int {
	score := 0
	name0Lower := strings.ToLower(product.Name0)
	itemCodeLower := strings.ToLower(product.ItemCode)

	for _, token := range tokens {
		tokenLower := strings.ToLower(token)

		// Exact match itemcode
		if itemCodeLower == tokenLower {
			score += 1000
		}

		// Exact match barcode
		for _, unit := range product.Units {
			if strings.ToLower(unit.Barcode) == tokenLower {
				score += 900
				break
			}
		}

		// Prefix match name
		if strings.HasPrefix(name0Lower, tokenLower) {
			score += 250
		} else if strings.Contains(name0Lower, tokenLower) {
			score += 50
		}

		// Prefix match itemcode
		if strings.HasPrefix(itemCodeLower, tokenLower) {
			score += 200
		}
	}

	// Bonus for shorter name (more specific)
	if len(product.Name0) < 30 {
		score += 10
	}

	return score
}

package tools

import (
	"bytes"
	"context"
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
)

// Thai NLP client with timeout
var thaiNLPClient = &http.Client{
	Timeout: 5 * time.Second,
}

// TokenizeRequest for Thai NLP service
type TokenizeRequest struct {
	Text string `json:"text"`
}

// TokenizeResponse from Thai NLP service
type TokenizeResponse struct {
	Status string   `json:"status"`
	Tokens []string `json:"tokens"`
}

// ProductSearchRequest represents the search request
type ProductSearchRequest struct {
	ShopID         string `json:"shopid"`
	Keyword        string `json:"keyword"`
	WHCode         string `json:"whcode"`
	LocationCode   string `json:"locationcode"`
	Limit          int    `json:"limit"`
	IncludeBalance bool   `json:"include_balance"`
}

// ProductUnit represents a product unit with price
type ProductUnit struct {
	Barcode     string  `json:"barcode"`
	UnitCode    string  `json:"unitcode"`
	UnitName    string  `json:"unitname"`
	UnitStand   float64 `json:"unitstand"`
	UnitDivide  float64 `json:"unitdivide"`
	Price1      float64 `json:"price1"`
	PriceRetail float64 `json:"price_retail"`
}

// LocationBalance represents stock balance by location
type LocationBalance struct {
	LocationCode string  `json:"location_code"`
	BalanceQty   float64 `json:"balance_qty"`
	BalanceWord  string  `json:"balance_word"`
}

// WarehouseBalance represents stock balance by warehouse
type WarehouseBalance struct {
	WarehouseCode string            `json:"warehouse_code"`
	BalanceQty    float64           `json:"balance_qty"`
	BalanceWord   string            `json:"balance_word"`
	Locations     []LocationBalance `json:"locations,omitempty"`
}

// BalanceInfo represents complete balance information
type BalanceInfo struct {
	Total      float64            `json:"total"`
	Formatted  string             `json:"formatted"`
	Warehouses []WarehouseBalance `json:"warehouses,omitempty"`
}

// ProductItem represents a found product
type ProductItem struct {
	ItemCode string        `json:"itemcode"`
	Name0    string        `json:"name0"`
	Units    []ProductUnit `json:"units"`
	Balance  *BalanceInfo  `json:"balance,omitempty"`
	Score    int           `json:"score"`
}

// ProductSearchResponse represents the search response
type ProductSearchResponse struct {
	Status   string        `json:"status"`
	Count    int           `json:"count"`
	Products []ProductItem `json:"products"`
	Tokens   []string      `json:"tokens"`
}

// SearchProducts performs Thai full-text search with stock balance
func SearchProducts(ctx context.Context, shopID, keyword, whcode, locationcode string, limit int, includeBalance bool) (*ProductSearchResponse, error) {
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	// Default limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	logger.Info("[MCP SearchProducts] shopid=%s, keyword=%s, limit=%d, include_balance=%v",
		shopID, keyword, limit, includeBalance)

	// 1. Tokenize keyword using Thai NLP
	tokens, err := tokenizeKeyword(keyword)
	if err != nil {
		logger.Error("[MCP SearchProducts] Tokenize failed: %v", err)
		tokens = []string{keyword}
	}
	logger.Info("[MCP SearchProducts] Tokens: %v", tokens)

	// 2. Connect to database
	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// 3. Search products
	products, err := searchProducts(db, tokens, limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	logger.Info("[MCP SearchProducts] Found %d products", len(products))

	if len(products) == 0 {
		return &ProductSearchResponse{
			Status:   "success",
			Count:    0,
			Products: []ProductItem{},
			Tokens:   tokens,
		}, nil
	}

	// 4. Get all units for found products
	itemCodes := make([]string, len(products))
	for i, p := range products {
		itemCodes[i] = p.ItemCode
	}

	unitsMap, err := fetchAllUnits(db, itemCodes)
	if err != nil {
		logger.Error("[MCP SearchProducts] Fetch units failed: %v", err)
	} else {
		for i := range products {
			if units, ok := unitsMap[products[i].ItemCode]; ok {
				products[i].Units = units
			}
		}
	}

	// 5. Calculate stock balance if requested
	if includeBalance {
		balanceMap, err := calculateBalancesByLocation(db, itemCodes, whcode, locationcode)
		if err != nil {
			logger.Error("[MCP SearchProducts] Calculate balance failed: %v", err)
		}

		packingMap, err := fetchPackingUnits(db, itemCodes)
		if err != nil {
			logger.Error("[MCP SearchProducts] Fetch packing units failed: %v", err)
		}

		for i := range products {
			itemCode := products[i].ItemCode
			defaultUnitName := getDefaultUnitName(products[i].Units)

			if balanceInfo, ok := balanceMap[itemCode]; ok && balanceInfo != nil {
				packingUnits := packingMap[itemCode]
				formatBalanceWithPacking(balanceInfo, packingUnits, defaultUnitName)
				products[i].Balance = balanceInfo
			} else {
				products[i].Balance = &BalanceInfo{
					Total:      0,
					Formatted:  "0",
					Warehouses: []WarehouseBalance{},
				}
			}
		}
	}

	// 6. Calculate scores and rank
	for i := range products {
		products[i].Score = calculateScore(products[i], tokens)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].Score > products[j].Score
	})

	return &ProductSearchResponse{
		Status:   "success",
		Count:    len(products),
		Products: products,
		Tokens:   tokens,
	}, nil
}

// tokenizeKeyword calls Thai NLP service
func tokenizeKeyword(keyword string) ([]string, error) {
	thaiNLPURL := getThaiNLPURL()

	requestBody, err := json.Marshal(TokenizeRequest{Text: keyword})
	if err != nil {
		return nil, err
	}

	resp, err := thaiNLPClient.Post(
		fmt.Sprintf("%s/tokenize", thaiNLPURL),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenizeResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Status != "success" || len(tokenResp.Tokens) == 0 {
		return strings.Fields(keyword), nil
	}

	var filtered []string
	for _, t := range tokenResp.Tokens {
		t = strings.TrimSpace(t)
		if len(t) > 1 {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return strings.Fields(keyword), nil
	}

	return filtered, nil
}

func getThaiNLPURL() string {
	return "http://thai-tokenizer:5000"
}

// searchProducts searches for products using ILIKE
func searchProducts(db *sql.DB, tokens []string, limit int) ([]ProductItem, error) {
	var conditions []string
	for _, token := range tokens {
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

	query := fmt.Sprintf(`
		SELECT DISTINCT itemcode, name0
		FROM productbarcode
		WHERE %s
		ORDER BY name0
		LIMIT %d
	`, whereClause, limit)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []ProductItem
	for rows.Next() {
		var itemCode, name0 string
		if err := rows.Scan(&itemCode, &name0); err != nil {
			continue
		}
		products = append(products, ProductItem{
			ItemCode: itemCode,
			Name0:    name0,
			Units:    []ProductUnit{},
		})
	}

	return products, rows.Err()
}

// fetchAllUnits gets all units for products
func fetchAllUnits(db *sql.DB, itemCodes []string) (map[string][]ProductUnit, error) {
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

	result := make(map[string][]ProductUnit)
	for rows.Next() {
		var itemCode, barcode, unitCode, unitName string
		var unitStand, unitDivide, price1, priceRetail float64

		if err := rows.Scan(&itemCode, &barcode, &unitCode, &unitName, &unitStand, &unitDivide, &price1, &priceRetail); err != nil {
			continue
		}

		result[itemCode] = append(result[itemCode], ProductUnit{
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

// calculateBalancesByLocation calculates stock balance by warehouse/location
func calculateBalancesByLocation(db *sql.DB, itemCodes []string, whcodeFilter, locationcodeFilter string) (map[string]*BalanceInfo, error) {
	if len(itemCodes) == 0 {
		return nil, nil
	}

	quoted := make([]string, len(itemCodes))
	for i, code := range itemCodes {
		code = strings.ReplaceAll(code, "'", "''")
		quoted[i] = fmt.Sprintf("'%s'", code)
	}

	whereExtra := ""
	if whcodeFilter != "" {
		whcodeFilter = strings.ReplaceAll(whcodeFilter, "'", "''")
		whereExtra += fmt.Sprintf(" AND whcode = '%s'", whcodeFilter)
	}
	if locationcodeFilter != "" {
		locationcodeFilter = strings.ReplaceAll(locationcodeFilter, "'", "''")
		whereExtra += fmt.Sprintf(" AND locationcode = '%s'", locationcodeFilter)
	}

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

		if itemMap[itemCode] == nil {
			itemMap[itemCode] = &itemData{
				TotalBalance: 0,
				Warehouses:   make(map[string]*warehouseData),
			}
		}
		itemMap[itemCode].TotalBalance += balance

		whKey := whcode
		whDisplay := whcode
		if whcode == "" || whcode == "X" {
			whKey = "__default__"
			whDisplay = ""
		}

		if itemMap[itemCode].Warehouses[whKey] == nil {
			itemMap[itemCode].Warehouses[whKey] = &warehouseData{
				WarehouseCode: whDisplay,
				BalanceQty:    0,
				Locations:     make(map[string]*locationData),
			}
		}
		itemMap[itemCode].Warehouses[whKey].BalanceQty += balance

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

	result := make(map[string]*BalanceInfo)
	for itemCode, item := range itemMap {
		balanceInfo := &BalanceInfo{
			Total:      item.TotalBalance,
			Formatted:  "",
			Warehouses: []WarehouseBalance{},
		}

		for _, wh := range item.Warehouses {
			whBalance := WarehouseBalance{
				WarehouseCode: wh.WarehouseCode,
				BalanceQty:    wh.BalanceQty,
				BalanceWord:   "",
				Locations:     []LocationBalance{},
			}

			for _, loc := range wh.Locations {
				whBalance.Locations = append(whBalance.Locations, LocationBalance{
					LocationCode: loc.LocationCode,
					BalanceQty:   loc.BalanceQty,
					BalanceWord:  "",
				})
			}

			balanceInfo.Warehouses = append(balanceInfo.Warehouses, whBalance)
		}

		result[itemCode] = balanceInfo
	}

	return result, nil
}

// fetchPackingUnits gets packing units for formatting
func fetchPackingUnits(db *sql.DB, itemCodes []string) (map[string][]models.ProductBarcodePackingStruct, error) {
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

// getDefaultUnitName finds the default unit name (ratio = 1)
func getDefaultUnitName(units []ProductUnit) string {
	if len(units) == 0 {
		return ""
	}

	for _, u := range units {
		ratio := u.UnitStand / u.UnitDivide
		if ratio == 1.0 {
			if u.UnitName != "" {
				return u.UnitName
			}
			return u.UnitCode
		}
	}

	lastUnit := units[len(units)-1]
	if lastUnit.UnitName != "" {
		return lastUnit.UnitName
	}
	return lastUnit.UnitCode
}

// formatBalanceWithPacking formats balance at all levels
func formatBalanceWithPacking(balanceInfo *BalanceInfo, packingUnits []models.ProductBarcodePackingStruct, defaultUnitName string) {
	if balanceInfo == nil {
		return
	}

	if len(packingUnits) > 0 {
		balanceInfo.Formatted = myglobal.CalcStockQtyWord(balanceInfo.Total, packingUnits)
	} else if defaultUnitName != "" {
		balanceInfo.Formatted = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(balanceInfo.Total), defaultUnitName)
	} else {
		balanceInfo.Formatted = myglobal.CalcStockQtyWordCut(balanceInfo.Total)
	}

	for i := range balanceInfo.Warehouses {
		if len(packingUnits) > 0 {
			balanceInfo.Warehouses[i].BalanceWord = myglobal.CalcStockQtyWord(balanceInfo.Warehouses[i].BalanceQty, packingUnits)
		} else if defaultUnitName != "" {
			balanceInfo.Warehouses[i].BalanceWord = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(balanceInfo.Warehouses[i].BalanceQty), defaultUnitName)
		} else {
			balanceInfo.Warehouses[i].BalanceWord = myglobal.CalcStockQtyWordCut(balanceInfo.Warehouses[i].BalanceQty)
		}

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

// calculateScore calculates relevance score
func calculateScore(product ProductItem, tokens []string) int {
	score := 0
	name0Lower := strings.ToLower(product.Name0)
	itemCodeLower := strings.ToLower(product.ItemCode)

	for _, token := range tokens {
		tokenLower := strings.ToLower(token)

		if itemCodeLower == tokenLower {
			score += 1000
		}

		for _, unit := range product.Units {
			if strings.ToLower(unit.Barcode) == tokenLower {
				score += 900
				break
			}
		}

		if strings.HasPrefix(name0Lower, tokenLower) {
			score += 250
		} else if strings.Contains(name0Lower, tokenLower) {
			score += 50
		}

		if strings.HasPrefix(itemCodeLower, tokenLower) {
			score += 200
		}
	}

	if len(product.Name0) < 30 {
		score += 10
	}

	return score
}

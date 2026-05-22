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
	"smlcloudplatform/internal/goapi/myollama"
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
	ShopID string `json:"shopid"`
	Keyword string `json:"keyword"`
	WHCode string `json:"whcode"`
	LocationCode string `json:"locationcode"`
	Limit int    `json:"limit"`
	IncludeBalance bool   `json:"include_balance"`
}

// ProductUnit represents a product unit with price
type ProductUnit struct {
	Barcode string  `json:"barcode"`
	UnitCode string  `json:"unitcode"`
	UnitName string  `json:"unit_name"`
	UnitStand float64 `json:"unitstand"`
	UnitDivide float64 `json:"unitdivide"`
	Price1 float64 `json:"price1"`
	PriceRetail float64 `json:"price_retail"`
}

// LocationBalance represents stock balance by location
type LocationBalance struct {
	LocationCode string  `json:"location_code"`
	BalanceQty float64 `json:"balance_qty"`
	BalanceWord string  `json:"balance_word"`
}

// WarehouseBalance represents stock balance by warehouse
type WarehouseBalance struct {
	WarehouseCode string            `json:"warehouse_code"`
	BalanceQty float64           `json:"balance_qty"`
	BalanceWord string            `json:"balance_word"`
	Locations []LocationBalance `json:"locations,omitempty"`
}

// BalanceInfo represents complete balance information
type BalanceInfo struct {
	Total float64            `json:"total"`
	Formatted string             `json:"formatted"`
	Warehouses []WarehouseBalance `json:"warehouses,omitempty"`
}

// ProductItem represents a found product
type ProductItem struct {
	ItemCode string        `json:"itemcode"`
	Name0 string        `json:"name0"`
	Units []ProductUnit `json:"units"`
	Balance *BalanceInfo  `json:"balance,omitempty"`
	Score int           `json:"score"`
}

// ProductSearchResponse represents the search response
type ProductSearchResponse struct {
	Status string        `json:"status"`
	Count int           `json:"count"`
	Products []ProductItem `json:"products"`
	Tokens []string      `json:"tokens"`
}

// SearchProducts performs smart Thai search with alias expansion + fuzzy fallback
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

	// 3. Three-phase search (no manual aliases — RAG handles cross-script)
	// Phase 1: ILIKE (fast, uses GIN trigram index)
	products, err := searchProductsILIKE(db, tokens, limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Phase 2: Fuzzy fallback if ILIKE returns few results
	if len(products) < 3 && len(tokens) > 0 {
		fuzzyProducts, fuzzyErr := searchProductsFuzzy(db, keyword, limit)
		if fuzzyErr == nil && len(fuzzyProducts) > len(products) {
			products = mergeProducts(products, fuzzyProducts, limit)
		}
	}

	// Phase 3: Vector similarity search (semantic) — fallback when text search finds little
	if len(products) < 3 {
		vectorProducts, vecErr := searchProductsVector(db, keyword, limit)
		if vecErr == nil && len(vectorProducts) > 0 {
			products = mergeProducts(products, vectorProducts, limit)
			logger.Info("[MCP SearchProducts] Vector search added %d products", len(vectorProducts))
		}
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

// tokenizeKeyword calls Thai NLP service with smart whitespace pre-split
func tokenizeKeyword(keyword string) ([]string, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("empty keyword")
	}

	// Step 1: Pre-split by whitespace to handle mixed Thai+English like "สี TOA"
	parts := strings.Fields(keyword)

	seen := make(map[string]bool)
	var allTokens []string
	addToken := func(t string) {
		t = strings.TrimSpace(t)
		lower := strings.ToLower(t)
		if lower == "" {
			return
		}
		// Filter out single ASCII chars but keep Thai chars (multi-byte)
		if len(t) == 1 {
			return
		}
		if !seen[lower] {
			seen[lower] = true
			allTokens = append(allTokens, t)
		}
	}

	// Step 2: For each part, try NLP tokenization for Thai text, keep English as-is
	for _, part := range parts {
		if isAsciiOnly(part) {
			// English/number — use as-is
			addToken(part)
		} else {
			// Thai text — try NLP tokenization
			nlpTokens, err := callThaiNLP(part)
			if err != nil || len(nlpTokens) == 0 {
				addToken(part)
			} else {
				for _, t := range nlpTokens {
					addToken(t)
				}
			}
		}
	}

	if len(allTokens) == 0 {
		return []string{keyword}, nil
	}

	return allTokens, nil
}

// callThaiNLP calls the Thai NLP tokenizer service for a single text part
func callThaiNLP(text string) ([]string, error) {
	thaiNLPURL := getThaiNLPURL()

	requestBody, err := json.Marshal(TokenizeRequest{Text: text})
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
		return nil, nil
	}

	var filtered []string
	for _, t := range tokenResp.Tokens {
		t = strings.TrimSpace(t)
		if len(t) > 1 {
			filtered = append(filtered, t)
		}
	}

	return filtered, nil
}

// isAsciiOnly checks if string contains only ASCII characters
func isAsciiOnly(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

func getThaiNLPURL() string {
	return "http://thai-tokenizer:5000"
}

// searchFields — columns to search (includes brand, category, group names)
var searchFields = []string{
	"itemcode", "barcode", "name0", "unitcode", "unit_name",
	"brandnames", "category_names", "group_names",
}

// mergeProducts combines two product lists, deduplicating by ItemCode
func mergeProducts(existing, additional []ProductItem, limit int) []ProductItem {
	seen := make(map[string]bool)
	for _, p := range existing {
		seen[p.ItemCode] = true
	}
	for _, p := range additional {
		if !seen[p.ItemCode] {
			existing = append(existing, p)
			seen[p.ItemCode] = true
		}
	}
	if len(existing) > limit {
		existing = existing[:limit]
	}
	return existing
}

// buildTokenCondition builds ILIKE OR condition for one token across all search fields
func buildTokenCondition(token string) string {
	token = strings.ReplaceAll(token, "'", "''")
	var parts []string
	for _, field := range searchFields {
		parts = append(parts, fmt.Sprintf("%s ILIKE '%%%s%%'", field, token))
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// searchProductsILIKE searches using ILIKE with OR + match count ranking
func searchProductsILIKE(db *sql.DB, tokens []string, limit int) ([]ProductItem, error) {
	var orConditions []string
	var matchCountParts []string

	for _, token := range tokens {
		orConditions = append(orConditions, buildTokenCondition(token))

		// Count how many tokens match per row for ranking
		token = strings.ReplaceAll(token, "'", "''")
		var caseParts []string
		for _, field := range searchFields {
			caseParts = append(caseParts, fmt.Sprintf("%s ILIKE '%%%s%%'", field, token))
		}
		matchPart := fmt.Sprintf("CASE WHEN (%s) THEN 1 ELSE 0 END", strings.Join(caseParts, " OR "))
		matchCountParts = append(matchCountParts, matchPart)
	}

	whereClause := strings.Join(orConditions, " OR ")
	matchCountExpr := strings.Join(matchCountParts, " + ")

	// Use subquery to calculate match_count with all columns, then deduplicate
	query := fmt.Sprintf(`
		SELECT itemcode, name0 FROM (
			SELECT itemcode, name0, (%s) as match_count,
				ROW_NUMBER() OVER (PARTITION BY itemcode ORDER BY (%s) DESC) as rn
			FROM productbarcode
			WHERE %s
		) ranked
		WHERE rn = 1
		ORDER BY match_count DESC, name0
		LIMIT %d
	`, matchCountExpr, matchCountExpr, whereClause, limit)

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

// searchProductsFuzzy uses pg_trgm similarity for fuzzy matching (typo-tolerant)
func searchProductsFuzzy(db *sql.DB, keyword string, limit int) ([]ProductItem, error) {
	keyword = strings.ReplaceAll(keyword, "'", "''")

	query := fmt.Sprintf(`
		SELECT itemcode, name0 FROM (
			SELECT DISTINCT ON (itemcode) itemcode, name0,
				similarity(name0, '%s') as sim
			FROM productbarcode
			WHERE similarity(name0, '%s') > 0.15
			   OR similarity(barcode, '%s') > 0.15
			   OR similarity(itemcode, '%s') > 0.15
			   OR similarity(brandnames, '%s') > 0.15
			   OR similarity(groupnames, '%s') > 0.15
			ORDER BY itemcode
		) sub
		ORDER BY sim DESC, name0
		LIMIT %d
	`, keyword, keyword, keyword, keyword, keyword, keyword, limit)

	rows, err := db.Query(query)
	if err != nil {
		// pg_trgm may not be available — gracefully degrade
		logger.Error("[MCP SearchProducts] Fuzzy search failed (pg_trgm?): %v", err)
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

// searchProductsVector — Phase 3: semantic search ด้วย pgvector
// แปลง keyword → embedding แล้วหา nearest neighbors (cosine similarity)
func searchProductsVector(db *sql.DB, keyword string, limit int) ([]ProductItem, error) {
	// ตรวจว่า pgvector + column พร้อม
	var colExists bool
	if err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'productbarcode' AND column_name = 'name_embedding'
		)
	`).Scan(&colExists); err != nil || !colExists {
		return nil, nil // ยังไม่มี column → skip quietly
	}

	// ตรวจว่ามี embedding อยู่จริง (ไม่ใช่ column ว่าง)
	var hasEmbeddings bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM productbarcode WHERE name_embedding IS NOT NULL LIMIT 1)`).Scan(&hasEmbeddings); err != nil || !hasEmbeddings {
		return nil, nil
	}

	// สร้าง embedding จาก keyword
	embedding, err := myollama.GenerateSingleEmbedding(keyword)
	if err != nil {
		logger.Warn("[MCP SearchProducts] Vector search skipped (Ollama error): %v", err)
		return nil, nil // graceful fallback
	}

	// แปลง embedding → pgvector string
	vecStr := float32SliceToVectorStr(embedding)

	// ค้นหา nearest neighbors ด้วย cosine distance (<=>)
	query := fmt.Sprintf(`
		SELECT itemcode, name0, (name_embedding <=> $1::vector) as distance
		FROM (
			SELECT DISTINCT ON (itemcode) itemcode, name0, name_embedding
			FROM productbarcode
			WHERE name_embedding IS NOT NULL
			ORDER BY itemcode, id
		) sub
		ORDER BY name_embedding <=> $1::vector
		LIMIT %d
	`, limit)

	rows, err := db.Query(query, vecStr)
	if err != nil {
		logger.Warn("[MCP SearchProducts] Vector query failed: %v", err)
		return nil, nil
	}
	defer rows.Close()

	var products []ProductItem
	for rows.Next() {
		var itemCode, name0 string
		var distance float64
		if err := rows.Scan(&itemCode, &name0, &distance); err != nil {
			continue
		}
		// cosine distance < 0.5 ถือว่าใกล้เคียงพอ (0=identical, 2=opposite)
		if distance > 0.5 {
			continue
		}
		products = append(products, ProductItem{
			ItemCode: itemCode,
			Name0:    name0,
			Units:    []ProductUnit{},
			Score:    int((1.0 - distance) * 500), // แปลง distance → score
		})
	}

	if len(products) > 0 {
		logger.Info("[MCP SearchProducts] Vector search found %d similar products (best distance=%.3f)", len(products), 1.0-float64(products[0].Score)/500.0)
	}

	return products, rows.Err()
}

// float32SliceToVectorStr แปลง []float32 → "[0.1,0.2,...]" สำหรับ pgvector
func float32SliceToVectorStr(v []float32) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, "%g", f)
	}
	sb.WriteByte(']')
	return sb.String()
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

// calculateScore calculates relevance score with multi-token match bonus
func calculateScore(product ProductItem, tokens []string) int {
	score := 0
	matchCount := 0
	name0Lower := strings.ToLower(product.Name0)
	itemCodeLower := strings.ToLower(product.ItemCode)

	for _, token := range tokens {
		tokenLower := strings.ToLower(token)
		matched := false

		if itemCodeLower == tokenLower {
			score += 1000
			matched = true
		}

		for _, unit := range product.Units {
			if strings.ToLower(unit.Barcode) == tokenLower {
				score += 900
				matched = true
				break
			}
		}

		if strings.HasPrefix(name0Lower, tokenLower) {
			score += 250
			matched = true
		} else if strings.Contains(name0Lower, tokenLower) {
			score += 50
			matched = true
		}

		if strings.HasPrefix(itemCodeLower, tokenLower) {
			score += 200
			matched = true
		}

		if matched {
			matchCount++
		}
	}

	// Bonus: products matching more tokens rank much higher
	if len(tokens) > 1 {
		score += matchCount * 500
	}

	if len(product.Name0) < 30 {
		score += 10
	}

	return score
}

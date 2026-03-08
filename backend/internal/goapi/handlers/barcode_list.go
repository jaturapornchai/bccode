package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// ==================== Barcode List (PostgreSQL) ====================

// BarcodeListRequest — Request body สำหรับดึงรายการบาร์โค้ดจาก PostgreSQL
type BarcodeListRequest struct {
	ShopID       string   `json:"shopid"`
	Keyword      string   `json:"keyword"`
	GroupCode    string   `json:"groupcode"`
	BrandCode    string   `json:"brandcode"`
	CategoryCode string   `json:"categorycode"`
	ClassCode    string   `json:"classcode"`
	DesignCode   string   `json:"designcode"`
	GradeCode    string   `json:"gradecode"`
	ModelCode    string   `json:"modelcode"`
	PatternCode  string   `json:"patterncode"`
	PriceMin     *float64 `json:"price_min"`
	PriceMax     *float64 `json:"price_max"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
	SortField    string   `json:"sort_field"`
	SortOrder    string   `json:"sort_order"`
}

// BarcodeListItem — รายการบาร์โค้ดสำหรับแสดงใน list (lightweight)
type BarcodeListItem struct {
	GuidFixed  string  `json:"guidfixed"`
	Barcode    string  `json:"barcode"`
	Name       string  `json:"name0"`
	UnitCode   string  `json:"unitcode"`
	UnitName   string  `json:"unitname"`
	ItemCode   string  `json:"itemcode"`
	GroupCode  string  `json:"groupcode"`
	GroupNames string  `json:"groupnames"`
	Price1           float64 `json:"price1"`
	ImageUri         string  `json:"imageuri"`
	StandValue       float64 `json:"standvalue"`
	DivideValue      float64 `json:"dividevalue"`
	UnitCount        int     `json:"unit_count"`
	AllUnitNames     string  `json:"all_unit_names"`
	BalanceQty       float64 `json:"balance_qty"`
	BalanceFormatted string  `json:"balance_formatted"`
}

// BarcodeListHandler — Handler สำหรับดึงรายการบาร์โค้ดจาก PostgreSQL
// POST /api/product/barcode/list
func BarcodeListHandler(c echo.Context) error {
	var req BarcodeListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: shopid",
		})
	}

	// Defaults
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Validate sort field (whitelist เพื่อป้องกัน SQL injection)
	useRelevanceSort := req.SortField == "relevance"
	validSortFields := map[string]string{
		"barcode": "pb.barcode", "name0": "pb.name0", "itemcode": "pb.itemcode",
		"groupnames": "pb.groupnames", "price1": "pb.price1", "unitname": "pb.unitname",
	}
	if !useRelevanceSort {
		if mapped, ok := validSortFields[req.SortField]; ok {
			req.SortField = mapped
		} else {
			req.SortField = "pb.barcode"
		}
	}
	sortOrder := "ASC"
	if strings.ToLower(req.SortOrder) == "desc" {
		sortOrder = "DESC"
	}

	// Connect to shop's PostgreSQL
	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		logger.Error("BarcodeListHandler: connect PG: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Database connection failed",
		})
	}

	// แยกคำค้นหา + expand aliases
	var searchWords []string
	if req.Keyword != "" {
		searchWords = strings.Fields(req.Keyword)
		searchWords = expandSearchAliases(db, searchWords)
	}

	// Build filter conditions (ไม่รวม keyword — จะใช้แยกใน two-phase search)
	filterConditions, filterArgs, filterArgIdx := buildFilterConditions(req)

	// ===== Two-Phase Search =====
	var items []BarcodeListItem
	var total int

	if len(searchWords) > 0 {
		// Phase 1: ILIKE search (เร็ว — ใช้ GIN trigram index)
		items, total, err = searchILIKE(db, searchWords, filterConditions, filterArgs, filterArgIdx, req)
		if err != nil {
			logger.Error("BarcodeListHandler: ILIKE search: %v", err)
		}

		// Phase 2: Fuzzy fallback ถ้า ILIKE ได้ผลน้อย
		if total < 3 && len(searchWords) == 1 {
			fuzzyItems, fuzzyTotal, fuzzyErr := searchFuzzy(db, searchWords[0], filterConditions, filterArgs, filterArgIdx, req)
			if fuzzyErr == nil && fuzzyTotal > total {
				items = fuzzyItems
				total = fuzzyTotal
			}
		}
	} else {
		// ไม่มี keyword — query ปกติ
		items, total, err = searchNoKeyword(db, filterConditions, filterArgs, filterArgIdx, req, sortOrder)
		if err != nil {
			logger.Error("BarcodeListHandler: no keyword search: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Query failed",
			})
		}
	}

	// Relevance scoring + sort
	if useRelevanceSort && len(searchWords) > 0 {
		items = sortByRelevance(items, searchWords)
	}

	// เพิ่มข้อมูลหน่วยนับ (unit_count + all_unit_names)
	items = enrichUnitInfo(db, items)

	// แปลงเป็น format ที่ Flutter เข้าใจ
	result := formatBarcodeItems(items)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
		"total":   total,
	})
}

// ==================== Filter Builder ====================

func buildFilterConditions(req BarcodeListRequest) ([]string, []interface{}, int) {
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	filterFields := []struct {
		code   string
		column string
	}{
		{req.GroupCode, "pb.groupcode"},
		{req.BrandCode, "pb.brandcode"},
		{req.CategoryCode, "pb.categorycode"},
		{req.ClassCode, "pb.classcode"},
		{req.DesignCode, "pb.designcode"},
		{req.GradeCode, "pb.gradecode"},
		{req.ModelCode, "pb.modelcode"},
		{req.PatternCode, "pb.patterncode"},
	}
	for _, f := range filterFields {
		if f.code != "" {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", f.column, argIdx))
			args = append(args, f.code)
			argIdx++
		}
	}

	if req.PriceMin != nil {
		conditions = append(conditions, fmt.Sprintf("pb.price1 >= $%d", argIdx))
		args = append(args, *req.PriceMin)
		argIdx++
	}
	if req.PriceMax != nil {
		conditions = append(conditions, fmt.Sprintf("pb.price1 <= $%d", argIdx))
		args = append(args, *req.PriceMax)
		argIdx++
	}

	return conditions, args, argIdx
}

// ==================== Search: ILIKE (Phase 1) ====================

func searchILIKE(db *sql.DB, words []string, filterConds []string, filterArgs []interface{}, argIdx int, req BarcodeListRequest) ([]BarcodeListItem, int, error) {
	conditions := append([]string{}, filterConds...)
	args := append([]interface{}{}, filterArgs...)

	for _, word := range words {
		kw := "%" + word + "%"
		conditions = append(conditions, fmt.Sprintf(
			"(pb.barcode ILIKE $%d OR pb.name0 ILIKE $%d OR pb.itemcode ILIKE $%d OR pb.groupnames ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx,
		))
		args = append(args, kw)
		argIdx++
	}

	return executeSearch(db, conditions, args, argIdx, req, "pb.barcode", "ASC")
}

// ==================== Search: Fuzzy (Phase 2) ====================

func searchFuzzy(db *sql.DB, keyword string, filterConds []string, filterArgs []interface{}, argIdx int, req BarcodeListRequest) ([]BarcodeListItem, int, error) {
	conditions := append([]string{}, filterConds...)
	args := append([]interface{}{}, filterArgs...)

	// ใช้ pg_trgm similarity — threshold 0.15 สำหรับภาษาไทย
	conditions = append(conditions, fmt.Sprintf(
		"(similarity(pb.name0, $%d) > 0.15 OR similarity(pb.barcode, $%d) > 0.15 OR similarity(pb.itemcode, $%d) > 0.15)",
		argIdx, argIdx, argIdx,
	))
	args = append(args, keyword)
	argIdx++

	return executeSearch(db, conditions, args, argIdx, req, fmt.Sprintf("similarity(pb.name0, '%s')", strings.ReplaceAll(keyword, "'", "''")), "DESC")
}

// ==================== Search: No Keyword ====================

func searchNoKeyword(db *sql.DB, filterConds []string, filterArgs []interface{}, argIdx int, req BarcodeListRequest, sortOrder string) ([]BarcodeListItem, int, error) {
	sortField := req.SortField
	if sortField == "relevance" {
		sortField = "pb.barcode"
	}
	return executeSearch(db, filterConds, filterArgs, argIdx, req, sortField, sortOrder)
}

// ==================== Execute Search ====================

func executeSearch(db *sql.DB, conditions []string, args []interface{}, argIdx int, req BarcodeListRequest, orderField string, orderDir string) ([]BarcodeListItem, int, error) {
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM productbarcode pb %s", whereClause)
	var total int
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		logger.Error("BarcodeListHandler: count: %v", err)
		total = 0
	}

	// Data
	dataQuery := fmt.Sprintf(`
		SELECT COALESCE(pb.guidfixed,''), pb.barcode, COALESCE(pb.name0,''), COALESCE(pb.unitcode,''), COALESCE(pb.unitname,''),
			   COALESCE(pb.itemcode,''), COALESCE(pb.groupcode,''), COALESCE(pb.groupnames,''),
			   COALESCE(pb.price1,0), COALESCE(pb.imageuri,''),
			   COALESCE(pb.standvalue,0), COALESCE(pb.dividevalue,0),
			   COALESCE(p.balanceqty,0), COALESCE(p.balanceqtyword,'')
		FROM productbarcode pb
		LEFT JOIN product p ON p.itemcode = pb.itemcode
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderField, orderDir, argIdx, argIdx+1)

	args = append(args, req.Limit, req.Offset)

	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []BarcodeListItem
	for rows.Next() {
		var item BarcodeListItem
		var price sql.NullFloat64
		if err := rows.Scan(
			&item.GuidFixed, &item.Barcode, &item.Name, &item.UnitCode, &item.UnitName,
			&item.ItemCode, &item.GroupCode, &item.GroupNames,
			&price, &item.ImageUri,
			&item.StandValue, &item.DivideValue,
			&item.BalanceQty, &item.BalanceFormatted,
		); err != nil {
			logger.Error("BarcodeListHandler: scan: %v", err)
			continue
		}
		if price.Valid {
			item.Price1 = price.Float64
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

// ==================== Relevance Scoring ====================

func calculateBarcodeSearchScore(item BarcodeListItem, words []string) int {
	score := 0
	nameLower := strings.ToLower(item.Name)
	barcodeLower := strings.ToLower(item.Barcode)
	itemCodeLower := strings.ToLower(item.ItemCode)

	for _, word := range words {
		w := strings.ToLower(word)

		// Exact match barcode: +1000
		if barcodeLower == w {
			score += 1000
		}
		// Exact match itemcode: +900
		if itemCodeLower == w {
			score += 900
		}
		// Prefix match name: +250
		if strings.HasPrefix(nameLower, w) {
			score += 250
		} else if strings.Contains(nameLower, w) {
			score += 50
		}
		// Prefix match itemcode: +200
		if strings.HasPrefix(itemCodeLower, w) {
			score += 200
		}
		// Prefix match barcode: +150
		if strings.HasPrefix(barcodeLower, w) {
			score += 150
		}
	}

	// Bonus for shorter name
	if len(item.Name) < 30 {
		score += 10
	}

	return score
}

func sortByRelevance(items []BarcodeListItem, words []string) []BarcodeListItem {
	sort.SliceStable(items, func(i, j int) bool {
		scoreI := calculateBarcodeSearchScore(items[i], words)
		scoreJ := calculateBarcodeSearchScore(items[j], words)
		return scoreI > scoreJ
	})
	return items
}

// ==================== Alias Expansion ====================

// expandSearchAliases — ตรวจแต่ละคำกับ search_aliases table
// เช่น "ทีโอเอ" → "TOA", "มากิต้า" → "MAKITA"
func expandSearchAliases(db *sql.DB, words []string) []string {
	result := make([]string, len(words))
	copy(result, words)

	for i, word := range result {
		var target string
		err := db.QueryRow(
			"SELECT target FROM search_aliases WHERE LOWER(alias) = LOWER($1) LIMIT 1",
			word,
		).Scan(&target)
		if err == nil && target != "" {
			logger.Info("BarcodeListHandler: alias expand: %s → %s", word, target)
			result[i] = target
		}
	}

	return result
}

// ==================== Enrich Unit Info ====================

// enrichUnitInfo — เพิ่มข้อมูล unit_count + all_unit_names (พร้อมอัตราส่วน) ให้แต่ละ item
func enrichUnitInfo(db *sql.DB, items []BarcodeListItem) []BarcodeListItem {
	if len(items) == 0 {
		return items
	}

	// Collect unique itemcodes
	itemcodeSet := map[string]bool{}
	for _, item := range items {
		if item.ItemCode != "" {
			itemcodeSet[item.ItemCode] = true
		}
	}

	if len(itemcodeSet) == 0 {
		for i := range items {
			items[i].UnitCount = 1
		}
		return items
	}

	// Build IN clause
	codes := make([]interface{}, 0, len(itemcodeSet))
	placeholders := make([]string, 0, len(itemcodeSet))
	idx := 1
	for code := range itemcodeSet {
		codes = append(codes, code)
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		idx++
	}

	// Query: ดึง unitname + อัตราส่วน per barcode per itemcode
	query := fmt.Sprintf(
		`SELECT itemcode, unitname,
			COALESCE(barcoderefunitstand, 0) as stand,
			COALESCE(barcoderefunitdivide, 0) as divide,
			COALESCE(barcoderef, '') as barcoderef
		FROM productbarcode
		WHERE itemcode IN (%s)
		ORDER BY itemcode, barcoderefunitstand ASC`,
		strings.Join(placeholders, ","),
	)

	rows, err := db.Query(query, codes...)
	if err != nil {
		logger.Error("enrichUnitInfo: query: %v", err)
		for i := range items {
			items[i].UnitCount = 1
		}
		return items
	}
	defer rows.Close()

	// เก็บ unit info per itemcode
	type unitDetail struct {
		UnitName   string
		Stand      float64
		Divide     float64
		BarcodeRef string
	}
	detailMap := map[string][]unitDetail{}

	for rows.Next() {
		var code, unitName, barcodeRef string
		var stand, divide float64
		if err := rows.Scan(&code, &unitName, &stand, &divide, &barcodeRef); err != nil {
			continue
		}
		detailMap[code] = append(detailMap[code], unitDetail{
			UnitName:   unitName,
			Stand:      stand,
			Divide:     divide,
			BarcodeRef: barcodeRef,
		})
	}

	// Build all_unit_names พร้อมอัตราส่วน
	for i := range items {
		details, ok := detailMap[items[i].ItemCode]
		if !ok || len(details) <= 1 {
			items[i].UnitCount = 1
			if ok && len(details) == 1 {
				items[i].AllUnitNames = details[0].UnitName
			}
			continue
		}

		items[i].UnitCount = len(details)

		// หา base unit (barcoderef ว่าง = หน่วยหลัก)
		baseUnit := ""
		for _, d := range details {
			if d.BarcodeRef == "" {
				baseUnit = d.UnitName
				break
			}
		}

		// สร้าง string เช่น "ชิ้น, กล่อง(40), โหล(12)"
		parts := make([]string, 0, len(details))
		for _, d := range details {
			if d.BarcodeRef == "" {
				// หน่วยหลัก
				parts = append(parts, d.UnitName)
			} else {
				// หน่วยย่อย/ใหญ่กว่า — แสดงอัตราส่วน
				ratio := d.Stand
				if d.Divide > 0 {
					ratio = d.Stand / d.Divide
				}
				if ratio > 0 && ratio != 1 {
					ratioStr := formatRatio(ratio)
					if baseUnit != "" {
						parts = append(parts, fmt.Sprintf("%s(%s %s)", d.UnitName, ratioStr, baseUnit))
					} else {
						parts = append(parts, fmt.Sprintf("%s(%s)", d.UnitName, ratioStr))
					}
				} else {
					parts = append(parts, d.UnitName)
				}
			}
		}
		items[i].AllUnitNames = strings.Join(parts, ", ")
	}

	return items
}

// formatRatio — แสดงตัวเลขอัตราส่วน ถ้าเป็นจำนวนเต็มไม่แสดงทศนิยม
func formatRatio(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}

// ==================== Format Response ====================

func formatBarcodeItems(items []BarcodeListItem) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]interface{}{
			"guidfixed":          item.GuidFixed,
			"barcode":            item.Barcode,
			"names":              []map[string]string{{"code": "th", "name": item.Name}},
			"itemunitcode":       item.UnitCode,
			"itemunitnames":      []map[string]string{{"code": "th", "name": item.UnitName}},
			"itemcode":           item.ItemCode,
			"groupcode":          item.GroupCode,
			"groupnames":         []map[string]string{{"code": "th", "name": item.GroupNames}},
			"prices":             []map[string]interface{}{{"keynumber": 1, "price": item.Price1}},
			"imageuri":           item.ImageUri,
			"standvalue":        item.StandValue,
			"dividevalue":       item.DivideValue,
			"unit_count":         item.UnitCount,
			"all_unit_names":     item.AllUnitNames,
			"balance_qty":        item.BalanceQty,
			"balance_formatted":  item.BalanceFormatted,
		})
	}
	return result
}

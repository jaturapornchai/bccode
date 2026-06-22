package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ==================== Barcode List (MongoDB Atlas) ====================

// BarcodeListRequest — Request body สำหรับดึงรายการบาร์โค้ด
type BarcodeListRequest struct {
	HoldingCode        string   `json:"holdingcode"`
	Keyword            string   `json:"keyword"`
	GroupCode          string   `json:"groupcode"`
	GroupCodeLegacy    string   `json:"groupcode"`
	BrandCode          string   `json:"brandcode"`
	BrandCodeLegacy    string   `json:"brandcode"`
	CategoryCode       string   `json:"categorycode"`
	ClassCode          string   `json:"classcode"`
	DesignCode         string   `json:"designcode"`
	GradeCode          string   `json:"gradecode"`
	ModelCode          string   `json:"modelcode"`
	PatternCode        string   `json:"patterncode"`
	ItemType           *int     `json:"itemtype"`
	ItemTypeLegacy     *int     `json:"itemtype"`
	MaterialType       *int     `json:"materialtype"`
	MaterialTypeLegacy *int     `json:"materialtype"`
	PriceMin           *float64 `json:"pricemin"`
	PriceMax           *float64 `json:"pricemax"`
	Limit              int      `json:"limit"`
	Offset             int      `json:"offset"`
	SortField          string   `json:"sortfield"`
	SortOrder          string   `json:"sortorder"`
}

func mapString(doc bson.M, key string) string {
	if val, ok := doc[key]; ok && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func mapFloat(doc bson.M, key string) float64 {
	if val, ok := doc[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0.0
}

func mapInt(doc bson.M, key string) int {
	if val, ok := doc[key]; ok && val != nil {
		switch v := val.(type) {
		case int64:
			return int(v)
		case int32:
			return int(v)
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

func mapIntAny(doc bson.M, keys ...string) int {
	for _, key := range keys {
		if _, ok := doc[key]; ok {
			return mapInt(doc, key)
		}
	}
	return 0
}

func mapStringAny(doc bson.M, keys ...string) string {
	for _, key := range keys {
		if val := mapString(doc, key); val != "" {
			return val
		}
	}
	return ""
}

func mapBool(doc bson.M, key string) bool {
	if val, ok := doc[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func mapNames(doc bson.M, key string) []map[string]string {
	result := []map[string]string{}
	if val, ok := doc[key]; ok && val != nil {
		if arr, ok := val.(primitive.A); ok {
			for _, item := range arr {
				if m, ok := item.(bson.M); ok {
					code, _ := m["code"].(string)
					name, _ := m["name"].(string)
					if code != "" {
						result = append(result, map[string]string{
							"code": code,
							"name": name,
						})
					}
				}
			}
		}
	}
	return result
}

func firstMappedName(names []map[string]string) string {
	for _, name := range names {
		if val := strings.TrimSpace(name["name"]); val != "" {
			return val
		}
	}
	return ""
}

func bsonMap(value interface{}) (bson.M, bool) {
	switch v := value.(type) {
	case bson.M:
		return v, true
	case map[string]interface{}:
		return bson.M(v), true
	default:
		return nil, false
	}
}

func barcodeStockDimensions(doc bson.M) []map[string]string {
	raw, ok := doc["dimensions"]
	if !ok || raw == nil {
		return nil
	}
	arr, ok := raw.(primitive.A)
	if !ok {
		return nil
	}
	dimensions := make([]map[string]string, 0, len(arr))
	for _, entry := range arr {
		dim, ok := bsonMap(entry)
		if !ok {
			continue
		}
		item := bson.M{}
		if rawItem, ok := dim["item"]; ok {
			if itemMap, ok := bsonMap(rawItem); ok {
				item = itemMap
			}
		}
		dimensionGuid := mapStringAny(dim, "guidfixed", "guidfixed", "guid", "dimensionguid", "dimension_code", "code")
		itemGuid := mapStringAny(item, "guidfixed", "guidfixed", "guid", "itemguid", "option_guid", "valuecode", "code")
		dimensionName := firstMappedName(mapNames(dim, "names"))
		itemName := firstMappedName(mapNames(item, "names"))
		if dimensionGuid == "" && itemGuid == "" && dimensionName == "" && itemName == "" {
			continue
		}
		dimensions = append(dimensions, map[string]string{
			"dimensionguid": dimensionGuid,
			"dimensionname": dimensionName,
			"itemguid":      itemGuid,
			"itemname":      itemName,
		})
	}
	return dimensions
}

func stockDimensionPart(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func stockDimensionKey(dimensions []map[string]string) string {
	if len(dimensions) == 0 {
		return ""
	}
	parts := make([]string, 0, len(dimensions))
	for _, dim := range dimensions {
		left := stockDimensionPart(dim["dimensionguid"])
		if left == "" {
			left = stockDimensionPart(dim["dimensionname"])
		}
		right := stockDimensionPart(dim["itemguid"])
		if right == "" {
			right = stockDimensionPart(dim["itemname"])
		}
		if left == "" && right == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", left, right))
	}
	return strings.Join(parts, "|")
}

type stockBalanceInfo struct {
	currentQty   float64
	reservedQty  float64
	availableQty float64
	totalValue   float64
}

func stockBalanceMapKey(itemCode, dimensionKey string) string {
	return itemCode + "\x00" + dimensionKey
}

// pgSchemaCap caches information_schema capability probes (does table/column
// exist) per holding+table+column. The PG schema only changes on migration —
// which redeploys (restarts) the backend — so a process-lifetime cache is safe
// and removes a catalog round-trip from every barcode-list request.
var (
	pgSchemaCapMu sync.RWMutex
	pgSchemaCap   = map[string]bool{}
)

func pgSchemaCapGet(key string) (bool, bool) {
	pgSchemaCapMu.RLock()
	v, ok := pgSchemaCap[key]
	pgSchemaCapMu.RUnlock()
	return v, ok
}

func pgSchemaCapSet(key string, val bool) {
	pgSchemaCapMu.Lock()
	pgSchemaCap[key] = val
	pgSchemaCapMu.Unlock()
}

func pgColumnExistsProbe(ctx context.Context, db *sql.DB, tableName, columnName string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)`,
		tableName, columnName,
	).Scan(&exists)
	return exists, err
}

func pgTableExistsProbe(ctx context.Context, db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`,
		tableName,
	).Scan(&exists)
	return exists, err
}

// pgColumnExistsCached returns whether a column exists, caching the result.
// A transient probe error is not cached (returns false for this request only).
func pgColumnExistsCached(ctx context.Context, db *sql.DB, holdingCode, tableName, columnName string) bool {
	key := "c|" + holdingCode + "|" + tableName + "|" + columnName
	if v, ok := pgSchemaCapGet(key); ok {
		return v
	}
	exists, err := pgColumnExistsProbe(ctx, db, tableName, columnName)
	if err != nil {
		return false
	}
	pgSchemaCapSet(key, exists)
	return exists
}

// pgTableExistsCached returns whether a table exists, caching the result.
func pgTableExistsCached(ctx context.Context, db *sql.DB, holdingCode, tableName string) bool {
	key := "t|" + holdingCode + "|" + tableName
	if v, ok := pgSchemaCapGet(key); ok {
		return v
	}
	exists, err := pgTableExistsProbe(ctx, db, tableName)
	if err != nil {
		return false
	}
	pgSchemaCapSet(key, exists)
	return exists
}

// BarcodeListHandler — Handler สำหรับดึงรายการบาร์โค้ดจาก MongoDB Atlas
// POST /api/product/barcode/list
func BarcodeListHandler(c echo.Context) error {
	var req BarcodeListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: holdingcode",
		})
	}
	if req.GroupCode == "" {
		req.GroupCode = req.GroupCodeLegacy
	}
	if req.BrandCode == "" {
		req.BrandCode = req.BrandCodeLegacy
	}
	if req.ItemType == nil {
		req.ItemType = req.ItemTypeLegacy
	}
	if req.MaterialType == nil {
		req.MaterialType = req.MaterialTypeLegacy
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

	// Connect to MongoDB Atlas
	_, atlasDB := GetAtlasConnection()
	if atlasDB == nil {
		logger.Error("BarcodeListHandler: MongoDB connection is nil")
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"message": "MongoDB is not connected",
		})
	}

	collection := atlasDB.Collection("productbarcodes")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build MongoDB filter
	filter := bson.M{"holdingcode": req.HoldingCode}

	if req.Keyword != "" {
		keyword := strings.TrimSpace(req.Keyword)
		words := strings.Fields(keyword)
		if len(words) > 1 {
			andConds := []bson.M{}
			for _, w := range words {
				andConds = append(andConds, bson.M{
					"$or": []bson.M{
						{"barcode": bson.M{"$regex": w, "$options": "i"}},
						{"itemcode": bson.M{"$regex": w, "$options": "i"}},
						{"names.name": bson.M{"$regex": w, "$options": "i"}},
					},
				})
			}
			filter["$and"] = andConds
		} else {
			filter["$or"] = []bson.M{
				{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"itemcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
			}
		}
	}

	// Apply other filters
	if req.GroupCode != "" {
		filter["groupcode"] = req.GroupCode
	}
	if req.BrandCode != "" {
		filter["brandcode"] = req.BrandCode
	}
	if req.CategoryCode != "" {
		filter["categorycode"] = req.CategoryCode
	}
	if req.ClassCode != "" {
		filter["classcode"] = req.ClassCode
	}
	if req.DesignCode != "" {
		filter["designcode"] = req.DesignCode
	}
	if req.GradeCode != "" {
		filter["gradecode"] = req.GradeCode
	}
	if req.ModelCode != "" {
		filter["modelcode"] = req.ModelCode
	}
	if req.PatternCode != "" {
		filter["patterncode"] = req.PatternCode
	}
	if req.ItemType != nil {
		filter["itemtype"] = *req.ItemType
	}
	if req.MaterialType != nil {
		filter["materialtype"] = *req.MaterialType
	}
	if req.PriceMin != nil {
		filter["prices.price"] = bson.M{"$gte": *req.PriceMin}
	}
	if req.PriceMax != nil {
		if minFilter, ok := filter["prices.price"].(bson.M); ok {
			minFilter["$lte"] = *req.PriceMax
		} else {
			filter["prices.price"] = bson.M{"$lte": *req.PriceMax}
		}
	}

	// Count total documents
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("BarcodeListHandler: count: %v", err)
		totalCount = 0
	}

	// Find options
	findOpts := options.Find()
	findOpts.SetLimit(int64(req.Limit))
	findOpts.SetSkip(int64(req.Offset))

	// Sort mapping
	sortField := "barcode"
	if req.SortField != "" {
		sf := strings.ToLower(req.SortField)
		if strings.Contains(sf, "barcode") {
			sortField = "barcode"
		} else if strings.Contains(sf, "name0") {
			sortField = "names.0.name"
		} else if strings.Contains(sf, "itemcode") {
			sortField = "itemcode"
		} else if strings.Contains(sf, "price1") {
			sortField = "prices.0.price"
		}
	}
	sortOrderVal := 1
	if strings.ToLower(req.SortOrder) == "desc" {
		sortOrderVal = -1
	}
	findOpts.SetSort(bson.M{sortField: sortOrderVal})

	// Execute find query
	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		logger.Error("BarcodeListHandler: find: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Query failed",
		})
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err := cursor.All(ctx, &docs); err != nil {
		logger.Error("BarcodeListHandler: decode: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Query failed",
		})
	}

	// Collect itemcodes and prepare initial response format
	itemCodes := []string{}
	itemCodeSet := map[string]bool{}
	resultList := []map[string]interface{}{}

	for _, doc := range docs {
		itemCode := mapString(doc, "itemcode")
		if itemCode != "" && !itemCodeSet[itemCode] {
			itemCodeSet[itemCode] = true
			itemCodes = append(itemCodes, itemCode)
		}

		// Prices parsing
		prices := []map[string]interface{}{}
		if val, ok := doc["prices"]; ok && val != nil {
			if arr, ok := val.(primitive.A); ok {
				for _, item := range arr {
					if m, ok := item.(bson.M); ok {
						keyNum := mapInt(m, "keynumber")
						if keyNum == 0 {
							keyNum = mapInt(m, "keynumber")
						}
						prices = append(prices, map[string]interface{}{
							"keynumber": keyNum,
							"price":     mapFloat(m, "price"),
						})
					}
				}
			}
		}
		if len(prices) == 0 {
			prices = append(prices, map[string]interface{}{"keynumber": 1, "price": 0.0})
		}

		itemType := mapIntAny(doc, "itemtype", "itemtype")
		materialType := mapIntAny(doc, "materialtype", "material_type")
		stockDimensions := barcodeStockDimensions(doc)
		dimensionKey := stockDimensionKey(stockDimensions)

		resultList = append(resultList, map[string]interface{}{
			"guidfixed":          mapString(doc, "guidfixed"),
			"barcode":            mapString(doc, "barcode"),
			"names":              mapNames(doc, "names"),
			"itemunitcode":       mapString(doc, "itemunitcode"),
			"itemunitnames":      mapNames(doc, "itemunitnames"),
			"itemcode":           itemCode,
			"barcoderef":         mapString(doc, "barcoderef"),
			"groupcode":          mapString(doc, "groupcode"),
			"groupnames":         mapNames(doc, "groupnames"),
			"brandcode":          mapString(doc, "brandcode"),
			"brandnames":         mapNames(doc, "brandnames"),
			"categorycode":       mapString(doc, "categorycode"),
			"category_code":      mapString(doc, "categorycode"),
			"categorynames":      mapNames(doc, "categorynames"),
			"classcode":          mapString(doc, "classcode"),
			"class_code":         mapString(doc, "classcode"),
			"classnames":         mapNames(doc, "classnames"),
			"designcode":         mapString(doc, "designcode"),
			"design_code":        mapString(doc, "designcode"),
			"designnames":        mapNames(doc, "designnames"),
			"gradecode":          mapString(doc, "gradecode"),
			"grade_code":         mapString(doc, "gradecode"),
			"gradenames":         mapNames(doc, "gradenames"),
			"modelcode":          mapString(doc, "modelcode"),
			"model_code":         mapString(doc, "modelcode"),
			"modelnames":         mapNames(doc, "modelnames"),
			"patterncode":        mapString(doc, "patterncode"),
			"pattern_code":       mapString(doc, "patterncode"),
			"patternnames":       mapNames(doc, "patternnames"),
			"groupsubonecode":    mapString(doc, "groupsubonecode"),
			"groupsubonenames":   mapNames(doc, "groupsubonenames"),
			"groupsubtwocode":    mapString(doc, "groupsubtwocode"),
			"groupsubtwonames":   mapNames(doc, "groupsubtwonames"),
			"prices":             prices,
			"imageuri":           mapString(doc, "imageuri"),
			"standvalue":         mapFloat(doc, "standvalue"),
			"dividevalue":        mapFloat(doc, "dividevalue"),
			"isstock":            mapInt(doc, "isstock"),
			"itemtype":           itemType,
			"materialtype":       materialType,
			"material_type":      materialType,
			"isusesubbarcodes":   mapBool(doc, "isusesubbarcodes"),
			"checksum":           mapString(doc, "checksum"),
			"holdingcode":        mapString(doc, "holdingcode"),
			"refbarcodes":        doc["refbarcodes"],
			"bom":                doc["bom"],
			"businesstypes":      doc["businesstypes"],
			"ignorebranches":     doc["ignorebranches"],
			"unit_count":         1,
			"all_unit_names":     "",
			"stock_dimensionkey": dimensionKey,
			"stockdimensions":    stockDimensions,
			"balanceqty":         0.0,
			"reservedqty":        0.0,
			"availableqty":       0.0,
			"balance_formatted":  "",
		})
	}

	// 1. Enrich Unit Info from MongoDB
	if len(itemCodes) > 0 {
		unitFilter := bson.M{
			"holdingcode": req.HoldingCode,
			"itemcode":    bson.M{"$in": itemCodes},
		}
		unitCursor, err := collection.Find(ctx, unitFilter)
		if err == nil {
			var unitDocs []bson.M
			if err := unitCursor.All(ctx, &unitDocs); err == nil {
				type unitInfo struct {
					name   string
					stand  float64
					divide float64
					isRef  bool
				}
				itemUnitsMap := map[string][]unitInfo{}
				for _, ud := range unitDocs {
					code := mapString(ud, "itemcode")
					if code == "" {
						continue
					}
					unitName := ""
					unitNames := mapNames(ud, "itemunitnames")
					if len(unitNames) > 0 {
						unitName = unitNames[0]["name"]
					}
					isRef := mapString(ud, "barcoderef") != ""

					itemUnitsMap[code] = append(itemUnitsMap[code], unitInfo{
						name:   unitName,
						stand:  mapFloat(ud, "standvalue"),
						divide: mapFloat(ud, "dividevalue"),
						isRef:  isRef,
					})
				}

				for _, res := range resultList {
					code, _ := res["itemcode"].(string)
					units, ok := itemUnitsMap[code]
					if !ok || len(units) <= 1 {
						res["unit_count"] = 1
						if ok && len(units) == 1 {
							res["all_unit_names"] = units[0].name
						}
						continue
					}
					res["unit_count"] = len(units)
					parts := []string{}
					baseUnit := ""
					for _, u := range units {
						if !u.isRef {
							baseUnit = u.name
							parts = append(parts, u.name)
							break
						}
					}
					for _, u := range units {
						if u.isRef {
							ratio := u.stand
							if u.divide > 0 {
								ratio = u.stand / u.divide
							}
							if ratio > 0 && ratio != 1 {
								ratioStr := ""
								if ratio == float64(int64(ratio)) {
									ratioStr = fmt.Sprintf("%d", int64(ratio))
								} else {
									ratioStr = fmt.Sprintf("%.2f", ratio)
								}
								if baseUnit != "" {
									parts = append(parts, fmt.Sprintf("%s(%s %s)", u.name, ratioStr, baseUnit))
								} else {
									parts = append(parts, fmt.Sprintf("%s(%s)", u.name, ratioStr))
								}
							} else {
								parts = append(parts, u.name)
							}
						}
					}
					res["all_unit_names"] = strings.Join(parts, ", ")
				}
			}
		}
	}

	// 2. Enrich accounting stock and marketplace dimension availability.
	if len(itemCodes) > 0 {
		accountingBalanceMap := map[string]stockBalanceInfo{}
		marketplaceBalanceMap := map[string]stockBalanceInfo{}

		db, err := mypg.PgSqlFastConnect(req.HoldingCode)
		if err == nil {
			placeholders := make([]string, len(itemCodes))
			args := make([]interface{}, 0, len(itemCodes)+1)
			args = append(args, req.HoldingCode)
			for i, code := range itemCodes {
				placeholders[i] = fmt.Sprintf("$%d", i+2)
				args = append(args, code)
			}

			hasReservedQty := pgColumnExistsCached(ctx, db, req.HoldingCode, "inventorystockbalances", "reservedqty")
			reservedExpr := "0::numeric"
			if hasReservedQty {
				reservedExpr = "COALESCE(SUM(reservedqty), 0)"
			}
			accountingQuery := fmt.Sprintf(
				`SELECT itemcode,
				        COALESCE(SUM(currentqty), 0),
				        %s,
				        COALESCE(SUM(currenttotalvalue), 0)
				   FROM inventorystockbalances
				  WHERE holdingcode = $1 AND itemcode IN (%s)
				  GROUP BY itemcode`,
				reservedExpr,
				strings.Join(placeholders, ","),
			)

			rows, err := db.QueryContext(ctx, accountingQuery, args...)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var code string
					var qty, reservedQty, totalValue float64
					if err := rows.Scan(&code, &qty, &reservedQty, &totalValue); err == nil {
						accountingBalanceMap[code] = stockBalanceInfo{
							currentQty:  qty,
							reservedQty: reservedQty,
							totalValue:  totalValue,
						}
					}
				}
			} else {
				logger.Warn("BarcodeListHandler: inventory stock balance query failed: %v", err)
			}

			hasMarketplaceStock := pgTableExistsCached(ctx, db, req.HoldingCode, "marketplacestockbalances") &&
				pgColumnExistsCached(ctx, db, req.HoldingCode, "marketplacestockbalances", "itemcode") &&
				pgColumnExistsCached(ctx, db, req.HoldingCode, "marketplacestockbalances", "dimensionkey") &&
				pgColumnExistsCached(ctx, db, req.HoldingCode, "marketplacestockbalances", "availableqty")
			if hasMarketplaceStock {
				marketplaceQuery := fmt.Sprintf(
					`SELECT itemcode,
					        COALESCE(dimensionkey, '') AS dimensionkey,
					        COALESCE(SUM(current_qty), 0),
					        COALESCE(SUM(reservedqty), 0),
					        COALESCE(SUM(availableqty), 0)
					   FROM marketplacestockbalances
					  WHERE holdingcode = $1 AND itemcode IN (%s)
					  GROUP BY itemcode, COALESCE(dimensionkey, '')`,
					strings.Join(placeholders, ","),
				)
				rows, err := db.QueryContext(ctx, marketplaceQuery, args...)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var code, dimensionKey string
						var qty, reservedQty, availableQty float64
						if err := rows.Scan(&code, &dimensionKey, &qty, &reservedQty, &availableQty); err == nil {
							marketplaceBalanceMap[stockBalanceMapKey(code, dimensionKey)] = stockBalanceInfo{
								currentQty:   qty,
								reservedQty:  reservedQty,
								availableQty: availableQty,
							}
						}
					}
				} else {
					logger.Warn("BarcodeListHandler: marketplace stock balance query failed: %v", err)
				}
			}
		} else {
			logger.Warn("BarcodeListHandler: PostgreSQL connection failed: %v", err)
		}

		for _, res := range resultList {
			code, _ := res["itemcode"].(string)
			dimensionKey, _ := res["stock_dimensionkey"].(string)
			if accountingBalance, ok := accountingBalanceMap[code]; ok {
				res["balanceqty"] = accountingBalance.currentQty
				res["product_balanceqty"] = accountingBalance.currentQty
				res["reservedqty"] = accountingBalance.reservedQty
				res["availableqty"] = accountingBalance.currentQty - accountingBalance.reservedQty
				res["balanceamount"] = accountingBalance.totalValue
				res["balance_formatted"] = fmt.Sprintf("%.4f", accountingBalance.currentQty-accountingBalance.reservedQty)
			}
			if dimensionKey == "" {
				continue
			}
			if marketplaceBalance, ok := marketplaceBalanceMap[stockBalanceMapKey(code, dimensionKey)]; ok {
				res["reservedqty"] = marketplaceBalance.reservedQty
				res["availableqty"] = marketplaceBalance.availableQty
				res["balance_formatted"] = fmt.Sprintf("%.4f", marketplaceBalance.availableQty)
			} else {
				res["availableqty"] = 0.0
				res["balance_formatted"] = "0.0000"
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resultList,
		"total":   totalCount,
	})
}

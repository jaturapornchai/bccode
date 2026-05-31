package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	ShopID             string   `json:"shopid"`
	Keyword            string   `json:"keyword"`
	GroupCode          string   `json:"group_code"`
	GroupCodeLegacy    string   `json:"groupcode"`
	BrandCode          string   `json:"brand_code"`
	BrandCodeLegacy    string   `json:"brandcode"`
	CategoryCode       string   `json:"categorycode"`
	ClassCode          string   `json:"classcode"`
	DesignCode         string   `json:"designcode"`
	GradeCode          string   `json:"gradecode"`
	ModelCode          string   `json:"modelcode"`
	PatternCode        string   `json:"patterncode"`
	ItemType           *int     `json:"item_type"`
	ItemTypeLegacy     *int     `json:"itemtype"`
	MaterialType       *int     `json:"materialtype"`
	MaterialTypeLegacy *int     `json:"material_type"`
	PriceMin           *float64 `json:"price_min"`
	PriceMax           *float64 `json:"price_max"`
	Limit              int      `json:"limit"`
	Offset             int      `json:"offset"`
	SortField          string   `json:"sort_field"`
	SortOrder          string   `json:"sort_order"`
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

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: shopid",
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

	collection := atlasDB.Collection("productBarcodes")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build MongoDB filter
	filter := bson.M{"shopid": req.ShopID}

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
		filter["group_code"] = req.GroupCode
	}
	if req.BrandCode != "" {
		filter["brand_code"] = req.BrandCode
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
		filter["item_type"] = *req.ItemType
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
						keyNum := mapInt(m, "key_number")
						if keyNum == 0 {
							keyNum = mapInt(m, "keynumber")
						}
						prices = append(prices, map[string]interface{}{
							"key_number": keyNum,
							"price":      mapFloat(m, "price"),
						})
					}
				}
			}
		}
		if len(prices) == 0 {
			prices = append(prices, map[string]interface{}{"key_number": 1, "price": 0.0})
		}

		itemType := mapIntAny(doc, "item_type", "itemtype")
		materialType := mapIntAny(doc, "materialtype", "material_type")

		resultList = append(resultList, map[string]interface{}{
			"guid_fixed":        mapString(doc, "guidfixed"),
			"barcode":           mapString(doc, "barcode"),
			"names":             mapNames(doc, "names"),
			"item_unit_code":    mapString(doc, "itemunitcode"),
			"itemunitnames":     mapNames(doc, "itemunitnames"),
			"itemcode":          itemCode,
			"barcoderef":        mapString(doc, "barcoderef"),
			"group_code":        mapString(doc, "group_code"),
			"groupcode":         mapString(doc, "group_code"),
			"group_names":       mapNames(doc, "group_names"),
			"groupnames":        "",
			"brand_code":        mapString(doc, "brand_code"),
			"brandcode":         mapString(doc, "brand_code"),
			"brandnames":        mapNames(doc, "brandnames"),
			"categorycode":      mapString(doc, "categorycode"),
			"category_code":     mapString(doc, "categorycode"),
			"category_names":    mapNames(doc, "category_names"),
			"categorynames":     mapNames(doc, "category_names"),
			"classcode":         mapString(doc, "classcode"),
			"class_code":        mapString(doc, "classcode"),
			"classnames":        mapNames(doc, "classnames"),
			"designcode":        mapString(doc, "designcode"),
			"design_code":       mapString(doc, "designcode"),
			"designnames":       mapNames(doc, "designnames"),
			"gradecode":         mapString(doc, "gradecode"),
			"grade_code":        mapString(doc, "gradecode"),
			"gradenames":        mapNames(doc, "gradenames"),
			"modelcode":         mapString(doc, "modelcode"),
			"model_code":        mapString(doc, "modelcode"),
			"modelnames":        mapNames(doc, "modelnames"),
			"patterncode":       mapString(doc, "patterncode"),
			"pattern_code":      mapString(doc, "patterncode"),
			"patternnames":      mapNames(doc, "patternnames"),
			"groupsubonecode":   mapString(doc, "groupsubonecode"),
			"groupsubonenames":  mapNames(doc, "groupsubonenames"),
			"groupsubtwocode":   mapString(doc, "groupsubtwocode"),
			"groupsubtwonames":  mapNames(doc, "groupsubtwonames"),
			"prices":            prices,
			"imageuri":          mapString(doc, "imageuri"),
			"standvalue":        mapFloat(doc, "standvalue"),
			"dividevalue":       mapFloat(doc, "dividevalue"),
			"isstock":           mapInt(doc, "isstock"),
			"itemtype":          itemType,
			"item_type":         itemType,
			"materialtype":      materialType,
			"material_type":     materialType,
			"isusesubbarcodes":  mapBool(doc, "isusesubbarcodes"),
			"checksum":          mapString(doc, "checksum"),
			"shopid":            mapString(doc, "shopid"),
			"refbarcodes":       doc["refbarcodes"],
			"bom":               doc["bom"],
			"businesstypes":     doc["businesstypes"],
			"ignorebranches":    doc["ignorebranches"],
			"unit_count":        1,
			"all_unit_names":    "",
			"balance_qty":       0.0,
			"balance_formatted": "",
		})
	}

	// 1. Enrich Unit Info from MongoDB
	if len(itemCodes) > 0 {
		unitFilter := bson.M{
			"shopid":   req.ShopID,
			"itemcode": bson.M{"$in": itemCodes},
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

	// 2. Enrich Balance Qty from PostgreSQL
	if len(itemCodes) > 0 {
		balanceMap := map[string]float64{}
		balanceWordMap := map[string]string{}

		db, err := mypg.PgSqlFastConnect(req.ShopID)
		if err == nil {
			placeholders := make([]string, len(itemCodes))
			args := make([]interface{}, len(itemCodes))
			for i, code := range itemCodes {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args[i] = code
			}
			balQuery := fmt.Sprintf(
				"SELECT itemcode, COALESCE(balanceqty, 0), COALESCE(balanceqtyword, '') FROM product WHERE itemcode IN (%s)",
				strings.Join(placeholders, ","),
			)
			rows, err := db.Query(balQuery, args...)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var code, word string
					var qty float64
					if err := rows.Scan(&code, &qty, &word); err == nil {
						balanceMap[code] = qty
						balanceWordMap[code] = word
					}
				}
			}
		}

		for _, res := range resultList {
			code, _ := res["itemcode"].(string)
			if qty, ok := balanceMap[code]; ok {
				res["balance_qty"] = qty
			}
			if word, ok := balanceWordMap[code]; ok {
				res["balance_formatted"] = word
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resultList,
		"total":   totalCount,
	})
}

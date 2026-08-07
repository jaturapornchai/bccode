package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"smlcloudplatform/internal/goapi/logger"
)

// ==================== Barcode List (MongoDB Atlas) ====================

// BarcodeListRequest — Request body สำหรับดึงรายการบาร์โค้ด
type BarcodeListRequest struct {
	HoldingCode string `json:"holdingcode"`
	Keyword     string `json:"keyword"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	SortField   string `json:"sortfield"`
	SortOrder   string `json:"sortorder"`
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

	holdingCode, companyCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, "")
	if scopeErr != nil {
		return c.JSON(scopeErr.Status, map[string]interface{}{
			"success": false,
			"code":    scopeErr.Code,
			"message": scopeErr.Message,
		})
	}
	req.HoldingCode = holdingCode
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

	// Build MongoDB filter. Exclude soft-deleted docs (deletedat set) — otherwise deleted barcodes
	// linger as ghosts and look like duplicates, since uniqueness only blocks LIVE barcodes.
	filter := bson.M{
		"holdingcode":  req.HoldingCode,
		"businesscode": companyCode,
		"deletedat":    bson.M{"$exists": false},
	}

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

	// Product classification belongs to Product. Barcode list only filters its own fields.
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

		resultList = append(resultList, map[string]interface{}{
			"guidfixed":      mapString(doc, "guidfixed"),
			"holdingcode":    req.HoldingCode,
			"businesscode":   companyCode,
			"barcode":        mapString(doc, "barcode"),
			"names":          mapNames(doc, "names"),
			"itemcode":       itemCode,
			"itemunitguid":   mapString(doc, "itemunitguid"),
			"itemunitcode":   mapString(doc, "itemunitcode"),
			"itemunitnames":  mapNames(doc, "itemunitnames"),
			"prices":         prices,
			"condition":      mapBool(doc, "condition"),
			"dividevalue":    mapFloat(doc, "dividevalue"),
			"standvalue":     mapFloat(doc, "standvalue"),
			"ismainbarcode":  mapBool(doc, "ismainbarcode"),
			"unit_count":     1,
			"all_unit_names": "",
		})
	}

	// 1. Enrich Unit Info from MongoDB
	if len(itemCodes) > 0 {
		unitFilter := bson.M{
			"holdingcode":  req.HoldingCode,
			"businesscode": companyCode,
			"itemcode":     bson.M{"$in": itemCodes},
			"deletedat":    bson.M{"$exists": false},
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resultList,
		"total":   totalCount,
	})
}

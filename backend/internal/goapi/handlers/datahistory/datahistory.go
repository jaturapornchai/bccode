package datahistory

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionName = "datahistory"
)

func databaseName() string {
	return config.NewServiceConfig().MongodbDatabaseName()
}

// ActionType - ประเภทการกระทำ
type ActionType string

const (
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
)

// ScreenType - ประเภทหน้าจอ
type ScreenType string

const (
	ScreenPurchaseOrder ScreenType = "purchaseorder"
	ScreenPurchase      ScreenType = "purchase"
	ScreenSale          ScreenType = "sale"
	ScreenStockAdjust   ScreenType = "stockadjust"
	ScreenTransfer      ScreenType = "transfer"
)

// FieldChange - รายละเอียดการเปลี่ยนแปลง field
type FieldChange struct {
	Field    string      `json:"field" bson:"field"`
	OldValue interface{} `json:"old_value" bson:"old_value"`
	NewValue interface{} `json:"new_value" bson:"new_value"`
}

// DataHistory - โครงสร้างข้อมูล history
type DataHistory struct {
	ID         primitive.ObjectID     `json:"_id,omitempty" bson:"_id,omitempty"`
	ShopID     string                 `json:"shopid" bson:"shopid"`
	ScreenType ScreenType             `json:"screen_type" bson:"screen_type"`
	Action     ActionType             `json:"action" bson:"action"`
	DocNo      string                 `json:"docno" bson:"docno"`
	GuidFixed  string                 `json:"guid_fixed" bson:"guid_fixed"`
	UserCode   string                 `json:"user_code" bson:"user_code"`
	UserName   string                 `json:"user_name" bson:"user_name"`
	Timestamp  time.Time              `json:"timestamp" bson:"timestamp"`
	DataBefore map[string]interface{} `json:"data_before" bson:"data_before"` // null ถ้า create
	DataAfter  map[string]interface{} `json:"data_after" bson:"data_after"`   // null ถ้า delete
	Changes    []FieldChange          `json:"changes" bson:"changes"`         // สรุปสิ่งที่เปลี่ยน
}

// getCollection - ดึง collection สำหรับ datahistory
func getCollection() (*mongo.Collection, error) {
	client, err := myglobal.MongoConnect()
	if err != nil {
		return nil, err
	}
	return client.Database(databaseName()).Collection(CollectionName), nil
}

// SaveHistory - บันทึก history ลง MongoDB
func SaveHistory(history DataHistory) error {
	collection, err := getCollection()
	if err != nil {
		logger.Error("[DataHistory] Failed to get collection: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ตั้งค่า timestamp ถ้ายังไม่มี
	if history.Timestamp.IsZero() {
		history.Timestamp = time.Now().UTC()
	}

	_, err = collection.InsertOne(ctx, history)
	if err != nil {
		logger.Error("[DataHistory] Failed to save history: %v", err)
		return err
	}

	logger.Debug("[DataHistory] Saved: shopid=%s, screen=%s, action=%s, docno=%s",
		history.ShopID, history.ScreenType, history.Action, history.DocNo)
	return nil
}

// HasPOHistory - ตรวจสอบว่า PO มี history อยู่แล้วหรือไม่ (ใช้แยก create กับ update)
func HasPOHistory(shopID, guidFixed string) bool {
	collection, err := getCollection()
	if err != nil {
		logger.Error("[DataHistory] Failed to get collection for check: %v", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ค้นหาจาก guidfixed เพราะ docno อาจเปลี่ยนได้
	filter := bson.M{
		"shopid":      shopID,
		"screen_type": ScreenPurchaseOrder,
		"guid_fixed":  guidFixed,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("[DataHistory] Failed to count history: %v", err)
		return false
	}

	return count > 0
}

// GetLastPOSnapshot - ดึง dataAfter ล่าสุดของ PO (ใช้เปรียบเทียบกับข้อมูลใหม่)
func GetLastPOSnapshot(shopID, guidFixed string) map[string]interface{} {
	collection, err := getCollection()
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"shopid":      shopID,
		"screen_type": ScreenPurchaseOrder,
		"guid_fixed":  guidFixed,
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var result DataHistory
	err = collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		return nil
	}

	return result.DataAfter
}

// poKeyFields - fields สำคัญที่ใช้เปรียบเทียบว่ามีการแก้ไขจริงหรือไม่
var poKeyFields = []string{
	"docno", "docdatetime", "custcode", "description", "discountword",
	"totaldiscount", "totalvalue", "total_amount", "totalvatvalue",
	"totalbeforevat", "totalaftervat", "vatrate", "vat_type",
	"iscancel", "cancelreason",
	"doc_currency", "exchange_rate", "totalamount_doc",
	"purchasetypecode",
}

// HasMeaningfulChanges - ตรวจสอบว่ามีการเปลี่ยนแปลง field สำคัญหรือไม่
// เปรียบเทียบ key fields + details array
func HasMeaningfulChanges(oldData, newData map[string]interface{}) bool {
	if oldData == nil || newData == nil {
		return true // ถ้าไม่มีข้อมูลเก่า ถือว่ามีการเปลี่ยนแปลง
	}

	// ตรวจสอบ key fields
	for _, field := range poKeyFields {
		if !compareValues(oldData[field], newData[field]) {
			logger.Debug("[DataHistory] Field changed: %s", field)
			return true
		}
	}

	// ตรวจสอบ details (จำนวนรายการ)
	oldDetails, _ := oldData["details"].([]interface{})
	newDetails, _ := newData["details"].([]interface{})
	if len(oldDetails) != len(newDetails) {
		logger.Debug("[DataHistory] Details count changed: %d -> %d", len(oldDetails), len(newDetails))
		return true
	}

	// ตรวจสอบแต่ละรายการ
	for i := range oldDetails {
		oldItem, _ := oldDetails[i].(map[string]interface{})
		newItem, _ := newDetails[i].(map[string]interface{})
		if oldItem == nil || newItem == nil {
			return true
		}
		// เปรียบเทียบ fields สำคัญของรายการสินค้า
		detailFields := []string{"itemcode", "qty", "price", "sum_amount", "unitcode", "whcode", "locationcode", "discount", "discountamount"}
		for _, f := range detailFields {
			if !compareValues(oldItem[f], newItem[f]) {
				logger.Debug("[DataHistory] Detail[%d].%s changed", i, f)
				return true
			}
		}
	}

	return false
}

// SavePOHistory - บันทึก history สำหรับ Purchase Order
func SavePOHistory(shopID, docNo, guidFixed, userCode, userName string, action ActionType, dataBefore, dataAfter map[string]interface{}) error {
	history := DataHistory{
		ShopID:     shopID,
		ScreenType: ScreenPurchaseOrder,
		Action:     action,
		DocNo:      docNo,
		GuidFixed:  guidFixed,
		UserCode:   userCode,
		UserName:   userName,
		Timestamp:  time.Now().UTC(),
		DataBefore: dataBefore,
		DataAfter:  dataAfter,
		Changes:    calculateChanges(dataBefore, dataAfter),
	}
	return SaveHistory(history)
}

// calculateChanges - คำนวณความแตกต่างระหว่าง before และ after
func calculateChanges(before, after map[string]interface{}) []FieldChange {
	var changes []FieldChange

	if before == nil && after == nil {
		return changes
	}

	// กรณี create - ไม่มี before
	if before == nil {
		for key, value := range after {
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: nil,
				NewValue: value,
			})
		}
		return changes
	}

	// กรณี delete - ไม่มี after
	if after == nil {
		for key, value := range before {
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: value,
				NewValue: nil,
			})
		}
		return changes
	}

	// กรณี update - เปรียบเทียบ before และ after
	allKeys := make(map[string]bool)
	for key := range before {
		allKeys[key] = true
	}
	for key := range after {
		allKeys[key] = true
	}

	for key := range allKeys {
		oldVal := before[key]
		newVal := after[key]

		// เปรียบเทียบค่า (simple comparison)
		if !compareValues(oldVal, newVal) {
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	return changes
}

// compareValues - เปรียบเทียบค่า 2 ค่า (รองรับ map, slice, และ nested objects)
func compareValues(a, b interface{}) bool {
	defer func() {
		if r := recover(); r != nil {
			// กรณี panic จากการเปรียบเทียบ type ที่เปรียบเทียบไม่ได้
			logger.Debug("[DataHistory] compareValues recover จาก panic: %v", r)
		}
	}()

	// ถ้าทั้งคู่เป็น nil ถือว่าเท่ากัน
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// ใช้ reflect.DeepEqual สำหรับ map, slice, และ type อื่นๆ ที่เปรียบเทียบด้วย == ไม่ได้
	aKind := reflect.TypeOf(a).Kind()
	if aKind == reflect.Map || aKind == reflect.Slice {
		return reflect.DeepEqual(a, b)
	}

	// สำหรับ type ทั่วไป ใช้ string comparison เพื่อความปลอดภัย
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// GetHistoryHandler - API endpoint สำหรับดึง history
func GetHistoryHandler(c echo.Context) error {
	shopID := c.QueryParam("shopid")
	screenType := c.QueryParam("screen_type")
	docNo := c.QueryParam("docno")

	if shopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "shopid is required",
		})
	}

	collection, err := getCollection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to connect to database",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// สร้าง filter
	filter := bson.M{"shopid": shopID}
	if screenType != "" {
		filter["screen_type"] = screenType
	}
	if docNo != "" {
		filter["docno"] = docNo
	}

	// เรียงลำดับตาม timestamp ล่าสุดก่อน
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	// จำกัดจำนวน
	limit := int64(100)
	opts.SetLimit(limit)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("[DataHistory] Failed to query: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to query history",
		})
	}
	defer cursor.Close(ctx)

	var results []DataHistory
	if err := cursor.All(ctx, &results); err != nil {
		logger.Error("[DataHistory] Failed to decode: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to decode history",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// GetPOHistoryHandler - API endpoint สำหรับดึง history ของใบสั่งซื้อ
func GetPOHistoryHandler(c echo.Context) error {
	shopID := c.QueryParam("shopid")
	docNo := c.QueryParam("docno")

	if shopID == "" || docNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "shopid and docno are required",
		})
	}

	collection, err := getCollection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to connect to database",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter สำหรับ PO เฉพาะ docno
	filter := bson.M{
		"shopid":      shopID,
		"screen_type": ScreenPurchaseOrder,
		"docno":       docNo,
	}

	// เรียงลำดับตาม timestamp ล่าสุดก่อน
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("[DataHistory] Failed to query PO history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to query history",
		})
	}
	defer cursor.Close(ctx)

	var results []DataHistory
	if err := cursor.All(ctx, &results); err != nil {
		logger.Error("[DataHistory] Failed to decode PO history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to decode history",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

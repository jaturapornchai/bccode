package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const unitCollection = "units"

// ==================== Unit Types ====================

// UnitNameEntry ชื่อหน่วยนับแต่ละภาษา
type UnitNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// UnitDocument เอกสารหน่วยนับใน MongoDB
type UnitDocument struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holding_code" bson:"holding_code"`
	GuidFixed   string             `json:"guid_fixed" bson:"guid_fixed"`
	UnitCode    string             `json:"unitcode" bson:"unitcode"`
	Names       []UnitNameEntry    `json:"names" bson:"names"`
	CreatedBy   string             `json:"createdby" bson:"createdby"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedBy   string             `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt   time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// ==================== List Units ====================

// ListUnitsResponse ผลลัพธ์จากการดึง/ค้นหาหน่วยนับ
type ListUnitsResponse struct {
	Units       []UnitDocument `json:"units"`
	Count       int            `json:"count"`
	Keyword     string         `json:"keyword,omitempty"`
	GeneratedAt time.Time      `json:"generated_at"`
}

// ListUnits ดึง/ค้นหาหน่วยนับ
func ListUnits(ctx context.Context, holdingCode, keyword string, limit int) (*ListUnitsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()

	// สร้าง filter
	filter := bson.M{
		"holding_code": holdingCode,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	// ถ้ามี keyword → ค้นหาใน unitcode, names.name
	if keyword != "" {
		keywordFilter := bson.M{
			"$or": []bson.M{
				{"unitcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListUnits] holdingCode=%s, keyword=%s, limit=%d", holdingCode, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(unitCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "unitcode", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var units []UnitDocument
	if err := cursor.All(ctx, &units); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}

	if units == nil {
		units = []UnitDocument{}
	}

	return &ListUnitsResponse{
		Units:       units,
		Count:       len(units),
		Keyword:     keyword,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Unit ====================

// CreateUnitResponse ผลลัพธ์จากการสร้างหน่วยนับ
type CreateUnitResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message"`
	Unit        UnitDocument `json:"unit"`
	KafkaSync   string       `json:"kafka_sync"`
	KafkaError  string       `json:"kafka_error,omitempty"`
	GeneratedAt time.Time    `json:"generated_at"`
}

// CreateUnit สร้างหน่วยนับใหม่
// ใช้ names[] เป็นชื่อแสดงผล
func CreateUnit(ctx context.Context, holdingCode, unitCode, namesJSON string) (*CreateUnitResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if unitCode == "" {
		return nil, fmt.Errorf("unitcode is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(unitCollection)

	// ตรวจสอบ unitcode ซ้ำ
	existFilter := bson.M{
		"holding_code": holdingCode,
		"unitcode":     unitCode,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}
	count, err := coll.CountDocuments(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ unitcode ซ้ำล้มเหลว: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("unitcode '%s' มีอยู่แล้วใน shop นี้", unitCode)
	}

	// Parse names JSON
	var names []UnitNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required — ต้องระบุชื่อหน่วยนับอย่างน้อย 1 ภาษา เช่น [{\"code\":\"th\",\"name\":\"ชิ้น\"}]")
	}

	now := time.Now()
	doc := UnitDocument{
		HoldingCode: holdingCode,
		GuidFixed:   uuid.New().String(),
		UnitCode:    unitCode,
		Names:       names,
		CreatedBy:   "mcp-tool",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้างหน่วยนับล้มเหลว: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	// Publish Kafka event (เหมือน mainapi)
	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicUnitCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateUnit] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreateUnit] สร้าง unitcode=%s สำเร็จ (shop=%s, kafka=%s)", unitCode, holdingCode, kafkaSync)

	return &CreateUnitResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้างหน่วยนับ '%s' สำเร็จ", unitCode),
		Unit:        doc,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Units (Bulk) ====================

// CreateUnitsRequest รายการหน่วยนับที่ต้องการสร้างพร้อมกัน
type CreateUnitsItem struct {
	UnitCode string          `json:"unitcode"`
	Names    []UnitNameEntry `json:"names"`
}

// CreateUnitsResponse ผลลัพธ์จากการสร้างหน่วยนับหลายรายการ
type CreateUnitsResponse struct {
	Success      bool           `json:"success"`
	Message      string         `json:"message"`
	Created      []UnitDocument `json:"created"`
	Skipped      []string       `json:"skipped,omitempty"`
	CreatedCount int            `json:"created_count"`
	SkippedCount int            `json:"skipped_count"`
	KafkaSync    string         `json:"kafka_sync"`
	KafkaError   string         `json:"kafka_error,omitempty"`
	GeneratedAt  time.Time      `json:"generated_at"`
}

// CreateUnits สร้างหน่วยนับหลายรายการพร้อมกัน
func CreateUnits(ctx context.Context, holdingCode, unitsJSON string) (*CreateUnitsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if unitsJSON == "" {
		return nil, fmt.Errorf("units JSON is required")
	}

	var items []CreateUnitsItem
	if err := json.Unmarshal([]byte(unitsJSON), &items); err != nil {
		return nil, fmt.Errorf("units JSON ไม่ถูกต้อง: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ต้องระบุหน่วยนับอย่างน้อย 1 รายการ")
	}
	if len(items) > 100 {
		return nil, fmt.Errorf("สร้างได้สูงสุด 100 รายการต่อครั้ง")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(unitCollection)

	// ดึง unitcode ที่มีอยู่แล้ว
	var unitCodes []string
	for _, item := range items {
		unitCodes = append(unitCodes, item.UnitCode)
	}

	existFilter := bson.M{
		"holding_code": holdingCode,
		"unitcode":     bson.M{"$in": unitCodes},
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}
	cursor, err := coll.Find(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ unitcode ซ้ำล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingCodes := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			UnitCode string `bson:"unitcode"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingCodes[doc.UnitCode] = true
		}
	}

	now := time.Now()
	var docsToInsert []interface{}
	var createdDocs []UnitDocument
	var skipped []string

	for _, item := range items {
		if item.UnitCode == "" {
			skipped = append(skipped, "(unitcode ว่าง)")
			continue
		}
		if existingCodes[item.UnitCode] {
			skipped = append(skipped, fmt.Sprintf("%s (มีอยู่แล้ว)", item.UnitCode))
			continue
		}

		names := item.Names
		if len(names) == 0 {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี names)", item.UnitCode))
			continue
		}

		doc := UnitDocument{
			HoldingCode: holdingCode,
			GuidFixed:   uuid.New().String(),
			UnitCode:    item.UnitCode,
			Names:       names,
			CreatedBy:   "mcp-tool",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		docsToInsert = append(docsToInsert, doc)
		createdDocs = append(createdDocs, doc)
	}

	kafkaSync := "skipped"
	kafkaError := ""

	if len(docsToInsert) > 0 {
		_, err = coll.InsertMany(ctx, docsToInsert)
		if err != nil {
			return nil, fmt.Errorf("สร้างหน่วยนับล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		// Publish Kafka event — bulk created (เหมือน mainapi)
		if err := publishToKafkaBatch(ctx, TopicUnitBulkCreated, docsToInsert); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP CreateUnits] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
		}
	}

	logger.Info("[MCP CreateUnits] สร้าง %d รายการสำเร็จ, ข้าม %d รายการ (shop=%s, kafka=%s)", len(createdDocs), len(skipped), holdingCode, kafkaSync)

	return &CreateUnitsResponse{
		Success:      true,
		Message:      fmt.Sprintf("สร้างหน่วยนับ %d รายการสำเร็จ", len(createdDocs)),
		Created:      createdDocs,
		Skipped:      skipped,
		CreatedCount: len(createdDocs),
		SkippedCount: len(skipped),
		KafkaSync:    kafkaSync,
		KafkaError:   kafkaError,
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== Update Unit ====================

// UpdateUnitResponse ผลลัพธ์จากการอัปเดตหน่วยนับ
type UpdateUnitResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message"`
	Unit        UnitDocument `json:"unit"`
	KafkaSync   string       `json:"kafka_sync"`
	KafkaError  string       `json:"kafka_error,omitempty"`
	GeneratedAt time.Time    `json:"generated_at"`
}

// UpdateUnit อัปเดตหน่วยนับตาม unitcode
func UpdateUnit(ctx context.Context, holdingCode, unitCode, namesJSON string) (*UpdateUnitResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if unitCode == "" {
		return nil, fmt.Errorf("unitcode is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(unitCollection)

	// ค้นหา unit ที่ต้องการอัปเดต
	filter := bson.M{
		"holding_code": holdingCode,
		"unitcode":     unitCode,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	var existing UnitDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบหน่วยนับ unitcode='%s'", unitCode)
	}

	// สร้าง update fields
	updateFields := bson.M{
		"updated_at": time.Now(),
		"updatedby":  "mcp-tool",
	}
	if namesJSON != "" {
		var names []UnitNameEntry
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["names"] = names
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดตหน่วยนับล้มเหลว: %w", err)
	}

	// ดึง document ที่อัปเดตแล้ว
	var updated UnitDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	// Publish Kafka event (เหมือน mainapi)
	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicUnitUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdateUnit] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP UpdateUnit] อัปเดต unitcode=%s สำเร็จ (shop=%s, kafka=%s)", unitCode, holdingCode, kafkaSync)

	return &UpdateUnitResponse{
		Success:     true,
		Message:     fmt.Sprintf("อัปเดตหน่วยนับ '%s' สำเร็จ", unitCode),
		Unit:        updated,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Unit ====================

// DeleteUnitResponse ผลลัพธ์จากการลบหน่วยนับ
type DeleteUnitResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	UnitCode    string    `json:"unitcode"`
	KafkaSync   string    `json:"kafka_sync"`
	KafkaError  string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

// DeleteUnit ลบหน่วยนับตาม unitcode
func DeleteUnit(ctx context.Context, holdingCode, unitCode string) (*DeleteUnitResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if unitCode == "" {
		return nil, fmt.Errorf("unitcode is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(unitCollection)

	filter := bson.M{
		"holding_code": holdingCode,
		"unitcode":     unitCode,
	}

	// ดึง document ก่อนลบ เพื่อส่ง Kafka event
	var docToDelete UnitDocument
	coll.FindOne(ctx, filter).Decode(&docToDelete)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ลบหน่วยนับล้มเหลว: %w", err)
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("ไม่พบหน่วยนับ unitcode='%s'", unitCode)
	}

	// Publish Kafka event (เหมือน mainapi)
	kafkaSync := "published"
	kafkaError := ""
	if docToDelete.UnitCode != "" {
		if err := publishToKafka(ctx, TopicUnitDeleted, docToDelete); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP DeleteUnit] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
		}
	} else {
		kafkaSync = "skipped"
	}

	logger.Info("[MCP DeleteUnit] ลบ unitcode=%s สำเร็จ (shop=%s, kafka=%s)", unitCode, holdingCode, kafkaSync)

	return &DeleteUnitResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบหน่วยนับ '%s' สำเร็จ", unitCode),
		UnitCode:    unitCode,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Units (Bulk) ====================

// DeleteUnitsResponse ผลลัพธ์จากการลบหน่วยนับหลายรายการ
type DeleteUnitsResponse struct {
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	Deleted       []string  `json:"deleted"`
	NotFound      []string  `json:"not_found,omitempty"`
	DeletedCount  int       `json:"deleted_count"`
	NotFoundCount int       `json:"not_found_count"`
	KafkaSync     string    `json:"kafka_sync"`
	KafkaError    string    `json:"kafka_error,omitempty"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// DeleteUnits ลบหน่วยนับหลายรายการพร้อมกัน
func DeleteUnits(ctx context.Context, holdingCode, unitcodesJSON string) (*DeleteUnitsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if unitcodesJSON == "" {
		return nil, fmt.Errorf("unitcodes is required")
	}

	var unitCodes []string
	if err := json.Unmarshal([]byte(unitcodesJSON), &unitCodes); err != nil {
		return nil, fmt.Errorf("unitcodes JSON ไม่ถูกต้อง (ต้องเป็น array of strings): %w", err)
	}
	if len(unitCodes) == 0 {
		return nil, fmt.Errorf("ต้องระบุ unitcode อย่างน้อย 1 รายการ")
	}
	if len(unitCodes) > 100 {
		return nil, fmt.Errorf("ลบได้สูงสุด 100 รายการต่อครั้ง")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(unitCollection)

	// ค้นหา unitcodes ที่มีอยู่จริง
	existFilter := bson.M{
		"holding_code": holdingCode,
		"unitcode":     bson.M{"$in": unitCodes},
	}
	cursor, err := coll.Find(ctx, existFilter, options.Find().SetProjection(bson.M{"unitcode": 1}))
	if err != nil {
		return nil, fmt.Errorf("ค้นหาหน่วยนับล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingCodes := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			UnitCode string `bson:"unitcode"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingCodes[doc.UnitCode] = true
		}
	}

	var deleted []string
	var notFound []string
	for _, code := range unitCodes {
		if existingCodes[code] {
			deleted = append(deleted, code)
		} else {
			notFound = append(notFound, code)
		}
	}

	// ดึง documents ก่อนลบ เพื่อส่ง Kafka event
	kafkaSync := "skipped"
	kafkaError := ""
	var docsToDelete []interface{}
	if len(deleted) > 0 {
		delCursor, _ := coll.Find(ctx, bson.M{
			"holding_code": holdingCode,
			"unitcode":     bson.M{"$in": deleted},
		})
		if delCursor != nil {
			var docs []UnitDocument
			delCursor.All(ctx, &docs)
			delCursor.Close(ctx)
			for _, d := range docs {
				docsToDelete = append(docsToDelete, d)
			}
		}
	}

	// ลบทั้งหมดที่มีอยู่ด้วย DeleteMany
	if len(deleted) > 0 {
		deleteFilter := bson.M{
			"holding_code": holdingCode,
			"unitcode":     bson.M{"$in": deleted},
		}
		_, err = coll.DeleteMany(ctx, deleteFilter)
		if err != nil {
			return nil, fmt.Errorf("ลบหน่วยนับล้มเหลว: %w", err)
		}

		// Publish Kafka event — bulk deleted (เหมือน mainapi)
		kafkaSync = "published"
		if len(docsToDelete) > 0 {
			if err := publishToKafkaBatch(ctx, TopicUnitBulkDeleted, docsToDelete); err != nil {
				kafkaSync = "failed"
				kafkaError = err.Error()
				logger.Warn("[MCP DeleteUnits] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
			}
		}
	}

	logger.Info("[MCP DeleteUnits] ลบ %d รายการสำเร็จ, ไม่พบ %d รายการ (shop=%s, kafka=%s)", len(deleted), len(notFound), holdingCode, kafkaSync)

	return &DeleteUnitsResponse{
		Success:       true,
		Message:       fmt.Sprintf("ลบหน่วยนับ %d รายการสำเร็จ", len(deleted)),
		Deleted:       deleted,
		NotFound:      notFound,
		DeletedCount:  len(deleted),
		NotFoundCount: len(notFound),
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Get Unit Schema ====================

// GetUnitSchema คืนโครงสร้างข้อมูลหน่วยนับ
func GetUnitSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  unitCollection,
		"description": "หน่วยนับ (Unit of Measure) — ใช้กำหนดหน่วยของสินค้า เช่น ชิ้น, กล่อง, กิโลกรัม",
		"fields": map[string]interface{}{
			"_id":          "ObjectID — MongoDB auto-generated ID",
			"holding_code": "string — Holding Code (tenant isolation)",
			"guid_fixed":   "string — UUID สำหรับอ้างอิงภายใน",
			"unitcode":     "string (required, max 100) — รหัสหน่วยนับ เช่น EA, BOX, KG, PACK",
			"names":        "array (required) — ชื่อหลายภาษา [{code:'th', name:'ชิ้น'}, {code:'en', name:'Each'}]",
			"createdby":    "string — ผู้สร้าง",
			"created_at":   "datetime — วันที่สร้าง",
			"updatedby":    "string — ผู้แก้ไขล่าสุด",
			"updated_at":   "datetime — วันที่แก้ไขล่าสุด",
			"deleted_at":   "datetime — วันที่ลบ (soft delete)",
		},
		"indexes": []string{
			"holding_code + unitcode (unique per shop)",
			"holding_code + guidfixed",
		},
		"examples": []map[string]interface{}{
			{
				"unitcode": "EA",
				"names":    []map[string]string{{"code": "th", "name": "ชิ้น"}, {"code": "en", "name": "Each"}},
			},
			{
				"unitcode": "BOX",
				"names":    []map[string]string{{"code": "th", "name": "กล่อง"}, {"code": "en", "name": "Box"}},
			},
			{
				"unitcode": "KG",
				"names":    []map[string]string{{"code": "th", "name": "กิโลกรัม"}, {"code": "en", "name": "Kilogram"}},
			},
		},
		"related_collections": []string{
			"productbarcodes — ใช้ itemunitcode อ้างอิง unitcode",
		},
	}
}

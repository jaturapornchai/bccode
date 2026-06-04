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

const productGroupCollection = "productGroups"

// ==================== Product Group Types ====================

// ProductGroupNameEntry ชื่อกลุ่มสินค้าแต่ละภาษา
type ProductGroupNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// ProductGroupDocument เอกสารกลุ่มสินค้าใน MongoDB
type ProductGroupDocument struct {
	ID          primitive.ObjectID      `json:"id" bson:"_id,omitempty"`
	HoldingCode string                  `json:"holding_code" bson:"holding_code"`
	GuidFixed   string                  `json:"guid_fixed" bson:"guid_fixed"`
	Code        string                  `json:"code" bson:"code"`
	Names       []ProductGroupNameEntry `json:"names" bson:"names"`
	CreatedBy   string                  `json:"createdby" bson:"createdby"`
	CreatedAt   time.Time               `json:"created_at" bson:"created_at"`
	UpdatedBy   string                  `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	UpdatedAt   time.Time               `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt   time.Time               `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// ==================== List Product Groups ====================

// ListProductGroupsResponse ผลลัพธ์จากการดึง/ค้นหากลุ่มสินค้า
type ListProductGroupsResponse struct {
	Groups      []ProductGroupDocument `json:"groups"`
	Count       int                    `json:"count"`
	Keyword     string                 `json:"keyword,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// ListProductGroups ดึง/ค้นหากลุ่มสินค้า
func ListProductGroups(ctx context.Context, holdingCode, keyword string, limit int) (*ListProductGroupsResponse, error) {
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

	filter := bson.M{
		"holding_code": holdingCode,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	if keyword != "" {
		keywordFilter := bson.M{
			"$or": []bson.M{
				{"code": bson.M{"$regex": keyword, "$options": "i"}},
				{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListProductGroups] holdingCode=%s, keyword=%s, limit=%d", holdingCode, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(productGroupCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []ProductGroupDocument
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}

	if groups == nil {
		groups = []ProductGroupDocument{}
	}

	return &ListProductGroupsResponse{
		Groups:      groups,
		Count:       len(groups),
		Keyword:     keyword,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Product Group ====================

// CreateProductGroupResponse ผลลัพธ์จากการสร้างกลุ่มสินค้า
type CreateProductGroupResponse struct {
	Success     bool                 `json:"success"`
	Message     string               `json:"message"`
	Group       ProductGroupDocument `json:"group"`
	KafkaSync   string               `json:"kafka_sync"`
	KafkaError  string               `json:"kafka_error,omitempty"`
	GeneratedAt time.Time            `json:"generated_at"`
}

// CreateProductGroup สร้างกลุ่มสินค้าใหม่
func CreateProductGroup(ctx context.Context, holdingCode, code, namesJSON string) (*CreateProductGroupResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productGroupCollection)

	// ตรวจสอบ code ซ้ำ
	existFilter := bson.M{
		"holding_code": holdingCode,
		"code":         code,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}
	count, err := coll.CountDocuments(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ code ซ้ำล้มเหลว: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("code '%s' มีอยู่แล้วใน shop นี้", code)
	}

	var names []ProductGroupNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required — ต้องระบุชื่อกลุ่มสินค้าอย่างน้อย 1 ภาษา เช่น [{\"code\":\"th\",\"name\":\"อาหาร\"}]")
	}

	now := time.Now()
	doc := ProductGroupDocument{
		HoldingCode: holdingCode,
		GuidFixed:   uuid.New().String(),
		Code:        code,
		Names:       names,
		CreatedBy:   "mcp-tool",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้างกลุ่มสินค้าล้มเหลว: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicProductGroupCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateProductGroup] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreateProductGroup] สร้าง code=%s สำเร็จ (shop=%s, kafka=%s)", code, holdingCode, kafkaSync)

	return &CreateProductGroupResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้างกลุ่มสินค้า '%s' สำเร็จ", code),
		Group:       doc,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Product Groups (Bulk) ====================

type CreateProductGroupsItem struct {
	Code  string                  `json:"code"`
	Names []ProductGroupNameEntry `json:"names"`
}

type CreateProductGroupsResponse struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Created      []ProductGroupDocument `json:"created"`
	Skipped      []string               `json:"skipped,omitempty"`
	CreatedCount int                    `json:"created_count"`
	SkippedCount int                    `json:"skipped_count"`
	KafkaSync    string                 `json:"kafka_sync"`
	KafkaError   string                 `json:"kafka_error,omitempty"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

func CreateProductGroups(ctx context.Context, holdingCode, groupsJSON string) (*CreateProductGroupsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if groupsJSON == "" {
		return nil, fmt.Errorf("groups JSON is required")
	}

	var items []CreateProductGroupsItem
	if err := json.Unmarshal([]byte(groupsJSON), &items); err != nil {
		return nil, fmt.Errorf("groups JSON ไม่ถูกต้อง: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ต้องระบุกลุ่มสินค้าอย่างน้อย 1 รายการ")
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
	coll := mongoClient.Database(dbName).Collection(productGroupCollection)

	var codes []string
	for _, item := range items {
		codes = append(codes, item.Code)
	}

	existFilter := bson.M{
		"holding_code": holdingCode,
		"code":         bson.M{"$in": codes},
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}
	cursor, err := coll.Find(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ code ซ้ำล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingCodes := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			Code string `bson:"code"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingCodes[doc.Code] = true
		}
	}

	now := time.Now()
	var docsToInsert []interface{}
	var createdDocs []ProductGroupDocument
	var skipped []string

	for _, item := range items {
		if item.Code == "" {
			skipped = append(skipped, "(code ว่าง)")
			continue
		}
		if existingCodes[item.Code] {
			skipped = append(skipped, fmt.Sprintf("%s (มีอยู่แล้ว)", item.Code))
			continue
		}
		if len(item.Names) == 0 {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี names)", item.Code))
			continue
		}

		doc := ProductGroupDocument{
			HoldingCode: holdingCode,
			GuidFixed:   uuid.New().String(),
			Code:        item.Code,
			Names:       item.Names,
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
			return nil, fmt.Errorf("สร้างกลุ่มสินค้าล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if err := publishToKafkaBatch(ctx, TopicProductGroupBulkCreated, docsToInsert); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP CreateProductGroups] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
		}
	}

	logger.Info("[MCP CreateProductGroups] สร้าง %d รายการสำเร็จ, ข้าม %d รายการ (shop=%s, kafka=%s)", len(createdDocs), len(skipped), holdingCode, kafkaSync)

	return &CreateProductGroupsResponse{
		Success:      true,
		Message:      fmt.Sprintf("สร้างกลุ่มสินค้า %d รายการสำเร็จ", len(createdDocs)),
		Created:      createdDocs,
		Skipped:      skipped,
		CreatedCount: len(createdDocs),
		SkippedCount: len(skipped),
		KafkaSync:    kafkaSync,
		KafkaError:   kafkaError,
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== Update Product Group ====================

type UpdateProductGroupResponse struct {
	Success     bool                 `json:"success"`
	Message     string               `json:"message"`
	Group       ProductGroupDocument `json:"group"`
	KafkaSync   string               `json:"kafka_sync"`
	KafkaError  string               `json:"kafka_error,omitempty"`
	GeneratedAt time.Time            `json:"generated_at"`
}

func UpdateProductGroup(ctx context.Context, holdingCode, code, namesJSON string) (*UpdateProductGroupResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productGroupCollection)

	filter := bson.M{
		"holding_code": holdingCode,
		"code":         code,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	var existing ProductGroupDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบกลุ่มสินค้า code='%s'", code)
	}

	updateFields := bson.M{
		"updated_at": time.Now(),
		"updatedby":  "mcp-tool",
	}
	if namesJSON != "" {
		var names []ProductGroupNameEntry
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["names"] = names
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดตกลุ่มสินค้าล้มเหลว: %w", err)
	}

	var updated ProductGroupDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicProductGroupUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdateProductGroup] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP UpdateProductGroup] อัปเดต code=%s สำเร็จ (shop=%s, kafka=%s)", code, holdingCode, kafkaSync)

	return &UpdateProductGroupResponse{
		Success:     true,
		Message:     fmt.Sprintf("อัปเดตกลุ่มสินค้า '%s' สำเร็จ", code),
		Group:       updated,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Product Group ====================

type DeleteProductGroupResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Code        string    `json:"code"`
	KafkaSync   string    `json:"kafka_sync"`
	KafkaError  string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

func DeleteProductGroup(ctx context.Context, holdingCode, code string) (*DeleteProductGroupResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productGroupCollection)

	filter := bson.M{
		"holding_code": holdingCode,
		"code":         code,
	}

	var docToDelete ProductGroupDocument
	coll.FindOne(ctx, filter).Decode(&docToDelete)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ลบกลุ่มสินค้าล้มเหลว: %w", err)
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("ไม่พบกลุ่มสินค้า code='%s'", code)
	}

	kafkaSync := "published"
	kafkaError := ""
	if docToDelete.Code != "" {
		if err := publishToKafka(ctx, TopicProductGroupDeleted, docToDelete); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP DeleteProductGroup] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
		}
	} else {
		kafkaSync = "skipped"
	}

	logger.Info("[MCP DeleteProductGroup] ลบ code=%s สำเร็จ (shop=%s, kafka=%s)", code, holdingCode, kafkaSync)

	return &DeleteProductGroupResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบกลุ่มสินค้า '%s' สำเร็จ", code),
		Code:        code,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Product Groups (Bulk) ====================

type DeleteProductGroupsResponse struct {
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

func DeleteProductGroups(ctx context.Context, holdingCode, codesJSON string) (*DeleteProductGroupsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if codesJSON == "" {
		return nil, fmt.Errorf("codes is required")
	}

	var codes []string
	if err := json.Unmarshal([]byte(codesJSON), &codes); err != nil {
		return nil, fmt.Errorf("codes JSON ไม่ถูกต้อง (ต้องเป็น array of strings): %w", err)
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("ต้องระบุ code อย่างน้อย 1 รายการ")
	}
	if len(codes) > 100 {
		return nil, fmt.Errorf("ลบได้สูงสุด 100 รายการต่อครั้ง")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productGroupCollection)

	existFilter := bson.M{
		"holding_code": holdingCode,
		"code":         bson.M{"$in": codes},
	}
	cursor, err := coll.Find(ctx, existFilter, options.Find().SetProjection(bson.M{"code": 1}))
	if err != nil {
		return nil, fmt.Errorf("ค้นหากลุ่มสินค้าล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingCodes := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			Code string `bson:"code"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingCodes[doc.Code] = true
		}
	}

	var deleted []string
	var notFound []string
	for _, c := range codes {
		if existingCodes[c] {
			deleted = append(deleted, c)
		} else {
			notFound = append(notFound, c)
		}
	}

	kafkaSync := "skipped"
	kafkaError := ""
	var docsToDelete []interface{}
	if len(deleted) > 0 {
		delCursor, _ := coll.Find(ctx, bson.M{
			"holding_code": holdingCode,
			"code":         bson.M{"$in": deleted},
		})
		if delCursor != nil {
			var docs []ProductGroupDocument
			delCursor.All(ctx, &docs)
			delCursor.Close(ctx)
			for _, d := range docs {
				docsToDelete = append(docsToDelete, d)
			}
		}
	}

	if len(deleted) > 0 {
		deleteFilter := bson.M{
			"holding_code": holdingCode,
			"code":         bson.M{"$in": deleted},
		}
		_, err = coll.DeleteMany(ctx, deleteFilter)
		if err != nil {
			return nil, fmt.Errorf("ลบกลุ่มสินค้าล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if len(docsToDelete) > 0 {
			if err := publishToKafkaBatch(ctx, TopicProductGroupBulkDeleted, docsToDelete); err != nil {
				kafkaSync = "failed"
				kafkaError = err.Error()
				logger.Warn("[MCP DeleteProductGroups] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
			}
		}
	}

	logger.Info("[MCP DeleteProductGroups] ลบ %d รายการสำเร็จ, ไม่พบ %d รายการ (shop=%s, kafka=%s)", len(deleted), len(notFound), holdingCode, kafkaSync)

	return &DeleteProductGroupsResponse{
		Success:       true,
		Message:       fmt.Sprintf("ลบกลุ่มสินค้า %d รายการสำเร็จ", len(deleted)),
		Deleted:       deleted,
		NotFound:      notFound,
		DeletedCount:  len(deleted),
		NotFoundCount: len(notFound),
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Get Product Group Schema ====================

func GetProductGroupSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  productGroupCollection,
		"description": "กลุ่มสินค้า (Product Group) — ใช้จัดกลุ่มสินค้า เช่น อาหาร, เครื่องดื่ม, อุปกรณ์",
		"fields": map[string]interface{}{
			"_id":          "ObjectID — MongoDB auto-generated ID",
			"holding_code": "string — Holding Code (tenant isolation)",
			"guid_fixed":   "string — UUID สำหรับอ้างอิงภายใน",
			"code":         "string (required, unique per shop) — รหัสกลุ่มสินค้า เช่น FOOD, DRINK, TOOL",
			"names":        "array (required) — ชื่อหลายภาษา [{code:'th', name:'อาหาร'}, {code:'en', name:'Food'}]",
			"createdby":    "string — ผู้สร้าง",
			"created_at":   "datetime — วันที่สร้าง",
			"updatedby":    "string — ผู้แก้ไขล่าสุด",
			"updated_at":   "datetime — วันที่แก้ไขล่าสุด",
			"deleted_at":   "datetime — วันที่ลบ (soft delete)",
		},
		"indexes": []string{
			"holding_code + code (unique per shop)",
			"holding_code + guidfixed",
		},
		"examples": []map[string]interface{}{
			{
				"code":  "FOOD",
				"names": []map[string]string{{"code": "th", "name": "อาหาร"}, {"code": "en", "name": "Food"}},
			},
			{
				"code":  "DRINK",
				"names": []map[string]string{{"code": "th", "name": "เครื่องดื่ม"}, {"code": "en", "name": "Beverage"}},
			},
		},
		"related_collections": []string{
			"productBarcodes — ใช้ groupcode อ้างอิง code ของกลุ่มสินค้า",
		},
	}
}

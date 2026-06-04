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

const productCategoryCollection = "productCategories"

// ==================== Product Category Types ====================

// ProductCategoryNameEntry ชื่อหมวดสินค้าแต่ละภาษา
type ProductCategoryNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// ProductCategoryDocument เอกสารหมวดสินค้าใน MongoDB
type ProductCategoryDocument struct {
	ID          primitive.ObjectID         `json:"id" bson:"_id,omitempty"`
	HoldingCode string                     `json:"holding_code" bson:"holding_code"`
	GuidFixed   string                     `json:"guid_fixed" bson:"guid_fixed"`
	Names       []ProductCategoryNameEntry `json:"names" bson:"names"`
	ParentGUID  string                     `json:"parent_guid,omitempty" bson:"parent_guid,omitempty"`
	GroupNumber int                        `json:"group_number,omitempty" bson:"group_number,omitempty"`
	ImageURI    string                     `json:"imageuri,omitempty" bson:"imageuri,omitempty"`
	IsDisabled  bool                       `json:"isdisabled,omitempty" bson:"isdisabled,omitempty"`
	CreatedBy   string                     `json:"createdby" bson:"createdby"`
	CreatedAt   time.Time                  `json:"created_at" bson:"created_at"`
	UpdatedBy   string                     `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	UpdatedAt   time.Time                  `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt   time.Time                  `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// ==================== List Product Categories ====================

type ListProductCategoriesResponse struct {
	Categories  []ProductCategoryDocument `json:"categories"`
	Count       int                       `json:"count"`
	Keyword     string                    `json:"keyword,omitempty"`
	GeneratedAt time.Time                 `json:"generated_at"`
}

func ListProductCategories(ctx context.Context, holdingCode, keyword string, limit int) (*ListProductCategoriesResponse, error) {
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
				{"guid_fixed": bson.M{"$regex": keyword, "$options": "i"}},
				{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListProductCategories] holdingCode=%s, keyword=%s, limit=%d", holdingCode, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "group_number", Value: 1}, {Key: "names.name", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var categories []ProductCategoryDocument
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}

	if categories == nil {
		categories = []ProductCategoryDocument{}
	}

	return &ListProductCategoriesResponse{
		Categories:  categories,
		Count:       len(categories),
		Keyword:     keyword,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Product Category ====================

type CreateProductCategoryResponse struct {
	Success     bool                    `json:"success"`
	Message     string                  `json:"message"`
	Category    ProductCategoryDocument `json:"category"`
	KafkaSync   string                  `json:"kafka_sync"`
	KafkaError  string                  `json:"kafka_error,omitempty"`
	GeneratedAt time.Time               `json:"generated_at"`
}

func CreateProductCategory(ctx context.Context, holdingCode, namesJSON, parentGUID string, groupNumber int) (*CreateProductCategoryResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	var names []ProductCategoryNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required — ต้องระบุชื่อหมวดสินค้าอย่างน้อย 1 ภาษา เช่น [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)

	now := time.Now()
	doc := ProductCategoryDocument{
		HoldingCode: holdingCode,
		GuidFixed:   uuid.New().String(),
		Names:       names,
		ParentGUID:  parentGUID,
		GroupNumber: groupNumber,
		CreatedBy:   "mcp-tool",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้างหมวดสินค้าล้มเหลว: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicProductCategoryCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateProductCategory] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreateProductCategory] สร้าง guidfixed=%s สำเร็จ (shop=%s, kafka=%s)", doc.GuidFixed, holdingCode, kafkaSync)

	return &CreateProductCategoryResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้างหมวดสินค้าสำเร็จ (guidfixed=%s)", doc.GuidFixed),
		Category:    doc,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Product Categories (Bulk) ====================

type CreateProductCategoriesItem struct {
	Names       []ProductCategoryNameEntry `json:"names"`
	ParentGUID  string                     `json:"parent_guid,omitempty"`
	GroupNumber int                        `json:"group_number,omitempty"`
}

type CreateProductCategoriesResponse struct {
	Success      bool                      `json:"success"`
	Message      string                    `json:"message"`
	Created      []ProductCategoryDocument `json:"created"`
	Skipped      []string                  `json:"skipped,omitempty"`
	CreatedCount int                       `json:"created_count"`
	SkippedCount int                       `json:"skipped_count"`
	KafkaSync    string                    `json:"kafka_sync"`
	KafkaError   string                    `json:"kafka_error,omitempty"`
	GeneratedAt  time.Time                 `json:"generated_at"`
}

func CreateProductCategories(ctx context.Context, holdingCode, categoriesJSON string) (*CreateProductCategoriesResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if categoriesJSON == "" {
		return nil, fmt.Errorf("categories JSON is required")
	}

	var items []CreateProductCategoriesItem
	if err := json.Unmarshal([]byte(categoriesJSON), &items); err != nil {
		return nil, fmt.Errorf("categories JSON ไม่ถูกต้อง: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ต้องระบุหมวดสินค้าอย่างน้อย 1 รายการ")
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
	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)

	now := time.Now()
	var docsToInsert []interface{}
	var createdDocs []ProductCategoryDocument
	var skipped []string

	for i, item := range items {
		if len(item.Names) == 0 {
			skipped = append(skipped, fmt.Sprintf("รายการที่ %d (ไม่มี names)", i+1))
			continue
		}

		doc := ProductCategoryDocument{
			HoldingCode: holdingCode,
			GuidFixed:   uuid.New().String(),
			Names:       item.Names,
			ParentGUID:  item.ParentGUID,
			GroupNumber: item.GroupNumber,
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
		_, err := coll.InsertMany(ctx, docsToInsert)
		if err != nil {
			return nil, fmt.Errorf("สร้างหมวดสินค้าล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if err := publishToKafkaBatch(ctx, TopicProductCategoryBulkCreated, docsToInsert); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP CreateProductCategories] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
		}
	}

	logger.Info("[MCP CreateProductCategories] สร้าง %d รายการสำเร็จ, ข้าม %d รายการ (shop=%s, kafka=%s)", len(createdDocs), len(skipped), holdingCode, kafkaSync)

	return &CreateProductCategoriesResponse{
		Success:      true,
		Message:      fmt.Sprintf("สร้างหมวดสินค้า %d รายการสำเร็จ", len(createdDocs)),
		Created:      createdDocs,
		Skipped:      skipped,
		CreatedCount: len(createdDocs),
		SkippedCount: len(skipped),
		KafkaSync:    kafkaSync,
		KafkaError:   kafkaError,
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== Update Product Category ====================

type UpdateProductCategoryResponse struct {
	Success     bool                    `json:"success"`
	Message     string                  `json:"message"`
	Category    ProductCategoryDocument `json:"category"`
	KafkaSync   string                  `json:"kafka_sync"`
	KafkaError  string                  `json:"kafka_error,omitempty"`
	GeneratedAt time.Time               `json:"generated_at"`
}

func UpdateProductCategory(ctx context.Context, holdingCode, guidFixed, namesJSON, parentGUID string, groupNumber int) (*UpdateProductCategoryResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if guidFixed == "" {
		return nil, fmt.Errorf("guidfixed is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)

	filter := bson.M{
		"holding_code": holdingCode,
		"guid_fixed":   guidFixed,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	var existing ProductCategoryDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบหมวดสินค้า guidfixed='%s'", guidFixed)
	}

	updateFields := bson.M{
		"updated_at": time.Now(),
		"updatedby":  "mcp-tool",
	}
	if namesJSON != "" {
		var names []ProductCategoryNameEntry
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["names"] = names
	}
	if parentGUID != "" {
		updateFields["parent_guid"] = parentGUID
	}
	if groupNumber > 0 {
		updateFields["group_number"] = groupNumber
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดตหมวดสินค้าล้มเหลว: %w", err)
	}

	var updated ProductCategoryDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicProductCategoryUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdateProductCategory] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP UpdateProductCategory] อัปเดต guidfixed=%s สำเร็จ (shop=%s, kafka=%s)", guidFixed, holdingCode, kafkaSync)

	return &UpdateProductCategoryResponse{
		Success:     true,
		Message:     fmt.Sprintf("อัปเดตหมวดสินค้า '%s' สำเร็จ", guidFixed),
		Category:    updated,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Product Category ====================

type DeleteProductCategoryResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	GuidFixed   string    `json:"guid_fixed"`
	KafkaSync   string    `json:"kafka_sync"`
	KafkaError  string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

func DeleteProductCategory(ctx context.Context, holdingCode, guidFixed string) (*DeleteProductCategoryResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if guidFixed == "" {
		return nil, fmt.Errorf("guidfixed is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)

	filter := bson.M{
		"holding_code": holdingCode,
		"guid_fixed":   guidFixed,
	}

	var docToDelete ProductCategoryDocument
	coll.FindOne(ctx, filter).Decode(&docToDelete)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ลบหมวดสินค้าล้มเหลว: %w", err)
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("ไม่พบหมวดสินค้า guidfixed='%s'", guidFixed)
	}

	kafkaSync := "published"
	kafkaError := ""
	if docToDelete.GuidFixed != "" {
		if err := publishToKafka(ctx, TopicProductCategoryDeleted, docToDelete); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP DeleteProductCategory] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
		}
	} else {
		kafkaSync = "skipped"
	}

	logger.Info("[MCP DeleteProductCategory] ลบ guidfixed=%s สำเร็จ (shop=%s, kafka=%s)", guidFixed, holdingCode, kafkaSync)

	return &DeleteProductCategoryResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบหมวดสินค้า '%s' สำเร็จ", guidFixed),
		GuidFixed:   guidFixed,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Product Categories (Bulk) ====================

type DeleteProductCategoriesResponse struct {
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

func DeleteProductCategories(ctx context.Context, holdingCode, guidfixedsJSON string) (*DeleteProductCategoriesResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holding_code is required")
	}
	if guidfixedsJSON == "" {
		return nil, fmt.Errorf("guidfixeds is required")
	}

	var guidfixeds []string
	if err := json.Unmarshal([]byte(guidfixedsJSON), &guidfixeds); err != nil {
		return nil, fmt.Errorf("guidfixeds JSON ไม่ถูกต้อง (ต้องเป็น array of strings): %w", err)
	}
	if len(guidfixeds) == 0 {
		return nil, fmt.Errorf("ต้องระบุ guidfixed อย่างน้อย 1 รายการ")
	}
	if len(guidfixeds) > 100 {
		return nil, fmt.Errorf("ลบได้สูงสุด 100 รายการต่อครั้ง")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)

	existFilter := bson.M{
		"holding_code": holdingCode,
		"guid_fixed":   bson.M{"$in": guidfixeds},
	}
	cursor, err := coll.Find(ctx, existFilter, options.Find().SetProjection(bson.M{"guid_fixed": 1}))
	if err != nil {
		return nil, fmt.Errorf("ค้นหาหมวดสินค้าล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingGuids := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			GuidFixed string `bson:"guid_fixed"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingGuids[doc.GuidFixed] = true
		}
	}

	var deleted []string
	var notFound []string
	for _, g := range guidfixeds {
		if existingGuids[g] {
			deleted = append(deleted, g)
		} else {
			notFound = append(notFound, g)
		}
	}

	kafkaSync := "skipped"
	kafkaError := ""
	var docsToDelete []interface{}
	if len(deleted) > 0 {
		delCursor, _ := coll.Find(ctx, bson.M{
			"holding_code": holdingCode,
			"guid_fixed":   bson.M{"$in": deleted},
		})
		if delCursor != nil {
			var docs []ProductCategoryDocument
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
			"guid_fixed":   bson.M{"$in": deleted},
		}
		_, err = coll.DeleteMany(ctx, deleteFilter)
		if err != nil {
			return nil, fmt.Errorf("ลบหมวดสินค้าล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if len(docsToDelete) > 0 {
			if err := publishToKafkaBatch(ctx, TopicProductCategoryBulkDeleted, docsToDelete); err != nil {
				kafkaSync = "failed"
				kafkaError = err.Error()
				logger.Warn("[MCP DeleteProductCategories] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
			}
		}
	}

	logger.Info("[MCP DeleteProductCategories] ลบ %d รายการสำเร็จ, ไม่พบ %d รายการ (shop=%s, kafka=%s)", len(deleted), len(notFound), holdingCode, kafkaSync)

	return &DeleteProductCategoriesResponse{
		Success:       true,
		Message:       fmt.Sprintf("ลบหมวดสินค้า %d รายการสำเร็จ", len(deleted)),
		Deleted:       deleted,
		NotFound:      notFound,
		DeletedCount:  len(deleted),
		NotFoundCount: len(notFound),
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Get Product Category Schema ====================

func GetProductCategorySchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  productCategoryCollection,
		"description": "หมวดสินค้า (Product Category) — ใช้จัดหมวดหมู่สินค้าแบบ hierarchy เช่น เนื้อสัตว์, ผัก, เครื่องปรุง",
		"fields": map[string]interface{}{
			"_id":          "ObjectID — MongoDB auto-generated ID",
			"holding_code": "string — Holding Code (tenant isolation)",
			"guid_fixed":   "string — UUID สำหรับอ้างอิง (ใช้เป็น key สำหรับ update/delete)",
			"names":        "array (required) — ชื่อหลายภาษา [{code:'th', name:'เนื้อสัตว์'}, {code:'en', name:'Meat'}]",
			"parent_guid":  "string (optional) — guidfixed ของหมวดแม่ (สำหรับ hierarchy)",
			"group_number": "number (optional) — ลำดับกลุ่ม/หมวด",
			"imageuri":     "string (optional) — URL รูปภาพหมวดสินค้า",
			"isdisabled":   "boolean (optional) — ปิดใช้งานหมวดนี้",
			"createdby":    "string — ผู้สร้าง",
			"created_at":   "datetime — วันที่สร้าง",
			"updatedby":    "string — ผู้แก้ไขล่าสุด",
			"updated_at":   "datetime — วันที่แก้ไขล่าสุด",
			"deleted_at":   "datetime — วันที่ลบ (soft delete)",
		},
		"indexes": []string{
			"holding_code + guidfixed (unique)",
			"holding_code + parentguid",
			"holding_code + groupnumber",
		},
		"examples": []map[string]interface{}{
			{
				"names":        []map[string]string{{"code": "th", "name": "เนื้อสัตว์"}, {"code": "en", "name": "Meat"}},
				"group_number": 1,
			},
			{
				"names":        []map[string]string{{"code": "th", "name": "ผักสด"}, {"code": "en", "name": "Fresh Vegetables"}},
				"group_number": 2,
			},
		},
		"related_collections": []string{
			"productBarcodes — ใช้ categorycode (= guidfixed) อ้างอิงหมวดสินค้า",
		},
	}
}

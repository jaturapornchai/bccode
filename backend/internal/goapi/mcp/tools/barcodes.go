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

const barcodeCollection = "productbarcodes"

// ==================== Auto-Lookup Helpers ====================

// lookupUnitNames ดึง names จาก units collection ตาม unitcode
// ใช้เมื่อผู้ใช้ระบุ itemunitcode แต่ไม่ได้ระบุ itemunitnames
func lookupUnitNames(ctx context.Context, holdingCode, unitCode string) []BarcodeNameEntry {
	if unitCode == "" {
		return nil
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(unitCollection)

	var doc struct {
		Names []BarcodeNameEntry `bson:"names"`
	}
	err := coll.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"unitcode":    unitCode,
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}).Decode(&doc)
	if err != nil {
		logger.Warn("[MCP] lookup unitnames for '%s' ไม่พบ: %v", unitCode, err)
		return nil
	}

	logger.Info("[MCP] auto-fill unitnames จาก units collection: unitcode=%s, names=%v", unitCode, doc.Names)
	return doc.Names
}

// lookupGroupNames ดึง names จาก productGroups collection ตาม groupcode
func lookupGroupNames(ctx context.Context, holdingCode, groupCode string) []BarcodeNameEntry {
	if groupCode == "" {
		return nil
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productGroupCollection)

	var doc struct {
		Names []BarcodeNameEntry `bson:"names"`
	}
	err := coll.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"code":        groupCode,
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}).Decode(&doc)
	if err != nil {
		logger.Warn("[MCP] lookup groupnames for '%s' ไม่พบ: %v", groupCode, err)
		return nil
	}

	logger.Info("[MCP] auto-fill groupnames จาก productGroups collection: code=%s", groupCode)
	return doc.Names
}

// lookupCategoryNames ดึง names จาก productCategories collection ตาม guidfixed
func lookupCategoryNames(ctx context.Context, holdingCode, categoryCode string) []BarcodeNameEntry {
	if categoryCode == "" {
		return nil
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(productCategoryCollection)

	var doc struct {
		Names []BarcodeNameEntry `bson:"names"`
	}
	err := coll.FindOne(ctx, bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   categoryCode,
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}).Decode(&doc)
	if err != nil {
		logger.Warn("[MCP] lookup categorynames for '%s' ไม่พบ: %v", categoryCode, err)
		return nil
	}

	logger.Info("[MCP] auto-fill categorynames จาก productCategories collection: guidfixed=%s", categoryCode)
	return doc.Names
}

// ==================== Barcode Types ====================

// BarcodeNameEntry ชื่อสินค้า/หน่วยนับแต่ละภาษา
type BarcodeNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// BarcodePriceEntry ราคาสินค้า
type BarcodePriceEntry struct {
	KeyNumber int     `json:"keynumber" bson:"keynumber"`
	Price     float64 `json:"price" bson:"price"`
}

// BarcodeDocument เอกสาร barcode ใน MongoDB (fields หลักที่ MCP ใช้)
type BarcodeDocument struct {
	ID               primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	HoldingCode      string              `json:"holdingcode" bson:"holdingcode"`
	GuidFixed        string              `json:"guidfixed" bson:"guidfixed"`
	Barcode          string              `json:"barcode" bson:"barcode"`
	ItemCode         string              `json:"itemcode" bson:"itemcode"`
	Names            []BarcodeNameEntry  `json:"names" bson:"names"`
	ItemUnitCode     string              `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames    []BarcodeNameEntry  `json:"itemunitnames" bson:"itemunitnames"`
	Prices           []BarcodePriceEntry `json:"prices" bson:"prices"`
	StandValue       float64             `json:"standvalue" bson:"standvalue"`
	DivideValue      float64             `json:"dividevalue" bson:"dividevalue"`
	IsMainBarcode    bool                `json:"ismainbarcode" bson:"ismainbarcode"`
	ImageURI         string              `json:"imageuri" bson:"imageuri"`
	GroupCode        string              `json:"groupcode" bson:"groupcode"`
	GroupNames       []BarcodeNameEntry  `json:"groupnames" bson:"groupnames"`
	BrandCode        string              `json:"brandcode" bson:"brandcode"`
	BrandNames       []BarcodeNameEntry  `json:"brandnames" bson:"brandnames"`
	CategoryCode     string              `json:"categorycode" bson:"categorycode"`
	CategoryNames    []BarcodeNameEntry  `json:"categorynames" bson:"categorynames"`
	RefBarcodes      []RefBarcodeEntry   `json:"refbarcodes,omitempty" bson:"refbarcodes,omitempty"`
	IsUseSubBarcodes bool                `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	ItemType         int8                `json:"itemtype" bson:"itemtype"`
	MaterialType     int8                `json:"materialtype" bson:"materialtype"`
	TaxType          int8                `json:"taxtype" bson:"taxtype"`
	VatType          int8                `json:"vattype" bson:"vattype"`
	CreatedBy        string              `json:"createdby" bson:"createdby"`
	CreatedAt        time.Time           `json:"createdat" bson:"createdat"`
	UpdatedBy        string              `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	UpdatedAt        time.Time           `json:"updatedat,omitempty" bson:"updatedat,omitempty"`
	DeletedBy        string              `json:"deletedby,omitempty" bson:"deletedby,omitempty"`
	DeletedAt        time.Time           `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
}

// ==================== List Barcodes ====================

// ListBarcodesResponse ผลลัพธ์จากการดึง/ค้นหา barcode
type ListBarcodesResponse struct {
	Barcodes    []BarcodeDocument `json:"barcodes"`
	Count       int               `json:"count"`
	Keyword     string            `json:"keyword,omitempty"`
	GeneratedAt time.Time         `json:"generatedat"`
}

// ListBarcodes ดึง/ค้นหา barcode
func ListBarcodes(ctx context.Context, holdingCode, keyword string, limit int) (*ListBarcodesResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
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

	// สร้าง filter — ไม่รวม soft deleted
	filter := bson.M{
		"holdingcode": holdingCode,
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}

	// ถ้ามี keyword → ค้นหาใน barcode, itemcode, names.name
	if keyword != "" {
		keywordFilter := bson.M{
			"$or": []bson.M{
				{"barcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"itemcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListBarcodes] holdingCode=%s, keyword=%s, limit=%d", holdingCode, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(barcodeCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "barcode", Value: 1}}).
		SetProjection(bson.M{
			"holdingcode": 1, "guidfixed": 1, "barcode": 1, "itemcode": 1,
			"names": 1, "itemunitcode": 1, "itemunitnames": 1,
			"prices": 1, "standvalue": 1, "dividevalue": 1,
			"ismainbarcode": 1, "isusesubbarcodes": 1,
			"refbarcodes": 1, "imageuri": 1,
			"groupcode": 1, "groupnames": 1,
			"brandcode": 1, "brandnames": 1,
			"categorycode": 1, "categorynames": 1,
			"itemtype": 1, "materialtype": 1, "taxtype": 1, "vattype": 1,
			"createdby": 1, "createdat": 1,
			"updatedby": 1, "updatedat": 1,
		})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var barcodes []BarcodeDocument
	if err := cursor.All(ctx, &barcodes); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}

	if barcodes == nil {
		barcodes = []BarcodeDocument{}
	}

	return &ListBarcodesResponse{
		Barcodes:    barcodes,
		Count:       len(barcodes),
		Keyword:     keyword,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Barcode ====================

// CreateBarcodeResponse ผลลัพธ์จากการสร้าง barcode
type CreateBarcodeResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message"`
	Barcode     BarcodeDocument `json:"barcode"`
	KafkaSync   string          `json:"kafkasync"`
	KafkaError  string          `json:"kafkaerror,omitempty"`
	GeneratedAt time.Time       `json:"generatedat"`
}

// CreateBarcode สร้าง barcode ใหม่
func CreateBarcode(ctx context.Context, holdingCode, barcode, itemCode, namesJSON, itemUnitCode, itemUnitNamesJSON, pricesJSON string, standValue, divideValue float64, isMainBarcode *bool, groupCode, groupNamesJSON, categoryCode, categoryNamesJSON string) (*CreateBarcodeResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}
	if itemCode == "" {
		return nil, fmt.Errorf("itemcode is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	// ตรวจสอบ barcode ซ้ำ
	existFilter := bson.M{
		"holdingcode": holdingCode,
		"barcode":     barcode,
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}
	count, err := coll.CountDocuments(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ barcode ซ้ำล้มเหลว: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("barcode '%s' มีอยู่แล้วใน shop นี้", barcode)
	}

	// Parse names JSON
	var names []BarcodeNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required — ต้องระบุชื่อสินค้าอย่างน้อย 1 ภาษา เช่น [{\"code\":\"th\",\"name\":\"สินค้า A\"}]")
	}

	// Parse unit names JSON (optional)
	var itemUnitNames []BarcodeNameEntry
	if itemUnitNamesJSON != "" {
		if err := json.Unmarshal([]byte(itemUnitNamesJSON), &itemUnitNames); err != nil {
			return nil, fmt.Errorf("itemunitnames JSON ไม่ถูกต้อง: %w", err)
		}
	}
	// Auto-lookup: ถ้าไม่ได้ระบุ itemunitnames แต่ระบุ itemunitcode → ดึงจาก units collection
	if len(itemUnitNames) == 0 && itemUnitCode != "" {
		if looked := lookupUnitNames(ctx, holdingCode, itemUnitCode); looked != nil {
			itemUnitNames = looked
		}
	}

	// Parse prices JSON (optional)
	var prices []BarcodePriceEntry
	if pricesJSON != "" {
		if err := json.Unmarshal([]byte(pricesJSON), &prices); err != nil {
			return nil, fmt.Errorf("prices JSON ไม่ถูกต้อง: %w", err)
		}
	}

	// Default standvalue/dividevalue
	if standValue == 0 {
		standValue = 1
	}
	if divideValue == 0 {
		divideValue = 1
	}

	// Parse group names JSON (optional)
	var groupNames []BarcodeNameEntry
	if groupNamesJSON != "" {
		if err := json.Unmarshal([]byte(groupNamesJSON), &groupNames); err != nil {
			return nil, fmt.Errorf("groupnames JSON ไม่ถูกต้อง: %w", err)
		}
	}
	// Auto-lookup: ถ้าไม่ได้ระบุ groupnames แต่ระบุ groupcode → ดึงจาก productGroups collection
	if len(groupNames) == 0 && groupCode != "" {
		if looked := lookupGroupNames(ctx, holdingCode, groupCode); looked != nil {
			groupNames = looked
		}
	}

	// Parse category names JSON (optional)
	var categoryNames []BarcodeNameEntry
	if categoryNamesJSON != "" {
		if err := json.Unmarshal([]byte(categoryNamesJSON), &categoryNames); err != nil {
			return nil, fmt.Errorf("categorynames JSON ไม่ถูกต้อง: %w", err)
		}
	}
	// Auto-lookup: ถ้าไม่ได้ระบุ categorynames แต่ระบุ categorycode → ดึงจาก productCategories collection
	if len(categoryNames) == 0 && categoryCode != "" {
		if looked := lookupCategoryNames(ctx, holdingCode, categoryCode); looked != nil {
			categoryNames = looked
		}
	}

	// Auto-detect ismainbarcode: หน่วยฐาน (standvalue=1, dividevalue=1) = true
	isMain := standValue == 1 && divideValue == 1
	if isMainBarcode != nil {
		isMain = *isMainBarcode
	}

	now := time.Now()
	guidFixed := uuid.New().String()
	doc := BarcodeDocument{
		HoldingCode:   holdingCode,
		GuidFixed:     guidFixed,
		Barcode:       barcode,
		ItemCode:      itemCode,
		Names:         names,
		ItemUnitCode:  itemUnitCode,
		ItemUnitNames: itemUnitNames,
		Prices:        prices,
		StandValue:    standValue,
		DivideValue:   divideValue,
		IsMainBarcode: isMain,
		GroupCode:     groupCode,
		GroupNames:    groupNames,
		CategoryCode:  categoryCode,
		CategoryNames: categoryNames,
		CreatedBy:     "mcp-tool",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้าง barcode ล้มเหลว: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	// Publish Kafka event (เหมือน mainapi)
	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicBarcodeCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateBarcode] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreateBarcode] สร้าง barcode=%s สำเร็จ (shop=%s, kafka=%s)", barcode, holdingCode, kafkaSync)

	return &CreateBarcodeResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้าง barcode '%s' สำเร็จ", barcode),
		Barcode:     doc,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Barcodes (Bulk) ====================

// CreateBarcodesItem รายการ barcode ที่ต้องการสร้างพร้อมกัน
type CreateBarcodesItem struct {
	Barcode       string              `json:"barcode"`
	ItemCode      string              `json:"itemcode"`
	Names         []BarcodeNameEntry  `json:"names"`
	ItemUnitCode  string              `json:"itemunitcode"`
	ItemUnitNames []BarcodeNameEntry  `json:"itemunitnames"`
	Prices        []BarcodePriceEntry `json:"prices"`
	StandValue    float64             `json:"standvalue"`
	DivideValue   float64             `json:"dividevalue"`
	IsMainBarcode *bool               `json:"ismainbarcode"`
}

// CreateBarcodesResponse ผลลัพธ์จากการสร้าง barcode หลายรายการ
type CreateBarcodesResponse struct {
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	Created      []BarcodeDocument `json:"created"`
	Skipped      []string          `json:"skipped,omitempty"`
	CreatedCount int               `json:"createdcount"`
	SkippedCount int               `json:"skippedcount"`
	KafkaSync    string            `json:"kafkasync"`
	KafkaError   string            `json:"kafkaerror,omitempty"`
	GeneratedAt  time.Time         `json:"generatedat"`
}

// CreateBarcodes สร้าง barcode หลายรายการพร้อมกัน
func CreateBarcodes(ctx context.Context, holdingCode, barcodesJSON string) (*CreateBarcodesResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if barcodesJSON == "" {
		return nil, fmt.Errorf("barcodes JSON is required")
	}

	var items []CreateBarcodesItem
	if err := json.Unmarshal([]byte(barcodesJSON), &items); err != nil {
		return nil, fmt.Errorf("barcodes JSON ไม่ถูกต้อง: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ต้องระบุ barcode อย่างน้อย 1 รายการ")
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
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	// ดึง barcode ที่มีอยู่แล้ว
	var barcodeValues []string
	for _, item := range items {
		barcodeValues = append(barcodeValues, item.Barcode)
	}

	existFilter := bson.M{
		"holdingcode": holdingCode,
		"barcode":     bson.M{"$in": barcodeValues},
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}
	cursor, err := coll.Find(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ barcode ซ้ำล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingBarcodes := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			Barcode string `bson:"barcode"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingBarcodes[doc.Barcode] = true
		}
	}

	now := time.Now()
	var docsToInsert []interface{}
	var createdDocs []BarcodeDocument
	var skipped []string

	for _, item := range items {
		if item.Barcode == "" {
			skipped = append(skipped, "(barcode ว่าง)")
			continue
		}
		if item.ItemCode == "" {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี itemcode)", item.Barcode))
			continue
		}
		if existingBarcodes[item.Barcode] {
			skipped = append(skipped, fmt.Sprintf("%s (มีอยู่แล้ว)", item.Barcode))
			continue
		}
		if len(item.Names) == 0 {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี names)", item.Barcode))
			continue
		}
		if item.ItemUnitCode == "" {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี itemunitcode)", item.Barcode))
			continue
		}
		if item.StandValue == 0 {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี standvalue)", item.Barcode))
			continue
		}
		if item.DivideValue == 0 {
			skipped = append(skipped, fmt.Sprintf("%s (ไม่มี dividevalue)", item.Barcode))
			continue
		}

		// Auto-detect ismainbarcode: หน่วยฐาน (standvalue=1, dividevalue=1) = true
		isMain := item.StandValue == 1 && item.DivideValue == 1
		if item.IsMainBarcode != nil {
			isMain = *item.IsMainBarcode
		}

		// Auto-lookup: ถ้าไม่ได้ระบุ itemunitnames แต่ระบุ itemunitcode → ดึงจาก units collection
		unitNames := item.ItemUnitNames
		if len(unitNames) == 0 && item.ItemUnitCode != "" {
			if looked := lookupUnitNames(ctx, holdingCode, item.ItemUnitCode); looked != nil {
				unitNames = looked
			}
		}

		doc := BarcodeDocument{
			HoldingCode:   holdingCode,
			GuidFixed:     uuid.New().String(),
			Barcode:       item.Barcode,
			ItemCode:      item.ItemCode,
			Names:         item.Names,
			ItemUnitCode:  item.ItemUnitCode,
			ItemUnitNames: unitNames,
			Prices:        item.Prices,
			StandValue:    item.StandValue,
			DivideValue:   item.DivideValue,
			IsMainBarcode: isMain,
			CreatedBy:     "mcp-tool",
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		docsToInsert = append(docsToInsert, doc)
		createdDocs = append(createdDocs, doc)
	}

	kafkaSync := "skipped"
	kafkaError := ""

	if len(docsToInsert) > 0 {
		_, err = coll.InsertMany(ctx, docsToInsert)
		if err != nil {
			return nil, fmt.Errorf("สร้าง barcode ล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		// Publish Kafka event — bulk created (เหมือน mainapi)
		if err := publishToKafkaBatch(ctx, TopicBarcodeBulkCreated, docsToInsert); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP CreateBarcodes] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
		}
	}

	logger.Info("[MCP CreateBarcodes] สร้าง %d รายการสำเร็จ, ข้าม %d รายการ (shop=%s, kafka=%s)", len(createdDocs), len(skipped), holdingCode, kafkaSync)

	return &CreateBarcodesResponse{
		Success:      true,
		Message:      fmt.Sprintf("สร้าง barcode %d รายการสำเร็จ", len(createdDocs)),
		Created:      createdDocs,
		Skipped:      skipped,
		CreatedCount: len(createdDocs),
		SkippedCount: len(skipped),
		KafkaSync:    kafkaSync,
		KafkaError:   kafkaError,
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== Update Barcode ====================

// UpdateBarcodeResponse ผลลัพธ์จากการอัปเดต barcode
type UpdateBarcodeResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message"`
	Barcode     BarcodeDocument `json:"barcode"`
	KafkaSync   string          `json:"kafkasync"`
	KafkaError  string          `json:"kafkaerror,omitempty"`
	GeneratedAt time.Time       `json:"generatedat"`
}

// UpdateBarcode อัปเดต barcode ตาม guidfixed
func UpdateBarcode(ctx context.Context, holdingCode, guidFixed, namesJSON, itemUnitCode, itemUnitNamesJSON, pricesJSON, groupCode, groupNamesJSON, categoryCode, categoryNamesJSON string) (*UpdateBarcodeResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
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
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	// ค้นหา barcode ที่ต้องการอัปเดต
	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guidFixed,
		"$or": []bson.M{
			{"deletedby": bson.M{"$exists": false}},
			{"deletedby": ""},
		},
	}

	var existing BarcodeDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบ barcode guidfixed='%s'", guidFixed)
	}

	// สร้าง update fields
	updateFields := bson.M{
		"updatedat": time.Now(),
		"updatedby": "mcp-tool",
	}
	if namesJSON != "" {
		var names []BarcodeNameEntry
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["names"] = names
	}
	if itemUnitCode != "" {
		updateFields["itemunitcode"] = itemUnitCode
	}
	if itemUnitNamesJSON != "" {
		var itemUnitNames []BarcodeNameEntry
		if err := json.Unmarshal([]byte(itemUnitNamesJSON), &itemUnitNames); err != nil {
			return nil, fmt.Errorf("itemunitnames JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["itemunitnames"] = itemUnitNames
	} else if itemUnitCode != "" {
		// Auto-lookup: ถ้าเปลี่ยน itemunitcode แต่ไม่ได้ระบุ itemunitnames → ดึงจาก units collection
		if looked := lookupUnitNames(ctx, holdingCode, itemUnitCode); looked != nil {
			updateFields["itemunitnames"] = looked
		}
	}
	if pricesJSON != "" {
		var prices []BarcodePriceEntry
		if err := json.Unmarshal([]byte(pricesJSON), &prices); err != nil {
			return nil, fmt.Errorf("prices JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["prices"] = prices
	}
	if groupCode != "" {
		updateFields["groupcode"] = groupCode
	}
	if groupNamesJSON != "" {
		var groupNames []BarcodeNameEntry
		if err := json.Unmarshal([]byte(groupNamesJSON), &groupNames); err != nil {
			return nil, fmt.Errorf("groupnames JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["groupnames"] = groupNames
	} else if groupCode != "" {
		// Auto-lookup: ถ้าเปลี่ยน groupcode แต่ไม่ได้ระบุ groupnames → ดึงจาก productGroups collection
		if looked := lookupGroupNames(ctx, holdingCode, groupCode); looked != nil {
			updateFields["groupnames"] = looked
		}
	}
	if categoryCode != "" {
		updateFields["categorycode"] = categoryCode
	}
	if categoryNamesJSON != "" {
		var categoryNames []BarcodeNameEntry
		if err := json.Unmarshal([]byte(categoryNamesJSON), &categoryNames); err != nil {
			return nil, fmt.Errorf("categorynames JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["categorynames"] = categoryNames
	} else if categoryCode != "" {
		// Auto-lookup: ถ้าเปลี่ยน categorycode แต่ไม่ได้ระบุ categorynames → ดึงจาก productCategories collection
		if looked := lookupCategoryNames(ctx, holdingCode, categoryCode); looked != nil {
			updateFields["categorynames"] = looked
		}
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดต barcode ล้มเหลว: %w", err)
	}

	// ดึง document ที่อัปเดตแล้ว
	var updated BarcodeDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	// Publish Kafka event (เหมือน mainapi)
	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicBarcodeUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdateBarcode] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP UpdateBarcode] อัปเดต guidfixed=%s สำเร็จ (shop=%s, kafka=%s)", guidFixed, holdingCode, kafkaSync)

	return &UpdateBarcodeResponse{
		Success:     true,
		Message:     fmt.Sprintf("อัปเดต barcode '%s' สำเร็จ", existing.Barcode),
		Barcode:     updated,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Barcode ====================

// DeleteBarcodeResponse ผลลัพธ์จากการลบ barcode
type DeleteBarcodeResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	GuidFixed   string    `json:"guidfixed"`
	KafkaSync   string    `json:"kafkasync"`
	KafkaError  string    `json:"kafkaerror,omitempty"`
	GeneratedAt time.Time `json:"generatedat"`
}

// DeleteBarcode ลบ barcode ตาม guidfixed
func DeleteBarcode(ctx context.Context, holdingCode, guidFixed string) (*DeleteBarcodeResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
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
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guidFixed,
	}

	// ดึง document ก่อนลบ เพื่อส่ง Kafka event
	var docToDelete BarcodeDocument
	coll.FindOne(ctx, filter).Decode(&docToDelete)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ลบ barcode ล้มเหลว: %w", err)
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("ไม่พบ barcode guidfixed='%s'", guidFixed)
	}

	// Publish Kafka event (เหมือน mainapi)
	kafkaSync := "published"
	kafkaError := ""
	if docToDelete.Barcode != "" {
		if err := publishToKafka(ctx, TopicBarcodeDeleted, docToDelete); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP DeleteBarcode] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
		}
	} else {
		kafkaSync = "skipped"
	}

	logger.Info("[MCP DeleteBarcode] ลบ guidfixed=%s สำเร็จ (shop=%s, kafka=%s)", guidFixed, holdingCode, kafkaSync)

	return &DeleteBarcodeResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบ barcode สำเร็จ (guidfixed=%s)", guidFixed),
		GuidFixed:   guidFixed,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Barcodes (Bulk) ====================

// DeleteBarcodesResponse ผลลัพธ์จากการลบ barcode หลายรายการ
type DeleteBarcodesResponse struct {
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	Deleted       []string  `json:"deleted"`
	NotFound      []string  `json:"notfound,omitempty"`
	DeletedCount  int       `json:"deletedcount"`
	NotFoundCount int       `json:"notfoundcount"`
	KafkaSync     string    `json:"kafkasync"`
	KafkaError    string    `json:"kafkaerror,omitempty"`
	GeneratedAt   time.Time `json:"generatedat"`
}

// DeleteBarcodes ลบ barcode หลายรายการพร้อมกัน
func DeleteBarcodes(ctx context.Context, holdingCode, guidfixedsJSON string) (*DeleteBarcodesResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if guidfixedsJSON == "" {
		return nil, fmt.Errorf("guidfixeds is required")
	}

	var guidFixeds []string
	if err := json.Unmarshal([]byte(guidfixedsJSON), &guidFixeds); err != nil {
		return nil, fmt.Errorf("guidfixeds JSON ไม่ถูกต้อง (ต้องเป็น array of strings): %w", err)
	}
	if len(guidFixeds) == 0 {
		return nil, fmt.Errorf("ต้องระบุ guidfixed อย่างน้อย 1 รายการ")
	}
	if len(guidFixeds) > 100 {
		return nil, fmt.Errorf("ลบได้สูงสุด 100 รายการต่อครั้ง")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	// ค้นหา guidfixed ที่มีอยู่จริง
	existFilter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   bson.M{"$in": guidFixeds},
	}
	cursor, err := coll.Find(ctx, existFilter, options.Find().SetProjection(bson.M{"guidfixed": 1}))
	if err != nil {
		return nil, fmt.Errorf("ค้นหา barcode ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingGuids := make(map[string]bool)
	for cursor.Next(ctx) {
		var doc struct {
			GuidFixed string `bson:"guidfixed"`
		}
		if err := cursor.Decode(&doc); err == nil {
			existingGuids[doc.GuidFixed] = true
		}
	}

	var deleted []string
	var notFound []string
	for _, gf := range guidFixeds {
		if existingGuids[gf] {
			deleted = append(deleted, gf)
		} else {
			notFound = append(notFound, gf)
		}
	}

	// ดึง documents ก่อนลบ เพื่อส่ง Kafka event
	kafkaSync := "skipped"
	kafkaError := ""
	var docsToDelete []interface{}
	if len(deleted) > 0 {
		delCursor, _ := coll.Find(ctx, bson.M{
			"holdingcode": holdingCode,
			"guidfixed":   bson.M{"$in": deleted},
		})
		if delCursor != nil {
			var docs []BarcodeDocument
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
			"holdingcode": holdingCode,
			"guidfixed":   bson.M{"$in": deleted},
		}
		_, err = coll.DeleteMany(ctx, deleteFilter)
		if err != nil {
			return nil, fmt.Errorf("ลบ barcode ล้มเหลว: %w", err)
		}

		// Publish Kafka event — bulk deleted (เหมือน mainapi)
		kafkaSync = "published"
		if len(docsToDelete) > 0 {
			if err := publishToKafkaBatch(ctx, TopicBarcodeBulkDeleted, docsToDelete); err != nil {
				kafkaSync = "failed"
				kafkaError = err.Error()
				logger.Warn("[MCP DeleteBarcodes] Kafka publish ล้มเหลว (แต่ MongoDB ลบแล้ว): %v", err)
			}
		}
	}

	logger.Info("[MCP DeleteBarcodes] ลบ %d รายการสำเร็จ, ไม่พบ %d รายการ (shop=%s, kafka=%s)", len(deleted), len(notFound), holdingCode, kafkaSync)

	return &DeleteBarcodesResponse{
		Success:       true,
		Message:       fmt.Sprintf("ลบ barcode %d รายการสำเร็จ", len(deleted)),
		Deleted:       deleted,
		NotFound:      notFound,
		DeletedCount:  len(deleted),
		NotFoundCount: len(notFound),
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Get Barcode Schema ====================

// GetBarcodeSchema คืนโครงสร้างข้อมูล barcode
func GetBarcodeSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  barcodeCollection,
		"description": "สินค้า/บาร์โค้ด (Product Barcode) — เอกสารหลักของสินค้า ประกอบด้วย barcode, ชื่อ, หน่วยนับ, ราคา, หมวดหมู่",
		"fields": map[string]interface{}{
			"_id":              "ObjectID — MongoDB auto-generated ID",
			"holdingcode":      "string — Holding Code (tenant isolation)",
			"guidfixed":        "string — UUID สำหรับอ้างอิงภายใน",
			"barcode":          "string (required) — รหัสบาร์โค้ด เช่น 8859100001234",
			"itemcode":         "string (required) — รหัสสินค้า เช่น SKU001",
			"names":            "array (required) — ชื่อหลายภาษา [{code:'th', name:'สินค้า A'}, {code:'en', name:'Product A'}]",
			"itemunitcode":     "string — รหัสหน่วยนับ เช่น EA, BOX",
			"itemunitnames":    "array — ชื่อหน่วยนับหลายภาษา [{code:'th', name:'ชิ้น'}]",
			"prices":           "array — ราคา [{keynumber:1, price:100.00}, {keynumber:2, price:90.00}]",
			"standvalue":       "float — ค่าตัวตั้ง (unit conversion) default=1",
			"dividevalue":      "float — ค่าตัวหาร (unit conversion) default=1",
			"ismainbarcode":    "bool — เป็น barcode หลักหรือไม่",
			"isusesubbarcodes": "bool — มี barcode อ้างอิง (reference/sub barcode) หรือไม่",
			"refbarcodes":      "array — barcode อ้างอิง [{barcode, condition, dividevalue, standvalue, qty}]",
			"imageuri":         "string — URL รูปสินค้า",
			"groupcode":        "string — รหัสกลุ่มสินค้า",
			"brandcode":        "string — รหัสยี่ห้อ",
			"categorycode":     "string — รหัสประเภท",
			"itemtype":         "int — ประเภทไอเทม (0=Stock, 1=Service, 2=Set, 3=Not Stock)",
			"materialtype":     "int — ประเภทสินค้าหลัก (0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural)",
			"taxtype":          "int — ประเภทภาษี",
			"vattype":          "int — ประเภท VAT",
			"createdby":        "string — ผู้สร้าง",
			"createdat":        "datetime — วันที่สร้าง",
			"updatedby":        "string — ผู้แก้ไขล่าสุด",
			"updatedat":        "datetime — วันที่แก้ไขล่าสุด",
			"deletedby":        "string — ผู้ลบ (soft delete)",
			"deletedat":        "datetime — วันที่ลบ (soft delete)",
		},
		"indexes": []string{
			"holdingcode + barcode (unique per shop)",
			"holdingcode + guidfixed",
			"holdingcode + itemcode",
		},
		"examples": []map[string]interface{}{
			{
				"barcode":      "8859100001234",
				"itemcode":     "SKU001",
				"names":        []map[string]string{{"code": "th", "name": "น้ำดื่ม 600ml"}, {"code": "en", "name": "Water 600ml"}},
				"itemunitcode": "EA",
				"prices":       []map[string]interface{}{{"keynumber": 1, "price": 10.00}},
			},
			{
				"barcode":      "BOX-SKU001",
				"itemcode":     "SKU001",
				"names":        []map[string]string{{"code": "th", "name": "น้ำดื่ม 600ml (ลัง)"}, {"code": "en", "name": "Water 600ml (Box)"}},
				"itemunitcode": "BOX",
				"prices":       []map[string]interface{}{{"keynumber": 1, "price": 200.00}},
				"standvalue":   24,
				"dividevalue":  1,
			},
		},
		"relatedcollections": []string{
			"units — หน่วยนับ (itemunitcode อ้างอิง unitcode)",
			"productGroups — กลุ่มสินค้า (groupcode)",
			"productBrands — ยี่ห้อ (brandcode)",
			"productCategories — ประเภท (categorycode)",
		},
	}
}

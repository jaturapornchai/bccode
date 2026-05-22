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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ==================== Reference Barcode Types ====================

// RefBarcodeEntry ข้อมูล barcode อ้างอิง (ใน array refbarcodes)
// ต้องมี names/itemunitnames เพื่อให้ frontend แสดงชื่อได้ (ตรงกับ RefProductBarcode)
type RefBarcodeEntry struct {
	GuidFixed string             `json:"guid_fixed" bson:"guid_fixed"`
	Names []BarcodeNameEntry `json:"names" bson:"names"`
	ItemUnitCode string             `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames []BarcodeNameEntry `json:"itemunitnames" bson:"itemunitnames"`
	Barcode string             `json:"barcode" bson:"barcode"`
	Condition bool               `json:"condition" bson:"condition"`
	DivideValue float64            `json:"dividevalue" bson:"dividevalue"`
	StandValue float64            `json:"standvalue" bson:"standvalue"`
	Qty float64            `json:"qty" bson:"qty"`
}

// ==================== Get Ref Barcodes ====================

// GetRefBarcodesResponse ผลลัพธ์จากการดึง reference barcodes
type GetRefBarcodesResponse struct {
	ShopID string             `json:"shopid"`
	ItemCode string             `json:"itemcode"`
	ProductName string             `json:"product_name"`
	Barcodes []RefBarcodeDetail `json:"barcodes"`
	UnitChain []UnitChainEntry   `json:"unit_chain"`
	Count int                `json:"count"`
	GeneratedAt time.Time          `json:"generated_at"`
}

// RefBarcodeDetail ข้อมูล barcode พร้อม reference
type RefBarcodeDetail struct {
	GuidFixed string              `json:"guid_fixed"`
	Barcode string              `json:"barcode"`
	ProductName string              `json:"product_name"`
	Names []BarcodeNameEntry  `json:"names"`
	ItemUnitCode string              `json:"item_unit_code"`
	UnitName string              `json:"unit_name"`
	ItemUnitNames []BarcodeNameEntry  `json:"itemunitnames"`
	StandValue float64             `json:"standvalue"`
	DivideValue float64             `json:"dividevalue"`
	IsMainBarcode bool                `json:"is_main_barcode"`
	IsUseSubBarcodes bool                `json:"isusesubbarcodes"`
	RefBarcodes []RefBarcodeEntry   `json:"refbarcodes"`
	Prices []BarcodePriceEntry `json:"prices"`
}

// UnitChainEntry แสดงสายหน่วยนับ
type UnitChainEntry struct {
	Barcode string  `json:"barcode"`
	ItemUnitCode string  `json:"item_unit_code"`
	UnitName string  `json:"unit_name"`
	StandValue float64 `json:"standvalue"`
	DivideValue float64 `json:"dividevalue"`
	RefBarcode string  `json:"ref_barcode,omitempty"`
	BaseQty float64 `json:"base_qty"`
}

// getThaiName ดึงชื่อภาษาไทย (ถ้าไม่มี → ภาษาแรกที่มี)
func getThaiName(names []BarcodeNameEntry) string {
	for _, n := range names {
		if n.Code == "th" && n.Name != "" {
			return n.Name
		}
	}
	for _, n := range names {
		if n.Name != "" {
			return n.Name
		}
	}
	return ""
}

// GetRefBarcodes ดึง reference barcodes ของสินค้า (itemcode)
func GetRefBarcodes(ctx context.Context, shopID, itemCode, barcode string) (*GetRefBarcodesResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if itemCode == "" && barcode == "" {
		return nil, fmt.Errorf("ต้องระบุ itemcode หรือ barcode อย่างน้อย 1 อย่าง")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	notDeletedFilter := []bson.M{
		{"deleted_by": bson.M{"$exists": false}},
		{"deleted_by": ""},
	}

	// ถ้าระบุ barcode แต่ไม่ระบุ itemcode → หา itemcode จาก barcode
	if itemCode == "" && barcode != "" {
		var doc struct {
			ItemCode string `bson:"itemcode"`
		}
		err := coll.FindOne(ctx, bson.M{
			"shopid":  shopID,
			"barcode": barcode,
			"$or":     notDeletedFilter,
		}).Decode(&doc)
		if err != nil {
			return nil, fmt.Errorf("ไม่พบ barcode '%s'", barcode)
		}
		itemCode = doc.ItemCode
	}

	// ดึง barcodes ทั้งหมดที่มี itemcode เดียวกัน
	filter := bson.M{
		"shopid":   shopID,
		"itemcode": itemCode,
		"$or":      notDeletedFilter,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "standvalue", Value: 1}}).
		SetProjection(bson.M{
			"guid_fixed": 1, "barcode": 1, "names": 1,
			"item_unit_code": 1, "itemunitnames": 1,
			"standvalue": 1, "dividevalue": 1,
			"is_main_barcode": 1, "isusesubbarcodes": 1,
			"refbarcodes": 1, "prices": 1,
		})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var barcodes []RefBarcodeDetail
	for cursor.Next(ctx) {
		var doc struct {
			GuidFixed string              `bson:"guid_fixed"`
			Barcode string              `bson:"barcode"`
			Names []BarcodeNameEntry  `bson:"names"`
			ItemUnitCode string              `bson:"item_unit_code"`
			ItemUnitNames []BarcodeNameEntry  `bson:"itemunitnames"`
			StandValue float64             `bson:"standvalue"`
			DivideValue float64             `bson:"dividevalue"`
			IsMainBarcode bool                `bson:"is_main_barcode"`
			IsUseSubBarcodes bool                `bson:"isusesubbarcodes"`
			RefBarcodes []RefBarcodeEntry   `bson:"refbarcodes"`
			Prices []BarcodePriceEntry `bson:"prices"`
		}
		if err := cursor.Decode(&doc); err != nil {
			logger.Warn("[MCP GetRefBarcodes] decode error: %v", err)
			continue
		}
		if doc.RefBarcodes == nil {
			doc.RefBarcodes = []RefBarcodeEntry{}
		}
		if doc.Names == nil {
			doc.Names = []BarcodeNameEntry{}
		}
		if doc.ItemUnitNames == nil {
			doc.ItemUnitNames = []BarcodeNameEntry{}
		}
		if doc.Prices == nil {
			doc.Prices = []BarcodePriceEntry{}
		}
		barcodes = append(barcodes, RefBarcodeDetail{
			GuidFixed:        doc.GuidFixed,
			Barcode:          doc.Barcode,
			ProductName:      getThaiName(doc.Names),
			Names:            doc.Names,
			ItemUnitCode:     doc.ItemUnitCode,
			UnitName:         getThaiName(doc.ItemUnitNames),
			ItemUnitNames:    doc.ItemUnitNames,
			StandValue:       doc.StandValue,
			DivideValue:      doc.DivideValue,
			IsMainBarcode:    doc.IsMainBarcode,
			IsUseSubBarcodes: doc.IsUseSubBarcodes,
			RefBarcodes:      doc.RefBarcodes,
			Prices:           doc.Prices,
		})
	}

	if barcodes == nil {
		barcodes = []RefBarcodeDetail{}
	}

	// สร้าง unit chain
	unitChain := buildUnitChain(barcodes)

	// ดึงชื่อสินค้าจาก barcode แรก (ทุก barcode ใน itemcode เดียวกันชื่อเดียวกัน)
	productName := ""
	if len(barcodes) > 0 {
		productName = barcodes[0].ProductName
	}

	logger.Info("[MCP GetRefBarcodes] shopID=%s, itemcode=%s, product=%s, count=%d", shopID, itemCode, productName, len(barcodes))

	return &GetRefBarcodesResponse{
		ShopID:      shopID,
		ItemCode:    itemCode,
		ProductName: productName,
		Barcodes:    barcodes,
		UnitChain:   unitChain,
		Count:       len(barcodes),
		GeneratedAt: time.Now(),
	}, nil
}

// buildUnitChain สร้างสายหน่วยนับจาก barcodes
func buildUnitChain(barcodes []RefBarcodeDetail) []UnitChainEntry {
	var chain []UnitChainEntry
	for _, b := range barcodes {
		refBarcode := ""
		if len(b.RefBarcodes) > 0 {
			refBarcode = b.RefBarcodes[0].Barcode
		}

		// คำนวณ base qty (จำนวนหน่วยฐาน)
		baseQty := b.StandValue
		if b.DivideValue > 0 {
			baseQty = b.StandValue / b.DivideValue
		}

		chain = append(chain, UnitChainEntry{
			Barcode:      b.Barcode,
			ItemUnitCode: b.ItemUnitCode,
			UnitName:     b.UnitName,
			StandValue:   b.StandValue,
			DivideValue:  b.DivideValue,
			RefBarcode:   refBarcode,
			BaseQty:      baseQty,
		})
	}

	if chain == nil {
		chain = []UnitChainEntry{}
	}
	return chain
}

// ==================== Set Ref Barcode ====================

// SetRefBarcodeResponse ผลลัพธ์จากการตั้ง reference barcode
type SetRefBarcodeResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Barcode string    `json:"barcode"`
	RefBarcode string    `json:"ref_barcode"`
	KafkaSync string    `json:"kafka_sync"`
	KafkaError string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

// SetRefBarcode ตั้ง reference barcode ให้ barcode ที่ระบุ
func SetRefBarcode(ctx context.Context, shopID, barcode, refBarcode string, qty, standValue, divideValue float64, condition bool) (*SetRefBarcodeResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}
	if refBarcode == "" {
		return nil, fmt.Errorf("ref_barcode is required")
	}
	if barcode == refBarcode {
		return nil, fmt.Errorf("barcode ไม่สามารถอ้างอิงตัวเองได้")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	notDeletedFilter := []bson.M{
		{"deleted_by": bson.M{"$exists": false}},
		{"deleted_by": ""},
	}

	// ตรวจสอบ barcode ต้นทาง
	var sourceDoc BarcodeDocument
	err := coll.FindOne(ctx, bson.M{
		"shopid":  shopID,
		"barcode": barcode,
		"$or":     notDeletedFilter,
	}).Decode(&sourceDoc)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบ barcode '%s'", barcode)
	}

	// ตรวจสอบ ref barcode ปลายทาง
	var refDoc BarcodeDocument
	err = coll.FindOne(ctx, bson.M{
		"shopid":  shopID,
		"barcode": refBarcode,
		"$or":     notDeletedFilter,
	}).Decode(&refDoc)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบ ref_barcode '%s'", refBarcode)
	}

	// ตรวจสอบ itemcode เดียวกัน
	if sourceDoc.ItemCode != refDoc.ItemCode {
		return nil, fmt.Errorf("barcode '%s' (itemcode=%s) และ ref_barcode '%s' (itemcode=%s) ต้องเป็น itemcode เดียวกัน",
			barcode, sourceDoc.ItemCode, refBarcode, refDoc.ItemCode)
	}

	// ตรวจสอบ circular reference
	if err := checkCircularRef(ctx, coll, shopID, barcode, refBarcode); err != nil {
		return nil, err
	}

	// Default values
	if standValue == 0 {
		standValue = sourceDoc.StandValue
	}
	if divideValue == 0 {
		divideValue = sourceDoc.DivideValue
	}

	// สร้าง ref entry — populate names จาก ref barcode document เพื่อให้ frontend แสดงชื่อได้
	refEntry := RefBarcodeEntry{
		GuidFixed:     refDoc.GuidFixed,
		Names:         refDoc.Names,
		ItemUnitCode:  refDoc.ItemUnitCode,
		ItemUnitNames: refDoc.ItemUnitNames,
		Barcode:       refBarcode,
		Condition:     condition,
		DivideValue:   divideValue,
		StandValue:    standValue,
		Qty:           qty,
	}

	// Update barcode — set refbarcodes + isusesubbarcodes
	now := time.Now()
	_, err = coll.UpdateOne(ctx, bson.M{
		"shopid":  shopID,
		"barcode": barcode,
		"$or":     notDeletedFilter,
	}, bson.M{
		"$set": bson.M{
			"refbarcodes":      []RefBarcodeEntry{refEntry},
			"isusesubbarcodes": true,
			"updatedby":        "mcp-tool",
			"updated_at":        now,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("อัปเดต refbarcodes ล้มเหลว: %w", err)
	}

	// Publish Kafka event
	kafkaSync := "published"
	kafkaError := ""
	var updated BarcodeDocument
	coll.FindOne(ctx, bson.M{"_id": sourceDoc.ID}).Decode(&updated)
	if err := publishToKafka(ctx, TopicBarcodeUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP SetRefBarcode] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP SetRefBarcode] barcode=%s → ref=%s สำเร็จ (shop=%s)", barcode, refBarcode, shopID)

	return &SetRefBarcodeResponse{
		Success:     true,
		Message:     fmt.Sprintf("ตั้ง reference barcode '%s' → '%s' สำเร็จ", barcode, refBarcode),
		Barcode:     barcode,
		RefBarcode:  refBarcode,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// checkCircularRef ตรวจสอบ circular reference
func checkCircularRef(ctx context.Context, coll *mongo.Collection, shopID, sourceBarcode, targetRefBarcode string) error {
	visited := map[string]bool{sourceBarcode: true}
	current := targetRefBarcode

	notDeletedFilter := []bson.M{
		{"deleted_by": bson.M{"$exists": false}},
		{"deleted_by": ""},
	}

	for i := 0; i < 10; i++ { // max 10 levels deep
		if visited[current] {
			return fmt.Errorf("circular reference detected: '%s' → '%s' จะทำให้เกิด loop", sourceBarcode, targetRefBarcode)
		}
		visited[current] = true

		var doc struct {
			RefBarcodes []RefBarcodeEntry `bson:"refbarcodes"`
		}
		err := coll.FindOne(ctx, bson.M{
			"shopid":  shopID,
			"barcode": current,
			"$or":     notDeletedFilter,
		}).Decode(&doc)
		if err != nil {
			break // ไม่พบ → ไม่มี loop
		}
		if len(doc.RefBarcodes) == 0 {
			break // ไม่มี ref → ไม่มี loop
		}
		current = doc.RefBarcodes[0].Barcode
	}
	return nil
}

// ==================== Create Multi-Unit Barcode ====================

// MultiUnitItem หน่วยนับที่ต้องการสร้าง
type MultiUnitItem struct {
	Barcode string              `json:"barcode"`
	ItemUnitCode string              `json:"item_unit_code"`
	ItemUnitNames []BarcodeNameEntry  `json:"itemunitnames"`
	Names []BarcodeNameEntry  `json:"names"`
	StandValue float64             `json:"standvalue"`
	DivideValue float64             `json:"dividevalue"`
	Prices []BarcodePriceEntry `json:"prices"`
}

// CreateMultiUnitBarcodeResponse ผลลัพธ์จากการสร้าง multi-unit barcode
type CreateMultiUnitBarcodeResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	ItemCode string            `json:"itemcode"`
	Created []BarcodeDocument `json:"created"`
	RefLinks []string          `json:"ref_links"`
	KafkaSync string            `json:"kafka_sync"`
	KafkaError string            `json:"kafka_error,omitempty"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// CreateMultiUnitBarcode สร้างสินค้าพร้อมหลายหน่วยนับ + ตั้ง reference อัตโนมัติ
func CreateMultiUnitBarcode(ctx context.Context, shopID, itemCode, namesJSON, unitsJSON string) (*CreateMultiUnitBarcodeResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if itemCode == "" {
		return nil, fmt.Errorf("itemcode is required")
	}
	if unitsJSON == "" {
		return nil, fmt.Errorf("units is required")
	}

	// Parse product names (ใช้ร่วมกันทุกหน่วย ถ้าหน่วยนั้นไม่ได้ระบุ names)
	var defaultNames []BarcodeNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &defaultNames); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}

	// Parse units
	var units []MultiUnitItem
	if err := json.Unmarshal([]byte(unitsJSON), &units); err != nil {
		return nil, fmt.Errorf("units JSON ไม่ถูกต้อง: %w", err)
	}
	if len(units) < 2 {
		return nil, fmt.Errorf("ต้องระบุอย่างน้อย 2 หน่วยนับ (เช่น ชิ้น + ลัง)")
	}
	if len(units) > 10 {
		return nil, fmt.Errorf("สร้างได้สูงสุด 10 หน่วยนับต่อ itemcode")
	}

	// Validate units
	for i, u := range units {
		if u.Barcode == "" {
			return nil, fmt.Errorf("units[%d].barcode is required", i)
		}
		if u.ItemUnitCode == "" {
			return nil, fmt.Errorf("units[%d].itemunitcode is required", i)
		}
		if u.StandValue <= 0 {
			return nil, fmt.Errorf("units[%d].standvalue ต้องมากกว่า 0", i)
		}
		if u.DivideValue <= 0 {
			units[i].DivideValue = 1
		}
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(barcodeCollection)

	// ตรวจสอบ barcode ซ้ำ
	var barcodeValues []string
	for _, u := range units {
		barcodeValues = append(barcodeValues, u.Barcode)
	}
	existFilter := bson.M{
		"shopid":  shopID,
		"barcode": bson.M{"$in": barcodeValues},
		"$or": []bson.M{
			{"deleted_by": bson.M{"$exists": false}},
			{"deleted_by": ""},
		},
	}
	existCount, err := coll.CountDocuments(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ barcode ซ้ำล้มเหลว: %w", err)
	}
	if existCount > 0 {
		return nil, fmt.Errorf("มี barcode ที่ซ้ำอยู่แล้วใน shop นี้ (พบ %d รายการ)", existCount)
	}

	now := time.Now()
	var docsToInsert []interface{}
	var createdDocs []BarcodeDocument
	var refLinks []string

	for i, u := range units {
		names := u.Names
		if len(names) == 0 {
			names = defaultNames
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("units[%d] ต้องมี names หรือระบุ names ระดับ product", i)
		}

		isMain := u.StandValue == 1 && u.DivideValue == 1

		// Auto-lookup: ถ้าไม่ได้ระบุ itemunitnames แต่ระบุ itemunitcode → ดึงจาก units collection
		unitNames := u.ItemUnitNames
		if len(unitNames) == 0 && u.ItemUnitCode != "" {
			if looked := lookupUnitNames(ctx, shopID, u.ItemUnitCode); looked != nil {
				unitNames = looked
			}
		}

		doc := BarcodeDocument{
			ShopID:        shopID,
			GuidFixed:     uuid.New().String(),
			Barcode:       u.Barcode,
			ItemCode:      itemCode,
			Names:         names,
			ItemUnitCode:  u.ItemUnitCode,
			ItemUnitNames: unitNames,
			Prices:        u.Prices,
			StandValue:    u.StandValue,
			DivideValue:   u.DivideValue,
			IsMainBarcode: isMain,
			CreatedBy:     "mcp-tool",
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		docsToInsert = append(docsToInsert, doc)
		createdDocs = append(createdDocs, doc)
	}

	// Insert ทั้งหมด
	_, err = coll.InsertMany(ctx, docsToInsert)
	if err != nil {
		return nil, fmt.Errorf("สร้าง barcodes ล้มเหลว: %w", err)
	}

	// หาหน่วยฐาน (standvalue=1, dividevalue=1) หรือ standvalue น้อยที่สุด
	var baseBarcode string
	for _, u := range units {
		if u.StandValue == 1 && u.DivideValue == 1 {
			baseBarcode = u.Barcode
			break
		}
	}
	if baseBarcode == "" {
		minStand := units[0].StandValue
		baseBarcode = units[0].Barcode
		for _, u := range units[1:] {
			if u.StandValue < minStand {
				minStand = u.StandValue
				baseBarcode = u.Barcode
			}
		}
	}

	// อัปเดต refbarcodes สำหรับหน่วยอื่นๆ → อ้างหน่วยฐาน
	notDeletedFilter := []bson.M{
		{"deleted_by": bson.M{"$exists": false}},
		{"deleted_by": ""},
	}

	// หา base barcode document เพื่อ populate names ใน refEntry
	var baseDoc BarcodeDocument
	for _, d := range createdDocs {
		if d.Barcode == baseBarcode {
			baseDoc = d
			break
		}
	}

	for _, u := range units {
		if u.Barcode == baseBarcode {
			continue
		}

		refEntry := RefBarcodeEntry{
			GuidFixed:     baseDoc.GuidFixed,
			Names:         baseDoc.Names,
			ItemUnitCode:  baseDoc.ItemUnitCode,
			ItemUnitNames: baseDoc.ItemUnitNames,
			Barcode:       baseBarcode,
			Condition:     false,
			DivideValue:   u.DivideValue,
			StandValue:    u.StandValue,
			Qty:           0,
		}

		_, updateErr := coll.UpdateOne(ctx, bson.M{
			"shopid":  shopID,
			"barcode": u.Barcode,
			"$or":     notDeletedFilter,
		}, bson.M{
			"$set": bson.M{
				"refbarcodes":      []RefBarcodeEntry{refEntry},
				"isusesubbarcodes": true,
			},
		})
		if updateErr != nil {
			logger.Warn("[MCP CreateMultiUnit] set ref for %s ล้มเหลว: %v", u.Barcode, updateErr)
		} else {
			refLinks = append(refLinks, fmt.Sprintf("%s → %s", u.Barcode, baseBarcode))
		}
	}

	// Publish Kafka events
	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafkaBatch(ctx, TopicBarcodeBulkCreated, docsToInsert); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateMultiUnit] Kafka publish ล้มเหลว: %v", err)
	}

	if refLinks == nil {
		refLinks = []string{}
	}

	logger.Info("[MCP CreateMultiUnit] สร้าง %d units สำหรับ itemcode=%s (shop=%s, refs=%d)",
		len(createdDocs), itemCode, shopID, len(refLinks))

	return &CreateMultiUnitBarcodeResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้าง %d หน่วยนับ สำหรับ itemcode '%s' สำเร็จ", len(createdDocs), itemCode),
		ItemCode:    itemCode,
		Created:     createdDocs,
		RefLinks:    refLinks,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

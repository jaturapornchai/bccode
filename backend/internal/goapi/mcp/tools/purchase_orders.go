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

const purchaseOrderCollection = "transactionPurchaseOrder"

// ==================== Kafka Topics ====================

const (
	TopicPurchaseOrderCreated = "when-purchase-order-created"
	TopicPurchaseOrderUpdated = "when-purchase-order-updated"
	TopicPurchaseOrderDeleted = "when-purchase-order-deleted"
)

// ==================== Purchase Order Types ====================

// PurchaseOrderNameEntry ชื่อหลายภาษา
type PurchaseOrderNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// PurchaseOrderDetail รายละเอียดสินค้าในใบสั่งซื้อ
type PurchaseOrderDetail struct {
	LineNumber          int                      `json:"linenumber" bson:"linenumber"`
	DocDatetime         time.Time                `json:"docdatetime" bson:"docdatetime"`
	Barcode             string                   `json:"barcode" bson:"barcode"`
	ItemCode            string                   `json:"itemcode" bson:"itemcode"`
	ItemNames           []PurchaseOrderNameEntry `json:"itemnames" bson:"itemnames"`
	UnitCode            string                   `json:"unitcode" bson:"unitcode"`
	UnitNames           []PurchaseOrderNameEntry `json:"unitnames" bson:"unitnames"`
	ItemType            int8                     `json:"itemtype" bson:"itemtype"`
	ItemGuid            string                   `json:"itemguid" bson:"itemguid"`
	Qty                 float64                  `json:"qty" bson:"qty"`
	Price               float64                  `json:"price" bson:"price"`
	Discount            string                   `json:"discount" bson:"discount"`
	DiscountAmount      float64                  `json:"discountamount" bson:"discountamount"`
	SumAmount           float64                  `json:"sumamount" bson:"sumamount"`
	SumAmountExcludeVat float64                  `json:"sumamountexcludevat" bson:"sumamountexcludevat"`
	TotalValueVat       float64                  `json:"totalvaluevat" bson:"totalvaluevat"`
	PriceExcludeVat     float64                  `json:"priceexcludevat" bson:"priceexcludevat"`
	VatType             int8                     `json:"vattype" bson:"vattype"`
	StandValue          float64                  `json:"standvalue" bson:"standvalue"`
	DivideValue         float64                  `json:"dividevalue" bson:"dividevalue"`
	WhCode              string                   `json:"whcode" bson:"whcode"`
	WhNames             []PurchaseOrderNameEntry `json:"whnames" bson:"whnames"`
	LocationCode        string                   `json:"locationcode" bson:"locationcode"`
	LocationNames       []PurchaseOrderNameEntry `json:"locationnames" bson:"locationnames"`
	Remark              string                   `json:"remark" bson:"remark"`
	Description         string                   `json:"description" bson:"description"`
	// Document Currency fields
	PriceDoc               float64 `json:"pricedoc" bson:"pricedoc,omitempty"`
	SumAmountDoc           float64 `json:"sumamountdoc" bson:"sumamountdoc,omitempty"`
	DiscountAmountDoc      float64 `json:"discountamountdoc" bson:"discountamountdoc,omitempty"`
	PriceExcludeVatDoc     float64 `json:"priceexcludevatdoc" bson:"priceexcludevatdoc,omitempty"`
	SumAmountExcludeVatDoc float64 `json:"sumamountexcludevatdoc" bson:"sumamountexcludevatdoc,omitempty"`
	TotalValueVatDoc       float64 `json:"totalvaluevatdoc" bson:"totalvaluevatdoc,omitempty"`
}

// PurchaseOrderBranch สาขา
type PurchaseOrderBranch struct {
	GuidFixed string                   `json:"guidfixed" bson:"guidfixed"`
	Code      string                   `json:"code" bson:"code"`
	Names     []PurchaseOrderNameEntry `json:"names" bson:"names"`
}

// PurchaseOrderWHTEntry ภาษีหัก ณ ที่จ่าย
type PurchaseOrderWHTEntry struct {
	Description string  `json:"description" bson:"description"`
	Rate        float64 `json:"rate" bson:"rate"`
	TaxBase     float64 `json:"taxbase" bson:"taxbase"`
	Amount      float64 `json:"amount" bson:"amount"`
	Note        string  `json:"note,omitempty" bson:"note,omitempty"`
}

// PurchaseOrderDocument เอกสารใบสั่งซื้อใน MongoDB
type PurchaseOrderDocument struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	GuidFixed   string             `json:"guidfixed" bson:"guidfixed"`
	// Header
	DocNo       string    `json:"docno" bson:"docno"`
	DocDatetime time.Time `json:"docdatetime" bson:"docdatetime"`
	TransFlag   int       `json:"transflag" bson:"transflag"`
	DocType     int8      `json:"doctype" bson:"doctype"`
	// Creditor
	CustCode  string                   `json:"custcode" bson:"custcode"`
	CustNames []PurchaseOrderNameEntry `json:"custnames" bson:"custnames"`
	// Financial
	Description    string  `json:"description" bson:"description"`
	DiscountWord   string  `json:"discountword" bson:"discountword"`
	TotalDiscount  float64 `json:"totaldiscount" bson:"totaldiscount"`
	TotalValue     float64 `json:"totalvalue" bson:"totalvalue"`
	TotalBeforeVat float64 `json:"totalbeforevat" bson:"totalbeforevat"`
	TotalAfterVat  float64 `json:"totalaftervat" bson:"totalaftervat"`
	TotalVatValue  float64 `json:"totalvatvalue" bson:"totalvatvalue"`
	TotalAmount    float64 `json:"totalamount" bson:"totalamount"`
	TotalExceptVat float64 `json:"totalexceptvat" bson:"totalexceptvat"`
	VatType        int8    `json:"vattype" bson:"vattype"`
	VatRate        float64 `json:"vatrate" bson:"vatrate"`
	TotalQty       float64 `json:"totalqty" bson:"totalqty"`
	RoundAmount    float64 `json:"roundamount" bson:"roundamount"`
	IsManualAmount bool    `json:"ismanualamount" bson:"ismanualamount"`
	IsCancel       bool    `json:"iscancel" bson:"iscancel"`
	Status         int8    `json:"status" bson:"status"`
	// Branch
	Branch PurchaseOrderBranch `json:"branch" bson:"branch"`
	// Multi-Currency
	Currency          string  `json:"currency" bson:"currency,omitempty"`
	CurrencySymbol    string  `json:"currencysymbol" bson:"currencysymbol,omitempty"`
	DocCurrency       string  `json:"doccurrency" bson:"doccurrency,omitempty"`
	DocCurrencySymbol string  `json:"doccurrencysymbol" bson:"doccurrencysymbol,omitempty"`
	ExchangeRate      float64 `json:"exchangerate" bson:"exchangerate"`
	// Document Currency Totals
	TotalValueDoc     float64 `json:"totalvaluedoc" bson:"totalvaluedoc,omitempty"`
	TotalDiscountDoc  float64 `json:"totaldiscountdoc" bson:"totaldiscountdoc,omitempty"`
	TotalVatValueDoc  float64 `json:"totalvatvaluedoc" bson:"totalvatvaluedoc,omitempty"`
	TotalBeforeVatDoc float64 `json:"totalbeforevatdoc" bson:"totalbeforevatdoc,omitempty"`
	TotalAfterVatDoc  float64 `json:"totalaftervatdoc" bson:"totalaftervatdoc,omitempty"`
	TotalAmountDoc    float64 `json:"totalamountdoc" bson:"totalamountdoc,omitempty"`
	// WHT
	WHTEntries []PurchaseOrderWHTEntry `json:"whtentries,omitempty" bson:"whtentries"`
	// Credit Terms
	CreditDays int    `json:"creditdays,omitempty" bson:"creditdays"`
	DueDate    string `json:"duedate,omitempty" bson:"duedate"`
	// Purchase Type (for PO Approval)
	PurchaseTypeCode  string                   `json:"purchasetypecode" bson:"purchasetypecode"`
	PurchaseTypeNames []PurchaseOrderNameEntry `json:"purchasetypenames" bson:"purchasetypenames"`
	// Creator/Updater
	CreatorCode string    `json:"creatorcode" bson:"creatorcode"`
	CreatorName string    `json:"creatorname" bson:"creatorname"`
	CreatedAt   time.Time `json:"createdat" bson:"createdat"`
	UpdaterCode string    `json:"modifiercode" bson:"updatercode"`
	UpdaterName string    `json:"modifiername" bson:"updatername"`
	UpdatedAt   time.Time `json:"updatedat" bson:"updatedat"`
	DeletedAt   time.Time `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	// Sale
	SaleCode string `json:"salecode" bson:"salecode"`
	SaleName string `json:"salename" bson:"salename"`
	// Details
	Details []PurchaseOrderDetail `json:"details" bson:"details"`
}

// ==================== List Purchase Orders ====================

type ListPurchaseOrdersResponse struct {
	PurchaseOrders []PurchaseOrderDocument `json:"purchaseorders"`
	Count          int                     `json:"count"`
	Keyword        string                  `json:"keyword,omitempty"`
	GeneratedAt    time.Time               `json:"generatedat"`
}

func ListPurchaseOrders(ctx context.Context, holdingCode, keyword string, limit int) (*ListPurchaseOrdersResponse, error) {
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

	filter := bson.M{
		"holdingcode": holdingCode,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	if keyword != "" {
		keywordFilter := bson.M{
			"$or": []bson.M{
				{"docno": bson.M{"$regex": keyword, "$options": "i"}},
				{"custcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"custnames.name": bson.M{"$regex": keyword, "$options": "i"}},
				{"description": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListPurchaseOrders] holdingCode=%s, keyword=%s, limit=%d", holdingCode, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(purchaseOrderCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "docdatetime", Value: -1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []PurchaseOrderDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}
	if docs == nil {
		docs = []PurchaseOrderDocument{}
	}

	return &ListPurchaseOrdersResponse{
		PurchaseOrders: docs,
		Count:          len(docs),
		Keyword:        keyword,
		GeneratedAt:    time.Now(),
	}, nil
}

// ==================== Create Purchase Order ====================

type CreatePurchaseOrderResponse struct {
	Success       bool                  `json:"success"`
	Message       string                `json:"message"`
	PurchaseOrder PurchaseOrderDocument `json:"purchaseorder"`
	KafkaSync     string                `json:"kafkasync"`
	KafkaError    string                `json:"kafkaerror,omitempty"`
	GeneratedAt   time.Time             `json:"generatedat"`
}

func CreatePurchaseOrder(ctx context.Context, holdingCode, docno, custcode, custnamesJSON, detailsJSON, description string, transflag int, vattype int8, vatrate float64, totalamount float64) (*CreatePurchaseOrderResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if docno == "" {
		return nil, fmt.Errorf("docno is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(purchaseOrderCollection)

	// ตรวจสอบ docno ซ้ำ
	existFilter := bson.M{
		"holdingcode": holdingCode,
		"docno":       docno,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}
	count, err := coll.CountDocuments(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ docno ซ้ำล้มเหลว: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("purchase order docno '%s' มีอยู่แล้วใน shop นี้", docno)
	}

	// Parse custnames JSON
	var custnames []PurchaseOrderNameEntry
	if custnamesJSON != "" {
		if err := json.Unmarshal([]byte(custnamesJSON), &custnames); err != nil {
			return nil, fmt.Errorf("custnames JSON ไม่ถูกต้อง: %w", err)
		}
	}

	// Parse details JSON
	var details []PurchaseOrderDetail
	if detailsJSON != "" {
		if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
			return nil, fmt.Errorf("details JSON ไม่ถูกต้อง: %w", err)
		}
	}

	now := time.Now()
	doc := PurchaseOrderDocument{
		HoldingCode: holdingCode,
		GuidFixed:   uuid.New().String(),
		DocNo:       docno,
		DocDatetime: now,
		TransFlag:   transflag,
		CustCode:    custcode,
		CustNames:   custnames,
		Description: description,
		VatType:     vattype,
		VatRate:     vatrate,
		TotalAmount: totalamount,
		Details:     details,
		CreatorCode: "mcp-tool",
		CreatorName: "MCP Tool",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้างใบสั่งซื้อล้มเหลว: %w", err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicPurchaseOrderCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreatePurchaseOrder] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreatePurchaseOrder] สร้าง docno=%s สำเร็จ (shop=%s, kafka=%s)", docno, holdingCode, kafkaSync)

	return &CreatePurchaseOrderResponse{
		Success:       true,
		Message:       fmt.Sprintf("สร้างใบสั่งซื้อ '%s' สำเร็จ", docno),
		PurchaseOrder: doc,
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Update Purchase Order ====================

type UpdatePurchaseOrderResponse struct {
	Success       bool                  `json:"success"`
	Message       string                `json:"message"`
	PurchaseOrder PurchaseOrderDocument `json:"purchaseorder"`
	KafkaSync     string                `json:"kafkasync"`
	KafkaError    string                `json:"kafkaerror,omitempty"`
	GeneratedAt   time.Time             `json:"generatedat"`
}

func UpdatePurchaseOrder(ctx context.Context, holdingCode, guidfixed, custnamesJSON, detailsJSON, description string, totalamount float64, status int8) (*UpdatePurchaseOrderResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if guidfixed == "" {
		return nil, fmt.Errorf("guidfixed is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(purchaseOrderCollection)

	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guidfixed,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	var existing PurchaseOrderDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบใบสั่งซื้อ guidfixed='%s'", guidfixed)
	}

	updateFields := bson.M{
		"updatedat":   time.Now(),
		"updatercode": "mcp-tool",
		"updatername": "MCP Tool",
	}

	if custnamesJSON != "" {
		var custnames []PurchaseOrderNameEntry
		if err := json.Unmarshal([]byte(custnamesJSON), &custnames); err != nil {
			return nil, fmt.Errorf("custnames JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["custnames"] = custnames
	}
	if detailsJSON != "" {
		var details []PurchaseOrderDetail
		if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
			return nil, fmt.Errorf("details JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["details"] = details
	}
	if description != "" {
		updateFields["description"] = description
	}
	if totalamount > 0 {
		updateFields["totalamount"] = totalamount
	}
	if status > 0 {
		updateFields["status"] = status
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดตใบสั่งซื้อล้มเหลว: %w", err)
	}

	var updated PurchaseOrderDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicPurchaseOrderUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdatePurchaseOrder] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP UpdatePurchaseOrder] อัปเดต guidfixed=%s สำเร็จ (shop=%s)", guidfixed, holdingCode)

	return &UpdatePurchaseOrderResponse{
		Success:       true,
		Message:       fmt.Sprintf("อัปเดตใบสั่งซื้อ '%s' สำเร็จ", existing.DocNo),
		PurchaseOrder: updated,
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Delete Purchase Order (Soft Delete) ====================

type DeletePurchaseOrderResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	GuidFixed   string    `json:"guidfixed"`
	DocNo       string    `json:"docno"`
	KafkaSync   string    `json:"kafkasync"`
	KafkaError  string    `json:"kafkaerror,omitempty"`
	GeneratedAt time.Time `json:"generatedat"`
}

func DeletePurchaseOrder(ctx context.Context, holdingCode, guidfixed string) (*DeletePurchaseOrderResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if guidfixed == "" {
		return nil, fmt.Errorf("guidfixed is required")
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(purchaseOrderCollection)

	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guidfixed,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	var existing PurchaseOrderDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบใบสั่งซื้อ guidfixed='%s'", guidfixed)
	}

	// Soft delete
	now := time.Now()
	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
		"deletedat":   now,
		"updatercode": "mcp-tool",
		"updatername": "MCP Tool",
		"updatedat":   now,
	}})
	if err != nil {
		return nil, fmt.Errorf("ลบใบสั่งซื้อล้มเหลว: %w", err)
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicPurchaseOrderDeleted, existing); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP DeletePurchaseOrder] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP DeletePurchaseOrder] ลบ docno=%s สำเร็จ (shop=%s)", existing.DocNo, holdingCode)

	return &DeletePurchaseOrderResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบใบสั่งซื้อ '%s' สำเร็จ", existing.DocNo),
		GuidFixed:   guidfixed,
		DocNo:       existing.DocNo,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Get Purchase Order Schema ====================

func GetPurchaseOrderSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  purchaseOrderCollection,
		"description": "ใบสั่งซื้อ (Purchase Order) — เอกสารสั่งซื้อสินค้าจากเจ้าหนี้/ซัพพลายเออร์",
		"fields": map[string]interface{}{
			"_id":              "ObjectID — MongoDB auto-generated ID",
			"holdingcode":      "string — Holding Code (tenant isolation)",
			"guidfixed":        "string — UUID สำหรับอ้างอิงภายใน",
			"docno":            "string (required) — เลขที่เอกสาร เช่น PO-2026-0001",
			"docdatetime":      "datetime — วันที่เอกสาร",
			"transflag":        "int — ประเภท transaction (6=ใบสั่งซื้อ)",
			"doctype":          "int8 — ประเภทเอกสาร",
			"custcode":         "string — รหัสเจ้าหนี้/ผู้ขาย",
			"custnames":        "array — ชื่อเจ้าหนี้หลายภาษา [{code:'th', name:'บริษัท ABC'}]",
			"description":      "string — คำอธิบาย/หมายเหตุ",
			"vattype":          "int8 — ประเภท VAT (0=ไม่มี, 1=รวม, 2=แยก)",
			"vatrate":          "float64 — อัตรา VAT %",
			"totalvalue":       "float64 — มูลค่ารวมสินค้า",
			"totaldiscount":    "float64 — ส่วนลดท้ายบิล",
			"totalbeforevat":   "float64 — มูลค่าก่อน VAT",
			"totalaftervat":    "float64 — มูลค่าหลัง VAT",
			"totalvatvalue":    "float64 — มูลค่า VAT",
			"totalamount":      "float64 — มูลค่ารวมทั้งหมด",
			"totalqty":         "float64 — จำนวนรวม",
			"status":           "int8 — สถานะ (0=ร่าง, 1=รออนุมัติ, 2=อนุมัติ, 3=ปฏิเสธ)",
			"iscancel":         "bool — ยกเลิกแล้วหรือไม่",
			"branch":           "object — สาขา {code, names}",
			"currency":         "string — สกุลเงินหลัก (THB)",
			"doccurrency":      "string — สกุลเงินเอกสาร (USD, JPY, etc.)",
			"exchangerate":     "float64 — อัตราแลกเปลี่ยน",
			"whtentries":       "array — ภาษีหัก ณ ที่จ่าย [{description, rate, taxbase, amount}]",
			"creditdays":       "int — เครดิตเทอม (วัน)",
			"duedate":          "string — วันครบกำหนดชำระ (YYYY-MM-DD)",
			"purchasetypecode": "string — ประเภทการจัดซื้อ (สำหรับ Approval)",
			"details":          "array — รายละเอียดสินค้า [{barcode, itemcode, itemnames, qty, price, sumamount, ...}]",
			"creatorcode":      "string — รหัสผู้สร้าง",
			"createdat":        "datetime — วันที่สร้าง",
			"updatedat":        "datetime — วันที่แก้ไขล่าสุด",
			"deletedat":        "datetime — วันที่ลบ (soft delete, zero = active)",
		},
		"examples": []map[string]interface{}{
			{
				"docno":     "PO-2026-0001",
				"transflag": 6,
				"custcode":  "CR-001",
				"custnames": []map[string]string{{"code": "th", "name": "บริษัท สมาร์ท ซัพพลาย จำกัด"}},
				"vattype":   1,
				"vatrate":   7,
				"details": []map[string]interface{}{
					{
						"barcode":   "8859100001234",
						"itemcode":  "SKU001",
						"qty":       100,
						"price":     50.0,
						"sumamount": 5000.0,
					},
				},
				"totalamount": 5350.0,
			},
		},
	}
}

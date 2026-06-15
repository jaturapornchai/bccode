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

const purchaseRequisitionCollection = "transactionPurchaseRequisition"

// ==================== Kafka Topics ====================

const (
	TopicPurchaseRequisitionCreated = "when-purchaserequisition-created"
	TopicPurchaseRequisitionUpdated = "when-purchaserequisition-updated"
	TopicPurchaseRequisitionDeleted = "when-purchaserequisition-deleted"
)

// ==================== PR Types ====================

// PRNameEntry ชื่อหลายภาษา
type PRNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// PRDetail รายละเอียดสินค้าในใบขอซื้อ
type PRDetail struct {
	LineNumber          int           `json:"linenumber" bson:"linenumber"`
	DocDatetime         time.Time     `json:"docdatetime" bson:"docdatetime"`
	Barcode             string        `json:"barcode" bson:"barcode"`
	ItemCode            string        `json:"itemcode" bson:"itemcode"`
	ItemNames           []PRNameEntry `json:"itemnames" bson:"itemnames"`
	UnitCode            string        `json:"unitcode" bson:"unitcode"`
	UnitNames           []PRNameEntry `json:"unitnames" bson:"unitnames"`
	ItemType            int8          `json:"itemtype" bson:"itemtype"`
	ItemGuid            string        `json:"itemguid" bson:"itemguid"`
	Qty                 float64       `json:"qty" bson:"qty"`
	Price               float64       `json:"price" bson:"price"`
	Discount            string        `json:"discount" bson:"discount"`
	DiscountAmount      float64       `json:"discountamount" bson:"discountamount"`
	SumAmount           float64       `json:"sumamount" bson:"sumamount"`
	SumAmountExcludeVat float64       `json:"sumamountexcludevat" bson:"sumamountexcludevat"`
	TotalValueVat       float64       `json:"totalvaluevat" bson:"totalvaluevat"`
	PriceExcludeVat     float64       `json:"priceexcludevat" bson:"priceexcludevat"`
	VatType             int8          `json:"vattype" bson:"vattype"`
	StandValue          float64       `json:"standvalue" bson:"standvalue"`
	DivideValue         float64       `json:"dividevalue" bson:"dividevalue"`
	WhCode              string        `json:"whcode" bson:"whcode"`
	WhNames             []PRNameEntry `json:"whnames" bson:"whnames"`
	LocationCode        string        `json:"locationcode" bson:"locationcode"`
	LocationNames       []PRNameEntry `json:"locationnames" bson:"locationnames"`
	Remark              string        `json:"remark" bson:"remark"`
	Description         string        `json:"description" bson:"description"`
}

// PRBranch สาขา
type PRBranch struct {
	GuidFixed string        `json:"guidfixed" bson:"guidfixed"`
	Code      string        `json:"code" bson:"code"`
	Names     []PRNameEntry `json:"names" bson:"names"`
}

// PurchaseRequisitionDocument เอกสารใบขอซื้อใน MongoDB
type PurchaseRequisitionDocument struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	GuidFixed   string             `json:"guidfixed" bson:"guidfixed"`
	// Header
	DocNo       string    `json:"docno" bson:"docno"`
	DocDatetime time.Time `json:"docdatetime" bson:"docdatetime"`
	TransFlag   int       `json:"transflag" bson:"transflag"`
	DocType     int8      `json:"doctype" bson:"doctype"`
	// Creditor (ผู้ขาย — กรณี PR ระบุได้แต่ไม่บังคับ)
	CustCode  string        `json:"custcode" bson:"custcode"`
	CustNames []PRNameEntry `json:"custnames" bson:"custnames"`
	// Financial
	Description    string  `json:"description" bson:"description"`
	TotalValue     float64 `json:"totalvalue" bson:"totalvalue"`
	TotalBeforeVat float64 `json:"totalbeforevat" bson:"totalbeforevat"`
	TotalAfterVat  float64 `json:"totalaftervat" bson:"totalaftervat"`
	TotalVatValue  float64 `json:"totalvatvalue" bson:"totalvatvalue"`
	TotalAmount    float64 `json:"totalamount" bson:"totalamount"`
	TotalExceptVat float64 `json:"totalexceptvat" bson:"totalexceptvat"`
	VatType        int8    `json:"vattype" bson:"vattype"`
	VatRate        float64 `json:"vatrate" bson:"vatrate"`
	TotalQty       float64 `json:"totalqty" bson:"totalqty"`
	IsCancel       bool    `json:"iscancel" bson:"iscancel"`
	Status         int8    `json:"status" bson:"status"`
	// Branch
	Branch PRBranch `json:"branch" bson:"branch"`
	// === PR-specific fields ===
	RequesterCode         string        `json:"requestercode" bson:"requestercode"`
	RequesterName         string        `json:"requestername" bson:"requestername"`
	DepartmentCode        string        `json:"departmentcode" bson:"departmentcode"`
	DepartmentNames       []PRNameEntry `json:"departmentnames" bson:"departmentnames"`
	Purpose               string        `json:"purpose" bson:"purpose"`
	BudgetCode            string        `json:"budgetcode,omitempty" bson:"budgetcode,omitempty"`
	BudgetAmount          float64       `json:"budgetamount,omitempty" bson:"budgetamount,omitempty"`
	Urgency               int8          `json:"urgency" bson:"urgency"` // 1=ปกติ 2=เร่งด่วน 3=เร่งด่วนมาก
	RequestedDeliveryDate string        `json:"requesteddeliverydate,omitempty" bson:"requesteddeliverydate,omitempty"`
	RefPODocNo            string        `json:"refpodocno,omitempty" bson:"refpodocno,omitempty"`
	RefRFQDocNo           string        `json:"refrfqdocno,omitempty" bson:"refrfqdocno,omitempty"`
	ConversionStatus      string        `json:"conversionstatus,omitempty" bson:"conversionstatus,omitempty"` // none/converted_to_rfq/converted_to_po
	// Purchase Type (for Approval)
	PurchaseTypeCode  string        `json:"purchasetypecode" bson:"purchasetypecode"`
	PurchaseTypeNames []PRNameEntry `json:"purchasetypenames" bson:"purchasetypenames"`
	// Creator/Updater
	CreatorCode string    `json:"creatorcode" bson:"creatorcode"`
	CreatorName string    `json:"creatorname" bson:"creatorname"`
	CreatedAt   time.Time `json:"createdat" bson:"createdat"`
	UpdaterCode string    `json:"modifiercode" bson:"updatercode"`
	UpdaterName string    `json:"modifiername" bson:"updatername"`
	UpdatedAt   time.Time `json:"updatedat" bson:"updatedat"`
	DeletedAt   time.Time `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
	// Details
	Details []PRDetail `json:"details" bson:"details"`
}

// ==================== List Purchase Requisitions ====================

type ListPurchaseRequisitionsResponse struct {
	PurchaseRequisitions []PurchaseRequisitionDocument `json:"purchaserequisitions"`
	Count                int                           `json:"count"`
	Keyword              string                        `json:"keyword,omitempty"`
	GeneratedAt          time.Time                     `json:"generatedat"`
}

func ListPurchaseRequisitions(ctx context.Context, holdingCode, keyword string, limit int) (*ListPurchaseRequisitionsResponse, error) {
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
				{"requestercode": bson.M{"$regex": keyword, "$options": "i"}},
				{"requestername": bson.M{"$regex": keyword, "$options": "i"}},
				{"departmentcode": bson.M{"$regex": keyword, "$options": "i"}},
				{"purpose": bson.M{"$regex": keyword, "$options": "i"}},
				{"description": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListPurchaseRequisitions] holdingCode=%s, keyword=%s, limit=%d", holdingCode, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(purchaseRequisitionCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "docdatetime", Value: -1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []PurchaseRequisitionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}
	if docs == nil {
		docs = []PurchaseRequisitionDocument{}
	}

	return &ListPurchaseRequisitionsResponse{
		PurchaseRequisitions: docs,
		Count:                len(docs),
		Keyword:              keyword,
		GeneratedAt:          time.Now(),
	}, nil
}

// ==================== Create Purchase Requisition ====================

type CreatePurchaseRequisitionResponse struct {
	Success             bool                        `json:"success"`
	Message             string                      `json:"message"`
	PurchaseRequisition PurchaseRequisitionDocument `json:"purchaserequisition"`
	KafkaSync           string                      `json:"kafkasync"`
	KafkaError          string                      `json:"kafkaerror,omitempty"`
	GeneratedAt         time.Time                   `json:"generatedat"`
}

func CreatePurchaseRequisition(ctx context.Context, holdingCode, docno, requesterCode, requesterName, departmentCode, departmentNamesJSON, purpose, budgetCode string, budgetAmount float64, urgency int8, requestedDeliveryDate, detailsJSON, description string, totalamount float64) (*CreatePurchaseRequisitionResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if docno == "" {
		return nil, fmt.Errorf("docno is required")
	}
	if requesterCode == "" {
		return nil, fmt.Errorf("requestercode is required")
	}
	if departmentCode == "" {
		return nil, fmt.Errorf("departmentcode is required")
	}
	if purpose == "" {
		return nil, fmt.Errorf("purpose is required")
	}
	if urgency < 1 || urgency > 3 {
		urgency = 1 // default ปกติ
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	coll := mongoClient.Database(dbName).Collection(purchaseRequisitionCollection)

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
		return nil, fmt.Errorf("ใบขอซื้อ docno '%s' มีอยู่แล้วใน shop นี้", docno)
	}

	// Parse department names JSON
	var departmentNames []PRNameEntry
	if departmentNamesJSON != "" {
		if err := json.Unmarshal([]byte(departmentNamesJSON), &departmentNames); err != nil {
			return nil, fmt.Errorf("departmentnames JSON ไม่ถูกต้อง: %w", err)
		}
	}

	// Parse details JSON
	var details []PRDetail
	if detailsJSON != "" {
		if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
			return nil, fmt.Errorf("details JSON ไม่ถูกต้อง: %w", err)
		}
	}

	now := time.Now()
	doc := PurchaseRequisitionDocument{
		HoldingCode:           holdingCode,
		GuidFixed:             uuid.New().String(),
		DocNo:                 docno,
		DocDatetime:           now,
		TransFlag:             21, // PR transflag
		RequesterCode:         requesterCode,
		RequesterName:         requesterName,
		DepartmentCode:        departmentCode,
		DepartmentNames:       departmentNames,
		Purpose:               purpose,
		BudgetCode:            budgetCode,
		BudgetAmount:          budgetAmount,
		Urgency:               urgency,
		RequestedDeliveryDate: requestedDeliveryDate,
		ConversionStatus:      "none",
		Description:           description,
		TotalAmount:           totalamount,
		Details:               details,
		CreatorCode:           "mcp-tool",
		CreatorName:           "MCP Tool",
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้างใบขอซื้อล้มเหลว: %w", err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicPurchaseRequisitionCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreatePurchaseRequisition] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreatePurchaseRequisition] สร้าง docno=%s สำเร็จ (shop=%s, kafka=%s)", docno, holdingCode, kafkaSync)

	return &CreatePurchaseRequisitionResponse{
		Success:             true,
		Message:             fmt.Sprintf("สร้างใบขอซื้อ '%s' สำเร็จ", docno),
		PurchaseRequisition: doc,
		KafkaSync:           kafkaSync,
		KafkaError:          kafkaError,
		GeneratedAt:         time.Now(),
	}, nil
}

// ==================== Update Purchase Requisition ====================

type UpdatePurchaseRequisitionResponse struct {
	Success             bool                        `json:"success"`
	Message             string                      `json:"message"`
	PurchaseRequisition PurchaseRequisitionDocument `json:"purchaserequisition"`
	KafkaSync           string                      `json:"kafkasync"`
	KafkaError          string                      `json:"kafkaerror,omitempty"`
	GeneratedAt         time.Time                   `json:"generatedat"`
}

func UpdatePurchaseRequisition(ctx context.Context, holdingCode, guidfixed, requesterCode, requesterName, departmentCode, departmentNamesJSON, purpose, budgetCode string, budgetAmount float64, urgency int8, requestedDeliveryDate, detailsJSON, description string, totalamount float64, status int8, conversionStatus string) (*UpdatePurchaseRequisitionResponse, error) {
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
	coll := mongoClient.Database(dbName).Collection(purchaseRequisitionCollection)

	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guidfixed,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	var existing PurchaseRequisitionDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบใบขอซื้อ guidfixed='%s'", guidfixed)
	}

	updateFields := bson.M{
		"updatedat":   time.Now(),
		"updatercode": "mcp-tool",
		"updatername": "MCP Tool",
	}

	if requesterCode != "" {
		updateFields["requestercode"] = requesterCode
	}
	if requesterName != "" {
		updateFields["requestername"] = requesterName
	}
	if departmentCode != "" {
		updateFields["departmentcode"] = departmentCode
	}
	if departmentNamesJSON != "" {
		var departmentNames []PRNameEntry
		if err := json.Unmarshal([]byte(departmentNamesJSON), &departmentNames); err != nil {
			return nil, fmt.Errorf("departmentnames JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["departmentnames"] = departmentNames
	}
	if purpose != "" {
		updateFields["purpose"] = purpose
	}
	if budgetCode != "" {
		updateFields["budgetcode"] = budgetCode
	}
	if budgetAmount > 0 {
		updateFields["budgetamount"] = budgetAmount
	}
	if urgency >= 1 && urgency <= 3 {
		updateFields["urgency"] = urgency
	}
	if requestedDeliveryDate != "" {
		updateFields["requesteddeliverydate"] = requestedDeliveryDate
	}
	if detailsJSON != "" {
		var details []PRDetail
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
	if conversionStatus != "" {
		updateFields["conversionstatus"] = conversionStatus
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดตใบขอซื้อล้มเหลว: %w", err)
	}

	var updated PurchaseRequisitionDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicPurchaseRequisitionUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdatePurchaseRequisition] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP UpdatePurchaseRequisition] อัปเดต guidfixed=%s สำเร็จ (shop=%s)", guidfixed, holdingCode)

	return &UpdatePurchaseRequisitionResponse{
		Success:             true,
		Message:             fmt.Sprintf("อัปเดตใบขอซื้อ '%s' สำเร็จ", existing.DocNo),
		PurchaseRequisition: updated,
		KafkaSync:           kafkaSync,
		KafkaError:          kafkaError,
		GeneratedAt:         time.Now(),
	}, nil
}

// ==================== Delete Purchase Requisition (Soft Delete) ====================

type DeletePurchaseRequisitionResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	GuidFixed   string    `json:"guidfixed"`
	DocNo       string    `json:"docno"`
	KafkaSync   string    `json:"kafkasync"`
	KafkaError  string    `json:"kafkaerror,omitempty"`
	GeneratedAt time.Time `json:"generatedat"`
}

func DeletePurchaseRequisition(ctx context.Context, holdingCode, guidfixed string) (*DeletePurchaseRequisitionResponse, error) {
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
	coll := mongoClient.Database(dbName).Collection(purchaseRequisitionCollection)

	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guidfixed,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	var existing PurchaseRequisitionDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบใบขอซื้อ guidfixed='%s'", guidfixed)
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
		return nil, fmt.Errorf("ลบใบขอซื้อล้มเหลว: %w", err)
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicPurchaseRequisitionDeleted, existing); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP DeletePurchaseRequisition] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP DeletePurchaseRequisition] ลบ docno=%s สำเร็จ (shop=%s)", existing.DocNo, holdingCode)

	return &DeletePurchaseRequisitionResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบใบขอซื้อ '%s' สำเร็จ", existing.DocNo),
		GuidFixed:   guidfixed,
		DocNo:       existing.DocNo,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Get Purchase Requisition Schema ====================

func GetPurchaseRequisitionSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  purchaseRequisitionCollection,
		"description": "ใบขอซื้อ (Purchase Requisition / PR) — เอกสารคำขอซื้อจากแผนก ส่งให้หัวหน้าอนุมัติ แล้วฝ่ายจัดซื้อดำเนินการ",
		"transflag":   21,
		"flow":        "PR (draft) → PR (approved) → RFQ หรือ PO",
		"fields": map[string]interface{}{
			"_id":                   "ObjectID — MongoDB auto-generated ID",
			"holdingcode":           "string — Holding Code (tenant isolation)",
			"guidfixed":             "string — UUID สำหรับอ้างอิงภายใน",
			"docno":                 "string (required) — เลขที่เอกสาร เช่น PR20260314-00001",
			"docdatetime":           "datetime — วันที่เอกสาร",
			"transflag":             "int — ประเภท transaction (21=ใบขอซื้อ)",
			"requestercode":         "string (required) — รหัสผู้ขอซื้อ (รหัสพนักงาน)",
			"requestername":         "string — ชื่อผู้ขอซื้อ",
			"departmentcode":        "string (required) — รหัสแผนก",
			"departmentnames":       "array — ชื่อแผนกหลายภาษา [{code:'th', name:'ฝ่ายไอที'}]",
			"purpose":               "string (required) — วัตถุประสงค์/เหตุผลในการขอซื้อ",
			"budgetcode":            "string — รหัสงบประมาณ",
			"budgetamount":          "float64 — วงเงินงบประมาณ",
			"urgency":               "int8 — ความเร่งด่วน (1=ปกติ, 2=เร่งด่วน, 3=เร่งด่วนมาก)",
			"requesteddeliverydate": "string — วันที่ต้องการรับ (YYYY-MM-DD)",
			"conversionstatus":      "string — สถานะแปลง (none/converted_to_rfq/converted_to_po)",
			"refpodocno":            "string — เลขที่ PO ที่สร้างจาก PR",
			"refrfqdocno":           "string — เลขที่ RFQ ที่สร้างจาก PR",
			"totalamount":           "float64 — มูลค่ารวมทั้งหมด",
			"status":                "int8 — สถานะ (0=ร่าง, 1=รออนุมัติ, 2=อนุมัติ, 3=ปฏิเสธ)",
			"details":               "array — รายละเอียดสินค้า [{barcode, itemcode, itemnames, qty, price, sumamount, ...}]",
			"createdat":             "datetime — วันที่สร้าง",
			"deletedat":             "datetime — วันที่ลบ (soft delete, zero = active)",
		},
		"examples": []map[string]interface{}{
			{
				"docno":                 "PR20260314-00001",
				"transflag":             21,
				"requestercode":         "EMP001",
				"requestername":         "สมชาย ใจดี",
				"departmentcode":        "IT",
				"departmentnames":       []map[string]string{{"code": "th", "name": "ฝ่ายไอที"}},
				"purpose":               "ซื้อคอมพิวเตอร์สำหรับพนักงานใหม่ 3 คน",
				"urgency":               2,
				"requesteddeliverydate": "2026-03-20",
				"details": []map[string]interface{}{
					{
						"barcode":   "8859100001234",
						"itemcode":  "NOTEBOOK001",
						"qty":       3,
						"price":     25000.0,
						"sumamount": 75000.0,
					},
				},
				"totalamount": 75000.0,
			},
		},
	}
}

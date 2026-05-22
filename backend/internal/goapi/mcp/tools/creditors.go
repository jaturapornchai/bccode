package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/myollama"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const creditorCollection = "creditors"

// ==================== Creditor Types ====================

// CreditorNameEntry ชื่อเจ้าหนี้แต่ละภาษา
type CreditorNameEntry struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
	IsAuto bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// CreditorAddress ที่อยู่เจ้าหนี้
type CreditorAddress struct {
	GUID string              `json:"guid" bson:"guid"`
	Address []string            `json:"address" bson:"address"`
	CountryCode string              `json:"country_code" bson:"country_code"`
	ProvinceCode string              `json:"province_code" bson:"province_code"`
	DistrictCode string              `json:"district_code" bson:"district_code"`
	SubDistrictCode string              `json:"sub_district_code" bson:"sub_district_code"`
	ZipCode string              `json:"zip_code" bson:"zip_code"`
	ContactNames []CreditorNameEntry `json:"contactnames" bson:"contactnames"`
	PhonePrimary string              `json:"phone_primary" bson:"phone_primary"`
	PhoneSecondary string              `json:"phone_secondary" bson:"phone_secondary"`
}

// CreditorDocument เอกสารเจ้าหนี้ใน MongoDB
type CreditorDocument struct {
	ID primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	ShopID string              `json:"shopid" bson:"shopid"`
	GuidFixed string              `json:"guid_fixed" bson:"guid_fixed"`
	Code string              `json:"code" bson:"code"`
	PersonalType int8                `json:"personal_type" bson:"personal_type"`
	Names []CreditorNameEntry `json:"names" bson:"names"`
	TaxId string              `json:"tax_id" bson:"tax_id"`
	Email string              `json:"email" bson:"email"`
	CreditDay int                 `json:"creditday" bson:"creditday"`
	BranchNumber string              `json:"branch_number" bson:"branch_number"`
	IsMember bool                `json:"ismember" bson:"ismember"`
	AddressForBilling CreditorAddress     `json:"addressforbilling" bson:"addressforbilling"`
	CreatedBy string              `json:"createdby" bson:"createdby"`
	CreatedAt time.Time           `json:"created_at" bson:"created_at"`
	UpdatedBy string              `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	UpdatedAt time.Time           `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt time.Time           `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// ==================== List Creditors ====================

type ListCreditorsResponse struct {
	Creditors []CreditorDocument `json:"creditors"`
	Count int                `json:"count"`
	Keyword string             `json:"keyword,omitempty"`
	GeneratedAt time.Time          `json:"generated_at"`
}

func ListCreditors(ctx context.Context, shopID, keyword string, limit int) (*ListCreditorsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
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
		"shopid": shopID,
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
				{"tax_id": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListCreditors] shopID=%s, keyword=%s, limit=%d", shopID, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(creditorCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var creditors []CreditorDocument
	if err := cursor.All(ctx, &creditors); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}
	if creditors == nil {
		creditors = []CreditorDocument{}
	}

	// Vector search fallback — ถ้า MongoDB regex หาได้น้อยกว่า 3 ลอง PG vector search
	if keyword != "" && len(creditors) < 3 {
		vectorCreditors, vecErr := searchCreditorsVector(shopID, keyword, limit)
		if vecErr == nil && len(vectorCreditors) > 0 {
			creditors = mergeCreditors(creditors, vectorCreditors, limit)
			logger.Info("[MCP ListCreditors] Vector search added results, total=%d", len(creditors))
		}
	}

	return &ListCreditorsResponse{
		Creditors:   creditors,
		Count:       len(creditors),
		Keyword:     keyword,
		GeneratedAt: time.Now(),
	}, nil
}

// searchCreditorsVector — ค้นหาเจ้าหนี้ด้วย pgvector cosine similarity
func searchCreditorsVector(shopID, keyword string, limit int) ([]CreditorDocument, error) {
	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, err
	}

	var colExists bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='creditor' AND column_name='name_embedding')`).Scan(&colExists)
	if !colExists {
		return nil, fmt.Errorf("no name_embedding column")
	}

	var hasEmb bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM creditor WHERE name_embedding IS NOT NULL LIMIT 1)`).Scan(&hasEmb)
	if !hasEmb {
		return nil, fmt.Errorf("no embeddings")
	}

	queryEmb, err := myollama.GenerateSingleEmbedding(keyword)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	vecStr := float32SliceToVectorString(queryEmb)
	query := `SELECT guidfixed, code, names, COALESCE(taxid,''), COALESCE(email,''),
			(name_embedding <=> $1::vector) as distance
		FROM creditor
		WHERE shopid = $2 AND name_embedding IS NOT NULL
		ORDER BY name_embedding <=> $1::vector
		LIMIT $3`

	rows, err := db.Query(query, vecStr, shopID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CreditorDocument
	for rows.Next() {
		var c CreditorDocument
		var namesJSON string
		var distance float64
		if err := rows.Scan(&c.GuidFixed, &c.Code, &namesJSON, &c.TaxId, &c.Email, &distance); err != nil {
			continue
		}
		if distance > 0.5 {
			continue
		}
		c.ShopID = shopID
		if namesJSON != "" {
			json.Unmarshal([]byte(namesJSON), &c.Names)
		}
		results = append(results, c)
	}
	return results, nil
}

// mergeCreditors — รวมผลลัพธ์โดยไม่ซ้ำ (ใช้ guidfixed เป็น key)
func mergeCreditors(existing, extra []CreditorDocument, limit int) []CreditorDocument {
	seen := make(map[string]bool, len(existing))
	for _, c := range existing {
		seen[c.GuidFixed] = true
	}
	merged := existing
	for _, c := range extra {
		if len(merged) >= limit {
			break
		}
		if !seen[c.GuidFixed] {
			merged = append(merged, c)
			seen[c.GuidFixed] = true
		}
	}
	return merged
}

// ==================== Search Creditors (semantic-first) ====================

// SearchCreditorsResponse — result สำหรับ search_creditors MCP tool
type SearchCreditorsResponse struct {
	Creditors []CreditorDocument `json:"creditors"`
	Count int                `json:"count"`
	Keyword string             `json:"keyword"`
	SearchMode string             `json:"search_mode"` // "vector", "regex", "combined"
	GeneratedAt time.Time          `json:"generated_at"`
}

// SearchCreditors — vector-first semantic search สำหรับ agent ใช้
func SearchCreditors(ctx context.Context, shopID, keyword string, limit int) (*SearchCreditorsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	logger.Info("[SearchEntity] search_creditors shopID=%s keyword=%s limit=%d", shopID, keyword, limit)

	// 1. Try pgvector first
	vectorResults, vecErr := searchCreditorsVector(shopID, keyword, limit)
	if vecErr == nil && len(vectorResults) > 0 {
		logger.Info("[SearchEntity] search_creditors vector found %d results", len(vectorResults))
		return &SearchCreditorsResponse{
			Creditors:   vectorResults,
			Count:       len(vectorResults),
			Keyword:     keyword,
			SearchMode:  "vector",
			GeneratedAt: time.Now(),
		}, nil
	}
	if vecErr != nil {
		logger.Info("[SearchEntity] search_creditors vector fallback (err: %v) — trying regex", vecErr)
	}

	// 2. Fallback: MongoDB regex search
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}
	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()

	baseFilter := bson.M{
		"shopid": shopID,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}
	keywordFilter := BuildEntityKeywordFilter(keyword, []string{"code", "names.name", "tax_id"})
	logger.Info("[SearchEntity] search_creditors tokens=%v", TokenizeEntityKeyword(keyword))
	filter := bson.M{"$and": []bson.M{baseFilter, keywordFilter}}

	coll := mongoClient.Database(dbName).Collection(creditorCollection)
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var creditors []CreditorDocument
	if err := cursor.All(ctx, &creditors); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}
	if creditors == nil {
		creditors = []CreditorDocument{}
	}

	logger.Info("[SearchEntity] search_creditors regex found %d results", len(creditors))
	return &SearchCreditorsResponse{
		Creditors:   creditors,
		Count:       len(creditors),
		Keyword:     keyword,
		SearchMode:  "regex",
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Creditor ====================

type CreateCreditorResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Creditor CreditorDocument `json:"creditor"`
	KafkaSync string           `json:"kafka_sync"`
	KafkaError string           `json:"kafka_error,omitempty"`
	GeneratedAt time.Time        `json:"generated_at"`
}

func CreateCreditor(ctx context.Context, shopID, code, namesJSON, taxid, email string, personaltype int8, creditday int, addressJSON string) (*CreateCreditorResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
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
	coll := mongoClient.Database(dbName).Collection(creditorCollection)

	// ตรวจสอบ code ซ้ำ
	existFilter := bson.M{
		"shopid": shopID,
		"code":   code,
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
		return nil, fmt.Errorf("creditor code '%s' มีอยู่แล้วใน shop นี้", code)
	}

	// Parse names JSON
	var names []CreditorNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required — ต้องระบุชื่อเจ้าหนี้อย่างน้อย 1 ภาษา เช่น [{\"code\":\"th\",\"name\":\"บริษัท ABC\"}]")
	}

	// Parse address JSON
	var address CreditorAddress
	if addressJSON != "" {
		if err := json.Unmarshal([]byte(addressJSON), &address); err != nil {
			return nil, fmt.Errorf("addressforbilling JSON ไม่ถูกต้อง: %w", err)
		}
		if address.GUID == "" {
			address.GUID = uuid.New().String()
		}
	}

	now := time.Now()
	doc := CreditorDocument{
		ShopID:            shopID,
		GuidFixed:         uuid.New().String(),
		Code:              code,
		PersonalType:      personaltype,
		Names:             names,
		TaxId:             taxid,
		Email:             email,
		CreditDay:         creditday,
		AddressForBilling: address,
		CreatedBy:         "mcp-tool",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("สร้างเจ้าหนี้ล้มเหลว: %w", err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicCreditorCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateCreditor] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreateCreditor] สร้าง code=%s สำเร็จ (shop=%s, kafka=%s)", code, shopID, kafkaSync)

	return &CreateCreditorResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้างเจ้าหนี้ '%s' สำเร็จ", code),
		Creditor:    doc,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Creditors (Bulk) ====================

type CreateCreditorsItem struct {
	Code string              `json:"code"`
	Names []CreditorNameEntry `json:"names"`
	PersonalType int8                `json:"personal_type"`
	TaxId string              `json:"tax_id"`
	Email string              `json:"email"`
	CreditDay int                 `json:"creditday"`
	AddressForBilling CreditorAddress     `json:"addressforbilling"`
}

type CreateCreditorsResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Created []CreditorDocument `json:"created"`
	Skipped []string           `json:"skipped,omitempty"`
	CreatedCount int                `json:"created_count"`
	SkippedCount int                `json:"skipped_count"`
	KafkaSync string             `json:"kafka_sync"`
	KafkaError string             `json:"kafka_error,omitempty"`
	GeneratedAt time.Time          `json:"generated_at"`
}

func CreateCreditors(ctx context.Context, shopID, creditorsJSON string) (*CreateCreditorsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if creditorsJSON == "" {
		return nil, fmt.Errorf("creditors JSON is required")
	}

	var items []CreateCreditorsItem
	if err := json.Unmarshal([]byte(creditorsJSON), &items); err != nil {
		return nil, fmt.Errorf("creditors JSON ไม่ถูกต้อง: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ต้องระบุเจ้าหนี้อย่างน้อย 1 รายการ")
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
	coll := mongoClient.Database(dbName).Collection(creditorCollection)

	// ดึง code ที่มีอยู่แล้ว
	var codes []string
	for _, item := range items {
		codes = append(codes, item.Code)
	}

	existFilter := bson.M{
		"shopid": shopID,
		"code":   bson.M{"$in": codes},
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
	var createdDocs []CreditorDocument
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

		addr := item.AddressForBilling
		if addr.GUID == "" {
			addr.GUID = uuid.New().String()
		}

		doc := CreditorDocument{
			ShopID:            shopID,
			GuidFixed:         uuid.New().String(),
			Code:              item.Code,
			PersonalType:      item.PersonalType,
			Names:             item.Names,
			TaxId:             item.TaxId,
			Email:             item.Email,
			CreditDay:         item.CreditDay,
			AddressForBilling: addr,
			CreatedBy:         "mcp-tool",
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		docsToInsert = append(docsToInsert, doc)
		createdDocs = append(createdDocs, doc)
	}

	kafkaSync := "skipped"
	kafkaError := ""

	if len(docsToInsert) > 0 {
		_, err = coll.InsertMany(ctx, docsToInsert)
		if err != nil {
			return nil, fmt.Errorf("สร้างเจ้าหนี้ล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if err := publishToKafkaBatch(ctx, TopicCreditorBulkCreated, docsToInsert); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP CreateCreditors] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
		}
	}

	logger.Info("[MCP CreateCreditors] สร้าง %d รายการสำเร็จ, ข้าม %d รายการ (shop=%s)", len(createdDocs), len(skipped), shopID)

	return &CreateCreditorsResponse{
		Success:      true,
		Message:      fmt.Sprintf("สร้างเจ้าหนี้ %d รายการสำเร็จ", len(createdDocs)),
		Created:      createdDocs,
		Skipped:      skipped,
		CreatedCount: len(createdDocs),
		SkippedCount: len(skipped),
		KafkaSync:    kafkaSync,
		KafkaError:   kafkaError,
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== Update Creditor ====================

type UpdateCreditorResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Creditor CreditorDocument `json:"creditor"`
	KafkaSync string           `json:"kafka_sync"`
	KafkaError string           `json:"kafka_error,omitempty"`
	GeneratedAt time.Time        `json:"generated_at"`
}

func UpdateCreditor(ctx context.Context, shopID, guidfixed, namesJSON, taxid, email string, creditday int, addressJSON string) (*UpdateCreditorResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
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
	coll := mongoClient.Database(dbName).Collection(creditorCollection)

	filter := bson.M{
		"shopid":    shopID,
		"guid_fixed": guidfixed,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	var existing CreditorDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบเจ้าหนี้ guidfixed='%s'", guidfixed)
	}

	updateFields := bson.M{
		"updated_at": time.Now(),
		"updatedby": "mcp-tool",
	}

	if namesJSON != "" {
		var names []CreditorNameEntry
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["names"] = names
	}
	if taxid != "" {
		updateFields["tax_id"] = taxid
	}
	if email != "" {
		updateFields["email"] = email
	}
	if creditday > 0 {
		updateFields["creditday"] = creditday
	}
	if addressJSON != "" {
		var address CreditorAddress
		if err := json.Unmarshal([]byte(addressJSON), &address); err != nil {
			return nil, fmt.Errorf("addressforbilling JSON ไม่ถูกต้อง: %w", err)
		}
		if address.GUID == "" {
			address.GUID = existing.AddressForBilling.GUID
			if address.GUID == "" {
				address.GUID = uuid.New().String()
			}
		}
		updateFields["addressforbilling"] = address
	}

	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("อัปเดตเจ้าหนี้ล้มเหลว: %w", err)
	}

	var updated CreditorDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicCreditorUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdateCreditor] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP UpdateCreditor] อัปเดต guidfixed=%s สำเร็จ (shop=%s)", guidfixed, shopID)

	return &UpdateCreditorResponse{
		Success:     true,
		Message:     fmt.Sprintf("อัปเดตเจ้าหนี้ '%s' สำเร็จ", existing.Code),
		Creditor:    updated,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Creditor (Soft Delete) ====================

type DeleteCreditorResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	GuidFixed string    `json:"guid_fixed"`
	Code string    `json:"code"`
	KafkaSync string    `json:"kafka_sync"`
	KafkaError string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

func DeleteCreditor(ctx context.Context, shopID, guidfixed string) (*DeleteCreditorResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
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
	coll := mongoClient.Database(dbName).Collection(creditorCollection)

	filter := bson.M{
		"shopid":    shopID,
		"guid_fixed": guidfixed,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}

	var existing CreditorDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบเจ้าหนี้ guidfixed='%s'", guidfixed)
	}

	// Soft delete
	now := time.Now()
	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
		"deleted_at": now,
		"updatedby": "mcp-tool",
		"updated_at": now,
	}})
	if err != nil {
		return nil, fmt.Errorf("ลบเจ้าหนี้ล้มเหลว: %w", err)
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicCreditorDeleted, existing); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP DeleteCreditor] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP DeleteCreditor] ลบ code=%s สำเร็จ (shop=%s)", existing.Code, shopID)

	return &DeleteCreditorResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบเจ้าหนี้ '%s' สำเร็จ", existing.Code),
		GuidFixed:   guidfixed,
		Code:        existing.Code,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Creditors (Bulk Soft Delete) ====================

type DeleteCreditorsResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Deleted []string  `json:"deleted"`
	NotFound []string  `json:"not_found,omitempty"`
	DeletedCount int       `json:"deleted_count"`
	NotFoundCount int       `json:"not_found_count"`
	KafkaSync string    `json:"kafka_sync"`
	KafkaError string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

func DeleteCreditors(ctx context.Context, shopID, guidfixedsJSON string) (*DeleteCreditorsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
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
	coll := mongoClient.Database(dbName).Collection(creditorCollection)

	// ค้นหา guidfixeds ที่มีอยู่จริง (ยังไม่ถูกลบ)
	existFilter := bson.M{
		"shopid":    shopID,
		"guid_fixed": bson.M{"$in": guidfixeds},
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": time.Time{}},
		},
	}
	cursor, err := coll.Find(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ค้นหาเจ้าหนี้ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingGUIDs := make(map[string]bool)
	var docsToDelete []interface{}
	for cursor.Next(ctx) {
		var doc CreditorDocument
		if err := cursor.Decode(&doc); err == nil {
			existingGUIDs[doc.GuidFixed] = true
			docsToDelete = append(docsToDelete, doc)
		}
	}

	var deleted []string
	var notFound []string
	for _, gf := range guidfixeds {
		if existingGUIDs[gf] {
			deleted = append(deleted, gf)
		} else {
			notFound = append(notFound, gf)
		}
	}

	kafkaSync := "skipped"
	kafkaError := ""

	if len(deleted) > 0 {
		now := time.Now()
		updateFilter := bson.M{
			"shopid":    shopID,
			"guid_fixed": bson.M{"$in": deleted},
		}
		_, err = coll.UpdateMany(ctx, updateFilter, bson.M{"$set": bson.M{
			"deleted_at": now,
			"updatedby": "mcp-tool",
			"updated_at": now,
		}})
		if err != nil {
			return nil, fmt.Errorf("ลบเจ้าหนี้ล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if len(docsToDelete) > 0 {
			if err := publishToKafkaBatch(ctx, TopicCreditorBulkDeleted, docsToDelete); err != nil {
				kafkaSync = "failed"
				kafkaError = err.Error()
				logger.Warn("[MCP DeleteCreditors] Kafka publish ล้มเหลว: %v", err)
			}
		}
	}

	logger.Info("[MCP DeleteCreditors] ลบ %d รายการสำเร็จ, ไม่พบ %d รายการ (shop=%s)", len(deleted), len(notFound), shopID)

	return &DeleteCreditorsResponse{
		Success:       true,
		Message:       fmt.Sprintf("ลบเจ้าหนี้ %d รายการสำเร็จ", len(deleted)),
		Deleted:       deleted,
		NotFound:      notFound,
		DeletedCount:  len(deleted),
		NotFoundCount: len(notFound),
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Get Creditor Schema ====================

func GetCreditorSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  creditorCollection,
		"description": "เจ้าหนี้ (Creditor) — ข้อมูลผู้ขาย/เจ้าหนี้ที่ธุรกิจต้องชำระเงินให้",
		"fields": map[string]interface{}{
			"_id":               "ObjectID — MongoDB auto-generated ID",
			"shopid":            "string — Shop ID (tenant isolation)",
			"guid_fixed":         "string — UUID สำหรับอ้างอิงภายใน",
			"code":              "string (required) — รหัสเจ้าหนี้ เช่น CR-001",
			"personal_type":      "int8 — ประเภท: 1=บุคคลธรรมดา, 2=นิติบุคคล",
			"names":             "array (required) — ชื่อหลายภาษา [{code:'th', name:'บริษัท ABC'}]",
			"tax_id":             "string — เลขผู้เสียภาษี",
			"email":             "string — อีเมล",
			"creditday":         "int — วงเงินเครดิต (วัน)",
			"branch_number":      "string — เลขที่สาขา",
			"ismember":          "bool — เป็นสมาชิกหรือไม่",
			"addressforbilling": "object — ที่อยู่สำหรับออกบิล {address, countrycode, provincecode, districtcode, subdistrictcode, zipcode, phoneprimary}",
			"createdby":         "string — ผู้สร้าง",
			"created_at":         "datetime — วันที่สร้าง",
			"updated_at":         "datetime — วันที่แก้ไขล่าสุด",
			"deleted_at":         "datetime — วันที่ลบ (soft delete, zero = active)",
		},
		"indexes": []string{
			"shopid + code (unique per shop)",
			"shopid + guidfixed",
		},
		"examples": []map[string]interface{}{
			{
				"code":         "CR-001",
				"personal_type": 2,
				"names":        []map[string]string{{"code": "th", "name": "บริษัท สมาร์ท ซัพพลาย จำกัด"}, {"code": "en", "name": "Smart Supply Co., Ltd."}},
				"tax_id":        "0105565012345",
				"creditday":    30,
			},
		},
	}
}

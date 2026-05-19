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

const debtorCollection = "debtors"

// ==================== Debtor Types ====================

// DebtorNameEntry ชื่อลูกหนี้แต่ละภาษา
type DebtorNameEntry struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

// DebtorAddress ที่อยู่ลูกหนี้
type DebtorAddress struct {
	GUID            string            `json:"guid" bson:"guid"`
	Address         []string          `json:"address" bson:"address"`
	CountryCode     string            `json:"countrycode" bson:"countrycode"`
	ProvinceCode    string            `json:"provincecode" bson:"provincecode"`
	DistrictCode    string            `json:"districtcode" bson:"districtcode"`
	SubDistrictCode string            `json:"subdistrictcode" bson:"subdistrictcode"`
	ZipCode         string            `json:"zipcode" bson:"zipcode"`
	ContactNames    []DebtorNameEntry `json:"contactnames" bson:"contactnames"`
	PhonePrimary    string            `json:"phoneprimary" bson:"phoneprimary"`
	PhoneSecondary  string            `json:"phonesecondary" bson:"phonesecondary"`
}

// DebtorDocument เอกสารลูกหนี้ใน MongoDB
type DebtorDocument struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopID            string             `json:"shopid" bson:"shopid"`
	GuidFixed         string             `json:"guidfixed" bson:"guidfixed"`
	Code              string             `json:"code" bson:"code"`
	PersonalType      int8               `json:"personaltype" bson:"personaltype"`
	Names             []DebtorNameEntry  `json:"names" bson:"names"`
	TaxId             string             `json:"taxid" bson:"taxid"`
	Email             string             `json:"email" bson:"email"`
	CreditDay         int                `json:"creditday" bson:"creditday"`
	BranchNumber      string             `json:"branchnumber" bson:"branchnumber"`
	IsMember          bool               `json:"ismember" bson:"ismember"`
	AddressForBilling DebtorAddress      `json:"addressforbilling" bson:"addressforbilling"`
	CreatedBy         string             `json:"createdby" bson:"createdby"`
	CreatedAt         time.Time          `json:"createdat" bson:"createdat"`
	UpdatedBy         string             `json:"updatedby,omitempty" bson:"updatedby,omitempty"`
	UpdatedAt         time.Time          `json:"updatedat,omitempty" bson:"updatedat,omitempty"`
	DeletedAt         time.Time          `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
}

// ==================== List Debtors ====================

type ListDebtorsResponse struct {
	Debtors     []DebtorDocument `json:"debtors"`
	Count       int              `json:"count"`
	Keyword     string           `json:"keyword,omitempty"`
	GeneratedAt time.Time        `json:"generated_at"`
}

func ListDebtors(ctx context.Context, shopID, keyword string, limit int) (*ListDebtorsResponse, error) {
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
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	if keyword != "" {
		keywordFilter := bson.M{
			"$or": []bson.M{
				{"code": bson.M{"$regex": keyword, "$options": "i"}},
				{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
				{"taxid": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		filter = bson.M{"$and": []bson.M{filter, keywordFilter}}
	}

	logger.Info("[MCP ListDebtors] shopID=%s, keyword=%s, limit=%d", shopID, keyword, limit)

	coll := mongoClient.Database(dbName).Collection(debtorCollection)
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var debtors []DebtorDocument
	if err := cursor.All(ctx, &debtors); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}
	if debtors == nil {
		debtors = []DebtorDocument{}
	}

	// Vector search fallback — ถ้า MongoDB regex หาได้น้อยกว่า 3 ลอง PG vector search
	if keyword != "" && len(debtors) < 3 {
		vectorDebtors, vecErr := searchDebtorsVector(shopID, keyword, limit)
		if vecErr == nil && len(vectorDebtors) > 0 {
			debtors = mergeDebtors(debtors, vectorDebtors, limit)
			logger.Info("[MCP ListDebtors] Vector search added results, total=%d", len(debtors))
		}
	}

	return &ListDebtorsResponse{
		Debtors:     debtors,
		Count:       len(debtors),
		Keyword:     keyword,
		GeneratedAt: time.Now(),
	}, nil
}

// searchDebtorsVector — ค้นหาลูกหนี้ด้วย pgvector cosine similarity
func searchDebtorsVector(shopID, keyword string, limit int) ([]DebtorDocument, error) {
	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, err
	}

	// ตรวจว่ามี column name_embedding
	var colExists bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='debtor' AND column_name='name_embedding')`).Scan(&colExists)
	if !colExists {
		return nil, fmt.Errorf("no name_embedding column")
	}

	// ตรวจว่ามี embeddings อยู่จริง
	var hasEmb bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM debtor WHERE name_embedding IS NOT NULL LIMIT 1)`).Scan(&hasEmb)
	if !hasEmb {
		return nil, fmt.Errorf("no embeddings")
	}

	// สร้าง embedding จาก keyword
	queryEmb, err := myollama.GenerateSingleEmbedding(keyword)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	vecStr := float32SliceToVectorString(queryEmb)
	query := `SELECT guidfixed, code, names, COALESCE(taxid,''), COALESCE(email,''),
			(name_embedding <=> $1::vector) as distance
		FROM debtor
		WHERE shopid = $2 AND name_embedding IS NOT NULL
		ORDER BY name_embedding <=> $1::vector
		LIMIT $3`

	rows, err := db.Query(query, vecStr, shopID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DebtorDocument
	for rows.Next() {
		var d DebtorDocument
		var namesJSON string
		var distance float64
		if err := rows.Scan(&d.GuidFixed, &d.Code, &namesJSON, &d.TaxId, &d.Email, &distance); err != nil {
			continue
		}
		if distance > 0.5 {
			continue
		}
		d.ShopID = shopID
		if namesJSON != "" {
			json.Unmarshal([]byte(namesJSON), &d.Names)
		}
		results = append(results, d)
	}
	return results, nil
}

// mergeDebtors — รวมผลลัพธ์โดยไม่ซ้ำ (ใช้ guidfixed เป็น key)
func mergeDebtors(existing, extra []DebtorDocument, limit int) []DebtorDocument {
	seen := make(map[string]bool, len(existing))
	for _, d := range existing {
		seen[d.GuidFixed] = true
	}
	merged := existing
	for _, d := range extra {
		if len(merged) >= limit {
			break
		}
		if !seen[d.GuidFixed] {
			merged = append(merged, d)
			seen[d.GuidFixed] = true
		}
	}
	return merged
}

// ==================== Search Debtors (semantic-first) ====================

// SearchDebtorsResponse — result สำหรับ search_debtors MCP tool
type SearchDebtorsResponse struct {
	Debtors     []DebtorDocument `json:"debtors"`
	Count       int              `json:"count"`
	Keyword     string           `json:"keyword"`
	SearchMode  string           `json:"search_mode"` // "vector", "regex", "combined"
	GeneratedAt time.Time        `json:"generated_at"`
}

// SearchDebtors — vector-first semantic search สำหรับ agent ใช้
// เรียก pgvector ก่อน ถ้าไม่ได้ผล fallback ไป MongoDB regex
func SearchDebtors(ctx context.Context, shopID, keyword string, limit int) (*SearchDebtorsResponse, error) {
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

	logger.Info("[SearchEntity] search_debtors shopID=%s keyword=%s limit=%d", shopID, keyword, limit)

	// 1. Try pgvector first
	vectorResults, vecErr := searchDebtorsVector(shopID, keyword, limit)
	if vecErr == nil && len(vectorResults) > 0 {
		logger.Info("[SearchEntity] search_debtors vector found %d results", len(vectorResults))
		return &SearchDebtorsResponse{
			Debtors:     vectorResults,
			Count:       len(vectorResults),
			Keyword:     keyword,
			SearchMode:  "vector",
			GeneratedAt: time.Now(),
		}, nil
	}
	if vecErr != nil {
		logger.Info("[SearchEntity] search_debtors vector fallback (err: %v) — trying regex", vecErr)
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
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}
	keywordFilter := BuildEntityKeywordFilter(keyword, []string{"code", "names.name", "taxid"})
	logger.Info("[SearchEntity] search_debtors tokens=%v", TokenizeEntityKeyword(keyword))
	filter := bson.M{"$and": []bson.M{baseFilter, keywordFilter}}

	coll := mongoClient.Database(dbName).Collection(debtorCollection)
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var debtors []DebtorDocument
	if err := cursor.All(ctx, &debtors); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}
	if debtors == nil {
		debtors = []DebtorDocument{}
	}

	logger.Info("[SearchEntity] search_debtors regex found %d results", len(debtors))
	return &SearchDebtorsResponse{
		Debtors:     debtors,
		Count:       len(debtors),
		Keyword:     keyword,
		SearchMode:  "regex",
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Debtor ====================

type CreateDebtorResponse struct {
	Success     bool           `json:"success"`
	Message     string         `json:"message"`
	Debtor      DebtorDocument `json:"debtor"`
	KafkaSync   string         `json:"kafka_sync"`
	KafkaError  string         `json:"kafka_error,omitempty"`
	GeneratedAt time.Time      `json:"generated_at"`
}

func CreateDebtor(ctx context.Context, shopID, code, namesJSON, taxid, email string, personaltype int8, creditday int, addressJSON string) (*CreateDebtorResponse, error) {
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
	coll := mongoClient.Database(dbName).Collection(debtorCollection)

	// ตรวจสอบ code ซ้ำ
	existFilter := bson.M{
		"shopid": shopID,
		"code":   code,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}
	count, err := coll.CountDocuments(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ตรวจสอบ code ซ้ำล้มเหลว: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("debtor code '%s' มีอยู่แล้วใน shop นี้", code)
	}

	var names []DebtorNameEntry
	if namesJSON != "" {
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("names is required — ต้องระบุชื่อลูกหนี้อย่างน้อย 1 ภาษา เช่น [{\"code\":\"th\",\"name\":\"ร้าน ABC\"}]")
	}

	var address DebtorAddress
	if addressJSON != "" {
		if err := json.Unmarshal([]byte(addressJSON), &address); err != nil {
			return nil, fmt.Errorf("addressforbilling JSON ไม่ถูกต้อง: %w", err)
		}
		if address.GUID == "" {
			address.GUID = uuid.New().String()
		}
	}

	now := time.Now()
	doc := DebtorDocument{
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
		return nil, fmt.Errorf("สร้างลูกหนี้ล้มเหลว: %w", err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		doc.ID = oid
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicDebtorCreated, doc); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP CreateDebtor] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
	}

	logger.Info("[MCP CreateDebtor] สร้าง code=%s สำเร็จ (shop=%s, kafka=%s)", code, shopID, kafkaSync)

	return &CreateDebtorResponse{
		Success:     true,
		Message:     fmt.Sprintf("สร้างลูกหนี้ '%s' สำเร็จ", code),
		Debtor:      doc,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Create Debtors (Bulk) ====================

type CreateDebtorsItem struct {
	Code              string            `json:"code"`
	Names             []DebtorNameEntry `json:"names"`
	PersonalType      int8              `json:"personaltype"`
	TaxId             string            `json:"taxid"`
	Email             string            `json:"email"`
	CreditDay         int               `json:"creditday"`
	AddressForBilling DebtorAddress     `json:"addressforbilling"`
}

type CreateDebtorsResponse struct {
	Success      bool             `json:"success"`
	Message      string           `json:"message"`
	Created      []DebtorDocument `json:"created"`
	Skipped      []string         `json:"skipped,omitempty"`
	CreatedCount int              `json:"created_count"`
	SkippedCount int              `json:"skipped_count"`
	KafkaSync    string           `json:"kafka_sync"`
	KafkaError   string           `json:"kafka_error,omitempty"`
	GeneratedAt  time.Time        `json:"generated_at"`
}

func CreateDebtors(ctx context.Context, shopID, debtorsJSON string) (*CreateDebtorsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if debtorsJSON == "" {
		return nil, fmt.Errorf("debtors JSON is required")
	}

	var items []CreateDebtorsItem
	if err := json.Unmarshal([]byte(debtorsJSON), &items); err != nil {
		return nil, fmt.Errorf("debtors JSON ไม่ถูกต้อง: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ต้องระบุลูกหนี้อย่างน้อย 1 รายการ")
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
	coll := mongoClient.Database(dbName).Collection(debtorCollection)

	var codes []string
	for _, item := range items {
		codes = append(codes, item.Code)
	}

	existFilter := bson.M{
		"shopid": shopID,
		"code":   bson.M{"$in": codes},
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
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
	var createdDocs []DebtorDocument
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

		doc := DebtorDocument{
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
			return nil, fmt.Errorf("สร้างลูกหนี้ล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if err := publishToKafkaBatch(ctx, TopicDebtorBulkCreated, docsToInsert); err != nil {
			kafkaSync = "failed"
			kafkaError = err.Error()
			logger.Warn("[MCP CreateDebtors] Kafka publish ล้มเหลว (แต่ MongoDB สำเร็จแล้ว): %v", err)
		}
	}

	logger.Info("[MCP CreateDebtors] สร้าง %d รายการสำเร็จ, ข้าม %d รายการ (shop=%s)", len(createdDocs), len(skipped), shopID)

	return &CreateDebtorsResponse{
		Success:      true,
		Message:      fmt.Sprintf("สร้างลูกหนี้ %d รายการสำเร็จ", len(createdDocs)),
		Created:      createdDocs,
		Skipped:      skipped,
		CreatedCount: len(createdDocs),
		SkippedCount: len(skipped),
		KafkaSync:    kafkaSync,
		KafkaError:   kafkaError,
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== Update Debtor ====================

type UpdateDebtorResponse struct {
	Success     bool           `json:"success"`
	Message     string         `json:"message"`
	Debtor      DebtorDocument `json:"debtor"`
	KafkaSync   string         `json:"kafka_sync"`
	KafkaError  string         `json:"kafka_error,omitempty"`
	GeneratedAt time.Time      `json:"generated_at"`
}

func UpdateDebtor(ctx context.Context, shopID, guidfixed, namesJSON, taxid, email string, creditday int, addressJSON string) (*UpdateDebtorResponse, error) {
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
	coll := mongoClient.Database(dbName).Collection(debtorCollection)

	filter := bson.M{
		"shopid":    shopID,
		"guidfixed": guidfixed,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	var existing DebtorDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบลูกหนี้ guidfixed='%s'", guidfixed)
	}

	updateFields := bson.M{
		"updatedat": time.Now(),
		"updatedby": "mcp-tool",
	}

	if namesJSON != "" {
		var names []DebtorNameEntry
		if err := json.Unmarshal([]byte(namesJSON), &names); err != nil {
			return nil, fmt.Errorf("names JSON ไม่ถูกต้อง: %w", err)
		}
		updateFields["names"] = names
	}
	if taxid != "" {
		updateFields["taxid"] = taxid
	}
	if email != "" {
		updateFields["email"] = email
	}
	if creditday > 0 {
		updateFields["creditday"] = creditday
	}
	if addressJSON != "" {
		var address DebtorAddress
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
		return nil, fmt.Errorf("อัปเดตลูกหนี้ล้มเหลว: %w", err)
	}

	var updated DebtorDocument
	coll.FindOne(ctx, bson.M{"_id": existing.ID}).Decode(&updated)

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicDebtorUpdated, updated); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP UpdateDebtor] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP UpdateDebtor] อัปเดต guidfixed=%s สำเร็จ (shop=%s)", guidfixed, shopID)

	return &UpdateDebtorResponse{
		Success:     true,
		Message:     fmt.Sprintf("อัปเดตลูกหนี้ '%s' สำเร็จ", existing.Code),
		Debtor:      updated,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Debtor (Soft Delete) ====================

type DeleteDebtorResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	GuidFixed   string    `json:"guidfixed"`
	Code        string    `json:"code"`
	KafkaSync   string    `json:"kafka_sync"`
	KafkaError  string    `json:"kafka_error,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

func DeleteDebtor(ctx context.Context, shopID, guidfixed string) (*DeleteDebtorResponse, error) {
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
	coll := mongoClient.Database(dbName).Collection(debtorCollection)

	filter := bson.M{
		"shopid":    shopID,
		"guidfixed": guidfixed,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}

	var existing DebtorDocument
	err := coll.FindOne(ctx, filter).Decode(&existing)
	if err != nil {
		return nil, fmt.Errorf("ไม่พบลูกหนี้ guidfixed='%s'", guidfixed)
	}

	now := time.Now()
	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
		"deletedat": now,
		"updatedby": "mcp-tool",
		"updatedat": now,
	}})
	if err != nil {
		return nil, fmt.Errorf("ลบลูกหนี้ล้มเหลว: %w", err)
	}

	kafkaSync := "published"
	kafkaError := ""
	if err := publishToKafka(ctx, TopicDebtorDeleted, existing); err != nil {
		kafkaSync = "failed"
		kafkaError = err.Error()
		logger.Warn("[MCP DeleteDebtor] Kafka publish ล้มเหลว: %v", err)
	}

	logger.Info("[MCP DeleteDebtor] ลบ code=%s สำเร็จ (shop=%s)", existing.Code, shopID)

	return &DeleteDebtorResponse{
		Success:     true,
		Message:     fmt.Sprintf("ลบลูกหนี้ '%s' สำเร็จ", existing.Code),
		GuidFixed:   guidfixed,
		Code:        existing.Code,
		KafkaSync:   kafkaSync,
		KafkaError:  kafkaError,
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Delete Debtors (Bulk Soft Delete) ====================

type DeleteDebtorsResponse struct {
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

func DeleteDebtors(ctx context.Context, shopID, guidfixedsJSON string) (*DeleteDebtorsResponse, error) {
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
	coll := mongoClient.Database(dbName).Collection(debtorCollection)

	existFilter := bson.M{
		"shopid":    shopID,
		"guidfixed": bson.M{"$in": guidfixeds},
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}
	cursor, err := coll.Find(ctx, existFilter)
	if err != nil {
		return nil, fmt.Errorf("ค้นหาลูกหนี้ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	existingGUIDs := make(map[string]bool)
	var docsToDelete []interface{}
	for cursor.Next(ctx) {
		var doc DebtorDocument
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
			"guidfixed": bson.M{"$in": deleted},
		}
		_, err = coll.UpdateMany(ctx, updateFilter, bson.M{"$set": bson.M{
			"deletedat": now,
			"updatedby": "mcp-tool",
			"updatedat": now,
		}})
		if err != nil {
			return nil, fmt.Errorf("ลบลูกหนี้ล้มเหลว: %w", err)
		}

		kafkaSync = "published"
		if len(docsToDelete) > 0 {
			if err := publishToKafkaBatch(ctx, TopicDebtorBulkDeleted, docsToDelete); err != nil {
				kafkaSync = "failed"
				kafkaError = err.Error()
				logger.Warn("[MCP DeleteDebtors] Kafka publish ล้มเหลว: %v", err)
			}
		}
	}

	logger.Info("[MCP DeleteDebtors] ลบ %d รายการสำเร็จ, ไม่พบ %d รายการ (shop=%s)", len(deleted), len(notFound), shopID)

	return &DeleteDebtorsResponse{
		Success:       true,
		Message:       fmt.Sprintf("ลบลูกหนี้ %d รายการสำเร็จ", len(deleted)),
		Deleted:       deleted,
		NotFound:      notFound,
		DeletedCount:  len(deleted),
		NotFoundCount: len(notFound),
		KafkaSync:     kafkaSync,
		KafkaError:    kafkaError,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Get Debtor Schema ====================

func GetDebtorSchema() map[string]interface{} {
	return map[string]interface{}{
		"collection":  debtorCollection,
		"description": "ลูกหนี้ (Debtor) — ข้อมูลลูกค้า/ลูกหนี้ที่ต้องชำระเงินให้ธุรกิจ",
		"fields": map[string]interface{}{
			"_id":               "ObjectID — MongoDB auto-generated ID",
			"shopid":            "string — Shop ID (tenant isolation)",
			"guidfixed":         "string — UUID สำหรับอ้างอิงภายใน",
			"code":              "string (required) — รหัสลูกหนี้ เช่น DB-001",
			"personaltype":      "int8 — ประเภท: 1=บุคคลธรรมดา, 2=นิติบุคคล",
			"names":             "array (required) — ชื่อหลายภาษา [{code:'th', name:'ร้าน XYZ'}]",
			"taxid":             "string — เลขผู้เสียภาษี",
			"email":             "string — อีเมล",
			"creditday":         "int — วงเงินเครดิต (วัน)",
			"branchnumber":      "string — เลขที่สาขา",
			"ismember":          "bool — เป็นสมาชิกหรือไม่",
			"addressforbilling": "object — ที่อยู่สำหรับออกบิล {address, countrycode, provincecode, districtcode, subdistrictcode, zipcode, phoneprimary}",
			"createdby":         "string — ผู้สร้าง",
			"createdat":         "datetime — วันที่สร้าง",
			"updatedat":         "datetime — วันที่แก้ไขล่าสุด",
			"deletedat":         "datetime — วันที่ลบ (soft delete, zero = active)",
		},
		"indexes": []string{
			"shopid + code (unique per shop)",
			"shopid + guidfixed",
		},
		"examples": []map[string]interface{}{
			{
				"code":         "DB-001",
				"personaltype": 1,
				"names":        []map[string]string{{"code": "th", "name": "ร้านค้า สมชาย"}, {"code": "en", "name": "Somchai Shop"}},
				"taxid":        "1234567890123",
				"creditday":    15,
			},
		},
	}
}

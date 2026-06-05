package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ==================== MongoDB Query (Readonly) ====================

// MongoQueryRequest คำขอ query MongoDB
type MongoQueryRequest struct {
	HoldingCode string `json:"holdingcode"`
	Database    string `json:"database"`   // ถ้าไม่ระบุจะใช้ค่า default จาก config
	Collection  string `json:"collection"` // ชื่อ collection ที่ต้องการ query
	Filter      string `json:"filter"`     // JSON filter (bson.M format)
	Limit       int    `json:"limit"`      // จำนวน documents สูงสุด (default=20, max=100)
}

// MongoQueryResponse ผลลัพธ์จาก query MongoDB
type MongoQueryResponse struct {
	Database    string                   `json:"database"`
	Collection  string                   `json:"collection"`
	Documents   []map[string]interface{} `json:"documents"`
	Count       int                      `json:"count"`
	Truncated   bool                     `json:"truncated"`
	ExecutionMs int64                    `json:"executionms"`
	GeneratedAt time.Time                `json:"generatedat"`
}

// QueryMongoDB ค้นหาข้อมูลใน MongoDB (readonly — ใช้ Find เท่านั้น)
func QueryMongoDB(ctx context.Context, holdingCode, database, collection, filterJSON string, limit int) (*MongoQueryResponse, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection is required")
	}

	// กำหนด limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// กำหนด database name
	if database == "" {
		svcConfig := serviceConfig.NewServiceConfig()
		database = svcConfig.MongodbDatabaseName()
	}

	logger.Info("[MongoDB Query] db=%s, collection=%s, filter=%s, limit=%d", database, collection, filterJSON, limit)

	// เชื่อมต่อ MongoDB
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	// สร้าง filter จาก JSON string
	var filter bson.M
	if filterJSON != "" && filterJSON != "{}" {
		if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
			return nil, fmt.Errorf("filter JSON ไม่ถูกต้อง: %w", err)
		}
	} else {
		filter = bson.M{}
	}

	// เพิ่ม holdingcode filter ถ้ามี (ป้องกันการดูข้อมูลข้าม shop)
	if holdingCode != "" {
		filter["holdingcode"] = holdingCode
	}

	startTime := time.Now()

	// Query MongoDB (readonly — Find เท่านั้น)
	coll := mongoClient.Database(database).Collection(collection)
	opts := options.Find().SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	// อ่านผลลัพธ์
	var documents []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			logger.Error("[MongoDB Query] decode error: %v", err)
			continue
		}
		documents = append(documents, doc)
	}

	if documents == nil {
		documents = []map[string]interface{}{}
	}

	return &MongoQueryResponse{
		Database:    database,
		Collection:  collection,
		Documents:   documents,
		Count:       len(documents),
		Truncated:   len(documents) >= limit,
		ExecutionMs: time.Since(startTime).Milliseconds(),
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== List MongoDB Collections ====================

// MongoListCollectionsResponse ผลลัพธ์รายการ collections
type MongoListCollectionsResponse struct {
	Database    string           `json:"database"`
	Collections []CollectionInfo `json:"collections"`
	Count       int              `json:"count"`
	GeneratedAt time.Time        `json:"generatedat"`
}

// CollectionInfo ข้อมูล collection
type CollectionInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListMongoDBCollections แสดงรายการ collections ใน database
func ListMongoDBCollections(ctx context.Context, database string) (*MongoListCollectionsResponse, error) {
	// กำหนด database name
	if database == "" {
		svcConfig := serviceConfig.NewServiceConfig()
		database = svcConfig.MongodbDatabaseName()
	}

	logger.Info("[MongoDB ListCollections] db=%s", database)

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	db := mongoClient.Database(database)
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("ดึงรายการ collections ล้มเหลว: %w", err)
	}

	var collInfos []CollectionInfo
	for _, name := range collections {
		collInfos = append(collInfos, CollectionInfo{
			Name: name,
			Type: "collection",
		})
	}

	return &MongoListCollectionsResponse{
		Database:    database,
		Collections: collInfos,
		Count:       len(collInfos),
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Aggregate MongoDB ====================

// MongoAggregateRequest คำขอ aggregation pipeline
type MongoAggregateRequest struct {
	HoldingCode string `json:"holdingcode"`
	Database    string `json:"database"`
	Collection  string `json:"collection"`
	Pipeline    string `json:"pipeline"` // JSON array ของ pipeline stages
	Limit       int    `json:"limit"`
}

// AggregateMongoDB รัน aggregation pipeline บน collection (readonly)
func AggregateMongoDB(ctx context.Context, holdingCode, database, collection, pipelineJSON string, limit int) (*MongoQueryResponse, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection is required")
	}
	if pipelineJSON == "" {
		return nil, fmt.Errorf("pipeline is required")
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	if database == "" {
		svcConfig := serviceConfig.NewServiceConfig()
		database = svcConfig.MongodbDatabaseName()
	}

	// ตรวจสอบ pipeline ว่าไม่มี write operations
	normalizedPipeline := strings.ToLower(pipelineJSON)
	forbiddenStages := []string{"$out", "$merge"}
	for _, stage := range forbiddenStages {
		if strings.Contains(normalizedPipeline, stage) {
			return nil, fmt.Errorf("pipeline มี stage ที่ไม่อนุญาต: %s (readonly mode)", stage)
		}
	}

	logger.Info("[MongoDB Aggregate] db=%s, collection=%s, limit=%d", database, collection, limit)

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}

	// Parse pipeline JSON
	var pipeline []bson.M
	if err := json.Unmarshal([]byte(pipelineJSON), &pipeline); err != nil {
		return nil, fmt.Errorf("pipeline JSON ไม่ถูกต้อง: %w", err)
	}

	// เพิ่ม $match holdingcode ถ้ามี (ป้องกันการดูข้อมูลข้าม shop)
	if holdingCode != "" {
		matchStage := bson.M{"$match": bson.M{"holdingcode": holdingCode}}
		pipeline = append([]bson.M{matchStage}, pipeline...)
	}

	// เพิ่ม $limit ท้าย pipeline
	pipeline = append(pipeline, bson.M{"$limit": limit})

	startTime := time.Now()

	coll := mongoClient.Database(database).Collection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	if documents == nil {
		documents = []map[string]interface{}{}
	}

	return &MongoQueryResponse{
		Database:    database,
		Collection:  collection,
		Documents:   documents,
		Count:       len(documents),
		Truncated:   len(documents) >= limit,
		ExecutionMs: time.Since(startTime).Milliseconds(),
		GeneratedAt: time.Now(),
	}, nil
}

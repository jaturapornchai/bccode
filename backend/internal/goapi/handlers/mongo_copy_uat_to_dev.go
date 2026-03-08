// โอนข้อมูล MongoDB จาก Production (MONGO_SERVER_IP) → Dev (MONGODB_URI / local)
// Production = MONGO_SERVER_IP (103.13.30.32) + bcaiclouddb
// Dev/Local  = MONGODB_URI (bootstrap.json → host.docker.internal:27017)

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// maskPassword - Helper function to mask password in logs
func maskPassword(uri string) string {
	start := 0
	end := len(uri)

	if idx := len("mongodb://"); len(uri) > idx {
		if at := idx; at < len(uri) {
			for i := at; i < len(uri); i++ {
				if uri[i] == '@' {
					start = idx
					end = i
					break
				}
			}
		}
	}

	if start > 0 && end > start {
		return uri[:start] + "****:****" + uri[end:]
	}
	return uri
}

// buildProductionMongoURI — สร้าง URI สำหรับเชื่อมต่อ Production MongoDB
// ลำดับความสำคัญ:
//  1. MONGODB_PRODUCTION_URI (full URI รวม SRV, TLS — จาก system_config category mongodb_production)
//  2. MONGO_SERVER_IP + MONGO_USERNAME + MONGO_PASSWORD (build URI เอง)
//
// return: URI, dbName, error
func buildProductionMongoURI() (string, string, error) {
	// วิธีที่ 1: ใช้ full URI จาก MONGODB_PRODUCTION_URI (รองรับ mongodb+srv://)
	productionURI := os.Getenv("MONGODB_PRODUCTION_URI")
	productionDB := os.Getenv("MONGODB_PRODUCTION_DB")

	if productionURI != "" {
		if productionDB == "" {
			productionDB = "dedepos"
		}
		return productionURI, productionDB, nil
	}

	// วิธีที่ 2: build URI จาก MONGO_SERVER_IP (แบบเดิม mongodb://)
	serverIP := os.Getenv("MONGO_SERVER_IP")
	serverPort := os.Getenv("MONGO_SERVER_PORT")
	username := os.Getenv("MONGO_USERNAME")
	password := os.Getenv("MONGO_PASSWORD")
	authDB := os.Getenv("MONGO_AUTH_DB")
	dbName := os.Getenv("MONGO_DB_NAME")

	if serverIP == "" {
		return "", "", fmt.Errorf("ไม่พบ config สำหรับ Production MongoDB — กรุณาตั้งค่า mongodb_production.uri ใน Setup Config")
	}
	if serverPort == "" {
		serverPort = "27017"
	}
	if username == "" {
		username = "admin"
	}
	if password == "" {
		return "", "", fmt.Errorf("MONGO_PASSWORD is required")
	}
	if authDB == "" {
		authDB = "admin"
	}
	if dbName == "" {
		dbName = "bcaiclouddb"
	}

	encodedUser := url.QueryEscape(username)
	encodedPass := url.QueryEscape(password)
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=%s",
		encodedUser, encodedPass, serverIP, serverPort, authDB)

	return uri, dbName, nil
}

// connectProductionMongo — เชื่อมต่อ Production MongoDB แล้ว return client + database
func connectProductionMongo(ctx context.Context) (*mongo.Client, *mongo.Database, string, error) {
	uri, dbName, err := buildProductionMongoURI()
	if err != nil {
		return nil, nil, "", err
	}

	logger.Info("เชื่อมต่อ Production MongoDB: %s (DB: %s)", maskPassword(uri), dbName)

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, "", fmt.Errorf("เชื่อมต่อ Production MongoDB ไม่ได้: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, "", fmt.Errorf("Ping Production MongoDB ไม่ได้: %v", err)
	}

	logger.Info("✓ เชื่อมต่อ Production MongoDB สำเร็จ")
	return client, client.Database(dbName), dbName, nil
}

// connectDevMongo — เชื่อมต่อ Dev/Local MongoDB (MONGODB_URI จาก bootstrap.json)
func connectDevMongo(ctx context.Context) (*mongo.Client, *mongo.Database, string, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, nil, "", fmt.Errorf("MONGODB_URI is required (dev MongoDB)")
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = os.Getenv("MONGO_DB_NAME")
	}
	if dbName == "" {
		dbName = "bcaiclouddevdb"
	}

	logger.Info("เชื่อมต่อ Dev MongoDB: %s (DB: %s)", maskPassword(uri), dbName)

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(100).
		SetMinPoolSize(10)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, "", fmt.Errorf("เชื่อมต่อ Dev MongoDB ไม่ได้: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, "", fmt.Errorf("Ping Dev MongoDB ไม่ได้: %v", err)
	}

	logger.Info("✓ เชื่อมต่อ Dev MongoDB สำเร็จ")
	return client, client.Database(dbName), dbName, nil
}

// CopyMongoUatToDevHandler - Copy collections จาก Production → Dev
func CopyMongoUatToDevHandler(c echo.Context) error {
	logger.Info("Copy MongoDB from Production to DEV endpoint hit")

	var payLoad models.PayLoadCopyMongoStruct
	err := json.NewDecoder(c.Request().Body).Decode(&payLoad)
	if err != nil {
		logger.Error("decoding JSON: %v", err)
		return c.String(http.StatusBadRequest, "Invalid JSON format")
	}

	if payLoad.SourceShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "source_shop_id is required",
			"code":  "MISSING_SOURCE_SHOP_ID",
		})
	}
	if payLoad.TargetShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "target_shop_id is required",
			"code":  "MISSING_TARGET_SHOP_ID",
		})
	}
	if payLoad.SourceShopID == payLoad.TargetShopID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "source_shop_id and target_shop_id cannot be the same",
			"code":  "SAME_SHOP_ID",
		})
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := copyMongoData(ctx, payLoad); err != nil {
			logger.Info("❌ Copy failed: %v", err)
		}
	}()

	return c.JSON(http.StatusAccepted, map[string]any{
		"message":        "Copy process started in background",
		"status":         "accepted",
		"code":           202,
		"source_shop_id": payLoad.SourceShopID,
		"target_shop_id": payLoad.TargetShopID,
		"note":           "Check server logs for progress",
	})
}

// copyMongoData — ทำงานจริงแบบ async
// SOURCE = Production MongoDB (MONGO_SERVER_IP)
// TARGET = Dev MongoDB (MONGODB_URI / bootstrap.json)
func copyMongoData(ctx context.Context, payLoad models.PayLoadCopyMongoStruct) error {
	logger.Info("=== Starting MongoDB Copy Process (Production → Dev) ===")

	// ✅ 1. Connect to SOURCE — Production MongoDB (MONGO_SERVER_IP)
	sourceClient, sourceDB, sourceDBName, err := connectProductionMongo(ctx)
	if err != nil {
		logger.Error("SOURCE connection: %v", err)
		return err
	}
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer dcancel()
		sourceClient.Disconnect(dctx)
	}()

	// ✅ 2. Connect to TARGET — Dev MongoDB (MONGODB_URI)
	targetClient, targetDB, targetDBName, err := connectDevMongo(ctx)
	if err != nil {
		logger.Error("TARGET connection: %v", err)
		return err
	}
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer dcancel()
		targetClient.Disconnect(dctx)
	}()

	// List all collections from source database
	collections, err := sourceDB.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		logger.Error("listing collections: %v", err)
		return err
	}

	logger.Info("Source Database (Production): %s", sourceDBName)
	logger.Info("Target Database (Dev): %s", targetDBName)
	logger.Info("Total collections: %d", len(collections))
	logger.Info("Source ShopID: %s", payLoad.SourceShopID)
	logger.Info("Target ShopID: %s", payLoad.TargetShopID)

	// Scan collections with shopid field
	logger.Info("\n=== Scanning Collections ===")
	collectionsWithShopID := []string{}
	for i, collectionName := range collections {
		logger.Info("  [%d/%d] Scanning: %s", i+1, len(collections), collectionName)

		collection := sourceDB.Collection(collectionName)
		filter := bson.M{"shopid": bson.M{"$exists": true}}

		countCtx, countCancel := context.WithTimeout(ctx, 5*time.Second)
		count, err := collection.CountDocuments(countCtx, filter)
		countCancel()

		if err != nil {
			logger.Info("      Error: %v", err)
			continue
		}

		if count > 0 {
			collectionsWithShopID = append(collectionsWithShopID, collectionName)
			logger.Info("      ✓ %d documents with shopid", count)
		}
	}

	logger.Info("\n=== Summary ===")
	logger.Info("Collections with shopid: %d/%d", len(collectionsWithShopID), len(collections))

	// Delete existing documents in target
	logger.Info("\n=== Cleaning Target ===")
	totalDeleted := int64(0)
	for _, collectionName := range collectionsWithShopID {
		collection := targetDB.Collection(collectionName)
		deleteFilter := bson.M{"shopid": payLoad.TargetShopID}

		deleteCtx, deleteCancel := context.WithTimeout(ctx, 30*time.Second)
		deleteResult, err := collection.DeleteMany(deleteCtx, deleteFilter)
		deleteCancel()

		if err != nil {
			logger.Info("❌ Error deleting in %s: %v", collectionName, err)
			continue
		}

		if deleteResult.DeletedCount > 0 {
			totalDeleted += deleteResult.DeletedCount
			logger.Info("✓ Deleted %d docs from %s", deleteResult.DeletedCount, collectionName)
		}
	}
	logger.Info("Total deleted: %d documents", totalDeleted)

	// Copy documents with proper timeout and batch processing
	logger.Info("\n=== Copying Data ===")
	totalCopied := 0
	failedCollections := []string{}

	for _, collectionName := range collectionsWithShopID {
		logger.Info("\n--- Processing: %s ---", collectionName)

		sourceCollection := sourceDB.Collection(collectionName)
		targetCollection := targetDB.Collection(collectionName)

		copyFilter := bson.M{"shopid": payLoad.SourceShopID}

		copyCtx, copyCancel := context.WithTimeout(ctx, 5*time.Minute)

		findOpts := options.Find().SetBatchSize(1000)
		cursor, err := sourceCollection.Find(copyCtx, copyFilter, findOpts)
		if err != nil {
			logger.Info("❌ Error finding in %s: %v", collectionName, err)
			failedCollections = append(failedCollections, collectionName)
			copyCancel()
			continue
		}

		var documents []interface{}
		batchSize := 0
		const maxBatchSize = 1000

		for cursor.Next(copyCtx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				logger.Info("❌ Error decoding doc in %s: %v", collectionName, err)
				continue
			}

			delete(doc, "_id")
			doc["shopid"] = payLoad.TargetShopID
			documents = append(documents, doc)
			batchSize++

			if batchSize >= maxBatchSize {
				if len(documents) > 0 {
					insertCtx, insertCancel := context.WithTimeout(copyCtx, 30*time.Second)
					_, err := targetCollection.InsertMany(insertCtx, documents, options.InsertMany().SetOrdered(false))
					insertCancel()

					if err != nil {
						logger.Info("⚠️  [%s] Batch insert had errors: %v", collectionName, err)
						successCount := insertOneByOne(copyCtx, targetCollection, documents, collectionName)
						totalCopied += successCount
					} else {
						totalCopied += len(documents)
						logger.Info("  ✓ [%s] Inserted batch of %d docs", collectionName, len(documents))
					}
					documents = []interface{}{}
					batchSize = 0
				}
			}
		}

		// Insert remaining documents
		if len(documents) > 0 {
			insertCtx, insertCancel := context.WithTimeout(copyCtx, 30*time.Second)
			_, err := targetCollection.InsertMany(insertCtx, documents, options.InsertMany().SetOrdered(false))
			insertCancel()

			if err != nil {
				logger.Info("⚠️  [%s] Final batch insert had errors: %v", collectionName, err)
				successCount := insertOneByOne(copyCtx, targetCollection, documents, collectionName)
				totalCopied += successCount
			} else {
				totalCopied += len(documents)
				logger.Info("✓ [%s] Inserted final batch of %d docs", collectionName, len(documents))
			}
		}

		cursor.Close(copyCtx)
		copyCancel()

		if err := cursor.Err(); err != nil {
			logger.Info("❌ Cursor error in %s: %v", collectionName, err)
			failedCollections = append(failedCollections, collectionName)
			continue
		}

		if len(documents) == 0 && batchSize == 0 {
			logger.Info("⚠️  No documents in %s with shopid %s", collectionName, payLoad.SourceShopID)
		}
	}

	logger.Info("\n=== Final Summary ===")
	logger.Info("Total copied: %d documents", totalCopied)
	logger.Info("Total deleted: %d documents", totalDeleted)
	logger.Info("Failed collections: %d", len(failedCollections))
	if len(failedCollections) > 0 {
		logger.Info("Failed list: %v", failedCollections)
	}
	logger.Info("=== Copy Process Completed (MongoDB only) ===")

	return nil
}

// insertOneByOne — insert ทีละ document เมื่อ batch insert fail (เช่น duplicate key)
func insertOneByOne(ctx context.Context, collection *mongo.Collection, documents []interface{}, collectionName string) int {
	successCount := 0
	skipCount := 0
	for i, doc := range documents {
		insertOneCtx, insertOneCancel := context.WithTimeout(ctx, 10*time.Second)
		_, errOne := collection.InsertOne(insertOneCtx, doc)
		insertOneCancel()
		if errOne != nil {
			skipCount++
			if i%100 == 0 || i == len(documents)-1 {
				docMap := doc.(bson.M)
				docID := "unknown"
				if id, ok := docMap["_id"]; ok {
					docID = fmt.Sprintf("%v", id)
				}
				logger.Info("    [%s][%d] ❌ Doc _id=%s: %v", collectionName, i+1, docID, errOne)
			}
		} else {
			successCount++
		}
	}
	logger.Info("  ✓ [%s] Inserted %d docs, skipped %d duplicates", collectionName, successCount, skipCount)
	return successCount
}

// PreviewCopyMongoHandler — ดูรายละเอียด collections ที่จะโอน ก่อนกดยืนยัน
// ดึงจาก Production MongoDB (MONGO_SERVER_IP)
func PreviewCopyMongoHandler(c echo.Context) error {
	logger.Info("Preview MongoDB Copy endpoint hit")

	var payLoad models.PayLoadCopyMongoStruct
	err := json.NewDecoder(c.Request().Body).Decode(&payLoad)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	if payLoad.SourceShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "source_shop_id is required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// เชื่อมต่อ Production MongoDB (MONGO_SERVER_IP)
	sourceClient, sourceDB, _, err := connectProductionMongo(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dcancel()
		sourceClient.Disconnect(dctx)
	}()

	// ดึง list collections
	collections, err := sourceDB.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("ดึง collection list ไม่ได้: %v", err)})
	}

	// นับ documents ที่มี shopid = source ในแต่ละ collection
	type CollectionInfo struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}

	var results []CollectionInfo
	totalDocs := int64(0)

	for _, collName := range collections {
		coll := sourceDB.Collection(collName)
		filter := bson.M{"shopid": payLoad.SourceShopID}

		countCtx, countCancel := context.WithTimeout(ctx, 5*time.Second)
		count, err := coll.CountDocuments(countCtx, filter)
		countCancel()

		if err != nil {
			logger.Info("Preview: error counting %s: %v", collName, err)
			continue
		}

		if count > 0 {
			results = append(results, CollectionInfo{Name: collName, Count: count})
			totalDocs += count
		}
	}

	logger.Info("Preview: %d collections, %d total documents for shop %s", len(results), totalDocs, payLoad.SourceShopID)

	return c.JSON(http.StatusOK, map[string]any{
		"success":          true,
		"source_shop_id":   payLoad.SourceShopID,
		"collections":      results,
		"collection_count": len(results),
		"total_documents":  totalDocs,
	})
}

// ListSourceShopsHandler — ดึง list shops จาก Production MongoDB (MONGO_SERVER_IP)
// สำหรับ Flutter เลือก shop ต้นทางก่อน copy
func ListSourceShopsHandler(c echo.Context) error {
	logger.Info("List source shops from Production MongoDB (MONGO_SERVER_IP)")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// เชื่อมต่อ Production MongoDB (MONGO_SERVER_IP)
	sourceClient, sourceDB, _, err := connectProductionMongo(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dcancel()
		sourceClient.Disconnect(dctx)
	}()

	collection := sourceDB.Collection("shops")

	// Query shops — เอาเฉพาะ field ที่ต้องการ
	projection := bson.M{
		"guidfixed":  1,
		"names":      1,
		"name1":      1,
		"branchcode": 1,
	}
	findOpts := options.Find().SetProjection(projection).SetSort(bson.M{"name1": 1})

	// ไม่เอา shop ที่ถูกลบ
	filter := bson.M{
		"deletedat": bson.M{"$exists": false},
	}

	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		logger.Error("querying shops: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("ดึงข้อมูลร้านค้าไม่ได้: %v", err),
		})
	}
	defer cursor.Close(ctx)

	var shops []bson.M
	if err := cursor.All(ctx, &shops); err != nil {
		logger.Error("decoding shops: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("อ่านข้อมูลร้านค้าไม่ได้: %v", err),
		})
	}

	logger.Info("Found %d shops from Production MongoDB", len(shops))

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    shops,
		"total":   len(shops),
	})
}

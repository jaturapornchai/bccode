// โอนข้อมูล MongoDB จาก UAT/PRO → DEV เฉพาะใน DEV mode
// Source UAT = MONGODB_UAT_URI + MONGODB_UAT_DB
// Source PRO = MONGODB_PRO_URI + MONGODB_PRO_DB หรือ MONGODB_PRODUCTION_URI + MONGODB_PRODUCTION_DB
// Target DEV = MONGODB_DEV_URI + MONGODB_DEV_DB หรือ legacy MONGODB_URI + MONGODB_DB/MONGO_DB_NAME

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoCopyEnvDev = "dev"
	mongoCopyEnvUAT = "uat"
	mongoCopyEnvPRO = "pro"
)

// maskPassword - Helper function to mask password in logs
func maskPassword(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.User == nil {
		return uri
	}
	parsed.User = url.UserPassword("****", "****")
	return parsed.String()
}

func normalizeMongoCopyEnv(value string, fallback string) (string, error) {
	env := strings.ToLower(strings.TrimSpace(value))
	if env == "" {
		env = fallback
	}
	switch env {
	case "dev", "development", "local":
		return mongoCopyEnvDev, nil
	case "uat", "test", "testing", "staging":
		return mongoCopyEnvUAT, nil
	case "pro", "prod", "production":
		return mongoCopyEnvPRO, nil
	default:
		return "", fmt.Errorf("invalid environment: %s", value)
	}
}

func isGoAPIDevelopmentMode() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("MODE")))
	return mode == "" || mode == "dev" || mode == "development" || mode == "local"
}

func validateMongoCopyEnvironment(payLoad models.PayLoadCopyMongoStruct) (string, string, error) {
	sourceEnv, err := normalizeMongoCopyEnv(payLoad.SourceEnvironment, mongoCopyEnvUAT)
	if err != nil {
		return "", "", err
	}
	targetEnv, err := normalizeMongoCopyEnv(payLoad.TargetEnvironment, mongoCopyEnvDev)
	if err != nil {
		return "", "", err
	}
	if !isGoAPIDevelopmentMode() {
		return "", "", fmt.Errorf("copy tool is allowed only in DEV mode")
	}
	if targetEnv != mongoCopyEnvDev {
		return "", "", fmt.Errorf("copy target must be DEV")
	}
	if sourceEnv != mongoCopyEnvUAT && sourceEnv != mongoCopyEnvPRO {
		return "", "", fmt.Errorf("copy source must be UAT or PRO")
	}
	return sourceEnv, targetEnv, nil
}

func buildUATMongoURI() (string, string, error) {
	uri := strings.TrimSpace(os.Getenv("MONGODB_UAT_URI"))
	dbName := strings.TrimSpace(os.Getenv("MONGODB_UAT_DB"))
	if dbName == "" {
		dbName = strings.TrimSpace(os.Getenv("MONGODB_UAT_DATABASE"))
	}
	if uri == "" {
		return "", "", fmt.Errorf("MONGODB_UAT_URI is required for UAT source")
	}
	if dbName == "" {
		return "", "", fmt.Errorf("MONGODB_UAT_DB is required for UAT source")
	}
	return uri, dbName, nil
}

func connectSourceMongo(ctx context.Context, sourceEnv string) (*mongo.Client, *mongo.Database, string, error) {
	switch sourceEnv {
	case mongoCopyEnvUAT:
		return connectUATMongo(ctx)
	case mongoCopyEnvPRO:
		return connectProductionMongo(ctx)
	default:
		return nil, nil, "", fmt.Errorf("unsupported source environment: %s", sourceEnv)
	}
}

func connectUATMongo(ctx context.Context) (*mongo.Client, *mongo.Database, string, error) {
	uri, dbName, err := buildUATMongoURI()
	if err != nil {
		return nil, nil, "", err
	}

	logger.Info("[DEV][CopyMongo] เชื่อมต่อ UAT MongoDB: %s (DB: %s)", maskPassword(uri), dbName)

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, "", fmt.Errorf("เชื่อมต่อ UAT MongoDB ไม่ได้: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, "", fmt.Errorf("Ping UAT MongoDB ไม่ได้: %v", err)
	}

	logger.Info("[DEV][CopyMongo] ✓ เชื่อมต่อ UAT MongoDB สำเร็จ")
	return client, client.Database(dbName), dbName, nil
}

// buildProductionMongoURI — สร้าง URI สำหรับเชื่อมต่อ Production MongoDB
// return: URI, dbName, error
func buildProductionMongoURI() (string, string, error) {
	productionURI := envFirst("MONGODB_PRO_URI", "MONGODB_PRODUCTION_URI")
	productionDB := envFirst("MONGODB_PRO_DB", "MONGODB_PRO_DATABASE", "MONGODB_PRODUCTION_DB")

	if productionURI == "" {
		return "", "", fmt.Errorf("MONGODB_PRO_URI is required for PRO source")
	}
	if productionDB == "" {
		return "", "", fmt.Errorf("MONGODB_PRO_DB is required for PRO source")
	}
	return productionURI, productionDB, nil
}

func buildDevMongoURI() (string, string, error) {
	uri := envFirst("MONGODB_DEV_URI", "MONGODB_URI")
	if uri == "" {
		return "", "", fmt.Errorf("MONGODB_DEV_URI is required (dev MongoDB)")
	}

	dbName := envFirst("MONGODB_DEV_DB", "MONGODB_DEV_DATABASE", "MONGODB_DB", "MONGO_DB_NAME")
	if dbName == "" {
		return "", "", fmt.Errorf("MONGODB_DEV_DB is required (dev MongoDB database)")
	}

	return uri, dbName, nil
}

// connectProductionMongo — เชื่อมต่อ Production MongoDB แล้ว return client + database
func connectProductionMongo(ctx context.Context) (*mongo.Client, *mongo.Database, string, error) {
	uri, dbName, err := buildProductionMongoURI()
	if err != nil {
		return nil, nil, "", err
	}

	logger.Info("[DEV][CopyMongo] เชื่อมต่อ PRO MongoDB: %s (DB: %s)", maskPassword(uri), dbName)

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, "", fmt.Errorf("เชื่อมต่อ PRO MongoDB ไม่ได้: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, "", fmt.Errorf("Ping PRO MongoDB ไม่ได้: %v", err)
	}

	logger.Info("[DEV][CopyMongo] ✓ เชื่อมต่อ PRO MongoDB สำเร็จ")
	return client, client.Database(dbName), dbName, nil
}

// connectDevMongo — เชื่อมต่อ Dev MongoDB (MONGODB_DEV_URI หรือ legacy MONGODB_URI)
func connectDevMongo(ctx context.Context) (*mongo.Client, *mongo.Database, string, error) {
	uri, dbName, err := buildDevMongoURI()
	if err != nil {
		return nil, nil, "", err
	}

	logger.Info("[DEV][CopyMongo] เชื่อมต่อ DEV MongoDB: %s (DB: %s)", maskPassword(uri), dbName)

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(100).
		SetMinPoolSize(10)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, "", fmt.Errorf("เชื่อมต่อ DEV MongoDB ไม่ได้: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, nil, "", fmt.Errorf("Ping DEV MongoDB ไม่ได้: %v", err)
	}

	logger.Info("[DEV][CopyMongo] ✓ เชื่อมต่อ DEV MongoDB สำเร็จ")
	return client, client.Database(dbName), dbName, nil
}

// CopyMongoUatToDevHandler - Copy collections จาก UAT/PRO → DEV เฉพาะใน DEV mode
func CopyMongoUatToDevHandler(c echo.Context) error {
	logger.Info("[DEV][CopyMongo] Copy MongoDB endpoint hit")

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
	sourceEnv, targetEnv, err := validateMongoCopyEnvironment(payLoad)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "DEV mode") {
			status = http.StatusForbidden
		}
		logger.Warn("[CopyMongo] blocked: %v", err)
		return c.JSON(status, map[string]string{
			"error": err.Error(),
			"code":  "ENVIRONMENT_COPY_BLOCKED",
		})
	}
	payLoad.SourceEnvironment = sourceEnv
	payLoad.TargetEnvironment = targetEnv

	logger.Info("[DEV][CopyMongo] accepted source_env=%s target_env=%s source_shop=%s target_shop=%s",
		sourceEnv, targetEnv, payLoad.SourceShopID, payLoad.TargetShopID)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := copyMongoData(ctx, payLoad); err != nil {
			logger.Info("❌ Copy failed: %v", err)
		}
	}()

	return c.JSON(http.StatusAccepted, map[string]any{
		"message":            "Copy process started in background",
		"status":             "accepted",
		"code":               202,
		"source_environment": sourceEnv,
		"target_environment": targetEnv,
		"source_shop_id":     payLoad.SourceShopID,
		"target_shop_id":     payLoad.TargetShopID,
		"note":               "Check server logs for progress",
	})
}

// copyMongoData — ทำงานจริงแบบ async
// SOURCE = UAT/PRO MongoDB
// TARGET = DEV MongoDB (MONGODB_DEV_URI / bootstrap.json)
func copyMongoData(ctx context.Context, payLoad models.PayLoadCopyMongoStruct) error {
	logger.Info("[DEV][CopyMongo] === Starting MongoDB Copy Process (%s → %s) ===", payLoad.SourceEnvironment, payLoad.TargetEnvironment)

	// ✅ 1. Connect to SOURCE — UAT/PRO MongoDB
	sourceClient, sourceDB, sourceDBName, err := connectSourceMongo(ctx, payLoad.SourceEnvironment)
	if err != nil {
		logger.Error("SOURCE connection: %v", err)
		return err
	}
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer dcancel()
		sourceClient.Disconnect(dctx)
	}()

	// ✅ 2. Connect to TARGET — Dev MongoDB (MONGODB_DEV_URI หรือ legacy MONGODB_URI)
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

	logger.Info("[DEV][CopyMongo] Source Database (%s): %s", payLoad.SourceEnvironment, sourceDBName)
	logger.Info("[DEV][CopyMongo] Target Database (%s): %s", payLoad.TargetEnvironment, targetDBName)
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
// ดึงจาก UAT/PRO MongoDB
func PreviewCopyMongoHandler(c echo.Context) error {
	logger.Info("[DEV][CopyMongo] Preview MongoDB Copy endpoint hit")

	var payLoad models.PayLoadCopyMongoStruct
	err := json.NewDecoder(c.Request().Body).Decode(&payLoad)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	if payLoad.SourceShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "source_shop_id is required"})
	}
	sourceEnv, targetEnv, err := validateMongoCopyEnvironment(payLoad)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "DEV mode") {
			status = http.StatusForbidden
		}
		logger.Warn("[CopyMongo] preview blocked: %v", err)
		return c.JSON(status, map[string]string{
			"error": err.Error(),
			"code":  "ENVIRONMENT_COPY_BLOCKED",
		})
	}
	payLoad.SourceEnvironment = sourceEnv
	payLoad.TargetEnvironment = targetEnv

	logger.Info("[DEV][CopyMongo] preview source_env=%s target_env=%s source_shop=%s",
		sourceEnv, targetEnv, payLoad.SourceShopID)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// เชื่อมต่อ UAT/PRO MongoDB
	sourceClient, sourceDB, _, err := connectSourceMongo(ctx, sourceEnv)
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
		"success":            true,
		"source_environment": sourceEnv,
		"target_environment": targetEnv,
		"source_shop_id":     payLoad.SourceShopID,
		"collections":        results,
		"collection_count":   len(results),
		"total_documents":    totalDocs,
	})
}

// ListSourceShopsHandler — ดึง list shops จาก UAT/PRO MongoDB
// สำหรับ Flutter เลือก shop ต้นทางก่อน copy
func ListSourceShopsHandler(c echo.Context) error {
	sourceEnv, err := normalizeMongoCopyEnv(c.QueryParam("source_environment"), mongoCopyEnvUAT)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if sourceEnv != mongoCopyEnvUAT && sourceEnv != mongoCopyEnvPRO {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "source_environment must be uat or pro"})
	}
	if !isGoAPIDevelopmentMode() {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "copy tool is allowed only in DEV mode",
			"code":  "ENVIRONMENT_COPY_BLOCKED",
		})
	}
	logger.Info("[DEV][CopyMongo] List source shops from %s MongoDB", sourceEnv)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// เชื่อมต่อ UAT/PRO MongoDB
	sourceClient, sourceDB, _, err := connectSourceMongo(ctx, sourceEnv)
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
		"guid_fixed": 1,
		"names":      1,
		"name1":      1,
		"branchcode": 1,
	}
	findOpts := options.Find().SetProjection(projection).SetSort(bson.M{"name1": 1})

	// ไม่เอา shop ที่ถูกลบ
	filter := bson.M{
		"deleted_at": bson.M{"$exists": false},
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

	logger.Info("[DEV][CopyMongo] Found %d shops from %s MongoDB", len(shops), sourceEnv)

	return c.JSON(http.StatusOK, map[string]any{
		"success":            true,
		"source_environment": sourceEnv,
		"data":               shops,
		"total":              len(shops),
	})
}

package datahistory

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	OldCollectionName = "datahistory"    // Collection เดิมที่ใช้ก่อนแยก
	CollectionPrefix  = "datahistory_"   // Prefix สำหรับ collection ใหม่ตาม screen_type
)

// MigrateHistoryHandler - API endpoint สำหรับย้ายข้อมูลจาก collection เดิมไปใหม่
// GET /api/datahistory/migrate
func MigrateHistoryHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client, err := myglobal.MongoConnect()
	if err != nil {
		logger.Error("[Migration] Failed to connect to MongoDB: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to connect to database",
		})
	}

	db := client.Database(DatabaseName)
	oldColl := db.Collection(OldCollectionName)

	// นับจำนวน documents ใน collection เดิม
	count, err := oldColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Error("[Migration] Failed to count documents: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to count documents",
		})
	}

	if count == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "No documents to migrate - old collection is empty",
			"summary": map[string]interface{}{
				"total":    0,
				"migrated": 0,
				"failed":   0,
			},
		})
	}

	logger.Info("[Migration] Found %d documents in old collection '%s'", count, OldCollectionName)

	// ดึงข้อมูลทั้งหมดจาก collection เดิม
	cursor, err := oldColl.Find(ctx, bson.M{})
	if err != nil {
		logger.Error("[Migration] Failed to query old collection: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to query old collection",
		})
	}
	defer cursor.Close(ctx)

	// สรุปการย้าย
	summary := make(map[string]int)
	totalMigrated := 0
	totalFailed := 0

	// วนลูปย้ายข้อมูล
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			logger.Error("[Migration] Failed to decode document: %v", err)
			totalFailed++
			continue
		}

		// ดึง screen_type
		screenType, ok := doc["screen_type"].(string)
		if !ok || screenType == "" {
			logger.Error("[Migration] Document missing screen_type: %v", doc["_id"])
			totalFailed++
			continue
		}

		// Insert ไปยัง collection ใหม่ (ใช้ prefix + screen_type)
		newCollectionName := CollectionPrefix + screenType
		newColl := db.Collection(newCollectionName)

		_, err = newColl.InsertOne(ctx, doc)
		if err != nil {
			// อาจเป็น duplicate key - ถ้าเคยรันไปแล้ว
			if mongo.IsDuplicateKeyError(err) {
				logger.Debug("[Migration] Duplicate document, skipping: %v", doc["_id"])
			} else {
				logger.Error("[Migration] Failed to insert to %s_%s: %v", CollectionPrefix, screenType, err)
				totalFailed++
				continue
			}
		}

		summary[screenType]++
		totalMigrated++

		if totalMigrated%100 == 0 {
			logger.Info("[Migration] Progress: %d/%d documents migrated...", totalMigrated, count)
		}
	}

	if err := cursor.Err(); err != nil {
		logger.Error("[Migration] Cursor error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Cursor error: %v", err),
		})
	}

	logger.Info("[Migration] Completed - Migrated: %d, Failed: %d", totalMigrated, totalFailed)

	// สร้าง detailed summary
	detailedSummary := make(map[string]int)
	for screenType, cnt := range summary {
		collectionName := CollectionPrefix + screenType
		detailedSummary[collectionName] = cnt
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Migration completed",
		"summary": map[string]interface{}{
			"total":       count,
			"migrated":    totalMigrated,
			"failed":      totalFailed,
			"collections": detailedSummary,
		},
		"note": fmt.Sprintf("Old collection '%s' still exists. Delete manually if needed: db.%s.drop()", OldCollectionName, OldCollectionName),
	})
}

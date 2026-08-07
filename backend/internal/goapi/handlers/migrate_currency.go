package handlers

import (
	"context"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

// MigrateCurrencyColumnsHandler adds currency columns to doc table
// GET /api/migrate/currency?holdingcode=xxx
func MigrateCurrencyColumnsHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "holdingcode is required",
		})
	}

	logger.Info("[MIGRATE-CURRENCY] Starting migration for shop: %s", holdingCode)

	// Get PostgreSQL connection
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("[MIGRATE-CURRENCY] Failed to connect to PostgreSQL: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "database connection failed",
		})
	}

	ctx := context.Background()

	// Add currency columns
	migrationSQL := `
-- Add multi-currency columns to doc table
ALTER TABLE doc
ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'THB',
ADD COLUMN IF NOT EXISTS currency_symbol VARCHAR(5) DEFAULT '฿',
ADD COLUMN IF NOT EXISTS doc_currency VARCHAR(10) DEFAULT 'THB',
ADD COLUMN IF NOT EXISTS doc_currency_symbol VARCHAR(5) DEFAULT '฿',
ADD COLUMN IF NOT EXISTS exchange_rate NUMERIC(18,6) DEFAULT 1,
ADD COLUMN IF NOT EXISTS totalvalue_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS totaldiscount_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS totalvatvalue_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS totalbeforevat_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS totalaftervat_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS totalamount_doc NUMERIC(18,2) DEFAULT 0;
`

	logger.Info("[MIGRATE-CURRENCY] Running migration: Adding currency columns...")
	_, err = db.ExecContext(ctx, migrationSQL)
	if err != nil {
		logger.Error("[MIGRATE-CURRENCY] Migration failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": fmt.Sprintf("migration failed: %v", err),
		})
	}

	logger.Success("[MIGRATE-CURRENCY] doc table migration completed!")

	// เพิ่ม multi-currency columns ให้ docdetail table
	docdetailMigrationSQL := `
ALTER TABLE docdetail
ADD COLUMN IF NOT EXISTS price_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS sumamount_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS discountamount_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS priceexcludevat_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS sumamountexcludevat_doc NUMERIC(18,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS totalvaluevat_doc NUMERIC(18,2) DEFAULT 0;
`

	logger.Info("[MIGRATE-CURRENCY] Running migration: Adding currency columns to docdetail...")
	_, err = db.ExecContext(ctx, docdetailMigrationSQL)
	if err != nil {
		logger.Error("[MIGRATE-CURRENCY] docdetail migration failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": fmt.Sprintf("docdetail migration failed: %v", err),
		})
	}

	logger.Success("[MIGRATE-CURRENCY] docdetail migration completed!")

	// เพิ่ม isdelete column สำหรับ soft delete
	softDeleteSQL := `
ALTER TABLE doc ADD COLUMN IF NOT EXISTS isdelete BOOLEAN DEFAULT FALSE;
`
	logger.Info("[MIGRATE-CURRENCY] Running migration: Adding isdelete column...")
	_, err = db.ExecContext(ctx, softDeleteSQL)
	if err != nil {
		logger.Warn("[MIGRATE-CURRENCY] isdelete migration warning: %v", err)
	} else {
		logger.Success("[MIGRATE-CURRENCY] isdelete column added!")
	}

	// Create indexes
	indexSQL := `
CREATE INDEX IF NOT EXISTS idxdoccurrency ON doc(currency);
CREATE INDEX IF NOT EXISTS idxdoccurrencyshop ON doc(currency, holdingcode);
CREATE INDEX IF NOT EXISTS idxdocisdelete ON doc(isdelete);
CREATE INDEX IF NOT EXISTS idxdocdocnotransflagisdelete ON doc(docno, transflag, isdelete);
`

	logger.Info("[MIGRATE-CURRENCY] Creating indexes...")
	_, err = db.ExecContext(ctx, indexSQL)
	if err != nil {
		logger.Warn("[MIGRATE-CURRENCY] Index creation warning: %v", err)
	} else {
		logger.Success("[MIGRATE-CURRENCY] Indexes created successfully!")
	}

	// Verify columns exist
	var columnCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'doc'
		AND column_name IN ('currency', 'currency_symbol', 'exchange_rate', 'totalamount_doc',
		                     'totalbeforevat_doc', 'totalaftervat_doc', 'totalvalue_doc',
		                     'totaldiscount_doc', 'totalvatvalue_doc')
	`).Scan(&columnCount)

	if err != nil {
		logger.Error("[MIGRATE-CURRENCY] Failed to verify columns: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "verification failed",
		})
	}

	logger.Success("[MIGRATE-CURRENCY] Verified: Found %d currency columns in doc table", columnCount)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       "Currency columns migration completed",
		"columns_added": columnCount,
		"expected":      9,
	})
}

// BackfillCurrencyDataHandler อ่าน currency data จาก MongoDB แล้วอัพเดท PostgreSQL
// สำหรับเอกสารเก่าที่ถูก Kafka consume ก่อนแก้ TagName: "json"
// GET /api/migrate/currency-backfill?holdingcode=xxx
func BackfillCurrencyDataHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "holdingcode is required",
		})
	}

	logger.Info("[BACKFILL-CURRENCY] เริ่ม backfill currency data สำหรับ shop: %s", holdingCode)

	// เชื่อมต่อ MongoDB
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "MongoDB connection failed",
		})
	}

	cfg := config.NewServiceConfig()
	mongoDBName := cfg.MongodbDatabaseName()
	mongoDB := mongoClient.Database(mongoDBName)

	// เชื่อมต่อ PostgreSQL
	pgDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("[BACKFILL-CURRENCY] PostgreSQL connection failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "PostgreSQL connection failed",
		})
	}

	ctx := context.Background()

	// อ่านเอกสารจาก MongoDB ที่มี doc_currency และ exchangerate
	// รวมทุก collection ที่เป็น transaction
	collections := []string{
		"transactionPurchaseOrder",
		"transactionPurchase",
		"transactionPurchaseReturn",
		"transactionSale",
		"transactionSaleReturn",
		"transactionSaleOrder",
	}

	totalUpdated := 0
	totalSkipped := 0

	for _, collName := range collections {
		collection := mongoDB.Collection(collName)

		// ค้นหาเอกสารที่มี doc_currency ไม่ว่าง และเป็น shop ที่ต้องการ
		filter := bson.M{
			"holdingcode":  holdingCode,
			"doc_currency": bson.M{"$exists": true, "$ne": ""},
		}

		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			logger.Error("[BACKFILL-CURRENCY] ค้นหาใน %s ล้มเหลว: %v", collName, err)
			continue
		}

		var docs []bson.M
		if err := cursor.All(ctx, &docs); err != nil {
			logger.Error("[BACKFILL-CURRENCY] อ่านข้อมูลจาก %s ล้มเหลว: %v", collName, err)
			continue
		}

		for _, doc := range docs {
			docNo, _ := doc["docno"].(string)
			if docNo == "" {
				continue
			}

			docCurrency, _ := doc["doc_currency"].(string)
			docCurrencySymbol, _ := doc["doc_currencysymbol"].(string)

			var exchangeRate float64
			switch v := doc["exchange_rate"].(type) {
			case float64:
				exchangeRate = v
			case int32:
				exchangeRate = float64(v)
			case int64:
				exchangeRate = float64(v)
			}

			var totalAmountDoc float64
			switch v := doc["totalamount_doc"].(type) {
			case float64:
				totalAmountDoc = v
			case int32:
				totalAmountDoc = float64(v)
			case int64:
				totalAmountDoc = float64(v)
			}

			// อัพเดท PostgreSQL
			updateSQL := `UPDATE doc SET doc_currency = $1, doc_currency_symbol = $2, exchange_rate = $3, totalamount_doc = $4 WHERE docno = $5`
			result, err := pgDB.ExecContext(ctx, updateSQL, docCurrency, docCurrencySymbol, exchangeRate, totalAmountDoc, docNo)
			if err != nil {
				logger.Error("[BACKFILL-CURRENCY] อัพเดท %s ล้มเหลว: %v", docNo, err)
				continue
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				totalUpdated++
				logger.Info("[BACKFILL-CURRENCY] อัพเดท %s สำเร็จ: doc_currency=%s, rate=%.4f, amount_doc=%.2f", docNo, docCurrency, exchangeRate, totalAmountDoc)
			} else {
				totalSkipped++
			}
		}

		if len(docs) > 0 {
			logger.Info("[BACKFILL-CURRENCY] %s: พบ %d เอกสาร", collName, len(docs))
		}
	}

	logger.Success("[BACKFILL-CURRENCY] เสร็จสิ้น: อัพเดท %d เอกสาร, ข้าม %d เอกสาร", totalUpdated, totalSkipped)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Backfill completed: updated %d, skipped %d", totalUpdated, totalSkipped),
		"total_updated": totalUpdated,
		"total_skipped": totalSkipped,
	})
}

// MigrateClickHouseSoftDeleteHandler เพิ่ม isdelete column ใน ClickHouse doc table
// GET /api/migrate/clickhouse-softdelete
func MigrateClickHouseSoftDeleteHandler(c echo.Context) error {
	logger.Info("[MIGRATE-SOFTDELETE] เริ่ม migration: เพิ่ม isdelete column ใน ClickHouse")

	chClient, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Error("[MIGRATE-SOFTDELETE] เชื่อมต่อ ClickHouse ล้มเหลว: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "ClickHouse connection failed",
		})
	}

	ctx := context.Background()

	// เพิ่ม isdelete column ใน doc table
	alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS isdelete Bool DEFAULT 0", myclickhouse.TableName("doc"))
	err = chClient.Exec(ctx, alterSQL)
	if err != nil {
		logger.Error("[MIGRATE-SOFTDELETE] เพิ่ม isdelete ใน doc ล้มเหลว: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": fmt.Sprintf("ALTER TABLE doc failed: %v", err),
		})
	}
	logger.Success("[MIGRATE-SOFTDELETE] เพิ่ม isdelete column ใน %s สำเร็จ", myclickhouse.TableName("doc"))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "ClickHouse soft delete migration completed",
	})
}

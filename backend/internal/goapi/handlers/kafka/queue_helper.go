package kafka

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypostgres"
	"time"
)

// AddDocToProcessQueue - เพิ่มเอกสารเข้า PostgreSQL queue แยกตาม holdingCode
// แทนที่ Redis queue ด้วย PostgreSQL
func AddDocToProcessQueue(holdingCode, docNo string, transFlag interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// แปลง transFlag เป็น string
	transFlagStr := convertTransFlagToString(transFlag)

	// ดึง PostgreSQL connection พร้อม retry
	var db *sql.DB
	var err error
	maxRetries := 3
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = myglobal.GetGlobalDatabaseConnection()
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			logger.Warn("Failed to get PostgreSQL connection (attempt %d/%d): %v, retrying...", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // exponential backoff
		}
	}

	if err != nil {
		logger.Error("Failed to get PostgreSQL connection after %d attempts: %v", maxRetries, err)
		return fmt.Errorf("failed to get database connection after retries: %w", err)
	}

	// สร้าง QueueItem
	queueItem := mypostgres.QueueItem{
		HoldingCode: holdingCode,
		DocNo:       docNo,
		TransFlag:   transFlagStr,
		CreatedAt:   time.Now(),
		RetryCount:  0,
	}

	// เพิ่มเข้า PostgreSQL queue
	qm := mypostgres.NewQueueManager(db)
	err = qm.AddToQueue(ctx, queueItem)
	if err != nil {
		logger.Error("Failed to add to PostgreSQL queue (shop=%s, doc=%s): %v",
			holdingCode, docNo, err)
		return err
	}

	logger.Debug("Added to PostgreSQL queue successfully: shop=%s, docno=%s, transflag=%s",
		holdingCode, docNo, transFlagStr)

	return nil
}

// AddDocToProcessQueueWithPriority - เพิ่มเอกสารเข้า queue พร้อม priority
func AddDocToProcessQueueWithPriority(holdingCode, docNo string, transFlag interface{}, priority int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	transFlagStr := convertTransFlagToString(transFlag)

	// ดึง PostgreSQL connection
	db, err := myglobal.GetGlobalDatabaseConnection()
	if err != nil {
		logger.Error("Failed to get PostgreSQL connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// สร้าง QueueItem (PostgreSQL จะจัดเรียงตาม createdat โดยอัตโนมัติ)
	queueItem := mypostgres.QueueItem{
		HoldingCode: holdingCode,
		DocNo:       docNo,
		TransFlag:   transFlagStr,
		CreatedAt:   time.Now(),
		RetryCount:  0,
	}

	qm := mypostgres.NewQueueManager(db)
	err = qm.AddToQueue(ctx, queueItem)
	if err != nil {
		logger.Error("Failed to add to PostgreSQL queue with priority: %v", err)
		return err
	}

	logger.Debug("Added to PostgreSQL queue with priority=%d: shop=%s, docno=%s",
		priority, holdingCode, docNo)

	return nil
}

// convertTransFlagToString - แปลง transFlag จาก int หรือ string เป็น string
func convertTransFlagToString(transFlag interface{}) string {
	switch v := transFlag.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

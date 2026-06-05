package mypostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"strings"
	"time"
)

// QueueItem - โครงสร้างข้อมูล queue item (เหมือน myredis.QueueItem)
type QueueItem struct {
	ID          int64     `json:"id,omitempty"`
	HoldingCode string    `json:"holdingcode"`
	DocNo       string    `json:"docno"`
	TransFlag   string    `json:"transflag"`
	RetryCount  int       `json:"retrycount"`
	CreatedAt   time.Time `json:"createdat"`
	Error       string    `json:"error,omitempty"`
}

// QueueManager - จัดการ queue ใน PostgreSQL
type QueueManager struct {
	db *sql.DB
}

// NewQueueManager - สร้าง QueueManager ใหม่
func NewQueueManager(db *sql.DB) *QueueManager {
	return &QueueManager{db: db}
}

// isDatabaseClosedError ตรวจสอบว่า error เกิดจากการปิด connection หรือไม่
func isDatabaseClosedError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, sql.ErrConnDone) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is closed") || strings.Contains(msg, "connection is already closed")
}

// runWithReconnect รันคำสั่งกับฐานข้อมูล ถ้า connection ปิดจะรีเฟรชและลองใหม่อีกครั้ง
func (qm *QueueManager) runWithReconnect(operation func(db *sql.DB) error) error {
	err := operation(qm.db)
	if !isDatabaseClosedError(err) {
		return err
	}

	logger.Warn("PostgreSQL queue connection closed. Attempting automatic refresh...")
	newDB, refreshErr := myglobal.RefreshGlobalDatabaseConnection()
	if refreshErr != nil {
		return errors.Join(err, fmt.Errorf("failed to refresh queue database connection: %w", refreshErr))
	}

	qm.db = newDB
	return operation(qm.db)
}

// AddToQueue - เพิ่มงานเข้า queue (แทนที่ Redis LPUSH)
func (qm *QueueManager) AddToQueue(ctx context.Context, item QueueItem) error {
	query := `
		INSERT INTO queues (holdingcode, doc_no, trans_flag, retry_count, createdat, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	err := qm.runWithReconnect(func(db *sql.DB) error {
		_, execErr := db.ExecContext(ctx, query,
			item.HoldingCode,
			item.DocNo,
			item.TransFlag,
			item.RetryCount,
			item.CreatedAt,
		)
		return execErr
	})

	if err != nil {
		logger.Error("Failed to add to queue: %v", err)
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	logger.Debug("Added to queue: shop=%s, doc=%s, trans=%s", item.HoldingCode, item.DocNo, item.TransFlag)
	return nil
}

// PopFromQueue - ดึงงานจาก queue (แทนที่ Redis LPOP)
// ใช้ FOR UPDATE SKIP LOCKED เพื่อป้องกัน race condition
func (qm *QueueManager) PopFromQueue(ctx context.Context, holdingCode string) (*QueueItem, error) {
	query := `
		UPDATE queues
		SET status = 'processing',
			processed_at = NOW()
		WHERE id = (
			SELECT id
			FROM queues
			WHERE holdingcode = $1
			  AND status = 'pending'
			ORDER BY createdat ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, holdingcode, doc_no, trans_flag, retry_count, createdat
	`

	var item QueueItem
	err := qm.runWithReconnect(func(db *sql.DB) error {
		return db.QueryRowContext(ctx, query, holdingCode).Scan(
			&item.ID,
			&item.HoldingCode,
			&item.DocNo,
			&item.TransFlag,
			&item.RetryCount,
			&item.CreatedAt,
		)
	})

	if err == sql.ErrNoRows {
		return nil, nil // ไม่มีงานใน queue
	}

	if err != nil {
		logger.Error("Failed to pop from queue: %v", err)
		return nil, fmt.Errorf("failed to pop from queue: %w", err)
	}

	logger.Debug("Popped from queue: shop=%s, doc=%s", item.HoldingCode, item.DocNo)
	return &item, nil
}

// RequeueItem - ใส่งานกลับเข้า queue (แทนที่ Redis LPUSH)
func (qm *QueueManager) RequeueItem(ctx context.Context, item QueueItem) error {
	query := `
		UPDATE queues
		SET status = 'pending',
			retry_count = $1,
			updatedat = NOW(),
			processed_at = NULL
		WHERE id = $2
	`

	err := qm.runWithReconnect(func(db *sql.DB) error {
		_, execErr := db.ExecContext(ctx, query, item.RetryCount, item.ID)
		return execErr
	})
	if err != nil {
		logger.Error("Failed to requeue item: %v", err)
		return fmt.Errorf("failed to requeue item: %w", err)
	}

	logger.Debug("Requeued item: shop=%s, doc=%s, retry=%d", item.HoldingCode, item.DocNo, item.RetryCount)
	return nil
}

// AddToDeadLetterQueue - ส่งงานล้มเหลวไปยัง DLQ (แทนที่ Redis LPUSH dead_letter_queue)
func (qm *QueueManager) AddToDeadLetterQueue(ctx context.Context, item QueueItem, errorMessage string) error {
	insertQuery := `
		INSERT INTO dead_letter_queue (holdingcode, doc_no, trans_flag, retry_count, createdat, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	updateQuery := `
		UPDATE queues
		SET status = 'failed',
			error_message = $1,
			updatedat = NOW()
		WHERE id = $2
	`

	operation := func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		success := false
		defer func() {
			if !success {
				_ = tx.Rollback()
			}
		}()

		if _, err := tx.ExecContext(ctx, insertQuery,
			item.HoldingCode,
			item.DocNo,
			item.TransFlag,
			item.RetryCount,
			item.CreatedAt,
			errorMessage,
		); err != nil {
			logger.Error("Failed to add to dead letter queue: %v", err)
			return fmt.Errorf("failed to add to DLQ: %w", err)
		}

		if _, err := tx.ExecContext(ctx, updateQuery, errorMessage, item.ID); err != nil {
			logger.Error("Failed to update queue item status: %v", err)
			return fmt.Errorf("failed to update queue item: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		success = true
		return nil
	}

	if err := qm.runWithReconnect(operation); err != nil {
		return err
	}

	logger.Warn("Added to dead letter queue: shop=%s, doc=%s, error=%s", item.HoldingCode, item.DocNo, errorMessage)
	return nil
}

// GetActiveShops - ดึงรายชื่อ shop ที่มีงาน (แทนที่ Redis KEYS queue:*)
func (qm *QueueManager) GetActiveShops(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT holdingcode
		FROM queues
		WHERE status = 'pending'
		ORDER BY holdingcode
	`

	var rows *sql.Rows
	err := qm.runWithReconnect(func(db *sql.DB) error {
		var queryErr error
		rows, queryErr = db.QueryContext(ctx, query)
		return queryErr
	})
	if err != nil {
		logger.Error("Failed to get active shops: %v", err)
		return nil, fmt.Errorf("failed to get active shops: %w", err)
	}
	defer rows.Close()

	var shops []string
	for rows.Next() {
		var holdingCode string
		if err := rows.Scan(&holdingCode); err != nil {
			logger.Error("Failed to scan holdingcode: %v", err)
			continue
		}
		shops = append(shops, holdingCode)
	}

	return shops, nil
}

// GetQueueLength - ดึงความยาว queue ของ shop (แทนที่ Redis LLEN)
func (qm *QueueManager) GetQueueLength(ctx context.Context, holdingCode string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM queues
		WHERE holdingcode = $1 AND status = 'pending'
	`

	var length int64
	err := qm.runWithReconnect(func(db *sql.DB) error {
		return db.QueryRowContext(ctx, query, holdingCode).Scan(&length)
	})
	if err != nil {
		logger.Error("Failed to get queue length: %v", err)
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}

	return length, nil
}

// QueueStats - สถิติของ queue
type QueueStats struct {
	HoldingCode     string     `json:"holdingcode"`
	PendingCount    int64      `json:"pendingcount"`
	ProcessingCount int64      `json:"processingcount"`
	CompletedCount  int64      `json:"completedcount"`
	FailedCount     int64      `json:"failedcount"`
	OldestItem      *time.Time `json:"oldestitem,omitempty"`
}

// GetQueueStats - ดึงสถิติ queue ของ shop (แทนที่ Redis custom stats)
func (qm *QueueManager) GetQueueStats(ctx context.Context, holdingCode string) (*QueueStats, error) {
	query := `
		SELECT
			holdingcode,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
			COUNT(*) FILTER (WHERE status = 'processing') as processing_count,
			COUNT(*) FILTER (WHERE status = 'completed') as completed_count,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_count,
			MIN(createdat) FILTER (WHERE status = 'pending') as oldest_item
		FROM queues
		WHERE holdingcode = $1
		GROUP BY holdingcode
	`

	var stats QueueStats
	err := qm.runWithReconnect(func(db *sql.DB) error {
		return db.QueryRowContext(ctx, query, holdingCode).Scan(
			&stats.HoldingCode,
			&stats.PendingCount,
			&stats.ProcessingCount,
			&stats.CompletedCount,
			&stats.FailedCount,
			&stats.OldestItem,
		)
	})

	if err == sql.ErrNoRows {
		return &QueueStats{
			HoldingCode:     holdingCode,
			PendingCount:    0,
			ProcessingCount: 0,
			CompletedCount:  0,
			FailedCount:     0,
		}, nil
	}

	if err != nil {
		logger.Error("Failed to get queue stats: %v", err)
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	return &stats, nil
}

// QueueSummary - สรุปข้อมูล queue ทั้งหมด
type QueueSummary struct {
	TotalShops      int                      `json:"totalshops"`
	TotalPending    int64                    `json:"totalpending"`
	TotalProcessing int64                    `json:"totalprocessing"`
	TotalFailed     int64                    `json:"totalfailed"`
	ShopStats       []map[string]interface{} `json:"shopstats"`
}

// GetQueueSummary - ดึงสรุปข้อมูล queue ทั้งหมด (แทนที่ Redis custom stats)
func (qm *QueueManager) GetQueueSummary(ctx context.Context) (*QueueSummary, error) {
	query := `
		SELECT
			holdingcode,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
			COUNT(*) FILTER (WHERE status = 'processing') as processing_count,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_count
		FROM queues
		WHERE status IN ('pending', 'processing', 'failed')
		GROUP BY holdingcode
		ORDER BY pending_count DESC
	`

	var rows *sql.Rows
	err := qm.runWithReconnect(func(db *sql.DB) error {
		var queryErr error
		rows, queryErr = db.QueryContext(ctx, query)
		return queryErr
	})
	if err != nil {
		logger.Error("Failed to get queue summary: %v", err)
		return nil, fmt.Errorf("failed to get queue summary: %w", err)
	}
	defer rows.Close()

	summary := &QueueSummary{
		ShopStats: make([]map[string]interface{}, 0),
	}

	for rows.Next() {
		var holdingCode string
		var pending, processing, failed int64

		if err := rows.Scan(&holdingCode, &pending, &processing, &failed); err != nil {
			logger.Error("Failed to scan shop stats: %v", err)
			continue
		}

		summary.TotalShops++
		summary.TotalPending += pending
		summary.TotalProcessing += processing
		summary.TotalFailed += failed

		summary.ShopStats = append(summary.ShopStats, map[string]interface{}{
			"holdingcode":      holdingCode,
			"pending_count":    pending,
			"processing_count": processing,
			"failed_count":     failed,
		})
	}

	return summary, nil
}

// MarkAsCompleted - ทำเครื่องหมายงานว่าเสร็จสมบูรณ์
func (qm *QueueManager) MarkAsCompleted(ctx context.Context, itemId int64) error {
	query := `
		UPDATE queues
		SET status = 'completed',
			updatedat = NOW()
		WHERE id = $1
	`

	err := qm.runWithReconnect(func(db *sql.DB) error {
		_, execErr := db.ExecContext(ctx, query, itemId)
		return execErr
	})
	if err != nil {
		logger.Error("Failed to mark as completed: %v", err)
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	return nil
}

// CleanupOldCompletedItems - ทำความสะอาดงานที่เสร็จแล้วเกินกำหนด
func (qm *QueueManager) CleanupOldCompletedItems(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM queues
		WHERE status = 'completed'
		  AND updatedat < $1
	`

	cutoffTime := time.Now().Add(-olderThan)
	var result sql.Result
	err := qm.runWithReconnect(func(db *sql.DB) error {
		var execErr error
		result, execErr = db.ExecContext(ctx, query, cutoffTime)
		return execErr
	})
	if err != nil {
		logger.Error("Failed to cleanup old items: %v", err)
		return 0, fmt.Errorf("failed to cleanup: %w", err)
	}

	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		logger.Info("Cleaned up %d old completed items", deleted)
	}

	return deleted, nil
}

// Backward compatibility functions to match myredis API

// AddToQueueCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.AddToQueue
func AddToQueue(ctx context.Context, db *sql.DB, item QueueItem) error {
	qm := NewQueueManager(db)
	return qm.AddToQueue(ctx, item)
}

// PopFromQueueCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.PopFromQueue
func PopFromQueue(ctx context.Context, db *sql.DB, holdingCode string) (*QueueItem, error) {
	qm := NewQueueManager(db)
	return qm.PopFromQueue(ctx, holdingCode)
}

// RequeueItemCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.RequeueItem
func RequeueItem(ctx context.Context, db *sql.DB, item QueueItem) error {
	qm := NewQueueManager(db)
	return qm.RequeueItem(ctx, item)
}

// AddToDeadLetterQueueCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.AddToDeadLetterQueue
func AddToDeadLetterQueue(ctx context.Context, db *sql.DB, item QueueItem, errorMessage string) error {
	qm := NewQueueManager(db)
	return qm.AddToDeadLetterQueue(ctx, item, errorMessage)
}

// GetActiveShopsCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.GetActiveShops
func GetActiveShops(ctx context.Context, db *sql.DB) ([]string, error) {
	qm := NewQueueManager(db)
	return qm.GetActiveShops(ctx)
}

// GetQueueLengthCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.GetQueueLength
func GetQueueLength(ctx context.Context, db *sql.DB, holdingCode string) (int64, error) {
	qm := NewQueueManager(db)
	return qm.GetQueueLength(ctx, holdingCode)
}

// GetQueueStatsCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.GetQueueStats
func GetQueueStats(ctx context.Context, db *sql.DB, holdingCode string) (map[string]interface{}, error) {
	qm := NewQueueManager(db)
	stats, err := qm.GetQueueStats(ctx, holdingCode)
	if err != nil {
		return nil, err
	}

	// แปลงเป็น map เพื่อ compatibility
	statsMap := make(map[string]interface{})
	data, _ := json.Marshal(stats)
	json.Unmarshal(data, &statsMap)

	return statsMap, nil
}

// GetQueueSummaryCompat - ฟังก์ชันเพื่อความเข้ากันได้กับ myredis.GetQueueSummary
func GetQueueSummary(ctx context.Context, db *sql.DB) (map[string]interface{}, error) {
	qm := NewQueueManager(db)
	summary, err := qm.GetQueueSummary(ctx)
	if err != nil {
		return nil, err
	}

	// แปลงเป็น map เพื่อ compatibility
	summaryMap := make(map[string]interface{})
	data, _ := json.Marshal(summary)
	json.Unmarshal(data, &summaryMap)

	return summaryMap, nil
}

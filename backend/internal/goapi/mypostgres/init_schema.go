package mypostgres

import (
	"context"
	"database/sql"
	"smlcloudplatform/internal/goapi/logger"
	"time"
)

// InitQueueSchema - สร้างตารางและ functions สำหรับ Queue System
func InitQueueSchema(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("กำลังตรวจสอบและสร้าง Queue Schema...")

	// SQL สำหรับสร้างตารางและ functions
	schemaSQL := `
-- Table: queues
CREATE TABLE IF NOT EXISTS queues (
    id BIGSERIAL PRIMARY KEY,
    holdingcode VARCHAR(100) NOT NULL,
    doc_no VARCHAR(100) NOT NULL,
    trans_flag VARCHAR(10) NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    processed_at TIMESTAMP NULL,
    error_message TEXT NULL
);

-- Table: dead_letter_queue
CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id BIGSERIAL PRIMARY KEY,
    holdingcode VARCHAR(100) NOT NULL,
    doc_no VARCHAR(100) NOT NULL,
    trans_flag VARCHAR(10) NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    createdat TIMESTAMP NOT NULL,
    failed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    error_message TEXT NOT NULL
);

-- Table: distributed_locks
CREATE TABLE IF NOT EXISTS distributed_locks (
    lock_key VARCHAR(500) PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    acquired_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_queues_shop_status ON queues(holdingcode, status);
CREATE INDEX IF NOT EXISTS idx_queues_createdat ON queues(createdat);
CREATE INDEX IF NOT EXISTS idx_queues_status ON queues(status);
CREATE INDEX IF NOT EXISTS idx_dlq_holdingcode ON dead_letter_queue(holdingcode);
CREATE INDEX IF NOT EXISTS idx_dlq_failed_at ON dead_letter_queue(failed_at);
CREATE INDEX IF NOT EXISTS idx_locks_expires_at ON distributed_locks(expires_at);
CREATE INDEX IF NOT EXISTS idx_queues_pop ON queues(holdingcode, createdat) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_queues_active_shops ON queues(holdingcode) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_queues_processing ON queues(holdingcode, processed_at) WHERE status = 'processing';

-- Function: อัปเดต updatedat อัตโนมัติ
CREATE OR REPLACE FUNCTION update_updatedat_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updatedat = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: อัปเดต updatedat
DROP TRIGGER IF EXISTS update_queues_updatedat ON queues;
CREATE TRIGGER update_queues_updatedat
    BEFORE UPDATE ON queues
    FOR EACH ROW
    EXECUTE FUNCTION update_updatedat_column();

-- Function: ทำความสะอาด expired locks
CREATE OR REPLACE FUNCTION cleanup_expired_locks()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM distributed_locks
    WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function: ดึงงานจาก queue
CREATE OR REPLACE FUNCTION pop_from_queue(p_holdingcode VARCHAR)
RETURNS TABLE(
    id BIGINT,
    holdingcode VARCHAR,
    doc_no VARCHAR,
    trans_flag VARCHAR,
    retry_count INTEGER,
    createdat TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    UPDATE queues
    SET status = 'processing',
        processed_at = NOW()
    WHERE queues.id = (
        SELECT queues.id
        FROM queues
        WHERE queues.holdingcode = p_holdingcode
          AND queues.status = 'pending'
        ORDER BY queues.createdat ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED
    )
    RETURNING
        queues.id,
        queues.holdingcode,
        queues.doc_no,
        queues.trans_flag,
        queues.retry_count,
        queues.createdat;
END;
$$ LANGUAGE plpgsql;

-- Function: ดึง active shops
CREATE OR REPLACE FUNCTION get_active_shops()
RETURNS TABLE(holdingcode VARCHAR, queue_count BIGINT) AS $$
BEGIN
    RETURN QUERY
    SELECT q.holdingcode, COUNT(*) as queue_count
    FROM queues q
    WHERE q.status = 'pending'
    GROUP BY q.holdingcode
    ORDER BY queue_count DESC;
END;
$$ LANGUAGE plpgsql;
`

	// Execute schema SQL
	_, err := db.ExecContext(ctx, schemaSQL)
	if err != nil {
		logger.Error("Failed to initialize queue schema: %v", err)
		return err
	}

	logger.Success("✅ Queue Schema initialized successfully")
	return nil
}

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
    docno VARCHAR(100) NOT NULL,
    transflag VARCHAR(10) NOT NULL,
    retrycount INTEGER NOT NULL DEFAULT 0,
    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    processedat TIMESTAMP NULL,
    errormessage TEXT NULL
);

-- Table: deadletterqueue
CREATE TABLE IF NOT EXISTS deadletterqueue (
    id BIGSERIAL PRIMARY KEY,
    holdingcode VARCHAR(100) NOT NULL,
    docno VARCHAR(100) NOT NULL,
    transflag VARCHAR(10) NOT NULL,
    retrycount INTEGER NOT NULL DEFAULT 0,
    createdat TIMESTAMP NOT NULL,
    failedat TIMESTAMP NOT NULL DEFAULT NOW(),
    errormessage TEXT NOT NULL
);

-- Table: distributedlocks
CREATE TABLE IF NOT EXISTS distributedlocks (
    lockkey VARCHAR(500) PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    acquiredat TIMESTAMP NOT NULL DEFAULT NOW(),
    expiresat TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idxqueuesshopstatus ON queues(holdingcode, status);
CREATE INDEX IF NOT EXISTS idxqueuescreatedat ON queues(createdat);
CREATE INDEX IF NOT EXISTS idxqueuesstatus ON queues(status);
CREATE INDEX IF NOT EXISTS idxdlqholdingcode ON deadletterqueue(holdingcode);
CREATE INDEX IF NOT EXISTS idxdlqfailedat ON deadletterqueue(failedat);
CREATE INDEX IF NOT EXISTS idxlocksexpiresat ON distributedlocks(expiresat);
CREATE INDEX IF NOT EXISTS idxqueuespop ON queues(holdingcode, createdat) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idxqueuesactiveshops ON queues(holdingcode) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idxqueuesprocessing ON queues(holdingcode, processedat) WHERE status = 'processing';

-- Function: อัปเดต updatedat อัตโนมัติ
CREATE OR REPLACE FUNCTION updateupdatedatcolumn()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updatedat = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: อัปเดต updatedat
DROP TRIGGER IF EXISTS updatequeuesupdatedat ON queues;
CREATE TRIGGER updatequeuesupdatedat
    BEFORE UPDATE ON queues
    FOR EACH ROW
    EXECUTE FUNCTION updateupdatedatcolumn();

-- Function: ทำความสะอาด expired locks
CREATE OR REPLACE FUNCTION cleanupexpiredlocks()
RETURNS INTEGER AS $$
DECLARE
    deletedcount INTEGER;
BEGIN
    DELETE FROM distributedlocks
    WHERE expiresat < NOW();
    GET DIAGNOSTICS deletedcount = ROW_COUNT;
    RETURN deletedcount;
END;
$$ LANGUAGE plpgsql;

-- Function: ดึงงานจาก queue
CREATE OR REPLACE FUNCTION popfromqueue(pholdingcode VARCHAR)
RETURNS TABLE(
    id BIGINT,
    holdingcode VARCHAR,
    docno VARCHAR,
    transflag VARCHAR,
    retrycount INTEGER,
    createdat TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    UPDATE queues
    SET status = 'processing',
        processedat = NOW()
    WHERE queues.id = (
        SELECT queues.id
        FROM queues
        WHERE queues.holdingcode = pholdingcode
          AND queues.status = 'pending'
        ORDER BY queues.createdat ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED
    )
    RETURNING
        queues.id,
        queues.holdingcode,
        queues.docno,
        queues.transflag,
        queues.retrycount,
        queues.createdat;
END;
$$ LANGUAGE plpgsql;

-- Function: ดึง active shops
CREATE OR REPLACE FUNCTION getactiveshops()
RETURNS TABLE(holdingcode VARCHAR, queuecount BIGINT) AS $$
BEGIN
    RETURN QUERY
    SELECT q.holdingcode, COUNT(*) as queuecount
    FROM queues q
    WHERE q.status = 'pending'
    GROUP BY q.holdingcode
    ORDER BY queuecount DESC;
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

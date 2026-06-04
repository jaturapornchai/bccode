-- ========================================
-- PostgreSQL Schema for Queue Management
-- แทนที่ Redis Queue System
-- ========================================

-- Table: queues
-- แทนที่ Redis List (queue:{holdingCode})
CREATE TABLE IF NOT EXISTS queues (
    id BIGSERIAL PRIMARY KEY,
    holding_code VARCHAR(100) NOT NULL,
    doc_no VARCHAR(100) NOT NULL,
    trans_flag VARCHAR(10) NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    processed_at TIMESTAMP NULL,
    error_message TEXT NULL,

    -- Indexes สำหรับ performance
    INDEX idx_queues_shop_status (holding_code, status),
    INDEX idx_queues_created_at (created_at),
    INDEX idx_queues_status (status)
);

-- Table: dead_letter_queue
-- แทนที่ Redis List (dead_letter_queue)
CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id BIGSERIAL PRIMARY KEY,
    holding_code VARCHAR(100) NOT NULL,
    doc_no VARCHAR(100) NOT NULL,
    trans_flag VARCHAR(10) NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    failed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    error_message TEXT NOT NULL,

    -- Indexes
    INDEX idx_dlq_holding_code (holding_code),
    INDEX idx_dlq_failed_at (failed_at)
);

-- Table: distributed_locks
-- แทนที่ Redis String (stock:calc:{holdingCode}:{itemCode})
CREATE TABLE IF NOT EXISTS distributed_locks (
    lock_key VARCHAR(500) PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    acquired_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,

    -- Index สำหรับ cleanup expired locks
    INDEX idx_locks_expires_at (expires_at)
);

-- ========================================
-- Functions & Triggers
-- ========================================

-- Trigger: อัปเดต updated_at อัตโนมัติ
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_queues_updated_at
    BEFORE UPDATE ON queues
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

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

-- Function: ดึงงานจาก queue (Pop with status update)
CREATE OR REPLACE FUNCTION pop_from_queue(p_holding_code VARCHAR)
RETURNS TABLE(
    id BIGINT,
    holding_code VARCHAR,
    doc_no VARCHAR,
    trans_flag VARCHAR,
    retry_count INTEGER,
    created_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    UPDATE queues
    SET status = 'processing',
        processed_at = NOW()
    WHERE queues.id = (
        SELECT queues.id
        FROM queues
        WHERE queues.holding_code = p_holding_code
          AND queues.status = 'pending'
        ORDER BY queues.created_at ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED
    )
    RETURNING
        queues.id,
        queues.holding_code,
        queues.doc_no,
        queues.trans_flag,
        queues.retry_count,
        queues.created_at;
END;
$$ LANGUAGE plpgsql;

-- Function: ดึง active shops (shops ที่มี pending queue)
CREATE OR REPLACE FUNCTION get_active_shops()
RETURNS TABLE(holding_code VARCHAR, queue_count BIGINT) AS $$
BEGIN
    RETURN QUERY
    SELECT q.holding_code, COUNT(*) as queue_count
    FROM queues q
    WHERE q.status = 'pending'
    GROUP BY q.holding_code
    ORDER BY queue_count DESC;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- Indexes สำหรับ Performance
-- ========================================

-- Index สำหรับ Pop operation
CREATE INDEX IF NOT EXISTS idx_queues_pop
    ON queues(holding_code, created_at)
    WHERE status = 'pending';

-- Index สำหรับ GetActiveShops
CREATE INDEX IF NOT EXISTS idx_queues_active_shops
    ON queues(holding_code)
    WHERE status = 'pending';

-- Partial Index สำหรับ processing items
CREATE INDEX IF NOT EXISTS idx_queues_processing
    ON queues(holding_code, processed_at)
    WHERE status = 'processing';

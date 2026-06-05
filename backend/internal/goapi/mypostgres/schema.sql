-- ========================================
-- PostgreSQL Schema for Queue Management
-- แทนที่ Redis Queue System
-- ========================================

-- Table: queues
-- แทนที่ Redis List (queue:{holdingCode})
CREATE TABLE IF NOT EXISTS queues (
    id BIGSERIAL PRIMARY KEY,
    holdingcode VARCHAR(100) NOT NULL,
    docno VARCHAR(100) NOT NULL,
    transflag VARCHAR(10) NOT NULL,
    retrycount INTEGER NOT NULL DEFAULT 0,
    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    processedat TIMESTAMP NULL,
    errormessage TEXT NULL,

    -- Indexes สำหรับ performance
    INDEX idxqueuesshopstatus (holdingcode, status),
    INDEX idxqueuescreatedat (createdat),
    INDEX idxqueuesstatus (status)
);

-- Table: deadletterqueue
-- แทนที่ Redis List (deadletterqueue)
CREATE TABLE IF NOT EXISTS deadletterqueue (
    id BIGSERIAL PRIMARY KEY,
    holdingcode VARCHAR(100) NOT NULL,
    docno VARCHAR(100) NOT NULL,
    transflag VARCHAR(10) NOT NULL,
    retrycount INTEGER NOT NULL DEFAULT 0,
    createdat TIMESTAMP NOT NULL,
    failedat TIMESTAMP NOT NULL DEFAULT NOW(),
    errormessage TEXT NOT NULL,

    -- Indexes
    INDEX idxdlqholdingcode (holdingcode),
    INDEX idxdlqfailedat (failedat)
);

-- Table: distributedlocks
-- แทนที่ Redis String (stock:calc:{holdingCode}:{itemCode})
CREATE TABLE IF NOT EXISTS distributedlocks (
    lockkey VARCHAR(500) PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    acquiredat TIMESTAMP NOT NULL DEFAULT NOW(),
    expiresat TIMESTAMP NOT NULL,

    -- Index สำหรับ cleanup expired locks
    INDEX idxlocksexpiresat (expiresat)
);

-- ========================================
-- Functions & Triggers
-- ========================================

-- Trigger: อัปเดต updatedat อัตโนมัติ
CREATE OR REPLACE FUNCTION updateupdatedatcolumn()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updatedat = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

-- Function: ดึงงานจาก queue (Pop with status update)
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

-- Function: ดึง active shops (shops ที่มี pending queue)
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

-- ========================================
-- Indexes สำหรับ Performance
-- ========================================

-- Index สำหรับ Pop operation
CREATE INDEX IF NOT EXISTS idxqueuespop
    ON queues(holdingcode, createdat)
    WHERE status = 'pending';

-- Index สำหรับ GetActiveShops
CREATE INDEX IF NOT EXISTS idxqueuesactiveshops
    ON queues(holdingcode)
    WHERE status = 'pending';

-- Partial Index สำหรับ processing items
CREATE INDEX IF NOT EXISTS idxqueuesprocessing
    ON queues(holdingcode, processedat)
    WHERE status = 'processing';

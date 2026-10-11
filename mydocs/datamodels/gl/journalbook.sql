-- ----------------------------------------------------------------------------
-- เล่มสมุดรายวัน (Journal Books)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS journal_books (
    company_code    VARCHAR(20) NOT NULL,
    code            VARCHAR(15) NOT NULL,              -- เช่น 'JV', 'PV', 'RV', 'SV', 'UV'
    name_th         VARCHAR(100) NOT NULL,
    name_en         VARCHAR(100),
    
    -- 1=ทั่วไป (JV), 2=จ่าย (PV), 3=รับ (RV), 4=ขาย (SV), 5=ซื้อ (UV), 6=ยอดยกมา
    book_type       SMALLINT NOT NULL CHECK (book_type BETWEEN 1 AND 6),
    
    -- 0=ปิดใช้งาน, 1=เปิดใช้งาน
    is_active       SMALLINT NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    
    PRIMARY KEY (company_code, code)
);

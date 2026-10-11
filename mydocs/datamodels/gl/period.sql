-- ----------------------------------------------------------------------------
-- งวดบัญชี (Fiscal Periods) สามารถกำหนดงวดได้เรื่อยๆ ไม่จำกัดปีบัญชี
-- year_no: ปีบัญชี เช่น 2026 (ต้องเป็นปี ค.ศ. ไม่ใช่ พ.ศ.)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS fiscal_periods (
    company_code    VARCHAR(20) NOT NULL,
    period_year     INT NOT NULL,              -- เช่น 2026
    period_no       SMALLINT NOT NULL,         -- งวดที่ 1 ถึง 12
    begin_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    
    -- สถานะงวด: 1=OPEN (เปิดบันทึกได้), 2=LOCKED (ล็อคชั่วคราว), 3=CLOSED (ปิดงวดถาวร)
    status          SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 3),
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    
    PRIMARY KEY (company_code, period_year, period_no)
);

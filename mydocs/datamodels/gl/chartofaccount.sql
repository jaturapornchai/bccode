-- สร้างตารางผังบัญชี (Chart of Accounts)
-- ผังบัญชีมีรูปแบบเป็นชั้น (level)
-- ประเภทมีบัญชีคุม และบัญชีปรกติ
-- ถ้าผังไหนเป็นหัว (มีตัวลูก) จะไม่สามารถบันทึกตัวเลขได้
-- ผังที่บันทึกตัวเลขได้ จะต้องเป็นตัวลูก ไม่มีหางมาต่อ
-- การเรียง ให้เรียงตามรหัสบัญชี
-- ทุกผังบัญชี จะต้องมี field หัวบัญชี ถ้าไม่มีให้เป็น null

CREATE TABLE IF NOT EXISTS chart_of_accounts (
    company_code         VARCHAR(20) NOT NULL,
    account_code         VARCHAR(20) NOT NULL,
    name_th              VARCHAR(150) NOT NULL,
    name_en              VARCHAR(150),
    
    -- 1=สินทรัพย์, 2=หนี้สิน, 3=ทุน, 4=รายได้, 5=ค่าใช้จ่าย
    account_group        SMALLINT NOT NULL CHECK (account_group BETWEEN 1 AND 5),
    
    -- 1=DEBIT, 2=CREDIT
    normal_balance       SMALLINT NOT NULL CHECK (normal_balance IN (1, 2)),
    
    parent_account_code  VARCHAR(20),
    account_level        SMALLINT NOT NULL CHECK (account_level BETWEEN 1 AND 12),
    
    -- 0=ไม่ใช่, 1=ใช่ (ใช้ตัวเลขแทน Boolean)
    is_header            SMALLINT NOT NULL DEFAULT 0 CHECK (is_header IN (0, 1)),
    allow_posting        SMALLINT NOT NULL DEFAULT 1 CHECK (allow_posting IN (0, 1)),
    is_cash              SMALLINT NOT NULL DEFAULT 0 CHECK (is_cash IN (0, 1)),
    is_active            SMALLINT NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),

    -- มิติวิเคราะห์ 6 มิติ (0=ไม่แยก, 1=แยกตามมิติ) ตาม Champ Parity
    depart_separate      SMALLINT NOT NULL DEFAULT 0 CHECK (depart_separate IN (0, 1)),
    project_separate     SMALLINT NOT NULL DEFAULT 0 CHECK (project_separate IN (0, 1)),
    allocate_separate    SMALLINT NOT NULL DEFAULT 0 CHECK (allocate_separate IN (0, 1)),
    job_separate         SMALLINT NOT NULL DEFAULT 0 CHECK (job_separate IN (0, 1)),
    part_separate        SMALLINT NOT NULL DEFAULT 0 CHECK (part_separate IN (0, 1)),
    side_separate        SMALLINT NOT NULL DEFAULT 0 CHECK (side_separate IN (0, 1)),
    
    -- วันที่และเวลาใช้ UTC+0 เสมอ
    created_at           TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),

    -- กำหนด Primary Key
    PRIMARY KEY (company_code, account_code),

    -- Foreign Key หัวบัญชีต้องอยู่ภายในบริษัทเดียวกัน
    FOREIGN KEY (company_code, parent_account_code) 
        REFERENCES chart_of_accounts(company_code, account_code) 
        ON DELETE RESTRICT,

    -- กฎ: ถ้าเป็นบัญชีคุม (is_header = 1) ห้ามลงบันทึกบัญชีเด็ดขาด (allow_posting = 1 ไม่ได้)
    CONSTRAINT chk_coa_header_posting CHECK (
        NOT (is_header = 1 AND allow_posting = 1)
    )
);

-- Index สำหรับค้นหาและจัดเรียงตามบริษัท
CREATE INDEX IF NOT EXISTS idx_coa_company_parent ON chart_of_accounts(company_code, parent_account_code);
CREATE INDEX IF NOT EXISTS idx_coa_company_group ON chart_of_accounts(company_code, account_group, account_code);

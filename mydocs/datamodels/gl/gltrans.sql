-- ----------------------------------------------------------------------------
-- หัวเอกสารห้องบัญชี: รับข้อมูล/คีย์ตรง ตรวจสอบ และปรับปรุงบัญชี ไม่ใช่ workflow หน้าบ้าน
-- ร่าง CREATE สำหรับฐานใหม่ ไม่ใช่ migration ของฐานที่มีอยู่
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS journal_entries (
    company_code    VARCHAR(20) NOT NULL,
    doc_no          VARCHAR(30) NOT NULL,              -- เช่น 'JV-202609-0001'
    doc_date        DATE NOT NULL,                     -- วันที่บันทึกบัญชี
    book_code       VARCHAR(15) NOT NULL,
    branch_code     VARCHAR(20) NOT NULL,              -- สาขาที่บันทึกรายการ (เลือกตั้งแต่ Login)
    
    ref_doc_no      VARCHAR(50),                       -- เลขที่เอกสารอ้างอิง (เช่น ใบกำกับภาษี/ใบเสร็จ)
    ref_date        DATE,                              -- วันที่เอกสารอ้างอิง
    
    description     VARCHAR(500),                      -- คำอธิบายรายการหัวเอกสาร
    
    -- แก้เป็น NUMERIC ไม่ระบุ scale เพื่อรองรับทศนิยมที่ผู้ใช้กำหนดเอง
    total_amount    NUMERIC NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    
    -- ประเภทแหล่งข้อมูลเท่านั้น ไม่บังคับติดตั้งระบบต้นทาง: 1=คีย์ตรง, 2=ขาย, 3=ซื้อ, 4=รับ, 5=จ่าย, 6=ภายนอกอื่น
    source_type     SMALLINT NOT NULL DEFAULT 1 CHECK (source_type BETWEEN 1 AND 6),
    source_system   VARCHAR(100), -- ชื่อระบบ/ชุดนำเข้าที่มี namespace ชัดเจน
    source_record_id VARCHAR(150), -- ID รายการต้นทางคงที่ ไม่ใช่เลขที่แสดงบนเอกสารอย่างเดียว
    revision        INT NOT NULL DEFAULT 1 CHECK (revision > 0), -- เพิ่มเมื่อข้อมูลบัญชีหรือรายละเอียดประกอบเปลี่ยน
    
    -- สถานะเอกสาร: 1=DRAFT (ร่าง), 2=POSTED (ผ่านรายการแล้ว), 3=CANCELLED (ยกเลิก)
    status          SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 3),
    
    period_year     INT NOT NULL,              -- เช่น 2026
    period_no       SMALLINT NOT NULL,         -- งวดที่ 1 ถึง 12

    created_by      VARCHAR(50),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    
    PRIMARY KEY (company_code, doc_no),
    CHECK ((source_system IS NULL AND source_record_id IS NULL) OR
           (source_system IS NOT NULL AND source_record_id IS NOT NULL
            AND btrim(source_system) <> '' AND btrim(source_record_id) <> '')),
    CHECK (source_type = 1 OR source_system IS NOT NULL),
    FOREIGN KEY (company_code, book_code) REFERENCES journal_books(company_code, code) ON DELETE RESTRICT,
    FOREIGN KEY (company_code, period_year, period_no) REFERENCES fiscal_periods(company_code, period_year, period_no) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_journal_entries_date ON journal_entries(company_code, doc_date);
CREATE INDEX IF NOT EXISTS idx_journal_entries_branch ON journal_entries(company_code, branch_code);
CREATE INDEX IF NOT EXISTS idx_journal_entries_book ON journal_entries(company_code, book_code);
CREATE INDEX IF NOT EXISTS idx_journal_entries_period ON journal_entries(company_code, period_year, period_no);
-- ป้องกันรับรายการเดิมซ้ำต่อบริษัท; แก้ข้อมูลต้นทางต้องตรวจ revision ไม่สร้างใบสำคัญใหม่เงียบ ๆ
CREATE UNIQUE INDEX IF NOT EXISTS idx_journal_entries_source
    ON journal_entries(company_code, source_system, source_record_id) WHERE source_system IS NOT NULL;

-- ----------------------------------------------------------------------------
-- รายละเอียดรายการบัญชี เดบิต/เครดิต (Journal Lines - Champ BCTransSub)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS journal_lines (
    company_code    VARCHAR(20) NOT NULL,
    doc_no          VARCHAR(30) NOT NULL,
    line_number     SMALLINT NOT NULL,                 -- ลำดับบรรทัด 1, 2, 3...
    
    account_code    VARCHAR(20) NOT NULL,              -- รหัสบัญชี (อ้างอิง chart_of_accounts)
    branch_code     VARCHAR(20) NOT NULL,              -- สาขาของบรรทัดรายการ (สำหรับแยกงบรายสาขา)
    department_code VARCHAR(20),                       -- แผนก (Optional)
    project_code    VARCHAR(20),                       -- โครงการ (Optional)
    allocate_code   VARCHAR(20),                       -- การจัดสรร (Optional)
    job_code        VARCHAR(20),                       -- งาน (Optional)
    part_code       VARCHAR(20),                       -- หน่วยงาน (Optional)
    side_code       VARCHAR(20),                       -- ฝ่าย (Optional)
    
    description     VARCHAR(255),                      -- คำอธิบายรายการแต่ละบรรทัด
    
    -- แก้เป็น NUMERIC ไม่ระบุ scale เพื่อรองรับทศนิยมที่ผู้ใช้กำหนดเอง
    debit           NUMERIC NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit          NUMERIC NOT NULL DEFAULT 0 CHECK (credit >= 0),
    
    period_year     INT NOT NULL,              -- เช่น 2026
    period_no       SMALLINT NOT NULL,         -- งวดที่ 1 ถึง 12

    status          SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 3), -- 1=Draft, 2=ปกติ, 3=ยกเลิก
    posted          SMALLINT NOT NULL DEFAULT 0 CHECK (posted IN (0, 1)), -- 0=ยังไม่ผ่านรายการ, 1=ผ่านรายการแล้ว

    created_at      TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    
    PRIMARY KEY (company_code, doc_no, line_number),
    FOREIGN KEY (company_code, doc_no) REFERENCES journal_entries(company_code, doc_no) ON DELETE CASCADE,
    FOREIGN KEY (company_code, account_code) REFERENCES chart_of_accounts(company_code, account_code) ON DELETE RESTRICT,
    
    -- กฎ: แต่ละบรรทัดต้องเป็น Debit หรือ Credit อย่างใดอย่างหนึ่งเท่านั้น (ห้ามใส่ทั้งคู่)
    CONSTRAINT chk_journal_debit_credit CHECK (
        (debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_journal_lines_acc ON journal_lines(company_code, account_code);
CREATE INDEX IF NOT EXISTS idx_journal_lines_branch ON journal_lines(company_code, branch_code);
CREATE INDEX IF NOT EXISTS idx_journal_lines_period ON journal_lines(company_code, account_code, period_year, period_no);

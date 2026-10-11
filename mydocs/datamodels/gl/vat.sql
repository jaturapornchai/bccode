-- ร่างสำหรับลุงจืดตรวจแก้ (อนุมัติให้ AI เพิ่มไฟล์เฉพาะงานนี้ 2026-09-20)
-- รันหลัง gltrans.sql; ไม่ใช่ migration สำหรับตารางที่มีอยู่แล้ว
-- 1 แถว = รายการใบกำกับภาษีของใบสำคัญ ไม่ใช่บรรทัดเดบิต/เครดิต
-- จำนวนเงิน THB: NUMERIC(16,2); ส่งผ่าน API เป็น decimal string
CREATE TABLE IF NOT EXISTS vat_records (
    company_code        VARCHAR(20) NOT NULL,
    doc_no              VARCHAR(30) NOT NULL,
    line_no             INT NOT NULL CHECK (line_no > 0),
    branch_code         VARCHAR(20) NOT NULL, -- สาขาของบริษัทเจ้าของรายการภาษี
    tax_type            SMALLINT NOT NULL CHECK (tax_type IN (1, 2)), -- 1=ซื้อ, 2=ขาย
    document_type       SMALLINT NOT NULL DEFAULT 1 CHECK (document_type IN (1, 2, 3)),
    -- 1=ใบกำกับ, 2=ใบเพิ่มหนี้, 3=ใบลดหนี้; เก็บยอดบวก รายงานใช้ประเภทกำหนดเครื่องหมาย
    tax_invoice_no      VARCHAR(50) NOT NULL CHECK (btrim(tax_invoice_no) <> ''),
    tax_invoice_date    DATE NOT NULL,
    original_invoice_no VARCHAR(50),
    original_invoice_date DATE,
    tax_period_year     INT, -- ค.ศ.; NULL คู่กับเดือนเมื่อยังไม่กำหนดงวดใช้สิทธิ
    tax_period_month    SMALLINT,
    partner_tax_id      VARCHAR(13), -- NULL สำหรับกรณีที่ไม่ต้องมี; ห้ามแต่งเลขแทน
    partner_branch_no   VARCHAR(5), -- 00000=สำนักงานใหญ่; NULL=ยังไม่ระบุ/ไม่ใช้
    partner_name        VARCHAR(255) NOT NULL CHECK (btrim(partner_name) <> ''),
    base_amount         NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (base_amount >= 0 AND base_amount <> 'NaN'::numeric),
    zero_rate_amount    NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (zero_rate_amount >= 0 AND zero_rate_amount <> 'NaN'::numeric),
    exempt_amount       NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (exempt_amount >= 0 AND exempt_amount <> 'NaN'::numeric),
    vat_rate            NUMERIC(5,2) NOT NULL CHECK (vat_rate BETWEEN 0 AND 100), -- ไม่ default อัตราภาษี
    vat_amount          NUMERIC(16,2) NOT NULL DEFAULT 0 CHECK (vat_amount >= 0 AND vat_amount <> 'NaN'::numeric),
    claim_status        SMALLINT, -- เฉพาะซื้อ: 1=ใช้สิทธิ, 2=ต้องห้าม, 3=รอใช้สิทธิ, 4=ไม่ใช้สิทธิ
    claim_reason        VARCHAR(500), -- เหตุผลเลื่อนงวด/ต้องห้าม/ไม่ใช้สิทธิ
    remark              VARCHAR(500),
    created_by          VARCHAR(50),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_code, doc_no, line_no),
    FOREIGN KEY (company_code, doc_no)
        REFERENCES journal_entries(company_code, doc_no) ON DELETE RESTRICT,
    CHECK (partner_tax_id IS NULL OR partner_tax_id ~ '^[0-9]{13}$'),
    CHECK (partner_branch_no IS NULL OR partner_branch_no ~ '^[0-9]{5}$'),
    CHECK ((tax_period_year IS NULL AND tax_period_month IS NULL) OR
           (tax_period_year IS NOT NULL AND tax_period_month IS NOT NULL
            AND tax_period_year BETWEEN 1900 AND 9999 AND tax_period_month BETWEEN 1 AND 12)),
    CHECK ((tax_type = 1 AND claim_status IS NOT NULL AND claim_status BETWEEN 1 AND 4)
        OR (tax_type = 2 AND claim_status IS NULL)),
    CHECK (tax_type <> 2 OR tax_period_year IS NOT NULL),
    CHECK (claim_status IS DISTINCT FROM 1 OR tax_period_year IS NOT NULL),
    CHECK (document_type = 1 OR
           (original_invoice_no IS NOT NULL AND btrim(original_invoice_no) <> '' AND original_invoice_date IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_vat_records_period
    ON vat_records(company_code, tax_type, tax_period_year, tax_period_month, branch_code);
-- ใช้ค้นหาใบกำกับซ้ำ ไม่ตั้ง UNIQUE จนยืนยันกรณีแบ่งใบกำกับ/ปรับปรุงรายการ
CREATE INDEX IF NOT EXISTS idx_vat_records_invoice
    ON vat_records(company_code, partner_tax_id, tax_invoice_no, tax_invoice_date);

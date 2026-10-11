-- รายละเอียดสำหรับห้องบัญชี รับจากภายนอกหรือคีย์ตรง ไม่ใช่ workflow ซื้อ/ขาย/สั่งจ่าย
-- ร่าง เจ้าหนี้ แบบ many-to-many; รันหลัง partners.sql และ gltrans.sql
-- เงิน THB NUMERIC ไม่บังคับ scale ตาม GL เดิม; service ต้องตรวจ scale/rounding ก่อนบันทึก
CREATE TABLE IF NOT EXISTS ap_documents (
 company_code VARCHAR(20) NOT NULL,
 document_id VARCHAR(50) NOT NULL, -- ID ภายใน ไม่ใช้เลขใบแจ้งหนี้ผู้ขายเป็น PK
 partner_code VARCHAR(20) NOT NULL,
 document_no VARCHAR(50) NOT NULL CHECK (btrim(document_no) <> ''),
 document_date DATE NOT NULL,
 due_date DATE,
 branch_code VARCHAR(20) NOT NULL,
 document_kind SMALLINT NOT NULL CHECK (document_kind IN (1,2,3,4,5)),
 -- หลักฐานบัญชี: 1=ตั้งหนี้, 2=ผลชำระที่เกิดแล้ว, 3=เพิ่มหนี้, 4=ลดหนี้, 5=ยอดยกมา
 balance_side SMALLINT NOT NULL CHECK (balance_side IN (1,2)), -- 1=เพิ่มหนี้, 2=ลดหนี้
 amount NUMERIC NOT NULL CHECK (amount > 0 AND amount < 'Infinity'::numeric),
 currency_code VARCHAR(3) NOT NULL DEFAULT 'THB' CHECK (currency_code = 'THB'),
 created_by VARCHAR(50) NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(company_code,document_id),
 UNIQUE(company_code,partner_code,document_id,balance_side),
 FOREIGN KEY(company_code,partner_code) REFERENCES gl_partners(company_code,partner_code) ON DELETE RESTRICT,
 CHECK ((document_kind IN (1,3) AND balance_side=1) OR
        (document_kind IN (2,4) AND balance_side=2) OR document_kind=5)
);
-- เอกสารหนึ่งใบเชื่อมหลายบรรทัด/หลายใบสำคัญ และบรรทัดเดียวกระจายหลายเอกสาร
CREATE TABLE IF NOT EXISTS ap_journal_allocations (
 company_code VARCHAR(20) NOT NULL,
 allocation_id VARCHAR(50) NOT NULL, -- ใช้เป็น idempotency key ต่อบริษัท
 document_id VARCHAR(50) NOT NULL,
 doc_no VARCHAR(30) NOT NULL,
 line_number SMALLINT NOT NULL,
 amount NUMERIC NOT NULL CHECK (amount > 0 AND amount < 'Infinity'::numeric),
 created_by VARCHAR(50) NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 reversed_at TIMESTAMPTZ,
 reversed_by VARCHAR(50),
 reversal_reason VARCHAR(500),
 PRIMARY KEY(company_code,allocation_id),
 FOREIGN KEY(company_code,document_id) REFERENCES ap_documents(company_code,document_id) ON DELETE RESTRICT,
 FOREIGN KEY(company_code,doc_no,line_number) REFERENCES journal_lines(company_code,doc_no,line_number) ON DELETE RESTRICT,
 CHECK ((reversed_at IS NULL AND reversed_by IS NULL AND reversal_reason IS NULL) OR
        (reversed_at IS NOT NULL AND reversed_by IS NOT NULL AND reversal_reason IS NOT NULL AND btrim(reversal_reason)<>''))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ap_journal_active_pair
 ON ap_journal_allocations(company_code,document_id,doc_no,line_number) WHERE reversed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ap_journal_line
 ON ap_journal_allocations(company_code,doc_no,line_number);
-- กระทบยอดหนี้ many-to-many ตามหลักฐานชำระ ไม่สร้างคำสั่งรับ/จ่ายเงินจริง
CREATE TABLE IF NOT EXISTS ap_settlements (
 company_code VARCHAR(20) NOT NULL,
 settlement_id VARCHAR(50) NOT NULL,
 partner_code VARCHAR(20) NOT NULL,
 debt_document_id VARCHAR(50) NOT NULL,
 payment_document_id VARCHAR(50) NOT NULL,
 debt_side SMALLINT NOT NULL DEFAULT 1 CHECK (debt_side=1),
 payment_side SMALLINT NOT NULL DEFAULT 2 CHECK (payment_side=2),
 settlement_date DATE NOT NULL,
 amount NUMERIC NOT NULL CHECK (amount > 0 AND amount < 'Infinity'::numeric),
 created_by VARCHAR(50) NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 reversed_at TIMESTAMPTZ,
 reversed_by VARCHAR(50),
 reversal_reason VARCHAR(500),
 PRIMARY KEY(company_code,settlement_id),
 FOREIGN KEY(company_code,partner_code,debt_document_id,debt_side)
   REFERENCES ap_documents(company_code,partner_code,document_id,balance_side) ON DELETE RESTRICT,
 FOREIGN KEY(company_code,partner_code,payment_document_id,payment_side)
   REFERENCES ap_documents(company_code,partner_code,document_id,balance_side) ON DELETE RESTRICT,
 CHECK ((reversed_at IS NULL AND reversed_by IS NULL AND reversal_reason IS NULL) OR
        (reversed_at IS NOT NULL AND reversed_by IS NOT NULL AND reversal_reason IS NOT NULL AND btrim(reversal_reason)<>''))
);
CREATE INDEX IF NOT EXISTS idx_ap_settlements_debt ON ap_settlements(company_code,debt_document_id);
CREATE INDEX IF NOT EXISTS idx_ap_settlements_payment ON ap_settlements(company_code,payment_document_id);
CREATE INDEX IF NOT EXISTS idx_ap_documents_partner ON ap_documents(company_code,partner_code,due_date);


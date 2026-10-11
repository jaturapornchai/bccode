-- สำหรับห้องบัญชีตรวจหลักฐานธนาคารเท่านั้น ไม่สั่งโอนหรืออนุมัติจ่ายเงิน
-- ร่าง Statement จริงจากธนาคาร และการจับคู่ many-to-many กับบรรทัด GL
CREATE TABLE IF NOT EXISTS gl_bank_accounts (
 company_code VARCHAR(20) NOT NULL,
 bank_account_code VARCHAR(20) NOT NULL,
 bank_name VARCHAR(150) NOT NULL,
 account_number VARCHAR(50) NOT NULL,
 account_name VARCHAR(255) NOT NULL,
 gl_account_code VARCHAR(20) NOT NULL,
 currency_code VARCHAR(3) NOT NULL DEFAULT 'THB' CHECK (currency_code='THB'),
 PRIMARY KEY(company_code,bank_account_code),
 UNIQUE(company_code,bank_account_code,gl_account_code),
 FOREIGN KEY(company_code,gl_account_code) REFERENCES chart_of_accounts(company_code,account_code) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS bank_statement_lines (
 company_code VARCHAR(20) NOT NULL,
 statement_line_id VARCHAR(50) NOT NULL,
 bank_account_code VARCHAR(20) NOT NULL,
 source_key VARCHAR(150) NOT NULL, -- ID ธนาคาร หรือ file hash + เลขแถว; ห้ามใช้วันที่+ยอดอย่างเดียว
 transaction_date DATE NOT NULL,
 value_date DATE,
 bank_reference VARCHAR(150),
 description VARCHAR(500),
 direction SMALLINT NOT NULL CHECK (direction IN (1,2)), -- มุมบริษัท: 1=เงินเข้า, 2=เงินออก
 amount NUMERIC NOT NULL CHECK (amount > 0 AND amount < 'Infinity'::numeric),
 balance_after NUMERIC CHECK (balance_after > '-Infinity'::numeric AND balance_after < 'Infinity'::numeric),
 imported_by VARCHAR(50) NOT NULL,
 imported_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(company_code,statement_line_id),
 UNIQUE(company_code,bank_account_code,source_key),
 UNIQUE(company_code,bank_account_code,statement_line_id,direction),
 FOREIGN KEY(company_code,bank_account_code) REFERENCES gl_bank_accounts(company_code,bank_account_code) ON DELETE RESTRICT
);
-- ระบุว่าบรรทัด GL ใช้บัญชีธนาคารใด; หนึ่งบรรทัดใช้บัญชีธนาคารเดียว
CREATE TABLE IF NOT EXISTS bank_journal_lines (
 company_code VARCHAR(20) NOT NULL,
 doc_no VARCHAR(30) NOT NULL,
 line_number SMALLINT NOT NULL,
 bank_account_code VARCHAR(20) NOT NULL,
 gl_account_code VARCHAR(20) NOT NULL,
 direction SMALLINT NOT NULL CHECK (direction IN (1,2)),
 PRIMARY KEY(company_code,doc_no,line_number),
 UNIQUE(company_code,bank_account_code,doc_no,line_number,direction),
 FOREIGN KEY(company_code,doc_no,line_number) REFERENCES journal_lines(company_code,doc_no,line_number) ON DELETE RESTRICT,
 FOREIGN KEY(company_code,bank_account_code,gl_account_code)
   REFERENCES gl_bank_accounts(company_code,bank_account_code,gl_account_code) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS bank_statement_matches (
 company_code VARCHAR(20) NOT NULL,
 match_id VARCHAR(50) NOT NULL,
 bank_account_code VARCHAR(20) NOT NULL,
 statement_line_id VARCHAR(50) NOT NULL,
 doc_no VARCHAR(30) NOT NULL,
 line_number SMALLINT NOT NULL,
 direction SMALLINT NOT NULL CHECK (direction IN (1,2)),
 amount NUMERIC NOT NULL CHECK (amount > 0 AND amount < 'Infinity'::numeric),
 created_by VARCHAR(50) NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 reversed_at TIMESTAMPTZ,
 reversed_by VARCHAR(50),
 reversal_reason VARCHAR(500),
 PRIMARY KEY(company_code,match_id),
 FOREIGN KEY(company_code,bank_account_code,statement_line_id,direction)
   REFERENCES bank_statement_lines(company_code,bank_account_code,statement_line_id,direction) ON DELETE RESTRICT,
 FOREIGN KEY(company_code,bank_account_code,doc_no,line_number,direction)
   REFERENCES bank_journal_lines(company_code,bank_account_code,doc_no,line_number,direction) ON DELETE RESTRICT,
 CHECK ((reversed_at IS NULL AND reversed_by IS NULL AND reversal_reason IS NULL) OR
        (reversed_at IS NOT NULL AND reversed_by IS NOT NULL AND reversal_reason IS NOT NULL AND btrim(reversal_reason)<>''))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bank_matches_active_pair
 ON bank_statement_matches(company_code,statement_line_id,doc_no,line_number) WHERE reversed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_bank_matches_journal ON bank_statement_matches(company_code,doc_no,line_number);
CREATE INDEX IF NOT EXISTS idx_bank_statement_date ON bank_statement_lines(company_code,bank_account_code,transaction_date);


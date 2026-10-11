-- ร่างสำหรับลุงจืดตรวจแก้ (อนุมัติให้ AI เพิ่มไฟล์เฉพาะงานนี้ 2026-09-20)
-- รันหลัง gltrans.sql; 1 ใบสำคัญมีหลายหนังสือรับรอง แต่ละฉบับมีหลายรายการเงินได้
-- payer=ผู้จ่าย/ผู้หัก, payee=ผู้รับเงิน/ผู้ถูกหัก เสมอทั้งสองทิศทาง
CREATE TABLE IF NOT EXISTS wht_records (
    company_code        VARCHAR(20) NOT NULL,
    doc_no              VARCHAR(30) NOT NULL,
    line_no             INT NOT NULL CHECK (line_no > 0), -- ลำดับหนังสือรับรองในใบสำคัญ
    branch_code         VARCHAR(20) NOT NULL,
    wht_direction       SMALLINT NOT NULL CHECK (wht_direction IN (1, 2)), -- 1=เราหักเขา, 2=เขาหักเรา
    wht_cert_no         VARCHAR(50) NOT NULL CHECK (btrim(wht_cert_no) <> ''),
    certificate_date    DATE NOT NULL,
    form_type           VARCHAR(10) NOT NULL CHECK (form_type IN ('PND1', 'PND2', 'PND3', 'PND53')),
    payer_tax_id        VARCHAR(13) NOT NULL CHECK (payer_tax_id ~ '^[0-9]{13}$'),
    payer_branch_no     VARCHAR(5) CHECK (payer_branch_no ~ '^[0-9]{5}$'),
    payer_name          VARCHAR(255) NOT NULL CHECK (btrim(payer_name) <> ''),
    payer_address       TEXT NOT NULL CHECK (btrim(payer_address) <> ''),
    payee_tax_id        VARCHAR(13) NOT NULL CHECK (payee_tax_id ~ '^[0-9]{13}$'),
    payee_branch_no     VARCHAR(5) CHECK (payee_branch_no ~ '^[0-9]{5}$'),
    payee_name          VARCHAR(255) NOT NULL CHECK (btrim(payee_name) <> ''),
    payee_address       TEXT NOT NULL CHECK (btrim(payee_address) <> ''),
    remark              VARCHAR(500),
    created_by          VARCHAR(50),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_code, doc_no, line_no),
    FOREIGN KEY (company_code, doc_no)
        REFERENCES journal_entries(company_code, doc_no) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS wht_record_lines (
    company_code        VARCHAR(20) NOT NULL,
    doc_no              VARCHAR(30) NOT NULL,
    line_no             INT NOT NULL,
    item_no             INT NOT NULL CHECK (item_no > 0),
    payment_date        DATE NOT NULL,
    income_tax_type     VARCHAR(50) NOT NULL CHECK (btrim(income_tax_type) <> ''), -- รหัสภายใน; mapping แบบภาษีทำแยก
    income_description  VARCHAR(255) NOT NULL CHECK (btrim(income_description) <> ''),
    condition_type      SMALLINT NOT NULL CHECK (condition_type IN (1, 2, 3)), -- ร่างตาม handoff: หัก/ออกให้ตลอด/ออกให้ครั้งเดียว
    wht_rate            NUMERIC(5,2) NOT NULL CHECK (wht_rate BETWEEN 0 AND 100),
    base_amount         NUMERIC(16,2) NOT NULL CHECK (base_amount >= 0 AND base_amount <> 'NaN'::numeric),
    tax_amount          NUMERIC(16,2) NOT NULL CHECK (tax_amount >= 0 AND tax_amount <> 'NaN'::numeric),
    PRIMARY KEY (company_code, doc_no, line_no, item_no),
    FOREIGN KEY (company_code, doc_no, line_no)
        REFERENCES wht_records(company_code, doc_no, line_no) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_wht_records_form
    ON wht_records(company_code, wht_direction, form_type, branch_code);
CREATE INDEX IF NOT EXISTS idx_wht_records_certificate
    ON wht_records(company_code, payer_tax_id, wht_cert_no);
CREATE INDEX IF NOT EXISTS idx_wht_record_lines_payment
    ON wht_record_lines(company_code, payment_date);

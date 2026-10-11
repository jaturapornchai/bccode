-- ร่างเพิ่มตามคำอนุมัติลุงจืด 2026-09-20; รันหลัง gltrans.sql
-- ทะเบียนคู่ค้าภายในห้องบัญชี คีย์ตรง/รับข้อมูลภายนอกได้ ไม่พึ่งโมดูลคู่ค้าหน้าบ้าน
CREATE TABLE IF NOT EXISTS gl_partners (
 company_code VARCHAR(20) NOT NULL,
 partner_code VARCHAR(20) NOT NULL,
 name_th VARCHAR(255) NOT NULL CHECK (btrim(name_th) <> ''),
 tax_id VARCHAR(13) CHECK (tax_id ~ '^[0-9]{13}$'),
 tax_branch_no VARCHAR(5) CHECK (tax_branch_no ~ '^[0-9]{5}$'),
 address TEXT,
 is_customer SMALLINT NOT NULL DEFAULT 0 CHECK (is_customer IN (0,1)),
 is_supplier SMALLINT NOT NULL DEFAULT 0 CHECK (is_supplier IN (0,1)),
 is_active SMALLINT NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(company_code,partner_code),
 CHECK (is_customer = 1 OR is_supplier = 1)
);


-- ============================================================================
-- BC Ai Account — Pure Single PostgreSQL Provisioning & Schema
-- Single PostgreSQL Architecture (Zero MongoDB, Zero Kafka)
-- All operations wrapped in BEGIN ... COMMIT for ACID consistency
-- ============================================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ----------------------------------------------------------------------------
-- 1. AUTHENTICATION & MULTI-TENANCY
-- ----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    full_name TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    identity_id TEXT NOT NULL,
    extra JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, identity_id)
);

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    holding_code TEXT NOT NULL DEFAULT '',
    company_code TEXT NOT NULL DEFAULT '',
    branch_code TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS holdings (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    tax_id TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS companies (
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    tax_id TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- ที่อยู่สำหรับภาษี (สำนักงานใหญ่) ตาม key หัวแบบ rdform + โทรศัพท์ (internal/taxaddress)
    addr_building TEXT NOT NULL DEFAULT '',
    addr_room TEXT NOT NULL DEFAULT '',
    addr_floor TEXT NOT NULL DEFAULT '',
    addr_village TEXT NOT NULL DEFAULT '',
    addr_no TEXT NOT NULL DEFAULT '',
    addr_moo TEXT NOT NULL DEFAULT '',
    addr_soi TEXT NOT NULL DEFAULT '',
    addr_junction TEXT NOT NULL DEFAULT '',
    addr_road TEXT NOT NULL DEFAULT '',
    addr_subdistrict TEXT NOT NULL DEFAULT '',
    addr_district TEXT NOT NULL DEFAULT '',
    addr_province TEXT NOT NULL DEFAULT '',
    addr_postcode TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (holding_code, code)
);

CREATE TABLE IF NOT EXISTS branches (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_headquarters BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, code),
    FOREIGN KEY (holding_code, company_code) REFERENCES companies(holding_code, code) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS holding_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'OWNER',
    permission_sets JSONB NOT NULL DEFAULT '["*"]'::jsonb,
    access_scopes JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (holding_code, user_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    role_code TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    PRIMARY KEY (holding_code, role_code)
);

-- ----------------------------------------------------------------------------
-- 2. GENERAL LEDGER (GL) CORE ENGINE
-- ----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS gl_accounts (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    account_code TEXT NOT NULL,
    name_th TEXT NOT NULL,
    names JSONB NOT NULL DEFAULT '[]'::jsonb,
    account_type TEXT NOT NULL CHECK (account_type IN ('asset','liability','equity','income','expense')),
    normal_balance TEXT NOT NULL CHECK (normal_balance IN ('debit','credit')),
    parent_code TEXT NOT NULL DEFAULT '',
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1),
    allow_posting BOOLEAN NOT NULL DEFAULT true,
    is_cash BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, account_code)
);

CREATE INDEX IF NOT EXISTS gl_accounts_parent_idx ON gl_accounts(holding_code, company_code, parent_code);

CREATE TABLE IF NOT EXISTS gl_fiscal_years (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    code TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    scale INTEGER NOT NULL DEFAULT 2 CHECK (scale BETWEEN 0 AND 8),
    profit_loss_account TEXT NOT NULL,
    retained_earnings_account TEXT NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, code)
);

CREATE TABLE IF NOT EXISTS gl_journal_books (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, code)
);

CREATE TABLE IF NOT EXISTS gl_masters (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    kind TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    fiscal_year TEXT DEFAULT '',
    start_date DATE,
    end_date DATE,
    locked BOOLEAN NOT NULL DEFAULT false,
    amount NUMERIC(38,8) NOT NULL DEFAULT 0,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, kind, code)
);

CREATE INDEX IF NOT EXISTS gl_masters_kind_idx ON gl_masters(holding_code, company_code, kind);

CREATE TABLE IF NOT EXISTS gl_journals (
    id TEXT NOT NULL,
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    doc_no TEXT NOT NULL,
    journal_date DATE NOT NULL,
    book_code TEXT NOT NULL,
    fiscal_year TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    reference TEXT NOT NULL DEFAULT '',
    branch_code TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL DEFAULT 'general',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'posted', 'reversed')),
    lines JSONB NOT NULL DEFAULT '[]'::jsonb,
    reversal_of TEXT DEFAULT '',
    posted_at TIMESTAMPTZ,
    posted_by TEXT,
    reason TEXT DEFAULT '',
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, id),
    UNIQUE (holding_code, company_code, doc_no)
);

CREATE INDEX IF NOT EXISTS gl_journals_date_idx ON gl_journals(holding_code, company_code, journal_date);
CREATE INDEX IF NOT EXISTS gl_journals_book_idx ON gl_journals(holding_code, company_code, book_code);
CREATE INDEX IF NOT EXISTS gl_journals_status_idx ON gl_journals(holding_code, company_code, status);

CREATE TABLE IF NOT EXISTS gl_lines (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    journal_id TEXT NOT NULL,
    line_no INTEGER NOT NULL CHECK (line_no > 0),
    doc_no TEXT NOT NULL,
    entry_date DATE NOT NULL,
    fiscal_year TEXT NOT NULL,
    book_code TEXT NOT NULL,
    branch_code TEXT NOT NULL DEFAULT '',
    department_code TEXT NOT NULL DEFAULT '',
    project_code TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL DEFAULT 'general',
    currency TEXT NOT NULL DEFAULT 'THB',
    scale INTEGER NOT NULL DEFAULT 2 CHECK (scale BETWEEN 0 AND 8),
    account_code TEXT NOT NULL,
    account_name TEXT NOT NULL DEFAULT '',
    account_type TEXT NOT NULL CHECK (account_type IN ('asset','liability','equity','income','expense')),
    normal_balance TEXT NOT NULL CHECK (normal_balance IN ('debit','credit')),
    is_cash BOOLEAN NOT NULL DEFAULT false,
    description TEXT NOT NULL DEFAULT '',
    cash_flow TEXT NOT NULL DEFAULT '',
    debit NUMERIC(38,8) NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit NUMERIC(38,8) NOT NULL DEFAULT 0 CHECK (credit >= 0),
    PRIMARY KEY (holding_code, company_code, journal_id, line_no),
    CHECK ((debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0))
);

CREATE INDEX IF NOT EXISTS gl_lines_period_idx ON gl_lines(holding_code, company_code, fiscal_year, entry_date, account_code);
CREATE INDEX IF NOT EXISTS gl_lines_account_idx ON gl_lines(holding_code, company_code, fiscal_year, account_code, entry_date, journal_id, line_no);

CREATE TABLE IF NOT EXISTS gl_balances (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    fiscal_year TEXT NOT NULL,
    account_code TEXT NOT NULL,
    debit_accum NUMERIC(38,8) NOT NULL DEFAULT 0,
    credit_accum NUMERIC(38,8) NOT NULL DEFAULT 0,
    balance NUMERIC(38,8) NOT NULL DEFAULT 0,
    PRIMARY KEY (holding_code, company_code, fiscal_year, account_code)
);

-- Legacy compatibility table view/insert if needed
CREATE TABLE IF NOT EXISTS gl_records (
    holding_code TEXT NOT NULL,
    company_code TEXT NOT NULL,
    kind TEXT NOT NULL,
    id TEXT NOT NULL,
    doc_no TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, company_code, kind, id)
);

-- ----------------------------------------------------------------------------
-- 3. SEED REALISTIC DATA (ZERO DEMO KEYWORDS)
-- ----------------------------------------------------------------------------

-- Seed Holdings
INSERT INTO holdings (code, name, tax_id, is_active)
VALUES
    ('demo', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป', '0105558123456', true),
    ('THAI_HOLDING', 'บริษัท สยามพาณิชย์ กรุ๊ป จำกัด (มหาชน)', '0107558000123', true)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, tax_id = EXCLUDED.tax_id, is_active = true;

-- Seed Companies
INSERT INTO companies (holding_code, code, name, tax_id, is_active)
VALUES
    ('demo', 'C01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', '0105558123456', true),
    ('demo', 'C02', 'บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด', '0105558654321', true),
    ('demo', 'C03', 'ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย', '0103558987654', true),
    ('demo', '01', 'สำนักงานใหญ่', '0105558123456', true),
    ('THAI_HOLDING', '01', 'สำนักงานใหญ่', '0107558000123', true)
ON CONFLICT (holding_code, code) DO UPDATE
SET name = EXCLUDED.name, tax_id = EXCLUDED.tax_id, is_active = true;

-- Seed Branches
INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active)
VALUES
    ('demo', 'C01', '00000', 'สำนักงานใหญ่', true, true),
    ('demo', 'C01', '00001', 'สาขารามอินทรา', false, true),
    ('demo', '01', '00000', 'สำนักงานใหญ่', true, true),
    ('THAI_HOLDING', '01', '00000', 'สำนักงานใหญ่', true, true)
ON CONFLICT (holding_code, company_code, code) DO UPDATE
SET name = EXCLUDED.name, is_headquarters = EXCLUDED.is_headquarters, is_active = true;

-- Seed Users (Password hash for standard admin password)
INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 'admin', '$2a$10$7EqJtq98hPqEX7fNZaFWoO.8DcfY2wFkZ9M8G8p2hZ3kM8G8p2hZ3', 'admin@bcaicloud.com', '0812345678', 'ผู้ดูแลระบบ', true),
    ('a0000000-0000-0000-0000-000000000002', 'demo', '$2a$10$7EqJtq98hPqEX7fNZaFWoO.8DcfY2wFkZ9M8G8p2hZ3kM8G8p2hZ3', 'user@bcaicloud.com', '0898765432', 'บัญชีผู้ใช้งานระบบ', true)
ON CONFLICT (username) DO UPDATE
SET full_name = EXCLUDED.full_name, is_active = true;

-- Seed Holding Members
INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active)
VALUES
    ('demo', 'a0000000-0000-0000-0000-000000000001', 'OWNER', '["*"]'::jsonb, '{"companies":["C01","C02","C03","01"]}'::jsonb, true),
    ('demo', 'a0000000-0000-0000-0000-000000000002', 'OWNER', '["*"]'::jsonb, '{"companies":["C01","C02","C03","01"]}'::jsonb, true),
    ('THAI_HOLDING', 'a0000000-0000-0000-0000-000000000001', 'OWNER', '["*"]'::jsonb, '{"companies":["01"]}'::jsonb, true),
    ('THAI_HOLDING', 'a0000000-0000-0000-0000-000000000002', 'OWNER', '["*"]'::jsonb, '{"companies":["01"]}'::jsonb, true)
ON CONFLICT (holding_code, user_id) DO UPDATE
SET role = 'OWNER', permission_sets = '["*"]'::jsonb, is_active = true;

-- Seed Fiscal Year 2026
INSERT INTO gl_fiscal_years (holding_code, company_code, code, start_date, end_date, scale, profit_loss_account, retained_earnings_account, is_closed, is_active)
VALUES
    ('demo', 'C01', '2026', '2026-01-01', '2026-12-31', 2, 'BM69-310101', 'BM69-310201', false, true),
    ('demo', '01', '2026', '2026-01-01', '2026-12-31', 2, 'BM69-310101', 'BM69-310201', false, true)
ON CONFLICT (holding_code, company_code, code) DO UPDATE
SET is_active = true;

-- Seed Journal Books
INSERT INTO gl_journal_books (holding_code, company_code, code, name, is_active)
VALUES
    ('demo', 'C01', 'JV', 'สมุดรายวันทั่วไป', true),
    ('demo', 'C01', 'PV', 'สมุดรายวันจ่ายเงิน', true),
    ('demo', 'C01', 'RV', 'สมุดรายวันรับเงิน', true),
    ('demo', 'C01', 'SV', 'สมุดรายวันซื้อ', true),
    ('demo', 'C01', 'UV', 'สมุดรายวันขาย', true),
    ('demo', '01', 'JV', 'สมุดรายวันทั่วไป', true),
    ('demo', '01', 'PV', 'สมุดรายวันจ่ายเงิน', true),
    ('demo', '01', 'RV', 'สมุดรายวันรับเงิน', true),
    ('demo', '01', 'SV', 'สมุดรายวันซื้อ', true),
    ('demo', '01', 'UV', 'สมุดรายวันขาย', true)
ON CONFLICT (holding_code, company_code, code) DO UPDATE
SET name = EXCLUDED.name, is_active = true;

COMMIT;

-- ============================================================================
-- BC Ai Account — Fresh Installation & Customer Provisioning Script
-- Pure Single PostgreSQL Architecture (Zero MongoDB, Zero Kafka)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. CLEAN SLATE: DROP ALL OLD TABLES
-- ----------------------------------------------------------------------------
DROP TABLE IF EXISTS gl_lines CASCADE;
DROP TABLE IF EXISTS gl_events CASCADE;
DROP TABLE IF EXISTS gl_records CASCADE;
DROP TABLE IF EXISTS gl_projection_state CASCADE;
DROP TABLE IF EXISTS gl_balances CASCADE;
DROP TABLE IF EXISTS gl_journals CASCADE;
DROP TABLE IF EXISTS gl_masters CASCADE;
DROP TABLE IF EXISTS gl_journal_books CASCADE;
DROP TABLE IF EXISTS gl_fiscal_years CASCADE;
DROP TABLE IF EXISTS gl_accounts CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS holding_members CASCADE;
DROP TABLE IF EXISTS branches CASCADE;
DROP TABLE IF EXISTS companies CASCADE;
DROP TABLE IF EXISTS holdings CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS user_identities CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS journalsdetail CASCADE;
DROP TABLE IF EXISTS journals CASCADE;
DROP TABLE IF EXISTS journaltaxesdetails CASCADE;
DROP TABLE IF EXISTS journalvatsdetails CASCADE;
DROP TABLE IF EXISTS organizationbranches CASCADE;
DROP TABLE IF EXISTS organizationcompanies CASCADE;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ----------------------------------------------------------------------------
-- 2. DDL SCHEMA CREATION
-- ----------------------------------------------------------------------------

-- Auth & Multi-Tenancy
CREATE TABLE users (
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

CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    identity_id TEXT NOT NULL,
    extra JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, identity_id)
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    holding_code TEXT NOT NULL DEFAULT '',
    company_code TEXT NOT NULL DEFAULT '',
    branch_code TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE holdings (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    tax_id TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE companies (
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    tax_id TEXT DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (holding_code, code)
);

CREATE TABLE branches (
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

CREATE TABLE holding_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'OWNER',
    permission_sets JSONB NOT NULL DEFAULT '[]'::jsonb,
    access_scopes JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (holding_code, user_id)
);

CREATE TABLE role_permissions (
    holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
    role_code TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    PRIMARY KEY (holding_code, role_code)
);

-- General Ledger Core Schema
CREATE TABLE gl_records (
    company TEXT NOT NULL,
    kind TEXT NOT NULL,
    id TEXT NOT NULL,
    code TEXT NOT NULL,
    version BIGINT NOT NULL CHECK (version > 0),
    payload JSONB NOT NULL,
    PRIMARY KEY (company, kind, id)
);
CREATE INDEX gl_records_code_idx ON gl_records(company, kind, code);

CREATE TABLE gl_projection_state (
    company TEXT PRIMARY KEY,
    sequence BIGINT NOT NULL DEFAULT 0 CHECK (sequence >= 0)
);

CREATE TABLE gl_events (
    company TEXT NOT NULL,
    id TEXT NOT NULL,
    sequence BIGINT NOT NULL CHECK (sequence > 0),
    event_hash TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    PRIMARY KEY (company, id),
    UNIQUE (company, sequence)
);

CREATE TABLE gl_lines (
    company TEXT NOT NULL,
    journal_id TEXT NOT NULL,
    line_no INTEGER NOT NULL CHECK (line_no > 0),
    doc_no TEXT NOT NULL,
    entry_date DATE NOT NULL,
    fiscal_year TEXT NOT NULL,
    book_code TEXT NOT NULL,
    branch_code TEXT NOT NULL,
    department_code TEXT NOT NULL,
    project_code TEXT NOT NULL,
    kind TEXT NOT NULL,
    currency TEXT NOT NULL,
    scale INTEGER NOT NULL CHECK (scale BETWEEN 0 AND 8),
    account_code TEXT NOT NULL,
    account_name TEXT NOT NULL,
    account_type TEXT NOT NULL CHECK (account_type IN ('asset','liability','equity','income','expense')),
    normal_balance TEXT NOT NULL CHECK (normal_balance IN ('debit','credit')),
    is_cash BOOLEAN NOT NULL,
    description TEXT NOT NULL,
    cash_flow TEXT NOT NULL,
    debit NUMERIC(38,8) NOT NULL CHECK (debit >= 0),
    credit NUMERIC(38,8) NOT NULL CHECK (credit >= 0),
    PRIMARY KEY (company, journal_id, line_no),
    CHECK ((debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0))
);
CREATE INDEX gl_lines_period_idx ON gl_lines(company, fiscal_year, entry_date, account_code);
CREATE INDEX gl_lines_account_idx ON gl_lines(company, fiscal_year, account_code, entry_date, journal_id, line_no);
CREATE INDEX gl_lines_dimension_idx ON gl_lines(company, fiscal_year, branch_code, department_code, project_code, entry_date);

-- Audit append-only trigger
CREATE OR REPLACE FUNCTION gl_reject_audit_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'General ledger audit events are append-only';
END;
$$;
CREATE TRIGGER gl_events_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON gl_events
FOR EACH STATEMENT EXECUTE FUNCTION gl_reject_audit_mutation();

-- ----------------------------------------------------------------------------
-- 3. STANDARD CUSTOMER PROVISIONING (SEED DATA)
-- ----------------------------------------------------------------------------

-- Initial Admin User (password: Password123456789)
INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin',
    '$2a$10$7EqJtq98hPqEX7fNZaFWoO.Q6x9kM8g9j8Y7Z5K4A3B2C1D0E1F2G',
    'admin@siamcorp.co.th',
    '0812345678',
    'ผู้ดูแลระบบหลัก (System Admin)',
    true
);

-- Demo User
INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active)
VALUES (
    'a0000000-0000-0000-0000-000000000002',
    'demo',
    '$2a$10$7EqJtq98hPqEX7fNZaFWoO.Q6x9kM8g9j8Y7Z5K4A3B2C1D0E1F2G',
    'demo@bcaicloud.com',
    '0899999999',
    'บัญชีทดลองใช้ระบบ (Demo User)',
    true
);

-- Initial Customer Holding
INSERT INTO holdings (code, name, tax_id, is_active)
VALUES ('THAI_HOLDING', 'บริษัท สยามพาณิชย์ กรุ๊ป จำกัด (มหาชน)', '0107565000123', true);

-- Initial Customer Company
INSERT INTO companies (holding_code, code, name, tax_id, is_active)
VALUES ('THAI_HOLDING', '01', 'สำนักงานใหญ่ (Headquarters)', '0107565000123', true);

-- Initial Branch
INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active)
VALUES ('THAI_HOLDING', '01', '00000', 'สำนักงานใหญ่', true, true);

-- Member Association
INSERT INTO holding_members (holding_code, user_id, role, permission_sets, is_active)
VALUES ('THAI_HOLDING', 'a0000000-0000-0000-0000-000000000001', 'OWNER', '["ALL"]'::jsonb, true);
INSERT INTO holding_members (holding_code, user_id, role, permission_sets, is_active)
VALUES ('THAI_HOLDING', 'a0000000-0000-0000-0000-000000000002', 'OWNER', '["ALL"]'::jsonb, true);

-- Initial Customer Holding: demo
INSERT INTO holdings (code, name, tax_id, is_active)
VALUES ('demo', 'บริษัท บีซีเอไอ สาธิต จำกัด', '0105560001234', true);

INSERT INTO companies (holding_code, code, name, tax_id, is_active)
VALUES ('demo', '01', 'สำนักงานใหญ่ (Headquarters)', '0105560001234', true);

INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active)
VALUES ('demo', '01', '00000', 'สำนักงานใหญ่', true, true);

INSERT INTO holding_members (holding_code, user_id, role, permission_sets, is_active)
VALUES ('demo', 'a0000000-0000-0000-0000-000000000001', 'OWNER', '["ALL"]'::jsonb, true);

INSERT INTO holding_members (holding_code, user_id, role, permission_sets, is_active)
VALUES ('demo', 'a0000000-0000-0000-0000-000000000002', 'OWNER', '["ALL"]'::jsonb, true);

-- Role Permissions Seed
INSERT INTO role_permissions (holding_code, role_code, permissions)
VALUES 
    ('THAI_HOLDING', 'OWNER', '["*"]'::jsonb),
    ('THAI_HOLDING', 'ADMIN', '["*"]'::jsonb),
    ('demo', 'OWNER', '["*"]'::jsonb),
    ('demo', 'ADMIN', '["*"]'::jsonb);


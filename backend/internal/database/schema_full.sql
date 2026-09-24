-- ============================================================================
-- BC Ai Account — Pure PostgreSQL Core Schema
-- Single PostgreSQL Architecture (No MongoDB, No Kafka)
-- ============================================================================

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
    provider TEXT NOT NULL, -- 'google', 'line', 'password'
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
    role TEXT NOT NULL DEFAULT 'USER', -- 'OWNER', 'ADMIN', 'USER'
    permission_sets JSONB NOT NULL DEFAULT '[]'::jsonb,
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
    kind TEXT NOT NULL, -- 'budgets', 'periods', 'mappings', 'account-groups', 'statement-templates'
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
    kind TEXT NOT NULL DEFAULT 'general', -- 'general', 'opening', 'closing'
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
CREATE INDEX IF NOT EXISTS gl_lines_dimension_idx ON gl_lines(holding_code, company_code, fiscal_year, branch_code, department_code, project_code, entry_date);

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

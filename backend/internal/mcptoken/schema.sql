CREATE TABLE IF NOT EXISTS mcp_access_tokens (
 id text PRIMARY KEY,
 holding_code text NOT NULL,
 company_code text NOT NULL,
 branch_code text NOT NULL DEFAULT '',
 kind text NOT NULL CHECK (kind IN ('api','mcp')),
 name text NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
 mode text NOT NULL CHECK (mode IN ('readonly','readwrite')),
 token_hash bytea NOT NULL CHECK (octet_length(token_hash)=32),
 created_by text NOT NULL,
 creator_username text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 expires_at timestamptz NOT NULL,
 revoked_at timestamptz,
 last_used_at timestamptz,
 CHECK (expires_at>created_at)
);
CREATE INDEX IF NOT EXISTS mcp_access_tokens_scope ON mcp_access_tokens(holding_code,company_code,branch_code,created_at DESC);
CREATE TABLE IF NOT EXISTS mcp_token_audit (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 token_id text NOT NULL REFERENCES mcp_access_tokens(id),
 holding_code text NOT NULL,
 company_code text NOT NULL,
 branch_code text NOT NULL,
 actor text NOT NULL,
 action text NOT NULL CHECK (action IN ('create','revoke','use')),
 occurred_at timestamptz NOT NULL DEFAULT now()
);

-- Preserve deployed single-company grants; never infer all-holding access.
CREATE TABLE IF NOT EXISTS mcp_token_companies (
 token_id text NOT NULL REFERENCES mcp_access_tokens(id),
 holding_code text NOT NULL,
 company_code text NOT NULL CHECK (company_code<>''),
 PRIMARY KEY(token_id,company_code)
);
CREATE INDEX IF NOT EXISTS mcp_token_companies_holding ON mcp_token_companies(holding_code,token_id);
INSERT INTO mcp_token_companies(token_id,holding_code,company_code)
 SELECT id,holding_code,company_code FROM mcp_access_tokens WHERE company_code<>''
 ON CONFLICT(token_id,company_code) DO NOTHING;

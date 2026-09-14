CREATE TABLE IF NOT EXISTS gl_records (
    company text NOT NULL, kind text NOT NULL, id text NOT NULL, code text NOT NULL,
    version bigint NOT NULL CHECK (version > 0), payload jsonb NOT NULL,
    PRIMARY KEY (company, kind, id)
);
CREATE INDEX IF NOT EXISTS gl_records_code_idx ON gl_records(company,kind,code);
CREATE TABLE IF NOT EXISTS gl_projection_state (
    company text PRIMARY KEY, sequence bigint NOT NULL DEFAULT 0 CHECK (sequence >= 0)
);
CREATE TABLE IF NOT EXISTS gl_events (
    company text NOT NULL, id text NOT NULL, sequence bigint NOT NULL CHECK (sequence > 0),
    event_hash text NOT NULL, occurred_at timestamptz NOT NULL, payload jsonb NOT NULL,
    PRIMARY KEY (company,id), UNIQUE (company,sequence)
);
CREATE TABLE IF NOT EXISTS gl_lines (
    company text NOT NULL, journal_id text NOT NULL, line_no integer NOT NULL CHECK (line_no > 0),
    doc_no text NOT NULL, entry_date date NOT NULL, fiscal_year text NOT NULL, book_code text NOT NULL,
    branch_code text NOT NULL, department_code text NOT NULL, project_code text NOT NULL,
    kind text NOT NULL, currency text NOT NULL, scale integer NOT NULL CHECK (scale BETWEEN 0 AND 8),
    account_code text NOT NULL, account_name text NOT NULL,
    account_type text NOT NULL CHECK (account_type IN ('asset','liability','equity','income','expense')),
    normal_balance text NOT NULL CHECK (normal_balance IN ('debit','credit')), is_cash boolean NOT NULL,
    description text NOT NULL, cash_flow text NOT NULL,
    debit numeric(38,8) NOT NULL CHECK (debit >= 0), credit numeric(38,8) NOT NULL CHECK (credit >= 0),
    PRIMARY KEY (company,journal_id,line_no), CHECK ((debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0))
);
CREATE INDEX IF NOT EXISTS gl_lines_period_idx ON gl_lines(company,fiscal_year,entry_date,account_code);
CREATE INDEX IF NOT EXISTS gl_lines_account_idx ON gl_lines(company,fiscal_year,account_code,entry_date,journal_id,line_no);
CREATE INDEX IF NOT EXISTS gl_lines_dimension_idx ON gl_lines(company,fiscal_year,branch_code,department_code,project_code,entry_date);

-- Delivery metadata belongs in the Mongo outbox. PostgreSQL audit snapshots
-- cannot be amended or removed by the normal application connection.
CREATE OR REPLACE FUNCTION gl_reject_audit_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'General ledger audit events are append-only';
END;
$$;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid='gl_events'::regclass AND tgname='gl_events_immutable') THEN
        CREATE TRIGGER gl_events_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON gl_events
        FOR EACH STATEMENT EXECUTE FUNCTION gl_reject_audit_mutation();
    END IF;
END;
$$;
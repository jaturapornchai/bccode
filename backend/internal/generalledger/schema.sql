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
CREATE TABLE IF NOT EXISTS gl_journal_review_events (
    company text NOT NULL, journal_id text NOT NULL,
    event_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    journal_version bigint NOT NULL CHECK (journal_version > 0),
    status smallint NOT NULL CHECK (status IN (1,2,3)),
    note text NOT NULL CHECK (length(note) <= 4000),
    reviewed_by text NOT NULL CHECK (btrim(reviewed_by) <> ''),
    reviewed_at timestamptz NOT NULL,
    CHECK (status <> 2 OR btrim(note) <> '')
);
CREATE INDEX IF NOT EXISTS gl_journal_review_current_idx ON gl_journal_review_events(company,journal_id,journal_version,event_no DESC);
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid='gl_journal_review_events'::regclass AND tgname='gl_journal_review_immutable') THEN
        CREATE TRIGGER gl_journal_review_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON gl_journal_review_events
        FOR EACH STATEMENT EXECUTE FUNCTION gl_reject_audit_mutation();
    END IF;
END; $$;

-- Stable source document identity survives new request IDs and projection rebuilds.
CREATE TABLE IF NOT EXISTS gl_source_journals (
    company text NOT NULL, source_system text NOT NULL, source_record_id text NOT NULL,
    journal_id text NOT NULL, request_hash text NOT NULL, sequence bigint NOT NULL,
    version bigint NOT NULL, created_at timestamptz NOT NULL,
    PRIMARY KEY (company,source_system,source_record_id), UNIQUE(company,journal_id),
    CHECK (length(btrim(source_system)) BETWEEN 1 AND 100),
    CHECK (length(btrim(source_record_id)) BETWEEN 1 AND 150)
);

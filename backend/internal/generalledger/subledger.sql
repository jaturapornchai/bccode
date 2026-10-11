-- Authoritative accounting-office evidence. These tables deliberately do not
-- reference gl_records: ledger projection rebuilds replace that cache.
CREATE TABLE IF NOT EXISTS gl_subledger_partners (
 company text NOT NULL, code text NOT NULL, version bigint NOT NULL CHECK(version>0), payload jsonb NOT NULL,
 PRIMARY KEY(company,code)
);
CREATE TABLE IF NOT EXISTS gl_subledger_bank_accounts (
 company text NOT NULL, code text NOT NULL, version bigint NOT NULL CHECK(version>0), payload jsonb NOT NULL,
 PRIMARY KEY(company,code)
);
CREATE TABLE IF NOT EXISTS gl_subledger_documents (
 company text NOT NULL,id text NOT NULL,ledger text NOT NULL CHECK(ledger IN ('ar','ap')),
 partner_code text NOT NULL,branch_code text NOT NULL,control_account_code text NOT NULL,
 amount numeric(34,8) NOT NULL CHECK(amount>0 AND amount<'Infinity'),side smallint NOT NULL CHECK(side IN(1,2)),
 version bigint NOT NULL CHECK(version>0),created_journal_id text NOT NULL,payload jsonb NOT NULL,
 PRIMARY KEY(company,id),FOREIGN KEY(company,partner_code) REFERENCES gl_subledger_partners(company,code)
);
CREATE INDEX IF NOT EXISTS gl_subledger_documents_partner ON gl_subledger_documents(company,ledger,partner_code,id);
CREATE TABLE IF NOT EXISTS gl_subledger_allocations (
 company text NOT NULL,id text NOT NULL,document_id text NOT NULL,journal_id text NOT NULL,line_number int NOT NULL CHECK(line_number>0),
 amount numeric(34,8) NOT NULL CHECK(amount>0 AND amount<'Infinity'),payload jsonb NOT NULL,created_at timestamptz NOT NULL,created_by text NOT NULL,
 reversed_at timestamptz,reversed_by text,reversal_reason text,
 PRIMARY KEY(company,id),FOREIGN KEY(company,document_id) REFERENCES gl_subledger_documents(company,id),
 CHECK((reversed_at IS NULL AND reversed_by IS NULL AND reversal_reason IS NULL) OR (reversed_at IS NOT NULL AND reversed_by IS NOT NULL AND reversal_reason IS NOT NULL AND length(btrim(reversal_reason))>0))
);
CREATE UNIQUE INDEX IF NOT EXISTS gl_subledger_allocation_pair ON gl_subledger_allocations(company,document_id,journal_id,line_number) WHERE reversed_at IS NULL;
CREATE INDEX IF NOT EXISTS gl_subledger_allocation_journal ON gl_subledger_allocations(company,journal_id,line_number);
CREATE TABLE IF NOT EXISTS gl_subledger_settlements (
 company text NOT NULL,id text NOT NULL,journal_id text NOT NULL,debt_id text NOT NULL,payment_id text NOT NULL,
 amount numeric(34,8) NOT NULL CHECK(amount>0 AND amount<'Infinity'),settlement_date date NOT NULL,payload jsonb NOT NULL,
 created_at timestamptz NOT NULL,created_by text NOT NULL,reversed_at timestamptz,reversed_by text,reversal_reason text,
 PRIMARY KEY(company,id),FOREIGN KEY(company,debt_id) REFERENCES gl_subledger_documents(company,id),FOREIGN KEY(company,payment_id) REFERENCES gl_subledger_documents(company,id),CHECK(debt_id<>payment_id),
 CHECK((reversed_at IS NULL AND reversed_by IS NULL AND reversal_reason IS NULL) OR (reversed_at IS NOT NULL AND reversed_by IS NOT NULL AND reversal_reason IS NOT NULL AND length(btrim(reversal_reason))>0))
);
CREATE INDEX IF NOT EXISTS gl_subledger_settlement_debt ON gl_subledger_settlements(company,debt_id);
CREATE INDEX IF NOT EXISTS gl_subledger_settlement_payment ON gl_subledger_settlements(company,payment_id);
CREATE TABLE IF NOT EXISTS gl_subledger_bank_lines (
 company text NOT NULL,journal_id text NOT NULL,line_number int NOT NULL CHECK(line_number>0),bank_account_code text NOT NULL,
 direction smallint NOT NULL CHECK(direction IN(1,2)),amount numeric(34,8) NOT NULL CHECK(amount>0 AND amount<'Infinity'),payload jsonb NOT NULL,
 PRIMARY KEY(company,journal_id,line_number),FOREIGN KEY(company,bank_account_code) REFERENCES gl_subledger_bank_accounts(company,code)
);
CREATE TABLE IF NOT EXISTS gl_subledger_statements (
 company text NOT NULL,id text NOT NULL,owner_journal_id text NOT NULL,bank_account_code text NOT NULL,source_key text NOT NULL,
 transaction_date date NOT NULL,direction smallint NOT NULL CHECK(direction IN(1,2)),amount numeric(34,8) NOT NULL CHECK(amount>0 AND amount<'Infinity'),
 payload jsonb NOT NULL,created_at timestamptz NOT NULL,created_by text NOT NULL,
 PRIMARY KEY(company,id),UNIQUE(company,bank_account_code,source_key),FOREIGN KEY(company,bank_account_code) REFERENCES gl_subledger_bank_accounts(company,code)
);
ALTER TABLE gl_subledger_statements ADD COLUMN IF NOT EXISTS owner_journal_id text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS gl_subledger_statement_date ON gl_subledger_statements(company,transaction_date,id);
CREATE TABLE IF NOT EXISTS gl_subledger_matches (
 company text NOT NULL,id text NOT NULL,owner_journal_id text NOT NULL,statement_id text NOT NULL,journal_id text NOT NULL,line_number int NOT NULL,
 amount numeric(34,8) NOT NULL CHECK(amount>0 AND amount<'Infinity'),payload jsonb NOT NULL,
 created_at timestamptz NOT NULL,created_by text NOT NULL,reversed_at timestamptz,reversed_by text,reversal_reason text,
 PRIMARY KEY(company,id),FOREIGN KEY(company,statement_id) REFERENCES gl_subledger_statements(company,id),
 FOREIGN KEY(company,journal_id,line_number) REFERENCES gl_subledger_bank_lines(company,journal_id,line_number),
 CHECK((reversed_at IS NULL AND reversed_by IS NULL AND reversal_reason IS NULL) OR (reversed_at IS NOT NULL AND reversed_by IS NOT NULL AND reversal_reason IS NOT NULL AND length(btrim(reversal_reason))>0))
);
CREATE UNIQUE INDEX IF NOT EXISTS gl_subledger_match_pair ON gl_subledger_matches(company,statement_id,journal_id,line_number) WHERE reversed_at IS NULL;
CREATE TABLE IF NOT EXISTS gl_subledger_audit (
 event_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,company text NOT NULL,journal_id text NOT NULL,
 action text NOT NULL,actor text NOT NULL,occurred_at timestamptz NOT NULL,payload jsonb NOT NULL
);
CREATE INDEX IF NOT EXISTS gl_subledger_audit_journal ON gl_subledger_audit(company,journal_id,event_no);

DROP TRIGGER IF EXISTS gl_subledger_audit_append_only ON gl_subledger_audit;
CREATE TRIGGER gl_subledger_audit_append_only BEFORE UPDATE OR DELETE OR TRUNCATE ON gl_subledger_audit FOR EACH STATEMENT EXECUTE FUNCTION gl_reject_audit_mutation();

ALTER TABLE gl_subledger_allocations ADD COLUMN IF NOT EXISTS reversed_effective_date date;

-- AP payment preparation (Champ BCPaymentTemp/BCPaymentTempSub2): reserves AP subledger debts for a later payment voucher; never posts.
-- Used/reserved amounts are computed live from settlements (APOpenItems/APReservations); lines keep a snapshot of the debt.
-- No FK to gl_subledger_documents: ERP edit/delete removes debt rows for real, so reservations are guarded by
-- cleanSubledger (ERP) and subledgerMutation.checkReservations (GL) with gl_err_ap_prepare_reserved instead.
-- Lives in the GL schema because the GL reservation guard reads it on every subledger mutation.
CREATE TABLE IF NOT EXISTS ap_payment_prepare (
 company text NOT NULL,
 docno text NOT NULL CHECK (length(btrim(docno)) > 0),
 paydate date NOT NULL,
 partner_code text NOT NULL,
 branch_code text NOT NULL DEFAULT '',
 remark text NOT NULL DEFAULT '',
 projectcode text NOT NULL DEFAULT '',
 allocatecode text NOT NULL DEFAULT '',
 billgroupcode text NOT NULL DEFAULT '',
 departmentcode text NOT NULL DEFAULT '',
 total_amount numeric(18,2) NOT NULL DEFAULT 0,
 total_home_amount numeric(18,2) NOT NULL DEFAULT 0,
 discount_amount numeric(18,2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
 net_amount numeric(18,2) NOT NULL DEFAULT 0,
 exchange_profit numeric(18,2) NOT NULL DEFAULT 0,
 iscancel boolean NOT NULL DEFAULT false,
 cancel_reason text NOT NULL DEFAULT '',
 cancelled_at timestamptz,
 cancelled_by text NOT NULL DEFAULT '',
 version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
 created_at timestamptz NOT NULL,
 created_by text NOT NULL,
 modified_at timestamptz NOT NULL,
 modified_by text NOT NULL,
 PRIMARY KEY (company, docno),
 CHECK (NOT iscancel OR length(btrim(cancel_reason)) > 0)
);
CREATE INDEX IF NOT EXISTS ap_payment_prepare_partner ON ap_payment_prepare (company, partner_code, paydate);
CREATE TABLE IF NOT EXISTS ap_payment_prepare_lines (
 company text NOT NULL,
 docno text NOT NULL,
 linenumber int NOT NULL CHECK (linenumber > 0),
 debt_document_id text NOT NULL,
 document_no text NOT NULL,
 document_date date NOT NULL,
 debt_amount numeric(18,2) NOT NULL,
 billtype smallint NOT NULL,
 sign smallint NOT NULL DEFAULT 1 CHECK (sign IN (1, -1)),
 description text NOT NULL DEFAULT '',
 invbalance numeric(18,2) NOT NULL,
 payamount numeric(18,2) NOT NULL CHECK (payamount > 0),
 PRIMARY KEY (company, docno, linenumber),
 UNIQUE (company, docno, debt_document_id),
 FOREIGN KEY (company, docno) REFERENCES ap_payment_prepare (company, docno)
);
CREATE INDEX IF NOT EXISTS ap_payment_prepare_lines_debt ON ap_payment_prepare_lines (company, debt_document_id);
CREATE TABLE IF NOT EXISTS ap_payment_prepare_events (
 event_no bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 company text NOT NULL,
 docno text NOT NULL,
 action text NOT NULL CHECK (action IN ('create', 'update', 'cancel')),
 actor text NOT NULL,
 occurred_at timestamptz NOT NULL,
 payload jsonb NOT NULL
);
CREATE INDEX IF NOT EXISTS ap_payment_prepare_events_doc ON ap_payment_prepare_events (company, docno, event_no);

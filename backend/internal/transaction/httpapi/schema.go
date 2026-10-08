package httpapi

import (
	"context"
	"database/sql"
	"sync"
)

// ตารางเอกสารธุรกรรม ERP (doc / docdetail / docref) และสมุดสต็อก (stock_ledger / stock_period_balance)
// โครงสร้างตรงกับฐานที่ใช้งานอยู่ — ถ้าฐานของ holding ยังไม่มีตาราง (กลุ่มกิจการใหม่) จะสร้างให้ครั้งแรกที่ใช้
// ทุกชื่อเป็น lowercase ตามกฎของโปรเจ็กต์
const docSchema = `
CREATE TABLE IF NOT EXISTS doc (
	id serial PRIMARY KEY,
	businesscode text NOT NULL DEFAULT '',
	transflag integer,
	docno text,
	custcode text,
	custname text NOT NULL DEFAULT '',
	description text NOT NULL DEFAULT '',
	remark text NOT NULL DEFAULT '',
	docdatetime timestamptz,
	perioddatetime timestamptz,
	taxdocno text,
	totalamount numeric(18,8),
	roundamount numeric(18,8),
	paytype integer,
	paycashamount numeric(18,8),
	paycashchange numeric(18,8),
	paycashbalance numeric(18,8),
	deliverycode text,
	checksum char(32) NOT NULL,
	branchid text,
	slipurl text,
	salechannelcode text,
	deliveryamount numeric(18,8),
	iscancel boolean DEFAULT false,
	cancelreason text,
	guidpos text,
	guidbranch text,
	guidfixed text,
	isref boolean DEFAULT false,
	islocked boolean DEFAULT false,
	isclosed boolean DEFAULT false,
	iscomparedsuccess integer DEFAULT 0,
	currency text DEFAULT 'THB',
	currency_symbol text DEFAULT '฿',
	doc_currency text DEFAULT 'THB',
	doc_currency_symbol text DEFAULT '฿',
	exchange_rate numeric(18,8) DEFAULT 1,
	totalvalue_doc numeric(18,8) DEFAULT 0,
	totaldiscount_doc numeric(18,8) DEFAULT 0,
	totalvatvalue_doc numeric(18,8) DEFAULT 0,
	totalbeforevat_doc numeric(18,8) DEFAULT 0,
	totalaftervat_doc numeric(18,8) DEFAULT 0,
	totalamount_doc numeric(18,8) DEFAULT 0,
	creator_code text DEFAULT '',
	creator_name text DEFAULT '',
	created_at timestamptz DEFAULT now(),
	modifier_code text DEFAULT '',
	modifier_name text DEFAULT '',
	modified_at timestamptz DEFAULT now(),
	isdelete boolean DEFAULT false,
	approval_status text DEFAULT '',
	isclosedmanual boolean DEFAULT false,
	closedmanual_by_code text DEFAULT '',
	closedmanual_by_name text DEFAULT '',
	closedmanual_at timestamptz DEFAULT '1970-01-01 00:00:00+00',
	closedmanual_reason text DEFAULT ''
);
ALTER TABLE doc ADD COLUMN IF NOT EXISTS custname text NOT NULL DEFAULT '';
ALTER TABLE doc ADD COLUMN IF NOT EXISTS description text NOT NULL DEFAULT '';
ALTER TABLE doc ADD COLUMN IF NOT EXISTS remark text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_doc_company_doc ON doc (businesscode, transflag, docno);
CREATE INDEX IF NOT EXISTS idx_doc_transflag_docno_docdatetime ON doc (transflag, docno, docdatetime);
CREATE INDEX IF NOT EXISTS idx_doc_transflag_docno_perioddatetime ON doc (transflag, docno, perioddatetime);
CREATE INDEX IF NOT EXISTS idx_doc_created_at ON doc (created_at);
CREATE INDEX IF NOT EXISTS idx_doc_creator_code ON doc (creator_code);
CREATE INDEX IF NOT EXISTS idx_doc_approval_status ON doc (approval_status) WHERE approval_status <> '';

CREATE TABLE IF NOT EXISTS docdetail (
	id serial PRIMARY KEY,
	businesscode text NOT NULL DEFAULT '',
	docdatetime timestamptz,
	docno text,
	docref text,
	description text,
	linenumber integer,
	transflag integer,
	calcflag integer,
	calcseq integer,
	iscancel boolean DEFAULT false,
	itemcode text,
	barcodemain text,
	barcode text,
	unitcode text,
	whcode text,
	locationcode text,
	totalqty numeric(18,8),
	unitstand numeric(18,8),
	unitdivide numeric(18,8),
	price numeric(18,2),
	priceexcludevat numeric(18,2),
	sumamount numeric(18,2),
	isupdated boolean DEFAULT false,
	iscalcstock integer DEFAULT 0,
	price_doc numeric(18,2) DEFAULT 0,
	sumamount_doc numeric(18,2) DEFAULT 0,
	discountamount_doc numeric(18,2) DEFAULT 0,
	priceexcludevat_doc numeric(18,2) DEFAULT 0,
	sumamountexcludevat_doc numeric(18,2) DEFAULT 0,
	totalvaluevat_doc numeric(18,2) DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_docdetail_company_doc ON docdetail (businesscode, transflag, docno);
CREATE INDEX IF NOT EXISTS idx_docdetail_company_item ON docdetail (businesscode, itemcode);
CREATE INDEX IF NOT EXISTS idx_docdetail_docno ON docdetail (docno);
CREATE INDEX IF NOT EXISTS idx_docdetail_docref ON docdetail (docref);
CREATE INDEX IF NOT EXISTS idx_docdetail_docdatetime ON docdetail (docdatetime);
CREATE INDEX IF NOT EXISTS idx_docdetail_itemcode ON docdetail (itemcode);
CREATE INDEX IF NOT EXISTS idx_docdetail_itemcode_tr ON docdetail (itemcode, transflag);
CREATE INDEX IF NOT EXISTS idx_docdetail_barcode ON docdetail (barcode);
CREATE INDEX IF NOT EXISTS idx_docdetail_barcodemain ON docdetail (barcodemain);
CREATE INDEX IF NOT EXISTS idx_docdetail_isupdated ON docdetail (isupdated);

CREATE TABLE IF NOT EXISTS docref (
	id serial PRIMARY KEY,
	businesscode text NOT NULL DEFAULT '',
	docno text,
	docnotransflag integer,
	docnoref text,
	docnoreftransflag integer
);
CREATE INDEX IF NOT EXISTS idx_docref_company_doc ON docref (businesscode, docnotransflag, docno);
CREATE INDEX IF NOT EXISTS idx_docref_docno ON docref (docno);
CREATE INDEX IF NOT EXISTS idx_docref_docnoref ON docref (docnoref);

CREATE TABLE IF NOT EXISTS stock_ledger (
	businesscode text NOT NULL,
	itemcode text NOT NULL,
	whcode text NOT NULL,
	docdatetime timestamptz NOT NULL,
	behindindex integer NOT NULL,
	docno text NOT NULL,
	linenumber integer NOT NULL,
	periodkey char(7) NOT NULL,
	locationcode text NOT NULL DEFAULT '',
	barcode text NOT NULL DEFAULT '',
	unitcode text NOT NULL DEFAULT '',
	docref text NOT NULL DEFAULT '',
	transflag integer NOT NULL,
	direction smallint NOT NULL,
	docqty numeric(18,8) NOT NULL DEFAULT 0,
	unitstand numeric(18,8) NOT NULL DEFAULT 1,
	unitdivide numeric(18,8) NOT NULL DEFAULT 1,
	price numeric(18,8) NOT NULL DEFAULT 0,
	docvalue numeric(18,8) NOT NULL DEFAULT 0,
	qty numeric(18,8) NOT NULL,
	unitcost numeric(18,8) NOT NULL,
	amount numeric(18,8) NOT NULL,
	balanceqty numeric(18,8) NOT NULL,
	balanceamount numeric(18,8) NOT NULL,
	avgcost numeric(18,8) NOT NULL,
	calculatedat timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT stock_ledger_pk PRIMARY KEY (businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber),
	CONSTRAINT stock_ledger_avgcost_chk CHECK (avgcost >= 0),
	CONSTRAINT stock_ledger_direction_chk CHECK (direction IN (-1, 1)),
	CONSTRAINT stock_ledger_periodkey_chk CHECK (periodkey ~ '^[0-9]{4}-[0-9]{2}$'),
	CONSTRAINT stock_ledger_qty_chk CHECK (qty >= 0),
	CONSTRAINT stock_ledger_unitcost_chk CHECK (unitcost >= 0)
);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_doc ON stock_ledger (businesscode, docno);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_item ON stock_ledger (businesscode, itemcode);

CREATE TABLE IF NOT EXISTS stock_period_balance (
	businesscode text NOT NULL,
	itemcode text NOT NULL,
	whcode text NOT NULL,
	periodkey char(7) NOT NULL,
	closeqty numeric(18,8) NOT NULL,
	closeamount numeric(18,8) NOT NULL,
	closeavgcost numeric(18,8) NOT NULL,
	hastrans boolean NOT NULL DEFAULT false,
	updatedat timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT stock_period_balance_pk PRIMARY KEY (businesscode, itemcode, whcode, periodkey)
);
`

var (
	schemaMu    sync.Mutex
	schemaReady = map[*sql.DB]bool{}
)

// ensureSchema สร้างตารางเอกสาร/สต็อกในฐานของ holding ครั้งแรกที่ใช้ (idempotent, ล็อกด้วย advisory lock กัน race ข้าม instance)
func ensureSchema(ctx context.Context, db *sql.DB) error {
	schemaMu.Lock()
	defer schemaMu.Unlock()
	if schemaReady[db] {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('bc_erp_transaction_schema'))`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, docSchema); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	schemaReady[db] = true
	return nil
}

package cheque

import (
	"context"
	"database/sql"
	"sync"
)

// ตารางของโมดูลเช็คและธนาคาร (docs/kms/champ/design-2026-10-11/cheque-bank.md §4.2)
// ทุกชื่อเป็น lowercase snake_case; ฐานของ holding ที่ยังไม่มีตารางจะถูกสร้างให้ครั้งแรกที่ใช้ (ไม่มี migration ช่วง dev)
const schemaSQL = `
-- ทะเบียนรหัส 5 ชนิด (ตาม "กำหนดรหัสต่างๆ" myfrmbasemaster.cpp:229-236)
CREATE TABLE IF NOT EXISTS bank_codes (
	company     text         NOT NULL,
	kind        text         NOT NULL CHECK (kind IN ('bank','bank_branch','bank_account_type','bank_service','cheque_ext_status')),
	code        varchar(10)  NOT NULL CHECK (btrim(code) <> ''),
	name        varchar(50)  NOT NULL CHECK (btrim(name) <> ''),
	bank_code   varchar(10),
	user_group  varchar(255) NOT NULL DEFAULT '',
	is_active   boolean      NOT NULL DEFAULT true,
	version     bigint       NOT NULL DEFAULT 1 CHECK (version > 0),
	created_at  timestamptz  NOT NULL DEFAULT now(),
	created_by  text         NOT NULL,
	updated_at  timestamptz  NOT NULL DEFAULT now(),
	updated_by  text         NOT NULL,
	PRIMARY KEY (company, kind, code),
	CHECK (kind = 'bank_branch' OR bank_code IS NULL)
);
-- Champ ใช้ collation ไม่แยกตัวพิมพ์ใหญ่เล็ก: BBL กับ bbl ต้องซ้ำกัน
CREATE UNIQUE INDEX IF NOT EXISTS ux_bank_codes_upper ON bank_codes (company, kind, upper(code));

-- ผู้ติดต่อธนาคาร (BCBankContact; จอ 5680) แยกตามธนาคาร + สาขา
CREATE TABLE IF NOT EXISTS bank_contacts (
	company          text         NOT NULL,
	bank_code        varchar(10)  NOT NULL,
	bank_branch_code varchar(10)  NOT NULL,
	contact_code     varchar(10)  NOT NULL CHECK (btrim(contact_code) <> ''),
	line_no          integer      NOT NULL CHECK (line_no > 0),
	name             varchar(50)  NOT NULL CHECK (btrim(name) <> ''),
	position         varchar(50)  NOT NULL DEFAULT '',
	address          varchar(255) NOT NULL DEFAULT '',
	telephone        varchar(30)  NOT NULL DEFAULT '',
	is_contact       boolean      NOT NULL DEFAULT true,
	created_at       timestamptz  NOT NULL DEFAULT now(),
	created_by       text         NOT NULL,
	updated_at       timestamptz  NOT NULL DEFAULT now(),
	updated_by       text         NOT NULL,
	PRIMARY KEY (company, bank_code, bank_branch_code, contact_code)
);
CREATE INDEX IF NOT EXISTS idx_bank_contacts_bank ON bank_contacts (company, bank_code, bank_branch_code, line_no);

-- ทะเบียนเช็ค (BCChqIn + BCChqOut รวมตารางเดียว แยกด้วย cheque_type: 0 = เช็ครับ, 1 = เช็คจ่าย)
-- status เป็น cache ของ event สถานะล่าสุดที่ยังไม่ถูกยกเลิก (คำนวณใหม่ได้ด้วย recalculate)
CREATE TABLE IF NOT EXISTS cheque_registers (
	company          text           NOT NULL,
	id               uuid           NOT NULL,
	cheque_type      smallint       NOT NULL CHECK (cheque_type IN (0,1)),
	doc_no           varchar(30)    NOT NULL CHECK (btrim(doc_no) <> ''),
	doc_date         date           NOT NULL,
	due_date         date           NOT NULL,
	partner_code     text           NOT NULL CHECK (btrim(partner_code) <> ''),
	cheque_no        varchar(20)    NOT NULL CHECK (btrim(cheque_no) <> ''),
	bank_code        varchar(10)    NOT NULL,
	bank_branch_code varchar(10)    NOT NULL,
	book_code        text,
	status           smallint       NOT NULL DEFAULT 0,
	ext_status_code  varchar(10),
	currency_code    varchar(3)     NOT NULL DEFAULT 'THB',
	exchange_rate    numeric(19,10) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
	amount           numeric(18,2)  NOT NULL CHECK (amount > 0),
	home_amount      numeric(18,2)  NOT NULL CHECK (home_amount > 0),
	description      varchar(255)   NOT NULL DEFAULT '',
	sale_code        text           NOT NULL DEFAULT '',
	department_code  text           NOT NULL DEFAULT '',
	source_kind      text           NOT NULL DEFAULT 'register' CHECK (source_kind IN ('register','erp','opening')),
	prepare_doc_no   varchar(30),
	print_count      integer        NOT NULL DEFAULT 0 CHECK (print_count >= 0),
	version          bigint         NOT NULL DEFAULT 1 CHECK (version > 0),
	created_at       timestamptz    NOT NULL DEFAULT now(),
	created_by       text           NOT NULL,
	updated_at       timestamptz    NOT NULL DEFAULT now(),
	updated_by       text           NOT NULL,
	-- ลบเช็คที่ไม่มีเอกสารใดใช้งานอยู่ (Champ cqfrmcheqin.cpp:438-470 ลบแถวจริง): BC เก็บแถว+ประวัติไว้เป็นหลักฐาน แต่ทุกรายการ/ยอด/การเลือกไม่นับ
	deleted_at       timestamptz,
	deleted_by       text,
	PRIMARY KEY (company, id),
	CHECK ((deleted_at IS NULL) = (deleted_by IS NULL)),
	CHECK ((cheque_type = 0 AND status BETWEEN 0 AND 6) OR (cheque_type = 1 AND status BETWEEN 0 AND 3)),
	CHECK (cheque_type = 0 OR (book_code IS NOT NULL AND doc_date <= due_date)),
	-- เช็ครับสถานะ <> 0 ต้องมีสมุด ยกเว้นยกเลิกเช็คในมือ (423 ชนิด 0 ไม่ใช้สมุด cqfrmcheqincancel.cpp)
	-- และขายลด (424 ไม่บังคับสมุด: การตรวจถูก comment ออกที่ cqfrmcheqinsale.cpp:374-382)
	CONSTRAINT cheque_registers_in_book_check CHECK (cheque_type = 1 OR status IN (0,4,5) OR book_code IS NOT NULL),
	CHECK (cheque_type = 0 OR sale_code = '')
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_in_no  ON cheque_registers (company, cheque_no, bank_code, bank_branch_code) WHERE cheque_type = 0 AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_out_no ON cheque_registers (company, cheque_no, book_code) WHERE cheque_type = 1 AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_doc_no ON cheque_registers (company, cheque_type, doc_no) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_cheque_select  ON cheque_registers (company, cheque_type, status, due_date);
CREATE INDEX IF NOT EXISTS ix_cheque_partner ON cheque_registers (company, cheque_type, partner_code);
CREATE INDEX IF NOT EXISTS ix_cheque_book    ON cheque_registers (company, book_code);
-- ตารางที่สร้างก่อน 2026-10-11 มี CHECK สมุดของเช็ครับแบบไม่มีชื่อ (status IN (0,4)) ซึ่ง CREATE TABLE IF NOT EXISTS ไม่แก้ให้:
-- ลบตัวเดิม (หาจากเนื้อหา เพราะชื่ออัตโนมัติขึ้นกับลำดับ CHECK) แล้วใส่ตัวที่มีชื่อ — ทำครั้งเดียว รอบถัดไปข้ามทั้งก้อน
DO $$
DECLARE old_check text;
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'cheque_registers'::regclass AND conname = 'cheque_registers_in_book_check') THEN
		FOR old_check IN SELECT conname FROM pg_constraint WHERE conrelid = 'cheque_registers'::regclass AND contype = 'c'
			AND pg_get_constraintdef(oid) LIKE '%book_code IS NOT NULL%' AND pg_get_constraintdef(oid) NOT LIKE '%doc_date%'
		LOOP
			EXECUTE format('ALTER TABLE cheque_registers DROP CONSTRAINT %I', old_check);
		END LOOP;
		ALTER TABLE cheque_registers ADD CONSTRAINT cheque_registers_in_book_check CHECK (cheque_type = 1 OR status IN (0,4,5) OR book_code IS NOT NULL);
	END IF;
END $$;

-- ประวัติของเช็ค (append-only; ยกเลิกด้วย voided_*): edit = แก้ฟิลด์หลักหลังยืนยัน, manual_status = แก้สถานะในจอทะเบียน (D8)
CREATE TABLE IF NOT EXISTS cheque_events (
	company     text        NOT NULL,
	id          bigserial,
	cheque_id   uuid        NOT NULL,
	event_kind  text        NOT NULL CHECK (event_kind IN ('create','update','edit','manual_status','status_doc','print','ext_status','delete')),
	doc_flag    integer,
	doc_no      varchar(30),
	event_date  date        NOT NULL,
	from_status smallint,
	to_status   smallint,
	book_code   text,
	detail      jsonb       NOT NULL DEFAULT '{}'::jsonb,
	reason      text        NOT NULL DEFAULT '',
	actor       text        NOT NULL,
	created_at  timestamptz NOT NULL DEFAULT now(),
	voided_at   timestamptz,
	voided_by   text,
	void_reason text,
	PRIMARY KEY (company, id),
	FOREIGN KEY (company, cheque_id) REFERENCES cheque_registers (company, id),
	CHECK ((voided_at IS NULL AND voided_by IS NULL AND void_reason IS NULL) OR (voided_at IS NOT NULL AND voided_by IS NOT NULL AND btrim(void_reason) <> ''))
);
CREATE INDEX IF NOT EXISTS ix_cheque_events_cheque ON cheque_events (company, cheque_id, event_date DESC, id DESC) WHERE voided_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_cheque_events_doc ON cheque_events (company, doc_flag, doc_no) WHERE doc_flag IS NOT NULL;

-- เอกสารเปลี่ยนสถานะ 8 ตัว (419..426)
CREATE TABLE IF NOT EXISTS cheque_docs (
	company       text          NOT NULL,
	transflag     integer       NOT NULL CHECK (transflag BETWEEN 419 AND 426),
	doc_no        varchar(30)   NOT NULL CHECK (btrim(doc_no) <> ''),
	doc_date      date          NOT NULL,
	branch_code   text          NOT NULL,
	book_code     text,
	cancel_type   smallint      CHECK (cancel_type BETWEEN 0 AND 3),
	cancel_passed boolean       NOT NULL DEFAULT false,
	description   varchar(255)  NOT NULL DEFAULT '',
	total_amount  numeric(18,2) NOT NULL CHECK (total_amount > 0),
	fee_amount    numeric(18,2) NOT NULL DEFAULT 0 CHECK (fee_amount >= 0),
	net_amount    numeric(18,2) NOT NULL,
	gl_status     text          NOT NULL DEFAULT 'none' CHECK (gl_status IN ('none','posted','no_mapping','not_required')),
	status        text          NOT NULL DEFAULT 'active' CHECK (status IN ('active','cancelled')),
	cancel_reason text,
	cancelled_at  timestamptz,
	cancelled_by  text,
	version       bigint        NOT NULL DEFAULT 1 CHECK (version > 0),
	created_at    timestamptz   NOT NULL DEFAULT now(),
	created_by    text          NOT NULL,
	updated_at    timestamptz   NOT NULL DEFAULT now(),
	updated_by    text          NOT NULL,
	PRIMARY KEY (company, transflag, doc_no),
	CHECK (net_amount = total_amount - fee_amount),
	CHECK (fee_amount <= total_amount),
	CHECK (transflag <> 419 OR fee_amount = 0),
	CHECK ((transflag = 423) = (cancel_type IS NOT NULL)),
	CHECK (transflag = 426 OR cancel_passed = false),
	CHECK (status = 'active' OR (cancelled_at IS NOT NULL AND btrim(cancel_reason) <> ''))
);
CREATE INDEX IF NOT EXISTS ix_cheque_docs_date ON cheque_docs (company, transflag, doc_date DESC, doc_no DESC);
CREATE TABLE IF NOT EXISTS cheque_doc_lines (
	company     text          NOT NULL,
	transflag   integer       NOT NULL,
	doc_no      varchar(30)   NOT NULL,
	line_no     integer       NOT NULL CHECK (line_no > 0),
	cheque_id   uuid          NOT NULL,
	from_status smallint      NOT NULL,
	amount      numeric(18,2) NOT NULL CHECK (amount > 0),
	fee_amount  numeric(18,2) NOT NULL DEFAULT 0 CHECK (fee_amount >= 0 AND fee_amount <= amount),
	net_amount  numeric(18,2) NOT NULL,
	PRIMARY KEY (company, transflag, doc_no, line_no),
	FOREIGN KEY (company, transflag, doc_no) REFERENCES cheque_docs (company, transflag, doc_no),
	FOREIGN KEY (company, cheque_id) REFERENCES cheque_registers (company, id),
	UNIQUE (company, transflag, doc_no, cheque_id),
	CHECK (net_amount = amount - fee_amount),
	CHECK (transflag <> 419 OR fee_amount = 0)
);
CREATE INDEX IF NOT EXISTS ix_cheque_doc_lines_cheque ON cheque_doc_lines (company, cheque_id);
-- ใบสำคัญ GL ต่อรอบแก้ไขของเอกสาร (แบบเดียวกับ erp_doc_journals)
CREATE TABLE IF NOT EXISTS cheque_doc_journals (
	company       text        NOT NULL,
	transflag     integer     NOT NULL,
	doc_no        varchar(30) NOT NULL,
	revision      integer     NOT NULL CHECK (revision > 0),
	journal_id    text        NOT NULL,
	journal_docno text        NOT NULL,
	voided_at     timestamptz,
	created_at    timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (company, transflag, doc_no, revision)
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cheque_doc_journals_active ON cheque_doc_journals (company, transflag, doc_no) WHERE voided_at IS NULL;

-- ลิงก์เช็ค ↔ เอกสารต้นทาง (ใบเสร็จ/ใบจ่าย/รายวัน) แทน BCRecMoney/BCPayMoney + Balance — เขียนโดยงาน CB-14
CREATE TABLE IF NOT EXISTS cheque_source_links (
	company       text          NOT NULL,
	id            uuid          NOT NULL,
	cheque_id     uuid          NOT NULL,
	source_kind   text          NOT NULL CHECK (source_kind IN ('erp_doc','gl_journal')),
	source_flag   integer,
	source_doc_no text          NOT NULL,
	source_date   date          NOT NULL,
	amount        numeric(18,2) NOT NULL CHECK (amount > 0),
	created_at    timestamptz   NOT NULL DEFAULT now(),
	created_by    text          NOT NULL,
	voided_at     timestamptz,
	voided_by     text,
	void_reason   text,
	PRIMARY KEY (company, id),
	FOREIGN KEY (company, cheque_id) REFERENCES cheque_registers (company, id)
);
CREATE INDEX IF NOT EXISTS ix_cheque_links_cheque ON cheque_source_links (company, cheque_id) WHERE voided_at IS NULL;

-- requestid ของคำสั่งเขียน (กันกดบันทึกซ้ำ): คำขอซ้ำได้ผลเดิมโดยไม่เขียนอะไรเพิ่ม
CREATE TABLE IF NOT EXISTS cheque_requests (
	company    text        NOT NULL,
	request_id varchar(80) NOT NULL,
	operation  text        NOT NULL,
	result     jsonb,
	created_by text        NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (company, request_id)
);
`

var (
	schemaMu    sync.Mutex
	schemaReady = map[*sql.DB]bool{}
)

// EnsureSchema สร้างตารางของโมดูลเช็ค/ธนาคารในฐานของ holding ครั้งแรกที่ใช้
// (idempotent; ล็อกด้วย advisory lock กัน race ข้าม instance)
func EnsureSchema(ctx context.Context, db *sql.DB) error {
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
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('bc_cheque_bank_schema'))`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, schemaSQL); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	schemaReady[db] = true
	return nil
}

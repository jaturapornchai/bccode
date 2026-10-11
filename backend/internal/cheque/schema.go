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

package build

import (
	"context"
	"database/sql"
	"fmt"

	"smlcloudplatform/internal/goapi/logger"
)

// Stock & Cost Engine v2 — ดู ADR docs/kms/decisions/2026-09-15-stock-cost-engine-v2.md
//
// หลักการที่ตารางชุดนี้บังคับไว้ในตัวเอง (ไม่พึ่งให้โค้ดจำ):
//   - PRIMARY KEY เป็นคีย์ทางธุรกิจจริง = ลำดับการคิดต้นทุน จึงมีแถวซ้ำไม่ได้ และ UPSERT ได้
//   - businesscode อยู่ในคีย์ตั้งแต่ต้น การลบข้ามบริษัทจึงเป็นไปไม่ได้
//   - CHECK constraint กันค่าที่เป็นไปไม่ได้ทางบัญชี (จำนวนติดลบ ต้นทุนติดลบ ทิศทางนอกเหนือ ±1)

// TableStockLedgerCreate สร้างตาราง stock_ledger — สมุดเคลื่อนไหวสต็อกพร้อมต้นทุนที่คิดแล้ว
// แทนที่ processstockcost เดิมที่ PK เป็น id SERIAL จึงไม่มีทาง UPSERT และต้อง DELETE ทิ้งก่อนเสมอ
func TableStockLedgerCreate(db *sql.DB) error {
	logger.Info("Creating stock_ledger table")

	const createTableQuery = `
		CREATE TABLE IF NOT EXISTS stock_ledger (
			businesscode  TEXT NOT NULL,
			itemcode      TEXT NOT NULL,
			whcode        TEXT NOT NULL,
			docdatetime   TIMESTAMPTZ NOT NULL,
			behindindex   INT NOT NULL,
			docno         TEXT NOT NULL,
			linenumber    INT NOT NULL,
			periodkey     CHAR(7) NOT NULL,
			locationcode  TEXT NOT NULL DEFAULT '',
			barcode       TEXT NOT NULL DEFAULT '',
			unitcode      TEXT NOT NULL DEFAULT '',
			docref        TEXT NOT NULL DEFAULT '',
			transflag     INT NOT NULL,
			direction     SMALLINT NOT NULL,
			docqty        NUMERIC(18,8) NOT NULL DEFAULT 0,
			unitstand     NUMERIC(18,8) NOT NULL DEFAULT 1,
			unitdivide    NUMERIC(18,8) NOT NULL DEFAULT 1,
			price         NUMERIC(18,8) NOT NULL DEFAULT 0,
			docvalue      NUMERIC(18,8) NOT NULL DEFAULT 0,
			qty           NUMERIC(18,8) NOT NULL,
			unitcost      NUMERIC(18,8) NOT NULL,
			amount        NUMERIC(18,8) NOT NULL,
			balanceqty    NUMERIC(18,8) NOT NULL,
			balanceamount NUMERIC(18,8) NOT NULL,
			avgcost       NUMERIC(18,8) NOT NULL,
			calculatedat  TIMESTAMPTZ NOT NULL DEFAULT now(),
			CONSTRAINT stock_ledger_pk PRIMARY KEY (businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber),
			CONSTRAINT stock_ledger_direction_chk CHECK (direction IN (-1, 1)),
			CONSTRAINT stock_ledger_qty_chk CHECK (qty >= 0),
			CONSTRAINT stock_ledger_unitcost_chk CHECK (unitcost >= 0),
			CONSTRAINT stock_ledger_avgcost_chk CHECK (avgcost >= 0),
			CONSTRAINT stock_ledger_periodkey_chk CHECK (periodkey ~ '^[0-9]{4}-[0-9]{2}$')
		)`

	const commentsAndIndexes = `
		COMMENT ON TABLE stock_ledger IS 'สมุดเคลื่อนไหวสต็อกพร้อมต้นทุนที่คำนวณแล้ว (แทน processstockcost)';
		COMMENT ON COLUMN stock_ledger.behindindex IS 'ลำดับเอกสารภายในวันเดียวกัน ส่วนหนึ่งของคีย์เรียงลำดับการคิดต้นทุน';
		COMMENT ON COLUMN stock_ledger.periodkey IS 'งวดบัญชีรูปแบบ YYYY-MM ตัดตามเขตเวลาธุรกิจ คำนวณโดย stockengine.PeriodKeyOf';
		COMMENT ON COLUMN stock_ledger.direction IS 'ทิศทาง: 1 = รับเข้า, -1 = จ่ายออก';
		COMMENT ON COLUMN stock_ledger.docqty IS 'จำนวนตามหน่วยนับในเอกสาร (qty คือจำนวนหน่วยฐานที่ใช้คิดต้นทุน)';
		COMMENT ON COLUMN stock_ledger.price IS 'ราคาต่อหน่วยตามเอกสาร';
		COMMENT ON COLUMN stock_ledger.docvalue IS 'มูลค่าตามเอกสารของบรรทัดนี้ (ยอดขายสำหรับขาออก ยอดซื้อสำหรับขาเข้า)';
		COMMENT ON COLUMN stock_ledger.balanceqty IS 'ยอดคงเหลือสะสมหลังรายการนี้';
		COMMENT ON COLUMN stock_ledger.avgcost IS 'ต้นทุนถัวเฉลี่ยถ่วงน้ำหนักหลังรายการนี้';

		CREATE INDEX IF NOT EXISTS idx_stock_ledger_period ON stock_ledger (businesscode, periodkey, itemcode, whcode);
		CREATE INDEX IF NOT EXISTS idx_stock_ledger_item ON stock_ledger (businesscode, itemcode, docdatetime);
		CREATE INDEX IF NOT EXISTS idx_stock_ledger_docno ON stock_ledger (businesscode, docno);
		CREATE INDEX IF NOT EXISTS idx_stock_ledger_transflag ON stock_ledger (businesscode, transflag, docdatetime);
		`

	return execInTx(db, "stock_ledger", createTableQuery, commentsAndIndexes)
}

// TableStockPeriodBalanceCreate สร้างตาราง stock_period_balance — ยอดคงเหลือ ณ สิ้นงวด
// ทำให้การคิดต้นทุนใหม่เริ่มจากงวดที่กระทบ ไม่ต้องไล่ตั้งแต่รายการแรกของสินค้า
// เทียบต้นแบบ Champ: bcstkwarehouseperiod ที่เก็บ 61 งวดเป็นคอลัมน์กว้าง ที่นี่เก็บเป็นแถว
func TableStockPeriodBalanceCreate(db *sql.DB) error {
	logger.Info("Creating stock_period_balance table")

	const createTableQuery = `
		CREATE TABLE IF NOT EXISTS stock_period_balance (
			businesscode TEXT NOT NULL,
			itemcode     TEXT NOT NULL,
			whcode       TEXT NOT NULL,
			periodkey    CHAR(7) NOT NULL,
			closeqty     NUMERIC(18,8) NOT NULL,
			closeamount  NUMERIC(18,8) NOT NULL,
			closeavgcost NUMERIC(18,8) NOT NULL,
			hastrans     BOOLEAN NOT NULL DEFAULT FALSE,
			updatedat    TIMESTAMPTZ NOT NULL DEFAULT now(),
			CONSTRAINT stock_period_balance_pk PRIMARY KEY (businesscode, itemcode, whcode, periodkey),
			CONSTRAINT stock_period_balance_avgcost_chk CHECK (closeavgcost >= 0),
			CONSTRAINT stock_period_balance_periodkey_chk CHECK (periodkey ~ '^[0-9]{4}-[0-9]{2}$')
		)`

	const commentsAndIndexes = `
		COMMENT ON TABLE stock_period_balance IS 'ยอดคงเหลือสต็อก ณ สิ้นงวด ใช้เป็นจุดตั้งต้นการคิดต้นทุนใหม่';
		COMMENT ON COLUMN stock_period_balance.hastrans IS 'งวดนี้มีรายการเคลื่อนไหวจริงหรือเป็นยอดยกมาจากงวดก่อน';

		CREATE INDEX IF NOT EXISTS idx_stock_period_balance_period ON stock_period_balance (businesscode, periodkey);
		`

	return execInTx(db, "stock_period_balance", createTableQuery, commentsAndIndexes)
}

// TableStockDirtyCreate สร้างตาราง stock_dirty — สมุดงานค้างที่ยุบงานซ้ำตั้งแต่ต้นทาง
// แทน timer ของระบบต้นแบบ: เอกสารหลายใบที่แตะสินค้าเดียวกันเหลือแถวเดียว
// โดย fromdate เก็บวันที่เก่าสุดที่กระทบ (ดู stockengine.MarkDirty)
func TableStockDirtyCreate(db *sql.DB) error {
	logger.Info("Creating stock_dirty table")

	const createTableQuery = `
		CREATE TABLE IF NOT EXISTS stock_dirty (
			businesscode TEXT NOT NULL,
			itemcode     TEXT NOT NULL,
			fromdate     TIMESTAMPTZ NOT NULL,
			reason       TEXT NOT NULL DEFAULT 'doc',
			enqueuedat   TIMESTAMPTZ NOT NULL DEFAULT now(),
			markedat     TIMESTAMPTZ NOT NULL DEFAULT now(),
			attempts     INT NOT NULL DEFAULT 0,
			lasterror    TEXT NOT NULL DEFAULT '',
			leaseowner   TEXT,
			leaseuntil   TIMESTAMPTZ,
			CONSTRAINT stock_dirty_pk PRIMARY KEY (businesscode, itemcode)
		)`

	const commentsAndIndexes = `
		COMMENT ON TABLE stock_dirty IS 'รายการสินค้าที่รอคำนวณต้นทุนใหม่ (แทนการวน timer ของระบบเดิม)';
		COMMENT ON COLUMN stock_dirty.fromdate IS 'วันเวลาที่เก่าสุดที่ถูกกระทบ คิดใหม่ตั้งแต่จุดนี้เท่านั้น';
		COMMENT ON COLUMN stock_dirty.enqueuedat IS 'เวลาที่เข้าคิวครั้งแรก ใช้เรียงลำดับก่อนหลัง ไม่เลื่อนตามการแตะครั้งหลัง';
		COMMENT ON COLUMN stock_dirty.markedat IS 'เวลาที่ถูกแตะล่าสุด ใช้ดูว่ามีเอกสารเข้ามาใหม่ระหว่างกำลังคำนวณหรือไม่';
		COMMENT ON COLUMN stock_dirty.leaseuntil IS 'เวลาหมดอายุการจองงาน หมดอายุแล้วให้ตัวอื่นหยิบไปทำต่อได้';

		CREATE INDEX IF NOT EXISTS idx_stock_dirty_ready ON stock_dirty (enqueuedat) WHERE leaseuntil IS NULL;
		`

	return execInTx(db, "stock_dirty", createTableQuery, commentsAndIndexes)
}

// TableStockDeadLetterCreate สร้างตาราง stock_dead_letter — งานคิดต้นทุนที่ล้มเหลวจนเลิกลองแล้ว
// แยกจากคิวงานเสียของเอกสาร เพราะงานที่นี่เป็นระดับสินค้า ไม่มีเลขที่เอกสารกำกับ
func TableStockDeadLetterCreate(db *sql.DB) error {
	logger.Info("Creating stock_dead_letter table")

	const createTableQuery = `
		CREATE TABLE IF NOT EXISTS stock_dead_letter (
			id           BIGSERIAL PRIMARY KEY,
			businesscode TEXT NOT NULL,
			itemcode     TEXT NOT NULL,
			fromdate     TIMESTAMPTZ NOT NULL,
			attempts     INT NOT NULL DEFAULT 0,
			lasterror    TEXT NOT NULL DEFAULT '',
			failedat     TIMESTAMPTZ NOT NULL DEFAULT now()
		)`

	const commentsAndIndexes = `
		COMMENT ON TABLE stock_dead_letter IS 'งานคิดต้นทุนสินค้าที่ล้มเหลวครบจำนวนครั้งแล้ว รอคนตรวจสอบ';
		COMMENT ON COLUMN stock_dead_letter.fromdate IS 'วันเวลาที่ต้องคิดใหม่ตั้งแต่ ใช้ตอนสั่งคิดซ้ำหลังแก้ต้นเหตุ';

		CREATE INDEX IF NOT EXISTS idx_stock_dead_letter_item ON stock_dead_letter (businesscode, itemcode);
		`

	return execInTx(db, "stock_dead_letter", createTableQuery, commentsAndIndexes)
}

// execInTx รัน DDL สองก้อน (สร้างตาราง แล้วตามด้วย comment/index) ในทรานแซกชันเดียว
func execInTx(db *sql.DB, table, createQuery, extraQuery string) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, createQuery); err != nil {
		return fmt.Errorf("create %s table: %w", table, err)
	}
	if _, err := tx.ExecContext(ctx, extraQuery); err != nil {
		return fmt.Errorf("create %s comments and indexes: %w", table, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit %s: %w", table, err)
	}

	logger.Info("✅ %s table created", table)
	return nil
}

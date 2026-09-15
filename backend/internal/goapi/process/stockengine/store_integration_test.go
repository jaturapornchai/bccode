//go:build integration

package stockengine

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var testTransFlags = []int{54, 12, 310, 48, 60, 58, 66, 44, 16, 56, 68, 72}

// engineTestDB เตรียม schema แยกต่อหนึ่งเทสต์ พร้อมตารางจริงที่ engine ใช้
func engineTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("BC_STOCK_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_STOCK_TEST_POSTGRES_DSN to an isolated PostgreSQL")
	}

	schema := "bc_stock_test_" + fmt.Sprintf("%x", uuid.New().ID())
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`); err != nil {
			t.Error(err)
		}
		admin.Close()
	})

	scoped, err := sql.Open("postgres", dsn+"&search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { scoped.Close() })

	createSchema(t, scoped)
	return scoped
}

func createSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE doc (
			businesscode TEXT NOT NULL DEFAULT '', docno TEXT, iscancel BOOLEAN DEFAULT FALSE)`,
		`CREATE TABLE docdetail (
			id SERIAL PRIMARY KEY, businesscode TEXT NOT NULL DEFAULT '',
			docdatetime TIMESTAMPTZ, docno TEXT, docref TEXT, linenumber INT,
			transflag INT, behindindex INT NOT NULL DEFAULT 0, itemcode TEXT,
			barcode TEXT, unitcode TEXT, whcode TEXT, locationcode TEXT,
			totalqty NUMERIC(18,8), unitstand NUMERIC(18,8), unitdivide NUMERIC(18,8),
			price NUMERIC(18,2), priceexcludevat NUMERIC(18,2), sumamount NUMERIC(18,2))`,
		`CREATE TABLE stock_ledger (
			businesscode TEXT NOT NULL, itemcode TEXT NOT NULL, whcode TEXT NOT NULL,
			docdatetime TIMESTAMPTZ NOT NULL, behindindex INT NOT NULL, docno TEXT NOT NULL,
			linenumber INT NOT NULL,
			periodkey CHAR(7) NOT NULL,
			locationcode TEXT NOT NULL DEFAULT '', barcode TEXT NOT NULL DEFAULT '',
			unitcode TEXT NOT NULL DEFAULT '', docref TEXT NOT NULL DEFAULT '',
			transflag INT NOT NULL, direction SMALLINT NOT NULL,
			qty NUMERIC(18,8) NOT NULL, unitcost NUMERIC(18,8) NOT NULL, amount NUMERIC(18,8) NOT NULL,
			balanceqty NUMERIC(18,8) NOT NULL, balanceamount NUMERIC(18,8) NOT NULL,
			avgcost NUMERIC(18,8) NOT NULL, calculatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
			CONSTRAINT stock_ledger_pk PRIMARY KEY (businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber),
			CONSTRAINT stock_ledger_direction_chk CHECK (direction IN (-1, 1)),
			CONSTRAINT stock_ledger_qty_chk CHECK (qty >= 0),
			CONSTRAINT stock_ledger_unitcost_chk CHECK (unitcost >= 0),
			CONSTRAINT stock_ledger_avgcost_chk CHECK (avgcost >= 0),
			CONSTRAINT stock_ledger_periodkey_chk CHECK (periodkey ~ '^[0-9]{4}-[0-9]{2}$'))`,
		`CREATE TABLE stock_period_balance (
			businesscode TEXT NOT NULL, itemcode TEXT NOT NULL, whcode TEXT NOT NULL,
			periodkey CHAR(7) NOT NULL, closeqty NUMERIC(18,8) NOT NULL,
			closeamount NUMERIC(18,8) NOT NULL, closeavgcost NUMERIC(18,8) NOT NULL,
			hastrans BOOLEAN NOT NULL DEFAULT FALSE, updatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
			CONSTRAINT stock_period_balance_pk PRIMARY KEY (businesscode, itemcode, whcode, periodkey),
			CONSTRAINT stock_period_balance_avgcost_chk CHECK (closeavgcost >= 0))`,
		`CREATE TABLE stock_dirty (
			businesscode TEXT NOT NULL, itemcode TEXT NOT NULL, fromdate TIMESTAMPTZ NOT NULL,
			reason TEXT NOT NULL DEFAULT 'doc', enqueuedat TIMESTAMPTZ NOT NULL DEFAULT now(),
			attempts INT NOT NULL DEFAULT 0, lasterror TEXT NOT NULL DEFAULT '',
			leaseowner TEXT, leaseuntil TIMESTAMPTZ,
			CONSTRAINT stock_dirty_pk PRIMARY KEY (businesscode, itemcode))`,
		`CREATE TABLE deadletterqueue (
			id BIGSERIAL PRIMARY KEY, holdingcode VARCHAR(100) NOT NULL, docno VARCHAR(100) NOT NULL,
			transflag VARCHAR(10) NOT NULL, retrycount INTEGER NOT NULL DEFAULT 0,
			createdat TIMESTAMP NOT NULL, failedat TIMESTAMP NOT NULL DEFAULT NOW(),
			errormessage TEXT NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create schema: %v\n%s", err, statement)
		}
	}
}

type docLine struct {
	docNo       string
	line        int
	transFlag   int
	qty         float64
	sumAmount   float64
	when        time.Time
	behindIndex int
	whCode      string
	cancelled   bool
}

func insertDoc(t *testing.T, db *sql.DB, business, item string, lines ...docLine) {
	t.Helper()
	seen := map[string]bool{}
	for _, l := range lines {
		wh := l.whCode
		if wh == "" {
			wh = "WH01"
		}
		if !seen[l.docNo] {
			if _, err := db.Exec(`INSERT INTO doc (businesscode, docno, iscancel) VALUES ($1,$2,$3)`,
				business, l.docNo, l.cancelled); err != nil {
				t.Fatalf("insert doc: %v", err)
			}
			seen[l.docNo] = true
		}
		if _, err := db.Exec(`
			INSERT INTO docdetail (businesscode, docdatetime, docno, linenumber, transflag, behindindex,
				itemcode, whcode, unitcode, totalqty, unitstand, unitdivide, price, priceexcludevat, sumamount)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'EA',$9,1,1,0,0,$10)`,
			business, l.when, l.docNo, l.line, l.transFlag, l.behindIndex,
			item, wh, l.qty, l.sumAmount); err != nil {
			t.Fatalf("insert docdetail: %v", err)
		}
	}
}

func at(day int) time.Time {
	return time.Date(2026, 3, day, 9, 0, 0, 0, time.UTC)
}

func ledgerSnapshot(t *testing.T, db *sql.DB, business, item string) []LedgerRow {
	t.Helper()
	rows, err := db.Query(`
		SELECT whcode, docdatetime, behindindex, docno, linenumber, direction, qty, unitcost, amount,
		       balanceqty, balanceamount, avgcost
		FROM stock_ledger WHERE businesscode = $1 AND itemcode = $2
		ORDER BY docdatetime, behindindex, docno, linenumber`, business, item)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	defer rows.Close()

	var out []LedgerRow
	for rows.Next() {
		var r LedgerRow
		if err := rows.Scan(&r.WhCode, &r.DocDateTime, &r.BehindIndex, &r.DocNo, &r.LineNumber,
			&r.Direction, &r.Qty, &r.UnitCost, &r.Amount, &r.BalanceQty, &r.BalanceAmount, &r.AvgCost); err != nil {
			t.Fatalf("scan ledger: %v", err)
		}
		out = append(out, r)
	}
	return out
}

func TestRecalculateWritesLedgerAndPeriodBalance(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	insertDoc(t, db, "BC01", "ITEM01",
		docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 10, sumAmount: 1000, when: at(1)},
		docLine{docNo: "PU002", line: 1, transFlag: 12, qty: 10, sumAmount: 1400, when: at(2)},
		docLine{docNo: "SA001", line: 1, transFlag: 44, qty: 5, when: at(3)},
	)

	scope := Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: at(1)}
	count, err := Recalculate(ctx, db, scope, testTransFlags, DefaultOptions())
	if err != nil {
		t.Fatalf("recalculate: %v", err)
	}
	if count != 3 {
		t.Fatalf("rows written = %d, want 3", count)
	}

	ledger := ledgerSnapshot(t, db, "BC01", "ITEM01")
	if len(ledger) != 3 {
		t.Fatalf("ledger rows = %d, want 3", len(ledger))
	}
	if ledger[2].UnitCost != 120 {
		t.Errorf("sale unit cost = %v, want 120", ledger[2].UnitCost)
	}
	if ledger[2].BalanceQty != 15 {
		t.Errorf("final balance = %v, want 15", ledger[2].BalanceQty)
	}

	var qty, amount float64
	if err := db.QueryRow(`
		SELECT closeqty, closeamount FROM stock_period_balance
		WHERE businesscode='BC01' AND itemcode='ITEM01' AND whcode='WH01' AND periodkey='2026-03'`).
		Scan(&qty, &amount); err != nil {
		t.Fatalf("read period balance: %v", err)
	}
	if qty != 15 || amount != 1800 {
		t.Errorf("period balance = (%v, %v), want (15, 1800)", qty, amount)
	}
}

// คิดซ้ำต้องได้ผลเท่าเดิม ไม่เกิดแถวซ้ำ เพราะคีย์หลักเป็นคีย์ทางธุรกิจจริง
func TestRecalculateIsIdempotent(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	insertDoc(t, db, "BC01", "ITEM01",
		docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 10, sumAmount: 1000, when: at(1)},
		docLine{docNo: "SA001", line: 1, transFlag: 44, qty: 4, when: at(2)},
	)
	scope := Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: at(1)}

	if _, err := Recalculate(ctx, db, scope, testTransFlags, DefaultOptions()); err != nil {
		t.Fatalf("first run: %v", err)
	}
	first := ledgerSnapshot(t, db, "BC01", "ITEM01")

	for round := 0; round < 3; round++ {
		if _, err := Recalculate(ctx, db, scope, testTransFlags, DefaultOptions()); err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		again := ledgerSnapshot(t, db, "BC01", "ITEM01")
		if len(again) != len(first) {
			t.Fatalf("round %d produced %d rows, want %d", round, len(again), len(first))
		}
		for i := range first {
			if first[i] != again[i] {
				t.Fatalf("round %d row %d differs:\n%+v\n%+v", round, i, first[i], again[i])
			}
		}
	}
}

// การคำนวณของบริษัทหนึ่งต้องไม่แตะข้อมูลของอีกบริษัท
func TestRecalculateDoesNotTouchOtherCompany(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	insertDoc(t, db, "BC01", "ITEM01", docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 10, sumAmount: 1000, when: at(1)})
	insertDoc(t, db, "BC02", "ITEM01", docLine{docNo: "PU900", line: 1, transFlag: 12, qty: 7, sumAmount: 700, when: at(1)})

	for _, business := range []string{"BC01", "BC02"} {
		if _, err := Recalculate(ctx, db, Scope{BusinessCode: business, ItemCode: "ITEM01", From: at(1)},
			testTransFlags, DefaultOptions()); err != nil {
			t.Fatalf("recalculate %s: %v", business, err)
		}
	}

	// คิดใหม่ของ BC01 อีกรอบ ต้องไม่ลบแถวของ BC02
	if _, err := Recalculate(ctx, db, Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: at(1)},
		testTransFlags, DefaultOptions()); err != nil {
		t.Fatalf("second recalculate: %v", err)
	}

	other := ledgerSnapshot(t, db, "BC02", "ITEM01")
	if len(other) != 1 {
		t.Fatalf("BC02 ledger rows = %d, want 1 (must survive BC01 recalculation)", len(other))
	}
	if other[0].BalanceQty != 7 {
		t.Errorf("BC02 balance = %v, want 7", other[0].BalanceQty)
	}
}

// เอกสารที่ถูกยกเลิกต้องไม่ถูกนำมาคิดต้นทุน
func TestCancelledDocumentIsExcluded(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	insertDoc(t, db, "BC01", "ITEM01",
		docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 10, sumAmount: 1000, when: at(1)},
		docLine{docNo: "PU002", line: 1, transFlag: 12, qty: 99, sumAmount: 9900, when: at(2), cancelled: true},
	)

	if _, err := Recalculate(ctx, db, Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: at(1)},
		testTransFlags, DefaultOptions()); err != nil {
		t.Fatalf("recalculate: %v", err)
	}

	ledger := ledgerSnapshot(t, db, "BC01", "ITEM01")
	if len(ledger) != 1 {
		t.Fatalf("ledger rows = %d, want 1 (cancelled document must be skipped)", len(ledger))
	}
	if ledger[0].BalanceQty != 10 {
		t.Errorf("balance = %v, want 10", ledger[0].BalanceQty)
	}
}

// คิดใหม่เฉพาะช่วงท้ายต้องได้ผลเท่ากับคิดใหม่ทั้งหมด และต้องไม่ลบประวัติงวดก่อน
func TestPartialRecalculationMatchesFullRecalculation(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	january := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)
	february := time.Date(2026, 2, 10, 9, 0, 0, 0, time.UTC)
	march := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)

	insertDoc(t, db, "BC01", "ITEM01",
		docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 10, sumAmount: 1000, when: january},
		docLine{docNo: "PU002", line: 1, transFlag: 12, qty: 10, sumAmount: 1400, when: february},
		docLine{docNo: "SA001", line: 1, transFlag: 44, qty: 5, when: march},
	)

	full := Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: january}
	if _, err := Recalculate(ctx, db, full, testTransFlags, DefaultOptions()); err != nil {
		t.Fatalf("full recalculate: %v", err)
	}
	expected := ledgerSnapshot(t, db, "BC01", "ITEM01")

	// คิดใหม่เฉพาะเดือนมีนาคม โดยอาศัยยอดยกมาปลายเดือนกุมภาพันธ์
	partial := Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: march}
	if _, err := Recalculate(ctx, db, partial, testTransFlags, DefaultOptions()); err != nil {
		t.Fatalf("partial recalculate: %v", err)
	}
	got := ledgerSnapshot(t, db, "BC01", "ITEM01")

	if len(got) != len(expected) {
		t.Fatalf("rows after partial = %d, want %d (history must survive)", len(got), len(expected))
	}
	for i := range expected {
		if expected[i] != got[i] {
			t.Fatalf("row %d differs after partial recalculation:\n%+v\n%+v", i, expected[i], got[i])
		}
	}
}

// งานที่แตะสินค้าเดียวกันหลายครั้งต้องยุบเหลือชิ้นเดียว และเลื่อนไปวันที่เก่าที่สุด
func TestMarkDirtyCoalescesToOldestDate(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	if err := MarkDirty(ctx, db, "BC01", "ITEM01", at(10), "doc"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}
	if err := MarkDirty(ctx, db, "BC01", "ITEM01", at(3), "doc"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}
	if err := MarkDirty(ctx, db, "BC01", "ITEM01", at(20), "doc"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}

	var count int
	var from time.Time
	if err := db.QueryRow(`SELECT COUNT(*), MIN(fromdate) FROM stock_dirty`).Scan(&count, &from); err != nil {
		t.Fatalf("read dirty: %v", err)
	}
	if count != 1 {
		t.Fatalf("dirty rows = %d, want 1 (three documents must coalesce)", count)
	}
	if !from.UTC().Equal(at(3)) {
		t.Errorf("fromdate = %v, want %v (oldest affected date)", from.UTC(), at(3))
	}
}

// งานที่ถูกจองแล้วต้องไม่ถูกหยิบซ้ำโดยตัวอื่น
func TestClaimDirtyLeasesExclusively(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	for _, item := range []string{"ITEM01", "ITEM02", "ITEM03"} {
		if err := MarkDirty(ctx, db, "BC01", item, at(1), "doc"); err != nil {
			t.Fatalf("mark dirty: %v", err)
		}
	}

	first, err := ClaimDirty(ctx, db, "worker-a", 2)
	if err != nil {
		t.Fatalf("claim a: %v", err)
	}
	second, err := ClaimDirty(ctx, db, "worker-b", 5)
	if err != nil {
		t.Fatalf("claim b: %v", err)
	}

	if len(first) != 2 {
		t.Fatalf("worker-a claimed %d, want 2", len(first))
	}
	if len(second) != 1 {
		t.Fatalf("worker-b claimed %d, want 1 (leased work must not be handed out twice)", len(second))
	}

	claimed := map[string]bool{}
	for _, item := range append(first, second...) {
		if claimed[item.ItemCode] {
			t.Fatalf("item %s was claimed twice", item.ItemCode)
		}
		claimed[item.ItemCode] = true
	}
}

// งานที่ล้มเหลวต้องกลับเข้าคิว ไม่หายเงียบ
func TestReleaseReturnsWorkToQueue(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	if err := MarkDirty(ctx, db, "BC01", "ITEM01", at(1), "doc"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}
	claimed, err := ClaimDirty(ctx, db, "worker-a", 1)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim: %v (got %d)", err, len(claimed))
	}
	if err := ReleaseDirty(ctx, db, claimed[0], fmt.Errorf("database unavailable")); err != nil {
		t.Fatalf("release: %v", err)
	}

	again, err := ClaimDirty(ctx, db, "worker-b", 1)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if len(again) != 1 {
		t.Fatalf("released work was not reclaimable (got %d items)", len(again))
	}
	if again[0].Attempts != 2 {
		t.Errorf("attempts = %d, want 2", again[0].Attempts)
	}
}

func TestMoveToDeadLetterRemovesFromQueue(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	if err := MarkDirty(ctx, db, "BC01", "ITEM01", at(1), "doc"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}
	claimed, _ := ClaimDirty(ctx, db, "worker-a", 1)
	if err := MoveToDeadLetter(ctx, db, claimed[0], fmt.Errorf("broken row")); err != nil {
		t.Fatalf("dead letter: %v", err)
	}

	var pending, dead int
	db.QueryRow(`SELECT COUNT(*) FROM stock_dirty`).Scan(&pending)
	db.QueryRow(`SELECT COUNT(*) FROM deadletterqueue`).Scan(&dead)
	if pending != 0 {
		t.Errorf("stock_dirty rows = %d, want 0", pending)
	}
	if dead != 1 {
		t.Errorf("deadletterqueue rows = %d, want 1", dead)
	}
}

// สองตัวที่คิดสินค้าเดียวกันพร้อมกัน ต้องมีตัวเดียวที่เขียนได้ อีกตัวต้องได้รับแจ้งว่าไม่ว่าง
// ไม่ใช่เขียนทับกันหรือเงียบหายไป
func TestConcurrentRecalculationIsSerialised(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	insertDoc(t, db, "BC01", "ITEM01",
		docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 10, sumAmount: 1000, when: at(1)})

	scope := Scope{BusinessCode: "BC01", ItemCode: "ITEM01", From: at(1)}
	result := Calculate(nil, []Movement{{
		WhCode: "WH01", DocDateTime: at(1), DocNo: "PU001", LineNumber: 1,
		TransFlag: 12, Qty: 10, UnitStand: 1, UnitDivide: 1, SumAmount: 1000,
	}}, DefaultOptions())

	// ตัวแรกจับล็อกแล้วค้างไว้ด้วยทรานแซกชันที่ยังไม่ commit
	holder, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var held bool
	if err := holder.QueryRowContext(ctx,
		`SELECT pg_try_advisory_xact_lock(hashtext($1))`, "stockengine:BC01:ITEM01").Scan(&held); err != nil {
		t.Fatal(err)
	}
	if !held {
		t.Fatal("could not acquire lock for the test holder")
	}

	err = Persist(ctx, db, scope, result)
	if err == nil {
		holder.Rollback()
		t.Fatal("second writer succeeded while the item was locked")
	}
	if err != ErrLockBusy {
		holder.Rollback()
		t.Fatalf("error = %v, want ErrLockBusy", err)
	}

	// ปล่อยล็อกแล้วต้องเขียนได้ตามปกติ
	if err := holder.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := Persist(ctx, db, scope, result); err != nil {
		t.Fatalf("persist after lock released: %v", err)
	}
	if rows := ledgerSnapshot(t, db, "BC01", "ITEM01"); len(rows) != 1 {
		t.Fatalf("ledger rows = %d, want 1", len(rows))
	}
}

// เขียนพร้อมกันหลายสินค้าต้องไม่ชนกัน
func TestParallelItemsDoNotBlockEachOther(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	items := []string{"ITEM01", "ITEM02", "ITEM03", "ITEM04"}
	for _, item := range items {
		insertDoc(t, db, "BC01", item,
			docLine{docNo: "PU-" + item, line: 1, transFlag: 12, qty: 5, sumAmount: 500, when: at(1)})
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(items))
	for _, item := range items {
		wg.Add(1)
		go func(item string) {
			defer wg.Done()
			if _, err := Recalculate(ctx, db, Scope{BusinessCode: "BC01", ItemCode: item, From: at(1)},
				testTransFlags, DefaultOptions()); err != nil {
				errs <- fmt.Errorf("%s: %w", item, err)
			}
		}(item)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("parallel recalculation failed: %v", err)
	}

	var total int
	db.QueryRow(`SELECT COUNT(*) FROM stock_ledger WHERE businesscode='BC01'`).Scan(&total)
	if total != len(items) {
		t.Errorf("ledger rows = %d, want %d", total, len(items))
	}
}

// ฐานข้อมูลต้องปฏิเสธข้อมูลที่เป็นไปไม่ได้ทางบัญชี แม้โค้ดจะพลาดส่งมา
func TestDatabaseRejectsImpossibleValues(t *testing.T) {
	db := engineTestDB(t)

	const insertPrefix = `INSERT INTO stock_ledger (businesscode,itemcode,whcode,docdatetime,behindindex,docno,linenumber,
		periodkey,locationcode,barcode,unitcode,docref,transflag,direction,qty,unitcost,amount,balanceqty,balanceamount,avgcost)
		VALUES (`

	// แถวนี้ถูกต้องทุกอย่าง ใช้ยืนยันว่าเทสต์ล้มเหลวเพราะข้อจำกัดที่ตั้งใจทดสอบจริง
	// ไม่ใช่เพราะคำสั่ง INSERT เขียนผิด
	const validRow = `'BC01','I1','WH01','2026-03-01 09:00:00+00',0,'OK1',1,'2026-03','','','','',12,1,1,1,1,1,1,1`
	if _, err := db.Exec(insertPrefix + validRow + `)`); err != nil {
		t.Fatalf("baseline row must be accepted, got: %v", err)
	}

	cases := []struct {
		name   string
		values string
	}{
		{"direction นอกเหนือจากรับเข้า/จ่ายออก", `'BC01','I1','WH01',now(),0,'D1',1,'2026-03','','','','',12,0,1,1,1,1,1,1`},
		{"จำนวนติดลบ", `'BC01','I1','WH01',now(),0,'D2',1,'2026-03','','','','',12,1,-5,1,1,1,1,1`},
		{"ต้นทุนติดลบ", `'BC01','I1','WH01',now(),0,'D3',1,'2026-03','','','','',12,1,1,-1,1,1,1,1`},
		{"ต้นทุนถัวเฉลี่ยติดลบ", `'BC01','I1','WH01',now(),0,'D4',1,'2026-03','','','','',12,1,1,1,1,1,1,-1`},
		{"รหัสงวดผิดรูปแบบ", `'BC01','I1','WH01',now(),0,'D5',1,'2026-3 ','','','','',12,1,1,1,1,1,1,1`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := db.Exec(insertPrefix + tc.values + `)`); err == nil {
				t.Error("database accepted a value that should be impossible")
			}
		})
	}
}

// คีย์หลักต้องกันแถวซ้ำได้จริง
func TestLedgerPrimaryKeyRejectsDuplicates(t *testing.T) {
	db := engineTestDB(t)
	const insert = `INSERT INTO stock_ledger (businesscode,itemcode,whcode,docdatetime,behindindex,docno,linenumber,
		periodkey,locationcode,barcode,unitcode,docref,transflag,direction,qty,unitcost,amount,balanceqty,balanceamount,avgcost)
		VALUES ('BC01','I1','WH01','2026-03-01 09:00:00+00',0,'D1',1,'2026-03','','','','',12,1,1,1,1,1,1,1)`

	if _, err := db.Exec(insert); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if _, err := db.Exec(insert); err == nil {
		t.Error("duplicate ledger row was accepted")
	}
}

// สัญญาณปลุกต้องถึง worker จริง ไม่ใช่รอรอบตรวจเอง
func TestWorkerWakesOnNotification(t *testing.T) {
	dsn := os.Getenv("BC_STOCK_TEST_POSTGRES_DSN")
	db := engineTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	insertDoc(t, db, "BC01", "ITEM01",
		docLine{docNo: "PU001", line: 1, transFlag: 12, qty: 6, sumAmount: 600, when: at(1)})

	worker := NewWorker(db, dsn, testTransFlags)
	worker.PollInterval = time.Hour // ตัดรอบตรวจเองออก เพื่อพิสูจน์ว่าถูกปลุกด้วยสัญญาณจริง

	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	time.Sleep(500 * time.Millisecond) // ให้ตัวรับสัญญาณพร้อมก่อน
	if err := MarkDirty(ctx, db, "BC01", "ITEM01", at(1), "doc"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if rows := ledgerSnapshot(t, db, "BC01", "ITEM01"); len(rows) == 1 {
			cancel()
			<-done
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	cancel()
	<-done
	t.Fatal("worker did not process the item after notification")
}

func TestQueueStatusReportsPendingWork(t *testing.T) {
	db := engineTestDB(t)
	ctx := context.Background()

	for _, item := range []string{"ITEM01", "ITEM02"} {
		if err := MarkDirty(ctx, db, "BC01", item, at(1), "doc"); err != nil {
			t.Fatalf("mark dirty: %v", err)
		}
	}
	if _, err := ClaimDirty(ctx, db, "worker-a", 1); err != nil {
		t.Fatalf("claim: %v", err)
	}

	status, err := LoadQueueStatus(ctx, db)
	if err != nil {
		t.Fatalf("queue status: %v", err)
	}
	if status.Pending != 1 {
		t.Errorf("pending = %d, want 1", status.Pending)
	}
	if status.Processing != 1 {
		t.Errorf("processing = %d, want 1", status.Processing)
	}
}

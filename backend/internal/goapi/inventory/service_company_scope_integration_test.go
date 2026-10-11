//go:build integration

package inventory

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"smlcloudplatform/internal/doctype"
)

// TestInventoryReportsNeverMixCompanies - รายงานเคลื่อนไหวสต็อก / สต็อกการ์ด / มูลค่าคงเหลือ บน PostgreSQL จริง:
// บริษัท 02 ต้องไม่เห็นสต็อกบริษัท 01 (เดิมมี OR s.businesscode = '01') และกลับกัน
// ใช้ schema แยก inv_scope_it ที่มีเฉพาะคอลัมน์ที่รายงานอ่าน เพื่อไม่ชนกับตารางจริงของแพ็กเกจอื่นในฐานทดสอบเดียวกัน
func TestInventoryReportsNeverMixCompanies(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN")
	}
	ctx := context.Background()
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	dropSchema := func() {
		if _, err := admin.Exec(`DROP SCHEMA IF EXISTS inv_scope_it CASCADE`); err != nil {
			t.Fatal(err)
		}
	}
	dropSchema()
	t.Cleanup(dropSchema)
	if _, err := admin.Exec(`
CREATE SCHEMA inv_scope_it;
CREATE TABLE inv_scope_it.stock_ledger (
	businesscode text NOT NULL, itemcode text NOT NULL, whcode text NOT NULL, docdatetime timestamptz NOT NULL,
	behindindex integer NOT NULL, docno text NOT NULL, linenumber integer NOT NULL, unitcode text NOT NULL DEFAULT '',
	docref text NOT NULL DEFAULT '', transflag integer NOT NULL, direction smallint NOT NULL,
	docqty numeric(18,8) NOT NULL, unitcost numeric(18,8) NOT NULL, amount numeric(18,8) NOT NULL,
	balanceqty numeric(18,8) NOT NULL, balanceamount numeric(18,8) NOT NULL, avgcost numeric(18,8) NOT NULL);
CREATE TABLE inv_scope_it.docdetail (businesscode text NOT NULL, docno text NOT NULL, itemcode text NOT NULL, description text);
CREATE TABLE inv_scope_it.stock_period_balance (
	businesscode text NOT NULL, itemcode text NOT NULL, whcode text NOT NULL, periodkey char(7) NOT NULL,
	closeqty numeric(18,8) NOT NULL, closeavgcost numeric(18,8) NOT NULL, closeamount numeric(18,8) NOT NULL,
	PRIMARY KEY (businesscode, itemcode, whcode, periodkey));
-- ตารางเดิมที่กรองแค่ holding: รายงานของบริษัทต้องไม่ถอยมาอ่าน (เดิมอ่านเมื่อบริษัทไม่มีแถวในสมุดสต็อก)
CREATE TABLE inv_scope_it.inventorycosttransactions (
	id serial, holdingcode text, itemcode text, transactiondate timestamptz, refdocno text, transactiontype text, transflag integer,
	qty numeric, unitcost numeric, totalcost numeric, balanceqty numeric, balancetotalvalue numeric, balanceavgcost numeric);
CREATE TABLE inv_scope_it.inventorystockbalances (
	holdingcode text, itemcode text, whcode text, currentqty numeric, currentavgcost numeric, currenttotalvalue numeric);
CREATE TABLE inv_scope_it.productcostingconfig (holdingcode text, itemcode text, costingmethod text);
INSERT INTO inv_scope_it.inventorycosttransactions(holdingcode, itemcode, transactiondate, refdocno, transactiontype, transflag,
	qty, unitcost, totalcost, balanceqty, balancetotalvalue, balanceavgcost)
	VALUES ('rungrueng', 'WTR-600', '2026-09-15 03:00:00+00', 'RR6909-0001', 'purchase', 12, 500, 6.5, 3250, 500, 3250, 6.5);
INSERT INTO inv_scope_it.inventorystockbalances(holdingcode, itemcode, whcode, currentqty, currentavgcost, currenttotalvalue)
	VALUES ('rungrueng', 'WTR-600', 'WH-01', 500, 6.5, 3250);`); err != nil {
		t.Fatal(err)
	}

	scoped, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := scoped.Query()
	query.Set("search_path", "inv_scope_it")
	scoped.RawQuery = query.Encode()
	db, err := sql.Open("postgres", scoped.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	at := time.Date(2026, 9, 15, 3, 0, 0, 0, time.UTC)
	companies := map[string]struct{ docno, name string }{
		"01": {docno: "IV6909-0101", name: "น้ำดื่มตราช้าง 600 มล. แพ็ก 12 ขวด"},
		"02": {docno: "IV6909-0201", name: "น้ำดื่มตราช้าง 600 มล. ลังละ 24 ขวด"},
	}
	for company, c := range companies {
		if _, err := db.Exec(`INSERT INTO stock_ledger(businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber,
			unitcode, transflag, direction, docqty, unitcost, amount, balanceqty, balanceamount, avgcost)
			VALUES($1, 'WTR-600', 'WH-01', $2, 1, $3, 1, 'ขวด', $4, -1, 12, 6.5, 78, 88, 572, 6.5)`,
			company, at, c.docno, doctype.SaleInvoice); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO docdetail(businesscode, docno, itemcode, description) VALUES($1, $2, 'WTR-600', $3)`,
			company, c.docno, c.name); err != nil {
			t.Fatal(err)
		}
		// ยอดปิดทุกงวดที่มีความเคลื่อนไหว: มูลค่าคงเหลือต้องใช้งวดล่าสุด (2026-09) ไม่บวกงวด 2026-08 ซ้ำ
		if _, err := db.Exec(`INSERT INTO stock_period_balance(businesscode, itemcode, whcode, periodkey, closeqty, closeavgcost, closeamount)
			VALUES($1, 'WTR-600', 'WH-01', '2026-08', 100, 6.5, 650), ($1, 'WTR-600', 'WH-01', '2026-09', 88, 6.5, 572)`, company); err != nil {
			t.Fatal(err)
		}
	}

	svc := NewInventoryCostingService(db)
	from, to := at.AddDate(0, 0, -1), at.AddDate(0, 0, 1)
	for company, want := range companies {
		movement, err := svc.GetStockMovementReport(ctx, company, "", from, to)
		if err != nil {
			t.Fatal(err)
		}
		if len(movement) != 1 || movement[0].DocNo != want.docno || movement[0].ItemName != want.name {
			t.Errorf("company %s stock movement = %+v, want only %s", company, movement, want.docno)
		}

		card, err := svc.GetStockCard(ctx, "rungrueng", company, "WTR-600", from, to)
		if err != nil {
			t.Fatal(err)
		}
		if len(card.Entries) != 1 || card.Entries[0].DocNo != want.docno || card.ItemName != want.name {
			t.Errorf("company %s stock card = %+v, want only %s", company, card, want.docno)
		}

		valuation, err := svc.GetInventoryValuation(ctx, "rungrueng", company)
		if err != nil {
			t.Fatal(err)
		}
		if len(valuation.Items) != 1 || valuation.Items[0].ItemName != want.name || valuation.Items[0].Qty != 88 || valuation.Items[0].TotalValue != 572 {
			t.Errorf("company %s valuation = %+v, want one item %q qty 88 value 572 (latest period only)", company, valuation.Items, want.name)
		}
	}

	// บริษัทที่ไม่มีสต็อก ต้องได้รายงานว่าง ไม่ใช่สต็อกบริษัท 01
	movement, err := svc.GetStockMovementReport(ctx, "03", "", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(movement) != 0 {
		t.Errorf("company 03 sees %+v", movement)
	}
	card, err := svc.GetStockCard(ctx, "rungrueng", "03", "WTR-600", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(card.Entries) != 0 {
		t.Errorf("company 03 stock card = %+v, want empty (no holding-wide fallback)", card.Entries)
	}
	valuation, err := svc.GetInventoryValuation(ctx, "rungrueng", "03")
	if err != nil {
		t.Fatal(err)
	}
	if len(valuation.Items) != 0 || valuation.TotalValue != 0 {
		t.Errorf("company 03 valuation = %+v, want empty (no holding-wide fallback)", valuation)
	}
}

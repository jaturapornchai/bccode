package stockengine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/lib/pq"
)

// คีย์เรียงลำดับการคิดต้นทุน ต้องตรงกับ SortMovements และกับดัชนี idx_docdetail_calcorder
// ทั้งสี่คอลัมน์รวมกันไม่ซ้ำ เพราะ (businesscode, docno, linenumber) เป็นคีย์ของบรรทัดเอกสาร
const calcOrderBy = "ORDER BY d.docdatetime, d.behindindex, d.docno, d.linenumber"

// movementColumns คือคอลัมน์ที่ต้องใช้คำนวณ เลือกเท่าที่ใช้จริงเพื่อไม่ดึงข้อมูลเกินจำเป็น
const movementColumns = `d.whcode, d.locationcode, d.barcode, d.unitcode, d.docref,
	d.docdatetime, d.behindindex, d.docno, d.linenumber, d.transflag, d.calcflag,
	d.totalqty, d.unitstand, d.unitdivide, d.price, d.priceexcludevat, d.sumamount`

var ledgerColumns = []string{
	"businesscode", "itemcode", "whcode", "docdatetime", "behindindex", "docno", "linenumber",
	"periodkey", "locationcode", "barcode", "unitcode", "docref", "transflag", "direction",
	"docqty", "unitstand", "unitdivide", "price", "docvalue",
	"qty", "unitcost", "amount", "balanceqty", "balanceamount", "avgcost",
}

// Scope ระบุขอบเขตของงานคำนวณหนึ่งชิ้น
type Scope struct {
	BusinessCode string
	ItemCode     string
	// From คือจุดเริ่มคิดใหม่ รายการก่อนหน้านี้ถือว่าถูกต้องแล้วและไม่ถูกแตะ
	From time.Time
}

// periodStart คืนวันแรกของงวดที่วันที่นั้นอยู่ ใช้เป็นขอบเขตการคิดใหม่จริง
// เพื่อให้ยอดยกมาที่อ่านมาเป็นยอดสิ้นงวดก่อนหน้าพอดี
func periodStart(from time.Time) time.Time {
	return time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
}

// LoadOpening อ่านยอดคงเหลือปลายงวดล่าสุดที่อยู่ก่อนงวดของ scope.From ใช้เป็นจุดตั้งต้นการคำนวณ
//
// อ่าน "งวดล่าสุดที่มีอยู่จริง" ไม่ใช่ "งวดก่อนหน้าหนึ่งเดือน" เพราะยอดปลายงวดถูกเขียนเฉพาะงวดที่มีรายการ
// สินค้าที่ไม่มีการเคลื่อนไหวมาสามเดือนจึงยังได้ยอดยกมาถูกต้อง ไม่ใช่ศูนย์
// คลังที่ยังไม่เคยมียอดจะไม่อยู่ในผลลัพธ์ ผู้เรียกถือว่าเริ่มจากศูนย์
func LoadOpening(ctx context.Context, db *sql.DB, scope Scope) (map[string]Balance, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT ON (whcode) whcode, closeqty, closeamount, closeavgcost
		FROM stock_period_balance
		WHERE businesscode = $1 AND itemcode = $2 AND periodkey < $3
		ORDER BY whcode, periodkey DESC`,
		scope.BusinessCode, scope.ItemCode, PeriodKeyOf(scope.From))
	if err != nil {
		return nil, fmt.Errorf("load opening balance: %w", err)
	}
	defer rows.Close()

	opening := make(map[string]Balance)
	for rows.Next() {
		var wh string
		var bal Balance
		if err := rows.Scan(&wh, &bal.Qty, &bal.Amount, &bal.AvgCost); err != nil {
			return nil, fmt.Errorf("scan opening balance: %w", err)
		}
		opening[wh] = bal
	}
	return opening, rows.Err()
}

// LoadMovements อ่านรายการเคลื่อนไหวตั้งแต่ต้นงวดของ scope.From เป็นต้นไป
// อ่านทีละแถวจาก cursor ไม่โหลดทั้งประวัติสินค้าเข้าหน่วยความจำ
// กรองเอกสารที่ถูกยกเลิกและถูกลบออกโดยดูจากหัวเอกสาร เพราะ docdetail.iscancel ไม่ถูกเขียนค่าจริง
// การลบเอกสารเป็นการทำเครื่องหมาย isdelete ไม่ได้ลบบรรทัดทิ้ง ถ้าไม่กรองของที่ลบแล้วจะยังกินสต็อกตลอดไป
func LoadMovements(ctx context.Context, db *sql.DB, scope Scope, transFlags []int) ([]Movement, error) {
	flags := make([]string, len(transFlags))
	for i, flag := range transFlags {
		flags[i] = fmt.Sprintf("%d", flag)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM docdetail d
		LEFT JOIN doc h ON h.businesscode = d.businesscode AND h.docno = d.docno
		WHERE d.businesscode = $1 AND d.itemcode = $2 AND d.docdatetime >= $3
		  AND d.transflag IN (%s)
		  AND COALESCE(h.iscancel, FALSE) = FALSE
		  AND COALESCE(h.isdelete, FALSE) = FALSE
		%s`, movementColumns, strings.Join(flags, ","), calcOrderBy)

	rows, err := db.QueryContext(ctx, query, scope.BusinessCode, scope.ItemCode, periodStart(scope.From))
	if err != nil {
		return nil, fmt.Errorf("query movements: %w", err)
	}
	defer rows.Close()

	movements := make([]Movement, 0, 256)
	for rows.Next() {
		var m Movement
		var whCode, locationCode, barcode, unitCode, docRef, docNo sql.NullString
		var qty, unitStand, unitDivide, price, priceExcludeVat, sumAmount sql.NullFloat64
		var behindIndex, lineNumber, transFlag, calcFlag sql.NullInt64

		if err := rows.Scan(&whCode, &locationCode, &barcode, &unitCode, &docRef,
			&m.DocDateTime, &behindIndex, &docNo, &lineNumber, &transFlag, &calcFlag,
			&qty, &unitStand, &unitDivide, &price, &priceExcludeVat, &sumAmount); err != nil {
			return nil, fmt.Errorf("scan movement: %w", err)
		}

		m.WhCode = whCode.String
		m.LocationCode = locationCode.String
		m.Barcode = barcode.String
		m.UnitCode = unitCode.String
		m.DocRef = docRef.String
		m.DocNo = docNo.String
		m.BehindIndex = int(behindIndex.Int64)
		m.LineNumber = int(lineNumber.Int64)
		m.TransFlag = int(transFlag.Int64)
		m.CalcFlag = int(calcFlag.Int64)
		m.Qty = qty.Float64
		m.UnitStand = unitStand.Float64
		m.UnitDivide = unitDivide.Float64
		m.Price = price.Float64
		m.PriceExcludeVat = priceExcludeVat.Float64
		m.SumAmount = sumAmount.Float64

		movements = append(movements, m)
	}
	return movements, rows.Err()
}

// Persist เขียนผลการคำนวณลงฐานข้อมูลในทรานแซกชันเดียว
//
// ทุกอย่างอยู่ในทรานแซกชันเดียวกันโดยตั้งใจ ผู้ใช้จึงไม่มีทางอ่านเจอสถานะระหว่างทาง
// ที่ข้อมูลถูกลบไปแล้วแต่ยังเขียนกลับไม่เสร็จ และถ้าโปรเซสตายกลางคันข้อมูลเดิมยังอยู่ครบ
//
// การจอง advisory lock อยู่ในทรานแซกชันเดียวกัน จึงถูกปล่อยอัตโนมัติเมื่อจบงาน
// ไม่ว่าจะสำเร็จหรือล้มเหลว ต่างจากล็อกที่เก็บเป็นแถวในตารางซึ่งค้างได้ถ้าโปรเซสตาย
//
// dirtyBefore คือเวลาที่อ่านรายการเคลื่อนไหวมา งานในคิวที่ถูกทำเครื่องหมายหลังเวลานั้น
// ต้องไม่ถูกลบทิ้ง เพราะเป็นเอกสารที่เข้ามาหลังจากอ่านข้อมูลไปแล้วและยังไม่ได้ถูกคิด
func Persist(ctx context.Context, db *sql.DB, scope Scope, result Result, dirtyBefore time.Time) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	locked, err := acquireAdvisoryLock(ctx, tx, scope)
	if err != nil {
		return err
	}
	if !locked {
		return ErrLockBusy
	}

	from := periodStart(scope.From)

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM stock_ledger
		WHERE businesscode = $1 AND itemcode = $2 AND docdatetime >= $3`,
		scope.BusinessCode, scope.ItemCode, from); err != nil {
		return fmt.Errorf("clear ledger from %s: %w", from.Format(time.RFC3339), err)
	}

	if err := copyLedgerRows(ctx, tx, scope, result.Rows); err != nil {
		return err
	}

	// ยอดปลายงวดเก่าที่อยู่ในช่วงคิดใหม่ต้องถูกล้างก่อน ไม่ใช่แค่ทับ
	// เอกสารที่ถูกลบจนงวดนั้นไม่เหลือรายการ จะไม่มีตัวไปทับยอดเดิม แล้วยอดผิดจะค้างเป็นยอดยกมาของงวดถัดไป
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM stock_period_balance
		WHERE businesscode = $1 AND itemcode = $2 AND periodkey >= $3`,
		scope.BusinessCode, scope.ItemCode, PeriodKeyOf(from)); err != nil {
		return fmt.Errorf("clear period balance from %s: %w", PeriodKeyOf(from), err)
	}

	if err := upsertPeriodBalances(ctx, tx, scope, result.Closing); err != nil {
		return err
	}

	// ลบเฉพาะงานที่เกิดก่อนเวลาที่อ่านข้อมูล เอกสารที่เข้ามาระหว่างคำนวณต้องยังค้างคิวไว้คิดรอบหน้า
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM stock_dirty
		WHERE businesscode = $1 AND itemcode = $2 AND markedat <= $3`,
		scope.BusinessCode, scope.ItemCode, dirtyBefore); err != nil {
		return fmt.Errorf("clear dirty flag: %w", err)
	}

	return tx.Commit()
}

// ErrLockBusy บอกว่ามีตัวอื่นกำลังคำนวณสินค้าตัวนี้อยู่
// ผู้เรียกต้องไม่กลืนข้อผิดพลาดนี้เงียบ ๆ งานยังค้างอยู่ใน stock_dirty และต้องถูกหยิบไปทำใหม่
var ErrLockBusy = fmt.Errorf("stockengine: another worker is calculating this item")

// acquireAdvisoryLock จองสิทธิ์คำนวณสินค้าตัวนี้ภายในทรานแซกชันปัจจุบัน
func acquireAdvisoryLock(ctx context.Context, tx *sql.Tx, scope Scope) (bool, error) {
	key := "stockengine:" + scope.BusinessCode + ":" + scope.ItemCode
	var locked bool
	if err := tx.QueryRowContext(ctx, `SELECT pg_try_advisory_xact_lock(hashtext($1))`, key).Scan(&locked); err != nil {
		return false, fmt.Errorf("acquire advisory lock: %w", err)
	}
	return locked, nil
}

func copyLedgerRows(ctx context.Context, tx *sql.Tx, scope Scope, rows []LedgerRow) error {
	if len(rows) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("stock_ledger", ledgerColumns...))
	if err != nil {
		return fmt.Errorf("prepare ledger copy: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.ExecContext(ctx,
			scope.BusinessCode, scope.ItemCode, row.WhCode, row.DocDateTime, row.BehindIndex, row.DocNo, row.LineNumber,
			PeriodKeyOf(row.DocDateTime),
			row.LocationCode, row.Barcode, row.UnitCode, row.DocRef, row.TransFlag, row.Direction,
			row.DocQty, row.UnitStand, row.UnitDivide, row.Price, row.DocValue,
			row.Qty, row.UnitCost, row.Amount, row.BalanceQty, row.BalanceAmount, row.AvgCost); err != nil {
			return fmt.Errorf("copy ledger row %s/%d: %w", row.DocNo, row.LineNumber, err)
		}
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("flush ledger copy: %w", err)
	}
	return nil
}

func upsertPeriodBalances(ctx context.Context, tx *sql.Tx, scope Scope, closing map[string]Balance) error {
	for key, bal := range closing {
		whCode, periodKey, ok := splitClosingKey(key)
		if !ok {
			return fmt.Errorf("malformed closing key %q", key)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO stock_period_balance
				(businesscode, itemcode, whcode, periodkey, closeqty, closeamount, closeavgcost, hastrans, updatedat)
			VALUES ($1,$2,$3,$4,$5,$6,$7,TRUE,now())
			ON CONFLICT (businesscode, itemcode, whcode, periodkey)
			DO UPDATE SET closeqty = EXCLUDED.closeqty,
			              closeamount = EXCLUDED.closeamount,
			              closeavgcost = EXCLUDED.closeavgcost,
			              hastrans = TRUE,
			              updatedat = now()`,
			scope.BusinessCode, scope.ItemCode, whCode, periodKey,
			bal.Qty, bal.Amount, bal.AvgCost); err != nil {
			return fmt.Errorf("upsert period balance %s/%s: %w", whCode, periodKey, err)
		}
	}
	return nil
}

func splitClosingKey(key string) (whCode, periodKey string, ok bool) {
	idx := strings.LastIndex(key, "|")
	if idx < 0 {
		return "", "", false
	}
	return key[:idx], key[idx+1:], true
}

// ScopesForItem คืนขอบเขตงานของสินค้าหนึ่งรหัสในทุกบริษัทที่มีเอกสารเคลื่อนไหวจริง
//
// รหัสสินค้าเดียวกันอยู่ได้หลายบริษัทในฐานข้อมูลเดียวกัน การคำนวณจึงต้องแยกทีละบริษัท
func ScopesForItem(ctx context.Context, db *sql.DB, itemCode string, transFlags []int) ([]Scope, error) {
	flags := make([]string, len(transFlags))
	for i, flag := range transFlags {
		flags[i] = fmt.Sprintf("%d", flag)
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT businesscode, MIN(docdatetime)
		FROM docdetail
		WHERE itemcode = $1 AND transflag IN (%s)
		GROUP BY businesscode`, strings.Join(flags, ",")), itemCode)
	if err != nil {
		return nil, fmt.Errorf("list companies for item %s: %w", itemCode, err)
	}
	defer rows.Close()

	scopes := []Scope{}
	for rows.Next() {
		scope := Scope{ItemCode: itemCode}
		if err := rows.Scan(&scope.BusinessCode, &scope.From); err != nil {
			return nil, fmt.Errorf("scan company for item %s: %w", itemCode, err)
		}
		scopes = append(scopes, scope)
	}
	return scopes, rows.Err()
}

// RecalculateSummary สรุปผลการคำนวณใหม่ทั้งฐานข้อมูล
type RecalculateSummary struct {
	Items         int
	Failed        int
	BusinessCodes []string
}

// RecalculateAll คำนวณต้นทุนใหม่ทั้งฐานข้อมูล ไล่ทีละสินค้าและรายงานความคืบหน้าตามจริง
//
// ใช้ตอนสร้างฐานข้อมูลใหม่จากต้นทาง ซึ่งต้องรอให้เสร็จก่อนบอกผู้ใช้ว่าจบแล้ว
// จึงคำนวณตรงไม่ผ่านคิว สินค้าที่ล้มเหลวถูกนับไว้และไม่หยุดตัวที่เหลือ
func RecalculateAll(ctx context.Context, db *sql.DB, transFlags []int, opt Options, progress func(processed, total int)) (RecalculateSummary, error) {
	summary := RecalculateSummary{}

	flags := make([]string, len(transFlags))
	for i, flag := range transFlags {
		flags[i] = fmt.Sprintf("%d", flag)
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT businesscode, itemcode, MIN(docdatetime)
		FROM docdetail
		WHERE transflag IN (%s) AND itemcode <> ''
		GROUP BY businesscode, itemcode
		ORDER BY businesscode, itemcode`, strings.Join(flags, ",")))
	if err != nil {
		return summary, fmt.Errorf("list items to recalculate: %w", err)
	}

	scopes := []Scope{}
	seenBusiness := map[string]bool{}
	for rows.Next() {
		var scope Scope
		if err := rows.Scan(&scope.BusinessCode, &scope.ItemCode, &scope.From); err != nil {
			rows.Close()
			return summary, fmt.Errorf("scan item to recalculate: %w", err)
		}
		scopes = append(scopes, scope)
		if !seenBusiness[scope.BusinessCode] {
			seenBusiness[scope.BusinessCode] = true
			summary.BusinessCodes = append(summary.BusinessCodes, scope.BusinessCode)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return summary, fmt.Errorf("read items to recalculate: %w", err)
	}

	for index, scope := range scopes {
		if ctx.Err() != nil {
			return summary, ctx.Err()
		}
		if _, err := Recalculate(ctx, db, scope, transFlags, opt); err != nil {
			logger.Error("recalculate %s/%s: %v", scope.BusinessCode, scope.ItemCode, err)
			summary.Failed++
		} else {
			summary.Items++
		}
		if progress != nil {
			progress(index+1, len(scopes))
		}
	}

	return summary, nil
}

// Recalculate คำนวณต้นทุนของสินค้าหนึ่งตัวใหม่ตั้งแต่จุดที่กำหนด แล้วบันทึกผล
func Recalculate(ctx context.Context, db *sql.DB, scope Scope, transFlags []int, opt Options) (int, error) {
	// จับเวลาของฐานข้อมูลก่อนอ่านข้อมูล ใช้ตัดสินว่างานในคิวชิ้นไหนถูกคิดไปแล้ว
	// ใช้เวลาของฐานข้อมูลไม่ใช่ของเครื่อง เพราะเวลาที่ทำเครื่องหมายงานก็มาจากฐานข้อมูลเดียวกัน
	var dirtyBefore time.Time
	if err := db.QueryRowContext(ctx, `SELECT now()`).Scan(&dirtyBefore); err != nil {
		return 0, fmt.Errorf("read database time: %w", err)
	}

	opening, err := LoadOpening(ctx, db, scope)
	if err != nil {
		return 0, err
	}
	if len(opening) == 0 {
		// ไม่มียอดยกมาเลย แต่สินค้าอาจมีประวัติมาก่อน เช่นตอนเปิดใช้ระบบครั้งแรก
		// ถ้าเริ่มคิดกลางทางโดยถือว่ายอดต้นงวดเป็นศูนย์ ยอดจะผิดตั้งแต่บรรทัดแรก
		earliest, err := earliestMovement(ctx, db, scope, transFlags)
		if err != nil {
			return 0, err
		}
		if !earliest.IsZero() {
			scope.From = earliest
		}
	}

	movements, err := LoadMovements(ctx, db, scope, transFlags)
	if err != nil {
		return 0, err
	}

	result := Calculate(opening, movements, opt)
	if err := Persist(ctx, db, scope, result, dirtyBefore); err != nil {
		return 0, err
	}
	return len(result.Rows), nil
}

// earliestMovement คืนวันที่ของเอกสารใบแรกสุดของสินค้าที่อยู่ก่อนงวดที่จะคิดใหม่
// คืนเวลาศูนย์ถ้าไม่มีประวัติก่อนหน้า แปลว่าเริ่มคิดจากจุดเดิมได้เลย
func earliestMovement(ctx context.Context, db *sql.DB, scope Scope, transFlags []int) (time.Time, error) {
	flags := make([]string, len(transFlags))
	for i, flag := range transFlags {
		flags[i] = fmt.Sprintf("%d", flag)
	}

	var earliest sql.NullTime
	err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT MIN(docdatetime)
		FROM docdetail
		WHERE businesscode = $1 AND itemcode = $2 AND docdatetime < $3
		  AND transflag IN (%s)`, strings.Join(flags, ",")),
		scope.BusinessCode, scope.ItemCode, periodStart(scope.From)).Scan(&earliest)
	if err != nil {
		return time.Time{}, fmt.Errorf("find earliest movement: %w", err)
	}
	if !earliest.Valid {
		return time.Time{}, nil
	}
	return earliest.Time, nil
}

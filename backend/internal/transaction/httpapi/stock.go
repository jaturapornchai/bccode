package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/doctype"
)

// สมุดสต็อก (stock_ledger) เป็นสายต่อเนื่องของแต่ละ (สินค้า, คลัง): ยอดคงเหลือและต้นทุนเฉลี่ยของแถวหนึ่งขึ้นกับแถวก่อนหน้า
// ตามลำดับ docdatetime, behindindex, docno, linenumber (behindindex = ลำดับที่บันทึกในเวลาเดียวกัน — หน้าจอส่งเวลาเดียวกันทั้งวัน
// ถ้าเรียงด้วยเลขที่เอกสารแทน ใบขายที่เลขน้อยกว่าจะมาก่อนใบซื้อวันเดียวกันและยอดคงเหลือติดลบ) — ดังนั้นเมื่อเอกสารถูกบันทึก แก้ หรือลบ ต้องคำนวณสายใหม่ตั้งแต่แถวแรกที่
// เปลี่ยน ไม่ใช่เพิ่มแถวต่อท้ายเฉยๆ (เอกสารย้อนหลังและเอกสารที่ถูกแก้/ลบจะทำให้แถวหลังจากนั้นผิดทั้งหมด)

const defaultWarehouse = "WH01"

func movesStock(dt doctype.DocType) bool {
	return dt.Stock == doctype.StockIn || dt.Stock == doctype.StockOut
}

type stockKey struct{ item, warehouse string }

// stockSlot คือตำแหน่งเดิมของแถวสต็อก (เวลา + ลำดับในเวลานั้น) ที่ถูกลบเพื่อลงใหม่ตอนแก้เอกสาร
type stockSlot struct {
	at    time.Time
	index int
}

// businessDay คือวันที่ตามปฏิทินของ holding (Asia/Bangkok) ของเวลาที่เก็บเป็น timestamptz — ใช้ทั้งวันที่ใบสำคัญและงวดสต็อก
func businessDay(t time.Time) string {
	return t.In(authmodels.HoldingLocation("")).Format("2006-01-02")
}

// replaceStockMovement ลบแถวสต็อกเดิมของเอกสาร (ถ้ามี) ลงแถวใหม่จาก doc (nil = ลบเอกสาร) แล้วคำนวณสายของทุก (สินค้า, คลัง) ที่เกี่ยวข้องใหม่
func (s *Store) replaceStockMovement(ctx context.Context, tx *sql.Tx, company string, dt doctype.DocType, doc *ErpTransactionDoc, docNo string, docDate time.Time, now time.Time) error {
	touched := map[stockKey]bool{}
	previous := map[stockKey]stockSlot{}
	var fromTime time.Time
	fromPeriod := ""
	track := func(at time.Time, period string) {
		if fromTime.IsZero() || at.Before(fromTime) {
			fromTime = at
		}
		if fromPeriod == "" || period < fromPeriod {
			fromPeriod = period
		}
	}

	removed, err := tx.QueryContext(ctx, `DELETE FROM stock_ledger WHERE businesscode=$1 AND transflag=$2 AND docno=$3 RETURNING itemcode, whcode, docdatetime, behindindex, periodkey`, company, dt.Flag, docNo)
	if err != nil {
		return fmt.Errorf("remove stock rows: %w", err)
	}
	for removed.Next() {
		var key stockKey
		var at time.Time
		var slot int
		var period string
		if err = removed.Scan(&key.item, &key.warehouse, &at, &slot, &period); err != nil {
			removed.Close()
			return err
		}
		touched[key] = true
		previous[key] = stockSlot{at: at, index: slot}
		track(at, period)
	}
	if err = removed.Err(); err != nil {
		removed.Close()
		return err
	}
	removed.Close()

	if doc != nil {
		added, err := s.insertStockRows(ctx, tx, company, dt, doc, docNo, docDate, now, previous)
		if err != nil {
			return err
		}
		if len(added) > 0 {
			track(docDate, businessDay(docDate)[:7])
		}
		for _, key := range added {
			touched[key] = true
		}
	}

	keys := make([]stockKey, 0, len(touched))
	for key := range touched {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].item != keys[j].item {
			return keys[i].item < keys[j].item
		}
		return keys[i].warehouse < keys[j].warehouse
	})
	for _, key := range keys {
		if err := rebuildStockChain(ctx, tx, company, key, fromTime, fromPeriod, now); err != nil {
			return fmt.Errorf("rebuild stock %s/%s: %w", key.item, key.warehouse, err)
		}
	}
	return nil
}

// insertStockRows ลงแถวของแต่ละบรรทัดโดยยังไม่คิดยอดคงเหลือ (เป็น 0) — rebuildStockChain คิดให้ตามลำดับที่ถูกต้อง
func (s *Store) insertStockRows(ctx context.Context, tx *sql.Tx, company string, dt doctype.DocType, doc *ErpTransactionDoc, docNo string, docDate time.Time, now time.Time, previous map[stockKey]stockSlot) ([]stockKey, error) {
	direction := 1
	if dt.Stock == doctype.StockOut {
		direction = -1
	}
	periodKey := businessDay(docDate)[:7]
	var keys []stockKey
	for i, detail := range doc.Details {
		itemCode := strings.TrimSpace(detail.ItemCode)
		if itemCode == "" {
			continue
		}
		whCode := strings.TrimSpace(detail.WhCode)
		if whCode == "" {
			whCode = defaultWarehouse
		}
		qty := decimal.NewFromFloat(detail.Qty).Abs()
		price := decimal.NewFromFloat(detail.Price)
		if price.IsNegative() {
			price = decimal.Zero
		}
		key := stockKey{itemCode, whCode}
		// ตำแหน่งในวินาทีเดียวกันของสินค้า/คลังเดียวกัน: แก้เอกสารที่วันเวลาเดิมคงตำแหน่งเดิม ไม่งั้นต่อท้ายใบล่าสุด (NULL = คิดในคำสั่ง INSERT)
		var slot any
		if old, ok := previous[key]; ok && old.at.Equal(docDate) {
			slot = old.index
		}
		lineNo := i + 1
		if detail.LineNumber > 0 {
			lineNo = detail.LineNumber
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_ledger (
			businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber,
			periodkey, locationcode, barcode, unitcode, docref, transflag, direction,
			docqty, unitstand, unitdivide, price, docvalue, qty, unitcost, amount,
			balanceqty, balanceamount, avgcost, calculatedat
		) VALUES (
			$1, $2, $3, $4, COALESCE($18::integer, (SELECT MAX(behindindex) + 1 FROM stock_ledger WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3 AND docdatetime=$4), 0), $5, $6,
			$7, $8, $9, $10, $11, $12, $13,
			$14, 1, 1, $15, $16, $14, 0, 0,
			0, 0, 0, $17
		)`, company, itemCode, whCode, docDate, docNo, lineNo,
			periodKey, detail.LocationCode, detail.Barcode, detail.UnitCode, doc.DocRef, dt.Flag, direction,
			qty.String(), price.String(), qty.Mul(price).Round(8).String(), now, slot); err != nil {
			return nil, fmt.Errorf("insert stock row line %d: %w", lineNo, err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

type stockRow struct {
	at                 time.Time
	behind, line, dir  int
	docNo              string
	qty, price, docVal decimal.Decimal
	unitCost, amount   decimal.Decimal
	balQty, balAmount  decimal.Decimal
	avgCost            decimal.Decimal
	changed            bool
}

type stockBalance struct{ qty, amount, avg decimal.Decimal }

// nextStockBalance คิดแถวถัดไปของสาย (ต้นทุนถัวเฉลี่ยเคลื่อนที่): รับเข้าใช้ราคาของเอกสาร จ่ายออกใช้ต้นทุนเฉลี่ยก่อนหน้า
// (ยังไม่มีต้นทุนเฉลี่ยจึงใช้ราคาของเอกสาร) มูลค่าคงเหลือไม่ติดลบ
func nextStockBalance(dir int, qty, price, docValue decimal.Decimal, last stockBalance) (unitCost, amount decimal.Decimal, next stockBalance) {
	if dir > 0 {
		unitCost, amount = price, docValue
		next.qty = last.qty.Add(qty)
		next.amount = last.amount.Add(amount)
		if next.qty.IsPositive() {
			next.avg = next.amount.DivRound(next.qty, 8)
		} else {
			next.avg = unitCost
		}
		return
	}
	unitCost = price
	if last.avg.IsPositive() {
		unitCost = last.avg
	}
	amount = qty.Mul(unitCost).Round(8)
	next.qty = last.qty.Sub(qty)
	next.amount = last.amount.Sub(amount)
	if next.amount.IsNegative() {
		next.amount = decimal.Zero
	}
	next.avg = last.avg
	if !next.avg.IsPositive() && next.qty.IsPositive() {
		next.avg = next.amount.DivRound(next.qty, 8)
	}
	return
}

// rebuildStockChain คิดแถวของ (สินค้า, คลัง) ตั้งแต่ from เป็นต้นไปใหม่จากยอดคงเหลือของแถวก่อนหน้า แล้วปรับยอดปิดงวดตั้งแต่ fromPeriod
func rebuildStockChain(ctx context.Context, tx *sql.Tx, company string, key stockKey, from time.Time, fromPeriod string, now time.Time) error {
	var last stockBalance
	var qtyText, amountText, avgText string
	err := tx.QueryRowContext(ctx, `SELECT balanceqty::text, balanceamount::text, avgcost::text FROM stock_ledger
		WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3 AND docdatetime < $4
		ORDER BY docdatetime DESC, behindindex DESC, docno DESC, linenumber DESC LIMIT 1`, company, key.item, key.warehouse, from).Scan(&qtyText, &amountText, &avgText)
	switch {
	case err == sql.ErrNoRows:
	case err != nil:
		return err
	default:
		if last, err = parseStockBalance(qtyText, amountText, avgText); err != nil {
			return err
		}
	}

	cursor, err := tx.QueryContext(ctx, `SELECT docdatetime, behindindex, docno, linenumber, direction,
		qty::text, price::text, docvalue::text, unitcost::text, amount::text, balanceqty::text, balanceamount::text, avgcost::text
		FROM stock_ledger WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3 AND docdatetime >= $4
		ORDER BY docdatetime, behindindex, docno, linenumber FOR UPDATE`, company, key.item, key.warehouse, from)
	if err != nil {
		return err
	}
	var chain []stockRow
	for cursor.Next() {
		var r stockRow
		var texts [8]string
		if err = cursor.Scan(&r.at, &r.behind, &r.docNo, &r.line, &r.dir, &texts[0], &texts[1], &texts[2], &texts[3], &texts[4], &texts[5], &texts[6], &texts[7]); err != nil {
			cursor.Close()
			return err
		}
		values := [8]*decimal.Decimal{&r.qty, &r.price, &r.docVal, &r.unitCost, &r.amount, &r.balQty, &r.balAmount, &r.avgCost}
		for i, text := range texts {
			if *values[i], err = decimal.NewFromString(text); err != nil {
				cursor.Close()
				return err
			}
		}
		chain = append(chain, r)
	}
	if err = cursor.Err(); err != nil {
		cursor.Close()
		return err
	}
	cursor.Close()

	for i := range chain {
		r := &chain[i]
		unitCost, amount, next := nextStockBalance(r.dir, r.qty, r.price, r.docVal, last)
		if !unitCost.Equal(r.unitCost) || !amount.Equal(r.amount) || !next.qty.Equal(r.balQty) || !next.amount.Equal(r.balAmount) || !next.avg.Equal(r.avgCost) {
			if _, err = tx.ExecContext(ctx, `UPDATE stock_ledger SET unitcost=$1, amount=$2, balanceqty=$3, balanceamount=$4, avgcost=$5, calculatedat=$6
				WHERE businesscode=$7 AND itemcode=$8 AND whcode=$9 AND docdatetime=$10 AND behindindex=$11 AND docno=$12 AND linenumber=$13`,
				unitCost.String(), amount.String(), next.qty.String(), next.amount.String(), next.avg.String(), now,
				company, key.item, key.warehouse, r.at, r.behind, r.docNo, r.line); err != nil {
				return err
			}
		}
		last = next
	}

	if fromPeriod == "" {
		return nil
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO stock_period_balance (businesscode, itemcode, whcode, periodkey, closeqty, closeamount, closeavgcost, hastrans, updatedat)
		SELECT DISTINCT ON (periodkey) businesscode, itemcode, whcode, periodkey, balanceqty, balanceamount, avgcost, true, $5
		FROM stock_ledger WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3 AND periodkey >= $4
		ORDER BY periodkey, docdatetime DESC, behindindex DESC, docno DESC, linenumber DESC
		ON CONFLICT (businesscode, itemcode, whcode, periodkey)
		DO UPDATE SET closeqty=EXCLUDED.closeqty, closeamount=EXCLUDED.closeamount, closeavgcost=EXCLUDED.closeavgcost, hastrans=true, updatedat=EXCLUDED.updatedat`,
		company, key.item, key.warehouse, fromPeriod, now); err != nil {
		return err
	}
	// งวดที่เคยมีความเคลื่อนไหวแต่ตอนนี้ไม่เหลือแถวสต็อกแล้ว (เอกสารถูกลบ/ย้ายวันที่) ต้องไม่ค้างยอดปิดเก่า
	_, err = tx.ExecContext(ctx, `DELETE FROM stock_period_balance WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3 AND periodkey >= $4 AND hastrans
		AND periodkey NOT IN (SELECT periodkey FROM stock_ledger WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3)`,
		company, key.item, key.warehouse, fromPeriod)
	return err
}

func parseStockBalance(qty, amount, avg string) (stockBalance, error) {
	var b stockBalance
	var err error
	if b.qty, err = decimal.NewFromString(qty); err != nil {
		return b, err
	}
	if b.amount, err = decimal.NewFromString(amount); err != nil {
		return b, err
	}
	b.avg, err = decimal.NewFromString(avg)
	return b, err
}

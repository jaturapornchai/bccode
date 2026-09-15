package stockengine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// ไฟล์นี้คือทางเดียวที่รายงานใช้อ่านยอดคงเหลือ เพื่อให้ทุกจอได้ตัวเลขชุดเดียวกัน
//
// ยอดคงเหลือต่อคลังอ่านจาก "แถวล่าสุด" ของสมุดสต็อก ไม่ได้บวกใหม่ทุกครั้ง
// เพราะ balanceqty/balanceamount ถูกคิดไว้แล้วตามลำดับต้นทุนที่ถูกต้อง
// ส่วนยอดต่อที่เก็บใช้การรวม direction * qty เพราะต้นทุนถัวเฉลี่ยคิดที่ระดับคลัง ไม่ใช่ที่เก็บ

// ลำดับนี้ต้องตรงกับคีย์หลักของ stock_ledger แถวสุดท้ายจึงเป็นแถวล่าสุดจริง
const ledgerLatestOrder = "docdatetime DESC, behindindex DESC, docno DESC, linenumber DESC"

// BalanceFilter ระบุขอบเขตการอ่านยอดคงเหลือ
//
// DateCondition คือเงื่อนไขวันที่ที่ผู้เรียกสร้างไว้แล้ว (เช่นจาก mypg.BuildDateConditionSQLWithTimeZone)
// ต้องอ้างตัวแปร $1 ตัวเดียวและอ้างคอลัมน์ docdatetime เท่านั้น
type BalanceFilter struct {
	BusinessCode  string
	DateCondition string
	DateArg       any
	ItemCodes     []string
}

// where สร้างเงื่อนไขและลำดับตัวแปรให้ทุก query ใช้ร่วมกัน
func (f BalanceFilter) where() (string, []any) {
	conditions := []string{}
	args := []any{}

	if f.DateCondition != "" {
		conditions = append(conditions, f.DateCondition)
		args = append(args, f.DateArg)
	}

	args = append(args, f.BusinessCode)
	conditions = append(conditions, fmt.Sprintf("businesscode = $%d", len(args)))

	if len(f.ItemCodes) > 0 {
		args = append(args, pq.Array(f.ItemCodes))
		conditions = append(conditions, fmt.Sprintf("itemcode = ANY($%d)", len(args)))
	}

	return strings.Join(conditions, " AND "), args
}

// WarehouseBalance ยอดคงเหลือของสินค้าหนึ่งตัวในคลังหนึ่ง
type WarehouseBalance struct {
	ItemCode string
	WhCode   string
	Qty      float64
	Amount   float64
	AvgCost  float64
}

// LocationBalance ยอดคงเหลือของสินค้าหนึ่งตัวในที่เก็บหนึ่ง (จำนวนอย่างเดียว ไม่มีต้นทุน)
type LocationBalance struct {
	ItemCode     string
	WhCode       string
	LocationCode string
	Qty          float64
}

// ItemBalance ยอดคงเหลือรวมทุกคลังของสินค้าหนึ่งตัว
type ItemBalance struct {
	ItemCode string
	Qty      float64
	Amount   float64
	AvgCost  float64
}

// LoadWarehouseBalances อ่านยอดคงเหลือแยกตามคลัง ณ วันที่ที่กำหนด
func LoadWarehouseBalances(ctx context.Context, db *sql.DB, filter BalanceFilter) ([]WarehouseBalance, error) {
	where, args := filter.where()

	query := fmt.Sprintf(`
		SELECT DISTINCT ON (itemcode, whcode)
			itemcode, whcode, balanceqty, balanceamount, avgcost
		FROM stock_ledger
		WHERE %s
		ORDER BY itemcode, whcode, %s`, where, ledgerLatestOrder)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load warehouse balances: %w", err)
	}
	defer rows.Close()

	balances := []WarehouseBalance{}
	for rows.Next() {
		var b WarehouseBalance
		if err := rows.Scan(&b.ItemCode, &b.WhCode, &b.Qty, &b.Amount, &b.AvgCost); err != nil {
			return nil, fmt.Errorf("scan warehouse balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

// LoadLocationBalances อ่านยอดคงเหลือแยกตามที่เก็บ ณ วันที่ที่กำหนด
func LoadLocationBalances(ctx context.Context, db *sql.DB, filter BalanceFilter) ([]LocationBalance, error) {
	where, args := filter.where()

	query := fmt.Sprintf(`
		SELECT itemcode, whcode, locationcode, SUM(direction * qty) AS balanceqty
		FROM stock_ledger
		WHERE %s
		GROUP BY itemcode, whcode, locationcode
		ORDER BY itemcode, whcode, locationcode`, where)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load location balances: %w", err)
	}
	defer rows.Close()

	balances := []LocationBalance{}
	for rows.Next() {
		var b LocationBalance
		if err := rows.Scan(&b.ItemCode, &b.WhCode, &b.LocationCode, &b.Qty); err != nil {
			return nil, fmt.Errorf("scan location balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

// LoadItemBalances อ่านยอดคงเหลือรวมทุกคลัง ณ วันที่ที่กำหนด
//
// ต้นทุนถัวเฉลี่ยระดับสินค้าคำนวณจากมูลค่ารวมหารจำนวนรวม ไม่ใช่หยิบค่าของคลังใดคลังหนึ่งมาใช้
func LoadItemBalances(ctx context.Context, db *sql.DB, filter BalanceFilter) ([]ItemBalance, error) {
	where, args := filter.where()

	query := fmt.Sprintf(`
		WITH perwarehouse AS (
			SELECT DISTINCT ON (itemcode, whcode)
				itemcode, balanceqty, balanceamount
			FROM stock_ledger
			WHERE %s
			ORDER BY itemcode, whcode, %s
		)
		SELECT
			itemcode,
			SUM(balanceqty) AS balanceqty,
			SUM(balanceamount) AS balanceamount,
			CASE WHEN SUM(balanceqty) <> 0 THEN SUM(balanceamount) / SUM(balanceqty) ELSE 0 END AS avgcost
		FROM perwarehouse
		GROUP BY itemcode
		ORDER BY itemcode`, where, ledgerLatestOrder)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load item balances: %w", err)
	}
	defer rows.Close()

	balances := []ItemBalance{}
	for rows.Next() {
		var b ItemBalance
		if err := rows.Scan(&b.ItemCode, &b.Qty, &b.Amount, &b.AvgCost); err != nil {
			return nil, fmt.Errorf("scan item balance: %w", err)
		}
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

package cheque

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ยอดที่ derive จากทะเบียน + เอกสารสถานะ แทนยอดสะสมของ Champ (ChqOnHand/ChqReturn ของลูกหนี้, ยอดเช็คจ่ายค้างของเจ้าหนี้)
// — ไม่มีคอลัมน์ยอดสะสมให้คลาดกัน จึงไม่ต้องล็อกยอดเมื่อแก้ทะเบียน (D8) (cheque-bank.md §4.5)

// PartnerBalance คือยอดเช็คค้างของลูกหนี้/เจ้าหนี้หนึ่งราย (สตริงทศนิยม 2 ตำแหน่ง)
type PartnerBalance struct {
	PartnerCode       string `json:"partner_code"`
	PartnerName       string `json:"partner_name"`
	ReceivedOnHand    string `json:"received_on_hand"`   // เช็ครับในมือ + นำฝากแล้วยังไม่ผ่าน (สถานะ 0, 1)
	ReceivedReturned  string `json:"received_returned"`  // เช็ครับคืน + นำเข้าใหม่ (สถานะ 3, 6)
	IssuedOutstanding string `json:"issued_outstanding"` // เช็คจ่ายที่ยังไม่ผ่าน + เตรียมจ่าย (สถานะ 0, 3)
}

// BalanceFilter คือเงื่อนไขค้นยอดคงค้างรายผู้ติดต่อ
type BalanceFilter struct {
	Partner string
	Page    int
	Limit   int
}

// PartnerBalancePage คือผลหนึ่งหน้า
type PartnerBalancePage struct {
	Items []PartnerBalance
	Total int
	Page  int
	Limit int
}

// PartnerBalances คือยอดเช็คค้างรายผู้ติดต่อ เฉพาะรายที่มียอดค้างอย่างน้อยหนึ่งช่อง เรียงตามรหัส
func PartnerBalances(ctx context.Context, q querier, company string, f BalanceFilter) (PartnerBalancePage, error) {
	page, limit, offset := pageWindow(f.Page, f.Limit, defaultRegisterLimit, maxRegisterLimit)
	result := PartnerBalancePage{Items: []PartnerBalance{}, Page: page, Limit: limit}
	args := []any{company}
	where := "r.company = $1 AND r.deleted_at IS NULL"
	if p := strings.TrimSpace(f.Partner); p != "" {
		args = append(args, p)
		where += " AND r.partner_code = $" + strconv.Itoa(len(args))
	}
	const sums = `SUM(r.home_amount) FILTER (WHERE r.cheque_type = 0 AND r.status IN (0,1)) AS on_hand,
		SUM(r.home_amount) FILTER (WHERE r.cheque_type = 0 AND r.status IN (3,6)) AS returned,
		SUM(r.home_amount) FILTER (WHERE r.cheque_type = 1 AND r.status IN (0,3)) AS issued`
	grouped := `SELECT r.partner_code, ` + sums + ` FROM cheque_registers r WHERE ` + where + ` GROUP BY r.partner_code
		HAVING COUNT(*) FILTER (WHERE (r.cheque_type = 0 AND r.status IN (0,1,3,6)) OR (r.cheque_type = 1 AND r.status IN (0,3))) > 0`
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM (`+grouped+`) g`, args...).Scan(&result.Total); err != nil {
		return PartnerBalancePage{}, err
	}
	n := len(args)
	rows, err := q.QueryContext(ctx, `SELECT g.partner_code, COALESCE(p.payload->>'name_th', ''),
			COALESCE(g.on_hand, 0)::numeric(18,2)::text, COALESCE(g.returned, 0)::numeric(18,2)::text, COALESCE(g.issued, 0)::numeric(18,2)::text
		FROM (`+grouped+`) g
		LEFT JOIN gl_subledger_partners p ON p.company = $1 AND p.code = g.partner_code
		ORDER BY g.partner_code`+fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n+1, n+2), append(args, limit, offset)...)
	if err != nil {
		return PartnerBalancePage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var b PartnerBalance
		if err = rows.Scan(&b.PartnerCode, &b.PartnerName, &b.ReceivedOnHand, &b.ReceivedReturned, &b.IssuedOutstanding); err != nil {
			return PartnerBalancePage{}, err
		}
		result.Items = append(result.Items, b)
	}
	return result, rows.Err()
}

// BookMovement คือยอดเช็คที่ผ่านเข้า/ออกสมุดเงินฝากหนึ่งเล่มในช่วงวันที่ — ใช้ตรวจกับยอด GL ของบัญชีสมุด
type BookMovement struct {
	BookCode string `json:"book_code"`
	In       string `json:"in_amount"`
	Out      string `json:"out_amount"`
	Net      string `json:"net_amount"`
}

// bookSignSQL คือทิศของยอดสุทธิต่อสมุดตามชนิดเอกสาร: เข้า = รับผ่าน/นำเข้าใหม่/ขายลด และยกเลิกเช็คจ่ายที่ผ่านแล้ว;
// ออก = รับคืน/จ่ายผ่าน และยกเลิกเช็ครับที่ผ่านแล้วหรือนำเข้าใหม่แล้ว; นำฝาก/ยกเลิกเช็คที่ยังไม่ผ่านไม่กระทบสมุด
// (cheque-bank.md §4.5 — กติกาของ 423 ชนิด 2 และ 426 ที่ไม่ติ๊กยังต้องให้ลุงจืดยืนยันตอนรับงาน)
const bookSignSQL = `CASE
	WHEN d.transflag IN (420, 422, 424) THEN 1
	WHEN d.transflag IN (421, 425) THEN -1
	WHEN d.transflag = 423 AND d.cancel_type IN (1, 3) THEN -1
	WHEN d.transflag = 426 AND d.cancel_passed THEN 1
	ELSE 0 END`

// BookMovements คือยอดเข้า/ออกของทุกสมุด (หรือสมุดเดียว) จากเอกสารสถานะที่ยังใช้งาน ในช่วงวันที่เอกสาร (ISO; ว่าง = ไม่จำกัด)
func BookMovements(ctx context.Context, q querier, company, book, from, to string) ([]BookMovement, error) {
	args := []any{company}
	where := []string{"d.company = $1", "d.status = 'active'", "d.book_code IS NOT NULL"}
	add := func(cond string, value any) {
		args = append(args, value)
		where = append(where, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(args))))
	}
	if b := strings.TrimSpace(book); b != "" {
		add("d.book_code = ?", b)
	}
	for _, d := range []struct{ value, field, cond string }{{from, "date_from", "d.doc_date >= ?"}, {to, "date_to", "d.doc_date <= ?"}} {
		if s := strings.TrimSpace(d.value); s != "" {
			if !validISODate(s) {
				return nil, badRequest("cheque_err_date_invalid", d.field)
			}
			add(d.cond, s)
		}
	}
	rows, err := q.QueryContext(ctx, `SELECT book_code, COALESCE(SUM(net_amount) FILTER (WHERE sign > 0), 0)::numeric(18,2)::text,
			COALESCE(SUM(net_amount) FILTER (WHERE sign < 0), 0)::numeric(18,2)::text, SUM(sign * net_amount)::numeric(18,2)::text
		FROM (SELECT d.book_code, d.net_amount, `+bookSignSQL+` AS sign FROM cheque_docs d WHERE `+strings.Join(where, " AND ")+`) m
		WHERE sign <> 0 GROUP BY book_code ORDER BY book_code`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BookMovement{}
	for rows.Next() {
		var m BookMovement
		if err = rows.Scan(&m.BookCode, &m.In, &m.Out, &m.Net); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

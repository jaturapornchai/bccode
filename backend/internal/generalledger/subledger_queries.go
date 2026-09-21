package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// All balances are computed by PostgreSQL NUMERIC. Draft links reserve capacity,
// but only posted journals contribute to confirmed settlements and matches.
const subledgerReadCTE = `WITH scoped_journals AS (
 SELECT r.id,r.payload FROM gl_records r WHERE r.company=$1 AND r.kind='journals'
 AND ($2='' OR r.payload->>'branchcode'=$2)
), visible_journals AS (SELECT id,payload FROM scoped_journals WHERE NOT COALESCE((payload->>'isdeleted')::boolean,false)), posted_journals AS (
 SELECT r.id,r.payload FROM visible_journals r WHERE r.payload->>'date'<=$4
 AND (r.payload->>'postedat' IS NULL OR (r.payload->>'postedat')::timestamptz < (($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok'))
 AND (r.payload->>'status'='posted' OR (r.payload->>'status'='reversed' AND NOT EXISTS(
 SELECT 1 FROM gl_records reversal WHERE reversal.company=$1 AND reversal.kind='journals'
 AND reversal.payload->>'reversalof'=r.id AND reversal.payload->>'date'<=$4 AND (reversal.payload->>'postedat' IS NULL OR (reversal.payload->>'postedat')::timestamptz < (($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok')))))
), active_allocations AS (
 SELECT a.* FROM gl_subledger_allocations a WHERE a.company=$1 AND a.created_at<(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok') AND (a.reversed_at IS NULL OR a.reversed_at>=(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok') OR a.reversed_effective_date>$4::date)
), active_settlements AS (
 SELECT s.* FROM gl_subledger_settlements s JOIN posted_journals j ON j.id=s.journal_id WHERE s.company=$1
 AND s.created_at<(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok') AND s.settlement_date<=$4::date AND (s.reversed_at IS NULL OR s.reversed_at>=(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok'))
), active_matches AS (
 SELECT m.* FROM gl_subledger_matches m JOIN posted_journals own ON own.id=m.owner_journal_id
 JOIN posted_journals target ON target.id=m.journal_id WHERE m.company=$1 AND m.created_at<(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok')
 AND (m.reversed_at IS NULL OR m.reversed_at>=(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok'))
) `

func subledgerSelect(kind string) (string, error) {
	switch kind {
	case "partners", "bank-accounts":
		table := "gl_subledger_partners"
		if kind == "bank-accounts" {
			table = "gl_subledger_bank_accounts"
		}
		return `SELECT d.code AS sortkey,d.payload FROM ` + table + ` d WHERE d.company=$1 AND ($3='' OR d.code ILIKE $3 OR d.payload->>'name_th' ILIKE $3 OR d.payload->>'bank_name' ILIKE $3 OR d.payload->>'account_name' ILIKE $3)`, nil
	case "documents":
		return `SELECT d.id AS sortkey,d.payload||jsonb_build_object('allocated_amount',COALESCE(a.allocated,0)::text,'posted_amount',COALESCE(a.posted,0)::text,'settled_amount',COALESCE(s.settled,0)::text,'remaining_amount',(COALESCE(a.posted,0)-COALESCE(s.settled,0))::text) AS payload
 FROM gl_subledger_documents d
 LEFT JOIN LATERAL(SELECT SUM(a.amount) FILTER(WHERE j.payload->>'status' IN('draft','posted') OR p.id IS NOT NULL) AS allocated,SUM(a.amount) FILTER(WHERE p.id IS NOT NULL) AS posted FROM active_allocations a JOIN visible_journals j ON j.id=a.journal_id LEFT JOIN posted_journals p ON p.id=a.journal_id WHERE a.document_id=d.id) a ON true
 LEFT JOIN LATERAL(SELECT SUM(s.amount) AS settled FROM active_settlements s WHERE s.debt_id=d.id OR s.payment_id=d.id) s ON true
 WHERE d.company=$1 AND ($2='' OR d.branch_code=$2) AND d.payload->>'document_date'<=$4
 AND ($3='' OR d.id ILIKE $3 OR d.payload->>'document_no' ILIKE $3 OR d.partner_code ILIKE $3)`, nil
	case "statements":
		return `SELECT d.id AS sortkey,d.payload||jsonb_build_object('owner_journal_id',d.owner_journal_id,'matched_amount',COALESCE(m.matched,0)::text,'remaining_amount',(d.amount-COALESCE(m.matched,0))::text) AS payload
 FROM gl_subledger_statements d JOIN scoped_journals own ON own.id=d.owner_journal_id
 LEFT JOIN LATERAL(SELECT SUM(m.amount) AS matched FROM active_matches m WHERE m.statement_id=d.id) m ON true
 WHERE d.company=$1 AND d.created_at<(($4::date+1)::timestamp AT TIME ZONE 'Asia/Bangkok') AND d.transaction_date<=$4::date AND ($3='' OR d.id ILIKE $3 OR d.owner_journal_id ILIKE $3 OR d.source_key ILIKE $3 OR d.payload->>'bank_reference' ILIKE $3 OR d.bank_account_code ILIKE $3)`, nil
	case "bank-lines":
		return `SELECT d.journal_id||':'||lpad(d.line_number::text,10,'0') AS sortkey,d.payload||jsonb_build_object('docno',j.payload->>'docno','date',j.payload->>'date','amount',d.amount::text,'matched_amount',COALESCE(m.matched,0)::text,'remaining_amount',(d.amount-COALESCE(m.matched,0))::text) AS payload
 FROM gl_subledger_bank_lines d JOIN posted_journals j ON j.id=d.journal_id
 LEFT JOIN LATERAL(SELECT SUM(m.amount) AS matched FROM active_matches m WHERE m.journal_id=d.journal_id AND m.line_number=d.line_number) m ON true
 WHERE d.company=$1 AND ($3='' OR d.journal_id ILIKE $3 OR (d.journal_id||':'||d.line_number::text) ILIKE $3 OR j.payload->>'docno' ILIKE $3 OR d.bank_account_code ILIKE $3)`, nil
	case "allocations":
		return `SELECT d.id AS sortkey,d.payload||jsonb_build_object('docno',j.payload->>'docno','document_no',doc.payload->>'document_no','created_at',d.created_at,'reversed_at',d.reversed_at,'reversed_by',d.reversed_by,'reversal_reason',d.reversal_reason) AS payload
 FROM gl_subledger_allocations d JOIN visible_journals j ON j.id=d.journal_id JOIN gl_subledger_documents doc ON doc.company=d.company AND doc.id=d.document_id
 WHERE d.company=$1 AND ($2='' OR doc.branch_code=$2) AND ($3='' OR d.id ILIKE $3 OR d.journal_id ILIKE $3 OR d.document_id ILIKE $3 OR j.payload->>'docno' ILIKE $3 OR doc.payload->>'document_no' ILIKE $3)`, nil
	case "settlements":
		return `SELECT d.id AS sortkey,d.payload||jsonb_build_object('journal_id',d.journal_id,'docno',j.payload->>'docno','debt_document_no',debt.payload->>'document_no','payment_document_no',payment.payload->>'document_no','created_at',d.created_at,'reversed_at',d.reversed_at,'reversed_by',d.reversed_by,'reversal_reason',d.reversal_reason) AS payload
 FROM gl_subledger_settlements d JOIN visible_journals j ON j.id=d.journal_id JOIN gl_subledger_documents debt ON debt.company=d.company AND debt.id=d.debt_id JOIN gl_subledger_documents payment ON payment.company=d.company AND payment.id=d.payment_id
 WHERE d.company=$1 AND ($2='' OR (debt.branch_code=$2 AND payment.branch_code=$2)) AND d.settlement_date<=$4::date
 AND ($3='' OR d.id ILIKE $3 OR d.journal_id ILIKE $3 OR d.debt_id ILIKE $3 OR d.payment_id ILIKE $3 OR debt.payload->>'document_no' ILIKE $3 OR payment.payload->>'document_no' ILIKE $3 OR j.payload->>'docno' ILIKE $3)`, nil
	case "matches":
		return `SELECT d.id AS sortkey,d.payload||jsonb_build_object('owner_journal_id',d.owner_journal_id,'docno',j.payload->>'docno','bank_reference',statement.payload->>'bank_reference','created_at',d.created_at,'reversed_at',d.reversed_at,'reversed_by',d.reversed_by,'reversal_reason',d.reversal_reason) AS payload
 FROM gl_subledger_matches d JOIN visible_journals j ON j.id=d.journal_id JOIN visible_journals own ON own.id=d.owner_journal_id JOIN gl_subledger_statements statement ON statement.company=d.company AND statement.id=d.statement_id JOIN scoped_journals imported ON imported.id=statement.owner_journal_id
 WHERE d.company=$1 AND ($3='' OR d.id ILIKE $3 OR d.journal_id ILIKE $3 OR d.owner_journal_id ILIKE $3 OR d.statement_id ILIKE $3 OR j.payload->>'docno' ILIKE $3 OR statement.payload->>'bank_reference' ILIKE $3)`, nil
	}
	return "", fmt.Errorf("ประเภทข้อมูลประกอบไม่ถูกต้อง")
}

func (p *Postgres) SubledgerList(ctx context.Context, scope Scope, kind, q string, page, limit int, asOf string) (Page, error) {
	out := Page{Items: []json.RawMessage{}}
	selectSQL, err := subledgerSelect(kind)
	if err != nil {
		return out, err
	}
	if asOf == "" {
		asOf = "9999-12-30"
	}
	if !validDate(asOf) {
		return out, fmt.Errorf("วันที่ข้อมูลประกอบไม่ถูกต้อง")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if page > 1000000 {
		return out, fmt.Errorf("เลขหน้ามากเกินกำหนด")
	}
	if len(q) > 200 {
		return out, fmt.Errorf("คำค้นยาวเกินกำหนด")
	}
	q = strings.TrimSpace(q)
	if q != "" {
		q = "%" + q + "%"
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return out, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	args := []any{scope.Company, scope.Branch, q, asOf}
	if err = tx.QueryRowContext(ctx, subledgerReadCTE+`SELECT COUNT(*) FROM (`+selectSQL+`) rows`, args...).Scan(&out.Total); err != nil {
		return out, err
	}
	rows, err := tx.QueryContext(ctx, subledgerReadCTE+selectSQL+` ORDER BY sortkey LIMIT $5 OFFSET $6`, append(args, limit, (page-1)*limit)...)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var key string
		var raw []byte
		if err = rows.Scan(&key, &raw); err != nil {
			rows.Close()
			return out, err
		}
		out.Items = append(out.Items, json.RawMessage(raw))
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT sequence FROM gl_projection_state WHERE company=$1),0)`, scope.Company).Scan(&out.Sequence); err != nil {
		return out, err
	}
	out.Page, out.Limit = page, limit
	err = tx.Commit()
	return out, err
}

func (p *Postgres) SubledgerReport(ctx context.Context, scope Scope, name string, q ReportQuery) (Report, error) {
	report := Report{Rows: []map[string]string{}, Totals: map[string]string{}, Warnings: []string{}}
	kind, filter := "documents", ""
	money := []string{"posted_amount", "settled_amount", "remaining_amount"}
	report.Columns = []ReportColumn{{Key: "document_no", Label: "เลขที่เอกสาร"}, {Key: "partner_code", Label: "คู่ค้า"}, {Key: "balance_side", Label: "ด้านหนี้"}, {Key: "document_date", Label: "วันที่"}, {Key: "due_date", Label: "ครบกำหนด"}, {Key: "posted_amount", Label: "ผ่านบัญชี", Amount: true}, {Key: "settled_amount", Label: "ตัดยอดแล้ว", Amount: true}, {Key: "remaining_amount", Label: "คงเหลือสุทธิ", Amount: true}}
	switch name {
	case "ar-outstanding":
		filter = `payload->>'ledger'='ar'`
	case "ap-outstanding":
		filter = `payload->>'ledger'='ap'`
	case "bank-unmatched":
		kind = "statements"
		filter = "true"
		money = []string{"amount", "matched_amount", "remaining_amount"}
		report.Columns = []ReportColumn{{Key: "bank_account_code", Label: "บัญชีธนาคาร"}, {Key: "transaction_date", Label: "วันที่"}, {Key: "bank_reference", Label: "อ้างอิง"}, {Key: "direction", Label: "รับ/จ่าย"}, {Key: "amount", Label: "ยอด Statement", Amount: true}, {Key: "matched_amount", Label: "จับคู่แล้ว", Amount: true}, {Key: "remaining_amount", Label: "ยังไม่จับคู่", Amount: true}}
	default:
		return report, ErrNotFound
	}
	if q.BranchCode != "" {
		if scope.Branch != "" && scope.Branch != q.BranchCode {
			return report, ErrNotFound
		}
		scope.Branch = q.BranchCode
	}
	asof := q.To
	if asof == "" {
		asof = time.Now().UTC().Format("2006-01-02")
	}
	if !validDate(asof) {
		return report, fmt.Errorf("วันที่รายงานไม่ถูกต้อง")
	}
	report.AsOf = asof
	selectSQL, err := subledgerSelect(kind)
	if err != nil {
		return report, err
	}
	if kind == "documents" {
		signed := []string{}
		for _, key := range money {
			signed = append(signed, "'"+key+"',((payload->>'"+key+"')::numeric * CASE WHEN payload->>'balance_side'='2' THEN -1 ELSE 1 END)::text")
		}
		selectSQL = "SELECT sortkey,payload||jsonb_build_object(" + strings.Join(signed, ",") + ") AS payload FROM (" + selectSQL + ") gross"
	}
	base := subledgerReadCTE + `, balances AS (` + selectSQL + `) `
	where := ` WHERE ` + filter + ` AND (payload->>'remaining_amount')::numeric<>0`
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return report, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return report, err
	}
	defer tx.Rollback()
	args := []any{scope.Company, scope.Branch, "", asof}
	aggregates := []string{"COUNT(*)"}
	for _, field := range money {
		aggregates = append(aggregates, `COALESCE(SUM((payload->>'`+field+`')::numeric),0)::text`)
	}
	values := make([]string, len(money))
	scan := []any{&report.TotalRows}
	for i := range values {
		scan = append(scan, &values[i])
	}
	if err = tx.QueryRowContext(ctx, base+`SELECT `+strings.Join(aggregates, ",")+` FROM balances`+where, args...).Scan(scan...); err != nil {
		return report, err
	}
	for i, key := range money {
		report.Totals[key] = values[i]
	}
	page, limit := q.Page, q.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	if page > 1000000 {
		return report, fmt.Errorf("เลขหน้ามากเกินกำหนด")
	}
	rows, err := tx.QueryContext(ctx, base+`SELECT payload FROM balances`+where+` ORDER BY sortkey LIMIT $5 OFFSET $6`, append(args, limit, (page-1)*limit)...)
	if err != nil {
		return report, err
	}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return report, err
		}
		var item map[string]json.RawMessage
		if err = json.Unmarshal(raw, &item); err != nil {
			rows.Close()
			return report, err
		}
		row := map[string]string{}
		for key, value := range item {
			var text string
			if json.Unmarshal(value, &text) == nil {
				row[key] = text
			} else {
				row[key] = string(value)
			}
		}
		report.Rows = append(report.Rows, row)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return report, err
	}
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT sequence FROM gl_projection_state WHERE company=$1),0)`, scope.Company).Scan(&report.Sequence); err != nil {
		return report, err
	}
	if report.TotalRows > int64(len(report.Rows)) {
		report.Warnings = append(report.Warnings, "แสดงหน้าที่ "+strconv.Itoa(page)+" ยอดรวมคำนวณจากทุกรายการ")
	}
	err = tx.Commit()
	return report, err
}

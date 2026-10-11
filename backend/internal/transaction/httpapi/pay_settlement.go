package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// prepareLineUseSQL คือยอดที่ใบจ่ายชำระหนี้ใช้จากบรรทัด l ของใบเตรียมจ่าย (Champ PaymentProcessBill ลดยอดเตรียมจ่ายเมื่อบันทึกใบจ่าย):
// settled = การตัดหนี้ที่ยังมีผลซึ่งระบุใบเตรียมจ่ายนี้, pending = บรรทัดใบจ่ายที่บันทึกแล้วแต่ยังไม่เกิดการตัดหนี้ (เช่นยังไม่มีผังลงบัญชี)
// params คงที่: $1 company, $2 transflag ใบจ่ายชำระหนี้, $3 prefix การตัดหนี้ของใบจ่าย, $4 ใบจ่ายที่ไม่นับ (ค่าว่าง = ไม่มี), $5 id บัญชีย่อยของใบนั้น
const prepareLineUseSQL = `COALESCE((SELECT SUM(s.amount) FROM gl_subledger_settlements s WHERE s.company=$1 AND s.reversed_at IS NULL
 AND s.debt_id=l.debt_document_id AND s.payload->>'prepare_docno'=l.docno AND s.payment_id<>$5),0) AS settled,
 COALESCE((SELECT SUM(dd.sumamount) FROM docdetail dd JOIN doc h ON h.businesscode=dd.businesscode AND h.transflag=dd.transflag AND h.docno=dd.docno
 WHERE dd.businesscode=$1 AND dd.transflag=$2 AND dd.paybillno=l.docno AND dd.debt_document_id=l.debt_document_id AND dd.docno<>$4
 AND NOT COALESCE(h.isdelete,false) AND NOT COALESCE(h.iscancel,false)
 AND NOT EXISTS(SELECT 1 FROM gl_subledger_settlements s WHERE s.company=$1 AND s.reversed_at IS NULL AND s.id=$3::text||dd.docno||'-'||dd.linenumber)),0) AS pending`

// prepareUseArgs คือ $1-$5 ของ prepareLineUseSQL
func prepareUseArgs(company, excludePV string) []any {
	spec, excludeID := subledgerSpecs["pay"], ""
	if excludePV != "" {
		excludeID = spec.docPrefix + excludePV
	}
	return []any{company, doctype.Pay, spec.settlePrefix, excludePV, excludeID}
}

// prepareLineUse คือบรรทัดหนึ่งของใบเตรียมจ่ายกับยอดที่ใบจ่ายใช้ไปแล้ว
type prepareLineUse struct {
	docNo, debtID       string
	lineNumber          int
	payAmount, settled  decimal.Decimal
	pending             decimal.Decimal
	documentNo          string
	documentDate        string
	debtAmount, balance string
	billType, sign      int
	description         string
}

// paid = ยอดที่ใบจ่ายใช้จากบรรทัดนี้ (ตัดหนี้แล้ว + บันทึกแล้วรอตัด)
func (u prepareLineUse) paid() decimal.Decimal { return u.settled.Add(u.pending) }

// loadPrepareLineUse อ่านทุกบรรทัดของใบเตรียมจ่ายที่ระบุพร้อมยอดที่ใช้ไป ในคำสั่งเดียว (ไม่วนอ่านทีละบรรทัด)
func loadPrepareLineUse(ctx context.Context, tx *sql.Tx, company string, prepareNos []string, excludePV string) ([]prepareLineUse, error) {
	rows, err := tx.QueryContext(ctx, `SELECT l.docno, l.debt_document_id, l.linenumber, l.payamount::text, u.settled::text, u.pending::text,
		l.document_no, l.document_date::text, l.debt_amount::text, l.invbalance::text, l.billtype, l.sign, l.description
		FROM ap_payment_prepare_lines l CROSS JOIN LATERAL (SELECT `+prepareLineUseSQL+`) u
		WHERE l.company=$1 AND l.docno=ANY($6) ORDER BY l.docno, l.linenumber`, append(prepareUseArgs(company, excludePV), pq.Array(prepareNos))...)
	if err != nil {
		return nil, fmt.Errorf("read prepare lines: %w", err)
	}
	defer rows.Close()
	var out []prepareLineUse
	for rows.Next() {
		var u prepareLineUse
		var pay, settled, pending string
		if err = rows.Scan(&u.docNo, &u.debtID, &u.lineNumber, &pay, &settled, &pending,
			&u.documentNo, &u.documentDate, &u.debtAmount, &u.balance, &u.billType, &u.sign, &u.description); err != nil {
			return nil, fmt.Errorf("scan prepare line: %w", err)
		}
		if u.payAmount, err = decimal.NewFromString(pay); err == nil {
			if u.settled, err = decimal.NewFromString(settled); err == nil {
				u.pending, err = decimal.NewFromString(pending)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("parse prepare line amounts: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// checkNotReserved: เอกสารหนี้ที่ใบเตรียมจ่าย (ยังไม่ยกเลิก) จองไว้และบรรทัดนั้นยังตัดหนี้ไม่ครบ แก้/ลบ/ยกเลิกไม่ได้
// (ไม่มี FK ไปแถวหนี้ เพื่อตอบเป็นข้อความไทยแทน error 23503 — design §4.1)
func checkNotReserved(ctx context.Context, tx *sql.Tx, company, debtID, docNo string) error {
	var prepareNo string
	err := tx.QueryRowContext(ctx, `SELECT l.docno FROM ap_payment_prepare_lines l
		JOIN ap_payment_prepare p ON p.company=l.company AND p.docno=l.docno AND NOT p.iscancel
		WHERE l.company=$1 AND l.debt_document_id=$2 AND l.payamount > COALESCE((SELECT SUM(s.amount) FROM gl_subledger_settlements s
			WHERE s.company=$1 AND s.reversed_at IS NULL AND s.debt_id=l.debt_document_id AND s.payload->>'prepare_docno'=l.docno),0)
		ORDER BY l.docno LIMIT 1`, company, debtID).Scan(&prepareNo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check payment preparation reservation: %w", err)
	}
	return erpError("ap_prepare_reserved", "", docNo, prepareNo)
}

// normalizePayLines: ช่องอ้างหนี้รายบรรทัดใช้เฉพาะเอกสารที่ตัดหนี้รายบรรทัด (ใบจ่ายชำระหนี้); reftype ตามว่ามาจากใบเตรียมจ่ายหรือไม่
func normalizePayLines(dt doctype.DocType, details []ErpDetailItem) {
	lineSettle := subledgerSpecs[dt.Key].lineSettle
	for i := range details {
		line := &details[i]
		if !lineSettle {
			line.DebtDocumentID, line.PayBillNo, line.RefType = "", "", 0
			continue
		}
		line.DebtDocumentID, line.PayBillNo, line.RefType = strings.TrimSpace(line.DebtDocumentID), strings.TrimSpace(line.PayBillNo), 0
		if line.PayBillNo != "" {
			line.RefType = 1
		}
	}
}

// checkPayPrepareLines ตรวจบรรทัดใบจ่ายที่ดึงจากใบเตรียมจ่าย ในธุรกรรมเดียวกับการบันทึก (design §4.5 ข้อ 3) ไม่ว่าใบจ่ายจะลงบัญชีได้หรือไม่:
// ล็อกหัวใบเตรียมจ่าย, ใบต้องยังไม่ยกเลิกและเป็นของเจ้าหนี้เดียวกับใบจ่าย, ยอดรวมต่อ (ใบเตรียมจ่าย, หนี้) ≤ ยอดที่เหลือในใบเตรียมจ่าย
// (ไม่นับบรรทัดของใบจ่ายนี้เอง — ตอนแก้ไขการตัดหนี้เดิมถูกถอนแล้วใน releasePosting)
func checkPayPrepareLines(ctx context.Context, tx *sql.Tx, company string, doc *ErpTransactionDoc, docNo string) error {
	type useKey struct{ prepareNo, debtID string }
	sums, firstLine := map[useKey]decimal.Decimal{}, map[useKey]int{}
	var prepareNos []string
	for i, line := range doc.Details {
		if line.PayBillNo == "" {
			continue
		}
		key := useKey{line.PayBillNo, line.DebtDocumentID}
		if _, seen := firstLine[key]; !seen {
			firstLine[key] = i
		}
		sums[key] = sums[key].Add(money(line.SumAmount))
		prepareNos = append(prepareNos, line.PayBillNo)
	}
	if len(prepareNos) == 0 {
		return nil
	}
	sort.Strings(prepareNos)
	prepareNos = compactStrings(prepareNos)
	rows, err := tx.QueryContext(ctx, `SELECT docno, partner_code, iscancel FROM ap_payment_prepare
		WHERE company=$1 AND docno=ANY($2) ORDER BY docno FOR UPDATE`, company, pq.Array(prepareNos))
	if err != nil {
		return fmt.Errorf("lock payment preparations: %w", err)
	}
	headers := map[string]struct {
		partner   string
		cancelled bool
	}{}
	for rows.Next() {
		var no, partner string
		var cancelled bool
		if err = rows.Scan(&no, &partner, &cancelled); err != nil {
			rows.Close()
			return fmt.Errorf("scan payment preparation: %w", err)
		}
		headers[no] = struct {
			partner   string
			cancelled bool
		}{partner, cancelled}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, no := range prepareNos {
		header, found := headers[no]
		switch {
		case !found:
			return erpBadRequest("ap_prepare_not_found", "paybillno")
		case header.cancelled:
			return erpError("ap_prepare_cancelled_locked", "paybillno")
		case header.partner != strings.TrimSpace(doc.CustCode):
			return erpBadRequest("ap_pay_prepare_creditor_mismatch", "paybillno", no)
		}
	}
	uses, err := loadPrepareLineUse(ctx, tx, company, prepareNos, docNo)
	if err != nil {
		return err
	}
	open := map[useKey]decimal.Decimal{}
	for _, u := range uses {
		open[useKey{u.docNo, u.debtID}] = u.payAmount.Sub(u.paid())
	}
	for key, sum := range sums {
		if sum.GreaterThan(open[key]) {
			i := firstLine[key]
			return erpError("ap_pay_prepare_over_open", fmt.Sprintf("details[%d].sumamount", i), payLineLabel(doc.Details[i]), key.prepareNo)
		}
	}
	return nil
}

func compactStrings(sorted []string) []string {
	out := sorted[:0]
	for i, value := range sorted {
		if i == 0 || value != sorted[i-1] {
			out = append(out, value)
		}
	}
	return out
}

// payLineLabel คือชื่อเอกสารหนี้ที่แสดงในข้อความผิดพลาดของบรรทัด (เลขที่ที่อ้าง ถ้าไม่มีใช้รหัสแถวหนี้)
func payLineLabel(line ErpDetailItem) string {
	if ref := strings.TrimSpace(line.DocRef); ref != "" {
		return ref
	}
	return line.DebtDocumentID
}

// settleLines ตัดหนี้รายบรรทัดของใบจ่ายชำระหนี้ (design §4.5, ตัวตรวจ §4.5.1): บรรทัดที่มี debt_document_id ตัดหนี้แถวนั้นเท่ากับยอดบรรทัด
// ตรวจทุกบรรทัดก่อนเขียน — หนี้ต้องเป็นบัญชีย่อยเดียวกัน คู่ค้า/บัญชีคุม/สาขา/สกุลเงิน (ว่าง = THB) เดียวกับใบจ่าย และเป็นฝั่งหนี้หลัง
// normalize แหล่ง (จอบัญชี side 1, ERP side ตามประเภท); ยอดรวมต่อหนี้ ≤ ยอดคงเหลือ as-of วันที่ใบจ่าย (gl.APOpenItems) และ ≤ ยอดคงเหลือ
// ปัจจุบันที่นับทุกการตัดหนี้ (กันใบจ่ายย้อนวันตัดซ้ำกับใบที่ตัดไปแล้วหลังวันนั้น); ยอดรวมทุกบรรทัด ≤ ยอดของใบจ่าย
func settleLines(ctx context.Context, tx *sql.Tx, scope gl.Scope, spec subledgerSpec, doc *ErpTransactionDoc, docNo, ownID, partner, control, date, journalID string, total decimal.Decimal, now time.Time) error {
	type payLine struct {
		index  int
		amount decimal.Decimal
	}
	var lines []payLine
	perDebt, firstLine := map[string]decimal.Decimal{}, map[string]int{}
	var ids []string
	sum := decimal.Zero
	for i, item := range doc.Details {
		if item.DebtDocumentID == "" {
			continue
		}
		amount := money(item.SumAmount)
		if !amount.IsPositive() {
			e := amountInvalidError()
			e.Field = fmt.Sprintf("details[%d].sumamount", i)
			return e
		}
		if _, seen := firstLine[item.DebtDocumentID]; !seen {
			firstLine[item.DebtDocumentID] = i
			ids = append(ids, item.DebtDocumentID)
		}
		perDebt[item.DebtDocumentID] = perDebt[item.DebtDocumentID].Add(amount)
		sum = sum.Add(amount)
		lines = append(lines, payLine{i, amount})
	}
	if len(lines) == 0 {
		return nil
	}
	if sum.GreaterThan(total) {
		return erpBadRequest("ap_pay_lines_over_total", "totalamount", sum.StringFixed(2), total.StringFixed(2))
	}
	sort.Strings(ids)
	debtNos, err := lockPayDebts(ctx, tx, scope, spec, partner, control, ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, ok := debtNos[id]; !ok {
			i := firstLine[id]
			return erpBadRequest("ap_pay_settle_mismatch", fmt.Sprintf("details[%d].debt_document_id", i), payLineLabel(doc.Details[i]))
		}
	}
	// as-of = ledgerAsOf (เหมือนจอเตรียมจ่าย): นับการตัดหนี้ของใบจ่ายที่ลงวันที่หลังกว่าด้วย
	open, _, err := gl.APOpenItems(ctx, tx, scope, gl.APOpenQuery{Partner: partner, AsOf: ledgerAsOf(date, now), IDs: ids, Limit: len(ids)})
	if err != nil {
		return ledgerFail(err)
	}
	remaining := map[string]decimal.Decimal{}
	for _, item := range open {
		remaining[item.ID] = item.Remaining.Decimal()
	}
	for _, id := range ids {
		current, err := debtOutstanding(ctx, tx, scope.Company, id)
		if err != nil {
			return err
		}
		limit := decimal.Min(remaining[id], current)
		if perDebt[id].GreaterThan(limit) {
			e := settleExceedsOutstandingError(debtNos[id], limit)
			e.Field = fmt.Sprintf("details[%d].sumamount", firstLine[id])
			return e
		}
	}

	prepareLines := map[[2]string]int{}
	if prepareNos := payPrepareNos(doc.Details); len(prepareNos) > 0 {
		uses, err := loadPrepareLineUse(ctx, tx, scope.Company, prepareNos, docNo)
		if err != nil {
			return err
		}
		for _, u := range uses {
			prepareLines[[2]string{u.docNo, u.debtID}] = u.lineNumber
		}
	}
	args := []any{scope.Company, journalID, ownID, date, now, scope.Actor}
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		item := doc.Details[line.index]
		amount := line.amount.StringFixed(2)
		payload := map[string]any{"debt_id": item.DebtDocumentID, "payment_id": ownID, "docno": docNo, "linenumber": item.LineNumber,
			"docref": item.DocRef, "amount": amount}
		if item.PayBillNo != "" {
			payload["prepare_docno"] = item.PayBillNo
			payload["prepare_linenumber"] = prepareLines[[2]string{item.PayBillNo, item.DebtDocumentID}]
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		n := len(args)
		values = append(values, fmt.Sprintf("($1, $%d, $2, $%d, $3, $%d, $4, $%d, $5, $6)", n+1, n+2, n+3, n+4))
		args = append(args, fmt.Sprintf("%s%s-%d", spec.settlePrefix, docNo, item.LineNumber), item.DebtDocumentID, amount, body)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_subledger_settlements (
		company, id, journal_id, debt_id, payment_id, amount, settlement_date, payload, created_at, created_by
	) VALUES `+strings.Join(values, ", "), args...); err != nil {
		return fmt.Errorf("insert subledger settlements: %w", err)
	}
	return nil
}

// lockPayDebts ล็อกแถวหนี้ที่บรรทัดอ้าง (FOR UPDATE เรียงตาม id: การตัดหนี้ใบเดียวกันพร้อมกันรอกัน แล้วอ่านยอดหลังอีกฝั่ง commit)
// และคืนเฉพาะแถวที่ตรงเงื่อนไข §4.5.1 (id → เลขที่เอกสารหนี้); แถวที่ขาดไป = ไม่ใช่หนี้ที่ใบจ่ายนี้ตัดได้
func lockPayDebts(ctx context.Context, tx *sql.Tx, scope gl.Scope, spec subledgerSpec, partner, control string, ids []string) (map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, COALESCE(payload->>'doc_no', payload->>'document_no', id),
			COALESCE(ledger=$3 AND partner_code=$4 AND control_account_code=$5 AND ($6='' OR branch_code=$6)
			AND side = CASE WHEN payload->>'balance_side' IS NOT NULL THEN 1 ELSE $7 END
			AND COALESCE(NULLIF(payload->>'currency_code',''),'THB')='THB', false)
		FROM gl_subledger_documents WHERE company=$1 AND id=ANY($2) ORDER BY id FOR UPDATE`,
		scope.Company, pq.Array(ids), spec.ledger, partner, control, scope.Branch, spec.settleSide)
	if err != nil {
		return nil, fmt.Errorf("lock debts to settle: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, no string
		var matches bool
		if err = rows.Scan(&id, &no, &matches); err != nil {
			return nil, fmt.Errorf("scan debt to settle: %w", err)
		}
		if matches {
			out[id] = no
		}
	}
	return out, rows.Err()
}

// payPrepareNos คือเลขที่ใบเตรียมจ่ายที่บรรทัดอ้าง (ไม่ซ้ำ เรียงคงที่)
func payPrepareNos(details []ErpDetailItem) []string {
	var nos []string
	for _, line := range details {
		if line.PayBillNo != "" {
			nos = append(nos, line.PayBillNo)
		}
	}
	sort.Strings(nos)
	return compactStrings(nos)
}

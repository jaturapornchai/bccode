package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// rowQuerier คือส่วนของ *sql.DB / *sql.Tx ที่การออกเลขที่เอกสารใช้ — ออกเลขในธุรกรรมเดียวกับการบันทึกเพื่อไม่ให้ซ้ำกัน
type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// serialized: เอกสารที่ลงบัญชีหรือเคลื่อนไหวสต็อกต้องคิวต่อกันทั้งบริษัท — สายสต็อกของสินค้าหนึ่งคำนวณจากแถวก่อนหน้า
// และการออกเลขที่ใบสำคัญต้องไม่แย่งกัน; ใช้ล็อกเดียวกับคำสั่งบัญชี (gl.LockCompanyTx) เพื่อไม่ให้เกิด deadlock ข้ามกัน
func serialized(dt doctype.DocType) bool {
	return doctype.PostsToGL(dt.Flag) || movesStock(dt)
}

// checkDocDates คือการตรวจงวดของเอกสาร (Champ IsPeriodClose) ก่อนเขียนทุกครั้ง ทุกประเภทเอกสาร ไม่ใช่เฉพาะที่ลงบัญชี:
// วันที่ของเอกสาร (วันตามปฏิทินของ holding) ต้องอยู่ในปีบัญชีที่เปิดและงวดที่ยังไม่ล็อก; วันที่ว่าง (NULL) ข้าม
func checkDocDates(ctx context.Context, tx *sql.Tx, scope gl.Scope, dates ...sql.NullTime) error {
	checked := make(map[string]bool, len(dates))
	for _, date := range dates {
		if !date.Valid {
			continue
		}
		day := businessDay(date.Time)
		if checked[day] {
			continue
		}
		checked[day] = true
		if _, err := gl.CheckSourceDateOpen(ctx, tx, scope, day); err != nil {
			return err
		}
	}
	return nil
}

// normalizeLines ให้ทุกบรรทัดมีเลขบรรทัดบวกและไม่ซ้ำกัน — เลขบรรทัดเป็นส่วนหนึ่งของกุญแจแถวสต็อก ซ้ำแล้วบันทึกไม่ได้
func normalizeLines(details []ErpDetailItem) {
	used := make(map[int]bool, len(details))
	for i := range details {
		n := details[i].LineNumber
		if n <= 0 {
			n = i + 1
		}
		for used[n] {
			n++
		}
		used[n] = true
		details[i].LineNumber = n
	}
}

// ขอบเขตตัวเลขที่รับ (Rule 10): คอลัมน์แคบสุดที่ค่าเหล่านี้ลงคือ numeric(18,8) จึงต้อง |x| < 1e10
// decimal.UnmarshalJSON รับ "1e2000000000" ได้ และ String/Round/Cmp ของค่านั้นขยายเป็น 10^2e9 (กิน RAM/CPU จนเครื่องล่ม)
// จึงตรวจเลขชี้กำลังและจำนวนบิตของตัวเลขก่อน (ถูก ไม่ขยายค่า) แล้วค่อยเทียบขนาด
const (
	amountMaxExponent = 10
	amountMinExponent = -20 // ทศนิยมไม่เกิน 20 ตำแหน่ง (PostgreSQL ปัดเหลือ scale ของคอลัมน์)
	amountMaxBits     = 128 // ราว 38 หลัก
)

var amountLimit = decimal.New(1, amountMaxExponent)

// maxAmountTokenBytes: ความยาวสูงสุดของตัวเลขหนึ่งช่องในคำขอ (นับเครื่องหมายคำพูด) — ค่าที่บันทึกได้จริงยาวไม่เกินราว 34 ไบต์
// (10 หลัก + ทศนิยม 20 ตำแหน่ง + เครื่องหมาย/จุด/คำพูด); ต้องตัดก่อนแปลงเพราะ decimal แปลงตัวเลขยาว n หลักใช้เวลา O(n^2)
// (ล้านหลัก ≈ 1–3 วินาทีต่อช่อง ทั้งแบบข้อความและตัวเลข JSON) — amountInRange ตรวจได้หลังแปลงเสร็จแล้วเท่านั้น
const maxAmountTokenBytes = 48

// decodeAmount แปลงตัวเลขดิบหนึ่งช่องจากคำขอ: ไม่มีช่อง = คงค่าเดิม (ศูนย์), ยาวเกิน = 400 บอกช่อง, ผิดรูป = error ของ decimal
func decodeAmount(field string, raw json.RawMessage, dst *decimal.Decimal) error {
	if len(raw) == 0 {
		return nil
	}
	if len(raw) > maxAmountTokenBytes {
		return amountOutOfRangeError(field)
	}
	return dst.UnmarshalJSON(raw)
}

// UnmarshalJSON ของรายการ: ช่องตัวเลขรับเป็นข้อความดิบก่อน (ช่องตื้นกว่าบัง Qty/Price/... ของ plain) แล้วตรวจความยาวก่อนแปลง
func (item *ErpDetailItem) UnmarshalJSON(data []byte) error {
	type plain ErpDetailItem
	wire := struct {
		*plain
		Qty            json.RawMessage `json:"qty"`
		Price          json.RawMessage `json:"price"`
		DiscountAmount json.RawMessage `json:"discountamount"`
		SumAmount      json.RawMessage `json:"sumamount"`
	}{plain: (*plain)(item)}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	for _, amount := range []struct {
		field string
		raw   json.RawMessage
		dst   *decimal.Decimal
	}{
		{"qty", wire.Qty, &item.Qty}, {"price", wire.Price, &item.Price},
		{"discountamount", wire.DiscountAmount, &item.DiscountAmount}, {"sumamount", wire.SumAmount, &item.SumAmount},
	} {
		if err := decodeAmount(amount.field, amount.raw, amount.dst); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalJSON ของเอกสาร: ตรวจความยาวตัวเลขหัวเอกสารแบบเดียวกับรายการ และแปลงรายการทีละบรรทัดเพื่อบอกช่องเป็น details[i].<ช่อง>
func (doc *ErpTransactionDoc) UnmarshalJSON(data []byte) error {
	type plain ErpTransactionDoc
	wire := struct {
		*plain
		TotalAmount    json.RawMessage   `json:"totalamount"`
		TotalValue     json.RawMessage   `json:"totalvalue"`
		TotalDiscount  json.RawMessage   `json:"totaldiscount"`
		TotalVatValue  json.RawMessage   `json:"totalvatvalue"`
		TotalBeforeVat json.RawMessage   `json:"totalbeforevat"`
		TotalAfterVat  json.RawMessage   `json:"totalaftervat"`
		VatRate        json.RawMessage   `json:"vatrate"`
		Details        []json.RawMessage `json:"details"`
	}{plain: (*plain)(doc)}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	for _, amount := range []struct {
		field string
		raw   json.RawMessage
		dst   *decimal.Decimal
	}{
		{"totalamount", wire.TotalAmount, &doc.TotalAmount}, {"totalvalue", wire.TotalValue, &doc.TotalValue},
		{"totaldiscount", wire.TotalDiscount, &doc.TotalDiscount}, {"totalvatvalue", wire.TotalVatValue, &doc.TotalVatValue},
		{"totalbeforevat", wire.TotalBeforeVat, &doc.TotalBeforeVat}, {"totalaftervat", wire.TotalAfterVat, &doc.TotalAfterVat},
		{"vatrate", wire.VatRate, &doc.VatRate},
	} {
		if err := decodeAmount(amount.field, amount.raw, amount.dst); err != nil {
			return err
		}
	}
	if wire.Details == nil {
		return nil
	}
	doc.Details = make([]ErpDetailItem, len(wire.Details))
	for i, raw := range wire.Details {
		if err := doc.Details[i].UnmarshalJSON(raw); err != nil {
			if user, ok := gl.AsUserError(err); ok && user.Code == "erp_amount_out_of_range" {
				return amountOutOfRangeError(fmt.Sprintf("details[%d].%s", i, user.Field))
			}
			return err
		}
	}
	return nil
}

func amountInRange(value decimal.Decimal) bool {
	exp := value.Exponent()
	if exp > amountMaxExponent || exp < amountMinExponent || value.Coefficient().BitLen() > amountMaxBits {
		return false
	}
	return value.Abs().LessThan(amountLimit)
}

// validateDocAmounts ตรวจทุกยอดเงิน/จำนวน/ราคา/อัตราภาษีของคำขอก่อนใช้งาน — เกินขอบเขต = 400 บอกช่องที่ผิด
func validateDocAmounts(doc *ErpTransactionDoc) error {
	header := []struct {
		field string
		value decimal.Decimal
	}{
		{"totalamount", doc.TotalAmount}, {"totalvalue", doc.TotalValue}, {"totaldiscount", doc.TotalDiscount},
		{"totalvatvalue", doc.TotalVatValue}, {"totalbeforevat", doc.TotalBeforeVat}, {"totalaftervat", doc.TotalAfterVat},
		{"vatrate", doc.VatRate},
	}
	for _, amount := range header {
		if !amountInRange(amount.value) {
			return amountOutOfRangeError(amount.field)
		}
	}
	for i, item := range doc.Details {
		for _, amount := range []struct {
			field string
			value decimal.Decimal
		}{{"qty", item.Qty}, {"price", item.Price}, {"discountamount", item.DiscountAmount}, {"sumamount", item.SumAmount}} {
			if !amountInRange(amount.value) {
				return amountOutOfRangeError(fmt.Sprintf("details[%d].%s", i, amount.field))
			}
		}
	}
	return nil
}

// SaveDoc บันทึกเอกสารหนึ่งใบเป็นธุรกรรมเดียว: หัว/รายการ/การอ้างอิง + สต็อก + ใบสำคัญบัญชี + บัญชีย่อย
// ทั้งหมดสำเร็จหรือถอยกลับพร้อมกัน; แก้ไข = กลับรายการผลเดิมก่อนแล้วลงใหม่
func (s *Store) SaveDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, doc *ErpTransactionDoc, isEdit bool) (*ErpTransactionDoc, error) {
	if err := validateDocAmounts(doc); err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if serialized(dt) {
		if err = gl.LockCompanyTx(ctx, tx, scope.Company); err != nil {
			return nil, ledgerFail(err)
		}
	}
	company, branch, transflag, actor := scope.Company, scope.Branch, dt.Flag, scope.Actor

	now := time.Now().UTC()
	docDate := now
	if doc.DocDateTime != "" {
		if t, err := time.Parse(time.RFC3339, doc.DocDateTime); err == nil {
			docDate = t
		} else if t, err := time.Parse("2006-01-02", doc.DocDateTime); err == nil {
			docDate = t
		}
	}

	docNo := strings.TrimSpace(doc.DocNo)
	if docNo == "" {
		if isEdit {
			return nil, sql.ErrNoRows
		}
		generated, err := s.NextDocNo(ctx, tx, company, transflag, dt.Prefix)
		if err != nil {
			return nil, fmt.Errorf("generate docno: %w", err)
		}
		docNo = generated
	}
	doc.DocNo = docNo
	normalizeLines(doc.Details)
	normalizePayLines(dt, doc.Details)
	dueDate, err := resolveDueDate(ctx, tx, company, dt, doc, docDate)
	if err != nil {
		return nil, err
	}
	creditDay := 0
	if doc.CreditDay != nil {
		creditDay = *doc.CreditDay
	}

	var docID int
	if isEdit {
		var currentCancel bool
		var currentApproval string
		var storedDate sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT COALESCE(iscancel, false), COALESCE(approval_status, ''), docdatetime
			FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno = $3 AND NOT COALESCE(isdelete, false) FOR UPDATE`,
			company, transflag, docNo).Scan(&currentCancel, &currentApproval, &storedDate)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("update doc %s: %w", docNo, sql.ErrNoRows)
		}
		if err != nil {
			return nil, fmt.Errorf("check doc status: %w", err)
		}
		if currentCancel || currentApproval == "approved" {
			return nil, docLockedError()
		}
		// แก้ไข = ย้ายเอกสารออกจากวันที่เดิมไปวันที่ใหม่ งวดของทั้งสองวันต้องยังเปิดอยู่
		if err = checkDocDates(ctx, tx, scope, storedDate, sql.NullTime{Time: docDate, Valid: true}); err != nil {
			return nil, err
		}

		if doctype.PostsToGL(transflag) {
			if err = s.releasePosting(ctx, tx, scope, dt, docNo, reasonEdit, now); err != nil {
				return nil, err
			}
		}
		updateQuery := `UPDATE doc SET
			docdatetime = $1, custcode = $2, totalamount = $3,
			totalvalue_doc = $4, totaldiscount_doc = $5, totalvatvalue_doc = $6,
			totalbeforevat_doc = $7, totalaftervat_doc = $8,
			custname = $9, description = $10, remark = $11,
			branchid = $12, modifier_code = $13, modified_at = $14,
			creditday = $15, perioddatetime = COALESCE($16::timestamptz, perioddatetime)
			WHERE businesscode = $17 AND transflag = $18 AND docno = $19 AND NOT COALESCE(isdelete, false)
			RETURNING id`
		err = tx.QueryRowContext(ctx, updateQuery,
			docDate, doc.CustCode, doc.TotalAmount,
			doc.TotalValue, doc.TotalDiscount, doc.TotalVatValue,
			doc.TotalBeforeVat, doc.TotalAfterVat,
			doc.CustName, doc.Description, doc.Remark,
			branch, actor, now,
			creditDay, dueDate,
			company, transflag, docNo,
		).Scan(&docID)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("update doc %s: %w", docNo, sql.ErrNoRows)
		}
		if err != nil {
			return nil, fmt.Errorf("update doc: %w", err)
		}
		if _, err = tx.ExecContext(ctx, "DELETE FROM docdetail WHERE businesscode = $1 AND transflag = $2 AND docno = $3", company, transflag, docNo); err != nil {
			return nil, fmt.Errorf("delete doc details: %w", err)
		}
	} else {
		// เลขที่ที่เคยใช้แล้วใช้ซ้ำไม่ได้ แม้เอกสารนั้นถูกลบ — ใบสำคัญและสมุดสต็อกของมันยังเป็นหลักฐานอ้างเลขนี้อยู่
		var taken bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno = $3)`, company, transflag, docNo).Scan(&taken); err != nil {
			return nil, fmt.Errorf("check docno: %w", err)
		}
		if taken {
			return nil, duplicateDocError(docNo)
		}
		if err = checkDocDates(ctx, tx, scope, sql.NullTime{Time: docDate, Valid: true}); err != nil {
			return nil, err
		}
		insertQuery := `INSERT INTO doc (
			businesscode, transflag, docno, custcode, docdatetime,
			totalamount, totalvalue_doc, totaldiscount_doc, totalvatvalue_doc,
			totalbeforevat_doc, totalaftervat_doc, branchid,
			custname, description, remark,
			creator_code, created_at, modifier_code, modified_at, checksum,
			creditday, perioddatetime
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, '',
			$20, $21
		) RETURNING id`
		err = tx.QueryRowContext(ctx, insertQuery,
			company, transflag, docNo, doc.CustCode, docDate,
			doc.TotalAmount, doc.TotalValue, doc.TotalDiscount, doc.TotalVatValue,
			doc.TotalBeforeVat, doc.TotalAfterVat, branch,
			doc.CustName, doc.Description, doc.Remark,
			actor, now, actor, now,
			creditDay, dueDate,
		).Scan(&docID)
		if err != nil {
			return nil, fmt.Errorf("insert doc: %w", err)
		}
	}

	insertDetailQuery := `INSERT INTO docdetail (
		businesscode, transflag, docno, docdatetime, linenumber,
		itemcode, barcode, unitcode, totalqty, price, sumamount,
		description, whcode, locationcode, docref,
		debt_document_id, paybillno, reftype
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15,
		$16, $17, $18
	)`
	docRefGlobal := strings.TrimSpace(doc.DocRef)
	for _, item := range doc.Details {
		desc := item.Description
		if desc == "" {
			desc = item.ItemName
		}
		itemRef := strings.TrimSpace(item.DocRef)
		if itemRef == "" {
			itemRef = docRefGlobal
		}
		if _, err = tx.ExecContext(ctx, insertDetailQuery,
			company, transflag, docNo, docDate, item.LineNumber,
			item.ItemCode, item.Barcode, item.UnitCode, item.Qty, item.Price, item.SumAmount,
			desc, item.WhCode, item.LocationCode, itemRef,
			item.DebtDocumentID, item.PayBillNo, item.RefType,
		); err != nil {
			return nil, fmt.Errorf("insert docdetail line %d: %w", item.LineNumber, err)
		}
	}
	if subledgerSpecs[dt.Key].lineSettle {
		if err = checkPayPrepareLines(ctx, tx, company, doc, docNo); err != nil {
			return nil, err
		}
	}

	// บันทึกความสัมพันธ์การอ้างอิงเอกสาร (Document Reference Linkage)
	if docRefGlobal != "" {
		refTransFlag := doc.DocRefTransFlag
		if refTransFlag == 0 {
			if refType, ok := doctype.ByDocNo(docRefGlobal); ok {
				refTransFlag = refType.Flag
			}
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM docref WHERE businesscode=$1 AND docno=$2`, company, docNo); err != nil {
			return nil, fmt.Errorf("reset docref: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO docref (businesscode, docno, docnotransflag, docnoref, docnoreftransflag)
			VALUES ($1, $2, $3, $4, $5)`, company, docNo, transflag, docRefGlobal, refTransFlag); err != nil {
			return nil, fmt.Errorf("insert docref: %w", err)
		}
		// อัปเดตเอกสารต้นทาง (PO, SO, QT, INV, PUR) ให้เป็นอ้างอิงแล้วและปิดเอกสาร — ยกเว้นเอกสารตัดหนี้รายบรรทัด (ใบจ่ายชำระหนี้):
		// จ่ายบางส่วนต้องไม่ปิดเอกสารหนี้/ใบเตรียมจ่าย; "ปิด" ของเอกสารหนี้ = ยอดคงเหลือในบัญชีย่อยเป็น 0 (design ap-payment-prepare §4.5 ข้อ 5)
		if !subledgerSpecs[dt.Key].lineSettle {
			if _, err = tx.ExecContext(ctx, `UPDATE doc SET isref = true, isclosed = true, modified_at = $1
				WHERE businesscode = $2 AND docno = $3`, now, company, docRefGlobal); err != nil {
				return nil, fmt.Errorf("close referenced doc: %w", err)
			}
		}
	}

	if movesStock(dt) {
		if err = s.replaceStockMovement(ctx, tx, company, dt, doc, docNo, docDate, now); err != nil {
			return nil, err
		}
	}
	posted := postingResult{}
	if doctype.PostsToGL(transflag) {
		if posted, err = s.postDocument(ctx, tx, scope, dt, doc, docNo, docDate, now); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	doc.ID = strconv.Itoa(docID)
	doc.DocDateTime = docDate.Format(time.RFC3339)
	doc.TransFlag = transflag
	doc.BranchCode = branch
	doc.GLStatus, doc.GLDocNo = posted.status, posted.docNo
	return doc, nil
}

// DeleteDoc ลบเอกสาร (ตั้งธง isdelete) พร้อมกลับรายการใบสำคัญ เอาแถวบัญชีย่อยของเอกสารออก และคำนวณสายสต็อกใหม่
// เอกสารที่ถูกตัดหนี้/จัดสรรแล้วลบไม่ได้ (erp_doc_settled); ไม่พบหรือถูกลบไปแล้ว = sql.ErrNoRows
func (s *Store) DeleteDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, docIdOrNo string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if serialized(dt) {
		if err = gl.LockCompanyTx(ctx, tx, scope.Company); err != nil {
			return ledgerFail(err)
		}
	}

	where := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false) AND "
	args := []any{scope.Company, dt.Flag}
	if idNum, err := strconv.Atoi(docIdOrNo); err == nil && idNum > 0 {
		where += "(id = $3 OR docno = $4)"
		args = append(args, idNum, docIdOrNo)
	} else {
		where += "docno = $3"
		args = append(args, docIdOrNo)
	}
	var docNo string
	var currentCancel bool
	var currentApproval string
	var docDate sql.NullTime
	query := fmt.Sprintf("SELECT docno, COALESCE(iscancel, false), COALESCE(approval_status, ''), docdatetime FROM doc %s FOR UPDATE", where)
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&docNo, &currentCancel, &currentApproval, &docDate); err != nil {
		return err
	}
	if currentCancel || currentApproval == "approved" {
		return docLockedError()
	}

	// Reject if any live downstream document exists
	var hasDownstream bool
	checkDownstreamQuery := `SELECT EXISTS(
		SELECT 1 FROM docref r
		JOIN doc d ON d.businesscode = r.businesscode AND d.docno = r.docno AND d.transflag = r.docnotransflag
		WHERE r.businesscode = $1 AND r.docnoref = $2 AND (r.docnoreftransflag = $3 OR r.docnoreftransflag = 0)
		  AND NOT COALESCE(d.isdelete, false) AND NOT COALESCE(d.iscancel, false)
	)`
	if err := tx.QueryRowContext(ctx, checkDownstreamQuery, scope.Company, docNo, dt.Flag).Scan(&hasDownstream); err != nil {
		return fmt.Errorf("check downstream: %w", err)
	}
	if hasDownstream {
		return docHasDownstreamError()
	}
	if err = checkDocDates(ctx, tx, scope, docDate); err != nil {
		return err
	}

	now := time.Now().UTC()
	if doctype.PostsToGL(dt.Flag) {
		if err = s.releasePosting(ctx, tx, scope, dt, docNo, reasonDelete, now); err != nil {
			return err
		}
	}
	if movesStock(dt) {
		if err = s.replaceStockMovement(ctx, tx, scope.Company, dt, nil, docNo, time.Time{}, now); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE doc SET isdelete = true, modifier_code = $1, modified_at = $2
		WHERE businesscode = $3 AND transflag = $4 AND docno = $5`, scope.Actor, now, scope.Company, dt.Flag, docNo); err != nil {
		return fmt.Errorf("delete doc: %w", err)
	}
	return tx.Commit()
}

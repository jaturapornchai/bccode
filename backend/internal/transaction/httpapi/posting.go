package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
)

// เอกสาร ERP ที่ลงบัญชี (doctype.PostingKeys) ลงผ่านเครื่องบัญชีกลาง (gl.ExecuteSourceTx) ในธุรกรรมเดียวกับเอกสาร
// ตามรูปแบบลงบัญชี (mappings) ที่บริษัทตั้งไว้ — ไม่มีรหัสบัญชีฝังในโค้ด และไม่มีทางลัดข้ามกฎของบัญชี
// (ปีบัญชีเปิด งวดไม่ล็อก สมุดรายวันใช้งานอยู่ เดบิต = เครดิต)
// แก้/ลบเอกสาร = กลับรายการใบสำคัญเดิมก่อน แล้วค่อยแก้ไขจึงลงใบใหม่ของรอบถัดไป

const (
	glStatusPosted    = "posted"     // ลงบัญชีแล้ว (glDocNo = เลขที่ใบสำคัญ)
	glStatusNoMapping = "no_mapping" // บริษัทยังไม่ได้ตั้งรูปแบบลงบัญชีของเอกสารประเภทนี้ — บันทึกเอกสารแล้วแต่ยังไม่ลงบัญชี
	glStatusNoAmount  = "no_amount"  // ยอดรวมเป็นศูนย์ ไม่มีรายการให้ลงบัญชี
)

const (
	actionCreateJournal  = "create"
	actionPostJournal    = "post"
	actionReverseJournal = "reverse"
)

const (
	reasonEdit   = "erp_gl_reverse_edit"
	reasonDelete = "erp_gl_reverse_delete"
	reasonCancel = "erp_gl_reverse_cancel"
)

// ledgerError ทำเครื่องหมายว่าความผิดพลาดมาจากเครื่องบัญชี/กฎลงบัญชี เพื่อให้ชั้น HTTP ตอบแบบเดียวกับจอบัญชี
// (ข้อความไทย/อังกฤษ 409 เมื่อเป็นข้อมูลที่ผู้ใช้แก้ได้) แทนที่จะเป็น 500 เปล่าๆ
type ledgerError struct{ err error }

func (e *ledgerError) Error() string { return e.err.Error() }
func (e *ledgerError) Unwrap() error { return e.err }

func ledgerFail(err error) error {
	if err == nil {
		return nil
	}
	return &ledgerError{err}
}

// erpError คือความผิดของเอกสารที่ผู้ใช้แก้ได้: ข้อความไทยมาจากแถว gl_err_<code> ใน languages.tsv แทนค่า {0}, {1}, … ด้วย args
// (ภาษาอื่น glhttp แปลจากแถวเดียวกัน) — ไม่มีข้อความผู้ใช้ฝังในโค้ด
func erpError(code, field string, args ...string) *gl.UserError {
	return &gl.UserError{Code: code, Message: thText("gl_err_"+code, args...), Field: field, Args: args}
}

func amountInvalidError() *gl.UserError {
	return erpError("erp_amount_invalid", "totalamount")
}

// amountOutOfRangeError: ตัวเลขในคำขอเกินขอบเขตที่บันทึกได้ (validateDocAmounts) — field = ชื่อช่อง เช่น details[0].qty
func amountOutOfRangeError(field string) *gl.UserError {
	e := erpError("erp_amount_out_of_range", field)
	e.Status = http.StatusBadRequest
	return e
}

func settledError(docNo string) *gl.UserError {
	return erpError("erp_doc_settled", "", docNo)
}

// settleExceedsOutstandingError: ยอดตัดชำระเกินหนี้คงเหลือ (Champ arfrmbcreceipt1view.cpp:1760 / apfrmappayment1view.cpp:1100
// ปฏิเสธแล้วบอกยอดสูงสุด) — {0} = เลขที่เอกสารหนี้ที่อ้างถึง, {1} = ยอดคงเหลือที่ตัดได้
func settleExceedsOutstandingError(debtDocNo string, outstanding decimal.Decimal) *gl.UserError {
	if outstanding.IsNegative() {
		outstanding = decimal.Zero
	}
	return erpError("erp_settle_exceeds_remaining", "totalamount", debtDocNo, outstanding.StringFixed(2))
}

func duplicateDocError(docNo string) *gl.UserError {
	return erpError("erp_doc_duplicate", "docno", docNo)
}

func docLockedError() *gl.UserError {
	return erpError("erp_doc_locked", "")
}

func docHasDownstreamError() *gl.UserError {
	return erpError("erp_doc_has_downstream", "")
}

func docNotPendingError() *gl.UserError {
	return erpError("erp_doc_not_pending", "")
}

func docNotApprovedError() *gl.UserError {
	return erpError("erp_doc_not_approved", "")
}

func docAlreadyCancelledError() *gl.UserError {
	return erpError("erp_doc_already_cancelled", "")
}

func cancelReasonRequiredError() *gl.UserError {
	return erpError("erp_cancel_reason_required", "reason")
}

// listDateInvalidError: วันที่ค้นหารายการต้องเป็น YYYY-MM-DD (field = ชื่อพารามิเตอร์ fromdate/todate)
func listDateInvalidError(field, value string) *gl.UserError {
	e := erpError("erp_list_date_invalid", field, value)
	e.Status = http.StatusBadRequest
	return e
}

// thText คือข้อความไทยจาก languages.tsv แทนค่า {0}, {1}, … — ใช้กับข้อความที่ลงในใบสำคัญ (ไทยเป็นต้นฉบับของบัญชี)
func thText(key string, args ...string) string {
	return langText(key, "th", args...)
}

// langText คือข้อความของ key ตามภาษาที่ขอ แทนค่า {0}, {1}, … (ไม่มีแถว = คืน key ตามกติกาของ language.Text)
func langText(key, lang string, args ...string) string {
	text := language.Text(key, lang)
	for i, arg := range args {
		text = strings.ReplaceAll(text, "{"+strconv.Itoa(i)+"}", arg)
	}
	return text
}

func docTypeLabel(dt doctype.DocType) string {
	return thText("gl_mapping_doc_" + strings.ReplaceAll(dt.Key, "-", "_"))
}

// glRequestID เป็นรหัสคำขอคงที่ของขั้นหนึ่งของเอกสารหนึ่งรอบ — ส่งซ้ำ (retry) แล้วเครื่องบัญชีคืนผลเดิม ไม่ลงซ้ำ
func glRequestID(company string, flag int, docNo string, revision int, action string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("erp\x00%s\x00%d\x00%s\x00%d\x00%s", company, flag, docNo, revision, action)))
	return "erp-" + hex.EncodeToString(sum[:])[:40]
}

type docAmounts struct{ total, vat, goods decimal.Decimal }

// money ปัดยอดเงินเป็นทศนิยมสองตำแหน่ง ปัดครึ่งออกจากศูนย์ (decimal.Round: 0.005 → 0.01)
func money(value decimal.Decimal) decimal.Decimal {
	return value.Round(2)
}

// postingAmounts คือยอดที่ลงบัญชีของเอกสาร — กติกาปัดเศษ: ปัดที่ระดับเอกสาร (ยอดรวม/ภาษี/ก่อนภาษีของหัวเอกสาร)
// ทีละยอดด้วย money (2 ตำแหน่ง ปัดครึ่งออกจากศูนย์) ไม่ปัดรายบรรทัด; มูลค่าสินค้า = ยอดรวม − ภาษี จึงเดบิต = เครดิตเสมอ
// ไม่มียอดภาษีแต่มียอดก่อนภาษีให้ใช้ส่วนต่างเป็นภาษี
func postingAmounts(doc *ErpTransactionDoc) (docAmounts, error) {
	total, vat, before := money(doc.TotalAmount), money(doc.TotalVatValue), money(doc.TotalBeforeVat)
	if total.IsNegative() || vat.IsNegative() || before.IsNegative() {
		return docAmounts{}, amountInvalidError()
	}
	if vat.IsZero() && before.IsPositive() && total.GreaterThan(before) {
		vat = total.Sub(before)
	}
	if vat.GreaterThan(total) {
		return docAmounts{}, amountInvalidError()
	}
	return docAmounts{total: total, vat: vat, goods: total.Sub(vat)}, nil
}

// mappingLines สร้างรายการบัญชีจากรูปแบบลงบัญชี; รายการที่ยอดเป็นศูนย์ (เช่น ภาษีของเอกสารที่ไม่มีภาษี) ข้ามไป
func mappingLines(rules []gl.MappingRule, amounts docAmounts, description string) []gl.Line {
	lines := make([]gl.Line, 0, len(rules))
	for _, rule := range rules {
		var amount decimal.Decimal
		switch rule.Source {
		case gl.MappingSourceTotal:
			amount = amounts.total
		case gl.MappingSourceGoods:
			amount = amounts.goods
		case gl.MappingSourceVat:
			amount = amounts.vat
		}
		if !amount.IsPositive() {
			continue
		}
		line := gl.Line{AccountCode: rule.AccountCode, Description: description, Debit: "0", Credit: "0"}
		text := gl.Amount(amount.StringFixed(2))
		if rule.Side == gl.MappingSideDebit {
			line.Debit = text
		} else {
			line.Credit = text
		}
		lines = append(lines, line)
	}
	return lines
}

type postingResult struct {
	status string
	docNo  string
}

// postDocument ลงบัญชีเอกสารตามรูปแบบลงบัญชีของประเภทนั้น (สร้างฉบับร่างแล้วผ่านรายการทันที) และลงบัญชีย่อยลูกหนี้/เจ้าหนี้
func (s *Store) postDocument(ctx context.Context, tx *sql.Tx, scope gl.Scope, dt doctype.DocType, doc *ErpTransactionDoc, docNo string, docDate, now time.Time) (postingResult, error) {
	// งวดปิด/ปีบัญชีปิดต้องหยุดก่อนทุกทางออก แม้เอกสารที่ยังไม่มีรูปแบบลงบัญชีหรือยอดเป็นศูนย์
	date := businessDay(docDate)
	year, err := gl.CheckSourceDateOpen(ctx, tx, scope, date)
	if err != nil {
		return postingResult{}, ledgerFail(err)
	}
	mapping, found, err := gl.MappingForSourceDoc(ctx, tx, scope, dt.Key)
	if err != nil {
		return postingResult{}, ledgerFail(err)
	}
	if !found {
		return postingResult{status: glStatusNoMapping}, nil
	}
	amounts, err := postingAmounts(doc)
	if err != nil {
		return postingResult{}, err
	}
	if amounts.total.IsZero() {
		return postingResult{status: glStatusNoAmount}, nil
	}

	var revision int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision), 0) + 1 FROM erp_doc_journals WHERE businesscode=$1 AND transflag=$2 AND docno=$3`, scope.Company, dt.Flag, docNo).Scan(&revision); err != nil {
		return postingResult{}, fmt.Errorf("next journal revision: %w", err)
	}

	description := thText("erp_gl_journal_desc", docTypeLabel(dt), docNo)
	sourceRecordID := fmt.Sprintf("%s:%s:r%d", dt.Key, docNo, revision)
	journal := gl.Journal{
		Date:           date,
		BookCode:       mapping.BookCode,
		FiscalYear:     year,
		Description:    description,
		Reference:      docNo,
		BranchCode:     scope.Branch,
		Kind:           "manual",
		SourceSystem:   "erp",
		SourceRecordID: sourceRecordID,
		Lines:          mappingLines(mapping.Rules, amounts, description),
	}
	created, err := gl.ExecuteSourceTx(ctx, tx, scope, gl.Command{
		Resource:  "journals",
		Action:    actionCreateJournal,
		RequestID: glRequestID(scope.Company, dt.Flag, docNo, revision, actionCreateJournal),
		Journal:   &journal,
	})
	if err != nil {
		return postingResult{}, ledgerFail(err)
	}
	if _, err = gl.ExecuteSourceTx(ctx, tx, scope, gl.Command{
		Resource:  "journals",
		Action:    actionPostJournal,
		ID:        created.ID,
		Version:   created.Version,
		RequestID: glRequestID(scope.Company, dt.Flag, docNo, revision, actionPostJournal),
	}); err != nil {
		return postingResult{}, ledgerFail(err)
	}
	info, ok, err := gl.SourceJournal(ctx, tx, scope, created.ID)
	if err != nil {
		return postingResult{}, ledgerFail(err)
	}
	if !ok {
		return postingResult{}, fmt.Errorf("journal %s of document %s not found after posting", created.ID, docNo)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO erp_doc_journals (businesscode, transflag, docno, revision, journal_id, journal_docno, source_record_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, scope.Company, dt.Flag, docNo, revision, created.ID, info.DocNo, sourceRecordID); err != nil {
		return postingResult{}, fmt.Errorf("link journal to document: %w", err)
	}
	if err = s.writeSubledger(ctx, tx, scope, dt, mapping, doc, docNo, date, created.ID, amounts.total, description, now); err != nil {
		return postingResult{}, err
	}
	return postingResult{status: glStatusPosted, docNo: info.DocNo}, nil
}

// releasePosting เอาผลลงบัญชีเดิมของเอกสารออกก่อนแก้/ลบ: ตรวจว่ายังไม่มีการตัดหนี้ของคนอื่นผูกอยู่ แล้วกลับรายการใบสำคัญ
func (s *Store) releasePosting(ctx context.Context, tx *sql.Tx, scope gl.Scope, dt doctype.DocType, docNo, reasonKey string, now time.Time) error {
	if err := cleanSubledger(ctx, tx, scope, dt, docNo); err != nil {
		return err
	}
	return s.voidDocJournal(ctx, tx, scope, dt, docNo, reasonKey, now)
}

// voidDocJournal กลับรายการใบสำคัญที่ใช้งานอยู่ของเอกสาร (ถ้ามี) แล้วทำเครื่องหมายว่าเลิกใช้ — ใบเดิมคงอยู่เป็นหลักฐาน
func (s *Store) voidDocJournal(ctx context.Context, tx *sql.Tx, scope gl.Scope, dt doctype.DocType, docNo, reasonKey string, now time.Time) error {
	var revision int
	var journalID string
	err := tx.QueryRowContext(ctx, `SELECT revision, journal_id FROM erp_doc_journals WHERE businesscode=$1 AND transflag=$2 AND docno=$3 AND voided_at IS NULL FOR UPDATE`,
		scope.Company, dt.Flag, docNo).Scan(&revision, &journalID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read document journal: %w", err)
	}
	info, found, err := gl.SourceJournal(ctx, tx, scope, journalID)
	if err != nil {
		return ledgerFail(err)
	}
	if !found {
		return fmt.Errorf("journal %s of document %s not found", journalID, docNo)
	}
	switch info.Status {
	case "posted":
		reverseDate, err := gl.SourceReversalDate(ctx, tx, scope, info.Date, businessDay(now))
		if err != nil {
			return ledgerFail(err)
		}
		if _, err = gl.ExecuteSourceTx(ctx, tx, scope, gl.Command{
			Resource:  "journals",
			Action:    actionReverseJournal,
			ID:        journalID,
			Version:   info.Version,
			Date:      reverseDate,
			DocNo:     info.DocNo + "-R",
			Reason:    thText(reasonKey, docNo),
			RequestID: glRequestID(scope.Company, dt.Flag, docNo, revision, actionReverseJournal),
		}); err != nil {
			return ledgerFail(err)
		}
	case "reversed":
		// ผู้ใช้กลับรายการใบนี้ที่จอบัญชีไปก่อนแล้ว ไม่ต้องกลับซ้ำ
	default:
		return fmt.Errorf("journal %s of document %s is %q, expected posted", journalID, docNo, info.Status)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE erp_doc_journals SET voided_at=$1 WHERE businesscode=$2 AND transflag=$3 AND docno=$4 AND revision=$5`,
		now, scope.Company, dt.Flag, docNo, revision); err != nil {
		return fmt.Errorf("void document journal: %w", err)
	}
	return nil
}

// subledgerSpec บอกว่าเอกสารประเภทหนึ่งอยู่ฝั่งไหนของบัญชีย่อย: side 1 = ด้านเดบิต 2 = ด้านเครดิตของบัญชีคุม
// (ลูกหนี้: ขาย = 1 หนี้, รับเงิน/รับคืน = 2; เจ้าหนี้: ซื้อ = 2 หนี้, จ่ายเงิน/ส่งคืน = 1)
type subledgerSpec struct {
	ledger       string // ar | ap
	docPrefix    string
	side         int
	controlSide  string // ด้านของรายการยอดรวมในรูปแบบลงบัญชีที่เป็นบัญชีคุม
	settleSide   int    // ด้านของหนี้ที่เอกสารนี้ตัด; 0 = ไม่ตัดหนี้
	settlePrefix string
	debt         bool // เอกสารนี้เองเป็นหนี้ (มีวันครบกำหนด)
	// excessAsCredit: ยอดเกินหนี้คงค้างไม่ปฏิเสธ — ตัดหนี้เท่าที่ค้าง ส่วนเกินคงเป็นเครดิตค้างของเอกสารนี้ให้ใช้ภายหลัง
	// (รับคืน/ส่งคืนสินค้า: Champ BILLFrmInvCreditNote4.cpp:4336-4350 ปิดการตรวจยอดเกินไว้ และ :1172-1181 ใบลดหนี้เก็บยอดคงเหลือของตัวเอง)
	excessAsCredit bool
}

var subledgerSpecs = map[string]subledgerSpec{
	"sale-invoice":        {ledger: "ar", docPrefix: "AR-", side: 1, controlSide: gl.MappingSideDebit, debt: true},
	"sale-invoice-return": {ledger: "ar", docPrefix: "CN-", side: 2, controlSide: gl.MappingSideCredit, settleSide: 1, settlePrefix: "SET-CN-", excessAsCredit: true},
	"paid":                {ledger: "ar", docPrefix: "RCP-", side: 2, controlSide: gl.MappingSideCredit, settleSide: 1, settlePrefix: "SET-AR-"},
	"purchase":            {ledger: "ap", docPrefix: "AP-", side: 2, controlSide: gl.MappingSideCredit, debt: true},
	"purchase-return":     {ledger: "ap", docPrefix: "PRN-", side: 1, controlSide: gl.MappingSideDebit, settleSide: 2, settlePrefix: "SET-PRN-", excessAsCredit: true},
	"pay":                 {ledger: "ap", docPrefix: "PAY-", side: 1, controlSide: gl.MappingSideDebit, settleSide: 2, settlePrefix: "SET-AP-"},
}

func debtPrefix(ledger string) string {
	if ledger == "ar" {
		return "AR-"
	}
	return "AP-"
}

func controlAccount(rules []gl.MappingRule, side string) string {
	for _, rule := range rules {
		if rule.Source == gl.MappingSourceTotal && rule.Side == side {
			return rule.AccountCode
		}
	}
	return ""
}

// writeSubledger ลงเอกสารลูกหนี้/เจ้าหนี้ของคู่ค้าที่มีในทะเบียนบัญชีย่อย และตัดหนี้เมื่อเอกสารอ้างถึงใบหนี้ที่มีอยู่
// บัญชีคุมมาจากรูปแบบลงบัญชี (ไม่ใช่รหัสที่ฝังในโค้ด); ถ้ารายการนั้นเป็นบัญชีเงินสด/ธนาคาร (ขายสด) ไม่มีหนี้ให้ลง
func (s *Store) writeSubledger(ctx context.Context, tx *sql.Tx, scope gl.Scope, dt doctype.DocType, mapping gl.Master, doc *ErpTransactionDoc, docNo, date, journalID string, total decimal.Decimal, description string, now time.Time) error {
	spec, ok := subledgerSpecs[dt.Key]
	partner := strings.TrimSpace(doc.CustCode)
	if !ok || partner == "" {
		return nil
	}
	control := controlAccount(mapping.Rules, spec.controlSide)
	if control == "" {
		return nil
	}
	cash, err := gl.SourceAccountIsCash(ctx, tx, scope, control)
	if err != nil {
		return ledgerFail(err)
	}
	if cash {
		return nil
	}
	var partnerExists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_partners WHERE company=$1 AND code=$2)`, scope.Company, partner).Scan(&partnerExists); err != nil {
		return fmt.Errorf("read subledger partner: %w", err)
	}
	if !partnerExists {
		return nil
	}

	amount := total.StringFixed(2)
	payload := map[string]any{"doc_no": docNo, "doc_date": date, "journal_id": journalID, "total": amount, "description": description}
	if spec.debt {
		payload["due_date"] = date
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ownID := spec.docPrefix + docNo
	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_subledger_documents (
		company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, $9, $10)`,
		scope.Company, ownID, spec.ledger, partner, scope.Branch, control, amount, spec.side, journalID, body); err != nil {
		return fmt.Errorf("insert subledger document: %w", err)
	}

	ref := strings.TrimSpace(doc.DocRef)
	if spec.settleSide == 0 || ref == "" {
		return nil
	}
	var debtID string
	for _, candidate := range []string{debtPrefix(spec.ledger) + ref, ref} {
		// FOR UPDATE: การตัดหนี้ใบเดียวกันพร้อมกันต้องรอกัน แล้วอ่านยอดคงเหลือหลังอีกฝั่ง commit
		// หนี้ต้องเป็นของคู่ค้าและบัญชีคุมเดียวกัน (และสาขาเดียวกันเมื่อทำงานในสาขา) แบบเดียวกับการตัดหนี้ของจอบัญชี
		// (subledger_allocations.go settlement) — อ้างเลขที่ของคู่ค้าอื่น = ไม่มีหนี้ให้ตัด
		// เอกสารที่สร้างจากจอบัญชี (มี balance_side) ใช้ side 1 = หนี้, 2 = ผลชำระ/ลดหนี้ ทุกบัญชีย่อย จึงตัดได้เฉพาะ side 1
		// (เงื่อนไข debt.Side==1 ของจอบัญชี) — ไม่อย่างนั้นใบจ่ายชำระ (settleSide 2) จะไปตัดเอกสารจ่ายชำระของจอบัญชี
		err = tx.QueryRowContext(ctx, `SELECT id FROM gl_subledger_documents
			WHERE company=$1 AND id=$2 AND ledger=$3 AND partner_code=$5 AND control_account_code=$6 AND ($7 = '' OR branch_code=$7)
				AND side = CASE WHEN payload->>'balance_side' IS NOT NULL THEN 1 ELSE $4 END
			FOR UPDATE`,
			scope.Company, candidate, spec.ledger, spec.settleSide, partner, control, scope.Branch).Scan(&debtID)
		if err == nil {
			break
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("find debt to settle: %w", err)
		}
		debtID = ""
	}
	if debtID == "" {
		return nil
	}
	outstanding, err := debtOutstanding(ctx, tx, scope.Company, debtID)
	if err != nil {
		return err
	}
	settle := total
	if total.GreaterThan(outstanding) {
		if !spec.excessAsCredit {
			return settleExceedsOutstandingError(ref, outstanding)
		}
		if !outstanding.IsPositive() {
			return nil
		}
		settle = outstanding
	}
	settleAmount := settle.StringFixed(2)
	settlement, err := json.Marshal(map[string]any{"debt_id": debtID, "payment_id": ownID, "docno": docNo, "docref": ref, "amount": settleAmount})
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_subledger_settlements (
		company, id, journal_id, debt_id, payment_id, amount, settlement_date, payload, created_at, created_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		scope.Company, spec.settlePrefix+docNo, journalID, debtID, ownID, settleAmount, date, settlement, now, scope.Actor); err != nil {
		return fmt.Errorf("insert subledger settlement: %w", err)
	}
	return nil
}

// debtOutstanding คือยอดหนี้คงเหลือของเอกสาร debtID ณ ตอนนี้ ตามเงื่อนไขการตัดหนี้ของจอบัญชี (subledger_allocations.go settlement):
// ยอดที่ผ่านบัญชีแล้ว − การตัดหนี้ทุกแถวที่ยังไม่ถอน (ไม่ดูสถานะใบสำคัญของการตัดหนี้ จึงเข้มกว่ารายงานหนี้คงค้างที่นับเฉพาะที่ผ่านรายการ); ยอดที่ผ่านบัญชี = ยอดจัดสรรจากใบสำคัญที่ผ่านรายการ (เอกสารที่สร้างจากจอบัญชี มี balance_side)
// หรือยอดเอกสารเมื่อใบสำคัญที่สร้างเอกสารยังผ่านรายการอยู่ (เอกสารจาก ERP) — การจัดสรรเป็นที่มาของหนี้ ไม่ใช่การลดหนี้
func debtOutstanding(ctx context.Context, tx *sql.Tx, company, debtID string) (decimal.Decimal, error) {
	var text string
	if err := tx.QueryRowContext(ctx, `SELECT (
		CASE WHEN d.payload->>'balance_side' IS NOT NULL THEN COALESCE((SELECT SUM(a.amount) FROM gl_subledger_allocations a
				JOIN gl_records r ON r.company=a.company AND r.kind='journals' AND r.id=a.journal_id
				WHERE a.company=d.company AND a.document_id=d.id AND a.reversed_at IS NULL AND r.payload->>'status'='posted'), 0)
			WHEN EXISTS(SELECT 1 FROM gl_records r WHERE r.company=d.company AND r.kind='journals' AND r.id=d.created_journal_id AND r.payload->>'status'='posted') THEN d.amount
			ELSE 0 END
		- COALESCE((SELECT SUM(s.amount) FROM gl_subledger_settlements s WHERE s.company=d.company AND (s.debt_id=d.id OR s.payment_id=d.id) AND s.reversed_at IS NULL), 0)
	)::text FROM gl_subledger_documents d WHERE d.company=$1 AND d.id=$2`, company, debtID).Scan(&text); err != nil {
		return decimal.Zero, fmt.Errorf("read debt outstanding: %w", err)
	}
	outstanding, err := decimal.NewFromString(text)
	if err != nil {
		return decimal.Zero, fmt.Errorf("parse debt outstanding %q: %w", text, err)
	}
	return outstanding, nil
}

// cleanSubledger ลบแถวบัญชีย่อยที่เอกสารนี้สร้างไว้ (เอกสารลูกหนี้/เจ้าหนี้ + การตัดหนี้ของตัวเอง) ก่อนลงใหม่หรือลบเอกสาร
// ถ้ามีการตัดหนี้/จัดสรรของเอกสารอื่นผูกอยู่แล้วจะไม่แตะ — ปล่อยให้แก้ได้จะทำให้ยอดค้างของอีกฝั่งเพี้ยนเงียบๆ
func cleanSubledger(ctx context.Context, tx *sql.Tx, scope gl.Scope, dt doctype.DocType, docNo string) error {
	spec, ok := subledgerSpecs[dt.Key]
	if !ok {
		return nil
	}
	ownID := spec.docPrefix + docNo
	ownSettlement := ""
	if spec.settleSide != 0 {
		ownSettlement = spec.settlePrefix + docNo
	}
	var bound int
	if err := tx.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM gl_subledger_settlements WHERE company=$1 AND (debt_id=$2 OR payment_id=$2) AND id<>$3)
		+ (SELECT count(*) FROM gl_subledger_allocations WHERE company=$1 AND document_id=$2)`, scope.Company, ownID, ownSettlement).Scan(&bound); err != nil {
		return fmt.Errorf("check subledger bindings: %w", err)
	}
	if bound > 0 {
		return settledError(docNo)
	}
	if ownSettlement != "" {
		if _, err := tx.ExecContext(ctx, `DELETE FROM gl_subledger_settlements WHERE company=$1 AND id=$2`, scope.Company, ownSettlement); err != nil {
			return fmt.Errorf("remove subledger settlement: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM gl_subledger_documents WHERE company=$1 AND id=$2`, scope.Company, ownID); err != nil {
		return fmt.Errorf("remove subledger document: %w", err)
	}
	return nil
}

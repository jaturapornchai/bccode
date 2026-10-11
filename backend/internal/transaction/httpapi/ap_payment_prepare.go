package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	gl "smlcloudplatform/internal/generalledger"
)

// ใบเตรียมจ่ายชำระหนี้ (Champ เมนู "บันทึกเตรียมจ่ายชำระหนี้" — design docs/kms/champ/design-2026-10-11/ap-payment-prepare.md §4):
// เอกสารจองยอดหนี้เจ้าหนี้ที่จะจ่าย ไม่ลงบัญชี ไม่แตะ gl_journals; ใบจ่ายชำระหนี้ดึงบรรทัดไปตัดหนี้ (docdetail.paybillno + debt_document_id)
// ยอดที่จ่ายได้/ยอดจองคำนวณสดจาก gl.APOpenItems ตัวเดียว; ยอดที่ใบจ่ายใช้ไปจาก prepareLineUseSQL ตัวเดียว (pay_settlement.go)

// apPrepareAmountPattern: ยอดเงินในคำขอเป็นข้อความทศนิยมไม่เกิน 2 ตำแหน่ง ไม่มีเครื่องหมาย (ตรงกับ numeric(18,2))
var apPrepareAmountPattern = regexp.MustCompile(`^[0-9]{1,16}(\.[0-9]{1,2})?$`)

// apPrepareDocNoPattern / apPrepareRouteNames: เลขที่ใบเตรียมจ่ายเป็น segment ของ path API จึงใช้ชุดอักขระเดียวกับ BFF
// (frontend/src/app/api/ap-payment-prepare/[[...preparePath]]/route.ts) และห้ามชนชื่อ path ของรายการ
var apPrepareDocNoPattern = regexp.MustCompile(`^[\p{L}\p{M}\p{N}_.-]+$`)

var apPrepareRouteNames = map[string]bool{"list": true, "outstanding": true, "creditors": true, ".": true, "..": true}

// billtype ของบรรทัด (คอลัมน์ S ของ Champ, design §4.5.3) กำหนดจากประเภทเอกสารหนี้เท่านั้น — ห้ามดูรหัสบัญชี
const (
	billTypePurchase   = 0
	billTypeBeginning  = 1
	billTypeCreditNote = 2
	billTypeDebitNote  = 3
	billTypeOtherDebt  = 4
)

// erpBillTypes: แถวหนี้จากเอกสาร ERP (หาประเภทจาก prefix ของ id ใน subledgerSpecs)
var erpBillTypes = map[string]int{"purchase": billTypePurchase, "apdebitnote": billTypeDebitNote, "apotherdebt": billTypeOtherDebt}

// glKindBillTypes: แถวหนี้จากจอบัญชีตาม document_kind (5 ยอดยกมา, 3 เพิ่มหนี้, 4 ลดหนี้, 1 ตั้งหนี้ ≈ ตั้งเจ้าหนี้อื่นๆ — คำตัดสิน D12)
var glKindBillTypes = map[int]int{5: billTypeBeginning, 3: billTypeDebitNote, 4: billTypeCreditNote, 1: billTypeOtherDebt}

// billTypeKeys คือ key ชื่อประเภทเอกสารใน languages.tsv ที่จอแสดงแทนตัวเลข (แบบ MyDesc ของ Champ)
var billTypeKeys = map[int]string{
	billTypePurchase:   "goods_receive",
	billTypeBeginning:  "ap_prepare_billtype_beginning",
	billTypeCreditNote: "gl_vat_document_credit_note",
	billTypeDebitNote:  "gl_vat_document_debit_note",
	billTypeOtherDebt:  "ap_prepare_billtype_otherdebt",
}

// apBillType คือ billtype และเครื่องหมายของแถวหนี้ (ลดหนี้ = −1); ไม่รู้ประเภท = เจ้าหนี้อื่นๆ
func apBillType(id string, fromGL bool, kind int) (billType, sign int) {
	billType = billTypeOtherDebt
	if fromGL {
		if bt, ok := glKindBillTypes[kind]; ok {
			billType = bt
		}
	} else {
		for key, spec := range subledgerSpecs {
			if bt, ok := erpBillTypes[key]; ok && strings.HasPrefix(id, spec.docPrefix) {
				billType = bt
			}
		}
	}
	if billType == billTypeCreditNote {
		return billType, -1
	}
	return billType, 1
}

type apPrepareLineInput struct {
	DebtDocumentID string `json:"debt_document_id"`
	PayAmount      string `json:"payamount"`
	Description    string `json:"description"`
}

// apPrepareInput คือคำขอบันทึก/แก้ (การจัดสรร/กลุ่มของบิลยังไม่มีทะเบียน — ไม่รับ เก็บค่าว่าง, design §4.3 ข้อ 10 / D8)
type apPrepareInput struct {
	DocNo          string               `json:"docno"`
	PayDate        string               `json:"paydate"`
	PartnerCode    string               `json:"partner_code"`
	Remark         string               `json:"remark"`
	ProjectCode    string               `json:"projectcode"`
	DepartmentCode string               `json:"departmentcode"`
	DiscountAmount string               `json:"discount_amount"`
	Version        int64                `json:"version"`
	Lines          []apPrepareLineInput `json:"lines"`
}

type apPrepareLine struct {
	LineNumber     int    `json:"linenumber"`
	DebtDocumentID string `json:"debt_document_id"`
	DocumentNo     string `json:"document_no"`
	DocumentDate   string `json:"document_date"`
	DebtAmount     string `json:"debt_amount"`
	BillType       int    `json:"billtype"`
	BillTypeKey    string `json:"billtype_key"`
	Sign           int    `json:"sign"`
	Description    string `json:"description"`
	InvBalance     string `json:"invbalance"`
	PayAmount      string `json:"payamount"`
	UsedAmount     string `json:"used_amount"`
	OpenAmount     string `json:"open_amount"`
}

type apPrepareDoc struct {
	DocNo           string          `json:"docno"`
	PayDate         string          `json:"paydate"`
	PartnerCode     string          `json:"partner_code"`
	PartnerName     string          `json:"partner_name"`
	BranchCode      string          `json:"branch_code"`
	Remark          string          `json:"remark"`
	ProjectCode     string          `json:"projectcode"`
	AllocateCode    string          `json:"allocatecode"`
	BillGroupCode   string          `json:"billgroupcode"`
	DepartmentCode  string          `json:"departmentcode"`
	TotalAmount     string          `json:"total_amount"`
	TotalHomeAmount string          `json:"total_home_amount"`
	DiscountAmount  string          `json:"discount_amount"`
	NetAmount       string          `json:"net_amount"`
	ExchangeProfit  string          `json:"exchange_profit"`
	BillStatus      int             `json:"billstatus"`
	IsCancel        bool            `json:"iscancel"`
	CancelReason    string          `json:"cancel_reason"`
	Version         int64           `json:"version"`
	Lines           []apPrepareLine `json:"lines"`
}

type apPrepareListRow struct {
	DocNo          string `json:"docno"`
	PayDate        string `json:"paydate"`
	PartnerCode    string `json:"partner_code"`
	PartnerName    string `json:"partner_name"`
	TotalAmount    string `json:"total_amount"`
	DiscountAmount string `json:"discount_amount"`
	NetAmount      string `json:"net_amount"`
	BillStatus     int    `json:"billstatus"`
	IsCancel       bool   `json:"iscancel"`
	Version        int64  `json:"version"`
}

type apOutstandingRow struct {
	DebtDocumentID  string `json:"debt_document_id"`
	DocumentNo      string `json:"document_no"`
	DocumentDate    string `json:"document_date"`
	DueDate         string `json:"due_date"`
	BillType        int    `json:"billtype"`
	BillTypeKey     string `json:"billtype_key"`
	Sign            int    `json:"sign"`
	Description     string `json:"description"`
	DebtAmount      string `json:"debt_amount"`
	RemainingAmount string `json:"remaining_amount"`
	ReservedAmount  string `json:"reserved_amount"`
	AvailableAmount string `json:"available_amount"`
}

type apPrepareEvent struct {
	EventNo    int64           `json:"event_no"`
	Action     string          `json:"action"`
	Actor      string          `json:"actor"`
	OccurredAt string          `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

// apPrepareFilter คือเงื่อนไขรายการใบเตรียมจ่าย: วันที่ YYYY-MM-DD รวมวันต้นและวันปลาย, status open|paid|cancelled (ค่าว่าง = ทั้งหมด)
type apPrepareFilter struct {
	Q, Partner, FromDate, ToDate, Status string
	Limit, Offset                        int
}

// ListPrepares คืนรายการใบเตรียมจ่ายของบริษัท + จำนวนทั้งหมด; billstatus 1 = ทุกบรรทัดถูกใบจ่ายใช้ครบแล้ว (design §4.2)
func (s *Store) ListPrepares(ctx context.Context, db *sql.DB, company string, filter apPrepareFilter) ([]apPrepareListRow, int, error) {
	for field, value := range map[string]string{"fromdate": filter.FromDate, "todate": filter.ToDate} {
		if value == "" {
			continue
		}
		if _, err := dayStart(value); err != nil {
			return nil, 0, listDateInvalidError(field, value)
		}
	}
	args := append(prepareUseArgs(company, ""), gl.SearchPattern(filter.Q), strings.TrimSpace(filter.Partner), filter.FromDate, filter.ToDate, filter.Status, filter.Limit, filter.Offset)
	rows, err := db.QueryContext(ctx, `WITH heads AS (
		SELECT p.docno, p.paydate, p.partner_code, p.total_amount, p.discount_amount, p.net_amount, p.iscancel, p.version,
			COALESCE((SELECT bool_and(u.settled + u.pending >= l.payamount) FROM ap_payment_prepare_lines l
				CROSS JOIN LATERAL (SELECT `+prepareLineUseSQL+`) u WHERE l.company=p.company AND l.docno=p.docno), false) AS paid
		FROM ap_payment_prepare p
		WHERE p.company=$1 AND ($6::text='' OR p.docno ILIKE $6 OR p.partner_code ILIKE $6 OR p.remark ILIKE $6)
			AND ($7::text='' OR p.partner_code=$7)
			AND p.paydate >= COALESCE(NULLIF($8::text,'')::date, p.paydate) AND p.paydate <= COALESCE(NULLIF($9::text,'')::date, p.paydate)
	)
	SELECT h.docno, h.paydate::text, h.partner_code, COALESCE(sp.payload->>'title_name',''), COALESCE(sp.payload->>'name_th',''),
		h.total_amount::text, h.discount_amount::text, h.net_amount::text, h.paid, h.iscancel, h.version, COUNT(*) OVER()
	FROM heads h LEFT JOIN gl_subledger_partners sp ON sp.company=$1 AND sp.code=h.partner_code
	WHERE $10::text='' OR ($10::text='cancelled' AND h.iscancel) OR ($10::text='paid' AND NOT h.iscancel AND h.paid)
		OR ($10::text='open' AND NOT h.iscancel AND NOT h.paid)
	ORDER BY h.paydate DESC, h.docno DESC LIMIT $11 OFFSET $12`, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list payment preparations: %w", err)
	}
	defer rows.Close()
	out, total := []apPrepareListRow{}, 0
	for rows.Next() {
		var row apPrepareListRow
		var title, name string
		var paid bool
		if err = rows.Scan(&row.DocNo, &row.PayDate, &row.PartnerCode, &title, &name, &row.TotalAmount, &row.DiscountAmount, &row.NetAmount,
			&paid, &row.IsCancel, &row.Version, &total); err != nil {
			return nil, 0, fmt.Errorf("scan payment preparation: %w", err)
		}
		row.PartnerName = gl.PartnerFullName(title, name)
		if paid {
			row.BillStatus = 1
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

// PrepareOutstanding คือหนี้ค้างจ่ายของเจ้าหนี้ที่เตรียมจ่ายได้ ด้วยกติกาเดียวกับการบันทึก: as-of = asOf หรือวันนี้ที่หลังกว่า,
// แก้ใบเดิม (excludeDocNo) = ไม่นับยอดจองของใบนั้นและอ่านหนี้ในสาขาของใบนั้น; ยอดเงินส่งออกเป็นทศนิยม 2 ตำแหน่ง
func (s *Store) PrepareOutstanding(ctx context.Context, db *sql.DB, scope gl.Scope, partner, asOf, q, excludeDocNo string, limit, offset int) ([]apOutstandingRow, int, error) {
	if strings.TrimSpace(partner) == "" {
		return nil, 0, erpBadRequest("ap_prepare_creditor_required", "partner_code")
	}
	if _, err := dayStart(asOf); err != nil {
		return nil, 0, listDateInvalidError("todate", asOf)
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	if excludeDocNo != "" {
		if scope.Branch, err = prepareBranch(ctx, tx, scope.Company, excludeDocNo); err != nil {
			return nil, 0, err
		}
	}
	items, total, err := gl.APOpenItems(ctx, tx, scope, gl.APOpenQuery{Partner: partner, AsOf: ledgerAsOf(asOf, time.Now()), Q: q,
		ExcludePrepare: excludeDocNo, Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	out := make([]apOutstandingRow, 0, len(items))
	for _, item := range items {
		billType, sign := apBillType(item.ID, item.FromGL, item.Kind)
		out = append(out, apOutstandingRow{
			DebtDocumentID: item.ID, DocumentNo: item.DocumentNo, DocumentDate: item.DocumentDate, DueDate: item.DueDate,
			BillType: billType, BillTypeKey: billTypeKeys[billType], Sign: sign, Description: item.Description,
			DebtAmount: moneyText(item.Amount), RemainingAmount: moneyText(item.Remaining), ReservedAmount: moneyText(item.Reserved),
			AvailableAmount: moneyText(item.Available),
		})
	}
	return out, total, tx.Commit()
}

// moneyText คือยอดเงินจากบัญชีย่อย (NUMERIC สเกล 8) ที่ส่งออก API: ปัดครั้งเดียวเป็น 2 ตำแหน่งด้วย money ให้ตรงกับยอดที่บันทึกรับได้
func moneyText(a gl.Amount) string {
	return money(a.Decimal()).StringFixed(2)
}

// ledgerAsOf คือวันที่ as-of ของยอดหนี้ที่ใช้ร่วมกันทั้งจอเลือกหนี้ การบันทึกใบเตรียมจ่าย และการตัดหนี้ของใบจ่าย: วันที่เอกสาร
// หรือวันนี้ที่หลังกว่า — as-of นับเวลาผ่านบัญชีด้วย เอกสารที่ลงวันที่ย้อนหลังจึงยังเห็นหนี้ที่เพิ่งผ่านบัญชีวันนี้
func ledgerAsOf(day string, now time.Time) string {
	if today := businessDay(now); today > day {
		return today
	}
	return day
}

// prepareBranch คือสาขาของใบเตรียมจ่าย — หนี้ของใบอ่านในสาขาของใบเสมอ ไม่ขึ้นกับสาขาของผู้แก้ (Champ ไม่มีสาขาในใบเตรียมจ่าย)
func prepareBranch(ctx context.Context, tx *sql.Tx, company, docNo string) (string, error) {
	var branch string
	err := tx.QueryRowContext(ctx, `SELECT branch_code FROM ap_payment_prepare WHERE company=$1 AND docno=$2`, company, docNo).Scan(&branch)
	if errors.Is(err, sql.ErrNoRows) {
		return "", prepareNotFound()
	}
	if err != nil {
		return "", fmt.Errorf("read payment preparation branch: %w", err)
	}
	return branch, nil
}

// GetPrepare อ่านใบเตรียมจ่ายพร้อมยอดที่ใบจ่ายใช้ไปต่อบรรทัด; ไม่พบ = nil, nil
func (s *Store) GetPrepare(ctx context.Context, db *sql.DB, company, docNo string) (*apPrepareDoc, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	doc := apPrepareDoc{}
	err = tx.QueryRowContext(ctx, `SELECT docno, paydate::text, partner_code, branch_code, remark, projectcode, allocatecode, billgroupcode,
		departmentcode, total_amount::text, total_home_amount::text, discount_amount::text, net_amount::text, exchange_profit::text,
		iscancel, cancel_reason, version FROM ap_payment_prepare WHERE company=$1 AND docno=$2`, company, docNo).Scan(
		&doc.DocNo, &doc.PayDate, &doc.PartnerCode, &doc.BranchCode, &doc.Remark, &doc.ProjectCode, &doc.AllocateCode, &doc.BillGroupCode,
		&doc.DepartmentCode, &doc.TotalAmount, &doc.TotalHomeAmount, &doc.DiscountAmount, &doc.NetAmount, &doc.ExchangeProfit,
		&doc.IsCancel, &doc.CancelReason, &doc.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read payment preparation: %w", err)
	}
	if partner, found, err := lookupPartner(ctx, tx, company, doc.PartnerCode); err != nil {
		return nil, err
	} else if found {
		doc.PartnerName = partner.Name
	}
	uses, err := loadPrepareLineUse(ctx, tx, company, []string{docNo}, "")
	if err != nil {
		return nil, err
	}
	doc.Lines = make([]apPrepareLine, 0, len(uses))
	for _, u := range uses {
		doc.Lines = append(doc.Lines, apPrepareLine{
			LineNumber: u.lineNumber, DebtDocumentID: u.debtID, DocumentNo: u.documentNo, DocumentDate: u.documentDate,
			DebtAmount: u.debtAmount, BillType: u.billType, BillTypeKey: billTypeKeys[u.billType], Sign: u.sign, Description: u.description,
			InvBalance: u.balance, PayAmount: u.payAmount.StringFixed(2), UsedAmount: u.paid().StringFixed(2),
			OpenAmount: u.payAmount.Sub(u.paid()).StringFixed(2),
		})
	}
	if fullyPaid(uses) {
		doc.BillStatus = 1
	}
	return &doc, tx.Commit()
}

// fullyPaid: ทุกบรรทัดถูกใบจ่ายใช้ครบยอดเตรียมจ่ายแล้ว (Champ BillStatus = 1 เมื่อจ่ายหมด bcprocess.cpp:490-515)
func fullyPaid(uses []prepareLineUse) bool {
	for _, u := range uses {
		if u.paid().LessThan(u.payAmount) {
			return false
		}
	}
	return len(uses) > 0
}

// preparedLine คือบรรทัดที่ผ่านการตรวจแล้วพร้อมบันทึก
type preparedLine struct {
	apPrepareLine
	amount decimal.Decimal
}

// SavePrepare บันทึกใบใหม่ (editNo = "") หรือแก้ใบเดิม ในธุรกรรมเดียว (design §4.3):
// ล็อกบริษัท → ตรวจหัว/งวด → ล็อกแถวหนี้ → ยอดแต่ละบรรทัดต้อง ≥ ยอดที่ใบจ่ายใช้ไปแล้ว (แก้ได้จนจ่ายครบ — คำตัดสิน D11)
// และ ≤ ยอดที่ตัดแล้ว + ยอดที่ยังเตรียมจ่ายได้ (gl.APOpenItems) → ยอดรวมคำนวณที่ backend → เขียนหัว/บรรทัด/เหตุการณ์
func (s *Store) SavePrepare(ctx context.Context, db *sql.DB, scope gl.Scope, in apPrepareInput, editNo string) (*apPrepareDoc, error) {
	docNo := strings.TrimSpace(in.DocNo)
	if editNo != "" {
		docNo = editNo
	}
	payDay, partnerCode, discount, amounts, err := validatePrepareInput(in, docNo)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	company := scope.Company
	if err = gl.LockCompanyTx(ctx, tx, company); err != nil {
		return nil, err
	}
	if err = checkPrepareMasters(ctx, tx, company, partnerCode, in); err != nil {
		return nil, err
	}
	oldPayDay := sql.NullTime{}
	lineScope := scope // แก้ใบเดิม: อ่านหนี้ในสาขาของใบ
	if editNo == "" {
		var exists bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ap_payment_prepare WHERE company=$1 AND docno=$2)`, company, docNo).Scan(&exists); err != nil {
			return nil, fmt.Errorf("check payment preparation number: %w", err)
		}
		if exists {
			return nil, erpError("ap_prepare_doc_exists", "docno")
		}
	} else if oldPayDay, lineScope.Branch, err = lockPrepareHeader(ctx, tx, company, docNo, in.Version); err != nil {
		return nil, err
	}
	if err = checkDocDates(ctx, tx, scope, oldPayDay, sql.NullTime{Time: payDay, Valid: true}); err != nil {
		return nil, err
	}
	lines, err := checkPrepareLines(ctx, tx, lineScope, docNo, partnerCode, businessDay(payDay), in.Lines, amounts, editNo != "")
	if err != nil {
		return nil, err
	}
	total := decimal.Zero
	for _, line := range lines {
		total = total.Add(line.amount.Mul(decimal.NewFromInt(int64(line.Sign))))
	}
	if discount.IsPositive() && discount.GreaterThan(total) {
		return nil, erpBadRequest("ap_prepare_discount_invalid", "discount_amount")
	}
	if err = writePrepare(ctx, tx, scope, editNo == "", docNo, businessDay(payDay), partnerCode, in, total, discount, lines); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit payment preparation: %w", err)
	}
	return s.GetPrepare(ctx, db, company, docNo)
}

// validatePrepareInput ตรวจรูปแบบคำขอที่ไม่ต้องอ่านฐานข้อมูล: เลขที่, วันที่, เจ้าหนี้, ส่วนลด, บรรทัดไม่ว่าง/ไม่ซ้ำ/ยอดถูกรูปแบบ
func validatePrepareInput(in apPrepareInput, docNo string) (time.Time, string, decimal.Decimal, []decimal.Decimal, error) {
	fail := func(err error) (time.Time, string, decimal.Decimal, []decimal.Decimal, error) {
		return time.Time{}, "", decimal.Zero, nil, err
	}
	if docNo == "" {
		return fail(erpBadRequest("ap_prepare_docno_required", "docno"))
	}
	if !apPrepareDocNoPattern.MatchString(docNo) || apPrepareRouteNames[docNo] {
		return fail(erpBadRequest("ap_prepare_docno_invalid", "docno", docNo))
	}
	payDay, err := dayStart(strings.TrimSpace(in.PayDate))
	if err != nil {
		return fail(erpBadRequest("source_date_invalid", "paydate"))
	}
	partnerCode := strings.TrimSpace(in.PartnerCode)
	if partnerCode == "" {
		return fail(erpBadRequest("ap_prepare_creditor_required", "partner_code"))
	}
	discount := decimal.Zero
	if raw := strings.TrimSpace(in.DiscountAmount); raw != "" {
		if !apPrepareAmountPattern.MatchString(raw) {
			return fail(erpBadRequest("ap_prepare_discount_invalid", "discount_amount"))
		}
		discount = decimal.RequireFromString(raw)
	}
	if len(in.Lines) == 0 {
		return fail(erpBadRequest("ap_prepare_no_lines", "lines"))
	}
	seen := map[string]bool{}
	amounts := make([]decimal.Decimal, len(in.Lines))
	for i, line := range in.Lines {
		id := strings.TrimSpace(line.DebtDocumentID)
		if id == "" {
			return fail(erpBadRequest("ap_prepare_debt_not_found", fmt.Sprintf("lines[%d].debt_document_id", i), id))
		}
		if seen[id] {
			return fail(erpBadRequest("ap_prepare_line_duplicate", fmt.Sprintf("lines[%d].debt_document_id", i), id))
		}
		seen[id] = true
		raw := strings.TrimSpace(line.PayAmount)
		if !apPrepareAmountPattern.MatchString(raw) {
			return fail(erpBadRequest("ap_prepare_amount_invalid", fmt.Sprintf("lines[%d].payamount", i)))
		}
		if amounts[i] = decimal.RequireFromString(raw); !amounts[i].IsPositive() {
			return fail(erpBadRequest("ap_prepare_amount_invalid", fmt.Sprintf("lines[%d].payamount", i)))
		}
	}
	return payDay, partnerCode, discount, amounts, nil
}

// checkPrepareMasters: เจ้าหนี้ต้องมีในทะเบียน เป็นเจ้าหนี้ และใช้งานอยู่; โครงการ/แผนก (ถ้ากรอก) ต้องมีในทะเบียน (design §4.3 ข้อ 2, 9)
func checkPrepareMasters(ctx context.Context, tx *sql.Tx, company, partnerCode string, in apPrepareInput) error {
	partner, found, err := lookupPartner(ctx, tx, company, partnerCode)
	if err != nil {
		return err
	}
	if !found || !partner.IsSupplier || !partner.IsActive {
		return erpBadRequest("ap_prepare_creditor_not_found", "partner_code")
	}
	for _, dim := range []struct{ kind, field, code string }{
		{gl.ProjectKind, "projectcode", in.ProjectCode}, {gl.DepartmentKind, "departmentcode", in.DepartmentCode},
	} {
		code := strings.TrimSpace(dim.code)
		if code == "" {
			continue
		}
		exists, err := gl.DimensionExists(ctx, tx, company, dim.kind, code)
		if err != nil {
			return err
		}
		if !exists {
			return erpBadRequest("ap_prepare_dimension_not_found", dim.field)
		}
	}
	return nil
}

// lockPrepareHeader ล็อกหัวใบเตรียมจ่ายที่จะแก้: ต้องมีอยู่ ยังไม่ยกเลิก และ version ตรง (optimistic lock)
// คืนวันที่จ่ายเดิมไว้ตรวจงวด และสาขาของใบ
func lockPrepareHeader(ctx context.Context, tx *sql.Tx, company, docNo string, version int64) (sql.NullTime, string, error) {
	var payDay, branch string
	var cancelled bool
	var current int64
	err := tx.QueryRowContext(ctx, `SELECT paydate::text, branch_code, iscancel, version FROM ap_payment_prepare WHERE company=$1 AND docno=$2 FOR UPDATE`,
		company, docNo).Scan(&payDay, &branch, &cancelled, &current)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullTime{}, "", prepareNotFound()
	}
	if err != nil {
		return sql.NullTime{}, "", fmt.Errorf("lock payment preparation: %w", err)
	}
	if cancelled {
		return sql.NullTime{}, "", erpError("ap_prepare_cancelled_locked", "")
	}
	if version != current {
		return sql.NullTime{}, "", erpError("ap_prepare_version_conflict", "version")
	}
	day, err := dayStart(payDay)
	if err != nil {
		return sql.NullTime{}, "", fmt.Errorf("parse payment preparation date %q: %w", payDay, err)
	}
	return sql.NullTime{Time: day, Valid: true}, branch, nil
}

// checkPrepareLines ตรวจยอดทุกบรรทัดกับหนี้ค้างจ่ายและยอดที่ใบจ่ายใช้ไปแล้ว แล้วคืนบรรทัดพร้อม snapshot ของเอกสารหนี้
// as-of = ledgerAsOf(วันที่จ่าย) — การตัดหนี้ที่เกิดหลังวันที่จ่ายก็ลดยอดที่จ่ายได้แล้ว; scope.Branch = สาขาของใบ
func checkPrepareLines(ctx context.Context, tx *sql.Tx, scope gl.Scope, docNo, partnerCode, payDay string, inputs []apPrepareLineInput, amounts []decimal.Decimal, isEdit bool) ([]preparedLine, error) {
	ids := make([]string, len(inputs))
	for i, line := range inputs {
		ids[i] = strings.TrimSpace(line.DebtDocumentID)
	}
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	debts, err := lockPrepareDebts(ctx, tx, scope.Company, partnerCode, sorted)
	if err != nil {
		return nil, err
	}
	used := map[string]prepareLineUse{}
	if isEdit {
		uses, err := loadPrepareLineUse(ctx, tx, scope.Company, []string{docNo}, "")
		if err != nil {
			return nil, err
		}
		if fullyPaid(uses) {
			return nil, erpError("ap_prepare_paid_locked", "")
		}
		kept := map[string]bool{}
		for _, id := range ids {
			kept[id] = true
		}
		for _, u := range uses {
			used[u.debtID] = u
			if !kept[u.debtID] && u.paid().IsPositive() {
				return nil, erpError("ap_prepare_below_paid", "lines", u.documentNo, u.paid().StringFixed(2))
			}
		}
	}
	open, _, err := gl.APOpenItems(ctx, tx, scope, gl.APOpenQuery{Partner: partnerCode, AsOf: ledgerAsOf(payDay, time.Now()), ExcludePrepare: docNo,
		IDs: sorted, Limit: len(sorted)})
	if err != nil {
		return nil, err
	}
	openByID := map[string]gl.APOpenItem{}
	for _, item := range open {
		openByID[item.ID] = item
	}
	lines := make([]preparedLine, 0, len(inputs))
	for i, input := range inputs {
		id, amount := ids[i], amounts[i]
		old, wasLine := used[id]
		item, found := openByID[id]
		label, isDebt := debts[id]
		// หนี้ที่ไม่มียอดคงเหลือ (ยังไม่ผ่านบัญชีหรือตัดครบแล้ว) เพิ่มเป็นบรรทัดใหม่ไม่ได้; บรรทัดเดิมคงไว้ได้ภายในเพดาน
		if !isDebt || !found || (!wasLine && !item.Remaining.Decimal().IsPositive()) {
			return nil, erpBadRequest("ap_prepare_debt_not_found", fmt.Sprintf("lines[%d].debt_document_id", i), id)
		}
		// เพดาน = ยอดหนี้ที่ผ่านบัญชี − การตัดหนี้และยอดจองของทุกอย่างที่ไม่ใช่ใบนี้ (as-of เดียวกันทั้งหมด)
		// = ยอดที่ยังเตรียมจ่ายได้ + ยอดที่ใบนี้ตัดหนี้แล้ว; ข้อความแสดงไม่ต่ำกว่า 0 แต่เทียบกับค่าจริง
		limit := money(item.Available.Decimal().Add(item.PrepareUsed.Decimal()))
		field := fmt.Sprintf("lines[%d].payamount", i)
		if amount.GreaterThan(limit) {
			return nil, erpError("ap_prepare_over_balance", field, label, decimal.Max(limit, decimal.Zero).StringFixed(2))
		}
		if amount.LessThan(old.paid()) {
			return nil, erpError("ap_prepare_below_paid", field, label, old.paid().StringFixed(2))
		}
		line := preparedLine{amount: amount, apPrepareLine: apPrepareLine{LineNumber: i + 1, DebtDocumentID: id, PayAmount: amount.StringFixed(2),
			Description: strings.TrimSpace(input.Description)}}
		line.BillType, line.Sign = apBillType(item.ID, item.FromGL, item.Kind)
		line.DocumentNo, line.DocumentDate, line.DebtAmount, line.InvBalance = item.DocumentNo, item.DocumentDate, moneyText(item.Amount), moneyText(item.Remaining)
		if line.Description == "" {
			line.Description = item.Description
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// lockPrepareDebts ล็อกแถวหนี้ที่เลือก (FOR UPDATE เรียงตาม id: สองคนเตรียมจ่ายหนี้ใบเดียวกันทำต่อกันทีละราย — design §4.2)
// คืนเฉพาะแถวที่เป็นบัญชีย่อยเจ้าหนี้ของเจ้าหนี้รายนี้ (id → เลขที่เอกสารหนี้)
func lockPrepareDebts(ctx context.Context, tx *sql.Tx, company, partnerCode string, ids []string) (map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, COALESCE(payload->>'doc_no', payload->>'document_no', id), COALESCE(ledger='ap' AND partner_code=$3, false)
		FROM gl_subledger_documents WHERE company=$1 AND id=ANY($2) ORDER BY id FOR UPDATE`, company, pq.Array(ids), partnerCode)
	if err != nil {
		return nil, fmt.Errorf("lock debts to prepare: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, no string
		var matches bool
		if err = rows.Scan(&id, &no, &matches); err != nil {
			return nil, fmt.Errorf("scan debt to prepare: %w", err)
		}
		if matches {
			out[id] = no
		}
	}
	return out, rows.Err()
}

// writePrepare เขียนหัว (ใหม่หรือแก้ + version+1) แทนบรรทัดทั้งชุด และบันทึกเหตุการณ์ลง ap_payment_prepare_events
func writePrepare(ctx context.Context, tx *sql.Tx, scope gl.Scope, isNew bool, docNo, payDay, partnerCode string, in apPrepareInput, total, discount decimal.Decimal, lines []preparedLine) error {
	now := time.Now().UTC()
	remark, project, department := strings.TrimSpace(in.Remark), strings.TrimSpace(in.ProjectCode), strings.TrimSpace(in.DepartmentCode)
	totalText, discountText, netText := total.StringFixed(2), discount.StringFixed(2), total.Sub(discount).StringFixed(2)
	action := "update"
	var err error
	if isNew {
		action = "create"
		_, err = tx.ExecContext(ctx, `INSERT INTO ap_payment_prepare (company, docno, paydate, partner_code, branch_code, remark, projectcode, departmentcode,
			total_amount, total_home_amount, discount_amount, net_amount, created_at, created_by, modified_at, modified_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, $10, $11, $12, $13, $12, $13)`,
			scope.Company, docNo, payDay, partnerCode, scope.Branch, remark, project, department, totalText, discountText, netText, now, scope.Actor)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE ap_payment_prepare SET paydate=$3, partner_code=$4, remark=$5, projectcode=$6, departmentcode=$7,
			total_amount=$8, total_home_amount=$8, discount_amount=$9, net_amount=$10, version=version+1, modified_at=$11, modified_by=$12
			WHERE company=$1 AND docno=$2`,
			scope.Company, docNo, payDay, partnerCode, remark, project, department, totalText, discountText, netText, now, scope.Actor)
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM ap_payment_prepare_lines WHERE company=$1 AND docno=$2`, scope.Company, docNo)
		}
	}
	if err != nil {
		return fmt.Errorf("write payment preparation: %w", err)
	}
	args := []any{scope.Company, docNo}
	values := make([]string, 0, len(lines))
	eventLines := make([]map[string]string, 0, len(lines))
	for _, line := range lines {
		n := len(args)
		values = append(values, fmt.Sprintf("($1, $2, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8, n+9, n+10))
		args = append(args, line.LineNumber, line.DebtDocumentID, line.DocumentNo, line.DocumentDate, line.DebtAmount, line.BillType, line.Sign,
			line.Description, line.InvBalance, line.PayAmount)
		eventLines = append(eventLines, map[string]string{"debt_document_id": line.DebtDocumentID, "payamount": line.PayAmount})
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO ap_payment_prepare_lines (company, docno, linenumber, debt_document_id, document_no, document_date,
		debt_amount, billtype, sign, description, invbalance, payamount) VALUES `+strings.Join(values, ", "), args...); err != nil {
		return fmt.Errorf("write payment preparation lines: %w", err)
	}
	return writePrepareEvent(ctx, tx, scope, docNo, action, now, map[string]any{
		"paydate": payDay, "partner_code": partnerCode, "remark": remark, "projectcode": project, "departmentcode": department,
		"total_amount": totalText, "discount_amount": discountText, "net_amount": netText, "lines": eventLines,
	})
}

func writePrepareEvent(ctx context.Context, tx *sql.Tx, scope gl.Scope, docNo, action string, at time.Time, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO ap_payment_prepare_events (company, docno, action, actor, occurred_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6)`, scope.Company, docNo, action, scope.Actor, at, body); err != nil {
		return fmt.Errorf("write payment preparation event: %w", err)
	}
	return nil
}

// CancelPrepare ยกเลิกระดับหัวเอกสารพร้อมเหตุผล (design §4.3 ข้อ 11-13, 15) และตรวจงวดของวันที่จ่าย ตาม Champ
// ConfirmDeleteData (apfrmapstatement1view.cpp:612-615 ห้ามเฉพาะ BillStatus ≠ 0 = จ่ายครบ; bcprocess.cpp:499-511):
// ใบที่จ่ายครบแล้วยกเลิกไม่ได้; ใบที่จ่ายบางส่วนยกเลิกได้ — การตัดหนี้ที่จ่ายแล้วคงอยู่ ยอดจองส่วนที่ยังไม่จ่ายถูกปล่อย
// (ยอดจองนับเฉพาะใบที่ยังไม่ยกเลิก — gl.APOpenItems)
func (s *Store) CancelPrepare(ctx context.Context, db *sql.DB, scope gl.Scope, docNo, reason string, version int64) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return erpBadRequest("ap_prepare_cancel_reason_required", "reason")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = gl.LockCompanyTx(ctx, tx, scope.Company); err != nil {
		return err
	}
	payDay, _, err := lockPrepareHeader(ctx, tx, scope.Company, docNo, version)
	if err != nil {
		return err
	}
	if err = checkDocDates(ctx, tx, scope, payDay); err != nil {
		return err
	}
	uses, err := loadPrepareLineUse(ctx, tx, scope.Company, []string{docNo}, "")
	if err != nil {
		return err
	}
	if fullyPaid(uses) {
		return erpError("ap_prepare_paid_locked", "")
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE ap_payment_prepare SET iscancel=true, cancel_reason=$3, cancelled_at=$4, cancelled_by=$5,
		version=version+1, modified_at=$4, modified_by=$5 WHERE company=$1 AND docno=$2`, scope.Company, docNo, reason, now, scope.Actor); err != nil {
		return fmt.Errorf("cancel payment preparation: %w", err)
	}
	if err = writePrepareEvent(ctx, tx, scope, docNo, "cancel", now, map[string]any{"reason": reason}); err != nil {
		return err
	}
	return tx.Commit()
}

// PrepareEvents คือประวัติการบันทึก/แก้/ยกเลิกของใบเตรียมจ่าย เรียงตามเวลา
func (s *Store) PrepareEvents(ctx context.Context, db *sql.DB, company, docNo string) ([]apPrepareEvent, error) {
	rows, err := db.QueryContext(ctx, `SELECT event_no, action, actor, occurred_at, payload FROM ap_payment_prepare_events
		WHERE company=$1 AND docno=$2 ORDER BY event_no`, company, docNo)
	if err != nil {
		return nil, fmt.Errorf("read payment preparation events: %w", err)
	}
	defer rows.Close()
	out := []apPrepareEvent{}
	for rows.Next() {
		var event apPrepareEvent
		var at time.Time
		var payload []byte
		if err = rows.Scan(&event.EventNo, &event.Action, &event.Actor, &at, &payload); err != nil {
			return nil, fmt.Errorf("scan payment preparation event: %w", err)
		}
		event.OccurredAt, event.Payload = at.UTC().Format(time.RFC3339), json.RawMessage(payload)
		out = append(out, event)
	}
	return out, rows.Err()
}

// apCreditor คือเจ้าหนี้ที่เลือกได้ในจอเตรียมจ่าย
type apCreditor struct {
	Code       string `json:"partner_code"`
	Name       string `json:"name"`
	CreditDays int    `json:"credit_days"`
}

// ListCreditors คือเจ้าหนี้ที่ใช้งานอยู่ของบริษัท (ค้นด้วยรหัส/ชื่อ) — จอเตรียมจ่ายใช้สิทธิ์ของตัวเอง ไม่ต้องมีสิทธิ์จอบัญชี
func (s *Store) ListCreditors(ctx context.Context, db *sql.DB, company, q string, limit, offset int) ([]apCreditor, int, error) {
	if q = strings.TrimSpace(q); q != "" {
		q = "%" + q + "%"
	}
	rows, err := db.QueryContext(ctx, `SELECT code, COALESCE(payload->>'title_name',''), COALESCE(payload->>'name_th',''),
		COALESCE((payload->>'credit_days')::int, 0), COUNT(*) OVER() FROM gl_subledger_partners
		WHERE company=$1 AND COALESCE((payload->>'is_supplier')::boolean, false) AND COALESCE((payload->>'is_active')::boolean, false)
			AND ($2::text='' OR code ILIKE $2 OR payload->>'name_th' ILIKE $2)
		ORDER BY code LIMIT $3 OFFSET $4`, company, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list creditors: %w", err)
	}
	defer rows.Close()
	out, total := []apCreditor{}, 0
	for rows.Next() {
		var c apCreditor
		var title, name string
		if err = rows.Scan(&c.Code, &title, &name, &c.CreditDays, &total); err != nil {
			return nil, 0, fmt.Errorf("scan creditor: %w", err)
		}
		c.Name = gl.PartnerFullName(title, name)
		out = append(out, c)
	}
	return out, total, rows.Err()
}

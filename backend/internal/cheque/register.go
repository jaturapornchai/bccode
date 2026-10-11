package cheque

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
)

// Register คือเช็คหนึ่งใบในทะเบียน (เงินเป็นสตริงทศนิยม 2 ตำแหน่ง; balance_amount = ยอดที่ยังไม่ถูกเอกสารต้นทางตัด)
type Register struct {
	ID             string    `json:"id"`
	ChequeType     int       `json:"cheque_type"`
	DocNo          string    `json:"doc_no"`
	DocDate        string    `json:"doc_date"`
	DueDate        string    `json:"due_date"`
	PartnerCode    string    `json:"partner_code"`
	PartnerName    string    `json:"partner_name"`
	ChequeNo       string    `json:"cheque_no"`
	BankCode       string    `json:"bank_code"`
	BankBranchCode string    `json:"bank_branch_code"`
	BookCode       string    `json:"book_code"`
	Status         int       `json:"status"`
	ExtStatusCode  string    `json:"ext_status_code"`
	CurrencyCode   string    `json:"currency_code"`
	ExchangeRate   string    `json:"exchange_rate"`
	Amount         string    `json:"amount"`
	HomeAmount     string    `json:"home_amount"`
	BalanceAmount  string    `json:"balance_amount"`
	Description    string    `json:"description"`
	SaleCode       string    `json:"sale_code"`
	DepartmentCode string    `json:"department_code"`
	SourceKind     string    `json:"source_kind"`
	PrepareDocNo   string    `json:"prepare_doc_no"`
	PrintCount     int       `json:"print_count"`
	Version        int64     `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      string    `json:"updated_by"`
}

// RegisterInput คือค่าที่ผู้ใช้กรอกในจอทะเบียน (เงิน/อัตราเป็นสตริงทศนิยม); Status nil = สถานะเช็คในมือ (สร้าง) / คงเดิม (แก้);
// Confirm = ผู้ใช้ยืนยันแก้ฟิลด์หลักของเช็คที่ถูกอ้างแล้ว หรือแก้สถานะในจอทะเบียน (D8)
type RegisterInput struct {
	ID             string
	Version        int64
	ChequeType     int
	DocNo          string
	DocDate        string
	DueDate        string
	PartnerCode    string
	ChequeNo       string
	BankCode       string
	BankBranchCode string
	BookCode       string
	Status         *int
	ExtStatusCode  string
	CurrencyCode   string
	ExchangeRate   string
	Amount         string
	Description    string
	SaleCode       string
	DepartmentCode string
	Confirm        bool
}

const (
	maxDocNoLen       = 30
	maxChequeNoLen    = 20
	maxDescriptionLen = 255
	maxRefCodeLen     = 50
	homeCurrency      = "THB"
)

var (
	moneyPattern = regexp.MustCompile(`^[0-9]{1,16}(\.[0-9]{1,2})?$`)
	ratePattern  = regexp.MustCompile(`^[0-9]{1,9}(\.[0-9]{1,10})?$`)
)

// parseMoney อ่านเงินจากสตริงทศนิยมไม่เกิน 2 ตำแหน่ง (ไม่รับ float, เครื่องหมาย, จุลภาค หรือรูปแบบยกกำลัง)
func parseMoney(value string) (decimal.Decimal, bool) {
	value = strings.TrimSpace(value)
	if !moneyPattern.MatchString(value) {
		return decimal.Zero, false
	}
	d, err := decimal.NewFromString(value)
	return d, err == nil
}

// money2 แสดงเงินเป็นสตริงทศนิยม 2 ตำแหน่ง
func money2(d decimal.Decimal) string { return d.StringFixed(2) }

// homeAmount = ปัดครึ่งขึ้น(amount × rate) 2 ตำแหน่ง (Champ XRound cqfrmcheqindeposit.cpp:475-476; ค่าบวกเสมอ จึงเท่ากับ decimal.Round)
func homeAmount(amount, rate decimal.Decimal) decimal.Decimal { return amount.Mul(rate).Round(2) }

func validISODate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

// cleanText ตัดช่องว่างหัวท้าย และตรวจว่าไม่ยาวเกินและไม่มีอักขระควบคุม
func cleanText(value, field string, max int, required bool) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return "", badRequest("cheque_err_required_field", field)
		}
		return "", nil
	}
	if utf8.RuneCountInString(value) > max {
		return "", badRequest("cheque_err_too_long", field, strconv.Itoa(max))
	}
	if strings.ContainsFunc(value, unicode.IsControl) {
		return "", badRequest("cheque_err_text_invalid", field)
	}
	return value, nil
}

// registerValues คือค่าในทะเบียนที่ตรวจรูปแบบแล้ว
type registerValues struct {
	chequeType                                               int
	docNo, docDate, dueDate, partner, chequeNo, bank, branch string
	book, extStatus, description, sale, department           string
	status                                                   *int
	rate, amount, home                                       decimal.Decimal
}

// normalizeRegister ตรวจรูปแบบของค่าที่กรอก (ไม่แตะฐานข้อมูล) ตามลำดับช่องบนจอ
func normalizeRegister(in RegisterInput) (registerValues, error) {
	var (
		v   = registerValues{chequeType: in.ChequeType, status: in.Status}
		err error
	)
	if in.ChequeType != TypeReceived && in.ChequeType != TypeIssued {
		return v, badRequest("cheque_err_type_invalid", "cheque_type")
	}
	if v.docNo, err = cleanText(in.DocNo, "doc_no", maxDocNoLen, false); err != nil {
		return v, err
	}
	for _, d := range []struct {
		value, field string
		out          *string
	}{{in.DocDate, "doc_date", &v.docDate}, {in.DueDate, "due_date", &v.dueDate}} {
		*d.out = strings.TrimSpace(d.value)
		if *d.out == "" {
			return v, badRequest("cheque_err_required_field", d.field)
		}
		if !validISODate(*d.out) {
			return v, badRequest("cheque_err_date_invalid", d.field)
		}
	}
	// เช็คจ่าย: วันที่จ่ายต้องไม่เกินวันที่ถึงกำหนด (เช็ครับรับเช็คที่ถึงกำหนดแล้วได้)
	if in.ChequeType == TypeIssued && v.docDate > v.dueDate {
		return v, badRequest("cheque_err_due_before_issue", "due_date")
	}
	if v.partner, err = cleanText(in.PartnerCode, "partner_code", maxRefCodeLen, true); err != nil {
		return v, err
	}
	if v.chequeNo, err = cleanText(in.ChequeNo, "cheque_no", maxChequeNoLen, true); err != nil {
		return v, err
	}
	if v.bank, err = cleanText(in.BankCode, "bank_code", maxCodeLen, true); err != nil {
		return v, err
	}
	if v.branch, err = cleanText(in.BankBranchCode, "bank_branch_code", maxCodeLen, true); err != nil {
		return v, err
	}
	if v.book, err = cleanText(in.BookCode, "book_code", maxRefCodeLen, in.ChequeType == TypeIssued); err != nil {
		if e, ok := AsError(err); ok && e.Key == "cheque_err_required_field" {
			return v, badRequest("cheque_err_book_required", "book_code")
		}
		return v, err
	}
	if v.extStatus, err = cleanText(in.ExtStatusCode, "ext_status_code", maxCodeLen, false); err != nil {
		return v, err
	}
	if v.amount, err = parseAmount(in.Amount); err != nil {
		return v, err
	}
	if v.rate, err = parseRate(in.CurrencyCode, in.ExchangeRate); err != nil {
		return v, err
	}
	v.home = homeAmount(v.amount, v.rate)
	if v.description, err = cleanText(in.Description, "description", maxDescriptionLen, false); err != nil {
		return v, err
	}
	if v.sale, err = cleanText(in.SaleCode, "sale_code", maxRefCodeLen, false); err != nil {
		return v, err
	}
	if in.ChequeType == TypeIssued {
		v.sale = "" // พนักงานขายมีเฉพาะเช็ครับ
	}
	if v.department, err = cleanText(in.DepartmentCode, "department_code", maxRefCodeLen, false); err != nil {
		return v, err
	}
	return v, nil
}

func parseAmount(value string) (decimal.Decimal, error) {
	amount, ok := parseMoney(value)
	if !ok || !amount.IsPositive() {
		return decimal.Zero, badRequest("cheque_err_amount_invalid", "amount")
	}
	return amount, nil
}

// parseRate: จนกว่า D15 (สกุลเงินต่างประเทศ) จะตัดสิน รับเฉพาะเงินบาท อัตรา 1 (Champ บังคับ 1 เมื่อไม่เปิดหลายสกุลเงิน cqfrmcheqin.cpp:1135)
func parseRate(currency, rate string) (decimal.Decimal, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency != "" && currency != homeCurrency {
		return decimal.Zero, badRequest("cheque_err_currency_not_supported", "currency_code")
	}
	rate = strings.TrimSpace(rate)
	if rate == "" {
		return decimal.NewFromInt(1), nil
	}
	if !ratePattern.MatchString(rate) {
		return decimal.Zero, badRequest("cheque_err_rate_invalid", "exchange_rate")
	}
	d, err := decimal.NewFromString(rate)
	if err != nil || !d.IsPositive() {
		return decimal.Zero, badRequest("cheque_err_rate_invalid", "exchange_rate")
	}
	if !d.Equal(decimal.NewFromInt(1)) {
		return decimal.Zero, badRequest("cheque_err_currency_not_supported", "exchange_rate")
	}
	return d, nil
}

// validRegisterStatus: สถานะที่ตั้งในจอทะเบียนได้ — ตอนสร้าง เช็ครับ 0..6 / เช็คจ่าย 0 หรือ 3 (เตรียมจ่าย);
// แก้ภายหลัง (D8) เลือกได้ทุกสถานะของชนิดนั้น; เช็ครับสถานะ ≠ 0 ต้องมีสมุด
func validRegisterStatus(chequeType, status int, creating bool) bool {
	if chequeType == TypeReceived {
		return status >= InOnHand && status <= InRedeposited
	}
	if creating {
		return status == OutOnHand || status == OutPrepared
	}
	return status >= OutOnHand && status <= OutPrepared
}

// registerBook: เช็ครับในมือระบุสมุดหรือไม่ก็ได้ (Champ cqfrmcheqin.cpp:391-400,488 บันทึก BookNo ตามที่เลือก ใช้กรองตอนนำเช็คเข้า 420/424);
// เช็ครับสถานะอื่นต้องมีสมุด ยกเว้นยกเลิกแล้ว (ยกเลิกเช็คในมือ 423 ชนิด 0 ไม่ใช้สมุด); เช็คจ่ายบังคับสมุดตั้งแต่ normalizeRegister
func registerBook(chequeType, status int, book string) (string, error) {
	if chequeType != TypeReceived || status == InOnHand {
		return book, nil
	}
	if book == "" && status != InCancelled {
		return "", badRequest("cheque_err_book_required", "book_code")
	}
	return book, nil
}

// checkRegisterRefs ตรวจรหัสที่อ้างข้ามตาราง (ล็อก FOR SHARE ไว้จนจบ tx): ลูกหนี้/เจ้าหนี้ตามชนิดเช็ค, ธนาคาร, สาขา, สมุด, สถานะเพิ่มเติม
// changed บอกว่าช่องไหนต้องตรวจ (ตอนแก้ ตรวจเฉพาะช่องที่เปลี่ยน เพื่อไม่ให้ทะเบียนเก่าบันทึกไม่ได้เพราะรหัสถูกปิดภายหลัง)
func checkRegisterRefs(ctx context.Context, tx *sql.Tx, company string, v registerValues, changed func(field string) bool) error {
	if changed("partner_code") {
		if err := requirePartner(ctx, tx, company, v.partner, v.chequeType); err != nil {
			return err
		}
	}
	for _, c := range []struct{ kind, code, field string }{
		{KindBank, v.bank, "bank_code"}, {KindBankBranch, v.branch, "bank_branch_code"}, {KindChequeExtStatus, v.extStatus, "ext_status_code"},
	} {
		if c.code == "" || !changed(c.field) {
			continue
		}
		ok, err := CodeExists(ctx, tx, company, c.kind, c.code)
		if err != nil {
			return err
		}
		if !ok {
			return notFound("bank_code_err_not_found", c.field, c.code)
		}
	}
	if v.book != "" && changed("book_code") {
		if _, err := requireBook(ctx, tx, company, v.book); err != nil {
			return err
		}
	}
	return nil
}

func requirePartner(ctx context.Context, tx *sql.Tx, company, code string, chequeType int) error {
	var customer, supplier, active bool
	err := tx.QueryRowContext(ctx, `SELECT COALESCE((payload->>'is_customer')::boolean, false), COALESCE((payload->>'is_supplier')::boolean, false),
		COALESCE((payload->>'is_active')::boolean, true) FROM gl_subledger_partners WHERE company = $1 AND code = $2 FOR SHARE`, company, code).Scan(&customer, &supplier, &active)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	role := customer
	if chequeType == TypeIssued {
		role = supplier
	}
	if errors.Is(err, sql.ErrNoRows) || !role || !active {
		return badRequest("cheque_err_partner_invalid", "partner_code", code)
	}
	return nil
}

// bankBook คือสมุดเงินฝากที่เอกสารอ้าง (gl_subledger_bank_accounts.payload)
type bankBook struct {
	code, glAccount string
}

// requireBook ตรวจว่าสมุดเงินฝากมีอยู่และใช้งานอยู่ (FOR SHARE กันถูกแก้/ลบระหว่าง tx)
func requireBook(ctx context.Context, tx *sql.Tx, company, code string) (bankBook, error) {
	var (
		b      = bankBook{code: code}
		active bool
	)
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(payload->>'gl_account_code', ''), COALESCE((payload->>'is_active')::boolean, false)
		FROM gl_subledger_bank_accounts WHERE company = $1 AND code = $2 FOR SHARE`, company, code).Scan(&b.glAccount, &active)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !active) {
		return b, badRequest("cheque_err_book_invalid", "book_code", code)
	}
	return b, err
}

// nextDocNo ออกเลขที่ <prefix><ปี พ.ศ. 2 หลัก><เดือน>-<ลำดับ 4 หลัก> ตามวันที่เอกสาร (รูปแบบเดียวกับเอกสาร ERP)
// ผู้เรียกต้องถือล็อกเลขที่ (lockNumbering) ก่อน
func nextDocNo(ctx context.Context, tx *sql.Tx, query string, prefix, date string, args ...any) (string, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}
	period := fmt.Sprintf("%s%02d%02d", prefix, (t.Year()+543)%100, int(t.Month()))
	var last sql.NullString
	if err = tx.QueryRowContext(ctx, query, append(args, period+"-%")...).Scan(&last); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	seq := 1
	if last.Valid {
		if i := strings.LastIndex(last.String, "-"); i >= 0 {
			if n, err := strconv.Atoi(last.String[i+1:]); err == nil {
				seq = n + 1
			}
		}
	}
	return fmt.Sprintf("%s-%04d", period, seq), nil
}

// lockNumbering กันสองคำขอออกเลขที่เดียวกันพร้อมกัน (ต่อบริษัท + ชุดเลข); ตัวคั่นต้องไม่ใช่ NUL เพราะ text ของ PostgreSQL รับไม่ได้
func lockNumbering(ctx context.Context, tx *sql.Tx, company, series string) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('bc_cheque_doc_no'), hashtext($1))`, company+"/"+series)
	return err
}

func registerPrefix(chequeType int) string {
	flag := 61 // บันทึกเช็ครับ CHQ-IN
	if chequeType == TypeIssued {
		flag = 418 // บันทึกเช็คจ่าย CHQ-OUT
	}
	dt, _ := doctype.ByFlag(flag)
	return dt.Prefix
}

// duplicateError แปลงการชน unique index ของทะเบียนเป็นข้อความที่บอกช่องที่ซ้ำ
func duplicateError(err error, v registerValues) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Constraint == "ux_cheque_doc_no" {
		return newError(http.StatusConflict, "cheque_err_duplicate_doc_no", "doc_no", v.docNo)
	}
	return newError(http.StatusConflict, "cheque_err_duplicate_number", "cheque_no", v.chequeNo)
}

const registerSelect = `SELECT r.id::text, r.cheque_type, r.doc_no, to_char(r.doc_date, 'YYYY-MM-DD'), to_char(r.due_date, 'YYYY-MM-DD'),
	r.partner_code, COALESCE(p.payload->>'name_th', ''), r.cheque_no, r.bank_code, r.bank_branch_code, COALESCE(r.book_code, ''),
	r.status, COALESCE(r.ext_status_code, ''), r.currency_code, trim_scale(r.exchange_rate)::text, r.amount::text, r.home_amount::text,
	(r.home_amount - COALESCE(l.linked, 0))::numeric(18,2)::text, r.description, r.sale_code, r.department_code, r.source_kind,
	COALESCE(r.prepare_doc_no, ''), r.print_count, r.version, r.created_at, r.created_by, r.updated_at, r.updated_by
FROM cheque_registers r
LEFT JOIN gl_subledger_partners p ON p.company = r.company AND p.code = r.partner_code
LEFT JOIN LATERAL (SELECT SUM(s.amount) AS linked FROM cheque_source_links s
	WHERE s.company = r.company AND s.cheque_id = r.id AND s.voided_at IS NULL) l ON true`

func scanRegister(row rowScanner) (Register, error) {
	var r Register
	err := row.Scan(&r.ID, &r.ChequeType, &r.DocNo, &r.DocDate, &r.DueDate, &r.PartnerCode, &r.PartnerName, &r.ChequeNo,
		&r.BankCode, &r.BankBranchCode, &r.BookCode, &r.Status, &r.ExtStatusCode, &r.CurrencyCode, &r.ExchangeRate,
		&r.Amount, &r.HomeAmount, &r.BalanceAmount, &r.Description, &r.SaleCode, &r.DepartmentCode, &r.SourceKind,
		&r.PrepareDocNo, &r.PrintCount, &r.Version, &r.CreatedAt, &r.CreatedBy, &r.UpdatedAt, &r.UpdatedBy)
	return r, err
}

func loadRegister(ctx context.Context, q querier, company, id string) (Register, error) {
	r, err := scanRegister(q.QueryRowContext(ctx, registerSelect+` WHERE r.company = $1 AND r.id = $2 AND r.deleted_at IS NULL`, company, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Register{}, notFound("cheque_err_not_found", "id")
	}
	return r, err
}

// lockedRegister คือค่าปัจจุบันของทะเบียนที่ล็อกแถวแล้ว (FOR UPDATE)
type lockedRegister struct {
	registerValues
	id, sourceKind string
	statusNow      int
	linked         decimal.Decimal
	version        int64
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// lockRegister ล็อกเช็คหนึ่งใบของชนิดที่ระบุ (ชนิดไม่ตรง = ไม่พบ เพราะจอรับ/จ่ายแยกกัน) แล้วตรวจ version
func lockRegister(ctx context.Context, tx *sql.Tx, company string, chequeType int, id string, version int64) (lockedRegister, error) {
	var (
		l              lockedRegister
		rate, amt, hom string
		linked         string
	)
	if !uuidPattern.MatchString(id) {
		return l, notFound("cheque_err_not_found", "id")
	}
	err := tx.QueryRowContext(ctx, `SELECT r.id::text, r.cheque_type, r.doc_no, to_char(r.doc_date, 'YYYY-MM-DD'), to_char(r.due_date, 'YYYY-MM-DD'),
			r.partner_code, r.cheque_no, r.bank_code, r.bank_branch_code, COALESCE(r.book_code, ''), r.status, COALESCE(r.ext_status_code, ''),
			r.exchange_rate::text, r.amount::text, r.home_amount::text, r.description, r.sale_code, r.department_code, r.source_kind, r.version,
			(SELECT COALESCE(SUM(s.amount), 0)::text FROM cheque_source_links s WHERE s.company = r.company AND s.cheque_id = r.id AND s.voided_at IS NULL)
		FROM cheque_registers r WHERE r.company = $1 AND r.id = $2 AND r.deleted_at IS NULL FOR UPDATE OF r`, company, id).Scan(
		&l.id, &l.chequeType, &l.docNo, &l.docDate, &l.dueDate, &l.partner, &l.chequeNo, &l.bank, &l.branch, &l.book, &l.statusNow,
		&l.extStatus, &rate, &amt, &hom, &l.description, &l.sale, &l.department, &l.sourceKind, &l.version, &linked)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && l.chequeType != chequeType) {
		return l, notFound("cheque_err_not_found", "id")
	}
	if err != nil {
		return l, err
	}
	l.rate, l.amount, l.home, l.linked = decimal.RequireFromString(rate), decimal.RequireFromString(amt), decimal.RequireFromString(hom), decimal.RequireFromString(linked)
	l.status = intPtr(l.statusNow)
	if l.version != version {
		return l, versionConflict()
	}
	return l, nil
}

// chequeRefs คือเอกสารที่อ้างเช็ค: statusDoc = เอกสารสถานะที่ยังใช้งาน, linkDoc = เอกสารต้นทางที่ยังตัดยอด,
// historyDoc = เอกสารแรกที่เคยอ้างเช็ค (รวมที่ยกเลิก/แก้ออกแล้ว) ใช้บอกเอกสารต้นทางในข้อความ cheque_err_from_source_doc
type chequeRefs struct{ statusDoc, linkDoc, historyDoc string }

func (r chequeRefs) active() string {
	if r.statusDoc != "" {
		return r.statusDoc
	}
	return r.linkDoc
}

func loadRefs(ctx context.Context, q querier, company, id string) (chequeRefs, error) {
	var r chequeRefs
	err := q.QueryRowContext(ctx, `SELECT
		COALESCE((SELECT doc_no FROM cheque_events WHERE company = $1 AND cheque_id = $2 AND event_kind = 'status_doc' AND voided_at IS NULL ORDER BY event_date DESC, id DESC LIMIT 1), ''),
		COALESCE((SELECT source_doc_no FROM cheque_source_links WHERE company = $1 AND cheque_id = $2 AND voided_at IS NULL ORDER BY source_date DESC, created_at DESC LIMIT 1), ''),
		COALESCE((SELECT doc_no FROM cheque_events WHERE company = $1 AND cheque_id = $2 AND event_kind = 'status_doc' ORDER BY id LIMIT 1),
			(SELECT source_doc_no FROM cheque_source_links WHERE company = $1 AND cheque_id = $2 ORDER BY created_at LIMIT 1), '')`,
		company, id).Scan(&r.statusDoc, &r.linkDoc, &r.historyDoc)
	return r, err
}

// CreateRegister บันทึกเช็คใหม่ในทะเบียน + event create ใน tx เดียว (เลขที่เอกสารว่าง = ออกเลขอัตโนมัติ)
func CreateRegister(ctx context.Context, db *sql.DB, scope Scope, requestID string, in RegisterInput) (Register, error) {
	v, err := normalizeRegister(in)
	if err != nil {
		return Register{}, err
	}
	status := InOnHand
	if v.status != nil {
		status = *v.status
	}
	if !validRegisterStatus(v.chequeType, status, true) {
		return Register{}, badRequest("cheque_err_status_invalid", "status")
	}
	if v.book, err = registerBook(v.chequeType, status, v.book); err != nil {
		return Register{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Register{}, err
	}
	defer tx.Rollback()
	var out Register
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "registers:create", &out); err != nil || replayed {
		return out, err
	}
	if err = checkRegisterRefs(ctx, tx, scope.Company, v, func(string) bool { return true }); err != nil {
		return Register{}, err
	}
	if v.docNo == "" {
		if err = lockNumbering(ctx, tx, scope.Company, "register:"+strconv.Itoa(v.chequeType)); err != nil {
			return Register{}, err
		}
		if v.docNo, err = nextDocNo(ctx, tx, `SELECT doc_no FROM cheque_registers WHERE company = $1 AND cheque_type = $2 AND doc_no LIKE $3 ORDER BY doc_no DESC LIMIT 1`,
			registerPrefix(v.chequeType), v.docDate, scope.Company, v.chequeType); err != nil {
			return Register{}, err
		}
	}
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO cheque_registers (company, id, cheque_type, doc_no, doc_date, due_date, partner_code, cheque_no,
			bank_code, bank_branch_code, book_code, status, ext_status_code, currency_code, exchange_rate, amount, home_amount,
			description, sale_code, department_code, created_by, updated_by)
		VALUES ($1, gen_random_uuid(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $20)
		RETURNING id::text`,
		scope.Company, v.chequeType, v.docNo, v.docDate, v.dueDate, v.partner, v.chequeNo, v.bank, v.branch, nullable(v.book), status,
		nullable(v.extStatus), homeCurrency, v.rate.String(), money2(v.amount), money2(v.home), v.description, v.sale, v.department, scope.Actor).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return Register{}, duplicateError(err, v)
		}
		return Register{}, err
	}
	if err = insertEvent(ctx, tx, scope, eventRow{chequeID: id, kind: eventCreate, date: v.docDate, to: intPtr(status), book: v.book}); err != nil {
		return Register{}, err
	}
	if out, err = loadRegister(ctx, tx, scope.Company, id); err != nil {
		return Register{}, err
	}
	if err = completeRequest(ctx, tx, scope, requestID, out); err != nil {
		return Register{}, err
	}
	return out, tx.Commit()
}

// change คือค่าเดิม → ค่าใหม่ของหนึ่งช่อง (เก็บใน detail ของ event)
type change struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// UpdateRegister แก้ทะเบียน (D8): ฟิลด์หลักของเช็คที่ถูกอ้างแล้วแก้ได้หลังผู้ใช้ยืนยัน (Confirm) และบันทึก event edit;
// สถานะแก้ตรงได้หลังยืนยันและบันทึก event manual_status; วันที่ถึงกำหนดแก้ได้เฉพาะเมื่อยังไม่มีเอกสารสถานะ;
// ไม่ลงบัญชีและไม่ตรวจงวดปิด (Champ ไม่ตรวจ เพราะทะเบียนไม่ลงบัญชี)
func UpdateRegister(ctx context.Context, db *sql.DB, scope Scope, requestID string, in RegisterInput) (Register, error) {
	v, err := normalizeRegister(in)
	if err != nil {
		return Register{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Register{}, err
	}
	defer tx.Rollback()
	var out Register
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "registers:update", &out); err != nil || replayed {
		return out, err
	}
	cur, err := lockRegister(ctx, tx, scope.Company, v.chequeType, strings.TrimSpace(in.ID), in.Version)
	if err != nil {
		return Register{}, err
	}
	if v.docNo == "" {
		v.docNo = cur.docNo
	}
	newStatus := cur.statusNow
	if v.status != nil {
		newStatus = *v.status
	}
	if v.book, err = registerBook(v.chequeType, newStatus, v.book); err != nil {
		return Register{}, err
	}

	main := map[string]change{}
	minor := map[string]change{}
	track := func(target map[string]change, field, from, to string) {
		if from != to {
			target[field] = change{From: from, To: to}
		}
	}
	track(main, "doc_no", cur.docNo, v.docNo)
	track(main, "doc_date", cur.docDate, v.docDate)
	track(main, "partner_code", cur.partner, v.partner)
	track(main, "cheque_no", cur.chequeNo, v.chequeNo)
	track(main, "bank_code", cur.bank, v.bank)
	track(main, "bank_branch_code", cur.branch, v.branch)
	track(main, "book_code", cur.book, v.book)
	track(main, "amount", money2(cur.amount), money2(v.amount))
	track(main, "exchange_rate", cur.rate.String(), v.rate.String())
	track(minor, "due_date", cur.dueDate, v.dueDate)
	track(minor, "ext_status_code", cur.extStatus, v.extStatus)
	track(minor, "description", cur.description, v.description)
	track(minor, "sale_code", cur.sale, v.sale)
	track(minor, "department_code", cur.department, v.department)

	refs, err := loadRefs(ctx, tx, scope.Company, cur.id)
	if err != nil {
		return Register{}, err
	}
	if len(main) > 0 && refs.active() != "" && !in.Confirm {
		// Champ CANNOT_UPDATE_REFERENCE / _OTHER (cqfrmcheqin.cpp:320-336) ถามยืนยันก่อนแก้
		return Register{}, newError(http.StatusConflict, "cheque_err_referenced", "confirm", refs.active())
	}
	if _, ok := minor["due_date"]; ok && refs.statusDoc != "" {
		return Register{}, badRequest("cheque_err_due_date_locked", "due_date", refs.statusDoc)
	}
	if newStatus != cur.statusNow {
		if !in.Confirm {
			return Register{}, newError(http.StatusConflict, "cheque_err_status_confirm", "status")
		}
		if !validRegisterStatus(v.chequeType, newStatus, false) {
			return Register{}, badRequest("cheque_err_status_invalid", "status")
		}
	}
	if v.home.LessThan(cur.linked) {
		return Register{}, newError(http.StatusConflict, "cheque_err_over_allocated", "amount", money2(cur.linked))
	}
	changedField := func(field string) bool {
		_, isMain := main[field]
		_, isMinor := minor[field]
		return isMain || isMinor
	}
	if err = checkRegisterRefs(ctx, tx, scope.Company, v, changedField); err != nil {
		return Register{}, err
	}
	if len(main) == 0 && len(minor) == 0 && newStatus == cur.statusNow {
		if out, err = loadRegister(ctx, tx, scope.Company, cur.id); err != nil {
			return Register{}, err
		}
		if err = completeRequest(ctx, tx, scope, requestID, out); err != nil {
			return Register{}, err
		}
		return out, tx.Commit()
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cheque_registers SET doc_no = $3, doc_date = $4, due_date = $5, partner_code = $6, cheque_no = $7,
			bank_code = $8, bank_branch_code = $9, book_code = $10, status = $11, ext_status_code = $12, exchange_rate = $13, amount = $14,
			home_amount = $15, description = $16, sale_code = $17, department_code = $18, version = version + 1, updated_at = now(), updated_by = $19
		WHERE company = $1 AND id = $2`,
		scope.Company, cur.id, v.docNo, v.docDate, v.dueDate, v.partner, v.chequeNo, v.bank, v.branch, nullable(v.book), newStatus,
		nullable(v.extStatus), v.rate.String(), money2(v.amount), money2(v.home), v.description, v.sale, v.department, scope.Actor); err != nil {
		if isUniqueViolation(err) {
			return Register{}, duplicateError(err, v)
		}
		return Register{}, err
	}
	today := businessDay(time.Now())
	if len(main) > 0 || len(minor) > 0 {
		kind := eventUpdate
		if len(main) > 0 {
			kind = eventEdit
		}
		for field, c := range minor {
			main[field] = c
		}
		if err = insertEvent(ctx, tx, scope, eventRow{chequeID: cur.id, kind: kind, date: today, book: v.book, detail: main}); err != nil {
			return Register{}, err
		}
	}
	if newStatus != cur.statusNow {
		if err = insertEvent(ctx, tx, scope, eventRow{chequeID: cur.id, kind: eventManualStatus, date: today,
			from: intPtr(cur.statusNow), to: intPtr(newStatus), book: v.book}); err != nil {
			return Register{}, err
		}
	}
	if out, err = loadRegister(ctx, tx, scope.Company, cur.id); err != nil {
		return Register{}, err
	}
	if err = completeRequest(ctx, tx, scope, requestID, out); err != nil {
		return Register{}, err
	}
	return out, tx.Commit()
}

// CancelRegister ลบเช็คที่ไม่มีเอกสารใดใช้งานอยู่ (Champ cqfrmcheqin.cpp:438-470 / cqfrmcheqout.cpp:665-667):
// ถูกอ้างอยู่ = MSG_REFERANCE, มาจากเอกสารต้นทาง = MSG_FROM_OTHER; เอกสารสถานะที่ยกเลิกหมดแล้วไม่นับเป็นการอ้าง
// (Champ ลบ Sub ของเอกสารสถานะเมื่อยกเลิก CQFrmBaseTransStatus.cpp:323-335) — BC ไม่ลบแถวจริง แต่ตั้ง deleted_at
// + event delete เพื่อให้เอกสารที่ยกเลิกแล้วและประวัติยังเปิดดูได้; รายการ/ยอด/การเลือกเช็คทุกจุดไม่นับแถวที่ลบ
func CancelRegister(ctx context.Context, db *sql.DB, scope Scope, requestID string, chequeType int, id string, version int64) (Register, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Register{}, err
	}
	defer tx.Rollback()
	var out Register
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "registers:cancel", &out); err != nil || replayed {
		return out, err
	}
	cur, err := lockRegister(ctx, tx, scope.Company, chequeType, strings.TrimSpace(id), version)
	if err != nil {
		return Register{}, err
	}
	refs, err := loadRefs(ctx, tx, scope.Company, cur.id)
	if err != nil {
		return Register{}, err
	}
	if doc := refs.active(); doc != "" {
		return Register{}, newError(http.StatusConflict, "cheque_err_referenced_delete", "id", doc)
	}
	if cur.sourceKind == "erp" {
		return Register{}, newError(http.StatusConflict, "cheque_err_from_source_doc", "id", refs.historyDoc)
	}
	if out, err = loadRegister(ctx, tx, scope.Company, cur.id); err != nil {
		return Register{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cheque_registers SET deleted_at = now(), deleted_by = $3, version = version + 1, updated_at = now(), updated_by = $3
		WHERE company = $1 AND id = $2 AND deleted_at IS NULL`, scope.Company, cur.id, scope.Actor); err != nil {
		return Register{}, err
	}
	if err = insertEvent(ctx, tx, scope, eventRow{chequeID: cur.id, kind: eventDelete, date: businessDay(time.Now()), book: cur.book}); err != nil {
		return Register{}, err
	}
	if err = completeRequest(ctx, tx, scope, requestID, out); err != nil {
		return Register{}, err
	}
	return out, tx.Commit()
}

// Reference คือหนึ่งแถวในตารางเอกสารอ้างอิงใต้หัวเช็ค (cqfrmcheqin.cpp:1227-1300)
type Reference struct {
	KindKey     string `json:"kind_key"`
	DocNo       string `json:"doc_no"`
	DocDate     string `json:"doc_date"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
}

// RegisterDetail คือเช็คหนึ่งใบ + การปรับปรุงล่าสุด + เอกสารที่อ้างถึง (ไม่รวมที่ถูกยกเลิก) เรียงตามวันที่
type RegisterDetail struct {
	Register
	LastEvent  *Event      `json:"last_event"`
	References []Reference `json:"references"`
}

func GetRegister(ctx context.Context, q querier, company, id string) (RegisterDetail, error) {
	id = strings.TrimSpace(id)
	if !uuidPattern.MatchString(id) {
		return RegisterDetail{}, notFound("cheque_err_not_found", "id")
	}
	r, err := loadRegister(ctx, q, company, id)
	if err != nil {
		return RegisterDetail{}, err
	}
	detail := RegisterDetail{Register: r, References: []Reference{}}
	if detail.LastEvent, err = lastEvent(ctx, q, company, id); err != nil {
		return RegisterDetail{}, err
	}
	rows, err := q.QueryContext(ctx, `SELECT flag, kind, doc_no, doc_date, amount, description FROM (
			SELECT e.doc_flag AS flag, 'status_doc' AS kind, e.doc_no, e.event_date AS sort_date, e.id AS sort_id,
				to_char(e.event_date, 'YYYY-MM-DD') AS doc_date, dl.amount::text AS amount, d.description
			FROM cheque_events e
			JOIN cheque_docs d ON d.company = e.company AND d.transflag = e.doc_flag AND d.doc_no = e.doc_no
			JOIN cheque_doc_lines dl ON dl.company = e.company AND dl.transflag = e.doc_flag AND dl.doc_no = e.doc_no AND dl.cheque_id = e.cheque_id
			WHERE e.company = $1 AND e.cheque_id = $2 AND e.event_kind = 'status_doc' AND e.voided_at IS NULL
			UNION ALL
			SELECT COALESCE(s.source_flag, 0), s.source_kind, s.source_doc_no, s.source_date, 0,
				to_char(s.source_date, 'YYYY-MM-DD'), s.amount::text, ''
			FROM cheque_source_links s WHERE s.company = $1 AND s.cheque_id = $2 AND s.voided_at IS NULL
		) refs ORDER BY sort_date, sort_id, doc_no`, company, id)
	if err != nil {
		return RegisterDetail{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			flag int
			kind string
			ref  Reference
		)
		if err = rows.Scan(&flag, &kind, &ref.DocNo, &ref.DocDate, &ref.Amount, &ref.Description); err != nil {
			return RegisterDetail{}, err
		}
		ref.KindKey = kind
		if dt, ok := doctype.ByFlag(flag); ok {
			ref.KindKey = dt.Key
		}
		detail.References = append(detail.References, ref)
	}
	return detail, rows.Err()
}

// RegisterFilter คือเงื่อนไขค้นหาทะเบียน (วันที่เป็น ISO YYYY-MM-DD; Status nil = ทุกสถานะ)
type RegisterFilter struct {
	ChequeType int
	Status     *int
	Partner    string
	Book       string
	DueFrom    string
	DueTo      string
	Query      string
	Page       int
	Limit      int
}

// RegisterPage คือผลค้นหาหนึ่งหน้า
type RegisterPage struct {
	Items []Register
	Total int
	Page  int
	Limit int
}

const (
	defaultRegisterLimit = 50
	maxRegisterLimit     = 500
)

// ListRegisters ค้นทะเบียนเช็คของชนิดหนึ่ง เรียงตามวันที่ถึงกำหนด
func ListRegisters(ctx context.Context, q querier, company string, f RegisterFilter) (RegisterPage, error) {
	if f.ChequeType != TypeReceived && f.ChequeType != TypeIssued {
		return RegisterPage{}, badRequest("cheque_err_type_invalid", "type")
	}
	where := []string{"r.company = $1", "r.cheque_type = $2"}
	args := []any{company, f.ChequeType}
	add := func(cond string, value any) {
		args = append(args, value)
		where = append(where, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(args))))
	}
	if f.Status != nil {
		add("r.status = ?", *f.Status)
	}
	if s := strings.TrimSpace(f.Partner); s != "" {
		add("r.partner_code = ?", s)
	}
	if s := strings.TrimSpace(f.Book); s != "" {
		add("r.book_code = ?", s)
	}
	for _, d := range []struct{ value, field, cond string }{{f.DueFrom, "due_from", "r.due_date >= ?"}, {f.DueTo, "due_to", "r.due_date <= ?"}} {
		if s := strings.TrimSpace(d.value); s != "" {
			if !validISODate(s) {
				return RegisterPage{}, badRequest("cheque_err_date_invalid", d.field)
			}
			add(d.cond, s)
		}
	}
	if s := strings.TrimSpace(f.Query); s != "" {
		add("(r.cheque_no ILIKE ? OR r.doc_no ILIKE ? OR r.partner_code ILIKE ? OR r.description ILIKE ?)", likePattern(s))
	}
	return pageRegisters(ctx, q, where, args, f.Page, f.Limit)
}

func pageRegisters(ctx context.Context, q querier, where []string, args []any, page, limit int) (RegisterPage, error) {
	page, limit, offset := pageWindow(page, limit, defaultRegisterLimit, maxRegisterLimit)
	result := RegisterPage{Items: []Register{}, Page: page, Limit: limit}
	cond := strings.Join(append(where, "r.deleted_at IS NULL"), " AND ")
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM cheque_registers r WHERE `+cond, args...).Scan(&result.Total); err != nil {
		return RegisterPage{}, err
	}
	n := len(args)
	rows, err := q.QueryContext(ctx, registerSelect+` WHERE `+cond+fmt.Sprintf(` ORDER BY r.due_date, r.cheque_no, r.id LIMIT $%d OFFSET $%d`, n+1, n+2),
		append(args, limit, offset)...)
	if err != nil {
		return RegisterPage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		r, err := scanRegister(rows)
		if err != nil {
			return RegisterPage{}, err
		}
		result.Items = append(result.Items, r)
	}
	return result, rows.Err()
}

// SelectableFilter คือเงื่อนไขของเช็คที่เอกสารสถานะเลือกได้ (cheque-bank.md §4.3)
type SelectableFilter struct {
	Flag         int
	DocDate      string
	Book         string
	CancelType   int
	CancelPassed bool
	Query        string
	Page         int
	Limit        int
}

// ListSelectable คือเช็คที่เลือกเข้าเอกสารสถานะได้ ณ วันที่เอกสาร (ตรวจซ้ำอีกครั้งตอนบันทึกหลังล็อกแถว)
func ListSelectable(ctx context.Context, q querier, company string, f SelectableFilter) (RegisterPage, error) {
	rule, ok := ruleFor(f.Flag)
	if !ok {
		return RegisterPage{}, badRequest("cheque_err_transflag_invalid", "transflag", strconv.Itoa(f.Flag))
	}
	if f.Flag == FlagReceivedCancel && (f.CancelType < CancelOnHand || f.CancelType > CancelRedeposited) {
		return RegisterPage{}, badRequest("cheque_err_cancel_type_invalid", "cancel_type")
	}
	where := []string{"r.company = $1", "r.cheque_type = $2", "r.status = ANY($3::smallint[])"}
	args := []any{company, rule.chequeType, pq.Array(selectableStatuses(f.Flag, f.CancelType, f.CancelPassed))}
	add := func(cond string, value any) {
		args = append(args, value)
		where = append(where, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(args))))
	}
	if d := strings.TrimSpace(f.DocDate); d != "" {
		if !validISODate(d) {
			return RegisterPage{}, badRequest("cheque_err_date_invalid", "doc_date")
		}
		switch rule.date {
		case dateDue:
			add("r.due_date <= ?", d)
		case dateReceived:
			add("r.doc_date <= ?", d)
		}
	}
	if b := strings.TrimSpace(f.Book); b != "" && bookPolicyFor(f.Flag, f.CancelType).match {
		add("r.book_code = ?", b)
	}
	if s := strings.TrimSpace(f.Query); s != "" {
		add("(r.cheque_no ILIKE ? OR r.doc_no ILIKE ? OR r.partner_code ILIKE ?)", likePattern(s))
	}
	return pageRegisters(ctx, q, where, args, f.Page, f.Limit)
}

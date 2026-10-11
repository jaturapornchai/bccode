package cheque

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lib/pq"
)

// ชนิดรหัสทั้ง 5 ใน bank_codes (Champ: BCBank / BCBankBranch / BCBANKSERVICE / BCBANKACCOUNT / สถานะเช็คเพิ่มเติม)
const (
	KindBank            = "bank"
	KindBankBranch      = "bank_branch"
	KindBankAccountType = "bank_account_type"
	KindBankService     = "bank_service"
	KindChequeExtStatus = "cheque_ext_status"
)

// CodeKinds เรียงตามลำดับในเมนู "กำหนดรหัสต่างๆ" ของ Champ
var CodeKinds = []string{KindBank, KindBankBranch, KindBankService, KindBankAccountType, KindChequeExtStatus}

// ความยาวสูงสุดตาม Champ (Code VARCHAR(10), Name VARCHAR(50), UserGroup VARCHAR(255))
const (
	maxCodeLen      = 10
	maxNameLen      = 50
	maxUserGroupLen = 255
)

// ValidCodeKind บอกว่า kind เป็นหนึ่งใน 5 ชนิดรหัส
func ValidCodeKind(kind string) bool {
	for _, k := range CodeKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// Scope คือ holding/บริษัท/สาขา และผู้ทำรายการ ที่ผ่านการยืนยัน session แล้ว (ผู้เรียกต้องตรวจสิทธิ์ก่อน);
// Holding และ Branch ใช้ตอนลงบัญชีเอกสารเปลี่ยนสถานะเช็ค
type Scope struct {
	Holding string
	Company string
	Branch  string
	Actor   string
}

// Code คือหนึ่งแถวใน bank_codes
type Code struct {
	Kind      string `json:"kind"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	BankCode  string `json:"bank_code"`
	UserGroup string `json:"user_group"`
	IsActive  bool   `json:"is_active"`
	Version   int64  `json:"version"`
}

// CodeInput คือค่าที่ผู้ใช้กรอก (IsActive ว่าง = ใช้งาน เมื่อสร้าง / คงค่าเดิม เมื่อแก้)
type CodeInput struct {
	Kind      string
	Code      string
	Name      string
	BankCode  string
	UserGroup string
	IsActive  *bool
	Version   int64
}

// CodeFilter คือเงื่อนไขค้นหารายการรหัส; BankCode กรองเฉพาะรหัสสาขา (สาขาที่ไม่ระบุธนาคารใช้ได้กับทุกธนาคาร)
type CodeFilter struct {
	Query      string
	BankCode   string
	ActiveOnly bool
	Page       int
	Limit      int
}

const (
	defaultCodeLimit = 50
	maxCodeLimit     = 500
	codeColumns      = `kind, code, name, COALESCE(bank_code, ''), user_group, is_active, version`
)

// querier คือส่วนร่วมของ *sql.DB และ *sql.Tx ที่ใช้อ่านข้อมูล
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type rowScanner interface{ Scan(dest ...any) error }

func scanCode(row rowScanner) (Code, error) {
	var c Code
	err := row.Scan(&c.Kind, &c.Code, &c.Name, &c.BankCode, &c.UserGroup, &c.IsActive, &c.Version)
	return c, err
}

// normalizeCode ตัดช่องว่างหัวท้ายและตรวจความถูกต้องของรหัสที่กรอก (ข้อความผิดพลาดเป็น key ใน languages.tsv)
func normalizeCode(in CodeInput) (CodeInput, error) {
	in.Kind = strings.TrimSpace(in.Kind)
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)
	in.BankCode = strings.TrimSpace(in.BankCode)
	in.UserGroup = strings.TrimSpace(in.UserGroup)
	if !ValidCodeKind(in.Kind) {
		return in, badRequest("bank_code_err_kind_invalid", "kind", in.Kind)
	}
	if in.Code == "" {
		return in, badRequest("bank_code_err_code_required", "code")
	}
	if utf8.RuneCountInString(in.Code) > maxCodeLen {
		return in, badRequest("bank_code_err_code_too_long", "code", fmt.Sprint(maxCodeLen))
	}
	if strings.ContainsFunc(in.Code, unicode.IsControl) {
		return in, badRequest("bank_code_err_code_invalid", "code")
	}
	if in.Name == "" {
		return in, badRequest("bank_code_err_name_required", "name")
	}
	if utf8.RuneCountInString(in.Name) > maxNameLen {
		return in, badRequest("bank_code_err_name_too_long", "name", fmt.Sprint(maxNameLen))
	}
	if utf8.RuneCountInString(in.UserGroup) > maxUserGroupLen {
		return in, badRequest("bank_code_err_group_too_long", "user_group", fmt.Sprint(maxUserGroupLen))
	}
	if in.Kind != KindBankBranch {
		in.BankCode = "" // ตัวกรองธนาคารมีเฉพาะรหัสสาขา
	}
	return in, nil
}

// likePattern ทำให้คำค้นเป็นรูปแบบ ILIKE ที่ตัดความหมายพิเศษของ % _ \ ออก
func likePattern(q string) string {
	escaper := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + escaper.Replace(q) + "%"
}

func pageWindow(page, limit, defaultLimit, maxLimit int) (int, int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if page <= 0 {
		page = 1
	}
	return page, limit, (page - 1) * limit
}

// CodePage คือผลค้นหารายการรหัสหนึ่งหน้า (Total = จำนวนทั้งหมดที่ตรงเงื่อนไข)
type CodePage struct {
	Items []Code
	Total int
	Page  int
	Limit int
}

// ListCodes คืนรายการรหัสชนิดหนึ่งของบริษัท เรียงตามรหัส พร้อมจำนวนทั้งหมด (แบ่งหน้าด้วย LIMIT/OFFSET)
func ListCodes(ctx context.Context, db querier, company, kind string, f CodeFilter) (CodePage, error) {
	kind = strings.TrimSpace(kind)
	if !ValidCodeKind(kind) {
		return CodePage{}, badRequest("bank_code_err_kind_invalid", "kind", kind)
	}
	where := []string{"company = $1", "kind = $2"}
	args := []any{company, kind}
	if q := strings.TrimSpace(f.Query); q != "" {
		args = append(args, likePattern(q))
		where = append(where, fmt.Sprintf("(code ILIKE $%[1]d OR name ILIKE $%[1]d)", len(args)))
	}
	if bank := strings.TrimSpace(f.BankCode); bank != "" && kind == KindBankBranch {
		args = append(args, bank)
		where = append(where, fmt.Sprintf("(bank_code = $%d OR bank_code IS NULL)", len(args)))
	}
	if f.ActiveOnly {
		where = append(where, "is_active")
	}
	condition := strings.Join(where, " AND ")

	var total int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM bank_codes WHERE "+condition, args...).Scan(&total); err != nil {
		return CodePage{}, err
	}
	page, limit, offset := pageWindow(f.Page, f.Limit, defaultCodeLimit, maxCodeLimit)
	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx,
		fmt.Sprintf("SELECT %s FROM bank_codes WHERE %s ORDER BY code LIMIT $%d OFFSET $%d", codeColumns, condition, len(args)-1, len(args)), args...)
	if err != nil {
		return CodePage{}, err
	}
	defer rows.Close()
	items := []Code{}
	for rows.Next() {
		c, err := scanCode(rows)
		if err != nil {
			return CodePage{}, err
		}
		items = append(items, c)
	}
	return CodePage{Items: items, Total: total, Page: page, Limit: limit}, rows.Err()
}

// CodeExists บอกว่ารหัสชนิดนี้มีอยู่ในทะเบียนของบริษัท (ใช้ตรวจรหัสที่เอกสารอื่นอ้างถึง)
// FOR SHARE กันรหัสถูกลบระหว่างที่ธุรกรรมของผู้เรียกยังไม่จบ
func CodeExists(ctx context.Context, tx querier, company, kind, code string) (bool, error) {
	var one int
	err := tx.QueryRowContext(ctx,
		`SELECT 1 FROM bank_codes WHERE company = $1 AND kind = $2 AND code = $3 FOR SHARE`, company, kind, code).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// LookupCode อ่านรหัสหนึ่งตัวของบริษัท (ชื่อ ธนาคารของสาขา สถานะ) ให้ผู้เรียกที่ต้องตรวจรหัสกับทะเบียน เช่น สมุดเงินฝาก;
// FOR SHARE เหมือน CodeExists — ไม่พบ = ok false
func LookupCode(ctx context.Context, tx querier, company, kind, code string) (Code, bool, error) {
	c, err := scanCode(tx.QueryRowContext(ctx,
		`SELECT `+codeColumns+` FROM bank_codes WHERE company = $1 AND kind = $2 AND code = $3 FOR SHARE`, company, kind, code))
	if err == sql.ErrNoRows {
		return Code{}, false, nil
	}
	return c, err == nil, err
}

// requireBank ตรวจว่ารหัสธนาคารที่สาขาอ้างถึงมีอยู่จริง
func requireBank(ctx context.Context, tx querier, company, bankCode string) error {
	if bankCode == "" {
		return nil
	}
	ok, err := CodeExists(ctx, tx, company, KindBank, bankCode)
	if err != nil {
		return err
	}
	if !ok {
		return notFound("bank_code_err_not_found", "bank_code", bankCode)
	}
	return nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// CreateCode เพิ่มรหัสใหม่ (version 1); รหัสซ้ำ (ไม่แยกตัวพิมพ์ใหญ่เล็ก) = bank_code_err_duplicate
func CreateCode(ctx context.Context, db *sql.DB, scope Scope, in CodeInput) (Code, error) {
	in, err := normalizeCode(in)
	if err != nil {
		return Code{}, err
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Code{}, err
	}
	defer tx.Rollback()
	if err = requireBank(ctx, tx, scope.Company, in.BankCode); err != nil {
		return Code{}, err
	}
	created, err := scanCode(tx.QueryRowContext(ctx,
		`INSERT INTO bank_codes (company, kind, code, name, bank_code, user_group, is_active, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8) RETURNING `+codeColumns,
		scope.Company, in.Kind, in.Code, in.Name, nullable(in.BankCode), in.UserGroup, active, scope.Actor))
	if err != nil {
		if isUniqueViolation(err) {
			return Code{}, newError(409, "bank_code_err_duplicate", "code", in.Code)
		}
		return Code{}, err
	}
	return created, tx.Commit()
}

// lockCode อ่านแถวพร้อมล็อก (FOR UPDATE) แล้วตรวจ version; ไม่พบ = 404, version ไม่ตรง = 409
func lockCode(ctx context.Context, tx querier, company, kind, code string, version int64) (Code, error) {
	current, err := scanCode(tx.QueryRowContext(ctx,
		`SELECT `+codeColumns+` FROM bank_codes WHERE company = $1 AND kind = $2 AND code = $3 FOR UPDATE`, company, kind, code))
	if err == sql.ErrNoRows {
		return Code{}, notFound("bank_code_err_not_found", "code", code)
	}
	if err != nil {
		return Code{}, err
	}
	if current.Version != version {
		return Code{}, versionConflict()
	}
	return current, nil
}

// UpdateCode แก้ชื่อ/กลุ่มผู้ใช้/ธนาคารของสาขา/สถานะใช้งาน (รหัสและชนิดเป็นคีย์ แก้ไม่ได้) ด้วย optimistic lock
func UpdateCode(ctx context.Context, db *sql.DB, scope Scope, in CodeInput) (Code, error) {
	in, err := normalizeCode(in)
	if err != nil {
		return Code{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Code{}, err
	}
	defer tx.Rollback()
	current, err := lockCode(ctx, tx, scope.Company, in.Kind, in.Code, in.Version)
	if err != nil {
		return Code{}, err
	}
	if err = requireBank(ctx, tx, scope.Company, in.BankCode); err != nil {
		return Code{}, err
	}
	active := current.IsActive
	if in.IsActive != nil {
		active = *in.IsActive
	}
	updated, err := scanCode(tx.QueryRowContext(ctx,
		`UPDATE bank_codes SET name = $4, bank_code = $5, user_group = $6, is_active = $7,
		        version = version + 1, updated_at = now(), updated_by = $8
		 WHERE company = $1 AND kind = $2 AND code = $3 RETURNING `+codeColumns,
		scope.Company, in.Kind, in.Code, in.Name, nullable(in.BankCode), in.UserGroup, active, scope.Actor))
	if err != nil {
		return Code{}, err
	}
	return updated, tx.Commit()
}

// DeleteCode ลบรหัสที่ยังไม่มีข้อมูลอื่นอ้างถึง; ถูกใช้แล้ว = bank_code_err_in_use (ให้ปิดการใช้งานแทน)
func DeleteCode(ctx context.Context, db *sql.DB, scope Scope, in CodeInput) (Code, error) {
	in.Kind, in.Code = strings.TrimSpace(in.Kind), strings.TrimSpace(in.Code)
	if !ValidCodeKind(in.Kind) {
		return Code{}, badRequest("bank_code_err_kind_invalid", "kind", in.Kind)
	}
	if in.Code == "" {
		return Code{}, badRequest("bank_code_err_code_required", "code")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Code{}, err
	}
	defer tx.Rollback()
	current, err := lockCode(ctx, tx, scope.Company, in.Kind, in.Code, in.Version)
	if err != nil {
		return Code{}, err
	}
	used, err := codeInUse(ctx, tx, scope.Company, in.Kind, in.Code)
	if err != nil {
		return Code{}, err
	}
	if used {
		return Code{}, newError(409, "bank_code_err_in_use", "code", in.Code)
	}
	if _, err = tx.ExecContext(ctx,
		`DELETE FROM bank_codes WHERE company = $1 AND kind = $2 AND code = $3`, scope.Company, in.Kind, in.Code); err != nil {
		return Code{}, err
	}
	return current, tx.Commit()
}

// usage คือที่ที่รหัสถูกอ้างถึง: ตาราง + เงื่อนไข (company = $1 เสมอ, รหัส = $2) — เป็นค่าคงที่ ไม่มีข้อมูลผู้ใช้ปนใน SQL
type usage struct{ table, condition string }

// codeUsage ครอบคลุมตารางที่มีอยู่แล้วและตารางของงานเช็คที่จะตามมา (cheque-bank.md §4.2);
// ตารางที่ยังไม่ถูกสร้างในฐานนั้นข้ามได้ (to_regclass) จึงไม่ต้องแก้ฟังก์ชันนี้เมื่อตารางเกิดทีหลัง
var codeUsage = map[string][]usage{
	KindBank: {
		{"bank_codes", "kind = 'bank_branch' AND bank_code = $2"},
		{"bank_contacts", "bank_code = $2"},
		{"gl_subledger_bank_accounts", "payload->>'bank_code' = $2"},
		{"cheque_registers", "bank_code = $2 AND deleted_at IS NULL"},
		{"cheque_prepare_docs", "bank_code = $2"},
		{"bank_service_notes", "bank_code = $2"},
		{"gl_journal_cheques", "bank_code = $2"},
	},
	KindBankBranch: {
		{"bank_contacts", "bank_branch_code = $2"},
		{"gl_subledger_bank_accounts", "payload->>'bank_branch_code' = $2"},
		{"cheque_registers", "bank_branch_code = $2 AND deleted_at IS NULL"},
		{"cheque_prepare_docs", "bank_branch_code = $2"},
		{"gl_journal_cheques", "bank_branch_code = $2"},
	},
	KindBankAccountType: {
		{"gl_subledger_bank_accounts", "payload->>'account_type_code' = $2"},
	},
	KindChequeExtStatus: {
		{"cheque_registers", "ext_status_code = $2 AND deleted_at IS NULL"},
	},
}

// codeInUse บอกว่ามีข้อมูลของบริษัทนี้อ้างถึงรหัสนั้นอยู่หรือไม่ (ไม่ได้ผูก FK เพราะบางที่อยู่ใน jsonb และข้ามโมดูล)
func codeInUse(ctx context.Context, tx querier, company, kind, code string) (bool, error) {
	probes := codeUsage[kind]
	if len(probes) == 0 {
		return false, nil
	}
	tables := make([]string, len(probes))
	for i, p := range probes {
		tables[i] = p.table
	}
	rows, err := tx.QueryContext(ctx, `SELECT t FROM unnest($1::text[]) AS t WHERE to_regclass(t) IS NOT NULL`, pq.Array(tables))
	if err != nil {
		return false, err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			rows.Close()
			return false, err
		}
		existing[table] = true
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return false, err
	}
	rows.Close()

	checks := []string{}
	for _, p := range probes {
		if existing[p.table] {
			checks = append(checks, fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE company = $1 AND %s)", p.table, p.condition))
		}
	}
	if len(checks) == 0 {
		return false, nil
	}
	var used bool
	err = tx.QueryRowContext(ctx, "SELECT "+strings.Join(checks, " OR "), company, code).Scan(&used)
	return used, err
}

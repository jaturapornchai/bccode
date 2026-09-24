package generalledger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/shopspring/decimal"
	"golang.org/x/text/unicode/norm"
)

type Scope struct {
	Holding string
	Company string
	Branch  string
	Actor   string
}

type Name struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
}

type Identity struct {
	ID           string    `json:"id" bson:"_id"`
	HoldingCode  string    `json:"holdingcode" bson:"holdingcode"`
	BusinessCode string    `json:"businesscode" bson:"businesscode"`
	Version      int64     `json:"version" bson:"__v"`
	CreatedAt    time.Time `json:"createdat" bson:"createdat"`
	CreatedBy    string    `json:"createdby" bson:"createdby"`
	UpdatedAt    time.Time `json:"updatedat" bson:"updatedat"`
	UpdatedBy    string    `json:"updatedby" bson:"updatedby"`
	IsDeleted    bool      `json:"isdeleted" bson:"isdeleted"`
}

type Account struct {
	Identity          `bson:",inline"`
	AccountCode       string `json:"accountcode" bson:"accountcode"`
	Names             []Name `json:"names" bson:"names"`
	AccountType       string `json:"accounttype" bson:"accounttype"`
	ParentAccountCode string `json:"parentaccountcode" bson:"parentaccountcode"`
	NormalBalance     string `json:"normalbalance" bson:"normalbalance"`
	AllowPosting      bool   `json:"allowposting" bson:"allowposting"`
	IsActive          bool   `json:"isactive" bson:"isactive"`
	AccountGroup      string `json:"accountgroup" bson:"accountgroup"`
	IsCash            bool   `json:"iscash" bson:"iscash"`
	Level             int    `json:"level" bson:"level"`
}

func (Account) CollectionName() string { return "chart_of_accounts" }
func (a Account) ThaiName() string {
	for _, n := range a.Names {
		if n.Code == "th" {
			return n.Name
		}
	}
	return ""
}

func (a Account) MarshalJSON() ([]byte, error) {
	type Alias Account
	var parent *string
	if strings.TrimSpace(a.ParentAccountCode) != "" {
		p := strings.TrimSpace(a.ParentAccountCode)
		parent = &p
	}
	return json.Marshal(&struct {
		Alias
		ParentAccountCode *string `json:"parentaccountcode"`
	}{
		Alias:             Alias(a),
		ParentAccountCode: parent,
	})
}

type FiscalYear struct {
	Identity                `bson:",inline"`
	Code                    string `json:"code" bson:"code"`
	StartDate               string `json:"startdate" bson:"startdate"`
	EndDate                 string `json:"enddate" bson:"enddate"`
	IsActive                bool   `json:"isactive" bson:"isactive"`
	Scale                   int    `json:"scale" bson:"scale"`
	ProfitLossAccount       string `json:"profitlossaccount" bson:"profitlossaccount"`
	RetainedEarningsAccount string `json:"retainedearningsaccount" bson:"retainedearningsaccount"`
	Closed                  bool   `json:"closed" bson:"closed"`
}

func (FiscalYear) CollectionName() string { return "fiscal_year" }

// AllocationRule splits one cost allocation code across target branches,
// departments, projects and accounts by a fixed percentage. Rates are exact
// decimals (percentage, 0–100) and must total exactly 100 within a code.
type AllocationRule struct {
	BranchCode     string `json:"branchcode,omitempty" bson:"branchcode,omitempty"`
	DepartmentCode string `json:"departmentcode,omitempty" bson:"departmentcode,omitempty"`
	ProjectCode    string `json:"projectcode,omitempty" bson:"projectcode,omitempty"`
	AccountCode    string `json:"accountcode,omitempty" bson:"accountcode,omitempty"`
	Rate           Amount `json:"rate" bson:"rate"`
}

// Auxiliary masters have separate collections. Kind is an API discriminator,
// never a client-supplied collection or table name.
type Master struct {
	Identity       `bson:",inline"`
	Kind           string                `json:"kind" bson:"-"`
	Code           string                `json:"code" bson:"code"`
	Name           string                `json:"name" bson:"name"`
	NameEn         string                `json:"nameen,omitempty" bson:"nameen,omitempty"`     // journal-books: English name
	BookType       int                   `json:"booktype,omitempty" bson:"booktype,omitempty"` // journal-books: 1 ทั่วไป 2 จ่าย 3 รับ 4 ขาย 5 ซื้อ 6 ยอดยกมา
	IsActive       bool                  `json:"isactive" bson:"isactive"`
	AccountCode    string                `json:"accountcode,omitempty" bson:"accountcode,omitempty"`
	FiscalYear     string                `json:"fiscalyear,omitempty" bson:"fiscalyear,omitempty"`
	StartDate      string                `json:"startdate,omitempty" bson:"startdate,omitempty"`
	EndDate        string                `json:"enddate,omitempty" bson:"enddate,omitempty"`
	Locked         bool                  `json:"locked" bson:"locked"`
	Amount         Amount                `json:"amount" bson:"amount"`
	BranchCode     string                `json:"branchcode,omitempty" bson:"branchcode,omitempty"`
	DepartmentCode string                `json:"departmentcode,omitempty" bson:"departmentcode,omitempty"`
	ProjectCode    string                `json:"projectcode,omitempty" bson:"projectcode,omitempty"`
	Direction      string                `json:"direction,omitempty" bson:"direction,omitempty"`
	BookCode       string                `json:"bookcode,omitempty" bson:"bookcode,omitempty"`
	ItemAccount    string                `json:"itemaccount,omitempty" bson:"itemaccount,omitempty"`
	CostAccount    string                `json:"costaccount,omitempty" bson:"costaccount,omitempty"`
	RevenueAccount string                `json:"revenueaccount,omitempty" bson:"revenueaccount,omitempty"`
	Rules          []MappingRule         `json:"rules,omitempty" bson:"rules,omitempty"`
	AllocateMode   string                `json:"allocatemode,omitempty" bson:"allocatemode,omitempty"`
	AllocateRules  []AllocationRule      `json:"allocaterules,omitempty" bson:"allocaterules,omitempty"`
	StatementType  string                `json:"statementtype,omitempty" bson:"statementtype,omitempty"`
	GlobalStyle    *StatementGlobalStyle `json:"globalstyle,omitempty" bson:"globalstyle,omitempty"`
	Rows           []StatementRow        `json:"rows,omitempty" bson:"rows,omitempty"`
}

type StatementStyle struct {
	FontFamily string `json:"fontfamily,omitempty" bson:"fontfamily,omitempty"`
	FontSize   string `json:"fontsize,omitempty" bson:"fontsize,omitempty"`
	FontWeight string `json:"fontweight,omitempty" bson:"fontweight,omitempty"`
	FontStyle  string `json:"fontstyle,omitempty" bson:"fontstyle,omitempty"`
	Align      string `json:"align,omitempty" bson:"align,omitempty"`
	Color      string `json:"color,omitempty" bson:"color,omitempty"`
	Indent     int    `json:"indent,omitempty" bson:"indent,omitempty"`
	Underline  string `json:"underline,omitempty" bson:"underline,omitempty"`
}

type StatementRow struct {
	ID            string         `json:"id" bson:"id"`
	RowNo         int            `json:"rowno" bson:"rowno"`
	RowType       string         `json:"rowtype" bson:"rowtype"`
	Title         string         `json:"title" bson:"title"`
	NoteNo        string         `json:"noteno,omitempty" bson:"noteno,omitempty"`
	AccountCodes  []string       `json:"accountcodes,omitempty" bson:"accountcodes,omitempty"`
	AccountGroup  string         `json:"accountgroup,omitempty" bson:"accountgroup,omitempty"`
	NormalBalance string         `json:"normalbalance,omitempty" bson:"normalbalance,omitempty"`
	Formula       string         `json:"formula,omitempty" bson:"formula,omitempty"`
	ReverseSign   bool           `json:"reversesign,omitempty" bson:"reversesign,omitempty"`
	ShowZero      bool           `json:"showzero,omitempty" bson:"showzero,omitempty"`
	Style         StatementStyle `json:"style,omitempty" bson:"style,omitempty"`
}

type StatementGlobalStyle struct {
	FontFamily     string `json:"fontfamily,omitempty" bson:"fontfamily,omitempty"`
	FontSize       string `json:"fontsize,omitempty" bson:"fontsize,omitempty"`
	Scale          int    `json:"scale,omitempty" bson:"scale,omitempty"`
	Compact        bool   `json:"compact,omitempty" bson:"compact,omitempty"`
	ShowNoteColumn bool   `json:"shownotecolumn,omitempty" bson:"shownotecolumn,omitempty"`
	ComparisonType string `json:"comparisontype,omitempty" bson:"comparisontype,omitempty"`
}

var MasterCollections = map[string]string{
	"account-groups": "gl_account_groups", "product-account-groups": "gl_product_account_groups",
	"mappings": "gl_account_mappings", "budgets": "gl_budgets", "periods": "gl_periods", "forecast": "gl_cash_forecast",
	"allocations":         "gl_allocations",
	"statement-templates": "gl_statement_templates",
	"journal-books":       "gl_journal_books",
}

func (m Master) CollectionName() string { return MasterCollections[m.Kind] }

type MappingRule struct {
	AccountCode string `json:"accountcode" bson:"accountcode"`
	Side        string `json:"side" bson:"side"`
	Source      string `json:"source" bson:"source"`
}

type Line struct {
	AccountCode    string `json:"accountcode" bson:"accountcode"`
	AccountName    string `json:"accountname,omitempty" bson:"accountname,omitempty"`
	Description    string `json:"description" bson:"description"`
	Debit          Amount `json:"debit" bson:"debit"`
	Credit         Amount `json:"credit" bson:"credit"`
	DepartmentCode string `json:"departmentcode" bson:"departmentcode"`
	ProjectCode    string `json:"projectcode" bson:"projectcode"`
	CashFlow       string `json:"cashflow" bson:"cashflow"`
}

type Journal struct {
	Details        *JournalDetails `json:"details,omitempty" bson:"details,omitempty"`
	SourceType     int             `json:"source_type,omitempty" bson:"source_type,omitempty"`
	SourceSystem   string          `json:"source_system,omitempty" bson:"source_system,omitempty"`
	SourceRecordID string          `json:"source_record_id,omitempty" bson:"source_record_id,omitempty"`
	Identity       `bson:",inline"`
	DocNo          string     `json:"docno" bson:"docno"`
	Date           string     `json:"date" bson:"date"`
	BookCode       string     `json:"bookcode" bson:"bookcode"`
	FiscalYear     string     `json:"fiscalyear" bson:"fiscalyear"`
	Description    string     `json:"description" bson:"description"`
	Reference      string     `json:"reference" bson:"reference"`
	BranchCode     string     `json:"branchcode" bson:"branchcode"`
	Kind           string     `json:"kind" bson:"kind"`
	Status         string     `json:"status" bson:"status"`
	Lines          []Line     `json:"lines" bson:"lines"`
	ReversalOf     string     `json:"reversalof,omitempty" bson:"reversalof,omitempty"`
	PostedAt       *time.Time `json:"postedat,omitempty" bson:"postedat,omitempty"`
	PostedBy       string     `json:"postedby,omitempty" bson:"postedby,omitempty"`
	Reason         string     `json:"reason,omitempty" bson:"reason,omitempty"`
}

func (Journal) CollectionName() string { return "gl_journals" }

type Command struct {
	Review     *ReviewInput `json:"review,omitempty"`
	TargetYear string       `json:"targetyear,omitempty"`
	prepared   *preparedProcess
	Resource   string      `json:"resource"`
	ID         string      `json:"id"`
	Action     string      `json:"action"`
	RequestID  string      `json:"requestid"`
	Version    int64       `json:"version"`
	Reason     string      `json:"reason"`
	Date       string      `json:"date"`
	DocNo      string      `json:"docno"`
	Account    *Account    `json:"account,omitempty"`
	FiscalYear *FiscalYear `json:"fiscalyear,omitempty"`
	Master     *Master     `json:"master,omitempty"`
	Journal    *Journal    `json:"journal,omitempty"`
}

// Code lengths are counted in runes (PostgreSQL VARCHAR(n) counts characters) and follow
// mydocs/datamodels/gl: journal_entries.doc_no VARCHAR(30), chart_of_accounts.account_code
// VARCHAR(20), journal_books.code VARCHAR(15). Codes without a column there keep the old 60.
const (
	DocNoMaxRunes        = 30
	AccountCodeMaxRunes  = 20
	BookCodeMaxRunes     = 15
	codeMaxRunes         = 60
	accountNameMaxRunes  = 150 // chart_of_accounts.name_th / name_en VARCHAR(150)
	accountNameLanguages = 12
)

// Punctuation allowed after the first character of a code. Space, | " ' , ; and every
// control/format character (NUL, zero-width, bidi marks) are rejected.
const codePunctuation = "_.-/#():"

// NormalizeCode trims and NFC-normalises a code so a pasted trailing space or a decomposed
// character cannot create a second, visually identical code. NFC does not reorder a Thai
// tone mark against an above vowel (both combining class 0), so such typos stay distinct.
func NormalizeCode(code string) string { return norm.NFC.String(strings.TrimSpace(code)) }

// checkCode validates one code field and names the field and the fix in Thai.
// Thai vowel signs and tone marks are Unicode combining marks (\p{M}), so they are allowed
// after the first character; the first character must be a letter or a digit.
func checkCode(value, field, label string, max int) error {
	if value == "" {
		return fieldError("code_required", field, "กรุณาระบุ"+label)
	}
	if !utf8.ValidString(value) {
		return fieldError("code_invisible_char", field, label+"มีอักขระที่อ่านไม่ได้ปนอยู่ (มักติดมาจากการคัดลอก) กรุณาลบแล้วพิมพ์ใหม่")
	}
	for i, r := range value {
		switch {
		case unicode.IsSpace(r):
			return fieldError("code_has_space", field, label+"ห้ามมีช่องว่าง กรุณาลบช่องว่างออก หรือใช้ - หรือ _ คั่นแทน")
		case unicode.Is(unicode.C, r):
			return fieldError("code_invisible_char", field, label+"มีอักขระที่มองไม่เห็นปนอยู่ (มักติดมาจากการคัดลอก) กรุณาลบแล้วพิมพ์ใหม่")
		case i == 0 && !unicode.IsLetter(r) && !unicode.IsNumber(r):
			return fieldError("code_invalid_start", field, label+"ต้องขึ้นต้นด้วยตัวอักษรหรือตัวเลข ห้ามขึ้นต้นด้วยสระ วรรณยุกต์ หรือเครื่องหมาย")
		case unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsNumber(r) || strings.ContainsRune(codePunctuation, r):
		default:
			return fieldError("code_invalid_char", field, fmt.Sprintf("%sมีอักขระ “%s” ที่ใช้ไม่ได้ ใช้ได้เฉพาะตัวอักษร ตัวเลข และเครื่องหมาย _ . - / # ( ) :", label, string(r)))
		}
	}
	if n := utf8.RuneCountInString(value); n > max {
		return fieldError("code_too_long", field, fmt.Sprintf("%sยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว) กรุณาย่อให้สั้นลง", label, max, n))
	}
	return nil
}

// CheckDocNo validates a journal document number exactly as Journal.Validate does, so a module
// that generates voucher numbers (fixed assets) can explain a failure before it posts anything.
func CheckDocNo(docNo, field string) error {
	return checkCode(docNo, field, "เลขที่เอกสาร", DocNoMaxRunes)
}

// validCode is the boolean form for callers that report their own message.
func validCode(code string) bool { return checkCode(code, "", "", codeMaxRunes) == nil }

// validRequestID keeps the idempotency key contract: 16–80 bytes of code characters
// (UUIDs, and the 64-hex-character digests the fixed-asset poster sends).
func validRequestID(id string) bool {
	return len(id) >= 16 && len(id) <= 80 && checkCode(id, "", "", 80) == nil
}

func validDate(value string) bool {
	t, err := time.Parse("2006-01-02", value)
	return err == nil && t.Format("2006-01-02") == value
}

func (a Account) Validate() error {
	if err := checkCode(a.AccountCode, "accountcode", "รหัสบัญชี", AccountCodeMaxRunes); err != nil {
		return err
	}
	if len(a.Names) > accountNameLanguages {
		return fieldError("account_names_too_many", "names", fmt.Sprintf("ชื่อบัญชีรองรับไม่เกิน %d ภาษา กรุณาลบชื่อภาษาที่ไม่ใช้ออก", accountNameLanguages))
	}
	if strings.TrimSpace(a.ThaiName()) == "" {
		return fieldError("account_name_th_required", "names", "กรุณาระบุชื่อบัญชีภาษาไทย")
	}
	if !contains([]string{"asset", "liability", "equity", "income", "expense"}, a.AccountType) {
		return fmt.Errorf("กรุณาเลือกหมวดบัญชี")
	}
	if !contains([]string{"debit", "credit"}, a.NormalBalance) {
		return fmt.Errorf("กรุณาเลือกด้านบัญชี")
	}
	if a.ParentAccountCode == a.AccountCode {
		return fmt.Errorf("บัญชีแม่ต้องไม่เป็นบัญชีเดียวกัน")
	}
	if a.IsCash && (a.AccountType != "asset" || !a.AllowPosting) {
		return fmt.Errorf("บัญชีเงินสดต้องเป็นบัญชีสินทรัพย์ที่ลงรายการได้")
	}
	if a.Level < 0 || a.Level > 12 {
		return fmt.Errorf("ระดับบัญชีต้องอยู่ระหว่าง 1 ถึง 12")
	}
	for _, n := range a.Names {
		if utf8.RuneCountInString(n.Name) > accountNameMaxRunes {
			return fieldError("account_name_too_long", "names", fmt.Sprintf("ชื่อบัญชียาวได้ไม่เกิน %d ตัวอักษร กรุณาย่อชื่อให้สั้นลง", accountNameMaxRunes))
		}
	}
	return nil
}

func (f FiscalYear) Validate() error {
	if err := checkCode(f.Code, "code", "รหัสปีบัญชี", codeMaxRunes); err != nil {
		return err
	}
	switch {
	case !validDate(f.StartDate):
		return fieldError("fiscal_year_start_invalid", "startdate", "วันเริ่มต้นปีบัญชีไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	case !validDate(f.EndDate):
		return fieldError("fiscal_year_end_invalid", "enddate", "วันสิ้นสุดปีบัญชีไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	case f.StartDate > f.EndDate:
		return fieldError("fiscal_year_range_invalid", "enddate", "วันสิ้นสุดปีบัญชีต้องไม่ก่อนวันเริ่มต้น กรุณาแก้วันที่ให้ถูกต้อง")
	}
	if f.Scale < 0 || f.Scale > 8 {
		return fmt.Errorf("กรุณากำหนดจำนวนทศนิยม 0–8 ตำแหน่ง")
	}
	return nil
}

// Free-text lengths follow mydocs/datamodels/gl: journal_entries.description VARCHAR(500),
// journal_entries.ref_doc_no VARCHAR(50), journal_lines.description VARCHAR(255) and
// journal_entries.reversal_reason VARCHAR(500) — counted in runes like PostgreSQL VARCHAR(n).
const (
	JournalDescriptionMaxRunes     = 500
	JournalReferenceMaxRunes       = 50
	JournalLineDescriptionMaxRunes = 255
	ReversalReasonMaxRunes         = 500
)

// checkJournalTextLimits checks only text that is new or changed (old = the stored draft, nil on
// create): a draft saved before these limits existed stays editable as long as its long text is untouched.
func checkJournalTextLimits(next Journal, old *Journal) error {
	if (old == nil || next.Description != old.Description) && utf8.RuneCountInString(strings.TrimSpace(next.Description)) > JournalDescriptionMaxRunes {
		return fieldError("journal_description_too_long", "description", fmt.Sprintf("คำอธิบายรายการต้องไม่เกิน %d ตัวอักษร — กรุณาย่อให้สั้นลง แล้วใส่รายละเอียดที่เหลือในคำอธิบายบรรทัด", JournalDescriptionMaxRunes))
	}
	if (old == nil || next.Reference != old.Reference) && utf8.RuneCountInString(strings.TrimSpace(next.Reference)) > JournalReferenceMaxRunes {
		return fieldError("journal_reference_too_long", "reference", fmt.Sprintf("เลขที่อ้างอิงต้องไม่เกิน %d ตัวอักษร — กรุณาใส่เฉพาะเลขที่เอกสารอ้างอิง", JournalReferenceMaxRunes))
	}
	previous := map[string]bool{}
	if old != nil {
		for _, l := range old.Lines {
			previous[l.Description] = true
		}
	}
	for i, l := range next.Lines {
		if !previous[l.Description] && utf8.RuneCountInString(strings.TrimSpace(l.Description)) > JournalLineDescriptionMaxRunes {
			return fieldError("journal_line_description_too_long", fmt.Sprintf("lines[%d].description", i), fmt.Sprintf("บรรทัดที่ %d: คำอธิบายบรรทัดต้องไม่เกิน %d ตัวอักษร — กรุณาย่อให้สั้นลง", i+1, JournalLineDescriptionMaxRunes))
		}
	}
	return nil
}

// truncateRunes cuts generated text (reversal descriptions built from the original's text) to max runes.
func truncateRunes(text string, max int) string {
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	return string([]rune(text)[:max])
}

func (j Journal) Validate(f FiscalYear, accounts map[string]Account) error {
	if err := checkCode(j.DocNo, "docno", "เลขที่เอกสาร", DocNoMaxRunes); err != nil {
		return err
	}
	if !validDate(j.Date) {
		return fieldError("journal_date_invalid", "date", "วันที่เอกสารไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	}
	if strings.TrimSpace(j.Description) == "" {
		return fieldError("journal_description_required", "description", "กรุณาระบุคำอธิบายรายการ")
	}
	// bookcode is checked against the company's journal-books master by the store
	// (checkJournalBook) — books are user-defined, so there is no fixed list here.
	if !contains([]string{"manual", "opening", "closing", "reversal", "mapping"}, j.Kind) {
		return fmt.Errorf("ประเภทรายการไม่ถูกต้อง")
	}
	if j.FiscalYear != f.Code || j.Date < f.StartDate || j.Date > f.EndDate || f.Closed || !f.IsActive {
		return fmt.Errorf("วันที่อยู่นอกปีบัญชีที่เปิดใช้งาน")
	}
	// Closing/opening journals are generated per branch from every non-zero
	// (account, department, project) balance, so they follow the process limit.
	maxLines := 500
	if j.Kind == "closing" || j.Kind == "opening" {
		maxLines = ProcessBalanceRowLimit
	}
	if len(j.Lines) < 2 || len(j.Lines) > maxLines {
		return fmt.Errorf("รายการบัญชีต้องมี 2–%d บรรทัด", maxLines)
	}
	debit, credit := decimal.Zero, decimal.Zero
	for index, line := range j.Lines {
		a, ok := accounts[line.AccountCode]
		if !ok || !a.IsActive || a.IsDeleted || !a.AllowPosting {
			return fmt.Errorf("บรรทัด %d: กรุณาเลือกบัญชีที่เปิดใช้งานและลงรายการได้", index+1)
		}
		if err := line.Debit.ValidateScale(f.Scale); err != nil {
			return err
		}
		if err := line.Credit.ValidateScale(f.Scale); err != nil {
			return err
		}
		d, c := line.Debit.Decimal(), line.Credit.Decimal()
		if d.IsNegative() || c.IsNegative() || (d.IsZero() == c.IsZero()) {
			return fmt.Errorf("บรรทัด %d: ระบุจำนวนเงินมากกว่าศูนย์ด้านเดบิตหรือเครดิตเพียงด้านเดียว", index+1)
		}
		if line.CashFlow != "" && !contains([]string{"operating", "investing", "financing"}, line.CashFlow) {
			return fmt.Errorf("ประเภทกระแสเงินสดไม่ถูกต้อง")
		}
		debit, credit = debit.Add(d), credit.Add(c)
	}
	if !debit.Equal(credit) {
		return fmt.Errorf("ยอดเดบิตและเครดิตไม่เท่ากัน ต่างกัน %s", debit.Sub(credit).Abs().StringFixed(int32(f.Scale)))
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// normalizeCommandCodes trims and NFC-normalises every code the ledger compares or keys on,
// before the request hash, on copies so the caller's structs are left untouched.
func normalizeCommandCodes(cmd *Command) {
	cmd.DocNo = NormalizeCode(cmd.DocNo)
	cmd.TargetYear = NormalizeCode(cmd.TargetYear)
	if cmd.Account != nil {
		a := *cmd.Account
		a.AccountCode = NormalizeCode(a.AccountCode)
		a.ParentAccountCode = NormalizeCode(a.ParentAccountCode)
		a.AccountGroup = NormalizeCode(a.AccountGroup)
		cmd.Account = &a
	}
	if cmd.FiscalYear != nil {
		f := *cmd.FiscalYear
		f.Code = NormalizeCode(f.Code)
		f.ProfitLossAccount = NormalizeCode(f.ProfitLossAccount)
		f.RetainedEarningsAccount = NormalizeCode(f.RetainedEarningsAccount)
		cmd.FiscalYear = &f
	}
	if cmd.Master != nil {
		m := *cmd.Master
		for _, code := range []*string{&m.Code, &m.AccountCode, &m.FiscalYear, &m.BranchCode, &m.DepartmentCode, &m.ProjectCode, &m.BookCode, &m.ItemAccount, &m.CostAccount, &m.RevenueAccount} {
			*code = NormalizeCode(*code)
		}
		cmd.Master = &m
	}
	if cmd.Journal != nil {
		j := *cmd.Journal
		j.DocNo = NormalizeCode(j.DocNo)
		j.BookCode = NormalizeCode(j.BookCode)
		j.FiscalYear = NormalizeCode(j.FiscalYear)
		j.BranchCode = NormalizeCode(j.BranchCode)
		j.Lines = append([]Line(nil), j.Lines...)
		for i := range j.Lines {
			j.Lines[i].AccountCode = NormalizeCode(j.Lines[i].AccountCode)
			j.Lines[i].DepartmentCode = NormalizeCode(j.Lines[i].DepartmentCode)
			j.Lines[i].ProjectCode = NormalizeCode(j.Lines[i].ProjectCode)
		}
		cmd.Journal = &j
	}
}

// rejectNUL refuses a command carrying U+0000 anywhere: PostgreSQL jsonb cannot store it,
// so without this check the INSERT failed as a generic 500 (text pasted from PDFs carries it).
func rejectNUL(cmd Command) error {
	path := nulPath(reflect.ValueOf(cmd), "")
	if path == "" {
		return nil
	}
	// Field is relative to the record the screen edits (journal.lines[1].description → lines[1].description).
	for _, record := range []string{"journal.", "account.", "master.", "fiscalyear.", "review."} {
		path = strings.TrimPrefix(path, record)
	}
	return fieldError("text_contains_nul", path, nulFieldLabel(path)+"มีอักขระที่มองไม่เห็น (NUL) ซึ่งมักติดมาจากการคัดลอกจาก PDF หรือโปรแกรมอื่น กรุณาลบข้อความในช่องนั้นแล้วพิมพ์ใหม่")
}

// nulPath returns the JSON path of the first string holding U+0000, or "".
func nulPath(v reflect.Value, path string) string {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return ""
		}
		return nulPath(v.Elem(), path)
	case reflect.String:
		if strings.IndexByte(v.String(), 0) >= 0 {
			return path
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name, _, _ := strings.Cut(tag, ",")
			child := path
			if !(f.Anonymous && name == "") {
				if name == "" {
					name = f.Name
				}
				child = strings.TrimPrefix(path+"."+name, ".")
			}
			if found := nulPath(v.Field(i), child); found != "" {
				return found
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if v.Kind() == reflect.Slice && bytes.IndexByte(v.Bytes(), 0) >= 0 {
				return path
			}
			return ""
		}
		for i := 0; i < v.Len(); i++ {
			if found := nulPath(v.Index(i), fmt.Sprintf("%s[%d]", path, i)); found != "" {
				return found
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if found := nulPath(iter.Value(), fmt.Sprintf("%s.%v", path, iter.Key())); found != "" {
				return found
			}
		}
	}
	return ""
}

var nulFieldLabels = map[string]string{
	"description": "คำอธิบาย", "reference": "เอกสารอ้างอิง", "reason": "เหตุผล", "name": "ชื่อ",
	"nameen": "ชื่อภาษาอังกฤษ", "docno": "เลขที่เอกสาร", "code": "รหัส", "accountcode": "รหัสบัญชี",
	"bookcode": "สมุดรายวัน", "branchcode": "สาขา", "remark": "หมายเหตุ",
}

// nulFieldLabel turns lines[2].description into "บรรทัดที่ 3 คำอธิบาย" for the Thai message.
func nulFieldLabel(path string) string {
	prefix := ""
	if rest, ok := strings.CutPrefix(path, "lines["); ok {
		if index, _, found := strings.Cut(rest, "]"); found {
			if n, err := strconv.Atoi(index); err == nil {
				prefix = fmt.Sprintf("บรรทัดที่ %d ", n+1)
			}
		}
	}
	last := path[strings.LastIndex(path, ".")+1:]
	if label, ok := nulFieldLabels[last]; ok {
		return prefix + label
	}
	return prefix + "ข้อความที่กรอก"
}

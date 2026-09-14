package generalledger

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
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

type FiscalYear struct {
	Identity                `bson:",inline"`
	Code                    string `json:"code" bson:"code"`
	StartDate               string `json:"startdate" bson:"startdate"`
	EndDate                 string `json:"enddate" bson:"enddate"`
	IsActive                bool   `json:"isactive" bson:"isactive"`
	Currency                string `json:"currency" bson:"currency"`
	Scale                   int    `json:"scale" bson:"scale"`
	ProfitLossAccount       string `json:"profitlossaccount" bson:"profitlossaccount"`
	RetainedEarningsAccount string `json:"retainedearningsaccount" bson:"retainedearningsaccount"`
	Closed                  bool   `json:"closed" bson:"closed"`
}

func (FiscalYear) CollectionName() string { return "fiscal_year" }

// Auxiliary masters have separate collections. Kind is an API discriminator,
// never a client-supplied collection or table name.
type Master struct {
	Identity       `bson:",inline"`
	Kind           string        `json:"kind" bson:"-"`
	Code           string        `json:"code" bson:"code"`
	Name           string        `json:"name" bson:"name"`
	IsActive       bool          `json:"isactive" bson:"isactive"`
	AccountCode    string        `json:"accountcode,omitempty" bson:"accountcode,omitempty"`
	FiscalYear     string        `json:"fiscalyear,omitempty" bson:"fiscalyear,omitempty"`
	StartDate      string        `json:"startdate,omitempty" bson:"startdate,omitempty"`
	EndDate        string        `json:"enddate,omitempty" bson:"enddate,omitempty"`
	Locked         bool          `json:"locked" bson:"locked"`
	Amount         Amount        `json:"amount" bson:"amount"`
	BranchCode     string        `json:"branchcode,omitempty" bson:"branchcode,omitempty"`
	DepartmentCode string        `json:"departmentcode,omitempty" bson:"departmentcode,omitempty"`
	ProjectCode    string        `json:"projectcode,omitempty" bson:"projectcode,omitempty"`
	Direction      string        `json:"direction,omitempty" bson:"direction,omitempty"`
	BookCode       string        `json:"bookcode,omitempty" bson:"bookcode,omitempty"`
	ItemAccount    string        `json:"itemaccount,omitempty" bson:"itemaccount,omitempty"`
	CostAccount    string        `json:"costaccount,omitempty" bson:"costaccount,omitempty"`
	RevenueAccount string                `json:"revenueaccount,omitempty" bson:"revenueaccount,omitempty"`
	Rules          []MappingRule         `json:"rules,omitempty" bson:"rules,omitempty"`
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
	"statement-templates": "gl_statement_templates",
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
	Identity    `bson:",inline"`
	DocNo       string     `json:"docno" bson:"docno"`
	Date        string     `json:"date" bson:"date"`
	BookCode    string     `json:"bookcode" bson:"bookcode"`
	FiscalYear  string     `json:"fiscalyear" bson:"fiscalyear"`
	Description string     `json:"description" bson:"description"`
	Reference   string     `json:"reference" bson:"reference"`
	BranchCode  string     `json:"branchcode" bson:"branchcode"`
	Currency    string     `json:"currency" bson:"currency"`
	Kind        string     `json:"kind" bson:"kind"`
	Status      string     `json:"status" bson:"status"`
	Lines       []Line     `json:"lines" bson:"lines"`
	ReversalOf  string     `json:"reversalof,omitempty" bson:"reversalof,omitempty"`
	PostedAt    *time.Time `json:"postedat,omitempty" bson:"postedat,omitempty"`
	PostedBy    string     `json:"postedby,omitempty" bson:"postedby,omitempty"`
	Reason      string     `json:"reason,omitempty" bson:"reason,omitempty"`
}

func (Journal) CollectionName() string { return "gl_journals" }

type Command struct {
	TargetYear string `json:"targetyear,omitempty"`
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

var codePattern = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}_.\-/]{0,59}$`)

func validCode(code string) bool { return codePattern.MatchString(code) }
func validDate(value string) bool {
	t, err := time.Parse("2006-01-02", value)
	return err == nil && t.Format("2006-01-02") == value
}

func (a Account) Validate() error {
	if len(a.Names) > 32 {
		return fmt.Errorf("ชื่อบัญชีรองรับไม่เกิน 32 ภาษา")
	}
	if !validCode(a.AccountCode) || strings.TrimSpace(a.ThaiName()) == "" || len(a.Names) > 12 {
		return fmt.Errorf("กรุณาระบุรหัสบัญชีและชื่อภาษาไทย")
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
		if len(n.Name) > 300 {
			return fmt.Errorf("ชื่อบัญชียาวเกินไป")
		}
	}
	return nil
}

func (f FiscalYear) Validate() error {
	if !validCode(f.Code) || !validDate(f.StartDate) || !validDate(f.EndDate) || f.StartDate > f.EndDate {
		return fmt.Errorf("กรุณาระบุรหัสและช่วงปีบัญชีให้ถูกต้อง")
	}
	if !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(f.Currency) || f.Scale < 0 || f.Scale > 8 {
		return fmt.Errorf("กรุณากำหนดสกุลเงินและจำนวนทศนิยม 0–8 ตำแหน่ง")
	}
	return nil
}

func (j Journal) Validate(f FiscalYear, accounts map[string]Account) error {
	if !validCode(j.DocNo) || !validDate(j.Date) || strings.TrimSpace(j.Description) == "" {
		return fmt.Errorf("กรุณาระบุเลขที่ วันที่ และคำอธิบายรายการ")
	}
	if !contains([]string{"JV", "UV", "SV", "RV", "PV"}, j.BookCode) {
		return fmt.Errorf("สมุดรายวันไม่ถูกต้อง")
	}
	if !contains([]string{"manual", "opening", "closing", "reversal", "mapping"}, j.Kind) {
		return fmt.Errorf("ประเภทรายการไม่ถูกต้อง")
	}
	if j.FiscalYear != f.Code || j.Date < f.StartDate || j.Date > f.EndDate || f.Closed || !f.IsActive {
		return fmt.Errorf("วันที่อยู่นอกปีบัญชีที่เปิดใช้งาน")
	}
	if j.Currency != f.Currency {
		return fmt.Errorf("สกุลเงินต้องตรงกับปีบัญชี")
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

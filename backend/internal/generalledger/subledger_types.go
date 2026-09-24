package generalledger

// JournalDetails records evidence for the accounting office. Amounts are exact
// decimal strings; line_number is one-based in the journal's Lines array.
type JournalDetails struct {
	Partners       []SubledgerPartner       `json:"partners,omitempty"`
	BankAccounts   []SubledgerBankAccount   `json:"bank_accounts,omitempty"`
	Documents      []SubledgerDocument      `json:"documents,omitempty"`
	Allocations    []SubledgerAllocation    `json:"allocations,omitempty"`
	Settlements    []SubledgerSettlement    `json:"settlements,omitempty"`
	BankLines      []SubledgerBankLine      `json:"bank_lines,omitempty"`
	StatementLines []SubledgerStatementLine `json:"statement_lines,omitempty"`
	Matches        []SubledgerMatch         `json:"matches,omitempty"`
	Withdrawals    []SubledgerWithdrawal    `json:"withdrawals,omitempty"`
	Withholdings   []SubledgerWithholding   `json:"withholdings,omitempty"`
	Vats           []SubledgerVat           `json:"vats,omitempty"`
}
type SubledgerPartner struct {
	Code      string `json:"partner_code"`
	Name      string `json:"name_th"`
	TaxID     string `json:"tax_id,omitempty"`
	TaxBranch string `json:"tax_branch_no,omitempty"`
	Address   string `json:"address,omitempty"`
	// ช่องไม่บังคับสำหรับไฟล์ยื่นภาษีหัก ณ ที่จ่ายด้วยสื่อ (Format กลาง ภ.ง.ด.3/53/2: TITLE_NAME, AMPHUR, PROVINCE, POSTAL_CODE)
	// เก็บใน payload ของทะเบียน ไม่มีคอลัมน์ใน mydocs/datamodels/gl/partners.sql — ตรวจที่ validatePartner
	TitleName    string `json:"title_name,omitempty"`    // คำนำหน้าชื่อ ≤ 100 ตัวอักษร
	AddrDistrict string `json:"addr_district,omitempty"` // อำเภอ/เขต ≤ 50 ตัวอักษร
	AddrProvince string `json:"addr_province,omitempty"` // จังหวัด ≤ 50 ตัวอักษร
	AddrPostcode string `json:"addr_postcode,omitempty"` // รหัสไปรษณีย์ ตัวเลข 5 หลัก (ว่างได้)
	IsCustomer   bool   `json:"is_customer"`
	IsSupplier   bool   `json:"is_supplier"`
	IsActive     bool   `json:"is_active"`
	Version      int64  `json:"version,omitempty"`
}
type SubledgerBankAccount struct {
	Code          string `json:"bank_account_code"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	GLAccountCode string `json:"gl_account_code"`
	Currency      string `json:"currency_code"`
	IsActive      bool   `json:"is_active"`
	Version       int64  `json:"version,omitempty"`
}
type SubledgerDocument struct {
	ID                 string `json:"id"`
	Ledger             string `json:"ledger"`
	PartnerCode        string `json:"partner_code"`
	DocumentNo         string `json:"document_no"`
	Date               string `json:"document_date"`
	DueDate            string `json:"due_date,omitempty"`
	BranchCode         string `json:"branch_code"`
	Kind               int    `json:"document_kind"`
	Side               int    `json:"balance_side"`
	Amount             Amount `json:"amount"`
	Currency           string `json:"currency_code"`
	ControlAccountCode string `json:"control_account_code"`
	Version            int64  `json:"version,omitempty"`
}
type SubledgerAllocation struct {
	ID         string `json:"id"`
	Ledger     string `json:"ledger"`
	DocumentID string `json:"document_id"`
	JournalID  string `json:"journal_id,omitempty"`
	LineNumber int    `json:"line_number"`
	Amount     Amount `json:"amount"`
}
type SubledgerSettlement struct {
	ID                string `json:"id"`
	Ledger            string `json:"ledger"`
	PartnerCode       string `json:"partner_code"`
	DebtDocumentID    string `json:"debt_document_id"`
	PaymentDocumentID string `json:"payment_document_id"`
	Date              string `json:"settlement_date"`
	Amount            Amount `json:"amount"`
}
type SubledgerBankLine struct {
	JournalID       string `json:"journal_id,omitempty"`
	LineNumber      int    `json:"line_number"`
	BankAccountCode string `json:"bank_account_code"`
	Direction       int    `json:"direction"`
}
type SubledgerStatementLine struct {
	ID              string  `json:"id"`
	BankAccountCode string  `json:"bank_account_code"`
	SourceKey       string  `json:"source_key"`
	Date            string  `json:"transaction_date"`
	ValueDate       string  `json:"value_date,omitempty"`
	Reference       string  `json:"bank_reference,omitempty"`
	Description     string  `json:"description,omitempty"`
	Direction       int     `json:"direction"`
	Amount          Amount  `json:"amount"`
	BalanceAfter    *Amount `json:"balance_after,omitempty"`
}
type SubledgerMatch struct {
	ID              string `json:"id"`
	StatementLineID string `json:"statement_line_id"`
	JournalID       string `json:"journal_id,omitempty"`
	LineNumber      int    `json:"line_number"`
	Amount          Amount `json:"amount"`
}
type SubledgerWithdrawal struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Reason string `json:"reason"`
}
type SubledgerOpenDocument struct {
	SubledgerDocument
	Allocated Amount `json:"allocated_amount"`
	Posted    Amount `json:"posted_amount"`
	Settled   Amount `json:"settled_amount"`
	Remaining Amount `json:"remaining_amount"`
}
type SubledgerStatementBalance struct {
	SubledgerStatementLine
	Matched   Amount `json:"matched_amount"`
	Remaining Amount `json:"remaining_amount"`
}
type SubledgerBankBalance struct {
	SubledgerBankLine
	DocNo     string `json:"docno"`
	Date      string `json:"date"`
	Amount    Amount `json:"amount"`
	Matched   Amount `json:"matched_amount"`
	Remaining Amount `json:"remaining_amount"`
}

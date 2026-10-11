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

// ความยาวสูงสุดตาม Champ BCBankContact (Code 10, Name 50, PersonPosition 50, Address 255, Telephone 30)
const (
	maxContactCodeLen = 10
	maxContactNameLen = 50
	maxContactPosLen  = 50
	maxContactAddrLen = 255
	maxContactTelLen  = 30
	// MaxContactsPerBranch กันคำขอที่ใหญ่ผิดปกติ — ผู้ติดต่อต่อสาขาจริงมีไม่กี่คน
	MaxContactsPerBranch = 200
)

// BankContact คือผู้ติดต่อหนึ่งคนของ ธนาคาร + สาขา
type BankContact struct {
	ContactCode string `json:"contact_code"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	Address     string `json:"address"`
	Telephone   string `json:"telephone"`
	IsContact   bool   `json:"is_contact"`
}

// BankContactInput คือแถวที่ผู้ใช้กรอก (IsContact ว่าง = ติดต่อ)
type BankContactInput struct {
	ContactCode string
	Name        string
	Position    string
	Address     string
	Telephone   string
	IsContact   *bool
}

func normalizeBankBranch(bankCode, branchCode string) (string, string, error) {
	bankCode, branchCode = strings.TrimSpace(bankCode), strings.TrimSpace(branchCode)
	if bankCode == "" {
		return "", "", badRequest("bank_contact_err_bank_required", "bank_code")
	}
	if branchCode == "" {
		return "", "", badRequest("bank_contact_err_branch_required", "bank_branch_code")
	}
	return bankCode, branchCode, nil
}

// ListBankContacts คืนผู้ติดต่อของ ธนาคาร + สาขา เรียงตามลำดับที่บันทึก (Champ CASHFrmBankBookNew.cpp:294-296)
func ListBankContacts(ctx context.Context, db querier, company, bankCode, branchCode string) ([]BankContact, error) {
	bankCode, branchCode, err := normalizeBankBranch(bankCode, branchCode)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx,
		`SELECT contact_code, name, position, address, telephone, is_contact
		 FROM bank_contacts WHERE company = $1 AND bank_code = $2 AND bank_branch_code = $3
		 ORDER BY line_no, contact_code LIMIT $4 OFFSET $5`,
		company, bankCode, branchCode, MaxContactsPerBranch, 0)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BankContact{}
	for rows.Next() {
		var c BankContact
		if err = rows.Scan(&c.ContactCode, &c.Name, &c.Position, &c.Address, &c.Telephone, &c.IsContact); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func hasControl(s string) bool { return strings.ContainsFunc(s, unicode.IsControl) }

// normalizeContacts ตัดช่องว่าง ตรวจช่องบังคับ ความยาว และรหัสซ้ำในชุดเดียวกัน (ไม่แยกตัวพิมพ์ใหญ่เล็ก)
func normalizeContacts(in []BankContactInput) ([]BankContactInput, error) {
	if len(in) > MaxContactsPerBranch {
		return nil, badRequest("bank_contact_err_too_many", "contacts", fmt.Sprint(MaxContactsPerBranch))
	}
	out := make([]BankContactInput, len(in))
	seen := make(map[string]bool, len(in))
	for i, c := range in {
		c.ContactCode = strings.TrimSpace(c.ContactCode)
		c.Name = strings.TrimSpace(c.Name)
		c.Position = strings.TrimSpace(c.Position)
		c.Address = strings.TrimSpace(c.Address)
		c.Telephone = strings.TrimSpace(c.Telephone)
		line := fmt.Sprint(i + 1)
		field := fmt.Sprintf("contacts[%d].", i)
		if c.ContactCode == "" {
			return nil, badRequest("bank_contact_err_required", field+"contact_code", line)
		}
		if c.Name == "" {
			return nil, badRequest("bank_contact_err_required", field+"name", line)
		}
		if hasControl(c.ContactCode) {
			return nil, badRequest("bank_contact_err_text_invalid", field+"contact_code", line)
		}
		for _, limit := range []struct {
			name, value string
			max         int
		}{
			{"contact_code", c.ContactCode, maxContactCodeLen}, {"name", c.Name, maxContactNameLen},
			{"position", c.Position, maxContactPosLen}, {"address", c.Address, maxContactAddrLen},
			{"telephone", c.Telephone, maxContactTelLen},
		} {
			if utf8.RuneCountInString(limit.value) > limit.max {
				return nil, badRequest("bank_contact_err_too_long", field+limit.name, line)
			}
		}
		key := strings.ToUpper(c.ContactCode)
		if seen[key] {
			return nil, badRequest("bank_contact_err_duplicate_code", field+"contact_code", c.ContactCode)
		}
		seen[key] = true
		out[i] = c
	}
	return out, nil
}

// SaveBankContacts บันทึกผู้ติดต่อ "ทั้งชุด" ของ ธนาคาร + สาขา ในธุรกรรมเดียว: รหัสที่มีอยู่ถูกแก้ รหัสใหม่ถูกเพิ่ม
// รหัสที่ไม่อยู่ในชุดถูกลบ (ชุดว่าง = ล้างทั้งสาขา) — พฤติกรรมเดียวกับ Champ ที่ลบแล้วใส่ใหม่ แต่คงผู้สร้าง/เวลาสร้างเดิมของแถวที่ไม่ถูกลบ
func SaveBankContacts(ctx context.Context, db *sql.DB, scope Scope, bankCode, branchCode string, contacts []BankContactInput) ([]BankContact, error) {
	bankCode, branchCode, err := normalizeBankBranch(bankCode, branchCode)
	if err != nil {
		return nil, err
	}
	contacts, err = normalizeContacts(contacts)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = requireBank(ctx, tx, scope.Company, bankCode); err != nil {
		return nil, err
	}
	ok, err := CodeExists(ctx, tx, scope.Company, KindBankBranch, branchCode)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, notFound("bank_code_err_not_found", "bank_branch_code", branchCode)
	}
	codes := make([]string, len(contacts))
	for i, c := range contacts {
		codes[i] = c.ContactCode
		active := true
		if c.IsContact != nil {
			active = *c.IsContact
		}
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO bank_contacts (company, bank_code, bank_branch_code, contact_code, line_no, name, position, address, telephone, is_contact, created_by, updated_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
			 ON CONFLICT (company, bank_code, bank_branch_code, contact_code) DO UPDATE
			 SET line_no = EXCLUDED.line_no, name = EXCLUDED.name, position = EXCLUDED.position, address = EXCLUDED.address,
			     telephone = EXCLUDED.telephone, is_contact = EXCLUDED.is_contact, updated_at = now(), updated_by = EXCLUDED.updated_by`,
			scope.Company, bankCode, branchCode, c.ContactCode, i+1, c.Name, c.Position, c.Address, c.Telephone, active, scope.Actor); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx,
		`DELETE FROM bank_contacts WHERE company = $1 AND bank_code = $2 AND bank_branch_code = $3 AND contact_code <> ALL($4::text[])`,
		scope.Company, bankCode, branchCode, pq.Array(codes)); err != nil {
		return nil, err
	}
	saved, err := ListBankContacts(ctx, tx, scope.Company, bankCode, branchCode)
	if err != nil {
		return nil, err
	}
	return saved, tx.Commit()
}

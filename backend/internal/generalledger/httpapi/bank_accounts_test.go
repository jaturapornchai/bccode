package httpapi

import (
	"strings"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

// JSON ที่จอสมุดเงินฝากส่งจริง (frontend/src/lib/cheque-api.ts sendBookCommand + BFF /api/gl/command)
const frontendBankBookCreate = `{
  "resource": "bank-accounts",
  "action": "create",
  "requestid": "11111111-2222-4333-8444-555555555555",
  "bankaccount": {
    "bank_account_code": "KBANK-01",
    "account_number": "123-4-56789-0",
    "account_name": "บริษัท รุ่งเรืองการค้า จำกัด",
    "bank_code": "KBANK",
    "bank_branch_code": "0123",
    "account_type_code": "CURRENT",
    "gl_account_code": "11120",
    "currency_code": "THB",
    "is_active": true
  }
}`

const frontendBankBookChangeNumber = `{
  "resource": "bank-accounts",
  "action": "change-account-number",
  "requestid": "11111111-2222-4333-8444-555555555556",
  "id": "KBANK-01",
  "version": 5,
  "newaccountnumber": "999-9-99999-9"
}`

func TestBankBookCommandsDecodeWithCurrentFrontendBodies(t *testing.T) {
	cmd, err := decodeCommand(strings.NewReader(frontendBankBookCreate))
	if err != nil {
		t.Fatal(err)
	}
	book := cmd.BankAccount
	if cmd.Resource != gl.BankAccountKind || book == nil || book.Code != "KBANK-01" || book.BankCode != "KBANK" || book.BankBranchCode != "0123" || book.AccountTypeCode != "CURRENT" || book.GLAccountCode != "11120" || book.Currency != "THB" {
		t.Fatalf("create body decoded wrong: %+v", cmd)
	}

	cmd, err = decodeCommand(strings.NewReader(frontendBankBookChangeNumber))
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Action != gl.ChangeAccountNumberAction || cmd.ID != "KBANK-01" || cmd.Version != 5 || cmd.NewAccountNumber != "999-9-99999-9" || cmd.BankAccount != nil {
		t.Fatalf("change-number body decoded wrong: %+v", cmd)
	}
}

func TestBankBookCommandRejectsUnknownFields(t *testing.T) {
	// ฟิลด์ที่ไม่รู้จัก (เช่น ยอดคงเหลือที่หน้าจอแก้ไม่ได้) ต้องไม่ถูกเงียบๆ ข้ามไป
	body := strings.Replace(frontendBankBookCreate, `"is_active": true`, `"is_active": true, "gl_balance": "100.00"`, 1)
	if _, err := decodeCommand(strings.NewReader(body)); err == nil {
		t.Fatal("unknown field gl_balance must be rejected")
	}
}

func TestBankBookPermissionsFollowTheBookBankScreen(t *testing.T) {
	if resourceScreens[gl.BankAccountKind] != bookBankScreen || bookBankScreen != "book-bank" {
		t.Fatalf("bank-accounts screen = %q", resourceScreens[gl.BankAccountKind])
	}
	screen := resourceScreens[gl.BankAccountKind]
	// เปิดจอได้ (อ่านรายการ) แต่ยังสร้าง/แก้/ลบไม่ได้จนกว่าจะได้สิทธิ์ของ action นั้น
	readOnly := map[string]bool{"book-bank": true}
	for _, action := range []string{"create", "update", "delete"} {
		if allowed(readOnly, screen, action) {
			t.Fatalf("screen entry alone must not grant %s", action)
		}
	}
	// "เปลี่ยนเลขที่บัญชี" ใช้สิทธิ์ update (http.go แมป action นี้เป็น update)
	editor := map[string]bool{"book-bank": true, "book-bank:update": true}
	if !allowed(editor, screen, "update") || allowed(editor, screen, "delete") || allowed(editor, screen, "create") {
		t.Fatal("update right must grant only update")
	}
	// สิทธิ์จอบัญชีแยกประเภทอื่นไม่ล้นมาถึงสมุดเงินฝาก
	if allowed(map[string]bool{"gl-journals": true, "gl-journals:update": true}, screen, "update") {
		t.Fatal("journal screen rights must not grant bank book commands")
	}
	if !allowed(map[string]bool{"*": true}, screen, "delete") {
		t.Fatal("owner wildcard must grant everything")
	}
}

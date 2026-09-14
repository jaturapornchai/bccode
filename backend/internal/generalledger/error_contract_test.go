package generalledger

import (
	"errors"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

const accountDupMessage = `E11000 duplicate key error collection: appdb.chart_of_accounts index: holdingcode_1_businesscode_1_accountcode_1 dup key: { holdingcode: "H1", businesscode: "B1", accountcode: "1100" }`

func writeError(message string) *mongo.WriteException {
	return &mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000, Message: message}}}
}

// QA UAT: creating an account with an already-used code must come back as a Thai
// 409 duplicate_code, both when the driver keeps the WriteException and when the
// transaction wrapper only leaves the E11000 text behind.
func TestDuplicateAccountCodeBecomesThaiDuplicateCode(t *testing.T) {
	const wantMessage = "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น"
	cases := map[string]error{
		"write exception on chart_of_accounts": duplicateKeyError("accounts", writeError(accountDupMessage)),
		"plain E11000 text (transaction wrapper)": transactionDuplicateError(
			errors.New(accountDupMessage)),
		"wrapped transaction error": transactionDuplicateError(
			&mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000, Message: accountDupMessage}}}),
	}
	for name, err := range cases {
		user, ok := AsUserError(err)
		if !ok {
			t.Fatalf("%s: expected a UserError, got %v", name, err)
		}
		if user.Code != CodeDuplicateCode {
			t.Errorf("%s: code = %q, want %q", name, user.Code, CodeDuplicateCode)
		}
		if user.HTTPStatus() != 409 {
			t.Errorf("%s: status = %d, want 409", name, user.HTTPStatus())
		}
		if user.Error() != wantMessage {
			t.Errorf("%s: message = %q, want %q", name, user.Error(), wantMessage)
		}
	}
}

func TestDuplicateIdBecomesDuplicateRequest(t *testing.T) {
	raw := writeError(`E11000 duplicate key error collection: appdb.gl_events index: _id_ dup key: { _id: "abc" }`)
	user, ok := AsUserError(transactionDuplicateError(raw))
	if !ok || user.Code != CodeDuplicateRequest || user.HTTPStatus() != 409 {
		t.Fatalf("duplicate _id: got %v (ok=%v)", raw, ok)
	}
	if !strings.Contains(user.Error(), "ถูกบันทึกไปแล้ว") {
		t.Fatalf("duplicate _id: unexpected Thai message %q", user.Error())
	}
}

// Real infrastructure failures must not be dressed up as user errors.
func TestNonDuplicateErrorsAreNotSwallowed(t *testing.T) {
	boom := errors.New("server selection error: context deadline exceeded")
	if got := duplicateKeyError("accounts", boom); got != boom {
		t.Fatalf("duplicateKeyError changed a non-duplicate error: %v", got)
	}
	if got := transactionDuplicateError(boom); got != boom {
		t.Fatalf("transactionDuplicateError changed a non-duplicate error: %v", got)
	}
}

// Guard regression: the version guard (mutations.go checkVersion) and the model
// validation path keep their Thai text and gain stable codes.
func TestAccountGuardsCarryCodesAndThaiMessages(t *testing.T) {
	stale := checkVersion(Command{Version: 3}, Identity{Version: 5})
	user, ok := AsUserError(stale)
	if !ok || user.Code != CodeStaleVersion || user.HTTPStatus() != 409 {
		t.Fatalf("stale version: got %v (ok=%v)", stale, ok)
	}
	if user.Error() != "รายการนี้เปลี่ยนไปแล้ว กรุณาโหลดข้อมูลล่าสุดก่อนบันทึก" {
		t.Fatalf("stale version message changed: %q", user.Error())
	}

	invalid := Account{AccountCode: "มีช่องว่าง"}.Validate()
	if invalid == nil {
		t.Fatal("expected Account.Validate to reject an invalid account code")
	}
	wrapped, ok := AsUserError(validationFailed(invalid))
	if !ok || wrapped.Code != CodeValidationFailed || wrapped.Error() != invalid.Error() {
		t.Fatalf("validation wrapping: got %v (ok=%v) want message %q", wrapped, ok, invalid.Error())
	}
	if wrapped.HTTPStatus() != 409 {
		t.Fatalf("validation status = %d, want 409", wrapped.HTTPStatus())
	}
}

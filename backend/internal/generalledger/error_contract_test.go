package generalledger

import "testing"

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

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

	invalid := Account{AccountCode: "มี ช่องว่าง"}.Validate()
	if invalid == nil {
		t.Fatal("expected Account.Validate to reject an invalid account code")
	}
	// ข้อผิดพลาดรายช่องมีรหัสของตัวเองอยู่แล้ว validationFailed ต้องส่งต่อโดยไม่ทับรหัสและชื่อช่อง
	wrapped, ok := AsUserError(validationFailed(invalid))
	if !ok || wrapped.Code != "code_has_space" || wrapped.Field != "accountcode" || wrapped.Error() != invalid.Error() {
		t.Fatalf("validation wrapping: got %v (ok=%v) want message %q", wrapped, ok, invalid.Error())
	}
	// ข้อมูลที่ผู้ใช้กรอกผิด = 400; 409 สงวนไว้สำหรับ version/รหัสซ้ำ/สถานะไม่ให้ทำ (errors.go fieldError vs conflictError)
	if wrapped.HTTPStatus() != 400 {
		t.Fatalf("validation status = %d, want 400", wrapped.HTTPStatus())
	}
}

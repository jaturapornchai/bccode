package httpapi

import (
	"net/http"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/microservice"
)

// ชนิดเช็ค/ชนิดเอกสารผิด ต้องถูกปฏิเสธก่อนแตะ session หรือฐานข้อมูล
func TestChequeRoutesRejectInvalidTypeBeforeAccess(t *testing.T) {
	for _, tc := range []struct {
		name  string
		run   func(*Http, microservice.IContext) error
		ctx   *testContext
		code  string
		field string
	}{
		{"registers type", (*Http).listRegisters, newContext(http.MethodGet, "/cheque/v1/registers?type=2", ""), "cheque_err_type_invalid", "type"},
		{"registers status", (*Http).listRegisters, newContext(http.MethodGet, "/cheque/v1/registers?type=0&status=x", ""), "cheque_err_status_invalid", "status"},
		{"selectable transflag", (*Http).listSelectable, newContext(http.MethodGet, "/cheque/v1/registers/selectable?transflag=44", ""), "cheque_err_transflag_invalid", "transflag"},
		{"selectable cancel type", (*Http).listSelectable, newContext(http.MethodGet, "/cheque/v1/registers/selectable?transflag=423&cancel_type=x", ""), "cheque_err_cancel_type_invalid", "cancel_type"},
		{"docs transflag", (*Http).listDocs, newContext(http.MethodGet, "/cheque/v1/docs/44", ""), "cheque_err_transflag_invalid", "transflag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.ctx.params["transflag"] = "44"
			if err := tc.run(handlerWith(nil, "*"), tc.ctx); err != nil {
				t.Fatal(err)
			}
			wantFailure(t, tc.ctx, http.StatusBadRequest, tc.code, tc.field)
		})
	}
}

func TestChequeCommandsRejectInvalidRequests(t *testing.T) {
	for _, tc := range []struct {
		name, body, code, field string
	}{
		{"registers unknown action", `{"resource":"registers","action":"delete","cheque_type":0}`, "cheque_err_action_invalid", "action"},
		{"registers without type", `{"resource":"registers","action":"create"}`, "cheque_err_type_invalid", "cheque_type"},
		{"registers unknown type", `{"resource":"registers","action":"create","cheque_type":5}`, "cheque_err_type_invalid", "cheque_type"},
		{"docs unknown action", `{"resource":"docs","action":"post","transflag":420}`, "cheque_err_action_invalid", "action"},
		{"docs non-cheque transflag", `{"resource":"docs","action":"create","transflag":44}`, "cheque_err_transflag_invalid", "transflag"},
		{"recalculate unknown action", `{"resource":"recalculate","action":"create"}`, "cheque_err_action_invalid", "action"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newContext(http.MethodPost, "/cheque/v1/command", tc.body)
			if err := handlerWith(nil, "*").command(c); err != nil {
				t.Fatal(err)
			}
			wantFailure(t, c, http.StatusBadRequest, tc.code, tc.field)
		})
	}
}

// เขียนต้องมีสิทธิ์ของเมนูนั้นเอง: ทะเบียนเช็คจ่ายไม่ได้สิทธิ์ทะเบียนเช็ครับ, เอกสารคนละชนิดไม่ได้สิทธิ์กัน
func TestChequeCommandsNeedTheirMenuPermission(t *testing.T) {
	for _, tc := range []struct {
		name        string
		body        string
		permissions []string
	}{
		{"read-only register menu", `{"resource":"registers","action":"create","cheque_type":0}`, []string{"cheque-received"}},
		{"issued permission on received cheque", `{"resource":"registers","action":"create","cheque_type":0}`, []string{"cheque-issued", "cheque-issued:create"}},
		{"create does not grant cancel", `{"resource":"registers","action":"cancel","cheque_type":1}`, []string{"cheque-issued", "cheque-issued:create"}},
		{"deposit permission on clear doc", `{"resource":"docs","action":"create","transflag":420}`, []string{"cheque-deposit", "cheque-deposit:create"}},
		{"update does not grant cancel", `{"resource":"docs","action":"cancel","transflag":421}`, []string{"cheque-received-return", "cheque-received-return:update"}},
		{"recalculate needs update", `{"resource":"recalculate","action":"apply"}`, []string{"cheque-recalculate"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newContext(http.MethodPost, "/cheque/v1/command", tc.body)
			if err := handlerWith(nil, tc.permissions...).command(c); err != nil {
				t.Fatal(err)
			}
			wantFailure(t, c, http.StatusForbidden, "cheque_err_action_forbidden", "")
		})
	}
}

func TestChequeReadRoutesNeedMenuPermission(t *testing.T) {
	c := newContext(http.MethodGet, "/cheque/v1/registers?type=0", "")
	if err := handlerWith(nil, "cheque-issued").listRegisters(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusForbidden, "cheque_err_read_forbidden", "")

	c = newContext(http.MethodGet, "/cheque/v1/docs/425", "")
	c.params["transflag"] = "425"
	if err := handlerWith(nil, "cheque-received-clear").listDocs(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusForbidden, "cheque_err_read_forbidden", "")

	c = newContext(http.MethodGet, "/cheque/v1/recalculate", "")
	if err := handlerWith(nil, "cheque-received").checkRecalculate(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusForbidden, "cheque_err_read_forbidden", "")
}

func TestChequePermissionHelpers(t *testing.T) {
	session := func(perms ...string) glhttp.SessionAccess {
		return glhttp.NewSessionAccess(gl.Scope{Holding: "H", Company: "01", Actor: "somchai"}, perms...)
	}
	// ยกเลิกได้ด้วยสิทธิ์ cancel หรือ delete ของเมนู (เหมือนเอกสาร ERP)
	if !allowsWrite(session("cheque-discount", "cheque-discount:delete"), "cheque-discount", actionCancel) {
		t.Fatal("delete permission must allow cancel")
	}
	if !allowsWrite(session("cheque-discount", "cheque-discount:cancel"), "cheque-discount", actionCancel) {
		t.Fatal("cancel permission must allow cancel")
	}
	if allowsWrite(session("cheque-discount", "cheque-discount:update"), "cheque-discount", actionCancel) {
		t.Fatal("update permission must not allow cancel")
	}
	// เห็นเช็ครับได้จากจอเอกสารที่เลือกเช็ครับ แต่ไม่เห็นเช็คจ่าย
	s := session("cheque-received-clear")
	if !canSeeType(s, 0) || canSeeType(s, 1) || !canSeeAnyCheque(s) {
		t.Fatal("received-clear menu sees received cheques only")
	}
	if canSeeAnyCheque(session("book-bank")) {
		t.Fatal("bank book menu alone does not open cheque registers")
	}
	if screen, ok := docScreen(426); !ok || screen != "cheque-issued-cancel" {
		t.Fatalf("docScreen(426) = %q, %v", screen, ok)
	}
	if _, ok := docScreen(44); ok {
		t.Fatal("44 is not a cheque status document")
	}
}

// เปิดทะเบียนตามรหัสต้องมีสิทธิ์อ่านเช็คอย่างน้อยหนึ่งชนิด (ไม่แตะฐานข้อมูลถ้าไม่มี)
func TestGetRegisterDeniedWithoutAnyChequeMenu(t *testing.T) {
	c := newContext(http.MethodGet, "/cheque/v1/registers/x", "")
	c.params["id"] = "11111111-1111-4111-8111-111111111111"
	if err := handlerWith(nil, "bank-code-kind-bank").getRegister(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusForbidden, "cheque_err_read_forbidden", "")
}

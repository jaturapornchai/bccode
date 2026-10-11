package httpapi

import (
	"context"
	"net/http"
	"testing"

	"smlcloudplatform/internal/doctype"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// chequeFlags คือ transflag ทั้ง 10 ชนิดที่โมดูลเช็ค (/cheque/v1) เป็นเจ้าของ:
// ทะเบียนเช็ครับ/จ่าย และเอกสารเปลี่ยนสถานะ 419-426
var chequeFlags = []int{61, 418, 419, 420, 421, 422, 423, 424, 425, 426}

// เอกสารเช็คต้องไม่ผ่านเส้นทางเอกสารทั่วไป /transaction/:module: ตารางและการลงบัญชีเป็นของ /cheque/v1
// (ถ้าเปิดทิ้งไว้ ผู้เรียกจะสร้างเอกสารเช็คที่ไม่ผ่านกติกาสถานะเช็คได้)
func TestAuthorizeRejectsChequeDocTypesBeforeSession(t *testing.T) {
	for _, flag := range chequeFlags {
		dt, ok := doctype.ByFlag(flag)
		if !ok {
			t.Fatalf("doctype registry lost cheque flag %d", flag)
		}
		t.Run(dt.Key, func(t *testing.T) {
			sessionCalled := false
			h := &Http{session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
				sessionCalled = true
				return glhttp.SessionAccess{}, nil
			}}
			request := newPermContext(dt.Key, "?lang=en")
			if h.authorize(request, actionRead) != nil {
				t.Fatalf("cheque doctype %q accepted on /transaction", dt.Key)
			}
			if sessionCalled {
				t.Fatal("session was checked before the cheque doctype was rejected")
			}
			if request.status != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", request.status)
			}
			body, ok := request.body.(apperr.Response)
			if !ok || body.Success {
				t.Fatalf("body = %#v", request.body)
			}
			if body.Code != "erp_err_cheque_module_moved" {
				t.Fatalf("code = %q, want erp_err_cheque_module_moved", body.Code)
			}
			if body.MessageTH == "" || body.MessageTH == body.Code {
				t.Fatalf("message_th = %q, want a Thai sentence", body.MessageTH)
			}
			if body.Message == "" || body.Message == body.Code {
				t.Fatalf("message = %q, want a translated sentence", body.Message)
			}
		})
	}
}

// เอกสารที่ไม่ใช่เช็คยังผ่านด่านนี้ไปตรวจ session ตามเดิม
func TestAuthorizeStillReachesSessionForNonChequeDocTypes(t *testing.T) {
	called := false
	h := &Http{session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
		called = true
		return glhttp.SessionAccess{}, context.Canceled
	}}
	request := newPermContext("sale-invoice", "")
	if h.authorize(request, actionRead) != nil {
		t.Fatal("failed session accepted")
	}
	if !called {
		t.Fatal("non-cheque doctype never reached the session check")
	}
}

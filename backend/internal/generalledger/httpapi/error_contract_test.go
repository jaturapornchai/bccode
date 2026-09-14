package httpapi

import (
	"encoding/json"
	"errors"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

// Contract QA/frontend rely on: `code` is the machine code, `message` is Thai.
func TestGLCommandErrorContract(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		status     int
		wantCode   string
		wantMsg    string
	}{
		{
			name:     "duplicate account code",
			err:      &gl.UserError{Code: gl.CodeDuplicateCode, Message: "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น"},
			status:   409,
			wantCode: "duplicate_code",
			wantMsg:  "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น",
		},
		{
			name:     "guard: account code cannot be changed",
			err:      &gl.UserError{Code: gl.CodeImmutableCode, Message: "รหัสบัญชีแก้ไม่ได้ กรุณาสร้างบัญชีใหม่"},
			status:   409,
			wantCode: "code_immutable",
			wantMsg:  "รหัสบัญชีแก้ไม่ได้ กรุณาสร้างบัญชีใหม่",
		},
		{
			name:     "guard: delete blocked by journal",
			err:      &gl.UserError{Code: gl.CodeReferenced, Message: "บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ"},
			status:   409,
			wantCode: "account_referenced",
			wantMsg:  "บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ",
		},
		{
			name:     "record not found",
			err:      gl.ErrNotFound,
			status:   404,
			wantCode: "not_found",
			wantMsg:  "ไม่พบรายการบัญชีในบริษัทหรือสาขานี้",
		},
		{
			name:     "projection still running",
			err:      gl.ErrProjectionPending,
			status:   409,
			wantCode: "GL_PROJECTION_PENDING",
			wantMsg:  "บันทึกแล้ว กำลังเตรียมข้อมูลรายงาน กรุณารอสักครู่แล้วโหลดใหม่",
		},
		{
			name:     "other Thai user message",
			err:      errors.New("วันที่เอกสารต้องอยู่ในงวดบัญชีที่เปิดอยู่"),
			status:   409,
			wantCode: "invalid_request",
			wantMsg:  "วันที่เอกสารต้องอยู่ในงวดบัญชีที่เปิดอยู่",
		},
		{
			name:     "infrastructure failure stays a server error",
			err:      errors.New(`server selection error: context deadline exceeded`),
			status:   503,
			wantCode: "unavailable",
			wantMsg:  "ระบบบัญชียังไม่พร้อม กรุณาลองใหม่ หากยังไม่สำเร็จให้ผู้ดูแลตรวจการเชื่อมต่อฐานข้อมูล",
		},
	}
	for _, tc := range cases {
		status, payload := errorPayloadFor(tc.err)
		if status != tc.status {
			t.Errorf("%s: status = %d, want %d", tc.name, status, tc.status)
		}
		if payload.Success {
			t.Errorf("%s: success should be false", tc.name)
		}
		if payload.ErrorCode != tc.wantCode && payload.Code != tc.wantCode {
			t.Errorf("%s: code = %q (errorcode = %q), want %q", tc.name, payload.Code, payload.ErrorCode, tc.wantCode)
		}
		if payload.Message != tc.wantMsg {
			t.Errorf("%s: message = %q, want %q", tc.name, payload.Message, tc.wantMsg)
		}
		if payload.StatusCode != tc.status {
			t.Errorf("%s: statuscode = %d, want %d", tc.name, payload.StatusCode, tc.status)
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("%s: marshal: %v", tc.name, err)
		}
		t.Logf("%s -> %d %s", tc.name, status, raw)
	}
}

func TestLocalFailResponsesCarryCode(t *testing.T) {
	if got := fallbackCode(400); got != "invalid_payload" {
		t.Errorf("400 code = %q", got)
	}
	if got := fallbackCode(403); got != "forbidden" {
		t.Errorf("403 code = %q", got)
	}
	if got := fallbackCode(409); got != "invalid_request" {
		t.Errorf("409 code = %q", got)
	}
}

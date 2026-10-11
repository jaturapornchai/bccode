package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// permContext ตอบเฉพาะเมธอดที่ authorize ใช้ (เมธอดอื่นของ IContext จะ panic ถ้าถูกเรียกโดยไม่ตั้งใจ)
type permContext struct {
	microservice.IContext
	req    *http.Request
	params map[string]string
	status int
	body   any
}

func newPermContext(module, query string) *permContext {
	return &permContext{
		req:    httptest.NewRequest(http.MethodGet, "/transaction/"+module+"/list"+query, nil),
		params: map[string]string{"module": module},
	}
}

func (c *permContext) Request() *http.Request              { return c.req }
func (c *permContext) ResponseWriter() http.ResponseWriter { return httptest.NewRecorder() }
func (c *permContext) Param(name string) string            { return c.params[name] }
func (c *permContext) QueryParam(name string) string       { return c.req.URL.Query().Get(name) }
func (c *permContext) Header(name string) string           { return c.req.Header.Get(name) }
func (c *permContext) Response(code int, data interface{}) {
	c.status, c.body = code, data
}

func httpWithPermissions(permissions ...string) *Http {
	return &Http{session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
		return glhttp.NewSessionAccess(gl.Scope{Holding: "H", Company: "C", Actor: "U"}, permissions...), nil
	}}
}

// จอที่ไม่มี <เมนู> = ปฏิเสธการอ่าน; มีแค่ <เมนู> = อ่านได้อย่างเดียว; เขียนต้องมี <เมนู>:<action> ตรงกัน; "*" ผ่านทั้งหมด
func TestAllowsDocTypePerMenuAndAction(t *testing.T) {
	sale, _ := doctype.ByKey("sale-invoice") // เมนู sale
	for _, tc := range []struct {
		name        string
		permissions []string
		action      string
		want        bool
	}{
		{"owner wildcard reads", []string{"*"}, actionRead, true},
		{"owner wildcard deletes", []string{"*"}, actionDelete, true},
		{"menu entry reads", []string{"sale"}, actionRead, true},
		{"menu entry alone is read-only for create", []string{"sale"}, actionCreate, false},
		{"menu entry alone is read-only for update", []string{"sale"}, actionUpdate, false},
		{"menu entry alone is read-only for delete", []string{"sale"}, actionDelete, false},
		{"create granted", []string{"sale", "sale:create"}, actionCreate, true},
		{"create does not grant update", []string{"sale", "sale:create"}, actionUpdate, false},
		{"action without the menu entry is useless", []string{"sale:create"}, actionCreate, false},
		{"other menu does not open this document", []string{"purchase", "purchase:create"}, actionRead, false},
		{"no permissions at all", nil, actionRead, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			access := glhttp.NewSessionAccess(gl.Scope{}, tc.permissions...)
			if got := allowsDocType(access, sale, tc.action); got != tc.want {
				t.Fatalf("allowsDocType(%v, %q) = %v, want %v", tc.permissions, tc.action, got, tc.want)
			}
		})
	}
}

// stock-count เปิดได้จากสองจอ (ใบตรวจนับ และผลตรวจนับ) สิทธิ์เมนูใดเมนูหนึ่งก็พอ แต่ action ต้องมาจากเมนูเดียวกัน
func TestAllowsDocTypeAcceptsEitherStockCountScreen(t *testing.T) {
	count, _ := doctype.ByKey("stock-count")
	for _, tc := range []struct {
		permissions []string
		action      string
		want        bool
	}{
		{[]string{"stock-count"}, actionRead, true},
		{[]string{"stock-count-sheet"}, actionRead, true},
		{[]string{"stock-count-sheet", "stock-count-sheet:update"}, actionUpdate, true},
		{[]string{"stock-count-sheet", "stock-count:update"}, actionUpdate, false},
	} {
		access := glhttp.NewSessionAccess(gl.Scope{}, tc.permissions...)
		if got := allowsDocType(access, count, tc.action); got != tc.want {
			t.Errorf("stock-count %v %q = %v, want %v", tc.permissions, tc.action, got, tc.want)
		}
	}
}

// ประเภทที่ไม่มี menu id ถูกปฏิเสธแม้ผู้ใช้จะมีสิทธิ์อื่น (fail-closed)
func TestAllowsDocTypeFailsClosedWithoutScreen(t *testing.T) {
	access := glhttp.NewSessionAccess(gl.Scope{}, "sale", "sale:create")
	if allowsDocType(access, doctype.DocType{Key: "ghost-type", Flag: 999}, actionRead) {
		t.Fatal("document type outside the registry must be denied")
	}
}

func TestAuthorizeDeniesWith403AndTranslatedMessage(t *testing.T) {
	for _, tc := range []struct {
		name        string
		module      string
		query       string
		permissions []string
		action      string
		wantCode    string
		wantEN      string
	}{
		{"read without menu", "sale-invoice", "?lang=en", []string{"purchase"}, actionRead, "forbidden", "You do not have permission to view this document type. Ask your administrator for access to this menu."},
		{"create with read-only role", "sale-invoice", "?lang=en", []string{"sale"}, actionCreate, "forbidden", "You do not have permission to create, edit or delete this document type. Ask your administrator for access."},
		{"delete with read-only role", "purchase", "", []string{"purchase"}, actionDelete, "forbidden", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := newPermContext(tc.module, tc.query)
			if access := httpWithPermissions(tc.permissions...).authorize(request, tc.action); access != nil {
				t.Fatalf("authorize granted access: %+v", access)
			}
			if request.status != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", request.status)
			}
			body, ok := request.body.(apperr.Response)
			if !ok {
				t.Fatalf("body = %T, want apperr.Response", request.body)
			}
			if body.Success || body.Code != tc.wantCode || body.StatusCode != http.StatusForbidden {
				t.Fatalf("body = %+v", body)
			}
			if body.MessageTH == "" || body.MessageTH == "erp_err_action_forbidden" || body.MessageTH == "erp_err_read_forbidden" {
				t.Fatalf("Thai message missing: %q", body.MessageTH)
			}
			if tc.wantEN != "" && body.Message != tc.wantEN {
				t.Fatalf("message = %q, want %q", body.Message, tc.wantEN)
			}
		})
	}
}

// อ่านกับเขียนใช้ข้อความคนละอัน เพื่อบอกผู้ใช้ว่าขาดสิทธิ์อะไร
func TestAuthorizeReadAndWriteDenialsDiffer(t *testing.T) {
	read := newPermContext("sale-invoice", "")
	httpWithPermissions().authorize(read, actionRead)
	write := newPermContext("sale-invoice", "")
	httpWithPermissions("sale").authorize(write, actionCreate)
	if read.body.(apperr.Response).MessageTH == write.body.(apperr.Response).MessageTH {
		t.Fatal("read and write denials share one message")
	}
}

func TestAuthorizeRejectsUnknownModuleBeforeSession(t *testing.T) {
	called := false
	h := &Http{session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
		called = true
		return glhttp.SessionAccess{}, nil
	}}
	request := newPermContext("not-a-module", "")
	if h.authorize(request, actionRead) != nil {
		t.Fatal("unknown module accepted")
	}
	if request.status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", request.status)
	}
	if called {
		t.Fatal("session resolved for a module that does not exist")
	}
}

// session ที่ยืนยันไม่ผ่านใช้สัญญา error เดียวกับ GL และต้องไม่เป็น 401 (UI จะวนขอ token ใหม่โดยไม่จำเป็น)
func TestAuthorizeSessionFailureIsNeverUnauthorized(t *testing.T) {
	h := &Http{session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
		return glhttp.SessionAccess{}, errors.New("กรุณาเลือกบริษัทก่อนใช้งานบัญชี")
	}}
	request := newPermContext("sale-invoice", "")
	if h.authorize(request, actionRead) != nil {
		t.Fatal("failed session accepted")
	}
	if request.status == http.StatusUnauthorized || request.status < 400 {
		t.Fatalf("status = %d", request.status)
	}
	if body, ok := request.body.(apperr.Response); !ok || body.Success {
		t.Fatalf("body = %#v", request.body)
	}
}

// จอเฉพาะงานเทียบ menu id ตรงตัว (doctype.ScreenGrant): ให้อ่านรายการ + งานของจอนั้นงานเดียว
// จอยกเลิกการอนุมัติไม่ให้ยกเลิกเอกสาร; จอที่ชื่อลงท้าย -cancel แต่เป็นจอหลักของเอกสาร (บันทึกยกเลิกเช็ครับ) ไม่ได้สิทธิ์ยกเลิกฟรี
func TestAllowsDocTypeUsesExactScreenGrants(t *testing.T) {
	pr, _ := doctype.ByKey("purchase-requisition")
	po, _ := doctype.ByKey("purchase-order")
	cheque, ok := doctype.ByKey("chequereceivedcancel")
	if !ok {
		t.Fatal("chequereceivedcancel not registered")
	}
	scope := gl.Scope{Holding: "H", Company: "C", Actor: "U"}
	for _, tc := range []struct {
		name        string
		dt          doctype.DocType
		permissions []string
		action      string
		want        bool
	}{
		{"approve-cancel screen reads", pr, []string{"purchase-requisition-approve-cancel"}, actionRead, true},
		{"approve-cancel screen unapproves", pr, []string{"purchase-requisition-approve-cancel"}, actionUnapprove, true},
		{"approve-cancel screen does not cancel", pr, []string{"purchase-requisition-approve-cancel"}, actionCancel, false},
		{"approve-cancel screen does not approve", pr, []string{"purchase-requisition-approve-cancel"}, actionApprove, false},
		{"approve-cancel screen does not delete", pr, []string{"purchase-requisition-approve-cancel"}, actionDelete, false},
		{"approve screen approves", pr, []string{"purchase-requisition-approve"}, actionApprove, true},
		{"approve screen does not unapprove", pr, []string{"purchase-requisition-approve"}, actionUnapprove, false},
		{"approve screen does not cancel", pr, []string{"purchase-requisition-approve"}, actionCancel, false},
		{"cancel screen cancels", pr, []string{"purchase-requisition-cancel"}, actionCancel, true},
		{"cancel screen does not approve", pr, []string{"purchase-requisition-cancel"}, actionApprove, false},
		{"po approve screen cannot edit", po, []string{"purchase-order-approve"}, actionUpdate, false},
		{"po approve screen cannot create", po, []string{"purchase-order-approve"}, actionCreate, false},
		{"main screen approve permission unapproves", pr, []string{"purchase-requisition", "purchase-requisition:approve"}, actionUnapprove, true},
		{"main screen update permission unapproves", pr, []string{"purchase-requisition", "purchase-requisition:update"}, actionUnapprove, true},
		{"main screen entry alone cannot unapprove", pr, []string{"purchase-requisition"}, actionUnapprove, false},
		{"cheque cancel document screen reads", cheque, []string{"cheque-received-cancel"}, actionRead, true},
		{"cheque cancel document screen entry does not cancel", cheque, []string{"cheque-received-cancel"}, actionCancel, false},
		{"cheque cancel document screen cancel permission", cheque, []string{"cheque-received-cancel", "cheque-received-cancel:cancel"}, actionCancel, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			session := glhttp.NewSessionAccess(scope, tc.permissions...)
			if got := allowsDocType(session, tc.dt, tc.action); got != tc.want {
				t.Fatalf("allowsDocType(%v, %s, %q) = %v, want %v", tc.permissions, tc.dt.Key, tc.action, got, tc.want)
			}
		})
	}
}

// ความผิดพลาดของคำขอตอบเป็น key ของ languages.tsv: code = key, ข้อความตามภาษาของคำขอ ไม่ใช่ key ดิบหรือข้อความเทคนิค
func TestFailRespondsWithLanguageKey(t *testing.T) {
	request := newPermContext("not-a-module", "?lang=en")
	if (&Http{}).authorize(request, actionRead) != nil {
		t.Fatal("unknown module accepted")
	}
	body, ok := request.body.(apperr.Response)
	if !ok || body.Code != "erp_err_invalid_module" {
		t.Fatalf("body = %#v, want code erp_err_invalid_module", request.body)
	}
	if !strings.Contains(body.MessageTH, "not-a-module") || !strings.Contains(body.MessageTH, "ไม่รู้จักประเภทเอกสาร") {
		t.Errorf("Thai message = %q", body.MessageTH)
	}
	if body.Message == "erp_err_invalid_module" || body.Message == "" {
		t.Errorf("message must be translated text, got %q", body.Message)
	}
}

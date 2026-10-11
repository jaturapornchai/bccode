package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/cheque"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// testContext ตอบเฉพาะเมธอดที่ handler ใช้ (เมธอดอื่นของ IContext จะ panic ถ้าถูกเรียกโดยไม่ตั้งใจ)
type testContext struct {
	microservice.IContext
	req    *http.Request
	params map[string]string
	status int
	body   any
	logs   []string
}

func newContext(method, target, body string) *testContext {
	return &testContext{req: httptest.NewRequest(method, target, strings.NewReader(body)), params: map[string]string{}}
}

func (c *testContext) Request() *http.Request              { return c.req }
func (c *testContext) ResponseWriter() http.ResponseWriter { return httptest.NewRecorder() }
func (c *testContext) Param(name string) string            { return c.params[name] }
func (c *testContext) QueryParam(name string) string       { return c.req.URL.Query().Get(name) }
func (c *testContext) Header(name string) string           { return c.req.Header.Get(name) }
func (c *testContext) Log(message string)                  { c.logs = append(c.logs, message) }
func (c *testContext) Response(code int, data interface{}) { c.status, c.body = code, data }

// handlerWith สร้าง Http ที่ session มีสิทธิ์ตามที่ให้ และ connect คืนฐานข้อมูลจำลอง (nil = ห้ามเชื่อมฐานข้อมูล)
func handlerWith(db *sql.DB, permissions ...string) *Http {
	return &Http{
		session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
			return glhttp.NewSessionAccess(gl.Scope{Holding: "H", Company: "01", Actor: "somchai"}, permissions...), nil
		},
		connect: func(string) (*sql.DB, error) {
			if db == nil {
				return nil, errors.New("test: database must not be opened")
			}
			return db, nil
		},
	}
}

func wantFailure(t *testing.T, c *testContext, status int, code, field string) apperr.Response {
	t.Helper()
	if c.status != status {
		t.Fatalf("status = %d, want %d (body %+v)", c.status, status, c.body)
	}
	body, ok := c.body.(apperr.Response)
	if !ok {
		t.Fatalf("body = %T, want apperr.Response", c.body)
	}
	if body.Success || body.Code != code || body.Field != field || body.StatusCode != status {
		t.Fatalf("body = %+v, want code %q field %q", body, code, field)
	}
	if body.MessageTH == "" || body.MessageTH == code {
		t.Fatalf("Thai message missing for %q: %q", code, body.MessageTH)
	}
	return body
}

// expectSchema ให้ EnsureSchema ผ่าน (ครั้งแรกต่อฐานข้อมูล)
func expectSchema(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS bank_codes`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
}

func TestListCodesRejectsUnknownKindBeforeAnyAccess(t *testing.T) {
	c := newContext(http.MethodGet, "/cheque/v1/codes/branch?lang=en", "")
	c.params["kind"] = "branch"
	if err := handlerWith(nil).listCodes(c); err != nil {
		t.Fatal(err)
	}
	body := wantFailure(t, c, http.StatusBadRequest, "bank_code_err_kind_invalid", "kind")
	if body.Message == "bank_code_err_kind_invalid" {
		t.Fatalf("English message missing: %q", body.Message)
	}
}

// อ่านรหัสได้เมื่อมีเมนูกำหนดรหัสชนิดใดก็ได้ หรือจอที่ต้องใช้รหัสในคอมโบ; เมนูอื่นของโมดูลอื่นไม่เปิดให้
func TestListCodesReadPermission(t *testing.T) {
	for _, tc := range []struct {
		name        string
		permissions []string
		want        bool
	}{
		{"owner wildcard", []string{"*"}, true},
		{"code screen of its own kind", []string{"bank-code-kind-bank-branch"}, true},
		{"book bank screen needs the combo data", []string{"book-bank"}, true},
		{"cheque screen needs the combo data", []string{"cheque-received"}, true},
		{"action entry without the menu entry", []string{"bank-code-kind-bank:create"}, false},
		{"unrelated menu", []string{"gl-journals"}, false},
		{"no permissions", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			access := glhttp.NewSessionAccess(gl.Scope{Company: "01"}, tc.permissions...)
			if got := canReadCodes(access); got != tc.want {
				t.Fatalf("canReadCodes(%v) = %v, want %v", tc.permissions, got, tc.want)
			}
		})
	}
}

func TestListCodesDeniedWithoutPermissionNeverOpensDatabase(t *testing.T) {
	c := newContext(http.MethodGet, "/cheque/v1/codes/bank", "")
	c.params["kind"] = cheque.KindBank
	if err := handlerWith(nil, "gl-journals").listCodes(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusForbidden, "cheque_err_read_forbidden", "")
}

func TestListCodesSessionFailureIsRelayed(t *testing.T) {
	h := &Http{session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
		return glhttp.SessionAccess{}, errors.New("no session")
	}}
	c := newContext(http.MethodGet, "/cheque/v1/codes/bank", "")
	c.params["kind"] = cheque.KindBank
	if err := h.listCodes(c); err != nil {
		t.Fatal(err)
	}
	if c.status < 400 {
		t.Fatalf("status = %d, want an error status", c.status)
	}
}

func TestListCodesReturnsPage(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectSchema(mock)
	mock.ExpectQuery(`SELECT count\(\*\) FROM bank_codes`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`FROM bank_codes`).WillReturnRows(
		sqlmock.NewRows([]string{"kind", "code", "name", "bank_code", "user_group", "is_active", "version"}).
			AddRow(cheque.KindBank, "BBL", "ธนาคารกรุงเทพ", "", "", true, 1))

	c := newContext(http.MethodGet, "/cheque/v1/codes/bank?q=กรุง&limit=20", "")
	c.params["kind"] = cheque.KindBank
	if err := handlerWith(db, "bank-code-kind-bank").listCodes(c); err != nil {
		t.Fatal(err)
	}
	body, ok := c.body.(map[string]any)
	if c.status != http.StatusOK || !ok || body["success"] != true || body["total"] != 1 {
		t.Fatalf("status %d body %+v logs %v", c.status, c.body, c.logs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCommandRejectsMalformedAndUnknownRequests(t *testing.T) {
	for _, tc := range []struct {
		name  string
		body  string
		code  string
		field string
	}{
		{"not json", `{"resource":`, "cheque_err_invalid_payload", ""},
		{"unknown resource", `{"resource":"cheques","action":"create"}`, "cheque_err_resource_invalid", "resource"},
		{"codes with unknown action", `{"resource":"codes","action":"purge","kind":"bank"}`, "cheque_err_action_invalid", "action"},
		{"codes with unknown kind", `{"resource":"codes","action":"create","kind":"branch"}`, "bank_code_err_kind_invalid", "kind"},
		{"contacts only support save", `{"resource":"bank-contacts","action":"create"}`, "cheque_err_action_invalid", "action"},
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

// เขียนรหัสต้องมีทั้ง <เมนูของชนิดนั้น> และ <เมนู>:<action> — สิทธิ์ของชนิดอื่นหรือ action อื่นไม่พอ
func TestCodesCommandWritePermission(t *testing.T) {
	for _, tc := range []struct {
		name        string
		body        string
		permissions []string
	}{
		{"menu entry alone is read-only", `{"resource":"codes","action":"create","kind":"bank","code":"BBL","name":"x"}`, []string{"bank-code-kind-bank"}},
		{"create does not grant update", `{"resource":"codes","action":"update","kind":"bank","code":"BBL","name":"x","version":1}`, []string{"bank-code-kind-bank", "bank-code-kind-bank:create"}},
		{"create does not grant delete", `{"resource":"codes","action":"delete","kind":"bank","code":"BBL","version":1}`, []string{"bank-code-kind-bank", "bank-code-kind-bank:create"}},
		{"permission of another kind", `{"resource":"codes","action":"create","kind":"bank","code":"BBL","name":"x"}`, []string{"bank-code-kind-bank-branch", "bank-code-kind-bank-branch:create"}},
		{"action without the menu entry", `{"resource":"codes","action":"create","kind":"bank","code":"BBL","name":"x"}`, []string{"bank-code-kind-bank:create"}},
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

func TestCodesCommandCreateReturnsSavedCode(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectSchema(mock)
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO bank_codes`).
		WithArgs("01", cheque.KindBank, "BBL", "ธนาคารกรุงเทพ", nil, "", true, "somchai").
		WillReturnRows(sqlmock.NewRows([]string{"kind", "code", "name", "bank_code", "user_group", "is_active", "version"}).
			AddRow(cheque.KindBank, "BBL", "ธนาคารกรุงเทพ", "", "", true, 1))
	mock.ExpectCommit()

	c := newContext(http.MethodPost, "/cheque/v1/command", `{"resource":"codes","action":"create","kind":"bank","code":"BBL","name":"ธนาคารกรุงเทพ"}`)
	if err := handlerWith(db, "bank-code-kind-bank", "bank-code-kind-bank:create").command(c); err != nil {
		t.Fatal(err)
	}
	body, ok := c.body.(map[string]any)
	if c.status != http.StatusOK || !ok || body["success"] != true {
		t.Fatalf("status %d body %+v", c.status, c.body)
	}
	if saved, ok := body["data"].(cheque.Code); !ok || saved.Code != "BBL" || saved.Version != 1 {
		t.Fatalf("data = %+v", body["data"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ความผิดของผู้ใช้ตอบด้วยข้อความที่แก้ได้ (409 + key) ส่วนความผิดของระบบตอบ 500 ข้อความกลางและไม่รั่วรายละเอียดฐานข้อมูล
func TestCodesCommandMapsUserAndSystemErrors(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectSchema(mock)
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WillReturnRows(sqlmock.NewRows([]string{"kind", "code", "name", "bank_code", "user_group", "is_active", "version"}).
		AddRow(cheque.KindBank, "BBL", "ธนาคารกรุงเทพ", "", "", true, 5))
	mock.ExpectRollback()
	perms := []string{"bank-code-kind-bank", "bank-code-kind-bank:update"}

	c := newContext(http.MethodPost, "/cheque/v1/command", `{"resource":"codes","action":"update","kind":"bank","code":"BBL","name":"x","version":3}`)
	if err := handlerWith(db, perms...).command(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusConflict, "cheque_err_version_conflict", "version")

	mock.ExpectBegin().WillReturnError(errors.New("pq: password authentication failed for user secret_user"))
	c = newContext(http.MethodPost, "/cheque/v1/command", `{"resource":"codes","action":"update","kind":"bank","code":"BBL","name":"x","version":3}`)
	if err := handlerWith(db, perms...).command(c); err != nil {
		t.Fatal(err)
	}
	body := wantFailure(t, c, http.StatusInternalServerError, "cheque_err_internal", "")
	if strings.Contains(body.Message, "secret_user") || strings.Contains(body.MessageTH, "secret_user") {
		t.Fatalf("database detail leaked to the client: %+v", body)
	}
	if len(c.logs) == 0 || !strings.Contains(c.logs[0], "secret_user") {
		t.Fatalf("technical detail must be logged server-side: %v", c.logs)
	}
}

func TestBankContactsCommandNeedsUpdatePermission(t *testing.T) {
	body := `{"resource":"bank-contacts","action":"save","bank_code":"BBL","bank_branch_code":"0123","contacts":[]}`
	for _, perms := range [][]string{{"bank-contact"}, {"bank-contact:update"}, {"book-bank", "book-bank:update"}} {
		c := newContext(http.MethodPost, "/cheque/v1/command", body)
		if err := handlerWith(nil, perms...).command(c); err != nil {
			t.Fatal(err)
		}
		wantFailure(t, c, http.StatusForbidden, "cheque_err_action_forbidden", "")
	}
}

func TestBankContactsCommandSavesWholeSet(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectSchema(mock)
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR SHARE`).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectQuery(`FOR SHARE`).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO bank_contacts`).
		WithArgs("01", "BBL", "0123", "C1", 1, "คุณสมศรี", "", "", "02-231-4333", true, "somchai").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM bank_contacts`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`FROM bank_contacts WHERE`).WillReturnRows(
		sqlmock.NewRows([]string{"contact_code", "name", "position", "address", "telephone", "is_contact"}).
			AddRow("C1", "คุณสมศรี", "", "", "02-231-4333", true))
	mock.ExpectCommit()

	c := newContext(http.MethodPost, "/cheque/v1/command",
		`{"resource":"bank-contacts","action":"save","bank_code":"BBL","bank_branch_code":"0123","contacts":[{"contact_code":"C1","name":"คุณสมศรี","telephone":"02-231-4333"}]}`)
	if err := handlerWith(db, "bank-contact", "bank-contact:update").command(c); err != nil {
		t.Fatal(err)
	}
	body, ok := c.body.(map[string]any)
	if c.status != http.StatusOK || !ok || body["total"] != 1 {
		t.Fatalf("status %d body %+v", c.status, c.body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// จอสมุดเงินฝากเห็นผู้ติดต่อได้ (อ่านอย่างเดียว) แต่จอที่ไม่เกี่ยวข้องห้าม
func TestListBankContactsReadPermission(t *testing.T) {
	denied := newContext(http.MethodGet, "/cheque/v1/bank-contacts?bank_code=BBL&bank_branch_code=0123", "")
	if err := handlerWith(nil, "bank-code-kind-bank").listBankContacts(denied); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, denied, http.StatusForbidden, "cheque_err_read_forbidden", "")

	for _, screen := range []string{"bank-contact", "book-bank"} {
		db, mock, _ := sqlmock.New()
		expectSchema(mock)
		mock.ExpectQuery(`FROM bank_contacts WHERE`).WithArgs("01", "BBL", "0123", cheque.MaxContactsPerBranch, 0).
			WillReturnRows(sqlmock.NewRows([]string{"contact_code", "name", "position", "address", "telephone", "is_contact"}))
		c := newContext(http.MethodGet, "/cheque/v1/bank-contacts?bank_code=BBL&bank_branch_code=0123", "")
		if err := handlerWith(db, screen).listBankContacts(c); err != nil {
			t.Fatal(err)
		}
		if c.status != http.StatusOK {
			t.Fatalf("%s: status %d body %+v", screen, c.status, c.body)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%s: %v", screen, err)
		}
		db.Close()
	}
}

func TestListBankContactsRequiresBankAndBranch(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectSchema(mock)
	c := newContext(http.MethodGet, "/cheque/v1/bank-contacts?bank_code=BBL", "")
	if err := handlerWith(db, "bank-contact").listBankContacts(c); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, c, http.StatusBadRequest, "bank_contact_err_branch_required", "bank_branch_code")
}

func TestTextSubstitutesArguments(t *testing.T) {
	if got := text("no_such_key_{0}", "th", "x"); got != "no_such_key_x" {
		t.Fatalf("text = %q", got)
	}
}

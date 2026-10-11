package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/config"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// Http คือ REST ของโมดูลเช็ค/ธนาคาร: /cheque/v1/* (อ่านด้วย GET, เขียนด้วย POST /cheque/v1/command)
type Http struct {
	ms *microservice.Microservice
	// session ยืนยัน session + สิทธิ์รายเมนู (ฉีดได้เพื่อทดสอบโดยไม่ต้องมีฐานข้อมูลกลาง)
	session func(context.Context, microservice.IContext) (glhttp.SessionAccess, error)
	// connect เปิดฐานข้อมูลของ holding (ฉีดได้เพื่อทดสอบ handler กับ sqlmock)
	connect func(holding string) (*sql.DB, error)
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	return &Http{ms: ms, session: glhttp.ResolveSessionAccess, connect: mypg.PgSqlFastConnect}
}

func (h *Http) RegisterHttp() {
	h.ms.GET("/cheque/v1/codes/:kind", h.listCodes)
	h.ms.GET("/cheque/v1/bank-contacts", h.listBankContacts)
	h.ms.POST("/cheque/v1/command", h.command)
}

// การกระทำที่ตรวจสิทธิ์: "" = เข้าจอ/อ่าน, อื่นๆ ต้องมีทั้ง <เมนู> และ <เมนู>:<action> (กฎเดียวกับ GL)
const (
	actionRead   = ""
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
)

const (
	bankContactScreen = "bank-contact"
	bookBankScreen    = "book-bank"
)

// codeScreens คือเมนูของทะเบียนรหัสแต่ละชนิด (frontend/src/lib/menu-data.ts) — เขียนรหัสได้เฉพาะผู้ที่มีสิทธิ์เมนูของชนิดนั้น
var codeScreens = map[string]string{
	cheque.KindBank:            "bank-code-kind-bank",
	cheque.KindBankBranch:      "bank-code-kind-bank-branch",
	cheque.KindBankService:     "bank-code-kind-bank-service",
	cheque.KindBankAccountType: "bank-code-kind-bank-account-type",
	cheque.KindChequeExtStatus: "bank-code-kind-cheque-ext-status",
}

// codeLookupScreens คือจอที่ต้องดึงรหัสไปใส่คอมโบ จึงอ่านรายการรหัสได้ (อ่านอย่างเดียว) นอกเหนือจากจอกำหนดรหัสเอง
var codeLookupScreens = []string{
	bookBankScreen, bankContactScreen, "bank-service",
	"cheque-received", "cheque-issued", "cheque-issued-prepare", "cheque-movement", "cheque-recalculate",
	"cheque-deposit", "cheque-received-clear", "cheque-received-return", "cheque-redeposit",
	"cheque-received-cancel", "cheque-discount", "cheque-issued-clear", "cheque-issued-cancel",
	"cheque-received-report", "cheque-issued-report",
}

// access คือผลของ authorize: session ที่ยืนยันแล้ว บริษัท/ผู้ทำรายการ และฐานข้อมูลของ holding
type access struct {
	ctx     context.Context
	scope   cheque.Scope
	session glhttp.SessionAccess
	db      *sql.DB
}

// authorize ตรวจตามลำดับ session/สมาชิกภาพ → สิทธิ์ (allowed) → เชื่อมฐานข้อมูล + สร้างตาราง ก่อนแตะข้อมูลใดๆ;
// ปฏิเสธ = ตอบ response เองแล้วคืน nil (ผู้เรียกคืน nil ต่อทันที)
func (h *Http) authorize(request microservice.IContext, write bool, allowed func(glhttp.SessionAccess) bool) *access {
	ctx := request.Request().Context()
	lang := requestLanguage(request)
	session, err := h.session(ctx, request)
	if err != nil {
		status, body := glhttp.SessionFailure(err, lang)
		request.Response(status, body)
		return nil
	}
	if session.Scope.Company == "" || !allowed(session) {
		key := "cheque_err_read_forbidden"
		if write {
			key = "cheque_err_action_forbidden"
		}
		_ = fail(request, http.StatusForbidden, key, "")
		return nil
	}
	db, err := h.connect(session.Scope.Holding)
	if err != nil {
		request.Log(fmt.Sprintf("cheque: connect holding database: %v", err))
		_ = fail(request, http.StatusServiceUnavailable, "cheque_err_database_unavailable", "")
		return nil
	}
	if err := cheque.EnsureSchema(ctx, db); err != nil {
		request.Log(fmt.Sprintf("cheque: ensure schema: %v", err))
		_ = fail(request, http.StatusInternalServerError, "cheque_err_internal", "")
		return nil
	}
	return &access{ctx: ctx, scope: cheque.Scope{Company: session.Scope.Company, Actor: session.Scope.Actor}, session: session, db: db}
}

func requestLanguage(request microservice.IContext) string {
	if lang := request.QueryParam("lang"); lang != "" {
		return lang
	}
	if lang := request.Header("Accept-Language"); lang != "" {
		return lang
	}
	return "th"
}

// text คือข้อความของ key ตามภาษาที่ขอ แทนค่า {0}, {1}, … (ไม่มีแถว = คืน key ตามกติกาของ language.Text)
func text(key, lang string, args ...string) string {
	value := language.Text(key, lang)
	for i, arg := range args {
		value = strings.ReplaceAll(value, "{"+strconv.Itoa(i)+"}", arg)
	}
	return value
}

// fail ตอบความผิดพลาดด้วย key ใน languages.tsv (code = key): message ตามภาษาของคำขอ, message_th เป็นภาษาไทย
func fail(request microservice.IContext, status int, key, field string, args ...string) error {
	lang := requestLanguage(request)
	appErr := &apperr.AppError{
		Code:       key,
		Message:    text(key, lang, args...),
		ThaiMsg:    text(key, "th", args...),
		HTTPStatus: status,
		Field:      field,
	}
	request.Response(status, appErr.ToResponse())
	return nil
}

// failure แปลง error ของโมดูลเป็น response: ความผิดของผู้ใช้ (cheque.Error) ตอบด้วยข้อความที่แก้ได้;
// อย่างอื่นเป็น 500 ข้อความกลาง (รายละเอียดทางเทคนิคเก็บใน log เท่านั้น ไม่ส่งถึงผู้ใช้)
func failure(request microservice.IContext, err error) error {
	if e, ok := cheque.AsError(err); ok {
		return fail(request, e.Status, e.Key, e.Field, e.Args...)
	}
	request.Log(fmt.Sprintf("cheque: %v", err))
	return fail(request, http.StatusInternalServerError, "cheque_err_internal", "")
}

func respond(request microservice.IContext, body map[string]any) error {
	body["success"] = true
	request.Response(http.StatusOK, body)
	return nil
}

func intParam(request microservice.IContext, name string) int {
	value, _ := strconv.Atoi(request.QueryParam(name))
	return value
}

// canReadCodes: เปิดดูรหัสได้เมื่อมีสิทธิ์เมนูกำหนดรหัสชนิดใดก็ได้ หรือจอที่ต้องใช้รหัสในคอมโบ
func canReadCodes(session glhttp.SessionAccess) bool {
	for _, screen := range codeScreens {
		if session.Allows(screen, actionRead) {
			return true
		}
	}
	for _, screen := range codeLookupScreens {
		if session.Allows(screen, actionRead) {
			return true
		}
	}
	return false
}

// GET /cheque/v1/codes/:kind?q=&page=&limit=&bank_code=&active=1
func (h *Http) listCodes(request microservice.IContext) error {
	kind := request.Param("kind")
	if !cheque.ValidCodeKind(kind) {
		return fail(request, http.StatusBadRequest, "bank_code_err_kind_invalid", "kind", kind)
	}
	a := h.authorize(request, false, canReadCodes)
	if a == nil {
		return nil
	}
	page, err := cheque.ListCodes(a.ctx, a.db, a.scope.Company, kind, cheque.CodeFilter{
		Query:      request.QueryParam("q"),
		BankCode:   request.QueryParam("bank_code"),
		ActiveOnly: request.QueryParam("active") == "1",
		Page:       intParam(request, "page"),
		Limit:      intParam(request, "limit"),
	})
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"items": page.Items, "total": page.Total, "page": page.Page, "limit": page.Limit})
}

// GET /cheque/v1/bank-contacts?bank_code=&bank_branch_code=
// จอสมุดเงินฝากแสดงผู้ติดต่ออย่างเดียว จึงอ่านได้ด้วยสิทธิ์จอสมุดเงินฝากเช่นกัน
func (h *Http) listBankContacts(request microservice.IContext) error {
	a := h.authorize(request, false, func(s glhttp.SessionAccess) bool {
		return s.Allows(bankContactScreen, actionRead) || s.Allows(bookBankScreen, actionRead)
	})
	if a == nil {
		return nil
	}
	items, err := cheque.ListBankContacts(a.ctx, a.db, a.scope.Company, request.QueryParam("bank_code"), request.QueryParam("bank_branch_code"))
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"items": items, "total": len(items)})
}

// maxCommandBodyBytes: ขนาดคำขอเขียนสูงสุด — ผู้ติดต่อ 200 คนเต็มทุกช่องยังไม่ถึงครึ่ง
const maxCommandBodyBytes = 1 << 20

// commandBody คือโครงคำสั่งเขียนทั้งหมด {resource, action, <payload>} — ฟิลด์ของแต่ละ resource อยู่ระดับเดียวกัน
type commandBody struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Version  int64  `json:"version"`

	// resource codes
	Kind      string `json:"kind"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	BankCode  string `json:"bank_code"`
	UserGroup string `json:"user_group"`
	IsActive  *bool  `json:"is_active"`

	// resource bank-contacts
	BankBranchCode string        `json:"bank_branch_code"`
	Contacts       []contactBody `json:"contacts"`
}

type contactBody struct {
	ContactCode string `json:"contact_code"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	Address     string `json:"address"`
	Telephone   string `json:"telephone"`
	IsContact   *bool  `json:"is_contact"`
}

// POST /cheque/v1/command
func (h *Http) command(request microservice.IContext) error {
	var body commandBody
	reader := http.MaxBytesReader(request.ResponseWriter(), request.Request().Body, maxCommandBodyBytes)
	if err := json.NewDecoder(reader).Decode(&body); err != nil {
		return fail(request, http.StatusBadRequest, "cheque_err_invalid_payload", "")
	}
	switch strings.TrimSpace(body.Resource) {
	case "codes":
		return h.codesCommand(request, body)
	case "bank-contacts":
		return h.bankContactsCommand(request, body)
	}
	return fail(request, http.StatusBadRequest, "cheque_err_resource_invalid", "resource", body.Resource)
}

func (h *Http) codesCommand(request microservice.IContext, body commandBody) error {
	action := strings.TrimSpace(body.Action)
	if action != actionCreate && action != actionUpdate && action != actionDelete {
		return fail(request, http.StatusBadRequest, "cheque_err_action_invalid", "action", body.Action)
	}
	screen, ok := codeScreens[strings.TrimSpace(body.Kind)]
	if !ok {
		return fail(request, http.StatusBadRequest, "bank_code_err_kind_invalid", "kind", body.Kind)
	}
	a := h.authorize(request, true, func(s glhttp.SessionAccess) bool { return s.Allows(screen, action) })
	if a == nil {
		return nil
	}
	input := cheque.CodeInput{
		Kind: body.Kind, Code: body.Code, Name: body.Name, BankCode: body.BankCode,
		UserGroup: body.UserGroup, IsActive: body.IsActive, Version: body.Version,
	}
	var (
		saved cheque.Code
		err   error
	)
	switch action {
	case actionCreate:
		saved, err = cheque.CreateCode(a.ctx, a.db, a.scope, input)
	case actionUpdate:
		saved, err = cheque.UpdateCode(a.ctx, a.db, a.scope, input)
	default:
		saved, err = cheque.DeleteCode(a.ctx, a.db, a.scope, input)
	}
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"data": saved})
}

// bank-contacts save: ผู้ติดต่อทั้งชุดของ ธนาคาร + สาขา (เพิ่ม/แก้/ลบรวมกันในคำสั่งเดียว) ใช้สิทธิ์ update ของเมนู
func (h *Http) bankContactsCommand(request microservice.IContext, body commandBody) error {
	if strings.TrimSpace(body.Action) != "save" {
		return fail(request, http.StatusBadRequest, "cheque_err_action_invalid", "action", body.Action)
	}
	a := h.authorize(request, true, func(s glhttp.SessionAccess) bool { return s.Allows(bankContactScreen, actionUpdate) })
	if a == nil {
		return nil
	}
	contacts := make([]cheque.BankContactInput, len(body.Contacts))
	for i, c := range body.Contacts {
		contacts[i] = cheque.BankContactInput{
			ContactCode: c.ContactCode, Name: c.Name, Position: c.Position,
			Address: c.Address, Telephone: c.Telephone, IsContact: c.IsContact,
		}
	}
	saved, err := cheque.SaveBankContacts(a.ctx, a.db, a.scope, body.BankCode, body.BankBranchCode, contacts)
	if err != nil {
		return failure(request, err)
	}
	return respond(request, map[string]any{"items": saved, "total": len(saved)})
}

// Package httpapi connects the exact ledger engine to the existing authenticated
// main API. No credentials or tenant identities are accepted from request bodies.
package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"smlcloudplatform/internal/config"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

var validHoldingRegex = regexp.MustCompile(`^[A-Za-z0-9_]{1,63}$`)

type Http struct {
	ms    *microservice.Microservice
	store *gl.PostgresStore
	pg    *gl.Postgres
	mu    sync.Mutex
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	ms.Logger.Infof("GL running in pure PostgreSQL direct mode (Zero Kafka, Zero Outbox)")
	return newRuntime(ms, cfg)
}

// RegisterProjectionWorker belongs to legacy consumer mode. In pure PostgreSQL mode,
// changes are synchronously committed via ACID transactions.
func RegisterProjectionWorker(ms *microservice.Microservice, cfg config.IConfig) error {
	ms.Logger.Infof("GL PostgreSQL mode: direct synchronous execution enabled, Kafka projection worker decommissioned")
	return nil
}

func newRuntime(ms *microservice.Microservice, cfg config.IConfig) *Http {
	projection := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		if !validHoldingRegex.MatchString(holding) {
			return nil, fmt.Errorf("รหัสกลุ่มบริษัทไม่ถูกต้อง")
		}
		return mypg.PgSqlFastConnect(holding)
	})
	store := gl.NewPostgresStore(projection)
	return &Http{ms: ms, store: store, pg: projection}
}

func (h *Http) initialize(ctx context.Context) error {
	return nil
}

func (h *Http) RegisterHttp() {
	h.ms.POST("/gl/v2/command", h.command)
	h.ms.GET("/gl/v2/reports/:report", h.report)
	h.ms.GET("/gl/v2/:resource", h.list)
	h.ms.GET("/gl/v2/:resource/:id", h.get)
}

type requestScope struct {
	Scope       gl.Scope
	Permissions map[string]bool
}

func (h *Http) scope(ctx context.Context, request microservice.IContext) (requestScope, error) {
	u := request.UserInfo()
	result := requestScope{
		Scope: gl.Scope{
			Holding: u.HoldingCode,
			Company: u.BusinessCode,
			Actor:   u.UID,
		},
		Permissions: map[string]bool{"*": true},
	}
	if u.HoldingCode == "" || u.BusinessCode == "" {
		return result, fmt.Errorf("กรุณาเลือกบริษัทก่อนใช้งานบัญชี")
	}

	// Connect PostgreSQL for holding
	db, err := mypg.PgSqlFastConnect(u.HoldingCode)
	if err != nil {
		return result, nil
	}

	// Check branch code
	if u.BranchUID != "" {
		var branchCode string
		err := db.QueryRowContext(ctx, `SELECT code FROM branches WHERE holding_code = $1 AND code = $2 AND is_active = true`, u.HoldingCode, u.BranchUID).Scan(&branchCode)
		if err == nil && branchCode != "" {
			result.Scope.Branch = branchCode
		} else {
			result.Scope.Branch = u.BranchUID
		}
	}

	// Check permissions from PostgreSQL holding_members & role_permissions if available
	var role string
	var permSetsJSON []byte
	err = db.QueryRowContext(ctx, `SELECT role, permission_sets FROM holding_members WHERE holding_code = $1 AND (user_id::text = $2 OR user_id::text = $3) AND is_active = true`, u.HoldingCode, u.UID, u.Username).Scan(&role, &permSetsJSON)
	if err == nil {
		if strings.EqualFold(role, "OWNER") || strings.EqualFold(role, "ADMIN") {
			result.Permissions["*"] = true
			return result, nil
		}
		result.Permissions = map[string]bool{}
		var permSets []string
		if len(permSetsJSON) > 0 {
			_ = json.Unmarshal(permSetsJSON, &permSets)
		}
		codes := append([]string{role}, permSets...)
		for _, c := range codes {
			var permsJSON []byte
			if err := db.QueryRowContext(ctx, `SELECT permissions FROM role_permissions WHERE holding_code = $1 AND role_code = $2`, u.HoldingCode, c).Scan(&permsJSON); err == nil {
				var perms []string
				if json.Unmarshal(permsJSON, &perms) == nil {
					for _, p := range perms {
						result.Permissions[p] = true
					}
				}
			}
		}
	}

	return result, nil
}

var resourceScreens = map[string]string{"accounts": "chart-of-accounts", "fiscal-years": "chart-of-accounts", "account-groups": "gl-account-groups", "product-account-groups": "gl-product-account-groups", "mappings": "gl-account-mapping", "budgets": "gl-budget", "periods": "period-lock", "forecast": "cash-flow-forecast", "allocations": "gl-allocation", "statement-templates": "financial-statement-designer", "journal-books": "gl-journal-books"}
var reportScreens = map[string]string{"ledger": "general-ledger", "trialbalance": "trial-balance", "pnl": "profit-loss", "balancesheet": "balance-sheet", "cashflow": "cash-flow", "cashflowforecast": "cash-flow-forecast", "financialgraphs": "financial-graphs", "project-pnl": "project-pnl", "dimensionpnl": "dimension-pnl", "projectsummary": "project-summary-report", "dashboard": "business-dashboard", "executivesummary": "executive-summary", "workingpaper": "working-paper", "daily-check": "daily-info", "annual-balances": "gl-annual-accumulated", "allocate": "gl-allocation", "gljournal": "gl-daily-report", "budgetcomparison": "budget-comparison-report"}

func allowed(p map[string]bool, screen, action string) bool {
	if p["*"] {
		return true
	}
	if screen == "" || !p[screen] {
		return false
	}
	return action == "" || p[screen+":"+action]
}
func anyLedger(p map[string]bool) bool {
	if p["*"] {
		return true
	}
	for _, screen := range resourceScreens {
		if p[screen] {
			return true
		}
	}
	for _, screen := range reportScreens {
		if p[screen] {
			return true
		}
	}
	for _, screen := range []string{"jv-journal", "uv-journal", "sv-journal", "rv-journal", "pv-journal", "gl-post", "gl-unpost", "gl-opening-balance", "financial-close", "gl-year-end", "gl-reprocess", "gl-recalculate-posted", "xbrl-export", "data-backup-export"} {
		if p[screen] {
			return true
		}
	}
	return false
}
func journalScreen(book, kind, status string) string {
	if kind == "opening" {
		return "gl-opening-balance"
	}
	if book != "" {
		switch strings.ToUpper(book) {
		case "JV":
			return "jv-journal"
		case "UV":
			return "uv-journal"
		case "SV":
			return "sv-journal"
		case "RV":
			return "rv-journal"
		case "PV":
			return "pv-journal"
		}
		return ""
	}
	if status == "draft" {
		return "gl-post"
	}
	if status == "posted" {
		return "gl-unpost"
	}
	return ""
}

func response(request microservice.IContext, data interface{}) error {
	request.Response(http.StatusOK, map[string]interface{}{"success": true, "data": data})
	return nil
}

func getRequestLanguage(ctx microservice.IContext) string {
	if lang := ctx.QueryParam("lang"); lang != "" {
		return lang
	}
	if lang := ctx.Header("Accept-Language"); lang != "" {
		return lang
	}
	return "th"
}

// errorPayload is an alias to apperr.Response to align GL error responses
// with the standard application error envelope across all microservices.
type errorPayload = apperr.Response

// errorPayloadFor maps a command error to (HTTP status, error contract body).
// User-caused failures (guards, duplicate codes, validation) are reported as
// 409 (or 404 for missing records) with translated message via language.Text(key, lang);
// anything else stays a server problem (503) and is logged by failure().
func errorPayloadFor(err error, lang ...string) (int, errorPayload) {
	reqLang := "th"
	if len(lang) > 0 && lang[0] != "" {
		reqLang = lang[0]
	}

	if errors.Is(err, gl.ErrProjectionPending) {
		msg := language.Text("gl_err_projection_pending", reqLang)
		if msg == "gl_err_projection_pending" {
			msg = err.Error()
		}
		appErr := &apperr.AppError{
			Code:       "projection_pending",
			Message:    msg,
			ThaiMsg:    err.Error(),
			HTTPStatus: http.StatusConflict,
		}
		res := appErr.ToResponse()
		res.ErrorCode = "GL_PROJECTION_PENDING"
		return http.StatusConflict, res
	}
	if user, ok := gl.AsUserError(err); ok {
		msg := user.Message
		if reqLang != "th" {
			key := "gl_err_" + user.Code
			translated := language.Text(key, reqLang)
			if translated != "" && translated != key {
				msg = translated
			}
		}
		appErr := &apperr.AppError{
			Code:       user.Code,
			Message:    msg,
			ThaiMsg:    user.Message,
			HTTPStatus: user.HTTPStatus(),
		}
		return appErr.StatusCode(), appErr.ToResponse()
	}
	if errors.Is(err, gl.ErrNotFound) {
		fallbackMsg := "ไม่พบรายการบัญชีในบริษัทหรือสาขานี้"
		msg := language.Text("gl_err_not_found", reqLang)
		if msg == "gl_err_not_found" {
			msg = fallbackMsg
		}
		appErr := &apperr.AppError{
			Code:       "not_found",
			Message:    msg,
			ThaiMsg:    fallbackMsg,
			HTTPStatus: http.StatusNotFound,
		}
		return http.StatusNotFound, appErr.ToResponse()
	}
	message := err.Error()
	if strings.ContainsAny(message, "กขคงจฉชซญดตถทธนบปผฝพพฟภมยรลวศษสหอฮ") && !strings.Contains(message, "SQLSTATE") && len(message) < 1500 {
		appErr := &apperr.AppError{
			Code:       "invalid_request",
			Message:    message,
			ThaiMsg:    message,
			HTTPStatus: http.StatusConflict,
		}
		return http.StatusConflict, appErr.ToResponse()
	}
	fallbackUnavailable := "ระบบบัญชียังไม่พร้อม กรุณาลองใหม่ หากยังไม่สำเร็จให้ผู้ดูแลตรวจการเชื่อมต่อฐานข้อมูล"
	unavailMsg := language.Text("gl_err_unavailable", reqLang)
	if unavailMsg == "gl_err_unavailable" {
		unavailMsg = fallbackUnavailable
	}
	appErr := &apperr.AppError{
		Code:       "unavailable",
		Message:    unavailMsg,
		ThaiMsg:    fallbackUnavailable,
		HTTPStatus: http.StatusServiceUnavailable,
	}
	return http.StatusServiceUnavailable, appErr.ToResponse()
}

// fallbackCode keeps every local fail() response machine-readable too.
func fallbackCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "invalid_payload"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "invalid_request"
	default:
		return "error"
	}
}

// decodeCommand reads exactly one command from the request body. Unknown fields
// are rejected so a typo (or a field from a newer/older client) never reaches the
// ledger silently; trailing data is rejected as well.
var errMultipleCommands = errors.New("ส่งข้อมูลบัญชีได้ครั้งละหนึ่งคำสั่ง")

// decodeFailure is the response for a body the ledger cannot read: a user-caused
// 400 with the stable invalid_payload code and a translated message (the raw English
// decoder text is logged, never shown).
func decodeFailure(err error, lang ...string) errorPayload {
	reqLang := "th"
	if len(lang) > 0 && lang[0] != "" {
		reqLang = lang[0]
	}
	msg := decodeErrorMessage(err, reqLang)
	appErr := &apperr.AppError{
		Code:       "invalid_payload",
		Message:    msg,
		ThaiMsg:    decodeErrorMessage(err, "th"),
		HTTPStatus: http.StatusBadRequest,
	}
	return appErr.ToResponse()
}

// decodeErrorMessage explains a JSON decode failure with language fallback without leaking the
// decoder's English text, its struct names or field types.
func decodeErrorMessage(err error, lang ...string) string {
	reqLang := "th"
	if len(lang) > 0 && lang[0] != "" {
		reqLang = lang[0]
	}
	fallbackInvalid := "รูปแบบข้อมูลบัญชีไม่ถูกต้อง"
	invalidMsg := language.Text("gl_err_invalid_payload", reqLang)
	if invalidMsg == "gl_err_invalid_payload" {
		invalidMsg = fallbackInvalid
	}

	if err == nil {
		return invalidMsg
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			if isAmountField(typeErr.Field) {
				return getAmountFieldMessage(reqLang)
			}
			if reqLang == "th" {
				return fmt.Sprintf("รูปแบบข้อมูลบัญชีไม่ถูกต้อง: ชนิดข้อมูลของฟิลด์ %s ไม่ถูกต้อง", typeErr.Field)
			}
			return fmt.Sprintf("%s: %s", invalidMsg, typeErr.Field)
		}
		if reqLang == "th" {
			return "รูปแบบข้อมูลบัญชีไม่ถูกต้อง: ชนิดข้อมูลไม่ถูกต้อง"
		}
		return invalidMsg
	}
	message := err.Error()
	if _, after, ok := strings.Cut(message, "unknown field "); ok {
		field := strings.Trim(strings.TrimSpace(after), `"`)
		if reqLang == "th" {
			return fmt.Sprintf("รูปแบบข้อมูลบัญชีไม่ถูกต้อง: ไม่รองรับฟิลด์ %s", field)
		}
		return fmt.Sprintf("%s: %s", invalidMsg, field)
	}
	if isAmountField(message) {
		return getAmountFieldMessage(reqLang)
	}
	return invalidMsg
}

const amountFieldMessage = "รูปแบบข้อมูลบัญชีไม่ถูกต้อง จำนวนเงินต้องส่งเป็นข้อความทศนิยม"

func getAmountFieldMessage(lang string) string {
	msg := language.Text("gl_err_amount_decimal", lang)
	if msg == "gl_err_amount_decimal" {
		return amountFieldMessage
	}
	return msg
}

func isAmountField(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "amount") || strings.Contains(lower, "debit") || strings.Contains(lower, "credit") {
		return true
	}
	return strings.Contains(text, "จำนวนเงิน")
}

func decodeCommand(reader io.Reader) (gl.Command, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var cmd gl.Command
	if err := decoder.Decode(&cmd); err != nil {
		return gl.Command{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return gl.Command{}, errMultipleCommands
	}
	return cmd, nil
}

func fail(request microservice.IContext, status int, message string) error {
	lang := getRequestLanguage(request)
	translated := message
	switch message {
	case "กรุณาเลือกบริษัทก่อนใช้งานบัญชี":
		translated = language.Text("gl_err_select_company", lang)
	case "ไม่มีสิทธิ์เข้าใช้บริษัทนี้":
		translated = language.Text("gl_err_company_forbidden", lang)
	case "ไม่มีสิทธิ์เข้าใช้สาขานี้":
		translated = language.Text("gl_err_branch_forbidden", lang)
	case "ไม่มีสิทธิ์อ่านข้อมูลบัญชีนี้":
		translated = language.Text("gl_err_read_forbidden", lang)
	case "ไม่มีสิทธิ์อ่านรายการบัญชีนี้":
		translated = language.Text("gl_err_read_forbidden", lang)
	case "ไม่มีสิทธิ์อ่านรายงานบัญชีนี้":
		translated = language.Text("gl_err_report_forbidden", lang)
	case "ไม่มีสิทธิ์ทำรายการบัญชีนี้":
		translated = language.Text("gl_err_action_forbidden", lang)
	case "เปลี่ยนสมุดรายวันหรือประเภทรายการเดิมไม่ได้":
		translated = language.Text("gl_err_journal_book_immutable", lang)
	case "ส่งข้อมูลบัญชีได้ครั้งละหนึ่งคำสั่ง":
		translated = language.Text("gl_err_multiple_commands", lang)
	case "รหัสกลุ่มบริษัทไม่ถูกต้อง":
		translated = language.Text("gl_err_holding_invalid", lang)
	default:
		if strings.HasPrefix(message, "gl_err_") {
			translated = language.Text(message, lang)
		}
	}
	if translated == "" || strings.HasPrefix(translated, "gl_err_") {
		translated = message
	}

	appErr := &apperr.AppError{
		Code:       fallbackCode(status),
		Message:    translated,
		ThaiMsg:    message,
		HTTPStatus: status,
	}
	request.Response(status, appErr.ToResponse())
	return nil
}

func failure(request microservice.IContext, err error) error {
	lang := getRequestLanguage(request)
	status, payload := errorPayloadFor(err, lang)
	if status >= 500 {
		log.Printf("[gl/v2] ledger failure: %v", err)
	}
	request.Response(status, payload)
	return nil
}

func (h *Http) begin(request microservice.IContext) (context.Context, context.CancelFunc, requestScope, error) {
	ctx, cancel := context.WithTimeout(request.Request().Context(), 30*time.Second)
	scope, err := h.scope(ctx, request)
	if err == nil {
		err = h.initialize(ctx)
	}
	return ctx, cancel, scope, err
}

func pageNumber(value string, fallback, max int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return fallback
	}
	if n > max {
		return max
	}
	return n
}
func (h *Http) startRead(ctx context.Context, scope gl.Scope, request microservice.IContext) (int64, error) {
	if err := h.store.Ready(ctx, scope); err != nil {
		return 0, err
	}
	v, err := h.pg.Version(ctx, scope)
	if err != nil {
		return 0, err
	}
	if raw := request.QueryParam("snapshot"); raw != "" {
		expected, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || expected != v {
			return 0, fmt.Errorf("ข้อมูลเปลี่ยนระหว่างส่งออก กรุณาโหลดและส่งออกใหม่")
		}
	}
	return v, nil
}
func (h *Http) endRead(ctx context.Context, scope gl.Scope, version int64) error {
	v, err := h.pg.Version(ctx, scope)
	if err != nil {
		return err
	}
	if v != version {
		return fmt.Errorf("ข้อมูลเปลี่ยนระหว่างอ่าน กรุณาโหลดข้อมูลใหม่")
	}
	return nil
}

func (h *Http) list(request microservice.IContext) error {
	ctx, cancel, scope, err := h.begin(request)
	defer cancel()
	if err != nil {
		return failure(request, err)
	}
	resource := request.Param("resource")
	screen := resourceScreens[resource]
	filters := gl.ListFilter{BookCode: request.QueryParam("bookcode"), Kind: request.QueryParam("kind"), Status: request.QueryParam("status")}
	if resource == "journals" {
		screen = journalScreen(filters.BookCode, filters.Kind, filters.Status)
	}
	lookup := (resource == "accounts" || resource == "fiscal-years") && anyLedger(scope.Permissions)
	if !allowed(scope.Permissions, screen, "") && !lookup && !allowed(scope.Permissions, "data-backup-export", "") {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านข้อมูลบัญชีนี้")
	}
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	page, err := h.pg.List(ctx, scope.Scope, resource, request.QueryParam("q"), pageNumber(request.QueryParam("page"), 1, 1000000), pageNumber(request.QueryParam("limit"), 100, 1000), filters)
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	page.Sequence = version
	return response(request, page)
}

func (h *Http) get(request microservice.IContext) error {
	ctx, cancel, scope, err := h.begin(request)
	defer cancel()
	if err != nil {
		return failure(request, err)
	}
	resource := request.Param("resource")
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	data, err := h.pg.Get(ctx, scope.Scope, resource, request.Param("id"))
	if err != nil {
		return failure(request, err)
	}
	screen := resourceScreens[resource]
	operationalRead := false
	if resource == "journals" {
		var j gl.Journal
		if err = json.Unmarshal(data, &j); err != nil {
			return failure(request, err)
		}
		screen = journalScreen(j.BookCode, j.Kind, "")
		operationalRead = (j.Status == "draft" && allowed(scope.Permissions, "gl-post", "")) || (j.Status == "posted" && allowed(scope.Permissions, "gl-unpost", ""))
	}
	if !operationalRead && !allowed(scope.Permissions, screen, "") && !((resource == "accounts" || resource == "fiscal-years") && anyLedger(scope.Permissions)) {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านรายการบัญชีนี้")
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	return response(request, data)
}

func (h *Http) report(request microservice.IContext) error {
	ctx, cancel, scope, err := h.begin(request)
	defer cancel()
	if err != nil {
		return failure(request, err)
	}
	name := request.Param("report")
	if !canReadReport(scope.Permissions, name) {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านรายงานบัญชีนี้")
	}
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	q := gl.ReportQuery{From: request.QueryParam("from"), To: request.QueryParam("to"), FiscalYear: request.QueryParam("fiscalyear"), AccountCode: request.QueryParam("accountcode"), BranchCode: request.QueryParam("branchcode"), DepartmentCode: request.QueryParam("departmentcode"), ProjectCode: request.QueryParam("projectcode"), BookCode: request.QueryParam("bookcode"), Page: pageNumber(request.QueryParam("page"), 1, 1000000), Limit: pageNumber(request.QueryParam("limit"), 100, 1000)}
	data, err := h.pg.Report(ctx, scope.Scope, name, q)
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	data.Sequence = version
	return response(request, data)
}

func (h *Http) command(request microservice.IContext) error {
	ctx, cancel, scope, err := h.begin(request)
	defer cancel()
	if err != nil {
		return failure(request, err)
	}
	cmd, err := decodeCommand(http.MaxBytesReader(request.ResponseWriter(), request.Request().Body, 2<<20))
	if errors.Is(err, errMultipleCommands) {
		log.Printf("[gl/v2] command body rejected: %v", err)
		return fail(request, http.StatusBadRequest, errMultipleCommands.Error())
	}
	if err != nil {
		log.Printf("[gl/v2] command body decode failed: %v", err)
		request.Response(http.StatusBadRequest, decodeFailure(err, getRequestLanguage(request)))
		return nil
	}
	screen := resourceScreens[cmd.Resource]
	action := cmd.Action
	if cmd.Resource == "processes" {
		screen = map[string]string{"close": "financial-close", "year-end": "gl-year-end", "recalculate": "gl-recalculate-posted", "reprocess": "gl-reprocess"}[cmd.Action]
		action = "update"
	}
	if action == "lock" || action == "unlock" {
		action = "update"
	}
	if cmd.Resource == "journals" {
		if cmd.Action == "post" {
			screen = "gl-post"
			action = "update"
		} else if cmd.Action == "reverse" {
			screen = "gl-unpost"
			action = "update"
		} else {
			j := cmd.Journal
			if cmd.ID != "" {
				old, e := h.store.JournalForAuthorization(ctx, scope.Scope, cmd.ID)
				if e != nil {
					return failure(request, e)
				}
				if j != nil && (j.BookCode != old.BookCode || j.Kind != old.Kind) {
					return fail(request, 409, "เปลี่ยนสมุดรายวันหรือประเภทรายการเดิมไม่ได้")
				}
				j = &old
			}
			if j != nil {
				screen = journalScreen(j.BookCode, j.Kind, "")
			}
		}
	}
	if !allowed(scope.Permissions, screen, action) {
		return fail(request, 403, "ไม่มีสิทธิ์ทำรายการบัญชีนี้")
	}
	result, err := h.store.Execute(ctx, scope.Scope, cmd)
	if err != nil {
		return failure(request, err)
	}
	return response(request, result)
}

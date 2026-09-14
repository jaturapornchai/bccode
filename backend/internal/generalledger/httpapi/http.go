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
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/generalledger/kafkatransport"
	access "smlcloudplatform/internal/organization/access"
	branchmodels "smlcloudplatform/internal/organization/branch/models"
	rolemodels "smlcloudplatform/internal/organization/rolepermission/models"
	"smlcloudplatform/pkg/microservice"
)

type Http struct {
	ms             *microservice.Microservice
	pst            microservice.IPersisterMongo
	store          *gl.Store
	pg             *gl.Postgres
	mu             sync.Mutex
	bus            *kafkatransport.Bus
	startupErr     error
	closeResources func()
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	h := newRuntime(ms, cfg)
	ms.RegisterBackgroundWorker(func(ctx context.Context) {
		defer h.closeResources()
		if h.startupErr != nil {
			ms.Logger.Warnf("GL Kafka relay unavailable; check broker and TLS configuration")
			return
		}
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if h.initialize(ctx) == nil {
					if err := h.store.DeliverPending(ctx); err != nil && ctx.Err() == nil {
						ms.Logger.Warnf("GL Kafka relay pending; inspect undelivered event IDs and retry status")
					}
				}
			}
		}
	})
	return h
}

// RegisterProjectionWorker belongs to consumer mode. API mode only relays
// committed MongoDB references; only this worker applies them to PostgreSQL.
func RegisterProjectionWorker(ms *microservice.Microservice, cfg config.IConfig) error {
	h := newRuntime(ms, cfg)
	if h.startupErr != nil {
		h.closeResources()
		return h.startupErr
	}
	ms.RegisterBackgroundWorker(func(ctx context.Context) {
		// Run closes its consumer before the databases and writer close.
		defer h.closeResources()
		h.bus.Run(ctx, func(ctx context.Context, ref gl.EventReference) error {
			if err := h.initialize(ctx); err != nil {
				return err
			}
			return h.store.ApplyReference(ctx, ref)
		}, func(error) {
			ms.Logger.Warnf("GL Kafka projection pending; consumer will retry the uncommitted event")
		})
	})
	return nil
}

func newRuntime(ms *microservice.Microservice, cfg config.IConfig) *Http {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	var poolMu sync.Mutex
	pools := map[string]*sql.DB{}
	pgcfg := cfg.PersisterConfig()
	projection := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		if !regexp.MustCompile(`^[A-Za-z0-9_]{1,63}$`).MatchString(holding) {
			return nil, fmt.Errorf("รหัสกลุ่มบริษัทไม่ถูกต้อง")
		}
		poolMu.Lock()
		defer poolMu.Unlock()
		if db, ok := pools[holding]; ok {
			return db, nil
		}
		address := url.URL{Scheme: "postgres", Host: net.JoinHostPort(pgcfg.Host(), pgcfg.Port()), Path: "/" + strings.ToLower(holding), User: url.UserPassword(pgcfg.Username(), pgcfg.Password())}
		query := url.Values{"sslmode": {pgcfg.SSLMode()}, "TimeZone": {"UTC"}}
		address.RawQuery = query.Encode()
		db, err := sql.Open("postgres", address.String())
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(8)
		db.SetMaxIdleConns(2)
		db.SetConnMaxIdleTime(5 * time.Minute)
		pools[holding] = db
		return db, nil
	})
	// Fixed group: mainapi and worker both run the consumer block in production,
	// so a per-container CONSUMER_GROUP_NAME would apply every event twice.
	bus, err := kafkatransport.New(cfg.MQConfig(), "gl-v2-projection")
	h := &Http{ms: ms, pst: pst, pg: projection, bus: bus, startupErr: err}
	h.closeResources = func() {
		if bus != nil {
			_ = bus.Close()
		}
		poolMu.Lock()
		defer poolMu.Unlock()
		for _, db := range pools {
			_ = db.Close()
		}
	}
	return h
}

func (h *Http) initialize(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.startupErr != nil {
		return h.startupErr
	}
	if h.store != nil {
		return nil
	}
	col, err := h.pst.Exec(ctx, gl.Event{})
	if err != nil {
		return err
	}
	h.store = gl.NewStore(col.Database(), h.pg, h.bus)
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
	result := requestScope{Scope: gl.Scope{Holding: u.HoldingCode, Company: u.BusinessCode, Actor: u.UID}, Permissions: map[string]bool{}}
	if u.HoldingCode == "" || u.BusinessCode == "" || u.CompanyUID == "" {
		return result, fmt.Errorf("กรุณาเลือกบริษัทก่อนใช้งานบัญชี")
	}
	membership, err := access.FindActiveMembership(ctx, h.pst, u, time.Now().UTC())
	if err != nil {
		return result, err
	}
	if err = access.RequireActiveHolding(ctx, h.pst, u.HoldingCode); err != nil {
		return result, err
	}
	if !access.AllowsCompany(membership.AccessScopes, u.CompanyUID) {
		return result, fmt.Errorf("ไม่มีสิทธิ์เข้าใช้บริษัทนี้")
	}
	if !access.AllowsAllBranches(membership.AccessScopes, u.CompanyUID) {
		if !access.AllowsBranch(membership.AccessScopes, u.CompanyUID, u.BranchUID) {
			return result, fmt.Errorf("ไม่มีสิทธิ์เข้าใช้สาขานี้")
		}
		var branch branchmodels.BranchDoc
		if err = h.pst.FindOne(ctx, branchmodels.BranchDoc{}, bson.M{"holdingcode": u.HoldingCode, "companyuid": u.CompanyUID, "branchuid": u.BranchUID, "isdeleted": false}, &branch); err != nil {
			return result, err
		}
		result.Scope.Branch = branch.Code
	}
	role := "USER"
	if membership.Role == authmodels.ROLE_OWNER {
		role = "OWNER"
	} else if membership.Role == authmodels.ROLE_ADMIN {
		role = "ADMIN"
	}
	codes := append([]string{role}, membership.PermissionSets...)
	var records []rolemodels.RolePermissionDoc
	if err = h.pst.Find(ctx, rolemodels.RolePermissionDoc{}, bson.M{"holdingcode": u.HoldingCode, "rolecode": bson.M{"$in": codes}, "isactive": true, "isdeleted": false}, &records, options.Find().SetLimit(int64(len(codes)+1))); err != nil {
		return result, err
	}
	hasRole := false
	for _, r := range records {
		if r.RoleCode == role {
			hasRole = true
		}
		for _, p := range r.Permissions {
			result.Permissions[p] = true
		}
	}
	if (role == "ADMIN" || role == "OWNER") && !hasRole {
		result.Permissions["*"] = true
	}
	return result, nil
}

var resourceScreens = map[string]string{"accounts": "chart-of-accounts", "fiscal-years": "chart-of-accounts", "account-groups": "gl-account-groups", "product-account-groups": "gl-product-account-groups", "mappings": "gl-account-mapping", "budgets": "gl-budget", "periods": "period-lock", "forecast": "cash-flow-forecast", "statement-templates": "financial-statement-designer"}
var reportScreens = map[string]string{"ledger": "general-ledger", "trialbalance": "trial-balance", "pnl": "profit-loss", "balancesheet": "balance-sheet", "cashflow": "cash-flow", "cashflowforecast": "cash-flow-forecast", "financialgraphs": "financial-graphs", "project-pnl": "project-pnl", "dimensionpnl": "dimension-pnl", "projectsummary": "project-summary-report", "dashboard": "business-dashboard", "executivesummary": "executive-summary", "workingpaper": "working-paper", "daily-check": "daily-info", "annual-balances": "gl-annual-accumulated"}

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

// errorPayload is the gl/v2 error contract. Every failure carries a stable
// machine code in `code` and the Thai text the UI shows in `message`.
type errorPayload struct {
	Success   bool   `json:"success"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	ErrorCode string `json:"errorcode,omitempty"`
}

// errorPayloadFor maps a command error to (HTTP status, error contract body).
// User-caused failures (guards, duplicate codes, validation) are reported as
// 409 (or 404 for missing records) with a Thai message; anything else stays a
// server problem (503) and is logged by failure().
func errorPayloadFor(err error) (int, errorPayload) {
	if errors.Is(err, gl.ErrProjectionPending) {
		return http.StatusConflict, errorPayload{
			Code:      "projection_pending",
			Message:   err.Error(),
			ErrorCode: "GL_PROJECTION_PENDING",
		}
	}
	if user, ok := gl.AsUserError(err); ok {
		return user.HTTPStatus(), errorPayload{Code: user.Code, Message: user.Error()}
	}
	if errors.Is(err, gl.ErrNotFound) {
		return http.StatusNotFound, errorPayload{Code: "not_found", Message: "ไม่พบรายการบัญชีในบริษัทหรือสาขานี้"}
	}
	message := err.Error()
	if strings.ContainsAny(message, "กขคงจฉชซญดตถทธนบปผฝพพฟภมยรลวศษสหอฮ") && !strings.Contains(message, "SQLSTATE") && len(message) < 1500 {
		return http.StatusConflict, errorPayload{Code: "invalid_request", Message: message}
	}
	return http.StatusServiceUnavailable, errorPayload{
		Code:    "unavailable",
		Message: "ระบบบัญชียังไม่พร้อม กรุณาลองใหม่ หากยังไม่สำเร็จให้ผู้ดูแลตรวจการเชื่อมต่อฐานข้อมูล",
	}
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
// 400 with the stable invalid_payload code and a Thai message (the raw English
// decoder text is logged, never shown).
func decodeFailure(err error) errorPayload {
	return errorPayload{Code: "invalid_payload", Message: decodeErrorMessage(err)}
}

// decodeErrorMessage explains a JSON decode failure in Thai without leaking the
// decoder's English text, its struct names or field types.
func decodeErrorMessage(err error) string {
	if err == nil {
		return "รูปแบบข้อมูลบัญชีไม่ถูกต้อง"
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			if isAmountField(typeErr.Field) {
				return amountFieldMessage
			}
			return fmt.Sprintf("รูปแบบข้อมูลบัญชีไม่ถูกต้อง: ชนิดข้อมูลของฟิลด์ %s ไม่ถูกต้อง", typeErr.Field)
		}
		return "รูปแบบข้อมูลบัญชีไม่ถูกต้อง: ชนิดข้อมูลไม่ถูกต้อง"
	}
	message := err.Error()
	if _, after, ok := strings.Cut(message, "unknown field "); ok {
		// The field itself is not accepted, so always name it; the amount wording
		// must not be used here or the UI blames the value format instead.
		field := strings.Trim(strings.TrimSpace(after), `"`)
		return fmt.Sprintf("รูปแบบข้อมูลบัญชีไม่ถูกต้อง: ไม่รองรับฟิลด์ %s", field)
	}
	if isAmountField(message) {
		return amountFieldMessage
	}
	return "รูปแบบข้อมูลบัญชีไม่ถูกต้อง"
}

const amountFieldMessage = "รูปแบบข้อมูลบัญชีไม่ถูกต้อง จำนวนเงินต้องส่งเป็นข้อความทศนิยม"

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
	request.Response(status, errorPayload{Code: fallbackCode(status), Message: message})
	return nil
}

func failure(request microservice.IContext, err error) error {
	status, payload := errorPayloadFor(err)
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
		request.Response(http.StatusBadRequest, decodeFailure(err))
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

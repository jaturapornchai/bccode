// Package httpapi serves the central report registry: GET /report/v1/registry, POST /report/v1/run and
// POST /report/v1/export (CSV). Company and branch come from the session only; every report menu checks its own
// view permission and print/export need the user permissions report:print / report:export.
package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"smlcloudplatform/internal/config"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"
	"smlcloudplatform/internal/report"
	"smlcloudplatform/pkg/microservice"
)

var validHolding = regexp.MustCompile(`^[A-Za-z0-9_]{1,63}$`)

var errInvalidHolding = errors.New("invalid holding code")

const maxBodyBytes = 1 << 20

// NamesFunc returns the company and branch names printed in the report header
type NamesFunc func(ctx context.Context, holding, company, branch, lang string) (string, string, error)

type Http struct {
	ms       *microservice.Microservice
	registry *report.Registry
	// session resolves the session scope + per-menu permissions (injectable for tests, like transaction/httpapi)
	session func(context.Context, microservice.IContext) (glhttp.SessionAccess, error)
	names   NamesFunc
	now     func() time.Time
}

func connectHolding(holding string) (*sql.DB, error) {
	if !validHolding.MatchString(holding) {
		return nil, errInvalidHolding
	}
	return mypg.PgSqlFastConnect(holding)
}

func NewHttp(ms *microservice.Microservice, cfg config.IConfig) *Http {
	engine := gl.NewPostgres(connectHolding)
	return New(ms, report.NewRegistry(report.GLDefinitions(engine, connectHolding)...), glhttp.ResolveSessionAccess, centralNames)
}

// New builds the handler on an explicit registry, session resolver and name lookup (tests)
func New(ms *microservice.Microservice, registry *report.Registry, session func(context.Context, microservice.IContext) (glhttp.SessionAccess, error), names NamesFunc) *Http {
	return &Http{ms: ms, registry: registry, session: session, names: names, now: time.Now}
}

func (h *Http) RegisterHttp() {
	h.ms.GET("/report/v1/registry", h.Registry)
	h.ms.POST("/report/v1/run", h.Run)
	h.ms.POST("/report/v1/export", h.Export)
}

func requestLanguage(request microservice.IContext) string {
	if lang := request.QueryParam("lang"); lang != "" {
		return language.Normalize(lang)
	}
	if lang := request.Header("Accept-Language"); lang != "" {
		return language.Normalize(lang)
	}
	return "th"
}

// respondError writes a report failure, or a session error in the GL session contract (403 no access, …)
func respondError(request microservice.IContext, err error, lang string) error {
	var failure *report.Failure
	if errors.As(err, &failure) {
		status, body := failure.Response(lang)
		request.Response(status, body)
		return nil
	}
	status, body := glhttp.SessionFailure(err, lang)
	request.Response(status, body)
	return nil
}

// access resolves the session; a company must be chosen before any report
func (h *Http) access(ctx context.Context, request microservice.IContext) (glhttp.SessionAccess, error) {
	user := request.UserInfo()
	if user.HoldingCode == "" || user.BusinessCode == "" {
		return glhttp.SessionAccess{}, report.NewFailure("REPORT_COMPANY_REQUIRED", "")
	}
	return h.session(ctx, request)
}

type registryResponse struct {
	CanPrint  bool                `json:"can_print"`
	CanExport bool                `json:"can_export"`
	Reports   []report.ReportView `json:"reports"`
}

// Registry lists the report menus the user may open (module=gl → GL only)
func (h *Http) Registry(request microservice.IContext) error {
	lang := requestLanguage(request)
	ctx, cancel := context.WithTimeout(request.Request().Context(), 15*time.Second)
	defer cancel()
	access, err := h.access(ctx, request)
	if err != nil {
		return respondError(request, err, lang)
	}
	views := h.registry.Views(request.QueryParam("module"), lang, func(menuID string) bool { return access.Allows(menuID, "") })
	request.Response(http.StatusOK, map[string]interface{}{"success": true, "data": registryResponse{
		CanPrint: access.Allows(report.PermissionPrint, ""), CanExport: access.Allows(report.PermissionExport, ""), Reports: views,
	}})
	return nil
}

// Run returns one page of a report on screen, or every row with printed-page totals when print=true
func (h *Http) Run(request microservice.IContext) error {
	resp, err := h.execute(request, false)
	if err != nil {
		return respondError(request, err, requestLanguage(request))
	}
	request.Response(http.StatusOK, map[string]interface{}{"success": true, "data": resp})
	return nil
}

// Export streams the whole report as CSV (UTF-8 BOM, RFC 4180, formula-safe text cells)
func (h *Http) Export(request microservice.IContext) error {
	lang := requestLanguage(request)
	resp, err := h.execute(request, true)
	if err != nil {
		return respondError(request, err, lang)
	}
	w := request.ResponseWriter()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", report.ContentDisposition("report-"+resp.MenuID+".csv", resp.Header.Title+".csv"))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if err := report.WriteCSV(w, resp, lang); err != nil {
		log.Printf("[report] csv export %s interrupted: %v", resp.MenuID, err)
	}
	return nil
}

// execute returns the response or an error: *report.Failure, or a session error of glhttp
func (h *Http) execute(request microservice.IContext, export bool) (report.RunResponse, error) {
	lang := requestLanguage(request)
	ctx, cancel := context.WithTimeout(request.Request().Context(), 90*time.Second)
	defer cancel()
	var req report.RunRequest
	if err := json.NewDecoder(io.LimitReader(request.Request().Body, maxBodyBytes)).Decode(&req); err != nil {
		return report.RunResponse{}, report.NewFailure("REPORT_BAD_REQUEST", "")
	}
	access, err := h.access(ctx, request)
	if err != nil {
		return report.RunResponse{}, err
	}
	plan, failure := h.registry.Prepare(req, access, access.Scope, export, lang)
	if failure != nil {
		return report.RunResponse{}, failure
	}
	result, failure := plan.Execute(ctx)
	if failure != nil {
		if failure.Code == "REPORT_ENGINE_FAILED" || failure.Code == "REPORT_TIMEOUT" {
			log.Printf("[report] %s failed: %v", req.MenuID, failure.Unwrap())
		}
		return report.RunResponse{}, failure
	}
	return plan.Response(result, h.header(ctx, request, access.Scope, plan, lang), lang), nil
}

func (h *Http) header(ctx context.Context, request microservice.IContext, scope gl.Scope, plan *report.Plan, lang string) report.Header {
	branch := plan.Source.Params["branch"]
	if branch == "" {
		branch = scope.Branch
	}
	company, branchName, err := h.names(ctx, scope.Holding, scope.Company, branch, lang)
	if err != nil {
		log.Printf("[report] header names %s/%s: %v", scope.Holding, scope.Company, err)
	}
	user := request.UserInfo()
	printedBy := user.Name
	if printedBy == "" {
		printedBy = user.Username
	}
	return report.Header{CompanyName: company, BranchName: branchName, PrintedBy: printedBy, PrintedAt: h.now().In(bangkok).Format(time.RFC3339)}
}

var bangkok = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Bangkok"); err == nil {
		return loc
	}
	return time.FixedZone("ICT", 7*60*60)
}()

// centralNames reads company/branch names from the central database, in the request language when present
func centralNames(ctx context.Context, holding, company, branch, lang string) (string, string, error) {
	db, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		return "", "", err
	}
	var companyName, branchName string
	err = db.QueryRowContext(ctx, `SELECT
		COALESCE((SELECT NULLIF(n->>'name','') FROM jsonb_array_elements(CASE WHEN jsonb_typeof(c.names)='array' THEN c.names ELSE '[]'::jsonb END) n WHERE n->>'code'=$3 LIMIT 1), c.name),
		COALESCE((SELECT COALESCE((SELECT NULLIF(n->>'name','') FROM jsonb_array_elements(CASE WHEN jsonb_typeof(b.names)='array' THEN b.names ELSE '[]'::jsonb END) n WHERE n->>'code'=$3 LIMIT 1), b.name)
			FROM branches b WHERE b.holding_code=c.holding_code AND b.company_code=c.code AND b.code=$4), '')
		FROM companies c WHERE c.holding_code=$1 AND c.code=$2`, holding, company, lang, branch).Scan(&companyName, &branchName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	return companyName, branchName, err
}

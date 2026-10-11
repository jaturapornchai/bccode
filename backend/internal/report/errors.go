package report

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/pkg/apperr"
)

// Failure is a report error with a stable REPORT_* code; the message is a languages.tsv key so every language
// gets the same text and the screen can point at the parameter named by Field
type Failure struct {
	Code   string
	Status int
	Key    string
	Field  string
	Args   []string
	cause  error
}

func (f *Failure) Error() string { return f.Code }
func (f *Failure) Unwrap() error { return f.cause }

// NewFailure = the failure of one REPORT_* code pointing at field (HTTP layer errors such as a malformed body)
func NewFailure(code, field string) *Failure { return fail(code, field) }

func fail(code string, field string, args ...string) *Failure {
	spec := failures[code]
	return &Failure{Code: code, Status: spec.status, Key: spec.key, Field: field, Args: args}
}

var failures = map[string]struct {
	status int
	key    string
}{
	"REPORT_UNKNOWN":                   {http.StatusNotFound, "report_error_unknown"},
	"REPORT_NOT_READY":                 {http.StatusConflict, "report_error_not_ready"},
	"REPORT_FORBIDDEN":                 {http.StatusForbidden, "report_error_forbidden"},
	"REPORT_VARIANT_FORBIDDEN":         {http.StatusForbidden, "report_error_variant_forbidden"},
	"REPORT_PRINT_FORBIDDEN":           {http.StatusForbidden, "report_error_print_forbidden"},
	"REPORT_EXPORT_FORBIDDEN":          {http.StatusForbidden, "report_error_export_forbidden"},
	"REPORT_COMPANY_REQUIRED":          {http.StatusBadRequest, "report_error_company_required"},
	"REPORT_BAD_REQUEST":               {http.StatusBadRequest, "report_error_bad_request"},
	"REPORT_PARAM_REQUIRED":            {http.StatusUnprocessableEntity, "report_error_param_required"},
	"REPORT_PARAM_INVALID":             {http.StatusUnprocessableEntity, "report_error_param_invalid"},
	"REPORT_DATE_RANGE_INVALID":        {http.StatusUnprocessableEntity, "report_error_date_range_invalid"},
	"REPORT_SORT_INVALID":              {http.StatusUnprocessableEntity, "report_error_option_invalid"},
	"REPORT_LIMIT_INVALID":             {http.StatusUnprocessableEntity, "report_error_limit_invalid"},
	"REPORT_EXPORT_FORMAT_UNAVAILABLE": {http.StatusUnprocessableEntity, "report_error_export_format_unavailable"},
	"REPORT_EXPORT_TOO_LARGE":          {http.StatusUnprocessableEntity, "report_error_export_too_large"},
	"REPORT_ENGINE_FAILED":             {http.StatusInternalServerError, "report_error_engine"},
	"REPORT_PROJECTION_PENDING":        {http.StatusConflict, "report_error_projection_pending"},
	"REPORT_TIMEOUT":                   {http.StatusGatewayTimeout, "report_error_timeout"},
	"REPORT_INTERNAL":                  {http.StatusInternalServerError, "report_error_internal"},
}

// engineFields maps GL engine error fields (reports.go) to the registry parameter the user can fix
var engineFields = map[string]string{
	"from": "date_from", "to": "date_to", "fiscalyear": "fiscal_year", "accountfrom": "account_from", "accountto": "account_to",
	"docnofrom": "doc_no_from", "docnoto": "doc_no_to", "budgetcode": "budget_from", "budgetcodeto": "budget_to",
	"poststatus": "post_status", "departmentcode": "departments", "projectcode": "projects",
	"allocatecode": "allocates", "jobcode": "jobs", "partcode": "parts", "sidecode": "sides",
	"bookcode": "books", "branchcode": "branch", "orderby": "sort",
}

// fromEngine classifies an error of a report source. User errors of the GL engine keep their gl_err_<code> text
// (with args); timeouts and pending projections get their own code; anything else is an engine failure whose
// detail goes to the log only
func fromEngine(err error) *Failure {
	var f *Failure
	if errors.As(err, &f) {
		return f
	}
	if user, ok := gl.AsUserError(err); ok {
		code := "REPORT_PARAM_INVALID"
		if user.Code == "report_range_outside_fiscal_year" || user.Code == "report_date_invalid" {
			code = "REPORT_DATE_RANGE_INVALID"
		}
		field := engineFields[user.Field]
		out := fail(code, field, user.Args...)
		out.Key = "gl_err_" + user.Code
		out.cause = err
		return out
	}
	if errors.Is(err, gl.ErrProjectionPending) {
		return &Failure{Code: "REPORT_PROJECTION_PENDING", Status: http.StatusConflict, Key: "report_error_projection_pending", cause: err}
	}
	var pqErr *pq.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &pqErr) && pqErr.Code == "57014" {
		out := fail("REPORT_TIMEOUT", "")
		out.cause = err
		return out
	}
	out := fail("REPORT_ENGINE_FAILED", "")
	out.cause = err
	return out
}

// Response renders a failure in the standard error envelope (apperr.Response) for the request language
func (f *Failure) Response(lang string) (int, apperr.Response) {
	message := fill(language.Text(f.Key, lang), f.Args)
	thai := fill(language.Text(f.Key, "th"), f.Args)
	appErr := apperr.New(f.Code, f.Status, message, thai).WithField(f.Field)
	return f.Status, appErr.ToResponse()
}

// fill replaces {0}, {1}, … with args (languages.tsv placeholder convention)
func fill(text string, args []string) string {
	for i, arg := range args {
		text = strings.ReplaceAll(text, "{"+strconv.Itoa(i)+"}", arg)
	}
	return text
}

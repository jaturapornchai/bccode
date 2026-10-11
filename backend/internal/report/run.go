package report

import (
	"context"
	"time"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
)

// Row metadata the registry adds to every source row
const (
	KindKey  = "_kind"  // detail|opening|carry|group_total|day_total|control_total|header
	LabelKey = "_label" // ready-made caption of a non-detail row (master lists)
	BoldKey  = "_bold"  // "true" = print bold although the row is a detail row (control accounts in the chart)

	KindDetail       = "detail"
	KindOpening      = "opening"
	KindCarry        = "carry"
	KindGroupTotal   = "group_total"
	KindDayTotal     = "day_total"
	KindControlTotal = "control_total"
	KindHeader       = "header"
)

const (
	defaultLimit       = 200
	maxLimit           = 1000
	defaultRowsPerPage = 40
	minRowsPerPage     = 5
	maxRowsPerPage     = 500
	// MaxPrintRows = rows one print/export may hold; the source reads MaxPrintRows+1 in the same snapshot to detect more
	MaxPrintRows  = 100000
	screenTimeout = 30 * time.Second
	printTimeout  = 60 * time.Second
)

const (
	PermissionPrint  = "report:print"
	PermissionExport = "report:export"
)

// SourceRequest = one validated run of one definition
type SourceRequest struct {
	Scope   gl.Scope
	Params  map[string]string
	Sort    string
	Page    int
	Limit   int
	All     bool // print/export: every row (up to MaxPrintRows+1) in one snapshot
	Timeout time.Duration
	Lang    string
}

type SourceColumn struct {
	Key     string
	Amount  bool
	Grouped bool
}

// FooterRow = a row printed after the grand total (e.g. working paper net profit); LabelKey + Args make its caption
type FooterRow struct {
	LabelKey string
	Args     []string
	Values   map[string]string
}

type SourceResult struct {
	Columns   []SourceColumn
	Rows      []map[string]string // every row carries KindKey
	Totals    map[string]string   // unrounded stored sums
	Footer    []FooterRow
	TotalRows int64
	AsOf      string
}

// Access is the session permission check (glhttp.SessionAccess)
type Access interface {
	Allows(screen, action string) bool
}

type RunRequest struct {
	MenuID      string            `json:"menu_id"`
	Params      map[string]string `json:"params"`
	Sort        string            `json:"sort"`
	Page        int               `json:"page"`
	Limit       int               `json:"limit"`
	Print       bool              `json:"print"`
	RowsPerPage int               `json:"rows_per_page"`
	Format      string            `json:"format"` // export only
}

// Plan = a run request that passed every check
type Plan struct {
	Def         *Definition
	Variant     *Variant
	Source      SourceRequest
	Print       bool
	RowsPerPage int
}

// Prepare checks permission, readiness, paging and parameters of one run. export = the CSV endpoint
func (r *Registry) Prepare(req RunRequest, access Access, scope gl.Scope, export bool, lang string) (*Plan, *Failure) {
	def, variant, ok := r.Lookup(req.MenuID)
	if !ok {
		return nil, fail("REPORT_UNKNOWN", "menu_id")
	}
	if !access.Allows(variant.MenuID, "") {
		return nil, fail("REPORT_FORBIDDEN", "")
	}
	if def.Status != StatusReady {
		return nil, fail("REPORT_NOT_READY", "")
	}
	if req.Print && !access.Allows(PermissionPrint, "") {
		return nil, fail("REPORT_PRINT_FORBIDDEN", "")
	}
	if export {
		if !access.Allows(PermissionExport, "") {
			return nil, fail("REPORT_EXPORT_FORBIDDEN", "")
		}
		if !contains(def.Exports, req.Format) {
			return nil, fail("REPORT_EXPORT_FORMAT_UNAVAILABLE", "format")
		}
	}
	plan := &Plan{Def: def, Variant: variant, Print: req.Print}
	all := req.Print || export
	if !all {
		page, limit, failure := screenPaging(req.Page, req.Limit)
		if failure != nil {
			return nil, failure
		}
		plan.Source.Page, plan.Source.Limit, plan.Source.Timeout = page, limit, screenTimeout
	} else {
		plan.Source.All, plan.Source.Timeout = true, printTimeout
	}
	if req.Print {
		rows := req.RowsPerPage
		if rows == 0 {
			rows = defaultRowsPerPage
		}
		if rows < minRowsPerPage || rows > maxRowsPerPage {
			return nil, fail("REPORT_LIMIT_INVALID", "rows_per_page")
		}
		plan.RowsPerPage = rows
	}
	params, failure := ResolveParams(def, variant, req.Params)
	if failure != nil {
		return nil, failure
	}
	// a branch session reads only its own branch (the GL engine enforces the same; this gives the field error)
	if scope.Branch != "" && params["branch"] != "" && params["branch"] != scope.Branch {
		return nil, fail("REPORT_FORBIDDEN", "branch")
	}
	sort, failure := checkSort(def, req.Sort)
	if failure != nil {
		return nil, failure
	}
	plan.Source.Scope, plan.Source.Params, plan.Source.Sort, plan.Source.Lang = scope, params, sort, lang
	return plan, nil
}

func screenPaging(page, limit int) (int, int, *Failure) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = defaultLimit
	}
	if page < 1 {
		return 0, 0, fail("REPORT_LIMIT_INVALID", "page")
	}
	if limit < 1 || limit > maxLimit {
		return 0, 0, fail("REPORT_LIMIT_INVALID", "limit")
	}
	return page, limit, nil
}

// Execute runs the plan's source; print/export fail with REPORT_EXPORT_TOO_LARGE above MaxPrintRows
func (p *Plan) Execute(ctx context.Context) (SourceResult, *Failure) {
	result, err := p.Def.Source.Run(ctx, p.Source)
	if err != nil {
		return result, fromEngine(err)
	}
	if p.Source.All && (result.TotalRows > MaxPrintRows || len(result.Rows) > MaxPrintRows) {
		return result, fail("REPORT_EXPORT_TOO_LARGE", "", "100,000")
	}
	return result, nil
}

// ---- run response ----

type ColumnView struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Scale int32  `json:"scale"`
	Group string `json:"group"`
}

type Condition struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Header struct {
	CompanyName string      `json:"company_name"`
	BranchName  string      `json:"branch_name"`
	Title       string      `json:"title"`
	Conditions  []Condition `json:"conditions"`
	PrintedBy   string      `json:"printed_by"`
	PrintedAt   string      `json:"printed_at"`
}

type RunResponse struct {
	MenuID       string              `json:"menu_id"`
	Code         string              `json:"code"`
	Columns      []ColumnView        `json:"columns"`
	Rows         []map[string]string `json:"rows"`
	Totals       map[string]string   `json:"totals"`
	TotalsLabels map[string]string   `json:"totals_labels"`
	Footer       []map[string]string `json:"footer"`
	PageTotals   []PageTotal         `json:"page_totals"`
	TotalRows    int64               `json:"total_rows"`
	Page         int                 `json:"page"`
	Limit        int                 `json:"limit"`
	AsOf         string              `json:"as_of"`
	Header       Header              `json:"header"`
	Warnings     []string            `json:"warnings"`
}

// catalog = the definition's column catalog by key
func (d *Definition) catalog() map[string]Column {
	out := make(map[string]Column, len(d.Columns))
	for _, c := range d.Columns {
		out[c.Key] = c
	}
	return out
}

// columnViews labels the source columns in the request language, keeping the source order
func (d *Definition) columnViews(v *Variant, cols []SourceColumn, lang string) []ColumnView {
	catalog := d.catalog()
	out := make([]ColumnView, 0, len(cols))
	for _, c := range cols {
		spec, ok := catalog[c.Key]
		if !ok {
			// a column the catalog misses still gets a language key, never engine text (registry_test checks coverage)
			spec = Column{Key: c.Key, LabelKey: "report_col_" + c.Key, Type: ColText}
			if c.Amount {
				spec.Type, spec.Scale = ColAmount, amountScale
			}
		}
		if key, ok := v.Labels[c.Key]; ok {
			spec.LabelKey = key
		}
		view := ColumnView{Key: c.Key, Label: language.Text(spec.LabelKey, lang), Type: spec.Type, Scale: spec.Scale}
		if c.Grouped && spec.GroupKey != "" {
			view.Group = language.Text(spec.GroupKey, lang)
		}
		out = append(out, view)
	}
	return out
}

// Response builds the run response; header is filled by the HTTP layer (company/branch names, printed by)
func (p *Plan) Response(result SourceResult, header Header, lang string) RunResponse {
	columns := p.Def.columnViews(p.Variant, result.Columns, lang)
	isColumn := map[string]bool{}
	for _, c := range columns {
		isColumn[c.Key] = true
	}
	labels := map[string]string{}
	for _, t := range p.Def.Totals {
		if _, present := result.Totals[t.Key]; present && !isColumn[t.Key] {
			labels[t.Key] = language.Text(t.LabelKey, lang)
		}
	}
	rows := result.Rows
	if rows == nil {
		rows = []map[string]string{}
	}
	resp := RunResponse{MenuID: p.Variant.MenuID, Code: p.Def.Code, Columns: columns, Rows: rows, Totals: grandTotals(p.Def.Totals, result.Totals), TotalsLabels: labels,
		Footer: footerRows(result.Footer, columns, lang), PageTotals: []PageTotal{}, TotalRows: result.TotalRows, Page: p.Source.Page, Limit: p.Source.Limit, AsOf: result.AsOf, Header: header, Warnings: []string{}}
	if p.Source.All {
		resp.Page, resp.Limit = 1, len(rows)
	}
	if p.Print {
		resp.PageTotals = pageTotals(rows, p.Def.Totals, p.RowsPerPage)
	}
	resp.Header.Title = language.Text(p.Variant.NameKey, lang)
	resp.Header.Conditions = p.conditions(lang)
	return resp
}

// footerRows renders footer captions and rounds their amounts once
func footerRows(footer []FooterRow, columns []ColumnView, lang string) []map[string]string {
	amount := map[string]bool{}
	for _, c := range columns {
		amount[c.Key] = c.Type == ColAmount || c.Type == ColQty
	}
	out := []map[string]string{}
	for _, f := range footer {
		row := map[string]string{"label": fill(language.Text(f.LabelKey, lang), f.Args)}
		for key, value := range f.Values {
			if amount[key] && value != "" {
				value = RoundAmount(value, amountScale)
			}
			row[key] = value
		}
		out = append(out, row)
	}
	return out
}

// conditions = the parameters printed under the report title: offered, non-empty, enum values by their label,
// booleans only when switched on
func (p *Plan) conditions(lang string) []Condition {
	out := []Condition{}
	for _, param := range offered(p.Def, p.Variant) {
		value := p.Source.Params[param.Key]
		if value == "" || param.Type == ParamBool && value != "true" {
			continue
		}
		c := Condition{Key: param.Key, Label: language.Text(param.LabelKey, lang), Value: value, Type: param.Type}
		if param.Type == ParamBool {
			c.Value = ""
		}
		for _, o := range param.Options {
			if o.Value == value {
				c.Value = language.Text(o.LabelKey, lang)
			}
		}
		out = append(out, c)
	}
	return out
}

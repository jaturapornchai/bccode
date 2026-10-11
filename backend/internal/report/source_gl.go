package report

import (
	"context"

	gl "smlcloudplatform/internal/generalledger"
)

// GLEngine is the GL report engine ((*gl.Postgres).Report): one REPEATABLE READ READ ONLY transaction per call
type GLEngine interface {
	Report(ctx context.Context, scope gl.Scope, name string, q gl.ReportQuery) (gl.Report, error)
}

// glSource runs one GL engine report; the definition's parameters map onto gl.ReportQuery
type glSource struct {
	engine GLEngine
	name   string
}

// glKinds = the registry row kind of each GL engine rowtype (unknown types are detail rows)
var glKinds = map[string]string{
	"opening": KindOpening, "carry": KindCarry,
	"total": KindGroupTotal, "doctotal": KindGroupTotal, "budgettotal": KindGroupTotal, "dimtotal": KindGroupTotal,
	"daytotal": KindDayTotal, "control": KindControlTotal,
	"header": KindHeader, "document": KindHeader, "budgetaccount": KindHeader,
}

func glQuery(req SourceRequest) gl.ReportQuery {
	p := req.Params
	flag := func(key string) bool { return p[key] == "true" }
	q := gl.ReportQuery{
		From: p["date_from"], To: p["date_to"], FiscalYear: p["fiscal_year"],
		AccountFrom: p["account_from"], AccountTo: p["account_to"], BranchCode: p["branch"],
		DepartmentCode: p["departments"], ProjectCode: p["projects"], BookCode: p["books"],
		DocNoFrom: p["doc_no_from"], DocNoTo: p["doc_no_to"], PostStatus: p["post_status"],
		BudgetCode: p["budget_code"], BudgetCodeTo: p["budget_to"],
		Summary: flag("summary"), Gross: flag("gross"), HideOpening: flag("hide_opening"), HideZero: flag("hide_zero"),
		ControlTotals: flag("control_totals"), DayTotals: flag("day_totals"),
		ByDepartment: flag("by_department"), ByProject: flag("by_project"), DimTotals: flag("dim_totals"),
		OrderBy: req.Sort, Page: req.Page, Limit: req.Limit, Timeout: req.Timeout,
	}
	if p["budget_from"] != "" {
		q.BudgetCode = p["budget_from"]
	}
	if req.All {
		q.RowCap = MaxPrintRows + 1
	}
	return q
}

func (s glSource) Run(ctx context.Context, req SourceRequest) (SourceResult, error) {
	rep, err := s.engine.Report(ctx, req.Scope, s.name, glQuery(req))
	if err != nil {
		return SourceResult{}, err
	}
	out := SourceResult{Columns: make([]SourceColumn, 0, len(rep.Columns)), Rows: make([]map[string]string, 0, len(rep.Rows)), Totals: rep.Totals, TotalRows: rep.TotalRows, AsOf: rep.AsOf}
	for _, c := range rep.Columns {
		out.Columns = append(out.Columns, SourceColumn{Key: c.Key, Amount: c.Amount, Grouped: c.Group != ""})
	}
	for _, row := range rep.Rows {
		copied := make(map[string]string, len(row)+1)
		for key, value := range row {
			copied[key] = value
		}
		kind, ok := glKinds[row["rowtype"]]
		if !ok {
			kind = KindDetail
		}
		copied[KindKey] = kind
		out.Rows = append(out.Rows, copied)
	}
	for _, f := range rep.Footer {
		values := map[string]string{}
		for key, value := range f {
			if key != "labelkey" {
				values[key] = value
			}
		}
		out.Footer = append(out.Footer, FooterRow{LabelKey: f["labelkey"], Values: values})
	}
	return out, nil
}

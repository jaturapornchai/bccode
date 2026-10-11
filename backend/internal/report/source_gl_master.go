package report

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/lib/pq"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
)

// Account groups in BC order (Champ 5520 combo CB_ACCOUNTGROUP_1..5, GLRepChartOfAccountDlg.cpp:183-219)
var accountGroups = []Option{
	{Value: "asset", LabelKey: "gl_asset"}, {Value: "liability", LabelKey: "gl_liability"}, {Value: "equity", LabelKey: "gl_equity"},
	{Value: "income", LabelKey: "gl_revenue"}, {Value: "expense", LabelKey: "gl_expense"},
}

// masterSource lists one GL master register like Champ's master detail reports 5520/5521/5531/5538: code range
// (from alone = that code, both = between, to alone ignored), ORDER BY code. Rows and counts are built here because
// the layout (group runs, control/posting counts) is presentation of a short list, not accounting calculation.
type masterSource struct {
	db    func(holding string) (*sql.DB, error)
	kind  string // gl_records kind
	chart bool   // Champ 5520 layout: account group runs, name display option, control/posting counts
}

type masterRecord struct {
	code, name1, name2, title, group, level string
	posting                                 bool
}

// masterQuery reads at most $6 records of kind $2 in one company; $3/$4 = code range, $5 = account groups
const masterQuery = `SELECT code,
	COALESCE((SELECT NULLIF(n->>'name','') FROM jsonb_array_elements(CASE WHEN jsonb_typeof(payload->'names')='array' THEN payload->'names' ELSE '[]'::jsonb END) n WHERE n->>'code'='th' LIMIT 1), payload->>'name', ''),
	COALESCE((SELECT NULLIF(n->>'name','') FROM jsonb_array_elements(CASE WHEN jsonb_typeof(payload->'names')='array' THEN payload->'names' ELSE '[]'::jsonb END) n WHERE n->>'code'='en' LIMIT 1), payload->>'nameen', ''),
	COALESCE(payload->>'titlename',''), COALESCE(payload->>'accounttype',''), COALESCE(payload->>'level',''),
	COALESCE((payload->>'allowposting')::boolean, true)
FROM gl_records
WHERE company=$1 AND kind=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)
	AND ($3='' OR CASE WHEN $4='' THEN code=$3 ELSE code COLLATE "C">=$3 AND code COLLATE "C"<=$4 END)
	AND (cardinality($5::text[])=0 OR COALESCE(payload->>'accounttype','')=ANY($5::text[]))
ORDER BY code COLLATE "C", id
LIMIT $6`

func (s masterSource) Run(ctx context.Context, req SourceRequest) (SourceResult, error) {
	groups, failure := groupRange(req.Params["group_from"], req.Params["group_to"], s.chart)
	if failure != nil {
		return SourceResult{}, failure
	}
	records, err := s.read(ctx, req, groups)
	if err != nil {
		return SourceResult{}, err
	}
	var result SourceResult
	if s.chart {
		result = chartRows(records, req.Params["show_name"], req.Lang)
	} else {
		result = registerRows(records, s.kind == "journal-books", req.Lang)
	}
	result.TotalRows = int64(len(result.Rows))
	if len(records) > MaxPrintRows {
		result.TotalRows = int64(len(records))
	}
	if !req.All {
		result.Rows = pageSlice(result.Rows, req.Page, req.Limit)
	}
	return result, nil
}

// groupRange = the account types from group_from to group_to (Champ default สินทรัพย์ → ค่าใช้จ่าย)
func groupRange(from, to string, chart bool) ([]string, *Failure) {
	if !chart {
		return []string{}, nil
	}
	index := func(value string, fallback int) int {
		for i, g := range accountGroups {
			if g.Value == value {
				return i
			}
		}
		return fallback
	}
	start, stop := index(from, 0), index(to, len(accountGroups)-1)
	if start > stop {
		return nil, fail("REPORT_PARAM_INVALID", "group_from")
	}
	out := []string{}
	for _, g := range accountGroups[start : stop+1] {
		out = append(out, g.Value)
	}
	return out, nil
}

func (s masterSource) read(ctx context.Context, req SourceRequest, groups []string) ([]masterRecord, error) {
	db, err := s.db(req.Scope.Holding)
	if err != nil {
		return nil, err
	}
	if err = gl.EnsureSchema(ctx, db); err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if req.Timeout > 0 {
		if _, err = tx.ExecContext(ctx, `SELECT set_config('statement_timeout', $1, true)`, strconv.FormatInt(req.Timeout.Milliseconds(), 10)); err != nil {
			return nil, err
		}
	}
	rows, err := tx.QueryContext(ctx, masterQuery, req.Scope.Company, s.kind, req.Params["code_from"], req.Params["code_to"], pq.Array(groups), MaxPrintRows+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []masterRecord{}
	for rows.Next() {
		var r masterRecord
		if err = rows.Scan(&r.code, &r.name1, &r.name2, &r.title, &r.group, &r.level, &r.posting); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func groupLabel(group, lang string) string {
	for _, g := range accountGroups {
		if g.Value == group {
			return language.Text(g.LabelKey, lang)
		}
	}
	return group
}

// chartName = Champ 5520 "แสดงชื่อ": both = "ชื่อ 1 / ชื่อ 2" (GLRepChartOfAccountView.cpp), name1, name2
func chartName(r masterRecord, show string) string {
	switch show {
	case "name1":
		return r.name1
	case "name2":
		return r.name2
	}
	if r.name2 == "" || r.name2 == r.name1 {
		return r.name1
	}
	return r.name1 + " / " + r.name2
}

// chartRows = Champ 5520 (GLRepChartOfAccountView.cpp:140-264): the group shows on the first account of each run of
// one group, a "รวม<หมวด> n รายการ" row closes the run, control accounts print bold, and the report ends with the
// control and posting account counts
func chartRows(records []masterRecord, show, lang string) SourceResult {
	out := SourceResult{Columns: []SourceColumn{{Key: "accountgroup"}, {Key: "accountcode"}, {Key: "accountname"}, {Key: "level"}}, Rows: []map[string]string{}}
	headers, details := 0, 0
	runGroup, runCount := "", 0
	closeRun := func() {
		if runCount == 0 {
			return
		}
		out.Rows = append(out.Rows, map[string]string{KindKey: KindGroupTotal, LabelKey: fill(language.Text("report_master_group_count", lang), []string{groupLabel(runGroup, lang), strconv.Itoa(runCount)})})
	}
	for _, r := range records {
		if r.group != runGroup || runCount == 0 {
			closeRun()
			runGroup, runCount = r.group, 0
		}
		row := map[string]string{KindKey: KindDetail, "accountgroup": "", "accountcode": r.code, "accountname": chartName(r, show), "level": r.level}
		if runCount == 0 {
			row["accountgroup"] = groupLabel(r.group, lang)
		}
		if r.posting {
			details++
		} else {
			headers++
			row[BoldKey] = "true"
		}
		runCount++
		out.Rows = append(out.Rows, row)
	}
	closeRun()
	out.Totals = map[string]string{"count": strconv.Itoa(headers + details), "header_count": strconv.Itoa(headers), "detail_count": strconv.Itoa(details)}
	out.Footer = []FooterRow{
		{LabelKey: "report_master_control_count", Args: []string{strconv.Itoa(headers)}},
		{LabelKey: "report_master_posting_count", Args: []string{strconv.Itoa(details)}},
	}
	return out
}

// registerRows = Champ 5521/5531/5538: code + name (+ voucher title for journal books), "รวมทั้งสิ้น n รายการ"
func registerRows(records []masterRecord, books bool, lang string) SourceResult {
	out := SourceResult{Columns: []SourceColumn{{Key: "code"}, {Key: "name"}}, Rows: []map[string]string{}}
	if books {
		out.Columns = append(out.Columns, SourceColumn{Key: "titlename"})
	}
	for _, r := range records {
		row := map[string]string{KindKey: KindDetail, "code": r.code, "name": r.name1}
		if books {
			row["titlename"] = r.title
		}
		out.Rows = append(out.Rows, row)
	}
	count := strconv.Itoa(len(records))
	out.Totals = map[string]string{"count": count}
	out.Footer = []FooterRow{{LabelKey: "report_master_grand_count", Args: []string{count}}}
	return out
}

func pageSlice(rows []map[string]string, page, limit int) []map[string]string {
	start := (page - 1) * limit
	if start >= len(rows) || start < 0 {
		return []map[string]string{}
	}
	end := start + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

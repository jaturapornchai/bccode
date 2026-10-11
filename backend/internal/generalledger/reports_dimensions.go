package generalledger

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Report filters and groupings of the six analysis registers (design docs/kms/champ/design-2026-10-11/gl-dimensions.md §3.7).
// Champ's report dialogs offer one multi-select per register (GLRepTrialBalanceDlg.cpp:160-190) and group by them
// (GLRepTrialBalanceDlg.cpp:687-715). A filter is a comma-separated code list, one code is the old single value.
//
// Parameters: department and project keep the positions their report already had ($7/$8 of the standard context, $4/$5 of
// the budget-detail context); the other registers are appended after the base parameters, so no base query is renumbered.

// dimensionFilter: the raw comma-separated code list the query carries for the register.
func (q ReportQuery) dimensionFilter(kind string) string {
	switch kind {
	case DepartmentKind:
		return q.DepartmentCode
	case ProjectKind:
		return q.ProjectCode
	case AllocateKind:
		return q.AllocateCode
	case JobKind:
		return q.JobCode
	case PartKind:
		return q.PartCode
	case SideKind:
		return q.SideCode
	}
	return ""
}

// groupsBy: the report lists one row per code of the register ("แยกแผนก" … of Champ's dialogs).
func (q ReportQuery) groupsBy(kind string) bool {
	switch kind {
	case DepartmentKind:
		return q.ByDepartment
	case ProjectKind:
		return q.ByProject
	case AllocateKind:
		return q.ByAllocate
	case JobKind:
		return q.ByJob
	case PartKind:
		return q.ByPart
	case SideKind:
		return q.BySide
	}
	return false
}

// codesOf: the selected codes of the register, never nil (pq.Array(nil) would send NULL, not an empty array).
func (r reportContext) codesOf(kind string) []string {
	if codes := r.codes[kind]; codes != nil {
		return codes
	}
	return []string{}
}

// hasDimensionFilter: some register is narrowed to chosen codes.
func (r reportContext) hasDimensionFilter() bool {
	for _, codes := range r.codes {
		if len(codes) > 0 {
			return true
		}
	}
	return false
}

// parseReportDimensions resolves the multi-value filters (every register, journal book) of a report query.
func (r *reportContext) parseReportDimensions(q ReportQuery) error {
	r.codes = make(map[string][]string, len(dimensionRegistry))
	for _, d := range dimensionRegistry {
		codes, err := parseReportCodeList(q.dimensionFilter(d.kind), d.jsonKey)
		if err != nil {
			return err
		}
		r.codes[d.kind] = codes
	}
	var err error
	r.books, err = parseReportCodeList(q.BookCode, "bookcode")
	return err
}

// addDimensionArgs registers the parameter number of every register's code array. Department and project are already in
// r.args at the given positions; the others are appended now, so call it right after the base arguments are set.
func (r *reportContext) addDimensionArgs(departParam, projectParam int) {
	r.params = map[string]int{DepartmentKind: departParam, ProjectKind: projectParam}
	for _, d := range dimensionRegistry {
		if _, placed := r.params[d.kind]; placed {
			continue
		}
		r.args = append(r.args, pq.Array(r.codesOf(d.kind)))
		r.params[d.kind] = len(r.args)
	}
}

// dimensionConditions = " AND (cardinality($n::text[])=0 OR <expr>=ANY($n::text[]))" for every register, where expr is the
// column (or JSON path) of the register in the query's own table. Every query that runs with r.args must contain them all.
func (r reportContext) dimensionConditions(expr func(d dimensionKind) string) string {
	var out strings.Builder
	for _, d := range dimensionRegistry {
		n := r.params[d.kind]
		fmt.Fprintf(&out, " AND (cardinality($%d::text[])=0 OR %s=ANY($%d::text[]))", n, expr(d), n)
	}
	return out.String()
}

// columnOf / jsonLineKey give dimensionConditions the column of a table alias (alias "" = no prefix) or the key of a journal line object.
func columnOf(alias string) func(d dimensionKind) string {
	return func(d dimensionKind) string { return alias + d.column }
}

func jsonLineKey(alias string) func(d dimensionKind) string {
	return func(d dimensionKind) string { return alias + "->>'" + d.jsonKey + "'" }
}

// dimensionColumnList = "alias.department_code, alias.project_code, …" in registry order.
func dimensionColumnList(alias string) string {
	columns := make([]string, len(dimensionRegistry))
	for i, d := range dimensionRegistry {
		columns[i] = alias + d.column
	}
	return strings.Join(columns, ", ")
}

// dimensionEmptyList = one ” per register ("”, ”, …"), for the heading and total rows of a UNION.
func dimensionEmptyList() string {
	empty := make([]string, len(dimensionRegistry))
	for i := range empty {
		empty[i] = "''"
	}
	return strings.Join(empty, ", ")
}

// dimensionEmptyAliases = "”::text AS departmentcode, …": the first branch of a UNION names (and types) the columns.
func dimensionEmptyAliases() string {
	aliased := make([]string, len(dimensionRegistry))
	for i, d := range dimensionRegistry {
		aliased[i] = "''::text AS " + d.jsonKey
	}
	return strings.Join(aliased, ", ")
}

// dimensionTextColumns = the report columns of the registers, headed by the register names (design OD-6).
func dimensionTextColumns() []ReportColumn {
	columns := make([]ReportColumn, len(dimensionRegistry))
	for i, d := range dimensionRegistry {
		columns[i] = textColumn(d.jsonKey, d.label)
	}
	return columns
}

// budgetActualDimensions = the actual movement of a budget line counts a budget's own dimensions only (” = any code):
// AND (br.<col>=” OR f.<col>=br.<col>) for every register.
func budgetActualDimensions() string {
	var out strings.Builder
	for _, d := range dimensionRegistry {
		fmt.Fprintf(&out, " AND (br.%[1]s='' OR f.%[1]s=br.%[1]s)", d.column)
	}
	return out.String()
}

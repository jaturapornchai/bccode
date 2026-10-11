package report

import "database/sql"

// GL report definitions, wave 1 (docs/kms/champ/report-parity-r0-gl.md). Champ menu ids in ChampMenuID; the
// parameters follow each Champ dialog as far as the GL engine supports them.

func boolParam(key, labelKey, def string) Param {
	return Param{Key: key, Type: ParamBool, LabelKey: labelKey, Default: def}
}

var (
	pDateFrom    = Param{Key: "date_from", Type: ParamDate, LabelKey: "report_param_date_from", Required: true}
	pDateTo      = Param{Key: "date_to", Type: ParamDate, LabelKey: "report_param_date_to", Required: true}
	pFiscalYear  = Param{Key: "fiscal_year", Type: ParamYear, LabelKey: "report_param_fiscal_year", Lookup: "fiscal_year"}
	pAccountFrom = Param{Key: "account_from", Type: ParamCode, LabelKey: "report_param_account_from", Lookup: "account"}
	pAccountTo   = Param{Key: "account_to", Type: ParamCode, LabelKey: "report_param_account_to", Lookup: "account"}
	pBranch      = Param{Key: "branch", Type: ParamCode, LabelKey: "report_param_branch", Lookup: "branch"}
	pDepartments = Param{Key: "departments", Type: ParamCodes, LabelKey: "report_param_departments", Lookup: "department"}
	pProjects    = Param{Key: "projects", Type: ParamCodes, LabelKey: "report_param_projects", Lookup: "project"}
	pAllocates   = Param{Key: "allocates", Type: ParamCodes, LabelKey: "report_param_allocates", Lookup: "allocate"}
	pJobs        = Param{Key: "jobs", Type: ParamCodes, LabelKey: "report_param_jobs", Lookup: "job"}
	pParts       = Param{Key: "parts", Type: ParamCodes, LabelKey: "report_param_parts", Lookup: "part"}
	pSides       = Param{Key: "sides", Type: ParamCodes, LabelKey: "report_param_sides", Lookup: "side"}
	pBooks       = Param{Key: "books", Type: ParamCodes, LabelKey: "report_param_books", Lookup: "book"}
	pDocNoFrom   = Param{Key: "doc_no_from", Type: ParamCode, LabelKey: "report_param_doc_no_from"}
	pDocNoTo     = Param{Key: "doc_no_to", Type: ParamCode, LabelKey: "report_param_doc_no_to"}

	// Champ "สถานะ" combo SHOW_ALL / SHOW_SOME_POSTTRANS / SHOW_SOME_NOTPOSTTRANS (GLRepTransDlg.cpp:83-85)
	postStatusOptions = []Option{
		{Value: "all", LabelKey: "report_post_status_all"},
		{Value: "posted", LabelKey: "report_post_status_posted"},
		{Value: "unposted", LabelKey: "report_post_status_unposted"},
	}
	pHideZero      = boolParam("hide_zero", "report_param_hide_zero", "false")
	pControlTotals = boolParam("control_totals", "report_param_control_totals", "true")
	pByDepartment  = boolParam("by_department", "report_param_by_department", "false")
	pByProject     = boolParam("by_project", "report_param_by_project", "false")
	pByAllocate    = boolParam("by_allocate", "report_param_by_allocate", "false")
	pByJob         = boolParam("by_job", "report_param_by_job", "false")
	pByPart        = boolParam("by_part", "report_param_by_part", "false")
	pBySide        = boolParam("by_side", "report_param_by_side", "false")
	pDimTotals     = boolParam("dim_totals", "report_param_dim_totals", "false")
)

func postStatusParam(def string) Param {
	return Param{Key: "post_status", Type: ParamEnum, LabelKey: "report_param_post_status", Default: def, Options: postStatusOptions}
}

func col(key, labelKey string) Column     { return Column{Key: key, LabelKey: labelKey, Type: ColText} }
func dateCol(key, labelKey string) Column { return Column{Key: key, LabelKey: labelKey, Type: ColDate} }
func amountCol(key, labelKey, groupKey string) Column {
	return Column{Key: key, LabelKey: labelKey, Type: ColAmount, GroupKey: groupKey, Scale: amountScale}
}

// pageAmount = an amount total that is also summed per printed page
func pageAmount(key, labelKey string) TotalKey {
	return TotalKey{Key: key, LabelKey: labelKey, Page: true}
}

var dimensionColumns = []Column{
	col("departmentcode", "report_col_department"),
	col("projectcode", "report_col_project"),
	col("allocatecode", "report_col_allocate"),
	col("jobcode", "report_col_job"),
	col("partcode", "report_col_part"),
	col("sidecode", "report_col_side"),
}

// GLDefinitions returns the wave-1 GL definitions on one engine and one holding database connector
func GLDefinitions(engine GLEngine, db func(holding string) (*sql.DB, error)) []*Definition {
	csv := []string{"csv"}
	journal := &Definition{
		Code: "gl_journal", Module: "gl", Family: "journal", Status: StatusReady, Drill: "journal", Exports: csv,
		Source: glSource{engine: engine, name: "gljournal"},
		Variants: []Variant{
			{MenuID: "gl-daily-report", ChampMenuID: "5532", NameKey: "gl_daily_report", Route: "/report/gljournal", Preset: map[string]string{"summary": "false"}, Landscape: true},
			{MenuID: "gl-daily-summary-report", ChampMenuID: "5533", NameKey: "gl_daily_summary_report", Route: "/report/gljournal-summary", Preset: map[string]string{"summary": "true"}, Omit: []string{"day_totals"},
				Labels: map[string]string{"date": "report_col_doc_date"}},
		},
		Params: []Param{pDateFrom, pDateTo, pDocNoFrom, pDocNoTo, pBooks, pAccountFrom, pAccountTo, pBranch, pDepartments, pProjects, pAllocates, pJobs, pParts, pSides, postStatusParam("all"),
			boolParam("day_totals", "report_param_day_totals", "false"), boolParam("summary", "report_param_summary", "false")},
		Columns: append([]Column{
			dateCol("date", "report_col_date"), col("bookcode", "report_col_book"), col("docno", "report_col_doc_no"),
			dateCol("taxinvoicedate", "report_col_tax_invoice_date"), col("taxinvoiceno", "report_col_tax_invoice_no"), col("referenceno", "report_col_reference"),
			col("accountcode", "report_col_account_chart"), col("accountname", "report_col_account_name"), col("description", "report_col_description"),
			amountCol("debit", "report_col_debit", ""), amountCol("credit", "report_col_credit", ""), col("branchcode", "report_col_branch"),
			amountCol("amount", "report_col_amount", ""),
		}, dimensionColumns...),
		Sorts:       []Option{{Value: "docno", LabelKey: "report_sort_doc_no"}, {Value: "date", LabelKey: "report_sort_doc_date"}},
		DefaultSort: "docno",
		Totals:      []TotalKey{pageAmount("debit", "report_col_debit"), pageAmount("credit", "report_col_credit"), pageAmount("amount", "report_col_amount")},
	}
	ledger := &Definition{
		Code: "gl_ledger", Module: "gl", Family: "ledger", Status: StatusReady, Drill: "journal", Exports: csv,
		Source: glSource{engine: engine, name: "ledger"},
		Variants: []Variant{
			{MenuID: "general-ledger", ChampMenuID: "5534", NameKey: "general_ledger", Route: "/report/ledger", Preset: map[string]string{"summary": "false"},
				Omit: []string{"control_totals", "by_department", "by_project", "by_allocate", "by_job", "by_part", "by_side", "dim_totals"}, Landscape: true},
			{MenuID: "general-ledger-summary", ChampMenuID: "5535", NameKey: "general_ledger_summary", Route: "/report/ledger-summary", Preset: map[string]string{"summary": "true"},
				Omit: []string{"date_from", "hide_opening", "day_totals"},
				// account columns named like the trial balance (Champ 5535 GLRepGLSumView)
				Labels: map[string]string{"accountcode": "report_col_account_code", "accountname": "report_col_item"}},
		},
		Params: []Param{pDateFrom, pDateTo, pAccountFrom, pAccountTo, pBooks, pBranch, pDepartments, pProjects, pAllocates, pJobs, pParts, pSides, postStatusParam("all"),
			boolParam("hide_opening", "report_param_hide_opening", "false"), boolParam("day_totals", "report_param_day_totals", "false"),
			pControlTotals, pByDepartment, pByProject, pByAllocate, pByJob, pByPart, pBySide, pDimTotals, boolParam("summary", "report_param_summary", "false")},
		Columns: append([]Column{
			col("accountcode", "report_col_account_chart"), col("accountname", "report_col_account_chart_name"), dateCol("date", "report_col_date"),
			col("bookcode", "report_col_book"), col("docno", "report_col_doc_no"), col("description", "report_col_description"),
			amountCol("debit", "report_col_debit", ""), amountCol("credit", "report_col_credit", ""), amountCol("balance", "report_col_balance", ""),
			col("linedescription", "report_col_line_description"), col("branchcode", "report_col_branch"),
		}, dimensionColumns...),
		Totals: []TotalKey{pageAmount("debit", "report_col_debit"), pageAmount("credit", "report_col_credit"),
			{Key: "opening", LabelKey: "report_col_group_opening"}, {Key: "balance", LabelKey: "report_col_balance"}},
	}
	trialBalance := &Definition{
		Code: "gl_trial_balance", Module: "gl", Family: "trial_balance", Status: StatusReady, Exports: csv,
		Source: glSource{engine: engine, name: "trialbalance"},
		Variants: []Variant{
			{MenuID: "trial-balance", ChampMenuID: "5506", NameKey: "trial_balance", Route: "/report/trialbalance", Preset: map[string]string{"gross": "false", "post_status": "posted"}, Landscape: true},
			{MenuID: "trial-balance-2", ChampMenuID: "5507", NameKey: "trial_balance_2", Route: "/report/trialbalance-2", Preset: map[string]string{"gross": "true"}},
		},
		Params: []Param{pDateFrom, pDateTo, pAccountFrom, pAccountTo, pBranch, pDepartments, pProjects, pAllocates, pJobs, pParts, pSides, postStatusParam("posted"),
			pHideZero, pControlTotals, pByDepartment, pByProject, pByAllocate, pByJob, pByPart, pBySide, pDimTotals, boolParam("gross", "report_param_gross", "false")},
		Columns: append([]Column{
			col("accountcode", "report_col_account_code"), col("accountname", "report_col_item"),
			amountCol("openingdebit", "report_col_debit", "report_col_group_opening"), amountCol("openingcredit", "report_col_credit", "report_col_group_opening"),
			amountCol("debit", "report_col_debit", "report_col_group_period"), amountCol("credit", "report_col_credit", "report_col_group_period"),
			amountCol("endingdebit", "report_col_debit", "report_col_group_ending"), amountCol("endingcredit", "report_col_credit", "report_col_group_ending"),
		}, dimensionColumns...),
		Totals: []TotalKey{pageAmount("openingdebit", "report_col_debit"), pageAmount("openingcredit", "report_col_credit"),
			pageAmount("debit", "report_col_debit"), pageAmount("credit", "report_col_credit"),
			pageAmount("endingdebit", "report_col_debit"), pageAmount("endingcredit", "report_col_credit"),
			{Key: "difference", LabelKey: "report_total_difference"}},
	}
	workingPaper := &Definition{
		Code: "gl_working_paper", Module: "gl", Family: "working_paper", Status: StatusReady, Exports: csv,
		Source: glSource{engine: engine, name: "workingpaper"},
		Variants: []Variant{
			{MenuID: "working-paper", ChampMenuID: "5499", NameKey: "working_paper", Route: "/gl/workingpaper", Landscape: true},
		},
		Params: []Param{pDateFrom, pDateTo, pAccountFrom, pAccountTo, pBranch, pDepartments, pProjects, pAllocates, pJobs, pParts, pSides, pHideZero, pControlTotals, pByDepartment, pByProject, pByAllocate, pByJob, pByPart, pBySide, pDimTotals},
		Columns: append([]Column{
			col("accountcode", "report_col_account_code"), col("accountname", "report_col_item"),
			amountCol("endingdebit", "report_col_debit", "report_col_group_trial_balance"), amountCol("endingcredit", "report_col_credit", "report_col_group_trial_balance"),
			amountCol("pnldebit", "report_col_debit", "report_col_group_profit_loss"), amountCol("pnlcredit", "report_col_credit", "report_col_group_profit_loss"),
			amountCol("bsdebit", "report_col_debit", "report_col_group_balance_sheet"), amountCol("bscredit", "report_col_credit", "report_col_group_balance_sheet"),
		}, dimensionColumns...),
		Totals: []TotalKey{pageAmount("endingdebit", "report_col_debit"), pageAmount("endingcredit", "report_col_credit"),
			pageAmount("pnldebit", "report_col_debit"), pageAmount("pnlcredit", "report_col_credit"),
			pageAmount("bsdebit", "report_col_debit"), pageAmount("bscredit", "report_col_credit"),
			{Key: "difference", LabelKey: "report_total_difference"}, {Key: "profit", LabelKey: "gl_wp_net_profit_loss"}},
	}
	budgetColumns := []Column{
		col("accountcode", "report_col_account_code"), col("accountname", "report_col_account_chart_name"), col("budgetcode", "report_col_code"),
		col("budgetname", "report_col_name_details"), col("branchcode", "report_col_branch"), col("departmentcode", "report_col_department"),
		col("projectcode", "report_col_project"), col("allocatecode", "report_col_allocate"), col("jobcode", "report_col_job"),
		col("partcode", "report_col_part"), col("sidecode", "report_col_side"),
		dateCol("startdate", "report_col_start_date"), dateCol("enddate", "report_col_end_date"),
		amountCol("budgetamount", "report_col_amount", ""), amountCol("actualamount", "report_col_budget_actual", ""),
		amountCol("variance", "report_col_variance", ""), amountCol("amount", "report_col_amount", ""),
	}
	budgetComparison := &Definition{
		Code: "gl_budget_comparison", Module: "gl", Family: "budget", Status: StatusReady, Exports: csv,
		Source: glSource{engine: engine, name: "budgetcomparison"},
		Variants: []Variant{
			{MenuID: "budget-comparison-report", ChampMenuID: "5530", NameKey: "budget_comparison_report", Route: "/report/budgetcomparison", Landscape: true},
		},
		Params:  []Param{pDateFrom, pDateTo, {Key: "budget_code", Type: ParamCode, LabelKey: "report_param_budget_code", Lookup: "budget"}, pAccountFrom, pAccountTo, pBranch, pDepartments, pProjects, pAllocates, pJobs, pParts, pSides},
		Columns: budgetColumns,
		Totals:  []TotalKey{pageAmount("budgetamount", "report_col_amount"), pageAmount("actualamount", "report_col_budget_actual"), pageAmount("variance", "report_col_variance")},
	}
	budgetDetail := &Definition{
		Code: "gl_budget_detail", Module: "gl", Family: "budget", Status: StatusReady, Exports: csv,
		Source: glSource{engine: engine, name: "budgetdetail"},
		Variants: []Variant{
			{MenuID: "gl-budget-detail", ChampMenuID: "5525", NameKey: "gl_budget_detail", Route: "/report/budget-detail", Landscape: true},
		},
		Params: []Param{pFiscalYear, {Key: "budget_from", Type: ParamCode, LabelKey: "report_param_budget_from", Lookup: "budget"}, {Key: "budget_to", Type: ParamCode, LabelKey: "report_param_budget_to", Lookup: "budget"},
			pAccountFrom, pAccountTo, pBranch, pDepartments, pProjects, pAllocates, pJobs, pParts, pSides},
		Columns: budgetColumns,
		Totals:  []TotalKey{pageAmount("amount", "report_col_amount"), {Key: "linecount", LabelKey: "report_total_line_count", Int: true}},
	}
	statement := &Definition{
		Code: "gl_statement", Module: "gl", Family: "statement", Status: StatusPlanned,
		Variants: []Variant{
			{MenuID: "financial-statement-designer", ChampMenuID: "5505", NameKey: "financial_statement_designer", Route: "/gl/statement-designer"},
		},
	}
	return []*Definition{
		masterDefinition("gl_journal_book_list", "gl-journal-books-report", "5531", "gl_journal_books_report", "/report/journal-books", "journal-books", "book",
			[]Column{col("code", "report_col_book_code"), col("name", "report_col_book_name"), col("titlename", "report_col_name_details")}, db),
		masterDefinition("gl_department_list", "gl-department-report", "5538", "gl_department_report", "/report/departments", "departments", "department",
			[]Column{col("code", "report_col_department_code"), col("name", "report_col_name_details")}, db),
		masterDefinition("gl_project_list", "gl-project-report", "5521", "gl_project_report", "/report/projects", "projects", "project",
			[]Column{col("code", "report_col_project_code"), col("name", "report_col_name_details")}, db),
		masterDefinition("gl_allocate_list", "gl-allocate-report", "5539", "gl_allocate_report", "/report/allocates", "allocate-codes", "allocate",
			[]Column{col("code", "report_col_allocate_code"), col("name", "report_col_name_details")}, db),
		masterDefinition("gl_job_list", "gl-job-report", "5540", "gl_job_report", "/report/jobs", "job-codes", "job",
			[]Column{col("code", "report_col_job_code"), col("name", "report_col_name_details")}, db),
		masterDefinition("gl_part_list", "gl-part-report", "5541", "gl_part_report", "/report/parts", "part-codes", "part",
			[]Column{col("code", "report_col_part_code"), col("name", "report_col_name_details")}, db),
		masterDefinition("gl_side_list", "gl-side-report", "5542", "gl_side_report", "/report/sides", "side-codes", "side",
			[]Column{col("code", "report_col_side_code"), col("name", "report_col_name_details")}, db),
		budgetDetail,
		chartDefinition(db),
		journal, ledger, budgetComparison, trialBalance, workingPaper, statement,
	}
}

func masterTotals() []TotalKey {
	return []TotalKey{{Key: "count", LabelKey: "report_total_count", Int: true, Page: true}}
}

// masterDefinition = Champ 5521/5531/5538 master detail list: code range only
func masterDefinition(code, menuID, champ, nameKey, route, kind, lookup string, columns []Column, db func(string) (*sql.DB, error)) *Definition {
	return &Definition{
		Code: code, Module: "gl", Family: "master", Status: StatusReady, Exports: []string{"csv"},
		Source:   masterSource{db: db, kind: kind},
		Variants: []Variant{{MenuID: menuID, ChampMenuID: champ, NameKey: nameKey, Route: route}},
		Params: []Param{{Key: "code_from", Type: ParamCode, LabelKey: "report_param_code_from", Lookup: lookup},
			{Key: "code_to", Type: ParamCode, LabelKey: "report_param_code_to", Lookup: lookup}},
		Columns: columns,
		Totals:  masterTotals(),
	}
}

// chartDefinition = Champ 5520 รายงานรายละเอียดผังบัญชี (GLRepChartOfAccountDlg.cpp:26-251)
func chartDefinition(db func(string) (*sql.DB, error)) *Definition {
	d := masterDefinition("gl_chart_of_accounts_list", "gl-chart-of-accounts-report", "5520", "gl_chart_of_accounts_report", "/report/chart-of-accounts", "accounts", "account",
		[]Column{col("accountgroup", "report_col_account_group"), col("accountcode", "report_col_account_code"), col("accountname", "report_col_account_chart_name"), {Key: "level", LabelKey: "report_col_level", Type: ColInt}}, db)
	d.Source = masterSource{db: db, kind: "accounts", chart: true}
	d.Params = append(d.Params,
		Param{Key: "group_from", Type: ParamEnum, LabelKey: "report_param_group_from", Default: "asset", Options: accountGroups},
		Param{Key: "group_to", Type: ParamEnum, LabelKey: "report_param_group_to", Default: "expense", Options: accountGroups},
		Param{Key: "show_name", Type: ParamEnum, LabelKey: "report_param_show_name", Default: "both", Options: []Option{
			{Value: "both", LabelKey: "report_show_name_both"}, {Value: "name1", LabelKey: "report_show_name_1"}, {Value: "name2", LabelKey: "report_show_name_2"}}},
	)
	d.Totals = append(d.Totals, TotalKey{Key: "header_count", LabelKey: "report_total_header_count", Int: true}, TotalKey{Key: "detail_count", LabelKey: "report_total_detail_count", Int: true})
	return d
}

package report

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/registry.json from the registry")

func testRegistry() *Registry { return NewRegistry(GLDefinitions(nil, nil)...) }

func TestRegistryCodesMenusAndRoutesAreUnique(t *testing.T) {
	codes, menus, routes := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, d := range testRegistry().Definitions() {
		if codes[d.Code] {
			t.Errorf("duplicate code %s", d.Code)
		}
		codes[d.Code] = true
		if d.Status == StatusReady && d.Source == nil {
			t.Errorf("%s is ready without a source", d.Code)
		}
		for _, v := range d.Variants {
			if menus[v.MenuID] || routes[v.Route] {
				t.Errorf("duplicate menu %s or route %s", v.MenuID, v.Route)
			}
			menus[v.MenuID], routes[v.Route] = true, true
		}
	}
}

// presets, omitted params and label overrides must name real params/columns, and sorts must contain the default
func TestVariantsReferToDefinitionParamsAndColumns(t *testing.T) {
	for _, d := range testRegistry().Definitions() {
		params, columns := map[string]bool{}, d.catalog()
		for _, p := range d.Params {
			params[p.Key] = true
		}
		for _, v := range d.Variants {
			for key := range v.Preset {
				if !params[key] {
					t.Errorf("%s preset %s is not a param", v.MenuID, key)
				}
			}
			for _, key := range v.Omit {
				if !params[key] {
					t.Errorf("%s omits unknown param %s", v.MenuID, key)
				}
			}
			for key := range v.Labels {
				if _, ok := columns[key]; !ok {
					t.Errorf("%s relabels unknown column %s", v.MenuID, key)
				}
			}
		}
		if len(d.Sorts) > 0 && d.DefaultSort == "" {
			t.Errorf("%s has sorts without a default", d.Code)
		}
		for _, c := range d.Columns {
			if c.Type == ColAmount && c.Scale != amountScale {
				t.Errorf("%s amount column %s scale %d", d.Code, c.Key, c.Scale)
			}
		}
	}
}

// languageRows reads backend/assets/language/languages.tsv: key → columns
func languageRows(t *testing.T) map[string][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "assets", "language", "languages.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string][]string{}
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if cols := strings.Split(line, "\t"); cols[0] != "" {
			rows[cols[0]] = cols
		}
	}
	return rows
}

// usedLanguageKeys = every language key the registry, errors, CSV and master lists can print
func usedLanguageKeys() []string {
	keys := []string{"report_master_control_count", "report_master_posting_count", "report_master_grand_count", "report_master_group_count",
		"gl_report_doc_total", "gl_report_total", "gl_report_day_total_count"}
	for _, f := range failures {
		keys = append(keys, f.key)
	}
	for _, code := range []string{"report_date_invalid", "report_post_status_invalid", "report_fiscal_year_unclear", "report_range_outside_fiscal_year"} {
		keys = append(keys, "gl_err_"+code)
	}
	for _, key := range glRowLabels {
		keys = append(keys, key)
	}
	for _, g := range accountGroups {
		keys = append(keys, g.LabelKey)
	}
	for _, d := range testRegistry().Definitions() {
		for _, v := range d.Variants {
			keys = append(keys, v.NameKey)
			for _, key := range v.Labels {
				keys = append(keys, key)
			}
		}
		for _, p := range d.Params {
			keys = append(keys, p.LabelKey)
			for _, o := range p.Options {
				keys = append(keys, o.LabelKey)
			}
		}
		for _, o := range d.Sorts {
			keys = append(keys, o.LabelKey)
		}
		for _, c := range d.Columns {
			keys = append(keys, c.LabelKey)
			if c.GroupKey != "" {
				keys = append(keys, c.GroupKey)
			}
		}
		for _, total := range d.Totals {
			keys = append(keys, total.LabelKey)
		}
	}
	return keys
}

// every printed key has a full 13-column row with real th/en text and the same {n} placeholders in both
func TestLanguageKeysExistWithThaiAndEnglish(t *testing.T) {
	rows := languageRows(t)
	placeholder := regexp.MustCompile(`\{\d\}`)
	for _, key := range usedLanguageKeys() {
		cols, ok := rows[key]
		if !ok {
			t.Errorf("languages.tsv misses %s", key)
			continue
		}
		if len(cols) != 13 {
			t.Errorf("%s has %d columns, want 13", key, len(cols))
			continue
		}
		for i, value := range cols {
			if strings.TrimSpace(value) == "" {
				t.Errorf("%s column %d is empty", key, i)
			}
		}
		if strings.Join(placeholder.FindAllString(cols[1], -1), ",") != strings.Join(placeholder.FindAllString(cols[2], -1), ",") {
			t.Errorf("%s th/en placeholders differ: %q / %q", key, cols[1], cols[2])
		}
	}
}

// engineColumns = the column keys one GL engine report method declares (textColumn/amountColumn/debitCreditColumns)
func engineColumns(t *testing.T, source, method string) []string {
	t.Helper()
	start := strings.Index(source, ") "+method+"(ctx context.Context")
	if start < 0 {
		t.Fatalf("engine method %s not found", method)
	}
	body := source[start:]
	if end := strings.Index(body[1:], "\nfunc "); end > 0 {
		body = body[:end+1]
	}
	keys := []string{}
	for _, m := range regexp.MustCompile(`(?:textColumn|amountColumn)\("([a-z0-9_]+)"`).FindAllStringSubmatch(body, -1) {
		keys = append(keys, m[1])
	}
	for _, m := range regexp.MustCompile(`debitCreditColumns\("([a-z0-9_]+)", "([a-z0-9_]+)"`).FindAllStringSubmatch(body, -1) {
		keys = append(keys, m[1], m[2])
	}
	if len(keys) == 0 {
		t.Fatalf("engine method %s declares no columns", method)
	}
	return keys
}

// the catalog labels every column the GL engine can return, so no engine Thai label reaches the screen
func TestCatalogCoversEngineColumns(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "generalledger", "reports.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	methods := map[string]string{"gljournal": "glJournal", "ledger": "ledger", "trialbalance": "trialBalance", "workingpaper": "trialBalance", "budgetcomparison": "budgetComparison", "budgetdetail": "budgetDetail"}
	// definitions sharing one engine method (trial balance + working paper) are checked against their joint catalog
	catalogs := map[string]map[string]bool{}
	for _, d := range testRegistry().Definitions() {
		if src, ok := d.Source.(glSource); ok {
			method := methods[src.name]
			if method == "" {
				t.Fatalf("%s: engine report %s has no method mapping", d.Code, src.name)
			}
			if catalogs[method] == nil {
				catalogs[method] = map[string]bool{}
			}
			for key := range d.catalog() {
				catalogs[method][key] = true
			}
		}
	}
	for method, catalog := range catalogs {
		for _, key := range engineColumns(t, source, method) {
			if !catalog[key] {
				t.Errorf("engine %s column %s has no catalog entry", method, key)
			}
		}
	}
}

type goldenEntry struct {
	MenuID      string `json:"menu_id"`
	ChampMenuID string `json:"champ_menu_id"`
	Route       string `json:"route"`
	Code        string `json:"code"`
	Status      string `json:"status"`
}

// testdata/registry.json pins every menu id ↔ route ↔ definition (go test ./internal/report -run Golden -update rewrites it)
func TestRegistryGolden(t *testing.T) {
	entries := []goldenEntry{}
	for _, d := range testRegistry().Definitions() {
		for _, v := range d.Variants {
			entries = append(entries, goldenEntry{MenuID: v.MenuID, ChampMenuID: v.ChampMenuID, Route: v.Route, Code: d.Code, Status: d.Status})
		}
	}
	got, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	path := filepath.Join("testdata", "registry.json")
	if *updateGolden {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ReplaceAll(string(want), "\r\n", "\n") != string(got) {
		t.Fatalf("registry differs from %s (run with -update after an intended change):\n%s", path, got)
	}
}

func TestViewsFilterByModuleAndPermission(t *testing.T) {
	r := testRegistry()
	all := r.Views("gl", "th", func(string) bool { return true })
	if len(all) != 18 {
		t.Fatalf("gl views = %d, want 18", len(all))
	}
	if len(r.Views("ap", "th", func(string) bool { return true })) != 0 {
		t.Fatal("another module must list nothing")
	}
	only := r.Views("gl", "en", func(menu string) bool { return menu == "trial-balance-2" })
	if len(only) != 1 || only[0].Name != "Trial Balance 2" || only[0].ChampMenuID != "5507" {
		t.Fatalf("filtered view = %+v", only)
	}
	for _, p := range only[0].Params {
		if p.Key == "gross" {
			t.Fatal("a preset param must not be offered")
		}
	}
}

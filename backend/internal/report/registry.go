// Package report is the central report registry (docs/kms/champ/design-2026-10-11/report-registry.md):
// one Definition per report family, one Variant per menu item that opens it. Every report menu opens the
// same screen; the screen asks the registry which parameters to show and runs the report by menu id, so
// the variant preset (e.g. "แบบสรุป") is locked server-side and cannot be overridden from the browser.
package report

import (
	"context"

	"smlcloudplatform/internal/goapi/language"
)

const (
	StatusReady   = "ready"
	StatusPlanned = "planned" // listed in the registry so the menu exists, but running it returns REPORT_NOT_READY

	ParamDate  = "date"  // YYYY-MM-DD
	ParamCode  = "code"  // one code (≤50 characters)
	ParamCodes = "codes" // comma-separated codes (≤200)
	ParamEnum  = "enum"  // one of Param.Options
	ParamBool  = "bool"  // "true" | "false"
	ParamYear  = "year"  // fiscal year code

	ColText   = "text"
	ColDate   = "date"
	ColAmount = "amount"
	ColQty    = "qty"
	ColInt    = "int"
)

// Option is one value of an enum parameter or of the sort select; LabelKey is a languages.tsv key
type Option struct {
	Value    string
	LabelKey string
}

type Param struct {
	Key      string
	Type     string
	LabelKey string
	Required bool
	Default  string
	Lookup   string // picker the screen offers: account|branch|book|department|project|budget|fiscal_year
	Options  []Option
}

// Column describes one result column the source may return; the run response keeps the source's column order
// and takes label/type/scale/group from here (GroupKey applies only when the source marks the column grouped)
type Column struct {
	Key      string
	LabelKey string
	Type     string
	GroupKey string
	Scale    int32
}

// TotalKey whitelists one total of the source. Int totals are counts (never rounded); Page totals are also
// summed per printed page over detail rows (Champ "ยอดรวม/ต่อหน้า")
type TotalKey struct {
	Key      string
	LabelKey string
	Int      bool
	Page     bool
}

type Variant struct {
	MenuID      string
	ChampMenuID string
	NameKey     string
	Route       string            // unique per variant: the menu opens a tab keyed by route
	Preset      map[string]string // locked parameter values of this menu
	Omit        []string          // parameters this menu does not offer
	Labels      map[string]string // column label keys of this menu that differ from the definition catalog
	Landscape   bool
}

type Definition struct {
	Code        string
	Module      string
	Family      string
	Status      string
	Variants    []Variant
	Params      []Param
	Columns     []Column
	Sorts       []Option
	DefaultSort string
	Totals      []TotalKey
	Drill       string // "journal" = rows with journalid open the journal
	Exports     []string
	Source      Source
}

// Source produces the rows of one definition inside one read-only snapshot
type Source interface {
	Run(ctx context.Context, req SourceRequest) (SourceResult, error)
}

// Registry is the ordered list of definitions with menu id lookup
type Registry struct {
	defs   []*Definition
	byMenu map[string]menuEntry
}

type menuEntry struct {
	def     *Definition
	variant *Variant
}

func NewRegistry(defs ...*Definition) *Registry {
	r := &Registry{defs: defs, byMenu: map[string]menuEntry{}}
	for _, d := range defs {
		for i := range d.Variants {
			r.byMenu[d.Variants[i].MenuID] = menuEntry{def: d, variant: &d.Variants[i]}
		}
	}
	return r
}

// Lookup finds the definition and variant a menu id opens
func (r *Registry) Lookup(menuID string) (*Definition, *Variant, bool) {
	e, ok := r.byMenu[menuID]
	return e.def, e.variant, ok
}

// Definitions returns the definitions in registry order
func (r *Registry) Definitions() []*Definition { return r.defs }

// ---- registry response ----

type OptionView struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type ParamView struct {
	Key      string       `json:"key"`
	Type     string       `json:"type"`
	Label    string       `json:"label"`
	Required bool         `json:"required"`
	Default  string       `json:"default"`
	Lookup   string       `json:"lookup"`
	Options  []OptionView `json:"options"`
}

// ReportView = one menu item the user may open, with the parameters that menu offers
type ReportView struct {
	MenuID      string       `json:"menu_id"`
	ChampMenuID string       `json:"champ_menu_id"`
	Code        string       `json:"code"`
	Module      string       `json:"module"`
	Family      string       `json:"family"`
	Status      string       `json:"status"`
	Name        string       `json:"name"`
	Route       string       `json:"route"`
	Landscape   bool         `json:"landscape"`
	Exports     []string     `json:"exports"`
	Drill       string       `json:"drill"`
	Params      []ParamView  `json:"params"`
	Sorts       []OptionView `json:"sorts"`
	DefaultSort string       `json:"default_sort"`
}

func options(opts []Option, lang string) []OptionView {
	out := make([]OptionView, 0, len(opts))
	for _, o := range opts {
		out = append(out, OptionView{Value: o.Value, Label: language.Text(o.LabelKey, lang)})
	}
	return out
}

// offered = the parameters a variant shows: definition params minus preset and omitted ones
func offered(d *Definition, v *Variant) []Param {
	out := make([]Param, 0, len(d.Params))
	for _, p := range d.Params {
		if _, locked := v.Preset[p.Key]; locked || contains(v.Omit, p.Key) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// View builds the registry entry of one variant in the request language
func View(d *Definition, v *Variant, lang string) ReportView {
	params := []ParamView{}
	for _, p := range offered(d, v) {
		params = append(params, ParamView{Key: p.Key, Type: p.Type, Label: language.Text(p.LabelKey, lang), Required: p.Required, Default: p.Default, Lookup: p.Lookup, Options: options(p.Options, lang)})
	}
	exports := append([]string{}, d.Exports...)
	return ReportView{MenuID: v.MenuID, ChampMenuID: v.ChampMenuID, Code: d.Code, Module: d.Module, Family: d.Family, Status: d.Status, Name: language.Text(v.NameKey, lang), Route: v.Route, Landscape: v.Landscape, Exports: exports, Drill: d.Drill, Params: params, Sorts: options(d.Sorts, lang), DefaultSort: d.DefaultSort}
}

// Views lists the variants of one module (all modules when module is "") that allow(menuID) accepts
func (r *Registry) Views(module, lang string, allow func(menuID string) bool) []ReportView {
	out := []ReportView{}
	for _, d := range r.defs {
		if module != "" && d.Module != module {
			continue
		}
		for i := range d.Variants {
			if allow(d.Variants[i].MenuID) {
				out = append(out, View(d, &d.Variants[i], lang))
			}
		}
	}
	return out
}

package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// แนะนำบัญชีให้บรรทัดของรูปแบบงบการเงิน (คำสั่ง statement-templates/suggest — คำนวณอย่างเดียว ไม่บันทึกอะไร)
// และตรวจรหัสบัญชีตอนบันทึกแม่แบบ. หลักฐานมาจากข้อมูลที่ผู้ใช้บันทึกไว้เท่านั้น: บัญชีเงินสดในผังบัญชี, บัญชีธนาคาร,
// บัญชีคุมของเอกสารลูกหนี้/เจ้าหนี้, กลุ่มบัญชีสินค้า, บัญชีกำไรสะสมของปีบัญชี, บัญชีในข้อมูลสินทรัพย์ถาวร และบรรทัด
// ชนิดเดียวกันของแม่แบบอื่น — ห้ามเดาจากรหัส ช่วงรหัส หรือชื่อบัญชี เพราะผังบัญชีแต่ละบริษัทต่างกัน; ทุกบัญชีที่แนะนำ
// ผู้ใช้ต้องติ๊กเลือกเอง (ADR docs/kms/decisions/2026-09-27-gl-statement-account-suggestions.md)

// statementSuggestKeys: frontend STATEMENT_SUGGEST_KEYS must match (general-ledger.test.ts reads this line)
var statementSuggestKeys = []string{"cash_and_equivalents", "trade_receivables", "inventories", "property_plant_equipment", "purchase_of_ppe", "trade_payables", "retained_earnings", "sales_revenue", "cost_of_sales", "depreciation_expense"}

// statementSuggestRule: หมวดบัญชีที่บรรทัดชนิดนี้รับ + แหล่งหลักฐาน (ทุกชนิดใช้ other_templates เพิ่มเสมอ)
type statementSuggestRule struct {
	accountType string
	sources     []string
}

// แหล่งหลักฐาน — ลำดับของเหตุผลที่ส่งกลับเรียงตามลำดับค่าคงที่นี้เสมอ (statementSuggestSourceOrder)
const (
	suggestIsCash         = "iscash"
	suggestBank           = "bank_accounts"
	suggestAR             = "ar_documents"
	suggestAP             = "ap_documents"
	suggestProductItem    = "product_item"
	suggestProductRevenue = "product_revenue"
	suggestProductCost    = "product_cost"
	suggestFiscalRetained = "fiscal_retained"
	suggestFAAsset        = "fa_asset"
	suggestFAAccum        = "fa_accum"
	suggestFAExpense      = "fa_expense"
	suggestOtherTemplates = "other_templates"
)

var statementSuggestSourceOrder = []string{suggestIsCash, suggestBank, suggestAR, suggestAP, suggestProductItem, suggestProductRevenue, suggestProductCost, suggestFiscalRetained, suggestFAAsset, suggestFAAccum, suggestFAExpense, suggestOtherTemplates}

var statementSuggestRules = map[string]statementSuggestRule{
	"cash_and_equivalents":     {accountType: "asset", sources: []string{suggestIsCash, suggestBank}},
	"trade_receivables":        {accountType: "asset", sources: []string{suggestAR}},
	"inventories":              {accountType: "asset", sources: []string{suggestProductItem}},
	"property_plant_equipment": {accountType: "asset", sources: []string{suggestFAAsset, suggestFAAccum}},
	"purchase_of_ppe":          {accountType: "asset", sources: []string{suggestFAAsset}},
	"trade_payables":           {accountType: "liability", sources: []string{suggestAP}},
	"retained_earnings":        {accountType: "equity", sources: []string{suggestFiscalRetained}},
	"sales_revenue":            {accountType: "income", sources: []string{suggestProductRevenue}},
	"cost_of_sales":            {accountType: "expense", sources: []string{suggestProductCost}},
	"depreciation_expense":     {accountType: "expense", sources: []string{suggestFAExpense}},
}

// ขนาดสูงสุดที่แนะนำได้ (เฉพาะคำสั่ง suggest ซึ่งใช้สิทธิ์อ่านและผลขยายตามผังบัญชี — การบันทึกแม่แบบไม่จำกัด)
const (
	statementSuggestMaxTargets  = 500
	statementSuggestMaxCodes    = 2000
	statementSuggestMaxFixCodes = 20000
	statementTemplateMaxDepth   = 12 // ระดับบัญชีสูงสุดของผังบัญชี (Account.Validate)
)

type StatementSuggestions struct {
	Targets []StatementSuggestionTarget `json:"targets"`
	Fixes   []StatementAccountFix       `json:"fixes"`
}

type StatementSuggestionTarget struct {
	Target      string                      `json:"target"` // "row" | "column"
	ID          string                      `json:"id"`
	RowNo       int                         `json:"rowno,omitempty"`
	SuggestKey  string                      `json:"suggestkey"`
	AccountType string                      `json:"accounttype"`
	Accounts    []StatementSuggestedAccount `json:"accounts"`
	InUse       []StatementAccountInUse     `json:"inuse"`
}

type StatementSuggestedAccount struct {
	AccountCode string                      `json:"accountcode"`
	AccountName string                      `json:"accountname"` // Account.ThaiName()
	IsActive    bool                        `json:"isactive"`
	Reasons     []StatementSuggestionReason `json:"reasons"`
}

type StatementSuggestionReason struct {
	Source    string   `json:"source"`
	Count     int      `json:"count"`
	Templates []string `json:"templates,omitempty"` // other_templates only: first 5 codes by code order; Count = total
}

type StatementAccountInUse struct {
	AccountCode string                      `json:"accountcode"`
	AccountName string                      `json:"accountname"`
	Target      string                      `json:"target"`
	ID          string                      `json:"id"`
	RowNo       int                         `json:"rowno,omitempty"`
	Title       string                      `json:"title"`
	Reasons     []StatementSuggestionReason `json:"reasons,omitempty"`
}

type StatementAccountFix struct {
	Target  string                     `json:"target"`
	ID      string                     `json:"id"`
	RowNo   int                        `json:"rowno,omitempty"`
	Title   string                     `json:"title"`
	Headers []StatementHeaderExpansion `json:"headers"`
	Removed []string                   `json:"removed"`
}

type StatementHeaderExpansion struct {
	AccountCode string                  `json:"accountcode"`
	AccountName string                  `json:"accountname"`
	Descendants []string                `json:"descendants"` // codes to add, after skipping
	Skipped     []StatementAccountInUse `json:"skipped"`     // posting descendants already in a conflicting target
}

// statementCodeTarget: บรรทัด/คอลัมน์ที่เครื่องคำนวณงบอ่านรหัสบัญชีจริง — ใช้ร่วมกันทั้งตรวจตอนบันทึก แนะนำบัญชี
// แก้บัญชีหัวข้อ และหาบัญชีที่มียอดแต่ไม่อยู่ในงบ (statements_unassigned.go) ต้องตรงกับ statementTemplateReadsAccount (SQL)
type statementCodeTarget struct {
	target     string // row | column
	id         string
	rowNo      int
	title      string
	basis      string // opening | closing | movement | column
	codes      []string
	suggestKey string
	index      int // ตำแหน่งใน rows หรือ columns ของแม่แบบ
}

// statementCodeTargets: งบทั่วไป = ทุกแถวชนิด account (ฐานตาม amountbasis หรือตามชนิดงบ);
// งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น = แถวบัญชีที่เป็นความเคลื่อนไหว + ทุกคอลัมน์ (statements_equity.go)
func statementCodeTargets(m Master) []statementCodeTarget {
	targets := []statementCodeTarget{}
	equity := m.StatementType == "equity"
	for i, row := range m.Rows {
		if row.RowType != "account" {
			continue
		}
		basis := statementTargetBasis(m.StatementType, row.AmountBasis)
		if equity {
			if equityRowBasis(row) != "movement" {
				continue
			}
			basis = "movement"
		}
		targets = append(targets, statementCodeTarget{target: "row", id: row.ID, rowNo: row.RowNo, title: row.Title, basis: basis, codes: row.AccountCodes, suggestKey: row.SuggestKey, index: i})
	}
	if equity {
		for i, column := range m.Columns {
			targets = append(targets, statementCodeTarget{target: "column", id: column.ID, title: column.Title, basis: "column", codes: column.AccountCodes, suggestKey: column.SuggestKey, index: i})
		}
	}
	return targets
}

func statementTargetBasis(statementType, amountBasis string) string {
	switch amountBasis {
	case "opening", "closing", "movement":
		return amountBasis
	}
	if statementPeriodic(statementType) {
		return "movement"
	}
	return "closing"
}

// statementTargetsConflict: บัญชีเดียวกันในสองตำแหน่งนี้ทำให้ยอดนับซ้ำ — ฐานเดียวกันเท่านั้น (ยอดต้นงวดกับปลายงวด
// ของบัญชีเดียวกันอยู่คนละบรรทัดได้) และไม่ใช่งบกระแสเงินสด (วิธีทางอ้อมใช้บัญชีเดียวกันทั้งบรรทัดกำไรและบรรทัดปรับปรุง)
func statementTargetsConflict(statementType string, a, b statementCodeTarget) bool {
	return statementType != "cash_flow" && (a.target != b.target || a.index != b.index) && a.basis == b.basis
}

// statementEvidence: ข้อมูลที่ผู้ใช้บันทึกไว้ในบริษัท ใช้เป็นเหตุผลแนะนำบัญชี
type statementEvidence struct {
	counts    map[string]map[string]int // แหล่ง → รหัสบัญชี → จำนวนข้อมูลที่ผูกบัญชีนี้ในบทบาทนั้น
	templates []Master                  // แม่แบบงบอื่นของบริษัท (ไม่รวมแม่แบบที่กำลังแก้) เรียงตามรหัส
}

func (ev *statementEvidence) add(source, code string, count int) {
	if code == "" || count <= 0 {
		return
	}
	if ev.counts == nil {
		ev.counts = map[string]map[string]int{}
	}
	if ev.counts[source] == nil {
		ev.counts[source] = map[string]int{}
	}
	ev.counts[source][code] += count
}

func (s *PostgresStore) suggestStatementAccounts(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	if cmd.Master == nil {
		return Result{}, fieldError("statement_template_suggest_payload_required", "master", "ไม่พบข้อมูลรูปแบบงบการเงินที่จะใช้แนะนำบัญชี")
	}
	m := *cmd.Master
	if statementSuggestTooLarge(m) {
		return Result{}, statementSuggestTooLargeError()
	}
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return Result{}, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	chart, err := loadStatementChart(ctx, tx, scope.Company)
	if err != nil {
		return Result{}, err
	}
	evidence, err := loadStatementEvidence(ctx, tx, scope.Company, cmd.ID)
	if err != nil {
		return Result{}, err
	}
	out, tooLarge := buildStatementSuggestions(m, chart, evidence)
	if tooLarge {
		return Result{}, statementSuggestTooLargeError()
	}
	return Result{Suggestions: &out}, nil
}

func statementSuggestTooLarge(m Master) bool {
	if len(m.Rows)+len(m.Columns) > statementSuggestMaxTargets {
		return true
	}
	for _, target := range statementCodeTargets(m) {
		if len(target.codes) > statementSuggestMaxCodes {
			return true
		}
	}
	return false
}

func statementSuggestTooLargeError() error {
	e := fieldError("statement_template_suggest_too_large", "master", fmt.Sprintf("รูปแบบงบการเงินนี้ใหญ่เกินกว่าจะแนะนำบัญชีได้ (ไม่เกิน %d บรรทัดรวมคอลัมน์ และไม่เกิน %d บัญชีต่อบรรทัด)", statementSuggestMaxTargets, statementSuggestMaxCodes))
	e.Args = []string{strconv.Itoa(statementSuggestMaxTargets), strconv.Itoa(statementSuggestMaxCodes)}
	return e
}

// buildStatementSuggestions: บัญชีที่แนะนำของทุกตำแหน่งที่มีชนิดสำหรับแนะนำ + รายการแก้บัญชีหัวข้อ/รหัสที่ไม่มีในผัง
// ของทุกตำแหน่งที่เครื่องคำนวณอ่าน; tooLarge = ผลการแก้บัญชีหัวข้อเกิน statementSuggestMaxFixCodes
func buildStatementSuggestions(m Master, chart map[string]Account, ev statementEvidence) (StatementSuggestions, bool) {
	out := StatementSuggestions{Targets: []StatementSuggestionTarget{}, Fixes: []StatementAccountFix{}}
	targets := statementCodeTargets(m)
	positions := map[string][]int{} // รหัสบัญชี → ตำแหน่ง (index ใน targets) ที่มีรหัสนี้ เรียงตามลำดับแม่แบบ
	for i, target := range targets {
		for _, code := range target.codes {
			if list := positions[code]; len(list) == 0 || list[len(list)-1] != i {
				positions[code] = append(list, i)
			}
		}
	}
	others := statementOtherTemplateEvidence(chart, ev.templates)
	for _, target := range targets {
		rule, ok := statementSuggestRules[target.suggestKey]
		if !ok {
			continue // ไม่มีชนิดหรือชนิดที่ไม่รู้จัก: ไม่แนะนำ (การบันทึกแม่แบบตรวจชนิดเอง)
		}
		out.Targets = append(out.Targets, statementSuggestTarget(m.StatementType, target, targets, positions, rule, chart, ev, others[target.suggestKey]))
	}
	fixes, tooLarge := statementAccountFixes(m.StatementType, targets, chart)
	if tooLarge {
		return StatementSuggestions{}, true
	}
	out.Fixes = fixes
	return out, false
}

// statementOtherTemplateEvidence: ชนิดบรรทัด → รหัสบัญชีที่ลงรายการได้ → รหัสแม่แบบอื่นที่ผูกบัญชีนี้ไว้ในบรรทัดชนิดเดียวกัน
// (แม่แบบหนึ่งนับครั้งเดียวต่อบัญชี, เรียงตามรหัสแม่แบบ)
func statementOtherTemplateEvidence(chart map[string]Account, templates []Master) map[string]map[string][]string {
	sets := map[string]map[string]map[string]bool{}
	for _, template := range templates {
		for _, target := range statementCodeTargets(template) {
			if _, ok := statementSuggestRules[target.suggestKey]; !ok {
				continue
			}
			for _, code := range target.codes {
				if account, ok := chart[code]; !ok || !account.AllowPosting {
					continue
				}
				if sets[target.suggestKey] == nil {
					sets[target.suggestKey] = map[string]map[string]bool{}
				}
				if sets[target.suggestKey][code] == nil {
					sets[target.suggestKey][code] = map[string]bool{}
				}
				sets[target.suggestKey][code][template.Code] = true
			}
		}
	}
	out := map[string]map[string][]string{}
	for key, byCode := range sets {
		out[key] = map[string][]string{}
		for code, set := range byCode {
			out[key][code] = slices.Sorted(maps.Keys(set))
		}
	}
	return out
}

func statementSuggestTarget(statementType string, target statementCodeTarget, all []statementCodeTarget, positions map[string][]int, rule statementSuggestRule, chart map[string]Account, ev statementEvidence, others map[string][]string) StatementSuggestionTarget {
	result := StatementSuggestionTarget{Target: target.target, ID: target.id, RowNo: target.rowNo, SuggestKey: target.suggestKey, AccountType: rule.accountType, Accounts: []StatementSuggestedAccount{}, InUse: []StatementAccountInUse{}}
	reasons := map[string][]StatementSuggestionReason{}
	for _, source := range statementSuggestSourceOrder {
		switch {
		case source == suggestOtherTemplates:
			for code, templates := range others {
				reasons[code] = append(reasons[code], StatementSuggestionReason{Source: source, Count: len(templates), Templates: templates[:min(len(templates), 5)]})
			}
		case !contains(rule.sources, source):
		case source == suggestIsCash:
			for code, account := range chart {
				if account.IsCash {
					reasons[code] = append(reasons[code], StatementSuggestionReason{Source: source, Count: 1})
				}
			}
		default:
			for code, count := range ev.counts[source] {
				reasons[code] = append(reasons[code], StatementSuggestionReason{Source: source, Count: count})
			}
		}
	}
	own := map[string]bool{}
	for _, code := range target.codes {
		own[code] = true
	}
	candidates := make([]string, 0, len(reasons))
	for code := range reasons {
		candidates = append(candidates, code)
	}
	sort.Strings(candidates)
	for _, code := range candidates {
		account, ok := chart[code]
		if !ok || !account.AllowPosting || account.AccountType != rule.accountType || own[code] {
			continue
		}
		if holder, found := statementConflictHolder(statementType, target, all, positions[code]); found {
			result.InUse = append(result.InUse, StatementAccountInUse{AccountCode: code, AccountName: account.ThaiName(), Target: holder.target, ID: holder.id, RowNo: holder.rowNo, Title: holder.title, Reasons: reasons[code]})
			continue
		}
		result.Accounts = append(result.Accounts, StatementSuggestedAccount{AccountCode: code, AccountName: account.ThaiName(), IsActive: account.IsActive, Reasons: reasons[code]})
	}
	return result
}

// statementConflictHolder: ตำแหน่งแรกของแม่แบบ (ตามลำดับ) ที่มีบัญชีนี้อยู่แล้วและจะทำให้ยอดนับซ้ำกับ target
func statementConflictHolder(statementType string, target statementCodeTarget, all []statementCodeTarget, holders []int) (statementCodeTarget, bool) {
	for _, index := range holders {
		if statementTargetsConflict(statementType, target, all[index]) {
			return all[index], true
		}
	}
	return statementCodeTarget{}, false
}

// statementAccountFixes: ทุกตำแหน่งที่เครื่องคำนวณอ่าน (มีหรือไม่มีชนิดก็ได้) — บัญชีหัวข้อได้ยอดศูนย์เสมอ
// (ยอดไม่รวมขึ้นบัญชีแม่) จึงเสนอแทนด้วยบัญชีย่อยที่ลงรายการได้ ยกเว้นบัญชีย่อยที่อยู่ในตำแหน่งอื่นที่ชนกันแล้ว
// (ถ้าเพิ่มจะนับซ้ำ); ตำแหน่งหลังที่ขยายบัญชีหัวข้อเดียวกันจะข้ามบัญชีย่อยที่ตำแหน่งก่อนหน้ารับไปแล้ว
func statementAccountFixes(statementType string, targets []statementCodeTarget, chart map[string]Account) ([]StatementAccountFix, bool) {
	fixes := []StatementAccountFix{}
	tree := newStatementTree(chart)
	holders := map[string]map[string][]int{} // ฐาน → รหัสบัญชีที่ลงรายการได้ → ตำแหน่งที่ถือบัญชีนี้
	hold := func(basis, code string, index int) {
		if holders[basis] == nil {
			holders[basis] = map[string][]int{}
		}
		if list := holders[basis][code]; len(list) == 0 || list[len(list)-1] != index {
			holders[basis][code] = append(list, index)
		}
	}
	for i, target := range targets {
		for _, code := range target.codes {
			if account, ok := chart[code]; ok && account.AllowPosting {
				hold(target.basis, code, i)
			}
		}
	}
	total := 0
	for i, target := range targets {
		headers, removed := expandStatementAccountCodes(tree, target.codes)
		if len(headers) == 0 && len(removed) == 0 {
			continue
		}
		fix := StatementAccountFix{Target: target.target, ID: target.id, RowNo: target.rowNo, Title: target.title, Headers: []StatementHeaderExpansion{}, Removed: removed}
		for _, header := range headers {
			expansion := StatementHeaderExpansion{AccountCode: header.code, AccountName: header.name, Descendants: []string{}, Skipped: []StatementAccountInUse{}}
			for _, code := range header.descendants {
				if holder, found := statementConflictHolder(statementType, target, targets, holders[target.basis][code]); found {
					expansion.Skipped = append(expansion.Skipped, StatementAccountInUse{AccountCode: code, AccountName: chart[code].ThaiName(), Target: holder.target, ID: holder.id, RowNo: holder.rowNo, Title: holder.title})
					continue
				}
				expansion.Descendants = append(expansion.Descendants, code)
				hold(target.basis, code, i)
			}
			total += len(expansion.Descendants) + len(expansion.Skipped)
			if total > statementSuggestMaxFixCodes {
				return nil, true
			}
			fix.Headers = append(fix.Headers, expansion)
		}
		fixes = append(fixes, fix)
	}
	return fixes, false
}

type statementHeader struct {
	code        string
	name        string
	descendants []string
}

// statementTree: ผังบัญชีที่ยังไม่ถูกลบของบริษัท + บัญชีลูกตามบัญชีแม่ (parentaccountcode) — ไม่ใช้ช่วงรหัสหรือชื่อ
type statementTree struct {
	chart    map[string]Account
	children map[string][]string
	memo     map[string][]string
}

func newStatementTree(chart map[string]Account) *statementTree {
	children := map[string][]string{}
	for code, account := range chart {
		if account.ParentAccountCode != "" {
			children[account.ParentAccountCode] = append(children[account.ParentAccountCode], code)
		}
	}
	return &statementTree{chart: chart, children: children, memo: map[string][]string{}}
}

// descendants: บัญชีที่ลงรายการได้ทุกระดับใต้บัญชีหัวข้อ (รวมบัญชีที่ปิดใช้งาน) เรียงตามรหัส; กันวนซ้ำและลึกไม่เกิน 12 ระดับ
func (t *statementTree) descendants(header string) []string {
	if cached, ok := t.memo[header]; ok {
		return cached
	}
	found := []string{}
	visited := map[string]bool{header: true}
	var walk func(code string, depth int)
	walk = func(code string, depth int) {
		if depth > statementTemplateMaxDepth {
			return
		}
		for _, child := range t.children[code] {
			if visited[child] {
				continue
			}
			visited[child] = true
			if t.chart[child].AllowPosting {
				found = append(found, child)
			}
			walk(child, depth+1)
		}
	}
	walk(header, 1)
	sort.Strings(found)
	t.memo[header] = found
	return found
}

// expandStatementAccountCodes: แยกรหัสของตำแหน่งหนึ่งเป็นบัญชีหัวข้อ (พร้อมบัญชีย่อยที่ลงรายการได้ แม้ไม่มีเลยก็ส่งกลับ)
// และรหัสที่ไม่มีในผังบัญชีหรือถูกลบแล้ว; บัญชีที่ลงรายการได้และ __current_earnings__ คงเดิม; รักษาลำดับ ตัดรหัสซ้ำ
func expandStatementAccountCodes(tree *statementTree, codes []string) ([]statementHeader, []string) {
	headers := []statementHeader{}
	removed := []string{}
	seen := map[string]bool{}
	for _, code := range codes {
		if code == statementCurrentEarnings || seen[code] {
			continue
		}
		seen[code] = true
		account, ok := tree.chart[code]
		switch {
		case !ok:
			removed = append(removed, code)
		case account.AllowPosting:
		default:
			headers = append(headers, statementHeader{code: code, name: account.ThaiName(), descendants: tree.descendants(code)})
		}
	}
	return headers, removed
}

// checkStatementSuggestKeys: ชนิดสำหรับแนะนำบัญชีต้องเป็นชนิดที่ระบบรู้จัก (ทุกบรรทัดและทุกคอลัมน์)
func checkStatementSuggestKeys(m Master) error {
	for i, row := range m.Rows {
		if _, ok := statementSuggestRules[row.SuggestKey]; row.SuggestKey != "" && !ok {
			e := fieldError("statement_template_suggest_key_invalid", fmt.Sprintf("rows[%d].suggestkey", i), fmt.Sprintf("บรรทัด %d: ไม่รู้จักชนิดสำหรับแนะนำบัญชี “%s”", row.RowNo, row.SuggestKey))
			e.Args = []string{strconv.Itoa(row.RowNo), row.SuggestKey}
			return e
		}
	}
	for i, column := range m.Columns {
		if _, ok := statementSuggestRules[column.SuggestKey]; column.SuggestKey != "" && !ok {
			e := fieldError("statement_template_column_suggest_key_invalid", fmt.Sprintf("columns[%d].suggestkey", i), fmt.Sprintf("คอลัมน์ “%s”: ไม่รู้จักชนิดสำหรับแนะนำบัญชี “%s”", column.Title, column.SuggestKey))
			e.Args = []string{column.Title, column.SuggestKey}
			return e
		}
	}
	return nil
}

// validateStatementTemplateAccounts: ตำแหน่งแรกที่เครื่องคำนวณอ่านแล้วมีรหัสที่ไม่มีในผัง ถูกลบ หรือเป็นบัญชีหัวข้อ
// (ยอดเป็นศูนย์เสมอโดยไม่มีใครรู้) — บัญชีที่ปิดใช้งานยังใช้ได้; รหัสในตำแหน่งที่เครื่องคำนวณไม่อ่านไม่ตรวจและไม่ลบทิ้ง
func validateStatementTemplateAccounts(m Master, accounts map[string]Account) error {
	for _, target := range statementCodeTargets(m) {
		bad := []string{}
		seen := map[string]bool{}
		for _, code := range target.codes {
			if code == statementCurrentEarnings || seen[code] {
				continue
			}
			seen[code] = true
			if account, ok := accounts[code]; !ok || account.IsDeleted || !account.AllowPosting {
				bad = append(bad, code)
			}
		}
		if len(bad) == 0 {
			continue
		}
		joined := strings.Join(bad, ", ")
		if target.target == "column" {
			e := fieldError("statement_template_column_accounts_invalid", fmt.Sprintf("columns[%d].accountcodes", target.index), fmt.Sprintf("คอลัมน์ “%s”: บัญชี %s ไม่มีในผังบัญชี ถูกลบแล้ว หรือเป็นบัญชีหัวข้อที่บันทึกบัญชีไม่ได้ (ยอดจะเป็นศูนย์เสมอ) — กดปุ่ม “แทนบัญชีหัวข้อด้วยบัญชีย่อย” ด้านบน หรือเลือกผังบัญชีของคอลัมน์นี้ใหม่", target.title, joined))
			e.Args = []string{target.title, joined}
			return e
		}
		e := fieldError("statement_template_row_accounts_invalid", fmt.Sprintf("rows[%d].accountcodes", target.index), fmt.Sprintf("บรรทัด %d “%s”: บัญชี %s ไม่มีในผังบัญชี ถูกลบแล้ว หรือเป็นบัญชีหัวข้อที่บันทึกบัญชีไม่ได้ (ยอดจะเป็นศูนย์เสมอ) — กดปุ่ม “แทนบัญชีหัวข้อด้วยบัญชีย่อย” ด้านบน หรือเลือกผังบัญชีของบรรทัดนี้ใหม่", target.rowNo, target.title, joined))
		e.Args = []string{strconv.Itoa(target.rowNo), target.title, joined}
		return e
	}
	return nil
}

// validateMasterStatement: ตรวจชนิดสำหรับแนะนำบัญชีและรหัสบัญชีของรูปแบบงบการเงินตอนสร้าง/แก้ไข (ชนิดอื่นผ่านเสมอ)
func (s *PostgresStore) validateMasterStatement(ctx context.Context, tx *sql.Tx, scope Scope, m *Master) error {
	if m.Kind != "statement-templates" {
		return nil
	}
	if err := checkStatementSuggestKeys(*m); err != nil {
		return err
	}
	lines := []Line{}
	for _, target := range statementCodeTargets(*m) {
		for _, code := range target.codes {
			if code != statementCurrentEarnings {
				lines = append(lines, Line{AccountCode: code})
			}
		}
	}
	if len(lines) == 0 {
		return nil
	}
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, lines)
	if err != nil {
		return err
	}
	return validateStatementTemplateAccounts(*m, accounts)
}

// statementTemplateReadsAccount: เงื่อนไข SQL ว่าแม่แบบ t (แถวใน gl_records) มีรหัสบัญชี $2 ในตำแหน่งที่เครื่องคำนวณงบอ่านจริง
// — ต้องตรงกับ statementCodeTargets (แถว account; งบส่วนของผู้ถือหุ้น = แถวความเคลื่อนไหว + ทุกคอลัมน์)
const statementTemplateReadsAccount = `(EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(t.payload->'rows','[]'::jsonb)) r
           WHERE r->>'rowtype'='account'
             AND (COALESCE(t.payload->>'statementtype','')<>'equity' OR COALESCE(r->>'amountbasis','') NOT IN ('opening','closing','other'))
             AND COALESCE(r->'accountcodes','[]'::jsonb) ? $2)
   OR (t.payload->>'statementtype'='equity' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(t.payload->'columns','[]'::jsonb)) c
           WHERE COALESCE(c->'accountcodes','[]'::jsonb) ? $2)))`

// statementTemplatesReadingAccount: รหัสแม่แบบงบ (ไม่เกิน 5 แม่แบบ ตามรหัส) ที่เครื่องคำนวณอ่านบัญชีนี้
func statementTemplatesReadingAccount(ctx context.Context, tx *sql.Tx, company, accountCode string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT t.code FROM gl_records t WHERE t.company=$1 AND t.kind='statement-templates' AND NOT COALESCE((t.payload->>'isdeleted')::boolean,false) AND `+statementTemplateReadsAccount+` ORDER BY t.code LIMIT 5`, company, accountCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := []string{}
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

// loadStatementChart: ผังบัญชีที่ยังไม่ถูกลบของบริษัท (รวมบัญชีที่ปิดใช้งาน)
func loadStatementChart(ctx context.Context, tx *sql.Tx, company string) (map[string]Account, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='accounts' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	chart := map[string]Account{}
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		var account Account
		if err = json.Unmarshal(data, &account); err != nil {
			return nil, err
		}
		chart[account.AccountCode] = account
	}
	return chart, rows.Err()
}

// loadStatementEvidence: นับข้อมูลที่ผูกบัญชีในแต่ละบทบาท + อ่านแม่แบบงบอื่น (ยกเว้น currentID ที่กำลังแก้ — แถวของมันมาจากคำขอ)
func loadStatementEvidence(ctx context.Context, tx *sql.Tx, company, currentID string) (statementEvidence, error) {
	ev := statementEvidence{counts: map[string]map[string]int{}, templates: []Master{}}
	scanCounts := func(query string, handle func(scan func(...any) error) error) error {
		rows, err := tx.QueryContext(ctx, query, company)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err = handle(rows.Scan); err != nil {
				return err
			}
		}
		return rows.Err()
	}
	err := scanCounts(`SELECT payload->>'gl_account_code', count(*) FROM gl_subledger_bank_accounts WHERE company=$1 AND COALESCE(payload->>'gl_account_code','')<>'' GROUP BY 1`, func(scan func(...any) error) error {
		var code string
		var count int
		if err := scan(&code, &count); err != nil {
			return err
		}
		ev.add(suggestBank, code, count)
		return nil
	})
	if err != nil {
		return ev, err
	}
	// เอกสารลูกหนี้/เจ้าหนี้ของใบสำคัญที่ยังไม่ถูกลบ (ใบที่กลับรายการแล้วยังนับ — บัญชีคุมยังบอกบทบาทของบัญชี)
	err = scanCounts(`SELECT d.ledger, d.control_account_code, count(*) FROM gl_subledger_documents d
    JOIN gl_records r ON r.company=d.company AND r.kind='journals' AND r.id=d.created_journal_id AND NOT COALESCE((r.payload->>'isdeleted')::boolean,false)
   WHERE d.company=$1 GROUP BY d.ledger, d.control_account_code`, func(scan func(...any) error) error {
		var ledger, code string
		var count int
		if err := scan(&ledger, &code, &count); err != nil {
			return err
		}
		switch ledger {
		case "ar":
			ev.add(suggestAR, code, count)
		case "ap":
			ev.add(suggestAP, code, count)
		}
		return nil
	})
	if err != nil {
		return ev, err
	}
	err = scanCounts(`SELECT COALESCE(payload->>'itemaccount',''),COALESCE(payload->>'revenueaccount',''),COALESCE(payload->>'costaccount','') FROM gl_records WHERE company=$1 AND kind='product-account-groups' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, func(scan func(...any) error) error {
		var item, revenue, cost string
		if err := scan(&item, &revenue, &cost); err != nil {
			return err
		}
		ev.add(suggestProductItem, item, 1)
		ev.add(suggestProductRevenue, revenue, 1)
		ev.add(suggestProductCost, cost, 1)
		return nil
	})
	if err != nil {
		return ev, err
	}
	// ไม่ใช้บัญชีกำไรขาดทุนของปีบัญชี: รายการปิดบัญชีลงบัญชีนี้ทั้งสองด้านในใบเดียวกัน ยอดจึงเป็นศูนย์เสมอ
	err = scanCounts(`SELECT COALESCE(payload->>'retainedearningsaccount','') FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, func(scan func(...any) error) error {
		var retained string
		if err := scan(&retained); err != nil {
			return err
		}
		ev.add(suggestFiscalRetained, retained, 1)
		return nil
	})
	if err != nil {
		return ev, err
	}
	// ตาราง fa_records สร้างเมื่อเปิดใช้ระบบสินทรัพย์ถาวรครั้งแรก (internal/fixedasset/records.go) — ไม่มีตาราง = ไม่มีหลักฐาน
	var hasFixedAssets bool
	if err = tx.QueryRowContext(ctx, `SELECT to_regclass('fa_records') IS NOT NULL`).Scan(&hasFixedAssets); err != nil {
		return ev, err
	}
	if hasFixedAssets {
		err = scanCounts(`SELECT COALESCE(payload->>'assetaccountcode',''),COALESCE(payload->>'accumdeprecaccountcode',''),COALESCE(payload->>'deprecexpenseaccountcode','') FROM fa_records WHERE company=$1 AND kind IN ('types','assets') AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, func(scan func(...any) error) error {
			var asset, accum, expense string
			if err := scan(&asset, &accum, &expense); err != nil {
				return err
			}
			ev.add(suggestFAAsset, asset, 1)
			ev.add(suggestFAAccum, accum, 1)
			ev.add(suggestFAExpense, expense, 1)
			return nil
		})
		if err != nil {
			return ev, err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT code, COALESCE(payload->>'statementtype',''), COALESCE(payload->'rows','[]'::jsonb), COALESCE(payload->'columns','[]'::jsonb) FROM gl_records WHERE company=$1 AND kind='statement-templates' AND id<>$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY code`, company, currentID)
	if err != nil {
		return ev, err
	}
	defer rows.Close()
	for rows.Next() {
		var template Master
		var rowsJSON, columnsJSON []byte
		if err = rows.Scan(&template.Code, &template.StatementType, &rowsJSON, &columnsJSON); err != nil {
			return ev, err
		}
		if err = json.Unmarshal(rowsJSON, &template.Rows); err != nil {
			return ev, err
		}
		if err = json.Unmarshal(columnsJSON, &template.Columns); err != nil {
			return ev, err
		}
		ev.templates = append(ev.templates, template)
	}
	return ev, rows.Err()
}

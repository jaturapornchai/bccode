package generalledger

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/lib/pq"
)

// The six line-level analysis dimensions of a journal line. Champ keeps each one in its own register table with a
// Code and a Name (Script\SQLSERVER_Script.sql: BCDepartment :460-468, BCProject :590-598; the other four are edited in the
// same generic form myfrmbasemaster.cpp), carries the code on every line (BCTransSub DepartCode / ProjectCode /
// AllocateCode / JobCode / PartCode / SideCode, :9430-9475) and lets an account narrow the codes it accepts
// (BCChartOfAccount.<k>Separate :9069-9075 + BCAccSeparate :9563-9576, "_ALL" = every code). Branch is the seventh
// Champ switch; BC keeps branch on the journal header (design OD-2) so it is not a register here.
// A journal line may only name an active code of its register (design docs/kms/champ/design-2026-10-11/gl-dimensions.md §3).
const (
	DepartmentKind = "departments"
	ProjectKind    = "projects"
	AllocateKind   = "allocate-codes"
	JobKind        = "job-codes"
	PartKind       = "part-codes"
	SideKind       = "side-codes"
)

// dimensionKind describes one register. Nothing in it is client input: column / jsonKey / flagKey / listKey / allKey are
// spliced into SQL and JSON paths, so they come from the dimensionRegistry table below only.
type dimensionKind struct {
	kind  string
	label string // Thai fallback of the register name (the screen reads gl_<errKey> from languages.tsv)
	// column is the gl_lines / gl_budgets column; jsonKey the key of a code on a line, template, allocation rule, forecast, budget row.
	column, jsonKey string
	// flagKey is the account flag AND the company switch key ("แยก…"); listKey the account's allowed-code list; allKey
	// the account's "เลือกทั้งหมด" box (Champ BCAccSeparate "_ALL").
	flagKey, listKey, allKey string
	errKey                   string // short name inside error codes: gl_err_journal_line_<errKey>_not_found …
	labelKey                 string // gl_<errKey>: register name
	codeMax, nameMax         int
	order                    int // position in Champ's screens (the 5th, branch, is not a register here)
}

// dimensionRegistry is in Champ's order (design §3.2): depart 1, project 2, allocate 3, job 4, part 6, side 7.
var dimensionRegistry = []dimensionKind{
	{kind: DepartmentKind, label: "แผนก", column: "department_code", jsonKey: "departmentcode", flagKey: "departseparate", listKey: "departcodes", allKey: "departallcodes", errKey: "depart", labelKey: "gl_department", codeMax: 10, nameMax: 50, order: 1},
	{kind: ProjectKind, label: "โครงการ", column: "project_code", jsonKey: "projectcode", flagKey: "projectseparate", listKey: "projectcodes", allKey: "projectallcodes", errKey: "project", labelKey: "gl_project", codeMax: 10, nameMax: 50, order: 2},
	{kind: AllocateKind, label: "การจัดสรร", column: "allocate_code", jsonKey: "allocatecode", flagKey: "allocateseparate", listKey: "allocatecodes", allKey: "allocateallcodes", errKey: "allocate", labelKey: "gl_allocate", codeMax: 10, nameMax: 50, order: 3},
	{kind: JobKind, label: "งาน", column: "job_code", jsonKey: "jobcode", flagKey: "jobseparate", listKey: "jobcodes", allKey: "joballcodes", errKey: "job", labelKey: "gl_job", codeMax: 15, nameMax: 255, order: 4},
	{kind: PartKind, label: "หน่วยงาน", column: "part_code", jsonKey: "partcode", flagKey: "partseparate", listKey: "partcodes", allKey: "partallcodes", errKey: "part", labelKey: "gl_part", codeMax: 10, nameMax: 255, order: 6},
	{kind: SideKind, label: "ฝ่าย", column: "side_code", jsonKey: "sidecode", flagKey: "sideseparate", listKey: "sidecodes", allKey: "sideallcodes", errKey: "side", labelKey: "gl_side", codeMax: 10, nameMax: 255, order: 7},
}

var dimensionKinds = func() map[string]dimensionKind {
	byKind := make(map[string]dimensionKind, len(dimensionRegistry))
	for _, d := range dimensionRegistry {
		byKind[d.kind] = d
	}
	return byKind
}()

// IsDimensionKind: the resource is one of the six analysis registers (departments … sides).
func IsDimensionKind(kind string) bool {
	_, ok := dimensionKinds[kind]
	return ok
}

// lineCodePtr points at the code a line carries for the register, nil for an unknown kind. One switch keeps the six
// flat Line fields (the stored journal JSON and every Line{...} literal keep working) behind the registry loops.
func (l *Line) lineCodePtr(kind string) *string {
	switch kind {
	case DepartmentKind:
		return &l.DepartmentCode
	case ProjectKind:
		return &l.ProjectCode
	case AllocateKind:
		return &l.AllocateCode
	case JobKind:
		return &l.JobCode
	case PartKind:
		return &l.PartCode
	case SideKind:
		return &l.SideCode
	}
	return nil
}

// dimensionCode is the code the line names for the register ("" = none).
func (l Line) dimensionCode(kind string) string {
	if p := l.lineCodePtr(kind); p != nil {
		return *p
	}
	return ""
}

// accountDimension points at the three account fields of one register: the flag, the allowed codes and "เลือกทั้งหมด".
type accountDimension struct {
	separate *bool
	codes    *[]string
	all      *bool
}

func (a *Account) accountDimension(kind string) accountDimension {
	switch kind {
	case DepartmentKind:
		return accountDimension{&a.DepartSeparate, &a.DepartCodes, &a.DepartAllCodes}
	case ProjectKind:
		return accountDimension{&a.ProjectSeparate, &a.ProjectCodes, &a.ProjectAllCodes}
	case AllocateKind:
		return accountDimension{&a.AllocateSeparate, &a.AllocateCodes, &a.AllocateAllCodes}
	case JobKind:
		return accountDimension{&a.JobSeparate, &a.JobCodes, &a.JobAllCodes}
	case PartKind:
		return accountDimension{&a.PartSeparate, &a.PartCodes, &a.PartAllCodes}
	case SideKind:
		return accountDimension{&a.SideSeparate, &a.SideCodes, &a.SideAllCodes}
	}
	var none bool
	var noCodes []string
	return accountDimension{&none, &noCodes, &none}
}

// ---- user-facing errors (codes get gl_err_<code> rows in languages.tsv) ----

// dimensionErrorCode builds the error code of one register, e.g. journal_line_job_not_found. Every code it can return is listed by
// DimensionErrorCodes so the language-row test (httpapi) sees them: a source scan cannot read a concatenated literal.
func dimensionErrorCode(prefix string, d dimensionKind, suffix string) string {
	return prefix + d.errKey + suffix
}

// DimensionErrorCodes lists every per-register error code (account_<k>_code_not_found, account_<k>_codes_need_flag and
// journal_line_<k>_not_found / _not_allowed / _not_separate / _disabled and budget_<k>_not_found / _not_separate / _not_allowed
// for each of the six registers).
func DimensionErrorCodes() []string {
	var codes []string
	for _, d := range dimensionRegistry {
		codes = append(codes,
			dimensionErrorCode("account_", d, "_code_not_found"), dimensionErrorCode("account_", d, "_codes_need_flag"),
			dimensionErrorCode("journal_line_", d, "_not_found"), dimensionErrorCode("journal_line_", d, "_not_allowed"),
			dimensionErrorCode("journal_line_", d, "_not_separate"), dimensionErrorCode("journal_line_", d, "_disabled"),
			dimensionErrorCode("budget_", d, "_not_found"), dimensionErrorCode("budget_", d, "_not_separate"), dimensionErrorCode("budget_", d, "_not_allowed"))
	}
	return codes
}

func dimensionNameRequired(kind string) error {
	return fieldError("dimension_name_required", "name", "กรุณาระบุชื่อ"+dimensionKinds[kind].label)
}

func dimensionNameTooLong(kind string, n int) error {
	d := dimensionKinds[kind]
	return fieldError("dimension_name_too_long", "name", fmt.Sprintf("ชื่อ%sยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว) กรุณาย่อชื่อให้สั้นลง", d.label, d.nameMax, n))
}

func accountDimensionCodeNotFound(kind, code string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("ไม่พบ%s “%s” ที่เปิดใช้งานในทะเบียน%s — เลือกรหัสที่มีในทะเบียน หรือเปิดใช้งานรหัสนั้นก่อน", d.label, code, d.label)
	err := fieldError(dimensionErrorCode("account_", d, "_code_not_found"), d.listKey, message)
	err.Args = []string{code}
	return err
}

func accountDimensionCodesNeedFlag(kind string) error {
	d := dimensionKinds[kind]
	return fieldError(dimensionErrorCode("account_", d, "_codes_need_flag"), d.listKey, fmt.Sprintf("กำหนด%sที่ใช้ได้ต้องเปิด “แยก%s” ของบัญชีนี้ก่อน หรือล้างรายการ%sที่เลือกไว้", d.label, d.label, d.label))
}

func lineField(index int, d dimensionKind) string {
	return fmt.Sprintf("lines[%d].%s", index, d.jsonKey)
}

func lineDimensionNotFound(kind string, index int, code string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("บรรทัด %d: ไม่พบ%s “%s” ที่เปิดใช้งานในทะเบียน%s — เลือกรหัสจากรายการ", index+1, d.label, code, d.label)
	err := fieldError(dimensionErrorCode("journal_line_", d, "_not_found"), lineField(index, d), message)
	err.Args = []string{strconv.Itoa(index + 1), code}
	return err
}

func lineDimensionNotAllowed(kind string, index int, account, code string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("บรรทัด %d: บัญชี %s กำหนด%sที่ใช้ได้ไว้ ไม่รวม%s “%s” — เลือก%sที่บัญชีนี้อนุญาต หรือแก้รายการที่ผังบัญชี", index+1, account, d.label, d.label, code, d.label)
	err := fieldError(dimensionErrorCode("journal_line_", d, "_not_allowed"), lineField(index, d), message)
	err.Args = []string{strconv.Itoa(index + 1), account, code}
	return err
}

// lineDimensionNotSeparate - Champ OnCheckSeparate "Chart_Of_Account <code> NOT_SEPARATE" (glfrmjournal.cpp:639-686).
func lineDimensionNotSeparate(kind string, index int, account string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("บรรทัด %d: ผังบัญชี %s ไม่ได้กำหนดให้แยก%s", index+1, account, d.label)
	err := fieldError(dimensionErrorCode("journal_line_", d, "_not_separate"), lineField(index, d), message)
	err.Args = []string{strconv.Itoa(index + 1), account}
	return err
}

// lineDimensionDisabled: the company switch of the register is off, so a NEW code cannot be put on a line. Champ hides the
// column then (glfrmjournal.cpp:227-256) and so cannot take the value; the API can, hence this small extra (design §3.4 item 3).
func lineDimensionDisabled(kind string, index int, code string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("บรรทัด %d: บริษัทยังไม่เปิดใช้ “แยก%s” จึงระบุ%s “%s” ไม่ได้ — เปิดที่หน้ากำหนดค่าเริ่มต้นบริษัทก่อน หรือล้างช่อง%sของบรรทัดนี้", index+1, d.label, d.label, code, d.label)
	err := fieldError(dimensionErrorCode("journal_line_", d, "_disabled"), lineField(index, d), message)
	err.Args = []string{strconv.Itoa(index + 1), code}
	return err
}

func dimensionInUseDelete(kind, code string) error {
	label := dimensionKinds[kind].label
	return conflictError("dimension_in_use_delete", "code", fmt.Sprintf("%s “%s” มีรายการบัญชี ใบสำคัญร่าง หรือข้อมูลบัญชีอื่นใช้อยู่แล้ว ลบไม่ได้ ถ้าไม่ต้องการใช้ต่อ ให้ปิดใช้งานแทน", label, code))
}

func dimensionInUseCode(kind, code string) error {
	label := dimensionKinds[kind].label
	return conflictError("dimension_in_use_code", "code", fmt.Sprintf("%s “%s” มีรายการบัญชี ใบสำคัญร่าง หรือข้อมูลบัญชีอื่นใช้อยู่แล้ว เปลี่ยนรหัสไม่ได้ ถ้าต้องการรหัสใหม่ ให้สร้าง%sใหม่แล้วปิดใช้งานของเดิม", label, code, label))
}

// ---- master validation ----

// validateDimensionMaster checks a register record: { code ≤ codeMax, name required ≤ nameMax, usergroup, isactive }.
// Only those fields (plus identity) survive, so a client cannot park unrelated master fields in the record.
// Lengths follow the design §3.1: job code 15, name 255 for job / part / side. Champ's BCPart / BCSide register Code is VARCHAR(15)
// (SQLSERVER_Script.sql:9148-9176) but the journal line column PartCode / SideCode is VARCHAR(10) (:9440-9441), so a longer code
// could never sit on a line; the design keeps 10 for both.
// usergroup is Champ's UserGroup column shown beside the code (myfrmbasemaster.cpp:509-536); it never filters the journal picker.
func validateDimensionMaster(m *Master) error {
	d := dimensionKinds[m.Kind]
	if err := checkCode(m.Code, "code", "รหัส"+d.label, d.codeMax); err != nil {
		return err
	}
	name := strings.TrimSpace(m.Name)
	switch n := utf8.RuneCountInString(name); {
	case n == 0:
		return dimensionNameRequired(m.Kind)
	case n > d.nameMax:
		return dimensionNameTooLong(m.Kind, n)
	}
	userGroup := strings.TrimSpace(m.UserGroup)
	if n := utf8.RuneCountInString(userGroup); n > dimensionUserGroupMaxRunes {
		return fieldError("dimension_usergroup_too_long", "usergroup", fmt.Sprintf("กลุ่มพนักงาน/ใช้กับรายวันยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว) กรุณาย่อให้สั้นลง", dimensionUserGroupMaxRunes, n))
	}
	*m = Master{Identity: m.Identity, Kind: m.Kind, Code: m.Code, Name: name, UserGroup: userGroup, IsActive: m.IsActive}
	return nil
}

// dimensionUserGroupMaxRunes: Champ BCDepartment / BCAllocate / BCPart / BCSide / BCJob .UserGroup VARCHAR(255) (SQLSERVER_Script.sql:370-382, 460-472, 9148-9190).
const dimensionUserGroupMaxRunes = 255

// dimensionInUse reports whether anything in the company still names the code: ledger lines, any live journal
// (drafts included — their lines are not in gl_lines yet), saved journal templates, allocation rules, cash
// forecasts, budgets and accounts whose allowed-code list names it. Deleted records do not count, so deleting a draft frees its codes.
func dimensionInUse(ctx context.Context, tx *sql.Tx, company, kind, code string) (bool, error) {
	d := dimensionKinds[kind]
	var used bool
	// d.column / d.jsonKey / d.listKey come from the dimensionRegistry table above, never from the request.
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM gl_lines WHERE company=$1 AND %[1]s=$2)
		OR EXISTS(SELECT 1 FROM gl_budgets WHERE company=$1 AND %[1]s=$2)
		OR EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND (
			(kind='journals' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'lines','[]'::jsonb)) l WHERE l->>'%[2]s'=$2))
			OR (kind='journal-templates' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'template'->'lines','[]'::jsonb)) l WHERE l->>'%[2]s'=$2))
			OR (kind='allocations' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'allocaterules','[]'::jsonb)) r WHERE r->>'%[2]s'=$2))
			OR (kind='forecast' AND payload->>'%[2]s'=$2)
			OR (kind='accounts' AND COALESCE(payload->'%[3]s','[]'::jsonb) @> to_jsonb($2::text))))`, d.column, d.jsonKey, d.listKey)
	err := tx.QueryRowContext(ctx, query, company, code).Scan(&used)
	return used, err
}

// guardDimensionUpdate allows name and isactive edits always; the code is frozen once anything uses it.
func guardDimensionUpdate(ctx context.Context, tx *sql.Tx, company, kind string, old, next Master) error {
	if next.Code == old.Code {
		return nil
	}
	used, err := dimensionInUse(ctx, tx, company, kind, old.Code)
	if err != nil {
		return err
	}
	if used {
		return dimensionInUseCode(kind, old.Code)
	}
	return nil
}

func guardDimensionDelete(ctx context.Context, tx *sql.Tx, company, kind string, old Master) error {
	used, err := dimensionInUse(ctx, tx, company, kind, old.Code)
	if err != nil {
		return err
	}
	if used {
		return dimensionInUseDelete(kind, old.Code)
	}
	return nil
}

// normalizeCodeList trims and NFC-normalises codes, drops blanks and repeats, and sorts them so the stored list is stable.
func normalizeCodeList(codes []string) []string {
	seen := make(map[string]bool, len(codes))
	var out []string
	for _, code := range codes {
		if code = NormalizeCode(code); code != "" && !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	sort.Strings(out)
	return out
}

// dimensionRegisters holds, per register kind, the codes a validation names: code -> isactive. A code missing from
// the map is not in the register (or was deleted).
type dimensionRegisters map[string]map[string]bool

// loadDimensionCodes reads the named codes of one register in a single query.
func loadDimensionCodes(ctx context.Context, tx *sql.Tx, company, kind string, codes []string) (map[string]bool, error) {
	found := map[string]bool{}
	if len(codes) == 0 {
		return found, nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT code, COALESCE((payload->>'isactive')::boolean,false) FROM gl_records WHERE company=$1 AND kind=$2 AND code=ANY($3) AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, kind, pq.Array(codes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		var active bool
		if err = rows.Scan(&code, &active); err != nil {
			return nil, err
		}
		found[code] = found[code] || active
	}
	return found, rows.Err()
}

// namedLineCodes collects, per register, the distinct codes the lines name (blank = none).
func namedLineCodes(lines []Line) map[string]map[string]bool {
	named := make(map[string]map[string]bool, len(dimensionRegistry))
	for _, d := range dimensionRegistry {
		named[d.kind] = map[string]bool{}
	}
	for _, line := range lines {
		for _, d := range dimensionRegistry {
			if code := strings.TrimSpace(line.dimensionCode(d.kind)); code != "" {
				named[d.kind][code] = true
			}
		}
	}
	return named
}

// lineNamesDimension reports whether any line carries a code in any register; the validation reads nothing else when none does.
func lineNamesDimension(lines []Line) bool {
	for _, set := range namedLineCodes(lines) {
		if len(set) > 0 {
			return true
		}
	}
	return false
}

// loadLineDimensions loads the codes the lines name — one query per register, none when no line names one in
// that register, so a voucher check never reads a whole register.
func loadLineDimensions(ctx context.Context, tx *sql.Tx, company string, lines []Line) (dimensionRegisters, error) {
	regs := dimensionRegisters{}
	for kind, set := range namedLineCodes(lines) {
		codes := make([]string, 0, len(set))
		for code := range set {
			codes = append(codes, code)
		}
		found, err := loadDimensionCodes(ctx, tx, company, kind, codes)
		if err != nil {
			return nil, err
		}
		regs[kind] = found
	}
	return regs, nil
}

// checkLineCode: the code must be an active register code and the account must accept it. An account that ticks
// "เลือกทั้งหมด" accepts every code; otherwise only the codes in its list — an empty list accepts none
// (Champ OnCheckComboData glfrmjournal.cpp:2815-2923: flag on + "_ALL" row = all, flag on + rows = those, flag on + no row = none).
func (regs dimensionRegisters) checkLineCode(kind string, index int, account Account, code string) error {
	if !regs[kind][code] {
		return lineDimensionNotFound(kind, index, code)
	}
	dim := account.accountDimension(kind)
	if !*dim.all && !contains(*dim.codes, code) {
		return lineDimensionNotAllowed(kind, index, account.AccountCode, code)
	}
	return nil
}

// checkJournalDimensions runs the dimension rules of a voucher (or template) the user is saving: it reads the company switches
// and the register codes the lines name, then applies checkLineSeparation. oldLines are the stored lines of the same record
// (nil for a new one) — a code they already hold is kept even when its switch has been turned off since.
func checkJournalDimensions(ctx context.Context, tx *sql.Tx, company string, lines []Line, accounts map[string]Account, oldLines []Line) error {
	settings, err := loadDimensionSettings(ctx, tx, company)
	if err != nil {
		return err
	}
	dims, err := loadLineDimensions(ctx, tx, company, lines)
	if err != nil {
		return err
	}
	return checkLineSeparation(lines, accounts, dims, settings, oldLines)
}

// checkAccountDimensionCodes validates the allowed-code lists of an account being saved: each code must be in the
// register, and active unless the account already listed it (a code deactivated later must not block editing the account).
func checkAccountDimensionCodes(ctx context.Context, tx *sql.Tx, company string, next Account, old *Account) error {
	for _, d := range dimensionRegistry {
		nextList := *next.accountDimension(d.kind).codes
		if len(nextList) == 0 {
			continue
		}
		var prev []string
		if old != nil {
			prev = *old.accountDimension(d.kind).codes
		}
		found, err := loadDimensionCodes(ctx, tx, company, d.kind, nextList)
		if err != nil {
			return err
		}
		for _, code := range nextList {
			active, exists := found[code]
			if !exists || (!active && !contains(prev, code)) {
				return accountDimensionCodeNotFound(d.kind, code)
			}
		}
	}
	return nil
}

// ---- budgets (design §3.11) ----

// dimensionCode is the code the budget is limited to for the register ("" = any).
func (b Budget) dimensionCode(kind string) string {
	switch kind {
	case DepartmentKind:
		return b.DepartmentCode
	case ProjectKind:
		return b.ProjectCode
	case AllocateKind:
		return b.AllocateCode
	case JobKind:
		return b.JobCode
	case PartKind:
		return b.PartCode
	case SideKind:
		return b.SideCode
	}
	return ""
}

func budgetDimensionNotFound(kind, code string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("ไม่พบ%s “%s” ที่เปิดใช้งานในทะเบียน%s — เลือกรหัสจากรายการ", d.label, code, d.label)
	err := fieldError(dimensionErrorCode("budget_", d, "_not_found"), d.jsonKey, message)
	err.Args = []string{code}
	return err
}

// budgetDimensionNotSeparate - Champ GLFrmBudget.cpp:471 "Chart_Of_Account <code> NOT_SEPARATE".
func budgetDimensionNotSeparate(kind, account string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("ผังบัญชี %s ไม่ได้กำหนดให้แยก%s จึงใช้ในงบประมาณที่ระบุ%sนี้ไม่ได้", account, d.label, d.label)
	err := fieldError(dimensionErrorCode("budget_", d, "_not_separate"), d.jsonKey, message)
	err.Args = []string{account}
	return err
}

func budgetDimensionNotAllowed(kind, account, code string) error {
	d := dimensionKinds[kind]
	message := fmt.Sprintf("บัญชี %s กำหนด%sที่ใช้ได้ไว้ ไม่รวม%s “%s” — เลือก%sที่บัญชีนี้อนุญาต หรือแก้รายการที่ผังบัญชี", account, d.label, d.label, code, d.label)
	err := fieldError(dimensionErrorCode("budget_", d, "_not_allowed"), d.jsonKey, message)
	err.Args = []string{account, code}
	return err
}

// checkBudgetDimensions: a budget limited to a register code must name an active code of the register, and every account
// in it must be flagged "แยก…" and accept the code (Champ ConfrimSaveData / OnCheckSeparate, GLFrmBudget.cpp:421-530; the
// allowed-code list is the journal rule of §3.5 item 5). Champ reads a dimension only when its company switch is on
// (GLFrmBudget.cpp:138-151, 441-447), so a register whose switch is off is not checked and the stored code is kept.
func checkBudgetDimensions(ctx context.Context, tx *sql.Tx, company string, b Budget, accounts map[string]Account) error {
	named := false
	for _, d := range dimensionRegistry {
		named = named || b.dimensionCode(d.kind) != ""
	}
	if !named {
		return nil
	}
	settings, err := loadDimensionSettings(ctx, tx, company)
	if err != nil {
		return err
	}
	for _, d := range dimensionRegistry {
		code := b.dimensionCode(d.kind)
		if code == "" || !settings.enabled(d.kind) {
			continue
		}
		found, err := loadDimensionCodes(ctx, tx, company, d.kind, []string{code})
		if err != nil {
			return err
		}
		if !found[code] {
			return budgetDimensionNotFound(d.kind, code)
		}
		for _, line := range b.Lines {
			account := accounts[line.AccountCode]
			dim := account.accountDimension(d.kind)
			if !*dim.separate {
				return budgetDimensionNotSeparate(d.kind, line.AccountCode)
			}
			if !*dim.all && !contains(*dim.codes, code) {
				return budgetDimensionNotAllowed(d.kind, line.AccountCode, code)
			}
		}
	}
	return nil
}

// ---- year-end balances (design §3.9): a closing / opening line keeps every dimension of the balance it moves ----

// dimensionCodes holds one code per register in registry order; as a map key it groups balances that agree on every dimension.
type dimensionCodes [6]string

func init() {
	if len(dimensionRegistry) != len(dimensionCodes{}) {
		panic("dimensionCodes needs one slot per analysis register")
	}
}

// applyTo copies the codes onto a generated line.
func (c dimensionCodes) applyTo(line *Line) {
	for i, d := range dimensionRegistry {
		*line.lineCodePtr(d.kind) = c[i]
	}
}

// less orders groups by register order so generated vouchers come out in a stable order.
func (c dimensionCodes) less(other dimensionCodes) bool {
	for i := range c {
		if c[i] != other[i] {
			return c[i] < other[i]
		}
	}
	return false
}

func (b *ProcessBalance) codePtr(kind string) *string {
	switch kind {
	case DepartmentKind:
		return &b.DepartmentCode
	case ProjectKind:
		return &b.ProjectCode
	case AllocateKind:
		return &b.AllocateCode
	case JobKind:
		return &b.JobCode
	case PartKind:
		return &b.PartCode
	case SideKind:
		return &b.SideCode
	}
	return nil
}

// dimensionCodes: the balance's codes in registry order.
func (b ProcessBalance) dimensionCodes() dimensionCodes {
	var codes dimensionCodes
	for i, d := range dimensionRegistry {
		codes[i] = *b.codePtr(d.kind)
	}
	return codes
}

// scanTargets lists the Scan destinations of processBalanceQuery.
func (b *ProcessBalance) scanTargets() []any {
	targets := []any{&b.AccountCode, &b.AccountName, &b.AccountType, &b.BranchCode}
	for _, d := range dimensionRegistry {
		targets = append(targets, b.codePtr(d.kind))
	}
	return append(targets, &b.Balance)
}

// processBalanceQuery: the signed balance of every account per branch and combination of dimension codes, up to a date ($3).
var processBalanceQuery = `SELECT account_code,MAX(account_name),MAX(account_type),branch_code,` + dimensionColumnList("") + `,SUM(debit-credit)::text
  FROM gl_lines WHERE company=$1 AND fiscal_year=$2 AND entry_date<=$3::date
  GROUP BY account_code,branch_code,` + dimensionColumnList("") + ` HAVING SUM(debit-credit)<>0
  ORDER BY branch_code,` + dimensionColumnList("") + `,account_code LIMIT $4`

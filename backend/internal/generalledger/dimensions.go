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

// Departments and projects are the dimension registers a journal line can name. Champ keeps them in
// BCDepartment / BCProject (Code VARCHAR(10) primary key, Name VARCHAR(50) NOT NULL — Script\SQLSERVER_Script.sql:460-468,
// 590-598) and edits them in the generic master form (myfrmbasemaster.cpp:223,242). A journal line may only name an active
// code of its register, and an account can narrow the codes to a list (Account.DepartCodes / ProjectCodes).
const (
	DepartmentKind = "departments"
	ProjectKind    = "projects"

	dimensionCodeMaxRunes = 10 // BCDepartment.Code / BCProject.Code VARCHAR(10)
	dimensionNameMaxRunes = 50 // BCDepartment.Name / BCProject.Name VARCHAR(50)
)

// dimensionKind describes one register: the Thai label used in messages, the gl_lines / gl_budgets column, the
// JSON key under which records elsewhere in gl_records hold a code of this register and the account JSON key of its
// allowed-code list (never client input).
type dimensionKind struct {
	label, column, jsonKey, accountListKey string
}

var dimensionKinds = map[string]dimensionKind{
	DepartmentKind: {label: "แผนก", column: "department_code", jsonKey: "departmentcode", accountListKey: "departcodes"},
	ProjectKind:    {label: "โครงการ", column: "project_code", jsonKey: "projectcode", accountListKey: "projectcodes"},
}

func isDimensionKind(kind string) bool {
	_, ok := dimensionKinds[kind]
	return ok
}

// ---- user-facing errors (codes get gl_err_<code> rows in languages.tsv) ----

func dimensionNameRequired(kind string) error {
	return fieldError("dimension_name_required", "name", "กรุณาระบุชื่อ"+dimensionKinds[kind].label)
}

func dimensionNameTooLong(kind string, n int) error {
	return fieldError("dimension_name_too_long", "name", fmt.Sprintf("ชื่อ%sยาวได้ไม่เกิน %d ตัวอักษร (ตอนนี้ %d ตัว) กรุณาย่อชื่อให้สั้นลง", dimensionKinds[kind].label, dimensionNameMaxRunes, n))
}

func accountDimensionCodeNotFound(kind, code string) error {
	label := dimensionKinds[kind].label
	message := fmt.Sprintf("ไม่พบ%s “%s” ที่เปิดใช้งานในทะเบียน%s — เลือกรหัสที่มีในทะเบียน หรือเปิดใช้งานรหัสนั้นก่อน", label, code, label)
	var err *UserError
	if kind == DepartmentKind {
		err = fieldError("account_depart_code_not_found", "departcodes", message)
	} else {
		err = fieldError("account_project_code_not_found", "projectcodes", message)
	}
	err.Args = []string{code}
	return err
}

func lineDimensionNotFound(kind string, index int, code string) error {
	label := dimensionKinds[kind].label
	message := fmt.Sprintf("บรรทัด %d: ไม่พบ%s “%s” ที่เปิดใช้งานในทะเบียน%s — เลือกรหัสจากรายการ", index+1, label, code, label)
	field := fmt.Sprintf("lines[%d].%s", index, dimensionKinds[kind].jsonKey)
	var err *UserError
	if kind == DepartmentKind {
		err = fieldError("journal_line_depart_not_found", field, message)
	} else {
		err = fieldError("journal_line_project_not_found", field, message)
	}
	err.Args = []string{strconv.Itoa(index + 1), code}
	return err
}

func lineDimensionNotAllowed(kind string, index int, account, code string) error {
	label := dimensionKinds[kind].label
	message := fmt.Sprintf("บรรทัด %d: บัญชี %s กำหนด%sที่ใช้ได้ไว้ ไม่รวม%s “%s” — เลือก%sที่บัญชีนี้อนุญาต หรือแก้รายการที่ผังบัญชี", index+1, account, label, label, code, label)
	field := fmt.Sprintf("lines[%d].%s", index, dimensionKinds[kind].jsonKey)
	var err *UserError
	if kind == DepartmentKind {
		err = fieldError("journal_line_depart_not_allowed", field, message)
	} else {
		err = fieldError("journal_line_project_not_allowed", field, message)
	}
	err.Args = []string{strconv.Itoa(index + 1), account, code}
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

// validateDimensionMaster checks a departments / projects record: { code ≤10, name required ≤50, isactive }.
// Only those fields (plus identity) survive, so a client cannot park unrelated master fields in the record.
func validateDimensionMaster(m *Master) error {
	label := dimensionKinds[m.Kind].label
	if err := checkCode(m.Code, "code", "รหัส"+label, dimensionCodeMaxRunes); err != nil {
		return err
	}
	name := strings.TrimSpace(m.Name)
	switch n := utf8.RuneCountInString(name); {
	case n == 0:
		return dimensionNameRequired(m.Kind)
	case n > dimensionNameMaxRunes:
		return dimensionNameTooLong(m.Kind, n)
	}
	*m = Master{Identity: m.Identity, Kind: m.Kind, Code: m.Code, Name: name, IsActive: m.IsActive}
	return nil
}

// dimensionInUse reports whether anything in the company still names the code: ledger lines, any live journal
// (drafts included — their lines are not in gl_lines yet), saved journal templates, allocation rules, cash
// forecasts, budgets and accounts whose allowed-code list names it. Deleted records do not count, so deleting a draft frees its codes.
func dimensionInUse(ctx context.Context, tx *sql.Tx, company, kind, code string) (bool, error) {
	d := dimensionKinds[kind]
	var used bool
	// d.column / d.jsonKey come from the dimensionKinds table above, never from the request.
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM gl_lines WHERE company=$1 AND %[1]s=$2)
		OR EXISTS(SELECT 1 FROM gl_budgets WHERE company=$1 AND %[1]s=$2)
		OR EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND (
			(kind='journals' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'lines','[]'::jsonb)) l WHERE l->>'%[2]s'=$2))
			OR (kind='journal-templates' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'template'->'lines','[]'::jsonb)) l WHERE l->>'%[2]s'=$2))
			OR (kind='allocations' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'allocaterules','[]'::jsonb)) r WHERE r->>'%[2]s'=$2))
			OR (kind='forecast' AND payload->>'%[2]s'=$2)
			OR (kind='accounts' AND COALESCE(payload->'%[3]s','[]'::jsonb) @> to_jsonb($2::text))))`, d.column, d.jsonKey, d.accountListKey)
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

// loadLineDimensions loads the department and project codes the lines name — one query per register, none when no
// line names one, so a voucher check never reads the whole register.
func loadLineDimensions(ctx context.Context, tx *sql.Tx, company string, lines []Line) (dimensionRegisters, error) {
	named := map[string]map[string]bool{DepartmentKind: {}, ProjectKind: {}}
	for _, line := range lines {
		if code := strings.TrimSpace(line.DepartmentCode); code != "" {
			named[DepartmentKind][code] = true
		}
		if code := strings.TrimSpace(line.ProjectCode); code != "" {
			named[ProjectKind][code] = true
		}
	}
	regs := dimensionRegisters{}
	for kind, set := range named {
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

// checkLineCode: the code must be an active register code and, when the account lists allowed codes, one of them.
func (regs dimensionRegisters) checkLineCode(kind string, index int, account, code string, allowedCodes []string) error {
	if !regs[kind][code] {
		return lineDimensionNotFound(kind, index, code)
	}
	if len(allowedCodes) > 0 && !contains(allowedCodes, code) {
		return lineDimensionNotAllowed(kind, index, account, code)
	}
	return nil
}

// checkAccountDimensionCodes validates the allowed-code lists of an account being saved: each code must be in the
// register, and active unless the account already listed it (a code deactivated later must not block editing the account).
func checkAccountDimensionCodes(ctx context.Context, tx *sql.Tx, company string, next Account, old *Account) error {
	lists := []struct {
		kind       string
		next, prev []string
	}{{kind: DepartmentKind, next: next.DepartCodes}, {kind: ProjectKind, next: next.ProjectCodes}}
	if old != nil {
		lists[0].prev, lists[1].prev = old.DepartCodes, old.ProjectCodes
	}
	for _, list := range lists {
		if len(list.next) == 0 {
			continue
		}
		found, err := loadDimensionCodes(ctx, tx, company, list.kind, list.next)
		if err != nil {
			return err
		}
		for _, code := range list.next {
			active, exists := found[code]
			if !exists || (!active && !contains(list.prev, code)) {
				return accountDimensionCodeNotFound(list.kind, code)
			}
		}
	}
	return nil
}

package generalledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Departments and projects are the dimension registers a journal line can name. Champ keeps them in
// BCDepartment / BCProject (Code VARCHAR(10) primary key, Name VARCHAR(50) NOT NULL — Script\SQLSERVER_Script.sql:460-468,
// 590-598) and edits them in the generic master form (myfrmbasemaster.cpp:223,242). Phase 1 only keeps the registers:
// journal lines still carry free codes, nothing validates a line against these masters yet.
const (
	DepartmentKind = "departments"
	ProjectKind    = "projects"

	dimensionCodeMaxRunes = 10 // BCDepartment.Code / BCProject.Code VARCHAR(10)
	dimensionNameMaxRunes = 50 // BCDepartment.Name / BCProject.Name VARCHAR(50)
)

// dimensionKind describes one register: the Thai label used in messages, the gl_lines / gl_budgets column and the
// JSON key under which records elsewhere in gl_records hold a code of this register (never client input).
type dimensionKind struct {
	label, column, jsonKey string
}

var dimensionKinds = map[string]dimensionKind{
	DepartmentKind: {label: "แผนก", column: "department_code", jsonKey: "departmentcode"},
	ProjectKind:    {label: "โครงการ", column: "project_code", jsonKey: "projectcode"},
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
// forecasts and budgets. Deleted records do not count, so deleting a draft frees its codes.
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
			OR (kind='forecast' AND payload->>'%[2]s'=$2)))`, d.column, d.jsonKey)
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

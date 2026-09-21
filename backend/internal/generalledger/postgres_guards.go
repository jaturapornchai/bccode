package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *PostgresStore) pgYear(ctx context.Context, tx *sql.Tx, scope Scope, code string) (FiscalYear, error) {
	var y FiscalYear
	var data []byte
	err := tx.QueryRowContext(ctx, `SELECT payload || jsonb_build_object('id',id,'version',version) FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, scope.Company, code).Scan(&data)
	if err == sql.ErrNoRows {
		return y, fmt.Errorf("กรุณาตั้งค่าปีบัญชีก่อนบันทึกรายการ")
	}
	if err != nil {
		return y, err
	}
	err = json.Unmarshal(data, &y)
	return y, err
}
func (s *PostgresStore) openPGYear(ctx context.Context, tx *sql.Tx, scope Scope, code string) (FiscalYear, error) {
	y, err := s.pgYear(ctx, tx, scope, code)
	if err != nil {
		return y, err
	}
	if y.Closed || !y.IsActive {
		return y, fmt.Errorf("ปีบัญชีปิดหรือไม่ได้เปิดใช้งาน")
	}
	return y, nil
}
func (s *PostgresStore) pgYearAt(ctx context.Context, tx *sql.Tx, scope Scope, date string) (FiscalYear, error) {
	var code string
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, scope.Company, date).Scan(&code)
	if err != nil {
		return FiscalYear{}, fmt.Errorf("ไม่พบปีบัญชีสำหรับวันที่กลับรายการ: %w", err)
	}
	return s.openPGYear(ctx, tx, scope, code)
}
func (s *PostgresStore) checkOpenDate(ctx context.Context, tx *sql.Tx, scope Scope, date string) error {
	var locked bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='periods' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND COALESCE((payload->>'locked')::boolean,false) AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, date).Scan(&locked)
	if err != nil {
		return err
	}
	if locked {
		return fmt.Errorf("งวดบัญชีนี้ถูกล็อก กรุณาเลือกวันที่ในงวดเปิด")
	}
	return nil
}
func pgChange(kind, id, code string, value any) ([]Change, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return []Change{{Kind: kind, ID: id, Code: code, Payload: string(data)}}, nil
}

func (s *PostgresStore) mutatePeriod(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	var old Master
	if cmd.Action != "create" {
		if err := s.loadRecord(ctx, tx, scope.Company, "periods", cmd.ID, &old); err != nil {
			return nil, err
		}
		if err := checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
		if _, err := s.openPGYear(ctx, tx, scope, old.FiscalYear); err != nil {
			return nil, err
		}
	}
	if cmd.Action == "lock" || cmd.Action == "unlock" {
		if strings.TrimSpace(cmd.Reason) == "" {
			return nil, fmt.Errorf("กรุณาระบุเหตุผลล็อกหรือปลดล็อกงวด")
		}
		if cmd.Action == "lock" {
			var pending bool
			err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='journals' AND payload->>'fiscalyear'=$2 AND payload->>'date'>=$3 AND payload->>'date'<=$4 AND payload->>'status'='draft' AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, old.FiscalYear, old.StartDate, old.EndDate).Scan(&pending)
			if err != nil {
				return nil, err
			}
			if pending {
				return nil, fmt.Errorf("ยังมีรายการร่างในงวด กรุณาผ่านรายการก่อนล็อก")
			}
		}
		old.Locked = cmd.Action == "lock"
		old.Identity = updateIdentity(scope, old.Identity, now)
		return pgChange("periods", old.ID, old.Code, old)
	}
	if cmd.Action == "delete" {
		var used bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='journals' AND payload->>'fiscalyear'=$2 AND payload->>'date'>=$3 AND payload->>'date'<=$4 AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, old.FiscalYear, old.StartDate, old.EndDate).Scan(&used)
		if err != nil {
			return nil, err
		}
		if used {
			return nil, fmt.Errorf("งวดนี้มีรายการบัญชีอ้างอิง ลบไม่ได้")
		}
		if old.Locked {
			return nil, fmt.Errorf("งวดถูกล็อก ห้ามลบ")
		}
		old.IsDeleted = true
		old.Identity = updateIdentity(scope, old.Identity, now)
		return pgChange("periods", old.ID, old.Code, old)
	}
	if cmd.Action != "create" && cmd.Action != "update" {
		return nil, fmt.Errorf("คำสั่งงวดบัญชีไม่ถูกต้อง")
	}
	if cmd.Master == nil {
		return nil, fmt.Errorf("ไม่พบข้อมูลงวดบัญชี")
	}
	next := *cmd.Master
	next.Kind = "periods"
	next.Locked = old.Locked
	if !validCode(next.Code) || !validDate(next.StartDate) || !validDate(next.EndDate) || next.StartDate > next.EndDate {
		return nil, fmt.Errorf("รหัสหรือวันที่งวดบัญชีไม่ถูกต้อง")
	}
	year, err := s.openPGYear(ctx, tx, scope, next.FiscalYear)
	if err != nil {
		return nil, err
	}
	if next.StartDate < year.StartDate || next.EndDate > year.EndDate {
		return nil, fmt.Errorf("งวดอยู่นอกปีบัญชี")
	}
	if cmd.Action == "update" && (next.Code != old.Code || next.FiscalYear != old.FiscalYear || next.StartDate != old.StartDate || next.EndDate != old.EndDate || !next.IsActive) {
		return nil, fmt.Errorf("เปลี่ยนขอบเขตงวดเดิมไม่ได้")
	}
	var overlaps bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='periods' AND id<>$2 AND payload->>'startdate'<=$3 AND payload->>'enddate'>=$4 AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, cmd.ID, next.EndDate, next.StartDate).Scan(&overlaps)
	if err != nil {
		return nil, err
	}
	if overlaps {
		return nil, fmt.Errorf("ช่วงงวดบัญชีทับซ้อน")
	}
	if err = s.checkDuplicateCode(ctx, tx, scope.Company, "periods", next.Code, cmd.ID); err != nil {
		return nil, err
	}
	if cmd.Action == "create" {
		next.Identity = newIdentity(scope, now)
	} else {
		next.Identity = updateIdentity(scope, old.Identity, now)
	}
	return pgChange("periods", next.ID, next.Code, next)
}

func (s *PostgresStore) validatePGAccount(ctx context.Context, tx *sql.Tx, scope Scope, next *Account, old *Account) error {
	if err := next.Validate(); err != nil {
		return err
	}
	if next.AccountGroup != "" {
		var exists bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='account-groups' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND COALESCE((payload->>'isactive')::boolean,false))`, scope.Company, next.AccountGroup).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return userError(CodeGroupNotFound, "ไม่พบกลุ่มผังบัญชีที่เปิดใช้งาน")
		}
	}
	if old != nil && (next.AccountType != old.AccountType || next.NormalBalance != old.NormalBalance || next.IsCash != old.IsCash || next.ParentAccountCode != old.ParentAccountCode || next.AllowPosting != old.AllowPosting) {
		var used bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_lines WHERE company=$1 AND account_code=$2)`, scope.Company, old.AccountCode).Scan(&used)
		if err != nil {
			return err
		}
		if used {
			return userError(CodePostedLocked, "บัญชีนี้ผ่านรายการแล้ว เปลี่ยนหมวด ด้านบัญชี หรือโครงสร้างไม่ได้")
		}
	}
	seen := map[string]bool{next.AccountCode: true}
	parent := next.ParentAccountCode
	parentLevel := 0
	depth := 0
	for parent != "" {
		if seen[parent] || depth >= 12 {
			return userError(CodeTreeInvalid, "โครงสร้างบัญชีวนซ้ำหรือมีระดับมากเกินไป")
		}
		seen[parent] = true
		depth++
		var data []byte
		err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='accounts' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, scope.Company, parent).Scan(&data)
		if err == sql.ErrNoRows {
			return userError(CodeParentNotFound, "ไม่พบบัญชีแม่ในบริษัทนี้")
		}
		if err != nil {
			return err
		}
		var a Account
		if err = json.Unmarshal(data, &a); err != nil {
			return err
		}
		if a.AllowPosting || !a.IsActive {
			return userError(CodeParentInvalid, "บัญชีแม่ต้องเป็นบัญชีคุมที่เปิดใช้งาน")
		}
		if depth == 1 {
			parentLevel = a.Level
			if parentLevel <= 0 {
				parentLevel = 1
			}
		}
		parent = a.ParentAccountCode
	}
	if next.Level <= 0 {
		next.Level = parentLevel + 1
	}
	if next.Level < 1 || next.Level > 12 {
		return userError(CodeLevelRange, "ระดับบัญชีต้องอยู่ระหว่าง 1 ถึง 12")
	}
	if parentLevel > 0 && next.Level <= parentLevel {
		return userError(CodeLevelShallow, "ระดับบัญชีต้องมากกว่าบัญชีแม่")
	}
	if old != nil {
		var invalid bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='accounts' AND payload->>'parentaccountcode'=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND ($3 OR COALESCE((payload->>'level')::integer,1)<=$4))`, scope.Company, old.AccountCode, next.AllowPosting, next.Level).Scan(&invalid)
		if err != nil {
			return err
		}
		if invalid {
			return userError(CodeHasChildren, "บัญชีนี้มีบัญชีลูก กรุณาตรวจระดับและการอนุญาตลงรายการ")
		}
	}
	return nil
}
func (s *PostgresStore) deletePGAccountGuard(ctx context.Context, tx *sql.Tx, scope Scope, code string) error {
	var used bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND ((kind='journals' AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'lines','[]'::jsonb)) l WHERE l->>'accountcode'=$2)) OR payload->>'parentaccountcode'=$2 OR payload->>'accountcode'=$2 AND kind<>'accounts' OR payload->>'profitlossaccount'=$2 OR payload->>'retainedearningsaccount'=$2 OR payload->>'itemaccount'=$2 OR payload->>'costaccount'=$2 OR payload->>'revenueaccount'=$2 OR EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'rules','[]'::jsonb)) r WHERE r->>'accountcode'=$2) OR EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'rows','[]'::jsonb)) r WHERE COALESCE(r->'accountcodes','[]'::jsonb) ? $2)))`, scope.Company, code).Scan(&used)
	if err != nil {
		return err
	}
	if used {
		return userError(CodeReferenced, "บัญชีนี้มีเอกสารหรือข้อมูลหลักอ้างอิง ลบไม่ได้")
	}
	return nil
}
func (s *PostgresStore) mutatePGFiscalYear(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	var old FiscalYear
	if cmd.Action != "create" {
		if err := s.loadRecord(ctx, tx, scope.Company, "fiscal-years", cmd.ID, &old); err != nil {
			return nil, err
		}
		if err := checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
		if old.Closed {
			return nil, fmt.Errorf("ปีบัญชีปิดแล้ว เปลี่ยนข้อมูลไม่ได้")
		}
	}
	if cmd.Action != "create" && cmd.Action != "update" && cmd.Action != "delete" {
		return nil, fmt.Errorf("คำสั่งปีบัญชีไม่ถูกต้อง")
	}
	var used bool
	if old.ID != "" {
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND payload->>'fiscalyear'=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, old.Code).Scan(&used)
		if err != nil {
			return nil, err
		}
	}
	if cmd.Action == "delete" {
		if used {
			return nil, fmt.Errorf("ปีบัญชีมีข้อมูลอ้างอิง ลบไม่ได้")
		}
		old.IsDeleted = true
		old.Identity = updateIdentity(scope, old.Identity, now)
		return pgChange("fiscal-years", old.ID, old.Code, old)
	}
	if cmd.FiscalYear == nil {
		return nil, fmt.Errorf("ไม่พบข้อมูลปีบัญชี")
	}
	next := *cmd.FiscalYear
	next.Closed = false
	if err := next.Validate(); err != nil {
		return nil, err
	}
	if old.ID != "" && (next.Code != old.Code || used && (next.StartDate != old.StartDate || next.EndDate != old.EndDate || next.Scale != old.Scale || !next.IsActive)) {
		return nil, fmt.Errorf("ปีบัญชีมีข้อมูลอ้างอิง เปลี่ยนช่วงวันหรือทศนิยมไม่ได้")
	}
	var overlap bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND id<>$2 AND payload->>'startdate'<=$3 AND payload->>'enddate'>=$4 AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, cmd.ID, next.EndDate, next.StartDate).Scan(&overlap)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, fmt.Errorf("ช่วงปีบัญชีทับซ้อน")
	}
	if next.ProfitLossAccount != "" && next.ProfitLossAccount == next.RetainedEarningsAccount {
		return nil, fmt.Errorf("บัญชีกำไรขาดทุนและกำไรสะสมต้องเป็นคนละบัญชี")
	}
	lines := []Line{}
	for _, code := range []string{next.ProfitLossAccount, next.RetainedEarningsAccount} {
		if code != "" {
			lines = append(lines, Line{AccountCode: code})
		}
	}
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, lines)
	if err != nil {
		return nil, err
	}
	for _, line := range lines {
		a, ok := accounts[line.AccountCode]
		if !ok || !a.IsActive || !a.AllowPosting || a.AccountType != "equity" {
			return nil, fmt.Errorf("บัญชีปิดผลกำไรต้องเป็นบัญชีทุนที่เปิดรับรายการ")
		}
	}
	if err = s.checkDuplicateCode(ctx, tx, scope.Company, "fiscal-years", next.Code, cmd.ID); err != nil {
		return nil, err
	}
	if cmd.Action == "create" {
		next.Identity = newIdentity(scope, now)
	} else {
		next.Identity = updateIdentity(scope, old.Identity, now)
	}
	return pgChange("fiscal-years", next.ID, next.Code, next)
}

// Protect only openings actually emitted by year-end, not a user's free-text reference.
func (s *PostgresStore) protectGeneratedOpening(ctx context.Context, tx *sql.Tx, scope Scope, j Journal) error {
	if j.Kind != "opening" {
		return nil
	}
	var generated bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_events e WHERE e.company=$1 AND e.payload->>'action'='processes:year-end' AND EXISTS(SELECT 1 FROM jsonb_array_elements(e.payload->'changes') c WHERE c->>'kind'='journals' AND c->>'id'=$2))`, scope.Company, j.ID).Scan(&generated)
	if err != nil {
		return err
	}
	if generated {
		return fmt.Errorf("ยอดยกมาจากปิดปีแก้ไขหรือลบไม่ได้")
	}
	return nil
}

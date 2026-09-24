package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"time"
)

func (s *PostgresStore) mutateProcess(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if scope.Branch != "" {
		return nil, fmt.Errorf("การประมวลผลต้องใช้สิทธิ์ครอบคลุมทุกสาขา")
	}
	if strings.TrimSpace(cmd.Reason) == "" {
		return nil, fmt.Errorf("กรุณาระบุเหตุผลก่อนประมวลผล")
	}
	if cmd.Action == "recalculate" || cmd.Action == "reprocess" {
		if err := rebuildPGTransaction(ctx, tx, scope); err != nil {
			return nil, err
		}
		return []Change{}, nil
	}
	if cmd.Action != "close" && cmd.Action != "year-end" {
		return nil, fmt.Errorf("คำสั่งประมวลผลไม่ถูกต้อง")
	}
	year, err := s.openPGYear(ctx, tx, scope, cmd.ID)
	if err != nil {
		return nil, err
	}
	if err = checkVersion(cmd, year.Identity); err != nil {
		return nil, err
	}
	// Several branches get "-001" style suffixes, so 4 characters of doc_no are reserved for them.
	if err := checkCode(cmd.DocNo, "docno", "เลขที่เอกสารประมวลผล", DocNoMaxRunes-4); err != nil {
		return nil, err
	}
	if !validDate(cmd.Date) {
		return nil, fieldError("process_date_invalid", "date", "วันที่ประมวลผลไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	}
	to := cmd.Date
	if cmd.Action == "year-end" {
		to = year.EndDate
	}
	if to < year.StartDate || to > year.EndDate {
		return nil, fmt.Errorf("วันที่ประมวลผลอยู่นอกปีบัญชี")
	}
	var pending bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='journals' AND payload->>'fiscalyear'=$2 AND payload->>'date'<=$3 AND payload->>'status'='draft' AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, year.Code, to).Scan(&pending)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, fmt.Errorf("ยังมีรายการร่าง กรุณาผ่านรายการก่อนประมวลผล")
	}
	snapshot, err := processBalancesTx(ctx, tx, scope, year.Code, to)
	if err != nil {
		return nil, err
	}
	target := year
	prepared := &preparedProcess{Year: year}
	if cmd.Action == "close" {
		if year.ProfitLossAccount == "" || year.RetainedEarningsAccount == "" {
			return nil, fmt.Errorf("กรุณากำหนดบัญชีกำไรขาดทุนและกำไรสะสม")
		}
	} else {
		target, err = s.openPGYear(ctx, tx, scope, cmd.TargetYear)
		if err != nil {
			return nil, err
		}
		end, _ := time.Parse("2006-01-02", year.EndDate)
		if target.StartDate != end.AddDate(0, 0, 1).Format("2006-01-02") || target.Scale != year.Scale || cmd.Date != target.StartDate {
			return nil, fmt.Errorf("ปีถัดไปและวันที่ยอดยกมาต้องต่อเนื่องกับปีเดิมและใช้ทศนิยมเดียวกัน")
		}
		var exists bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='journals' AND payload->>'fiscalyear'=$2 AND payload->>'kind'='opening' AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, scope.Company, target.Code).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("ปีถัดไปมียอดยกมาแล้ว")
		}
		prepared.CloseYear = true
	}
	byBranch := map[string][]Line{}
	type dimension struct{ branch, department, project string }
	closing := map[dimension]decimal.Decimal{}
	for _, row := range snapshot.Rows {
		value, err := decimal.NewFromString(row.Balance)
		if err != nil {
			return nil, err
		}
		if value.IsZero() {
			continue
		}
		isResult := row.AccountType == "income" || row.AccountType == "expense"
		if cmd.Action == "close" && !isResult {
			continue
		}
		if cmd.Action == "year-end" && isResult {
			return nil, fmt.Errorf("ยังไม่ได้ผ่านรายการปิดรายได้และค่าใช้จ่าย กรุณาปิดงบก่อนประมวลผลสิ้นปี")
		}
		if cmd.Action == "close" {
			value = value.Neg()
			key := dimension{row.BranchCode, row.DepartmentCode, row.ProjectCode}
			closing[key] = closing[key].Add(value)
		}
		line, err := signedLine(row.AccountCode, value, "ยอดประมวลผล "+year.Code)
		if err != nil {
			return nil, err
		}
		line.DepartmentCode = row.DepartmentCode
		line.ProjectCode = row.ProjectCode
		byBranch[row.BranchCode] = append(byBranch[row.BranchCode], line)
	}
	keys := make([]dimension, 0, len(closing))
	for key := range closing {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.branch != b.branch {
			return a.branch < b.branch
		}
		if a.department != b.department {
			return a.department < b.department
		}
		return a.project < b.project
	})
	for _, key := range keys {
		balance := closing[key]
		if balance.IsZero() {
			continue
		}
		for _, item := range []struct {
			code  string
			value decimal.Decimal
		}{{year.ProfitLossAccount, balance.Neg()}, {year.ProfitLossAccount, balance}, {year.RetainedEarningsAccount, balance.Neg()}} {
			line, err := signedLine(item.code, item.value, "โอนผลกำไรขาดทุนเข้ากำไรสะสม")
			if err != nil {
				return nil, err
			}
			line.DepartmentCode = key.department
			line.ProjectCode = key.project
			byBranch[key.branch] = append(byBranch[key.branch], line)
		}
	}
	if len(byBranch) == 0 && cmd.Action == "close" {
		return nil, fmt.Errorf("ไม่มียอดคงเหลือที่ต้องประมวลผล")
	}
	// Generated vouchers go to a book chosen by its type, never by a fixed code.
	book, err := processBookCode(ctx, tx, scope.Company, prepared.CloseYear)
	if err != nil {
		return nil, err
	}
	branches := make([]string, 0, len(byBranch))
	for branch := range byBranch {
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	for i, branch := range branches {
		docno := cmd.DocNo
		if len(branches) > 1 {
			docno = fmt.Sprintf("%s-%03d", cmd.DocNo, i+1)
		}
		j := Journal{DocNo: docno, Date: cmd.Date, BookCode: book, FiscalYear: target.Code, Description: "ปิดงบบัญชี " + year.Code, Kind: "closing", Status: "draft", Reason: cmd.Reason, Reference: "CLOSE:" + year.Code + ":" + cmd.Date, BranchCode: branch, Lines: byBranch[branch]}
		if prepared.CloseYear {
			j.Kind = "opening"
			j.Reference = "YEAR-END:" + year.Code
			j.Description = "ยอดยกมาจากปี " + year.Code
		}
		if err := s.validateJournalLines(ctx, tx, scope, &j, true); err != nil {
			return nil, err
		}
		prepared.Journals = append(prepared.Journals, j)
	}
	changes := []Change{}
	for _, j := range prepared.Journals {
		if err = s.checkDuplicateCode(ctx, tx, scope.Company, "journals", j.DocNo, ""); err != nil {
			return nil, err
		}
		j.Identity = newIdentity(scope, now)
		part, e := pgChange("journals", j.ID, j.DocNo, j)
		if e != nil {
			return nil, e
		}
		changes = append(changes, part...)
	}
	if prepared.CloseYear {
		year.Closed = true
		year.Identity = updateIdentity(scope, year.Identity, now)
		part, e := pgChange("fiscal-years", year.ID, year.Code, year)
		if e != nil {
			return nil, e
		}
		changes = append(changes, part...)
	}
	return changes, nil
}

func processBalancesTx(ctx context.Context, tx *sql.Tx, scope Scope, year, to string) (ProcessBalanceSnapshot, error) {
	result := ProcessBalanceSnapshot{Rows: []ProcessBalance{}}
	rows, err := tx.QueryContext(ctx, `SELECT account_code,MAX(account_name),MAX(account_type),branch_code,department_code,project_code,SUM(debit-credit)::text FROM gl_lines WHERE company=$1 AND fiscal_year=$2 AND entry_date<=$3::date GROUP BY account_code,branch_code,department_code,project_code HAVING SUM(debit-credit)<>0 ORDER BY branch_code,department_code,project_code,account_code LIMIT $4`, scope.Company, year, to, ProcessBalanceRowLimit+1)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var b ProcessBalance
		if err = rows.Scan(&b.AccountCode, &b.AccountName, &b.AccountType, &b.BranchCode, &b.DepartmentCode, &b.ProjectCode, &b.Balance); err != nil {
			return result, err
		}
		result.Rows = append(result.Rows, b)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	if len(result.Rows) > ProcessBalanceRowLimit {
		return result, fmt.Errorf("ยอดแยกตามมิติเกินขนาดที่รองรับ")
	}
	return result, nil
}

func rebuildPGTransaction(ctx context.Context, tx *sql.Tx, scope Scope) error {
	var err error
	var version int64
	err = tx.QueryRowContext(ctx, `SELECT sequence FROM gl_projection_state WHERE company=$1`, scope.Company).Scan(&version)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1`, scope.Company); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM gl_records WHERE company=$1`, scope.Company); err != nil {
		return err
	}
	// Keyset batches bound memory; rows must close before statements on tx.
	for after := int64(0); after < version; {
		rows, err := tx.QueryContext(ctx, `SELECT sequence,event_hash,payload FROM gl_events WHERE company=$1 AND sequence>$2 ORDER BY sequence LIMIT 100`, scope.Company, after)
		if err != nil {
			return err
		}
		events := []Event{}
		for rows.Next() {
			var sequence int64
			var hash string
			var data []byte
			if err = rows.Scan(&sequence, &hash, &data); err != nil {
				rows.Close()
				return err
			}
			var event Event
			if err = json.Unmarshal(data, &event); err != nil {
				rows.Close()
				return err
			}
			actual, _, err := eventDigest(event)
			if err != nil || actual != hash || sequence != event.Sequence || sequence != after+int64(len(events))+1 {
				rows.Close()
				return fmt.Errorf("ประวัติบัญชีไม่ต่อเนื่องหรือไม่ตรงกับต้นฉบับ")
			}
			events = append(events, event)
		}
		rowErr := rows.Err()
		rows.Close()
		if rowErr != nil {
			return rowErr
		}
		if len(events) == 0 {
			return fmt.Errorf("ประวัติบัญชีขาดหาย")
		}
		for _, event := range events {
			if err = applyChanges(ctx, tx, event); err != nil {
				return err
			}
			after = event.Sequence
		}
		if after > version {
			return fmt.Errorf("ลำดับประวัติบัญชีเกินยอดประมวลผล")
		}
	}
	return nil
}

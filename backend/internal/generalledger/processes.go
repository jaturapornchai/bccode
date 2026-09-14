package generalledger

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"sort"
	"strings"
	"time"
)

type preparedProcess struct {
	Sequence  int64
	Journals  []Journal
	Year      FiscalYear
	CloseYear bool
}

// PostgreSQL supplies exact balances with their dimensions. Mongo serializes
// this snapshot against subsequent company writes before creating any drafts.
func (s *Store) prepareProcess(ctx context.Context, scope Scope, cmd Command) (*preparedProcess, error) {
	if scope.Branch != "" {
		return nil, fmt.Errorf("การประมวลผลบัญชีต้องใช้สิทธิ์ระดับบริษัทที่ครอบคลุมทุกสาขา")
	}
	if err := s.Ready(ctx, scope); err != nil {
		return nil, err
	}
	sequence, err := s.projection.Version(ctx, scope)
	if err != nil {
		return nil, err
	}
	prepared := &preparedProcess{Sequence: sequence}
	if strings.TrimSpace(cmd.Reason) == "" {
		return nil, fmt.Errorf("กรุณาระบุเหตุผลก่อนประมวลผล")
	}
	if cmd.Action == "recalculate" {
		return prepared, nil
	}
	if cmd.Action == "reprocess" {
		return nil, fmt.Errorf("เอกสารซื้อขายเดิมยังไม่มีจำนวนเงินต้นฉบับแบบทศนิยมแม่นยำและขอบเขตบริษัทที่ยืนยัน จึงยังสร้างบัญชีอัตโนมัติไม่ได้")
	}
	if !contains([]string{"close", "year-end"}, cmd.Action) {
		return nil, fmt.Errorf("คำสั่งประมวลผลไม่ถูกต้อง")
	}
	year, err := s.yearByCode(ctx, scope, cmd.ID)
	if err != nil {
		return nil, err
	}
	if year.Closed {
		return nil, fmt.Errorf("ปีบัญชีนี้ปิดแล้ว")
	}
	if cmd.Version != year.Version {
		return nil, fmt.Errorf("ปีบัญชีเปลี่ยนแล้ว กรุณาโหลดข้อมูลใหม่")
	}
	prepared.Year = year
	if !validCode(cmd.DocNo) || !validDate(cmd.Date) {
		return nil, fmt.Errorf("กรุณาระบุเลขที่เอกสารและวันที่ประมวลผล")
	}
	to := cmd.Date
	if cmd.Action == "year-end" {
		to = year.EndDate
	}
	if to < year.StartDate || to > year.EndDate {
		return nil, fmt.Errorf("วันที่ประมวลผลต้องอยู่ในปีบัญชี")
	}
	if pending, err := s.referenced(ctx, scope, bson.M{"fiscalyear": year.Code, "isdeleted": false, "status": "draft", "date": bson.M{"$lte": to}}); err != nil {
		return nil, err
	} else if pending {
		return nil, fmt.Errorf("ยังมีรายการร่าง กรุณาตรวจและผ่านรายการให้ครบก่อนประมวลผล")
	}
	reader, ok := s.projection.(interface {
		ProcessBalances(context.Context, Scope, string, string) (ProcessBalanceSnapshot, error)
	})
	if !ok {
		return nil, fmt.Errorf("ระบบรายงานยังไม่รองรับยอดประมวลผล")
	}
	snapshot, err := reader.ProcessBalances(ctx, scope, year.Code, to)
	if err != nil {
		return nil, err
	}
	if snapshot.Sequence != sequence {
		return nil, fmt.Errorf("ข้อมูลเปลี่ยนระหว่างเตรียมประมวลผล กรุณาตรวจยอดแล้วลองใหม่")
	}
	target := year
	if cmd.Action == "close" {
		if year.ProfitLossAccount == "" || year.RetainedEarningsAccount == "" {
			return nil, fmt.Errorf("กรุณากำหนดบัญชีกำไรขาดทุนและกำไรสะสมในปีบัญชีก่อน")
		}
	} else {
		if cmd.TargetYear == "" {
			return nil, fmt.Errorf("กรุณาเลือกปีบัญชีถัดไปสำหรับยอดยกมา")
		}
		target, err = s.yearByCode(ctx, scope, cmd.TargetYear)
		if err != nil {
			return nil, err
		}
		end, _ := time.Parse("2006-01-02", year.EndDate)
		if target.StartDate != end.AddDate(0, 0, 1).Format("2006-01-02") || target.Closed || target.Currency != year.Currency || target.Scale != year.Scale || cmd.Date != target.StartDate {
			return nil, fmt.Errorf("ปีถัดไปต้องต่อเนื่องและใช้สกุลเงินกับทศนิยมเดียวกัน วันที่เอกสารต้องเป็นวันเริ่มปีถัดไป")
		}
		if exists, err := s.referenced(ctx, scope, bson.M{"fiscalyear": target.Code, "kind": "opening", "isdeleted": false}); err != nil {
			return nil, err
		} else if exists {
			return nil, fmt.Errorf("ปีถัดไปมียอดยกมาแล้ว กรุณาตรวจสอบเพื่อไม่ให้ยกยอดซ้ำ")
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
		j := Journal{DocNo: docno, Date: cmd.Date, BookCode: "JV", FiscalYear: target.Code, Description: "ปิดงบบัญชี " + year.Code, Currency: year.Currency, Kind: "closing", Status: "draft", Reason: cmd.Reason, Reference: "CLOSE:" + year.Code + ":" + cmd.Date, BranchCode: branch, Lines: byBranch[branch]}
		if prepared.CloseYear {
			j.Kind = "opening"
			j.Reference = "YEAR-END:" + year.Code
			j.Description = "ยอดยกมาจากปี " + year.Code
		}
		if err := s.checkJournal(ctx, scope, &j); err != nil {
			return nil, err
		}
		prepared.Journals = append(prepared.Journals, j)
	}
	version, err := s.projection.Version(ctx, scope)
	if err != nil {
		return nil, err
	}
	if version != sequence {
		return nil, fmt.Errorf("ข้อมูลเปลี่ยนระหว่างเตรียมประมวลผล กรุณาตรวจยอดแล้วลองใหม่")
	}
	return prepared, nil
}

func signedLine(code string, value decimal.Decimal, description string) (Line, error) {
	line := Line{AccountCode: code, Description: description}
	amount, err := ParseAmount(value.Abs().String())
	if err != nil {
		return line, fmt.Errorf("ยอดรวมเกินขนาดจำนวนเงินที่รองรับ กรุณาตรวจสอบบัญชี %s", code)
	}
	if value.IsNegative() {
		line.Credit = amount
	} else {
		line.Debit = amount
	}
	return line, nil
}

func (s *Store) processMutation(ctx context.Context, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.prepared == nil {
		return nil, fmt.Errorf("กรุณาเตรียมข้อมูลก่อนประมวลผล")
	}
	if cmd.Action == "recalculate" {
		return []Change{}, nil
	}
	changes := []Change{}
	for i, j := range cmd.prepared.Journals {
		if err := s.checkJournal(ctx, scope, &j); err != nil {
			return nil, err
		}
		created := cmd
		created.Action = "create"
		created.RequestID = fmt.Sprintf("%s-%d", cmd.RequestID, i)
		j.Identity = identityFor(scope, created, Identity{}, now)
		change, err := s.save(ctx, scope, "journals", j.ID, j.DocNo, j, 0)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	if cmd.prepared.CloseYear {
		year, err := s.yearByCode(ctx, scope, cmd.prepared.Year.Code)
		if err != nil {
			return nil, err
		}
		if year.Version != cmd.prepared.Year.Version || year.Closed {
			return nil, fmt.Errorf("ปีบัญชีเปลี่ยนแล้ว กรุณาเตรียมประมวลผลใหม่")
		}
		yearCmd := cmd
		yearCmd.Action = "update"
		year.Identity = identityFor(scope, yearCmd, year.Identity, now)
		year.Closed = true
		change, err := s.save(ctx, scope, "fiscal-years", year.ID, year.Code, year, cmd.prepared.Year.Version)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// MaxGeneratedPeriods bounds "สร้างงวดอัตโนมัติ" (Champ 5511 F9): a fiscal year longer than this is
// almost surely a typing mistake in its dates, and one command must not write hundreds of periods.
const MaxGeneratedPeriods = 24

type periodSlot struct{ start, end string }

// monthlyPeriodSlots splits a fiscal year into calendar months clipped to the year's dates, so a
// year that starts mid-month gets a short first period and every other period is a whole month
// (the VAT month). Champ adds months to the first start date instead; calendar months keep each
// period inside one tax month.
func monthlyPeriodSlots(startDate, endDate string) ([]periodSlot, error) {
	start, startErr := time.Parse(time.DateOnly, startDate)
	end, endErr := time.Parse(time.DateOnly, endDate)
	if startErr != nil || endErr != nil || end.Before(start) {
		return nil, fieldError("fiscal_year_range_invalid", "enddate", "วันสิ้นสุดปีบัญชีต้องไม่ก่อนวันเริ่มต้น กรุณาแก้วันที่ให้ถูกต้อง")
	}
	slots := []periodSlot{}
	for current := start; !current.After(end); {
		last := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		if last.After(end) {
			last = end
		}
		if len(slots) == MaxGeneratedPeriods {
			return nil, &UserError{Code: "periods_too_many", Status: http.StatusConflict, Args: []string{strconv.Itoa(MaxGeneratedPeriods)},
				Message: fmt.Sprintf("ปีบัญชียาวเกิน %d เดือน จึงสร้างงวดรายเดือนอัตโนมัติไม่ได้ กรุณาตรวจวันเริ่มต้นและวันสิ้นสุดปีบัญชี", MaxGeneratedPeriods)}
		}
		slots = append(slots, periodSlot{start: current.Format(time.DateOnly), end: last.Format(time.DateOnly)})
		current = last.AddDate(0, 0, 1)
	}
	return slots, nil
}

// planMonthlyPeriods returns the monthly periods a fiscal year still needs (code <year>-NN, NN = month
// number in the year). A month that already has a period with exactly its dates is skipped; an
// existing period that cuts across a month would leave an overlap or a gap, so the plan is refused.
// The result together with the skipped periods covers the year day by day without overlap.
func planMonthlyPeriods(year FiscalYear, existing []Master) ([]Master, error) {
	slots, err := monthlyPeriodSlots(year.StartDate, year.EndDate)
	if err != nil {
		return nil, err
	}
	planned := []Master{}
	for index, slot := range slots {
		covered := false
		for _, period := range existing {
			if period.EndDate < slot.start || period.StartDate > slot.end {
				continue
			}
			if period.StartDate == slot.start && period.EndDate == slot.end {
				covered = true
				continue
			}
			return nil, &UserError{Code: "period_generate_overlap", Status: http.StatusConflict, Args: []string{period.Code},
				Message: fmt.Sprintf("งวด %s มีช่วงวันที่คร่อมเดือน ทำให้งวดรายเดือนทับซ้อนหรือขาดช่วง กรุณาลบงวด %s แล้วสร้างงวดอัตโนมัติอีกครั้ง", period.Code, period.Code)}
		}
		if !covered {
			planned = append(planned, Master{Kind: "periods", Code: fmt.Sprintf("%s-%02d", year.Code, index+1), FiscalYear: year.Code, StartDate: slot.start, EndDate: slot.end, IsActive: true})
		}
	}
	return planned, nil
}

// generatePeriods is periods action "generate": create the missing monthly periods of one open
// fiscal year (TargetYear = year code) in the command's transaction.
func (s *PostgresStore) generatePeriods(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.TargetYear == "" {
		return nil, fieldError("period_year_required", "targetyear", "กรุณาเลือกปีบัญชีก่อนสร้างงวดอัตโนมัติ")
	}
	year, err := s.openPGYear(ctx, tx, scope, cmd.TargetYear)
	if err != nil {
		return nil, err
	}
	existing, err := loadPeriodsInRange(ctx, tx, scope.Company, year.StartDate, year.EndDate, "")
	if err != nil {
		return nil, err
	}
	planned, err := planMonthlyPeriods(year, existing)
	if err != nil {
		return nil, err
	}
	if len(planned) == 0 {
		return nil, &UserError{Code: "periods_complete", Status: http.StatusConflict, Args: []string{year.Code},
			Message: fmt.Sprintf("ปีบัญชี %s มีงวดครบทุกเดือนแล้ว ไม่มีงวดที่ต้องสร้างเพิ่ม", year.Code)}
	}
	return s.newPeriodChanges(ctx, tx, scope, planned, nil, now)
}

// yearPeriodChanges keeps a fiscal year's periods in step with the year: deletes the periods of old
// (a deleted year or a year whose dates changed) and creates monthly periods for next (nil = delete).
// Only called while the year has no journal, budget or locked period (mutatePGFiscalYear "used").
func (s *PostgresStore) yearPeriodChanges(ctx context.Context, tx *sql.Tx, scope Scope, old FiscalYear, next *FiscalYear, now time.Time) ([]Change, error) {
	changes := []Change{}
	released := map[string]bool{}
	if old.ID != "" {
		periods, err := loadPeriods(ctx, tx, `company=$1 AND kind='periods' AND payload->>'fiscalyear'=$2`, scope.Company, old.Code)
		if err != nil {
			return nil, err
		}
		for _, period := range periods {
			period.Kind = "periods"
			period.IsDeleted = true
			period.Identity = updateIdentity(scope, period.Identity, now)
			change, err := pgChange("periods", period.ID, period.Code, period)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change...)
			released[period.Code] = true
		}
	}
	if next == nil {
		return changes, nil
	}
	// This year's own periods are being replaced above; only other years' periods can collide.
	existing, err := loadPeriodsInRange(ctx, tx, scope.Company, next.StartDate, next.EndDate, old.Code)
	if err != nil {
		return nil, err
	}
	planned, err := planMonthlyPeriods(*next, existing)
	if err != nil {
		return nil, err
	}
	created, err := s.newPeriodChanges(ctx, tx, scope, planned, released, now)
	if err != nil {
		return nil, err
	}
	return append(changes, created...), nil
}

// newPeriodChanges turns planned periods into create changes; released = codes freed by periods
// deleted in the same command (not yet visible as deleted to the duplicate check).
func (s *PostgresStore) newPeriodChanges(ctx context.Context, tx *sql.Tx, scope Scope, planned []Master, released map[string]bool, now time.Time) ([]Change, error) {
	changes := []Change{}
	for _, period := range planned {
		if err := checkCode(period.Code, "code", "รหัสงวดบัญชี", codeMaxRunes); err != nil {
			return nil, err
		}
		if !released[period.Code] {
			if err := s.checkDuplicateCode(ctx, tx, scope.Company, "periods", period.Code, ""); err != nil {
				return nil, err
			}
		}
		period.Identity = newIdentity(scope, now)
		change, err := pgChange("periods", period.ID, period.Code, period)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change...)
	}
	return changes, nil
}

// loadPeriodsInRange lists live periods touching [start, end], except those of skipYear.
func loadPeriodsInRange(ctx context.Context, tx *sql.Tx, company, start, end, skipYear string) ([]Master, error) {
	return loadPeriods(ctx, tx, `company=$1 AND kind='periods' AND payload->>'startdate'<=$3 AND payload->>'enddate'>=$2 AND ($4='' OR COALESCE(payload->>'fiscalyear','')<>$4)`, company, start, end, skipYear)
}

func loadPeriods(ctx context.Context, tx *sql.Tx, where string, args ...any) ([]Master, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload || jsonb_build_object('id',id,'version',version) FROM gl_records WHERE `+where+` AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY payload->>'startdate', id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	periods := []Master{}
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var period Master
		if err := json.Unmarshal(data, &period); err != nil {
			return nil, err
		}
		periods = append(periods, period)
	}
	return periods, rows.Err()
}

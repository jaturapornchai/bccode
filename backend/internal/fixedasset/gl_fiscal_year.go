package fixedasset

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	gl "smlcloudplatform/internal/generalledger"
)

// Error codes of the voucher date / GL fiscal-year checks; the screen text of each is
// gl_err_<code> in languages.tsv (th column = Message).
const (
	codeDateInvalid           = "fa_date_invalid"
	codePostYearInvalid       = "fa_post_year_invalid"
	codeFiscalYearNotFound    = "fa_fiscal_year_not_found"
	codeFiscalYearClosed      = "fa_fiscal_year_closed"
	codeFiscalYearAmbiguous   = "fa_fiscal_year_ambiguous"
	codeDateOutsidePeriodYear = "fa_date_outside_period_year"
)

// checkVoucherDate refuses a voucher date that is not a real YYYY-MM-DD (ค.ศ.) date, before it
// is used to find the fiscal year — a typo would otherwise read as "no fiscal year".
func checkVoucherDate(date, field string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return faFieldError(codeDateInvalid, field, fmt.Sprintf("วันที่ %q ไม่ถูกต้อง กรุณาระบุเป็น ปี ค.ศ.-เดือน-วัน เช่น 2026-07-31", date))
	}
	return nil
}

// depreciationVoucherDate checks the ค.ศ. year of the schedule period to post (the calendar year
// calculator.go writes, same bounds as the tax report) and returns it canonical, the voucher date
// and the period's last day. A blank date is the period's last day: the screen sends no date, and
// the expense belongs to the period it covers (accrual) — "today" would file December's depreciation
// posted in January into the next fiscal year.
func depreciationVoucherDate(fiscalYear string, period int, date string) (year, voucherDate, periodEnd string, err error) {
	y, convErr := strconv.Atoi(strings.TrimSpace(fiscalYear))
	if convErr != nil || y < taxReportMinYear || y > taxReportMaxYear {
		return "", "", "", faFieldError(codePostYearInvalid, "fiscalyear", fmt.Sprintf("ปีของงวดค่าเสื่อมราคา %q ไม่ถูกต้อง กรุณาระบุปี ค.ศ. 4 หลัก เช่น 2026 (ถ้าเป็นปี พ.ศ. ให้ลบ 543)", fiscalYear))
	}
	periodEnd = time.Date(y, time.Month(period)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	if date == "" {
		date = periodEnd
	}
	if err := checkVoucherDate(date, "date"); err != nil {
		return "", "", "", err
	}
	return strconv.Itoa(y), date, periodEnd, nil
}

// depreciationFiscalYear returns the GL fiscal-year code of a depreciation voucher: the year that
// holds the period's last day, which the voucher date must share — a period is never filed in
// another year's statements, and a closed year stays closed.
func depreciationFiscalYear(years []gl.FiscalYear, year string, period int, date, periodEnd string) (string, error) {
	periodCode, err := pickFiscalYear(years, periodEnd, "fiscalyear")
	if err != nil {
		return "", err
	}
	code, err := pickFiscalYear(years, date, "date")
	if err != nil {
		return "", err
	}
	if code != periodCode {
		return "", faFieldError(codeDateOutsidePeriodYear, "date", fmt.Sprintf("วันที่ใบสำคัญ %s อยู่ในปีบัญชี %s แต่ค่าเสื่อมราคางวด %d/%s อยู่ในปีบัญชี %s กรุณาใช้วันที่ในปีบัญชี %s หรือเว้นว่างไว้ ระบบจะใช้วันสิ้นงวด %s", date, code, period, year, periodCode, periodCode, periodEnd))
	}
	return code, nil
}

// fiscalYears reads the company's GL fiscal years (the list already drops deleted ones).
func (p *GLPoster) fiscalYears(ctx context.Context, scope Scope) ([]gl.FiscalYear, error) {
	page, err := p.ledger.List(ctx, glScope(scope), "fiscal-years", "", 1, 1000, gl.ListFilter{})
	if err != nil {
		return nil, err
	}
	years := make([]gl.FiscalYear, 0, len(page.Items))
	for _, raw := range page.Items {
		var year gl.FiscalYear
		if err := json.Unmarshal(raw, &year); err != nil {
			return nil, err
		}
		years = append(years, year)
	}
	return years, nil
}

// fiscalYearCodeAt returns the code of the company's GL fiscal year that contains date — a voucher
// belongs to the year of its own date, as in Champ (BCPeriod: DocDate BETWEEN StartDate AND StopDate).
// A ค.ศ. year is never a GL fiscal-year code: a company may code its years in พ.ศ. ("2569") and start
// them in any month. GL re-checks the year inside its own transaction.
func (p *GLPoster) fiscalYearCodeAt(ctx context.Context, scope Scope, date, field string) (string, error) {
	years, err := p.fiscalYears(ctx, scope)
	if err != nil {
		return "", err
	}
	return pickFiscalYear(years, date, field)
}

// pickFiscalYear chooses the one fiscal year whose [startdate, enddate] holds date (both ends
// inclusive). No year, a closed or inactive year, or overlapping years are refused with a field
// error — never the first row, never the nearest year, never a ค.ศ. year used as the code.
func pickFiscalYear(years []gl.FiscalYear, date, field string) (string, error) {
	var found []gl.FiscalYear
	for _, year := range years {
		if !year.IsDeleted && year.StartDate <= date && date <= year.EndDate {
			found = append(found, year)
		}
	}
	switch {
	case len(found) == 0:
		return "", faFieldError(codeFiscalYearNotFound, field, fmt.Sprintf("ไม่พบปีบัญชีที่ครอบคลุมวันที่ %s กรุณาเพิ่มปีบัญชีที่เมนู ปีบัญชีและบัญชีปิดปี หรือเลือกวันที่อื่น", date))
	case len(found) > 1:
		return "", faFieldError(codeFiscalYearAmbiguous, field, fmt.Sprintf("วันที่ %s อยู่ในปีบัญชีมากกว่าหนึ่งปี (%s, %s) กรุณาแก้ช่วงวันที่ของปีบัญชีที่เมนู ปีบัญชีและบัญชีปิดปี ให้ไม่ทับซ้อนกัน", date, found[0].Code, found[1].Code))
	}
	year := found[0]
	if year.Closed || !year.IsActive {
		return "", faFieldError(codeFiscalYearClosed, field, fmt.Sprintf("วันที่ %s อยู่ในปีบัญชี %s ที่ปิดแล้วหรือไม่ได้เปิดใช้งาน กรุณาเลือกวันที่ในปีบัญชีที่เปิดอยู่", date, year.Code))
	}
	return year.Code, nil
}

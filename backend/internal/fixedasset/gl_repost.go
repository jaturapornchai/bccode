package fixedasset

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	authmodels "smlcloudplatform/internal/authentication/models"
	gl "smlcloudplatform/internal/generalledger"
)

// Error codes of posting a period again after its journal was reversed; the screen text of each is
// gl_err_<code> in languages.tsv (th column = Message).
const (
	codeDocNoReversed           = "fa_docno_reversed"
	codeDocNoInUse              = "fa_docno_in_use"
	codeDocNoTooLongToReverse   = "fa_docno_too_long_to_reverse"
	codeJournalNotPosted        = "fa_journal_not_posted"
	codePeriodPostInProgress    = "fa_period_post_in_progress"
	reversalDocNoPrefix         = "REV-"
	journalStatusPosted         = "posted"
	journalStatusReversed       = "reversed"
	journalStatusesPageLimit    = 100
	reversalDocNoMaxSourceRunes = gl.DocNoMaxRunes - len(reversalDocNoPrefix)
	periodPostLockClass         = "fa-depreciation-post"
)

// businessDate is the Thai calendar date of now: a voucher dated from "today" between 00:00 and
// 07:00 in Thailand must not carry yesterday's UTC date.
func businessDate(now time.Time) string {
	return now.In(authmodels.HoldingLocation("")).Format("2006-01-02")
}

// journalInfo is what choosing a voucher number needs to know about a live journal.
type journalInfo struct {
	Status    string
	Reference string
}

// journalsByDocNo returns every live journal whose document number, description or reference
// contains search, keyed by exact document number. GL keeps a reversed journal and its number for
// audit and checks numbers company-wide, so the lookup ignores the session branch and reads every
// page.
func (p *GLPoster) journalsByDocNo(ctx context.Context, scope Scope, search string) (map[string]journalInfo, error) {
	sc := glScope(scope)
	sc.Branch = ""
	journals := map[string]journalInfo{}
	for page := 1; ; page++ {
		result, err := p.ledger.List(ctx, sc, "journals", search, page, journalStatusesPageLimit, gl.ListFilter{})
		if err != nil {
			return nil, err
		}
		for _, raw := range result.Items {
			var journal gl.Journal
			if err := json.Unmarshal(raw, &journal); err != nil {
				return nil, err
			}
			if !journal.IsDeleted {
				journals[journal.DocNo] = journalInfo{Status: journal.Status, Reference: journal.Reference}
			}
		}
		if len(result.Items) == 0 || int64(page*journalStatusesPageLimit) >= result.Total {
			return journals, nil
		}
	}
}

// depreciationDocNo picks the voucher number of the depreciation posting whose GL reference is
// reference. The GL request IDs are derived from the number, so the number of an unfinished posting
// of the same period (a draft or posted journal no schedule row is marked with yet) is kept: the
// retry replays the same GL request instead of posting twice. Any other live number is never sent
// again — GL would replay the old journal (reversed, or already marked on the schedule) or refuse a
// number another document uses — so a generated number moves on to the next suffix
// (GJ-FA-2026-07-2, -3, ...) and a typed number is refused on the number field.
func (p *GLPoster) depreciationDocNo(ctx context.Context, scope Scope, docNo, reference string, generated bool, marked func(docNo string) (bool, error)) (string, error) {
	journals, err := p.journalsByDocNo(ctx, scope, docNo)
	if err != nil {
		return "", err
	}
	for n := 1; ; n++ {
		candidate := docNo
		if n > 1 {
			candidate = fmt.Sprintf("%s-%d", docNo, n)
		}
		if journal, exists := journals[candidate]; exists {
			retry := false
			if journal.Status != journalStatusReversed && journal.Reference == reference {
				done, err := marked(candidate)
				if err != nil {
					return "", err
				}
				retry = !done
			}
			if !retry {
				if generated {
					continue
				}
				if journal.Status == journalStatusReversed {
					return "", faFieldError(codeDocNoReversed, "docno", fmt.Sprintf("เลขที่ใบสำคัญ %s ถูกกลับรายการแล้ว ใช้ซ้ำไม่ได้ กรุณาระบุเลขที่ใหม่ หรือเว้นว่างไว้ให้ระบบสร้างเลขที่ให้", candidate))
				}
				return "", faFieldError(codeDocNoInUse, "docno", fmt.Sprintf("เลขที่ใบสำคัญ %s ถูกใช้กับเอกสารอื่นแล้ว กรุณาระบุเลขที่ใหม่ หรือเว้นว่างไว้ให้ระบบสร้างเลขที่ให้", candidate))
			}
		}
		// The reversal is named REV-<number> within the GL length; a cut name would collide with
		// the reversal of another number that starts the same way.
		if utf8.RuneCountInString(candidate) > reversalDocNoMaxSourceRunes {
			return "", faFieldError(codeDocNoTooLongToReverse, "docno", fmt.Sprintf("เลขที่ใบสำคัญ %s ยาวเกิน %d ตัวอักษร จึงกลับรายการภายหลังไม่ได้ กรุณาระบุเลขที่ใบสำคัญไม่เกิน %d ตัวอักษร", candidate, reversalDocNoMaxSourceRunes, reversalDocNoMaxSourceRunes))
		}
		return candidate, nil
	}
}

// depreciationRowsMarked reports whether a finished posting already marked schedule rows with docNo.
func depreciationRowsMarked(ctx context.Context, db *sql.DB, company, docNo string) (bool, error) {
	n, err := countRecords(ctx, db, company, kindDepreciation, ` AND payload->>'journaldocno' = $3 AND COALESCE((payload->>'isposted')::boolean, false)`, docNo)
	return n > 0, err
}

// checkJournalPosted refuses to mark schedule rows against a journal that GL does not hold as
// posted — a replayed GL request returns its old result even after the journal was reversed.
func (p *GLPoster) checkJournalPosted(ctx context.Context, scope Scope, docNo string) error {
	journals, err := p.journalsByDocNo(ctx, scope, docNo)
	if err != nil {
		return err
	}
	if status := journals[docNo].Status; status != journalStatusPosted {
		return faFieldError(codeJournalNotPosted, "", fmt.Sprintf("ใบสำคัญ %s ไม่ได้อยู่ในสถานะผ่านรายการ (สถานะ %q) ระบบจึงยังไม่ทำเครื่องหมายค่าเสื่อมราคาว่าผ่านแล้ว กรุณาตรวจใบสำคัญที่เมนูสมุดรายวัน", docNo, status))
	}
	return nil
}

// lockDepreciationPeriod serializes the postings of one company period: two postings started
// together (for example with different typed voucher numbers) would both read the same unposted
// rows and post them twice. The GL calls run in their own transactions, so the lock is a session
// lock on a connection of its own. It is only tried, never waited for, so a queued request never
// holds a pooled connection that the running posting needs.
func lockDepreciationPeriod(ctx context.Context, db *sql.DB, company, year string, period int) (func(), error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s:%s:%d", company, year, period)
	var locked bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtext($1), hashtext($2))`, periodPostLockClass, key).Scan(&locked); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if !locked {
		_ = conn.Close()
		return nil, &gl.UserError{Code: codePeriodPostInProgress, Status: http.StatusConflict, Message: fmt.Sprintf("ค่าเสื่อมราคางวด %d/%s กำลังถูกผ่านรายการโดยผู้ใช้อื่น กรุณารอสักครู่แล้วตรวจสถานะก่อนลองใหม่", period, year)}
	}
	return func() {
		// A fresh context: the request may already be cancelled, and the lock must still be freed.
		if _, err := conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtext($1), hashtext($2))`, periodPostLockClass, key); err != nil {
			// Drop the connection instead of pooling it, so the session lock ends with it.
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
		_ = conn.Close()
	}, nil
}

// reversalDate dates the reversal of a depreciation journal on the journal's own date while GL
// still takes vouchers on that date: the period then nets to zero and posting it again puts the
// corrected figure back into the same period, as Champ re-transfers into the same period. A date
// in a closed fiscal year or a locked period cannot change any more, so the reversal goes to today
// by the Thai calendar (never before the original); posting that period again is refused by the
// same closed year or locked period, so nothing is counted twice.
func (p *GLPoster) reversalDate(ctx context.Context, scope Scope, original string, now time.Time) (string, error) {
	open, err := p.dateOpen(ctx, scope, original)
	if err != nil {
		return "", err
	}
	if open {
		return original, nil
	}
	date := businessDate(now)
	if date < original {
		date = original
	}
	// GL files the reversal in the fiscal year of its date; say so in Thai before GL is asked.
	if _, err := p.fiscalYearCodeAt(ctx, scope, date, ""); err != nil {
		return "", err
	}
	return date, nil
}

// dateOpen reports whether GL still takes a voucher dated date: an open, active fiscal year holds
// it and no locked period covers it.
func (p *GLPoster) dateOpen(ctx context.Context, scope Scope, date string) (bool, error) {
	if _, err := p.fiscalYearCodeAt(ctx, scope, date, ""); err != nil {
		if user, ok := gl.AsUserError(err); ok && (user.Code == codeFiscalYearClosed || user.Code == codeFiscalYearNotFound) {
			return false, nil
		}
		return false, err
	}
	page, err := p.ledger.List(ctx, glScope(scope), "periods", "", 1, 1000, gl.ListFilter{})
	if err != nil {
		return false, err
	}
	for _, raw := range page.Items {
		var period gl.Master
		if err := json.Unmarshal(raw, &period); err != nil {
			return false, err
		}
		if !period.IsDeleted && period.Locked && period.StartDate <= date && date <= period.EndDate {
			return false, nil
		}
	}
	return true, nil
}

// setDepreciationPosted writes the posted flag of the schedule rows of one journal in a single
// transaction: rows written one by one could leave part of a period posted, and the next posting
// of that period would send different lines under the same number.
func setDepreciationPosted(ctx context.Context, db *sql.DB, scope Scope, items []DepreciationScheduleItem, docNo string, now time.Time) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, it := range items {
		it.IsPosted = docNo != ""
		it.JournalDocNo = docNo
		it.PostedAt = nil
		if it.IsPosted {
			it.PostedAt = &now
		}
		it.UpdatedAt = now
		it.UpdatedBy = scope.Actor
		if err := putRecord(ctx, tx, scope.Company, kindDepreciation, it.ID, depreciationKey(it), it); err != nil {
			return err
		}
	}
	return tx.Commit()
}

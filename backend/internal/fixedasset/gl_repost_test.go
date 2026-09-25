package fixedasset

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	gl "smlcloudplatform/internal/generalledger"
)

// statusLedger serves journals for the voucher-number lookup the way GL List does: a
// case-insensitive substring on the number, pages of the requested size, and the total count. It
// fails the test on any write.
type statusLedger struct {
	t        *testing.T
	journals []gl.Journal
	scopes   []gl.Scope
}

func (l *statusLedger) Execute(context.Context, gl.Scope, gl.Command) (gl.Result, error) {
	l.t.Fatal("choosing a voucher number must not write to GL")
	return gl.Result{}, nil
}

func (l *statusLedger) List(_ context.Context, scope gl.Scope, resource, search string, page, limit int, _ gl.ListFilter) (gl.Page, error) {
	l.scopes = append(l.scopes, scope)
	if resource != "journals" {
		l.t.Fatalf("unexpected list %s", resource)
	}
	var found []gl.Journal
	for _, j := range l.journals {
		if strings.Contains(strings.ToLower(j.DocNo), strings.ToLower(search)) {
			found = append(found, j)
		}
	}
	result := gl.Page{Total: int64(len(found)), Page: page, Limit: limit}
	for i := (page - 1) * limit; i < len(found) && i < page*limit; i++ {
		raw, err := json.Marshal(found[i])
		if err != nil {
			l.t.Fatal(err)
		}
		result.Items = append(result.Items, raw)
	}
	return result, nil
}

const july = "FA-2026-07" // the GL reference of the July 2026 depreciation posting

func journalWith(docNo, status, reference string) gl.Journal {
	return gl.Journal{DocNo: docNo, Status: status, Reference: reference}
}

// markedRows stands in for the schedule rows: the numbers a finished posting marked on them.
func markedRows(docNos ...string) func(string) (bool, error) {
	return func(docNo string) (bool, error) {
		for _, d := range docNos {
			if d == docNo {
				return true, nil
			}
		}
		return false, nil
	}
}

func TestDepreciationDocNoSkipsNumbersThatCannotBeReplayed(t *testing.T) {
	ctx := context.Background()
	scope := Scope{Holding: "RUNGRUENG", Company: "01", Branch: "00100", Actor: "fa-unit"}
	for _, tc := range []struct {
		name     string
		journals []gl.Journal
		marked   []string
		want     string
	}{
		{"first posting", nil, nil, "GJ-FA-2026-07"},
		{"retry after GL posted but before the rows were marked", []gl.Journal{journalWith("GJ-FA-2026-07", "posted", july)}, nil, "GJ-FA-2026-07"},
		{"retry after the post step failed on a draft", []gl.Journal{journalWith("GJ-FA-2026-07", "draft", july)}, nil, "GJ-FA-2026-07"},
		{"July finished, then an asset registered late adds a July row", []gl.Journal{journalWith("GJ-FA-2026-07", "posted", july)}, []string{"GJ-FA-2026-07"}, "GJ-FA-2026-07-2"},
		{"posted again after a reversal", []gl.Journal{journalWith("GJ-FA-2026-07", "reversed", july), journalWith("REV-GJ-FA-2026-07", "posted", "GJ-FA-2026-07")}, nil, "GJ-FA-2026-07-2"},
		{"third time", []gl.Journal{journalWith("GJ-FA-2026-07", "reversed", july), journalWith("GJ-FA-2026-07-2", "reversed", july)}, nil, "GJ-FA-2026-07-3"},
		{"retry of the second number", []gl.Journal{journalWith("GJ-FA-2026-07", "reversed", july), journalWith("GJ-FA-2026-07-2", "posted", july)}, nil, "GJ-FA-2026-07-2"},
		{"the number was typed on another voucher in the journal screen", []gl.Journal{journalWith("GJ-FA-2026-07", "posted", "")}, nil, "GJ-FA-2026-07-2"},
	} {
		ledger := &statusLedger{t: t, journals: tc.journals}
		got, err := NewGLPoster(nil, ledger, nil).depreciationDocNo(ctx, scope, "GJ-FA-2026-07", july, true, markedRows(tc.marked...))
		if err != nil || got != tc.want {
			t.Fatalf("%s: got %q %v, want %q", tc.name, got, err, tc.want)
		}
		// GL checks numbers company-wide, so the lookup must not be limited to the session branch.
		for _, s := range ledger.scopes {
			if s.Branch != "" || s.Company != "01" || s.Holding != "RUNGRUENG" {
				t.Fatalf("%s: lookup scope %+v must be company-wide", tc.name, s)
			}
		}
	}
}

func TestDepreciationDocNoReadsEveryPage(t *testing.T) {
	// 150 journals match the search and the reversed base sorts onto page 2.
	journals := make([]gl.Journal, 0, 151)
	for i := 0; i < 150; i++ {
		journals = append(journals, journalWith(fmt.Sprintf("GJ-FA-2026-07-X%03d", i), "posted", ""))
	}
	journals = append(journals, journalWith("GJ-FA-2026-07", "reversed", july))
	ledger := &statusLedger{t: t, journals: journals}
	got, err := NewGLPoster(nil, ledger, nil).depreciationDocNo(context.Background(), Scope{Company: "01"}, "GJ-FA-2026-07", july, true, markedRows())
	if err != nil || got != "GJ-FA-2026-07-2" {
		t.Fatalf("got %q %v; the reversed number on page 2 must be seen", got, err)
	}
	if len(ledger.scopes) != 2 {
		t.Fatalf("expected 2 pages to be read, read %d", len(ledger.scopes))
	}
}

func TestDepreciationDocNoRefusesTypedNumbersItCannotUse(t *testing.T) {
	ctx := context.Background()
	scope := Scope{Company: "01"}
	ledger := &statusLedger{t: t, journals: []gl.Journal{
		journalWith("FA-DEP-2569-07", "reversed", july),
		journalWith("FA-DEP-2569-06", "posted", "FA-2026-06"),
		journalWith("JV-RENT-0925", "posted", ""),
		journalWith("FA-DEP-2569-07B", "posted", july),
		journalWith("FA-DEP-2569-07C", "posted", july),
	}}
	poster := NewGLPoster(nil, ledger, nil)
	marked := markedRows("FA-DEP-2569-06", "FA-DEP-2569-07C")
	for _, tc := range []struct{ docNo, code string }{
		{"FA-DEP-2569-07", codeDocNoReversed},                      // reversed: GL would replay the reversed journal
		{"FA-DEP-2569-06", codeDocNoInUse},                         // June's finished posting
		{"JV-RENT-0925", codeDocNoInUse},                           // another document
		{"FA-DEP-2569-07C", codeDocNoInUse},                        // July already finished under this number
		{"FA-DEP-2569-JULY-0001-ABCDE", codeDocNoTooLongToReverse}, // 27 characters: REV- would not fit
	} {
		_, err := poster.depreciationDocNo(ctx, scope, tc.docNo, july, false, marked)
		user, ok := gl.AsUserError(err)
		if !ok || user.Code != tc.code || user.Field != "docno" || !strings.Contains(user.Message, tc.docNo) {
			t.Fatalf("%s: got %#v, want %s on docno", tc.docNo, err, tc.code)
		}
	}
	// A new number, and the retry of this period's unfinished posting, are sent as typed.
	for _, docNo := range []string{"FA-DEP-2569-08", "FA-DEP-2569-07B"} {
		if got, err := poster.depreciationDocNo(ctx, scope, docNo, july, false, marked); err != nil || got != docNo {
			t.Fatalf("%s: got %q %v", docNo, got, err)
		}
	}

	// A 15-character book: the next generated number would not fit inside REV-<number>, so it is
	// refused before posting instead of failing at the next reversal.
	base := "GENERALBOOK0001-FA-2026-07"
	ledger = &statusLedger{t: t, journals: []gl.Journal{journalWith(base, "reversed", july)}}
	_, err := NewGLPoster(nil, ledger, nil).depreciationDocNo(ctx, scope, base, july, true, markedRows())
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeDocNoTooLongToReverse || user.Field != "docno" {
		t.Fatalf("long book: %#v", err)
	}
	// The first posting of the same base still fits (REV- + 26 characters = 30).
	if got, err := NewGLPoster(nil, &statusLedger{t: t}, nil).depreciationDocNo(ctx, scope, base, july, true, markedRows()); err != nil || got != base {
		t.Fatalf("long book first posting: %q %v", got, err)
	}
}

func TestDepreciationDocNoPassesOnTheRowLookupError(t *testing.T) {
	ledger := &statusLedger{t: t, journals: []gl.Journal{journalWith("GJ-FA-2026-07", "posted", july)}}
	failed := fmt.Errorf("database gone")
	_, err := NewGLPoster(nil, ledger, nil).depreciationDocNo(context.Background(), Scope{Company: "01"}, "GJ-FA-2026-07", july, true, func(string) (bool, error) { return false, failed })
	if err != failed {
		t.Fatalf("got %v, want the lookup error", err)
	}
}

func TestCheckJournalPostedRefusesAReplayedReversal(t *testing.T) {
	ctx := context.Background()
	ledger := &statusLedger{t: t, journals: []gl.Journal{journalWith("GJ-FA-2026-07", "reversed", july), journalWith("GJ-FA-2026-07-2", "posted", july)}}
	poster := NewGLPoster(nil, ledger, nil)
	err := poster.checkJournalPosted(ctx, Scope{Company: "01"}, "GJ-FA-2026-07")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeJournalNotPosted || !strings.Contains(user.Message, "reversed") {
		t.Fatalf("reversed journal: %#v", err)
	}
	if err := poster.checkJournalPosted(ctx, Scope{Company: "01"}, "GJ-FA-2026-07-2"); err != nil {
		t.Fatalf("posted journal: %v", err)
	}
	err = poster.checkJournalPosted(ctx, Scope{Company: "01"}, "GJ-FA-2026-08")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeJournalNotPosted {
		t.Fatalf("missing journal: %#v", err)
	}
}

func TestBusinessDateIsTheThaiCalendarDate(t *testing.T) {
	for _, tc := range []struct {
		now  time.Time
		want string
	}{
		{time.Date(2026, 10, 1, 16, 59, 59, 0, time.UTC), "2026-10-01"}, // 23:59:59 in Thailand
		{time.Date(2026, 10, 1, 17, 0, 0, 0, time.UTC), "2026-10-02"},   // midnight in Thailand
		{time.Date(2026, 12, 31, 18, 30, 0, 0, time.UTC), "2027-01-01"}, // new year in Thailand first
		{time.Date(2026, 10, 2, 3, 0, 0, 0, time.UTC), "2026-10-02"},
	} {
		if got := businessDate(tc.now); got != tc.want {
			t.Fatalf("%s: got %s, want %s", tc.now.Format(time.RFC3339), got, tc.want)
		}
	}
}

// calendarLedger serves fiscal years and periods and fails the test on any write.
type calendarLedger struct {
	t       *testing.T
	years   []gl.FiscalYear
	periods []gl.Master
}

func (l *calendarLedger) Execute(context.Context, gl.Scope, gl.Command) (gl.Result, error) {
	l.t.Fatal("choosing the reversal date must not write to GL")
	return gl.Result{}, nil
}

func (l *calendarLedger) List(_ context.Context, _ gl.Scope, resource, _ string, _, _ int, _ gl.ListFilter) (gl.Page, error) {
	page := gl.Page{}
	var items []any
	switch resource {
	case "fiscal-years":
		for _, y := range l.years {
			items = append(items, y)
		}
	case "periods":
		for _, p := range l.periods {
			items = append(items, p)
		}
	default:
		l.t.Fatalf("unexpected list %s", resource)
	}
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			l.t.Fatal(err)
		}
		page.Items = append(page.Items, raw)
	}
	return page, nil
}

func TestReversalDateKeepsTheOriginalDateWhileItIsOpen(t *testing.T) {
	ctx := context.Background()
	scope := Scope{Company: "01"}
	nightOf1October := time.Date(2026, 10, 1, 18, 30, 0, 0, time.UTC) // 2 October in Thailand
	ledger := &calendarLedger{t: t,
		years: []gl.FiscalYear{
			{Code: "2568", StartDate: "2025-01-01", EndDate: "2025-12-31", IsActive: true, Closed: true},
			{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true},
		},
		periods: []gl.Master{
			{Code: "2569-06", StartDate: "2026-06-01", EndDate: "2026-06-30", Locked: true},
			{Code: "2569-07", StartDate: "2026-07-01", EndDate: "2026-07-31"},
			{Code: "2569-08", StartDate: "2026-08-01", EndDate: "2026-08-31", Locked: true, Identity: gl.Identity{IsDeleted: true}},
		},
	}
	poster := NewGLPoster(nil, ledger, nil)
	for _, tc := range []struct{ original, want string }{
		{"2026-07-31", "2026-07-31"}, // open: the period nets to zero, the re-post lands in July again
		{"2026-08-31", "2026-08-31"}, // a deleted lock does not count
		{"2026-06-30", "2026-10-02"}, // locked period: today by the Thai calendar, not 1 October UTC
		{"2025-12-31", "2026-10-02"}, // closed year: today
	} {
		got, err := poster.reversalDate(ctx, scope, tc.original, nightOf1October)
		if err != nil || got != tc.want {
			t.Fatalf("%s: got %q %v, want %s", tc.original, got, err, tc.want)
		}
	}
	// Closed year and today outside every fiscal year: explained in Thai, nothing is chosen.
	_, err := poster.reversalDate(ctx, scope, "2025-12-31", time.Date(2027, 2, 1, 3, 0, 0, 0, time.UTC))
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearNotFound || !strings.Contains(user.Message, "2027-02-01") {
		t.Fatalf("no year for today: %#v", err)
	}
}

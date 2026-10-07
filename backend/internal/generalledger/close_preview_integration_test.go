//go:build integration

package generalledger

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/shopspring/decimal"
)

func (f *pgIntegrityFixture) closeTestJournal(doc, date, branch, dept, project, debit, credit, amount string) Journal {
	f.ensureDimension(DepartmentKind, dept)
	f.ensureDimension(ProjectKind, project)
	j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: "2026", Description: "test", Kind: "manual", BranchCode: branch,
		Lines: []Line{{AccountCode: debit, Debit: Amount(amount), Credit: Amount("0"), DepartmentCode: dept, ProjectCode: project}, {AccountCode: credit, Debit: Amount("0"), Credit: Amount(amount), DepartmentCode: dept, ProjectCode: project}}}
	return f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID)
}

// rowCounts: records, lines, events and the event sequence of the test company — a preview must change none of them.
func (f *pgIntegrityFixture) rowCounts() [4]int64 {
	f.t.Helper()
	var n [4]int64
	for i, query := range []string{
		`SELECT COUNT(*) FROM gl_records WHERE company='C'`,
		`SELECT COUNT(*) FROM gl_lines WHERE company='C'`,
		`SELECT COUNT(*) FROM gl_events WHERE company='C'`,
		`SELECT sequence FROM gl_projection_state WHERE company='C'`,
	} {
		if err := f.db.QueryRow(query).Scan(&n[i]); err != nil {
			f.t.Fatal(err)
		}
	}
	return n
}

func (f *pgIntegrityFixture) previewClose(date, docno string) (ClosePreview, error) {
	f.t.Helper()
	year := f.years["2026"]
	result, err := f.store.Execute(f.ctx, f.scope, Command{Resource: "processes", Action: ClosePreviewAction, ID: "2026", Version: year.Version, Date: date, DocNo: docno})
	if err != nil {
		return ClosePreview{}, err
	}
	if result.ClosePreview == nil {
		f.t.Fatal("preview answered without closepreview")
	}
	return *result.ClosePreview, nil
}

func (f *pgIntegrityFixture) closingDrafts() []Journal {
	f.t.Helper()
	page, err := f.store.List(f.ctx, f.scope, "journals", "", 1, 100, ListFilter{Kind: "closing"})
	if err != nil {
		f.t.Fatal(err)
	}
	drafts := []Journal{}
	for _, item := range page.Items {
		var j Journal
		if err = json.Unmarshal(item, &j); err != nil {
			f.t.Fatal(err)
		}
		drafts = append(drafts, j)
	}
	sort.Slice(drafts, func(i, j int) bool { return drafts[i].DocNo < drafts[j].DocNo })
	return drafts
}

func sameDecimal(t *testing.T, got, want string) {
	t.Helper()
	g, err := decimal.NewFromString(got)
	if err != nil {
		t.Fatalf("not a decimal: %q", got)
	}
	w, _ := decimal.NewFromString(want)
	if !g.Equal(w) {
		t.Fatalf("got %s want %s", got, want)
	}
}

// Champ 5504: the entries shown before saving must be the entries that get saved. The preview and the
// real close run one generator, the preview runs read-only, and closing up to a month is cumulative.
func TestClosePreviewEqualsCreatedClosingDrafts(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.post(f.closeTestJournal("S-B1-JAN", "2026-01-10", "B1", "D", "P", "1000", "4000", "100.50"))
	f.post(f.closeTestJournal("S-B2-JAN", "2026-01-20", "B2", "D2", "P2", "1000", "4000", "200"))
	f.post(f.closeTestJournal("E-B1-FEB", "2026-02-10", "B1", "D", "P", "5000", "1000", "30.25"))
	f.post(f.closeTestJournal("E-B2-FEB", "2026-02-12", "B2", "D2", "P2", "5000", "1000", "260"))
	f.post(f.closeTestJournal("S-B1-APR", "2026-04-10", "B1", "D", "P", "1000", "4000", "70"))
	pendingDraft := f.closeTestJournal("S-B1-DRAFT", "2026-02-15", "B1", "D", "P", "1000", "4000", "5")

	// A draft dated inside the range: the preview still shows the entries but flags it; the real close refuses.
	before := f.rowCounts()
	flagged, err := f.previewClose("2026-02-28", "")
	if err != nil || !flagged.PendingDrafts {
		t.Fatalf("pending draft not flagged: %+v %v", flagged, err)
	}
	year := f.years["2026"]
	f.deny(f.scope, Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE-FEB", Date: "2026-02-28", Reason: "close"})
	if after := f.rowCounts(); after != before {
		t.Fatalf("preview or refused close wrote rows: before %v after %v", before, after)
	}
	f.post(pendingDraft)

	// Cumulative up to February: branch B1 income 105.50 - expense 30.25, branch B2 income 200 - expense 260 (a loss).
	before = f.rowCounts()
	preview, err := f.previewClose("2026-02-28", "CLOSE-FEB")
	if err != nil {
		t.Fatal(err)
	}
	if after := f.rowCounts(); after != before {
		t.Fatalf("preview wrote rows: before %v after %v", before, after)
	}
	if preview.PendingDrafts || !preview.Balanced || preview.Scale != 2 || preview.Date != "2026-02-28" || len(preview.Branches) != 2 {
		t.Fatalf("preview header: %+v", preview)
	}
	sameDecimal(t, preview.TotalDebit, preview.TotalCredit)
	sameDecimal(t, preview.NetProfit, "15.25") // (105.50-30.25) + (200-260)
	b1, b2 := preview.Branches[0], preview.Branches[1]
	if b1.BranchCode != "B1" || b2.BranchCode != "B2" || b1.DocNo != "CLOSE-FEB-001" || b2.DocNo != "CLOSE-FEB-002" {
		t.Fatalf("branches: %+v %+v", b1, b2)
	}
	sameDecimal(t, b1.NetProfit, "75.25")
	sameDecimal(t, b2.NetProfit, "-60")
	for _, branch := range preview.Branches {
		sameDecimal(t, branch.TotalDebit, branch.TotalCredit)
		transfers := 0
		for _, line := range branch.Lines {
			if line.Transfer {
				transfers++
				if line.AccountCode != "3200" && line.AccountCode != "3100" {
					t.Fatalf("transfer line on %s: %+v", line.AccountCode, line)
				}
				if line.DepartmentCode == "" || line.ProjectCode == "" {
					t.Fatalf("transfer line lost its dimensions: %+v", line)
				}
			} else if line.AccountCode != "4000" && line.AccountCode != "5000" {
				t.Fatalf("closing line on %s: %+v", line.AccountCode, line)
			}
			if line.AccountName == "" {
				t.Fatalf("line without account name: %+v", line)
			}
		}
		if transfers != 3 {
			t.Fatalf("branch %s: %d transfer lines, want profit/loss in+out and retained earnings = 3", branch.BranchCode, transfers)
		}
	}
	if noNumber, err := f.previewClose("2026-02-28", ""); err != nil || noNumber.Branches[0].DocNo != "" || noNumber.Branches[1].DocNo != "" {
		t.Fatalf("preview without a number must not invent one: %+v %v", noNumber.Branches, err)
	}

	result := f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE-FEB", Date: "2026-02-28", Reason: "close"})
	drafts := f.closingDrafts()
	if result.CreatedJournals != 2 || len(drafts) != 2 {
		t.Fatalf("created %d journals, %d closing drafts", result.CreatedJournals, len(drafts))
	}
	for i, draft := range drafts {
		shown := preview.Branches[i]
		if draft.Status != "draft" || draft.DocNo != shown.DocNo || draft.BranchCode != shown.BranchCode || len(draft.Lines) != len(shown.Lines) {
			t.Fatalf("draft %s differs from the preview: %+v vs %+v", draft.DocNo, draft, shown)
		}
		for n, line := range draft.Lines {
			want := shown.Lines[n]
			if line.AccountCode != want.AccountCode || line.AccountName != want.AccountName || line.DepartmentCode != want.DepartmentCode || line.ProjectCode != want.ProjectCode {
				t.Fatalf("draft %s line %d: %+v vs preview %+v", draft.DocNo, n, line, want)
			}
			sameDecimal(t, line.Debit.Decimal().String(), want.Debit)
			sameDecimal(t, line.Credit.Decimal().String(), want.Credit)
		}
	}

	// After the closing vouchers are posted February has nothing left to close, for the preview and the command alike.
	for _, draft := range drafts {
		f.post(draft)
	}
	_, previewErr := f.previewClose("2026-02-28", "CLOSE-FEB2")
	_, closeErr := f.execute(f.scope, Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE-FEB2", Date: "2026-02-28", Reason: "close"})
	if previewErr == nil || closeErr == nil || previewErr.Error() != closeErr.Error() {
		t.Fatalf("preview error %v, close error %v", previewErr, closeErr)
	}

	// Closing up to April only carries the movement since the February close (B1 April income 70).
	april, err := f.previewClose("2026-04-30", "CLOSE-APR")
	if err != nil || len(april.Branches) != 1 || april.Branches[0].BranchCode != "B1" {
		t.Fatalf("april preview: %+v %v", april, err)
	}
	sameDecimal(t, april.NetProfit, "70")
	sameDecimal(t, april.TotalDebit, april.TotalCredit)
	aprilResult := f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE-APR", Date: "2026-04-30", Reason: "close"})
	aprilDraft := f.journal(aprilResult.ID)
	if len(aprilDraft.Lines) != len(april.Branches[0].Lines) {
		t.Fatalf("april draft lines %d, preview %d", len(aprilDraft.Lines), len(april.Branches[0].Lines))
	}

	// Branch-limited users can not preview a company-wide close; a stale year version is refused like the command.
	narrow := f.scope
	narrow.Branch = "B1"
	if _, err = f.store.Execute(f.ctx, narrow, Command{Resource: "processes", Action: ClosePreviewAction, ID: "2026", Version: year.Version, Date: "2026-04-30"}); err == nil {
		t.Fatal("branch-limited preview accepted")
	}
	if _, err = f.store.Execute(f.ctx, f.scope, Command{Resource: "processes", Action: ClosePreviewAction, ID: "2026", Version: year.Version + 7, Date: "2026-04-30"}); err == nil {
		t.Fatal("stale year version accepted")
	}
}

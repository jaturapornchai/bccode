//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func (f *pgIntegrityFixture) draftOn(doc, date string) Journal { return f.draftFrom(doc, date, "") }

// draftFrom creates a draft; a non-empty system makes it a voucher imported from that system (source type 2 + record id), like the import API does.
func (f *pgIntegrityFixture) draftFrom(doc, date, system string) Journal {
	f.t.Helper()
	j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: date[:4], Description: "ขายวัสดุก่อสร้าง รับชำระเงินสดหน้าร้าน", Kind: "manual", Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100")}}}
	if system != "" {
		j.SourceType, j.SourceSystem, j.SourceRecordID = 2, system, "SRC-"+doc
	}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	return f.journal(r.ID)
}

func (f *pgIntegrityFixture) livePeriods(year string) []Master {
	f.t.Helper()
	rows, err := f.db.Query(`SELECT payload || jsonb_build_object('id',id,'version',version) FROM gl_records WHERE company='C' AND kind='periods' AND payload->>'fiscalyear'=$1 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY payload->>'startdate'`, year)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	periods := []Master{}
	for rows.Next() {
		var raw []byte
		var p Master
		if err := rows.Scan(&raw); err != nil || json.Unmarshal(raw, &p) != nil {
			f.t.Fatalf("read period: %v", err)
		}
		periods = append(periods, p)
	}
	return periods
}

func (f *pgIntegrityFixture) period(code string) Master {
	f.t.Helper()
	var raw []byte
	if err := f.db.QueryRow(`SELECT payload || jsonb_build_object('id',id,'version',version) FROM gl_records WHERE company='C' AND kind='periods' AND code=$1 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, code).Scan(&raw); err != nil {
		f.t.Fatalf("period %s: %v", code, err)
	}
	var p Master
	if err := json.Unmarshal(raw, &p); err != nil {
		f.t.Fatal(err)
	}
	return p
}

func (f *pgIntegrityFixture) failWith(cmd Command, code string) *UserError {
	f.t.Helper()
	_, err := f.execute(f.scope, cmd)
	user, ok := AsUserError(err)
	if !ok || user.Code != code {
		f.t.Fatalf("%s/%s: want %s, got %v", cmd.Resource, cmd.Action, code, err)
	}
	return user
}

const testApprover = "นายสมชาย ใจดี"

// batchInput = the journals to post plus the approver Champ requires on every batch.
func batchInput(journals ...Journal) *PostBatchInput {
	return &PostBatchInput{Items: batchItems(journals...), Approver: testApprover}
}

func batchItems(journals ...Journal) []PostBatchItem {
	items := []PostBatchItem{}
	for _, j := range journals {
		items = append(items, PostBatchItem{ID: j.ID, Version: j.Version})
	}
	return items
}

func (f *pgIntegrityFixture) postCheck(filter PostBatchInput) PostCheck {
	f.t.Helper()
	r := f.run(Command{Resource: PostBatchKind, Action: "check", PostBatch: &filter})
	if r.PostCheck == nil {
		f.t.Fatal("check returned no rows object")
	}
	return *r.PostCheck
}

// Champ 5512: check shows every draft with the post-rule result; posting is all-or-nothing in one
// transaction, one audit event per journal plus the batch, and a retry returns the same batch.
func TestPostBatchCheckAllOrNothingAndReplay(t *testing.T) {
	f := newPGIntegrityFixture(t)
	a := f.draftOn("JV001", "2026-01-05")
	b := f.draftFrom("JV002", "2026-01-06", "pos-branch") // imported voucher: the Source column shows its system
	bad := f.draftOn("JV003", "2026-01-07")
	// Only a legacy/tampered row can be unbalanced — the save path refuses it.
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{lines,0,debit}','"150"') WHERE company='C' AND kind='journals' AND id=$1`, bad.ID); err != nil {
		t.Fatal(err)
	}
	f.draftOn("JV900", "2027-03-01") // outside the date filter

	check := f.postCheck(PostBatchInput{BookCode: "JV", DateFrom: "2026-01-01", DateTo: "2026-01-31"})
	if len(check.Rows) != 3 || check.More || check.Limit != PostBatchMaxJournals {
		t.Fatalf("want 3 rows in January, got %+v", check)
	}
	if !check.Rows[0].OK || !check.Rows[1].OK || check.Rows[0].DocNo != "JV001" || check.Rows[0].Debit != "100" || check.Rows[0].Version != a.Version || check.Rows[0].SourceSystem != "" || check.Rows[1].SourceSystem != "pos-branch" {
		t.Fatalf("good rows wrong: %+v", check.Rows[:2])
	}
	if check.Rows[2].OK || !strings.Contains(check.Rows[2].Message, "เดบิตและเครดิตไม่เท่ากัน") {
		t.Fatalf("unbalanced row not flagged: %+v", check.Rows[2])
	}
	if got := f.postCheck(PostBatchInput{DocNoFrom: "JV002", DocNoTo: "JV003"}); len(got.Rows) != 2 || got.Rows[0].DocNo != "JV002" {
		t.Fatalf("doc no range filter wrong: %+v", got.Rows)
	}
	if f.countRows(`SELECT count(*) FROM gl_lines`) != 0 || f.countRows(`SELECT count(*) FROM gl_records WHERE kind='journals' AND payload->>'status'<>'draft'`) != 0 {
		t.Fatal("check saved something")
	}

	events := f.countRows(`SELECT count(*) FROM gl_events`)
	user := f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: batchInput(a, b, bad)}, "post_batch_item_failed")
	if len(user.Args) != 2 || user.Args[0] != "JV003" {
		t.Fatalf("failure must name the journal: %+v", user)
	}
	if f.countRows(`SELECT count(*) FROM gl_events`) != events || f.countRows(`SELECT count(*) FROM gl_lines`) != 0 ||
		f.countRows(`SELECT count(*) FROM gl_records WHERE kind=$1`, PostBatchKind) != 0 || f.journal(a.ID).Status != "draft" {
		t.Fatal("a failed batch left something behind")
	}

	cmd := Command{Resource: PostBatchKind, Action: "create", RequestID: uuid.NewString(), Reason: "ผ่านรายการประจำเดือนมกราคม 2569", PostBatch: &PostBatchInput{Items: batchItems(a, b), Remark: "ผ่านรายการประจำเดือนมกราคม 2569", Check1: "  นางสาวสุดา มั่นคง ", Approver: " " + testApprover + " "}}
	r := f.run(cmd)
	if r.BatchNo != "000001" || r.PostedJournals != 2 || r.ID == "" || r.Version != 1 {
		t.Fatalf("batch result wrong: %+v", r)
	}
	for _, j := range []Journal{a, b} {
		if got := f.journal(j.ID); got.Status != "posted" || got.PostBatchNo != "000001" || got.PostedBy != "tester" {
			t.Fatalf("journal %s not posted in batch: %+v", j.DocNo, got)
		}
	}
	if f.countRows(`SELECT count(*) FROM gl_lines`) != 4 || f.countRows(`SELECT count(*) FROM gl_events WHERE payload->>'action'='journals:post'`) != 2 ||
		f.countRows(`SELECT count(*) FROM gl_events WHERE payload->>'action'='post-batches:create'`) != 1 {
		t.Fatal("batch must write each journal's post event, its lines and one batch event")
	}
	var raw []byte
	if err := f.db.QueryRow(`SELECT payload FROM gl_records WHERE company='C' AND kind=$1 AND code='000001'`, PostBatchKind).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var batch PostBatch
	if err := json.Unmarshal(raw, &batch); err != nil || batch.Count != 2 || len(batch.Journals) != 2 || batch.Journals[0].DocNo != "JV001" || batch.PostedBy != "tester" || batch.Remark != "ผ่านรายการประจำเดือนมกราคม 2569" ||
		batch.Check1 != "นางสาวสุดา มั่นคง" || batch.Check2 != "" || batch.Approver != testApprover {
		t.Fatalf("batch record wrong: %v %+v", err, batch)
	}

	replay := f.run(cmd)
	if replay.BatchNo != "000001" || replay.PostedJournals != 2 || replay.Sequence != r.Sequence || f.countRows(`SELECT count(*) FROM gl_records WHERE kind=$1`, PostBatchKind) != 1 {
		t.Fatalf("retry must return the saved batch: %+v vs %+v", replay, r)
	}
	if rest := f.postCheck(PostBatchInput{DateTo: "2026-12-31"}); len(rest.Rows) != 1 || rest.Rows[0].DocNo != "JV003" {
		t.Fatalf("posted journals must leave the list: %+v", rest.Rows)
	}
	f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: batchInput(a)}, "post_batch_item_failed") // already posted
	next := f.draftOn("JV004", "2026-01-08")
	if r2 := f.run(Command{Resource: PostBatchKind, Action: "create", PostBatch: batchInput(next)}); r2.BatchNo != "000002" {
		t.Fatalf("batch no must run: %+v", r2)
	}
	f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: &PostBatchInput{}}, "post_batch_empty")
	f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: &PostBatchInput{Items: []PostBatchItem{{ID: bad.ID, Version: bad.Version}, {ID: bad.ID, Version: bad.Version}}, Approver: testApprover}}, "post_batch_item_invalid")
	f.deny(f.scope, Command{Resource: PostBatchKind, Action: "delete", ID: r.ID, Version: r.Version})
}

// Champ GLDlgPosttransDetail: the approver is required (blank = refused, nothing posted), the reviewers
// are optional, and each name fits BCPostTrans VARCHAR(50) counted in characters, not bytes.
func TestPostBatchSignatories(t *testing.T) {
	f := newPGIntegrityFixture(t)
	a := f.draftOn("JV201", "2026-01-05")
	items := batchItems(a)
	events := f.countRows(`SELECT count(*) FROM gl_events`)
	for _, blank := range []string{"", "   "} {
		user := f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: &PostBatchInput{Items: items, Approver: blank, Check1: "นางสาวสุดา มั่นคง"}}, "post_batch_approver_required")
		if user.Field != "approver" {
			t.Fatalf("approver error must point at the field: %+v", user)
		}
	}
	long := strings.Repeat("ก", PostBatchPersonMaxRunes+1) // 51 characters, 153 bytes
	for field, input := range map[string]PostBatchInput{
		"check1":   {Items: items, Approver: testApprover, Check1: long},
		"check2":   {Items: items, Approver: testApprover, Check2: long},
		"approver": {Items: items, Approver: long},
	} {
		input := input
		if user := f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: &input}, "post_batch_person_too_long"); user.Field != field || len(user.Args) != 1 || user.Args[0] != "50" {
			t.Fatalf("%s: wrong error %+v", field, user)
		}
	}
	if f.countRows(`SELECT count(*) FROM gl_events`) != events || f.journal(a.ID).Status != "draft" || f.countRows(`SELECT count(*) FROM gl_records WHERE kind=$1`, PostBatchKind) != 0 {
		t.Fatal("a refused batch left something behind")
	}

	fifty := strings.Repeat("ก", PostBatchPersonMaxRunes) // exactly 50 characters is allowed
	r := f.run(Command{Resource: PostBatchKind, Action: "create", PostBatch: &PostBatchInput{Items: items, Approver: fifty, Check1: fifty, Check2: " นายวิชัย รักดี "}})
	var raw []byte
	if err := f.db.QueryRow(`SELECT payload FROM gl_records WHERE company='C' AND kind=$1 AND code=$2`, PostBatchKind, r.BatchNo).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var batch PostBatch
	if err := json.Unmarshal(raw, &batch); err != nil || batch.Approver != fifty || batch.Check1 != fifty || batch.Check2 != "นายวิชัย รักดี" {
		t.Fatalf("batch signatories wrong: %v %+v", err, batch)
	}
}

func TestPostBatchRejectsLockedPeriodAndStaleVersion(t *testing.T) {
	f := newPGIntegrityFixture(t)
	jan := f.draftOn("JV101", "2026-01-10")
	feb := f.draftOn("JV102", "2026-02-10")
	// A draft inside a locked period is a legacy state (locking refuses periods with drafts).
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{locked}','true') WHERE company='C' AND kind='periods' AND code='2026-02'`); err != nil {
		t.Fatal(err)
	}
	check := f.postCheck(PostBatchInput{})
	if len(check.Rows) != 2 || !check.Rows[0].OK || check.Rows[1].OK || !strings.Contains(check.Rows[1].Message, "ถูกล็อก") {
		t.Fatalf("locked-period row not flagged: %+v", check.Rows)
	}
	user := f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: batchInput(jan, feb)}, "post_batch_item_failed")
	if user.Args[0] != "JV102" || f.journal(jan.ID).Status != "draft" {
		t.Fatalf("locked period must stop the whole batch: %+v", user)
	}
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{locked}','false') WHERE company='C' AND kind='periods' AND code='2026-02'`); err != nil {
		t.Fatal(err)
	}

	edited := f.journal(jan.ID)
	edited.Description = "ขายวัสดุก่อสร้าง ให้ลูกค้าโครงการก่อสร้าง A เป็นเงินเชื่อ"
	f.run(Command{Resource: "journals", Action: "update", ID: edited.ID, Version: edited.Version, Journal: &edited})
	user = f.failWith(Command{Resource: PostBatchKind, Action: "create", PostBatch: batchInput(feb, jan)}, "post_batch_item_failed")
	if user.Args[0] != "JV101" || f.journal(feb.ID).Status != "draft" || f.countRows(`SELECT count(*) FROM gl_lines`) != 0 {
		t.Fatalf("stale version must stop the whole batch: %+v", user)
	}
	if r := f.run(Command{Resource: PostBatchKind, Action: "create", PostBatch: batchInput(feb, f.journal(jan.ID))}); r.PostedJournals != 2 {
		t.Fatalf("fresh versions must post: %+v", r)
	}
}

// Champ 5511 F9: a fiscal year comes with its monthly periods; "generate" fills only missing months
// and refuses a period that would overlap; periods follow the year's dates until something uses it.
func TestFiscalYearMonthlyPeriods(t *testing.T) {
	f := newPGIntegrityFixture(t)
	periods := f.livePeriods("2026")
	if len(periods) != 12 || periods[0].Code != "2026-01" || periods[11].EndDate != "2026-12-31" || periods[1].EndDate != "2026-02-28" {
		t.Fatalf("new year must come with 12 monthly periods: %+v", periods)
	}
	f.failWith(Command{Resource: "periods", Action: "generate", TargetYear: "2026"}, "periods_complete")
	f.failWith(Command{Resource: "periods", Action: "generate"}, "period_year_required")

	march := f.period("2026-03")
	f.run(Command{Resource: "periods", Action: "delete", ID: march.ID, Version: march.Version, Reason: "ทดสอบสร้างใหม่"})
	if r := f.run(Command{Resource: "periods", Action: "generate", TargetYear: "2026"}); r.CreatedPeriods != 1 {
		t.Fatalf("generate must add only March: %+v", r)
	}
	if again := f.period("2026-03"); again.StartDate != "2026-03-01" || again.EndDate != "2026-03-31" || again.ID == march.ID {
		t.Fatalf("March not recreated: %+v", again)
	}

	april := f.period("2026-04")
	f.run(Command{Resource: "periods", Action: "delete", ID: april.ID, Version: april.Version, Reason: "ทดสอบช่วงคร่อม"})
	half := Master{Code: "APR-H1", FiscalYear: "2026", StartDate: "2026-04-01", EndDate: "2026-04-15", IsActive: true}
	f.run(Command{Resource: "periods", Action: "create", Master: &half})
	f.failWith(Command{Resource: "periods", Action: "generate", TargetYear: "2026"}, "period_generate_overlap")
	if n := len(f.livePeriods("2026")); n != 12 {
		t.Fatalf("refused generate must not add periods: %d", n)
	}

	y := FiscalYear{Code: "2028", StartDate: "2028-01-01", EndDate: "2028-06-30", IsActive: true, Scale: 2, ProfitLossAccount: "3200", RetainedEarningsAccount: "3100"}
	r := f.run(Command{Resource: "fiscal-years", Action: "create", FiscalYear: &y})
	if n := len(f.livePeriods("2028")); n != 6 {
		t.Fatalf("half year must have 6 periods, got %d", n)
	}
	y.ID, y.Version, y.EndDate = r.ID, r.Version, "2028-12-31"
	r = f.run(Command{Resource: "fiscal-years", Action: "update", ID: y.ID, Version: y.Version, FiscalYear: &y})
	if p := f.livePeriods("2028"); len(p) != 12 || p[11].EndDate != "2028-12-31" || f.countRows(`SELECT count(*) FROM gl_records WHERE kind='periods' AND code='2028-01' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`) != 1 {
		t.Fatalf("periods must follow the new year end: %+v", p)
	}
	y.Version = r.Version
	jan := f.period("2028-01")
	f.run(Command{Resource: "periods", Action: "lock", ID: jan.ID, Version: jan.Version, Reason: "ปิดงวดมกราคม"})
	moved := y
	moved.EndDate = "2028-11-30"
	f.deny(f.scope, Command{Resource: "fiscal-years", Action: "update", ID: y.ID, Version: y.Version, FiscalYear: &moved})
	jan = f.period("2028-01")
	f.run(Command{Resource: "periods", Action: "unlock", ID: jan.ID, Version: jan.Version, Reason: "แก้ไขปีบัญชี"})
	f.run(Command{Resource: "fiscal-years", Action: "delete", ID: y.ID, Version: y.Version})
	if n := len(f.livePeriods("2028")); n != 0 {
		t.Fatalf("deleting the year must delete its periods, %d left", n)
	}
}

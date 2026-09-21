//go:build integration

package generalledger

import (
	"context"
	"testing"
)

func TestJournalReviewLifecycle(t *testing.T) {
	pg, db := glpgTestDB(t)
	ctx := context.Background()
	if _, err := pg.database(ctx, "H"); err != nil {
		t.Fatal(err)
	}
	_, err := db.Exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES ('C','journals','j','JV1',1,'{"holdingcode":"H","businesscode":"C","id":"j","version":1,"docno":"JV1","status":"draft","branchcode":"B","lines":[]}')`)
	if err != nil {
		t.Fatal(err)
	}
	store := NewPostgresStore(pg)
	scope := Scope{Holding: "H", Company: "C", Actor: "reviewer", Branch: "B"}
	zero := int64(0)
	cmd := Command{Resource: "journals", Action: "review", ID: "j", Version: 1, RequestID: "review-request-0001", Review: &ReviewInput{Status: 2, Note: "missing invoice", ExpectedEventNo: &zero}}
	result, err := store.Execute(ctx, scope, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if result.Version != 1 {
		t.Fatal("review changed journal version")
	}
	review, err := pg.JournalReview(ctx, scope, "j")
	if err != nil || review.Status != 2 || len(review.Events) != 1 {
		t.Fatalf("review=%+v err=%v", review, err)
	}
	if _, err = store.Execute(ctx, scope, cmd); err != nil {
		t.Fatal(err)
	}
	changed := cmd
	changed.Review = &ReviewInput{Status: 3, Note: "different", ExpectedEventNo: &zero}
	if _, err = store.Execute(ctx, scope, changed); err == nil {
		t.Fatal("accepted idempotency payload change")
	}
	latest := review.EventNo
	outcomes := make(chan error, 2)
	for _, requestID := range []string{"review-concurrent-01", "review-concurrent-02"} {
		next := cmd
		next.RequestID = requestID
		next.Review = &ReviewInput{Status: 3, ExpectedEventNo: &latest}
		go func(c Command) { _, e := store.Execute(ctx, scope, c); outcomes <- e }(next)
	}
	successes := 0
	for range 2 {
		if <-outcomes == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent successes=%d", successes)
	}
	cmd.RequestID = "review-request-0002"
	if _, err = store.Execute(ctx, scope, cmd); err == nil {
		t.Fatal("accepted stale review")
	}
	if _, err = db.Exec(`UPDATE gl_journal_review_events SET note='tamper'`); err == nil {
		t.Fatal("audit mutated")
	}
	if _, err = db.Exec(`UPDATE gl_records SET version=2 WHERE company='C' AND id='j'`); err != nil {
		t.Fatal(err)
	}
	review, err = pg.JournalReview(ctx, scope, "j")
	if err != nil || review.Status != 1 || review.EventNo != 0 || len(review.Events) != 2 {
		t.Fatalf("stale review reused %+v %v", review, err)
	}
	if _, err = store.Execute(ctx, scope, cmd); err == nil {
		t.Fatal("accepted stale journal")
	}
	denied := scope
	denied.Company = "other"
	if _, err = pg.JournalReview(ctx, denied, "j"); err == nil {
		t.Fatal("tenant leak")
	}
	denied = scope
	denied.Branch = "other"
	if _, err = pg.JournalReview(ctx, denied, "j"); err == nil {
		t.Fatal("branch leak")
	}
}

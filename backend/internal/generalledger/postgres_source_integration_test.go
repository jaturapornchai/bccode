//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"sync"
	"testing"
)

func TestPostgresSourceIdentitySurvivesNewRequestIDs(t *testing.T) {
	pg, db := glpgTestDB(t)
	glpgSeed(t, pg, "A")
	glpgSeed(t, pg, "B")
	ctx := context.Background()
	store := NewPostgresStore(pg)
	scope := Scope{Holding: "H", Company: "A", Actor: "source-test"}
	journal := glpgJournal("A", "IMPORT-1", "2026-01-10", "manual", "draft", "1000", "4000", "0.30")
	journal.SourceType = 2
	journal.SourceSystem = "external-ledger"
	journal.SourceRecordID = "stable-record-1"
	command := Command{Resource: "journals", Action: "create", RequestID: uuid.NewString(), Journal: &journal}
	first, err := store.Execute(ctx, scope, command)
	if err != nil {
		t.Fatal(err)
	}
	const workers = 4
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			retry := command
			retry.RequestID = uuid.NewString()
			got, e := store.Execute(ctx, scope, retry)
			if e != nil || got.ID != first.ID || got.Sequence != first.Sequence {
				t.Errorf("source retry got=%+v err=%v", got, e)
			}
		}()
	}
	wg.Wait()
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM gl_records WHERE company='A' AND kind='journals'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicate journals=%d err=%v", count, err)
	}
	raw, err := store.Get(ctx, scope, "journals", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	var persisted Journal
	if err = json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.SourceRecordID != journal.SourceRecordID {
		t.Fatal("source round trip changed")
	}
	glpgDecimalEqual(t, string(persisted.Lines[0].Debit), "0.30")
	changed := journal
	changed.Description = "changed payload"
	command.Journal = &changed
	command.RequestID = uuid.NewString()
	if _, err = store.Execute(ctx, scope, command); err == nil {
		t.Fatal("accepted changed payload for existing source")
	}
	changed = journal
	changed.SourceRecordID = "other-record"
	command.Journal = &changed
	command.Action = "update"
	command.ID = first.ID
	command.Version = first.Version
	command.RequestID = uuid.NewString()
	if _, err = store.Execute(ctx, scope, command); err == nil {
		t.Fatal("accepted source identity change")
	}
	command.Journal = &journal
	command.Action = "create"
	command.ID = ""
	command.RequestID = uuid.NewString()
	other := scope
	other.Company = "B"
	if second, err := store.Execute(ctx, other, command); err != nil || second.ID == first.ID {
		t.Fatalf("company isolation result=%+v err=%v", second, err)
	}
}

func TestJournalSourceValidation(t *testing.T) {
	for _, journal := range []Journal{{SourceType: 7}, {SourceType: 2}, {SourceSystem: "erp"}, {SourceRecordID: "one"}, {SourceSystem: " erp", SourceRecordID: "one"}} {
		if err := validateJournalSource(journal); err == nil {
			t.Errorf("accepted malformed source %+v", journal)
		}
	}
	if err := validateJournalSource(Journal{}); err != nil {
		t.Fatal("manual compatibility", err)
	}
}

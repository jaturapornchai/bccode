//go:build integration

package generalledger

import "testing"

// Champ runs a voucher number when the user leaves it blank (BRunningEdit) and keeps a typed one.
func TestPostgresJournalAutoDocNo(t *testing.T) {
	f := newPGIntegrityFixture(t)
	first := f.draft("", "B1")
	second := f.draft("", "B1")
	if first.DocNo != "JV6901-0001" || second.DocNo != "JV6901-0002" {
		t.Fatalf("auto numbers = %q, %q; want JV6901-0001, JV6901-0002", first.DocNo, second.DocNo)
	}
	if typed := f.draft("JV-MANUAL-9", "B1"); typed.DocNo != "JV-MANUAL-9" {
		t.Fatalf("typed docno replaced: %q", typed.DocNo)
	}
	preview, err := f.store.PreviewJournalDocNo(f.ctx, f.scope, "JV", "2026-01-31")
	if err != nil || preview != "JV6901-0003" {
		t.Fatalf("preview = %q, %v; want JV6901-0003", preview, err)
	}
	// A deleted draft keeps its number: the next voucher must not reuse it.
	f.run(Command{Resource: "journals", Action: "delete", ID: second.ID, Version: second.Version})
	if third := f.draft("", "B1"); third.DocNo != "JV6901-0003" {
		t.Fatalf("after delete got %q, want JV6901-0003", third.DocNo)
	}

	books := f.books()
	book := books["JV"]
	book.DocFormat = "YYMM-@#####"
	book.TitleName = "ใบสำคัญทั่วไป"
	f.run(Command{Resource: "journal-books", Action: "update", ID: book.ID, Version: book.Version, Master: &book})
	if got := f.draft("", "B1"); got.DocNo != "6901-JV00001" {
		t.Fatalf("custom format docno = %q, want 6901-JV00001", got.DocNo)
	}
	if stored := f.books()["JV"]; stored.TitleName != "ใบสำคัญทั่วไป" || stored.DocFormat != "YYMM-@#####" {
		t.Fatalf("book format/title not stored: %+v", stored)
	}
	book = f.books()["JV"]
	bad := book
	bad.DocFormat = "JV-YYMM"
	f.failCode(Command{Resource: "journal-books", Action: "update", ID: book.ID, Version: book.Version, Master: &bad}, "journal_book_docformat_invalid", "docformat")
}

package generalledger

import "testing"

func TestChampParityMasterCollections(t *testing.T) {
	if MasterCollections["journal-books"] != "gl_journal_books" {
		t.Fatalf("expected MasterCollections[journal-books] to be gl_journal_books, got %s", MasterCollections["journal-books"])
	}
}

package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

func TestCheckJournalBranch(t *testing.T) {
	ctx := context.Background()
	found := scopeDB(t, &scopeDriver{})
	missing := scopeDB(t, &scopeDriver{branchErr: sql.ErrNoRows})
	broken := scopeDB(t, &scopeDriver{branchErr: errors.New("db down")})
	connect := func(db *sql.DB) func(string) (*sql.DB, error) {
		return func(string) (*sql.DB, error) { return db, nil }
	}
	branchScope := requestScope{Scope: gl.Scope{Holding: "H", Company: "C", Branch: "B"}}
	companyScope := requestScope{Scope: gl.Scope{Holding: "H", Company: "C"}}
	cases := []struct {
		name     string
		scope    requestScope
		connect  func(string) (*sql.DB, error)
		branch   string
		wantCode string
		wantErr  bool
	}{
		{name: "branch session blank uses session", scope: branchScope, branch: ""},
		{name: "branch session same branch", scope: branchScope, branch: " B "},
		{name: "branch session other branch", scope: branchScope, branch: "X", wantCode: "journal_branch_outside_session"},
		{name: "company session requires branch", scope: companyScope, connect: connect(found), branch: "", wantCode: "journal_branch_required"},
		{name: "company session active branch", scope: companyScope, connect: connect(found), branch: "B"},
		{name: "company session unknown branch", scope: companyScope, connect: connect(missing), branch: "สาขาปลอม", wantCode: "journal_branch_not_found"},
		{name: "company session lookup failure fails closed", scope: companyScope, connect: connect(broken), branch: "B", wantErr: true},
		{name: "company session no registry fails closed", scope: companyScope, branch: "B", wantErr: true},
	}
	for _, c := range cases {
		err := checkJournalBranch(ctx, c.connect, c.scope, &gl.Journal{BranchCode: c.branch})
		if c.wantCode == "" && !c.wantErr {
			if err != nil {
				t.Fatalf("%s: unexpected error %v", c.name, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("%s: expected error", c.name)
		}
		if c.wantCode != "" {
			user, ok := gl.AsUserError(err)
			if !ok || user.Code != c.wantCode || user.Field != "branchcode" || user.Message == "" {
				t.Fatalf("%s: got %#v want code %s on field branchcode", c.name, err, c.wantCode)
			}
		}
	}
}

// CheckJournalBranch is the same rule for journals other modules (fixed assets) post.
func TestCheckJournalBranchExported(t *testing.T) {
	ctx := context.Background()
	session := gl.Scope{Holding: "H", Company: "C", Branch: "B"}
	if user, ok := gl.AsUserError(CheckJournalBranch(ctx, nil, session, "X")); !ok || user.Code != "journal_branch_outside_session" {
		t.Fatalf("other branch in a branch session = %#v", user)
	}
	if err := CheckJournalBranch(ctx, nil, session, ""); err != nil {
		t.Fatalf("blank must mean the session branch: %v", err)
	}
	if err := CheckJournalBranch(ctx, nil, gl.Scope{Holding: "H", Company: "C"}, "B"); err == nil {
		t.Fatal("a company session without the branch registry must fail closed")
	}
}

// A budget branch is optional (blank = every branch); a named branch in a company-wide
// session must be an active branch of the company, and lookups fail closed.
func TestCheckBudgetBranch(t *testing.T) {
	ctx := context.Background()
	connect := func(db *sql.DB) func(string) (*sql.DB, error) {
		return func(string) (*sql.DB, error) { return db, nil }
	}
	found := connect(scopeDB(t, &scopeDriver{}))
	missing := connect(scopeDB(t, &scopeDriver{branchErr: sql.ErrNoRows}))
	branchScope := requestScope{Scope: gl.Scope{Holding: "H", Company: "C", Branch: "B"}}
	companyScope := requestScope{Scope: gl.Scope{Holding: "H", Company: "C"}}
	for _, c := range []struct {
		name     string
		scope    requestScope
		connect  func(string) (*sql.DB, error)
		branch   string
		wantCode string
		wantErr  bool
	}{
		{name: "company session blank = every branch", scope: companyScope, branch: ""},
		{name: "company session active branch", scope: companyScope, connect: found, branch: " B "},
		{name: "company session unknown branch", scope: companyScope, connect: missing, branch: "สาขาปลอม", wantCode: "journal_branch_not_found"},
		{name: "company session no registry fails closed", scope: companyScope, branch: "B", wantErr: true},
		{name: "branch session left to the store", scope: branchScope, branch: "X"},
	} {
		err := checkBudgetBranch(ctx, c.connect, c.scope, &gl.Budget{BranchCode: c.branch})
		if c.wantCode == "" && !c.wantErr {
			if err != nil {
				t.Fatalf("%s: unexpected error %v", c.name, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("%s: expected error", c.name)
		}
		if user, ok := gl.AsUserError(err); c.wantCode != "" && (!ok || user.Code != c.wantCode || user.Field != "branchcode") {
			t.Fatalf("%s: got %#v want %s", c.name, err, c.wantCode)
		}
	}
}

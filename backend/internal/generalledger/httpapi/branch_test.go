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

package httpapi

import (
	"context"
	"database/sql"
	"errors"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/mcptoken"
)

// CheckJournalBranch applies the same header-branch rule to a journal another module (fixed
// assets) posts through the ledger for this session; branch is the voucher header branch and
// connect opens the control database that holds the branch registry.
func CheckJournalBranch(ctx context.Context, connect func(string) (*sql.DB, error), session gl.Scope, branch string) error {
	return checkJournalBranch(ctx, connect, requestScope{Scope: session}, &gl.Journal{BranchCode: branch})
}

// checkJournalBranch validates the voucher header branch against the organisation registry
// (audit 2026-09-24: a company-wide session saved blank or made-up branches).
//   - branch-scoped session: blank means the session branch (the store fills it in); any
//     other branch is refused with a message naming both branches.
//   - company-wide session: the branch is required and must be an active branch of the company.
func checkJournalBranch(ctx context.Context, connect func(string) (*sql.DB, error), scope requestScope, j *gl.Journal) error {
	branch := gl.NormalizeCode(j.BranchCode)
	session := scope.Scope.Branch
	if session != "" {
		if branch == "" || branch == session {
			return nil // resolveScope already verified the session branch is active
		}
		return gl.JournalBranchOutsideSession(branch, session)
	}
	if branch == "" {
		return gl.JournalBranchRequired()
	}
	if connect == nil {
		return errScopeDenied
	}
	db, err := connect(mcptoken.ControlDatabase)
	if err != nil || db == nil {
		return errScopeDenied
	}
	var code string
	err = db.QueryRowContext(ctx, `SELECT code FROM branches WHERE holding_code=$1 AND company_code=$2 AND code=$3 AND is_active=true`, scope.Scope.Holding, scope.Scope.Company, branch).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return gl.JournalBranchNotFound(branch)
	}
	return err
}

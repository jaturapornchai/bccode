package shop

import (
	"context"
	"strings"

	"smlcloudplatform/internal/authentication/models"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"smlcloudplatform/pkg/apperr"
)

// accessScopeError rejects a membership save; key names the languages.tsv row the
// handler shows in the caller's language (shopUserErrorRow).
type accessScopeError struct {
	status *apperr.AppError
	key    string
}

var (
	errAccessScopeInvalid        = &accessScopeError{status: apperr.ErrValidation, key: "ss_err_access_scope_invalid"}
	errAccessScopeExceedsGrantor = &accessScopeError{status: apperr.ErrForbidden, key: "ss_err_access_scope_exceeds_grantor"}
)

func (e *accessScopeError) Error() string { return e.key }

func accessScopeType(scope models.AccessScope) string {
	return strings.ToLower(strings.TrimSpace(scope.ScopeType))
}

// hydrateAccessScopes validates saved scope rules against this Holding and resolves the
// identities scope enforcement reads (company UID = company code, branch UID = branch
// code). A holding rule collapses the list to that single rule: readers expand it at
// read time, so companies created later are covered. An unknown scope type or a
// company/branch outside this Holding rejects the save; nothing is promoted to holding.
func hydrateAccessScopes(ctx context.Context, db dbtx, holdingCode string, scopes []models.AccessScope) ([]models.AccessScope, error) {
	needBranches := false
	for _, scope := range scopes {
		switch accessScopeType(scope) {
		case "holding", "company":
		case "branch":
			needBranches = true
		default:
			return nil, errAccessScopeInvalid
		}
	}
	if models.HasHoldingScope(scopes) {
		return []models.AccessScope{{ScopeType: "holding"}}, nil
	}
	companies, branches, err := holdingOrganization(ctx, db, holdingCode, needBranches)
	if err != nil {
		return nil, err
	}
	hydrated := make([]models.AccessScope, 0, len(scopes))
	for _, scope := range scopes {
		companyRef := scope.CompanyUID
		if strings.TrimSpace(companyRef) == "" {
			companyRef = scope.BusinessCode
		}
		companyUID, ok := companies[companyModels.NormalizeCompanyCode(companyRef)]
		if !ok {
			return nil, errAccessScopeInvalid
		}
		scope.ScopeType = accessScopeType(scope)
		scope.CompanyUID, scope.BusinessCode = companyUID, companyUID
		if scope.ScopeType == "company" {
			// A company rule names no branch. A leftover branch id is unvalidated and would turn
			// it into a rule ScopesAllowCompanySelection no longer treats as company-wide.
			scope.BranchUID, scope.BranchCode = "", ""
		}
		if scope.ScopeType == "branch" {
			branchRef := scope.BranchUID
			if strings.TrimSpace(branchRef) == "" {
				branchRef = scope.BranchCode
			}
			branchUID, ok := resolveBranchCode(branches[companyUID], branchRef)
			if !ok {
				return nil, errAccessScopeInvalid
			}
			scope.BranchUID, scope.BranchCode = branchUID, branchUID
		}
		hydrated = append(hydrated, scope)
	}
	return hydrated, nil
}

// holdingOrganization loads the Holding's company codes (keyed by normalized code) and,
// when asked, its branch codes per company. Closed companies/branches may still be
// granted: every reader opens active ones only.
func holdingOrganization(ctx context.Context, db dbtx, holdingCode string, withBranches bool) (map[string]string, map[string]map[string]bool, error) {
	companies := map[string]string{}
	rows, err := db.QueryContext(ctx, `SELECT code FROM companies WHERE holding_code = $1`, holdingCode)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, nil, err
		}
		companies[companyModels.NormalizeCompanyCode(code)] = code
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	branches := map[string]map[string]bool{}
	if !withBranches {
		return companies, branches, nil
	}
	branchRows, err := db.QueryContext(ctx, `SELECT company_code, code FROM branches WHERE holding_code = $1`, holdingCode)
	if err != nil {
		return nil, nil, err
	}
	defer branchRows.Close()
	for branchRows.Next() {
		var companyCode, code string
		if err := branchRows.Scan(&companyCode, &code); err != nil {
			return nil, nil, err
		}
		if branches[companyCode] == nil {
			branches[companyCode] = map[string]bool{}
		}
		branches[companyCode][code] = true
	}
	return companies, branches, branchRows.Err()
}

// resolveBranchCode matches a branch identifier exactly, then as a Thai tax branch code
// ("1" → "00001").
func resolveBranchCode(codes map[string]bool, value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value != "" && codes[value] {
		return value, true
	}
	if normalized, err := branchModels.NormalizeThaiTaxBranchCode(value); err == nil && codes[normalized] {
		return normalized, true
	}
	return "", false
}

package shop

import (
	"context"
	"strings"

	"smlcloudplatform/internal/authentication/models"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	companyModels "smlcloudplatform/internal/organization/company/models"
)

// hydrateAccessScopes resolves saved scope rules to the identities scope enforcement
// reads (company UID = company code, branch UID = branch code) and expands a
// holding-wide scope into one scope per company (allow-list semantics).
func hydrateAccessScopes(ctx context.Context, db dbtx, holdingCode string, scopes []models.AccessScope) ([]models.AccessScope, error) {
	expanded := make([]models.AccessScope, 0, len(scopes))
	for _, scope := range scopes {
		switch strings.ToLower(strings.TrimSpace(scope.ScopeType)) {
		case "holding":
			companies, err := allCompanyScopes(ctx, db, holdingCode)
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, companies...)
		case "branch":
			if scope.CompanyUID == "" {
				scope.CompanyUID = companyModels.NormalizeCompanyCode(scope.BusinessCode)
			}
			if scope.BranchUID == "" {
				scope.BranchUID = strings.TrimSpace(scope.BranchCode)
				if code, err := branchModels.NormalizeThaiTaxBranchCode(scope.BranchCode); err == nil {
					scope.BranchUID = code
				}
			}
			expanded = append(expanded, scope)
		default:
			if scope.CompanyUID == "" {
				scope.CompanyUID = companyModels.NormalizeCompanyCode(scope.BusinessCode)
			}
			expanded = append(expanded, scope)
		}
	}
	return expanded, nil
}

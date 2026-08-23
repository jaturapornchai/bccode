package access

import (
	"sort"
	"strings"

	authmodels "smlcloudplatform/internal/authentication/models"
)

func AllowedCompanyUIDs(scopes []authmodels.AccessScope) []string {
	unique := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		scopeType := strings.ToLower(strings.TrimSpace(scope.ScopeType))
		companyUID := normalizeUID(scope.CompanyUID)
		validBranch := scopeType == "branch" && normalizeUID(scope.BranchUID) != ""
		if companyUID != "" && (scopeType == "company" || validBranch) {
			unique[companyUID] = struct{}{}
		}
	}

	codes := make([]string, 0, len(unique))
	for code := range unique {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

func AllowsCompany(scopes []authmodels.AccessScope, companyUID string) bool {
	companyUID = normalizeUID(companyUID)
	if companyUID == "" {
		return false
	}
	for _, allowed := range AllowedCompanyUIDs(scopes) {
		if allowed == companyUID {
			return true
		}
	}
	return false
}

func AllowsAllBranches(scopes []authmodels.AccessScope, companyUID string) bool {
	companyUID = normalizeUID(companyUID)
	if companyUID == "" {
		return false
	}
	for _, scope := range scopes {
		if strings.EqualFold(strings.TrimSpace(scope.ScopeType), "company") && scope.AllBranches && normalizeUID(scope.CompanyUID) == companyUID {
			return true
		}
	}
	return false
}

func AllowedBranchUIDs(scopes []authmodels.AccessScope, companyUID string) []string {
	companyUID = normalizeUID(companyUID)
	unique := make(map[string]struct{})
	for _, scope := range scopes {
		if !strings.EqualFold(strings.TrimSpace(scope.ScopeType), "branch") || normalizeUID(scope.CompanyUID) != companyUID {
			continue
		}
		if branchUID := normalizeUID(scope.BranchUID); branchUID != "" {
			unique[branchUID] = struct{}{}
		}
	}

	codes := make([]string, 0, len(unique))
	for code := range unique {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

func AllowsBranch(scopes []authmodels.AccessScope, companyUID string, branchUID string) bool {
	companyUID = normalizeUID(companyUID)
	branchUID = normalizeUID(branchUID)
	if companyUID == "" || branchUID == "" {
		return false
	}
	if AllowsAllBranches(scopes, companyUID) {
		return true
	}
	for _, allowed := range AllowedBranchUIDs(scopes, companyUID) {
		if allowed == branchUID {
			return true
		}
	}
	return false
}

func normalizeUID(value string) string {
	return strings.TrimSpace(value)
}

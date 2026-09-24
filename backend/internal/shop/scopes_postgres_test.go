package shop

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/authentication/models"
)

func holdingOrgMock(t *testing.T, withBranches bool) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectQuery("SELECT code FROM companies WHERE holding_code").WithArgs("rungrueng").
		WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("01").AddRow("C02"))
	if withBranches {
		mock.ExpectQuery("SELECT company_code, code FROM branches WHERE holding_code").WithArgs("rungrueng").
			WillReturnRows(sqlmock.NewRows([]string{"company_code", "code"}).AddRow("01", "00000").AddRow("01", "00001"))
	}
	return db, mock
}

func TestHydrateAccessScopesKeepsHoldingRule(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, err := hydrateAccessScopes(context.Background(), db, "rungrueng", []models.AccessScope{
		{ScopeType: "company", BusinessCode: "01", AllBranches: true},
		{ScopeType: " Holding ", CompanyUID: "C02"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []models.AccessScope{{ScopeType: "holding"}}) {
		t.Fatalf("holding rule must be stored alone, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHydrateAccessScopesRejectsUnknownType(t *testing.T) {
	for _, scope := range []models.AccessScope{{}, {ScopeType: "everything"}, {ScopeType: "holdings", BusinessCode: "01"}} {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		// A holding rule next to an invalid one must not hide it.
		_, err = hydrateAccessScopes(context.Background(), db, "rungrueng", []models.AccessScope{{ScopeType: "holding"}, scope})
		db.Close()
		if !errors.Is(err, errAccessScopeInvalid) {
			t.Fatalf("scope %+v: err = %v, want errAccessScopeInvalid", scope, err)
		}
	}
}

func TestHydrateAccessScopesResolvesCompanyAndBranch(t *testing.T) {
	db, mock := holdingOrgMock(t, true)
	got, err := hydrateAccessScopes(context.Background(), db, "rungrueng", []models.AccessScope{
		{ScopeType: "Company", BusinessCode: " c02 ", AllBranches: true},
		{ScopeType: "branch", CompanyUID: "01", BranchCode: "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []models.AccessScope{
		{ScopeType: "company", CompanyUID: "C02", BusinessCode: "C02", AllBranches: true},
		{ScopeType: "branch", CompanyUID: "01", BusinessCode: "01", BranchUID: "00001", BranchCode: "00001"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hydrated = %+v, want %+v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// A company rule names no branch: a branch id sent with it is unvalidated and would stop the
// rule from being company-wide, so hydrate drops it.
func TestHydrateAccessScopesClearsBranchOnCompanyRule(t *testing.T) {
	db, mock := holdingOrgMock(t, false)
	got, err := hydrateAccessScopes(context.Background(), db, "rungrueng", []models.AccessScope{
		{ScopeType: "company", BusinessCode: "01", BranchUID: "99999", BranchCode: "not-a-branch", AllBranches: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []models.AccessScope{{ScopeType: "company", CompanyUID: "01", BusinessCode: "01", AllBranches: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hydrated = %+v, want %+v", got, want)
	}
	if !models.ScopesAllowCompanySelection(got, "01") {
		t.Fatal("hydrated company rule must stay company-wide")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHydrateAccessScopesRejectsForeignOrganization(t *testing.T) {
	for name, scope := range map[string]models.AccessScope{
		"company of another holding": {ScopeType: "company", BusinessCode: "X99"},
		"company without identity":   {ScopeType: "company", AllBranches: true},
		"companyuid only, unknown":   {ScopeType: "company", CompanyUID: "C03"},
		"missing branch":             {ScopeType: "branch", CompanyUID: "01", BranchUID: "00009"},
		"branch of another company":  {ScopeType: "branch", CompanyUID: "C02", BranchUID: "00001"},
		"branch without identity":    {ScopeType: "branch", CompanyUID: "01"},
	} {
		t.Run(name, func(t *testing.T) {
			db, _ := holdingOrgMock(t, scope.ScopeType == "branch")
			if _, err := hydrateAccessScopes(context.Background(), db, "rungrueng", []models.AccessScope{scope}); !errors.Is(err, errAccessScopeInvalid) {
				t.Fatalf("err = %v, want errAccessScopeInvalid", err)
			}
		})
	}
}

func TestAccessScopeErrorsMapToHTTPStatus(t *testing.T) {
	if got := shopUserAppError(errAccessScopeInvalid, "th"); got.StatusCode() != http.StatusBadRequest || got.Code != "VALIDATION_FAILED" || got.ThaiMsg == "" {
		t.Fatalf("invalid scope error = %+v", got)
	}
	if got := shopUserAppError(errAccessScopeExceedsGrantor, "en"); got.StatusCode() != http.StatusForbidden || got.Code != "FORBIDDEN" || got.Message == got.ThaiMsg {
		t.Fatalf("grantor error = %+v", got)
	}
}

func TestCheckScopeGrant(t *testing.T) {
	companyA := models.AccessScope{ScopeType: "company", CompanyUID: "A", BusinessCode: "A", AllBranches: true}
	companyB := models.AccessScope{ScopeType: "company", CompanyUID: "B", BusinessCode: "B", AllBranches: true}
	branchA1 := models.AccessScope{ScopeType: "branch", CompanyUID: "A", BranchUID: "00001"}
	branchA2 := models.AccessScope{ScopeType: "branch", CompanyUID: "A", BranchUID: "00002"}
	holding := models.AccessScope{ScopeType: "holding"}
	grantor := func(role models.UserRole, scopes ...models.AccessScope) models.ShopUser {
		return models.ShopUser{ShopUserBase: models.ShopUserBase{Role: role}, AccessScopes: scopes}
	}
	request := func(role models.UserRole, scopes ...models.AccessScope) *models.UserRoleRequest {
		return &models.UserRoleRequest{Role: role, AccessScopes: scopes}
	}
	adminA := grantor(models.ROLE_ADMIN, companyA)
	for _, tc := range []struct {
		name    string
		grantor models.ShopUser
		req     *models.UserRoleRequest
		allowed bool
	}{
		{name: "owner grants holding", grantor: grantor(models.ROLE_OWNER, companyA), req: request(models.ROLE_USER, holding), allowed: true},
		{name: "holding-wide admin grants holding", grantor: grantor(models.ROLE_ADMIN, holding), req: request(models.ROLE_ADMIN, holding), allowed: true},
		{name: "company admin cannot grant holding", grantor: adminA, req: request(models.ROLE_USER, holding)},
		{name: "company admin cannot grant another company", grantor: adminA, req: request(models.ROLE_USER, companyB)},
		{name: "company admin cannot widen self", grantor: adminA, req: request(models.ROLE_ADMIN, companyA, companyB)},
		{name: "company admin cannot create holding-wide admin", grantor: adminA, req: request(models.ROLE_ADMIN)},
		{name: "company admin grants own company", grantor: adminA, req: request(models.ROLE_USER, companyA), allowed: true},
		{name: "company admin grants own branch", grantor: adminA, req: request(models.ROLE_USER, branchA1), allowed: true},
		{name: "company admin keeps a user without scope", grantor: adminA, req: request(models.ROLE_USER), allowed: true},
		{name: "branch admin cannot grant company", grantor: grantor(models.ROLE_ADMIN, branchA1), req: request(models.ROLE_USER, companyA)},
		{name: "branch admin cannot grant other branch", grantor: grantor(models.ROLE_ADMIN, branchA1), req: request(models.ROLE_USER, branchA2)},
		{name: "branch admin grants own branch", grantor: grantor(models.ROLE_ADMIN, branchA1), req: request(models.ROLE_USER, branchA1), allowed: true},
		{name: "company-only admin cannot grant all branches", grantor: grantor(models.ROLE_ADMIN, models.AccessScope{ScopeType: "company", CompanyUID: "A"}), req: request(models.ROLE_USER, companyA)},
		{name: "unknown rule type is not granted", grantor: adminA, req: request(models.ROLE_USER, models.AccessScope{ScopeType: "everything", CompanyUID: "A"})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkScopeGrant(tc.grantor, tc.req)
			if tc.allowed && err != nil {
				t.Fatalf("grant rejected: %v", err)
			}
			if !tc.allowed && !errors.Is(err, errAccessScopeExceedsGrantor) {
				t.Fatalf("err = %v, want errAccessScopeExceedsGrantor", err)
			}
		})
	}
}

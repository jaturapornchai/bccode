//go:build integration

package shop

import (
	"context"
	"errors"
	"testing"
	"time"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb/centraldbtest"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	shopmodels "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/apperr"
	msmodels "smlcloudplatform/pkg/microservice/models"
)

func TestPostgresMembershipFailsClosed(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('h', 'Holding'), ('other', 'Other')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('h', 'C', 'Company')`)
	var uid string
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('member') RETURNING id::text`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	repo := &ShopUserPostgresRepository{db: db}
	svc := NewShopUserService(repo)
	assertNoManagerAccess := func() {
		t.Helper()
		if e := svc.SaveUserFullProfile("h", "member", &models.UserRoleRequest{Username: "member", Role: models.ROLE_OWNER}); e == nil {
			t.Fatal("self promotion allowed")
		}
		list, _, e := repo.FindByUserUIDPage(ctx, uid, msmodels.Pageable{})
		if e != nil || len(list) != 0 {
			t.Fatalf("inaccessible Holding listed: %v %v", list, e)
		}
	}

	if _, err := repo.FindByHoldingCodeAndUserUID(ctx, "h", uid); err == nil {
		t.Fatal("missing membership allowed")
	}
	assertNoManagerAccess()

	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, is_active) VALUES ('h', $1, 'OWNER', false)`, uid)
	disabled, err := repo.FindByHoldingCodeAndUserUID(ctx, "h", uid)
	if err != nil || !disabled.IsAccessDisabled {
		t.Fatalf("disabled membership must be visible as disabled: %+v %v", disabled, err)
	}
	assertNoManagerAccess()

	centraldbtest.Exec(t, db, `UPDATE holding_members SET is_active = true`)
	owner, err := repo.FindByHoldingCodeAndUserUID(ctx, "h", uid)
	// A manager without explicit scopes reads back as the holding rule, not a company snapshot.
	if err != nil || owner.Role != models.ROLE_OWNER || owner.IsAccessDisabled || len(owner.AccessScopes) != 1 || !models.HasHoldingScope(owner.AccessScopes) {
		t.Fatalf("active owner lost scope: %+v %v", owner, err)
	}
	if _, err = repo.FindByHoldingCodeAndUserUID(ctx, "h", "member"); err == nil {
		t.Fatal("username accepted as trusted UID")
	}
	list, _, err := repo.FindByUserUIDPage(ctx, uid, msmodels.Pageable{})
	if err != nil || len(list) != 1 || list[0].HoldingCode != "h" || list[0].Role != models.ROLE_OWNER {
		t.Fatalf("Holding list not membership scoped: %+v %v", list, err)
	}

	centraldbtest.Exec(t, db, `UPDATE users SET is_active = false`)
	if _, err := repo.FindByHoldingCodeAndUserUID(ctx, "h", uid); err == nil {
		t.Fatal("inactive user allowed")
	}
	assertNoManagerAccess()
	centraldbtest.Exec(t, db, `UPDATE users SET is_active = true`)
	centraldbtest.Exec(t, db, `UPDATE holdings SET is_active = false WHERE code = 'h'`)
	if _, err := repo.FindByHoldingCodeAndUserUID(ctx, "h", uid); err == nil {
		t.Fatal("inactive Holding allowed")
	}
	assertNoManagerAccess()
}

func TestPostgresCreateShopRoundTripAndStatusFence(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	var uid string
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('owner@example.com') RETURNING id::text`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	shopRepo := NewShopPostgresRepository(db)
	userRepo := NewShopUserPostgresRepository(db)
	now := time.Date(2026, 9, 23, 8, 0, 0, 0, time.UTC)
	svc := NewShopService(shopRepo, userRepo, func() time.Time { return now })

	th, name := "th", "กลุ่มกิจการรุ่งเรืองกรุ๊ป"
	input := shopmodels.Shop{HoldingCode: "Rungrueng", Names: []common.NameX{{Code: &th, Name: &name}}, Telephone: "021234567"}
	input.Settings.TaxID = "0105560000001"
	code, err := svc.CreateShop(uid, "owner@example.com", input)
	if err != nil || code != "rungrueng" {
		t.Fatalf("CreateShop = %q %v", code, err)
	}

	info, err := svc.InfoShop("rungrueng")
	if err != nil || info.Telephone != "021234567" || info.Settings.TaxID != "0105560000001" || orgaccess.PrimaryName(info.Names) != name || !info.IsActive {
		t.Fatalf("profile did not round-trip: %+v %v", info, err)
	}
	member, err := userRepo.FindByHoldingCodeAndUserUID(ctx, "rungrueng", uid)
	if err != nil || member.Role != models.ROLE_OWNER || member.IsAccessDisabled {
		t.Fatalf("creator is not active OWNER: %+v %v", member, err)
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM organization_audits WHERE holding_code = 'rungrueng' AND action = 'holding.created'`).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("holding.created audits = %d %v", audits, err)
	}

	_, err = svc.CreateShop(uid, "owner@example.com", input)
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) || appErr.Code != "DUPLICATE" {
		t.Fatalf("duplicate CreateShop error = %v", err)
	}

	if err := svc.ChangeShopStatus("rungrueng", uid, "owner@example.com", false, true, "stale"); !errors.Is(err, orgaccess.ErrStatusChangeConflict) {
		t.Fatalf("stale status change error = %v", err)
	}
	if err := svc.ChangeShopStatus("rungrueng", uid, "owner@example.com", true, false, "ปิดกิจการชั่วคราว"); err != nil {
		t.Fatal(err)
	}
	var active bool
	var reason string
	if err := db.QueryRow(`SELECT h.is_active, a.reason FROM holdings h JOIN organization_audits a ON a.holding_code = h.code AND a.action = 'organization.status.changed' WHERE h.code = 'rungrueng'`).Scan(&active, &reason); err != nil || active || reason != "ปิดกิจการชั่วคราว" {
		t.Fatalf("status change not stored with audit: active=%v reason=%q %v", active, reason, err)
	}
	if err := svc.UpdateShop("rungrueng", "owner@example.com", input); err != nil {
		t.Fatalf("closed Holding profile update: %v", err)
	}
}

func TestPostgresSaveFullProfileCreatesUserMembership(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	var ownerUID string
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('owner@example.com') RETURNING id::text`).Scan(&ownerUID); err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name, created_by) VALUES ('h', 'Holding', $1)`, ownerUID)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('h', 'C01', 'Company')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role) VALUES ('h', $1, 'OWNER')`, ownerUID)
	repo := NewShopUserPostgresRepository(db)
	svc := NewShopUserService(repo)

	err := svc.SaveUserFullProfile("h", "owner@example.com", &models.UserRoleRequest{
		Username:     "staff@example.com",
		Role:         models.ROLE_USER,
		AccessScopes: []models.AccessScope{{ScopeType: "company", CompanyUID: "C01"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	staff, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "staff@example.com")
	if err != nil || staff.Role != models.ROLE_USER || staff.UserUID == "" || len(staff.AccessScopes) != 1 || staff.AccessScopes[0].CompanyUID != "C01" {
		t.Fatalf("staff membership = %+v %v", staff, err)
	}

	if err := svc.DeleteUserPermissionShop("h", "owner@example.com", "staff@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "staff@example.com"); err == nil {
		t.Fatal("deleted membership still found")
	}
	if _, err := repo.FindByHoldingCodeAndUserUID(ctx, "h", ownerUID); err != nil {
		t.Fatalf("owner membership lost: %v", err)
	}
}

func TestPostgresHydrateAccessScopesAgainstCentralSchema(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('h', 'Holding'), ('other', 'Other')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('h', 'C01', 'Company'), ('h', 'C02', 'Closed'), ('other', 'X01', 'Other company')`)
	centraldbtest.Exec(t, db, `UPDATE companies SET is_active = false WHERE code = 'C02'`)
	centraldbtest.Exec(t, db, `INSERT INTO branches (holding_code, company_code, code, name) VALUES ('h', 'C01', '00001', 'Branch'), ('other', 'X01', '00002', 'Other branch')`)

	got, err := hydrateAccessScopes(ctx, db, "h", []models.AccessScope{
		{ScopeType: "company", BusinessCode: "c01", AllBranches: true},
		{ScopeType: "branch", CompanyUID: "C01", BranchCode: "1"},
	})
	if err != nil || len(got) != 2 || got[0].CompanyUID != "C01" || got[1].BranchUID != "00001" {
		t.Fatalf("hydrated = %+v err=%v", got, err)
	}
	for _, foreign := range []models.AccessScope{
		{ScopeType: "company", BusinessCode: "X01"},
		{ScopeType: "branch", CompanyUID: "C01", BranchUID: "00002"},
	} {
		if _, err := hydrateAccessScopes(ctx, db, "h", []models.AccessScope{foreign}); !errors.Is(err, errAccessScopeInvalid) {
			t.Fatalf("foreign scope %+v err = %v", foreign, err)
		}
	}

	repo := &ShopUserPostgresRepository{db: db}
	if code, err := repo.ResolveCompanyUID(ctx, "h", "c01"); err != nil || code != "C01" {
		t.Fatalf("active company = %q %v", code, err)
	}
	if _, err := repo.ResolveCompanyUID(ctx, "h", "C02"); err == nil {
		t.Fatal("closed company resolved: a holding scope would open it")
	}
}

// The member list (access audit) and the detail must agree that an OWNER/ADMIN without explicit
// scopes is holding-wide, and a company-restricted ADMIN must not narrow or remove that member
// (review 2026-09-24).
func TestPostgresHoldingWideAdminListedAndProtectedFromCompanyAdmin(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	uids := map[string]string{}
	for _, username := range []string{"owner", "group-admin", "admin-a"} {
		var uid string
		if err := db.QueryRow(`INSERT INTO users (username) VALUES ($1) RETURNING id::text`, username).Scan(&uid); err != nil {
			t.Fatal(err)
		}
		uids[username] = uid
	}
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name, created_by) VALUES ('h', 'Holding', $1)`, uids["owner"])
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('h', 'C01', 'Company A'), ('h', 'C02', 'Company B')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role) VALUES ('h', $1, 'OWNER'), ('h', $2, 'ADMIN')`, uids["owner"], uids["group-admin"])
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, access_scopes) VALUES ('h', $1, 'ADMIN', $2)`,
		uids["admin-a"], `[{"scopetype":"company","companyuid":"C01","businesscode":"C01","allbranches":true}]`)
	repo := &ShopUserPostgresRepository{db: db}

	list, _, err := repo.FindByUserInShopPageWithProfileMatches(ctx, "h", msmodels.Pageable{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]models.ShopUser{}
	for _, member := range list {
		listed[member.Username] = member
	}
	for _, username := range []string{"owner", "group-admin"} {
		detail, err := repo.FindByHoldingCodeAndUsername(ctx, "h", username)
		if err != nil {
			t.Fatal(err)
		}
		for source, scopes := range map[string][]models.AccessScope{"list": listed[username].AccessScopes, "detail": detail.AccessScopes} {
			if len(scopes) != 1 || !models.HasHoldingScope(scopes) {
				t.Fatalf("%s %s scopes = %+v, want the single holding rule", source, username, scopes)
			}
		}
	}
	if scopes := listed["admin-a"].AccessScopes; len(scopes) != 1 || scopes[0].CompanyUID != "C01" {
		t.Fatalf("company admin listed with %+v", scopes)
	}

	svc := NewShopUserService(repo)
	err = svc.SaveUserFullProfile("h", "admin-a", &models.UserRoleRequest{
		Username: "group-admin", Role: models.ROLE_ADMIN,
		AccessScopes: []models.AccessScope{{ScopeType: "company", CompanyUID: "C01", BusinessCode: "C01", AllBranches: true}},
	})
	if !errors.Is(err, errAccessScopeExceedsGrantor) {
		t.Fatalf("narrowing a holding-wide admin: err = %v", err)
	}
	if err := svc.DeleteUserPermissionShop("h", "admin-a", "group-admin"); !errors.Is(err, errAccessScopeExceedsGrantor) {
		t.Fatalf("removing a holding-wide admin: err = %v", err)
	}
	var role, scopes string
	if err := db.QueryRow(`SELECT role, access_scopes::text FROM holding_members WHERE holding_code = 'h' AND user_id = $1`, uids["group-admin"]).Scan(&role, &scopes); err != nil {
		t.Fatalf("holding-wide admin row: %v", err)
	}
	if role != "ADMIN" || (scopes != "[]" && scopes != "{}") {
		t.Fatalf("holding-wide admin row changed: role=%s scopes=%s", role, scopes)
	}

	if err := repo.SaveFullProfile(ctx, "h", &models.UserRoleRequest{Username: "x", UserUID: "not-a-uuid"}); !errors.Is(err, errMemberRequestInvalid) {
		t.Fatalf("malformed useruid: err = %v, want errMemberRequestInvalid", err)
	}
}

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
	if err != nil || owner.Role != models.ROLE_OWNER || owner.IsAccessDisabled || len(owner.AccessScopes) != 1 || owner.AccessScopes[0].CompanyUID != "C" {
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

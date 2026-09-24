//go:build integration

package shop

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb/centraldbtest"
)

// shopUserTestHolding is holding "h" (timezone Asia/Bangkok) owned by owner@example.com, plus
// holding "other" with its own member global01 (a password account).
func shopUserTestHolding(t *testing.T) (*sql.DB, ShopUserService, IShopUserRepository) {
	t.Helper()
	db := centraldbtest.New(t)
	var ownerUID, globalUID string
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('owner@example.com') RETURNING id::text`).Scan(&ownerUID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO users (username, password_hash, email, full_name) VALUES ('global01', 'hash', 'global01@other.co.th', 'มาลี ศรีสุข') RETURNING id::text`).Scan(&globalUID); err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name, created_by, profile) VALUES
		('h', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', $1, '{"settings":{"timezone":"Asia/Bangkok"}}'),
		('other', 'บริษัท หอมกรุ่น คอฟฟี่ จำกัด', $2, '{}')`, ownerUID, globalUID)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('h', 'C01', 'สำนักงานใหญ่')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role) VALUES ('h', $1, 'OWNER'), ('other', $2, 'OWNER')`, ownerUID, globalUID)
	repo := NewShopUserPostgresRepository(db)
	return db, NewShopUserService(repo), repo
}

// Position, department, picture and expiry are saved on the membership, read back, and the expiry
// closes access at 00:00 of that date in the Holding's timezone.
func TestPostgresSaveFullProfilePersistsMembershipFields(t *testing.T) {
	db, svc, repo := shopUserTestHolding(t)
	ctx := context.Background()
	avatar, thumb := "/goapi/s3/file/avatar/somchai.webp", "/goapi/s3/file/avatar/somchai-thumb.webp"
	err := svc.SaveUserFullProfile("h", "owner@example.com", &models.UserRoleRequest{
		Username: "somchai01", Email: "somchai@rungrueng.co.th", UserProfileName: "สมชาย ใจดี",
		Role: models.ROLE_USER, Position: "พนักงานบัญชี", Department: "ฝ่ายบัญชี",
		Avatar: &avatar, AvatarThumb: &thumb, AccessExpiryDate: "2026-12-31",
		AccessScopes: []models.AccessScope{{ScopeType: "company", CompanyUID: "C01"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var position, department, storedAvatar, storedThumb, expiry string
	if err = db.QueryRow(`SELECT m.position, m.department, m.avatar, m.avatar_thumb, m.access_expiry_date::text
		FROM holding_members m JOIN users u ON u.id = m.user_id WHERE m.holding_code = 'h' AND u.username = 'somchai01'`).
		Scan(&position, &department, &storedAvatar, &storedThumb, &expiry); err != nil {
		t.Fatal(err)
	}
	if position != "พนักงานบัญชี" || department != "ฝ่ายบัญชี" || storedAvatar != avatar || storedThumb != thumb || expiry != "2026-12-31" {
		t.Fatalf("PG row = %q %q %q %q %q", position, department, storedAvatar, storedThumb, expiry)
	}

	member, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "somchai01")
	if err != nil {
		t.Fatal(err)
	}
	// usable through the end of 2026-12-31 (Bangkok): access ends at 00:00 of 2027-01-01
	bangkokMidnight := time.Date(2027, 1, 1, 0, 0, 0, 0, time.FixedZone("ICT", 7*3600))
	if member.Position != "พนักงานบัญชี" || member.Department != "ฝ่ายบัญชี" || member.Avatar != avatar ||
		member.AvatarThumb != thumb || !member.AccessExpiryDate.Equal(bangkokMidnight) {
		t.Fatalf("read back = %+v (expiry %s)", member, member.AccessExpiryDate)
	}
	profile, err := svc.InfoShopByUser("h", "somchai01")
	if err != nil || profile.AccessExpiryDate != "2026-12-31" || profile.Avatar != avatar || profile.Position != "พนักงานบัญชี" {
		t.Fatalf("profile = %+v (%v)", profile, err)
	}

	// Edit without picture fields keeps the stored picture; an empty expiry removes it.
	err = svc.SaveUserFullProfile("h", "owner@example.com", &models.UserRoleRequest{
		EditUsername: "somchai01", Username: "somchai01", Role: models.ROLE_USER, Position: "หัวหน้าฝ่ายบัญชี",
		AccessScopes: []models.AccessScope{{ScopeType: "company", CompanyUID: "C01"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var expiryNull bool
	if err = db.QueryRow(`SELECT m.position, m.avatar, m.access_expiry_date IS NULL FROM holding_members m
		JOIN users u ON u.id = m.user_id WHERE m.holding_code = 'h' AND u.username = 'somchai01'`).Scan(&position, &storedAvatar, &expiryNull); err != nil {
		t.Fatal(err)
	}
	if position != "หัวหน้าฝ่ายบัญชี" || storedAvatar != avatar || !expiryNull {
		t.Fatalf("after edit: position %q avatar %q expiry cleared %v", position, storedAvatar, expiryNull)
	}
}

// Adding a usercode that is already a member is a 409; another Holding's login account can never
// be attached by usercode (adversarial review 2026-09-24: it leaked the account's email and locked
// its home Holding out of fixing it); an account no Holding uses needs the explicit addexistinguser
// flag and may not change that account's email.
func TestPostgresSaveFullProfileDuplicateAndExistingAccount(t *testing.T) {
	db, svc, repo := shopUserTestHolding(t)
	ctx := context.Background()
	add := func(req models.UserRoleRequest) error {
		req.Role = models.ROLE_USER
		req.AccessScopes = []models.AccessScope{{ScopeType: "company", CompanyUID: "C01"}}
		return svc.SaveUserFullProfile("h", "owner@example.com", &req)
	}
	if err := add(models.UserRoleRequest{Username: "somchai01", UserProfileName: "สมชาย ใจดี"}); err != nil {
		t.Fatal(err)
	}
	for _, username := range []string{"somchai01", "SomChai01"} {
		err := add(models.UserRoleRequest{Username: username, UserProfileName: "คนอื่น"})
		if !errors.Is(err, errMemberAlreadyExists) {
			t.Fatalf("add %s again: %v", username, err)
		}
		if got := shopUserAppError(err, "th"); got.StatusCode() != http.StatusConflict || got.Field != "username" {
			t.Fatalf("add %s again answered %d field %q", username, got.StatusCode(), got.Field)
		}
	}
	var name string
	if err := db.QueryRow(`SELECT full_name FROM users WHERE username = 'somchai01'`).Scan(&name); err != nil || name != "สมชาย ใจดี" {
		t.Fatalf("duplicate add overwrote the account: %q (%v)", name, err)
	}

	if err := add(models.UserRoleRequest{Username: "global01"}); !errors.Is(err, errUsernameTaken) {
		t.Fatalf("add another Holding's account without the flag: %v", err)
	}
	if err := add(models.UserRoleRequest{Username: "global01", AddExistingUser: true}); !errors.Is(err, errUsernameTaken) {
		t.Fatalf("attach another Holding's account with the flag: %v", err)
	}
	if _, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "global01"); err == nil {
		t.Fatal("another Holding's account was attached")
	}
	var other int
	if err := db.QueryRow(`SELECT count(*) FROM holding_members m JOIN users u ON u.id = m.user_id WHERE u.username = 'global01' AND m.holding_code = 'other' AND m.role = 'OWNER'`).Scan(&other); err != nil || other != 1 {
		t.Fatalf("membership in the other Holding changed: %d (%v)", other, err)
	}
	// From here global01 has left its Holding: an account no Holding uses may be joined with the flag,
	// and without it the answer is errLoginExists (the screen's attach dialog), not errUsernameTaken.
	centraldbtest.Exec(t, db, `DELETE FROM holding_members WHERE holding_code = 'other'`)
	if err := add(models.UserRoleRequest{Username: "global01"}); !errors.Is(err, errLoginExists) {
		t.Fatalf("attachable account without the flag: %v", err)
	}
	if err := add(models.UserRoleRequest{Username: "global01", Email: "changed@rungrueng.co.th", AddExistingUser: true}); !errors.Is(err, errUserEmailLocked) {
		t.Fatalf("add existing account with a new email: %v", err)
	}
	if err := add(models.UserRoleRequest{Username: "global01", AddExistingUser: true}); err != nil {
		t.Fatalf("add existing account with the flag: %v", err)
	}
	var email, password string
	if err := db.QueryRow(`SELECT email, password_hash FROM users WHERE username = 'global01'`).Scan(&email, &password); err != nil ||
		email != "global01@other.co.th" || password != "hash" {
		t.Fatalf("existing account changed: %q %q (%v)", email, password, err)
	}
	if _, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "global01"); err != nil {
		t.Fatalf("existing account not added: %v", err)
	}

	// A claimed account's usercode is locked for this Holding's admin.
	err := svc.SaveUserFullProfile("h", "owner@example.com", &models.UserRoleRequest{
		EditUsername: "global01", Username: "global02", Role: models.ROLE_USER,
		AccessScopes: []models.AccessScope{{ScopeType: "company", CompanyUID: "C01"}},
	})
	if !errors.Is(err, errUserCodeLocked) {
		t.Fatalf("rename of a shared account: %v", err)
	}
}

// review 2026-09-24: a login account removed from THIS business group can be added back — the
// same account is reactivated without the "attach an existing account" confirmation, and the
// removal is recorded (who removed whom). An account used by another group is still a 409.
func TestPostgresReAddRemovedLoginAccount(t *testing.T) {
	db, svc, repo := shopUserTestHolding(t)
	ctx := context.Background()
	add := func(req models.UserRoleRequest) error {
		req.Role = models.ROLE_USER
		req.AccessScopes = []models.AccessScope{{ScopeType: "company", CompanyUID: "C01"}}
		return svc.SaveUserFullProfile("h", "owner@example.com", &req)
	}
	if err := add(models.UserRoleRequest{Username: "somsri01", UserProfileName: "สมศรี มีสุข", Position: "พนักงานบัญชี"}); err != nil {
		t.Fatal(err)
	}
	first, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "somsri01")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteUserPermissionShop("h", "owner@example.com", "somsri01"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "somsri01"); !errors.Is(err, ErrShopUserNotFound) {
		t.Fatalf("membership not removed: %v", err)
	}
	var actor, before string
	if err := db.QueryRow(`SELECT a.actor_uid, a.before_state->>'username' FROM organization_audits a JOIN users o ON o.id::text = a.actor_uid
		WHERE a.holding_code = 'h' AND a.action = 'member_removed' AND a.target_code = $1 AND o.username = 'owner@example.com'`, first.UserUID).Scan(&actor, &before); err != nil || before != "somsri01" {
		t.Fatalf("removal not recorded with the admin: %q %q (%v)", actor, before, err)
	}

	// Re-add WITHOUT addexistinguser: the same account comes back, no new login account.
	if err := add(models.UserRoleRequest{Username: "somsri01", UserProfileName: "สมศรี มีสุข", Position: "หัวหน้าบัญชี"}); err != nil {
		t.Fatalf("re-add of a removed account: %v", err)
	}
	again, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "somsri01")
	if err != nil || again.UserUID != first.UserUID || again.Position != "หัวหน้าบัญชี" || again.IsAccessDisabled {
		t.Fatalf("reactivated membership = %+v (%v)", again, err)
	}
	var accounts int
	if err := db.QueryRow(`SELECT count(*) FROM users WHERE LOWER(username) = 'somsri01'`).Scan(&accounts); err != nil || accounts != 1 {
		t.Fatalf("login accounts for somsri01 = %d (%v)", accounts, err)
	}

	// review 2026-09-24: removed here, and somebody has signed in with it (password/Google) — no silent
	// reconnect: errLoginExists (the attach dialog), nothing written, then the confirmed attach.
	if err := svc.DeleteUserPermissionShop("h", "owner@example.com", "somsri01"); err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `UPDATE users SET password_hash = 'hash' WHERE id = $1::uuid`, first.UserUID)
	if err := add(models.UserRoleRequest{Username: "somsri01", UserProfileName: "สมศรี คนใหม่"}); !errors.Is(err, errLoginExists) {
		t.Fatalf("re-add of a claimed removed account without confirmation: %v", err)
	}
	var claimedName string
	if err := db.QueryRow(`SELECT full_name FROM users WHERE id = $1::uuid`, first.UserUID).Scan(&claimedName); err != nil || claimedName != "สมศรี มีสุข" {
		t.Fatalf("unconfirmed re-add changed the claimed account: %q (%v)", claimedName, err)
	}
	if _, err := repo.FindByHoldingCodeAndUsername(ctx, "h", "somsri01"); !errors.Is(err, ErrShopUserNotFound) {
		t.Fatalf("unconfirmed re-add restored the membership: %v", err)
	}
	if err := add(models.UserRoleRequest{Username: "somsri01", AddExistingUser: true}); err != nil {
		t.Fatalf("confirmed re-add of a claimed account: %v", err)
	}

	// Removed here but now used by another group: still refused (409 username).
	if err := svc.DeleteUserPermissionShop("h", "owner@example.com", "somsri01"); err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role) VALUES ('other', $1, 'USER')`, first.UserUID)
	if err := add(models.UserRoleRequest{Username: "somsri01"}); !errors.Is(err, errUsernameTaken) {
		t.Fatalf("re-add of an account another group uses: %v", err)
	}
	// Never a member here and active in another group: 409 as before.
	if err := add(models.UserRoleRequest{Username: "global01"}); !errors.Is(err, errUsernameTaken) {
		t.Fatalf("another group's account: %v", err)
	}
}

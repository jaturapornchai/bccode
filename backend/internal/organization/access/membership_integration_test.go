//go:build integration

package access_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb/centraldbtest"
	orgpolicy "smlcloudplatform/internal/organization/access"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

const (
	ownerUID = "11111111-1111-1111-1111-111111111111"
	adminUID = "22222222-2222-2222-2222-222222222222"
	userUID  = "33333333-3333-3333-3333-333333333333"
	offUID   = "44444444-4444-4444-4444-444444444444"
)

func seedMembership(t *testing.T) *sql.DB {
	db := centraldbtest.New(t)
	centraldbtest.Exec(t, db, `INSERT INTO users (id, username, full_name, is_active) VALUES
		($1,'owner','เจ้าของ',true), ($2,'admin','ผู้ดูแล',true), ($3,'user','พนักงาน',true), ($4,'off','ปิดใช้',false)`, ownerUID, adminUID, userUID, offUID)
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng','กลุ่มกิจการรุ่งเรืองกรุ๊ป'), ('closed','ปิดแล้ว')`)
	centraldbtest.Exec(t, db, `UPDATE holdings SET is_active=false WHERE code='closed'`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('rungrueng','01','บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด'), ('rungrueng','02','บริษัท หอมกรุ่น คอฟฟี่ จำกัด')`)
	centraldbtest.Exec(t, db, `UPDATE companies SET is_active=false WHERE code='02'`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes) VALUES
		('rungrueng',$1,'OWNER','["*"]','{}'),
		('rungrueng',$2,'admin','[]','[]'),
		('rungrueng',$3,'USER','["ACCOUNTANT"]','[{"scopetype":"branch","companyuid":"01","branchuid":"00001"}]'),
		('rungrueng',$4,'OWNER','[]','{}'),
		('closed',$1,'OWNER','[]','{}')`, ownerUID, adminUID, userUID, offUID)
	return db
}

func info(uid, holding string) micromodels.UserInfo {
	return micromodels.UserInfo{UID: uid, HoldingCode: holding}
}

func TestFindActiveMembershipAgainstCentralSchema(t *testing.T) {
	db := seedMembership(t)
	ctx := context.Background()
	now := time.Now()

	owner, err := orgpolicy.FindActiveMembership(ctx, db, info(ownerUID, "rungrueng"), now)
	if err != nil {
		t.Fatalf("owner: %v", err)
	}
	if owner.Role != authmodels.ROLE_OWNER || owner.MembershipUID == "" || owner.Username != "owner" {
		t.Fatalf("owner membership = %#v", owner)
	}
	// Manager without explicit scopes covers every ACTIVE company.
	if len(owner.AccessScopes) != 1 || owner.AccessScopes[0].CompanyUID != "01" || !owner.AccessScopes[0].AllBranches {
		t.Fatalf("owner scopes = %#v", owner.AccessScopes)
	}

	user, err := orgpolicy.FindActiveMembership(ctx, db, info(userUID, "rungrueng"), now)
	if err != nil || user.Role != authmodels.ROLE_USER || len(user.PermissionSets) != 1 {
		t.Fatalf("user membership = %#v err=%v", user, err)
	}
	if !orgpolicy.AllowsBranch(user.AccessScopes, "01", "00001") || orgpolicy.AllowsAllBranches(user.AccessScopes, "01") {
		t.Fatalf("user scopes = %#v", user.AccessScopes)
	}

	for name, userInfo := range map[string]micromodels.UserInfo{
		"inactive user":    info(offUID, "rungrueng"),
		"inactive holding": info(ownerUID, "closed"),
		"not a member":     info(adminUID, "closed"),
		"no holding":       info(ownerUID, ""),
		"no uid":           info("", "rungrueng"),
		"malformed uid":    info("not-a-uuid", "rungrueng"),
	} {
		if _, err := orgpolicy.FindActiveMembership(ctx, db, userInfo, now); !errors.Is(err, orgpolicy.ErrActiveMembershipRequired) {
			t.Fatalf("%s: err = %v, want ErrActiveMembershipRequired", name, err)
		}
	}

	if _, err := orgpolicy.FindActiveHoldingManager(ctx, db, info(adminUID, "rungrueng"), now); err != nil {
		t.Fatalf("admin must be a manager: %v", err)
	}
	if _, err := orgpolicy.FindActiveHoldingManager(ctx, db, info(userUID, "rungrueng"), now); !errors.Is(err, orgpolicy.ErrHoldingManagerRequired) {
		t.Fatalf("user manager err = %v", err)
	}
	if _, err := orgpolicy.FindActiveHoldingOwner(ctx, db, info(adminUID, "rungrueng"), now); !errors.Is(err, orgpolicy.ErrHoldingOwnerRequired) {
		t.Fatalf("admin owner err = %v", err)
	}

	if err := orgpolicy.RequireActiveHolding(ctx, db, " rungrueng "); err != nil {
		t.Fatalf("active holding: %v", err)
	}
	for _, code := range []string{"closed", "missing", ""} {
		if err := orgpolicy.RequireActiveHolding(ctx, db, code); !errors.Is(err, orgpolicy.ErrActiveHoldingRequired) {
			t.Fatalf("holding %q err = %v", code, err)
		}
	}
}

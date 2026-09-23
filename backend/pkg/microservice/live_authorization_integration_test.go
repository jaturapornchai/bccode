//go:build integration

package microservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"smlcloudplatform/internal/centraldb/centraldbtest"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

func TestLiveAuthorizationAgainstCentralSchema(t *testing.T) {
	db := centraldbtest.New(t)
	var ownerUID, staffUID string
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('owner') RETURNING id::text`).Scan(&ownerUID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('staff') RETURNING id::text`).Scan(&staffUID); err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('rungrueng', '01', 'A'), ('rungrueng', '02', 'B')`)
	centraldbtest.Exec(t, db, `INSERT INTO branches (holding_code, company_code, code, name) VALUES ('rungrueng', '01', '00000', 'สำนักงานใหญ่'), ('rungrueng', '02', '00000', 'สำนักงานใหญ่'), ('rungrueng', '02', '00001', 'สาขาบางนา')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes) VALUES
		('rungrueng', $1, 'OWNER', '["*"]', '{}'),
		('rungrueng', $2, 'USER', '[]', '[{"scopetype":"branch","companyuid":"02","branchuid":"00001"}]')`, ownerUID, staffUID)

	owner, err := microservice.SelectWorkspaceLiveForTest(db, models.UserInfo{UID: ownerUID, CompanyUID: "stale"}, "rungrueng", "01", "00000")
	if err != nil || owner.CompanyUID != "01" || owner.BranchUID != "00000" || owner.Role != 2 || owner.MembershipUID == "" {
		t.Fatalf("owner workspace = %+v %v", owner, err)
	}
	if _, err := microservice.AuthorizeLiveForTest(db, owner); err != nil {
		t.Fatalf("owner re-authorize: %v", err)
	}

	staff := models.UserInfo{UID: staffUID}
	if got, err := microservice.SelectWorkspaceLiveForTest(db, staff, "rungrueng", "02", "00001"); err != nil || got.CompanyUID != "02" {
		t.Fatalf("staff branch workspace = %+v %v", got, err)
	}
	for _, target := range [][2]string{{"02", ""}, {"02", "00000"}, {"01", "00000"}, {"99", ""}} {
		if _, err := microservice.SelectWorkspaceLiveForTest(db, staff, "rungrueng", target[0], target[1]); !errors.Is(err, microservice.ErrLiveWorkspaceAccess) {
			t.Fatalf("staff %v error = %v, want ErrLiveWorkspaceAccess", target, err)
		}
	}
	if _, err := microservice.SelectWorkspaceLiveForTest(db, staff, "other", "", ""); !errors.Is(err, microservice.ErrLiveWorkspaceAccess) {
		t.Fatalf("non-member Holding error = %v", err)
	}

	centraldbtest.Exec(t, db, `UPDATE branches SET is_active = false WHERE company_code = '01' AND code = '00000'`)
	if _, err := microservice.AuthorizeLiveForTest(db, owner); !errors.Is(err, microservice.ErrLiveWorkspaceAccess) {
		t.Fatalf("closed branch error = %v", err)
	}
	centraldbtest.Exec(t, db, `UPDATE holding_members SET is_active = false WHERE user_id = $1`, ownerUID)
	if _, err := microservice.AuthorizeLiveForTest(db, models.UserInfo{UID: ownerUID, HoldingCode: "rungrueng"}); !errors.Is(err, microservice.ErrLiveWorkspaceAccess) {
		t.Fatalf("disabled membership error = %v", err)
	}
	loginOnly, err := microservice.AuthorizeLiveForTest(db, models.UserInfo{UID: ownerUID, HoldingCode: ""})
	if err != nil || loginOnly.HoldingCode != "" {
		t.Fatalf("login-only = %+v %v", loginOnly, err)
	}
	centraldbtest.Exec(t, db, `UPDATE users SET is_active = false WHERE id = $1`, ownerUID)
	if _, err := microservice.AuthorizeLiveForTest(db, models.UserInfo{UID: ownerUID}); !errors.Is(err, microservice.ErrLiveUserAccess) {
		t.Fatalf("disabled user error = %v", err)
	}
}

func TestRevokedMembershipClearsWorkspaceButKeepsLoginSession(t *testing.T) {
	db := centraldbtest.New(t)
	var uid string
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ('member') RETURNING id::text`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng', 'Holding')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('rungrueng', '01', 'A')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, access_scopes) VALUES ('rungrueng', $1, 'USER', '[{"scopetype":"company","companyuid":"01"}]')`, uid)
	cacher, err := microservice.NewCacher(db)
	if err != nil {
		t.Fatal(err)
	}
	authService := microservice.NewAuthServicePrefix("test-auth-", "test-refresh-", cacher, time.Hour, 24*time.Hour, db)
	accessToken, _, err := authService.CreateSession(models.UserInfo{Username: "member", UID: uid})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := authService.SelectShop(microservice.AUTHTYPE_BEARER, accessToken, "rungrueng", "01", "", 0); err != nil {
		t.Fatalf("select workspace: %v", err)
	}
	selected, err := authService.AuthenticateAccessToken(context.Background(), accessToken)
	if err != nil || selected.CompanyUID != "01" {
		t.Fatalf("selected workspace = %#v, err=%v", selected, err)
	}

	centraldbtest.Exec(t, db, `UPDATE holding_members SET is_active = false WHERE user_id = $1`, uid)
	if _, err := authService.AuthenticateAccessToken(context.Background(), accessToken); !errors.Is(err, microservice.ErrLiveWorkspaceAccess) {
		t.Fatalf("revoked membership error = %v, want ErrLiveWorkspaceAccess", err)
	}
	loginOnly, err := authService.AuthenticateAccessToken(context.Background(), accessToken)
	if err != nil || loginOnly.UID != uid || loginOnly.HoldingCode != "" {
		t.Fatalf("login-only identity = %#v, err=%v", loginOnly, err)
	}
}

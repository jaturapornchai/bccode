//go:build integration

package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"smlcloudplatform/internal/centraldb/centraldbtest"
	orgpolicy "smlcloudplatform/internal/organization/access"
	msmodels "smlcloudplatform/pkg/microservice/models"
)

// review 2026-09-24: 50 ทวิ ตรวจสิทธิ์จากสมาชิกภาพจริงในฐานควบคุมกลาง — บริษัท/สาขาในสิทธิ์ และวันหมดอายุ (ใช้ได้ถึงสิ้นวัน)
func TestWhtAccessForFollowsMembershipScopes(t *testing.T) {
	db := centraldbtest.New(t)
	const (
		ownerUID   = "11111111-1111-1111-1111-111111111111"
		branchUID  = "33333333-3333-3333-3333-333333333333"
		expiredUID = "66666666-6666-6666-6666-666666666666"
	)
	centraldbtest.Exec(t, db, `INSERT INTO users (id, username, full_name, is_active) VALUES
		($1,'owner','เจ้าของกิจการ',true), ($2,'branch','พนักงานบัญชีสาขาลาดหลุมแก้ว',true), ($3,'expired','พนักงานบัญชีพ้นสภาพ',true)`,
		ownerUID, branchUID, expiredUID)
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng','กลุ่มกิจการรุ่งเรืองกรุ๊ป')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES
		('rungrueng','01','บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด'), ('rungrueng','02','บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, access_expiry_date) VALUES
		('rungrueng',$1,'OWNER','["*"]','{}',NULL),
		('rungrueng',$2,'USER','["ACCOUNTANT"]','[{"scopetype":"branch","companyuid":"01","branchuid":"00001"}]',NULL),
		('rungrueng',$3,'OWNER','["*"]','{}',DATE '2026-09-30')`, ownerUID, branchUID, expiredUID)
	bangkok := time.FixedZone("ICT", 7*60*60)
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, bangkok)
	user := func(uid, branch string) msmodels.UserInfo {
		return msmodels.UserInfo{UID: uid, HoldingCode: "rungrueng", BranchUID: branch}
	}

	for _, tc := range []struct {
		name    string
		user    msmodels.UserInfo
		company string
		want    error
		branch  string // สาขาของใบสำคัญที่ต้องผ่าน allowsBranch
		denied  string // สาขาของใบสำคัญที่ต้องถูกปฏิเสธ
	}{
		{name: "owner covers every company", user: user(ownerUID, ""), company: "02", branch: "00009"},
		{name: "branch user in own branch", user: user(branchUID, "00001"), company: "01", branch: "00001", denied: "00002"},
		{name: "branch user other company", user: user(branchUID, "00001"), company: "02", want: errWhtScopeDenied},
		{name: "branch user selects another branch", user: user(branchUID, "00002"), company: "01", want: errWhtScopeDenied},
		{name: "expired membership", user: user(expiredUID, ""), company: "01", want: orgpolicy.ErrAccessExpired},
		{name: "not a member", user: msmodels.UserInfo{UID: ownerUID, HoldingCode: "siam"}, company: "01", want: orgpolicy.ErrActiveMembershipRequired},
	} {
		access, err := whtAccessFor(context.Background(), db, tc.user, tc.company, now)
		if (tc.want == nil) != (err == nil) || (tc.want != nil && !errors.Is(err, tc.want)) {
			t.Fatalf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
		if tc.branch != "" && !access.allowsBranch(tc.branch) {
			t.Fatalf("%s: branch %s must be allowed", tc.name, tc.branch)
		}
		if tc.denied != "" && access.allowsBranch(tc.denied) {
			t.Fatalf("%s: branch %s must be refused", tc.name, tc.denied)
		}
	}
	// วันที่หมดอายุเองยังใช้งานได้จนสิ้นวัน (เวลาไทยของกลุ่มกิจการ)
	if _, err := whtAccessFor(context.Background(), db, user(expiredUID, ""), "01", time.Date(2026, 9, 30, 23, 59, 0, 0, bangkok)); err != nil {
		t.Fatalf("last usable day: %v", err)
	}
}

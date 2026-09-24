//go:build integration

package businesstype

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/centraldb/centraldbtest"
	"smlcloudplatform/internal/goapi/language"
	orgpolicy "smlcloudplatform/internal/organization/access"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

// review 2026-09-24: business types are a Holding setting — reading needs an active, unexpired
// membership of the selected Holding and writing needs OWNER/ADMIN (was: any logged-in session).
func TestBusinessTypeAccessFollowsHoldingSettingsRule(t *testing.T) {
	db := centraldbtest.New(t)
	const (
		ownerUID   = "11111111-1111-1111-1111-111111111111"
		userUID    = "33333333-3333-3333-3333-333333333333"
		expiredUID = "66666666-6666-6666-6666-666666666666"
	)
	centraldbtest.Exec(t, db, `INSERT INTO users (id, username, full_name, is_active) VALUES
		($1,'owner','เจ้าของ',true), ($2,'user','พนักงานบัญชี',true), ($3,'expired','พนักงานพ้นสภาพ',true)`, ownerUID, userUID, expiredUID)
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng','กลุ่มกิจการรุ่งเรืองกรุ๊ป'), ('siam','บริษัท สยามพาณิชย์ โฮลดิ้ง จำกัด')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, access_expiry_date) VALUES
		('rungrueng',$1,'OWNER','["*"]','{}',NULL),
		('rungrueng',$2,'USER','["ACCOUNTANT"]','[]',NULL),
		('rungrueng',$3,'ADMIN','[]','{}',DATE '2026-09-30')`, ownerUID, userUID, expiredUID)
	user := func(uid, holding string) micromodels.UserInfo {
		return micromodels.UserInfo{UID: uid, HoldingCode: holding}
	}
	bangkok := time.FixedZone("ICT", 7*60*60)
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, bangkok)

	for _, tc := range []struct {
		name  string
		user  micromodels.UserInfo
		write bool
		want  error
	}{
		{"owner writes", user(ownerUID, "rungrueng"), true, nil},
		{"member reads", user(userUID, "rungrueng"), false, nil},
		{"member cannot write", user(userUID, "rungrueng"), true, orgpolicy.ErrHoldingManagerRequired},
		{"not a member of the other Holding", user(ownerUID, "siam"), false, orgpolicy.ErrActiveMembershipRequired},
		{"expired admin cannot read", user(expiredUID, "rungrueng"), false, orgpolicy.ErrAccessExpired},
	} {
		err := businessTypeAccess(db, tc.user, tc.write, now)
		if (tc.want == nil) != (err == nil) || (tc.want != nil && !errors.Is(err, tc.want)) {
			t.Fatalf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	// The expiry date itself is still usable (through its end in the Holding's timezone).
	if err := businessTypeAccess(db, user(expiredUID, "rungrueng"), true, time.Date(2026, 9, 30, 23, 59, 0, 0, bangkok)); err != nil {
		t.Fatalf("last usable day: %v", err)
	}

	// Refusals are 403 with the Thai message of the languages.tsv row (the raw key means the row is missing).
	for _, tc := range []struct {
		err  error
		text string
	}{
		{orgpolicy.ErrHoldingManagerRequired, language.Text("ss_err_business_type_manager_required", "th")},
		{orgpolicy.ErrActiveMembershipRequired, language.Text("ss_err_holding_membership_required", "th")},
		{orgpolicy.ErrAccessExpired, "ผู้ดูแลกลุ่มกิจการ"},
	} {
		r := &businessTypeRequest{}
		_ = respondBusinessTypeAccessError(r, tc.err)
		raw, _ := json.Marshal(r.data)
		if r.status != http.StatusForbidden || strings.HasPrefix(tc.text, "ss_err_") || !strings.Contains(string(raw), tc.text) {
			t.Fatalf("%v: status %d body %s", tc.err, r.status, raw)
		}
	}
}

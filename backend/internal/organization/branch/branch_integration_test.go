//go:build integration

package branch

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	authModels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb/centraldbtest"
)

func TestBranchQueriesAgainstCentralSchema(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng','กลุ่มกิจการรุ่งเรืองกรุ๊ป')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name) VALUES ('rungrueng','01','บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด'), ('rungrueng','02','บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด')`)
	centraldbtest.Exec(t, db, `INSERT INTO branches (holding_code, company_code, code, name, settings) VALUES
		('rungrueng','01','00000','สำนักงานใหญ่','{"timezone":"Asia/Bangkok","basecurrency":"THB"}'),
		('rungrueng','01','00001','สาขาลาดหลุมแก้ว','{}'),
		('rungrueng','02','00000','สำนักงานใหญ่','{}'),
		('rungrueng','02','00001','สาขาบางนา','{}')`)

	scopes := []authModels.AccessScope{
		{ScopeType: "company", CompanyUID: "01", AllBranches: true},
		{ScopeType: "branch", CompanyUID: "02", BranchUID: "00001"},
	}
	predicate, args := scopedBranchPredicate(scopes, []interface{}{"rungrueng"})
	list, err := queryBranches(ctx, db, "rungrueng", `SELECT `+branchColumns+` FROM branches b WHERE b.holding_code = $1 AND `+predicate+` ORDER BY b.company_code, b.code`, args...)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[2].CompanyUID != "02" || list[2].BranchUID != "00001" {
		t.Fatalf("scoped list = %#v", list)
	}
	if list[0].Timezone != "Asia/Bangkok" || list[0].BaseCurrency != "THB" || list[0].Names[0].Name == nil {
		t.Fatalf("settings/names not decoded: %#v", list[0])
	}

	predicate, args = scopedBranchPredicate(nil, []interface{}{"rungrueng"})
	if none, err := queryBranches(ctx, db, "rungrueng", `SELECT `+branchColumns+` FROM branches b WHERE b.holding_code = $1 AND `+predicate, args...); err != nil || len(none) != 0 {
		t.Fatalf("empty scopes must match nothing: %d %v", len(none), err)
	}

	if _, err := findBranch(ctx, db, "rungrueng", "", "00001", false); !errors.Is(err, errBranchAmbiguous) {
		t.Fatalf("shared code without company err = %v", err)
	}
	got, err := findBranch(ctx, db, "rungrueng", "02", "1", false)
	if err != nil || got.CompanyUID != "02" || got.Code != "00001" {
		t.Fatalf("company-qualified lookup = %#v err=%v", got, err)
	}
	if _, err := findBranch(ctx, db, "rungrueng", "01", "00009", false); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing branch err = %v", err)
	}
	if err := requireActiveBranchCompany(ctx, db, "rungrueng", "03"); err == nil {
		t.Fatal("unknown company must be rejected")
	}
}

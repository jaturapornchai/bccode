package branch

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	authModels "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	branchModels "smlcloudplatform/internal/organization/branch/models"
)

func TestPrepareBranchUpdateRequiresIANATimezone(t *testing.T) {
	for _, test := range []struct {
		name     string
		timezone string
		want     string
		wantErr  error
	}{
		{name: "valid", timezone: " Asia/Bangkok ", want: "Asia/Bangkok"},
		{name: "missing", wantErr: branchModels.ErrBranchTimezoneRequired},
		{name: "offset", timezone: "+07:00", wantErr: branchModels.ErrBranchTimezoneInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := branchModels.BranchOrgDoc{CompanyGuid: "company-id", Code: "00001", BranchSettings: branchModels.BranchSettings{Timezone: test.timezone}, Names: validBranchNames()}
			err := prepareBranchUpdate(&req)
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && err != nil {
				t.Fatal(err)
			}
			if test.wantErr == nil && req.Timezone != test.want {
				t.Fatalf("timezone = %q, want %q", req.Timezone, test.want)
			}
		})
	}
}

func TestPrepareBranchCreateForcesActive(t *testing.T) {
	req := branchModels.BranchOrgDoc{
		HoldingUID: "client-holding", CompanyUID: "company-id", CompanyGuid: "company-id",
		BranchUID: "client-branch", GuidFixed: "client-guid",
		Code: "1", BranchSettings: branchModels.BranchSettings{Timezone: "UTC"}, Names: validBranchNames(),
	}
	now := time.Date(2026, 8, 14, 1, 2, 3, 0, time.FixedZone("test", 7*60*60))
	if err := prepareBranchCreate(&req, "holding", "owner@example.com", now); err != nil {
		t.Fatal(err)
	}
	if !req.IsActive {
		t.Fatalf("unsafe saved status: %#v", req)
	}
	// PostgreSQL identity: branch UID is the branch code, company UID the company code.
	if req.HoldingUID != "holding" || req.BranchUID != "00001" || req.GuidFixed != "00001" || req.BranchCode != "00001" || req.BusinessCode != "company-id" {
		t.Fatalf("unsafe saved stable identity/code: %#v", req)
	}
	if req.BusinessTypes == nil || req.CreatedAt.Location() != time.UTC || req.UpdatedAt.Location() != time.UTC {
		t.Fatalf("canonical array/timestamps missing: %#v", req)
	}
}

func TestValidateBranchNamesRequiresAtLeastOneNonBlankLocalizedName(t *testing.T) {
	code, name, blank := "th", "สาขาทดสอบ", " "
	for _, test := range []struct {
		name    string
		names   common.JSONB
		wantErr bool
	}{
		{name: "missing", wantErr: true},
		{name: "blank name", names: common.JSONB{{Code: &code, Name: &blank}}, wantErr: true},
		{name: "blank language", names: common.JSONB{{Code: &blank, Name: &name}}, wantErr: true},
		{name: "valid", names: validBranchNames()},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateBranchNames(test.names)
			if test.wantErr != (err != nil) {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func validBranchNames() common.JSONB {
	code, name := "th", "สาขาทดสอบ"
	return common.JSONB{{Code: &code, Name: &name}}
}

func TestPreserveBranchParentRejectsClientMove(t *testing.T) {
	existing := branchModels.BranchOrgDoc{HoldingCode: "holding", HoldingUID: "holding-uid", CompanyUID: "company-uid", BranchUID: "branch-uid"}
	for _, req := range []branchModels.BranchOrgDoc{
		{HoldingUID: "other-holding"},
		{CompanyUID: "other-company"},
		{BranchUID: "other-branch"},
	} {
		if err := preserveBranchParent(&req, existing); !errors.Is(err, errBranchParentChange) {
			t.Fatalf("request %#v error = %v", req, err)
		}
	}
	req := branchModels.BranchOrgDoc{}
	if err := preserveBranchParent(&req, existing); err != nil {
		t.Fatal(err)
	}
	if req.HoldingUID != "holding-uid" || req.CompanyUID != "company-uid" || req.BranchUID != "branch-uid" {
		t.Fatalf("server lineage not preserved: %#v", req)
	}
}

func TestBranchCodeCannotChangeThroughOrdinaryUpdate(t *testing.T) {
	existing := branchModels.BranchOrgDoc{Code: "00001", BranchCode: "00001"}
	if err := requireUnchangedBranchCode(existing, "1"); err != nil {
		t.Fatal(err)
	}
	if err := requireUnchangedBranchCode(existing, "00002"); !errors.Is(err, errBranchCodeChange) {
		t.Fatalf("error = %v, want %v", err, errBranchCodeChange)
	}
}

func TestBranchCreateResponseExposesSavedEntity(t *testing.T) {
	payload, err := json.Marshal(orgaccess.OrganizationCreateResponse{Entity: branchModels.BranchOrgDoc{
		HoldingUID: "holding-uid", CompanyUID: "company-uid", BranchUID: "branch-uid", CompanyGuid: "company-uid",
	}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, required := range []string{`"entity"`, `"holdinguid":"holding-uid"`, `"companyuid":"company-uid"`, `"branchuid":"branch-uid"`} {
		if !strings.Contains(text, required) {
			t.Fatalf("response %s missing %s", text, required)
		}
	}
}

func TestScopedBranchPredicateFailsClosedAndBindsScopes(t *testing.T) {
	scopes := []authModels.AccessScope{
		{ScopeType: "company", CompanyUID: "01", AllBranches: true},
		{ScopeType: "branch", CompanyUID: "02", BranchUID: "00001"},
	}
	predicate, args := scopedBranchPredicate(scopes, []interface{}{"holding"})
	if !strings.Contains(predicate, "$2::text[]") || !strings.Contains(predicate, "unnest($3::text[], $4::text[])") || len(args) != 4 {
		t.Fatalf("predicate = %s args = %d", predicate, len(args))
	}
	if _, emptyArgs := scopedBranchPredicate(nil, nil); len(emptyArgs) != 3 {
		t.Fatalf("empty scopes must still bind empty arrays: %d", len(emptyArgs))
	}
}

func TestBranchSettingsRoundTripKeepsFlatJSON(t *testing.T) {
	doc := branchModels.BranchOrgDoc{Code: "00000", BranchSettings: branchModels.BranchSettings{Timezone: "Asia/Bangkok", BaseCurrency: "THB"}}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"timezone":"Asia/Bangkok"`) || strings.Contains(string(raw), "BranchSettings") {
		t.Fatalf("branch settings must stay flat in JSON: %s", raw)
	}
	if !isHeadquarters(doc) {
		t.Fatal("branch 00000 is the head office")
	}
}

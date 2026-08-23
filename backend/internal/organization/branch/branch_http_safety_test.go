package branch

import (
	"context"
	"encoding/json"
	"errors"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	"strings"
	"testing"
	"time"

	branchModels "smlcloudplatform/internal/organization/branch/models"
	companyModels "smlcloudplatform/internal/organization/company/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
			req := branchModels.BranchOrgDoc{CompanyGuid: "company-id", Code: "00001", Timezone: test.timezone, Names: validBranchNames()}
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
		BranchUID: "client-branch", GuidFixed: "client-guid", IsDeleted: true, Version: 99,
		Code: "1", Timezone: "UTC", Names: validBranchNames(),
	}
	now := time.Date(2026, 8, 14, 1, 2, 3, 0, time.FixedZone("test", 7*60*60))
	if err := prepareBranchCreate(&req, "holding", "holding-uid", "owner@example.com", "branch-uid", now); err != nil {
		t.Fatal(err)
	}
	if !req.IsActive || req.IsDeleted || req.Version != 0 || req.ID.IsZero() {
		t.Fatalf("unsafe saved status/ID: %#v", req)
	}
	if req.HoldingUID != "holding-uid" || req.BranchUID != "branch-uid" || req.GuidFixed != "branch-uid" || req.BranchCode != "00001" {
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

type companyFinderStub struct {
	company companyModels.CompanyDoc
	filter  bson.M
}

func (stub *companyFinderStub) FindOne(_ context.Context, _ interface{}, filter interface{}, decode interface{}, _ ...*options.FindOneOptions) error {
	stub.filter = filter.(bson.M)
	*(decode.(*companyModels.CompanyDoc)) = stub.company
	return nil
}

func TestFindActiveBranchCompanyRequiresCanonicalActiveParent(t *testing.T) {
	active := companyModels.CompanyDoc{ID: primitive.NewObjectID(), HoldingUID: "holding-uid", CompanyUID: "company-id"}
	active.IsActive = true
	finder := &companyFinderStub{company: active}
	got, err := findActiveBranchCompany(context.Background(), finder, "holding", "holding-uid", "company-id")
	if err != nil {
		t.Fatal(err)
	}
	if got.CompanyUID != "company-id" || finder.filter["isactive"] != true || finder.filter["holdinguid"] != "holding-uid" || finder.filter["companyuid"] != "company-id" {
		t.Fatalf("company lookup is not fail-closed: %#v", finder.filter)
	}
	if deleted, ok := finder.filter["isdeleted"].(bson.M); !ok || deleted["$ne"] != true {
		t.Fatalf("company lookup does not reject deleted parent: %#v", finder.filter)
	}

	inactive := &companyFinderStub{company: companyModels.CompanyDoc{ID: primitive.NewObjectID(), HoldingUID: "holding-uid", CompanyUID: "company-id"}}
	if _, err := findActiveBranchCompany(context.Background(), inactive, "holding", "holding-uid", "company-id"); !errors.Is(err, branchModels.ErrBranchCompanyNotFound) {
		t.Fatalf("inactive company error = %v", err)
	}
	legacy := companyModels.CompanyDoc{ID: primitive.NewObjectID(), GuidFixed: "company-id"}
	legacy.IsActive = true
	if _, err := findActiveBranchCompany(context.Background(), &companyFinderStub{company: legacy}, "holding", "holding-uid", "company-id"); !errors.Is(err, branchModels.ErrBranchCompanyNotFound) {
		t.Fatalf("legacy parent without stable IDs error = %v", err)
	}
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

func TestBranchManagementFilterExcludesSoftDeleted(t *testing.T) {
	filter := visibleBranchFilter("holding")
	deleted, ok := filter["isdeleted"].(bson.M)
	if !ok || deleted["$ne"] != true {
		t.Fatalf("unsafe management filter: %#v", filter)
	}
}

func TestBranchCreateResponseExposesSavedEntity(t *testing.T) {
	payload, err := json.Marshal(orgaccess.OrganizationCreateResponse{Entity: branchModels.BranchOrgDoc{
		HoldingUID: "holding-uid", CompanyUID: "company-uid", BranchUID: "branch-uid", CompanyGuid: "company-uid",
	}, KafkaSync: "outbox"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, required := range []string{`"entity"`, `"holdinguid":"holding-uid"`, `"companyuid":"company-uid"`, `"branchuid":"branch-uid"`, `"kafka_sync":"outbox"`} {
		if !strings.Contains(text, required) {
			t.Fatalf("response %s missing %s", text, required)
		}
	}
}

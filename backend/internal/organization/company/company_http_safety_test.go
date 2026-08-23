package company

import (
	"encoding/json"
	"errors"
	"os"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCreateCompanyDoesNotSilentlyCreateDefaultBranch(t *testing.T) {
	source, err := os.ReadFile("company_http.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"ensureDefaultHeadOfficeBranch", "defaultbranchid"} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("CreateCompany still contains implicit Branch behavior: %s", forbidden)
		}
	}
}

func TestPrepareCompanyCreateOverridesClientIdentityAndStatus(t *testing.T) {
	code, name := "th", "บริษัททดสอบ"
	req := companyModels.CompanyDoc{
		HoldingUID: "client-holding", CompanyUID: "client-company", GuidFixed: "client-guid",
		IsDeleted: true, Version: 99,
	}
	req.Code = " cmp-001 "
	req.Names = common.JSONB{{Code: &code, Name: &name}}
	now := time.Date(2026, 8, 14, 1, 2, 3, 0, time.FixedZone("test", 7*60*60))
	if err := prepareCompanyCreate(&req, "holding-code", "holding-uid", "owner@example.com", "company-uid", now); err != nil {
		t.Fatal(err)
	}
	if req.HoldingUID != "holding-uid" || req.CompanyUID != "company-uid" || req.GuidFixed != "company-uid" || req.IsDeleted || req.Version != 0 || !req.IsActive {
		t.Fatalf("unsafe saved identity/status: %#v", req)
	}
	if req.ID.IsZero() || req.Code != "CMP-001" {
		t.Fatalf("saved identity/code is incomplete: %#v", req)
	}
	if req.CreatedAt.Location() != time.UTC || req.UpdatedAt.Location() != time.UTC {
		t.Fatalf("timestamps are not UTC: %v %v", req.CreatedAt, req.UpdatedAt)
	}
}

func TestCompanyCodeCannotChangeThroughOrdinaryUpdate(t *testing.T) {
	if err := requireUnchangedCompanyCode(" cmp-001 ", "CMP-001"); err != nil {
		t.Fatal(err)
	}
	if err := requireUnchangedCompanyCode("CMP-001", "CMP-002"); !errors.Is(err, errCompanyCodeChange) {
		t.Fatalf("error = %v, want %v", err, errCompanyCodeChange)
	}
}

func TestCompanyIdentityFilterSupportsCanonicalAndLegacyIDs(t *testing.T) {
	filter := companyIdentityFilter("holding", "company-id")
	identities, ok := filter["$or"].(bson.A)
	if !ok || len(identities) != 2 {
		t.Fatalf("identity filter = %#v", filter)
	}
	if deleted, ok := filter["isdeleted"].(bson.M); !ok || deleted["$ne"] != true {
		t.Fatalf("identity filter does not reject deleted records: %#v", filter)
	}
}

func TestCompanyManagementFilterExcludesSoftDeleted(t *testing.T) {
	filter := visibleCompanyFilter("holding")
	deleted, ok := filter["isdeleted"].(bson.M)
	if !ok || deleted["$ne"] != true {
		t.Fatalf("unsafe management filter: %#v", filter)
	}
}

func TestCompanyCreateResponseExposesSavedEntity(t *testing.T) {
	payload, err := json.Marshal(orgaccess.OrganizationCreateResponse{Entity: companyModels.CompanyDoc{CompanyUID: "company-uid", HoldingUID: "holding-uid"}, KafkaSync: "outbox"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, required := range []string{`"entity"`, `"companyuid":"company-uid"`, `"holdinguid":"holding-uid"`, `"kafka_sync":"outbox"`} {
		if !strings.Contains(text, required) {
			t.Fatalf("response %s missing %s", text, required)
		}
	}
}

func TestValidateCompanyNamesRequiresAtLeastOneNonBlankLocalizedName(t *testing.T) {
	code, name := "th", "บริษัททดสอบ"
	blank := " "

	tests := []struct {
		name    string
		names   common.JSONB
		wantErr bool
	}{
		{name: "missing", wantErr: true},
		{name: "blank name", names: common.JSONB{{Code: &code, Name: &blank}}, wantErr: true},
		{name: "blank language", names: common.JSONB{{Code: &blank, Name: &name}}, wantErr: true},
		{name: "valid", names: common.JSONB{{Code: &code, Name: &name}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCompanyNames(test.names)
			if test.wantErr != (err != nil) {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

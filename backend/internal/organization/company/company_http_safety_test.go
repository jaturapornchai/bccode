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
	if err := prepareCompanyCreate(&req, "holding-code", "owner@example.com", now); err != nil {
		t.Fatal(err)
	}
	if req.HoldingUID != "holding-code" || req.CompanyUID != "CMP-001" || req.GuidFixed != "CMP-001" || req.IsDeleted || req.Version != 0 || !req.IsActive {
		t.Fatalf("unsafe saved identity/status: %#v", req)
	}
	if req.ID != "CMP-001" || req.Code != "CMP-001" {
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

func TestCompanyCreateResponseExposesSavedEntity(t *testing.T) {
	payload, err := json.Marshal(orgaccess.OrganizationCreateResponse{Entity: companyModels.CompanyDoc{CompanyUID: "company-uid", HoldingUID: "holding-uid"}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, required := range []string{`"entity"`, `"companyuid":"company-uid"`, `"holdinguid":"holding-uid"`} {
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

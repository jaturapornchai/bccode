package models

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPrepareCreateBranchRequiresCompanyGuid(t *testing.T) {
	req := BranchPg{Code: "00001"}

	err := PrepareCreateBranch(&req, "shop-1", time.Now(), func() string { return "branch-1" })
	if err == nil {
		t.Fatal("expected missing company_guid error")
	}
	if !strings.Contains(err.Error(), "company_guid") {
		t.Fatalf("error = %q, want company_guid", err.Error())
	}
}

func TestPrepareCreateBranchScopesTenantAndNormalizesCode(t *testing.T) {
	now := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)
	req := BranchPg{
		ShopID:      "client-shop",
		Code:        "1",
		CompanyGuid: " company-1 ",
	}

	err := PrepareCreateBranch(&req, "token-shop", now, func() string { return "branch-1" })
	if err != nil {
		t.Fatalf("prepare create branch: %v", err)
	}

	if req.ShopID != "token-shop" {
		t.Fatalf("ShopID = %q, want token-shop", req.ShopID)
	}
	if req.CompanyGuid != "company-1" {
		t.Fatalf("CompanyGuid = %q, want company-1", req.CompanyGuid)
	}
	if req.Code != "00001" {
		t.Fatalf("Code = %q, want 00001", req.Code)
	}
	if req.GuidFixed != "branch-1" {
		t.Fatalf("GuidFixed = %q, want branch-1", req.GuidFixed)
	}
	if !req.IsActive {
		t.Fatal("expected branch to be active on create")
	}
	if !req.CreatedAt.Equal(now) || !req.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps = %v/%v, want %v", req.CreatedAt, req.UpdatedAt, now)
	}
}

func TestBranchDuplicateLookupScopesByCompanyGuid(t *testing.T) {
	db := newDryRunDB(t)

	stmt := BranchDuplicateLookup(db, "shop-1", "company-1", "00001", "branch-1").
		First(&BranchPg{}).Statement

	assertSQLContains(t, stmt.SQL.String(), "shopid", "company_guid", "code", "guid_fixed")
	assertVars(t, stmt.Vars, "shop-1", "company-1", "00001", "branch-1")
}

func TestCompanyBranchCountLookupScopesByCompanyGuid(t *testing.T) {
	db := newDryRunDB(t)

	stmt := CompanyBranchCountLookup(db, "shop-1", "company-1").Count(new(int64)).Statement

	assertSQLContains(t, stmt.SQL.String(), "shopid", "company_guid")
	assertVars(t, stmt.Vars, "shop-1", "company-1")
}

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm dry run db: %v", err)
	}
	return db
}

func assertSQLContains(t *testing.T, sql string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(sql, part) {
			t.Fatalf("SQL %q does not contain %q", sql, part)
		}
	}
}

func assertVars(t *testing.T, got []interface{}, want ...interface{}) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("vars length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("vars[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

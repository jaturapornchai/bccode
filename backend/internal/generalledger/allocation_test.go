package generalledger

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestAllocationModelSerialization(t *testing.T) {
	master := Master{
		Kind:         "allocations",
		Code:         "ALLOC-RENT",
		Name:         "ปันส่วนค่าเช่าสำนักงาน",
		AccountCode:  "530101",
		AllocateMode: "percent",
		AllocateRules: []AllocationRule{
			{BranchCode: "HQ", DepartmentCode: "SALES", ProjectCode: "PRJ1", AccountCode: "530101", Rate: "40"},
			{BranchCode: "HQ", DepartmentCode: "ADMIN", ProjectCode: "", AccountCode: "530101", Rate: "35"},
			{BranchCode: "HQ", DepartmentCode: "IT", ProjectCode: "PRJ2", AccountCode: "530101", Rate: "25"},
		},
		IsActive: true,
	}

	data, err := json.Marshal(master)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Master
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Code != "ALLOC-RENT" || decoded.AllocateMode != "percent" {
		t.Errorf("unexpected decoded master: %+v", decoded)
	}
	if len(decoded.AllocateRules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(decoded.AllocateRules))
	}
	if decoded.AllocateRules[0].DepartmentCode != "SALES" || decoded.AllocateRules[0].Rate != "40" {
		t.Errorf("unexpected rule 0: %+v", decoded.AllocateRules[0])
	}

	// Sum of rates must equal 100 exactly
	total := decimal.Zero
	for _, rule := range decoded.AllocateRules {
		r, err := decimal.NewFromString(string(rule.Rate))
		if err != nil {
			t.Fatalf("invalid decimal: %v", err)
		}
		total = total.Add(r)
	}
	if !total.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected sum of 100, got %s", total.String())
	}
}

func TestAllocationDimensionCodeValidation(t *testing.T) {
	store := &Store{}
	ctx := context.Background()
	scope := Scope{Holding: "H", Company: "C"}

	validCodes := []string{"SALES", "DEPT-01", "PRJ_100", "HQ", "BRANCH.2", "DEPT/01"}
	for _, c := range validCodes {
		if err := store.checkAllocationDepartment(ctx, scope, c); err != nil {
			t.Errorf("expected valid department %q, got err: %v", c, err)
		}
		if err := store.checkAllocationProject(ctx, scope, c); err != nil {
			t.Errorf("expected valid project %q, got err: %v", c, err)
		}
		if err := store.checkAllocationBranch(ctx, scope, c); err != nil {
			t.Errorf("expected valid branch %q, got err: %v", c, err)
		}
	}

	invalidCodes := []string{"", "SALES@123", "PRJ 100", "BRANCH#1", " 123", "a!b"}
	for _, c := range invalidCodes {
		if err := store.checkAllocationDepartment(ctx, scope, c); err == nil {
			t.Errorf("expected error for department %q, got nil", c)
		}
		if err := store.checkAllocationProject(ctx, scope, c); err == nil {
			t.Errorf("expected error for project %q, got nil", c)
		}
		if err := store.checkAllocationBranch(ctx, scope, c); err == nil {
			t.Errorf("expected error for branch %q, got nil", c)
		}
	}
}

package rolepermission

import (
	"reflect"
	"testing"

	rolemodels "smlcloudplatform/internal/organization/rolepermission/models"
)

func TestUnionPermissionsMergesSetsAndCollapsesWildcard(t *testing.T) {
	got := unionPermissions([]rolemodels.RolePermissionDoc{
		{Permissions: []string{"employee", "employee:update"}},
		{Permissions: []string{"currency", "employee"}},
	})
	if want := []string{"currency", "employee", "employee:update"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unionPermissions() = %v, want %v", got, want)
	}
	if got := unionPermissions([]rolemodels.RolePermissionDoc{{Permissions: []string{"employee"}}, {Permissions: []string{"*"}}}); !reflect.DeepEqual(got, []string{"*"}) {
		t.Fatalf("wildcard union = %v, want [*]", got)
	}
	if got := unionPermissions(nil); len(got) != 0 {
		t.Fatalf("empty union = %v, want empty", got)
	}
}

func TestHasRoleRecord(t *testing.T) {
	records := []rolemodels.RolePermissionDoc{{RoleCode: "ACCOUNTING"}}
	if hasRoleRecord(records, "OWNER") {
		t.Fatal("OWNER should not be found")
	}
	if !hasRoleRecord(append(records, rolemodels.RolePermissionDoc{RoleCode: "OWNER"}), "OWNER") {
		t.Fatal("OWNER should be found")
	}
}

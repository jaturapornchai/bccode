//go:build integration

package employee

import (
	"testing"

	"smlcloudplatform/internal/centraldb/centraldbtest"
	"smlcloudplatform/internal/shop/employee/models"
)

func TestEmployeeColumnsMatchCentralSchema(t *testing.T) {
	db := centraldbtest.New(t)
	roles := []string{"cashier"}
	branches := []models.EmployeeBranch{{Code: "00000"}}
	raw, err := marshalEmployeeJSON(models.Employee{Code: "E001", Roles: &roles, Branches: &branches,
		AccessScopes: []models.EmployeeAccessScope{{ScopeType: "company", CompanyUID: "01"}}})
	if err != nil {
		t.Fatal(err)
	}
	centraldbtest.Exec(t, db, `INSERT INTO employees (id, holding_code, code, name, roles, branches, access_scopes, contact, profile_picture, profile_picture_thumb)
		VALUES ('e1', 'h', 'E001', 'สมชาย ใจดี', $1, $2, $3, $4, '/goapi/s3/file/a.webp', '/goapi/s3/file/a-thumb.webp')`,
		raw.roles, raw.branches, raw.scopes, raw.contact)

	info, err := scanEmployee(db.QueryRow(`SELECT ` + employeeColumns + ` FROM employees WHERE id = 'e1'`))
	if err != nil {
		t.Fatal(err)
	}
	if info.GuidFixed != "e1" || (*info.Roles)[0] != "cashier" || (*info.Branches)[0].Code != "00000" ||
		info.AccessScopes[0].CompanyUID != "01" || info.ProfilePictureThumb != "/goapi/s3/file/a-thumb.webp" {
		t.Fatalf("employee round-trip = %+v", info)
	}
}

package shop

import (
	"net/http"
	"strings"
	"testing"

	"smlcloudplatform/internal/authentication/models"
)

// NUL in member text is a 400 naming the field, not a 500 "try again" (adversarial review 2026-09-24).
func TestMemberNULErrorNamesField(t *testing.T) {
	if err := memberNULError(&models.UserRoleRequest{Username: "somchai01", Position: "หัวหน้าฝ่ายบัญชี"}, "th"); err != nil {
		t.Fatalf("clean request rejected: %v", err)
	}
	err := memberNULError(&models.UserRoleRequest{Username: "somchai01", Position: "ฝ่าย\x00บัญชี"}, "en")
	if err == nil || err.StatusCode() != http.StatusBadRequest || err.Field != "position" {
		t.Fatalf("NUL in position: %+v", err)
	}
	if !strings.Contains(err.ThaiMsg, "ตำแหน่ง") || !strings.Contains(err.ThaiMsg, "NUL") || !strings.Contains(err.Message, "Position") {
		t.Fatalf("message must name the field: th=%q en=%q", err.ThaiMsg, err.Message)
	}
	if err = memberNULError(&models.UserRoleRequest{Username: "somchai01", PermissionSets: []string{"ACC\x00"}}, "th"); err == nil || err.Field != "permissionsets[0]" {
		t.Fatalf("NUL in a permission set: %+v", err)
	}
}

package httpapi

import (
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

// คำสั่งที่คำนวณอย่างเดียว (ไม่บันทึก) ใช้แค่สิทธิ์เปิดจอ — คำสั่งอื่นของ resource เดียวกันยังต้องมีสิทธิ์แก้ไข
func TestCalculationOnlyCommands(t *testing.T) {
	for _, cmd := range []gl.Command{{Resource: "budgets", Action: "spread"}, {Resource: "statement-templates", Action: "suggest"}} {
		if !calculationOnly(cmd) {
			t.Fatalf("%s/%s must be calculation only", cmd.Resource, cmd.Action)
		}
	}
	for _, cmd := range []gl.Command{{Resource: "statement-templates", Action: "update"}, {Resource: "statement-templates", Action: "create"}, {Resource: "budgets", Action: "suggest"}, {Resource: "accounts", Action: "suggest"}} {
		if calculationOnly(cmd) {
			t.Fatalf("%s/%s must need its normal permission", cmd.Resource, cmd.Action)
		}
	}
}

// แนะนำบัญชีของรูปแบบงบการเงินใช้สิทธิ์อ่านจอออกแบบงบการเงิน (financial-statement-designer) เท่านั้น
func TestStatementSuggestNeedsDesignerReadPermission(t *testing.T) {
	screen := resourceScreens["statement-templates"]
	if screen != "financial-statement-designer" {
		t.Fatalf("statement-templates screen = %q", screen)
	}
	if !allowed(map[string]bool{"financial-statement-designer": true}, screen, "") {
		t.Fatal("designer read permission must allow suggest")
	}
	if allowed(map[string]bool{}, screen, "") || allowed(map[string]bool{"gl-journals": true}, screen, "") {
		t.Fatal("suggest allowed without the designer screen")
	}
}

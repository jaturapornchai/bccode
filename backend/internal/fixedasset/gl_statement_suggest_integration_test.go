//go:build integration

package fixedasset_test

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	fa "smlcloudplatform/internal/fixedasset"
	gl "smlcloudplatform/internal/generalledger"
)

// คำสั่งแนะนำบัญชีของรูปแบบงบการเงิน (GL) อ่านบัญชีจากประเภทสินทรัพย์ถาวรที่บันทึกผ่าน store จริง (ชื่อฟิลด์ JSON จริงของ fa_records)
func TestGLStatementSuggestReadsFixedAssetAccounts(t *testing.T) {
	db := faIntegrationDB(t)
	ctx := context.Background()
	connect := func(string) (*sql.DB, error) { return db, nil }
	ledger := gl.NewPostgresStore(gl.NewPostgres(connect))
	glScope := gl.Scope{Holding: "RUNGRUENG_FA", Company: "01", Branch: "00000", Actor: "fa-it"}
	nonce := time.Now().UnixNano()
	step := 0
	execute := func(cmd gl.Command) gl.Result {
		t.Helper()
		step++
		cmd.RequestID = fmt.Sprintf("fa-suggest-%03d-%d", step, nonce)
		result, err := ledger.Execute(ctx, glScope, cmd)
		if err != nil {
			t.Fatalf("GL %s %s: %v", cmd.Resource, cmd.Action, err)
		}
		return result
	}
	for _, a := range []struct {
		code, parent, name, kind, normal string
		posting                          bool
	}{
		{"12000", "", "สินทรัพย์ไม่หมุนเวียน", "asset", "debit", false},
		{"12110", "12000", "อาคารและอุปกรณ์", "asset", "debit", true},
		{"12120", "12000", "ค่าเสื่อมราคาสะสม-อาคารและอุปกรณ์", "asset", "credit", true},
		{"50000", "", "ค่าใช้จ่าย", "expense", "debit", false},
		{"52100", "50000", "ค่าเสื่อมราคา-อาคารและอุปกรณ์", "expense", "debit", true},
	} {
		execute(gl.Command{Resource: "accounts", Action: "create", Account: &gl.Account{AccountCode: a.code, ParentAccountCode: a.parent, Names: []gl.Name{{Code: "th", Name: a.name}},
			AccountType: a.kind, NormalBalance: a.normal, AllowPosting: a.posting, IsActive: true}})
	}
	store := fa.NewStore(connect)
	scope := fa.Scope{Holding: "RUNGRUENG_FA", Company: "01", Actor: "fa-it"}
	for _, code := range []string{"BLD", "EQP"} {
		if _, err := store.CreateAssetType(ctx, scope, fa.AssetType{TypeCode: code, Names: []fa.Name{{Code: "th", Name: "อาคารและอุปกรณ์ " + code}}, DefaultUsefulLifeYears: 20, DefaultDeprecPercent: "5",
			AssetAccountCode: "12110", AccumDeprecAccountCode: "12120", DeprecExpenseAccountCode: "52100"}, time.Date(2026, 9, 27, 3, 0, 0, 0, time.UTC)); err != nil {
			t.Fatal(err)
		}
	}

	result := execute(gl.Command{Resource: "statement-templates", Action: "suggest", Master: &gl.Master{Code: "CF-01", Name: "งบกระแสเงินสด", StatementType: "cash_flow", IsActive: true,
		Rows: []gl.StatementRow{
			{ID: "cf-3", RowNo: 3, RowType: "account", Title: "ค่าเสื่อมราคา", SuggestKey: "depreciation_expense"},
			{ID: "cf-10", RowNo: 10, RowType: "account", Title: "ซื้อที่ดิน อาคารและอุปกรณ์", SuggestKey: "purchase_of_ppe"},
		}}})
	reasons := func(source string) []gl.StatementSuggestionReason {
		return []gl.StatementSuggestionReason{{Source: source, Count: 2}}
	}
	want := [][]gl.StatementSuggestedAccount{
		{{AccountCode: "52100", AccountName: "ค่าเสื่อมราคา-อาคารและอุปกรณ์", IsActive: true, Reasons: reasons("fa_expense")}},
		{{AccountCode: "12110", AccountName: "อาคารและอุปกรณ์", IsActive: true, Reasons: reasons("fa_asset")}},
	}
	if result.Suggestions == nil || len(result.Suggestions.Targets) != 2 {
		t.Fatalf("cash flow suggestions = %+v", result.Suggestions)
	}
	for i, target := range result.Suggestions.Targets {
		if !reflect.DeepEqual(target.Accounts, want[i]) {
			t.Fatalf("%s accounts = %+v", target.ID, target.Accounts)
		}
	}
	result = execute(gl.Command{Resource: "statement-templates", Action: "suggest", Master: &gl.Master{Code: "BS-01", Name: "งบแสดงฐานะการเงิน", StatementType: "balance_sheet", IsActive: true,
		Rows: []gl.StatementRow{{ID: "bs-230", RowNo: 230, RowType: "account", Title: "ที่ดิน อาคารและอุปกรณ์ - สุทธิ", SuggestKey: "property_plant_equipment"}}}})
	wantPPE := []gl.StatementSuggestedAccount{
		{AccountCode: "12110", AccountName: "อาคารและอุปกรณ์", IsActive: true, Reasons: reasons("fa_asset")},
		{AccountCode: "12120", AccountName: "ค่าเสื่อมราคาสะสม-อาคารและอุปกรณ์", IsActive: true, Reasons: reasons("fa_accum")},
	}
	if got := result.Suggestions.Targets[0].Accounts; !reflect.DeepEqual(got, wantPPE) {
		t.Fatalf("ppe accounts = %+v", got)
	}
}

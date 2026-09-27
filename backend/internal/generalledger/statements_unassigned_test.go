package generalledger

import (
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

func unassignedBalance(accountType, name, balance, movement string) statementBalance {
	return statementBalance{accountType: accountType, name: name, balance: decimal.RequireFromString(balance), movement: decimal.RequireFromString(movement)}
}

func unassignedBalances(extra map[string]statementBalance) map[string]statementBalance {
	balances := map[string]statementBalance{
		"1000": unassignedBalance("asset", "เงินสดในมือ", "100", "10"),
		"2000": unassignedBalance("liability", "เจ้าหนี้การค้า", "-50", "-5"),
		"3000": unassignedBalance("equity", "ทุนจดทะเบียน", "-30", "0"),
		"4000": unassignedBalance("income", "รายได้จากการขาย", "-40", "-40"),
		"5000": unassignedBalance("expense", "ค่าเช่าสำนักงาน", "20", "20"),
	}
	for code, balance := range extra {
		balances[code] = balance
	}
	earnings := statementBalance{accountType: "equity"}
	for _, balance := range balances {
		if balance.accountType == "income" || balance.accountType == "expense" {
			earnings.balance = earnings.balance.Add(balance.balance)
			earnings.movement = earnings.movement.Add(balance.movement)
		}
	}
	balances[statementCurrentEarnings] = earnings
	return balances
}

func unassignedCodes(items []StatementUnassigned) []string {
	codes := []string{}
	for _, item := range items {
		codes = append(codes, item.AccountCode)
	}
	return codes
}

func unassignedTemplate(statementType string, rows ...StatementRow) Master {
	return Master{StatementType: statementType, Rows: rows}
}

func TestStatementUnassignedBalanceSheet(t *testing.T) {
	balances := unassignedBalances(nil)
	bs := unassignedTemplate("balance_sheet", StatementRow{ID: "a", RowNo: 10, RowType: "account", AccountCodes: []string{"1000"}},
		StatementRow{ID: "h", RowNo: 20, RowType: "header", AccountCodes: []string{"2000"}}) // รหัสในบรรทัดหัวข้อไม่นับว่าอยู่ในงบ
	got := statementUnassigned(bs, balances, "amount", "2569", 2)
	want := []StatementUnassigned{
		{Key: "amount", FiscalYear: "2569", AccountCode: "2000", AccountName: "เจ้าหนี้การค้า", AccountType: "liability", Basis: "closing", Amount: "50.00"},
		{Key: "amount", FiscalYear: "2569", AccountCode: "3000", AccountName: "ทุนจดทะเบียน", AccountType: "equity", Basis: "closing", Amount: "30.00"},
		// กำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี: รายได้ 40 - ค่าใช้จ่าย 20 = กำไร 20 (ส่วนของเจ้าของเครดิตเป็นบวก) เป็นรายการเดียว
		{Key: "amount", FiscalYear: "2569", AccountCode: statementCurrentEarnings, AccountType: "equity", Basis: "closing", Amount: "20.00"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("balance sheet unassigned = %+v", got)
	}
	// มีบรรทัดกำไรที่ยังไม่ปิดบัญชีแล้ว → ไม่แจ้งกำไร และไม่แจ้งบัญชีรายได้/ค่าใช้จ่ายรายตัว
	withEarnings := bs
	withEarnings.Rows = append(append([]StatementRow{}, bs.Rows...), StatementRow{ID: "e", RowNo: 30, RowType: "account", AccountCodes: []string{"2000", "3000", statementCurrentEarnings}})
	if got = statementUnassigned(withEarnings, balances, "amount", "2569", 2); got != nil {
		t.Fatalf("covered balance sheet = %+v", got)
	}
	// รายได้กับค่าใช้จ่ายหักกันเป็นศูนย์ → ไม่แจ้ง
	zeroNet := unassignedBalances(map[string]statementBalance{"4000": unassignedBalance("income", "รายได้", "-20", "-20")})
	if got = statementUnassigned(withoutCurrentEarnings(withEarnings), zeroNet, "amount", "2569", 2); len(got) != 0 {
		t.Fatalf("zero net earnings flagged = %+v", got)
	}
	// สินทรัพย์ที่อยู่แค่บรรทัดยอดต้นงวดยังไม่อยู่ในงบฐานะการเงิน
	opening := unassignedTemplate("balance_sheet", StatementRow{ID: "o", RowNo: 10, RowType: "account", AmountBasis: "opening", AccountCodes: []string{"1000"}},
		StatementRow{ID: "r", RowNo: 20, RowType: "account", AccountCodes: []string{"2000", "3000", statementCurrentEarnings}})
	if got = statementUnassigned(opening, balances, "prioramount", "2568", 2); !reflect.DeepEqual(unassignedCodes(got), []string{"1000"}) || got[0].Amount != "100.00" || got[0].Key != "prioramount" || got[0].FiscalYear != "2568" {
		t.Fatalf("opening-only asset = %+v", got)
	}
}

// withoutCurrentEarnings: แม่แบบเดิมแต่เอา __current_earnings__ ออกจากทุกบรรทัด
func withoutCurrentEarnings(m Master) Master {
	rows := []StatementRow{}
	for _, row := range m.Rows {
		codes := []string{}
		for _, code := range row.AccountCodes {
			if code != statementCurrentEarnings {
				codes = append(codes, code)
			}
		}
		row.AccountCodes = codes
		rows = append(rows, row)
	}
	m.Rows = rows
	return m
}

func TestStatementUnassignedProfitAndLoss(t *testing.T) {
	balances := unassignedBalances(nil)
	pnl := unassignedTemplate("pnl", StatementRow{ID: "a", RowNo: 10, RowType: "account", AccountCodes: []string{"4000"}})
	want := []StatementUnassigned{{Key: "amount", FiscalYear: "2569", AccountCode: "5000", AccountName: "ค่าเช่าสำนักงาน", AccountType: "expense", Basis: "movement", Amount: "20.00"}}
	if got := statementUnassigned(pnl, balances, "amount", "2569", 2); !reflect.DeepEqual(got, want) {
		t.Fatalf("pnl unassigned = %+v", got)
	}
	withEarnings := unassignedTemplate("pnl", StatementRow{ID: "a", RowNo: 10, RowType: "account", AccountCodes: []string{statementCurrentEarnings}})
	if got := statementUnassigned(withEarnings, balances, "amount", "2569", 2); got != nil {
		t.Fatalf("pnl with earnings = %+v", got)
	}
	openingOnly := unassignedTemplate("pnl", StatementRow{ID: "o", RowNo: 10, RowType: "account", AmountBasis: "opening", AccountCodes: []string{"4000"}},
		StatementRow{ID: "e", RowNo: 20, RowType: "account", AccountCodes: []string{"5000"}})
	if got := statementUnassigned(openingOnly, balances, "amount", "2569", 2); !reflect.DeepEqual(unassignedCodes(got), []string{"4000"}) || got[0].Amount != "40.00" {
		t.Fatalf("income only in an opening row = %+v", got)
	}
}

func TestStatementUnassignedRoundingSignAndTypes(t *testing.T) {
	balances := unassignedBalances(map[string]statementBalance{
		"1500": unassignedBalance("asset", "ปัดเป็นศูนย์", "0.004", "0"),
		"1600": unassignedBalance("asset", "ศูนย์", "0", "0"),
		"2100": unassignedBalance("liability", "ด้านเดบิต", "5.555", "0"),
	})
	bs := unassignedTemplate("balance_sheet", StatementRow{ID: "a", RowNo: 10, RowType: "account", AccountCodes: []string{"1000", "2000", "3000", statementCurrentEarnings}})
	got := statementUnassigned(bs, balances, "amount", "2569", 2)
	if !reflect.DeepEqual(unassignedCodes(got), []string{"2100"}) || got[0].Amount != "-5.56" {
		t.Fatalf("rounding/sign = %+v", got)
	}
	if got = statementUnassigned(bs, balances, "amount", "2569", 0); len(got) != 1 || got[0].Amount != "-6" {
		t.Fatalf("scale 0 = %+v", got)
	}
	for _, statementType := range []string{"production_cost", "cash_flow", "equity", "custom", ""} {
		if got := statementUnassigned(unassignedTemplate(statementType), balances, "amount", "2569", 2); got != nil {
			t.Fatalf("%s must not be checked: %+v", statementType, got)
		}
	}
}

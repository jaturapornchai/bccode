package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestStandardChartEmbeddedFileIsValid(t *testing.T) {
	chart, err := StandardChart()
	if err != nil {
		t.Fatalf("embedded standard chart: %v", err)
	}
	if len(chart) < 100 {
		t.Errorf("standard chart has %d accounts, want at least 100", len(chart))
	}
	maxLevel, roots := 0, 0
	for i, a := range chart {
		if a.Level > maxLevel {
			maxLevel = a.Level
		}
		if a.ParentAccountCode == "" {
			roots++
		}
		if i > 0 && (chart[i-1].Level > a.Level || (chart[i-1].Level == a.Level && chart[i-1].AccountCode >= a.AccountCode)) {
			t.Fatalf("not ordered parents-first at %s", a.AccountCode)
		}
	}
	if maxLevel != 5 {
		t.Errorf("deepest level = %d, want 5", maxLevel)
	}
	if roots != 5 {
		t.Errorf("level-1 accounts = %d, want 5 (asset, liability, equity, income, expense)", roots)
	}
}

func TestStandardChartChangesAreFreshRecords(t *testing.T) {
	chart, err := StandardChart()
	if err != nil {
		t.Fatal(err)
	}
	scope := Scope{Holding: "H", Company: "C", Actor: "tester"}
	changes, err := standardChartChanges(scope, time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), chart)
	if err != nil || len(changes) != len(chart) {
		t.Fatalf("changes = %d, err = %v, want %d", len(changes), err, len(chart))
	}
	ids := map[string]bool{}
	for i, c := range changes {
		var a Account
		if err := json.Unmarshal([]byte(c.Payload), &a); err != nil {
			t.Fatal(err)
		}
		if c.Kind != "accounts" || c.Code != chart[i].AccountCode || a.ID != c.ID || ids[c.ID] || a.Version != 1 || a.HoldingCode != "H" || a.BusinessCode != "C" || a.Level != chart[i].Level {
			t.Fatalf("change %d (%s) = %+v payload %+v", i, c.Code, c, a)
		}
		ids[c.ID] = true
	}
}

// miniChart is the smallest tree that satisfies every rule: a header chain to level 3 with posting leaves.
func miniChart() []Account {
	name := func(s string) []Name { return []Name{{Code: "th", Name: s}} }
	return []Account{
		{AccountCode: "1", Names: name("สินทรัพย์"), AccountType: "asset", NormalBalance: "debit", IsActive: true, Level: 1},
		{AccountCode: "11", Names: name("สินทรัพย์หมุนเวียน"), AccountType: "asset", ParentAccountCode: "1", NormalBalance: "debit", IsActive: true, Level: 2},
		{AccountCode: "111", Names: name("เงินสด"), AccountType: "asset", ParentAccountCode: "11", NormalBalance: "debit", AllowPosting: true, IsActive: true, IsCash: true, Level: 3},
		{AccountCode: "112", Names: name("ลูกหนี้การค้า"), AccountType: "asset", ParentAccountCode: "11", NormalBalance: "debit", AllowPosting: true, IsActive: true, Level: 3},
		{AccountCode: "2", Names: name("หนี้สิน"), AccountType: "liability", NormalBalance: "credit", IsActive: true, Level: 1},
		{AccountCode: "21", Names: name("เจ้าหนี้การค้า"), AccountType: "liability", ParentAccountCode: "2", NormalBalance: "credit", AllowPosting: true, IsActive: true, Level: 2},
	}
}

func TestValidateStandardChart(t *testing.T) {
	if err := validateStandardChart(miniChart()); err != nil {
		t.Fatalf("mini chart should pass: %v", err)
	}
	find := func(chart []Account, code string) *Account {
		for i := range chart {
			if chart[i].AccountCode == code {
				return &chart[i]
			}
		}
		t.Fatalf("no account %s", code)
		return nil
	}
	cases := []struct {
		name   string
		mutate func(chart []Account) []Account
		want   string
	}{
		{"empty chart", func(c []Account) []Account { return nil }, "ว่างเปล่า"},
		{"missing parent", func(c []Account) []Account { find(c, "112").ParentAccountCode = "999"; return c }, "ไม่พบบัญชีแม่"},
		{"posting header", func(c []Account) []Account { find(c, "11").AllowPosting = true; return c }, "ต้องเป็นบัญชีหัวข้อ"},
		{"non-posting leaf", func(c []Account) []Account { find(c, "112").AllowPosting = false; return c }, "ต้องลงรายการได้"},
		{"duplicate code", func(c []Account) []Account {
			dup := *find(c, "21")
			dup.Names = []Name{{Code: "th", Name: "เจ้าหนี้อื่น"}}
			return append(c, dup)
		}, "ซ้ำในผังมาตรฐาน"},
		{"duplicate Thai name", func(c []Account) []Account {
			find(c, "112").Names = []Name{{Code: "th", Name: " เงินสด "}}
			return c
		}, "ซ้ำกันระหว่างรหัส"},
		{"iscash on liability", func(c []Account) []Account { find(c, "21").IsCash = true; return c }, "บัญชีเงินสดต้องเป็นบัญชีสินทรัพย์"},
		{"wrong parent level", func(c []Account) []Account { find(c, "111").Level = 4; return c }, "แต่บัญชีแม่"},
		{"root not level 1", func(c []Account) []Account { find(c, "1").Level = 2; return c }, "ต้องเป็นระดับ 1"},
		{"type differs from root", func(c []Account) []Account {
			find(c, "112").AccountType = "liability"
			find(c, "112").NormalBalance = "credit"
			return c
		}, "ต้องตรงกับบัญชีระดับ 1"},
		{"active under inactive parent", func(c []Account) []Account { find(c, "11").IsActive = false; return c }, "ต้องเปิดใช้งาน"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStandardChart(tc.mutate(miniChart()))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want one containing %q", err, tc.want)
			}
		})
	}
}

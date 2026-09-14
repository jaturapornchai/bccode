package generalledger

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func TestAmountExactBoundaries(t *testing.T) {
	for _, input := range []string{"0", "0.1", "0.2", "0.30", "99999999999999999999999999.12345678", "-0.005"} {
		t.Run(input, func(t *testing.T) {
			a, err := ParseAmount(input)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(struct {
				Amount Amount `json:"amount"`
			}{a})
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Amount Amount `json:"amount"`
			}
			if err = json.Unmarshal(data, &decoded); err != nil || a != decoded.Amount {
				t.Fatalf("JSON round-trip: %s %v", data, err)
			}
			stored, err := bson.Marshal(struct {
				Amount Amount `bson:"amount"`
			}{a})
			if err != nil {
				t.Fatal(err)
			}
			if bson.Raw(stored).Lookup("amount").Type != bsontype.Decimal128 {
				t.Fatal("money not stored as Decimal128")
			}
			var loaded struct {
				Amount Amount `bson:"amount"`
			}
			if err = bson.Unmarshal(stored, &loaded); err != nil || loaded.Amount != a {
				t.Fatalf("BSON round-trip: %s %v", loaded.Amount, err)
			}
		})
	}
	a, _ := ParseAmount("0.1")
	b, _ := ParseAmount("0.2")
	if a.Decimal().Add(b.Decimal()).String() != "0.3" {
		t.Fatal("inexact addition")
	}
	for _, input := range []string{`0.1`, `null`, `"1e2"`, `"NaN"`, `"1,000.00"`, `"01"`, `"1.123456789"`, `""`} {
		var a Amount
		if err := json.Unmarshal([]byte(input), &a); err == nil {
			t.Errorf("accepted invalid money %s", input)
		}
	}
	for _, input := range []string{"0.004", "0.005", "0.006", "-0.005"} {
		a, _ := ParseAmount(input)
		if err := a.ValidateScale(2); err == nil {
			t.Errorf("silently rounded %s", input)
		}
	}
	for _, input := range []string{"0.00", "0.10", "-12.34"} {
		a, _ := ParseAmount(input)
		if err := a.ValidateScale(2); err != nil {
			t.Fatal(err)
		}
	}
	bad, _ := bson.Marshal(bson.M{"amount": 0.1})
	var loaded struct {
		Amount Amount `bson:"amount"`
	}
	if err := bson.Unmarshal(bad, &loaded); err == nil {
		t.Fatal("accepted BSON double")
	}
}

func TestJournalValidation(t *testing.T) {
	accounts := map[string]Account{"101": {AccountCode: "101", IsActive: true, AllowPosting: true}, "301": {AccountCode: "301", IsActive: true, AllowPosting: true}}
	year := FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Currency: "THB", Scale: 2}
	journal := Journal{DocNo: "JV-001", Date: "2026-09-11", BookCode: "JV", FiscalYear: "2026", Description: "ทุนเริ่มต้น", Currency: "THB", Kind: "manual", Lines: []Line{{AccountCode: "101", Debit: "0.1"}, {AccountCode: "101", Debit: "0.2"}, {AccountCode: "301", Credit: "0.3"}}}
	if err := journal.Validate(year, accounts); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*Journal){"unbalanced": func(j *Journal) { j.Lines[2].Credit = "0.31" }, "negative": func(j *Journal) { j.Lines[0].Debit = "-0.1" }, "both sides": func(j *Journal) { j.Lines[0].Credit = "0.1" }, "precision": func(j *Journal) { j.Lines[0].Debit = "0.105" }, "missing account": func(j *Journal) { j.Lines[0].AccountCode = "missing" }, "outside year": func(j *Journal) { j.Date = "2025-12-31" }, "invalid date": func(j *Journal) { j.Date = "2026-02-30" }, "currency": func(j *Journal) { j.Currency = "USD" }}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := journal
			j.Lines = append([]Line(nil), journal.Lines...)
			mutate(&j)
			if err := j.Validate(year, accounts); err == nil {
				t.Fatal("accepted invalid journal")
			}
		})
	}
}

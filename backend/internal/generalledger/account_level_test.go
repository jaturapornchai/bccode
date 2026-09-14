package generalledger

import (
	"encoding/json"
	"testing"
)

func TestAccountLevelValidation(t *testing.T) {
	cases := []struct {
		name    string
		level   int
		wantErr bool
	}{
		{"zero allowed for auto-calc", 0, false},
		{"level 1", 1, false},
		{"level 12", 12, false},
		{"negative level", -1, true},
		{"level 13 out of range", 13, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := Account{
				AccountCode:   "101",
				Names:         []Name{{Code: "th", Name: "เงินสด"}},
				AccountType:   "asset",
				NormalBalance: "debit",
				AllowPosting:  true,
				IsActive:      true,
				Level:         tc.level,
			}
			err := a.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestAccountLevelJsonSerialization(t *testing.T) {
	a := Account{
		AccountCode:   "1100",
		Names:         []Name{{Code: "th", Name: "ลูกหนี้การค้า"}},
		AccountType:   "asset",
		NormalBalance: "debit",
		AllowPosting:  false,
		IsActive:      true,
		Level:         2,
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("Marshal() err = %v", err)
	}

	var decoded Account
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() err = %v", err)
	}
	if decoded.Level != 2 {
		t.Fatalf("Decoded Level = %d, want 2", decoded.Level)
	}
}

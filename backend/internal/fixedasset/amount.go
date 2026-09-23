package fixedasset

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// Amount represents an exact financial decimal amount stored as a decimal string.
type Amount string

func (a Amount) Decimal() decimal.Decimal {
	if a == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(string(a))
	if err != nil {
		return decimal.Zero
	}
	return d
}

func (a Amount) String() string {
	return string(a)
}

func AmountFromDecimal(d decimal.Decimal) Amount {
	return Amount(d.StringFixed(2))
}

func AmountFromFloat(f float64) Amount {
	return Amount(decimal.NewFromFloat(f).StringFixed(2))
}

func AmountFromInt(i int64) Amount {
	return Amount(decimal.NewFromInt(i).StringFixed(2))
}

func (a Amount) MarshalJSON() ([]byte, error) {
	s := string(a)
	if s == "" {
		s = "0.00"
	}
	return json.Marshal(s)
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*a = Amount(s)
		return nil
	}
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*a = Amount(decimal.NewFromFloat(f).StringFixed(2))
		return nil
	}
	return fmt.Errorf("invalid amount JSON")
}

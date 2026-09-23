package generalledger

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/shopspring/decimal"
)

// Amount is an exact decimal at every boundary. JSON numbers and implicit
// rounding are rejected; the company's currency scale is checked separately.
type Amount string

var amountPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,25})(\.[0-9]{1,8})?$`)

func ParseAmount(value string) (Amount, error) {
	if !amountPattern.MatchString(value) {
		return "", fmt.Errorf("จำนวนเงินต้องเป็นเลขทศนิยมไม่เกิน 8 ตำแหน่ง และไม่มีเครื่องหมายคั่นหลักพัน")
	}
	d, err := decimal.NewFromString(value)
	if err != nil {
		return "", err
	}
	return Amount(d.String()), nil
}

func (a Amount) Decimal() decimal.Decimal {
	if a == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(string(a))
	if err != nil {
		panic("invalid internal ledger amount")
	}
	return d
}

func (a Amount) MarshalJSON() ([]byte, error) {
	if a == "" {
		return []byte(`"0"`), nil
	}
	if _, err := ParseAmount(string(a)); err != nil {
		return nil, err
	}
	return json.Marshal(string(a))
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("จำนวนเงินต้องส่งเป็นข้อความทศนิยม")
	}
	parsed, err := ParseAmount(value)
	if err == nil {
		*a = parsed
	}
	return err
}

func amountFromDecimal(d decimal.Decimal) Amount { return Amount(d.String()) }

func (a Amount) ValidateScale(scale int) error {
	if scale < 0 || scale > 8 {
		return fmt.Errorf("จำนวนตำแหน่งทศนิยมไม่ถูกต้อง")
	}
	if _, err := ParseAmount(string(a)); a != "" && err != nil {
		return err
	}
	if !a.Decimal().Equal(a.Decimal().Truncate(int32(scale))) {
		return fmt.Errorf("จำนวนเงินมีทศนิยมเกิน %d ตำแหน่ง กรุณาตรวจสอบก่อนบันทึก", scale)
	}
	return nil
}

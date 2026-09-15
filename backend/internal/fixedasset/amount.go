package fixedasset

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Amount represents an exact financial decimal amount compatible with BSON Decimal128.
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

func (a Amount) MarshalBSONValue() (bsontype.Type, []byte, error) {
	val := string(a)
	if val == "" {
		val = "0"
	}
	d, err := primitive.ParseDecimal128(val)
	if err != nil {
		return bsontype.Null, nil, err
	}
	return bson.MarshalValue(d)
}

func (a *Amount) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bsontype.Decimal128:
		d, ok := (bson.RawValue{Type: t, Value: data}).Decimal128OK()
		if !ok {
			return fmt.Errorf("invalid BSON Decimal128")
		}
		exact, err := decimal.NewFromString(d.String())
		if err != nil {
			return err
		}
		*a = Amount(exact.StringFixed(2))
		return nil
	case bsontype.Double:
		f, ok := (bson.RawValue{Type: t, Value: data}).DoubleOK()
		if !ok {
			return fmt.Errorf("invalid BSON Double")
		}
		*a = Amount(decimal.NewFromFloat(f).StringFixed(2))
		return nil
	case bsontype.String:
		s, ok := (bson.RawValue{Type: t, Value: data}).StringValueOK()
		if !ok {
			return fmt.Errorf("invalid BSON String")
		}
		*a = Amount(s)
		return nil
	case bsontype.Int32:
		i, ok := (bson.RawValue{Type: t, Value: data}).Int32OK()
		if !ok {
			return fmt.Errorf("invalid BSON Int32")
		}
		*a = Amount(fmt.Sprintf("%d.00", i))
		return nil
	case bsontype.Int64:
		i, ok := (bson.RawValue{Type: t, Value: data}).Int64OK()
		if !ok {
			return fmt.Errorf("invalid BSON Int64")
		}
		*a = Amount(fmt.Sprintf("%d.00", i))
		return nil
	default:
		*a = Amount("0.00")
		return nil
	}
}

package generalledger

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAmountSmallestSupportedBSONRoundTrip(t *testing.T) {
	for _, value := range []string{"0.00000001", "0.0000001", "0.000001", "-0.00000001"} {
		t.Run(value, func(t *testing.T) {
			amount, err := ParseAmount(value)
			if err != nil {
				t.Fatal(err)
			}
			original := struct {
				Amount Amount `bson:"amount"`
			}{amount}
			raw, err := bson.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var restored struct {
				Amount Amount `bson:"amount"`
			}
			if err := bson.Unmarshal(raw, &restored); err != nil || restored.Amount != amount {
				t.Fatalf("BSON %s round-trip: got %s; %v", bson.Raw(raw).Lookup("amount").Decimal128().String(), restored.Amount, err)
			}
		})
	}
}
